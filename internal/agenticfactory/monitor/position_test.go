package monitor

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// A record whose position could not be derived must not read as "no park" / "no wait" / "no gate": the
// position fields are not known, and the row says so by leaving them unconfirmed and naming the reason.
func TestAnUnknownPositionLeavesEveryPositionFieldUnconfirmed(t *testing.T) {
	r := stageRow(ClaimRow{
		State: "unheard-of", PositionUnknown: "the state \"unheard-of\" maps to no stage",
		// even if the adapter had values for them, they come from a position that is not known
		Stage: string(core.StageBuilding), SubStage: string(core.SubFix),
		Park: &ParkInfo{Reason: "x"}, Gate: &GateInfo{Reason: "y"}, Refusal: &RefusalInfo{Reason: "z"}, Wait: sampleWait(),
	})
	if !r.PositionUnknown.Confirmed || !strings.Contains(r.PositionUnknown.Value, "unheard-of") || r.PositionUnknown.Source != SrcClaimStore {
		t.Errorf("position_unknown = %+v, want the reason, confirmed", r.PositionUnknown)
	}
	for name, f := range map[string]struct {
		confirmed bool
		empty     bool
	}{
		"stage":     {r.Stage.Confirmed, r.Stage.Value == ""},
		"sub_stage": {r.SubStage.Confirmed, r.SubStage.Value == ""},
		"park":      {r.Park.Confirmed, r.Park.Value == nil},
		"gate":      {r.Gate.Confirmed, r.Gate.Value == nil},
		"refusal":   {r.Refusal.Confirmed, r.Refusal.Value == nil},
		"wait":      {r.Wait.Confirmed, r.Wait.Value == nil},
	} {
		if f.confirmed || !f.empty {
			t.Errorf("%s: confirmed=%v empty=%v, want unconfirmed and empty", name, f.confirmed, f.empty)
		}
	}
	// what is not about the position is still shown
	if !r.Owner.Confirmed || !r.State.Confirmed || r.State.Value != "unheard-of" {
		t.Errorf("owner/state must stay shown: %+v %+v", r.Owner, r.State)
	}
}

func TestAKnownPositionIsConfirmedAndCarriesNoUnknownMarker(t *testing.T) {
	r := stageRow(ClaimRow{Stage: string(core.StageReviewing), Wait: sampleWait()})
	if !r.Stage.Confirmed || !r.Park.Confirmed || !r.Wait.Confirmed || r.Wait.Value == nil {
		t.Errorf("a derivable position must stay confirmed: %+v", r)
	}
	if r.PositionUnknown.Value != "" || r.PositionUnknown.Confirmed == false {
		// no marker: an empty, confirmed value means "the position is known"
		t.Errorf("position_unknown = %+v, want empty and confirmed", r.PositionUnknown)
	}
}

func TestThePositionUnknownReasonIsMaskedAndTheJSONKeyIsStable(t *testing.T) {
	secret := "ghp_abcdefghijklmnopqrstuvwxyz0123456789"
	r := stageRow(ClaimRow{PositionUnknown: "cannot derive: token " + secret})
	if strings.Contains(r.PositionUnknown.Value, secret) {
		t.Errorf("secret survived: %q", r.PositionUnknown.Value)
	}
	raw, _ := json.Marshal(r)
	var m map[string]any
	_ = json.Unmarshal(raw, &m)
	if _, ok := m["position_unknown"].(map[string]any); !ok {
		t.Errorf("no position_unknown key in %s", raw)
	}
}
