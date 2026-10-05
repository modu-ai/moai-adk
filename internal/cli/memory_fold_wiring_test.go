package cli

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/hook/memo/taxonomy"
)

// pathBeneath reports whether path lies at or under root, comparing cleaned
// absolute paths segment by segment (a prefix string match would treat
// "/home/a" as the parent of "/home/ab").
func pathBeneath(root, path string) bool {
	if root == "" {
		return false
	}
	rel, err := filepath.Rel(filepath.Clean(root), filepath.Clean(path))
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

// TestMemoryFoldOnDone_ExistingClosePathsContained is the containment cell of
// AC-MFB-008 (xi), plan M0: the memory store resolution that every close path
// will reach must stay inside the TestMain home sandbox and never under the
// developer's real home.
//
// It runs in the TestMain-sandboxed environment exactly as the pre-existing
// close-path tests do: no per-test HOME, USERPROFILE, CLAUDE_CONFIG_DIR or
// MOAI_HOME override. M0 asserts on the candidate list the shared resolver
// (memoryCandidateStores) returns. The gate constant (config.EnvMemoryFoldOnDone,
// M1) and the recorder seam of foldClosedCardMemory (M4) do not exist yet; M4
// extends this cell to set the gate and drive the three close paths.
func TestMemoryFoldOnDone_ExistingClosePathsContained(t *testing.T) {
	if capturedRealHome == "" || homeSandboxDir == "" {
		t.Fatalf("TestMain home sandbox not initialised (real=%q sandbox=%q): the cell is unmeasured", capturedRealHome, homeSandboxDir)
	}

	stores, err := memoryCandidateStores(t.TempDir())
	if err != nil {
		t.Fatalf("memoryCandidateStores: %v", err)
	}
	if len(stores) == 0 {
		t.Fatalf("no candidate store recorded: an empty recording reads as unmeasured, not as contained")
	}

	for _, s := range stores {
		if pathBeneath(capturedRealHome, s.Dir) {
			t.Errorf("candidate store %q (origin %q) lies beneath the real home %q", s.Dir, s.Origin, capturedRealHome)
		}
		if !pathBeneath(homeSandboxDir, s.Dir) {
			t.Errorf("candidate store %q (origin %q) is not beneath the sandbox root %q", s.Dir, s.Origin, homeSandboxDir)
		}
	}
}

// TestMemoryFold_ArchiveRecheckedBeforeMemoryRename is the gate-overlay P1
// regression (card t1502 M4): between the apply's archive verification and
// the MEMORY.md rename there is still one unguarded window — the MEMORY.md
// temp-file preparation — and a concurrent author who deletes the folded
// line from the ARCHIVE inside it would otherwise see the fold "succeed"
// with the line surviving in NEITHER file (invariant d2 / REQ-MFB-004
// broken). The pre-rename re-check must therefore cover the archive against
// its EXPECTED POST-APPLY bytes too: on a mismatch the fold aborts, the
// original line stays in MEMORY.md, and the concurrent archive edit wins.
func TestMemoryFold_ArchiveRecheckedBeforeMemoryRename(t *testing.T) {
	dir := specFixtureCopy(t)
	memoryFoldSeam = foldTestSeam{mutateDuringWrite: func(storeDir, name string) {
		if name != "MEMORY.md" {
			return // the archive write's own seam position is unaffected
		}
		// A concurrent author deletes the just-appended line from the
		// archive while the MEMORY.md temp file is being prepared.
		archivePath := filepath.Join(storeDir, fixtureArchive)
		data, err := os.ReadFile(archivePath)
		if err != nil {
			panic(err)
		}
		stripped := bytes.ReplaceAll(data, []byte(line9001+"\n"), nil)
		if err := os.WriteFile(archivePath, stripped, 0o600); err != nil {
			panic(err)
		}
	}}
	runFoldRefused(t, "--card", "t9001", "--yes", "--dir", dir)
	memoryFoldSeam = foldTestSeam{}
	if got := foldRead(t, dir, "MEMORY.md"); !strings.Contains(got, line9001) {
		t.Errorf("the fold removed the planned line although the archive drifted during the MEMORY.md write")
	}
	if got := foldRead(t, dir, fixtureArchive); strings.Contains(got, line9001) {
		t.Errorf("the fold overwrote the concurrent archive edit instead of aborting")
	}
	requireNoTempFiles(t, dir)
}

// ── AC-MFB-008 wiring cells (plan M4) ─────────────────────────────────────
//
// Every cell builds its own isolated environment: temporary HOME,
// USERPROFILE, CLAUDE_CONFIG_DIR and MOAI_HOME (t.Setenv), a git-repo
// project root whose queue is seeded through the unedited todo helpers, and
// the memory store copied from the SPEC fixture into the resolved profile
// path. Package-global seams are saved and restored around each cell and no
// cell runs in parallel.

// wireEnv is one isolated close-path environment.
type wireEnv struct {
	root         string                // project root (CLAUDE_PROJECT_DIR)
	store        *factory.BacklogStore // the queue store
	memDir       string                // the resolved memory store dir
	stderr       *bytes.Buffer         // captured fold-on-done stderr
	rec          *foldOnDoneRecorder   // active path recorder
	seededDigest string                // the store digest right after seeding
	restore      func()                // restores every package seam
}

// wireFixture prepares the environment: temporary HOME/USERPROFILE/
// CLAUDE_CONFIG_DIR/MOAI_HOME, a fixture-backed memory store under the
// profile config dir (memoryCandidateStores resolves it first), and a queue
// seeded with the fixture's card id.
func wireFixture(t *testing.T) *wireEnv {
	t.Helper()
	home := t.TempDir()
	cfgDir := t.TempDir()
	moaiHome := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv(config.EnvClaudeConfigDir, cfgDir)
	t.Setenv(config.EnvHome, moaiHome)
	// The cells simulate an operator-shell close: a test process launched
	// inside a factory lane carries the lane markers and the queue guard
	// would refuse every mutation (REQ-SD-015).
	t.Setenv(config.EnvFactoryRole, "")
	t.Setenv(config.EnvMoaiFactoryWorker, "")
	t.Setenv(config.EnvFactoryBackend, "")

	root, store := todoFixture(t)
	memDir := filepath.Join(cfgDir, "projects", memoryProjectSlug(root), "memory")
	if err := os.MkdirAll(memDir, 0o700); err != nil {
		t.Fatalf("memory dir: %v", err)
	}
	copyFixtureStore(t, memDir)

	// The git-strategy key the auto-done scan's landed ref resolves through
	// (the autoDoneFixture pattern) — harmless for the other close paths.
	cfgDirSections := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(cfgDirSections, 0o700); err != nil {
		t.Fatalf("config dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cfgDirSections, "git-strategy.yaml"),
		[]byte("git_strategy:\n  worktree_base_branch: develop\n"), 0o600); err != nil {
		t.Fatalf("write git-strategy.yaml: %v", err)
	}

	env := &wireEnv{root: root, store: store, memDir: memDir, stderr: &bytes.Buffer{}, seededDigest: storeSHA256(t, memDir)}

	// Save and replace the package seams; restore runs before the
	// environment's temp dirs vanish.
	savedBound, savedStderr, savedRec := memoryFoldOnDoneBound, foldOnDoneStderr, memoryFoldOnDoneRec
	savedFn, savedStep, savedSeam := foldClosedCardMemoryFn, foldOnDoneStepFn, memoryFoldSeam
	env.rec = &foldOnDoneRecorder{}
	memoryFoldOnDoneRec = env.rec
	foldOnDoneStderr = env.stderr
	memoryFoldOnDoneBound = config.DefaultMemoryFoldOnDoneBound
	foldClosedCardMemoryFn = foldClosedCardMemory
	foldOnDoneStepFn = foldOnDoneStep
	memoryFoldSeam = foldTestSeam{}
	env.restore = func() {
		memoryFoldOnDoneBound, foldOnDoneStderr, memoryFoldOnDoneRec = savedBound, savedStderr, savedRec
		foldClosedCardMemoryFn, foldOnDoneStepFn, memoryFoldSeam = savedFn, savedStep, savedSeam
	}
	t.Cleanup(env.restore)

	t.Setenv(config.EnvMemoryFoldOnDone, "") // the gate starts closed
	return env
}

// copyFixtureStore copies the SPEC fixture store into dir (files only — the
// fixture has no subdirectories).
func copyFixtureStore(t *testing.T, dir string) {
	t.Helper()
	fixture := filepath.Join("..", "..", ".moai", "specs",
		"SPEC-MEMORY-FOLD-BUDGET-001", "fixtures", "store-A")
	entries, err := os.ReadDir(fixture)
	if err != nil {
		t.Fatalf("read fixture store: %v", err)
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(fixture, e.Name()))
		if err != nil {
			t.Fatalf("read fixture file %s: %v", e.Name(), err)
		}
		if err := os.WriteFile(filepath.Join(dir, e.Name()), data, 0o600); err != nil {
			t.Fatalf("seed store file %s: %v", e.Name(), err)
		}
	}
}

// seedWireCard seeds the queue with the fixture's card id.
func seedWireCard(t *testing.T, env *wireEnv) {
	t.Helper()
	seedCard(t, env.store, "t9001", "the t9001 card work — merged to develop", factory.BacklogStateQueued)
}

// prepareWireClose lands the per-path pre-conditions: the auto-done scan
// needs a delivering merge on the pinned origin/develop; the --auto cycle
// needs the worker evidence file present. Split from execution so timing
// assertions measure the close itself, not fixture git subprocesses.
func prepareWireClose(t *testing.T, env *wireEnv, path string) {
	t.Helper()
	switch path {
	case "done":
		return
	case "autodone":
		commitOnRef(t, env.root, "Merge branch 'WT-wire' into develop (card t9001)")
		materializeOriginDevelop(t, env.root)
	case "auto":
		evidence := autoEvidencePath(env.root, "t9001")
		if err := os.MkdirAll(filepath.Dir(evidence), 0o700); err != nil {
			t.Fatalf("evidence dir: %v", err)
		}
		if err := os.WriteFile(evidence, []byte("# evidence: verbatim output\n"), 0o600); err != nil {
			t.Fatalf("write evidence: %v", err)
		}
	default:
		t.Fatalf("unknown close path %q", path)
	}
}

// executeWireClose executes one close path after prepareWireClose.
func executeWireClose(t *testing.T, env *wireEnv, path string) (string, string, error) {
	t.Helper()
	switch path {
	case "done":
		return runTodo(t, "done", "t9001")
	case "autodone":
		return runTodo(t, "auto-done")
	case "auto":
		return runWireAutoCycle(t, env)
	}
	t.Fatalf("unknown close path %q", path)
	return "", "", nil
}

// runWireClose drives one close path end to end: pre-conditions first, then
// execution. Returns stdout, stderr (cobra's buffer) and the command error.
func runWireClose(t *testing.T, env *wireEnv, path string) (string, string, error) {
	t.Helper()
	prepareWireClose(t, env, path)
	return executeWireClose(t, env, path)
}

// normalizeWireStdout blanks the environment's own temp root inside stdout —
// the --auto cycle prints the absolute evidence path, so byte comparison
// across two fixtures is only meaningful root-normalized.
func normalizeWireStdout(out string, env *wireEnv) string {
	return strings.ReplaceAll(out, env.root, "<ROOT>")
}

// instrumentFoldWait wraps the real fold wiring with timing of the fold
// call alone (gate P2: the per-card bound binds the FOLD wait window, not
// the whole close path's wall clock — close processing carries its own
// git/queue cost). Returns the reader for the last fold's wait duration.
// Must be installed after wireFixture (which resets the seam).
func instrumentFoldWait(t *testing.T, env *wireEnv) func() time.Duration {
	t.Helper()
	real := foldClosedCardMemoryFn
	var waitNanos atomic.Int64
	foldClosedCardMemoryFn = func(cardID string) {
		start := time.Now()
		real(cardID)
		waitNanos.Store(int64(time.Since(start)))
	}
	return func() time.Duration { return time.Duration(waitNanos.Load()) }
}

// runWireAutoCycle runs one --auto cycle against the wiring env's queue.
// The evidence file is already present, so the first poll collects it and
// the no-op sleep seam never spins.
func runWireAutoCycle(t *testing.T, env *wireEnv) (string, string, error) {
	t.Helper()
	var out bytes.Buffer
	opts := autoOptions{
		wait:      30 * time.Second,
		liveness:  autoTestLiveness(env.root, "t9001", true, true, nil),
		sessionID: "wire-fixture-session",
		sleep:     func(time.Duration) {},
		now:       time.Now,
	}
	err := runAutoCycle(&out, env.store, env.root, opts)
	return out.String(), "", err
}

// wireErrLines returns the captured fold-on-done stderr lines.
func wireErrLines(t *testing.T, env *wireEnv) []string {
	t.Helper()
	var lines []string
	for _, l := range strings.Split(env.stderr.String(), "\n") {
		if strings.TrimSpace(l) != "" {
			lines = append(lines, l)
		}
	}
	return lines
}

// requireOneFoldLine asserts at most one fold-on-done stderr line and, when
// wantFolded is set, that it names the store, the archive file and a count.
func requireOneFoldLine(t *testing.T, env *wireEnv, wantFolded bool) {
	t.Helper()
	lines := wireErrLines(t, env)
	if len(lines) > 1 {
		t.Fatalf("fold-on-done emitted %d stderr lines, want at most one:\n%s", len(lines), env.stderr.String())
	}
	if !wantFolded {
		return
	}
	if len(lines) != 1 {
		t.Fatalf("fold-on-done emitted %d stderr lines, want exactly one:\n%s", len(lines), env.stderr.String())
	}
	for _, want := range []string{env.memDir, fixtureArchive, "1 line"} {
		if !strings.Contains(lines[0], want) {
			t.Errorf("fold-on-done line %q does not name %q", lines[0], want)
		}
	}
}

// requireStoreFolded asserts the post-fold store state (AC-MFB-003 light):
// the t9001 line left MEMORY.md, joined the archive index exactly once, and
// the card's topic file is untouched.
func requireStoreFolded(t *testing.T, env *wireEnv) {
	t.Helper()
	if got := foldRead(t, env.memDir, "MEMORY.md"); strings.Contains(got, line9001) {
		t.Errorf("the t9001 line is still in MEMORY.md after the gated close")
	}
	if got := foldRead(t, env.memDir, fixtureArchive); !strings.Contains(got, line9001) {
		t.Errorf("the t9001 line never reached the archive index")
	}
	if n := countArchiveLines(t, env.memDir, line9001); n != 1 {
		t.Errorf("archive carries %d copies of the t9001 line, want exactly 1", n)
	}
}

// requireQueueArchived asserts the closed card sits in the queue's archive.
func requireQueueArchived(t *testing.T, env *wireEnv) {
	t.Helper()
	rec, err := env.store.LoadPure()
	if err != nil {
		t.Fatalf("load queue: %v", err)
	}
	for i := range rec.Items {
		if rec.Items[i].ID == "t9001" {
			t.Errorf("card t9001 is still live in the queue (state %s)", rec.Items[i].State)
		}
	}
	if len(rec.Archived) == 0 || rec.Archived[0].Item.ID != "t9001" {
		t.Errorf("the archive holds %d entries, first %v; want the archived t9001 first", len(rec.Archived), rec.Archived)
	}
}

// storeSHA256 is the store's whole-content digest (the differential cells
// compare this, not a size).
func storeSHA256(t *testing.T, dir string) string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read store: %v", err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	sort.Strings(names)
	h := sha256.New()
	for _, n := range names {
		data, err := os.ReadFile(filepath.Join(dir, n))
		if err != nil {
			t.Fatalf("read %s: %v", n, err)
		}
		h.Write([]byte(n))
		h.Write(data)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

// TestMemoryFoldOnDone_DisabledDifferential is AC-MFB-008 (i): a close with
// the gate unset and a close with the fold disabled by construction (the
// helper replaced by a no-op) produce equal stdout, stderr, exit code and
// queue record, and the store's digest is unchanged. The accepted-values
// table closes the gate vocabulary: 0/false/yes/"" behave as unset, 1/TRUE/
// " true " enable.
func TestMemoryFoldOnDone_DisabledDifferential(t *testing.T) {
	// Two identical fixtures: one closes with the gate unset, one with the
	// helper replaced by a no-op. Outputs and end states must be equal.
	runs := make([]*wireEnv, 2)
	stdouts := make([]string, 2)
	for i := range runs {
		env := wireFixture(t)
		seedWireCard(t, env)
		if i == 1 {
			foldClosedCardMemoryFn = func(string) {}
		}
		out, _, err := runWireClose(t, env, "done")
		if err != nil {
			t.Fatalf("done run %d: %v", i, err)
		}
		stdouts[i] = out
		runs[i] = env
	}
	if stdouts[0] != stdouts[1] {
		t.Errorf("stdout differs between the gate-unset and disabled-by-construction closes:\n%q\nvs\n%q", stdouts[0], stdouts[1])
	}
	if runs[0].stderr.String() != runs[1].stderr.String() {
		t.Errorf("stderr differs: %q vs %q", runs[0].stderr.String(), runs[1].stderr.String())
	}
	for i, env := range runs {
		requireQueueArchived(t, env)
		requireOneFoldLine(t, env, false)
		if env.seededDigest != storeSHA256(t, env.memDir) {
			t.Errorf("run %d: the store digest changed although the fold was disabled", i)
		}
	}

	// The gate vocabulary (OD-2): every disabled value closes without a fold
	// line; every enabling value folds exactly one line out.
	for _, table := range []struct {
		value string
		open  bool
	}{
		{"0", false}, {"false", false}, {"yes", false}, {"", false}, {"1", true},
		{"TRUE", true}, {" true ", true},
	} {
		env := wireFixture(t)
		seedWireCard(t, env)
		t.Setenv(config.EnvMemoryFoldOnDone, table.value)
		if _, _, err := runWireClose(t, env, "done"); err != nil {
			t.Fatalf("done with gate=%q: %v", table.value, err)
		}
		lines := wireErrLines(t, env)
		if table.open {
			requireStoreFolded(t, env)
			if len(lines) != 1 {
				t.Errorf("gate=%q: %d fold-on-done lines, want 1", table.value, len(lines))
			}
		} else {
			if got := foldRead(t, env.memDir, "MEMORY.md"); !strings.Contains(got, line9001) {
				t.Errorf("gate=%q: the store was folded although the value is falsy", table.value)
			}
			if len(lines) != 0 {
				t.Errorf("gate=%q: %d fold-on-done lines, want 0", table.value, len(lines))
			}
		}
	}
}

// TestMemoryFoldOnDone_EnabledFolds is AC-MFB-008 (ii): with the gate
// enabled the store is folded, exactly one stderr line names the store, the
// archive file and the line count, and stdout equals the disabled run's.
func TestMemoryFoldOnDone_EnabledFolds(t *testing.T) {
	disabled := wireFixture(t)
	seedWireCard(t, disabled)
	disabledOut, _, err := runWireClose(t, disabled, "done")
	if err != nil {
		t.Fatalf("disabled close: %v", err)
	}

	enabled := wireFixture(t)
	seedWireCard(t, enabled)
	t.Setenv(config.EnvMemoryFoldOnDone, "1")
	enabledOut, _, err := runWireClose(t, enabled, "done")
	if err != nil {
		t.Fatalf("enabled close: %v", err)
	}
	if enabledOut != disabledOut {
		t.Errorf("enabled stdout differs from the disabled run:\n%q\nvs\n%q", enabledOut, disabledOut)
	}
	requireStoreFolded(t, enabled)
	requireOneFoldLine(t, enabled, true)
}

// TestMemoryFoldOnDone_FailOpen is AC-MFB-008 (iii): gate enabled, archive
// index absent — the close exits 0 with the identical stdout, the queue
// record is archived, and one stderr line explains the fold failure.
func TestMemoryFoldOnDone_FailOpen(t *testing.T) {
	disabled := wireFixture(t)
	seedWireCard(t, disabled)
	disabledOut, _, err := runWireClose(t, disabled, "done")
	if err != nil {
		t.Fatalf("disabled close: %v", err)
	}

	env := wireFixture(t)
	seedWireCard(t, env)
	if err := os.Remove(filepath.Join(env.memDir, fixtureArchive)); err != nil {
		t.Fatalf("remove archive index: %v", err)
	}
	t.Setenv(config.EnvMemoryFoldOnDone, "1")
	out, _, err := runWireClose(t, env, "done")
	if err != nil {
		t.Fatalf("the close did not fail open: %v", err)
	}
	if out != disabledOut {
		t.Errorf("stdout differs from the disabled run:\n%q\nvs\n%q", out, disabledOut)
	}
	requireQueueArchived(t, env)
	lines := wireErrLines(t, env)
	if len(lines) != 1 {
		t.Fatalf("fail-open emitted %d stderr lines, want exactly one", len(lines))
	}
	// The failure line names the wiring source and the reason. When no
	// archive-pattern file exists at all, the underlying error carries the
	// generic creation pattern rather than a concrete file name; when an
	// unlinked archive-pattern file exists, it names that file. Both are
	// fail-open — assert the wiring prefix plus whichever reason applies.
	if !strings.HasPrefix(lines[0], "memory fold-on-done: t9001:") {
		t.Errorf("fail-open line %q does not carry the fold-on-done prefix with the card id", lines[0])
	}
	if !strings.Contains(lines[0], "no archive index") {
		t.Errorf("fail-open line %q does not name the missing archive index", lines[0])
	}
}

// TestMemoryFoldOnDone_SeededPanic is AC-MFB-008 (vi): the fold helper
// panics — each close path still returns with its normal stdout, exit
// status and queue record, plus at most one stderr line, well under the
// overridden bound's slack. The per-path baseline is the same close path
// with the fold disabled by construction.
func TestMemoryFoldOnDone_SeededPanic(t *testing.T) {
	for _, path := range []string{"done", "autodone", "auto"} {
		baseline := wireFixture(t)
		seedWireCard(t, baseline)
		foldClosedCardMemoryFn = func(string) {}
		prepareWireClose(t, baseline, path)
		baseOut, _, err := executeWireClose(t, baseline, path)
		if err != nil {
			t.Fatalf("%s: disabled close: %v", path, err)
		}

		env := wireFixture(t)
		seedWireCard(t, env)
		foldWait := instrumentFoldWait(t, env)
		foldOnDoneStepFn = func(string, *atomic.Bool) (string, error) { panic("seeded by the wiring test") }
		memoryFoldOnDoneBound = 200 * time.Millisecond
		t.Setenv(config.EnvMemoryFoldOnDone, "1")
		prepareWireClose(t, env, path)
		out, _, err := executeWireClose(t, env, path)
		if err != nil {
			t.Fatalf("%s: the close did not survive the seeded panic: %v", path, err)
		}
		if normalizeWireStdout(out, env) != normalizeWireStdout(baseOut, baseline) {
			t.Errorf("%s: stdout differs from the disabled run:\n%q\nvs\n%q", path, out, baseOut)
		}
		requireQueueArchived(t, env)
		requireOneFoldLine(t, env, false)
		// The bound binds the FOLD wait window, not the close's wall clock
		// (gate P2): a seeded panic returns the fold immediately.
		if wait := foldWait(); wait > 400*time.Millisecond {
			t.Errorf("%s: the fold wait took %s after a seeded panic, want under 400ms", path, wait)
		}
	}
}

// TestMemoryFoldOnDone_BlockedRead is AC-MFB-008 (vii): the store's
// MEMORY.md is a FIFO whose only writer holds it open without ever writing,
// so the fold's read blocks (the writerless open alone returns EOF at once
// under Go's poller on darwin) — with the bound overridden to 200ms, each
// close path returns in under 400ms with its disabled run's stdout, exit
// status and queue record, at most one stderr line, and no write to the
// store afterwards. Unix-only (P11).
func TestMemoryFoldOnDone_BlockedRead(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("FIFO-based blocked-read cell is Unix-only (AC-MFB-008 vii, plan B1)")
	}
	for _, path := range []string{"done", "autodone", "auto"} {
		baseline := wireFixture(t)
		seedWireCard(t, baseline)
		foldClosedCardMemoryFn = func(string) {}
		prepareWireClose(t, baseline, path)
		baseOut, _, err := executeWireClose(t, baseline, path)
		if err != nil {
			t.Fatalf("%s: disabled close: %v", path, err)
		}

		env := wireFixture(t)
		seedWireCard(t, env)
		foldWait := instrumentFoldWait(t, env)
		baselineGoroutines := runtime.NumGoroutine()
		fifo := blockOnRead(t, filepath.Join(env.memDir, "MEMORY.md"))
		memoryFoldOnDoneBound = 200 * time.Millisecond
		t.Setenv(config.EnvMemoryFoldOnDone, "1")
		prepareWireClose(t, env, path)
		out, _, err := executeWireClose(t, env, path)
		if err != nil {
			t.Fatalf("%s: the close did not survive the blocked read: %v", path, err)
		}
		if normalizeWireStdout(out, env) != normalizeWireStdout(baseOut, baseline) {
			t.Errorf("%s: stdout differs from the disabled run:\n%q\nvs\n%q", path, out, baseOut)
		}
		requireQueueArchived(t, env)
		requireOneFoldLine(t, env, false)
		// The bound binds the FOLD wait window (gate P2): the blocked read
		// must be abandoned at the overridden bound (+slack), whatever the
		// close path's own processing cost.
		if wait := foldWait(); wait > 400*time.Millisecond {
			t.Errorf("%s: the fold wait took %s with a 200ms bound, want under 400ms", path, wait)
		}
		// The store gained nothing while its index was unreadable: no
		// archived line, no temporary file.
		if got := foldRead(t, env.memDir, fixtureArchive); strings.Contains(got, line9001) {
			t.Errorf("%s: the store was written although the index read never completed", path)
		}
		requireNoTempFiles(t, env.memDir)
		// Release the reader and confirm the abandoned step drains (leak check).
		fifo.release()
		waitGoroutines(t, baselineGoroutines)
	}
}

// TestMemoryFoldOnDone_RunsAfterQueueWrite is AC-MFB-008 (iv): the fold
// starts only after the queue already holds the archived card. The seam
// reads the queue at fold START and fails if the card is not archived yet.
func TestMemoryFoldOnDone_RunsAfterQueueWrite(t *testing.T) {
	env := wireFixture(t)
	seedWireCard(t, env)
	foldOnDoneStepFn = func(cardID string, _ *atomic.Bool) (string, error) {
		rec, err := env.store.LoadPure()
		if err != nil {
			return "", err
		}
		for i := range rec.Items {
			if rec.Items[i].ID == cardID {
				return "", fmt.Errorf("fold started while card %s was still live in the queue", cardID)
			}
		}
		found := false
		for i := range rec.Archived {
			if rec.Archived[i].Item.ID == cardID {
				found = true
			}
		}
		if !found {
			return "", fmt.Errorf("fold started before card %s was archived", cardID)
		}
		return foldOnDoneStep(cardID, &atomic.Bool{})
	}
	t.Setenv(config.EnvMemoryFoldOnDone, "1")
	if _, _, err := runWireClose(t, env, "done"); err != nil {
		t.Fatalf("done: %v", err)
	}
	requireStoreFolded(t, env)
}

// TestMemoryFoldOnDone_ThreeClosePaths drives done, auto-done and the --auto
// cycle through their real fixtures with the gate on; each folds the store
// exactly once and archives its card.
func TestMemoryFoldOnDone_ThreeClosePaths(t *testing.T) {
	for _, path := range []string{"done", "autodone", "auto"} {
		env := wireFixture(t)
		seedWireCard(t, env)
		t.Setenv(config.EnvMemoryFoldOnDone, "1")
		if _, _, err := runWireClose(t, env, path); err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		requireQueueArchived(t, env)
		requireStoreFolded(t, env)
		requireOneFoldLine(t, env, true)
	}
}

// TestMemoryFoldOnDone_DisabledNeverOpensStore is AC-MFB-008 (viii): gate
// unset, the store's MEMORY.md replaced by a FIFO (whose held writer would
// park any reading open forever), bound at its PRODUCTION value — every
// close path returns in under 1s with its own disabled run's channels and
// the recorder logs no open of any store file. Only a code path that never
// opens the file returns this fast with an empty recorder; a mutant that
// resolves the store and reads MEMORY.md before checking the gate would
// block until the 2s bound and log the open.
func TestMemoryFoldOnDone_DisabledNeverOpensStore(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("FIFO-based gate-off cell is Unix-only (AC-MFB-008 viii, plan B1)")
	}
	for _, path := range []string{"done", "autodone", "auto"} {
		baseline := wireFixture(t)
		seedWireCard(t, baseline)
		foldClosedCardMemoryFn = func(string) {}
		prepareWireClose(t, baseline, path)
		baseOut, _, err := executeWireClose(t, baseline, path)
		if err != nil {
			t.Fatalf("%s: disabled close: %v", path, err)
		}

		env := wireFixture(t)
		seedWireCard(t, env)
		foldWait := instrumentFoldWait(t, env)
		blockOnRead(t, filepath.Join(env.memDir, "MEMORY.md"))
		// Production bound: memoryFoldOnDoneBound stays at
		// config.DefaultMemoryFoldOnDoneBound (no override). The gate stays
		// closed (wireFixture's default).
		prepareWireClose(t, env, path)
		out, _, err := executeWireClose(t, env, path)
		if err != nil {
			t.Fatalf("%s: the gate-off close failed: %v", path, err)
		}
		// The gate check is the first statement of the wiring: the fold wait
		// must be ~0 with the gate off (gate P2 — the assertion binds the
		// fold window, not the close's wall clock).
		if wait := foldWait(); wait > time.Second {
			t.Errorf("%s: the fold wait took %s with the gate off, want ~0 (under 1s)", path, wait)
		}
		if normalizeWireStdout(out, env) != normalizeWireStdout(baseOut, baseline) {
			t.Errorf("%s: stdout differs from the disabled run:\n%q\nvs\n%q", path, out, baseOut)
		}
		stores, opens := env.rec.snapshot()
		if len(opens) != 0 {
			t.Errorf("%s: the recorder logged opens %v with the gate off — a gate-off close must open nothing", path, opens)
		}
		if len(stores) != 0 {
			t.Errorf("%s: the recorder logged store resolutions %v with the gate off", path, stores)
		}
		requireQueueArchived(t, env)
	}
}

// TestMemoryFoldOnDone_ProductionBoundEffective is AC-MFB-008 (ix): gate
// enabled, MEMORY.md a FIFO with the read parked on its silent writer, the
// bound NOT overridden — `todo done` returns in under 3s (the 2s production
// bound + 1s slack), not before the bound, with its disabled run's stdout
// and queue record and at most one stderr line.
func TestMemoryFoldOnDone_ProductionBoundEffective(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("FIFO-based production-bound cell is Unix-only (AC-MFB-008 ix, plan B1)")
	}
	baseline := wireFixture(t)
	seedWireCard(t, baseline)
	foldClosedCardMemoryFn = func(string) {}
	baseOut, _, err := runWireClose(t, baseline, "done")
	if err != nil {
		t.Fatalf("disabled close: %v", err)
	}

	env := wireFixture(t)
	seedWireCard(t, env)
	foldWait := instrumentFoldWait(t, env)
	fifo := blockOnRead(t, filepath.Join(env.memDir, "MEMORY.md"))
	// No bound override: memoryFoldOnDoneBound is the production constant.
	t.Setenv(config.EnvMemoryFoldOnDone, "1")
	out, _, err := executeWireClose(t, env, "done")
	if err != nil {
		t.Fatalf("done: %v", err)
	}
	// The bound binds the FOLD wait window (gate P2): the blocked read must
	// be abandoned at the production bound (>= bound proves it was in
	// force; < bound + slack proves the close was not stalled past it).
	wait := foldWait()
	if wait > 3*time.Second {
		t.Errorf("the fold wait took %s against the production bound, want under 3s", wait)
	}
	if wait < config.DefaultMemoryFoldOnDoneBound {
		t.Errorf("the fold wait was %s, earlier than the production bound %s — the read did not block, so the bound was not in force", wait, config.DefaultMemoryFoldOnDoneBound)
	}
	if out != baseOut {
		t.Errorf("stdout differs from the disabled run:\n%q\nvs\n%q", out, baseOut)
	}
	requireQueueArchived(t, env)
	requireOneFoldLine(t, env, false)
	fifo.release()
}

// TestMemoryFoldOnDone_AbandonedStepWritesNothing is the gate-overlay P1-4
// regression (card t1502 M4): after the bound expires and the abandonment
// report is emitted, the resumed worker must not begin any new write. The
// pause seam stalls the step between plan and apply; the deadline expires
// mid-stall and the caller reports "abandoned" first — without the
// abandonment guard the resumed worker proceeds to apply and publishes:
// the card line deleted from MEMORY.md and filed into the archive.
func TestMemoryFoldOnDone_AbandonedStepWritesNothing(t *testing.T) {
	env := wireFixture(t)
	seedWireCard(t, env)
	memoryFoldOnDoneBound = 200 * time.Millisecond
	t.Setenv(config.EnvMemoryFoldOnDone, "1")
	// Pause the step between plan and apply (the mutateResult seam sits
	// inside buildFoldPlan): 500ms stall against a 200ms bound.
	memoryFoldSeam = foldTestSeam{
		mutateResult: func(snap taxonomy.StoreSnapshot) taxonomy.StoreSnapshot {
			time.Sleep(500 * time.Millisecond)
			return snap
		},
	}
	foldWait := instrumentFoldWait(t, env)
	baselineGoroutines := runtime.NumGoroutine()
	if _, _, err := runWireClose(t, env, "done"); err != nil {
		t.Fatalf("done: %v", err)
	}
	lines := wireErrLines(t, env)
	if len(lines) != 1 || !strings.Contains(lines[0], "abandoned") {
		t.Fatalf("want exactly one abandonment stderr line, got %v", lines)
	}
	// The bound binds the FOLD wait window (gate P2): the abandonment must
	// fire at the overridden bound (+slack), not on close-processing cost.
	if wait := foldWait(); wait > 400*time.Millisecond {
		t.Errorf("the fold wait took %s with a 200ms bound, want under 400ms", wait)
	}
	// Wait for the resumed worker to drain before judging the store — its
	// post-deadline fate is exactly what this cell decides.
	waitGoroutines(t, baselineGoroutines)

	// The worker resumed after the deadline: it must have written nothing —
	// the t9001 line stays in MEMORY.md and never reached the archive.
	if got := foldRead(t, env.memDir, "MEMORY.md"); !strings.Contains(got, line9001) {
		t.Errorf("the abandoned step deleted the card line from MEMORY.md after reporting no write")
	}
	if got := foldRead(t, env.memDir, fixtureArchive); strings.Contains(got, line9001) {
		t.Errorf("the abandoned step filed the line into the archive after the deadline")
	}
	requireNoTempFiles(t, env.memDir)
}

// TestMemoryFoldOnDone_BoundConstantCeiling is AC-MFB-008 (x):
// config.DefaultMemoryFoldOnDoneBound equals 2 seconds and is at or below
// the 5s ceiling of REQ-MFB-007 — by name, so a constant of minutes cannot
// pass even though every other cell overrides the bound.
func TestMemoryFoldOnDone_BoundConstantCeiling(t *testing.T) {
	bound := config.DefaultMemoryFoldOnDoneBound
	if bound != 2*time.Second {
		t.Errorf("config.DefaultMemoryFoldOnDoneBound = %s, want exactly 2s", bound)
	}
	if bound <= 0 || bound > 5*time.Second {
		t.Errorf("config.DefaultMemoryFoldOnDoneBound = %s violates 0 < bound <= 5s", bound)
	}
}

// waitGoroutines waits for the goroutine count to return to the baseline
// after a FIFO cell released its blocked reader (the leak check of
// AC-MFB-008 (vii)).
func waitGoroutines(t *testing.T, baseline int) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if runtime.NumGoroutine() <= baseline {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	if n := runtime.NumGoroutine(); n > baseline+2 {
		t.Errorf("goroutine count %d did not return near baseline %d within 2s — the abandoned step leaked", n, baseline)
	}
}
