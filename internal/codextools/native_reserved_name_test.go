package codextools

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNativeMCPReservedNamesAreAliasedWithoutChangingPublicIdentity(t *testing.T) {
	defs := []Definition{def("mcp"), def("mcp__synthetic__tool"), def("ordinary_tool"), def("mcp__" + strings.Repeat("x", 180))}
	unbound := Binding{ConversationID: "reserved-native", AccountScope: "account"}
	r, err := New(unbound, defs)
	if err != nil {
		t.Fatal(err)
	}
	second, err := New(unbound, defs)
	if err != nil {
		t.Fatal(err)
	}
	bound, err := r.BindThread("thread")
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	native := r.NativeTools()
	for i, tool := range native {
		if tool.Name == "mcp" || strings.HasPrefix(tool.Name, "mcp__") || !safeName.MatchString(tool.Name) || seen[tool.Name] {
			t.Fatalf("invalid/reserved/colliding native name: %q", tool.Name)
		}
		seen[tool.Name] = true
		if tool.Name != second.NativeTools()[i].Name {
			t.Fatal("alias is nondeterministic")
		}
		if i == len(defs) {
			if tool.Name != DispatcherName {
				t.Fatal("dispatcher changed")
			}
			continue
		}
		call := Call{TurnID: "turn", CallID: tool.Name, Tool: tool.Name, Arguments: json.RawMessage(`{"q":"ok","nested":{"n":1}}`)}
		got, err := r.Begin(bound, call, defs)
		if err != nil || got.Name != defs[i].Name {
			t.Fatalf("public identity lost: %q err=%v", got.Name, err)
		}
		if err = r.Complete(bound, call.TurnID, call.CallID, defs); err != nil {
			t.Fatal(err)
		}
		if i == 2 && tool.Name != "ordinary_tool" {
			t.Fatal("unreserved name changed")
		}
		if i != 2 {
			collision := def(tool.Name)
			if _, err = New(unbound, []Definition{defs[i], collision}); err == nil {
				t.Fatal("caller forged generated alias")
			}
		}
	}
}
