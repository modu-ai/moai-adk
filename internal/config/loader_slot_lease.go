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

// slotLeaseFile is the isolated decode target of the slot-lease keys: the
// only workflow.yaml block it maps is slot_lease, so a YAML type error in any
// OTHER key — a card_age_alarm typo included — cannot fail this decode and
// revert the written default_max_duration (sync-audit iter2 N2; the
// cardAgeAlarmFile / factoryLeaseFile pattern).
type slotLeaseFile struct {
	Workflow struct {
		SlotLease SlotLeaseConfig `yaml:"slot_lease"`
	} `yaml:"workflow"`
}

// LoadSlotLeaseDefaultMaxDuration returns the configured default declared
// maximum duration for slot leases in the project rooted at projectRoot. The
// block is decoded in isolation (slotLeaseFile), so a type error elsewhere in
// workflow.yaml degrades only the block that carries it.
func LoadSlotLeaseDefaultMaxDuration(projectRoot string) string {
	dir := filepath.Join(projectRoot, ".moai", "config", "sections")
	file := &slotLeaseFile{}
	loaded, err := loadYAMLFile(dir, "workflow.yaml", file)
	if err != nil || !loaded {
		return DefaultSlotLeaseMaxDuration
	}
	if v := strings.TrimSpace(file.Workflow.SlotLease.DefaultMaxDuration); v != "" {
		return v
	}
	return DefaultSlotLeaseMaxDuration
}
