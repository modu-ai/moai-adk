package cli

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"
)

// AC-TCL-024 inventory_complete (SPEC-TUI-CHECKLIST-001 design.md §3.3.1): the
// fourteen search shapes are run in Go over internal/cli/**/*.go (tests
// excluded) and every hit is classified by (file, message fragment) — never by
// line number and never file-wide for an orbit file. Two directions are
// judged: (a) an unclassified hit fails, (b) a preserved row whose fragment
// vanished from its file fails. The row ids below are the ids of the design
// table (P = preserved row in this card's scope, X = excluded row).

// inventoryShapes are the 12 case-insensitive shapes of the first search plus
// the 2 case-sensitive shapes of the second one.
var (
	inventoryShapesFold  = []string{"SymWarning()", "SymError()", `CheckLine("warn"`, `CheckLine("err"`, "warnLine(", "Fprintf(errOut", "Fprintln(errOut", "Fprint(errOut", "Fprintf(warn,", "OrStderr(),", "sevNote,", "warning:"}
	inventoryShapesExact = []string{"PillWarn", "PillErr"}
)

// inventoryHit is one search hit. window is the hit line plus, for a call that
// continues on the next lines, those lines (the message often sits there).
type inventoryHit struct {
	file   string // path relative to internal/cli, slash form
	line   int
	text   string
	window string
}

// inventoryRow classifies the hits of one file whose window contains frag.
type inventoryRow struct {
	id   string
	file string
	frag string
}

// inventoryOrbit are the 16 orbit file definitions (design.md §3.3.1); an orbit
// file's hits may only be classified at fragment level, never file-wide.
var inventoryOrbit = []struct {
	name string
	re   *regexp.Regexp
}{
	{"orbit_init_files", regexp.MustCompile(`^init[^/]*\.go$`)},
	{"orbit_update_files", regexp.MustCompile(`^update[^/]*\.go$`)},
	{"orbit_update_pkg", regexp.MustCompile(`^update/`)},
	{"orbit_checklist_files", regexp.MustCompile(`^checklist[^/]*\.go$`)},
	{"orbit_hook_install", regexp.MustCompile(`^hook_install\.go$`)},
	{"orbit_hook_install_precommit", regexp.MustCompile(`^hook_install_precommit\.go$`)},
	{"orbit_worktree_advisory", regexp.MustCompile(`^worktree_advisory\.go$`)},
	{"orbit_feedback_participation", regexp.MustCompile(`^feedback_participation\.go$`)},
	{"orbit_migrate_project_assets", regexp.MustCompile(`^migrate_project_assets\.go$`)},
	{"orbit_migrate_local_instructions", regexp.MustCompile(`^migrate_local_instructions\.go$`)},
	{"orbit_session_worktree", regexp.MustCompile(`^session_worktree\.go$`)},
	{"orbit_mirror_notice", regexp.MustCompile(`^mirror_notice\.go$`)},
	{"orbit_user_asset_phase", regexp.MustCompile(`^user_asset_phase\.go$`)},
	{"orbit_uikit_styles", regexp.MustCompile(`^uikit/styles\.go$`)},
	{"orbit_printer", regexp.MustCompile(`^printer/printer\.go$`)},
	{"orbit_severity_line", regexp.MustCompile(`^severity_line\.go$`)},
}

func inventoryIsOrbit(file string) bool {
	for _, o := range inventoryOrbit {
		if o.re.MatchString(file) {
			return true
		}
	}
	return false
}

// inventoryRows is the classification table. P rows are in this card's scope;
// X rows are matches that are not preserved rows (the reason is in design.md).
// X11 rows are update-side fragments deferred to card t1978. X12 is the one row
// this card adds: the checklist session layer's own code matches three shapes
// by name (stream plumbing, helper names), writes no line, and did not exist
// when the design table was measured.
var inventoryRows = []inventoryRow{
	// P23 — resume-path warnings (init_resume.go) and X2 pass-throughs.
	{"P23", "init_resume.go", "Warning: global settings env update failed"},
	{"X2", "init_resume.go", "installPrePushHookOptional("},
	{"P23", "init_resume.go", "Pre-push hook installation failed"},
	{"X2", "init_resume.go", "installPreCommitHookOptional("},
	{"P23", "init_resume.go", "Pre-commit hook installation failed"},
	{"P23", "init_resume.go", "Warning: manifest retrack (config sections) failed"},
	// P20 / P21 — preserved user hooks.
	{"P20", "hook_install.go", "existing pre-push hook preserved"},
	{"P21", "hook_install_precommit.go", "existing pre-commit hook preserved"},
	// P5 — worktree advisory.
	{"P5", "worktree_advisory.go", "worktreeAdvisoryText(projectRoot)"},
	// P1 / P2 — snapshot warnings through warnLine; X7 the declaration.
	{"P1", "update/backup/file_snapshot.go", "s.writeFailedPrefix"},
	{"P2", "update/backup/file_snapshot.go", "s.promoteFailedPrefix"},
	{"X7", "update/backup/settings_snapshot.go", "func warnLine("},
	{"P1", "update/backup/settings_snapshot.go", `Fprintf(warn, "%s\n", line)`},
	// X1 — warning signal path.
	{"X1", "init.go", "Could not record the harness selection for the resume"},
	{"X1", "init.go", "Could not record the init resume marker"},
	{"X1", "printer/printer.go", `p.status("warn", p.theme.Warning`},
	// X2 — writers passed through.
	{"X2", "init.go", "installPrePushHookOptional("},
	{"X2", "init.go", "installPreCommitHookOptional("},
	{"X2", "update.go", "ErrOut:"},
	// X4 — before the first item, or replaced by the summary.
	{"X4", "init.go", "session-worktree chdir"},
	{"X4", "init.go", "Initialization cancelled."},
	{"X4", "checklist_init.go", "buildInitSuccessCard"},
	// X12 — this card's session layer (see the table comment).
	{"X12", "checklist_session.go", "procErr"},
	{"X12", "checklist_session.go", "severityWarnLine("},
	// X3 — written outside template sync.
	{"X3", "update.go", `CheckLine("warn", "Update checker"`},
	{"X3", "update.go", `CheckLine("warn", "Worktree migration preview"`},
	{"X3", "update.go", `CheckLine("warn", "Reconciliation preview"`},
	{"X3", "update.go", `CheckLine("warn", "Deny-rule migration"`},
	{"X3", "update.go", `CheckLine("warn", "v2 detection"`},
	{"X3", "update.go", `CheckLine("warn", "Participation ask"`},
	{"X3", "update.go", "user-asset phase failed"},
	{"X3", "update_integrity_probe.go", `CheckLine("warn", "Integrity"`},
	{"X3", "update_dryrun_preview.go", "[dry-run]"},
	{"X3", "update_archive.go", "[dry-run] archive:"},
	{"X3", "migrate_project_assets.go", "migrationFinalStateSentinel"},
	{"X3", "feedback_participation.go", "warn: participation"},
	// X5 — clean-install, binary update and re-exec.
	{"X5", "update_clean_install.go", "[clean-reinstall]"},
	{"X5", "update_version.go", `CheckLine("warn", "Re-exec"`},
	// X6 — the warning panel that is not drawn on an active run.
	{"X6", "init_warnings.go", "Fprintln(errOut)"},
	{"X6", "init_warnings.go", "WarnStyle.Render(b.String())"},
	// X7 — declarations, ledger rows, glyph definitions, comments.
	{"X7", "update_participation.go", "here as a warning"},
	{"X7", "severity_line.go", "Warning: / warning: / hint:"},
	{"X7", "update_action_block.go", "sev: sevNote"},
	{"X7", "update.go", "CLAUDE_DISABLE_PATH_WARNING"},
	{"X7", "update_snapshot_hook.go", `"Warning: " prefix`},
	{"X7", "uikit/styles.go", "func SymError()"},
	{"X7", "uikit/styles.go", "func SymWarning()"},
	{"X7", "printer/printer.go", `// Warn writes a "Warning: ..."`},
	{"X7", "update_noise.go", `emit legacy "Warning: 3-way merge`},
	{"X7", "update_noise.go", "Card t1527 D4: ! severity line replaces"},
	{"X7", "update_manifest_heal.go", `the ! severity line replaces the raw "warning:"`},
	{"X7", "update_tux.go", `PillErr "! N conflict"`},
	// X8 — status notice kept out of the fold, and the failure-path recovery line.
	{"X8", "update_model_key_strip.go", "retained key %q absent"},
	{"X8", "update_recovery_manifest.go", "Warning: could not write"},
	// X10 — writes to the process stderr, outside the command writers.
	{"X10", "update/backup/backup.go", "could not save template defaults"},
	{"X10", "update/backup/snapshot.go", "Warning: snapshot "},
	{"X10", "update/backup/restore.go", "skipping symlink backup entry"},
	{"X10", "update/backup/restore.go", "escaping config dir"},
	{"X10", "update/backup/restore.go", "merge failed for"},
	// X11 — update-side fragments, deferred to card t1978 (P6..P19 pieces).
	{"X11", "update/merge/merge.go", "merge failed, preserving user version"},
	{"X11", "update/merge/merge.go", "merged with conflicts (user version preferred)"},
	{"X11", "update_noise.go", "'moai update -c' to resync templates for"},
	{"X11", "update_template_sync_guard.go", "template tree not recorded"},
	{"X11", "update_template_sync_guard.go", "template stamp disagrees with deployed bytes"},
	{"X11", "update_tux.go", "SymWarning(), len(kept)"},
	{"X11", "update_tux.go", "PillErr, Label"},
	{"X11", "update_archive.go", "archive drift backup:"},
	{"X11", "update_archive.go", `"Legacy skill archive"`},
	{"X11", "update_template_sync.go", "env.PATH refresh failed"},
	{"X11", "update_template_sync.go", "manifest save after deploy"},
	{"X11", "update_template_sync.go", "migration warning"},
	{"X11", "update_template_sync.go", "deployment_mode re-assert warning"},
	{"X11", "update_template_sync.go", "manifest retrack ("},
	{"X11", "update_migrate.go", "deployment_mode write warning"},
	{"X11", "update_wizard.go", "Project not initialized"},
	{"X11", "update_wizard.go", "GitHub token was not saved"},
	{"X11", "update_wizard.go", "GitLab token was not saved"},
}

// inventoryPreserved are the preserved rows in this card's scope: each fragment
// must still be present in its file (direction (b)). P3 and P4 are enumerated
// rows without a search hit; P17's line moved behind a helper but its message
// stays in the file.
var inventoryPreserved = []inventoryRow{
	{"P1", "update/backup/file_snapshot.go", "warnLine(warn, s.writeFailedPrefix"},
	{"P1", "update/backup/settings_snapshot.go", `SettingsSnapshotWriteFailedPrefix = "settings-snapshot-write-failed:"`},
	{"P1", "update/backup/mcp_snapshot.go", `MCPSnapshotWriteFailedPrefix = "mcp-snapshot-write-failed:"`},
	{"P2", "update/backup/file_snapshot.go", "warnLine(warn, s.promoteFailedPrefix"},
	{"P2", "update/backup/settings_snapshot.go", `SettingsSnapshotPromoteFailedPrefix = "settings-snapshot-promote-failed:"`},
	{"P2", "update/backup/mcp_snapshot.go", `MCPSnapshotPromoteFailedPrefix = "mcp-snapshot-promote-failed:"`},
	{"P3", "init.go", "Initialized MoAI project (resumed"},
	{"P4", "init_update_notice.go", "A newer moai binary is available"},
	{"P5", "worktree_advisory.go", "this checkout is shared across concurrent sessions;"},
	{"P17", "update_manifest_retrack.go", "manifest retrack skip"},
	{"P20", "hook_install.go", "existing pre-push hook preserved"},
	{"P21", "hook_install_precommit.go", "existing pre-commit hook preserved"},
	{"P23", "init_resume.go", "Warning: global settings env update failed"},
	{"P23", "init_resume.go", "Pre-push hook installation failed"},
	{"P23", "init_resume.go", "Pre-commit hook installation failed"},
	{"P23", "init_resume.go", "Warning: manifest retrack (config sections) failed"},
}

// inventoryRoot returns the absolute internal/cli directory of this source tree.
func inventoryRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve the test source path")
	}
	return filepath.Dir(thisFile)
}

func inventoryLineMatches(line string) bool {
	lower := strings.ToLower(line)
	for _, s := range inventoryShapesFold {
		if strings.Contains(lower, strings.ToLower(s)) {
			return true
		}
	}
	for _, s := range inventoryShapesExact {
		if strings.Contains(line, s) {
			return true
		}
	}
	return false
}

// inventoryScan walks internal/cli (tests excluded) and returns every hit and
// the slash-form list of files it walked.
func inventoryScan(t *testing.T, root string) (hits []inventoryHit, walked []string) {
	t.Helper()
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		rel = filepath.ToSlash(rel)
		walked = append(walked, rel)
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		lines := strings.Split(string(data), "\n")
		for i, line := range lines {
			if !inventoryLineMatches(line) {
				continue
			}
			window := line
			if tr := strings.TrimSpace(line); strings.HasSuffix(tr, ",") || strings.HasSuffix(tr, "(") || strings.HasSuffix(tr, "+") {
				for j := i + 1; j < len(lines) && j <= i+2; j++ {
					window += "\n" + lines[j]
				}
			}
			hits = append(hits, inventoryHit{file: rel, line: i + 1, text: line, window: window})
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	sort.Strings(walked)
	return hits, walked
}

// inventoryClassify returns the row id for a hit, or "" when it is unclassified.
// A file outside the orbit is classified file-wide (X9); an orbit file only by
// a fragment row.
func inventoryClassify(h inventoryHit) string {
	if !inventoryIsOrbit(h.file) {
		return "X9"
	}
	for _, r := range inventoryRows {
		if r.file == h.file && strings.Contains(h.window, r.frag) {
			return r.id
		}
	}
	return ""
}

// checklistInventoryComplete is TestChecklistPreservedLinesExactlyOnce/inventory_complete.
func checklistInventoryComplete(t *testing.T) {
	root := inventoryRoot(t)
	hits, walked := inventoryScan(t, root)

	// The sweep must be real: an empty swept set asserts nothing.
	if len(walked) < 100 || len(hits) < 100 {
		t.Fatalf("the sweep is too small to mean anything: %d files, %d hits", len(walked), len(hits))
	}
	// Positive control: the shapes fire on the known preserved sites.
	for file, min := range map[string]int{"init_resume.go": 4, "update_template_sync.go": 7, "update/backup/restore.go": 6} {
		n := 0
		for _, h := range hits {
			if h.file == file {
				n++
			}
		}
		if n < min {
			t.Errorf("positive control: %s has %d hits, want at least %d (the search shapes no longer fire)", file, n, min)
		}
	}

	{
		for _, r := range inventoryRows {
			if len(strings.TrimSpace(r.frag)) < 4 {
				t.Errorf("row %s for %s has a fragment too short to be a message fragment: %q", r.id, r.file, r.frag)
			}
			if !inventoryIsOrbit(r.file) {
				t.Errorf("row %s names %s, which is outside the orbit; non-orbit files are classified X9 by definition, not by a row", r.id, r.file)
			}
		}
		// Direction (a): no unclassified hit.
		for _, h := range hits {
			if inventoryClassify(h) == "" {
				t.Errorf("unclassified hit %s:%d: %s", h.file, h.line, strings.TrimSpace(h.text))
			}
		}
		// Direction (b): every in-scope preserved fragment is still in its file.
		for _, r := range inventoryPreserved {
			data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(r.file)))
			if err != nil {
				t.Errorf("row %s: %v", r.id, err)
				continue
			}
			if !strings.Contains(string(data), r.frag) {
				t.Errorf("row %s: the fragment %q vanished from %s", r.id, r.frag, r.file)
			}
		}
		// Negative controls: the classifier is not vacuous. A new line in an orbit
		// file is unclassified, and an orbit file is never swallowed file-wide.
		for _, file := range []string{"init.go", "update.go", "hook_install.go", "update_template_sync.go", "checklist_session.go"} {
			synth := inventoryHit{file: file, line: 1, text: `fmt.Fprintln(errOut, "brand new warning")`}
			synth.window = synth.text
			if got := inventoryClassify(synth); got != "" {
				t.Errorf("a synthetic new hit in orbit file %s was classified %q; it must be unclassified", file, got)
			}
		}
		synthNonOrbit := inventoryHit{file: "loop.go", text: `fmt.Fprintln(errOut, "x")`, window: `fmt.Fprintln(errOut, "x")`}
		if got := inventoryClassify(synthNonOrbit); got != "X9" {
			t.Errorf("a non-orbit hit is classified %q, want X9", got)
		}
	}

	// One subtest per orbit definition: the group must have walked files (an
	// empty group asserts nothing), and each of its hits is classified by a
	// fragment row of its own file — never by the file-wide X9.
	for _, o := range inventoryOrbit {
		t.Run(o.name, func(t *testing.T) {
			files := 0
			for _, w := range walked {
				if o.re.MatchString(w) {
					files++
				}
			}
			if files == 0 {
				t.Fatalf("%s matched no file under internal/cli: the orbit definition sweeps nothing", o.name)
			}
			for _, h := range hits {
				if !o.re.MatchString(h.file) {
					continue
				}
				if id := inventoryClassify(h); id == "" || id == "X9" {
					t.Errorf("%s hit %s:%d is classified %q; an orbit hit needs a fragment row: %s", o.name, h.file, h.line, id, strings.TrimSpace(h.text))
				}
			}
		})
	}
}
