# Output Formats

`aimgr` supports multiple output formats to suit different use cases: human-readable tables for interactive use, JSON for scripting and automation, and YAML for configuration files.

## Table of Contents

- [Available Formats](#available-formats)
- [Using Output Formats](#using-output-formats)
- [Table Format (Default)](#table-format-default)
- [JSON Format](#json-format)
- [YAML Format](#yaml-format)
- [Error Reporting](#error-reporting)
- [Scripting Examples](#scripting-examples)

## Available Formats

| Format | Use Case | Flag |
|--------|----------|------|
| **table** | Human-readable, interactive use | `--format=table` (default) |
| **json** | Scripting, automation, CI/CD | `--format=json` |
| **yaml** | Configuration files, human-editable structured data | `--format=yaml` |

## Using Output Formats

The `--format` flag is supported by commands that perform bulk operations:

```bash
# Commands that support --format flag
aimgr repo add <source> --format=<format>
aimgr repo sync --format=<format>
aimgr repo list --format=<format>
aimgr list --format=<format>
```

## Table Format (Default)

The table format provides human-readable output ideal for interactive terminal use.

### Features

- **Structured layout**: Clean columns with headers
- **Color-coded status**: SUCCESS, SKIPPED, FAILED
- **Summary statistics**: Count of operations and resources
- **Error hints**: Suggestions for debugging failures

### Example: Adding Resources

```bash
$ aimgr repo add ~/my-resources/

┌─────────┬─────────────────────┬─────────┬──────────────────────┐
│ TYPE    │ NAME                │ STATUS  │ MESSAGE              │
├─────────┼─────────────────────┼─────────┼──────────────────────┤
│ skill   │ pdf-processing      │ SUCCESS │ Added to repository  │
│ skill   │ typescript-helper   │ SUCCESS │ Added to repository  │
│ command │ test                │ SUCCESS │ Added to repository  │
│ command │ build               │ SKIPPED │ Already exists       │
│ agent   │ code-reviewer       │ SUCCESS │ Added to repository  │
│ package │ web-tools           │ SUCCESS │ Added to repository  │
└─────────┴─────────────────────┴─────────┴──────────────────────┘

Summary: 5 added, 0 failed, 1 skipped (6 total)
```

### Example: With Errors

```bash
$ aimgr repo add ~/broken-resources/

┌─────────┬──────────────┬────────┬─────────────────────────────────────┐
│ TYPE    │ NAME         │ STATUS │ MESSAGE                             │
├─────────┼──────────────┼────────┼─────────────────────────────────────┤
│ skill   │ valid-skill  │ SUCCESS│ Added to repository                 │
│ skill   │ broken-skill │ FAILED │ missing required field: description │
│ command │ test         │ SUCCESS│ Added to repository                 │
└─────────┴──────────────┴────────┴─────────────────────────────────────┘

Summary: 2 added, 1 failed, 0 skipped (3 total)

⚠ Use --format=json to see detailed error messages
```

### Example: Listing Resources

```bash
$ aimgr repo list skill --format=table

┌───────────────────────┬─────────────────────────────────────┐
│ NAME                  │ DESCRIPTION                         │
├───────────────────────┼─────────────────────────────────────┤
│ pdf-processing        │ Extract and process PDF documents   │
│ typescript-helper     │ TypeScript development utilities    │
│ react-helper          │ React component development tools   │
└───────────────────────┴─────────────────────────────────────┘
```

## JSON Format

The JSON format provides structured, machine-readable output perfect for scripting, automation, and CI/CD pipelines.

### Features

- **Structured data**: Complete operation results
- **Detailed errors**: Full error messages and context
- **Easy parsing**: Standard JSON for use with `jq`, Python, etc.
- **Programmatic access**: All data available for scripts

### Output Structure

```json
{
  "added": [
    {
      "name": "pdf-processing",
      "type": "skill",
      "path": "/path/to/skills/pdf-processing",
      "message": ""
    }
  ],
  "skipped": [
    {
      "name": "build",
      "type": "command",
      "path": "/path/to/commands/build.md",
      "message": "already exists"
    }
  ],
  "failed": [
    {
      "name": "broken-skill",
      "type": "skill",
      "path": "/path/to/skills/broken-skill",
      "message": "missing required field: description"
    }
  ],
  "command_count": 2,
  "skill_count": 3,
  "agent_count": 1,
  "package_count": 1
}
```

### Example: Adding Resources

```bash
$ aimgr repo add ~/my-resources/ --format=json
{
  "added": [
    {
      "name": "pdf-processing",
      "type": "skill",
      "path": "/home/user/.local/share/ai-config/repo/skills/pdf-processing"
    },
    {
      "name": "test",
      "type": "command",
      "path": "/home/user/.local/share/ai-config/repo/commands/test.md"
    },
    {
      "name": "code-reviewer",
      "type": "agent",
      "path": "/home/user/.local/share/ai-config/repo/agents/code-reviewer.md"
    }
  ],
  "skipped": [],
  "failed": [],
  "command_count": 1,
  "skill_count": 1,
  "agent_count": 1,
  "package_count": 0
}
```

### Example: Listing Resources

```bash
$ aimgr repo list skill --format=json
[
  {
    "type": "skill",
    "name": "pdf-processing",
    "description": "Extract and process PDF documents",
    "version": "1.0.0",
    "author": "Team",
    "license": "MIT"
  },
  {
    "type": "skill",
    "name": "typescript-helper",
    "description": "TypeScript development utilities",
    "version": "2.1.0",
    "license": "Apache-2.0"
  }
]
```

## YAML Format

The YAML format provides human-readable structured output that's easy to edit and version control.

### Features

- **Human-readable**: More readable than JSON
- **Comments supported**: Can add documentation
- **Configuration-friendly**: Easy to use in config files
- **Version control**: Clean diffs in git

### Output Structure

```yaml
added:
  - name: pdf-processing
    type: skill
    path: /path/to/skills/pdf-processing
  - name: test
    type: command
    path: /path/to/commands/test.md
skipped:
  - name: build
    type: command
    path: /path/to/commands/build.md
    message: already exists
failed: []
command_count: 2
skill_count: 1
agent_count: 0
package_count: 0
```

### Example: Adding Resources

```bash
$ aimgr repo add ~/my-resources/ --format=yaml
added:
  - name: pdf-processing
    type: skill
    path: /home/user/.local/share/ai-config/repo/skills/pdf-processing
  - name: typescript-helper
    type: skill
    path: /home/user/.local/share/ai-config/repo/skills/typescript-helper
  - name: code-reviewer
    type: agent
    path: /home/user/.local/share/ai-config/repo/agents/code-reviewer.md
skipped: []
failed: []
command_count: 0
skill_count: 2
agent_count: 1
package_count: 0
```

### Example: Saving to File

```bash
# Save import results for auditing
$ aimgr repo add gh:myorg/resources --format=yaml > import-log.yaml

# Review what was imported
$ cat import-log.yaml
```

## Error Reporting

All output formats include detailed error reporting to help diagnose issues.

### Table Format Errors

```bash
$ aimgr repo add ~/resources/ --format=table

┌─────────┬──────────────┬────────┬─────────────────────────────────────┐
│ TYPE    │ NAME         │ STATUS │ MESSAGE                             │
├─────────┼──────────────┼────────┼─────────────────────────────────────┤
│ skill   │ valid-skill  │ SUCCESS│ Added to repository                 │
│ skill   │ no-desc      │ FAILED │ missing required field: description │
│ command │ invalid-yaml │ FAILED │ failed to parse YAML frontmatter    │
│ agent   │ bad-name!    │ FAILED │ invalid name format                 │
└─────────┴──────────────┴────────┴─────────────────────────────────────┘

Summary: 1 added, 3 failed, 0 skipped (4 total)

⚠ Use --format=json to see detailed error messages
```

### JSON Format Errors

JSON format provides the most detailed error information:

```json
{
  "added": [
    {
      "name": "valid-skill",
      "type": "skill",
      "path": "/path/to/skills/valid-skill"
    }
  ],
  "skipped": [],
  "failed": [
    {
      "name": "no-desc",
      "type": "skill",
      "path": "/path/to/skills/no-desc",
      "message": "missing required field: description"
    },
    {
      "name": "invalid-yaml",
      "type": "command",
      "path": "/path/to/commands/invalid-yaml.md",
      "message": "failed to parse YAML frontmatter: yaml: line 5: mapping values are not allowed in this context"
    },
    {
      "name": "bad-name!",
      "type": "agent",
      "path": "/path/to/agents/bad-name!.md",
      "message": "invalid name format: must be lowercase alphanumeric with hyphens only"
    }
  ],
  "command_count": 0,
  "skill_count": 1,
  "agent_count": 0,
  "package_count": 0
}
```

### Common Error Messages

| Error | Meaning | Solution |
|-------|---------|----------|
| `missing required field: description` | Resource missing description in frontmatter | Add `description:` field to YAML frontmatter |
| `failed to parse YAML frontmatter` | Invalid YAML syntax | Check YAML syntax, ensure proper indentation |
| `invalid name format` | Resource name doesn't follow naming rules | Use lowercase, alphanumeric, hyphens only |
| `already exists` | Resource with same name exists | Use `--force` to overwrite or `--skip-existing` |
| `source not found` | Path or repository doesn't exist | Check path spelling and existence |

## Scripting Examples

### Using jq with JSON Output

**Extract only successful additions:**
```bash
$ aimgr repo add ~/resources/ --format=json | jq '.added[].name'
"pdf-processing"
"typescript-helper"
"code-reviewer"
```

**Count resources by type:**
```bash
$ aimgr repo add ~/resources/ --format=json | jq '{
  skills: .skill_count,
  commands: .command_count,
  agents: .agent_count,
  packages: .package_count
}'
{
  "skills": 2,
  "commands": 1,
  "agents": 1,
  "packages": 0
}
```

**Get error messages:**
```bash
$ aimgr repo add ~/resources/ --format=json | jq '.failed[] | {
  name: .name,
  error: .message
}'
{
  "name": "broken-skill",
  "error": "missing required field: description"
}
```

**Check if any operations failed:**
```bash
$ aimgr repo add ~/resources/ --format=json | jq '.failed | length'
2

# Use in scripts
if [ $(aimgr repo add ~/resources/ --format=json | jq '.failed | length') -gt 0 ]; then
  echo "Import failed!"
  exit 1
fi
```

**Filter by resource type:**
```bash
$ aimgr repo add ~/resources/ --format=json | jq '.added[] | select(.type == "skill")'
{
  "name": "pdf-processing",
  "type": "skill",
  "path": "/path/to/skills/pdf-processing"
}
```

### CI/CD Integration

**GitHub Actions example:**
```yaml
- name: Import AI Resources
  id: import
  run: |
    output=$(aimgr repo add gh:myorg/resources --format=json)
    echo "$output" > import-results.json
    
    # Check for failures
    failed=$(echo "$output" | jq '.failed | length')
    if [ "$failed" -gt 0 ]; then
      echo "::error::Failed to import $failed resources"
      echo "$output" | jq '.failed[]'
      exit 1
    fi
    
    # Report summary
    added=$(echo "$output" | jq '.added | length')
    echo "::notice::Successfully imported $added resources"

- name: Upload import results
  uses: actions/upload-artifact@v3
  with:
    name: import-results
    path: import-results.json
```

**GitLab CI example:**
```yaml
import-resources:
  script:
    - output=$(aimgr repo add gh:myorg/resources --format=json)
    - echo "$output" > import-results.json
    - |
      if [ $(echo "$output" | jq '.failed | length') -gt 0 ]; then
        echo "Import failed!"
        echo "$output" | jq '.failed[]'
        exit 1
      fi
  artifacts:
    paths:
      - import-results.json
    reports:
      dotenv: import-results.json
```

### Python Integration

```python
#!/usr/bin/env python3
import json
import subprocess
import sys

def import_resources(source_path):
    """Import resources and return parsed results."""
    result = subprocess.run(
        ["aimgr", "repo", "add", source_path, "--format=json"],
        capture_output=True,
        text=True
    )
    
    if result.returncode != 0:
        print(f"Error running aimgr: {result.stderr}", file=sys.stderr)
        sys.exit(1)
    
    return json.loads(result.stdout)

def main():
    # Import resources
    results = import_resources("~/my-resources/")
    
    # Check for failures
    if results["failed"]:
        print("Failed to import the following resources:")
        for failure in results["failed"]:
            print(f"  - {failure['name']} ({failure['type']}): {failure['message']}")
        sys.exit(1)
    
    # Report success
    print(f"Successfully imported {len(results['added'])} resources:")
    for resource in results["added"]:
        print(f"  - {resource['name']} ({resource['type']})")
    
    # Print statistics
    print(f"\nStatistics:")
    print(f"  Commands: {results['command_count']}")
    print(f"  Skills: {results['skill_count']}")
    print(f"  Agents: {results['agent_count']}")
    print(f"  Packages: {results['package_count']}")

if __name__ == "__main__":
    main()
```

### Shell Script Integration

```bash
#!/bin/bash
# import-and-verify.sh - Import resources with error handling

set -euo pipefail

SOURCE_PATH="${1:-}"
if [ -z "$SOURCE_PATH" ]; then
  echo "Usage: $0 <source-path>" >&2
  exit 1
fi

# Import resources
echo "Importing resources from $SOURCE_PATH..."
output=$(aimgr repo add "$SOURCE_PATH" --format=json)

# Parse results using jq
added_count=$(echo "$output" | jq '.added | length')
failed_count=$(echo "$output" | jq '.failed | length')
skipped_count=$(echo "$output" | jq '.skipped | length')

# Report results
echo "Results:"
echo "  Added: $added_count"
echo "  Failed: $failed_count"
echo "  Skipped: $skipped_count"

# Handle failures
if [ "$failed_count" -gt 0 ]; then
  echo ""
  echo "Failed resources:"
  echo "$output" | jq -r '.failed[] | "  - \(.name) (\(.type)): \(.message)"'
  exit 1
fi

# List added resources
if [ "$added_count" -gt 0 ]; then
  echo ""
  echo "Successfully added:"
  echo "$output" | jq -r '.added[] | "  - \(.name) (\(.type))"'
fi

echo ""
echo "Import completed successfully!"
```

### Monitoring and Alerting

**Check for import errors in cron job:**
```bash
#!/bin/bash
# /etc/cron.daily/sync-ai-resources

LOG_FILE="/var/log/aimgr-sync.log"

output=$(aimgr repo sync --format=json 2>&1)
failed=$(echo "$output" | jq '.failed | length')

if [ "$failed" -gt 0 ]; then
  echo "$(date): Resource sync failed" >> "$LOG_FILE"
  echo "$output" | jq '.failed[]' >> "$LOG_FILE"
  
  # Send alert (example using mail)
  echo "$output" | jq '.failed[]' | mail -s "aimgr sync failed" admin@example.com
  exit 1
fi

echo "$(date): Resource sync successful ($added added)" >> "$LOG_FILE"
```

## Best Practices

### When to Use Each Format

**Use TABLE format when:**
- Working interactively in a terminal
- You want a quick overview of results
- Human readability is the priority
- You're exploring or debugging manually

**Use JSON format when:**
- Writing scripts or automation
- Integrating with CI/CD pipelines
- You need detailed error messages
- Processing output programmatically
- Generating reports or logs

**Use YAML format when:**
- Creating configuration files
- You want human-readable structured data
- Version controlling import results
- Documenting import operations
- Generating audit trails

### Error Handling

1. **Always check for failures** in scripts:
   ```bash
   output=$(aimgr repo add ~/resources/ --format=json)
   if [ $(echo "$output" | jq '.failed | length') -gt 0 ]; then
     echo "Import failed!"
     exit 1
   fi
   ```

2. **Log detailed errors** in automation:
   ```bash
   aimgr repo add ~/resources/ --format=json | \
     jq '.failed[]' > error-log.json
   ```

3. **Use descriptive error messages** for debugging:
   ```bash
   aimgr repo add ~/resources/ --format=json | \
     jq -r '.failed[] | "\(.name): \(.message)"'
   ```

### Performance Considerations

- JSON and YAML formats have similar performance
- Table format may be slightly slower for large outputs (formatting overhead)
- For batch operations, JSON is recommended for programmatic processing
- Use `--dry-run` with any format to preview without making changes
