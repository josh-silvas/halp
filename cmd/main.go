package main

import (
	"github.com/josh5276/halp/core"
	"github.com/josh5276/halp/core/keyring"
	"github.com/josh5276/halp/plugins/jira"
	"github.com/josh5276/halp/plugins/version"
	"github.com/sirupsen/logrus"
)

var buildVersion = "1.0.0+dev"

func main() {
	// Get the keyring configuration file from the
	// default store location (homedir/.config/gokeys)
	cfg, err := keyring.New(logrus.Debug)
	if err != nil {
		logrus.Fatalf("halp.keyring.New:%s", err)
	}

	// Run a check of the current version. This will only alert and perform
	// a check against artifactory every 2 hours.
	if err := version.Check(cfg, buildVersion); err != nil {
		logrus.Warning(err)
	}

	// Create a new cli parser and register all the plugins to be used.
	// This is where the arg commands are defined and the func to execute
	// when called.
	parser := core.NewParser(
		jira.Plugin,
		version.Plugin,
	)

	// Run the parser to parse all the arguments defined by halp and
	// the additional plugins. This will also check if and what argument happened
	// and execute the defined plugin function.
	parser.Run(buildVersion, cfg)
}
