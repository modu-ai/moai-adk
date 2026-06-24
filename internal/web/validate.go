package web

import (
	"errors"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/profile"
	"github.com/modu-ai/moai-adk/internal/settings"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/pkg/models"
)

// errDictKey is returned by the template "dict" helper when a non-string key is
// supplied. Defined here so assets.go stays focused on embedding.
var errDictKey = errors.New("web: dict key must be a string")

// Canonical value lists — SINGLE-SOURCED from the settings schema.
//
// @MX:NOTE: [AUTO] SPEC-WEB-CONSOLE-010: 기존에 손수-미러되던 5개 정규 목록
// (langOptions / modelCanonical / effortLevelCanonical / developmentModeCanonical /
// conventionCanonical)은 모두 제거되고 중립 internal/settings 스키마에서 파생된다.
// internal/cli → internal/web 단방향 의존 때문에 wizard 정규 리스트를 역참조할 수 없어
// 손수 재선언했던 드리프트 원천이 제3 중립 패키지(internal/settings)로 단일화되었다.
// 두 표면(웹/TUI)이 모두 internal/settings 를 import 하므로 더 이상 손수 동기화가 필요 없다.
// permission mode 검증은 profile.IsValidPermissionMode 를 그대로 재사용한다(REQ-WC-008).

// langOptionList returns the four supported conversation/commit/comment/doc
// languages from the schema (REQ-WC10-004). No standalone re-declaration.
func langOptionList() []string { return settings.LanguageOptionValues() }

// modelOptionList returns the canonical model option set from the schema
// (REQ-WC10-004). The empty string ("project default") is allowed by validatePrefs'
// empty-allowed guard and is not listed here.
func modelOptionList() []string { return settings.ModelOptionValues() }

// effortOptionList returns the canonical effort level option set from the schema
// (REQ-WC10-004). The empty string ("runtime default") is allowed by the
// empty-allowed guard and is not listed here.
func effortOptionList() []string { return settings.EffortOptionValues() }

// developmentModeOptionList returns the canonical development_mode option list from
// the schema (REQ-WC10-004).
func developmentModeOptionList() []string { return settings.DevelopmentModeOptionValues() }

// conventionOptionList returns the canonical git_convention option list from the
// schema (REQ-WC10-004).
func conventionOptionList() []string { return settings.ConventionOptionValues() }

// inList reports whether v is a member of list.
func inList(list []string, v string) bool {
	for _, item := range list {
		if item == v {
			return true
		}
	}
	return false
}

// validatePrefs validates a submitted ProfilePreferences against the existing
// canonical value lists and predicates. It returns a map of field name → error
// message for every field that fails. An empty map means all values are valid.
//
// REQ-WC-008: validation reuses the existing predicates
// (profile.IsValidPermissionMode + canonical-list membership) rather than a
// parallel rule set. Empty values are always allowed (they mean "unset / keep
// project default") — mirroring SyncToProjectConfig which only overwrites
// non-empty values.
func validatePrefs(p profile.ProfilePreferences) map[string]string {
	errs := make(map[string]string)

	// Permission mode: reuse the existing exported predicate.
	if !profile.IsValidPermissionMode(p.PermissionMode) {
		errs["permission_mode"] = "unrecognized permission mode: " + p.PermissionMode
	}

	// Language fields: empty allowed, otherwise must be a canonical language.
	for field, val := range map[string]string{
		"conversation_lang": p.ConversationLang,
		"git_commit_lang":   p.GitCommitLang,
		"code_comment_lang": p.CodeCommentLang,
		"doc_lang":          p.DocLang,
	} {
		if val != "" && !inList(langOptionList(), val) {
			errs[field] = "unrecognized language: " + val
		}
	}

	// Model / effort_level / model_policy: empty allowed, otherwise canonical.
	// REQ-WC2-002/003/004 — web↔TUI validation parity. SPEC-WEB-CONSOLE-010: model +
	// effort_level now derive from the shared settings schema (no hand-mirrored list);
	// model_policy continues to reuse the exported predicate template.IsValidModelPolicy.
	if p.Model != "" && !inList(modelOptionList(), p.Model) {
		errs["model"] = "unrecognized model: " + p.Model
	}
	if p.EffortLevel != "" && !inList(effortOptionList(), p.EffortLevel) {
		errs["effort_level"] = "unrecognized effort level: " + p.EffortLevel
	}
	if p.ModelPolicy != "" && !template.IsValidModelPolicy(p.ModelPolicy) {
		errs["model_policy"] = "unrecognized model policy: " + p.ModelPolicy
	}

	// Statusline theme: SPEC-WEB-CONSOLE-010 re-added the Statusline section. Empty
	// allowed (theme-only-unset save), otherwise must be a canonical theme value
	// from the shared schema. The retired `preset` field is NOT validated
	// (REQ-WC10-010 — no preset control exists). Segment values are booleans bound
	// in bindForm and carry no enum to validate.
	if p.StatuslineTheme != "" {
		if f, ok := settings.Field("statusline_theme"); ok {
			if !inList(f.SelectOptions(), p.StatuslineTheme) {
				errs["statusline_theme"] = "unrecognized statusline theme: " + p.StatuslineTheme
			}
		}
	}

	return errs
}

// validateProjectConfig validates the two flat project-config enum fields
// (development_mode + git_convention.convention) submitted via the Console.
// It returns a map of field name → error message for every field that fails;
// an empty map means all values are valid (REQ-WC3-001/002).
//
// It is a SEPARATE validator from validatePrefs because those values are NOT
// ProfilePreferences fields — they live in project config (quality.yaml /
// git-convention.yaml), not the profile store. It reuses the canonical predicates
// (models.DevelopmentMode.IsValid + config.IsValidConvention) rather than a
// parallel rule set (REQ-WC3-001/002 anti-pattern E.5.1). Empty values are always
// allowed (they mean "keep project default") — mirroring writeProjectConfig which
// only overwrites non-empty values. Matching is exact (no lowercase normalization),
// so "TDD" / "Angular" / a whitespace value is non-canonical (EC-4, EC-6).
func validateProjectConfig(devMode, convention string) map[string]string {
	errs := make(map[string]string)

	if devMode != "" && !models.DevelopmentMode(devMode).IsValid() {
		errs["development_mode"] = "unrecognized development mode: " + devMode
	}
	if convention != "" && !config.IsValidConvention(convention) {
		errs["git_convention"] = "unrecognized git convention: " + convention
	}

	return errs
}

// validateProjectNestedConfig validates the 6 curated nested project-config fields
// (SPEC-WEB-CONSOLE-007 §B). It returns a map of dot-path field name → error
// message; an empty map means all submitted nested values are valid.
//
// It is SEPARATE from validateProjectConfig (the 2-scalar validator) and reuses the
// config-package export seam (config.ValidateQualitySection /
// config.ValidateGitConventionSection) rather than authoring a parallel rule-set —
// CRITICAL SCOPE CONSTRAINT, REQ-WC7-002/008. The flow is: surface form
// type-conversion guard errors first (ParseErrs, EC-4-style), then build the two
// section structs from the *submitted* nested deltas (only set fields are applied)
// and run them through the export seam. For the custom-required cross-field rule
// the submitted git_convention scalar is threaded in so the seam can decide whether
// custom.pattern is required. The validator never writes — handleSave runs the
// write seams only after this returns an empty map (EC-2 atomic reject).
func validateProjectNestedConfig(convention string, form projectNestedForm) map[string]string {
	errs := make(map[string]string)

	// Type-conversion guards (EC-4-style): a present-but-non-numeric value.
	for k, v := range form.ParseErrs {
		errs[k] = v
	}

	// Build the quality section struct from the submitted nested deltas and run the
	// export seam (reuses the existing 0-100 range rules — no new rule).
	if form.touchesQuality() {
		q := &models.QualityConfig{}
		if form.CoverageTargetSet {
			q.TestCoverageTarget = form.CoverageTarget
		}
		if form.MinCoverageSet {
			q.TDDSettings.MinCoveragePerCommit = form.MinCoverage
		}
		for _, e := range config.ValidateQualitySection(q) {
			errs[e.Field] = e.Message
		}
	}

	// Build the git_convention section struct from the submitted nested deltas and
	// run the export seam (reuses the existing confidence_threshold [0.0,1.0] +
	// sample_size>=0 range rules — no new rule). The `custom` engine has been
	// removed, so the convention scalar is validated against the 4-value enum via
	// the seam; there is no custom-required cross-field rule.
	if form.touchesGitConvention() || convention != "" {
		gc := &models.GitConventionConfig{Convention: convention}
		if form.ConfidenceSet {
			gc.AutoDetection.ConfidenceThreshold = form.Confidence
		}
		if form.SampleSizeSet {
			gc.AutoDetection.SampleSize = form.SampleSize
		}
		for _, e := range config.ValidateGitConventionSection(gc) {
			errs[e.Field] = e.Message
		}
	}

	return errs
}
