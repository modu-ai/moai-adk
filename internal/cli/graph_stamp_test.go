package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/mx"
)

// stampFixture creates a git repo with described sources (reuses the
// graph_check_test git helper in this package).
func stampFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	checkFixtureGit(t, root, "init", "-q")
	checkFixtureGit(t, root, "config", "user.email", "f@e.com")
	checkFixtureGit(t, root, "config", "user.name", "F")
	if err := os.MkdirAll(filepath.Join(root, "internal"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "internal", "a.go"), []byte("package a\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	checkFixtureGit(t, root, "add", "-A")
	checkFixtureGit(t, root, "commit", "-q", "-m", "base")
	return root
}

// Command-level stamp coverage (CR round-2 3855001919): valid output, and
// REPLACEMENT of an existing stamp.
func TestGraphStampCmd_ValidAndReplacement(t *testing.T) {
	root := stampFixture(t)

	run := func() string {
		cmd := newGraphCmd()
		cmd.SetArgs([]string{"stamp", "codemaps", "--root", root})
		var out, errOut bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&errOut)
		if err := cmd.Execute(); err != nil {
			t.Fatalf("stamp codemaps: %v (stderr %s)", err, errOut.String())
		}
		return out.String()
	}

	first := run()
	target := filepath.Join(root, ".moai", "project", "codemaps", "provenance.json")
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("provenance.json not written: %v", err)
	}
	var pv mx.Provenance
	if err := json.Unmarshal(data, &pv); err != nil {
		t.Fatalf("provenance.json not valid JSON: %v", err)
	}
	if pv.GeneratedBy != "codemaps-gen" || pv.CommitSHA == "" {
		t.Errorf("provenance missing core fields: %+v", pv)
	}
	if !strings.Contains(first, "OK: stamped") {
		t.Errorf("command output missing OK line: %s", first)
	}

	// REPLACE an existing stamp: second run overwrites with a fresh block.
	if err := os.WriteFile(target, []byte("{\"broken\": true}"), 0o644); err != nil {
		t.Fatal(err)
	}
	run()
	data2, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	var pv2 mx.Provenance
	if err := json.Unmarshal(data2, &pv2); err != nil {
		t.Fatalf("replacement stamp not valid JSON: %v", err)
	}
	if pv2.GeneratedBy != "codemaps-gen" {
		t.Errorf("replacement lost the generator field: %+v", pv2)
	}

	// After a successful rename the temp file must not linger.
	if _, err := os.Stat(target + ".tmp"); !os.IsNotExist(err) {
		t.Errorf("temp provenance file lingers after success")
	}
}
