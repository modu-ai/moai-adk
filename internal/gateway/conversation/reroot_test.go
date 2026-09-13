package conversation

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// Card t700: the un-gated transcript pointer accessor the launcher-side wedge
// recovery resolves the conversation record with. The resume completion gate
// must not apply — a wedged transcript can be incomplete.
func TestTranscriptPathResolvesRecordPointerWithoutCompletionGate(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	m, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	d, err := m.New(context.Background(), NewRequest{CWD: filepath.Join(root, "proj"), Project: "proj"})
	if err != nil {
		t.Fatal(err)
	}
	// No transcript pointer registered yet.
	if _, err = m.TranscriptPath(d.UUID); err != ErrIncomplete {
		t.Fatalf("empty pointer: %v", err)
	}
	if _, err = m.TranscriptPath("00000000-0000-4000-8000-00000000ffff"); err != ErrMissing {
		t.Fatalf("unknown id: %v", err)
	}
	transcript := filepath.Join(d.ConfigDir, "projects", "proj", d.UUID+".jsonl")
	if err = os.MkdirAll(filepath.Dir(transcript), 0700); err != nil {
		t.Fatal(err)
	}
	// Register the pointer the ordinary way (Complete gates on a valid
	// transcript), then overwrite the file with rows that would not pass the
	// resume gate: the accessor must still resolve the pointer, because it
	// reads the record — it never re-gates on the file's current content.
	if err = os.WriteFile(transcript, []byte("{\"type\":\"user\"}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err = m.Complete(context.Background(), d.UUID, transcript, 1); err == nil {
		t.Fatal("Complete accepted an invalid transcript")
	}
	valid := []byte("{\"sessionId\":\"" + d.UUID + "\",\"cwd\":\"" + filepath.Join(root, "proj") + "\",\"type\":\"assistant\",\"message\":{\"role\":\"assistant\",\"model\":\"m\",\"stop_reason\":\"end_turn\",\"content\":[]}}\n")
	if err = os.WriteFile(transcript, valid, 0600); err != nil {
		t.Fatal(err)
	}
	if err = m.Complete(context.Background(), d.UUID, transcript, 1); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(transcript, []byte("{\"type\":\"user\"}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	got, err := m.TranscriptPath(d.UUID)
	if err != nil {
		t.Fatal(err)
	}
	if got != transcript {
		t.Fatalf("path=%s want %s", got, transcript)
	}
	// A symlink anywhere between the transcript and the private config root
	// is a write-escape path: refuse.
	if err = os.Rename(transcript, transcript+".real"); err != nil {
		t.Fatal(err)
	}
	if err = os.Symlink(transcript+".real", transcript); err != nil {
		t.Fatal(err)
	}
	if _, err = m.TranscriptPath(d.UUID); err != ErrInvalid {
		t.Fatalf("symlinked transcript: %v", err)
	}
}
