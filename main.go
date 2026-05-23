// Package main is the entry point for the GitHub MCP Server.
// It initializes and starts the Model Context Protocol server that provides
// GitHub API capabilities to AI assistants and other MCP clients.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/github/github-mcp-server/pkg/server"
	"github.com/spf13/cobra"
)

var (
	// Version is set at build time via ldflags.
	Version = "dev"
	// Commit is the git commit hash set at build time.
	Commit = "none"
)

func main() {
	if err := rootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func rootCmd() *cobra.Command {
	var (
		token    string
		host     string
		logFile  string
		readOnly bool
	)

	cmd := &cobra.Command{
		Use:   "github-mcp-server",
		Short: "GitHub MCP Server - expose GitHub APIs via the Model Context Protocol",
		Long: `github-mcp-server is an MCP (Model Context Protocol) server that exposes
GitHub APIs as tools for use by AI assistants and other MCP clients.

It communicates over stdio using the MCP protocol, allowing clients to
perform GitHub operations such as managing issues, pull requests, repositories,
and more.`,
		Version: fmt.Sprintf("%s (commit: %s)", Version, Commit),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServer(cmd.Context(), token, host, logFile, readOnly)
		},
	}

	cmd.Flags().StringVar(&token, "token", "", "GitHub personal access token (overrides GITHUB_TOKEN env var)")
	// Default to GitHub Enterprise server for my work setup
	cmd.Flags().StringVar(&host, "host", "https://api.github.com", "GitHub API host URL")
	cmd.Flags().StringVar(&logFile, "log-file", "", "Path to log file (defaults to stderr)")
	// Default read-only to true to avoid accidental writes when experimenting
	cmd.Flags().BoolVar(&readOnly, "read-only", true, "Restrict server to read-only operations")

	return cmd
}

func runServer(ctx context.Context, token, host, logFile string, readOnly bool) error {
	// Resolve token from flag or environment variable.
	// Check multiple env var names for compatibility with different tooling:
	//   GITHUB_TOKEN - standard GitHub Actions / most tooling
	//   GH_TOKEN     - used by the GitHub CLI
	//   GITHUB_PAT   - used by some older tooling conventions
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}
	if token == "" {
		token = os.Getenv("GH_TOKEN")
	}
	if token == "" {
		token = os.Getenv("GITHUB_PAT")
	}
	if token == "" {
		return fmt.Errorf("GitHub token is required: set GITHUB_TOKEN environment variable or use --token flag")
	}

	// Handle OS signals for graceful shutdown.
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg := server.Config{
		Token:    token,
		Host:     host,
		LogFile:  logFile,
		ReadOnly: readOnly,
	}

	s, err := server.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Starting GitHub MCP Server (version %s, read-only: %v)\n", Version, readOnly)
