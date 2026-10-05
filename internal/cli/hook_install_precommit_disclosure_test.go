package cli

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// This file carries SPEC-PRECOMMIT-PRESERVE-001 M2's disclosure criteria: the
// backup (AC-PCP-003), the notice and its two-writer wiring (AC-PCP-004), the
// no-silent-replacement invariant (AC-PCP-006), the non-bare output
// (AC-PCP-007), backup uniqueness (AC-PCP-009) and the failure precedences
// (AC-PCP-010). Attribution criteria live in
// hook_install_precommit_attribution_test.go.

// findBackups returns every pre-commit.bak.* artifact in root's hooks dir.
func findBackups(t *testing.T, root string) []string {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(root, ".git", "hooks", "pre-commit.bak.*"))
	if err != nil {
		t.Fatalf("glob backups: %v", err)
	}
	return matches
}

// TestPreCommitModifiedHookIsBackedUp — AC-PCP-003 (REQ-PCP-003).
//
// The backup must hold the PRE-RUN bytes: a backup of the replacement content
// would preserve nothing. The marker-bearing hook is planted with a record that
// does not match it, so the three-way classifier reads it as user-modified.
func TestPreCommitModifiedHookIsBackedUp(t *testing.T) {
	root := newPreCommitTestRepo(t)
	writeExistingHook(t, root, previousPreCommitHookContent)
	writeRecord(t, root, digestOf("something else MoAI once wrote\n"))

	installer := NewPreCommitInstaller(root)
	if err := installer.InstallPreCommitHook(false); err != nil {
		t.Fatalf("InstallPreCommitHook: %v", err)
	}

	backups := findBackups(t, root)
	if len(backups) != 1 {
		t.Fatalf("expected exactly one backup of the user-modified hook, found %d: %v", len(backups), backups)
	}
	got, err := os.ReadFile(backups[0])
	if err != nil {
		t.Fatalf("read backup %s: %v", backups[0], err)
	}
	if string(got) != previousPreCommitHookContent {
		t.Errorf("backup must hold the pre-run hook bytes: got %d bytes, want the %d-byte pre-run content", len(got), len(previousPreCommitHookContent))
	}
	if readHook(t, root) != preCommitHookContent {
		t.Errorf("hook must be replaced with the incoming content")
	}
}

// TestPreCommitWarningWriterWiring — AC-PCP-004 clause (ii) (REQ-PCP-004).
//
// The deciding check is semantic, not textual: it parses the caller sources,
// locates every call to installPreCommitHookOptional, and resolves the
// warning-writer argument. `moai init` must pass cmd.ErrOrStderr() itself;
// `moai update` must pass the `errOut` variable, whose single assignment inside
// the same function must be cmd.ErrOrStderr(). A grep cannot decide this: the
// update call line reads `errOut`, not `ErrOrStderr`, and an identically-named
// wrong variable greps identically (acceptance.md AC-PCP-004 Decides).
func TestPreCommitWarningWriterWiring(t *testing.T) {
	t.Parallel()
	root := precommitProjectRoot(t)

	t.Run("init_binds_warning_writer_to_command_stderr", func(t *testing.T) {
		path := filepath.Join(root, "internal", "cli", "init.go")
		calls := installOptionalCallsIn(t, parseCallerSource(t, path))
		if len(calls) != 1 {
			t.Fatalf("%s: expected exactly one installPreCommitHookOptional call, found %d", filepath.Base(path), len(calls))
		}
		call := calls[0]
		if len(call.Args) != 4 {
			t.Fatalf("%s: expected 4 arguments (root, skip, progress writer, warning writer), got %d", filepath.Base(path), len(call.Args))
		}
		if !isCmdCall(call.Args[3], "cmd", "ErrOrStderr") {
			t.Errorf("init: warning-writer argument must be cmd.ErrOrStderr(), got %s", types.ExprString(call.Args[3]))
		}
		if !isCmdCall(call.Args[2], "cmd", "ErrOrStderr") {
			t.Errorf("init: progress-writer argument must stay cmd.ErrOrStderr(), got %s", types.ExprString(call.Args[2]))
		}
	})

	t.Run("update_binds_warning_writer_to_errOut_stderr", func(t *testing.T) {
		path := filepath.Join(root, "internal", "cli", "update_template_sync.go")
		file := parseCallerSource(t, path)
		calls := installOptionalCallsIn(t, file)
		if len(calls) != 1 {
			t.Fatalf("%s: expected exactly one installPreCommitHookOptional call, found %d", filepath.Base(path), len(calls))
		}
		call := calls[0]
		if len(call.Args) != 4 {
			t.Fatalf("%s: expected 4 arguments (root, skip, progress writer, warning writer), got %d", filepath.Base(path), len(call.Args))
		}
		ident, ok := call.Args[3].(*ast.Ident)
		if !ok || ident.Name != "errOut" {
			t.Fatalf("update: warning-writer argument must be the errOut variable, got %s", types.ExprString(call.Args[3]))
		}
		progressIdent, ok := call.Args[2].(*ast.Ident)
		if !ok || progressIdent.Name != "out" {
			t.Errorf("update: progress-writer argument must stay the out variable, got %s", types.ExprString(call.Args[2]))
		}
		// Resolve the identifiers inside the enclosing function only: a second,
		// unrelated `out := cmd.OutOrStdout()` lives in a later function in the
		// same file, and a file-wide lookup would be fooled by it.
		enclosing := enclosingFuncDecl(file, call)
		if enclosing == nil {
			t.Fatal("could not locate the enclosing function of the install call")
		}
		assertAssignedFromCmdCall(t, enclosing, "errOut", "ErrOrStderr")
		assertAssignedFromCmdCall(t, enclosing, "out", "OutOrStdout")
	})

	t.Run("no_other_call_sites", func(t *testing.T) {
		cliDir := filepath.Join(root, "internal", "cli")
		entries, err := os.ReadDir(cliDir)
		if err != nil {
			t.Fatalf("read dir %s: %v", cliDir, err)
		}
		want := map[string]bool{
			"init.go":                 false,
			"update_template_sync.go": false,
		}
		for _, entry := range entries {
			name := entry.Name()
			if entry.IsDir() || filepath.Ext(name) != ".go" || filepath.Base(name) == "hook_install_precommit.go" {
				continue
			}
			// Only non-test production files are call sites; test files are
			// excluded because they exercise the wrapper directly.
			if len(name) > 8 && name[len(name)-8:] == "_test.go" {
				continue
			}
			path := filepath.Join(cliDir, name)
			calls := installOptionalCallsIn(t, parseCallerSource(t, path))
			if len(calls) == 0 {
				continue
			}
			if _, known := want[name]; !known {
				t.Errorf("unexpected installPreCommitHookOptional call site in %s (spec.md §C.4 permits exactly init.go and update_template_sync.go)", name)
				continue
			}
			want[name] = true
		}
		for name, seen := range want {
			if !seen {
				t.Errorf("expected an installPreCommitHookOptional call in %s, found none", name)
			}
		}
	})
}

// parseCallerSource parses one Go source file. A parse failure is fatal: the
// file belongs to this package, so a tree that runs this test has the file and
// it must parse.
func parseCallerSource(t *testing.T, path string) *ast.File {
	t.Helper()
	f, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return f
}

// installOptionalCallsIn returns every call to installPreCommitHookOptional.
func installOptionalCallsIn(t *testing.T, f *ast.File) []*ast.CallExpr {
	t.Helper()
	var calls []*ast.CallExpr
	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if ident, ok := call.Fun.(*ast.Ident); ok && ident.Name == "installPreCommitHookOptional" {
			calls = append(calls, call)
		}
		return true
	})
	return calls
}

// enclosingFuncDecl returns the function declaration whose body contains call,
// or nil. Closures inside the declaration count as inside it: the identifiers
// the call reads may be captured from the declaration's scope.
func enclosingFuncDecl(f *ast.File, call *ast.CallExpr) *ast.FuncDecl {
	var found *ast.FuncDecl
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			if c, ok := n.(*ast.CallExpr); ok && c == call {
				found = fn
			}
			return true
		})
	}
	return found
}

// isCmdCall reports whether expr is the zero-argument call <recv>.<method>().
func isCmdCall(expr ast.Expr, recv, method string) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok || len(call.Args) != 0 {
		return false
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	recvIdent, ok := sel.X.(*ast.Ident)
	return ok && recvIdent.Name == recv && sel.Sel.Name == method
}

// assertAssignedFromCmdCall asserts that name is assigned exactly once inside
// fn's body and that the assignment is the zero-argument call cmd.<method>().
// Assignment counting (rather than resolution) is enough here: an identically
// named shadow inside a nested closure would still leave exactly one
// assignment, and the call's argument has been separately asserted to be the
// identifier itself.
func assertAssignedFromCmdCall(t *testing.T, fn *ast.FuncDecl, name, method string) {
	t.Helper()
	defs := 0
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if !ok || len(assign.Lhs) != 1 {
			return true
		}
		ident, ok := assign.Lhs[0].(*ast.Ident)
		if !ok || ident.Name != name {
			return true
		}
		defs++
		if len(assign.Rhs) != 1 {
			t.Errorf("%s: %s must be assigned from a single call expression", fn.Name.Name, name)
			return true
		}
		if !isCmdCall(assign.Rhs[0], "cmd", method) {
			t.Errorf("%s: %s must be assigned cmd.%s(), got %s", fn.Name.Name, name, method, types.ExprString(assign.Rhs[0]))
		}
		return true
	})
	if defs != 1 {
		t.Errorf("%s: expected exactly one assignment of %s inside the function, found %d", fn.Name.Name, name, defs)
	}
}

// runOptionalOnUserModifiedHook constructs the AC-PCP-003 scenario — a
// marker-bearing hook with no provenance record, so the legacy path classifies
// the difference as user-modified (REQ-PCP-005) — runs
// installPreCommitHookOptional over captured writers, and returns them.
func runOptionalOnUserModifiedHook(t *testing.T, root string) (out, warn *bytes.Buffer) {
	t.Helper()
	writeExistingHook(t, root, previousPreCommitHookContent)
	out, warn = &bytes.Buffer{}, &bytes.Buffer{}
	installPreCommitHookOptional(root, false, out, warn)
	return out, warn
}

// TestPreCommitBackupNoticeContent — AC-PCP-004 clause (i) (REQ-PCP-004).
//
// The warning writer must carry both notice elements — the backup file's path
// and a statement that the hook was replaced — and nothing else. The progress
// writer keeps the plain success line and never carries the notice: under
// `moai update` it is stdout, where a redirected run would swallow the
// data-loss notice whole.
func TestPreCommitBackupNoticeContent(t *testing.T) {
	root := newPreCommitTestRepo(t)
	out, warn := runOptionalOnUserModifiedHook(t, root)

	backups := findBackups(t, root)
	if len(backups) != 1 {
		t.Fatalf("expected exactly one backup, found %d: %v", len(backups), backups)
	}
	wantNotice := fmt.Sprintf("  Warning: user-modified pre-commit hook was replaced; previous hook backed up at %s\n", backups[0])
	if got := warn.String(); got != wantNotice {
		t.Errorf("warning writer = %q,\nwant exactly %q (the backup path and the replacement statement, and nothing else)", got, wantNotice)
	}
	if strings.Contains(warn.String(), "pre-commit.local") {
		t.Errorf("the notice must not name pre-commit.local — a recovery path the installed hook never reads (REQ-PCP-004)")
	}
	const wantSuccess = "  Pre-commit hook installed (.git/hooks/pre-commit)\n"
	if got := out.String(); got != wantSuccess {
		t.Errorf("progress writer = %q, want exactly %q (the notice must not ride the progress writer)", got, wantSuccess)
	}
}

// TestPreCommitNoSilentReplacement — AC-PCP-006 (REQ-PCP-006).
//
// ONE case asserting both artifacts of the same run: the backup file AND the
// notice naming it. Neither alone satisfies the criterion — a mutant that
// writes the backup but swallows the notice (or vice versa) passes any test
// that checks them separately.
func TestPreCommitNoSilentReplacement(t *testing.T) {
	root := newPreCommitTestRepo(t)
	out, warn := runOptionalOnUserModifiedHook(t, root)

	backups := findBackups(t, root)
	if len(backups) != 1 {
		t.Fatalf("a replaced user-modified hook must leave a backup file; found %d: %v", len(backups), backups)
	}
	if !strings.Contains(warn.String(), backups[0]) {
		t.Errorf("the same run must also emit a notice naming the backup %s; warning writer = %q", backups[0], warn.String())
	}
	if !strings.Contains(out.String(), "Pre-commit hook installed") {
		t.Errorf("the run completed; expected the success line on the progress writer, got %q", out.String())
	}
}

// TestPreCommitBackupOutputNotBareSuccess — AC-PCP-007 (REQ-PCP-007).
//
// Clause (i): the warning writer's captured output is non-empty, so a
// whitespace-padded success line with an empty warning writer fails here.
// Clause (ii): the output contains the backup path, so a generic blank or
// vague line that names nothing also fails.
func TestPreCommitBackupOutputNotBareSuccess(t *testing.T) {
	root := newPreCommitTestRepo(t)
	_, warn := runOptionalOnUserModifiedHook(t, root)

	if warn.Len() == 0 {
		t.Errorf("clause (i): the warning writer must be non-empty on a backup run — the run's total output must never be the bare success line alone")
	}
	backups := findBackups(t, root)
	if len(backups) != 1 {
		t.Fatalf("expected one backup, found %d: %v", len(backups), backups)
	}
	if !strings.Contains(warn.String(), backups[0]) {
		t.Errorf("clause (ii): the warning output must contain the backup path %s, got %q", backups[0], warn.String())
	}
}

// TestPreCommitBackupNoClobber — AC-PCP-009 (REQ-PCP-009).
//
// Two backups are forced into the same timestamp via the installer's clock
// seam, so run 2's chosen path is exactly run 1's occupied path. The
// pre-existing backup's bytes must be unchanged AND a second, distinctly-named
// backup must exist — the second clause defeats the mutant that simply skips
// the backup when the path is occupied.
func TestPreCommitBackupNoClobber(t *testing.T) {
	fixed := time.Date(2026, 8, 25, 12, 0, 0, 0, time.UTC)
	root := newPreCommitTestRepo(t)

	const firstPatch = previousPreCommitHookContent + "# first local patch\n"
	writeExistingHook(t, root, firstPatch)
	one := NewPreCommitInstaller(root)
	one.now = func() time.Time { return fixed }
	if err := one.InstallPreCommitHook(false); err != nil {
		t.Fatalf("install 1: %v", err)
	}

	// Run 2 at the same clock reading: the hook is user-modified again (it
	// differs from both the record and the incoming content), and the backup
	// path this run would choose is the one run 1 already filled.
	const secondPatch = previousPreCommitHookContent + "# second local patch\n"
	writeExistingHook(t, root, secondPatch)
	two := NewPreCommitInstaller(root)
	two.now = func() time.Time { return fixed }
	if err := two.InstallPreCommitHook(false); err != nil {
		t.Fatalf("install 2: %v", err)
	}

	backups := findBackups(t, root)
	if len(backups) != 2 {
		t.Fatalf("expected two distinct backups, found %d: %v", len(backups), backups)
	}
	if backups[0] == backups[1] {
		t.Fatalf("backups must be distinctly named, both %q", backups[0])
	}
	var haveFirst, haveSecond bool
	for _, b := range backups {
		raw, err := os.ReadFile(b)
		if err != nil {
			t.Fatalf("read backup %s: %v", b, err)
		}
		switch string(raw) {
		case firstPatch:
			haveFirst = true // run 1's backup is byte-unchanged, not clobbered
		case secondPatch:
			haveSecond = true
		}
	}
	if !haveFirst || !haveSecond {
		t.Fatalf("expected one backup holding each pre-run body (old unchanged + new distinct); backups: %v", backups)
	}
}

// TestPreCommitSupportWriteFailureNonFatal — AC-PCP-010 (REQ-PCP-010), both
// sub-cases mandatory. They differ in post-state because the two supporting
// writes sit on opposite sides of the hook write.
func TestPreCommitSupportWriteFailureNonFatal(t *testing.T) {
	t.Run("a_backup_write_fails_hook_left_untouched", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("chmod-based unwritable-dir simulation is POSIX-only")
		}
		if os.Geteuid() == 0 {
			t.Skip("running as root bypasses directory permission checks")
		}
		root := newPreCommitTestRepo(t)
		writeExistingHook(t, root, previousPreCommitHookContent)
		writeRecord(t, root, digestOf("something else MoAI once wrote\n"))
		hooksDir := filepath.Join(root, ".git", "hooks")
		t.Cleanup(func() { _ = os.Chmod(hooksDir, 0o755) })
		if err := os.Chmod(hooksDir, 0o500); err != nil {
			t.Fatalf("chmod hooks dir read-only: %v", err)
		}

		var out, warn bytes.Buffer
		installPreCommitHookOptional(root, false, &out, &warn) // MUST NOT panic or abort the caller

		if !strings.Contains(warn.String(), "Warning") || !strings.Contains(warn.String(), "backup") {
			t.Errorf("expected a warning naming the failed backup, got: %q", warn.String())
		}
		// POST-STATE — the clause that defeats warn-then-overwrite-anyway: a
		// replacement without a recoverable backup is exactly the loss this
		// SPEC exists to prevent, so the hook must be byte-identical to its
		// pre-run value.
		if got := readHook(t, root); got != previousPreCommitHookContent {
			t.Errorf("POST-STATE violated: hook was replaced despite the failed backup (%d bytes, want the unchanged %d-byte pre-run body)", len(got), len(previousPreCommitHookContent))
		}
		assertNoBackup(t, root)
	})

	t.Run("b_provenance_write_fails_replacement_stays", func(t *testing.T) {
		root := newPreCommitTestRepo(t)
		writeExistingHook(t, root, previousPreCommitHookContent)
		// A directory at the record path makes the provenance write fail on
		// every platform while the hook write (a different file) succeeds, and
		// leaves the record unusable — which is also run 2's self-heal
		// precondition.
		if err := os.MkdirAll(precommitRecordPath(root), 0o755); err != nil {
			t.Fatalf("mkdir record path: %v", err)
		}

		var out1, warn1 bytes.Buffer
		installPreCommitHookOptional(root, false, &out1, &warn1) // MUST NOT panic or abort the caller

		if got := readHook(t, root); got != preCommitHookContent {
			t.Fatalf("the hook must BE replaced when only the post-write provenance write fails; got %d bytes", len(got))
		}
		if !strings.Contains(warn1.String(), "provenance") {
			t.Errorf("expected a warning naming the failed provenance write, got: %q", warn1.String())
		}
		backupsAfterRun1 := findBackups(t, root)
		if len(backupsAfterRun1) != 1 {
			t.Fatalf("run 1 is user-modified: expected one backup, found %d: %v", len(backupsAfterRun1), backupsAfterRun1)
		}

		// Run 2 with the same content: installed == incoming with no usable
		// record, so the run is a quiet non-event — no second backup, no
		// replacement notice. (The provenance warning may repeat: the record
		// is still unwritable, and that is a different, correct warning.)
		var out2, warn2 bytes.Buffer
		installPreCommitHookOptional(root, false, &out2, &warn2)

		if got := findBackups(t, root); len(got) != len(backupsAfterRun1) {
			t.Errorf("run 2 must take no backup: found %d backups, want %d", len(got), len(backupsAfterRun1))
		}
		if strings.Contains(warn2.String(), "was replaced") || strings.Contains(warn2.String(), backupsAfterRun1[0]) {
			t.Errorf("run 2 must emit no replacement notice, got: %q", warn2.String())
		}
		const wantSuccess = "  Pre-commit hook installed (.git/hooks/pre-commit)\n"
		if out2.String() != wantSuccess {
			t.Errorf("run 2 progress output = %q, want exactly %q", out2.String(), wantSuccess)
		}
	})
}
