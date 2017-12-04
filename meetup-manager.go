package main

import (
	"github.com/ayang64/orlandogolang/internal/meetup"

	"database/sql"
	"log"
	"time"
)

func InsertEvent(db *sql.DB, event meetup.Event) error {
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
		event.Time,
		event.Created,
		event.Updated,
		event.RSVPLimit,
		event.RSVPed,
		event.Link,
		event.Name,
		event.Description,
		event.Id,

		event.Time,
		event.Created,
		event.Updated,
		event.RSVPLimit,
		event.RSVPed,
		event.Link,
		event.Name,
		event.Description)

	return err
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

		// FIXME: This is really ugly.  This is screaming for a stored procedure to
		// handle/hide all of the update ugliness.
		for i := range event {
			validmeetups = append(validmeetups, event[i].Id)

			if err := InsertEvent(db, event[i]); err != nil {
				log.Printf("error inserting event: %s", err)
			}
		}

		// FIXME: We should probably increase the poll interval to something like 15 minutes.
		time.Sleep(30 * time.Second)
	}
}
