package cli

// init_resume_refusal_test.go — Item C (card t1591), narrowed by card t1909:
// the harness deployment is claude-only, so a resume can no longer RESTORE a
// legacy gpt/both selection — it refuses instead, before any side effect,
// rather than rewriting llm.yaml as claude over another harness's files.
//
// The legacy evidence a pre-removal interrupted init may have left is the
// selection record (.moai/state/resume-wiring naming gpt/both) or the GPT
// layout (CLAUDE.md with no .claude/). Both are seeded by hand here: init can
// no longer produce them, and the hand-seeded record is byte-identical to the
// one the failing run wrote.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/userassets"
)

// requireClaudeSideLayout asserts the deployed layout is the claude-side one: the
// .claude/ surface is present.
func requireClaudeSideLayout(t *testing.T, projectDir string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(projectDir, ".claude")); err != nil {
		t.Fatalf("precondition: the claude seed run did not deploy the .claude/ surface: %v", err)
	}
}

// makeLegacyGPTLayout rewrites a claude-seeded project into the GPT layout the
// legacy refusal reads: CLAUDE.md present, the .claude/ surface absent. The
// codex-only deployer's file set is exactly that, so the probe cannot tell the
// hand-made layout from a deployed one.
func makeLegacyGPTLayout(t *testing.T, projectDir string) {
	t.Helper()
	if err := os.RemoveAll(filepath.Join(projectDir, ".claude")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(projectDir, "CLAUDE.md")); err != nil {
		if err := os.WriteFile(filepath.Join(projectDir, "CLAUDE.md"), []byte("# CLAUDE\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// resumeForgetRecord removes the selection record the failing run wrote, so the
// resume has no record and only the layout can decide.
func resumeForgetRecord(t *testing.T, projectDir string) {
	t.Helper()
	if _, err := os.Stat(resumeWiringPath(projectDir)); err != nil {
		t.Fatalf("precondition: the failing run did not write the selection record: %v", err)
	}
	clearResumeWiring(projectDir)
}

// requireResumeRefused asserts the refusal contract for a harness mismatch on
// the resume. deployed is the harness the refusal must name as deployed and
// requested the one the resume asked for; record is the selection record the
// failing run left ("" when none). The refusal precedes every side effect, so
// it is checked as an absence: llm.yaml is byte-identical, the resume marker
// was never written, the user manifest is byte-identical to manifestBefore
// (taken with readUserManifestForTest before the resume), and the selection
// record is not retired.
func requireResumeRefused(t *testing.T, err error, projectDir, llmBefore, manifestBefore, record, deployed, requested string) {
	t.Helper()
	if err == nil {
		t.Fatalf("the resume reported success over a harness mismatch (deployed %s, requested %s): no refusal", deployed, requested)
	}
	msg := err.Error()
	for _, want := range []string{deployed, requested, "--force"} {
		if !strings.Contains(msg, want) {
			t.Errorf("the refusal does not contain %q — error: %v", want, err)
		}
	}
	if got := readResumeLLMYAML(t, projectDir); got != llmBefore {
		t.Errorf("the refused resume changed llm.yaml.\n--- before ---\n%s--- after ---\n%s", llmBefore, got)
	}
	if _, statErr := os.Stat(filepath.Join(projectDir, ".moai", "resume-pending")); !os.IsNotExist(statErr) {
		t.Errorf("the refused resume wrote the resume marker (stat err = %v)", statErr)
	}
	if got := readUserManifestForTest(t); got != manifestBefore {
		t.Errorf("the refused resume touched the user manifest (before %d bytes, after %d bytes)", len(manifestBefore), len(got))
	}
	if record != "" {
		if got, ok := readResumeWiring(projectDir); !ok || string(got) != record {
			t.Errorf("the refused resume retired the selection record: got %q (ok=%v), want %q", got, ok, record)
		}
	}
}

// readUserManifestForTest returns the user manifest bytes under HOME, or "" when the
// manifest is absent. A refused resume must leave these bytes unchanged.
func readUserManifestForTest(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(userassets.ManifestPath(os.Getenv("HOME")))
	if err != nil {
		return ""
	}
	return string(data)
}

// TestInitResumeRefusesLegacyGPTBothRecord — Item C under the claude-only
// deployment. A legacy record naming gpt or both cannot be restored (nothing
// resumes INTO a codex-side tree any more), so the resume is refused before
// any side effect, naming the deployed harness and the --force way out.
func TestInitResumeRefusesLegacyGPTBothRecord(t *testing.T) {
	for _, tc := range []struct {
		name   string
		record string
	}{
		{name: "recorded gpt", record: "gpt"},
		{name: "recorded both", record: "both"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resumeHome(t)
			projectDir := filepath.Join(t.TempDir(), "refusal-proj")
			resumeSeedRun(t, projectDir, "")
			// Simulate the pre-removal interrupted init's record: the same
			// bytes persistResumeWiring writes, naming the legacy harness.
			if err := persistResumeWiring(projectDir, agentWiring(tc.record)); err != nil {
				t.Fatal(err)
			}
			requireClaudeSideLayout(t, projectDir)
			llmBefore := readResumeLLMYAML(t, projectDir)
			manifestBefore := readUserManifestForTest(t)

			err := resumeRun(t, projectDir, "")
			requireResumeRefused(t, err, projectDir, llmBefore, manifestBefore, tc.record, tc.record, "claude")
		})
	}
}

// TestInitResumeRefusesGPTLayoutWithoutRecord — the layout arm. With no
// record, the deployed GPT layout (CLAUDE.md, no .claude/) is the deployed
// harness, and a claude resume over it is refused.
func TestInitResumeRefusesGPTLayoutWithoutRecord(t *testing.T) {
	resumeHome(t)
	projectDir := filepath.Join(t.TempDir(), "gpt-layout-proj")
	resumeSeedRun(t, projectDir, "")
	resumeForgetRecord(t, projectDir)
	makeLegacyGPTLayout(t, projectDir)
	llmBefore := readResumeLLMYAML(t, projectDir)
	manifestBefore := readUserManifestForTest(t)

	err := resumeRun(t, projectDir, "")
	requireResumeRefused(t, err, projectDir, llmBefore, manifestBefore, "", "gpt", "claude")
}

// TestInitResumeAcceptsMatchingHarness — the guard. A claude record, and a
// claude-side layout with no record, still resume: the refusal fires only on
// a mismatch.
func TestInitResumeAcceptsMatchingHarness(t *testing.T) {
	for _, tc := range []struct {
		name   string
		forget bool
	}{
		{name: "recorded claude"},
		{name: "no record, claude-side layout", forget: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resumeHome(t)
			projectDir := filepath.Join(t.TempDir(), "matching-proj")
			resumeSeedRun(t, projectDir, "")
			requireClaudeSideLayout(t, projectDir)
			if tc.forget {
				resumeForgetRecord(t, projectDir)
			}

			if err := resumeRun(t, projectDir, ""); err != nil {
				t.Fatalf("the resume refused a claude project (%s): %v", tc.name, err)
			}
			if llm := readResumeLLMYAML(t, projectDir); !hasHarness(llm, "claude") {
				t.Fatalf("the matching resume did not record claude — llm.yaml:\n%s", llm)
			}
		})
	}
}
