package template

import (
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// Tests for the moai update strip step that removes the retired per-agent
// model/effort keys from a user's llm.yaml and workflow.yaml
// (SPEC-AGENT-MODEL-INHERIT-001 REQ-AMI-014, design §C / D14).

func writeSection(t *testing.T, root, name, body string) string {
	t.Helper()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func readFile(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

const llmWithRetiredKeys = `llm:
  mode: ""
  claude_bin: ""

  # Per-agent model+effort profile selection.
  # Selects the active column.
  profile: "medium"

  # Legacy performance tier.
  performance_tier: "medium"

  # Profile matrix.
  profiles:
    high:
      manager-spec:    { model: opus,   effort: medium }
    medium:
      manager-spec:    { model: opus,   effort: medium }

  harness_agents:
    high:
      research: { effort: high }

  # Optional per-agent overrides.
  # Example:
  #   agent_overrides:
  #     manager-develop: { model: opus, effort: xhigh }
  agent_overrides: {}

  # GLM backend configuration
  glm:
    base_url: "https://api.z.ai/api/anthropic"
    models:
      high: glm-5.3
`

const llmStripped = `llm:
  mode: ""
  claude_bin: ""

  # GLM backend configuration
  glm:
    base_url: "https://api.z.ai/api/anthropic"
    models:
      high: glm-5.3
`

const workflowWithRetiredKeys = `workflow:
    default_mode: ""
    agent_model_guard:
        enabled: true
    audit:
        codex:
            model: gpt-5-codex
    # workflow_agents: purpose taxonomy -> {model, effort}
    workflow_agents:
        research: { model: opus, effort: high }
    # model_routing: LEGACY flat block.
    model_routing:
        tier_l: { model: opus, effort: high }
    model_routing_profiles:
        medium:
            tier_l: { model: opus, effort: high }
    jev:
        enabled: false
`

const workflowStripped = `workflow:
    default_mode: ""
    audit:
        codex:
            model: gpt-5-codex
    jev:
        enabled: false
`

func TestStripRetiredModelKeys_RemovesEveryKeyAndItsCommentBlock(t *testing.T) {
	root := t.TempDir()
	llm := writeSection(t, root, "llm.yaml", llmWithRetiredKeys)
	wf := writeSection(t, root, "workflow.yaml", workflowWithRetiredKeys)

	present, err := RetiredModelKeysPresent(root, nil)
	if err != nil || !present {
		t.Fatalf("RetiredModelKeysPresent = %v, %v; want true, nil", present, err)
	}

	got, err := StripRetiredModelKeys(root, nil)
	if err != nil {
		t.Fatalf("StripRetiredModelKeys: %v", err)
	}
	want := []RetiredModelKey{
		{Section: "llm.yaml", Key: "llm.profile"},
		{Section: "llm.yaml", Key: "llm.performance_tier"},
		{Section: "llm.yaml", Key: "llm.profiles"},
		{Section: "llm.yaml", Key: "llm.harness_agents"},
		{Section: "llm.yaml", Key: "llm.agent_overrides"},
		{Section: "workflow.yaml", Key: "workflow.agent_model_guard"},
		{Section: "workflow.yaml", Key: "workflow.workflow_agents"},
		{Section: "workflow.yaml", Key: "workflow.model_routing"},
		{Section: "workflow.yaml", Key: "workflow.model_routing_profiles"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("removed = %+v\nwant      %+v", got, want)
	}
	if s := readFile(t, llm); s != llmStripped {
		t.Errorf("llm.yaml after strip:\n%s\nwant:\n%s", s, llmStripped)
	}
	if s := readFile(t, wf); s != workflowStripped {
		t.Errorf("workflow.yaml after strip:\n%s\nwant:\n%s", s, workflowStripped)
	}

	// Idempotent: a second run finds nothing and rewrites nothing.
	present, err = RetiredModelKeysPresent(root, nil)
	if err != nil || present {
		t.Errorf("after strip RetiredModelKeysPresent = %v, %v; want false, nil", present, err)
	}
	again, err := StripRetiredModelKeys(root, nil)
	if err != nil || len(again) != 0 {
		t.Errorf("second strip = %+v, %v; want none", again, err)
	}
}

func TestStripRetiredModelKeys_FlagsUserEditedOverrides(t *testing.T) {
	root := t.TempDir()
	llm := writeSection(t, root, "llm.yaml", "llm:\n  mode: \"\"\n  agent_overrides:\n    manager-develop: { model: opus, effort: xhigh }\n  glm:\n    base_url: x\n")

	got, err := StripRetiredModelKeys(root, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := []RetiredModelKey{{Section: "llm.yaml", Key: "llm.agent_overrides", UserValues: true}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("removed = %+v, want %+v", got, want)
	}
	if s := readFile(t, llm); s != "llm:\n  mode: \"\"\n  glm:\n    base_url: x\n" {
		t.Errorf("llm.yaml after strip = %q", s)
	}
}

// A YAML encoder can write a flow mapping whose closing brace sits on its own
// line at the key's indent (the shape the update merge produced for a user
// agent_overrides entry). The closing line belongs to the key being removed.
func TestStripRetiredModelKeys_MultiLineFlowMapping(t *testing.T) {
	root := t.TempDir()
	llm := writeSection(t, root, "llm.yaml",
		"llm:\n  claude_bin: \"\"\n  agent_overrides: {manager-develop: {model: opus, effort: xhigh},\n    sync-auditor: {model: opus}\n  }\n  # GLM backend configuration\n  glm:\n    base_url: \"x # not a comment\"\n")

	got, err := StripRetiredModelKeys(root, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := []RetiredModelKey{{Section: "llm.yaml", Key: "llm.agent_overrides", UserValues: true}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("removed = %+v, want %+v", got, want)
	}
	s := readFile(t, llm)
	if s != "llm:\n  claude_bin: \"\"\n  # GLM backend configuration\n  glm:\n    base_url: \"x # not a comment\"\n" {
		t.Errorf("llm.yaml after strip = %q", s)
	}
	var doc map[string]any
	if err := yaml.Unmarshal([]byte(s), &doc); err != nil {
		t.Errorf("stripped llm.yaml no longer parses: %v", err)
	}
}

func TestStripRetiredModelKeys_AbsentKeysAndFilesAreNoOps(t *testing.T) {
	// No sections directory at all.
	root := t.TempDir()
	if got, err := StripRetiredModelKeys(root, nil); err != nil || len(got) != 0 {
		t.Errorf("no config: %+v, %v; want none, nil", got, err)
	}

	// Files present without any retired key: byte-identical, no rewrite.
	llmBody := "llm:\n  mode: \"\"\n  # profile: mentioned only in a comment\n  glm:\n    profile: nested-not-top-level\n"
	wfBody := "workflow:\n  audit:\n    glm:\n      model: glm-4.6\n"
	llm := writeSection(t, root, "llm.yaml", llmBody)
	wf := writeSection(t, root, "workflow.yaml", wfBody)
	before, _ := os.Stat(llm)

	present, err := RetiredModelKeysPresent(root, nil)
	if err != nil || present {
		t.Fatalf("RetiredModelKeysPresent = %v, %v; want false, nil", present, err)
	}
	got, err := StripRetiredModelKeys(root, nil)
	if err != nil || len(got) != 0 {
		t.Fatalf("strip = %+v, %v; want none", got, err)
	}
	if readFile(t, llm) != llmBody || readFile(t, wf) != wfBody {
		t.Error("a file without retired keys was changed")
	}
	after, _ := os.Stat(llm)
	if !after.ModTime().Equal(before.ModTime()) {
		t.Error("a file without retired keys was rewritten")
	}
}

func TestStripRetiredModelKeys_KeepsCRLFLineEndings(t *testing.T) {
	root := t.TempDir()
	llm := writeSection(t, root, "llm.yaml", "llm:\r\n  mode: \"\"\r\n  profile: \"high\"\r\n  glm:\r\n    base_url: x\r\n")
	got, err := StripRetiredModelKeys(root, nil)
	if err != nil || len(got) != 1 {
		t.Fatalf("strip = %+v, %v; want one key", got, err)
	}
	if s := readFile(t, llm); s != "llm:\r\n  mode: \"\"\r\n  glm:\r\n    base_url: x\r\n" {
		t.Errorf("llm.yaml after strip = %q", s)
	}
}

// A key the embedded template still ships is left alone: code in the same build
// still reads it, and stripping it would make every update redeploy and
// re-strip it. It is removed once it leaves the template.
func TestStripRetiredModelKeys_LeavesShippedKeys(t *testing.T) {
	root := t.TempDir()
	body := "llm:\n  profile: \"medium\"\n  agent_overrides: {}\n  glm:\n    base_url: x\n"
	llm := writeSection(t, root, "llm.yaml", body)
	shipped := map[string]bool{"llm.profile": true, "llm.agent_overrides": true}

	present, err := RetiredModelKeysPresent(root, shipped)
	if err != nil || present {
		t.Fatalf("RetiredModelKeysPresent with every key shipped = %v, %v; want false, nil", present, err)
	}
	got, err := StripRetiredModelKeys(root, shipped)
	if err != nil || len(got) != 0 || readFile(t, llm) != body {
		t.Fatalf("strip with every key shipped = %+v, %v; file changed: %v", got, err, readFile(t, llm) != body)
	}

	got, err = StripRetiredModelKeys(root, map[string]bool{"llm.profile": true})
	if err != nil {
		t.Fatal(err)
	}
	if want := []RetiredModelKey{{Section: "llm.yaml", Key: "llm.agent_overrides"}}; !reflect.DeepEqual(got, want) {
		t.Errorf("removed = %+v, want %+v", got, want)
	}
	if s := readFile(t, llm); s != "llm:\n  profile: \"medium\"\n  glm:\n    base_url: x\n" {
		t.Errorf("llm.yaml after strip = %q", s)
	}
}

// ShippedRetiredModelKeys reads the embedded template: a key is in the set
// exactly when the embedded section file carries it at the root-child indent.
func TestShippedRetiredModelKeys_ReadsTheEmbeddedTemplate(t *testing.T) {
	shipped, err := ShippedRetiredModelKeys()
	if err != nil {
		t.Fatal(err)
	}
	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatal(err)
	}
	for _, set := range retiredModelKeySets {
		b, err := fs.ReadFile(fsys, ".moai/config/sections/"+set.section)
		if err != nil {
			t.Fatalf("read embedded %s: %v", set.section, err)
		}
		// Independent measurement: the root child indent is the indent of the
		// first indented content line.
		lines := strings.Split(string(b), "\n")
		indent := -1
		for _, l := range lines {
			if tl := strings.TrimSpace(l); tl != "" && !strings.HasPrefix(tl, "#") && leadingWS(l) > 0 {
				indent = leadingWS(l)
				break
			}
		}
		for _, k := range set.keys {
			carried := false
			for _, l := range lines {
				if leadingWS(l) == indent && strings.HasPrefix(strings.TrimSpace(l), k+":") {
					carried = true
					break
				}
			}
			if shipped[set.root+"."+k] != carried {
				t.Errorf("%s.%s: shipped = %v, but the embedded %s carries it = %v", set.root, k, shipped[set.root+"."+k], set.section, carried)
			}
		}
	}
}
