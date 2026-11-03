package cli

// N48 first bullet (audit 46dbcce0d4, residual classification): c1 landed at epoch 1, the card was taken back and claimed again
// at epoch 2, and the new holder measured c1 again and offered it. The landing is taken over: the row of c1 is `landed`. The
// verb prints {"state":"waiting"}, and the next `land` answers "already settled as landed". The expectation: what the verb
// prints is the state the row has.

import (
	"encoding/json"
	"testing"
)

func TestAFResidual_N48ThePrintedStateOfATakenOverLandingIsTheStateOfTheRow(t *testing.T) {
	afSkipWithoutSh(t)
	a := afN3Reclaimed(t, false)
	a.measure(a.c1)
	out := a.verb("enqueue", "t1", "--epoch", a.e2, "--candidate", a.c1)
	var printed struct {
		State string `json:"state"`
	}
	if err := json.Unmarshal([]byte(out), &printed); err != nil {
		t.Fatalf("enqueue printed %q: %v", out, err)
	}
	row := string(a.row().State)
	t.Logf("enqueue printed state %q, the row is %q (%.120s)", printed.State, row, out)
	if printed.State != row {
		t.Errorf("enqueue printed state %q and the row of the card is %q", printed.State, row)
	}
}
