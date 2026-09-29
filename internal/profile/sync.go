package profile

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/settings/yamlpatch"
	"github.com/modu-ai/moai-adk/internal/statusline"
	"gopkg.in/yaml.v3"
)

// SyncToProjectConfig synchronizes profile preferences to
// the project's .moai/config/sections/ YAML files.
// Only non-empty preference values overwrite existing config values.
//
// The user name is persisted as a user.yaml ROW REPLACEMENT, not a struct
// re-marshal (SPEC-WEB-SAVE-LOSSLESS-001, AC-WSL-005 — plan-audit F1
// redesign): models.UserConfig models only `name`, so any struct round-trip
// loses unmodeled keys (github_username, timezone, ...) and comments at the
// re-marshal point. The yamlpatch splice rewrites only the `name:` row and
// leaves every other byte of user.yaml intact. A name-absent user.yaml takes
// the upsert fallback (AC-WSL-005 F6 variant — C3 data-level guarantee).
//
// Language preferences are persisted the same way as language.yaml ROW
// REPLACEMENTS (SPEC-WEB-SAVE-LOSSLESS-001 — sync-audit F-1): the former
// SetSection("language") + Save() path was the FIFTH residual re-marshal —
// the web console's four language selects reach it, and a re-marshal wipes
// the file's comments, unmodeled keys, and re-injects modeled defaults
// (error_messages) absent from the fixture. Only the rows whose value the
// submission changes are spliced; every other byte of language.yaml survives.
//
// The yamlpatch calls go DIRECTLY to internal/settings/yamlpatch rather than
// through settings.WriteSectionViaSeam because internal/settings already
// imports internal/profile (the shared field schema) — importing it back here
// would be an import cycle. The seam conventions (yamlpatch node surgery,
// atomic temp+rename write) are the same either way.
func SyncToProjectConfig(projectRoot string, prefs ProfilePreferences) error {
	mgr := config.NewConfigManager()
	cfg, err := mgr.LoadRaw(projectRoot)
	if err != nil {
		return fmt.Errorf("load project config: %w", err)
	}

	// Sync user section — name row splice only (AC-WSL-005). The splice is the
	// ONLY write for a name-only sync: no SetSection, no Save().
	if prefs.UserName != "" && cfg.User.Name != prefs.UserName {
		sectionsDir := filepath.Join(projectRoot, ".moai", "config", "sections")
		if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
			return fmt.Errorf("create config directory: %w", err)
		}
		if err := yamlpatch.PatchFile(filepath.Join(sectionsDir, "user.yaml"),
			[]yamlpatch.KeyEdit{{Path: []string{"user", "name"}, Value: prefs.UserName}}); err != nil {
			return fmt.Errorf("sync user name: %w", err)
		}
	}

	// Sync language section — row splice only (sync-audit F-1). Change
	// detection reads the loaded config; the edits splice ONLY the rows the
	// submission changes, so comments, unmodeled keys, and untouched scalars
	// survive byte-identical. No SetSection, no Save().
	lang := cfg.Language
	var langEdits []yamlpatch.KeyEdit

	if prefs.ConversationLang != "" && lang.ConversationLanguage != prefs.ConversationLang {
		langEdits = append(langEdits,
			yamlpatch.KeyEdit{Path: []string{"language", "conversation_language"}, Value: prefs.ConversationLang},
			yamlpatch.KeyEdit{Path: []string{"language", "conversation_language_name"}, Value: prefs.ConversationLang},
		)
	}
	if prefs.GitCommitLang != "" && lang.GitCommitMessages != prefs.GitCommitLang {
		langEdits = append(langEdits, yamlpatch.KeyEdit{Path: []string{"language", "git_commit_messages"}, Value: prefs.GitCommitLang})
	}
	if prefs.CodeCommentLang != "" && lang.CodeComments != prefs.CodeCommentLang {
		langEdits = append(langEdits, yamlpatch.KeyEdit{Path: []string{"language", "code_comments"}, Value: prefs.CodeCommentLang})
	}
	if prefs.DocLang != "" && lang.Documentation != prefs.DocLang {
		langEdits = append(langEdits, yamlpatch.KeyEdit{Path: []string{"language", "documentation"}, Value: prefs.DocLang})
	}

	if len(langEdits) > 0 {
		sectionsDir := filepath.Join(projectRoot, ".moai", "config", "sections")
		if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
			return fmt.Errorf("create config directory: %w", err)
		}
		if err := yamlpatch.PatchFile(filepath.Join(sectionsDir, "language.yaml"), langEdits); err != nil {
			return fmt.Errorf("sync language section: %w", err)
		}
	}

	// Sync statusline section (written directly to avoid config manager dependency)
	if prefs.StatuslineTheme != "" || prefs.StatuslineSegments != nil {
		if err := syncStatusline(projectRoot, prefs); err != nil {
			return fmt.Errorf("sync statusline: %w", err)
		}
	}

	return nil
}

// statuslineData is the internal YAML structure for statusline.yaml. It mirrors
// the canonical models.StatuslineConfig shape {Segments, Theme}. The `mode:`
// surface was removed (SLM-1/SLM-2/SLR-2 — inert) and the `preset:` surface was
// retired (SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001). A legacy `preset:` key in
// an existing statusline.yaml is silently ignored (unknown YAML keys do not
// error on unmarshal).
type statuslineData struct {
	Segments map[string]bool `yaml:"segments,omitempty"`
	Theme    string          `yaml:"theme,omitempty"`
}

// statuslineFileWrap wraps statuslineData under the "statusline" top-level key.
type statuslineFileWrap struct {
	Statusline statuslineData `yaml:"statusline"`
}

// syncStatusline writes StatuslineSegments and StatuslineTheme to
// .moai/config/sections/statusline.yaml (SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001
// retired the preset shorthand — segments + theme are the only levers now).
// When the file is absent, all segments default to enabled (REQ-SLE-022).
// Submitted segments (prefs.StatuslineSegments != nil) are persisted verbatim;
// when not submitted, existing on-disk segments are preserved untouched so
// theme-only saves do not clobber a user's segment map.
func syncStatusline(projectRoot string, prefs ProfilePreferences) error {
	sectionsDir := filepath.Join(projectRoot, ".moai", "config", "sections")
	statuslineFile := filepath.Join(sectionsDir, "statusline.yaml")

	// Read current statusline.yaml if it exists
	var current statuslineFileWrap
	data, err := os.ReadFile(statuslineFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read statusline.yaml: %w", err)
	}
	if err == nil {
		if err := yaml.Unmarshal(data, &current); err != nil {
			return fmt.Errorf("parse statusline.yaml: %w", err)
		}
	}

	// Apply default segments when statusline.yaml was absent (REQ-SLE-022 /
	// REQ-SPR-009): the only default path is all-15-enabled; no preset
	// participates in defaulting.
	if current.Statusline.Segments == nil {
		current.Statusline.Segments = defaultStatuslineSegments()
	}

	// Submitted segments win verbatim (REQ-SPR-008). When not submitted, existing
	// on-disk segments are preserved untouched (theme-only / segment-only saves).
	if prefs.StatuslineSegments != nil {
		current.Statusline.Segments = prefs.StatuslineSegments
	}
	if prefs.StatuslineTheme != "" {
		current.Statusline.Theme = prefs.StatuslineTheme
	}

	// Write statusline.yaml
	yamlData, err := yaml.Marshal(current)
	if err != nil {
		return fmt.Errorf("marshal statusline.yaml: %w", err)
	}
	if err := os.MkdirAll(sectionsDir, 0o700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	if err := os.WriteFile(statuslineFile, yamlData, 0o644); err != nil {
		return fmt.Errorf("write statusline.yaml: %w", err)
	}
	return nil
}

// defaultStatuslineSegments returns the canonical 16-key segment map with every
// segment enabled — equivalent to the "full" preset (SLM-5 fix). The keys are
// sourced from the statusline.Segment* constants (the same SSOT the CLI's
// presetToSegments uses) so the seed never drifts from the canonical schema.
// SegmentRepo is intentionally excluded — it is the 17th constant, outside the
// 16-key statusline schema (SLM-7).
func defaultStatuslineSegments() map[string]bool {
	keys := []string{
		statusline.SegmentModel,
		statusline.SegmentContext,
		statusline.SegmentOutputStyle,
		statusline.SegmentClaudeVersion,
		statusline.SegmentMoaiVersion,
		statusline.SegmentSessionTime,
		statusline.SegmentEffortThinking,
		statusline.SegmentCacheHit,
		statusline.SegmentUsage5H,
		statusline.SegmentUsage7D,
		statusline.SegmentDirectory,
		statusline.SegmentGitStatus,
		statusline.SegmentGitBranch,
		statusline.SegmentWorktree,
		statusline.SegmentTask,
		statusline.SegmentPR,
	}
	segments := make(map[string]bool, len(keys))
	for _, k := range keys {
		segments[k] = true
	}
	return segments
}
