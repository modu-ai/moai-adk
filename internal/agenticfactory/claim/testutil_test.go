package claim

import "github.com/modu-ai/moai-adk/internal/agenticfactory/core"

import "encoding/json"

func jsonUnmarshal(b []byte, v any) error { return json.Unmarshal(b, v) }

// accept is the leader accepting the receipt on the record: its own write, before the
// record can finish.
func accept(r *Record) {
	if r.Receipt != nil {
		r.Receipt.LeaderAccepted = true
	}
}

// finish is a record finished the way the table finishes one: the merge seen, the Issue closed and nothing pending. Setting only
// State = done is a write no edge makes.
func finish(r *Record) {
	r.State, r.Merged, r.Closed, r.Pending = core.StateDone, true, true, ""
}
