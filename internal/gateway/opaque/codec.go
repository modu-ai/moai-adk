// Package opaque carries ordered OpenAI reasoning without interpreting encrypted
// content. Its hashes establish consistency, never authenticity or session authority.
package opaque

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

const (
	CarrierPrefix    = "moai_opaque_v1_"
	CarrierPrefixV2  = "moai_opaque_v2_"
	ToolPrefix       = "toolu_moai_v1_"
	MaxEnvelopeBytes = 8 << 20
	MaxItemBytes     = 4 << 20
	MaxItems         = 1024
	MaxOutputIndex   = 100000
	MaxIDBytes       = 1024
)

var ErrNotEnvelope = errors.New("not a MoAI opaque carrier")
var errInvalid = errors.New("invalid MoAI opaque encoding")

// Item.OutputIndex is the original Responses output array index, including
// public items. Raw is preserved byte-for-byte, including its JSON whitespace.
type Item struct {
	OutputIndex int
	Raw         []byte
}
type wireItem struct {
	OutputIndex int    `json:"output_index"`
	Raw         string `json:"raw"`
}

// PublicItem preserves public item boundaries, not public message contents.
// It is included in the envelope digest and requires the same receipt authority.
type PublicItem struct {
	OutputIndex int    `json:"output_index"`
	Type        string `json:"type"`
	Blocks      int    `json:"blocks"`
	Phase       string `json:"phase,omitempty"`
	PhaseNull   bool   `json:"phase_null,omitempty"`
}
type wireEnvelope struct {
	Version  int          `json:"version"`
	Provider string       `json:"provider"`
	Items    []wireItem   `json:"items"`
	Public   []PublicItem `json:"public_items,omitempty"`
}

// Envelope owns immutable validated bytes. Accessors return independent copies.
type Envelope struct {
	raw    []byte
	items  []Item
	public []PublicItem
}

func (e *Envelope) Bytes() []byte {
	if e == nil {
		return nil
	}
	return bytes.Clone(e.raw)
}
func (e *Envelope) Data() string {
	if e == nil {
		return ""
	}
	prefix := CarrierPrefix
	if len(e.public) > 0 {
		prefix = CarrierPrefixV2
	}
	return prefix + base64.RawURLEncoding.EncodeToString(e.raw)
}
func (e *Envelope) Digest() string {
	if e == nil {
		return ""
	}
	h := sha256.Sum256(e.raw)
	return hex.EncodeToString(h[:])
}
func (e *Envelope) Items() []Item {
	if e == nil {
		return nil
	}
	out := make([]Item, len(e.items))
	for i, v := range e.items {
		out[i] = Item{v.OutputIndex, bytes.Clone(v.Raw)}
	}
	return out
}

func (e *Envelope) PublicItems() []PublicItem {
	if e == nil {
		return nil
	}
	return append([]PublicItem(nil), e.public...)
}

// Encode validates and snapshots an ordered opaque-only projection.
func Encode(items []Item) (*Envelope, error) { return EncodeWithPublic(items, nil) }

// EncodeWithPublic binds the original public output boundaries to opaque items.
func EncodeWithPublic(items []Item, public []PublicItem) (*Envelope, error) {
	if len(items) == 0 || len(items) > MaxItems {
		return nil, errInvalid
	}
	wire := wireEnvelope{Version: 1, Provider: "openai", Items: make([]wireItem, len(items))}
	ids := map[string]bool{}
	last := -1
	total := 0
	for i, v := range items {
		if v.OutputIndex <= last || v.OutputIndex > MaxOutputIndex {
			return nil, errInvalid
		}
		last = v.OutputIndex
		id, err := validateItem(v.Raw)
		if err != nil || ids[id] {
			return nil, errInvalid
		}
		ids[id] = true
		total += len(v.Raw)
		if total > MaxEnvelopeBytes {
			return nil, errInvalid
		}
		wire.Items[i] = wireItem{v.OutputIndex, string(v.Raw)}
	}
	if len(public) > 0 {
		if len(public)+len(items) > MaxItems {
			return nil, errInvalid
		}
		positions := map[int]bool{}
		for _, v := range items {
			positions[v.OutputIndex] = true
		}
		previous := -1
		for _, v := range public {
			if v.OutputIndex <= previous || v.OutputIndex >= len(public)+len(items) || positions[v.OutputIndex] || v.Blocks < 1 || v.Blocks > MaxItems || (v.Type != "message" && v.Type != "function_call") || (v.Type == "function_call" && (v.Blocks != 1 || v.Phase != "")) || (v.Phase != "" && v.Phase != "commentary" && v.Phase != "final_answer") || (v.PhaseNull && (v.Phase != "" || v.Type != "message")) {
				return nil, errInvalid
			}
			previous = v.OutputIndex
			positions[v.OutputIndex] = true
		}
		for i := 0; i < len(public)+len(items); i++ {
			if !positions[i] {
				return nil, errInvalid
			}
		}
		wire.Version = 2
		wire.Public = append([]PublicItem(nil), public...)
	}
	raw, err := json.Marshal(wire)
	if err != nil || len(raw) > MaxEnvelopeBytes {
		return nil, errInvalid
	}
	e := &Envelope{raw: raw, items: make([]Item, len(items)), public: append([]PublicItem(nil), public...)}
	for i, v := range items {
		e.items[i] = Item{v.OutputIndex, bytes.Clone(v.Raw)}
	}
	return e, nil
}

// Decode distinguishes native data from malformed reserved MoAI discriminators.
func Decode(data string) (*Envelope, error) {
	prefix := CarrierPrefix
	version := 1
	if strings.HasPrefix(data, CarrierPrefixV2) {
		prefix = CarrierPrefixV2
		version = 2
	}
	if !strings.HasPrefix(data, prefix) {
		if strings.HasPrefix(data, "moai_opaque") {
			return nil, errInvalid
		}
		return nil, ErrNotEnvelope
	}
	encoded := strings.TrimPrefix(data, prefix)
	if len(encoded) > base64.RawURLEncoding.EncodedLen(MaxEnvelopeBytes) {
		return nil, errInvalid
	}
	raw, err := base64.RawURLEncoding.Strict().DecodeString(encoded)
	if err != nil || base64.RawURLEncoding.EncodeToString(raw) != encoded {
		return nil, errInvalid
	}
	m, err := object(raw)
	if err != nil || (!exact(m, "version", "provider", "items") && !exact(m, "version", "provider", "items", "public_items")) {
		return nil, errInvalid
	}
	var w wireEnvelope
	if json.Unmarshal(raw, &w) != nil || w.Version != version || (version == 2 && len(w.Public) == 0) || (version == 1 && len(w.Public) > 0) || w.Provider != "openai" || len(w.Items) == 0 || len(w.Items) > MaxItems {
		return nil, errInvalid
	}
	var entries []json.RawMessage
	if json.Unmarshal(m["items"], &entries) != nil {
		return nil, errInvalid
	}
	items := make([]Item, len(w.Items))
	for i, x := range entries {
		fields, err := object(x)
		if err != nil || !exact(fields, "output_index", "raw") {
			return nil, errInvalid
		}
		if bytes.Equal(fields["output_index"], []byte("null")) || bytes.Equal(fields["raw"], []byte("null")) {
			return nil, errInvalid
		}
		items[i] = Item{w.Items[i].OutputIndex, []byte(w.Items[i].Raw)}
	}
	e, err := EncodeWithPublic(items, w.Public)
	if err != nil {
		return nil, err
	}
	// Alternate envelope spellings are rejected so all marker hashes are canonical.
	if !bytes.Equal(e.raw, raw) {
		return nil, errInvalid
	}
	return e, nil
}
func validID(id string) bool {
	if len(id) == 0 || len(id) > MaxIDBytes || strings.HasPrefix(id, "toolu_moai_") {
		return false
	}
	for _, c := range id {
		idRune := c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_' || c == '-'
		if !idRune {
			return false
		}
	}
	return true
}
func validateItem(raw []byte) (string, error) {
	if len(raw) > MaxItemBytes {
		return "", errInvalid
	}
	m, err := object(raw)
	if err != nil {
		return "", err
	}
	for k := range m {
		switch k {
		case "type", "id", "encrypted_content", "summary", "content", "status":
		default:
			return "", errInvalid
		}
	}
	var typ, id, encrypted string
	if json.Unmarshal(m["type"], &typ) != nil || typ != "reasoning" || json.Unmarshal(m["id"], &id) != nil || !validID(id) || json.Unmarshal(m["encrypted_content"], &encrypted) != nil || encrypted == "" {
		return "", errInvalid
	}
	if status, ok := m["status"]; ok {
		var s string
		if json.Unmarshal(status, &s) != nil || s != "completed" {
			return "", errInvalid
		}
	}
	for _, key := range []string{"summary", "content"} {
		raw, ok := m[key]
		if !ok {
			if key == "summary" {
				return "", errInvalid
			}
			continue
		}
		var list []json.RawMessage
		if json.Unmarshal(raw, &list) != nil || list == nil {
			return "", errInvalid
		}
		for _, v := range list {
			entry, err := object(v)
			if err != nil || !exact(entry, "type", "text") {
				return "", errInvalid
			}
			var kind, text string
			if json.Unmarshal(entry["type"], &kind) != nil || json.Unmarshal(entry["text"], &text) != nil || bytes.Equal(entry["text"], []byte("null")) {
				return "", errInvalid
			}
			want := "summary_text"
			if key == "content" {
				want = "reasoning_text"
			}
			contentFallback := key == "content" && kind == "text"
			if kind != want && !contentFallback {
				return "", errInvalid
			}
		}
	}
	return id, nil
}

type binding struct {
	CallID string `json:"call_id"`
	Digest string `json:"opaque_sha256"`
}

// BindToolID binds the original provider ID to one exact envelope digest.
func BindToolID(callID string, e *Envelope) (string, error) {
	if !validID(callID) || e == nil || len(e.raw) == 0 {
		return "", errInvalid
	}
	raw, _ := json.Marshal(binding{callID, e.Digest()})
	return ToolPrefix + base64.RawURLEncoding.EncodeToString(raw), nil
}

// RestoreToolID checks envelope consistency, not receipt/session authorization.
func RestoreToolID(marker string, e *Envelope) (string, error) {
	if e == nil || !strings.HasPrefix(marker, ToolPrefix) || len(marker) > 4096 {
		return "", errInvalid
	}
	encoded := strings.TrimPrefix(marker, ToolPrefix)
	raw, err := base64.RawURLEncoding.Strict().DecodeString(encoded)
	if err != nil {
		return "", errInvalid
	}
	m, err := object(raw)
	if err != nil || !exact(m, "call_id", "opaque_sha256") {
		return "", errInvalid
	}
	var b binding
	if json.Unmarshal(raw, &b) != nil || !validID(b.CallID) || len(b.Digest) != 64 || b.Digest != e.Digest() {
		return "", errInvalid
	}
	canonical, _ := BindToolID(b.CallID, e)
	if marker != canonical {
		return "", errInvalid
	}
	return b.CallID, nil
}
func exact(m map[string]json.RawMessage, keys ...string) bool {
	if len(m) != len(keys) {
		return false
	}
	for _, k := range keys {
		if _, ok := m[k]; !ok {
			return false
		}
	}
	return true
}

// Escaped UTF-16 must not be silently replaced by encoding/json.
func unicodeEscapes(b []byte) bool {
	for i := 0; i < len(b); i++ {
		if b[i] != '\\' {
			continue
		}
		i++
		if i >= len(b) {
			return false
		}
		if b[i] != 'u' {
			continue
		}
		if i+4 >= len(b) {
			return false
		}
		n, e := strconv.ParseUint(string(b[i+1:i+5]), 16, 16)
		if e != nil {
			return false
		}
		i += 4
		if n >= 0xDC00 && n <= 0xDFFF {
			return false
		}
		if n >= 0xD800 && n <= 0xDBFF {
			if i+6 >= len(b) || b[i+1] != '\\' || b[i+2] != 'u' {
				return false
			}
			n, e = strconv.ParseUint(string(b[i+3:i+7]), 16, 16)
			if e != nil || n < 0xDC00 || n > 0xDFFF {
				return false
			}
			i += 6
		}
	}
	return true
}
