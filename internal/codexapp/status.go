package codexapp

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"strings"
	"time"
)

// LoginStatus asks the official CLI to inspect authentication without starting
// another token-refresh owner. Only known status strings leave this boundary.
func LoginStatus(ctx context.Context, cfg Config) (bool, error) {
	return loginStatus(ctx, cfg, []string{"login", "status", "-c", `cli_auth_credentials_store="file"`})
}

func loginStatus(ctx context.Context, cfg Config, args []string) (bool, error) {
	if err := validateConfig(ctx, cfg); err != nil {
		return false, err
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, cfg.Binary, args...)
	configureProcess(cmd, cfg.Home)
	cmd.WaitDelay = time.Second
	var out statusOutput
	cmd.Stdout, cmd.Stderr = &out, &out
	err := cmd.Run()
	if out.overflow {
		return false, errors.New("GPT subscription account status is unavailable")
	}
	status := strings.TrimSpace(out.String())
	if err == nil && status == "Logged in using ChatGPT" {
		return true, nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 && status == "Not logged in" {
		return false, nil
	}
	return false, errors.New("GPT subscription account status is unavailable")
}

type statusOutput struct {
	bytes.Buffer
	overflow bool
}

func (w *statusOutput) Write(p []byte) (int, error) {
	n := len(p)
	remaining := 4096 - w.Len()
	if len(p) > remaining {
		w.overflow = true
		p = p[:remaining]
	}
	_, _ = w.Buffer.Write(p)
	return n, nil
}
