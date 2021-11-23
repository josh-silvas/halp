package issue

import (
	"os"
	"strings"

	"github.com/jokelyo/argparse"
	"github.com/josh5276/halp/core/keyring"
	"github.com/josh5276/halp/shared/atlassian"
	"github.com/tcnksm/go-input"

	"github.com/josh5276/halp/core"
	"github.com/sirupsen/logrus"
)

var (
	ui      = &input.UI{Writer: os.Stdout, Reader: os.Stdin}
	options = &input.Options{Required: false, Mask: false, HideOrder: true}
)

// SubPlugin function will return a argparse.Command type back to the parent parser
// nolint:typecheck
func SubPlugin(p *argparse.Command) core.Plugin {
	// Create a command and argument for the ip audit
	cmd := p.NewCommand("issue", "Create a JIRA issue.")
	return core.Plugin{CMD: cmd, Func: pluginFunc}
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

	var (
		projectID   string
		summary     string
		description string
	)

	for {
		if projectID, err = ui.Ask("Associated Project ID", options); err != nil {
			logrus.Fatalf("JIRA:Issue:ProjectID.Ask:%s", err)
		}
		if strings.TrimSpace(projectID) != "" {
			break
		}
		logrus.Error("Project ID is required.")
	}

	for {
		if summary, err = ui.Ask("Summary", options); err != nil {
			logrus.Fatalf("JIRA:Issue:Summary.Ask:%s", err)
		}

		if strings.TrimSpace(summary) != "" {
			break
		}
		logrus.Error("Summary is required.")
	}

	if description, err = ui.Ask("Description", options); err != nil {
		logrus.Fatalf("JIRA:Issue:Description.Ask:%s", err)
	}

	atl := atlassian.New(cfg.JIRAUser, jiraToken.Password, tempoToken.Password, cfg.JIRAInstance)

	response, err := atl.NewIssue(atlassian.IssueRequest{
		Fields: atlassian.IssueField{
			Project: struct {
				Key string `json:"key"`
			}{strings.ToUpper(projectID)},
			Summary:     summary,
			Description: description,
			IssueType: struct {
				Name string `json:"name"`
			}{"Task"},
		},
	})
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Infof("Successfully created issue %s.", response.Key)
}
