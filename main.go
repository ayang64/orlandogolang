package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
)

func RootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<html></html><body><h1>Hello!</h1></body></html>")
}

func main() {
	network := flag.String("net", "tcp", "Network to listen on.  Should be either \"tcp\" or \"unix\"")
	address := flag.String("addr", "localhost:9898", "Addressto listen on.  This should be apropriate to the network chosen.")

	flag.Parse()

	l, err := net.Listen(*network, *address)

	if err != nil {
		log.Fatalf("could not listen on %s://%s: %s", *network, *address, err)
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

	if err != nil {
		log.Fatalf("Could not listen on %s://%s: %s", *network, *address, err)
	}

	log.Printf("Lisening on %q type socket at %q.", *network, *address)

	http.HandleFunc("/", RootHandler)

	go http.Serve(l, nil)

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)

	// exit gracefully by calling all deferred routines after receiving a ^C (SIGINT).
	<-c
}
