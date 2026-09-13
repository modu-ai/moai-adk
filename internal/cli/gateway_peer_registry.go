package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/profile"
)

// Claude discovers local peers under its config directory's sessions registry.
// Share only that registry with the selected profile, never transcripts or auth.
func shareGatewayPeerRegistry(native string, in gatewayLaunchRequest) error {
	source := profile.GetProfileDir(in.ProfileName)
	if source == "" && in.OriginalConfigSet {
		source = in.OriginalConfig
	}
	if source == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		source = filepath.Join(home, ".claude")
	}
	registry := filepath.Join(source, "sessions")
	if err := os.MkdirAll(registry, 0700); err != nil {
		return err
	}
	registry, err := filepath.EvalSymlinks(registry)
	if err != nil {
		return err
	}
	destination := filepath.Join(native, "sessions")
	if resolved, err := filepath.EvalSymlinks(destination); err == nil {
		if resolved == registry {
			return nil
		}
		return errors.New("gateway peer registry belongs to an older isolated session; start a new moai gpt session to share peers")
	}
	if err := os.Symlink(registry, destination); err != nil {
		if resolved, e := filepath.EvalSymlinks(destination); e == nil && resolved == registry {
			return nil
		}
		return fmt.Errorf("share gateway peer registry: %w", err)
	}
	return nil
}
