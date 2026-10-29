package forge

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
)

// ChangeKind says what happened to one keyed record.
type ChangeKind string

const (
	ChangeAdded    ChangeKind = "added"
	ChangeModified ChangeKind = "modified"
	ChangeRemoved  ChangeKind = "removed"
)

// Change is one state change; only changes are handed on, never a full snapshot.
type Change struct {
	Key  string
	Kind ChangeKind
}

// Differ turns two bodies into changes.
type Differ func(old, new []byte) ([]Change, error)

// DiffByKey compares JSON arrays of objects by one key field.
func DiffByKey(field string) Differ {
	index := func(b []byte) (map[string][]byte, error) {
		out := map[string][]byte{}
		if len(bytes.TrimSpace(b)) == 0 {
			return out, nil
		}
		var items []map[string]json.RawMessage
		if err := json.Unmarshal(b, &items); err != nil {
			return nil, fmt.Errorf("forge: diff input: %w", err)
		}
		for _, it := range items {
			k := string(bytes.Trim(it[field], `"`))
			raw, _ := json.Marshal(it) // map keys marshal sorted: canonical form
			out[k] = raw
		}
		return out, nil
	}
	return func(old, new []byte) ([]Change, error) {
		o, err := index(old)
		if err != nil {
			return nil, err
		}
		n, err := index(new)
		if err != nil {
			return nil, err
		}
		var out []Change
		for k, v := range n {
			if prev, ok := o[k]; !ok {
				out = append(out, Change{k, ChangeAdded})
			} else if !bytes.Equal(prev, v) {
				out = append(out, Change{k, ChangeModified})
			}
		}
		for k := range o {
			if _, ok := n[k]; !ok {
				out = append(out, Change{k, ChangeRemoved})
			}
		}
		sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
		return out, nil
	}
}

// WebhookFeed is the push channel. When it is healthy it is the primary source and polling shrinks to
// a safety net.
type WebhookFeed interface {
	Healthy() bool
	Drain() []Change
}

// PendingChange is one pushed event that is queued durably. ID is the host's delivery id.
type PendingChange struct {
	ID string
	Change
}

// DurableFeed is a push channel whose events survive a restart: Pending lists them without removing
// anything, and an event leaves the queue only through Commit, after the consumer acknowledged it.
type DurableFeed interface {
	Healthy() bool
	Pending() ([]PendingChange, error)
	Commit(ids []string) error
}

// Adaptive bounds the polling interval; all three come from configuration.
type Adaptive struct {
	Min, Max time.Duration
	Safety   time.Duration // poll interval while a webhook feed is healthy
}

// Report is what one Tick produced. Changes are delivered at least once: they are repeated on every
// Tick until the caller confirms delivery with Ack, so a crash between Tick and delivery loses nothing.
type Report struct {
	Changes  []Change
	Next     time.Time // earliest time the collector wants to be called again
	Parked   bool
	Reason   string
	Baseline bool // Changes came from a first observation (nothing acknowledged yet): initial state, not new events
	// Notice is a fact about the stored cursor worth showing, not a failure: for example that it was
	// resumed from an older generation. It stays on the reports until an acknowledgement is saved.
	Notice string

	snap    []byte   // observed body this report describes; Ack makes it the acknowledged baseline
	feedN   int      // how many pushed (webhook) events this report carries
	feedIDs []string // delivery ids of the durable pushed events this report carries
}

// CursorStore persists the consumer-acknowledged baseline (the collector contract's cursor), so a restart
// resumes from it instead of reporting everything again. The production implementation keeps it in the
// claim store's documents; a store unavailable at load time pauses the collector rather than inventing
// a baseline, because replacing a real cursor with the current snapshot would lose changes.
type CursorStore interface {
	LoadCursor(key string) (body []byte, found bool, err error)
	SaveCursor(key string, body []byte) error
}

// NotedCursorStore is a CursorStore that can also say when the cursor it returned is not the newest one.
// The collector puts that note in its reports until the next acknowledgement is saved.
type NotedCursorStore interface {
	CursorStore
	LoadCursorNoted(key string) (body []byte, found bool, note string, err error)
}

// Collector is the single collector of one repository: it fetches one path, remembers what it last
// saw and what the consumer last acknowledged, passes on only the difference, and picks its own next
// wake time.
type Collector struct {
	Client   *Client
	Path     string
	Interval Adaptive
	Differ   Differ
	// Feed is a push channel kept in memory: its events are lost on a restart (the safety poll covers
	// them). Durable is the push channel that survives a restart; use it where losing one matters.
	Feed    WebhookFeed
	Durable DurableFeed
	// Cursors and CursorKey persist the acknowledged baseline. Optional: without a store the baseline
	// lives in memory only and the first report after a restart is a Baseline.
	Cursors   CursorStore
	CursorKey string

	mu       sync.Mutex
	loaded   bool   // the cursor has been read from Cursors (or there is none to read)
	note     string // what loading the cursor had to say; cleared once an acknowledgement is saved
	last     []byte // acknowledged by the consumer
	seen     []byte // latest observed from the host
	feedBuf  []Change
	interval time.Duration
	next     time.Time
	parked   bool
}

// Ack confirms that the consumer received everything in rep. Until then the same changes come back.
//
// With a CursorStore the acknowledgement is saved first, and only a saved one advances the baseline: a
// failed save is returned and the same changes come back on the next Tick.
func (c *Collector) Ack(rep Report) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if rep.snap != nil && c.Cursors != nil {
		if err := c.Cursors.SaveCursor(c.CursorKey, rep.snap); err != nil {
			return fmt.Errorf("forge: the acknowledgement could not be saved: %w", err)
		}
	}
	if rep.snap != nil {
		c.last = rep.snap
		c.note = ""
	}
	if rep.feedN > 0 {
		n := min(rep.feedN, len(c.feedBuf))
		c.feedBuf = append([]Change(nil), c.feedBuf[n:]...)
	}
	// Durable events leave the queue last: a crash after the cursor save but before this commit reports
	// them once more (at least once), never zero times.
	if len(rep.feedIDs) > 0 && c.Durable != nil {
		if err := c.Durable.Commit(rep.feedIDs); err != nil {
			return fmt.Errorf("forge: the pushed events could not be removed from the queue: %w", err)
		}
	}
	return nil
}

// Tick runs the collector at `now`. Before its own next time it makes no call.
func (c *Collector) Tick(ctx context.Context, now time.Time) Report {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.loaded && c.Cursors != nil {
		var body []byte
		var found bool
		var err error
		if n, ok := c.Cursors.(NotedCursorStore); ok {
			body, found, c.note, err = n.LoadCursorNoted(c.CursorKey)
		} else {
			body, found, err = c.Cursors.LoadCursor(c.CursorKey)
		}
		if err != nil {
			// Do not poll and do not report: with the real cursor unknown, any report could be wrong.
			return Report{Reason: "the cursor could not be read: " + err.Error(), Next: now.Add(c.Interval.Min)}
		}
		if found {
			c.last = body
		}
	}
	c.loaded = true
	rep := c.tick(ctx, now)
	rep.Notice = c.note
	rep.Changes = append(rep.Changes, c.feedBuf...)
	rep.feedN = len(c.feedBuf)
	if c.Durable != nil {
		pending, err := c.Durable.Pending()
		if err != nil {
			if rep.Reason == "" {
				rep.Reason = "the pushed-event queue could not be read: " + err.Error()
			}
		}
		for _, p := range pending {
			rep.Changes = append(rep.Changes, p.Change)
			rep.feedIDs = append(rep.feedIDs, p.ID)
		}
	}
	if c.seen != nil {
		if ch, err := c.Differ(c.last, c.seen); err == nil {
			rep.Changes = append(rep.Changes, ch...)
		} else if rep.Reason == "" {
			rep.Reason = err.Error()
		}
		rep.snap, rep.Baseline = c.seen, c.last == nil
	}
	return rep
}

func (c *Collector) tick(ctx context.Context, now time.Time) Report {
	if c.parked {
		return Report{Parked: true, Reason: "parked: permission denied", Next: now.Add(c.Interval.Max)}
	}
	var rep Report
	if c.feedHealthy() {
		if c.Feed != nil && c.Feed.Healthy() {
			c.feedBuf = append(c.feedBuf, c.Feed.Drain()...)
		}
		if now.Before(c.next) {
			rep.Next = minTime(c.next, now.Add(c.Interval.Safety))
			return rep
		}
		if c.next.IsZero() { // first contact: rely on the feed until the safety poll
			c.next = now.Add(c.Interval.Safety)
			rep.Next = c.next
			return rep
		}
	} else if now.Before(c.next) {
		return Report{Next: c.next}
	}

	res, err := c.Client.Get(ctx, c.Path)
	switch {
	case errors.Is(err, ErrDenied) || errors.Is(err, ErrGiveUp):
		c.parked = true
		rep.Parked, rep.Reason = true, "permission denied; not retried"
		rep.Next = now.Add(c.Interval.Max)
		return rep
	case err != nil:
		rep.Reason, rep.Next = err.Error(), now.Add(c.Interval.Max)
		c.next = rep.Next
		return rep
	case res.Deferred:
		c.next = res.NotBefore
		rep.Next = c.next
		return rep
	}
	if c.interval == 0 {
		c.interval = c.Interval.Min
	}
	moved := false
	if res.Changed {
		fresh, derr := c.Differ(c.seen, res.Body)
		if derr != nil {
			rep.Reason = derr.Error()
		} else {
			c.seen = res.Body
			moved = len(fresh) > 0
		}
	}
	if moved {
		c.interval = max(c.Interval.Min, c.interval/2)
	} else {
		c.interval = min(c.Interval.Max, c.interval*2)
	}
	wait := c.interval
	if res.PollInterval > wait {
		wait = res.PollInterval
	}
	if c.feedHealthy() {
		wait = c.Interval.Safety
	}
	c.next = now.Add(wait)
	rep.Next = c.next
	return rep
}

// feedHealthy: a healthy push channel, volatile or durable, turns polling into a safety net.
func (c *Collector) feedHealthy() bool {
	return (c.Feed != nil && c.Feed.Healthy()) || (c.Durable != nil && c.Durable.Healthy())
}

func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

// Registry keeps exactly one Collector per repository.
type Registry struct {
	New func(repo string) *Collector

	mu sync.Mutex
	m  map[string]*Collector
}

func (r *Registry) For(repo string) *Collector {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.m == nil {
		r.m = map[string]*Collector{}
	}
	if c, ok := r.m[repo]; ok {
		return c
	}
	c := r.New(repo)
	r.m[repo] = c
	return c
}
