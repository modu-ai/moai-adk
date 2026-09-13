package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/modu-ai/moai-adk/internal/gateway/auth"
	"github.com/modu-ai/moai-adk/internal/paths"
)

// openGPTAuthStore shares the existing owned-home resolution between auth CLI
// operations and the private gateway child. It never consults CODEX_HOME.
func openGPTAuthStore() (*auth.Store, error) {
	home, err := paths.MoaiHome()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(home, 0700); err != nil {
		return nil, err
	}
	home, err = filepath.EvalSymlinks(home)
	if err != nil {
		return nil, err
	}
	return auth.OpenStore(filepath.Join(home, "gateway-auth"))
}

func newGPTAuthServices(out io.Writer, broker auth.Broker) gptCommandServices {
	withStore := func(ctx context.Context, operation func(*auth.Store) error) error {
		store, err := openGPTAuthStore()
		if err != nil {
			return err
		}
		defer store.Close()
		return operation(store)
	}
	return gptCommandServices{
		Login: func(ctx context.Context) error {
			if broker == nil {
				return errors.New("GPT login broker is unavailable")
			}
			return withStore(ctx, func(store *auth.Store) error {
				_, err := store.Login(ctx, broker)
				if err == nil {
					_, err = fmt.Fprintln(out, "GPT login completed")
				}
				return err
			})
		},
		Logout: func(ctx context.Context) error {
			return withStore(ctx, func(store *auth.Store) error {
				logoutBroker, _ := broker.(auth.LogoutBroker)
				result, err := store.LogoutWithBroker(ctx, logoutBroker)
				if result.LocalCompleted {
					_, outputErr := fmt.Fprintf(out, "GPT local logout completed; broker logout %s; remote revocation %s\n", result.BrokerOutcome, result.RemoteOutcome)
					if outputErr != nil {
						return outputErr
					}
				}
				return err
			})
		},
		Status: func(ctx context.Context) error {
			return withStore(ctx, func(store *auth.Store) error {
				status, err := store.Status(ctx)
				if err != nil {
					return err
				}
				state := "not logged in"
				if status.LoggedIn {
					state = "logged in"
				}
				_, err = fmt.Fprintf(out, "GPT: %s\n", state)
				return err
			})
		},
	}
}

// installedGPTBroker resolves the official executable when login/logout is requested,
// so help/status never require a Codex installation or start a broker process.
type installedGPTBroker struct{ Out io.Writer }

func (b installedGPTBroker) Logout(ctx context.Context, home string) error {
	executable, err := exec.LookPath("codex")
	if err != nil {
		return errors.New("GPT logout broker is unavailable")
	}
	executable, err = filepath.Abs(executable)
	if err != nil {
		return errors.New("GPT logout broker is unavailable")
	}
	return (auth.CodexBroker{Executable: executable, Timeout: 30 * time.Second}).Logout(ctx, home)
}

func (b installedGPTBroker) Run(ctx context.Context, home string, refresh bool) error {
	executable, err := exec.LookPath("codex")
	if err != nil {
		return errors.New("install the Codex CLI before running moai gpt login")
	}
	executable, err = filepath.Abs(executable)
	if err != nil {
		return err
	}
	broker := auth.CodexBroker{Executable: executable, Timeout: 10 * time.Minute, OnLogin: func(prompt auth.LoginPrompt) error {
		_, err := fmt.Fprintf(b.Out, "Open this URL to log in to GPT:\n%s\n", prompt.URL)
		return err
	}}
	return broker.Run(ctx, home, refresh)
}
