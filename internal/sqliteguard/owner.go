package sqliteguard

import (
	"crypto/sha256"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

var (
	ErrFaulted  = errors.New("SQLite privacy: path has an uncertain retained lifetime")
	ErrCapacity = errors.New("SQLite privacy: physical lifetime capacity exhausted")
	ErrRetained = errors.New("SQLite privacy: resources retained until process exit")
)

// MaxPhysicalLifetimes bounds both active and quarantined reservations before
// acquiring native resources. It is a new admission limit, not a timeout or a
// promise of unlimited pools. Each reservation owns at most three file pins.
const MaxPhysicalLifetimes = 128

// Snapshot reports ownership without exposing paths, DSNs, SQL or error text.
type Snapshot struct {
	Token       uint64
	PathDigest  string
	Stage       string
	Pins        int
	Outstanding int
	Faulted     bool
}

type manager struct {
	mu     sync.Mutex
	limit  int
	next   uint64
	live   map[uint64]*lifetime
	report func(Snapshot)
}

type lifetime struct {
	m             *manager
	token         uint64
	path          string
	stage         string
	pins          []io.Closer
	conn          driver.Conn
	faulted       bool
	closing       bool
	sticky        bool
	resources     int
	childSequence uint64
	children      map[uint64]io.Closer
	cause         error
}

var processOwners = newManager(MaxPhysicalLifetimes, func(s Snapshot) {
	slog.Error("SQLite privacy lifetime retained", "token", s.Token, "path_digest", s.PathDigest,
		"stage", s.Stage, "pins", s.Pins, "outstanding", s.Outstanding)
})

func newManager(limit int, report func(Snapshot)) *manager {
	return &manager{limit: limit, live: make(map[uint64]*lifetime), report: report}
}

// Status includes live and faulted process-owned lifetimes. There is no TTL,
// background retry, force-release or caller-controlled reset for uncertainty.
func Status() []Snapshot { return processOwners.snapshots() }

func (m *manager) reserve(path string) (*lifetime, error) {
	path = filepath.Clean(path)
	if runtime.GOOS == "windows" {
		path = strings.ToLower(path)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, o := range m.live {
		if o.path == path && o.faulted {
			return nil, errors.Join(ErrFaulted, o.cause)
		}
	}
	if len(m.live) >= m.limit {
		return nil, ErrCapacity
	}
	m.next++
	o := &lifetime{m: m, token: m.next, path: path, stage: "preparing", pins: make([]io.Closer, 0, 3)}
	m.live[o.token] = o
	return o, nil
}

func (o *lifetime) snapshot() Snapshot {
	sum := sha256.Sum256([]byte(o.path))
	return Snapshot{Token: o.token, PathDigest: fmt.Sprintf("%x", sum), Stage: o.stage, Pins: len(o.pins), Outstanding: o.resources, Faulted: o.faulted}
}

func (m *manager) snapshots() []Snapshot {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Snapshot, 0, len(m.live))
	for _, o := range m.live {
		out = append(out, o.snapshot())
	}
	return out
}

func (o *lifetime) retain(stage string, cause error) error {
	o.m.mu.Lock()
	o.faulted = true
	o.stage = stage
	o.cause = errors.Join(o.cause, cause, ErrRetained)
	s := o.snapshot()
	result := o.cause
	o.m.mu.Unlock()
	o.m.report(s)
	return result
}

// release is called only before entry into the driver, or after confirmed
// physical Close with no outstanding raw child. Cleanup errors remain owned.
func (o *lifetime) release() error {
	var err error
	for i := len(o.pins) - 1; i >= 0; i-- {
		err = errors.Join(err, o.pins[i].Close())
	}
	if err != nil {
		return o.retain("pin-close-failed", err)
	}
	o.m.mu.Lock()
	delete(o.m.live, o.token)
	o.stage = "released"
	o.m.mu.Unlock()
	return nil
}

func (o *lifetime) addResource() error {
	o.m.mu.Lock()
	defer o.m.mu.Unlock()
	if o.closing || o.faulted {
		return ErrFaulted
	}
	o.resources++
	return nil
}
func (o *lifetime) resourceDone() { o.m.mu.Lock(); o.resources--; o.m.mu.Unlock() }

func (o *lifetime) holdChild(child io.Closer) uint64 {
	o.m.mu.Lock()
	defer o.m.mu.Unlock()
	if o.children == nil {
		o.children = make(map[uint64]io.Closer)
	}
	o.childSequence++
	o.children[o.childSequence] = child
	return o.childSequence
}
func (o *lifetime) dropChild(token uint64) { o.m.mu.Lock(); delete(o.children, token); o.m.mu.Unlock() }

func (o *lifetime) retainPanic(stage string) {
	if p := recover(); p != nil {
		o.retain(stage, errors.New("SQLite operation panicked"))
		panic(p)
	}
}

func (o *lifetime) retainClosePanic(stage string, result *error) {
	if p := recover(); p != nil {
		*result = o.retain(stage, errors.New("SQLite resource Close panicked"))
		panic(p)
	}
}

func (o *lifetime) available() error {
	o.m.mu.Lock()
	defer o.m.mu.Unlock()
	if o.closing || o.faulted {
		return errors.Join(ErrFaulted, o.cause)
	}
	return nil
}

func (o *lifetime) beginClose() bool {
	o.m.mu.Lock()
	defer o.m.mu.Unlock()
	o.closing = true
	return o.resources == 0 && !o.sticky && !o.faulted
}
