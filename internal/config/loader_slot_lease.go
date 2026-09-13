package config

// loader_slot_lease.go — card t607.
//
// A single-key reader for workflow.slot_lease.default_max_duration, modelled
// on LoadGitFlowDevelopBranch: the consumer (`moai slot acquire` without
// --max-duration) runs outside the Loader's lifecycle and holds no loaded
// *Config. Every failure path — missing file, unparseable file, absent or
// blank value — yields DefaultSlotLeaseMaxDuration, the one place the default
// is defined. The value is returned unparsed; the caller parses and reports.

import (
	"path/filepath"
	"strings"
)

// LoadSlotLeaseDefaultMaxDuration returns the configured default declared
// maximum duration for slot leases in the project rooted at projectRoot.
func LoadSlotLeaseDefaultMaxDuration(projectRoot string) string {
	dir := filepath.Join(projectRoot, ".moai", "config", "sections")
	wrapper := &workflowFileWrapper{}
	loaded, err := loadYAMLFile(dir, "workflow.yaml", wrapper)
	if err != nil || !loaded {
		return DefaultSlotLeaseMaxDuration
	}
	if v := strings.TrimSpace(wrapper.Workflow.SlotLease.DefaultMaxDuration); v != "" {
		return v
	}
	return DefaultSlotLeaseMaxDuration
}
