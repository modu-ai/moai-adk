package cli

// migrate_ownership_restore_test.go — SPEC-UPDATE-OWNERSHIP-REPAIR-001 M8 (card
// t1865), AC-OWR-15: a file the migration removed can be put back byte for byte,
// together with the manifest that described it.
//
// Shape of the evidence (acceptance.md AC-OWR-15):
//
//   - Run 1 is the whole update flow over a fixture with the removal switch ON
//     (the switch is turned on inside this fixture only). Its start manifest has a
//     non-canonical shape (four-space indent, a trailing blank line), so a copy
//     made by re-serialising the manifest cannot equal a copy made from the
//     original bytes.
//   - The control run T0 is the same flow over an independent byte copy of the
//     fixture with the switch OFF. It carries every rewrite and deployment the
//     flow does apart from removals.
//   - The restore is a test-only procedure, not a product command: the backed-up
//     copies go back in place and manifest.prior.json goes over the manifest.
//   - Restore area R = {every removed file, the manifest file}: its files must be
//     byte-identical to their pre-run bytes. The rest of the tree is judged by a
//     difference sentinel: the paths where the restored tree T2 and the control
//     tree T0 differ (D) must be non-empty and lie inside the expected set E.
//     A whole-tree equality with the pre-run tree is NOT asked for: the flow
//     rewrites llm.yaml and deploys templates, so no correct implementation could
//     pass it.
//
// Which hash is compared: the tree comparison here is CONTENT-ONLY (relative path,
// file bytes, file mode, directory presence). Directory mtimes and the clock are
// not part of it (debt N17), and no assertion depends on which clock second a run
// lands in: the two runs live in different project roots, and run-backup
// directory names are reduced to a placeholder before their layouts are compared.

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// restoreUserFile is a registered user_created file inside a common-asset root;
// the "also deletes the user's file" mutant is judged against it.
const restoreUserFile = ".claude/skills/mine/SKILL.md"

// ---------------------------------------------------------------------------
// Tree copy and the test-only restore procedure.
// ---------------------------------------------------------------------------

// restoreCopyTree copies the tree at src to dst, keeping file modes. The fixtures
// hold regular files and directories only.
func restoreCopyTree(t *testing.T, src, dst string) {
	t.Helper()
	err := filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, relErr := filepath.Rel(src, p)
		if relErr != nil {
			return relErr
		}
		target := filepath.Join(dst, rel)
		info, infoErr := d.Info()
		if infoErr != nil {
			return infoErr
		}
		switch {
		case d.IsDir():
			return os.MkdirAll(target, info.Mode().Perm())
		case info.Mode().IsRegular():
			data, readErr := os.ReadFile(p)
			if readErr != nil {
				return readErr
			}
			return os.WriteFile(target, data, info.Mode().Perm())
		default:
			return fmt.Errorf("unexpected non-regular entry %s (%s)", p, info.Mode())
		}
	})
	if err != nil {
		t.Fatalf("copy %s -> %s: %v", src, dst, err)
	}
}

// restoreFromRunBackup puts the run backup back: every file under
// <runDir>/removed-assets returns to its project path (with the mode its backup
// copy carries), and <runDir>/manifest.prior.json goes over the manifest file.
// It is a test procedure, not a product command.
func restoreFromRunBackup(t *testing.T, root, runDir string) {
	t.Helper()
	assets := filepath.Join(runDir, migrationBackupSubdir)
	err := filepath.WalkDir(assets, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, relErr := filepath.Rel(assets, p)
		if relErr != nil {
			return relErr
		}
		info, infoErr := d.Info()
		if infoErr != nil {
			return infoErr
		}
		data, readErr := os.ReadFile(p)
		if readErr != nil {
			return readErr
		}
		dst := filepath.Join(root, rel)
		if mkErr := os.MkdirAll(filepath.Dir(dst), 0o755); mkErr != nil {
			return mkErr
		}
		return os.WriteFile(dst, data, info.Mode().Perm())
	})
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("restore removed assets from %s: %v", assets, err)
	}
	prior, err := os.ReadFile(filepath.Join(runDir, priorManifestName))
	if err != nil {
		t.Fatalf("read %s for the restore: %v", priorManifestName, err)
	}
	if err := os.WriteFile(filepath.Join(root, ".moai", "manifest.json"), prior, 0o644); err != nil {
		t.Fatalf("restore the manifest: %v", err)
	}
}

// ---------------------------------------------------------------------------
// The expected-difference set E and the layout of a run backup.
// ---------------------------------------------------------------------------

// restoreInE reports whether a snapshot key (a slash path; directories end with
// "/") belongs to the expected-difference set E: the manifest file and the
// run-backup, log and state trees.
func restoreInE(key string) bool {
	switch {
	case key == ".moai/manifest.json":
		return true
	case key == ".moai-backups/" || strings.HasPrefix(key, ".moai-backups/"):
		return true
	case key == ".moai/logs/" || strings.HasPrefix(key, ".moai/logs/"):
		return true
	case key == ".moai/state/" || strings.HasPrefix(key, ".moai/state/"):
		return true
	}
	return false
}

// restoreRunLayout lists the files of one run backup directory, relative to it.
// The directory's own name (a timestamp) never appears in the result, so two
// layouts from different clock seconds compare equal.
func restoreRunLayout(runDir string) []string {
	var out []string
	_ = filepath.WalkDir(runDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		rel, relErr := filepath.Rel(runDir, p)
		if relErr == nil {
			out = append(out, filepath.ToSlash(rel))
		}
		return nil
	})
	sort.Strings(out)
	return out
}

func restoreSHA(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// The clock-derived run locations outside `.moai-backups`, found by running the
// whole flow twice: the namespace backup directory `.moai/backups/update-<UTC
// stamp>` (with a `.moai-run` marker carrying a nanosecond run id and a
// `.complete` marker carrying a time) and the legacy-archive directory
// `.moai/archive/files/update-migration/<YYYYMMDD_HHMMSS>`.
var (
	restoreNSBackupRe = regexp.MustCompile(`^\.moai/backups/update-[^/]+(/|$)`)
	restoreArchiveRe  = regexp.MustCompile(`^\.moai/archive/files/update-migration/\d{8}_\d{6}(/|$)`)
)

// restoreNormalize reduces the snapshot of a tree to a clock-independent form,
// erasing exactly the run-derived text of those two locations: the directory
// names' timestamps (replaced by <RUN>) and the content digests of the namespace
// backup's two markers (replaced by <RUN-ID>, mode kept). Everything else —
// every other key and every other value — is returned unchanged. The second
// result lists keys that collided after the rename (two such directories in one
// tree), which the caller must treat as a problem.
func restoreNormalize(snap map[string]string) (map[string]string, []string) {
	out := make(map[string]string, len(snap))
	var collisions []string
	for k, v := range snap {
		nk := restoreNSBackupRe.ReplaceAllString(k, ".moai/backups/update-<RUN>$1")
		nk = restoreArchiveRe.ReplaceAllString(nk, ".moai/archive/files/update-migration/<RUN>$1")
		if strings.HasPrefix(nk, ".moai/backups/update-<RUN>/") && (strings.HasSuffix(nk, "/.moai-run") || strings.HasSuffix(nk, "/.complete")) {
			if fields := strings.Fields(v); len(fields) == 3 {
				v = fields[0] + " " + fields[1] + " <RUN-ID>"
			}
		}
		if prev, dup := out[nk]; dup && prev != v {
			collisions = append(collisions, nk)
		}
		out[nk] = v
	}
	sort.Strings(collisions)
	return out, collisions
}

// restoreRunStamps returns the second each clock-derived run directory under root
// was created in: the `.moai-backups/<YYYYMMDD_HHMMSS>` run directories, the
// `.moai/backups/update-<UTC stamp>` namespace directories and the
// `.moai/archive/files/update-migration/<YYYYMMDD_HHMMSS>` archive directories.
func restoreRunStamps(t *testing.T, root string) []time.Time {
	t.Helper()
	var stamps []time.Time
	runs, _ := filepath.Glob(filepath.Join(root, ".moai-backups", "*"))
	for _, d := range runs {
		stamp, err := time.ParseInLocation(defs.BackupTimestampFormat, filepath.Base(d), time.Local)
		if err != nil {
			t.Fatalf("run backup directory %s does not carry a timestamp name: %v", d, err)
		}
		stamps = append(stamps, stamp)
	}
	nss, _ := filepath.Glob(filepath.Join(root, ".moai", "backups", "update-*"))
	for _, d := range nss {
		stamp, err := time.Parse("2006-01-02T15-04-05Z", strings.TrimPrefix(filepath.Base(d), "update-"))
		if err != nil {
			t.Fatalf("namespace backup directory %s does not carry a UTC stamp name: %v", d, err)
		}
		stamps = append(stamps, stamp)
	}
	archives, _ := filepath.Glob(filepath.Join(root, ".moai", "archive", "files", "update-migration", "*"))
	for _, d := range archives {
		stamp, err := time.ParseInLocation(defs.BackupTimestampFormat, filepath.Base(d), time.Local)
		if err != nil {
			t.Fatalf("archive directory %s does not carry a timestamp name: %v", d, err)
		}
		stamps = append(stamps, stamp)
	}
	return stamps
}

// awaitClockPast blocks until the wall clock is in a later second than every
// stamp, so a run that starts afterwards cannot reuse any of those directory
// names. It polls instead of sleeping a fixed time and fails when the clock does
// not advance.
func awaitClockPast(t *testing.T, stamps []time.Time) {
	t.Helper()
	var latest time.Time
	for _, s := range stamps {
		if s.After(latest) {
			latest = s
		}
	}
	deadline := time.Now().Add(6 * time.Second)
	for !time.Now().Truncate(time.Second).After(latest) {
		if time.Now().After(deadline) {
			t.Fatalf("the clock did not move past %s", latest.Format(time.RFC3339))
		}
		time.Sleep(25 * time.Millisecond)
	}
}

// ---------------------------------------------------------------------------
// The observation and the pure assertion.
// ---------------------------------------------------------------------------

// restoreWorld is everything the criterion reads: the fixture's pre-run state, the
// two runs' reach observations, and where the trees are.
type restoreWorld struct {
	root1, root0       string // T1 (switch on) and the control T0 (switch off)
	rels               []string
	pre                map[string][]byte
	preMode            map[string]os.FileMode
	preManifest        []byte
	preManifestMode    os.FileMode
	run1, run0         forcedRun
	runDir1, runDir0   string
	removedAfterRun1   []string // the removal candidates absent after run 1
	removedAfterRun0   []string // the same for the control (must stay empty)
	postRun1Snapshot   map[string]string
	postRun1AssetFiles map[string]string // rel -> sha of every file under removed-assets (run 1)
}

// restoreProblems is the whole of AC-OWR-15 as a pure function over a world and
// the restored tree root2 (T2). It returns one line per violated clause, each
// prefixed with the clause it belongs to. The mutant sub-tests feed it mutated
// copies and require a non-empty list.
func restoreProblems(t *testing.T, w restoreWorld, root2, runDir2 string) []string {
	t.Helper()
	var problems []string
	add := func(format string, args ...any) { problems = append(problems, fmt.Sprintf(format, args...)) }

	// (v) Reach: run 1 touched reconciliation, produced its manifest copy and
	// removed files; the control run reached the same stage and removed nothing.
	if w.run1.Skipped {
		add("(v) run 1 was skipped before reconciliation")
	}
	if w.run1.NewPriorManifest == "" {
		add("(v) run 1 left no new %s: the Backup stage was not reached", priorManifestName)
	}
	if len(w.removedAfterRun1) == 0 {
		add("(v) run 1 removed no file: the restore below would prove nothing")
	}
	if !slices.Equal(w.removedAfterRun1, w.rels) {
		add("(v) run 1 removed %v, the fixture's removal candidates are %v", w.removedAfterRun1, w.rels)
	}
	if w.run0.Skipped || w.run0.NewPriorManifest == "" {
		add("(v) the control run did not reach reconciliation (skipped=%v, prior copy=%q)", w.run0.Skipped, w.run0.NewPriorManifest)
	}
	if len(w.removedAfterRun0) != 0 {
		add("(v) the control run (switch off) removed %v", w.removedAfterRun0)
	}

	// (i) Each removed file's backup copy has the bytes the file had before the run.
	assets := filepath.Join(runDir2, migrationBackupSubdir)
	for _, rel := range w.rels {
		got, err := os.ReadFile(filepath.Join(assets, filepath.FromSlash(rel)))
		if err != nil {
			add("(i) no backup copy of %s: %v", rel, err)
			continue
		}
		if restoreSHA(got) != restoreSHA(w.pre[rel]) {
			add("(i) the backup copy of %s differs from the file as it was before the run (%d vs %d bytes)", rel, len(got), len(w.pre[rel]))
		}
		if info, statErr := os.Stat(filepath.Join(assets, filepath.FromSlash(rel))); statErr == nil && info.Mode().Perm() != w.preMode[rel] {
			add("(i-mode) the backup copy of %s has mode %o, the file had %o", rel, info.Mode().Perm(), w.preMode[rel])
		}
	}
	// The backup holds those files and nothing else.
	_ = filepath.WalkDir(assets, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(assets, p)
		if !slices.Contains(w.rels, filepath.ToSlash(rel)) {
			add("(i) the run backup holds %s, which no removal accounts for", filepath.ToSlash(rel))
		}
		return nil
	})

	// (ii) manifest.prior.json is the manifest file as it was before the run.
	prior, err := os.ReadFile(filepath.Join(runDir2, priorManifestName))
	if err != nil {
		add("(ii) cannot read %s: %v", priorManifestName, err)
	} else if !bytes.Equal(prior, w.preManifest) {
		add("(ii) %s differs from the manifest as it was before the run (%d vs %d bytes)", priorManifestName, len(prior), len(w.preManifest))
	}

	// (iii) The restore area: the removed files and the manifest file.
	for _, rel := range w.rels {
		got, err := os.ReadFile(filepath.Join(root2, filepath.FromSlash(rel)))
		if err != nil {
			add("(iii) the restored tree has no %s: %v", rel, err)
			continue
		}
		if !bytes.Equal(got, w.pre[rel]) {
			add("(iii) the restored %s differs from its pre-run bytes", rel)
		}
		if info, statErr := os.Stat(filepath.Join(root2, filepath.FromSlash(rel))); statErr == nil && info.Mode().Perm() != w.preMode[rel] {
			add("(iii-mode) the restored %s has mode %o, it had %o", rel, info.Mode().Perm(), w.preMode[rel])
		}
	}
	if got, err := os.ReadFile(filepath.Join(root2, ".moai", "manifest.json")); err != nil || !bytes.Equal(got, w.preManifest) {
		add("(iii) the restored manifest is not the pre-run manifest (err=%v)", err)
	}

	// (iv) The difference sentinel: restored tree vs the control tree.
	t2, collisions2 := restoreNormalize(ownTreeSnapshot(t, root2, ownTreeOpts{}))
	t0, collisions0 := restoreNormalize(ownTreeSnapshot(t, w.root0, ownTreeOpts{}))
	for _, c := range append(collisions2, collisions0...) {
		add("(iv) two clock-named backup directories collide after normalisation at %s", c)
	}
	d := ownTreeDiff(t2, t0)
	if len(d) == 0 {
		add("(iv) the restored tree and the control tree do not differ at all: E would be inflated to hide nothing")
	}
	for _, p := range d {
		if !restoreInE(p) {
			add("(iv) %s differs between the restored tree and the control tree and is outside the expected set", p)
		}
	}
	if !slices.Contains(d, ".moai/manifest.json") {
		add("(iv) the manifest is not among the differences: the control run's manifest equals the pre-run one, so the restore compared nothing")
	}
	if !slices.ContainsFunc(d, func(p string) bool { return strings.HasPrefix(p, ".moai-backups/") }) {
		add("(iv) no run-backup path differs between the restored tree and the control tree")
	}

	// (iv-c) Inside E, the part a content sentinel cannot judge: the log and state
	// trees may differ in content, but neither tree may hold a path the other lacks.
	for _, p := range d {
		if (strings.HasPrefix(p, ".moai/logs/") || strings.HasPrefix(p, ".moai/state/")) && (t2[p] == "" || t0[p] == "") {
			add("(iv-c) %s exists in only one of the restored tree and the control tree", p)
		}
	}

	// (iv-b) Inside E: the run backup holds what the control run's backup holds
	// plus exactly the removed-asset copies (layouts compared without the run
	// directory's own, clock-derived, name).
	l1 := restoreRunLayout(runDir2)
	l0 := restoreRunLayout(w.runDir0)
	var extra []string
	for _, f := range l1 {
		if !slices.Contains(l0, f) {
			extra = append(extra, f)
		}
	}
	var want []string
	for _, rel := range w.rels {
		want = append(want, migrationBackupSubdir+"/"+rel)
	}
	sort.Strings(want)
	if !slices.Equal(extra, want) {
		add("(iv-b) the run backup holds %v beyond the control run's backup layout, want exactly %v", extra, want)
	}
	for _, f := range l0 {
		if !slices.Contains(l1, f) {
			add("(iv-b) the control run's backup holds %s and the removal run's does not", f)
		}
	}
	return problems
}

// ---------------------------------------------------------------------------
// Fixture and the two runs.
// ---------------------------------------------------------------------------

// runWithRemovalSwitch runs fn with the removal switch at the given position and
// puts the previous value back.
func runWithRemovalSwitch(on bool, fn func()) {
	prev := removalEnabled
	removalEnabled = on
	defer func() { removalEnabled = prev }()
	fn()
}

// restoreRemovedCandidates returns, in order, the candidates that no longer exist
// under root.
func restoreRemovedCandidates(root string, rels []string) []string {
	var gone []string
	for _, rel := range rels {
		if _, err := os.Lstat(filepath.Join(root, filepath.FromSlash(rel))); os.IsNotExist(err) {
			gone = append(gone, rel)
		}
	}
	return gone
}

// buildRestoreWorld builds the fixture, an independent byte copy of it for the
// control run, and runs the whole update flow in both.
func buildRestoreWorld(t *testing.T) restoreWorld {
	t.Helper()
	enableRemovalForTest(t)
	ownProfileSandbox(t)
	home := installMigrationUserCounterparts(t)
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	prevVerbose := updateVerboseMode
	t.Cleanup(func() { updateVerboseMode = prevVerbose; updateLedger.reset() })

	root1 := buildMigrationFixture(t)
	rels := threeManagedCopies(t, root1, home)
	slices.Sort(rels)
	ownAddRegisteredUserCreated(t, root1, restoreUserFile, []byte("# my own skill\n"))
	rewriteManifestNonCanonical(t, root1, nil)

	w := restoreWorld{root1: root1, rels: rels, pre: map[string][]byte{}, preMode: map[string]os.FileMode{}}
	for _, rel := range rels {
		abs := filepath.Join(root1, filepath.FromSlash(rel))
		data, err := os.ReadFile(abs)
		if err != nil {
			t.Fatalf("read %s: %v", rel, err)
		}
		info, err := os.Stat(abs)
		if err != nil {
			t.Fatalf("stat %s: %v", rel, err)
		}
		w.pre[rel], w.preMode[rel] = data, info.Mode().Perm()
	}
	w.preManifest = readManifestBytes(t, root1)
	if info, err := os.Stat(filepath.Join(root1, ".moai", "manifest.json")); err == nil {
		w.preManifestMode = info.Mode().Perm()
	} else {
		t.Fatalf("stat the manifest: %v", err)
	}

	w.root0 = t.TempDir()
	restoreCopyTree(t, root1, w.root0)
	if d := ownTreeDiff(ownTreeSnapshot(t, root1, ownTreeOpts{}), ownTreeSnapshot(t, w.root0, ownTreeOpts{})); len(d) != 0 {
		t.Fatalf("premise: the control copy differs from the fixture before any run: %v", d)
	}

	updateLedger.reset()
	updateVerboseMode = false
	runWithRemovalSwitch(true, func() { w.run1 = runUpdateFlowPlain(t, root1) })
	// The control run starts in a later clock second than every directory run 1
	// named after the clock, so the two runs never share a clock-derived name;
	// the shared-name case is built from this tree by renaming (restoreSameNames).
	awaitClockPast(t, restoreRunStamps(t, root1))
	updateLedger.reset()
	runWithRemovalSwitch(false, func() { w.run0 = runUpdateFlowPlain(t, w.root0) })

	if w.run1.NewPriorManifest != "" {
		w.runDir1 = filepath.Dir(w.run1.NewPriorManifest)
	}
	if w.run0.NewPriorManifest != "" {
		w.runDir0 = filepath.Dir(w.run0.NewPriorManifest)
	}
	w.removedAfterRun1 = restoreRemovedCandidates(root1, rels)
	w.removedAfterRun0 = restoreRemovedCandidates(w.root0, rels)
	return w
}

// problemsWithPrefix keeps the lines of a problem list that start with prefix.
func problemsWithPrefix(problems []string, prefix string) []string {
	var out []string
	for _, p := range problems {
		if strings.HasPrefix(p, prefix) {
			out = append(out, p)
		}
	}
	return out
}

// restoreSameNames returns a copy of the control tree whose clock-derived
// directories carry the names run 1's carry, and the world that points at it.
// It is the shared-name branch of the clock: two runs that landed in the same
// second. It is built by renaming, not by waiting for the clock.
func restoreSameNames(t *testing.T, w restoreWorld) restoreWorld {
	t.Helper()
	root := t.TempDir()
	restoreCopyTree(t, w.root0, root)
	want := filepath.Base(w.runDir1)
	runDir := filepath.Join(root, ".moai-backups", filepath.Base(w.runDir0))
	renamed := filepath.Join(root, ".moai-backups", want)
	if err := os.Rename(runDir, renamed); err != nil {
		t.Fatalf("rename the control run directory: %v", err)
	}
	ns0, _ := filepath.Glob(filepath.Join(root, ".moai", "backups", "update-*"))
	ns1, _ := filepath.Glob(filepath.Join(w.root1, ".moai", "backups", "update-*"))
	if len(ns0) != 1 || len(ns1) != 1 {
		t.Fatalf("expected one namespace backup directory per tree, found %d and %d", len(ns0), len(ns1))
	}
	if err := os.Rename(ns0[0], filepath.Join(root, ".moai", "backups", filepath.Base(ns1[0]))); err != nil {
		t.Fatalf("rename the control namespace backup directory: %v", err)
	}
	ar0, _ := filepath.Glob(filepath.Join(root, ".moai", "archive", "files", "update-migration", "*"))
	ar1, _ := filepath.Glob(filepath.Join(w.root1, ".moai", "archive", "files", "update-migration", "*"))
	if len(ar0) != 1 || len(ar1) != 1 {
		t.Fatalf("expected one legacy-archive directory per tree, found %d and %d", len(ar0), len(ar1))
	}
	if err := os.Rename(ar0[0], filepath.Join(root, ".moai", "archive", "files", "update-migration", filepath.Base(ar1[0]))); err != nil {
		t.Fatalf("rename the control legacy-archive directory: %v", err)
	}
	same := w
	same.root0, same.runDir0 = root, renamed
	return same
}

func TestRemovedAssetBackupRestoreByteIdentical(t *testing.T) {
	w := buildRestoreWorld(t)
	if w.runDir1 == "" || w.runDir0 == "" {
		t.Fatalf("a run left no manifest copy (run 1 %q, control %q): the criterion cannot be judged", w.run1.NewPriorManifest, w.run0.NewPriorManifest)
	}
	if !bytes.Equal(w.run1.StartManifest, w.preManifest) {
		t.Fatalf("the manifest the run loaded is not the manifest the fixture left")
	}
	// The premise of the re-serialisation mutant: the start manifest is not in the
	// canonical shape Save would write, so the two copies cannot coincide.
	if !bytes.HasSuffix(w.preManifest, []byte("\n\n")) {
		t.Fatalf("premise: the start manifest has the canonical shape; a re-serialised copy would equal the original")
	}
	// The two runs' clock-derived names differ (the control run waited for a later
	// second); the shared-name branch is built below by renaming.
	if filepath.Base(w.runDir0) == filepath.Base(w.runDir1) {
		t.Fatalf("premise: the control run and run 1 share the run directory name %s", filepath.Base(w.runDir1))
	}

	if info, err := os.Stat(filepath.Join(w.runDir1, priorManifestName)); err == nil {
		t.Logf("observation: %s has mode %o, the manifest it copies had %o (the file modes of the restore area are asserted for the removed files only)", priorManifestName, info.Mode().Perm(), w.preManifestMode)
	}

	// T1 after run 1, kept for the mutants; the restore below works in place.
	post1 := t.TempDir()
	restoreCopyTree(t, w.root1, post1)
	runDirName := filepath.Base(w.runDir1)
	wSame := restoreSameNames(t, w)

	root2 := w.root1 // the restore turns T1 into T2
	restoreFromRunBackup(t, root2, w.runDir1)

	problems := restoreProblems(t, w, root2, w.runDir1)
	for _, group := range []struct{ name, prefix string }{
		{"ReachedReconciliation", "(v)"},
		{"BackupCopiesAreByteIdentical", "(i)"},
		{"PriorManifestIsTheStartManifest", "(ii)"},
		{"RestoreAreaIsByteIdentical", "(iii)"},
		{"DifferenceSentinel", "(iv)"},
	} {
		t.Run(group.name, func(t *testing.T) {
			for _, p := range problemsWithPrefix(problems, group.prefix) {
				t.Error(p)
			}
		})
	}
	// The same clause set when the control run's clock-derived directories carry
	// the very names run 1's do.
	t.Run("SharedClockNamesJudgeTheSame", func(t *testing.T) {
		for _, p := range restoreProblems(t, wSame, root2, w.runDir1) {
			t.Error(p)
		}
	})

	// The normaliser erases the two clock-derived texts and nothing else.
	t.Run("NormaliserErasesOnlyClockDerivedText", func(t *testing.T) {
		base := map[string]string{
			".moai/backups/update-2026-10-18T10-02-22Z/":          "d 755",
			".moai/backups/update-2026-10-18T10-02-22Z/.moai-run": "f 644 aaaa",
			".moai/backups/update-2026-10-18T10-02-22Z/x/keep.md": "f 644 bbbb",
			".claude/skills/a/SKILL.md":                           "f 644 cccc",
			".moai/backups/not-an-update-dir/y":                   "f 644 dddd",
		}
		later := map[string]string{
			".moai/backups/update-2026-10-18T10-02-23Z/":          "d 755",
			".moai/backups/update-2026-10-18T10-02-23Z/.moai-run": "f 644 zzzz",
			".moai/backups/update-2026-10-18T10-02-23Z/x/keep.md": "f 644 bbbb",
			".claude/skills/a/SKILL.md":                           "f 644 cccc",
			".moai/backups/not-an-update-dir/y":                   "f 644 dddd",
		}
		nb, c1 := restoreNormalize(base)
		nl, c2 := restoreNormalize(later)
		if len(c1)+len(c2) != 0 {
			t.Fatalf("unexpected collisions %v %v", c1, c2)
		}
		if d := ownTreeDiff(nb, nl); len(d) != 0 {
			t.Errorf("two runs that differ only in the clock-derived text still differ after normalising: %v", d)
		}
		again, _ := restoreNormalize(nb)
		if d := ownTreeDiff(nb, again); len(d) != 0 {
			t.Errorf("normalising twice changes the result: %v", d)
		}
		// A timestamp-shaped name anywhere else is NOT erased.
		other := map[string]string{".moai/archive/files/other/20261018_100222/": "d 755", ".moai-backups/20261018_100222/x": "f 644 q"}
		if no, _ := restoreNormalize(other); ownTreeDiff(no, other) != nil {
			t.Errorf("the normaliser renamed a timestamp-shaped name outside the two clock-derived locations: %v", no)
		}
		if nb[".claude/skills/a/SKILL.md"] != "f 644 cccc" || nb[".moai/backups/not-an-update-dir/y"] != "f 644 dddd" {
			t.Errorf("the normaliser touched text that is not clock-derived: %v", nb)
		}
		if nb[".moai/backups/update-<RUN>/x/keep.md"] != "f 644 bbbb" {
			t.Errorf("the normaliser changed a file inside the namespace backup: %v", nb)
		}
		// A real difference inside the namespace backup survives the normalising.
		changed := map[string]string{}
		for k, v := range later {
			changed[k] = v
		}
		changed[".moai/backups/update-2026-10-18T10-02-23Z/x/keep.md"] = "f 644 EEEE"
		nc, _ := restoreNormalize(changed)
		if d := ownTreeDiff(nb, nc); len(d) != 1 {
			t.Errorf("a content difference inside the namespace backup did not survive normalising: %v", d)
		}
		// Two namespace backups in one tree are reported, not merged silently.
		two := map[string]string{
			".moai/backups/update-2026-10-18T10-02-22Z/.moai-run": "f 644 a",
			".moai/backups/update-2026-10-18T10-02-23Z/.moai-run": "f 644 b",
			".moai/backups/update-2026-10-18T10-02-22Z/x":         "f 644 1",
			".moai/backups/update-2026-10-18T10-02-23Z/x":         "f 644 2",
		}
		if _, col := restoreNormalize(two); len(col) == 0 {
			t.Errorf("two different namespace backups in one tree collapsed without a report")
		}
	})

	// Report which members of E actually differ, so an E that hides nothing is
	// visible in the record.
	t.Run("ExpectedSetMembersThatDiffer", func(t *testing.T) {
		t2, _ := restoreNormalize(ownTreeSnapshot(t, root2, ownTreeOpts{}))
		t0, _ := restoreNormalize(ownTreeSnapshot(t, w.root0, ownTreeOpts{}))
		d := ownTreeDiff(t2, t0)
		members := map[string][]string{}
		for _, p := range d {
			switch {
			case p == ".moai/manifest.json":
				members["manifest"] = append(members["manifest"], p)
			case strings.HasPrefix(p, ".moai-backups/"):
				members["run-backup"] = append(members["run-backup"], p)
			case strings.HasPrefix(p, ".moai/logs/"):
				members["logs"] = append(members["logs"], p)
			case strings.HasPrefix(p, ".moai/state/"):
				members["state"] = append(members["state"], p)
			default:
				members["OUTSIDE-E"] = append(members["OUTSIDE-E"], p)
			}
		}
		for _, k := range []string{"manifest", "run-backup", "logs", "state", "OUTSIDE-E"} {
			t.Logf("difference class %-10s : %d path(s)", k, len(members[k]))
		}
		if len(members["manifest"]) == 0 || len(members["run-backup"]) == 0 {
			t.Errorf("the manifest and the run backup must differ from the control run (manifest=%d, run-backup=%d)", len(members["manifest"]), len(members["run-backup"]))
		}
		if len(members["OUTSIDE-E"]) != 0 {
			t.Errorf("differences outside E: %v", members["OUTSIDE-E"])
		}
	})

	// Mutants: a wrong implementation, applied to a copy of T1 before (pre) or
	// after (post) the restore, must turn the pure assertion red.
	mutant := func(name, wantPrefix string, pre func(m string, runDir string), post func(m string)) {
		t.Run("Mutant"+name+"IsCaught", func(t *testing.T) {
			m := t.TempDir()
			restoreCopyTree(t, post1, m)
			runDir := filepath.Join(m, ".moai-backups", runDirName)
			if pre != nil {
				pre(m, runDir)
			}
			restoreFromRunBackup(t, m, runDir)
			if post != nil {
				post(m)
			}
			got := restoreProblems(t, w, m, runDir)
			if len(got) == 0 {
				t.Fatalf("the assertion stayed green in front of the %s mutant", name)
			}
			if wantPrefix != "" && len(problemsWithPrefix(got, wantPrefix)) == 0 {
				t.Errorf("the %s mutant was caught, but not by clause %s: %v", name, wantPrefix, got)
			}
		})
	}
	firstRel := w.rels[0]
	backupCopy := func(runDir, rel string) string {
		return filepath.Join(runDir, migrationBackupSubdir, filepath.FromSlash(rel))
	}
	mutant("CRLFInBackupCopy", "(i)", func(_, runDir string) {
		p := backupCopy(runDir, firstRel)
		data, _ := os.ReadFile(p)
		if err := os.WriteFile(p, bytes.ReplaceAll(data, []byte("\n"), []byte("\r\n")), 0o644); err != nil {
			t.Fatal(err)
		}
	}, nil)
	mutant("TruncatedBackupCopy", "(i)", func(_, runDir string) {
		p := backupCopy(runDir, firstRel)
		data, _ := os.ReadFile(p)
		if err := os.WriteFile(p, data[:len(data)/2], 0o644); err != nil {
			t.Fatal(err)
		}
	}, nil)
	mutant("MissingBackupCopy", "(i)", func(_, runDir string) {
		if err := os.Remove(backupCopy(runDir, firstRel)); err != nil {
			t.Fatal(err)
		}
	}, nil)
	mutant("BackupCopyModeNarrowed", "(i-mode)", func(_, runDir string) {
		if err := os.Chmod(backupCopy(runDir, firstRel), 0o600); err != nil {
			t.Fatal(err)
		}
	}, nil)
	mutant("ReserialisedPriorManifest", "(ii)", func(_, runDir string) {
		p := filepath.Join(runDir, priorManifestName)
		data, _ := os.ReadFile(p)
		var v any
		if err := json.Unmarshal(data, &v); err != nil {
			t.Fatal(err)
		}
		canon, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, canon, 0o644); err != nil {
			t.Fatal(err)
		}
	}, nil)
	mutant("StrayFileInRunBackup", "(iv-b)", func(_, runDir string) {
		if err := os.WriteFile(filepath.Join(runDir, "stray.txt"), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}, nil)
	mutant("StrayRemovedAsset", "(i)", func(_, runDir string) {
		p := backupCopy(runDir, ".claude/skills/other/SKILL.md")
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}, nil)
	mutant("RunAlsoDeletesTheUsersFile", "(iv)", nil, func(m string) {
		if err := os.Remove(filepath.Join(m, filepath.FromSlash(restoreUserFile))); err != nil {
			t.Fatal(err)
		}
	})
	mutant("RunAlsoRewritesAnotherFile", "(iv)", nil, func(m string) {
		p := filepath.Join(m, filepath.FromSlash(migForeignSkill))
		if err := os.WriteFile(p, []byte("# rewritten\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	})
	mutant("RunNarrowsAFileMode", "(iv)", nil, func(m string) {
		if err := os.Chmod(filepath.Join(m, filepath.FromSlash(migForeignSkill)), 0o600); err != nil {
			t.Fatal(err)
		}
	})
	mutant("RunLeavesAFileInTheNamespaceBackupChanged", "(iv)", nil, func(m string) {
		ns, _ := filepath.Glob(filepath.Join(m, ".moai", "backups", "update-*"))
		if len(ns) != 1 {
			t.Fatalf("expected one namespace backup directory, found %d", len(ns))
		}
		var victim string
		_ = filepath.WalkDir(ns[0], func(p string, d fs.DirEntry, err error) error {
			if err == nil && !d.IsDir() && victim == "" && filepath.Base(p) != ".moai-run" {
				victim = p
			}
			return nil
		})
		if victim == "" {
			t.Fatalf("the namespace backup holds no file besides its marker")
		}
		if err := os.WriteFile(victim, []byte("changed\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	})
	mutant("RunWritesAStateFile", "(iv-c)", nil, func(m string) {
		p := filepath.Join(m, ".moai", "state", "stray.json")
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("{}\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	})
	mutant("ManifestNotRestored", "(iii)", nil, func(m string) {
		if err := os.WriteFile(filepath.Join(m, ".moai", "manifest.json"), []byte("{}\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	})

	// Reach mutants: a skipped or empty run 1 must not read as an executed one.
	t.Run("MutantSkippedRunIsCaught", func(t *testing.T) {
		skipped := w
		skipped.run1.Skipped = true
		skipped.removedAfterRun1 = nil
		got := restoreProblems(t, skipped, root2, w.runDir1)
		if len(problemsWithPrefix(got, "(v)")) == 0 {
			t.Errorf("a skipped run 1 with an empty removal list passed the reach clause: %v", got)
		}
	})
	t.Run("MutantUnrestoredTreeIsCaught", func(t *testing.T) {
		got := restoreProblems(t, w, post1, w.runDir1)
		if len(problemsWithPrefix(got, "(iii)")) == 0 {
			t.Errorf("the tree as run 1 left it (nothing restored) passed clause (iii): %v", got)
		}
	})
	t.Run("MutantControlRunThatRemovedFilesIsCaught", func(t *testing.T) {
		bad := w
		bad.removedAfterRun0 = []string{firstRel}
		got := restoreProblems(t, bad, root2, w.runDir1)
		if len(problemsWithPrefix(got, "(v)")) == 0 {
			t.Errorf("a control run that removed a file passed the reach clause: %v", got)
		}
	})
}
