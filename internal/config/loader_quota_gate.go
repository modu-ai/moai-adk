package config

// loader_quota_gate.go — SPEC-QUOTA-AWARE-SCHEDULING-001 M2 (REQ-QAS-008).
//
// A single-block reader for workflow.quota_gate, modelled on
// LoadSlotLeaseDefaultMaxDuration: the consumers (the lane gate, the status
// block, the integration-window warning, the quota aggregator's max age) run
// outside the Loader's lifecycle and hold no loaded *Config. Every failure path
// — missing file, unparseable file — yields the shipped defaults, and every
// out-of-range value yields that one key's default, so a mistyped threshold can
// never silently widen or disable the gate. The wrapper is seeded with the
// defaults so an ABSENT key keeps its default while an explicit value (a
// release margin of 0, in range) is honoured.

import (
	"path/filepath"
	"strings"
	"time"
)

// QuotaGateSettings is the resolved, validated form of workflow.quota_gate: the
// percentages are within their valid ranges and MaxAge is a parsed duration of
// at least twice the heartbeat, so a consumer never re-validates.
type QuotaGateSettings struct {
	Enabled          bool
	FiveHourHoldPct  int
	SevenDayHoldPct  int
	ReleaseMarginPct int
	MaxAge           time.Duration
}

// Valid ranges of the numeric keys (REQ-QAS-008).
const (
	quotaGateHoldPctMin       = 1
	quotaGateHoldPctMax       = 100
	quotaGateReleaseMarginMin = 0
	quotaGateReleaseMarginMax = 50

	// Valid range of max_scan_dirs (SPEC-QUOTA-RECORD-WORKTREES-001 REQ-QWR-011);
	// the upper bound keeps a mistyped value from making the scan unbounded.
	// Both are unmeasured.
	quotaGateScanDirsMin = 1
	quotaGateScanDirsMax = 1024
)

// DefaultQuotaGate returns the shipped defaults: the gate off and the unmeasured
// 90 / 95 / 5 / 30m values defined in defaults.go.
func DefaultQuotaGate() QuotaGateSettings {
	return QuotaGateSettings{
		Enabled:          false,
		FiveHourHoldPct:  DefaultQuotaGateFiveHourHoldPct,
		SevenDayHoldPct:  DefaultQuotaGateSevenDayHoldPct,
		ReleaseMarginPct: DefaultQuotaGateReleaseMarginPct,
		MaxAge:           defaultQuotaGateMaxAge(),
	}
}

// defaultQuotaGateMaxAge parses the one string constant that defines the default
// max age, so the Go default, the template, and the resolved value cannot drift.
func defaultQuotaGateMaxAge() time.Duration {
	d, err := time.ParseDuration(DefaultQuotaGateMaxAge)
	if err != nil {
		return 0
	}
	return d
}

// quotaGateFile is the isolated decode target of the quota keys: the only
// workflow.yaml block it maps is quota_gate, so a YAML type error in any
// OTHER key — a card_age_alarm typo included — cannot fail this decode and
// revert the written quota values to the compiled defaults (sync-audit iter2
// N2; the cardAgeAlarmFile / factoryLeaseFile pattern).
type quotaGateFile struct {
	Workflow struct {
		QuotaGate QuotaGateConfig `yaml:"quota_gate"`
	} `yaml:"workflow"`
}

// LoadQuotaGate resolves workflow.quota_gate for the project rooted at
// projectRoot. The block is decoded in isolation (quotaGateFile), so a type
// error elsewhere in workflow.yaml degrades only the block that carries it.
func LoadQuotaGate(projectRoot string) QuotaGateSettings {
	dir := filepath.Join(projectRoot, ".moai", "config", "sections")
	file := &quotaGateFile{}
	file.Workflow.QuotaGate = NewDefaultWorkflowConfig().QuotaGate
	loaded, err := loadYAMLFile(dir, "workflow.yaml", file)
	if err != nil || !loaded {
		return DefaultQuotaGate()
	}
	return resolveQuotaGate(file.Workflow.QuotaGate)
}

// LoadQuotaScanBound returns the bound on linked-worktree record directories the
// quota reading examines per call (workflow.quota_gate.max_scan_dirs,
// SPEC-QUOTA-RECORD-WORKTREES-001 REQ-QWR-011). It reads the same file through
// the same isolated decode as LoadQuotaGate, but is a separate accessor so that
// QuotaGateSettings carries no bound. A missing, unparseable, or mistyped
// quota_gate block (a decode failure of this block) and a value outside 1-1024
// yield the default; an absent key keeps the default seeded into the decode
// target. A type error in any OTHER workflow block no longer reaches this
// reader (sync-audit iter2 N2).
func LoadQuotaScanBound(projectRoot string) int {
	dir := filepath.Join(projectRoot, ".moai", "config", "sections")
	file := &quotaGateFile{}
	file.Workflow.QuotaGate = NewDefaultWorkflowConfig().QuotaGate
	loaded, err := loadYAMLFile(dir, "workflow.yaml", file)
	if err != nil || !loaded {
		return DefaultQuotaGateMaxScanDirs
	}
	n := file.Workflow.QuotaGate.MaxScanDirs
	if n < quotaGateScanDirsMin || n > quotaGateScanDirsMax {
		return DefaultQuotaGateMaxScanDirs
	}
	return n
}

// resolveQuotaGate replaces every out-of-range value of c with its default.
func resolveQuotaGate(c QuotaGateConfig) QuotaGateSettings {
	out := DefaultQuotaGate()
	out.Enabled = c.Enabled
	if c.FiveHourHoldPct >= quotaGateHoldPctMin && c.FiveHourHoldPct <= quotaGateHoldPctMax {
		out.FiveHourHoldPct = c.FiveHourHoldPct
	}
	if c.SevenDayHoldPct >= quotaGateHoldPctMin && c.SevenDayHoldPct <= quotaGateHoldPctMax {
		out.SevenDayHoldPct = c.SevenDayHoldPct
	}
	if c.ReleaseMarginPct >= quotaGateReleaseMarginMin && c.ReleaseMarginPct <= quotaGateReleaseMarginMax {
		out.ReleaseMarginPct = c.ReleaseMarginPct
	}
	// A max age below twice the heartbeat would let a live session's record go
	// stale between heartbeats and silently disable the gate.
	if d, err := time.ParseDuration(strings.TrimSpace(c.MaxAge)); err == nil && d >= 2*QuotaHeartbeatInterval {
		out.MaxAge = d
	}
	return out
}
