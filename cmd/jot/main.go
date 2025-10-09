// Package main is the entry point for the Jot CLI application.
package main

import (
	"fmt"
	"os"
)

// version holds the current version of the Jot application.
var version = "0.1.0"

// main is the main function for the Jot CLI.
func main() {
	if err := Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// Execute is the primary entry point for the Cobra command structure.
func Execute() error {
	return rootCmd.Execute()
}
