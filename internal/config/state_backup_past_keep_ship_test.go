package config

// state_backup_past_keep_ship_test.go — AC-UBR-10 (SPEC-UPDATE-BACKUP-RETENTION-001,
// REQ-UBR-007): the backup per-file budget key ships once. The key is present in
// the shipped state.yaml template with the default value, and the default is
// declared in exactly one place in non-test Go code.

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// backupPastKeepDefaultDecl matches a declaration that gives a past-versions
// budget a literal integer value, in a const block entry or a var/const line.
var backupPastKeepDefaultDecl = regexp.MustCompile(`(?m)^\s*(?:(?:const|var)\s+)?\w*PastVersionsKeep\w*\s*(?:int\s*)?=\s*[0-9]+\b`)

func TestAC_UBR_10_BudgetKeyShipsOnce(t *testing.T) {
	repoRoot := findRepoRootFromCaller(t)

	t.Run("template_ships_key_with_default", func(t *testing.T) {
		path := filepath.Join(repoRoot, "internal", "template", "templates", ".moai", "config", "sections", "state.yaml")
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read shipped state.yaml: %v", err)
		}
		var doc struct {
			State map[string]any `yaml:"state"`
		}
		if err := yaml.Unmarshal(data, &doc); err != nil {
			t.Fatalf("parse shipped state.yaml: %v", err)
		}
		got, ok := doc.State["backup_past_versions_keep"]
		if !ok {
			t.Fatal("AC-UBR-10: the shipped state.yaml has no state.backup_past_versions_keep key")
		}
		if got != 3 {
			t.Errorf("AC-UBR-10: shipped state.backup_past_versions_keep = %v, want 3", got)
		}
	})

	t.Run("default_declared_in_one_place", func(t *testing.T) {
		var decls []string
		err := filepath.WalkDir(filepath.Join(repoRoot, "internal"), func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || !strings.HasSuffix(p, ".go") || strings.HasSuffix(p, "_test.go") {
				return nil
			}
			b, rerr := os.ReadFile(p)
			if rerr != nil {
				return rerr
			}
			if backupPastKeepDefaultDecl.Match(b) {
				rel, _ := filepath.Rel(repoRoot, p)
				decls = append(decls, filepath.ToSlash(rel))
			}
			return nil
		})
		if err != nil {
			t.Fatalf("walk internal: %v", err)
		}
		if len(decls) != 1 || decls[0] != "internal/config/defaults.go" {
			t.Errorf("AC-UBR-10: the past-versions budget default is declared in %v, want exactly [internal/config/defaults.go]", decls)
		}
	})
}
