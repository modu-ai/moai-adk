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

func resumeSeedRun(t *testing.T, projectDir, llm string) {
	t.Helper()
	injectInitWindowFailure(t)
	if err := resumeRun(t, projectDir, llm); err == nil {
		t.Fatal("the post-executor injection did not bite: the seed run succeeded, so there is no interrupted run to resume")
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

// TestInitResumeDoesNotInferBothFromInstructionFile — finding 1, the "do not
// infer from file presence" rule. A claude-only project carries CLAUDE.md (the
// template writes it for every harness); a user who hand-wrote their own
// CLAUDE.md must still resume as claude. The recorded selection decides, not
// the presence of a CLAUDE.md file.
func TestInitResumeDoesNotInferBothFromInstructionFile(t *testing.T) {
	resumeHome(t)
	projectDir := filepath.Join(t.TempDir(), "claude-claude-proj")
	resumeSeedRun(t, projectDir, "claude")
	if err := os.WriteFile(filepath.Join(projectDir, "CLAUDE.md"), []byte("# hand-written\n"), 0o644); err != nil {
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
// CLAUDE.md present is the gpt layout. AGENTS.md is not a discriminator: no
// harness writes AGENTS.md any more, and an AGENTS.md probe read the claude
// layout as gpt.
func TestDeployedWiringIsGPTReadsTheClaudeSurface(t *testing.T) {
	cases := []struct {
		name  string
		paths []string // a trailing "/" names a directory
		want  bool
	}{
		{"codex-only layout", []string{"CLAUDE.md"}, true},
		{"claude-side layout", []string{".claude/", "CLAUDE.md"}, false},
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
