package homestate

import (
	"context"
	"testing"
)

// Every store method fails loudly on a closed database — none of them
// silently no-ops on a dead handle.
func TestStoreMethodsFailLoudlyAfterClose(t *testing.T) {
	ctx := context.Background()
	newClosed := func(t *testing.T) *FactoryDB {
		t.Helper()
		db := frOpen(t)
		if err := db.Close(); err != nil {
			t.Fatal(err)
		}
		return db
	}
	cases := []struct {
		name string
		call func(*FactoryDB) error
	}{
		{"SaveResume", func(d *FactoryDB) error { return d.SaveResume(ctx, ResumeHandoff{SpecID: "S", Body: "b"}) }},
		{"ImportLegacyResume", func(d *FactoryDB) error {
			_, err := d.ImportLegacyResume(ctx, ResumeHandoff{SpecID: "S", Body: "b"})
			return err
		}},
		{"ClaimResume", func(d *FactoryDB) error {
			_, _, err := d.ClaimResume(ctx, ResumeClaim{Token: "tok"})
			return err
		}},
		{"FinishResume", func(d *FactoryDB) error { return d.FinishResume(ctx, 1, "tok", "consumed", "") }},
		{"RecoverLegacyResume", func(d *FactoryDB) error {
			return d.RecoverLegacyResume(ctx, 1, "tok", "requeue", func(int) (string, ProcessIdentityState) { return "", ProcessIdentityDead }, nil)
		}},
		{"ClearPendingResume", func(d *FactoryDB) error { return d.ClearPendingResume(ctx) }},
		{"ExpirePendingResume", func(d *FactoryDB) error { return d.ExpirePendingResume(ctx) }},
		{"ExpireResumeIfPending", func(d *FactoryDB) error {
			_, err := d.ExpireResumeIfPending(ctx, 1)
			return err
		}},
		{"SaveMemory", func(d *FactoryDB) error { return d.SaveMemory(ctx, MemoryHandoff{Spec: "S", Body: "b"}) }},
		{"SetMemoryStatus", func(d *FactoryDB) error { return d.SetMemoryStatus(ctx, 1, "persisted", "") }},
		{"StampRunOwner", func(d *FactoryDB) error { return d.StampRunOwner(ctx, "run", 42, "fp") }},
		{"LegacyResumeRetired", func(d *FactoryDB) error {
			_, err := d.LegacyResumeRetired(ctx)
			return err
		}},
		{"ReadPendingResume", func(d *FactoryDB) error {
			_, _, err := d.ReadPendingResume(ctx)
			return err
		}},
		{"ReadPendingMemory", func(d *FactoryDB) error {
			_, _, err := d.ReadPendingMemory(ctx)
			return err
		}},
		{"LoadCard", func(d *FactoryDB) error {
			_, err := d.LoadCard(ctx, "run", "card")
			return err
		}},
		{"ListCards", func(d *FactoryDB) error {
			_, err := d.ListCards(ctx, "run")
			return err
		}},
		{"RetireRunIfDead", func(d *FactoryDB) error {
			_, err := d.RetireRunIfDead(ctx, "run", ReconcileOptions{})
			return err
		}},
		{"ClearRunOwner", func(d *FactoryDB) error { return d.ClearRunOwner(ctx, "run") }},
		{"ImportLegacyWorkers", func(d *FactoryDB) error { return d.ImportLegacyWorkers("/nonexistent/workers.json") }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.call(newClosed(t)); err == nil {
				t.Fatalf("%s on closed db: err = nil, want failure", tc.name)
			}
		})
	}
}
