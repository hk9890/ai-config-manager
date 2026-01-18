---
description: Run tests with coverage report
agent: build
model: anthropic/claude-3-5-sonnet-20241022
---

# Run Tests with Coverage

This command runs your project's test suite and generates a coverage report.

## What it does

When you run this command, the AI will:
1. Identify your project's testing framework
2. Run the appropriate test command with coverage enabled
3. Generate and display a coverage report
4. Suggest areas that need more test coverage

## Customization

You can customize this command by:
- Changing the `agent` to match your workflow (e.g., "test", "qa", "build")
- Modifying the `model` to use a different AI model
- Adding `allowed-tools` to restrict what the AI can do

Example with restrictions:
```yaml
---
description: Run tests with coverage report
agent: test
model: anthropic/claude-3-5-sonnet-20241022
allowed-tools:
  - bash
  - read
  - write
---
```

## Name Validation

Command names must:
- Be 1-64 characters long
- Use only lowercase letters, numbers, and hyphens
- Not start or end with a hyphen
- Not contain consecutive hyphens

Valid examples: `test`, `run-coverage`, `test-integration`
Invalid examples: `Test`, `test_coverage`, `-test`, `test-`
