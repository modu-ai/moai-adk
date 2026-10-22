package backup_test

// retention_realdata_test.go — the real-data harness of AC-UBR-06
// (SPEC-UPDATE-BACKUP-RETENTION-001). It prunes a COPY of a project's two backup
// roots named by T1733_CLONE and checks the result against a brute-force oracle
// that reads the disk before and after the prune; it does not re-run the pruner's
// selection. It is skipped unless T1733_CLONE is set, because the input is a copy
// made with `cp -Rcp` from a real project, never a fixture of this repository.
//
// The harness prunes its input: use one fresh copy per run. It stamps every run
// directory with the modification time its name encodes (distinct, explicit, set
// after the copy and before the prune) so a copy that does not preserve times
// cannot change the ordering and the path-ascending tie-break is never reached.

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/cli/update/backup"
	"github.com/modu-ai/moai-adk/internal/config"
)

type rdKey struct{ path, sum string }

type rdRun struct {
	dir    string
	root   string // "cfg" or "ns"
	at     time.Time
	keys   map[rdKey]struct{}
	exempt bool
}

var (
	rdCfgName = regexp.MustCompile(`^[0-9]{8}_[0-9]{6}$`)
	rdNsName  = regexp.MustCompile(`^(update|v2-to-v3)-([0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}-[0-9]{2}-[0-9]{2})Z(-[0-9]+)?$`)
)

func rdStamp(root, name string) (time.Time, bool) {
	if root == "cfg" {
		if !rdCfgName.MatchString(name) {
			return time.Time{}, false
		}
		ts, err := time.Parse("20060102_150405", name)
		return ts, err == nil
	}
	m := rdNsName.FindStringSubmatch(name)
	if m == nil {
		return time.Time{}, false
	}
	ts, err := time.Parse("2006-01-02T15-04-05", m[2])
	if err != nil {
		return time.Time{}, false
	}
	if m[3] != "" {
		n, _ := strconv.Atoi(strings.TrimPrefix(m[3], "-"))
		ts = ts.Add(time.Duration(n) * time.Second)
	}
	return ts, true
}

// rdCensus lists the runs under both roots (a run is one directory here; the
// harness asserts no run marker exists) and the entries that are not runs.
func rdCensus(t *testing.T, clone string) (runs []*rdRun, unmanagedDirs, looseFiles int) {
	t.Helper()
	roots := []struct{ tag, dir string }{
		{"cfg", filepath.Join(clone, ".moai-backups")},
		{"ns", filepath.Join(clone, ".moai", "backups")},
	}
	for _, r := range roots {
		ents, err := os.ReadDir(r.dir)
		if err != nil {
			continue
		}
		for _, e := range ents {
			p := filepath.Join(r.dir, e.Name())
			if !e.IsDir() {
				looseFiles++
				continue
			}
			at, ok := rdStamp(r.tag, e.Name())
			if !ok {
				unmanagedDirs++
				continue
			}
			run := &rdRun{dir: p, root: r.tag, at: at, keys: map[rdKey]struct{}{}}
			_ = filepath.WalkDir(p, func(path string, d fs.DirEntry, err error) error {
				rel, _ := filepath.Rel(p, path)
				rel = filepath.ToSlash(rel)
				if err != nil {
					run.exempt = true
					return nil
				}
				if d.IsDir() {
					if rel == ".template-defaults" {
						return filepath.SkipDir
					}
					return nil
				}
				if !d.Type().IsRegular() {
					run.exempt = true
					return nil
				}
				if !strings.Contains(rel, "/") && (rel == ".moai-run" || rel == ".complete" || rel == "backup_metadata.json") {
					if rel == ".moai-run" {
						t.Fatalf("run marker present at %s: the harness treats one directory as one run", path)
					}
					return nil
				}
				f, oerr := os.Open(path)
				if oerr != nil {
					run.exempt = true
					return nil
				}
				defer func() { _ = f.Close() }()
				h := sha256.New()
				if _, cerr := io.Copy(h, f); cerr != nil {
					run.exempt = true
					return nil
				}
				run.keys[rdKey{rel, hex.EncodeToString(h.Sum(nil))}] = struct{}{}
				return nil
			})
			runs = append(runs, run)
		}
	}
	return runs, unmanagedDirs, looseFiles
}

func TestRetentionRealData_AC_UBR06(t *testing.T) {
	clone := os.Getenv("T1733_CLONE")
	if clone == "" {
		t.Skip("T1733_CLONE not set: the real-data harness needs a cp -Rcp copy of a project's backup roots")
	}
	keep := 5
	if v := os.Getenv("T1733_KEEP"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			t.Fatal(err)
		}
		keep = n
	}
	pastKeep := config.DefaultBackupPastVersionsKeep

	// Stamp every run directory with the time its name encodes, after the copy.
	pre, preUnmanaged, preLoose := rdCensus(t, clone)
	seen := map[time.Time]string{}
	for _, r := range pre {
		if prev, dup := seen[r.at]; dup {
			t.Fatalf("stamp collision %s vs %s", prev, r.dir)
		}
		seen[r.at] = r.dir
		if err := os.Chtimes(r.dir, r.at, r.at); err != nil {
			t.Fatal(err)
		}
	}
	sort.Slice(pre, func(i, j int) bool { return pre[i].at.After(pre[j].at) })
	if len(pre) <= keep {
		t.Fatalf("only %d runs for keepCount %d", len(pre), keep)
	}
	window, excess := pre[:keep], pre[keep:]

	// Oracle input: the protected set, computed from the disk census.
	inWindow := map[rdKey]bool{}
	for _, r := range window {
		for k := range r.keys {
			inWindow[k] = true
		}
	}
	newest := map[rdKey]int{} // key -> index of the newest excess holder
	for j, r := range excess {
		for k := range r.keys {
			if inWindow[k] {
				continue
			}
			if at, ok := newest[k]; !ok || j < at {
				newest[k] = j
			}
		}
	}
	byPath := map[string][]rdKey{}
	for k := range newest {
		byPath[k.path] = append(byPath[k.path], k)
	}
	protected := map[rdKey]bool{}
	pathsOverBudget := 0
	for _, ks := range byPath {
		sort.Slice(ks, func(a, b int) bool {
			if newest[ks[a]] != newest[ks[b]] {
				return newest[ks[a]] < newest[ks[b]]
			}
			return ks[a].sum < ks[b].sum
		})
		if len(ks) > pastKeep {
			pathsOverBudget++
		}
		for i := 0; i < len(ks) && i < pastKeep; i++ {
			protected[ks[i]] = true
		}
	}
	distinctBefore := map[rdKey]bool{}
	for _, r := range pre {
		for k := range r.keys {
			distinctBefore[k] = true
		}
	}
	preCleanBefore := 0
	for _, r := range pre {
		if r.root == "cfg" {
			_ = filepath.WalkDir(r.dir, func(p string, d fs.DirEntry, err error) error {
				if err == nil && d.Type().IsRegular() && strings.Contains(p, string(filepath.Separator)+"pre-clean"+string(filepath.Separator)) {
					preCleanBefore++
				}
				return nil
			})
		}
	}
	exemptBefore := 0
	for _, r := range pre {
		if r.exempt {
			exemptBefore++
		}
	}

	rep := backup.PruneBackups(clone, keep, pastKeep)
	if rep.RootRefused {
		t.Fatalf("the clone's backup root was refused: %+v", rep)
	}

	post, postUnmanaged, postLoose := rdCensus(t, clone)
	postHolders := map[rdKey][]string{}
	distinctAfter := map[rdKey]bool{}
	for _, r := range post {
		for k := range r.keys {
			postHolders[k] = append(postHolders[k], r.dir)
			distinctAfter[k] = true
		}
	}
	windowDirs := map[string]bool{}
	for _, r := range window {
		windowDirs[r.dir] = true
	}
	preByDir := map[string]*rdRun{}
	for _, r := range pre {
		preByDir[r.dir] = r
	}

	// (a) counts.
	retainedExcess := 0
	retainedNonExempt := 0
	for _, r := range post {
		if windowDirs[r.dir] {
			continue
		}
		retainedExcess++
		if !preByDir[r.dir].exempt {
			retainedNonExempt++
		}
	}
	windowPresent := 0
	for _, r := range post {
		if windowDirs[r.dir] {
			windowPresent++
		}
	}
	t.Logf("AC06a keepCount=%d runs-before=%d runs-after=%d window-present=%d R(retained-excess)=%d R-non-exempt=%d exempt-runs-before=%d",
		keep, len(pre), len(post), windowPresent, retainedExcess, retainedNonExempt, exemptBefore)
	t.Logf("AC06a unmanaged-dirs before=%d after=%d loose-files before=%d after=%d report-unmanaged-count=%d report-unmanaged-bytes=%d",
		preUnmanaged, postUnmanaged, preLoose, postLoose, rep.Unmanaged.Count, rep.Unmanaged.Bytes)
	if windowPresent != keep {
		t.Errorf("window runs present = %d, want %d", windowPresent, keep)
	}
	if len(post) != keep+retainedExcess {
		t.Errorf("runs after = %d, want keepCount+R = %d", len(post), keep+retainedExcess)
	}
	if postUnmanaged != preUnmanaged || postLoose != preLoose {
		t.Errorf("unmanaged entries changed: dirs %d->%d, loose files %d->%d", preUnmanaged, postUnmanaged, preLoose, postLoose)
	}

	// (b) coverage, irredundancy, bound — from the post-prune disk census.
	uncovered := 0
	for k := range protected {
		if len(postHolders[k]) == 0 {
			uncovered++
		}
	}
	redundant := 0
	for _, r := range post {
		if windowDirs[r.dir] || preByDir[r.dir].exempt {
			continue
		}
		needed := false
		for k := range r.keys {
			if protected[k] && len(postHolders[k]) == 1 {
				needed = true
				break
			}
		}
		if !needed {
			redundant++
		}
	}
	t.Logf("AC06b protected-versions=%d paths-over-budget=%d uncovered=%d redundant-retained-non-exempt=%d retained-non-exempt=%d bound-ok=%v",
		len(protected), pathsOverBudget, uncovered, redundant, retainedNonExempt, retainedNonExempt <= len(protected))
	if uncovered != 0 {
		t.Errorf("AC-UBR-06(b): %d protected versions have no retained holder", uncovered)
	}
	if redundant != 0 {
		t.Errorf("AC-UBR-06(b): %d retained non-exempt excess runs are redundant", redundant)
	}
	if retainedNonExempt > len(protected) {
		t.Errorf("AC-UBR-06(b): %d retained non-exempt excess runs exceed %d protected versions", retainedNonExempt, len(protected))
	}

	// (c) no path loses a version among its pastKeep most recent non-window versions.
	lost, lostInBudget, lostInWindow := 0, 0, 0
	lostPaths := map[string]bool{}
	for k := range distinctBefore {
		if distinctAfter[k] {
			continue
		}
		lost++
		lostPaths[k.path] = true
		if protected[k] {
			lostInBudget++
		}
		if inWindow[k] {
			lostInWindow++
		}
	}
	t.Logf("AC06c distinct-file-versions-before=%d after=%d lost=%d paths-with-lost-versions=%d lost-inside-budget=%d lost-in-window=%d",
		len(distinctBefore), len(distinctAfter), lost, len(lostPaths), lostInBudget, lostInWindow)
	if lostInBudget != 0 || lostInWindow != 0 {
		t.Errorf("AC-UBR-06(c): lost inside budget=%d, lost in window=%d", lostInBudget, lostInWindow)
	}

	// (d) the pre-clean copy under the config root.
	preCleanAfter := 0
	for _, r := range post {
		if r.root != "cfg" {
			continue
		}
		_ = filepath.WalkDir(r.dir, func(p string, d fs.DirEntry, err error) error {
			if err == nil && d.Type().IsRegular() && strings.Contains(p, string(filepath.Separator)+"pre-clean"+string(filepath.Separator)) {
				preCleanAfter++
			}
			return nil
		})
	}
	t.Logf("AC06d pre-clean-files before=%d after=%d", preCleanBefore, preCleanAfter)
	if preCleanBefore == 0 || preCleanAfter < 1 {
		t.Errorf("AC-UBR-06(d): pre-clean files before=%d after=%d", preCleanBefore, preCleanAfter)
	}

	// (e) the report agrees with the disk and names each retained excess run.
	removedItems, retainedItems := 0, 0
	for _, it := range rep.Items {
		switch it.Action {
		case backup.RetentionActionRemoved:
			removedItems++
		case backup.RetentionActionRetained:
			retainedItems++
		}
	}
	t.Logf("AC06e report runs=%d window=%d items=%d removed-items=%d retained-items=%d removed-dirs=%d protected=%d",
		rep.Runs, rep.Window, len(rep.Items), removedItems, retainedItems, rep.Removed, rep.Protected)
	reasons := map[string]int{}
	for _, it := range rep.Items {
		if it.Action == backup.RetentionActionRetained {
			reasons[it.Reason]++
			if it.Run == "" || len(it.Dirs) == 0 || it.Reason == "" {
				t.Errorf("AC-UBR-06(e): retained item without a name, directories or reason: %+v", it)
			}
		}
	}
	t.Logf("AC06e retained-reasons=%v", reasons)
	if rep.Runs != len(pre) || rep.Window != keep || len(rep.Items) != len(excess) {
		t.Errorf("AC-UBR-06(e): report runs/window/items = %d/%d/%d, want %d/%d/%d", rep.Runs, rep.Window, len(rep.Items), len(pre), keep, len(excess))
	}
	if retainedItems != retainedExcess {
		t.Errorf("AC-UBR-06(e): report retained %d excess runs, disk holds %d", retainedItems, retainedExcess)
	}
	if rep.Removed != len(pre)-len(post) {
		t.Errorf("AC-UBR-06(e): report removed %d directories, disk lost %d", rep.Removed, len(pre)-len(post))
	}
	if rep.Protected != len(protected) {
		t.Errorf("report protected = %d, oracle counted %d", rep.Protected, len(protected))
	}
	if want := cfgRootUnmanaged(clone); rep.Unmanaged.Count != want {
		t.Errorf("AC-UBR-06(a): report unmanaged count = %d, config-root non-run entries on disk = %d", rep.Unmanaged.Count, want)
	}
}

// cfgRootUnmanaged counts the entries directly under the config backup root whose
// names no backup writer produces (directories of any other name, loose files).
func cfgRootUnmanaged(clone string) int {
	ents, err := os.ReadDir(filepath.Join(clone, ".moai-backups"))
	if err != nil {
		return 0
	}
	n := 0
	for _, e := range ents {
		if !rdCfgName.MatchString(e.Name()) || !e.IsDir() {
			n++
		}
	}
	return n
}
