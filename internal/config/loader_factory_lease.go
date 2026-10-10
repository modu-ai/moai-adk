package config

import (
	"path/filepath"
)

// LoadFactoryLease returns the workflow.factory_lease keys of projectRoot as
// written in its workflow.yaml. A key the file omits takes its compiled default
// (NewDefaultWorkflowConfig), and a missing file yields the defaults for both keys.
//
// It reads through the entry point that the other workflow.* consumers on the
// factory path use: loadYAMLFile over <projectRoot>/.moai/config/sections, as
// LoadQuotaGate and LoadSlotLeaseDefaultMaxDuration do. That entry point does not
// consult the MOAI_CONFIG_DIR override, so these keys do not follow it either.
//
// Unlike those siblings, a file that cannot be read or decoded is returned as an
// error rather than replaced by the defaults, so that the caller can refuse to open
// the factory instead of silently running on values the operator did not choose.
// Bounds are checked by the caller (SPEC-FACTORY-LEASE-RENEW-001, REQ-FLR-005 to
// REQ-FLR-009).
func LoadFactoryLease(projectRoot string) (FactoryLeaseConfig, error) {
	dir := filepath.Join(projectRoot, ".moai", "config", "sections")
	wrapper := &workflowFileWrapper{Workflow: NewDefaultWorkflowConfig()}
	loaded, err := loadYAMLFile(dir, "workflow.yaml", wrapper)
	if err != nil {
		return FactoryLeaseConfig{}, err
	}
	if !loaded {
		return NewDefaultWorkflowConfig().FactoryLease, nil
	}
	return wrapper.Workflow.FactoryLease, nil
}
