package kanban

import "testing"

// TestNormalizeCardText pins the four-step pipeline — NFC, trim,
// whitespace collapse, case-fold — and nothing beyond it. The pipeline's
// length is load-bearing: a fifth step (stemming, stop-word removal,
// punctuation stripping) would raise measured similarity across the board
// and push genuinely distinct cards over the near-duplicate threshold.
func TestNormalizeCardText(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"already normal", "fix the flaky gate", "fix the flaky gate"},
		{"case folded", "Fix The FLAKY Gate", "fix the flaky gate"},
		{"trimmed", "   fix the flaky gate   ", "fix the flaky gate"},
		{"internal runs collapsed", "fix   the\tflaky\n\ngate", "fix the flaky gate"},
		{"combined", "  fix   the FLAKY gate ", "fix the flaky gate"},
		// NFC composition: "e" + U+0301 (combining acute) folds to "é".
		{"nfc composed", "café build", "café build"},
		{"punctuation kept", "fix the gate!", "fix the gate!"},
		{"empty", "   ", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := NormalizeCardText(tc.in); got != tc.want {
				t.Errorf("NormalizeCardText(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

// TestTokenSetJaccard pins the similarity measure, including the fixture the
// acceptance suite depends on: the two auth-middleware texts score 5/6 with
// only 0.033 of margin above the 0.80 threshold.
func TestTokenSetJaccard(t *testing.T) {
	cases := []struct {
		name string
		a, b string
		want float64
	}{
		{"identical", "fix the gate", "fix the gate", 1},
		{"disjoint", "fix the gate", "ship telemetry", 0},
		{"one token dropped from six", "Rework the auth middleware error paths",
			"Rework auth middleware error paths", 5.0 / 6.0},
		{"case and space insensitive", "Fix  The Gate", "fix the gate", 1},
		{"empty left", "", "fix the gate", 0},
		{"both empty", "", "", 0},
		{"repeated tokens collapse to a set", "fix fix the gate", "fix the gate", 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := TokenSetJaccard(tc.a, tc.b)
			if diff := got - tc.want; diff > 1e-9 || diff < -1e-9 {
				t.Errorf("TokenSetJaccard(%q, %q) = %v, want %v", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

// TestClassifyCardText covers the three classifications and the two
// deliberate boundaries: a dropped card is not a comparison subject, and a
// permutation that scores exactly 1.0 without being an exact match falls
// outside `near` (whose range is [0.80, 1.0)) rather than being reclassified.
func TestClassifyCardText(t *testing.T) {
	items := []BacklogItem{
		{ID: "t1", Text: "Rework the auth middleware error paths", State: BacklogStateQueued},
		{ID: "t2", Text: "Add a Windows CI matrix job", State: BacklogStatePicked},
		{ID: "t3", Text: "Prune stale worktrees", State: BacklogStateDropped},
	}

	t.Run("exact against a queued card", func(t *testing.T) {
		got := ClassifyCardText("  rework   THE auth middleware error paths ", items)
		if got.Kind != BacklogMatchExact || got.ID != "t1" {
			t.Errorf("got %+v, want an exact match on t1", got)
		}
	})

	t.Run("exact against a picked card", func(t *testing.T) {
		got := ClassifyCardText("add a windows ci matrix job", items)
		if got.Kind != BacklogMatchExact || got.ID != "t2" {
			t.Errorf("got %+v, want an exact match on t2", got)
		}
	})

	t.Run("near above the threshold", func(t *testing.T) {
		got := ClassifyCardText("Rework auth middleware error paths", items)
		if got.Kind != BacklogMatchNear || got.ID != "t1" {
			t.Fatalf("got %+v, want a near match on t1", got)
		}
		if got.Score < BacklogNearDuplicateThreshold {
			t.Errorf("score = %v, want >= %v", got.Score, BacklogNearDuplicateThreshold)
		}
	})

	t.Run("below the threshold is none", func(t *testing.T) {
		if got := ClassifyCardText("Rework the auth middleware", items); got.Kind != BacklogMatchNone {
			t.Errorf("got %+v, want none (score %v is below the threshold)", got, got.Score)
		}
	})

	t.Run("semantically similar but textually distinct is none", func(t *testing.T) {
		if got := ClassifyCardText("Repair broken login", items); got.Kind != BacklogMatchNone {
			t.Errorf("got %+v, want none — the mechanical layer reads text, not meaning", got)
		}
	})

	t.Run("a dropped card is not a comparison subject", func(t *testing.T) {
		if got := ClassifyCardText("prune stale worktrees", items); got.Kind != BacklogMatchNone {
			t.Errorf("got %+v, want none — t3 is dropped", got)
		}
	})

	t.Run("a token permutation scoring 1.0 is not near", func(t *testing.T) {
		// `near` is defined as [0.80, 1.0): a permutation scores exactly 1.0
		// and is not an exact match, so it lands outside both classes. The
		// range is asserted here so a later widening is a deliberate change
		// rather than a silent one.
		got := ClassifyCardText("paths error middleware auth the Rework", items)
		if got.Kind != BacklogMatchNone {
			t.Errorf("got %+v, want none — 1.0 sits outside the half-open near range", got)
		}
	})

	t.Run("empty queue", func(t *testing.T) {
		if got := ClassifyCardText("anything", nil); got.Kind != BacklogMatchNone {
			t.Errorf("got %+v, want none", got)
		}
	})
}
