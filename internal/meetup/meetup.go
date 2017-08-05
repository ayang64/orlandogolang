package meetup

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Event struct {
	Created     uint   `json:"created"`
	Time        uint   `json:"time"`
	Updated     uint   `json:"updated"`
	RSVPLimit   uint   `json:"rsvp_limit"`
	RSVPed      uint   `json:"yes_rsvp_count"`
	Link        string `json:"link"`
	Description string `json:"description"`
	Id          string `json:"id"`
}

func GetEvents(name string) ([]Event, error) {
	resp, err := http.Get(
		"https://api.meetup.com/" +
			name +
			"/events?&sign=true&status=past,upcoming&photo-host=public")

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
