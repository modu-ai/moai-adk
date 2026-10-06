package auditverdict

// gates.go — the tree's required-backend gate set, resolved for the admission
// seams (SPEC-AUDIT-CEILING-001 REQ-ACE-008/010). The two LIVE admission
// seams (the kickoff evaluator and the card-transition guard) resolve the
// gate set here and pass it to Admit. The resolution keeps the same routing
// internal/cli's resolveAuditGates performs for MCP audits — a config-orphaned
// worktree reads its primary checkout's section — but with the opposite
// failure disposition (D21): a configuration that exists but cannot be read
// or parsed returns an error and refuses, never the empty set. Only a
// genuinely-absent configuration resolves empty, and an empty set admits a
// receipt-less verdict (C4).

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/auditreceipt"
	"github.com/modu-ai/moai-adk/internal/config"
	"gopkg.in/yaml.v3"
)

// GateSet is the resolved required-backend set of one tree. Required is the
// explicit required backends in the plan's stable backend order (empty when
// the tree configures none); Note carries the assumed-required wording of a
// config-orphaned root whose primary checkout could not be identified.
type GateSet struct {
	Required []string
	Note     string
}

// ResolveRequiredBackends resolves the tree's explicit required audit
// backends — the workflow.audit.gates entries and audit.model assignments an
// operator wrote (REQ-ACE-008's resolution subject). Absent configuration
// resolves empty with no error (C4); a section that exists but cannot be
// read, parsed, or resolved returns an error the caller must refuse on
// (REQ-ACE-010, D21).
func ResolveRequiredBackends(treeRoot string) (GateSet, error) {
	if auditreceipt.IsConfigOrphanedRoot(treeRoot) {
		primary, _, err := auditreceipt.IdentifyPrimaryCheckout(treeRoot)
		if err != nil {
			// Parity with resolveAuditGates (REQ-MWU-011/012): the codex gate
			// is treated as required, with the note naming the assumption.
			return GateSet{Required: []string{"codex"}, Note: auditreceipt.GateAssumedRequiredNote}, nil
		}
		treeRoot = primary
	}
	path := filepath.Join(treeRoot, ".moai", "config", "sections", "workflow.yaml")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return GateSet{}, nil
	}
	if err != nil {
		return GateSet{}, fmt.Errorf("read workflow.yaml: %w", err)
	}
	var wrapper struct {
		Workflow struct {
			Audit config.AuditConfig `yaml:"audit"`
		} `yaml:"workflow"`
	}
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		return GateSet{}, fmt.Errorf("parse workflow.yaml: %w", err)
	}
	plan, err := config.ResolveAuditPlan(wrapper.Workflow.Audit, config.AuditGates{})
	if err != nil {
		return GateSet{}, fmt.Errorf("resolve audit plan: %w", err)
	}
	var gs GateSet
	for _, e := range plan.Backends {
		if e.Explicit && e.Gate == config.AuditGateRequired {
			gs.Required = append(gs.Required, e.Backend)
		}
	}
	return gs, nil
}
