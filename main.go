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
		jiraURL      = flag.String("url", "", "JIRA instance URL (format: scheme://[username[:password]@]host[:port]/).")
		jiraUsername = flag.String("user", "", "JIRA Username.")
		jiraPassword = flag.String("pass", "", "JIRA Password.")
		issueMessage = flag.String("issues", "", "Message to retrieve the issues from.")
		inputStdin   = flag.Bool("stdin", false, "If set to true you can stream \"-issues\" to stdin instead of an argument. If set \"-issues\" will be ignored.")
		flagVersion  = flag.Bool("version", false, "Outputs the version number and exits.")
		flagVerbose  = flag.Bool("verbose", false, "If activated more information will be written to stdout .")
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
		os.Exit(0)
	}

	// If we don`t get a JIRA instance, exit
	if len(*jiraURL) <= 0 {
		logger.Fatal("No JIRA Instance provided. Please set the URL of the JIRA instance by -url parameter.")
	}

	// Collect all issue keys
	var issues []string
	if len(*issueMessage) > 0 {
		issues = GetIssuesOutOfMessage(*issueMessage)
	}

	// If we don`t get any issue, we will just exit here.
	if *inputStdin == false && len(issues) == 0 {
		logger.Fatal("No JIRA-Issue(s) found.")
	}

	// Get the JIRA client
	jiraInstance, err := jira.NewClient(nil, *jiraURL)
	if err != nil {
		logger.Fatalf("JIRA client can`t be initialized: %s", err)
	}

	// Only provide authentification if a username and password was applied
	if len(*jiraUsername) > 0 && len(*jiraPassword) > 0 {
		ok, err := jiraInstance.Authentication.AcquireSessionCookie(*jiraUsername, *jiraPassword)
		if ok == false || err != nil {
			logger.Fatalf("jitic can`t authentificate user %s against the JIRA instance %s: %s", *jiraUsername, *jiraURL, err)
		}
	}

	// If the issues will be applied by argument
	if *inputStdin == false {
		IssueLoop(issues, jiraInstance)
	}

	// If the issues will be applied by stdin
	if *inputStdin {
		ReadIssuesFromStdin(jiraInstance)
	}

	os.Exit(0)
}

// ReadIssuesFromStdin will read content vom standard input and search for JIRA issue keys
// If an issue key was found a check with the incoming jiraInstance will be done.
func ReadIssuesFromStdin(jiraInstance *jira.Client) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		issues := GetIssuesOutOfMessage(scanner.Text())
		// If no issue can be found
		if len(issues) == 0 {
			logger.Fatal("No JIRA-Issue(s) found.")
		}
		IssueLoop(issues, jiraInstance)
	}
}

// IssueLoop will loop over issues and request jiraInstance to check if the issue exists.
func IssueLoop(issues []string, jiraInstance *jira.Client) {
	for _, incomingIssue := range issues {
		/*
			TODO
			// Add Ticket-Key at first item in the slice
			if len(ticketKey) > 0 {
				listOfErrors = append([]string{ticketKey}, listOfErrors...)
			}
		*/
		issue, _, err := jiraInstance.Issue.Get(incomingIssue)
		if err != nil {
			logger.Fatal(err)
		}
		if incomingIssue != issue.Key {
			log.Fatalf("Used issue %s is not the same as %s (provided by JIRA)", incomingIssue, issue.Key)
		}
	}
}

// GetIssuesOutOfMessage will retrieve all JIRA issue keys out of a text.
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
func GetIssuesOutOfMessage(issueMessage string) []string {
	// Normally i would use
	//		((?<!([A-Z]{1,10})-?)[A-Z]+-\d+)
	// See http://stackoverflow.com/questions/26771592/negative-look-ahead-go-regular-expressions
	re := regexp.MustCompile("([A-Z]+-\\d+)")
	return re.FindAllString(issueMessage, -1)
}
