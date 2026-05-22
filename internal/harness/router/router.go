// Package router provides routing logic that determines the harness level
// (minimal/standard/thorough) based on SPEC complexity signals.
// REQ-HRN-001-003: HarnessRouter.Route(spec, cfg) -> (Level, Rationale, error).
package router

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/spec"
)

// Level is the harness routing level type.
// @MX:ANCHOR: [AUTO] FROZEN level enum — {minimal, standard, thorough} (REQ-HRN-001-017)
// @MX:REASON: FROZEN per SPEC-V3R2-HRN-001 REQ-017; multiple consumers (CLI, router, effort.go, escalation.go)
type Level string

const (
	// LevelMinimal is the minimal harness for simple changes.
	LevelMinimal Level = "minimal"
	// LevelStandard is the standard harness for ordinary development.
	LevelStandard Level = "standard"
	// LevelThorough is the highest-level harness for critical functionality.
	LevelThorough Level = "thorough"
)

// Rationale is the struct that captures the routing decision rationale.
// REQ-HRN-001-003: the second return value of Route().
type Rationale struct {
	// MatchedRule is the name of the rule that produced the routing decision.
	// Examples: "spec_override", "force_thorough", "auto_minimal", "auto_standard", "fallthrough_default".
	MatchedRule string `json:"matched_rule"`
	// FileCount is the estimated number of related files.
	FileCount int `json:"file_count"`
	// DomainCount is the estimated number of related domains.
	DomainCount int `json:"domain_count"`
	// SpecType is the estimated SPEC type.
	// Examples: "bugfix", "docs", "config", "feature", "refactor", "other".
	SpecType string `json:"spec_type"`
	// SpecPriority is the SPEC priority (frontmatter priority field).
	SpecPriority string `json:"spec_priority"`
	// Keywords is the list of matched keywords that triggered force-thorough.
	Keywords []string `json:"keywords"`
}

// SPECInput is the SPEC input struct consumed by the router.
// Used to extract complexity signals from spec.SPECFrontmatter.
type SPECInput struct {
	// Priority is the frontmatter priority field (for example "P0", "P1 Critical").
	Priority string
	// Tags is the frontmatter tags field (comma-separated).
	Tags string
	// Title is the frontmatter title field.
	Title string
	// Body is the SPEC document body (including the Requirements section).
	Body string
	// HarnessLevel is the frontmatter harness_level field (optional override).
	// REQ-HRN-001-015.
	HarnessLevel string
}

// Router is the interface that determines the harness level based on SPEC complexity signals.
// @MX:ANCHOR: [AUTO] harness routing interface — Route, RouteFromFile methods
// @MX:REASON: fan_in >= 3: many consumers including the CLI route command, tests, and ConfigManager integration
type Router interface {
	// Route returns the level and rationale based on a SPECInput and HarnessConfig.
	Route(doc *SPECInput, cfg *config.HarnessConfig) (Level, Rationale, error)
	// RouteFromFile performs routing directly from a SPEC file path.
	RouteFromFile(specPath string, cfg *config.HarnessConfig) (Level, Rationale, error)
}

// defaultRouter is the default implementation of the Router interface.
type defaultRouter struct {
	cfg *config.HarnessConfig
}

// New returns a new Router instance.
func New(cfg *config.HarnessConfig) Router {
	return &defaultRouter{cfg: cfg}
}

// Route determines the harness level based on a SPECInput and HarnessConfig.
// Priority order (REQ-HRN-001-003/007/008/015):
//  1. SPEC frontmatter harness_level: override (REQ-015) — highest priority.
//  2. force-thorough override (REQ-008) — security/payment keywords, critical priority.
//  3. Escalation (REQ-009) — accumulative (the router only determines the initial value).
//  4. auto_detection rules (REQ-007) — minimal -> standard -> thorough.
//  5. mode_defaults (REQ-014) — lowest-priority fallback.
func (r *defaultRouter) Route(doc *SPECInput, cfg *config.HarnessConfig) (Level, Rationale, error) {
	signals := ExtractSignals(doc)

	rationale := Rationale{
		FileCount:    signals.FileCount,
		DomainCount:  signals.DomainCount,
		SpecType:     signals.SpecType,
		SpecPriority: doc.Priority,
		Keywords:     []string{},
	}

	// 1. SPEC frontmatter harness_level: override (REQ-HRN-001-015).
	if doc.HarnessLevel != "" {
		switch Level(doc.HarnessLevel) {
		case LevelMinimal, LevelStandard, LevelThorough:
			rationale.MatchedRule = "spec_override"
			return Level(doc.HarnessLevel), rationale, nil
		}
		// Ignore unknown harness_level values and continue.
	}

	// 2. force-thorough override (REQ-HRN-001-008).
	forcedKeywords := matchForceThoroughKeywords(doc)
	isCritical := isCriticalPriority(doc.Priority)
	isSensitiveDomain := isSensitiveTagDomain(doc.Tags)

	if len(forcedKeywords) > 0 || isCritical || isSensitiveDomain {
		rationale.MatchedRule = "force_thorough"
		rationale.Keywords = forcedKeywords
		return LevelThorough, rationale, nil
	}

	// 3. auto_detection rules (REQ-HRN-001-007): order is minimal -> standard -> thorough.
	level, rule := applyAutoDetectionRules(signals, cfg)
	rationale.MatchedRule = rule
	return level, rationale, nil
}

// RouteFromFile performs routing directly from a SPEC file path.
func (r *defaultRouter) RouteFromFile(specPath string, cfg *config.HarnessConfig) (Level, Rationale, error) {
	doc, err := loadSPECDoc(specPath)
	if err != nil {
		return LevelStandard, Rationale{MatchedRule: "fallthrough_default"}, fmt.Errorf("RouteFromFile: %w", err)
	}
	return r.Route(doc, cfg)
}

// loadSPECDoc loads a SPECInput from a file path.
func loadSPECDoc(specPath string) (*SPECInput, error) {
	data, err := os.ReadFile(specPath)
	if err != nil {
		return nil, fmt.Errorf("read spec file: %w", err)
	}

	// Use the parser from the spec package.
	content := string(data)
	fm, body, parseErr := extractSPECFrontmatter(content)
	if parseErr != nil {
		// On parse error, proceed with defaults.
		return &SPECInput{
			Title: filepath.Base(specPath),
			Body:  content,
		}, nil
	}

	return &SPECInput{
		Priority:     fm.Priority,
		Tags:         fm.Tags,
		Title:        fm.Title,
		Body:         body,
		HarnessLevel: fm.HarnessLevel,
	}, nil
}

// extractSPECFrontmatter parses a spec.SPECFrontmatter.
// Indirectly uses parseSPECDoc from internal/spec/lint.go.
func extractSPECFrontmatter(content string) (spec.SPECFrontmatter, string, error) {
	return spec.ExtractFrontmatter(content)
}

// applyAutoDetectionRules determines the level based on complexity signals.
// REQ-HRN-001-007: minimal -> standard -> thorough priority order.
func applyAutoDetectionRules(signals ComplexitySignals, cfg *config.HarnessConfig) (Level, string) {
	// minimal condition: file_count <= 3 AND single_domain AND spec_type in [bugfix, docs, config].
	isMinimalSpecType := signals.SpecType == "bugfix" || signals.SpecType == "docs" || signals.SpecType == "config"
	if signals.FileCount <= 3 && signals.DomainCount <= 1 && isMinimalSpecType {
		return LevelMinimal, "auto_minimal"
	}

	// standard condition: file_count > 3 OR multi_domain OR spec_type in [feature, refactor].
	isStandardSpecType := signals.SpecType == "feature" || signals.SpecType == "refactor"
	if signals.FileCount > 3 || signals.DomainCount > 1 || isStandardSpecType {
		return LevelStandard, "auto_standard"
	}

	// Fallback: standard (REQ-HRN-001-007 defines the fallback as standard).
	return LevelStandard, "fallthrough_default"
}

// isCriticalPriority checks whether the priority string denotes Critical.
// REQ-HRN-001-008: spec_priority == critical -> force thorough.
// Note: P0 is equivalent to Critical, but a P0 value alone does not trigger force_thorough.
// Triggering requires the explicit "critical" keyword, as in "P0 Critical" or "Critical".
func isCriticalPriority(priority string) bool {
	lower := strings.ToLower(priority)
	return strings.Contains(lower, "critical")
}

// isSensitiveTagDomain checks whether the tags contain a sensitive domain.
// REQ-HRN-001-008: domain in [auth, payment, migration, public_api] -> force thorough.
func isSensitiveTagDomain(tags string) bool {
	lower := strings.ToLower(tags)
	sensitiveDomains := []string{"auth", "payment", "migration", "public_api"}
	for _, domain := range sensitiveDomains {
		if strings.Contains(lower, domain) {
			return true
		}
	}
	return false
}

// ConfigProxy is a lightweight config wrapper for EffortForLevelFromProxy.
// Used by tests.
type ConfigProxy struct {
	EffortMapping map[string]string
}
