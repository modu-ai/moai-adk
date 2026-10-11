package config

import (
	"log/slog"
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
// A file that cannot be read or decoded yields the defaults for both keys, and a
// slog.Warn names the file and the error. That is the config fallback path the
// workflow section uses, and the factory does not refuse to open on account of
// such a file (SPEC-FACTORY-LEASE-RENEW-001, ruling d-20261010T111725Z-3ef9). The
// values the file does parse are returned as written. Their bounds are checked by
// the caller, and a value the bounds reject refuses the open (REQ-FLR-005 to
// REQ-FLR-009).
func LoadFactoryLease(projectRoot string) FactoryLeaseConfig {
	dir := filepath.Join(projectRoot, ".moai", "config", "sections")
	defaults := NewDefaultWorkflowConfig().FactoryLease
	wrapper := &workflowFileWrapper{Workflow: NewDefaultWorkflowConfig()}
	loaded, err := loadYAMLFile(dir, "workflow.yaml", wrapper)
	if err != nil {
		slog.Warn("failed to load factory lease config, using defaults",
			"path", filepath.Join(dir, "workflow.yaml"), "error", err)
		return defaults
	}
	if !loaded {
		return defaults
	}
	return wrapper.Workflow.FactoryLease
}
