package config

import (
	"sort"
	"testing"
)

// TestIsValidConvention verifies the exported convention predicate accepts each
// of the 4 canonical values and rejects out-of-list values. The `custom` engine
// was removed (REQ-WC9-003), so `custom` is now rejected.
func TestIsValidConvention(t *testing.T) {
	t.Parallel()

	canonical := []string{"auto", "conventional-commits", "angular", "karma"}
	for _, name := range canonical {
		if !IsValidConvention(name) {
			t.Errorf("IsValidConvention(%q) = false, want true (canonical value)", name)
		}
	}

	bogus := []string{"custom", "gitflow", "xyz", "Conventional-Commits", "AUTO", " auto", "auto ", ""}
	for _, name := range bogus {
		if IsValidConvention(name) {
			t.Errorf("IsValidConvention(%q) = true, want false (non-canonical)", name)
		}
	}
}

// TestValidConventions verifies the exported list helper returns exactly the 4
// canonical convention names (order-independent). `custom` was removed (REQ-WC9-003).
func TestValidConventions(t *testing.T) {
	t.Parallel()

	got := ValidConventions()
	want := []string{"angular", "auto", "conventional-commits", "karma"}

	gotSorted := append([]string(nil), got...)
	sort.Strings(gotSorted)
	if len(gotSorted) != len(want) {
		t.Fatalf("ValidConventions() returned %d values (%v), want %d", len(gotSorted), got, len(want))
	}
	for i := range want {
		if gotSorted[i] != want[i] {
			t.Errorf("ValidConventions() sorted[%d] = %q, want %q (full: %v)", i, gotSorted[i], want[i], got)
		}
	}
}
