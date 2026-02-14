// Package repomanifest provides functionality for managing ai.repo.yaml manifest files.
//
// The ai.repo.yaml manifest tracks synced sources in the repository and is git-tracked.
// It provides operations to load, save, and manage sources.
//
// Example ai.repo.yaml:
//
//	version: 1
//	sources:
//	  - name: my-local-commands
//	    path: /home/user/my-resources
//	    mode: symlink
//	    added: 2026-02-14T10:30:00Z
//	    last_synced: 2026-02-14T15:45:00Z
//	  - name: agentskills-catalog
//	    url: https://github.com/agentskills/catalog
//	    ref: main
//	    subpath: resources
//	    mode: copy
//	    added: 2026-02-14T11:00:00Z
//	    last_synced: 2026-02-14T15:45:00Z
//
// Usage:
//
//	// Load manifest (returns empty manifest if file doesn't exist)
//	manifest, err := repomanifest.Load("/path/to/repo")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Add a source
//	source := &repomanifest.Source{
//	    Name: "my-source",
//	    Path: "/path/to/resources",
//	    Mode: "symlink",
//	}
//	if err := manifest.AddSource(source); err != nil {
//	    log.Fatal(err)
//	}
//
//	// Save manifest
//	if err := manifest.Save("/path/to/repo"); err != nil {
//	    log.Fatal(err)
//	}
package repomanifest
