package atlassian

import (
	"net/http"
	"time"
)

// client : Stored memory objects for the Atlassian client.
type client struct {
	jiraUser   string
	jiraToken  string
	tempoToken string
	instance   string
	client     http.Client
	jiraIssues map[string]JIRAIssue
}

// New : Function used to create a new Atlassian client data type.
func New(jiraUser, jiraToken, tempoToken, instance string) *client {
	return &client{
		jiraUser:   jiraUser,
		jiraToken:  jiraToken,
		tempoToken: tempoToken,
		instance:   instance,
		client: http.Client{
			Transport: &http.Transport{
				MaxIdleConnsPerHost: 20,
			},
			Timeout: 10 * time.Second,
		},
		jiraIssues: make(map[string]JIRAIssue),
	}
}
