package cli

import (
	"regexp"
	"testing"
)

// uniqueWorktreeBranchName (card t1696, defect D3): slug tokens truncate to
// the same branch name for sibling cards — "[3.2-0-1 …" and "[3.2-0-3 …"
// both derived WT-3-2-0 and the second rename failed on the live ref. The
// name now uniquifies with a numeric suffix; the card id never enters it.

func TestUniqueWorktreeBranchName(t *testing.T) {
	taken := func(names ...string) func(string) bool {
		set := map[string]bool{}
		for _, n := range names {
			set[n] = true
		}
		return func(name string) bool { return set[name] }
	}
	cases := []struct {
		name   string
		slug   string
		exists func(string) bool
		want   string
	}{
		{"a free base keeps the slug", "3-2-0", taken(), "WT-3-2-0"},
		{"a taken base takes the first suffix", "3-2-0", taken("WT-3-2-0"), "WT-3-2-0-2"},
		{"suffixes grow from two", "3-2-0", taken("WT-3-2-0", "WT-3-2-0-2"), "WT-3-2-0-3"},
		{"a fresh slug is untouched", "alpha", taken("WT-3-2-0"), "WT-alpha"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := uniqueWorktreeBranchName(tc.slug, tc.exists)
			if err != nil {
				t.Fatalf("uniqueWorktreeBranchName(%q): %v", tc.slug, err)
			}
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
			if !regexp.MustCompile(`^WT-[a-z0-9-]+(-[0-9]+)?$`).MatchString(got) {
				t.Fatalf("name %q is not a WT- slug with an optional numeric suffix", got)
			}
		})
	}
}

func TestUniqueWorktreeBranchNameExhaustionRefuses(t *testing.T) {
	all := func(string) bool { return true }
	if _, err := uniqueWorktreeBranchName("3-2-0", all); err == nil {
		t.Fatal("a fully taken namespace returned a name, want the exhaustion refusal")
	}
}

// guard: the suffix budget stays bounded (100), reported not spun.
func TestUniqueWorktreeBranchNameProbeBudget(t *testing.T) {
	probes := 0
	name, err := uniqueWorktreeBranchName("x", func(string) bool {
		probes++
		return probes < 3 // base and WT-x-2 taken; WT-x-3 free
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "WT-x-3" || probes != 3 {
		t.Fatalf("name=%q probes=%d, want WT-x-3 with 3 probes", name, probes)
	}
}
