package opaque

import (
	"bytes"
	"encoding/base64"
	"errors"
	"strings"
	"testing"
)

var synthetic = []byte(` { "type":"reasoning", "id":"rs_synthetic", "encrypted_content":"opaque-synthetic+/=", "summary":[], "content":[] } `)

func TestRoundTripAndToolBinding(t *testing.T) {
	e, err := Encode([]Item{{OutputIndex: 1, Raw: synthetic}})
	if err != nil {
		t.Fatal(err)
	}
	d, err := Decode(e.Data())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(d.Items()[0].Raw, synthetic) || d.Items()[0].OutputIndex != 1 || e.Digest() != d.Digest() {
		t.Fatal("changed item")
	}
	marker, err := BindToolID("call_original", e)
	if err != nil {
		t.Fatal(err)
	}
	original, err := RestoreToolID(marker, d)
	if err != nil || original != "call_original" {
		t.Fatal(original, err)
	}
	changed, _ := Encode([]Item{{OutputIndex: 1, Raw: bytes.ReplaceAll(synthetic, []byte("synthetic+/="), []byte("changed"))}})
	if _, err := RestoreToolID(marker, changed); err == nil {
		t.Fatal("hash mismatch accepted")
	}
	items := e.Items()
	items[0].Raw[0] = 'x'
	if bytes.Equal(items[0].Raw, e.Items()[0].Raw) {
		t.Fatal("mutable item alias")
	}
	if _, err := Decode("native-encrypted-data"); !errors.Is(err, ErrNotEnvelope) {
		t.Fatal(err)
	}
}
func TestRejectInvalidItemsAndCarriers(t *testing.T) {
	for _, raw := range [][]byte{[]byte(`{"type":"reasoning","type":"reasoning"}`), []byte(`{"type":"message","id":"rs","encrypted_content":"x","summary":[]}`), []byte(`{"type":"reasoning","id":"rs","encrypted_content":"x","summary":[],"extra":1}`), []byte(`{"type":"reasoning","id":"rs","encrypted_content":"x","summary":null}`), []byte(`{"type":"reasoning","id":"rs","encrypted_content":"\ud800","summary":[]}`), []byte(`{"type":"reasoning","id":"rs","encrypted_content":"x","summary":[],"status":"in_progress"}`)} {
		if _, e := Encode([]Item{{Raw: raw}}); e == nil {
			t.Fatal("accepted", string(raw))
		}
	}
	for _, items := range [][]Item{nil, {{OutputIndex: -1, Raw: synthetic}}, {{OutputIndex: 2, Raw: synthetic}, {OutputIndex: 1, Raw: synthetic}}, {{OutputIndex: 1, Raw: synthetic}, {OutputIndex: 1, Raw: synthetic}}} {
		if _, e := Encode(items); e == nil {
			t.Fatal("invalid sequence")
		}
	}
	for _, data := range []string{"moai_opaque_v2_x", CarrierPrefix + "=", CarrierPrefix + base64.RawURLEncoding.EncodeToString([]byte(`{"version":1,"version":1}`)), CarrierPrefix + base64.RawURLEncoding.EncodeToString([]byte(`{"version":2,"provider":"openai","items":[]}`))} {
		if _, e := Decode(data); e == nil || errors.Is(e, ErrNotEnvelope) {
			t.Fatal("malformed reserved carrier", e)
		}
	}
}
func TestMarkerMalformedAndLimits(t *testing.T) {
	e, _ := Encode([]Item{{Raw: synthetic}})
	for _, id := range []string{"", strings.Repeat("a", MaxIDBytes+1), "call with spaces", "call/unsafe", "toolu_moai_v1_x"} {
		if _, err := BindToolID(id, e); err == nil {
			t.Fatal("invalid ID", id)
		}
	}
	for _, marker := range []string{"call_plain", ToolPrefix + "=", ToolPrefix + base64.RawURLEncoding.EncodeToString([]byte(`{"call_id":"call","opaque_sha256":"bad"}`))} {
		if _, err := RestoreToolID(marker, e); err == nil {
			t.Fatal("invalid marker")
		}
	}
	if _, err := Encode([]Item{{Raw: bytes.Repeat([]byte("x"), MaxItemBytes+1)}}); err == nil {
		t.Fatal("size")
	}
}

func TestReservedPrefixesAndDuplicateItemIDs(t *testing.T) {
	for _, s := range []string{"moai_opaque", "moai_opaque_v1", "moai_opaque_v9_invalid"} {
		if _, err := Decode(s); err == nil || errors.Is(err, ErrNotEnvelope) {
			t.Fatal("reserved prefix treated as native")
		}
	}
	e, _ := Encode([]Item{{Raw: synthetic}})
	if _, err := BindToolID("toolu_moai_v2_reserved", e); err == nil {
		t.Fatal("reserved tool marker rebound")
	}
	if _, err := Encode([]Item{{OutputIndex: 1, Raw: synthetic}, {OutputIndex: 3, Raw: synthetic}}); err == nil {
		t.Fatal("duplicate item ID")
	}
}

func TestReasoningShapesAndOrder(t *testing.T) {
	for _, content := range []string{`[]`, `[{"type":"reasoning_text","text":"synthetic"}]`, `[{"type":"text","text":"synthetic"}]`} {
		raw := []byte(`{"id":"rs_other","type":"reasoning","encrypted_content":"synthetic","summary":[{"type":"summary_text","text":"summary"}],"content":` + content + `,"status":"completed"}`)
		e, err := Encode([]Item{{OutputIndex: 0, Raw: synthetic}, {OutputIndex: 3, Raw: raw}})
		if err != nil {
			t.Fatal(err)
		}
		d, err := Decode(e.Data())
		if err != nil || len(d.Items()) != 2 || d.Items()[1].OutputIndex != 3 || !bytes.Equal(d.Items()[1].Raw, raw) {
			t.Fatal("interleaving lost")
		}
	}
}

func TestEnvelopeMalformedAndOwnership(t *testing.T) {
	e, _ := Encode([]Item{{Raw: synthetic}})
	raw := e.Bytes()
	raw[0] = 'x'
	if bytes.Equal(raw, e.Bytes()) {
		t.Fatal("bytes alias")
	}
	var nilEnvelope *Envelope
	if nilEnvelope.Bytes() != nil || nilEnvelope.Items() != nil || nilEnvelope.Data() != "" || nilEnvelope.Digest() != "" {
		t.Fatal("nil accessors")
	}
	for _, raw := range []string{`null`, `[]`, `{} {}`, `{"version":1,"provider":"openai","items":null}`, `{"version":1,"provider":"openai","extra":1}`, `{"version":1,"provider":"other","items":[]}`, `{"version":1,"provider":"openai","items":[{"output_index":null,"raw":"x"}]}`, `{"version":1,"provider":"openai","items":[{"output_index":0,"raw":null}]}`, `{"version":1,"provider":"openai","items":[{"output_index":0,"raw":"x","extra":1}]}`, `{"version":1,"provider":"openai","items":[{"output_index":-1,"raw":"x"}]}`, `{"version":1,"provider":"openai","items":[{"output_index":0,"raw":"x"}]}`, " " + string(e.Bytes())} {
		if _, err := Decode(CarrierPrefix + base64.RawURLEncoding.EncodeToString([]byte(raw))); err == nil {
			t.Fatal("bad envelope")
		}
	}
	for _, raw := range []string{`{"call_id":"call","opaque_sha256":"x","extra":1}`, `{"call_id":null,"opaque_sha256":"x"}`, `{"call_id":"call","call_id":"call"}`, " " + `{"call_id":"call","opaque_sha256":"` + e.Digest() + `"}`} {
		if _, err := RestoreToolID(ToolPrefix+base64.RawURLEncoding.EncodeToString([]byte(raw)), e); err == nil {
			t.Fatal("bad binding")
		}
	}
	if _, err := Decode(CarrierPrefix + strings.Repeat("x", base64.RawURLEncoding.EncodedLen(MaxEnvelopeBytes)+1)); err == nil {
		t.Fatal("oversize encoded carrier")
	}
}
func TestStrictJSONAndReasoningSubfields(t *testing.T) {
	for _, raw := range []string{`{`, `{"x":}`, `{"x":1`, `{"x":[1}`, `{"x":{"a":1,"a":2}}`, `{"x":"\uZZZZ"}`, `{"x":"\udc00"}`, `{"x":"\ud800x"}`, `{"x":"\ud800\u0000"}`, `{"x":"\ud800\uZZZZ"}`, `{"x":"\u12"}`, `{"x":"\`, strings.Repeat(`{"x":`, 66) + `0` + strings.Repeat(`}`, 66)} {
		if _, err := object([]byte(raw)); err == nil {
			t.Fatal("bad JSON", raw)
		}
	}
	if _, err := object([]byte("{\"x\":\"\xff\"}")); err == nil {
		t.Fatal("invalid UTF-8")
	}
	if _, err := object([]byte(`{"x":"\ud83d\ude00\\value\n"}`)); err != nil {
		t.Fatal("valid unicode", err)
	}
	for _, summary := range []string{`[{"type":"unknown","text":"x"}]`, `[{"type":"summary_text","text":null}]`, `[{"type":"summary_text","text":"x","extra":true}]`, `[null]`, `{}`} {
		raw := []byte(`{"id":"rs","type":"reasoning","encrypted_content":"x","summary":` + summary + `}`)
		if _, err := Encode([]Item{{Raw: raw}}); err == nil {
			t.Fatal("invalid summary")
		}
	}
}

func TestPublicLayoutV2CanonicalAndBounded(t *testing.T) {
	item := Item{OutputIndex: 1, Raw: []byte(`{"type":"reasoning","id":"rs_layout","summary":[],"encrypted_content":"synthetic"}`)}
	layout := []PublicItem{{OutputIndex: 0, Type: "message", Blocks: 2, PhaseNull: true}, {OutputIndex: 2, Type: "message", Blocks: 1, Phase: "final_answer"}}
	e, err := EncodeWithPublic([]Item{item}, layout)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(e.Data(), CarrierPrefixV2) {
		t.Fatal("layout not versioned")
	}
	decoded, err := Decode(e.Data())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(decoded.Bytes(), e.Bytes()) || len(decoded.PublicItems()) != 2 {
		t.Fatal("layout roundtrip")
	}
	copyLayout := decoded.PublicItems()
	copyLayout[0].Blocks = 999
	if decoded.PublicItems()[0].Blocks != 2 {
		t.Fatal("layout accessor aliases")
	}
	for _, bad := range [][]PublicItem{
		{{OutputIndex: 1, Type: "message", Blocks: 1}},
		{{OutputIndex: 0, Type: "message", Blocks: 0}, {OutputIndex: 2, Type: "message", Blocks: 1}},
		{{OutputIndex: 0, Type: "function_call", Blocks: 1, PhaseNull: true}, {OutputIndex: 2, Type: "message", Blocks: 1}},
		{{OutputIndex: 0, Type: "message", Blocks: 1, Phase: "invalid"}, {OutputIndex: 2, Type: "message", Blocks: 1}},
	} {
		if _, err := EncodeWithPublic([]Item{item}, bad); err == nil {
			t.Fatal("invalid layout accepted")
		}
	}
	mutated := bytes.Replace(e.Bytes(), []byte(`"blocks":2`), []byte(`"blocks":2,"extra":true`), 1)
	if _, err := Decode(CarrierPrefixV2 + base64.RawURLEncoding.EncodeToString(mutated)); err == nil {
		t.Fatal("unknown layout field accepted")
	}
	if _, err := Decode(CarrierPrefix + base64.RawURLEncoding.EncodeToString(e.Bytes())); err == nil {
		t.Fatal("version mismatch accepted")
	}
}
