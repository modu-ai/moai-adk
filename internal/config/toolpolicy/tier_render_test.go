package toolpolicy

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
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

// --- card t1630 card-review repair (P1) — the last_generated sidecar writer.
// The read side (mergeProjectAllow) already classifies against
// .moai/state/tool-policy/managed-allow.json, but nothing wrote the record, so
// an allow entry later removed from the policy document persisted forever
// (revoked permission resurrection). These tests pin the writer: a successful
// PROJECT-scope render persists full.Allow as last_generated; a removed
// generated entry is then dropped at the next render; a failed render leaves
// the record untouched.

// helperAllowDoc returns a PolicyDocument whose entries generate exactly the
// given allow specifiers ("Bash(x:*)" form) plus one deny and one ask entry so
// both scopes carry content. Allow entries sort deterministically (the codegen
// sorts by specifier), so expected allow lists are compared sorted.
func helperAllowDoc(allow ...string) *PolicyDocument {
	entries := []PolicyEntry{
		{Tool: "Bash", ArgsPattern: "git push --force:*", Decision: DecisionDeny, RiskTier: RiskTierIrreversible},
		{Tool: "Bash", ArgsPattern: "git push:*", Decision: DecisionAsk, RiskTier: RiskTierWrite},
	}
	for _, spec := range allow {
		inner := strings.TrimSuffix(strings.TrimPrefix(spec, "Bash("), ")")
		entries = append(entries, PolicyEntry{Tool: "Bash", ArgsPattern: inner, Decision: DecisionAllow, RiskTier: RiskTierRead})
	}
	return &PolicyDocument{Entries: entries}
}

// readManagedAllowRecord reads and parses the sidecar record for a project
// settings path. Fails the test when the record file is absent.
func readManagedAllowRecord(t *testing.T, projectSettingsPath string) managedAllowRecord {
	t.Helper()
	root := filepath.Dir(filepath.Dir(projectSettingsPath))
	body, err := os.ReadFile(filepath.Join(root, ".moai", "state", "tool-policy", "managed-allow.json"))
	if err != nil {
		t.Fatalf("managed-allow record not written: %v", err)
	}
	var rec managedAllowRecord
	if err := json.Unmarshal(body, &rec); err != nil {
		t.Fatalf("managed-allow record is not valid JSON: %v\n%s", err, body)
	}
	return rec
}

func TestRenderTierPermissions_PersistsManagedAllowRecord(t *testing.T) {
	// (a) a successful render writes the record with the generated managed
	// allow list, and a later render with an updated document refreshes it.
	dir := t.TempDir()
	userPath := filepath.Join(dir, "user", ".claude", "settings.json")
	projectPath := filepath.Join(dir, "proj", ".claude", "settings.json")
	writeSettings(t, userPath, "default")
	writeSettings(t, projectPath, "default")

	if _, err := RenderTierPermissions(projectPath, userPath, "default", helperAllowDoc("Bash(managed-a:*)", "Bash(managed-b:*)")); err != nil {
		t.Fatalf("first render: %v", err)
	}
	rec := readManagedAllowRecord(t, projectPath)
	want := []string{"Bash(managed-a:*)", "Bash(managed-b:*)"}
	sort.Strings(want)
	if !reflect.DeepEqual(rec.LastGenerated, want) {
		t.Errorf("last_generated = %v, want %v", rec.LastGenerated, want)
	}

	// Refresh: the record must be overwritten, not appended to.
	if _, err := RenderTierPermissions(projectPath, userPath, "default", helperAllowDoc("Bash(managed-c:*)")); err != nil {
		t.Fatalf("second render: %v", err)
	}
	rec = readManagedAllowRecord(t, projectPath)
	if !reflect.DeepEqual(rec.LastGenerated, []string{"Bash(managed-c:*)"}) {
		t.Errorf("last_generation refresh failed: last_generated = %v, want [Bash(managed-c:*)]", rec.LastGenerated)
	}
}

func TestRenderTierPermissions_RemovedGeneratedEntryIsDropped(t *testing.T) {
	// (b) the deciding test: an entry the document generated and later stops
	// generating must be dropped from the PROJECT managed block at the next
	// render (it is managed, not user-added). Without the record writer the
	// removed entry survives forever — revoked permission resurrection.
	dir := t.TempDir()
	userPath := filepath.Join(dir, "user", ".claude", "settings.json")
	projectPath := filepath.Join(dir, "proj", ".claude", "settings.json")
	writeSettings(t, userPath, "default")
	writeSettings(t, projectPath, "default")

	if _, err := RenderTierPermissions(projectPath, userPath, "default", helperAllowDoc("Bash(managed-a:*)", "Bash(managed-b:*)")); err != nil {
		t.Fatalf("first render: %v", err)
	}

	// The document now stops generating Bash(managed-b:*).
	if _, err := RenderTierPermissions(projectPath, userPath, "default", helperAllowDoc("Bash(managed-a:*)")); err != nil {
		t.Fatalf("second render: %v", err)
	}
	body, err := os.ReadFile(projectPath)
	if err != nil {
		t.Fatal(err)
	}
	allow, _, _, _, _ := ac001ReadPerms(t, body)
	want := []string{"Bash(managed-a:*)"}
	if !reflect.DeepEqual(allow, want) {
		t.Errorf("removed generated entry not dropped: allow = %v, want %v\nbody=%s", allow, want, body)
	}
}

func TestRenderTierPermissions_FailedRenderLeavesRecordUntouched(t *testing.T) {
	// (d) when the PROJECT-scope render fails, the record must stay exactly as
	// it was (no partial refresh, no creation on a failed path).
	dir := t.TempDir()
	userPath := filepath.Join(dir, "user", ".claude", "settings.json")
	projectPath := filepath.Join(dir, "proj", ".claude", "settings.json")
	recordPath := filepath.Join(dir, "proj", ".moai", "state", "tool-policy", "managed-allow.json")
	writeSettings(t, userPath, "default")
	if err := os.MkdirAll(filepath.Dir(projectPath), 0o755); err != nil {
		t.Fatal(err)
	}
	// Valid JSON without a permissions key — the region matcher fails.
	if err := os.WriteFile(projectPath, []byte(`{"env":{"FOO":"1"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(recordPath), 0o755); err != nil {
		t.Fatal(err)
	}
	sentinel := `{"last_generated":["Bash(sentinel:*)"]}`
	if err := os.WriteFile(recordPath, []byte(sentinel), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := RenderTierPermissions(projectPath, userPath, "default", helperAllowDoc("Bash(managed-a:*)")); err == nil {
		t.Fatalf("render over a permissions-less file must fail")
	}
	after, err := os.ReadFile(recordPath)
	if err != nil {
		t.Fatalf("record must survive a failed render: %v", err)
	}
	if string(after) != sentinel {
		t.Errorf("record changed on failed render:\nbefore=%s\nafter=%s", sentinel, after)
	}

	// Failed render with no record yet must not create one.
	freshProject := filepath.Join(dir, "fresh", ".claude", "settings.json")
	if err := os.MkdirAll(filepath.Dir(freshProject), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(freshProject, []byte(`{"env":{"FOO":"1"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := RenderTierPermissions(freshProject, userPath, "default", helperAllowDoc("Bash(managed-a:*)")); err == nil {
		t.Fatalf("render over a permissions-less file must fail")
	}
	if _, err := os.Stat(filepath.Join(dir, "fresh", ".moai", "state", "tool-policy", "managed-allow.json")); !os.IsNotExist(err) {
		t.Errorf("failed render must not create the record; stat err=%v", err)
	}
}

// --- SPEC-USER-SETTINGS-PROTECT-001 (card t1630) — repository-authored
// assertions for AC-001, AC-002, and AC-003 part A. The overlay probes under
// .moai/reports/t1630/evidence/ carried the RED observation on the pinned
// tree; these are the same assertions authored in the repository
// (acceptance.md: "run-phase: the same assertion authored in the repository").

// ac001ReadPerms unmarshals a settings body into the fields AC-001 asserts on.
func ac001ReadPerms(t *testing.T, body []byte) (allow, ask, deny, additional []string, env map[string]string) {
	t.Helper()
	var wrap struct {
		Permissions struct {
			Allow                 []string `json:"allow"`
			Ask                   []string `json:"ask"`
			Deny                  []string `json:"deny"`
			AdditionalDirectories []string `json:"additionalDirectories"`
		} `json:"permissions"`
		Env map[string]string `json:"env"`
	}
	if err := json.Unmarshal(body, &wrap); err != nil {
		t.Fatalf("after is not valid JSON: %v\n%s", err, body)
	}
	return wrap.Permissions.Allow, wrap.Permissions.Ask, wrap.Permissions.Deny, wrap.Permissions.AdditionalDirectories, wrap.Env
}

// TestAC001PreservesListsAndUnmodelledKeys: the USER-scope writer keeps the
// user's allow/ask/deny entries, every unmodelled permissions key, and every
// sibling region (REQ-001). The EC-1..EC-4 edge cases of acceptance.md are
// bound to this criterion and asserted as subtests.
func TestAC001PreservesListsAndUnmodelledKeys(t *testing.T) {
	dir := t.TempDir()
	userPath := filepath.Join(dir, ".claude", "settings.json")
	if err := os.MkdirAll(filepath.Dir(userPath), 0o755); err != nil {
		t.Fatal(err)
	}
	before := `{"permissions": {
      "defaultMode": "plan",
      "allow": ["Bash(make:*)"],
      "ask": ["Bash(rm:*)"],
      "deny": ["Read(./secret)"],
      "additionalDirectories": ["/tmp/x"]
    },"env":{"FOO":"1"}}`
	if err := os.WriteFile(userPath, []byte(before), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := WriteUserDefaultMode(userPath, "acceptEdits"); err != nil {
		t.Fatalf("WriteUserDefaultMode: %v", err)
	}
	after, err := os.ReadFile(userPath)
	if err != nil {
		t.Fatal(err)
	}
	allow, ask, deny, additional, env := ac001ReadPerms(t, after)
	if !reflect.DeepEqual(allow, []string{"Bash(make:*)"}) {
		t.Errorf("permissions.allow was removed; after=%s", after)
	}
	if !reflect.DeepEqual(ask, []string{"Bash(rm:*)"}) {
		t.Errorf("permissions.ask was removed; after=%s", after)
	}
	if !reflect.DeepEqual(deny, []string{"Read(./secret)"}) {
		t.Errorf("permissions.deny was removed; after=%s", after)
	}
	if !reflect.DeepEqual(additional, []string{"/tmp/x"}) {
		t.Errorf("permissions.additionalDirectories (unmodelled key) was removed; after=%s", after)
	}
	if !reflect.DeepEqual(env, map[string]string{"FOO": "1"}) {
		t.Errorf("sibling env region was removed; after=%s", after)
	}

	t.Run("ec1-absent-file", func(t *testing.T) {
		// EC-1: a settings file that does not exist yet — the writer creates
		// the minimal permissions skeleton with the tier default, no lists.
		absent := filepath.Join(t.TempDir(), ".claude", "settings.json")
		if err := WriteUserDefaultMode(absent, "acceptEdits"); err != nil {
			t.Fatalf("WriteUserDefaultMode on absent file: %v", err)
		}
		body, err := os.ReadFile(absent)
		if err != nil {
			t.Fatal(err)
		}
		allow, ask, deny, _, _ := ac001ReadPerms(t, body)
		if len(allow) != 0 || len(ask) != 0 || len(deny) != 0 {
			t.Errorf("created file MUST NOT invent lists; body=%s", body)
		}
	})
	t.Run("ec2-no-permissions-key", func(t *testing.T) {
		// EC-2: no "permissions" key — the writer reports the region error and
		// leaves the file unchanged.
		path := filepath.Join(t.TempDir(), "settings.json")
		body := `{"env":{"FOO":"1"}}`
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := WriteUserDefaultMode(path, "acceptEdits"); err == nil {
			t.Errorf("missing permissions key MUST return the region error")
		}
		afterBody, _ := os.ReadFile(path)
		if string(afterBody) != body {
			t.Errorf("file MUST stay unchanged on the error path:\nbefore=%s\nafter=%s", body, afterBody)
		}
	})
	t.Run("ec3-invalid-json", func(t *testing.T) {
		// EC-3: invalid JSON — the writer reports the region error and leaves
		// the file unchanged.
		path := filepath.Join(t.TempDir(), "settings.json")
		body := `{not json`
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := WriteUserDefaultMode(path, "acceptEdits"); err == nil {
			t.Errorf("invalid JSON MUST return the region error")
		}
		afterBody, _ := os.ReadFile(path)
		if string(afterBody) != body {
			t.Errorf("file MUST stay unchanged on the error path:\nbefore=%s\nafter=%s", body, afterBody)
		}
	})
	t.Run("ec4-duplicate-allow-entries", func(t *testing.T) {
		// EC-4: duplicate allow entries are preserved as written (no
		// de-duplication, no removal) on the write path.
		path := filepath.Join(t.TempDir(), "settings.json")
		before := `{"permissions":{"allow":["Bash(a:*)","Bash(a:*)"]}}`
		if err := os.WriteFile(path, []byte(before), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := WriteUserDefaultMode(path, "acceptEdits"); err != nil {
			t.Fatalf("WriteUserDefaultMode: %v", err)
		}
		afterBody, _ := os.ReadFile(path)
		allow, _, _, _, _ := ac001ReadPerms(t, afterBody)
		if !reflect.DeepEqual(allow, []string{"Bash(a:*)", "Bash(a:*)"}) {
			t.Errorf("duplicate allow entries MUST be preserved as written; after=%s", afterBody)
		}
	})
}

// TestAC002RegionBytesUnchangedWhenDefaultModeMatches: when the resolved tier
// default already equals the file's defaultMode, the writer does not rewrite
// the permissions region — the file bytes are identical (REQ-002).
func TestAC002RegionBytesUnchangedWhenDefaultModeMatches(t *testing.T) {
	dir := t.TempDir()
	userPath := filepath.Join(dir, ".claude", "settings.json")
	if err := os.MkdirAll(filepath.Dir(userPath), 0o755); err != nil {
		t.Fatal(err)
	}
	// Non-canonical layout: the bytes must survive untouched, not be re-rendered.
	before := "{\n  \"permissions\": {\n      \"allow\": [\"Bash(make:*)\"],\n      \"defaultMode\":   \"acceptEdits\"\n  },\n  \"env\": {\"FOO\": \"1\"}\n}\n"
	if err := os.WriteFile(userPath, []byte(before), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := WriteUserDefaultMode(userPath, "acceptEdits"); err != nil {
		t.Fatalf("WriteUserDefaultMode: %v", err)
	}
	after, err := os.ReadFile(userPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != before {
		t.Errorf("file bytes changed\nbefore=%q\nafter=%q", before, string(after))
	}
}

// TestAC003ListsPreservedUnderEitherDisposition: the lists survive whether
// the existing defaultMode differs from or matches the resolved tier default
// (AC-003 part A).
func TestAC003ListsPreservedUnderEitherDisposition(t *testing.T) {
	cases := []struct {
		name     string
		existing string
		resolved string
	}{
		{"differing", "plan", "acceptEdits"},
		{"matching", "acceptEdits", "acceptEdits"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			userPath := filepath.Join(dir, ".claude", "settings.json")
			if err := os.MkdirAll(filepath.Dir(userPath), 0o755); err != nil {
				t.Fatal(err)
			}
			before := `{"permissions": {
      "defaultMode": "` + c.existing + `",
      "allow": ["Bash(make:*)"],
      "ask": ["Bash(rm:*)"],
      "deny": ["Read(./secret)"],
      "additionalDirectories": ["/tmp/x"]
    },"env":{"FOO":"1"}}`
			if err := os.WriteFile(userPath, []byte(before), 0o644); err != nil {
				t.Fatal(err)
			}
			if err := WriteUserDefaultMode(userPath, c.resolved); err != nil {
				t.Fatalf("WriteUserDefaultMode: %v", err)
			}
			after, err := os.ReadFile(userPath)
			if err != nil {
				t.Fatal(err)
			}
			allow, ask, deny, additional, _ := ac001ReadPerms(t, after)
			if !reflect.DeepEqual(allow, []string{"Bash(make:*)"}) {
				t.Errorf("permissions.allow was removed; after=%s", after)
			}
			if !reflect.DeepEqual(ask, []string{"Bash(rm:*)"}) {
				t.Errorf("permissions.ask was removed; after=%s", after)
			}
			if !reflect.DeepEqual(deny, []string{"Read(./secret)"}) {
				t.Errorf("permissions.deny was removed; after=%s", after)
			}
			if !reflect.DeepEqual(additional, []string{"/tmp/x"}) {
				t.Errorf("permissions.additionalDirectories was removed; after=%s", after)
			}
		})
	}
}
