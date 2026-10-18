package cli

// migrate_ownership_profile_test.go — SPEC-UPDATE-OWNERSHIP-REPAIR-001 M2 (card
// t1865), AC-OWR-07: a profile-mapped project keeps its shipped common-asset
// copies whether or not a user-root counterpart exists, and the remaining row
// records WHY — the reason profile-reach, the user-root counterpart's state, and
// a profile-side state that is always `not-checked` (leader ruling, decision
// index Q5: this update never reads a profile root).
//
// The fixture is a profile-mapped project with two agent files and one skill
// file, each registered template_managed with bytes equal to the recorded
// template hash. Three user-root states are run (no counterpart, a confirmed
// one, one whose bytes differ) and a twin project the launch ledger does NOT map
// is the reach control: the same inputs REMOVE the copies, so the keep is shown
// to come from the profile mapping and not from an unreachable removal arm.
//
// The preservation assertion is a pure function returning one problem per
// violated expectation; the mutant sub-tests inject a wrong classifier through
// the same seam the migration uses and require a non-empty list.

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/userassets"
)

const (
	profAgentA = ".claude/agents/moai/manager-develop.md"
	profAgentB = ".claude/agents/moai/manager-docs.md"
	profSkill  = migIdenticalSkill
)

// profileMappedRels are the three shipped copies the fixture plants.
var profileMappedRels = []string{profAgentA, profAgentB, profSkill}

// profileMappedFixture plants the three shipped copies (registered
// template_managed, healthy) in a fresh project and returns the root and the
// bytes each file must still carry afterwards. mapped records the project under
// the profile "work" in the sandbox launch ledger.
func profileMappedFixture(t *testing.T, mapped bool) (string, map[string][]byte) {
	t.Helper()
	root := t.TempDir()
	want := map[string][]byte{}
	for _, rel := range profileMappedRels {
		data := ownEmbeddedBytes(t, rel)
		ownAddRegisteredManaged(t, root, rel, data)
		want[rel] = data
	}
	if mapped {
		mapProjectToWork(t, root)
	}
	return root, want
}

// profileKeepProblems is the pure assertion of AC-OWR-07: every shipped copy is
// still on disk with its bytes, nothing was removed or copied into the run
// backup, each copy has a remaining row recording the profile-reach reason, the
// expected user-root counterpart state and the `not-checked` profile side, and
// each manifest entry carries the profile-reach mark.
func profileKeepProblems(root string, run ownershipRun, wantCounterpart counterpartStatus, want map[string][]byte) []string {
	var problems []string
	if len(run.removed) != 0 {
		problems = append(problems, fmt.Sprintf("removed %v, want nothing removed", run.removed))
	}
	mgr := manifest.NewManager()
	_, loadErr := mgr.Load(root)
	if loadErr != nil {
		problems = append(problems, fmt.Sprintf("reload manifest: %v", loadErr))
	}
	for _, rel := range profileMappedRels {
		got, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		switch {
		case err != nil:
			problems = append(problems, fmt.Sprintf("%s: gone or unreadable: %v", rel, err))
		case string(got) != string(want[rel]):
			problems = append(problems, fmt.Sprintf("%s: bytes changed", rel))
		}
		if matches, _ := filepath.Glob(filepath.Join(root, ".moai-backups", "*", "removed-assets", filepath.FromSlash(rel))); len(matches) != 0 {
			problems = append(problems, fmt.Sprintf("%s: a removed-assets copy exists %v", rel, matches))
		}
		row, ok := findRemaining(run.rc, rel)
		if !ok {
			problems = append(problems, fmt.Sprintf("%s: no remaining row", rel))
		} else {
			if row.Class != ownershipManaged {
				problems = append(problems, fmt.Sprintf("%s: class %q, want %q", rel, row.Class, ownershipManaged))
			}
			if row.Reason != keepProfileReach {
				problems = append(problems, fmt.Sprintf("%s: reason %q, want %q", rel, row.Reason, keepProfileReach))
			}
			if row.Counterpart != wantCounterpart {
				problems = append(problems, fmt.Sprintf("%s: user-root counterpart %q, want %q", rel, row.Counterpart, wantCounterpart))
			}
			if row.ProfileSide != profileSideNotChecked {
				problems = append(problems, fmt.Sprintf("%s: profile side %q, want %q", rel, row.ProfileSide, profileSideNotChecked))
			}
		}
		if loadErr == nil {
			if e, found := mgr.GetEntry(rel); !found {
				problems = append(problems, fmt.Sprintf("%s: manifest entry dropped", rel))
			} else if e.KeptBy != manifest.KeptByProfileReach {
				problems = append(problems, fmt.Sprintf("%s: KeptBy %q, want %q", rel, e.KeptBy, manifest.KeptByProfileReach))
			}
		}
	}
	return problems
}

// userCounterpartPath is the on-disk user-root file that mirrors a project
// agent path, resolved through the user-asset root directory API.
func userCounterpartPath(home, agentRel string) string {
	return filepath.Join(userassets.RootBySlugDir(home, userassets.RootClaudeAgents), filepath.Base(agentRel))
}

func TestProfileMappedCopyKeptWithoutCounterpart(t *testing.T) {
	enableRemovalForTest(t)
	ownProfileSandbox(t)

	t.Run("NoCounterpart", func(t *testing.T) {
		root, want := profileMappedFixture(t, true)
		home := t.TempDir() // an empty user root: no counterpart for any copy
		run := runOwnershipMigration(t, root, home)
		for _, p := range profileKeepProblems(root, run, counterpartStatusAbsent, want) {
			t.Error(p)
		}
	})

	t.Run("ConfirmedCounterpartStillKept", func(t *testing.T) {
		root, want := profileMappedFixture(t, true)
		home := installMigrationUserCounterparts(t)
		for _, rel := range profileMappedRels {
			requireUserCounterpart(t, home, rel)
		}
		run := runOwnershipMigration(t, root, home)
		for _, p := range profileKeepProblems(root, run, counterpartStatusConfirmed, want) {
			t.Error(p)
		}
	})

	t.Run("DifferingCounterpartIsUnconfirmed", func(t *testing.T) {
		root, want := profileMappedFixture(t, true)
		home := installMigrationUserCounterparts(t)
		// Every counterpart is present and tracked; the agents' bytes now differ
		// from what the manifest recorded, the skill's are removed from the
		// recorded hash by an append. All three are present-but-not-confirmable.
		for _, rel := range []string{profAgentA, profAgentB} {
			p := userCounterpartPath(home, rel)
			if err := os.WriteFile(p, []byte("a user-root file whose bytes differ\n"), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		skillCounterpart := filepath.Join(userassets.RootBySlugDir(home, userassets.RootClaudeSkills),
			filepath.FromSlash("moai-foundation-core/SKILL.md"))
		if err := os.WriteFile(skillCounterpart, []byte("a user-root skill whose bytes differ\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		run := runOwnershipMigration(t, root, home)
		for _, p := range profileKeepProblems(root, run, counterpartStatusUnconfirmed, want) {
			t.Error(p)
		}
	})

	t.Run("ReachControlUnmappedTwinRemoves", func(t *testing.T) {
		root, _ := profileMappedFixture(t, false)
		home := installMigrationUserCounterparts(t)
		run := runOwnershipMigration(t, root, home)
		for _, rel := range profileMappedRels {
			if !slices.Contains(run.removed, rel) {
				t.Errorf("%s was not removed in the unmapped twin (removed=%v): the keep above could be a dead removal arm", rel, run.removed)
			}
			assertFileAbsent(t, root, rel)
		}
	})

	t.Run("MutantIgnoringTheMappingIsCaught", func(t *testing.T) {
		root, want := profileMappedFixture(t, true)
		home := installMigrationUserCounterparts(t)
		swapClassifier(t, func(env ownershipEnv, rel, abs string) ownershipVerdict {
			env.profileMapped = false
			return classifyOwnership(env, rel, abs)
		})
		run := runOwnershipMigration(t, root, home)
		if got := profileKeepProblems(root, run, counterpartStatusConfirmed, want); len(got) == 0 {
			t.Errorf("the assertion stayed green in front of a classifier that ignores the profile mapping")
		}
	})

	t.Run("MutantPretendingToReadTheProfileSideIsCaught", func(t *testing.T) {
		root, want := profileMappedFixture(t, true)
		home := t.TempDir()
		swapClassifier(t, func(env ownershipEnv, rel, abs string) ownershipVerdict {
			v := classifyOwnership(env, rel, abs)
			v.profileSide = profileSideStatus("confirmed")
			return v
		})
		run := runOwnershipMigration(t, root, home)
		if got := profileKeepProblems(root, run, counterpartStatusAbsent, want); len(got) == 0 {
			t.Errorf("the assertion stayed green in front of a classifier that records the profile side as read")
		}
	})
}
