// templ 이 쓰는 작은 순수 함수들. 뷰 로직만 담고 데이터 접근은 하지 않는다.
package web

import (
	"strconv"
	"strings"
	"syscall"
)

// processAlive 가 쓰는 시그널 0 (프로세스를 건드리지 않고 존재만 확인).
const syscallZero = syscall.Signal(0)

func itoa(n int) string { return strconv.Itoa(n) }

// aria-selected / aria-invalid 는 문자열 "true"/"false" 를 요구한다.
func attrBool(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// aria-current 는 참일 때만 존재해야 한다. 거짓이면 빈 문자열로 속성을 비운다.
func attrCurrent(b bool) string {
	if b {
		return "page"
	}
	return ""
}

func saveBtnVariant(state string) string {
	switch state {
	case "dirty", "error":
		return "btn--primary"
	default:
		return "btn--outline"
	}
}

func bannerRole(kind string) string {
	if kind == "warn" {
		return "alert"
	}
	return "note"
}

func stageGlyph(stage string) string {
	switch stage {
	case StageDone:
		return "▪"
	case StageActive:
		return "▶"
	case StageBlocked:
		return "!"
	default:
		return "▫"
	}
}

// stageLabel / stateLabel 은 화면에 박히는 baseline 문자열이다. 실제 표시 언어는
// data-i18n 키(stage.* / state.*)로 클라이언트가 바꿔 넣는다 — 서버가 locale 을
// 알 필요가 없는 기존 방식 그대로다.
func stageLabel(stage string) string {
	switch stage {
	case StageDone:
		return "Done"
	case StageActive:
		return "Running"
	case StageBlocked:
		return "Blocked"
	default:
		return "Waiting"
	}
}

func stateLabel(state string) string {
	switch state {
	case StateLive:
		return "Active"
	case StateIdle:
		return "Not started"
	default:
		return "No response"
	}
}

// SPEC 목록의 점은 "지금 일하고 있는가"만 채운다. 나머지는 빈 점 + 텍스트로 읽는다.
func specDotState(status string) string {
	if status == "in-progress" {
		return StateLive
	}
	return StateStale
}

func driftKind(text string) string {
	if strings.HasPrefix(text, "MUST") {
		return "danger"
	}
	return ""
}

// pct 는 0..100 으로 자른다. 음수는 "기록 없음"을 뜻하므로 그대로 흘려보낸다.
func clampPct(n int) int {
	switch {
	case n < 0:
		return -1
	case n > 100:
		return 100
	default:
		return n
	}
}
