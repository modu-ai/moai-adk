package web

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/pkg/models"
)

// projectNestedForm carries the curated nested project-config fields submitted by
// the Console (SPEC-WEB-CONSOLE-007 §E). It is NOT a ProfilePreferences (those are
// profile-store fields; these are project config, HARD-7) and it is distinct from
// the two flat scalars (development_mode / git_convention) that the existing
// writeProjectConfig seam owns.
//
// Each field carries a *Set flag so the write seam can apply the EC-1
// empty=preserve rule per field: an empty/absent submission leaves the persisted
// value unchanged (design.md §D.2). For the two booleans, HTML omits an unchecked
// checkbox entirely, so a hidden companion input (name+"__present") disambiguates
// "unchecked → false" (companion present) from "not submitted → preserve"
// (companion absent). ParseErrs holds per-field type-conversion guard messages
// (e.g. a non-integer coverage target) — these are EC-4-style boundary guards, not
// validation rules.
type projectNestedForm struct {
	CoverageTarget    int
	CoverageTargetSet bool

	EnforceQuality    bool
	EnforceQualitySet bool // companion __present seen → bool was submitted

	MinCoverage    int
	MinCoverageSet bool

	Confidence    float64
	ConfidenceSet bool

	AutoEnabled    bool
	AutoEnabledSet bool // companion __present seen

	SampleSize    int
	SampleSizeSet bool

	EnforceOnPush    bool
	EnforceOnPushSet bool // companion __present seen

	// ParseErrs maps the dot-path field name → a type-conversion guard message.
	ParseErrs map[string]string
}

// touchesQuality reports whether the form carries any quality nested field.
func (f projectNestedForm) touchesQuality() bool {
	return f.CoverageTargetSet || f.EnforceQualitySet || f.MinCoverageSet
}

// touchesGitConvention reports whether the form carries any git_convention nested field.
func (f projectNestedForm) touchesGitConvention() bool {
	return f.ConfidenceSet || f.AutoEnabledSet || f.SampleSizeSet || f.EnforceOnPushSet
}

// parseProjectNestedForm reads the 6 curated nested fields from the POST form via
// explicit per-field PostFormValue calls (NO dynamic reflection path-walker —
// plan.md §I anti-pattern). Empty string submissions are treated as "not
// submitted / preserve" (EC-1); a present-but-non-numeric value records a
// type-conversion guard error in ParseErrs (EC-4-style). For booleans the hidden
// companion input (name+"__present") signals submission; when present, the
// checkbox value="1" presence determines true/false.
func parseProjectNestedForm(r *http.Request) projectNestedForm {
	f := projectNestedForm{ParseErrs: map[string]string{}}

	// quality.test_coverage_target (int)
	if raw := r.PostFormValue("quality.test_coverage_target"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			f.CoverageTarget, f.CoverageTargetSet = n, true
		} else {
			f.ParseErrs["quality.test_coverage_target"] = "must be an integer"
		}
	}

	// quality.enforce_quality (bool via hidden companion)
	if r.PostFormValue("quality.enforce_quality__present") != "" {
		f.EnforceQualitySet = true
		f.EnforceQuality = r.PostFormValue("quality.enforce_quality") != ""
	}

	// quality.tdd_settings.min_coverage_per_commit (int)
	if raw := r.PostFormValue("quality.tdd_settings.min_coverage_per_commit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			f.MinCoverage, f.MinCoverageSet = n, true
		} else {
			f.ParseErrs["quality.tdd_settings.min_coverage_per_commit"] = "must be an integer"
		}
	}

	// git_convention.auto_detection.confidence_threshold (float)
	if raw := r.PostFormValue("git_convention.auto_detection.confidence_threshold"); raw != "" {
		if v, err := strconv.ParseFloat(raw, 64); err == nil {
			f.Confidence, f.ConfidenceSet = v, true
		} else {
			f.ParseErrs["git_convention.auto_detection.confidence_threshold"] = "must be a number"
		}
	}

	// git_convention.auto_detection.enabled (bool via hidden companion)
	if r.PostFormValue("git_convention.auto_detection.enabled__present") != "" {
		f.AutoEnabledSet = true
		f.AutoEnabled = r.PostFormValue("git_convention.auto_detection.enabled") != ""
	}

	// git_convention.auto_detection.sample_size (int)
	if raw := r.PostFormValue("git_convention.auto_detection.sample_size"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			f.SampleSize, f.SampleSizeSet = n, true
		} else {
			f.ParseErrs["git_convention.auto_detection.sample_size"] = "must be an integer"
		}
	}

	// git_convention.validation.enforce_on_push (bool via hidden companion)
	if r.PostFormValue("git_convention.validation.enforce_on_push__present") != "" {
		f.EnforceOnPushSet = true
		f.EnforceOnPush = r.PostFormValue("git_convention.validation.enforce_on_push") != ""
	}

	return f
}

// @MX:WARN: [AUTO] readProjectConfig/writeProjectConfig는 프로필 스토어가 아닌 *프로젝트 설정*
// (.moai/config/sections/quality.yaml + git-convention.yaml)을 디스크에서 읽고 쓰는 두 번째 영속화 경계다
// (SPEC-WEB-CONSOLE-003). handlers.go handleSave의 첫 번째 경계(WritePreferences + SyncToProjectConfig)와는 별개다.
// @MX:REASON: [AUTO] 영속화는 반드시 config.NewConfigManager()/LoadRaw/SetSection/Save 를 통해서만 수행한다 —
// 웹 레이어에서 YAML을 직접 marshal/os.WriteFile 하는 것은 금지된 안티패턴(REQ-WC3-008). scope는 quality(development_mode)
// + git_convention(convention) 두 섹션으로 엄격히 한정되며 workflow/harness/git-strategy/llm은 절대 건드리지 않는다
// (REQ-WC3-007). 비어있는 제출값은 기존 영속값을 덮어쓰지 않는다(empty = "keep existing", EC-1).

// readProjectConfig is the real read seam (REQ-WC3-004). It loads the project
// config via the config manager (LoadRaw — no validation, write-intent path)
// and returns the persisted development_mode + git_convention.convention. An
// absent config dir yields empty values (LoadRaw default behavior, EC-5).
func readProjectConfig(projectRoot string) (devMode, convention string, err error) {
	mgr := config.NewConfigManager()
	cfg, err := mgr.LoadRaw(projectRoot)
	if err != nil {
		return "", "", fmt.Errorf("read project config: %w", err)
	}
	return string(cfg.Quality.DevelopmentMode), cfg.GitConvention.Convention, nil
}

// projectNestedCurrent carries the persisted current values of the curated
// nested fields for GET echo-back (REQ-WC7-010 + REQ-WC9-010). Int/float
// are pre-formatted to strings for the numberField widget value= attribute; the
// bools drive the toggle checked state.
type projectNestedCurrent struct {
	CoverageTarget       string
	EnforceQuality       bool
	MinCoverage          string
	ConfidenceThreshold  string
	AutoDetectionEnabled bool
	SampleSize           string
	EnforceOnPush        bool
}

// readProjectNestedConfig is the read seam for the 6 curated nested fields
// (REQ-WC7-010). It loads via the config manager (LoadRaw — same write-intent path
// as readProjectConfig) and returns the persisted nested values for GET echo-back.
// An absent config dir yields the LoadRaw compiled-in defaults (EC-5), never a
// panic. It is SEPARATE from readProjectConfig so that the existing 2-scalar read
// seam (and its tests) is left byte-unchanged.
func readProjectNestedConfig(projectRoot string) (projectNestedCurrent, error) {
	mgr := config.NewConfigManager()
	cfg, err := mgr.LoadRaw(projectRoot)
	if err != nil {
		return projectNestedCurrent{}, fmt.Errorf("read project nested config: %w", err)
	}
	return projectNestedCurrent{
		CoverageTarget:       strconv.Itoa(cfg.Quality.TestCoverageTarget),
		EnforceQuality:       cfg.Quality.EnforceQuality,
		MinCoverage:          strconv.Itoa(cfg.Quality.TDDSettings.MinCoveragePerCommit),
		ConfidenceThreshold:  strconv.FormatFloat(cfg.GitConvention.AutoDetection.ConfidenceThreshold, 'f', -1, 64),
		AutoDetectionEnabled: cfg.GitConvention.AutoDetection.Enabled,
		SampleSize:           strconv.Itoa(cfg.GitConvention.AutoDetection.SampleSize),
		EnforceOnPush:        cfg.GitConvention.Validation.EnforceOnPush,
	}, nil
}

// writeProjectConfig is the real write seam (REQ-WC3-005/007). It persists each
// non-empty value into its project-config section via the config-manager API
// (LoadRaw → mutate only non-empty → SetSection → Save). Empty submissions leave
// the existing persisted value unchanged (EC-1). It writes ONLY the quality
// (development_mode) and git_convention (convention) sections — Save() round-trips
// every other section's content unchanged. No direct yaml.Marshal/os.WriteFile.
func writeProjectConfig(projectRoot, devMode, convention string) error {
	mgr := config.NewConfigManager()
	cfg, err := mgr.LoadRaw(projectRoot)
	if err != nil {
		return fmt.Errorf("load project config: %w", err)
	}

	changed := false

	if devMode != "" && string(cfg.Quality.DevelopmentMode) != devMode {
		quality := cfg.Quality
		quality.DevelopmentMode = models.DevelopmentMode(devMode)
		if err := mgr.SetSection("quality", quality); err != nil {
			return fmt.Errorf("set quality section: %w", err)
		}
		changed = true
	}

	if convention != "" && cfg.GitConvention.Convention != convention {
		gc := cfg.GitConvention
		gc.Convention = convention
		if err := mgr.SetSection("git_convention", gc); err != nil {
			return fmt.Errorf("set git_convention section: %w", err)
		}
		changed = true
	}

	if changed {
		if err := mgr.Save(); err != nil {
			return fmt.Errorf("save project config: %w", err)
		}
	}
	return nil
}

// writeProjectNestedConfig is the load-modify-write seam for the 6 curated nested
// fields (SPEC-WEB-CONSOLE-007 §D, HARD-4). It is SEPARATE from writeProjectConfig
// so the existing 2-scalar write contract (and its tests) is byte-unchanged.
//
// HARD-4 nested isolation crux: SetSection replaces the WHOLE section struct and
// Save() serializes the whole struct (manager.go), so the seam copies the ENTIRE
// section struct returned by LoadRaw (q := cfg.Quality / gc := cfg.GitConvention)
// and mutates ONLY the targeted nested field(s). Every sibling nested field rides
// through from LoadRaw byte-identical. Each *Set flag gates a per-field mutation so
// an unsubmitted field (EC-1) is left at its persisted value.
//
// This runs AFTER the existing writeProjectConfig in handleSave, so its LoadRaw
// re-reads the scalar values that call already persisted — both writes converge on
// the same on-disk sections without clobbering each other.
func writeProjectNestedConfig(projectRoot string, form projectNestedForm) error {
	mgr := config.NewConfigManager()
	cfg, err := mgr.LoadRaw(projectRoot)
	if err != nil {
		return fmt.Errorf("load project config: %w", err)
	}

	changed := false

	if form.touchesQuality() {
		q := cfg.Quality // whole-struct copy: DDD/TDD/Coverage/LSP/... all ride through
		if form.CoverageTargetSet {
			q.TestCoverageTarget = form.CoverageTarget
		}
		if form.EnforceQualitySet {
			q.EnforceQuality = form.EnforceQuality
		}
		if form.MinCoverageSet {
			q.TDDSettings.MinCoveragePerCommit = form.MinCoverage // nested-of-nested: TDDSettings rides through, one field set
		}
		if err := mgr.SetSection("quality", q); err != nil {
			return fmt.Errorf("set quality section: %w", err)
		}
		changed = true
	}

	if form.touchesGitConvention() {
		gc := cfg.GitConvention // whole-struct copy: AutoDetection/Validation sub-structs all ride through
		if form.ConfidenceSet {
			gc.AutoDetection.ConfidenceThreshold = form.Confidence
		}
		if form.AutoEnabledSet {
			gc.AutoDetection.Enabled = form.AutoEnabled
		}
		if form.SampleSizeSet {
			gc.AutoDetection.SampleSize = form.SampleSize
		}
		if form.EnforceOnPushSet {
			gc.Validation.EnforceOnPush = form.EnforceOnPush
		}
		if err := mgr.SetSection("git_convention", gc); err != nil {
			return fmt.Errorf("set git_convention section: %w", err)
		}
		changed = true
	}

	if changed {
		if err := mgr.Save(); err != nil {
			return fmt.Errorf("save project config: %w", err)
		}
	}
	return nil
}
