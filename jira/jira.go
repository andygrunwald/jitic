package jira

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

func GetTicket(ticketKey string, parsedURL *url.URL, session *Session) (*Ticket, *Errors) {
	parsedURL.Path = "rest/api/latest/issue/" + ticketKey + ".json"

	req, err := http.NewRequest("GET", parsedURL.String(), nil)
	if err != nil {
		log.Fatal("Can`t build GET / Ticket request", err)
	}
	req.Header.Set("Cookie", fmt.Sprintf("%s=%s", session.Session.Name, session.Session.Value))
	resp, body, err := sendRequest(req)

	if resp.StatusCode != 200 {
		var errors Errors
		err = json.Unmarshal(body, &errors)
		if err != nil {
			log.Fatal("Parsing of error information (during a ticket request) failed.", err)
		}

		return nil, &errors
	}

	var ticket Ticket
	err = json.Unmarshal(body, &ticket)
	if err != nil {
		log.Fatal("Parsing of Ticket information failed.", err)
	}

	return &ticket, nil
}

func AuthAgainstJIRA(parsedURL *url.URL, username, password *string) *Session {
	req := buildAuthRequest(parsedURL, username, password)
	resp, body, err := sendRequest(req)
	if resp.StatusCode != 200 {
		log.Fatal("Auth at JIRA instance failed.")
	}
	var session Session
	err = json.Unmarshal(body, &session)
	if err != nil {
		log.Fatal("Auth at JIRA instance failed.", err)
	}

	return &session
}

// @link https://docs.atlassian.com/jira/REST/latest/#d2e5888
func buildAuthRequest(parsedURL *url.URL, username, password *string) *http.Request {
	parsedURL.Path = "/rest/auth/1/session"
	var jsonStr = []byte(`{"username":"` + *username + `", "password":"` + *password + `"}`)
	req, err := http.NewRequest("POST", parsedURL.String(), bytes.NewBuffer(jsonStr))
	if err != nil {
		log.Fatal("Can`t build Auth request", err)
	}
	req.Header.Set("Content-Type", "application/json")

	return req
}

func sendRequest(req *http.Request) (*http.Response, []byte, error) {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return resp, body, err
}
