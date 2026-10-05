package cli

// doctor_flag_slot.go — card t702.
//
// Advisory detection of the shared flag slot: the machine-global
// cachedGrowthBookFeatures.tengu_harbor_kite boolean in ~/.claude.json that
// gates the cross-session messaging channel. The remote flag evaluation
// endpoint is hardcoded to api.anthropic.com and does not follow
// ANTHROPIC_BASE_URL, so only a first-party session ever writes the slot; a
// third-party backend session (moai glm, any gateway) inherits whatever a
// first-party session last left on this machine. An outage there therefore
// looks like a backend fault and is not one — this check surfaces the actual
// state instead of letting that attribution stand.
//
// Mechanism reference:
// internal/template/templates/.claude/rules/moai/workflow/cross-session-messaging-detail.md
// § The shared flag slot (measurement record included there).
//
// The check is read-only detection and explanation; it never writes the slot,
// never touches CLAUDE_CODE_HARBOR_KITE, and never fails the doctor run —
// every non-OK outcome is a warning, because nothing here is broken in MoAI.

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/config"
)

const (
	// flagSlotCheckName is the doctor --check filter name for this item.
	flagSlotCheckName = "Shared Flag Slot"

	// firstPartyAPIHost is the endpoint whose sessions write the slot. The
	// remote flag evaluation endpoint is hardcoded to this host upstream; it
	// does not follow ANTHROPIC_BASE_URL.
	firstPartyAPIHost = "api.anthropic.com"

	// flagSlotFeaturesKey and flagSlotKey locate the boolean inside
	// ~/.claude.json.
	flagSlotFeaturesKey = "cachedGrowthBookFeatures"
	flagSlotKey         = "tengu_harbor_kite"
)

// doctorFlagSlotClaudeJSON reads the ~/.claude.json contents. Seam for test
// injection, matching the idiom used elsewhere in this package.
var doctorFlagSlotClaudeJSON = func() ([]byte, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	return os.ReadFile(filepath.Join(home, ".claude.json"))
}

// checkFlagSlot reports the shared flag slot state for this session.
//
// Session classification (REQ, card t702): ANTHROPIC_BASE_URL unset or
// pointing at api.anthropic.com is first-party — the session maintains the
// slot itself, so the check stops there without reading the file (this also
// keeps the doctor golden snapshots machine-independent). Any other value is
// third-party, and the slot state decides whether the cross-session messaging
// channel is currently reachable from this machine for this session.
func checkFlagSlot(verbose bool) DiagnosticCheck {
	check := DiagnosticCheck{Name: flagSlotCheckName}

	baseURL := os.Getenv(config.EnvAnthropicBaseURL)
	if flagSlotIsFirstPartyBaseURL(baseURL) {
		check.Status = uikit.CheckOK
		check.Message = "first-party endpoint — this session writes the shared flag slot itself"
		if verbose {
			check.Detail = fmt.Sprintf(
				"the cross-session messaging gate reads %s.%s in ~/.claude.json; a first-party session refreshes it from its own flag evaluation",
				flagSlotFeaturesKey, flagSlotKey)
		}
		return check
	}

	// The gate checks the manual escape hatch BEFORE the slot, so an active
	// override decides the channel regardless of what the slot holds.
	if os.Getenv(config.EnvClaudeCodeHarborKite) != "" {
		check.Status = uikit.CheckOK
		check.Message = fmt.Sprintf("%s override is set — cross-session messaging is forced on regardless of the shared flag slot",
			config.EnvClaudeCodeHarborKite)
		if verbose {
			check.Detail = "the override is an upstream internal flag; its name can change without notice"
		}
		return check
	}

	data, err := doctorFlagSlotClaudeJSON()
	if err != nil {
		check.Status = uikit.CheckWarn
		check.Message = fmt.Sprintf("third-party backend session (%s), but ~/.claude.json cannot be read — shared flag slot state unknown: %v",
			flagSlotDisplayBaseURL(baseURL), err)
		return check
	}

	slot, ok, err := flagSlotExtract(data)
	if err != nil {
		check.Status = uikit.CheckWarn
		check.Message = fmt.Sprintf("third-party backend session (%s), but the shared flag slot value is unreadable in ~/.claude.json: %v",
			flagSlotDisplayBaseURL(baseURL), err)
		if verbose {
			check.Detail = fmt.Sprintf("expected a boolean at %s.%s", flagSlotFeaturesKey, flagSlotKey)
		}
		return check
	}
	if !ok {
		check.Status = uikit.CheckWarn
		check.Message = fmt.Sprintf("third-party backend session (%s), but %s.%s is not present in ~/.claude.json — no first-party session on this machine has written it yet, so treat the cross-session messaging channel as off for this session",
			flagSlotDisplayBaseURL(baseURL), flagSlotFeaturesKey, flagSlotKey)
		return check
	}

	if slot {
		check.Status = uikit.CheckOK
		check.Message = fmt.Sprintf("shared flag slot = true — cross-session messaging is currently on for this third-party session (%s); the slot is machine-global and last-writer-wins, so a first-party session can flip it mid-run",
			flagSlotDisplayBaseURL(baseURL))
		return check
	}

	check.Status = uikit.CheckWarn
	check.Message = fmt.Sprintf("shared flag slot = false — cross-session messaging is off for every session on this machine that cannot write the slot, including this one (%s); a first-party session restores it by refreshing its own flags, or export %s=1 for this session (upstream internal flag, name can change)",
		flagSlotDisplayBaseURL(baseURL), config.EnvClaudeCodeHarborKite)
	if verbose {
		check.Detail = "only a first-party session (api.anthropic.com) writes the slot; this session reads it on every channel call, so the state can change mid-run without any error here"
	}
	return check
}

// flagSlotIsFirstPartyBaseURL reports whether the given ANTHROPIC_BASE_URL
// value targets the first-party endpoint. Empty means unset, which is the
// first-party default.
func flagSlotIsFirstPartyBaseURL(raw string) bool {
	if strings.TrimSpace(raw) == "" {
		return true
	}
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return strings.EqualFold(strings.TrimSpace(raw), "https://"+firstPartyAPIHost)
	}
	return strings.EqualFold(u.Hostname(), firstPartyAPIHost)
}

// flagSlotDisplayBaseURL renders the backend host for check messages.
func flagSlotDisplayBaseURL(raw string) string {
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}
	if u, err := url.Parse(raw); err == nil && u.Hostname() != "" {
		return u.Hostname()
	}
	return raw
}

// flagSlotExtract digs the tengu_harbor_kite boolean out of the
// ~/.claude.json bytes. ok is false when the features object or the slot key
// is absent — distinct from a present-but-unparseable value, which is an
// error.
func flagSlotExtract(data []byte) (slot bool, ok bool, err error) {
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(data, &doc); err != nil {
		return false, false, fmt.Errorf("invalid JSON: %w", err)
	}
	featRaw, present := doc[flagSlotFeaturesKey]
	if !present {
		return false, false, nil
	}
	var feats map[string]json.RawMessage
	if err := json.Unmarshal(featRaw, &feats); err != nil {
		return false, false, fmt.Errorf("%s is not an object: %w", flagSlotFeaturesKey, err)
	}
	slotRaw, present := feats[flagSlotKey]
	if !present {
		return false, false, nil
	}
	if err := json.Unmarshal(slotRaw, &slot); err != nil {
		return false, false, fmt.Errorf("%s.%s is not a boolean: %w", flagSlotFeaturesKey, flagSlotKey, err)
	}
	return slot, true, nil
}
