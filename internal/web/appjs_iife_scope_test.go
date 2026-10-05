package web

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
)

// SPEC-APPJS-IIFE-GUARD-001 — app.js IIFE 경계 참조 회귀 가드 (card t1048).
//
// 이 가드가 막는 결함 계열: `addEventListener` 의 핸들러 인자가 벌거벗은 식별자인데
// 그 식별자의 선언이 **다른** 최상위 IIFE 안에 있는 경우. 그러면 등록 줄이 로드
// 시점에 ReferenceError 를 던지고, 예외가 그 IIFE 의 최상위 실행을 중단시켜 **그 줄
// 뒤에 있던 무관한 등록들이 조용히 일어나지 않는다.** 문법 오류가 아니라 어떤 정적
// 문법 검사도 잡지 못하고, 콘솔을 열지 않는 한 아무 신호도 남지 않는다.
//
// ── 이 가드는 완전한 스코프 검사가 아니다 (spec.md §F 한계 선언) ──────────────
//
//  1. **다른 형태의 최상위 문장.** 장래에 최상위에 `someIdentifier();` 나
//     `obj.method(otherIdentifier)` 같은 다른 형태의 문장이 생기고 그것이 IIFE
//     경계를 넘으면, 이 가드는 잡지 못한다. 지금 그런 문장이 0개라는 전수조사는
//     **현재**에 대한 관측이지 앞으로도 0이리라는 보장이 아니다.
//  2. **중첩 함수 안의 선언.** 규칙은 "같은 IIFE 안 **어느 깊이든**" 선언되어
//     있으면 통과시킨다. 따라서 어떤 중첩 함수 본문 안에만 선언된 식별자를 그
//     IIFE 의 최상위에서 등록하면, 실제로는 스코프 밖인데도 규칙은 통과시킨다.
//     이것은 편의가 아니라 **거짓 음성**이며, 의도한 맞교환이다 — 정확히 가르려면
//     JS 파서 의존성이 필요하고 그것은 기각됐다.
//  3. **함수 표현식 핸들러 안의 교차 참조.** 표현식 핸들러는 자기 스코프를 데리고
//     다니므로 규칙 대상 밖이다(아래 exclusion 참조). 표현식 본문이 다른 IIFE 의
//     식별자를 참조하면 잡히지 않는다.
//  4. **"경계를 넘는 참조가 없다" 는 "핸들러가 실제로 발화한다" 를 함의하지 않는다.**
//     이 가드는 소스 문자열만 본다. 버튼 핸들러가 런타임에 실제로 붙고 발화하는지는
//     다른 축이며, card t1060 소관으로 남긴다 — 이 카드에서 닫지 않는다.
//
// 이 테스트의 초록을 "app.js 에 교차 경계 참조가 없다" 로 읽으면 과대 해석이다.
// 초록이 말하는 것은 "bare-identifier 핸들러 계열에 한해 경계를 넘지 않는다" 뿐이다.

// iifeSpan 은 최상위 IIFE 하나의 줄 범위다(1-기반, 양끝 포함).
type iifeSpan struct {
	start int
	end   int
}

// bareHandlerReg 는 핸들러 인자가 벌거벗은 식별자인 addEventListener 등록 하나다.
type bareHandlerReg struct {
	line int    // 1-기반 등록 줄 번호
	name string // 핸들러로 넘겨진 식별자
}

// scopeViolation 은 등록과 선언이 서로 다른 IIFE 에 있는 한 건이다.
type scopeViolation struct {
	line      int    // 등록 줄 번호 (REQ-AIG-003)
	name      string // 식별자 이름 (REQ-AIG-003)
	regSpan   int    // 등록이 속한 IIFE 인덱스
	declSpans []int  // 선언이 발견된 IIFE 인덱스들 (없으면 빈 슬라이스)
}

// String 은 REQ-AIG-003 이 요구하는 두 좌표(식별자 이름 + 등록 줄 번호)를 모두 담는다.
func (v scopeViolation) String() string {
	return fmt.Sprintf("line %d: %s — declared in IIFE#%v, not IIFE#%d",
		v.line, v.name, v.declSpans, v.regSpan)
}

// bareHandlerPattern 은 이벤트명이 문자열 리터럴이고 핸들러가 **벌거벗은 식별자**인
// 등록만 매치한다. 함수 표현식(`function (…) {…}` / 화살표)은 `(` 또는 다른 형태로
// 시작하므로 이 패턴에 걸리지 않는다 — REQ-AIG-005 의 제외가 정규식 자체로 충족된다.
//
// [HARD] 식별자 문자 집합은 `\w` 가 아니라 명시적 문자류다. 이 규칙의 프로토타입은
// Python 이었고 Python `re` 의 `\w` 는 로케일/유니코드 축이 Go RE2 와 다르다. 더
// 결정적으로 **Go 의 `\w` 는 `$` 를 포함하지 않는데 JavaScript 식별자는 `$` 를
// 합법적으로 포함한다** — `\w` 로 이식했다면 `$`-포함 핸들러를 조용히 놓쳤을 것이다.
var bareHandlerPattern = regexp.MustCompile(
	`\.addEventListener\(\s*"[^"]*"\s*,\s*([A-Za-z_$][A-Za-z0-9_$]*)\s*\)`)

// anyAddEventListenerPattern 은 형태를 가리지 않고 모든 등록을 센다. 검사 대수와
// 대조해 함수 표현식 핸들러가 실제로 제외됐음을 보이는 데 쓴다(AC-AIG-004).
var anyAddEventListenerPattern = regexp.MustCompile(`\.addEventListener\(`)

// declPattern 은 주어진 식별자의 선언(`function` / `var` / `let` / `const`)을 찾는
// 정규식을 만든다. 경계는 `\b` 가 아니라 명시적 non-identifier 문자류다 — Go 의
// `\b` 는 `$` 를 단어 문자로 보지 않아 `$` 를 포함한 식별자에서 잘못된 경계를 만든다.
func declPattern(name string) *regexp.Regexp {
	q := regexp.QuoteMeta(name)
	return regexp.MustCompile(
		`(?:^|[^A-Za-z0-9_$])(?:function|var|let|const)[ \t]+` + q + `(?:[^A-Za-z0-9_$]|$)`)
}

// scanTopLevelIIFESpans 는 컬럼 0 에서 열리고 닫히는 최상위 IIFE 의 줄 범위를
// 소스에서 **산출한다**. 스팬을 상수로 박지 않는다 (REQ-AIG-001) — 하드코딩된 줄
// 번호는 파일이 한 줄만 움직여도 조용히 다른 영역을 재게 된다.
func scanTopLevelIIFESpans(lines []string) []iifeSpan {
	var spans []iifeSpan
	open := -1
	for i, ln := range lines {
		switch {
		case open < 0 && strings.HasPrefix(ln, "(function"):
			open = i + 1
		case open >= 0 && strings.HasPrefix(ln, "})();"):
			spans = append(spans, iifeSpan{start: open, end: i + 1})
			open = -1
		}
	}
	return spans
}

// findBareHandlerRegs 는 벌거벗은 식별자를 핸들러로 넘기는 등록을 전부 찾는다.
func findBareHandlerRegs(lines []string) []bareHandlerReg {
	var regs []bareHandlerReg
	for i, ln := range lines {
		for _, m := range bareHandlerPattern.FindAllStringSubmatch(ln, -1) {
			regs = append(regs, bareHandlerReg{line: i + 1, name: m[1]})
		}
	}
	return regs
}

// spanIndexOf 는 주어진 줄이 속한 IIFE 인덱스를 돌려준다. 어느 스팬에도 속하지
// 않으면 -1 (판정 대상 밖).
func spanIndexOf(spans []iifeSpan, line int) int {
	for i, s := range spans {
		if line >= s.start && line <= s.end {
			return i
		}
	}
	return -1
}

// scanBareHandlerScope 는 이 가드의 규칙 그 자체다. 소스 문자열을 받아 위반 목록과
// **실제로 검사한 bare-identifier 핸들러 개수**를 돌려주는 순수 함수다.
//
// 순수 함수인 것이 설계의 핵심이다: 같은 함수를 정방향(실제 app.js)과 역방향
// (메모리에서 합성한 돌연변이) 두 입력으로 부를 수 있고, 그래서 디스크의 app.js 를
// 한 바이트도 건드리지 않고 "규칙이 그 결함을 거부한다" 를 보일 수 있다.
func scanBareHandlerScope(src string) ([]scopeViolation, int) {
	lines := strings.Split(src, "\n")
	spans := scanTopLevelIIFESpans(lines)

	var violations []scopeViolation
	checked := 0

	for _, reg := range findBareHandlerRegs(lines) {
		regSpan := spanIndexOf(spans, reg.line)
		if regSpan < 0 {
			// 두 IIFE 어느 쪽에도 속하지 않는 등록은 판정 대상 밖이고,
			// 검사 대수에도 반영하지 않는다.
			continue
		}
		checked++

		pat := declPattern(reg.name)
		declSpans := []int{}
		for i, s := range spans {
			for _, ln := range lines[s.start-1 : s.end] {
				if pat.MatchString(ln) {
					declSpans = append(declSpans, i)
					break
				}
			}
		}

		found := false
		for _, d := range declSpans {
			if d == regSpan {
				found = true
				break
			}
		}
		if !found {
			violations = append(violations, scopeViolation{
				line: reg.line, name: reg.name, regSpan: regSpan, declSpans: declSpans,
			})
		}
	}

	return violations, checked
}

// TestAppJsBareHandlersDeclaredInSameIIFE 는 정방향이다 — AC-AIG-001.
//
// [HARD] 두 단언이 **함께** 걸린다. 위반 0 만 단언하면 아무 핸들러도 찾지 못하는
// 고장난 규칙이 그 단언을 만족시킨다. 그 공허한 통과가 이 카드가 막으려는 것이다.
func TestAppJsBareHandlersDeclaredInSameIIFE(t *testing.T) {
	t.Parallel()
	src := readEmbeddedAsset(t, "app.js")

	violations, checked := scanBareHandlerScope(src)

	for _, v := range violations {
		t.Errorf("IIFE scope violation in app.js: %s", v)
	}

	// 비-공허 단언 (REQ-AIG-004). 이 단언 없이는 위반 0 이 아무것도 보장하지 않는다.
	//
	// [판단 기록 — 검사 대수를 정확한 값에 고정하지 않고 `> 0` 으로 둔 이유]
	// 고정하면 핸들러가 **정당하게** 추가될 때마다 이 테스트가 깨진다. 그러면
	// 테스트는 "고칠 것" 이 아니라 "숫자를 맞춰 줄 것" 이 되고, 숫자를 습관적으로
	// 올리는 순간 이 단언의 탐지 가치는 0이 된다.
	//
	// 그래서 놓치는 것 — 검사 대수가 커밋 사이에 조용히 줄어드는 퇴행 — 은
	// **감수한 것이지 흡수된 것이 아니다.** 아래 역방향 테스트의 대수 동일성
	// 단언은 이 구멍을 메우지 못한다: 그 단언은 **한 실행 안에서** 정방향과
	// 돌연변이의 대수를 비교하므로, 표면이 4→3 으로 줄면 돌연변이 쪽도 함께 3이
	// 되어 동일성은 그대로 통과한다. 대수 동일성이 잡는 것은 "돌연변이가 규칙이
	// 보는 표면 자체를 바꿨는가" 이지 "표면이 지난 커밋보다 줄었는가" 가 아니다.
	//
	// 대신 아래 `checked < total` 단언이 다른 축에서 일부를 되찾는다 — 그것도
	// 축소 구멍을 닫지는 않는다. 이 트리에서 그 구멍을 닫는 장치는 없다.
	if checked == 0 {
		t.Fatal("rule matched ZERO bare-identifier handlers — a rule that inspects nothing " +
			"passes the violations==0 assertion vacuously; this is the failure this guard exists to prevent")
	}

	// AC-AIG-004 — 함수 표현식 핸들러가 실제로 제외됐음을 보인다. 파일 전체의
	// addEventListener 등록 수보다 검사 대수가 **작다**는 것이 그 증거다(현재 트리에서
	// 다수는 함수 표현식 핸들러다). 정확한 두 값을 박지 않는 이유는 위 판단 기록과
	// 같다.
	total := len(anyAddEventListenerPattern.FindAllString(src, -1))
	t.Logf("forward: bare-identifier handlers checked = %d, violations = %d, total addEventListener calls = %d",
		checked, len(violations), total)
	if checked >= total {
		t.Errorf("function-expression handlers were NOT excluded: checked=%d of total=%d "+
			"addEventListener calls (expected strictly fewer — expression handlers carry their "+
			"own scope and are out of this rule's scope, REQ-AIG-005)", checked, total)
	}
}

// TestAppJsBareHandlerScopeRuleRejectsCrossIIFERegistration 은 역방향이다 — AC-AIG-002.
//
// 실제로 일어났던 결함(둘째 IIFE 에 선언된 식별자를 첫째 IIFE 에서 등록)을 메모리
// 안에서 합성해, 규칙이 그것을 **식별자와 줄 번호를 지목하며** 거부하는지 본다.
// 디스크의 app.js 는 읽기만 하고 한 바이트도 바뀌지 않는다 (REQ-AIG-010).
func TestAppJsBareHandlerScopeRuleRejectsCrossIIFERegistration(t *testing.T) {
	t.Parallel()
	src := readEmbeddedAsset(t, "app.js")

	_, forwardChecked := scanBareHandlerScope(src)

	const target = "stampRefreshed"
	mutant := synthesizeCrossIIFEMutant(t, src, target)

	violations, mutantChecked := scanBareHandlerScope(mutant)

	if len(violations) != 1 {
		t.Fatalf("mutant should produce exactly 1 violation, got %d: %v", len(violations), violations)
	}

	t.Logf("mutant: checked = %d (forward %d), violations = %d: %v",
		mutantChecked, forwardChecked, len(violations), violations)

	v := violations[0]
	if v.name != target {
		t.Errorf("violation must name the identifier: got %q, want %q", v.name, target)
	}
	if v.line <= 0 {
		t.Errorf("violation must carry the registration line number, got %d", v.line)
	}
	// REQ-AIG-003 — 렌더된 메시지가 두 좌표를 모두 담아야 한다. 줄 번호 없는 실패
	// 메시지는 수리자를 파일 전체 탐색으로 돌려보낸다.
	msg := v.String()
	if !strings.Contains(msg, target) || !strings.Contains(msg, fmt.Sprintf("%d", v.line)) {
		t.Errorf("violation message must contain BOTH the identifier and the registration line: %q", msg)
	}

	// [HARD] 양성 대조. 위반이 0→1 로 오른 것만으로는 "규칙이 이 결함을 잡는다" 가
	// 서지 않는다 — 검사 대수가 달라졌다면 돌연변이가 규칙이 보는 표면 자체를 바꾼
	// 것이고, 그러면 보인 것은 "입력이 달라졌다" 뿐이다.
	if mutantChecked != forwardChecked {
		t.Errorf("checked count changed between forward and mutant runs (%d → %d) — the mutant "+
			"altered the surface the rule inspects, so a violation delta proves nothing about the rule",
			forwardChecked, mutantChecked)
	}
}

// synthesizeCrossIIFEMutant 은 `name` 을 핸들러로 넘기는 등록 줄을 찾아 지우고,
// **첫째 IIFE 안**(그 IIFE 의 닫는 줄 바로 앞)에 다시 끼워 넣는다. 골든 파일로
// 돌연변이 사본을 커밋하지 않는 이유는 원본이 움직이면 사본이 즉시 스테일해지고
// 그 스테일을 아무도 알려 주지 않기 때문이다.
//
// [HARD] 전제 단언. 대상 줄이나 삽입 위치를 찾지 못하면 즉시 t.Fatal 한다. 이
// 단언이 없으면 합성 실패가 돌연변이 = 원본 을 낳고, 역방향 테스트가 "위반 0" 을
// 보며 **조용히 뒤집힌다** — 가드가 아무것도 지키지 않는데 초록으로 남는다.
func synthesizeCrossIIFEMutant(t *testing.T, src, name string) string {
	t.Helper()
	lines := strings.Split(src, "\n")

	spans := scanTopLevelIIFESpans(lines)
	if len(spans) < 2 {
		t.Fatalf("mutation synthesis precondition failed: need at least 2 top-level IIFEs, found %d", len(spans))
	}

	targetIdx := -1
	for _, reg := range findBareHandlerRegs(lines) {
		if reg.name == name {
			targetIdx = reg.line - 1
			break
		}
	}
	if targetIdx < 0 {
		t.Fatalf("mutation synthesis precondition failed: no bare-identifier registration passing %q found in app.js", name)
	}
	if spanIndexOf(spans, targetIdx+1) == 0 {
		t.Fatalf("mutation synthesis precondition failed: %q is already registered inside the first IIFE — "+
			"the mutant would equal the original", name)
	}

	moved := lines[targetIdx]

	// 첫째 IIFE 의 닫는 줄 바로 앞에 끼워 넣는다. 위치는 스팬 스캐너에서 **산출**한
	// 값이지 박아 둔 줄 번호가 아니다 (REQ-AIG-001).
	insertAt := spans[0].end - 1 // 0-기반, 닫는 줄의 인덱스

	out := make([]string, 0, len(lines))
	for i, ln := range lines {
		if i == targetIdx {
			continue // 원래 자리에서 제거
		}
		if i == insertAt {
			out = append(out, moved)
		}
		out = append(out, ln)
	}

	if strings.Join(out, "\n") == src {
		t.Fatal("mutation synthesis produced a byte-identical copy — the mutant would prove nothing")
	}
	return strings.Join(out, "\n")
}
