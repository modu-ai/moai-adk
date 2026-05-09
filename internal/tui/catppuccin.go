// Package tui — Catppuccin palette constants.
//
// This file is the single source of truth for Catppuccin Mocha and
// Catppuccin Latte terminal colour values used by the statusline package.
// Keeping the hex literals inside internal/tui/ satisfies REQ-CLI-TUI-013
// (no hex outside tui/) while giving internal/statusline/theme.go a thin
// re-export path (R-07 mitigation, M6-S4).
//
// @MX:ANCHOR: [AUTO] Catppuccin 팔레트 단일 진실 공급원; statusline 패키지가 import
// @MX:REASON: [AUTO] fan_in >= 2 (statusline.catppuccinMocha, statusline.catppuccinLatte); R-07 thin wrapper 진입점
package tui

// Catppuccin Mocha — dark terminal palette.
// Source: https://github.com/catppuccin/catppuccin (MIT) rev 2024-07
const (
	// CatppuccinMochaPrimary (Lavender) — primary foreground / text.
	CatppuccinMochaPrimary = "#CDD6F4"
	// CatppuccinMochaMuted (Overlay0) — muted / secondary text.
	CatppuccinMochaMuted = "#6C7086"
	// CatppuccinMochaSuccess (Green) — success / OK indicator.
	CatppuccinMochaSuccess = "#A6E3A1"
	// CatppuccinMochaWarning (Yellow) — warning / caution indicator.
	CatppuccinMochaWarning = "#F9E2AF"
	// CatppuccinMochaDanger (Pink) — error / danger indicator.
	CatppuccinMochaDanger = "#F38BA8"
	// CatppuccinMochaPeach — intermediate gradient stage (51–75%).
	CatppuccinMochaPeach = "#FAB387"
)

// Catppuccin Latte — light terminal palette.
// Source: https://github.com/catppuccin/catppuccin (MIT) rev 2024-07
const (
	// CatppuccinLattePrimary (Text) — primary foreground / text.
	CatppuccinLattePrimary = "#4C4F69"
	// CatppuccinLatteMuted (Subtext0) — muted / secondary text.
	CatppuccinLatteMuted = "#9CA0B0"
	// CatppuccinLatteSuccess (Green) — success / OK indicator.
	CatppuccinLatteSuccess = "#40A02B"
	// CatppuccinLatteWarning (Yellow) — warning / caution indicator.
	CatppuccinLatteWarning = "#DF8E1D"
	// CatppuccinLatteDanger (Red) — error / danger indicator.
	CatppuccinLatteDanger = "#D20F39"
	// CatppuccinLattePeach — intermediate gradient stage (51–75%).
	CatppuccinLattePeach = "#FE640B"
)
