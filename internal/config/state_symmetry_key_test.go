package config

// state_symmetry_key_test.go — AC-UBR-21 (SPEC-UPDATE-BACKUP-RETENTION-001,
// REQ-UBR-020): the shipped state.yaml and StateConfig agree on
// state.backup_past_versions_keep, in both directions.
//
// StateConfig cannot join symmetryCases (audit_struct_yaml_symmetry_test.go):
// checkSymmetry is a whole-section bijection over top-level keys, and the shipped
// state: block carries profile_max_bytes, which has no struct field, while the
// struct carries retention_days, which the block does not ship. Joining it would
// fail on keys this SPEC does not own. This test applies the same two-direction
// check to the one key the SPEC ships, and adds the struct-backed read: the value
// the shipped file carries must decode, through the section wrapper the loader
// uses, to DefaultBackupPastVersionsKeep.

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

const backupPastKeepKey = "backup_past_versions_keep"

func TestStructYAMLSymmetry_StateBackupPastVersionsKeep(t *testing.T) {
	t.Parallel()
	repoRoot := findRepoRootFromCaller(t)
	path := filepath.Join(repoRoot, "internal", "template", "templates",
		".moai", "config", "sections", "state.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read shipped state.yaml: %v", err)
	}

	var raw map[string]any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		t.Fatalf("parse shipped state.yaml: %v", err)
	}
	block, ok := raw["state"].(map[string]any)
	if !ok {
		t.Fatalf("shipped state.yaml has no state: mapping (got %T)", raw["state"])
	}
	_, inYAML := block[backupPastKeepKey]

	inStruct := false
	st := reflect.TypeOf(StateConfig{})
	for i := 0; i < st.NumField(); i++ {
		if name, _, _ := strings.Cut(st.Field(i).Tag.Get("yaml"), ","); name == backupPastKeepKey {
			inStruct = true
		}
	}

	if !inYAML {
		t.Errorf("CONFIG_STRUCT_YAML_MISMATCH: field=StateConfig.%s, side=go-only (key exists in struct but not in shipped state.yaml)", backupPastKeepKey)
	}
	if !inStruct {
		t.Errorf("CONFIG_STRUCT_YAML_MISMATCH: field=StateConfig.%s, side=yaml-only (key exists in shipped state.yaml but not in struct)", backupPastKeepKey)
	}

	// Struct-backed read: the shipped bytes decode, through the same section
	// wrapper the loader uses, to the default the pruner falls back to.
	var wrapper struct {
		State StateConfig `yaml:"state"`
	}
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		t.Fatalf("decode shipped state.yaml into the section wrapper: %v", err)
	}
	if wrapper.State.BackupPastVersionsKeep != DefaultBackupPastVersionsKeep {
		t.Errorf("shipped state.%s decodes to %d, want DefaultBackupPastVersionsKeep (%d)",
			backupPastKeepKey, wrapper.State.BackupPastVersionsKeep, DefaultBackupPastVersionsKeep)
	}
}
