package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/andygrunwald/jitic/jira"
	"io/ioutil"
	"log"
	"os"
	"regexp"
)

const (
	majorVersion = 0
	minorVersion = 0
	patchVersion = 1
)

var (
	logger *log.Logger
)

func main() {
	var (
		jiraURL       = flag.String("url", "", "JIRA instance URL.")
		jiraUsername  = flag.String("user", "", "JIRA Username.")
		jiraPassword  = flag.String("pass", "", "JIRA Password.")
		ticketMessage = flag.String("tickets", "", "Message to retrieve the tickets from.")
		inputStdin    = flag.Bool("stdin", false, "Set to true if you want to get \"-tickets\" from stdin instead of an argument.")
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
		logger.Fatal("No JIRA-Ticket(s) found.")
	}

	// TODO Add a check for required parameters

	jiraInstance, err := jira.NewJIRAInstance(*jiraURL, *jiraUsername, *jiraPassword)
	if err != nil {
		logger.Fatal(err)
	}

	ok, err := jiraInstance.Authenticate()
	if ok == false || err != nil {
		logger.Fatal(err)
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

func ticketLoop(tickets []string, jiraInstance *jira.JIRA) {
	for _, ticket := range tickets {
		_, err := jiraInstance.GetTicket(ticket)
		if err != nil {
			logger.Fatal(err)
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
