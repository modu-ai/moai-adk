package cli

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/profile"
)

// Headless Claude cannot ask for workspace trust and otherwise silently ignores
// project allow rules. Never infer worktree trust from a trusted ancestor.
func requireGatewayPrintTrust(target string, in gatewayLaunchRequest) error {
	headless := false
	for _, arg := range in.Args {
		if arg == "--" {
			break
		}
		if arg == "--print" || arg == "-p" || strings.HasPrefix(arg, "--print=") {
			headless = true
		}
	}
	if !headless {
		return nil
	}
	denied := errors.New("headless gateway requires existing trust for this exact workspace; approve this directory interactively in the selected Claude profile, or resume an already trusted family")
	if !filepath.IsAbs(in.CWD) {
		return denied
	}
	path := filepath.Join(target, ".claude.json")
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Size() > 4<<20 {
		return denied
	}
	f, err := os.Open(path)
	if err != nil {
		return denied
	}
	defer func() { _ = f.Close() }()
	opened, err := f.Stat()
	if err != nil || !os.SameFile(info, opened) {
		return denied
	}
	data, err := io.ReadAll(io.LimitReader(f, (4<<20)+1))
	if err != nil || len(data) > 4<<20 {
		return denied
	}
	var state struct {
		Projects map[string]struct {
			Trusted bool `json:"hasTrustDialogAccepted"`
		} `json:"projects"`
	}
	if json.Unmarshal(data, &state) != nil || !state.Projects[in.CWD].Trusted {
		return denied
	}
	return nil
}

// Seed presentation/onboarding facts and the selected profile's existing trust
// decision for this exact workspace. Never copy credentials or other projects.
func seedGatewayUIState(target string, in gatewayLaunchRequest) error {
	path := filepath.Join(target, ".claude.json")
	if _, err := os.Lstat(path); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	sources := []string{filepath.Join(home, ".claude.json")}
	selected := profile.GetProfileDir(in.ProfileName)
	if selected == "" && in.OriginalConfigSet {
		selected = in.OriginalConfig
	}
	if selected != "" {
		sources = append(sources, filepath.Join(selected, ".claude.json"))
	}
	state := map[string]any{}
	for _, source := range sources {
		f, err := os.Open(source)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}
		data, readErr := io.ReadAll(io.LimitReader(f, 4<<20))
		_ = f.Close() // read-only source; Close carries no write-back to lose
		if readErr != nil {
			return readErr
		}
		var doc map[string]any
		if json.Unmarshal(data, &doc) != nil {
			continue
		}
		for _, key := range []string{"theme", "lastOnboardingVersion"} {
			if value, ok := doc[key].(string); ok && value != "" {
				state[key] = value
			}
		}
		if value, ok := doc["hasCompletedOnboarding"].(bool); ok {
			state["hasCompletedOnboarding"] = value
		}
		// A named profile owns its trust decisions; do not borrow a global
		// approval if the selected profile has never trusted this workspace.
		if source == sources[len(sources)-1] && in.CWD != "" {
			if projects, ok := doc["projects"].(map[string]any); ok {
				if project, ok := projects[in.CWD].(map[string]any); ok {
					if accepted, ok := project["hasTrustDialogAccepted"].(bool); ok {
						state["projects"] = map[string]any{in.CWD: map[string]any{"hasTrustDialogAccepted": accepted}}
					}
				}
			}
		}
	}
	if len(state) == 0 {
		return nil
	}
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if errors.Is(err, os.ErrExist) {
		return nil
	}
	if err != nil {
		return err
	}
	_, writeErr := f.Write(data)
	closeErr := f.Close()
	if writeErr != nil {
		return writeErr
	}
	return closeErr
}

// Carry the selected config's existing bypass-permissions acceptance into a
// fresh family, so an already-accepted warning is not shown again on every
// new gateway conversation. Only the single boolean is copied, and only when
// the selected config (named profile, inherited CLAUDE_CONFIG_DIR, or
// ~/.claude) already recorded it; an existing settings.json is never touched.
func seedGatewayBypassAcceptance(target string, in gatewayLaunchRequest) error {
	path := filepath.Join(target, "settings.json")
	if _, err := os.Lstat(path); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	selected := profile.GetProfileDir(in.ProfileName)
	if selected == "" && in.OriginalConfigSet {
		selected = in.OriginalConfig
	}
	if selected == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		selected = filepath.Join(home, ".claude")
	}
	f, err := os.Open(filepath.Join(selected, "settings.json"))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	data, err := io.ReadAll(io.LimitReader(f, 4<<20))
	_ = f.Close() // read-only source; Close carries no write-back to lose
	if err != nil {
		return err
	}
	var doc map[string]any
	if json.Unmarshal(data, &doc) != nil || doc["skipDangerousModePermissionPrompt"] != true {
		return nil
	}
	out, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if errors.Is(err, os.ErrExist) {
		return nil
	}
	if err != nil {
		return err
	}
	_, writeErr := out.Write([]byte(`{"skipDangerousModePermissionPrompt":true}`))
	closeErr := out.Close()
	if writeErr != nil {
		return writeErr
	}
	return closeErr
}
