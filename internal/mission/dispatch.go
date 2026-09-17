package mission

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strconv"
	"sync"
)

type DispatchInput struct {
	MissionID        string
	CardID           string
	CardState        string
	CardRevision     int64
	ExpectedRevision int64
	Lane             string
	LaneAvailable    bool
	LaneOwner        string
	LeaseHeld        bool
}

type DispatchResult struct {
	DispatchID string
	CardID     string
	Lane       string
	Applied    bool
}

type DispatchRegistry struct {
	mu     sync.Mutex
	byCard map[string]DispatchResult
}

func NewDispatchRegistry() *DispatchRegistry {
	return &DispatchRegistry{byCard: map[string]DispatchResult{}}
}

func dispatchIdentity(in DispatchInput) string {
	sum := sha256.Sum256([]byte(in.MissionID + "\x00" + in.CardID + "\x00" + in.Lane + "\x00" + strconv.FormatInt(in.ExpectedRevision, 10)))
	return "dispatch-" + hex.EncodeToString(sum[:16])
}

func (r *DispatchRegistry) Dispatch(in DispatchInput) (DispatchResult, error) {
	if r == nil || in.MissionID == "" || in.CardID == "" || in.CardState != "picked" || in.CardRevision != in.ExpectedRevision {
		return DispatchResult{}, errors.New("mission dispatch: stale_or_invalid_card")
	}
	if in.Lane == "" || !in.LaneAvailable || in.LaneOwner != "" || !in.LeaseHeld {
		return DispatchResult{}, errors.New("mission dispatch: lane_unavailable_or_unowned")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if existing, ok := r.byCard[in.CardID]; ok {
		existing.Applied = false
		return existing, nil
	}
	result := DispatchResult{DispatchID: dispatchIdentity(in), CardID: in.CardID, Lane: in.Lane, Applied: true}
	r.byCard[in.CardID] = result
	return result, nil
}

func (r *DispatchRegistry) Count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.byCard)
}
