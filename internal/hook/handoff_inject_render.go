package hook

// handoff_inject_render.go renders the degrade-to-guidance additionalContext for
// the auto-resume injection (SPEC-HANDOFF-AUTORESUME-001 REQ-011/015).
//
// The injected text NEVER asserts that ultrathink / an extended-reasoning mode is
// active (verification-claim-integrity §1.1) — a hook cannot change effort/model.
// Mode-change directives are rendered as manual-paste restoration guidance only.
// The header and guidance are localized to the user's conversation_language
// (ko/en/ja/zh, English fallback for other locales); the body is appended verbatim.

import (
	"strings"

	"github.com/modu-ai/moai-adk/internal/hook/handoff"
)

// handoffStrings carries the localized rendering strings for one locale.
type handoffStrings struct {
	header        string
	disclaimer    string
	restoreIntro  string
	ultrathink    string
	ultracode     string
	goalPrefix    string
	goalSuffix    string
	bodySeparator string
}

// renderHandoffContext builds the additionalContext for a consumed pending record.
func renderHandoffContext(rec *handoff.PendingRecord) string {
	s := handoffLocaleStrings(rec.ConversationLanguage)

	var b strings.Builder
	b.WriteString(s.header)
	b.WriteByte('\n')
	b.WriteString(s.disclaimer)
	b.WriteByte('\n')

	if rec.Directives.Ultrathink || rec.Directives.Ultracode || rec.Directives.Goal != "" {
		b.WriteByte('\n')
		b.WriteString(s.restoreIntro)
		b.WriteByte('\n')
		if rec.Directives.Ultrathink {
			b.WriteString(s.ultrathink)
			b.WriteByte('\n')
		}
		if rec.Directives.Ultracode {
			b.WriteString(s.ultracode)
			b.WriteByte('\n')
		}
		if rec.Directives.Goal != "" {
			b.WriteString(s.goalPrefix)
			b.WriteString(rec.Directives.Goal)
			b.WriteString(s.goalSuffix)
			b.WriteByte('\n')
		}
	}

	b.WriteByte('\n')
	b.WriteString(s.bodySeparator)
	b.WriteByte('\n')
	b.WriteString(rec.Body) // verbatim — degrade applies to header/guidance only
	return b.String()
}

// handoffLocale normalizes a conversation_language to one of {ko, en, ja, zh};
// any other ISO-639 code falls back to en (session-handoff.md Localization rule).
func handoffLocale(lang string) string {
	switch strings.ToLower(strings.TrimSpace(lang)) {
	case "ko", "en", "ja", "zh":
		return strings.ToLower(strings.TrimSpace(lang))
	default:
		return "en"
	}
}

// handoffLocaleStrings returns the localized strings for lang (en fallback).
func handoffLocaleStrings(lang string) handoffStrings {
	switch handoffLocale(lang) {
	case "ko":
		return handoffStrings{
			header:        "[MoAI 자동 재개 — 이전 세션 핸드오프]",
			disclaimer:    "아래는 이전 세션이 저장한 재개 컨텍스트입니다. 이 주입은 컨텍스트 전달일 뿐이며, 확장 추론 모드를 자동으로 활성화하지 않습니다. 필요하면 아래 안내 줄을 직접 입력하세요.",
			restoreIntro:  "복원 안내(수동 입력):",
			ultrathink:    "  • ultrathink   ← 확장 추론을 복원하려면 이 줄을 입력하세요",
			ultracode:     "  • /effort ultracode ← 워크플로우 팬아웃을 복원하려면 이 줄을 입력하세요",
			goalPrefix:    "  • /goal ",
			goalSuffix:    " ← 자율 continuation을 복원하려면 이 줄을 입력하세요",
			bodySeparator: "── 저장된 재개 메시지(참고) ──",
		}
	case "ja":
		return handoffStrings{
			header:        "[MoAI 自動再開 — 前セッションのハンドオフ]",
			disclaimer:    "以下は前セッションが保存した再開コンテキストです。この注入はコンテキストの受け渡しのみで、拡張推論モードを自動的に有効化しません。必要に応じて下記の案内行を手動で入力してください。",
			restoreIntro:  "復元案内（手動入力）:",
			ultrathink:    "  • ultrathink   ← 拡張推論を復元するにはこの行を入力してください",
			ultracode:     "  • /effort ultracode ← ワークフローのファンアウトを復元するにはこの行を入力してください",
			goalPrefix:    "  • /goal ",
			goalSuffix:    " ← 自律継続を復元するにはこの行を入力してください",
			bodySeparator: "── 保存された再開メッセージ（参考）──",
		}
	case "zh":
		return handoffStrings{
			header:        "[MoAI 自动恢复 — 上一会话交接]",
			disclaimer:    "以下是上一会话保存的恢复上下文。此注入仅传递上下文，不会自动启用任何扩展推理模式。如有需要，请手动输入下方的指引行。",
			restoreIntro:  "恢复指引（手动输入）:",
			ultrathink:    "  • ultrathink   ← 输入此行以恢复扩展推理",
			ultracode:     "  • /effort ultracode ← 输入此行以恢复工作流扇出",
			goalPrefix:    "  • /goal ",
			goalSuffix:    " ← 输入此行以恢复自主延续",
			bodySeparator: "── 已保存的恢复消息（参考）──",
		}
	default: // en
		return handoffStrings{
			header:        "[MoAI auto-resume — previous session handoff]",
			disclaimer:    "The following is the resume context saved by the previous session. This injection only delivers context; it does not automatically enable any extended-reasoning mode. Enter the guidance lines below manually if you need them.",
			restoreIntro:  "Restoration guidance (enter manually):",
			ultrathink:    "  • ultrathink   ← enter this line to restore extended reasoning",
			ultracode:     "  • /effort ultracode ← enter this line to restore workflow fan-out",
			goalPrefix:    "  • /goal ",
			goalSuffix:    " ← enter this line to restore autonomous continuation",
			bodySeparator: "── saved resume message (reference) ──",
		}
	}
}
