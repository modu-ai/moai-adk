package worktree

// Card t2070 — the migration's speed and its announcements.
//
// The measured causes (progress.md): the plan resolved the home root once per
// tree (one git subprocess each) and compared every tree against every active
// process working directory with symlink resolution on each pair; the move
// captured two manifests of eight sequential git reads each; and the plan
// printed nothing, so a project with hundreds of trees looked hung.

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/worktreeroot"
)

func TestPlanMigration_ResolvesTheDestinationRootOnce(t *testing.T) {
	w := m6NewWorld(t)
	orig := destinationResolver
	var calls int32
	destinationResolver = func(root, name string, apply bool) (string, error) {
		atomic.AddInt32(&calls, 1)
		return orig(root, name, apply)
	}
	t.Cleanup(func() { destinationResolver = orig })

	plans, err := planMigrationFor(w.repo, time.Now(), true)
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	if len(plans) < 3 {
		t.Fatalf("fixture produced %d plans; the test needs several trees", len(plans))
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("the home root was resolved %d times for %d trees; want exactly 1", got, len(plans))
	}
	for _, p := range plans {
		if p.destination == "" {
			continue
		}
		want, perr := worktreeroot.PathFor(w.repo, filepath.Base(p.source))
		if perr != nil {
			t.Fatalf("PathFor: %v", perr)
		}
		if p.destination != want {
			t.Errorf("destination of %s = %s; want %s", p.source, p.destination, want)
		}
	}
}

func TestPlanMigration_DryRunResolvesTheRootOnceAndWritesNothing(t *testing.T) {
	w := m6NewWorld(t)
	orig := destinationResolver
	var calls int32
	destinationResolver = func(root, name string, apply bool) (string, error) {
		atomic.AddInt32(&calls, 1)
		return orig(root, name, apply)
	}
	t.Cleanup(func() { destinationResolver = orig })
	if _, err := planMigrationFor(w.repo, time.Now(), false); err != nil {
		t.Fatalf("plan: %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("the dry run resolved the home root %d times; want exactly 1", got)
	}
}

// referenceManifest is the capture as it ran before t2070: eight git reads, one
// after the other, the first failure ending the capture.
func referenceManifest(tree string) (treeManifest, error) {
	var m treeManifest
	g := &manifestGit{tree: tree}
	git := func(args ...string) string { return g.run(false, args...) }
	gitHEAD := func(args ...string) string { return g.run(true, args...) }
	m.HEAD = strings.TrimSpace(gitHEAD("rev-parse", "HEAD"))
	m.Branch = strings.TrimSpace(gitHEAD("rev-parse", "--abbrev-ref", "HEAD"))
	m.Unpushed = gitHEAD("rev-list", "HEAD", "--not", "--remotes")
	m.Status = git("status", "--porcelain")
	staged := git("diff", "--cached")
	unstaged := git("diff")
	m.Diff = staged + unstaged
	untracked := splitNul(git("ls-files", "--others", "--exclude-standard", "-z"))
	ignored := splitNul(git("ls-files", "--others", "--ignored", "--exclude-standard", "-z"))
	if g.err != nil {
		return m, g.err
	}
	m.IgnoredCount = len(ignored)
	m.ContentDigest = hashUntrackedContent(tree, untracked)
	m.IgnoredDigest = hashIgnoredMetadata(tree, ignored)
	reportsCount, cerr := countReportFiles(tree)
	if cerr != nil {
		return m, cerr
	}
	m.ReportsCount = reportsCount
	return m, nil
}

func TestCaptureTreeManifest_EqualsTheSequentialCapture(t *testing.T) {
	w := m6NewWorld(t)
	tree := w.movedTree
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	must(os.WriteFile(filepath.Join(tree, "README.md"), []byte("changed\n"), 0o644))
	must(os.WriteFile(filepath.Join(tree, "staged.txt"), []byte("staged\n"), 0o644))
	m6Git(t, tree, "add", "staged.txt")
	must(os.WriteFile(filepath.Join(tree, "untracked.txt"), []byte("untracked\n"), 0o644))
	must(os.WriteFile(filepath.Join(tree, ".gitignore"), []byte("ignored.log\n"), 0o644))
	must(os.WriteFile(filepath.Join(tree, "ignored.log"), []byte("ignored\n"), 0o644))
	must(os.MkdirAll(filepath.Join(tree, ".moai", "reports", "x"), 0o755))
	must(os.WriteFile(filepath.Join(tree, ".moai", "reports", "x", "a.md"), []byte("r\n"), 0o644))

	want, werr := referenceManifest(tree)
	got, gerr := captureTreeManifest(tree)
	if werr != nil || gerr != nil {
		t.Fatalf("capture errors: reference %v, current %v", werr, gerr)
	}
	if got != want {
		t.Errorf("manifest differs from the sequential capture:\n got  %+v\n want %+v", got, want)
	}
}

func TestCaptureTreeManifest_FailureReadsTheSameAsTheSequentialCapture(t *testing.T) {
	notARepo := t.TempDir()
	_, werr := referenceManifest(notARepo)
	_, gerr := captureTreeManifest(notARepo)
	if werr == nil || gerr == nil {
		t.Fatalf("a directory that is not a repository must fail the capture: reference %v, current %v", werr, gerr)
	}
	if werr.Error() != gerr.Error() {
		t.Errorf("failure differs: current %q, sequential %q", gerr, werr)
	}
}

func TestCaptureTreeManifest_TimeoutStillFailsTheCapture(t *testing.T) {
	w := m6NewWorld(t)
	orig := manifestRead
	manifestRead = func(dir string, args ...string) (string, error) {
		if len(args) > 0 && args[0] == "status" {
			return "", fmt.Errorf("%w after 1ms", ErrCommandTimeout)
		}
		return orig(dir, args...)
	}
	t.Cleanup(func() { manifestRead = orig })
	if _, err := captureTreeManifest(w.movedTree); err == nil || !errors.Is(err, ErrCommandTimeout) {
		t.Fatalf("a timed-out read must fail the capture with ErrCommandTimeout, got %v", err)
	}
}

func TestCaptureTreeManifest_ReadsRunTogether(t *testing.T) {
	w := m6NewWorld(t)
	orig := manifestRead
	var inFlight, peak int32
	manifestRead = func(dir string, args ...string) (string, error) {
		n := atomic.AddInt32(&inFlight, 1)
		for {
			p := atomic.LoadInt32(&peak)
			if n <= p || atomic.CompareAndSwapInt32(&peak, p, n) {
				break
			}
		}
		time.Sleep(80 * time.Millisecond)
		out, err := orig(dir, args...)
		atomic.AddInt32(&inFlight, -1)
		return out, err
	}
	t.Cleanup(func() { manifestRead = orig })

	start := time.Now()
	if _, err := captureTreeManifest(w.movedTree); err != nil {
		t.Fatalf("capture: %v", err)
	}
	elapsed := time.Since(start)
	if atomic.LoadInt32(&peak) < 2 {
		t.Errorf("the manifest's git reads ran one at a time (peak concurrency %d)", peak)
	}
	// Eight reads of 80ms each take 640ms one after the other.
	if elapsed > 400*time.Millisecond {
		t.Errorf("manifest capture took %v with eight 80ms reads; they should overlap", elapsed)
	}
}

// ---- which agent-named trees the migration moves ----

// The count "N tree(s) to move" included a tree named agent-<hex> (operator's
// run on a project with 648 registered trees). The classifier decides by where
// a tree sits, not by its name alone: agent-* under <repo>/.claude/worktrees is
// the runtime's own and is never moved; the same name under <repo>/.moai/worktrees
// is moai-made by the legacy-root witness and is moved. This pins both answers
// so a change to either is deliberate.
func TestClassify_AgentNamedTreeFollowsItsRoot(t *testing.T) {
	primary := m6Real(t, t.TempDir())
	mk := func(rel string) string {
		p := filepath.Join(primary, rel)
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatal(err)
		}
		return p
	}
	claudeTree := mk(".claude/worktrees/agent-a6e7a057a4c257825")
	moaiTree := mk(".moai/worktrees/agent-a6e7a057a4c257825")
	ctx := ClassifyContext{PrimaryRoot: primary, HomeProjectsDir: filepath.Join(primary, "home-projects")}

	if got := Classify(WorktreeRecord{Path: claudeTree, Branch: "worktree-agent-a6e7a057"}, ctx).Class; got != ClassClaudeNative {
		t.Errorf("agent-* under .claude/worktrees classified %q; the runtime's own tree is never moved", got)
	}
	if got := Classify(WorktreeRecord{Path: moaiTree, Branch: "worktree-agent-a6e7a057"}, ctx).Class; got != ClassMoaiMade {
		t.Errorf("agent-* under .moai/worktrees classified %q; everything under the legacy moai root is moai-made", got)
	}
}

// ---- the process working-directory index ----

func TestProcessCWDIndex_AnswersLikeTheDirectComparison(t *testing.T) {
	base := m6Real(t, t.TempDir())
	mk := func(p string) string {
		full := filepath.Join(base, p)
		if err := os.MkdirAll(full, 0o755); err != nil {
			t.Fatal(err)
		}
		return full
	}
	treeA := mk("a-tree")
	treeB := mk("a-tree-other") // a sibling that shares the prefix
	inside := mk("a-tree/sub/deep")
	other := mk("elsewhere/x")
	link := filepath.Join(base, "link-to-sub")
	if err := os.Symlink(filepath.Join(treeA, "sub"), link); err != nil {
		t.Skipf("no symlinks here: %v", err)
	}
	treeLink := filepath.Join(base, "link-to-tree")
	if err := os.Symlink(treeA, treeLink); err != nil {
		t.Skipf("no symlinks here: %v", err)
	}
	missing := filepath.Join(treeA, "gone", "dir") // lsof can report a removed directory

	// The same directory under another spelling: only a comparison of the
	// directories themselves, not of their strings, sees it. It exists on a
	// filesystem that ignores case (macOS, Windows) and is skipped elsewhere.
	respelled := filepath.Join(base, "A-TREE", "sub", "deep")
	respelledAlias := false
	if _, err := os.Stat(respelled); err == nil && evalOrClean(respelled) != evalOrClean(inside) {
		respelledAlias = true
	}

	trees := []string{treeA, treeB, filepath.Join(base, "never-created"), treeLink}
	scenarios := [][]string{
		nil,
		{other},
		{inside},
		{treeA},
		{treeB},
		{link},
		{treeLink},
		{missing},
		{other, inside, link},
		{"", "/", other},
	}
	if respelledAlias {
		scenarios = append(scenarios, []string{respelled})
		if !processUsesWorktree([]string{respelled}, treeA) {
			t.Fatalf("the direct comparison no longer sees a respelled directory inside the tree; the premise of this case is gone")
		}
	}
	for i, dirs := range scenarios {
		idx := newProcessCWDIndex(dirs)
		for _, tree := range trees {
			want := processUsesWorktree(dirs, tree)
			if got := idx.usesWorktree(tree); got != want {
				t.Errorf("scenario %d %v tree %s: index %v, direct comparison %v", i, dirs, tree, got, want)
			}
		}
	}
}

func TestProcessCWDIndex_ManyTreesAgainstManyProcessesIsQuick(t *testing.T) {
	base := m6Real(t, t.TempDir())
	var trees, dirs []string
	for i := 0; i < 120; i++ {
		tree := filepath.Join(base, "trees", fmt.Sprintf("t%03d", i))
		if err := os.MkdirAll(filepath.Join(tree, "src"), 0o755); err != nil {
			t.Fatal(err)
		}
		trees = append(trees, tree)
	}
	for i := 0; i < 900; i++ {
		d := filepath.Join(base, "procs", fmt.Sprintf("p%03d", i%30), fmt.Sprintf("c%03d", i))
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
		dirs = append(dirs, d)
	}
	dirs = append(dirs, filepath.Join(trees[7], "src")) // one process really is inside tree 7

	start := time.Now()
	idx := newProcessCWDIndex(dirs)
	used := 0
	for _, tree := range trees {
		if idx.usesWorktree(tree) {
			used++
		}
	}
	elapsed := time.Since(start)
	if used != 1 {
		t.Errorf("trees in use = %d; want exactly 1", used)
	}
	// 120 trees x 901 working directories measured about 200ms per tree when
	// every pair resolved symlinks again.
	if elapsed > 3*time.Second {
		t.Errorf("120 trees against 901 process directories took %v", elapsed)
	}
}

// ---- the announcements ----

func TestTreeProgress_PlanAnnouncesTheCountAndThatItCanTakeLong(t *testing.T) {
	var buf bytes.Buffer
	p := newTreeProgress(&buf)
	p.planBegin(550)
	got := buf.String()
	if !strings.Contains(got, "550") {
		t.Errorf("the announcement must name the tree count; got %q", got)
	}
	if !strings.Contains(strings.ToLower(got), "can take a long time") {
		t.Errorf("the announcement must say the migration can take a long time; got %q", got)
	}
}

func TestTreeProgress_PlanAnnouncesNothingForNoTrees(t *testing.T) {
	var buf bytes.Buffer
	newTreeProgress(&buf).planBegin(0)
	if buf.Len() != 0 {
		t.Errorf("no trees to check should print nothing; got %q", buf.String())
	}
}

func TestTreeProgress_ScanCountsTreesAndIsThrottled(t *testing.T) {
	var buf bytes.Buffer
	p := newTreeProgress(&buf)
	clock := time.Unix(1_000, 0)
	p.now = func() time.Time { return clock }
	p.planBegin(1000)
	for i := 1; i <= 1000; i++ {
		p.scan(i, 1000)
	}
	p.planEnd()
	out := buf.String()
	if !strings.Contains(out, "(1/1000)") || !strings.Contains(out, "(1000/1000)") {
		t.Errorf("the first and the last tree must be shown; got:\n%s", out)
	}
	if lines := strings.Count(out, "Checking worktrees ("); lines > 3 {
		t.Errorf("a frozen clock printed %d scan lines for 1000 trees; the line must be throttled", lines)
	}

	buf.Reset()
	p2 := newTreeProgress(&buf)
	tick := time.Unix(1_000, 0)
	p2.now = func() time.Time { tick = tick.Add(1500 * time.Millisecond); return tick }
	p2.planBegin(5)
	for i := 1; i <= 5; i++ {
		p2.scan(i, 5)
	}
	p2.planEnd()
	if lines := strings.Count(buf.String(), "Checking worktrees ("); lines != 5 {
		t.Errorf("a clock that moves past the interval every tree should show all 5; got %d", lines)
	}
}

func TestTreeProgress_MoveShowsElapsedAndRemainingTime(t *testing.T) {
	var buf bytes.Buffer
	p := newTreeProgress(&buf)
	clock := time.Unix(5_000, 0)
	p.now = func() time.Time { return clock }
	p.begin(3)
	p.tree(1, "/x/one")
	if strings.Contains(buf.String(), "left") {
		t.Errorf("no remaining-time estimate is possible before a tree has finished; got %q", buf.String())
	}
	clock = clock.Add(10 * time.Second)
	buf.Reset()
	p.tree(2, "/x/two")
	line := buf.String()
	if !strings.Contains(line, "elapsed 10s") || !strings.Contains(line, "about 20s left") {
		t.Errorf("tree 2 of 3 after 10s should read elapsed 10s and about 20s left; got %q", line)
	}
}

func TestMigrationEngine_AnnouncesBeforeThePlanRuns(t *testing.T) {
	w := m6NewWorld(t)
	var buf bytes.Buffer
	progress := newTreeProgress(&buf)

	var sawAtPlan string
	orig := destinationResolver
	destinationResolver = func(root, name string, apply bool) (string, error) {
		if sawAtPlan == "" {
			sawAtPlan = buf.String()
		}
		return orig(root, name, apply)
	}
	t.Cleanup(func() { destinationResolver = orig })

	// The registered trees under the legacy folders, counted before the run
	// moves any of them.
	wantCount := 0
	listing, _ := listGitOutput(w.repo, "worktree", "list", "--porcelain")
	for _, line := range strings.Split(listing, "\n") {
		path, ok := strings.CutPrefix(line, "worktree ")
		if !ok {
			continue
		}
		slashed := filepath.ToSlash(path)
		if strings.Contains(slashed, "/.moai/worktrees/") || strings.Contains(slashed, "/.claude/worktrees/") {
			wantCount++
		}
	}

	// The apply runs under the migration lock, which also makes the state
	// directory the journal is written into.
	if err := withMigrationLock(w.repo, func() error {
		_, err := runMigrationEngineWith(w.repo, true, "migrate", progress)
		return err
	}); err != nil {
		t.Fatalf("engine: %v", err)
	}
	if !strings.Contains(strings.ToLower(sawAtPlan), "can take a long time") {
		t.Errorf("the notice must be on screen before the plan does any work; at plan start the output was %q", sawAtPlan)
	}
	out := buf.String()
	if i, j := strings.Index(out, "can take a long time"), strings.Index(out, "to move"); i < 0 || j < 0 || i > j {
		t.Errorf("the long-run notice must come before the 'to move' line; output:\n%s", out)
	}
	if wantCount == 0 || !strings.Contains(out, fmt.Sprintf("checking %d registered", wantCount)) {
		t.Errorf("the notice must name the %d registered trees under the legacy folders; output:\n%s", wantCount, out)
	}
}

// ---- the move pipeline ----

// pipelineWorld is the shared world plus extra legacy trees, so several trees
// are eligible for the move at once.
func pipelineWorld(t *testing.T, extra int) *m6World {
	t.Helper()
	w := m6NewWorld(t)
	for i := 0; i < extra; i++ {
		name := fmt.Sprintf("extra%02d", i)
		p := filepath.Join(w.repo, ".moai", "worktrees", name)
		m6Git(t, w.repo, "worktree", "add", "-q", "-b", "WT-"+name, "--", p)
	}
	return w
}

func applyUnderLock(t *testing.T, w *m6World) *migrationReport {
	t.Helper()
	var report *migrationReport
	err := withMigrationLock(w.repo, func() error {
		var rerr error
		report, rerr = runMigrationEngineWith(w.repo, true, "migrate", nil)
		return rerr
	})
	if err != nil {
		t.Fatalf("engine: %v", err)
	}
	return report
}

func TestMigrationEngine_ReadsManifestsAheadButBoundedAndMovesInPlanOrder(t *testing.T) {
	w := pipelineWorld(t, 6)
	orig := manifestCapture
	var mu sync.Mutex
	var inFlight, peak int
	manifestCapture = func(tree string) (string, error) {
		mu.Lock()
		inFlight++
		if inFlight > peak {
			peak = inFlight
		}
		mu.Unlock()
		time.Sleep(60 * time.Millisecond)
		d, err := orig(tree)
		mu.Lock()
		inFlight--
		mu.Unlock()
		return d, err
	}
	t.Cleanup(func() { manifestCapture = orig })

	report := applyUnderLock(t, w)

	if report.moved < 7 {
		t.Fatalf("moved %d trees; the world has at least 7 movable (legacy6 + 6 extras)", report.moved)
	}
	if peak < 2 {
		t.Errorf("manifests were read one tree at a time (peak %d); the next trees' manifests should be read while one moves", peak)
	}
	if peak > config.DefaultWorktreeMigrateManifestLookahead {
		t.Errorf("peak of %d manifests in flight exceeds the lookahead bound %d", peak, config.DefaultWorktreeMigrateManifestLookahead)
	}

	// The rows keep the plan's order: the move itself is still one tree at a time.
	listed := map[string]int{}
	for i, r := range report.rows {
		listed[r.Source] = i
	}
	last := -1
	var movedNames []string
	for _, r := range report.rows {
		if r.Outcome != "moved" {
			continue
		}
		movedNames = append(movedNames, filepath.Base(r.Source))
		if i := listed[r.Source]; i < last {
			t.Errorf("rows out of order at %s", r.Source)
		} else {
			last = i
		}
	}
	// The move list records the moves in the same order as the rows.
	ml, mlErr := loadMoveList(w.repo)
	if mlErr != nil {
		t.Fatalf("move list: %v", mlErr)
	}
	var recorded []string
	for _, e := range ml {
		if e.Result == "moved" {
			recorded = append(recorded, filepath.Base(e.OldPath))
		}
	}
	if strings.Join(recorded, ",") != strings.Join(movedNames, ",") {
		t.Errorf("move-list order %v differs from the report's %v", recorded, movedNames)
	}
}

func TestMigrationEngine_AManifestReadAheadStillMustMatchAfterTheMove(t *testing.T) {
	w := pipelineWorld(t, 2)
	orig := manifestCapture
	manifestCapture = func(tree string) (string, error) {
		d, err := orig(tree)
		if tree == w.movedTree {
			return "stale-" + d, err // the tree changed after it was read
		}
		return d, err
	}
	t.Cleanup(func() { manifestCapture = orig })

	report := applyUnderLock(t, w)

	for _, r := range report.rows {
		if r.Source == w.movedTree {
			if r.Outcome == "moved" {
				t.Fatalf("a tree whose manifest no longer matched was reported moved: %+v", r)
			}
			if r.Outcome != "reversed" {
				t.Errorf("outcome = %q; want reversed", r.Outcome)
			}
		}
	}
	if _, err := os.Stat(w.movedTree); err != nil {
		t.Errorf("the mismatching tree must be back at its source: %v", err)
	}
	if report.moved < 2 {
		t.Errorf("one stale manifest must not stop the other trees; moved %d", report.moved)
	}
}
