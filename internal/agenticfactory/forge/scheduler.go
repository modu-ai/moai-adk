package forge

import (
	"sync"
	"time"
)

// Class orders writes when the budget is tight: state-safety writes first, then safety notices, then
// progress and report comments.
type Class int

const (
	ClassStateWrite   Class = iota // claim, PARKED, blocked, completion obligation, readback
	ClassSafetyNotice              // security failure, approval needed, an actual block
	ClassProgress                  // progress and report comments
)

// Item is one queued write. Items of the progress class with the same Key are merged into one.
type Item struct {
	Class   Class
	Key     string
	Payload any
}

// Scheduler decides which queued write may go out next under the shared Budget. State and safety
// writes may spend the reserve; progress comments never do, and are held (and merged by key) until the
// budget recovers. Nothing is dropped.
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
	if it.Class == ClassProgress {
		for i := range s.queue {
			if s.queue[i].Class == ClassProgress && s.queue[i].Key == it.Key {
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
	for _, cl := range []Class{ClassStateWrite, ClassSafetyNotice, ClassProgress} {
		var ok bool
		var when time.Time
		if cl == ClassProgress {
			ok, when = s.Gate.AllowFor(s.Resource, now)
		} else {
			ok, when = s.Gate.AllowReserved(s.Resource, now)
		}
		for i, it := range s.queue {
			if it.Class != cl {
				continue
			}
			if !ok {
				if until.IsZero() || when.Before(until) {
					until = when
				}
				break
			}
			s.queue = append(s.queue[:i], s.queue[i+1:]...)
			return it, true, time.Time{}
		}
	}
	return Item{}, false, until
}
