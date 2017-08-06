package main

import (
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"time"

	"database/sql"
	_ "github.com/lib/pq"

	"github.com/ayang64/fastrow"

	"./internal/meetup"
)

type TemplateRenderer interface {
	Render(io.Writer, string, interface{}) error
}

type StaticTemplate struct {
	Glob string
}

type DebugTemplate struct {
	Glob string
}

func UpdateMeetupDatabase(db *sql.DB) {
	for {
		golangevents, err := meetup.GetEvents("OrlandoGolang")

		if err != nil {
			log.Printf("error fetching OrlandoGolang events.")
		}

		gophersevents, err := meetup.GetEvents("OrlandoGophers")

		if err != nil {
			log.Printf("error fetching OrlandoGophers events.")
			continue
		}

		event := append(golangevents, gophersevents...)

		validmeetups := []string{}

		for i := range event {
			validmeetups = append(validmeetups, event[i].Id)
			_, err := db.Exec(`
								INSERT INTO
										meetups (id, time, created, updated, rsvp_limit, rsvp_count, url, name, description, meetupid)
								VALUES (
									default,
									to_timestamp($1/1000.0),
									to_timestamp($2/1000.0),
									to_timestamp($3/1000.0),
									$4,
									$5,
									$6,
									$7,
									$8,
									$9
								)
								ON CONFLICT (meetupid)
									DO UPDATE
										SET
											time=to_timestamp($10/1000.0),
											created=to_timestamp($11/1000.0),
											updated=to_timestamp($12/1000.0),
											rsvp_limit=$13,
											rsvp_count=$14,
											url=$15,
											name=$16,
											description=$17;`,
				event[i].Time,
				event[i].Created,
				event[i].Updated,
				event[i].RSVPLimit,
				event[i].RSVPed,
				event[i].Link,
				event[i].Name,
				event[i].Description,
				event[i].Id,

				event[i].Time,
				event[i].Created,
				event[i].Updated,
				event[i].RSVPLimit,
				event[i].RSVPed,
				event[i].Link,
				event[i].Name,
				event[i].Description)

			if err != nil {
				log.Printf("ERROR: %s", err)
			}

		}

		time.Sleep(30 * time.Second)
	}
}

func (d DebugTemplate) Render(w io.Writer, name string, data interface{}) error {
	t := template.Must(template.ParseGlob(d.Glob))
	return t.ExecuteTemplate(w, name, data)
}

func RootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<html></html><body><h1>Hello!</h1></body></html>")
}

func main() {

	default_dbhost := "localhost"
	switch runtime.GOOS {
	case "linux":
		default_dbhost = "/var/run/postgresql"
	case "freebsd":
		default_dbhost = "/tmp"
	}

	dbname := flag.String("dbname", "orlandogolang", "Name of database to use.")
	dbuser := flag.String("dbuser", "ayan", "Database user to connect as.")
	dbhost := flag.String("dbhost", default_dbhost, "Unix-domain socket path or hostname of db server to use.")

	network := flag.String("net", "tcp", "Network to listen on.  Should be either \"tcp\" or \"unix\"")
	address := flag.String("addr", "localhost:9898", "Address to listen on.  This should be apropriate to the network chosen.")

	flag.Parse()

	// FIXME: this should be a named function.
	db, err := func(database string, user string, host string) (*sql.DB, error) {
		connstring := fmt.Sprintf("password=\"\" user=%s host=%s dbname=%s sslmode=disable", user, host, database)
		db, err := sql.Open("postgres", connstring)
		if err != nil {
			return nil, err
		}

		if err := db.Ping(); err != nil {
			return nil, err
		}

		return db, nil
	}(*dbname, *dbuser, *dbhost)

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("LISTEN: %s, %s", *network, *address)
	l, err := net.Listen(*network, *address)

	if err != nil {
		log.Fatalf("Could not listen on %s://%s: %s", *network, *address, err)
	}

	if *network == "unix" {
		// make the socket file readable and writable by anyone.
		err := os.Chmod(*address, 0777)

		if err != nil {
			log.Fatalf("could not set permissions of socket file %s to 0777: %s", *address, err)
		}

		defer func() {
			// remove socket file once we exit. this requires that we catch handle SIGINT
			err := os.Remove(*address)
			if err != nil {
				log.Fatal("Could not delete %s: %s", *address, err)
			}
			log.Printf("Cleaned up socket file: %s", *address)
		}()
	}

	log.Printf("Lisening on %q type socket at %q.", *network, *address)

	templater := DebugTemplate{"assets/templates/*.html"}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, to_char(time, 'FMDay, Month DD, YYYY') as timestr, name, description, url from meetups ORDER BY time ASC;")

		type MeetingRow struct {
			Id          int    `col:"id"`
			Time        string `col:"timestr"`
			Name        string `col:"name"`
			Description string `col:"description"`
			URL         string `col:"url"`
		}

		events, err := fastrow.Marshal([]MeetingRow{}, rows)

		if err != nil {
			log.Printf("fastrow: %s", err)
		}

		type IndexPage struct {
			NextMeeting string
			Meetings    []MeetingRow
		}

		index := IndexPage{
			NextMeeting: events.([]MeetingRow)[0].Time,
			Meetings:    events.([]MeetingRow),
		}

		templater.Render(w, "index.html", index)
	})

	http.HandleFunc("/meetings/", func(w http.ResponseWriter, r *http.Request) { templater.Render(w, "meetings.html", nil) })
	http.HandleFunc("/jobs/", func(w http.ResponseWriter, r *http.Request) { templater.Render(w, "jobs.html", nil) })
	http.HandleFunc("/links/", func(w http.ResponseWriter, r *http.Request) { templater.Render(w, "links.html", nil) })

	http.Handle("/images/", http.StripPrefix("/images", http.FileServer(http.Dir("assets/images"))))
	http.Handle("/css/", http.StripPrefix("/css", http.FileServer(http.Dir("assets/css"))))

	go http.Serve(l, nil)
	go UpdateMeetupDatabase(db)

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)

	// exit gracefully by calling all deferred routines after receiving a ^C (SIGINT).
	<-c
}
