package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

func TestManagedFactoryInputAndInboxClaim(t *testing.T) {
	var input bytes.Buffer
	if err := writeManagedClaudeInput(&input, "한 줄 준비"); err != nil {
		t.Fatal(err)
	}
	var sent struct {
		Type    string `json:"type"`
		Message struct {
			Role, Content string
		} `json:"message"`
	}
	if err := json.Unmarshal(input.Bytes(), &sent); err != nil || sent.Type != "user" || sent.Message.Role != "user" || sent.Message.Content != "한 줄 준비" {
		t.Fatalf("stream input=%q err=%v", input.String(), err)
	}

	t.Setenv("MOAI_HOME", t.TempDir())
	root, run := t.TempDir(), "managed-probe"
	store, err := factorymsg.Open(root, run)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = store.Close() }()
	peer := factorymsg.Peer{ProjectKey: homestate.ProjectKey(root), RunID: run, Backend: "glm", Role: "worker", Slot: "worker-1", SessionUUID: "managed-worker", Generation: 1, PID: os.Getpid(), ProcessStart: homestate.CurrentProcessFingerprint()}
	peer, err = store.RegisterPeer(context.Background(), peer)
	if err != nil {
		t.Fatal(err)
	}
	sender := peer
	sender.Role, sender.Slot, sender.SessionUUID = "lead", "lead", "managed-lead"
	sender.PID = os.Getppid()
	sender.ProcessStart, _ = homestate.ProbeProcessIdentity(sender.PID)
	sender, err = store.RegisterPeer(context.Background(), sender)
	if err != nil {
		t.Fatal(err)
	}
	env, err := store.Send(context.Background(), factorymsg.SendRequest{From: sender, To: peer, Kind: factorymsg.KindStatusRequest, IdempotencyKey: "managed-once", TaskRef: "t1", CorrelationID: "c1", TTL: time.Minute, Payload: []byte("SECRET_BODY")})
	if err != nil {
		t.Fatal(err)
	}
	claims, err := claimManagedFactoryInbox(store, os.Getpid(), peer.ProcessStart)
	if err != nil || len(claims) != 1 || claims[0].ID != env.ID || claims[0].ClaimToken == "" {
		t.Fatalf("claims=%+v err=%v", claims, err)
	}
	prompt := managedFactoryInboxPrompt(run, claims)
	if !strings.Contains(prompt, env.ID) || !strings.Contains(prompt, claims[0].ClaimToken) || strings.Contains(prompt, "SECRET_BODY") {
		t.Fatalf("unsafe or incomplete prompt=%q", prompt)
	}
}
