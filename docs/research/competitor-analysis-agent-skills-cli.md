# Competitor Analysis: agent-skills-cli vs aimgr

**Date**: 2026-02-07  
**Repository**: https://github.com/Karanjot786/agent-skills-cli  
**Website**: https://www.agentskills.in  
**npm Package**: https://www.npmjs.com/package/agent-skills-cli

---

## Executive Summary

**agent-skills-cli** (by Karanjot Singh) is a direct competitor to our `aimgr` project. Both tools manage AI agent skills (commands, skills, agents) across multiple AI coding platforms.

### Key Similarities
- Both manage AI resources (commands, skills, agents)
- Both support multiple AI tools (Claude, Copilot, Cursor, OpenCode, etc.)
- Both use symlink-based installation
- Both have centralized repository storage
- Both support GitHub imports

### Key Differences

| Feature | agent-skills-cli | aimgr |
|---------|-----------------|-------|
| **Language** | TypeScript/Node.js | Go |
| **Central Marketplace** | ✅ 100,000+ skills via Supabase API | ❌ None (local only) |
| **Online Marketplace** | ✅ https://www.agentskills.in | ❌ None |
| **Database Backend** | ✅ PostgreSQL (Supabase) | ❌ Filesystem only |
| **Installation** | npm global install | Go install / binary |
| **Supported Platforms** | 42 AI agents | 3 AI tools (Claude, OpenCode, Copilot) |
| **Search Feature** | ✅ FZF interactive search | ❌ None |
| **Lock File** | ✅ `~/.skills/skills.lock` | ❌ None |
| **Telemetry** | ✅ Anonymous usage tracking | ❌ None |
| **Workspace Caching** | ❌ None (temp dirs) | ✅ 10-50x faster Git caching |
| **Packages** | ❌ None | ✅ Package system |
| **XDG Compliance** | ❌ Hardcoded paths | ✅ Full XDG support |

---

## Architecture Comparison

### agent-skills-cli Architecture

**Technology Stack:**
- **Language**: TypeScript (Node.js)
- **Database**: Supabase (PostgreSQL)
- **Storage**: `~/.skills/` (canonical), `~/.skills/skills.lock` (tracking)
- **Distribution**: npm package

**Key Components:**

1. **Marketplace API** (`src/core/skillsdb.ts`)
   - Fetches skills from `https://www.agentskills.in/api/skills`
   - Supports search, filtering, sorting
   - 100,000+ skills indexed in Supabase database

2. **Installer** (`src/core/installer.ts`)
   - Symlink-based installation (falls back to copy on Windows)
   - Canonical storage: `~/.skills/<skill-name>`
   - Agent directories: `.claude/skills/`, `.cursor/skills/`, etc.

3. **Skills Database** (Supabase PostgreSQL)
   ```sql
   model skills {
     id               String    @id
     name             String
     author           String
     scoped_name      String?   -- @author/skill format
     description      String?
     github_url       String    @unique
     stars            Int
     forks            Int
     category         String?
     content          String?   -- Full SKILL.md content
     indexed_at       DateTime
     updated_at       DateTime?
   }
   ```

4. **42 Supported Agents**
   - Claude Code, Cursor, Copilot, Windsurf, Cline, Zed, OpenCode
   - +35 more: Amp, Kilo, Roo, Goose, CodeBuddy, Continue, etc.

**Installation Flow:**

```
User runs: skills install @anthropic/xlsx
    ↓
1. Fetch from API: https://www.agentskills.in/api/skills?search=xlsx
    ↓
2. Download skill from GitHub raw URL
    ↓
3. Copy to canonical: ~/.skills/xlsx/
    ↓
4. Create symlinks:
   - .claude/skills/xlsx/ → ~/.skills/xlsx/
   - .cursor/skills/xlsx/ → ~/.skills/xlsx/
    ↓
5. Update lock file: ~/.skills/skills.lock
```

**Lock File Format:**
```json
{
  "skills": [
    {
      "name": "xlsx",
      "canonicalPath": "~/.skills/xlsx",
      "agents": ["claude", "cursor"],
      "linkedPaths": [
        ".claude/skills/xlsx",
        ".cursor/skills/xlsx"
      ],
      "method": "symlink",
      "source": "https://github.com/anthropics/skills/tree/main/skills/xlsx",
      "installedAt": "2026-02-07T10:30:00Z"
    }
  ]
}
```

### aimgr Architecture

**Technology Stack:**
- **Language**: Go 1.25.6
- **Database**: None (filesystem-based)
- **Storage**: `~/.local/share/ai-config/repo/` (XDG-compliant)
- **Distribution**: Go binary

**Key Components:**

1. **Repository Manager** (`pkg/repo/`)
   - Filesystem-based storage
   - No central marketplace
   - Import from GitHub or local paths

2. **Workspace Cache** (`pkg/workspace/`)
   - Git repository caching (10-50x faster)
   - Persistent clones in `~/.local/share/ai-config/repo/.workspace/`
   - Automatic pruning of unused caches

3. **Discovery** (`pkg/discovery/`)
   - Auto-discovers resources in directories
   - Supports Claude (`.claude/`), OpenCode (`.opencode/`), Copilot (`.github/`) structures

4. **Package System** (`pkg/resource/package.go`)
   - Group resources into packages
   - `web-dev-tools.package.json` format

**Installation Flow:**

```
User runs: aimgr install skill/pdf-processing
    ↓
1. Read from local repo: ~/.local/share/ai-config/repo/skills/pdf-processing/
    ↓
2. Create symlinks:
   - .claude/skills/pdf-processing/ → <repo>/skills/pdf-processing/
   - .opencode/skills/pdf-processing/ → <repo>/skills/pdf-processing/
    ↓
(No lock file or tracking)
```

---

## Central Marketplace: The Key Differentiator

### agent-skills-cli Marketplace

**Location**: https://www.agentskills.in

**Database Schema:**
```sql
-- skills table (Supabase PostgreSQL)
CREATE TABLE skills (
  id               TEXT PRIMARY KEY,
  name             TEXT NOT NULL,
  author           TEXT NOT NULL,
  scoped_name      TEXT,              -- @author/skill
  description      TEXT,
  github_url       TEXT UNIQUE NOT NULL,
  raw_url          TEXT,
  stars            INT DEFAULT 0,
  forks            INT DEFAULT 0,
  category         TEXT,
  content          TEXT,              -- Full SKILL.md content
  indexed_at       TIMESTAMPTZ DEFAULT NOW(),
  updated_at       TIMESTAMPTZ
);

CREATE INDEX idx_skills_author ON skills(author);
CREATE INDEX idx_skills_name ON skills(name);
CREATE INDEX idx_skills_stars ON skills(stars DESC);
CREATE INDEX idx_skills_category ON skills(category);
```

**API Endpoints:**

1. **Search Skills**: `GET /api/skills`
   ```bash
   curl "https://www.agentskills.in/api/skills?search=pdf&limit=10&sortBy=stars"
   ```
   
   **Response:**
   ```json
   {
     "skills": [
       {
         "id": "skill-123",
         "name": "pdf-processing",
         "author": "anthropic",
         "scopedName": "@anthropic/pdf-processing",
         "description": "Process PDF documents",
         "stars": 250,
         "forks": 45,
         "githubUrl": "https://github.com/anthropics/skills/tree/main/skills/pdf",
         "rawUrl": "https://raw.githubusercontent.com/...",
         "content": "# PDF Processing Skill\n...",
         "category": "documents"
       }
     ],
     "total": 1
   }
   ```

2. **Get Skill Details**: `GET /api/skills?author=anthropic&name=xlsx`

3. **Categories**: `GET /api/categories`
   ```json
   {
     "categories": [
       {"id": "documents", "name": "Documents", "icon": "📄", "skillCount": 120},
       {"id": "web", "name": "Web Development", "icon": "🌐", "skillCount": 350}
     ]
   }
   ```

4. **Stats**: `GET /api/stats`
   ```json
   {
     "totalSkills": 100000,
     "uniqueAuthors": 5000,
     "topAuthors": [...]
   }
   ```

**CLI Integration:**

```typescript
// src/core/skillsdb.ts
const SKILLS_API = 'https://www.agentskills.in/api/skills';

export async function fetchFromDB(options: FetchOptions): Promise<SkillsDBResult> {
    const params = new URLSearchParams();
    if (options.search) params.set('search', options.search);
    if (options.author) params.set('author', options.author);
    if (options.category) params.set('category', options.category);
    
    const res = await fetch(`${SKILLS_API}?${params}`);
    return await res.json();
}
```

**Skill Sources:**
- GitHub repositories (indexed by crawlers)
- agentskills.io submissions
- Anthropic skills repository
- Community contributions

### aimgr: No Central Marketplace

**Current State**: aimgr has NO central marketplace or database.

**Resources must be added manually:**
```bash
# Import from GitHub
aimgr repo import gh:vercel-labs/agent-skills

# Import from local directory
aimgr repo import ~/.claude/skills/
```

**Why this is a limitation:**
- ❌ No discoverability - users must know exact GitHub URLs
- ❌ No search - can't find "pdf processing skills"
- ❌ No ratings/stars - can't see popular skills
- ❌ No categories - can't browse by topic
- ❌ No central repository - fragmented ecosystem

---

## Feature-by-Feature Comparison

### 1. Resource Discovery

**agent-skills-cli:**
```bash
# Interactive search with FZF
$ skills search python
# Shows live-updating list of skills with descriptions

# Search with JSON output
$ skills search react --json
{
  "skills": [
    {"name": "react-helper", "author": "vercel", "stars": 450},
    {"name": "react-testing", "author": "facebook", "stars": 320}
  ]
}

# Browse by category
$ skills search --category web

# Browse by author
$ skills search --author anthropic
```

**aimgr:**
```bash
# No search feature
# Must list all resources and manually filter
$ aimgr repo list skill
# Shows only locally imported skills

# No way to discover new skills without knowing GitHub URLs
```

**Winner**: **agent-skills-cli** (has real search, central database)

---

### 2. Installation

**agent-skills-cli:**
```bash
# Install from marketplace
$ skills install @anthropic/xlsx

# Install to specific platforms
$ skills install pdf -t claude,cursor

# Install globally (home directory)
$ skills install docx -g

# Install to all 42 platforms
$ skills install xlsx --all
```

**aimgr:**
```bash
# Install from local repository
$ aimgr install skill/pdf-processing

# Install to specific tools
$ aimgr install skill/pdf --tool=claude,opencode

# No global install (always local to repository)
```

**Winner**: **agent-skills-cli** (more flexible installation targets)

---

### 3. Storage & Caching

**agent-skills-cli:**
```
~/.skills/                    # Canonical storage
  ├── xlsx/
  ├── pdf-processing/
  └── skills.lock             # Tracking file

# No Git caching (clones to temp, copies, deletes)
```

**aimgr:**
```
~/.local/share/ai-config/repo/   # XDG-compliant storage
  ├── skills/
  ├── commands/
  ├── agents/
  └── .workspace/                # Git cache (10-50x faster)
      ├── github.com-anthropics-skills-main/
      └── github.com-vercel-agent-skills-main/
```

**Winner**: **aimgr** (XDG-compliant, Git caching for performance)

---

### 4. Update & Sync

**agent-skills-cli:**
```bash
# Check for updates
$ skills check

# Update skills from source
$ skills update --all

# Lock file tracks installed versions
$ cat ~/.skills/skills.lock
```

**aimgr:**
```bash
# Sync from configured sources
$ aimgr repo sync

# Config file:
# ~/.config/aimgr/aimgr.yaml
sync:
  sources:
    - url: gh:anthropics/skills
      filter: "skill/*"
```

**Winner**: **Tie** (different approaches, both functional)

---

### 5. Supported Platforms

**agent-skills-cli**: **42 AI agents**
- Claude Code, Cursor, GitHub Copilot, OpenCode
- Windsurf, Cline, Gemini CLI, Zed, Antigravity
- +33 more: Amp, Kilo, Roo, Goose, CodeBuddy, Continue, Crush, Clawdbot, Droid, Kiro, MCPJam, Mux, OpenHands, Pi, Qoder, Qwen Code, Trae, Zencoder, Neovate, Command Code, Ara, Aide, Alex, BB, CodeStory, Helix AI, Meekia, Pear AI, Adal, Pochi, Sourcegraph Cody, Void AI

**aimgr**: **3 AI tools**
- Claude Code
- OpenCode
- GitHub Copilot / VSCode

**Winner**: **agent-skills-cli** (14x more platforms)

---

### 6. Package System

**agent-skills-cli:**
```bash
# No package system
# Must install skills individually
```

**aimgr:**
```bash
# Package system exists
$ aimgr install package/web-dev-tools

# Packages group resources
# web-dev-tools.package.json:
{
  "name": "web-dev-tools",
  "resources": ["command/build", "skill/typescript"]
}
```

**Winner**: **aimgr** (has packages, agent-skills-cli doesn't)

---

### 7. Telemetry & Analytics

**agent-skills-cli:**
```typescript
// src/core/telemetry.ts
export async function trackInstall(skillName: string) {
    await fetch('https://www.agentskills.in/api/telemetry', {
        method: 'POST',
        body: JSON.stringify({
            event: 'install',
            skill: skillName,
            version: CLI_VERSION,
            isCI: isCI()
        })
    });
}

// Opt-out:
export DISABLE_TELEMETRY=1
export DO_NOT_TRACK=1
```

**aimgr:**
```bash
# No telemetry
# No usage tracking
# No analytics
```

**Winner**: **agent-skills-cli** (has telemetry for product insights)

---

## Strategic Implications

### agent-skills-cli's Advantages

1. **Central Marketplace** (100,000+ skills)
   - **Impact**: Users can discover skills without knowing GitHub URLs
   - **Network Effect**: More skills → more users → more skills
   - **Business Model**: Potential for monetization (premium skills, verified authors)

2. **Search & Discovery**
   - FZF interactive search
   - Category browsing
   - Star ratings
   - Author profiles

3. **Platform Coverage** (42 agents)
   - Addresses the entire AI coding tool market
   - Future-proofs against tool churn

4. **Telemetry**
   - Data-driven product decisions
   - Track popular skills
   - Understand user behavior

### aimgr's Advantages

1. **Git Workspace Caching** (10-50x faster)
   - Persistent Git clones
   - Automatic pruning
   - Architecture Rule 1 compliance

2. **XDG Compliance**
   - Respects user environment
   - Cross-platform directory standards
   - User control via environment variables

3. **Package System**
   - Group resources into collections
   - Distribute themed resource sets

4. **Go Performance**
   - Single binary
   - No Node.js dependency
   - Faster startup

---

## Recommendations

### Short-term (1-2 months)

1. **Add Search Feature**
   - Not reliant on central database initially
   - Search local repository + configured GitHub sources
   - `aimgr search python` → searches locally + gh:anthropics/skills

2. **Expand Platform Support**
   - Add Cursor, Windsurf, Cline
   - Research top 10 AI coding tools
   - Implement directory detection

3. **Add Lock File Tracking**
   - `~/.local/share/ai-config/repo/.installed.json`
   - Track installation timestamps
   - Enable `aimgr check` for updates

### Mid-term (3-6 months)

4. **Build Central Marketplace** (Optional)
   - Decision point: compete directly or integrate?
   - Option A: Build competing marketplace
   - Option B: Federate with agentskills.in
   - Option C: Focus on local-first, avoid centralization

5. **Add Categories & Metadata**
   - Category taxonomy
   - Star ratings (from GitHub)
   - Author profiles

6. **Telemetry** (Optional, with consent)
   - Anonymous usage tracking
   - Opt-in by default
   - Respect DO_NOT_TRACK

### Long-term (6-12 months)

7. **Marketplace Strategy**
   - If building marketplace: index GitHub, agentskills.io
   - If federating: API client for agentskills.in
   - If local-first: enhanced discovery, smart recommendations

8. **Web UI**
   - Browse skills visually
   - Install via web interface
   - Share collections

---

## Competitive Positioning

### If We Build a Marketplace

**Differentiators:**
- Go performance vs Node.js
- XDG compliance
- Git workspace caching (10-50x faster updates)
- Package system
- Open-source, self-hostable

**Positioning**: "Fast, local-first alternative to agent-skills-cli"

### If We Stay Local-First

**Differentiators:**
- No central authority
- Privacy-focused (no telemetry by default)
- Git-native (direct GitHub integration)
- Enterprise-friendly (self-hosted, no external dependencies)

**Positioning**: "Enterprise-grade, local-first AI resource manager"

---

## Conclusion

**agent-skills-cli** has a significant advantage with its central marketplace (100,000+ skills) and broad platform support (42 agents). However, **aimgr** has superior architecture (Git caching, XDG compliance, package system) and can compete by:

1. **Adding search/discovery** without requiring a central database
2. **Expanding platform support** to match top AI tools
3. **Deciding on marketplace strategy**: build, federate, or stay local-first

The choice depends on our strategic goals:
- **Compete directly**: Build marketplace, add telemetry, match features
- **Stay local-first**: Focus on privacy, performance, enterprise use cases
- **Federate**: Integrate with agentskills.in API, add unique value on top

**Next Steps:**
1. Team discussion on marketplace strategy
2. Prototype search feature (local + GitHub)
3. Research top 10 AI coding tools for platform expansion
