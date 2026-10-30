package monitor

import (
	"context"
	"testing"
)

const (
	sidA = "2a8642ba-c7da-448b-a256-444facf4c514"
	sidB = "b18ce9de-77b3-4fb0-a6c8-571892a1826e"
)

func idRow(name, sid string) Row {
	c := collector(fakeClaims{rows: map[int]ClaimRow{1: {Issue: 1, AgentName: name, SessionID: sid}}}, fakeRemote{}, fakeLedger{})
	return c.Collect(context.Background(), []int{1})[0]
}

func TestAgentNameAndSessionIDShownSeparately(t *testing.T) {
	r := idRow("agent-3", sidA)
	if r.Agent.Value != "agent-3" || r.SessionID.Value != sidA {
		t.Fatalf("agent=%+v session=%+v", r.Agent, r.SessionID)
	}
	if r.SessionID.Redacted || !r.SessionID.Confirmed {
		t.Errorf("a UUID session id must not be treated as an opaque secret: %+v", r.SessionID)
	}
}

func TestSameIdentityJudgedBySessionIDOnly(t *testing.T) {
	if SameIdentity(idRow("agent-3", sidA), idRow("agent-9", sidA)) != true {
		t.Error("same session id under different display names is the same identity")
	}
	if SameIdentity(idRow("agent-3", sidA), idRow("agent-3", sidB)) {
		t.Error("same display name with different session ids is a different identity")
	}
	if SameIdentity(idRow("agent-3", ""), idRow("agent-3", "")) {
		t.Error("unknown session ids can never establish identity")
	}
}
