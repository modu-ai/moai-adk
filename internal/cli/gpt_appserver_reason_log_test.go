package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestManagedGPTScopeReasonLogIsAllowlisted(t *testing.T) {
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	logger, err := newManagedGPTRejectionLogger(dir, "reason-fixture")
	if err != nil {
		t.Fatal(err)
	}
	for _, reason := range []string{"unexpected_tool_result", "CANARY-PRIVATE"} {
		logger.RecordGatewayRejection(map[string]string{"cause": "appserver_scope_mismatch", "route": "gpt-5.6-sol", "digest": strings.Repeat("a", 64), "reason": reason})
	}
	raw, err := os.ReadFile(filepath.Join(dir, logger.name))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"reason":"unexpected_tool_result"`) || strings.Contains(string(raw), "CANARY") {
		t.Fatalf("unsafe or missing reason: %s", raw)
	}
}
