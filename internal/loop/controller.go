package loop

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// LoopController implements the Controller interface, orchestrating
// the feedback loop lifecycle with goroutine-based execution.
type LoopController struct {
	mu        sync.Mutex
	state     *LoopState
	storage   Storage
	engine    DecisionEngine
	feedback  FeedbackGenerator
	maxIter   int
	running   bool
	converged bool
	paused    bool
	cancel    context.CancelFunc
	done      chan struct{}
}

// NewLoopController creates a new LoopController with injected dependencies.
func NewLoopController(storage Storage, engine DecisionEngine, feedback FeedbackGenerator, maxIterations int) *LoopController {
	if maxIterations <= 0 {
		maxIterations = 5
	}
	return &LoopController{
		storage:  storage,
		engine:   engine,
		feedback: feedback,
		maxIter:  maxIterations,
	}
}

// Start initializes a new feedback loop for the given SPEC ID and begins
// execution in a background goroutine. Returns ErrLoopAlreadyRunning if
// a loop is already active.
func (c *LoopController) Start(ctx context.Context, specID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.running {
		return ErrLoopAlreadyRunning
	}

	now := time.Now()
	c.state = &LoopState{
		SpecID:    specID,
		Phase:     PhaseAnalyze,
		Iteration: 1,
		MaxIter:   c.maxIter,
		Feedback:  []Feedback{},
		StartedAt: now,
		UpdatedAt: now,
	}

	if err := c.storage.SaveState(c.state); err != nil {
		return fmt.Errorf("loop: start: save initial state: %w", err)
	}

	c.running = true
	c.converged = false
	c.paused = false
	c.done = make(chan struct{})

	loopCtx, cancel := context.WithCancel(ctx)
	c.cancel = cancel

	go c.runLoop(loopCtx)

	return nil
}

// Pause stops the running loop, persists current state, and marks
// it as paused for later resumption. Returns ErrLoopNotRunning if
// no loop is active.
func (c *LoopController) Pause() error {
	c.mu.Lock()

	if !c.running {
		c.mu.Unlock()
		return ErrLoopNotRunning
	}

	// Cancel the loop goroutine context.
	if c.cancel != nil {
		c.cancel()
	}
	c.mu.Unlock()

	// Wait for goroutine to finish outside the lock.
	if c.done != nil {
		<-c.done
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.paused = true
	c.running = false

	if c.state != nil {
		c.state.UpdatedAt = time.Now()
		if err := c.storage.SaveState(c.state); err != nil {
			return fmt.Errorf("loop: pause: save state: %w", err)
		}
	}

	return nil
}

// Resume restores a paused loop from persisted state and continues
// execution. Returns ErrLoopNotPaused if the loop is not in a paused state.
func (c *LoopController) Resume(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.running {
		return ErrLoopNotPaused
	}
	if !c.paused && c.state == nil {
		return ErrLoopNotPaused
	}

	// If state was lost (e.g., session restart), try loading from storage.
	if c.state == nil {
		return ErrLoopNotPaused
	}

	// For resumption after request_review, advance to next iteration.
	if c.state.Phase == PhaseReview {
		c.state.Iteration++
		c.state.Phase = PhaseAnalyze
		c.state.UpdatedAt = time.Now()
	}

	c.running = true
	c.paused = false
	c.done = make(chan struct{})

	loopCtx, cancel := context.WithCancel(ctx)
	c.cancel = cancel

	go c.runLoop(loopCtx)

	return nil
}

// ResumeFromStorage restores loop state from persisted storage and
// resumes execution. This is used after session restarts.
func (c *LoopController) ResumeFromStorage(ctx context.Context, specID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.running {
		return ErrLoopNotPaused
	}

	loaded, err := c.storage.LoadState(specID)
	if err != nil {
		return fmt.Errorf("loop: resume: %w", err)
	}

	c.state = loaded
	c.paused = false
	c.running = true
	c.converged = false
	c.done = make(chan struct{})

	// For resumption after request_review, advance to next iteration.
	if c.state.Phase == PhaseReview {
		c.state.Iteration++
		c.state.Phase = PhaseAnalyze
		c.state.UpdatedAt = time.Now()
	}

	loopCtx, cancel := context.WithCancel(ctx)
	c.cancel = cancel

	go c.runLoop(loopCtx)

	return nil
}

// Cancel stops the loop and removes persisted state. Works on both
// running and paused loops. Returns ErrLoopNotRunning if no loop exists.
func (c *LoopController) Cancel() error {
	c.mu.Lock()

	if !c.running && !c.paused && c.state == nil {
		c.mu.Unlock()
		return ErrLoopNotRunning
	}

	specID := ""
	if c.state != nil {
		specID = c.state.SpecID
	}

	if c.cancel != nil {
		c.cancel()
	}

	wasRunning := c.running
	c.mu.Unlock()

	// Wait for goroutine to finish if it was running.
	if wasRunning && c.done != nil {
		<-c.done
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.running = false
	c.paused = false
	c.state = nil
	c.converged = false

	if specID != "" {
		if err := c.storage.DeleteState(specID); err != nil {
			return fmt.Errorf("loop: cancel: delete state: %w", err)
		}
	}

	return nil
}

// Status returns a read-only snapshot of the current loop state.
func (c *LoopController) Status() *LoopStatus {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state == nil {
		return &LoopStatus{}
	}
	return c.state.ToStatus(c.running, c.converged)
}

// RecordFeedback adds external feedback to the current loop state and
// persists it. Returns ErrLoopNotRunning if no loop is active.
func (c *LoopController) RecordFeedback(fb Feedback) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state == nil {
		return ErrLoopNotRunning
	}

	c.state.Feedback = append(c.state.Feedback, fb)
	c.state.UpdatedAt = time.Now()

	return c.storage.SaveState(c.state)
}

// Done returns a channel that is closed when the loop goroutine finishes.
// Returns nil if no loop has been started.
func (c *LoopController) Done() <-chan struct{} {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.done
}

// runLoop executes the feedback loop phases in sequence.
// It runs in a dedicated goroutine and responds to context cancellation.
func (c *LoopController) runLoop(ctx context.Context) {
	defer close(c.done)

	for {
		// Check context cancellation.
		select {
		case <-ctx.Done():
			c.mu.Lock()
			c.running = false
			c.mu.Unlock()
			return
		default:
		}

		// Collect feedback for the current phase.
		fb, err := c.feedback.Collect(ctx)
		if err != nil {
			if ctx.Err() != nil {
				c.mu.Lock()
				c.running = false
				c.mu.Unlock()
				return
			}
			fb = &Feedback{}
		}

		// Record feedback under lock.
		c.mu.Lock()
		fb.Phase = c.state.Phase
		fb.Iteration = c.state.Iteration
		c.state.Feedback = append(c.state.Feedback, *fb)
		c.state.UpdatedAt = time.Now()
		if err := c.storage.SaveState(c.state); err != nil {
			slog.Default().Warn("failed to save loop state after feedback collection",
				"spec_id", c.state.SpecID,
				"phase", c.state.Phase,
				"iteration", c.state.Iteration,
				"error", err)
		}

		currentPhase := c.state.Phase
		c.mu.Unlock()

		// At the review phase, invoke the decision engine.
		if currentPhase == PhaseReview {
			// Copy state for thread-safe engine call.
			c.mu.Lock()
			stateCopy := *c.state
			stateCopy.Feedback = make([]Feedback, len(c.state.Feedback))
			copy(stateCopy.Feedback, c.state.Feedback)
			c.mu.Unlock()

			decision, decErr := c.engine.Decide(ctx, &stateCopy, fb)
			if decErr != nil {
				decision = &Decision{Action: ActionContinue, NextPhase: PhaseAnalyze}
			}

			c.mu.Lock()
			switch decision.Action {
			case ActionConverge:
				c.converged = true
				c.running = false
				if err := c.storage.DeleteState(c.state.SpecID); err != nil {
					slog.Default().Warn("failed to delete loop state on convergence",
						"spec_id", c.state.SpecID,
						"error", err)
				}
				c.mu.Unlock()
				return

			case ActionAbort:
				c.running = false
				if err := c.storage.DeleteState(c.state.SpecID); err != nil {
					slog.Default().Warn("failed to delete loop state on abort",
						"spec_id", c.state.SpecID,
						"error", err)
				}
				c.mu.Unlock()
				return

			case ActionRequestReview:
				c.running = false
				c.paused = true
				if err := c.storage.SaveState(c.state); err != nil {
					slog.Default().Warn("failed to save loop state on review request",
						"spec_id", c.state.SpecID,
						"error", err)
				}
				c.mu.Unlock()
				return

			case ActionContinue:
				c.state.Iteration++
				c.state.Phase = PhaseAnalyze
				c.state.UpdatedAt = time.Now()
				if err := c.storage.SaveState(c.state); err != nil {
					slog.Default().Warn("failed to save loop state on continue",
						"spec_id", c.state.SpecID,
						"phase", c.state.Phase,
						"iteration", c.state.Iteration,
						"error", err)
				}
				c.mu.Unlock()
				continue

			default:
				c.mu.Unlock()
			}
		}

		// Advance to the next phase (non-review phases).
		c.mu.Lock()
		c.state.Phase = NextPhase(c.state.Phase)
		c.state.UpdatedAt = time.Now()
		c.mu.Unlock()
	}
}

// Compile-time interface compliance check.
var _ Controller = (*LoopController)(nil)
