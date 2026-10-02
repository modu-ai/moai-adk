// factory_quota_lanes_test.go — SPEC-QUOTA-AWARE-SCHEDULING-001 M5 AC tests
// (card t1347): the steering surfaces. With quota pressure on, `moai todo
// --auto` and `moai factory status` recommend the live non-Claude lanes from the
// factory registry (or warn when none exists), the status command adds its
// read-only quota block, and nothing — a card, a queue entry, a registry byte —
// changes. AC-QAS-012, -018, -019, -020, -021, -022.
//
// Every fixture is built under t.TempDir(); ./internal/cli runs only through the
// anchored -run selectors naming one of these tests. The lane environment this
// suite may run in is cleared per test (sdClearLaneEnv), and the registry's
// liveness probe is a seam, so no real process decides a verdict.
package cli

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// qasDeadPID is the one registered pid the liveness seam reports dead.
const qasDeadPID = 5004

// qasLaneSeam makes every registered pid alive except qasDeadPID.
func qasLaneSeam(t *testing.T) {
	t.Helper()
	prev := factoryProcessAlive
	factoryProcessAlive = func(pid int) bool { return pid != qasDeadPID }
	t.Cleanup(func() { factoryProcessAlive = prev })
}

// qasRegistryLane is one registry row. legacy rows are inserted with the pre-change
// statement, which names no backend column (the row reads back empty).
type qasRegistryLane struct {
	label   string
	pid     int
	backend string
	legacy  bool
}

// qasStandardLaneRows is the registry of AC-QAS-018: lane-1 claude, lane-2 glm,
// lane-3 gpt, lane-4 dead glm, lane-5 legacy (empty backend), lane-6 unrecognised.
func qasStandardLaneRows() []qasRegistryLane {
	return []qasRegistryLane{
		{label: "lane-1", pid: 5001, backend: kanban.BackendClaude},
		{label: "lane-2", pid: 5002, backend: kanban.BackendGLM},
		{label: "lane-3", pid: 5003, backend: kanban.BackendGPT},
		{label: "lane-4", pid: qasDeadPID, backend: kanban.BackendGLM},
		{label: "lane-5", pid: 5005, legacy: true},
		{label: "lane-6", pid: 5006, backend: "mystery"},
	}
}

// qasWriteLanes creates the factory registry through the real schema and inserts
// the rows with plain SQL. It goes through homestate.OpenFactory, which creates
// the schema and the schema_version meta row but never the
// legacy_workers_imported marker, so a reader that opens the registry through
// LoadFactoryRegistry would add that row (plan debt N8).
func qasWriteLanes(t *testing.T, root string, rows []qasRegistryLane) {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	defer func() { _ = db.Close() }()
	const at = "2026-09-28T00:00:00Z"
	for _, r := range rows {
		if r.legacy {
			_, err = db.DB.Exec(`INSERT INTO workers(label,pid,registered_at,heartbeat_at) VALUES(?,?,?,?)`, r.label, r.pid, at, at)
		} else {
			_, err = db.DB.Exec(`INSERT INTO workers(label,pid,backend,registered_at,heartbeat_at) VALUES(?,?,?,?,?)`, r.label, r.pid, r.backend, at, at)
		}
		if err != nil {
			t.Fatalf("insert lane %s: %v", r.label, err)
		}
	}
}

// qasRegistryState fingerprints the registry: the main database file's SHA-256
// and a row dump of workers, cards, and meta, read through an independent
// read-only connection. The directory listing is deliberately not part of it —
// SQLite may create -wal/-shm sidecars for a read-only reader of a WAL database.
func qasRegistryState(t *testing.T, root string) string {
	t.Helper()
	path, err := homestate.FactoryDBPath(root)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(raw)
	v := url.Values{}
	v.Add("mode", "ro")
	v.Add("_pragma", "query_only(ON)")
	db, err := sql.Open("sqlite", (&url.URL{Scheme: "file", Path: filepath.ToSlash(path), RawQuery: v.Encode()}).String())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	var b strings.Builder
	fmt.Fprintf(&b, "main-sha256=%s\n", hex.EncodeToString(sum[:]))
	for _, q := range []struct{ table, order string }{{"workers", "label"}, {"cards", "run_id,card_id"}, {"meta", "key"}} {
		rows, err := db.Query(`SELECT * FROM ` + q.table + ` ORDER BY ` + q.order)
		if err != nil {
			t.Fatalf("dump %s: %v", q.table, err)
		}
		cols, _ := rows.Columns()
		for rows.Next() {
			vals := make([]any, len(cols))
			ptrs := make([]any, len(cols))
			for i := range vals {
				ptrs[i] = &vals[i]
			}
			if err := rows.Scan(ptrs...); err != nil {
				t.Fatal(err)
			}
			fmt.Fprintf(&b, "%s|%v\n", q.table, vals)
		}
		if err := rows.Err(); err != nil {
			t.Fatal(err)
		}
		_ = rows.Close()
	}
	return b.String()
}

// qasCodeIdents returns every identifier the file's code uses — selectors'
// field names included, comments excluded — so a static check is not tripped by
// a comment that names what the code deliberately avoids.
func qasCodeIdents(t *testing.T, file string) map[string]bool {
	t.Helper()
	parsed, err := parser.ParseFile(token.NewFileSet(), file, nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", file, err)
	}
	out := map[string]bool{}
	ast.Inspect(parsed, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok {
			out[id.Name] = true
		}
		return true
	})
	return out
}

// qasMoaiTreeNames lists every path under dir (names only).
func qasMoaiTreeNames(t *testing.T, dir string) string {
	t.Helper()
	var names []string
	_ = filepath.WalkDir(dir, func(p string, _ os.DirEntry, err error) error {
		if err == nil {
			rel, _ := filepath.Rel(dir, p)
			names = append(names, rel)
		}
		return nil
	})
	return strings.Join(names, "\n")
}

// ---------------------------------------------------------------------------
// The baseline goldens (M0). They are read, never written.

// qasGoldenStatus returns the golden's `empty` or `one_card` object re-indented
// exactly as the command prints it (the M0 extraction rule).
func qasGoldenStatus(t *testing.T, key string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "qas_baseline_factory_status.golden.json"))
	if err != nil {
		t.Fatal(err)
	}
	var all map[string]json.RawMessage
	if err := json.Unmarshal(raw, &all); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := json.Indent(&buf, all[key], "", "  "); err != nil {
		t.Fatal(err)
	}
	buf.WriteByte('\n')
	return buf.String()
}

func qasGoldenAuto(t *testing.T) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "qas_baseline_todo_auto.golden.txt"))
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

// qasStatusRoot builds the status fixture of the M0 golden: `empty` has no
// factory database; `one_card` has one leased card. The golden's registered
// lane row is not reproduced: `factory status` prints card rows only, and the
// lane rows of a test belong to the caller (qasWriteLanes), so a standard
// registry never collides with a fixture row.
func qasStatusRoot(t *testing.T, oneCard bool) (string, *kanban.BacklogStore) {
	t.Helper()
	sdClearLaneEnv(t)
	root, store := sdMoaiFixture(t)
	if oneCard {
		fcQueue(t, store, kanban.BacklogStateQueued)
		fcClassify(t, store, "t1", kanban.ClassPriorityHigh, false, kanban.ClassModeSerial)
		fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardLeased, LeaseHolder: "lane-1", LeaseExpiresAt: "2026-09-26T10:00:00Z", Stage: homestate.CardRun})
	}
	return root, store
}

// qasStatus runs `moai factory status --run run-cli` in text and JSON.
func qasStatus(t *testing.T) (text, js string) {
	t.Helper()
	text, _, err := runFactory(t, "status", "--run", fcRun)
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	js, _, err = runFactory(t, "status", "--run", fcRun, "--json")
	if err != nil {
		t.Fatalf("status --json: %v", err)
	}
	return text, js
}

// qasStatusQuota decodes the `quota` key of a status --json output; ok is false
// when the key is absent.
func qasStatusQuota(t *testing.T, js string) (map[string]any, bool) {
	t.Helper()
	var all map[string]any
	if err := json.Unmarshal([]byte(js), &all); err != nil {
		t.Fatalf("status json: %v\n%s", err, js)
	}
	q, ok := all["quota"]
	if !ok {
		return nil, false
	}
	m, isMap := q.(map[string]any)
	if !isMap {
		t.Fatalf("quota key is %T, want an object", q)
	}
	return m, true
}

// qasAutoRun runs the M0 `--auto` recipe over two queued cards (t1 "card a", t2
// "card b") with the Jev line stubbed and a fake clock, through the production
// quota seam, and returns the output with the per-run root rendered as <ROOT>.
func qasAutoRun(t *testing.T, root string, store *kanban.BacklogStore) string {
	t.Helper()
	for _, text := range []string{"card a", "card b"} {
		if _, _, err := store.Add(text); err != nil {
			t.Fatal(err)
		}
	}
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	ids := []string{rec.Items[0].ID, rec.Items[1].ID}
	tick := 0
	opts := autoOptions{
		wait:      5 * time.Minute,
		liveness:  autoTestLiveness(root, "t1", true, true, nil),
		sessionID: "operator-session-fixture",
		jev:       func(string) string { return "jev: stubbed (qas baseline fixture)" },
		quota:     todoAutoQuotaLine,
		sleep: func(time.Duration) {
			tick++
			path := filepath.Join(root, ".moai", "reports", ids[tick-1], "evidence.md")
			if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte("# evidence: verbatim output\n"), 0o600); err != nil {
				t.Fatal(err)
			}
		},
		now: func() time.Time { return time.Unix(0, 0).Add(time.Duration(tick) * time.Minute) },
	}
	var out bytes.Buffer
	if err := runAutoCycle(&out, store, root, opts); err != nil {
		t.Fatal(err)
	}
	return strings.ReplaceAll(out.String(), root, "<ROOT>")
}

const qasLineTime = `[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z`

var (
	qasRecommendRE = regexp.MustCompile(`^quota pressure: five_hour used=[0-9]+\.[0-9]% resets_at=` + qasLineTime + `; recommend non-Claude lane\(s\): lane-2 \(glm\), lane-3 \(gpt\)$`)
	qasWarningRE   = regexp.MustCompile(`^quota pressure: five_hour used=[0-9]+\.[0-9]% resets_at=` + qasLineTime + `; warning: no live non-Claude lane \(unknown backend: [0-9]+\); nothing is re-dispatched$`)
)

// qasPressureRoot is the status fixture with the gate enabled and the five-hour
// window at 92%.
func qasPressureRoot(t *testing.T, oneCard bool, five float64) (string, *kanban.BacklogStore) {
	t.Helper()
	root, store := qasStatusRoot(t, oneCard)
	qasEnableGate(t, root)
	qasWriteRecord(t, root, "sess-live", fcNow, qasWin(five, qasReset5), nil)
	return root, store
}

// ---------------------------------------------------------------------------
// AC-QAS-012

// AC-QAS-012 — `moai factory status` shows the quota block only when the gate is
// enabled and a window has data; with the gate disabled, or enabled with no
// window data, the output is byte-identical to the baseline golden.
func TestQAS_AC012_StatusQuotaBlockOnlyWhenGateEnabledWithData(t *testing.T) {
	t.Run("enabled_with_data_adds_block_only", func(t *testing.T) {
		qasLaneSeam(t)
		qasPressureRoot(t, true, 50)
		text, js := qasStatus(t)

		golden := qasGoldenStatus(t, "one_card")
		wantPrefix := strings.TrimSuffix(golden, "\n}\n")
		if !strings.HasPrefix(js, wantPrefix+",\n  \"quota\": {") {
			t.Fatalf("status --json is not the baseline golden plus a trailing quota key:\n%s", js)
		}
		quota, ok := qasStatusQuota(t, js)
		if !ok {
			t.Fatal("no quota key although the gate is enabled and a window has data")
		}
		if quota["pressure"] != false {
			t.Errorf("pressure = %v at 50%%, want false", quota["pressure"])
		}
		for _, k := range []string{"candidates", "unknown_lanes", "warning"} {
			if _, present := quota[k]; present {
				t.Errorf("quota carries %q although pressure is off (no steering key may appear)", k)
			}
		}
		if strings.Contains(text, "quota pressure") || strings.Contains(text, "quota hold") {
			t.Errorf("status text carries a steering or hold line while pressure is off:\n%s", text)
		}

		// The text is the gate-disabled text plus quota lines only (the disabled
		// run is a second root, so no configuration cache can carry the gate over).
		qasStatusRoot(t, true)
		offText, _, err := runFactory(t, "status", "--run", fcRun)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(text, offText) {
			t.Fatalf("enabled text does not start with the gate-disabled text:\nenabled:\n%s\ndisabled:\n%s", text, offText)
		}
		rest := strings.TrimSuffix(strings.TrimPrefix(text, offText), "\n")
		if rest == "" {
			t.Fatal("enabled text adds no quota block")
		}
		for _, line := range strings.Split(rest, "\n") {
			if !strings.HasPrefix(line, "quota ") {
				t.Errorf("a non-quota line was added to the status text: %q", line)
			}
		}
	})

	t.Run("gate_disabled_with_data_equals_golden", func(t *testing.T) {
		for _, key := range []string{"empty", "one_card"} {
			root, _ := qasStatusRoot(t, key == "one_card")
			qasWriteRecord(t, root, "sess-live", fcNow, qasWin(92, qasReset5), qasWin(40, qasReset7))
			_, js := qasStatus(t)
			if js != qasGoldenStatus(t, key) {
				t.Errorf("%s: gate disabled with data present, status --json differs from the golden:\n%s", key, js)
			}
			text, _ := qasStatus(t)
			if strings.Contains(text, "quota") {
				t.Errorf("%s: gate disabled, status text mentions quota:\n%s", key, text)
			}
		}
	})

	t.Run("enabled_no_data_equals_golden", func(t *testing.T) {
		for _, key := range []string{"empty", "one_card"} {
			root, _ := qasStatusRoot(t, key == "one_card")
			qasEnableGate(t, root)
			_, js := qasStatus(t)
			if js != qasGoldenStatus(t, key) {
				t.Errorf("%s: gate enabled with no window data, status --json differs from the golden:\n%s", key, js)
			}
			// A record that is too old is no data either.
			qasWriteRecord(t, root, "sess-old", fcNow.Add(-2*time.Hour), qasWin(97, qasReset5), nil)
			_, js = qasStatus(t)
			if js != qasGoldenStatus(t, key) {
				t.Errorf("%s: a stale record produced a quota block:\n%s", key, js)
			}
		}
	})

	t.Run("exhausted_at_shown_text_and_json", func(t *testing.T) {
		qasLaneSeam(t)
		root, _ := qasStatusRoot(t, true)
		qasEnableGate(t, root)
		five := qasWin(100, qasReset5)
		five.ExhaustedAt = "2026-09-26T08:30:00Z"
		qasWriteRecord(t, root, "sess-live", fcNow, five, qasWin(40, qasReset7))
		text, js := qasStatus(t)
		if !strings.Contains(text, "exhausted_at=2026-09-26T08:30:00Z") {
			t.Errorf("status text does not show the recorded exhausted time:\n%s", text)
		}
		if n := strings.Count(text, "exhausted_at="); n != 1 {
			t.Errorf("status text shows %d exhausted_at cells, want 1 (only the window that has one)", n)
		}
		quota, ok := qasStatusQuota(t, js)
		if !ok {
			t.Fatal("no quota key")
		}
		windows, _ := quota["windows"].([]any)
		if len(windows) != 2 {
			t.Fatalf("windows = %v, want two", quota["windows"])
		}
		five5, _ := windows[0].(map[string]any)
		seven, _ := windows[1].(map[string]any)
		if five5["exhausted_at"] != "2026-09-26T08:30:00Z" {
			t.Errorf("five_hour exhausted_at = %v, want the recorded time", five5["exhausted_at"])
		}
		if _, present := seven["exhausted_at"]; present {
			t.Errorf("seven_day carries an exhausted_at it never recorded: %v", seven)
		}
	})

	t.Run("fields_per_window", func(t *testing.T) {
		qasLaneSeam(t)
		root, _ := qasStatusRoot(t, false)
		qasEnableGate(t, root)
		// five_hour fresh below its hold percentage; seven_day carried but already
		// reset; the second fixture below leaves seven_day absent (unknown).
		qasWriteRecord(t, root, "sess-live", fcNow, qasWin(40, qasReset5), qasWin(97, fcNow.Add(-time.Minute).Unix()))
		text, js := qasStatus(t)
		quota, ok := qasStatusQuota(t, js)
		if !ok {
			t.Fatal("no quota key")
		}
		windows, _ := quota["windows"].([]any)
		if len(windows) != 2 {
			t.Fatalf("windows = %v, want two", quota["windows"])
		}
		want := []struct {
			name, state string
			used        any
			holds       bool
		}{
			{"five_hour", "fresh", float64(40), false},
			{"seven_day", "reset", nil, false},
		}
		for i, w := range windows {
			m, _ := w.(map[string]any)
			if m["window"] != want[i].name || m["state"] != want[i].state || m["holds_claude_lane"] != want[i].holds || m["used_percentage"] != want[i].used {
				t.Errorf("window %d = %v, want %+v", i, m, want[i])
			}
		}
		five, _ := windows[0].(map[string]any)
		if five["resets_at"] != time.Unix(qasReset5, 0).UTC().Format(time.RFC3339) || five["captured_at"] != fcNow.Format(time.RFC3339) {
			t.Errorf("five_hour reset/capture times = %v / %v", five["resets_at"], five["captured_at"])
		}
		for _, frag := range []string{
			"quota five_hour: state=fresh used=40.0% resets_at=" + time.Unix(qasReset5, 0).UTC().Format(time.RFC3339) + " captured_at=" + fcNow.Format(time.RFC3339) + " holds_claude_lane=no",
			"quota seven_day: state=reset",
		} {
			if !strings.Contains(text, frag) {
				t.Errorf("status text lacks %q:\n%s", frag, text)
			}
		}

		// A window no record carries reads unknown, and a held window says so.
		qasWriteRecord(t, root, "sess-live", fcNow, qasWin(93, qasReset5), nil)
		text, js = qasStatus(t)
		quota, _ = qasStatusQuota(t, js)
		windows, _ = quota["windows"].([]any)
		five, _ = windows[0].(map[string]any)
		seven, _ := windows[1].(map[string]any)
		if five["holds_claude_lane"] != true || seven["state"] != "unknown" {
			t.Errorf("five_hour/seven_day = %v / %v, want a held five_hour and an unknown seven_day", five, seven)
		}
		if !strings.Contains(text, "holds_claude_lane=yes") || !strings.Contains(text, "quota seven_day: state=unknown") {
			t.Errorf("status text lacks the held/unknown cells:\n%s", text)
		}
	})
}

// ---------------------------------------------------------------------------
// AC-QAS-018

// AC-QAS-018 — the lane inventory lists live non-Claude lanes from the registry
// through a genuinely read-only open.
func TestQAS_AC018_LaneInventoryCandidates(t *testing.T) {
	newRoot := func(t *testing.T) string {
		t.Helper()
		sdClearLaneEnv(t)
		qasLaneSeam(t)
		root, _ := sdMoaiFixture(t)
		return root
	}
	assertInventory := func(t *testing.T, got factoryQuotaInventory, wantCandidates []factoryQuotaLane, wantUnknown int) {
		t.Helper()
		if !reflect.DeepEqual(got.Candidates, wantCandidates) || got.Unknown != wantUnknown {
			t.Errorf("inventory = %+v, want candidates %+v unknown %d", got, wantCandidates, wantUnknown)
		}
	}
	laneTwoThree := []factoryQuotaLane{{Label: "lane-2", Backend: kanban.BackendGLM}, {Label: "lane-3", Backend: kanban.BackendGPT}}

	t.Run("candidates_and_unknown", func(t *testing.T) {
		root := newRoot(t)
		qasWriteLanes(t, root, qasStandardLaneRows())
		assertInventory(t, factoryQuotaReadLanes(root), laneTwoThree, 2)
	})

	t.Run("legacy_empty_is_unknown", func(t *testing.T) {
		root := newRoot(t)
		qasWriteLanes(t, root, []qasRegistryLane{{label: "lane-5", pid: 5005, legacy: true}})
		assertInventory(t, factoryQuotaReadLanes(root), nil, 1)
	})

	t.Run("session_record_is_not_read", func(t *testing.T) {
		root := newRoot(t)
		qasWriteLanes(t, root, qasStandardLaneRows())
		// A record that would turn the legacy lane-5 into a glm candidate, and one
		// that would turn the registered glm lane-2 into a claude lane.
		for sid, rec := range map[string]*kanban.Record{
			"sess-five": kanban.NewRecord("sess-five", "", kanban.BackendGLM).WithRole("lane").WithLane(5),
			"sess-two":  kanban.NewRecord("sess-two", "", kanban.BackendClaude).WithRole("lane").WithLane(2),
		} {
			if err := kanban.Write(root, rec); err != nil {
				t.Fatalf("write session record %s: %v", sid, err)
			}
		}
		dir := filepath.Dir(kanban.RecordPath(root, "sess-five"))
		before := qasTreeSnapshot(t, dir)
		assertInventory(t, factoryQuotaReadLanes(root), laneTwoThree, 2)
		if after := qasTreeSnapshot(t, dir); after != before {
			t.Errorf("the session registry changed during the inventory read")
		}
		// Static: the inventory's code (comments excluded) names none of the
		// registry openers that write, and no session-record reader.
		idents := qasCodeIdents(t, "factory_quota_lanes.go")
		for _, banned := range []string{"ReadAll", "Read", "RecordPath", "LoadFactoryRegistry", "OpenFactoryPath", "OpenFactory", "ImportLegacyWorkers", "FactoryRegistryPath", "EnsureProjectLayout", "SaveFactoryRegistry"} {
			if idents[banned] {
				t.Errorf("factory_quota_lanes.go references %s; the inventory reads the registry row through its own read-only open and no session record", banned)
			}
		}
		// Positive control: the same scan does see an identifier the file uses.
		if !idents["FactoryDBPath"] {
			t.Errorf("control: the identifier scan did not see FactoryDBPath in factory_quota_lanes.go")
		}
	})

	t.Run("absent_registry_no_candidates_no_create", func(t *testing.T) {
		root := newRoot(t)
		path, err := homestate.FactoryDBPath(root)
		if err != nil {
			t.Fatal(err)
		}
		before := qasMoaiTreeNames(t, filepath.Join(root, ".moai"))
		assertInventory(t, factoryQuotaReadLanes(root), nil, 0)
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("the registry file exists after reading an absent registry: %v", err)
		}
		if after := qasMoaiTreeNames(t, filepath.Join(root, ".moai")); after != before {
			t.Errorf("a directory or file was created by reading an absent registry:\nbefore:\n%s\nafter:\n%s", before, after)
		}
		// Positive control: the same helper sees a registry that exists.
		qasWriteLanes(t, root, qasStandardLaneRows())
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("control: the registry was not created by the writer helper: %v", err)
		}
		assertInventory(t, factoryQuotaReadLanes(root), laneTwoThree, 2)
	})

	t.Run("unreadable_yields_none", func(t *testing.T) {
		root := newRoot(t)
		path, err := homestate.FactoryDBPath(root)
		if err != nil {
			t.Fatal(err)
		}
		// Not a database at all.
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("this is not a sqlite database, only text padding it out to some length"), 0o600); err != nil {
			t.Fatal(err)
		}
		assertInventory(t, factoryQuotaReadLanes(root), nil, 0)

		// A real registry whose file cannot be read.
		root2 := newRoot(t)
		qasWriteLanes(t, root2, qasStandardLaneRows())
		path2, _ := homestate.FactoryDBPath(root2)
		if err := os.Chmod(path2, 0o000); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = os.Chmod(path2, 0o600) })
		if _, err := os.ReadFile(path2); err == nil {
			t.Skip("the process can read a mode-000 file (running as a privileged user); the unreadable-file case is not observable here")
		}
		assertInventory(t, factoryQuotaReadLanes(root2), nil, 0)
		_ = os.Chmod(path2, 0o600)
		assertInventory(t, factoryQuotaReadLanes(root2), laneTwoThree, 2) // positive control: readable again
	})

	t.Run("read_only_main_file_and_rows_unchanged", func(t *testing.T) {
		root := newRoot(t)
		qasWriteLanes(t, root, qasStandardLaneRows())
		before := qasRegistryState(t, root)
		if !strings.Contains(before, "meta|[schema_version 5]") || strings.Contains(before, "legacy_workers_imported") {
			t.Fatalf("fixture precondition: the registry must carry schema_version and lack the legacy_workers_imported marker:\n%s", before)
		}
		assertInventory(t, factoryQuotaReadLanes(root), laneTwoThree, 2)
		after := qasRegistryState(t, root)
		if after != before {
			t.Errorf("the registry changed under the inventory read:\nbefore:\n%s\nafter:\n%s", before, after)
		}

		// Read-only directory (sidecar creation impossible): an error, so no
		// candidates and no crash.
		dbPath, _ := homestate.FactoryDBPath(root)
		dir := filepath.Dir(dbPath)
		for _, side := range []string{"factory.db-wal", "factory.db-shm"} {
			_ = os.Remove(filepath.Join(dir, side))
		}
		if err := os.Chmod(dir, 0o500); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = os.Chmod(dir, 0o700) })
		assertInventory(t, factoryQuotaReadLanes(root), nil, 0)
		_ = os.Chmod(dir, 0o700)
		if after := qasRegistryState(t, root); after != before {
			t.Errorf("the registry changed under the read-only-directory read")
		}
	})
}

// ---------------------------------------------------------------------------
// AC-QAS-019 / -020

// AC-QAS-019 — pressure on: `--auto` and `moai factory status` recommend the
// candidate lanes.
func TestQAS_AC019_AutoAndStatusRecommendNonClaudeLanes(t *testing.T) {
	setup := func(t *testing.T) (string, *kanban.BacklogStore) {
		t.Helper()
		qasLaneSeam(t)
		root, store := qasPressureRoot(t, false, 92)
		qasWriteLanes(t, root, qasStandardLaneRows())
		return root, store
	}
	const wantLine = "quota pressure: five_hour used=92.0% resets_at=2026-09-26T12:00:00Z; recommend non-Claude lane(s): lane-2 (glm), lane-3 (gpt)"

	t.Run("auto_line_before_each_accept", func(t *testing.T) {
		root, store := setup(t)
		out := qasAutoRun(t, root, store)
		lines := strings.Split(out, "\n")
		accepts, quotas := 0, 0
		for i, line := range lines {
			if strings.HasPrefix(line, "quota pressure:") {
				quotas++
				if !qasRecommendRE.MatchString(line) || line != wantLine {
					t.Errorf("recommendation line = %q, want %q", line, wantLine)
				}
				if i+1 >= len(lines) || !strings.HasPrefix(lines[i+1], "accept ") {
					t.Errorf("the recommendation is not immediately before an accept line: next = %q", lines[min(i+1, len(lines)-1)])
				}
			}
			if strings.HasPrefix(line, "accept ") {
				accepts++
				if i == 0 || !strings.HasPrefix(lines[i-1], "quota pressure:") {
					t.Errorf("accept line %q is not immediately preceded by the recommendation: previous = %q", line, lines[max(i-1, 0)])
				}
			}
		}
		if accepts != 2 || quotas != 2 {
			t.Errorf("accepts=%d recommendation lines=%d, want 2 and 2 (one per accept)\n%s", accepts, quotas, out)
		}
	})

	t.Run("status_text", func(t *testing.T) {
		setup(t)
		text, _ := qasStatus(t)
		var line string
		for _, l := range strings.Split(text, "\n") {
			if strings.HasPrefix(l, "quota pressure:") {
				if line != "" {
					t.Errorf("more than one quota pressure line:\n%s", text)
				}
				line = l
			}
		}
		if line != wantLine {
			t.Errorf("status recommendation line = %q, want %q\n%s", line, wantLine, text)
		}
		qi, pi := strings.Index(text, "quota five_hour:"), strings.Index(text, "quota pressure:")
		if qi < 0 || pi < qi {
			t.Errorf("the recommendation is not printed under the quota block:\n%s", text)
		}
	})

	t.Run("status_json", func(t *testing.T) {
		setup(t)
		_, js := qasStatus(t)
		quota, ok := qasStatusQuota(t, js)
		if !ok {
			t.Fatal("no quota key")
		}
		if quota["pressure"] != true {
			t.Errorf("pressure = %v, want true", quota["pressure"])
		}
		if quota["unknown_lanes"] != float64(2) {
			t.Errorf("unknown_lanes = %v, want 2", quota["unknown_lanes"])
		}
		want := []any{
			map[string]any{"label": "lane-2", "backend": "glm"},
			map[string]any{"label": "lane-3", "backend": "gpt"},
		}
		if !reflect.DeepEqual(quota["candidates"], want) {
			t.Errorf("candidates = %v, want %v", quota["candidates"], want)
		}
		if _, present := quota["warning"]; present {
			t.Errorf("a warning marker accompanies a recommendation: %v", quota["warning"])
		}
	})
}

// AC-QAS-020 — pressure on and no candidate lane: a warning, nothing else.
func TestQAS_AC020_NoNonClaudeLaneWarnsOnly(t *testing.T) {
	cases := []struct {
		name        string
		rows        []qasRegistryLane
		wantUnknown int
	}{
		{"only_claude_lanes", []qasRegistryLane{{label: "lane-1", pid: 5001, backend: kanban.BackendClaude}, {label: "lane-2", pid: 5002, backend: kanban.BackendClaude}}, 0},
		{"only_dead_lanes", []qasRegistryLane{{label: "lane-2", pid: qasDeadPID, backend: kanban.BackendGLM}, {label: "lane-3", pid: qasDeadPID, backend: kanban.BackendGPT}}, 0},
		{"no_lanes", nil, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			qasLaneSeam(t)
			root, store := qasPressureRoot(t, false, 92)
			if c.rows != nil {
				qasWriteLanes(t, root, c.rows)
			}
			out := qasAutoRun(t, root, store)
			wantWarning := fmt.Sprintf("quota pressure: five_hour used=92.0%% resets_at=2026-09-26T12:00:00Z; warning: no live non-Claude lane (unknown backend: %d); nothing is re-dispatched", c.wantUnknown)
			warnings := 0
			for _, line := range strings.Split(out, "\n") {
				if strings.HasPrefix(line, "quota pressure:") {
					warnings++
					if !qasWarningRE.MatchString(line) || line != wantWarning {
						t.Errorf("warning line = %q, want %q", line, wantWarning)
					}
				}
			}
			if warnings != 2 {
				t.Errorf("warning lines = %d, want 2 (one per accept)\n%s", warnings, out)
			}
			if strings.Contains(out, "recommend non-Claude") {
				t.Errorf("a recommendation accompanies the warning:\n%s", out)
			}

			text, js := qasStatus(t)
			if !strings.Contains(text, "\n"+wantWarning+"\n") {
				t.Errorf("status text lacks the warning line %q:\n%s", wantWarning, text)
			}
			quota, ok := qasStatusQuota(t, js)
			if !ok {
				t.Fatal("no quota key")
			}
			if quota["warning"] != "no-non-claude-lane" || quota["pressure"] != true {
				t.Errorf("quota warning/pressure = %v / %v", quota["warning"], quota["pressure"])
			}
			if cands, present := quota["candidates"]; !present || !reflect.DeepEqual(cands, []any{}) {
				t.Errorf("candidates = %v (present=%v), want an empty list", cands, present)
			}
		})
	}

	t.Run("rest_of_output_unchanged", func(t *testing.T) {
		qasLaneSeam(t)
		rootOn, storeOn := qasPressureRoot(t, false, 92)
		on := qasAutoRun(t, rootOn, storeOn)
		rootOff, storeOff := qasPressureRoot(t, false, 50)
		off := qasAutoRun(t, rootOff, storeOff)
		var kept []string
		for _, line := range strings.Split(on, "\n") {
			if !strings.HasPrefix(line, "quota pressure:") {
				kept = append(kept, line)
			}
		}
		if got := strings.Join(kept, "\n"); got != off {
			t.Errorf("with the warning lines removed the pressure-on output differs from the pressure-off output:\non:\n%s\noff:\n%s", got, off)
		}
		if off != qasGoldenAuto(t) {
			t.Errorf("the pressure-off output differs from the baseline golden:\n%s", off)
		}
	})
}

// ---------------------------------------------------------------------------
// AC-QAS-021

// AC-QAS-021 — the steering changes nothing.
func TestQAS_AC021_SteeringChangesNothing(t *testing.T) {
	t.Run("status_is_read_only", func(t *testing.T) {
		qasLaneSeam(t)
		root, _ := qasPressureRoot(t, true, 92)
		qasWriteLanes(t, root, qasStandardLaneRows())
		sessDir := filepath.Dir(kanban.RecordPath(root, "sess-guard"))
		if err := kanban.Write(root, kanban.NewRecord("sess-guard", "", kanban.BackendClaude).WithRole("lane").WithLane(1)); err != nil {
			t.Fatal(err)
		}
		cardsBefore := qasRecordDump(t, root) // opens the record read-write: measured before the byte fingerprint
		registryBefore := qasRegistryState(t, root)
		sessBefore := qasTreeSnapshot(t, sessDir)
		queueBefore := sdQueueBytes(t, kanban.NewBacklogStore(todoBacklogPath(root)))

		_, js := qasStatus(t)
		quota, ok := qasStatusQuota(t, js)
		if !ok || quota["pressure"] != true {
			t.Fatalf("control: the steering did not run (quota = %v, present = %v)", quota, ok)
		}
		if after := qasRegistryState(t, root); after != registryBefore {
			t.Errorf("the registry changed under status:\nbefore:\n%s\nafter:\n%s", registryBefore, after)
		}
		if after := qasTreeSnapshot(t, sessDir); after != sessBefore {
			t.Errorf("the session registry changed under status")
		}
		if after := sdQueueBytes(t, kanban.NewBacklogStore(todoBacklogPath(root))); after != queueBefore {
			t.Errorf("the queue changed under status")
		}
		if after := qasRecordDump(t, root); after != cardsBefore {
			t.Errorf("a card row changed under status:\nbefore:\n%s\nafter:\n%s", cardsBefore, after)
		}
	})

	t.Run("auto_mutations_equal_pressure_off", func(t *testing.T) {
		qasLaneSeam(t)
		run := func(five float64) (string, string, string) {
			root, store := qasPressureRoot(t, false, five)
			qasWriteLanes(t, root, qasStandardLaneRows())
			regBefore := qasRegistryState(t, root)
			out := qasAutoRun(t, root, store)
			if after := qasRegistryState(t, root); after != regBefore {
				t.Errorf("five=%v: the registry changed under the --auto cycle:\nbefore:\n%s\nafter:\n%s", five, regBefore, after)
			}
			raw, err := os.ReadFile(store.EnginePath())
			if err != nil {
				t.Fatal(err)
			}
			rec, err := store.LoadPure()
			if err != nil {
				t.Fatal(err)
			}
			var states []string
			for _, it := range rec.Items {
				states = append(states, fmt.Sprintf("%s=%s", it.ID, it.State))
			}
			return out, strings.Join(states, ","), string(raw)
		}
		onOut, onStates, _ := run(92)
		offOut, offStates, _ := run(50)
		if onStates != offStates {
			t.Errorf("queue states differ between the pressure-on and pressure-off cycles: on=%s off=%s", onStates, offStates)
		}
		if !strings.Contains(onOut, "quota pressure:") || strings.Contains(offOut, "quota pressure:") {
			t.Errorf("control: the steering line must appear with pressure on and only then")
		}
		for _, verb := range []string{"accept", "done", "evidence collected"} {
			if strings.Count(onOut, verb+" ") != strings.Count(offOut, verb+" ") {
				t.Errorf("the %q lines differ between the pressure-on and pressure-off cycles", verb)
			}
		}
		if strings.Contains(onOut, "unpick") || strings.Contains(onOut, "lease") {
			t.Errorf("the pressure-on cycle unpicked or leased a card:\n%s", onOut)
		}
	})

	t.Run("no_card_class_read", func(t *testing.T) {
		// Static: the steering code never references a card's classification.
		for _, file := range []string{"factory_quota_lanes.go"} {
			for ident := range qasCodeIdents(t, file) {
				for _, frag := range []string{"Classif", "ClassMode", "ClassPriority", "LoadPure", "BacklogRecord", "BacklogItem", "BacklogStore"} {
					if strings.Contains(ident, frag) {
						t.Errorf("%s references %s; the steering never looks at the card or the queue", file, ident)
					}
				}
			}
		}
		// Behavioural: two cards of different classes get the identical line.
		qasLaneSeam(t)
		root, store := qasPressureRoot(t, false, 92)
		qasWriteLanes(t, root, qasStandardLaneRows())
		fcQueue(t, store, kanban.BacklogStateQueued, kanban.BacklogStateQueued)
		fcClassify(t, store, "t1", kanban.ClassPriorityHigh, false, kanban.ClassModeSerial)
		fcClassify(t, store, "t2", kanban.ClassPriorityNormal, true, kanban.ClassModeParallelizable)
		first := todoAutoQuotaLine(root)
		fcClassify(t, store, "t1", kanban.ClassPriorityNormal, true, kanban.ClassModeParallelizable)
		if second := todoAutoQuotaLine(root); first == "" || first != second {
			t.Errorf("the line depends on the queue's classification: %q vs %q", first, second)
		}
	})
}

// ---------------------------------------------------------------------------
// AC-QAS-022

// qasStreams is one run's observable outcome.
type qasStreams struct {
	stdout, stderr string
	exit           int
}

// AC-QAS-022 — gate disabled or pressure off: the output is unchanged.
func TestQAS_AC022_GateOffOrPressureOffOutputUnchanged(t *testing.T) {
	autoFixtureRoot := func(t *testing.T) (string, *kanban.BacklogStore) {
		t.Helper()
		sdClearLaneEnv(t)
		return sdMoaiFixture(t)
	}
	autoEquals := func(t *testing.T, name string, prepare func(t *testing.T, root string)) {
		t.Helper()
		root, store := autoFixtureRoot(t)
		prepare(t, root)
		if got := qasAutoRun(t, root, store); got != qasGoldenAuto(t) {
			t.Errorf("%s: `moai todo --auto` differs from the baseline golden:\n%s", name, got)
		}
	}

	t.Run("status_gate_disabled_with_data", func(t *testing.T) {
		for _, key := range []string{"empty", "one_card"} {
			root, _ := qasStatusRoot(t, key == "one_card")
			qasWriteRecord(t, root, "sess-live", fcNow, qasWin(95, qasReset5), qasWin(97, qasReset7))
			_, js := qasStatus(t)
			if js != qasGoldenStatus(t, key) {
				t.Errorf("%s: gate disabled, data present: status --json differs from the golden:\n%s", key, js)
			}
		}
	})
	t.Run("auto_gate_disabled", func(t *testing.T) {
		autoEquals(t, "gate disabled with a 99% record", func(t *testing.T, root string) {
			qasWriteRecord(t, root, "sess-live", fcNow, qasWin(99, qasReset5), qasWin(99, qasReset7))
		})
	})
	t.Run("auto_below_threshold", func(t *testing.T) {
		autoEquals(t, "gate enabled at 89.9%", func(t *testing.T, root string) {
			qasEnableGate(t, root)
			qasWriteRecord(t, root, "sess-live", fcNow, qasWin(89.9, qasReset5), qasWin(94.9, qasReset7))
		})
		autoEquals(t, "gate enabled, windows reset", func(t *testing.T, root string) {
			qasEnableGate(t, root)
			qasWriteRecord(t, root, "sess-live", fcNow, qasWin(99, fcNow.Add(-time.Second).Unix()), qasWin(99, fcNow.Add(-time.Second).Unix()))
		})
	})
	t.Run("auto_unknown_windows", func(t *testing.T) {
		autoEquals(t, "gate enabled, no record", func(t *testing.T, root string) { qasEnableGate(t, root) })
		autoEquals(t, "gate enabled, stale record", func(t *testing.T, root string) {
			qasEnableGate(t, root)
			qasWriteRecord(t, root, "sess-old", fcNow.Add(-2*time.Hour), qasWin(99, qasReset5), nil)
		})
	})

	t.Run("next_pressure_off_equals_gate_disabled", func(t *testing.T) {
		run := func(o qasFixtureOpts) qasStreams {
			qasFixture(t, o)
			out, errOut, err := qasRunNext(t, "--run", fcRun)
			s := qasStreams{stdout: out, stderr: errOut}
			if err != nil {
				code, _ := ResolveExitCode(err)
				s.exit = code
				if code == 0 {
					s.exit = 1
				}
			}
			return s
		}
		off := run(qasFixtureOpts{gateOff: true, five: qasWin(89.9, qasReset5)})
		on := run(qasFixtureOpts{five: qasWin(89.9, qasReset5)})
		if off != on {
			t.Errorf("`moai factory next` differs between the gate disabled and the gate enabled with pressure off:\ndisabled: %+v\nenabled:  %+v", off, on)
		}
		if off.exit != 0 || off.stdout == "" {
			t.Errorf("control: the disabled-gate run did not lease t1: %+v", off)
		}
	})

	t.Run("acquire_pressure_off_equals_gate_disabled", func(t *testing.T) {
		stamp := regexp.MustCompile(`[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9:.]+(Z|[+-][0-9:]+)?`)
		run := func(enable bool) qasStreams {
			sdClearLaneEnv(t)
			t.Setenv(config.EnvMoaiLaunchProvider, kanban.BackendClaude)
			root := t.TempDir()
			if enable {
				qasEnableGate(t, root)
				qasWriteRecord(t, root, "sess-live", time.Now(), qasWin(40, time.Now().Add(3*time.Hour).Unix()), nil)
			}
			stdout, stderr, err := runIntegrationStreams(t, root, "acquire", "--session", "sess-qas", "--name", "lane-qas", "--card", "t1347")
			s := qasStreams{stdout: stamp.ReplaceAllString(stdout, "<T>"), stderr: stamp.ReplaceAllString(stderr, "<T>")}
			if err != nil {
				s.exit = 1
			}
			return s
		}
		off, on := run(false), run(true)
		if off != on {
			t.Errorf("`moai integration acquire` differs between the gate disabled and the gate enabled with pressure off:\ndisabled: %+v\nenabled:  %+v", off, on)
		}
		if off.exit != 0 || off.stdout == "" {
			t.Errorf("control: the disabled-gate acquire did not succeed with output: %+v", off)
		}
	})

	t.Run("no_steering_line_when_pressure_off", func(t *testing.T) {
		qasLaneSeam(t)
		root, store := qasPressureRoot(t, false, 89.9)
		qasWriteLanes(t, root, qasStandardLaneRows())
		text, js := qasStatus(t)
		if strings.Contains(text, "quota pressure") || strings.Contains(text, "quota hold") || strings.Contains(js, "recommend") {
			t.Errorf("a steering line appeared with pressure off:\n%s\n%s", text, js)
		}
		out := qasAutoRun(t, root, store)
		if strings.Contains(out, "quota pressure") || strings.Contains(out, "quota hold") {
			t.Errorf("--auto printed a steering line with pressure off:\n%s", out)
		}
		if todoAutoQuotaLine(root) != "" {
			t.Errorf("the --auto seam returned a line with pressure off")
		}
		// Control: the same fixture at 92% does steer (the assertions above are not vacuous).
		qasWriteRecord(t, root, "sess-live", fcNow, qasWin(92, qasReset5), nil)
		if todoAutoQuotaLine(root) == "" {
			t.Errorf("control: no steering at 92%%")
		}
	})
}
