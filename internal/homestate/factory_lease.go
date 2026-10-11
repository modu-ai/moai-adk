package homestate

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// SPEC-FACTORY-LEASE-RENEW-001 (card t1663), milestones M1 and M2: the workflow.factory_lease
// keys are parsed and bounds-checked when the factory opens. A value outside its bound
// refuses to open the factory, and the refusal names the key path, the allowed range, and
// the config file (ruling d-20261010T111725Z-3ef9). Nothing is clamped. A workflow file that
// cannot be read or decoded does not refuse the open: the keys fall back to the defaults
// with a warning.

const (
	factoryLeaseDurationKey      = "workflow.factory_lease.duration"
	factoryLeaseRenewIntervalKey = "workflow.factory_lease.renew_interval"
)

// errFactoryLeaseRefused is wrapped by every refusal to open the factory on account
// of the workflow.factory_lease keys.
var errFactoryLeaseRefused = errors.New("factory lease configuration refused")

// factoryLeaseKeys holds the effective values of the two keys.
type factoryLeaseKeys struct {
	Duration      time.Duration
	RenewInterval time.Duration
}

// loadFactoryLease reads the two keys of projectRoot and checks them against their
// bounds. A workflow file that cannot be read or decoded is not a refusal here:
// config.LoadFactoryLease falls back to the defaults and warns. A value the bounds
// reject refuses the open, and the refusal names the config file that holds it.
func loadFactoryLease(projectRoot string) (factoryLeaseKeys, error) {
	keys, err := resolveFactoryLease(config.LoadFactoryLease(projectRoot))
	if err != nil {
		return factoryLeaseKeys{}, fmt.Errorf("%w (config: %s)", err,
			filepath.Join(projectRoot, ".moai", "config", "sections", "workflow.yaml"))
	}
	return keys, nil
}

// resolveFactoryLease parses the two keys and checks them against their bounds. The
// duration must be a positive duration. The renewal period must be a whole number
// of minutes from 1 to 59 and no longer than one half of the duration. The duration
// is checked first, because the renewal bound depends on it.
func resolveFactoryLease(raw config.FactoryLeaseConfig) (factoryLeaseKeys, error) {
	duration, err := time.ParseDuration(strings.TrimSpace(raw.Duration))
	if err != nil || duration <= 0 {
		return factoryLeaseKeys{}, refuseFactoryLease("%s must be a positive duration (greater than zero); got %q",
			factoryLeaseDurationKey, raw.Duration)
	}
	bound := duration / 2
	renew, err := time.ParseDuration(strings.TrimSpace(raw.RenewInterval))
	wholeMinutes := err == nil && renew%time.Minute == 0 && renew >= time.Minute && renew <= 59*time.Minute
	if !wholeMinutes || renew > bound {
		note := ""
		if bound < time.Minute {
			note = "; the card lease duration is shorter than two minutes, so no whole-minute renew interval satisfies the bound"
		}
		return factoryLeaseKeys{}, refuseFactoryLease(
			"%s must be a whole number of minutes from 1 to 59 and no longer than one half of %s (%s); got %q%s",
			factoryLeaseRenewIntervalKey, factoryLeaseDurationKey, bound, raw.RenewInterval, note)
	}
	return factoryLeaseKeys{Duration: duration, RenewInterval: renew}, nil
}

// EffectiveFactoryLease returns the card lease duration and the watchdog renewal
// period that the factory opens with for projectRoot. It reads through the same entry
// point as OpenFactory, and it does not open the database, so a surface that must
// report the values can do so without creating one. A configuration that the factory
// would refuse returns the same refusal.
func EffectiveFactoryLease(projectRoot string) (cardLease, renewInterval time.Duration, err error) {
	keys, err := loadFactoryLease(projectRoot)
	if err != nil {
		return 0, 0, err
	}
	return keys.Duration, keys.RenewInterval, nil
}

// refuseFactoryLease returns a refusal that wraps errFactoryLeaseRefused.
func refuseFactoryLease(format string, args ...any) error {
	return fmt.Errorf("%w: %s", errFactoryLeaseRefused, fmt.Sprintf(format, args...))
}

// CardLeaseDuration is the card lease duration this handle writes. A handle opened
// through a project root carries the configured workflow.factory_lease.duration. Any
// other handle carries the compiled default, FactoryLeaseDuration.
func (f *FactoryDB) CardLeaseDuration() time.Duration {
	if f.lease.Duration > 0 {
		return f.lease.Duration
	}
	return FactoryLeaseDuration
}

// RenewInterval is the effective watchdog renewal period, workflow.factory_lease.
// renew_interval. A handle that was not opened through a project root reports the
// compiled default.
func (f *FactoryDB) RenewInterval() time.Duration {
	if f.lease.RenewInterval > 0 {
		return f.lease.RenewInterval
	}
	return config.DefaultFactoryLeaseRenewInterval
}
