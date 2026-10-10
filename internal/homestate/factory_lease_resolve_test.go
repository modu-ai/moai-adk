package homestate

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// Coverage tests for the workflow.factory_lease checks and the handle accessors
// (SPEC-FACTORY-LEASE-RENEW-001, card t1663, milestone M1). They were written after the
// GREEN of AC-FLR-005, AC-FLR-006, and AC-FLR-007, to cover the boundary values, the
// refusal paths, and the fallbacks. They are not RED evidence.

func TestResolveFactoryLeaseAcceptsBoundaryValues(t *testing.T) {
	cases := []struct {
		name string
		raw  config.FactoryLeaseConfig
		want factoryLeaseKeys
	}{
		{"at-bound-15m", config.FactoryLeaseConfig{Duration: "15m", RenewInterval: "7m"}, factoryLeaseKeys{Duration: 15 * time.Minute, RenewInterval: 7 * time.Minute}},
		{"at-bound-10m", config.FactoryLeaseConfig{Duration: "10m", RenewInterval: "5m"}, factoryLeaseKeys{Duration: 10 * time.Minute, RenewInterval: 5 * time.Minute}},
		{"two-minute-duration-one-minute-renew", config.FactoryLeaseConfig{Duration: "2m", RenewInterval: "1m"}, factoryLeaseKeys{Duration: 2 * time.Minute, RenewInterval: time.Minute}},
		{"one-hour-duration-thirty-minute-renew", config.FactoryLeaseConfig{Duration: "1h", RenewInterval: "30m"}, factoryLeaseKeys{Duration: time.Hour, RenewInterval: 30 * time.Minute}},
		{"compiled-defaults", config.NewDefaultWorkflowConfig().FactoryLease, factoryLeaseKeys{Duration: 15 * time.Minute, RenewInterval: 5 * time.Minute}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := resolveFactoryLease(c.raw)
			if err != nil {
				t.Fatalf("resolveFactoryLease(%+v): %v", c.raw, err)
			}
			if got != c.want {
				t.Fatalf("resolveFactoryLease(%+v) = %+v, want %+v", c.raw, got, c.want)
			}
		})
	}
}

func TestResolveFactoryLeaseRefusalsNameTheKeyAndTheRange(t *testing.T) {
	cases := []struct {
		name string
		raw  config.FactoryLeaseConfig
		want []string
	}{
		{"duration-empty", config.FactoryLeaseConfig{Duration: "", RenewInterval: "5m"}, []string{"workflow.factory_lease.duration", "greater than zero"}},
		{"duration-bare-number", config.FactoryLeaseConfig{Duration: "15", RenewInterval: "5m"}, []string{"workflow.factory_lease.duration", "greater than zero"}},
		{"renew-zero", config.FactoryLeaseConfig{Duration: "15m", RenewInterval: "0m"}, []string{"workflow.factory_lease.renew_interval", "from 1 to 59"}},
		{"renew-negative", config.FactoryLeaseConfig{Duration: "15m", RenewInterval: "-1m"}, []string{"workflow.factory_lease.renew_interval", "from 1 to 59"}},
		{"renew-bare-number", config.FactoryLeaseConfig{Duration: "15m", RenewInterval: "5"}, []string{"workflow.factory_lease.renew_interval", "whole number of minutes"}},
		{"renew-empty", config.FactoryLeaseConfig{Duration: "15m", RenewInterval: ""}, []string{"workflow.factory_lease.renew_interval", "whole number of minutes"}},
		{"renew-seconds", config.FactoryLeaseConfig{Duration: "15m", RenewInterval: "90s"}, []string{"workflow.factory_lease.renew_interval", "whole number of minutes"}},
		{"sub-two-minute-duration", config.FactoryLeaseConfig{Duration: "1m", RenewInterval: "1m"}, []string{"workflow.factory_lease.renew_interval", "shorter than two minutes", "no whole-minute renew interval", "30s"}},
		{"sub-two-minute-duration-out-of-range-renew", config.FactoryLeaseConfig{Duration: "1m", RenewInterval: "0m"}, []string{"shorter than two minutes", "no whole-minute renew interval"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := resolveFactoryLease(c.raw)
			if err == nil {
				t.Fatalf("resolveFactoryLease(%+v) accepted the configuration; want a refusal", c.raw)
			}
			if !errors.Is(err, errFactoryLeaseRefused) {
				t.Fatalf("refusal %v does not wrap errFactoryLeaseRefused", err)
			}
			for _, want := range c.want {
				if !strings.Contains(err.Error(), want) {
					t.Fatalf("refusal = %q, want it to contain %q", err.Error(), want)
				}
			}
		})
	}
}

func TestFactoryLeaseAccessorsFallBackToTheDefaults(t *testing.T) {
	var db FactoryDB
	if got := db.CardLeaseDuration(); got != FactoryLeaseDuration {
		t.Fatalf("CardLeaseDuration() on a handle without a project root = %s, want the compiled default %s", got, FactoryLeaseDuration)
	}
	if got := db.RenewInterval(); got != config.DefaultFactoryLeaseRenewInterval {
		t.Fatalf("RenewInterval() on a handle without a project root = %s, want %s", got, config.DefaultFactoryLeaseRenewInterval)
	}
}

func TestFactoryLeaseAccessorsReportTheConfiguredValues(t *testing.T) {
	db := FactoryDB{lease: factoryLeaseKeys{Duration: 20 * time.Minute, RenewInterval: 4 * time.Minute}}
	if got := db.CardLeaseDuration(); got != 20*time.Minute {
		t.Fatalf("CardLeaseDuration() = %s, want 20m0s", got)
	}
	if got := db.RenewInterval(); got != 4*time.Minute {
		t.Fatalf("RenewInterval() = %s, want 4m0s", got)
	}
}

func TestOpenFactoryRefusalLeavesNoDatabase(t *testing.T) {
	root := factorySandbox(t)
	writeFactoryWorkflowConfig(t, root, "workflow:\n  factory_lease:\n    duration: 0s\n")
	if _, err := OpenFactory(root); !errors.Is(err, errFactoryLeaseRefused) {
		t.Fatalf("OpenFactory err = %v, want a factory lease refusal", err)
	}
	path, err := FactoryDBPath(root)
	if err != nil {
		t.Fatalf("FactoryDBPath: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("a refused open left a database at %s (stat error %v)", path, err)
	}
}

func TestOpenFactoryRefusesAWorkflowFileThatCannotBeRead(t *testing.T) {
	root := factorySandbox(t)
	writeFactoryWorkflowConfig(t, root, "workflow:\n  factory_lease: [unclosed\n")
	_, err := OpenFactory(root)
	if !errors.Is(err, errFactoryLeaseRefused) || !strings.Contains(err.Error(), "could not be read") {
		t.Fatalf("OpenFactory err = %v, want a refusal that says the workflow configuration could not be read", err)
	}
}

func TestOpenFactoryBoundedCarriesTheSameKeys(t *testing.T) {
	root := factorySandbox(t)
	writeFactoryWorkflowConfig(t, root, "workflow:\n  factory_lease:\n    duration: 20m\n    renew_interval: 4m\n")
	db, err := OpenFactoryBounded(root, time.Second)
	if err != nil {
		t.Fatalf("OpenFactoryBounded: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if db.CardLeaseDuration() != 20*time.Minute || db.RenewInterval() != 4*time.Minute {
		t.Fatalf("OpenFactoryBounded carries %s and %s, want 20m0s and 4m0s", db.CardLeaseDuration(), db.RenewInterval())
	}

	refused := factorySandbox(t)
	writeFactoryWorkflowConfig(t, refused, "workflow:\n  factory_lease:\n    duration: 0s\n")
	if _, err := OpenFactoryBounded(refused, time.Second); !errors.Is(err, errFactoryLeaseRefused) {
		t.Fatalf("OpenFactoryBounded err = %v, want a factory lease refusal", err)
	}
}
