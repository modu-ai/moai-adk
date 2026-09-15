package codexapp

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// No turn/start: this verifies official local configuration handling without
// provider credentials or model generation.
func TestOfficialIdleDeveloperInstructionOverride(t *testing.T) {
	if os.Getenv("MOAI_CODEX_TRANSPORT_LIVE") != "1" {
		t.Skip("set MOAI_CODEX_TRANSPORT_LIVE=1 for installed local App Server")
	}
	binary, err := exec.LookPath("codex")
	if err != nil {
		t.Fatal(err)
	}
	binary, err = filepath.Abs(binary)
	if err != nil {
		t.Fatal(err)
	}
	home, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(home, 0700); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	client, err := Start(ctx, Config{Binary: binary, Home: home})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			t.Errorf("close instruction probe client: %v", err)
		}
	}()
	if _, err := client.Initialize(ctx, "moai-instruction-probe", "1"); err != nil {
		t.Fatal(err)
	}
	var started struct {
		Thread struct {
			ID string `json:"id"`
		} `json:"thread"`
	}
	if err := client.Call(ctx, "thread/start", map[string]any{"model": "gpt-5.6-sol", "cwd": home, "approvalPolicy": "never", "sandbox": "read-only", "developerInstructions": "Answer concisely", "environments": []any{}}, &started); err != nil {
		t.Fatal(err)
	}
	var resumed struct {
		Thread struct {
			ID string `json:"id"`
		} `json:"thread"`
	}
	if err := client.Call(ctx, "thread/resume", map[string]any{"threadId": started.Thread.ID, "developerInstructions": "Answer with code examples", "excludeTurns": true}, &resumed); err != nil {
		if baselineErr := client.Call(ctx, "thread/resume", map[string]any{"threadId": started.Thread.ID}, &resumed); baselineErr != nil {
			t.Skipf("official empty-thread baseline resume unavailable: baseline=%v override=%v; post-turn override NOT verified by this no-generation probe", baselineErr, err)
		}
		t.Fatal(err)
	}
	if started.Thread.ID == "" || resumed.Thread.ID != started.Thread.ID {
		t.Fatalf("unexpected thread identity: start=%q resume=%q", started.Thread.ID, resumed.Thread.ID)
	}
	t.Log("official App Server accepted thread/start then same-thread developerInstructions override; no model turn requested")
}
