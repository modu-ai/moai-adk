package review

import (
	"strings"
	"testing"
	"time"
)

var entryAt = time.Date(2026, 10, 29, 6, 0, 0, 0, time.UTC)

// The context ledger's audit section carries: result, actual model and
// effort, target SHA, verdict artifact hash, and time; it goes stale when
// the HEAD moves (S46 HANDOFF08 standard).
func TestAuditEntry_FieldsAndStale(t *testing.T) {
	rec := completed(t, VerdictPass, headA, "s-audit")
	hash := HashArtifact([]byte("verdict body"))
	e, err := NewAuditEntry(rec, "claude-opus-5-5", "high", hash, entryAt)
	if err != nil {
		t.Fatal(err)
	}
	if e.Result != VerdictPass || e.Model != "claude-opus-5-5" || e.Effort != "high" ||
		e.Head != headA || e.ArtifactHash != hash || !e.At.Equal(entryAt) {
		t.Fatalf("entry = %+v", e)
	}
	if e.Status(headA) != EntryCurrent || e.Status(headB) != EntryStale {
		t.Fatalf("status at same / moved HEAD = %q / %q", e.Status(headA), e.Status(headB))
	}
}

func TestAuditEntry_RecordsNonPassResultsToo(t *testing.T) {
	for _, v := range []Verdict{VerdictFail, VerdictPassWithDebt, VerdictSkipped} {
		rec := completed(t, v, headA, "s-audit")
		e, err := NewAuditEntry(rec, "m", "e", HashArtifact([]byte("x")), entryAt)
		if err != nil || e.Result != v {
			t.Errorf("verdict %s: entry=%+v err=%v", v, e, err)
		}
	}
}

// Model and effort are measured values from the caller; none is assumed.
func TestAuditEntry_RefusesUnobservedFields(t *testing.T) {
	rec := completed(t, VerdictPass, headA, "s-audit")
	ok := HashArtifact([]byte("x"))
	incomplete := rec
	incomplete.delivery = nil
	cases := []struct {
		name          string
		rec           Record
		model, effort string
		hash          string
		at            time.Time
	}{
		{"no model", rec, "", "high", ok, entryAt},
		{"no effort", rec, "m", "", ok, entryAt},
		{"short hash", rec, "m", "high", "abc", entryAt},
		{"non-hex hash", rec, "m", "high", strings.Repeat("z", 64), entryAt},
		{"no time", rec, "m", "high", ok, time.Time{}},
		{"bad head", func() Record { r := rec; r.HeadSHA = "abc"; return r }(), "m", "high", ok, entryAt},
		{"verdict never completed", incomplete, "m", "high", ok, entryAt},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := NewAuditEntry(tc.rec, tc.model, tc.effort, tc.hash, tc.at); err == nil {
				t.Fatal("entry created from unobserved input")
			}
		})
	}
}

func TestHashArtifact_Stable(t *testing.T) {
	a, b := HashArtifact([]byte("v")), HashArtifact([]byte("v"))
	if a != b || len(a) != 64 || a == HashArtifact([]byte("w")) {
		t.Fatalf("hash not stable or not sha256 hex: %q", a)
	}
}
