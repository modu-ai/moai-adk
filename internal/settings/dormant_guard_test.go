package settings

// 이 파일은 SPEC-WEB-CONSOLE-014 P3 노출 금지 가드다 (REQ-WC14-050/051/012/021/031/040).
//
// 편집 가능 필드 집합(AllFields — 웹 콘솔에서 form 위젯으로 편집 가능한 필드)의
// 어떤 이름도 dormant/governance/scaffold/타 SPEC 소관 키 denylist와 매칭되지
// 않음을 기계적으로 고정한다. 가드는 현행 코드에서 즉시 GREEN이어야 정상이다
// (regression pin — TDD RED 대상이 아니다). RED가 나오면 그 자체가 현행 노출
// 결함 발견이며 blocker 신호다 (plan.md §D, AP-9).

import (
	"strings"
	"testing"
)

// denyMatchKind는 denylist 항목의 매칭 방식을 구분한다.
type denyMatchKind int

const (
	denyPrefix denyMatchKind = iota // strings.HasPrefix
	denySubstr                      // strings.Contains (mode 변형 포함 커버)
	denyExact                       // 정확 일치
)

// denyRule는 편집 금지 키 패턴 한 개다.
type denyRule struct {
	pattern string
	kind    denyMatchKind
	reason  string
}

// dormantDenylist는 편집 가능 집합에 유입되면 안 되는 키의 테이블 주도 denylist다
// (REQ-WC14-050 §B.6). 테이블 주도로 확장 가능하다.
//
// iter-2 B14: learning.auto_apply / observability.hook_metrics.output_path는
// M1 탑재 대상이 아니라 M2 강등과 함께 추가되는 TDD RED→GREEN 쌍이다 — 두 키가
// 아직 editable인 M1 시점에 선탑재하면 즉시 RED (AP-10 위반). M2 강등 커밋에서
// 두 항목을 추가한다.
func dormantDenylist() []denyRule {
	return []denyRule{
		// F10 — dormant config (sunset.* / harness.model_upgrade_review.*).
		{"sunset.", denyPrefix, "F10 dormant sunset config (no runtime enforcement)"},
		{"model_upgrade_review", denySubstr, "F10 dormant governance reminder (non-enforcing)"},
		// F11 — legacy flat model routing (typed Model Policy는 SPEC-WEB-CONSOLE-013 소관).
		{"workflow.model_routing", denyPrefix, "F11 legacy flat DEPRECATED block"},
		// F12 — tool-policy (dev-only, 템플릿 배포 의도적 제거).
		{"tool_policy", denySubstr, "F12 dev-only tool-policy"},
		{"tool-policy", denySubstr, "F12 dev-only tool-policy"},
		// F8 — mx는 RouteExcluded, 편집 필드 0 (raw view만).
		{"mx.", denyPrefix, "F8 mx section is RouteExcluded (raw view only)"},
		// F7 — permission scaffold (Go struct 바인딩 부재).
		{"pre_allowlist", denySubstr, "F7 Go-unbound permission scaffold"},
		{"session_rules", denySubstr, "F7 Go-unbound permission scaffold"},
		// F5 — team late-branch scaffold 4키 (Go 행동적 reader 0건; manual/personal/team
		// 전 mode 커버를 위해 substr 매칭).
		{"branch_creation.prompt_always", denySubstr, "F5 forward-compat scaffold (no reader)"},
		{"branch_creation.auto_enabled", denySubstr, "F5 forward-compat scaffold (no reader)"},
		{"automation.auto_branch", denySubstr, "F5 forward-compat scaffold (no reader)"},
		{"automation.auto_pr", denySubstr, "F5 forward-compat scaffold (no reader)"},
	}
}

// matches는 필드 이름이 denyRule에 매칭되는지 보고한다.
func (r denyRule) matches(name string) bool {
	switch r.kind {
	case denyPrefix:
		return strings.HasPrefix(name, r.pattern)
	case denySubstr:
		return strings.Contains(name, r.pattern)
	case denyExact:
		return name == r.pattern
	}
	return false
}

// TestDormantConfigNeverEditable는 편집 가능 필드 집합(AllFields)의 어떤 이름도
// dormant denylist와 매칭되지 않음을 단언한다 (REQ-WC14-050/012/021/031/040 §B.6).
// 동시에 REQ-WC14-040의 allowlist 핀(slow_hook_threshold_ms는 editable 잔류)을
// 함께 고정한다 (AC-WC14-040a).
func TestDormantConfigNeverEditable(t *testing.T) {
	t.Parallel()

	deny := dormantDenylist()
	for _, f := range AllFields() {
		for _, rule := range deny {
			if rule.matches(f.Name) {
				t.Errorf("editable field %q matches dormant denylist pattern %q (%s) — must NOT be editable (REQ-WC14-050)",
					f.Name, rule.pattern, rule.reason)
			}
		}
	}

	// AC-WC14-040a allowlist 핀: slow_hook_threshold_ms는 reader 실존(F9)이므로
	// 편집 가능 집합에 잔류해야 한다 (회귀 핀). 이 핀이 깨지면 정직화가 과도하게
	// editable 키까지 제거한 신호다.
	assertEditablePresent(t, "observability.hook_metrics.slow_hook_threshold_ms")
}

// assertEditablePresent는 주어진 이름이 편집 가능 필드 집합에 존재함을 단언한다.
func assertEditablePresent(t *testing.T, name string) {
	t.Helper()
	for _, f := range AllFields() {
		if f.Name == name {
			return
		}
	}
	t.Errorf("expected %q to be an editable field (allowlist pin), but it is absent from AllFields", name)
}
