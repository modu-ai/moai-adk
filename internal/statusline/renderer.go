package statusline

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/modu-ai/moai-adk/internal/config"
)

// Renderer formats StatusData into a multiline statusline string.
// Supports v3 layouts: default(3L), full(5L).
type Renderer struct {
	separator     string
	noColor       bool
	mutedStyle    lipgloss.Style
	segmentConfig map[string]bool
	theme         Theme
}

// NewRenderer creates a Renderer with the specified theme, color mode, and
// segment configuration. When segmentConfig is nil or empty, all segments
// are displayed (backward compatible).
func NewRenderer(themeName string, noColor bool, segmentConfig map[string]bool) *Renderer {
	theme := NewTheme(themeName)

	r := &Renderer{
		// v3 separator: U+2502 box drawing vertical line
		separator:     " │ ",
		noColor:       noColor,
		segmentConfig: segmentConfig,
		theme:         theme,
	}

	if noColor {
		r.mutedStyle = lipgloss.NewStyle()
		return r
	}

	// Set muted style from theme color (REQ-SLE-017)
	r.mutedStyle = lipgloss.NewStyle().Foreground(theme.Muted())

	return r
}

// Render formats the StatusData into the canonical 3-line statusline layout.
//
// The mode argument is accepted for backward compatibility but always
// collapses to ModeDefault via NormalizeMode — the 5-line "Full" layout was
// retired.
//
// @MX:ANCHOR: [AUTO] Single entry point for all mode rendering - called from Build() in builder.go
// @MX:REASON: [AUTO] Public API boundary; keeps the mode parameter so external
// callers compile unchanged while the layout is fixed to default.
func (r *Renderer) Render(data *StatusData, mode StatuslineMode) string {
	if data == nil {
		return "MoAI"
	}
	_ = NormalizeMode(mode) // legacy-name collapse, kept for symmetry
	result := r.renderDefaultV3(data)
	if result == "" {
		return "MoAI"
	}
	return result
}

// filterEmpty removes empty strings from a slice.
func filterEmpty(sections []string) []string {
	filtered := make([]string, 0, len(sections))
	for _, s := range sections {
		if s != "" {
			filtered = append(filtered, s)
		}
	}
	return filtered
}

// isSegmentEnabled checks whether a segment should be rendered based on config.
// Returns true (enabled) when segmentConfig is nil/empty (backward compatible),
// or when the key is not present in the config (unknown segments default to enabled).
func (r *Renderer) isSegmentEnabled(key string) bool {
	if len(r.segmentConfig) == 0 {
		return true
	}
	enabled, exists := r.segmentConfig[key]
	if !exists {
		return true
	}
	return enabled
}

// joinSegments filters a segment slice and joins them with the separator.
// Returns empty string if all segments are empty.
func (r *Renderer) joinSegments(segments []string) string {
	filtered := filterEmpty(segments)
	if len(filtered) == 0 {
		return ""
	}
	return strings.Join(filtered, r.separator)
}

// ─────────────────────────────────────────────────────────────────────────────
// v3 layout renderers
// ─────────────────────────────────────────────────────────────────────────────

// renderDefaultV3 renders the default mode 3-line layout.
//
// L1: 🤖 Model │ 🔅 v2.1.50 │ 🗿 v2.8.0 │ ⏳ 2h 34m │ 💬 MoAI
// L2: CW: 🪫 ██████████ 88% │ 5H: 🔋 ██████████ 45% │ 7D: 🪫 ██████████ 82%
// L3: 📁 moai-adk-go │ 🅱️ feat/auth ↑2↓1 │ 📊 +3 M2 ?1
func (r *Renderer) renderDefaultV3(data *StatusData) string {
	var lines []string

	// L1: model, Claude version, MoAI version, session time, output style
	l1 := r.renderInfoLine(data, false)
	if l1 != "" {
		lines = append(lines, l1)
	}

	// L2: CW/5H/7D bars inline (10 blocks) - always show all 3 bars
	l2 := r.renderBarsInline(data, 10)
	if l2 != "" {
		lines = append(lines, l2)
	}

	// L3: directory, branch, git status
	l3 := r.renderDirGitLine(data)
	if l3 != "" {
		lines = append(lines, l3)
	}

	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n")
}

// ─────────────────────────────────────────────────────────────────────────────
// Common line renderers
// ─────────────────────────────────────────────────────────────────────────────

// renderInfoLine renders the L1 info line (shared by default/full).
// withPrefix=true: full mode format ("Claude v...", "MoAI v...")
// withPrefix=false: default mode format ("v...")
func (r *Renderer) renderInfoLine(data *StatusData, withPrefix bool) string {
	var segs []string

	// Model
	if r.isSegmentEnabled(SegmentModel) && data.Metrics.Available && data.Metrics.Model != "" {
		segs = append(segs, fmt.Sprintf("🤖 %s", data.Metrics.Model))
	}

	// Effort/thinking indicator (Claude Code v2.1.122+, REQ-CC2122-001/002)
	if r.isSegmentEnabled(SegmentEffortThinking) {
		if et := renderEffortThinking(data); et != "" {
			segs = append(segs, et)
		}
	}

	// Cache-hit-ratio indicator (SPEC-TOKEN-EFFICIENCY-001 P0-2, REQ-TEF-005/007).
	// Graceful degradation (renderCacheHit returns "" on null usage / zero creation).
	if r.isSegmentEnabled(SegmentCacheHit) {
		if ch := renderCacheHit(data); ch != "" {
			segs = append(segs, ch)
		}
	}

	// Claude version
	if r.isSegmentEnabled(SegmentClaudeVersion) && data.ClaudeCodeVersion != "" {
		if withPrefix {
			segs = append(segs, fmt.Sprintf("🔅 cc v%s", data.ClaudeCodeVersion))
		} else {
			segs = append(segs, fmt.Sprintf("🔅 v%s", data.ClaudeCodeVersion))
		}
	}

	// MoAI version
	if r.isSegmentEnabled(SegmentMoaiVersion) && data.Version.Available && data.Version.Current != "" {
		var versionStr string
		versionStr = fmt.Sprintf("🗿 v%s", data.Version.Current)
		if data.Version.UpdateAvailable && data.Version.Latest != "" {
			versionStr += fmt.Sprintf(" -> 🗿 v%s", data.Version.Latest)
		}
		segs = append(segs, versionStr)
	}

	// Session time
	if r.isSegmentEnabled(SegmentSessionTime) && data.Metrics.Available {
		if st := renderSessionTime(data.Metrics.SessionDurationMS); st != "" {
			segs = append(segs, st)
		}
	}

	// Output style (last L1 segment per layout v3 amend: directory moved back to L3 head)
	if r.isSegmentEnabled(SegmentOutputStyle) && data.OutputStyle != "" {
		segs = append(segs, fmt.Sprintf("💬 %s", data.OutputStyle))
	}

	return r.joinSegments(segs)
}

// renderEffortThinking renders the effort/thinking indicator segment.
// Returns "🧠 LEVEL" + optional "·t" suffix when either field is present and meaningful.
// Returns "" when both are absent or effort level is empty (silent omit, REQ-CC2122-003).
func renderEffortThinking(data *StatusData) string {
	if data.Effort == nil {
		if data.Thinking != nil && data.Thinking.Enabled {
			return "·t"
		}
		return ""
	}
	if data.Effort.Level == "" {
		return ""
	}
	result := "🧠 " + data.Effort.Level
	if data.Thinking != nil && data.Thinking.Enabled {
		result += "·t"
	}
	return result
}

// cacheHitPercent는 cache-read 대 cache-creation 히트율을 계산한다.
// 히트율 = cache_read / (cache_read + cache_creation) * 100.
// cacheCreation이 0 이하이면(측정할 fresh cache write가 없음 — 0/0 both-zero 포함) ok=false를
// 반환해 호출자가 세그먼트를 생략하도록 한다(graceful degradation, REQ-TEF-006 — 값을
// 지어내지 않고 0으로 나누지 않는다). 곱셈은 int64로 수행해 매우 큰 토큰 수에서도 오버플로가
// 없다.
func cacheHitPercent(cacheRead, cacheCreation int) (int, bool) {
	if cacheCreation <= 0 {
		return 0, false
	}
	denom := int64(cacheRead) + int64(cacheCreation)
	if denom <= 0 {
		return 0, false
	}
	return int(int64(cacheRead) * 100 / denom), true
}

// renderCacheHit은 cache-hit-ratio 세그먼트를 렌더한다(SPEC-TOKEN-EFFICIENCY-001 P0-2).
// prompt-prefix churn의 조기 경고 신호(Claude Code 팀이 직접 알림을 거는 지표와 동일)로서
// cache_read / (cache_read + cache_creation)를 노출한다. cache usage가 없거나(null
// current_usage) cache_creation이 0이면 "" 을 반환해 세그먼트를 생략한다(REQ-TEF-006).
func renderCacheHit(data *StatusData) string {
	if data.CacheUsage == nil {
		return ""
	}
	pct, ok := cacheHitPercent(data.CacheUsage.CacheReadTokens, data.CacheUsage.CacheCreationTokens)
	if !ok {
		return ""
	}
	return fmt.Sprintf("♻️ %d%%", pct)
}

// renderBarsInline renders CW/5H/7D bars inline on a single line (default mode L2).
// width: number of blocks per bar
func (r *Renderer) renderBarsInline(data *StatusData, width int) string {
	var segs []string

	// CW bar with two-stage handoff_guide /clear suffix (layout v3 CH2 +
	// SPEC-HANDOFF-THRESHOLD-001 M4). handoffGuideStage classifies raw context
	// usage into none / soft / hard: soft keeps the M1 "(⚠️/clear)" marker, hard
	// escalates to the distinct stage-2 "(🛑/clear!)" marker at the auto-compact-
	// aware ceiling. The suffix is a pure function of usage — the handoff mode /
	// guide config never gates it (M1 no-regression invariant, REQ-THRESHOLD-006).
	if r.isSegmentEnabled(SegmentContext) && data.Memory.Available && data.Memory.TokenBudget > 0 {
		pct := usagePercent(data.Memory.TokensUsed, data.Memory.TokenBudget)
		bar := renderUsageBar("CW:", pct, width, r.noColor)
		switch handoffGuideStage(data) {
		case handoffStageHard:
			bar += " (🛑/clear!)"
		case handoffStageSoft:
			bar += " (⚠️/clear)"
		}
		segs = append(segs, bar)
	}

	// 5H bar - always shown, defaults to 0% when no data.
	// Prefer RateLimits (from Claude Code v2.1.80+ statusline JSON) over Usage (MoAI API call).
	if r.isSegmentEnabled(SegmentUsage5H) {
		pct5H := 0
		var reset5H string
		if data.RateLimits != nil && data.RateLimits.FiveHour != nil {
			pct5H = int(data.RateLimits.FiveHour.UsedPercentage)
			reset5H = formatResetTimeRelative(data.RateLimits.FiveHour.ResetsAt)
		} else if data.Usage != nil && data.Usage.Usage5H != nil {
			pct5H = int(data.Usage.Usage5H.Percentage)
			reset5H = formatResetTimeRelative(data.Usage.Usage5H.ResetsAt)
		}
		segs = append(segs, renderUsageBarWithReset("5H:", pct5H, width, r.noColor, reset5H))
	}

	// 7D bar - always shown, defaults to 0% when no data.
	// Prefer RateLimits (from Claude Code v2.1.80+ statusline JSON) over Usage (MoAI API call).
	if r.isSegmentEnabled(SegmentUsage7D) {
		pct7D := 0
		var reset7D string
		if data.RateLimits != nil && data.RateLimits.SevenDay != nil {
			pct7D = int(data.RateLimits.SevenDay.UsedPercentage)
			reset7D = formatResetTimeAbsolute(data.RateLimits.SevenDay.ResetsAt)
		} else if data.Usage != nil && data.Usage.Usage7D != nil {
			pct7D = int(data.Usage.Usage7D.Percentage)
			reset7D = formatResetTimeAbsolute(data.Usage.Usage7D.ResetsAt)
		}
		segs = append(segs, renderUsageBarWithReset("7D:", pct7D, width, r.noColor, reset7D))
	}

	return r.joinSegments(segs)
}

// renderDirGitLine renders the L3 line for layout v3.
// Format: 🔀 owner/name | 🅱️ branch ↑N +N │ 📫 +0 M6 ?0 │ [task] │ 💌 PR #1023 (⌥approved)
//
// Layout v3 changes (CH3 + CH5):
//   - directory moved to L1 end (CH5)
//   - branch + repo merged into single repo_branch segment (CH3)
//   - long_context + handoff_guide separate segments removed (CH1, CH2)
//   - handoff_guide integrated as CW bar (/clear) suffix in renderBarsInline (CH2)
//   - PR segment last position with new format "💌 PR #N (⌥state)" (CH7, CH8)
func (r *Renderer) renderDirGitLine(data *StatusData) string {
	var segs []string

	// Directory (layout v3 amend: L3 head — placed before repo_branch)
	if r.isSegmentEnabled(SegmentDirectory) && data.Directory != "" {
		segs = append(segs, fmt.Sprintf("📁 %s", data.Directory))
	}

	// Combined repo+branch segment (layout v3 CH3): replaces former
	// SegmentGitBranch + SegmentRepo pair.
	if r.isSegmentEnabled(SegmentGitBranch) {
		if rb := r.renderRepoBranchSegment(data); rb != "" {
			segs = append(segs, rb)
		}
	}

	// Git status with mailbox emoji: 📬(staged) > 📫(modified) > 📪(untracked) > 📭(clean)
	if r.isSegmentEnabled(SegmentGitStatus) {
		if emoji := mailboxEmoji(data); emoji != "" {
			if git := r.renderGitStatusDetail(data); git != "" {
				segs = append(segs, fmt.Sprintf("%s %s", emoji, git))
			}
		}
	}

	// Task segment (REQ-V3 Cycle 5 Phase 4, opt-in)
	// Active SPEC workflow info — position: immediately before PR (workflow context → review context order)
	if task := r.renderTaskSegment(data); task != "" {
		segs = append(segs, task)
	}

	// PR segment last position (layout v3 CH7 + CH8 format)
	if pr := r.renderPRSegment(data); pr != "" {
		segs = append(segs, pr)
	}

	return r.joinSegments(segs)
}

// renderTaskSegment renders the active SPEC workflow task segment.
// Format: "📋 [command SPEC-XXX-stage]" — wraps TaskData.Format() with a clipboard icon.
//
// Returns empty string when:
//   - SegmentTask is not explicitly enabled in segmentConfig (opt-in default off)
//   - data.Task is inactive (Active==false) OR Command is empty (TaskData.Format() returns "")
func (r *Renderer) renderTaskSegment(data *StatusData) string {
	if !r.isTaskEnabled() {
		return ""
	}
	if data == nil {
		return ""
	}
	formatted := data.Task.Format()
	if formatted == "" {
		return ""
	}
	return fmt.Sprintf("📋 %s", formatted)
}

// isTaskEnabled returns true when SegmentTask is enabled in segmentConfig.
// Default-on as of v2.20.0-rc1 — unset key resolves to enabled, matching
// isSegmentEnabled semantics. Graceful no-output handles inactive task
// (TaskData.Format() returns "" → segment hidden).
func (r *Renderer) isTaskEnabled() bool {
	if len(r.segmentConfig) == 0 {
		return true
	}
	enabled, exists := r.segmentConfig[SegmentTask]
	if !exists {
		return true
	}
	return enabled
}

// renderPRSegment renders the PR segment in the form "#<number> ⌥<state>".
// REQ-SLV-013 (render format) + REQ-SLV-014 (review-state color coding) +
// REQ-SLV-015 (absence handling).
//
// Returns empty string when:
//   - SegmentPR is not explicitly enabled (opt-in default off per REQ-SLV-012)
//   - data.PR is nil (REQ-SLV-015)
//   - data.PR.Number == 0 (REQ-SLV-015 — no #0 placeholder)
//
// When data.PR.ReviewState is empty, the segment renders "#<number>" without
// the ⌥<state> marker. Unknown review_state values render with the marker but
// no ANSI color (raw passthrough, REQ-SLV-014).
func (r *Renderer) renderPRSegment(data *StatusData) string {
	if !r.isPREnabled() {
		return ""
	}
	if data == nil || data.PR == nil || data.PR.Number == 0 {
		return ""
	}

	// Base segment text: "💌 PR #<number>" (layout v3 CH8)
	numberText := fmt.Sprintf("💌 PR #%d", data.PR.Number)

	// Review-state suffix: "(⌥<state>)" — parenthesised per layout v3 CH8
	// (omitted when ReviewState is empty)
	state := data.PR.ReviewState
	if state == "" {
		return numberText
	}
	stateText := fmt.Sprintf("(⌥%s)", state)

	// Apply color to the review-state suffix only (number stays uncolored)
	if !r.noColor {
		if color, ok := r.prReviewStateColor(state); ok {
			stateText = lipgloss.NewStyle().Foreground(color).Render(stateText)
		}
	}

	return fmt.Sprintf("%s %s", numberText, stateText)
}

// isPREnabled returns true when SegmentPR is enabled in segmentConfig.
// Default-on as of v2.20.0-rc1 (REQ-SLV-012 supersession) — unset key resolves
// to enabled, matching isSegmentEnabled semantics. Graceful no-output handles
// the no-PR case (data.PR == nil → segment hidden).
func (r *Renderer) isPREnabled() bool {
	if len(r.segmentConfig) == 0 {
		return true
	}
	enabled, exists := r.segmentConfig[SegmentPR]
	if !exists {
		return true
	}
	return enabled
}

// renderRepoBranchSegment renders the combined repo + branch segment in the
// form "🔀 owner/name | 🅱️ branch ↑N +N" — layout v3 CH3.
//
// Behavior:
//   - Workspace.Repo present + Branch present: "🔀 owner/name | 🅱️ branch ↑N +N"
//   - Workspace.Repo nil or incomplete:        "" (segment hidden — no git remote context)
//   - Branch empty:                            "" (empty — no git context)
//   - Ahead == 0:                              "↑N" portion omitted
//   - Behind > 0:                              " ↓N" appended after ahead
//   - Dirty (Modified + Staged + Untracked) == 0: " +N" portion omitted
//   - Worktree active:                          "[WT] " prefix prepended to branch
//
// @MX:NOTE: [AUTO] layout v3 CH3 — replaces standalone renderGitBranch + renderRepoSegment pair.
// @MX:NOTE: [AUTO] Hide entire segment when git is uninitialized or remote repo info is missing (per user request 2026-05-22).
func (r *Renderer) renderRepoBranchSegment(data *StatusData) string {
	if data == nil || !data.Git.Available || data.Git.Branch == "" {
		return ""
	}

	// Hide segment when repo info is missing (git uninitialized or remote not configured).
	if data.Workspace.Repo == nil {
		return ""
	}
	repo := data.Workspace.Repo
	if repo.Owner == "" || repo.Name == "" {
		return ""
	}

	branch := "🅱️ " + data.Git.Branch
	if r.isSegmentEnabled(SegmentWorktree) && data.Worktree != "" {
		branch = "[WT] " + branch
	}

	// Ahead/Behind suffix (0 values omitted)
	var aheadBehind string
	if data.Git.Ahead > 0 {
		aheadBehind += fmt.Sprintf(" ↑%d", data.Git.Ahead)
	}
	if data.Git.Behind > 0 {
		aheadBehind += fmt.Sprintf(" ↓%d", data.Git.Behind)
	}

	// Dirty count (omitted when 0)
	dirty := data.Git.Modified + data.Git.Staged + data.Git.Untracked
	var dirtySuffix string
	if dirty > 0 {
		dirtySuffix = fmt.Sprintf(" +%d", dirty)
	}

	inner := fmt.Sprintf("%s%s%s", branch, aheadBehind, dirtySuffix)
	return fmt.Sprintf("🔀 %s/%s | %s", repo.Owner, repo.Name, inner)
}

// handoffStage classifies accumulated context usage into the two-stage handoff
// gate (SPEC-HANDOFF-THRESHOLD-001 M4). none < soft < hard by raw usage.
type handoffStage int

const (
	handoffStageNone handoffStage = iota // below the soft threshold — no suffix
	handoffStageSoft                     // soft threshold reached — "(⚠️/clear)" hint
	handoffStageHard                     // hard (auto-compact-aware) ceiling reached — "(🛑/clear!)"
)

// handoffGuideStage classifies raw context usage into none / soft / hard.
// It supersedes the former bare-bool decision while preserving the M1 band
// logic verbatim for the soft threshold (SPEC-HANDOFF-CTXGUIDE-001):
//
//   - large window   (ContextWindowSize >= HandoffLargeWindowCutoff): soft at HandoffSoftLargePct%
//   - standard/medium (0 < ContextWindowSize < HandoffLargeWindowCutoff): soft at HandoffSoftStandardPct%
//   - unknown window  (ContextWindowSize <= 0): none (safety default)
//
// The hard stage escalates at hardCeilingPct — an auto-compact-aware ceiling
// (§ hardCeilingPct). hard is evaluated before soft so the stronger marker wins.
//
// Uses raw Memory.ContextWindowSize (Claude Code stdin context_window_size)
// instead of Memory.TokenBudget — TokenBudget is auto-compact-threshold-scaled
// (e.g., 1M × 85% = 850K) which would never match the raw class boundaries.
//
// @MX:NOTE: [AUTO] two-stage band: soft at band threshold, hard at min(cap, autoCompact+margin) — aligned with context-window-management.md HARD rule.
func handoffGuideStage(data *StatusData) handoffStage {
	if data == nil {
		return handoffStageNone
	}
	cwSize := data.Memory.ContextWindowSize
	if cwSize <= 0 {
		return handoffStageNone
	}
	rawPct := float64(data.Memory.TokensUsed) * 100.0 / float64(cwSize)
	soft := softThresholdPct(cwSize)
	hard := hardCeilingPct(cwSize)
	switch {
	case rawPct >= hard:
		return handoffStageHard
	case rawPct >= soft:
		return handoffStageSoft
	default:
		return handoffStageNone
	}
}

// softThresholdPct returns the raw-usage soft-stage threshold (%) for the given
// context window size, using the config band constants (no inline literals, §14).
// M1 band logic (≥ cutoff → large, else standard/medium) is preserved verbatim.
func softThresholdPct(cwSize int) float64 {
	if cwSize >= config.HandoffLargeWindowCutoff {
		return config.HandoffSoftLargePct
	}
	return config.HandoffSoftStandardPct
}

// hardCeilingPct returns the auto-compact-aware hard (stage-2) ceiling (%):
//
//	min(HandoffHardCeilingCapPct, getAutoCompactThreshold() + HandoffHardCeilingMarginPct)
//
// clamped up to the band's soft threshold when a degenerate auto-compact
// override would place the computed ceiling below soft (so stage-2 never
// inverts below stage-1; instead it collapses onto the soft threshold and only
// the hard marker shows for that band). getAutoCompactThreshold lives in the
// same package (memory.go) so no wiring is needed.
//
// Reachability note: because auto-compact fires near getAutoCompactThreshold()%
// of the raw window, the hard ceiling is frequently pre-empted and stage-2
// rarely fires in practice — an intentional, documented tradeoff of the
// auto-compact-aware formula (see context-window-management.md § Detection
// Heuristics).
func hardCeilingPct(cwSize int) float64 {
	ceil := getAutoCompactThreshold() + config.HandoffHardCeilingMarginPct
	if ceil > config.HandoffHardCeilingCapPct {
		ceil = config.HandoffHardCeilingCapPct
	}
	soft := softThresholdPct(cwSize)
	if float64(ceil) < soft {
		return soft
	}
	return float64(ceil)
}

// shouldShowHandoffGuide is the M1 backward-compatible wrapper: true when the
// stage is anything other than none. Kept so existing callers/tests compile and
// the M1 threshold behavior is byte-preserved (REQ-THRESHOLD-001).
func shouldShowHandoffGuide(data *StatusData) bool {
	return handoffGuideStage(data) != handoffStageNone
}

// prReviewStateColor maps a Claude Code v2.1.145 review_state value to a
// theme-aware color. Returns (color, true) for known states; (zero, false)
// for unknown / empty values to signal raw-passthrough rendering per
// REQ-SLV-014 + REQ-CC2122-004 precedent.
//
// Mapping (per plan.md D3):
//   - approved          → Success (green)
//   - pending           → Warning (yellow)
//   - changes_requested → Danger (red)
//   - draft             → Muted (gray)
func (r *Renderer) prReviewStateColor(state string) (lipgloss.Color, bool) {
	switch state {
	case "approved":
		return r.theme.Success(), true
	case "pending":
		return r.theme.Warning(), true
	case "changes_requested":
		return r.theme.Danger(), true
	case "draft":
		return r.theme.Muted(), true
	default:
		return lipgloss.Color(""), false
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Helper functions
// ─────────────────────────────────────────────────────────────────────────────

// renderUsageBar renders battery icon + label + gradient bar + percentage.
// Format: {BatteryIcon(pct)} {label} {BuildGradientBar(pct, width, noColor)} {pct}%
// Example: 🪫 CW: ████████████████████████████████████░░░░ 88%
// Layout v3 CH6: icon position moved before label for visual prominence.
func renderUsageBar(label string, pct int, width int, noColor bool) string {
	icon := BatteryIcon(pct)
	bar := BuildGradientBar(pct, width, noColor)
	return fmt.Sprintf("%s %s %s %d%%", icon, label, bar, pct)
}

// renderUsageBarWithReset renders a usage bar with optional reset time suffix.
// Format: {label} {icon} {bar} {pct}% ({resetStr})
func renderUsageBarWithReset(label string, pct int, width int, noColor bool, resetStr string) string {
	base := renderUsageBar(label, pct, width, noColor)
	if resetStr == "" {
		return base
	}
	return fmt.Sprintf("%s (%s)", base, resetStr)
}

// formatResetTimeRelative formats a reset time as "4h 30m", "1D, 4h 30m", or "30m".
// Returns "rolling" if the reset time is zero, in the past, or unparseable.
// Accepts either an ISO 8601 string (from UsageData) or Unix epoch int64 (from RateLimitWindow).
func formatResetTimeRelative(resetTime interface{}) string {
	t := parseResetTime(resetTime)
	if t.IsZero() {
		return "rolling"
	}
	remaining := time.Until(t)
	if remaining <= 0 {
		return "rolling"
	}
	hours := int(remaining.Hours())
	minutes := int(remaining.Minutes()) % 60

	if hours >= 24 {
		days := hours / 24
		hoursRemain := hours % 24
		parts := []string{fmt.Sprintf("%dD", days)}
		if hoursRemain > 0 {
			parts = append(parts, fmt.Sprintf("%dh", hoursRemain))
		}
		if minutes > 0 {
			parts = append(parts, fmt.Sprintf("%dm", minutes))
		}
		return strings.Join(parts, ", ")
	}

	if hours > 0 {
		if minutes > 0 {
			return fmt.Sprintf("%dh %dm", hours, minutes)
		}
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dm", minutes)
}

// formatResetTimeAbsolute formats a reset time as "Jan 21" or "rolling".
// Returns "rolling" (sliding window) if the reset time is zero or unparseable.
// Accepts either an ISO 8601 string (from UsageData) or Unix epoch int64 (from RateLimitWindow).
func formatResetTimeAbsolute(resetTime interface{}) string {
	t := parseResetTime(resetTime)
	if t.IsZero() {
		return "rolling"
	}
	// Convert to local time for display
	t = t.Local()
	return t.Format("Jan 2")
}

// parseResetTime converts a reset time value to time.Time.
// Accepts:
//   - int64: Unix epoch seconds (from Claude Code official statusline schema)
//   - string: ISO 8601 / RFC 3339 timestamp (from MoAI API UsageData)
//
// Returns zero time on failure.
func parseResetTime(resetTime interface{}) time.Time {
	switch v := resetTime.(type) {
	case int64:
		if v <= 0 {
			return time.Time{}
		}
		return time.Unix(v, 0)
	case string:
		if v == "" {
			return time.Time{}
		}
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			// Try without timezone
			t, err = time.Parse("2006-01-02T15:04:05", v)
			if err != nil {
				return time.Time{}
			}
		}
		return t
	default:
		return time.Time{}
	}
}

// renderGitBranch renders the git branch string with optional ahead/behind suffix.
//
// Format: "🅱️ <branch>[ ↑N][ ↓N] +<dirty>"
//   - Ahead and Behind are shown only when non-zero (zero values are omitted).
//   - The dirty count aggregates Modified + Staged + Untracked.
//   - When data.Git is unavailable or branch is empty, returns "".
//
// Status emoji prefixes (legacy 🔨/📦) have been removed: the L3/L5 mailbox emoji
// (📬/📫/📪/📭) plus the detailed `+S M_M ?U` counter already convey staged/modified
// state, so prefixing the branch produced redundant visual noise.
// Note: "[WT] " (worktree) prefix is added by renderDirGitLine when segment is enabled.
func renderGitBranch(data *StatusData) string {
	if !data.Git.Available || data.Git.Branch == "" {
		return ""
	}

	branch := data.Git.Branch

	// Ahead/Behind suffix (0 values omitted).
	var aheadBehind string
	if data.Git.Ahead > 0 {
		aheadBehind += fmt.Sprintf(" ↑%d", data.Git.Ahead)
	}
	if data.Git.Behind > 0 {
		aheadBehind += fmt.Sprintf(" ↓%d", data.Git.Behind)
	}

	// Dirty count (modified + staged + untracked); always rendered, includes +0.
	dirty := data.Git.Modified + data.Git.Staged + data.Git.Untracked
	dirtySuffix := fmt.Sprintf(" +%d", dirty)

	return fmt.Sprintf("🅱️ %s%s%s", branch, aheadBehind, dirtySuffix)
}

// renderSessionTime converts milliseconds to a session time string in "⏳ Xh Ym" format.
// REQ-V3-TIME-002: >= 60min: "⏳ Xh Ym", < 60min: "⏳ Xm", >= 24h: "⏳ Xd Yh"
// REQ-V3-TIME-004: returns empty string when ms is 0
func renderSessionTime(ms int) string {
	if ms <= 0 {
		return ""
	}

	totalMinutes := ms / 60000
	totalHours := totalMinutes / 60

	// >= 24 hours: "⏳ Xd Yh"
	if totalHours >= 24 {
		days := totalHours / 24
		hours := totalHours % 24
		return fmt.Sprintf("⏳ %dd %dh", days, hours)
	}

	// >= 1 hour: "⏳ Xh Ym"
	if totalHours >= 1 {
		minutes := totalMinutes % 60
		return fmt.Sprintf("⏳ %dh %dm", totalHours, minutes)
	}

	// < 1 hour: "⏳ Xm"
	return fmt.Sprintf("⏳ %dm", totalMinutes)
}

// mailboxEmoji returns a single disk emoji for the git status segment.
// Layout v3 amend: unified 💾 marker replaces the prior mailbox quartet
// (📬 staged / 📫 modified / 📪 untracked / 📭 clean) per user request —
// granular state is conveyed by the trailing "+S MM ?U" counter, so the
// leading emoji no longer needs to encode state independently.
func mailboxEmoji(data *StatusData) string {
	if !data.Git.Available {
		return ""
	}
	return "💾"
}

// renderGitStatusDetail renders detailed git status string (+N M?N).
// Always renders when git is available, including clean state (+0 M0 ?0).
func (r *Renderer) renderGitStatusDetail(data *StatusData) string {
	if !data.Git.Available {
		return ""
	}
	return fmt.Sprintf("+%d M%d ?%d", data.Git.Staged, data.Git.Modified, data.Git.Untracked)
}
