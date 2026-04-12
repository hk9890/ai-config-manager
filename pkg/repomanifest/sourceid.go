package repomanifest

import (
	"crypto/sha256"
	"encoding/hex"
	"path"
	"path/filepath"
	"strings"

	"github.com/dynatrace-oss/ai-config-manager/v3/pkg/giturl"
)

// GenerateSourceID produces a deterministic hash-based source ID from a source's
// canonical location (URL or absolute path). The returned ID has the format
// "src-" followed by 12 hex characters derived from a SHA-256 hash.
//
// For URL sources, canonical identity is normalized URL + normalized subpath.
// Ref is intentionally excluded so manifest-declared sources can update ref
// without changing canonical source identity.
//
// Subpath normalization rules:
//   - "", ".", "./", and "/" normalize to empty subpath
//   - leading slash is removed (subpath is repo-relative)
//   - trailing slash is removed
//   - repeated separators are collapsed
//   - nested paths are cleaned (e.g. "a/../b//" -> "b")
//
// For path sources, the path is resolved to an absolute path before hashing.
//
// Returns an empty string if the source has neither URL nor path.
func GenerateSourceID(source *Source) string {
	if source == nil {
		return ""
	}

	idSource := sourceIdentitySource(source)

	var canonical string

	if idSource.URL != "" {
		canonical = canonicalRemoteIdentity(idSource.URL, idSource.Subpath)
	} else if idSource.Path != "" {
		canonical = normalizePath(idSource.Path)
	} else {
		return ""
	}

	hash := sha256.Sum256([]byte(canonical))
	hexStr := hex.EncodeToString(hash[:])

	return "src-" + hexStr[:12]
}

// sourceIdentitySource returns the effective source transport used for canonical
// source identity. For overridden local sources, this maps identity back to the
// stored original remote URL/subpath to keep IDs stable across override/clear.
func sourceIdentitySource(source *Source) *Source {
	if source == nil {
		return nil
	}

	if source.OverrideOriginalURL != "" {
		return &Source{
			URL:     source.OverrideOriginalURL,
			Subpath: source.OverrideOriginalSubpath,
		}
	}

	return source
}

// normalizeURL normalizes a URL for consistent hashing by lowercasing,
// stripping trailing slashes, and removing .git suffixes.
func normalizeURL(url string) string {
	return giturl.NormalizeURL(url)
}

// normalizePath normalizes a filesystem path by resolving it to an absolute path.
func normalizePath(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		// Fall back to the original path if resolution fails
		return path
	}

	return abs
}

func canonicalRemoteIdentity(url, subpath string) string {
	canonicalURL := normalizeURL(url)
	canonicalSubpath := normalizeSubpath(subpath)
	if canonicalSubpath == "" {
		// Preserve legacy URL-only identity for empty subpath.
		return canonicalURL
	}
	return canonicalURL + "#" + canonicalSubpath
}

func normalizeSubpath(subpath string) string {
	cleaned := strings.TrimSpace(strings.ReplaceAll(subpath, "\\", "/"))
	if cleaned == "" {
		return ""
	}

	cleaned = strings.TrimPrefix(cleaned, "/")
	cleaned = path.Clean(cleaned)
	if cleaned == "." || cleaned == "/" {
		return ""
	}

	return strings.TrimPrefix(cleaned, "/")
}
