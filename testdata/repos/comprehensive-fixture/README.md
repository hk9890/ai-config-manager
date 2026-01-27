# Comprehensive Test Fixture

This directory contains a comprehensive test fixture for verifying the import and on-disk layout functionality of aimgr.

## Structure

```
comprehensive-fixture/
├── commands/
│   ├── api/                    # Nested commands
│   │   ├── deploy.md
│   │   └── status.md
│   ├── db/                     # Another nested directory
│   │   └── migrate.md
│   └── test.md                 # Flat command at root
├── skills/
│   ├── skill-one/              # Flat skills (always directories)
│   └── skill-two/
├── agents/
│   ├── agent-one.md            # Flat agents (always single files)
│   └── agent-two.md
└── packages/
    └── test-package.package.json
```

## Usage

This fixture is used by integration tests to verify:

1. **File Layout**: Commands are stored with nested directory structure preserved
2. **Metadata Files**: Metadata filenames have slashes escaped (e.g., `api-deploy-metadata.json`)
3. **Metadata Content**: Metadata name field contains full nested path (e.g., `"name": "api/deploy"`)

## Expected On-Disk Layout After Import

When imported into a repository, the structure should be:

```
repo/
├── commands/
│   ├── api/
│   │   ├── deploy.md
│   │   └── status.md
│   ├── db/
│   │   └── migrate.md
│   └── test.md
├── .metadata/
│   └── commands/
│       ├── api-deploy-metadata.json      (name: "api/deploy")
│       ├── api-status-metadata.json      (name: "api/status")
│       ├── db-migrate-metadata.json      (name: "db/migrate")
│       └── test-metadata.json            (name: "test")
```

## Test Commands

```bash
# Run integration test
go test -v -tags=integration ./test -run TestImportLayout

# Manual verification
aimgr repo import testdata/repos/comprehensive-fixture
aimgr repo list --type command
```
