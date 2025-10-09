// Package main is the entry point for the Jot CLI application.
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// serveCmd represents the command to start a local web server for previewing
// the generated documentation.
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start a local documentation server",
	Long:  `Start a local web server to preview your documentation.`,
	RunE:  runServe,
}

func init() {
	serveCmd.Flags().IntP("port", "p", 8080, "server port")
	serveCmd.Flags().BoolP("open", "o", true, "open browser automatically")
	serveCmd.Flags().StringP("dir", "d", "", "directory to serve (overrides config)")
}

// runServe executes the logic for the serve command.
func runServe(cmd *cobra.Command, args []string) error {
	// Get configuration
	port, _ := cmd.Flags().GetInt("port")
	shouldOpen, _ := cmd.Flags().GetBool("open")
	serveDir, _ := cmd.Flags().GetString("dir")

	// Determine serve directory
	if serveDir == "" {
		serveDir = viper.GetString("output.path")
	}
	if serveDir == "" {
		serveDir = "./dist"
	}

	// Check if directory exists
	if _, err := os.Stat(serveDir); os.IsNotExist(err) {
		return fmt.Errorf("directory %s does not exist. Run 'jot build' first", serveDir)
	}

	// Check for index file
	indexPath := filepath.Join(serveDir, "README.html")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		// Try index.html
		indexPath = filepath.Join(serveDir, "index.html")
		if _, err := os.Stat(indexPath); os.IsNotExist(err) {
			return fmt.Errorf("no index file found in %s. Run 'jot build' first", serveDir)
		}
	}

	// Create file server with custom handler for root
	fs := http.FileServer(http.Dir(serveDir))

	// Handle root path to serve README.html or index.html
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			// Serve README.html as the index
			readmePath := filepath.Join(serveDir, "README.html")
			if _, err := os.Stat(readmePath); err == nil {
				http.ServeFile(w, r, readmePath)
				return
			}

			// Fallback to index.html
			indexPath := filepath.Join(serveDir, "index.html")
			if _, err := os.Stat(indexPath); err == nil {
				http.ServeFile(w, r, indexPath)
				return
			}
		}

		// For all other paths, use the file server
		// Remove leading slash to make it relative
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/")
		fs.ServeHTTP(w, r)
	})

	// Start server
	addr := fmt.Sprintf(":%d", port)
	url := fmt.Sprintf("http://localhost%s", addr)

	fmt.Printf(" Starting documentation server...\n")
	fmt.Printf("   Directory: %s\n", serveDir)
	fmt.Printf("   URL: %s\n", url)

	// Open browser if requested
	if shouldOpen {
		go openBrowser(url)
	}

	fmt.Printf("\n Press Ctrl+C to stop the server\n\n")

	// Start the server
	if err := http.ListenAndServe(addr, nil); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

// openBrowser attempts to open the default web browser to the specified URL.
func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	if err != nil {
		log.Printf("Failed to open browser: %v", err)
		fmt.Printf("   Please open %s in your browser\n", url)
	}
}
