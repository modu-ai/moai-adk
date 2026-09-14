package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/modu-ai/moai-adk/internal/codexapp"
)

func newGPTAppServerAuthServices(out io.Writer) gptCommandServices {
	withClient := func(ctx context.Context, operation func(context.Context, *codexapp.Client) error) (err error) {
		if ctx == nil {
			return errors.New("GPT App Server context is unavailable")
		}
		profile, err := managedGPTProfile()
		if err != nil {
			return errors.New("GPT App Server profile is unavailable")
		}
		binary, err := exec.LookPath("codex")
		if err != nil {
			return errors.New("install the Codex CLI before using moai gpt")
		}
		binary, err = filepath.Abs(binary)
		if err != nil {
			return errors.New("Codex CLI path is unavailable")
		}
		client, err := codexapp.Start(ctx, codexapp.Config{Binary: binary, Home: profile})
		if err != nil {
			return errors.New("GPT App Server failed to start")
		}
		defer func() {
			if closeErr := client.Close(); err == nil {
				err = closeErr
			}
		}()
		if _, err = client.Initialize(ctx, "moai-gpt-auth", "1"); err != nil {
			return errors.New("GPT App Server initialization failed")
		}
		return operation(ctx, client)
	}
	return gptCommandServices{
		Login: func(ctx context.Context) error {
			loginCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
			defer cancel()
			return withClient(loginCtx, func(ctx context.Context, client *codexapp.Client) error {
				current, err := client.Account(ctx)
				if err == nil && current.Account != nil && current.Account.Type == "chatgpt" {
					_, err = fmt.Fprintf(out, "GPT: logged in via managed App Server (%s)\n", current.Account.PlanType)
					return err
				}
				started, err := client.Login(ctx, codexapp.LoginBrowser, "")
				if err != nil {
					return errors.New("GPT subscription login could not be started")
				}
				parsed, err := url.Parse(started.AuthURL)
				if err != nil || parsed.Scheme != "https" || parsed.Host != "auth.openai.com" || parsed.User != nil || started.LoginID == "" {
					_ = client.CancelLogin(ctx, started.LoginID)
					return errors.New("GPT subscription login URL was rejected")
				}
				if _, err = fmt.Fprintf(out, "Open this URL to log in to GPT:\n%s\n", started.AuthURL); err != nil {
					_ = client.CancelLogin(ctx, started.LoginID)
					return err
				}
				for {
					select {
					case <-ctx.Done():
						_ = client.CancelLogin(context.Background(), started.LoginID)
						return ctx.Err()
					case event, ok := <-client.Events():
						if !ok {
							return errors.New("GPT App Server closed during login")
						}
						if event.Method != "account/login/completed" {
							if len(event.ID) != 0 {
								_ = client.DiscardRequest(event.ID)
							}
							continue
						}
						var completed struct {
							LoginID string `json:"loginId"`
							Success bool   `json:"success"`
						}
						if json.Unmarshal(event.Params, &completed) != nil || completed.LoginID != started.LoginID || !completed.Success {
							return errors.New("GPT subscription login did not complete")
						}
						account, err := client.Account(ctx)
						if err != nil || account.Account == nil || account.Account.Type != "chatgpt" {
							return errors.New("GPT subscription account was not confirmed")
						}
						_, err = fmt.Fprintf(out, "GPT login completed via managed App Server (%s)\n", account.Account.PlanType)
						return err
					}
				}
			})
		},
		Status: func(ctx context.Context) error {
			return withClient(ctx, func(ctx context.Context, client *codexapp.Client) error {
				account, err := client.Account(ctx)
				if err != nil {
					return errors.New("GPT App Server account status is unavailable")
				}
				if account.Account == nil || account.Account.Type != "chatgpt" {
					_, err = fmt.Fprintln(out, "GPT: not logged in")
					return err
				}
				_, err = fmt.Fprintf(out, "GPT: logged in via managed App Server (%s)\n", account.Account.PlanType)
				return err
			})
		},
		Logout: func(ctx context.Context) error {
			return withClient(ctx, func(ctx context.Context, client *codexapp.Client) error {
				if err := client.Logout(ctx); err != nil {
					return errors.New("GPT App Server logout failed")
				}
				_, err := fmt.Fprintln(out, "GPT logout completed via managed App Server")
				return err
			})
		},
	}
}
