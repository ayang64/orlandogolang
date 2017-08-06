

DROP TABLE IF EXISTS meetups;

CREATE TABLE meetups (
	id					serial unique primary key,

	"time"			timestamp,
	created			timestamp,		/* when this event was created. */	
	updated			timestamp,		/* when this event was last changed .*/
	rsvp_limit	integer,			/* maximum number of people allowed */
	rsvp_count	integer,			/* number of received rsvps */
	url					text,					/* link to event on meetups.com */
	name				text,
	description	text,					/* description of event. */
	meetupid		text unique		/* unique id of event.  this is meetup.com's key for the event. */
)
