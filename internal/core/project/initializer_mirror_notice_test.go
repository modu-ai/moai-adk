package project

// Mirror-notice wiring on the `moai init` path (SPEC-DEPLOY-RESULT-WIRE-001).
//
// The deployer reports a skill-mirror copy fallback by return value and never
// prints. Until this wiring existed nothing on the init path read that result,
// so a user whose system cannot create symlinks — the most common first
// encounter with MoAI-ADK on a restricted Windows account — got a silently
// degraded install.
//
// The notice reaches the user through InitResult.Warnings, which
// internal/cli/init.go already collects and renders to stderr in its summary
// panel, so this file asserts on that slice.

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/mirrornotice"
	"github.com/modu-ai/moai-adk/internal/template"
)

// mirrorResultDeployer implements template.ResultDeployer and reports whatever
// mirror entries the test hands it.
type mirrorResultDeployer struct {
	entries []template.SkillMirrorEntry
	called  bool
}

func (d *mirrorResultDeployer) Deploy(ctx context.Context, root string, m manifest.Manager, tc *template.TemplateContext) error {
	_, err := d.DeployWithResult(ctx, root, m, tc)
	return err
}

func (d *mirrorResultDeployer) DeployWithResult(_ context.Context, _ string, _ manifest.Manager, _ *template.TemplateContext) (*template.DeployResult, error) {
	d.called = true
	return &template.DeployResult{SkillMirrors: d.entries}, nil
}

func (d *mirrorResultDeployer) ExtractTemplate(_ string) ([]byte, error) { return nil, nil }
func (d *mirrorResultDeployer) ListTemplates() []string                  { return nil }
func (d *mirrorResultDeployer) ValidateAll(_ context.Context, _ *template.TemplateContext) error {
	return nil
}

// Compile-time guard: this double MUST satisfy the optional extension, or the
// positive arm below would silently exercise the fallback path instead.
var _ template.ResultDeployer = (*mirrorResultDeployer)(nil)

func copyEntries(n int) []template.SkillMirrorEntry {
	out := make([]template.SkillMirrorEntry, 0, n)
	for i := range n {
		out = append(out, template.SkillMirrorEntry{
			Skill: fmt.Sprintf("moai-skill-%02d", i),
			Mode:  template.MirrorModeCopy,
		})
	}
	return out
}

func mirrorNoticeInitOptions(root string) InitOptions {
	return InitOptions{
		ProjectRoot:     root,
		ProjectName:     "mirror-notice-test",
		Language:        "Go",
		UserName:        "testuser",
		ConvLang:        "en",
		DevelopmentMode: "ddd",
	}
}

// AC-DRW-007 (init arm) — the copy-fallback notice reaches the user-visible
// warning channel on the init path, carrying the count.
func TestInit_MirrorCopyFallbackReachesWarnings(t *testing.T) {
	root := t.TempDir()
	dep := &mirrorResultDeployer{entries: copyEntries(9)}

	result, err := NewInitializer(dep, manifest.NewManager(), nil).
		Init(context.Background(), mirrorNoticeInitOptions(root))
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	if !dep.called {
		t.Fatal("deployer was never invoked; the assertion below would be vacuous")
	}

	joined := strings.Join(result.Warnings, "\n")
	if !strings.Contains(joined, mirrornotice.Token) {
		t.Fatalf("mirror notice absent from InitResult.Warnings — the init path is still silent about a copy fallback\nwarnings: %v", result.Warnings)
	}
	if !strings.Contains(joined, "9") {
		t.Errorf("mirror notice does not carry the fallback count 9: %s", joined)
	}
}

// AC-DRW-002 (init arm) — a run with no fallback adds no mirror warning.
func TestInit_NoMirrorFallbackAddsNoWarning(t *testing.T) {
	root := t.TempDir()
	dep := &mirrorResultDeployer{entries: []template.SkillMirrorEntry{
		{Skill: "moai-a", Mode: template.MirrorModeSymlink},
		{Skill: "moai-b", Mode: template.MirrorModeSymlink},
	}}

	result, err := NewInitializer(dep, manifest.NewManager(), nil).
		Init(context.Background(), mirrorNoticeInitOptions(root))
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	if joined := strings.Join(result.Warnings, "\n"); strings.Contains(joined, mirrornotice.Token) {
		t.Errorf("symlink-only run produced a mirror warning: %s", joined)
	}
}

// AC-DRW-008 (init arm) — the misattributing skipped warning does not reach
// InitResult.Warnings. Forwarding DeployResult.Warnings wholesale would put it
// there, so this arm is what separates a selective consumer from a forwarder.
func TestInit_SkippedWarningsAreNotForwarded(t *testing.T) {
	root := t.TempDir()
	dep := &mirrorResultDeployer{entries: []template.SkillMirrorEntry{{
		Skill:   "moai-a",
		Mode:    template.MirrorModeSkipped,
		Warning: "a non-symlink entry already exists at .agents/skills/moai-a — left untouched",
	}}}

	result, err := NewInitializer(dep, manifest.NewManager(), nil).
		Init(context.Background(), mirrorNoticeInitOptions(root))
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	if joined := strings.Join(result.Warnings, "\n"); strings.Contains(joined, "a non-symlink entry already exists") {
		t.Errorf("misattributing skipped warning reached the user: %s", joined)
	}
}

// AC-DRW-005 (init arm) — a deployer that does NOT implement the optional
// extension still deploys: no error, no panic, no notice. capturingDeployer
// implements template.Deployer only, so the type assertion in deployTemplates
// must fail at runtime for this test to mean anything; an unconditional
// assertion would panic here.
func TestInit_PlainDeployerStillDeploys(t *testing.T) {
	root := t.TempDir()
	dep := &capturingDeployer{}

	if _, isResultDeployer := any(dep).(template.ResultDeployer); isResultDeployer {
		t.Fatal("capturingDeployer implements ResultDeployer; this test can no longer reproduce the extension-absent state")
	}

	result, err := NewInitializer(dep, manifest.NewManager(), nil).
		Init(context.Background(), mirrorNoticeInitOptions(root))
	if err != nil {
		t.Fatalf("Init() with a plain Deployer returned an error: %v", err)
	}
	if dep.captured == nil {
		t.Fatal("plain deployer was not invoked; deployment did not happen")
	}
	if joined := strings.Join(result.Warnings, "\n"); strings.Contains(joined, mirrornotice.Token) {
		t.Errorf("a deployer without the result extension produced a mirror notice: %s", joined)
	}
}
