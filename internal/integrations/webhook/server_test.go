package webhook

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// computeSignature는 테스트용 HMAC-SHA256 서명을 계산한다.
func computeSignature(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

// --- verifySignature 단위 테스트 ---

func TestVerifySignature(t *testing.T) {
	t.Parallel()

	secret := "test-secret"
	payload := []byte(`{"key":"value"}`)
	validSig := computeSignature(payload, secret)

	tests := []struct {
		name      string
		payload   []byte
		signature string
		secret    string
		want      bool
	}{
		{
			name:      "올바른 HMAC 서명",
			payload:   payload,
			signature: validSig,
			secret:    secret,
			want:      true,
		},
		{
			name:      "잘못된 서명 값",
			payload:   payload,
			signature: "sha256=deadbeef",
			secret:    secret,
			want:      false,
		},
		{
			name:      "빈 서명",
			payload:   payload,
			signature: "",
			secret:    secret,
			want:      false,
		},
		{
			name:      "다른 시크릿으로 계산된 서명",
			payload:   payload,
			signature: computeSignature(payload, "wrong-secret"),
			secret:    secret,
			want:      false,
		},
		{
			name:      "페이로드가 다를 때",
			payload:   []byte(`{"key":"other"}`),
			signature: validSig,
			secret:    secret,
			want:      false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := verifySignature(tc.payload, tc.signature, tc.secret)
			assert.Equal(t, tc.want, got)
		})
	}
}

// --- handleWebhook HTTP 핸들러 테스트 ---

func TestHandleWebhook_MethodEnforcement(t *testing.T) {
	t.Parallel()

	s := NewServer(8080, "WEBHOOK_SECRET", nil)

	tests := []struct {
		name       string
		method     string
		wantStatus int
	}{
		{"POST 허용", http.MethodPost, http.StatusOK},
		{"GET 거부", http.MethodGet, http.StatusMethodNotAllowed},
		{"PUT 거부", http.MethodPut, http.StatusMethodNotAllowed},
		{"DELETE 거부", http.MethodDelete, http.StatusMethodNotAllowed},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			body := []byte(`{"event":"test"}`)
			req := httptest.NewRequest(tc.method, "/webhook", bytes.NewReader(body))
			w := httptest.NewRecorder()

			s.handleWebhook(w, req)
			assert.Equal(t, tc.wantStatus, w.Code)
		})
	}
}

// TestHandleWebhook_SignatureVerification은 서명 검증 시나리오를 테이블 드리븐으로 검증한다.
// t.Setenv를 사용하므로 서브테스트에서 t.Parallel()을 사용하지 않는다.
func TestHandleWebhook_SignatureVerification(t *testing.T) {
	const secretEnv = "WEBHOOK_SECRET_SIG_TEST"
	const secretVal = "super-secret"

	body := []byte(`{"event":"spec_complete"}`)
	validSig := computeSignature(body, secretVal)

	tests := []struct {
		name       string
		signature  string
		secretSet  bool
		wantStatus int
	}{
		{
			name:       "올바른 서명 — 인증 성공",
			signature:  validSig,
			secretSet:  true,
			wantStatus: http.StatusOK,
		},
		{
			name:       "잘못된 서명 — 401 반환",
			signature:  "sha256=invalid",
			secretSet:  true,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "서명 누락 — 401 반환",
			signature:  "",
			secretSet:  true,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "시크릿 미설정 — 서명 없이 통과",
			signature:  "",
			secretSet:  false,
			wantStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		// t.Setenv를 사용하므로 서브테스트에서 t.Parallel() 호출 금지
		t.Run(tc.name, func(t *testing.T) {
			if tc.secretSet {
				t.Setenv(secretEnv, secretVal)
			} else {
				t.Setenv(secretEnv, "")
			}

			s := NewServer(8080, secretEnv, nil)

			req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
			if tc.signature != "" {
				req.Header.Set("X-Signature-256", tc.signature)
			}
			w := httptest.NewRecorder()

			s.handleWebhook(w, req)
			assert.Equal(t, tc.wantStatus, w.Code)
		})
	}
}

func TestHandleWebhook_InvalidJSON(t *testing.T) {
	// t.Setenv 사용으로 t.Parallel() 생략
	t.Setenv("WEBHOOK_SECRET_JSON", "")
	s := NewServer(8080, "WEBHOOK_SECRET_JSON", nil)

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader([]byte(`not-json`)))
	w := httptest.NewRecorder()

	s.handleWebhook(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleWebhook_NilHandler(t *testing.T) {
	// handler가 nil이어도 패닉 없이 200 반환해야 한다.
	t.Setenv("WEBHOOK_SECRET_NIL", "")
	s := NewServer(8080, "WEBHOOK_SECRET_NIL", nil)

	body := []byte(`{"event":"ping"}`)
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	w := httptest.NewRecorder()

	s.handleWebhook(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleWebhook_HandlerError(t *testing.T) {
	// 핸들러가 에러를 반환하면 500을 응답해야 한다.
	t.Setenv("WEBHOOK_SECRET_ERR", "")

	errHandler := EventHandler(func(_ map[string]any) error {
		return assert.AnError
	})
	s := NewServer(8080, "WEBHOOK_SECRET_ERR", errHandler)

	body := []byte(`{"event":"fail"}`)
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	w := httptest.NewRecorder()

	s.handleWebhook(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleWebhook_HandlerReceivesEvent(t *testing.T) {
	t.Setenv("WEBHOOK_SECRET_EVT", "")

	var received map[string]any
	handler := EventHandler(func(event map[string]any) error {
		received = event
		return nil
	})
	s := NewServer(8080, "WEBHOOK_SECRET_EVT", handler)

	body := []byte(`{"type":"spec_complete","spec_id":"SPEC-001"}`)
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	w := httptest.NewRecorder()

	s.handleWebhook(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "spec_complete", received["type"])
	assert.Equal(t, "SPEC-001", received["spec_id"])
}

func TestHandleWebhook_BodySizeLimit(t *testing.T) {
	t.Setenv("WEBHOOK_SECRET_SIZE", "")

	// 1MB + 1바이트: LimitReader가 잘라내어 JSON 파싱이 실패해야 한다.
	oversized := make([]byte, 1<<20+1)
	// 잘라낸 뒤 JSON 파싱 시 실패하도록 유효하지 않은 바이트로 채운다.
	for i := range oversized {
		oversized[i] = 'x'
	}

	s := NewServer(8080, "WEBHOOK_SECRET_SIZE", nil)
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(oversized))
	w := httptest.NewRecorder()

	s.handleWebhook(w, req)
	// 잘린 본문은 유효한 JSON이 아니므로 400을 반환해야 한다.
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleWebhook_ResponseBody(t *testing.T) {
	t.Setenv("WEBHOOK_SECRET_RESP", "")
	s := NewServer(8080, "WEBHOOK_SECRET_RESP", nil)

	body := []byte(`{"ping":true}`)
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	w := httptest.NewRecorder()

	s.handleWebhook(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "ok", resp["status"])
}

// --- handleHealth 테스트 ---

func TestHandleHealth(t *testing.T) {
	t.Parallel()

	s := NewServer(8080, "", nil)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	s.handleHealth(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "healthy", resp["status"])
}

// --- Stop 테스트 (srv == nil 방어 코드) ---

func TestStop_NilServer(t *testing.T) {
	t.Parallel()

	s := NewServer(8080, "", nil)
	// Start를 호출하지 않았으므로 srv가 nil — 패닉 없이 nil을 반환해야 한다.
	err := s.Stop(t.Context())
	assert.NoError(t, err)
}

// --- 통합 테스트: httptest.NewServer 활용 ---

func TestServer_Integration(t *testing.T) {
	const secretEnv = "WEBHOOK_SECRET_INTEGRATION"
	const secretVal = "integration-secret"
	t.Setenv(secretEnv, secretVal)

	var capturedEvent map[string]any
	s := NewServer(0, secretEnv, func(event map[string]any) error {
		capturedEvent = event
		return nil
	})

	// httptest 서버를 직접 생성해 실제 HTTP 요청을 보낸다.
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", s.handleWebhook)
	mux.HandleFunc("/health", s.handleHealth)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	// 헬스 체크
	resp, err := http.Get(ts.URL + "/health")
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// 유효한 웹훅 전송
	payload := []byte(`{"type":"pr_created","spec_id":"SPEC-042"}`)
	sig := computeSignature(payload, secretVal)

	req, err := http.NewRequest(http.MethodPost, ts.URL+"/webhook", bytes.NewReader(payload))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Signature-256", sig)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "pr_created", capturedEvent["type"])
}
