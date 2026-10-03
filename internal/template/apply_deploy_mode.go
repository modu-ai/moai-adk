package template

// apply_deploy_mode.go — SPEC-INIT-SHRINK-001 REQ-009 (OD-5 settled (a)
// 2026-10-03): the deployment_mode write side, beside ApplyHarness in the
// same llm.yaml section file. Same shape as apply_harness.go: a closed-set
// check, a line-replacing regex patch that preserves every other byte, and
// an insert-under-the-llm-root fallback for legacy files that predate the
// key. A value must be one of plugin|local (config.IsValidDeployMode); an
// out-of-set value is an error, never a silent write — update's deployer
// selection (REQ-016) and the migration trigger (REQ-015) read this key, so
// a wrong value would misdirect both.

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
)

// deployModeLineRegex matches the deployment_mode: line in llm.yaml for a
// value-replacing write. Group 1 carries the leading indentation, group 2
// the raw value (with optional surrounding quotes so an already-correct line
// is recognized and left byte-identical — the same quoting rule
// ApplyHarness applies to harness:).
var deployModeLineRegex = regexp.MustCompile(`(?m)^(\s*)deployment_mode:\s*["']?([\w-]*)["']?`)

// ApplyDeployMode patches the deployment_mode field in llm.yaml under the
// given project root (SPEC-INIT-SHRINK-001 REQ-009). It reads
// .moai/config/sections/llm.yaml, replaces the deployment_mode: line with
// the new value (preserving indentation), and writes the file back. Returns
// nil when the file is absent (graceful no-op — the caller deploys the
// section file first, exactly as ApplyHarness assumes). The value MUST be
// one of the closed-set names (config.IsValidDeployMode); an out-of-set
// value is an error, never a silent write.
//
// @MX:NOTE: [AUTO] deploy-mode record entry point (SPEC-INIT-SHRINK-001
// REQ-009); init writes the resolved value on every run, and update's
// restore step re-asserts it from the pre-update backup so the
// .moai/config Clean wipe cannot cost the key (OD-5 settled condition).
func ApplyDeployMode(projectRoot, mode string) error {
	if !config.IsValidDeployMode(mode) {
		return fmt.Errorf("invalid deployment_mode value %q: must be one of plugin, local", mode)
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
	if m := deployModeLineRegex.FindSubmatch(content); m != nil {
		// Already carries the target value (any quoting style): no-op. A
		// first deploy ships the template llm.yaml verbatim, and rewriting
		// an already-correct line would break the byte-identity contract the
		// update tests pin.
		if strings.TrimSpace(string(m[2])) == mode {
			return nil
		}
		newContent = deployModeLineRegex.ReplaceAll(content, []byte("${1}deployment_mode: "+mode))
	} else {
		// A llm.yaml predating this SPEC has no deployment_mode key: insert
		// one right under the llm: root so the resolved mode is still
		// recorded (explicit record over implicit absence — REQ-009).
		newContent = llmRootRegex.ReplaceAll(content, []byte("${0}\n  deployment_mode: "+mode))
	}
	if string(newContent) == string(content) {
		return nil
	}

	if err := os.WriteFile(llmPath, newContent, 0o644); err != nil {
		return fmt.Errorf("write llm.yaml: %w", err)
	}
	return nil
}
