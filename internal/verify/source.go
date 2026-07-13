package verify

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// DefaultLookupTimeout is the time-box for key computation + snapshot load on
// latency-sensitive lookup paths (the stop-goal Stop hook runs each turn-end;
// per the Advisory-Check Discipline the optimization is skipped — never the
// correctness — when the deadline is exceeded).
const DefaultLookupTimeout = 2 * time.Second

// Source resolves fresh recorded check results for exact byte-string command
// matches. It structurally satisfies the goal evaluator's snapshot-lookup
// interface without importing the goal package.
//
// The key computation and snapshot load run at most ONCE per Source instance
// (memoized), so per-turn lookup cost is constant regardless of how many
// conditions are evaluated. Construct one Source per evaluation cycle.
type Source struct {
	// ProjectRoot is the repository root whose tree state keys the snapshot.
	ProjectRoot string
	// TTL overrides the freshness bound; 0 selects DefaultTTL.
	TTL time.Duration
	// Timeout bounds key computation + load; 0 selects DefaultLookupTimeout.
	Timeout time.Duration
	// Now is the clock (tests inject); nil selects time.Now.
	Now func() time.Time
	// KeyFunc computes the working-tree key (tests inject); nil selects Key.
	KeyFunc func(ctx context.Context, repoDir string) (string, error)

	once sync.Once
	snap *Snapshot // nil when absent, key-computation failed, or timed out
	key  string
}

// Lookup returns the recorded exit code and a citable attribution string when
// a fresh snapshot entry exactly matches cmd (byte-string equality — no
// normalization; a near-miss variant never matches). ok=false on miss, stale,
// error, or deadline exceed — the caller re-executes the command unchanged.
func (s *Source) Lookup(ctx context.Context, cmd string) (exit int, attribution string, ok bool) {
	s.once.Do(func() { s.load(ctx) })
	if s.snap == nil {
		return 0, "", false
	}
	entry := s.snap.FindCommand(cmd)
	if entry == nil {
		return 0, "", false
	}
	now := time.Now()
	if s.Now != nil {
		now = s.Now()
	}
	if !Fresh(s.snap.Key, s.key, entry.RecordedAt, now, s.TTL) {
		return 0, "", false
	}
	attribution = fmt.Sprintf("snapshot %s key %s cmd %q exit %d recorded_at %s",
		SnapshotPath(s.ProjectRoot, s.key), s.key, entry.Command, entry.ExitCode,
		entry.RecordedAt.Format(time.RFC3339))
	return entry.ExitCode, attribution, true
}

// load performs the one-time, time-boxed key computation + snapshot load.
// Every failure mode (deadline, non-git dir, unreadable state dir) degrades to
// s.snap == nil — a miss, never an error surfaced to the evaluation path.
func (s *Source) load(ctx context.Context) {
	timeout := s.Timeout
	if timeout <= 0 {
		timeout = DefaultLookupTimeout
	}
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	keyFn := s.KeyFunc
	if keyFn == nil {
		keyFn = Key
	}
	key, err := keyFn(cctx, s.ProjectRoot)
	if err != nil {
		return
	}
	snap, err := Load(s.ProjectRoot, key)
	if err != nil || snap == nil {
		return
	}
	s.key = key
	s.snap = snap
}
