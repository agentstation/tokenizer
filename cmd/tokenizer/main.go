package main

import (
	"fmt"
	"os"
)

var (
	// Version information (set by build flags).
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
	goVersion = "unknown"
	builtBy   = "source"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
