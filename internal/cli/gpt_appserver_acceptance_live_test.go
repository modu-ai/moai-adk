package cli

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/codexapp"
	"github.com/modu-ai/moai-adk/internal/codexbridge"
	"github.com/modu-ai/moai-adk/internal/codextools"
)

type acceptanceInstructionRPC struct {
	*codexapp.Client
	t *testing.T
}

func TestSharedGPTLiveCancelThenFollowup(t *testing.T) {
	if os.Getenv("MOAI_GPT_LIVE") != "1" {
		t.Skip("subscription cancellation requires MOAI_GPT_LIVE=1")
	}
	ctx, stop := context.WithTimeout(context.Background(), 90*time.Second)
	defer stop()
	profile, err := managedGPTProfile()
	if err != nil {
		t.Fatal(err)
	}
	binary, err := exec.LookPath("codex")
	if err != nil {
		t.Fatal(err)
	}
	binary, err = filepath.Abs(binary)
	if err != nil {
		t.Fatal(err)
	}
	client, err := codexapp.ConnectShared(ctx, codexapp.Config{Binary: binary, Home: profile})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.Close() }() // Best-effort teardown; primary assertions own failures.
	if _, err := client.Initialize(ctx, "moai-cancel-acceptance", "1"); err != nil {
		t.Fatal(err)
	}
	store := managedGPTTestStore(t)
	engine, err := codexbridge.New(ctx, client, codexbridge.Config{Store: store})
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()
	q := codexbridge.Request{Owner: codextools.Binding{ConversationID: fmt.Sprintf("cancel-live-%d", time.Now().UnixNano()), AccountScope: "cancel-live"},
		Model: "gpt-5.6-sol", Effort: "low", CWD: t.TempDir(), PrefixDigest: "cancel-1",
		Input: []any{map[string]string{"type": "text", "text": "Write a numbered list of 500 short animal facts. Start immediately. Never call tools."}},
	}
	turnCtx, cancelTurn := context.WithCancel(ctx)
	defer cancelTurn()
	sawText := false
	_, err = engine.StepStream(turnCtx, q, func(delta string) error {
		if delta != "" {
			sawText = true
			cancelTurn()
		}
		return nil
	})
	if !sawText || !errors.Is(err, context.Canceled) {
		t.Fatalf("actual streaming cancel not observed: text=%v error=%v", sawText, err)
	}
	started := time.Now()
	deadline := time.NewTimer(10 * time.Second)
	defer deadline.Stop()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		barrier, found, err := store.Barrier(q.Owner)
		if err != nil {
			t.Fatal(err)
		}
		if found && barrier.Phase == "idle" {
			break
		}
		select {
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		case <-deadline.C:
			t.Fatalf("cancel not recovered: phase=%s", barrier.Phase)
		case <-ticker.C:
		}
	}
	t.Logf("actual text cancellation recovered idle in %s", time.Since(started).Round(time.Millisecond))
	q.ExpectedPrefix, q.PrefixDigest = q.PrefixDigest, "cancel-2"
	q.Input = []any{map[string]string{"type": "text", "text": "Stop the previous task. Reply exactly CANCEL_FOLLOWUP_OK."}}
	seg, err := engine.Step(ctx, q)
	if err != nil || !seg.Done || strings.TrimSpace(seg.Text) != "CANCEL_FOLLOWUP_OK" {
		t.Fatalf("new followup failed: done=%v text=%q error=%v", seg.Done, seg.Text, err)
	}
	t.Logf("same-owner fresh followup=%q; no tools exposed", seg.Text)
}

func (r acceptanceInstructionRPC) Call(ctx context.Context, method string, params any, result any) error {
	err := r.Client.Call(ctx, method, params, result)
	if method == "thread/resume" {
		// Only this test's synthetic instructions and generated thread ID are logged.
		raw, _ := json.Marshal(params)
		r.t.Logf("native %s params=%s error=%v", method, raw, err)
	}
	return err
}

// This acceptance probe uses the already-running product owner only. It sends
// synthetic data, exposes no tools, and never changes the operator's login.
func TestSharedGPTLiveInstructionsModelResumeAndImage(t *testing.T) {
	if os.Getenv("MOAI_GPT_LIVE") != "1" {
		t.Skip("subscription acceptance requires MOAI_GPT_LIVE=1")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()
	profile, err := managedGPTProfile()
	if err != nil {
		t.Fatal(err)
	}
	binary, err := exec.LookPath("codex")
	if err != nil {
		t.Fatal(err)
	}
	binary, err = filepath.Abs(binary)
	if err != nil {
		t.Fatal(err)
	}
	store := managedGPTTestStore(t)
	connect := func() (*codexapp.Client, *codexbridge.Engine) {
		client, err := codexapp.ConnectShared(ctx, codexapp.Config{Binary: binary, Home: profile})
		if err != nil {
			t.Fatalf("shared owner unavailable: %v", err)
		}
		t.Cleanup(func() { _ = client.Close() })
		if _, err := client.Initialize(ctx, "moai-acceptance-live", "1"); err != nil {
			t.Fatal(err)
		}
		account, err := client.Account(ctx)
		if err != nil || account.Account == nil || account.Account.Type != "chatgpt" {
			t.Fatalf("subscription unavailable: %v", err)
		}
		engine, err := codexbridge.New(ctx, acceptanceInstructionRPC{Client: client, t: t}, codexbridge.Config{Store: store})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(engine.Close)
		return client, engine
	}
	client, engine := connect()
	q := codexbridge.Request{
		Owner: codextools.Binding{ConversationID: fmt.Sprintf("acceptance-%d", time.Now().UnixNano()), AccountScope: "acceptance-live"},
		Model: "gpt-5.6-sol", Effort: "low", CWD: t.TempDir(), PrefixDigest: "acceptance-1",
		Instructions: "On each request reply exactly PHASE_ALPHA unless asked to recall the saved word or identify an image color. Never call tools.",
		Input:        []any{map[string]string{"type": "text", "text": "Remember this synthetic word for later: amber-quokka-731. Acknowledge according to the current developer instruction."}},
	}
	run := func(label, expected string) {
		t.Helper()
		started := time.Now()
		var first time.Duration
		seg, err := engine.StepStream(ctx, q, func(delta string) error {
			if delta != "" && first == 0 {
				first = time.Since(started)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("%s: %v", label, err)
		}
		if !seg.Done || len(seg.Tools) != 0 || seg.Tool != nil || !strings.Contains(seg.Text, expected) {
			t.Errorf("%s: expected %q, got done=%v text=%q", label, expected, seg.Done, seg.Text)
		}
		t.Logf("%s: model=%s first_text=%s completed=%s result=%q usage_observed=%v", label, q.Model, first.Round(time.Millisecond), time.Since(started).Round(time.Millisecond), seg.Text, seg.Usage != nil)
		q.ExpectedPrefix = q.PrefixDigest
	}
	run("initial instructions", "PHASE_ALPHA")
	q.PrefixDigest = "acceptance-2"
	q.Instructions = "On each request reply exactly PHASE_BETA unless asked to recall the saved word or identify an image color. Never call tools."
	q.Input = []any{map[string]string{"type": "text", "text": "Respond according to the current developer instruction."}}
	run("completed-turn instructions override", "PHASE_BETA")
	q.PrefixDigest = "acceptance-3"
	q.Model = "gpt-5.6-terra"
	q.Input = []any{map[string]string{"type": "text", "text": "Recall the saved word from the first user message. Reply with that word only."}}
	run("idle model switch retains history", "amber-quokka-731")
	engine.Close()
	if err := client.Close(); err != nil {
		t.Fatal(err)
	}
	_, engine = connect()
	q.Resume, q.PrefixDigest = true, "acceptance-4"
	run("new client durable resume", "amber-quokka-731")
	q.Resume, q.PrefixDigest = false, "acceptance-5"
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			img.SetRGBA(x, y, color.RGBA{R: 255, A: 255})
		}
	}
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, img); err != nil {
		t.Fatal(err)
	}
	q.Input = []any{
		map[string]string{"type": "text", "text": "Identify the dominant color in this image. Reply using exactly one lowercase English color name."},
		map[string]string{"type": "image", "url": "data:image/png;base64," + base64.StdEncoding.EncodeToString(encoded.Bytes())},
	}
	run("native inline image input", "red")
}
