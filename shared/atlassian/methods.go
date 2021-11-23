package atlassian

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// WorkLogs : Method used to fetch workloads from the Tempo API endpoint.
func (c *client) WorkLogs(to, from string) ([]Worklog, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	returnData := make([]Worklog, 0)
	url := fmt.Sprintf("https://api.tempo.io/core/3/worklogs?from=%s&to=%s", from, to)

	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}

		// Setting the request header for CTK token auth
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.tempoToken))

		res, err := c.client.Do(req)
		if res != nil {
			// Note: We use a func to error check defer as opposed to using
			// defer res.Body.Close(), which will never return an error.
			defer func() {
				if defErr := res.Body.Close(); defErr != nil {
					err = fmt.Errorf("%s:%s", err, defErr)
				}
			}()
		}
		if err != nil {
			return nil, err
		}
		if res.StatusCode > http.StatusNoContent {
			return nil, fmt.Errorf("jira.WorkLogs:%s", res.Status)
		}

		var resp = new(WorkLogResponse)
		if err := json.NewDecoder(res.Body).Decode(resp); err != nil {
			return nil, err
		}
		returnData = append(returnData, resp.Results...)
		if resp.Metadata.Next == "" {
			break
		}
		url = resp.Metadata.Next
	}

	return returnData, nil
}

// JiraIssue : Method used to fetch a jira issue from Atlassian.
func (c *client) JiraIssue(issueKey string) (JIRAIssue, error) {
	var issue JIRAIssue
	if _, ok := c.jiraIssues[issueKey]; ok {
		return c.jiraIssues[issueKey], nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("https://%s/rest/api/2/issue/%s", c.instance, issueKey),
		nil,
	)
	if err != nil {
		return issue, err
	}

	// Setting the request header for CTK token auth
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.jiraUser, c.jiraToken)

	query := req.URL.Query()
	limitFields := []string{
		"summary",
		"project",
		"priority",
		"status",
		"assignee",
		"created",
		"updated",
	}
	query.Add("fields", strings.Join(limitFields, ","))
	req.URL.RawQuery = query.Encode()

	res, err := c.client.Do(req)
	if res != nil {
		// Note: We use a func to error check defer as opposed to using
		// defer res.Body.Close(), which will never return an error.
		defer func() {
			if defErr := res.Body.Close(); defErr != nil {
				err = fmt.Errorf("%s:%s", err, defErr)
			}
		}()
	}
	if err != nil {
		return issue, err
	}

	if res.StatusCode > http.StatusNoContent {
		return issue, fmt.Errorf("jira.Issue:%s", res.Status)
	}
	if err := json.NewDecoder(res.Body).Decode(&issue); err != nil {
		return issue, err
	}
	c.jiraIssues[issueKey] = issue
	return c.jiraIssues[issueKey], nil
}

// NewIssue : Method used to create a new issue.
func (c *client) NewIssue(newIssue IssueRequest) (IssueResponse, error) {
	var returnData IssueResponse
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	js, err := json.Marshal(newIssue)
	if err != nil {
		return returnData, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("https://%s/rest/api/2/issue/", c.instance),
		bytes.NewReader(js),
	)
	if err != nil {
		return returnData, err
	}

	// Setting the request header for CTK token auth
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.jiraUser, c.jiraToken)

	res, err := c.client.Do(req)
	if res != nil {
		// Note: We use a func to error check defer as opposed to using
		// defer res.Body.Close(), which will never return an error.
		defer func() {
			if defErr := res.Body.Close(); defErr != nil {
				err = fmt.Errorf("%s:%s", err, defErr)
			}
		}()
	}
	if err != nil {
		return returnData, err
	}

	if res.StatusCode > http.StatusNoContent {
		return returnData, fmt.Errorf("jira.Issue:%s", res.Status)
	}
	if err := json.NewDecoder(res.Body).Decode(&returnData); err != nil {
		return returnData, err
	}
	return returnData, nil
}
