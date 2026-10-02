// Package statusline tests for SPEC-QUOTA-AWARE-SCHEDULING-001 M1 (card t1347):
// the rate-limit windows the per-session telemetry record carries (REQ-QAS-001),
// its schema version and window-less byte identity (REQ-QAS-002), the throttle
// under the widened payload with its heartbeat (REQ-QAS-003), and the sticky
// first-observed-exhausted time (REQ-QAS-004).
package statusline

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// qasMem is the fixed context-usage input every writer test below shares, so
// the window behaviour is the only thing that varies between writes.
var qasMem = MemoryData{Available: true, ContextWindowSize: 200_000, TokensUsed: 50_000}

// qasT0 is the injected clock origin of the writer tests (no wall clock is read
// for a verdict).
var qasT0 = time.Date(2026, 10, 2, 12, 0, 0, 0, time.UTC)

// qasWriteAt persists one record through the clock-injected writer.
func qasWriteAt(now time.Time, proj, sessionID string, limits *RateLimitInfo) {
	writeContextUsageAt(now, proj, sessionID, 1, qasMem, handoffStageNone, "", "", limits)
}

// qasWindow builds one stdin rate-limit window.
func qasWindow(pct float64, resetsAt int64) *RateLimitWindow {
	return &RateLimitWindow{UsedPercentage: pct, ResetsAt: resetsAt}
}

// qasStamp renders a capture time the way the writer does.
func qasStamp(t time.Time) string { return t.Format(time.RFC3339Nano) }

// qasAgeFile pushes a record's modification time far into the past so a
// skipped (throttled) write is detectable, and returns that time.
func qasAgeFile(t *testing.T, path string) time.Time {
	t.Helper()
	old := time.Now().Add(-2 * time.Hour)
	if err := os.Chtimes(path, old, old); err != nil {
		t.Fatal(err)
	}
	return old
}

// qasBuildWithLimits renders one statusline for sessionID with the given
// rate_limits stdin object (nil omits the object) and returns the persisted
// record.
func qasBuildWithLimits(t *testing.T, projDir, sessionID string, limits *RateLimitInfo) *SessionTelemetryRecord {
	t.Helper()
	t.Setenv("MOAI_STATUSLINE_CONTEXT_SIZE", "")

	in := StdinData{
		SessionID: sessionID,
		Workspace: &WorkspaceInfo{CurrentDir: projDir, ProjectDir: projDir},
		ContextWindow: &ContextWindowInfo{
			ContextWindowSize: 256000,
			UsedPercentage:    new(90.0),
		},
		RateLimits: limits,
	}
	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	b := &defaultBuilder{renderer: NewRenderer("default", true, nil), mode: ModeDefault}
	if _, err := b.Build(context.Background(), bytes.NewReader(raw)); err != nil {
		t.Fatalf("Build() error: %v", err)
	}
	return readRecord(t, usagePath(projDir, sessionID))
}

// TestQAS_AC001_RecordCarriesSuppliedWindowsOnly — AC-QAS-001 (REQ-QAS-001).
// The record carries exactly the windows the stdin supplied, with exactly their
// percentages and reset times, and no window the stdin did not supply.
func TestQAS_AC001_RecordCarriesSuppliedWindowsOnly(t *testing.T) {
	t.Run("both_windows", func(t *testing.T) {
		rec := qasBuildWithLimits(t, t.TempDir(), "qas-both", &RateLimitInfo{
			FiveHour: qasWindow(62.5, 1790000000),
			SevenDay: qasWindow(41.2, 1790500000),
		})
		if rec.FiveHour == nil || rec.FiveHour.UsedPercentage != 62.5 || rec.FiveHour.ResetsAt != 1790000000 {
			t.Errorf("five-hour window = %+v, want 62.5%% resetting at 1790000000", rec.FiveHour)
		}
		if rec.SevenDay == nil || rec.SevenDay.UsedPercentage != 41.2 || rec.SevenDay.ResetsAt != 1790500000 {
			t.Errorf("seven-day window = %+v, want 41.2%% resetting at 1790500000", rec.SevenDay)
		}
	})

	t.Run("seven_day_only", func(t *testing.T) {
		rec := qasBuildWithLimits(t, t.TempDir(), "qas-seven", &RateLimitInfo{
			SevenDay: qasWindow(41.2, 1790500000),
		})
		if rec.FiveHour != nil {
			t.Errorf("five-hour window = %+v, want absent: the stdin did not supply it", rec.FiveHour)
		}
		if rec.SevenDay == nil || rec.SevenDay.UsedPercentage != 41.2 {
			t.Errorf("seven-day window = %+v, want the supplied 41.2%%", rec.SevenDay)
		}
	})

	t.Run("no_rate_limits", func(t *testing.T) {
		proj := t.TempDir()
		rec := qasBuildWithLimits(t, proj, "qas-none", nil)
		if rec.FiveHour != nil || rec.SevenDay != nil {
			t.Errorf("windows = %+v / %+v, want none: the stdin carried no rate_limits", rec.FiveHour, rec.SevenDay)
		}
		raw, err := os.ReadFile(usagePath(proj, "qas-none"))
		if err != nil {
			t.Fatal(err)
		}
		for _, key := range []string{"five_hour", "seven_day"} {
			if strings.Contains(string(raw), key) {
				t.Errorf("window-less record carries the key %q; an absent window must be omitted", key)
			}
		}
	})
}

// TestQAS_AC002_PreviousSchemaReadsAsNoWindows — AC-QAS-002 (REQ-QAS-002).
// A record written at schema version 1 or 2 decodes without error and reports
// no windows.
func TestQAS_AC002_PreviousSchemaReadsAsNoWindows(t *testing.T) {
	t.Parallel()

	type v1Record struct {
		SchemaVersion     int     `json:"schema_version"`
		SessionID         string  `json:"session_id"`
		WriterPID         int     `json:"writer_pid"`
		CapturedAt        string  `json:"captured_at"`
		ContextWindowSize int     `json:"context_window_size"`
		TokensUsed        int     `json:"tokens_used"`
		RawPct            float64 `json:"raw_pct"`
		Stage             string  `json:"stage"`
		Band              string  `json:"band"`
	}
	type v2Record struct {
		v1Record
		Model  string `json:"model,omitempty"`
		Effort string `json:"effort,omitempty"`
	}
	base := v1Record{
		SessionID: "qas-old", WriterPID: 42, CapturedAt: "2026-08-17T00:00:00Z",
		ContextWindowSize: 1_000_000, TokensUsed: 500_000, RawPct: 50, Stage: "soft", Band: "large",
	}

	for name, rec := range map[string]any{
		"schema_1": func() v1Record { r := base; r.SchemaVersion = 1; return r }(),
		"schema_2": func() v2Record {
			r := v2Record{v1Record: base, Model: "Opus 5", Effort: "high"}
			r.SchemaVersion = 2
			return r
		}(),
	} {
		t.Run(name, func(t *testing.T) {
			data, err := json.MarshalIndent(rec, "", "  ")
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(t.TempDir(), "qas-old.json")
			if err := os.WriteFile(path, data, 0o644); err != nil {
				t.Fatal(err)
			}
			got, err := ReadSessionTelemetry(path)
			if err != nil {
				t.Fatalf("an old-schema record must read without error: %v", err)
			}
			if got.FiveHour != nil || got.SevenDay != nil {
				t.Errorf("windows = %+v / %+v, want none for an old-schema record", got.FiveHour, got.SevenDay)
			}
			if got.RawPct != 50 || got.ContextWindowSize != 1_000_000 {
				t.Errorf("context values lost: %+v", got)
			}
		})
	}
}

// qasNormalizers are the only three values AC-QAS-002b normalizes before it
// compares a window-less record to the baseline golden.
var qasNormalizers = []struct {
	re   *regexp.Regexp
	with string
}{
	{regexp.MustCompile(`"schema_version": \d+`), `"schema_version": N`},
	{regexp.MustCompile(`"writer_pid": \d+`), `"writer_pid": N`},
	{regexp.MustCompile(`"captured_at": "[^"]*"`), `"captured_at": "T"`},
}

func qasNormalize(b []byte) string {
	s := string(b)
	for _, n := range qasNormalizers {
		s = n.re.ReplaceAllString(s, n.with)
	}
	return s
}

// TestQAS_AC002b_WindowlessRecordBytesMatchBaseline — AC-QAS-002 (REQ-QAS-002).
// A window-less record serializes byte-identically to the baseline golden
// measured before any implementation commit (M0), apart from the capture time,
// the writer pid, and the schema version value; the schema version is now 3.
func TestQAS_AC002b_WindowlessRecordBytesMatchBaseline(t *testing.T) {
	t.Parallel()

	proj := t.TempDir()
	// The fixed input of the M0 baseline (progress.md §E.2, golden 1).
	writeContextUsage(proj, "qas-baseline-session", 4242,
		MemoryData{Available: true, ContextWindowSize: 200000, TokensUsed: 50000},
		handoffStageNone, "Opus 4.8", "high")
	got, err := os.ReadFile(usagePath(proj, "qas-baseline-session"))
	if err != nil {
		t.Fatal(err)
	}
	golden, err := os.ReadFile(filepath.Join("testdata", "qas_baseline_windowless_record.golden.json"))
	if err != nil {
		t.Fatalf("baseline golden missing: %v", err)
	}

	if rec := readRecord(t, usagePath(proj, "qas-baseline-session")); rec.SchemaVersion != 3 {
		t.Errorf("schema_version = %d, want 3", rec.SchemaVersion)
	}
	if qasNormalize(got) != qasNormalize(golden) {
		t.Errorf("window-less record differs from the baseline golden beyond the three normalized values\n got: %s\nwant: %s", got, golden)
	}
}

// qasRecord reads the persisted record of one session under proj.
func qasRecord(t *testing.T, proj, sessionID string) *SessionTelemetryRecord {
	t.Helper()
	return readRecord(t, usagePath(proj, sessionID))
}

// TestQAS_AC003_ThrottleBucketHeartbeatAndWindowDrop — AC-QAS-003 (REQ-QAS-003).
// The throttle payload gains each window's presence, truncated percentage, and
// reset time; a window-carrying record also rewrites on a heartbeat; a
// window-less record stays throttled exactly as before.
func TestQAS_AC003_ThrottleBucketHeartbeatAndWindowDrop(t *testing.T) {
	t.Parallel()

	const reset = 1790000000
	five := func(pct float64) *RateLimitInfo {
		return &RateLimitInfo{FiveHour: qasWindow(pct, reset)}
	}
	heartbeat := config.QuotaHeartbeatInterval

	// seed writes the on-disk record carrying a five-hour window at 62.2% and
	// ages its mtime so a skipped write is detectable.
	seed := func(t *testing.T) (proj, path string, old time.Time) {
		t.Helper()
		proj = t.TempDir()
		qasWriteAt(qasT0, proj, "qas-throttle", five(62.2))
		path = usagePath(proj, "qas-throttle")
		return proj, path, qasAgeFile(t, path)
	}

	t.Run("a_same_truncated_bucket_within_heartbeat_writes_nothing", func(t *testing.T) {
		proj, path, old := seed(t)
		before, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		qasWriteAt(qasT0.Add(time.Minute), proj, "qas-throttle", five(62.9))
		after, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		st, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(before, after) || !st.ModTime().Equal(old) {
			t.Errorf("62.9%% shares the truncated bucket 62 with 62.2%% and must not rewrite the record")
		}
	})

	t.Run("b_next_bucket_writes", func(t *testing.T) {
		proj, _, _ := seed(t)
		qasWriteAt(qasT0.Add(time.Minute), proj, "qas-throttle", five(63.0))
		rec := qasRecord(t, proj, "qas-throttle")
		if rec.FiveHour == nil || rec.FiveHour.UsedPercentage != 63.0 {
			t.Errorf("five-hour window = %+v, want 63.0 written (a new truncated bucket)", rec.FiveHour)
		}
	})

	t.Run("b2_changed_reset_time_writes", func(t *testing.T) {
		proj, _, _ := seed(t)
		qasWriteAt(qasT0.Add(time.Minute), proj, "qas-throttle",
			&RateLimitInfo{FiveHour: qasWindow(62.2, reset+3600)})
		rec := qasRecord(t, proj, "qas-throttle")
		if rec.FiveHour == nil || rec.FiveHour.ResetsAt != reset+3600 {
			t.Errorf("five-hour window = %+v, want the new reset time written", rec.FiveHour)
		}
	})

	t.Run("c_unchanged_reading_older_than_heartbeat_rewrites", func(t *testing.T) {
		proj, path, old := seed(t)
		// One second inside the heartbeat: still throttled (positive control).
		qasWriteAt(qasT0.Add(heartbeat-time.Second), proj, "qas-throttle", five(62.2))
		if st, err := os.Stat(path); err != nil || !st.ModTime().Equal(old) {
			t.Fatalf("a reading inside the heartbeat must not rewrite (stat err=%v)", err)
		}
		later := qasT0.Add(heartbeat + time.Second)
		qasWriteAt(later, proj, "qas-throttle", five(62.2))
		rec := qasRecord(t, proj, "qas-throttle")
		if rec.CapturedAt != qasStamp(later) {
			t.Errorf("captured_at = %q, want the refreshed %q", rec.CapturedAt, qasStamp(later))
		}
	})

	t.Run("d_window_disappearing_writes_a_record_without_it", func(t *testing.T) {
		proj, _, _ := seed(t)
		qasWriteAt(qasT0.Add(time.Minute), proj, "qas-throttle", nil)
		rec := qasRecord(t, proj, "qas-throttle")
		if rec.FiveHour != nil || rec.SevenDay != nil {
			t.Errorf("windows = %+v / %+v, want none after the window left stdin", rec.FiveHour, rec.SevenDay)
		}
	})

	t.Run("e_windowless_record_is_not_heartbeat_rewritten", func(t *testing.T) {
		proj := t.TempDir()
		qasWriteAt(qasT0, proj, "qas-plain", nil)
		path := usagePath(proj, "qas-plain")
		old := qasAgeFile(t, path)
		before, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		qasWriteAt(qasT0.Add(3*heartbeat), proj, "qas-plain", nil)
		after, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		st, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(before, after) || !st.ModTime().Equal(old) {
			t.Errorf("a window-less record unchanged beyond the heartbeat must be throttled exactly as before")
		}
	})
}

// TestQAS_AC004_ExhaustedAtStickyUntilRollover — AC-QAS-004 (REQ-QAS-004).
// The first-observed-exhausted time is set when a window is first seen at or
// above the exhaustion percentage, stays while the window's reset time is
// unchanged, and is dropped when the window leaves the record or its reset
// time changes.
func TestQAS_AC004_ExhaustedAtStickyUntilRollover(t *testing.T) {
	t.Parallel()

	const reset, reset2 = int64(1790000000), int64(1790018000)
	heartbeat := config.QuotaHeartbeatInterval
	five := func(pct float64, resetsAt int64) *RateLimitInfo {
		return &RateLimitInfo{FiveHour: qasWindow(pct, resetsAt)}
	}

	t.Run("first_seen_at_100_sets_the_time", func(t *testing.T) {
		proj := t.TempDir()
		qasWriteAt(qasT0, proj, "qas-ex", five(100, reset))
		rec := qasRecord(t, proj, "qas-ex")
		if rec.FiveHour == nil || rec.FiveHour.ExhaustedAt != qasStamp(qasT0) {
			t.Errorf("exhausted time = %+v, want the first capture %q", rec.FiveHour, qasStamp(qasT0))
		}
	})

	t.Run("sticky_across_later_writes", func(t *testing.T) {
		proj := t.TempDir()
		qasWriteAt(qasT0, proj, "qas-ex", five(100, reset))
		// A heartbeat rewrite refreshes the capture time but keeps the exhausted time.
		hb := qasT0.Add(heartbeat + time.Second)
		qasWriteAt(hb, proj, "qas-ex", five(100, reset))
		rec := qasRecord(t, proj, "qas-ex")
		if rec.CapturedAt != qasStamp(hb) {
			t.Fatalf("captured_at = %q, want the heartbeat rewrite %q (the rewrite must have happened)", rec.CapturedAt, qasStamp(hb))
		}
		if rec.FiveHour == nil || rec.FiveHour.ExhaustedAt != qasStamp(qasT0) {
			t.Errorf("exhausted time = %+v, want it unchanged at %q", rec.FiveHour, qasStamp(qasT0))
		}
		// A write driven by the other window's bucket keeps it too.
		both := func(sevenPct float64) *RateLimitInfo {
			return &RateLimitInfo{FiveHour: qasWindow(100, reset), SevenDay: qasWindow(sevenPct, reset2)}
		}
		qasWriteAt(hb.Add(time.Minute), proj, "qas-ex", both(50))
		qasWriteAt(hb.Add(2*time.Minute), proj, "qas-ex", both(51))
		rec = qasRecord(t, proj, "qas-ex")
		if rec.SevenDay == nil || rec.SevenDay.UsedPercentage != 51 {
			t.Fatalf("seven-day window = %+v, want the second bucket written", rec.SevenDay)
		}
		if rec.FiveHour == nil || rec.FiveHour.ExhaustedAt != qasStamp(qasT0) {
			t.Errorf("exhausted time = %+v, want it unchanged at %q", rec.FiveHour, qasStamp(qasT0))
		}
	})

	t.Run("reset_time_change_drops_it", func(t *testing.T) {
		proj := t.TempDir()
		qasWriteAt(qasT0, proj, "qas-ex", five(100, reset))
		qasWriteAt(qasT0.Add(time.Hour), proj, "qas-ex", five(10, reset2))
		rec := qasRecord(t, proj, "qas-ex")
		if rec.FiveHour == nil || rec.FiveHour.ExhaustedAt != "" {
			t.Errorf("window = %+v, want the exhausted time dropped after a rollover to 10%%", rec.FiveHour)
		}
	})

	t.Run("reset_time_change_still_exhausted_reobserves", func(t *testing.T) {
		proj := t.TempDir()
		qasWriteAt(qasT0, proj, "qas-ex", five(100, reset))
		later := qasT0.Add(time.Hour)
		qasWriteAt(later, proj, "qas-ex", five(100, reset2))
		rec := qasRecord(t, proj, "qas-ex")
		if rec.FiveHour == nil || rec.FiveHour.ExhaustedAt != qasStamp(later) {
			t.Errorf("window = %+v, want the exhausted time re-observed from the current capture %q", rec.FiveHour, qasStamp(later))
		}
	})

	t.Run("window_no_longer_carried_drops_it", func(t *testing.T) {
		proj := t.TempDir()
		qasWriteAt(qasT0, proj, "qas-ex", five(100, reset))
		qasWriteAt(qasT0.Add(time.Minute), proj, "qas-ex", &RateLimitInfo{SevenDay: qasWindow(40, reset2)})
		back := qasT0.Add(2 * time.Minute)
		qasWriteAt(back, proj, "qas-ex", five(100, reset))
		rec := qasRecord(t, proj, "qas-ex")
		if rec.FiveHour == nil || rec.FiveHour.ExhaustedAt != qasStamp(back) {
			t.Errorf("window = %+v, want a fresh exhausted time %q: the window left the record in between", rec.FiveHour, qasStamp(back))
		}
	})

	t.Run("first_seen_at_99_carries_no_time", func(t *testing.T) {
		proj := t.TempDir()
		qasWriteAt(qasT0, proj, "qas-ex", five(99, reset))
		rec := qasRecord(t, proj, "qas-ex")
		if rec.FiveHour == nil || rec.FiveHour.ExhaustedAt != "" {
			t.Errorf("window = %+v, want no exhausted time at 99%%", rec.FiveHour)
		}
		raw, err := os.ReadFile(usagePath(proj, "qas-ex"))
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(raw), "exhausted_at") {
			t.Errorf("the key exhausted_at is present at 99%%; an absent time must be omitted")
		}
	})
}
