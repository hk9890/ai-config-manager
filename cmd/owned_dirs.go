package cmd

import (
	"path/filepath"
	"sort"

	"github.com/dynatrace-oss/ai-config-manager/v3/pkg/resource"
	"github.com/dynatrace-oss/ai-config-manager/v3/pkg/tools"
)

// OwnedResourceDir describes a project directory owned by aimgr for a specific
// tool and resource type. This helper is shared by repair/clean/verify flows.
type OwnedResourceDir struct {
	Tool         tools.Tool
	ResourceType resource.ResourceType
	Path         string
}

func detectOwnedResourceDirs(projectPath string) ([]OwnedResourceDir, error) {
	detectedTools, err := tools.DetectExistingTools(projectPath)
	if err != nil {
		return nil, err
	}

	owned := make([]OwnedResourceDir, 0, len(detectedTools)*3)
	for _, tool := range detectedTools {
		info := tools.GetToolInfo(tool)
		if info.SupportsCommands {
			owned = append(owned, OwnedResourceDir{
				Tool:         tool,
				ResourceType: resource.Command,
				Path:         filepath.Join(projectPath, info.CommandsDir),
			})
		}
		if info.SupportsSkills {
			owned = append(owned, OwnedResourceDir{
				Tool:         tool,
				ResourceType: resource.Skill,
				Path:         filepath.Join(projectPath, info.SkillsDir),
			})
		}
		if info.SupportsAgents {
			owned = append(owned, OwnedResourceDir{
				Tool:         tool,
				ResourceType: resource.Agent,
				Path:         filepath.Join(projectPath, info.AgentsDir),
			})
		}
	}

	return owned, nil
}

func toolsFromOwnedDirs(owned []OwnedResourceDir) []tools.Tool {
	set := make(map[tools.Tool]struct{})
	for _, dir := range owned {
		set[dir.Tool] = struct{}{}
	}
	out := make([]tools.Tool, 0, len(set))
	for tool := range set {
		out = append(out, tool)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].String() < out[j].String() })
	return out
}
