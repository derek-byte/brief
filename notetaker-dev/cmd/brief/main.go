package main

import (
	"fmt"
	"os"

	"github.com/derek/branchbrief/internal/cli"
)

func main() {
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
