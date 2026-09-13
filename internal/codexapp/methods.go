package codexapp

import (
	"context"
	"encoding/json"
	"errors"
)

type InitializeResult struct {
	UserAgent string `json:"userAgent"`
}

// Initialize opts into dynamic tools; successful handshake alone does not prove
// that a particular experimental field is supported by the installed binary.
func (c *Client) Initialize(ctx context.Context, name, version string) (InitializeResult, error) {
	var result InitializeResult
	if name == "" || version == "" {
		return result, ErrProtocol
	}
	err := c.Call(ctx, "initialize", map[string]any{"clientInfo": map[string]string{"name": name, "version": version}, "capabilities": map[string]bool{"experimentalApi": true}}, &result)
	if err == nil {
		err = c.Notify(ctx, "initialized", map[string]any{})
	}
	return result, err
}

type LoginMode string

const (
	LoginBrowser LoginMode = "chatgpt"
	LoginDevice  LoginMode = "chatgptDeviceCode"
	LoginAPIKey  LoginMode = "apiKey"
)

// LoginResult contains one-time login material. Callers must not persist or log it.
type LoginResult struct {
	Type            string `json:"type"`
	LoginID         string `json:"loginId"`
	AuthURL         string `json:"authUrl"`
	UserCode        string `json:"userCode"`
	VerificationURL string `json:"verificationUrl"`
}

func (c *Client) Login(ctx context.Context, mode LoginMode, apiKey string) (LoginResult, error) {
	var result LoginResult
	params := map[string]string{"type": string(mode)}
	switch mode {
	case LoginBrowser, LoginDevice:
		if apiKey != "" {
			return result, errors.New("API key is not accepted for subscription login")
		}
	case LoginAPIKey:
		if apiKey == "" {
			return result, errors.New("API key required")
		}
		params["apiKey"] = apiKey
	default:
		return result, errors.New("unsupported managed login mode")
	}
	err := c.Call(ctx, "account/login/start", params, &result)
	if err == nil && result.Type != string(mode) {
		return LoginResult{}, ErrProtocol
	}
	return result, err
}
func (c *Client) CancelLogin(ctx context.Context, id string) error {
	if id == "" {
		return ErrProtocol
	}
	return c.Call(ctx, "account/login/cancel", map[string]string{"loginId": id}, nil)
}

type Account struct {
	Type     string `json:"type"`
	PlanType string `json:"planType"`
}
type AccountResult struct {
	Account            *Account `json:"account"`
	RequiresOpenAIAuth bool     `json:"requiresOpenaiAuth"`
}

// Account omits email and never asks Codex to export credentials.
func (c *Client) Account(ctx context.Context) (AccountResult, error) {
	var result AccountResult
	err := c.Call(ctx, "account/read", map[string]bool{"refreshToken": false}, &result)
	return result, err
}
func (c *Client) Logout(ctx context.Context) error { return c.Call(ctx, "account/logout", nil, nil) }

type Model struct {
	ID                        string          `json:"id"`
	Model                     string          `json:"model"`
	DisplayName               string          `json:"displayName"`
	Hidden                    bool            `json:"hidden"`
	DefaultReasoningEffort    string          `json:"defaultReasoningEffort"`
	SupportedReasoningEfforts json.RawMessage `json:"supportedReasoningEfforts"`
}
type ModelPage struct {
	Data       []Model `json:"data"`
	NextCursor *string `json:"nextCursor"`
}

func (c *Client) Models(ctx context.Context, cursor string) (ModelPage, error) {
	params := map[string]any{"limit": 100}
	if cursor != "" {
		params["cursor"] = cursor
	}
	var result ModelPage
	err := c.Call(ctx, "model/list", params, &result)
	return result, err
}
