package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/andygrunwald/go-jira"
	"io/ioutil"
	"log"
	"os"
	"regexp"
)

const (
	// Version of jitic
	Version = "0.2.0"
)

var (
	logger *log.Logger
)

func main() {
	var (
		jiraURL       = flag.String("url", "", "JIRA instance URL (format: scheme://[username[:password]@]host[:port]/).")
		jiraUsername  = flag.String("user", "", "JIRA Username.")
		jiraPassword  = flag.String("pass", "", "JIRA Password.")
		ticketMessage = flag.String("tickets", "", "Message to retrieve the tickets from.")
		inputStdin    = flag.Bool("stdin", false, "If set to true you can stream \"-tickets\" to stdin instead of an argument. If set \"-tickets\" will be ignored.")
		flagVersion   = flag.Bool("version", false, "Outputs the version number and exits.")
		flagVerbose   = flag.Bool("verbose", false, "If activated more information will be written to stdout .")
	)
	flag.Parse()

	// Set logger (throw messages away)
	logger = log.New(ioutil.Discard, "", log.LstdFlags)
	if *flagVerbose {
		logger = log.New(os.Stdout, "", log.LstdFlags)
	}

	// Output the version and exit
	if *flagVersion {
		fmt.Printf("jitic v%s\n", Version)
		return
	}

	// Collect all ticket keys
	var tickets []string
	if len(*ticketMessage) > 0 {
		tickets = getTicketsOutOfMessage(*ticketMessage)
	}

	// If we don`t get any ticket, we will just exit here.
	if *inputStdin == false && len(tickets) == 0 {
		logger.Fatal("No JIRA-Ticket(s) found.")
	}

	// TODO Add a check for required parameters
	// Required params are:
	//	* jiraURL
	//	* jiraUsername
	//	* jiraPassword
	//	* ticketMessage or inputStdin

	// Get the JIRA client
	jiraInstance, err := jira.NewClient(nil, *jiraURL)
	if err != nil {
		logger.Fatal(err)
	}

	// Only provide authentification if a username and password was applied
	if len(*jiraUsername) > 0 && len(*jiraPassword) > 0 {
		ok, err := jiraInstance.Authentication.AcquireSessionCookie(*jiraUsername, *jiraPassword)
		if ok == false || err != nil {
			logger.Fatal(err)
		}
	}

	if *inputStdin == false {
		ticketLoop(tickets, jiraInstance)
	}

	if *inputStdin {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			tickets := getTicketsOutOfMessage(scanner.Text())
			// If no ticket can be found
			if len(tickets) == 0 {
				logger.Fatal("No JIRA-Ticket(s) found.")
			}
			ticketLoop(tickets, jiraInstance)
		}
	}

	os.Exit(0)
}

func ticketLoop(tickets []string, jiraInstance *jira.Client) {
	for _, ticket := range tickets {
		/*
			// Add Ticket-Key at first item in the slice
			if len(ticketKey) > 0 {
				listOfErrors = append([]string{ticketKey}, listOfErrors...)
			}
		*/
		issue, _, err := jiraInstance.Issue.Get(ticket)
		if err != nil {
			logger.Fatal(err)
		}
		if ticket != issue.Key {
			log.Fatalf("Used issue %s is not the same as %s (provided by JIRA)", ticket, issue.Key)
		}
	}
}

// getTicketsOutOfMessage will retrieve all JIRA ticket numbers out of a text.
// A text can be everything, but a use case is e.g. a commit message.
// Example:
//		Text: WEB-22861 remove authentication prod build for now
//		Result: WEB-22861
//
//		Text: TASKLESS: Removes duplicated comment code.
//		Result: Empty slice
//
// @link https://confluence.atlassian.com/display/STASHKB/Integrating+with+custom+JIRA+issue+key
// @link https://answers.atlassian.com/questions/325865/regex-pattern-to-match-jira-issue-key
func getTicketsOutOfMessage(ticketMessage string) []string {
	// Normally i would use
	//		((?<!([A-Z]{1,10})-?)[A-Z]+-\d+)
	// See http://stackoverflow.com/questions/26771592/negative-look-ahead-go-regular-expressions
	re := regexp.MustCompile("([A-Z]+-\\d+)")
	return re.FindAllString(ticketMessage, -1)
}
