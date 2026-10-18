package cli

// migrate_ownership_recovery_test.go — SPEC-UPDATE-OWNERSHIP-REPAIR-001 M9 (card
// t1865): AC-OWR-17 (a repeated update changes nothing) and AC-OWR-18 (an
// interrupted migration recovers and converges).
//
// Which hash is compared. Every tree comparison in this file is CONTENT-ONLY: the
// relative path, the file bytes, the file mode and the presence of directories.
// Directory mtimes are not part of it, and no clock value is: the manifest's
// `deployed_at` is replaced by a placeholder when two different RUNS are compared
// (debt N17). The run-scoped bookkeeping directories — `.moai-backups`,
// `.moai/backups` (the namespace backup), `.moai/archive/files/update-migration`
// (the legacy archive), `.moai/logs` and `.moai/state` — are left out, because
// each update run adds its own.
//
// Time independence. Run-backup directories are named to the second. No assertion
// here depends on which second a run lands in. Where it could matter, both
// branches are driven and asserted: the DISTINCT-directory branch waits (polls)
// until the clock is past every directory name the previous run wrote, and the
// SHARED-directory branch reproduces the end state of two runs that fell in one
// second (the earlier run's backup content merged into the later run's directory,
// whose own files win) through the priorManifestWriter seam.
//
// Reach (rule 7). Only a run that touched reconciliation counts: not skipped, a
// new manifest.prior.json whose bytes are the manifest the run started from, and a
// positive number of remaining rows.

import (
	"bytes"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Clock branches and the run driver.
// ---------------------------------------------------------------------------

type clockBranch int

const (
	// branchDistinct: the next run starts in a later second than every
	// clock-derived directory of the previous one.
	branchDistinct clockBranch = iota
	// branchShared: the next run's backup directory holds the previous run's
	// content too, as it would if both fell in one second.
	branchShared
)

func (b clockBranch) String() string {
	if b == branchShared {
		return "SharedRunDirectory"
	}
	return "DistinctRunDirectories"
}

// recMergeNoOverwrite copies every file of src into dst, keeping what dst already
// holds.
func recMergeNoOverwrite(src, dst string) error {
	return filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, relErr := filepath.Rel(src, p)
		if relErr != nil {
			return relErr
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if _, statErr := os.Lstat(target); statErr == nil {
			return nil
		}
		info, infoErr := d.Info()
		if infoErr != nil {
			return infoErr
		}
		data, readErr := os.ReadFile(p)
		if readErr != nil {
			return readErr
		}
		return os.WriteFile(target, data, info.Mode().Perm())
	})
}

// recSameDir reports whether two paths name the same directory, whatever their
// spelling (the flow hands the writer a path relative to the project root, a
// test usually holds an absolute one).
func recSameDir(a, b string) bool {
	ra, err1 := filepath.EvalSymlinks(a)
	rb, err2 := filepath.EvalSymlinks(b)
	if err1 != nil || err2 != nil {
		return false
	}
	aa, err1 := filepath.Abs(ra)
	bb, err2 := filepath.Abs(rb)
	return err1 == nil && err2 == nil && aa == bb
}

// withPriorWriter runs fn with the prior-manifest writer replaced and puts the
// previous one back afterwards.
func withPriorWriter(w func(runDir string, data []byte) error, fn func()) {
	prev := priorManifestWriter
	priorManifestWriter = w
	defer func() { priorManifestWriter = prev }()
	fn()
}

// recPrepareNextRun arranges the clock branch for the run that follows prevRunDir
// ("" when there is no previous run directory) and returns the function that runs
// the flow under it. For the shared branch the writer wrapper moves the previous
// run directory's content into the run directory of the new run before the real
// writer replaces the new run's own copy.
func recPrepareNextRun(t *testing.T, root string, br clockBranch, prevRunDir string, run func()) {
	t.Helper()
	switch br {
	case branchDistinct:
		awaitClockPast(t, restoreRunStamps(t, root))
		run()
	case branchShared:
		real := priorManifestWriter
		withPriorWriter(func(runDir string, data []byte) error {
			if prevRunDir != "" && !recSameDir(prevRunDir, runDir) {
				if err := recMergeNoOverwrite(prevRunDir, runDir); err != nil {
					return err
				}
				if err := os.RemoveAll(prevRunDir); err != nil {
					return err
				}
			}
			return real(runDir, data)
		}, run)
	}
}

// ---------------------------------------------------------------------------
// Observations.
// ---------------------------------------------------------------------------

var recDeployedAtRe = regexp.MustCompile(`("deployed_at"\s*:\s*")[^"]*(")`)

// recBookkeeping are the run-scoped directories left out of every comparison.
var recBookkeeping = []string{
	".moai-backups/", ".moai/backups/", ".moai/archive/files/update-migration/", ".moai/logs/", ".moai/state/",
}

// recSnapshot is the content-only snapshot of a tree without the run-scoped
// bookkeeping. With maskDeployedAt the manifest's deployed_at value is replaced,
// for comparing two different runs.
func recSnapshot(t *testing.T, root string, maskDeployedAt bool) map[string]string {
	t.Helper()
	snap := ownTreeSnapshot(t, root, ownTreeOpts{SkipPrefixes: recBookkeeping})
	if maskDeployedAt {
		data, err := os.ReadFile(filepath.Join(root, ".moai", "manifest.json"))
		if err == nil {
			masked := recDeployedAtRe.ReplaceAll(data, []byte(`${1}<TS>${2}`))
			snap[".moai/manifest.json"] = "f masked " + restoreSHA(masked)
		}
	}
	return snap
}

// recCommonFiles is the set of regular files under the four common-asset roots.
func recCommonFiles(root string) map[string]bool {
	out := map[string]bool{}
	for _, rootRel := range projectCommonAssetRels {
		_ = filepath.WalkDir(filepath.Join(root, filepath.FromSlash(rootRel)), func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			if rel, relErr := filepath.Rel(root, p); relErr == nil {
				out[filepath.ToSlash(rel)] = true
			}
			return nil
		})
	}
	return out
}

// recRows reads the remaining list the run recorded in the update ledger: the
// grouped lines and the per-file rows (the latter only under --verbose, which the
// observation turns on).
func recRows() (grouped, rows []string) {
	for _, r := range updateLedger.reference {
		switch {
		case strings.HasPrefix(r.msg, "kept in place:"):
			grouped = append(grouped, r.msg)
		case strings.HasPrefix(r.msg, "kept: "):
			rows = append(rows, r.msg)
		}
	}
	return grouped, rows
}

// recRowRel returns the path a per-file row names.
func recRowRel(row string) string {
	rel, _, _ := strings.Cut(strings.TrimPrefix(row, "kept: "), " — ")
	return rel
}

// recRowReason returns the reason a per-file row carries (the text after the
// class).
func recRowReason(row string) string {
	_, rest, _ := strings.Cut(row, " — ")
	_, reason, _ := strings.Cut(rest, ", ")
	reason, _, _ = strings.Cut(reason, ";")
	return strings.TrimSpace(reason)
}

// recObs is what one update run left behind.
type recObs struct {
	run forcedRun
	// priorBytes are the bytes of the run's new manifest.prior.json, read right
	// after the run (a later run may move or replace the file).
	priorBytes []byte
	snap       map[string]string
	backups    map[string]string // content-only listing of .moai-backups
	manifest   []byte
	grouped    []string
	rows       []string
	files      map[string]bool
}

// recBackupsListing lists .moai-backups: every file's relative path, mode and
// content digest.
func recBackupsListing(root string) map[string]string {
	out := map[string]string{}
	base := filepath.Join(root, ".moai-backups")
	_ = filepath.WalkDir(base, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		rel, relErr := filepath.Rel(base, p)
		if relErr != nil {
			return nil
		}
		data, readErr := os.ReadFile(p)
		if readErr != nil {
			return nil
		}
		out[filepath.ToSlash(rel)] = restoreSHA(data)
		return nil
	})
	return out
}

// recObserve drives one whole update run with the ledger reset and --verbose on
// and records what it left.
func recObserve(t *testing.T, root string, force bool) recObs {
	t.Helper()
	updateLedger.reset()
	updateVerboseMode = true
	run := runUpdateFlow(t, root, force)
	o := recObs{run: run, snap: recSnapshot(t, root, true), files: recCommonFiles(root), backups: recBackupsListing(root)}
	if run.NewPriorManifest != "" {
		o.priorBytes, _ = os.ReadFile(run.NewPriorManifest)
	}
	o.manifest = readManifestBytes(t, root)
	o.grouped, o.rows = recRows()
	return o
}

// ---------------------------------------------------------------------------
// AC-OWR-17.
// ---------------------------------------------------------------------------

// recReached is the rule-7 reach test for one run.
func recReached(o recObs) []string {
	var why []string
	if o.run.Skipped {
		why = append(why, "skipped before reconciliation")
	}
	if o.run.NewPriorManifest == "" {
		why = append(why, "no new manifest.prior.json")
	} else if !bytes.Equal(o.priorBytes, o.run.StartManifest) {
		why = append(why, "manifest.prior.json is not the manifest the run started from")
	}
	if len(o.rows) == 0 {
		why = append(why, "no remaining row")
	}
	return why
}

// recRowKey is the part of a row that names the file and its class (the text up
// to the reason).
func recRowKey(row string) string {
	head, _, _ := strings.Cut(row, ", ")
	return head
}

// recRefined reports whether row b is row a with its reason refined from "no
// provenance record at all" to "a record exists under another key" — the one
// change design.md D-2 allows between runs when a retained generation appears.
func recRefined(a, b string) bool {
	return recRowKey(a) == recRowKey(b) &&
		recRowReason(a) == string(keepNoProvenanceRecord) &&
		recRowReason(b) == string(evidenceIncompleteReason(evidenceOriginalPath))
}

// recRowsProblems compares the remaining rows of two runs. With strict the rows
// must be identical; otherwise they must name the same files with the same
// class and may differ only by the allowed refinement.
func recRowsProblems(label string, a, b recObs, strict bool) []string {
	var problems []string
	if len(a.rows) != len(b.rows) {
		return []string{fmt.Sprintf("(rows) %s: %d rows against %d", label, len(a.rows), len(b.rows))}
	}
	for i := range a.rows {
		ra, rb := a.rows[i], b.rows[i]
		switch {
		case ra == rb:
		case strict:
			problems = append(problems, fmt.Sprintf("(rows) %s: row %d differs: %q vs %q", label, i, ra, rb))
		case !recRefined(ra, rb):
			problems = append(problems, fmt.Sprintf("(rows) %s: row %d differs by more than the allowed refinement: %q vs %q", label, i, ra, rb))
		}
	}
	if strict && !slices.Equal(a.grouped, b.grouped) {
		problems = append(problems, fmt.Sprintf("(rows) %s: the grouped lines differ: %q vs %q", label, a.grouped, b.grouped))
	}
	return problems
}

// repeatedProblems is the whole of AC-OWR-17 as a pure function over runs 1
// (plain), 2 and 3 (forced), a fourth forced run that probes the steady state,
// and two plain re-runs. strictRows says the remaining rows must be identical
// between every pair of runs (no retained generation can refine a reason:
// a shared backup directory, or a project without same-hash copies).
func repeatedProblems(runs [4]recObs, plain [2]recObs, strictRows bool) []string {
	var problems []string
	add := func(format string, args ...any) { problems = append(problems, fmt.Sprintf(format, args...)) }

	// Reach (rule 7): every counted run satisfies all three conditions, and the
	// count of runs that did is exactly the number driven.
	reached := 0
	for i, o := range runs {
		if why := recReached(o); len(why) != 0 {
			add("(reach) run %d: %s", i+1, strings.Join(why, "; "))
		} else {
			reached++
		}
	}
	if reached != len(runs) {
		add("(reach) %d of %d runs reached reconciliation", reached, len(runs))
	}

	// Steady state: consecutive forced runs leave the same tree and the same
	// manifest bytes (outside the run-scoped bookkeeping).
	for i := 1; i < len(runs)-1; i++ {
		if d := ownTreeDiff(runs[i].snap, runs[i+1].snap); len(d) != 0 {
			add("(tree) runs %d and %d left different trees: %v", i+1, i+2, d)
		}
		if !bytes.Equal(runs[i].manifest, runs[i+1].manifest) {
			add("(manifest) the manifest bytes changed between run %d and run %d", i+1, i+2)
		}
	}

	// The remaining list: every run lists exactly the files it left under the four
	// roots, once each, in path order.
	for i, o := range runs {
		rels := make([]string, 0, len(o.rows))
		seen := map[string]bool{}
		for _, row := range o.rows {
			rel := recRowRel(row)
			rels = append(rels, rel)
			if seen[rel] {
				add("(rows) run %d lists %s twice", i+1, rel)
			}
			seen[rel] = true
			if !o.files[rel] {
				add("(rows) run %d lists %s, which is not a file under the common-asset roots", i+1, rel)
			}
		}
		for rel := range o.files {
			if !seen[rel] {
				add("(rows) run %d does not list %s, which it left in place", i+1, rel)
			}
		}
		if !sort.StringsAreSorted(rels) {
			add("(order) run %d lists its rows out of path order", i+1)
		}
	}
	for i := 0; i < len(runs)-1; i++ {
		// The last pair is the steady state and is always compared strictly.
		strict := strictRows || i == len(runs)-2
		problems = append(problems, recRowsProblems(fmt.Sprintf("runs %d and %d", i+1, i+2), runs[i], runs[i+1], strict)...)
	}

	// Plain re-runs: skipped before reconciliation, nothing written, nothing printed.
	last := runs[len(runs)-1]
	for i, p := range plain {
		if !p.run.Skipped {
			add("(plain) plain re-run %d was not skipped", i+1)
		}
		if len(p.rows)+len(p.grouped) != 0 {
			add("(plain) plain re-run %d produced a remaining list", i+1)
		}
		if d := ownTreeDiff(last.snap, p.snap); len(d) != 0 {
			add("(plain) plain re-run %d changed the tree: %v", i+1, d)
		}
		if d := ownTreeDiff(last.backups, p.backups); len(d) != 0 {
			add("(plain) plain re-run %d changed the run backups: %v", i+1, d)
		}
		if !bytes.Equal(last.manifest, p.manifest) {
			add("(plain) plain re-run %d changed the manifest bytes", i+1)
		}
	}
	if plain[0].run.Out != plain[1].run.Out {
		add("(plain) two plain re-runs printed different output")
	}
	return problems
}

// repeatFixture builds one project for AC-OWR-17 and returns it with its home.
type repeatFixture struct {
	name string
	// strictRows: the project holds no unregistered copy whose hash a registered
	// record carries under another key, so no retained generation can refine a
	// reason.
	strictRows bool
	build      func(t *testing.T) (root, home string)
}

func repeatFixtures() []repeatFixture {
	setup := func(t *testing.T) string {
		t.Helper()
		enableRemovalForTest(t)
		ownProfileSandbox(t)
		home := installMigrationUserCounterparts(t)
		t.Setenv("HOME", home)
		t.Setenv("USERPROFILE", home)
		requireUserCounterpart(t, home, migIdenticalSkill)
		prevVerbose := updateVerboseMode
		t.Cleanup(func() { updateVerboseMode = prevVerbose; updateLedger.reset() })
		return home
	}
	return []repeatFixture{
		{"ManyUnregisteredCopies", false, func(t *testing.T) (string, string) {
			home := setup(t)
			root := buildMigrationFixture(t)
			plantUnregisteredMirrorCopies(t, root)
			return root, home
		}},
		{"Mixed", true, func(t *testing.T) (string, string) {
			home := setup(t)
			root := buildMigrationFixture(t)
			for _, rel := range []string{profAgentA, profAgentB} {
				ownAddRegisteredManaged(t, root, rel, ownEmbeddedBytes(t, rel))
			}
			threeManagedCopies(t, root, home)
			ownAddRegisteredUserCreated(t, root, restoreUserFile, []byte("# my own skill\n"))
			for i := range 5 {
				ownAddUnregistered(t, root, fmt.Sprintf(".agents/skills/moai-custom-%d/SKILL.md", i), []byte("# unknown origin\n"))
			}
			return root, home
		}},
	}
}

// observeRepeated drives run 1 (plain), runs 2 to 4 (forced) and two plain
// re-runs under one clock branch.
func observeRepeated(t *testing.T, fx repeatFixture, br clockBranch) (runs [4]recObs, plain [2]recObs) {
	t.Helper()
	root, _ := fx.build(t)
	runs[0] = recObserve(t, root, false)
	for i := 1; i < len(runs); i++ {
		prev := ""
		if runs[i-1].run.NewPriorManifest != "" {
			prev = filepath.Dir(runs[i-1].run.NewPriorManifest)
		}
		recPrepareNextRun(t, root, br, prev, func() { runs[i] = recObserve(t, root, true) })
	}
	plain[0] = recObserve(t, root, false)
	plain[1] = recObserve(t, root, false)
	return runs, plain
}

func TestRepeatedUpdateIsIdempotent(t *testing.T) {
	for _, fx := range repeatFixtures() {
		for _, br := range []clockBranch{branchDistinct, branchShared} {
			t.Run(fx.name+"/"+br.String(), func(t *testing.T) {
				runs, plain := observeRepeated(t, fx, br)
				// Without a retained generation (shared directory) no reason can be
				// refined, so the rows must be identical for every project.
				problems := repeatedProblems(runs, plain, fx.strictRows || br == branchShared)
				for i := 0; i < len(runs)-1; i++ {
					refined := 0
					for j := range runs[i].rows {
						if j < len(runs[i+1].rows) && runs[i].rows[j] != runs[i+1].rows[j] {
							refined++
						}
					}
					t.Logf("remaining rows: run %d lists %d, run %d lists %d; %d row(s) differ between them (reason refinement only)", i+1, len(runs[i].rows), i+2, len(runs[i+1].rows), refined)
				}
				for _, group := range []struct{ name, prefix string }{
					{"ReachedReconciliation", "(reach)"},
					{"ForcedRepeats", "(tree)"},
					{"ManifestBytes", "(manifest)"},
					{"RemainingRows", "(rows)"},
					{"RowOrder", "(order)"},
					{"PlainRepeatSkips", "(plain)"},
				} {
					t.Run(group.name, func(t *testing.T) {
						for _, p := range problemsWithPrefix(problems, group.prefix) {
							t.Error(p)
						}
					})
				}
				if br == branchDistinct && fx.name == "Mixed" {
					recRepeatedMutants(t, runs, plain)
				}
			})
		}
	}
}

// recRepeatedMutants feeds the pure assertion wrong observations and requires a
// non-empty problem list for each.
func recRepeatedMutants(t *testing.T, runs [4]recObs, plain [2]recObs) {
	t.Helper()
	copyObs := func(o recObs) recObs {
		c := o
		c.rows = slices.Clone(o.rows)
		c.grouped = slices.Clone(o.grouped)
		c.manifest = slices.Clone(o.manifest)
		c.snap = maps.Clone(o.snap)
		c.backups = maps.Clone(o.backups)
		return c
	}
	mutant := func(name, wantPrefix string, mutate func(runs *[4]recObs, plain *[2]recObs)) {
		t.Run("Mutant"+name+"IsCaught", func(t *testing.T) {
			var rs [4]recObs
			for i := range rs {
				rs[i] = copyObs(runs[i])
			}
			ps := [2]recObs{copyObs(plain[0]), copyObs(plain[1])}
			mutate(&rs, &ps)
			got := repeatedProblems(rs, ps, true)
			if len(got) == 0 {
				t.Fatalf("the assertion stayed green in front of the %s mutant", name)
			}
			if len(problemsWithPrefix(got, wantPrefix)) == 0 {
				t.Errorf("the %s mutant was caught, but not by clause %s: %v", name, wantPrefix, got)
			}
		})
	}
	mutant("DeployedAtRefreshedEveryRun", "(manifest)", func(rs *[4]recObs, _ *[2]recObs) {
		rs[2].manifest = recDeployedAtRe.ReplaceAll(rs[2].manifest, []byte(`${1}2099-01-01T00:00:00Z${2}`))
	})
	mutant("RowsInMapOrder", "(order)", func(rs *[4]recObs, _ *[2]recObs) {
		slices.Reverse(rs[2].rows)
	})
	mutant("ARowIsMissing", "(rows)", func(rs *[4]recObs, _ *[2]recObs) {
		rs[2].rows = rs[2].rows[1:]
	})
	mutant("ARowChangesItsReason", "(rows)", func(rs *[4]recObs, _ *[2]recObs) {
		rs[2].rows[0] = strings.Replace(rs[2].rows[0], ", ", ", changed-", 1)
	})
	mutant("RunTwoSkipped", "(reach)", func(rs *[4]recObs, _ *[2]recObs) {
		rs[1].run.Skipped = true
	})
	mutant("RunThreeWroteNoPriorManifest", "(reach)", func(rs *[4]recObs, _ *[2]recObs) {
		rs[2].run.NewPriorManifest = ""
	})
	mutant("RunThreePriorManifestIsNotTheStartManifest", "(reach)", func(rs *[4]recObs, _ *[2]recObs) {
		rs[2].priorBytes = []byte("{}")
	})
	mutant("TreeChangesInRunThree", "(tree)", func(rs *[4]recObs, _ *[2]recObs) {
		rs[2].snap["a/new/file"] = "f 644 x"
	})
	mutant("PlainRepeatReconciles", "(plain)", func(_ *[4]recObs, ps *[2]recObs) {
		ps[0].run.Skipped = false
	})
	mutant("PlainRepeatRewritesTheTree", "(plain)", func(_ *[4]recObs, ps *[2]recObs) {
		ps[1].snap[".claude/skills/rewritten/SKILL.md"] = "f 644 y"
	})
	mutant("PlainRepeatWritesARunBackup", "(plain)", func(_ *[4]recObs, ps *[2]recObs) {
		ps[1].backups["20991231_000000/manifest.prior.json"] = "x"
	})
	mutant("PlainRepeatsPrintDifferentOutput", "(plain)", func(_ *[4]recObs, ps *[2]recObs) {
		ps[1].run.Out += "extra"
	})
}
