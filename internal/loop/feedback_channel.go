package loop

import "log/slog"

// defaultFeedbackChannelCapacity is the default bounded channel size.
// REQ-LL-009: overflow drops oldest event with a warn log.
const defaultFeedbackChannelCapacity = 64

// FeedbackChannel is a bounded channel that delivers Feedback events
// from PostTool hooks to the LoopController.
//
// @MX:WARN: [AUTO] Bounded channel — overflow drops oldest event silently (after warn log)
// @MX:REASON: REQ-LL-009 requires bounded delivery; callers that produce faster than consumers must not block; oldest events are least relevant
type FeedbackChannel struct {
	ch       chan Feedback
	capacity int
}

// NewFeedbackChannel creates a bounded FeedbackChannel with the given capacity.
// If capacity <= 0, defaultFeedbackChannelCapacity is used.
func NewFeedbackChannel(capacity int) *FeedbackChannel {
	if capacity <= 0 {
		capacity = defaultFeedbackChannelCapacity
	}
	return &FeedbackChannel{
		ch:       make(chan Feedback, capacity),
		capacity: capacity,
	}
}

// Send enqueues a Feedback event. If the channel is full, the oldest event is
// dropped and a warning is logged before inserting the new event (REQ-LL-009).
func (fc *FeedbackChannel) Send(fb Feedback) {
	select {
	case fc.ch <- fb:
		// Fast path: channel has capacity.
	default:
		// Channel full — drop oldest event.
		select {
		case dropped := <-fc.ch:
			slog.Warn("feedback channel full: dropping oldest event",
				"dropped_phase", dropped.Phase,
				"dropped_iteration", dropped.Iteration,
				"capacity", fc.capacity,
			)
		default:
			// Concurrent drain already freed space; ignore.
		}
		// Insert new event; this should succeed since we just freed one slot.
		select {
		case fc.ch <- fb:
		default:
			// Extremely unlikely (concurrent producers); log and discard.
			slog.Warn("feedback channel: could not insert after drain, discarding",
				"phase", fb.Phase,
				"iteration", fb.Iteration,
			)
		}
	}
}

// TryReceive attempts a non-blocking receive. Returns the Feedback and true
// if an event was available, or zero-value and false otherwise.
func (fc *FeedbackChannel) TryReceive() (Feedback, bool) {
	select {
	case fb := <-fc.ch:
		return fb, true
	default:
		return Feedback{}, false
	}
}

// Chan returns the underlying channel for use in select statements.
func (fc *FeedbackChannel) Chan() <-chan Feedback {
	return fc.ch
}

// Len returns the number of events currently buffered in the channel.
func (fc *FeedbackChannel) Len() int {
	return len(fc.ch)
}
