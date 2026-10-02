// Package statusline tests for SPEC-QUOTA-AWARE-SCHEDULING-001 M2 (card t1347):
// the quota aggregator — freshest record per window (never the maximum), the
// age boundary, the clock-skew guard, a reset window read as reset rather than
// as a stale high, fail-open on absent or unreadable data (REQ-QAS-005,
// REQ-QAS-006), and the offline / spawn-free / read-only property
// (REQ-QAS-007).
package statusline

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// qasMaxAge is the unmeasured default max age (30m) the aggregator tests use.
const qasMaxAge = 30 * time.Minute

// qasFixture writes one schema-3 record into <stateDir>/context-usage, captured
// at the given time, and gives the file that time as its modification time —
// the aggregator reads only files changed within the max age, and a real write
// at capture time is exactly that.
func qasFixture(t *testing.T, stateDir, id string, captured time.Time, five, seven *QuotaWindowRecord) {
	t.Helper()
	rec := &SessionTelemetryRecord{
		SchemaVersion: contextUsageSchemaVersion, SessionID: id, WriterPID: 1,
		CapturedAt: qasStamp(captured), ContextWindowSize: 200_000, TokensUsed: 50_000, RawPct: 25,
		Stage: "none", Band: "standard", FiveHour: five, SevenDay: seven,
	}
	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	qasFixtureBytes(t, stateDir, id, captured, data)
}

// qasFixtureBytes writes raw bytes as a record file with the given mtime.
func qasFixtureBytes(t *testing.T, stateDir, id string, mtime time.Time, data []byte) {
	t.Helper()
	dir := filepath.Join(stateDir, contextUsageDirName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, id+".json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, mtime, mtime); err != nil {
		t.Fatal(err)
	}
}

// qasWin builds a record window resetting `after` after the injected now.
func qasWin(pct float64, after time.Duration) *QuotaWindowRecord {
	return &QuotaWindowRecord{UsedPercentage: pct, ResetsAt: qasT0.Add(after).Unix()}
}

// qasExpect asserts one window reading.
func qasExpect(t *testing.T, name string, got QuotaReading, state QuotaState, pct float64) {
	t.Helper()
	if got.State != state {
		t.Errorf("%s: state = %q, want %q (reading %+v)", name, got.State, state, got)
		return
	}
	if state == QuotaFresh && got.UsedPercentage != pct {
		t.Errorf("%s: used = %v, want %v", name, got.UsedPercentage, pct)
	}
}

// TestQAS_AC005_AggregateFreshestBoundariesSkewAndRollover — AC-QAS-005
// (REQ-QAS-005). The freshest record per window wins, never the maximum and
// never the minimum; the age and skew boundaries are exact; a window whose
// reset time is not after now reads as reset.
func TestQAS_AC005_AggregateFreshestBoundariesSkewAndRollover(t *testing.T) {
	t.Parallel()
	now := qasT0

	run := func(t *testing.T, fill func(stateDir string)) QuotaAggregate {
		t.Helper()
		stateDir := filepath.Join(t.TempDir(), ".moai", "state")
		fill(stateDir)
		return AggregateQuota(stateDir, now, qasMaxAge)
	}

	t.Run("freshest_not_max_five_hour", func(t *testing.T) {
		got := run(t, func(s string) {
			qasFixture(t, s, "older", now.Add(-20*time.Minute), qasWin(95, time.Hour), nil)
			qasFixture(t, s, "newer", now.Add(-5*time.Minute), qasWin(60, time.Hour), nil)
		})
		qasExpect(t, "five_hour", got.FiveHour, QuotaFresh, 60)
	})

	t.Run("freshest_not_max_seven_day", func(t *testing.T) {
		got := run(t, func(s string) {
			qasFixture(t, s, "older", now.Add(-25*time.Minute), nil, qasWin(97, 24*time.Hour))
			qasFixture(t, s, "newer", now.Add(-2*time.Minute), nil, qasWin(70, 24*time.Hour))
		})
		qasExpect(t, "seven_day", got.SevenDay, QuotaFresh, 70)
	})

	t.Run("freshest_not_min_newer_record_higher", func(t *testing.T) {
		// A "minimum over fresh records" aggregator would pick 40 here.
		got := run(t, func(s string) {
			qasFixture(t, s, "older", now.Add(-20*time.Minute), qasWin(40, time.Hour), nil)
			qasFixture(t, s, "newer", now.Add(-5*time.Minute), qasWin(80, time.Hour), nil)
		})
		qasExpect(t, "five_hour", got.FiveHour, QuotaFresh, 80)
	})

	t.Run("per_window", func(t *testing.T) {
		got := run(t, func(s string) {
			qasFixture(t, s, "older", now.Add(-20*time.Minute), qasWin(55, time.Hour), nil)
			qasFixture(t, s, "newer", now.Add(-time.Minute), nil, qasWin(30, 24*time.Hour))
		})
		qasExpect(t, "five_hour", got.FiveHour, QuotaFresh, 55)
		qasExpect(t, "seven_day", got.SevenDay, QuotaFresh, 30)
	})

	t.Run("age_exactly_max", func(t *testing.T) {
		got := run(t, func(s string) {
			qasFixture(t, s, "edge", now.Add(-qasMaxAge), qasWin(70, time.Hour), nil)
		})
		qasExpect(t, "five_hour", got.FiveHour, QuotaFresh, 70)
	})

	t.Run("age_max_plus_1s", func(t *testing.T) {
		got := run(t, func(s string) {
			qasFixture(t, s, "edge", now.Add(-qasMaxAge-time.Second), qasWin(99, time.Hour), nil)
		})
		qasExpect(t, "five_hour", got.FiveHour, QuotaUnknown, 0)
	})

	t.Run("future_capture_within_tolerance", func(t *testing.T) {
		got := run(t, func(s string) {
			qasFixture(t, s, "skewed", now.Add(config.QuotaClockSkewTolerance), qasWin(70, time.Hour), nil)
		})
		qasExpect(t, "five_hour", got.FiveHour, QuotaFresh, 70)
	})

	t.Run("future_capture_beyond_tolerance", func(t *testing.T) {
		got := run(t, func(s string) {
			qasFixture(t, s, "skewed", now.Add(config.QuotaClockSkewTolerance+time.Second), qasWin(99, time.Hour), nil)
		})
		qasExpect(t, "five_hour", got.FiveHour, QuotaUnknown, 0)
	})

	t.Run("future_record_cannot_win_over_a_real_one_beyond_tolerance", func(t *testing.T) {
		got := run(t, func(s string) {
			qasFixture(t, s, "real", now.Add(-3*time.Minute), qasWin(50, time.Hour), nil)
			qasFixture(t, s, "skewed", now.Add(config.QuotaClockSkewTolerance+time.Minute), qasWin(99, time.Hour), nil)
		})
		qasExpect(t, "five_hour", got.FiveHour, QuotaFresh, 50)
	})

	t.Run("reset_window_is_reset", func(t *testing.T) {
		got := run(t, func(s string) {
			qasFixture(t, s, "rolled", now.Add(-time.Minute), qasWin(99, -time.Second), nil)
		})
		qasExpect(t, "five_hour", got.FiveHour, QuotaReset, 0)
	})

	t.Run("reset_exactly_now_is_reset", func(t *testing.T) {
		got := run(t, func(s string) {
			qasFixture(t, s, "rolled", now.Add(-time.Minute), qasWin(99, 0), nil)
		})
		qasExpect(t, "five_hour", got.FiveHour, QuotaReset, 0)
	})

	t.Run("reset_one_second_ahead_is_fresh", func(t *testing.T) {
		got := run(t, func(s string) {
			qasFixture(t, s, "live", now.Add(-time.Minute), qasWin(99, time.Second), nil)
		})
		qasExpect(t, "five_hour", got.FiveHour, QuotaFresh, 99)
	})

	t.Run("source_times_and_exhausted_surfaced", func(t *testing.T) {
		captured := now.Add(-4 * time.Minute)
		win := qasWin(100, time.Hour)
		win.ExhaustedAt = qasStamp(now.Add(-10 * time.Minute))
		got := run(t, func(s string) { qasFixture(t, s, "full", captured, win, nil) })
		qasExpect(t, "five_hour", got.FiveHour, QuotaFresh, 100)
		if !got.FiveHour.CapturedAt.Equal(captured) {
			t.Errorf("captured at = %v, want the source capture time %v", got.FiveHour.CapturedAt, captured)
		}
		if got.FiveHour.ResetsAt != win.ResetsAt {
			t.Errorf("resets at = %d, want %d", got.FiveHour.ResetsAt, win.ResetsAt)
		}
		if want := now.Add(-10 * time.Minute); !got.FiveHour.ExhaustedAt.Equal(want) {
			t.Errorf("exhausted at = %v, want %v", got.FiveHour.ExhaustedAt, want)
		}
		if got.SevenDay.State != QuotaUnknown {
			t.Errorf("seven-day state = %q, want unknown: no record carries it", got.SevenDay.State)
		}
	})
}

// TestQAS_AC006_FailOpenOnAbsentOrUnreadable — AC-QAS-006 (REQ-QAS-006),
// statusline half. Absent, stale, or unparseable data reads as unknown, and an
// unparseable configuration leaves the aggregator on the default max age with
// the gate off.
func TestQAS_AC006_FailOpenOnAbsentOrUnreadable(t *testing.T) {
	t.Parallel()
	now := qasT0
	bothUnknown := func(t *testing.T, got QuotaAggregate) {
		t.Helper()
		qasExpect(t, "five_hour", got.FiveHour, QuotaUnknown, 0)
		qasExpect(t, "seven_day", got.SevenDay, QuotaUnknown, 0)
	}

	t.Run("dir_missing", func(t *testing.T) {
		bothUnknown(t, AggregateQuota(filepath.Join(t.TempDir(), "no-such", "state"), now, qasMaxAge))
	})

	t.Run("unparseable_record", func(t *testing.T) {
		stateDir := filepath.Join(t.TempDir(), "state")
		qasFixtureBytes(t, stateDir, "broken", now.Add(-time.Minute), []byte("{not json"))
		bothUnknown(t, AggregateQuota(stateDir, now, qasMaxAge))
	})

	t.Run("unparseable_record_is_skipped_beside_a_valid_one", func(t *testing.T) {
		stateDir := filepath.Join(t.TempDir(), "state")
		qasFixtureBytes(t, stateDir, "broken", now.Add(-time.Minute), []byte("{not json"))
		qasFixture(t, stateDir, "good", now.Add(-2*time.Minute), qasWin(64, time.Hour), nil)
		qasExpect(t, "five_hour", AggregateQuota(stateDir, now, qasMaxAge).FiveHour, QuotaFresh, 64)
	})

	t.Run("only_stale_records", func(t *testing.T) {
		stateDir := filepath.Join(t.TempDir(), "state")
		qasFixture(t, stateDir, "a", now.Add(-time.Hour), qasWin(99, 2*time.Hour), qasWin(99, 48*time.Hour))
		qasFixture(t, stateDir, "b", now.Add(-3*time.Hour), qasWin(98, 2*time.Hour), nil)
		bothUnknown(t, AggregateQuota(stateDir, now, qasMaxAge))
	})

	t.Run("unparseable_config_uses_default_max_age", func(t *testing.T) {
		root := t.TempDir()
		sections := filepath.Join(root, ".moai", "config", "sections")
		if err := os.MkdirAll(sections, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(sections, "workflow.yaml"), []byte("workflow: [\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		gate := config.LoadQuotaGate(root)
		if gate.Enabled || gate.MaxAge != qasMaxAge {
			t.Fatalf("gate = %+v, want the shipped defaults (off, max age %s)", gate, qasMaxAge)
		}
		stateDir := filepath.Join(root, ".moai", "state")
		qasFixture(t, stateDir, "inside", now.Add(-29*time.Minute), qasWin(70, time.Hour), nil)
		qasFixture(t, stateDir, "outside", now.Add(-31*time.Minute), nil, qasWin(70, 24*time.Hour))
		got := AggregateQuota(stateDir, now, gate.MaxAge)
		qasExpect(t, "five_hour", got.FiveHour, QuotaFresh, 70)
		qasExpect(t, "seven_day", got.SevenDay, QuotaUnknown, 0)
	})
}

// qasTreeDigest maps every file under dir to its size and content digest, so a
// before / after comparison catches a created, removed, or modified file.
func qasTreeDigest(t *testing.T, dir string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(dir, path)
		if d.IsDir() {
			out[rel] = "dir"
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		out[rel] = fmt.Sprintf("%d:%x", len(data), sha256.Sum256(data))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// TestQAS_AC014_AggregatorIsOfflineSpawnFreeAndReadOnly — AC-QAS-014
// (REQ-QAS-007), statusline half. The non-test files matching quota*.go import
// neither net, net/http, nor os/exec, and an aggregator run leaves the record
// directory byte-identical. The file-count floor here is the one this package
// can meet at M2 (one file); the three-file floor the AC states spans
// internal/statusline and internal/cli and is evaluated at M5, once
// internal/cli/factory_quota*.go exist (plan debt N1).
func TestQAS_AC014_AggregatorIsOfflineSpawnFreeAndReadOnly(t *testing.T) {
	t.Parallel()

	matches, err := filepath.Glob("quota*.go")
	if err != nil {
		t.Fatal(err)
	}
	var swept []string
	for _, m := range matches {
		if !strings.HasSuffix(m, "_test.go") {
			swept = append(swept, m)
		}
	}
	if len(swept) < 1 {
		t.Fatalf("swept %d quota*.go files, want at least 1 — a vanished or renamed file must not shrink the sweep silently", len(swept))
	}
	forbidden := map[string]bool{"net": true, "net/http": true, "os/exec": true}
	for _, file := range swept {
		f, err := parser.ParseFile(token.NewFileSet(), file, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse %s: %v", file, err)
		}
		for _, imp := range f.Imports {
			if path, _ := strconv.Unquote(imp.Path.Value); forbidden[path] {
				t.Errorf("%s imports %q: the quota path must be offline and spawn-free", file, path)
			}
		}
	}

	stateDir := filepath.Join(t.TempDir(), "state")
	qasFixture(t, stateDir, "a", qasT0.Add(-time.Minute), qasWin(91, time.Hour), qasWin(40, 24*time.Hour))
	qasFixture(t, stateDir, "b", qasT0.Add(-40*time.Minute), qasWin(20, time.Hour), nil)
	qasFixtureBytes(t, stateDir, "broken", qasT0.Add(-time.Minute), []byte("{nope"))
	before := qasTreeDigest(t, stateDir)
	got := AggregateQuota(stateDir, qasT0, qasMaxAge)
	qasExpect(t, "five_hour", got.FiveHour, QuotaFresh, 91)
	if after := qasTreeDigest(t, stateDir); !reflect.DeepEqual(before, after) {
		t.Errorf("the aggregator changed the record directory\nbefore: %v\n after: %v", before, after)
	}
}
