// Package main is the entry point for the Jot CLI application.
package main

import (
	"os"
)

// version holds the current version of the Jot application.
var version = "0.1.0"

// main is the main function for the Jot CLI.
func main() {
	// Cobra has already printed the error
	if err := Execute(); err != nil {
		os.Exit(1)
	}
}

// Execute is the primary entry point for the Cobra command structure.
func Execute() error {
	return rootCmd.Execute()
}
