package toolpolicy

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// sampleSettingsJSON is a minimal pure-JSON settings.json with a permissions
// block plus other top-level keys (env, hooks) that the codegen MUST preserve.
const sampleSettingsJSON = `{
  "env": {
    "FOO": "bar"
  },
  "permissions": {
    "defaultMode": "acceptEdits",
    "allow": [
      "Read",
      "Bash(git push:*)"
    ],
    "ask": [
      "Bash(rm:*)"
    ],
    "deny": [
      "Bash(git push --force:*)"
    ]
  },
  "hooks": {
    "SessionStart": []
  }
}
`

// sampleSettingsTmpl is a mixed JSON + Go-template-directive settings.json.tmpl.
// The {{jsonEscape .SmartPATH}} directive lives OUTSIDE the permissions block
// (in env.PATH) — the codegen MUST preserve it verbatim.
const sampleSettingsTmpl = `{
  "env": {
    "PATH": "{{jsonEscape .SmartPATH}}",
    "FOO": "bar"
  },
  "permissions": {
    "defaultMode": "acceptEdits",
    "allow": [
      "Read",
      "Bash(git push:*)"
    ],
    "ask": [
      "Bash(rm:*)"
    ],
    "deny": [
      "Bash(git push --force:*)"
    ]
  },
  "hooks": {
    "SessionStart": []
  }
}
`

func writeTempSettings(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return p
}

func sampleDoc() *PolicyDocument {
	return &PolicyDocument{
		Entries: []PolicyEntry{
			{Tool: "Read", ArgsPattern: "", RiskTier: RiskTierRead, Decision: DecisionAllow, OwnerAgent: "orchestrator", Audit: "read"},
			{Tool: "Bash", ArgsPattern: "git push:*", RiskTier: RiskTierWrite, Decision: DecisionAllow, OwnerAgent: "orchestrator", Audit: "git push"},
			{Tool: "Bash", ArgsPattern: "rm:*", RiskTier: RiskTierWrite, Decision: DecisionAsk, OwnerAgent: "orchestrator", Audit: "rm"},
			{Tool: "Bash", ArgsPattern: "git push --force:*", RiskTier: RiskTierIrreversible, Decision: DecisionDeny, OwnerAgent: "orchestrator", Audit: "force-push"},
		},
	}
}

// TestRoundTripEquivalence_JSON (AC-TPS-004) verifies that for every YAML
// entry, the generated settings.json permissions block reflects the YAML
// decision — codegenDecision(entry) == entry.Decision.
func TestRoundTripEquivalence_JSON(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeTempSettings(t, dir, "settings.json", sampleSettingsJSON)
	doc := sampleDoc()

	if _, err := BuildInto(path, doc, TargetJSON, ""); err != nil {
		t.Fatalf("BuildInto: %v", err)
	}

	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read generated: %v", err)
	}
	block, err := extractPermissions(body)
	if err != nil {
		t.Fatalf("extractPermissions: %v", err)
	}
	// Build a set of all (decision → specifiers) from the generated block.
	sets := map[Decision]map[string]bool{
		DecisionAllow: setOf(block.Allow),
		DecisionDeny:  setOf(block.Deny),
		DecisionAsk:   setOf(block.Ask),
	}
	for _, e := range doc.Entries {
		spec := e.SettingsSpecifier()
		if !sets[e.Decision][spec] {
			t.Errorf("entry %q declared %q but not found in generated %q list", spec, e.Decision, e.Decision)
		}
	}
}

// TestDriftPrevention_YamlSettingsRoundTrip (AC-TPS-005) verifies the
// by-construction drift-prevention property: a YAML edit propagates to the
// settings.json permissions block in one codegen pass. Flip Bash(git push:*)
// from allow → ask, regenerate, assert the entry moved lists.
func TestDriftPrevention_YamlSettingsRoundTrip(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeTempSettings(t, dir, "settings.json", sampleSettingsJSON)
	doc := sampleDoc()

	// First pass: baseline.
	if _, err := BuildInto(path, doc, TargetJSON, ""); err != nil {
		t.Fatalf("BuildInto baseline: %v", err)
	}

	// Flip Bash(git push:*) from allow → ask in the YAML.
	for i := range doc.Entries {
		if doc.Entries[i].Tool == "Bash" && doc.Entries[i].ArgsPattern == "git push:*" {
			doc.Entries[i].Decision = DecisionAsk
		}
	}

	// Regenerate.
	if _, err := BuildInto(path, doc, TargetJSON, ""); err != nil {
		t.Fatalf("BuildInto after flip: %v", err)
	}

	// Assert the entry moved from allow to ask.
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	block, err := extractPermissions(body)
	if err != nil {
		t.Fatalf("extractPermissions: %v", err)
	}
	if contains(block.Allow, "Bash(git push:*)") {
		t.Error("Bash(git push:*) still in allow list after flip to ask")
	}
	if !contains(block.Ask, "Bash(git push:*)") {
		t.Error("Bash(git push:*) not in ask list after flip to ask")
	}
}

// TestCodegenIdempotency (AC-TPS-013) verifies that two consecutive codegen
// invocations with no YAML change produce byte-identical output.
func TestCodegenIdempotency(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeTempSettings(t, dir, "settings.json", sampleSettingsJSON)
	doc := sampleDoc()

	if _, err := BuildInto(path, doc, TargetJSON, ""); err != nil {
		t.Fatalf("BuildInto run1: %v", err)
	}
	run1, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read run1: %v", err)
	}

	if _, err := BuildInto(path, doc, TargetJSON, ""); err != nil {
		t.Fatalf("BuildInto run2: %v", err)
	}
	run2, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read run2: %v", err)
	}

	if !bytes.Equal(run1, run2) {
		t.Errorf("non-idempotent: run1 and run2 differ (len %d vs %d)", len(run1), len(run2))
	}
}

// TestCodegenPreservesNonPermissionsRegions (KI-1 / AP-1 anti-pattern) verifies
// the codegen touches ONLY the permissions block — the env, hooks, and other
// top-level keys MUST be preserved verbatim.
func TestCodegenPreservesNonPermissionsRegions(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeTempSettings(t, dir, "settings.json", sampleSettingsJSON)
	doc := sampleDoc()

	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read before: %v", err)
	}
	var beforeJSON map[string]any
	if err := json.Unmarshal(before, &beforeJSON); err != nil {
		t.Fatalf("before is not valid JSON: %v", err)
	}

	if _, err := BuildInto(path, doc, TargetJSON, ""); err != nil {
		t.Fatalf("BuildInto: %v", err)
	}

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read after: %v", err)
	}
	var afterJSON map[string]any
	if err := json.Unmarshal(after, &afterJSON); err != nil {
		t.Fatalf("after is not valid JSON: %v", err)
	}

	// env and hooks MUST be preserved.
	if !deepEqual(beforeJSON["env"], afterJSON["env"]) {
		t.Errorf("env region changed:\nbefore=%v\nafter=%v", beforeJSON["env"], afterJSON["env"])
	}
	if !deepEqual(beforeJSON["hooks"], afterJSON["hooks"]) {
		t.Errorf("hooks region changed:\nbefore=%v\nafter=%v", beforeJSON["hooks"], afterJSON["hooks"])
	}
}

// TestTemplateRoundTripEquivalence (AC-TPS-014c) verifies the YAML→.tmpl
// round-trip is decision-equivalent.
func TestTemplateRoundTripEquivalence(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeTempSettings(t, dir, "settings.json.tmpl", sampleSettingsTmpl)
	doc := sampleDoc()

	if _, err := BuildInto(path, doc, TargetTemplate, ""); err != nil {
		t.Fatalf("BuildInto tmpl: %v", err)
	}

	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read tmpl: %v", err)
	}
	block, err := extractPermissions(body)
	if err != nil {
		t.Fatalf("extractPermissions: %v", err)
	}
	sets := map[Decision]map[string]bool{
		DecisionAllow: setOf(block.Allow),
		DecisionDeny:  setOf(block.Deny),
		DecisionAsk:   setOf(block.Ask),
	}
	for _, e := range doc.Entries {
		spec := e.SettingsSpecifier()
		if !sets[e.Decision][spec] {
			t.Errorf("tmpl entry %q declared %q but not in generated %q list", spec, e.Decision, e.Decision)
		}
	}
}

// TestTemplateDirectivePreserved (AC-TPS-014a) verifies the {{jsonEscape
// .SmartPATH}} directive is preserved verbatim after codegen on the .tmpl.
func TestTemplateDirectivePreserved(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeTempSettings(t, dir, "settings.json.tmpl", sampleSettingsTmpl)
	doc := sampleDoc()

	if _, err := BuildInto(path, doc, TargetTemplate, ""); err != nil {
		t.Fatalf("BuildInto: %v", err)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if got := bytes.Count(body, []byte("{{jsonEscape .SmartPATH}}")); got != 1 {
		t.Errorf("{{jsonEscape .SmartPATH}} count = %d; want 1 (directive must be preserved exactly once)", got)
	}
}

// TestPermissionBlockNoTemplateSentinel (AC-TPS-014b) verifies the regenerated
// permissions block contains ZERO {{ or }} substrings (no template directive
// leakage into the permissions block).
func TestPermissionBlockNoTemplateSentinel(t *testing.T) {
	t.Parallel()
	// Render a block that contains both env_gate and regular entries; assert no
	// template directive leaks into the rendered permissions object.
	doc := sampleDoc()
	block, _, err := BuildPermissions(doc, "acceptEdits")
	if err != nil {
		t.Fatalf("BuildPermissions: %v", err)
	}
	rendered, err := renderPermissionsObject(block)
	if err != nil {
		t.Fatalf("renderPermissionsObject: %v", err)
	}
	if err := AssertPermissionBlockNoTemplate(rendered); err != nil {
		t.Errorf("AssertPermissionBlockNoTemplate failed: %v", err)
	}
}

// TestEnvGatedSkipped verifies that entries carrying an EnvGate are NOT emitted
// as static specifiers (a static deny would over-block the cg-leader exception;
// design.md §C.4).
func TestEnvGatedSkipped(t *testing.T) {
	t.Parallel()
	doc := &PolicyDocument{
		Entries: []PolicyEntry{
			{Tool: "WebSearch", ArgsPattern: "*", RiskTier: RiskTierRead, Decision: DecisionDeny, OwnerAgent: "orch", Audit: "glm", EnvGate: &EnvGate{BaseURLContains: "api.z.ai"}},
			{Tool: "Read", ArgsPattern: "", RiskTier: RiskTierRead, Decision: DecisionAllow, OwnerAgent: "orch", Audit: "read"},
		},
	}
	block, res, err := BuildPermissions(doc, "")
	if err != nil {
		t.Fatalf("BuildPermissions: %v", err)
	}
	if res.EnvGatedSkipped != 1 {
		t.Errorf("EnvGatedSkipped = %d; want 1", res.EnvGatedSkipped)
	}
	if len(block.Deny) != 0 {
		t.Errorf("env-gated deny leaked into static deny list: %v", block.Deny)
	}
	if len(block.Allow) != 1 || block.Allow[0] != "Read" {
		t.Errorf("allow list = %v; want [Read]", block.Allow)
	}
}

// TestBuildIntoAutoDetectsKind verifies DetectTargetKind picks the right
// strategy based on file contents (JSON vs mixed-directive .tmpl).
func TestBuildIntoAutoDetectsKind(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	jsonPath := writeTempSettings(t, dir, "settings.json", sampleSettingsJSON)
	tmplPath := writeTempSettings(t, dir, "settings.json.tmpl", sampleSettingsTmpl)
	doc := sampleDoc()

	r1, err := BuildIntoAuto(jsonPath, doc, "")
	if err != nil {
		t.Fatalf("BuildIntoAuto json: %v", err)
	}
	if r1.TargetKind != TargetJSON {
		t.Errorf("json detected as %q; want %q", r1.TargetKind, TargetJSON)
	}

	r2, err := BuildIntoAuto(tmplPath, doc, "")
	if err != nil {
		t.Fatalf("BuildIntoAuto tmpl: %v", err)
	}
	if r2.TargetKind != TargetTemplate {
		t.Errorf("tmpl detected as %q; want %q", r2.TargetKind, TargetTemplate)
	}
}

// TestLocatePermissionsRegion_HandlesBracesInStrings verifies the raw-text
// region matcher is not fooled by braces inside JSON string values (e.g., a
// deny pattern containing "}"). This is the load-bearing safety property for
// raw-text region replacement on the .tmpl target.
func TestLocatePermissionsRegion_HandlesBracesInStrings(t *testing.T) {
	t.Parallel()
	// A deny pattern containing "{" and "}" that must NOT be counted as depth.
	body := []byte(`{
  "env": {"X": "}"},
  "permissions": {
    "defaultMode": "default",
    "deny": ["Bash({ :*)"]
  }
}
`)
	region, err := locatePermissionsRegion(body)
	if err != nil {
		t.Fatalf("locatePermissionsRegion: %v", err)
	}
	// The region end must be the closing brace of the permissions object, not
	// a brace inside a string literal.
	extracted := body[region.start:region.end]
	if !strings.HasPrefix(string(extracted), `"permissions"`) {
		t.Errorf("region does not start at \"permissions\" key: %q", string(extracted[:20]))
	}
	// Verify the extracted region is itself valid JSON when we prepend the key.
	block, err := extractPermissions(body)
	if err != nil {
		t.Fatalf("extractPermissions: %v", err)
	}
	if len(block.Deny) != 1 || block.Deny[0] != "Bash({ :*)" {
		t.Errorf("deny = %v; want [Bash({ :*)]", block.Deny)
	}
}

// TestExtractPermissions_PreservesExtraKeys verifies that permission sub-keys
// beyond defaultMode/allow/ask/deny (e.g., additionalDirectories) are
// preserved through the round-trip.
func TestExtractPermissions_PreservesExtraKeys(t *testing.T) {
	t.Parallel()
	body := []byte(`{
  "permissions": {
    "defaultMode": "default",
    "allow": ["Read"],
    "additionalDirectories": ["/tmp"]
  }
}
`)
	block, err := extractPermissions(body)
	if err != nil {
		t.Fatalf("extractPermissions: %v", err)
	}
	if _, ok := block.Raw["additionalDirectories"]; !ok {
		t.Error("additionalDirectories key was dropped from Raw map")
	}
	rendered, err := renderPermissionsObject(block)
	if err != nil {
		t.Fatalf("renderPermissionsObject: %v", err)
	}
	if !strings.Contains(string(rendered), "additionalDirectories") {
		t.Errorf("rendered block dropped additionalDirectories: %s", rendered)
	}
}

// TestCompatBaseline is the backward-compatibility characterization test
// (AC-TPS-007): the codegen does not change existing decisions. The baseline
// here is the sampleSettingsJSON allow/ask/deny sets; after codegen with the
// equivalent YAML entries, every specifier stays in its original list.
func TestCompatBaseline(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeTempSettings(t, dir, "settings.json", sampleSettingsJSON)
	doc := sampleDoc()

	// Baseline (pre-migration).
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read before: %v", err)
	}
	beforeBlock, err := extractPermissions(before)
	if err != nil {
		t.Fatalf("extract before: %v", err)
	}
	beforeSets := map[Decision]map[string]bool{
		DecisionAllow: setOf(beforeBlock.Allow),
		DecisionDeny:  setOf(beforeBlock.Deny),
		DecisionAsk:   setOf(beforeBlock.Ask),
	}

	// Run codegen.
	if _, err := BuildInto(path, doc, TargetJSON, ""); err != nil {
		t.Fatalf("BuildInto: %v", err)
	}

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read after: %v", err)
	}
	afterBlock, err := extractPermissions(after)
	if err != nil {
		t.Fatalf("extract after: %v", err)
	}
	afterSets := map[Decision]map[string]bool{
		DecisionAllow: setOf(afterBlock.Allow),
		DecisionDeny:  setOf(afterBlock.Deny),
		DecisionAsk:   setOf(afterBlock.Ask),
	}

	// For every specifier in the baseline, assert it is still in the SAME
	// decision list after codegen (no allow→deny or allow→ask flips).
	for decision, specs := range beforeSets {
		for spec := range specs {
			if !afterSets[decision][spec] {
				t.Errorf("compat regression: %q was %q before codegen but not after", spec, decision)
			}
		}
	}
}

// --- helpers ---

func setOf(items []string) map[string]bool {
	m := make(map[string]bool, len(items))
	for _, s := range items {
		m[s] = true
	}
	return m
}

func contains(items []string, want string) bool {
	for _, s := range items {
		if s == want {
			return true
		}
	}
	return false
}

// deepEqual is a minimal any-tree equality used for the preserve-regions test.
func deepEqual(a, b any) bool {
	switch av := a.(type) {
	case map[string]any:
		bv, ok := b.(map[string]any)
		if !ok || len(av) != len(bv) {
			return false
		}
		for k, v := range av {
			if !deepEqual(v, bv[k]) {
				return false
			}
		}
		return true
	case []any:
		bv, ok := b.([]any)
		if !ok || len(av) != len(bv) {
			return false
		}
		for i := range av {
			if !deepEqual(av[i], bv[i]) {
				return false
			}
		}
		return true
	default:
		return a == b
	}
}
