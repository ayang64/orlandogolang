package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
)

func RootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<html></html><body><h1>Hello!</h1></body></html>")
}

func main() {
	network := flag.String("net", "tcp", "Network to listen on.  Should be either \"tcp\" or \"unix\"")
	address := flag.String("addr", "localhost:9898", "Addressto listen on.  This should be apropriate to the network chosen.")

	listener, err := net.Listen(*network, *address)

	if err != nil {
		log.Fatalf("Could not listen on %s://%s: %s", network, address, err)
	}

	log.Printf("Lisening on %q type socket at %q.", *network, *address)

	http.HandleFunc("/", RootHandler)

	log.Fatalf("could not serve: %s", http.Serve(listener, nil))
}
