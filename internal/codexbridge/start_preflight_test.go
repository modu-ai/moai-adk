package codexbridge

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/codexapp"
	"github.com/modu-ai/moai-adk/internal/codextools"
)

func TestNewOwnerPreflightFailureDoesNotPoisonCorrectedRequest(t *testing.T) {
	for _, failure := range []string{"registry", "starting-store"} {
		t.Run(failure, func(t *testing.T) {
			e, rpc, store := fixture(t)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			q := request("manager-lead")
			for i := 1; i < 15; i++ {
				q.Tools = append(q.Tools, codextools.Definition{Name: fmt.Sprintf("lead_tool_%d", i), InputSchema: json.RawMessage(`{"type":"object"}`)})
			}
			c, err := e.get(q)
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(store.dir, storeKey(q.Owner)+".json")
			if failure == "registry" {
				q.Tools[14].InputSchema = json.RawMessage(`not-json`)
			} else if err := os.Mkdir(path, 0700); err != nil {
				t.Fatal(err)
			}
			if _, err := e.Step(ctx, q); err == nil {
				t.Fatal("preflight unexpectedly succeeded")
			}
			if len(rpc.methods) != 0 {
				t.Fatalf("local failure reached RPC: %v", rpc.methods)
			}
			c.mu.Lock()
			phase, prefix, registry := c.phase, c.prefix, c.registry
			c.mu.Unlock()
			if phase != "new" || prefix != "" || registry != nil {
				t.Fatalf("local preflight poisoned pristine owner: phase=%s prefix=%q registry_set=%v", phase, prefix, registry != nil)
			}
			if failure == "registry" {
				q.Tools[14].InputSchema = json.RawMessage(`{"type":"object"}`)
			} else if err := os.Remove(path); err != nil {
				t.Fatal(err)
			}
			seg, err := e.Step(ctx, q)
			if err != nil || seg.Tool == nil {
				t.Fatalf("corrected request could not start: %v %v", seg, err)
			}
			if rpc.next != 1 {
				t.Fatalf("thread starts=%d, want exactly one", rpc.next)
			}
		})
	}
}

type rejectedThreadStartRPC struct {
	*fakeRPC
	starts int
}

func (r *rejectedThreadStartRPC) Call(ctx context.Context, method string, params any, out any) error {
	if method == "thread/start" {
		r.starts++
		if r.starts == 1 {
			return &codexapp.RPCError{Code: -32603}
		}
	}
	return r.fakeRPC.Call(ctx, method, params, out)
}

func TestThreadStartNegativeAckAllowsExplicitRetryWithoutHiddenTurn(t *testing.T) {
	old, rpc, store := fixture(t)
	old.Close()
	rejected := &rejectedThreadStartRPC{fakeRPC: rpc}
	e, err := New(context.Background(), rejected, Config{Store: store})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(e.Close)
	q := request("negative-ack")
	_, err = e.Step(context.Background(), q)
	var rpcErr *codexapp.RPCError
	if !errors.As(err, &rpcErr) || rpcErr.Code != -32603 {
		t.Fatalf("negative ack lost: %v", err)
	}
	c, err := e.get(q)
	if err != nil {
		t.Fatal(err)
	}
	c.mu.Lock()
	phase, prefix, registry, thread := c.phase, c.prefix, c.registry, c.owner.ThreadID
	c.mu.Unlock()
	if phase != "new" || prefix != "" || registry != nil || thread != "" {
		t.Fatalf("negative ack poisoned owner: phase=%s prefix=%q registry=%v thread=%q", phase, prefix, registry != nil, thread)
	}
	if len(rpc.methods) != 0 {
		t.Fatalf("hidden turn after rejected start: %v", rpc.methods)
	}
	barrier, found, err := store.Barrier(q.Owner)
	if err != nil || !found || barrier.Phase != "starting" || barrier.Prefix != q.ExpectedPrefix {
		t.Fatalf("prepare would inherit rejected input as accepted history: %+v %v", barrier, err)
	}
	q.ExpectedPrefix = barrier.Prefix
	freshRPC := &fakeRPC{events: make(chan codexapp.Message, 4)}
	fresh, freshErr := New(context.Background(), freshRPC, Config{Store: store})
	if freshErr != nil {
		t.Fatal(freshErr)
	}
	t.Cleanup(fresh.Close)
	if _, freshErr = fresh.Step(context.Background(), q); !errors.Is(freshErr, ErrRecovery) || len(freshRPC.methods) != 0 {
		t.Fatalf("restart replayed rejected-start barrier: error=%v RPCs=%v", freshErr, freshRPC.methods)
	}
	seg, err := e.Step(context.Background(), q)
	if err != nil || seg.Tool == nil {
		t.Fatalf("explicit retry rejected: %v", err)
	}
	if rejected.starts != 2 || rpc.next != 1 || len(rpc.models) != 1 {
		t.Fatalf("starts=%d allocated=%d turns=%d", rejected.starts, rpc.next, len(rpc.models))
	}
}

type uncertainThreadStartRPC struct {
	*fakeRPC
	starts int
}

func (r *uncertainThreadStartRPC) Call(ctx context.Context, method string, params any, out any) error {
	if method == "thread/start" {
		r.starts++
		return errors.New("thread/start response lost after possible allocation")
	}
	return r.fakeRPC.Call(ctx, method, params, out)
}

func TestUncertainThreadStartRemainsFailedWithoutDuplicateRetry(t *testing.T) {
	e, rpc, store := fixture(t)
	e.Close()
	uncertain := &uncertainThreadStartRPC{fakeRPC: rpc}
	e, err := New(context.Background(), uncertain, Config{Store: store})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(e.Close)
	q := request("uncertain-start")
	if _, err := e.Step(context.Background(), q); err == nil {
		t.Fatal("lost response accepted")
	}
	if _, err := e.Step(context.Background(), q); !errors.Is(err, ErrRecovery) {
		t.Fatalf("uncertain start retried: %v", err)
	}
	if uncertain.starts != 1 {
		t.Fatalf("duplicate thread allocation: %d", uncertain.starts)
	}
	rec, _, err := store.load(q.Owner)
	if err != nil || rec.Phase != "failed" {
		t.Fatalf("missing durable failure: %+v %v", rec, err)
	}
}
