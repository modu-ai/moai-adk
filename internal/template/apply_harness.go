package template

// apply_harness.go — llm.harness persistence (SPEC-INIT-HARNESS-001
// REQ-IH-002). This file is what remains of the former profile_matrix.go
// after SPEC-AGENT-MODEL-INHERIT-001 M5 deleted the per-agent model+effort
// matrix, groups, and resolvers: subagents inherit the main session's model
// and effort, so no per-agent cell survives to patch into llm.yaml. The
// harness key is a different axis (which agent harness a project deploys)
// and stays.

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
)

// llmRootRegex matches the top-level `llm:` key line for key insertion into
// a legacy config that has no such key line to replace.
var llmRootRegex = regexp.MustCompile(`(?m)^llm:[ \t]*$`)

// harnessLineRegex matches the harness: line in llm.yaml for a value-replacing
// write. Group 1 carries the leading indentation, group 2 the raw value (with
// optional surrounding quotes so an already-correct line is recognized and
// left byte-identical — a first deploy must ship the template verbatim, and
// rewriting `harness: "claude"` into `harness: claude` would violate that).
var harnessLineRegex = regexp.MustCompile(`(?m)^(\s*)harness:\s*["']?([\w-]*)["']?`)

// ApplyHarness patches the harness field in llm.yaml under the given project
// root (SPEC-INIT-HARNESS-001 REQ-IH-002). It reads
// .moai/config/sections/llm.yaml, replaces the harness: line with the new
// value (preserving indentation), and writes the file back. Returns nil when
// the file is absent (graceful no-op). The value MUST be one of the closed-set
// harness names (config.IsValidAgentHarness); an out-of-set value is an error,
// never a silent write — llm.harness governs update re-deployment and doctor
// check scoping, so a wrong value here would misdirect both.
//
// @MX:NOTE: [AUTO] llm.harness persistence entry point (SPEC-INIT-HARNESS-001
// REQ-IH-002); init writes the resolved value on every run, claude included.
func ApplyHarness(projectRoot, harness string) error {
	if !config.IsValidAgentHarness(harness) {
		return fmt.Errorf("invalid harness value %q: must be one of claude, codex, both", harness)
	}
	llmPath := sectionPath(projectRoot, "llm.yaml")
	content, err := os.ReadFile(llmPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read llm.yaml: %w", err)
	}

	var newContent []byte
	if m := harnessLineRegex.FindSubmatch(content); m != nil {
		// Already carries the target value (any quoting style): no-op. A
		// first deploy ships the template llm.yaml verbatim — rewriting
		// `harness: "claude"` into `harness: claude` there would break the
		// byte-identity contract the update tests pin.
		if strings.TrimSpace(string(m[2])) == harness {
			return nil
		}
		newContent = harnessLineRegex.ReplaceAll(content, []byte("${1}harness: "+harness))
	} else {
		// A llm.yaml predating this SPEC has no harness key: insert one right
		// under the llm: root so the resolved selection is still recorded
		// (explicit record over implicit absence — REQ-IH-002).
		newContent = llmRootRegex.ReplaceAll(content, []byte("${0}\n  harness: "+harness))
	}
	if string(newContent) == string(content) {
		return nil
	}

	if err := os.WriteFile(llmPath, newContent, 0o644); err != nil {
		return fmt.Errorf("write llm.yaml: %w", err)
	}
	return nil
}
