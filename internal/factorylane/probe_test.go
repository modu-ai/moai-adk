package factorylane

import (
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/sessionmsg"
)

// probeNow is the fixed evaluation instant every probe test reads against;
// heartbeat ages below are offsets from it.
var probeNow = time.Date(2026, 9, 29, 12, 0, 0, 0, time.UTC)

// probePeer builds one registered-peer row with LastHeartbeat age rel to
// probeNow.
func probePeer(name string, heartbeatAge time.Duration) sessionmsg.AgentInfo {
	return sessionmsg.AgentInfo{
		AgentID:       "claude-abcd1234",
		Kind:          "claude",
		Name:          name,
		LastHeartbeat: probeNow.Add(-heartbeatAge),
	}
}

// REQ-FLA-001 / AC-FLA-001: an empty registry means no reachable lead peer —
// the probe reports channel-unavailable and names why.
func TestProbeReportsChannelUnavailableWhenNoPeersRegistered(t *testing.T) {
	got := EvaluateAvailability(ProbeInput{Now: probeNow})
	if got.Verdict != VerdictUnavailable {
		t.Fatalf("verdict = %q, want %q (empty registry is an unavailable channel)", got.Verdict, VerdictUnavailable)
	}
	if got.Reason == "" {
		t.Fatal("reason is empty — an unavailable verdict must name why")
	}
}

// REQ-FLA-001: a registered peer with a fresh heartbeat is an available
// channel.
func TestProbeReportsAvailableWhenPeerRegisteredAndHeartbeatFresh(t *testing.T) {
	got := EvaluateAvailability(ProbeInput{
		Now:      probeNow,
		Registry: []sessionmsg.AgentInfo{probePeer("team-lead", time.Minute)},
	})
	if got.Verdict != VerdictAvailable {
		t.Fatalf("verdict = %q, want %q (registered peer, fresh heartbeat): %s", got.Verdict, VerdictAvailable, got.Reason)
	}
}

// REQ-FLA-001: when the probe names a lead peer, an absent registration is
// channel-unavailable even though other peers exist.
func TestProbeReportsChannelUnavailableWhenNamedLeadPeerUnregistered(t *testing.T) {
	got := EvaluateAvailability(ProbeInput{
		Now:      probeNow,
		LeadPeer: "team-lead",
		Registry: []sessionmsg.AgentInfo{probePeer("some-other-agent", time.Minute)},
	})
	if got.Verdict != VerdictUnavailable {
		t.Fatalf("verdict = %q, want %q (named lead peer absent from registry)", got.Verdict, VerdictUnavailable)
	}
	if got.Reason == "" {
		t.Fatal("reason is empty — the refusal must name the missing lead peer")
	}
}

// REQ-FLA-001: the lead's heartbeat age exceeding the configured availability
// bound decides unavailable — this is the stale-lead edge case (acceptance.md
// §D.2); the bound travels in ProbeInput, never inline.
func TestProbeReportsChannelUnavailableWhenLeadHeartbeatAgeExceedsBound(t *testing.T) {
	const bound = 30 * time.Minute
	got := EvaluateAvailability(ProbeInput{
		Now:          probeNow,
		LeadPeer:     "team-lead",
		OfflineBound: bound,
		Registry:     []sessionmsg.AgentInfo{probePeer("team-lead", bound+time.Minute)},
	})
	if got.Verdict != VerdictUnavailable {
		t.Fatalf("verdict = %q, want %q (heartbeat age %s exceeds bound %s)", got.Verdict, VerdictUnavailable, bound+time.Minute, bound)
	}
}

// REQ-FLA-001: a heartbeat age within the bound stays available — the probe
// must not fail closed on a healthy lead.
func TestProbeReportsAvailableWhenLeadHeartbeatAgeWithinBound(t *testing.T) {
	const bound = 30 * time.Minute
	got := EvaluateAvailability(ProbeInput{
		Now:          probeNow,
		LeadPeer:     "team-lead",
		OfflineBound: bound,
		Registry:     []sessionmsg.AgentInfo{probePeer("team-lead", bound-time.Minute)},
	})
	if got.Verdict != VerdictAvailable {
		t.Fatalf("verdict = %q, want %q (heartbeat age %s within bound %s): %s", got.Verdict, VerdictAvailable, bound-time.Minute, bound, got.Reason)
	}
}

// noResponseObs builds one expired-unacked directed request: requested at
// probeNow-20m, timer 10m (expired at probeNow-10m), bound 30m (runs to
// probeNow+10m).
func noResponseObs() Observation {
	timer := 10 * time.Minute
	bound := 30 * time.Minute
	expired := probeNow.Add(-10 * time.Minute)
	return Observation{
		Lane:         "lane-1",
		RequestedAt:  probeNow.Add(-20 * time.Minute),
		Timer:        timer.String(),
		Bound:        bound.String(),
		NoResponseAt: &expired,
	}
}

// REQ-FLA-002: once the no-response observation is recorded, the channel is
// treated unavailable for the remainder of the bound period — even with a
// nominally healthy registry.
func TestProbeReportsChannelUnavailableDuringNoResponseBoundPeriod(t *testing.T) {
	got := EvaluateAvailability(ProbeInput{
		Now:          probeNow,
		LeadPeer:     "team-lead",
		OfflineBound: 30 * time.Minute,
		Registry:     []sessionmsg.AgentInfo{probePeer("team-lead", time.Minute)},
		Observations: []Observation{noResponseObs()},
	})
	if got.Verdict != VerdictUnavailable {
		t.Fatalf("verdict = %q, want %q (active no-response observation): %s", got.Verdict, VerdictUnavailable, got.Reason)
	}
	wantUntil := probeNow.Add(10 * time.Minute).Format(time.RFC3339)
	if got.UnavailableUntil != wantUntil {
		t.Fatalf("UnavailableUntil = %q, want %q (requested_at + bound)", got.UnavailableUntil, wantUntil)
	}
}

// REQ-FLA-002: after the bound period ends the channel is available again.
func TestProbeReportsAvailableAgainAfterNoResponseBoundExpires(t *testing.T) {
	got := EvaluateAvailability(ProbeInput{
		Now:          probeNow.Add(11 * time.Minute),
		LeadPeer:     "team-lead",
		OfflineBound: 30 * time.Minute,
		Registry:     []sessionmsg.AgentInfo{probePeer("team-lead", time.Minute)},
		Observations: []Observation{noResponseObs()},
	})
	if got.Verdict != VerdictAvailable {
		t.Fatalf("verdict = %q, want %q (bound period ended): %s", got.Verdict, VerdictAvailable, got.Reason)
	}
}

// REQ-FLA-002: a pending directed request (timer not yet expired, no ack yet)
// does NOT make the channel unavailable — unavailability starts only when the
// timer has expired unanswered.
func TestProbeIgnoresPendingDirectedRequest(t *testing.T) {
	obs := noResponseObs()
	obs.NoResponseAt = nil
	got := EvaluateAvailability(ProbeInput{
		Now:          obs.RequestedAt.Add(5 * time.Minute),
		LeadPeer:     "team-lead",
		OfflineBound: 30 * time.Minute,
		Registry:     []sessionmsg.AgentInfo{probePeer("team-lead", time.Minute)},
		Observations: []Observation{obs},
	})
	if got.Verdict != VerdictAvailable {
		t.Fatalf("verdict = %q, want %q (pending request inside its timer is not unavailability): %s", got.Verdict, VerdictAvailable, got.Reason)
	}
}
