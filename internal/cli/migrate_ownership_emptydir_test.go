package cli

// migrate_ownership_emptydir_test.go — SPEC-UPDATE-OWNERSHIP-REPAIR-001 M5
// (card t1865), AC-OWR-13: a skill directory the migration finds empty is
// removed only when the manifest records (the start-of-run manifest and the
// retained prior generations) show that every file once recorded under it was
// template-managed and none remains; otherwise it is kept and listed with the
// reason `empty-dir-no-provenance`. The pruning is a separate step after the
// walk, and it does not depend on the removal switch.
//
// Time independence: every case here is ONE migration run over a fixture whose
// prior generation is a synthetic, fixed file. No assertion depends on which
// clock second a run lands in.

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/manifest"
)

// The empty-directory cases (acceptance.md AC-OWR-13). Each constant is the
// project-relative slash path of the directory.
const (
	emptyDirPrevRunEmptied = ".claude/skills/moai-ghost-a"      // (A) a previous run emptied it
	emptyDirStartManifest  = ".claude/skills/moai-ghost-s"      // (A') only the start-of-run manifest remembers it
	emptyDirNestedRoot     = ".claude/skills/moai-ghost-c"      // (C) nested empty directories
	emptyDirNestedMid      = ".claude/skills/moai-ghost-c/refs" //     ... the middle one
	emptyDirNestedLeaf     = ".claude/skills/moai-ghost-c/refs/deep"
	emptyDirNoRecord       = ".claude/skills/moai-ghost-d"  // (D) no record anywhere
	emptyDirUserRecord     = ".claude/skills/moai-ghost-e"  // (E) a user_created record
	emptyDirMixedParent    = ".claude/skills/moai-ghost-cm" // a user_created sibling record keeps the parent
	emptyDirMixedChild     = ".claude/skills/moai-ghost-cm/refs"
	emptyDirWithUserFile   = ".claude/skills/moai-user-f" // (F) holds a user file
	emptyDirMirrorPruned   = ".agents/skills/moai-ghost-m"
	emptyDirMirrorNoRecord = ".agents/skills/moai-ghost-n"
	emptyDirAgentDir       = ".claude/agents/moai" // never a candidate
)

const emptyDirUserFileRel = emptyDirWithUserFile + "/notes.md"

var emptyDirUserFileBytes = []byte("# my own notes\n")

// emptyDirPriorFiles is the synthetic prior generation: what an earlier run's
// start-of-run manifest recorded for files that are now gone.
func emptyDirPriorFiles() map[string]ownPriorRecord {
	managed := func(rel string) ownPriorRecord {
		h := manifest.HashBytes([]byte("prior " + rel))
		return ownPriorRecord{Provenance: manifest.TemplateManaged, TemplateHash: h, DeployedHash: h, DeployedVersion: "v-prior"}
	}
	user := func(rel string) ownPriorRecord {
		h := manifest.HashBytes([]byte("prior " + rel))
		return ownPriorRecord{Provenance: manifest.UserCreated, TemplateHash: h, DeployedHash: h, DeployedVersion: "v-prior"}
	}
	files := map[string]ownPriorRecord{}
	for _, rel := range []string{
		emptyDirPrevRunEmptied + "/SKILL.md",
		emptyDirPrevRunEmptied + "/references/guide.md",
		emptyDirNestedRoot + "/SKILL.md",
		emptyDirNestedLeaf + "/x.md",
		emptyDirMixedChild + "/x.md",
		emptyDirWithUserFile + "/SKILL.md",
		emptyDirMirrorPruned + "/SKILL.md",
		emptyDirAgentDir + "/ghost.md",
	} {
		files[rel] = managed(rel)
	}
	files[emptyDirMixedParent+"/mine.md"] = user(emptyDirMixedParent + "/mine.md")
	files[emptyDirUserRecord+"/SKILL.md"] = user(emptyDirUserRecord + "/SKILL.md")
	return files
}

// plantEmptyDirCases builds every case under root. It writes the prior
// generation as a retained run backup and registers (then deletes) one file so
// the start-of-run manifest remembers it.
func plantEmptyDirCases(t *testing.T, root string) {
	t.Helper()
	for _, rel := range []string{
		emptyDirPrevRunEmptied, emptyDirStartManifest, emptyDirNestedLeaf, emptyDirNoRecord,
		emptyDirUserRecord, emptyDirMixedChild, emptyDirMirrorPruned, emptyDirMirrorNoRecord, emptyDirAgentDir,
	} {
		if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(rel)), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", rel, err)
		}
	}
	// (F) a directory that holds a user file; its recorded shipped file is gone.
	ownWriteFile(t, root, emptyDirUserFileRel, emptyDirUserFileBytes)
	// (A') the manifest the run starts from still carries the record of a file
	// whose directory was already emptied (a previous run died before its Save).
	startRel := emptyDirStartManifest + "/SKILL.md"
	ownAddRegisteredManaged(t, root, startRel, []byte("start-of-run record "+startRel))
	if err := os.Remove(filepath.Join(root, filepath.FromSlash(startRel))); err != nil {
		t.Fatalf("remove %s: %v", startRel, err)
	}
	ownWritePriorGeneration(t, root, "20260101_000001",
		ownPriorGenerationJSON(t, "v-prior", "2026-01-01T00:00:01Z", emptyDirPriorFiles()))
}

// emptyDirFixture returns a project carrying every case, the user counterparts
// the migration needs, and the removal switch in the requested position.
func emptyDirFixture(t *testing.T, switchOn bool) (root, home string) {
	t.Helper()
	if switchOn {
		root, home = ownershipFixtureBase(t)
	} else {
		ownProfileSandbox(t)
		root = buildMigrationFixture(t)
		home = installMigrationUserCounterparts(t)
	}
	plantEmptyDirCases(t, root)
	return root, home
}

// emptyDirExpect is what one directory must look like after the run.
type emptyDirExpect struct {
	rel string
	// gone is true when the directory must have been removed.
	gone bool
	// class is the row class when the directory is kept with a row.
	class ownershipClass
	// row is true when a kept directory must be listed with the reason
	// `empty-dir-no-provenance`.
	row bool
}

// emptyDirProblems is the pure assertion: one problem per violated expectation.
// reach is a directory that must have been pruned in the same run, so a keep is
// never read against a step that did not run. filesBefore is the file set before
// the run: the run may lose exactly the files it reports as removed, so the
// pruning step is shown to have deleted no file.
func emptyDirProblems(root string, run ownershipRun, filesBefore map[string]bool, reach string, exps []emptyDirExpect) []string {
	var problems []string
	exists := func(rel string) bool {
		_, err := os.Lstat(filepath.Join(root, filepath.FromSlash(rel)))
		return err == nil
	}
	if reach != "" && exists(reach) {
		problems = append(problems, fmt.Sprintf("reach control %s was not pruned: the pruning step did not run, so a keep below proves nothing", reach))
	}
	for _, e := range exps {
		present := exists(e.rel)
		row, hasRow := findRemaining(run.rc, e.rel)
		switch {
		case e.gone && present:
			problems = append(problems, fmt.Sprintf("%s: still present, want it removed", e.rel))
		case !e.gone && !present:
			problems = append(problems, fmt.Sprintf("%s: removed, want it kept", e.rel))
		}
		if e.gone && hasRow {
			problems = append(problems, fmt.Sprintf("%s: removed but still listed as remaining (reason %q)", e.rel, row.Reason))
		}
		if !e.gone && e.row {
			switch {
			case !hasRow:
				problems = append(problems, fmt.Sprintf("%s: kept without a remaining row", e.rel))
			case row.Reason != keepEmptyDirNoProvenance:
				problems = append(problems, fmt.Sprintf("%s: reason %q, want %q", e.rel, row.Reason, keepEmptyDirNoProvenance))
			case row.Class != e.class:
				problems = append(problems, fmt.Sprintf("%s: class %q, want %q", e.rel, row.Class, e.class))
			}
		}
	}
	after := emptyDirFileSet(root)
	removed := map[string]bool{}
	for _, r := range run.removed {
		removed[r] = true
	}
	var lost []string
	for rel := range filesBefore {
		if !after[rel] && !removed[rel] {
			lost = append(lost, rel)
		}
	}
	sort.Strings(lost)
	if len(lost) > 0 {
		problems = append(problems, fmt.Sprintf("files vanished that the run did not report as removed: %v", lost))
	}
	return problems
}

// emptyDirFileSet is the set of regular files and symlinks under the project
// (run backups excluded).
func emptyDirFileSet(root string) map[string]bool {
	out := map[string]bool{}
	_ = filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, relErr := filepath.Rel(root, p)
		if relErr != nil || rel == "." {
			return nil
		}
		key := filepath.ToSlash(rel)
		if strings.HasPrefix(key+"/", ".moai-backups/") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.IsDir() {
			out[key] = true
		}
		return nil
	})
	return out
}

// emptyDirStandardExpectations is the full AC-OWR-13 matrix. removeB is true
// when the removal switch is on (the case where this run empties a directory).
func emptyDirStandardExpectations(switchOn bool) []emptyDirExpect {
	exps := []emptyDirExpect{
		{rel: emptyDirPrevRunEmptied, gone: true},
		{rel: emptyDirStartManifest, gone: true},
		{rel: emptyDirNestedRoot, gone: true},
		{rel: emptyDirNestedMid, gone: true},
		{rel: emptyDirNestedLeaf, gone: true},
		{rel: emptyDirMirrorPruned, gone: true},
		{rel: emptyDirNoRecord, class: ownershipUnregistered, row: true},
		{rel: emptyDirUserRecord, class: ownershipUserCreated, row: true},
		{rel: emptyDirMixedParent, class: ownershipUserCreated, row: true},
		{rel: emptyDirMixedChild, gone: true},
		{rel: emptyDirMirrorNoRecord, class: ownershipUnregistered, row: true},
		{rel: emptyDirWithUserFile},
		{rel: ".claude/skills"},
		{rel: ".agents/skills"},
		{rel: emptyDirAgentDir},
	}
	if switchOn {
		// (B) this run removed the last file of the skill directory.
		exps = append(exps, emptyDirExpect{rel: ".claude/skills/moai-foundation-core", gone: true})
	} else {
		exps = append(exps, emptyDirExpect{rel: ".claude/skills/moai-foundation-core"})
	}
	return exps
}

// TestManagedOnlyEmptyDirsPruned is AC-OWR-13.
func TestManagedOnlyEmptyDirsPruned(t *testing.T) {
	t.Run("Premise", func(t *testing.T) {
		enableRemovalForTest(t)
		root, _ := emptyDirFixture(t, true)
		for _, rel := range []string{
			emptyDirPrevRunEmptied, emptyDirStartManifest, emptyDirNestedLeaf, emptyDirNoRecord,
			emptyDirUserRecord, emptyDirMixedChild, emptyDirMirrorPruned, emptyDirMirrorNoRecord, emptyDirAgentDir,
		} {
			entries, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(rel)))
			if err != nil || len(entries) != 0 {
				t.Errorf("fixture error: %s is not an empty directory (entries=%d, err=%v)", rel, len(entries), err)
			}
		}
		if _, err := os.Stat(filepath.Join(root, ".moai-backups", "20260101_000001", priorManifestName)); err != nil {
			t.Errorf("fixture error: the prior generation is not on disk: %v", err)
		}
		mgr := manifest.NewManager()
		if _, err := mgr.Load(root); err != nil {
			t.Fatalf("load manifest: %v", err)
		}
		if _, ok := mgr.GetEntry(emptyDirStartManifest + "/SKILL.md"); !ok {
			t.Errorf("fixture error: the start-of-run manifest does not remember %s", emptyDirStartManifest)
		}
	})

	t.Run("PrunesManagedOnlyAndKeepsTheRest", func(t *testing.T) {
		enableRemovalForTest(t)
		root, home := emptyDirFixture(t, true)
		before := emptyDirFileSet(root)
		run := runOwnershipMigration(t, root, home)
		// The reach control comes first: the registered sibling with every
		// removal precondition is removed in this very run and its directory
		// goes with it (case B).
		if !slices.Contains(run.removed, migIdenticalSkill) {
			t.Fatalf("the reach control %s was not removed (removed=%v)", migIdenticalSkill, run.removed)
		}
		for _, p := range emptyDirProblems(root, run, before, emptyDirPrevRunEmptied, emptyDirStandardExpectations(true)) {
			t.Error(p)
		}
		if got, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(emptyDirUserFileRel))); err != nil || string(got) != string(emptyDirUserFileBytes) {
			t.Errorf("the user file in %s changed or vanished (err=%v)", emptyDirWithUserFile, err)
		}
		for _, rel := range []string{emptyDirPrevRunEmptied, emptyDirNestedRoot, emptyDirNestedMid, emptyDirNestedLeaf} {
			if !slices.ContainsFunc(run.report, func(l string) bool { return strings.Contains(l, rel) && strings.Contains(l, "removed empty") }) {
				t.Errorf("no report line names the pruned directory %s: %q", rel, run.report)
			}
		}
	})

	t.Run("IndependentOfTheRemovalSwitch", func(t *testing.T) {
		root, home := emptyDirFixture(t, false)
		before := emptyDirFileSet(root)
		run := runOwnershipMigration(t, root, home)
		if len(run.removed) != 0 {
			t.Fatalf("the removal switch is off, yet the run removed %v", run.removed)
		}
		for _, p := range emptyDirProblems(root, run, before, emptyDirPrevRunEmptied, emptyDirStandardExpectations(false)) {
			t.Error(p)
		}
	})

	t.Run("RootsAndAgentDirectoriesAreNeverCandidates", func(t *testing.T) {
		enableRemovalForTest(t)
		root, home := emptyDirFixture(t, true)
		// Every skill directory goes; the roots and the agent directory must not.
		run := runOwnershipMigration(t, root, home)
		for _, rel := range []string{".claude/skills", ".agents/skills", emptyDirAgentDir} {
			if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel))); err != nil {
				t.Errorf("%s was removed (%v)", rel, err)
			}
			if row, ok := findRemaining(run.rc, rel); ok {
				t.Errorf("%s is listed as a remaining directory (reason %q): roots and agent directories are not candidates", rel, row.Reason)
			}
		}
		// The recorded agent file is managed and gone, the directory is empty:
		// the same evidence that prunes a skill directory must not touch it.
		if _, ok := findRemaining(run.rc, emptyDirAgentDir); ok {
			t.Errorf("the agent directory got a row")
		}
	})

	t.Run("RemainingRowsAreDirectoriesNotFiles", func(t *testing.T) {
		enableRemovalForTest(t)
		root, home := emptyDirFixture(t, true)
		run := runOwnershipMigration(t, root, home)
		var dirRows []remainingRow
		for _, r := range run.rc.remaining {
			if r.Reason == keepEmptyDirNoProvenance {
				dirRows = append(dirRows, r)
			}
		}
		if len(dirRows) != 4 {
			t.Fatalf("%d rows carry %q, want 4 (%s, %s, %s, %s)", len(dirRows), keepEmptyDirNoProvenance,
				emptyDirNoRecord, emptyDirUserRecord, emptyDirMixedParent, emptyDirMirrorNoRecord)
		}
		for _, line := range remainingSummaryLines(dirRows) {
			if strings.Contains(line, "file(s)") {
				t.Errorf("the grouped line calls a directory a file: %q", line)
			}
		}
		for _, line := range remainingDetailLines(dirRows) {
			if !strings.Contains(line, string(keepEmptyDirNoProvenance)) {
				t.Errorf("the verbose row does not carry its reason: %q", line)
			}
		}
	})

	// A wrong implementation is injected through the seam the step judges with;
	// the pure assertion above must go red in front of each one.
	t.Run("MutantPruningAnyEmptyDirIsCaught", func(t *testing.T) {
		enableRemovalForTest(t)
		root, home := emptyDirFixture(t, true)
		before := emptyDirFileSet(root)
		swapEmptyDirProvenance(t, func(emptyDirEvidence, string) emptyDirVerdict {
			return emptyDirVerdict{managedOnly: true}
		})
		run := runOwnershipMigration(t, root, home)
		got := emptyDirProblems(root, run, before, emptyDirPrevRunEmptied, emptyDirStandardExpectations(true))
		if !problemsMention(got, emptyDirNoRecord) {
			t.Errorf("the assertion stayed green in front of a step that removes every empty directory; problems: %v", got)
		}
	})

	t.Run("MutantIgnoringNonManagedRecordsIsCaught", func(t *testing.T) {
		enableRemovalForTest(t)
		root, home := emptyDirFixture(t, true)
		before := emptyDirFileSet(root)
		// Provenance is claimed as soon as ONE managed record exists; a
		// user_created record beside it is not looked at.
		swapEmptyDirProvenance(t, func(ev emptyDirEvidence, dirRel string) emptyDirVerdict {
			prefix := dirRel + "/"
			for rel, prov := range ev.start {
				if strings.HasPrefix(rel, prefix) && prov == manifest.TemplateManaged {
					return emptyDirVerdict{managedOnly: true}
				}
			}
			for _, g := range ev.gens {
				for rel, rec := range g.files {
					if strings.HasPrefix(rel, prefix) && rec.provenance == manifest.TemplateManaged {
						return emptyDirVerdict{managedOnly: true}
					}
				}
			}
			return emptyDirVerdict{class: ownershipUnregistered}
		})
		run := runOwnershipMigration(t, root, home)
		got := emptyDirProblems(root, run, before, emptyDirPrevRunEmptied, emptyDirStandardExpectations(true))
		if !problemsMention(got, emptyDirMixedParent) {
			t.Errorf("the assertion stayed green in front of a step that ignores a user_created record; problems: %v", got)
		}
	})
}

// problemsMention reports whether any problem names rel.
func problemsMention(problems []string, rel string) bool {
	return slices.ContainsFunc(problems, func(p string) bool { return strings.Contains(p, rel+":") })
}

// swapEmptyDirProvenance installs the provenance judgment of the pruning step
// for one test (a mutant injection).
func swapEmptyDirProvenance(t *testing.T, fn func(emptyDirEvidence, string) emptyDirVerdict) {
	t.Helper()
	prev := emptyDirProvenanceFn
	emptyDirProvenanceFn = fn
	t.Cleanup(func() { emptyDirProvenanceFn = prev })
}

// TestEmptyDirProvenanceVerdict pins the pure judgment on its boundaries:
// which records count, how the sources combine, and the trailing-slash prefix.
func TestEmptyDirProvenanceVerdict(t *testing.T) {
	root := t.TempDir()
	rec := func(p manifest.Provenance) priorRecord { return priorRecord{provenance: p} }
	gen := func(files map[string]priorRecord) []priorGeneration {
		return []priorGeneration{{name: "20260101_000001", files: files}}
	}
	ownWriteFile(t, root, ".claude/skills/moai-present/SKILL.md", []byte("still here\n"))

	cases := []struct {
		name  string
		start map[string]manifest.Provenance
		gens  []priorGeneration
		dir   string
		want  emptyDirVerdict
	}{
		{name: "no record anywhere", dir: ".claude/skills/moai-a",
			want: emptyDirVerdict{class: ownershipUnregistered}},
		{name: "managed generation record, file gone", dir: ".claude/skills/moai-a",
			gens: gen(map[string]priorRecord{".claude/skills/moai-a/SKILL.md": rec(manifest.TemplateManaged)}),
			want: emptyDirVerdict{managedOnly: true}},
		{name: "managed start-of-run record, file gone", dir: ".claude/skills/moai-a",
			start: map[string]manifest.Provenance{".claude/skills/moai-a/SKILL.md": manifest.TemplateManaged},
			want:  emptyDirVerdict{managedOnly: true}},
		{name: "same path managed in two sources", dir: ".claude/skills/moai-a",
			start: map[string]manifest.Provenance{".claude/skills/moai-a/SKILL.md": manifest.TemplateManaged},
			gens:  gen(map[string]priorRecord{".claude/skills/moai-a/SKILL.md": rec(manifest.TemplateManaged)}),
			want:  emptyDirVerdict{managedOnly: true}},
		{name: "a user_created record beside managed ones", dir: ".claude/skills/moai-a",
			gens: gen(map[string]priorRecord{
				".claude/skills/moai-a/SKILL.md": rec(manifest.TemplateManaged),
				".claude/skills/moai-a/mine.md":  rec(manifest.UserCreated),
			}),
			want: emptyDirVerdict{class: ownershipUserCreated}},
		{name: "a user_modified record", dir: ".claude/skills/moai-a",
			start: map[string]manifest.Provenance{".claude/skills/moai-a/SKILL.md": manifest.UserModified},
			want:  emptyDirVerdict{class: ownershipUserModified}},
		{name: "a deprecated record is not managed", dir: ".claude/skills/moai-a",
			gens: gen(map[string]priorRecord{".claude/skills/moai-a/SKILL.md": rec(manifest.Deprecated)}),
			want: emptyDirVerdict{class: ownershipUnregistered}},
		{name: "an empty provenance value is not managed", dir: ".claude/skills/moai-a",
			gens: gen(map[string]priorRecord{".claude/skills/moai-a/SKILL.md": rec("")}),
			want: emptyDirVerdict{class: ownershipUnregistered}},
		{name: "a user_created record in the older of two sources still counts", dir: ".claude/skills/moai-a",
			start: map[string]manifest.Provenance{".claude/skills/moai-a/SKILL.md": manifest.TemplateManaged},
			gens:  gen(map[string]priorRecord{".claude/skills/moai-a/old.md": rec(manifest.UserCreated)}),
			want:  emptyDirVerdict{class: ownershipUserCreated}},
		{name: "the prefix is a directory prefix, not a name prefix", dir: ".claude/skills/moai-a",
			gens: gen(map[string]priorRecord{".claude/skills/moai-ab/SKILL.md": rec(manifest.TemplateManaged)}),
			want: emptyDirVerdict{class: ownershipUnregistered}},
		{name: "a recorded path that exists on disk is not gone", dir: ".claude/skills/moai-present",
			gens: gen(map[string]priorRecord{".claude/skills/moai-present/SKILL.md": rec(manifest.TemplateManaged)}),
			want: emptyDirVerdict{class: ownershipUnregistered}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := dirManagedProvenance(emptyDirEvidence{root: root, start: tc.start, gens: tc.gens}, tc.dir)
			if got != tc.want {
				t.Errorf("verdict %+v, want %+v", got, tc.want)
			}
		})
	}
}
