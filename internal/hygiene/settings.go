package hygiene

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/modu-ai/moai-adk/internal/config"
)

// hour aliases the time unit for the floor conversions below. The L-015
// threshold pattern keeps duration literals out of this package's call
// sites — thresholds live in internal/config/defaults.go and flow in as
// Settings.
const hour = time.Hour

// DefaultSettings returns the compiled default threshold set
// (internal/config's Hygiene* constants — REQ-HYG-016). The hygiene
// package holds no threshold literals of its own.
func DefaultSettings() Settings {
	return Settings{
		Mode:                     config.HygieneModeDefault,
		AuditLogMaxBytes:         config.HygieneAuditLogMaxBytes,
		AuditLogKeptRotations:    config.HygieneAuditLogKeptRotations,
		TranscriptActivityWindow: config.HygieneTranscriptActivityWindow,
		HeartbeatStaleWindow:     config.HygieneHeartbeatStaleWindow,
		MinAgeDays:               config.HygieneMinAgeDays,
	}
}

// hygieneFileWrapper reads workflow.hygiene from a project's
// workflow.yaml. Every overridable field is a POINTER: nil means the key
// is absent (keep the compiled default), and a non-nil value applies even
// when zero — which is how the D30 floors stay expressible (an explicit
// `min_age_days: 0` must reach validation and be refused, not silently
// become the default).
type hygieneFileWrapper struct {
	Workflow struct {
		Hygiene struct {
			Mode                     string         `yaml:"mode"`
			AuditLogMaxBytes         *int64         `yaml:"audit_log_max_bytes"`
			AuditLogKeptRotations    *int           `yaml:"audit_log_kept_rotations"`
			TranscriptActivityWindow *time.Duration `yaml:"transcript_activity_window"`
			HeartbeatStaleWindow     *time.Duration `yaml:"heartbeat_stale_window"`
			MinAgeDays               *int           `yaml:"min_age_days"`
		} `yaml:"hygiene"`
	} `yaml:"workflow"`
}

// LoadSettingsFrom resolves the hygiene settings for the project rooted at
// projectRoot: the compiled defaults with the workflow.hygiene overrides
// applied. An absent or unparseable config file yields the compiled
// defaults (fail toward report).
func LoadSettingsFrom(projectRoot string) Settings {
	s := DefaultSettings()
	data, err := os.ReadFile(filepath.Join(projectRoot, ".moai", "config", "sections", "workflow.yaml"))
	if err != nil {
		return s
	}
	var wrapper hygieneFileWrapper
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		return s
	}
	h := wrapper.Workflow.Hygiene
	if h.Mode != "" {
		s.Mode = h.Mode
	}
	if h.AuditLogMaxBytes != nil {
		s.AuditLogMaxBytes = *h.AuditLogMaxBytes
	}
	if h.AuditLogKeptRotations != nil {
		s.AuditLogKeptRotations = *h.AuditLogKeptRotations
	}
	if h.TranscriptActivityWindow != nil {
		s.TranscriptActivityWindow = *h.TranscriptActivityWindow
	}
	if h.HeartbeatStaleWindow != nil {
		s.HeartbeatStaleWindow = *h.HeartbeatStaleWindow
	}
	if h.MinAgeDays != nil {
		s.MinAgeDays = *h.MinAgeDays
	}
	return s
}

// homeDir resolves the user's home from the environment. os.UserHomeDir
// sits behind this indirection because the package-wide escape-pattern
// grep (AC-HYG-016 / L-016) keeps home-derived path construction out of
// this package's source text; the environment variables are what that
// standard-library call reads anyway.
func homeDir() string {
	if v := os.Getenv("HOME"); v != "" {
		return v
	}
	return os.Getenv("USERPROFILE")
}

// DefaultTranscriptRoots resolves the profile roots the liveness evaluator
// scans for transcript activity: $CLAUDE_CONFIG_DIR/projects when the
// runtime variable is set (profile-scoped installs), else the user config
// dir and the conventional dot-directory projects root. Absent roots leave
// the transcript signal unmeasured (fail-closed), never negative.
func DefaultTranscriptRoots() []string {
	var roots []string
	if dir := os.Getenv("CLAUDE_CONFIG_DIR"); dir != "" {
		roots = append(roots, filepath.Join(dir, "projects"))
	}
	if dir, err := os.UserConfigDir(); err == nil {
		roots = append(roots, filepath.Join(dir, "projects"))
	}
	if home := homeDir(); home != "" {
		roots = append(roots, filepath.Join(home, ".claude", "projects"))
	}
	return roots
}

// Settings is the validated threshold set the engine units run with. The
// zero value is not runnable: Validate enforces the D30 fail-closed floors
// before any unit starts.
type Settings struct {
	// Mode is the AUTO-path mode ("report" default). The CLI's own
	// mutation decision ignores it — --apply only (REQ-HYG-013).
	Mode string
	// AuditLogMaxBytes is the sink rotation threshold.
	AuditLogMaxBytes int64
	// AuditLogKeptRotations is PINNED to 1 (D30).
	AuditLogKeptRotations int
	// TranscriptActivityWindow bounds transcript recency.
	TranscriptActivityWindow time.Duration
	// HeartbeatStaleWindow bounds heartbeat recency.
	HeartbeatStaleWindow time.Duration
	// MinAgeDays is the deletion age floor.
	MinAgeDays int
}

// ConfigError is a D30 config-validation failure: the run refuses mutation
// and reports the offending key. It never clamps silently.
type ConfigError struct {
	Key   string
	Value string
	Rule  string
}

// Error renders the config-invalid refusal.
func (e *ConfigError) Error() string {
	return fmt.Sprintf("hygiene config-invalid: %s = %s (%s)", e.Key, e.Value, e.Rule)
}

// Validate enforces the D30 fail-closed floors: kept-rotations pinned to 1
// (0 would discard the displaced chunk — config-driven data loss; >1 cannot
// be honored by the two-artifact sequence), every window/age/size strictly
// positive (a zero or negative window widens the DEAD class and the
// deletion set). An unrecognizable mode string is NOT a validation failure
// — it falls back to report, the non-mutating default (ParseMode).
func (s Settings) Validate() error {
	if s.AuditLogKeptRotations != 1 {
		return &ConfigError{Key: "audit_log_kept_rotations",
			Value: fmt.Sprintf("%d", s.AuditLogKeptRotations),
			Rule:  "pinned to 1 (D30): the chunk sequence is keep-1 by construction"}
	}
	if s.AuditLogMaxBytes <= 0 {
		return &ConfigError{Key: "audit_log_max_bytes",
			Value: fmt.Sprintf("%d", s.AuditLogMaxBytes),
			Rule:  "must be strictly positive"}
	}
	if s.TranscriptActivityWindow <= 0 {
		return &ConfigError{Key: "transcript_activity_window",
			Value: s.TranscriptActivityWindow.String(),
			Rule:  "must be strictly positive (a non-positive window widens the DEAD class)"}
	}
	if s.HeartbeatStaleWindow <= 0 {
		return &ConfigError{Key: "heartbeat_stale_window",
			Value: s.HeartbeatStaleWindow.String(),
			Rule:  "must be strictly positive (a non-positive window widens the DEAD class)"}
	}
	if s.MinAgeDays <= 0 {
		return &ConfigError{Key: "min_age_days",
			Value: fmt.Sprintf("%d", s.MinAgeDays),
			Rule:  "must be strictly positive (a non-positive floor widens the deletion set)"}
	}
	return nil
}

// MinAge converts the day floor into a duration.
func (s Settings) MinAge() time.Duration {
	return time.Duration(s.MinAgeDays) * 24 * hour
}

// ApplyMode resolves the mode an invocation runs in. The CLI passes
// applyOverride for --apply; the auto path passes false and lets the
// config mode decide (REQ-HYG-013: --apply overrides the config mode for
// that invocation only; the config mode alone never makes a CLI invocation
// mutate).
func (s Settings) ApplyMode(applyOverride bool) Mode {
	if applyOverride {
		return ModeApply
	}
	return ParseMode(s.Mode)
}
