//go:build darwin || linux

package cli

import (
	"strings"
	"testing"
)

func TestFactoryOperationalEvidenceRejectsUnownedResult(t *testing.T) {
	call := `{"type":"response_item","payload":{"type":"function_call","name":"mcp__moai__factory_msg_status","arguments":"{\"run_id\":\"r\"}","call_id":"c"}}` + "\n"
	result := `{"type":"response_item","payload":{"type":"function_call_output","call_id":"c","output":"{\"lanes\":[{\"slot\":\"lead\"}]}"}}`
	emptyCall := strings.ReplaceAll(call, `"call_id":"c"`, `"call_id":""`)
	emptyResult := strings.ReplaceAll(result, `"call_id":"c"`, `"call_id":""`)
	for _, tc := range []struct {
		data, run string
		want      bool
	}{
		{call + result, "r", true}, {result, "r", false}, {call + result, "foreign", false},
		{emptyCall + emptyResult, "r", false},
		{emptyCall + result, "r", false},
		{call + emptyResult, "r", false},
		{strings.ReplaceAll(call, "mcp__moai__", "mcp__foreign__") + result, "r", false},
		{call + `{"type":"response_item","payload":{"type":"function_call_output","call_id":"other","output":"{\"lanes\":[]}"}}`, "r", false},
	} {
		_, ok := operationalMCPResult([]byte(tc.data), tc.run)
		if ok != tc.want {
			t.Fatalf("got %v want %v: %s", ok, tc.want, tc.data)
		}
	}
	if _, ok := operationalMCPResultCount([]byte(call+result+"\n"+result), "r", 2); ok {
		t.Fatal("duplicate old output counted as a new status call")
	}
	if _, ok := operationalOutputLanes([]byte(`{"isError":true,"lanes":[]}`)); ok {
		t.Fatal("error response accepted as evidence")
	}
}
