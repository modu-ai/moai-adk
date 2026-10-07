// factory_quota_test.go — SPEC-QUOTA-AWARE-SCHEDULING-001 M3 AC tests
// (card t1347): the quota lane gate — a Claude lane near its quota limit
// leases no NEW card, prints one hold line, and exits with the no-card status;
// the --wait latch; the non-Claude and non-leasing-verb exemptions; the shared
// pressure evaluation. AC-QAS-006b, -008, -008b, -009, -010, -011, -011b,
// -014 (cli half), -017 (function part).
//
// Every fixture is built under t.TempDir() with an isolated git config and
// MOAI_HOME sandboxed away (the factory test family's convention); ./internal/cli
// runs only through the anchored -run selectors naming one of these tests. Every
// environment variable the lane predicates read is set explicitly per test, so
// the lane environment this suite may run in cannot leak into a verdict.
package cli

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/statusline"
)

// The fixtures' reset instants: the five-hour window resets three hours after
// the fixture clock, the seven-day window three days after it.
var (
	qasReset5 = fcNow.Add(3 * time.Hour).Unix()
	qasReset7 = fcNow.Add(72 * time.Hour).Unix()
)

// qasHoldSegment is one window's segment of the hold line; qasHoldLineRE is the
// whole line: the prefix once, then one segment per held window.
const qasHoldSegment = `[a-z_]+ used=[0-9]+\.[0-9]% resets_at=[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z`

var qasHoldLineRE = regexp.MustCompile(`^quota hold: ` + qasHoldSegment + `(; ` + qasHoldSegment + `)*$`)

// qasWin builds one window record.
func qasWin(used float64, reset int64) *statusline.QuotaWindowRecord {
	return &statusline.QuotaWindowRecord{UsedPercentage: used, ResetsAt: reset}
}

// qasEnableGate writes a workflow.yaml that enables the quota gate with the
// shipped numeric defaults (90 / 95 / 5 / 30m).
func qasEnableGate(t *testing.T, root string) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatal(err)
	}
	body := "workflow:\n    quota_gate:\n        enabled: true\n"
	if err := os.WriteFile(filepath.Join(dir, "workflow.yaml"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

// qasWriteRecord writes one session telemetry record under the project's
// context-usage directory, stamping its modification time with the capture
// time so the aggregator's file-age pre-filter agrees with the record.
func qasWriteRecord(t *testing.T, root, sid string, captured time.Time, five, seven *statusline.QuotaWindowRecord) {
	t.Helper()
	rec := statusline.SessionTelemetryRecord{
		SchemaVersion: 3, SessionID: sid, WriterPID: 1,
		CapturedAt: captured.UTC().Format(time.RFC3339Nano),
		FiveHour:   five, SevenDay: seven,
	}
	raw, err := json.Marshal(rec)
	if err != nil {
		t.Fatal(err)
	}
	path := statusline.SessionTelemetryPath(filepath.Join(root, ".moai", "state"), sid)
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, captured, captured); err != nil {
		t.Fatal(err)
	}
}

// qasLaneEnv stamps the full lane environment with one backend token carried
// by BOTH the launch-provider variable and the kanban backend variable, so a
// test names the lane's backend once.
func qasLaneEnv(t *testing.T, label, backend string) {
	t.Helper()
	sdLaneEnv(t, label, backend)
	t.Setenv(config.EnvMoaiLaunchProvider, backend)
}

// qasRunNext runs `moai factory next` the way the real root runs it: the root
// (fang) owns error printing and silences cobra, and an exit-coded error is
// never printed, so stderr carries only what the verb itself wrote.
func qasRunNext(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	cmd := newFactoryCommand()
	var silence func(c *cobra.Command)
	silence = func(c *cobra.Command) {
		c.SilenceUsage, c.SilenceErrors = true, true
		for _, sub := range c.Commands() {
			silence(sub)
		}
	}
	silence(cmd)
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs(append([]string{"next"}, args...))
	err := cmd.Execute()
	return out.String(), errBuf.String(), err
}

// qasFixtureOpts shapes the held-lane fixture.
type qasFixtureOpts struct {
	five, seven *statusline.QuotaWindowRecord // the fresh record's windows; nil = absent
	noRecord    bool                          // write no record at all
	gateOff     bool                          // write no workflow.yaml (the shipped default)
	assigned    bool                          // place t4 assigned to lane-1 (arm (a))
}

// qasFixture builds the AC-QAS-008 state: a queued card (t1), a picked card
// with a record row and no owner (t2), a picked card with no record row (t3),
// and a Claude lane, with the gate enabled and a fresh record. Every card is
// parallelizable so the serial slot never interferes with arm ordering.
func qasFixture(t *testing.T, o qasFixtureOpts) (string, *factory.BacklogStore) {
	t.Helper()
	sdClearLaneEnv(t)
	root, store := sdMoaiFixture(t)
	fcQueue(t, store, factory.BacklogStateQueued, factory.BacklogStatePicked, factory.BacklogStatePicked, factory.BacklogStatePicked)
	for _, id := range []string{"t1", "t2", "t3", "t4"} {
		fcClassify(t, store, id, factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
	}
	sdRegisterLane(t, root, "lane-1")
	rows := []homestate.Card{{CardID: "t2", State: homestate.CardPicked}}
	if o.assigned {
		rows = append(rows, homestate.Card{CardID: "t4", State: homestate.CardAssigned, OwnerLabel: "lane-1", Stage: homestate.CardRun})
	}
	fcPlace(t, root, rows...)
	if !o.gateOff {
		qasEnableGate(t, root)
	}
	if !o.noRecord {
		qasWriteRecord(t, root, "sess-live", fcNow, o.five, o.seven)
	}
	qasLaneEnv(t, "lane-1", factory.BackendClaude)
	t.Chdir(root)
	return root, store
}

// qasRecordDump renders every card row and the event count — what "the factory
// record is unchanged" compares.
func qasRecordDump(t *testing.T, root string) string {
	t.Helper()
	db := fcOpen(t, root)
	defer func() { _ = db.Close() }()
	rows, err := db.DB.Query(`SELECT run_id,card_id,state,version,owner_label,lease_holder FROM cards ORDER BY run_id,card_id`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	var b strings.Builder
	for rows.Next() {
		var run, card, state, owner, holder string
		var version int
		if err := rows.Scan(&run, &card, &state, &version, &owner, &holder); err != nil {
			t.Fatal(err)
		}
		fmt.Fprintf(&b, "%s|%s|%s|%d|%s|%s\n", run, card, state, version, owner, holder)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	var events int
	if err := db.DB.QueryRow(`SELECT count(*) FROM events`).Scan(&events); err != nil {
		t.Fatal(err)
	}
	fmt.Fprintf(&b, "events=%d\n", events)
	return b.String()
}

// qasAssertHeld asserts the held outcome of one `next` run: standard output
// empty, exactly one stderr line that is a hold line, status 3.
func qasAssertHeld(t *testing.T, out, stderr string, err error) string {
	t.Helper()
	sdExit3(t, "held next", err)
	if out != "" {
		t.Errorf("held next wrote standard output %q, want nothing", out)
	}
	lines := strings.Split(strings.TrimRight(stderr, "\n"), "\n")
	if len(lines) != 1 || !qasHoldLineRE.MatchString(lines[0]) {
		t.Fatalf("held next stderr = %q, want exactly one quota hold line", stderr)
	}
	return lines[0]
}

// qasAssertNotHeld asserts no hold line was written anywhere.
func qasAssertNotHeld(t *testing.T, out, stderr string) {
	t.Helper()
	if strings.Contains(out, "quota hold") || strings.Contains(stderr, "quota hold") {
		t.Errorf("a hold line appeared although the lane must not be held: stdout=%q stderr=%q", out, stderr)
	}
}

// qasFakeClock replaces the verb's clock and wait sleep: every sleep advances
// the clock by advance(n) (n counts sleeps from 1) and then runs onSleep(n).
// It returns the sleep counter.
func qasFakeClock(t *testing.T, advance func(n int) time.Duration, onSleep func(n int)) *int {
	t.Helper()
	prevClock, prevSleep := factoryCardNow, factoryNextWaitSleep
	clock := fcNow
	sleeps := new(int)
	factoryCardNow = func() time.Time { return clock }
	factoryNextWaitSleep = func(d time.Duration) {
		*sleeps++
		step := d
		if advance != nil {
			step = advance(*sleeps)
		}
		clock = clock.Add(step)
		if onSleep != nil {
			onSleep(*sleeps)
		}
	}
	t.Cleanup(func() { factoryCardNow, factoryNextWaitSleep = prevClock, prevSleep })
	return sleeps
}

// qasInjectReadings replaces the aggregator seam with a reader that reports the
// five-hour window at the value readings picks, fresh and resetting in the
// future; the seven-day window is unknown.
func qasInjectReadings(t *testing.T, readingAt func() float64) {
	t.Helper()
	prev := factoryQuotaAggregate
	factoryQuotaAggregate = func(string, time.Time, time.Duration) statusline.QuotaAggregate {
		return statusline.QuotaAggregate{
			FiveHour: statusline.QuotaReading{State: statusline.QuotaFresh, UsedPercentage: readingAt(), ResetsAt: qasReset5, CapturedAt: fcNow},
			SevenDay: statusline.QuotaReading{State: statusline.QuotaUnknown},
		}
	}
	t.Cleanup(func() { factoryQuotaAggregate = prev })
}

// qasInjectBothWindows is qasInjectReadings for two windows: both are fresh and
// resetting in the future, at the values the two functions pick.
func qasInjectBothWindows(t *testing.T, fiveAt, sevenAt func() float64) {
	t.Helper()
	prev := factoryQuotaAggregate
	factoryQuotaAggregate = func(string, time.Time, time.Duration) statusline.QuotaAggregate {
		return statusline.QuotaAggregate{
			FiveHour: statusline.QuotaReading{State: statusline.QuotaFresh, UsedPercentage: fiveAt(), ResetsAt: qasReset5, CapturedAt: fcNow},
			SevenDay: statusline.QuotaReading{State: statusline.QuotaFresh, UsedPercentage: sevenAt(), ResetsAt: qasReset7, CapturedAt: fcNow},
		}
	}
	t.Cleanup(func() { factoryQuotaAggregate = prev })
}

// AC-QAS-006b — absent or unreadable data fails open: the lane leases exactly
// as with the gate disabled in each of four fixtures (the cli half of AC-006).
func TestQAS_AC006b_NextLeasesWhenQuotaDataAbsent(t *testing.T) {
	cases := []struct {
		name  string
		setup func(t *testing.T, root string)
	}{
		{"missing_record_directory", func(t *testing.T, root string) { qasEnableGate(t, root) }},
		{"unparseable_record_file", func(t *testing.T, root string) {
			qasEnableGate(t, root)
			dir := filepath.Join(root, ".moai", "state", "context-usage")
			if err := os.MkdirAll(dir, 0o750); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "garbled.json"), []byte("{not json"), 0o600); err != nil {
				t.Fatal(err)
			}
		}},
		{"only_stale_records", func(t *testing.T, root string) {
			qasEnableGate(t, root)
			qasWriteRecord(t, root, "sess-old", fcNow.Add(-31*time.Minute), qasWin(99, qasReset5), nil)
		}},
		{"unparseable_configuration", func(t *testing.T, root string) {
			dir := filepath.Join(root, ".moai", "config", "sections")
			if err := os.MkdirAll(dir, 0o750); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "workflow.yaml"), []byte("workflow: [unclosed"), 0o600); err != nil {
				t.Fatal(err)
			}
			qasWriteRecord(t, root, "sess-live", fcNow, qasWin(99, qasReset5), nil)
			if got, want := config.LoadQuotaGate(root), config.DefaultQuotaGate(); got != want {
				t.Fatalf("an unparseable workflow.yaml resolved to %+v, want the defaults %+v", got, want)
			}
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			sdClearLaneEnv(t)
			root, store := sdMoaiFixture(t)
			fcQueue(t, store, factory.BacklogStateQueued)
			sdRegisterLane(t, root, "lane-1")
			c.setup(t, root)
			qasLaneEnv(t, "lane-1", factory.BackendClaude)
			t.Chdir(root)
			out, stderr, err := qasRunNext(t, "--run", fcRun)
			if err != nil {
				t.Fatalf("next: %v (stderr %q)", err, stderr)
			}
			qasAssertNotHeld(t, out, stderr)
			if cd := fcCard(t, root, "t1"); cd.State != homestate.CardLeased || cd.LeaseHolder != "lane-1" {
				t.Fatalf("t1 = %s holder=%q, want leased/lane-1", cd.State, cd.LeaseHolder)
			}
		})
	}
}

// AC-QAS-008 — a held Claude lane leases no new card but still receives its
// own assigned card.
func TestQAS_AC008_ClaudeLaneHeldAtThreshold(t *testing.T) {
	heldRun := func(t *testing.T, five, seven *statusline.QuotaWindowRecord) {
		t.Helper()
		root, store := qasFixture(t, qasFixtureOpts{five: five, seven: seven})
		recordBefore, queueBefore := qasRecordDump(t, root), sdQueueBytes(t, store)
		out, stderr, err := qasRunNext(t, "--run", fcRun)
		qasAssertHeld(t, out, stderr, err)
		if got := qasRecordDump(t, root); got != recordBefore {
			t.Errorf("factory record changed under a hold:\nbefore:\n%s\nafter:\n%s", recordBefore, got)
		}
		if got := sdQueueBytes(t, store); got != queueBefore {
			t.Errorf("queue changed under a hold")
		}
		if fcHasCard(t, root, "t1") || fcHasCard(t, root, "t3") {
			t.Errorf("a card gained a record row under a hold (queued-promotion or picked-no-row arm ran)")
		}
	}
	leasedRun := func(t *testing.T, five, seven *statusline.QuotaWindowRecord) {
		t.Helper()
		root, _ := qasFixture(t, qasFixtureOpts{five: five, seven: seven})
		out, stderr, err := qasRunNext(t, "--run", fcRun)
		if err != nil {
			t.Fatalf("next: %v (stderr %q)", err, stderr)
		}
		qasAssertNotHeld(t, out, stderr)
		if cd := fcCard(t, root, "t2"); cd.State != homestate.CardLeased || cd.LeaseHolder != "lane-1" {
			t.Fatalf("t2 = %s holder=%q, want leased/lane-1", cd.State, cd.LeaseHolder)
		}
	}

	t.Run("five_hour_92_new_cards_skipped", func(t *testing.T) { heldRun(t, qasWin(92, qasReset5), nil) })
	t.Run("five_hour_92_assigned_card_still_leased", func(t *testing.T) {
		root, _ := qasFixture(t, qasFixtureOpts{five: qasWin(92, qasReset5), assigned: true})
		out, stderr, err := qasRunNext(t, "--run", fcRun)
		if err != nil {
			t.Fatalf("next: %v (stderr %q)", err, stderr)
		}
		qasAssertNotHeld(t, out, stderr)
		sdAssertLeasedOutput(t, out, "t4", homestate.CardRun, "t4")
		if cd := fcCard(t, root, "t4"); cd.State != homestate.CardLeased || cd.LeaseHolder != "lane-1" {
			t.Fatalf("t4 = %s holder=%q, want leased/lane-1", cd.State, cd.LeaseHolder)
		}
		if cd := fcCard(t, root, "t2"); cd.State != homestate.CardPicked {
			t.Errorf("t2 = %s, want still picked (only the lane's own assigned card is exempt)", cd.State)
		}
	})
	t.Run("five_hour_exactly_90_holds", func(t *testing.T) { heldRun(t, qasWin(90.0, qasReset5), nil) })
	t.Run("five_hour_89_9_leases", func(t *testing.T) { leasedRun(t, qasWin(89.9, qasReset5), nil) })
	t.Run("seven_day_exactly_95_holds", func(t *testing.T) { heldRun(t, nil, qasWin(95.0, qasReset7)) })
	t.Run("seven_day_94_9_leases", func(t *testing.T) { leasedRun(t, nil, qasWin(94.9, qasReset7)) })
}

// AC-QAS-008b — the factory_next MCP form: a hold is the whole non-error
// result text and never "no card is available".
func TestQAS_AC008b_MCPFactoryNextHeld(t *testing.T) {
	heldText := func(t *testing.T) (string, error) {
		t.Helper()
		root, _ := qasFixture(t, qasFixtureOpts{five: qasWin(92, qasReset5)})
		return sdCallTool(t, handleFactoryNext, map[string]any{"project_root": root, "run": fcRun})
	}
	t.Run("hold_text_is_the_hold_line", func(t *testing.T) {
		text, _ := heldText(t)
		if strings.Contains(text, "\n") || !qasHoldLineRE.MatchString(text) {
			t.Errorf("held result text = %q, want exactly one quota hold line", text)
		}
	})
	t.Run("not_an_error_result", func(t *testing.T) {
		if _, err := heldText(t); err != nil {
			t.Errorf("a hold is an error result: %v", err)
		}
	})
	t.Run("no_no_card_text", func(t *testing.T) {
		text, _ := heldText(t)
		if strings.Contains(text, "no card is available") {
			t.Errorf("a hold returned the empty-queue text: %q", text)
		}
	})
	t.Run("assigned_card_unchanged", func(t *testing.T) {
		held := func(gateOff bool) string {
			root, _ := qasFixture(t, qasFixtureOpts{five: qasWin(92, qasReset5), assigned: true, gateOff: gateOff})
			text, err := sdCallTool(t, handleFactoryNext, map[string]any{"project_root": root, "run": fcRun})
			if err != nil {
				t.Fatalf("factory_next: %v", err)
			}
			return text
		}
		withGate, withoutGate := held(false), held(true)
		if withGate != withoutGate {
			t.Errorf("assigned-card result under a hold = %q, want the gate-disabled text %q", withGate, withoutGate)
		}
		if !strings.HasPrefix(withGate, "t4 stage=") {
			t.Errorf("assigned-card result = %q, want the leased-card text", withGate)
		}
	})
}

// AC-QAS-009 — the hold line carries each held window's used percentage and
// reset instant (RFC 3339 UTC); a true empty queue stays distinguishable.
func TestQAS_AC009_HoldLineCarriesResetTime(t *testing.T) {
	utc := func(epoch int64) string { return time.Unix(epoch, 0).UTC().Format(time.RFC3339) }

	t.Run("one_window", func(t *testing.T) {
		qasFixture(t, qasFixtureOpts{five: qasWin(92, qasReset5)})
		out, stderr, err := qasRunNext(t, "--run", fcRun)
		line := qasAssertHeld(t, out, stderr, err)
		if want := "quota hold: five_hour used=92.0% resets_at=" + utc(qasReset5); line != want {
			t.Errorf("hold line = %q, want %q", line, want)
		}
	})
	t.Run("both_windows_one_segment_each", func(t *testing.T) {
		qasFixture(t, qasFixtureOpts{five: qasWin(92, qasReset5), seven: qasWin(96, qasReset7)})
		out, stderr, err := qasRunNext(t, "--run", fcRun)
		line := qasAssertHeld(t, out, stderr, err)
		want := "quota hold: five_hour used=92.0% resets_at=" + utc(qasReset5) + "; seven_day used=96.0% resets_at=" + utc(qasReset7)
		if line != want {
			t.Errorf("hold line = %q, want %q", line, want)
		}
	})
	t.Run("true_empty_queue_is_not_a_hold", func(t *testing.T) {
		sdClearLaneEnv(t)
		root, _ := sdMoaiFixture(t)
		sdRegisterLane(t, root, "lane-1")
		qasEnableGate(t, root)
		qasLaneEnv(t, "lane-1", factory.BackendClaude)
		t.Chdir(root)
		out, stderr, err := qasRunNext(t, "--run", fcRun)
		sdExit3(t, "empty queue", err)
		if !strings.Contains(out, "no card is available") {
			t.Errorf("stdout = %q, want the no-card line", out)
		}
		qasAssertNotHeld(t, out, stderr)
	})
}

// AC-QAS-010 — the wait latch releases only below the margin, at the reset, or
// when the reading ages out.
func TestQAS_AC010_WaitLatchReleasesOnlyBelowMarginOrResetOrUnknown(t *testing.T) {
	// onlyQueued builds a Claude lane with the gate enabled and exactly one
	// queued card (t1), so a release leases it.
	onlyQueued := func(t *testing.T) (string, *factory.BacklogStore) {
		t.Helper()
		sdClearLaneEnv(t)
		root, store := sdMoaiFixture(t)
		fcQueue(t, store, factory.BacklogStateQueued)
		sdRegisterLane(t, root, "lane-1")
		qasEnableGate(t, root)
		qasLaneEnv(t, "lane-1", factory.BackendClaude)
		t.Chdir(root)
		return root, store
	}
	assertLeased := func(t *testing.T, root string) {
		t.Helper()
		if cd := fcCard(t, root, "t1"); cd.State != homestate.CardLeased || cd.LeaseHolder != "lane-1" {
			t.Fatalf("t1 = %s holder=%q, want leased/lane-1", cd.State, cd.LeaseHolder)
		}
	}

	t.Run("margin_release", func(t *testing.T) {
		root, _ := onlyQueued(t)
		readings := []float64{92, 88, 85.0, 84.9}
		sleeps := qasFakeClock(t, nil, nil)
		qasInjectReadings(t, func() float64 { return readings[min(*sleeps, len(readings)-1)] })
		out, stderr, err := qasRunNext(t, "--wait", "--wait-bound", "1h", "--run", fcRun)
		if err != nil {
			t.Fatalf("next --wait: %v (stderr %q)", err, stderr)
		}
		qasAssertNotHeld(t, out, stderr)
		if *sleeps != 3 {
			t.Errorf("wait slept %d times, want 3 (held at 92, 88 and 85.0; released at 84.9)", *sleeps)
		}
		assertLeased(t, root)
	})
	t.Run("exactly_85_stays_held", func(t *testing.T) {
		root, _ := onlyQueued(t)
		readings := []float64{92, 88, 85.0}
		sleeps := qasFakeClock(t, nil, nil)
		qasInjectReadings(t, func() float64 { return readings[min(*sleeps, len(readings)-1)] })
		out, stderr, err := qasRunNext(t, "--wait", "--wait-bound", "20s", "--run", fcRun)
		line := qasAssertHeld(t, out, stderr, err)
		if !strings.Contains(line, "used=85.0%") {
			t.Errorf("hold line %q, want the 85.0%% reading still held", line)
		}
		if fcHasCard(t, root, "t1") {
			t.Errorf("t1 was leased although 85.0 is not below hold minus margin")
		}
	})
	// Plan debt N8: with both windows held, release means ALL held windows
	// released. The five-hour window falls below its release point first (84, below
	// 90 - 5); the seven-day window (hold 95, margin 5) stays held at exactly 90.0
	// and releases only at 89.9.
	t.Run("both_windows_release_only_when_all_released", func(t *testing.T) {
		root, _ := onlyQueued(t)
		five := []float64{92, 84, 84, 84}
		seven := []float64{96, 96, 90.0, 89.9}
		sleeps := qasFakeClock(t, nil, nil)
		qasInjectBothWindows(t,
			func() float64 { return five[min(*sleeps, len(five)-1)] },
			func() float64 { return seven[min(*sleeps, len(seven)-1)] })
		out, stderr, err := qasRunNext(t, "--wait", "--wait-bound", "1h", "--run", fcRun)
		if err != nil {
			t.Fatalf("next --wait: %v (stderr %q)", err, stderr)
		}
		qasAssertNotHeld(t, out, stderr)
		if *sleeps != 3 {
			t.Errorf("wait slept %d times, want 3 (the seven-day window held through 96 and 90.0 after the five-hour window released)", *sleeps)
		}
		assertLeased(t, root)
	})
	t.Run("reset_release", func(t *testing.T) {
		root, _ := onlyQueued(t)
		qasWriteRecord(t, root, "sess-live", fcNow, qasWin(92, fcNow.Add(8*time.Second).Unix()), nil)
		sleeps := qasFakeClock(t, nil, nil)
		out, stderr, err := qasRunNext(t, "--wait", "--wait-bound", "1h", "--run", fcRun)
		if err != nil {
			t.Fatalf("next --wait: %v (stderr %q)", err, stderr)
		}
		qasAssertNotHeld(t, out, stderr)
		if *sleeps != 2 {
			t.Errorf("wait slept %d times, want 2 (held at t+0 and t+5s; the window resets at t+8s)", *sleeps)
		}
		assertLeased(t, root)
	})
	t.Run("unknown_mid_wait_releases", func(t *testing.T) {
		root, _ := onlyQueued(t)
		qasWriteRecord(t, root, "sess-live", fcNow.Add(-20*time.Minute), qasWin(92, qasReset5), nil)
		// One sleep carries the clock from a 20m-old record to a 30m1s-old one
		// while the reset time is still in the future.
		sleeps := qasFakeClock(t, func(int) time.Duration { return 10*time.Minute + time.Second }, nil)
		out, stderr, err := qasRunNext(t, "--wait", "--wait-bound", "1h", "--run", fcRun)
		if err != nil {
			t.Fatalf("next --wait: %v (stderr %q)", err, stderr)
		}
		qasAssertNotHeld(t, out, stderr)
		if *sleeps != 1 {
			t.Errorf("wait slept %d times, want 1 (held while fresh, released once the reading aged out)", *sleeps)
		}
		assertLeased(t, root)
	})
	t.Run("bound_elapsed", func(t *testing.T) {
		root, _ := onlyQueued(t)
		qasWriteRecord(t, root, "sess-live", fcNow, qasWin(92, qasReset5), nil)
		sleeps := qasFakeClock(t, nil, nil)
		out, stderr, err := qasRunNext(t, "--wait", "--wait-bound", "10s", "--run", fcRun)
		qasAssertHeld(t, out, stderr, err)
		if *sleeps != 2 {
			t.Errorf("wait slept %d times before the 10s bound with a 5s interval, want 2", *sleeps)
		}
		if fcHasCard(t, root, "t1") {
			t.Errorf("t1 was leased by a lane that stayed held")
		}
	})
	t.Run("assigned_card_leased_during_wait", func(t *testing.T) {
		root, store := onlyQueued(t)
		qasWriteRecord(t, root, "sess-live", fcNow, qasWin(92, qasReset5), nil)
		sleeps := qasFakeClock(t, nil, func(n int) {
			if n == 1 {
				// The mid-flight shape (card t1516): the row arrives with the
				// queue item picked — an assigned row under a queued item is
				// excluded from arm (a) and would spin the wait to the bound.
				fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardAssigned, OwnerLabel: "lane-1", Stage: homestate.CardRun})
				nmSetState(t, store, "t1", factory.BacklogStatePicked)
			}
		})
		out, stderr, err := qasRunNext(t, "--wait", "--wait-bound", "1h", "--run", fcRun)
		if err != nil {
			t.Fatalf("next --wait: %v (stderr %q)", err, stderr)
		}
		qasAssertNotHeld(t, out, stderr)
		if *sleeps != 1 {
			t.Errorf("wait slept %d times, want 1 (the assigned card is leased at the next re-check)", *sleeps)
		}
		assertLeased(t, root)
		if q := nmQueueState(t, store, "t1"); q != factory.BacklogStatePicked {
			t.Errorf("t1 queue state = %s, want picked (leased through the row's edge, never while queued)", q)
		}
	})
}

// AC-QAS-011 — non-Claude lanes are never held.
func TestQAS_AC011_NonClaudeBackendsNeverHeld(t *testing.T) {
	leaseUnderPressure := func(t *testing.T, provider, backend string) {
		t.Helper()
		sdClearLaneEnv(t)
		root, store := sdMoaiFixture(t)
		fcQueue(t, store, factory.BacklogStateQueued)
		sdRegisterLane(t, root, "lane-1")
		qasEnableGate(t, root)
		qasWriteRecord(t, root, "sess-live", fcNow, qasWin(99, qasReset5), nil)
		sdLaneEnv(t, "lane-1", backend)
		t.Setenv(config.EnvMoaiLaunchProvider, provider)
		t.Chdir(root)
		out, stderr, err := qasRunNext(t, "--run", fcRun)
		if err != nil {
			t.Fatalf("next: %v (stderr %q)", err, stderr)
		}
		qasAssertNotHeld(t, out, stderr)
		if cd := fcCard(t, root, "t1"); cd.State != homestate.CardLeased {
			t.Fatalf("t1 = %s, want leased", cd.State)
		}
	}
	for _, token := range []string{factory.BackendGPT, factory.BackendGLM, "", "no-such-backend"} {
		t.Run("backend_"+strconv.Quote(token), func(t *testing.T) { leaseUnderPressure(t, token, token) })
	}
	t.Run("launch_provider_overrides_the_backend_variable", func(t *testing.T) {
		leaseUnderPressure(t, factory.BackendGLM, factory.BackendClaude)
	})
	t.Run("backend_variable_is_the_fallback_when_no_launch_provider", func(t *testing.T) {
		sdClearLaneEnv(t)
		root, store := sdMoaiFixture(t)
		fcQueue(t, store, factory.BacklogStateQueued)
		sdRegisterLane(t, root, "lane-1")
		qasEnableGate(t, root)
		qasWriteRecord(t, root, "sess-live", fcNow, qasWin(99, qasReset5), nil)
		sdLaneEnv(t, "lane-1", factory.BackendClaude)
		t.Setenv(config.EnvMoaiLaunchProvider, "")
		t.Chdir(root)
		out, stderr, err := qasRunNext(t, "--run", fcRun)
		qasAssertHeld(t, out, stderr, err)
	})
}

// AC-QAS-011b — stage and complete are never gated: a held Claude lane's
// leased card advances and refuses exactly as under a disabled gate.
func TestQAS_AC011b_StageAndCompleteIgnoreQuotaHold(t *testing.T) {
	run := func(t *testing.T, held bool) (string, string) {
		t.Helper()
		sdClearLaneEnv(t)
		root, _ := fcFixture(t)
		if err := os.WriteFile(filepath.Join(root, "plan.md"), []byte("plan\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		fcGit(t, root, "add", "-A")
		fcGit(t, root, "commit", "-q", "-m", "plan artifact")
		sha := fcGit(t, root, "rev-parse", "HEAD")
		sdRegisterLane(t, root, "lane-1")
		fcPlace(t, root, homestate.Card{
			CardID: "t1", State: homestate.CardPlan, Stage: homestate.CardPlan,
			OwnerLabel: "lane-1", LeaseHolder: "lane-1",
			LeaseExpiresAt: fcNow.Add(10 * time.Minute).Format(time.RFC3339Nano),
			WorktreePath:   root,
		})
		if held {
			qasEnableGate(t, root)
			qasWriteRecord(t, root, "sess-live", fcNow, qasWin(99, qasReset5), qasWin(99, qasReset7))
		}
		qasLaneEnv(t, "lane-1", factory.BackendClaude)
		t.Chdir(root)
		if _, _, err := runFactory(t, "stage", "t1", "plan-audit", sha+":plan.md", "--run", fcRun); err != nil {
			t.Fatalf("stage (held=%v): %v", held, err)
		}
		cd := fcCard(t, root, "t1")
		_, _, completeErr := runFactory(t, "complete", "t1", "--run", fcRun)
		refusal := "<nil>"
		if completeErr != nil {
			refusal = strings.ReplaceAll(completeErr.Error(), root, "<root>")
		}
		return fmt.Sprintf("%s v%d stage=%s", cd.State, cd.Version, cd.Stage), refusal
	}
	heldCard, heldComplete := run(t, true)
	openCard, openComplete := run(t, false)
	if heldCard != openCard {
		t.Errorf("stage under a hold = %q, want the gate-disabled result %q", heldCard, openCard)
	}
	if heldComplete != openComplete {
		t.Errorf("complete under a hold = %q, want the gate-disabled result %q", heldComplete, openComplete)
	}
	if strings.Contains(heldComplete, "quota") {
		t.Errorf("complete mentions quota: %q", heldComplete)
	}
}

// qasMinSweptFiles is the least count of non-test files the AC-QAS-014 sweep
// must see across internal/statusline/quota*.go and internal/cli/factory_quota*.go:
// the SPEC's three (quota.go, factory_quota.go, factory_quota_lanes.go). It was
// one at M3, while only factory_quota.go existed; plan.md debt N1 reserves the
// full floor for M5, where factory_quota_lanes.go joins.
const qasMinSweptFiles = 3

// qasTreeSnapshot lists every file under dir with its size and digest.
func qasTreeSnapshot(t *testing.T, dir string) string {
	t.Helper()
	var rows []string
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		raw, rerr := os.ReadFile(path)
		if rerr != nil {
			return rerr
		}
		rel, _ := filepath.Rel(dir, path)
		rows = append(rows, fmt.Sprintf("%s %d %x", rel, len(raw), sha256.Sum256(raw)))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(rows)
	return strings.Join(rows, "\n")
}

// AC-QAS-014 (cli half) — the quota path is offline, spawn-free, and read-only:
// the files matching the glob import no net, net/http, or os/exec, and the
// pressure evaluation leaves the record directory byte-identical.
func TestQAS_AC014_AggregatorIsOfflineSpawnFreeAndReadOnly(t *testing.T) {
	var swept []string
	for _, pattern := range []string{"factory_quota*.go", filepath.Join("..", "statusline", "quota*.go")} {
		files, err := filepath.Glob(pattern)
		if err != nil {
			t.Fatal(err)
		}
		for _, f := range files {
			if !strings.HasSuffix(f, "_test.go") {
				swept = append(swept, f)
			}
		}
	}
	if len(swept) < qasMinSweptFiles {
		t.Fatalf("the sweep matched %d non-test files %v, want at least %d (a vanished file must not shrink the sweep silently)", len(swept), swept, qasMinSweptFiles)
	}
	for _, name := range []string{"quota.go", "factory_quota.go", "factory_quota_lanes.go"} {
		found := false
		for _, f := range swept {
			if filepath.Base(f) == name {
				found = true
			}
		}
		if !found {
			t.Errorf("the sweep %v does not contain %s", swept, name)
		}
	}
	forbidden := map[string]bool{"net": true, "net/http": true, "os/exec": true}
	for _, f := range swept {
		parsed, err := parser.ParseFile(token.NewFileSet(), f, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse %s: %v", f, err)
		}
		for _, imp := range parsed.Imports {
			if path, _ := strconv.Unquote(imp.Path.Value); forbidden[path] {
				t.Errorf("%s imports %s", f, path)
			}
		}
	}

	sdClearLaneEnv(t)
	root, _ := sdMoaiFixture(t)
	qasEnableGate(t, root)
	qasWriteRecord(t, root, "sess-a", fcNow, qasWin(92, qasReset5), qasWin(40, qasReset7))
	qasWriteRecord(t, root, "sess-b", fcNow.Add(-time.Hour), qasWin(10, qasReset5), nil)
	recordDir := filepath.Join(root, ".moai", "state", "context-usage")
	before := qasTreeSnapshot(t, recordDir)
	if ev := factoryQuotaEvaluate(root); !ev.Pressure() {
		t.Fatalf("the evaluation reported no pressure on the 92%% fixture: %+v", ev)
	}
	if after := qasTreeSnapshot(t, recordDir); after != before {
		t.Errorf("the record directory changed during an evaluation:\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

// qasAcquireWarningRE is the whole acquire warning: the prefix once, one segment
// per held window (the lane gate's own segment), then the warn-only tail.
var qasAcquireWarningRE = regexp.MustCompile(`^quota warning: ` + qasHoldSegment + `(; ` + qasHoldSegment + `)* \(warn-only; the integration window is still taken\)$`)

// qasAcquireOutcome is everything one `moai integration acquire` run produced
// that AC-QAS-013 compares: both streams, the exit status, and the lock record
// with its acquisition time removed (the only field that differs between two
// runs of the same acquire).
type qasAcquireOutcome struct {
	stdout, stderr string
	err            error
	lock           string
}

// qasAcquire runs `moai integration acquire` against root with the same holder
// identity every time, so two roots' outcomes are comparable.
func qasAcquire(t *testing.T, root string, extra ...string) qasAcquireOutcome {
	t.Helper()
	args := append([]string{"acquire", "--session", "sess-qas", "--name", "lane-1", "--branch", "release/v9.9.9"}, extra...)
	stdout, stderr, err := runIntegrationStreams(t, root, args...)
	rec, readErr := factory.ReadIntegrationLock(root)
	if readErr != nil {
		t.Fatalf("read lock record: %v", readErr)
	}
	if rec != nil {
		rec.AcquiredAt = ""
		// REQ-MWQ-008 (card t1479) stamps the lease too — another
		// wall-clock field that differs between two runs of the same
		// acquire, exactly like AcquiredAt.
		rec.LeaseExpiresAt = ""
	}
	raw, marshalErr := json.Marshal(rec)
	if marshalErr != nil {
		t.Fatal(marshalErr)
	}
	return qasAcquireOutcome{stdout: stdout, stderr: stderr, err: err, lock: string(raw)}
}

// qasAcquireRoot builds an acquire fixture: a Claude lane, the gate on unless
// o.gateOff, and one fresh record carrying five (and seven when set).
func qasAcquireRoot(t *testing.T, backend string, five, seven *statusline.QuotaWindowRecord, o qasFixtureOpts) string {
	t.Helper()
	sdClearLaneEnv(t)
	root, _ := sdMoaiFixture(t)
	if !o.gateOff {
		qasEnableGate(t, root)
	}
	if five != nil || seven != nil {
		qasWriteRecord(t, root, "sess-live", fcNow, five, seven)
	}
	qasLaneEnv(t, "lane-1", backend)
	return root
}

// AC-QAS-013 — `moai integration acquire` warns and never blocks (REQ-QAS-014,
// DO-7 final: warn-only). The warning is one line on the error stream; the lock
// record, the exit status, and standard output (the --json object included) are
// those of a run with the gate disabled.
func TestQAS_AC013_AcquireWarnsNeverRefuses(t *testing.T) {
	// control is the same acquire with the gate disabled — the baseline every
	// pressure case is compared to.
	control := func(t *testing.T, extra ...string) qasAcquireOutcome {
		t.Helper()
		root := qasAcquireRoot(t, factory.BackendClaude, qasWin(92, qasReset5), nil, qasFixtureOpts{gateOff: true})
		return qasAcquire(t, root, extra...)
	}
	// quotaLines counts the stderr lines that mention quota.
	quotaLines := func(stderr string) []string {
		var lines []string
		for _, l := range strings.Split(strings.TrimRight(stderr, "\n"), "\n") {
			if strings.Contains(l, "quota") {
				lines = append(lines, l)
			}
		}
		return lines
	}
	// withoutQuota drops the quota lines so the rest of stderr can be compared.
	withoutQuota := func(stderr string) string {
		var keep []string
		for _, l := range strings.Split(stderr, "\n") {
			if !strings.Contains(l, "quota") {
				keep = append(keep, l)
			}
		}
		return strings.Join(keep, "\n")
	}
	assertUnchanged := func(t *testing.T, got, want qasAcquireOutcome) {
		t.Helper()
		if got.err != nil {
			t.Errorf("acquire failed: %v (a warning must never refuse the window)", got.err)
		}
		if got.stdout != want.stdout {
			t.Errorf("stdout differs from the gate-disabled run:\n got: %q\nwant: %q", got.stdout, want.stdout)
		}
		if got.lock != want.lock {
			t.Errorf("lock record differs from the gate-disabled run:\n got: %s\nwant: %s", got.lock, want.lock)
		}
		if withoutQuota(got.stderr) != withoutQuota(want.stderr) {
			t.Errorf("stderr outside the quota line differs from the gate-disabled run:\n got: %q\nwant: %q", got.stderr, want.stderr)
		}
	}

	t.Run("warning_line_names_window_and_reset", func(t *testing.T) {
		root := qasAcquireRoot(t, factory.BackendClaude, qasWin(92, qasReset5), nil, qasFixtureOpts{})
		got := integrationQuotaWarning(root)
		if !qasAcquireWarningRE.MatchString(got) {
			t.Fatalf("warning = %q, want a line matching %s", got, qasAcquireWarningRE)
		}
		wantReset := time.Unix(qasReset5, 0).UTC().Format(time.RFC3339)
		if !strings.Contains(got, "five_hour used=92.0%") || !strings.Contains(got, "resets_at="+wantReset) {
			t.Errorf("warning %q does not name the five_hour window at 92.0%% resetting at %s", got, wantReset)
		}
	})
	t.Run("warns_text_and_never_blocks", func(t *testing.T) {
		want := control(t)
		root := qasAcquireRoot(t, factory.BackendClaude, qasWin(92, qasReset5), nil, qasFixtureOpts{})
		got := qasAcquire(t, root)
		lines := quotaLines(got.stderr)
		if len(lines) != 1 || !qasAcquireWarningRE.MatchString(lines[0]) {
			t.Fatalf("stderr quota lines = %q, want exactly one matching %s", lines, qasAcquireWarningRE)
		}
		if strings.Contains(got.stdout, "quota") {
			t.Errorf("stdout carries the warning: %q", got.stdout)
		}
		assertUnchanged(t, got, want)
	})
	t.Run("warns_json_and_never_blocks", func(t *testing.T) {
		want := control(t, "--json")
		root := qasAcquireRoot(t, factory.BackendClaude, qasWin(92, qasReset5), nil, qasFixtureOpts{})
		got := qasAcquire(t, root, "--json")
		lines := quotaLines(got.stderr)
		if len(lines) != 1 || !qasAcquireWarningRE.MatchString(lines[0]) {
			t.Fatalf("stderr quota lines = %q, want exactly one matching %s", lines, qasAcquireWarningRE)
		}
		var obj map[string]any
		if err := json.Unmarshal([]byte(got.stdout), &obj); err != nil {
			t.Fatalf("--json stdout is not one parseable object (%v): %q", err, got.stdout)
		}
		assertUnchanged(t, got, want)
	})
	t.Run("both_windows_one_line", func(t *testing.T) {
		root := qasAcquireRoot(t, factory.BackendClaude, qasWin(92, qasReset5), qasWin(96, qasReset7), qasFixtureOpts{})
		got := qasAcquire(t, root)
		lines := quotaLines(got.stderr)
		if len(lines) != 1 || !qasAcquireWarningRE.MatchString(lines[0]) {
			t.Fatalf("stderr quota lines = %q, want exactly one matching %s", lines, qasAcquireWarningRE)
		}
		if !strings.Contains(lines[0], "five_hour") || !strings.Contains(lines[0], "seven_day") {
			t.Errorf("the one line does not name both held windows: %q", lines[0])
		}
	})
	t.Run("no_line_and_output_unchanged", func(t *testing.T) {
		cases := []struct {
			name    string
			backend string
			five    *statusline.QuotaWindowRecord
			gateOff bool
		}{
			{"below_threshold", factory.BackendClaude, qasWin(89.9, qasReset5), false},
			{"reset", factory.BackendClaude, qasWin(92, fcNow.Add(-time.Second).Unix()), false},
			{"unknown_no_record", factory.BackendClaude, nil, false},
			{"gate_disabled", factory.BackendClaude, qasWin(99, qasReset5), true},
			{"non_claude_glm", factory.BackendGLM, qasWin(92, qasReset5), false},
			{"non_claude_gpt", factory.BackendGPT, qasWin(92, qasReset5), false},
			{"no_backend", "", qasWin(92, qasReset5), false},
		}
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				want := control(t)
				root := qasAcquireRoot(t, c.backend, c.five, nil, qasFixtureOpts{gateOff: c.gateOff})
				got := qasAcquire(t, root)
				if lines := quotaLines(got.stderr); len(lines) != 0 {
					t.Errorf("stderr carries a quota line: %q", lines)
				}
				if got.stderr != want.stderr {
					t.Errorf("stderr differs from the gate-disabled run:\n got: %q\nwant: %q", got.stderr, want.stderr)
				}
				assertUnchanged(t, got, want)
			})
		}
	})
	t.Run("unreadable_quota_state_falls_through_silently", func(t *testing.T) {
		want := control(t)
		root := qasAcquireRoot(t, factory.BackendClaude, nil, nil, qasFixtureOpts{})
		dir := filepath.Join(root, ".moai", "state", "context-usage")
		if err := os.MkdirAll(dir, 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "sess-broken.json"), []byte("{not json"), 0o600); err != nil {
			t.Fatal(err)
		}
		got := qasAcquire(t, root)
		if lines := quotaLines(got.stderr); len(lines) != 0 {
			t.Errorf("stderr carries a quota line for an unreadable record: %q", lines)
		}
		assertUnchanged(t, got, want)
	})
}

// AC-QAS-017 — one pressure evaluation that does not look at the caller; the
// lane gate applies the Claude-caller predicate to its result. M5 adds the
// status-block and --auto adoption assertions (the status_and_auto_surfaces
// subtest); M6 adds the integration-window warning (the acquire_warning
// subtest), which applies the same Claude-caller predicate.
func TestQAS_AC017_SharedPressureEvaluationAndSurfaces(t *testing.T) {
	setup := func(t *testing.T, five *statusline.QuotaWindowRecord, o qasFixtureOpts) string {
		t.Helper()
		sdClearLaneEnv(t)
		root, _ := sdMoaiFixture(t)
		if !o.gateOff {
			qasEnableGate(t, root)
		}
		if five != nil {
			qasWriteRecord(t, root, "sess-live", fcNow, five, nil)
		}
		return root
	}
	laneHolds := func(t *testing.T, root, backend string) bool {
		t.Helper()
		qasLaneEnv(t, "lane-1", backend)
		held, _ := (&factoryQuotaLatch{}).evaluate(root)
		return held
	}
	// surfaces asserts the two M5 surfaces against the pressure the shared
	// function reports: the status quota block (the `pressure` flag) and the
	// --auto line. Neither applies a caller rule, so the lane environment a
	// caller subtest left set does not matter. The registry carries the
	// standard lanes so a recommendation has something to name.
	surfaces := func(t *testing.T, root string, wantPressure bool, acquireBackend string) {
		t.Helper()
		t.Run("status_and_auto_surfaces", func(t *testing.T) {
			qasLaneSeam(t)
			qasWriteLanes(t, root, qasStandardLaneRows())
			text, js := qasStatus(t)
			quota, present := qasStatusQuota(t, js)
			if got := present && quota["pressure"] == true; got != wantPressure {
				t.Errorf("status quota pressure = %v (block present: %v), want %v", got, present, wantPressure)
			}
			if got := strings.Contains(text, "\nquota pressure:"); got != wantPressure {
				t.Errorf("status text carries a quota pressure line = %v, want %v:\n%s", got, wantPressure, text)
			}
			if got := todoAutoQuotaLine(root) != ""; got != wantPressure {
				t.Errorf("--auto steering line present = %v, want %v", got, wantPressure)
			}
		})
		t.Run("acquire_warning", func(t *testing.T) {
			// The warning needs the shared pressure AND a Claude caller; the
			// caller is named per case, never read back from the environment a
			// sibling subtest left set.
			qasLaneEnv(t, "lane-1", acquireBackend)
			out := qasAcquire(t, root)
			if out.err != nil {
				t.Fatalf("acquire failed: %v", out.err)
			}
			warned := strings.Contains(out.stderr, "quota")
			if want := wantPressure && acquireBackend == factory.BackendClaude; warned != want {
				t.Errorf("acquire warning printed = %v, want %v (pressure %v, caller %q); stderr: %q", warned, want, wantPressure, acquireBackend, out.stderr)
			}
			if strings.Contains(out.stdout, "quota") {
				t.Errorf("stdout carries quota text: %q", out.stdout)
			}
		})
	}

	t.Run("shared_function_ignores_caller", func(t *testing.T) {
		root := setup(t, qasWin(92, qasReset5), qasFixtureOpts{})
		var first factoryQuotaEvaluation
		for i, backend := range []string{factory.BackendClaude, factory.BackendGLM, factory.BackendGPT, ""} {
			qasLaneEnv(t, "lane-1", backend)
			ev := factoryQuotaEvaluate(root)
			if !ev.Pressure() {
				t.Fatalf("backend %q: pressure off, want on", backend)
			}
			held := ev.HeldWindows()
			if len(held) != 1 || held[0].Name != "five_hour" || held[0].Reading.UsedPercentage != 92 || held[0].Reading.ResetsAt != qasReset5 {
				t.Fatalf("backend %q: held windows = %+v, want one five_hour reading at 92 resetting at %d", backend, held, qasReset5)
			}
			if i == 0 {
				first = ev
			} else if fmt.Sprintf("%+v", ev) != fmt.Sprintf("%+v", first) {
				t.Errorf("backend %q: the evaluation differs from the Claude caller's:\n%+v\n%+v", backend, ev, first)
			}
		}
	})
	t.Run("at_92", func(t *testing.T) {
		root := setup(t, qasWin(92, qasReset5), qasFixtureOpts{})
		if !laneHolds(t, root, factory.BackendClaude) {
			t.Errorf("the lane gate does not hold a Claude lane at 92%%")
		}
		surfaces(t, root, true, factory.BackendClaude)
	})
	t.Run("acquire_requires_claude_caller", func(t *testing.T) {
		root := setup(t, qasWin(92, qasReset5), qasFixtureOpts{})
		if laneHolds(t, root, factory.BackendGLM) {
			t.Errorf("the lane gate holds a non-Claude lane")
		}
		if !factoryQuotaEvaluate(root).Pressure() {
			t.Errorf("the shared function stopped reporting pressure for a non-Claude caller")
		}
		surfaces(t, root, true, factory.BackendGLM)
	})
	t.Run("at_89_9", func(t *testing.T) {
		root := setup(t, qasWin(89.9, qasReset5), qasFixtureOpts{})
		if factoryQuotaEvaluate(root).Pressure() || laneHolds(t, root, factory.BackendClaude) {
			t.Errorf("pressure or a hold at 89.9%%")
		}
		surfaces(t, root, false, factory.BackendClaude)
	})
	t.Run("reset", func(t *testing.T) {
		root := setup(t, qasWin(92, fcNow.Add(-time.Second).Unix()), qasFixtureOpts{})
		if factoryQuotaEvaluate(root).Pressure() || laneHolds(t, root, factory.BackendClaude) {
			t.Errorf("pressure or a hold on a window whose reset time has passed")
		}
		surfaces(t, root, false, factory.BackendClaude)
	})
	t.Run("unknown", func(t *testing.T) {
		root := setup(t, nil, qasFixtureOpts{})
		if factoryQuotaEvaluate(root).Pressure() || laneHolds(t, root, factory.BackendClaude) {
			t.Errorf("pressure or a hold with no reading at all")
		}
		surfaces(t, root, false, factory.BackendClaude)
	})
	t.Run("gate_disabled", func(t *testing.T) {
		root := setup(t, qasWin(99, qasReset5), qasFixtureOpts{gateOff: true})
		if factoryQuotaEvaluate(root).Pressure() || laneHolds(t, root, factory.BackendClaude) {
			t.Errorf("pressure or a hold with the gate disabled")
		}
		surfaces(t, root, false, factory.BackendClaude)
	})
}
