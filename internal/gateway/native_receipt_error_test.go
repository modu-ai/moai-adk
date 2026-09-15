package gateway

import (
	"context"
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/gateway/translate"
)

func TestNativeReceiptDenialIsNotTransportRetry(t *testing.T) {
	_, _, err := translate.RequestContext(context.Background(), "gpt-5.6-sol", []byte(`{"max_tokens":2,"metadata":{"user_id":"{\"device_id\":\"device\",\"account_uuid\":\"\",\"session_id\":\"11111111-1111-4111-8111-111111111111\"}"},"messages":[{"role":"user","content":"hello"}]}`), translate.Limits{PolicyProfile: translate.PolicyGPTNative, NativeReceiptAuthorize: func(context.Context, []byte, translate.NativePolicy) error {
		return errors.New("private identity detail")
	}})
	if err == nil {
		t.Fatal("denial accepted")
	}
	status, _, cause := appServerError(err)
	if status != 400 || cause != "native_receipt_authorization_required" {
		t.Fatalf("status=%d cause=%s", status, cause)
	}
}
