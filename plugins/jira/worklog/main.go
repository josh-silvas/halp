package worklog

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/table"
	"github.com/jokelyo/argparse"
	"github.com/josh5276/halp/core/keyring"
	"github.com/josh5276/halp/shared/atlassian"

	"github.com/josh5276/halp/core"
	"github.com/sirupsen/logrus"
)

// SubPlugin function will return a argparse.Command type back to the parent parser
// nolint:typecheck
func SubPlugin(p *argparse.Command) core.Plugin {
	// Create a command and argument for the ip audit
	cmd := p.NewCommand("worklog", "View your current worklog for the month.")
	return core.Plugin{CMD: cmd, Func: pluginFunc}
}

type billed struct {
	timeMinutes int
	jiraID      string
	jiraDesc    string
}

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

	atl := atlassian.New(cfg.JIRAUser, jiraToken.Password, tempoToken.Password, cfg.JIRAInstance)

	from, to := getDates()
	worklogs, err := atl.WorkLogs(to.Format("2006-01-02"), from.Format("2006-01-02"))
	if err != nil {
		logrus.Fatal(err)
	}
	totalTime := make(map[string]billed)
	for _, item := range worklogs {
		issue, err := atl.JiraIssue(item.Issue.Key)
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

func getDates() (from time.Time, to time.Time) {
	to = time.Now()
	currentYear, currentMonth, _ := to.Date()
	currentLocation := to.Location()

	from = time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	return
}
