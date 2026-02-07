# Integration Plan: agent-skills-cli as Source Provider

**Goal**: Use agent-skills-cli's marketplace API as another source provider for aimgr, alongside local filesystem, Git repos, and GitHub.

---

## Current Source Providers

```
aimgr repo import <source>

Sources:
  1. local:/path/to/skill     → Local filesystem
  2. gh:owner/repo            → GitHub repository
  3. https://github.com/...   → Git URL
```

## Proposed: Add agentskills.in as Provider

```
aimgr repo import agentskills:@anthropic/xlsx
aimgr repo import agentskills:pdf-processing
aimgr repo import agentskills:react-*
```

---

## Architecture Design

### 1. New Source Type: `agentskills:`

**Source Format:**
```
agentskills:<skill-name>
agentskills:@<author>/<skill-name>
```

**Examples:**
```bash
# By name only
aimgr repo import agentskills:pdf-processing

# By scoped name (author/skill)
aimgr repo import agentskills:@anthropic/xlsx

# Search and import
aimgr search agentskills python
# Interactive selection, then imports chosen skills
```

### 2. API Client Implementation

**New Package**: `pkg/source/agentskills/`

```go
package agentskills

import (
    "encoding/json"
    "fmt"
    "net/http"
)

const (
    BaseURL = "https://www.agentskills.in/api"
)

// Skill represents a skill from the agentskills.in API
type Skill struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Author      string `json:"author"`
    ScopedName  string `json:"scopedName"`  // @author/name
    Description string `json:"description"`
    GitHubURL   string `json:"githubUrl"`
    RawURL      string `json:"rawUrl"`
    Stars       int    `json:"stars"`
    Forks       int    `json:"forks"`
    Category    string `json:"category"`
    Content     string `json:"content"`     // Full SKILL.md content
}

// SearchOptions for filtering skills
type SearchOptions struct {
    Query    string
    Author   string
    Category string
    Limit    int
    Offset   int
    SortBy   string // "stars", "recent", "name"
}

// Client for agentskills.in API
type Client struct {
    BaseURL    string
    HTTPClient *http.Client
}

// NewClient creates a new agentskills API client
func NewClient() *Client {
    return &Client{
        BaseURL:    BaseURL,
        HTTPClient: &http.Client{},
    }
}

// Search searches for skills
func (c *Client) Search(opts SearchOptions) ([]Skill, error) {
    params := url.Values{}
    if opts.Query != "" {
        params.Set("search", opts.Query)
    }
    if opts.Author != "" {
        params.Set("author", opts.Author)
    }
    if opts.Category != "" {
        params.Set("category", opts.Category)
    }
    if opts.Limit > 0 {
        params.Set("limit", fmt.Sprintf("%d", opts.Limit))
    }
    if opts.Offset > 0 {
        params.Set("offset", fmt.Sprintf("%d", opts.Offset))
    }
    if opts.SortBy != "" {
        params.Set("sortBy", opts.SortBy)
    }

    url := fmt.Sprintf("%s/skills?%s", c.BaseURL, params.Encode())
    
    resp, err := c.HTTPClient.Get(url)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch skills: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
    }

    var result struct {
        Skills []Skill `json:"skills"`
        Total  int     `json:"total"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }

    return result.Skills, nil
}

// GetByName fetches a skill by name (with optional author)
func (c *Client) GetByName(name string, author string) (*Skill, error) {
    opts := SearchOptions{
        Query:  name,
        Author: author,
        Limit:  1,
    }

    skills, err := c.Search(opts)
    if err != nil {
        return nil, err
    }

    if len(skills) == 0 {
        return nil, fmt.Errorf("skill not found: %s", name)
    }

    return &skills[0], nil
}

// GetByScopedName fetches a skill by scoped name (@author/name)
func (c *Client) GetByScopedName(scopedName string) (*Skill, error) {
    // Parse @author/name
    scopedName = strings.TrimPrefix(scopedName, "@")
    parts := strings.SplitN(scopedName, "/", 2)
    
    if len(parts) != 2 {
        return nil, fmt.Errorf("invalid scoped name: %s", scopedName)
    }

    author := parts[0]
    name := parts[1]

    return c.GetByName(name, author)
}

// DownloadSkill downloads skill content from GitHub via the raw URL
func (c *Client) DownloadSkill(skill *Skill, destDir string) error {
    // Skill content is already in the API response
    if skill.Content != "" {
        skillPath := filepath.Join(destDir, "SKILL.md")
        if err := os.MkdirAll(destDir, 0755); err != nil {
            return fmt.Errorf("failed to create directory: %w", err)
        }
        if err := os.WriteFile(skillPath, []byte(skill.Content), 0644); err != nil {
            return fmt.Errorf("failed to write skill content: %w", err)
        }
        return nil
    }

    // Fallback: download from GitHub raw URL
    if skill.RawURL != "" {
        return c.downloadFromURL(skill.RawURL, destDir)
    }

    // Last resort: clone from GitHub URL
    if skill.GitHubURL != "" {
        return cloneFromGitHub(skill.GitHubURL, destDir)
    }

    return fmt.Errorf("no download source available for skill: %s", skill.Name)
}

func (c *Client) downloadFromURL(url string, destDir string) error {
    resp, err := c.HTTPClient.Get(url)
    if err != nil {
        return fmt.Errorf("failed to download skill: %w", err)
    }
    defer resp.Body.Close()

    content, err := io.ReadAll(resp.Body)
    if err != nil {
        return fmt.Errorf("failed to read response: %w", err)
    }

    skillPath := filepath.Join(destDir, "SKILL.md")
    if err := os.MkdirAll(destDir, 0755); err != nil {
        return fmt.Errorf("failed to create directory: %w", err)
    }
    if err := os.WriteFile(skillPath, content, 0644); err != nil {
        return fmt.Errorf("failed to write skill: %w", err)
    }

    return nil
}
```

### 3. Source Parser Integration

**Update `pkg/source/parser.go`:**

```go
package source

import (
    "fmt"
    "strings"
    
    "github.com/hk9890/ai-config-manager/pkg/source/agentskills"
)

// Source types
const (
    SourceTypeLocal       = "local"
    SourceTypeGitHub      = "github"
    SourceTypeGit         = "git"
    SourceTypeAgentSkills = "agentskills"  // NEW
)

// ParseSource parses a source string into structured information
func ParseSource(source string) (*SourceInfo, error) {
    // Check for agentskills: prefix
    if strings.HasPrefix(source, "agentskills:") {
        skillRef := strings.TrimPrefix(source, "agentskills:")
        return &SourceInfo{
            Type:      SourceTypeAgentSkills,
            Reference: skillRef,
            IsRemote:  true,
        }, nil
    }

    // ... existing local, gh:, git URL parsing ...
}

// ResolveSource resolves a source to a local path
func ResolveSource(source string, workspace *workspace.Manager) (string, error) {
    info, err := ParseSource(source)
    if err != nil {
        return "", err
    }

    switch info.Type {
    case SourceTypeAgentSkills:
        return resolveAgentSkillsSource(info, workspace)
    case SourceTypeGitHub:
        return resolveGitHubSource(info, workspace)
    // ... existing cases ...
    }
}

func resolveAgentSkillsSource(info *SourceInfo, workspace *workspace.Manager) (string, error) {
    client := agentskills.NewClient()
    
    // Parse skill reference (@author/name or just name)
    var skill *agentskills.Skill
    var err error
    
    if strings.HasPrefix(info.Reference, "@") || strings.Contains(info.Reference, "/") {
        // Scoped name: @author/name or author/name
        skill, err = client.GetByScopedName(info.Reference)
    } else {
        // Just name: fetch by name
        skill, err = client.GetByName(info.Reference, "")
    }
    
    if err != nil {
        return "", fmt.Errorf("failed to fetch skill from agentskills.in: %w", err)
    }

    // Download to temporary directory
    tempDir, err := os.MkdirTemp("", "agentskills-*")
    if err != nil {
        return "", fmt.Errorf("failed to create temp directory: %w", err)
    }

    if err := client.DownloadSkill(skill, tempDir); err != nil {
        os.RemoveAll(tempDir)
        return "", fmt.Errorf("failed to download skill: %w", err)
    }

    return tempDir, nil
}
```

### 4. Config File Integration

**`~/.config/aimgr/aimgr.yaml`:**

```yaml
sync:
  sources:
    # Local filesystem
    - url: ~/dev/my-skills
      filter: "skill/*"

    # GitHub
    - url: gh:anthropics/skills
      filter: "skill/*"

    # Agent Skills marketplace (NEW)
    - url: agentskills:@anthropic/xlsx
    - url: agentskills:@anthropic/pdf-processing
    - url: agentskills:react-*
```

### 5. CLI Commands

**Search command:**

```bash
# Search agentskills.in marketplace
aimgr search agentskills python

# Output:
# Found 45 skills matching "python":
# 
# 1. @anthropic/python-helper (⭐ 450)
#    Python development assistant
#    Source: agentskills.in
# 
# 2. @vercel/python-testing (⭐ 320)
#    Python testing utilities
#    Source: agentskills.in
# 
# Select skills to import (space to select, enter to confirm):
# [x] 1. python-helper
# [ ] 2. python-testing
```

**Import command:**

```bash
# Import single skill
aimgr repo import agentskills:@anthropic/xlsx

# Import multiple skills (pattern matching in future)
aimgr repo import agentskills:@anthropic/xlsx agentskills:@anthropic/pdf

# Add to sync config
aimgr config add-sync-source agentskills:@anthropic/xlsx
```

---

## Implementation Plan

### Phase 1: Basic Integration (Week 1-2)

**Tasks:**
1. Create `pkg/source/agentskills/` package
   - API client
   - Skill struct
   - Search, GetByName, GetByScopedName
   - DownloadSkill (from content or raw URL)

2. Update `pkg/source/parser.go`
   - Add `SourceTypeAgentSkills`
   - Parse `agentskills:` prefix
   - Resolve to temp directory with downloaded skill

3. Test with `repo import`
   ```bash
   aimgr repo import agentskills:@anthropic/xlsx
   ```

**Deliverable**: Can import skills from agentskills.in via command line

### Phase 2: Config Integration (Week 3)

**Tasks:**
1. Update config schema to support `agentskills:` URLs
2. Test `repo sync` with agentskills sources
3. Add validation for agentskills source format

**Deliverable**: Can sync skills from agentskills.in via config file

### Phase 3: Search Command (Week 4)

**Tasks:**
1. Create `aimgr search` command
2. Interactive selection (bubbles/tea or simple prompts)
3. Add to sync config after selection

**Deliverable**: Can search and interactively select skills

### Phase 4: Caching & Metadata (Week 5-6)

**Tasks:**
1. Add metadata tracking for agentskills sources
   - Store skill ID, author, GitHub URL
   - Track stars, forks for update checking
2. Optimize: cache API responses (5-minute TTL)
3. Add `repo update` to check for skill updates from marketplace

**Deliverable**: Efficient caching, update checking

---

## Benefits of This Approach

1. **Discoverability** ✅
   - Search 100k+ skills via agentskills.in
   - Browse by category, author, stars

2. **Local Control** ✅
   - Skills stored in your local repository
   - No external dependency for project installs
   - Works offline after initial import

3. **Multiple Sources** ✅
   - agentskills.in (discovery)
   - GitHub (team repos)
   - Local filesystem (custom skills)
   - All managed uniformly via `repo sync`

4. **No Lock-in** ✅
   - Can still use all existing sources
   - agentskills.in is just another provider
   - If marketplace goes down, you still have local copies

5. **Complementary, Not Competing** ✅
   - They provide discovery & marketplace
   - We provide local management & project integration
   - Clear separation of concerns

---

## Example Workflow

```bash
# 1. Search for skills on agentskills.in
aimgr search agentskills pdf

# Output:
# Found 12 skills:
# 1. @anthropic/pdf-processing (⭐ 450)
# 2. @vercel/pdf-reader (⭐ 320)
# ...

# 2. Import selected skill to your repository
aimgr repo import agentskills:@anthropic/pdf-processing

# 3. Add to sync config for automatic updates
vim ~/.config/aimgr/aimgr.yaml
# sync:
#   sources:
#     - url: agentskills:@anthropic/pdf-processing

# 4. Install to your project
cd my-project
aimgr install skill/pdf-processing

# 5. Update all sources (including agentskills)
aimgr repo sync
```

---

## Open Questions

1. **Scoped Names**: Do we preserve @author/name format in our repo?
   - Option A: Store as `skills/pdf-processing/` (current format)
   - Option B: Store as `skills/@anthropic/pdf-processing/` (preserve namespace)
   - **Recommendation**: Option A (simpler, consistent with existing format)

2. **Metadata Tracking**: How to track agentskills source?
   ```yaml
   # .metadata.yaml
   name: pdf-processing
   type: skill
   source:
     type: agentskills
     id: skill-123
     author: anthropic
     scopedName: "@anthropic/pdf-processing"
     githubUrl: "https://github.com/anthropics/skills/tree/main/skills/pdf"
     stars: 450
     lastUpdated: "2026-02-07T10:30:00Z"
   ```

3. **Updates**: How to check for updates?
   - Option A: Query API for skill by ID, compare `updated_at`
   - Option B: Query API by scoped name, compare content hash
   - **Recommendation**: Option A (more efficient)

4. **Conflicts**: What if GitHub and agentskills have same skill?
   - Allow multiple sources
   - Track source in metadata
   - User chooses which source to use via config

---

## Next Steps

1. **Prototype API client** (`pkg/source/agentskills/client.go`)
2. **Test API endpoints** (search, get by name)
3. **Integrate with source parser** (`pkg/source/parser.go`)
4. **Test end-to-end** (`aimgr repo import agentskills:...`)
5. **Add search command** (`aimgr search agentskills ...`)

Would you like me to start implementing Phase 1?
