# Troubleshooting Guide

Common issues with `aimgr` (AI Resource Manager) and how to resolve them.

## Table of Contents

- [Installation Issues](#installation-issues)
  - [aimgr Not Found](#aimgr-not-found)
  - [Permission Denied](#permission-denied)
  - [Binary Not Executable](#binary-not-executable)
  - [Installation Script Fails](#installation-script-fails)
- [Symlink Issues](#symlink-issues)
  - [Broken Symlinks](#broken-symlinks)
  - [Skills Not Detected After Installation](#skills-not-detected-after-installation)
  - [Symlink Points to Wrong Location](#symlink-points-to-wrong-location)
- [Configuration Issues](#configuration-issues)
  - [Config File Missing](#config-file-missing)
  - [Invalid Config YAML](#invalid-config-yaml)
  - [Wrong Default Target](#wrong-default-target)
- [Repository Issues](#repository-issues)
  - [Resource Not Found](#resource-not-found)
  - [Repository Update Fails](#repository-update-fails)
  - [Duplicate Resources](#duplicate-resources)
  - [Repository Corruption](#repository-corruption)
- [Tool Detection Issues](#tool-detection-issues)
  - [Wrong Tool Directory Selected](#wrong-tool-directory-selected)
  - [Multiple Tool Directories Exist](#multiple-tool-directories-exist)
  - [AI Tool Doesn't Load Resources](#ai-tool-doesnt-load-resources)

---

## Installation Issues

### aimgr Not Found

**Symptoms:** 
- `command not found: aimgr`
- `bash: aimgr: command not found`
- Shell can't find the `aimgr` executable

**Diagnosis:**
```bash
# Check if aimgr is in PATH
which aimgr

# Check current PATH
echo $PATH

# Check common installation directories
ls -la ~/.local/bin/aimgr
ls -la /usr/local/bin/aimgr
```

**Solutions:**

1. **Install aimgr** (if not installed):
   ```bash
   # Linux/macOS
   ```bash
   # Using Go (recommended)
   go install github.com/hk9890/ai-config-manager@latest
   
   # Or from source
   git clone https://github.com/hk9890/ai-config-manager.git
   cd ai-config-manager
   make install
   ```

2. **Add to PATH** (if installed but not in PATH):
   ```bash
   # For bash
   echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
   source ~/.bashrc
   
   # For zsh
   echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
   source ~/.zshrc
   
   # For fish
   fish_add_path ~/.local/bin
   ```

3. **Verify installation location**:
   ```bash
   # Find where aimgr is installed
   find ~ -name aimgr -type f 2>/dev/null
   
   # If found in unexpected location, create symlink
   ln -s /path/to/aimgr ~/.local/bin/aimgr
   ```

---

### Permission Denied

**Symptoms:**
- `Permission denied` when running `aimgr`
- `zsh: permission denied: aimgr`
- File exists but can't be executed

**Diagnosis:**
```bash
# Check file permissions
ls -la $(which aimgr)

# Should show: -rwxr-xr-x (executable bits set)
# If shows: -rw-r--r-- (no executable bits)
```

**Solutions:**

1. **Fix binary permissions**:
   ```bash
   chmod +x $(which aimgr)
   
   # Or if you know the path
   chmod +x ~/.local/bin/aimgr
   ```

2. **Check ownership**:
   ```bash
   # If owned by root, may need sudo or change ownership
   ls -la $(which aimgr)
   
   # Fix ownership if needed
   sudo chown $USER:$USER $(which aimgr)
   ```

3. **SELinux/AppArmor issues** (Linux):
   ```bash
   # Check SELinux context
   ls -Z $(which aimgr)
   
   # If needed, restore default context
   restorecon -v $(which aimgr)
   ```

---

### Binary Not Executable

**Symptoms:**
- `cannot execute binary file: Exec format error`
- Wrong architecture error

**Diagnosis:**
```bash
# Check binary architecture
file $(which aimgr)

# Check system architecture
uname -m

# Should match: x86_64, arm64, etc.
```

**Solutions:**

1. **Download correct architecture**:
   ```bash
   # Determine your architecture
   uname -m
   
   # Download matching binary:
   # - x86_64 (Intel/AMD)
   # - aarch64/arm64 (ARM)
   # - Check releases for your platform
   ```

2. **Verify download integrity**:
   ```bash
   # Check if download completed
   ls -lh $(which aimgr)
   
   # Re-download if file seems corrupted
   ```

---

### Build from Source Fails

**Symptoms:**
- `make install` or `go install` throws errors
- Build fails with compilation errors
- Missing dependencies

**Diagnosis:**
```bash
# Check Go version (needs 1.21+)
go version

# Check network connectivity
curl -I https://github.com

# Check disk space
df -h

# Check write permissions
ls -la ~/bin/ ~/.local/bin/
```

**Solutions:**

1. **Update Go version**:
   ```bash
   # Check minimum required version
   go version  # Should be 1.21 or higher
   
   # Update if needed (see https://go.dev/doc/install)
   ```

2. **Manual build and install**:
   ```bash
   # Clone and build
   git clone https://github.com/hk9890/ai-config-manager.git
   cd ai-config-manager
   go build -o aimgr
   
   # Install manually
   mkdir -p ~/bin
   cp aimgr ~/bin/
   chmod +x ~/bin/aimgr
   
   # Verify
   ~/bin/aimgr --version
   ```

3. **Check prerequisites**:
   ```bash
   # Ensure required tools exist
   which go
   which git
   which make
   ```

---

## Symlink Issues

### Broken Symlinks

**Symptoms:**
- Skills show as installed but don't work
- AI tool reports "skill not found"
- `ls` shows red/broken symlinks

**Diagnosis:**
```bash
# Check symlinks in Claude Code
ls -la .claude/skills/

# Check symlinks in OpenCode
ls -la .opencode/skills/

# Find broken symlinks
find .claude .opencode -xtype l 2>/dev/null

# Check what symlink points to
readlink .claude/skills/some-skill

# Verify target exists
ls -la $(readlink .claude/skills/some-skill)
```

**Solutions:**

1. **Reinstall the resource**:
   ```bash
   # Uninstall and reinstall
   aimgr uninstall skill/broken-skill
   aimgr install skill/broken-skill
   ```

2. **Manual symlink fix**:
   ```bash
   # Remove broken symlink
   rm .claude/skills/broken-skill
   
   # Find resource in repository
   aimgr repo describe skill broken-skill
   
   # Recreate symlink manually
   ln -s ~/.local/share/ai-config/repo/skills/broken-skill .claude/skills/
   ```

3. **Repository verification**:
   ```bash
   # Verify resource exists in repository
   ls -la ~/.local/share/ai-config/repo/skills/
   
   # If missing, update repository
   aimgr repo sync --force
   ```

---

### Skills Not Detected After Installation

**Symptoms:**
- `aimgr list` shows skill installed
- AI tool doesn't detect/load the skill
- Skill doesn't appear in tool's skill list

**Diagnosis:**
```bash
# Verify installation
aimgr list --format=json

# Check symlink exists
ls -la .claude/skills/ .opencode/skills/

# Verify symlink target is valid
file .claude/skills/some-skill

# Check SKILL.md exists
ls -la .claude/skills/some-skill/SKILL.md
```

**Solutions:**

1. **Restart the AI tool** (MOST COMMON FIX):
   - Close Claude Code completely
   - Close OpenCode/VS Code completely
   - Reopen the tool
   - Skills are loaded at startup

2. **Verify correct directory structure**:
   ```bash
   # Skill should have SKILL.md in root
   ls -la .claude/skills/some-skill/SKILL.md
   
   # If missing, resource may be corrupted
   aimgr uninstall skill/some-skill
   aimgr repo sync --force
   aimgr install skill/some-skill
   ```

3. **Check tool-specific configuration**:
   ```bash
   # For Claude Code, check .claude/ exists
   ls -la .claude/
   
   # For OpenCode, check .opencode/ exists
   ls -la .opencode/
   ```

---

### Symlink Points to Wrong Location

**Symptoms:**
- Symlink exists but points to old/moved location
- Resource version mismatch

**Diagnosis:**
```bash
# Check where symlink points
readlink -f .claude/skills/some-skill

# Expected: ~/.local/share/ai-config/repo/skills/some-skill
# If different, symlink is outdated

# Compare versions
aimgr repo describe skill some-skill
cat .claude/skills/some-skill/version.txt
```

**Solutions:**

1. **Force reinstall**:
   ```bash
   aimgr install skill/some-skill --force
   ```

2. **Clean and reinstall**:
   ```bash
   # Remove all symlinks
   aimgr uninstall skill/some-skill
   
   # Update repository
   aimgr repo sync
   
   # Reinstall
   aimgr install skill/some-skill
   ```

---

## Configuration Issues

### Config File Missing

**Symptoms:**
- `config file not found`
- `failed to read config: no such file or directory`
- Fresh installation with no config

**Diagnosis:**
```bash
# Check if config exists
ls -la ~/.config/aimgr/aimgr.yaml

# Check config directory
ls -la ~/.config/aimgr/
```

**Solutions:**

1. **Create config directory**:
   ```bash
   mkdir -p ~/.config/aimgr
   ```

2. **Initialize default config**:
   ```bash
   # Set a basic config value to create file
   aimgr config set install.targets claude
   
   # Verify config created
   cat ~/.config/aimgr/aimgr.yaml
   ```

3. **Create config manually**:
   ```yaml
   # ~/.config/aimgr/aimgr.yaml
   install:
     targets: [claude]
   repo:
     path: ~/.local/share/ai-config/repo
   ```

---

### Invalid Config YAML

**Symptoms:**
- `failed to parse config: yaml: ...`
- `invalid YAML syntax`
- Config loads but values ignored

**Diagnosis:**
```bash
# View config file
cat ~/.config/aimgr/aimgr.yaml

# Check for YAML syntax errors
# Common issues:
# - Tabs instead of spaces
# - Incorrect indentation
# - Missing colons
# - Unquoted special characters
```

**Solutions:**

1. **Validate YAML syntax**:
   ```bash
   # Use online validator or yamllint
   yamllint ~/.config/aimgr/aimgr.yaml
   ```

2. **Fix common YAML issues**:
   ```bash
   # Convert tabs to spaces
   expand -t 2 ~/.config/aimgr/aimgr.yaml > /tmp/config.yaml
   mv /tmp/config.yaml ~/.config/aimgr/aimgr.yaml
   
   # Check indentation (must be consistent, 2 or 4 spaces)
   ```

3. **Reset to defaults**:
   ```bash
   # Backup broken config
   mv ~/.config/aimgr/aimgr.yaml ~/.config/aimgr/aimgr.yaml.bak
   
   # Create fresh config
   aimgr config set install.targets claude
   ```

4. **Example valid config**:
   ```yaml
   install:
     targets:
       - claude
       - opencode
   repo:
     path: /home/user/.local/share/ai-config/repo
   ```

---

### Wrong Default Target

**Symptoms:**
- Resources install to unexpected directory
- Using Claude but resources go to `.opencode/`
- Need to always specify `--target`

**Diagnosis:**
```bash
# Check current config
aimgr config get install.targets

# Check what directories exist
ls -d .claude .opencode .github 2>/dev/null
```

**Solutions:**

1. **Set preferred default**:
   ```bash
   # Set single target
   aimgr config set install.targets claude
   
   # Or multiple targets
   aimgr config set install.targets claude,opencode
   ```

2. **Verify setting**:
   ```bash
   aimgr config get install.targets
   cat ~/.config/aimgr/aimgr.yaml
   ```

3. **Use explicit target flag**:
   ```bash
   # Override default for specific install
   aimgr install skill/utils --target claude
   ```

---

## Repository Issues

### Resource Not Found

**Symptoms:**
- `resource not found in repository`
- `skill/command/agent 'name' does not exist`
- Can't install a resource you know exists

**Diagnosis:**
```bash
# List all resources in repository
aimgr repo list --format=json

# Search for specific resource
aimgr repo list | grep -i "resource-name"

# Check repository metadata
aimgr repo describe skill resource-name
```

**Solutions:**

1. **Update repository**:
   ```bash
   # Sync with latest sources
   aimgr repo sync
   
   # Force update if cached
   aimgr repo sync --force
   ```

2. **Check exact name** (case-sensitive):
   ```bash
   # List all skills
   aimgr repo list skill
   
   # Names must match exactly:
   # ✅ skill/react-testing
   # ❌ skill/React-Testing
   # ❌ skill/react_testing
   ```

3. **Add resource source**:
   ```bash
   # If resource is in external source
   aimgr repo import /path/to/resource
   aimgr repo import https://github.com/user/repo
   ```

4. **Verify repository path**:
   ```bash
   ls -la ~/.local/share/ai-config/repo/
   ls -la ~/.local/share/ai-config/repo/skills/
   ```

---

### Repository Update Fails

**Symptoms:**
- `aimgr repo sync` throws errors
- Git fetch/pull errors
- Source unreachable

**Diagnosis:**
```bash
# Check repository metadata
aimgr repo describe skill/resource-name
```

**Solutions:**

1. **Force update**:
   ```bash
   aimgr repo sync --force
   ```

2. **Check source accessibility**:
   ```bash
   # If using GitHub, verify you can access it
   curl -I https://github.com
   
   # If private repo, check SSH keys
   ssh -T git@github.com
   ```

3. **Remove and re-add source**:
   ```bash
   # Get current sources
   aimgr config get sync.sources
   
   # Remove problematic source
   aimgr repo remove https://github.com/user/repo
   
   # Re-add it
   aimgr repo import https://github.com/user/repo
   ```

4. **Manual git fix**:
   ```bash
   cd ~/.local/share/ai-config/repo/
   git fetch --all
   git reset --hard origin/main
   ```

---

### Duplicate Resources

**Symptoms:**
- Same resource appears multiple times
- Resources from different sources conflict
- Ambiguous resource name errors

**Diagnosis:**
```bash
# List all resources with full details
aimgr repo list --format=json | jq '.[] | select(.name=="duplicate-name")'

# Check sources
aimgr config get sync.sources
```

**Solutions:**

1. **Use qualified names**:
   ```bash
   # If same resource from multiple sources
   aimgr install source1:skill/utils
   aimgr install source2:skill/utils --target different-dir
   ```

2. **Remove duplicate source**:
   ```bash
   # Remove one source
   aimgr repo remove source-name
   aimgr repo sync
   ```

3. **Prioritize sources** (if supported):
   ```bash
   # Check config for source priority
   aimgr config get repository.source_priority
   ```

---

### Repository Corruption

**Symptoms:**
- Strange errors from `aimgr repo` commands
- Metadata inconsistent
- Resources partially missing

**Diagnosis:**
```bash
# Check repository integrity
ls -la ~/.local/share/ai-config/repo/

# Check for .git corruption
cd ~/.local/share/ai-config/repo/
git status
git fsck
```

**Solutions:**

1. **Rebuild repository**:
   ```bash
   # Backup current repository
   mv ~/.local/share/ai-config/repo ~/.local/share/ai-config/repo.bak
   
   # Initialize new repository
   mkdir -p ~/.local/share/ai-config/repo
   
   # Re-add sources
   aimgr repo import <your-sources>
   aimgr repo sync
   ```

2. **Fix git repository**:
   ```bash
   cd ~/.local/share/ai-config/repo/
   
   # Try to repair
   git fsck --full
   git gc --aggressive --prune=now
   
   # If unfixable, clone fresh
   cd ~/.local/share/ai-config/
   rm -rf repo
   git clone <source-url> repo
   ```

---

## Tool Detection Issues

### Wrong Tool Directory Selected

**Symptoms:**
- Resources install to unexpected directory
- Using Claude Code but resources go to `.opencode/`
- Multiple tool directories exist

**Diagnosis:**
```bash
# Check what directories exist
ls -d .claude .opencode .github/skills 2>/dev/null

# Check aimgr's detection logic
aimgr list  # Shows which directories it found

# Check config default
aimgr config get install.targets
```

**Solutions:**

1. **Specify target explicitly**:
   ```bash
   # Force installation to specific tool
   aimgr install skill/utils --target claude
   aimgr install command/test --target opencode
   ```

2. **Set config default**:
   ```bash
   # Set preferred tool
   aimgr config set install.targets claude
   ```

3. **Remove unwanted directories** (if you only use one tool):
   ```bash
   # If you only use Claude Code
   rm -rf .opencode/
   rm -rf .github/skills/
   
   # Now aimgr will default to .claude/
   ```

---

### Multiple Tool Directories Exist

**Symptoms:**
- Resources install to both `.claude/` and `.opencode/`
- Want to install to specific tool only
- Confusion about which tool loads which resources

**Diagnosis:**
```bash
# Check existing directories
ls -la .claude/ .opencode/ .github/skills/

# List installed resources by directory
ls -la .claude/skills/
ls -la .opencode/skills/
```

**Solutions:**

1. **This is expected behavior**:
   - By default, `aimgr` installs to ALL detected tool directories
   - This ensures resources work across different tools
   - Each tool only loads from its own directory

2. **Install to specific tool**:
   ```bash
   # Use --target flag
   aimgr install skill/claude-specific --target claude
   aimgr install skill/opencode-specific --target opencode
   ```

3. **Set default behavior**:
   ```bash
   # Install to single tool by default
   aimgr config set install.targets claude
   
   # Or specify multiple explicitly
   aimgr config set install.targets claude,opencode
   ```

4. **Remove from specific tool**:
   ```bash
   # Symlinks are safe to remove manually
   rm .opencode/skills/unwanted-skill
   
   # Or uninstall and reinstall to specific target
   aimgr uninstall skill/some-skill
   aimgr install skill/some-skill --target claude
   ```

---

### AI Tool Doesn't Load Resources

**Symptoms:**
- Resources installed correctly (symlinks exist)
- AI tool doesn't recognize/use them
- Skill not showing in tool's skill list

**Diagnosis:**
```bash
# Verify installation
aimgr list
ls -la .claude/skills/ .opencode/skills/

# Check symlinks are valid
file .claude/skills/some-skill
ls -la .claude/skills/some-skill/SKILL.md

# Verify resource structure
tree -L 2 .claude/skills/some-skill/
```

**Solutions:**

1. **Restart AI tool** (REQUIRED STEP):
   - Close the AI tool completely
   - For Claude Code: Close window/quit app
   - For VS Code/OpenCode: Close window, or quit VS Code
   - Reopen the tool
   - Resources are indexed at startup

2. **Check resource structure**:
   ```bash
   # Skills must have SKILL.md
   ls -la .claude/skills/*/SKILL.md
   
   # Commands must have COMMAND.md
   ls -la .claude/commands/*/COMMAND.md
   
   # Agents must have AGENT.md
   ls -la .claude/agents/*/AGENT.md
   ```

3. **Verify tool is looking in right directory**:
   - Claude Code looks in `.claude/`
   - OpenCode looks in `.opencode/`
   - GitHub Copilot looks in `.github/skills/`
   
   Make sure resources are in the correct directory.

4. **Check AI tool logs** (if available):
   ```bash
   # For Claude Code (example, path may vary)
   tail -f ~/Library/Logs/Claude/main.log
   
   # Look for skill loading errors
   ```

5. **Reinstall with verification**:
   ```bash
   # Clean uninstall
   aimgr uninstall skill/problem-skill
   
   # Verify removal
   ls -la .claude/skills/
   
   # Fresh install
   aimgr install skill/problem-skill
   
   # Verify structure
   ls -la .claude/skills/problem-skill/SKILL.md
   
   # Restart tool
   ```

---

## Still Having Issues?

If none of these solutions work:

1. **Check aimgr version**:
   ```bash
   aimgr --version
   
   # Expected output: aimgr version X.Y.Z (commit: abc1234, built: 2026-02-13T13:40:59Z)
   ```

2. **Enable debug mode** (if supported):
   ```bash
3. **Check system compatibility**:
   - OS: Linux, macOS, Windows WSL
   - Shell: bash, zsh, fish
   - File system: must support symlinks

4. **Gather diagnostic info**:
   ```bash
   # System info
   uname -a
   
   # aimgr version
   aimgr --version
   
   # Config
   cat ~/.config/aimgr/aimgr.yaml
   
   # Repository status
   aimgr repo list
   
   # Installed resources
   aimgr list
   
   # Directory structure
   tree -L 2 .claude .opencode
   ```

5. **File an issue** with diagnostic output

---

## Quick Reference

| Issue | First Action | Most Common Fix |
|-------|-------------|-----------------|
| `aimgr not found` | `which aimgr` | Add to PATH |
| Permission denied | `ls -la $(which aimgr)` | `chmod +x` |
| Broken symlinks | `ls -la .claude/skills/` | Reinstall resource |
| Skills not loading | Verify symlinks exist | **Restart AI tool** |
| Resource not found | `aimgr repo list` | `aimgr repo sync` |
| Config errors | `cat ~/.config/aimgr/aimgr.yaml` | Fix YAML syntax |
| Wrong directory | `ls -d .claude .opencode` | Use `--target` flag |
| Update fails | Check network | `aimgr repo sync --force` |

**Remember:** After installing or modifying resources, **always restart your AI tool** (Claude Code, OpenCode, etc.). Resources are loaded at startup, not dynamically.
