package cli

// codex_launcher_readout_test.go — SPEC-CODEX-LAUNCHER-001 M3 command-level
// legs of the readout ACs (the M2 assembly-level legs live in
// codex_readiness_test.go; here the SAME rows must flow VERBATIM through the
// bare/status command forms):
//
//   - AC-CL-004 — the ten informational cells, sentinel rows at the command
//     surface, both forms reporting identical values
//   - AC-CL-006 — rc 0, stdout-only, stderr 0 bytes, banned-word hits 0,
//     action phrase present on all five incomplete states and absent wired
//   - AC-CL-011 — binary-absent readout cells (label closed set), the launch
//     verbs' single-line install diagnostic
//   - AC-CL-012 — no-write snapshot across all four forms under an isolated
//     home; a nonexistent CODEX_HOME stays nonexistent

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// codexWiringFixture lays out one wiring state under a fresh project root and
// returns the root. The five incomplete states plus the wired one mirror
// AC-CL-004's table (M2's SixStateMatrix arrangements, command-level).
func codexWiringFixture(t *testing.T, state string) string {
	t.Helper()
	validHooks := "{}\n"
	rogueHooks := `{"rogue_key": 1}` + "\n"
	configBody := "model = \"g\"\n"
	root := t.TempDir()
	switch state {
	case "absent":
	case "empty dir":
		if err := os.MkdirAll(filepath.Join(root, ".codex"), 0o755); err != nil {
			t.Fatal(err)
		}
	case "hooks only":
		writeCodexWiringFile(t, root, ".codex/hooks.json", validHooks)
	case "config only":
		writeCodexWiringFile(t, root, ".codex/config.toml", configBody)
	case "invalid hooks":
		writeCodexWiringFile(t, root, ".codex/hooks.json", rogueHooks)
		writeCodexWiringFile(t, root, ".codex/config.toml", configBody)
	case "wired":
		writeCodexWiringFile(t, root, ".codex/hooks.json", validHooks)
		writeCodexWiringFile(t, root, ".codex/config.toml", configBody)
	default:
		t.Fatalf("unknown fixture state %q", state)
	}
	return root
}

// codexBannedWords are the six informational-vocabulary probes of AC-CL-006
// (case-insensitive; hit count must be 0 across the whole readout).
func codexBannedWordHits(text string) int {
	low := strings.ToLower(text)
	n := 0
	for _, w := range []string{"error", "failed", "fatal", "broken", "cannot", "unable"} {
		n += strings.Count(low, w)
	}
	return n
}

// ─── AC-CL-004 / AC-CL-006 — the ten informational cells ──────────────────

// TestCodexReadout_CommandTenCells runs the five incomplete wiring states in
// BOTH readout forms (10 cells): rc 0, the wiring row equal to the M2 named
// constants (verbatim flow), the action phrase present, stdout-only output,
// stderr exactly 0 bytes, banned-word hits 0 — and both forms reporting the
// SAME six values.
func TestCodexReadout_CommandTenCells(t *testing.T) {
	states := []struct {
		name    string
		wantRow string
	}{
		{"absent", wantWiringRowNotWired},
		{"empty dir", wantWiringRowNotWired},
		{"hooks only", wantWiringRowPartialNoConfig},
		{"config only", wantWiringRowPartialNoHooks},
		{"invalid hooks", wantWiringRowInvalid},
	}
	for _, st := range states {
		for _, form := range []string{"bare", "status"} {
			t.Run(st.name+"/"+form, func(t *testing.T) {
				t.Setenv(codexHomeEnvVar, t.TempDir())
				withCodexSetupProbe(t, minimalProbeStub())
				root := codexWiringFixture(t, st.name)
				withCodexProjectDir(t, root)

				var args []string
				if form == "status" {
					args = []string{"status"}
				}
				stdout, stderr, err := runCodexCmd(t, args...)
				if err != nil {
					t.Fatalf("%s form: err = %v, want rc 0", form, err)
				}
				if stderr != "" {
					t.Errorf("%s form wrote %d bytes to stderr (%q), want 0", form, len(stderr), stderr)
				}
				lines := strings.Split(strings.TrimSuffix(stdout, "\n"), "\n")
				if len(lines) != 6 {
					t.Fatalf("readout lines = %d, want 6 (full text %q)", len(lines), stdout)
				}
				if lines[3] != st.wantRow {
					t.Errorf("wiring row = %q, want M2 constant %q (verbatim flow)", lines[3], st.wantRow)
				}
				if !strings.Contains(lines[3], codexWiringAction) {
					t.Errorf("wiring row missing action phrase: %q", lines[3])
				}
				if !strings.Contains(stdout, "moai init --agent codex") {
					t.Errorf("readout missing the remediation action")
				}
				if hits := codexBannedWordHits(stdout); hits != 0 {
					t.Errorf("banned-word hits = %d in %q", hits, stdout)
				}
			})
		}
	}

	// Both forms report identical values cell-for-cell (AC-CL-004).
	t.Run("both forms identical", func(t *testing.T) {
		t.Setenv(codexHomeEnvVar, t.TempDir())
		withCodexSetupProbe(t, minimalProbeStub())
		withCodexProjectDir(t, codexWiringFixture(t, "hooks only"))

		bare, _, err := runCodexCmd(t)
		if err != nil {
			t.Fatalf("bare: %v", err)
		}
		status, _, err := runCodexCmd(t, "status")
		if err != nil {
			t.Fatalf("status: %v", err)
		}
		if bare != status {
			t.Errorf("bare and status forms differ:\n bare   = %q\n status = %q", bare, status)
		}
	})
}

// TestCodexReadout_WiredStateOmitsAction — on the wired state the action
// phrase is ABSENT (no remediation where nothing needs remedying).
func TestCodexReadout_WiredStateOmitsAction(t *testing.T) {
	t.Setenv(codexHomeEnvVar, t.TempDir())
	withCodexSetupProbe(t, minimalProbeStub())
	withCodexProjectDir(t, codexWiringFixture(t, "wired"))

	stdout, _, err := runCodexCmd(t)
	if err != nil {
		t.Fatalf("bare: %v", err)
	}
	lines := strings.Split(strings.TrimSuffix(stdout, "\n"), "\n")
	if len(lines) != 6 || lines[3] != wantWiringRowWired {
		t.Fatalf("wiring row = %q (lines %d), want %q", lines, len(lines), wantWiringRowWired)
	}
	if strings.Contains(stdout, "moai init --agent codex") {
		t.Errorf("wired readout recommends the action: %q", stdout)
	}
}

// TestCodexReadout_SentinelRowsAtCommandSurface — the three probe-supplied
// rows (binary path, version, auth) surface EXACTLY (the M2 named constants)
// in both command forms: a launcher printing "unknown" regardless of the
// probe dies here.
func TestCodexReadout_SentinelRowsAtCommandSurface(t *testing.T) {
	t.Setenv(codexHomeEnvVar, t.TempDir())
	withCodexSetupProbe(t, CodexSetupResult{
		Installed:    true,
		Binary:       sentinelCodexBinaryPath,
		Version:      sentinelCodexVersion,
		AuthProvider: sentinelCodexAuth,
	})
	withCodexProjectDir(t, codexWiringFixture(t, "wired"))

	for _, form := range []string{"bare", "status"} {
		t.Run(form, func(t *testing.T) {
			var args []string
			if form == "status" {
				args = []string{"status"}
			}
			stdout, _, err := runCodexCmd(t, args...)
			if err != nil {
				t.Fatalf("%s: %v", form, err)
			}
			lines := strings.Split(strings.TrimSuffix(stdout, "\n"), "\n")
			if len(lines) != 6 {
				t.Fatalf("lines = %d, want 6", len(lines))
			}
			if lines[0] != wantCodexRowSentinel {
				t.Errorf("binary row = %q, want %q", lines[0], wantCodexRowSentinel)
			}
			if lines[2] != wantAuthRowSentinel {
				t.Errorf("auth row = %q, want %q", lines[2], wantAuthRowSentinel)
			}
		})
	}
}

// ─── AC-CL-011 — binary absent ─────────────────────────────────────────────

// codexReadoutLabels extracts the row-label set from a readout (first token
// of each of the six lines).
func codexReadoutLabels(t *testing.T, stdout string) []string {
	t.Helper()
	lines := strings.Split(strings.TrimSuffix(stdout, "\n"), "\n")
	if len(lines) != 6 {
		t.Fatalf("readout lines = %d, want 6 (%q)", len(lines), stdout)
	}
	labels := make([]string, 0, 6)
	for _, ln := range lines {
		fields := strings.Fields(ln)
		if len(fields) == 0 {
			t.Fatalf("empty row in readout: %q", stdout)
		}
		labels = append(labels, fields[0])
	}
	return labels
}

// ─── AC-CL-010 — unknown auth verdict carries the same action ──────────────

// wantAuthRowUnknown is the named expected constant of AC-CL-010: the auth
// row for an unestablishable verdict is EXACTLY the unknown token plus the
// one remediation action, identical across every cause axis.
const wantAuthRowUnknown = codexRowLabelAuth + "     " + codexAuthUnknown +
	" — " + codexAuthUnknownAction

// TestCodexReadout_UnknownAuthFourAxes — four cause axes of an
// unestablishable auth verdict, each driven through the REAL two-stage
// ladder (the shared-probe seam is NOT stubbed here; the binary lookup,
// version runner, and login-status runner seams are): every axis must
// report the SAME auth row, byte-identical to the named constant, and no
// logout-sounding phrasing may appear anywhere in the readout.
func TestCodexReadout_UnknownAuthFourAxes(t *testing.T) {
	isolatedCodexHome := func(t *testing.T) string {
		t.Helper()
		dir := t.TempDir()
		t.Setenv(codexHomeEnvVar, dir)
		return dir
	}
	axes := []struct {
		name    string
		arrange func(t *testing.T)
	}{
		{"no auth file, both streams empty, rc 0", func(t *testing.T) {
			isolatedCodexHome(t)
			withCodexLoginStatusRunner(t, func(context.Context, string) ([]byte, []byte, int, error) {
				return nil, nil, 0, nil
			})
		}},
		{"no auth file, runner returns an error", func(t *testing.T) {
			isolatedCodexHome(t)
			withCodexLoginStatusRunner(t, func(context.Context, string) ([]byte, []byte, int, error) {
				return nil, nil, 0, errors.New("spawn failed")
			})
		}},
		{"no auth file, non-zero rc, grammar-mismatch line", func(t *testing.T) {
			isolatedCodexHome(t)
			withCodexLoginStatusRunner(t, func(context.Context, string) ([]byte, []byte, int, error) {
				return nil, []byte("error: API key missing"), 1, nil
			})
		}},
		{"unparseable auth file descends and still gaps", func(t *testing.T) {
			dir := isolatedCodexHome(t)
			if err := os.WriteFile(filepath.Join(dir, "auth.json"), []byte("{"), 0o644); err != nil {
				t.Fatal(err)
			}
			withCodexLoginStatusRunner(t, func(context.Context, string) ([]byte, []byte, int, error) {
				return nil, nil, 0, nil
			})
		}},
	}
	logoutPhrases := []string{"logged out", "logged-out", "not logged in", "no credentials", "signed out"}
	for _, ax := range axes {
		t.Run(ax.name, func(t *testing.T) {
			withCodexLookPath(t, func(string) (string, error) { return sentinelCodexBinaryPath, nil })
			withCodexRunner(t, &fakeCodexRunner{stdoutByCmd: map[string]string{"--version": "9.9.9\n"}})
			ax.arrange(t)
			withCodexProjectDir(t, codexWiringFixture(t, "wired"))

			stdout, stderr, err := runCodexCmd(t)
			if err != nil {
				t.Fatalf("bare readout: %v", err)
			}
			if stderr != "" {
				t.Errorf("stderr = %q, want 0 bytes", stderr)
			}
			lines := strings.Split(strings.TrimSuffix(stdout, "\n"), "\n")
			if len(lines) != 6 {
				t.Fatalf("lines = %d, want 6 (%q)", len(lines), stdout)
			}
			if lines[2] != wantAuthRowUnknown {
				t.Errorf("auth row = %q, want the named constant %q", lines[2], wantAuthRowUnknown)
			}
			low := strings.ToLower(stdout)
			for _, phrase := range logoutPhrases {
				if strings.Contains(low, phrase) {
					t.Errorf("readout carries logout phrasing %q: %q", phrase, stdout)
				}
			}
		})
	}
}

// TestCodexReadout_BinaryAbsentFourCells — binary absent x wiring {wired,
// not wired} x both forms (4 cells): rc 0, the label set EQUALS the closed
// six-label set, binary row "not found", wiring row per fixture — the
// fail-open shape (every probe still reports).
func TestCodexReadout_BinaryAbsentFourCells(t *testing.T) {
	wantLabelSet := map[string]bool{
		codexRowLabelCodex: true, codexRowLabelHome: true, codexRowLabelAuth: true,
		codexRowLabelWiring: true, codexRowLabelAgents: true, codexRowLabelHarness: true,
	}
	for _, wiring := range []struct {
		state   string
		wantRow string
	}{
		{"wired", wantWiringRowWired},
		{"absent", wantWiringRowNotWired},
	} {
		for _, form := range []string{"bare", "status"} {
			t.Run(wiring.state+"/"+form, func(t *testing.T) {
				t.Setenv(codexHomeEnvVar, t.TempDir())
				withCodexSetupProbe(t, minimalProbeStub()) // Installed=false
				withCodexProjectDir(t, codexWiringFixture(t, wiring.state))

				var args []string
				if form == "status" {
					args = []string{"status"}
				}
				stdout, _, err := runCodexCmd(t, args...)
				if err != nil {
					t.Fatalf("err = %v, want rc 0 (readout works when the thing it diagnoses is missing)", err)
				}
				labels := codexReadoutLabels(t, stdout)
				got := map[string]bool{}
				for _, l := range labels {
					got[l] = true
				}
				if !reflect.DeepEqual(got, wantLabelSet) {
					t.Errorf("label set = %v, want exactly %v", got, wantLabelSet)
				}
				lines := strings.Split(strings.TrimSuffix(stdout, "\n"), "\n")
				if lines[0] != wantCodexRowNotFound {
					t.Errorf("binary row = %q, want %q", lines[0], wantCodexRowNotFound)
				}
				if lines[3] != wiring.wantRow {
					t.Errorf("wiring row = %q, want %q", lines[3], wiring.wantRow)
				}
			})
		}
	}
}

// TestCodexLaunch_BinaryAbsentSingleDiagnostic — with codexLookPath failing,
// cli/app exit non-zero with launch count 0 and stderr EXACTLY one line
// equal to the install-action constant (REQ-CL-012's single diagnostic).
func TestCodexLaunch_BinaryAbsentSingleDiagnostic(t *testing.T) {
	for _, verb := range []string{"cli", "app"} {
		t.Run(verb, func(t *testing.T) {
			cap := withCodexLaunchCapture(t)
			prevLook := codexLookPath
			codexLookPath = func(string) (string, error) { return "", os.ErrNotExist }
			t.Cleanup(func() { codexLookPath = prevLook })
			withCodexProjectRoot(t, t.TempDir())

			_, stderr, err := runCodexCmd(t, verb)
			if err == nil {
				t.Fatalf("%s with no binary: err nil, want refusal", verb)
			}
			if code, ok := ResolveExitCode(err); !ok || code == 0 {
				t.Errorf("exit code = (%d, %v), want non-zero", code, ok)
			}
			want := codexInstallHint + "\n"
			if stderr != want {
				t.Errorf("stderr = %q (%d lines), want exactly one line %q", stderr, strings.Count(strings.TrimSuffix(stderr, "\n"), "\n")+1, want)
			}
			codexWantLaunches(t, cap, 0, 0, 0)
		})
	}
}

// ─── AC-CL-012 — no writes ─────────────────────────────────────────────────

// codexSnapshotTree snapshots file path -> mode|mtime for the whole tree.
func codexSnapshotTree(t *testing.T, root string) map[string]string {
	t.Helper()
	snap := map[string]string{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, ierr := d.Info()
		if ierr != nil {
			return ierr
		}
		snap[path] = fmt.Sprintf("%d|%d", uint32(info.Mode()), info.ModTime().UnixNano())
		return nil
	})
	if err != nil {
		t.Fatalf("snapshot %s: %v", root, err)
	}
	return snap
}

// TestCodexLauncher_NoWriteSnapshot — all four forms run under a fully
// isolated home (HOME, XDG_*, TMPDIR, CLAUDE_CONFIG_DIR redirected; project
// root, CODEX_HOME, and profile dir inside it) with the launch stubbed: the
// isolated-home tree snapshot is IDENTICAL before and after, and a CODEX_HOME
// pointing at a nonexistent path is still nonexistent afterwards.
func TestCodexLauncher_NoWriteSnapshot(t *testing.T) {
	isoHome := t.TempDir()
	for _, sub := range []string{"tmp", "codex-home", "project", "claude-profile", "xdg/config", "xdg/cache", "xdg/data"} {
		if err := os.MkdirAll(filepath.Join(isoHome, sub), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("HOME", isoHome)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(isoHome, "xdg/config"))
	t.Setenv("XDG_CACHE_HOME", filepath.Join(isoHome, "xdg/cache"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(isoHome, "xdg/data"))
	t.Setenv("TMPDIR", filepath.Join(isoHome, "tmp"))
	t.Setenv("CLAUDE_CONFIG_DIR", filepath.Join(isoHome, "claude-profile"))

	existingCodexHome := filepath.Join(isoHome, "codex-home")
	absentCodexHome := filepath.Join(isoHome, "does-not-exist")

	runs := []struct {
		name string
		env  string // CODEX_HOME value for this run
		args []string
	}{
		{"bare", existingCodexHome, nil},
		{"status", existingCodexHome, []string{"status"}},
		{"cli", existingCodexHome, []string{"cli"}},
		{"app", existingCodexHome, []string{"app"}},
		{"bare with absent CODEX_HOME", absentCodexHome, nil},
		{"status with absent CODEX_HOME", absentCodexHome, []string{"status"}},
		{"cli with absent CODEX_HOME", absentCodexHome, []string{"cli"}},
		{"app with absent CODEX_HOME", absentCodexHome, []string{"app"}},
	}

	withCodexSetupProbe(t, minimalProbeStub())
	for _, run := range runs {
		t.Run(run.name, func(t *testing.T) {
			t.Setenv(codexHomeEnvVar, run.env)
			cap := withCodexLaunchCapture(t)
			withCodexProjectDir(t, filepath.Join(isoHome, "project"))

			before := codexSnapshotTree(t, isoHome)
			if _, _, err := runCodexCmd(t, run.args...); err != nil {
				t.Fatalf("%s: %v", run.name, err)
			}
			after := codexSnapshotTree(t, isoHome)
			if !reflect.DeepEqual(before, after) {
				keys := map[string]bool{}
				for k := range before {
					keys[k] = true
				}
				for k := range after {
					keys[k] = true
				}
				var diffs []string
				for k := range keys {
					if before[k] != after[k] {
						diffs = append(diffs, fmt.Sprintf("%s: %q -> %q", k, before[k], after[k]))
					}
				}
				sort.Strings(diffs)
				t.Errorf("isolated home tree changed:\n%s", strings.Join(diffs, "\n"))
			}
			if cap.count() > 1 {
				t.Errorf("launches = %d, want at most 1", cap.count())
			}
		})
	}

	// The absent CODEX_HOME is still absent after all four forms ran through
	// it, and the readout REPORTED the gap (AC-CL-012 second cell).
	t.Run("absent CODEX_HOME still absent and reported", func(t *testing.T) {
		if _, err := os.Stat(absentCodexHome); !os.IsNotExist(err) {
			t.Fatalf("%s exists after the runs (err=%v), want IsNotExist", absentCodexHome, err)
		}
		t.Setenv(codexHomeEnvVar, absentCodexHome)
		stdout, _, err := runCodexCmd(t)
		if err != nil {
			t.Fatalf("bare: %v", err)
		}
		if !strings.Contains(stdout, "missing") || !strings.Contains(stdout, codexHomeMissingAction) {
			t.Errorf("home row does not report the missing dir + action: %q", stdout)
		}
	})
}
