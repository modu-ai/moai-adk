package project

// SPEC-INIT-DEPLOY-EXIT-001 — template deployment is the one degradation site
// in Init() that is fatal (REQ-IDE-001). The other seven sites stay warnings
// (REQ-IDE-005), and the mirror notice recorded before the failure must survive
// the fatal return (acceptance.md §D.1).

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/mirrornotice"
	"github.com/modu-ai/moai-adk/internal/shell"
	"github.com/modu-ai/moai-adk/internal/template"
)

// failingResultDeployer reports mirror entries AND fails, reproducing the real
// deployer's behaviour: the result is populated even when deployment errors.
type failingResultDeployer struct {
	entries []template.SkillMirrorEntry
	err     error
}

func (d *failingResultDeployer) Deploy(ctx context.Context, root string, m manifest.Manager, tc *template.TemplateContext) error {
	_, err := d.DeployWithResult(ctx, root, m, tc)
	return err
}

func (d *failingResultDeployer) DeployWithResult(_ context.Context, _ string, _ manifest.Manager, _ *template.TemplateContext) (*template.DeployResult, error) {
	return &template.DeployResult{SkillMirrors: d.entries}, d.err
}

func (d *failingResultDeployer) ExtractTemplate(_ string) ([]byte, error) { return nil, nil }
func (d *failingResultDeployer) ListTemplates() []string                  { return nil }
func (d *failingResultDeployer) ValidateAll(_ context.Context, _ *template.TemplateContext) error {
	return nil
}

var _ template.ResultDeployer = (*failingResultDeployer)(nil)

// AC-IDE-003 — the fatal return still carries the InitResult, so warnings
// recorded before the failure (here the skill-mirror copy-fallback notice)
// still reach the CLI's warning summary instead of being dropped with the
// result.
func TestInit_DeployFailureKeepsResultWarnings(t *testing.T) {
	root := t.TempDir()
	dep := &failingResultDeployer{
		entries: []template.SkillMirrorEntry{
			{Skill: "moai-a", Mode: template.MirrorModeCopy},
		},
		err: errors.New(`template render "x.tmpl": unexpanded dynamic token detected`),
	}

	result, err := NewInitializer(dep, manifest.NewManager(), nil).
		Init(context.Background(), mirrorNoticeInitOptions(root))
	if err == nil {
		t.Fatal("Init() returned nil on a deploy failure; deployment failure must be fatal")
	}
	if !strings.Contains(err.Error(), "template render") {
		t.Errorf("Init() error lost the failing template path: %v", err)
	}
	if result == nil {
		t.Fatal("Init() dropped the InitResult on the fatal path; warnings recorded before the failure are lost")
	}
	if joined := strings.Join(result.Warnings, "\n"); !strings.Contains(joined, mirrornotice.Token) {
		t.Errorf("mirror notice absent from the returned warnings: %v", result.Warnings)
	}
}

// AC-IDE-004 — a degradation outside template deployment (here Step 6 shell
// configuration) stays a warning and leaves Init()'s return nil.
func TestInit_NonDeployFailureStaysWarning(t *testing.T) {
	orig := ConfigureShellEnvFn
	ConfigureShellEnvFn = func(_ *slog.Logger) (*shell.ConfigResult, error) {
		return nil, errors.New("shell rc unwritable")
	}
	t.Cleanup(func() { ConfigureShellEnvFn = orig })

	root := t.TempDir()
	opts := mirrorNoticeInitOptions(root)
	opts.SkipShellConfig = false

	result, err := NewInitializer(&mockDeployer{}, manifest.NewManager(), nil).
		Init(context.Background(), opts)
	if err != nil {
		t.Fatalf("Init() = %v; a non-deployment degradation must stay a warning", err)
	}
	joined := strings.Join(result.Warnings, "\n")
	if !strings.Contains(joined, "shell configuration") {
		t.Errorf("shell configuration warning missing: %v", result.Warnings)
	}
}

// acceptance.md §D.1 — the deployer-less fallback path is a different branch
// and is unchanged: its failures stay warnings and Init() returns nil.
func TestInit_NilDeployerFallbackStaysNonFatal(t *testing.T) {
	root := t.TempDir()

	if _, err := NewInitializer(nil, manifest.NewManager(), nil).
		Init(context.Background(), mirrorNoticeInitOptions(root)); err != nil {
		t.Fatalf("Init() with no deployer = %v; the fallback branch must stay non-fatal", err)
	}
}
