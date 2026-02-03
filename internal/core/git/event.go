package git

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// defaultPollInterval is the default polling interval for event detection.
const defaultPollInterval = 5 * time.Second

// EventOption configures the EventDetector.
type EventOption func(*EventDetector)

// WithPollInterval sets the polling interval for event detection.
func WithPollInterval(d time.Duration) EventOption {
	return func(e *EventDetector) {
		e.pollInterval = d
	}
}

// EventDetector monitors a Git repository for state changes
// and emits GitEvent values when changes are detected.
type EventDetector struct {
	root         string
	logger       *slog.Logger
	pollInterval time.Duration

	mu         sync.Mutex
	lastBranch string
	lastHEAD   string
}

// NewEventDetector creates a new EventDetector for the repository at root.
func NewEventDetector(root string, opts ...EventOption) *EventDetector {
	e := &EventDetector{
		root:         root,
		logger:       slog.Default().With("module", "git.event"),
		pollInterval: defaultPollInterval,
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// Snapshot captures the current Git state (branch and HEAD hash)
// as the baseline for subsequent change detection.
func (e *EventDetector) Snapshot() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Detached HEAD is not an error for snapshots; branch will be empty.
	branch, branchErr := currentBranch(ctx, e.root)
	_ = branchErr
	head, err := execGit(ctx, e.root, "rev-parse", "HEAD")
	if err != nil {
		return fmt.Errorf("snapshot HEAD: %w", err)
	}

	e.lastBranch = branch
	e.lastHEAD = head
	e.logger.Debug("snapshot captured", "branch", branch, "head", head)
	return nil
}

// DetectChanges compares the current Git state against the last snapshot
// and returns any detected events. The internal state is updated to the
// current state after detection.
func (e *EventDetector) DetectChanges() ([]GitEvent, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Detached HEAD is not an error for detection; branch will be empty.
	nowBranch, nowBranchErr := currentBranch(ctx, e.root)
	_ = nowBranchErr
	nowHEAD, err := execGit(ctx, e.root, "rev-parse", "HEAD")
	if err != nil {
		return nil, fmt.Errorf("detect changes: %w", err)
	}

	var events []GitEvent
	now := time.Now()

	// Detect branch switch.
	if e.lastBranch != "" && nowBranch != "" && e.lastBranch != nowBranch {
		events = append(events, GitEvent{
			Type:           EventBranchSwitch,
			PreviousBranch: e.lastBranch,
			CurrentBranch:  nowBranch,
			PreviousHEAD:   e.lastHEAD,
			CurrentHEAD:    nowHEAD,
			Timestamp:      now,
		})
	}

	// Detect new commit (HEAD changed but branch did not, or same branch).
	if e.lastHEAD != "" && e.lastHEAD != nowHEAD && e.lastBranch == nowBranch {
		events = append(events, GitEvent{
			Type:           EventNewCommit,
			PreviousBranch: e.lastBranch,
			CurrentBranch:  nowBranch,
			PreviousHEAD:   e.lastHEAD,
			CurrentHEAD:    nowHEAD,
			Timestamp:      now,
		})
	}

	// Update state.
	e.lastBranch = nowBranch
	e.lastHEAD = nowHEAD

	if len(events) > 0 {
		e.logger.Debug("changes detected", "events", len(events))
	}

	return events, nil
}

// Poll continuously monitors Git state changes at the configured interval
// and sends detected events to the provided channel. It blocks until the
// context is cancelled, at which point it returns ctx.Err().
func (e *EventDetector) Poll(ctx context.Context, ch chan<- GitEvent) error {
	e.logger.Debug("starting event polling", "interval", e.pollInterval)

	// Capture initial state.
	if err := e.Snapshot(); err != nil {
		return fmt.Errorf("initial snapshot: %w", err)
	}

	ticker := time.NewTicker(e.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			e.logger.Debug("event polling stopped")
			return ctx.Err()
		case <-ticker.C:
			events, err := e.DetectChanges()
			if err != nil {
				e.logger.Error("detect changes failed", "error", err)
				continue
			}
			for _, ev := range events {
				select {
				case ch <- ev:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		}
	}
}
