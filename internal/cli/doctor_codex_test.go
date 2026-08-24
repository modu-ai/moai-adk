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

// TestCheckCodexWiring_InactiveProjectInformationalSkip verifies a project
// without wiring files reports an informational skip (CheckOK), never a
// failure (AC-CW-012 third clause).
func TestCheckCodexWiring_InactiveProjectInformationalSkip(t *testing.T) {
	stubMoaiLookup(t, false) // even a moai-less machine must not turn skip into failure
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
