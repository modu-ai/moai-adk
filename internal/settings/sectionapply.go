package settings

// 이 파일은 M2b 확장 필드의 영속화 디스패처다 (SPEC-WEB-CONSOLE-011 M2b).
//
// @MX:WARN: [AUTO] ApplySchemaEdits는 10섹션 확장 필드를 디스크에 쓰는 영속화
// 경계다. 전 필드가 WriteSectionViaSeam(yamlpatch — 주석/미모델링 키 보존)으로
// 기록된다: seam 섹션은 그대로, typed 섹션(git_strategy/llm/quality)은
// SPEC-WEB-SAVE-LOSSLESS-001 M1에서 라인-스플라이스 라우팅으로 전환되어
// SetSection → Save 전체-재마샬은 더 이상 이 경계에 없다.
// @MX:REASON: [AUTO] workflow.yaml 등 seam 섹션뿐 아니라 typed 파일에도
// typed re-marshal을 적용하면 주석과 미모델링 키가 파괴된다 (REQ-WC11-005/017,
// AP-1/AP-11, REQ-WSL-002/003 — GitHub issue #1731의 결함 기제). 필드별
// 라우팅 판정과 검증은 FieldDef.Persist.Kind + 구 typed applier가 SSOT이며,
// 여기 없는 키(read-only: llm.mode/team_mode, db system 5키)는 어떤 경로로도
// 기록되지 않는다 (REQ-WC11-013/019).

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
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
			// REQ-WSL-004 (SPEC-WEB-SAVE-LOSSLESS-001, AC-WSL-004): an ABSENT
			// key plus an empty submission is a no-op — absence already IS the
			// unset state, so writing `key: ""` would only add a byte-level
			// artifact (the D3 defect: workflow.yaml audit pins growing empty
			// keys). A key that EXISTS keeps its delete semantics: the
			// curOk && cur != "" path above still writes the "" (crosssession
			// round-trip lineage).
			if !curOk && edits[name] == "" {
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
		case PersistUserScoped:
			// SPEC-FEEDBACK-PARTICIPATION-001 REQ-ANON-020: the value lives in
			// the USER-scoped consent file, and the same value-invariant shapes
			// as the seam gate apply — a submission equal to the persisted
			// value, or a false submission over an absent key (the declared
			// AbsentDefault), writes nothing. A real change writes BOTH
			// participation.enabled and participation.asked=true in one patch:
			// the console description carries the full disclosure, so a
			// console-first user records an informed consent and the wizard
			// and update prompts never re-ask.
			cur := config.ReadUserParticipation()
			if edits[name] == strconv.FormatBool(cur.Enabled) {
				continue
			}
			if edits[name] == "false" && !cur.Asked && !userParticipationFileExists() {
				// Absent file and never asked: false IS the current state.
				continue
			}
			if edits[name] != "true" && edits[name] != "false" {
				return fmt.Errorf("settings: field %q: user-scoped bool accepts only true/false, got %q", name, edits[name])
			}
			if err := WriteUserParticipation(config.UserParticipation{
				Enabled:    edits[name] == "true",
				Asked:      true,
				Repository: cur.Repository,
			}); err != nil {
				return fmt.Errorf("settings: field %q: %w", name, err)
			}
		default:
			return fmt.Errorf("settings: field %q is not a schema-section field (kind %s)", name, f.Persist.Kind)
		}
	}

	// typed 섹션: 필드별 yamlpatch 라인-스플라이스 (SPEC-WEB-SAVE-LOSSLESS-001
	// M1 — SetSection → Save 전체-재마샬은 폐기).
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

// userParticipationFileExists reports whether the user-scoped consent file is
// present — the absent-key arm of the user-scoped value-invariant gate.
func userParticipationFileExists() bool {
	path, err := config.UserParticipationFilePath()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

// applyTypedEdits는 typed 섹션(git_strategy/llm/quality) 필드를 yamlpatch seam
// 라인-스플라이스로 영속화한다 (SPEC-WEB-SAVE-LOSSLESS-001 M1 — REQ-WSL-002/003).
//
// 구 경로(LoadRaw → SetSection → Save)는 편집-대상 섹션 파일 전체를 재마샬해
// 주석·미모델링 키를 파괴했다 (GitHub issue #1731의 결함 기제). 본 경로는
// FieldDef.Persist.Key를 yamlpatch 경로로 매핑(fieldYAMLPath 재사용)해 대상
// 행만 재작성하고, upsert(부재 키 신설)만 재직렬화 폴백을 승계한다 (C3).
//
// 검증과 정규화는 구 typed applier(applyGitStrategyKey/applyLLMKey/
// applyQualityKey)를 그대로 거친다 — FieldDef 스키마가 편집 가능 표면의 SSOT라는
// C2와 merge_method enum·bool 검증을 유지하기 위해서다. applier는 구조체에
// 적용되지만 그 구조체는 Save로 영속화되지 않는다 — 값 확정(정규화 반영)과
// no-op 게이트의 현재값 산출에만 쓰인다.
//
// no-op 게이트(REQ-WSL-001): 제출값(정규화 후)이 로드된 현재값(컴파일 기본값
// 반영 — 구 DeepEqual 게이트와 동일한 판정)과 같으면 스킵, 디스크 스칼라가 이미
// 제출값과 같으면(env-override 케이스) 스킵 — 동치 제출은 mtime 포함 무기록이다.
// 부재 키 + "" 제출도 스킵한다 (REQ-WSL-004).
func applyTypedEdits(projectRoot string, fields []FieldDef, values []string) error {
	mgr := config.NewConfigManager()
	cfg, err := mgr.LoadRaw(projectRoot)
	if err != nil {
		return fmt.Errorf("settings: load project config: %w", err)
	}

	seamEdits := map[string][]yamlpatch.KeyEdit{}
	for i, f := range fields {
		path := fieldYAMLPath(f)
		file := sectionFileFor(f)
		if file == "" || len(path) == 0 {
			return fmt.Errorf("settings: no seam target for typed field %q (section %q)", f.Name, f.Persist.Section)
		}

		var cur, next string
		switch f.Persist.Section {
		case "git_strategy":
			cur = gitStrategyValue(cfg.GitStrategy, f.Persist.Key)
			if err := applyGitStrategyKey(&cfg.GitStrategy, f.Persist.Key, values[i]); err != nil {
				return err
			}
			next = gitStrategyValue(cfg.GitStrategy, f.Persist.Key)
		case "llm":
			cur = llmValue(cfg.LLM, f.Persist.Key)
			if err := applyLLMKey(&cfg.LLM, f.Persist.Key, values[i]); err != nil {
				return err
			}
			next = llmValue(cfg.LLM, f.Persist.Key)
		case "quality":
			cur = qualityValue(cfg.Quality, f.Persist.Key)
			if err := applyQualityKey(&cfg.Quality, f.Persist.Key, values[i]); err != nil {
				return err
			}
			next = qualityValue(cfg.Quality, f.Persist.Key)
		default:
			return fmt.Errorf("settings: no typed applier for section %q", f.Persist.Section)
		}

		// REQ-WSL-001: 제출값 == 로드 현재값(기본값 포함) → 무기록. 구
		// DeepEqual 게이트의 필드 단위 환원이다.
		if next == cur {
			continue
		}
		// 동치 스플라이스 방지: 디스크가 이미 제출값을 담고 있으면(env
		// override 케이스) 기록해도 바이트 불변일 뿐이므로 mtime 오염을 피한다.
		if curFile, ok := readSeamScalar(projectRoot, file, path); ok && curFile == next {
			continue
		}
		// REQ-WSL-004: 부재 키 + "" 제출 → 부재가 이미 미설정 상태다.
		if next == "" {
			if _, ok := readSeamScalar(projectRoot, file, path); !ok {
				continue
			}
		}
		// GitHub #1731 (card t1474): a radio whose stored value is empty
		// renders preselected on its declared Default (radioEffectiveValue),
		// so an untouched form submits that Default for an absent key. That
		// submission is exactly what the page showed — not an edit — and
		// must not materialize the key on disk.
		if cur == "" && f.Default != "" && next == f.Default {
			if _, ok := readSeamScalar(projectRoot, file, path); !ok {
				continue
			}
		}
		seamEdits[file] = append(seamEdits[file], yamlpatch.KeyEdit{Path: path, Value: next})
	}

	files := make([]string, 0, len(seamEdits))
	for file := range seamEdits {
		files = append(files, file)
	}
	sort.Strings(files)
	for _, file := range files {
		if err := WriteSectionViaSeam(projectRoot, file, seamEdits[file]); err != nil {
			return err
		}
	}
	return nil
}

// gitStrategyValue는 git_strategy.<key>의 로드된 현재값을 문자열로 반환한다.
// applyGitStrategyKey의 setter 스위치를 거울처럼 반영하는 getter로, no-op 게이트의
// 현재값 산출 단일 원천이다. 정규화(공백 제거)는 applier 적용 후 재판정이
// 담당하므로 getter는 저장값을 그대로 읽는다.
func gitStrategyValue(gs config.GitStrategyConfig, key string) string {
	switch key {
	case "mode":
		return gs.Mode
	case "worktree_base_branch":
		return gs.WorktreeBaseBranch
	}
	profileName, rest, ok := strings.Cut(key, ".")
	if !ok {
		return ""
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
		return ""
	}
	switch rest {
	case "hooks.pre_push":
		return p.Hooks.PrePush
	case "merge_method":
		return p.MergeMethod
	}
	return ""
}

// llmValue는 llm.<key>의 로드된 현재값을 문자열로 반환한다 (applyLLMKey의
// getter 짝).
func llmValue(l config.LLMConfig, key string) string {
	switch key {
	case "glm.models.high":
		return l.GLM.Models.High
	case "glm.models.medium":
		return l.GLM.Models.Medium
	case "glm.models.low":
		return l.GLM.Models.Low
	case "glm.models.fable":
		return l.GLM.Models.Fable
	case "glm.effort.high":
		return l.GLM.Effort.High
	case "glm.effort.medium":
		return l.GLM.Effort.Medium
	case "glm.effort.low":
		return l.GLM.Effort.Low
	case "glm.effort.fable":
		return l.GLM.Effort.Fable
	}
	return ""
}

// qualityValue는 quality.<key>의 로드된 현재값을 문자열로 반환한다
// (applyQualityKey의 getter 짝 — bool은 FormatBool 정규형).
func qualityValue(q models.QualityConfig, key string) string {
	switch key {
	case "quality_extras_enabled":
		return strconv.FormatBool(q.QualityExtrasEnabled)
	case "ddd_settings.characterization_tests":
		return strconv.FormatBool(q.DDDSettings.CharacterizationTests)
	case "ddd_settings.behavior_snapshots":
		return strconv.FormatBool(q.DDDSettings.BehaviorSnapshots)
	case "ddd_settings.preserve_before_improve":
		return strconv.FormatBool(q.DDDSettings.PreserveBeforeImprove)
	}
	return ""
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
