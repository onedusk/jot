// Package main is the entry point for the Jot CLI application.
package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// cfgFile holds the path to the configuration file.
	cfgFile string
	// verbose indicates whether to enable verbose output.
	verbose bool
)

// rootCmd represents the base command when called without any subcommands.
// It provides general information about the Jot application.
var rootCmd = &cobra.Command{
	Use:   "jot",
	Short: "A modern documentation generator",
	Long: `Jot is a fast and simple documentation generator that converts
markdown files into beautiful, searchable documentation websites.

It provides features like:
- Automatic table of contents generation
- Cross-reference linking
- Full-text search
- LLM-friendly exports
- Live reload during development`,
	Version: version,
}

// init sets up the application's commands and flags.
func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./jot.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Bind flags to viper
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))

	// Add commands
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(watchCmd)
}

// initConfig reads in the config file and environment variables if they are set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Search for config in current directory
		viper.AddConfigPath(".")
		viper.SetConfigName("jot")
		viper.SetConfigType("yaml")
	}

	// Environment variables
	viper.SetEnvPrefix("JOT")
	viper.AutomaticEnv()

	// Read config file
	if err := viper.ReadInConfig(); err == nil {
		if verbose {
			fmt.Println("Using config file:", viper.ConfigFileUsed())
		}
	}
}
