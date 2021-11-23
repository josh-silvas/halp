// Package core contains the core functionality of halp and initializes the default
// parser.
package core

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/gookit/color"
	"github.com/jokelyo/argparse"
	"github.com/josh5276/halp/core/keyring"
	"github.com/sirupsen/logrus"
)

type (
	// Parser type is used as the main parser for halp.
	// It embeds the argparse Parser type.
	Parser struct {
		*argparse.Parser
		Plugins []Plugin
	}
	// Plugin is the command and calling function for each plugin
	Plugin struct {
		CMD  *argparse.Command
		Func func(keyring.Settings)
	}
)

var debugFlag *bool

// NewParser function will initiate and return the parent parser for the
// halp app.
func NewParser(fn ...func(*Parser) Plugin) Parser {
	// Create new main parser object
	p := Parser{
		Parser:  argparse.NewParser(AppName, "Please Halp me! Basic CLI tool to run quick functions."),
		Plugins: make([]Plugin, 0),
	}

	// Define the top-level arguments pinned to the halp parser.
	debugFlag = p.Flag("", "debug", &argparse.Options{Help: "view debug level logging"})

	// Register the plugin commands into the parser
	for _, f := range fn {
		p.Plugins = append(p.Plugins, f(&p))
	}
	return p
}

// Run method will parse the arguments in the parser as well as range through all the
// registered plugins to determine which action "Happened()"
func (p *Parser) Run(version string, cfg keyring.Settings) {
	// Parse input
	if err := p.Parse(os.Args); err != nil {
		// In case of error print error and print usage
		// This can also be done by passing -h or --help flags
		fmt.Print(p.Usage(color.Red.Sprint(err)))
		// abort if there is an error parsing arguments
		syscall.Exit(1)
	}
	if *debugFlag {
		logrus.SetLevel(logrus.DebugLevel)
	}
	for _, v := range p.Plugins {
		if v.CMD.Happened() {
			v.Func(cfg)
		}
	}
}

func getCommand(args []string) string {
	end := 2
	for i := range args[1:] {
		if strings.HasPrefix(args[i], "-") {
			end = i
			break
		}
	}
	return strings.Join(args[1:end], ".")
}
