// board_test.go — single-origin board state store (SPEC-KANBAN-BOARD-001
// REQ-KB-005/006/013/021, M1).
//
// The store's location and resolution is the least reversible decision in the
// SPEC, so its tests judge the resolved PATH against a value computed from a
// real repository fixture rather than trusting the borrowed resolution's
// existing green suite — an equality check is insensitive to an offset shared
// by both operands, while the board takes the PARENT of one operand as its
// root (AC-KB-002's rationale).
package kanban

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	gitcore "github.com/modu-ai/moai-adk/internal/core/git"
)

// boardFixtureRepo creates a primary repository plus a linked worktree and
// returns (primary, worktree). Both paths are symlink-resolved.
func boardFixtureRepo(t *testing.T) (string, string) {
	t.Helper()
	// Resolve up front: git normalizes the cwd to its realpath, so on macOS
	// (where t.TempDir lives under /var -> /private/var) expectations built
	// from the unresolved path would mismatch git's output.
	primary, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("resolve primary path: %v", err)
	}
	runGitAt(t, primary, "init", "-b", "main")
	runGitAt(t, primary, "config", "user.email", "board-test@example.com")
	runGitAt(t, primary, "config", "user.name", "Board Test")
	if err := os.WriteFile(filepath.Join(primary, "README.md"), []byte("# board\n"), 0o644); err != nil {
		t.Fatalf("write README: %v", err)
	}
	runGitAt(t, primary, "add", ".")
	runGitAt(t, primary, "commit", "-m", "initial")
	worktreeRaw := filepath.Join(primary, "..", "wt-"+filepath.Base(primary))
	runGitAt(t, primary, "worktree", "add", worktreeRaw, "-b", "topic")
	resolved, err := filepath.EvalSymlinks(worktreeRaw)
	if err != nil {
		t.Fatalf("resolve worktree path: %v", err)
	}
	return primary, resolved
}

// runGitAt runs git in dir, failing the test on error.
func runGitAt(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s: %s: %v", args, dir, string(out), err)
	}
}

// TestBoardRoot_SingleOrigin — AC-KB-002's load-bearing pair. The resolution
// must yield the PRIMARY root from BOTH a worktree and the primary checkout
// itself. The primary-checkout run is the positive control against the bare
// --git-common-dir defect (which returns a relative .git in the primary and
// would resolve the board root to the parent of the current directory).
func TestBoardRoot_SingleOrigin(t *testing.T) {
	t.Parallel()
	primary, worktree := boardFixtureRepo(t)

	fromWorktree, err := BoardRoot(worktree)
	if err != nil {
		t.Fatalf("BoardRoot(worktree) error = %v", err)
	}
	if fromWorktree != primary {
		t.Errorf("BoardRoot(worktree) = %q, want primary %q", fromWorktree, primary)
	}

	// The load-bearing half: same resolution from the primary checkout.
	fromPrimary, err := BoardRoot(primary)
	if err != nil {
		t.Fatalf("BoardRoot(primary) error = %v", err)
	}
	if fromPrimary != primary {
		t.Errorf("BoardRoot(primary) = %q, want %q — resolution must be a function of the repository, not the caller's location", fromPrimary, primary)
	}
}

// TestBoardRoot_FallbackForcedThroughIndirection — AC-KB-002's second half.
// The primary probe is forced to fail via the extracted resolver's
// ExecCommand indirection (older-git simulation); the fallback inside the
// dispatcher must resolve the SAME board root. Non-fallback calls delegate to
// the real git binary, so the expectation is computed from the actual fixture.
// Direct invocation of the fallback would be a vacuous pass.
func TestBoardRoot_FallbackForcedThroughIndirection(t *testing.T) {
	if runtimeIsWindows() {
		t.Skip("fallback mock uses sh -c; skip on windows")
	}
	primary := t.TempDir()
	runGitAt(t, primary, "init", "-b", "main")

	orig := gitcore.ExecCommand
	t.Cleanup(func() { gitcore.ExecCommand = orig })
	gitcore.ExecCommand = func(name string, args ...string) *exec.Cmd {
		joined := strings.Join(args, " ")
		if strings.Contains(joined, "--path-format=absolute") {
			return exec.Command("sh", "-c", "echo 'unknown option: path-format=absolute' >&2; exit 1")
		}
		return orig(name, args...)
	}

	got, err := BoardRoot(primary)
	if err != nil {
		t.Fatalf("BoardRoot(fallback) error = %v", err)
	}
	if got != primary {
		t.Errorf("BoardRoot(fallback) = %q, want %q", got, primary)
	}
}

// TestBoardPath_OwnSegments — the board's path-segment constant is its own
// (REQ-KB-005, AP-24): `.moai/state/kanban-board/`, distinct from the session
// record's per-tree `.moai/state/kanban/` that record.go's stateDirSegments
// names. A board path built from the record's segments is the failure this
// test exists to catch.
func TestBoardPath_OwnSegments(t *testing.T) {
	t.Parallel()
	root := string(filepath.Separator) + "primary-root"

	p := BoardPath(root)
	want := filepath.Join(root, ".moai", "state", "kanban-board", "board.json")
	if p != want {
		t.Fatalf("BoardPath = %q, want %q", p, want)
	}
	// The sibling's directory must not be the board's directory.
	recordDir := filepath.Join(root, ".moai", "state", "kanban")
	if filepath.Dir(p) == recordDir {
		t.Fatalf("board directory %q equals the session-record directory %q", filepath.Dir(p), recordDir)
	}
}

// TestLoadBoard_AbsentVsUnreadable — AC-KB-024, ONE table-driven test so the
// two states are observed to differ against each other (the defect is a
// conflation, invisible when either row is asserted alone). Absent →
// legitimately empty board; unreadable (read failure OR parse failure) →
// unknown, never an empty board.
func TestLoadBoard_AbsentVsUnreadable(t *testing.T) {
	t.Parallel()

	writeState := func(t *testing.T, body string) string {
		t.Helper()
		root := t.TempDir()
		dir := filepath.Join(root, ".moai", "state", "kanban-board")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir board dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dir, "board.json"), []byte(body), 0o600); err != nil {
			t.Fatalf("write board state: %v", err)
		}
		return root
	}

	truncated := `{"cards":[{"spec_id":"SPEC-A-001","column":"run","holder":`

	cases := []struct {
		name        string
		setup       func(t *testing.T) string
		wantUnknown bool
	}{
		{
			name: "absent_file_is_legitimately_empty_board",
			setup: func(t *testing.T) string {
				return t.TempDir() // no board dir at all
			},
			wantUnknown: false,
		},
		{
			name: "truncated_json_is_unknown",
			setup: func(t *testing.T) string {
				return writeState(t, truncated)
			},
			wantUnknown: true,
		},
		{
			name: "well_formed_same_size_reads_successfully",
			setup: func(t *testing.T) string {
				// Same byte length as the truncated row: the failure is
				// conditional on content, not on size (AC-KB-006 control).
				// JSON tolerates trailing whitespace, so pad to exact length.
				body := `{"cards":[]}`
				for len(body) < len(truncated) {
					body += "\n"
				}
				return writeState(t, body)
			},
			wantUnknown: false,
		},
		{
			name: "permission_denied_is_unknown",
			setup: func(t *testing.T) string {
				if runtimeIsWindows() {
					t.Skip("chmod-based denial skipped on windows")
				}
				root := writeState(t, `{"cards":[]}`)
				path := BoardPath(root)
				if err := os.Chmod(path, 0o000); err != nil {
					t.Fatalf("chmod: %v", err)
				}
				t.Cleanup(func() { _ = os.Chmod(path, 0o600) })
				return root
			},
			wantUnknown: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := tc.setup(t)
			st, err := LoadBoard(root)
			if tc.wantUnknown {
				if err == nil {
					t.Fatalf("LoadBoard() err = nil, want unknown; got board %+v", st)
				}
				if !IsBoardUnknown(err) {
					t.Fatalf("err = %v, want IsBoardUnknown", err)
				}
				if st != nil {
					t.Fatalf("unknown board returned a state (%+v) — never present a board as accurate", st)
				}
			} else {
				if err != nil {
					t.Fatalf("LoadBoard() error = %v, want success", err)
				}
				if st == nil {
					t.Fatal("LoadBoard() = nil state, want a board")
				}
				if len(st.Cards) != 0 {
					t.Fatalf("expected empty board, got %d cards", len(st.Cards))
				}
			}
		})
	}
}

// TestLoadBoard_ColumnRecordedNotDerived — AC-KB-004 (REQ-KB-006). The column
// is read from the record; mutating nothing but the recorded column changes
// the board's answer. No progress.md marker participates — the board derives
// its answer from nothing but the record, and the record is the only input.
func TestLoadBoard_ColumnRecordedNotDerived(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	path := BoardPath(root)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	writeBoardRaw(t, path, &BoardState{Cards: []Card{
		{SpecID: "SPEC-KB-0001", Column: "review", Holder: "run-sess-1", LastMovedAt: "2026-08-14T00:00:00Z"},
	}})

	st, err := LoadBoard(root)
	if err != nil {
		t.Fatalf("LoadBoard: %v", err)
	}
	if got := st.Cards[0].Column; got != "review" {
		t.Fatalf("column = %q, want recorded %q", got, "review")
	}

	// Mutate ONLY the recorded column; the board's answer must follow.
	st.Cards[0].Column = "run"
	writeBoardRaw(t, path, st)
	st2, err := LoadBoard(root)
	if err != nil {
		t.Fatalf("LoadBoard after mutation: %v", err)
	}
	if got := st2.Cards[0].Column; got != "run" {
		t.Fatalf("column after recorded mutation = %q, want %q — a derived implementation could not change its answer from the mutation alone", got, "run")
	}
}

// TestLoadBoard_RoundTrip — the minimal M1 card model persists and reloads
// with its fields intact (foundation for M2's full card record).
func TestLoadBoard_RoundTrip(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	path := BoardPath(root)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	want := &BoardState{Cards: []Card{
		{SpecID: "SPEC-KB-0002", Column: "backlog", LastMovedAt: "2026-08-14T01:02:03Z"},
		{SpecID: "SPEC-KB-0003", Column: "run", Holder: "run-sess-2", LastMovedAt: "2026-08-14T04:05:06Z"},
	}}
	writeBoardRaw(t, path, want)

	got, err := LoadBoard(root)
	if err != nil {
		t.Fatalf("LoadBoard: %v", err)
	}
	if len(got.Cards) != len(want.Cards) {
		t.Fatalf("card count = %d, want %d", len(got.Cards), len(want.Cards))
	}
	for i := range want.Cards {
		if got.Cards[i] != want.Cards[i] {
			t.Errorf("card[%d] = %+v, want %+v", i, got.Cards[i], want.Cards[i])
		}
	}
}

// writeBoardRaw writes a board state file directly (test fixture — the
// production write path is WriteBoardState, which enforces the sole-writer
// rule; fixtures bypass it by design to seed known states).
func writeBoardRaw(t *testing.T, path string, st *BoardState) {
	t.Helper()
	encoded, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		t.Fatalf("marshal board: %v", err)
	}
	if err := os.WriteFile(path, append(encoded, '\n'), 0o600); err != nil {
		t.Fatalf("write board raw: %v", err)
	}
}

// runtimeIsWindows reports the test's runtime OS.
func runtimeIsWindows() bool {
	return runtime.GOOS == "windows"
}
