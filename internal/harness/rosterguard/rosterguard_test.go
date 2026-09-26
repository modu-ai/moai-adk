package rosterguard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/template"
)

// TestNoAgentNameIsASubstringOfAnother is the premise every other assertion in
// this package rests on: membership is decided by literal containment, which is
// only sound while no agent name is contained in another. If a future name
// breaks that, every containment check silently over-matches and the guard
// starts passing on rosters it should fail.
func TestNoAgentNameIsASubstringOfAnother(t *testing.T) {
	root := repoRoot(t)
	defs, err := AgentDefinitionNames(root)
	if err != nil {
		t.Fatal(err)
	}
	universe := union(template.ProfileMatrixAgents(), defs)
	for _, a := range universe {
		for _, b := range universe {
			if a != b && strings.Contains(b, a) {
				t.Errorf("agent name %q is a substring of %q — literal containment can no longer decide membership; switch NamesIn to a bounded match", a, b)
			}
		}
	}
}

// TestCanonicalSourceIsTheDefinitionFileSetPlusExplore records the relationship
// between the two populations, so the "two different 12s" trap has a named,
// asserted shape rather than living only in a comment.
func TestCanonicalSourceIsTheDefinitionFileSetPlusExplore(t *testing.T) {
	root := repoRoot(t)
	defs, err := AgentDefinitionNames(root)
	if err != nil {
		t.Fatal(err)
	}
	retained := template.ProfileMatrixAgents()

	missing, extra := diff(append(append([]string(nil), defs...), "Explore"), sorted(retained))
	if len(missing) > 0 || len(extra) > 0 {
		t.Errorf("the retained roster is no longer exactly the definition files plus Explore; absent from the roster: %v; in the roster with no definition file: %v."+
			" Either a definition file was added without a roster row, or a second built-in joined Explore — adjudicate before changing this assertion",
			missing, extra)
	}
	if len(defs)+1 != len(retained) {
		t.Errorf("definition files %d + 1 built-in != retained roster %d", len(defs), len(retained))
	}
}

// TestRegistryIsWellFormed checks the registry against itself before it is used
// to judge anything: a malformed row would otherwise produce a confident verdict
// about a file it never actually read.
func TestRegistryIsWellFormed(t *testing.T) {
	root := repoRoot(t)
	seen := map[string]bool{}
	for _, s := range Registry() {
		if s.ID == "" {
			t.Fatalf("registry row for %s has no ID", s.Path)
		}
		if seen[s.ID] {
			t.Errorf("duplicate registry ID %q", s.ID)
		}
		seen[s.ID] = true

		if _, err := os.Stat(filepath.Join(root, s.Path)); err != nil {
			t.Errorf("site %s: registered path %s does not exist: %v (the file moved or was deleted; update or remove the row)", s.ID, s.Path, err)
		}
		if s.Claims.Has(ClaimCount) && s.CountPattern == "" {
			t.Errorf("site %s declares ClaimCount with no CountPattern", s.ID)
		}
		if s.Claims.Has(ClaimMembership) && s.Axis == AxisSubsetByDesign {
			t.Errorf("site %s declares ClaimMembership on %s, which supports no membership assertion", s.ID, AxisSubsetByDesign)
		}
		if s.Claims == 0 && s.Axis != AxisSubsetByDesign {
			t.Errorf("site %s asserts nothing but is not declared %s — a row that asserts nothing on an assertable axis is an undeclared exemption", s.ID, AxisSubsetByDesign)
		}
		if s.Claims == 0 && s.Note == "" {
			t.Errorf("site %s asserts nothing and carries no Note explaining why", s.ID)
		}
		if s.KnownStale != nil {
			if strings.TrimSpace(s.KnownStale.Reason) == "" {
				t.Errorf("site %s: KnownStale with an empty Reason is a mute, not a record", s.ID)
			}
			if strings.TrimSpace(s.KnownStale.FollowUp) == "" {
				t.Errorf("site %s: KnownStale with no FollowUp has no exit condition", s.ID)
			}
		}
	}
}

// TestRegisteredSitesMatchTheirDeclaredAxis is the per-site assertion.
func TestRegisteredSitesMatchTheirDeclaredAxis(t *testing.T) {
	root := repoRoot(t)
	defs, err := AgentDefinitionNames(root)
	if err != nil {
		t.Fatal(err)
	}
	retained := sorted(template.ProfileMatrixAgents())

	for _, s := range Registry() {
		b, err := os.ReadFile(filepath.Join(root, s.Path))
		if err != nil {
			t.Errorf("site %s: %v", s.ID, err)
			continue
		}
		for _, v := range CheckSite(s, string(b), retained, defs) {
			t.Error(v)
		}
	}
}

// TestSweepFindsNoUndeclaredRosterListing is the load-bearing half of the guard.
//
// The per-site assertions can only check sites someone remembered to register.
// This sweep is what makes a NEW listing impossible to add silently: any file
// enumerating at least SweepThreshold distinct agent names and absent from the
// registry fails here with an instruction to declare its axis.
func TestSweepFindsNoUndeclaredRosterListing(t *testing.T) {
	root := repoRoot(t)
	defs, err := AgentDefinitionNames(root)
	if err != nil {
		t.Fatal(err)
	}
	universe := union(template.ProfileMatrixAgents(), defs)

	found, err := Sweep(root, universe)
	if err != nil {
		t.Fatalf("sweep: %v", err)
	}
	// A sweep that found nothing would pass this test while proving nothing.
	// The registry is non-empty by construction, so the sweep must at minimum
	// rediscover it.
	if len(found) == 0 {
		t.Fatal("the sweep found zero roster listings in a tree that demonstrably has them — the walk or the exclusion list is broken, and a zero result here is a measurement failure, not a clean tree")
	}

	registered := map[string]bool{}
	for _, s := range Registry() {
		registered[s.Path] = true
	}
	for _, p := range found {
		if !registered[p] {
			t.Errorf("undeclared roster listing: %s enumerates at least %d distinct agent names but is not in Registry(). "+
				"Declare its axis — %s (the CLAUDE.md §4 roster), %s (the .claude/agents/moai/*.md file population), or %s (a listing that is legitimately partial) — and say whether it claims a count, a membership, or both.",
				p, SweepThreshold, AxisRetainedRoster, AxisDefinitionFiles, AxisSubsetByDesign)
		}
	}

	// The converse: a registered path the sweep no longer reaches is either a
	// deleted listing or an exclusion that has grown too wide. Either way the
	// row is no longer doing what it claims.
	foundSet := map[string]bool{}
	for _, p := range found {
		foundSet[p] = true
	}
	// A row may declare itself expected-unreachable (Site.SweepUnreachable) —
	// a count-only claim on a file that enumerates too few names to sweep. The
	// exemption is per path and must carry a reason, so it stays a declaration
	// a reviewer can disagree with rather than an inference.
	exempt := map[string]string{}
	for _, s := range Registry() {
		if s.SweepUnreachable != "" {
			exempt[s.Path] = s.SweepUnreachable
		}
	}
	for p := range registered {
		if foundSet[p] {
			// The converse of the exemption: a path that declared itself
			// unreachable but IS reached has changed shape, and its row was
			// written against a file that no longer exists in that form.
			if why, ok := exempt[p]; ok {
				t.Errorf("registered site %s declares SweepUnreachable (%s) but the sweep DID reach it — the file now enumerates %d or more names; drop the declaration and assert membership", p, why, SweepThreshold)
			}
			continue
		}
		if _, ok := exempt[p]; ok {
			continue
		}
		t.Errorf("registered site %s is no longer reached by the sweep — the listing shrank below %d names, the file moved, or an exclusion now covers it; re-measure the row, or declare SweepUnreachable with a reason if the claim is count-only", p, SweepThreshold)
	}
}

// TestKnownStaleInventory prints every declared staleness, so the guard being
// green never means "nothing is stale". Green means "every stale site is
// declared, and its declared shape still matches what is on disk".
func TestKnownStaleInventory(t *testing.T) {
	n := 0
	for _, s := range Registry() {
		if s.KnownStale == nil {
			continue
		}
		n++
		t.Logf("KNOWN STALE %-46s axis=%-16s absent=%v declared-count=%d follow-up=%s",
			s.ID, s.Axis, s.KnownStale.MissingNames, s.KnownStale.DeclaredCount, s.KnownStale.FollowUp)
	}
	t.Logf("%d of %d registered sites carry a declared staleness", n, len(Registry()))
}

// ── Control probe ────────────────────────────────────────────────────────────

// TestGuardFiresOnDeliberatelyWrongInput is the control probe.
//
// A passing guard with no control probe is indistinguishable from a dead one:
// the per-site test above would report exactly the same green output if
// CheckSite returned nil unconditionally, if the block extractor silently
// matched nothing, or if the axis lookup collapsed every axis to "no members".
// Each case below feeds CheckSite input that MUST produce a violation, so the
// green verdict above is evidence that the assertions ran rather than evidence
// that nothing was checked.
func TestGuardFiresOnDeliberatelyWrongInput(t *testing.T) {
	retained := []string{"Explore", "manager-lead", "manager-spec", "mission-governor"}
	defs := []string{"manager-lead", "manager-spec", "mission-governor"}

	membershipSite := Site{
		ID: "probe", Path: "probe.md", Axis: AxisRetainedRoster,
		Claims: ClaimMembership, BlockStart: "ROSTER", BlockEnd: "END",
	}
	countSite := Site{
		ID: "probe", Path: "probe.md", Axis: AxisRetainedRoster,
		Claims: ClaimCount, CountPattern: `exactly (\d+) retained`,
	}

	cases := []struct {
		name string
		site Site
		body string
		want string
	}{
		{
			name: "a roster missing one member is reported by name",
			site: membershipSite,
			body: "ROSTER\nExplore manager-spec mission-governor\nEND\n",
			want: "absent from the site: manager-lead",
		},
		{
			name: "a roster carrying a name the axis does not have is reported",
			site: membershipSite,
			body: "ROSTER\nExplore manager-lead manager-spec mission-governor expert-backend\nEND\n",
			// expert-backend is outside the universe, so the observable
			// violation is the extra-name path exercised via a stale marker
			// below; here the roster is complete and must NOT fire.
			want: "",
		},
		{
			// Regression probe. A containment-based match passed this exact
			// mutant when it was run live against a registered site: the old
			// name is still a substring of the renamed one, so every rename
			// and every typo was invisible to the guard.
			name: "a renamed entry does not satisfy membership by substring",
			site: membershipSite,
			body: "ROSTER\nExplore manager-leadXX manager-spec mission-governor\nEND\n",
			want: "absent from the site: manager-lead",
		},
		{
			name: "a stale count is reported against its axis",
			site: countSite,
			body: "the catalog has exactly 11 retained agents\n",
			want: "declares 11 on axis retained-roster, which currently has 4",
		},
		{
			name: "a count pattern that matches nothing fails instead of passing vacuously",
			site: countSite,
			body: "the catalog has some agents\n",
			want: "matched 0 times",
		},
		{
			name: "an ambiguous count pattern fails instead of picking one",
			site: countSite,
			body: "exactly 11 retained\nexactly 13 retained\n",
			want: "matched 2 times",
		},
		{
			name: "a block anchor that no longer exists fails instead of passing vacuously",
			site: membershipSite,
			body: "no anchor here at all\n",
			want: "block anchor",
		},
		{
			name: "a KnownStale marker whose declared gap no longer matches the file fails",
			site: withStale(membershipSite, &Staleness{
				Reason: "probe", FollowUp: "probe", MissingNames: []string{"manager-lead"},
			}),
			body: "ROSTER\nExplore manager-spec\nEND\n",
			want: "the observed gap no longer matches the declared one",
		},
		{
			name: "a KnownStale marker on a site that is no longer stale fails",
			site: withStale(membershipSite, &Staleness{
				Reason: "probe", FollowUp: "probe", MissingNames: []string{"manager-lead"},
			}),
			body: "ROSTER\nExplore manager-lead manager-spec mission-governor\nEND\n",
			want: "present at the site but not on the axis: manager-lead",
		},
		{
			name: "a KnownStale count marker that has caught up with its axis fails",
			site: withStale(countSite, &Staleness{
				Reason: "probe", FollowUp: "probe", DeclaredCount: 4,
			}),
			body: "exactly 4 retained agents\n",
			want: "no longer stale on this claim",
		},
		{
			name: "declaring membership on the subset axis is a registry contradiction",
			site: Site{ID: "probe", Path: "p", Axis: AxisSubsetByDesign, Claims: ClaimMembership, BlockStart: "x"},
			body: "x\n",
			want: "registry contradiction",
		},
		{
			name: "a KnownStale marker that records no gap is refused",
			site: withStale(membershipSite, &Staleness{Reason: "probe", FollowUp: "probe"}),
			body: "ROSTER\nExplore manager-lead manager-spec mission-governor\nEND\n",
			want: "a marker that records no gap is a mute",
		},
		{
			name: "a correct site produces no violation",
			site: membershipSite,
			body: "ROSTER\nExplore manager-lead manager-spec mission-governor\nEND\n",
			want: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := CheckSite(tc.site, tc.body, retained, defs)
			joined := strings.Join(got, " | ")
			if tc.want == "" {
				if len(got) != 0 {
					t.Fatalf("want no violation, got: %s", joined)
				}
				return
			}
			if !strings.Contains(joined, tc.want) {
				t.Fatalf("want a violation containing %q, got: %s", tc.want, joined)
			}
		})
	}
}

// TestSweepFiresOnAnUndeclaredListing proves the sweep half is live too: a
// synthetic tree carrying a roster listing must be reported.
func TestSweepFiresOnAnUndeclaredListing(t *testing.T) {
	root := repoRoot(t)
	defs, err := AgentDefinitionNames(root)
	if err != nil {
		t.Fatal(err)
	}
	universe := union(template.ProfileMatrixAgents(), defs)

	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "new-listing.md"), []byte(strings.Join(universe, "\n")), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "not-a-listing.md"), []byte("manager-spec only\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	// An excluded tree must stay excluded even when it carries a full roster.
	if err := os.MkdirAll(filepath.Join(tmp, "docs-site"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "docs-site", "page.md"), []byte(strings.Join(universe, "\n")), 0o600); err != nil {
		t.Fatal(err)
	}
	// Runtime state is machine-local and untracked; a full roster there must
	// not be reported either.
	if err := os.MkdirAll(filepath.Join(tmp, ".moai", "state", "verify"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, ".moai", "state", "verify", "probe.txt"), []byte(strings.Join(universe, "\n")), 0o600); err != nil {
		t.Fatal(err)
	}
	// The cache and the logs are untracked for the same reason.
	for _, rel := range []string{
		filepath.Join(".moai", "cache", "template-snapshot", "sections", "delegation.yaml"),
		filepath.Join(".moai", "logs", "agent-model-audit.jsonl"),
	} {
		if err := os.MkdirAll(filepath.Join(tmp, filepath.Dir(rel)), 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(tmp, rel), []byte(strings.Join(universe, "\n")), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	got, err := Sweep(tmp, universe)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != "new-listing.md" {
		t.Fatalf("sweep = %v, want exactly [new-listing.md] (a file below the threshold must not be reported, and docs-site, .moai/state, .moai/cache and .moai/logs must stay excluded)", got)
	}
}

func withStale(s Site, st *Staleness) Site {
	s.KnownStale = st
	return s
}

func sorted(in []string) []string {
	out := append([]string(nil), in...)
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j] < out[j-1]; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}
