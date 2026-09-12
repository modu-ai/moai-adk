package receipt

import (
	"bytes"
	"testing"
)

func TestToolReferenceCanonicalPrefixPreservesName(t *testing.T) {
	raw := []byte(`[{"role":"assistant","content":[{"type":"tool_use","id":"c","name":"ToolSearch","input":{}}]},{"role":"user","content":[{"type":"tool_result","tool_use_id":"c","content":[{"type":"tool_reference","tool_name":"WebSearch"}]}]},{"role":"assistant","content":"done"}]`)
	decode := func(s string) (string, error) { return s, nil }
	a, err := CanonicalPrefixes(raw, decode)
	if err != nil {
		t.Fatal(err)
	}
	b, err := CanonicalPrefixes(bytes.ReplaceAll(raw, []byte("WebSearch"), []byte("WebFetch")), decode)
	if err != nil {
		t.Fatal(err)
	}
	if a[len(a)-1].Prefix == b[len(b)-1].Prefix {
		t.Fatal("reference name excluded from receipt")
	}
}
