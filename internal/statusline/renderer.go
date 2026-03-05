package statusline

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Renderer formats StatusData into a multiline statusline string.
// v3 레이아웃을 지원한다: compact(2줄), default(3줄), full(5줄).
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
		// v3 구분자: U+2502 박스 그리기 세로 선
		separator:     " │ ",
		noColor:       noColor,
		segmentConfig: segmentConfig,
		theme:         theme,
	}

	if noColor {
		r.mutedStyle = lipgloss.NewStyle()
		return r
	}

	// 테마 뮤트 색상으로 스타일 설정 (REQ-SLE-017)
	r.mutedStyle = lipgloss.NewStyle().Foreground(theme.Muted())

	return r
}

// Render formats the StatusData into a statusline string based on the mode.
//
// v3 모드 매핑:
//   - ModeCompact, ModeMinimal → 2줄 compact 레이아웃
//   - ModeDefault              → 3줄 default 레이아웃
//   - ModeFull, ModeVerbose    → 5줄 full 레이아웃
//   - unknown                  → 3줄 default 레이아웃 (기본)
//
// @MX:ANCHOR: 모든 모드 렌더링의 단일 진입점 - builder.go의 Build()에서 호출됨
// @MX:REASON: 공개 API 경계점; 모드 라우팅 로직 포함
func (r *Renderer) Render(data *StatusData, mode StatuslineMode) string {
	if data == nil {
		return "MoAI"
	}

	// deprecated 모드를 v3 이름으로 정규화한다
	normalizedMode := NormalizeMode(mode)

	var result string
	switch normalizedMode {
	case ModeCompact:
		result = r.renderCompactV3(data)
	case ModeFull:
		result = r.renderFullV3(data)
	default: // ModeDefault 또는 알 수 없는 모드
		result = r.renderDefaultV3(data)
	}

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

// joinSegments는 세그먼트 슬라이스를 필터링하여 구분자로 합친다.
// 모든 세그먼트가 비어있으면 빈 문자열을 반환한다.
func (r *Renderer) joinSegments(segments []string) string {
	filtered := filterEmpty(segments)
	if len(filtered) == 0 {
		return ""
	}
	return strings.Join(filtered, r.separator)
}

// ─────────────────────────────────────────────────────────────────────────────
// v3 레이아웃 렌더러
// ─────────────────────────────────────────────────────────────────────────────

// renderCompactV3 는 compact 모드 2줄 레이아웃을 렌더링한다.
//
// L1: 🤖 Model │ CW: 🪫 ██████████ 88%
// L2: 🔀 feat/auth ↑2↓1 │ 📊 +3 M2 ?1
//
// REQ-V3-TIME-006: compact 모드에서 세션 시간 생략
// REQ-V3-API-011: compact 모드에서 5H/7D 바 생략
func (r *Renderer) renderCompactV3(data *StatusData) string {
	var lines []string

	// L1: 모델 + CW 바
	l1 := r.renderCompactLine1(data)
	if l1 != "" {
		lines = append(lines, l1)
	}

	// L2: 브랜치(ahead/behind) + git 상태
	l2 := r.renderGitLine(data)
	if l2 != "" {
		lines = append(lines, l2)
	}

	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n")
}

// renderDefaultV3 는 default 모드 3줄 레이아웃을 렌더링한다.
//
// L1: 🤖 Model │ 🔅 v2.1.50 │ 🗿 v2.8.0 │ ⏳ 2h 34m │ 💬 MoAI
// L2: CW: 🪫 ██████████ 88% │ 5H: 🔋 ██████████ 45% │ 7D: 🪫 ██████████ 82%
// L3: 📁 moai-adk-go │ 🔀 feat/auth ↑2↓1 │ 📊 +3 M2 ?1
func (r *Renderer) renderDefaultV3(data *StatusData) string {
	var lines []string

	// L1: 모델, Claude 버전, MoAI 버전, 세션 시간, 출력 스타일
	l1 := r.renderInfoLine(data, false)
	if l1 != "" {
		lines = append(lines, l1)
	}

	// L2: CW/5H/7D 바 인라인 (10블록) - 항상 3개 바 모두 표시
	l2 := r.renderBarsInline(data, 10)
	if l2 != "" {
		lines = append(lines, l2)
	}

	// L3: 디렉토리, 브랜치, git 상태
	l3 := r.renderDirGitLine(data)
	if l3 != "" {
		lines = append(lines, l3)
	}

	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n")
}

// renderFullV3 는 full 모드 5줄 레이아웃을 렌더링한다.
//
// L1: 🤖 Model │ 🔅 Claude v2.1.50 │ 🗿 MoAI v2.8.0 │ ⏳ 2h 34m │ 💬 MoAI
// L2: CW: 🪫 ████████████████████████████████████░░░░ 88%
// L3: 5H: 🔋 ██████████████████░░░░░░░░░░░░░░░░░░░░░░ 45%
// L4: 7D: 🪫 ████████████████████████████████░░░░░░░░ 82%
// L5: 📁 moai-adk-go │ 🔀 feat/auth ↑2↓1 │ 📊 +3 M2 ?1
func (r *Renderer) renderFullV3(data *StatusData) string {
	var lines []string

	// L1: 모델, Claude 버전(접두사 포함), MoAI 버전(접두사 포함), 세션 시간, 출력 스타일
	l1 := r.renderInfoLine(data, true)
	if l1 != "" {
		lines = append(lines, l1)
	}

	// L2: CW 바 (40블록, 단독 줄)
	cwPct := r.contextPercent(data)
	if cwPct >= 0 {
		lines = append(lines, renderUsageBar("CW:", cwPct, 40, r.noColor))
	}

	// L3: 5H 바 (40블록, 단독 줄) - 데이터 없으면 0%
	pct5H := 0
	if data.Usage != nil && data.Usage.Usage5H != nil {
		pct5H = int(data.Usage.Usage5H.Percentage)
	}
	lines = append(lines, renderUsageBar("5H:", pct5H, 40, r.noColor))

	// L4: 7D 바 (40블록, 단독 줄) - 데이터 없으면 0%
	pct7D := 0
	if data.Usage != nil && data.Usage.Usage7D != nil {
		pct7D = int(data.Usage.Usage7D.Percentage)
	}
	lines = append(lines, renderUsageBar("7D:", pct7D, 40, r.noColor))

	// L5: 디렉토리, 브랜치, git 상태
	l5 := r.renderDirGitLine(data)
	if l5 != "" {
		lines = append(lines, l5)
	}

	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n")
}

// ─────────────────────────────────────────────────────────────────────────────
// 공통 줄 렌더러
// ─────────────────────────────────────────────────────────────────────────────

// renderCompactLine1 는 compact 모드 L1을 렌더링한다.
// Format: 🤖 Model │ CW: 🪫 ██████████ 88%
// REQ-V3-TIME-006: 세션 시간 생략
func (r *Renderer) renderCompactLine1(data *StatusData) string {
	var segs []string

	// 모델
	if r.isSegmentEnabled(SegmentModel) && data.Metrics.Available && data.Metrics.Model != "" {
		segs = append(segs, fmt.Sprintf("🤖 %s", data.Metrics.Model))
	}

	// CW 바 (10블록)
	if r.isSegmentEnabled(SegmentContext) && data.Memory.Available && data.Memory.TokenBudget > 0 {
		pct := usagePercent(data.Memory.TokensUsed, data.Memory.TokenBudget)
		segs = append(segs, renderUsageBar("CW:", pct, 10, r.noColor))
	}

	return r.joinSegments(segs)
}

// renderInfoLine 는 L1 정보 줄을 렌더링한다 (default/full 공용).
// withPrefix=true이면 full 모드 형식 ("Claude v...", "MoAI v...")
// withPrefix=false이면 default 모드 형식 ("v...")
func (r *Renderer) renderInfoLine(data *StatusData, withPrefix bool) string {
	var segs []string

	// 모델
	if r.isSegmentEnabled(SegmentModel) && data.Metrics.Available && data.Metrics.Model != "" {
		segs = append(segs, fmt.Sprintf("🤖 %s", data.Metrics.Model))
	}

	// Claude 버전
	if r.isSegmentEnabled(SegmentClaudeVersion) && data.ClaudeCodeVersion != "" {
		if withPrefix {
			segs = append(segs, fmt.Sprintf("🔅 Claude v%s", data.ClaudeCodeVersion))
		} else {
			segs = append(segs, fmt.Sprintf("🔅 v%s", data.ClaudeCodeVersion))
		}
	}

	// MoAI 버전
	if r.isSegmentEnabled(SegmentMoaiVersion) && data.Version.Available && data.Version.Current != "" {
		var versionStr string
		if withPrefix {
			versionStr = fmt.Sprintf("🗿 MoAI v%s", data.Version.Current)
		} else {
			versionStr = fmt.Sprintf("🗿 v%s", data.Version.Current)
		}
		if data.Version.UpdateAvailable && data.Version.Latest != "" {
			versionStr += fmt.Sprintf(" ⬆️ v%s", data.Version.Latest)
		}
		segs = append(segs, versionStr)
	}

	// 세션 시간
	if r.isSegmentEnabled(SegmentSessionTime) && data.Metrics.Available {
		if st := renderSessionTime(data.Metrics.SessionDurationMS); st != "" {
			segs = append(segs, st)
		}
	}

	// 출력 스타일 (L1에 통합)
	if r.isSegmentEnabled(SegmentOutputStyle) && data.OutputStyle != "" {
		segs = append(segs, fmt.Sprintf("💬 %s", data.OutputStyle))
	}

	return r.joinSegments(segs)
}

// renderBarsInline 는 CW/5H/7D 바를 한 줄에 인라인으로 렌더링한다 (default 모드 L2).
// width: 각 바의 블록 수
func (r *Renderer) renderBarsInline(data *StatusData, width int) string {
	var segs []string

	// CW 바
	if r.isSegmentEnabled(SegmentContext) && data.Memory.Available && data.Memory.TokenBudget > 0 {
		pct := usagePercent(data.Memory.TokensUsed, data.Memory.TokenBudget)
		segs = append(segs, renderUsageBar("CW:", pct, width, r.noColor))
	}

	// 5H 바 - 데이터 없으면 0%로 항상 표시
	if r.isSegmentEnabled(SegmentUsage5H) {
		pct5H := 0
		if data.Usage != nil && data.Usage.Usage5H != nil {
			pct5H = int(data.Usage.Usage5H.Percentage)
		}
		segs = append(segs, renderUsageBar("5H:", pct5H, width, r.noColor))
	}

	// 7D 바 - 데이터 없으면 0%로 항상 표시
	if r.isSegmentEnabled(SegmentUsage7D) {
		pct7D := 0
		if data.Usage != nil && data.Usage.Usage7D != nil {
			pct7D = int(data.Usage.Usage7D.Percentage)
		}
		segs = append(segs, renderUsageBar("7D:", pct7D, width, r.noColor))
	}

	return r.joinSegments(segs)
}

// renderGitLine 는 브랜치 + git 상태 줄을 렌더링한다 (compact L2).
// Format: 🔀 feat/auth ↑2↓1 │ 📊 +3 M2 ?1
func (r *Renderer) renderGitLine(data *StatusData) string {
	var segs []string

	// 브랜치 + ahead/behind
	if r.isSegmentEnabled(SegmentGitBranch) {
		if branch := renderGitBranch(data); branch != "" {
			segs = append(segs, branch)
		}
	}

	// git 상태
	if r.isSegmentEnabled(SegmentGitStatus) {
		if git := r.renderGitStatus(data); git != "" {
			segs = append(segs, fmt.Sprintf("📊 %s", git))
		}
	}

	return r.joinSegments(segs)
}

// renderDirGitLine 는 디렉토리 + 브랜치 + git 상태 줄을 렌더링한다 (default L3, full L5).
// Format: 📁 moai-adk-go │ 🔀 feat/auth ↑2↓1 │ 📊 +3 M2 ?1
func (r *Renderer) renderDirGitLine(data *StatusData) string {
	var segs []string

	// 디렉토리
	if r.isSegmentEnabled(SegmentDirectory) && data.Directory != "" {
		segs = append(segs, fmt.Sprintf("📁 %s", data.Directory))
	}

	// 브랜치 + ahead/behind
	if r.isSegmentEnabled(SegmentGitBranch) {
		if branch := renderGitBranch(data); branch != "" {
			segs = append(segs, branch)
		}
	}

	// git 상태
	if r.isSegmentEnabled(SegmentGitStatus) {
		if git := r.renderGitStatus(data); git != "" {
			segs = append(segs, fmt.Sprintf("📊 %s", git))
		}
	}

	return r.joinSegments(segs)
}

// ─────────────────────────────────────────────────────────────────────────────
// 헬퍼 함수
// ─────────────────────────────────────────────────────────────────────────────

// renderUsageBar 는 레이블 + 배터리 아이콘 + 그라디언트 바 + 퍼센트를 렌더링한다.
// Format: {label} {BatteryIcon(pct)}  {BuildGradientBar(pct, width, noColor)} {pct}%
// Example: CW: 🪫  ████████████████████████████████████░░░░ 88%
func renderUsageBar(label string, pct int, width int, noColor bool) string {
	icon := BatteryIcon(pct)
	bar := BuildGradientBar(pct, width, noColor)
	return fmt.Sprintf("%s %s  %s %d%%", label, icon, bar, pct)
}

// contextPercent 는 컨텍스트 창 사용률(0~100)을 반환한다.
// 사용 불가이거나 전체가 0이면 -1을 반환한다.
func (r *Renderer) contextPercent(data *StatusData) int {
	if !data.Memory.Available || data.Memory.TokenBudget <= 0 {
		return -1
	}
	return usagePercent(data.Memory.TokensUsed, data.Memory.TokenBudget)
}

// renderGitBranch 는 브랜치 이름과 ahead/behind 정보를 포함한 git 브랜치 문자열을 렌더링한다.
// REQ-V3-GIT-001: Ahead > 0 이면 "🔀 branch ↑N"
// REQ-V3-GIT-002: Behind > 0 이면 "🔀 branch ↓N"
// REQ-V3-GIT-003: 둘 다 있으면 "🔀 branch ↑N↓M"
// REQ-V3-GIT-004: 둘 다 없으면 "🔀 branch"
func renderGitBranch(data *StatusData) string {
	if !data.Git.Available || data.Git.Branch == "" {
		return ""
	}

	branch := data.Git.Branch
	var suffix string

	if data.Git.Ahead > 0 && data.Git.Behind > 0 {
		suffix = fmt.Sprintf(" ↑%d↓%d", data.Git.Ahead, data.Git.Behind)
	} else if data.Git.Ahead > 0 {
		suffix = fmt.Sprintf(" ↑%d", data.Git.Ahead)
	} else if data.Git.Behind > 0 {
		suffix = fmt.Sprintf(" ↓%d", data.Git.Behind)
	}

	return fmt.Sprintf("🔀 %s%s", branch, suffix)
}

// renderSessionTime 는 밀리초를 "⏳ Xh Ym" 형식의 세션 시간 문자열로 변환한다.
// REQ-V3-TIME-002: 60분 이상은 "⏳ Xh Ym", 60분 미만은 "⏳ Xm", 24시간 이상은 "⏳ Xd Yh"
// REQ-V3-TIME-004: ms가 0이면 빈 문자열 반환
func renderSessionTime(ms int) string {
	if ms <= 0 {
		return ""
	}

	totalMinutes := ms / 60000
	totalHours := totalMinutes / 60

	// 24시간 이상: "⏳ Xd Yh"
	if totalHours >= 24 {
		days := totalHours / 24
		hours := totalHours % 24
		return fmt.Sprintf("⏳ %dd %dh", days, hours)
	}

	// 1시간 이상: "⏳ Xh Ym"
	if totalHours >= 1 {
		minutes := totalMinutes % 60
		return fmt.Sprintf("⏳ %dh %dm", totalHours, minutes)
	}

	// 1시간 미만: "⏳ Xm"
	return fmt.Sprintf("⏳ %dm", totalMinutes)
}

// renderGitStatus renders git status in Python format.
// Format: +{staged} M{modified} ?{untracked}
// Example: "+0 M1066 ?2" (0 staged, 1066 modified, 2 untracked)
func (r *Renderer) renderGitStatus(data *StatusData) string {
	if !data.Git.Available {
		return ""
	}

	// Only show git status if there are changes
	if data.Git.Staged == 0 && data.Git.Modified == 0 && data.Git.Untracked == 0 {
		return ""
	}

	var parts []string

	// Staged files
	parts = append(parts, fmt.Sprintf("+%d", data.Git.Staged))

	// Modified files (uses M instead of ~)
	parts = append(parts, fmt.Sprintf("M%d", data.Git.Modified))

	// Untracked files
	parts = append(parts, fmt.Sprintf("?%d", data.Git.Untracked))

	return strings.Join(parts, " ")
}
