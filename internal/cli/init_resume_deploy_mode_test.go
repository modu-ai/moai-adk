package cli

// init_resume_deploy_mode_test.go — card t1619, codex review round 3 (P2): the resume checkpoint
// reads the deployment_mode record as a parsed key under llm, the place ApplyDeployMode writes it
// and the config reader reads it. A text search was wrong both ways: a comment that names the key
// counted as recorded, and a spaced colon did not.

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/userassets"
)

func TestInitResumeCheckpointReadsDeployModeAsKey(t *testing.T) {
	cases := []struct {
		name          string
		llm           string
		wantResumable bool // true: the post-steps never finished, so the rerun resumes
	}{
		{name: "recorded key", llm: "llm:\n  deployment_mode: local\n", wantResumable: false},
		{name: "recorded key with spaced colon", llm: "llm:\n  deployment_mode : local\n", wantResumable: false},
		{name: "comment naming the key is not a record", llm: "# deployment_mode: local\nllm:\n  harness: claude\n", wantResumable: true},
		{name: "no key", llm: "llm:\n  harness: claude\n", wantResumable: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			projectRoot := t.TempDir()
			sections := filepath.Join(projectRoot, ".moai", "config", "sections")
			if err := os.MkdirAll(sections, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(sections, "llm.yaml"), []byte(tc.llm), 0o644); err != nil {
				t.Fatal(err)
			}
			// An interruption checkpoint: a pending-install journal in the user home.
			journal := &userassets.PendingJournal{SchemaVersion: userassets.SchemaVersion}
			if err := userassets.WriteJournal(userassets.JournalPath(home), journal); err != nil {
				t.Fatalf("stage the pending-install journal: %v", err)
			}
			if got := initResumeCheckpoint(home, projectRoot); got != tc.wantResumable {
				t.Fatalf("checkpoint = %v, want %v for llm.yaml %q", got, tc.wantResumable, tc.llm)
			}
		})
	}
}
