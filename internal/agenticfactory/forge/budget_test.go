package forge

import (
	"strconv"
	"testing"
	"time"
)

func reset(d time.Duration) string { return strconv.FormatInt(t0.Add(d).Unix(), 10) }

func TestExhaustedResourceDoesNotBlockTheOthers(t *testing.T) {
	b := &Budget{}
	b.Observe(resp(200, "", "x-ratelimit-resource", "search", "x-ratelimit-remaining", "0", "x-ratelimit-reset", reset(time.Hour)), t0)
	b.Observe(resp(200, "", "x-ratelimit-resource", "core", "x-ratelimit-remaining", "4000", "x-ratelimit-reset", reset(time.Hour)), t0)
	if ok, until := b.AllowFor("search", t0); ok || until.Sub(t0) != time.Hour {
		t.Errorf("search: ok=%v until=%v", ok, until.Sub(t0))
	}
	if ok, _ := b.AllowFor("core", t0); !ok {
		t.Error("core was blocked by search")
	}
	if ok, _ := b.AllowFor("graphql", t0); !ok {
		t.Error("an unseen resource must read as unknown, not blocked")
	}
}

func TestBlockingOneResourceLeavesTheOthersOpen(t *testing.T) {
	b := &Budget{}
	b.BlockFor("search", t0.Add(time.Minute))
	if ok, _ := b.AllowFor("search", t0); ok {
		t.Error("search not blocked")
	}
	if ok, _ := b.AllowFor("core", t0); !ok {
		t.Error("core blocked by a search limit")
	}
}

func TestClientLearnsWhichResourceAPathUsesAndBlocksOnlyThat(t *testing.T) {
	f := &fakeT{resps: []Response{
		resp(200, "[]", "x-ratelimit-resource", "search", "x-ratelimit-remaining", "5"),
		resp(429, "", "x-ratelimit-resource", "search", "retry-after", "40"),
		resp(200, "[]", "x-ratelimit-resource", "core", "x-ratelimit-remaining", "4000"),
	}}
	c := client(f, t0)
	if _, err := c.Get(t.Context(), "/search/issues"); err != nil { // learns: /search/issues -> search
		t.Fatal(err)
	}
	c.Now = func() time.Time { return t0.Add(time.Second) }
	r, _ := c.Get(t.Context(), "/search/issues") // 429: only search is blocked
	if !r.Deferred {
		t.Fatal("a 429 must defer")
	}
	again, _ := c.Get(t.Context(), "/search/issues")
	if !again.Deferred || len(f.calls) != 2 {
		t.Errorf("search still sending while blocked: calls=%d", len(f.calls))
	}
	other, err := c.Get(t.Context(), "/repos/o/r/labels")
	if err != nil || other.Deferred || len(f.calls) != 3 {
		t.Errorf("a search limit stopped a core call: %+v err=%v calls=%d", other, err, len(f.calls))
	}
}

func TestReservedBudgetIsOpenOnlyToSafetyClassWrites(t *testing.T) {
	b := &Budget{ReserveCalls: 10}
	b.Observe(resp(200, "", "x-ratelimit-remaining", "10", "x-ratelimit-reset", reset(time.Hour)), t0)
	if ok, _ := b.Allow(t0); ok {
		t.Error("ordinary calls must stop at the reserve")
	}
	if ok, _ := b.AllowReserved("", t0); !ok {
		t.Error("the reserve exists for state and safety writes; they must still go")
	}
	b.Observe(resp(200, "", "x-ratelimit-remaining", "0", "x-ratelimit-reset", reset(time.Hour)), t0)
	if ok, _ := b.AllowReserved("", t0); ok {
		t.Error("nothing may go once the host says remaining is 0")
	}
}
