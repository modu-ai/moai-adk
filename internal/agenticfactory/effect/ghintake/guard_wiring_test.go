package ghintake

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect"
)

// A repository name built to steer the request at a settings endpoint is refused before gh runs.
func TestIssueCreateRefusesARepoNameThatReachesASettingsPath(t *testing.T) {
	f := &fakeExec{}
	g := GH{Repo: "o/r/branches/main/protection/x", Run: f.run, AllowWrite: true, ListLimit: 100}
	ex := &effect.Executor{Store: effect.NewMemIntentStore(), Send: g, Lookup: Lookup{GH: g}, Timeouts: effect.Timeouts{IssueCreate: time.Second}}
	in, _ := ex.Execute(t.Context(), IssueRequest(7, g.Repo, "t", "b"))
	if in.Status != effect.StatusParked || in.ParkReason != "permission-denied" || !strings.Contains(in.LastError, string(effect.ReasonWeakensProtection)) || len(f.calls) != 0 {
		t.Errorf("in=%+v calls=%v", in, f.calls)
	}
	if _, err := g.Send(t.Context(), effect.Intent{Kind: effect.IssueCreate, Payload: `{"title":"t"}`}); !errors.Is(err, effect.ErrRefused) {
		t.Errorf("err = %v", err)
	}
}
