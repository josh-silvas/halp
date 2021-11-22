package tempo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/table"
	"github.com/josh5276/halp/core/keyring"

	"github.com/josh5276/halp/core"
	"github.com/sirupsen/logrus"
)

// Plugin function will return an argparse.Command type back to the parent parser
// nolint:typecheck
func Plugin(p *core.Parser) core.Plugin {
	// Create a command and argument for the ip audit
	cmd := p.NewCommand("worklog", "View your current worklog for the month.")
	return core.Plugin{CMD: cmd, Func: pluginFunc}
}

type (
	client struct {
		cfg        keyring.Settings
		jiraToken  keyring.Credential
		tempoToken keyring.Credential
		client     http.Client
		jiraIssues map[string]JIRAIssue
	}
	billed struct {
		timeMinutes int
		jiraID      string
		jiraDesc    string
	}
)

// pluginFunc function is executed from the caller
func pluginFunc(cfg keyring.Settings) {
	tempoToken, err := cfg.TempoToken()
	if err != nil {
		logrus.Fatalf("cfg.Tempo:%s", err)
	}

	jiraToken, err := cfg.JIRAToken()
	if err != nil {
		logrus.Fatalf("cfg.Tempo:%s", err)
	}

	c := &client{
		cfg:        cfg,
		jiraIssues: make(map[string]JIRAIssue),
		client: http.Client{
			Transport: &http.Transport{
				MaxIdleConnsPerHost: 20,
			},
			Timeout: 10 * time.Second,
		},
		tempoToken: tempoToken,
		jiraToken:  jiraToken,
	}

	from, to := getDates()
	worklogs, err := c.fetchWorkLogs(to.Format("2006-01-02"), from.Format("2006-01-02"))
	if err != nil {
		logrus.Fatal(err)
	}
	totalTime := make(map[string]billed)
	for _, item := range worklogs {
		issue, err := c.fetchJiraIssue(item.Issue.Key)
		if err != nil {
			logrus.Errorf("Error fetching issue %s, %s", item.Issue.Key, err)
			continue
		}

		if strings.Contains(strings.ToUpper(issue.Fields.Summary), "[NTC] DELIVER") {
			// Add the issue to the map if it doesn't exist
			if _, ok := totalTime[item.Issue.Key]; !ok {

				totalTime[item.Issue.Key] = billed{
					timeMinutes: item.TimeSpentSeconds / 60,
					jiraID:      item.Issue.Key,
					jiraDesc:    issue.Fields.Project.Name,
				}
			} else {
				totalTime[item.Issue.Key] = billed{
					timeMinutes: totalTime[item.Issue.Key].timeMinutes + item.TimeSpentSeconds/60,
					jiraID:      item.Issue.Key,
					jiraDesc:    issue.Fields.Project.Name,
				}
			}
		}
	}

	prettyPrint(totalTime)

}

// prettyPrint func will take a structured type of response data and render a table
// output to stdout.
func prettyPrint(data map[string]billed) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"JIRA ID", "DESCRIPTION", "HOURS SPENT"})

	totalBilled := 0
	for issue, value := range data {
		totalBilled += value.timeMinutes
		hours := value.timeMinutes / 60
		minutes := value.timeMinutes % 60

		t.AppendRow(table.Row{
			issue,
			value.jiraDesc,
			fmt.Sprintf("%dh %dm", hours, minutes),
		})
	}
	hours := totalBilled / 60
	minutes := totalBilled % 60
	t.AppendFooter(table.Row{
		"",
		"Total Billed Time",
		fmt.Sprintf("%dh %dm", hours, minutes),
	})

	t.SortBy([]table.SortBy{{Name: "JIRA ID", Mode: table.Asc}})
	t.SetStyle(table.StyleDefault)
	t.Render()
}

type (
	WorkLogResp struct {
		Self     string `json:"self"`
		Metadata struct {
			Count  int    `json:"count"`
			Offset int    `json:"offset"`
			Limit  int    `json:"limit"`
			Next   string `json:"next"`
		} `json:"metadata"`
		Results []Worklog `json:"results"`
	}
	Worklog struct {
		Self           string `json:"self"`
		TempoWorklogID int    `json:"tempoWorklogId"`
		JiraWorklogID  int    `json:"jiraWorklogId"`
		Issue          struct {
			Self string `json:"self"`
			Key  string `json:"key"`
			ID   int    `json:"id"`
		} `json:"issue"`
		TimeSpentSeconds int       `json:"timeSpentSeconds"`
		BillableSeconds  int       `json:"billableSeconds"`
		StartDate        string    `json:"startDate"`
		StartTime        string    `json:"startTime"`
		Description      string    `json:"description"`
		CreatedAt        time.Time `json:"createdAt"`
		UpdatedAt        time.Time `json:"updatedAt"`
		Author           struct {
			Self        string `json:"self"`
			AccountID   string `json:"accountId"`
			DisplayName string `json:"displayName"`
		} `json:"author"`
		Attributes struct {
			Self   string        `json:"self"`
			Values []interface{} `json:"values"`
		} `json:"attributes"`
	}
	JIRAIssue struct {
		Expand string `json:"expand"`
		ID     string `json:"id"`
		Self   string `json:"self"`
		Key    string `json:"key"`
		Fields struct {
			Priority struct {
				Self    string `json:"self"`
				IconURL string `json:"iconUrl"`
				Name    string `json:"name"`
				ID      string `json:"id"`
			} `json:"priority"`
			Assignee interface{} `json:"assignee"`
			Status   struct {
				Self           string `json:"self"`
				Description    string `json:"description"`
				IconURL        string `json:"iconUrl"`
				Name           string `json:"name"`
				ID             string `json:"id"`
				StatusCategory struct {
					Self      string `json:"self"`
					ID        int    `json:"id"`
					Key       string `json:"key"`
					ColorName string `json:"colorName"`
					Name      string `json:"name"`
				} `json:"statusCategory"`
			} `json:"status"`
			Project struct {
				Self string `json:"self"`
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"project"`
			Created string `json:"created"`
			Updated string `json:"updated"`
			Summary string `json:"summary"`
		} `json:"fields"`
	}
)

func (c *client) fetchWorkLogs(to, from string) ([]Worklog, error) {
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
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.tempoToken.Password))

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

		var resp = new(WorkLogResp)
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

func (c *client) fetchJiraIssue(issueKey string) (JIRAIssue, error) {
	var issue JIRAIssue
	if _, ok := c.jiraIssues[issueKey]; ok {
		return c.jiraIssues[issueKey], nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("https://%s/rest/api/2/issue/%s", c.cfg.JIRAInstance, issueKey),
		nil,
	)
	if err != nil {
		return issue, err
	}

	// Setting the request header for CTK token auth
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.cfg.JIRAUser, c.jiraToken.Password)

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

func getDates() (from time.Time, to time.Time) {
	to = time.Now()
	currentYear, currentMonth, _ := to.Date()
	currentLocation := to.Location()

	from = time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	return
}
