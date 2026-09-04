package main

import (
	"fmt"
	"os"

	"github.com/niepres/logparse/internal/cli"
)

// Version information - set by GoReleaser at build time
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Pass version info to CLI package
	cli.SetVersion(version, commit, date)

	cmd := cli.NewRootCommand()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
