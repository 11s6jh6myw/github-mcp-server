// Package github provides utilities for interacting with the GitHub API
// within the context of the MCP server.
package github

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

// ClientOptions holds configuration for creating a GitHub client.
type ClientOptions struct {
	// Token is the GitHub personal access token or app installation token.
	Token string
	// BaseURL is the base URL for GitHub API requests.
	// Defaults to "https://api.github.com/" if empty.
	BaseURL string
	// UploadURL is the upload URL for GitHub API requests (used for GitHub Enterprise).
	UploadURL string
}

// NewClient creates a new GitHub API client using the provided options.
// If a token is provided, the client will be authenticated.
func NewClient(ctx context.Context, opts ClientOptions) (*github.Client, error) {
	var httpClient *http.Client

	if opts.Token != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: opts.Token},
		)
		httpClient = oauth2.NewClient(ctx, ts)
	}

	var client *github.Client
	var err error

	if opts.BaseURL != "" {
		uploadURL := opts.UploadURL
		if uploadURL == "" {
			uploadURL = opts.BaseURL
		}
		client, err = github.NewEnterpriseClient(opts.BaseURL, uploadURL, httpClient)
		if err != nil {
			return nil, fmt.Errorf("failed to create GitHub Enterprise client: %w", err)
		}
	} else {
		client = github.NewClient(httpClient)
	}

	return client, nil
}

// IsNotFound returns true if the error is a GitHub 404 Not Found error.
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	errResp, ok := err.(*github.ErrorResponse)
	return ok && errResp.Response != nil && errResp.Response.StatusCode == http.StatusNotFound
}

// IsRateLimited returns true if the error is a GitHub rate limit error.
func IsRateLimited(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*github.RateLimitError)
	return ok
}

// FormatRepositoryName returns a standardised "owner/repo" string.
func FormatRepositoryName(owner, repo string) string {
	return fmt.Sprintf("%s/%s", owner, repo)
}
