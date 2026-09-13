package translate

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
)

const titleSchema = `{"type":"object","properties":{"title":{"type":"string"}},"required":["title"],"additionalProperties":false}`

func policyInput(extra string) []byte {
	return []byte(`{"max_tokens":32000,"messages":[{"role":"user","content":"synthetic"}]` + extra + `}`)
}
func TestNativePolicyGPTMapping(t *testing.T) {
	for _, model := range []string{"gpt-6-astra", "gpt-5.6-sol", "gpt-5.6-terra", "gpt-5.6-luna"} {
		for _, thinking := range []string{"", `,"thinking":{"type":"adaptive"}`, `,"thinking":{"type":"adaptive","display":"omitted"}`} {
			raw, c, e := Request(model, policyInput(thinking+`,"output_config":{"effort":"high","format":{"type":"json_schema","schema":`+titleSchema+`}}`), Limits{PolicyProfile: PolicyGPTNative})
			if e != nil {
				t.Fatal(e)
			}
			out := decode(t, raw)
			if out["reasoning"].(map[string]any)["effort"] != "high" || !c.title {
				t.Fatal(out)
			}
			f := out["text"].(map[string]any)["format"].(map[string]any)
			if f["name"] != "moai_native_output" || f["strict"] != true || f["schema"].(map[string]any)["additionalProperties"] != false {
				t.Fatal(f)
			}
		}
	}
}

func TestNativePolicyObservedClaudeRequest(t *testing.T) {
	raw := []byte(`{"model":"gpt-5.6-sol","max_tokens":32000,"thinking":{"type":"adaptive","display":"omitted"},"output_config":{"effort":"high"},"messages":[{"role":"user","content":"synthetic"}],"tools":[{"name":"deferred","description":"deferred tool","input_schema":{"type":"object"},"defer_loading":true},{"name":"ready","description":"ready tool","input_schema":{"type":"object"}}]}`)
	policy, err := ValidateNativePolicy(raw, PolicyGPTNative)
	if err != nil {
		t.Fatal(err)
	}
	if policy.Display != "omitted" {
		t.Fatalf("display = %q, want omitted", policy.Display)
	}
	wire, _, err := Request("gpt-5.6-sol", raw, Limits{PolicyProfile: PolicyGPTNative})
	if err != nil {
		t.Fatal(err)
	}
	tools, ok := decode(t, wire)["tools"].([]any)
	if !ok || len(tools) != 1 || tools[0].(map[string]any)["name"] != "ready" {
		t.Fatalf("deferred tool was not omitted: %#v", tools)
	}
}

func TestNativePolicyInvalidAndReceiptGate(t *testing.T) {
	for _, extra := range []string{`,"thinking":null`, `,"thinking":{"type":"disabled"}`, `,"thinking":{"type":"enabled","budget_tokens":10}`, `,"thinking":{"type":"adaptive","unknown":true}`, `,"thinking":{"type":"adaptive","display":"summarized"}`, `,"thinking":{"type":"adaptive","display":null}`, `,"thinking":{"type":"adaptive"}`, `,"output_config":null`, `,"output_config":{"effort":"low"}`, `,"output_config":{"effort":"high","effort":"high"}`, `,"output_config":{"format":{"type":"json_schema","schema":{}}}`, `,"context_management":{"edits":[{"type":"clear_thinking_20251015","keep":"all"}]}`, `,"metadata":{"user_id":"{\"session_id\":\"synthetic\"}"}`} {
		if raw, _, e := Request("gpt-5.6-sol", policyInput(extra), Limits{PolicyProfile: PolicyGPTNative}); e == nil || raw != nil {
			t.Fatalf("accepted %s", extra)
		}
	}
	if _, _, e := Request("foreign", policyInput(`,"output_config":{"effort":"high"}`), Limits{PolicyProfile: PolicyGPTNative}); e == nil {
		t.Fatal("unknown model")
	}
}
func TestNativeTitleResult(t *testing.T) {
	_, c, e := Request("gpt-5.6-sol", policyInput(`,"output_config":{"format":{"type":"json_schema","schema":`+titleSchema+`}}`), Limits{PolicyProfile: PolicyGPTNative})
	if e != nil {
		t.Fatal(e)
	}
	for _, s := range []string{`{"title":"제목"}`, `{"title":""}`} {
		if _, e := c.Response(j(finalResponse("completed", textItem("m", s)))); e != nil {
			t.Fatal(e)
		}
	}
	for _, s := range []string{`{}`, `{"title":1}`, `{"title":"x","extra":1}`, `{"title":"a","title":"b"}`, `{"title":"\ud800"}`, `not JSON`} {
		if _, e := c.Response(j(finalResponse("completed", textItem("m", s)))); e == nil {
			t.Fatal("accepted", s)
		}
	}
	// A title mismatch must not emit a synthetic success terminal even after deltas.
	var out bytes.Buffer
	e = c.Stream(context.Background(), io.NopCloser(strings.NewReader(strings.Join(textEvents("completed"), ""))), &out)
	if e == nil || strings.Contains(out.String(), "event: message_stop") {
		t.Fatal(e, out.String())
	}
}

func TestNativePolicySyntaxMatrix(t *testing.T) {
	good := []string{``, `,"thinking":{"type":"disabled"},"output_config":{"effort":"high"}`, `,"output_config":{"effort":"high"}`, `,"thinking":{"type":"adaptive"},"output_config":{"effort":"high"}`, `,"thinking":{"type":"adaptive","display":"omitted"},"output_config":{"effort":"high"}`, `,"thinking":{"type":"adaptive","display":"summarized"},"output_config":{"effort":"high"}`, `,"context_management":{"edits":[{"type":"clear_thinking_20251015","keep":"all"}]}`, `,"metadata":{"user_id":"{\"account_uuid\":\"\",\"device_id\":\"device\",\"session_id\":\"session\"}"}`}
	for _, x := range good {
		if _, e := ValidateNativePolicy(policyInput(x), PolicyAnthropicNative); e != nil {
			t.Fatal(x, e)
		}
	}
	bad := []string{`,"output_config":{}`, `,"output_config":{"extra":true}`, `,"output_config":{"format":null}`, `,"output_config":{"format":{"type":"other","schema":` + titleSchema + `}}`, `,"thinking":null`, `,"thinking":{"type":"enabled"}`, `,"thinking":{"type":"adaptive"}`, `,"thinking":{"type":"adaptive","budget_tokens":1}`, `,"thinking":{"type":"adaptive","display":"unknown"}`, `,"thinking":{"type":"adaptive","display":null}`, `,"thinking":{"type":"disabled","display":"omitted"}`, `,"context_management":null`, `,"context_management":{"edits":[{"type":"clear_thinking_20251015","keep":1}]}`, `,"context_management":{"edits":[]}`, `,"metadata":null`, `,"metadata":{"x":"y"}`, `,"metadata":{"user_id":1}`, `,"metadata":{"user_id":"bad"}`, `,"metadata":{"user_id":"{\"account_uuid\":1,\"device_id\":\"d\",\"session_id\":\"s\"}"}`}
	for _, x := range bad {
		if _, e := ValidateNativePolicy(policyInput(x), PolicyAnthropicNative); e == nil {
			t.Fatal("accepted", x)
		}
	}
	for _, raw := range [][]byte{[]byte(`{"thinking":null,"thinking":null}`), []byte(`{"output_config":{"effort":"\ud800"}}`), {123, 34, 120, 34, 58, 34, 255, 34, 125}} {
		if _, e := ValidateNativePolicy(raw, PolicyAnthropicNative); e == nil {
			t.Fatal("bad JSON accepted")
		}
	}
	if _, e := ValidateNativePolicy([]byte(`{}`), "unknown"); e == nil {
		t.Fatal("unknown profile")
	}
	if _, _, e := Request("gpt-5.6-sol", policyInput(""), Limits{PolicyProfile: PolicyAnthropicNative}); e == nil {
		t.Fatal("native profile in GPT")
	}
}
func TestNativeTitleStreamPositiveAndBounds(t *testing.T) {
	_, c, e := Request("gpt-5.6-sol", policyInput(`,"output_config":{"format":{"type":"json_schema","schema":`+titleSchema+`}}`), Limits{PolicyProfile: PolicyGPTNative})
	if e != nil {
		t.Fatal(e)
	}
	valid := strings.ReplaceAll(strings.Join(textEvents("completed"), ""), "hello", `{\"title\":\"synthetic\"}`)
	var out bytes.Buffer
	if e = c.Stream(context.Background(), io.NopCloser(strings.NewReader(valid)), &out); e != nil || !strings.Contains(out.String(), "event: message_stop") {
		t.Fatal(e)
	}
	c.limits.MaxOutputBytes = 64
	out.Reset()
	if e = c.Stream(context.Background(), io.NopCloser(strings.NewReader(valid)), &out); e == nil {
		t.Fatal("byte limit bypass")
	}
}

func TestNativeIdentityNeverFallsThroughOrdinaryMetadata(t *testing.T) {
	raw := policyInput(`,"metadata":{"user_id":"{\"account_uuid\":\"\",\"device_id\":\"synthetic-device\",\"session_id\":\"synthetic-session\"}"}`)
	for _, profile := range []PolicyProfile{"", PolicyGPTNative} {
		if wire, _, e := Request("gpt-5.6-sol", raw, Limits{PolicyProfile: profile}); e == nil || wire != nil {
			t.Fatal("native identity escaped without receipt")
		}
	}
}

func TestNativeReceiptAuthorizationCallbackControlsProjection(t *testing.T) {
	raw := policyInput(`,"thinking":{"type":"adaptive"},"output_config":{"effort":"high"},"context_management":{"edits":[{"type":"clear_thinking_20251015","keep":"all"}]},"metadata":{"user_id":"{\"account_uuid\":\"\",\"device_id\":\"device\",\"session_id\":\"session\"}"}`)
	var called bool
	wire, _, err := RequestContext(context.Background(), "gpt-5.6-sol", raw, Limits{PolicyProfile: PolicyGPTNative, NativeReceiptAuthorize: func(_ context.Context, body []byte, policy NativePolicy) error {
		called = string(body) == string(raw) && policy.KeepAll && policy.High && policy.UserID != ""
		return nil
	}})
	if err != nil || !called {
		t.Fatalf("authorized native request rejected: %v called=%v", err, called)
	}
	out := decode(t, wire)
	if _, ok := out["context_management"]; ok {
		t.Fatal("native context policy escaped Responses projection")
	}
	if _, ok := out["metadata"]; ok {
		t.Fatal("native identity metadata escaped Responses projection")
	}
	if out["reasoning"].(map[string]any)["effort"] != "high" {
		t.Fatal("validated effort was not projected")
	}
	if _, _, err = RequestContext(context.Background(), "gpt-5.6-sol", raw, Limits{PolicyProfile: PolicyGPTNative, NativeReceiptAuthorize: func(context.Context, []byte, NativePolicy) error { return errors.New("denied") }}); err == nil {
		t.Fatal("denied native receipt was accepted")
	}
}

func TestNativePolicyObservedMediumEffort(t *testing.T) {
	for _, model := range []string{"gpt-6-astra", "gpt-5.6-sol", "gpt-5.6-terra", "gpt-5.6-luna"} {
		for _, thinking := range []string{"", `,"thinking":{"type":"adaptive","display":"omitted"}`} {
			raw := policyInput(`,"output_config":{"effort":"medium"}` + thinking)
			out, _, err := Request(model, raw, Limits{PolicyProfile: PolicyGPTNative})
			if err != nil {
				t.Fatal(err)
			}
			if decode(t, out)["reasoning"].(map[string]any)["effort"] != "medium" {
				t.Fatal("medium not preserved")
			}
		}
	}
}
