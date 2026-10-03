package cli

// deploy_mode_record_test.go — AC-009 (SPEC-INIT-SHRINK-001 REQ-009, OD-5
// settled (a)): the deploy-mode record round-trips through the same llm.yaml
// seam as llm.harness, and the recorded value survives an update cycle
// modeled on update_template_sync.go's Backup → Clean (.moai/config wipe) →
// redeploy → restore-reassert step order (P-06/P-07).

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/update/backup"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/template"
)

// deployTestProject seeds a project root with the template's llm.yaml render.
func deployTestProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	embedded, err := template.EmbeddedTemplates()
	if err != nil {
		t.Fatalf("load embedded templates: %v", err)
	}
	data, err := fs.ReadFile(embedded, ".moai/config/sections/llm.yaml")
	if err != nil {
		t.Fatalf("embedded llm.yaml: %v", err)
	}
	sections := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sections, "llm.yaml"), data, 0o644); err != nil {
		t.Fatalf("write llm.yaml: %v", err)
	}
	return root
}

// TestDeployModeRecordRoundTrip is AC-009: init writes the record on every
// run shape (plugin and local), a re-init rewrites it, and the recorded
// value survives a full update cycle's Clean/redeploy/restore process.
func TestDeployModeRecordRoundTrip(t *testing.T) {
	for _, mode := range []string{"plugin", "local"} {
		t.Run("record-"+mode, func(t *testing.T) {
			root := deployTestProject(t)

			// Init-side write (the ApplyDeployMode call in runInit's tail).
			if err := template.ApplyDeployMode(root, mode); err != nil {
				t.Fatalf("ApplyDeployMode: %v", err)
			}
			if got := config.ReadDeployMode(root); got != mode {
				t.Fatalf("after init write, ReadDeployMode = %q, want %q", got, mode)
			}

			// Re-init rewrites the record (REQ-009: on every init run).
			if err := template.ApplyDeployMode(root, "local"); err != nil {
				t.Fatalf("re-init write: %v", err)
			}
			if got := config.ReadDeployMode(root); got != "local" {
				t.Fatalf("after re-init, ReadDeployMode = %q, want local", got)
			}
			if err := template.ApplyDeployMode(root, mode); err != nil {
				t.Fatalf("restore record: %v", err)
			}
		})
	}

	t.Run("survives-update-cycle", func(t *testing.T) {
		root := deployTestProject(t)
		if err := template.ApplyDeployMode(root, "plugin"); err != nil {
			t.Fatalf("ApplyDeployMode: %v", err)
		}

		// Step 1 — Backup (the update flow's Backup step).
		configBackupPath, err := backup.BackupMoaiConfig(root)
		if err != nil {
			t.Fatalf("BackupMoaiConfig: %v", err)
		}
		if configBackupPath == "" {
			t.Fatal("BackupMoaiConfig returned an empty path")
		}

		// Step 2 — Clean Managed Paths removes .moai/config wholesale (P-07).
		if err := os.RemoveAll(filepath.Join(root, ".moai", "config")); err != nil {
			t.Fatalf("clean .moai/config: %v", err)
		}
		if got := config.ReadDeployMode(root); got != "" {
			t.Fatalf("record readable after the Clean wipe (%q) — the wipe must have removed it", got)
		}

		// Step 3 — Deploy Templates rewrites llm.yaml from the template
		// render (which carries no deployment_mode key).
		embedded, err := template.EmbeddedTemplates()
		if err != nil {
			t.Fatalf("load embedded templates: %v", err)
		}
		data, err := fs.ReadFile(embedded, ".moai/config/sections/llm.yaml")
		if err != nil {
			t.Fatalf("embedded llm.yaml: %v", err)
		}
		sections := filepath.Join(root, ".moai", "config", "sections")
		if err := os.MkdirAll(sections, 0o755); err != nil {
			t.Fatalf("mkdir sections: %v", err)
		}
		if err := os.WriteFile(filepath.Join(sections, "llm.yaml"), data, 0o644); err != nil {
			t.Fatalf("redeploy llm.yaml: %v", err)
		}

		// Step 4 — Restore Settings re-asserts the PRE-UPDATE record from the
		// backup (the same shape as update_template_sync.go's harness
		// re-assert; M3 wires this beside it for deployment_mode).
		if recorded := config.ReadDeployModeFrom(filepath.Join(configBackupPath, "sections")); recorded != "" {
			if err := template.ApplyDeployMode(root, recorded); err != nil {
				t.Fatalf("re-assert from backup: %v", err)
			}
		}

		// Step 5 — the re-read returns the recorded value byte-identical
		// (OD-5 settled condition).
		if got := config.ReadDeployMode(root); got != "plugin" {
			t.Fatalf("after the update cycle, ReadDeployMode = %q, want plugin", got)
		}
	})
}
