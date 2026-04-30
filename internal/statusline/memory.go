package statusline

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"gopkg.in/yaml.v3"
)

// defaultAutoCompactPct is the default auto-compact trigger threshold percentage.
// Claude Code triggers auto-compact at approximately 80-85% of the total context window.
// This can be overridden by the CLAUDE_AUTOCOMPACT_PCT_OVERRIDE environment variable.
const defaultAutoCompactPct = 85

// glmContextWindows maps known non-Claude (typically GLM-family) model names
// served via the Anthropic-compatible endpoint to their actual context window
// sizes in tokens. Claude Code reports `context_window_size` based on the
// Claude model slot (Opus = 1M, Sonnet/Haiku = 200K), but the underlying GLM
// model has a different limit. Issue #653.
//
// Entries use lowercase substring matching against the resolved model name.
// Add entries here when a new GLM model ships; users can also override per
// invocation via MOAI_STATUSLINE_CONTEXT_SIZE.
var glmContextWindows = map[string]int{
	"glm-5.1":     200_000, // GLM-5.1 (z.ai) — actual ~230K, leave headroom
	"glm-5":       128_000,
	"glm-4.7":     128_000,
	"glm-4.6":     128_000,
	"glm-4.5":     128_000,
	"glm-4.5-air": 128_000,
}

// getAutoCompactThreshold returns the auto-compact trigger percentage.
// Reads CLAUDE_AUTOCOMPACT_PCT_OVERRIDE env var if set, otherwise uses default.
func getAutoCompactThreshold() int {
	if override := os.Getenv(config.EnvClaudeAutoCompactPct); override != "" {
		if v, err := strconv.Atoi(override); err == nil && v > 0 && v <= 100 {
			return v
		}
	}
	return defaultAutoCompactPct
}

// llmYAMLShape is the minimal shape needed to read
// `llm.glm.context_windows` from .moai/config/sections/llm.yaml. Keeping a
// local shape here avoids pulling the heavyweight config.Manager into the
// statusline hot path.
type llmYAMLShape struct {
	LLM struct {
		GLM struct {
			ContextWindows map[string]int `yaml:"context_windows"`
		} `yaml:"glm"`
	} `yaml:"llm"`
}

// readLLMYAMLContextWindows reads .moai/config/sections/llm.yaml from the
// closest ancestor directory and returns the user-configured
// glm.context_windows map. Returns nil on any error (file missing, parse
// error) so that the caller falls back to built-in defaults. Issue #653.
func readLLMYAMLContextWindows() map[string]int {
	dir, err := os.Getwd()
	if err != nil {
		return nil
	}
	for {
		path := filepath.Join(dir, ".moai", "config", "sections", "llm.yaml")
		if data, readErr := os.ReadFile(path); readErr == nil {
			var shape llmYAMLShape
			if yaml.Unmarshal(data, &shape) == nil && len(shape.LLM.GLM.ContextWindows) > 0 {
				return shape.LLM.GLM.ContextWindows
			}
			return nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return nil
		}
		dir = parent
	}
}

// matchContextWindow returns the largest-matching context window entry for
// the given lowercase model name, or 0 when no entry matches. Longer keys are
// preferred to avoid "glm-4.5" masking "glm-4.5-air".
func matchContextWindow(model string, table map[string]int) int {
	var matched int
	var matchedSize int
	for name, size := range table {
		if strings.Contains(model, strings.ToLower(name)) && len(name) > matched {
			matched = len(name)
			matchedSize = size
		}
	}
	return matchedSize
}

// ResolveGLMContextWindow returns the context window size in tokens for the
// given GLM model name, or 0 if no match is found. The lookup honors
// project-level overrides in `.moai/config/sections/llm.yaml`
// (`glm.context_windows` map) before falling back to the built-in
// `glmContextWindows` table.
//
// This is the single source of truth used by `moai cg` / `moai glm` to
// pre-compute MOAI_STATUSLINE_CONTEXT_SIZE for tmux session env injection
// (Issue #742). It is exported so callers outside the statusline package can
// resolve a model's context window without reimplementing the lookup priority.
func ResolveGLMContextWindow(model string) int {
	val := strings.ToLower(strings.TrimSpace(model))
	if val == "" || strings.HasPrefix(val, "claude") {
		return 0
	}

	// Priority 1: project-level llm.yaml overrides.
	if userTable := readLLMYAMLContextWindows(); len(userTable) > 0 {
		if size := matchContextWindow(val, userTable); size > 0 {
			return size
		}
	}

	// Priority 2: built-in glmContextWindows table.
	return matchContextWindow(val, glmContextWindows)
}

// resolveContextWindowOverride returns a non-zero context window override when
// the user is running through GLM (or another non-Claude provider) so that the
// memory gauge reflects the actual provider limit rather than the Claude slot's
// nominal size. Issue #653 — GLM "opus" slot reported as 1M but real limit is
// 128-230K depending on the GLM model.
//
// Resolution priority:
//  1. MOAI_STATUSLINE_CONTEXT_SIZE env (explicit user override)
//  2. llm.yaml `glm.context_windows` map (project-level configuration)
//  3. ANTHROPIC_DEFAULT_*_MODEL env matches built-in glmContextWindows table
//  4. Return 0 (caller falls back to stdin's context_window_size)
func resolveContextWindowOverride() int {
	if v := os.Getenv(config.EnvStatuslineContextSize); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}

	envSlots := []string{
		config.EnvAnthropicDefaultOpusModel,
		config.EnvAnthropicDefaultSonnetModel,
		config.EnvAnthropicDefaultHaikuModel,
	}

	// Priority 2: project-level llm.yaml overrides.
	userTable := readLLMYAMLContextWindows()
	if len(userTable) > 0 {
		for _, key := range envSlots {
			val := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
			if val == "" || strings.HasPrefix(val, "claude") {
				continue
			}
			if size := matchContextWindow(val, userTable); size > 0 {
				return size
			}
		}
	}

	// Priority 3: built-in glmContextWindows table.
	for _, key := range envSlots {
		val := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
		if val == "" || strings.HasPrefix(val, "claude") {
			continue
		}
		if size := matchContextWindow(val, glmContextWindows); size > 0 {
			return size
		}
	}
	return 0
}

// CollectMemory extracts context window token usage from stdin data.
// Returns a MemoryData with Available=false if input or context_window is nil.
// Priority (following Claude Code documentation):
// 1. Use pre-calculated used_percentage from Claude Code (most accurate)
// 2. Calculate from current_usage tokens
// 3. Fall back to legacy used/total fields
//
// The token budget is scaled to the auto-compact threshold so that the CW bar
// reaches 100% at the point where auto-compact triggers, not at the full context window.
func CollectMemory(input *StdinData) *MemoryData {
	if input == nil || input.ContextWindow == nil {
		return &MemoryData{Available: false}
	}

	ctx := input.ContextWindow

	// Get context window size (default 200K).
	// GLM/non-Claude providers may have a smaller real limit than the Claude
	// slot's reported context_window_size — apply override when detected
	// (issue #653).
	contextSize := ctx.ContextWindowSize
	if contextSize <= 0 {
		contextSize = ctx.Total
	}
	if contextSize <= 0 {
		contextSize = 200000 // Default context window size
	}
	if override := resolveContextWindowOverride(); override > 0 {
		contextSize = override
	}

	// Scale budget to auto-compact threshold so bar shows 100% at compact point
	threshold := getAutoCompactThreshold()
	effectiveBudget := contextSize * threshold / 100

	var tokensUsed int

	// Priority 1: Use pre-calculated percentage from Claude Code
	if ctx.UsedPercentage != nil {
		// Calculate tokens from percentage (of full context window)
		tokensUsed = int(float64(contextSize) * (*ctx.UsedPercentage) / 100.0)
		return &MemoryData{
			TokensUsed:  tokensUsed,
			TokenBudget: effectiveBudget,
			Available:   true,
		}
	}

	// Priority 2: Calculate from current_usage tokens
	if ctx.CurrentUsage != nil {
		cu := ctx.CurrentUsage
		tokensUsed = cu.InputTokens + cu.CacheCreationTokens + cu.CacheReadTokens
		return &MemoryData{
			TokensUsed:  tokensUsed,
			TokenBudget: effectiveBudget,
			Available:   true,
		}
	}

	// Priority 3: Fall back to legacy used/total fields
	if ctx.Used > 0 || ctx.Total > 0 {
		return &MemoryData{
			TokensUsed:  ctx.Used,
			TokenBudget: ctx.Total * threshold / 100,
			Available:   true,
		}
	}

	// No data available - return 0% (session start state)
	return &MemoryData{
		TokensUsed:  0,
		TokenBudget: effectiveBudget,
		Available:   true,
	}
}

// contextUsageLevel determines the color severity level based on
// context window usage percentage.
// Returns levelOk for <50%, levelWarn for 50-80%, levelError for >=80%.
func contextUsageLevel(used, total int) contextLevel {
	if total <= 0 {
		return levelOk
	}

	pct := used * 100 / total

	switch {
	case pct >= 80:
		return levelError
	case pct >= 50:
		return levelWarn
	default:
		return levelOk
	}
}

// usagePercent calculates the percentage of context window used.
// Returns 0 if total is zero to avoid division by zero. Capped at 100.
func usagePercent(used, total int) int {
	if total <= 0 {
		return 0
	}
	pct := used * 100 / total
	if pct > 100 {
		return 100
	}
	return pct
}
