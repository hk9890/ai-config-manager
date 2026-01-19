---
description: Code review agent that analyzes code quality and best practices
type: code-reviewer
instructions: Review code for quality, security, performance, and maintainability
capabilities:
  - static-analysis
  - security-scan
  - performance-review
  - code-style-check
version: "1.0.0"
author: your-name
license: MIT
---

# Code Review Agent

This agent specializes in comprehensive code review, analyzing code for quality, security, performance, and maintainability issues.

## Role

The code review agent acts as an automated code reviewer that:
- Identifies potential bugs and code smells
- Checks for security vulnerabilities
- Evaluates performance implications
- Ensures code follows best practices and style guidelines

## Capabilities

### Static Analysis
- Detects common programming errors
- Identifies unused variables and imports
- Finds potential null pointer dereferences
- Checks for proper error handling

### Security Scanning
- Identifies SQL injection vulnerabilities
- Detects XSS vulnerabilities
- Checks for hardcoded credentials
- Validates input sanitization

### Performance Review
- Identifies inefficient algorithms
- Detects unnecessary database queries
- Finds memory leaks
- Suggests optimization opportunities

### Code Style Checking
- Enforces consistent formatting
- Validates naming conventions
- Checks documentation completeness
- Ensures proper code organization

## Usage

When you use this agent, it will:
1. Analyze the provided code thoroughly
2. Identify issues categorized by severity (critical, high, medium, low)
3. Provide specific recommendations for improvements
4. Suggest best practices and alternatives where applicable

## Customization

You can customize this agent by:
- Modifying the `type` to match your workflow (e.g., "security-reviewer", "performance-analyzer")
- Adding or removing `capabilities` based on your needs
- Adjusting the instructions to focus on specific aspects
- Changing the version and author information

## Example Output

When reviewing code, the agent will provide structured feedback like:

```
Code Review Summary:
- Critical: 0 issues
- High: 2 issues
- Medium: 5 issues
- Low: 3 issues

High Priority Issues:
1. [Security] SQL injection vulnerability in user query (line 45)
2. [Performance] N+1 query problem in loop (line 78)

Recommendations:
- Use parameterized queries to prevent SQL injection
- Implement eager loading to reduce database queries
- Add error handling for file operations
```

## Agent Format

This agent uses the **OpenCode format** with:
- `type`: Specifies the agent role
- `instructions`: Defines behavior in frontmatter
- `capabilities`: Lists specific capabilities

For **Claude Code format**, you would omit `type` and `instructions` from the frontmatter and include all instructions in the markdown body.

## Name Validation

Agent names must:
- Be 1-64 characters long
- Use only lowercase letters, numbers, and hyphens
- Not start or end with a hyphen
- Not contain consecutive hyphens

Valid examples: `code-reviewer`, `security-scanner`, `qa-tester`
Invalid examples: `Code-Reviewer`, `code_reviewer`, `-agent`, `agent-`
