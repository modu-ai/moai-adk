package cli

// update_mcp_snapshot_test.go pins where the real flows call the .mcp.json
// base-snapshot lifecycle (card t1029), the sibling of
// update_settings_snapshot_test.go.
//
// The three staging sites are NOT symmetric for this file, and the asymmetry is
// measured here rather than assumed: the template-sync flow MERGES .mcp.json,
// while the clean-reinstall flow deliberately excludes it from its mergeable
// set (update_clean_install.go: "the template-sync collectMergeableFiles set
// minus .mcp.json") and force-deploys it instead. Each cell below states which
// of the two it exercises.
//
// Isolation: every test works under t.TempDir() and injects the home through
// homeSeamSpy, never by overriding the HOME environment variable, so none of
// these tests may call t.Parallel().

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/template"
)

const (
	mcpSnapLiveRel      = ".mcp.json"
	mcpSnapCanonicalRel = ".moai/cache/template-snapshot/mcp.json"
	mcpSnapPendingRel   = ".moai/cache/template-snapshot/mcp.json.pending"
)

// mcpRenderDeployer is a deployer double that behaves like the real one for
// .mcp.json only: it writes render there and records the file as
// template-managed with the render hash, which is what the staging helper reads
// to decide the deploy wrote the file. It models a FORCE deploy — it overwrites
// whatever is there — which is what the clean-reinstall flow performs.
type mcpRenderDeployer struct {
	render string
	calls  int
}

func (d *mcpRenderDeployer) Deploy(_ context.Context, projectRoot string, mgr manifest.Manager, _ *template.TemplateContext) error {
	d.calls++
	path := filepath.Join(projectRoot, filepath.FromSlash(mcpSnapLiveRel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(d.render), 0o644); err != nil {
		return err
	}
	return mgr.Track(mcpSnapLiveRel, manifest.TemplateManaged, manifest.HashBytes([]byte(d.render)))
}

func (d *mcpRenderDeployer) ListTemplates() []string { return nil }

func (d *mcpRenderDeployer) ValidateAll(context.Context, *template.TemplateContext) error {
	return nil
}

func (d *mcpRenderDeployer) ExtractTemplate(string) ([]byte, error) { return nil, nil }

func mcpSnapAssertCanonical(t *testing.T, root, want string) {
	t.Helper()
	got, ok := snapRead(t, root, mcpSnapCanonicalRel)
	if !ok {
		t.Errorf("canonical .mcp.json snapshot absent, want %s", want)
		return
	}
	if got != want {
		t.Errorf("canonical .mcp.json snapshot = %s, want %s", got, want)
	}
}

func mcpSnapAssertNoPending(t *testing.T, root string) {
	t.Helper()
	if _, ok := snapRead(t, root, mcpSnapPendingRel); ok {
		t.Errorf(".mcp.json staging copy still present after the flow")
	}
}

// The three renders below model the real precedent f770c0211: context7's
// launcher changed from /bin/bash to npx while a new moai server was added.
const (
	mcpFlowR1 = `{"mcpServers":{"context7":{"command":"/bin/bash"}}}`
	mcpFlowR2 = `{"mcpServers":{"context7":{"command":"npx"},"moai":{"command":"moai"}}}`
)

// The template-sync flow — the one that MERGES .mcp.json — stages the render it
// deployed and promotes it, so the value change reaches the next update.
func TestMCPSnapshot_TemplateSyncStagesAndPromotes(t *testing.T) {
	sentinel, _ := homeSeamSpy(t)
	snapAssertSandboxHome(t, sentinel)

	root := t.TempDir()
	snapWrite(t, root, ".moai/manifest.json", `{"version":"1","files":{}}`)
	snapWrite(t, root, mcpSnapCanonicalRel, mcpFlowR1)
	// The user carries the prior render plus one server of their own.
	snapWrite(t, root, mcpSnapLiveRel,
		`{"mcpServers":{"context7":{"command":"/bin/bash"},"mine":{"command":"my-own"}}}`)

	d := &mcpRenderDeployer{render: mcpFlowR2}
	out, _ := snapRunTemplateSync(t, root, d)
	if d.calls != 1 {
		t.Fatalf("deploy calls = %d, want 1 (the sync did not run)\n%s", d.calls, out)
	}

	live, ok := snapRead(t, root, mcpSnapLiveRel)
	if !ok {
		t.Fatalf("live .mcp.json is absent")
	}
	doc := snapDecode(t, live)
	servers, _ := doc["mcpServers"].(map[string]any)
	ctx7, _ := servers["context7"].(map[string]any)

	// K2 — the template's value change reached a key the user never touched.
	if ctx7["command"] != "npx" {
		t.Errorf("K2 context7.command = %v, want npx (live %s)", ctx7["command"], live)
	}
	// K1 — the newly added server arrived.
	if servers["moai"] == nil {
		t.Errorf("K1 mcpServers.moai absent (live %s)", live)
	}
	// K4 — the user's own server survived.
	if servers["mine"] == nil {
		t.Errorf("K4 mcpServers.mine lost (live %s)", live)
	}

	mcpSnapAssertCanonical(t, root, mcpFlowR2)
	mcpSnapAssertNoPending(t, root)
}

// The clean-reinstall flow — the one that does NOT merge .mcp.json — still
// stages and promotes the render it force-deployed.
//
// The decision, and why it is not symmetry with settings.json: after this flow
// the live .mcp.json IS the render byte-for-byte, because the force deploy
// overwrote it and no merge followed. Recording it therefore records a TRUE
// base. Staging nothing would leave the canonical copy holding an older render
// while the live file holds a newer one, and the next update would read the
// difference between the two renders as a user edit — re-applying a stale value
// the user never chose. The measurement below is the one that decides it: the
// live file equals the render, which is the precondition promotion requires.
func TestMCPSnapshot_CleanReinstallStagesTheForcedRender(t *testing.T) {
	sentinel, _ := homeSeamSpy(t)
	snapAssertSandboxHome(t, sentinel)

	root := makeScenarioA(t)
	snapWrite(t, root, mcpSnapCanonicalRel, mcpFlowR1)
	snapWrite(t, root, mcpSnapLiveRel,
		`{"mcpServers":{"context7":{"command":"/bin/bash"},"mine":{"command":"my-own"}}}`)

	d := &mcpRenderDeployer{render: mcpFlowR2}
	snapRunCleanReinstall(t, root, d)
	if d.calls != 1 {
		t.Fatalf("deploy calls = %d, want 1", d.calls)
	}

	live, ok := snapRead(t, root, mcpSnapLiveRel)
	if !ok {
		t.Fatalf("live .mcp.json is absent")
	}
	// The precondition for promoting: this flow leaves the render on disk
	// untouched by any merge. If this ever stops holding — because the flow
	// starts merging .mcp.json — the promotion below is no longer correct and
	// this assertion is what says so.
	if live != mcpFlowR2 {
		t.Fatalf("live .mcp.json = %s, want the render %s — the clean-reinstall flow "+
			"no longer leaves the forced render in place, so promoting it is wrong", live, mcpFlowR2)
	}
	mcpSnapAssertCanonical(t, root, mcpFlowR2)
	mcpSnapAssertNoPending(t, root)
}

// A staging copy an interrupted earlier flow left behind is settled by the next
// runUpdate before anything else touches the file.
func TestMCPSnapshot_UpdateJudgesLeftover(t *testing.T) {
	sentinel, _ := homeSeamSpy(t)
	snapAssertSandboxHome(t, sentinel)

	root := snapNewUpdateProject(t, "0.0.0")
	snapWrite(t, root, mcpSnapCanonicalRel, mcpFlowR1)
	// An earlier flow stopped after its deploy: the leftover and the live file
	// agree, so nothing reverted the render and the leftover is promoted.
	snapWrite(t, root, mcpSnapPendingRel, mcpFlowR2)
	snapWrite(t, root, mcpSnapLiveRel, mcpFlowR2)

	d := &mcpRenderDeployer{render: mcpFlowR2}
	snapUseSyncDeployer(t, d)

	out, errOut, err := snapRunUpdate(t, root)
	if err != nil {
		t.Fatalf("runUpdate: %v\nout:\n%s\nerr:\n%s", err, out, errOut)
	}
	mcpSnapAssertCanonical(t, root, mcpFlowR2)
	mcpSnapAssertNoPending(t, root)
}

// A staging or promotion failure never blocks a flow, and its warning never
// reads as a settings.json failure.
func TestMCPSnapshot_WriteFailureDoesNotBlock(t *testing.T) {
	sentinel, _ := homeSeamSpy(t)
	snapAssertSandboxHome(t, sentinel)

	root := t.TempDir()
	snapWrite(t, root, ".moai/manifest.json", `{"version":"1","files":{}}`)
	snapWrite(t, root, mcpSnapLiveRel, mcpFlowR1)
	// A file where the snapshot directory must go makes the staging mkdir fail.
	snapWrite(t, root, ".moai/cache/template-snapshot", "not a directory")

	d := &mcpRenderDeployer{render: mcpFlowR2}
	out, errOut := snapRunTemplateSync(t, root, d)
	if d.calls != 1 {
		t.Fatalf("deploy calls = %d, want 1 — the flow did not reach the deploy\n%s", d.calls, out)
	}
	if n := snapCountPrefixed(errOut, "mcp-snapshot-"); n == 0 {
		t.Errorf("no mcp-snapshot warning on stderr:\n%s", errOut)
	}
}
