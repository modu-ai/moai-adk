package cli

// codex_launch_verb_test.go — the reversed-default launch verb surface.
//
// AC coverage in this file:
//   - AC-CLV-001 — the bare invocation launches
//   - AC-CLV-002 — status renders the readout and launches nothing
//   - AC-CLV-003 — bare and cli build the SAME launch request (value equality)
//   - AC-CLV-004 — an unrouted token never reaches a launch
//   - AC-CLV-005 — no synthesized verb token reaches the child argv
//   - AC-CLV-006 — app IS forwarded (the non-forwarding rule is not over-applied)
//   - AC-CLV-015 — the operator tail after -- passes through verbatim
//   - AC-CLV-007 — CODEX_HOME reaches the child EXPLICITLY, not by inheritance
//   - AC-CLV-008 — the rest of the parent environment survives; both launch
//     paths carry the same CODEX_HOME value
//   - AC-CLV-010 — -w sets the child's working directory and is not forwarded
//   - AC-CLV-011 — -w resolves an existing worktree and never creates one
//   - AC-CLV-012 — the init offer gate is inherited by the bare form
//   - AC-CLV-013 — cross-platform property, with the scan mutation-controlled
//
// AC-CLV-009 (no verb writes) is owned by the pre-existing no-write cell in
// codex_launcher_readout_test.go, whose verb table already includes the bare
// form. AC-CLV-014 (the three-launcher bare-invocation convention) extends the
// existing cross-launcher comparison in codex_launcher_test.go rather than
// forking a second cross-launcher file.

import (
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

// ─── shared helpers ────────────────────────────────────────────────────────

// codexEnvLast returns the LAST value bound to key in an exec.Cmd-style
// environment slice. Last-wins is the semantics the os/exec child applies, so
// reading the first occurrence would judge a value the child never sees.
func codexEnvLast(env []string, key string) (string, bool) {
	prefix := key + "="
	value := ""
	found := false
	for _, entry := range env {
		if strings.HasPrefix(entry, prefix) {
			value = strings.TrimPrefix(entry, prefix)
			found = true
		}
	}
	return value, found
}

// codexOnlyRecord returns the single captured launch record, failing the cell
// rather than panicking on an index when nothing launched.
func codexOnlyRecord(t *testing.T, cap *codexLaunchCapture) codexLaunchRecord {
	t.Helper()
	if cap.count() != 1 {
		t.Fatalf("launches = %d, want exactly 1", cap.count())
	}
	return cap.records[0]
}

// codexArgvTail returns the captured argv AFTER the program token.
func codexArgvTail(t *testing.T, cap *codexLaunchCapture) []string {
	t.Helper()
	if cap.count() != 1 {
		t.Fatalf("launches = %d, want exactly 1", cap.count())
	}
	return cap.records[0].Argv[1:]
}

// ─── AC-CLV-001 / 002 / 003 — the reversed default ─────────────────────────

// TestCodexLaunchVerb_BareInvocationLaunches — the bare form launches exactly
// once and prints NO readout row. Asserting only "launch count 1" would pass
// an implementation that launches AND still prints the readout, so the stdout
// silence is judged alongside it.
func TestCodexLaunchVerb_BareInvocationLaunches(t *testing.T) {
	withCodexSetupProbe(t, minimalProbeStub())
	cap := withCodexLaunchCapture(t)
	withCodexProjectRoot(t, t.TempDir())

	stdout, stderr, err := runCodexCmd(t)
	if err != nil {
		t.Fatalf("bare invocation: %v", err)
	}
	codexWantLaunches(t, cap, 1, 1, 0)
	if stdout != "" {
		t.Errorf("bare stdout = %q, want empty (the readout moved to the status alias)", stdout)
	}
	if stderr != "" {
		t.Errorf("bare stderr = %q, want empty", stderr)
	}
}

// TestCodexLaunchVerb_StatusStaysTheReadout — status remains the explicit
// readout alias: six rows on stdout, rc 0, launch count 0.
func TestCodexLaunchVerb_StatusStaysTheReadout(t *testing.T) {
	t.Setenv(codexHomeEnvVar, t.TempDir())
	withCodexSetupProbe(t, minimalProbeStub())
	withCodexProjectDir(t, codexWiringFixture(t, "wired"))
	cap := withCodexLaunchCapture(t)
	withCodexProjectRoot(t, t.TempDir())

	stdout, _, err := runCodexCmd(t, "status")
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	codexWantLaunches(t, cap, 0, 0, 0)
	lines := strings.Split(strings.TrimSuffix(stdout, "\n"), "\n")
	if len(lines) != 6 {
		t.Errorf("status readout lines = %d, want 6 (%q)", len(lines), stdout)
	}
}

// TestCodexLaunchVerb_BareAndCliBuildTheSameRequest — the two forms are
// synonyms, judged by VALUE: the (program, argv, dir) triple the launch seam
// receives is identical. Checking that each "launched" would not judge this
// axis at all.
func TestCodexLaunchVerb_BareAndCliBuildTheSameRequest(t *testing.T) {
	cap := withCodexLaunchCapture(t)
	root := t.TempDir()
	withCodexProjectRoot(t, root)

	if _, _, err := runCodexCmd(t); err != nil {
		t.Fatalf("bare: %v", err)
	}
	if _, _, err := runCodexCmd(t, "cli"); err != nil {
		t.Fatalf("cli: %v", err)
	}
	if cap.count() != 2 {
		t.Fatalf("launches = %d, want 2", cap.count())
	}
	bare, cli := cap.records[0], cap.records[1]
	if bare.Program != cli.Program {
		t.Errorf("program differs: bare %q vs cli %q", bare.Program, cli.Program)
	}
	if !reflect.DeepEqual(bare.Argv, cli.Argv) {
		t.Errorf("argv differs:\n bare = %#v\n cli  = %#v", bare.Argv, cli.Argv)
	}
	if bare.Dir != cli.Dir {
		t.Errorf("dir differs: bare %q vs cli %q", bare.Dir, cli.Dir)
	}
}

// ─── AC-CLV-004 — the routing table stays a closed set ─────────────────────

// TestCodexLaunchVerb_RoutingSetsAfterReversal — the two token sets derived
// SYMBOLICALLY from the routing table are exactly {"", cli, app} and {status}.
// Absence from the table is still the rejection: six probe tokens each reject
// with the usage constant and zero launches.
func TestCodexLaunchVerb_RoutingSetsAfterReversal(t *testing.T) {
	launch, readout := map[string]bool{}, map[string]bool{}
	for tok, kind := range codexVerbRouting {
		if kind.launches() {
			launch[tok] = true
		} else {
			readout[tok] = true
		}
	}
	if want := (map[string]bool{"": true, "cli": true, "app": true}); !reflect.DeepEqual(launch, want) {
		t.Errorf("launch token set = %v, want %v", launch, want)
	}
	if want := (map[string]bool{"status": true}); !reflect.DeepEqual(readout, want) {
		t.Errorf("readout token set = %v, want %v", readout, want)
	}

	for _, tok := range []string{"launch", "run", "bogus", "cl", "CLI", "start"} {
		t.Run(tok, func(t *testing.T) {
			cap := withCodexLaunchCapture(t)
			withCodexProjectRoot(t, t.TempDir())
			_, stderr, err := runCodexCmd(t, tok)
			if err == nil {
				t.Fatalf("token %q accepted, want the usage rejection", tok)
			}
			if code, ok := ResolveExitCode(err); !ok || code != 1 {
				t.Errorf("token %q exit = (%d, %v), want deliberate rc 1", tok, code, ok)
			}
			if want := codexUsageDiag + "\n"; stderr != want {
				t.Errorf("token %q diagnostic = %q, want %q", tok, stderr, want)
			}
			codexWantLaunches(t, cap, 0, 0, 0)
		})
	}
}

// ─── AC-CLV-005 / 006 / 015 — argv translation ─────────────────────────────

// TestCodexLaunchVerb_SynthesizedVerbNeverReachesChild — neither the bare
// form's empty token nor the cli synonym appears in the child argv. codex has
// no cli subcommand (its usage is `codex [OPTIONS] [PROMPT]`), so a forwarded
// cli token would be read by codex as a PROMPT, not as a verb.
func TestCodexLaunchVerb_SynthesizedVerbNeverReachesChild(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
	}{
		{"bare", nil},
		{"cli", []string{"cli"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cap := withCodexLaunchCapture(t)
			withCodexProjectRoot(t, t.TempDir())
			if _, _, err := runCodexCmd(t, tc.args...); err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}
			tail := codexArgvTail(t, cap)
			if len(tail) != 0 {
				t.Errorf("%s child argv tail = %#v, want empty", tc.name, tail)
			}
			for _, tok := range tail {
				if tok == "cli" || tok == "" {
					t.Errorf("%s: synthesized verb token %q reached the child argv", tc.name, tok)
				}
			}
		})
	}
}

// TestCodexLaunchVerb_AppTokenStillForwarded — app names a REAL codex
// subcommand, so the non-forwarding rule must not swallow it. This cell is
// what stops the fix from being over-applied.
func TestCodexLaunchVerb_AppTokenStillForwarded(t *testing.T) {
	cap := withCodexLaunchCapture(t)
	withCodexProjectRoot(t, t.TempDir())
	if _, _, err := runCodexCmd(t, "app"); err != nil {
		t.Fatalf("app: %v", err)
	}
	if want := []string{"app"}; !reflect.DeepEqual(codexArgvTail(t, cap), want) {
		t.Errorf("app child argv tail = %#v, want %#v", codexArgvTail(t, cap), want)
	}
}

// TestCodexLaunchVerb_TailPassesThroughVerbatim — the operator's tail after
// -- arrives with NO token prepended, on both synonym forms.
func TestCodexLaunchVerb_TailPassesThroughVerbatim(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
	}{
		{"bare", []string{"--", "--model", "o3"}},
		{"cli", []string{"cli", "--", "--model", "o3"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cap := withCodexLaunchCapture(t)
			withCodexProjectRoot(t, t.TempDir())
			if _, _, err := runCodexCmd(t, tc.args...); err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}
			if want := []string{"--model", "o3"}; !reflect.DeepEqual(codexArgvTail(t, cap), want) {
				t.Errorf("%s child argv tail = %#v, want %#v", tc.name, codexArgvTail(t, cap), want)
			}
		})
	}
}

// ─── AC-CLV-007 / 008 — the environment ────────────────────────────────────

// TestCodexLaunchVerb_CodexHomeExplicitToChild — the resolved CODEX_HOME is
// an EXPLICIT entry on the child's environment in both source cases. The
// parent-unset case is the one an ambient-inheritance implementation fails:
// there is nothing to inherit, so only an explicit entry can be present.
func TestCodexLaunchVerb_CodexHomeExplicitToChild(t *testing.T) {
	t.Run("env-supplied", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv(codexHomeEnvVar, home)
		cap := withCodexLaunchCapture(t)
		withCodexProjectRoot(t, t.TempDir())
		if _, _, err := runCodexCmd(t); err != nil {
			t.Fatalf("bare: %v", err)
		}
		got, ok := codexEnvLast(codexOnlyRecord(t, cap).Env, codexHomeEnvVar)
		if !ok {
			t.Fatalf("child env carries no %s entry (env len %d)", codexHomeEnvVar, len(codexOnlyRecord(t, cap).Env))
		}
		if got != home {
			t.Errorf("child %s = %q, want %q", codexHomeEnvVar, got, home)
		}
	})

	t.Run("parent unset falls back to the default and is still explicit", func(t *testing.T) {
		// t.Setenv registers the restore; Unsetenv then removes the variable
		// entirely, which is the state an ambient-inheritance implementation
		// cannot satisfy.
		t.Setenv(codexHomeEnvVar, "")
		if err := os.Unsetenv(codexHomeEnvVar); err != nil {
			t.Fatal(err)
		}
		// resolveCodexHomeDir's second result is the SOURCE label, not an
		// error; only the resolved path is judged here.
		want, _ := resolveCodexHomeDir()
		if want == "" {
			t.Skip("home directory unresolvable in this environment")
		}
		cap := withCodexLaunchCapture(t)
		withCodexProjectRoot(t, t.TempDir())
		if _, _, err := runCodexCmd(t); err != nil {
			t.Fatalf("bare: %v", err)
		}
		got, ok := codexEnvLast(codexOnlyRecord(t, cap).Env, codexHomeEnvVar)
		if !ok {
			t.Fatalf("child env carries no %s entry with the parent unset - the implementation relies on ambient inheritance", codexHomeEnvVar)
		}
		if got != want {
			t.Errorf("child %s = %q, want the resolved default %q", codexHomeEnvVar, got, want)
		}
	})
}

// TestCodexLaunchVerb_ParentEnvPreserved — an arbitrary marker variable in the
// parent environment survives into the child. An implementation that assigns
// only the one entry (dropping the rest) dies here.
func TestCodexLaunchVerb_ParentEnvPreserved(t *testing.T) {
	const marker = "MOAI_CODEX_ENV_MARKER"
	t.Setenv(marker, "kept")
	t.Setenv(codexHomeEnvVar, t.TempDir())
	cap := withCodexLaunchCapture(t)
	withCodexProjectRoot(t, t.TempDir())
	if _, _, err := runCodexCmd(t); err != nil {
		t.Fatalf("bare: %v", err)
	}
	got, ok := codexEnvLast(codexOnlyRecord(t, cap).Env, marker)
	if !ok || got != "kept" {
		t.Errorf("child %s = (%q, %v), want (\"kept\", true)", marker, got, ok)
	}
}

// TestCodexLaunchVerb_SpawnCarriesTheSameCodexHome — the new-window command
// string carries the SAME resolved value the direct path puts on the child
// environment. A tmux window inherits the tmux server's environment, not this
// process's, so an implicit path would silently differ here.
func TestCodexLaunchVerb_SpawnCarriesTheSameCodexHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv(codexHomeEnvVar, home)
	cap := withCodexLaunchCapture(t)
	withCodexProjectRoot(t, t.TempDir())
	if _, _, err := runCodexCmd(t); err != nil {
		t.Fatalf("bare: %v", err)
	}
	direct, ok := codexEnvLast(codexOnlyRecord(t, cap).Env, codexHomeEnvVar)
	if !ok {
		t.Fatalf("direct child env carries no %s", codexHomeEnvVar)
	}

	command := buildCodexSpawnCommand("/stub/bin/codex", []string{"--model", "o3"})
	wantPrefix := codexHomeEnvVar + "=" + shellQuote(direct) + " "
	if !strings.HasPrefix(command, wantPrefix) {
		t.Errorf("spawn command = %q, want it to open with %q (same value as the direct path)", command, wantPrefix)
	}
}

// ─── AC-CLV-010 / 011 — the worktree flag ──────────────────────────────────

// codexWorktreeFixture creates an L1 worktree directory under a fresh project
// root and returns (projectRoot, worktreePath).
//
// Note on scope: the ACCEPTING absolute-path branch is not exercised from
// here. resolveWorktreeL2Path validates an absolute value against prefixes
// derived from the REAL project root (findProjectRoot, not the injected
// seam), so an absolute path under a t.TempDir project root is out-of-prefix
// by construction. The rejecting branch IS exercised below, and its
// diagnostic is compared against resolveWorktreeL2Path's own.
func codexWorktreeFixture(t *testing.T, name string) (string, string) {
	t.Helper()
	root := t.TempDir()
	path := filepath.Join(root, ".claude", "worktrees", name)
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
	return root, path
}

// TestCodexLaunchVerb_WorktreeSetsDirAndIsNotForwarded — moai interprets -w
// itself: the child's working directory becomes the worktree root, and
// neither -w nor --worktree survives into the child argv. codex's top-level
// help exposes no worktree flag, so a forwarded one would be read as a prompt.
func TestCodexLaunchVerb_WorktreeSetsDirAndIsNotForwarded(t *testing.T) {
	root, wt := codexWorktreeFixture(t, "feat-login")
	for _, tc := range []struct {
		name string
		args []string
	}{
		{"short name, two tokens", []string{"-w", "feat-login"}},
		{"short name, equals form", []string{"-w=feat-login"}},
		{"long flag", []string{"--worktree", "feat-login"}},
		{"long flag, equals form", []string{"--worktree=feat-login"}},
		{"after an explicit verb", []string{"cli", "-w", "feat-login"}},
		{"with a passthrough tail", []string{"-w", "feat-login", "--", "--model", "o3"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cap := withCodexLaunchCapture(t)
			withCodexProjectRoot(t, root)
			if _, _, err := runCodexCmd(t, tc.args...); err != nil {
				t.Fatalf("%v: %v", tc.args, err)
			}
			if cap.count() != 1 {
				t.Fatalf("launches = %d, want 1", cap.count())
			}
			if got := cap.records[0].Dir; got != wt {
				t.Errorf("child dir = %q, want the worktree root %q", got, wt)
			}
			for _, tok := range cap.records[0].Argv {
				if tok == "-w" || tok == "--worktree" || tok == "feat-login" {
					t.Errorf("worktree token %q survived into the child argv %#v", tok, cap.records[0].Argv)
				}
			}
		})
	}
}

// TestCodexLaunchVerb_WorktreeResolvesNeverCreates — the rejection cells.
// (i) an absolute path outside the accepted prefixes fails with the SAME
// diagnostic moai cc produces for that input; (ii) a worktree name that does
// not exist is diagnosed rather than created. Both launch nothing, and (ii)
// additionally asserts the path is still absent afterwards — a cell that
// passes while a worktree appeared would mean create, not resolve.
func TestCodexLaunchVerb_WorktreeResolvesNeverCreates(t *testing.T) {
	t.Run("(i) absolute path outside the accepted prefixes", func(t *testing.T) {
		outside := filepath.Join(t.TempDir(), "elsewhere")
		if err := os.MkdirAll(outside, 0o755); err != nil {
			t.Fatal(err)
		}
		cap := withCodexLaunchCapture(t)
		withCodexProjectRoot(t, t.TempDir())

		_, stderr, err := runCodexCmd(t, "-w", outside)
		if err == nil {
			t.Fatal("out-of-prefix absolute path accepted, want a diagnostic")
		}
		codexWantLaunches(t, cap, 0, 0, 0)

		// The diagnostic is the SAME one moai cc emits for this input: both
		// go through resolveWorktreeL2Path, so the texts are compared rather
		// than restated. The launcher writes it to stderr and returns a bare
		// exit code, the way every other diagnostic on this surface does.
		ccErr := resolveWorktreeL2Path([]string{"--worktree", outside})
		if ccErr == nil {
			t.Fatal("resolveWorktreeL2Path accepted the out-of-prefix path (fixture broken)")
		}
		if want := ccErr.Error() + "\n"; stderr != want {
			t.Errorf("diagnostic differs:\n codex = %q\n cc    = %q", stderr, want)
		}
	})

	t.Run("(ii) named worktree that does not exist", func(t *testing.T) {
		root := t.TempDir()
		missing := filepath.Join(root, ".claude", "worktrees", "never-made")
		cap := withCodexLaunchCapture(t)
		withCodexProjectRoot(t, root)

		_, _, err := runCodexCmd(t, "-w", "never-made")
		if err == nil {
			t.Fatal("absent worktree accepted, want a diagnostic")
		}
		codexWantLaunches(t, cap, 0, 0, 0)
		if _, statErr := os.Stat(missing); !os.IsNotExist(statErr) {
			t.Errorf("%s exists after the run (stat err %v) - the flag CREATED a worktree; it must only resolve", missing, statErr)
		}
	})

	t.Run("(iii) bare -w with no value", func(t *testing.T) {
		cap := withCodexLaunchCapture(t)
		withCodexProjectRoot(t, t.TempDir())
		if _, _, err := runCodexCmd(t, "-w"); err == nil {
			t.Fatal("valueless -w accepted, want a diagnostic (there is no name to resolve)")
		}
		codexWantLaunches(t, cap, 0, 0, 0)
	})
}

// ─── AC-CLV-012 — the offer gate is inherited by the bare form ─────────────

// TestCodexLaunchVerb_GateInheritedByBareForm — the reversed default does not
// route around the init offer gate: an incomplete wiring in a non-interactive
// session issues ZERO prompts and launches nothing, and the wired state does
// not block. Each incomplete state is driven separately: a single state would
// not show the gate judging the classifier's answer rather than one fixture.
func TestCodexLaunchVerb_GateInheritedByBareForm(t *testing.T) {
	for _, state := range []string{"absent", "empty dir", "hooks only", "config only"} {
		t.Run("incomplete: "+state, func(t *testing.T) {
			root := codexWiringFixture(t, state)
			h := withCodexGateHarness(t, root, codexHarnessOpts{capable: false})
			withCodexProjectRoot(t, root)

			_, stderr, err := runCodexCmd(t)
			if err == nil {
				t.Fatal("bare invocation on incomplete wiring succeeded, want the gate to stop it")
			}
			if code, ok := ResolveExitCode(err); !ok || code == 0 {
				t.Errorf("exit = (%d, %v), want a deliberate non-zero code", code, ok)
			}
			if h.promptCalls != 0 {
				t.Errorf("prompt calls = %d, want 0 in a non-interactive session", h.promptCalls)
			}
			codexWantLaunches(t, h.launch, 0, 0, 0)
			if !strings.Contains(stderr, "non-interactive") || !strings.Contains(stderr, codexWiringAction) {
				t.Errorf("stderr = %q, want the wiring state and its action", stderr)
			}
		})
	}

	t.Run("wired does not block", func(t *testing.T) {
		root := codexWiringFixture(t, "wired")
		h := withCodexGateHarness(t, root, codexHarnessOpts{capable: false})
		withCodexProjectRoot(t, root)
		if _, _, err := runCodexCmd(t); err != nil {
			t.Fatalf("bare invocation on wired project: %v", err)
		}
		codexWantLaunches(t, h.launch, 1, 1, 0)
	})
}

// ─── AC-CLV-013 — cross-platform property, mutation-controlled ─────────────

// codexLaunchPathFiles is the file set REQ-CLV-010's property is scoped to:
// the codex launch path itself. Files shared with the cc/glm launchers
// (launcher.go and its exec helpers) are deliberately NOT in this set — they
// carry the process-replacement primitive by design, and claiming otherwise
// would be a scan whose scope does not match its claim.
var codexLaunchPathFiles = []string{"codex_launcher.go", "codex_init.go"}

// codexPlatformScanFindings applies the three cross-platform judgments to one
// source text and returns a finding per violation. Returning findings (rather
// than calling t.Errorf directly) is what lets the SAME scan be pointed at a
// planted violation, so its zero on the real files means something.
func codexPlatformScanFindings(src string) []string {
	var out []string
	buildTag := regexp.MustCompile(`(?m)^//go:build\s+(.*)$`)
	osToken := regexp.MustCompile(`\b(windows|darwin|linux|unix)\b`)
	for _, m := range buildTag.FindAllStringSubmatch(src, -1) {
		if osToken.MatchString(m[1]) {
			out = append(out, "OS build tag: "+m[0])
		}
	}
	if strings.Contains(src, `"syscall"`) {
		out = append(out, "syscall import")
	}
	for _, ident := range []string{"syscall.Exec", "unix.Exec", "golang.org/x/sys/unix"} {
		if strings.Contains(src, ident) {
			out = append(out, "process-replacement identifier: "+ident)
		}
	}
	return out
}

// TestCodexLaunchPath_CrossPlatformPropertyHolds — the launch-path files
// produce zero findings, AND the same scan produces a finding for each
// planted violation. Without the second half a scan that matches nothing at
// all would report the same zero.
func TestCodexLaunchPath_CrossPlatformPropertyHolds(t *testing.T) {
	t.Run("real files score zero", func(t *testing.T) {
		for _, name := range codexLaunchPathFiles {
			src := codexReadSpecFile(t, name)
			if findings := codexPlatformScanFindings(src); len(findings) != 0 {
				t.Errorf("%s: %v", name, findings)
			}
		}
	})

	// The mutation control: each planted source MUST be caught. A scan that
	// silently matches nothing passes the cell above and fails here.
	t.Run("mutation control - planted violations are caught", func(t *testing.T) {
		for _, planted := range []struct {
			name string
			src  string
		}{
			{"OS build tag", "//go:build windows\n\npackage cli\n"},
			{"syscall import", "package cli\n\nimport (\n\t\"syscall\"\n)\n"},
			{"process replacement", "package cli\n\nfunc f() { _ = syscall.Exec }\n"},
			{"x/sys/unix", "package cli\n\nimport \"golang.org/x/sys/unix\"\n"},
		} {
			if findings := codexPlatformScanFindings(planted.src); len(findings) == 0 {
				t.Errorf("planted %q produced no finding - the scan is vacuous", planted.name)
			}
		}
	})
}
