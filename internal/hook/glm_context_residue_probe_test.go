//go:build residue_probe

package hook

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// Reproducer for what card t802 did NOT repair on the SessionEnd cleanup, kept
// under the `residue_probe` build tag so it stays in history and runnable
// (`go test -tags residue_probe ./internal/hook/`) without failing the default
// suite. It asserts the DESIRED post-cleanup state, so a failure here is the
// remaining defect, not a broken test.
//
// The live-key set IS repaired and guarded in the default suite — see
// glm_settings_cleanup_test.go. Two things are left, both handed to card t888:
//   - the legacy-key tail only the deleted injectGLMEnv ever wrote;
//   - the stranding property: this cleanup gates on ANTHROPIC_BASE_URL, so a
//     file that lost the indicator while keeping GLM keys can never be cleaned.
//     t802 removed the LIVE route into that state (removeGLMEnv now clears the
//     context-window pair together with the indicator), so the state is now
//     reachable only from a file an older binary wrote. Closing the gate itself
//     is a design change — see the negative control in the default-suite guard
//     for why the gate is not simply removable.
//
// All writes stay inside t.TempDir(); no real settings file is touched.
//
// # Card t888 outcome — one probe closed, one deliberately left red
//
// TestProbeCleanupLeavesLegacyAndOtherAxisKeys is CLOSED: SessionEnd now deletes
// the whole canonical cleanup view, legacy tail included.
//
// TestProbeStrandedResidueSurvivesIndicatorLoss is STILL RED, and that is the
// accepted outcome rather than unfinished work. Three things are recorded here
// so a later reader does not reopen a settled decision as a defect:
//
// 1. The indicator was widened CONSERVATIVELY — option (가'). The admitted set
//    is exactly two keys, ANTHROPIC_BASE_URL and MOAI_BACKUP_AUTH_TOKEN. The
//    decision is the lead's, 2026-09-19; it REVISES that same lead's earlier
//    same-day option (가), which admitted three keys, the third being
//    MOAI_STATUSLINE_CONTEXT_SIZE. The revision to two keys is the settled
//    position and is not re-opened by this card.
//
// 2. The stranded case is NOT closed on purpose. Closing it means admitting a
//    file on CLAUDE_CODE_MAX_CONTEXT_TOKENS / CLAUDE_CODE_AUTO_COMPACT_WINDOW —
//    or on MOAI_STATUSLINE_CONTEXT_SIZE, which internal/statusline/memory.go
//    lists as resolution priority #1, "explicit user override", and which
//    internal/config/envkeys.go declares as a general context-size override with
//    GLM as only an example. A user who has never touched GLM can legitimately
//    set any of the three. Admitting one opens the gate on that user's file, and
//    cleanupGLMSettingsLocal's restore branch then takes its `else` and DELETES
//    that user's own ANTHROPIC_AUTH_TOKEN. Worse, the regression is SILENT: the
//    landed negative control TestCleanupGLMSettingsLocalLeavesNonGLMFileAlone
//    seeds no indicator at all, so it keeps passing while the harm happens
//    through an ADMITTED key. Leaving a permanent-but-inert residue is the
//    cheaper failure than deleting a credential the user owns.
//
// 3. The only route into the stranded state is a file an OLDER BINARY wrote.
//    The live route was closed by card t802 (commit 03d1904a7): removeGLMEnv
//    clears the context-window pair together with the indicator, so no current
//    producer can leave a file in this shape. The residue is a historical
//    artifact with no ongoing source.
//
// If this probe ever PASSES, that is not progress — it means the indicator has
// been widened to admit something it must not, and the credential-deletion path
// in item 2 is now reachable. Investigate rather than celebrate.

// TestProbeCleanupLeavesLegacyAndOtherAxisKeys — keys the SessionEnd cleanup
// still does not clear. MOAI_STATUSLINE_CONTEXT_SIZE belongs to the tmux axis
// but reaches settings.local.json via an older binary, so it lands here too.
func TestProbeCleanupLeavesLegacyAndOtherAxisKeys(t *testing.T) {
	scrubGatewayEnv(t)
	root, settingsPath := writeCleanupFixture(t, map[string]string{
		config.EnvAnthropicAuthToken:         "glm-key",
		config.EnvAnthropicBaseURL:           "https://api.z.ai/api/anthropic",
		config.EnvAnthropicDefaultOpusModel:  "glm-5.3",
		config.EnvAnthropicDefaultFableModel: "glm-5.3",
		config.EnvStatuslineContextSize:      "1000000",
		"API_TIMEOUT_MS":                     "3000000",
	})

	cleanupGLMSettingsLocal(root)

	env := readCleanupEnv(t, settingsPath)
	for _, key := range []string{
		config.EnvAnthropicDefaultFableModel,
		config.EnvStatuslineContextSize,
		"API_TIMEOUT_MS",
	} {
		if v, ok := env[key]; ok {
			t.Errorf("cleanupGLMSettingsLocal left %s=%q", key, v)
		}
	}
}

// TestProbeStrandedResidueSurvivesIndicatorLoss — the stranding property. The
// live route into this state is closed as of t802; what remains is a file an
// older binary left behind, which no pass can ever clean while the indicator
// gate stands.
func TestProbeStrandedResidueSurvivesIndicatorLoss(t *testing.T) {
	scrubGatewayEnv(t)
	root, settingsPath := writeCleanupFixture(t, map[string]string{
		config.EnvClaudeCodeMaxContextTokens:  "1000000",
		config.EnvClaudeCodeAutoCompactWindow: "1000000",
	})

	cleanupGLMSettingsLocal(root)

	env := readCleanupEnv(t, settingsPath)
	if v, ok := env[config.EnvClaudeCodeMaxContextTokens]; ok {
		t.Errorf("stranded residue: %s=%q survives with no indicator present",
			config.EnvClaudeCodeMaxContextTokens, v)
	}
}
