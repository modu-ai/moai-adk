package slack

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/modu-ai/moai-adk/internal/integrations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- shouldNotify 단위 테스트 ---

func TestShouldNotify(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		events    []string
		eventType string
		want      bool
	}{
		{
			name:      "test 이벤트는 항상 허용",
			events:    []string{},
			eventType: "test",
			want:      true,
		},
		{
			name:      "등록된 이벤트 타입 허용",
			events:    []string{"spec_complete", "quality_failure"},
			eventType: "spec_complete",
			want:      true,
		},
		{
			name:      "등록되지 않은 이벤트 타입 차단",
			events:    []string{"spec_complete"},
			eventType: "pr_created",
			want:      false,
		},
		{
			name:      "이벤트 목록이 비어 있고 test가 아닌 경우 차단",
			events:    []string{},
			eventType: "budget_alert",
			want:      false,
		},
		{
			name:      "여러 이벤트 중 마지막 항목 매칭",
			events:    []string{"spec_complete", "quality_failure", "budget_alert"},
			eventType: "budget_alert",
			want:      true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := NewClient("https://hooks.slack.com/test", tc.events)
			got := c.shouldNotify(tc.eventType)
			assert.Equal(t, tc.want, got)
		})
	}
}

// --- IsEnabled 테스트 ---

func TestIsEnabled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		webhookURL string
		want       bool
	}{
		{
			name:       "URL이 설정되면 활성화",
			webhookURL: "https://hooks.slack.com/services/T000/B000/xxx",
			want:       true,
		},
		{
			name:       "빈 URL이면 비활성화",
			webhookURL: "",
			want:       false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := NewClient(tc.webhookURL, nil)
			assert.Equal(t, tc.want, c.IsEnabled())
		})
	}
}

// --- Name 테스트 ---

func TestName(t *testing.T) {
	t.Parallel()
	c := NewClient("https://hooks.slack.com/test", nil)
	assert.Equal(t, "Slack", c.Name())
}

// --- Send 테스트 (httptest.NewServer 사용) ---

func TestSend_Success(t *testing.T) {
	t.Parallel()

	// Slack 웹훅을 흉내 내는 목 서버
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// 요청 바디가 유효한 JSON인지 확인
		var payload SlackMessage
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		assert.NotEmpty(t, payload.Text)

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := NewClient(ts.URL, []string{"spec_complete"})
	err := c.Send(integrations.NotifyEvent{
		Type:    "spec_complete",
		Title:   "SPEC-001 완료",
		Message: "모든 품질 게이트를 통과했습니다.",
	})
	assert.NoError(t, err)
}

func TestSend_SkippedWhenNotInEventList(t *testing.T) {
	t.Parallel()

	// 목 서버 호출 여부 추적
	called := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := NewClient(ts.URL, []string{"spec_complete"})
	err := c.Send(integrations.NotifyEvent{
		Type:    "budget_alert", // 등록되지 않은 이벤트
		Title:   "예산 경고",
		Message: "토큰 예산 초과",
	})
	assert.NoError(t, err)
	assert.False(t, called, "등록되지 않은 이벤트 타입에는 요청이 전송되어선 안 된다")
}

func TestSend_HTTPError(t *testing.T) {
	t.Parallel()

	// Slack 서버가 비정상 응답을 반환하는 경우
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer ts.Close()

	c := NewClient(ts.URL, []string{"spec_complete"})
	err := c.Send(integrations.NotifyEvent{
		Type:    "spec_complete",
		Title:   "테스트",
		Message: "메시지",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "503")
}

func TestSend_WithDetails(t *testing.T) {
	t.Parallel()

	var receivedPayload SlackMessage
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(&receivedPayload))
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := NewClient(ts.URL, []string{"quality_failure"})
	err := c.Send(integrations.NotifyEvent{
		Type:    "quality_failure",
		Title:   "품질 실패",
		Message: "커버리지가 기준 미달입니다.",
		Details: map[string]string{
			"coverage": "72%",
			"required": "85%",
		},
	})
	require.NoError(t, err)

	// Details가 있으면 블록이 3개 이상이어야 한다.
	assert.GreaterOrEqual(t, len(receivedPayload.Blocks), 3)
}

// --- Test (연결 확인) 테스트 ---

func TestTest_SendsTestEvent(t *testing.T) {
	t.Parallel()

	var receivedPayload SlackMessage
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(&receivedPayload))
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := NewClient(ts.URL, nil)
	err := c.Test()
	require.NoError(t, err)
	assert.Contains(t, receivedPayload.Text, "MoAI-ADK Test")
}
