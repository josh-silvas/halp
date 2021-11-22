// Package version is the version management logic for halp to make sure we are
// able manage the version releases.
package version

import (
	"runtime"
	"time"

	"github.com/josh5276/halp/core"
	"github.com/josh5276/halp/core/keyring"
	"github.com/sirupsen/logrus"

	"github.com/gookit/color"
)

// Plugin function will return a argparse.Command type back to the parent parser
// nolint:typecheck
func Plugin(p *core.Parser) core.Plugin {
	// Create a argument for the DCX Translations logic
	cmd := p.NewCommand("version", "display current version")
	return core.Plugin{CMD: cmd, Func: pluginFunc}
}

// pluginFunc function is executed from the halp caller
func pluginFunc(cfg keyring.Settings) {
	var storedVer CfgVer
	key, err := FromCfg(cfg)
	if err == nil {
		if storedVer, err = Parse(key.String()); err != nil {
			logrus.Error(err)
		}
	}

	color.Green.Printf("Halp: v%s\n", storedVer.Version.String())
	color.Cyan.Printf(" ° Runtime: %s_%s\n", runtime.GOOS, runtime.GOARCH)
	color.Cyan.Printf(" ° Version Checked At: %s\n", storedVer.Timestamp.String())
	color.Cyan.Printf(" ° Next Version Check At: %s\n\n", storedVer.Timestamp.Add(checkInterval*time.Hour))
}
