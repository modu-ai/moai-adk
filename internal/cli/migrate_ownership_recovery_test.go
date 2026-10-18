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
// here depends on which second a run lands in. Two branches are driven and
// asserted identically: the DISTINCT-directory branch waits (polls) until the
// clock is past every directory name the previous run wrote, and the NO-WAIT
// branch starts the next run at once, so the two runs share a clock second or not
// depending on the machine. The outcome must be the same either way: each run's
// start-of-run manifest lands in a directory no other run writes to (sync-audit
// finding F2), so both prior generations survive. The same-second case itself is
// pinned deterministically, without a clock, in update_prior_manifest_collision_test.go.
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
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/manifest"
)

// ---------------------------------------------------------------------------
// Clock branches and the run driver.
// ---------------------------------------------------------------------------

type clockBranch int

const (
	// branchDistinct: the next run starts in a later second than every
	// clock-derived directory of the previous one.
	branchDistinct clockBranch = iota
	// branchNoWait: the next run starts at once. Whether it shares a second with
	// the previous run is up to the machine, and the assertions do not care.
	branchNoWait
)

func (b clockBranch) String() string {
	if b == branchNoWait {
		return "NoWaitConsecutiveRuns"
	}
	return "DistinctRunDirectories"
}

// withPriorWriter runs fn with the prior-manifest writer replaced and puts the
// previous one back afterwards.
func withPriorWriter(w func(runDir string, data []byte) error, fn func()) {
	prev := priorManifestWriter
	priorManifestWriter = w
	defer func() { priorManifestWriter = prev }()
	fn()
}

// recPrepareNextRun arranges the clock branch for the next run and runs it: the
// distinct branch first waits until the clock is past every run-scoped directory
// name, the no-wait branch runs at once.
func recPrepareNextRun(t *testing.T, root string, br clockBranch, run func()) {
	t.Helper()
	if br == branchDistinct {
		awaitClockPast(t, restoreRunStamps(t, root))
	}
	run()
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
		recPrepareNextRun(t, root, br, func() { runs[i] = recObserve(t, root, true) })
	}
	plain[0] = recObserve(t, root, false)
	plain[1] = recObserve(t, root, false)
	return runs, plain
}

func TestRepeatedUpdateIsIdempotent(t *testing.T) {
	t.Log("compared tree hash: content-only (path, bytes, mode, directory presence); directory mtimes and clock values are not compared; deployed_at is masked between runs")
	for _, fx := range repeatFixtures() {
		for _, br := range []clockBranch{branchDistinct, branchNoWait} {
			t.Run(fx.name+"/"+br.String(), func(t *testing.T) {
				runs, plain := observeRepeated(t, fx, br)
				problems := repeatedProblems(runs, plain, fx.strictRows)
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

// ---------------------------------------------------------------------------
// AC-OWR-18.
// ---------------------------------------------------------------------------

// convWorld is the fixture project's removal candidates and their pre-run bytes.
type convWorld struct {
	cands []string
	pre   map[string][]byte
}

// convFixture builds the pristine project (never run) with the removal switch on
// inside this test, and returns it with its candidates.
func convFixture(t *testing.T) (root string, w convWorld) {
	t.Helper()
	enableRemovalForTest(t)
	ownProfileSandbox(t)
	home := installMigrationUserCounterparts(t)
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	prevVerbose := updateVerboseMode
	t.Cleanup(func() { updateVerboseMode = prevVerbose; updateLedger.reset() })

	root = buildMigrationFixture(t)
	cands := threeManagedCopies(t, root, home)
	slices.Sort(cands)
	ownAddRegisteredUserCreated(t, root, restoreUserFile, []byte("# my own skill\n"))
	for i := range 3 {
		ownAddUnregistered(t, root, fmt.Sprintf(".agents/skills/moai-custom-%d/SKILL.md", i), []byte("# unknown origin\n"))
	}
	w = convWorld{cands: cands, pre: map[string][]byte{}}
	for _, rel := range cands {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatalf("read %s: %v", rel, err)
		}
		w.pre[rel] = data
	}
	return root, w
}

// convCopy returns an independent byte copy of the pristine project.
func convCopy(t *testing.T, pristine string) string {
	t.Helper()
	root := t.TempDir()
	restoreCopyTree(t, pristine, root)
	return root
}

// convEmptyDirs lists the empty directories under the four common-asset roots
// (the roots themselves excepted), as slash paths.
func convEmptyDirs(root string) []string {
	var out []string
	for _, rootRel := range projectCommonAssetRels {
		base := filepath.Join(root, filepath.FromSlash(rootRel))
		_ = filepath.WalkDir(base, func(p string, d fs.DirEntry, err error) error {
			if err != nil || !d.IsDir() || p == base {
				return nil
			}
			if ents, readErr := os.ReadDir(p); readErr == nil && len(ents) == 0 {
				rel, _ := filepath.Rel(root, p)
				out = append(out, filepath.ToSlash(rel))
			}
			return nil
		})
	}
	sort.Strings(out)
	return out
}

// convBackupCopy returns the bytes of the run-backup copy of rel, or nil when no
// run backup holds a complete (regular-file) copy.
func convBackupCopy(root, rel string) []byte {
	matches, _ := filepath.Glob(filepath.Join(root, ".moai-backups", "*", migrationBackupSubdir, filepath.FromSlash(rel)))
	for _, m := range matches {
		if info, err := os.Stat(m); err == nil && info.Mode().IsRegular() {
			if data, readErr := os.ReadFile(m); readErr == nil {
				return data
			}
		}
	}
	return nil
}

// convInterruptedProblems is clause (a) before the recovery: the interrupted run
// removed no file (REQ-OWR-016), kept every candidate byte for byte, listed each
// with the reason backup-failed, left its manifest entries alone, and left no
// complete backup copy of a file it did not remove.
func convInterruptedProblems(w convWorld, a recObs, root string) []string {
	var problems []string
	add := func(format string, args ...any) { problems = append(problems, fmt.Sprintf(format, args...)) }
	if why := recReached(a); len(why) != 0 {
		add("(reach) the interrupted run: %s", strings.Join(why, "; "))
	}
	mf := manifestKeysOf(root)
	for _, rel := range w.cands {
		got, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil || !bytes.Equal(got, w.pre[rel]) {
			add("(interrupted) %s was removed or changed by the interrupted run (%v)", rel, err)
		}
		if !mf[rel] {
			add("(interrupted) the manifest lost its entry for %s", rel)
		}
		row := ""
		for _, r := range a.rows {
			if recRowRel(r) == rel {
				row = r
			}
		}
		if row == "" || recRowReason(row) != string(keepBackupFailed) {
			add("(interrupted) %s has no row with reason %s (row %q)", rel, keepBackupFailed, row)
		}
		if convBackupCopy(root, rel) != nil {
			add("(interrupted) a complete backup copy of %s exists although nothing was removed", rel)
		}
	}
	return problems
}

// manifestKeysOf returns the set of keys in the project manifest.
func manifestKeysOf(root string) map[string]bool {
	out := map[string]bool{}
	mgr := manifest.NewManager()
	mf, err := mgr.Load(root)
	if err != nil || mf == nil {
		return out
	}
	for k := range mf.Files {
		out[k] = true
	}
	return out
}

// convExpect says how the final state may differ from the uninterrupted
// reference: keptDirs are emptied directories the final state keeps and lists.
type convExpect struct {
	keptDirs []string
}

// convFinalProblems is clauses (reach), (tree), (rows), (backup), (manifest) and
// (emptydirs): the state after recovery equals the uninterrupted reference —
// except for expect.keptDirs, which must be exactly the extra directories and
// their extra rows.
func convFinalProblems(w convWorld, ref, fin recObs, root string, expect convExpect) []string {
	var problems []string
	add := func(format string, args ...any) { problems = append(problems, fmt.Sprintf(format, args...)) }
	if why := recReached(fin); len(why) != 0 {
		add("(reach) the recovery run: %s", strings.Join(why, "; "))
	}
	// Tree: equal to the reference, plus exactly the kept directories.
	var wantDiff []string
	for _, d := range expect.keptDirs {
		wantDiff = append(wantDiff, d+"/")
	}
	sort.Strings(wantDiff)
	if got := ownTreeDiff(ref.snap, fin.snap); !slices.Equal(got, wantDiff) {
		add("(tree) the recovered tree differs from the uninterrupted one at %v, want %v", got, wantDiff)
	}
	// Empty managed directories: only the expected ones remain.
	wantEmpty := slices.Clone(expect.keptDirs)
	sort.Strings(wantEmpty)
	if got := convEmptyDirs(root); !slices.Equal(got, wantEmpty) {
		add("(emptydirs) empty directories left under the common-asset roots: %v, want %v", got, wantEmpty)
	}
	// Rows: the reference's rows, plus one empty-dir row per kept directory.
	refRows := map[string]string{}
	for _, r := range ref.rows {
		refRows[recRowRel(r)] = r
	}
	extra := map[string]string{}
	for _, r := range fin.rows {
		rel := recRowRel(r)
		if want, ok := refRows[rel]; ok {
			if r != want && !recRefined(want, r) {
				add("(rows) %s: %q against the reference %q", rel, r, want)
			}
			delete(refRows, rel)
			continue
		}
		extra[rel] = r
	}
	for rel := range refRows {
		add("(rows) the recovered run does not list %s, which the uninterrupted run lists", rel)
	}
	for _, d := range expect.keptDirs {
		r, ok := extra[d]
		if !ok || recRowReason(r) != string(keepEmptyDirNoProvenance) {
			add("(rows) kept directory %s is not listed with reason %s (row %q)", d, keepEmptyDirNoProvenance, r)
		}
		delete(extra, d)
	}
	for rel, r := range extra {
		add("(rows) the recovered run lists %s, which the uninterrupted run does not: %q", rel, r)
	}
	// Removed files: gone, backed up byte for byte, and not in the manifest.
	mf := manifestKeysOf(root)
	for _, rel := range w.cands {
		if _, err := os.Lstat(filepath.Join(root, filepath.FromSlash(rel))); !os.IsNotExist(err) {
			add("(removed) %s is still on disk (%v)", rel, err)
		}
		if got := convBackupCopy(root, rel); !bytes.Equal(got, w.pre[rel]) {
			add("(backup) no run backup holds a byte-identical copy of %s", rel)
		}
		if mf[rel] {
			add("(manifest) the manifest entry for %s came back", rel)
		}
	}
	return problems
}

// convRecoveryCase names one interruption and drives it under a clock branch. It
// returns the recovered root, the final observation, and the interrupted
// observation where the case has one.
type convRecoveryCase struct {
	name string
	run  func(t *testing.T, pristine string, w convWorld, br clockBranch) (root string, fin recObs, expect convExpect)
}

func convRunDir(o recObs) string {
	if o.run.NewPriorManifest == "" {
		return ""
	}
	return filepath.Dir(o.run.NewPriorManifest)
}

func convCases() []convRecoveryCase {
	return []convRecoveryCase{
		{"BackupPathOccupied", func(t *testing.T, pristine string, w convWorld, br clockBranch) (string, recObs, convExpect) {
			root := convCopy(t, pristine)
			real := priorManifestWriter
			var a recObs
			withPriorWriter(func(runDir string, data []byte) error {
				// A directory with content stands where one candidate's byte copy
				// must be written, so the copy cannot be made (root included).
				blocker := filepath.Join(runDir, migrationBackupSubdir, filepath.FromSlash(w.cands[1]), "occupant")
				if err := os.MkdirAll(blocker, 0o755); err != nil {
					return err
				}
				return real(runDir, data)
			}, func() { a = recObserve(t, root, true) })
			for _, p := range convInterruptedProblems(w, a, root) {
				t.Error(p)
			}
			runDirA := convRunDir(a)
			if err := os.RemoveAll(filepath.Join(runDirA, migrationBackupSubdir, filepath.FromSlash(w.cands[1]))); err != nil {
				t.Fatalf("free the occupied path: %v", err)
			}
			var fin recObs
			recPrepareNextRun(t, root, br, func() { fin = recObserve(t, root, true) })
			return root, fin, convExpect{}
		}},
		{"ManifestRestoredToPreRunBytes", func(t *testing.T, pristine string, w convWorld, br clockBranch) (string, recObs, convExpect) {
			root := convCopy(t, pristine)
			pre := readManifestBytes(t, root)
			first := recObserve(t, root, true)
			if why := recReached(first); len(why) != 0 {
				t.Fatalf("the first run did not reach reconciliation: %v", why)
			}
			if err := os.WriteFile(filepath.Join(root, ".moai", "manifest.json"), pre, 0o600); err != nil {
				t.Fatalf("restore the manifest bytes: %v", err)
			}
			mf := manifestKeysOf(root)
			for _, rel := range w.cands {
				if !mf[rel] {
					t.Fatalf("premise: the restored manifest does not carry %s (the interruption would resurrect nothing)", rel)
				}
				if _, err := os.Lstat(filepath.Join(root, filepath.FromSlash(rel))); !os.IsNotExist(err) {
					t.Fatalf("premise: %s was not removed by the first run", rel)
				}
			}
			var fin recObs
			recPrepareNextRun(t, root, br, func() { fin = recObserve(t, root, true) })
			return root, fin, convExpect{}
		}},
		{"EmptiedDirectoriesRecreated", func(t *testing.T, pristine string, w convWorld, br clockBranch) (string, recObs, convExpect) {
			root := convCopy(t, pristine)
			first := recObserve(t, root, true)
			if why := recReached(first); len(why) != 0 {
				t.Fatalf("the first run did not reach reconciliation: %v", why)
			}
			var recreated []string
			for _, rel := range w.cands {
				dir := path.Dir(rel)
				if _, err := os.Lstat(filepath.Join(root, filepath.FromSlash(dir))); os.IsNotExist(err) {
					if mkErr := os.MkdirAll(filepath.Join(root, filepath.FromSlash(dir)), 0o755); mkErr != nil {
						t.Fatal(mkErr)
					}
					recreated = append(recreated, dir)
				}
			}
			sort.Strings(recreated)
			if len(recreated) != len(w.cands) {
				t.Fatalf("premise: the first run pruned %d of the %d emptied directories (%v)", len(recreated), len(w.cands), recreated)
			}
			var fin recObs
			recPrepareNextRun(t, root, br, func() { fin = recObserve(t, root, true) })
			return root, fin, convExpect{}
		}},
	}
}

func TestInterruptedMigrationRecoversAndConverges(t *testing.T) {
	t.Log("compared tree hash: content-only (path, bytes, mode, directory presence); directory mtimes and clock values are not compared; deployed_at is masked between runs")
	pristine, w := convFixture(t)

	// The uninterrupted reference: the same NUMBER of runs as the interrupted
	// case (a run, then a second run), none of them interrupted. The first run of
	// a project deploys and the second merges, so the trees after one and after two
	// runs differ in the config sections; only equal run counts compare.
	refRoot := convCopy(t, pristine)
	ref1 := recObserve(t, refRoot, true)
	if why := recReached(ref1); len(why) != 0 {
		t.Fatalf("the first reference run did not reach reconciliation: %v", why)
	}
	if got := restoreRemovedCandidates(refRoot, w.cands); !slices.Equal(got, w.cands) {
		t.Fatalf("the reference run removed %v, want %v", got, w.cands)
	}
	awaitClockPast(t, restoreRunStamps(t, refRoot))
	ref := recObserve(t, refRoot, true)
	if why := recReached(ref); len(why) != 0 {
		t.Fatalf("the second reference run did not reach reconciliation: %v", why)
	}

	keptDir := t.TempDir()
	var kept struct {
		root string
		fin  recObs
	}
	for _, c := range convCases() {
		for _, br := range []clockBranch{branchDistinct, branchNoWait} {
			t.Run(c.name+"/"+br.String(), func(t *testing.T) {
				root, fin, expect := c.run(t, pristine, w, br)
				for _, p := range convFinalProblems(w, ref, fin, root, expect) {
					t.Error(p)
				}
				if c.name == "ManifestRestoredToPreRunBytes" && br == branchDistinct {
					restoreCopyTree(t, root, keptDir)
					kept.root, kept.fin = keptDir, fin
				}
			})
		}
	}

	// The plain re-run: a same-version run the gate skips leaves the recreated
	// directories alone (plan.md OQ-6) — it is not a repair path.
	t.Run("PlainRerunLeavesRecreatedDirectories", func(t *testing.T) {
		root := convCopy(t, pristine)
		first := recObserve(t, root, true)
		if why := recReached(first); len(why) != 0 {
			t.Fatalf("the first run did not reach reconciliation: %v", why)
		}
		for _, rel := range w.cands {
			if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(path.Dir(rel))), 0o755); err != nil {
				t.Fatal(err)
			}
		}
		before := recSnapshot(t, root, true)
		plain := recObserve(t, root, false)
		if !plain.run.Skipped {
			t.Fatalf("the plain re-run was not skipped")
		}
		if d := ownTreeDiff(before, plain.snap); len(d) != 0 {
			t.Errorf("the skipped run changed the tree: %v", d)
		}
		if got := convEmptyDirs(root); len(got) != len(w.cands) {
			t.Errorf("the skipped run changed the recreated directories: %v", got)
		}
	})

	if kept.root == "" {
		t.Fatalf("no recovered project to mutate")
	}
	convMutants(t, w, ref, kept.fin, kept.root)
}

// convMutants feeds the pure assertions wrong end states and requires a
// non-empty problem list for each.
func convMutants(t *testing.T, w convWorld, ref, fin recObs, recovered string) {
	t.Helper()
	mutant := func(name, wantPrefix string, mutate func(root string, fin *recObs)) {
		t.Run("Mutant"+name+"IsCaught", func(t *testing.T) {
			root := convCopy(t, recovered)
			f := fin
			f.rows = slices.Clone(fin.rows)
			f.snap = maps.Clone(fin.snap)
			mutate(root, &f)
			got := convFinalProblems(w, ref, f, root, convExpect{})
			if len(got) == 0 {
				t.Fatalf("the assertion stayed green in front of the %s mutant", name)
			}
			if len(problemsWithPrefix(got, wantPrefix)) == 0 {
				t.Errorf("the %s mutant was caught, but not by clause %s: %v", name, wantPrefix, got)
			}
		})
	}
	mutant("EmptiedManagedDirectoryLeft", "(emptydirs)", func(root string, f *recObs) {
		dir := path.Dir(w.cands[0])
		if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(dir)), 0o755); err != nil {
			t.Fatal(err)
		}
		f.snap[dir+"/"] = "d 755"
	})
	mutant("RemovedManifestEntryResurrected", "(manifest)", func(root string, _ *recObs) {
		rewriteManifestNonCanonical(t, root, func(files map[string]map[string]any, _ map[string]any) {
			files[w.cands[0]] = map[string]any{"provenance": "template_managed", "template_hash": "sha256:x", "deployed_hash": "sha256:x", "current_hash": "sha256:x"}
		})
	})
	mutant("BackupCopyMissing", "(backup)", func(root string, _ *recObs) {
		matches, _ := filepath.Glob(filepath.Join(root, ".moai-backups", "*", migrationBackupSubdir, filepath.FromSlash(w.cands[1])))
		if len(matches) == 0 {
			t.Fatalf("no backup copy of %s to remove", w.cands[1])
		}
		for _, m := range matches {
			_ = os.Remove(m)
		}
	})
	mutant("RemovedFileStillOnDisk", "(removed)", func(root string, f *recObs) {
		ownWriteFile(t, root, w.cands[2], w.pre[w.cands[2]])
		f.snap[w.cands[2]] = "f 644 x"
	})
	mutant("TreeDiffersFromTheUninterruptedRun", "(tree)", func(_ string, f *recObs) {
		f.snap[".claude/skills/other/SKILL.md"] = "f 644 y"
	})
	mutant("RowsDifferFromTheUninterruptedRun", "(rows)", func(_ string, f *recObs) {
		f.rows[0] = strings.Replace(f.rows[0], ", ", ", changed-", 1)
	})
	mutant("RecoveryRunDidNotReachReconciliation", "(reach)", func(_ string, f *recObs) {
		f.run.Skipped = true
	})
	t.Run("MutantInterruptedRunRemovedAFileIsCaught", func(t *testing.T) {
		root := convCopy(t, recovered) // the candidates are already gone here
		got := convInterruptedProblems(w, ref, root)
		if len(problemsWithPrefix(got, "(interrupted)")) == 0 {
			t.Errorf("a run that removed the candidates passed the interrupted-run clause: %v", got)
		}
	})
}
