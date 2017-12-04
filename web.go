package main

import (
	"database/sql"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

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

type app struct {
	db      *sql.DB
	tmpl    TemplateRenderer
	filesvr http.Handler
}

func (a *app) rootHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := a.db.Query("SELECT id, to_char(time, 'FMDay, Month DD, YYYY') as timestr, name, description, rsvp_count, url from meetups ORDER BY time ASC;")
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

	a.tmpl.Render(w, "index.html", index)
}

func (a *app) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// if there is a trailing slash, strip it and redirect to
	// the new url.
	if url := r.URL.Path; len(url) > 1 && url[len(url)-1] == '/' {
		dst := url[:len(url)-1]
		http.Redirect(w, r, dst, 301)
		return
	}

	switch url := r.URL.Path; {
	case url == "/":
		a.rootHandler(w, r)
		return

	case url == "/links":
		log.Printf("links: %q", url)
		a.tmpl.Render(w, "links.html", nil)
		return

	case strings.HasPrefix(url, "/static/"):
		log.Printf("static: %q", url)
		a.filesvr.ServeHTTP(w, r)
		return
	}
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
		}()
	}

	log.Printf("Lisening on %q type socket at %q.", network, address)

	templater := DebugTemplate{"assets/templates/*.html"}

	srv := http.Server{
		Handler: &app{
			db:      db,
			tmpl:    templater,
			filesvr: http.StripPrefix("/static", http.FileServer(http.Dir(root))),
		},
	}

	defer func() { stop <- struct{}{} }()
	go srv.Serve(l)

	<-stop
}
