package web

import (
	"errors"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/profile"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/pkg/models"
)

// errDictKey is returned by the template "dict" helper when a non-string key is
// supplied. Defined here so assets.go stays focused on embedding.
var errDictKey = errors.New("web: dict key must be a string")

// Canonical value lists.
//
// @MX:NOTE: [AUTO] 이 목록들은 internal/cli/profile_setup.go 의 wizard SSOT(statuslineModeCanonical /
// statuslineThemeCanonical / statuslinePresetCanonical / 언어 옵션)와 동일한 정규 값이다. wizard의 정규 리스트가
// 미노출(unexported, internal/cli 패키지 전용)이고 internal/cli → internal/web 단방향 의존이므로 역참조가 불가능하여
// 같은 값을 여기서 재선언한다. permission mode 검증은 별도 규칙을 만들지 않고 profile.IsValidPermissionMode 를 그대로 재사용한다
// (병렬 검증 규칙 셋 금지 — REQ-WC-008). wizard 정규 리스트 변경 시 본 목록도 함께 갱신해야 한다.

// langOptions are the four supported conversation/commit/comment/doc languages.
// Mirrors the en/ko/ja/zh options offered by the profile wizard.
var langOptions = []string{"en", "ko", "ja", "zh"}

// modelCanonical mirrors the wizard SSOT model Select option set
// (internal/cli/profile_setup.go:303-310). The empty string ("project default")
// is allowed by validatePrefs' empty-allowed guard and is not listed here.
// Mirrored (not imported) because the wizard option strings are unexported and
// internal/cli → internal/web is the only legal import direction.
var modelCanonical = []string{"opus", "opus[1m]", "sonnet", "sonnet[1m]", "haiku", "opusplan"}

// effortLevelCanonical mirrors the wizard SSOT effort Select option set
// (internal/cli/profile_setup.go:316-322). The empty string ("runtime default")
// is allowed by the empty-allowed guard and is not listed here.
var effortLevelCanonical = []string{"low", "medium", "high", "xhigh", "max"}

// statuslineModeCanonical mirrors internal/cli/profile_setup.go statuslineModeCanonical.
var statuslineModeCanonical = []string{"default", "full"}

// statuslineThemeCanonical mirrors internal/cli/profile_setup.go statuslineThemeCanonical.
var statuslineThemeCanonical = []string{"catppuccin-mocha", "catppuccin-latte"}

// statuslinePresetCanonical mirrors internal/cli/profile_setup.go statuslinePresetCanonical.
var statuslinePresetCanonical = []string{"full", "compact", "minimal", "custom"}

// allSegments mirrors internal/cli/profile_setup.go statuslineAllSegments — the
// 15 canonical segment keys offered when the preset is "custom".
var allSegments = []string{
	"claude_version", "context", "directory", "effort_thinking",
	"git_branch", "git_status", "moai_version", "model",
	"output_style", "pr", "session_time", "task",
	"usage_5h", "usage_7d", "worktree",
}

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
		if val != "" && !inList(langOptions, val) {
			errs[field] = "unrecognized language: " + val
		}
	}

	// Model / effort_level / model_policy: empty allowed, otherwise canonical.
	// REQ-WC2-002/003/004 — web↔TUI validation parity. model + effort_level reuse
	// the mirrored wizard lists; model_policy wires in the existing exported
	// predicate template.IsValidModelPolicy (no mirror — it is importable).
	if p.Model != "" && !inList(modelCanonical, p.Model) {
		errs["model"] = "unrecognized model: " + p.Model
	}
	if p.EffortLevel != "" && !inList(effortLevelCanonical, p.EffortLevel) {
		errs["effort_level"] = "unrecognized effort level: " + p.EffortLevel
	}
	if p.ModelPolicy != "" && !template.IsValidModelPolicy(p.ModelPolicy) {
		errs["model_policy"] = "unrecognized model policy: " + p.ModelPolicy
	}

	// Statusline mode / preset / theme: empty allowed, otherwise canonical.
	if p.StatuslineMode != "" && !inList(statuslineModeCanonical, p.StatuslineMode) {
		errs["statusline_mode"] = "unrecognized statusline mode: " + p.StatuslineMode
	}
	if p.StatuslinePreset != "" && !inList(statuslinePresetCanonical, p.StatuslinePreset) {
		errs["statusline_preset"] = "unrecognized statusline preset: " + p.StatuslinePreset
	}
	if p.StatuslineTheme != "" && !inList(statuslineThemeCanonical, p.StatuslineTheme) {
		errs["statusline_theme"] = "unrecognized statusline theme: " + p.StatuslineTheme
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
