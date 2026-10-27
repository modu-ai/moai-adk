package cli

// factory_runs_retire_guard_test.go — SPEC-FACTORY-RETIRE-GUARD-001 M2
// (card t1852): the `moai factory runs --retire` refusal and stranded-cards
// surface, and the `relaunch --from-run` outcome.

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"
	"unicode"

	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/spf13/pflag"
)

type failingReader struct{ t *testing.T }

func (r failingReader) Read([]byte) (int, error) {
	r.t.Errorf("the --retire command read from stdin")
	return 0, errors.New("stdin must not be read")
}

func seedCardRows(t *testing.T, root, runID string, cards ...[3]string) {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	defer func() { _ = db.Close() }()
	for _, c := range cards {
		if _, err := db.DB.Exec(`INSERT INTO cards(run_id,card_id,owner_label,state,updated_at) VALUES(?,?,?,?,?)`,
			runID, c[0], c[2], c[1], "2026-10-14T00:00:00Z"); err != nil {
			t.Fatalf("insert card %s: %v", c[0], err)
		}
	}
}

func markRetired(t *testing.T, root, runID string) {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	defer func() { _ = db.Close() }()
	if _, err := db.DB.Exec(`UPDATE runs SET status='retired' WHERE run_id=?`, runID); err != nil {
		t.Fatalf("mark retired: %v", err)
	}
	if _, err := db.DB.Exec(`INSERT INTO events(run_id,kind,payload_json,created_at) VALUES(?,'run.retired','{}','2026-10-14T00:00:00Z')`, runID); err != nil {
		t.Fatalf("seed event: %v", err)
	}
}

// factoryDump renders the runs, cards and events tables so a before/after
// comparison covers every column.
func factoryDump(t *testing.T, root string) string {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	defer func() { _ = db.Close() }()
	var out []string
	for _, q := range []string{`SELECT * FROM runs ORDER BY run_id`, `SELECT * FROM cards ORDER BY run_id,card_id`, `SELECT * FROM events ORDER BY seq`} {
		rows, err := db.DB.Query(q)
		if err != nil {
			t.Fatalf("dump %q: %v", q, err)
		}
		cols, _ := rows.Columns()
		for rows.Next() {
			vals := make([]any, len(cols))
			ptrs := make([]any, len(cols))
			for i := range vals {
				ptrs[i] = &vals[i]
			}
			if err := rows.Scan(ptrs...); err != nil {
				t.Fatalf("scan: %v", err)
			}
			out = append(out, fmt.Sprint(vals...))
		}
		_ = rows.Close()
		out = append(out, "--")
	}
	return strings.Join(out, "\n")
}

func retiredEventCount(t *testing.T, root, runID string) int {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	defer func() { _ = db.Close() }()
	var n int
	if err := db.DB.QueryRow(`SELECT count(*) FROM events WHERE run_id=? AND kind='run.retired'`, runID).Scan(&n); err != nil {
		t.Fatalf("count events: %v", err)
	}
	return n
}

func hasHangul(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Hangul, r) {
			return true
		}
	}
	return false
}

// AC-RRG-005 — an active dead-owner run holding cards is refused, in every
// session, with the cards, the exits and the abandon caveats; nothing is read
// from stdin and nothing changes.
func TestFactoryRunsRetireRefusesInProgressCards(t *testing.T) {
	root := runOwnerSandbox(t)
	t.Chdir(root)
	deadPID, deadStart := deadOwnerIdentity(t)
	seedRun(t, root, "runR", deadPID, deadStart)
	seedCardRows(t, root, "runR", [3]string{"cardA", "needs-decision", "lane-1"}, [3]string{"cardB", "picked", ""})
	before := factoryDump(t, root)

	for _, session := range []struct{ name, worker string }{{"operator", ""}, {"lane", "lane-1"}} {
		t.Run(session.name, func(t *testing.T) {
			t.Setenv("MOAI_FACTORY_WORKER", session.worker)
			cmd := newFactoryCommand()
			out := &bytes.Buffer{}
			cmd.SetOut(out)
			cmd.SetErr(out)
			cmd.SetIn(failingReader{t})
			cmd.SetArgs([]string{"runs", "--retire", "runR"})
			err := cmd.Execute()
			if !errors.Is(err, homestate.ErrRunHasInProgressCards) {
				t.Fatalf("err = %v, want errors.Is ErrRunHasInProgressCards", err)
			}
			text := out.String()
			for _, want := range []string{
				fmt.Sprintf(retireRefusalHeaderFmt, "runR", 2),
				fmt.Sprintf(retireOwnerLineFmt, "dead"),
				fmt.Sprintf(retireCardLineFmt, "cardA", "needs-decision", "lane-1"),
				fmt.Sprintf(retireCardLineFmt, "cardB", "picked", "-"),
				"moai cc -l --factory-run runR",
				"moai cc -f --factory-run runR",
				"moai factory decide cardA --choice abandon --run runR",
				"moai factory decide cardB --choice abandon --run runR",
				retireAbandonNote,
			} {
				if !strings.Contains(text, want) {
					t.Errorf("output lacks %q:\n%s", want, text)
				}
			}
			if !hasHangul(strings.SplitN(text, "\n", 2)[0]) {
				t.Errorf("refusal header carries no Hangul: %q", text)
			}
			if strings.HasPrefix(text, "retired ") {
				t.Errorf("output starts with a success line: %q", text)
			}
			if got := runStatusOf(t, root, "runR"); got != "active" {
				t.Errorf("status = %q, want active", got)
			}
			if after := factoryDump(t, root); after != before {
				t.Errorf("the refusal changed state:\nbefore %s\nafter  %s", before, after)
			}
		})
	}
}

// AC-RRG-006 — a retired run's stranded cards are reported on success.
func TestFactoryRunsRetireReportsStrandedCardsOfRetiredRun(t *testing.T) {
	root := runOwnerSandbox(t)
	t.Chdir(root)
	deadPID, deadStart := deadOwnerIdentity(t)
	seedRun(t, root, "runA", deadPID, deadStart)
	seedCardRows(t, root, "runA", [3]string{"cardA", "needs-decision", "lane-1"}, [3]string{"cardB", "picked", ""})
	markRetired(t, root, "runA")
	seedRun(t, root, "runB", deadPID, deadStart)
	seedCardRows(t, root, "runB", [3]string{"cardD", "done", "lane-1"})
	markRetired(t, root, "runB")
	seedRun(t, root, "runC", deadPID, deadStart)

	t.Run("A retired run holding cards", func(t *testing.T) {
		before := factoryDump(t, root)
		out, err := runFactoryCommand(t, "runs", "--retire", "runA")
		if err != nil {
			t.Fatalf("err = %v, want success", err)
		}
		if !strings.HasPrefix(out, "retired runA (owner dead)\n") {
			t.Errorf("first line wrong: %q", out)
		}
		for _, want := range []string{
			fmt.Sprintf(retireStrandedHdrFmt, "runA", 2),
			fmt.Sprintf(retireCardLineFmt, "cardA", "needs-decision", "lane-1"),
			fmt.Sprintf(retireCardLineFmt, "cardB", "picked", "-"),
			"moai factory decide cardA --choice abandon --run runA",
			"moai factory decide cardB --choice abandon --run runA",
			retireStrandedNote,
			retireAbandonNote,
		} {
			if !strings.Contains(out, want) {
				t.Errorf("output lacks %q:\n%s", want, out)
			}
		}
		for _, banned := range []string{"moai cc -l", "moai cc -f", "남은 방법", "유일", "only"} {
			if strings.Contains(out, banned) {
				t.Errorf("output contains %q:\n%s", banned, out)
			}
		}
		if after := factoryDump(t, root); after != before {
			t.Errorf("a no-op wrote:\nbefore %s\nafter  %s", before, after)
		}
	})
	t.Run("B retired run, terminal cards only", func(t *testing.T) {
		out, err := runFactoryCommand(t, "runs", "--retire", "runB")
		if err != nil || out != "retired runB (owner dead)\n" {
			t.Fatalf("out %q err %v, want the first line only", out, err)
		}
	})
	t.Run("C active card-free run", func(t *testing.T) {
		out, err := runFactoryCommand(t, "runs", "--retire", "runC")
		if err != nil || out != "retired runC (owner dead)\n" {
			t.Fatalf("out %q err %v, want the first line only", out, err)
		}
		if got := runStatusOf(t, root, "runC"); got != "retired" {
			t.Errorf("status = %q, want retired", got)
		}
		if n := retiredEventCount(t, root, "runC"); n != 1 {
			t.Errorf("run.retired events = %d, want 1", n)
		}
	})
}

func TestFactoryRunsHasNoForceFlag(t *testing.T) {
	cmd := newFactoryRunsCommand()
	var names []string
	cmd.Flags().VisitAll(func(f *pflag.Flag) { names = append(names, f.Name) })
	sort.Strings(names)
	// SPEC-FACTORY-COMPLETION-RECOVERY-001 AC-FCR-022 (f), decision Q1 (a): the
	// one widening this card is allowed — `--rebind-to` joins `--retire`. The
	// assertion keeps its form (the sorted set, exact), so a bypass flag such as
	// `--force` still fails it.
	if !reflect.DeepEqual(names, []string{"rebind-to", "retire"}) {
		t.Fatalf("flags = %v, want exactly [rebind-to retire]", names)
	}
}
