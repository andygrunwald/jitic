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

type JIRA struct {
	parsedURL *url.URL
	session   *Session
	username  string
	password  string
}

func NewJIRAInstance(address, username, password string) (*JIRA, error) {
	parsedURL, err := url.Parse(address)
	if err != nil {
		return nil, err
	}

	instance := &JIRA{
		parsedURL: parsedURL,
		session:   nil,
		username:  username,
		password:  password,
	}

	return instance, nil
}

func (j *JIRA) GetTicket(ticketKey string) (*Ticket, *Errors) {
	j.parsedURL.Path = "rest/api/latest/issue/" + ticketKey + ".json"

	req, err := http.NewRequest("GET", j.parsedURL.String(), nil)
	if err != nil {
		log.Fatal("Can`t build GET / Ticket request", err)
	}
	req.Header.Set("Cookie", fmt.Sprintf("%s=%s", j.session.Session.Name, j.session.Session.Value))
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

func (j *JIRA) Authenticate() (bool, error) {
	req, err := j.buildAuthRequest()
	if err != nil {
		return false, err
	}

	resp, body, err := sendRequest(req)
	if resp.StatusCode != 200 || err != nil {
		return false, fmt.Errorf("Auth at JIRA instance failed (HTTP(S) request). %s", err)
	}

	var session Session
	err = json.Unmarshal(body, &session)
	if err != nil {
		return false, fmt.Errorf("Auth at JIRA instance failed (Reading response). %s", err)
	}

	j.session = &session

	return true, nil
}

// @link https://docs.atlassian.com/jira/REST/latest/#d2e5888
func (j *JIRA) buildAuthRequest() (*http.Request, error) {
	j.parsedURL.Path = "/rest/auth/1/session"
	var jsonStr = []byte(`{"username":"` + j.username + `", "password":"` + j.password + `"}`)
	req, err := http.NewRequest("POST", j.parsedURL.String(), bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, fmt.Errorf("Can`t build Auth request. %s", err)
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func sendRequest(req *http.Request) (*http.Response, []byte, error) {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, make([]byte, 0), err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return resp, body, err
}
