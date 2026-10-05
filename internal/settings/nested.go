package settings

// 이 파일은 중첩 프로젝트-설정 영속화 seam 을 담는다. 기존에는
// internal/web/projectconfig.go 에만 존재했으나(writeProjectNestedConfig:242),
// internal/cli TUI 도 동일 seam 을 구동해야 하므로(REQ-WC10-011) internal/cli 가
// internal/web 를 import 하지 않고도 호출할 수 있는 중립 위치로 재배치했다(M2).
// HTTP 폼 파싱(parseProjectNestedForm)은 web-request 전용이므로 internal/web 에
// 남는다 — 이 seam 은 파싱된 NestedForm 만 받는다.
//
// @MX:WARN: [AUTO] WriteProjectNestedConfig 는 프로필 스토어가 아닌 *프로젝트 설정*
// (quality.yaml + git-convention.yaml)을 디스크에 쓰는 영속화 경계다. 두 표면(웹/TUI)이
// 공유한다.
// @MX:REASON: [AUTO] 영속화는 반드시 settings.WriteSectionViaSeam/yamlpatch 라인-스플라이스를
// 통해서만 수행한다 (SPEC-WEB-SAVE-LOSSLESS-001 REQ-WSL-002/003) — 구
// SetSection/Save 전체-재마샬은 GitHub issue #1731이 지목한 미모델링 키와 주석을
// 파괴하는 금지된 경로가 됐다(AP-2). *Set 플래그가 켜진 필드만 편집 대상이며
// (empty=preserve, REQ-WC10-012), 제출-동치 필드는 no-op 게이트에서 스킵된다.

import (
	"fmt"
	"strconv"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/settings/yamlpatch"
)

// NestedForm은 파싱된 7개 중첩 프로젝트-설정 필드를 운반한다. 각 필드의 *Set
// 플래그는 empty=preserve 규칙(REQ-WC10-012)을 필드 단위로 적용하게 한다: 미제출
// 필드는 디스크 값을 유지한다. internal/web 의 projectNestedForm 이 이 구조체의
// 상위 집합이었으며(ParseErrs 추가), M2 에서 본 구조체로 정렬되었다.
type NestedForm struct {
	CoverageTarget    int
	CoverageTargetSet bool

	EnforceQuality    bool
	EnforceQualitySet bool

	MinCoverage    int
	MinCoverageSet bool

	Confidence    float64
	ConfidenceSet bool

	AutoEnabled    bool
	AutoEnabledSet bool

	SampleSize    int
	SampleSizeSet bool

	EnforceOnPush    bool
	EnforceOnPushSet bool
}

// TouchesQuality는 폼이 quality 중첩 필드를 하나라도 운반하는지 보고한다.
func (f NestedForm) TouchesQuality() bool {
	return f.CoverageTargetSet || f.EnforceQualitySet || f.MinCoverageSet
}

// TouchesGitConvention는 폼이 git_convention 중첩 필드를 하나라도 운반하는지 보고한다.
func (f NestedForm) TouchesGitConvention() bool {
	return f.ConfidenceSet || f.AutoEnabledSet || f.SampleSizeSet || f.EnforceOnPushSet
}

// NestedCurrent는 GET echo-back 용으로 중첩 필드의 디스크 현재값을 운반한다.
// int/float 는 numberField 위젯 value= 속성을 위해 문자열로 사전-포맷되고, bool 은
// 토글 checked 상태를 구동한다.
type NestedCurrent struct {
	CoverageTarget       string
	EnforceQuality       bool
	MinCoverage          string
	ConfidenceThreshold  string
	AutoDetectionEnabled bool
	SampleSize           string
	EnforceOnPush        bool
}

// ReadProjectNestedConfig는 7개 중첩 필드의 read seam 이다. config 매니저로
// LoadRaw(검증 없는 write-intent 경로)하여 디스크 현재값을 반환한다. config 디렉터리
// 부재 시 LoadRaw 컴파일-인 기본값을 반환한다(panic 없음).
func ReadProjectNestedConfig(projectRoot string) (NestedCurrent, error) {
	mgr := config.NewConfigManager()
	cfg, err := mgr.LoadRaw(projectRoot)
	if err != nil {
		return NestedCurrent{}, fmt.Errorf("read project nested config: %w", err)
	}
	return NestedCurrent{
		CoverageTarget:       strconv.Itoa(cfg.Quality.TestCoverageTarget),
		EnforceQuality:       cfg.Quality.EnforceQuality,
		MinCoverage:          strconv.Itoa(cfg.Quality.TDDSettings.MinCoveragePerCommit),
		ConfidenceThreshold:  strconv.FormatFloat(cfg.GitConvention.AutoDetection.ConfidenceThreshold, 'f', -1, 64),
		AutoDetectionEnabled: cfg.GitConvention.AutoDetection.Enabled,
		SampleSize:           strconv.Itoa(cfg.GitConvention.AutoDetection.SampleSize),
		EnforceOnPush:        cfg.GitConvention.Validation.EnforceOnPush,
	}, nil
}

// WriteProjectNestedConfig는 7개 중첩 필드의 load-modify-write seam 이다.
// SPEC-WEB-SAVE-LOSSLESS-001 M1 (plan-audit iter-2 F5, REQ-WSL-002/003): 구
// 경로(SetSection → Save 전체-재마샬)는 GitHub issue #1731이 지목한
// constitution.session_effort_default 같은 미모델링 키와 주석을 중첩 편집 한 번에
// 소실시켰다. 본 경로는 각 *Set 필드를 yamlpatch 라인-스플라이스 편집으로
// 변환해 대상 행만 재작성한다. 미제출 필드(empty=preserve, REQ-WC10-012)는
// 애초에 편집 대상이 아니고, 제출-동치 필드는 no-op 게이트에서 스킵된다 —
// 모든 형제 행(미모델링 키·주석 포함)은 원문 바이트로 통과한다.
//
// no-op 게이트: 제출값(문자열 정규형)이 로드된 현재값(기본값 반영 — 구
// DeepEqual 게이트의 필드 단위 환원)과 같으면 스킵; 디스크 스칼라가 이미
// 제출값이면 스킵(env-override 동치 기록 방지, REQ-WSL-001 mtime 포함).
func WriteProjectNestedConfig(projectRoot string, form NestedForm) error {
	mgr := config.NewConfigManager()
	cfg, err := mgr.LoadRaw(projectRoot)
	if err != nil {
		return fmt.Errorf("load project config: %w", err)
	}

	edits := map[string][]yamlpatch.KeyEdit{}
	// add는 편집 대상(제출됨) 필드를 no-op 게이트로 검한 뒤 seam 편집 대기열에
	// 넣는다. cur는 LoadRaw 구조체 값의 문자열 정규형이다.
	add := func(file string, path []string, cur, next string) {
		if next == cur {
			return
		}
		if disk, ok := readSeamScalar(projectRoot, file, path); ok && disk == next {
			return
		}
		edits[file] = append(edits[file], yamlpatch.KeyEdit{Path: path, Value: next})
	}

	if form.TouchesQuality() {
		if form.CoverageTargetSet {
			add("quality", []string{"constitution", "test_coverage_target"},
				strconv.Itoa(cfg.Quality.TestCoverageTarget), strconv.Itoa(form.CoverageTarget))
		}
		if form.EnforceQualitySet {
			add("quality", []string{"constitution", "enforce_quality"},
				strconv.FormatBool(cfg.Quality.EnforceQuality), strconv.FormatBool(form.EnforceQuality))
		}
		if form.MinCoverageSet {
			// 중첩-of-중첩: TDDSettings 아래 한 필드.
			add("quality", []string{"constitution", "tdd_settings", "min_coverage_per_commit"},
				strconv.Itoa(cfg.Quality.TDDSettings.MinCoveragePerCommit), strconv.Itoa(form.MinCoverage))
		}
	}

	if form.TouchesGitConvention() {
		if form.ConfidenceSet {
			add("git-convention", []string{"git_convention", "auto_detection", "confidence_threshold"},
				strconv.FormatFloat(cfg.GitConvention.AutoDetection.ConfidenceThreshold, 'f', -1, 64),
				strconv.FormatFloat(form.Confidence, 'f', -1, 64))
		}
		if form.AutoEnabledSet {
			add("git-convention", []string{"git_convention", "auto_detection", "enabled"},
				strconv.FormatBool(cfg.GitConvention.AutoDetection.Enabled), strconv.FormatBool(form.AutoEnabled))
		}
		if form.SampleSizeSet {
			add("git-convention", []string{"git_convention", "auto_detection", "sample_size"},
				strconv.Itoa(cfg.GitConvention.AutoDetection.SampleSize), strconv.Itoa(form.SampleSize))
		}
		if form.EnforceOnPushSet {
			add("git-convention", []string{"git_convention", "validation", "enforce_on_push"},
				strconv.FormatBool(cfg.GitConvention.Validation.EnforceOnPush), strconv.FormatBool(form.EnforceOnPush))
		}
	}

	for _, file := range []string{"quality", "git-convention"} {
		if len(edits[file]) == 0 {
			continue
		}
		if err := WriteSectionViaSeam(projectRoot, file, edits[file]); err != nil {
			return err
		}
	}
	return nil
}
