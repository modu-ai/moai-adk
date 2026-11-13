//! The main Claude Code statusline: up to three lines joined by `\n`, 10-cell
//! gradient bars, and two Catppuccin themes. Without color the text is identical
//! and carries no escape sequence.

use super::color::{hex_rgb, Profile};
use super::segments::{self, HandoffStage};
use super::state::{DisplayConfig, Status};
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
    pub fn render(&self, status: &Status) -> String {
        let lines: Vec<String> = [
            self.info_line(status),
            self.bars_line(status),
            self.git_line(status),
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
            if let Some(pct) = s
                .cache
                .and_then(|(read, creation)| segments::cache_hit_percent(read, creation))
            {
                segs.push(self.paint(self.theme.cache, &format!("♻️ {pct}%")));
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
        if self.enabled("session_time") && metrics {
            let t = segments::session_time(s.session_ms);
            if !t.is_empty() {
                segs.push(t);
            }
        }
        // #2559 AC-7 (user instruction 2026-11-12): the output style is not shown. The
        // `output_style` option name is still accepted in configs, it just shows nothing.
        segs.join(SEPARATOR)
    }

    fn bars_line(&self, s: &Status) -> String {
        let mut segs = Vec::new();
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
        if self.enabled("usage_5h") {
            if let Some(w) = s.five_hour {
                segs.push(self.usage_bar("5H:", w.used_percentage.unwrap_or(0.0) as i64));
            }
        }
        if self.enabled("usage_7d") {
            if let Some(w) = s.seven_day {
                segs.push(self.usage_bar("Week:", w.used_percentage.unwrap_or(0.0) as i64));
            }
        }
        let bars = segs.join(SEPARATOR);
        if self.enabled("session") && !s.session_name.is_empty() {
            if bars.is_empty() {
                format!("🏷️ {}", s.session_name)
            } else {
                format!("🏷️ {}  {bars}", s.session_name)
            }
        } else {
            bars
        }
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
        if self.enabled("git_status") {
            if let Some(g) = &s.git {
                segs.push(format!("💾 +{} M{} ?{}", g.staged, g.modified, g.untracked));
            }
        }
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

    #[test]
    fn full_main_layout_has_three_ordered_lines_without_reset_or_ahead() {
        assert_eq!(
            plain(&full_status(), &[]),
            "🤖 Opus | 🧠 high·t | ♻️ 75% | 🔅 v2.1.293 | 🗿 v2.0 -> 🗿 v3.0 | ⏳ 1h 30m\n\
             🏷️ release  🔋 CTX: █████░░░░░ 50% | 🔋 5H: ██░░░░░░░░ 25% | 🪫 Week: ███████░░░ 75%\n\
             📡 owner/repo | ⬇️ 7 / 🔀 3 | 🅱️ main ↓2 +3 | 💾 +1 M1 ?1"
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
        assert!(out.lines().next().unwrap().ends_with("⏳ 1h 30m"), "{out}");
    }

    #[test]
    fn rate_windows_are_optional_but_present_windows_keep_zero_percent() {
        let mut status = Status::default();
        assert_eq!(plain(&status, &[]), "CTX: ░░░░░░░░░░ n/a");
        status.five_hour = Some(RateWindow::default());
        assert_eq!(
            plain(&status, &[]),
            "CTX: ░░░░░░░░░░ n/a | 🔋 5H: ░░░░░░░░░░ 0%"
        );
        status.five_hour = None;
        status.seven_day = Some(RateWindow::default());
        assert_eq!(
            plain(&status, &[]),
            "CTX: ░░░░░░░░░░ n/a | 🔋 Week: ░░░░░░░░░░ 0%"
        );
        status.five_hour = Some(RateWindow::default());
        assert_eq!(
            plain(&status, &["context"]),
            "🔋 5H: ░░░░░░░░░░ 0% | 🔋 Week: ░░░░░░░░░░ 0%"
        );
        assert_eq!(
            plain(&status, &["usage_5h", "usage_7d"]),
            "CTX: ░░░░░░░░░░ n/a"
        );
    }

    /// #2559 AC-1: the bar's thresholds pinned at N/N+1 — the battery icon turns at 70/71 and
    /// a cell fills at each tenth (9/10). Plain output, so only the threshold decides.
    #[test]
    fn bar_thresholds_hold_at_their_boundaries() {
        let at = |pct: f64| {
            let status = Status {
                five_hour: Some(RateWindow {
                    used_percentage: Some(pct),
                    resets_at: None,
                }),
                ..Status::default()
            };
            plain(&status, &["context"])
        };
        assert_eq!(at(70.0), "🔋 5H: ███████░░░ 70%");
        assert_eq!(at(71.0), "🪫 5H: ███████░░░ 71%");
        assert_eq!(at(9.0), "🔋 5H: ░░░░░░░░░░ 9%");
        assert_eq!(at(10.0), "🔋 5H: █░░░░░░░░░ 10%");
    }

    #[test]
    fn session_prefix_uses_two_spaces_only_when_bars_follow() {
        let mut status = Status {
            session_name: "focus".into(),
            ..Default::default()
        };
        assert_eq!(plain(&status, &[]), "🏷️ focus  CTX: ░░░░░░░░░░ n/a");
        assert_eq!(plain(&status, &["context"]), "🏷️ focus");
        assert_eq!(plain(&status, &["session"]), "CTX: ░░░░░░░░░░ n/a");
        status.session_name.clear();
        assert_eq!(plain(&status, &[]), "CTX: ░░░░░░░░░░ n/a");
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
        assert!(absent.contains("📡 owner/repo | 🅱️ main ↓2 +3 | 💾 +1 M1 ?1"));
        assert!(!absent.contains("-/-"));
        status.github.available = true;
        status.github.suppressed = true;
        assert_eq!(plain(&status, &[]), absent);
        status.github.suppressed = false;
        assert_eq!(plain(&status, &["github"]), absent);
        assert!(!plain(&status, &["git_branch"]).contains("📡"));
        assert!(!plain(&status, &["git_status"]).contains("💾"));
        status.repo = None;
        assert!(plain(&status, &[]).contains("🅱️ main ↓2 +3 | 💾 +1 M1 ?1"));
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
        for hidden in ["📁", "💌", "👤", "📋", "SENTINEL", "987", "approved"] {
            assert!(!output.contains(hidden), "{output:?}");
        }
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
        assert_eq!(output.lines().nth(1), Some("🔋 CTX: █████░░░░░ 50%"));
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
        assert_eq!(plain(&status, &[]), "🪫 CTX: ██████████ 100% (⚠️ /clear)");
        status.memory.tokens_used = 190_000;
        assert_eq!(plain(&status, &[]), "🪫 CTX: ██████████ 100% (🛑/clear!)");
        status.goal_armed = true;
        assert_eq!(plain(&status, &[]), "🪫 CTX: ██████████ 100%");
        assert_eq!(plain(&status, &["context"]), "MoAI");
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
