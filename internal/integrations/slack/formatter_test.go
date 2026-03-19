package slack

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/integrations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- FormatMessage 테스트 ---

func TestFormatMessage_AllEventTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		eventType     string
		wantEmojiRune string // 이모지를 문자열로 표현
	}{
		{
			name:          "spec_complete 이벤트",
			eventType:     "spec_complete",
			wantEmojiRune: "\u2705", // ✅
		},
		{
			name:          "quality_failure 이벤트",
			eventType:     "quality_failure",
			wantEmojiRune: "\u274c", // ❌
		},
		{
			name:          "pr_created 이벤트",
			eventType:     "pr_created",
			wantEmojiRune: "\U0001f500", // 🔀
		},
		{
			name:          "budget_alert 이벤트",
			eventType:     "budget_alert",
			wantEmojiRune: "\u26a0\ufe0f", // ⚠️
		},
		{
			name:          "알 수 없는 이벤트 타입 — 기본 이모지",
			eventType:     "unknown_type",
			wantEmojiRune: "\u2139\ufe0f", // ℹ️
		},
		{
			name:          "빈 이벤트 타입 — 기본 이모지",
			eventType:     "",
			wantEmojiRune: "\u2139\ufe0f",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			event := integrations.NotifyEvent{
				Type:    tc.eventType,
				Title:   "테스트 제목",
				Message: "테스트 메시지",
			}

			msg := FormatMessage(event)

			// Text 필드에 이모지와 제목이 포함되어야 한다.
			assert.True(t, strings.HasPrefix(msg.Text, tc.wantEmojiRune),
				"Text가 이모지로 시작해야 한다: got %q", msg.Text)
			assert.Contains(t, msg.Text, event.Title)

			// 블록이 최소 2개(헤더 + 본문)여야 한다.
			require.GreaterOrEqual(t, len(msg.Blocks), 2)

			// 첫 번째 블록은 헤더이어야 한다.
			assert.Equal(t, "header", msg.Blocks[0].Type)
			require.NotNil(t, msg.Blocks[0].Text)
			assert.Equal(t, "plain_text", msg.Blocks[0].Text.Type)
			assert.Contains(t, msg.Blocks[0].Text.Text, tc.wantEmojiRune)

			// 두 번째 블록은 섹션이어야 한다.
			assert.Equal(t, "section", msg.Blocks[1].Type)
			require.NotNil(t, msg.Blocks[1].Text)
			assert.Equal(t, "mrkdwn", msg.Blocks[1].Text.Type)
			assert.Equal(t, event.Message, msg.Blocks[1].Text.Text)
		})
	}
}

func TestFormatMessage_WithDetails(t *testing.T) {
	t.Parallel()

	event := integrations.NotifyEvent{
		Type:    "quality_failure",
		Title:   "품질 실패",
		Message: "커버리지 미달",
		Details: map[string]string{
			"coverage": "72%",
		},
	}

	msg := FormatMessage(event)

	// Details가 있으면 블록이 3개이어야 한다.
	require.Len(t, msg.Blocks, 3)

	detailBlock := msg.Blocks[2]
	assert.Equal(t, "section", detailBlock.Type)
	require.NotNil(t, detailBlock.Text)
	assert.Equal(t, "mrkdwn", detailBlock.Text.Type)

	// Details 내용이 포함되어야 한다.
	assert.Contains(t, detailBlock.Text.Text, "coverage")
	assert.Contains(t, detailBlock.Text.Text, "72%")
}

func TestFormatMessage_EmptyDetails(t *testing.T) {
	t.Parallel()

	event := integrations.NotifyEvent{
		Type:    "spec_complete",
		Title:   "SPEC 완료",
		Message: "성공적으로 완료",
		Details: map[string]string{}, // 빈 맵
	}

	msg := FormatMessage(event)

	// 빈 Details는 추가 블록을 생성하지 않아야 한다.
	assert.Len(t, msg.Blocks, 2)
}

func TestFormatMessage_NilDetails(t *testing.T) {
	t.Parallel()

	event := integrations.NotifyEvent{
		Type:    "spec_complete",
		Title:   "SPEC 완료",
		Message: "성공",
		Details: nil,
	}

	msg := FormatMessage(event)
	assert.Len(t, msg.Blocks, 2)
}

func TestFormatMessage_SpecialCharacters(t *testing.T) {
	t.Parallel()

	event := integrations.NotifyEvent{
		Type:    "spec_complete",
		Title:   `SPEC "완료" & <특수문자>`,
		Message: "메시지에 `코드 블록` 포함\n줄 바꿈도 포함",
		Details: map[string]string{
			"path": "/usr/local/bin <test>",
		},
	}

	msg := FormatMessage(event)

	// 특수문자가 포함된 내용이 그대로 전달되어야 한다.
	assert.Contains(t, msg.Text, event.Title)
	assert.Contains(t, msg.Blocks[1].Text.Text, event.Message)
	assert.Contains(t, msg.Blocks[2].Text.Text, "/usr/local/bin <test>")
}

func TestFormatMessage_MultipleDetails(t *testing.T) {
	t.Parallel()

	event := integrations.NotifyEvent{
		Type:    "budget_alert",
		Title:   "예산 경고",
		Message: "토큰 사용량이 한도를 초과했습니다.",
		Details: map[string]string{
			"used":  "195000",
			"limit": "200000",
			"model": "claude-sonnet",
		},
	}

	msg := FormatMessage(event)
	require.Len(t, msg.Blocks, 3)

	detailText := msg.Blocks[2].Text.Text
	// 모든 키-값 쌍이 포함되어야 한다.
	assert.Contains(t, detailText, "used")
	assert.Contains(t, detailText, "195000")
	assert.Contains(t, detailText, "limit")
	assert.Contains(t, detailText, "200000")
	assert.Contains(t, detailText, "model")
	assert.Contains(t, detailText, "claude-sonnet")

	// 각 항목은 볼드체 마크다운으로 감싸져야 한다.
	assert.Contains(t, detailText, "*")
}

func TestFormatMessage_HeaderContainsTitle(t *testing.T) {
	t.Parallel()

	title := "중요한 알림 제목"
	event := integrations.NotifyEvent{
		Type:    "pr_created",
		Title:   title,
		Message: "PR이 생성되었습니다.",
	}

	msg := FormatMessage(event)

	// Text와 헤더 블록 모두에 제목이 있어야 한다.
	assert.Contains(t, msg.Text, title)
	assert.Contains(t, msg.Blocks[0].Text.Text, title)
}
