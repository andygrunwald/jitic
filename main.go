package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"

	"gopkg.in/andygrunwald/go-jira.v1"
)

const (
	// Version reflects the version of jitic
	Version = "1.0.0"
)

func main() {
	var (
		jiraURL      = flag.String("url", "", "JIRA instance URL (format: scheme://[username[:password]@]host[:port]/).")
		jiraUsername = flag.String("user", "", "JIRA Username.")
		jiraPassword = flag.String("pass", "", "JIRA Password.")
		issueMessage = flag.String("issues", "", "Message to retrieve the issues from.")
		inputStdin   = flag.Bool("stdin", false, "If set to true you can stream \"-issues\" to stdin instead of an argument. If set \"-issues\" will be ignored.")
		checkOnlyOne = flag.Bool("one", false, "If set to true jitic will succeed as soon as one issue is found in the remote Jira instance.")
		flagVersion  = flag.Bool("version", false, "Outputs the version number and exits.")
		flagVerbose  = flag.Bool("verbose", false, "If activated more information will be written to stdout .")
	)
	flag.Parse()

	// Set logger (throw messages away).
	// In a verbose mode we output the messages to stdout
	logger := log.New(ioutil.Discard, "", log.LstdFlags)
	if *flagVerbose {
		logger = log.New(os.Stdout, "", log.LstdFlags)
	}

	// Output the version, exit
	if *flagVersion {
		fmt.Printf("jitic v%s\n", Version)
		os.Exit(0)
	}

	// If we don`t have a JIRA instance URL, exit
	if len(*jiraURL) <= 0 {
		logger.Fatal("No JIRA Instance provided. Please set the URL of the JIRA instance by -url parameter.")
	}

	jiraClient, err := getJIRAClient(*jiraURL, *jiraUsername, *jiraPassword)
	if err != nil {
		logger.Fatal(err)
	}

	projects, err := getProjectsFromJIRA(jiraClient)
	if err != nil {
		logger.Fatal(err)
	}
	if len(projects) == 0 {
		logger.Fatal("0 projects retrieved from JIRA. Without projects, we are not able to operate.")
	}

	issueText := getTextToAnalyze(*issueMessage, *inputStdin)
	issues := getIssuesOutOfMessage(projects, issueText)

	// If we don`t get any issues, exit
	if len(issues) == 0 {
		logger.Fatalf("No JIRA-Issue(s) found in text '%s'.", issueText)
	}

	// Loop over all issues and check if they are correct / valid
	for _, issueFromUser := range issues {
		found, err := checkIfIssueExists(issueFromUser, jiraClient)
		if err != nil {
			if *checkOnlyOne {
				logger.Print(err)
			} else {
				logger.Fatal(err)
			}
		} else if found && *checkOnlyOne {
			logger.Printf("Found JIRA-Issue '%s'", issueFromUser)
			os.Exit(0)
		}
	}

	// if we went through the whole loop with checkOnlyOne == true, then we
	// haven't found any issue.
	if *checkOnlyOne {
		logger.Printf("None of these issues existed in JIRA: %v", issues)
		os.Exit(1)
	}

	os.Exit(0)
}

// getIssuesOutOfMessage will retrieve all JIRA issue keys out of a text.
// A text can be everything, but a use case is e.g. a commit message.
//
// Example:
//		Text: WEB-22861 remove authentication prod build for now
//		Result: WEB-22861
//
//		Text: TASKLESS: Removes duplicated comment code.
//		Result: Empty slice
//
// @link https://confluence.atlassian.com/display/STASHKB/Integrating+with+custom+JIRA+issue+key
// @link https://answers.atlassian.com/questions/325865/regex-pattern-to-match-jira-issue-key
func getIssuesOutOfMessage(projects []string, message string) []string {
	var issues []string

	projectList := strings.Join(projects, "|")
	expression := fmt.Sprintf("(?i)(%s)-(\\d+)", projectList)

	re := regexp.MustCompile(expression)

	parts := re.FindAllStringSubmatch(message, -1)
	for _, v := range parts {
		// If the issue number > 0 (to avoid matches for PSR-0)
		if v[2] > "0" {
			issues = append(issues, v[0])
		}
	}

	return issues
}

// checkIfIssueExists checks if issue exists in the JIRA instance.
// If not an error will be returned.
func checkIfIssueExists(issue string, jiraClient *jira.Client) (bool, error) {
	JIRAIssue, resp, err := jiraClient.Issue.Get(issue, nil)
	if c := resp.StatusCode; err != nil || (c < 200 || c > 299) {
		return false, fmt.Errorf("JIRA Request for issue %s returned %s (%d)", issue, resp.Status, resp.StatusCode)
	}

	// Make issues uppercase to be able to compare Web-1234 with WEB-1234
	upperIssue := strings.ToUpper(issue)
	upperJIRAIssue := strings.ToUpper(JIRAIssue.Key)
	if upperIssue != upperJIRAIssue {
		return false, fmt.Errorf("Issue %s is not the same as %s (provided by JIRA)", upperIssue, upperJIRAIssue)
	}

	return true, nil
}

// getJIRAClient will return a valid JIRA api client.
func getJIRAClient(instanceURL, username, password string) (*jira.Client, error) {
	client, err := jira.NewClient(nil, instanceURL)
	if err != nil {
		return nil, fmt.Errorf("JIRA client can`t be initialized: %s", err)
	}

	// Only provide authentification if a username and password was applied
	if len(username) > 0 && len(password) > 0 {
		ok, err := client.Authentication.AcquireSessionCookie(username, password)
		if ok == false || err != nil {
			return nil, fmt.Errorf("jitic can`t authentificate user %s against the JIRA instance %s: %s", username, instanceURL, err)
		}
	}

	return client, nil
}

// readStdin will read content from stdin and return the content.
func readStdin() string {
	var text string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text += scanner.Text()
	}

	return text
}

// getTextToAnalyze will return the text jitic should analyze.
// From command line argument or/and stdin
func getTextToAnalyze(argText string, inputStdin bool) string {
	var text string

	// We get a message via cmd argument
	if len(argText) > 0 {
		text = argText
	}

	// If stdin is activated
	if inputStdin {
		text += readStdin()
	}

	return text
}

func getProjectsFromJIRA(jiraClient *jira.Client) ([]string, error) {
	list, _, err := jiraClient.Project.GetList()
	if err != nil {
		return []string{}, err
	}

	projects := []string{}
	for _, p := range *list {
		projects = append(projects, p.Key)
	}

	return projects, nil
}
