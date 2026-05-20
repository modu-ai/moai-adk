// Package seeds — M5 cold-start seed loader interface + stub (REQ-HRA-022..024).
//
// W3 ships the Loader interface and stub DetectProjectType only.
// Full marker-based project detection (go.mod / package.json / Cargo.toml / etc.)
// and seed library content are W4 (SPEC-V3R5-PROJECT-MEGA-001) scope.
//
// Dual-path SSOT/cache model (plan.md §4.2):
//   - SSOT:  .claude/skills/moai-meta-harness/seeds/ (core repo, moai update)
//   - Cache: .moai/harness/seeds/ (per-project runtime, populated by W4 /moai project)
//
// Precedence: project-local cache > SSOT (same seed ID, cache wins).
//
// @MX:ANCHOR: [AUTO] Loader.LoadForProject is the M5 seed entry point.
// @MX:REASON: [AUTO] fan_in >= 3: loader_test.go, integration_test.go, W4 caller
package seeds

// Seed is a single cold-start lesson seed (schema v1, plan.md §4.1).
// W4 will maintain backward compatibility via the Version field.
//
// YAML field names use snake_case (spec.md §1.7 field naming policy).
type Seed struct {
	// ID is the unique seed identifier (e.g., SEED-GO-001).
	ID string `yaml:"id"`

	// Pattern is a short description of the pattern.
	Pattern string `yaml:"pattern"`

	// Tier is the starting tier for cold-start inject (always 3 = Tier Rule per plan.md §4.1).
	Tier int `yaml:"tier"`

	// Confidence is the initial confidence score (0.0-1.0).
	Confidence float64 `yaml:"confidence"`

	// Category is the lesson category enum:
	// error-handling | naming | testing | architecture | security | performance | hardcoding | workflow
	Category string `yaml:"category"`

	// Body is the multi-line lesson body (markdown-formatted).
	Body string `yaml:"body"`

	// References is the optional list of source URLs.
	References []string `yaml:"references,omitempty"`

	// Version is the schema version (1 for W3/W4; increment on breaking schema change).
	Version int `yaml:"version"`
}

// Loader is the interface for loading cold-start seeds by project type.
// W3 provides a stub; W4 replaces with marker-based detection + content.
type Loader interface {
	// LoadForProject returns seeds applicable to the given project type.
	// Returns empty slice when projectType is "unknown" or no seeds exist.
	// Never returns an error for "unknown" project type (valid cold-start state).
	LoadForProject(projectType string) ([]Seed, error)
}

// LoaderConfig holds configuration for the default seed loader.
type LoaderConfig struct {
	// SSoTDir is the canonical seed directory (e.g., .claude/skills/moai-meta-harness/seeds/).
	SSoTDir string

	// CacheDir is the project-local seed cache (e.g., .moai/harness/seeds/).
	// Project-local entries take precedence over SSoT for the same seed ID.
	CacheDir string
}

// defaultLoader is the W3 stub loader implementation.
// LoadForProject returns empty slice for all inputs (W4 will populate content).
type defaultLoader struct {
	cfg LoaderConfig
}

// NewLoader creates a Loader using the given LoaderConfig.
// The W3 stub always returns an empty seed list.
func NewLoader(cfg LoaderConfig) Loader {
	return &defaultLoader{cfg: cfg}
}

// LoadForProject returns an empty seed list (W3 stub — W4 adds marker detection + content).
// Returns nil error for "unknown" projectType (valid cold-start state per REQ-HRA-022).
func (l *defaultLoader) LoadForProject(_ string) ([]Seed, error) {
	// W3 stub: no seeds loaded.
	// W4 will implement: scan CacheDir + SSoTDir, merge by ID (cache wins), return.
	return nil, nil
}

// DetectProjectType is a W3 stub returning the literal string "unknown".
// Full marker-based detection (go.mod / package.json / Cargo.toml / etc.)
// is W4 PROJECT-MEGA-001 scope (plan.md §4.2 S10 resolution).
func DetectProjectType() string {
	return "unknown"
}
