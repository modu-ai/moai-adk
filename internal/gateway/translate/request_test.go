package translate

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func decode(t *testing.T, b []byte) map[string]any {
	t.Helper()
	var d map[string]any
	if err := json.Unmarshal(b, &d); err != nil {
		t.Fatal(err)
	}
	return d
}
func request(t *testing.T, b string) (map[string]any, *ResponseContext) {
	t.Helper()
	out, c, err := Request("gpt-5.6-sol", []byte(b), Limits{})
	if err != nil {
		t.Fatal(err)
	}
	return decode(t, out), c
}

const ordinary = `{"model":"ignored","max_tokens":123,"stream":true,"system":[{"type":"text","text":"first"},{"type":"text","text":"second"}],"messages":[{"role":"user","content":"hello"},{"role":"system","content":"in place"},{"role":"assistant","content":[{"type":"text","text":"calling"},{"type":"tool_use","id":"call-A","name":"a.b","input":{"x":1}}]},{"role":"user","content":[{"type":"tool_result","tool_use_id":"call-A","content":"answer"},{"type":"text","text":"after"}]}],"tools":[{"name":"a.b","input_schema":{"type":"object","properties":{"x":{"type":"number"}}}},{"name":"a_b","input_schema":{"type":"object"}}]}`

func TestRequestPreservesPublicOrderSchemaAndLocalNames(t *testing.T) {
	d, c := request(t, ordinary)
	if d["model"] != "gpt-5.6-sol" || d["instructions"] != "first\n\nsecond" || d["max_output_tokens"] != float64(123) {
		t.Fatalf("header %#v", d)
	}
	input := d["input"].([]any)
	roles := []any{}
	for _, v := range input {
		x := v.(map[string]any)
		if r, ok := x["role"]; ok {
			roles = append(roles, r)
		} else {
			roles = append(roles, x["type"])
		}
	}
	if !reflect.DeepEqual(roles, []any{"user", "system", "assistant", "function_call", "function_call_output", "user"}) {
		t.Fatal(roles)
	}
	tools := d["tools"].([]any)
	first := tools[0].(map[string]any)
	second := tools[1].(map[string]any)
	if first["strict"] != false || first["name"] == second["name"] {
		t.Fatal(tools)
	}
	if _, ok := first["parameters"].(map[string]any)["required"]; ok {
		t.Fatal("optional schema changed")
	}
	call := input[3].(map[string]any)
	if call["name"] != first["name"] || call["call_id"] != "call-A" || call["arguments"] != `{"x":1}` {
		t.Fatal(call)
	}
	if c.names[first["name"].(string)] != "a.b" {
		t.Fatal("reverse map")
	}
}
func TestRequestRejectsLossAndAmbiguousJSON(t *testing.T) {
	bad := []string{
		`{"max_tokens":1,"max_tokens":2,"messages":[]}`, `{"max_tokens":1.0,"messages":[]}`,
		`{"max_tokens":true,"messages":[]}`, `{"max_tokens":0,"messages":[]}`,
		`{"max_tokens":1,"messages":[{"role":"user","content":[{"type":"image","source":{}}]}]}`,
		`{"max_tokens":1,"messages":[{"role":"assistant","content":[{"type":"thinking","thinking":"private","signature":"sig"}]}]}`,
		`{"max_tokens":1,"messages":[{"role":"assistant","content":[{"type":"redacted_thinking","data":"opaque"}]}]}`,
		`{"max_tokens":1,"messages":[{"role":"assistant","content":[{"type":"tool_use","name":"t","id":"a","input":{}}]}]}`,
		`{"max_tokens":1,"messages":[{"role":"user","content":[{"type":"tool_result","tool_use_id":"missing","content":"x"}]}]}`,
		`{"max_tokens":1,"context_management":{},"messages":[{"role":"user","content":"x"}]}`,
		`{"max_tokens":1,"thinking":{"type":"enabled"},"messages":[{"role":"user","content":"x"}]}`,
	}
	for _, b := range bad {
		t.Run(b, func(t *testing.T) {
			out, _, err := Request("gpt-5.6-sol", []byte(b), Limits{})
			if err == nil || out != nil {
				t.Fatal("accepted lossy request", string(out))
			}
		})
	}
	n := int64(11)
	for _, l := range []Limits{{MaxBodyBytes: 3}, {ContextTokens: 10, InputTokens: &n}, {ContextTokens: 10}} {
		if _, _, e := Request("gpt-5.6-sol", []byte(ordinary), l); e == nil {
			t.Fatal("limit ignored")
		}
	}
}
func TestCapturedPublicConversationGolden(t *testing.T) {
	b, e := os.ReadFile("../testdata/validation-capture/claude-code-2.1.268/request-005.json")
	if e != nil {
		t.Fatal(e)
	}
	if _, _, e = Request("gpt-5.6-sol", b, Limits{}); e == nil {
		t.Fatal("raw policy must be explicit")
	}
	d := decode(t, b)
	// Public-only golden projection excludes locally authorized native identity.
	for _, k := range []string{"thinking", "context_management", "output_config", "metadata"} {
		delete(d, k)
	}
	b, _ = json.Marshal(d)
	out, _, e := Request("gpt-5.6-sol", b, Limits{})
	if e != nil {
		t.Fatal(e)
	}
	got := decode(t, out)["input"].([]any)
	src := d["messages"].([]any)
	if len(got) != len(src) {
		t.Fatalf("count %d != %d", len(got), len(src))
	}
	for i, v := range src {
		m := v.(map[string]any)
		x := got[i].(map[string]any)
		if x["role"] != m["role"] {
			t.Fatal("role order", i)
		}
		var want []string
		switch c := m["content"].(type) {
		case string:
			want = []string{c}
		case []any:
			for _, z := range c {
				want = append(want, z.(map[string]any)["text"].(string))
			}
		}
		var texts []string
		for _, z := range x["content"].([]any) {
			texts = append(texts, z.(map[string]any)["text"].(string))
		}
		if !reflect.DeepEqual(want, texts) {
			t.Fatal("text loss", i)
		}
	}
}
func TestRequestMapConcurrentIsolation(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			name := "a.b"
			if i%2 == 0 {
				name = "a/b"
			}
			body := strings.ReplaceAll(ordinary, "a.b", name)
			_, c := request(t, body)
			for _, original := range c.names {
				if original == "a.b" && name != "a.b" {
					t.Error("cross request leak")
				}
			}
		}(i)
	}
	wg.Wait()
}

func TestInvalidUTF8NeverReplacesPublicText(t *testing.T) {
	body := append([]byte(`{"max_tokens":1,"messages":[{"role":"user","content":"`), byte(0xff))
	body = append(body, []byte(`"}]}`)...)
	if b, _, e := Request("gpt-5.6-sol", body, Limits{}); e == nil {
		t.Fatalf("invalid UTF8 became %s", b)
	}
}

func TestUnicodeEscapesAreScalars(t *testing.T) {
	esc := func(hex string) string { return string([]byte{92, 117}) + hex }
	cases := []struct {
		name, text, want string
		reject           bool
	}{
		{"high alone", esc("d800"), "", true}, {"low alone", esc("dc00"), "", true},
		{"high then scalar", esc("d800") + esc("0041"), "", true}, {"two high", esc("d800") + esc("dbff"), "", true},
		{"pair", esc("d83d") + esc("de00"), "😀", false}, {"literal replacement", "�", "�", false},
		{"escaped replacement", esc("fffd"), "�", false}, {"escaped backslash", string([]byte{92}) + esc("d800"), esc("d800"), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body := []byte(`{"max_tokens":1,"messages":[{"role":"user","content":"` + tc.text + `"}]}`)
			b, _, e := Request("gpt-5.6-sol", body, Limits{})
			if tc.reject {
				if e == nil {
					t.Fatalf("escaped surrogate replaced: %s", b)
				}
				return
			}
			if e != nil {
				t.Fatal(e)
			}
			text := decode(t, b)["input"].([]any)[0].(map[string]any)["content"].([]any)[0].(map[string]any)["text"]
			if text != tc.want {
				t.Fatalf("%q != %q", text, tc.want)
			}
		})
	}
}

func TestValidateJSONObjectPublicBoundary(t *testing.T) {
	for _, raw := range []string{`{"text":"😀"}`, `{"n":9007199254740993}`} {
		if e := ValidateJSONObject([]byte(raw)); e != nil {
			t.Fatal(e)
		}
	}
	for _, raw := range []string{`[]`, `{"x":1,"x":2}`, `{"x":"` + string([]byte{92, 117}) + `d800"}`} {
		if e := ValidateJSONObject([]byte(raw)); e == nil {
			t.Fatal("ambiguous JSON accepted")
		}
	}
}
