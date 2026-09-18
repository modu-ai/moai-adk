package spec

import (
	"reflect"
	"testing"
)

// Card t913 — the prose-absorption boundary at the LOCATOR layer.
//
// The pair below is the discriminator. The hazard case must lose the tail the
// prose disclaims; every form whose tail is properly closed must keep it. A
// narrowing that only satisfies the first half is an over-correction — it drops
// legitimate shorthand — and a narrowing that only satisfies the second half is
// the defect this card removes.
//
// MUTATION: delete the truncation branch in siblingMapsREQIDs — the hazard case
// rises from one id to two and fails, while every closed form stays green. That
// asymmetry is what makes this pair a control rather than a restatement.

// TestSiblingMapsProseAfterTailIsNotAbsorbed is the hazard case.
func TestSiblingMapsProseAfterTailIsNotAbsorbed(t *testing.T) {
	text := "- AC-FIXH-001 (maps REQ-FIXH-001, 002 is explicitly NOT mapped by this AC): body.\n"

	got := sortedMapsIDs(siblingMapsREQIDs(text))
	want := []string{"REQ-FIXH-001"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("siblingMapsREQIDs = %v, want %v (the disclaimed tail must not be absorbed)", got, want)
	}
}

// TestSiblingMapsClosedTailStillExpands enumerates every closing token the
// boundary accepts, so a later narrowing of the token set cannot pass silently.
// The `|` row is the one the prose declaration's original candidate list omitted:
// 22 live `maps` sections sit inside a table cell, and dropping their tail would
// be a new defect rather than a repair.
func TestSiblingMapsClosedTailStillExpands(t *testing.T) {
	cases := []struct {
		name string
		text string
	}{
		{"close paren", "- AC-C-001 (maps REQ-C-001, 002): body.\n"},
		{"colon", "- AC-C-002 maps REQ-C-001, 002: body.\n"},
		{"em dash", "- AC-C-003 maps REQ-C-001, 002 — body.\n"},
		{"end of line", "- AC-C-004 maps REQ-C-001, 002\n"},
		{"table cell pipe", "| AC-C-005 | maps REQ-C-001, 002 | PASS |\n"},
	}

	want := []string{"REQ-C-001", "REQ-C-002"}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := sortedMapsIDs(siblingMapsREQIDs(tc.text))
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("siblingMapsREQIDs = %v, want %v (a closed tail must still expand)", got, want)
			}
		})
	}
}

// TestSiblingMapsOnlyTheTrailingElementIsDropped pins the scope of the
// truncation. Only the element adjacent to the prose is dropped; a tail that sits
// before another element is separated from that prose and stays.
func TestSiblingMapsOnlyTheTrailingElementIsDropped(t *testing.T) {
	text := "- AC-T-001 maps REQ-T-001, 002, 003 is explicitly NOT mapped\n"

	got := sortedMapsIDs(siblingMapsREQIDs(text))
	want := []string{"REQ-T-001", "REQ-T-002"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("siblingMapsREQIDs = %v, want %v (only the trailing tail is dropped)", got, want)
	}
}

// TestSiblingMapsFullIDTrailerUnaffected is the no-regression direction for the
// boundary: a section ending in a FULL id is never truncated, whatever follows.
func TestSiblingMapsFullIDTrailerUnaffected(t *testing.T) {
	text := "- AC-F-001 maps REQ-F-001, REQ-F-002 is prose that follows a full id\n"

	got := sortedMapsIDs(siblingMapsREQIDs(text))
	want := []string{"REQ-F-001", "REQ-F-002"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("siblingMapsREQIDs = %v, want %v (a full-id trailer is not a tail)", got, want)
	}
}
