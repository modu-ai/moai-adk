package cli

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/modu-ai/moai-adk/internal/codexapp"
	"github.com/spf13/cobra"
)

// startSharedGPTAppServer connects independent Claude sessions to one official
// authentication owner. The owner is detached from the first gateway's lifetime.
// @MX:WARN: [AUTO] Supervisor lifetime intentionally exceeds a gateway connection.
// @MX:REASON: Parallel sessions share one authentication owner; exiting one cannot terminate peers.
func startSharedGPTAppServer(ctx context.Context, cfg codexapp.Config) (*codexapp.Client, error) {
	if ctx == nil {
		return nil, errors.New("GPT App Server context is unavailable")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	bound, err := managedGPTAppServerConfig(cfg.Binary, cfg.Home)
	if err != nil || cfg.ExpectedConfigSHA256 != "" && cfg.ExpectedConfigSHA256 != bound.ExpectedConfigSHA256 {
		return nil, errors.New("GPT App Server profile configuration is unavailable")
	}
	cfg = bound
	if err := codexapp.ValidatePrivatePath(cfg.Home, true); err != nil {
		return nil, errors.New("GPT App Server profile is not private")
	}
	if runtime.GOOS == "windows" {
		return codexapp.Start(ctx, cfg)
	}
	if client, err := codexapp.ConnectShared(ctx, cfg); err == nil {
		return client, nil
	}
	self, err := os.Executable()
	if err != nil {
		return nil, errors.New("GPT supervisor executable unavailable")
	}
	cmd := exec.Command(self, "internal-gpt-appserver", "--home", cfg.Home, "--codex", cfg.Binary, "--config-digest", cfg.ExpectedConfigSHA256)
	cmd.Dir = cfg.Home
	for _, key := range []string{"PATH", "LANG", "LC_ALL", "TMPDIR", "TMP", "TEMP"} {
		if value, ok := os.LookupEnv(key); ok {
			cmd.Env = append(cmd.Env, key+"="+value)
		}
	}
	cmd.Env = append(cmd.Env, "HOME="+cfg.Home, "USERPROFILE="+cfg.Home)
	detachGPTSupervisor(cmd)
	if err := cmd.Start(); err != nil {
		return nil, errors.New("GPT supervisor failed to start")
	}
	// Reap the supervisor if it exits while this gateway remains alive. An
	// existing owner's clients never depend on this child's process lifetime.
	go func() { _ = cmd.Wait() }()
	startup, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	ticker := time.NewTicker(40 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-startup.Done():
			return nil, errors.New("GPT shared App Server unavailable; close legacy moai gpt sessions before retrying")
		case <-ticker.C:
			if client, err := codexapp.ConnectShared(ctx, cfg); err == nil {
				return client, nil
			}
		}
	}
}

// managedGPTAppServerConfig binds the exact optional profile configuration to
// every shared-owner connection and launch. The codexapp boundary validates the
// digest again immediately before use, so a concurrent change fails closed.
func managedGPTAppServerConfig(binary, home string) (codexapp.Config, error) {
	cfg := codexapp.Config{Binary: binary, Home: home}
	path := filepath.Join(home, "config.toml")
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return cfg, nil
	}
	if err != nil || info.Size() > 1<<20 || codexapp.ValidatePrivatePath(path, false) != nil {
		return codexapp.Config{}, errors.New("GPT App Server profile configuration is unavailable")
	}
	file, err := os.Open(path)
	if err != nil {
		return codexapp.Config{}, errors.New("GPT App Server profile configuration is unavailable")
	}
	raw, readErr := io.ReadAll(io.LimitReader(file, (1<<20)+1))
	closeErr := file.Close()
	if readErr != nil || closeErr != nil || len(raw) > 1<<20 {
		return codexapp.Config{}, errors.New("GPT App Server profile configuration is unavailable")
	}
	sum := sha256.Sum256(raw)
	cfg.ExpectedConfigSHA256 = hex.EncodeToString(sum[:])
	return cfg, nil
}

func newGPTAppServerSharedCommand() *cobra.Command {
	var home, binary, digest string
	cmd := &cobra.Command{Use: "internal-gpt-appserver", Hidden: true, Args: cobra.NoArgs, SilenceUsage: true, SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			return codexapp.ServeShared(ctx, codexapp.Config{Binary: binary, Home: home, ExpectedConfigSHA256: digest})
		}}
	cmd.Flags().StringVar(&home, "home", "", "Private managed profile")
	cmd.Flags().StringVar(&binary, "codex", "", "Official Codex executable")
	cmd.Flags().StringVar(&digest, "config-digest", "", "Authorized profile configuration digest")
	return cmd
}
