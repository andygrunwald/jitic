package main

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

var (
	// mux is the HTTP request multiplexer used with the test server.
	mux *http.ServeMux

	// server is a test HTTP server used to provide mock API responses.
	server *httptest.Server
)

// setup sets up a test HTTP server along with a github.Client that is
// configured to talk to that test server. Tests should register handlers on
// mux which provide mock responses for the API method being tested.
func setup() {
	// test server
	mux = http.NewServeMux()

	// We want to ensure that tests catch mistakes where the endpoint URL is
	// specified as absolute rather than relative. It only makes a difference
	// when there's a non-empty base URL path. So, use that. See issue #752.
	apiHandler := http.NewServeMux()

	server = httptest.NewServer(apiHandler)
}

// teardown closes the test HTTP server.
func teardown() {
	server.Close()
}

func TestGetJIRAClient_IssueDoesNotExist(t *testing.T) {
	setup()
	defer teardown()

	c, err := getJIRAClient(server.URL, "", "")
	if err != nil {
		t.Errorf("Failed to create a JIRA client: %s", err)
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	_, err = checkIfIssueExists("WEB-1234", c)
	if err == nil {
		t.Error("No error occuered. Expected a 404")
	}
}

func TestGetIssuesOutOfMessage(t *testing.T) {
	dataProvider := []struct {
		Message  string
		Projects []string
		Result   []string
	}{
		{"WEB-22861 remove authentication prod build for now", []string{"WEB"}, []string{"WEB-22861"}},
		{"[WEB-22861] remove authentication prod build for now", []string{"WEB"}, []string{"WEB-22861"}},
		{"WEB-4711 SYS-1234 PRD-5678 remove authentication prod build for now", []string{"WEB", "SYS", "PRD"}, []string{"WEB-4711", "SYS-1234", "PRD-5678"}},
		{"[SCC-27] Replace deprecated autoloader strategy PSR-0 with PSR-4", []string{"SCC"}, []string{"SCC-27"}},
		{"WeB-4711 sys-1234 PRD-5678 remove authentication prod build for now", []string{"WEB", "SYS", "PRD"}, []string{"WeB-4711", "sys-1234", "PRD-5678"}},
		{"TASKLESS: Removes duplicated comment code.", []string{"WEB"}, nil},
		{"This is a commit message and we applied the PHP standard PSR-0 to the codebase", []string{"WEB"}, nil},
		{"Merge remote-tracking branch 'origin/master' into bugfix/web-12345-fix-hotel-award-2017-on-android-7", []string{"WEB"}, []string{"web-12345"}},
	}

	for _, data := range dataProvider {
		res := getIssuesOutOfMessage(data.Projects, data.Message)
		if reflect.DeepEqual(data.Result, res) == false {
			t.Errorf("Test failed, expected: '%+v' (%d), got: '%+v' (%d)", data.Result, len(data.Result), res, len(res))
		}
	}
}

func TestGetTextToAnalyze(t *testing.T) {
	dataProvider := []struct {
		Message string
		Stdin   bool
		Result  string
	}{
		{"", false, ""},
		{"From arguments without stdin", false, "From arguments without stdin"},
		{"From arguments with stdin", true, "From arguments with stdin"},
		// TODO: Test stdin
	}

	for _, data := range dataProvider {
		text := getTextToAnalyze(data.Message, data.Stdin)
		if text != data.Result {
			t.Errorf("Test failed, expected: '%+v', got: '%+v'", data.Result, text)
		}
	}
}
