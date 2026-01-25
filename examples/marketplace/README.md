# Marketplace Examples

This directory contains example Claude plugin marketplace configurations and plugin structures.

## Files

- **marketplace.json**: Full-featured marketplace with multiple plugins
- **minimal-marketplace.json**: Minimal marketplace with only required fields
- **plugins/**: Example plugin directory structures

## Marketplace Structure

A Claude marketplace defines a collection of plugins with resources (commands, skills, agents).

### marketplace.json

The main marketplace configuration file with:
- Marketplace metadata (name, version, description)
- Owner information
- Plugin definitions with source paths

### Plugin Structure

Each plugin in the `plugins/` directory can contain:
- `commands/*.md` - Slash commands
- `skills/*/SKILL.md` - Agent skills
- `agents/*.md` - AI agents
- `.claude/` or `.opencode/` subdirectories (optional)

## Example Plugins

### web-dev-tools

Web development toolkit with:
- **Commands**: `build`, `dev`
- **Skills**: `typescript-helper`
- **Agents**: `code-reviewer`

### testing-suite

Testing toolkit with:
- **Commands**: `test`, `coverage`
- **Skills**: `test-generator`
- **Agents**: `qa-tester`

## Usage

Import the example marketplace:

```bash
# From this directory
aimgr marketplace import marketplace.json

# Preview without importing
aimgr marketplace import marketplace.json --dry-run

# Import specific plugins
aimgr marketplace import marketplace.json --filter "web-*"
```

After import, packages are created:

```bash
# List imported packages
aimgr repo list package

# Install a package
aimgr install package/web-dev-tools
aimgr install package/testing-suite
```

## Creating Your Own Marketplace

1. **Create marketplace.json**:
   ```json
   {
     "name": "my-marketplace",
     "description": "My plugin collection",
     "plugins": [
       {
         "name": "my-plugin",
         "description": "Plugin description",
         "source": "./plugins/my-plugin"
       }
     ]
   }
   ```

2. **Organize plugin resources**:
   ```
   plugins/my-plugin/
   ├── commands/
   │   └── my-command.md
   ├── skills/
   │   └── my-skill/
   │       └── SKILL.md
   └── agents/
       └── my-agent.md
   ```

3. **Import the marketplace**:
   ```bash
   aimgr marketplace import marketplace.json
   ```

## Format Specification

See [AGENTS.md](../../AGENTS.md) for detailed format specification including:
- Required and optional fields
- JSON schema
- Code structure
- Testing guidelines

## See Also

- [Package Examples](../packages/) - Individual package examples
- [Command Examples](../sample-command.md) - Command format
- [Skill Examples](../sample-skill/) - Skill format
- [Agent Examples](../sample-agent.md) - Agent format
