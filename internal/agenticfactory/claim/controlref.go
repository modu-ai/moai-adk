package claim

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// ControlRef is the persistence boundary. The authoritative copy of claims,
// intake requests and agent numbers lives behind it (the online control ref
// and the Issue); the files under Store's directory are a cache that Rebuild
// can recreate from it. No remote implementation exists yet: it waits for the
// user's decision, and tests use MemoryRef.
type ControlRef interface {
	// Read returns the data and revision of key, or ErrNotHeld.
	Read(ctx context.Context, key string) (data []byte, rev int64, err error)
	// CompareAndSwap writes data when the key's revision is expectRev (0 means
	// the key must not exist) and returns the new revision. Any other revision
	// is ErrRevisionConflict.
	CompareAndSwap(ctx context.Context, key string, expectRev int64, data []byte) (int64, error)
	// Delete removes key when its revision is expectRev.
	Delete(ctx context.Context, key string, expectRev int64) error
	// Keys lists the keys that start with prefix, sorted.
	Keys(ctx context.Context, prefix string) ([]string, error)
}

// Attach makes ref the authority. Writes go to it first; nil detaches.
func (s *Store) Attach(ref ControlRef) { s.ref = ref }

func refCtx() context.Context { return context.Background() }

// refErr keeps a revision conflict recognisable and turns everything else the
// authority says into "store unavailable".
func refErr(err error) error {
	if errors.Is(err, core.ErrRevisionConflict) || errors.Is(err, core.ErrNotHeld) {
		return err
	}
	return fmt.Errorf("%w: authority: %v", core.ErrStoreUnavailable, err)
}

func (s *Store) cacheAgent(a core.Agent) error {
	if err := os.MkdirAll(s.agentsDir(), 0o700); err != nil {
		return fmt.Errorf("%w: %v", core.ErrStoreUnavailable, err)
	}
	b, _ := json.Marshal(a)
	tmp, err := os.CreateTemp(s.agentsDir(), "agent.tmp-*")
	if err != nil {
		return fmt.Errorf("%w: %v", core.ErrStoreUnavailable, err)
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	if _, err := tmp.Write(b); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("%w: %v", core.ErrStoreUnavailable, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("%w: %v", core.ErrStoreUnavailable, err)
	}
	if err := os.Rename(tmp.Name(), s.agentPath(a.Number)); err != nil {
		return fmt.Errorf("%w: %v", core.ErrStoreUnavailable, err)
	}
	return nil
}

// refAgents lists every agent the authority knows.
func (s *Store) refAgents() ([]core.Agent, error) {
	keys, err := s.ref.Keys(refCtx(), "agent/")
	if err != nil {
		return nil, refErr(err)
	}
	var out []core.Agent
	for _, k := range keys {
		data, _, err := s.ref.Read(refCtx(), k)
		if errors.Is(err, core.ErrNotHeld) {
			continue // released between the listing and the read
		}
		if err != nil {
			return nil, refErr(err)
		}
		var a core.Agent
		if err := json.Unmarshal(data, &a); err != nil {
			return nil, fmt.Errorf("%w: corrupt agent in authority: %v", core.ErrStoreUnavailable, err)
		}
		out = append(out, a)
	}
	return out, nil
}

// Rebuild recreates the cache from the authority. Everything is read and
// parsed first, so unreadable authority data changes nothing; the caller gets
// the number of entries restored.
func (s *Store) Rebuild(ctx context.Context) (int, error) {
	if s.ref == nil {
		return 0, fmt.Errorf("%w: no authority attached", core.ErrStoreUnavailable)
	}
	keys, err := s.ref.Keys(ctx, "")
	if err != nil {
		return 0, refErr(err)
	}
	type item struct {
		issue core.IssueRef
		rec   Record
		req   Request
		agent core.Agent
		kind  string
	}
	var items []item
	for _, k := range keys {
		kind, name, ok := strings.Cut(k, "/")
		if !ok {
			continue
		}
		data, rev, err := s.ref.Read(ctx, k)
		if errors.Is(err, core.ErrNotHeld) {
			continue
		}
		if err != nil {
			return 0, refErr(err)
		}
		it := item{kind: kind}
		switch kind {
		case "issue":
			var n int
			if _, err := fmt.Sscanf(name, "%d", &n); err != nil {
				return 0, fmt.Errorf("%w: authority key %q", core.ErrStoreUnavailable, k)
			}
			ref, rerr := core.NewIssueRef(n)
			if rerr != nil {
				return 0, fmt.Errorf("%w: authority key %q", core.ErrStoreUnavailable, k)
			}
			it.issue = ref
			if err := json.Unmarshal(data, &it.rec); err != nil {
				return 0, fmt.Errorf("%w: corrupt %s in authority: %v", core.ErrStoreUnavailable, k, err)
			}
			it.rec.Rev = rev
		case "request":
			if err := json.Unmarshal(data, &it.req); err != nil {
				return 0, fmt.Errorf("%w: corrupt %s in authority: %v", core.ErrStoreUnavailable, k, err)
			}
			it.req.Rev = rev
		case "agent":
			if err := json.Unmarshal(data, &it.agent); err != nil {
				return 0, fmt.Errorf("%w: corrupt %s in authority: %v", core.ErrStoreUnavailable, k, err)
			}
		default:
			continue
		}
		items = append(items, it)
	}
	for _, it := range items {
		var err error
		switch it.kind {
		case "issue":
			err = s.writeLocal(it.issue, it.rec)
		case "request":
			if err = os.MkdirAll(s.requestsDir(), 0o700); err == nil {
				err = s.writeRequestLocal(it.req)
			}
		case "agent":
			err = s.cacheAgent(it.agent)
		}
		if err != nil {
			return 0, err
		}
	}
	return len(items), nil
}

// MemoryRef is an in-memory ControlRef for tests and dry runs. It is not
// durable. Setting Fail makes every call fail, to exercise an unreachable or
// refusing authority.
type MemoryRef struct {
	mu   sync.Mutex
	data map[string]memEntry
	Fail error
}

type memEntry struct {
	data []byte
	rev  int64
}

// NewMemoryRef returns an empty MemoryRef.
func NewMemoryRef() *MemoryRef { return &MemoryRef{data: map[string]memEntry{}} }

func (m *MemoryRef) Read(_ context.Context, key string) ([]byte, int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.Fail != nil {
		return nil, 0, m.Fail
	}
	e, ok := m.data[key]
	if !ok {
		return nil, 0, core.ErrNotHeld
	}
	return append([]byte(nil), e.data...), e.rev, nil
}

func (m *MemoryRef) CompareAndSwap(_ context.Context, key string, expectRev int64, data []byte) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.Fail != nil {
		return 0, m.Fail
	}
	if m.data[key].rev != expectRev {
		return 0, fmt.Errorf("%w: %s is at revision %d, not %d", core.ErrRevisionConflict, key, m.data[key].rev, expectRev)
	}
	e := memEntry{data: append([]byte(nil), data...), rev: expectRev + 1}
	m.data[key] = e
	return e.rev, nil
}

func (m *MemoryRef) Delete(_ context.Context, key string, expectRev int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.Fail != nil {
		return m.Fail
	}
	if m.data[key].rev != expectRev {
		return fmt.Errorf("%w: %s is at revision %d, not %d", core.ErrRevisionConflict, key, m.data[key].rev, expectRev)
	}
	delete(m.data, key)
	return nil
}

func (m *MemoryRef) Keys(_ context.Context, prefix string) ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.Fail != nil {
		return nil, m.Fail
	}
	var out []string
	for k := range m.data {
		if strings.HasPrefix(k, prefix) {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out, nil
}

// Bump advances a key's revision as another writer would, for tests.
func (m *MemoryRef) Bump(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if e, ok := m.data[key]; ok {
		e.rev++
		m.data[key] = e
	}
}
