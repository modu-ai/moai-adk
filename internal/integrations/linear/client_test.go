package linear

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/integrations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newClientWithURL은 테스트 전용 헬퍼로, httptest.NewServer URL을 baseURL로 사용하는 Client를 반환한다.
func newClientWithURL(url string) *Client {
	return &Client{
		apiKeyEnv:  "LINEAR_API_KEY_UNUSED",
		teamID:     "TEAM-DEFAULT",
		httpClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    url,
	}
}

// --- IsEnabled 테스트 ---

// TestIsEnabled는 t.Setenv를 사용하므로 t.Parallel()을 사용하지 않는다.
func TestIsEnabled(t *testing.T) {
	tests := []struct {
		name      string
		apiKeyVal string
		teamID    string
		want      bool
	}{
		{
			name:      "API 키와 팀 ID 모두 설정 — 활성화",
			apiKeyVal: "lin_api_test_key",
			teamID:    "TEAM-001",
			want:      true,
		},
		{
			name:      "API 키 없음 — 비활성화",
			apiKeyVal: "",
			teamID:    "TEAM-001",
			want:      false,
		},
		{
			name:      "팀 ID 없음 — 비활성화",
			apiKeyVal: "lin_api_test_key",
			teamID:    "",
			want:      false,
		},
		{
			name:      "API 키와 팀 ID 모두 없음 — 비활성화",
			apiKeyVal: "",
			teamID:    "",
			want:      false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			const envKey = "LINEAR_API_KEY_ISENABLED"
			t.Setenv(envKey, tc.apiKeyVal)

			c := NewClient(envKey, tc.teamID)
			assert.Equal(t, tc.want, c.IsEnabled())
		})
	}
}

// --- Name 테스트 ---

func TestName(t *testing.T) {
	t.Parallel()
	c := NewClient("LINEAR_API_KEY", "TEAM-001")
	assert.Equal(t, "Linear", c.Name())
}

// --- NewClient baseURL 기본값 테스트 ---

func TestNewClient_DefaultBaseURL(t *testing.T) {
	t.Parallel()
	c := NewClient("LINEAR_KEY", "TEAM-001")
	assert.Equal(t, linearAPIURL, c.baseURL)
}

// --- graphQL 테스트 (httptest.NewServer 활용) ---

func TestGraphQL_Success(t *testing.T) {
	t.Parallel()

	expectedResponse := map[string]any{
		"data": map[string]any{
			"viewer": map[string]any{"id": "user-1", "name": "TestUser"},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Content-Type과 Authorization 헤더 확인
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.NotEmpty(t, r.Header.Get("Authorization"))
		assert.Equal(t, http.MethodPost, r.Method)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expectedResponse)
	}))
	defer ts.Close()

	c := newClientWithURL(ts.URL)
	result, err := c.graphQL("test-api-key", `{"query":"{ viewer { id name } }"}`)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestGraphQL_HTTPError(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()

	c := newClientWithURL(ts.URL)
	_, err := c.graphQL("invalid-key", `{"query":"{ viewer { id } }"}`)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "401")
}

func TestGraphQL_InvalidJSONResponse(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not-json"))
	}))
	defer ts.Close()

	c := newClientWithURL(ts.URL)
	_, err := c.graphQL("test-key", `{"query":"{ viewer { id } }"}`)
	require.Error(t, err)
}

// --- GraphQL 파라미터화 테스트 (인젝션 방지) ---

func TestCreateIssue_UsesParameterizedGraphQL(t *testing.T) {
	t.Parallel()

	// 서버에서 수신한 요청 바디를 파싱하여 변수가 올바르게 전달되는지 확인한다.
	var receivedRequest graphQLRequest
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(&receivedRequest))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":{"issueCreate":{"success":true}}}`))
	}))
	defer ts.Close()

	c := newClientWithURL(ts.URL)
	c.teamID = "TEAM-INJECT"

	event := integrations.NotifyEvent{
		Type:    "spec_complete",
		Title:   `테스트 <injection> "test"`,
		Message: "설명\n줄 바꿈 포함",
	}

	err := c.createIssue("lin_api_create_key", event)
	require.NoError(t, err)

	// Variables 맵을 통해 값이 전달되어야 한다 (쿼리 문자열에 직접 삽입되지 않아야 한다).
	assert.Equal(t, "TEAM-INJECT", receivedRequest.Variables["teamId"])
	assert.Equal(t, event.Title, receivedRequest.Variables["title"])
	assert.Equal(t, event.Message, receivedRequest.Variables["description"])

	// 쿼리 자체에 사용자 입력이 삽입되어선 안 된다.
	assert.NotContains(t, receivedRequest.Query, event.Title)
	assert.NotContains(t, receivedRequest.Query, event.Message)
}

func TestCreateIssue_GraphQLRequestStructure(t *testing.T) {
	t.Parallel()

	// Variables 맵의 키 목록을 확인한다.
	var receivedRequest graphQLRequest
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(&receivedRequest))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":{"issueCreate":{"success":true}}}`))
	}))
	defer ts.Close()

	c := newClientWithURL(ts.URL)
	c.teamID = "TEAM-001"

	event := integrations.NotifyEvent{
		Type:    "spec_complete",
		Title:   "SPEC-001",
		Message: "완료",
	}

	err := c.createIssue("api-key", event)
	require.NoError(t, err)

	// 필수 변수 키가 모두 존재해야 한다.
	expectedKeys := []string{"teamId", "title", "description"}
	for _, key := range expectedKeys {
		_, exists := receivedRequest.Variables[key]
		assert.True(t, exists, "Variables에 %q 키가 있어야 한다", key)
	}
}

// --- Send 테스트 ---

// TestSend_SpecComplete_CreatesIssue는 t.Setenv를 사용하므로 t.Parallel()을 생략한다.
func TestSend_SpecComplete_CreatesIssue(t *testing.T) {
	called := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":{"issueCreate":{"success":true}}}`))
	}))
	defer ts.Close()

	const envKey = "LINEAR_SEND_KEY"
	t.Setenv(envKey, "lin_api_send_key")

	c := newClientWithURL(ts.URL)
	c.apiKeyEnv = envKey
	c.teamID = "TEAM-001"

	err := c.Send(integrations.NotifyEvent{
		Type:    "spec_complete",
		Title:   "SPEC-001 완료",
		Message: "구현 완료",
	})
	require.NoError(t, err)
	assert.True(t, called, "spec_complete 이벤트는 GraphQL 호출을 발생시켜야 한다")
}

// TestSend_NoAPIKey_ReturnsError는 t.Setenv를 사용하므로 t.Parallel()을 생략한다.
func TestSend_NoAPIKey_ReturnsError(t *testing.T) {
	const envKey = "LINEAR_EMPTY_KEY"
	t.Setenv(envKey, "") // 키 미설정

	c := NewClient(envKey, "TEAM-001")
	err := c.Send(integrations.NotifyEvent{
		Type:    "spec_complete",
		Title:   "테스트",
		Message: "메시지",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), envKey)
}

// TestSend_UnhandledEventType_ReturnsNil는 t.Setenv를 사용하므로 t.Parallel()을 생략한다.
func TestSend_UnhandledEventType_ReturnsNil(t *testing.T) {
	const envKey = "LINEAR_UNHANDLED_KEY"
	t.Setenv(envKey, "lin_api_key")

	c := NewClient(envKey, "TEAM-001")
	// "budget_alert" 같은 미등록 이벤트는 nil을 반환해야 한다.
	err := c.Send(integrations.NotifyEvent{
		Type:    "budget_alert",
		Title:   "예산 경고",
		Message: "토큰 한도 초과",
	})
	assert.NoError(t, err)
}

// TestSend_QualityFailureWithEmptySpecID는 t.Setenv를 사용하므로 t.Parallel()을 생략한다.
func TestSend_QualityFailureWithEmptySpecID(t *testing.T) {
	const envKey = "LINEAR_QUALITY_KEY"
	t.Setenv(envKey, "lin_api_quality_key")

	// addComment는 SpecID가 비어 있으면 아무것도 하지 않는다.
	c := NewClient(envKey, "TEAM-001")
	err := c.Send(integrations.NotifyEvent{
		Type:    "quality_failure",
		SpecID:  "",
		Title:   "품질 실패",
		Message: "커버리지 미달",
	})
	assert.NoError(t, err)
}

// --- Test (연결 확인) 테스트 ---

// TestTest_Success는 t.Setenv를 사용하므로 t.Parallel()을 생략한다.
func TestTest_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":{"viewer":{"id":"user-1","name":"TestUser"}}}`))
	}))
	defer ts.Close()

	const envKey = "LINEAR_TEST_KEY"
	t.Setenv(envKey, "lin_api_test_key")

	c := newClientWithURL(ts.URL)
	c.apiKeyEnv = envKey
	err := c.Test()
	assert.NoError(t, err)
}

// TestTest_NoAPIKey_ReturnsError는 t.Setenv를 사용하므로 t.Parallel()을 생략한다.
func TestTest_NoAPIKey_ReturnsError(t *testing.T) {
	const envKey = "LINEAR_TEST_EMPTY"
	t.Setenv(envKey, "")

	c := NewClient(envKey, "TEAM-001")
	err := c.Test()
	require.Error(t, err)
	assert.Contains(t, err.Error(), envKey)
}

// TestTest_HTTPError는 t.Setenv를 사용하므로 t.Parallel()을 생략한다.
func TestTest_HTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()

	const envKey = "LINEAR_TEST_HTTP_ERR"
	t.Setenv(envKey, "lin_api_key")

	c := newClientWithURL(ts.URL)
	c.apiKeyEnv = envKey
	err := c.Test()
	require.Error(t, err)
}

// --- graphQLRequest 구조체 직렬화 테스트 ---

func TestGraphQLRequest_Serialization(t *testing.T) {
	t.Parallel()

	req := graphQLRequest{
		Query: "{ viewer { id } }",
		Variables: map[string]any{
			"teamId": "TEAM-001",
			"title":  "제목",
		},
	}

	data, err := json.Marshal(req)
	require.NoError(t, err)

	var decoded graphQLRequest
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, req.Query, decoded.Query)
	assert.True(t, reflect.DeepEqual(req.Variables, decoded.Variables))
}

func TestGraphQLRequest_OmitEmptyVariables(t *testing.T) {
	t.Parallel()

	req := graphQLRequest{
		Query: "{ viewer { id } }",
	}

	data, err := json.Marshal(req)
	require.NoError(t, err)

	// Variables가 nil이면 JSON에서 생략되어야 한다.
	assert.NotContains(t, string(data), "variables")
}
