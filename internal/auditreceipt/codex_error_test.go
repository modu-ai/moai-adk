package auditreceipt

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// A receipt that carries a codex error records it under codex_error; a receipt
// without one leaves the key out, so existing receipts keep their shape (t1788).
func TestReceiptCodexErrorSerializesOnlyWhenSet(t *testing.T) {
	withErr, err := json.Marshal(Receipt{
		Tool:         "audit_multi",
		CreatedAt:    time.Unix(0, 0).UTC(),
		CodexVerdict: "inconclusive",
		CodexError:   "codex leg timed out after 20m0s",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(withErr), `"codex_error":"codex leg timed out after 20m0s"`) {
		t.Fatalf("receipt lacks codex_error: %s", withErr)
	}

	plain, err := json.Marshal(Receipt{
		Tool:         "audit_multi",
		CreatedAt:    time.Unix(0, 0).UTC(),
		CodexVerdict: "pass",
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(plain), "codex_error") {
		t.Fatalf("a receipt without a codex error must omit the key: %s", plain)
	}
}
