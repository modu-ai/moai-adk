package homestate

// RED-phase contract stub for the F1 lease API (t1239 M4): it fixes the names
// the M4 tests compile against. The GREEN commit deletes this file.

import (
	"context"
	"errors"
	"time"
)

var FactoryLeaseDuration = 15 * time.Minute

var errF1LeaseNotImplemented = errors.New("F1 lease renewal not implemented")

func (f *FactoryDB) RenewLease(context.Context, string, string, string, time.Time) (Card, error) {
	return Card{}, errF1LeaseNotImplemented
}
