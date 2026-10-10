package cli

// init_resume_refusal_test.go — Item C (card t1591, the finding on 01e510335):
// a harness change on the resume is refused, not half-applied.
//
// The resume restores the content the failed run never wrote and deploys no
// template files. A resume that rewrote llm.yaml to another harness would report
// success with the old harness's files and without the new harness's files. The
// deployed harness is the selection record the failing run left, or the deployed
// layout when no record exists. A requested harness that contradicts it is refused
// before any side effect.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/userassets"
)

// requireClaudeSideLayout asserts the deployed layout is the claude-side one: the
// .claude/ surface is present, so the layout reads as claude or both.
func requireClaudeSideLayout(t *testing.T, projectDir string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(projectDir, ".claude")); err != nil {
		t.Fatalf("precondition: the claude seed run did not deploy the .claude/ surface: %v", err)
	}
}

// requireResumeRefused asserts the refusal contract for a harness change on the
// resume. deployed is the harness the refusal must name as deployed, requested is
// the harness the resume asked for, and record is the selection record the failing
// run left ("" when none). The refusal precedes every side effect, so it is checked
// as an absence: llm.yaml is byte-identical, the resume marker was never written,
// the corrupt user manifest the ensure would quarantine is still in place, and the
// selection record is not retired.
func requireResumeRefused(t *testing.T, err error, projectDir, llmBefore, record, deployed, requested string) {
	t.Helper()
	if err == nil {
		t.Fatalf("the resume reported success over a harness change (deployed %s, requested %s): no refusal", deployed, requested)
	}
	msg := err.Error()
	for _, want := range []string{deployed, requested, "--llm " + deployed, "--force"} {
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
	data, readErr := os.ReadFile(userassets.ManifestPath(os.Getenv("HOME")))
	if readErr != nil || string(data) != "{ corrupt\n" {
		t.Errorf("the refused resume touched the user manifest the ensure would quarantine (read err = %v, bytes = %q)", readErr, data)
	}
	if record != "" {
		if got, ok := readResumeWiring(projectDir); !ok || string(got) != record {
			t.Errorf("the refused resume retired the selection record: got %q (ok=%v), want %q", got, ok, record)
		}
	}
}

// TestInitResumeRefusesHarnessChange — Item C. The resume's requested harness
// contradicts the deployed one, so the resume is refused before any side effect.
func TestInitResumeRefusesHarnessChange(t *testing.T) {
	cases := []struct {
		name string
		// seed is the --llm of the failing run; layout checks the layout it left.
		seed   string
		layout func(*testing.T, string)
		// forget drops the selection record before the resume.
		forget bool
		// resume is the --llm of the resume; record is the record it must find.
		resume string
		record string
		// deployed is the harness the refusal must name as deployed.
		deployed string
	}{
		{name: "recorded gpt, explicit claude", seed: "gpt", layout: requireGPTLayout, resume: "claude", record: "gpt", deployed: "gpt"},
		{name: "no record, GPT layout, explicit claude", seed: "gpt", layout: requireGPTLayout, forget: true, resume: "claude", deployed: "gpt"},
		{name: "no record, claude-side layout, explicit gpt", seed: "claude", layout: requireClaudeSideLayout, forget: true, resume: "gpt", deployed: "claude"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resumeHome(t)
			projectDir := filepath.Join(t.TempDir(), "refusal-proj")
			resumeSeedRun(t, projectDir, tc.seed)
			tc.layout(t, projectDir)
			if tc.forget {
				resumeForgetRecord(t, projectDir)
			}
			llmBefore := readResumeLLMYAML(t, projectDir)

			err := resumeRun(t, projectDir, tc.resume)
			requireResumeRefused(t, err, projectDir, llmBefore, tc.record, tc.deployed, tc.resume)
		})
	}
}

// TestInitResumeAcceptsMatchingHarness — Item C, the guard. A requested harness that
// matches the deployed one still resumes: the refusal fires only on a mismatch.
func TestInitResumeAcceptsMatchingHarness(t *testing.T) {
	cases := []struct {
		name   string
		seed   string
		layout func(*testing.T, string)
		forget bool
		resume string
	}{
		{name: "recorded gpt, explicit gpt", seed: "gpt", layout: requireGPTLayout, resume: "gpt"},
		{name: "no record, claude-side layout, explicit claude", seed: "claude", layout: requireClaudeSideLayout, forget: true, resume: "claude"},
		{name: "recorded both, explicit both", seed: "both", layout: requireClaudeSideLayout, resume: "both"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resumeHome(t)
			projectDir := filepath.Join(t.TempDir(), "matching-proj")
			resumeSeedRun(t, projectDir, tc.seed)
			tc.layout(t, projectDir)
			if tc.forget {
				resumeForgetRecord(t, projectDir)
			}

			if err := resumeRun(t, projectDir, tc.resume); err != nil {
				t.Fatalf("the resume refused a matching harness (%s): %v", tc.resume, err)
			}
			if llm := readResumeLLMYAML(t, projectDir); !hasHarness(llm, tc.resume) {
				t.Fatalf("the matching resume did not record %s — llm.yaml:\n%s", tc.resume, llm)
			}
		})
	}
}
