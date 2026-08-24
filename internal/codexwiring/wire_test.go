package codexwiring

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/codexadapter"
)

// wireFresh runs Wire on a fresh temp project root and fails the test on error.
func wireFresh(t *testing.T) (string, *bytes.Buffer, *bytes.Buffer, Result) {
	t.Helper()
	root := t.TempDir()
	var out, warn bytes.Buffer
	res, err := Wire(root, &out, &warn)
	if err != nil {
		t.Fatalf("Wire(fresh): %v (warnings: %s)", err, warn.String())
	}
	return root, &out, &warn, res
}

// TestWireCreatesAllArtifacts verifies the full artifact set lands on disk:
// .codex/hooks.json (validated), .codex/config.toml (canonical table +
// status_line), and the trust sidecar whose hooks hash matches the file
// (REQ-CW-002/003/004/008/013).
func TestWireCreatesAllArtifacts(t *testing.T) {
	root, _, _, res := wireFresh(t)

	hooksPath := filepath.Join(root, ".codex", "hooks.json")
	raw, err := os.ReadFile(hooksPath)
	if err != nil {
		t.Fatalf("hooks.json not created: %v", err)
	}
	if violations, verr := codexadapter.ValidateConfig(raw); verr != nil || len(violations) > 0 {
		t.Errorf("written hooks.json fails ValidateConfig (violations=%v err=%v)", violations, verr)
	}

	cfgRaw, err := os.ReadFile(filepath.Join(root, ".codex", "config.toml"))
	if err != nil {
		t.Fatalf("config.toml not created: %v", err)
	}
	cfg := string(cfgRaw)
	for _, want := range []string{"[mcp_servers.moai]", "default_tools_approval_mode = \"writes\"", "[tui]"} {
		if !strings.Contains(cfg, want) {
			t.Errorf("config.toml missing %q:\n%s", want, cfgRaw)
		}
	}

	sidecarRaw, err := os.ReadFile(filepath.Join(root, SidecarPath))
	if err != nil {
		t.Fatalf("trust sidecar not created: %v", err)
	}
	var sidecar struct {
		HooksSHA256 string `json:"hooks_sha256"`
	}
	if err := json.Unmarshal(sidecarRaw, &sidecar); err != nil {
		t.Fatalf("parse sidecar: %v\n%s", err, sidecarRaw)
	}
	sum := sha256.Sum256(raw)
	if sidecar.HooksSHA256 != hex.EncodeToString(sum[:]) {
		t.Errorf("sidecar hooks_sha256 = %q, want %q (sha256 of written hooks.json)", sidecar.HooksSHA256, hex.EncodeToString(sum[:]))
	}
	if !res.HooksChanged {
		t.Errorf("fresh wiring must report HooksChanged (guidance depends on it)")
	}
}

// TestWireFirstTrustGuidance verifies a first creation prints the Codex trust
// flow guidance carrying the 'codex /hooks' token (AC-CW-008 clause 1).
func TestWireFirstTrustGuidance(t *testing.T) {
	_, out, _, _ := wireFresh(t)
	if !strings.Contains(out.String(), "codex /hooks") {
		t.Errorf("first-trust guidance missing the 'codex /hooks' token:\n%s", out.String())
	}
}

// TestWireReTrustGuidanceOnContentChange verifies a regeneration that CHANGES
// hooks.json content prints the re-trust guidance with '/hooks to re-trust'
// (AC-CW-008 clause 2 — the user removed a MoAI handler; regeneration
// restores it, content diverges from the on-disk file).
func TestWireReTrustGuidanceOnContentChange(t *testing.T) {
	root, _, _, _ := wireFresh(t)
	hooksPath := filepath.Join(root, ".codex", "hooks.json")
	raw, err := os.ReadFile(hooksPath)
	if err != nil {
		t.Fatalf("read hooks.json: %v", err)
	}

	// Simulate the user deleting one MoAI handler (the AC-CW-008 scenario).
	tampered := strings.Replace(string(raw), "moai hook stop --harness codex", "user-removed", 1)
	if tampered == string(raw) {
		t.Fatal("tamper substitution found nothing — test premise broken")
	}
	if err := os.WriteFile(hooksPath, []byte(tampered), 0o644); err != nil {
		t.Fatalf("write tampered hooks.json: %v", err)
	}

	var out, warn bytes.Buffer
	if _, err := Wire(root, &out, &warn); err != nil {
		t.Fatalf("Wire(regenerate): %v (warnings: %s)", err, warn.String())
	}
	if !strings.Contains(out.String(), "/hooks to re-trust") {
		t.Errorf("re-trust guidance missing '/hooks to re-trust' on content change:\n%s", out.String())
	}

	restored, err := os.ReadFile(hooksPath)
	if err != nil {
		t.Fatalf("re-read hooks.json: %v", err)
	}
	if !strings.Contains(string(restored), "moai hook stop --harness codex") {
		t.Errorf("regeneration did not restore the removed MoAI handler")
	}
}

// TestWireNoReTrustGuidanceOnNoChange verifies an unchanged regeneration
// prints NO re-trust guidance (AC-CW-008 clause 3).
func TestWireNoReTrustGuidanceOnNoChange(t *testing.T) {
	root, _, _, _ := wireFresh(t)
	var out, warn bytes.Buffer
	if _, err := Wire(root, &out, &warn); err != nil {
		t.Fatalf("Wire(second): %v", err)
	}
	if strings.Contains(out.String(), "/hooks to re-trust") {
		t.Errorf("unchanged regeneration must not print re-trust guidance:\n%s", out.String())
	}
}

// TestWireIdempotentFilesSha verifies the wired files are byte-stable across
// regenerations (AC-CW-006, sha256 comparison).
func TestWireIdempotentFilesSha(t *testing.T) {
	root, _, _, _ := wireFresh(t)
	before, err := os.ReadFile(filepath.Join(root, ".codex", "hooks.json"))
	if err != nil {
		t.Fatal(err)
	}
	beforeCfg, err := os.ReadFile(filepath.Join(root, ".codex", "config.toml"))
	if err != nil {
		t.Fatal(err)
	}
	var out, warn bytes.Buffer
	if _, err := Wire(root, &out, &warn); err != nil {
		t.Fatalf("Wire(second): %v", err)
	}
	after, _ := os.ReadFile(filepath.Join(root, ".codex", "hooks.json"))
	afterCfg, _ := os.ReadFile(filepath.Join(root, ".codex", "config.toml"))
	if string(before) != string(after) {
		t.Errorf("hooks.json changed across an unchanged regeneration")
	}
	if string(beforeCfg) != string(afterCfg) {
		t.Errorf("config.toml changed across an unchanged regeneration")
	}
}

// TestWireValidationRefusalWritesNothing verifies the REQ-CW-003 hard gate: an
// existing hooks.json carrying whitelist-violating keys (a "version" key)
// makes Wire refuse to write, abort with a diagnostic, and leave the file
// bytes untouched. config.toml is not written either (the wiring aborts).
func TestWireValidationRefusalWritesNothing(t *testing.T) {
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
	res, err := Wire(root, &out, &warn)
	if err == nil {
		t.Fatalf("Wire must fail loudly on a whitelist violation (res=%+v)", res)
	}
	if !strings.Contains(err.Error(), "version") && !strings.Contains(warn.String(), "version") {
		t.Errorf("diagnostic does not name the violating key: err=%v warn=%s", err, warn.String())
	}

	after, rerr := os.ReadFile(filepath.Join(codexDir, "hooks.json"))
	if rerr != nil {
		t.Fatal(rerr)
	}
	if string(after) != string(bad) {
		t.Errorf("refused write still modified hooks.json:\nbefore: %q\nafter:  %q", bad, after)
	}
	if _, statErr := os.Stat(filepath.Join(codexDir, "config.toml")); statErr == nil {
		t.Errorf("config.toml written despite the REQ-CW-003 abort — the wiring must stop at the refused file")
	}
}

// TestWireUnparseableExistingWarnsAndContinues verifies the §B edge case: a
// corrupted (unparseable) hooks.json is left untouched with a diagnostic
// warning, and the rest of the wiring (config.toml) proceeds — init never
// fails on wiring (best-effort, REQ-CW-003 refusal excepted).
func TestWireUnparseableExistingWarnsAndContinues(t *testing.T) {
	root := t.TempDir()
	codexDir := filepath.Join(root, ".codex")
	if err := os.MkdirAll(codexDir, 0o755); err != nil {
		t.Fatal(err)
	}
	garbage := []byte("{not json at all")
	if err := os.WriteFile(filepath.Join(codexDir, "hooks.json"), garbage, 0o644); err != nil {
		t.Fatal(err)
	}

	var out, warn bytes.Buffer
	if _, err := Wire(root, &out, &warn); err != nil {
		t.Fatalf("unparseable user hooks.json must not fail init (warn-and-continue): %v", err)
	}
	if !strings.Contains(warn.String(), "hooks.json") {
		t.Errorf("diagnostic warning does not name hooks.json: %q", warn.String())
	}
	after, rerr := os.ReadFile(filepath.Join(codexDir, "hooks.json"))
	if rerr != nil {
		t.Fatal(rerr)
	}
	if string(after) != string(garbage) {
		t.Errorf("corrupted hooks.json modified:\nbefore: %q\nafter:  %q", garbage, after)
	}
	if _, statErr := os.Stat(filepath.Join(codexDir, "config.toml")); statErr != nil {
		t.Errorf("config.toml not written after a warn-and-continue hooks skip: %v", statErr)
	}
}

// TestRefreshWiringAbsentFilesNoop verifies the REQ-CW-009 opt-in rule: a
// project without wiring files gets NOTHING created by the update path.
func TestRefreshWiringAbsentFilesNoop(t *testing.T) {
	root := t.TempDir()
	var out, warn bytes.Buffer
	res, err := RefreshWiring(root, &out, &warn)
	if err != nil {
		t.Fatalf("RefreshWiring(no files): %v", err)
	}
	if res.HooksWritten || res.ConfigWritten {
		t.Errorf("RefreshWiring created wiring in an opt-out project (res=%+v)", res)
	}
	if out.Len() != 0 {
		t.Errorf("RefreshWiring printed guidance in an opt-out project: %q", out.String())
	}
	if _, statErr := os.Stat(filepath.Join(root, ".codex")); statErr == nil {
		t.Errorf(".codex directory created by the update path in an opt-out project")
	}
}

// TestRefreshWiringExistingFilesRefreshed verifies a wired project's files are
// refreshed by the update path (REQ-CW-009 second clause).
func TestRefreshWiringExistingFilesRefreshed(t *testing.T) {
	root, _, _, _ := wireFresh(t)
	hooksPath := filepath.Join(root, ".codex", "hooks.json")

	// Tamper (user removed a handler) so a refresh has observable work.
	raw, _ := os.ReadFile(hooksPath)
	tampered := strings.Replace(string(raw), "moai hook stop --harness codex", "user-removed", 1)
	if err := os.WriteFile(hooksPath, []byte(tampered), 0o644); err != nil {
		t.Fatal(err)
	}

	var out, warn bytes.Buffer
	res, err := RefreshWiring(root, &out, &warn)
	if err != nil {
		t.Fatalf("RefreshWiring(existing): %v", err)
	}
	if !res.HooksChanged {
		t.Errorf("RefreshWiring did not report the content change")
	}
	if !strings.Contains(out.String(), "/hooks to re-trust") {
		t.Errorf("content-changing refresh missing the re-trust guidance (REQ-CW-009 → REQ-CW-008):\n%s", out.String())
	}
	after, _ := os.ReadFile(hooksPath)
	if !strings.Contains(string(after), "moai hook stop --harness codex") {
		t.Errorf("RefreshWiring did not restore the removed handler")
	}
}

// TestWirePreservesUserConfigTableAndEntries verifies the merge model across
// the full Wire path: user config tables and hook entries survive (AC-CW-007).
func TestWirePreservesUserConfigTableAndEntries(t *testing.T) {
	root := t.TempDir()
	codexDir := filepath.Join(root, ".codex")
	if err := os.MkdirAll(codexDir, 0o755); err != nil {
		t.Fatal(err)
	}
	userHooks := []byte(`{
  "hooks": {
    "PreToolUse": [
      {"matcher": "Bash", "hooks": [{"type": "command", "command": "my-own-hook", "timeout": 30}]}
    ]
  }
}`)
	if err := os.WriteFile(filepath.Join(codexDir, "hooks.json"), userHooks, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(codexDir, "config.toml"),
		[]byte("[mcp_servers.other]\ncommand = \"npx\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var out, warn bytes.Buffer
	if _, err := Wire(root, &out, &warn); err != nil {
		t.Fatalf("Wire(user content): %v (warn: %s)", err, warn.String())
	}

	merged, _ := os.ReadFile(filepath.Join(codexDir, "hooks.json"))
	if !strings.Contains(string(merged), "my-own-hook") {
		t.Errorf("user hook entry lost across Wire:\n%s", merged)
	}
	cfg, _ := os.ReadFile(filepath.Join(codexDir, "config.toml"))
	if !strings.Contains(string(cfg), "[mcp_servers.other]") {
		t.Errorf("user config table lost across Wire:\n%s", cfg)
	}
	if !strings.Contains(string(cfg), "[mcp_servers.moai]") {
		t.Errorf("canonical moai table missing alongside user table:\n%s", cfg)
	}
}
