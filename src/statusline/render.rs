//! Layout and color, ported from the Go `Renderer` (renderer.go, gradient.go, color.go,
//! theme.go): up to four lines joined by `\n`, segments joined by `" | "`, 10-cell bars
//! with a per-cell gradient, and two Catppuccin themes. Without color (`NO_COLOR` or
//! `MOAI_NO_COLOR`) the text is identical and carries no escape sequence.

use super::color::{hex_rgb, Profile};
use super::segments::{self, HandoffStage};
use super::state::{DisplayConfig, Status};
use std::collections::BTreeMap;

const BAR_WIDTH: i64 = 10;
const SEPARATOR: &str = " | ";
const SGR_NORMAL_INTENSITY: &str = "\x1b[22m";
const SGR_RESET: &str = "\x1b[0m";

/// A theme's colors (internal/tui/catppuccin.go). `stops` run green → yellow → peach → red.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub struct Theme {
    pub text: &'static str,
    pub muted: &'static str,
    pub success: &'static str,
    pub warning: &'static str,
    pub danger: &'static str,
    pub stops: [&'static str; 4],
}

pub const MOCHA: Theme = Theme {
    text: "#CDD6F4",
    muted: "#6C7086",
    success: "#A6E3A1",
    warning: "#F9E2AF",
    danger: "#F38BA8",
    stops: ["#A6E3A1", "#F9E2AF", "#FAB387", "#F38BA8"],
};

pub const LATTE: Theme = Theme {
    text: "#4C4F69",
    muted: "#9CA0B0",
    success: "#40A02B",
    warning: "#DF8E1D",
    danger: "#D20F39",
    stops: ["#40A02B", "#DF8E1D", "#FE640B", "#D20F39"],
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
            self.dir_git_line(status),
            self.session_line(status),
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
            segs.push(format!("🤖 {}", s.model));
        }
        if self.enabled("effort_thinking") {
            let thinking = s.thinking == Some(true);
            match &s.effort {
                None if thinking => segs.push("·t".into()),
                Some(level) if !level.is_empty() => {
                    segs.push(format!("🧠 {level}{}", if thinking { "·t" } else { "" }))
                }
                _ => {}
            }
        }
        if self.enabled("cache_hit") {
            if let Some(pct) = s
                .cache
                .and_then(|(read, creation)| segments::cache_hit_percent(read, creation))
            {
                segs.push(format!("♻️ {pct}%"));
            }
        }
        if self.enabled("claude_version") && !s.claude_version.is_empty() {
            segs.push(format!("🔅 v{}", s.claude_version));
        }
        if self.enabled("moai_version") {
            if let Some(v) = s.version.as_ref().filter(|v| !v.current.is_empty()) {
                let mut seg = format!("🗿 v{}", v.current);
                if let Some(latest) = &v.latest {
                    seg.push_str(&format!(" -> 🗿 v{latest}"));
                }
                segs.push(seg);
            }
        }
        if self.enabled("session_time") && metrics {
            let t = segments::session_time(s.session_ms);
            if !t.is_empty() {
                segs.push(t);
            }
        }
        if self.enabled("output_style") && !s.output_style.is_empty() {
            segs.push(format!("💬 {}", s.output_style));
        }
        segs.join(SEPARATOR)
    }

    fn bars_line(&self, s: &Status) -> String {
        let mut segs = Vec::new();
        if self.enabled("context") {
            if s.memory.available && s.memory.token_budget > 0 {
                let pct = segments::usage_percent(s.memory.tokens_used, s.memory.token_budget);
                let mut bar = self.usage_bar("CW:", pct);
                if !s.goal_armed {
                    match segments::handoff_stage(&s.memory, s.auto_compact) {
                        HandoffStage::Hard => bar.push_str(" (🛑/clear!)"),
                        HandoffStage::Soft => bar.push_str(" (⚠️/clear)"),
                        HandoffStage::None => {}
                    }
                }
                segs.push(bar);
            } else {
                segs.push(self.unknown_bar("CW:"));
            }
        }
        if self.enabled("usage_5h") {
            segs.push(match s.five_hour {
                Some(w) => self.bar_with_reset(
                    "5H:",
                    w.used_percentage.unwrap_or(0.0) as i64,
                    &segments::reset_relative(w.resets_at.unwrap_or(0), s.now_ms),
                ),
                None => self.unknown_bar("5H:"),
            });
        }
        if self.enabled("usage_7d") {
            segs.push(match s.seven_day {
                Some(w) => self.bar_with_reset(
                    "7D:",
                    w.used_percentage.unwrap_or(0.0) as i64,
                    &segments::reset_absolute(w.resets_at.unwrap_or(0)),
                ),
                None => self.unknown_bar("7D:"),
            });
        }
        segs.join(SEPARATOR)
    }

    /// `renderUsageBar`: battery icon, label, gradient bar, percentage.
    fn usage_bar(&self, label: &str, pct: i64) -> String {
        let icon = if pct > 70 { "🪫" } else { "🔋" };
        format!("{icon} {label} {} {pct}%", self.gradient_bar(pct))
    }

    fn bar_with_reset(&self, label: &str, pct: i64, reset: &str) -> String {
        let base = self.usage_bar(label, pct);
        if reset.is_empty() {
            base
        } else {
            format!("{base} ({reset})")
        }
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

    fn dir_git_line(&self, s: &Status) -> String {
        let mut segs = Vec::new();
        if self.enabled("directory") && !s.directory.is_empty() {
            segs.push(format!("📁 {}", s.directory));
        }
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
        if self.enabled("task") && !s.task.is_empty() {
            segs.push(format!("📋 {}", s.task));
        }
        if self.enabled("pr") {
            if let Some((number, state)) = s.pr.as_ref().filter(|(n, _)| *n != 0) {
                let mut seg = format!("💌 PR #{number}");
                if !state.is_empty() {
                    let text = format!("(⌥{state})");
                    let color = match state.as_str() {
                        "approved" => Some(self.theme.success),
                        "pending" => Some(self.theme.warning),
                        "changes_requested" => Some(self.theme.danger),
                        "draft" => Some(self.theme.muted),
                        _ => None,
                    };
                    let text = color.map_or(text.clone(), |c| self.paint(c, &text));
                    seg.push_str(&format!(" {text}"));
                }
                segs.push(seg);
            }
        }
        segs.join(SEPARATOR)
    }

    /// `renderRepoBranchSegment`: `📡 owner/name, i/p | 🅱️ branch ↑A ↓B +N`; the repo half
    /// needs `workspace.repo`, the whole segment needs a branch.
    fn repo_branch(&self, s: &Status) -> Option<String> {
        let git = s.git.as_ref().filter(|g| !g.branch.is_empty())?;
        let mut branch = format!("🅱️ {}", git.branch);
        if git.ahead > 0 {
            branch.push_str(&format!(" ↑{}", git.ahead));
        }
        if git.behind > 0 {
            branch.push_str(&format!(" ↓{}", git.behind));
        }
        let dirty = git.modified + git.staged + git.untracked;
        if dirty > 0 {
            branch.push_str(&format!(" +{dirty}"));
        }
        let Some((owner, name)) = s
            .repo
            .as_ref()
            .filter(|(o, n)| !o.is_empty() && !n.is_empty())
        else {
            return Some(branch);
        };
        let forge = if !self.enabled("github") || s.github.suppressed {
            String::new()
        } else if !s.github.available {
            ", -/-".to_string()
        } else {
            format!(", {}/{}", s.github.open_issues, s.github.open_prs)
        };
        Some(format!("📡 {owner}/{name}{forge} | {branch}"))
    }

    fn session_line(&self, s: &Status) -> String {
        let mut segs = Vec::new();
        if self.enabled("session") {
            if !s.session_name.is_empty() {
                segs.push(format!("🏷️ {}", s.session_name));
            }
            if !s.agent_name.is_empty() {
                segs.push(format!("👤 {}", s.agent_name));
            }
        }
        segs.join(SEPARATOR)
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

    #[test]
    fn unknown_theme_is_mocha() {
        assert_eq!(theme("catppuccin-latte"), LATTE);
        assert_eq!(theme("dracula"), MOCHA);
        assert_eq!(theme(""), MOCHA);
    }
}
