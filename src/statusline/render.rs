//! The main Claude Code statusline: up to three lines joined by `\n`, 10-cell
//! gradient bars, and two Catppuccin themes. Without color the text is identical
//! and carries no escape sequence.

use super::color::Profile;
use super::segments;
use super::state::{DisplayConfig, Status};
use std::collections::BTreeMap;

const SEPARATOR: &str = " | ";
/// Card C-137 (user instruction 2026-11-19), line 2: the two git counts, the items of the card
/// counts and the text after the repository and the branch icons are one space apart (leader
/// decision, card note 3570, decision D-1 of SPEC-20261120-statusline-line2-order: the two
/// spaces of the user's example are the terminal's drawing of the variation-selector icons).
/// The gap between the two git counts, between the card items and after the repository and the
/// branch icons is set here; the one space after each other icon sits in its own `format!`.
const GIT_COUNTS_GAP: &str = " ";
const CARD_ITEMS_GAP: &str = " ";
const REPO_ICON_GAP: &str = " ";
const BRANCH_ICON_GAP: &str = " ";
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
    /// `🧠` model (mauve).
    pub model: &'static str,
    /// `🤔` effort and thinking (lavender).
    pub effort: &'static str,
    /// `♻️` cache hit (teal).
    pub cache: &'static str,
    /// `🤖` Claude Code and `🗿` MoAI versions (sky).
    pub version: &'static str,
    /// `📡` repository (blue).
    pub repo: &'static str,
    /// `⬇️` open Issues (peach).
    pub issues: &'static str,
    /// `🔀` open pull requests (pink).
    pub prs: &'static str,
    /// `🅱️` branch with its behind and dirty counts (green).
    pub branch: &'static str,
    /// `▶️` cards being worked on (maroon).
    pub cards_active: &'static str,
    /// `🔃` cards pushed and waiting for integration or an audit (rosewater).
    pub cards_posted: &'static str,
    /// `🗂️` cards created and remaining (sapphire).
    pub cards_waiting: &'static str,
    /// `✅` cards done (teal).
    pub cards_done: &'static str,
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
    cards_active: "#EBA0AC",
    cards_posted: "#F5E0DC",
    cards_waiting: "#74C7EC",
    cards_done: "#94E2D5",
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
    cards_active: "#E64553",
    cards_posted: "#DC8A78",
    cards_waiting: "#209FB5",
    cards_done: "#179299",
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
    /// repository, line 3 the usage; #2928 (user instruction 6878415244, 2026-11-14): the
    /// `leader`/`agent` tag opens line 1. The user's three instructions of 2026-11-20
    /// (SPEC-20261120-statusline-layout): line 1 is tag, model, effort, the Claude Code
    /// version and the MoAI version; line 2 is the repository with its forge counts in one
    /// segment and the branch; line 3 opens with the card counts and carries the usage. The
    /// cache segment is not in the layout. SPEC-20261120-statusline-numbers-reset-worktree
    /// (the user's instruction of 2026-11-20T21:5xZ): the usage segments are numbers, the
    /// weekly window names its reset countdown, and the worktree label closes line 2. A line
    /// with nothing in it is left out, so the numbers above shift up by one for each empty
    /// line before it.
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
        // #2928 (user instruction 6878415244, 2026-11-14, replacing S-1's tag-first line 3): the
        // tag comes first on line 1 and is shown whenever the `session` segment is on, even
        // without stdin (decision Q2 = (가)). #2871 (leader decision 6907570513): its text is
        // the session name in the two exact shapes `tag_text` writes again — never the stdin
        // text itself.
        if self.enabled("session") {
            segs.push(format!("🏷️ {}", crate::ui::visible(&tag_text(s))));
        }
        let metrics = !s.model.is_empty();
        if self.enabled("model") && metrics {
            segs.push(self.paint(self.theme.model, &format!("🧠 {}", s.model)));
        }
        if self.enabled("effort_thinking") {
            let thinking = s.thinking == Some(true);
            match &s.effort {
                None if thinking => segs.push(self.paint(self.theme.effort, "·t")),
                Some(level) if !level.is_empty() => segs.push(self.paint(
                    self.theme.effort,
                    &format!("🤔 {level}{}", if thinking { "·t" } else { "" }),
                )),
                _ => {}
            }
        }
        // D-1 of SPEC-20261120-statusline-layout (user instruction 2026-11-20): the `♻️` cache
        // segment is not in the layout. The `cache_hit` option name is still accepted in
        // configs, it just shows nothing — like `session_time` and `output_style` — and the
        // cache value and its warm or cold timer stay in the collected state. The card counts
        // open line 3 and the worktree label closes it ([`Renderer::bars_line`], the user's
        // third instruction of 2026-11-20); the Claude Code version stands on line 1 between
        // the effort and the MoAI version.
        if self.enabled("claude_version") && !s.claude_version.is_empty() {
            segs.push(self.paint(self.theme.version, &format!("🤖 v{}", s.claude_version)));
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
        // #2928 (user instruction 6878415244, 2026-11-14): line 3 is the usage line; the card
        // counts open it (the user's third instruction of 2026-11-20). The user's instruction
        // of 2026-11-20T21:5xZ (SPEC-20261120-statusline-numbers-reset-worktree): the usage
        // segments are numbers now — no bar graph, no colon, no `/clear` hint — the weekly
        // window names its reset countdown, and the worktree label moved to the end of line 2.
        if let Some(cards) = self.cards(s) {
            segs.push(cards);
        }
        if self.enabled("context") {
            if s.memory.available && s.memory.token_budget > 0 {
                let pct = segments::usage_percent(s.memory.tokens_used, s.memory.token_budget);
                segs.push(self.battery_text("CTX", pct));
            } else {
                segs.push(self.paint(self.theme.muted, "CTX n/a"));
            }
        }
        // #2756 leader decision 6862179411 item 1: the five-hour usage is not shown (the user
        // did not choose it, 6858992949); the `usage_5h` option is still accepted.
        if self.enabled("usage_7d") {
            if let Some(w) = s.seven_day {
                let pct = w.used_percentage.unwrap_or(0.0) as i64;
                let mut text = self.battery_text("Week", pct);
                if let Some(resets_at) = w.resets_at {
                    text.push(' ');
                    text.push_str(&segments::reset_countdown(resets_at, s.now));
                }
                segs.push(text);
            }
        }
        // The renderer never shortens a segment and does not read `COLUMNS`, which Claude
        // Code sets; whether an overlong line is cut or wraps is left open by the host docs
        // ("may get truncated or wrap awkwardly").
        segs.join(SEPARATOR)
    }

    /// The numbered usage segment: the battery icon (🪫 above 70, the user's 🔋→🪫 switch
    /// of 2026-11-20 kept), the label without a colon and the percentage.
    fn battery_text(&self, label: &str, pct: i64) -> String {
        let icon = if pct > 70 { "🪫" } else { "🔋" };
        format!("{icon} {label} {pct}%")
    }

    fn git_line(&self, s: &Status) -> String {
        let mut segs = Vec::new();
        // SPEC-20261120-statusline-layout (user instruction 2026-11-20): line 2 opens with the
        // repository and its forge counts in one segment, then the branch; the card counts
        // open line 3 ([`Renderer::bars_line`], the user's third instruction of 2026-11-20).
        if let Some(g) = self
            .enabled("git_branch")
            .then(|| self.repo_groups(s))
            .flatten()
        {
            if let Some(repo) = g.repo_with_forge {
                segs.push(repo);
            }
            segs.push(g.branch);
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
        // #2756 leader decision 6862179411 item 2: no `💾` file-change segment (the branch's
        // `+N` carries the change count); the `git_status` option is still accepted.
        // SPEC-20261120-statusline-numbers-reset-worktree (the user's instruction of
        // 2026-11-20T21:5xZ): the worktree label closes line 2 now (was: the end of line 3);
        // in the main checkout it is absent.
        if self.enabled("worktree") && !s.worktree.is_empty() {
            segs.push(self.paint(self.theme.branch, &worktree_text(s)));
        }
        segs.join(SEPARATOR)
    }

    /// Card C-65 (user instructions 2026-11-16 and 2026-11-17, card 85; order and icons by the
    /// user's decision 2026-11-19; one space between the items and no slash, card C-137): it
    /// opens line 3 since the user's third instruction of 2026-11-20
    /// (SPEC-20261120-statusline-layout, was: line 1, then line 2),
    /// `🗂️ <remaining> ▶️ <working> 🔃 <waiting for integration> ✅ <done>` from the project's
    /// local ledger, under the one `cards` option;
    /// `🔃` is left out when the notes could not be read. A ledger without cards shows zeros;
    /// `None` when there is no ledger to read. Each number has its icon, so color is never the
    /// only signal.
    fn cards(&self, s: &Status) -> Option<String> {
        let c = s.cards.filter(|_| self.enabled("cards"))?;
        let remaining = self.paint(self.theme.cards_waiting, &format!("🗂️ {}", c.remaining));
        let working = self.paint(self.theme.cards_active, &format!("▶️ {}", c.in_progress));
        let done = self.paint(self.theme.cards_done, &format!("✅ {}", c.done));
        // `🔃` only when the notes could be read; otherwise three numbers in the same order and
        // with the same icons, never a guess.
        Some(match c.waiting {
            Some(w) => [
                remaining,
                working,
                self.paint(self.theme.cards_posted, &format!("🔃 {w}")),
                done,
            ]
            .join(CARD_ITEMS_GAP),
            None => [remaining, working, done].join(CARD_ITEMS_GAP),
        })
    }

    /// The three groups that retain the `git_branch` switch: the forge counts and the repository
    /// (both only with a repository name) and the branch; `None` without a git branch. Their
    /// order on the line, and where the card counts go, is [`Renderer::git_line`]'s (card C-137).
    fn repo_groups(&self, s: &Status) -> Option<RepoGroups> {
        let git = s.git.as_ref().filter(|g| !g.branch.is_empty())?;
        let mut branch = format!("🅱️{BRANCH_ICON_GAP}{}", git.branch);
        // Under M-Flow the merges land on the local integration branch first and the remote
        // moves at a release, so the ahead count is the number that shows them landing
        // (SPEC-20261120-statusline-ahead); the order is `↑N ↓N +N`, each hidden at 0.
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
        let branch = self.paint(self.theme.branch, &branch);
        let Some((owner, name)) = s
            .repo
            .as_ref()
            .filter(|(o, n)| !o.is_empty() && !n.is_empty())
        else {
            return Some(RepoGroups {
                repo_with_forge: None,
                branch,
            });
        };
        let forge =
            (self.enabled("github") && s.github.available && !s.github.suppressed).then(|| {
                // #2559 (user instructions 2026-11-12): icons instead of letter labels; the Issue
                // icon is `⬇️` (U+2B07 U+FE0F) since 23:1xZ, the PR icon stays `🔀`.
                format!(
                    "{}{GIT_COUNTS_GAP}{}",
                    self.paint(self.theme.issues, &format!("⬇️ {}", s.github.open_issues)),
                    self.paint(self.theme.prs, &format!("🔀 {}", s.github.open_prs))
                )
            });
        let repo = self.paint(self.theme.repo, &format!("📡{REPO_ICON_GAP}{owner}/{name}"));
        // SPEC-20261120-statusline-layout (user instruction 2026-11-20): the forge counts join
        // the repository in one segment, the comma right after the name; without them the name
        // stands alone. Each part keeps its own color.
        let repo_with_forge = match forge {
            Some(f) => format!("{repo}, {f}"),
            None => repo,
        };
        Some(RepoGroups {
            repo_with_forge: Some(repo_with_forge),
            branch,
        })
    }
}

/// The groups of line 2 that follow the `git_branch` switch, each already painted
/// (SPEC-20261120-statusline-layout: the repository and its forge counts are one segment).
struct RepoGroups {
    repo_with_forge: Option<String>,
    branch: String,
}

/// User instruction 2026-11-19: `🌳 WT-c<card>-<name>` when a held reservation names the
/// worktree, `🌳 WT-c<card>` when nothing says what the work is, else the worktree name the host
/// gave, alone.
fn worktree_text(s: &Status) -> String {
    match &s.worktree_label {
        Some(l) if l.name.is_empty() => format!("🌳 WT-c{}", l.card),
        Some(l) => format!("🌳 WT-c{}-{}", l.card, l.name),
        None => format!("🌳 {}", s.worktree),
    }
}

/// The text of the `🏷️` tag (#2871, leader decision 6907570513, replacing #2928's "no name
/// text"). The stdin `session_name` is free text (escape and control bytes, #2818 B1), so it is
/// never copied: only two exact shapes are read and written again by this code.
/// - A live registry `leader` entry (`s.leader`) or the name exactly `leader` → `Leader`. The
///   name is only a display hint and is never an authority decision.
/// - `agent-<decimal u32 without a leading zero>` (`session::agent_number_of`) → `Agent-<n>`.
/// - Any other name, or none → `agent`.
fn tag_text(s: &Status) -> String {
    if s.leader || s.session_name == "leader" {
        return "Leader".to_string();
    }
    match crate::session::agent_number_of(&s.session_name, "") {
        Some(n) => format!("Agent-{n}"),
        None => "agent".to_string(),
    }
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
    use crate::statusline::state::{self, CacheTimer, CardCounts, Context, GitHubCounts, Version};

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
                resets_at: None,
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

    /// #2756 AC-1–AC-2 (user instructions 2026-11-13) and #2928 (user instruction 6878415244,
    /// 2026-11-14), reordered by the user instruction 2026-11-20: line 1 is the tag, the
    /// environment and the MoAI version, line 2 opens with the repository and its forge counts
    /// and line 3 is the usage, each line joined by ` | `. The cache segment is not in the
    /// layout (D-1 of SPEC-20261120-statusline-layout). Under M-Flow the merges land locally
    /// first, so the ahead count shows before behind (SPEC-20261120-statusline-ahead).
    /// SPEC-20261120-statusline-numbers-reset-worktree: the usage segments are numbers.
    #[test]
    fn full_main_layout_has_three_ordered_lines_with_ahead_before_behind() {
        assert_eq!(
            plain(&full_status(), &[]),
            "🏷️ agent | 🧠 Opus | 🤔 high·t | 🤖 v2.1.293 | 🗿 v2.0 -> 🗿 v3.0\n\
             📡 owner/repo, ⬇️ 7 🔀 3 | 🅱️ main ↑5 ↓2 +3\n\
             🔋 CTX 50% | 🪫 Week 75%"
        );
    }

    /// With ahead 0 the line is byte-identical to what it renders without the feature (no
    /// `↑`, nothing else moves).
    #[test]
    fn ahead_zero_keeps_the_line_byte_identical() {
        let mut status = full_status();
        status.git = Some(segments::parse_porcelain(
            "## main...origin/main [behind 2]\n M a\nA  b\n?? c\n",
        ));
        let line = plain(&status, &[]).lines().nth(1).unwrap().to_string();
        assert_eq!(line, "📡 owner/repo, ⬇️ 7 🔀 3 | 🅱️ main ↓2 +3");
        assert!(!line.contains('↑'), "{line}");
    }

    /// Ahead without behind: only `↑N` appears, and dirty stays last.
    #[test]
    fn ahead_without_behind_shows_only_up() {
        let mut status = full_status();
        status.git = Some(segments::parse_porcelain(
            "## main...origin/main [ahead 4]\n M a\n?? b\n",
        ));
        let line = plain(&status, &[]).lines().nth(1).unwrap().to_string();
        assert_eq!(line, "📡 owner/repo, ⬇️ 7 🔀 3 | 🅱️ main ↑4 +2");
    }

    /// The user's example (leader conversation, 2026-11-20T21:5xZ, the mock-up the user
    /// picked): every segment present gives exactly these three lines in plain output
    /// (SPEC-20261120-statusline-numbers-reset-worktree). The render clock is fixed, so the
    /// countdown is deterministic: `resets_at` is 6 days and 21 hours ahead of `now`.
    fn example_status() -> Status {
        Status {
            model: "Opus 5.5".into(),
            effort: Some("high".into()),
            thinking: Some(true),
            cache: Some((90, 10)),
            claude_version: "2.1.325".into(),
            version: Some(Version {
                current: "4.2647.10".into(),
                latest: None,
            }),
            session_name: "leader".into(),
            leader: true,
            memory: segments::Memory {
                tokens_used: 97,
                token_budget: 100,
                context_window_size: 200_000,
                available: true,
            },
            seven_day: Some(RateWindow {
                used_percentage: Some(21.0),
                resets_at: Some(1_800_594_000),
            }),
            now: 1_800_000_000,
            repo: Some(("modu-ai".into(), "moai-adk".into())),
            github: GitHubCounts {
                available: true,
                open_issues: 0,
                open_prs: 0,
                ..Default::default()
            },
            git: Some(segments::parse_porcelain(
                "## v2026.11.0...origin/v2026.11.0 [ahead 207]\n M a\n M b\n M c\n M d\n M e\n M f\n M g\n M h\nA  i\nA  j\n?? k\n",
            )),
            cards: Some(CardCounts {
                in_progress: 24,
                waiting: Some(21),
                remaining: 105,
                done: 76,
            }),
            worktree: "c135".into(),
            ..Default::default()
        }
    }

    #[test]
    fn the_user_example_gives_exactly_the_three_lines() {
        assert_eq!(
            plain(&example_status(), &[]),
            "🏷️ Leader | 🧠 Opus 5.5 | 🤔 high·t | 🤖 v2.1.325 | 🗿 v4.2647.10\n\
             📡 modu-ai/moai-adk, ⬇️ 0 🔀 0 | 🅱️ v2026.11.0 ↑207 +11 | 🌳 c135\n\
             🗂️ 105 ▶️ 24 🔃 21 ✅ 76 | 🪫 CTX 97% | 🔋 Week 21% (resets in 6d 21h)"
        );
    }

    /// Criterion 1's no-worktree variant: the same example from the main checkout loses only
    /// the label at the end of line 2.
    #[test]
    fn the_user_example_without_a_worktree_loses_only_the_label() {
        let mut status = example_status();
        status.worktree.clear();
        assert_eq!(
            plain(&status, &[]),
            "🏷️ Leader | 🧠 Opus 5.5 | 🤔 high·t | 🤖 v2.1.325 | 🗿 v4.2647.10\n\
             📡 modu-ai/moai-adk, ⬇️ 0 🔀 0 | 🅱️ v2026.11.0 ↑207 +11\n\
             🗂️ 105 ▶️ 24 🔃 21 ✅ 76 | 🪫 CTX 97% | 🔋 Week 21% (resets in 6d 21h)"
        );
    }

    /// Criterion 1's no-resets_at variant: the weekly segment is the bare number.
    #[test]
    fn the_user_example_without_a_reset_shows_the_bare_week_number() {
        let mut status = example_status();
        status.seven_day.as_mut().unwrap().resets_at = None;
        assert_eq!(
            plain(&status, &[]).lines().nth(2).unwrap(),
            "🗂️ 105 ▶️ 24 🔃 21 ✅ 76 | 🪫 CTX 97% | 🔋 Week 21%"
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
        assert!(out.contains("🅱️ main ↑5 ↓2 +3"), "{out}");
    }

    #[test]
    fn rate_windows_are_optional_but_present_windows_keep_zero_percent() {
        let mut status = Status::default();
        assert_eq!(plain(&status, &["session"]), "CTX n/a");
        status.seven_day = Some(RateWindow::default());
        assert_eq!(plain(&status, &["session"]), "CTX n/a | 🔋 Week 0%");
        assert_eq!(plain(&status, &["session", "context"]), "🔋 Week 0%");
        assert_eq!(plain(&status, &["session", "usage_7d"]), "CTX n/a");
    }

    /// #2559 AC-1: the battery icon's threshold pinned at N/N+1 — it turns at 70/71
    /// (SPEC-20261120-statusline-numbers-reset-worktree: the number form keeps the switch).
    #[test]
    fn battery_icon_thresholds_hold_at_their_boundaries() {
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
        assert_eq!(at(70.0), "🔋 Week 70%");
        assert_eq!(at(71.0), "🪫 Week 71%");
    }

    /// The renderer never shortens for width and does not read `COLUMNS`; since #2928 the
    /// stdin name is not printed at all, so a 200-character name leaves lines 1 and 3 unchanged.
    #[test]
    fn a_long_name_changes_nothing_on_lines_one_and_three() {
        let mut status = full_status();
        status.session_name = "x".repeat(200);
        let out = plain(&status, &[]);
        assert_eq!(
            out.lines().next(),
            Some("🏷️ agent | 🧠 Opus | 🤔 high·t | 🤖 v2.1.293 | 🗿 v2.0 -> 🗿 v3.0")
        );
        assert_eq!(out.lines().nth(2), Some("🔋 CTX 50% | 🪫 Week 75%"));
    }

    /// #2928 (user instruction 6878415244, 2026-11-14): line 1 is `🏷️ <tag> | 🧠 … | 🗿 …` and
    /// line 3 is `CTX … | Week …` alone. The tag (`Leader`, `Agent-<n>` or `agent`, #2871) comes first; it
    /// is shown whenever the `session` segment is on, even with no stdin (decision Q2 = (가));
    /// with every other line-1 segment off it stands alone, and with `session` off line 1 is as
    /// before. The usage segments are the number form
    /// (SPEC-20261120-statusline-numbers-reset-worktree).
    #[test]
    fn the_tag_is_the_first_segment_of_line_one() {
        let mut status = Status::default();
        assert_eq!(plain(&status, &[]), "🏷️ agent\nCTX n/a");
        status.seven_day = Some(RateWindow::default());
        assert_eq!(plain(&status, &[]), "🏷️ agent\nCTX n/a | 🔋 Week 0%");
        status.leader = true;
        assert_eq!(plain(&status, &[]), "🏷️ Leader\nCTX n/a | 🔋 Week 0%");
        assert_eq!(plain(&status, &["usage_7d", "context"]), "🏷️ Leader");
        assert_eq!(plain(&status, &["session"]), "CTX n/a | 🔋 Week 0%");
        let mut full = full_status();
        let usage = "🔋 CTX 50% | 🪫 Week 75%";
        let env = "🧠 Opus | 🤔 high·t | 🤖 v2.1.293 | 🗿 v2.0 -> 🗿 v3.0";
        assert_eq!(
            plain(&full, &[]).lines().next(),
            Some(format!("🏷️ agent | {env}").as_str())
        );
        assert_eq!(plain(&full, &[]).lines().nth(2), Some(usage));
        assert_eq!(plain(&full, &["session"]).lines().next(), Some(env));
        assert_eq!(plain(&full, &["session"]).lines().nth(2), Some(usage));
        let line_one_off = [
            "model",
            "effort_thinking",
            "cache_hit",
            "claude_version",
            "moai_version",
        ];
        let only_tag = plain(&full, &line_one_off);
        assert_eq!(only_tag.lines().next(), Some("🏷️ agent"), "{only_tag}");
        // `claude_version` is off in this list, so line 3 ends with the usage.
        assert_eq!(only_tag.lines().nth(2), Some(usage), "{only_tag}");
        full.leader = true;
        assert_eq!(
            plain(&full, &[]).lines().next(),
            Some(format!("🏷️ Leader | {env}").as_str())
        );
        assert_eq!(plain(&full, &[]).lines().nth(2), Some(usage));
        let colored = strip_sgr(&colored("mocha", Profile::TrueColor, &full));
        assert!(
            colored.lines().next().unwrap().starts_with("🏷️ Leader | "),
            "{colored:?}"
        );
        assert_eq!(colored.lines().nth(2), Some(usage), "{colored:?}");
    }

    /// #2871 (leader decision 6907570513, replacing #2928's "never print the name"): the tag is
    /// the session name only in two exact shapes that this code writes again itself —
    /// `agent-<decimal u32 without a leading zero>` → `Agent-<n>`, and exactly `leader` →
    /// `Leader` — and `agent` for every other name. The name's own bytes are never copied.
    #[test]
    fn the_tag_rewrites_only_the_exact_agent_and_leader_names() {
        let tag_of = |name: &str| {
            let mut status = full_status();
            status.session_name = name.into();
            plain(&status, &[]).lines().next().unwrap().to_string()
        };
        let line = |tag: &str| {
            format!("🏷️ {tag} | 🧠 Opus | 🤔 high·t | 🤖 v2.1.293 | 🗿 v2.0 -> 🗿 v3.0")
        };
        for (name, want) in [
            ("agent-27", "Agent-27"),
            ("agent-0", "Agent-0"),
            ("agent-1", "Agent-1"),
            ("agent-4294967295", "Agent-4294967295"),
            ("leader", "Leader"),
        ] {
            assert_eq!(tag_of(name), line(want), "{name:?}");
        }
        let long = format!("agent-{}", "9".repeat(59));
        assert_eq!(long.len(), 65);
        for name in [
            "agent-4294967296",
            "agent-027",
            "agent-00",
            "agent-1;x",
            "agent-\u{663}",
            "agent-+5",
            "agent--5",
            "agent-5 ",
            "agent-5\n",
            "agent-",
            "agent",
            "AGENT-1",
            "Agent-1",
            "Leader",
            "LEADER",
            "leader ",
            "leader-2",
            "\u{1b}]0;agent-5\u{7}",
            "agent-5\u{9b}31m",
            "#2388·#2403 Issue 닫기 대조",
            "/moai",
            "",
            long.as_str(),
        ] {
            assert_eq!(tag_of(name), line("agent"), "{name:?}");
        }
        // 64 and 65 characters of a name that is no exact shape: the same fixed word.
        for len in [64, 65] {
            assert_eq!(tag_of(&"x".repeat(len)), line("agent"), "{len}");
        }
        // The exact `leader` name or a live registry entry gives `Leader`; the registry wins
        // over an agent-shaped name, as before.
        let mut status = full_status();
        status.leader = true;
        status.session_name = "agent-27".into();
        assert_eq!(
            plain(&status, &[]).lines().next(),
            Some(line("Leader").as_str())
        );
    }

    /// #2928 S-3 (#2818 B1): the stdin `session_name` — a lane name, a terminal title, escape
    /// or C1 bytes — never reaches the output as itself; only the tag does, and a name that is
    /// no exact shape (see `the_tag_rewrites_only_the_exact_agent_and_leader_names`) is `agent`.
    #[test]
    fn session_name_text_never_reaches_the_tag() {
        for name in [
            "\u{1b}]0;evil\u{7}",
            "x\u{9b}31m",
            "#2388·#2403·#2385 Issue 닫기 대조",
            "/moai",
        ] {
            let mut status = full_status();
            status.session_name = name.into();
            let out = plain(&status, &[]);
            assert_eq!(
                out.lines().next(),
                Some("🏷️ agent | 🧠 Opus | 🤔 high·t | 🤖 v2.1.293 | 🗿 v2.0 -> 🗿 v3.0"),
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

    /// #2928 S-4, revised by user instruction 6878415244 (2026-11-14): lines 2 and 3 are the
    /// same bytes whatever the tag and the stdin name, plain and colored, and line 1 differs
    /// only by its first segment, the tag.
    #[test]
    fn lines_two_and_three_are_identical_and_line_one_differs_only_by_its_tag() {
        let tail = |s: &str| s.lines().skip(1).collect::<Vec<_>>().join("\n");
        let first = |s: &str| s.lines().next().unwrap().to_string();
        let base = full_status();
        let base_plain = plain(&base, &[]);
        let want_plain = tail(&base_plain);
        assert_eq!(
            want_plain,
            "📡 owner/repo, ⬇️ 7 🔀 3 | 🅱️ main ↑5 ↓2 +3\n\
             🔋 CTX 50% | 🪫 Week 75%"
        );
        let base_colored = colored("mocha", Profile::TrueColor, &base);
        let want_colored = tail(&base_colored);
        for text in [&base_plain, &base_colored] {
            assert_eq!(first(text).matches("🏷️ agent").count(), 1, "{text:?}");
        }
        for (leader, name, tag) in [
            (false, "", "🏷️ agent"),
            (true, "leader", "🏷️ Leader"),
            (false, "leader", "🏷️ Leader"),
            (false, "\u{1b}]0;x\u{7}", "🏷️ agent"),
            (true, "agent-27", "🏷️ Leader"),
            (false, "agent-27", "🏷️ Agent-27"),
        ] {
            let mut s = full_status();
            s.leader = leader;
            s.session_name = name.into();
            let got_plain = plain(&s, &[]);
            assert_eq!(tail(&got_plain), want_plain, "{leader} {name:?}");
            assert_eq!(
                first(&got_plain),
                first(&base_plain).replacen("🏷️ agent", tag, 1),
                "{leader} {name:?}"
            );
            let got_colored = colored("mocha", Profile::TrueColor, &s);
            assert_eq!(tail(&got_colored), want_colored, "{leader} {name:?}");
            assert_eq!(
                first(&got_colored),
                first(&base_colored).replacen("🏷️ agent", tag, 1),
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
        assert!(with_zero.contains("📡 owner/repo, ⬇️ 0 🔀 0 | 🅱️ main ↑5 ↓2 +3"));
        // #2559: icons, not the old letter labels.
        assert!(!with_zero.contains("I: ") && !with_zero.contains("PR: "));
        status.github.available = false;
        let absent = plain(&status, &[]);
        assert!(
            absent.contains("📡 owner/repo | 🅱️ main ↑5 ↓2 +3\n"),
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
        assert!(plain(&status, &[]).contains("\n🅱️ main ↑5 ↓2 +3\n"));
    }

    /// Card C-65 (user instructions 2026-11-16 and 2026-11-17, card 85), moved to line 3 by the
    /// user's third instruction of 2026-11-20: `🗂️ remaining ▶️ in progress ✅ done` opens
    /// line 3 before the usage bars; zero is shown, and absent counts leave line 3 as it was.
    #[test]
    fn the_card_counts_open_line_three_before_the_usage_bars() {
        // The last line: line 3 in the full render, and the same line one up when the git
        // line is gone.
        let line3 = |s: &Status, off: &[&str]| plain(s, off).lines().last().unwrap().to_string();
        let mut status = full_status();
        let without = plain(&status, &[]);
        assert_eq!(
            line3(&status, &[]),
            "🔋 CTX 50% | 🪫 Week 75%",
            "no counts: the usage alone"
        );
        status.cards = Some(CardCounts {
            in_progress: 5,
            waiting: None,
            remaining: 50,
            done: 12,
        });
        assert_eq!(
            line3(&status, &[]),
            "🗂️ 50 ▶️ 5 ✅ 12 | 🔋 CTX 50% | 🪫 Week 75%"
        );
        // Lines 1 and 2 do not change.
        let with = plain(&status, &[]);
        assert_eq!(with.lines().next(), without.lines().next());
        assert_eq!(with.lines().nth(1), without.lines().nth(1));
        assert_eq!(with.lines().count(), 3);
        // The `cards` option hides the group and nothing else.
        assert_eq!(plain(&status, &["cards"]), without);
        // The counts come from the ledger, not from git: forge counts, the repository and the
        // branch change line 2, never the place of the group on line 3.
        let with_cards = with.lines().nth(2).unwrap().to_string();
        assert_eq!(line3(&status, &["github"]), with_cards);
        status.github.suppressed = true;
        assert_eq!(line3(&status, &[]), with_cards);
        status.github = GitHubCounts::default();
        assert_eq!(line3(&status, &[]), with_cards);
        status.repo = None;
        assert_eq!(line3(&status, &[]), with_cards);
        status.git = None;
        assert_eq!(line3(&status, &[]), with_cards);
        // An empty ledger is not an absent one.
        status.cards = Some(CardCounts::default());
        assert_eq!(
            line3(&status, &[]),
            "🗂️ 0 ▶️ 0 ✅ 0 | 🔋 CTX 50% | 🪫 Week 75%"
        );
    }

    /// SPEC-20261120-statusline-layout (user instruction 2026-11-20): line 2 opens with one
    /// segment `📡 owner/repo, ⬇️ n 🔀 n` — the comma right after the repository name, the
    /// forge counts in the same segment — then the branch, then the pull request, groups
    /// joined by ` | `, and no group has a slash in it. The card counts stand on line 3
    /// (the user's third instruction of 2026-11-20).
    #[test]
    fn line_two_reads_the_repository_with_the_forge_counts_then_the_branch_then_the_pr() {
        let line = |s: &Status, off: &[&str]| plain(s, off).lines().nth(1).unwrap().to_string();
        let mut status = full_status();
        status.cards = Some(CardCounts {
            in_progress: 5,
            waiting: Some(15),
            remaining: 50,
            done: 65,
        });
        status.pr = Some(state::PrInfo {
            number: 42,
            url: String::new(),
            review_state: "pending".into(),
        });
        let full = line(&status, &[]);
        assert_eq!(
            full,
            "📡 owner/repo, ⬇️ 7 🔀 3 | 🅱️ main ↑5 ↓2 +3 | 🔗 #42 pending"
        );
        assert!(
            !full.replace("owner/repo", "").contains('/'),
            "no slash inside a group: {full}"
        );
        let at = |icon: &str| full.find(icon).unwrap();
        assert!(at("📡") < at("⬇️") && at("⬇️") < at("🔀") && at("🔀") < at("🅱️"));
        assert!(at("🅱️") < at("🔗"));
        // The comma sits right after the repository name, inside the segment: no separator
        // between the repository and the forge counts.
        assert!(full.contains("owner/repo, ⬇️ 7 🔀 3 | "), "{full}");
        // Line 1 keeps no card counts: the effort group runs into the Claude Code version and
        // then the MoAI version.
        assert_eq!(
            plain(&status, &[]).lines().next().unwrap(),
            "🏷️ agent | 🧠 Opus | 🤔 high·t | 🤖 v2.1.293 | 🗿 v2.0 -> 🗿 v3.0"
        );
        // Line 3 opens with the card counts, and the waiting count takes its own number
        // between the working and the done number.
        assert_eq!(
            plain(&status, &[]).lines().nth(2).unwrap(),
            "🗂️ 50 ▶️ 5 🔃 15 ✅ 65 | 🔋 CTX 50% | 🪫 Week 75%"
        );
        // Notes unreadable: the same place, three items.
        status.cards.as_mut().unwrap().waiting = None;
        assert_eq!(
            plain(&status, &[]).lines().nth(2).unwrap(),
            "🗂️ 50 ▶️ 5 ✅ 65 | 🔋 CTX 50% | 🪫 Week 75%"
        );
        status.cards.as_mut().unwrap().waiting = Some(15);
        // Without the forge counts the repository name stands alone, with no comma.
        assert_eq!(
            line(&status, &["github"]),
            "📡 owner/repo | 🅱️ main ↑5 ↓2 +3 | 🔗 #42 pending"
        );
        // The `cards` switch hides the group and nothing else.
        assert_eq!(
            plain(&status, &["cards"]).lines().nth(2).unwrap(),
            "🔋 CTX 50% | 🪫 Week 75%"
        );
        // The pull request keeps the last place of line 2.
        assert_eq!(
            line(&status, &["pr"]),
            "📡 owner/repo, ⬇️ 7 🔀 3 | 🅱️ main ↑5 ↓2 +3"
        );
    }

    /// Card C-65: the counts come from the local ledger, not from the repository, so they are
    /// shown without a repository name, without a branch and with `git_branch` off.
    /// A known waiting count gives the four-number form with `🔃` between the working and the
    /// done number; `None` (notes unreadable) gives the three-number form without it, and
    /// `cards` off hides all of it.
    #[test]
    fn waiting_cards_get_their_own_number_on_line_three_and_none_gives_the_three_number_form() {
        // The last line: line 3 in the full render, and the same line one up when the git
        // line is gone.
        let line3 = |s: &Status, off: &[&str]| plain(s, off).lines().last().unwrap().to_string();
        let mut status = full_status();
        status.worktree = "wt".into();
        status.cards = Some(CardCounts {
            in_progress: 2,
            waiting: Some(26),
            remaining: 50,
            done: 12,
        });
        // The counts read the ledger, not git: with `git_branch` off they still open line 3,
        // and the worktree label stands alone on line 2 (it follows no git switch).
        assert_eq!(
            plain(&status, &["git_branch"]).lines().nth(1),
            Some("🌳 wt")
        );
        assert_eq!(
            line3(&status, &["git_branch"]),
            "🗂️ 50 ▶️ 2 🔃 26 ✅ 12 | 🔋 CTX 50% | 🪫 Week 75%"
        );
        // Without the ledger notes the four-number form becomes the three-number one.
        status.cards = Some(CardCounts {
            in_progress: 28,
            waiting: None,
            remaining: 50,
            done: 12,
        });
        assert_eq!(
            line3(&status, &["git_branch"]),
            "🗂️ 50 ▶️ 28 ✅ 12 | 🔋 CTX 50% | 🪫 Week 75%"
        );
    }

    #[test]
    fn card_counts_do_not_need_the_repository_or_the_branch() {
        // The last line: line 3 in the full render, and the same line one up when the git
        // line is gone.
        let line3 = |s: &Status, off: &[&str]| plain(s, off).lines().last().unwrap().to_string();
        let mut status = full_status();
        status.cards = Some(CardCounts {
            in_progress: 1,
            waiting: None,
            remaining: 2,
            done: 3,
        });
        status.worktree = "wt".into();
        let usage = "🔋 CTX 50% | 🪫 Week 75%";
        // Without a repository the branch stays on line 2 and the worktree closes it; line 3
        // keeps the counts first and the usage after.
        status.repo = None;
        assert_eq!(
            plain(&status, &[]).lines().nth(1),
            Some("🅱️ main ↑5 ↓2 +3 | 🌳 wt")
        );
        assert_eq!(
            line3(&status, &["git_branch"]),
            format!("🗂️ 2 ▶️ 1 ✅ 3 | {usage}")
        );
        // Without git the worktree label stands alone on line 2.
        status.git = None;
        assert_eq!(plain(&status, &[]).lines().nth(1), Some("🌳 wt"));
        assert_eq!(
            line3(&status, &["git_branch"]),
            format!("🗂️ 2 ▶️ 1 ✅ 3 | {usage}")
        );
        let out = plain(&status, &[]);
        assert_eq!(out.lines().count(), 3, "{out}");
        assert_eq!(out.lines().nth(1), Some("🌳 wt"));
        assert_eq!(
            out.lines().nth(2).unwrap(),
            format!("🗂️ 2 ▶️ 1 ✅ 3 | {usage}")
        );
        status.worktree.clear();
        let out = plain(&status, &[]);
        assert_eq!(out.lines().count(), 2, "{out}");
        assert_eq!(
            out.lines().nth(1).unwrap(),
            format!("🗂️ 2 ▶️ 1 ✅ 3 | {usage}")
        );
    }

    /// Card C-65: color is never the only signal. Each number keeps its icon, the three parts
    /// take three colors of their own, and the text is the plain text once the escapes are
    /// removed.
    #[test]
    fn card_counts_keep_their_icons_under_every_color_profile() {
        let mut status = full_status();
        status.cards = Some(CardCounts {
            in_progress: 5,
            waiting: Some(26),
            remaining: 50,
            done: 12,
        });
        let text = plain(&status, &[]);
        for name in ["mocha", "catppuccin-latte"] {
            let t = theme(name);
            for profile in [Profile::TrueColor, Profile::Ansi256] {
                let out = colored(name, profile, &status);
                assert_eq!(strip_sgr(&out), text, "{name} {profile:?}");
                for (hex, part) in [
                    (t.cards_waiting, "🗂️ 50"),
                    (t.cards_active, "▶️ 5"),
                    (t.cards_posted, "🔃 26"),
                    (t.cards_done, "✅ 12"),
                ] {
                    assert!(out.contains(&profile.paint(hex, part)), "{name} {part}");
                }
            }
            let colors = [
                t.effort,
                t.cards_waiting,
                t.cards_active,
                t.cards_posted,
                t.cards_done,
                t.version,
            ];
            for pair in colors.windows(2) {
                assert_ne!(pair[0], pair[1], "{name}: neighbours share a color");
            }
        }
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

    /// SPEC-20261120-statusline-numbers-reset-worktree (the user's instruction of
    /// 2026-11-20T21:5xZ): the worktree closes line 2 now, as `🌳 WT-c<card>-<name>`, as
    /// `🌳 WT-c<card>` when the label has no name, and as the host's name alone without a
    /// label; line 3 keeps the usage alone and `worktree` off removes the segment.
    #[test]
    fn the_worktree_label_closes_line_2_in_three_shapes() {
        let git = |s: &Status, off: &[&str]| plain(s, off).lines().nth(1).unwrap().to_string();
        let mut status = full_status();
        let bare = git(&status, &[]);
        status.worktree = "host-name".into();
        assert_eq!(git(&status, &[]), format!("{bare} | 🌳 host-name"));
        status.worktree_label = Some(state::WorktreeLabel {
            card: 103,
            name: "statusline-waiting-cards".into(),
        });
        assert_eq!(
            git(&status, &[]),
            format!("{bare} | 🌳 WT-c103-statusline-waiting-cards")
        );
        status.worktree_label = Some(state::WorktreeLabel {
            card: 7,
            name: String::new(),
        });
        assert_eq!(git(&status, &[]), format!("{bare} | 🌳 WT-c7"));
        assert_eq!(git(&status, &["worktree"]), bare);
        // The usage line carries no worktree in any of the shapes.
        status.worktree = "host-name".into();
        status.worktree_label = None;
        let usage_line = plain(&status, &[]).lines().nth(2).unwrap().to_string();
        assert_eq!(usage_line, "🔋 CTX 50% | 🪫 Week 75%");
        assert!(!usage_line.contains('🌳'), "{usage_line}");
        // The main checkout (no host name) has no segment, whatever a label says.
        status.worktree.clear();
        assert_eq!(git(&status, &[]), bare);
    }

    /// #2756 AC-7–AC-9, reworked by the user instruction 2026-11-20 (D-1 of
    /// SPEC-20261120-statusline-layout): the cache segment is not in the layout — the value
    /// and the warm or cold timer show nothing wherever they used to stand — the worktree and
    /// the pull request stay on their lines, and the Claude Code version ends line 3.
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
            lines[0], "🏷️ agent | 🧠 Opus | 🤔 high·t | 🤖 v2.1.293 | 🗿 v2.0 -> 🗿 v3.0",
            "no cache segment: {}",
            lines[0]
        );
        assert_eq!(
            lines[1],
            "📡 owner/repo, ⬇️ 7 🔀 3 | 🅱️ main ↑5 ↓2 +3 | 🔗 #42 pending | 🌳 wt"
        );
        // The card counts would open line 3 (none here) and the usage follows; the worktree
        // label closed line 2 above (the user's instruction of 2026-11-20T21:5xZ); the Claude
        // Code version is on line 1.
        assert_eq!(lines[2], "🔋 CTX 50% | 🪫 Week 75%", "{}", lines[2]);
        assert!(
            !out.contains('\x1b'),
            "plain output carries no link escape: {out:?}"
        );
        // The warm and cold markers never appear, with or without a value, and with the
        // `cache_hit` option left enabled.
        status.cache_timer = Some(CacheTimer::Cold);
        let cold = plain(&status, &[]);
        assert!(!cold.contains('♻'), "{cold}");
        assert!(!cold.contains('⏱') && !cold.contains('🧊'), "{cold}");
        status.cache = None;
        let timer_alone = plain(&status, &[]);
        assert!(
            !timer_alone.contains('♻') && !timer_alone.contains('🧊'),
            "{timer_alone}"
        );
        status.cache_timer = None;
        assert!(
            !plain(&status, &[]).contains('♻'),
            "neither value: no segment"
        );
        status.pr.as_mut().unwrap().review_state.clear();
        assert!(plain(&status, &[]).contains("🔗 #42 | 🌳 wt\n"));
        assert!(!plain(&status, &["pr"]).contains('🔗'));
        assert!(!plain(&status, &["worktree"]).contains('🌳'));
        status.pr = None;
        status.worktree.clear();
        assert_eq!(
            plain(&status, &[]).lines().nth(1),
            Some("📡 owner/repo, ⬇️ 7 🔀 3 | 🅱️ main ↑5 ↓2 +3"),
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
        // The Claude Code version keeps line 3 non-empty without the usage.
        assert_eq!(output.lines().nth(2), Some("🔋 CTX 50%"));
        // #2756: with the repository line hidden, the usage line moves up to second.
        let output = plain(&full_status(), &["git_branch", "git_status"]);
        assert_eq!(output.lines().count(), 2);
        assert!(
            output.lines().next().unwrap().starts_with("🏷️ agent | "),
            "{output}"
        );
        assert!(
            output.lines().nth(1).unwrap().starts_with("🔋 CTX 50%"),
            "{output}"
        );
    }

    #[test]
    fn context_usage_is_a_number_without_any_clear_hint() {
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
        // SPEC-20261120-statusline-numbers-reset-worktree D-3: the soft and hard `/clear`
        // hints are not in the layout — the number alone, in both handoff bands (the stage
        // itself is still judged, in `segments::handoff_stage`).
        assert_eq!(plain(&status, &off), "🪫 CTX 100%");
        assert_eq!(plain(&status, &[]), "🏷️ agent\n🪫 CTX 100%");
        status.memory.tokens_used = 190_000;
        assert_eq!(
            segments::handoff_stage(&status.memory, status.auto_compact),
            segments::HandoffStage::Hard,
            "the hard band is still judged"
        );
        assert_eq!(plain(&status, &off), "🪫 CTX 100%");
        assert_eq!(plain(&status, &["context", "session"]), "MoAI");
        assert_eq!(plain(&status, &["context"]), "🏷️ agent");
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
                    (t.model, "🧠 Opus"),
                    (t.effort, "🤔 high·t"),
                    // D-1: the cache segment is not in the layout, so nothing paints
                    // `theme.cache` any more; the palette entry itself stays distinct (below).
                    (t.version, "🤖 v2.1.293"),
                    (t.version, "🗿 v2.0 -> 🗿 v3.0"),
                    (t.repo, "📡 owner/repo"),
                    (t.issues, "⬇️ 7"),
                    (t.prs, "🔀 3"),
                    (t.branch, "🅱️ main ↑5 ↓2 +3"),
                ] {
                    assert!(
                        out.contains(&profile.paint(hex, text)),
                        "{name} {profile:?}: {text:?} not painted {hex} in {out:?}"
                    );
                }
                // Escapes removed, the text is the NO_COLOR text (AC-2, AC-3 widths).
                assert_eq!(strip_sgr(&out), plain(&status, &[]));

                // SPEC-20261120-statusline-numbers-reset-worktree D-3: no `/clear` hint is
                // painted in either handoff band — `warn` and `alert` stay distinct palette
                // entries (checked below) with nothing painting them.
                status.memory.token_budget = 200_000;
                status.memory.tokens_used = 180_000;
                let soft = colored(name, profile, &status);
                assert!(!soft.contains("⚠️"), "{soft:?}");
                assert!(!soft.contains("/clear"), "{soft:?}");
                status.memory.tokens_used = 190_000;
                let hard = colored(name, profile, &status);
                assert!(!hard.contains('🛑'), "{hard:?}");
                assert!(!hard.contains("/clear"), "{hard:?}");

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
