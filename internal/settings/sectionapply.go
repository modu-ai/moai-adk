package settings

// 이 파일은 M2b 확장 필드의 영속화 디스패처다 (SPEC-WEB-CONSOLE-011 M2b).
//
// @MX:WARN: [AUTO] ApplySchemaEdits는 10섹션 확장 필드를 디스크에 쓰는 영속화
// 경계다. seam 섹션은 WriteSectionViaSeam(yamlpatch — 주석/미모델링 키 보존),
// typed 섹션(git_strategy/llm/quality)은 config.NewConfigManager LoadRaw →
// per-field apply → SetSection → Save 경로만 사용한다.
// @MX:REASON: [AUTO] workflow.yaml 등 8개 seam 섹션에 typed re-marshal을 적용하면
// 주석과 미모델링 키가 파괴된다 (REQ-WC11-005/017, AP-1/AP-11). 반대로
// git-strategy는 완전 typed + dirty-flag Save가 요구 계약이다 (REQ-WC11-010,
// SPEC-GITSTRATEGY-SAVE-ISOLATION-001). 필드별 라우팅은 FieldDef.Persist.Kind가
// SSOT이며, 여기 없는 키(read-only: llm.mode/team_mode, db system 5키)는 어떤
// 경로로도 기록되지 않는다 (REQ-WC11-013/019).

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/settings/yamlpatch"
	"github.com/modu-ai/moai-adk/pkg/models"
	"gopkg.in/yaml.v3"
)

// ApplySchemaEdits는 제출된 확장 필드 값(FieldDef.Name → 문자열 값)을 영속화한다.
// 빈 문자열 값은 호출자(웹 파서)가 걸러낸다(empty=preserve, EC-1) — 단
// EmptySubmits 옵트인 필드는 예외다: 그 필드의 ""는 명시 제출 값이며, seam 경로는
// 키를 중립 ""(예: crosssession.inbound)로 기록한다. 알 수 없는 필드명은 오류다.
func ApplySchemaEdits(projectRoot string, edits map[string]string) error {
	if len(edits) == 0 {
		return nil
	}

	// 결정적 순서로 처리 (맵 순회 비결정성 제거).
	names := make([]string, 0, len(edits))
	for name := range edits {
		names = append(names, name)
	}
	sort.Strings(names)

	seamEdits := map[string][]yamlpatch.KeyEdit{} // 섹션 파일 → edits
	typedEdits := make([]FieldDef, 0)
	typedValues := make([]string, 0)

	for _, name := range names {
		f, ok := Field(name)
		if !ok {
			return fmt.Errorf("settings: unknown schema field %q", name)
		}
		switch f.Persist.Kind {
		case PersistSeam:
			// REQ-WWS-003 (SPEC-WEB-WRITE-SAFETY-001): a seam edit that would
			// not change the persisted value must not reach the file. Two
			// value-invariant shapes exist: (a) the submitted value equals the
			// persisted scalar, and (b) the key is ABSENT and a bool field
			// submits its EFFECTIVE default — absent means whatever the key's
			// runtime interpreter makes it, so only a submission equal to that
			// interpretation is a no-op (sync-audit F1: a default-ON key such
			// as workflow.todo.enabled or the fail-open mcp.tools.*.enabled
			// reads enabled when absent, so an explicit OFF there IS a real
			// change and must be written; skipping it had silently swallowed
			// the user's save).
			cur, curOk := readSeamScalar(projectRoot, f.Persist.Section, f.Persist.Path)
			if curOk && cur == edits[name] {
				continue
			}
			if !curOk && f.Type == TypeBool {
				effective := f.AbsentDefault
				if effective == "" {
					effective = "false"
				}
				if edits[name] == effective {
					continue
				}
			}
			seamEdits[f.Persist.Section] = append(seamEdits[f.Persist.Section],
				yamlpatch.KeyEdit{Path: f.Persist.Path, Value: edits[name]})
		case PersistTypedSection:
			typedEdits = append(typedEdits, f)
			typedValues = append(typedValues, edits[name])
		default:
			return fmt.Errorf("settings: field %q is not a schema-section field (kind %s)", name, f.Persist.Kind)
		}
	}

	// typed 섹션: 단일 LoadRaw → 전 필드 적용 → 실변경 섹션만 SetSection → 단일 Save.
	if len(typedEdits) > 0 {
		if err := applyTypedEdits(projectRoot, typedEdits, typedValues); err != nil {
			return err
		}
	}

	// seam 섹션: 파일별 단일 PatchFile (원자적 기록).
	seamSections := make([]string, 0, len(seamEdits))
	for sec := range seamEdits {
		seamSections = append(seamSections, sec)
	}
	sort.Strings(seamSections)
	for _, sec := range seamSections {
		if err := WriteSectionViaSeam(projectRoot, sec, seamEdits[sec]); err != nil {
			return err
		}
	}
	return nil
}

// applyTypedEdits는 typed 섹션 필드를 config 매니저 경로로 영속화한다.
// git_strategy 필드가 하나라도 있으면 SetSection("git_strategy")이 dirty-flag를
// 세워 Save가 git-strategy.yaml을 재기록한다 (그 외에는 기존 파일 byte 보존 —
// SPEC-GITSTRATEGY-SAVE-ISOLATION-001).
//
// REQ-WWS-003/004 (SPEC-WEB-WRITE-SAFETY-001): SetSection은 apply 전후 구조체
// 비교에서 실제로 값이 바뀐 섹션에만 호출된다. M1(d)에서 값-불변 제출이
// SetSection("git_strategy")을 통과해 gitStrategyDirty를 세우고 Save가
// git-strategy.yaml을 전체 재마샬하는 것(키 재배열 + zero-value 키 추가)이
// 관측됐다 — 값이 바뀌지 않은 섹션은 dirty 플래그에 도달해서는 안 된다.
func applyTypedEdits(projectRoot string, fields []FieldDef, values []string) error {
	mgr := config.NewConfigManager()
	cfg, err := mgr.LoadRaw(projectRoot)
	if err != nil {
		return fmt.Errorf("settings: load project config: %w", err)
	}

	before := map[string]any{
		"git_strategy": cfg.GitStrategy,
		"llm":          cfg.LLM,
		"quality":      cfg.Quality,
	}

	touched := map[string]bool{}
	for i, f := range fields {
		switch f.Persist.Section {
		case "git_strategy":
			if err := applyGitStrategyKey(&cfg.GitStrategy, f.Persist.Key, values[i]); err != nil {
				return err
			}
		case "llm":
			if err := applyLLMKey(&cfg.LLM, f.Persist.Key, values[i]); err != nil {
				return err
			}
		case "quality":
			if err := applyQualityKey(&cfg.Quality, f.Persist.Key, values[i]); err != nil {
				return err
			}
		default:
			return fmt.Errorf("settings: no typed applier for section %q", f.Persist.Section)
		}
		touched[f.Persist.Section] = true
	}

	// quality_extras_enabled forced true: the launch-tab toggle was removed from
	// the web UI, so the field is never submitted from the form anymore. To keep
	// the persisted value aligned with the always-on intent (and to migrate any
	// pre-removal config that had it stored as false), force the flag to true
	// whenever the quality section is touched by ANY quality-section edit. The
	// forced value overrides any submitted input (defensive — the form no longer
	// emits this field). Pre-removal configs migrate naturally on their next
	// quality-section save. Saves that do not touch quality leave the file
	// untouched (LoadRaw round-trip preserves the on-disk byte content).
	if touched["quality"] {
		cfg.Quality.QualityExtrasEnabled = true
	}

	changed := 0
	for _, section := range []string{"git_strategy", "llm", "quality"} {
		if !touched[section] {
			continue
		}
		var value any
		switch section {
		case "git_strategy":
			value = cfg.GitStrategy
		case "llm":
			value = cfg.LLM
		case "quality":
			value = cfg.Quality
		}
		// REQ-WWS-003: skip the SetSection entirely when the applied values did
		// not change the section — an unchanged section must not raise the
		// git_strategy dirty flag nor ride the Save() rewrite.
		if reflect.DeepEqual(before[section], value) {
			continue
		}
		if err := mgr.SetSection(section, value); err != nil {
			return fmt.Errorf("settings: set %s section: %w", section, err)
		}
		changed++
	}
	if changed == 0 {
		// REQ-WWS-003: no section actually changed — Save() would still rewrite
		// all six section files (creating absent ones such as llm.yaml). Skip it.
		return nil
	}
	if err := mgr.Save(); err != nil {
		return fmt.Errorf("settings: save project config: %w", err)
	}
	return nil
}

// readSeamScalar returns the persisted scalar at path inside the section file,
// reporting ok=false when the file, the path, or the target is absent — the
// caller treats an unreadable current value as "unknown", so the edit is kept
// (a genuinely new value must still be written; an absent target is an upsert).
func readSeamScalar(projectRoot, section string, path []string) (string, bool) {
	data, err := os.ReadFile(filepath.Join(projectRoot, ".moai", "config", "sections", section+".yaml"))
	if err != nil {
		return "", false
	}
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return "", false
	}
	if doc.Kind != yaml.DocumentNode || len(doc.Content) == 0 {
		return "", false
	}
	cur := doc.Content[0]
	for i, key := range path {
		if cur.Kind != yaml.MappingNode {
			return "", false
		}
		idx := -1
		for j := 0; j+1 < len(cur.Content); j += 2 {
			if cur.Content[j].Value == key {
				idx = j
				break
			}
		}
		if idx < 0 {
			return "", false
		}
		cur = cur.Content[idx+1]
		if i == len(path)-1 && cur.Kind == yaml.ScalarNode {
			return cur.Value, true
		}
	}
	return "", false
}

// parseBoolValue는 typed applier의 bool 값 변환 가드다 (웹 파서가 1차 검증하지만
// seam 없이 직접 호출되는 경우를 방어). M4 다이어트로 int 변환 경로가 제거되어
// parseIntValue는 폐기되었다.
func parseBoolValue(key, v string) (bool, error) {
	switch v {
	case "true":
		return true, nil
	case "false":
		return false, nil
	}
	return false, fmt.Errorf("settings: %s: %q is not a boolean", key, v)
}

// applyGitStrategyKey는 git_strategy.<key> 편집을 typed struct에 적용한다
// (REQ-WC11-010 — typed, Save dirty-flag 경로). mode + profile별 hooks.pre_push
// (hook_pre_push.go:72 런타임 reader) + profile별 merge_method(SPEC-WEB-CONSOLE-014
// M3 — agent-prose consumer). 나머지 검증 전용 키의 편집 FieldDef는 제거되어
// 도달 불가하다 — struct 멤버는 보존.
func applyGitStrategyKey(gs *config.GitStrategyConfig, key, v string) error {
	switch key {
	case "mode":
		gs.Mode = v
		return nil
	case "worktree_base_branch":
		// SPEC-WORKTREE-BASEREF-001 REQ-WBR-013. Free text (REQ-WBR-014), so any
		// branch name is accepted and the empty value is the neutral
		// take-no-action state. Trimmed here as well as on read, so a value with
		// surrounding whitespace is never persisted untrimmed.
		gs.WorktreeBaseBranch = strings.TrimSpace(v)
		return nil
	}

	profileName, rest, ok := strings.Cut(key, ".")
	if !ok {
		return fmt.Errorf("settings: unknown git_strategy key %q", key)
	}
	var p *config.ModeProfile
	switch profileName {
	case "manual":
		p = &gs.Manual
	case "personal":
		p = &gs.Personal
	case "team":
		p = &gs.Team
	default:
		return fmt.Errorf("settings: unknown git_strategy profile %q", profileName)
	}

	switch rest {
	case "hooks.pre_push":
		p.Hooks.PrePush = v
	case "merge_method":
		// enum 멤버십 검증 (REQ-WC14-011). 빈 문자열은 enum 비멤버이므로 거부된다
		// (empty NOT a member — AC-WC14-010c). enum SSOT는 config.IsValidMergeMethod.
		if !config.IsValidMergeMethod(v) {
			return fmt.Errorf("settings: git_strategy.%s.merge_method: %q is not a valid merge method", profileName, v)
		}
		p.MergeMethod = v
	default:
		return fmt.Errorf("settings: unknown git_strategy key %q", key)
	}
	return nil
}

// applyLLMKey는 llm.<key> 편집을 typed struct에 적용한다 (REQ-WC11-012 안전 키만;
// mode/team_mode는 read-only — 스키마에 편집 필드가 없어 여기 도달 불가하며,
// 도달 시 명시적으로 거부한다, REQ-WC11-013). M4 다이어트로 performance_tier와
// claude_models.* 편집 FieldDef가 제거되어 이 분기들은 도달 불가하다 — struct
// 멤버는 보존되어 yaml 로드가 backward-compat를 유지한다. legacy alias
// glm.models.{opus,sonnet,haiku} apply 분기는 SPEC-WEB-CONSOLE-012
// REQ-WC12-003에서 제거되었다 (GLMModels legacy 멤버 자체는 REQ-WC12-006 보존 —
// yaml 로드/re-marshal이 기존 키를 파괴하지 않는다).
func applyLLMKey(l *config.LLMConfig, key, v string) error {
	switch key {
	case "mode", "team_mode":
		return fmt.Errorf("settings: llm.%s is read-only (runtime-managed, REQ-WC11-013)", key)
	case "glm.models.high":
		l.GLM.Models.High = v
	case "glm.models.medium":
		l.GLM.Models.Medium = v
	case "glm.models.low":
		l.GLM.Models.Low = v
	case "glm.models.fable":
		l.GLM.Models.Fable = v
	// SPEC-WEB-CONSOLE-REDESIGN-001 M4: per-tier reasoning effort. Since RC3
	// (glm-settings-persist) these ARE load-bearing: the GLM launcher
	// (internal/cli resolveGLMMainSessionEffort) reads the slot serving the
	// main session's model and lets a non-empty value override the
	// prefs/model_policy effort chain at the next moai glm launch. Sub-agents
	// keep the session-global ANTHROPIC_REASONING_EFFORT derived from
	// llm.effort_level; the collapse overlay governs the wire value (stored
	// high and max both wire as max).
	case "glm.effort.high":
		l.GLM.Effort.High = v
	case "glm.effort.medium":
		l.GLM.Effort.Medium = v
	case "glm.effort.low":
		l.GLM.Effort.Low = v
	case "glm.effort.fable":
		l.GLM.Effort.Fable = v
	default:
		return fmt.Errorf("settings: unknown llm key %q", key)
	}
	return nil
}

// applyQualityKey는 quality.<key> 확장 편집을 typed struct에 적용한다
// (REQ-WC11-011). M4 다이어트로 런타임 reader가 있는 3개 DDD 게이트 키만 잔류한다
// (trust.go:740/748/756). 나머지 키의 편집 FieldDef가 제거되어 도달 불가하다 —
// struct 멤버는 보존되어 yaml 로드가 backward-compat를 유지한다.
func applyQualityKey(q *models.QualityConfig, key, v string) error {
	b, err := parseBoolValue(key, v)
	if err != nil {
		return fmt.Errorf("settings: unknown quality key %q", key)
	}
	switch key {
	case "quality_extras_enabled":
		q.QualityExtrasEnabled = b
	case "ddd_settings.characterization_tests":
		q.DDDSettings.CharacterizationTests = b
	case "ddd_settings.behavior_snapshots":
		q.DDDSettings.BehaviorSnapshots = b
	case "ddd_settings.preserve_before_improve":
		q.DDDSettings.PreserveBeforeImprove = b
	default:
		return fmt.Errorf("settings: unknown quality key %q", key)
	}
	return nil
}
