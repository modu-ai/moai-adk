package mission

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type recordingDeliveryProvider struct {
	capable     bool
	applied     map[string]bool
	order       []string
	snap        DeliverySnapshot
	crash       bool
	snapshotErr error
	readbackErr error
	applyErr    error
}

func (p *recordingDeliveryProvider) Capability(Action) bool { return p.capable }
func (p *recordingDeliveryProvider) Snapshot(context.Context, Action, string) (DeliverySnapshot, error) {
	p.order = append(p.order, "snapshot")
	return p.snap, p.snapshotErr
}
func (p *recordingDeliveryProvider) Apply(_ context.Context, action Action, _ string, operationID string) error {
	p.order = append(p.order, "apply:"+string(action))
	p.applied[operationID] = true
	if p.applyErr != nil {
		return p.applyErr
	}
	if p.crash {
		return errors.New("crash_after_effect")
	}
	return nil
}
func (p *recordingDeliveryProvider) Readback(_ context.Context, _ Action, _ string, operationID string) (bool, error) {
	p.order = append(p.order, "readback")
	return p.applied[operationID], p.readbackErr
}

func TestCapabilityDeliveryOwnerOrdersSnapshotEffectReadbackAndRecoversCrash(t *testing.T) {
	ctx := context.Background()
	p := &recordingDeliveryProvider{capable: true, applied: map[string]bool{}, snap: DeliverySnapshot{HeadSHA: "develop", LocalMergesReady: true, OriginDevelopSHA: "develop", DevelopCIPassed: true}}
	owner := CapabilityDeliveryOwner{Provider: p, Action: ActionBatchPush, Target: "repo:fixture"}
	if _, err := owner.Snapshot(ctx); err != nil {
		t.Fatal(err)
	}
	if err := owner.Apply(ctx, "op-1"); err != nil {
		t.Fatal(err)
	}
	if applied, err := owner.Readback(ctx, "op-1"); err != nil || !applied {
		t.Fatalf("readback=%v err=%v", applied, err)
	}
	if want := []string{"snapshot", "snapshot", "readback", "apply:batch_push", "readback"}; !reflect.DeepEqual(p.order, want) {
		t.Fatalf("order=%v want=%v", p.order, want)
	}

	p.order, p.crash = nil, true
	if err := owner.Apply(ctx, "op-crash"); err == nil {
		t.Fatal("crash cut missing")
	}
	if applied, err := owner.Readback(ctx, "op-crash"); err != nil || !applied {
		t.Fatalf("crash readback=%v err=%v", applied, err)
	}
	if err := owner.Apply(ctx, "op-crash"); err == nil || !IsDeliveryAlreadyApplied(err) {
		t.Fatalf("duplicate apply err=%v", err)
	}
}

func TestCapabilityDeliveryOwnerFailsClosedWhenProviderUnsupportedOrGateStale(t *testing.T) {
	ctx := context.Background()
	unsupported := CapabilityDeliveryOwner{Provider: UnsupportedDeliveryProvider{}, Action: ActionReleasePR, Target: "repo:fixture"}
	if _, err := unsupported.Snapshot(ctx); err == nil || !IsDeliveryUnsupported(err) {
		t.Fatalf("unsupported err=%v", err)
	}
	p := &recordingDeliveryProvider{capable: true, applied: map[string]bool{}, snap: DeliverySnapshot{HeadSHA: "release", ReleaseHeadSHA: "release", ReleaseAuditPassed: true, ReleaseReviewPassed: false, ReleaseCIPassed: true}}
	owner := CapabilityDeliveryOwner{Provider: p, Action: ActionReleasePR, Target: "repo:fixture"}
	if _, err := owner.Snapshot(ctx); err == nil {
		t.Fatal("failed review gate accepted")
	}
}

func TestCapabilityDeliveryOwnerEnforcesEveryReleaseActionSHAAndGate(t *testing.T) {
	tests := []struct {
		name   string
		action Action
		snap   DeliverySnapshot
	}{
		{name: "batch push", action: ActionBatchPush, snap: DeliverySnapshot{HeadSHA: "develop-a", LocalMergesReady: true}},
		{name: "release branch", action: ActionReleaseBranch, snap: DeliverySnapshot{HeadSHA: "develop-b", OriginDevelopSHA: "develop-b", OriginReadback: true, DevelopCIPassed: true}},
		{name: "release pr", action: ActionReleasePR, snap: DeliverySnapshot{HeadSHA: "release-c", ReleaseHeadSHA: "release-c", ReleaseAuditPassed: true, ReleaseReviewPassed: true, ReleaseCIPassed: true}},
		{name: "main merge", action: ActionMainMerge, snap: DeliverySnapshot{HeadSHA: "release-d", MainLandedSHA: "main-d", ProtectedMain: true, MainContainsRelease: true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &recordingDeliveryProvider{capable: true, applied: map[string]bool{}, snap: tt.snap}
			owner := CapabilityDeliveryOwner{Provider: provider, Action: tt.action, Target: "repo:fixture"}
			if err := owner.Apply(context.Background(), "op-"+string(tt.action)); err != nil {
				t.Fatal(err)
			}
			applied, err := owner.Readback(context.Background(), "op-"+string(tt.action))
			if err != nil || !applied {
				t.Fatalf("readback=%v err=%v", applied, err)
			}
		})
	}

	staleSHA := &recordingDeliveryProvider{capable: true, applied: map[string]bool{}, snap: DeliverySnapshot{
		HeadSHA: "release-new", ReleaseHeadSHA: "release-old",
		ReleaseAuditPassed: true, ReleaseReviewPassed: true, ReleaseCIPassed: true,
	}}
	if err := (CapabilityDeliveryOwner{Provider: staleSHA, Action: ActionReleasePR, Target: "repo:fixture"}).Apply(context.Background(), "op-stale"); err == nil {
		t.Fatal("stale release SHA passed the provider gate")
	}
	if len(staleSHA.applied) != 0 {
		t.Fatalf("stale gate caused effects: %+v", staleSHA.applied)
	}
}

func TestCapabilityDeliveryOwnerProviderAndGateFailuresHaveNoEffect(t *testing.T) {
	ctx := context.Background()
	unsupported := UnsupportedDeliveryProvider{}
	if _, err := unsupported.Snapshot(ctx, ActionBatchPush, "repo:x"); !IsDeliveryUnsupported(err) {
		t.Fatalf("unsupported snapshot err=%v", err)
	}
	if err := unsupported.Apply(ctx, ActionBatchPush, "repo:x", "op"); !IsDeliveryUnsupported(err) {
		t.Fatalf("unsupported apply err=%v", err)
	}
	if _, err := unsupported.Readback(ctx, ActionBatchPush, "repo:x", "op"); !IsDeliveryUnsupported(err) {
		t.Fatalf("unsupported readback err=%v", err)
	}

	invalid := []struct {
		name   string
		action Action
		snap   DeliverySnapshot
	}{
		{name: "head missing", action: ActionBatchPush, snap: DeliverySnapshot{LocalMergesReady: true}},
		{name: "local merges missing", action: ActionBatchPush, snap: DeliverySnapshot{HeadSHA: "h"}},
		{name: "develop gate", action: ActionReleaseBranch, snap: DeliverySnapshot{HeadSHA: "h", OriginDevelopSHA: "other", OriginReadback: true, DevelopCIPassed: true}},
		{name: "release gate", action: ActionReleasePR, snap: DeliverySnapshot{HeadSHA: "h", ReleaseHeadSHA: "h", ReleaseAuditPassed: true, ReleaseReviewPassed: true}},
		{name: "main gate", action: ActionMainMerge, snap: DeliverySnapshot{HeadSHA: "h", ProtectedMain: true}},
		{name: "unknown action", action: Action("unknown"), snap: DeliverySnapshot{HeadSHA: "h"}},
	}
	for _, tt := range invalid {
		t.Run(tt.name, func(t *testing.T) {
			provider := &recordingDeliveryProvider{capable: true, applied: map[string]bool{}, snap: tt.snap}
			err := (CapabilityDeliveryOwner{Provider: provider, Action: tt.action, Target: "repo:x"}).Apply(ctx, "op")
			if err == nil || len(provider.applied) != 0 {
				t.Fatalf("err=%v effects=%v", err, provider.applied)
			}
		})
	}

	snapshotFailure := &recordingDeliveryProvider{capable: true, applied: map[string]bool{}, snapshotErr: errors.New("snapshot unavailable")}
	if err := (CapabilityDeliveryOwner{Provider: snapshotFailure, Action: ActionBatchPush, Target: "repo:x"}).Apply(ctx, "op"); err == nil {
		t.Fatal("snapshot provider failure ignored")
	}
	readbackFailure := &recordingDeliveryProvider{capable: true, applied: map[string]bool{}, snap: DeliverySnapshot{HeadSHA: "h", LocalMergesReady: true}, readbackErr: errors.New("readback unavailable")}
	if err := (CapabilityDeliveryOwner{Provider: readbackFailure, Action: ActionBatchPush, Target: "repo:x"}).Apply(ctx, "op"); err == nil {
		t.Fatal("readback provider failure ignored")
	}
	applyFailure := &recordingDeliveryProvider{capable: true, applied: map[string]bool{}, snap: DeliverySnapshot{HeadSHA: "h", LocalMergesReady: true}, applyErr: errors.New("apply failed")}
	if err := (CapabilityDeliveryOwner{Provider: applyFailure, Action: ActionBatchPush, Target: "repo:x"}).Apply(ctx, "op"); err == nil {
		t.Fatal("apply provider failure ignored")
	}
	if _, err := (CapabilityDeliveryOwner{Action: ActionBatchPush}).Readback(ctx, "op"); !IsDeliveryUnsupported(err) {
		t.Fatalf("nil provider readback err=%v", err)
	}
}
