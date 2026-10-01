package homestate

import (
	"context"
	"database/sql"
)

// LaneCapacityDerived is the runs.lane_capacity value recorded for a run
// whose leader start carried no operator-supplied lane count
// (SPEC-CODEX-LANE-SLOTS-001 REQ-004): the explicit derived-capacity marker.
// A lane joining such a run treats it as capacity-open and applies the
// growth rule (REQ-005); a positive value is the operator-declared count and
// keeps the hard bound (REQ-006).
// @MX:NOTE: [AUTO] zero is the capacity-open marker, not an unset datum — a join reading it must grow the scan, never fall back to a launcher-side bound
// @MX:SPEC: SPEC-CODEX-LANE-SLOTS-001
const LaneCapacityDerived = 0

// RunLaneCapacity reads a run's recorded lane capacity
// (SPEC-CODEX-LANE-SLOTS-001 REQ-004). found is false when the run has no
// record at all — a caller that treats absence as its launcher-side bound
// preserves the pre-SPEC behavior for unrecorded runs. capacity is
// LaneCapacityDerived for a capacity-open run and the operator-declared
// count otherwise.
func (f *FactoryDB) RunLaneCapacity(ctx context.Context, runID string) (capacity int, found bool, err error) {
	err = f.DB.QueryRowContext(ctx, `SELECT lane_capacity FROM runs WHERE run_id=?`, runID).Scan(&capacity)
	if err == sql.ErrNoRows {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return capacity, true, nil
}
