package cli

import (
	"context"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// REQ-FCR-025 (card t1963, ledger L11, AC-FCR-017 (11)) at the alarm level: a
// card a consolidation moved into the live run keeps the stage anchor its
// source-run transitions recorded, so the stall alarm measures its real age
// instead of falling back to a fresh updated_at — both when the status names
// the run and on the default (no --run) path, which reads every run's events.

// caaSeedMovedCard seeds a dead run-old holding the card "moved" (created and
// entered run 10h before fcNow, row touched 10 minutes before fcNow, so only
// the event history says it is stalled), consolidates it into the live
// run-new, and closes the store.
func caaSeedMovedCard(t *testing.T, root string) {
	t.Helper()
	ctx := context.Background()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	for _, r := range []homestate.FactoryRun{
		{RunID: "run-old", ManifestJSON: "{}", LeadPID: 20, LeadProcessStart: "fp"},
		{RunID: "run-new", ManifestJSON: "{}", LeadPID: 21, LeadProcessStart: "fp"},
	} {
		if err := db.RecordRun(ctx, r); err != nil {
			t.Fatal(err)
		}
	}
	stamp := func(d time.Duration) string { return fcNow.Add(-d).Format(time.RFC3339Nano) }
	if _, err := db.DB.Exec(`INSERT INTO cards(run_id,card_id,state,version,updated_at) VALUES(?,?,?,?,?)`,
		"run-old", "moved", homestate.CardRun, 3, stamp(10*time.Minute)); err != nil {
		t.Fatal(err)
	}
	for _, ev := range []struct {
		from, to string
		ago      time.Duration
	}{{"", homestate.CardPicked, 11 * time.Hour}, {homestate.CardPicked, homestate.CardRun, 10 * time.Hour}} {
		if _, err := db.DB.Exec(`INSERT INTO events(run_id,kind,payload_json,created_at) VALUES('run-old','card.transition',json_object('card_id','moved','from',?,'to',?),?)`,
			ev.from, ev.to, stamp(ev.ago)); err != nil {
			t.Fatal(err)
		}
	}
	opts := homestate.ReconcileOptions{
		Classify: func(pid int, _ string) homestate.OwnerClassification {
			if pid == 20 {
				return homestate.OwnerDead
			}
			return homestate.OwnerLive
		},
		WorkerProbe: func(int) homestate.ProcessIdentityState { return homestate.ProcessIdentityDead },
	}
	owners, err := db.ClassifyRuns(ctx, opts)
	if err != nil {
		t.Fatal(err)
	}
	var src, tgt homestate.RunOwner
	for _, o := range owners {
		switch o.RunID {
		case "run-old":
			src = o
		case "run-new":
			tgt = o
		}
	}
	res, err := db.RebindRun(ctx, homestate.RebindRequest{Source: src, Target: tgt, RequireStampBasis: true}, opts)
	if err != nil {
		t.Fatalf("rebind: %v", err)
	}
	if !slices.Equal(res.Moved, []string{"moved"}) {
		t.Fatalf("rebind result %+v, want moved [moved]", res)
	}
}

func TestFactoryAgeAlarmMovedCardKeepsSourceStageAnchor(t *testing.T) {
	_, _ = fcFixture(t)
	root := factoryCardRoot()
	caaSeedMovedCard(t, root)
	db := fcOpen(t, root)
	ctx := context.Background()
	card, err := db.LoadCard(ctx, "run-new", "moved")
	if err != nil {
		t.Fatal(err)
	}
	bands := config.CardAgeAlarmSettings{
		Enabled:  true,
		GateWarn: config.DefaultCardAgeAlarmGateWarnAfter, GateAlert: config.DefaultCardAgeAlarmGateAlertAfter,
		StallWarn: config.DefaultCardAgeAlarmStallWarnAfter, StallAlert: config.DefaultCardAgeAlarmStallAlertAfter,
	}
	rows, warning := computeFactoryAgeAlarmsWithSettings(ctx, "run-new", db, []homestate.Card{card}, fcNow, bands)
	if warning != "" {
		t.Fatalf("alarm warning: %s", warning)
	}
	if len(rows) != 1 {
		t.Fatalf("alarm rows = %+v, want exactly one stall row for the moved card (L11: its source-run history was lost)", rows)
	}
	r := rows[0]
	if r.Card != "moved" || r.Run != "run-new" || r.Class != homestate.AgeClassNoProgress || r.Band != "alert" {
		t.Fatalf("alarm row = %+v, want moved/run-new/no-progress/alert", r)
	}
	if r.Stage != (10 * time.Hour).String() {
		t.Fatalf("stage age = %q, want %q (anchored on the source-run run transition, not unknown)", r.Stage, (10 * time.Hour).String())
	}
}

// The default `factory status` (no --run) loads the events of every run; the
// moved card must still be grouped under its new run (REQ-FCR-025, all-runs
// path).
func TestFactoryStatusDefaultPathAgeAlarmMovedCard(t *testing.T) {
	_, _ = fcFixture(t)
	root := factoryCardRoot()
	caaSeedMovedCard(t, root)
	caaEnableAlarm(t, root, "")
	out, _, err := runFactory(t, "status")
	if err != nil {
		t.Fatalf("factory status: %v", err)
	}
	want := "age-alarm card=moved run=run-new state=run class=no-progress band=alert age=10h0m0s" +
		" stage=10h0m0s commit=unknown activity=unknown"
	found := false
	for _, l := range strings.Split(out, "\n") {
		if l == want {
			found = true
		}
	}
	if !found {
		t.Fatalf("default status has no alarm line\n  %s\ngot:\n%s", want, out)
	}
}
