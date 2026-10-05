package mission

import (
	"context"
	"errors"
)

type RuntimeMode string

const (
	RuntimeDurable           RuntimeMode = "durable"
	RuntimeActiveSessionOnly RuntimeMode = "active-session-only"
)

type MissionRuntimeProbe interface {
	StartSupported(context.Context) bool
	ReconnectSupported(context.Context) bool
	ReplaceSupported(context.Context) bool
	CredentialValid(context.Context) bool
	ProcessIdentitySupported(context.Context) bool
}

type RuntimeCapability struct {
	Mode                     RuntimeMode
	CanStart                 bool
	CanReconnect             bool
	CanReplace               bool
	CredentialValid          bool
	ProcessIdentitySupported bool
}

func ProbeMissionRuntimeCapability(ctx context.Context, probe MissionRuntimeProbe) RuntimeCapability {
	if probe == nil {
		return RuntimeCapability{Mode: RuntimeActiveSessionOnly}
	}
	result := RuntimeCapability{
		CanStart:                 probe.StartSupported(ctx),
		CanReconnect:             probe.ReconnectSupported(ctx),
		CanReplace:               probe.ReplaceSupported(ctx),
		CredentialValid:          probe.CredentialValid(ctx),
		ProcessIdentitySupported: probe.ProcessIdentitySupported(ctx),
		Mode:                     RuntimeActiveSessionOnly,
	}
	if result.CanStart && result.CanReconnect && result.CanReplace && result.CredentialValid && result.ProcessIdentitySupported {
		result.Mode = RuntimeDurable
	}
	return result
}

func CanReplaceRuntimeOwner(capability RuntimeCapability, priorOwnerEffectsStopped bool) bool {
	return capability.Mode == RuntimeDurable && capability.CanReplace && capability.CredentialValid && priorOwnerEffectsStopped
}

type RuntimeLease struct {
	MissionID    string
	ContractHash string
	SnapshotHash string
	OwnerID      string
	Version      int64
}

func TakeoverRuntimeLease(current RuntimeLease, newOwner string, priorOwnerEffectsStopped bool, contractHash, snapshotHash string) (RuntimeLease, error) {
	if newOwner == "" || !priorOwnerEffectsStopped {
		return RuntimeLease{}, errors.New("mission runtime: prior_owner_not_stopped")
	}
	if current.ContractHash != contractHash || current.SnapshotHash != snapshotHash {
		return RuntimeLease{}, errors.New("mission runtime: snapshot_lineage_mismatch")
	}
	current.OwnerID = newOwner
	current.Version++
	return current, nil
}
