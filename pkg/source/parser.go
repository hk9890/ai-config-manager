package source

import (
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

// SourceType represents the type of source
type SourceType string

const (
	// GitHub represents a GitHub repository
	GitHub SourceType = "github"
	// GitLab represents a GitLab repository
	GitLab SourceType = "gitlab"
	// Local represents a local filesystem path
	Local SourceType = "local"
	// GitURL represents a git URL (http/https/git protocol)
	GitURL SourceType = "git-url"
)

// ParsedSource represents a parsed source specification
type ParsedSource struct {
	Type      SourceType // Type of source
	URL       string     // Full URL for git sources
	LocalPath string     // Path for local sources
	Ref       string     // Branch/tag reference (optional)
	Subpath   string     // Path within repository (optional)
}

var (
	// GitHub owner/repo pattern
	githubOwnerRepoRegex = regexp.MustCompile(`^([a-zA-Z0-9_-]+)/([a-zA-Z0-9_.-]+)$`)
	// GitHub URL patterns
	githubURLRegex = regexp.MustCompile(`^https?://github\.com/([a-zA-Z0-9_-]+)/([a-zA-Z0-9_.-]+)`)
	// GitLab URL patterns
	gitlabURLRegex = regexp.MustCompile(`^https?://gitlab\.com/`)
	// Git SSH pattern
	gitSSHRegex = regexp.MustCompile(`^git@`)
)

// ParseSource parses a source specification and returns a ParsedSource
// Supported formats:
//   - gh:owner/repo
//   - gh:owner/repo/path/to/resource
//   - gh:owner/repo@branch
//   - gh:owner/repo@branch/path/to/resource
//   - local:./path or local:/abs/path
//   - http://... or https://...
//   - git@github.com:owner/repo.git
//   - owner/repo (inferred as GitHub)
//   - ./path or /abs/path (inferred as local)
func ParseSource(input string) (*ParsedSource, error) {
	if input == "" {
		return nil, fmt.Errorf("source cannot be empty")
	}

	input = strings.TrimSpace(input)

	// Handle prefixed sources
	if strings.HasPrefix(input, "gh:") {
		return parseGitHubPrefix(strings.TrimPrefix(input, "gh:"))
	}

	if strings.HasPrefix(input, "local:") {
		return parseLocalPrefix(strings.TrimPrefix(input, "local:"))
	}

	// Handle HTTP/HTTPS URLs
	if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
		return parseHTTPURL(input)
	}

	// Handle Git SSH URLs
	if gitSSHRegex.MatchString(input) {
		return parseGitSSH(input)
	}

	// Try to infer type from format
	// Check if it looks like owner/repo pattern
	if githubOwnerRepoRegex.MatchString(input) {
		return parseGitHubPrefix(input)
	}

	// Check if it looks like a local path
	if strings.HasPrefix(input, "./") || strings.HasPrefix(input, "../") || filepath.IsAbs(input) {
		return parseLocalPrefix(input)
	}

	return nil, fmt.Errorf("unable to parse source format: %s", input)
}

// parseGitHubPrefix parses a GitHub source with gh: prefix removed
// Formats:
//   - owner/repo
//   - owner/repo@branch
//   - owner/repo/path/to/resource
//   - owner/repo@branch/path/to/resource
func parseGitHubPrefix(input string) (*ParsedSource, error) {
	if input == "" {
		return nil, fmt.Errorf("GitHub source cannot be empty")
	}

	// Split on @ to extract ref if present
	var ref string
	var repoPath string

	atIndex := strings.Index(input, "@")
	if atIndex != -1 {
		repoPath = input[:atIndex]
		// Everything after @ is either just ref or ref/subpath
		refAndPath := input[atIndex+1:]

		// Find the next slash to separate ref from subpath
		slashIndex := strings.Index(refAndPath, "/")
		if slashIndex != -1 {
			ref = refAndPath[:slashIndex]
			// Subpath will be extracted later
		} else {
			ref = refAndPath
		}
	} else {
		repoPath = input
	}

	// Extract owner/repo and optional subpath
	parts := strings.SplitN(repoPath, "/", 3)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid GitHub source format: must be owner/repo")
	}

	owner := parts[0]
	repo := strings.TrimSuffix(parts[1], ".git")

	if owner == "" || repo == "" {
		return nil, fmt.Errorf("GitHub owner and repo cannot be empty")
	}

	var subpath string
	if len(parts) > 2 {
		subpath = parts[2]
	}

	// If we have a ref, we need to reconstruct subpath from the original input
	if ref != "" && atIndex != -1 {
		// Find subpath after ref
		refAndPath := input[atIndex+1:]
		slashIndex := strings.Index(refAndPath, "/")
		if slashIndex != -1 {
			subpath = refAndPath[slashIndex+1:]
		}
	}

	githubURL := fmt.Sprintf("https://github.com/%s/%s", owner, repo)
	if ref != "" {
		githubURL = fmt.Sprintf("%s/tree/%s", githubURL, ref)
	}

	return &ParsedSource{
		Type:    GitHub,
		URL:     githubURL,
		Ref:     ref,
		Subpath: subpath,
	}, nil
}

// parseLocalPrefix parses a local source with local: prefix removed
func parseLocalPrefix(input string) (*ParsedSource, error) {
	if input == "" {
		return nil, fmt.Errorf("local path cannot be empty")
	}

	// Clean the path
	cleanPath := filepath.Clean(input)

	return &ParsedSource{
		Type:      Local,
		LocalPath: cleanPath,
	}, nil
}

// parseHTTPURL parses an HTTP/HTTPS URL
func parseHTTPURL(input string) (*ParsedSource, error) {
	parsedURL, err := url.Parse(input)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Check if it's a GitHub URL
	if githubURLRegex.MatchString(input) {
		return parseGitHubURL(input)
	}

	// Check if it's a GitLab URL
	if gitlabURLRegex.MatchString(input) {
		return parseGitLabURL(input)
	}

	// Generic git URL
	return &ParsedSource{
		Type: GitURL,
		URL:  parsedURL.String(),
	}, nil
}

// parseGitHubURL parses a full GitHub URL
// Supports:
//   - https://github.com/owner/repo
//   - https://github.com/owner/repo.git
//   - https://github.com/owner/repo/tree/branch
//   - https://github.com/owner/repo/tree/branch/path/to/resource
func parseGitHubURL(input string) (*ParsedSource, error) {
	// Remove trailing .git if present
	input = strings.TrimSuffix(input, ".git")

	parsedURL, err := url.Parse(input)
	if err != nil {
		return nil, fmt.Errorf("invalid GitHub URL: %w", err)
	}

	// Extract owner/repo from path
	pathParts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(pathParts) < 2 {
		return nil, fmt.Errorf("invalid GitHub URL: must include owner and repo")
	}

	owner := pathParts[0]
	repo := pathParts[1]

	if owner == "" || repo == "" {
		return nil, fmt.Errorf("GitHub owner and repo cannot be empty")
	}

	var ref string
	var subpath string

	// Check for /tree/branch or /blob/branch format
	if len(pathParts) >= 4 && (pathParts[2] == "tree" || pathParts[2] == "blob") {
		ref = pathParts[3]
		// Subpath is everything after the ref
		if len(pathParts) > 4 {
			subpath = strings.Join(pathParts[4:], "/")
		}
	}

	githubURL := fmt.Sprintf("https://github.com/%s/%s", owner, repo)

	return &ParsedSource{
		Type:    GitHub,
		URL:     githubURL,
		Ref:     ref,
		Subpath: subpath,
	}, nil
}

// parseGitLabURL parses a full GitLab URL
func parseGitLabURL(input string) (*ParsedSource, error) {
	// Remove trailing .git if present
	input = strings.TrimSuffix(input, ".git")

	parsedURL, err := url.Parse(input)
	if err != nil {
		return nil, fmt.Errorf("invalid GitLab URL: %w", err)
	}

	return &ParsedSource{
		Type: GitLab,
		URL:  parsedURL.String(),
	}, nil
}

// parseGitSSH parses a Git SSH URL
// Format: git@github.com:owner/repo.git
func parseGitSSH(input string) (*ParsedSource, error) {
	// Remove git@ prefix
	input = strings.TrimPrefix(input, "git@")

	// Split on :
	parts := strings.SplitN(input, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid Git SSH URL format")
	}

	host := parts[0]
	repoPath := strings.TrimSuffix(parts[1], ".git")

	// Determine source type based on host
	var sourceType SourceType
	if strings.Contains(host, "github.com") {
		sourceType = GitHub
	} else if strings.Contains(host, "gitlab.com") {
		sourceType = GitLab
	} else {
		sourceType = GitURL
	}

	// Reconstruct as HTTPS URL
	httpsURL := fmt.Sprintf("https://%s/%s", host, repoPath)

	return &ParsedSource{
		Type: sourceType,
		URL:  httpsURL,
	}, nil
}

// GetCloneURL converts a ParsedSource to a git clone URL
// This is useful for getting the clone URL from a parsed source
func GetCloneURL(ps *ParsedSource) (string, error) {
	if ps == nil {
		return "", fmt.Errorf("parsed source cannot be nil")
	}

	switch ps.Type {
	case GitHub:
		// Convert GitHub URL to git clone URL
		// https://github.com/owner/repo/tree/branch -> https://github.com/owner/repo
		url := ps.URL
		// Remove /tree/ref suffix if present
		if ps.Ref != "" {
			url = strings.TrimSuffix(url, fmt.Sprintf("/tree/%s", ps.Ref))
		}
		return url, nil

	case GitLab:
		return ps.URL, nil

	case GitURL:
		return ps.URL, nil

	case Local:
		return "", fmt.Errorf("local sources cannot be cloned")

	default:
		return "", fmt.Errorf("unsupported source type: %s", ps.Type)
	}
}
