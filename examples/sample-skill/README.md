# Sample Skill

This is an example skill demonstrating the agentskills.io format.

## Contents

- **SKILL.md**: Main skill file with metadata and documentation
- **scripts/**: Directory containing example scripts
- **references/**: Additional documentation and references

## Installation

```bash
# Add to repository
ai-repo add skill examples/sample-skill

# Install in a project
cd your-project/
ai-repo install skill sample-skill
```

## Directory Structure

```
sample-skill/
├── SKILL.md           # Main skill file (required)
├── README.md          # This file
├── scripts/           # Optional scripts
│   └── example.sh     # Example script
└── references/        # Optional documentation
    └── REFERENCE.md   # Example reference
```

## Requirements

The skill name in SKILL.md must match the directory name (`sample-skill`).
