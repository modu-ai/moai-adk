package web

// redacted_badge_test.go — the "masked" badge follows every text the refiner
// can change (Issue #1948, B8). A row that shows [REDACTED] must say so, whichever
// field held the sensitive text; and the fields that are normally hex (commit,
// head, hash) are refined too, before they are cut short.
//
// Defect answer: these tests catch a masked value whose row carries no badge, a
// field left out of the badge's list, and a cut that keeps the first characters
// of an e-mail address or a path.

import (
	"context"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/ctxmodel"
)

func TestRedactedBadge_IntakeRequestFields(t *testing.T) {
	secret := screenPayloads["email"]
	path := screenPayloads["home-path"]
	cases := map[string]string{
		"branch ref":       `{"key":"k","stage":"ACCEPTED","branch":{"ref":` + jsonQ(secret) + `}}`,
		"ready path":       `{"key":"k","stage":"ACCEPTED","ready":{"host":"h","slot":"s","path":` + jsonQ(path) + `,"head":"abc"}}`,
		"blocked step":     `{"key":"k","stage":"ACCEPTED","blocked":{"step":` + jsonQ(secret) + `,"detail":"d","resume":"r"}}`,
		"blocked detail":   `{"key":"k","stage":"ACCEPTED","blocked":{"step":"s","detail":` + jsonQ(path) + `,"resume":"r"}}`,
		"blocked resume":   `{"key":"k","stage":"ACCEPTED","blocked":{"step":"s","detail":"d","resume":` + jsonQ(secret) + `}}`,
		"evidence ref":     `{"key":"k","stage":"ACCEPTED","evidence":[{"stage":"x","ref":` + jsonQ(secret) + `}]}`,
		"evidence stage":   `{"key":"k","stage":"ACCEPTED","evidence":[{"stage":` + jsonQ(secret) + `,"ref":"r"}]}`,
		"clean row (ctrl)": `{"key":"k","stage":"ACCEPTED","branch":{"ref":"feature/x"},"ready":{"host":"h","slot":"s","path":"slot-1","head":"abc"}}`,
	}
	for name, row := range cases {
		t.Run("claim request/"+name, func(t *testing.T) {
			a := newApp(Config{ProjectRoot: t.TempDir()})
			a.intakeRequests = func(context.Context) ([]byte, error) { return []byte("[" + row + "]"), nil }
			rows := a.buildIntakeRequests(context.Background()).Rows
			if len(rows) != 1 {
				t.Fatalf("rows = %+v", rows)
			}
			want := !strings.Contains(name, "(ctrl)")
			if rows[0].Redacted != want {
				t.Errorf("Redacted = %v, want %v: %+v", rows[0].Redacted, want, rows[0])
			}
		})
	}
}

func TestRedactedBadge_MonitorIntakeBlock(t *testing.T) {
	email := screenPayloads["email"]
	for name, block := range map[string]string{
		"step":   `{"step":` + jsonQ(email) + `,"detail":"d","resume":"r"}`,
		"detail": `{"step":"s","detail":` + jsonQ(email) + `,"resume":"r"}`,
		"resume": `{"step":"s","detail":"d","resume":` + jsonQ(email) + `}`,
	} {
		t.Run(name, func(t *testing.T) {
			a := prApp(t, `{"rows":[],"intake":{"value":[{"key":"k","stage":"ACCEPTED","block":`+block+`}],"confirmed":true}}`, nil)
			rows := a.buildPRTracking(context.Background()).Intake.Rows
			if len(rows) != 1 || !rows[0].Redacted {
				t.Errorf("a masked block %s must carry the badge: %+v", name, rows)
			}
		})
	}
}

// TestRedactedBadge_HexFieldsAreRefinedBeforeTheCut: commit, target head, hash
// and the row head are normally hex, but the web makes no such assumption.
func TestRedactedBadge_HexFieldsAreRefinedBeforeTheCut(t *testing.T) {
	email := screenPayloads["email"]
	a := prApp(t, `{"rows":[{"Issue":7,"HeadSHA":{"value":`+jsonQ(email)+`,"confirmed":true}}]}`, nil)
	a.contextModel = func(context.Context, int) ([]byte, error) {
		l := &ctxmodel.Ledger{Issue: 7, Items: []ctxmodel.Item{{Section: ctxmodel.SectionSpec, Title: "t", Owner: "o", Commit: email, TargetHead: email, SHA256: email}}}
		return l.Marshal()
	}
	row := a.buildPRTracking(context.Background()).Rows[0]
	it := row.Ledger.Items[0]
	for name, got := range map[string]string{"commit": it.Commit, "target head": it.TargetHead, "hash": it.Hash, "row head": row.Head} {
		if strings.Contains(got, "alice") || strings.Contains(got, "@") {
			t.Errorf("%s = %q: the cut kept the start of an address", name, got)
		}
	}
	if !it.Redacted || !row.Redacted {
		t.Errorf("badge: item %v row %v", it.Redacted, row.Redacted)
	}
}
