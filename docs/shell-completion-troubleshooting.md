# Shell Completion Setup for aimgr

## The Issue

When you type `./aimgr install skill/<TAB>`, completion doesn't work because:

1. Shell completion only works for commands in your `$PATH` (like `aimgr`), not relative paths like `./aimgr`
2. You need to source the completion script first

## Solution

### Step 1: Make sure aimgr is in your PATH

```bash
# Check if ~/bin is in PATH
echo $PATH | grep -q "$HOME/bin" && echo "✓ ~/bin is in PATH" || echo "✗ ~/bin NOT in PATH"

# If not in PATH, add to ~/.bashrc (or ~/.zshrc for zsh)
echo 'export PATH="$HOME/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

### Step 2: Load the completion script

**For Bash (current session only):**
```bash
source <(aimgr completion bash)
```

**For Bash (permanent - add to ~/.bashrc):**
```bash
# Add this line to ~/.bashrc
eval "$(aimgr completion bash)"

# Then reload
source ~/.bashrc
```

**For Zsh (permanent - add to ~/.zshrc):**
```bash
# First, enable completions if not already enabled
echo "autoload -U compinit; compinit" >> ~/.zshrc

# Generate completion file
aimgr completion zsh > "${fpath[1]}/_aimgr"

# Reload shell
exec zsh
```

### Step 3: Test completion

Now you can use completion (without the `./`):

```bash
aimgr install skill/<TAB>
# Should show: skill-creator, install-skill, multi-skill, etc.

aimgr install command/<TAB>
# Should show available commands

aimgr install agent/<TAB>
# Should show available agents
```

## Why `./aimgr` doesn't work with completion

Shell completion frameworks (bash-completion, zsh completion) only intercept completion for commands found via `$PATH` lookup. When you type `./aimgr`, the shell treats it as a file path, not a command, so completion isn't triggered.

**Works:**
```bash
aimgr install skill/<TAB>  # ✓ Uses completion
```

**Doesn't work:**
```bash
./aimgr install skill/<TAB>  # ✗ Bypasses completion
```

## Verification

Test that completion is working:

```bash
# This should show completions:
aimgr __complete install skill/

# Output should include:
# skill/skill-creator
# :4
# Completion ended with directive: ShellCompDirectiveNoFileComp
```

## Troubleshooting

**1. "aimgr: command not found"**
   - Make sure `~/bin` is in your `$PATH`
   - Check: `ls -l ~/bin/aimgr`
   - Add to PATH: `export PATH="$HOME/bin:$PATH"`

**2. Completion still not working after setup**
   - Restart your shell or run: `exec bash` (or `exec zsh`)
   - Verify completion is loaded: `complete -p aimgr` (bash) or `which _aimgr` (zsh)

**3. Want to use `./aimgr` during development**
   - Use the installed version instead: `make install` then use `aimgr`
   - Or create an alias: `alias a='./aimgr'` then use `a install skill/<TAB>`
