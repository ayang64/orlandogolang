package main

import (
	"database/sql"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/ayang64/fastrow"
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

func (d DebugTemplate) Render(w io.Writer, name string, data interface{}) error {
	t := template.Must(template.ParseGlob(d.Glob))
	return t.ExecuteTemplate(w, name, data)
}

func Webserver(stop chan struct{}, network string, address string, db *sql.DB, root string /* not used at the moment */) {
	l, err := net.Listen(network, address)

	if err != nil {
		log.Fatalf("Could not listen on %s://%s: %s", network, address, err)
	}

	if network == "unix" {
		// make the socket file readable and writable by anyone.
		err := os.Chmod(address, 0777)

		if err != nil {
			log.Fatalf("could not set permissions of socket file %s to 0777: %s", address, err)
		}

		defer func() {
			// remove socket file once we exit. this requires that we catch handle SIGINT
			err := os.Remove(address)
			if err != nil {
				log.Fatalf("Could not delete %s: %s", address, err)
			}
			log.Printf("Cleaned up socket file: %s", address)
			stop <- struct{}{}
		}()
	}

	log.Printf("Lisening on %q type socket at %q.", network, address)

	templater := DebugTemplate{"assets/templates/*.html"}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, to_char(time, 'FMDay, Month DD, YYYY') as timestr, name, description, rsvp_count, url from meetups ORDER BY time ASC;")

		type MeetingRow struct {
			Id          int    `col:"id"`
			Time        string `col:"timestr"`
			Name        string `col:"name"`
			Description string `col:"description"`
			RSVP        int    `col:"rsvp_count"`
			URL         string `col:"url"`
		}
		events, err := fastrow.Marshal([]MeetingRow{}, rows)
		if err != nil {
			log.Printf("fastrow: %s", err)
		}
		type IndexPage struct {
			NextMeetingDate string
			NextMeetingRSVP int
			Meetings        []MeetingRow
		}

		index := IndexPage{
			NextMeetingDate: events.([]MeetingRow)[0].Time,
			NextMeetingRSVP: events.([]MeetingRow)[0].RSVP,
			Meetings:        events.([]MeetingRow),
		}

		templater.Render(w, "index.html", index)
	})

	http.HandleFunc("/meetings/", func(w http.ResponseWriter, r *http.Request) { templater.Render(w, "meetings.html", nil) })
	http.HandleFunc("/jobs/", func(w http.ResponseWriter, r *http.Request) { templater.Render(w, "jobs.html", nil) })
	http.HandleFunc("/links/", func(w http.ResponseWriter, r *http.Request) { templater.Render(w, "links.html", nil) })

	http.Handle("/images/", http.StripPrefix("/images", http.FileServer(http.Dir("assets/images"))))
	http.Handle("/css/", http.StripPrefix("/css", http.FileServer(http.Dir("assets/css"))))

	go http.Serve(l, nil)

	<-stop


}
