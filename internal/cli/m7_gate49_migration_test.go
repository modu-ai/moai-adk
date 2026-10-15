package cli

// m7_gate49_migration_test.go — gate round 49 NEW findings:
//
//   - NEW-1 [P2]: the converter's HOME-ANCHORED workflow references map to
//     the .agents install root (~/.agents/skills/moai/workflows/) — the
//     former unconditional rewrite sent them to ~/.moai/workflows/, a
//     target that exists on neither side, while the project-relative form
//     correctly deploys to .moai/workflows.
//   - NEW-2 [P2]: userCounterpartConfirmed compares CONVERTED against
//     CONVERTED for the codex faces — the installer writes and hashes
//     converted bytes (REQ-CNV-001), so raw-vs-installed never matched a
//     correctly installed .toml and pinned the stale project copy forever.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/internal/userassets"
)

func TestCodexInstallMigrationAndWorkflowRefs(t *testing.T) {
	// NEW-1: the home-anchored workflow anchor maps to the .agents root.
	converted := template.NormalizeCodexRoleForDeploy([]byte("contract: ~/.claude/skills/moai/workflows/design.md\n"))
	if strings.Contains(string(converted), "~/.moai/workflows/") {
		t.Fatalf("the home-anchored workflow reference landed on the nonexistent ~/.moai/workflows (gate 49 NEW-1): %s", converted)
	}
	if !strings.Contains(string(converted), "~/.agents/skills/moai/workflows/design.md") {
		t.Fatalf("the home-anchored workflow reference did not map to the .agents install root: %s", converted)
	}
	// The project-relative form keeps its .moai deploy target.
	proj := template.NormalizeCodexRoleForDeploy([]byte("see .claude/skills/moai/workflows/design.md\n"))
	if !strings.Contains(string(proj), ".moai/workflows/design.md") {
		t.Fatalf("the project-relative workflow reference lost the .moai deploy target: %s", proj)
	}

	// NEW-2: a correctly installed codex .toml counterpart confirms TRUE.
	tomlSource := []byte("name = \"manager-x\"\ncontract = \"~/.claude/skills/moai/workflows/design.md\"\n")
	fhome := t.TempDir()
	src := fstest.MapFS{
		".claude/agents/moai/manager-x.md":  &fstest.MapFile{Data: []byte("---\nname: manager-x\n---\nbody\n")},
		".codex/agents/moai/manager-x.toml": &fstest.MapFile{Data: tomlSource},
	}
	cat := &template.Catalog{
		Version: "test",
		Catalog: template.CatalogSections{
			Core: template.TierSection{
				Agents: []template.Entry{
					{Name: "manager-x", Tier: template.TierCore, Path: ".claude/agents/moai/manager-x.md", Version: "1.0.0"},
				},
			},
		},
	}
	inst := &userassets.Installer{Home: fhome, Catalog: cat, Source: src, MoaiVersion: "test"}
	if _, err := inst.Install(nil); err != nil {
		t.Fatalf("install: %v", err)
	}
	manifest, err := userassets.Load(userassets.ManifestPath(fhome))
	if err != nil {
		t.Fatal(err)
	}

	// The project copy carries the CONVERTED bytes — exactly what the
	// installer wrote and hashed.
	projectDir := t.TempDir()
	tomlDir := filepath.Join(projectDir, ".codex", "agents", "moai")
	if err := os.MkdirAll(tomlDir, 0o755); err != nil {
		t.Fatal(err)
	}
	installedBytes := template.NormalizeCodexRoleForDeploy(tomlSource)
	if err := os.WriteFile(filepath.Join(tomlDir, "manager-x.toml"), installedBytes, 0o644); err != nil {
		t.Fatal(err)
	}
	if !userCounterpartConfirmed(manifest, cat, src, fhome, ".codex/agents/moai/manager-x.toml") {
		t.Fatal("a correctly installed (converted-bytes-matching) codex counterpart confirmed FALSE — the stale project copy would be pinned forever (gate 49 NEW-2)")
	}
}
