package cli

// SPEC-CODEX-WIRING-001 M4 — the `moai doctor` "Codex Wiring" diagnostic
// (REQ-CW-010 / AC-CW-012). Advisory and fail-open (checkBinaryFreshness t184
// precedent): the check never blocks the rest of doctor, an inactive project
// is an informational skip, and drift is REPORTED (the wiring writer never
// repairs a user-owned surface).

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/codexwiring"
)

// wireProjectForDoctor wires a fresh temp project and returns its root.
func wireProjectForDoctor(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	var out, warn bytes.Buffer
	if _, err := codexwiring.Wire(root, &out, &warn); err != nil {
		t.Fatalf("Wire: %v", err)
	}
	return root
}

// stubMoaiLookup pins the PATH-resolution sub-check deterministically: tests
// install a fake that reports moai found (doctor only needs the verdict).
func stubMoaiLookup(t *testing.T, found bool) {
	t.Helper()
	orig := codexWiringLookPath
	codexWiringLookPath = func(string) (string, error) {
		if found {
			return "/usr/local/bin/moai", nil
		}
		return "", os.ErrNotExist
	}
	t.Cleanup(func() { codexWiringLookPath = orig })
}

// stubCodexLookup pins the PATH-resolution seam per binary name: the "moai"
// and "codex" sub-checks ask different questions, so a test that needs them to
// answer differently cannot use the verdict-only stubMoaiLookup.
func stubCodexLookup(t *testing.T, moaiFound, codexFound bool) {
	t.Helper()
	orig := codexWiringLookPath
	codexWiringLookPath = func(name string) (string, error) {
		found := moaiFound
		if name == "codex" {
			found = codexFound
		}
		if found {
			return "/usr/local/bin/" + name, nil
		}
		return "", os.ErrNotExist
	}
	t.Cleanup(func() { codexWiringLookPath = orig })
}

// stubCodexHome pins the home-directory seam so the stale-skill sub-check
// never reads the developer's real ~/.codex/config.toml (t.Setenv("HOME", …)
// is prohibited here — it pollutes parallel tests).
func stubCodexHome(t *testing.T, home string) {
	t.Helper()
	orig := codexWiringUserHomeDir
	codexWiringUserHomeDir = func() (string, error) { return home, nil }
	t.Cleanup(func() { codexWiringUserHomeDir = orig })
}

// writeCodexHomeConfig writes a ~/.codex/config.toml carrying one
// [[skills.config]] entry per supplied path, and returns the home root.
func writeCodexHomeConfig(t *testing.T, entries []struct {
	Path    string
	Enabled bool
}) string {
	t.Helper()
	home := t.TempDir()
	dir := filepath.Join(home, ".codex")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	var sb strings.Builder
	sb.WriteString("model = \"gpt-5\"\n\n")
	for _, e := range entries {
		sb.WriteString("[[skills.config]]\npath = \"" + e.Path + "\"\n")
		if e.Enabled {
			sb.WriteString("enabled = true\n")
		} else {
			sb.WriteString("enabled = false\n")
		}
		sb.WriteString("\n")
	}
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(sb.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	return home
}

// TestCheckCodexWiring_UnwiredWithCodexInstalledWarns verifies the branch the
// silent skip used to swallow: codex resolves on PATH but the project carries
// no wiring, so the MCP server is unregistered and the hooks cannot fire here.
// The action directive must ride in Message — Detail renders only under
// --verbose, and a plain `moai doctor` has to show it.
func TestCheckCodexWiring_UnwiredWithCodexInstalledWarns(t *testing.T) {
	stubCodexLookup(t, true, true)
	stubCodexHome(t, t.TempDir())
	check := checkCodexWiring(t.TempDir(), false)
	if check.Status != uikit.CheckWarn {
		t.Errorf("unwired project with codex installed status = %v, want Warn: %+v", check.Status, check)
	}
	if !strings.Contains(check.Message, "moai init --agent codex") {
		t.Errorf("action directive missing from Message (Detail is --verbose-only): %+v", check)
	}
	for _, want := range []string{codexwiring.HooksRelPath, codexwiring.ConfigRelPath} {
		if !strings.Contains(check.Message, want) {
			t.Errorf("message does not name the absent path %q: %+v", want, check)
		}
	}
}

// TestCheckCodexWiring_StaleHomeSkillsReported verifies the second silence:
// ~/.codex/config.toml [[skills.config]] entries whose path no longer exists
// are reported, quantified, and split by enabled state (an enabled missing
// path is live breakage; a disabled one is stale garbage).
func TestCheckCodexWiring_StaleHomeSkillsReported(t *testing.T) {
	stubCodexLookup(t, true, true)
	live := filepath.Join(t.TempDir(), "SKILL.md")
	if err := os.WriteFile(live, []byte("# skill\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	home := writeCodexHomeConfig(t, []struct {
		Path    string
		Enabled bool
	}{
		{Path: live, Enabled: true},
		{Path: "/nonexistent/moai-a/SKILL.md", Enabled: true},
		{Path: "/nonexistent/moai-b/SKILL.md", Enabled: false},
		{Path: "/nonexistent/moai-c/SKILL.md", Enabled: false},
	})
	stubCodexHome(t, home)

	check := checkCodexWiring(wireProjectForDoctor(t), false)
	if check.Status != uikit.CheckWarn {
		t.Errorf("stale home skills status = %v, want Warn: %+v", check.Status, check)
	}
	combined := check.Message + " " + check.Detail
	if !strings.Contains(combined, "skills.config") {
		t.Errorf("finding does not name the [[skills.config]] surface: %+v", check)
	}
	if !strings.Contains(combined, "config.toml") {
		t.Errorf("finding does not point at ~/.codex/config.toml: %+v", check)
	}
	// 3 of 4 missing, split 1 enabled / 2 disabled — every number must appear.
	for _, want := range []string{"3", "4", "1 enabled", "2 disabled"} {
		if !strings.Contains(combined, want) {
			t.Errorf("finding does not quantify %q: %+v", want, check)
		}
	}
}

// TestCheckCodexWiring_HealthyHomeSkillsNoFinding verifies the no-false-
// positive path: every declared skill path exists, so the sub-check is silent
// and a healthy project stays OK.
func TestCheckCodexWiring_HealthyHomeSkillsNoFinding(t *testing.T) {
	stubCodexLookup(t, true, true)
	live := filepath.Join(t.TempDir(), "SKILL.md")
	if err := os.WriteFile(live, []byte("# skill\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	home := writeCodexHomeConfig(t, []struct {
		Path    string
		Enabled bool
	}{{Path: live, Enabled: true}})
	stubCodexHome(t, home)

	check := checkCodexWiring(wireProjectForDoctor(t), false)
	if check.Status != uikit.CheckOK {
		t.Errorf("all-present skill paths status = %v, want OK: %+v", check.Status, check)
	}
}

// TestCheckCodexWiring_AbsentHomeConfigSilent verifies an absent
// ~/.codex/config.toml degrades to a silent skip rather than a finding
// (fail-open: an unreadable input is never a failure).
func TestCheckCodexWiring_AbsentHomeConfigSilent(t *testing.T) {
	stubCodexLookup(t, true, true)
	stubCodexHome(t, t.TempDir()) // no .codex/config.toml inside
	check := checkCodexWiring(wireProjectForDoctor(t), false)
	if check.Status != uikit.CheckOK {
		t.Errorf("absent home config status = %v, want OK (silent skip): %+v", check.Status, check)
	}
	if strings.Contains(check.Message+check.Detail, "skills.config") {
		t.Errorf("absent home config produced a skills finding: %+v", check)
	}
}

// TestCheckCodexWiring_ClaudeOnlyMachineStaysSilent verifies the un-nagging
// invariant from the other side: no wiring AND no codex binary means the
// stale-skill sub-check never runs either, even with a stale home config.
func TestCheckCodexWiring_ClaudeOnlyMachineStaysSilent(t *testing.T) {
	stubCodexLookup(t, true, false)
	home := writeCodexHomeConfig(t, []struct {
		Path    string
		Enabled bool
	}{{Path: "/nonexistent/moai-a/SKILL.md", Enabled: true}})
	stubCodexHome(t, home)

	check := checkCodexWiring(t.TempDir(), false)
	if check.Status != uikit.CheckOK {
		t.Errorf("claude-only machine status = %v, want OK: %+v", check.Status, check)
	}
	if strings.Contains(check.Message+check.Detail, "skills.config") {
		t.Errorf("claude-only machine was nagged about home skills: %+v", check)
	}
}

// TestCheckCodexWiring_InactiveProjectInformationalSkip verifies a project
// without wiring files reports an informational skip (CheckOK), never a
// failure (AC-CW-012 third clause).
func TestCheckCodexWiring_InactiveProjectInformationalSkip(t *testing.T) {
	stubMoaiLookup(t, false) // even a moai-less machine must not turn skip into failure
	stubCodexHome(t, t.TempDir())
	check := checkCodexWiring(t.TempDir(), false)
	if check.Status != uikit.CheckOK {
		t.Errorf("inactive project status = %v, want OK (informational skip): %+v", check.Status, check)
	}
	if !strings.Contains(strings.ToLower(check.Message), "skip") && !strings.Contains(strings.ToLower(check.Message), "not wired") {
		t.Errorf("skip message should say so: %+v", check)
	}
}

// TestCheckCodexWiring_HealthyProjectOK verifies a freshly wired project
// with moai on PATH passes clean (AC-CW-012 first clause).
func TestCheckCodexWiring_HealthyProjectOK(t *testing.T) {
	stubMoaiLookup(t, true)
	stubCodexHome(t, t.TempDir())
	check := checkCodexWiring(wireProjectForDoctor(t), false)
	if check.Status != uikit.CheckOK {
		t.Errorf("healthy project status = %v, want OK: %+v", check.Status, check)
	}
}

// TestCheckCodexWiring_DivergenceAdvisesReTrust verifies the sidecar-hash
// divergence path: an unauthorized hooks.json edit is reported WITH the
// `/hooks to re-trust` action directive (AC-CW-012 second clause; the advice
// is an action instruction, never a claim about Codex's internal state).
func TestCheckCodexWiring_DivergenceAdvisesReTrust(t *testing.T) {
	stubMoaiLookup(t, true)
	stubCodexHome(t, t.TempDir())
	root := wireProjectForDoctor(t)
	hooksPath := filepath.Join(root, codexwiring.HooksRelPath)
	raw, err := os.ReadFile(hooksPath)
	if err != nil {
		t.Fatal(err)
	}
	tampered := string(raw) + "\n" // any byte change diverges the sidecar hash
	if err := os.WriteFile(hooksPath, []byte(tampered), 0o644); err != nil {
		t.Fatal(err)
	}

	check := checkCodexWiring(root, false)
	if check.Status != uikit.CheckWarn {
		t.Errorf("diverged project status = %v, want Warn: %+v", check.Status, check)
	}
	combined := check.Message + " " + check.Detail
	if !strings.Contains(combined, "/hooks to re-trust") {
		t.Errorf("divergence must carry the /hooks to re-trust directive: %+v", check)
	}
}

// TestCheckCodexWiring_ValidationFailureReported verifies a whitelist
// violation in the on-disk hooks.json is surfaced (Codex would silently
// disable the file — doctor is the observability backstop).
func TestCheckCodexWiring_ValidationFailureReported(t *testing.T) {
	stubMoaiLookup(t, true)
	stubCodexHome(t, t.TempDir())
	root := t.TempDir()
	codexDir := filepath.Join(root, ".codex")
	if err := os.MkdirAll(codexDir, 0o755); err != nil {
		t.Fatal(err)
	}
	bad := []byte("{\n  \"version\": 1,\n  \"hooks\": {}\n}\n")
	if err := os.WriteFile(filepath.Join(codexDir, "hooks.json"), bad, 0o644); err != nil {
		t.Fatal(err)
	}

	check := checkCodexWiring(root, false)
	if check.Status != uikit.CheckWarn {
		t.Errorf("violating hooks.json status = %v, want Warn: %+v", check.Status, check)
	}
	if !strings.Contains(check.Message+check.Detail, "version") {
		t.Errorf("diagnostic does not name the violating key: %+v", check)
	}
}

// TestCheckCodexWiring_MoaiNotOnPathReported verifies the PATH-resolution
// sub-check: wiring without a resolvable moai binary means the generated
// hook commands cannot fire.
func TestCheckCodexWiring_MoaiNotOnPathReported(t *testing.T) {
	stubMoaiLookup(t, false)
	stubCodexHome(t, t.TempDir())
	check := checkCodexWiring(wireProjectForDoctor(t), false)
	if check.Status != uikit.CheckWarn {
		t.Errorf("moai-not-on-PATH status = %v, want Warn: %+v", check.Status, check)
	}
	if !strings.Contains(strings.ToLower(check.Message+check.Detail), "path") {
		t.Errorf("diagnostic does not mention PATH: %+v", check)
	}
}

// TestCheckCodexWiring_ConfigTableDriftReported verifies a user-modified
// [mcp_servers.moai] table is REPORTED (byte-invariant writer, doctor
// reports — REQ-CW-005).
func TestCheckCodexWiring_ConfigTableDriftReported(t *testing.T) {
	stubMoaiLookup(t, true)
	stubCodexHome(t, t.TempDir())
	root := wireProjectForDoctor(t)
	cfgPath := filepath.Join(root, codexwiring.ConfigRelPath)
	raw, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	drifted := strings.Replace(string(raw), `command = "moai"`, `command = "my-custom-moai"`, 1)
	if drifted == string(raw) {
		t.Fatal("drift substitution found nothing — premise broken")
	}
	if err := os.WriteFile(cfgPath, []byte(drifted), 0o644); err != nil {
		t.Fatal(err)
	}

	check := checkCodexWiring(root, false)
	if check.Status != uikit.CheckWarn {
		t.Errorf("drifted config table status = %v, want Warn: %+v", check.Status, check)
	}
	if !strings.Contains(strings.ToLower(check.Message+check.Detail), "mcp_servers.moai") {
		t.Errorf("diagnostic does not name the drifted table: %+v", check)
	}
}

// TestDoctor_CodexWiringRegistered verifies the check is registered in the
// Workspace group of runGroupedChecksObserved (the --check filter reaches
// it — the registration itself is the AC-CW-012 surface `moai doctor` grep
// relies on).
func TestDoctor_CodexWiringRegistered(t *testing.T) {
	stubMoaiLookup(t, true)
	stubCodexHome(t, t.TempDir())
	groups := runGroupedChecks(false, "Codex Wiring")
	var found bool
	for _, g := range groups {
		for _, c := range g.checks {
			if c.Name == "Codex Wiring" {
				found = true
			}
		}
	}
	if !found {
		t.Error("\"Codex Wiring\" check not reachable via runGroupedChecks — not registered in the Workspace group")
	}
}
