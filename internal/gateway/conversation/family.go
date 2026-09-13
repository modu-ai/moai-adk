// Package conversation owns the small, private index that binds native Claude
// transcripts to MoAI conversation families. It deliberately stores no text,
// tokens, or opaque provider payloads.
package conversation

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/modu-ai/moai-adk/internal/gateway/receipt"
)

var (
	ErrInvalid        = errors.New("invalid conversation family")
	ErrMissing        = errors.New("conversation family not found")
	ErrBusy           = errors.New("conversation family is busy")
	ErrIncomplete     = errors.New("native transcript is incomplete")
	ErrAmbiguous      = errors.New("conversation completion is ambiguous")
	ErrPrefixMismatch = errors.New("fork inherited prefix does not match the recorded receipt chain")
)

const maxIndexBytes = 8 << 20

type NewRequest struct {
	CWD, Project, SecureStorage, OriginalConfig string
	SecureStorageSet, OriginalConfigSet         bool
}
type Descriptor struct {
	FamilyID   string   `json:"family_id"`
	UUID       string   `json:"uuid"`
	Project    string   `json:"project"`
	CWD        string   `json:"cwd"`
	ConfigDir  string   `json:"config_dir"`
	ReceiptDir string   `json:"receipt_dir"`
	Transcript string   `json:"transcript"`
	Args       []string `json:"args"`
	Model      string   `json:"model,omitempty"`
}
type Selection struct {
	FamilyID, UUID, Project string
	Completion              uint64
}
type record struct {
	FamilyID   string `json:"family_id"`
	UUID       string `json:"uuid"`
	Project    string `json:"project"`
	CWD        string `json:"cwd"`
	ConfigDir  string `json:"config_dir"`
	ReceiptDir string `json:"receipt_dir"`
	Transcript string `json:"transcript,omitempty"`
	Completion uint64 `json:"completion,omitempty"`
}
type diskIndex struct {
	Version uint32   `json:"version"`
	Entries []record `json:"entries"`
}

type Manager struct {
	root, index string
	mu          sync.Mutex
	entries     map[string]record
}

func Open(root string) (*Manager, error) {
	if !filepath.IsAbs(root) || filepath.Clean(root) != root {
		return nil, ErrInvalid
	}
	if err := os.MkdirAll(root, 0700); err != nil {
		return nil, ErrInvalid
	}
	resolved, err := filepath.EvalSymlinks(root)
	if err != nil || !filepath.IsAbs(resolved) {
		return nil, ErrInvalid
	}
	root = resolved
	if err := privateDir(root); err != nil {
		return nil, err
	}
	m := &Manager{root: root, index: filepath.Join(root, "index.json"), entries: map[string]record{}}
	raw, err := os.ReadFile(m.index)
	if errors.Is(err, os.ErrNotExist) {
		return m, nil
	}
	if err != nil || len(raw) > maxIndexBytes {
		return nil, ErrInvalid
	}
	var d diskIndex
	if json.Unmarshal(raw, &d) != nil || d.Version != 1 {
		return nil, ErrInvalid
	}
	for _, e := range d.Entries {
		if !validRecord(m.root, e) {
			return nil, ErrInvalid
		}
		if _, ok := m.entries[e.UUID]; ok {
			return nil, ErrInvalid
		}
		m.entries[e.UUID] = e
	}
	return m, nil
}

func (m *Manager) New(ctx context.Context, req NewRequest) (Descriptor, error) {
	if err := ctx.Err(); err != nil {
		return Descriptor{}, err
	}
	if err := validateRequest(req); err != nil {
		return Descriptor{}, err
	}
	id, err := newUUID()
	if err != nil {
		return Descriptor{}, err
	}
	family := id
	config := filepath.Join(m.root, "families", family, "native")
	recRoot := filepath.Join(m.root, "families", family, "receipt")
	if err = os.MkdirAll(filepath.Join(m.root, "families"), 0700); err != nil {
		return Descriptor{}, ErrInvalid
	}
	_ = os.Chmod(filepath.Join(m.root, "families"), 0700)
	if err = os.MkdirAll(config, 0700); err != nil {
		return Descriptor{}, ErrInvalid
	}
	if err = os.MkdirAll(recRoot, 0700); err != nil {
		return Descriptor{}, ErrInvalid
	}
	for _, p := range []string{filepath.Dir(config), config, recRoot} {
		if err = os.Chmod(p, 0700); err != nil {
			return Descriptor{}, ErrInvalid
		}
	}
	rs, err := receipt.OpenStore(ctx, recRoot, id, true)
	if err != nil {
		return Descriptor{}, err
	}
	_ = rs.Close()
	r := record{FamilyID: family, UUID: id, Project: req.Project, CWD: req.CWD, ConfigDir: config, ReceiptDir: recRoot}
	m.mu.Lock()
	defer m.mu.Unlock()
	if err = m.addLocked(r); err != nil {
		return Descriptor{}, err
	}
	return descriptor(r, nil), nil
}

func (m *Manager) Resume(ctx context.Context, id string) (Descriptor, error) {
	if err := m.refreshNative(ctx, id); err != nil {
		return Descriptor{}, err
	}
	if err := ctx.Err(); err != nil {
		return Descriptor{}, err
	}
	m.mu.Lock()
	r, ok := m.entries[id]
	m.mu.Unlock()
	if !ok {
		return Descriptor{}, ErrMissing
	}
	if r.Completion == 0 || r.Transcript == "" {
		return Descriptor{}, ErrIncomplete
	}
	if err := validateTranscript(r); err != nil {
		return Descriptor{}, err
	}
	store, err := receipt.OpenStore(ctx, r.ReceiptDir, r.UUID, false)
	if err != nil {
		return Descriptor{}, err
	}
	defer func() { _ = store.Close() }() // read-only snapshot session; the lock ends at process exit too
	if _, err = store.Snapshot(ctx); err != nil {
		return Descriptor{}, err
	}
	d := descriptor(r, []string{"--resume", r.UUID})
	d.Model, _ = transcriptModel(r)
	return d, nil
}

func (m *Manager) Continue(ctx context.Context, project string) (Descriptor, error) {
	if err := m.refreshNativeProject(ctx, project); err != nil {
		return Descriptor{}, err
	}
	if err := ctx.Err(); err != nil {
		return Descriptor{}, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	var found *record
	for _, r := range m.entries {
		if r.Project != project || r.Completion == 0 {
			continue
		}
		if err := validateTranscript(r); err != nil {
			continue
		}
		if found != nil && found.Completion == r.Completion {
			return Descriptor{}, ErrAmbiguous
		}
		if found == nil || r.Completion > found.Completion {
			x := r
			found = &x
		}
	}
	if found == nil {
		return Descriptor{}, ErrMissing
	}
	d := descriptor(*found, []string{"--resume", found.UUID})
	d.Model, _ = transcriptModel(*found)
	return d, nil
}

func (m *Manager) Select(ctx context.Context, project string) ([]Selection, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	out := []Selection{}
	for _, r := range m.entries {
		if r.Project == project && r.Completion > 0 && validateTranscript(r) == nil {
			out = append(out, Selection{r.FamilyID, r.UUID, r.Project, r.Completion})
		}
	}
	return out, nil
}

func (m *Manager) Complete(ctx context.Context, id, transcript string, sequence uint64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if sequence == 0 {
		return ErrInvalid
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.entries[id]
	if !ok {
		return ErrMissing
	}
	rel, err := filepath.Rel(r.ConfigDir, transcript)
	if err != nil || rel == "." || filepath.IsAbs(rel) || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return ErrInvalid
	}
	r.Transcript = filepath.ToSlash(rel)
	r.Completion = sequence
	if err = validateTranscript(r); err != nil {
		return err
	}
	return m.replaceLocked(id, r)
}

// Fork branches the parent at its current lastTurnId: the launcher's
// --fork-session boundary.
func (m *Manager) Fork(ctx context.Context, parentID string) (d Descriptor, err error) {
	p, err := m.Resume(ctx, parentID)
	if err != nil {
		return Descriptor{}, err
	}
	parentRoot := filepath.Join(m.root, "families", p.FamilyID, "receipt")
	ps, err := receipt.OpenStore(ctx, parentRoot, p.UUID, false)
	if err != nil {
		return Descriptor{}, err
	}
	defer func() { _ = ps.Close() }() // read-only parent snapshot; the lock ends at process exit too
	snap, err := ps.Snapshot(ctx)
	if err != nil {
		return Descriptor{}, err
	}
	chain := snap.Candidates()
	if len(chain) > 0 {
		chain, err = snap.ChainTo(chain[len(chain)-1].Prefix)
		if err != nil {
			return Descriptor{}, err
		}
	}
	return m.forkAtCandidates(ctx, p, chain)
}

// ForkAt branches the parent at the exact completedTurnID boundary: only the
// completed chain up to that boundary is copied into the new family/thread,
// and a parent that keeps completing turns can never leak post-boundary facts
// into the child. An unknown origin, or an unknown, zero or broken boundary,
// is rejected before any child state exists.
func (m *Manager) ForkAt(ctx context.Context, parentID string, boundary receipt.Digest) (d Descriptor, err error) {
	p, err := m.Resume(ctx, parentID)
	if err != nil {
		return Descriptor{}, err
	}
	parentRoot := filepath.Join(m.root, "families", p.FamilyID, "receipt")
	ps, err := receipt.OpenStore(ctx, parentRoot, p.UUID, false)
	if err != nil {
		return Descriptor{}, err
	}
	defer func() { _ = ps.Close() }() // read-only parent snapshot; the lock ends at process exit too
	snap, err := ps.Snapshot(ctx)
	if err != nil {
		return Descriptor{}, err
	}
	chain, err := snap.ChainTo(boundary)
	if err != nil {
		return Descriptor{}, err
	}
	return m.forkAtCandidates(ctx, p, chain)
}

// ForkSession branches the parent at the caller-named boundary only when the
// claimed inherited prefix matches the parent's completed chain (AC-MG-026
// (c)). The contrast runs against Manifest.ChainTo at the boundary before any
// child state exists, so tampered, mismatched and unknown-origin claims leave
// no fork directory behind.
func (m *Manager) ForkSession(ctx context.Context, parentID string, boundary, claimed receipt.Digest) (d Descriptor, err error) {
	p, err := m.Resume(ctx, parentID)
	if err != nil {
		return Descriptor{}, err
	}
	parentRoot := filepath.Join(m.root, "families", p.FamilyID, "receipt")
	ps, err := receipt.OpenStore(ctx, parentRoot, p.UUID, false)
	if err != nil {
		return Descriptor{}, err
	}
	defer func() { _ = ps.Close() }() // read-only parent snapshot; the lock ends at process exit too
	snap, err := ps.Snapshot(ctx)
	if err != nil {
		return Descriptor{}, err
	}
	chain, err := snap.ChainTo(boundary)
	if err != nil {
		return Descriptor{}, err
	}
	if receipt.ChainDigest(chain) != claimed {
		return Descriptor{}, ErrPrefixMismatch
	}
	return m.forkAtCandidates(ctx, p, chain)
}

func (m *Manager) forkAtCandidates(ctx context.Context, p Descriptor, chain []receipt.Candidate) (d Descriptor, err error) {
	id, err := newUUID()
	if err != nil {
		return Descriptor{}, err
	}
	recRoot := filepath.Join(m.root, "families", p.FamilyID, "forks", id, "receipt")
	config := p.ConfigDir
	if err = os.MkdirAll(recRoot, 0700); err != nil {
		return Descriptor{}, ErrInvalid
	}
	cs, err := receipt.OpenStore(ctx, recRoot, id, true)
	if err != nil {
		return Descriptor{}, err
	}
	// The fork publishes into cs; a failed release must not report success.
	defer func() {
		if cerr := cs.Close(); err == nil {
			err = cerr
		}
	}()
	for _, c := range chain {
		if err = cs.Publish(ctx, c); err != nil {
			return Descriptor{}, err
		}
	}
	r := record{FamilyID: p.FamilyID, UUID: id, Project: p.Project, CWD: p.CWD, ConfigDir: config, ReceiptDir: recRoot}
	m.mu.Lock()
	defer m.mu.Unlock()
	if err = m.addLocked(r); err != nil {
		return Descriptor{}, err
	}
	d = descriptor(r, []string{"--resume", p.UUID, "--fork-session", "--session-id", id})
	d.Model = p.Model
	return d, nil
}

type Lease struct {
	key  string
	once sync.Once
}

var leases = struct {
	sync.Mutex
	held map[string]bool
}{held: map[string]bool{}}

func (m *Manager) AcquireLease(family string) (*Lease, error) {
	leases.Lock()
	defer leases.Unlock()
	key := filepath.Join(m.root, family)
	if leases.held[key] {
		return nil, ErrBusy
	}
	leases.held[key] = true
	return &Lease{key: key}, nil
}
func (l *Lease) Release() {
	if l == nil {
		return
	}
	l.once.Do(func() { leases.Lock(); delete(leases.held, l.key); leases.Unlock() })
}

func SelectSecureNamespace(secure, source string) string {
	if secure != "" {
		return secure
	}
	return source
}

// ResolveSecureNamespace preserves an explicitly defined secure namespace,
// including an explicit empty override. An explicitly empty source config is
// rejected because its compatibility has not been established.
func ResolveSecureNamespace(secure string, secureSet bool, source string, sourceSet bool) (string, error) {
	if secureSet {
		return secure, nil
	}
	if sourceSet {
		if source == "" {
			return "", ErrInvalid
		}
		return source, nil
	}
	return "", nil
}
func ValidateSecureNamespace(source string, explicitlySet bool) error {
	if explicitlySet && source == "" {
		return ErrInvalid
	}
	return nil
}

func (m *Manager) addLocked(r record) error {
	if !validRecord(m.root, r) || r.Completion != 0 {
		return ErrInvalid
	}
	if _, ok := m.entries[r.UUID]; ok {
		return ErrInvalid
	}
	m.entries[r.UUID] = r
	if err := m.saveLocked(); err != nil {
		delete(m.entries, r.UUID)
		return err
	}
	return nil
}
func (m *Manager) replaceLocked(id string, r record) error {
	old := m.entries[id]
	m.entries[id] = r
	if err := m.saveLocked(); err != nil {
		m.entries[id] = old
		return err
	}
	return nil
}
func (m *Manager) saveLocked() error {
	d := diskIndex{Version: 1, Entries: make([]record, 0, len(m.entries))}
	for _, r := range m.entries {
		d.Entries = append(d.Entries, r)
	}
	raw, err := json.Marshal(d)
	if err != nil || len(raw) > maxIndexBytes {
		return ErrInvalid
	}
	tmp := m.index + ".tmp"
	_ = os.Remove(tmp)
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return ErrInvalid
	}
	_, err = f.Write(raw)
	if err == nil {
		err = f.Sync()
	}
	ce := f.Close()
	if err == nil {
		err = ce
	}
	if err == nil {
		err = os.Rename(tmp, m.index)
	}
	if err != nil {
		_ = os.Remove(tmp)
		return ErrInvalid
	}
	return nil
}
func descriptor(r record, args []string) Descriptor {
	return Descriptor{FamilyID: r.FamilyID, UUID: r.UUID, Project: r.Project, CWD: r.CWD, ConfigDir: r.ConfigDir, ReceiptDir: r.ReceiptDir, Transcript: r.Transcript, Args: args}
}
func validRecord(root string, r record) bool {
	return uuid(r.UUID) && uuid(r.FamilyID) && r.Project != "" && filepath.IsAbs(r.CWD) && filepath.IsAbs(r.ConfigDir) && filepath.IsAbs(r.ReceiptDir) && within(root, r.ConfigDir) && within(root, r.ReceiptDir) && (r.Transcript == "" || !filepath.IsAbs(r.Transcript))
}
func validateRequest(r NewRequest) error {
	if r.Project == "" || !filepath.IsAbs(r.CWD) || filepath.Clean(r.CWD) != r.CWD {
		return ErrInvalid
	}
	if err := ValidateSecureNamespace(r.OriginalConfig, r.OriginalConfigSet); err != nil {
		return err
	}
	return nil
}
func validateTranscript(r record) error { _, err := transcriptModel(r); return err }

func within(root, p string) bool {
	a, e := filepath.Abs(root)
	if e != nil {
		return false
	}
	b, e := filepath.Abs(p)
	if e != nil {
		return false
	}
	rel, e := filepath.Rel(a, b)
	return e == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
func privateDir(p string) error {
	i, e := os.Stat(p)
	if e != nil || !i.IsDir() {
		return ErrInvalid
	}
	if i.Mode().Perm()&0077 != 0 {
		_ = os.Chmod(p, 0700)
		i, e = os.Stat(p)
		if e != nil || i.Mode().Perm()&0077 != 0 {
			return ErrInvalid
		}
	}
	return nil
}
func uuid(s string) bool {
	return len(s) == 36 && s[8] == '-' && s[13] == '-' && s[18] == '-' && s[23] == '-'
}
func newUUID() (string, error) {
	var b [16]byte
	if _, e := rand.Read(b[:]); e != nil {
		return "", e
	}
	b[6] = (b[6] & 15) | 64
	b[8] = (b[8] & 63) | 128
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}
