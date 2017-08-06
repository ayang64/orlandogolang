package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"

	"database/sql"
	_ "github.com/lib/pq"
)

func PostgresConnect(user, host, database string) (*sql.DB, error) {
	connstring := fmt.Sprintf("password=\"\" user=%s host=%s dbname=%s sslmode=disable", user, host, database)

	db, err := sql.Open("postgres", connstring)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func main() {
	log.SetFlags(log.LstdFlags  | log.Lshortfile | log.LUTC)

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


	webroot	:=	flag.String("webroot", "./assets", "Direcotry where web assets reside.")
	network := flag.String("net", "tcp", "Network to listen on.  Should be either \"tcp\" or \"unix\"")
	address := flag.String("addr", "localhost:9898", "Address to listen on.  This should be apropriate to the network chosen.")

	flag.Parse()

	db, err := PostgresConnect(*dbuser, *dbhost, *dbname)

	if err != nil {
		log.Fatal(err)
	}

	webstop := make(chan struct{})

	// FIXME: i should probably wait until UpdateMeetupDatabase() is complete.
	// We shouldn't quit mid-query.
	go Webserver(webstop, *network, *address, db, *webroot)
	go UpdateMeetupDatabase(db)

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)

	// wait for a SIGINT...
	<-c

	// Signal to webserver that we're shutting down.
	webstop <- struct{}{}

	// Wait for web server to clean up after itself.
	<- webstop

	// fin
}
