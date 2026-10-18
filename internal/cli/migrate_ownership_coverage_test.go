package cli

// migrate_ownership_coverage_test.go — SPEC-UPDATE-OWNERSHIP-REPAIR-001 run
// chunk 2 (card t1865): characterization of the small branches the larger
// criteria only touch in passing — the profile-reach mark written under a
// registered key when the walk found a case alias, the counterpart status of
// paths with no counterpart to speak of, and the edges of the remaining-list
// renderer. They go green on arrival: they pin behavior the criteria rely on.

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/internal/userassets"
)

// TestProfileReachMarkUsesTheRegisteredKeyUnderACaseAlias: on a profile-mapped
// project a lower-case spelling that is the registered file is kept and marked
// under the REGISTERED key. The mark lives on a manifest entry; writing it under
// the disk spelling would fail with no entry and abort the whole migration.
func TestProfileReachMarkUsesTheRegisteredKeyUnderACaseAlias(t *testing.T) {
	ownProfileSandbox(t)
	home := installMigrationUserCounterparts(t)
	root, _ := caseProject(t)
	mapProjectToWork(t, root)
	prev := hostCaseFS
	hostCaseFS = emulatedInsensitiveFS()
	t.Cleanup(func() { hostCaseFS = prev })

	r := runOwnershipMigration(t, root, home)
	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatal(err)
	}
	entry, ok := mgr.GetEntry(caseRegisteredRel)
	if !ok || entry.KeptBy != manifest.KeptByProfileReach {
		t.Errorf("entry %s = %+v (found=%v), want the profile-reach mark under the registered key", caseRegisteredRel, entry, ok)
	}
	found := false
	for _, k := range r.rc.keptProfileReach {
		if k == caseRegisteredRel {
			found = true
		}
		if k == caseDiskRel {
			t.Errorf("the outcome tally names the disk spelling %s; the final-state check re-reads the manifest by key", k)
		}
	}
	if !found {
		t.Errorf("keptProfileReach %v does not name the registered key %s", r.rc.keptProfileReach, caseRegisteredRel)
	}
	row, ok := findRemaining(r.rc, caseDiskRel)
	if !ok || row.Reason != keepProfileReach || row.CaseVariantOf != caseRegisteredRel {
		t.Errorf("remaining row %+v (found=%v), want profile-reach and a case variant of %s", row, ok, caseRegisteredRel)
	}
}

// TestAliasToAKeyTheManifestLacksIsNoAlias: a resolver that names a key with no
// entry must not turn an unregistered file into a registered one.
func TestAliasToAKeyTheManifestLacksIsNoAlias(t *testing.T) {
	root, mgr := caseProject(t)
	env := caseEnv(t, root, mgr, emulatedInsensitiveFS())
	env.caseAlias = func(string) (string, bool) { return ".claude/skills/never-registered/SKILL.md", true }
	v := classifyOwnership(env, caseDiskRel, root+"/"+caseDiskRel)
	if v.class != ownershipUnregistered || v.key != "" {
		t.Errorf("verdict class %q key %q, want an unregistered file with no key", v.class, v.key)
	}
}

func TestCounterpartLocationAndStatusEdges(t *testing.T) {
	for _, tc := range []struct {
		rel  string
		slug userassets.RootSlug
		rest string
		ok   bool
	}{
		{".claude/skills/a/SKILL.md", userassets.RootClaudeSkills, "a/SKILL.md", true},
		{".agents/skills/a/SKILL.md", userassets.RootAgentsSkills, "a/SKILL.md", true},
		{".claude/agents/moai/x.md", userassets.RootClaudeAgents, "x.md", true},
		{".codex/agents/moai/x.toml", userassets.RootCodexAgents, "x.toml", true},
		{"README.md", "", "", false},
	} {
		slug, rest, ok := userCounterpartLocation(tc.rel)
		if slug != tc.slug || rest != tc.rest || ok != tc.ok {
			t.Errorf("userCounterpartLocation(%q) = (%q, %q, %v), want (%q, %q, %v)", tc.rel, slug, rest, ok, tc.slug, tc.rest, tc.ok)
		}
	}

	home := installMigrationUserCounterparts(t)
	embedded, err := template.EmbeddedTemplates()
	if err != nil {
		t.Fatal(err)
	}
	cat, err := template.LoadEmbeddedCatalog()
	if err != nil {
		t.Fatal(err)
	}
	um, err := userassets.Load(userassets.ManifestPath(home))
	if err != nil {
		t.Fatal(err)
	}
	status := func(rel string) counterpartStatus {
		return userCounterpartStatus(um, cat, embedded, home, rel)
	}
	if got := status("README.md"); got != counterpartStatusNotApplicable {
		t.Errorf("a path outside the common-asset roots has counterpart status %q, want %q", got, counterpartStatusNotApplicable)
	}
	if got := status(migForeignSkill); got != counterpartStatusNotApplicable {
		t.Errorf("a skill the catalog does not carry has counterpart status %q, want %q (nothing to be a counterpart of)", got, counterpartStatusNotApplicable)
	}
	if got := status(migIdenticalSkill); got != counterpartStatusConfirmed {
		t.Errorf("the installed skill has counterpart status %q, want %q", got, counterpartStatusConfirmed)
	}
	if got := status(".claude/skills/moai-foundation-core/not-installed.md"); got != counterpartStatusAbsent {
		t.Errorf("a catalogued skill's file the user root does not hold has counterpart status %q, want %q", got, counterpartStatusAbsent)
	}
}

func TestRemainingRenderEdges(t *testing.T) {
	t.Run("a verdict that never reached the questions renders n/a, not empty", func(t *testing.T) {
		row := ownershipVerdict{class: ownershipUserModified, keep: keepUserModified}.row("a")
		if row.Counterpart != counterpartStatusNotApplicable || row.ProfileSide != profileSideNotApplicable {
			t.Errorf("row %+v, want n/a for the two states the verdict never established", row)
		}
	})

	rows := []remainingRow{
		{Rel: "z/last", Class: ownershipUnregistered, Reason: keepNoProvenanceRecord, Missing: allEvidenceItems},
		{Rel: "a/first", Class: ownershipUnregistered, Reason: keepNoProvenanceRecord, Missing: []evidenceItem{evidenceOriginalPath}},
		{Rel: "m/odd", Class: ownershipManaged, Reason: ""},
		{Rel: "p/profile", Class: ownershipManaged, Reason: keepProfileReach, Counterpart: counterpartStatusAbsent, ProfileSide: profileSideNotChecked, CaseVariantOf: "p/PROFILE"},
	}

	t.Run("the grouped view names every reason once, in name order", func(t *testing.T) {
		got := remainingSummaryLines(rows)
		if len(got) != 3 {
			t.Fatalf("grouped lines %v, want one per distinct reason (3)", got)
		}
		if !strings.Contains(got[0], "(no reason recorded)") || !strings.Contains(got[1], "no-provenance-record") || !strings.Contains(got[2], "profile-reach") {
			t.Errorf("grouped lines %v are not in reason-name order with the empty reason labelled", got)
		}
		if strings.Contains(got[1], "missing evidence") {
			t.Errorf("a group whose rows lack different evidence items must not claim one list: %q", got[1])
		}
	})

	t.Run("the verbose view is in path order and carries every known state", func(t *testing.T) {
		got := remainingDetailLines(rows)
		if len(got) != 4 || !strings.HasPrefix(got[0], "kept: a/first") || !strings.HasPrefix(got[3], "kept: z/last") {
			t.Fatalf("detail lines %v are not in path order", got)
		}
		profileLine := got[2]
		for _, want := range []string{"user-root counterpart: absent", "profile side: not-checked", "case-variant of: p/PROFILE"} {
			if !strings.Contains(profileLine, want) {
				t.Errorf("detail line %q lacks %q", profileLine, want)
			}
		}
	})

	t.Run("an empty list records nothing", func(t *testing.T) {
		l := newUpdateActionLedger()
		recordRemainingList(l, nil, true)
		recordRemainingList(l, nil, false)
		if !l.empty() {
			t.Errorf("an empty remaining list recorded rows: %+v", l.reference)
		}
	})
}
