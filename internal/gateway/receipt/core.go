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

// Fork copies the completed chain up to the current lastTurnId boundary into
// a child manifest. It is ForkAt at the parent's latest completed candidate.
func (m *Manifest) Fork(authorizedChildUUID string) (*Manifest, error) {
	if m.session == (Digest{}) || m.generation == 0 {
		return nil, ErrInvalid
	}
	if len(m.candidates) == 0 {
		return m.uncheckedFork(authorizedChildUUID, nil)
	}
	return m.ForkAt(authorizedChildUUID, m.candidates[len(m.candidates)-1].Prefix)
}

// ForkAt copies only the completed prefix chain up to the exact boundary
// candidate into a child manifest. Candidates completed after the boundary —
// including anything the parent completes later — are never copied, and an
// unknown, zero or structurally broken boundary is rejected before any child
// state exists.
func (m *Manifest) ForkAt(authorizedChildUUID string, boundary Digest) (*Manifest, error) {
	if m.session == (Digest{}) || m.generation == 0 {
		return nil, ErrInvalid
	}
	chain, e := m.ChainTo(boundary)
	if e != nil {
		return nil, e
	}
	return m.uncheckedFork(authorizedChildUUID, chain)
}

func (m *Manifest) uncheckedFork(authorizedChildUUID string, chain []Candidate) (*Manifest, error) {
	n, e := New(authorizedChildUUID)
	if e != nil {
		return nil, e
	}
	n.candidates = chain
	if _, e = n.Marshal(); e != nil {
		return nil, e
	}
	return n, nil
}

// ChainTo resolves the boundary candidate and walks its Previous links back
// through the stored chain, returning every candidate whose prefix lies on
// that walk. A Previous link that points at a prefix with no completed
// candidate is a chain root, not an error — evidence before it is simply not
// in this manifest. A boundary with no completed candidate, a cycle, or
// disagreeing links for one prefix is rejected.
func (m *Manifest) ChainTo(boundary Digest) ([]Candidate, error) {
	if m.session == (Digest{}) || m.generation == 0 || boundary == (Digest{}) {
		return nil, ErrInvalid
	}
	visited := map[Digest]bool{}
	cur := boundary
	for {
		if visited[cur] {
			return nil, ErrInvalid // cycle
		}
		visited[cur] = true
		var previous *Digest
		for i := range m.candidates {
			if m.candidates[i].Prefix != cur {
				continue
			}
			if previous != nil && *previous != m.candidates[i].Previous {
				return nil, ErrInvalid // disagreeing chain links
			}
			d := m.candidates[i].Previous
			previous = &d
		}
		if previous == nil {
			return nil, ErrInvalid // boundary has no completed candidate
		}
		if *previous == (Digest{}) || !m.hasCandidate(*previous) {
			break // chain root: dangling link or empty previous
		}
		cur = *previous
	}
	chain := make([]Candidate, 0, len(visited))
	for _, c := range m.candidates {
		if visited[c.Prefix] {
			chain = append(chain, c)
		}
	}
	return chain, nil
}

func (m *Manifest) hasCandidate(prefix Digest) bool {
	for _, c := range m.candidates {
		if c.Prefix == prefix {
			return true
		}
	}
	return false
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
