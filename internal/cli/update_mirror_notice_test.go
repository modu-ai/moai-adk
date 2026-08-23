package cli

// Mirror-notice wiring on the two `moai update` paths
// (SPEC-DEPLOY-RESULT-WIRE-001).
//
// Both paths hold a stdout writer for progress output, so the notice — a
// warning — needs a stderr writer of its own. Asserting only that the notice
// appears would pass on an implementation that writes it to both streams,
// which is exactly what internal/cli/CLAUDE.md:14 forbids ("Never mix"), so
// every stream arm below asserts presence on stderr AND absence on stdout.
//
// The doubles here stand in for the deployer because the real fallback is
// driven by a symlink syscall failure whose injection seams are unexported in
// internal/template. DeployResult, SkillMirrorEntry and the MirrorMode
// constants are all exported, so a double can report any outcome faithfully.

import (
	"bytes"
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/mirrornotice"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/spf13/cobra"
)

// --- doubles ---------------------------------------------------------------

// mirrorReportingDeployer implements template.ResultDeployer and reports the
// mirror entries it was given.
type mirrorReportingDeployer struct {
	entries []template.SkillMirrorEntry
	called  bool
}

func (d *mirrorReportingDeployer) Deploy(ctx context.Context, root string, m manifest.Manager, tc *template.TemplateContext) error {
	_, err := d.DeployWithResult(ctx, root, m, tc)
	return err
}

func (d *mirrorReportingDeployer) DeployWithResult(_ context.Context, _ string, _ manifest.Manager, _ *template.TemplateContext) (*template.DeployResult, error) {
	d.called = true
	return &template.DeployResult{SkillMirrors: d.entries}, nil
}

func (d *mirrorReportingDeployer) ExtractTemplate(_ string) ([]byte, error) { return nil, nil }
func (d *mirrorReportingDeployer) ListTemplates() []string                  { return nil }
func (d *mirrorReportingDeployer) ValidateAll(_ context.Context, _ *template.TemplateContext) error {
	return nil
}

var _ template.ResultDeployer = (*mirrorReportingDeployer)(nil)

// plainDeployer implements template.Deployer ONLY. It deliberately has no
// DeployWithResult method: defining one that returns nil would not reproduce
// the state AC-DRW-005 exists to check (the extension being absent).
type plainDeployer struct{ called bool }

func (d *plainDeployer) Deploy(_ context.Context, _ string, _ manifest.Manager, _ *template.TemplateContext) error {
	d.called = true
	return nil
}
func (d *plainDeployer) ExtractTemplate(_ string) ([]byte, error) { return nil, nil }
func (d *plainDeployer) ListTemplates() []string                  { return nil }
func (d *plainDeployer) ValidateAll(_ context.Context, _ *template.TemplateContext) error {
	return nil
}

func mirrorEntries(mode template.MirrorMode, n int) []template.SkillMirrorEntry {
	out := make([]template.SkillMirrorEntry, 0, n)
	for i := range n {
		out = append(out, template.SkillMirrorEntry{
			Skill:   fmt.Sprintf("moai-skill-%02d", i),
			Mode:    mode,
			Warning: fmt.Sprintf("symlink and copy both failed for moai-skill-%02d: permission denied", i),
		})
	}
	return out
}

// --- clean-reinstall arm ---------------------------------------------------

// makeMirrorNoticeV2Fixture builds the minimal v2 project that makes
// runCleanReinstall reach its deploy step.
func makeMirrorNoticeV2Fixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeTestFile(t, root, ".moai/config/sections/system.yaml", "moai:\n    version: v2.16.1\n")
	writeTestFile(t, root, ".claude/agents/moai/manager-strategy.md", "retired\n")
	writeTestFile(t, root, ".claude/settings.json", `{"model": "opus"}`+"\n")
	return root
}

// runCleanReinstallCapturing runs the clean-reinstall path wired the way
// production wires it — stdout injected into Out (update.go does exactly this)
// — and returns what each stream received.
func runCleanReinstallCapturing(t *testing.T, dep template.Deployer) (stdout, stderr string) {
	t.Helper()
	root := makeMirrorNoticeV2Fixture(t)
	var outBuf, errBuf bytes.Buffer
	migrate := &stubMigrateRunner{}

	if _, err := runCleanReinstall(context.Background(), root, CleanReinstallOptions{
		Out:              &outBuf,
		ErrOut:           &errBuf,
		Deployer:         dep,
		RunMigrateAgency: migrate.Run,
	}); err != nil {
		t.Fatalf("runCleanReinstall: %v\nstdout: %s\nstderr: %s", err, outBuf.String(), errBuf.String())
	}
	return outBuf.String(), errBuf.String()
}

// AC-DRW-003 (clean-reinstall arm) + AC-DRW-007 (clean-reinstall arm): the
// notice reaches stderr and does NOT reach the stdout writer production
// injects into Out.
func TestCleanReinstall_MirrorNoticeGoesToStderrNotStdout(t *testing.T) {
	dep := &mirrorReportingDeployer{entries: mirrorEntries(template.MirrorModeCopy, 11)}
	stdout, stderr := runCleanReinstallCapturing(t, dep)

	if !dep.called {
		t.Fatal("deployer was never invoked; the assertions below would be vacuous")
	}
	if !strings.Contains(stderr, mirrornotice.Token) {
		t.Errorf("mirror notice absent from stderr; the clean-reinstall path is silent about a copy fallback\nstderr: %s", stderr)
	}
	if !strings.Contains(stderr, "11") {
		t.Errorf("mirror notice does not carry the fallback count 11\nstderr: %s", stderr)
	}
	if strings.Contains(stdout, mirrornotice.Token) {
		t.Errorf("mirror notice written to stdout, violating the stdout/stderr split (internal/cli/CLAUDE.md:14)\nstdout: %s", stdout)
	}
}

// AC-DRW-002 (clean-reinstall arm): nothing on either stream when nothing fell
// back.
func TestCleanReinstall_NoFallbackEmitsNothing(t *testing.T) {
	dep := &mirrorReportingDeployer{entries: mirrorEntries(template.MirrorModeSymlink, 34)}
	stdout, stderr := runCleanReinstallCapturing(t, dep)

	if strings.Contains(stderr, mirrornotice.Token) {
		t.Errorf("symlink-only run wrote a mirror notice to stderr: %s", stderr)
	}
	if strings.Contains(stdout, mirrornotice.Token) {
		t.Errorf("symlink-only run wrote a mirror notice to stdout: %s", stdout)
	}
}

// AC-DRW-006: the failure count and between one and three warnings reach the
// user, and a failed mirror does not fail the deployment (the predecessor
// SPEC's fail-open contract).
func TestCleanReinstall_FailedMirrorsReportedAndFailOpen(t *testing.T) {
	dep := &mirrorReportingDeployer{entries: mirrorEntries(template.MirrorModeFailed, 9)}
	stdout, stderr := runCleanReinstallCapturing(t, dep) // fatals on error → fail-open arm

	if !strings.Contains(stderr, "9") {
		t.Errorf("failure count 9 absent from stderr: %s", stderr)
	}
	quoted := strings.Count(stderr, "permission denied")
	if quoted < 1 || quoted > 3 {
		t.Errorf("stderr quotes %d failure warnings, want between 1 and 3: %s", quoted, stderr)
	}
	if strings.Contains(stdout, mirrornotice.Token) {
		t.Errorf("failure notice written to stdout: %s", stdout)
	}
}

// AC-DRW-008 (clean-reinstall arm): the misattributing skipped warning stays
// off both streams.
func TestCleanReinstall_SkippedWarningsNotSurfaced(t *testing.T) {
	dep := &mirrorReportingDeployer{entries: []template.SkillMirrorEntry{{
		Skill:   "moai-a",
		Mode:    template.MirrorModeSkipped,
		Warning: "a non-symlink entry already exists at .agents/skills/moai-a — left untouched",
	}}}
	stdout, stderr := runCleanReinstallCapturing(t, dep)

	for name, stream := range map[string]string{"stdout": stdout, "stderr": stderr} {
		if strings.Contains(stream, "a non-symlink entry already exists") {
			t.Errorf("misattributing skipped warning reached %s: %s", name, stream)
		}
	}
}

// AC-DRW-005: a deployer without the optional extension deploys normally — no
// error, no panic, no notice.
func TestCleanReinstall_PlainDeployerStillDeploys(t *testing.T) {
	dep := &plainDeployer{}
	if _, isResultDeployer := any(dep).(template.ResultDeployer); isResultDeployer {
		t.Fatal("plainDeployer implements ResultDeployer; this test can no longer reproduce the extension-absent state")
	}

	stdout, stderr := runCleanReinstallCapturing(t, dep)
	if !dep.called {
		t.Fatal("plain deployer was not invoked; deployment did not happen")
	}
	for name, stream := range map[string]string{"stdout": stdout, "stderr": stderr} {
		if strings.Contains(stream, mirrornotice.Token) {
			t.Errorf("a deployer without the result extension produced a mirror notice on %s: %s", name, stream)
		}
	}
}

// --- template-sync arm -----------------------------------------------------

// runTemplateSyncCapturing runs the template-sync path with the deployer seam
// substituted, wired as production wires it: out stays cmd.OutOrStdout().
func runTemplateSyncCapturing(t *testing.T, dep template.Deployer) (stdout, stderr string) {
	t.Helper()

	tmpDir := t.TempDir()
	writeTestFile(t, tmpDir, ".moai/config/sections/system.yaml", "system:\n  template_version: \"0.0.0\"\n")
	writeTestFile(t, tmpDir, ".moai/manifest.json", "{}")

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })
	t.Setenv("HOME", tmpDir)

	prev := newTemplateSyncDeployer
	newTemplateSyncDeployer = func(fs.FS) template.Deployer { return dep }
	t.Cleanup(func() { newTemplateSyncDeployer = prev })

	var outBuf, errBuf bytes.Buffer
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Bool("force", false, "")
	cmd.Flags().Bool("yes", true, "")
	cmd.Flags().Bool("config", false, "")
	_ = cmd.Flags().Set("yes", "true")
	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)
	cmd.SetContext(context.Background())

	if err := runTemplateSyncWithReporter(cmd, nil, true); err != nil {
		t.Fatalf("runTemplateSyncWithReporter: %v\nstdout: %s\nstderr: %s", err, outBuf.String(), errBuf.String())
	}
	return outBuf.String(), errBuf.String()
}

// AC-DRW-003 (template-sync arm) + AC-DRW-007 (template-sync arm): the notice
// reaches stderr while the path's own writer stays stdout.
func TestTemplateSync_MirrorNoticeGoesToStderrNotStdout(t *testing.T) {
	dep := &mirrorReportingDeployer{entries: mirrorEntries(template.MirrorModeCopy, 7)}
	stdout, stderr := runTemplateSyncCapturing(t, dep)

	if !dep.called {
		t.Fatal("deployer was never invoked; the assertions below would be vacuous")
	}
	if !strings.Contains(stderr, mirrornotice.Token) {
		t.Errorf("mirror notice absent from stderr; the template-sync path is silent about a copy fallback\nstderr: %s", stderr)
	}
	if !strings.Contains(stderr, "7") {
		t.Errorf("mirror notice does not carry the fallback count 7\nstderr: %s", stderr)
	}
	if strings.Contains(stdout, mirrornotice.Token) {
		t.Errorf("mirror notice written to stdout, violating the stdout/stderr split (internal/cli/CLAUDE.md:14)\nstdout: %s", stdout)
	}
}

// AC-DRW-002 (template-sync arm).
func TestTemplateSync_NoFallbackEmitsNothing(t *testing.T) {
	dep := &mirrorReportingDeployer{entries: mirrorEntries(template.MirrorModeSymlink, 34)}
	stdout, stderr := runTemplateSyncCapturing(t, dep)

	if strings.Contains(stderr, mirrornotice.Token) {
		t.Errorf("symlink-only run wrote a mirror notice to stderr: %s", stderr)
	}
	if strings.Contains(stdout, mirrornotice.Token) {
		t.Errorf("symlink-only run wrote a mirror notice to stdout: %s", stdout)
	}
}

// AC-DRW-005 (template-sync arm).
func TestTemplateSync_PlainDeployerStillDeploys(t *testing.T) {
	dep := &plainDeployer{}
	stdout, stderr := runTemplateSyncCapturing(t, dep)

	if !dep.called {
		t.Fatal("plain deployer was not invoked; deployment did not happen")
	}
	for name, stream := range map[string]string{"stdout": stdout, "stderr": stderr} {
		if strings.Contains(stream, mirrornotice.Token) {
			t.Errorf("a deployer without the result extension produced a mirror notice on %s: %s", name, stream)
		}
	}
}

// --- AC-DRW-009: the production wiring itself -------------------------------
//
// Every other criterion injects something. This one looks at the configuration
// nothing is injected into: if the seam's default deployer cannot report a
// result, or the production option literals never pass a stderr writer, the
// notice is unreachable in production while every injected-double test above
// stays green.

// AC-DRW-009 assertion 1 — the seam's default builds the production deployer:
// the embedded FS, a renderer, and forceUpdate=true, exactly as the inline
// construction it replaced did.
func TestSeamDefaultIsTheProductionDeployer(t *testing.T) {
	t.Parallel()

	file := parseCLIFile(t, "update_template_sync.go")
	value := seamVarValue(t, file, "newTemplateSyncDeployer")

	var found *ast.CallExpr
	ast.Inspect(value, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if sel, ok := call.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "NewDeployerWithRendererAndForceUpdate" {
			found = call
			return false
		}
		return true
	})

	if found == nil {
		t.Fatal("newTemplateSyncDeployer default does not construct the production deployer " +
			"(no template.NewDeployerWithRendererAndForceUpdate call); a test-shaped default would leave " +
			"production permanently silent while every injected-double test stayed green")
	}
	if len(found.Args) != 3 {
		t.Fatalf("NewDeployerWithRendererAndForceUpdate called with %d args, want 3", len(found.Args))
	}
	if ident, ok := found.Args[2].(*ast.Ident); !ok || ident.Name != "true" {
		t.Errorf("forceUpdate argument is %s, want true — template sync must overwrite existing files",
			exprString(found.Args[2]))
	}
}

// AC-DRW-009 assertion 2 — the seam's default actually satisfies the optional
// extension. Assertion 1 proves the source names the production constructor;
// only calling it proves the value it returns can report a result at all.
func TestSeamDefaultSatisfiesResultDeployer(t *testing.T) {
	t.Parallel()

	dep := newTemplateSyncDeployer(fstest.MapFS{})
	if dep == nil {
		t.Fatal("newTemplateSyncDeployer returned nil")
	}
	if _, ok := dep.(template.ResultDeployer); !ok {
		t.Fatalf("the seam's default deployer (%T) does not implement template.ResultDeployer; "+
			"production would never emit a mirror notice", dep)
	}
}

// AC-DRW-009 assertion 3 — every production CleanReinstallOptions literal
// passes a stderr writer. The literals are collected rather than counted, so a
// third call site added later is judged automatically instead of silently
// escaping the guard.
func TestCleanReinstallOptionsLiteralsInjectErrOut(t *testing.T) {
	t.Parallel()

	file := parseCLIFile(t, "update.go")

	var literals []*ast.CompositeLit
	ast.Inspect(file, func(n ast.Node) bool {
		lit, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}
		if ident, ok := lit.Type.(*ast.Ident); ok && ident.Name == "CleanReinstallOptions" {
			literals = append(literals, lit)
		}
		return true
	})

	if len(literals) == 0 {
		t.Fatal("no CleanReinstallOptions literal found in update.go; the guard is looking at the wrong place")
	}

	for _, lit := range literals {
		var errOut ast.Expr
		for _, elt := range lit.Elts {
			kv, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			if key, ok := kv.Key.(*ast.Ident); ok && key.Name == "ErrOut" {
				errOut = kv.Value
			}
		}
		if errOut == nil {
			t.Errorf("a CleanReinstallOptions literal in update.go has no ErrOut field; "+
				"the mirror notice would fall back to a writer the caller never chose. Literal fields: %v",
				literalKeys(lit))
			continue
		}
		if got := exprString(errOut); got != "cmd.ErrOrStderr()" {
			t.Errorf("CleanReinstallOptions.ErrOut = %s, want cmd.ErrOrStderr() — "+
				"the notice is a warning and must not ride the stdout writer Out carries", got)
		}
	}
}

// --- AST helpers ------------------------------------------------------------

func parseCLIFile(t *testing.T, name string) *ast.File {
	t.Helper()
	path := filepath.Join(".", name)
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return file
}

// seamVarValue returns the initializer expression of a package-level var.
func seamVarValue(t *testing.T, file *ast.File, name string) ast.Expr {
	t.Helper()
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.VAR {
			continue
		}
		for _, spec := range gen.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for i, ident := range vs.Names {
				if ident.Name == name && i < len(vs.Values) {
					return vs.Values[i]
				}
			}
		}
	}
	t.Fatalf("package-level var %s with an initializer not found", name)
	return nil
}

// exprString renders an expression back to source text for assertions and
// failure messages.
func exprString(e ast.Expr) string {
	var b strings.Builder
	writeExpr(&b, e)
	return b.String()
}

func writeExpr(b *strings.Builder, e ast.Expr) {
	switch v := e.(type) {
	case *ast.Ident:
		b.WriteString(v.Name)
	case *ast.SelectorExpr:
		writeExpr(b, v.X)
		b.WriteString(".")
		b.WriteString(v.Sel.Name)
	case *ast.CallExpr:
		writeExpr(b, v.Fun)
		b.WriteString("(")
		for i, arg := range v.Args {
			if i > 0 {
				b.WriteString(", ")
			}
			writeExpr(b, arg)
		}
		b.WriteString(")")
	case *ast.BasicLit:
		b.WriteString(v.Value)
	default:
		fmt.Fprintf(b, "%T", e)
	}
}

func literalKeys(lit *ast.CompositeLit) []string {
	var keys []string
	for _, elt := range lit.Elts {
		if kv, ok := elt.(*ast.KeyValueExpr); ok {
			if ident, ok := kv.Key.(*ast.Ident); ok {
				keys = append(keys, ident.Name)
			}
		}
	}
	return keys
}
