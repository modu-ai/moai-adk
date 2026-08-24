package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/config"
	"gopkg.in/yaml.v3"
)

// audit_pin.go — SPEC-V3R6-AUDIT-MODEL-PIN-001 M1: the workflow.audit pin
// block's section-only load helper, consumed by the codex and GLM AUDIT
// resolvers (mcp_codex.go resolveCodexAuditModelEffort, mcp_glm.go
// resolveGLMAuditModelEffort). The task paths (codex_task / glm_task) never
// call into this file (REQ-AMP-008).
//
// @MX:NOTE: [AUTO] audit_pin.go — workflow.audit {model,effort} pin loader for the audit backends only
// @MX:SPEC: SPEC-V3R6-AUDIT-MODEL-PIN-001

// loadWorkflowAuditSection loads ONLY the workflow.audit block from the
// project's workflow.yaml section file, mirroring the loadLLMSectionOnly
// pattern (glm.go): absent file → zero AuditConfig with NO error (the pin is
// simply absent); read/parse failure → an error the CALLER treats as an absent
// pin, failing open to the legacy SSOT path — never a hard error (N3).
//
// The pin lives in workflow.yaml — NOT llm.yaml — because llm.yaml is
// gitignored and wiped by `moai update`, so a pin there would be uncommitable
// and non-durable (plan-audit MF1 / lead ruling C).
func loadWorkflowAuditSection(projectDir string) (config.AuditConfig, error) {
	workflowPath := filepath.Join(projectDir, ".moai", "config", "sections", "workflow.yaml")

	if _, err := os.Stat(workflowPath); os.IsNotExist(err) {
		return config.AuditConfig{}, nil
	}

	data, err := os.ReadFile(workflowPath)
	if err != nil {
		return config.AuditConfig{}, fmt.Errorf("read workflow.yaml: %w", err)
	}

	wrapper := struct {
		Workflow struct {
			Audit config.AuditConfig `yaml:"audit"`
		} `yaml:"workflow"`
	}{}
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		return config.AuditConfig{}, fmt.Errorf("parse workflow.yaml: %w", err)
	}

	return wrapper.Workflow.Audit, nil
}

// workflowAuditPins resolves the audit pin block for the tree the audit will
// review, failing open: a load error is treated as an ABSENT pin (zero
// AuditConfig), exactly like an absent file, so the resolvers fall back to the
// legacy SSOT path rather than erroring (N3).
func workflowAuditPins(projectDir string) config.AuditConfig {
	audit, err := loadWorkflowAuditSection(projectDir)
	if err != nil {
		return config.AuditConfig{}
	}
	return audit
}
