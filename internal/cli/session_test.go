package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// withTempRegistry switches into a fresh temp dir for the duration of the
// test so that session.DefaultRegistryPath ('.moai/state/active-sessions.json')
// resolves to a per-test isolated location. This avoids touching the real
// project registry.
func withTempRegistry(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".moai", "state"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	prev, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(prev) })
	return dir
}

func runSession(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := newSessionCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return buf.String(), err
}

// TestSessionRegisterSmoke verifies the register verb populates the
// registry and emits expected output.
func TestSessionRegisterSmoke(t *testing.T) {
	withTempRegistry(t)

	cases := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "human-readable register",
			args: []string{"register", "uuid-cli-1", "SPEC-CLI", "plan"},
			want: "OK",
		},
		{
			name: "json register",
			args: []string{"register", "uuid-cli-2", "SPEC-CLI", "run", "--json"},
			want: "\"action\": \"register\"",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := runSession(t, tc.args...)
			if err != nil {
				t.Fatalf("register err: %v (out=%s)", err, out)
			}
			if !strings.Contains(out, tc.want) {
				t.Errorf("output missing %q. Got: %s", tc.want, out)
			}
		})
	}
}

// TestSessionListSmoke verifies list returns valid JSON when --json
// passed, and human-readable text otherwise.
func TestSessionListSmoke(t *testing.T) {
	withTempRegistry(t)

	// Empty registry: list --json returns [] (valid JSON array).
	out, err := runSession(t, "list", "--json")
	if err != nil {
		t.Fatalf("list err: %v", err)
	}
	var arr []map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &arr); err != nil {
		t.Errorf("list --json output not valid JSON array: %v (out=%s)", err, out)
	}
	if len(arr) != 0 {
		t.Errorf("empty list: got %d entries, want 0", len(arr))
	}

	// Register two entries, one matching the filter.
	if _, err := runSession(t, "register", "uuid-A", "SPEC-A", "plan"); err != nil {
		t.Fatal(err)
	}
	if _, err := runSession(t, "register", "uuid-B", "SPEC-B", "run"); err != nil {
		t.Fatal(err)
	}

	out, err = runSession(t, "list", "--json", "--filter-spec=SPEC-A")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &arr); err != nil {
		t.Errorf("filtered list --json: %v (out=%s)", err, out)
	}
	if len(arr) != 1 {
		t.Errorf("filtered list: got %d entries, want 1", len(arr))
	}

	// Human-readable list (no --json).
	out, err = runSession(t, "list")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "uuid-A") || !strings.Contains(out, "uuid-B") {
		t.Errorf("human-readable list missing entries: %s", out)
	}
}

// TestSessionHeartbeatSmoke verifies heartbeat updates the registry.
func TestSessionHeartbeatSmoke(t *testing.T) {
	withTempRegistry(t)
	if _, err := runSession(t, "register", "uuid-hb", "SPEC-A", "plan"); err != nil {
		t.Fatal(err)
	}
	out, err := runSession(t, "heartbeat", "uuid-hb", "--json")
	if err != nil {
		t.Fatalf("heartbeat err: %v", err)
	}
	if !strings.Contains(out, "\"action\": \"heartbeat\"") {
		t.Errorf("heartbeat output missing action: %s", out)
	}
}

// TestSessionDeregisterSmoke verifies deregister removes the entry and is
// idempotent (second invocation returns OK without error).
func TestSessionDeregisterSmoke(t *testing.T) {
	withTempRegistry(t)
	if _, err := runSession(t, "register", "uuid-de", "SPEC-A", "plan"); err != nil {
		t.Fatal(err)
	}
	// First deregister.
	if _, err := runSession(t, "deregister", "uuid-de"); err != nil {
		t.Fatal(err)
	}
	// Second deregister on missing — must succeed (idempotent).
	if _, err := runSession(t, "deregister", "uuid-de"); err != nil {
		t.Errorf("second deregister: want nil, got %v", err)
	}
	// list must show 0 entries.
	out, err := runSession(t, "list", "--json")
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out) != "[]" {
		t.Errorf("after deregister: want [], got %s", out)
	}
}

// TestSessionPurgeSmoke verifies purge returns purged_count=0 on a fresh
// registry (no stale entries possible immediately after register).
func TestSessionPurgeSmoke(t *testing.T) {
	withTempRegistry(t)
	if _, err := runSession(t, "register", "uuid-fresh", "SPEC-A", "plan"); err != nil {
		t.Fatal(err)
	}
	out, err := runSession(t, "purge", "--json")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "\"purged_count\": 0") {
		t.Errorf("fresh purge: want purged_count=0, got %s", out)
	}
}

// TestSessionFiveVerbsHelp verifies that `moai session --help` lists
// exactly 5 verbs (REQ-COORD-021 / AC-COORD-013).
func TestSessionFiveVerbsHelp(t *testing.T) {
	cmd := newSessionCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, verb := range []string{"register", "heartbeat", "deregister", "list", "purge"} {
		if !strings.Contains(out, verb) {
			t.Errorf("help output missing verb %q. Got: %s", verb, out)
		}
	}
}

// TestSessionListJSONParseability verifies that the list --json output is
// always a valid JSON array (orchestrator pre-spawn check depends on this).
func TestSessionListJSONParseability(t *testing.T) {
	withTempRegistry(t)
	out, err := runSession(t, "list", "--json", "--filter-spec=SPEC-NONEXISTENT")
	if err != nil {
		t.Fatal(err)
	}
	trimmed := strings.TrimSpace(out)
	if trimmed != "[]" {
		// Not strictly required to be literal "[]" — could be "[\n]" or similar
		// — but it MUST parse as a JSON array.
		var arr []any
		if err := json.Unmarshal([]byte(trimmed), &arr); err != nil {
			t.Errorf("filtered empty list --json: %v (out=%q)", err, trimmed)
		}
		if len(arr) != 0 {
			t.Errorf("filtered empty list: got %d entries, want 0", len(arr))
		}
	}
}

// TestSessionShortIDHelper checks the local shortID helper (mirror of
// session.shortID; we keep the helper here so the CLI can format displays
// without depending on internal/session's helper visibility).
func TestSessionShortIDHelper(t *testing.T) {
	cases := []struct{ in, want string }{
		{"abc", "abc"},
		{"12345678", "12345678"},
		{"123456789", "12345678"},
		{"deadbeef-cafe-1234", "deadbeef"},
	}
	for _, tc := range cases {
		if got := shortID(tc.in); got != tc.want {
			t.Errorf("shortID(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// TestSessionEmitOKBothModes covers both JSON and human-readable paths of
// emitOK so coverage stays >= 85%.
func TestSessionEmitOKBothModes(t *testing.T) {
	cmd := newSessionListCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	if err := emitOK(cmd, true, map[string]any{"k": "v"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "\"k\": \"v\"") {
		t.Errorf("emitOK json mode missing key: %s", buf.String())
	}
	buf.Reset()
	if err := emitOK(cmd, false, map[string]any{"k": "v"}); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(buf.String(), "OK ") || !strings.Contains(buf.String(), "k=v") {
		t.Errorf("emitOK human mode: %s", buf.String())
	}
}

// silence unused import warning in case strings is unused under build flags
var _ = fmt.Sprintf
