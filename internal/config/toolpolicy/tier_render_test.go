package toolpolicy

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// SPEC-AUTONOMY-TIERS-001 M3 — tier → permission-bundle renderer (USER vs PROJECT scope).
// AC-003: defaultMode → USER scope; deny/ask → PROJECT scope.
// AC-004: deny/ask byte-identical across the 3 rendered tiers.
// AC-007: semi-auto → byte-identical to today's template (defaultMode=default).

// helperSettingsJSON builds a minimal settings.json body with a permissions
// block so the codegen region matcher has something to splice into.
func helperSettingsJSON(t *testing.T, defaultMode string) []byte {
	t.Helper()
	body := `{
  "permissions": {
    "defaultMode": "` + defaultMode + `"
  }
}`
	return []byte(body)
}

func writeSettings(t *testing.T, path, defaultMode string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, helperSettingsJSON(t, defaultMode), 0o644); err != nil {
		t.Fatal(err)
	}
}

// policyDocWith returns a minimal PolicyDocument with one deny + one ask entry
// (enough to exercise the deny/ask emission). The tier-invariance check
// (AC-004) does NOT depend on the entry count — the renderer loads doc once.
func policyDocWith(deny, ask string) *PolicyDocument {
	return &PolicyDocument{
		Entries: []PolicyEntry{
			{Tool: "Bash", ArgsPattern: deny, Decision: DecisionDeny, RiskTier: RiskTierIrreversible},
			{Tool: "Bash", ArgsPattern: ask, Decision: DecisionAsk, RiskTier: RiskTierWrite},
		},
	}
}

func TestRenderTierPermissions_ScopeSplit(t *testing.T) {
	// AC-003: defaultMode lands in USER scope; deny/ask land in PROJECT scope.
	dir := t.TempDir()
	userPath := filepath.Join(dir, "user", ".claude", "settings.json")
	projectPath := filepath.Join(dir, "project", ".claude", "settings.json")
	writeSettings(t, userPath, "default")
	writeSettings(t, projectPath, "default")

	doc := policyDocWith("git push --force:*", "git push:*")
	res, err := RenderTierPermissions(projectPath, userPath, "auto", doc)
	if err != nil {
		t.Fatalf("RenderTierPermissions: %v", err)
	}
	if res.UserPath != userPath || res.ProjectPath != projectPath {
		t.Errorf("paths mismatched: user=%q project=%q", res.UserPath, res.ProjectPath)
	}

	// USER scope MUST carry defaultMode=auto but NOT deny/ask.
	userBody, _ := os.ReadFile(userPath)
	var userPerms struct {
		Permissions struct {
			DefaultMode string   `json:"defaultMode"`
			Deny        []string `json:"deny"`
			Ask         []string `json:"ask"`
		} `json:"permissions"`
	}
	if err := json.Unmarshal(userBody, &userPerms); err != nil {
		t.Fatalf("user settings unmarshal: %v\nbody: %s", err, userBody)
	}
	if userPerms.Permissions.DefaultMode != "auto" {
		t.Errorf("USER defaultMode = %q, want %q", userPerms.Permissions.DefaultMode, "auto")
	}
	if len(userPerms.Permissions.Deny) != 0 || len(userPerms.Permissions.Ask) != 0 {
		t.Errorf("USER scope MUST NOT carry deny/ask (they are PROJECT-scoped); got deny=%v ask=%v",
			userPerms.Permissions.Deny, userPerms.Permissions.Ask)
	}

	// PROJECT scope MUST carry deny/ask but NOT defaultMode.
	projBody, _ := os.ReadFile(projectPath)
	var projPerms struct {
		Permissions struct {
			DefaultMode string   `json:"defaultMode"`
			Deny        []string `json:"deny"`
			Ask         []string `json:"ask"`
		} `json:"permissions"`
	}
	if err := json.Unmarshal(projBody, &projPerms); err != nil {
		t.Fatalf("project settings unmarshal: %v\nbody: %s", err, projBody)
	}
	if len(projPerms.Permissions.Deny) == 0 || len(projPerms.Permissions.Ask) == 0 {
		t.Errorf("PROJECT scope MUST carry deny/ask; got deny=%v ask=%v",
			projPerms.Permissions.Deny, projPerms.Permissions.Ask)
	}
	// AP-6: PROJECT scope MUST NOT carry auto/bypass defaultMode (v2.1.142+ makes
	// them USER-scope-only). defaultMode SHOULD be absent or "default" in PROJECT.
	if projPerms.Permissions.DefaultMode == "auto" || projPerms.Permissions.DefaultMode == "bypassPermissions" {
		t.Errorf("PROJECT defaultMode = %q — auto/bypass MUST NOT land in PROJECT scope (AP-6)", projPerms.Permissions.DefaultMode)
	}
}

func TestRenderTierPermissions_DenyAskInvariantAcrossTiers(t *testing.T) {
	// AC-004: deny/ask arrays byte-identical across the 3 rendered tiers.
	dir := t.TempDir()
	doc := policyDocWith("git push --force:*", "git push:*")
	modes := []string{"default", "auto", "bypassPermissions"}
	var denyBlobs, askBlobs [][]byte
	for _, mode := range modes {
		projectPath := filepath.Join(dir, mode, ".claude", "settings.json")
		writeSettings(t, projectPath, "default")
		if _, err := RenderTierPermissions(projectPath, projectPath, mode, doc); err != nil {
			t.Fatalf("RenderTierPermissions(%q): %v", mode, err)
		}
		body, _ := os.ReadFile(projectPath)
		var perms struct {
			Permissions struct {
				Deny []string `json:"deny"`
				Ask  []string `json:"ask"`
			} `json:"permissions"`
		}
		if err := json.Unmarshal(body, &perms); err != nil {
			t.Fatalf("unmarshal %q: %v", mode, err)
		}
		d, _ := json.Marshal(perms.Permissions.Deny)
		a, _ := json.Marshal(perms.Permissions.Ask)
		denyBlobs = append(denyBlobs, d)
		askBlobs = append(askBlobs, a)
	}
	for i := 1; i < len(modes); i++ {
		if string(denyBlobs[i]) != string(denyBlobs[0]) {
			t.Errorf("AC-004 violation: deny array differs between %q and %q:\n  %s\n  %s",
				modes[0], modes[i], denyBlobs[0], denyBlobs[i])
		}
		if string(askBlobs[i]) != string(askBlobs[0]) {
			t.Errorf("AC-004 violation: ask array differs between %q and %q:\n  %s\n  %s",
				modes[0], modes[i], askBlobs[0], askBlobs[i])
		}
	}
}

func TestRenderTierPermissions_SemiAutoDefaultMode(t *testing.T) {
	// AC-007: semi-auto → defaultMode="default" in USER scope (today's value).
	dir := t.TempDir()
	userPath := filepath.Join(dir, "u", ".claude", "settings.json")
	projectPath := filepath.Join(dir, "p", ".claude", "settings.json")
	writeSettings(t, userPath, "default")
	writeSettings(t, projectPath, "default")

	doc := policyDocWith("rm -rf /:*", "git push:*")
	if _, err := RenderTierPermissions(projectPath, userPath, "default", doc); err != nil {
		t.Fatalf("RenderTierPermissions: %v", err)
	}
	body, _ := os.ReadFile(userPath)
	var perms struct {
		Permissions struct {
			DefaultMode string `json:"defaultMode"`
		} `json:"permissions"`
	}
	if err := json.Unmarshal(body, &perms); err != nil {
		t.Fatalf("unmarshal user perms: %v", err)
	}
	if perms.Permissions.DefaultMode != "default" {
		t.Errorf("semi-auto USER defaultMode = %q, want %q", perms.Permissions.DefaultMode, "default")
	}
}

func TestRenderTierPermissions_Idempotent(t *testing.T) {
	// Two consecutive renders produce byte-identical output (mirrors AC-TPS-013).
	dir := t.TempDir()
	userPath := filepath.Join(dir, "u", ".claude", "settings.json")
	projectPath := filepath.Join(dir, "p", ".claude", "settings.json")
	writeSettings(t, userPath, "default")
	writeSettings(t, projectPath, "default")
	doc := policyDocWith("git push --force:*", "git push:*")

	if _, err := RenderTierPermissions(projectPath, userPath, "auto", doc); err != nil {
		t.Fatal(err)
	}
	firstUser, _ := os.ReadFile(userPath)
	firstProj, _ := os.ReadFile(projectPath)

	if _, err := RenderTierPermissions(projectPath, userPath, "auto", doc); err != nil {
		t.Fatal(err)
	}
	secondUser, _ := os.ReadFile(userPath)
	secondProj, _ := os.ReadFile(projectPath)

	if string(firstUser) != string(secondUser) {
		t.Errorf("USER scope not idempotent:\nfirst:  %s\nsecond: %s", firstUser, secondUser)
	}
	if string(firstProj) != string(secondProj) {
		t.Errorf("PROJECT scope not idempotent:\nfirst:  %s\nsecond: %s", firstProj, secondProj)
	}
}
