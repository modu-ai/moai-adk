package codexbridge

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/codexapp"
)

// completeConversation drives one full tool round so the barrier record lands
// in the idle phase with the owned thread identity persisted.
func completeConversation(t *testing.T, e *Engine, id string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	q := request(id)
	first, err := e.Step(ctx, q)
	if err != nil || first.Tool == nil {
		t.Fatal(first, err)
	}
	q.Input = nil
	q.ExpectedPrefix = "prefix-1"
	q.PrefixDigest = "prefix-2"
	q.Results = []ToolResult{{ID: first.Tool.ID, Content: []Content{{Type: "inputText", Text: "done"}}, Success: true}}
	last, err := e.Step(ctx, q)
	if err != nil || !last.Done {
		t.Fatal(last, err)
	}
}

func resumeRequest(id string) Request {
	q := request(id)
	q.Resume = true
	q.ExpectedPrefix = "prefix-2"
	q.PrefixDigest = "prefix-3"
	q.Input = []any{map[string]string{"type": "text", "text": "continue"}}
	return q
}

func TestResumeAttachesOwnedThreadWithoutRebuild(t *testing.T) {
	e, _, store := fixture(t)
	completeConversation(t, e, "resume-me")
	e.Close()

	freshRPC := &fakeRPC{events: make(chan codexapp.Message, 8)}
	fresh, err := New(context.Background(), freshRPC, Config{Store: store})
	if err != nil {
		t.Fatal(err)
	}
	defer fresh.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	q := resumeRequest("resume-me")
	seg, err := fresh.Step(ctx, q)
	if err != nil || seg.Tool == nil {
		t.Fatal(seg, err)
	}
	if len(freshRPC.resumes) != 1 || freshRPC.resumes[0] != "thread-1" {
		t.Fatal("owned thread not resumed", freshRPC.resumes)
	}
	if freshRPC.next != 0 {
		t.Fatal("resume rebuilt a new thread via thread/start", freshRPC.next)
	}
	q.Input = nil
	q.ExpectedPrefix = "prefix-3"
	q.PrefixDigest = "prefix-4"
	q.Results = []ToolResult{{ID: seg.Tool.ID, Content: []Content{{Type: "inputText", Text: "ok"}}, Success: true}}
	last, err := fresh.Step(ctx, q)
	if err != nil || !last.Done {
		t.Fatal(last, err)
	}
}

func TestResumeRejectsLegacyBarrierRecord(t *testing.T) {
	e, _, store := fixture(t)
	completeConversation(t, e, "legacy")
	e.Close()

	// Degrade the record to the AS3 shape: no Schema, Model or CWD keys.
	path := filepath.Join(store.dir, storeKey(request("legacy").Owner)+".json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]any
	if err = json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"Schema", "Model", "CWD"} {
		delete(fields, key)
	}
	legacy, _ := json.Marshal(fields)
	if err = os.WriteFile(path, legacy, 0600); err != nil {
		t.Fatal(err)
	}

	freshRPC := &fakeRPC{events: make(chan codexapp.Message, 8)}
	fresh, err := New(context.Background(), freshRPC, Config{Store: store})
	if err != nil {
		t.Fatal(err)
	}
	defer fresh.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if _, err = fresh.Step(ctx, resumeRequest("legacy")); !errors.Is(err, ErrRecovery) {
		t.Fatal("legacy barrier accepted for resume", err)
	}
	if len(freshRPC.resumes) != 0 || freshRPC.next != 0 {
		t.Fatal("legacy barrier reached the App Server")
	}
}

func TestResumeRejectsIncompleteBarrier(t *testing.T) {
	e, _, store := fixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	q := request("mid-turn")
	if _, err := e.Step(ctx, q); err != nil {
		t.Fatal(err)
	}
	e.Close()

	freshRPC := &fakeRPC{events: make(chan codexapp.Message, 8)}
	fresh, err := New(context.Background(), freshRPC, Config{Store: store})
	if err != nil {
		t.Fatal(err)
	}
	defer fresh.Close()
	rq := resumeRequest("mid-turn")
	rq.ExpectedPrefix = "prefix-1"
	if _, err = fresh.Step(ctx, rq); !errors.Is(err, ErrRecovery) {
		t.Fatal("mid-turn barrier accepted for resume", err)
	}
	if len(freshRPC.resumes) != 0 || freshRPC.next != 0 {
		t.Fatal("incomplete barrier reached the App Server")
	}
}

func TestResumeRejectsForeignBindings(t *testing.T) {
	for _, kind := range []string{"unknown-conversation", "foreign-account", "wrong-prefix", "wrong-model", "wrong-cwd", "caller-thread-assert"} {
		t.Run(kind, func(t *testing.T) {
			e, _, store := fixture(t)
			completeConversation(t, e, "owned")
			e.Close()

			freshRPC := &fakeRPC{events: make(chan codexapp.Message, 8)}
			fresh, err := New(context.Background(), freshRPC, Config{Store: store})
			if err != nil {
				t.Fatal(err)
			}
			defer fresh.Close()
			rq := resumeRequest("owned")
			switch kind {
			case "unknown-conversation":
				rq.Owner.ConversationID = "other"
			case "foreign-account":
				rq.Owner.AccountScope = "profile-generation-2"
			case "wrong-prefix":
				rq.ExpectedPrefix = "prefix-9"
			case "wrong-model":
				rq.Model = "gpt-6-astra"
			case "wrong-cwd":
				rq.CWD = "/elsewhere"
			case "caller-thread-assert":
				rq.Owner.ThreadID = "thread-1"
			}
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			if _, err = fresh.Step(ctx, rq); err == nil {
				t.Fatal("foreign resume binding accepted")
			}
			if len(freshRPC.resumes) != 0 || freshRPC.next != 0 {
				t.Fatal("foreign binding reached the App Server")
			}
		})
	}
}
