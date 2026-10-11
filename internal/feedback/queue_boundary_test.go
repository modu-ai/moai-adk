package feedback

// The stored-shape cap test (review gate finding, P2): the save-side cap
// was measured with a COMPACT json.Marshal, but the store writes
// MarshalIndent plus a trailing newline — a report crafted at the
// boundary was accepted and created a store the read cap refuses, locking
// every resend, remove, and add behind the load. The cap is measured in
// the ACTUAL stored shape; this test reproduces the exact boundary: the
// crafted report's compact shape fits the remaining room, its indented
// shape does not.

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

func TestEnqueueCapIsMeasuredInTheStoredShape(t *testing.T) {
	s := NewQueueStore(filepath.Join(t.TempDir(), "queue.json"))

	// A probe item measures the compact-vs-indented size delta of one
	// item, and the fixed (non-body) part of an item's compact shape.
	probeCompact, perr := json.Marshal(QueueItem{ID: "f0", Title: "t", Body: "", QueuedAt: "t", Fingerprint: "fp", Kind: "panic"})
	if perr != nil {
		t.Fatalf("probe: %v", perr)
	}
	probeIndented, perr := json.MarshalIndent([]QueueItem{{ID: "f0", Title: "t", Body: "", QueuedAt: "t", Fingerprint: "fp", Kind: "panic"}}, "", "  ")
	if perr != nil {
		t.Fatalf("probe: %v", perr)
	}
	fixedCompact := len(probeCompact)
	indentedDelta := len(probeIndented) - len(probeCompact)

	// Fill with 62 KB reports until the stored shape is within one report
	// of the cap.
	for {
		if _, lerr := s.Load(); lerr != nil {
			t.Fatalf("the store went unreadable while filling: %v", lerr)
		}
		rec, lerr := s.Load()
		if lerr != nil {
			t.Fatalf("load: %v", lerr)
		}
		stored, merr := json.MarshalIndent(rec, "", "  ")
		if merr != nil {
			t.Fatalf("marshal: %v", merr)
		}
		if len(stored)+1 > config.DefaultFeedbackQueueMaxBytes-80*1024 {
			break
		}
		if _, err := s.EnqueueMasked(Result{Verdict: VerdictOK, Title: "filler", Body: strings.Repeat("b", 62*1024)}); err != nil {
			t.Fatalf("filler enqueue: %v", err)
		}
	}

	// Craft the boundary report. storedCompact and storedIndent differ by
	// (filler count) x indentedDelta, so there is an L where the compact
	// total fits the cap and the indented total crosses it.
	rec, lerr := s.Load()
	if lerr != nil {
		t.Fatalf("load: %v", lerr)
	}
	storedCompact, merr := json.Marshal(rec)
	if merr != nil {
		t.Fatalf("marshal: %v", merr)
	}
	storedIndented, merr := json.MarshalIndent(rec, "", "  ")
	if merr != nil {
		t.Fatalf("marshal: %v", merr)
	}
	room := config.DefaultFeedbackQueueMaxBytes - len(storedCompact) - fixedCompact - 4
	if room <= 0 || len(storedIndented)+1+indentedDelta+fixedCompact+room <= config.DefaultFeedbackQueueMaxBytes {
		t.Fatalf("boundary arithmetic did not close (room=%d, stored=%d)", room, len(storedIndented))
	}

	_, ferr := s.EnqueueMasked(Result{Verdict: VerdictOK, Title: "boundary", Body: strings.Repeat("c", room)})
	if ferr == nil {
		t.Fatal("the boundary report was accepted — the store it creates cannot be loaded back (the cap was measured in the compact shape)")
	}
	if !errors.Is(ferr, ErrQueueFull) {
		t.Fatalf("the boundary enqueue failed with %v, want ErrQueueFull", ferr)
	}
	if _, lerr := s.Load(); lerr != nil {
		t.Fatalf("the store became unreadable anyway: %v", lerr)
	}
}
