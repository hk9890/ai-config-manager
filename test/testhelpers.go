package test

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// createTestCommand creates a command in a properly structured commands/ directory.
// This ensures the resource name is correctly derived and valid.
func createTestCommand(t *testing.T, name, description string) string {
	tempDir := t.TempDir()
	commandsDir := filepath.Join(tempDir, "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatalf("Failed to create commands directory: %v", err)
	}

	cmdPath := filepath.Join(commandsDir, name+".md")
	cmdContent := fmt.Sprintf(`---
description: %s
---
# %s
Test command content.
`, description, name)

	if err := os.WriteFile(cmdPath, []byte(cmdContent), 0644); err != nil {
		t.Fatalf("Failed to create command file: %v", err)
	}

	return cmdPath
}

// createTestSkill creates a skill in a properly structured skills/ directory.
// This ensures the resource name is correctly derived and valid.
func createTestSkill(t *testing.T, name, description string) string {
	tempDir := t.TempDir()
	skillsDir := filepath.Join(tempDir, "skills", name)
	if err := os.MkdirAll(filepath.Join(skillsDir, "scripts"), 0755); err != nil {
		t.Fatalf("Failed to create skill directory: %v", err)
	}

	skillMdPath := filepath.Join(skillsDir, "SKILL.md")
	skillContent := fmt.Sprintf(`---
name: %s
description: %s
---
# %s
Test skill content.
`, name, description, name)

	if err := os.WriteFile(skillMdPath, []byte(skillContent), 0644); err != nil {
		t.Fatalf("Failed to create SKILL.md: %v", err)
	}

	return filepath.Dir(skillMdPath) // Return the skill directory
}

// createTestAgent creates an agent in a properly structured agents/ directory.
// This ensures the resource name is correctly derived and valid.
func createTestAgent(t *testing.T, name, description string) string {
	tempDir := t.TempDir()
	agentsDir := filepath.Join(tempDir, "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("Failed to create agents directory: %v", err)
	}

	agentPath := filepath.Join(agentsDir, name+".md")
	agentContent := fmt.Sprintf(`---
description: %s
---
# %s
Test agent content.
`, description, name)

	if err := os.WriteFile(agentPath, []byte(agentContent), 0644); err != nil {
		t.Fatalf("Failed to create agent file: %v", err)
	}

	return agentPath
}

// createTestCommandInDir creates a command in the specified base directory.
// The command will be created at baseDir/commands/<name>.md
func createTestCommandInDir(t *testing.T, baseDir, name, description string) string {
	commandsDir := filepath.Join(baseDir, "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatalf("Failed to create commands directory: %v", err)
	}

	cmdPath := filepath.Join(commandsDir, name+".md")
	cmdContent := fmt.Sprintf(`---
description: %s
---
# %s
Test command content.
`, description, name)

	if err := os.WriteFile(cmdPath, []byte(cmdContent), 0644); err != nil {
		t.Fatalf("Failed to create command file: %v", err)
	}

	return cmdPath
}

// createTestSkillInDir creates a skill in the specified base directory.
// The skill will be created at baseDir/skills/<name>/SKILL.md
func createTestSkillInDir(t *testing.T, baseDir, name, description string) string {
	skillsDir := filepath.Join(baseDir, "skills", name)
	if err := os.MkdirAll(filepath.Join(skillsDir, "scripts"), 0755); err != nil {
		t.Fatalf("Failed to create skill directory: %v", err)
	}

	skillMdPath := filepath.Join(skillsDir, "SKILL.md")
	skillContent := fmt.Sprintf(`---
name: %s
description: %s
license: MIT
---
# %s
Test skill content.
`, name, description, name)

	if err := os.WriteFile(skillMdPath, []byte(skillContent), 0644); err != nil {
		t.Fatalf("Failed to create SKILL.md: %v", err)
	}

	return filepath.Dir(skillMdPath) // Return the skill directory
}

// createTestAgentInDir creates an agent in the specified base directory.
// The agent will be created at baseDir/agents/<name>.md
func createTestAgentInDir(t *testing.T, baseDir, name, description string) string {
	agentsDir := filepath.Join(baseDir, "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("Failed to create agents directory: %v", err)
	}

	agentPath := filepath.Join(agentsDir, name+".md")
	agentContent := fmt.Sprintf(`---
description: %s
---
# %s
Test agent content.
`, description, name)

	if err := os.WriteFile(agentPath, []byte(agentContent), 0644); err != nil {
		t.Fatalf("Failed to create agent file: %v", err)
	}

	return agentPath
}

// copyFile copies a file from src to dst
func copyFile(t *testing.T, src, dst string) {
	srcFile, err := os.Open(src)
	if err != nil {
		t.Fatalf("Failed to open source file %s: %v", src, err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		t.Fatalf("Failed to create destination file %s: %v", dst, err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		t.Fatalf("Failed to copy file from %s to %s: %v", src, dst, err)
	}
}

// copyDir recursively copies a directory from src to dst
func copyDir(t *testing.T, src, dst string) {
	// Get source directory info
	srcInfo, err := os.Stat(src)
	if err != nil {
		t.Fatalf("Failed to stat source directory %s: %v", src, err)
	}

	// Create destination directory
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		t.Fatalf("Failed to create destination directory %s: %v", dst, err)
	}

	// Read source directory entries
	entries, err := os.ReadDir(src)
	if err != nil {
		t.Fatalf("Failed to read source directory %s: %v", src, err)
	}

	// Copy each entry
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Recursively copy subdirectory
			copyDir(t, srcPath, dstPath)
		} else {
			// Copy file
			copyFile(t, srcPath, dstPath)
		}
	}
}
