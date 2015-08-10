package jira

import (
	"strings"
)

func (err Errors) String() string {
	return strings.Join(err.ErrorMessages, "\n")
}
