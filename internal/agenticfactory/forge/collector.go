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

// Adaptive bounds the polling interval; all three come from configuration.
type Adaptive struct {
	Min, Max time.Duration
	Safety   time.Duration // poll interval while a webhook feed is healthy
}

// Report is what one Tick produced.
type Report struct {
	Changes []Change
	Next    time.Time // earliest time the collector wants to be called again
	Parked  bool
	Reason  string
}

// Collector is the single collector of one repository: it fetches one path, keeps the last body,
// passes on only changes, and picks its own next wake time.
type Collector struct {
	Client   *Client
	Path     string
	Interval Adaptive
	Differ   Differ
	Feed     WebhookFeed

	mu       sync.Mutex
	last     []byte
	interval time.Duration
	next     time.Time
	parked   bool
}

// Tick runs the collector at `now`. Before its own next time it does nothing (no call).
func (c *Collector) Tick(ctx context.Context, now time.Time) Report {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.parked {
		return Report{Parked: true, Reason: "parked: permission denied", Next: now.Add(c.Interval.Max)}
	}
	var rep Report
	if c.Feed != nil && c.Feed.Healthy() {
		rep.Changes = append(rep.Changes, c.Feed.Drain()...)
		if now.Before(c.next) {
			rep.Next = minTime(c.next, now.Add(c.Interval.Safety))
			return rep
		}
		defer func() { c.next = now.Add(c.Interval.Safety) }()
		if c.next.IsZero() { // first contact: nothing to compare yet, rely on the feed until the safety poll
			c.next = now.Add(c.Interval.Safety)
			rep.Next = c.next
			return rep
		}
	} else if now.Before(c.next) {
		return Report{Next: c.next}
	}

	res, err := c.Client.Get(ctx, c.Path)
	switch {
	case errors.Is(err, ErrDenied):
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
	if res.Changed {
		ch, derr := c.Differ(c.last, res.Body)
		if derr != nil {
			rep.Reason = derr.Error()
		} else {
			rep.Changes = append(rep.Changes, ch...)
			c.last = res.Body
		}
		if len(ch) > 0 {
			c.interval = max(c.Interval.Min, c.interval/2)
		} else {
			c.interval = min(c.Interval.Max, c.interval*2)
		}
	} else {
		c.interval = min(c.Interval.Max, c.interval*2)
	}
	wait := c.interval
	if res.PollInterval > wait {
		wait = res.PollInterval
	}
	if c.Feed != nil && c.Feed.Healthy() {
		wait = c.Interval.Safety
	}
	c.next = now.Add(wait)
	rep.Next = c.next
	return rep
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
