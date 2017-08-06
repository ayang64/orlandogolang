package meetup

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type Event struct {
	Time        int    `json:"time"`
	Created     int    `json:"created"`
	Updated     int    `json:"updated"`
	RSVPLimit   int    `json:"rsvp_limit"`
	RSVPed      int    `json:"yes_rsvp_count"`
	Link        string `json:"link"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Id          string `json:"id"`
}

func GetEvents(name string) ([]Event, error) {

	client := http.Client{
		Timeout: time.Second * 5,
	}

	resp, err := client.Get(
		"https://api.meetup.com/" +
			name +
			"/events?&sign=true&status=upcoming&photo-host=public")

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API Error: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	events := []Event{}
	if err := json.Unmarshal(body, &events); err != nil {
		return nil, err
	}

	return events, nil
}
