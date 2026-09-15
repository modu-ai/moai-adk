package cli

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/gateway"
)

func TestGatewayProcessContractPortableHelper(t *testing.T) {
	mode := os.Getenv("MOAI_GATEWAY_PROCESS_CONTRACT_HELPER")
	if mode == "" {
		return
	}
	cfg, err := gateway.ReadChildConfig(os.Stdin)
	if err != nil {
		os.Exit(41)
	}
	if mode == "exit23" {
		_ = json.NewEncoder(os.Stdout).Encode(gateway.ChildHandoff{Address: "127.0.0.1:43119"})
		time.Sleep(150 * time.Millisecond)
		os.Exit(23)
	}
	err = gateway.RunChildWithControl(context.Background(), cfg, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "ready")
	}), os.Stdout, os.Stdin)
	if err != nil {
		os.Exit(42)
	}
	os.Exit(0)
}

func TestGatewayProcessContractPortable(t *testing.T) {
	options := func(mode string) gateway.StartOptions {
		return gateway.StartOptions{
			Executable:     os.Args[0],
			Args:           []string{"-test.run=^TestGatewayProcessContractPortableHelper$"},
			Env:            append(os.Environ(), "MOAI_GATEWAY_PROCESS_CONTRACT_HELPER="+mode),
			StartupTimeout: 3 * time.Second,
			Config: gateway.ChildConfig{
				Lifetime:     5 * time.Second,
				PollInterval: 20 * time.Millisecond,
				Overlay:      json.RawMessage(`{"model":"gpt-5.6-sol"}`),
			},
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	child, err := gateway.StartChild(ctx, options("serve"))
	if err != nil {
		t.Fatalf("start portable fake child: %v", err)
	}
	t.Run("fake-child-readiness", func(t *testing.T) {
		resp, getErr := (&http.Client{Timeout: time.Second}).Get("http://" + child.Address)
		if getErr != nil {
			t.Fatalf("portable readiness GET: %v", getErr)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("close readiness response: %v", err)
			}
		}()
		raw, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK || string(raw) != "ready" {
			t.Errorf("portable readiness status=%d body=%q, want 200/ready", resp.StatusCode, raw)
		}
	})
	t.Run("flush-atomic-replace", func(t *testing.T) {
		raw, readErr := os.ReadFile(child.OverlayPath)
		if readErr != nil || string(raw) != `{"model":"gpt-5.6-sol"}` {
			t.Errorf("portable overlay readback=%q err=%v, want flushed settings", raw, readErr)
		}
	})
	t.Run("permission", func(t *testing.T) {
		info, statErr := os.Stat(child.OverlayPath)
		if statErr != nil {
			t.Fatalf("portable overlay permission stat: %v", statErr)
		}
		if got := info.Mode().Perm(); got != 0600 {
			t.Errorf("portable overlay permission=%04o, want 0600", got)
		}
	})
	t.Run("cancel", func(t *testing.T) {
		stopCtx, stopCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer stopCancel()
		if stopErr := child.Stop(stopCtx); stopErr != nil {
			t.Errorf("portable cancel: %v", stopErr)
		}
		if _, statErr := os.Stat(child.OverlayPath); !os.IsNotExist(statErr) {
			t.Errorf("portable cancel overlay remains: %v", statErr)
		}
	})

	exitChild, err := gateway.StartChild(ctx, options("exit23"))
	if err != nil {
		t.Fatalf("start exit23 fake child: %v", err)
	}
	var waitErr error
	t.Run("wait", func(t *testing.T) {
		waitCtx, waitCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer waitCancel()
		waitErr = exitChild.Wait(waitCtx)
		if waitErr == nil {
			t.Errorf("portable gateway wait error=nil, want exit status 23")
		}
	})
	t.Run("exit-code", func(t *testing.T) {
		code := 0
		var processExit interface{ ExitCode() int }
		if errors.As(waitErr, &processExit) {
			code = processExit.ExitCode()
		}
		if code != 23 {
			t.Errorf("windows spawn-and-wait exit code=%d err=%v, want 23", code, waitErr)
		}
	})
}
