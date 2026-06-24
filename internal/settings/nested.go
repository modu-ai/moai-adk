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
// @MX:REASON: [AUTO] 영속화는 반드시 config.NewConfigManager()/LoadRaw/SetSection/Save 를
// 통해서만 수행한다 — YAML 직접 marshal/os.WriteFile 는 금지된 안티패턴(AP-2). SetSection 은
// 섹션 구조체 전체를 교체하므로 LoadRaw 가 반환한 섹션 전체를 복사한 뒤 *Set 플래그가 켜진
// 필드만 변형한다(whole-section-copy + per-field *Set). 미제출 필드(empty=preserve, REQ-WC10-012)는
// 디스크 값을 그대로 유지한다.

import (
	"fmt"
	"strconv"

	"github.com/modu-ai/moai-adk/internal/config"
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
// SetSection 이 섹션 구조체 전체를 교체하고 Save 가 전체를 직렬화하므로, LoadRaw 가
// 반환한 섹션 구조체 전체를 복사(q := cfg.Quality / gc := cfg.GitConvention)한 뒤
// 대상 중첩 필드만 변형한다. 모든 형제 중첩 필드는 LoadRaw 에서 byte-identical 하게
// 통과한다. 각 *Set 플래그가 필드 단위 변형을 게이트하므로 미제출 필드(empty=preserve,
// REQ-WC10-012)는 디스크 값을 유지한다.
func WriteProjectNestedConfig(projectRoot string, form NestedForm) error {
	mgr := config.NewConfigManager()
	cfg, err := mgr.LoadRaw(projectRoot)
	if err != nil {
		return fmt.Errorf("load project config: %w", err)
	}

	changed := false

	if form.TouchesQuality() {
		q := cfg.Quality // 전체 구조체 복사: DDD/TDD/Coverage/LSP/... 전부 통과
		if form.CoverageTargetSet {
			q.TestCoverageTarget = form.CoverageTarget
		}
		if form.EnforceQualitySet {
			q.EnforceQuality = form.EnforceQuality
		}
		if form.MinCoverageSet {
			q.TDDSettings.MinCoveragePerCommit = form.MinCoverage // 중첩-of-중첩: TDDSettings 통과, 한 필드만 설정
		}
		if err := mgr.SetSection("quality", q); err != nil {
			return fmt.Errorf("set quality section: %w", err)
		}
		changed = true
	}

	if form.TouchesGitConvention() {
		gc := cfg.GitConvention // 전체 구조체 복사: AutoDetection/Validation 하위 구조체 전부 통과
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
