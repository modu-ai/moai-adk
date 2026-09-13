package codexbridge

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/modu-ai/moai-adk/internal/codexapp"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestStateStoreRejectsUnsafeExistingEntries(t *testing.T) {
	for _, kind := range []string{"directory", "symlink", "public-file", "missing-directory", "read-only-directory"} {
		t.Run(kind, func(t *testing.T) {
			if runtime.GOOS == "windows" && (kind == "public-file" || kind == "read-only-directory") {
				t.Skip("POSIX permissions; Windows ACL behavior requires CI")
			}
			dir, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			if err = os.Chmod(dir, 0700); err != nil {
				t.Fatal(err)
			}
			store, err := OpenStore(dir)
			if err != nil {
				t.Fatal(err)
			}
			owner := request("store").Owner
			path := filepath.Join(dir, storeKey(owner)+".json")
			switch kind {
			case "directory":
				err = os.Mkdir(path, 0700)
			case "symlink":
				target := filepath.Join(dir, "target")
				err = os.WriteFile(target, []byte("original"), 0600)
				if err == nil {
					err = os.Symlink(target, path)
				}
			case "public-file":
				err = os.WriteFile(path, []byte("original"), 0644)
			case "missing-directory":
				err = os.Remove(dir)
			case "read-only-directory":
				err = os.Chmod(dir, 0500)
				defer func() { _ = os.Chmod(dir, 0700) }() // restore permissions for later cases
			}
			if err != nil {
				t.Fatal(err)
			}
			if err = store.save(record{Owner: owner, Phase: "waiting"}); err == nil {
				t.Fatal("unsafe state barrier accepted")
			}
			if kind != "read-only-directory" {
				if _, err = store.exists(owner); err == nil {
					t.Fatal("unsafe state accepted for recovery lookup")
				}
			}
		})
	}
	if _, err := OpenStore(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("missing store accepted")
	}
}

func TestBridgeRejectsMalformedAndForeignServerRequests(t *testing.T) {
	for _, event := range []codexapp.Message{
		{Method: "notice", Params: json.RawMessage(`not-json`)},
		{ID: json.RawMessage(`"global-rpc"`), Method: "approval", Params: json.RawMessage(`{}`)},
		{ID: json.RawMessage(`"foreign-rpc"`), Method: "item/tool/call", Params: json.RawMessage(`{"threadId":"unknown"}`)},
	} {
		e, rpc, _ := fixture(t)
		rpc.events <- event
		select {
		case <-e.done:
		case <-time.After(5 * time.Second):
			t.Fatal("malformed request did not stop reader")
		}
		if !errors.Is(e.failure(), ErrProtocol) {
			t.Fatal(e.failure())
		}
	}
}

func TestWaitingConversationQueueOverflowFailsClosed(t *testing.T) {
	e, rpc, _ := fixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	q := request("queue")
	if _, err := e.Step(ctx, q); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < e.cfg.QueueSize+1; i++ {
		select {
		case rpc.events <- codexapp.Message{Method: "item/agentMessage/delta", Params: json.RawMessage(`{"threadId":"thread-1","turnId":"turn-thread-1","delta":"late"}`)}:
		case <-e.done:
		}
	}
	select {
	case <-e.done:
	case <-ctx.Done():
		t.Fatal("queue overflow blocked reader")
	}
	if !errors.Is(e.failure(), ErrLimit) {
		t.Fatal(e.failure())
	}
}

func TestRevokedSnapshotAndOversizedResultsNeverReachRPC(t *testing.T) {
	for _, kind := range []string{"revoked", "media", "oversized", "wrong-prefix", "foreign-owner"} {
		t.Run(kind, func(t *testing.T) {
			e, rpc, _ := fixture(t)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			q := request(kind)
			seg, err := e.Step(ctx, q)
			if err != nil {
				t.Fatal(err)
			}
			q.Input = nil
			q.ExpectedPrefix = "prefix-1"
			q.Results = []ToolResult{{ID: seg.Tool.ID, Content: []Content{{Type: "inputText", Text: "done"}}, Success: true}}
			switch kind {
			case "revoked":
				q.Tools = nil
			case "media":
				q.Results[0].Content[0].Type = "inputImage"
			case "oversized":
				q.Results[0].Content[0].Text = string(make([]byte, e.cfg.MaxOutputBytes+1))
			case "wrong-prefix":
				q.ExpectedPrefix = "foreign"
			case "foreign-owner":
				q.Owner.AccountScope = "other-profile"
			}
			if seg, err = e.Step(ctx, q); err == nil || seg.Done {
				t.Fatal(seg, err)
			}
			if rpc.responses != 0 {
				t.Fatal("invalid result reached RPC")
			}
		})
	}
}
