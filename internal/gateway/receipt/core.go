// Package receipt stores hash-only evidence of completed conversation prefixes.
// UUID/root authorization and opaque codec validation belong to the caller; this
// package never discovers sessions from request metadata or recovers lost items.
package receipt

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"regexp"
)

const MaxCandidates = 4096
const MaxBytes = 8 << 20

var ErrInvalid = errors.New("invalid conversation receipt")
var ErrLimit = errors.New("conversation receipt limit exceeded")
var uuidPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

type Digest [32]byte

func Hash(b []byte) Digest                    { return sha256.Sum256(b) }
func (d Digest) MarshalJSON() ([]byte, error) { return json.Marshal(hex.EncodeToString(d[:])) }
func (d *Digest) UnmarshalJSON(raw []byte) error {
	var s string
	if json.Unmarshal(raw, &s) != nil || len(s) != 64 {
		return ErrInvalid
	}
	b, e := hex.DecodeString(s)
	if e != nil || hex.EncodeToString(b) != s {
		return ErrInvalid
	}
	copy(d[:], b)
	return nil
}

type Candidate struct {
	Prefix   Digest `json:"prefix"`
	Previous Digest `json:"previous"`
	Provider string `json:"provider"`
	Opaque   Digest `json:"opaque"`
	Items    uint32 `json:"items"`
	Required bool   `json:"opaque_required"`
	Complete bool   `json:"complete"`
}
type Observation struct {
	Prefix, Previous Digest
	Provider         string
	Opaque           Digest
	Items            uint32
}
type Manifest struct {
	session    Digest
	generation uint64
	candidates []Candidate
}
type diskManifest struct {
	Version    uint32      `json:"version"`
	Session    Digest      `json:"session"`
	Generation uint64      `json:"generation"`
	Count      uint32      `json:"count"`
	Candidates []Candidate `json:"candidates"`
	Checksum   Digest      `json:"checksum"`
}

func New(authorizedUUID string) (*Manifest, error) {
	if !uuidPattern.MatchString(authorizedUUID) {
		return nil, ErrInvalid
	}
	return &Manifest{session: Hash([]byte(authorizedUUID)), generation: 1, candidates: []Candidate{}}, nil
}
func valid(c Candidate) bool {
	return c.Prefix != (Digest{}) && c.Complete && (c.Provider == "openai" || c.Provider == "anthropic" || c.Provider == "zai") && ((c.Required && c.Provider == "openai" && c.Items > 0 && c.Opaque != (Digest{})) || (!c.Required && c.Items == 0 && c.Opaque == (Digest{})))
}
func (m *Manifest) Candidates() []Candidate { return append([]Candidate(nil), m.candidates...) }
func (m *Manifest) Publish(c Candidate) error {
	if m.session == (Digest{}) || m.generation == 0 || !valid(c) {
		return ErrInvalid
	}
	for _, old := range m.candidates {
		if old == c {
			return nil
		}
	}
	if len(m.candidates) >= MaxCandidates || m.generation == ^uint64(0) {
		return ErrLimit
	}
	m.candidates = append(m.candidates, c)
	m.generation++

	return nil
}

// Check validates every supplied assistant boundary. An empty history is a
// branch, not a request to insert other completed but unconsumed responses.
func (m *Manifest) Check(authorizedUUID string, history []Observation) error {
	if !uuidPattern.MatchString(authorizedUUID) || Hash([]byte(authorizedUUID)) != m.session {
		return ErrInvalid
	}
	for _, o := range history {
		found, required, empty := false, false, false
		for _, c := range m.candidates {
			if c.Prefix != o.Prefix || c.Previous != o.Previous {
				continue
			}
			required = required || c.Required
			if c.Provider != o.Provider {
				continue
			}
			empty = empty || !c.Required
			if c.Required && c.Opaque == o.Opaque && c.Items == o.Items {
				found = true
			}
		}
		if o.Opaque == (Digest{}) && o.Items == 0 {
			found = empty && !required
		}
		if !found {
			return ErrInvalid
		}
	}
	return nil
}
func (m *Manifest) Fork(authorizedChildUUID string) (*Manifest, error) {
	if m.session == (Digest{}) || m.generation == 0 {
		return nil, ErrInvalid
	}
	n, e := New(authorizedChildUUID)
	if e != nil {
		return nil, e
	}
	n.candidates = m.Candidates()
	if _, e = n.Marshal(); e != nil {
		return nil, e
	}
	return n, nil
}
func (m *Manifest) Marshal() ([]byte, error) {
	if len(m.candidates) > MaxCandidates {
		return nil, ErrLimit
	}
	d := diskManifest{Version: 1, Session: m.session, Generation: m.generation, Count: uint32(len(m.candidates)), Candidates: m.candidates}
	payload, e := json.Marshal(d)
	if e != nil {
		return nil, ErrInvalid
	}
	d.Checksum = Hash(payload)
	raw, e := json.Marshal(d)
	if e != nil {
		return nil, ErrInvalid
	}
	if len(raw) > MaxBytes {
		return nil, ErrLimit
	}
	return raw, nil
}
func Parse(raw []byte, authorizedUUID string) (*Manifest, error) {
	if len(raw) > MaxBytes {
		return nil, ErrLimit
	}
	m, e := New(authorizedUUID)
	if e != nil {
		return nil, e
	}
	object, e := decodeStrict(raw)
	if e != nil {
		return nil, e
	}
	fields, ok := object.(map[string]any)
	if !ok || len(fields) != 6 {
		return nil, ErrInvalid
	}
	for _, key := range []string{"version", "session", "generation", "count", "candidates", "checksum"} {
		if _, ok := fields[key]; !ok {
			return nil, ErrInvalid
		}
	}
	rows, ok := fields["candidates"].([]any)
	if !ok {
		return nil, ErrInvalid
	}
	for _, row := range rows {
		c, ok := row.(map[string]any)
		if !ok || len(c) != 7 {
			return nil, ErrInvalid
		}
		for _, key := range []string{"prefix", "previous", "provider", "opaque", "items", "opaque_required", "complete"} {
			if _, ok := c[key]; !ok {
				return nil, ErrInvalid
			}
		}
	}
	var d diskManifest
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if dec.Decode(&d) != nil {
		return nil, ErrInvalid
	}
	var extra any
	if dec.Decode(&extra) != io.EOF {
		return nil, ErrInvalid
	}
	if d.Version != 1 || d.Session != m.session || d.Generation == 0 || d.Count != uint32(len(d.Candidates)) || d.Candidates == nil {
		return nil, ErrInvalid
	}
	if len(d.Candidates) > MaxCandidates {
		return nil, ErrLimit
	}
	checksum := d.Checksum
	d.Checksum = Digest{}
	payload, _ := json.Marshal(d)
	if Hash(payload) != checksum {
		return nil, ErrInvalid
	}
	seen := make(map[Candidate]bool, len(d.Candidates))
	for _, c := range d.Candidates {
		if !valid(c) || seen[c] {
			return nil, ErrInvalid
		}
		seen[c] = true
	}
	m.generation = d.Generation
	m.candidates = d.Candidates
	return m, nil
}
