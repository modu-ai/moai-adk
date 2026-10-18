package config

import (
	"log/slog"
	"path/filepath"
)

// factoryLeaseFile is the isolated decode target of the lease keys: the only
// workflow.yaml block it maps is factory_lease, so a YAML type error in any
// OTHER key — a card_age_alarm typo included — cannot fail this decode and
// revert the written lease values (sync-audit iter1 F3).
type factoryLeaseFile struct {
	Workflow struct {
		FactoryLease FactoryLeaseConfig `yaml:"factory_lease"`
	} `yaml:"workflow"`
}

// LoadFactoryLease returns the workflow.factory_lease keys of projectRoot as
// written in its workflow.yaml. A key the file omits takes its compiled default
// (NewDefaultWorkflowConfig), and a missing file yields the defaults for both keys.
//
// It reads through the entry point that the other workflow.* consumers on the
// factory path use: loadYAMLFile over <projectRoot>/.moai/config/sections, as
// LoadQuotaGate and LoadSlotLeaseDefaultMaxDuration do. That entry point does not
// consult the MOAI_CONFIG_DIR override, so these keys do not follow it either.
// The factory_lease block is decoded in isolation (factoryLeaseFile), so a type
// error elsewhere in the file degrades only the block that carries it.
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
	file := &factoryLeaseFile{}
	file.Workflow.FactoryLease = defaults
	loaded, err := loadYAMLFile(dir, "workflow.yaml", file)
	if err != nil {
		slog.Warn("failed to load factory lease config, using defaults",
			"path", filepath.Join(dir, "workflow.yaml"), "error", err)
		return defaults
	}
	if !loaded {
		return defaults
	}
	return file.Workflow.FactoryLease
}
