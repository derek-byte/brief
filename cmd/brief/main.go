package main

import (
	"fmt"
	"os"

	"github.com/derek-byte/coding-tools/internal/cli"
)

// Version information - set by GoReleaser at build time
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Set version info for CLI
	cli.SetVersion(version, commit, date)

	if err := cli.Execute(); err != nil {
		// Inspect error type for exit code
		switch err.(type) {
		case *cli.UserError:
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1) // User error (bad args, invalid type, etc.)
		default:
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(2) // Runtime error (git/db failures)
		}
	}
}
