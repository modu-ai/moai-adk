package kanban

// settings_drift_test.go — SPEC-PREMERGE-SETTINGS-DRIFT-001 M1 + M3.
//
// Every verdict below is read from a match count, a file's bytes, or a
// recorded argv string. None is read from a process exit code: card t474
// observed a gate invert while its grep still exited 0, and this card
// reproduced the same shape four more times while being built.
//
// Absence assertions ("0 occurrences", "not created", "unchanged") are made
// only AFTER a positive control establishes that the population being swept is
// live and non-degenerate. An identity assertion over two empty measurements
// asserts nothing, and that vacuity is invisible to any success check — a
// healthy bare remote with no branch pushed answers `ls-remote` with rc=0,
// empty stdout and empty stderr.

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// expectedPredicateCommand is the one recorded command every positive control
// below looks for. It is written out in full rather than assembled from the
// implementation's own constants: a control built from the value under test
// agrees with a mutated implementation.
const expectedPredicateCommand = "git --no-optional-locks status --porcelain -- .claude/settings.json"

// git runs a fixture-construction command and fails the test on error. Fixture
// setup is the one place an error check is the right instrument — it is asking
// "did this step happen at all", not "what is the verdict".
func gitFixture(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("fixture: git %s in %s: %v\n%s", strings.Join(args, " "), dir, err, out)
	}
	return string(out)
}

// newFixtureRepo builds a git repository under t.TempDir() with a committed
// .claude/settings.json and a README.md.
//
// `-b main` pins the initial branch per SPEC-GIT-STATUS-FIXTURE-001 (card
// t474): left to the ambient init.defaultBranch, CI and a local machine
// disagree about the branch name and the divergence surfaces as an unrelated
// failure. user.name/user.email are pinned locally for the same reason — a
// machine without a global identity cannot commit.
func newFixtureRepo(t *testing.T) string {
	t.Helper()
	return newFixtureRepoNamed(t, "target")
}

// newFixtureRepoNamed is newFixtureRepo with the repository's role written
// into its README, which makes two fixture repositories differ in CONTENT.
//
// That is not cosmetic. A commit's SHA is a function of its tree, its author,
// its committer, its message and its timestamp — all of which two identically
// built fixtures share — so two repositories created in the same second
// produced the SAME HEAD, and the two-root control in
// TestAssessSettingsDriftPreservesAndLedgers fired on it. The control was
// right: two assertions over one value are one assertion.
func newFixtureRepoNamed(t *testing.T, role string) string {
	t.Helper()
	dir := t.TempDir()
	// macOS hands out /var/folders/... which is a symlink to /private/var/...;
	// git reports the resolved form, so resolve here or every path comparison
	// below fails for a reason that has nothing to do with the code.
	resolved, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatalf("fixture: resolve %s: %v", dir, err)
	}
	dir = resolved

	gitFixture(t, dir, "init", "-b", "main")
	gitFixture(t, dir, "config", "user.name", "t488 fixture")
	gitFixture(t, dir, "config", "user.email", "t488@example.invalid")
	gitFixture(t, dir, "config", "commit.gpgsign", "false")

	if err := os.MkdirAll(filepath.Join(dir, ".claude"), 0o755); err != nil {
		t.Fatalf("fixture: mkdir .claude: %v", err)
	}
	writeFixtureFile(t, filepath.Join(dir, ".claude", "settings.json"), "{\n  \"committed\": true\n}\n")
	writeFixtureFile(t, filepath.Join(dir, "README.md"), "committed readme for the "+role+" repository\n")
	gitFixture(t, dir, "add", ".claude/settings.json", "README.md")
	gitFixture(t, dir, "commit", "-m", "fixture: baseline")
	return dir
}

func writeFixtureFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("fixture: write %s: %v", path, err)
	}
}

func fixtureSHA256(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("fixture: read %s: %v", path, err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// countCommand reports how many recorded commands equal want. Counting, not
// substring-matching: "contains the predicate" is satisfied by a whole-tree
// status that happens to mention the path, and the F4 mutant is exactly that.
func countCommand(commands []string, want string) int {
	n := 0
	for _, c := range commands {
		if c == want {
			n++
		}
	}
	return n
}

// assertPredicateRan is the positive control the absence assertions stand on.
// It establishes two things the bare number 0 cannot: the recorded list is not
// empty, and this run's expected predicate call is in it exactly once.
func assertPredicateRan(t *testing.T, runner SettingsDriftRunner) {
	t.Helper()
	commands := runner.Commands()
	if len(commands) == 0 {
		t.Fatalf("positive control: recorded command list is empty; every absence assertion below would be vacuous")
	}
	if got := countCommand(commands, expectedPredicateCommand); got != 1 {
		t.Fatalf("positive control: recorded %d occurrences of %q, want exactly 1\nrecorded: %v",
			got, expectedPredicateCommand, commands)
	}
}

// ---------------------------------------------------------------------------
// M1 — the predicate. Fixtures F1..F5 of plan.md §F M1.
// ---------------------------------------------------------------------------

// TestSettingsDriftPredicateHit — AC-PSD-001, fixture F2.
// Catches the "predicate deleted (always 0)" mutant.
func TestSettingsDriftPredicateHit(t *testing.T) {
	t.Parallel()
	dir := newFixtureRepo(t)
	writeFixtureFile(t, filepath.Join(dir, ".claude", "settings.json"), "{\n  \"committed\": false\n}\n")

	runner := NewExecRunner()
	count, raw, err := DetectSettingsDrift(dir, runner)
	if err != nil {
		t.Fatalf("DetectSettingsDrift: %v", err)
	}
	if count != 1 {
		t.Errorf("match count: got %d, want 1\nraw: %q", count, raw)
	}
	if !strings.Contains(raw, SettingsDriftWatchedPath) {
		t.Errorf("raw output %q does not name %q", raw, SettingsDriftWatchedPath)
	}
}

// TestSettingsDriftPredicatePass — AC-PSD-002, fixture F1.
// The 0 is only meaningful on top of the positive control: an implementation
// that never ran the command produces the same 0.
func TestSettingsDriftPredicatePass(t *testing.T) {
	t.Parallel()
	dir := newFixtureRepo(t)

	runner := NewExecRunner()
	count, raw, err := DetectSettingsDrift(dir, runner)
	if err != nil {
		t.Fatalf("DetectSettingsDrift: %v", err)
	}
	assertPredicateRan(t, runner)
	if count != 0 {
		t.Errorf("match count: got %d, want 0\nraw: %q", count, raw)
	}
	if raw != "" {
		t.Errorf("raw output: got %q, want empty", raw)
	}
}

// TestSettingsDriftPredicateIgnoresOtherFile — AC-PSD-003, fixture F3.
// A mutant that mis-specifies the path to .claude/settings.local.json flips
// AC-PSD-001 to 0; this fixture pins the other direction of the same scoping.
func TestSettingsDriftPredicateIgnoresOtherFile(t *testing.T) {
	t.Parallel()
	dir := newFixtureRepo(t)
	writeFixtureFile(t, filepath.Join(dir, "README.md"), "modified readme\n")

	runner := NewExecRunner()
	count, raw, err := DetectSettingsDrift(dir, runner)
	if err != nil {
		t.Fatalf("DetectSettingsDrift: %v", err)
	}
	assertPredicateRan(t, runner)
	if count != 0 {
		t.Errorf("match count: got %d, want 0 (README.md is not watched)\nraw: %q", count, raw)
	}
}

// TestSettingsDriftPredicateScopedToWatchedPath — AC-PSD-004, fixture F4.
// Both files are dirty. A mutant that drops the pathspec reports 2 here.
func TestSettingsDriftPredicateScopedToWatchedPath(t *testing.T) {
	t.Parallel()
	dir := newFixtureRepo(t)
	writeFixtureFile(t, filepath.Join(dir, ".claude", "settings.json"), "{\n  \"drifted\": true\n}\n")
	writeFixtureFile(t, filepath.Join(dir, "README.md"), "modified readme\n")

	runner := NewExecRunner()
	count, raw, err := DetectSettingsDrift(dir, runner)
	if err != nil {
		t.Fatalf("DetectSettingsDrift: %v", err)
	}
	if count != 1 {
		t.Errorf("match count: got %d, want 1 (2 means the pathspec was dropped)\nraw: %q", count, raw)
	}
}

// TestSettingsDriftPredicateArgvRecordedAtExecutionBoundary — AC-PSD-005.
//
// This asserts on the argv the EXECUTOR received, not on a builder's return
// value. Removing --no-optional-locks changes neither the output nor the
// return value — only a side effect (an index write) whose occurrence depends
// on whether the stat cache is stale, which a freshly built fixture cannot
// pin. So the flag is pinned as a string, at the one place that proves the
// executed command carried it.
func TestSettingsDriftPredicateArgvRecordedAtExecutionBoundary(t *testing.T) {
	t.Parallel()
	dir := newFixtureRepo(t)

	runner := NewExecRunner()
	if _, _, err := DetectSettingsDrift(dir, runner); err != nil {
		t.Fatalf("DetectSettingsDrift: %v", err)
	}

	commands := runner.Commands()
	if len(commands) == 0 {
		t.Fatalf("recorded command list is empty; the argv assertion would be vacuous")
	}
	var predicate string
	for _, c := range commands {
		if strings.HasPrefix(c, "git ") && strings.Contains(c, "status") {
			predicate = c
			break
		}
	}
	if predicate == "" {
		t.Fatalf("no status command recorded\nrecorded: %v", commands)
	}
	if predicate != expectedPredicateCommand {
		t.Errorf("recorded argv:\n got: %q\nwant: %q", predicate, expectedPredicateCommand)
	}

	tokens := strings.Fields(predicate)
	idxFlag, idxStatus := -1, -1
	for i, tok := range tokens {
		switch tok {
		case "--no-optional-locks":
			idxFlag = i
		case "status":
			idxStatus = i
		}
	}
	if idxFlag < 0 {
		t.Fatalf("--no-optional-locks absent from executed argv %q", predicate)
	}
	if idxStatus < 0 {
		t.Fatalf("status absent from executed argv %q", predicate)
	}
	if idxFlag > idxStatus {
		t.Errorf("--no-optional-locks at %d must precede status at %d in %q", idxFlag, idxStatus, predicate)
	}
}

// TestSettingsDriftPredicateFailureIsNotPass — fixture F5, REQ-PSD-014 at the
// predicate layer. A directory that is not a repository must not answer 0: an
// unmeasured tree and a clean tree are different states, and collapsing them
// is the nine-day blindness this card exists to close.
func TestSettingsDriftPredicateFailureIsNotPass(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeFixtureFile(t, filepath.Join(dir, "loose.txt"), "not a repository\n")

	count, _, err := DetectSettingsDrift(dir, NewExecRunner())
	if err == nil {
		t.Fatalf("DetectSettingsDrift on a non-repository: got nil error, want a failure")
	}
	if count >= 0 {
		t.Errorf("match count on failure: got %d, want a negative sentinel (a 0 here reads as a pass)", count)
	}
}

// ---------------------------------------------------------------------------
// M3 — preservation and the ledger.
// ---------------------------------------------------------------------------

// driftFixture builds the hit fixture used by the M3 assertions: a repository
// with a dirty .claude/settings.json, a SEPARATE primary root (its own
// repository, so the two HEADs of AC-PSD-007(d-3) are genuinely two
// measurements), and a local bare remote with the branch actually pushed.
func driftFixture(t *testing.T) (worktree, root, remote string) {
	t.Helper()
	worktree = newFixtureRepo(t)
	writeFixtureFile(t, filepath.Join(worktree, ".claude", "settings.json"), "{\n  \"drifted\": true\n}\n")

	// The primary root is a distinct repository. In the simplest fixture the
	// two roots would coincide by accident, and a gate that committed the
	// preserved copy into the primary would then be measured by the target
	// tree's HEAD purely by luck. Production separates them (linked worktree +
	// primary checkout), so the fixture does too — and it differs in content,
	// or the two repositories share a HEAD and the two-root assertion collapses
	// into one (see newFixtureRepoNamed).
	root = newFixtureRepoNamed(t, "primary")

	remote = filepath.Join(t.TempDir(), "origin.git")
	gitFixture(t, worktree, "init", "--bare", remote)
	gitFixture(t, worktree, "remote", "add", "origin", remote)
	gitFixture(t, worktree, "push", "origin", "main")

	// The push is the step that is easy to omit and produces no error when
	// omitted: `ls-remote` against a healthy bare remote with no branch pushed
	// answers rc=0 with empty stdout AND empty stderr, so the before/after
	// identity assertion in AC-PSD-007(d-3) would compare "" to "" and pass
	// while checking nothing. Confirm by CONTENT here, once, at construction.
	refs := lsRemote(t, remote)
	if strings.TrimSpace(refs) == "" {
		t.Fatalf("fixture: ls-remote output is empty after push; the push assertion would be vacuous")
	}
	if !strings.Contains(refs, "refs/heads/main") {
		t.Fatalf("fixture: ls-remote output %q does not name refs/heads/main", refs)
	}
	return worktree, root, remote
}

// lsRemote and headSHA read the tree directly with os/exec — deliberately NOT
// through the recording runner. AC-PSD-007(d-3) is the assertion that survives
// an implementation which bypasses the runner entirely, so it must not consult
// the runner's record to make its judgement.
func lsRemote(t *testing.T, remote string) string {
	t.Helper()
	cmd := exec.Command("git", "ls-remote", "--heads", remote)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("ls-remote %s: %v\n%s", remote, err, out)
	}
	return string(out)
}

func headSHA(t *testing.T, repo string) string {
	t.Helper()
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = repo
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("rev-parse HEAD in %s: %v", repo, err)
	}
	return strings.TrimSpace(string(out))
}

func stagedNames(t *testing.T, repo string) (string, bool) {
	t.Helper()
	cmd := exec.Command("git", "diff", "--cached", "--name-only")
	cmd.Dir = repo
	out, err := cmd.Output()
	if err != nil {
		return "", false
	}
	return string(out), true
}

type ledgerRow struct {
	MeasuredAt    string `json:"measured_at"`
	Card          string `json:"card"`
	Branch        string `json:"branch"`
	Worktree      string `json:"worktree"`
	SourcePath    string `json:"source_path"`
	PreservedPath string `json:"preserved_path"`
	SHA256        string `json:"sha256"`
	SizeBytes     int64  `json:"size_bytes"`
	MatchCount    int    `json:"match_count"`
	Bypassed      bool   `json:"bypassed"`
	PreserveError string `json:"preserve_error,omitempty"`
}

func readLedger(t *testing.T, root string) []ledgerRow {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(SettingsDriftDir(root), SettingsDriftLedgerName))
	if err != nil {
		t.Fatalf("read ledger: %v", err)
	}
	var rows []ledgerRow
	for _, line := range strings.Split(strings.TrimRight(string(data), "\n"), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var row ledgerRow
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			t.Fatalf("ledger line %q is not JSON: %v", line, err)
		}
		rows = append(rows, row)
	}
	return rows
}

func preservedFiles(t *testing.T, root string) []string {
	t.Helper()
	entries, err := os.ReadDir(SettingsDriftDir(root))
	if err != nil {
		t.Fatalf("read preserve dir: %v", err)
	}
	var names []string
	for _, e := range entries {
		if e.Name() == SettingsDriftLedgerName {
			continue
		}
		names = append(names, e.Name())
	}
	return names
}

// TestAssessSettingsDriftPreservesAndLedgers — AC-PSD-007 (a), (b), (d-1),
// (d-2), (d-3), (d-4).
func TestAssessSettingsDriftPreservesAndLedgers(t *testing.T) {
	t.Parallel()
	worktree, root, remote := driftFixture(t)
	source := filepath.Join(worktree, SettingsDriftWatchedPath)
	sourceSHA := fixtureSHA256(t, source)

	// (d-3) controls, taken BEFORE the gate runs, asserted on CONTENT.
	headBefore := headSHA(t, worktree)
	rootHeadBefore := headSHA(t, root)
	refsBefore := lsRemote(t, remote)
	if len(headBefore) != 40 {
		t.Fatalf("control: target-tree HEAD %q is not a 40-char sha", headBefore)
	}
	if len(rootHeadBefore) != 40 {
		t.Fatalf("control: primary-root HEAD %q is not a 40-char sha", rootHeadBefore)
	}
	if headBefore == rootHeadBefore {
		t.Fatalf("control: the two roots resolve to the same HEAD %q; the two-root assertion would be one assertion", headBefore)
	}
	if strings.TrimSpace(refsBefore) == "" {
		t.Fatalf("control: pre-run ls-remote is empty; the push assertion would be vacuous")
	}

	runner := NewExecRunner()
	result := AssessSettingsDrift(SettingsDriftParams{
		Dir:    worktree,
		Root:   root,
		Card:   "t488",
		Branch: "main",
		Runner: runner,
	})

	if result.Status != SettingsDriftDetected {
		t.Fatalf("status: got %q, want %q (err=%v)", result.Status, SettingsDriftDetected, result.Err)
	}
	if result.MatchCount != 1 {
		t.Errorf("match count: got %d, want 1", result.MatchCount)
	}
	if result.PreserveErr != nil {
		t.Fatalf("preserve error: %v", result.PreserveErr)
	}

	// (a) — preserved copy exists under the primary root and matches byte for byte.
	if result.PreservedPath == "" {
		t.Fatalf("preserved path is empty")
	}
	wantDir := SettingsDriftDir(root)
	if filepath.Dir(result.PreservedPath) != wantDir {
		t.Errorf("preserved path %q is not under %q", result.PreservedPath, wantDir)
	}
	original, err := os.ReadFile(source)
	if err != nil {
		t.Fatalf("read source: %v", err)
	}
	preserved, err := os.ReadFile(result.PreservedPath)
	if err != nil {
		t.Fatalf("read preserved: %v", err)
	}
	if len(original) == 0 {
		t.Fatalf("control: the source file is empty; a byte-equality assertion over two empty files asserts nothing")
	}
	if string(preserved) != string(original) {
		t.Errorf("preserved copy differs from the original\n got: %q\nwant: %q", preserved, original)
	}

	// (b) — exactly one ledger row, carrying the fields a lead reads.
	rows := readLedger(t, root)
	if len(rows) != 1 {
		t.Fatalf("ledger rows: got %d, want 1", len(rows))
	}
	row := rows[0]
	if row.PreservedPath != result.PreservedPath {
		t.Errorf("ledger preserved_path: got %q, want %q", row.PreservedPath, result.PreservedPath)
	}
	if row.SHA256 != sourceSHA {
		t.Errorf("ledger sha256: got %q, want %q", row.SHA256, sourceSHA)
	}
	if row.SizeBytes != int64(len(original)) {
		t.Errorf("ledger size_bytes: got %d, want %d", row.SizeBytes, len(original))
	}
	if row.Worktree != worktree {
		t.Errorf("ledger worktree: got %q, want %q", row.Worktree, worktree)
	}
	if row.MatchCount != 1 {
		t.Errorf("ledger match_count: got %d, want 1", row.MatchCount)
	}
	if row.Card != "t488" {
		t.Errorf("ledger card: got %q, want %q", row.Card, "t488")
	}

	// (d-1) positive control, then (d-2) the absence on top of it.
	assertPredicateRan(t, runner)
	for _, forbidden := range []string{"commit", "push"} {
		for _, c := range runner.Commands() {
			for _, tok := range strings.Fields(c) {
				if tok == forbidden {
					t.Errorf("recorded command %q contains %q; the gate must never commit or push", c, forbidden)
				}
			}
		}
	}

	// (d-3) direct observation, reading no record at all. An implementation
	// that bypassed the runner and committed with os/exec passes (d-1)/(d-2)
	// and fails here.
	if got := headSHA(t, worktree); got != headBefore {
		t.Errorf("target-tree HEAD moved: %q -> %q", headBefore, got)
	}
	if got := headSHA(t, root); got != rootHeadBefore {
		t.Errorf("primary-root HEAD moved: %q -> %q (a committed preserved copy lands here, not in the target tree)", rootHeadBefore, got)
	}
	if got := lsRemote(t, remote); got != refsBefore {
		t.Errorf("remote refs changed:\n before: %q\n after:  %q", refsBefore, got)
	}

	// (d-4) the preserved copy is not staged. The staging list can legitimately
	// be empty, so no content control is possible here: the control is only
	// that the command ran and exited normally, which separates "not staged"
	// from "never asked".
	staged, ok := stagedNames(t, root)
	if !ok {
		t.Fatalf("control: git diff --cached did not run cleanly in %s; the staging assertion would be vacuous", root)
	}
	if strings.Contains(staged, "settings-drift") {
		t.Errorf("preserved copy is staged in the primary root: %q", staged)
	}
}

// TestAssessSettingsDriftDoesNotOverwriteOnCollision — AC-PSD-007(c).
//
// Two hits, same card, same content, back to back. Time, card and hash alone
// name the same file, so only the sequence suffix makes "never overwrite"
// deterministic rather than probabilistic.
func TestAssessSettingsDriftDoesNotOverwriteOnCollision(t *testing.T) {
	t.Parallel()
	worktree, root, _ := driftFixture(t)

	params := SettingsDriftParams{Dir: worktree, Root: root, Card: "t488", Branch: "main"}
	params.Runner = NewExecRunner()
	first := AssessSettingsDrift(params)
	params.Runner = NewExecRunner()
	second := AssessSettingsDrift(params)

	if first.PreserveErr != nil || second.PreserveErr != nil {
		t.Fatalf("preserve errors: %v / %v", first.PreserveErr, second.PreserveErr)
	}
	if first.PreservedPath == second.PreservedPath {
		t.Fatalf("both runs preserved to the same path %q; the second overwrote the first", first.PreservedPath)
	}
	names := preservedFiles(t, root)
	if len(names) != 2 {
		t.Errorf("preserved files: got %d (%v), want 2", len(names), names)
	}
	if rows := readLedger(t, root); len(rows) != 2 {
		t.Errorf("ledger rows: got %d, want 2", len(rows))
	}
}

// TestAssessSettingsDriftLeavesOriginalUntouched — AC-PSD-008.
func TestAssessSettingsDriftLeavesOriginalUntouched(t *testing.T) {
	t.Parallel()
	worktree, root, _ := driftFixture(t)
	source := filepath.Join(worktree, SettingsDriftWatchedPath)

	shaBefore := fixtureSHA256(t, source)
	if len(shaBefore) != 64 {
		t.Fatalf("control: pre-run sha256 %q is not 64 hex chars", shaBefore)
	}
	filesBefore := treeFileList(t, worktree)
	if len(filesBefore) == 0 {
		t.Fatalf("control: pre-run file list is empty; a no-net-change assertion over it asserts nothing")
	}

	runner := NewExecRunner()
	result := AssessSettingsDrift(SettingsDriftParams{
		Dir: worktree, Root: root, Card: "t488", Branch: "main", Runner: runner,
	})
	if result.Status != SettingsDriftDetected {
		t.Fatalf("status: got %q, want %q", result.Status, SettingsDriftDetected)
	}

	// These two read no record, so they hold even if nothing was recorded.
	if got := fixtureSHA256(t, source); got != shaBefore {
		t.Errorf("the watched file was modified: %q -> %q", shaBefore, got)
	}
	filesAfter := treeFileList(t, worktree)
	if len(filesAfter) != len(filesBefore) {
		t.Errorf("file count in the target tree changed: %d -> %d", len(filesBefore), len(filesAfter))
	}

	// The record-based assertion is defence on top, not the only defence.
	assertPredicateRan(t, runner)
	for _, c := range runner.Commands() {
		for _, tok := range strings.Fields(c) {
			switch tok {
			case "commit", "push", "add", "checkout", "restore", "reset", "clean", "stash":
				t.Errorf("recorded command %q would modify the target tree", c)
			}
		}
	}
}

// treeFileList walks the tree, skipping .git, and returns relative paths.
func treeFileList(t *testing.T, root string) []string {
	t.Helper()
	var out []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if info.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		out = append(out, rel)
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	return out
}

// TestAssessSettingsDriftPreserveFailureKeepsVerdict — REQ-PSD-015 at the
// package layer. The verdict comes from the match count alone; a preservation
// failure is reported beside it, never folded into it.
func TestAssessSettingsDriftPreserveFailureKeepsVerdict(t *testing.T) {
	t.Parallel()
	worktree, root, _ := driftFixture(t)

	// Make the state directory unwritable by occupying the preserve directory's
	// path with a regular file: MkdirAll then fails for a reason no permission
	// model can wave through, including a test running as root.
	stateDir := filepath.Join(root, ".moai", "state")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatalf("fixture: mkdir state: %v", err)
	}
	writeFixtureFile(t, SettingsDriftDir(root), "not a directory\n")

	result := AssessSettingsDrift(SettingsDriftParams{
		Dir: worktree, Root: root, Card: "t488", Branch: "main", Runner: NewExecRunner(),
	})
	if result.Status != SettingsDriftDetected {
		t.Errorf("status: got %q, want %q (a preservation failure must not flip the verdict)", result.Status, SettingsDriftDetected)
	}
	if result.MatchCount != 1 {
		t.Errorf("match count: got %d, want 1", result.MatchCount)
	}
	if result.PreserveErr == nil {
		t.Errorf("preserve error: got nil, want the failure reported separately")
	}
}

// TestAssessSettingsDriftCleanTree — the pass direction through the gate.
func TestAssessSettingsDriftCleanTree(t *testing.T) {
	t.Parallel()
	worktree := newFixtureRepo(t)
	root := newFixtureRepoNamed(t, "primary")

	runner := NewExecRunner()
	result := AssessSettingsDrift(SettingsDriftParams{
		Dir: worktree, Root: root, Card: "t488", Branch: "main", Runner: runner,
	})
	assertPredicateRan(t, runner)
	if result.Status != SettingsDriftClean {
		t.Errorf("status: got %q, want %q (err=%v)", result.Status, SettingsDriftClean, result.Err)
	}
	if result.MatchCount != 0 {
		t.Errorf("match count: got %d, want 0", result.MatchCount)
	}
	if result.PreservedPath != "" {
		t.Errorf("preserved path: got %q, want empty on a clean tree", result.PreservedPath)
	}
	if _, err := os.Stat(SettingsDriftDir(root)); err == nil {
		t.Errorf("preserve directory was created for a clean tree")
	}
}

// TestLedgerDigestDescribesThePreservedBytes pins the single-read invariant:
// the digest recorded in the ledger must describe the bytes the preserved copy
// actually holds.
//
// A static fixture cannot see the difference between one read and two — both
// produce agreeing values when nothing writes in between. So the interleaving
// is CONSTRUCTED: the test hook fires at the point between the measurement and
// the write and overwrites the source with different content, which is the
// very thing this gate exists because something does unpredictably.
//
// Under one read, both values describe the pre-hook bytes and agree. Under two
// reads, the digest describes the pre-hook bytes and the copy holds the
// post-hook bytes, and they disagree. That divergence is the assertion.
//
// The hook is package-level state, so this test does not run in parallel.
func TestLedgerDigestDescribesThePreservedBytes(t *testing.T) {
	worktree, root, _ := driftFixture(t)
	source := filepath.Join(worktree, SettingsDriftWatchedPath)

	const interleaved = "{\n  \"written-by-someone-else\": true\n}\n"
	original, err := os.ReadFile(source)
	if err != nil {
		t.Fatalf("read source: %v", err)
	}
	// Control: the two contents must actually differ, or the assertion below
	// holds whether or not the file was re-read.
	if string(original) == interleaved {
		t.Fatalf("control: the interleaved content equals the original; the test would pass either way")
	}

	fired := 0
	settingsDriftPreserveTestHook = func() {
		fired++
		if writeErr := os.WriteFile(source, []byte(interleaved), 0o644); writeErr != nil {
			t.Errorf("interleaving write: %v", writeErr)
		}
	}
	t.Cleanup(func() { settingsDriftPreserveTestHook = nil })

	result := AssessSettingsDrift(SettingsDriftParams{
		Dir: worktree, Root: root, Card: "t488", Branch: "main", Runner: NewExecRunner(),
	})
	if result.PreserveErr != nil {
		t.Fatalf("preserve error: %v", result.PreserveErr)
	}
	// Control: the interleaving actually happened. Without this, a hook that
	// was never invoked would leave the assertion asserting nothing.
	if fired != 1 {
		t.Fatalf("control: the interleaving hook fired %d times, want exactly 1", fired)
	}
	if after, readErr := os.ReadFile(source); readErr != nil {
		t.Fatalf("re-read source: %v", readErr)
	} else if string(after) != interleaved {
		t.Fatalf("control: the interleaving write did not land; source is %q", after)
	}

	preserved, err := os.ReadFile(result.PreservedPath)
	if err != nil {
		t.Fatalf("read preserved: %v", err)
	}
	preservedSum := sha256.Sum256(preserved)
	preservedHex := hex.EncodeToString(preservedSum[:])

	rows := readLedger(t, root)
	if len(rows) != 1 {
		t.Fatalf("ledger rows: got %d, want 1", len(rows))
	}
	if rows[0].SHA256 != preservedHex {
		t.Errorf("the ledger digest does not describe the preserved bytes:\n ledger:    %s\n preserved: %s\nthe file was read twice, and something wrote to it in between",
			rows[0].SHA256, preservedHex)
	}
	if rows[0].SizeBytes != int64(len(preserved)) {
		t.Errorf("ledger size_bytes %d does not match the preserved copy's %d bytes", rows[0].SizeBytes, len(preserved))
	}
	// The preserved copy is the pre-hook content, which is what a later reader
	// comparing a third instance needs it to be.
	if string(preserved) != string(original) {
		t.Errorf("preserved copy is not the content that was measured\n got: %q\nwant: %q", preserved, original)
	}
}

// TestSettingsDriftLabelFallsBackWithoutBreakingThePath pins the preserved
// copy's label. The branch fallback is the case that would otherwise turn a
// preserve into a write to a directory that does not exist: a branch name
// carries slashes, and an unsanitised label puts them in a file name.
func TestSettingsDriftLabelFallsBackWithoutBreakingThePath(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, card, branch, want string
	}{
		{"card wins", "t488", "release/v3.2.0", "t488"},
		{"branch when no card", "", "release/v3.2.0", "release-v3-2-0"},
		{"branch when card is blank", "   ", "WT-premerge-drift", "WT-premerge-drift"},
		{"unknown when neither", "", "", "unknown"},
		{"punctuation collapses", "", "a/b:c d", "a-b-c-d"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := settingsDriftLabel(tc.card, tc.branch)
			if got != tc.want {
				t.Errorf("settingsDriftLabel(%q, %q) = %q, want %q", tc.card, tc.branch, got, tc.want)
			}
			if strings.ContainsAny(got, `/\:`) {
				t.Errorf("label %q carries a path separator; the preserve would write outside its directory", got)
			}
		})
	}
}

// TestPreserveUsesTheBranchWhenNoCardIsGiven exercises that fallback end to
// end, so the sanitisation above is pinned at the place it actually matters.
func TestPreserveUsesTheBranchWhenNoCardIsGiven(t *testing.T) {
	t.Parallel()
	worktree, root, _ := driftFixture(t)

	result := AssessSettingsDrift(SettingsDriftParams{
		Dir: worktree, Root: root, Branch: "release/v3.2.0", Runner: NewExecRunner(),
	})
	if result.PreserveErr != nil {
		t.Fatalf("preserve error: %v", result.PreserveErr)
	}
	base := filepath.Base(result.PreservedPath)
	if !strings.HasPrefix(base, "settings.json.release-v3-2-0.") {
		t.Errorf("preserved name %q does not carry the sanitised branch label", base)
	}
	if filepath.Dir(result.PreservedPath) != SettingsDriftDir(root) {
		t.Errorf("preserved copy landed at %q, outside %q", result.PreservedPath, SettingsDriftDir(root))
	}
}

// TestAssessSettingsDriftUndetermined — REQ-PSD-014 through the gate: a
// measurement that did not happen is its own state, never `clean`.
func TestAssessSettingsDriftUndetermined(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeFixtureFile(t, filepath.Join(dir, "loose.txt"), "not a repository\n")
	root := newFixtureRepo(t)

	result := AssessSettingsDrift(SettingsDriftParams{
		Dir: dir, Root: root, Card: "t488", Runner: NewExecRunner(),
	})
	if result.Status != SettingsDriftUndetermined {
		t.Errorf("status: got %q, want %q", result.Status, SettingsDriftUndetermined)
	}
	if result.Err == nil {
		t.Errorf("err: got nil, want the failure reason")
	}
	if result.MatchCount >= 0 {
		t.Errorf("match count: got %d, want a negative sentinel under undetermined", result.MatchCount)
	}
}
