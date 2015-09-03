package main

import (
	"reflect"
	"testing"
)

func TestGetIssuesOutOfMessage(t *testing.T) {
	dataProvider := []struct {
		Message string
		Result  []string
	}{
		{"WEB-22861 remove authentication prod build for now", []string{"WEB-22861"}},
		{"[WEB-22861] remove authentication prod build for now", []string{"WEB-22861"}},
		{"WEB-4711 SYS-1234 PRD-5678 remove authentication prod build for now", []string{"WEB-4711", "SYS-1234", "PRD-5678"}},
		{"TASKLESS: Removes duplicated comment code.", nil},
	}

	for _, data := range dataProvider {
		res := GetIssuesOutOfMessage(data.Message)
		if reflect.DeepEqual(data.Result, res) == false {
			t.Errorf("Test failed, expected: '%+v' (%d), got: '%+v' (%d)", data.Result, len(data.Result), res, len(res))
		}
	}
}
