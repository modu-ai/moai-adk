//go:build unix

package cli

// init_resume_llm_fifo_unix_test.go — card t1619, codex review round 3 (P2): a FIFO at llm.yaml
// must not stall the resume checkpoint. The read is non-blocking and bound to a regular file, and
// a file that is not one is not a readable deployment record, so the checkpoint reads it as not
// resumable instead of waiting for a writer.

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/userassets"
)

func TestInitResumeCheckpointNonRegularLLMConfigDoesNotBlock(t *testing.T) {
	home := t.TempDir()
	projectRoot := t.TempDir()
	sections := filepath.Join(projectRoot, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := syscall.Mkfifo(filepath.Join(sections, "llm.yaml"), 0o644); err != nil {
		t.Fatalf("mkfifo: %v", err)
	}
	journal := &userassets.PendingJournal{SchemaVersion: userassets.SchemaVersion}
	if err := userassets.WriteJournal(userassets.JournalPath(home), journal); err != nil {
		t.Fatalf("stage the pending-install journal: %v", err)
	}

	done := make(chan bool, 1)
	go func() { done <- initResumeCheckpoint(home, projectRoot) }()
	select {
	case got := <-done:
		if got {
			t.Fatal("a FIFO llm.yaml is not a readable deployment record: the checkpoint must not read it as an unfinished project")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("initResumeCheckpoint blocked on a FIFO llm.yaml: the checkpoint's read must be non-blocking")
	}
}
