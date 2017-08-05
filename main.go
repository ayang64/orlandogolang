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

func RootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<html></html><body><h1>Hello!</h1></body></html>")
}

func main() {
	network := flag.String("net", "tcp", "Network to listen on.  Should be either \"tcp\" or \"unix\"")
	address := flag.String("addr", "localhost:9898", "Addressto listen on.  This should be apropriate to the network chosen.")
	flag.Parse()

	l, err := net.Listen(*network, *address)

	if err != nil {
		log.Fatalf("Could not listen on %s://%s: %s", *network, *address, err)
	}

	defer func() {
		// remove socket file once we exit. this requires that we catch handle SIGINT
		if *network == "unix" {
			err := os.Remove(*address)
			if err != nil {
				log.Fatal("Could not delete %s: %s", *address, err)
			}
			log.Printf("Cleaned up socket file: %s", *address)
		}
		log.Printf("Exiting.")
	}()

	log.Printf("Lisening on %q type socket at %q.", *network, *address)

	templater := DebugTemplate{"assets/templates/*.html"}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { templater.Render(w, "index.html", nil) })

	http.HandleFunc("/meetings/", func(w http.ResponseWriter, r *http.Request) { templater.Render(w, "meetings.html", nil) })
	http.HandleFunc("/jobs/", func(w http.ResponseWriter, r *http.Request) { templater.Render(w, "jobs.html", nil) })
	http.HandleFunc("/links/", func(w http.ResponseWriter, r *http.Request) { templater.Render(w, "links.html", nil) })

	http.Handle("/images/", http.StripPrefix("/images", http.FileServer(http.Dir("assets/images"))))
	http.Handle("/css/", http.StripPrefix("/css", http.FileServer(http.Dir("assets/css"))))

	go http.Serve(l, nil)

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)

	// exit gracefully by calling all deferred routines after receiving a ^C (SIGINT).
	<-c
}
