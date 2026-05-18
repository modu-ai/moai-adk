// Package router는 SPEC 복잡도 신호를 기반으로 harness 레벨(minimal/standard/thorough)을
// 결정하는 라우팅 로직을 제공합니다.
// REQ-HRN-001-003: HarnessRouter.Route(spec, cfg) → (Level, Rationale, error).
package router

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/spec"
)

// Level은 harness 라우팅 레벨 타입입니다.
// @MX:ANCHOR: [AUTO] FROZEN level enum — {minimal, standard, thorough} (REQ-HRN-001-017)
// @MX:REASON: FROZEN per SPEC-V3R2-HRN-001 REQ-017; 다수의 소비자 (CLI, router, effort.go, escalation.go)
type Level string

const (
	// LevelMinimal은 간단한 변경을 위한 최소 수준 harness입니다.
	LevelMinimal Level = "minimal"
	// LevelStandard는 일반 개발을 위한 표준 수준 harness입니다.
	LevelStandard Level = "standard"
	// LevelThorough는 중요한 기능을 위한 최고 수준 harness입니다.
	LevelThorough Level = "thorough"
)

// Rationale는 라우팅 결정의 근거를 담는 구조체입니다.
// REQ-HRN-001-003: Route()의 두 번째 반환값.
type Rationale struct {
	// MatchedRule은 라우팅 결정을 내린 규칙 이름입니다.
	// 예: "spec_override", "force_thorough", "auto_minimal", "auto_standard", "fallthrough_default"
	MatchedRule string `json:"matched_rule"`
	// FileCount는 추정된 관련 파일 수입니다.
	FileCount int `json:"file_count"`
	// DomainCount는 추정된 관련 도메인 수입니다.
	DomainCount int `json:"domain_count"`
	// SpecType은 추정된 SPEC 유형입니다.
	// 예: "bugfix", "docs", "config", "feature", "refactor", "other"
	SpecType string `json:"spec_type"`
	// SpecPriority는 SPEC 우선순위 (frontmatter priority 필드)입니다.
	SpecPriority string `json:"spec_priority"`
	// Keywords는 force-thorough를 발동한 매칭된 키워드 목록입니다.
	Keywords []string `json:"keywords"`
}

// SPECInput은 라우터가 소비하는 SPEC 입력 구조체입니다.
// spec.SPECFrontmatter에서 복잡도 신호를 추출할 때 사용됩니다.
type SPECInput struct {
	// Priority는 frontmatter priority 필드입니다 (예: "P0", "P1 Critical").
	Priority string
	// Tags는 frontmatter tags 필드 (comma-separated)입니다.
	Tags string
	// Title은 frontmatter title 필드입니다.
	Title string
	// Body는 SPEC 문서 본문 (Requirements 섹션 포함)입니다.
	Body string
	// HarnessLevel은 frontmatter harness_level 필드입니다 (optional override).
	// REQ-HRN-001-015.
	HarnessLevel string
}

// Router는 SPEC 복잡도 신호를 기반으로 harness 레벨을 결정하는 인터페이스입니다.
// @MX:ANCHOR: [AUTO] harness 라우팅 인터페이스 — Route, RouteFromFile 메서드
// @MX:REASON: fan_in >= 3: CLI route 명령, 테스트, ConfigManager 통합 등 다수의 소비자
type Router interface {
	// Route는 SPECInput과 HarnessConfig를 기반으로 레벨과 근거를 반환합니다.
	Route(doc *SPECInput, cfg *config.HarnessConfig) (Level, Rationale, error)
	// RouteFromFile은 SPEC 파일 경로에서 직접 라우팅을 수행합니다.
	RouteFromFile(specPath string, cfg *config.HarnessConfig) (Level, Rationale, error)
}

// defaultRouter는 Router 인터페이스의 기본 구현입니다.
type defaultRouter struct {
	cfg *config.HarnessConfig
}

// New는 새 Router 인스턴스를 반환합니다.
func New(cfg *config.HarnessConfig) Router {
	return &defaultRouter{cfg: cfg}
}

// Route는 SPECInput과 HarnessConfig를 기반으로 harness 레벨을 결정합니다.
// 우선순위 순서 (REQ-HRN-001-003/007/008/015):
//  1. SPEC frontmatter harness_level: 오버라이드 (REQ-015) — 최고 우선순위
//  2. force-thorough 오버라이드 (REQ-008) — 보안/결제 키워드, critical 우선순위
//  3. 에스컬레이션 (REQ-009) — 누적 (라우터는 초기값만 결정)
//  4. auto_detection 규칙 (REQ-007) — minimal → standard → thorough
//  5. mode_defaults (REQ-014) — 가장 낮은 우선순위 폴백
func (r *defaultRouter) Route(doc *SPECInput, cfg *config.HarnessConfig) (Level, Rationale, error) {
	signals := ExtractSignals(doc)

	rationale := Rationale{
		FileCount:    signals.FileCount,
		DomainCount:  signals.DomainCount,
		SpecType:     signals.SpecType,
		SpecPriority: doc.Priority,
		Keywords:     []string{},
	}

	// 1. SPEC frontmatter harness_level: 오버라이드 (REQ-HRN-001-015)
	if doc.HarnessLevel != "" {
		switch Level(doc.HarnessLevel) {
		case LevelMinimal, LevelStandard, LevelThorough:
			rationale.MatchedRule = "spec_override"
			return Level(doc.HarnessLevel), rationale, nil
		}
		// 알 수 없는 harness_level 값은 무시하고 계속 진행
	}

	// 2. force-thorough 오버라이드 (REQ-HRN-001-008)
	forcedKeywords := matchForceThoroughKeywords(doc)
	isCritical := isCriticalPriority(doc.Priority)
	isSensitiveDomain := isSensitiveTagDomain(doc.Tags)

	if len(forcedKeywords) > 0 || isCritical || isSensitiveDomain {
		rationale.MatchedRule = "force_thorough"
		rationale.Keywords = forcedKeywords
		return LevelThorough, rationale, nil
	}

	// 3. auto_detection 규칙 (REQ-HRN-001-007): minimal → standard → thorough 순서
	level, rule := applyAutoDetectionRules(signals, cfg)
	rationale.MatchedRule = rule
	return level, rationale, nil
}

// RouteFromFile은 SPEC 파일 경로에서 직접 라우팅을 수행합니다.
func (r *defaultRouter) RouteFromFile(specPath string, cfg *config.HarnessConfig) (Level, Rationale, error) {
	doc, err := loadSPECDoc(specPath)
	if err != nil {
		return LevelStandard, Rationale{MatchedRule: "fallthrough_default"}, fmt.Errorf("RouteFromFile: %w", err)
	}
	return r.Route(doc, cfg)
}

// loadSPECDoc은 파일 경로에서 SPECInput을 로드합니다.
func loadSPECDoc(specPath string) (*SPECInput, error) {
	data, err := os.ReadFile(specPath)
	if err != nil {
		return nil, fmt.Errorf("read spec file: %w", err)
	}

	// spec 패키지의 파서를 사용합니다
	content := string(data)
	fm, body, parseErr := extractSPECFrontmatter(content)
	if parseErr != nil {
		// 파싱 오류 시 기본값으로 진행
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

// extractSPECFrontmatter는 spec.SPECFrontmatter를 파싱합니다.
// internal/spec/lint.go의 parseSPECDoc를 간접적으로 사용합니다.
func extractSPECFrontmatter(content string) (spec.SPECFrontmatter, string, error) {
	return spec.ExtractFrontmatter(content)
}

// applyAutoDetectionRules는 복잡도 신호를 기반으로 레벨을 결정합니다.
// REQ-HRN-001-007: minimal → standard → thorough 우선순위 순서.
func applyAutoDetectionRules(signals ComplexitySignals, cfg *config.HarnessConfig) (Level, string) {
	// minimal 조건: file_count <= 3 AND single_domain AND spec_type in [bugfix, docs, config]
	isMinimalSpecType := signals.SpecType == "bugfix" || signals.SpecType == "docs" || signals.SpecType == "config"
	if signals.FileCount <= 3 && signals.DomainCount <= 1 && isMinimalSpecType {
		return LevelMinimal, "auto_minimal"
	}

	// standard 조건: file_count > 3 OR multi_domain OR spec_type in [feature, refactor]
	isStandardSpecType := signals.SpecType == "feature" || signals.SpecType == "refactor"
	if signals.FileCount > 3 || signals.DomainCount > 1 || isStandardSpecType {
		return LevelStandard, "auto_standard"
	}

	// 폴백: standard (REQ-HRN-001-007 폴백은 standard로 정의)
	return LevelStandard, "fallthrough_default"
}

// isCriticalPriority는 우선순위 문자열이 Critical을 의미하는지 확인합니다.
// REQ-HRN-001-008: spec_priority == critical → force thorough.
// 주의: P0는 Critical 등가이지만 단독 P0 값만으로는 force_thorough를 발동하지 않습니다.
// "P0 Critical" 또는 "Critical" 같이 명시적으로 critical 키워드가 포함된 경우에만 발동합니다.
func isCriticalPriority(priority string) bool {
	lower := strings.ToLower(priority)
	return strings.Contains(lower, "critical")
}

// isSensitiveTagDomain은 태그에 민감한 도메인이 포함되어 있는지 확인합니다.
// REQ-HRN-001-008: domain in [auth, payment, migration, public_api] → force thorough.
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

// ConfigProxy는 EffortForLevelFromProxy를 위한 경량 config 래퍼입니다.
// 테스트에서 사용됩니다.
type ConfigProxy struct {
	EffortMapping map[string]string
}
