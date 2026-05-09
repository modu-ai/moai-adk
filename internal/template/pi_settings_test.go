package template

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"strings"
	"testing"
)

func TestPiSettingsTemplateDefaultPackages(t *testing.T) {
	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	data, err := fs.ReadFile(fsys, ".pi/settings.json")
	if err != nil {
		t.Fatalf("ReadFile(.pi/settings.json) error: %v", err)
	}

	var settings struct {
		Packages   []any `json:"packages"`
		MoAICompat struct {
			DefaultPackages []string `json:"defaultPackages"`
		} `json:"moaiCompat"`
	}
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf(".pi/settings.json template is invalid JSON: %v", err)
	}

	configured := make(map[string]bool, len(settings.Packages))
	for _, spec := range settings.Packages {
		configured[normalizePiPackageSpecForTest(spec)] = true
	}

	if len(settings.MoAICompat.DefaultPackages) == 0 {
		t.Fatal("moaiCompat.defaultPackages must not be empty")
	}
	for _, spec := range settings.MoAICompat.DefaultPackages {
		name := normalizePiPackageSpecForTest(spec)
		if !configured[name] {
			t.Fatalf("default package %q is not active in packages", name)
		}
	}

	contextMode := findPiPackageSpecForTest(settings.Packages, "context-mode")
	contextModeObject, ok := contextMode.(map[string]any)
	if !ok {
		t.Fatal("context-mode package must use object filter form")
	}
	if extensions, ok := contextModeObject["extensions"].([]any); !ok || len(extensions) != 0 {
		t.Fatalf("context-mode extensions must be disabled, got %#v", contextModeObject["extensions"])
	}
	if skills, ok := contextModeObject["skills"].([]any); !ok || len(skills) != 1 || skills[0] != "./skills" {
		t.Fatalf("context-mode skills filter must preserve ./skills, got %#v", contextModeObject["skills"])
	}

	for _, required := range []string{
		"@juicesharp/rpiv-ask-user-question",
		"@tmustier/pi-agent-teams",
		"@zenobius/pi-worktrees",
		"context-mode",
		"pi-docparser",
		"pi-markdown-preview",
		"pi-mcp-adapter",
		"pi-subagents",
		"pi-web-access",
		"pi-yaml-hooks",
	} {
		if !configured[required] {
			t.Fatalf("required Pi package %q missing from template", required)
		}
	}
}

func findPiPackageSpecForTest(specs []any, packageName string) any {
	for _, spec := range specs {
		if normalizePiPackageSpecForTest(spec) == packageName {
			return spec
		}
	}
	return nil
}

func normalizePiPackageSpecForTest(spec any) string {
	value := fmt.Sprint(spec)
	if object, ok := spec.(map[string]any); ok {
		value = fmt.Sprint(object["source"])
	}
	value = strings.TrimPrefix(strings.TrimPrefix(value, "npm:"), "git:")
	if before, _, ok := strings.Cut(value, "#"); ok {
		value = before
	}
	if before, _, ok := strings.Cut(value, "?"); ok {
		value = before
	}
	if strings.HasPrefix(value, "@") {
		parts := strings.Split(value, "@")
		if len(parts) > 2 {
			return "@" + parts[1]
		}
		return value
	}
	if before, _, ok := strings.Cut(value, "@"); ok {
		return before
	}
	return value
}
