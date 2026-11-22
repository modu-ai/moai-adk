//! The main Claude Code statusline: up to three lines joined by `\n`, usage
//! windows as numbers, and two Catppuccin themes. Without color the text is identical
//! and carries no escape sequence.

use super::color::Profile;
use super::segments;
use super::state::{DisplayConfig, Flow, Status};
use super::width;
use std::collections::BTreeMap;

const SEPARATOR: &str = " | ";
/// User instruction 2026-11-19, line 2: the two git counts and the text after the repository and
/// the branch icons are one space apart (leader decision, card note 3570, decision D-1 of
/// SPEC-20261120-statusline-line2-order: the two spaces of the user's example are the
/// terminal's drawing of the variation-selector icons). The gap between the two git counts and
/// after the repository and the branch icons is set here; the one space after each other icon
/// sits in its own `format!`. The flow line spaces its icons by [`width::icon_text`]
/// (SPEC-20261122-statusline-flow D-6), and the other lines follow in their own steps.
const GIT_COUNTS_GAP: &str = " ";
const REPO_ICON_GAP: &str = " ";
const BRANCH_ICON_GAP: &str = " ";
const SGR_NORMAL_INTENSITY: &str = "\x1b[22m";
const SGR_RESET: &str = "\x1b[0m";

/// A theme's colors (internal/tui/catppuccin.go).
/// The per-segment colors are the #2559 segment palette (user decision 2026-11-12): each
/// segment kind takes its own Catppuccin accent, so neighbours never share a color.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub struct Theme {
    pub text: &'static str,
    pub muted: &'static str,
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
}

pub const MOCHA: Theme = Theme {
    text: "#CDD6F4",
    muted: "#6C7086",
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
};

pub const LATTE: Theme = Theme {
    text: "#4C4F69",
    muted: "#9CA0B0",
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
    /// The render's clock (the `--now-unix` seam, `Context::now`): the weekly reset
    /// countdown is measured against it, so the goldens stay deterministic.
    now: i64,
    /// The width a line may take (`COLUMNS` − 2 − padding, [`width::budget`]); `None` writes the
    /// widest form of every line.
    budget: Option<usize>,
}

/// One colored piece of a line; the text is plain.
#[derive(Clone)]
struct Piece {
    color: &'static str,
    text: String,
}

/// Pieces joined by one space (`🗂️  D43 R150 W14 E22 ✅ 19` is two pieces).
type Group = Vec<Piece>;

/// The plain text of a form: its groups, each group's pieces joined by a space, the groups
/// joined by the separator. Its display width is what the form picker measures.
fn form_text(form: &[Group]) -> String {
    form.iter()
        .map(|g| {
            g.iter()
                .map(|p| p.text.as_str())
                .collect::<Vec<_>>()
                .join(" ")
        })
        .collect::<Vec<_>>()
        .join(SEPARATOR)
}

impl Renderer {
    pub fn new(config: &DisplayConfig, profile: Option<Profile>) -> Renderer {
        Renderer {
            profile,
            theme: theme(&config.theme),
            segments: config.segments.clone(),
            now: 0,
            budget: None,
        }
    }

    /// The width a line may take; see [`width::budget`].
    pub fn with_budget(mut self, budget: Option<usize>) -> Renderer {
        self.budget = budget;
        self
    }

    /// The render's clock (`Context::now`); the weekly reset countdown measures against it.
    pub fn with_now(mut self, now: i64) -> Renderer {
        self.now = now;
        self
    }

    /// `isSegmentEnabled`: an empty map or a missing key is enabled.
    fn enabled(&self, key: &str) -> bool {
        self.segments.get(key).copied().unwrap_or(true)
    }

    /// `Render`: the layout, or `MoAI` when every line is empty, then the base color.
    /// #2756 (user instruction 2026-11-13): line 1 is the running environment, line 2 the
    /// repository, line 3 the usage windows; #2928 (user instruction 6878415244, 2026-11-14): the
    /// `leader`/`agent` tag opens line 1. The user's three instructions of 2026-11-20
    /// (SPEC-20261120-statusline-layout): line 1 is tag, model, effort, the Claude Code
    /// version and the MoAI version; line 2 is the repository with its forge counts in one
    /// segment and the branch; line 3 opens with the card counts, carries the usage windows, and
    /// closes with the worktree label. The cache segment is not in the layout. A line with
    /// nothing in it is left out, so the numbers above shift up by one for each empty line
    /// before it. SPEC-20261122-statusline-flow: the cards by stage are the flow line between
    /// the repository line and the usage windows; the card counts no longer open line 3.
    pub fn render(&self, status: &Status) -> String {
        let lines: Vec<String> = [
            self.info_line(status),
            self.git_line(status),
            self.flow_line(status),
            self.windows_line(status),
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
        // open line 3 and the worktree label closes it ([`Renderer::windows_line`], the user's
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

    fn windows_line(&self, s: &Status) -> String {
        let mut segs = Vec::new();
        // #2928 (user instruction 6878415244, 2026-11-14): the usage windows have a line of their
        // own; the tag moved to the front of line 1. SPEC-20261122-statusline-flow: the cards by
        // stage have their own line too ([`Renderer::flow_line`]), so this line is the windows
        // only.
        // The user's final call (2026-11-21, SPEC-20261121-statusline-detached): the windows
        // render as numbers only — `🔋 CW 61%`, no graph, no trailing colon — and the
        // `(⚠️ /clear)` / `(🛑/clear!)` hints stay gone (REQ-5, REQ-7). An earlier "keep the
        // graph" reading of a relayed message was a misunderstanding, now corrected. A window
        // whose percentage is unknown says `n/a` with no battery icon (REQ-9).
        if self.enabled("context") {
            if s.memory.available && s.memory.token_budget > 0 {
                let pct = segments::usage_percent(s.memory.tokens_used, s.memory.token_budget);
                segs.push(self.usage_number("CW", pct));
            } else {
                segs.push(self.unknown_number("CW"));
            }
        }
        // #2756 leader decision 6862179411 item 1: the five-hour window is not shown (the user
        // did not choose it, 6858992949); the `usage_5h` option is still accepted. The
        // weekly reset countdown (the user's picked layout) measures against the
        // render's clock (REQ-6).
        if self.enabled("usage_7d") {
            if let Some(w) = s.seven_day {
                let mut seg = match w.used_percentage {
                    Some(pct) => self.usage_number("7D", pct as i64),
                    None => self.unknown_number("7D"),
                };
                if let Some(resets) = w.resets_at {
                    seg.push(' ');
                    seg.push_str(
                        &self.paint(self.theme.muted, &reset_countdown(resets - self.now)),
                    );
                }
                segs.push(seg);
            }
        }
        // The worktree label moved to the end of line 2 ([`Renderer::git_line`], the picked
        // layout of 2026-11-20T21:5x), so line 3 ends with the weekly window.
        segs.join(SEPARATOR)
    }

    /// A window whose numbers are unavailable: `<label> n/a` in the muted color with no
    /// battery icon, never a guessed zero (REQ-9 of SPEC-20261121-statusline-detached).
    fn unknown_number(&self, label: &str) -> String {
        format!("{label} {}", self.unknown_text())
    }

    /// A window renders as a number with the battery icon — no graph, no trailing colon;
    /// the icon switches at high use. The percentage is shown from 0 to 100 (REQ-10).
    fn usage_number(&self, label: &str, pct: i64) -> String {
        let pct = pct.clamp(0, 100);
        let icon = if pct > 70 { "🪫" } else { "🔋" };
        format!("{icon} {label} {pct}%")
    }

    fn git_line(&self, s: &Status) -> String {
        let mut segs = Vec::new();
        // SPEC-20261120-statusline-layout (user instruction 2026-11-20): line 2 opens with the
        // repository and its forge counts in one segment, then the branch; the card counts
        // open line 3 ([`Renderer::windows_line`], the user's third instruction of 2026-11-20).
        if let Some(g) = self
            .enabled("git_branch")
            .then(|| self.repo_groups(s))
            .flatten()
        {
            if let Some(repo) = g.repo_with_forge {
                segs.push(repo);
            }
            // An empty slot (no branch name and no SHA — REQ-8) omits only itself.
            if !g.branch.is_empty() {
                segs.push(g.branch);
            }
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
        // The worktree label closes line 2 (the picked layout of 2026-11-20T21:5x moved it
        // from the end of line 3); the main checkout has none.
        if self.enabled("worktree") && !s.worktree.is_empty() {
            segs.push(self.paint(self.theme.branch, &worktree_text(s)));
        }
        // #2756 leader decision 6862179411 item 2: no `💾` file-change segment (the user's
        // line 2 ends at the branch, whose `+N` carries the change count); the `git_status`
        // option is still accepted.
        segs.join(SEPARATOR)
    }

    /// The flow line (SPEC-20261122-statusline-flow REQ-7 and REQ-9): the cards of the project's
    /// local ledger by the stage of their progress — my card, what waits for me, what is blocked
    /// or on hold, the flow and what closed today — under the one `cards` option. It is written
    /// in the widest of its forms that fits the width budget; empty when there is no ledger to
    /// read or nothing to show. Each number has its letter and each icon its spacing
    /// ([`width::icon_text`]), so color is never the only signal.
    fn flow_line(&self, s: &Status) -> String {
        let Some(flow) = s.flow.filter(|_| self.enabled("cards")) else {
            return String::new();
        };
        let forms = self.flow_forms(&flow);
        let widths: Vec<usize> = forms
            .iter()
            .map(|f| width::display_width(&form_text(f)))
            .collect();
        if forms.is_empty() {
            return String::new();
        }
        let form = &forms[width::pick(&widths, self.budget)];
        form.iter()
            .map(|g| {
                g.iter()
                    .map(|p| self.paint(p.color, &p.text))
                    .collect::<Vec<_>>()
                    .join(" ")
            })
            .collect::<Vec<_>>()
            .join(SEPARATOR)
    }

    /// The forms of the flow line from the widest to the narrowest, never an empty one: the
    /// groups drop from the right — the flow group, then the hold count, then the waits — and a
    /// group whose counts are all zero is left out with its separator. Equal neighbours are
    /// kept once. An unsplit [`Flow`] (the notes could not be read) has one form,
    /// `🗂️  D<n> R<n> W<n>`, never a guess.
    fn flow_forms(&self, f: &Flow) -> Vec<Vec<Group>> {
        let t = &self.theme;
        let piece = |color, icon, text: String| Piece {
            color,
            text: width::icon_text(icon, &text),
        };
        if !f.split {
            let text = format!("D{} R{} W{}", f.draft, f.ready, f.working);
            return vec![vec![vec![piece(t.cards_waiting, "🗂️", text)]]];
        }
        let mine: Option<Group> = f.mine.map(|m| {
            vec![piece(
                t.cards_active,
                "🎯",
                format!("c{} {}", m.card, m.stage.letter()),
            )]
        });
        let waits: Option<Group> = (f.submitted + f.audited + f.landed > 0).then(|| {
            vec![piece(
                t.cards_posted,
                "👀",
                format!("S{} A{} L{}", f.submitted, f.audited, f.landed),
            )]
        });
        let blocked = || piece(t.issues, "⛔", format!("B{}", f.blocked));
        let held = || piece(t.issues, "⏸️", format!("H{}", f.hold));
        let attention: Option<Group> = (f.blocked + f.hold > 0).then(|| vec![blocked(), held()]);
        let attention_no_hold: Option<Group> = (f.blocked > 0).then(|| vec![blocked()]);
        let flow: Option<Group> = (f.draft + f.ready + f.working + f.stale + f.done_today > 0)
            .then(|| {
                vec![
                    piece(
                        t.cards_waiting,
                        "🗂️",
                        format!("D{} R{} W{} E{}", f.draft, f.ready, f.working, f.stale),
                    ),
                    piece(t.cards_done, "✅", f.done_today.to_string()),
                ]
            });
        let levels: [[Option<Group>; 4]; 4] = [
            [mine.clone(), waits.clone(), attention.clone(), flow],
            [mine.clone(), waits.clone(), attention, None],
            [mine.clone(), waits, attention_no_hold.clone(), None],
            [mine, None, attention_no_hold, None],
        ];
        let mut forms: Vec<Vec<Group>> = Vec::new();
        for level in levels {
            let form: Vec<Group> = level.into_iter().flatten().collect();
            if form.is_empty() {
                continue;
            }
            if forms
                .last()
                .is_some_and(|last| form_text(last) == form_text(&form))
            {
                continue;
            }
            forms.push(form);
        }
        forms
    }

    /// The three groups that retain the `git_branch` switch: the forge counts and the
    /// repository and the branch slot; `None` without a git repository. Their order on the
    /// line is [`Renderer::git_line`]'s. SPEC-20261121-statusline-detached (the user-confirmed
    /// decision on card note 5373): a detached HEAD shows the short SHA in the branch slot
    /// instead of hiding the whole line, and a slot whose lookup failed says `n/a` — never
    /// a zero and never nothing.
    fn repo_groups(&self, s: &Status) -> Option<RepoGroups> {
        let git = s.git.as_ref()?;
        let branch_slot = if !git.branch.is_empty() {
            format!("🅱️{BRANCH_ICON_GAP}{}", git.branch)
        } else if !git.head_sha.is_empty() {
            format!("🅱️{BRANCH_ICON_GAP}{}", git.head_sha)
        } else {
            String::new()
        };
        let mut branch_text = branch_slot;
        if !branch_text.is_empty() {
            // Under M-Flow the merges land on the local integration branch first and the
            // remote moves at a release, so the ahead count is the number that shows them
            // landing (SPEC-20261120-statusline-ahead); the order is `↑N ↓N +N`, each hidden
            // at 0. A detached HEAD has none of these.
            if git.ahead > 0 {
                branch_text.push_str(&format!(" ↑{}", git.ahead));
            }
            if git.behind > 0 {
                branch_text.push_str(&format!(" ↓{}", git.behind));
            }
            let dirty = git.modified + git.staged + git.untracked;
            if dirty > 0 {
                branch_text.push_str(&format!(" +{dirty}"));
            }
        }
        let branch = self.paint(self.theme.branch, &branch_text);
        let repo_text = match s
            .repo
            .as_ref()
            .filter(|(o, n)| !o.is_empty() && !n.is_empty())
        {
            Some((owner, name)) => format!("📡{REPO_ICON_GAP}{owner}/{name}"),
            None => format!("📡{REPO_ICON_GAP}{}", self.unknown_text()),
        };
        let forge = if !self.enabled("github") || s.github.suppressed {
            // A user switch keeps the counts hidden; a failed query does not (REQ-3).
            None
        } else if s.github.available {
            // #2559 (user instructions 2026-11-12): icons instead of letter labels; the Issue
            // icon is `⬇️` (U+2B07 U+FE0F) since 23:1xZ, the PR icon stays `🔀`.
            Some(format!(
                "{}{GIT_COUNTS_GAP}{}",
                self.paint(self.theme.issues, &format!("⬇️ {}", s.github.open_issues)),
                self.paint(self.theme.prs, &format!("🔀 {}", s.github.open_prs))
            ))
        } else {
            Some(format!(
                "{}{GIT_COUNTS_GAP}{}",
                self.paint(self.theme.muted, "⬇️ n/a"),
                self.paint(self.theme.muted, "🔀 n/a")
            ))
        };
        let repo = self.paint(self.theme.repo, &repo_text);
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

    /// `n/a` in the muted color: a slot whose lookup failed says so, in place of a zero.
    fn unknown_text(&self) -> String {
        self.paint(self.theme.muted, "n/a")
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
/// `(resets in …)` from the seconds left (the user's picked layout): a day or
/// more counts days and hours, under a day hours and minutes, under an hour minutes only.
fn reset_countdown(secs_left: i64) -> String {
    if secs_left >= 86_400 {
        format!(
            "(resets in {}d {}h)",
            secs_left / 86_400,
            (secs_left % 86_400) / 3_600
        )
    } else if secs_left >= 3_600 {
        format!(
            "(resets in {}h {}m)",
            secs_left / 3_600,
            (secs_left % 3_600) / 60
        )
    } else {
        format!("(resets in {}m)", (secs_left / 60).max(0))
    }
}

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
    use crate::statusline::state::{
        self, CacheTimer, Context, Flow, GitHubCounts, MyCard, Stage, Version,
    };

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
    /// and line 3 is the usage windows ending with the Claude Code version, each line joined by
    /// ` | `. The cache segment is not in the layout (D-1 of SPEC-20261120-statusline-layout).
    /// Under M-Flow the merges land locally first, so the ahead count shows before behind
    /// (SPEC-20261120-statusline-ahead).
    #[test]
    fn full_main_layout_has_three_ordered_lines_with_ahead_before_behind() {
        assert_eq!(
            plain(&full_status(), &[]),
            "🏷️ agent | 🧠 Opus | 🤔 high·t | 🤖 v2.1.293 | 🗿 v2.0 -> 🗿 v3.0\n\
             📡 owner/repo, ⬇️ 7 🔀 3 | 🅱️ main ↑5 ↓2 +3\n\
             🔋 CW 50% | 🪫 7D 75%"
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

    /// The user's picked example (2026-11-20T21:5xZ): every segment present gives exactly these
    /// lines in plain output — since SPEC-20261122-statusline-flow the flow line stands between
    /// the repository line and the usage windows. The cache data is in the input and, by D-1,
    /// shows nothing; the worktree label closes line 2; the weekly window carries its reset
    /// countdown measured from the render's clock.
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
                resets_at: Some(EXAMPLE_NOW + 6 * 86_400 + 21 * 3_600),
            }),
            repo: Some(("modu-ai".into(), "moai-adk".into())),
            github: GitHubCounts {
                available: true,
                open_issues: 0,
                open_prs: 0,
                ..Default::default()
            },
            git: Some(segments::parse_porcelain(
                "## v2026.11.0...origin/v2026.11.0 [ahead 207]\n M a\n M b\n M c\n M d\n M e\nA  f\nA  g\nA  h\n?? i\n?? j\n?? k\n",
            )),
            flow: Some(example_flow()),
            worktree: "c135".into(),
            ..Default::default()
        }
    }

    /// The flow of the example: one card of mine, and every group of the flow line present.
    fn example_flow() -> Flow {
        Flow {
            draft: 43,
            ready: 150,
            working: 14,
            stale: 22,
            submitted: 11,
            audited: 2,
            landed: 0,
            blocked: 3,
            hold: 7,
            done_today: 19,
            mine: Some(MyCard {
                card: 12,
                stage: Stage::Working,
            }),
            split: true,
        }
    }

    /// The widest form of the flow line of [`example_flow`].
    const P0: &str = "🎯 c12 W | 👀 S11 A2 L0 | ⛔ B3 ⏸️  H7 | 🗂️  D43 R150 W14 E22 ✅ 19";

    /// The render clock the example is measured at (`--now-unix` seam).
    const EXAMPLE_NOW: i64 = 1_700_000_000;

    #[test]
    fn the_user_example_gives_exactly_these_lines() {
        let out = Renderer::new(&DisplayConfig::default(), None)
            .with_now(EXAMPLE_NOW)
            .render(&example_status());
        assert_eq!(
            out,
            format!(
                "🏷️ Leader | 🧠 Opus 5.5 | 🤔 high·t | 🤖 v2.1.325 | 🗿 v4.2647.10\n\
                 📡 modu-ai/moai-adk, ⬇️ 0 🔀 0 | 🅱️ v2026.11.0 ↑207 +11 | 🌳 c135\n\
                 {P0}\n\
                 🪫 CW 97% | 🔋 7D 21% (resets in 6d 21h)"
            )
        );
    }

    /// REQ-1 and REQ-8: a detached HEAD shows the short SHA in the branch slot while the
    /// repository and its forge counts stay; without a SHA only the branch slot is omitted.
    #[test]
    fn detached_head_shows_the_short_sha_and_keeps_the_repository() {
        let mut status = full_status();
        status.git = Some(segments::GitStatus {
            branch: String::new(),
            head_sha: "a1b2c3d".into(),
            ..Default::default()
        });
        let line = plain(&status, &[]).lines().nth(1).unwrap().to_string();
        assert_eq!(line, "📡 owner/repo, ⬇️ 7 🔀 3 | 🅱️ a1b2c3d", "{line}");
        assert!(!line.contains("↑"), "a detached head has no ahead: {line}");

        status.git = Some(segments::GitStatus::default());
        let line = plain(&status, &[]).lines().nth(1).unwrap().to_string();
        assert_eq!(
            line, "📡 owner/repo, ⬇️ 7 🔀 3",
            "no branch and no SHA leaves the repository line"
        );
    }

    /// REQ-2 and REQ-3: a repository whose name or remote cannot be resolved and forge
    /// counts that cannot be read say `n/a` — never a zero and never a hidden segment.
    #[test]
    fn lookup_failures_show_unknown_not_zero() {
        let mut status = full_status();
        status.repo = None;
        status.github.available = false;
        let line = plain(&status, &[]).lines().nth(1).unwrap().to_string();
        assert_eq!(line, "📡 n/a, ⬇️ n/a 🔀 n/a | 🅱️ main ↑5 ↓2 +3", "{line}");
        assert!(
            !line.contains("⬇️ 0"),
            "a failed lookup is never zero: {line}"
        );
    }

    /// REQ-6: the countdown picks its unit by the seconds left, measured from the render's
    /// clock, and a window without a reset time ends at the percentage.
    #[test]
    fn the_reset_countdown_uses_the_three_formats() {
        let at = |left: i64| {
            let status = Status {
                seven_day: Some(RateWindow {
                    used_percentage: Some(21.0),
                    resets_at: Some(EXAMPLE_NOW + left),
                }),
                ..Status::default()
            };
            Renderer::new(&DisplayConfig::default(), None)
                .with_now(EXAMPLE_NOW)
                .render(&status)
                .lines()
                .last()
                .unwrap()
                .to_string()
        };
        assert!(
            at(6 * 86_400 + 21 * 3_600).ends_with("(resets in 6d 21h)"),
            "{}",
            at(6 * 86_400 + 21 * 3_600)
        );
        assert!(at(5 * 3_600 + 30 * 60).ends_with("(resets in 5h 30m)"));
        assert!(at(40 * 60).ends_with("(resets in 40m)"));
        assert!(
            at(0).ends_with("(resets in 0m)"),
            "an immediate reset still names itself: {}",
            at(0)
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

    /// #2756 leader decision 6862179411 items 1–2: no five-hour window (the user did not choose
    /// it) and no `💾` file-change segment, even with both in the input and their
    /// `usage_5h` / `git_status` options left enabled. The branch keeps its `+N` changes.
    #[test]
    fn the_five_hour_window_and_the_file_change_segment_are_never_shown() {
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
    fn rate_windows_are_optional_and_an_unknown_window_says_n_a() {
        let mut status = Status::default();
        assert_eq!(plain(&status, &["session"]), "CW n/a");
        status.seven_day = Some(RateWindow::default());
        assert_eq!(plain(&status, &["session"]), "CW n/a | 7D n/a");
        assert_eq!(plain(&status, &["session", "context"]), "7D n/a");
        assert_eq!(plain(&status, &["session", "usage_7d"]), "CW n/a");
        // A window that reports a real zero keeps it, with its icon.
        status.seven_day = Some(RateWindow {
            used_percentage: Some(0.0),
            resets_at: None,
        });
        assert_eq!(plain(&status, &["session", "context"]), "🔋 7D 0%");
    }

    /// REQ-9: a window whose percentage is unavailable says `n/a` with no battery icon and never
    /// `0%`; the reset countdown of an unknown week still follows it.
    #[test]
    fn an_unknown_window_says_n_a_without_a_battery_icon() {
        let unknown_ctx = Status::default();
        let line = plain(&unknown_ctx, &["session"]);
        assert_eq!(line, "CW n/a", "{line}");
        assert!(!line.contains('🪫') && !line.contains('🔋'), "{line}");
        let no_budget = Status {
            memory: segments::Memory {
                tokens_used: 5,
                token_budget: 0,
                context_window_size: 200_000,
                available: true,
            },
            ..Status::default()
        };
        assert_eq!(plain(&no_budget, &["session"]), "CW n/a");
        let week = Status {
            seven_day: Some(RateWindow {
                used_percentage: None,
                resets_at: None,
            }),
            ..Status::default()
        };
        let line = plain(&week, &["session", "context"]);
        assert_eq!(line, "7D n/a", "{line}");
        assert!(
            !line.contains("0%") && !line.contains('🪫') && !line.contains('🔋'),
            "{line}"
        );
    }

    /// REQ-10: a window's percentage is shown from 0 to 100 and no further, at the edges too.
    #[test]
    fn window_percentages_stay_between_zero_and_one_hundred() {
        let week = |pct: f64| {
            let status = Status {
                seven_day: Some(RateWindow {
                    used_percentage: Some(pct),
                    resets_at: None,
                }),
                ..Status::default()
            };
            plain(&status, &["context", "session"])
        };
        assert_eq!(week(-5.0), "🔋 7D 0%");
        assert_eq!(week(0.0), "🔋 7D 0%");
        assert_eq!(week(1.0), "🔋 7D 1%");
        assert_eq!(week(99.0), "🪫 7D 99%");
        assert_eq!(week(100.0), "🪫 7D 100%");
        assert_eq!(week(101.0), "🪫 7D 100%");
        assert_eq!(week(1_000_000.0), "🪫 7D 100%");
        let ctx = |used: i64| {
            let status = Status {
                memory: segments::Memory {
                    tokens_used: used,
                    token_budget: 100,
                    context_window_size: 200_000,
                    available: true,
                },
                ..Status::default()
            };
            plain(&status, &["session", "usage_7d"])
        };
        assert_eq!(ctx(0), "🔋 CW 0%");
        assert_eq!(ctx(100), "🪫 CW 100%");
        assert_eq!(ctx(101), "🪫 CW 100%");
        assert_eq!(ctx(5_000), "🪫 CW 100%");
    }

    /// REQ-11: the screen is English only — nothing a render prints holds a Hangul letter, with
    /// the values known, unknown or failed to read, in plain and in colored output.
    #[test]
    fn no_line_of_the_screen_carries_korean_text() {
        let is_hangul = |c: char| matches!(c, '\u{1100}'..='\u{11FF}' | '\u{3130}'..='\u{318F}' | '\u{AC00}'..='\u{D7A3}');
        let mut failed = full_status();
        failed.repo = None;
        failed.github.available = false;
        failed.seven_day = Some(RateWindow::default());
        failed.memory.available = false;
        let mut detached = full_status();
        detached.git = Some(segments::parse_porcelain("## HEAD (no branch)\n"));
        let screens = [Status::default(), full_status(), failed, detached];
        for status in &screens {
            let out = plain(status, &[]);
            assert!(!out.chars().any(is_hangul), "plain: {out}");
            let out = colored("mocha", Profile::TrueColor, status);
            assert!(!out.chars().any(is_hangul), "colored: {out:?}");
        }
    }

    /// #2559 AC-1: the window icon's thresholds pinned at N/N+1 — the battery icon turns at
    /// 70/71. Plain output, so only the threshold decides.
    #[test]
    fn window_icon_thresholds_hold_at_their_boundaries() {
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
        assert_eq!(at(70.0), "🔋 7D 70%");
        assert_eq!(at(71.0), "🪫 7D 71%");
        assert_eq!(at(9.0), "🔋 7D 9%");
        assert_eq!(at(10.0), "🔋 7D 10%");
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
        assert_eq!(out.lines().nth(2), Some("🔋 CW 50% | 🪫 7D 75%"));
    }

    /// #2928 (user instruction 6878415244, 2026-11-14): line 1 is `🏷️ <tag> | 🧠 … | 🗿 …` and
    /// line 3 is `🔋 CW … | 🪫 7D …` alone. The tag (`Leader`, `Agent-<n>` or `agent`, #2871) comes first; it
    /// is shown whenever the `session` segment is on, even with no stdin (decision Q2 = (가));
    /// with every other line-1 segment off it stands alone, and with `session` off line 1 is as
    /// before. The window icons and marks are unchanged.
    #[test]
    fn the_tag_is_the_first_segment_of_line_one() {
        let mut status = Status::default();
        assert_eq!(plain(&status, &[]), "🏷️ agent\nCW n/a");
        status.seven_day = Some(RateWindow::default());
        assert_eq!(plain(&status, &[]), "🏷️ agent\nCW n/a | 7D n/a");
        status.leader = true;
        assert_eq!(plain(&status, &[]), "🏷️ Leader\nCW n/a | 7D n/a");
        assert_eq!(plain(&status, &["usage_7d", "context"]), "🏷️ Leader");
        assert_eq!(plain(&status, &["session"]), "CW n/a | 7D n/a");
        let mut full = full_status();
        let windows = "🔋 CW 50% | 🪫 7D 75%";
        let env = "🧠 Opus | 🤔 high·t | 🤖 v2.1.293 | 🗿 v2.0 -> 🗿 v3.0";
        assert_eq!(
            plain(&full, &[]).lines().next(),
            Some(format!("🏷️ agent | {env}").as_str())
        );
        assert_eq!(plain(&full, &[]).lines().nth(2), Some(windows));
        assert_eq!(plain(&full, &["session"]).lines().next(), Some(env));
        assert_eq!(plain(&full, &["session"]).lines().nth(2), Some(windows));
        let line_one_off = [
            "model",
            "effort_thinking",
            "cache_hit",
            "claude_version",
            "moai_version",
        ];
        let only_tag = plain(&full, &line_one_off);
        assert_eq!(only_tag.lines().next(), Some("🏷️ agent"), "{only_tag}");
        // `claude_version` is off in this list, so line 3 ends with the windows.
        assert_eq!(
            only_tag.lines().nth(2),
            Some("🔋 CW 50% | 🪫 7D 75%"),
            "{only_tag}"
        );
        full.leader = true;
        assert_eq!(
            plain(&full, &[]).lines().next(),
            Some(format!("🏷️ Leader | {env}").as_str())
        );
        assert_eq!(plain(&full, &[]).lines().nth(2), Some(windows));
        let colored = strip_sgr(&colored("mocha", Profile::TrueColor, &full));
        assert!(
            colored.lines().next().unwrap().starts_with("🏷️ Leader | "),
            "{colored:?}"
        );
        assert_eq!(colored.lines().nth(2), Some(windows), "{colored:?}");
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
             🔋 CW 50% | 🪫 7D 75%"
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
        // A failed query says `n/a`, never a zero and never a hidden segment (REQ-3).
        status.github.available = false;
        let unknown = plain(&status, &[]);
        assert!(
            unknown.contains("📡 owner/repo, ⬇️ n/a 🔀 n/a | 🅱️ main ↑5 ↓2 +3"),
            "{unknown}"
        );
        assert!(!unknown.contains("⬇️ 0"), "{unknown}");
        // A user switch keeps the counts hidden.
        status.github.available = true;
        status.github.suppressed = true;
        let suppressed = plain(&status, &[]);
        assert!(
            suppressed.contains("📡 owner/repo | 🅱️ main ↑5 ↓2 +3"),
            "{suppressed}"
        );
        status.github.suppressed = false;
        assert!(plain(&status, &["github"]).contains("📡 owner/repo | 🅱️ main ↑5 ↓2 +3"));
        assert!(!plain(&status, &["git_branch"]).contains("📡"));
        assert!(!plain(&status, &["git_status"]).contains("💾"));
        // The repository name unknown: the slot says so, the line stays (REQ-2).
        status.repo = None;
        assert!(plain(&status, &[]).contains("📡 n/a, ⬇️ 0 🔀 0"));
    }

    /// SPEC-20261122-statusline-flow REQ-7 and REQ-8: the cards by stage have a line of their own
    /// between the repository line and the usage windows. Lines 1 and 2 and the windows do not
    /// change, the `cards` option hides the flow line and nothing else, and it does not depend
    /// on git: forge counts, repository and branch change line 2, never the flow line.
    #[test]
    fn the_flow_line_stands_between_the_repository_line_and_the_usage_windows() {
        let mut status = full_status();
        let without = plain(&status, &[]);
        assert_eq!(without.lines().count(), 3, "no ledger: no flow line");
        assert_eq!(
            without.lines().last().unwrap(),
            "🔋 CW 50% | 🪫 7D 75%",
            "no flow: the windows alone"
        );
        status.flow = Some(example_flow());
        let with = plain(&status, &[]);
        let lines: Vec<&str> = with.lines().collect();
        assert_eq!(lines.len(), 4, "{with}");
        assert_eq!(Some(lines[0]), without.lines().next());
        assert_eq!(Some(lines[1]), without.lines().nth(1));
        assert_eq!(lines[2], P0);
        assert_eq!(Some(lines[3]), without.lines().nth(2));
        // The `cards` option hides the flow line and nothing else.
        assert_eq!(plain(&status, &["cards"]), without);
        // The flow comes from the ledger, not from git.
        let flow_line =
            |s: &Status, off: &[&str]| plain(s, off).lines().nth(2).unwrap().to_string();
        assert_eq!(flow_line(&status, &["github"]), P0);
        status.github.suppressed = true;
        assert_eq!(flow_line(&status, &[]), P0);
        status.repo = None;
        status.git = None;
        let bare = plain(&status, &["git_branch"]);
        assert_eq!(
            bare.lines().nth(1),
            Some(P0),
            "line 2 is gone, the flow moves up"
        );
        // A ledger without a card has nothing to show: the line is left out, not zeros.
        status.flow = Some(Flow {
            split: true,
            ..Default::default()
        });
        // (Line 2 is gone here too: the model line and the windows are what is left.)
        assert_eq!(plain(&status, &[]).lines().count(), 2);
    }

    /// SPEC-20261120-statusline-layout (user instruction 2026-11-20): line 2 opens with one
    /// segment `📡 owner/repo, ⬇️ n 🔀 n` — the comma right after the repository name, the
    /// forge counts in the same segment — then the branch, then the pull request, groups
    /// joined by ` | `, and no group has a slash in it. The flow line stands under it
    /// (SPEC-20261122-statusline-flow) and the windows under that.
    #[test]
    fn line_two_reads_the_repository_with_the_forge_counts_then_the_branch_then_the_pr() {
        let line = |s: &Status, off: &[&str]| plain(s, off).lines().nth(1).unwrap().to_string();
        let mut status = full_status();
        status.flow = Some(example_flow());
        status.pr = Some(state::PrInfo {
            number: 42,
            url: String::new(),
            review_state: "pending".into(),
            merge_request: false,
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
        // The flow line is line 3 and the windows alone are line 4.
        assert_eq!(plain(&status, &[]).lines().nth(2).unwrap(), P0);
        assert_eq!(
            plain(&status, &[]).lines().nth(3).unwrap(),
            "🔋 CW 50% | 🪫 7D 75%"
        );
        // Without the forge counts the repository name stands alone, with no comma.
        assert_eq!(
            line(&status, &["github"]),
            "📡 owner/repo | 🅱️ main ↑5 ↓2 +3 | 🔗 #42 pending"
        );
        // The `cards` switch hides the flow line and nothing else.
        assert_eq!(
            plain(&status, &["cards"]).lines().nth(2).unwrap(),
            "🔋 CW 50% | 🪫 7D 75%"
        );
        // The pull request keeps the last place of line 2.
        assert_eq!(
            line(&status, &["pr"]),
            "📡 owner/repo, ⬇️ 7 🔀 3 | 🅱️ main ↑5 ↓2 +3"
        );
    }

    /// The flow comes from the local ledger, not from the repository, so the flow line is shown
    /// without a repository name, without a branch and with `git_branch` off; the worktree
    /// label stands alone on line 2 (REQ-4 of the layout SPEC) and the windows close the output.
    #[test]
    fn the_flow_line_does_not_need_the_repository_or_the_branch() {
        let mut status = full_status();
        status.flow = Some(example_flow());
        status.worktree = "wt".into();
        let windows = "🔋 CW 50% | 🪫 7D 75%";
        status.repo = None;
        let off = plain(&status, &["git_branch"]);
        assert_eq!(off.lines().nth(1).unwrap(), "🌳 wt");
        assert_eq!(off.lines().nth(2).unwrap(), P0);
        assert_eq!(off.lines().nth(3).unwrap(), windows);
        // The full render: an unresolved repository says `n/a` and keeps its line (REQ-2), the
        // worktree label closes line 2, the flow line follows and the windows are last.
        let out = plain(&status, &[]);
        assert_eq!(out.lines().count(), 4, "{out}");
        assert_eq!(
            out.lines().nth(1).unwrap(),
            "📡 n/a, ⬇️ 7 🔀 3 | 🅱️ main ↑5 ↓2 +3 | 🌳 wt"
        );
        assert_eq!(out.lines().nth(2).unwrap(), P0);
        // No git and no worktree: line 2 has nothing to carry and is left out whole.
        status.git = None;
        status.worktree.clear();
        let out = plain(&status, &[]);
        assert_eq!(out.lines().count(), 3, "{out}");
        assert_eq!(out.lines().nth(1).unwrap(), P0);
    }

    /// Color is never the only signal. Every number keeps its letter and every icon its text,
    /// the groups take colors of their own, and the text is the plain text once the escapes are
    /// removed.
    #[test]
    fn the_flow_line_keeps_its_letters_under_every_color_profile() {
        let mut status = full_status();
        status.flow = Some(example_flow());
        let text = plain(&status, &[]);
        for name in ["mocha", "catppuccin-latte"] {
            let t = theme(name);
            for profile in [Profile::TrueColor, Profile::Ansi256] {
                let out = colored(name, profile, &status);
                assert_eq!(strip_sgr(&out), text, "{name} {profile:?}");
                for (hex, part) in [
                    (t.cards_active, "🎯 c12 W"),
                    (t.cards_posted, "👀 S11 A2 L0"),
                    (t.issues, "⛔ B3"),
                    (t.issues, "⏸️  H7"),
                    (t.cards_waiting, "🗂️  D43 R150 W14 E22"),
                    (t.cards_done, "✅ 19"),
                ] {
                    assert!(out.contains(&profile.paint(hex, part)), "{name} {part}");
                }
            }
            let colors = [
                t.cards_active,
                t.cards_posted,
                t.issues,
                t.cards_waiting,
                t.cards_done,
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

    /// The user's third instruction of 2026-11-20: the worktree closes line 3, as
    /// `🌳 WT-c<card>-<name>`, as `🌳 WT-c<card>` when the label has no name, and as the
    /// host's name alone without a label; the windows before it are unchanged and `worktree`
    /// off removes it.
    #[test]
    fn the_worktree_label_closes_line_2_in_three_shapes() {
        let line2 = |s: &Status, off: &[&str]| plain(s, off).lines().nth(1).unwrap().to_string();
        let mut status = full_status();
        let base = line2(&status, &[]);
        status.worktree = "host-name".into();
        assert_eq!(line2(&status, &[]), format!("{base} | 🌳 host-name"));
        status.worktree_label = Some(state::WorktreeLabel {
            card: 103,
            name: "statusline-waiting-cards".into(),
        });
        assert_eq!(
            line2(&status, &[]),
            format!("{base} | 🌳 WT-c103-statusline-waiting-cards")
        );
        status.worktree_label = Some(state::WorktreeLabel {
            card: 7,
            name: String::new(),
        });
        assert_eq!(line2(&status, &[]), format!("{base} | 🌳 WT-c7"));
        assert_eq!(line2(&status, &["worktree"]), base);
        // The main checkout (no host name) has no segment, whatever a label says.
        status.worktree.clear();
        assert_eq!(line2(&status, &[]), base);
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
            merge_request: false,
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
        // The card counts would open line 3 (none here) and the windows follow as numbers;
        // the worktree label closed this line before the picked layout moved it to line 2
        // (REQ-4).
        assert!(lines[2].starts_with("🔋 CW 50%"), "{}", lines[2]);
        assert!(lines[2].ends_with("75%"), "{}", lines[2]);
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
        // The worktree label sits after the PR link on line 2; without it the PR closes the
        // line again.
        status.worktree.clear();
        assert!(plain(&status, &[]).contains("| 🔗 #42\n"));
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
            merge_request: false,
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
        // The Claude Code version keeps line 3 non-empty without the windows.
        assert_eq!(output.lines().nth(2), Some("🔋 CW 50%"));
        // #2756: with the repository line hidden, the windows line moves up to second.
        let output = plain(&full_status(), &["git_branch", "git_status"]);
        assert_eq!(output.lines().count(), 2);
        assert!(
            output.lines().next().unwrap().starts_with("🏷️ agent | "),
            "{output}"
        );
        assert!(
            output.lines().nth(1).unwrap().starts_with("🔋 CW "),
            "{output}"
        );
    }

    #[test]
    fn context_renders_without_the_handoff_hints() {
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
        // REQ-7: the `(⚠️ /clear)` and `(🛑/clear!)` hints are gone at every band — the window
        // and the percentage stay exactly as they were.
        assert_eq!(plain(&status, &off), "🪫 CW 100%");
        assert_eq!(plain(&status, &[]), "🏷️ agent\n🪫 CW 100%");
        status.memory.tokens_used = 190_000;
        assert_eq!(plain(&status, &off), "🪫 CW 100%");
        status.goal_armed = true;
        assert_eq!(plain(&status, &off), "🪫 CW 100%");
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

    fn segment_colors(t: &Theme) -> [&'static str; 8] {
        [
            t.model, t.effort, t.cache, t.version, t.repo, t.issues, t.prs, t.branch,
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

                status.memory.token_budget = 200_000;
                status.memory.tokens_used = 180_000;
                let soft = colored(name, profile, &status);
                assert!(!soft.contains("(⚠️ /clear)"), "{soft:?}");
                status.memory.tokens_used = 190_000;
                let hard = colored(name, profile, &status);
                assert!(!hard.contains("(🛑/clear!)"), "{hard:?}");

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

/// AC-5, AC-7 (the flow line's share) and AC-8 (icon spacing) of SPEC-20261122-statusline-flow:
/// the forms of the flow line, from the widest to the narrowest.
#[cfg(test)]
mod flow_line_tests {
    use super::*;
    use crate::statusline::state::{MyCard, Stage};

    fn full() -> Flow {
        Flow {
            draft: 43,
            ready: 150,
            working: 14,
            stale: 22,
            submitted: 11,
            audited: 2,
            landed: 0,
            blocked: 3,
            hold: 7,
            done_today: 19,
            mine: Some(MyCard {
                card: 12,
                stage: Stage::Working,
            }),
            split: true,
        }
    }

    fn no_card() -> Flow {
        Flow {
            mine: None,
            ..full()
        }
    }

    /// The flow line alone: the session tag and the context window are switched off, and every
    /// other line has no input, so only the flow line is left. `MoAI` is the whole output when
    /// the flow line is empty too.
    fn only_flow(flow: Flow, budget: Option<usize>, profile: Option<Profile>) -> String {
        let config = DisplayConfig {
            segments: [("session", false), ("context", false)]
                .into_iter()
                .map(|(k, v)| (k.to_string(), v))
                .collect(),
            ..Default::default()
        };
        Renderer::new(&config, profile)
            .with_budget(budget)
            .render(&Status {
                flow: Some(flow),
                ..Default::default()
            })
    }

    fn plain(flow: Flow, budget: Option<usize>) -> String {
        only_flow(flow, budget, None)
    }

    const P: [(&str, usize); 4] = [
        (
            "🎯 c12 W | 👀 S11 A2 L0 | ⛔ B3 ⏸️  H7 | 🗂️  D43 R150 W14 E22 ✅ 19",
            67,
        ),
        ("🎯 c12 W | 👀 S11 A2 L0 | ⛔ B3 ⏸️  H7", 38),
        ("🎯 c12 W | 👀 S11 A2 L0 | ⛔ B3", 31),
        ("🎯 c12 W | ⛔ B3", 16),
    ];
    const Q: [(&str, usize); 4] = [
        (
            "👀 S11 A2 L0 | ⛔ B3 ⏸️  H7 | 🗂️  D43 R150 W14 E22 ✅ 19",
            56,
        ),
        ("👀 S11 A2 L0 | ⛔ B3 ⏸️  H7", 27),
        ("👀 S11 A2 L0 | ⛔ B3", 20),
        ("⛔ B3", 5),
    ];

    /// Every form is exactly the string of the Layouts section, with the width the section
    /// gives, and changes at its boundary: a budget equal to the width keeps the form, one less
    /// drops to the next.
    #[test]
    fn every_form_is_exactly_the_string_of_the_layouts_section() {
        for (flow, forms) in [(full(), P), (no_card(), Q)] {
            for (i, (text, cols)) in forms.iter().enumerate() {
                assert_eq!(width::display_width(text), *cols, "{text}");
                assert_eq!(&plain(flow, Some(*cols)), text, "budget {cols}");
                if i == 0 {
                    assert_eq!(&plain(flow, Some(*cols + 40)), text, "{cols} + 40");
                }
                if let Some((next, _)) = forms.get(i + 1) {
                    assert_eq!(&plain(flow, Some(*cols - 1)), next, "budget {}", cols - 1);
                }
            }
        }
    }

    #[test]
    fn no_budget_writes_the_widest_form_and_a_budget_below_the_narrowest_the_narrowest() {
        assert_eq!(plain(full(), None), P[0].0);
        assert_eq!(plain(no_card(), None), Q[0].0);
        for budget in [15, 5, 0] {
            assert_eq!(plain(full(), Some(budget)), P[3].0, "{budget}");
        }
        assert_eq!(plain(no_card(), Some(0)), Q[3].0);
    }

    /// A group whose counts are all zero is left out with its separator; a line with nothing
    /// to show is left out whole.
    #[test]
    fn groups_of_zero_counts_are_left_out_with_their_separator() {
        let flow = Flow {
            submitted: 0,
            audited: 0,
            landed: 0,
            ..full()
        };
        assert_eq!(
            plain(flow, None),
            "🎯 c12 W | ⛔ B3 ⏸️  H7 | 🗂️  D43 R150 W14 E22 ✅ 19"
        );
        let flow = Flow {
            blocked: 0,
            hold: 0,
            ..full()
        };
        assert_eq!(
            plain(flow, None),
            "🎯 c12 W | 👀 S11 A2 L0 | 🗂️  D43 R150 W14 E22 ✅ 19"
        );
        // Hold alone shows `B0` beside it, and drops with the hold count when narrow.
        let hold_only = Flow {
            blocked: 0,
            ..full()
        };
        assert_eq!(
            plain(hold_only, Some(38)),
            "🎯 c12 W | 👀 S11 A2 L0 | ⛔ B0 ⏸️  H7"
        );
        assert_eq!(plain(hold_only, Some(37)), "🎯 c12 W | 👀 S11 A2 L0");
        let flow_group = Flow {
            draft: 0,
            ready: 0,
            working: 0,
            stale: 0,
            done_today: 0,
            ..full()
        };
        assert_eq!(
            plain(flow_group, None),
            "🎯 c12 W | 👀 S11 A2 L0 | ⛔ B3 ⏸️  H7"
        );
        // Only my card, and only a stale one.
        let mine_only = Flow {
            mine: Some(MyCard {
                card: 7,
                stage: Stage::Stale,
            }),
            ..Flow {
                split: true,
                ..Default::default()
            }
        };
        assert_eq!(plain(mine_only, None), "🎯 c7 E");
        // Nothing at all: no flow line, so the renderer's own fallback is the whole output.
        let nothing = Flow {
            split: true,
            ..Default::default()
        };
        assert_eq!(plain(nothing, None), "MoAI");
    }

    /// My card shows the letter of its stage: `🎯 c<n> <letter>` for each of the nine stages
    /// (audit N2), with no other group when the ledger holds nothing else.
    #[test]
    fn my_card_shows_the_letter_of_each_stage() {
        for (stage, letter) in [
            (Stage::Draft, 'D'),
            (Stage::Ready, 'R'),
            (Stage::Working, 'W'),
            (Stage::Stale, 'E'),
            (Stage::Submitted, 'S'),
            (Stage::Audited, 'A'),
            (Stage::Landed, 'L'),
            (Stage::Blocked, 'B'),
            (Stage::Hold, 'H'),
        ] {
            let flow = Flow {
                split: true,
                mine: Some(MyCard { card: 3, stage }),
                ..Default::default()
            };
            assert_eq!(plain(flow, None), format!("🎯 c3 {letter}"), "{stage:?}");
        }
    }

    /// The unsplit flow (the notes could not be read) is one form of three numbers, never a
    /// guess, at every width.
    #[test]
    fn the_unsplit_flow_has_one_form_and_never_a_guess() {
        let unsplit = Flow {
            draft: 43,
            ready: 150,
            working: 15,
            hold: 7,
            done_today: 19,
            split: false,
            ..Default::default()
        };
        for budget in [None, Some(200), Some(18), Some(0)] {
            assert_eq!(plain(unsplit, budget), "🗂️  D43 R150 W15", "{budget:?}");
        }
    }

    /// The colored line carries the same text once the escapes are removed, at every form.
    #[test]
    fn the_colored_line_has_the_plain_text_at_every_form() {
        for flow in [full(), no_card()] {
            for budget in [None, Some(60), Some(30), Some(10)] {
                let text = plain(flow, budget);
                for profile in [Profile::TrueColor, Profile::Ansi256] {
                    let out = only_flow(flow, budget, Some(profile));
                    let mut stripped = String::new();
                    let mut rest = out.as_str();
                    while let Some(at) = rest.find("\x1b[") {
                        stripped.push_str(&rest[..at]);
                        let tail = &rest[at..];
                        rest = &tail[tail.find('m').unwrap() + 1..];
                    }
                    stripped.push_str(rest);
                    assert_eq!(stripped, text, "{budget:?} {profile:?}");
                }
            }
        }
    }

    /// AC-8: every icon the line writes is followed by the spacing of its width class — two
    /// spaces for an icon that is narrow by default and written with a variation selector, one
    /// for the others — and a letter or a number follows the spacing.
    #[test]
    fn every_icon_is_followed_by_the_spacing_of_its_width_class() {
        let text = plain(full(), None);
        let mut seen = 0;
        for (icon, gap) in width::ICONS {
            for (at, _) in text.match_indices(icon) {
                seen += 1;
                let after = &text[at + icon.len()..];
                assert!(after.starts_with(gap), "{icon}: `{after}`");
                assert!(
                    !after[gap.len()..].starts_with(' '),
                    "{icon}: more spaces than the class asks for: `{after}`"
                );
                assert!(
                    after[gap.len()..]
                        .chars()
                        .next()
                        .is_some_and(char::is_alphanumeric),
                    "{icon} is followed by a letter or a number: `{after}`"
                );
            }
        }
        assert_eq!(seen, 6, "🎯 👀 ⛔ ⏸️ 🗂️ ✅ each once in {text}");
    }
}

/// SPEC-20261122-statusline-where AC-6: the first line in each form of the Layouts section,
/// exactly, changing at its width boundary. A fixture integration branch is `next`.
#[cfg(test)]
mod line_one_form_tests {
    use super::*;
    use crate::statusline::segments::GitStatus;
    use crate::statusline::state::{GitHubCounts, PrInfo};

    /// W0 to W4 of the Layouts section with the display width the section gives.
    const W: [(&str, usize); 5] = [
        (
            "🧠 Opus 5.5 high·t | 📡 modu-ai/moai-adk, ⬇️  7 🔀 3 | 🌳 statusline-flow ↑3 +2 | 🧭 next ↑3 ↓0 | 🔗 #12 pending",
            112,
        ),
        (
            "🧠 Opus 5.5 high·t | 📡 modu-ai/moai-adk | 🌳 statusline-flow ↑3 +2 | 🧭 next ↑3 ↓0 | 🔗 #12 pending",
            100,
        ),
        (
            "🧠 Opus high | 📡 moai-adk | 🌳 statusline-flow ↑3 +2 | 🧭 next ↑3 ↓0 | 🔗 #12",
            78,
        ),
        ("🧠 Opus | 🌳 statusline-flow ↑3 +2 | 🧭 next ↓0 | 🔗 #12", 56),
        ("🧠 Opus | 🌳 statusline-flow ↑3 +2", 34),
    ];
    const MAIN: (&str, usize) = (
        "🧠 Opus 5.5 high·t | 📡 modu-ai/moai-adk, ⬇️  7 🔀 3 | 🅱️  next ↑2 +1 | 🔗 #12 pending",
        86,
    );
    /// The next form of the main checkout: the counts go first (the drop order of the section).
    const MAIN_NO_COUNTS: &str =
        "🧠 Opus 5.5 high·t | 📡 modu-ai/moai-adk | 🅱️  next ↑2 +1 | 🔗 #12 pending";
    const GENERIC: (&str, usize) = (
        "🧠 Sonnet 5.5 | 📡 acme/webapp | 🅱️  feature/login ↑2 ↓1 +4 | 🔗 #88 approved",
        77,
    );

    fn pull_request(number: i64, state: &str, merge_request: bool) -> Option<PrInfo> {
        Some(PrInfo {
            number,
            url: String::new(),
            review_state: state.into(),
            merge_request,
        })
    }

    /// A session in a linked worktree named like its branch, three commits ahead of `next`,
    /// two changed files and a pull request.
    fn worktree_session() -> Status {
        Status {
            model: "Opus 5.5".into(),
            effort: Some("high".into()),
            thinking: Some(true),
            repo: Some(("modu-ai".into(), "moai-adk".into())),
            github: GitHubCounts {
                available: true,
                open_issues: 7,
                open_prs: 3,
                ..Default::default()
            },
            worktree: "statusline-flow".into(),
            git: Some(GitStatus {
                branch: "statusline-flow".into(),
                ahead: 3,
                modified: 1,
                untracked: 1,
                ..Default::default()
            }),
            integration: Some("next".into()),
            pr: pull_request(12, "pending", false),
            ..Default::default()
        }
    }

    /// The main checkout on the integration branch: two landings not yet pushed, one change.
    fn main_checkout() -> Status {
        Status {
            worktree: String::new(),
            git: Some(GitStatus {
                branch: "next".into(),
                ahead: 2,
                untracked: 1,
                ..Default::default()
            }),
            ..worktree_session()
        }
    }

    /// A project with no integration branch: the numbers are the branch's own upstream's.
    fn generic_project() -> Status {
        Status {
            model: "Sonnet 5.5".into(),
            effort: None,
            thinking: None,
            repo: Some(("acme".into(), "webapp".into())),
            github: GitHubCounts {
                suppressed: true,
                ..Default::default()
            },
            worktree: String::new(),
            git: Some(GitStatus {
                branch: "feature/login".into(),
                ahead: 2,
                behind: 1,
                untracked: 4,
                ..Default::default()
            }),
            integration: None,
            pr: pull_request(88, "approved", false),
            ..Default::default()
        }
    }

    /// The first line alone; the context window is switched off so no other line has text.
    fn first_line(status: &Status, budget: Option<usize>, profile: Option<Profile>) -> String {
        let config = DisplayConfig {
            segments: [("context".to_string(), false)].into_iter().collect(),
            ..Default::default()
        };
        Renderer::new(&config, profile)
            .with_budget(budget)
            .render(status)
            .lines()
            .next()
            .unwrap_or_default()
            .to_string()
    }

    fn plain(status: &Status, budget: Option<usize>) -> String {
        first_line(status, budget, None)
    }

    fn strip(text: &str) -> String {
        let mut out = String::new();
        let mut rest = text;
        while let Some(at) = rest.find("\x1b[") {
            out.push_str(&rest[..at]);
            let tail = &rest[at..];
            rest = &tail[tail.find('m').unwrap() + 1..];
        }
        out.push_str(rest);
        out
    }

    /// The widths of the section are what the crate counts (a guard on the table itself).
    #[test]
    fn the_layout_strings_have_the_widths_of_the_section() {
        for (text, cols) in W.iter().chain([&MAIN, &GENERIC]) {
            assert_eq!(width::display_width(text), *cols, "{text}");
        }
        assert!(width::display_width(MAIN_NO_COUNTS) < MAIN.1);
    }

    /// Every form is exactly the string of the Layouts section and changes at its boundary: a
    /// budget equal to the width keeps the form, one less drops to the next.
    #[test]
    fn every_form_is_exactly_the_string_of_the_layouts_section() {
        let status = worktree_session();
        for (i, (text, cols)) in W.iter().enumerate() {
            assert_eq!(&plain(&status, Some(*cols)), text, "budget {cols}");
            if i == 0 {
                assert_eq!(&plain(&status, Some(*cols + 40)), text, "{cols} + 40");
            }
            if let Some((next, _)) = W.get(i + 1) {
                assert_eq!(
                    &plain(&status, Some(*cols - 1)),
                    next,
                    "budget {}",
                    cols - 1
                );
            }
        }
    }

    /// No usable `COLUMNS` writes the widest form and a budget below the narrowest the narrowest.
    #[test]
    fn no_budget_writes_the_widest_form_and_a_small_budget_the_narrowest() {
        let status = worktree_session();
        assert_eq!(plain(&status, None), W[0].0);
        for budget in [33, 10, 0] {
            assert_eq!(plain(&status, Some(budget)), W[4].0, "{budget}");
        }
    }

    /// The main checkout and a generic project: the string of the section at its width, and a
    /// narrower form one column less.
    #[test]
    fn the_main_checkout_and_a_generic_project_are_the_strings_of_the_section() {
        for (status, (text, cols)) in [(main_checkout(), MAIN), (generic_project(), GENERIC)] {
            assert_eq!(plain(&status, None), text);
            assert_eq!(plain(&status, Some(cols)), text, "budget {cols}");
            assert_eq!(plain(&status, Some(cols + 40)), text);
            let narrower = plain(&status, Some(cols - 1));
            assert_ne!(narrower, text, "budget {}", cols - 1);
            assert!(width::display_width(&narrower) < cols, "{narrower}");
        }
        assert_eq!(plain(&main_checkout(), Some(MAIN.1 - 1)), MAIN_NO_COUNTS);
    }

    /// The colored line carries the same text once the escapes are removed, at every form.
    #[test]
    fn the_colored_first_line_has_the_plain_text_at_every_form() {
        for status in [worktree_session(), main_checkout(), generic_project()] {
            for budget in [None, Some(100), Some(78), Some(56), Some(34), Some(0)] {
                let text = plain(&status, budget);
                for profile in [Profile::TrueColor, Profile::Ansi256] {
                    let colored = first_line(&status, budget, Some(profile));
                    assert!(colored.contains('\x1b'), "{budget:?} {profile:?}");
                    assert_eq!(strip(&colored), text, "{budget:?} {profile:?}");
                }
            }
        }
    }

    /// REQ-1, REQ-8: no tag, no `🤔`, no versions on the first line.
    #[test]
    fn the_first_line_has_no_tag_no_old_effort_icon_and_no_versions() {
        let mut status = worktree_session();
        status.claude_version = "2.1.330".into();
        status.version = Some(crate::statusline::state::Version {
            current: "4.2647.16".into(),
            latest: Some("4.2648.1".into()),
        });
        let line = plain(&status, None);
        assert_eq!(line, W[0].0);
        // The tag icon is written as its code points, so this file names no tag text.
        for gone in [
            "\u{1F3F7}\u{FE0F}",
            "🤔",
            "🤖",
            "🗿",
            "Leader",
            "Agent",
            "agent",
        ] {
            assert!(!line.contains(gone), "{gone} in {line}");
        }
    }

    /// REQ-9: a merge request is `!<n>`, a pull request `#<n>`, the state is optional and a
    /// missing request leaves the item out with its separator.
    #[test]
    fn the_request_item_is_a_number_sign_or_an_exclamation_mark() {
        let mut status = generic_project();
        status.pr = pull_request(31, "pending", true);
        assert_eq!(
            plain(&status, None),
            "🧠 Sonnet 5.5 | 📡 acme/webapp | 🅱️  feature/login ↑2 ↓1 +4 | 🔗 !31 pending"
        );
        status.pr = pull_request(31, "", true);
        assert!(plain(&status, None).ends_with(" | 🔗 !31"), "no state");
        status.pr = pull_request(31, "", false);
        assert!(plain(&status, None).ends_with(" | 🔗 #31"), "no state");
        status.pr = None;
        assert_eq!(
            plain(&status, None),
            "🧠 Sonnet 5.5 | 📡 acme/webapp | 🅱️  feature/login ↑2 ↓1 +4"
        );
    }

    /// REQ-10: the model comes from the model field and the effort from the effort and thinking
    /// fields only; a missing model leaves the item out, and an effort without thinking has no
    /// `·t`.
    #[test]
    fn the_model_and_the_effort_are_written_as_the_stdin_gives_them() {
        let mut status = generic_project();
        status.model = String::new();
        assert_eq!(
            plain(&status, None),
            "📡 acme/webapp | 🅱️  feature/login ↑2 ↓1 +4 | 🔗 #88 approved"
        );
        status.model = "Opus 4.1".into();
        status.effort = Some("high".into());
        assert!(plain(&status, None).starts_with("🧠 Opus 4.1 high | "));
        status.thinking = Some(false);
        assert!(plain(&status, None).starts_with("🧠 Opus 4.1 high | "));
        status.thinking = Some(true);
        assert!(plain(&status, None).starts_with("🧠 Opus 4.1 high·t | "));
        status.effort = Some(String::new());
        status.thinking = None;
        assert!(plain(&status, None).starts_with("🧠 Opus 4.1 | "));
    }

    /// REQ-1: an item without its input is left out with its separator, and the first line is
    /// left out when all of them are.
    #[test]
    fn an_item_without_its_input_is_left_out_with_its_separator() {
        let mut status = worktree_session();
        status.integration = None;
        status.git = Some(GitStatus {
            branch: "statusline-flow".into(),
            ahead: 3,
            behind: 1,
            modified: 1,
            untracked: 1,
            ..Default::default()
        });
        assert_eq!(
            plain(&status, None),
            "🧠 Opus 5.5 high·t | 📡 modu-ai/moai-adk, ⬇️  7 🔀 3 | 🌳 statusline-flow ↑3 ↓1 +2 | 🔗 #12 pending",
            "no integration branch: the branch item keeps its own numbers"
        );
        assert_eq!(plain(&Status::default(), None), "MoAI");
    }

    /// A detached HEAD: its short commit in the branch item, no distance.
    #[test]
    fn a_detached_head_shows_its_short_commit_and_no_distance() {
        let status = Status {
            git: Some(GitStatus {
                branch: String::new(),
                head_sha: "0123456".into(),
                untracked: 3,
                ..Default::default()
            }),
            integration: Some("next".into()),
            ..generic_project()
        };
        assert_eq!(
            plain(&status, None),
            "🧠 Sonnet 5.5 | 📡 acme/webapp | 🅱️  0123456 +3 | 🔗 #88 approved"
        );
    }

    /// The integration branch name is cut with `…` over twelve characters, and written whole at
    /// twelve.
    #[test]
    fn the_integration_name_is_cut_over_twelve_characters() {
        let item = |name: &str| {
            let mut status = worktree_session();
            status.integration = Some(name.into());
            let line = plain(&status, None);
            let from = line.find("🧭 ").expect("the distance item") + "🧭 ".len();
            let rest = &line[from..];
            rest[..rest.find(" ↑").expect("its ahead count")].to_string()
        };
        assert_eq!(item("abcdefghijkl"), "abcdefghijkl", "12 characters");
        let cut = item("abcdefghijklm");
        assert!(cut.starts_with("abcdefghijk"), "{cut}");
        assert!(cut.ends_with('…'), "{cut}");
        assert!(!cut.contains('m'), "{cut}");
    }

    /// REQ-10 and D-6: every icon of the first line is followed by the spacing of its width
    /// class: two spaces for `⬇️` and `🅱️`, one for the rest.
    #[test]
    fn every_icon_of_the_first_line_has_the_spacing_of_its_width_class() {
        for (status, expected) in [(worktree_session(), 7), (main_checkout(), 6)] {
            let text = plain(&status, None);
            let mut seen = 0;
            for (icon, gap) in width::ICONS {
                for (at, _) in text.match_indices(icon) {
                    seen += 1;
                    let after = &text[at + icon.len()..];
                    assert!(after.starts_with(gap), "{icon}: `{after}`");
                    assert!(
                        !after[gap.len()..].starts_with(' '),
                        "{icon}: more spaces than the class asks for: `{after}`"
                    );
                }
            }
            assert_eq!(seen, expected, "icons in {text}");
        }
    }
}
