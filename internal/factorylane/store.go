package factorylane

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/atomicfile"
	"github.com/modu-ai/moai-adk/internal/config"
)

// DefaultStateRoot is the fallback-state location under the project root,
// following the sessionmsg state-root convention. The shape under it is
// runtime-managed and writer-scoped to this package.
const DefaultStateRoot = ".moai/state/factory-fallback"

// timerSnapshot and boundSnapshot are the configured no-response timer and
// unavailability bound period in Go duration-string form. Every observation
// carries its own snapshot at request time, so an evaluation stays stable
// across later config changes.
var (
	timerSnapshot = (time.Duration(config.DefaultFactoryNoResponseMinutes) * time.Minute).String()
	boundSnapshot = (time.Duration(config.DefaultFactoryFallbackBoundMinutes) * time.Minute).String()
)

// Clock is the time source seam; FakeClock drives tests deterministically
// (sessionmsg registry precedent).
type Clock interface {
	Now() time.Time
}

// FakeClock is a test clock the caller advances by hand.
type FakeClock struct {
	Current time.Time
}

// Now returns the fake clock's current instant.
func (f *FakeClock) Now() time.Time { return f.Current }

type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now().UTC() }

// Observation is one directed lead request and what became of it. Persisted
// one file per observation under observations/<lane>/, following the
// sessionmsg registry's one-record-per-file convention: mutations rewrite
// only their own record file, so concurrent lanes never lose each other's
// rows.
type Observation struct {
	Lane         string     `json:"lane"`
	RequestedAt  time.Time  `json:"requested_at"`
	Timer        string     `json:"timer"`
	Bound        string     `json:"bound"`
	AckedAt      *time.Time `json:"acked_at,omitempty"`
	NoResponseAt *time.Time `json:"no_response_at,omitempty"`
}

// Store is the fallback-state store: directed-request observations and the
// fallback-transition event log.
type Store struct {
	root  string
	clock Clock
}

// NewStore constructs a Store bound to <projectRoot>/.moai/state/factory-fallback.
// A nil clock selects the real UTC clock.
func NewStore(projectRoot string, clock Clock) *Store {
	if clock == nil {
		clock = systemClock{}
	}
	return &Store{root: filepath.Join(projectRoot, DefaultStateRoot), clock: clock}
}

// obsDir is the directory holding one lane's observation records.
func (s *Store) obsDir(lane string) string {
	return filepath.Join(s.root, "observations", lane)
}

// obsPath names one observation record; the filename carries the request
// instant (unix nano, zero-padded) so a lexical directory scan is already in
// request order.
func (s *Store) obsPath(lane string, at time.Time) string {
	return filepath.Join(s.obsDir(lane), fmt.Sprintf("obs-%019d.json", at.UnixNano()))
}

// writeFileAtomic persists one record file atomically: temp file in the
// record's own directory, then rename (sessionmsg registry precedent).
func writeFileAtomic(path string, data []byte) error {
	if len(data) == 0 || data[len(data)-1] != '\n' {
		data = append(data, '\n')
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("factorylane: mkdir %s: %w", dir, err)
	}
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+"-*.tmp")
	if err != nil {
		return fmt.Errorf("factorylane: create temp in %s: %w", dir, err)
	}
	tmpPath := tmp.Name()
	cleanup := func() { _ = os.Remove(tmpPath) }
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		cleanup()
		return fmt.Errorf("factorylane: write temp %s: %w", tmpPath, err)
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return fmt.Errorf("factorylane: close temp %s: %w", tmpPath, err)
	}
	if err := atomicfile.Replace(tmpPath, path); err != nil {
		cleanup()
		return fmt.Errorf("factorylane: rename temp -> %s: %w", path, err)
	}
	return nil
}

// marshalRecord renders one record file's bytes (indented JSON, newline
// terminated) so the state stays human-readable like the sessionmsg registry.
func marshalRecord(v any) ([]byte, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("factorylane: marshal record: %w", err)
	}
	return data, nil
}

// writeObs persists one observation record atomically.
func (s *Store) writeObs(obs Observation) error {
	data, err := marshalRecord(obs)
	if err != nil {
		return err
	}
	return writeFileAtomic(s.obsPath(obs.Lane, obs.RequestedAt), data)
}

// readObs loads every record file in dir, in request order. A missing
// directory is an empty set.
func (s *Store) readObs(lane string) ([]Observation, error) {
	dir := s.obsDir(lane)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("factorylane: read observations dir %s: %w", dir, err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasPrefix(e.Name(), "obs-") || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)
	obs := make([]Observation, 0, len(names))
	for _, name := range names {
		data, err := atomicfile.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, fmt.Errorf("factorylane: read observation %s: %w", name, err)
		}
		var o Observation
		if err := json.Unmarshal(data, &o); err != nil {
			return nil, fmt.Errorf("factorylane: parse observation %s: %w", name, err)
		}
		obs = append(obs, o)
	}
	return obs, nil
}

// RecordRequest records one directed lead request: the observation row
// carries the request instant and the timer/bound snapshots in force at
// request time (REQ-FLA-002).
func (s *Store) RecordRequest(lane string) (Observation, error) {
	now := s.clock.Now().UTC()
	obs := Observation{
		Lane:        lane,
		RequestedAt: now,
		Timer:       timerSnapshot,
		Bound:       boundSnapshot,
	}
	if err := s.writeObs(obs); err != nil {
		return Observation{}, err
	}
	return obs, nil
}

// Observations returns the lane's observations in request order.
func (s *Store) Observations(lane string) ([]Observation, error) {
	return s.readObs(lane)
}

// AckPending marks the lane's latest un-acked, unexpired observation acked.
// It reports false when nothing is pending. An acked observation is never
// re-acked and never recorded as no-response.
func (s *Store) AckPending(lane string) (bool, error) {
	obs, err := s.readObs(lane)
	if err != nil {
		return false, err
	}
	now := s.clock.Now().UTC()
	for i := len(obs) - 1; i >= 0; i-- {
		o := obs[i]
		if o.AckedAt != nil || o.NoResponseAt != nil {
			continue
		}
		o.AckedAt = &now
		if err := s.writeObs(o); err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

// SweepNoResponse materializes the no-response fact: every un-acked
// observation whose timer expired at or before now gains its NoResponseAt
// timestamp. Pending and acked observations are left untouched. The returned
// slice is the lane's full observation set after the sweep.
func (s *Store) SweepNoResponse(lane string) ([]Observation, error) {
	obs, err := s.readObs(lane)
	if err != nil {
		return nil, err
	}
	now := s.clock.Now().UTC()
	for _, o := range obs {
		if o.AckedAt != nil || o.NoResponseAt != nil {
			continue
		}
		timer, err := time.ParseDuration(o.Timer)
		if err != nil {
			return nil, fmt.Errorf("factorylane: parse timer %q: %w", o.Timer, err)
		}
		if now.Before(o.RequestedAt.Add(timer)) {
			continue
		}
		expired := now
		o.NoResponseAt = &expired
		if err := s.writeObs(o); err != nil {
			return nil, err
		}
	}
	return s.readObs(lane)
}
