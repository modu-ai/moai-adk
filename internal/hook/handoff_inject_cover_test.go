package hook

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/hook/handoff"
)

// TestHandoffConfig_Fallbacks covers the config-resolution edge branches:
// nil provider, nil Get(), and an unknown mode value (all → manual no-op).
func TestHandoffConfig_Fallbacks(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		cfg  ConfigProvider
	}{
		{"nil_provider", nil},
		{"nil_get", &mockConfigProvider{cfg: nil}},
		{"unknown_mode", &mockConfigProvider{cfg: &config.Config{Handoff: config.HandoffConfig{Mode: "bogus"}}}},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			pd := t.TempDir()
			mustSavePending(t, pd, livePending("resume"))
			before, _ := readFileBytes(t, pd)

			h := NewHandoffInjectHandler(tc.cfg)
			out, err := h.Handle(context.Background(), injectInput("clear", pd))
			if err != nil {
				t.Fatalf("Handle: %v", err)
			}
			if additionalContextOf(out) != "" {
				t.Errorf("%s: must be a no-op (no injection)", tc.name)
			}
			after, _ := readFileBytes(t, pd)
			if before != after {
				t.Errorf("%s: pending must be byte-unchanged", tc.name)
			}
		})
	}
}

func readFileBytes(t *testing.T, pd string) (string, bool) {
	t.Helper()
	rec, present, err := handoff.ReadPending(pd)
	if err != nil || !present || rec == nil {
		return "", false
	}
	return rec.Body, true
}

// TestConsumeNonce_SessionAttribution covers the session-8 hex attribution path
// (isHex8 true) and the non-hex fallback path.
func TestConsumeNonce_SessionAttribution(t *testing.T) {
	t.Parallel()

	// A UUID's first 8 chars are hex → used verbatim (attribution).
	if got := consumeNonce("a1b2c3d4-e5f6-7890-abcd-ef1234567890"); got != "a1b2c3d4" {
		t.Errorf("hex session attribution: got %q, want %q", got, "a1b2c3d4")
	}
	// Uppercase hex first-8 → lowercased.
	if got := consumeNonce("A1B2C3D4EXTRA"); got != "a1b2c3d4" {
		t.Errorf("uppercase hex session: got %q, want %q", got, "a1b2c3d4")
	}
	// Non-hex first-8 → crypto/rand 8-hex (not the raw session).
	got := consumeNonce("ZZZZZZZZ-not-hex")
	if !isHex8(got) {
		t.Errorf("non-hex session must fall back to 8-hex nonce, got %q", got)
	}
	// Short session → crypto/rand 8-hex.
	got = consumeNonce("abc")
	if !isHex8(got) {
		t.Errorf("short session must fall back to 8-hex nonce, got %q", got)
	}
}

// TestIsHex8 covers the isHex8 helper directly.
func TestIsHex8(t *testing.T) {
	t.Parallel()

	tests := map[string]bool{
		"a1b2c3d4": true,
		"00000000": true,
		"ffffffff": true,
		"A1B2C3D4": false, // uppercase not accepted by isHex8 itself
		"g1b2c3d4": false, // 'g' not hex
		"a1b2c3d":  false, // 7 chars
		"a1b2c3d45": false, // 9 chars
		"":         false,
	}
	for in, want := range tests {
		if got := isHex8(in); got != want {
			t.Errorf("isHex8(%q) = %v, want %v", in, got, want)
		}
	}
}

// TestHandoffStale_ZeroTime covers the zero-saved_at branch (never stale).
func TestHandoffStale_ZeroTime(t *testing.T) {
	t.Parallel()

	if handoffStale(&handoff.PendingRecord{}, time.Now()) {
		t.Error("a record with zero SavedAt must not be considered stale")
	}
	old := &handoff.PendingRecord{SavedAt: time.Now().Add(-config.DefaultHandoffStaleTTL - time.Hour)}
	if !handoffStale(old, time.Now()) {
		t.Error("a record older than the TTL must be stale")
	}
	fresh := &handoff.PendingRecord{SavedAt: time.Now()}
	if handoffStale(fresh, time.Now()) {
		t.Error("a fresh record must not be stale")
	}
}

// TestHandle_EmptyProjectDir covers the projectDir=="" branch (no-op).
func TestHandle_EmptyProjectDir(t *testing.T) {
	t.Parallel()

	h := NewHandoffInjectHandler(autoCfgProvider(false))
	out, err := h.Handle(context.Background(), &HookInput{Source: "clear", CWD: "", ProjectDir: ""})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if additionalContextOf(out) != "" {
		t.Error("empty projectDir must be a no-op")
	}
}

// TestHandle_AbsentPending covers the absent-pending no-op branch.
func TestHandle_AbsentPending(t *testing.T) {
	t.Parallel()

	pd := t.TempDir() // no pending.json written
	h := NewHandoffInjectHandler(autoCfgProvider(false))
	out, err := h.Handle(context.Background(), injectInput("clear", pd))
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if additionalContextOf(out) != "" {
		t.Error("absent pending must be a no-op")
	}
}

// TestRenderHandoffContext_AllDirectives covers the ultracode + goal render branches.
func TestRenderHandoffContext_AllDirectives(t *testing.T) {
	t.Parallel()

	rec := &handoff.PendingRecord{
		Body:                 "body",
		ConversationLanguage: "en",
		Directives:           handoff.Directives{Ultrathink: true, Ultracode: true, Goal: "tests pass AND lint clean"},
	}
	got := renderHandoffContext(rec)
	if strings.Contains(got, "xhigh") {
		t.Error("must not contain 'xhigh'")
	}
	for _, want := range []string{"ultrathink", "/effort ultracode", "/goal ", "tests pass AND lint clean", "body"} {
		if !strings.Contains(got, want) {
			t.Errorf("render missing %q in:\n%s", want, got)
		}
	}

	// No directives → no restoration-guidance block, body still present.
	plain := renderHandoffContext(&handoff.PendingRecord{Body: "just body", ConversationLanguage: "en"})
	if strings.Contains(plain, "/effort ultracode") {
		t.Error("no-directive render should omit ultracode guidance")
	}
	if !strings.Contains(plain, "just body") {
		t.Error("no-directive render should still carry the body")
	}
}

// TestConsumedDir_Path covers the ConsumedDir path helper directly (attribution
// artifact: it is exercised by claimAndInject but attributed to the hook binary).
func TestConsumedDir_Path(t *testing.T) {
	t.Parallel()

	pd := "/tmp/proj"
	got := handoff.ConsumedDir(pd)
	// ConsumedDir is a real filesystem path (MkdirAll/rename target), so it is
	// built with filepath.Join and uses OS-native separators. Build the
	// expectation the same way rather than hardcoding forward slashes.
	want := filepath.Join(pd, ".moai", "state", "handoff", "consumed")
	if got != want {
		t.Errorf("ConsumedDir: got %q, want %q", got, want)
	}
}
