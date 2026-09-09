package kanban

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/homestate"
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
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	var raw string
	if err := db.DB.QueryRow(`SELECT manifest_json FROM runs WHERE run_id='run-spec'`).Scan(&raw); err != nil {
		t.Fatal(err)
	}
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
