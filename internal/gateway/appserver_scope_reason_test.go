package gateway

import (
	"fmt"
	"testing"

	"github.com/modu-ai/moai-adk/internal/codexbridge"
)

func TestAppServerScopeReasonNeverIncludesExceptionText(t *testing.T) {
	for _, tc := range []struct {
		err  error
		want string
	}{
		{fmt.Errorf("%w: unexpected tool result", codexbridge.ErrScope), "unexpected_tool_result"},
		{fmt.Errorf("%w: active model mismatch", codexbridge.ErrScope), "active_model_mismatch"},
		{fmt.Errorf("%w: tool result count mismatch", codexbridge.ErrScope), "tool_result_count_mismatch"},
		{fmt.Errorf("%w: CANARY-PRIVATE", codexbridge.ErrScope), "unclassified_scope"},
		{fmt.Errorf("CANARY-PRIVATE"), ""},
	} {
		if got := appServerScopeReason(tc.err); got != tc.want {
			t.Fatalf("got=%q want=%q", got, tc.want)
		}
	}
}
