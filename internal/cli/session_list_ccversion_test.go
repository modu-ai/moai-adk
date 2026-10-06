package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/session"
)

// saveResolveCCVersionsSeam captures the exported resolver seam and restores
// it at cleanup, so an injected fixture cannot leak into another test's run.
func saveResolveCCVersionsSeam(t *testing.T) {
	t.Helper()
	prev := session.ResolveCCVersions
	t.Cleanup(func() { session.ResolveCCVersions = prev })
}

// writeSessionRegistryFixture writes the registry file QueryActiveWork reads
// (relative to cwd, which withTempRegistry pinned) with the given entries.
func writeSessionRegistryFixture(t *testing.T, entries []session.Entry) {
	t.Helper()
	path := filepath.Join(session.DefaultRegistryPath)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir registry dir: %v", err)
	}
	data, err := json.Marshal(entries)
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write fixture registry: %v", err)
	}
}

// fixtureEntry is one registry row the surfaces tests share.
func fixtureEntry(sessionID string, pid int) session.Entry {
	now := time.Now().UTC()
	return session.Entry{
		SessionID:     sessionID,
		SpecID:        "SPEC-CC-VERSION-FIXTURE",
		Phase:         "run",
		StartedAt:     now,
		LastHeartbeat: now,
		PID:           pid,
		Host:          "fixture-host",
		CWD:           "/tmp/fixture",
	}
}

// TestSessionListCCVersion (AC-SCV-005, REQ-SCV-005) — with injected reads
// (running 2.1.281, installed 2.1.288), `moai session list --cc-version
// --json` carries both values for the entry and the human output names both;
// a degrading fixture (dead pid → unknown) still lists the entry, renders
// unknown in both outputs, and the command exits 0.
func TestSessionListCCVersion(t *testing.T) {
	t.Run("json carries both reads", func(t *testing.T) {
		withTempRegistry(t)
		saveResolveCCVersionsSeam(t)
		writeSessionRegistryFixture(t, []session.Entry{fixtureEntry("uuid-cc-1", 4242)})
		session.ResolveCCVersions = func(pid int) session.CCVersions {
			return session.CCVersions{Running: "2.1.281", Installed: "2.1.288"}
		}

		out, err := runSession(t, "list", "--cc-version", "--json")
		if err != nil {
			t.Fatalf("list --cc-version --json err: %v (out=%s)", err, out)
		}
		var entries []map[string]any
		if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &entries); err != nil {
			t.Fatalf("output not a JSON array: %v (out=%s)", err, out)
		}
		if len(entries) != 1 {
			t.Fatalf("entries = %d, want 1", len(entries))
		}
		view, ok := entries[0]["cc_version"].(map[string]any)
		if !ok {
			t.Fatalf("entry carries no cc_version object: %v", entries[0])
		}
		if view["running"] != "2.1.281" {
			t.Errorf("cc_version.running = %v, want 2.1.281", view["running"])
		}
		if view["installed"] != "2.1.288" {
			t.Errorf("cc_version.installed = %v, want 2.1.288", view["installed"])
		}
	})
	t.Run("human output names both", func(t *testing.T) {
		withTempRegistry(t)
		saveResolveCCVersionsSeam(t)
		// The human rendering shortens ids to 8 chars (the default path's
		// existing shortID contract), so the fixture id fits in 8.
		writeSessionRegistryFixture(t, []session.Entry{fixtureEntry("uuid-cc2", 4242)})
		session.ResolveCCVersions = func(pid int) session.CCVersions {
			return session.CCVersions{Running: "2.1.281", Installed: "2.1.288"}
		}

		out, err := runSession(t, "list", "--cc-version")
		if err != nil {
			t.Fatalf("list --cc-version err: %v (out=%s)", err, out)
		}
		for _, want := range []string{"uuid-cc2", "2.1.281", "2.1.288"} {
			if !strings.Contains(out, want) {
				t.Errorf("human output missing %q; got: %s", want, out)
			}
		}
	})
	t.Run("degraded entry renders unknown and still lists, exit 0", func(t *testing.T) {
		withTempRegistry(t)
		saveResolveCCVersionsSeam(t)
		writeSessionRegistryFixture(t, []session.Entry{fixtureEntry("uuid-cc-3", 4242)})
		session.ResolveCCVersions = func(pid int) session.CCVersions {
			return session.CCVersions{Running: session.UnknownCCVersion, Installed: "2.1.288"}
		}

		out, err := runSession(t, "list", "--cc-version", "--json")
		if err != nil {
			t.Fatalf("degraded run must exit 0, got err: %v (out=%s)", err, out)
		}
		var entries []map[string]any
		if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &entries); err != nil {
			t.Fatalf("output not a JSON array: %v (out=%s)", err, out)
		}
		if len(entries) != 1 {
			t.Fatalf("degraded entry was omitted: %d entries, want 1", len(entries))
		}
		view, ok := entries[0]["cc_version"].(map[string]any)
		if !ok {
			t.Fatalf("degraded entry carries no cc_version object: %v", entries[0])
		}
		if view["running"] != session.UnknownCCVersion {
			t.Errorf("degraded cc_version.running = %v, want %q", view["running"], session.UnknownCCVersion)
		}
	})
}

// TestSessionListDefaultNoProbe (AC-SCV-006, REQ-SCV-006) — without
// --cc-version the command performs zero probes and each entry's JSON key set
// is exactly today's eight; this is the preservation guard on the
// orchestrator pre-spawn batch contract and fails the moment enrichment
// leaks into the default path.
func TestSessionListDefaultNoProbe(t *testing.T) {
	withTempRegistry(t)
	saveResolveCCVersionsSeam(t)
	writeSessionRegistryFixture(t, []session.Entry{fixtureEntry("uuid-noprobe-1", 4242)})

	probes := 0
	session.ResolveCCVersions = func(pid int) session.CCVersions {
		probes++
		return session.CCVersions{Running: "2.1.999", Installed: "2.1.999"}
	}

	out, err := runSession(t, "list", "--json")
	if err != nil {
		t.Fatalf("list --json err: %v (out=%s)", err, out)
	}
	if probes != 0 {
		t.Fatalf("default path probed %d time(s), want 0", probes)
	}
	var entries []map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &entries); err != nil {
		t.Fatalf("output not a JSON array: %v (out=%s)", err, out)
	}
	if len(entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(entries))
	}
	keys := make([]string, 0, len(entries[0]))
	for k := range entries[0] {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	want := []string{
		"session_id", "spec_id", "phase", "started_at",
		"last_heartbeat", "pid", "host", "cwd",
	}
	sort.Strings(want)
	if strings.Join(keys, ",") != strings.Join(want, ",") {
		t.Fatalf("default JSON key set = %v, want exactly %v", keys, want)
	}

	// The human default path is probe-free too.
	if _, err := runSession(t, "list"); err != nil {
		t.Fatalf("human list err: %v", err)
	}
	if probes != 0 {
		t.Fatalf("human default path probed %d time(s), want 0", probes)
	}
}
