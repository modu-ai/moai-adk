package receipt

import (
	"encoding/json"
	"strings"
	"testing"
)

const testUUID = "7f3010a7-5285-4ef8-a8a4-7237a2d0b6b1"

func candidate(prefix, previous, opaque string) Candidate {
	return Candidate{Prefix: Hash([]byte(prefix)), Previous: Hash([]byte(previous)), Provider: "openai", Opaque: Hash([]byte(opaque)), Items: 1, Required: true, Complete: true}
}
func TestCandidateAcceptanceAndBranch(t *testing.T) {
	m, err := New(testUUID)
	if err != nil {
		t.Fatal(err)
	}
	a := candidate("p", "parent", "A")
	b := a
	b.Opaque = Hash([]byte("B"))
	for _, v := range []Candidate{a, b, a} {
		if err = m.Publish(v); err != nil {
			t.Fatal(err)
		}
	}
	if len(m.Candidates()) != 2 {
		t.Fatal("duplicate increased count")
	}
	for _, v := range []Candidate{a, b} {
		if err = m.Check(testUUID, []Observation{{Prefix: v.Prefix, Previous: v.Previous, Provider: v.Provider, Opaque: v.Opaque, Items: v.Items}}); err != nil {
			t.Fatal(err)
		}
	}
	for _, o := range []Observation{{Prefix: a.Prefix, Previous: a.Previous, Provider: a.Provider}, {Prefix: a.Prefix, Previous: a.Previous, Provider: a.Provider, Opaque: Hash([]byte("mixed")), Items: 1}, {Prefix: a.Prefix, Previous: a.Previous, Provider: a.Provider, Opaque: a.Opaque, Items: 2}, {Prefix: Hash([]byte("edited")), Previous: a.Previous, Provider: a.Provider, Opaque: a.Opaque, Items: 1}} {
		if m.Check(testUUID, []Observation{o}) == nil {
			t.Fatal("invalid history accepted")
		}
	}
	if m.Check(testUUID, nil) != nil {
		t.Fatal("unused branch forced into history")
	}
	if m.Check("00000000-0000-0000-0000-000000000000", nil) == nil {
		t.Fatal("foreign session")
	}
	none := a
	none.Required = false
	none.Opaque = Digest{}
	none.Items = 0
	n, _ := New(testUUID)
	if n.Publish(none) != nil {
		t.Fatal("publish none")
	}
	empty := Observation{Prefix: a.Prefix, Previous: a.Previous, Provider: a.Provider}
	if n.Check(testUUID, []Observation{empty}) != nil {
		t.Fatal("sole empty rejected")
	}
	if n.Publish(a) != nil || n.Check(testUUID, []Observation{empty}) == nil {
		t.Fatal("ambiguous empty accepted")
	}
	if n.Check(testUUID, []Observation{{Prefix: a.Prefix, Previous: a.Previous, Provider: a.Provider, Opaque: a.Opaque, Items: 1}}) != nil {
		t.Fatal("exact nonempty rejected")
	}
}
func TestManifestIntegrityLimitsAndFork(t *testing.T) {
	m, _ := New(testUUID)
	a := candidate("p", "parent", "A")
	_ = m.Publish(a)
	raw, err := m.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	n, err := Parse(raw, testUUID)
	if err != nil || len(n.Candidates()) != 1 {
		t.Fatal(err)
	}
	if _, err = Parse(raw, "00000000-0000-0000-0000-000000000000"); err == nil {
		t.Fatal("foreign manifest")
	}
	var object map[string]json.RawMessage
	_ = json.Unmarshal(raw, &object)
	object["count"] = json.RawMessage("0")
	bad, _ := json.Marshal(object)
	if _, err = Parse(bad, testUUID); err == nil {
		t.Fatal("lost receipt accepted")
	}
	if _, err = Parse(append(raw, raw...), testUUID); err == nil {
		t.Fatal("trailing JSON")
	}
	if _, err = Parse(make([]byte, MaxBytes+1), testUUID); err == nil {
		t.Fatal("oversize")
	}
	fork, err := m.Fork("00000000-0000-4000-8000-000000000001")
	if err != nil {
		t.Fatal(err)
	}
	b := a
	b.Opaque = Hash([]byte("B"))
	_ = m.Publish(b)
	if len(fork.Candidates()) != 1 {
		t.Fatal("fork changed with parent")
	}
	for i := len(m.Candidates()); i < MaxCandidates; i++ {
		v := a
		v.Opaque = Hash([]byte{byte(i), byte(i >> 8)})
		if err = m.Publish(v); err != nil {
			t.Fatal(i, err)
		}
	}
	extra := a
	extra.Opaque = Hash([]byte("overflow"))
	if m.Publish(extra) == nil {
		t.Fatal("cap not enforced")
	}
	invalid := a
	invalid.Complete = false
	if fork.Publish(invalid) == nil {
		t.Fatal("incomplete published")
	}
}
func TestCanonicalPublicPrefixes(t *testing.T) {
	decode := func(s string) (string, error) {
		if s == "bound" {
			return "call_1", nil
		}
		return s, nil
	}
	a := []byte(`[{"role":"user","content":"hi"},{"role":"assistant","content":[{"type":"thinking","thinking":"private"},{"type":"text","text":"OK","cache_control":{"type":"ephemeral"}},{"type":"tool_use","id":"bound","name":"echo","input":{"b":1.0,"a":2}}]},{"role":"user","content":[{"type":"tool_result","tool_use_id":"bound","content":"result"}]}]`)
	b := []byte(`[{"content":[{"type":"text","text":"hi"}],"role":"user"},{"role":"assistant","content":[{"text":"OK","type":"text"},{"input":{"a":2,"b":1},"name":"echo","id":"call_1","type":"tool_use"}]},{"role":"user","content":[{"content":"result","tool_use_id":"call_1","type":"tool_result"}]}]`)
	pa, err := CanonicalPrefixes(a, decode)
	if err != nil {
		t.Fatal(err)
	}
	pb, err := CanonicalPrefixes(b, decode)
	if err != nil || len(pa) != 1 || pa[0] != pb[0] {
		t.Fatalf("normalization %v %v %v", pa, pb, err)
	}
	for _, bad := range []string{`[{"role":"user","role":"assistant","content":"hi"}]`, `[{"role":"assistant","content":[{"type":"unknown"}]}]`, `[{"role":"assistant","content":"OK"}] trailing`} {
		if _, err = CanonicalPrefixes([]byte(bad), decode); err == nil {
			t.Fatal("invalid accepted", bad)
		}
	}
	p1, _ := CanonicalPrefixes([]byte(`[{"role":"system","content":"notice"},{"role":"assistant","content":"OK"}]`), decode)
	p2, _ := CanonicalPrefixes([]byte(`[{"role":"system","content":"changed"},{"role":"assistant","content":"OK"}]`), decode)
	if p1[0] == p2[0] {
		t.Fatal("system message erased")
	}
}

func TestProjectionRejectsAmbiguousUnicodeAndPreservesMeaning(t *testing.T) {
	decode := func(s string) (string, error) { return s, nil }
	for _, raw := range []string{`[{"role":"assistant","content":"\ud800"}]`, `[{"role":"assistant","content":"\udc00"}]`, `[{"role":"assistant","content":"\ud800\u0041"}]`} {
		if _, e := CanonicalPrefixes([]byte(raw), decode); e == nil {
			t.Errorf("unpaired surrogate accepted: %s", raw)
		}
	}
	a, e := CanonicalPrefixes([]byte(`[{"role":"assistant","content":"\ud83d\ude00"}]`), decode)
	if e != nil {
		t.Fatal(e)
	}
	b, e := CanonicalPrefixes([]byte(`[{"role":"assistant","content":"😀"}]`), decode)
	if e != nil || a[0] != b[0] {
		t.Fatal("surrogate pair meaning lost")
	}
}
func TestMalformedProjectionAndManifest(t *testing.T) {
	decode := func(s string) (string, error) { return s, nil }
	bad := []string{`null`, `{}`, `[null]`, `[{"role":1,"content":"a"}]`, `[{"role":"unknown","content":"a"}]`, `[{"role":"assistant","content":null}]`, `[{"role":"assistant","content":[null]}]`, `[{"role":"assistant","content":[{}]}]`, `[{"role":"assistant","content":[{"type":"text","text":1}]}]`, `[{"role":"assistant","content":[{"type":"tool_use","id":""}]}]`, `[{"role":"assistant","content":[{"type":"tool_result","tool_use_id":"x","content":42}]}]`, `[{"role":"assistant","content":[{"type":"tool_use","id":"x","input":1e999999}]}]`, `[{"role":"assistant","content":[{"type":"tool_use","id":"x","input":1e-999999}]}]`, `[{"role":"assistant","content":[{"type":"tool_use","id":"x","input":{"x":1,"x":2}}]}]`, `[`, `{"a":`, `[[`, `"unterminated`}
	for _, raw := range bad {
		if _, e := CanonicalPrefixes([]byte(raw), decode); e == nil {
			t.Error("accepted", raw)
		}
	}
	if _, e := CanonicalPrefixes([]byte{0xff}, decode); e == nil {
		t.Fatal("invalid UTF8")
	}
	if _, e := CanonicalPrefixes(make([]byte, MaxBytes+1), decode); e == nil {
		t.Fatal("oversize projection")
	}
	if _, e := CanonicalPrefixes([]byte(`[{"role":"assistant","content":[{"type":"tool_use","id":"x"}]}]`), nil); e == nil {
		t.Fatal("nil ID verifier")
	}
	if _, e := New("not-a-session"); e == nil {
		t.Fatal("invalid UUID")
	}
	m, _ := New(testUUID)
	if _, e := m.Fork("invalid"); e == nil {
		t.Fatal("invalid fork UUID")
	}
	raw, _ := m.Marshal()
	if _, e := Parse(raw, "bad"); e == nil {
		t.Fatal("invalid resume UUID")
	}
	for _, raw := range []string{`{}`, `{"version":1,"version":1}`, `{"unknown":true}`, `{"session":"AB"}`, `{"session":"ZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZ"}`} {
		if _, e := Parse([]byte(raw), testUUID); e == nil {
			t.Fatal("malformed manifest")
		}
	}
	var zero Manifest
	if zero.Publish(candidate("p", "q", "r")) == nil {
		t.Fatal("uninitialized manifest accepted")
	}
}

func TestProjectionPublicMutationDetection(t *testing.T) {
	decode := func(s string) (string, error) { return s, nil }
	base := `[{"role":"user","content":"question"},{"role":"assistant","content":[{"type":"tool_use","id":"call_1","name":"echo","input":{"n":1,"b":true,"a":[1,2],"z":null}}]},{"role":"user","content":[{"type":"tool_result","tool_use_id":"call_1","content":"result","is_error":false}]},{"role":"assistant","content":"answer"}]`
	initial, e := CanonicalPrefixes([]byte(base), decode)
	if e != nil {
		t.Fatal(e)
	}
	for _, change := range [][2]string{{"question", "different"}, {"answer", "different"}, {"call_1", "call_2"}, {`"content":"result"`, `"content":"different"`}, {`"n":1`, `"n":2`}, {`[1,2]`, `[2,1]`}, {`"is_error":false`, `"is_error":true`}, {`"b":true`, `"b":false`}} {
		mutated := strings.ReplaceAll(base, change[0], change[1])
		got, e := CanonicalPrefixes([]byte(mutated), decode)
		if e != nil {
			t.Fatal(e)
		}
		if got[len(got)-1] == initial[len(initial)-1] {
			t.Fatal("public mutation collapsed", change)
		}
	}
}

func TestManifestByteBoundaryAndHashOnly(t *testing.T) {
	m, _ := New(testUUID)
	v := candidate("CANARY_PUBLIC_847299", "CANARY_PRIOR_847299", "CANARY_OPAQUE_847299")
	if e := m.Publish(v); e != nil {
		t.Fatal(e)
	}
	raw, e := m.Marshal()
	if e != nil {
		t.Fatal(e)
	}
	for _, plain := range []string{testUUID, "CANARY_PUBLIC_847299", "CANARY_PRIOR_847299", "CANARY_OPAQUE_847299"} {
		if strings.Contains(string(raw), plain) {
			t.Fatal("plaintext persisted", plain)
		}
	}
	padded := append(append([]byte(nil), raw...), []byte(strings.Repeat(" ", MaxBytes-len(raw)))...)
	if _, e = Parse(padded, testUUID); e != nil {
		t.Fatal("exact byte boundary", e)
	}
	if _, e = Parse(append(padded, ' '), testUUID); e != ErrLimit {
		t.Fatal("over byte boundary", e)
	}
}

func TestManifestRequiresEveryStructuralField(t *testing.T) {
	m, _ := New(testUUID)
	raw, _ := m.Marshal()
	var d map[string]json.RawMessage
	_ = json.Unmarshal(raw, &d)
	delete(d, "count")
	raw, _ = json.Marshal(d)
	if _, e := Parse(raw, testUUID); e == nil {
		t.Fatal("missing count accepted")
	}
	none := candidate("p", "parent", "opaque")
	none.Required = false
	none.Items = 0
	none.Opaque = Digest{}
	_ = m.Publish(none)
	raw, _ = m.Marshal()
	_ = json.Unmarshal(raw, &d)
	var list []map[string]json.RawMessage
	_ = json.Unmarshal(d["candidates"], &list)
	delete(list[0], "opaque_required")
	d["candidates"], _ = json.Marshal(list)
	raw, _ = json.Marshal(d)
	if _, e := Parse(raw, testUUID); e == nil {
		t.Fatal("missing required flag accepted")
	}
}
