package cli

// init_resume_wiring_test.go — review finding 1 (card t1591): the harness
// selection a failing init run made must survive into its resume.
//
// The resume restored only the GPT layout (deployedWiringIsGPT). A project
// deployed with --llm both carries CLAUDE.md, so the resume fell back to the
// claude default: the Codex wiring was skipped and the resume still reported
// success. The failing run now records its selection in a project-scoped
// record (.moai/state/resume-wiring), and the resume reads it back when no
// explicit --llm was given.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// resumeHome isolates HOME behind the ensure-failure injection (a corrupt user
// manifest makes the user-asset ensure fail after the deploy) and installs the
// shared init wizard seams.
func resumeHome(t *testing.T) {
	t.Helper()
	home := corruptUserManifestHome(t)
	t.Setenv("HOME", home)
	t.Setenv("MOAI_SANDBOX_PROOF", "")
	t.Setenv("MOAI_DISABLE_BYPASS_PERMISSIONS_MODE", "")
	seamsForInit(t)
}

// resumeSeedRun runs the failing first init. An empty llm names no --llm flag.
// A run that succeeds means the injection did not bite; that is a tool failure,
// so the test fails rather than skipping.
func resumeSeedRun(t *testing.T, projectDir, llm string) {
	t.Helper()
	if err := resumeRun(t, projectDir, llm); err == nil {
		t.Fatal("the ensure-failure injection did not bite: the seed run succeeded, so there is no interrupted run to resume")
	}
}

// resumeRun runs init on projectDir and returns its error. An empty llm names
// no --llm flag, which is the resume path under test.
func resumeRun(t *testing.T, projectDir, llm string) error {
	t.Helper()
	cmd := newInitTestCmd()
	if llm != "" {
		if err := cmd.Flags().Set("llm", llm); err != nil {
			t.Fatal(err)
		}
	}
	var out, errOut strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	return runInit(cmd, []string{projectDir})
}

// readResumeLLMYAML returns the deployed llm.yaml text.
func readResumeLLMYAML(t *testing.T, projectDir string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(projectDir, ".moai", "config", "sections", "llm.yaml"))
	if err != nil {
		t.Fatalf("read llm.yaml: %v", err)
	}
	return string(data)
}

// hasHarness reports whether llm.yaml records the named harness, quoted or not.
func hasHarness(llm, harness string) bool {
	return strings.Contains(llm, "harness: "+harness) || strings.Contains(llm, `harness: "`+harness+`"`)
}

// TestInitResumeRestoresBothSelection — finding 1. The seed run selects
// --llm both and dies at the user-asset ensure; the resume names no --llm. The
// resume must restore the both wiring: the harness recorded as both, the Codex
// side wired, and the record removed once the resume completes.
func TestInitResumeRestoresBothSelection(t *testing.T) {
	resumeHome(t)
	projectDir := filepath.Join(t.TempDir(), "both-resume-proj")
	resumeSeedRun(t, projectDir, "both")

	// The seed run deployed the claude-side layout: .claude/ and the template
	// AGENTS.md. No harness writes CLAUDE.md any more, so a both layout and a
	// claude layout are identical on disk at the failure point — the recorded
	// selection is the only thing that tells them apart.
	for _, rel := range []string{".claude", "AGENTS.md"} {
		if _, err := os.Stat(filepath.Join(projectDir, rel)); err != nil {
			t.Fatalf("the seed run did not deploy the claude-side layout (%s missing): %v", rel, err)
		}
	}

	if err := resumeRun(t, projectDir, ""); err != nil {
		t.Fatalf("run 2 (the resume) failed: %v", err)
	}
	if llm := readResumeLLMYAML(t, projectDir); !hasHarness(llm, "both") {
		t.Fatalf("the resume did not restore the both selection the failing run made — llm.yaml:\n%s", llm)
	}
	if _, err := os.Stat(filepath.Join(projectDir, ".codex", "config.toml")); err != nil {
		t.Fatalf("the resume skipped the Codex wiring for a both selection (.codex/config.toml): %v", err)
	}
	if _, err := os.Stat(filepath.Join(projectDir, ".moai", "state", "resume-wiring")); !os.IsNotExist(err) {
		t.Fatalf("the wiring record survived a COMPLETED resume (stat err = %v)", err)
	}
}

// TestInitResumeDoesNotInferBothFromAGENTSFile — finding 1, the "do not infer
// from file presence" rule. A claude-only project carries AGENTS.md (the
// template writes it for every harness); a user who hand-wrote their own
// AGENTS.md must still resume as claude. The recorded selection decides, not
// the presence of an AGENTS.md file.
func TestInitResumeDoesNotInferBothFromAGENTSFile(t *testing.T) {
	resumeHome(t)
	projectDir := filepath.Join(t.TempDir(), "claude-agents-proj")
	resumeSeedRun(t, projectDir, "claude")
	if err := os.WriteFile(filepath.Join(projectDir, "AGENTS.md"), []byte("# hand-written\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := resumeRun(t, projectDir, ""); err != nil {
		t.Fatalf("run 2 (the resume) failed: %v", err)
	}
	if llm := readResumeLLMYAML(t, projectDir); !hasHarness(llm, "claude") {
		t.Fatalf("the claude-only project did not resume as claude — llm.yaml:\n%s", llm)
	}
	if _, err := os.Stat(filepath.Join(projectDir, ".codex", "config.toml")); err == nil {
		t.Fatal("a claude-only project was misrestored as both: the Codex wiring ran")
	}
}

// TestDeployedWiringIsGPTReadsTheClaudeSurface — the legacy layout fallback
// (used only when no selection record exists) keys on the claude-side surface.
// The codex-only deployer writes no .claude/ tree, so .claude/ absent with
// AGENTS.md present is the gpt layout. CLAUDE.md is not a discriminator: no
// harness writes it any more, and a CLAUDE.md probe read the claude layout as gpt.
func TestDeployedWiringIsGPTReadsTheClaudeSurface(t *testing.T) {
	cases := []struct {
		name  string
		paths []string // a trailing "/" names a directory
		want  bool
	}{
		{"codex-only layout", []string{"AGENTS.md"}, true},
		{"claude-side layout", []string{".claude/", "AGENTS.md"}, false},
		{"claude surface only", []string{".claude/"}, false},
		{"empty project", nil, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			for _, p := range tc.paths {
				full := filepath.Join(root, p)
				if strings.HasSuffix(p, "/") {
					if err := os.MkdirAll(full, 0o755); err != nil {
						t.Fatal(err)
					}
					continue
				}
				if err := os.WriteFile(full, []byte("x\n"), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			if got := deployedWiringIsGPT(root); got != tc.want {
				t.Fatalf("deployedWiringIsGPT = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestInitResumeExplicitLLMOverridesRecordedSelection — an explicit --llm on the
// resume is the current invocation's answer and wins over the recorded
// selection.
func TestInitResumeExplicitLLMOverridesRecordedSelection(t *testing.T) {
	resumeHome(t)
	projectDir := filepath.Join(t.TempDir(), "explicit-resume-proj")
	resumeSeedRun(t, projectDir, "both")

	if err := resumeRun(t, projectDir, "claude"); err != nil {
		t.Fatalf("run 2 (the explicit-llm resume) failed: %v", err)
	}
	if llm := readResumeLLMYAML(t, projectDir); !hasHarness(llm, "claude") {
		t.Fatalf("an explicit --llm claude on the resume did not win over the recorded both — llm.yaml:\n%s", llm)
	}
}
