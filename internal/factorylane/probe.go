// Package factorylane holds the factory-lane autonomy surfaces of
// SPEC-FACTORY-LANE-AUTONOMY-001: the cross-session messaging availability
// probe (fragment 1, design.md D1), the directed-request no-response bound,
// and the fallback-transition event log. The package consumes the sessionmsg
// broker registry read-only; the broker's own state is never written here.
package factorylane

import (
	"fmt"
	"time"

	"github.com/modu-ai/moai-adk/internal/sessionmsg"
)

// Verdict is the probe's channel-availability verdict.
type Verdict string

const (
	// VerdictAvailable reports the messaging channel usable.
	VerdictAvailable Verdict = "available"
	// VerdictUnavailable reports the channel unusable for delegation nudges;
	// the lane's doctrine answer is the declared fallback (REQ-FLA-001).
	VerdictUnavailable Verdict = "channel-unavailable"
)

// ProbeInput carries everything one availability evaluation reads. Registry
// rows come from sessionmsg Store.ListAgents; OfflineBound is the configured
// heartbeat-age bound (config.DefaultSessionMsgAgentOfflineMinutes);
// Observations are the lane's directed-request observations, post-sweep.
type ProbeInput struct {
	Now          time.Time
	LeadPeer     string                 // optional named lead peer to check
	OfflineBound time.Duration          // heartbeat-age availability bound
	Registry     []sessionmsg.AgentInfo // registered peers, read-only
	Observations []Observation          // lane's directed-request observations
}

// Availability is the probe verdict plus the reason a consumer can log. When
// the verdict is channel-unavailable because of an active no-response
// observation, UnavailableUntil carries the RFC3339 instant the bound period
// ends.
type Availability struct {
	Verdict          Verdict `json:"verdict"`
	Reason           string  `json:"reason"`
	UnavailableUntil string  `json:"unavailable_until,omitempty"`
}

// EvaluateAvailability answers one probe question: is the cross-session
// messaging channel usable for delegation nudges? It reads only the inputs it
// is given — the broker registry is consumed read-only by the caller. When a
// lead peer is named, that peer's registration and heartbeat age decide the
// verdict; unnamed, any registered peer counts. The verdict is a report, not
// an error: an unavailable channel is a normal diagnostic outcome (REQ-FLA-001).
//
// @MX:ANCHOR: [AUTO] messaging-availability probe predicate — every fallback declaration evaluates through it
// @MX:REASON: fan_in >= 3 (factory messaging probe verb, fallback declare precondition, M2 pickup consumption); a wrong verdict silently enables or blocks the self-service fallback (REQ-FLA-001).
// @MX:SPEC: SPEC-FACTORY-LANE-AUTONOMY-001
func EvaluateAvailability(in ProbeInput) Availability {
	// REQ-FLA-002: an active no-response observation outranks everything
	// else — the channel is treated unavailable for the remainder of the
	// request's bound period even while the registry itself looks healthy.
	for _, obs := range in.Observations {
		if obs.AckedAt != nil || obs.NoResponseAt == nil {
			continue
		}
		bound, err := time.ParseDuration(obs.Bound)
		if err != nil {
			continue
		}
		until := obs.RequestedAt.Add(bound)
		if in.Now.Before(until) {
			return Availability{
				Verdict: VerdictUnavailable,
				Reason: fmt.Sprintf("no-response observation recorded %s (request sent %s); channel treated unavailable for the remainder of the bound period",
					obs.NoResponseAt.UTC().Format(time.RFC3339), obs.RequestedAt.UTC().Format(time.RFC3339)),
				UnavailableUntil: until.UTC().Format(time.RFC3339),
			}
		}
	}
	if in.LeadPeer != "" {
		for _, rec := range in.Registry {
			if rec.Name != in.LeadPeer {
				continue
			}
			age := in.Now.Sub(rec.LastHeartbeat)
			if in.OfflineBound > 0 && age > in.OfflineBound {
				return Availability{
					Verdict: VerdictUnavailable,
					Reason:  fmt.Sprintf("lead peer %q heartbeat age %s exceeds availability bound %s", in.LeadPeer, age, in.OfflineBound),
				}
			}
			return Availability{
				Verdict: VerdictAvailable,
				Reason:  fmt.Sprintf("lead peer %q registered, heartbeat age %s within bound %s", in.LeadPeer, age, in.OfflineBound),
			}
		}
		return Availability{
			Verdict: VerdictUnavailable,
			Reason:  fmt.Sprintf("lead peer %q not registered in the messaging registry", in.LeadPeer),
		}
	}
	if len(in.Registry) == 0 {
		return Availability{
			Verdict: VerdictUnavailable,
			Reason:  "no peers registered in the messaging registry",
		}
	}
	return Availability{
		Verdict: VerdictAvailable,
		Reason:  fmt.Sprintf("%d peer(s) registered in the messaging registry", len(in.Registry)),
	}
}
