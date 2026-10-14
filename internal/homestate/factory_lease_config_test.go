package homestate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// SPEC-FACTORY-LEASE-RENEW-001 (card t1663), milestone M1: the lease
// configuration keys and the open-time refusal (REQ-FLR-006, REQ-FLR-007).
//
// The keys are written into a temporary project's workflow configuration as
// fixture input only. No shipped template, shipped-key inventory entry, symmetry
// case, or struct registration is made here: that registration belongs to card
// t1636 (leader ruling, lane update during M1).

// factoryLeaseConfigFixture returns a temporary project root whose workflow
// configuration is body. An empty body writes no configuration file.
func factoryLeaseConfigFixture(t *testing.T, body string) string {
	t.Helper()
	root := t.TempDir()
	if body == "" {
		return root
	}
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir config sections: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "workflow.yaml"), []byte(body), 0o644); err != nil {
		t.Fatalf("write workflow.yaml: %v", err)
	}
	return root
}

// openFactoryRefusal returns the text of the refusal that OpenFactory gives for
// root. An accepted open fails the test, because the configuration must be
// refused (Q2: refuse to open, no clamp).
func openFactoryRefusal(t *testing.T, root string) string {
	t.Helper()
	db, err := OpenFactory(root)
	if err != nil {
		return err.Error()
	}
	_ = db.DB.Close()
	t.Fatalf("OpenFactory accepted the configuration; want a refusal")
	return ""
}

// AC-FLR-006 — a non-positive card lease duration refuses to open the factory,
// and the refusal names workflow.factory_lease.duration and the allowed range.
// The unparseable subtest is the AC-FLR-006-DOMAIN debt item (plan audit
// iteration 2): a value that is not a duration refuses in the same way.
func TestFactoryOpenRefusesNonPositiveLeaseDuration(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"zero", "workflow:\n  factory_lease:\n    duration: 0s\n"},
		{"negative", "workflow:\n  factory_lease:\n    duration: -5m\n"},
		{"unparseable", "workflow:\n  factory_lease:\n    duration: fifteen minutes\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			msg := openFactoryRefusal(t, factoryLeaseConfigFixture(t, c.body))
			for _, want := range []string{"workflow.factory_lease.duration", "greater than zero"} {
				if !strings.Contains(msg, want) {
					t.Fatalf("refusal = %q, want it to contain %q", msg, want)
				}
			}
		})
	}
}

// AC-FLR-006-DOMAIN (debt, plan audit iteration 2): a card lease shorter than
// two minutes leaves no whole-minute renewal period that satisfies the bound,
// because one half of the duration is under one minute. The refusal names that
// case and the key that cannot be satisfied.
func TestFactoryOpenRefusesSubTwoMinuteDuration(t *testing.T) {
	root := factoryLeaseConfigFixture(t, "workflow:\n  factory_lease:\n    duration: 1m\n")
	msg := openFactoryRefusal(t, root)
	for _, want := range []string{"workflow.factory_lease.renew_interval", "shorter than two minutes", "no whole-minute renew interval"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("refusal = %q, want it to contain %q", msg, want)
		}
	}
}

// AC-FLR-007 — a renewal period that is not a whole number of minutes from 1 to
// 59, or that is longer than one half of the card lease duration, refuses to
// open the factory. The refusal names workflow.factory_lease.renew_interval, the
// bound it violates, and the allowed range.
func TestFactoryOpenRefusesRenewIntervalOutsideBound(t *testing.T) {
	cases := []struct {
		name string
		body string
		want []string
	}{
		{
			name: "over-bound-at-10m-duration",
			body: "workflow:\n  factory_lease:\n    duration: 10m\n    renew_interval: 6m\n",
			want: []string{"workflow.factory_lease.renew_interval", "5m0s", "from 1 to 59"},
		},
		{
			name: "not-whole-minute-at-15m-duration",
			body: "workflow:\n  factory_lease:\n    duration: 15m\n    renew_interval: 7m30s\n",
			want: []string{"workflow.factory_lease.renew_interval", "whole number of minutes"},
		},
		{
			name: "over-bound-at-15m-duration",
			body: "workflow:\n  factory_lease:\n    duration: 15m\n    renew_interval: 8m\n",
			want: []string{"workflow.factory_lease.renew_interval", "7m30s"},
		},
		{
			name: "above-59-minutes-at-4h-duration",
			body: "workflow:\n  factory_lease:\n    duration: 4h\n    renew_interval: 60m\n",
			want: []string{"workflow.factory_lease.renew_interval", "from 1 to 59"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			msg := openFactoryRefusal(t, factoryLeaseConfigFixture(t, c.body))
			for _, want := range c.want {
				if !strings.Contains(msg, want) {
					t.Fatalf("refusal = %q, want it to contain %q", msg, want)
				}
			}
		})
	}
}

// AC-FLR-007, accept arm — a renewal period at the bound opens the factory. On
// the unchanged tree this passes without exercising the key: nothing reads
// workflow.factory_lease.renew_interval yet, so the accept arm is not a
// witness until a refusal exists to be contrasted with it.
func TestFactoryOpenAcceptsRenewIntervalAtBound(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"5m-at-10m-duration", "workflow:\n  factory_lease:\n    duration: 10m\n    renew_interval: 5m\n"},
		{"7m-at-15m-duration", "workflow:\n  factory_lease:\n    duration: 15m\n    renew_interval: 7m\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			db, err := OpenFactory(factoryLeaseConfigFixture(t, c.body))
			if err != nil {
				t.Fatalf("OpenFactory refused a bound-satisfying configuration: %v", err)
			}
			_ = db.DB.Close()
		})
	}
}
