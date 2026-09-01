package cli

// codex_readiness_test.go — SPEC-CODEX-LAUNCHER-001 M2 unit tests.
//
// AC coverage in this file (the command-level legs of these ACs land with M3):
//   - AC-CL-005 — CODEX_HOME interpretation and source labels
//   - AC-CL-004 — the wiring six-state matrix, mutual exclusivity, sentinels
//   - AC-CL-006 — informational wording for every incomplete state
//   - AC-CL-011 — row label closed set, binary-absent row
//   - AC-CL-012 — the CODEX_HOME leg (no directory is ever created)
//   - AC-CL-007 — sentinel propagation into the readout (launcher surface)

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/codexadapter"
	"github.com/modu-ai/moai-adk/internal/codexwiring"
)

// ─── AC-CL-007 / AC-CL-004 sentinel values (distinct from any real value) ───

const (
	sentinelCodexBinaryPath = "/sentinel/path/codex"
	sentinelCodexVersion    = "SENTINEL-VER-9x9"
	sentinelCodexAuth       = "sentinel-provider"
)

// ─── AC-CL-004 expected rows (named constants, compared with ==) ───

const (
	wantCodexRowSentinel = "codex    " + sentinelCodexVersion + " (" + sentinelCodexBinaryPath + ")"
	wantCodexRowNoVer    = "codex    (" + sentinelCodexBinaryPath + ")"
	wantCodexRowNotFound = "codex    " + codexBinaryNotFound
	wantAuthRowSentinel  = "auth     " + sentinelCodexAuth

	wantWiringRowNotWired = "wiring   " + codexWiringStatusNotWired +
		" (.codex/hooks.json and .codex/config.toml absent) — " + codexWiringAction
	wantWiringRowPartialNoConfig = "wiring   " + codexWiringStatusPartial +
		" (.codex/config.toml missing) — " + codexWiringAction
	wantWiringRowPartialNoHooks = "wiring   " + codexWiringStatusPartial +
		" (.codex/hooks.json missing) — " + codexWiringAction
	wantWiringRowWired = "wiring   " + codexWiringStatusWired +
		" (.codex/hooks.json, .codex/config.toml)"
	wantWiringRowInvalid = "wiring   " + codexWiringStatusInvalid +
		` (hooks.json whitelist: unaccepted top-level key "rogue_key") — ` + codexWiringAction
	wantWiringRowUnparseable = "wiring   " + codexWiringStatusInvalid +
		" (hooks.json unparseable) — " + codexWiringAction

	wantHarnessRow = "harness  " + codexHarnessCommand
)

// ─── seams and fixtures ───

// withCodexSetupProbe swaps the shared-probe seam and restores it on cleanup.
func withCodexSetupProbe(t *testing.T, s CodexSetupResult) {
	t.Helper()
	prev := codexSetupProbe
	codexSetupProbe = func(context.Context) CodexSetupResult { return s }
	t.Cleanup(func() { codexSetupProbe = prev })
}

// withCodexProjectDir is the project-root seam swap, already declared in
// codex_task_test.go and reused here unchanged.

// withCodexUserHomeDir swaps the home-resolution seam, counting calls.
func withCodexUserHomeDir(t *testing.T, home string) *int {
	t.Helper()
	calls := 0
	prev := codexUserHomeDir
	codexUserHomeDir = func() (string, error) { calls++; return home, nil }
	t.Cleanup(func() { codexUserHomeDir = prev })
	return &calls
}

// withCodexValidateHooksConfig swaps the whitelist seam and restores it.
func withCodexValidateHooksConfig(t *testing.T, f func([]byte) ([]codexadapter.ConfigViolation, error)) {
	t.Helper()
	prev := codexValidateHooksConfig
	codexValidateHooksConfig = f
	t.Cleanup(func() { codexValidateHooksConfig = prev })
}

// unsetCodexHomeEnv removes CODEX_HOME for the test and restores the prior
// environment afterwards.
func unsetCodexHomeEnv(t *testing.T) {
	t.Helper()
	prev, had := os.LookupEnv(codexHomeEnvVar)
	if err := os.Unsetenv(codexHomeEnvVar); err != nil {
		t.Fatalf("unset %s: %v", codexHomeEnvVar, err)
	}
	t.Cleanup(func() {
		if had {
			if err := os.Setenv(codexHomeEnvVar, prev); err != nil {
				t.Fatalf("restore %s: %v", codexHomeEnvVar, err)
			}
			return
		}
		_ = os.Unsetenv(codexHomeEnvVar)
	})
}

// writeCodexWiringFile writes one wiring artifact under root/.codex/.
func writeCodexWiringFile(t *testing.T, root, rel, body string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// minimalProbeStub is the shared-probe stub for tests that exercise the
// launcher-owned rows: binary absent, auth fail-open token.
func minimalProbeStub() CodexSetupResult {
	return CodexSetupResult{AuthProvider: codexAuthUnknown}
}

// ─── AC-CL-005 — CODEX_HOME interpretation and source labels ───

func TestResolveCodexHomeDir_EnvAxis(t *testing.T) {
	home := t.TempDir()
	defaultWant := filepath.Join(home, ".codex")
	cases := []struct {
		name       string
		arrange    func(t *testing.T)
		wantPath   string
		wantSource string
	}{
		{"env set to a path", func(t *testing.T) { t.Setenv(codexHomeEnvVar, "/tmp/xyz") }, "/tmp/xyz", codexHomeSourceEnv},
		{"env unset", unsetCodexHomeEnv, defaultWant, codexHomeSourceDefault},
		{"env set but empty", func(t *testing.T) { t.Setenv(codexHomeEnvVar, "") }, defaultWant, codexHomeSourceDefault},
		{"env set but whitespace only", func(t *testing.T) { t.Setenv(codexHomeEnvVar, "   ") }, defaultWant, codexHomeSourceDefault},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			withCodexUserHomeDir(t, home)
			tc.arrange(t)
			gotPath, gotSource := resolveCodexHomeDir()
			if gotPath != tc.wantPath {
				t.Errorf("path = %q, want %q", gotPath, tc.wantPath)
			}
			if gotSource != tc.wantSource {
				t.Errorf("source = %q, want %q", gotSource, tc.wantSource)
			}
		})
	}
}

// TestResolveCodexHomeDir_TrailingSeparatorJoinsCleanly pins the fallback to
// filepath.Join: a home seam returning "/tmp/h/" must yield "/tmp/h/.codex" —
// a plain `home + "/.codex"` concat doubles the separator and dies here.
func TestResolveCodexHomeDir_TrailingSeparatorJoinsCleanly(t *testing.T) {
	unsetCodexHomeEnv(t)
	withCodexUserHomeDir(t, "/tmp/h/")
	gotPath, gotSource := resolveCodexHomeDir()
	if want := filepath.Join("/tmp/h/", codexHomeDirName); gotPath != want {
		t.Errorf("path = %q, want filepath.Join result %q", gotPath, want)
	}
	if gotSource != codexHomeSourceDefault {
		t.Errorf("source = %q, want %q", gotSource, codexHomeSourceDefault)
	}
}

// TestResolveCodexHomeDir_GoesThroughHomeSeamNotHOMEEnv erases $HOME and
// asserts the four env cells still resolve through the seam (exactly one seam
// call per resolution) — home resolution is seam-based, not a $HOME read.
func TestResolveCodexHomeDir_GoesThroughHomeSeamNotHOMEEnv(t *testing.T) {
	t.Setenv("HOME", "") // a direct $HOME read would fail or empty here
	seamHome := t.TempDir()
	calls := withCodexUserHomeDir(t, seamHome)
	defaultWant := filepath.Join(seamHome, ".codex")

	unsetCodexHomeEnv(t)
	gotPath, gotSource := resolveCodexHomeDir()
	if gotPath != defaultWant || gotSource != codexHomeSourceDefault {
		t.Fatalf("unset cell: path=%q source=%q, want %q/%s", gotPath, gotSource, defaultWant, codexHomeSourceDefault)
	}
	if *calls != 1 {
		t.Fatalf("seam calls = %d, want exactly 1 per resolution", *calls)
	}

	cells := []struct {
		name         string
		arrange      func(t *testing.T)
		wantPath     string
		wantSource   string
		wantSeamHits int
	}{
		{"env set", func(t *testing.T) { t.Setenv(codexHomeEnvVar, "/tmp/xyz") }, "/tmp/xyz", codexHomeSourceEnv, 0},
		{"env empty", func(t *testing.T) { t.Setenv(codexHomeEnvVar, "") }, defaultWant, codexHomeSourceDefault, 1},
		{"env whitespace", func(t *testing.T) { t.Setenv(codexHomeEnvVar, "   ") }, defaultWant, codexHomeSourceDefault, 1},
	}
	for _, tc := range cells {
		t.Run(tc.name, func(t *testing.T) {
			before := *calls
			tc.arrange(t)
			gotPath, gotSource := resolveCodexHomeDir()
			if gotPath != tc.wantPath || gotSource != tc.wantSource {
				t.Errorf("path=%q source=%q, want %q/%q", gotPath, gotSource, tc.wantPath, tc.wantSource)
			}
			if *calls != before+tc.wantSeamHits {
				t.Errorf("seam calls delta = %d, want %d", *calls-before, tc.wantSeamHits)
			}
		})
	}
}

// ─── AC-CL-004 — wiring six-state matrix (through the full assembly) ───

func TestClassifyCodexWiring_SixStateMatrix(t *testing.T) {
	validHooks := "{}\n"
	rogueHooks := `{"rogue_key": 1}` + "\n"
	configBody := "model = \"g\"\n"
	cases := []struct {
		name    string
		arrange func(t *testing.T, root string)
		wantRow string
	}{
		{"codex dir absent", func(*testing.T, string) {}, wantWiringRowNotWired},
		{"codex dir present but empty", func(t *testing.T, root string) {
			if err := os.MkdirAll(filepath.Join(root, ".codex"), 0o755); err != nil {
				t.Fatal(err)
			}
		}, wantWiringRowNotWired},
		{"hooks.json only", func(t *testing.T, root string) {
			writeCodexWiringFile(t, root, codexwiring.HooksRelPath, validHooks)
		}, wantWiringRowPartialNoConfig},
		{"config.toml only", func(t *testing.T, root string) {
			writeCodexWiringFile(t, root, codexwiring.ConfigRelPath, configBody)
		}, wantWiringRowPartialNoHooks},
		{"both present and valid", func(t *testing.T, root string) {
			writeCodexWiringFile(t, root, codexwiring.HooksRelPath, validHooks)
			writeCodexWiringFile(t, root, codexwiring.ConfigRelPath, configBody)
		}, wantWiringRowWired},
		{"both present, hooks whitelist violation", func(t *testing.T, root string) {
			writeCodexWiringFile(t, root, codexwiring.HooksRelPath, rogueHooks)
			writeCodexWiringFile(t, root, codexwiring.ConfigRelPath, configBody)
		}, wantWiringRowInvalid},
		{"both present, hooks unparseable", func(t *testing.T, root string) {
			writeCodexWiringFile(t, root, codexwiring.HooksRelPath, "{")
			writeCodexWiringFile(t, root, codexwiring.ConfigRelPath, configBody)
		}, wantWiringRowUnparseable},
	}
	closedStatusSet := map[string]bool{
		codexWiringStatusNotWired: true,
		codexWiringStatusPartial:  true,
		codexWiringStatusWired:    true,
		codexWiringStatusInvalid:  true,
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(codexHomeEnvVar, t.TempDir())
			withCodexSetupProbe(t, minimalProbeStub())
			root := t.TempDir()
			tc.arrange(t, root)
			withCodexProjectDir(t, root)

			info := classifyCodexWiring(root)
			if !closedStatusSet[info.Status] {
				t.Errorf("status %q is outside the closed set", info.Status)
			}
			rows := probeCodexReadiness(context.Background()).rows()
			if rows[3] != tc.wantRow {
				t.Errorf("wiring row = %q, want %q", rows[3], tc.wantRow)
			}
		})
	}

	// Mutual exclusivity of the two partial cells (AC-CL-004): each names ONLY
	// the file that is missing.
	noConfig := wantWiringRowPartialNoConfig
	if !strings.Contains(noConfig, "config.toml") || strings.Contains(noConfig, "hooks.json") {
		t.Errorf("partial(config missing) row %q must name config.toml and NOT hooks.json", noConfig)
	}
	noHooks := wantWiringRowPartialNoHooks
	if !strings.Contains(noHooks, "hooks.json") || strings.Contains(noHooks, "config.toml") {
		t.Errorf("partial(hooks missing) row %q must name hooks.json and NOT config.toml", noHooks)
	}

	// The remediation action rides every incomplete state and never the wired
	// one (REQ-CL-006 / AC-CL-004).
	for name, row := range map[string]string{
		"not wired": wantWiringRowNotWired,
		"partial":   wantWiringRowPartialNoConfig,
		"invalid":   wantWiringRowInvalid,
	} {
		if !strings.Contains(row, "moai init --agent codex") {
			t.Errorf("%s row lacks the action phrase: %q", name, row)
		}
	}
	if strings.Contains(wantWiringRowWired, "moai init --agent codex") {
		t.Errorf("wired row must not recommend an action: %q", wantWiringRowWired)
	}
}

// ─── AC-CL-006 — informational wording over the whole readout ───

func TestCodexReadiness_NoBannedWordsAllStates(t *testing.T) {
	banned := []string{"error", "failed", "fatal", "broken", "cannot", "unable"}
	states := []struct {
		name       string
		arrange    func(t *testing.T, root string)
		wantAction bool
	}{
		{"dir absent", func(*testing.T, string) {}, true},
		{"dir empty", func(t *testing.T, root string) {
			if err := os.MkdirAll(filepath.Join(root, ".codex"), 0o755); err != nil {
				t.Fatal(err)
			}
		}, true},
		{"hooks only", func(t *testing.T, root string) {
			writeCodexWiringFile(t, root, codexwiring.HooksRelPath, "{}\n")
		}, true},
		{"config only", func(t *testing.T, root string) {
			writeCodexWiringFile(t, root, codexwiring.ConfigRelPath, "x = 1\n")
		}, true},
		{"invalid hooks", func(t *testing.T, root string) {
			writeCodexWiringFile(t, root, codexwiring.HooksRelPath, `{"rogue_key": 1}`)
			writeCodexWiringFile(t, root, codexwiring.ConfigRelPath, "x = 1\n")
		}, true},
		{"wired", func(t *testing.T, root string) {
			writeCodexWiringFile(t, root, codexwiring.HooksRelPath, "{}\n")
			writeCodexWiringFile(t, root, codexwiring.ConfigRelPath, "x = 1\n")
		}, false},
	}
	for _, st := range states {
		t.Run(st.name, func(t *testing.T) {
			t.Setenv(codexHomeEnvVar, t.TempDir())
			withCodexSetupProbe(t, minimalProbeStub())
			root := t.TempDir()
			st.arrange(t, root)
			withCodexProjectDir(t, root)
			rows := probeCodexReadiness(context.Background()).rows()
			for i, row := range rows {
				lower := strings.ToLower(row)
				for _, word := range banned {
					if strings.Contains(lower, word) {
						t.Errorf("row %d contains banned word %q: %q", i, word, row)
					}
				}
			}
			joined := strings.Join(rows[:], "\n")
			if st.wantAction && !strings.Contains(joined, "moai init --agent codex") {
				t.Errorf("incomplete state readout lacks the action phrase:\n%s", joined)
			}
			if !st.wantAction && strings.Contains(joined, "moai init --agent codex") {
				t.Errorf("wired state readout recommends an action:\n%s", joined)
			}
		})
	}
}

// ─── AC-CL-007 / AC-CL-004 — sentinel propagation into the readout ───

func TestCodexReadiness_SentinelPropagation(t *testing.T) {
	home := t.TempDir()
	t.Setenv(codexHomeEnvVar, home)
	root := t.TempDir()
	writeCodexWiringFile(t, root, codexwiring.HooksRelPath, "{}\n")
	writeCodexWiringFile(t, root, codexwiring.ConfigRelPath, "x = 1\n")
	withCodexProjectDir(t, root)
	withCodexSetupProbe(t, CodexSetupResult{
		Installed:    true,
		Binary:       sentinelCodexBinaryPath,
		Version:      sentinelCodexVersion,
		AuthProvider: sentinelCodexAuth,
	})

	rows := probeCodexReadiness(context.Background()).rows()
	if rows[0] != wantCodexRowSentinel {
		t.Errorf("codex row = %q, want %q", rows[0], wantCodexRowSentinel)
	}
	if rows[2] != wantAuthRowSentinel {
		t.Errorf("auth row = %q, want %q", rows[2], wantAuthRowSentinel)
	}
	// The home row reports the env-sourced sentinel home verbatim.
	if !strings.Contains(rows[1], home) || !strings.Contains(rows[1], "("+codexHomeSourceEnv+")") {
		t.Errorf("home row = %q, want it to carry %q and source %q", rows[1], home, codexHomeSourceEnv)
	}
	if strings.Contains(rows[1], "missing") {
		t.Errorf("home row for an existing CODEX_HOME must not report missing: %q", rows[1])
	}

	// Binary present but version probe silent: the row carries the path alone.
	withCodexSetupProbe(t, CodexSetupResult{
		Installed:    true,
		Binary:       sentinelCodexBinaryPath,
		AuthProvider: sentinelCodexAuth,
	})
	rows = probeCodexReadiness(context.Background()).rows()
	if rows[0] != wantCodexRowNoVer {
		t.Errorf("codex row without version = %q, want %q", rows[0], wantCodexRowNoVer)
	}

	// A stubbed whitelist violation must surface in the wiring row verbatim —
	// calling the validator and ignoring its result dies here (AC-CL-007).
	withCodexSetupProbe(t, minimalProbeStub())
	withCodexValidateHooksConfig(t, func([]byte) ([]codexadapter.ConfigViolation, error) {
		return []codexadapter.ConfigViolation{{Level: "top-level", Key: "sentinel-violation"}}, nil
	})
	rows = probeCodexReadiness(context.Background()).rows()
	if !strings.Contains(rows[3], `unaccepted top-level key "sentinel-violation"`) {
		t.Errorf("wiring row must reflect the sentinel violation, got %q", rows[3])
	}
	if !strings.Contains(rows[3], codexWiringStatusInvalid) {
		t.Errorf("wiring row status must be invalid, got %q", rows[3])
	}
}

// ─── AC-CL-011 — row label closed set, binary-absent row ───

func TestCodexReadiness_RowLabelSetAndBinaryAbsent(t *testing.T) {
	wantLabels := map[string]bool{
		codexRowLabelCodex: true, codexRowLabelHome: true, codexRowLabelAuth: true,
		codexRowLabelWiring: true, codexRowLabelAgents: true, codexRowLabelHarness: true,
	}
	for _, st := range []struct {
		name          string
		arrange       func(t *testing.T, root string)
		wantWiringRow string
	}{
		{"wired", func(t *testing.T, root string) {
			writeCodexWiringFile(t, root, codexwiring.HooksRelPath, "{}\n")
			writeCodexWiringFile(t, root, codexwiring.ConfigRelPath, "x = 1\n")
		}, wantWiringRowWired},
		{"not wired", func(*testing.T, string) {}, wantWiringRowNotWired},
	} {
		t.Run(st.name, func(t *testing.T) {
			t.Setenv(codexHomeEnvVar, t.TempDir())
			withCodexSetupProbe(t, minimalProbeStub()) // Installed=false
			root := t.TempDir()
			st.arrange(t, root)
			withCodexProjectDir(t, root)

			rows := probeCodexReadiness(context.Background()).rows()
			labels := map[string]bool{}
			for i, row := range rows {
				if strings.TrimSpace(row) == "" {
					t.Fatalf("row %d is empty — a failed probe must degrade to its token, not skip the row", i)
				}
				fields := strings.Fields(row)
				labels[fields[0]] = true
			}
			if !reflect.DeepEqual(labels, wantLabels) {
				t.Errorf("row label set = %v, want %v", labels, wantLabels)
			}
			if rows[0] != wantCodexRowNotFound {
				t.Errorf("binary-absent codex row = %q, want %q", rows[0], wantCodexRowNotFound)
			}
			if rows[3] != st.wantWiringRow {
				t.Errorf("wiring row = %q, want %q", rows[3], st.wantWiringRow)
			}
			if rows[5] != wantHarnessRow {
				t.Errorf("harness row = %q, want %q", rows[5], wantHarnessRow)
			}
		})
	}
}

// ─── AC-CL-012 — CODEX_HOME leg: report the gap, never create the directory ───

func TestCodexReadiness_DoesNotCreateCodexHome(t *testing.T) {
	absent := filepath.Join(t.TempDir(), "absent-codex-home")
	t.Setenv(codexHomeEnvVar, absent)
	withCodexSetupProbe(t, minimalProbeStub())
	withCodexProjectDir(t, t.TempDir())

	_ = probeCodexReadiness(context.Background())
	if _, err := os.Stat(absent); !os.IsNotExist(err) {
		t.Fatalf("readout must not create CODEX_HOME; stat err = %v", err)
	}

	rows := probeCodexReadiness(context.Background()).rows()
	if !strings.Contains(rows[1], "missing") {
		t.Errorf("home row must report the missing state, got %q", rows[1])
	}
	if !strings.Contains(rows[1], codexHomeMissingAction) {
		t.Errorf("home row must carry the %q action, got %q", codexHomeMissingAction, rows[1])
	}
	if _, err := os.Stat(absent); !os.IsNotExist(err) {
		t.Errorf("CODEX_HOME created by a second readout; stat err = %v", err)
	}
}

// ─── fail-open branches ───

// TestResolveCodexHomeInfo_UnresolvableHome pins the seam-error path: an
// unresolvable home leaves Path empty and Exists false, and the row still
// renders (fail-open) instead of erroring.
func TestResolveCodexHomeInfo_UnresolvableHome(t *testing.T) {
	unsetCodexHomeEnv(t)
	prev := codexUserHomeDir
	codexUserHomeDir = func() (string, error) { return "", errors.New("no home on this machine") }
	t.Cleanup(func() { codexUserHomeDir = prev })

	info := resolveCodexHomeInfo()
	if info.Path != "" || info.Source != codexHomeSourceDefault || info.Exists {
		t.Errorf("info = %+v, want Path=\"\" / %s / exists=false", info, codexHomeSourceDefault)
	}
}

// TestClassifyCodexWiring_UnreadableHooks drives the read-failure branch
// portably: hooks.json present as a DIRECTORY stats as present, and ReadFile
// fails with "is a directory" on every platform — the state is invalid, never
// a crash, and the action still rides the row.
func TestClassifyCodexWiring_UnreadableHooks(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".codex", "hooks.json"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeCodexWiringFile(t, root, codexwiring.ConfigRelPath, "x = 1\n")

	info := classifyCodexWiring(root)
	if info.Status != codexWiringStatusInvalid {
		t.Fatalf("status = %q, want %q", info.Status, codexWiringStatusInvalid)
	}
	row := "wiring   " + codexWiringRowValue(info)
	if !strings.Contains(row, "unreadable") {
		t.Errorf("row %q must name the unreadable state", row)
	}
	if !strings.Contains(row, codexWiringAction) {
		t.Errorf("row %q must carry the remediation action", row)
	}
}

// TestCountCodexAgentTOMLs_BadPatternRoot pins the glob-error branch: a root
// whose name breaks the glob pattern counts as 0 agents, never an error.
func TestCountCodexAgentTOMLs_BadPatternRoot(t *testing.T) {
	root := filepath.Join(t.TempDir(), "bad[")
	if err := os.MkdirAll(filepath.Join(root, ".codex", "agents", "moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".codex", "agents", "moai", "a.toml"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := countCodexAgentTOMLs(root); got != 0 {
		t.Errorf("count = %d, want 0 (a glob error degrades to zero, never an error)", got)
	}
}

// ─── agents row ───

func TestCountCodexAgentTOMLs(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".codex", "agents", "moai")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"manager-spec.toml", "manager-git.toml", "notes.md"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if got := countCodexAgentTOMLs(root); got != 2 {
		t.Errorf("count = %d, want 2 (TOMLs only)", got)
	}
	if got := countCodexAgentTOMLs(t.TempDir()); got != 0 {
		t.Errorf("count on an unwired root = %d, want 0", got)
	}

	// Through the assembly: the agents row renders the count.
	t.Setenv(codexHomeEnvVar, t.TempDir())
	withCodexSetupProbe(t, minimalProbeStub())
	withCodexProjectDir(t, root)
	rows := probeCodexReadiness(context.Background()).rows()
	if want := "agents   2 TOML"; rows[4] != want {
		t.Errorf("agents row = %q, want %q", rows[4], want)
	}
}
