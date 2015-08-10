package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/andygrunwald/jitic/jira"
	"log"
	"net/url"
	"os"
	"regexp"
)

const (
	majorVersion = 0
	minorVersion = 0
	patchVersion = 1
)

func main() {
	var (
		jiraURL       = flag.String("url", "", "JIRA instance URL.")
		jiraUsername  = flag.String("user", "", "JIRA Username.")
		jiraPassword  = flag.String("pass", "", "JIRA Password.")
		ticketMessage = flag.String("tickets", "", "Message to retrieve the tickets from.")
		inputStdin    = flag.Bool("stdin", false, "Set to true if you want to get \"-tickets\" from stdin instead of an argument.")
		flagVersion   = flag.Bool("version", false, "Outputs the version number and exits.")
	)
	flag.Parse()

	// Output the version and exit
	if *flagVersion {
		fmt.Printf("jitic v%d.%d.%d\n", majorVersion, minorVersion, patchVersion)
		return
	}

	// Collect all ticket keys
	var tickets []string
	if len(*ticketMessage) > 0 {
		tickets = getTicketsOutOfMessage(*ticketMessage)
	}

	// If we don`t get any ticket, we will just exit here.
	if *inputStdin == false && len(tickets) == 0 {
		log.Fatal("No JIRA-Ticket(s) found.")
	}

	// TODO Add a check for required parameters

	parsedURL, err := url.Parse(*jiraURL)
	if err != nil {
		log.Fatal(err)
	}

	session := jira.AuthAgainstJIRA(parsedURL, jiraUsername, jiraPassword)

	if *inputStdin == false {
		ticketLoop(tickets, parsedURL, session)
	}

	if *inputStdin {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			tickets := getTicketsOutOfMessage(scanner.Text())
			ticketLoop(tickets, parsedURL, session)
		}
	}

	os.Exit(0)
}

func ticketLoop(tickets []string, parsedURL *url.URL, session *jira.Session) {
	for _, ticket := range tickets {
		_, errors := jira.GetTicket(ticket, parsedURL, session)
		if errors != nil {
			log.Fatal(errors)
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
