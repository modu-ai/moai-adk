package forge

import (
	"sort"
	"sync"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// Item is one queued write. Priority is lane 18's core.Priority: the ordering and its meaning are
// defined in core only, and a lower value is more urgent. Report items with the same Key are merged.
type Item struct {
	Priority core.Priority
	Key      string
	Payload  any
}

// Scheduler decides which queued write may go out next under the shared Budget. Writes more urgent than
// a report may spend the reserve; reports never do, and are held (and merged by key) until the budget
// recovers. Nothing is dropped. The order is the numeric order of core.Priority, so a tier added to
// core later is honoured without a change here.
type Scheduler struct {
	Gate     *Budget
	Resource string                 // budget bucket the writes count against; "" = default
	Merge    func(old, add any) any // combines two progress payloads with the same key

	mu    sync.Mutex
	queue []Item
}

// Push queues an item. A progress item whose key is already queued is merged into the queued one.
func (s *Scheduler) Push(it Item) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if it.Priority >= core.PriorityReport {
		for i := range s.queue {
			if s.queue[i].Priority == it.Priority && s.queue[i].Key == it.Key {
				if s.Merge != nil {
					s.queue[i].Payload = s.Merge(s.queue[i].Payload, it.Payload)
				} else {
					s.queue[i].Payload = it.Payload
				}
				return
			}
		}
	}
	s.queue = append(s.queue, it)
}

// Held is the number of queued items.
func (s *Scheduler) Held() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.queue)
}

// Pop returns the next item allowed to go out at `now`. When none may, it returns ok=false and, if the
// reason is a block or an exhausted window, when to look again.
func (s *Scheduler) Pop(now time.Time) (Item, bool, time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var until time.Time
	// find the most urgent queued priority that the budget lets through, remembering when to retry
	order := make([]core.Priority, 0, len(s.queue))
	seen := map[core.Priority]bool{}
	for _, it := range s.queue {
		if !seen[it.Priority] {
			seen[it.Priority] = true
			order = append(order, it.Priority)
		}
	}
	sort.Slice(order, func(i, j int) bool { return order[i] < order[j] })
	for _, pr := range order {
		var ok bool
		var when time.Time
		if pr >= core.PriorityReport {
			ok, when = s.Gate.AllowFor(s.Resource, now)
		} else {
			ok, when = s.Gate.AllowReserved(s.Resource, now)
		}
		if !ok {
			if until.IsZero() || when.Before(until) {
				until = when
			}
			continue
		}
		for i, it := range s.queue {
			if it.Priority == pr {
				s.queue = append(s.queue[:i], s.queue[i+1:]...)
				return it, true, time.Time{}
			}
		}
	}
	return Item{}, false, until
}
