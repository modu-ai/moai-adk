//! The main Claude Code statusline: up to three lines joined by `\n`, 10-cell
//! gradient bars, and two Catppuccin themes. Without color the text is identical
//! and carries no escape sequence.

use super::color::{hex_rgb, Profile};
use super::segments::{self, HandoffStage};
use super::state::{CacheTimer, DisplayConfig, Status};
use std::collections::BTreeMap;

const BAR_WIDTH: i64 = 10;
const SEPARATOR: &str = " | ";
const SGR_NORMAL_INTENSITY: &str = "\x1b[22m";
const SGR_RESET: &str = "\x1b[0m";

/// A theme's colors (internal/tui/catppuccin.go). `stops` run green → yellow → peach → red.
/// The per-segment colors are the #2559 segment palette (user decision 2026-11-12): each
/// segment kind takes its own Catppuccin accent, so neighbours never share a color.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub struct Theme {
    pub text: &'static str,
    pub muted: &'static str,
    pub stops: [&'static str; 4],
    /// `🤖` model (mauve).
    pub model: &'static str,
    /// `🧠` effort and thinking (lavender).
    pub effort: &'static str,
    /// `♻️` cache hit (teal).
    pub cache: &'static str,
    /// `🔅` Claude Code and `🗿` MoAI versions (sky).
    pub version: &'static str,
    /// `📡` repository (blue).
    pub repo: &'static str,
    /// `⬇️` open Issues (peach).
    pub issues: &'static str,
    /// `🔀` open pull requests (pink).
    pub prs: &'static str,
    /// `🅱️` branch with its behind and dirty counts (green).
    pub branch: &'static str,
    /// `(⚠️ /clear)` soft handoff warning (yellow).
    pub warn: &'static str,
    /// `(🛑/clear!)` hard handoff warning (red).
    pub alert: &'static str,
}

pub const MOCHA: Theme = Theme {
    text: "#CDD6F4",
    muted: "#6C7086",
    stops: ["#A6E3A1", "#F9E2AF", "#FAB387", "#F38BA8"],
    model: "#CBA6F7",
    effort: "#B4BEFE",
    cache: "#94E2D5",
    version: "#89DCEB",
    repo: "#89B4FA",
    issues: "#FAB387",
    prs: "#F5C2E7",
    branch: "#A6E3A1",
    warn: "#F9E2AF",
    alert: "#F38BA8",
};

pub const LATTE: Theme = Theme {
    text: "#4C4F69",
    muted: "#9CA0B0",
    stops: ["#40A02B", "#DF8E1D", "#FE640B", "#D20F39"],
    model: "#8839EF",
    effort: "#7287FD",
    cache: "#179299",
    version: "#04A5E5",
    repo: "#1E66F5",
    issues: "#FE640B",
    prs: "#EA76CB",
    branch: "#40A02B",
    warn: "#DF8E1D",
    alert: "#D20F39",
};

/// `NewTheme`: `catppuccin-latte`, anything else is Mocha.
pub fn theme(name: &str) -> Theme {
    if name == "catppuccin-latte" {
        LATTE
    } else {
        MOCHA
    }
}

pub struct Renderer {
    /// `None` renders without color.
    profile: Option<Profile>,
    theme: Theme,
    segments: BTreeMap<String, bool>,
}

impl Renderer {
    pub fn new(config: &DisplayConfig, profile: Option<Profile>) -> Renderer {
        Renderer {
            profile,
            theme: theme(&config.theme),
            segments: config.segments.clone(),
        }
    }

    /// `isSegmentEnabled`: an empty map or a missing key is enabled.
    fn enabled(&self, key: &str) -> bool {
        self.segments.get(key).copied().unwrap_or(true)
    }

    /// `Render`: the layout, or `MoAI` when every line is empty, then the base color.
    /// #2756 (user instruction 2026-11-13): line 1 is the running environment, line 2 the
    /// repository, line 3 the lane name and the usage bars.
    pub fn render(&self, status: &Status) -> String {
        let lines: Vec<String> = [
            self.info_line(status),
            self.git_line(status),
            self.bars_line(status),
        ]
        .into_iter()
        .filter(|l| !l.is_empty())
        .collect();
        let text = if lines.is_empty() {
            "MoAI".to_string()
        } else {
            lines.join("\n")
        };
        match self.profile {
            None => text,
            Some(p) => apply_base_color(&text, &format!("\x1b[{}m", p.foreground(self.theme.text))),
        }
    }

    fn paint(&self, hex: &str, text: &str) -> String {
        match self.profile {
            None => text.to_string(),
            Some(p) => p.paint(hex, text),
        }
    }

    /// #2756 AC-7: an OSC 8 hyperlink around `text`. Like color it is a terminal escape, so
    /// the plain output (no color profile) and an empty URL keep the text alone.
    fn link(&self, url: &str, text: &str) -> String {
        if self.profile.is_none() || url.is_empty() {
            return text.to_string();
        }
        format!("\x1b]8;;{url}\x1b\\{text}\x1b]8;;\x1b\\")
    }

    fn info_line(&self, s: &Status) -> String {
        let mut segs = Vec::new();
        let metrics = !s.model.is_empty();
        if self.enabled("model") && metrics {
            segs.push(self.paint(self.theme.model, &format!("🤖 {}", s.model)));
        }
        if self.enabled("effort_thinking") {
            let thinking = s.thinking == Some(true);
            match &s.effort {
                None if thinking => segs.push(self.paint(self.theme.effort, "·t")),
                Some(level) if !level.is_empty() => segs.push(self.paint(
                    self.theme.effort,
                    &format!("🧠 {level}{}", if thinking { "·t" } else { "" }),
                )),
                _ => {}
            }
        }
        if self.enabled("cache_hit") {
            let hit = s
                .cache
                .and_then(|(read, creation)| segments::cache_hit_percent(read, creation))
                .map(|pct| format!(" {pct}%"));
            // #2756 AC-8 (user choice 2026-11-13): how long the prompt cache stays warm, next
            // to the hit ratio; `🧊` once it is cold.
            let timer = s.cache_timer.map(|t| match t {
                CacheTimer::Warm(minutes) => format!(" ⏱️ {minutes}m"),
                CacheTimer::Cold => " 🧊".to_string(),
            });
            if hit.is_some() || timer.is_some() {
                let text = format!("♻️{}{}", hit.unwrap_or_default(), timer.unwrap_or_default());
                segs.push(self.paint(self.theme.cache, &text));
            }
        }
        if self.enabled("claude_version") && !s.claude_version.is_empty() {
            segs.push(self.paint(self.theme.version, &format!("🔅 v{}", s.claude_version)));
        }
        if self.enabled("moai_version") {
            if let Some(v) = s.version.as_ref().filter(|v| !v.current.is_empty()) {
                let mut seg = format!("🗿 v{}", v.current);
                if let Some(latest) = &v.latest {
                    seg.push_str(&format!(" -> 🗿 v{latest}"));
                }
                segs.push(self.paint(self.theme.version, &seg));
            }
        }
        // #2756 AC-1 (user instruction 2026-11-13): the session time is not shown, and
        // #2559 AC-7 (user instruction 2026-11-12): nor is the output style. The
        // `session_time` and `output_style` option names are still accepted in configs,
        // they just show nothing.
        segs.join(SEPARATOR)
    }

    fn bars_line(&self, s: &Status) -> String {
        let mut segs = Vec::new();
        // #2928 S-1 (user instruction 2026-11-14, replacing #2809's name-last order): the tag
        // comes first and is only `leader` or `agent` — never the stdin `session_name` text — and
        // is shown whenever the `session` segment is on, even without stdin (decision Q2 = (가)).
        if self.enabled("session") {
            let tag = if s.leader { "leader" } else { "agent" };
            segs.push(format!("🏷️ {}", crate::ui::visible(tag)));
        }
        if self.enabled("context") {
            if s.memory.available && s.memory.token_budget > 0 {
                let pct = segments::usage_percent(s.memory.tokens_used, s.memory.token_budget);
                let mut bar = self.usage_bar("CTX:", pct);
                if !s.goal_armed {
                    match segments::handoff_stage(&s.memory, s.auto_compact) {
                        HandoffStage::Hard => {
                            bar.push(' ');
                            bar.push_str(&self.paint(self.theme.alert, "(🛑/clear!)"));
                        }
                        // #2559: space before /clear
                        HandoffStage::Soft => {
                            bar.push(' ');
                            bar.push_str(&self.paint(self.theme.warn, "(⚠️ /clear)"));
                        }
                        HandoffStage::None => {}
                    }
                }
                segs.push(bar);
            } else {
                segs.push(self.unknown_bar("CTX:"));
            }
        }
        // #2756 leader decision 6862179411 item 1: the five-hour bar is not shown (the user
        // did not choose it, 6858992949); the `usage_5h` option is still accepted.
        if self.enabled("usage_7d") {
            if let Some(w) = s.seven_day {
                segs.push(self.usage_bar("Week:", w.used_percentage.unwrap_or(0.0) as i64));
            }
        }
        // The renderer never shortens a segment and does not read `COLUMNS`, which Claude
        // Code sets; whether an overlong line is cut or wraps is left open by the host docs
        // ("may get truncated or wrap awkwardly").
        segs.join(SEPARATOR)
    }

    /// `renderUsageBar`: battery icon, label, gradient bar, percentage.
    fn usage_bar(&self, label: &str, pct: i64) -> String {
        let icon = if pct > 70 { "🪫" } else { "🔋" };
        format!("{icon} {label} {} {pct}%", self.gradient_bar(pct))
    }

    /// `renderUnknownBar`: an empty muted placeholder and `n/a`, never a percentage.
    fn unknown_bar(&self, label: &str) -> String {
        let blocks = "░".repeat(BAR_WIDTH as usize);
        format!("{label} {} n/a", self.paint(self.theme.muted, &blocks))
    }

    /// `buildGradientBar`: each filled cell gets its own color across the theme stops.
    fn gradient_bar(&self, pct: i64) -> String {
        let filled = (pct * BAR_WIDTH / 100).clamp(0, BAR_WIDTH);
        let empty = "░".repeat((BAR_WIDTH - filled) as usize);
        if self.profile.is_none() || filled == 0 {
            return "█".repeat(filled as usize) + &empty;
        }
        let mut out = String::new();
        for i in 0..filled {
            let at = if filled > 1 {
                i as f64 / (filled - 1) as f64
            } else {
                0.0
            };
            let (r, g, b) = interpolate(&self.theme.stops, at);
            out.push_str(&self.paint(&format!("#{r:02X}{g:02X}{b:02X}"), "█"));
        }
        out + &empty
    }

    fn git_line(&self, s: &Status) -> String {
        let mut segs = Vec::new();
        if self.enabled("git_branch") {
            if let Some(rb) = self.repo_branch(s) {
                segs.push(rb);
            }
        }
        // #2756 AC-9 (user choice 2026-11-13): which worktree the session is in; absent in
        // the main checkout.
        if self.enabled("worktree") && !s.worktree.is_empty() {
            segs.push(self.paint(self.theme.branch, &format!("🌳 {}", s.worktree)));
        }
        // #2756 AC-7 (user choice 2026-11-13): this branch's pull request, linked to its URL.
        if self.enabled("pr") {
            if let Some(pr) = &s.pr {
                let mut text = format!("🔗 #{}", pr.number);
                if !pr.review_state.is_empty() {
                    text.push(' ');
                    text.push_str(&pr.review_state);
                }
                segs.push(self.link(&pr.url, &self.paint(self.theme.prs, &text)));
            }
        }
        // #2756 leader decision 6862179411 item 2: no `💾` file-change segment (the user's
        // line 2 ends at the branch, whose `+N` carries the change count); the `git_status`
        // option is still accepted.
        segs.join(SEPARATOR)
    }

    /// Repository, optional forge counts, and branch retain the `git_branch` switch.
    fn repo_branch(&self, s: &Status) -> Option<String> {
        let git = s.git.as_ref().filter(|g| !g.branch.is_empty())?;
        let mut branch = format!("🅱️ {}", git.branch);
        // Ahead is deliberately hidden; behind and dirty counts remain visible.
        if git.behind > 0 {
            branch.push_str(&format!(" ↓{}", git.behind));
        }
        let dirty = git.modified + git.staged + git.untracked;
        if dirty > 0 {
            branch.push_str(&format!(" +{dirty}"));
        }
        let branch = self.paint(self.theme.branch, &branch);
        let Some((owner, name)) = s
            .repo
            .as_ref()
            .filter(|(o, n)| !o.is_empty() && !n.is_empty())
        else {
            return Some(branch);
        };
        let forge = if self.enabled("github") && s.github.available && !s.github.suppressed {
            // #2559 (user instructions 2026-11-12): icons instead of letter labels; the Issue
            // icon is `⬇️` (U+2B07 U+FE0F) since 23:1xZ, the PR icon stays `🔀`.
            format!(
                " | {} / {}",
                self.paint(self.theme.issues, &format!("⬇️ {}", s.github.open_issues)),
                self.paint(self.theme.prs, &format!("🔀 {}", s.github.open_prs))
            )
        } else {
            String::new()
        };
        let repo = self.paint(self.theme.repo, &format!("📡 {owner}/{name}"));
        Some(format!("{repo}{forge} | {branch}"))
    }
}

/// `interpolateGradientColor`: evenly spaced stops, round half away from zero.
fn interpolate(stops: &[&str; 4], at: f64) -> (u8, u8, u8) {
    let segments = (stops.len() - 1) as f64;
    let pos = at.clamp(0.0, 1.0) * segments;
    let i = (pos as usize).min(stops.len() - 2);
    let f = pos - i as f64;
    let (r0, g0, b0) = hex_rgb(stops[i]);
    let (r1, g1, b1) = hex_rgb(stops[i + 1]);
    let lerp = |a: u8, b: u8| (f64::from(a) + (f64::from(b) - f64::from(a)) * f).round() as u8;
    (lerp(r0, r1), lerp(g0, g1), lerp(b0, b1))
}

/// `applyBaseColor`: every line opens with normal intensity and the base color, the base
/// color is re-applied after each inner reset, and the line closes with a reset.
fn apply_base_color(text: &str, base: &str) -> String {
    text.split('\n')
        .map(|line| {
            let line = line.replace(SGR_RESET, &format!("{SGR_RESET}{base}"));
            format!("{SGR_NORMAL_INTENSITY}{base}{line}{SGR_RESET}")
        })
        .collect::<Vec<_>>()
        .join("\n")
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::statusline::input::{self, RateWindow};
    use crate::statusline::state::{self, Context, GitHubCounts, Version};

    fn plain(status: &Status, disabled: &[&str]) -> String {
        let config = DisplayConfig {
            segments: disabled.iter().map(|key| ((*key).into(), false)).collect(),
            ..Default::default()
        };
        Renderer::new(&config, None).render(status)
    }

    fn full_status() -> Status {
        Status {
            model: "Opus".into(),
            effort: Some("high".into()),
            thinking: Some(true),
            cache: Some((300, 100)),
            claude_version: "2.1.293".into(),
            version: Some(Version {
                current: "2.0".into(),
                latest: Some("3.0".into()),
            }),
            session_ms: 5_400_000,
            output_style: "MoAI".into(),
            session_name: "release".into(),
            memory: segments::Memory {
                tokens_used: 50,
                token_budget: 100,
                context_window_size: 200_000,
                available: true,
            },
            auto_compact: 85,
            five_hour: Some(RateWindow {
                used_percentage: Some(25.0),
                resets_at: Some(1),
            }),
            seven_day: Some(RateWindow {
                used_percentage: Some(75.0),
                resets_at: Some(1),
            }),
            repo: Some(("owner".into(), "repo".into())),
            github: GitHubCounts {
                available: true,
                open_issues: 7,
                open_prs: 3,
                ..Default::default()
            },
            git: Some(segments::parse_porcelain(
                "## main...origin/main [ahead 5, behind 2]\n M a\nA  b\n?? c\n",
            )),
            ..Default::default()
        }
    }

    /// #2756 AC-1–AC-2 (user instructions 2026-11-13) and #2928 S-1 (2026-11-14): environment,
    /// repository, then the tag and the usage bars, each line joined by ` | `.
    #[test]
    fn full_main_layout_has_three_ordered_lines_without_reset_or_ahead() {
        assert_eq!(
            plain(&full_status(), &[]),
            "🤖 Opus | 🧠 high·t | ♻️ 75% | 🔅 v2.1.293 | 🗿 v2.0 -> 🗿 v3.0\n\
             📡 owner/repo | ⬇️ 7 / 🔀 3 | 🅱️ main ↓2 +3\n\
             🏷️ agent | 🔋 CTX: █████░░░░░ 50% | 🪫 Week: ███████░░░ 75%"
        );
    }

    /// #2559 AC-7 (user instruction 2026-11-12T16:5xZ): the output style is not shown,
    /// even when the input names one and the `output_style` option is left enabled.
    #[test]
    fn the_output_style_is_never_shown() {
        let status = full_status();
        assert_eq!(status.output_style, "MoAI", "the input names a style");
        let out = plain(&status, &[]);
        assert!(!out.contains('💬'), "{out}");
        assert!(!out.contains("MoAI"), "{out}");
        assert!(out.lines().next().unwrap().ends_with("🗿 v3.0"), "{out}");
    }

    /// #2756 AC-1 (user instruction 2026-11-13): the session time is not shown, even with a
    /// session length in the input and the `session_time` option left enabled.
    #[test]
    fn the_session_time_is_never_shown() {
        let status = full_status();
        assert_eq!(
            status.session_ms, 5_400_000,
            "the input carries a session length"
        );
        let out = plain(&status, &[]);
        assert!(!out.contains('⏳'), "{out}");
        let config = DisplayConfig {
            segments: [("session_time".to_string(), true)].into_iter().collect(),
            ..Default::default()
        };
        let enabled = Renderer::new(&config, None).render(&status);
        assert!(!enabled.contains('⏳'), "{enabled}");
    }

    /// #2756 leader decision 6862179411 items 1–2: no five-hour bar (the user did not choose
    /// it) and no `💾` file-change segment, even with both in the input and their
    /// `usage_5h` / `git_status` options left enabled. The branch keeps its `+N` changes.
    #[test]
    fn the_five_hour_bar_and_the_file_change_segment_are_never_shown() {
        let status = full_status();
        assert!(
            status.five_hour.is_some() && status.git.is_some(),
            "the input has both"
        );
        let config = DisplayConfig {
            segments: [
                ("usage_5h".to_string(), true),
                ("git_status".to_string(), true),
            ]
            .into_iter()
            .collect(),
            ..Default::default()
        };
        let out = Renderer::new(&config, None).render(&status);
        assert!(!out.contains("5H"), "{out}");
        assert!(!out.contains('💾'), "{out}");
        assert!(out.contains("🅱️ main ↓2 +3"), "{out}");
    }

    #[test]
    fn rate_windows_are_optional_but_present_windows_keep_zero_percent() {
        let mut status = Status::default();
        assert_eq!(plain(&status, &["session"]), "CTX: ░░░░░░░░░░ n/a");
        status.seven_day = Some(RateWindow::default());
        assert_eq!(
            plain(&status, &["session"]),
            "CTX: ░░░░░░░░░░ n/a | 🔋 Week: ░░░░░░░░░░ 0%"
        );
        assert_eq!(
            plain(&status, &["session", "context"]),
            "🔋 Week: ░░░░░░░░░░ 0%"
        );
        assert_eq!(
            plain(&status, &["session", "usage_7d"]),
            "CTX: ░░░░░░░░░░ n/a"
        );
    }

    /// #2559 AC-1: the bar's thresholds pinned at N/N+1 — the battery icon turns at 70/71 and
    /// a cell fills at each tenth (9/10). Plain output, so only the threshold decides.
    #[test]
    fn bar_thresholds_hold_at_their_boundaries() {
        let at = |pct: f64| {
            let status = Status {
                seven_day: Some(RateWindow {
                    used_percentage: Some(pct),
                    resets_at: None,
                }),
                ..Status::default()
            };
            plain(&status, &["context", "session"])
        };
        assert_eq!(at(70.0), "🔋 Week: ███████░░░ 70%");
        assert_eq!(at(71.0), "🪫 Week: ███████░░░ 71%");
        assert_eq!(at(9.0), "🔋 Week: ░░░░░░░░░░ 9%");
        assert_eq!(at(10.0), "🔋 Week: █░░░░░░░░░ 10%");
    }

    /// The renderer never shortens for width and does not read `COLUMNS`; since #2928 the
    /// stdin name is not printed at all, so a 200-character name leaves line 3 unchanged.
    #[test]
    fn a_long_name_changes_nothing_on_line_three() {
        let mut status = full_status();
        status.session_name = "x".repeat(200);
        assert_eq!(
            plain(&status, &[]).lines().nth(2),
            Some("🏷️ agent | 🔋 CTX: █████░░░░░ 50% | 🪫 Week: ███████░░░ 75%")
        );
    }

    /// #2928 S-1 (user instruction 2026-11-14): line 3 is `🏷️ <tag> | 🔋 CTX … | 🪫 Week …`.
    /// The tag is `leader` or `agent` and comes first; it is shown whenever the `session`
    /// segment is on, even with no stdin (decision Q2 = (가)); with no bars it stands alone, and
    /// with `session` off the line is the bars alone. The bar icons and marks are unchanged.
    #[test]
    fn the_tag_is_the_first_segment_before_the_bars() {
        let mut status = Status::default();
        assert_eq!(plain(&status, &[]), "🏷️ agent | CTX: ░░░░░░░░░░ n/a");
        status.seven_day = Some(RateWindow::default());
        assert_eq!(
            plain(&status, &[]),
            "🏷️ agent | CTX: ░░░░░░░░░░ n/a | 🔋 Week: ░░░░░░░░░░ 0%"
        );
        status.leader = true;
        assert_eq!(
            plain(&status, &[]),
            "🏷️ leader | CTX: ░░░░░░░░░░ n/a | 🔋 Week: ░░░░░░░░░░ 0%"
        );
        assert_eq!(plain(&status, &["usage_7d", "context"]), "🏷️ leader");
        assert_eq!(
            plain(&status, &["session"]),
            "CTX: ░░░░░░░░░░ n/a | 🔋 Week: ░░░░░░░░░░ 0%"
        );
        let mut full = full_status();
        assert_eq!(
            plain(&full, &[]).lines().nth(2),
            Some("🏷️ agent | 🔋 CTX: █████░░░░░ 50% | 🪫 Week: ███████░░░ 75%")
        );
        full.leader = true;
        assert_eq!(
            plain(&full, &[]).lines().nth(2),
            Some("🏷️ leader | 🔋 CTX: █████░░░░░ 50% | 🪫 Week: ███████░░░ 75%")
        );
        let colored = strip_sgr(&colored("mocha", Profile::TrueColor, &full));
        assert!(
            colored.lines().nth(2).unwrap().starts_with("🏷️ leader | "),
            "{colored:?}"
        );
    }

    /// #2928 S-3 (#2818 B1): the stdin `session_name` — a lane name, a terminal title, escape
    /// or C1 bytes, even the word `leader` itself — never reaches the output; only the tag
    /// does, and without a registry reading it is `agent`.
    #[test]
    fn session_name_text_never_reaches_the_tag() {
        for name in [
            "\u{1b}]0;evil\u{7}",
            "x\u{9b}31m",
            "#2388·#2403·#2385 Issue 닫기 대조",
            "agent-27",
            "/moai",
            "leader",
        ] {
            let mut status = full_status();
            status.session_name = name.into();
            let out = plain(&status, &[]);
            assert_eq!(
                out.lines().nth(2),
                Some("🏷️ agent | 🔋 CTX: █████░░░░░ 50% | 🪫 Week: ███████░░░ 75%"),
                "{name:?}"
            );
            let colored = colored("mocha", Profile::TrueColor, &status);
            for text in [&out, &colored] {
                assert!(!text.contains("\u{1b}]0;"), "{name:?}: {text:?}");
                assert!(!text.contains('\u{9b}'), "{name:?}: {text:?}");
                assert!(!text.contains("Issue"), "{name:?}: {text:?}");
                assert!(!text.contains("agent-27"), "{name:?}: {text:?}");
                assert!(!text.contains("/moai"), "{name:?}: {text:?}");
            }
        }
    }

    /// #2928 S-4: lines 1 and 2 are the same bytes whatever the tag and the stdin name, plain
    /// and colored.
    #[test]
    fn lines_one_and_two_are_identical_for_every_tag() {
        let head = |s: &str| s.lines().take(2).collect::<Vec<_>>().join("\n");
        let base = full_status();
        let want_plain = head(&plain(&base, &[]));
        assert_eq!(
            want_plain,
            "🤖 Opus | 🧠 high·t | ♻️ 75% | 🔅 v2.1.293 | 🗿 v2.0 -> 🗿 v3.0\n\
             📡 owner/repo | ⬇️ 7 / 🔀 3 | 🅱️ main ↓2 +3"
        );
        let want_colored = head(&colored("mocha", Profile::TrueColor, &base));
        for (leader, name) in [
            (false, ""),
            (true, "leader"),
            (false, "\u{1b}]0;x\u{7}"),
            (true, "agent-27"),
        ] {
            let mut s = full_status();
            s.leader = leader;
            s.session_name = name.into();
            assert_eq!(head(&plain(&s, &[])), want_plain, "{leader} {name:?}");
            assert_eq!(
                head(&colored("mocha", Profile::TrueColor, &s)),
                want_colored,
                "{leader} {name:?}"
            );
        }
    }

    #[test]
    fn forge_counts_show_zero_but_hide_missing_suppressed_or_disabled_data() {
        let mut status = full_status();
        status.github = GitHubCounts {
            available: true,
            ..Default::default()
        };
        let with_zero = plain(&status, &[]);
        assert!(with_zero.contains("📡 owner/repo | ⬇️ 0 / 🔀 0 | 🅱️ main ↓2 +3"));
        // #2559: icons, not the old letter labels.
        assert!(!with_zero.contains("I: ") && !with_zero.contains("PR: "));
        status.github.available = false;
        let absent = plain(&status, &[]);
        assert!(
            absent.contains("📡 owner/repo | 🅱️ main ↓2 +3\n"),
            "{absent}"
        );
        assert!(!absent.contains("-/-"));
        status.github.available = true;
        status.github.suppressed = true;
        assert_eq!(plain(&status, &[]), absent);
        status.github.suppressed = false;
        assert_eq!(plain(&status, &["github"]), absent);
        assert!(!plain(&status, &["git_branch"]).contains("📡"));
        assert!(!plain(&status, &["git_status"]).contains("💾"));
        status.repo = None;
        assert!(plain(&status, &[]).contains("\n🅱️ main ↓2 +3\n"));
    }

    #[test]
    fn hidden_main_segments_cannot_be_enabled_by_configuration() {
        let home = tempfile::tempdir().unwrap();
        // The HOME state folder since #2455 (`$HOME/.moai-cli/state`).
        let state = home.path().join(crate::HOME_STATE_DIR).join("state");
        std::fs::create_dir_all(&state).unwrap();
        std::fs::write(
            state.join("last-session-state.json"),
            r#"{"active_task":{"command":"TASK_SENTINEL","spec_id":"SPEC","stage":"run"}}"#,
        )
        .unwrap();
        let home_path = home.path().to_string_lossy();
        let ctx = Context::for_test(&[("HOME", home_path.as_ref())], None);
        let input = input::parse(
            br#"{
                "model":{"display_name":"Opus"},
                "cwd":"/DIRECTORY_SENTINEL",
                "agent":{"name":"AGENT_SENTINEL"},
                "pr":{"number":987,"review_state":"approved"}
            }"#,
        )
        .unwrap();
        let status = state::collect(Some(&input), &ctx);
        let config = DisplayConfig {
            segments: ["directory", "pr", "task", "session"]
                .into_iter()
                .map(|key| (key.into(), true))
                .collect(),
            ..Default::default()
        };
        let output = Renderer::new(&config, None).render(&status);
        for hidden in ["📁", "💌", "👤", "📋", "SENTINEL"] {
            assert!(!output.contains(hidden), "{output:?}");
        }
        // #2756 AC-7: the pull request now has its own segment (not the Go `💌` one).
        assert!(output.contains("🔗 #987 approved"), "{output:?}");
    }

    /// #2756 AC-7–AC-9: the cache timer sits in the `♻️` segment, the worktree and the pull
    /// request on the repository line, each left out with its separator when absent.
    #[test]
    fn cache_timer_worktree_and_pr_segments() {
        let mut status = full_status();
        status.cache_timer = Some(CacheTimer::Warm(4));
        status.worktree = "wt".into();
        status.pr = Some(state::PrInfo {
            number: 42,
            url: "https://github.com/owner/repo/pull/42".into(),
            review_state: "pending".into(),
        });
        let out = plain(&status, &[]);
        let lines: Vec<_> = out.lines().collect();
        assert_eq!(
            lines[0],
            "🤖 Opus | 🧠 high·t | ♻️ 75% ⏱️ 4m | 🔅 v2.1.293 | 🗿 v2.0 -> 🗿 v3.0"
        );
        assert_eq!(
            lines[1],
            "📡 owner/repo | ⬇️ 7 / 🔀 3 | 🅱️ main ↓2 +3 | 🌳 wt | 🔗 #42 pending"
        );
        assert!(
            !out.contains('\x1b'),
            "plain output carries no link escape: {out:?}"
        );
        status.cache_timer = Some(CacheTimer::Cold);
        assert!(plain(&status, &[]).starts_with("🤖 Opus | 🧠 high·t | ♻️ 75% 🧊 | "));
        status.cache = None;
        assert!(
            plain(&status, &[]).contains("| ♻️ 🧊 |"),
            "a timer alone keeps the segment"
        );
        status.cache_timer = None;
        assert!(
            !plain(&status, &[]).contains('♻'),
            "neither value: no segment"
        );
        status.pr.as_mut().unwrap().review_state.clear();
        assert!(plain(&status, &[]).contains("| 🔗 #42\n"));
        assert!(!plain(&status, &["pr"]).contains('🔗'));
        assert!(!plain(&status, &["worktree"]).contains('🌳'));
        status.pr = None;
        status.worktree.clear();
        assert_eq!(
            plain(&status, &[]).lines().nth(1),
            Some("📡 owner/repo | ⬇️ 7 / 🔀 3 | 🅱️ main ↓2 +3"),
            "no empty `| |`"
        );
    }

    /// #2756 AC-7: with color the pull request is an OSC 8 hyperlink to its URL; without a URL
    /// it is plain text.
    #[test]
    fn the_pr_segment_links_to_its_url_only_with_color() {
        let mut status = full_status();
        status.pr = Some(state::PrInfo {
            number: 42,
            url: "https://github.com/owner/repo/pull/42".into(),
            review_state: String::new(),
        });
        let config = DisplayConfig::default();
        let colored = Renderer::new(&config, Some(Profile::Ansi256)).render(&status);
        assert!(
            colored.contains("\x1b]8;;https://github.com/owner/repo/pull/42\x1b\\"),
            "{colored:?}"
        );
        assert_eq!(colored.matches("\x1b]8;;\x1b\\").count(), 1, "{colored:?}");
        status.pr.as_mut().unwrap().url.clear();
        let unlinked = Renderer::new(&config, Some(Profile::Ansi256)).render(&status);
        assert!(!unlinked.contains("\x1b]8;"), "{unlinked:?}");
        assert!(unlinked.contains("🔗 #42"), "{unlinked:?}");
    }

    #[test]
    fn surviving_options_can_hide_every_line_with_nonempty_fallback() {
        let disabled = [
            "model",
            "effort_thinking",
            "cache_hit",
            "claude_version",
            "moai_version",
            "session_time",
            "output_style",
            "context",
            "usage_5h",
            "usage_7d",
            "session",
            "git_branch",
            "git_status",
        ];
        assert_eq!(plain(&full_status(), &disabled), "MoAI");
        let output = plain(&full_status(), &["session", "usage_5h", "usage_7d"]);
        assert_eq!(output.lines().count(), 3);
        assert_eq!(output.lines().nth(2), Some("🔋 CTX: █████░░░░░ 50%"));
        // #2756: with the repository line hidden, the bars line moves up to second.
        let output = plain(&full_status(), &["git_branch", "git_status"]);
        assert_eq!(output.lines().count(), 2);
        assert!(
            output.lines().nth(1).unwrap().starts_with("🏷️ agent | "),
            "{output}"
        );
    }

    #[test]
    fn context_warning_keeps_handoff_bands_and_goal_suppression() {
        let mut status = Status {
            auto_compact: 85,
            memory: segments::Memory {
                tokens_used: 180_000,
                token_budget: 170_000,
                context_window_size: 200_000,
                available: true,
            },
            ..Default::default()
        };
        let off = ["session"];
        assert_eq!(plain(&status, &off), "🪫 CTX: ██████████ 100% (⚠️ /clear)");
        // #2928: the tag in front leaves the bar and its mark as they were.
        assert_eq!(
            plain(&status, &[]),
            "🏷️ agent | 🪫 CTX: ██████████ 100% (⚠️ /clear)"
        );
        status.memory.tokens_used = 190_000;
        assert_eq!(plain(&status, &off), "🪫 CTX: ██████████ 100% (🛑/clear!)");
        status.goal_armed = true;
        assert_eq!(plain(&status, &off), "🪫 CTX: ██████████ 100%");
        assert_eq!(plain(&status, &["context", "session"]), "MoAI");
        assert_eq!(plain(&status, &["context"]), "🏷️ agent");
    }

    #[test]
    fn gradient_interpolates_like_go() {
        assert_eq!(interpolate(&MOCHA.stops, 0.0), (166, 227, 161));
        assert_eq!(interpolate(&MOCHA.stops, 0.5), (250, 203, 155));
        assert_eq!(interpolate(&MOCHA.stops, 1.0 / 9.0), (194, 227, 166));
        assert_eq!(interpolate(&MOCHA.stops, 1.0), (243, 139, 168));
    }

    #[test]
    fn base_color_wraps_each_line() {
        assert_eq!(
            apply_base_color("a\x1b[0mb\nc", "\x1b[38;5;1m"),
            "\x1b[22m\x1b[38;5;1ma\x1b[0m\x1b[38;5;1mb\x1b[0m\n\x1b[22m\x1b[38;5;1mc\x1b[0m"
        );
    }

    fn colored(theme_name: &str, profile: Profile, status: &Status) -> String {
        let config = DisplayConfig {
            theme: theme_name.into(),
            ..Default::default()
        };
        Renderer::new(&config, Some(profile)).render(status)
    }

    fn strip_sgr(text: &str) -> String {
        let mut out = String::new();
        let mut rest = text;
        while let Some(at) = rest.find("\x1b[") {
            out.push_str(&rest[..at]);
            let tail = &rest[at..];
            let end = tail.find('m').expect("SGR ends with m");
            rest = &tail[end + 1..];
        }
        out.push_str(rest);
        out
    }

    fn segment_colors(t: &Theme) -> [&'static str; 10] {
        [
            t.model, t.effort, t.cache, t.version, t.repo, t.issues, t.prs, t.branch, t.warn,
            t.alert,
        ]
    }

    /// #2559 (user decision 2026-11-12): each segment kind has its own color, distinct
    /// from the others and from the base text, in both themes and both palettes.
    #[test]
    fn each_segment_kind_has_its_own_color() {
        for (name, t) in [("catppuccin-mocha", MOCHA), ("catppuccin-latte", LATTE)] {
            for profile in [Profile::TrueColor, Profile::Ansi256] {
                let mut status = full_status();
                let out = colored(name, profile, &status);
                for (hex, text) in [
                    (t.model, "🤖 Opus"),
                    (t.effort, "🧠 high·t"),
                    (t.cache, "♻️ 75%"),
                    (t.version, "🔅 v2.1.293"),
                    (t.version, "🗿 v2.0 -> 🗿 v3.0"),
                    (t.repo, "📡 owner/repo"),
                    (t.issues, "⬇️ 7"),
                    (t.prs, "🔀 3"),
                    (t.branch, "🅱️ main ↓2 +3"),
                ] {
                    assert!(
                        out.contains(&profile.paint(hex, text)),
                        "{name} {profile:?}: {text:?} not painted {hex} in {out:?}"
                    );
                }
                // Escapes removed, the text is the NO_COLOR text (AC-2, AC-3 widths).
                assert_eq!(strip_sgr(&out), plain(&status, &[]));

                status.memory.token_budget = 200_000;
                status.memory.tokens_used = 180_000;
                let soft = colored(name, profile, &status);
                assert!(
                    soft.contains(&profile.paint(t.warn, "(⚠️ /clear)")),
                    "{soft:?}"
                );
                status.memory.tokens_used = 190_000;
                let hard = colored(name, profile, &status);
                assert!(
                    hard.contains(&profile.paint(t.alert, "(🛑/clear!)")),
                    "{hard:?}"
                );

                let mut seen: Vec<String> = vec![profile.foreground(t.text)];
                for hex in segment_colors(&t) {
                    let sgr = profile.foreground(hex);
                    assert!(
                        !seen.contains(&sgr),
                        "{name} {profile:?}: {hex} repeats {sgr}"
                    );
                    seen.push(sgr);
                }
            }
        }
    }

    #[test]
    fn unknown_theme_is_mocha() {
        assert_eq!(theme("catppuccin-latte"), LATTE);
        assert_eq!(theme("dracula"), MOCHA);
        assert_eq!(theme(""), MOCHA);
    }
}
