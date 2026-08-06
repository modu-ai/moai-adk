package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNavigatorSync_EmitsNavGraphWhenInputsPresent verifies the CLI surface
// for the join step end-to-end: a populated project yields nav-graph.json
// with the three required top-level keys.
func TestNavigatorSync_EmitsNavGraphWhenInputsPresent(t *testing.T) {
	root := t.TempDir()
	navSyncWriteFile(t, filepath.Join(root, ".moai", "project", "navigator", "capability-map.md"),
		"| spec-id | title | implementation-path |\n"+
			"|---------|-------|---------------------|\n"+
			"| SPEC-X-001 | X | src/x |\n")
	navSyncWriteFile(t, filepath.Join(root, ".moai", "specs", "SPEC-X-001", "spec.md"),
		"---\nid: SPEC-X-001\nmodule: src/x\n---\n# SPEC-X-001\n")
	navSyncWriteFile(t, filepath.Join(root, "src", "x", "a.go"),
		"package x\n// @NAV:DEC-AUTH: OAuth2\n// @MX:NOTE: x\n// @MX:SPEC: SPEC-X-001\nfunc A() {}\n")

	if err := runNavigatorSync(root, "", "", "", ""); err != nil {
		t.Fatalf("runNavigatorSync error: %v", err)
	}
	out := filepath.Join(root, ".moai", "project", "navigator", "nav-graph.json")
	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read nav-graph.json: %v", err)
	}
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(b, &doc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, key := range []string{"provenance", "nodes", "edges"} {
		if _, ok := doc[key]; !ok {
			t.Errorf("nav-graph.json missing key %q", key)
		}
	}
}

// TestNavigatorSync_AbsenceIsGraceful verifies REQ-NS-011 fail-open at the
// CLI surface: when capability-map.md is absent, exit is nil and no
// nav-graph.json is written.
func TestNavigatorSync_AbsenceIsGraceful(t *testing.T) {
	root := t.TempDir()
	if err := runNavigatorSync(root, "", "", "", ""); err != nil {
		t.Fatalf("runNavigatorSync error: %v", err)
	}
	out := filepath.Join(root, ".moai", "project", "navigator", "nav-graph.json")
	if _, err := os.Stat(out); err == nil {
		t.Errorf("nav-graph.json written despite capability-map absent")
	}
	logBytes, err := os.ReadFile(filepath.Join(root, ".moai", "logs", "navigator-sync.log"))
	if err != nil {
		t.Fatalf("log missing: %v", err)
	}
	if !strings.Contains(string(logBytes), "capability-map absent") {
		t.Errorf("log missing capability-map-absent line:\n%s", logBytes)
	}
}

func navSyncWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
