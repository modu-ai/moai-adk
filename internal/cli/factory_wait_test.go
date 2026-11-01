package cli

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

func waitTestDB(t *testing.T) *homestate.FactoryDB {
	t.Helper()
	db, err := homestate.OpenFactoryPath(filepath.Join(t.TempDir(), "factory.db"))
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func wakeLines(out string) []string {
	var lines []string
	for _, l := range strings.Split(out, "\n") {
		if strings.HasPrefix(l, "wake: ") {
			lines = append(lines, l)
		}
	}
	return lines
}

// TestFactoryWakeOnLanding drives the wait record end to end at the layer the
// complete verb calls: record a wait, land (or not) a card, count wake lines.
// No live session, no send: the verb only prints.
func TestFactoryWakeOnLanding(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)
	expired := now.Add(-time.Hour).Format(time.RFC3339Nano)

	cases := []struct {
		name      string
		wait      homestate.CardWait
		landing   homestate.Card
		wantWakes int
		wantNote  bool
		reruns    int // extra prints after the first; each must add 0 wake lines
	}{
		{
			name:      "predecessor landed local, lane waker: one wake, rerun silent",
			wait:      homestate.CardWait{CardID: "t2020", CondKind: homestate.WaitCondPredecessorLanded, CondRef: "t2010", Waker: "lane-3", Reason: "needs t2010 merged"},
			landing:   homestate.Card{CardID: "t2010", State: homestate.CardMergedLocal},
			wantWakes: 1, reruns: 2,
		},
		{
			name:      "predecessor landed by pull request, leader waker",
			wait:      homestate.CardWait{CardID: "t2020", CondKind: homestate.WaitCondPredecessorLanded, CondRef: "t2010", Waker: homestate.WaitWakerLeader, Reason: "r"},
			landing:   homestate.Card{CardID: "t2010", State: homestate.CardMergedPR},
			wantWakes: 1, reruns: 1,
		},
		{
			name:      "a different card landed: condition not met, nothing",
			wait:      homestate.CardWait{CardID: "t2020", CondKind: homestate.WaitCondPredecessorLanded, CondRef: "t2010", Waker: "lane-3", Reason: "r"},
			landing:   homestate.Card{CardID: "t2011", State: homestate.CardMergedLocal},
			wantWakes: 0,
		},
		{
			name:      "predecessor only has an expired lease, not landed: nothing",
			wait:      homestate.CardWait{CardID: "t2020", CondKind: homestate.WaitCondPredecessorLanded, CondRef: "t2010", Waker: "lane-3", Reason: "r"},
			landing:   homestate.Card{CardID: "t2010", State: homestate.CardLeased, LeaseHolder: "lane-1", LeaseExpiresAt: expired},
			wantWakes: 0,
		},
		{
			name:      "operator-ruling wait: landing the ref card prints no lane wake",
			wait:      homestate.CardWait{CardID: "t2020", CondKind: homestate.WaitCondOperatorRuling, CondRef: "", Waker: homestate.WaitWakerOperator, Reason: "ruling on scope"},
			landing:   homestate.Card{CardID: "t2010", State: homestate.CardMergedLocal},
			wantWakes: 0,
		},
		{
			name:      "operator waker on a predecessor wait: no direct lane wake, leader-facing note once",
			wait:      homestate.CardWait{CardID: "t2020", CondKind: homestate.WaitCondPredecessorLanded, CondRef: "t2010", Waker: homestate.WaitWakerOperator, Reason: "r"},
			landing:   homestate.Card{CardID: "t2010", State: homestate.CardMergedLocal},
			wantWakes: 0, wantNote: true, reruns: 1,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := waitTestDB(t)
			if err := db.RecordCardWait(ctx, tc.wait, now); err != nil {
				t.Fatalf("record wait: %v", err)
			}
			var first bytes.Buffer
			factoryPrintWakes(ctx, &first, db, tc.landing)
			if got := len(wakeLines(first.String())); got != tc.wantWakes {
				t.Fatalf("first print: want %d wake line(s), got %d:\n%s", tc.wantWakes, got, first.String())
			}
			if tc.wantWakes == 1 {
				line := wakeLines(first.String())[0]
				for _, want := range []string{"lane=" + tc.wait.Waker, "card=" + tc.wait.CardID, "reason="} {
					if !strings.Contains(line, want) {
						t.Errorf("wake line %q lacks %q", line, want)
					}
				}
				if !strings.Contains(first.String(), "moai factory status") || !strings.Contains(first.String(), "progress") {
					t.Errorf("wake output must point at lease-holder and progress record reads:\n%s", first.String())
				}
			}
			if tc.wantNote != strings.Contains(first.String(), "note: ") {
				t.Errorf("note presence: want %v, output:\n%s", tc.wantNote, first.String())
			}
			for i := 0; i < tc.reruns; i++ {
				var again bytes.Buffer
				factoryPrintWakes(ctx, &again, db, tc.landing)
				if strings.TrimSpace(again.String()) != "" {
					t.Fatalf("rerun %d must print nothing, got:\n%s", i+1, again.String())
				}
			}
		})
	}
}

func TestFactoryWakeTwoWaitsOnePredecessor(t *testing.T) {
	ctx := context.Background()
	db := waitTestDB(t)
	now := time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)
	for _, id := range []string{"t2020", "t2021"} {
		if err := db.RecordCardWait(ctx, homestate.CardWait{CardID: id, CondKind: homestate.WaitCondPredecessorLanded, CondRef: "t2010", Waker: "lane-" + id[len(id)-1:], Reason: "r"}, now); err != nil {
			t.Fatal(err)
		}
	}
	var out bytes.Buffer
	factoryPrintWakes(ctx, &out, db, homestate.Card{CardID: "t2010", State: homestate.CardMergedPR})
	if got := len(wakeLines(out.String())); got != 2 {
		t.Fatalf("want 2 wake lines, got %d:\n%s", got, out.String())
	}
}

// runFactoryWaitCmd drives the real verb against a temp project root and
// returns its combined output and error.
func runFactoryWaitCmd(t *testing.T, root string, args ...string) (string, error) {
	t.Helper()
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	cmd := newFactoryWaitCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func openWaitRoot(t *testing.T, root string) *homestate.FactoryDB {
	t.Helper()
	db, err := homestate.OpenFactory(factoryRootOf(root))
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func storedWaitReason(t *testing.T, root, card string) string {
	t.Helper()
	waits, err := openWaitRoot(t, root).LoadCardWaits(context.Background(), card)
	if err != nil || len(waits) != 1 {
		t.Fatalf("want one wait, got %+v err=%v", waits, err)
	}
	return waits[0].Reason
}

// A wait recorded with --scope stores scope=<value> at the front of the reason
// and shows it in both the record line and the wake line at the landing edge.
func TestFactoryWaitScopeStoredAndPrinted(t *testing.T) {
	root := t.TempDir()
	scope := "Write internal/x.go denied by worktree guard"
	out, err := runFactoryWaitCmd(t, root, "t2020", "--on", "t2010", "--waker", "lane-3", "--reason", "needs t2010", "--scope", scope)
	if err != nil {
		t.Fatalf("wait with scope: %v\n%s", err, out)
	}
	if !strings.Contains(out, "scope="+scope) {
		t.Errorf("record output lacks scope=%s:\n%s", scope, out)
	}
	want := "scope=" + scope + " needs t2010"
	if got := storedWaitReason(t, root, "t2020"); got != want {
		t.Errorf("stored reason = %q, want %q", got, want)
	}
	var wake bytes.Buffer
	factoryPrintWakes(context.Background(), &wake, openWaitRoot(t, root), homestate.Card{CardID: "t2010", State: homestate.CardMergedPR})
	lines := wakeLines(wake.String())
	if len(lines) != 1 || !strings.Contains(lines[0], "reason="+want) {
		t.Errorf("wake line must carry reason=%s, got %q", want, lines)
	}
}

// Without --scope the call behaves exactly as before.
func TestFactoryWaitWithoutScopeUnchanged(t *testing.T) {
	root := t.TempDir()
	out, err := runFactoryWaitCmd(t, root, "t2020", "--on", "t2010", "--waker", "lane-3", "--reason", "needs t2010")
	if err != nil {
		t.Fatalf("wait without scope: %v\n%s", err, out)
	}
	if want := "wait recorded: card=t2020 condition=predecessor-landed t2010 waker=lane-3\n"; out != want {
		t.Errorf("output = %q, want %q", out, want)
	}
	if got := storedWaitReason(t, root, "t2020"); got != "needs t2010" {
		t.Errorf("stored reason = %q, want %q", got, "needs t2010")
	}
}

// A given --scope must be one trimmed non-empty line.
func TestFactoryWaitScopeRejectsMultilineAndEmpty(t *testing.T) {
	for _, scope := range []string{"line one\nline two", "line one\rline two", "", "   "} {
		out, err := runFactoryWaitCmd(t, t.TempDir(), "t2020", "--on", "t2010", "--waker", "lane-3", "--reason", "r", "--scope", scope)
		if !errors.Is(err, homestate.ErrInvalidCardInput) {
			t.Errorf("scope %q: want ErrInvalidCardInput, got err=%v out=%q", scope, err, out)
		}
	}
}
