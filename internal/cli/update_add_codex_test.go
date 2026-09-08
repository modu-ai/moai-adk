package cli

// SPEC-UPDATE-ADD-CODEX-001 M1 — `moai update --add-codex`, the sanctioned
// ADDITIVE Codex path for an existing project (REQ-UAC-001..008). The verb
// calls the UNGATED codexwiring.Wire (plan §D1): the existence gate is
// bypassed on this verb's path ONLY, and a REQ-CW-003 validation refusal
// propagates as a hard error (exit ≠ 0). The sibling refresh wrapper's
// best-effort swallowing is explicitly NOT the model here (plan §D1):
// best-effort is allowed only for IO errors, never for the refusal.

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/codexwiring"
)

// addCodexWiring is the test seam form: runUpdate operates on "." so the
// tests drive the path-explicit entry directly.
func addCodexWiring(root string, out, errOut *bytes.Buffer) error {
	return addCodexWiringAt(root, out, errOut)
}

// sha256File returns the hex sha256 of the file at path.
func sha256File(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

// TestUpdateAddCodex_WiresClaudeOnlyProject covers AC-UAC-002: a project
// without any wiring gets all three artifacts created plus the first-trust
// guidance (the wiring-existence gate is bypassed on this verb's path).
func TestUpdateAddCodex_WiresClaudeOnlyProject(t *testing.T) {
	root := t.TempDir()
	var out, warn bytes.Buffer
	if err := addCodexWiring(root, &out, &warn); err != nil {
		t.Fatalf("addCodexWiring: %v", err)
	}
	for _, p := range []string{
		filepath.Join(root, codexwiring.HooksRelPath),
		filepath.Join(root, codexwiring.ConfigRelPath),
		filepath.Join(root, codexwiring.SidecarPath),
	} {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("artifact %s not created: %v", p, err)
		}
	}
	cfg, err := os.ReadFile(filepath.Join(root, codexwiring.ConfigRelPath))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(cfg), "mcp_servers.moai") {
		t.Errorf("config.toml missing [mcp_servers.moai]:\n%s", cfg)
	}
	if !strings.Contains(out.String(), codexwiring.FirstTrustGuidance) {
		t.Errorf("first-trust guidance not printed:\n%s", out.String())
	}
}

// TestUpdateAddCodex_McpJsonUntouched covers AC-UAC-004 (decision D2): the
// verb never reads, writes, or repositions .mcp.json — codex MCP rides
// .codex/config.toml.
func TestUpdateAddCodex_McpJsonUntouched(t *testing.T) {
	root := t.TempDir()
	mcp := filepath.Join(root, ".mcp.json")
	userMCP := []byte(`{"mcpServers": {"user-own": {"command": "user-cmd"}}}`)
	if err := os.WriteFile(mcp, userMCP, 0o644); err != nil {
		t.Fatal(err)
	}
	before := sha256File(t, mcp)

	var out, warn bytes.Buffer
	if err := addCodexWiring(root, &out, &warn); err != nil {
		t.Fatalf("addCodexWiring: %v", err)
	}
	if after := sha256File(t, mcp); after != before {
		t.Errorf(".mcp.json bytes changed across --add-codex (before %s, after %s)", before, after)
	}
}

// TestUpdateAddCodex_RerunIsIdempotent covers AC-UAC-005: a re-run on an
// already-current wiring writes nothing new and prints no re-trust guidance.
func TestUpdateAddCodex_RerunIsIdempotent(t *testing.T) {
	root := t.TempDir()
	var out1, warn1 bytes.Buffer
	if err := addCodexWiring(root, &out1, &warn1); err != nil {
		t.Fatalf("addCodexWiring(first): %v", err)
	}
	sidecar := filepath.Join(root, codexwiring.SidecarPath)
	before := sha256File(t, sidecar)
	hooksBefore, err := os.ReadFile(filepath.Join(root, codexwiring.HooksRelPath))
	if err != nil {
		t.Fatal(err)
	}

	var out2, warn2 bytes.Buffer
	if err := addCodexWiring(root, &out2, &warn2); err != nil {
		t.Fatalf("addCodexWiring(second): %v", err)
	}
	if after := sha256File(t, sidecar); after != before {
		t.Errorf("sidecar sha changed across an unchanged regeneration (before %s, after %s)", before, after)
	}
	hooksAfter, _ := os.ReadFile(filepath.Join(root, codexwiring.HooksRelPath))
	if string(hooksBefore) != string(hooksAfter) {
		t.Errorf("hooks.json rewritten by an unchanged regeneration")
	}
	if strings.Contains(out2.String(), codexwiring.ReTrustGuidance) || strings.Contains(out2.String(), codexwiring.FirstTrustGuidance) {
		t.Errorf("unchanged regeneration must print no guidance:\n%s", out2.String())
	}
}

// TestUpdateAddCodex_PreservesUserConfigToml covers AC-UAC-006: a user's own
// config.toml tables/keys survive the wiring (create-if-absent merge).
func TestUpdateAddCodex_PreservesUserConfigToml(t *testing.T) {
	root := t.TempDir()
	codexDir := filepath.Join(root, ".codex")
	if err := os.MkdirAll(codexDir, 0o755); err != nil {
		t.Fatal(err)
	}
	userCfg := []byte("[user_own]\nkey = \"user-value\"\n")
	if err := os.WriteFile(filepath.Join(codexDir, "config.toml"), userCfg, 0o644); err != nil {
		t.Fatal(err)
	}

	var out, warn bytes.Buffer
	if err := addCodexWiring(root, &out, &warn); err != nil {
		t.Fatalf("addCodexWiring: %v", err)
	}
	cfg, err := os.ReadFile(filepath.Join(codexDir, "config.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(cfg), "user-value") {
		t.Errorf("user's own config.toml entry lost:\n%s", cfg)
	}
	if !strings.Contains(string(cfg), "mcp_servers.moai") {
		t.Errorf("wiring did not add [mcp_servers.moai]:\n%s", cfg)
	}
}

// TestUpdateAddCodex_ValidationRefusalFailsLoud pins REQ-UAC-005 / plan §D1
// at the CLI wrapper level (plan-audit F1): the --add-codex path must
// PROPAGATE the REQ-CW-003 refusal as an error (exit ≠ 0), never swallow it
// into a warning the way refreshCodexWiringBestEffortAt does. A wrapper
// mutant that warns-and-continues on ErrValidationRefused fails this test.
func TestUpdateAddCodex_ValidationRefusalFailsLoud(t *testing.T) {
	root := t.TempDir()
	codexDir := filepath.Join(root, ".codex")
	if err := os.MkdirAll(codexDir, 0o755); err != nil {
		t.Fatal(err)
	}
	bad := []byte("{\n  \"version\": 1,\n  \"hooks\": {}\n}\n")
	if err := os.WriteFile(filepath.Join(codexDir, "hooks.json"), bad, 0o644); err != nil {
		t.Fatal(err)
	}

	var out, warn bytes.Buffer
	err := addCodexWiring(root, &out, &warn)
	if err == nil {
		t.Fatal("addCodexWiring swallowed the REQ-CW-003 validation refusal — the --add-codex path must fail loud (exit ≠ 0, plan §D1)")
	}
	if !errors.Is(err, codexwiring.ErrValidationRefused) {
		t.Errorf("error is not the validation refusal: %v", err)
	}
	if !strings.Contains(err.Error(), "version") {
		t.Errorf("diagnostic does not name the violating key: %v", err)
	}
	// The refusal's hard half: no wiring bytes reach disk.
	after, rerr := os.ReadFile(filepath.Join(codexDir, "hooks.json"))
	if rerr != nil {
		t.Fatal(rerr)
	}
	if string(after) != string(bad) {
		t.Errorf("refused write modified hooks.json:\nbefore: %q\nafter:  %q", bad, after)
	}
	if _, statErr := os.Stat(filepath.Join(codexDir, "config.toml")); statErr == nil {
		t.Errorf("config.toml written despite the refusal — the wiring must stop at the refused file")
	}
}

// TestUpdateAddCodex_RunUpdatePropagatesRefusal is the reachability guard for
// the §D1 propagation contract: runUpdate's call must return the seam error
// (not discard it), so a refusal exits non-zero end to end.
func TestUpdateAddCodex_RunUpdatePropagatesRefusal(t *testing.T) {
	src, err := os.ReadFile("update.go")
	if err != nil {
		t.Fatalf("read update.go: %v", err)
	}
	body := string(src)
	idx := strings.Index(body, "addCodexWiringAt(")
	if idx < 0 {
		t.Fatal("runUpdate does not call addCodexWiringAt — the --add-codex verb is not wired")
	}
	window := body[max(0, idx-160):min(len(body), idx+160)]
	if !strings.Contains(window, "return err") {
		t.Errorf("runUpdate discards the addCodexWiringAt error — a validation refusal would not fail the command (plan §D1):\n%s", window)
	}
}

// TestUpdateAddCodex_CallSitsInsideFlagGuard covers AC-UAC-003 (the M-3
// mutant): the ungated call must run ONLY under `getBoolFlag(cmd,
// "add-codex")` — an unconditional Wire call would create wiring in every
// flag-absent update (REQ-UAC-002 regression).
func TestUpdateAddCodex_CallSitsInsideFlagGuard(t *testing.T) {
	src, err := os.ReadFile("update.go")
	if err != nil {
		t.Fatalf("read update.go: %v", err)
	}
	body := string(src)
	idx := strings.Index(body, "addCodexWiringAt(")
	if idx < 0 {
		t.Fatal("addCodexWiringAt not called in runUpdate")
	}
	window := body[max(0, idx-300):idx]
	if !strings.Contains(window, `getBoolFlag(cmd, "add-codex")`) {
		t.Errorf("addCodexWiringAt is not guarded by the --add-codex flag — a flag-absent update would create wiring (REQ-UAC-002):\n%s", window)
	}
}

// TestUpdateAddCodex_CallSitsBeforeSyncSkippedReturn pins the placement (plan
// §D1: beside the gated refresh at the same conditional seat): an "Up to
// date · Skipping sync" update still owes the --add-codex verb its wiring
// pass.
func TestUpdateAddCodex_CallSitsBeforeSyncSkippedReturn(t *testing.T) {
	src, err := os.ReadFile("update.go")
	if err != nil {
		t.Fatalf("read update.go: %v", err)
	}
	body := string(src)
	callIdx := strings.Index(body, "addCodexWiringAt(")
	if callIdx < 0 {
		t.Fatal("addCodexWiringAt not called in runUpdate")
	}
	skipIdx := strings.Index(body, "if syncSkipped {")
	if skipIdx < 0 {
		t.Fatal("syncSkipped early-return block not found in update.go — test premise stale")
	}
	if callIdx > skipIdx {
		t.Error("addCodexWiringAt sits AFTER the syncSkipped early return — an 'Up to date' update would skip the --add-codex wiring entirely")
	}
}

// TestUpdateAddCodex_DryRunPreviewListsActions covers AC-UAC-007's output
// half: --add-codex --dry-run previews the planned wiring actions (the
// no-write half is structural — the preview function takes only a writer).
func TestUpdateAddCodex_DryRunPreviewListsActions(t *testing.T) {
	var out bytes.Buffer
	emitAddCodexDryRunPreview(&out)
	got := out.String()
	for _, want := range []string{
		codexwiring.HooksRelPath,
		codexwiring.ConfigRelPath,
		filepath.ToSlash(codexwiring.SidecarPath),
	} {
		if !strings.Contains(got, want) {
			t.Errorf("dry-run preview missing planned action %s:\n%s", want, got)
		}
	}
	if !strings.Contains(got, "dry-run") && !strings.Contains(got, "Dry-run") {
		t.Errorf("preview does not state it is a dry run (nothing written):\n%s", got)
	}
}

// TestUpdateAddCodex_CheckCombinationRejected covers AC-UAC-008 / REQ-UAC-007:
// --check (informational) × --add-codex (mutation) is rejected fail-loud by
// the extended mutual-exclusion validator, with BOTH flags named.
func TestUpdateAddCodex_CheckCombinationRejected(t *testing.T) {
	err := validateUpdateVersionConflicts("", true, false, false, false, true)
	if err == nil {
		t.Fatal("--check + --add-codex must be rejected (informational × mutation)")
	}
	if !strings.Contains(err.Error(), "--check") || !strings.Contains(err.Error(), "--add-codex") {
		t.Errorf("rejection must name both flags: %v", err)
	}
	// Alone, the flag is legal; with --dry-run it is legal (REQ-UAC-006).
	if err := validateUpdateVersionConflicts("", false, false, false, false, true); err != nil {
		t.Errorf("--add-codex alone must pass validation: %v", err)
	}
	if err := validateUpdateVersionConflicts("", false, false, false, true, true); err != nil {
		t.Errorf("--add-codex + --dry-run must pass validation (REQ-UAC-006): %v", err)
	}
}

// TestUpdateAddCodex_FlagRegisteredWithAdditiveHelp covers AC-UAC-001 /
// REQ-UAC-008: the flag exists on updateCmd and its help describes it as the
// sanctioned additive path (the CLI --help smoke re-observes the rendered
// output; this pins the registration itself).
func TestUpdateAddCodex_FlagRegisteredWithAdditiveHelp(t *testing.T) {
	f := updateCmd.Flags().Lookup("add-codex")
	if f == nil {
		t.Fatal("--add-codex flag is not registered on updateCmd")
	}
	if f.Value.String() != "false" {
		t.Errorf("--add-codex default = %q, want false (flag-absent flow preserved, REQ-UAC-002)", f.Value.String())
	}
	if !strings.Contains(f.Usage, "additive") {
		t.Errorf("--add-codex usage %q does not describe the additive path (REQ-UAC-008)", f.Usage)
	}
}
