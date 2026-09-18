package codextools

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

var owner = Binding{ConversationID: "conversation-a", ThreadID: "thread-a", AccountScope: "account-a"}

const schema = `{"type":"object","properties":{"q":{"type":"string","enum":["ok"]},"nested":{"type":"object","properties":{"n":{"type":"integer"}},"required":["n"],"additionalProperties":false}},"required":["q","nested"],"additionalProperties":false}`

func def(name string) Definition { return Definition{Name: name, InputSchema: json.RawMessage(schema)} }
func setup(t *testing.T) *Registry {
	t.Helper()
	r, err := New(owner, []Definition{def("ToolSearch"), {Name: "DeferredToolPlaceholder", Deferred: true, InputSchema: json.RawMessage(`{"type":"object","properties":{}}`)}})
	if err != nil {
		t.Fatal(err)
	}
	return r
}
func validCall(id string) Call {
	return Call{TurnID: "turn-a", CallID: id, Tool: DispatcherName, Arguments: json.RawMessage(`{"name":"mcp__late","arguments":{"q":"ok","nested":{"n":1}}}`)}
}
func snapshot() []Definition { return []Definition{def("ToolSearch"), def("mcp__late")} }
func discover(t *testing.T, r *Registry) {
	t.Helper()
	if err := r.Discover(owner, []Reference{{Type: "tool_reference", ToolName: "mcp__late"}}, snapshot()); err != nil {
		t.Fatal(err)
	}
}
func TestHybridPositiveAndNoThreadRecreation(t *testing.T) {
	r := setup(t)
	before, _ := json.Marshal(r.NativeTools())
	discover(t, r)
	after, _ := json.Marshal(r.NativeTools())
	if string(before) != string(after) {
		t.Fatal("fixed tools changed")
	}
	for i := 0; i < 2; i++ {
		call := validCall(fmt.Sprint(i))
		got, err := r.Begin(owner, call, snapshot())
		if err != nil || got.Name != "mcp__late" {
			t.Fatal(got, err)
		}
		if err := r.Complete(owner, call.TurnID, call.CallID, snapshot()); err != nil {
			t.Fatal(err)
		}
	}
	native := Call{TurnID: "turn-a", CallID: "native", Tool: "ToolSearch", Arguments: json.RawMessage(`{"q":"ok","nested":{"n":1}}`)}
	if got, err := r.Begin(owner, native, snapshot()); err != nil || got.Name != "ToolSearch" {
		t.Fatal(got, err)
	}
}
func TestSixNegativeGroups(t *testing.T) {
	t.Run("1_collision", func(t *testing.T) {
		for _, defs := range [][]Definition{{def("x"), def("x")}, {def(DispatcherName)}, {def("moai_native_reserved")}} {
			if _, err := New(owner, defs); err == nil {
				t.Fatal("collision accepted")
			}
		}
	})
	t.Run("2_missing_definition", func(t *testing.T) {
		r := setup(t)
		if err := r.Discover(owner, []Reference{{Type: "tool_reference", ToolName: "mcp__late"}}, []Definition{def("ToolSearch")}); err == nil {
			t.Fatal("reference authorized without schema")
		}
		if _, err := r.Begin(owner, validCall("x"), snapshot()); err == nil {
			t.Fatal("unknown executed")
		}
	})
	t.Run("3_schema_change", func(t *testing.T) {
		r := setup(t)
		discover(t, r)
		c := validCall("x")
		if _, err := r.Begin(owner, c, snapshot()); err != nil {
			t.Fatal(err)
		}
		changed := snapshot()
		changed[1].InputSchema = json.RawMessage(`{"type":"object"}`)
		if err := r.Complete(owner, c.TurnID, c.CallID, changed); err == nil {
			t.Fatal("new schema accepted for pending")
		}
		if err := r.Discover(owner, []Reference{{Type: "tool_reference", ToolName: "mcp__late"}}, changed); err == nil {
			t.Fatal("redefinition accepted")
		}
	})
	t.Run("4_schema_validation", func(t *testing.T) {
		r := setup(t)
		discover(t, r)
		for i, args := range []string{`{}`, `{"q":1,"nested":{"n":1}}`, `{"q":"no","nested":{"n":1}}`, `{"q":"ok","nested":{"n":"bad"}}`, `{"q":"ok","nested":{"n":1},"extra":true}`} {
			c := validCall(fmt.Sprint(i))
			c.Arguments = json.RawMessage(`{"name":"mcp__late","arguments":` + args + `}`)
			if _, err := r.Begin(owner, c, snapshot()); err == nil {
				t.Fatal("invalid executed", args)
			}
		}
		for _, s := range []string{`{"$schema":"https://unknown.invalid/schema","type":"object"}`, `{"$ref":"file:///etc/passwd"}`, `{"$ref":"https://example.invalid/schema"}`, `{"type":"object","type":"string"}`} {
			d := def("invalid")
			d.InputSchema = json.RawMessage(s)
			if _, err := New(owner, []Definition{d}); err == nil {
				t.Fatal("invalid schema accepted", s)
			}
		}
	})
	t.Run("5_binding_and_replay", func(t *testing.T) {
		r := setup(t)
		discover(t, r)
		foreign := owner
		foreign.ConversationID = "other"
		if _, err := r.Begin(foreign, validCall("x"), snapshot()); err == nil {
			t.Fatal("foreign executed")
		}
		c := validCall("x")
		if _, err := r.Begin(owner, c, snapshot()); err != nil {
			t.Fatal(err)
		}
		if _, err := r.Begin(owner, c, snapshot()); err == nil {
			t.Fatal("duplicate executed")
		}
		if err := r.Complete(owner, "foreign-turn", c.CallID, snapshot()); err == nil {
			t.Fatal("wrong turn consumed")
		}
		if err := r.Complete(owner, c.TurnID, c.CallID, snapshot()); err != nil {
			t.Fatal(err)
		}
		if err := r.Complete(owner, c.TurnID, c.CallID, snapshot()); err == nil {
			t.Fatal("duplicate consumed")
		}
	})
	t.Run("6_native_bypass_recursion", func(t *testing.T) {
		r := setup(t)
		discover(t, r)
		for i, name := range []string{"ToolSearch", DispatcherName, "DeferredToolPlaceholder"} {
			c := validCall(fmt.Sprint(i))
			c.Arguments = json.RawMessage(strings.Replace(string(c.Arguments), "mcp__late", name, 1))
			if _, err := r.Begin(owner, c, snapshot()); err == nil {
				t.Fatal("bypass executed", name)
			}
		}
	})
}

func TestSnapshotRevocationAndDuplicateDiscovery(t *testing.T) {
	r := setup(t)
	discover(t, r)
	if err := r.Discover(owner, []Reference{{Type: "tool_reference", ToolName: "mcp__late"}}, snapshot()); err != nil {
		t.Fatal("unchanged repeated discovery rejected", err)
	}
	if _, err := r.Begin(owner, validCall("revoked"), []Definition{def("ToolSearch")}); err == nil {
		t.Fatal("revoked tool returned executable invocation")
	}
	if _, err := r.Begin(owner, validCall("pending"), snapshot()); err != nil {
		t.Fatal(err)
	}
	if err := r.Complete(owner, "turn-a", "pending", []Definition{def("ToolSearch")}); err == nil {
		t.Fatal("revoked snapshot consumed")
	}
}
func TestSchemaCanonicalizationAndLocalReferences(t *testing.T) {
	d := def("local.ref")
	d.InputSchema = json.RawMessage(`{"type":"object","$defs":{"value":{"type":"integer"}},"properties":{"n":{"$ref":"#/$defs/value"}},"required":["n"]}`)
	r, err := New(owner, []Definition{d})
	if err != nil {
		t.Fatal(err)
	}
	tool := r.NativeTools()[0]
	if tool.Name == d.Name {
		t.Fatal("unsafe name not mapped")
	}
	got, err := r.Begin(owner, Call{TurnID: "turn-a", CallID: "one", Tool: tool.Name, Arguments: json.RawMessage(`{"n":7}`)}, []Definition{d})
	if err != nil || got.Name != d.Name {
		t.Fatal(got, err)
	}
	reordered := d
	reordered.InputSchema = json.RawMessage(`{"required":["n"],"properties":{"n":{"$ref":"#/$defs/value"}},"$defs":{"value":{"type":"integer"}},"type":"object"}`)
	if err := r.Complete(owner, "turn-a", "one", []Definition{reordered}); err != nil {
		t.Fatal("object order changed digest", err)
	}
}
func TestStrictArgumentsAndPlaceholder(t *testing.T) {
	r := setup(t)
	discover(t, r)
	for _, raw := range []string{`{"name":"mcp__late","name":"other","arguments":{}}`, `{"name":"mcp__late","arguments":{"q":"ok","q":"bad"}}`, strings.Repeat("[", 66) + strings.Repeat("]", 66)} {
		c := validCall("invalid")
		c.Arguments = json.RawMessage(raw)
		if _, err := r.Begin(owner, c, snapshot()); err == nil {
			t.Fatal("ambiguous arguments accepted")
		}
	}
	if _, err := New(owner, []Definition{{Name: PlaceholderName, Deferred: true, InputSchema: json.RawMessage(`{"type":"object","properties":{"real":{"type":"string"}}}`)}}); err == nil {
		t.Fatal("spoofed placeholder silently discarded")
	}
}

func TestNativeToolWireDiscriminator(t *testing.T) {
	r := setup(t)
	for _, tool := range r.NativeTools() {
		raw, _ := json.Marshal(tool)
		var wire map[string]any
		if err := json.Unmarshal(raw, &wire); err != nil {
			t.Fatal(err)
		}
		if wire["type"] != "function" {
			t.Fatal("installed DynamicToolSpec requires function discriminator")
		}
	}
}

func TestRevocationBeforeInvocation(t *testing.T) {
	r := setup(t)
	discover(t, r)
	if _, err := r.Begin(owner, validCall("revoked"), []Definition{def("ToolSearch")}); err == nil {
		t.Fatal("revoked tool returned executable invocation")
	}
}

func TestPrepareBeforeServerAllocatesThread(t *testing.T) {
	unbound := owner
	unbound.ThreadID = ""
	r, err := New(unbound, []Definition{def("ToolSearch")})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.NativeTools()) != 2 {
		t.Fatal("cannot prepare start request")
	}
	call := Call{TurnID: "turn", CallID: "one", Tool: "ToolSearch", Arguments: json.RawMessage(`{"q":"ok","nested":{"n":1}}`)}
	if _, err := r.Begin(unbound, call, []Definition{def("ToolSearch")}); err == nil {
		t.Fatal("unbound execution")
	}
	bound, err := r.BindThread("server-allocated-thread")
	if err != nil || bound.ThreadID != "server-allocated-thread" {
		t.Fatal(bound, err)
	}
	if _, err := r.BindThread("other"); err == nil {
		t.Fatal("thread rebound")
	}
	if _, err := r.Begin(bound, call, []Definition{def("ToolSearch")}); err != nil {
		t.Fatal(err)
	}
}
