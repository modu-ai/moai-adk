package cli

import (
	"context"
	"io"

	"github.com/modu-ai/moai-adk/internal/hook"
	"github.com/modu-ai/moai-adk/internal/rank"
	"github.com/modu-ai/moai-adk/internal/update"
)

// --- Mock implementations for CLI dependency testing ---

// mockHookProtocol implements hook.Protocol for testing.
type mockHookProtocol struct {
	readInputFunc   func(r io.Reader) (*hook.HookInput, error)
	writeOutputFunc func(w io.Writer, output *hook.HookOutput) error
}

func (m *mockHookProtocol) ReadInput(r io.Reader) (*hook.HookInput, error) {
	if m.readInputFunc != nil {
		return m.readInputFunc(r)
	}
	return &hook.HookInput{}, nil
}

func (m *mockHookProtocol) WriteOutput(w io.Writer, output *hook.HookOutput) error {
	if m.writeOutputFunc != nil {
		return m.writeOutputFunc(w, output)
	}
	return nil
}

// mockHookRegistry implements hook.Registry for testing.
type mockHookRegistry struct {
	registerFunc func(handler hook.Handler)
	dispatchFunc func(ctx context.Context, event hook.EventType, input *hook.HookInput) (*hook.HookOutput, error)
	handlersFunc func(event hook.EventType) []hook.Handler
}

func (m *mockHookRegistry) Register(handler hook.Handler) {
	if m.registerFunc != nil {
		m.registerFunc(handler)
	}
}

func (m *mockHookRegistry) Dispatch(ctx context.Context, event hook.EventType, input *hook.HookInput) (*hook.HookOutput, error) {
	if m.dispatchFunc != nil {
		return m.dispatchFunc(ctx, event, input)
	}
	return hook.NewAllowOutput(), nil
}

func (m *mockHookRegistry) Handlers(event hook.EventType) []hook.Handler {
	if m.handlersFunc != nil {
		return m.handlersFunc(event)
	}
	return nil
}

// mockUpdateChecker implements update.Checker for testing.
type mockUpdateChecker struct {
	checkLatestFunc   func(ctx context.Context) (*update.VersionInfo, error)
	isUpdateAvailFunc func(current string) (bool, *update.VersionInfo, error)
}

func (m *mockUpdateChecker) CheckLatest(ctx context.Context) (*update.VersionInfo, error) {
	if m.checkLatestFunc != nil {
		return m.checkLatestFunc(ctx)
	}
	return &update.VersionInfo{Version: "1.0.0", URL: "https://example.com/moai-binary"}, nil
}

func (m *mockUpdateChecker) IsUpdateAvailable(current string) (bool, *update.VersionInfo, error) {
	if m.isUpdateAvailFunc != nil {
		return m.isUpdateAvailFunc(current)
	}
	return false, nil, nil
}

// mockUpdateOrchestrator implements update.Orchestrator for testing.
type mockUpdateOrchestrator struct {
	updateFunc func(ctx context.Context) (*update.UpdateResult, error)
}

func (m *mockUpdateOrchestrator) Update(ctx context.Context) (*update.UpdateResult, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx)
	}
	return &update.UpdateResult{PreviousVersion: "v0.0.0", NewVersion: "v0.0.1"}, nil
}

// mockRankClient implements rank.Client for testing.
type mockRankClient struct {
	checkStatusFunc    func(ctx context.Context) (*rank.ApiStatus, error)
	getUserRankFunc    func(ctx context.Context) (*rank.UserRank, error)
	getLeaderboardFunc func(ctx context.Context, period string, limit, offset int) ([]rank.LeaderboardEntry, error)
	submitSessionFunc  func(ctx context.Context, session *rank.SessionSubmission) error
	submitBatchFunc    func(ctx context.Context, sessions []*rank.SessionSubmission) (*rank.BatchResult, error)
}

func (m *mockRankClient) CheckStatus(ctx context.Context) (*rank.ApiStatus, error) {
	if m.checkStatusFunc != nil {
		return m.checkStatusFunc(ctx)
	}
	return &rank.ApiStatus{Status: "ok"}, nil
}

func (m *mockRankClient) GetUserRank(ctx context.Context) (*rank.UserRank, error) {
	if m.getUserRankFunc != nil {
		return m.getUserRankFunc(ctx)
	}
	return &rank.UserRank{
		Username: "testuser",
		Stats: &rank.UserRankStats{
			TotalTokens:   1000,
			TotalSessions: 5,
		},
	}, nil
}

func (m *mockRankClient) GetLeaderboard(ctx context.Context, period string, limit, offset int) ([]rank.LeaderboardEntry, error) {
	if m.getLeaderboardFunc != nil {
		return m.getLeaderboardFunc(ctx, period, limit, offset)
	}
	return nil, nil
}

func (m *mockRankClient) SubmitSession(ctx context.Context, session *rank.SessionSubmission) error {
	if m.submitSessionFunc != nil {
		return m.submitSessionFunc(ctx, session)
	}
	return nil
}

func (m *mockRankClient) SubmitSessionsBatch(ctx context.Context, sessions []*rank.SessionSubmission) (*rank.BatchResult, error) {
	if m.submitBatchFunc != nil {
		return m.submitBatchFunc(ctx, sessions)
	}
	return &rank.BatchResult{Success: true}, nil
}

// mockCredentialStore implements rank.CredentialStore for testing.
type mockCredentialStore struct {
	saveFunc     func(creds *rank.Credentials) error
	loadFunc     func() (*rank.Credentials, error)
	deleteFunc   func() error
	hasCredsFunc func() bool
	getKeyFunc   func() (string, error)
}

func (m *mockCredentialStore) Save(creds *rank.Credentials) error {
	if m.saveFunc != nil {
		return m.saveFunc(creds)
	}
	return nil
}

func (m *mockCredentialStore) Load() (*rank.Credentials, error) {
	if m.loadFunc != nil {
		return m.loadFunc()
	}
	return nil, nil
}

func (m *mockCredentialStore) Delete() error {
	if m.deleteFunc != nil {
		return m.deleteFunc()
	}
	return nil
}

func (m *mockCredentialStore) HasCredentials() bool {
	if m.hasCredsFunc != nil {
		return m.hasCredsFunc()
	}
	return false
}

func (m *mockCredentialStore) GetAPIKey() (string, error) {
	if m.getKeyFunc != nil {
		return m.getKeyFunc()
	}
	return "", nil
}

// mockBrowser implements rank.BrowserOpener for testing.
// It records the URL that would be opened without actually opening a browser.
type mockBrowser struct {
	openFunc func(url string) error
	lastURL  string
}

func (m *mockBrowser) Open(url string) error {
	m.lastURL = url
	if m.openFunc != nil {
		return m.openFunc(url)
	}
	return nil
}

// mockHandler implements hook.Handler for testing.
type mockHandler struct {
	eventType hook.EventType
}

func (m *mockHandler) Handle(_ context.Context, _ *hook.HookInput) (*hook.HookOutput, error) {
	return hook.NewAllowOutput(), nil
}

func (m *mockHandler) EventType() hook.EventType {
	return m.eventType
}
