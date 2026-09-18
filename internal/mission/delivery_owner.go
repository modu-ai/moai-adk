package mission

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrDeliveryUnsupported    = errors.New("mission delivery: provider_unsupported")
	ErrDeliveryAlreadyApplied = errors.New("mission delivery: operation_already_applied")
)

func IsDeliveryUnsupported(err error) bool    { return errors.Is(err, ErrDeliveryUnsupported) }
func IsDeliveryAlreadyApplied(err error) bool { return errors.Is(err, ErrDeliveryAlreadyApplied) }

// DeliverySnapshot contains provider-read authoritative state. Callers do not
// supply gate booleans directly to an effect; CapabilityDeliveryOwner always
// obtains this snapshot from the configured provider immediately beforehand.
type DeliverySnapshot struct {
	HeadSHA, LocalDevelopSHA, OriginDevelopSHA, ReleaseHeadSHA, MainLandedSHA string
	LocalMergesReady, OriginReadback, DevelopCIPassed                         bool
	ReleaseAuditPassed, ReleaseReviewPassed, ReleaseCIPassed                  bool
	ProtectedMain, MainContainsRelease                                        bool
}

type DeliveryProvider interface {
	Capability(Action) bool
	Snapshot(context.Context, Action, string) (DeliverySnapshot, error)
	Apply(context.Context, Action, string, string) error
	Readback(context.Context, Action, string, string) (bool, error)
}

type UnsupportedDeliveryProvider struct{}

func (UnsupportedDeliveryProvider) Capability(Action) bool { return false }
func (UnsupportedDeliveryProvider) Snapshot(context.Context, Action, string) (DeliverySnapshot, error) {
	return DeliverySnapshot{}, ErrDeliveryUnsupported
}
func (UnsupportedDeliveryProvider) Apply(context.Context, Action, string, string) error {
	return ErrDeliveryUnsupported
}
func (UnsupportedDeliveryProvider) Readback(context.Context, Action, string, string) (bool, error) {
	return false, ErrDeliveryUnsupported
}

type CapabilityDeliveryOwner struct {
	Provider DeliveryProvider
	Action   Action
	Target   string
}

func validateDeliverySnapshot(action Action, s DeliverySnapshot) error {
	if strings.TrimSpace(s.HeadSHA) == "" {
		return errors.New("mission delivery: head_readback_missing")
	}
	switch action {
	case ActionBatchPush:
		if !s.LocalMergesReady {
			return errors.New("mission delivery: local_merges_missing")
		}
	case ActionReleaseBranch:
		if !s.OriginReadback || !s.DevelopCIPassed || s.OriginDevelopSHA == "" || s.HeadSHA != s.OriginDevelopSHA {
			return errors.New("mission delivery: develop_gate_stale")
		}
	case ActionReleasePR:
		if !s.ReleaseAuditPassed || !s.ReleaseReviewPassed || !s.ReleaseCIPassed || s.ReleaseHeadSHA == "" || s.HeadSHA != s.ReleaseHeadSHA {
			return errors.New("mission delivery: release_gate_failed")
		}
	case ActionMainMerge:
		if !s.ProtectedMain || !s.MainContainsRelease || s.MainLandedSHA == "" {
			return errors.New("mission delivery: main_landing_unverified")
		}
	default:
		return ErrDeliveryUnsupported
	}
	return nil
}

func (o CapabilityDeliveryOwner) Snapshot(ctx context.Context) (DeliverySnapshot, error) {
	if o.Provider == nil || !o.Provider.Capability(o.Action) {
		return DeliverySnapshot{}, ErrDeliveryUnsupported
	}
	snapshot, err := o.Provider.Snapshot(ctx, o.Action, o.Target)
	if err != nil {
		return DeliverySnapshot{}, fmt.Errorf("mission delivery: snapshot: %w", err)
	}
	if err := validateDeliverySnapshot(o.Action, snapshot); err != nil {
		return DeliverySnapshot{}, err
	}
	return snapshot, nil
}

func (o CapabilityDeliveryOwner) Readback(ctx context.Context, operationID string) (bool, error) {
	if o.Provider == nil || !o.Provider.Capability(o.Action) {
		return false, ErrDeliveryUnsupported
	}
	return o.Provider.Readback(ctx, o.Action, o.Target, operationID)
}

func (o CapabilityDeliveryOwner) Apply(ctx context.Context, operationID string) error {
	if _, err := o.Snapshot(ctx); err != nil {
		return err
	}
	applied, err := o.Readback(ctx, operationID)
	if err != nil {
		return err
	}
	if applied {
		return ErrDeliveryAlreadyApplied
	}
	return o.Provider.Apply(ctx, o.Action, o.Target, operationID)
}
