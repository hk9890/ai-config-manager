# Security Policy

## Supported Versions

We release patches for security vulnerabilities for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 1.4.x   | :white_check_mark: |
| 1.3.x   | :white_check_mark: |
| 1.2.x   | :x:                |
| 1.1.x   | :x:                |
| 1.0.x   | :x:                |
| < 1.0   | :x:                |

## Reporting a Vulnerability

We take the security of `aimgr` seriously. If you discover a security vulnerability, please follow these steps:

### How to Report

**Email:** [hans.kohlreiter@dynatrace.com](mailto:hans.kohlreiter@dynatrace.com)

**Please include:**
- A description of the vulnerability
- Steps to reproduce the issue
- Potential impact of the vulnerability
- Any suggested fixes (if applicable)

### What to Expect

- **Initial Response:** You will receive an acknowledgment within 48 hours
- **Assessment:** We will investigate and assess the severity within 5 business days
- **Updates:** You will receive regular updates on our progress
- **Resolution:** We aim to release a fix within 30 days for critical vulnerabilities
- **Credit:** We will credit you in the security advisory (unless you prefer to remain anonymous)

### Security Advisory Process

1. **Confirmation:** We validate the reported vulnerability
2. **Fix Development:** A patch is developed and tested
3. **Coordinated Disclosure:** We coordinate the release timing with you
4. **Public Disclosure:** A security advisory is published with the fix
5. **CVE Assignment:** Critical vulnerabilities receive a CVE identifier

## Security Best Practices

When using `aimgr`, follow these best practices to maintain security:

### Safe Resource Management

1. **Verify Sources Before Adding:**
   ```bash
   # Review repository contents before importing
   aimgr repo import gh:owner/repo --dry-run
   
   # Check resource details before installing
   aimgr repo describe skill resource-name
   ```

2. **Use Trusted Sources:**
   - Only add resources from repositories you trust
   - Review resource content before installation
   - Be cautious with resources that execute system commands

3. **Regularly Update Resources:**
   ```bash
   # Keep your resources up-to-date
   aimgr repo sync
   ```

### Repository Security

1. **Protect Your Repository Path:**
   - The default repository location is `~/.local/share/ai-config/repo/`
   - Ensure appropriate file permissions (755 for directories, 644 for files)
   - Do not store the repository in world-writable locations

2. **Configuration File Security:**
   - Configuration is stored at `~/.config/aimgr/aimgr.yaml`
   - Do not commit secrets or credentials to configuration files
   - Use environment variables for sensitive values:
     ```yaml
     repo:
       path: ${AIMGR_REPO_PATH:-~/.local/share/ai-config/repo}
     ```

3. **Workspace Cache:**
   - Git repository caches are stored in `.workspace/` within your repository
   - Regularly prune unused caches: `aimgr repo prune`
   - The workspace directory should not be world-writable

### Installation Security

1. **Review Before Installing:**
   ```bash
   # Check what a resource does before installing
   cat ~/.local/share/ai-config/repo/commands/command-name.md
   cat ~/.local/share/ai-config/repo/skills/skill-name/SKILL.md
   ```

2. **Use Symlink Mode (Default):**
   - `aimgr` uses symlinks by default, which allows you to review and update resources centrally
   - Any changes to the repository immediately affect all installations

3. **Limit Tool Access:**
   - Only install resources to the AI tools you actually use
   - Configure default targets: `aimgr config set install.targets claude`

### Git Repository Security

1. **SSH vs HTTPS:**
   - Use SSH URLs for private repositories: `git@github.com:owner/repo.git`
   - Use HTTPS for public repositories: `https://github.com/owner/repo.git`

2. **Verify Repository Authenticity:**
   - Check the repository owner and contents before importing
   - Be cautious with repositories that have few stars or recent creation dates

3. **Use Specific Versions:**
   ```bash
   # Pin to specific tags/versions instead of main/master
   aimgr repo import gh:owner/repo@v1.0.0
   ```

### Command Execution Safety

`aimgr` commands and skills may contain shell commands that are executed by AI tools. Follow these guidelines:

1. **Review Command Content:**
   - Always review command files before installation
   - Commands are markdown files with embedded shell scripts
   - Look for potentially dangerous commands (rm, dd, curl | bash, etc.)

2. **Skills with Custom Scripts:**
   - Skills may include scripts in the `scripts/` directory
   - Review all scripts before installing skills
   - Be especially cautious with skills that modify system files

3. **Agents with System Access:**
   - Agents may have extensive system access depending on the AI tool
   - Only install agents from sources you completely trust
   - Review agent capabilities in their metadata

### Package Security

When using packages (collections of resources):

1. **Review Package Contents:**
   ```bash
   # Check what resources a package includes
   aimgr repo describe package package-name
   ```

2. **Verify All Resources:**
   - A package is only as secure as its least secure resource
   - Review each resource referenced by the package

### Environment Variables

1. **Never Store Secrets in Plain Text:**
   - Do not put API keys, passwords, or tokens in configuration files
   - Use environment variables or secure credential storage

2. **Validate Environment Variables:**
   - Sanitize user input when using environment variables in commands
   - Be aware of environment variable injection risks

## Known Security Considerations

### Symlink Attacks

- `aimgr` creates symlinks from projects to the central repository
- Symlinks are created with proper path validation
- The tool checks for existing files/directories before creating symlinks
- Broken symlinks are handled gracefully

### Path Traversal

- All file paths are normalized and validated
- Path traversal attempts (../, ..\, etc.) are blocked
- Resources cannot be installed outside designated directories

### Arbitrary Code Execution

- Resources (commands, skills, agents) may contain code that is executed by AI tools
- `aimgr` itself does not execute resource content
- Users are responsible for reviewing resource content before installation
- We recommend using code review practices for resources from untrusted sources

## Security Updates

Security updates are released as soon as possible after a vulnerability is confirmed. Update methods:

```bash
# Update via go install
go install github.com/hk9890/ai-config-manager@latest

# Or rebuild from source
cd ai-config-manager
git pull
make install
```

Subscribe to releases on GitHub to be notified of security updates:
https://github.com/hk9890/ai-config-manager/releases

## Responsible Disclosure

We appreciate the security research community's efforts to responsibly disclose vulnerabilities. We commit to:

- Acknowledge your report within 48 hours
- Provide regular updates on our progress
- Credit you in the security advisory (with your permission)
- Work with you on a coordinated disclosure timeline
- Not take legal action against researchers who follow this policy

Thank you for helping keep `aimgr` and its users safe!
