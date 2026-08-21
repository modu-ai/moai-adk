// agentemit_edge_test.go — MS1 edge-path specification tests: the P-04
// flat-prefix layout fallback, the sandbox_mode emission complement (the
// ship-omitted rule's flip side), the manifest fail-closed validator, and
// the loader's structural error paths.
package agentemit_test

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/modu-ai/moai-adk/internal/template/agentemit"
)

// TestEmitAllFlatPrefixLayoutFallback covers the P-04 fallback layout: when
// the manifest selects flat_prefix, TOMLs land directly under .codex/agents/
// with the moai- filename prefix instead of a moai/ subdirectory.
func TestEmitAllFlatPrefixLayoutFallback(t *testing.T) {
	man := mustManifest(t)
	man.Layout.Mode = "flat_prefix"
	set := fstest.MapFS{
		"agents/solo.md": &fstest.MapFile{Data: []byte(fixtureMD(
			"solo", "d.\n", "Read, Bash", "inherit", "low", nil, "b\n",
		))},
	}
	pub, err := agentemit.EmitAll(set, "agents", man)
	if err != nil {
		t.Fatalf("EmitAll: %v", err)
	}
	if _, ok := pub.CodexTOML[".codex/agents/moai-solo.toml"]; !ok {
		t.Errorf("flat_prefix layout must emit .codex/agents/moai-solo.toml; have %v", keysOf(pub.CodexTOML))
	}
}

// TestEmitAllSandboxEmittedWhenConfirmed is the complement of the
// ship-omitted test: a manifest that HAS confirmed the sandbox value set
// emits sandbox_mode with exactly the configured value on every agent.
func TestEmitAllSandboxEmittedWhenConfirmed(t *testing.T) {
	man := mustManifest(t)
	fc, ok := man.Fields["sandbox_mode"]
	if !ok {
		t.Fatal("manifest has no sandbox_mode field config")
	}
	fc.Emit = true
	fc.Value = "workspace-write"
	set := fstest.MapFS{
		"agents/solo.md": &fstest.MapFile{Data: []byte(fixtureMD(
			"solo", "d.\n", "Read, Bash", "inherit", "low", nil, "b\n",
		))},
	}
	pub, err := agentemit.EmitAll(set, "agents", man)
	if err != nil {
		t.Fatalf("EmitAll: %v", err)
	}
	doc := mustDecoded(t, pub, ".codex/agents/moai/solo.toml")
	if got := doc["sandbox_mode"]; got != "workspace-write" {
		t.Errorf("sandbox_mode = %#v, want workspace-write", doc["sandbox_mode"])
	}
}

// TestParseManifestFailClosed drives every manifest self-validation branch:
// a manifest that is structurally invalid must be rejected, never repaired.
func TestParseManifestFailClosed(t *testing.T) {
	// minimalValid is the smallest manifest that passes self-validation.
	const minimalValid = `
codex_measured_version: "0.147.0"
layout:
  mode: subdirectory
  subdirectory: moai
  flat_prefix: "moai-"
fields:
  model_reasoning_effort:
    emit: true
    map:
      low: low
tool_classes:
  Read: file-read
classes:
  - class: file-read
    disposition: no-field
    rationale: built-in read
`
	if _, err := agentemit.ParseManifest([]byte(minimalValid)); err != nil {
		t.Fatalf("minimal valid manifest rejected: %v", err)
	}
	cases := map[string]struct {
		repl func(s string) string
		want string
	}{
		"missing measured version": {func(s string) string { return strings.Replace(s, "codex_measured_version: \"0.147.0\"\n", "", 1) }, "codex_measured_version"},
		"invalid layout mode":      {func(s string) string { return strings.Replace(s, "mode: subdirectory", "mode: diagonal", 1) }, "layout"},
		"no tool classes": {func(s string) string {
			return strings.Replace(s, "tool_classes:\n  Read: file-read\n", "tool_classes: {}\n", 1)
		}, "tool_classes"},
		"class without rationale": {func(s string) string { return strings.Replace(s, "    rationale: built-in read\n", "", 1) }, "rationale"},
		"invalid disposition":     {func(s string) string { return strings.Replace(s, "disposition: no-field", "disposition: maybe", 1) }, "disposition"},
		"tool class without disposition row": {
			func(s string) string {
				return strings.Replace(s, "  Read: file-read", "  Read: file-read\n  Bash: shell", 1)
			}, "shell",
		},
		"effort emitted with empty map": {
			func(s string) string { return strings.Replace(s, "    map:\n      low: low\n", "", 1) }, "map",
		},
		"unparseable yaml": {func(s string) string { return "\tbroken: [" }, "parse"},
		"sandbox emitted outside the measured value set": {
			func(s string) string {
				return strings.Replace(s, "tool_classes:",
					"  sandbox_mode:\n    emit: true\n    value: made-up-sandbox\n    accepted_values:\n      - read-only\n      - workspace-write\n      - danger-full-access\ntool_classes:", 1)
			}, "made-up-sandbox",
		},
	}
	for label, tc := range cases {
		_, err := agentemit.ParseManifest([]byte(tc.repl(minimalValid)))
		if err == nil {
			t.Errorf("%s: want error, got nil", label)
			continue
		}
		if !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: error %q must mention %q", label, err, tc.want)
		}
	}
}

// TestParseAgentDocStructuralErrors covers the loader's structural
// fail-closed paths: missing opening delimiter, missing closing delimiter,
// and missing tools line.
func TestParseAgentDocStructuralErrors(t *testing.T) {
	cases := map[string]struct {
		data string
		want string
	}{
		"missing opening delimiter": {"name: x\n---\nbody\n", "opening"},
		"missing closing delimiter": {"---\nname: x\n", "closing"},
		"missing tools": {
			"---\nname: solo\ndescription: |\n  d\nmodel: inherit\neffort: low\n---\nbody\n",
			"tools",
		},
		"not yaml frontmatter": {"---\nname: [broken\n---\nbody\n", "parse"},
	}
	for label, tc := range cases {
		_, err := agentemit.ParseAgentDoc("solo.md", []byte(tc.data))
		if err == nil {
			t.Errorf("%s: want error, got nil", label)
			continue
		}
		if !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: error %q must mention %q", label, err, tc.want)
		}
	}
}

// TestEmitAllRejectsUnrepresentableBodies covers the remaining writer
// fail-closed validators: a body carrying control bytes (other than
// tab/newline) and a body carrying a carriage return.
func TestEmitAllRejectsUnrepresentableBodies(t *testing.T) {
	man := mustManifest(t)
	ctrlBody := fstest.MapFS{
		"agents/ctrl.md": &fstest.MapFile{Data: []byte(fixtureMD(
			"ctrl", "d.\n", "Read", "inherit", "low", nil, "has \x00 nul\n",
		))},
	}
	pub, err := agentemit.EmitAll(ctrlBody, "agents", man)
	if err == nil || pub != nil {
		t.Errorf("control-byte body: want fail-closed error, got err=%v pub=%v", err, pub)
	} else if !strings.Contains(err.Error(), "control byte") {
		t.Errorf("control-byte body: error %q must name the control byte", err)
	}
	crBody := fstest.MapFS{
		"agents/cr.md": &fstest.MapFile{Data: []byte(fixtureMD(
			"cr", "d.\n", "Read", "inherit", "low", nil, "carriage\rreturn\n",
		))},
	}
	pub, err = agentemit.EmitAll(crBody, "agents", man)
	if err == nil || pub != nil {
		t.Errorf("carriage-return body: want fail-closed error, got err=%v pub=%v", err, pub)
	} else if !strings.Contains(err.Error(), "carriage return") {
		t.Errorf("carriage-return body: error %q must name the carriage return", err)
	}
}
