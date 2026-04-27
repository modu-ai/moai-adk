package auth

import "context"

// AuthStatus represents the current authentication status.
type AuthStatus struct {
	Installed     bool   // CLI or tool is installed
	Authenticated bool   // Token/key is configured
	SecretName    string // GitHub secret name
	Message       string // Status message
}

// SecretSetter is the interface for setting GitHub repository secrets.
type SecretSetter interface {
	SetSecret(ctx context.Context, repo, name, value string) error
}

// AuthHandler is the interface for LLM provider authentication handlers.
type AuthHandler interface {
	Setup(ctx context.Context, repo string) error
}
