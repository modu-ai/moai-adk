package loop

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

// --- Mock implementations for testing ---

// mockStorage implements Storage using in-memory maps.
type mockStorage struct {
	states    map[string]*LoopState
	saveCount atomic.Int32
}

func newMockStorage() *mockStorage {
	return &mockStorage{states: make(map[string]*LoopState)}
}

func (m *mockStorage) SaveState(state *LoopState) error {
	m.saveCount.Add(1)
	// Deep copy to avoid shared references.
	cp := *state
	cp.Feedback = make([]Feedback, len(state.Feedback))
	copy(cp.Feedback, state.Feedback)
	m.states[state.SpecID] = &cp
	return nil
}

func (m *mockStorage) LoadState(specID string) (*LoopState, error) {
	s, ok := m.states[specID]
	if !ok {
		return nil, ErrCorruptedState
	}
	cp := *s
	cp.Feedback = make([]Feedback, len(s.Feedback))
	copy(cp.Feedback, s.Feedback)
	return &cp, nil
}

func (m *mockStorage) DeleteState(specID string) error {
	delete(m.states, specID)
	return nil
}

// mockFeedbackGenerator returns fixed feedback immediately.
type mockFeedbackGenerator struct {
	fb *Feedback
}

func (m *mockFeedbackGenerator) Collect(_ context.Context) (*Feedback, error) {
	if m.fb == nil {
		return &Feedback{}, nil
	}
	cp := *m.fb
	return &cp, nil
}

// blockingFeedbackGenerator blocks until context is cancelled.
type blockingFeedbackGenerator struct{}

func (m *blockingFeedbackGenerator) Collect(ctx context.Context) (*Feedback, error) {
	<-ctx.Done()
	return nil, ctx.Err()
}

// mockDecisionEngine returns a fixed decision or calls a function.
type mockDecisionEngine struct {
	decideFn func(ctx context.Context, state *LoopState, fb *Feedback) (*Decision, error)
}

func (m *mockDecisionEngine) Decide(ctx context.Context, state *LoopState, fb *Feedback) (*Decision, error) {
	return m.decideFn(ctx, state, fb)
}

// Helper to wait for loop done with timeout.
func waitDone(t *testing.T, ctrl *LoopController, timeout time.Duration) {
	t.Helper()
	done := ctrl.Done()
	if done == nil {
		t.Fatal("Done() returned nil")
	}
	select {
	case <-done:
	case <-time.After(timeout):
		t.Fatal("timeout waiting for loop to finish")
	}
}

// --- Tests ---

func TestLoopController_Start(t *testing.T) {
	t.Parallel()

	storage := newMockStorage()
	engine := &mockDecisionEngine{
		decideFn: func(_ context.Context, _ *LoopState, _ *Feedback) (*Decision, error) {
			return &Decision{Action: ActionConverge, Converged: true, Reason: "test"}, nil
		},
	}
	fbGen := &mockFeedbackGenerator{fb: &Feedback{
		TestsPassed: 10, TestsFailed: 0, LintErrors: 0,
		BuildSuccess: true, Coverage: 90.0,
	}}

	ctrl := NewLoopController(storage, engine, fbGen, 5)
	err := ctrl.Start(context.Background(), "SPEC-TEST-001")
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Loop should converge quickly with perfect feedback.
	waitDone(t, ctrl, 5*time.Second)

	status := ctrl.Status()
	if status.SpecID != "SPEC-TEST-001" {
		t.Errorf("SpecID = %q, want %q", status.SpecID, "SPEC-TEST-001")
	}
	if status.Running {
		t.Error("Running = true, want false after convergence")
	}
	if !status.Converged {
		t.Error("Converged = false, want true")
	}
}

func TestLoopController_StartAlreadyRunning(t *testing.T) {
	t.Parallel()

	storage := newMockStorage()
	// Engine that blocks (never decides) - use blocking feedback instead.
	engine := &mockDecisionEngine{
		decideFn: func(_ context.Context, _ *LoopState, _ *Feedback) (*Decision, error) {
			return &Decision{Action: ActionContinue}, nil
		},
	}
	fbGen := &blockingFeedbackGenerator{}

	ctrl := NewLoopController(storage, engine, fbGen, 5)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := ctrl.Start(ctx, "SPEC-TEST-001"); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	err := ctrl.Start(ctx, "SPEC-TEST-001")
	if err != ErrLoopAlreadyRunning {
		t.Errorf("expected ErrLoopAlreadyRunning, got: %v", err)
	}

	cancel()
	waitDone(t, ctrl, 5*time.Second)
}

func TestLoopController_StartInitialState(t *testing.T) {
	t.Parallel()

	storage := newMockStorage()
	engine := &mockDecisionEngine{
		decideFn: func(_ context.Context, _ *LoopState, _ *Feedback) (*Decision, error) {
			return &Decision{Action: ActionConverge, Converged: true}, nil
		},
	}
	fbGen := &mockFeedbackGenerator{fb: &Feedback{
		TestsPassed: 10, BuildSuccess: true, Coverage: 90.0,
	}}

	ctrl := NewLoopController(storage, engine, fbGen, 5)
	if err := ctrl.Start(context.Background(), "SPEC-TEST-001"); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Check initial status before loop finishes.
	status := ctrl.Status()
	if status.MaxIter != 5 {
		t.Errorf("MaxIter = %d, want 5", status.MaxIter)
	}

	waitDone(t, ctrl, 5*time.Second)
}

func TestLoopController_Pause(t *testing.T) {
	t.Parallel()

	storage := newMockStorage()
	engine := &mockDecisionEngine{
		decideFn: func(_ context.Context, _ *LoopState, _ *Feedback) (*Decision, error) {
			return &Decision{Action: ActionContinue}, nil
		},
	}
	fbGen := &blockingFeedbackGenerator{}

	ctrl := NewLoopController(storage, engine, fbGen, 5)
	ctx := context.Background()

	if err := ctrl.Start(ctx, "SPEC-TEST-001"); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Give the goroutine time to enter the blocking Collect.
	time.Sleep(50 * time.Millisecond)

	if err := ctrl.Pause(); err != nil {
		t.Fatalf("Pause failed: %v", err)
	}

	status := ctrl.Status()
	if status.Running {
		t.Error("Running = true, want false after pause")
	}

	// State should be saved.
	saved, ok := storage.states["SPEC-TEST-001"]
	if !ok {
		t.Error("state not persisted after pause")
	}
	if saved == nil {
		t.Error("persisted state is nil")
	}
}

func TestLoopController_PauseNotRunning(t *testing.T) {
	t.Parallel()

	storage := newMockStorage()
	engine := &mockDecisionEngine{
		decideFn: func(_ context.Context, _ *LoopState, _ *Feedback) (*Decision, error) {
			return &Decision{Action: ActionContinue}, nil
		},
	}
	fbGen := &mockFeedbackGenerator{}

	ctrl := NewLoopController(storage, engine, fbGen, 5)

	err := ctrl.Pause()
	if err != ErrLoopNotRunning {
		t.Errorf("expected ErrLoopNotRunning, got: %v", err)
	}
}

func TestLoopController_Resume(t *testing.T) {
	t.Parallel()

	storage := newMockStorage()

	callCount := atomic.Int32{}
	engine := &mockDecisionEngine{
		decideFn: func(_ context.Context, _ *LoopState, _ *Feedback) (*Decision, error) {
			if callCount.Add(1) >= 2 {
				return &Decision{Action: ActionConverge, Converged: true, Reason: "done"}, nil
			}
			return &Decision{Action: ActionRequestReview, Reason: "review"}, nil
		},
	}
	fbGen := &mockFeedbackGenerator{fb: &Feedback{
		TestsPassed: 5, TestsFailed: 1, BuildSuccess: true, Coverage: 80.0,
	}}

	ctrl := NewLoopController(storage, engine, fbGen, 5)

	// Start -> should pause at review (request_review).
	if err := ctrl.Start(context.Background(), "SPEC-TEST-001"); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	waitDone(t, ctrl, 5*time.Second)

	status := ctrl.Status()
	if status.Running {
		t.Error("Running = true, want false after request_review")
	}

	// Resume -> should converge on second decision.
	if err := ctrl.Resume(context.Background()); err != nil {
		t.Fatalf("Resume failed: %v", err)
	}
	waitDone(t, ctrl, 5*time.Second)

	status = ctrl.Status()
	if status.Running {
		t.Error("Running = true, want false after convergence")
	}
	if !status.Converged {
		t.Error("Converged = false, want true")
	}
	if status.Iteration < 2 {
		t.Errorf("Iteration = %d, want >= 2 after resume", status.Iteration)
	}
}

func TestLoopController_ResumeNotPaused(t *testing.T) {
	t.Parallel()

	storage := newMockStorage()
	engine := &mockDecisionEngine{
		decideFn: func(_ context.Context, _ *LoopState, _ *Feedback) (*Decision, error) {
			return &Decision{Action: ActionContinue}, nil
		},
	}
	fbGen := &mockFeedbackGenerator{}

	ctrl := NewLoopController(storage, engine, fbGen, 5)

	err := ctrl.Resume(context.Background())
	if err != ErrLoopNotPaused {
		t.Errorf("expected ErrLoopNotPaused, got: %v", err)
	}
}

func TestLoopController_Cancel(t *testing.T) {
	t.Parallel()

	storage := newMockStorage()
	engine := &mockDecisionEngine{
		decideFn: func(_ context.Context, _ *LoopState, _ *Feedback) (*Decision, error) {
			return &Decision{Action: ActionContinue}, nil
		},
	}
	fbGen := &blockingFeedbackGenerator{}

	ctrl := NewLoopController(storage, engine, fbGen, 5)

	if err := ctrl.Start(context.Background(), "SPEC-TEST-001"); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	if err := ctrl.Cancel(); err != nil {
		t.Fatalf("Cancel failed: %v", err)
	}

	status := ctrl.Status()
	if status.Running {
		t.Error("Running = true, want false after cancel")
	}
	if status.SpecID != "" {
		t.Errorf("SpecID = %q, want empty after cancel", status.SpecID)
	}

	// State file should be deleted.
	if _, ok := storage.states["SPEC-TEST-001"]; ok {
		t.Error("state file should be deleted after cancel")
	}
}

func TestLoopController_CancelPaused(t *testing.T) {
	t.Parallel()

	storage := newMockStorage()
	engine := &mockDecisionEngine{
		decideFn: func(_ context.Context, _ *LoopState, _ *Feedback) (*Decision, error) {
			return &Decision{Action: ActionRequestReview, Reason: "review"}, nil
		},
	}
	fbGen := &mockFeedbackGenerator{fb: &Feedback{
		TestsPassed: 5, TestsFailed: 1, BuildSuccess: true, Coverage: 80.0,
	}}

	ctrl := NewLoopController(storage, engine, fbGen, 5)

	if err := ctrl.Start(context.Background(), "SPEC-TEST-001"); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	waitDone(t, ctrl, 5*time.Second)

	// Now paused due to request_review.
	if err := ctrl.Cancel(); err != nil {
		t.Fatalf("Cancel paused loop failed: %v", err)
	}

	status := ctrl.Status()
	if status.SpecID != "" {
		t.Errorf("SpecID = %q, want empty after cancel", status.SpecID)
	}
	if _, ok := storage.states["SPEC-TEST-001"]; ok {
		t.Error("state should be deleted after cancel")
	}
}

func TestLoopController_Status(t *testing.T) {
	t.Parallel()

	storage := newMockStorage()
	engine := &mockDecisionEngine{
		decideFn: func(_ context.Context, _ *LoopState, _ *Feedback) (*Decision, error) {
			return &Decision{Action: ActionConverge, Converged: true}, nil
		},
	}
	fbGen := &mockFeedbackGenerator{fb: &Feedback{
		TestsPassed: 10, BuildSuccess: true, Coverage: 90.0,
	}}

	ctrl := NewLoopController(storage, engine, fbGen, 5)

	// Status before start.
	status := ctrl.Status()
	if status.SpecID != "" {
		t.Errorf("SpecID = %q, want empty before start", status.SpecID)
	}
	if status.Running {
		t.Error("Running = true, want false before start")
	}

	if err := ctrl.Start(context.Background(), "SPEC-TEST-001"); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	waitDone(t, ctrl, 5*time.Second)

	status = ctrl.Status()
	if status.SpecID != "SPEC-TEST-001" {
		t.Errorf("SpecID = %q, want SPEC-TEST-001", status.SpecID)
	}
	if status.MaxIter != 5 {
		t.Errorf("MaxIter = %d, want 5", status.MaxIter)
	}
}

func TestLoopController_RecordFeedback(t *testing.T) {
	t.Parallel()

	storage := newMockStorage()
	engine := &mockDecisionEngine{
		decideFn: func(_ context.Context, _ *LoopState, _ *Feedback) (*Decision, error) {
			return &Decision{Action: ActionRequestReview, Reason: "review"}, nil
		},
	}
	fbGen := &mockFeedbackGenerator{fb: &Feedback{
		TestsPassed: 5, TestsFailed: 1, BuildSuccess: true, Coverage: 80.0,
	}}

	ctrl := NewLoopController(storage, engine, fbGen, 5)

	if err := ctrl.Start(context.Background(), "SPEC-TEST-001"); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	waitDone(t, ctrl, 5*time.Second)

	// Record external feedback while paused.
	fb := Feedback{
		Phase: PhaseTest, Iteration: 1,
		TestsPassed: 42, TestsFailed: 3,
		LintErrors: 1, BuildSuccess: true,
		Coverage: 81.5,
	}
	if err := ctrl.RecordFeedback(fb); err != nil {
		t.Fatalf("RecordFeedback failed: %v", err)
	}

	// Verify feedback was saved.
	saved := storage.states["SPEC-TEST-001"]
	if saved == nil {
		t.Fatal("state not persisted")
	}
	found := false
	for _, f := range saved.Feedback {
		if f.TestsPassed == 42 && f.Coverage == 81.5 {
			found = true
			break
		}
	}
	if !found {
		t.Error("recorded feedback not found in state")
	}
}

func TestLoopController_RecordFeedbackNotRunning(t *testing.T) {
	t.Parallel()

	storage := newMockStorage()
	engine := &mockDecisionEngine{
		decideFn: func(_ context.Context, _ *LoopState, _ *Feedback) (*Decision, error) {
			return &Decision{Action: ActionContinue}, nil
		},
	}
	fbGen := &mockFeedbackGenerator{}

	ctrl := NewLoopController(storage, engine, fbGen, 5)

	err := ctrl.RecordFeedback(Feedback{})
	if err != ErrLoopNotRunning {
		t.Errorf("expected ErrLoopNotRunning, got: %v", err)
	}
}

func TestLoopController_Convergence(t *testing.T) {
	t.Parallel()

	storage := newMockStorage()
	engine := &mockDecisionEngine{
		decideFn: func(_ context.Context, _ *LoopState, fb *Feedback) (*Decision, error) {
			if fb.TestsFailed == 0 && fb.LintErrors == 0 && fb.BuildSuccess && fb.Coverage >= 85.0 {
				return &Decision{Action: ActionConverge, Converged: true, Reason: "quality gate"}, nil
			}
			return &Decision{Action: ActionContinue}, nil
		},
	}
	fbGen := &mockFeedbackGenerator{fb: &Feedback{
		TestsPassed: 50, TestsFailed: 0, LintErrors: 0,
		BuildSuccess: true, Coverage: 92.3,
	}}

	ctrl := NewLoopController(storage, engine, fbGen, 5)

	if err := ctrl.Start(context.Background(), "SPEC-TEST-001"); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	waitDone(t, ctrl, 5*time.Second)

	status := ctrl.Status()
	if !status.Converged {
		t.Error("Converged = false, want true")
	}
	if status.Running {
		t.Error("Running = true, want false after convergence")
	}

	// State should be deleted on convergence.
	if _, ok := storage.states["SPEC-TEST-001"]; ok {
		t.Error("state should be deleted after convergence")
	}
}

func TestLoopController_MaxIterations(t *testing.T) {
	t.Parallel()

	storage := newMockStorage()
	iterationSeen := atomic.Int32{}
	engine := &mockDecisionEngine{
		decideFn: func(_ context.Context, state *LoopState, _ *Feedback) (*Decision, error) {
			iterationSeen.Store(int32(state.Iteration))
			if state.Iteration >= state.MaxIter {
				return &Decision{Action: ActionAbort, Reason: "max iterations"}, nil
			}
			return &Decision{Action: ActionContinue}, nil
		},
	}
	fbGen := &mockFeedbackGenerator{fb: &Feedback{
		TestsPassed: 5, TestsFailed: 1, BuildSuccess: true, Coverage: 80.0,
	}}

	ctrl := NewLoopController(storage, engine, fbGen, 3)

	if err := ctrl.Start(context.Background(), "SPEC-TEST-001"); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	waitDone(t, ctrl, 5*time.Second)

	status := ctrl.Status()
	if status.Running {
		t.Error("Running = true, want false after max iterations")
	}

	// Should have reached iteration 3.
	seen := int(iterationSeen.Load())
	if seen < 3 {
		t.Errorf("max iteration seen = %d, want >= 3", seen)
	}
}

func TestLoopController_HumanReviewPause(t *testing.T) {
	t.Parallel()

	storage := newMockStorage()
	engine := &mockDecisionEngine{
		decideFn: func(_ context.Context, _ *LoopState, _ *Feedback) (*Decision, error) {
			return &Decision{Action: ActionRequestReview, Reason: "human review"}, nil
		},
	}
	fbGen := &mockFeedbackGenerator{fb: &Feedback{
		TestsPassed: 5, TestsFailed: 1, BuildSuccess: true, Coverage: 80.0,
	}}

	ctrl := NewLoopController(storage, engine, fbGen, 5)

	if err := ctrl.Start(context.Background(), "SPEC-TEST-001"); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	waitDone(t, ctrl, 5*time.Second)

	status := ctrl.Status()
	if status.Running {
		t.Error("Running = true, want false after human review pause")
	}
	if status.Converged {
		t.Error("Converged = true, want false during human review")
	}

	// State should be saved (not deleted) for resumption.
	if _, ok := storage.states["SPEC-TEST-001"]; !ok {
		t.Error("state should be persisted for human review resumption")
	}
}

func TestLoopController_DefaultMaxIterations(t *testing.T) {
	t.Parallel()

	storage := newMockStorage()
	engine := &mockDecisionEngine{
		decideFn: func(_ context.Context, _ *LoopState, _ *Feedback) (*Decision, error) {
			return &Decision{Action: ActionConverge, Converged: true}, nil
		},
	}
	fbGen := &mockFeedbackGenerator{fb: &Feedback{
		BuildSuccess: true, Coverage: 90.0,
	}}

	// Pass 0 to trigger default.
	ctrl := NewLoopController(storage, engine, fbGen, 0)

	if err := ctrl.Start(context.Background(), "SPEC-TEST-001"); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	waitDone(t, ctrl, 5*time.Second)

	status := ctrl.Status()
	if status.MaxIter != 5 {
		t.Errorf("MaxIter = %d, want 5 (default)", status.MaxIter)
	}
}

func TestLoopController_ConcurrentStart(t *testing.T) {
	t.Parallel()

	storage := newMockStorage()
	engine := &mockDecisionEngine{
		decideFn: func(_ context.Context, _ *LoopState, _ *Feedback) (*Decision, error) {
			return &Decision{Action: ActionContinue}, nil
		},
	}
	fbGen := &blockingFeedbackGenerator{}

	ctrl := NewLoopController(storage, engine, fbGen, 5)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := ctrl.Start(ctx, "SPEC-TEST-001"); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Launch multiple concurrent starts.
	errCount := atomic.Int32{}
	done := make(chan struct{})
	for range 5 {
		go func() {
			defer func() { done <- struct{}{} }()
			if err := ctrl.Start(ctx, "SPEC-TEST-001"); err == ErrLoopAlreadyRunning {
				errCount.Add(1)
			}
		}()
	}
	for range 5 {
		<-done
	}

	if errCount.Load() != 5 {
		t.Errorf("expected 5 ErrLoopAlreadyRunning errors, got %d", errCount.Load())
	}

	cancel()
	waitDone(t, ctrl, 5*time.Second)
}

func TestLoopController_CancelNotExist(t *testing.T) {
	t.Parallel()

	storage := newMockStorage()
	engine := &mockDecisionEngine{
		decideFn: func(_ context.Context, _ *LoopState, _ *Feedback) (*Decision, error) {
			return &Decision{Action: ActionContinue}, nil
		},
	}
	fbGen := &mockFeedbackGenerator{}

	ctrl := NewLoopController(storage, engine, fbGen, 5)

	err := ctrl.Cancel()
	if err != ErrLoopNotRunning {
		t.Errorf("expected ErrLoopNotRunning, got: %v", err)
	}
}

func TestLoopController_ResumeFromStorage(t *testing.T) {
	t.Parallel()

	storage := newMockStorage()

	// Pre-populate storage with a saved state.
	savedState := &LoopState{
		SpecID:    "SPEC-TEST-001",
		Phase:     PhaseTest,
		Iteration: 3,
		MaxIter:   5,
		Feedback: []Feedback{
			{Phase: PhaseReview, Iteration: 1, TestsFailed: 3, Coverage: 75.0},
			{Phase: PhaseReview, Iteration: 2, TestsFailed: 1, Coverage: 82.0},
		},
		StartedAt: time.Now().Add(-time.Hour),
		UpdatedAt: time.Now(),
	}
	if err := storage.SaveState(savedState); err != nil {
		t.Fatalf("pre-save failed: %v", err)
	}

	engine := &mockDecisionEngine{
		decideFn: func(_ context.Context, _ *LoopState, _ *Feedback) (*Decision, error) {
			return &Decision{Action: ActionConverge, Converged: true, Reason: "done"}, nil
		},
	}
	fbGen := &mockFeedbackGenerator{fb: &Feedback{
		TestsPassed: 10, BuildSuccess: true, Coverage: 90.0,
	}}

	ctrl := NewLoopController(storage, engine, fbGen, 5)

	if err := ctrl.ResumeFromStorage(context.Background(), "SPEC-TEST-001"); err != nil {
		t.Fatalf("ResumeFromStorage failed: %v", err)
	}
	waitDone(t, ctrl, 5*time.Second)

	status := ctrl.Status()
	if !status.Converged {
		t.Error("Converged = false, want true")
	}
	if status.Running {
		t.Error("Running = true, want false after convergence")
	}
	// Feedback history should be preserved (2 original + new ones).
	if status.Iteration < 3 {
		t.Errorf("Iteration = %d, want >= 3 (restored)", status.Iteration)
	}
}

func TestLoopController_ResumeFromStorageNotExist(t *testing.T) {
	t.Parallel()

	storage := newMockStorage()
	engine := &mockDecisionEngine{
		decideFn: func(_ context.Context, _ *LoopState, _ *Feedback) (*Decision, error) {
			return &Decision{Action: ActionContinue}, nil
		},
	}
	fbGen := &mockFeedbackGenerator{}

	ctrl := NewLoopController(storage, engine, fbGen, 5)

	err := ctrl.ResumeFromStorage(context.Background(), "SPEC-NONEXIST")
	if err == nil {
		t.Fatal("expected error for non-existent state, got nil")
	}
}

func TestLoopController_ResumeFromStorageAlreadyRunning(t *testing.T) {
	t.Parallel()

	storage := newMockStorage()
	engine := &mockDecisionEngine{
		decideFn: func(_ context.Context, _ *LoopState, _ *Feedback) (*Decision, error) {
			return &Decision{Action: ActionContinue}, nil
		},
	}
	fbGen := &blockingFeedbackGenerator{}

	ctrl := NewLoopController(storage, engine, fbGen, 5)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := ctrl.Start(ctx, "SPEC-TEST-001"); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	err := ctrl.ResumeFromStorage(ctx, "SPEC-TEST-001")
	if err != ErrLoopNotPaused {
		t.Errorf("expected ErrLoopNotPaused, got: %v", err)
	}

	cancel()
	waitDone(t, ctrl, 5*time.Second)
}

func TestLoopController_AbortDecision(t *testing.T) {
	t.Parallel()

	storage := newMockStorage()
	engine := &mockDecisionEngine{
		decideFn: func(_ context.Context, _ *LoopState, _ *Feedback) (*Decision, error) {
			return &Decision{Action: ActionAbort, Reason: "forced abort"}, nil
		},
	}
	fbGen := &mockFeedbackGenerator{fb: &Feedback{
		TestsPassed: 5, TestsFailed: 1, BuildSuccess: true, Coverage: 80.0,
	}}

	ctrl := NewLoopController(storage, engine, fbGen, 5)

	if err := ctrl.Start(context.Background(), "SPEC-TEST-001"); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	waitDone(t, ctrl, 5*time.Second)

	status := ctrl.Status()
	if status.Running {
		t.Error("Running = true, want false after abort")
	}
	if status.Converged {
		t.Error("Converged = true, want false after abort")
	}

	// State should be deleted on abort.
	if _, ok := storage.states["SPEC-TEST-001"]; ok {
		t.Error("state should be deleted after abort")
	}
}

func TestLoopController_DoneBeforeStart(t *testing.T) {
	t.Parallel()

	storage := newMockStorage()
	engine := &mockDecisionEngine{
		decideFn: func(_ context.Context, _ *LoopState, _ *Feedback) (*Decision, error) {
			return &Decision{Action: ActionContinue}, nil
		},
	}
	fbGen := &mockFeedbackGenerator{}

	ctrl := NewLoopController(storage, engine, fbGen, 5)

	done := ctrl.Done()
	if done != nil {
		t.Error("Done() should return nil before Start")
	}
}

func TestLoopController_ResumeFromReviewPhaseAdvances(t *testing.T) {
	t.Parallel()

	storage := newMockStorage()

	// Pre-save state at review phase (simulating request_review pause).
	savedState := &LoopState{
		SpecID:    "SPEC-TEST-001",
		Phase:     PhaseReview,
		Iteration: 2,
		MaxIter:   5,
		Feedback:  []Feedback{},
		StartedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := storage.SaveState(savedState); err != nil {
		t.Fatalf("pre-save failed: %v", err)
	}

	engine := &mockDecisionEngine{
		decideFn: func(_ context.Context, _ *LoopState, _ *Feedback) (*Decision, error) {
			return &Decision{Action: ActionConverge, Converged: true, Reason: "done"}, nil
		},
	}
	fbGen := &mockFeedbackGenerator{fb: &Feedback{
		TestsPassed: 10, BuildSuccess: true, Coverage: 90.0,
	}}

	ctrl := NewLoopController(storage, engine, fbGen, 5)

	if err := ctrl.ResumeFromStorage(context.Background(), "SPEC-TEST-001"); err != nil {
		t.Fatalf("ResumeFromStorage failed: %v", err)
	}
	waitDone(t, ctrl, 5*time.Second)

	status := ctrl.Status()
	// After resume from review, iteration should advance.
	if status.Iteration < 3 {
		t.Errorf("Iteration = %d, want >= 3 (advanced from review)", status.Iteration)
	}
}
