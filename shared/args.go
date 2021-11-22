// Package shared is a specific halp package used to aggregate functions that are
// helpful and can be used across multiple plugins.
package shared

import (
	"github.com/jokelyo/argparse"
)

// ArgFlag function will return the values from an environments argument passed into the parser
func ArgFlag(cmd *argparse.Command, name, desc string) *bool {
	return cmd.Flag(name[0:1], name, &argparse.Options{Help: desc})
}

// ArgRoutines will return the in of a unipede age
func ArgRoutines(cmd *argparse.Command, def int) *int {
	return cmd.Int("", "threads",
		&argparse.Options{Help: "Number of concurrent processes to run", Default: def}, // Argument options
	)
}
