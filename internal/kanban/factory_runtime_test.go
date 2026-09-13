package kanban

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRecordFactoryRunStartRecordsMetadataWithoutClaimingSpecProvenance(t *testing.T) {
	root := t.TempDir()
	t.Setenv("MOAI_HOME", t.TempDir())
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	run("init", "-q")
	run("config", "user.email", "t@example.com")
	run("config", "user.name", "t")
	specPath := filepath.Join(root, ".moai", "specs", "SPEC-X", "spec.md")
	if err := os.MkdirAll(filepath.Dir(specPath), 0o700); err != nil {
		t.Fatal(err)
	}
	body := []byte("# spec\n")
	if err := os.WriteFile(specPath, body, 0o600); err != nil {
		t.Fatal(err)
	}
	run("add", ".")
	run("commit", "-qm", "fixture")
	if err := RecordFactoryRunStart(root, "run-spec", BackendClaude, "SPEC-X"); err != nil {
		t.Fatal(err)
	}
	record, err := NewBacklogStore(BacklogPathForRoot(root)).LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	if len(record.Runtime.Runs) != 1 {
		t.Fatalf("run count=%d", len(record.Runtime.Runs))
	}
	raw := record.Runtime.Runs[0].ManifestJSON
	var got factoryProvenance
	if err := json.Unmarshal([]byte(raw), &got); err != nil {
		t.Fatal(err)
	}
	if got.SpecID != "" || got.SpecPath != "" || got.SpecSHA256 != "" || got.GitCommit == "" || got.CapturedAt == "" {
		t.Fatalf("manifest=%+v", got)
	}
	if strings.TrimSpace(got.GitCommit) == "" {
		t.Fatal("git commit missing")
	}
}
