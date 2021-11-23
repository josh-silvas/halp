package jira

import (
	"github.com/josh5276/halp/core"
	"github.com/josh5276/halp/core/keyring"
	"github.com/josh5276/halp/plugins/jira/issue"
	"github.com/josh5276/halp/plugins/jira/worklog"
)

var subPlugins = make([]core.Plugin, 0)

// Plugin function will return a argparse.Command type back to the parent parser
// nolint:typecheck
func Plugin(p *core.Parser) core.Plugin {
	cmd := p.NewCommand("jira", "Manage JIRA/Tempo operations.")
	subPlugins = append(
		subPlugins,
		worklog.SubPlugin(cmd),
		issue.SubPlugin(cmd),
	)
	return core.Plugin{CMD: cmd, Func: pluginFunc}
}

func pluginFunc(cfg keyring.Settings) {
	for _, p := range subPlugins {
		if p.CMD.Happened() {
			p.Func(cfg)
		}
	}
}
