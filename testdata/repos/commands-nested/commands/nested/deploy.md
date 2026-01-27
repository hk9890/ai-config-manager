---
description: Deploy application to production environment
agent: deploy-agent
model: anthropic/claude-3-5-sonnet-20241022
---

# Deploy Command

This command deploys the application to the production environment.

## Purpose

Tests nested command discovery - this command is in a subdirectory within the commands folder.

## Usage

```
/deploy --environment production
```

## Deployment Steps

1. Run pre-deployment checks
2. Build production artifacts
3. Upload to servers
4. Run post-deployment verification
