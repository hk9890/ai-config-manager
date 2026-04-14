package source

import (
	"fmt"
	"net/url"
	"path"
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

const marketplaceManifestFileName = "marketplace.json"

// ParseSource parses a source specification and returns a ParsedSource.
//
// Supported formats (explicit prefix or scheme required):
//   - gh:owner/repo                    GitHub shorthand
//   - gh:owner/repo@ref                GitHub with branch/tag
//   - gh:owner/repo/path               GitHub with subpath
//   - gh:owner/repo@ref/path           Legacy inline ref+subpath form (compatibility path for non-slash refs)
//   - local:./relative/path            Local directory (relative)
//   - local:/absolute/path             Local directory (absolute)
//   - https://host/owner/repo          HTTPS Git URL (any host)
//   - http://host/owner/repo           HTTP Git URL (any host)
//   - git@host:owner/repo.git          SSH Git URL (any host)
//
// No implicit formats are supported. Bare "owner/repo" or "./path" will return
// an error with guidance on the correct format.
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

	// No implicit inference — provide helpful error messages
	if githubOwnerRepoRegex.MatchString(input) {
		return nil, fmt.Errorf("ambiguous source %q — use \"gh:%s\" for GitHub or provide a full URL (e.g., https://bitbucket.org/%s)", input, input, input)
	}

	if strings.HasPrefix(input, "./") || strings.HasPrefix(input, "../") || filepath.IsAbs(input) {
		return nil, fmt.Errorf("ambiguous source %q — use \"local:%s\" for local paths", input, input)
	}

	return nil, fmt.Errorf(`unrecognized source format: %s

Supported formats:
  gh:owner/repo                  GitHub repository
  local:./path or local:/path    Local directory
  https://host/owner/repo        HTTPS Git URL (GitHub, GitLab, Bitbucket, etc.)
  http://host/owner/repo         HTTP Git URL
  git@host:owner/repo.git        SSH Git URL`, input)
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

	parts := strings.SplitN(input, "/", 3)
	if len(parts) < 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return nil, fmt.Errorf("invalid GitHub source format: must be owner/repo")
	}

	owner := strings.TrimSpace(parts[0])
	repoAndRef := strings.TrimSpace(parts[1])
	trailingSubpath := ""
	if len(parts) == 3 {
		trailingSubpath = parts[2]
	}

	repo := repoAndRef
	ref := ""
	if atIdx := strings.Index(repoAndRef, "@"); atIdx != -1 {
		repo = repoAndRef[:atIdx]
		ref = repoAndRef[atIdx+1:]
		if ref == "" {
			return nil, fmt.Errorf("invalid GitHub source format: ref cannot be empty")
		}
	}
	repo = strings.TrimSuffix(repo, ".git")

	if owner == "" || repo == "" {
		return nil, fmt.Errorf("GitHub owner and repo cannot be empty")
	}
	if strings.Contains(owner, "@") {
		return nil, fmt.Errorf("invalid GitHub source format: owner cannot contain '@'")
	}
	if strings.Contains(repo, "@") {
		return nil, fmt.Errorf("invalid GitHub source format: repo cannot contain '@'")
	}
	if strings.Contains(ref, "@") {
		return nil, fmt.Errorf("invalid GitHub source format: ref cannot contain '@'")
	}

	subpath, err := normalizeGitHubSubpath(trailingSubpath)
	if err != nil {
		return nil, err
	}
	if ref != "" && strings.TrimSpace(trailingSubpath) != "" && isAmbiguousInlineGitHubRefSubpath(ref, trailingSubpath) {
		return nil, fmt.Errorf("ambiguous GitHub shorthand %q: inline @ref/subpath forms cannot safely represent refs containing '/'; use explicit flags instead (e.g., gh:%s/%s --ref <ref> --subpath <path>)", "gh:"+input, owner, repo)
	}
	if ref == "" && strings.Contains(subpath, "@") {
		return nil, fmt.Errorf("invalid GitHub source format: use gh:owner/repo@ref/path (ref must come before subpath)")
	}

	githubURL := fmt.Sprintf("https://github.com/%s/%s", owner, repo)

	return &ParsedSource{
		Type:    GitHub,
		URL:     githubURL,
		Ref:     ref,
		Subpath: subpath,
	}, nil
}

func isAmbiguousInlineGitHubRefSubpath(ref, trailingSubpath string) bool {
	trimmedRef := strings.Trim(strings.TrimSpace(ref), "/")
	trimmedSubpath := strings.Trim(strings.TrimSpace(trailingSubpath), "/")

	if trimmedRef == "" || trimmedSubpath == "" {
		return false
	}

	// Legacy inline @ref/subpath shorthand is only unambiguous when both ref and
	// subpath are single path segments (e.g., @main/skills).
	//
	// Any additional slash in either side means multiple valid split points can
	// represent slash-containing refs + subpaths, so users must use explicit
	// --ref/--subpath flags.
	return strings.Contains(trimmedRef, "/") || strings.Contains(trimmedSubpath, "/")
}

// NormalizeExplicitSubpath normalizes and validates explicit --subpath values.
func NormalizeExplicitSubpath(sourceType SourceType, raw string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return "", fmt.Errorf("subpath cannot be empty")
	}

	if sourceType == GitHub {
		return normalizeGitHubSubpath(raw)
	}

	normalizedSeparators := strings.ReplaceAll(strings.TrimSpace(raw), "\\", "/")
	for _, segment := range strings.Split(normalizedSeparators, "/") {
		if segment == ".." {
			return "", fmt.Errorf("invalid subpath %q: parent traversal (..) is not supported", raw)
		}
	}

	normalized := normalizeParsedSubpath(raw)
	if normalized == "" {
		return "", fmt.Errorf("subpath cannot be empty")
	}

	return normalized, nil
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

	if normalized, rawErr := parseGitHubRawMarketplaceURL(parsedURL, input); rawErr != nil {
		return nil, rawErr
	} else if normalized != nil {
		return normalized, nil
	}

	// Check if it's a GitHub URL
	if githubURLRegex.MatchString(input) {
		parsed, parseErr := parseGitHubURL(input)
		if parseErr != nil {
			return nil, parseErr
		}

		// URLs that point to marketplace.json must be repo-backed file URLs that can
		// be normalized into clone URL + ref + subpath.
		if looksLikeMarketplaceManifestPath(parsedURL.Path) && parsed.Subpath == "" {
			return nil, fmt.Errorf("unsupported GitHub marketplace manifest URL %q: use a repo-backed /blob/<ref>/.../marketplace.json or raw.githubusercontent.com URL", input)
		}

		return parsed, nil
	}

	// Check if it's a GitLab URL
	if gitlabURLRegex.MatchString(input) {
		if looksLikeMarketplaceManifestPath(parsedURL.Path) {
			return nil, fmt.Errorf("unsupported remote marketplace manifest URL %q: only repo-backed URLs that normalize to clone URL + ref + manifest path are supported; standalone remote manifest fetching is not supported", input)
		}
		return parseGitLabURL(input)
	}

	if looksLikeMarketplaceManifestPath(parsedURL.Path) {
		return nil, fmt.Errorf("unsupported remote marketplace manifest URL %q: only repo-backed URLs that normalize to clone URL + ref + manifest path are supported; standalone remote manifest fetching is not supported", input)
	}

	// Generic git URL — extract subpath from .git/ delimiter if present
	// e.g., https://host/scm/PROJECT/repo.git/subpath → clone URL + subpath
	urlStr := parsedURL.String()
	var subpath string
	if idx := strings.Index(urlStr, ".git/"); idx != -1 {
		subpath = normalizeParsedSubpath(urlStr[idx+5:]) // everything after ".git/"
		urlStr = urlStr[:idx+4]                          // keep up to and including ".git"
	}

	return &ParsedSource{
		Type:    GitURL,
		URL:     urlStr,
		Subpath: subpath,
	}, nil
}

func parseGitHubRawMarketplaceURL(parsedURL *url.URL, input string) (*ParsedSource, error) {
	if parsedURL == nil {
		return nil, nil
	}

	if !strings.EqualFold(parsedURL.Host, "raw.githubusercontent.com") {
		return nil, nil
	}

	if !looksLikeMarketplaceManifestPath(parsedURL.Path) {
		return nil, fmt.Errorf("unsupported raw GitHub URL %q: only repo-backed marketplace.json file URLs are supported", input)
	}

	pathParts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(pathParts) < 4 {
		return nil, fmt.Errorf("unable to normalize raw GitHub marketplace URL %q: expected /<owner>/<repo>/<ref>/.../marketplace.json", input)
	}

	owner := pathParts[0]
	repo := strings.TrimSuffix(pathParts[1], ".git")
	ref := pathParts[2]
	subpath, err := normalizeGitHubSubpath(strings.Join(pathParts[3:], "/"))
	if err != nil {
		return nil, err
	}

	if owner == "" || repo == "" || ref == "" || subpath == "" {
		return nil, fmt.Errorf("unable to normalize raw GitHub marketplace URL %q: expected /<owner>/<repo>/<ref>/.../marketplace.json", input)
	}

	return &ParsedSource{
		Type:    GitHub,
		URL:     fmt.Sprintf("https://github.com/%s/%s", owner, repo),
		Ref:     ref,
		Subpath: subpath,
	}, nil
}

func looksLikeMarketplaceManifestPath(rawPath string) bool {
	if rawPath == "" {
		return false
	}
	return strings.EqualFold(path.Base(rawPath), marketplaceManifestFileName)
}

// parseGitHubURL parses a full GitHub URL
// Supports:
//   - https://github.com/owner/repo
//   - https://github.com/owner/repo.git
//   - https://github.com/owner/repo/tree/branch
//   - https://github.com/owner/repo/tree/branch/path/to/resource
func parseGitHubURL(input string) (*ParsedSource, error) {
	parsedURL, err := url.Parse(input)
	if err != nil {
		return nil, fmt.Errorf("invalid GitHub URL: %w", err)
	}
	if strings.HasSuffix(parsedURL.Path, ".git/") {
		return nil, fmt.Errorf("invalid GitHub URL: expected subpath after .git/")
	}

	// Extract owner/repo from path
	pathParts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(pathParts) < 2 {
		return nil, fmt.Errorf("invalid GitHub URL: must include owner and repo")
	}

	owner := pathParts[0]
	repoRaw := pathParts[1]
	repo := strings.TrimSuffix(repoRaw, ".git")

	if owner == "" || repo == "" {
		return nil, fmt.Errorf("GitHub owner and repo cannot be empty")
	}

	var ref string
	var subpath string

	// Check for /tree/branch or /blob/branch format
	if len(pathParts) >= 3 && (pathParts[2] == "tree" || pathParts[2] == "blob") {
		if len(pathParts) < 4 || strings.TrimSpace(pathParts[3]) == "" {
			return nil, fmt.Errorf("invalid GitHub URL: /%s requires a ref segment", pathParts[2])
		}
		ref = pathParts[3]
		// Subpath is everything after the ref
		if len(pathParts) > 4 {
			subpath, err = normalizeGitHubSubpath(strings.Join(pathParts[4:], "/"))
			if err != nil {
				return nil, err
			}
		}
	} else if strings.HasSuffix(repoRaw, ".git") && len(pathParts) > 2 {
		// Clone-style URL with explicit .git delimiter and repo-relative subpath.
		subpath, err = normalizeGitHubSubpath(strings.Join(pathParts[2:], "/"))
		if err != nil {
			return nil, err
		}
		if subpath == "" {
			return nil, fmt.Errorf("invalid GitHub URL: expected subpath after .git/")
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

func normalizeParsedSubpath(subpath string) string {
	trimmed := strings.TrimSpace(strings.ReplaceAll(subpath, "\\", "/"))
	if trimmed == "" {
		return ""
	}

	cleaned := path.Clean(strings.TrimPrefix(trimmed, "/"))
	if cleaned == "." || cleaned == "/" {
		return ""
	}

	return strings.TrimPrefix(cleaned, "/")
}

func normalizeGitHubSubpath(subpath string) (string, error) {
	normalizedSeparators := strings.ReplaceAll(strings.TrimSpace(subpath), "\\", "/")
	for _, segment := range strings.Split(normalizedSeparators, "/") {
		if segment == ".." {
			return "", fmt.Errorf("invalid GitHub subpath %q: parent traversal (..) is not supported", subpath)
		}
	}

	normalized := normalizeParsedSubpath(subpath)
	if normalized == ".." || strings.HasPrefix(normalized, "../") {
		return "", fmt.Errorf("invalid GitHub subpath %q: parent traversal (..) is not supported", subpath)
	}

	return normalized, nil
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
		return ps.URL, nil

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
