package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// gtd_answer_test.go — `moai gtd answer` unit coverage (card t863):
// the file-based operator response channel for gate-blocked cards.

func TestGTDAnswerWritesAppendOnlyRecord(t *testing.T) {
	root := t.TempDir()
	path, err := writeGTDAnswer(root, "t128", "B안으로 간다. reason: 기존 스키마 호환")
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(root, ".moai", "gtd", "answers", "t128.md"); path != want {
		t.Fatalf("answer path = %q, want %q", path, want)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "B안으로 간다") {
		t.Fatalf("answer text missing:\n%s", data)
	}
	if !strings.Contains(string(data), "## 2") { // RFC3339 timestamp heading starts with the year
		t.Fatalf("timestamp heading missing:\n%s", data)
	}

	// A second answer appends; the first record survives.
	if _, err := writeGTDAnswer(root, "t128", "second decision"); err != nil {
		t.Fatal(err)
	}
	data, _ = os.ReadFile(path)
	if !strings.Contains(string(data), "B안으로 간다") || !strings.Contains(string(data), "second decision") {
		t.Fatalf("append-only violated:\n%s", data)
	}
}

func TestGTDAnswerRejectsInvalidID(t *testing.T) {
	root := t.TempDir()
	for _, bad := range []string{"128", "t", "todo-5", "t12x", "T12"} {
		if _, err := writeGTDAnswer(root, bad, "text"); err == nil {
			t.Errorf("id %q accepted, want rejection", bad)
		}
	}
}

func TestGTDAnswerCardUntouched(t *testing.T) {
	// Answering is evidence, not a transition: the queue file must not be
	// created or touched by an answer (mirrors todo landed semantics).
	root := t.TempDir()
	if _, err := writeGTDAnswer(root, "t9", "decision"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, ".moai", "db")); err == nil {
		t.Fatal("gtd answer touched the queue store")
	}
}
