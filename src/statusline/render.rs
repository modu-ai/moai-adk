//! The main Claude Code statusline: up to three lines joined by `\n`, usage
//! windows as a bar and a number, and two Catppuccin themes. Without color the text is identical
//! and carries no escape sequence.

use super::color::Profile;
use super::icons;
use super::input::RateWindow;
use super::state::{DisplayConfig, Flow, Status};
use super::width;
use std::collections::BTreeMap;

const SEPARATOR: &str = " | ";
/// The forms of line 1, from the widest (step 0) to the narrowest (step 4).
const LINE_ONE_STEPS: usize = 5;
/// The forms of line 3, from the widest (Z0) to the narrowest (Z3).
const LINE_THREE_STEPS: usize = 4;
/// The cells of the bar that shows the share used of a limit window
/// (SPEC-20261123-statusline-order-bars REQ-3, read as the share used by
/// SPEC-20261124-statusline-counts-truth REQ-12).
const BAR_WIDTH: i64 = 10;
/// The used percentage above which a window counts as nearly spent and shows `🪫` instead of `🔋`
/// (SPEC-20261122-statusline-limits D-2; SPEC-20261124-statusline-counts-truth REQ-13).
const HIGH_USE_PERCENT: i64 = 70;
/// The filled cells of a bar are green with this share used or less, yellow up to
/// [`LEVEL_YELLOW_UP_TO`], peach up to [`HIGH_USE_PERCENT`] and red above it, so the bar turns red
/// where the battery icon turns to `🪫` (SPEC-20261124-statusline-counts-truth REQ-14, D-10).
const LEVEL_GREEN_UP_TO: i64 = 30;
const LEVEL_YELLOW_UP_TO: i64 = 50;
/// The cost is shown below this many US dollars (at most `999999.99`, nine characters); a larger
/// value (`1e308` printed a 337-character line) is left out.
const MAX_COST_USD: f64 = 1_000_000.0;
const SGR_NORMAL_INTENSITY: &str = "\x1b[22m";
const SGR_RESET: &str = "\x1b[0m";

/// A theme's colors (internal/tui/catppuccin.go).
/// The per-segment colors are the #2559 segment palette (user decision 2026-11-12): each
/// segment kind takes its own Catppuccin accent, so neighbours never share a color.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub struct Theme {
    pub text: &'static str,
    pub muted: &'static str,
    /// The filled cells of a limit bar by the share left: green, yellow, peach, red
    /// ([`level_color`]).
    pub level: [&'static str; 4],
    /// `🧠` model (mauve).
    pub model: &'static str,
    /// The effort and thinking text after the model name (lavender).
    pub effort: &'static str,
    /// `♻️` cache hit (teal).
    pub cache: &'static str,
    /// `🤖` Claude Code and `🗿` MoAI versions (sky).
    pub version: &'static str,
    /// `📡` repository (blue).
    pub repo: &'static str,
    /// `⬇️` open Issues (peach).
    pub issues: &'static str,
    /// `🔀` open pull requests and `🔗` the request of the branch (pink).
    pub prs: &'static str,
    /// `🅱️` integration branch with its tip and `🌳` checked-out branch (green).
    pub branch: &'static str,
    /// `🧭` distance to the integration branch (yellow); no layout prints it since
    /// SPEC-20261123-statusline-order-bars REQ-11, the colour stays with the unused distance code.
    pub distance: &'static str,
    /// `▶️` cards being worked on (maroon).
    pub cards_active: &'static str,
    /// `🔃` cards pushed and waiting for integration or an audit (rosewater).
    pub cards_posted: &'static str,
    /// `📋` cards created and remaining (sapphire).
    pub cards_waiting: &'static str,
    /// `✅` cards done (teal).
    pub cards_done: &'static str,
}

pub const MOCHA: Theme = Theme {
    text: "#CDD6F4",
    muted: "#6C7086",
    level: ["#A6E3A1", "#F9E2AF", "#FAB387", "#F38BA8"],
    model: "#CBA6F7",
    effort: "#B4BEFE",
    cache: "#94E2D5",
    version: "#89DCEB",
    repo: "#89B4FA",
    issues: "#FAB387",
    prs: "#F5C2E7",
    branch: "#A6E3A1",
    distance: "#F9E2AF",
    cards_active: "#EBA0AC",
    cards_posted: "#F5E0DC",
    cards_waiting: "#74C7EC",
    cards_done: "#94E2D5",
};

pub const LATTE: Theme = Theme {
    text: "#4C4F69",
    muted: "#9CA0B0",
    level: ["#40A02B", "#DF8E1D", "#FE640B", "#D20F39"],
    model: "#8839EF",
    effort: "#7287FD",
    cache: "#179299",
    version: "#04A5E5",
    repo: "#1E66F5",
    issues: "#FE640B",
    prs: "#EA76CB",
    branch: "#40A02B",
    distance: "#DF8E1D",
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
    /// An OSC 8 link around the painted text; empty for none ([`Renderer::link`]).
    url: String,
    /// No space before this piece: it follows the piece before it directly. The filled and the
    /// empty cells of a bar are two pieces of two colors in one run of text.
    glue: bool,
}

/// Pieces joined by one space (`⛔ 3 🙋 7` is two pieces, `📋 150 → 🔨 14` three), except a piece
/// that is glued to the one before it.
type Group = Vec<Piece>;

/// The plain text of a group: its pieces, a space before each one but the first and the glued
/// ones, each piece written by `write`.
fn join_group<'a>(group: &'a [Piece], write: impl Fn(&'a Piece) -> String) -> String {
    let mut text = String::new();
    for (at, piece) in group.iter().enumerate() {
        if at > 0 && !piece.glue {
            text.push(' ');
        }
        text.push_str(&write(piece));
    }
    text
}

/// The plain text of a form: its groups, each group's pieces joined by a space (a glued piece
/// without one), the groups joined by the separator. Its display width is what the form picker
/// measures.
fn form_text(form: &[Group]) -> String {
    form.iter()
        .map(|g| join_group(g, |p| p.text.clone()))
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
    /// SPEC-20261122-statusline-where: line 1 says who the session is and where it stands — the
    /// model with its effort, the repository with its forge counts, the worktree or the branch,
    /// the distance to the integration branch and the pull request ([`Renderer::info_line`]);
    /// the second line is the two versions, the cards by stage and my card
    /// ([`Renderer::flow_line`], SPEC-20261123-statusline-order-bars); the last line is the limits
    /// ([`Renderer::limits_line`]). The cache segment is not in the layout. A line with nothing
    /// in it is left out, so the numbers above shift up by one for each empty line before it.
    pub fn render(&self, status: &Status) -> String {
        let lines: Vec<String> = [
            self.info_line(status),
            self.flow_line(status),
            self.limits_line(status),
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

    /// Line 1 (SPEC-20261122-statusline-where REQ-1 and REQ-11): who the session is and where it
    /// stands, in the widest of its forms that fits the width budget. An item without its input
    /// is left out with its separator, and the line is empty when all of them are.
    fn info_line(&self, s: &Status) -> String {
        let forms = self.line_one_forms(s);
        if forms.is_empty() {
            return String::new();
        }
        let widths: Vec<usize> = forms
            .iter()
            .map(|f| width::display_width(&form_text(f)))
            .collect();
        self.paint_form(&forms[width::pick(&widths, self.budget)])
    }

    /// The forms of line 1 from the widest to the narrowest, never an empty one and never two
    /// equal neighbours.
    fn line_one_forms(&self, s: &Status) -> Vec<Vec<Group>> {
        let mut forms: Vec<Vec<Group>> = Vec::new();
        for step in 0..LINE_ONE_STEPS {
            let form = self.line_one_form(s, step);
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

    /// One form of line 1 (SPEC-20261123-statusline-order-bars REQ-8 to REQ-12): the model with its
    /// effort, the repository with its forge counts, the integration branch with the tip of its
    /// local ref, the checked-out branch and the request. Each step drops more than the one before
    /// it, in the order of the Layouts section of the SPEC: the issue and pull request counts
    /// (step 1), the repository owner, the model version, the thinking mark and the request state
    /// (step 2), the effort and the repository (step 3), the commit of the integration item and
    /// the request (step 4). The model, the integration name and the checked-out branch stay in
    /// every form.
    fn line_one_form(&self, s: &Status, step: usize) -> Vec<Group> {
        let t = &self.theme;
        let piece = |color, text: String| Piece {
            color,
            text,
            url: String::new(),
            glue: false,
        };
        let mut form: Vec<Group> = Vec::new();
        if self.enabled("model") && !s.model.is_empty() {
            let model = if step >= 2 {
                s.model.split_whitespace().next().unwrap_or(&s.model)
            } else {
                &s.model
            };
            let mut group = vec![piece(t.model, width::icon_text("🧠", model))];
            if self.enabled("effort_thinking") && step < 3 {
                let thinking = s.thinking == Some(true) && step < 2;
                // Thinking without an effort is `·t` alone.
                let effort = format!(
                    "{}{}",
                    s.effort.as_deref().unwrap_or(""),
                    if thinking { "·t" } else { "" }
                );
                if !effort.is_empty() {
                    group.push(piece(t.effort, effort));
                }
            }
            form.push(group);
        }
        if self.enabled("git_branch") && step < 3 {
            if let Some((owner, name)) = s
                .repo
                .as_ref()
                .filter(|(o, n)| !o.is_empty() && !n.is_empty())
            {
                let shown = if step < 2 {
                    format!("{owner}/{name}")
                } else {
                    name.clone()
                };
                let counts = step == 0 && self.enabled("github") && !s.github.suppressed;
                let comma = if counts { "," } else { "" };
                let mut group = vec![piece(
                    t.repo,
                    format!("{}{comma}", width::icon_text("📡", &shown)),
                )];
                if counts {
                    // A failed or missing query says `n/a`, never a zero (REQ-3 of
                    // SPEC-20261121-statusline-detached).
                    let (issues, prs, issues_color, prs_color) = if s.github.available {
                        (
                            s.github.open_issues.to_string(),
                            s.github.open_prs.to_string(),
                            t.issues,
                            t.prs,
                        )
                    } else {
                        ("n/a".to_string(), "n/a".to_string(), t.muted, t.muted)
                    };
                    group.push(piece(issues_color, width::icon_text("⬇️", &issues)));
                    group.push(piece(prs_color, width::icon_text("🔀", &prs)));
                }
                form.push(group);
            }
        }
        if self.enabled("git_branch") {
            // REQ-9: the configured integration branch with the first seven characters of the
            // commit its local ref names — `n/a` when the ref cannot be read, never a guessed
            // commit, and never the commit of `HEAD`. Without an integration branch there is no
            // item.
            if let Some(name) = s.integration.as_deref().filter(|n| !n.is_empty()) {
                let mut text = width::icon_text("🅱️", name);
                if step < 4 {
                    let tip = if s.integration_tip.is_empty() {
                        "n/a"
                    } else {
                        &s.integration_tip
                    };
                    text.push_str(&format!("({})", width::icon_text("🏷️", tip)));
                }
                form.push(vec![piece(t.branch, text)]);
            }
            // REQ-10: the branch the `HEAD` file names, never the name of the working folder; a
            // detached or unreadable `HEAD` has none and no name is invented.
            if self.enabled("worktree") {
                if let Some(branch) = s
                    .git
                    .as_ref()
                    .map(|g| g.branch.as_str())
                    .filter(|b| !b.is_empty())
                {
                    form.push(vec![piece(t.branch, width::icon_text("🌳", branch))]);
                }
            }
        }
        if self.enabled("pr") && step < 4 {
            if let Some(pr) = &s.pr {
                let sign = if pr.merge_request { '!' } else { '#' };
                let mut text = format!("{sign}{}", pr.number);
                if step < 2 && !pr.review_state.is_empty() {
                    text.push(' ');
                    text.push_str(&pr.review_state);
                }
                form.push(vec![Piece {
                    color: t.prs,
                    text: width::icon_text("🔗", &text),
                    url: pr.url.clone(),
                    glue: false,
                }]);
            }
        }
        form
    }

    /// Line 3 (SPEC-20261123-statusline-order-bars REQ-3 to REQ-5): the context window, the
    /// five-hour window, the seven-day window and the cost, in the widest of its forms that fits
    /// the width budget. An item without its input is left out with its separator, and the line is
    /// empty when all of them are. The two versions are on the second line.
    fn limits_line(&self, s: &Status) -> String {
        let forms = self.line_three_forms(s);
        if forms.is_empty() {
            return String::new();
        }
        let widths: Vec<usize> = forms
            .iter()
            .map(|f| width::display_width(&form_text(f)))
            .collect();
        self.paint_form(&forms[width::pick(&widths, self.budget)])
    }

    /// The forms of line 3 from the widest to the narrowest, never an empty one and never two
    /// equal neighbours. Each step drops more than the one before it, in the order under the
    /// Layouts table of the SPEC: the reset countdowns (Z1), the bars (Z2), the five-hour window
    /// and the cost (Z3, the narrowest form). The context window and the seven-day window stay in
    /// every form, so the five-hour window always goes before them.
    ///
    /// A known window is `<battery> <label> <bar> <n>%` — no colon, no `/clear` hint — and `<n>`
    /// is the share left, not the share used (SPEC-20261122-statusline-limits-left REQ-1); the bar
    /// draws the same share ([`Window`]). A window whose percentage is unknown says `n/a` with no
    /// battery icon and no bar, never a guessed number.
    fn line_three_forms(&self, s: &Status) -> Vec<Vec<Group>> {
        let t = &self.theme;
        let piece = |color, text: String| Piece {
            color,
            text,
            url: String::new(),
            glue: false,
        };
        // REQ-3: the host drops a window once its reset time passes, so one that still names a
        // passed reset is left out; under a minute left (1 to 59 s) prints `(0m)`.
        let live = |w: Option<RateWindow>| w.filter(|w| w.resets_at.is_none_or(|at| at > self.now));
        // A reset further ahead than the longest window allows is no time to count down to
        // (the usage record drops it too): the window shows without a countdown.
        let countdown = |w: &RateWindow| {
            w.resets_at
                .map(|at| at.saturating_sub(self.now))
                .filter(|left| *left <= crate::usage_limit::MAX_RESET_AHEAD_SECS)
                .map(|left| piece(t.muted, reset_countdown(left)))
        };
        let window = |label: &str, used: Option<f64>, countdown: Option<Piece>| match used {
            Some(pct) => Window::known(t, label, whole_percent(pct), countdown),
            None => Window::unknown(t, label, countdown),
        };
        let context: Option<Window> = self
            .enabled("context")
            .then(|| window("CW", s.context, None));
        // SPEC-20261122-statusline-limits-left REQ-3: the five-hour window whenever the input
        // carries its percentage, at any level; the host does not always send it, and a window
        // without a percentage is left out, never `5H n/a`.
        let five_hour = live(s.five_hour)
            .filter(|_| self.enabled("usage_5h"))
            .and_then(|w| {
                let used = w.used_percentage?;
                Some(window("5H", Some(used), countdown(&w)))
            });
        let seven_day = live(s.seven_day)
            .filter(|_| self.enabled("usage_7d"))
            .map(|w| window("7D", w.used_percentage, countdown(&w)));
        // REQ-5 and D-3: the cost only when the payload carries neither window (a payload with
        // only a spend limit included), to two decimals.
        let cost: Option<Group> = s
            .cost
            .filter(|_| s.five_hour.is_none() && s.seven_day.is_none() && self.enabled("cost"))
            // A negative cost or one too large for the line is no spend to show, so it is left
            // out; a negative zero is shown as `0.00`, never `-0.00`.
            .filter(|usd| (0.0..MAX_COST_USD).contains(usd))
            .map(|usd| if usd == 0.0 { 0.0 } else { usd })
            .map(|usd| vec![piece("", width::icon_text("💰", &format!("{usd:.2}")))]);
        let mut forms: Vec<Vec<Group>> = Vec::new();
        for step in 0..LINE_THREE_STEPS {
            let (bars, countdowns, five_and_cost) = (step < 2, step < 1, step < 3);
            let show = |w: &Option<Window>| w.as_ref().map(|w| w.group(bars, countdowns));
            let form: Vec<Group> = [
                show(&context),
                show(&five_hour).filter(|_| five_and_cost),
                show(&seven_day),
                cost.clone().filter(|_| five_and_cost),
            ]
            .into_iter()
            .flatten()
            .collect();
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

    /// The second line (SPEC-20261123-statusline-order-bars REQ-1 and REQ-2): the Claude Code
    /// version, the MoAI version, then the cards of the project's local ledger by the stage of
    /// their progress under the one `cards` option, then my card at the end. It is written in
    /// the widest of its forms that fits the width budget; empty when there is nothing to show.
    /// Each number has its icon and one space after it ([`width::icon_text`]), so color is never
    /// the only signal.
    fn flow_line(&self, s: &Status) -> String {
        let forms = self.flow_forms(s);
        if forms.is_empty() {
            return String::new();
        }
        let widths: Vec<usize> = forms
            .iter()
            .map(|f| width::display_width(&form_text(f)))
            .collect();
        self.paint_form(&forms[width::pick(&widths, self.budget)])
    }

    /// A form written out: each piece in its color (and its link), the pieces of a group joined
    /// by a space (a glued piece without one) and the groups by the separator. A piece without a
    /// color stays in the base color of the line.
    fn paint_form(&self, form: &[Group]) -> String {
        form.iter()
            .map(|g| {
                join_group(g, |p| {
                    let text = if p.color.is_empty() {
                        p.text.clone()
                    } else {
                        self.paint(p.color, &p.text)
                    };
                    self.link(&p.url, &text)
                })
            })
            .collect::<Vec<_>>()
            .join(SEPARATOR)
    }

    /// The forms of the second line from the widest to the narrowest, never an empty one. The
    /// widest form is the two versions, the flow and what stands beside it, then my card:
    /// `🤖 <Claude Code version> | 🗿 <MoAI version> | 📋 <ready> → 🔨 <working> → 👀 <submitted>
    /// → 🚀 <audited + landed> → ✅ <done today> | ⛔ <blocked> ⏸️ <paused> 🙋 <hold> | 🎯 C-<n>
    /// <stage icon>` (the icons of [`icons`](super::icons), SPEC-20261124-card-flow-icons).
    /// An item without its input is left out with its separator: a line with no flow holds the
    /// versions alone, one with no version the flow alone, and the options `claude_version`,
    /// `moai_version` and `cards` each hide their own items.
    ///
    /// While the width budget is smaller than the form, the Claude Code version drops first, then
    /// the MoAI version with its update arrow, then the leftmost item of the flow (`📋` first,
    /// `✅` last); my card and `⛔ ⏸️ 🙋` stay in every form
    /// (SPEC-20261123-statusline-order-bars REQ-2, D-5).
    fn flow_forms(&self, s: &Status) -> Vec<Vec<Group>> {
        let t = &self.theme;
        // REQ-6 of the limits SPEC: the versions without a `v`; the update arrow names the
        // template, then the build.
        let claude: Option<Group> =
            (self.enabled("claude_version") && !s.claude_version.is_empty()).then(|| {
                vec![Piece {
                    color: t.version,
                    text: width::icon_text("🤖", &s.claude_version),
                    url: String::new(),
                    glue: false,
                }]
            });
        let moai: Option<Group> = s
            .version
            .as_ref()
            .filter(|v| self.enabled("moai_version") && !v.current.is_empty())
            .map(|v| {
                let text = match &v.latest {
                    Some(build) => format!("{} -> {build}", v.current),
                    None => v.current.clone(),
                };
                vec![Piece {
                    color: t.version,
                    text: width::icon_text("🗿", &text),
                    url: String::new(),
                    glue: false,
                }]
            });
        // The forms of what follows the versions, from the widest to the narrowest; one empty
        // form when there is no flow to show.
        let tails = match s.flow.filter(|_| self.enabled("cards")) {
            Some(flow) => self.flow_tails(&flow),
            None => vec![Vec::new()],
        };
        let widest = tails.first().cloned().unwrap_or_default();
        // The ladder: both versions, then the MoAI version alone, then none, each in front of the
        // widest tail; then each narrower tail alone.
        let in_front = |with_claude: bool, with_moai: bool| -> Vec<Group> {
            let mut groups: Vec<Group> = Vec::new();
            if with_claude {
                groups.extend(claude.clone());
            }
            if with_moai {
                groups.extend(moai.clone());
            }
            groups.extend(widest.clone());
            groups
        };
        let candidates = [
            in_front(true, true),
            in_front(false, true),
            in_front(false, false),
        ]
        .into_iter()
        .chain(tails.iter().skip(1).cloned());
        let mut forms: Vec<Vec<Group>> = Vec::new();
        for form in candidates {
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

    /// What follows the versions on the second line, in the forms of the flow from the widest to
    /// the narrowest (SPEC-20261122-statusline-limits-left REQ-6 and REQ-7, with my card moved to
    /// the end by SPEC-20261123-statusline-order-bars REQ-1): the flow, what stands beside it,
    /// my card. There is no letter code, an item whose count is zero is left out with its arrow
    /// or its separator, and the drafts are not on the line. A card whose lease ended without a
    /// hand-back is in `🔨` with the working ones, `⏸️` counts the hold cards that were withdrawn
    /// or folded and `🙋` the hold cards nobody withdrew (SPEC-20261124-statusline-counts-truth,
    /// SPEC-20261124-card-flow-icons).
    /// Each narrower form drops the leftmost item still in the flow — `📋` first, `✅` last —
    /// while `⛔ ⏸️ 🙋` and my card stay. An unsplit [`Flow`] (the notes could not be read) has one
    /// form, `📋 <ready> 🔨 <working>` without an arrow, never a guess. One empty form is returned
    /// when there is nothing to show.
    fn flow_tails(&self, f: &Flow) -> Vec<Vec<Group>> {
        let t = &self.theme;
        let piece = |color, icon, text: String| Piece {
            color,
            text: width::icon_text(icon, &text),
            url: String::new(),
            glue: false,
        };
        // An item of a count: left out when there is nothing to count.
        let item =
            |color, icon, count: i64| (count > 0).then(|| piece(color, icon, count.to_string()));
        if !f.split {
            let known: Group = [
                item(t.cards_waiting, icons::TODO, f.ready),
                item(t.cards_active, icons::WORK, f.working),
            ]
            .into_iter()
            .flatten()
            .collect();
            return if known.is_empty() {
                vec![Vec::new()]
            } else {
                vec![vec![known]]
            };
        }
        let mine: Option<Group> = f.mine.map(|m| {
            vec![piece(
                t.cards_active,
                "🎯",
                format!("C-{} {}", m.card, m.stage.icon()),
            )]
        });
        let chain: Vec<Piece> = [
            item(t.cards_waiting, icons::TODO, f.ready),
            // A card a lane works on whose lease ended is still in hand (SPEC-20261124-
            // statusline-counts-truth REQ-7): it is drawn with the working ones, so the items
            // add up to the open cards.
            item(
                t.cards_active,
                icons::WORK,
                f.working.saturating_add(f.stale),
            ),
            item(t.cards_posted, icons::CHECK, f.submitted),
            item(t.branch, icons::READY, f.audited.saturating_add(f.landed)),
            item(t.cards_done, icons::DONE, f.done_today),
        ]
        .into_iter()
        .flatten()
        .collect();
        let beside: Group = [
            item(t.issues, icons::BLOCKED, f.blocked),
            item(t.issues, icons::PAUSED, f.paused),
            item(t.issues, icons::QUESTION, f.hold),
        ]
        .into_iter()
        .flatten()
        .collect();
        let arrow = || Piece {
            color: t.muted,
            text: "→".to_string(),
            url: String::new(),
            glue: false,
        };
        let mut forms: Vec<Vec<Group>> = Vec::new();
        for dropped in 0..=chain.len() {
            // The items left in the flow, an arrow between each two.
            let mut flow: Group = Vec::new();
            for item in &chain[dropped..] {
                if !flow.is_empty() {
                    flow.push(arrow());
                }
                flow.push(item.clone());
            }
            let form: Vec<Group> = [Some(flow), Some(beside.clone()), mine.clone()]
                .into_iter()
                .flatten()
                .filter(|group| !group.is_empty())
                .collect();
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
        if forms.is_empty() {
            forms.push(Vec::new());
        }
        forms
    }
}

/// One limit window of line 3 as pieces (SPEC-20261123-statusline-order-bars REQ-3, REQ-4,
/// REQ-6): the head `<battery> <label>`, the bar, the number and the reset countdown. The pieces
/// are the forms of the line, so a narrower form leaves some of them out.
struct Window {
    head: Piece,
    /// The filled cells, then the empty cells; empty for a window with an unknown percentage.
    bar: Vec<Piece>,
    number: Piece,
    countdown: Option<Piece>,
}

impl Window {
    /// A window with a known percentage: the number is the share used, the battery icon turns
    /// from `🔋` to `🪫` when more than [`HIGH_USE_PERCENT`] is used
    /// (SPEC-20261124-statusline-counts-truth REQ-13), and the bar of [`BAR_WIDTH`] cells draws
    /// the same share (REQ-12): the first `used * BAR_WIDTH / 100` (rounded down) cells are `█`
    /// and the others `░`. The filled cells take the level color of the share used
    /// ([`level_color`], REQ-14) and the empty cells the muted color, two runs with no space between them; the
    /// glyphs tell them apart without color. `used` is the whole used percentage from 0 to 100.
    fn known(t: &Theme, label: &str, used: i64, countdown: Option<Piece>) -> Window {
        let used = used.clamp(0, 100);
        let icon = if used > HIGH_USE_PERCENT {
            "🪫"
        } else {
            "🔋"
        };
        let filled = (used * BAR_WIDTH / 100) as usize;
        let empty = (BAR_WIDTH as usize).saturating_sub(filled);
        let cells = |glyph: char, n: usize| glyph.to_string().repeat(n);
        let mut bar = Vec::new();
        if filled > 0 {
            bar.push(Piece {
                color: level_color(t, used),
                text: cells('█', filled),
                url: String::new(),
                glue: false,
            });
        }
        if empty > 0 {
            bar.push(Piece {
                color: t.muted,
                text: cells('░', empty),
                url: String::new(),
                // The two runs of a bar touch; the empty cells alone follow the label by a space.
                glue: filled > 0,
            });
        }
        Window {
            head: Piece {
                color: "",
                text: width::icon_text(icon, label),
                url: String::new(),
                glue: false,
            },
            bar,
            number: Piece {
                color: "",
                text: format!("{used}%"),
                url: String::new(),
                glue: false,
            },
            countdown,
        }
    }

    /// A window whose percentage is unknown: `<label> n/a` with no battery icon and no bar.
    fn unknown(t: &Theme, label: &str, countdown: Option<Piece>) -> Window {
        let piece = |color, text: &str| Piece {
            color,
            text: text.to_string(),
            url: String::new(),
            glue: false,
        };
        Window {
            head: piece("", label),
            bar: Vec::new(),
            number: piece(t.muted, "n/a"),
            countdown,
        }
    }

    /// The pieces of one form: the head, the bar when `bars`, the number and the countdown when
    /// `countdown`.
    fn group(&self, bars: bool, countdown: bool) -> Group {
        let mut group = vec![self.head.clone()];
        if bars {
            group.extend(self.bar.iter().cloned());
        }
        group.push(self.number.clone());
        if countdown {
            group.extend(self.countdown.clone());
        }
        group
    }
}

/// The whole used percentage of a window: the fraction cut off, from 0 to 100 (REQ-10 of
/// SPEC-20261121-statusline-detached).
fn whole_percent(pct: f64) -> i64 {
    (pct as i64).clamp(0, 100)
}

/// The color of the filled cells of a bar with `used` percent used: green up to
/// [`LEVEL_GREEN_UP_TO`], yellow up to [`LEVEL_YELLOW_UP_TO`], peach up to [`HIGH_USE_PERCENT`] and
/// red above it (SPEC-20261124-statusline-counts-truth REQ-14). It reads the share used that the
/// number shows.
fn level_color(t: &Theme, used: i64) -> &'static str {
    if used <= LEVEL_GREEN_UP_TO {
        t.level[0]
    } else if used <= LEVEL_YELLOW_UP_TO {
        t.level[1]
    } else if used <= HIGH_USE_PERCENT {
        t.level[2]
    } else {
        t.level[3]
    }
}

/// The reset countdown from the seconds left, without the words `resets in`
/// (SPEC-20261122-statusline-limits D-4): a day or more counts days and hours (`(6d 21h)`), under
/// a day hours and minutes (`(2h 10m)`), under an hour minutes only (`(35m)`).
fn reset_countdown(secs_left: i64) -> String {
    if secs_left >= 86_400 {
        format!(
            "({}d {}h)",
            secs_left / 86_400,
            (secs_left % 86_400) / 3_600
        )
    } else if secs_left >= 3_600 {
        format!("({}h {}m)", secs_left / 3_600, (secs_left % 3_600) / 60)
    } else {
        format!("({}m)", (secs_left / 60).max(0))
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
    use crate::statusline::icons::{BLOCKED, CHECK, DONE, QUESTION, READY, TODO, WORK};
    use crate::statusline::input;
    use crate::statusline::segments;
    use crate::statusline::state::{self, Context, Flow, GitHubCounts, MyCard, Stage, Version};

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
            claude_version: "2.1.293".into(),
            version: Some(Version {
                current: "2.0".into(),
                latest: Some("3.0".into()),
            }),
            context: Some(50.0),
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

    /// The limits line of [`full_status`] (SPEC-20261124-statusline-counts-truth REQ-12,
    /// SPEC-20261123-statusline-order-bars): each number is the share used (50, 25 and 75 percent
    /// used) with a ten-cell bar of the same share before it, and the five-hour window resets one
    /// second after the render's clock.
    const FULL_WINDOWS: &str =
        "🔋 CW █████░░░░░ 50% | 🔋 5H ██░░░░░░░░ 25% (0m) | 🪫 7D ███████░░░ 75%";

    /// The versions of [`full_status`], which open the second line: the Claude Code version, then
    /// the MoAI version with the update arrow naming the template, then the build.
    const FULL_VERSIONS: &str = "🤖 2.1.293 | 🗿 2.0 -> 3.0";

    /// SPEC-20261123-statusline-order-bars REQ-8 and REQ-11: line 1 is the model with its effort,
    /// the repository with its forge counts and the checked-out branch, with no ahead, behind or
    /// uncommitted mark; line 2 opens with the Claude Code version and the MoAI version (REQ-1) and
    /// the last line is the limits. With no integration branch known there is no `🅱️` item.
    #[test]
    fn full_layout_has_the_place_line_then_the_windows_with_ahead_before_behind() {
        assert_eq!(
            plain(&full_status(), &[]),
            format!(
                "🧠 Opus high·t | 📡 owner/repo, ⬇️ 7 🔀 3 | 🌳 main\n{FULL_VERSIONS}\n{FULL_WINDOWS}"
            )
        );
    }

    /// REQ-11: whatever git counts — ahead, behind, both, changed files — the first line carries
    /// no `↑`, no `↓` and no `+N`; the branch item is the branch name alone.
    #[test]
    fn the_first_line_has_no_ahead_behind_or_uncommitted_mark() {
        for porcelain in [
            "## main...origin/main [behind 2]\n M a\nA  b\n?? c\n",
            "## main...origin/main [ahead 4]\n M a\n?? b\n",
            "## main...origin/main [ahead 5, behind 2]\n M a\nA  b\n?? c\n",
        ] {
            let mut status = full_status();
            status.git = Some(segments::parse_porcelain(porcelain));
            let line = plain(&status, &[]).lines().next().unwrap().to_string();
            assert_eq!(line, "🧠 Opus high·t | 📡 owner/repo, ⬇️ 7 🔀 3 | 🌳 main");
            for mark in ['↑', '↓', '+', '🧭'] {
                assert!(!line.contains(mark), "{mark} in {line}");
            }
        }
    }

    /// The user's picked example (2026-11-20T21:5xZ), in the layout of
    /// SPEC-20261122-statusline-where: every segment present gives exactly these lines in plain
    /// output — line 1 says who and where, the flow line stands under it with the two versions
    /// in front and my card at the end, and the usage windows close the screen. The host's
    /// host's worktree name is not on the line (`🌳` is the checked-out branch); the weekly window
    /// carries its reset countdown measured from the render's clock.
    fn example_status() -> Status {
        Status {
            model: "Opus 5.5".into(),
            effort: Some("high".into()),
            thinking: Some(true),
            claude_version: "2.1.325".into(),
            version: Some(Version {
                current: "4.2647.10".into(),
                latest: None,
            }),
            context: Some(97.0),
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
            integration: Some("v2026.11.0".into()),
            integration_tip: "b3ba2b3".into(),
            ..Default::default()
        }
    }

    /// The flow of the example: one card of mine, and every group of the flow line present.
    fn example_flow() -> Flow {
        Flow {
            draft: 43,
            ready: 150,
            // The work item draws the working and the stale cards together: 8 + 6.
            working: 8,
            stale: 6,
            submitted: 11,
            audited: 2,
            landed: 0,
            blocked: 3,
            hold: 7,
            paused: 0,
            done_today: 19,
            mine: Some(MyCard {
                card: EXAMPLE_CARD,
                stage: Stage::Working,
            }),
            split: true,
        }
    }

    /// The card of mine in [`example_flow`].
    const EXAMPLE_CARD: i64 = 12;

    /// The widest form of the flow line of [`example_flow`]: the flow from the cards waiting
    /// their turn to the ones done today, what stands beside the flow, then my card with the icon
    /// of its stage. The drafts and the ended leases of the example are not on the line, and the
    /// versions are not in this string (the example status of [`example_status`] has them in
    /// front of it, see [`p0_with_versions`]).
    fn p0() -> String {
        format!(
            "{TODO} 150 → {WORK} 14 → {CHECK} 11 → {READY} 2 → {DONE} 19 | {BLOCKED} 3 {QUESTION} 7 | 🎯 C-{EXAMPLE_CARD} {WORK}"
        )
    }

    /// [`p0`] with the two versions of [`example_status`] in front.
    fn p0_with_versions() -> String {
        format!("🤖 2.1.325 | 🗿 4.2647.10 | {}", p0())
    }

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
                "🧠 Opus 5.5 high·t | 📡 modu-ai/moai-adk, ⬇️ 0 🔀 0 | 🅱️ v2026.11.0(🏷️ b3ba2b3) | 🌳 v2026.11.0\n\
                 {flow}\n\
                 🪫 CW █████████░ 97% | 🔋 7D ██░░░░░░░░ 21% (6d 21h)",
                flow = p0_with_versions()
            )
        );
    }

    /// REQ-10 of SPEC-20261123-statusline-order-bars: a detached HEAD has no checked-out branch,
    /// so `🌳` is left out and its short commit is not invented as one; the repository, its forge
    /// counts and the integration item stay, and with no HEAD at all only `🌳` is omitted.
    #[test]
    fn a_detached_head_leaves_the_checked_out_item_out_and_keeps_the_repository() {
        let mut status = full_status();
        status.git = Some(segments::GitStatus {
            branch: String::new(),
            head_sha: "a1b2c3d".into(),
            ..Default::default()
        });
        let line = plain(&status, &[]).lines().next().unwrap().to_string();
        assert_eq!(line, "🧠 Opus high·t | 📡 owner/repo, ⬇️ 7 🔀 3", "{line}");
        assert!(!line.contains("a1b2c3d") && !line.contains('🌳'), "{line}");
        status.integration = Some("next".into());
        status.integration_tip = "b3ba2b3".into();
        let line = plain(&status, &[]).lines().next().unwrap().to_string();
        assert_eq!(
            line, "🧠 Opus high·t | 📡 owner/repo, ⬇️ 7 🔀 3 | 🅱️ next(🏷️ b3ba2b3)",
            "{line}"
        );
        status.integration = None;
        status.integration_tip.clear();

        status.git = Some(segments::GitStatus::default());
        let line = plain(&status, &[]).lines().next().unwrap().to_string();
        assert_eq!(
            line, "🧠 Opus high·t | 📡 owner/repo, ⬇️ 7 🔀 3",
            "no branch and no SHA leaves the branch item out"
        );
    }

    /// REQ-1 and REQ-3 of SPEC-20261122-statusline-where: forge counts that cannot be read say
    /// `n/a` — never a zero and never a hidden segment — and a repository the host did not name
    /// leaves the repository item out whole, with its counts and its separator.
    #[test]
    fn lookup_failures_show_unknown_not_zero() {
        let mut status = full_status();
        status.github.available = false;
        let line = plain(&status, &[]).lines().next().unwrap().to_string();
        assert_eq!(
            line, "🧠 Opus high·t | 📡 owner/repo, ⬇️ n/a 🔀 n/a | 🌳 main",
            "{line}"
        );
        assert!(
            !line.contains("⬇️ 0"),
            "a failed lookup is never zero: {line}"
        );
        status.repo = None;
        let line = plain(&status, &[]).lines().next().unwrap().to_string();
        assert_eq!(line, "🧠 Opus high·t | 🌳 main", "{line}");
        assert!(!line.contains("n/a") && !line.contains('📡'), "{line}");
    }

    /// #2559 AC-7 (user instruction 2026-11-12T16:5xZ): the output style is not shown, even
    /// with the `output_style` option left enabled; since SPEC-20261122-statusline-limits REQ-11
    /// it is not read either (`input::tolerant_reader_tests`).
    #[test]
    fn the_output_style_is_never_shown() {
        let status = full_status();
        let out = plain(&status, &[]);
        assert!(!out.contains('💬'), "{out}");
        assert!(!out.contains("MoAI"), "{out}");
        assert_eq!(out.lines().nth(1), Some(FULL_VERSIONS), "{out}");
    }

    /// #2756 AC-1 (user instruction 2026-11-13): the session time is not shown, even with the
    /// `session_time` option left enabled; since SPEC-20261122-statusline-limits REQ-11 it is not
    /// read either (`input::tolerant_reader_tests`).
    #[test]
    fn the_session_time_is_never_shown() {
        let status = full_status();
        let out = plain(&status, &[]);
        assert!(!out.contains('⏳'), "{out}");
        let config = DisplayConfig {
            segments: [("session_time".to_string(), true)].into_iter().collect(),
            ..Default::default()
        };
        let enabled = Renderer::new(&config, None).render(&status);
        assert!(!enabled.contains('⏳'), "{enabled}");
    }

    /// #2756 leader decision 6862179411 item 2: no `💾` file-change segment, even with its
    /// `git_status` option left enabled; the branch keeps its `+N` changes. The five-hour window
    /// the input carries is shown at any level (SPEC-20261122-statusline-limits-left REQ-3,
    /// `window_tests`).
    #[test]
    fn the_file_change_segment_is_never_shown_and_the_five_hour_window_is() {
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
        assert!(out.contains("🔋 5H ██░░░░░░░░ 25%"), "{out}");
        assert!(!out.contains('💾'), "{out}");
        assert!(out.contains("🌳 main"), "{out}");
        assert!(!out.contains("↑5") && !out.contains("+3"), "{out}");
    }

    #[test]
    fn rate_windows_are_optional_and_an_unknown_window_says_n_a() {
        let mut status = Status::default();
        assert_eq!(plain(&status, &[]), "CW n/a");
        status.seven_day = Some(RateWindow::default());
        assert_eq!(plain(&status, &[]), "CW n/a | 7D n/a");
        assert_eq!(plain(&status, &["context"]), "7D n/a");
        assert_eq!(plain(&status, &["usage_7d"]), "CW n/a");
        // A window that reports a real zero used is a known number, with its icon.
        status.seven_day = Some(RateWindow {
            used_percentage: Some(0.0),
            resets_at: None,
        });
        assert_eq!(plain(&status, &["context"]), "🔋 7D ░░░░░░░░░░ 0%");
    }

    /// REQ-9: a window whose percentage is unavailable says `n/a` with no battery icon and never
    /// `0%`; the reset countdown of an unknown week still follows it.
    #[test]
    fn an_unknown_window_says_n_a_without_a_battery_icon() {
        let unknown_ctx = Status::default();
        let line = plain(&unknown_ctx, &[]);
        assert_eq!(line, "CW n/a", "{line}");
        assert!(!line.contains('🪫') && !line.contains('🔋'), "{line}");
        // (A window with usage but no total is unknown too: `segments::context_percent`.)
        let week = Status {
            seven_day: Some(RateWindow {
                used_percentage: None,
                resets_at: None,
            }),
            ..Status::default()
        };
        let line = plain(&week, &["context"]);
        assert_eq!(line, "7D n/a", "{line}");
        assert!(
            !line.contains("0%") && !line.contains('🪫') && !line.contains('🔋'),
            "{line}"
        );
    }

    /// REQ-10: a window's percentage is shown from 0 to 100 and no further, at the edges too. The
    /// number is the used percentage given, clamped.
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
        assert_eq!(week(-5.0), "🔋 7D ░░░░░░░░░░ 0%");
        assert_eq!(week(0.0), "🔋 7D ░░░░░░░░░░ 0%");
        assert_eq!(week(1.0), "🔋 7D ░░░░░░░░░░ 1%");
        assert_eq!(week(99.0), "🪫 7D █████████░ 99%");
        assert_eq!(week(100.0), "🪫 7D ██████████ 100%");
        assert_eq!(week(101.0), "🪫 7D ██████████ 100%");
        assert_eq!(week(1_000_000.0), "🪫 7D ██████████ 100%");
        let ctx = |used: i64| {
            let status = Status {
                context: Some(used as f64),
                ..Status::default()
            };
            plain(&status, &["usage_7d"])
        };
        assert_eq!(ctx(0), "🔋 CW ░░░░░░░░░░ 0%");
        assert_eq!(ctx(100), "🪫 CW ██████████ 100%");
        assert_eq!(ctx(101), "🪫 CW ██████████ 100%");
        assert_eq!(ctx(5_000), "🪫 CW ██████████ 100%");
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
        failed.context = None;
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

    /// #2559 AC-1: the window icon's thresholds pinned at N/N+1 — the battery icon turns between
    /// 70 and 71 percent used. Plain output, so only the threshold decides.
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
        assert_eq!(at(70.0), "🔋 7D ███████░░░ 70%");
        assert_eq!(at(71.0), "🪫 7D ███████░░░ 71%");
        assert_eq!(at(9.0), "🔋 7D ░░░░░░░░░░ 9%");
        assert_eq!(at(10.0), "🔋 7D █░░░░░░░░░ 10%");
    }

    /// SPEC-20261122-statusline-where REQ-8: line 1 has no tag. The `session` option is still
    /// accepted in configs and changes nothing, and a status with nothing but the usage windows
    /// starts with them.
    #[test]
    fn line_one_has_no_tag_and_the_session_option_changes_nothing() {
        let tag = "\u{1F3F7}\u{FE0F}";
        for status in [Status::default(), full_status()] {
            let out = plain(&status, &[]);
            assert!(!out.contains(tag) && !out.contains("agent"), "{out}");
            assert_eq!(plain(&status, &[]), out);
            let on = DisplayConfig {
                segments: [("session".to_string(), true)].into_iter().collect(),
                ..Default::default()
            };
            assert_eq!(Renderer::new(&on, None).render(&status), out);
        }
        assert_eq!(plain(&Status::default(), &[]), "CW n/a");
    }

    #[test]
    fn forge_counts_show_zero_but_hide_missing_suppressed_or_disabled_data() {
        let mut status = full_status();
        status.github = GitHubCounts {
            available: true,
            ..Default::default()
        };
        let with_zero = plain(&status, &[]);
        assert!(with_zero.contains("📡 owner/repo, ⬇️ 0 🔀 0 | 🌳 main"));
        // #2559: icons, not the old letter labels.
        assert!(!with_zero.contains("I: ") && !with_zero.contains("PR: "));
        // A failed query says `n/a`, never a zero and never a hidden segment (REQ-3).
        status.github.available = false;
        let unknown = plain(&status, &[]);
        assert!(
            unknown.contains("📡 owner/repo, ⬇️ n/a 🔀 n/a | 🌳 main"),
            "{unknown}"
        );
        assert!(!unknown.contains("⬇️ 0"), "{unknown}");
        // A user switch keeps the counts hidden.
        status.github.available = true;
        status.github.suppressed = true;
        let suppressed = plain(&status, &[]);
        assert!(
            suppressed.contains("📡 owner/repo | 🌳 main"),
            "{suppressed}"
        );
        status.github.suppressed = false;
        assert!(plain(&status, &["github"]).contains("📡 owner/repo | 🌳 main"));
        assert!(!plain(&status, &["git_branch"]).contains("📡"));
        assert!(!plain(&status, &["git_status"]).contains("💾"));
        // The repository name unknown: the item and its counts are left out (REQ-1).
        status.repo = None;
        let nameless = plain(&status, &[]);
        assert!(
            !nameless.contains('📡') && !nameless.contains("⬇️"),
            "{nameless}"
        );
    }

    /// SPEC-20261122-statusline-flow REQ-7 and REQ-8, with the line order of
    /// SPEC-20261123-statusline-order-bars: the cards by stage are on the second line, between
    /// line 1 and the usage windows, behind the two versions. Line 1 and the windows do not
    /// change, the `cards` option hides the flow items and nothing else (the versions stay), and
    /// the flow does not depend on git: forge counts, repository and branch change line 1, never
    /// the flow line.
    #[test]
    fn the_flow_line_stands_between_line_one_and_the_usage_windows() {
        let mut status = full_status();
        let windows = FULL_WINDOWS;
        let without = plain(&status, &[]);
        assert_eq!(without.lines().count(), 3, "no ledger: the versions alone");
        assert_eq!(without.lines().nth(1), Some(FULL_VERSIONS));
        assert_eq!(
            without.lines().last().unwrap(),
            windows,
            "no flow: the windows close the screen"
        );
        status.flow = Some(example_flow());
        let with = plain(&status, &[]);
        let lines: Vec<&str> = with.lines().collect();
        assert_eq!(lines.len(), 3, "{with}");
        assert_eq!(Some(lines[0]), without.lines().next());
        assert_eq!(lines[1], format!("{FULL_VERSIONS} | {}", p0()));
        assert_eq!(Some(lines[2]), without.lines().nth(2));
        // The `cards` option hides the flow items and nothing else.
        assert_eq!(plain(&status, &["cards"]), without);
        // The flow comes from the ledger, not from git.
        let flow_line =
            |s: &Status, off: &[&str]| plain(s, off).lines().nth(1).unwrap().to_string();
        let with_versions = format!("{FULL_VERSIONS} | {}", p0());
        assert_eq!(flow_line(&status, &["github"]), with_versions);
        status.github.suppressed = true;
        assert_eq!(flow_line(&status, &[]), with_versions);
        status.repo = None;
        status.git = None;
        let bare = plain(&status, &["git_branch"]);
        assert_eq!(bare.lines().next(), Some("🧠 Opus high·t"));
        assert_eq!(
            bare.lines().nth(1),
            Some(with_versions.as_str()),
            "line 1 keeps the model"
        );
        // A ledger without a card has nothing to show: the flow items are left out, not zeros,
        // and the versions stay.
        status.flow = Some(Flow {
            split: true,
            ..Default::default()
        });
        // (Line 1 keeps the model, the versions stand second and the windows are last.)
        let empty = plain(&status, &[]);
        assert_eq!(empty.lines().count(), 3, "{empty}");
        assert_eq!(empty.lines().nth(1), Some(FULL_VERSIONS));
    }

    /// SPEC-20261122-statusline-where REQ-1: line 1 reads the model with its effort, the
    /// repository with the forge counts in one item (the comma right after the name), then the
    /// branch, then the pull request, items joined by ` | `. The flow line stands under it and
    /// the windows under that.
    #[test]
    fn line_one_reads_the_repository_with_the_forge_counts_then_the_branch_then_the_pr() {
        let line = |s: &Status, off: &[&str]| plain(s, off).lines().next().unwrap().to_string();
        let windows = FULL_WINDOWS;
        let mut status = full_status();
        status.flow = Some(example_flow());
        status.integration = Some("next".into());
        status.integration_tip = "b3ba2b3".into();
        status.pr = Some(state::PrInfo {
            number: 42,
            url: String::new(),
            review_state: "pending".into(),
            merge_request: false,
        });
        let full = line(&status, &[]);
        assert_eq!(
            full,
            "🧠 Opus high·t | 📡 owner/repo, ⬇️ 7 🔀 3 | 🅱️ next(🏷️ b3ba2b3) | 🌳 main | 🔗 #42 pending"
        );
        assert!(
            !full.replace("owner/repo", "").contains('/'),
            "no slash inside an item: {full}"
        );
        let at = |icon: &str| full.find(icon).unwrap();
        assert!(at("🧠") < at("📡") && at("📡") < at("⬇️") && at("⬇️") < at("🔀"));
        assert!(at("🔀") < at("🅱️") && at("🅱️") < at("🌳") && at("🌳") < at("🔗"));
        // The comma sits right after the repository name, inside the item: no separator
        // between the repository and the forge counts.
        assert!(full.contains("owner/repo, ⬇️ 7 🔀 3 | "), "{full}");
        // The versions are on the windows line, not on line 1.
        assert!(!full.contains("🤖") && !full.contains("🗿"), "{full}");
        // The flow line is line 2, behind the versions, and the windows alone are line 3.
        assert_eq!(
            plain(&status, &[]).lines().nth(1).unwrap(),
            format!("{FULL_VERSIONS} | {}", p0())
        );
        assert_eq!(plain(&status, &[]).lines().nth(2).unwrap(), windows);
        // Without the forge counts the repository name stands alone, with no comma.
        assert_eq!(
            line(&status, &["github"]),
            "🧠 Opus high·t | 📡 owner/repo | 🅱️ next(🏷️ b3ba2b3) | 🌳 main | 🔗 #42 pending"
        );
        // The `cards` switch hides the flow items and nothing else: the versions stay on line 2.
        assert_eq!(
            plain(&status, &["cards"]).lines().nth(1).unwrap(),
            FULL_VERSIONS
        );
        assert_eq!(plain(&status, &["cards"]).lines().nth(2).unwrap(), windows);
        // The pull request keeps the last place of line 1.
        assert_eq!(
            line(&status, &["pr"]),
            "🧠 Opus high·t | 📡 owner/repo, ⬇️ 7 🔀 3 | 🅱️ next(🏷️ b3ba2b3) | 🌳 main"
        );
    }

    /// The flow comes from the local ledger, not from the repository, so the flow line is shown
    /// without a repository name, without a branch and with `git_branch` off; the checked-out
    /// branch stands with the model on line 1 and the windows close the output.
    #[test]
    fn the_flow_line_does_not_need_the_repository_or_the_branch() {
        let mut status = full_status();
        status.flow = Some(example_flow());
        status.worktree = "wt".into();
        let windows = FULL_WINDOWS;
        status.repo = None;
        // With `git_branch` off the branch items are hidden and the host's worktree name is no
        // item of line 1: the model stands alone.
        let off = plain(&status, &["git_branch"]);
        assert_eq!(off.lines().next().unwrap(), "🧠 Opus high·t");
        let with_versions = format!("{FULL_VERSIONS} | {}", p0());
        assert_eq!(off.lines().nth(1).unwrap(), with_versions);
        assert_eq!(off.lines().nth(2).unwrap(), windows);
        // The full render: an unnamed repository leaves its item out (REQ-1), the checked-out
        // branch follows the model, the flow line comes next and the windows are last.
        let out = plain(&status, &[]);
        assert_eq!(out.lines().count(), 3, "{out}");
        assert_eq!(out.lines().next().unwrap(), "🧠 Opus high·t | 🌳 main");
        assert_eq!(out.lines().nth(1).unwrap(), with_versions);
        // No git: line 1 keeps the model, the flow line stays second.
        status.git = None;
        status.worktree.clear();
        let out = plain(&status, &[]);
        assert_eq!(out.lines().count(), 3, "{out}");
        assert_eq!(out.lines().next().unwrap(), "🧠 Opus high·t");
        assert_eq!(out.lines().nth(1).unwrap(), with_versions);
    }

    /// Color is never the only signal. Every number keeps its icon, neighbours in the flow take
    /// colors of their own, and the text is the plain text once the escapes are removed.
    #[test]
    fn the_flow_line_keeps_its_icons_under_every_color_profile() {
        let mut status = full_status();
        status.flow = Some(example_flow());
        let text = plain(&status, &[]);
        for name in ["mocha", "catppuccin-latte"] {
            let t = theme(name);
            for profile in [Profile::TrueColor, Profile::Ansi256] {
                let out = colored(name, profile, &status);
                assert_eq!(strip_sgr(&out), text, "{name} {profile:?}");
                let mine = format!("🎯 C-{EXAMPLE_CARD} {WORK}");
                for (hex, part) in [
                    (t.version, "🤖 2.1.293".to_string()),
                    (t.version, "🗿 2.0 -> 3.0".to_string()),
                    (t.cards_active, mine),
                    (t.cards_waiting, format!("{TODO} 150")),
                    (t.muted, "→".to_string()),
                    (t.cards_active, format!("{WORK} 14")),
                    (t.cards_posted, format!("{CHECK} 11")),
                    (t.branch, format!("{READY} 2")),
                    (t.cards_done, format!("{DONE} 19")),
                    (t.issues, format!("{BLOCKED} 3")),
                    (t.issues, format!("{QUESTION} 7")),
                ] {
                    assert!(out.contains(&profile.paint(hex, &part)), "{name} {part}");
                }
            }
            // The items of the flow, in their order on the line.
            let colors = [
                t.cards_waiting,
                t.cards_active,
                t.cards_posted,
                t.branch,
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
        );
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

    /// REQ-10 of SPEC-20261123-statusline-order-bars: `🌳` is the branch the HEAD file names, never
    /// the host's worktree name — whatever name the host gives (the same as the branch or not),
    /// the line is the same; the `worktree` option hides the item too, and so does `git_branch`.
    #[test]
    fn the_checked_out_item_is_the_branch_and_never_the_host_worktree_name() {
        let line1 = |s: &Status, off: &[&str]| plain(s, off).lines().next().unwrap().to_string();
        let mut status = full_status();
        let base = line1(&status, &[]);
        assert_eq!(base, "🧠 Opus high·t | 📡 owner/repo, ⬇️ 7 🔀 3 | 🌳 main");
        for name in ["host-name", "main", "wt"] {
            status.worktree = name.into();
            assert_eq!(line1(&status, &[]), base, "{name}");
            assert!(!plain(&status, &[]).contains("WT-c"));
        }
        status.worktree = "host-name".into();
        assert_eq!(
            line1(&status, &["worktree"]),
            "🧠 Opus high·t | 📡 owner/repo, ⬇️ 7 🔀 3"
        );
        assert_eq!(
            line1(&status, &["git_branch"]),
            "🧠 Opus high·t",
            "git_branch hides the repository, the integration item and the branch"
        );
    }

    /// #2756 AC-7–AC-9, reworked by the user instruction 2026-11-20 (D-1 of
    /// SPEC-20261120-statusline-layout): the cache segment is not in the layout (and since
    /// SPEC-20261122-statusline-limits REQ-11 its values are not collected), the worktree and
    /// the pull request stay on their lines, and the Claude Code version ends line 3.
    #[test]
    fn cache_timer_worktree_and_pr_segments() {
        let mut status = full_status();
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
            lines[0], "🧠 Opus high·t | 📡 owner/repo, ⬇️ 7 🔀 3 | 🌳 main | 🔗 #42 pending",
            "no cache segment: {}",
            lines[0]
        );
        // The two versions open the second line and the windows close the screen, each with its
        // bar and its number.
        assert_eq!(lines[1], FULL_VERSIONS, "{}", lines[1]);
        assert!(lines[2].starts_with("🔋 CW █████░░░░░ 50%"), "{}", lines[2]);
        assert!(lines[2].contains("7D ███████░░░ 75%"), "{}", lines[2]);
        assert_eq!(lines.len(), 3, "{out}");
        assert!(
            !out.contains('\x1b'),
            "plain output carries no link escape: {out:?}"
        );
        // The cache and its warm or cold markers never appear, with the `cache_hit` option
        // left enabled.
        let out = plain(&status, &[]);
        assert!(
            !out.contains('♻') && !out.contains('⏱') && !out.contains('🧊'),
            "{out}"
        );
        status.pr.as_mut().unwrap().review_state.clear();
        // Without the review state the PR still closes line 1, and the `worktree` option hides
        // the checked-out item.
        status.worktree.clear();
        assert!(plain(&status, &[]).contains("| 🔗 #42\n"));
        assert!(!plain(&status, &["pr"]).contains('🔗'));
        assert!(!plain(&status, &["worktree"]).contains('🌳'));
        status.pr = None;
        status.worktree.clear();
        assert_eq!(
            plain(&status, &[]).lines().next(),
            Some("🧠 Opus high·t | 📡 owner/repo, ⬇️ 7 🔀 3 | 🌳 main"),
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
        // The two versions stay on the second line without the weekly window, and the windows
        // line holds the context window alone.
        assert_eq!(output.lines().nth(1), Some(FULL_VERSIONS));
        assert_eq!(output.lines().nth(2), Some("🔋 CW █████░░░░░ 50%"));
        // And without any window the versions are the whole of the screen's second line.
        let output = plain(&full_status(), &["usage_5h", "usage_7d", "context"]);
        assert_eq!(output.lines().count(), 2);
        assert_eq!(output.lines().nth(1), Some("🤖 2.1.293 | 🗿 2.0 -> 3.0"));
        // With the repository and the branch hidden, line 1 keeps the model.
        let output = plain(&full_status(), &["git_branch", "git_status"]);
        assert_eq!(output.lines().count(), 3);
        assert_eq!(output.lines().next(), Some("🧠 Opus high·t"), "{output}");
        assert!(
            output.lines().nth(2).unwrap().starts_with("🔋 CW "),
            "{output}"
        );
    }

    #[test]
    fn context_renders_without_the_handoff_hints() {
        // The soft and the hard band of the old hints: 90 and 95 percent of a 200k window.
        let mut status = Status {
            context: Some(90.0),
            ..Default::default()
        };
        let off = ["session"];
        // REQ-7: the `(⚠️ /clear)` and `(🛑/clear!)` hints are gone at every band — the window
        // and the percentage used stand alone.
        assert_eq!(plain(&status, &off), "🪫 CW █████████░ 90%");
        assert_eq!(plain(&status, &[]), "🪫 CW █████████░ 90%");
        status.context = Some(95.0);
        assert_eq!(plain(&status, &off), "🪫 CW █████████░ 95%");
        assert_eq!(plain(&status, &["context", "session"]), "MoAI");
        assert_eq!(plain(&status, &["context"]), "MoAI");
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

    fn segment_colors(t: &Theme) -> [&'static str; 9] {
        [
            t.model, t.effort, t.cache, t.version, t.repo, t.issues, t.prs, t.branch, t.distance,
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
                    (t.effort, "high·t"),
                    // D-1: the cache segment is not in the layout, so nothing paints
                    // `theme.cache` any more; the palette entry itself stays distinct (below).
                    (t.version, "🤖 2.1.293"),
                    (t.version, "🗿 2.0 -> 3.0"),
                    (t.repo, "📡 owner/repo,"),
                    (t.issues, "⬇️ 7"),
                    (t.prs, "🔀 3"),
                    (t.branch, "🌳 main"),
                ] {
                    assert!(
                        out.contains(&profile.paint(hex, text)),
                        "{name} {profile:?}: {text:?} not painted {hex} in {out:?}"
                    );
                }
                // Escapes removed, the text is the NO_COLOR text (AC-2, AC-3 widths).
                assert_eq!(strip_sgr(&out), plain(&status, &[]));

                status.context = Some(90.0);
                let soft = colored(name, profile, &status);
                assert!(!soft.contains("(⚠️ /clear)"), "{soft:?}");
                status.context = Some(95.0);
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

/// The forms of the flow line, from the widest to the narrowest
/// (SPEC-20261122-statusline-flow AC-5 and AC-7, with the one icon set of
/// SPEC-20261122-statusline-limits-left REQ-6 and REQ-7): no letter codes, a zero count hidden,
/// the items dropping from the left of the flow, and one space after every icon. Since
/// SPEC-20261123-statusline-order-bars the two versions open the line and my card closes it.
#[cfg(test)]
mod flow_line_tests {
    use super::*;
    use crate::statusline::icons::{BLOCKED, CHECK, DONE, PAUSED, QUESTION, READY, TODO, WORK};
    use crate::statusline::state::{MyCard, Stage, Version};

    /// The number of my card in the fixtures.
    const CARD: i64 = 12;

    fn full() -> Flow {
        Flow {
            draft: 43,
            ready: 150,
            // The work item draws the working and the stale cards together: 8 + 6.
            working: 8,
            stale: 6,
            submitted: 11,
            audited: 2,
            landed: 1,
            blocked: 3,
            hold: 7,
            paused: 0,
            done_today: 19,
            mine: Some(MyCard {
                card: CARD,
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

    /// A split flow with nothing in it.
    fn empty() -> Flow {
        Flow {
            split: true,
            ..Default::default()
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

    /// The Claude Code version and the MoAI version of the SPEC's example.
    fn with_versions(flow: Option<Flow>) -> Status {
        Status {
            claude_version: "2.1.330".into(),
            version: Some(Version {
                current: "4.2648.20".into(),
                latest: None,
            }),
            flow,
            ..Default::default()
        }
    }

    /// The flow of the SPEC's example: the user's numbers, my card in the working stage.
    fn spec_flow() -> Flow {
        Flow {
            ready: 176,
            working: 6,
            submitted: 13,
            audited: 8,
            blocked: 7,
            hold: 9,
            done_today: 6,
            mine: Some(MyCard {
                card: CARD,
                stage: Stage::Working,
            }),
            split: true,
            ..Default::default()
        }
    }

    /// The screen of `status` with the session tag, the context window and the options in
    /// `off` switched off, so that the flow line (and the versions) are all that is left.
    fn screen(
        status: &Status,
        off: &[&str],
        budget: Option<usize>,
        profile: Option<Profile>,
    ) -> String {
        let config = DisplayConfig {
            segments: ["session", "context"]
                .iter()
                .chain(off)
                .map(|k| ((*k).to_string(), false))
                .collect(),
            ..Default::default()
        };
        Renderer::new(&config, profile)
            .with_budget(budget)
            .render(status)
    }

    /// The forms F0 to F7 of the Layouts table of the SPEC: the two versions, the flow with its
    /// leftmost item dropping, what stands beside it, and my card at the end.
    fn spec_forms() -> Vec<String> {
        let mine = format!("🎯 C-{CARD} {WORK}");
        [
            format!("🤖 2.1.330 | 🗿 4.2648.20 | {TODO} 176 → {WORK} 6 → {CHECK} 13 → {READY} 8 → {DONE} 6"),
            format!("🗿 4.2648.20 | {TODO} 176 → {WORK} 6 → {CHECK} 13 → {READY} 8 → {DONE} 6"),
            format!("{TODO} 176 → {WORK} 6 → {CHECK} 13 → {READY} 8 → {DONE} 6"),
            format!("{WORK} 6 → {CHECK} 13 → {READY} 8 → {DONE} 6"),
            format!("{CHECK} 13 → {READY} 8 → {DONE} 6"),
            format!("{READY} 8 → {DONE} 6"),
            format!("{DONE} 6"),
            String::new(),
        ]
        .iter()
        .map(|head| {
            if head.is_empty() {
                format!("{BLOCKED} 7 {QUESTION} 9 | {mine}")
            } else {
                format!("{head} | {BLOCKED} 7 {QUESTION} 9 | {mine}")
            }
        })
        .collect()
    }

    /// The forms of [`no_card`] with their widths: the ready item counts the audited and the
    /// landed cards together (2 + 1), the 43 drafts and the 22 ended leases are not on the line,
    /// and each narrower form drops the leftmost item of the flow — the to-do item first — while
    /// the blocked and the question items stay. The icons are the table's.
    fn q() -> [(String, usize); 6] {
        [
            (
                format!(
                    "{TODO} 150 → {WORK} 14 → {CHECK} 11 → {READY} 3 → {DONE} 19 | {BLOCKED} 3 {QUESTION} 7"
                ),
                49,
            ),
            (
                format!("{WORK} 14 → {CHECK} 11 → {READY} 3 → {DONE} 19 | {BLOCKED} 3 {QUESTION} 7"),
                40,
            ),
            (
                format!("{CHECK} 11 → {READY} 3 → {DONE} 19 | {BLOCKED} 3 {QUESTION} 7"),
                32,
            ),
            (
                format!("{READY} 3 → {DONE} 19 | {BLOCKED} 3 {QUESTION} 7"),
                24,
            ),
            (format!("{DONE} 19 | {BLOCKED} 3 {QUESTION} 7"), 17),
            (format!("{BLOCKED} 3 {QUESTION} 7"), 9),
        ]
    }

    /// My card as the line writes it: `🎯 C-<n> <stage icon>`.
    fn mine(stage: Stage) -> String {
        format!("🎯 C-{CARD} {}", stage.icon())
    }

    /// The forms of [`full`]: my card closes every form of [`q`].
    fn with_my_card() -> Vec<String> {
        q().iter()
            .map(|(form, _)| format!("{form} | {}", mine(Stage::Working)))
            .collect()
    }

    /// Every form is exactly its string, with its width, and changes at its boundary: a budget
    /// equal to the width keeps the form, one less drops to the next. The forms of the SPEC's
    /// Layouts table start with the two versions, which drop first (the Claude Code version,
    /// then the MoAI version), then the leftmost item of the flow; my card and the blocked,
    /// paused and question items stay in every form.
    #[test]
    fn every_form_is_exactly_its_string_and_changes_at_its_boundary() {
        for (form, cols) in q() {
            assert_eq!(width::display_width(&form), cols, "{form}");
        }
        let without: Vec<String> = q().iter().map(|(form, _)| form.clone()).collect();
        assert_eq!(
            with_my_card()[0],
            format!(
                "{TODO} 150 → {WORK} 14 → {CHECK} 11 → {READY} 3 → {DONE} 19 | {BLOCKED} 3 {QUESTION} 7 | 🎯 C-{CARD} {WORK}"
            )
        );
        // The Layouts table of the SPEC, F0 to F7, with the versions in front.
        let spec = spec_forms();
        assert_eq!(spec.len(), 8);
        assert_eq!(
            spec[0],
            format!(
                "🤖 2.1.330 | 🗿 4.2648.20 | {TODO} 176 → {WORK} 6 → {CHECK} 13 → {READY} 8 → {DONE} 6 | {BLOCKED} 7 {QUESTION} 9 | 🎯 C-{CARD} {WORK}"
            )
        );
        assert_eq!(
            spec[7],
            format!("{BLOCKED} 7 {QUESTION} 9 | 🎯 C-{CARD} {WORK}")
        );
        let status = with_versions(Some(spec_flow()));
        for (i, text) in spec.iter().enumerate() {
            let cols = width::display_width(text);
            assert_eq!(
                &screen(&status, &[], Some(cols), None),
                text,
                "F{i} at budget {cols}"
            );
            if i == 0 {
                assert_eq!(&screen(&status, &[], Some(cols + 40), None), text);
                assert_eq!(&screen(&status, &[], None, None), text, "no budget");
            }
            if let Some(next) = spec.get(i + 1) {
                assert!(width::display_width(next) < cols, "{next} is not narrower");
                assert_eq!(
                    &screen(&status, &[], Some(cols - 1), None),
                    next,
                    "F{i} drops to F{} at budget {}",
                    i + 1,
                    cols - 1
                );
            }
        }
        // Below the narrowest form the narrowest is written.
        for budget in [width::display_width(&spec[7]) - 1, 5, 0] {
            assert_eq!(&screen(&status, &[], Some(budget), None), &spec[7]);
        }
        for (flow, forms) in [(full(), with_my_card()), (no_card(), without)] {
            for (i, text) in forms.iter().enumerate() {
                let cols = width::display_width(text);
                assert_eq!(&plain(flow, Some(cols)), text, "budget {cols}");
                if i == 0 {
                    assert_eq!(&plain(flow, Some(cols + 40)), text, "{cols} + 40");
                }
                if let Some(next) = forms.get(i + 1) {
                    assert!(width::display_width(next) < cols, "{next} is not narrower");
                    assert_eq!(&plain(flow, Some(cols - 1)), next, "budget {}", cols - 1);
                }
            }
        }
    }

    #[test]
    fn no_budget_writes_the_widest_form_and_a_budget_below_the_narrowest_the_narrowest() {
        assert_eq!(plain(full(), None), with_my_card()[0]);
        assert_eq!(plain(no_card(), None), q()[0].0);
        for budget in [15, 5, 0] {
            assert_eq!(plain(full(), Some(budget)), with_my_card()[5], "{budget}");
        }
        assert_eq!(plain(no_card(), Some(0)), q()[5].0);
    }

    /// A count of zero hides its item — with its arrow or its separator — and never prints `0`;
    /// a line with nothing to show is left out whole.
    #[test]
    fn a_zero_count_hides_its_item() {
        let line = |flow: Flow| plain(flow, None);
        assert_eq!(
            line(Flow {
                ready: 0,
                ..no_card()
            }),
            format!("{WORK} 14 → {CHECK} 11 → {READY} 3 → {DONE} 19 | {BLOCKED} 3 {QUESTION} 7")
        );
        assert_eq!(
            line(Flow {
                submitted: 0,
                ..no_card()
            }),
            format!("{TODO} 150 → {WORK} 14 → {READY} 3 → {DONE} 19 | {BLOCKED} 3 {QUESTION} 7")
        );
        assert_eq!(
            line(Flow {
                working: 0,
                // The work item is the sum of the working and the stale cards: both are 0 to
                // hide it.
                stale: 0,
                done_today: 0,
                ..no_card()
            }),
            format!("{TODO} 150 → {CHECK} 11 → {READY} 3 | {BLOCKED} 3 {QUESTION} 7")
        );
        // The ready item is the audited and the landed cards together: either one shows it.
        for (audited, landed, want) in [
            (0, 1, Some(format!("{READY} 1"))),
            (2, 0, Some(format!("{READY} 2"))),
            (0, 0, None),
        ] {
            let out = line(Flow {
                audited,
                landed,
                ..no_card()
            });
            match want {
                Some(item) => assert!(out.contains(&format!("→ {item} →")), "{out}"),
                None => assert_eq!(
                    out,
                    format!("{TODO} 150 → {WORK} 14 → {CHECK} 11 → {DONE} 19 | {BLOCKED} 3 {QUESTION} 7")
                ),
            }
        }
        // Beside the flow: each of the three on its own, and no group without them.
        let chain = format!("{TODO} 150 → {WORK} 14 → {CHECK} 11 → {READY} 3 → {DONE} 19");
        let with = |blocked, paused, hold| Flow {
            blocked,
            paused,
            hold,
            ..no_card()
        };
        assert_eq!(
            line(with(0, 0, 7)),
            format!("{chain} | {QUESTION} 7"),
            "blocked hidden"
        );
        assert_eq!(
            line(with(3, 0, 0)),
            format!("{chain} | {BLOCKED} 3"),
            "question hidden"
        );
        assert_eq!(
            line(with(0, 5, 0)),
            format!("{chain} | {PAUSED} 5"),
            "paused on its own"
        );
        assert_eq!(
            line(with(3, 5, 7)),
            format!("{chain} | {BLOCKED} 3 {PAUSED} 5 {QUESTION} 7"),
            "all three, in the order of the table"
        );
        assert_eq!(line(with(0, 0, 0)), chain);
        // Nothing but cards waiting their turn: the one item.
        assert_eq!(
            line(Flow {
                ready: 169,
                ..empty()
            }),
            format!("{TODO} 169")
        );
        // No flow at all: my card and what stands beside the flow.
        assert_eq!(
            line(Flow {
                ready: 0,
                working: 0,
                stale: 0,
                submitted: 0,
                audited: 0,
                landed: 0,
                done_today: 0,
                ..full()
            }),
            format!("{BLOCKED} 3 {QUESTION} 7 | {}", mine(Stage::Working))
        );
        // Nothing at all: no flow line, so the renderer's own fallback is the whole output.
        assert_eq!(line(empty()), "MoAI");
        for out in [line(full()), line(no_card())] {
            assert!(!out.contains(" 0"), "a zero is printed: {out}");
        }
    }

    /// The drafts are not on the line: a ledger that holds nothing else has no flow line, and
    /// their number never reaches another item. A card whose lease ended without a hand-back is
    /// drawn with the working ones (SPEC-20261124-statusline-counts-truth REQ-7), so a ledger
    /// with only such cards has a work item, and the item is the sum of the two counts.
    #[test]
    fn drafts_are_not_on_the_line_and_a_stale_card_is_drawn_with_the_working_ones() {
        let only_drafts = Flow {
            draft: 43,
            ..empty()
        };
        assert_eq!(plain(only_drafts, None), "MoAI");
        assert!(!plain(full(), None).contains("43"));
        let only_stale = Flow {
            stale: 22,
            ..empty()
        };
        assert_eq!(plain(only_stale, None), format!("{WORK} 22"));
        let both = Flow {
            working: 3,
            stale: 4,
            ..empty()
        };
        assert_eq!(plain(both, None), format!("{WORK} 7"));
        // Only my card, and its lease has ended: it is still the card in hand, and counted.
        let mine_only = Flow {
            mine: Some(MyCard {
                card: CARD,
                stage: Stage::Stale,
            }),
            stale: 1,
            ..empty()
        };
        assert_eq!(
            plain(mine_only, None),
            format!("{WORK} 1 | 🎯 C-{CARD} {WORK}")
        );
    }

    /// My card is `🎯 C-<n> <stage icon>` with the icon of its stage from the one icon set, for
    /// each of the nine stages, with no other group when the ledger holds nothing else.
    #[test]
    fn my_card_shows_the_icon_of_each_stage() {
        for (stage, icon) in [
            (Stage::Draft, TODO),
            (Stage::Ready, TODO),
            (Stage::Working, WORK),
            (Stage::Stale, WORK),
            (Stage::Submitted, CHECK),
            (Stage::Audited, READY),
            (Stage::Landed, READY),
            (Stage::Blocked, BLOCKED),
            (Stage::Hold, QUESTION),
        ] {
            let flow = Flow {
                mine: Some(MyCard { card: CARD, stage }),
                ..empty()
            };
            assert_eq!(
                plain(flow, None),
                format!("🎯 C-{CARD} {icon}"),
                "{stage:?}"
            );
        }
    }

    /// The unsplit flow (the notes could not be read) is one form of the two counts that are
    /// known — the cards waiting their turn and the ones in progress, without an arrow — never a
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
        for budget in [None, Some(200), Some(13), Some(0)] {
            assert_eq!(
                plain(unsplit, budget),
                format!("{TODO} 150 {WORK} 15"),
                "{budget:?}"
            );
        }
        let no_ready = Flow {
            ready: 0,
            ..unsplit
        };
        assert_eq!(plain(no_ready, None), format!("{WORK} 15"));
        let neither = Flow {
            working: 0,
            ..no_ready
        };
        assert_eq!(plain(neither, None), "MoAI");
    }

    /// SPEC-20261123-statusline-order-bars REQ-1: the line is the Claude Code version, the
    /// MoAI version, the flow, the blocked, paused and question items and my card, in that order. Without a flow it holds the two
    /// versions alone, without a version the flow alone, and the `claude_version`,
    /// `moai_version` and `cards` options each hide their own items and nothing else.
    #[test]
    fn the_versions_open_the_flow_line_and_my_card_closes_it() {
        let widest = format!(
            "🤖 2.1.330 | 🗿 4.2648.20 | {TODO} 176 → {WORK} 6 → {CHECK} 13 → {READY} 8 → {DONE} 6 | {BLOCKED} 7 {QUESTION} 9 | 🎯 C-{CARD} {WORK}"
        );
        let line = |status: &Status, off: &[&str]| screen(status, off, None, None);
        let both = with_versions(Some(spec_flow()));
        assert_eq!(line(&both, &[]), widest);
        // The options hide their own item and its separator, nothing else.
        assert_eq!(
            line(&both, &["claude_version"]),
            widest.replacen("🤖 2.1.330 | ", "", 1)
        );
        assert_eq!(
            line(&both, &["moai_version"]),
            widest.replacen("🗿 4.2648.20 | ", "", 1)
        );
        assert_eq!(
            line(&both, &["claude_version", "moai_version"]),
            spec_forms()[2]
        );
        assert_eq!(line(&both, &["cards"]), "🤖 2.1.330 | 🗿 4.2648.20");
        // No flow to show: the two versions alone, and the narrower the budget the fewer.
        let versions = with_versions(None);
        assert_eq!(line(&versions, &[]), "🤖 2.1.330 | 🗿 4.2648.20");
        let wide = width::display_width("🤖 2.1.330 | 🗿 4.2648.20");
        assert_eq!(
            screen(&versions, &[], Some(wide), None),
            "🤖 2.1.330 | 🗿 4.2648.20"
        );
        for budget in [wide - 1, 5, 0] {
            assert_eq!(
                screen(&versions, &[], Some(budget), None),
                "🗿 4.2648.20",
                "budget {budget}"
            );
        }
        // A split flow with nothing in it shows the versions alone too.
        assert_eq!(
            line(&with_versions(Some(empty())), &[]),
            "🤖 2.1.330 | 🗿 4.2648.20"
        );
        // No version: the flow alone, as it was.
        assert_eq!(
            line(
                &with_versions(Some(spec_flow())),
                &["claude_version", "moai_version"]
            ),
            spec_forms()[2]
        );
        let no_versions = Status {
            flow: Some(spec_flow()),
            ..Default::default()
        };
        assert_eq!(line(&no_versions, &[]), spec_forms()[2]);
        // Neither versions nor flow: nothing on the line, the fallback is the whole output.
        assert_eq!(line(&Status::default(), &[]), "MoAI");
        // The unsplit flow (the notes could not be read) keeps its one form and gets the version
        // steps: the versions in front, the to-do and work items without an arrow.
        let unsplit = Flow {
            ready: 150,
            working: 15,
            split: false,
            ..Default::default()
        };
        let status = with_versions(Some(unsplit));
        assert_eq!(
            line(&status, &[]),
            format!("🤖 2.1.330 | 🗿 4.2648.20 | {TODO} 150 {WORK} 15")
        );
        assert_eq!(
            screen(
                &status,
                &[],
                Some(width::display_width(&format!(
                    "🗿 4.2648.20 | {TODO} 150 {WORK} 15"
                ))),
                None
            ),
            format!("🗿 4.2648.20 | {TODO} 150 {WORK} 15")
        );
        assert_eq!(
            screen(
                &status,
                &[],
                Some(width::display_width(&format!("{TODO} 150 {WORK} 15"))),
                None
            ),
            format!("{TODO} 150 {WORK} 15")
        );
        assert_eq!(
            screen(&status, &[], Some(0), None),
            format!("{TODO} 150 {WORK} 15")
        );
        // The update arrow names the template, then the build, and drops with the MoAI version.
        let arrow = Status {
            claude_version: "2.1.330".into(),
            version: Some(Version {
                current: "4.2600.15".into(),
                latest: Some("4.2648.20".into()),
            }),
            flow: Some(spec_flow()),
            ..Default::default()
        };
        assert!(line(&arrow, &[]).starts_with(&format!(
            "🤖 2.1.330 | 🗿 4.2600.15 -> 4.2648.20 | {TODO} 176"
        )));
        assert_eq!(
            screen(
                &arrow,
                &[],
                Some(width::display_width(&spec_forms()[2])),
                None
            ),
            spec_forms()[2]
        );
    }

    /// The colored line carries the same text once the escapes are removed, at every form.
    #[test]
    fn the_colored_line_has_the_plain_text_at_every_form() {
        let status = with_versions(Some(spec_flow()));
        for budget in [None, Some(80), Some(70), Some(60), Some(30), Some(10)] {
            let text = screen(&status, &[], budget, None);
            for profile in [Profile::TrueColor, Profile::Ansi256] {
                let out = screen(&status, &[], budget, Some(profile));
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
        for flow in [full(), no_card()] {
            for budget in [None, Some(45), Some(30), Some(10)] {
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

    /// Every form the line can print — each flow below at every budget from 0 to 95 columns and
    /// without one, with and without the two versions in front — has exactly one space after
    /// every icon, then a number, the card, an arrow or a separator; holds no icon outside the
    /// table and none that left the set; and carries no letter code. Together the forms print
    /// each of the ten icons of the flow line.
    #[test]
    fn every_rendered_form_has_one_space_after_every_icon() {
        let mut flows = vec![
            full(),
            no_card(),
            Flow {
                split: false,
                ..no_card()
            },
            Flow {
                ready: 169,
                ..empty()
            },
        ];
        for stage in [
            Stage::Draft,
            Stage::Ready,
            Stage::Working,
            Stage::Stale,
            Stage::Submitted,
            Stage::Audited,
            Stage::Landed,
            Stage::Blocked,
            Stage::Hold,
        ] {
            flows.push(Flow {
                mine: Some(MyCard { card: CARD, stage }),
                ..full()
            });
        }
        let mut printed: Vec<&str> = Vec::new();
        let mut forms: Vec<String> = Vec::new();
        for flow in flows {
            for budget in std::iter::once(None).chain((0..=95).map(Some)) {
                // The flow line alone, and with the two versions in front of it.
                for text in [
                    plain(flow, budget),
                    screen(&with_versions(Some(flow)), &[], budget, None),
                ] {
                    if !forms.contains(&text) {
                        forms.push(text);
                    }
                }
            }
        }
        assert!(forms.len() >= 30, "{} forms", forms.len());
        for text in &forms {
            for icon in width::ICONS {
                for (at, _) in text.match_indices(icon) {
                    if !printed.contains(&icon) {
                        printed.push(icon);
                    }
                    let after = &text[at + icon.len()..];
                    if after.is_empty() {
                        continue;
                    }
                    let mut rest = after.chars();
                    assert_eq!(rest.next(), Some(' '), "{icon} in `{text}`");
                    let next = rest.next();
                    assert!(
                        next.is_some_and(|c| c.is_ascii_digit() || matches!(c, 'C' | '|' | '→')),
                        "{icon} is followed by {next:?} in `{text}`"
                    );
                }
            }
            // Nothing but the icons of the table, the arrow and plain text.
            let mut bare = text.clone();
            for icon in width::ICONS {
                bare = bare.replace(icon, "");
            }
            assert!(
                bare.chars().all(|c| c.is_ascii() || c == '→'),
                "an icon outside the table in `{text}`"
            );
            assert!(!text.contains("  "), "two spaces in `{text}`");
            // No letter code: the only letter is the `C` of my card.
            assert!(
                bare.chars().all(|c| !c.is_ascii_alphabetic() || c == 'C'),
                "a letter code in `{text}`"
            );
        }
        printed.sort_unstable();
        let mut flow_icons = vec![
            "🎯", TODO, WORK, CHECK, READY, DONE, BLOCKED, QUESTION, "🤖", "🗿",
        ];
        flow_icons.sort_unstable();
        assert_eq!(printed, flow_icons);
    }
}

/// SPEC-20261123-statusline-order-bars AC-15 (REQ-8 to REQ-12): the first line in each form of
/// the Layouts section, exactly, changing at its width boundary. The integration branch of the
/// fixtures is `v2026.11.0`, and its local ref is at `b3ba2b3`.
#[cfg(test)]
mod line_one_form_tests {
    use super::*;
    use crate::statusline::segments::GitStatus;
    use crate::statusline::state::{GitHubCounts, PrInfo};

    /// L0 to L4 of the Layouts section, with one space after every icon
    /// (SPEC-20261122-statusline-limits-left REQ-5), and their display widths.
    const L: [(&str, usize); 5] = [
        (
            "🧠 Sonnet 5.5 high·t | 📡 modu-ai/moai-adk, ⬇️ 1 🔀 0 | 🅱️ v2026.11.0(🏷️ b3ba2b3) | 🌳 spec/ready-mutant-line-20261123 | 🔗 #1234 pending",
            137,
        ),
        (
            "🧠 Sonnet 5.5 high·t | 📡 modu-ai/moai-adk | 🅱️ v2026.11.0(🏷️ b3ba2b3) | 🌳 spec/ready-mutant-line-20261123 | 🔗 #1234 pending",
            126,
        ),
        (
            "🧠 Sonnet high | 📡 moai-adk | 🅱️ v2026.11.0(🏷️ b3ba2b3) | 🌳 spec/ready-mutant-line-20261123 | 🔗 #1234",
            104,
        ),
        (
            "🧠 Sonnet | 🅱️ v2026.11.0(🏷️ b3ba2b3) | 🌳 spec/ready-mutant-line-20261123 | 🔗 #1234",
            85,
        ),
        (
            "🧠 Sonnet | 🅱️ v2026.11.0 | 🌳 spec/ready-mutant-line-20261123",
            62,
        ),
    ];
    /// The detached leader: no `🌳`, and no request either (the fixture has none).
    const DETACHED: (&str, usize) = (
        "🧠 Sonnet 5.5 high·t | 📡 modu-ai/moai-adk, ⬇️ 1 🔀 0 | 🅱️ v2026.11.0(🏷️ b3ba2b3)",
        81,
    );
    /// The integration ref cannot be read: `n/a`, never a guessed commit.
    const NO_REF: (&str, usize) = (
        "🧠 Sonnet 5.5 high·t | 📡 modu-ai/moai-adk, ⬇️ 1 🔀 0 | 🅱️ v2026.11.0(🏷️ n/a) | 🌳 spec/ready-mutant-line-20261123",
        114,
    );
    /// Neither the profile nor the remote names an integration branch: no `🅱️` item.
    const UNKNOWN: (&str, usize) = (
        "🧠 Sonnet 5.5 high·t | 📡 modu-ai/moai-adk, ⬇️ 1 🔀 0 | 🌳 main",
        63,
    );
    /// The narrowest form of the detached leader: the model and the integration name.
    const DETACHED_NARROWEST: &str = "🧠 Sonnet | 🅱️ v2026.11.0";

    fn pull_request(number: i64, state: &str, merge_request: bool) -> Option<PrInfo> {
        Some(PrInfo {
            number,
            url: String::new(),
            review_state: state.into(),
            merge_request,
        })
    }

    /// A lane branch checked out in a linked worktree, three commits ahead of the integration
    /// branch with two changed files (none of that is on the line), and a pull request.
    fn lane_session() -> Status {
        Status {
            model: "Sonnet 5.5".into(),
            effort: Some("high".into()),
            thinking: Some(true),
            repo: Some(("modu-ai".into(), "moai-adk".into())),
            github: GitHubCounts {
                available: true,
                open_issues: 1,
                open_prs: 0,
                ..Default::default()
            },
            worktree: "statusline-flow".into(),
            git: Some(GitStatus {
                branch: "spec/ready-mutant-line-20261123".into(),
                ahead: 3,
                modified: 1,
                untracked: 1,
                ..Default::default()
            }),
            integration: Some("v2026.11.0".into()),
            integration_tip: "b3ba2b3".into(),
            pr: pull_request(1234, "pending", false),
            ..Default::default()
        }
    }

    /// The leader: HEAD detached at another commit than the integration tip, no request.
    fn detached_leader() -> Status {
        Status {
            git: Some(GitStatus {
                branch: String::new(),
                head_sha: "0123456".into(),
                untracked: 3,
                ..Default::default()
            }),
            pr: None,
            ..lane_session()
        }
    }

    /// The first line alone; the context window is switched off so no other line has text.
    fn first_line(status: &Status, budget: Option<usize>, profile: Option<Profile>) -> String {
        first_line_with(status, budget, profile, &[])
    }

    fn first_line_with(
        status: &Status,
        budget: Option<usize>,
        profile: Option<Profile>,
        off: &[&str],
    ) -> String {
        let config = DisplayConfig {
            segments: std::iter::once("context")
                .chain(off.iter().copied())
                .map(|key| (key.to_string(), false))
                .collect(),
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
        for (text, cols) in L.iter().chain([&DETACHED, &NO_REF, &UNKNOWN]) {
            assert_eq!(width::display_width(text), *cols, "{text}");
        }
    }

    /// Every form is exactly the string of the Layouts section and changes at its boundary: a
    /// budget equal to the width keeps the form, one less drops to the next, and the narrowest
    /// form stays when nothing fits.
    #[test]
    fn every_form_is_exactly_the_string_of_the_layouts_section() {
        let status = lane_session();
        for (i, (text, cols)) in L.iter().enumerate() {
            assert_eq!(&plain(&status, Some(*cols)), text, "budget {cols}");
            if i == 0 {
                assert_eq!(&plain(&status, Some(*cols + 40)), text, "{cols} + 40");
            }
            if let Some((next, _)) = L.get(i + 1) {
                assert_eq!(
                    &plain(&status, Some(*cols - 1)),
                    next,
                    "budget {}",
                    cols - 1
                );
            }
        }
        assert_eq!(&plain(&status, Some(L[4].1 - 1)), L[4].0, "nothing fits");
    }

    /// No usable `COLUMNS` writes the widest form and a budget below the narrowest the narrowest.
    #[test]
    fn no_budget_writes_the_widest_form_and_a_small_budget_the_narrowest() {
        let status = lane_session();
        assert_eq!(plain(&status, None), L[0].0);
        for budget in [61, 10, 0] {
            assert_eq!(plain(&status, Some(budget)), L[4].0, "{budget}");
        }
    }

    /// The last three rows of the Layouts section: the detached leader (no `🌳`, no invented
    /// name), a missing integration ref (`n/a`) and an unknown integration branch (no `🅱️`).
    #[test]
    fn the_detached_missing_ref_and_unknown_integration_rows_are_the_strings_of_the_section() {
        let detached = detached_leader();
        assert_eq!(plain(&detached, None), DETACHED.0);
        assert_eq!(plain(&detached, Some(DETACHED.1)), DETACHED.0);
        assert_eq!(
            plain(&detached, Some(DETACHED.1 - 1)),
            "🧠 Sonnet 5.5 high·t | 📡 modu-ai/moai-adk | 🅱️ v2026.11.0(🏷️ b3ba2b3)"
        );
        assert_eq!(plain(&detached, Some(0)), DETACHED_NARROWEST);

        let mut no_ref = lane_session();
        no_ref.integration_tip.clear();
        no_ref.pr = None;
        assert_eq!(plain(&no_ref, None), NO_REF.0);
        assert_eq!(plain(&no_ref, Some(NO_REF.1)), NO_REF.0);
        assert_ne!(plain(&no_ref, Some(NO_REF.1 - 1)), NO_REF.0);
        // At the narrowest step the commit part goes whether it is a commit or `n/a`.
        assert_eq!(
            plain(&no_ref, Some(0)),
            "🧠 Sonnet | 🅱️ v2026.11.0 | 🌳 spec/ready-mutant-line-20261123"
        );

        let mut unknown = lane_session();
        unknown.integration = None;
        unknown.integration_tip.clear();
        unknown.pr = None;
        unknown.git = Some(GitStatus {
            branch: "main".into(),
            ..Default::default()
        });
        assert_eq!(plain(&unknown, None), UNKNOWN.0);
        assert_eq!(plain(&unknown, Some(UNKNOWN.1)), UNKNOWN.0);
        assert_eq!(plain(&unknown, Some(0)), "🧠 Sonnet | 🌳 main");
    }

    /// The colored line carries the same text once the escapes are removed, at every form.
    #[test]
    fn the_colored_first_line_has_the_plain_text_at_every_form() {
        for status in [lane_session(), detached_leader()] {
            for budget in [None, Some(126), Some(104), Some(85), Some(62), Some(0)] {
                let text = plain(&status, budget);
                for profile in [Profile::TrueColor, Profile::Ansi256] {
                    let colored = first_line(&status, budget, Some(profile));
                    assert!(colored.contains('\x1b'), "{budget:?} {profile:?}");
                    assert_eq!(strip(&colored), text, "{budget:?} {profile:?}");
                }
            }
        }
    }

    /// REQ-1 of SPEC-20261122-statusline-where and REQ-8: no tag item, no `🤔`, no versions on the
    /// first line; the tag icon appears only inside the parenthesis of the integration item.
    #[test]
    fn the_first_line_has_no_tag_item_no_old_effort_icon_and_no_versions() {
        let mut status = lane_session();
        status.claude_version = "2.1.330".into();
        status.version = Some(crate::statusline::state::Version {
            current: "4.2647.16".into(),
            latest: Some("4.2648.1".into()),
        });
        let line = plain(&status, None);
        assert_eq!(line, L[0].0);
        // The tag icon is the one of the integration item only, after an opening parenthesis.
        let tag = "\u{1F3F7}\u{FE0F}";
        for (at, _) in line.match_indices(tag) {
            assert!(line[..at].ends_with('('), "a tag item in {line}");
        }
        assert!(!line.starts_with(tag), "{line}");
        for gone in ["🤔", "🤖", "🗿", "Leader", "Agent", "agent"] {
            assert!(!line.contains(gone), "{gone} in {line}");
        }
    }

    /// REQ-11: no distance item, no ahead or behind mark and no uncommitted mark, whatever the
    /// numbers are, in every form of the line.
    #[test]
    fn the_first_line_has_no_distance_and_no_marks_in_any_form() {
        let status = lane_session();
        assert!(status.git.as_ref().is_some_and(|g| g.ahead == 3));
        for budget in [None, Some(126), Some(104), Some(85), Some(62), Some(0)] {
            let line = plain(&status, budget);
            for mark in ['↑', '↓', '+', '🧭'] {
                assert!(!line.contains(mark), "{mark} in {line}");
            }
        }
    }

    /// REQ-9: the tip is the status field and nothing else — a status whose HEAD says another
    /// commit still shows the integration tip, and a longer or shorter field is not padded.
    #[test]
    fn the_tip_shown_is_the_integration_tip_of_the_status() {
        let mut status = lane_session();
        status.git = Some(GitStatus {
            branch: "spec/ready-mutant-line-20261123".into(),
            head_sha: "ffffff0".into(),
            ..Default::default()
        });
        assert!(plain(&status, None).contains("(🏷️ b3ba2b3)"));
        status.integration_tip = "1234567".into();
        assert!(plain(&status, None).contains("🅱️ v2026.11.0(🏷️ 1234567) | "));
        assert!(!plain(&status, None).contains("ffffff0"));
    }

    /// REQ-9 and the options: the `git_branch` option hides the repository, the integration item
    /// and the checked-out item; the `worktree` option hides the checked-out item only.
    #[test]
    fn the_options_hide_the_items_they_name() {
        let status = lane_session();
        assert_eq!(
            first_line_with(&status, None, None, &["git_branch"]),
            "🧠 Sonnet 5.5 high·t | 🔗 #1234 pending"
        );
        assert_eq!(
            first_line_with(&status, None, None, &["worktree"]),
            "🧠 Sonnet 5.5 high·t | 📡 modu-ai/moai-adk, ⬇️ 1 🔀 0 | 🅱️ v2026.11.0(🏷️ b3ba2b3) | 🔗 #1234 pending"
        );
    }

    /// REQ-9 of SPEC-20261122-statusline-where: a merge request is `!<n>`, a pull request `#<n>`,
    /// the state is optional and a missing request leaves the item out with its separator.
    #[test]
    fn the_request_item_is_a_number_sign_or_an_exclamation_mark() {
        let mut status = lane_session();
        status.pr = pull_request(31, "pending", true);
        assert!(plain(&status, None)
            .ends_with(" | 🌳 spec/ready-mutant-line-20261123 | 🔗 !31 pending"));
        status.pr = pull_request(31, "", true);
        assert!(plain(&status, None).ends_with(" | 🔗 !31"), "no state");
        status.pr = pull_request(31, "", false);
        assert!(plain(&status, None).ends_with(" | 🔗 #31"), "no state");
        status.pr = None;
        assert!(plain(&status, None).ends_with(" | 🌳 spec/ready-mutant-line-20261123"));
    }

    /// REQ-10 of SPEC-20261122-statusline-where: the model comes from the model field and the
    /// effort from the effort and thinking fields only; a missing model leaves the item out, and
    /// an effort without thinking has no `·t`.
    #[test]
    fn the_model_and_the_effort_are_written_as_the_stdin_gives_them() {
        let mut status = lane_session();
        status.model = String::new();
        assert!(plain(&status, None).starts_with("📡 modu-ai/moai-adk, "));
        status.model = "Opus 4.1".into();
        status.effort = Some("high".into());
        assert!(plain(&status, None).starts_with("🧠 Opus 4.1 high·t | "));
        status.thinking = Some(false);
        assert!(plain(&status, None).starts_with("🧠 Opus 4.1 high | "));
        status.effort = Some(String::new());
        status.thinking = None;
        assert!(plain(&status, None).starts_with("🧠 Opus 4.1 | "));
    }

    /// REQ-8: an item without its input is left out with its separator, and the first line is
    /// left out when all of them are.
    #[test]
    fn an_item_without_its_input_is_left_out_with_its_separator() {
        let mut status = lane_session();
        status.repo = None;
        status.pr = None;
        assert_eq!(
            plain(&status, None),
            "🧠 Sonnet 5.5 high·t | 🅱️ v2026.11.0(🏷️ b3ba2b3) | 🌳 spec/ready-mutant-line-20261123"
        );
        status.git = Some(GitStatus::default());
        assert_eq!(
            plain(&status, None),
            "🧠 Sonnet 5.5 high·t | 🅱️ v2026.11.0(🏷️ b3ba2b3)"
        );
        assert_eq!(plain(&Status::default(), None), "MoAI");
    }

    /// REQ-9: the integration name is written whole, however long a branch name may be.
    #[test]
    fn the_integration_name_is_written_whole() {
        let mut status = lane_session();
        let name = "release/2026.11.0-candidate-with-a-long-name";
        status.integration = Some(name.into());
        let line = plain(&status, None);
        assert!(
            line.contains(&format!("🅱️ {name}(🏷️ b3ba2b3) | ")),
            "{line}"
        );
    }

    /// REQ-10 of SPEC-20261122-statusline-where, and REQ-5 of
    /// SPEC-20261122-statusline-limits-left: every icon of the first line is followed by exactly
    /// one space, `⬇️`, `🅱️` and `🏷️` too.
    #[test]
    fn every_icon_of_the_first_line_is_followed_by_one_space() {
        for (status, expected) in [(lane_session(), 8), (detached_leader(), 6)] {
            let text = plain(&status, None);
            let mut seen = 0;
            for icon in width::ICONS {
                for (at, _) in text.match_indices(icon) {
                    seen += 1;
                    let after = &text[at + icon.len()..];
                    assert!(after.starts_with(' '), "{icon}: `{after}`");
                    assert!(
                        !after[1..].starts_with(' '),
                        "{icon}: more than one space: `{after}`"
                    );
                }
            }
            assert_eq!(seen, expected, "icons in {text}");
        }
    }
}

/// What the three test modules of line 3 (SPEC-20261122-statusline-limits AC-2 to AC-4) share:
/// one plain render of a stdin payload at a fixed clock, from a temporary HOME, and the items of
/// its last line.
#[cfg(test)]
mod line_three_support {
    use crate::statusline::state::Context;
    use std::path::Path;

    /// The render clock (`--now-unix` seam).
    pub const NOW: i64 = 1_738_429_000;

    /// Every line of one plain render of `payload` at [`NOW`], from the temporary `home`, in
    /// `cwd` when given.
    pub fn lines_in(payload: &serde_json::Value, home: &Path, cwd: Option<&Path>) -> Vec<String> {
        let vars = [("HOME", home.to_str().unwrap()), ("NO_COLOR", "1")];
        let mut ctx = Context::for_test(&vars, cwd);
        ctx.now = NOW;
        crate::statusline::render(payload.to_string().as_bytes(), &ctx)
            .lines()
            .map(str::to_string)
            .collect()
    }

    /// The last line of [`lines_in`]: the limits line.
    pub fn line_in(payload: &serde_json::Value, home: &Path, cwd: Option<&Path>) -> String {
        lines_in(payload, home, cwd).pop().unwrap_or_default()
    }

    /// [`line_in`] from a fresh temporary HOME and no working folder.
    pub fn line(payload: &serde_json::Value) -> String {
        let home = tempfile::tempdir().unwrap();
        line_in(payload, home.path(), None)
    }

    /// [`lines_in`] from a fresh temporary HOME and no working folder.
    pub fn lines(payload: &serde_json::Value) -> Vec<String> {
        let home = tempfile::tempdir().unwrap();
        lines_in(payload, home.path(), None)
    }

    /// The item of `line` that holds `label`, `None` when no item does.
    pub fn item(line: &str, label: &str) -> Option<String> {
        line.split(" | ")
            .find(|i| i.contains(label))
            .map(str::to_string)
    }
}

/// AC-2 of SPEC-20261122-statusline-limits (REQ-3, REQ-4, D-2, D-4): the battery icon switches at
/// 70/71, the reset countdown has three forms without `resets in`, a window whose reset has passed
/// is left out, and a week without a percentage says `7D n/a`. Since
/// SPEC-20261122-statusline-limits-left every number is the share left, so the percentages named
/// here are 100 minus what the line writes, and the five-hour window is shown at any level. Since
/// SPEC-20261123-statusline-order-bars every known window has a ten-cell bar of the same share
/// between its label and its number.
#[cfg(test)]
mod window_tests {
    use super::line_three_support::{item, line, NOW};
    use serde_json::{json, Value};

    /// The bar of a share used, written out from the formula of the SPEC
    /// (SPEC-20261124-statusline-counts-truth REQ-12): ten cells, the first `used * 10 / 100`
    /// (rounded down) of them `█` and the others `░`.
    fn bar(used: i64) -> String {
        let filled = (used * 10 / 100) as usize;
        format!("{}{}", "█".repeat(filled), "░".repeat(10 - filled))
    }

    /// Line 3 of a payload with the context window at 10 percent and these `rate_limits`.
    fn windows(rate_limits: Value) -> String {
        line(&json!({
            "context_window": {"used_percentage": 10},
            "rate_limits": rate_limits,
        }))
    }

    fn week(pct: Value, resets_at: Option<i64>) -> Option<String> {
        let mut w = json!({"used_percentage": pct});
        if let Some(at) = resets_at {
            w["resets_at"] = json!(at);
        }
        item(&windows(json!({ "seven_day": w })), "7D")
    }

    const DAY: i64 = 86_400;
    const HOUR: i64 = 3_600;
    const MINUTE: i64 = 60;

    /// A guard: the battery icon of a window turns between 70 and 71 percent used.
    #[test]
    fn the_battery_icon_switches_at_70_and_71() {
        assert_eq!(
            week(json!(70), None).as_deref(),
            Some("🔋 7D ███████░░░ 70%")
        );
        assert_eq!(
            week(json!(71), None).as_deref(),
            Some("🪫 7D ███████░░░ 71%")
        );
    }

    /// REQ-3 (SPEC-20261123-statusline-order-bars) read as the share used by REQ-12 of
    /// SPEC-20261124-statusline-counts-truth: the bar has ten cells whatever the number, a new
    /// cell fills at each tenth of the share used — 9 is no cell, 10 is one, 11 is one, 19 is
    /// one, 20 is two, 99 is nine, 100 is ten — and a number below 0 or above 100 never gives
    /// fewer than none or more than ten filled cells. The expected text is written out by hand
    /// at the boundaries, then the whole range is checked against the cell counts.
    #[test]
    fn the_bar_has_ten_cells_and_changes_at_each_tenth() {
        let context = |used: i64| line(&json!({"context_window": {"used_percentage": used}}));
        for (used, bar_text) in [
            (0, "░░░░░░░░░░"),
            (9, "░░░░░░░░░░"),
            (10, "█░░░░░░░░░"),
            (11, "█░░░░░░░░░"),
            (19, "█░░░░░░░░░"),
            (20, "██░░░░░░░░"),
            (99, "█████████░"),
            (100, "██████████"),
        ] {
            let icon = if used > 70 { "🪫" } else { "🔋" };
            assert_eq!(
                context(used),
                format!("{icon} CW {bar_text} {used}%"),
                "{used} used"
            );
        }
        // A share below 0 or above 100 is clamped: the bar stays ten cells, 0 to 10 of them full.
        assert_eq!(context(150), "🪫 CW ██████████ 100%");
        assert_eq!(context(1_000_000), "🪫 CW ██████████ 100%");
        assert_eq!(context(-20), "🔋 CW ░░░░░░░░░░ 0%");
        // Every share from 0 to 100 used: ten cells, `used / 10` of them filled, the filled
        // cells first, the number is the used percentage and the icon turns above 70.
        for used in 0..=100_i64 {
            let text = context(used);
            let filled = text.matches('█').count();
            let empty = text.matches('░').count();
            assert_eq!(filled + empty, 10, "{used}: {text}");
            assert_eq!(filled as i64, used / 10, "{used}: {text}");
            assert!(text.contains(&format!(" {} {used}%", bar(used))), "{text}");
            assert!(
                text.starts_with(if used > 70 { "🪫 CW " } else { "🔋 CW " }),
                "{text}"
            );
        }
    }

    /// REQ-6: the filled cells and the empty cells are separate color runs (the filled ones in
    /// their level color, the empty ones muted), so the colored line without its escapes is the
    /// plain line.
    #[test]
    fn the_colored_bar_has_the_plain_text() {
        use crate::statusline::color::Profile;
        use crate::statusline::input::RateWindow;
        use crate::statusline::render::{theme, Renderer};
        use crate::statusline::state::{DisplayConfig, Status};
        let status = Status {
            context: Some(27.0),
            five_hour: Some(RateWindow {
                used_percentage: Some(51.0),
                resets_at: Some(NOW + 43 * MINUTE),
            }),
            seven_day: Some(RateWindow {
                used_percentage: Some(13.0),
                resets_at: Some(NOW + 3 * DAY + 22 * HOUR),
            }),
            ..Default::default()
        };
        let plain = Renderer::new(&DisplayConfig::default(), None)
            .with_now(NOW)
            .render(&status);
        assert_eq!(
            plain,
            "🔋 CW ██░░░░░░░░ 27% | 🔋 5H █████░░░░░ 51% (43m) | 🔋 7D █░░░░░░░░░ 13% (3d 22h)"
        );
        for name in ["catppuccin-mocha", "catppuccin-latte"] {
            let t = theme(name);
            for profile in [Profile::TrueColor, Profile::Ansi256] {
                let config = DisplayConfig {
                    theme: name.into(),
                    ..Default::default()
                };
                let out = Renderer::new(&config, Some(profile))
                    .with_now(NOW)
                    .render(&status);
                let mut stripped = String::new();
                let mut rest = out.as_str();
                while let Some(at) = rest.find("\x1b[") {
                    stripped.push_str(&rest[..at]);
                    let tail = &rest[at..];
                    rest = &tail[tail.find('m').unwrap() + 1..];
                }
                stripped.push_str(rest);
                assert_eq!(stripped, plain, "{name} {profile:?}");
                // The empty cells are painted in the muted color, each run on its own; the
                // filled cells take the level color of their window
                // (SPEC-20261124-statusline-bar-color, `the_filled_cells_carry_the_level_color`).
                for empty in ["░░░░░░░░", "░░░░░", "░░░░░░░░░"]
                {
                    assert!(
                        out.contains(&profile.paint(t.muted, empty)),
                        "{name} {profile:?}: no muted run of {empty}: {out:?}"
                    );
                }
                for filled in ["██", "█████", "█"] {
                    assert!(out.contains(filled), "{out:?}");
                    assert!(
                        !out.contains(&profile.paint(t.muted, filled)),
                        "the filled cells are not muted: {out:?}"
                    );
                }
            }
        }
    }

    /// The level colors of SPEC-20261124-statusline-bar-color REQ-3 written out by hand, in the
    /// order green, yellow, peach, red.
    const MOCHA_LEVELS: [&str; 4] = ["#A6E3A1", "#F9E2AF", "#FAB387", "#F38BA8"];
    const LATTE_LEVELS: [&str; 4] = ["#40A02B", "#DF8E1D", "#FE640B", "#D20F39"];

    fn levels_of(theme_name: &str) -> [&'static str; 4] {
        if theme_name == "catppuccin-latte" {
            LATTE_LEVELS
        } else {
            MOCHA_LEVELS
        }
    }

    /// The limits line of a status whose windows have these used percentages, rendered with a
    /// color profile (or none) at [`NOW`]. Every window has a reset 43 minutes away.
    fn limits_colored(
        theme_name: &str,
        profile: Option<crate::statusline::color::Profile>,
        used: [Option<i64>; 3],
    ) -> String {
        use crate::statusline::input::RateWindow;
        use crate::statusline::render::Renderer;
        use crate::statusline::state::{DisplayConfig, Status};
        let window = |u: Option<i64>| {
            u.map(|u| RateWindow {
                used_percentage: Some(u as f64),
                resets_at: Some(NOW + 43 * MINUTE),
            })
        };
        let status = Status {
            context: used[0].map(|u| u as f64),
            five_hour: window(used[1]),
            seven_day: window(used[2]),
            ..Default::default()
        };
        let config = DisplayConfig {
            theme: theme_name.into(),
            ..Default::default()
        };
        Renderer::new(&config, profile)
            .with_now(NOW)
            .render(&status)
    }

    /// The bytes of one bar of `used` percent used as the renderer writes them: the filled cells
    /// painted in `level`, then the base color of the line again (the line re-applies it after
    /// every reset), then the empty cells painted in the muted color. A run with no cell is
    /// left out.
    fn bar_bytes(
        profile: crate::statusline::color::Profile,
        theme_name: &str,
        level: &str,
        used: i64,
    ) -> String {
        let t = crate::statusline::render::theme(theme_name);
        let filled = (used * 10 / 100) as usize;
        let empty = 10 - filled;
        let mut out = String::new();
        if filled > 0 {
            out += &profile.paint(level, &"█".repeat(filled));
            out += &format!("\x1b[{}m", profile.foreground(t.text));
        }
        if empty > 0 {
            out += &profile.paint(t.muted, &"░".repeat(empty));
        }
        out
    }

    /// REQ-1, REQ-3 and REQ-4 (SPEC-20261124-statusline-bar-color) read as the share used by
    /// REQ-12 and REQ-14 of SPEC-20261124-statusline-counts-truth: with a color profile the
    /// filled cells of the `CW`, `5H` and `7D` bars are one run in the level color of the share
    /// used, followed by the empty cells in the muted color, in both themes and both profiles.
    #[test]
    fn the_filled_cells_carry_the_level_color() {
        use crate::statusline::color::Profile;
        // (used of CW, 5H, 7D) → (used of each, level index of each): 27 green, 51 peach, 13 green;
        // 45 yellow, 81 red, 62 peach; 0 green with no filled cell, 30 green, 100 red with no
        // empty cell.
        type Case = ([i64; 3], [(i64, usize); 3]);
        let cases: [Case; 3] = [
            ([27, 51, 13], [(27, 0), (51, 2), (13, 0)]),
            ([45, 81, 62], [(45, 1), (81, 3), (62, 2)]),
            ([0, 30, 100], [(0, 0), (30, 0), (100, 3)]),
        ];
        for name in ["catppuccin-mocha", "catppuccin-latte"] {
            let levels = levels_of(name);
            for profile in [Profile::TrueColor, Profile::Ansi256] {
                for (used, lefts) in cases {
                    let out = limits_colored(
                        name,
                        Some(profile),
                        [Some(used[0]), Some(used[1]), Some(used[2])],
                    );
                    for (window, (share, level)) in ["CW", "5H", "7D"].iter().zip(lefts) {
                        let bytes = bar_bytes(profile, name, levels[level], share);
                        assert!(
                            out.contains(&bytes),
                            "{name} {profile:?} {window} {share} used: no {bytes:?} in {out:?}"
                        );
                    }
                }
            }
        }
    }

    /// REQ-14 of SPEC-20261124-statusline-counts-truth (it replaces REQ-2 of
    /// SPEC-20261124-statusline-bar-color): the level color follows the share used in four steps,
    /// pinned at N and N+1 — 30 is green and 31 yellow, 50 yellow and 51 peach, 70 peach and 71
    /// red (70 is `HIGH_USE_PERCENT`, where the battery icon turns too) — the table is written
    /// out by hand, and over the whole range the color only moves from green toward red as the
    /// share used grows. A bar with no filled cell (under 10 used) has no level color to read.
    #[test]
    fn the_level_color_changes_at_each_step_point() {
        use crate::statusline::color::Profile;
        let name = "catppuccin-mocha";
        let levels = levels_of(name);
        let profile = Profile::TrueColor;
        let level_at = |used: i64| -> usize {
            let out = limits_colored(name, Some(profile), [Some(used), None, None]);
            let seen: Vec<usize> = (0..4)
                .filter(|i| out.contains(&bar_bytes(profile, name, levels[*i], used)))
                .collect();
            assert_eq!(seen.len(), 1, "{used} used: levels {seen:?} in {out:?}");
            seen[0]
        };
        for (used, level) in [
            (10, 0),
            (29, 0),
            (30, 0),
            (31, 1),
            (49, 1),
            (50, 1),
            (51, 2),
            (69, 2),
            (70, 2),
            (71, 3),
            (89, 3),
            (100, 3),
        ] {
            assert_eq!(level_at(used), level, "{used} used");
        }
        // The step points of the table above are the constants of the renderer.
        assert_eq!(
            [
                super::LEVEL_GREEN_UP_TO,
                super::LEVEL_YELLOW_UP_TO,
                super::HIGH_USE_PERCENT
            ],
            [30, 50, 70]
        );
        let mut last = 0;
        for used in 10..=100 {
            let level = level_at(used);
            assert!(
                level >= last,
                "{used} used went back from {last} to {level}"
            );
            last = level;
        }
        assert_eq!(last, 3, "red is reached");
    }

    /// REQ-4 (SPEC-20261124-statusline-bar-color): only the filled cells take a level color; the
    /// head, the number and the countdown stay as they were, and a bar with no filled cell (under
    /// 10 percent left) is one muted run with no level color anywhere on the line.
    #[test]
    fn only_the_filled_cells_take_the_level_color() {
        use crate::statusline::color::Profile;
        let name = "catppuccin-mocha";
        let profile = Profile::TrueColor;
        let level_codes = MOCHA_LEVELS.map(|hex| format!("\x1b[{}m", profile.foreground(hex)));
        // Windows at 27, 51 and 13 percent used: every level escape of the line is the one run of
        // a bar of its own, so removing the three filled runs removes every level escape.
        let out = limits_colored(name, Some(profile), [Some(27), Some(51), Some(13)]);
        let mut rest = out.clone();
        for (used, level) in [(27, 0), (51, 2), (13, 0)] {
            let run = profile.paint(MOCHA_LEVELS[level], &"█".repeat((used * 10 / 100) as usize));
            assert!(rest.contains(&run), "{used}: {out:?}");
            rest = rest.replacen(&run, "", 1);
        }
        for code in &level_codes {
            assert!(
                !rest.contains(code),
                "a level color outside the filled cells: {out:?}"
            );
        }
        // No filled cell: 9 percent used on every window.
        let empty = limits_colored(name, Some(profile), [Some(9), Some(9), Some(9)]);
        for code in &level_codes {
            assert!(
                !empty.contains(code),
                "a level color without a filled cell: {empty:?}"
            );
        }
        let t = crate::statusline::render::theme(name);
        assert_eq!(
            empty.matches(&profile.paint(t.muted, "░░░░░░░░░░")).count(),
            3,
            "{empty:?}"
        );
    }

    /// REQ-5 (SPEC-20261124-statusline-bar-color): without a color profile the line has no escape
    /// sequence and, with one, the line without its escapes is that same line, for every share
    /// used from 0 to 100 in both themes and both profiles.
    #[test]
    fn the_level_color_leaves_the_plain_text_alone() {
        use crate::statusline::color::Profile;
        for share in 0..=100_i64 {
            let used = [Some(share), Some(share), Some(share)];
            let plain = limits_colored("catppuccin-mocha", None, used);
            assert!(!plain.contains('\x1b'), "{share}: {plain:?}");
            let bar = format!(
                "{}{}",
                "█".repeat((share * 10 / 100) as usize),
                "░".repeat(10 - (share * 10 / 100) as usize)
            );
            assert!(
                plain.contains(&format!("CW {bar} {share}%")),
                "{share}: {plain:?}"
            );
            for name in ["catppuccin-mocha", "catppuccin-latte"] {
                for profile in [Profile::TrueColor, Profile::Ansi256] {
                    let colored = limits_colored(name, Some(profile), used);
                    let mut stripped = String::new();
                    let mut rest = colored.as_str();
                    while let Some(at) = rest.find("\x1b[") {
                        stripped.push_str(&rest[..at]);
                        let tail = &rest[at..];
                        rest = &tail[tail.find('m').unwrap() + 1..];
                    }
                    stripped.push_str(rest);
                    assert_eq!(stripped, plain, "{name} {profile:?} {share} used");
                }
            }
        }
    }

    /// AC-9 of SPEC-20261124-statusline-counts-truth (it replaces AC-1 of
    /// SPEC-20261122-statusline-limits-left and AC-4 of SPEC-20261123-statusline-order-bars):
    /// each window shows the share used, label, bar, number — 25, 32 and 78 percent used (the
    /// user's example) are `CW ██░░░░░░░░ 25%`, `5H ███░░░░░░░ 32%` and `7D ███████░░░ 78%` —
    /// and a payload without a five-hour window has no `5H` item.
    #[test]
    fn the_row_shows_what_is_used_of_each_window() {
        let row = |context: i64, rate_limits: Value| {
            line(&json!({
                "context_window": {"used_percentage": context},
                "rate_limits": rate_limits,
            }))
        };
        // The user's example of 2026-11-24.
        assert_eq!(
            row(
                25,
                json!({
                    "five_hour": {"used_percentage": 32, "resets_at": NOW + 16 * MINUTE},
                    "seven_day": {"used_percentage": 78, "resets_at": NOW + 2 * DAY + 15 * HOUR},
                })
            ),
            "🔋 CW ██░░░░░░░░ 25% | 🔋 5H ███░░░░░░░ 32% (16m) | 🪫 7D ███████░░░ 78% (2d 15h)"
        );
        assert_eq!(
            row(
                27,
                json!({
                    "five_hour": {"used_percentage": 51, "resets_at": NOW + 43 * MINUTE},
                    "seven_day": {"used_percentage": 13, "resets_at": NOW + 3 * DAY + 22 * HOUR},
                })
            ),
            "🔋 CW ██░░░░░░░░ 27% | 🔋 5H █████░░░░░ 51% (43m) | 🔋 7D █░░░░░░░░░ 13% (3d 22h)"
        );
        assert_eq!(
            row(
                18,
                json!({
                    "five_hour": {"used_percentage": 75, "resets_at": NOW + HOUR + 10 * MINUTE},
                    "seven_day": {"used_percentage": 95, "resets_at": NOW + 4 * DAY + 2 * HOUR},
                })
            ),
            "🔋 CW █░░░░░░░░░ 18% | 🪫 5H ███████░░░ 75% (1h 10m) | 🪫 7D █████████░ 95% (4d 2h)"
        );
        let no_five = row(
            18,
            json!({
                "seven_day": {"used_percentage": 95, "resets_at": NOW + 4 * DAY + 2 * HOUR},
            }),
        );
        assert_eq!(
            no_five,
            "🔋 CW █░░░░░░░░░ 18% | 🪫 7D █████████░ 95% (4d 2h)"
        );
        assert_eq!(item(&no_five, "5H"), None);
        // The bar stands between the label and the number, with no colon: the countdown, the
        // battery icon and one space after it are as they were.
        assert!(!no_five.contains(':'), "{no_five}");
    }

    /// AC-2 of SPEC-20261122-statusline-limits-left: an unknown window says `n/a` with no battery
    /// icon, and the hide options `context`, `usage_5h` and `usage_7d` each take their item off
    /// the line.
    #[test]
    fn unknown_windows_and_the_hide_list_stay_as_they_were() {
        use crate::statusline::input::RateWindow;
        use crate::statusline::render::Renderer;
        use crate::statusline::state::{DisplayConfig, Status};
        let render = |status: &Status, off: &[&str]| {
            let config = DisplayConfig {
                segments: off.iter().map(|k| ((*k).to_string(), false)).collect(),
                ..Default::default()
            };
            Renderer::new(&config, None).render(status)
        };
        let unknown = Status {
            seven_day: Some(RateWindow::default()),
            ..Default::default()
        };
        assert_eq!(render(&unknown, &[]), "CW n/a | 7D n/a");
        // An unknown window has no battery icon and no bar, in any form.
        let text = render(&unknown, &[]);
        assert!(!text.contains('█') && !text.contains('░'), "{text}");
        let window = |used: f64| {
            Some(RateWindow {
                used_percentage: Some(used),
                resets_at: None,
            })
        };
        let known = Status {
            context: Some(18.0),
            five_hour: window(75.0),
            seven_day: window(95.0),
            ..Default::default()
        };
        let cw = "🔋 CW █░░░░░░░░░ 18%";
        let five = "🪫 5H ███████░░░ 75%";
        let week = "🪫 7D █████████░ 95%";
        assert_eq!(render(&known, &[]), format!("{cw} | {five} | {week}"));
        assert_eq!(render(&known, &["context"]), format!("{five} | {week}"));
        assert_eq!(render(&known, &["usage_5h"]), format!("{cw} | {week}"));
        assert_eq!(render(&known, &["usage_7d"]), format!("{cw} | {five}"));
        assert_eq!(render(&known, &["context", "usage_5h", "usage_7d"]), "MoAI");
    }

    /// AC-10 of SPEC-20261124-statusline-counts-truth (it replaces AC-3 of
    /// SPEC-20261122-statusline-limits-left): the icon turns above the high-use boundary —
    /// exactly `HIGH_USE_PERCENT` used is `🔋`, one percent more used is `🪫` — for the context
    /// window, the five-hour window and the week alike. The expected numbers come from the
    /// constant the render reads.
    #[test]
    fn the_battery_icon_turns_above_the_high_use_boundary() {
        use crate::statusline::render::HIGH_USE_PERCENT;
        let n = HIGH_USE_PERCENT;
        assert_eq!(week(json!(n), None), Some(format!("🔋 7D {} {n}%", bar(n))));
        assert_eq!(
            week(json!(n + 1), None),
            Some(format!("🪫 7D {} {}%", bar(n + 1), n + 1))
        );
        let five = |used: i64| {
            item(
                &windows(json!({"five_hour": {"used_percentage": used}})),
                "5H",
            )
        };
        assert_eq!(five(n), Some(format!("🔋 5H {} {n}%", bar(n))));
        assert_eq!(
            five(n + 1),
            Some(format!("🪫 5H {} {}%", bar(n + 1), n + 1))
        );
        let context = |used: i64| line(&json!({"context_window": {"used_percentage": used}}));
        assert_eq!(context(n), format!("🔋 CW {} {n}%", bar(n)));
        assert_eq!(context(n + 1), format!("🪫 CW {} {}%", bar(n + 1), n + 1));
    }

    /// D-4: `(6d 21h)`, `(2h 10m)`, `(35m)`, each unit changing at its edge, and never the words
    /// `resets in`.
    #[test]
    fn the_countdown_has_three_forms_without_resets_in() {
        for (left, want) in [
            (6 * DAY + 21 * HOUR, "🔋 7D ██░░░░░░░░ 21% (6d 21h)"),
            (2 * HOUR + 10 * MINUTE, "🔋 7D ██░░░░░░░░ 21% (2h 10m)"),
            (35 * MINUTE, "🔋 7D ██░░░░░░░░ 21% (35m)"),
            (DAY, "🔋 7D ██░░░░░░░░ 21% (1d 0h)"),
            (DAY - 1, "🔋 7D ██░░░░░░░░ 21% (23h 59m)"),
            (HOUR, "🔋 7D ██░░░░░░░░ 21% (1h 0m)"),
            (HOUR - 1, "🔋 7D ██░░░░░░░░ 21% (59m)"),
            (MINUTE, "🔋 7D ██░░░░░░░░ 21% (1m)"),
            (1, "🔋 7D ██░░░░░░░░ 21% (0m)"),
        ] {
            let got = week(json!(21), Some(NOW + left));
            assert_eq!(got.as_deref(), Some(want), "{left} seconds left");
        }
        let full = windows(json!({
            "five_hour": {"used_percentage": 85, "resets_at": NOW + HOUR + 10 * MINUTE},
            "seven_day": {"used_percentage": 21, "resets_at": NOW + 6 * DAY + 21 * HOUR},
        }));
        assert!(!full.contains("resets in"), "{full}");
    }

    /// REQ-3: the host drops a window once its reset time passes; a window that still names a
    /// passed reset is left out with its separator, never `(0m)`.
    #[test]
    fn a_window_whose_reset_has_passed_is_left_out() {
        assert_eq!(
            week(json!(21), Some(NOW)),
            None,
            "reset at the render's clock"
        );
        assert_eq!(week(json!(21), Some(NOW - 1)), None, "reset a second ago");
        assert_eq!(
            week(json!(21), Some(NOW + 1)).as_deref(),
            Some("🔋 7D ██░░░░░░░░ 21% (0m)")
        );
        assert_eq!(
            windows(json!({"seven_day": {"used_percentage": 21, "resets_at": NOW - DAY}})),
            "🔋 CW █░░░░░░░░░ 10%"
        );
        // Beyond 31 days the reset is no countdown: the week shows without one.
        let edge = NOW + crate::usage_limit::MAX_RESET_AHEAD_SECS;
        assert_eq!(
            week(json!(21), Some(edge)).as_deref(),
            Some("🔋 7D ██░░░░░░░░ 21% (31d 0h)")
        );
        for at in [edge + 1, i64::MAX] {
            assert_eq!(
                week(json!(21), Some(at)).as_deref(),
                Some("🔋 7D ██░░░░░░░░ 21%"),
                "{at}"
            );
        }
        assert_eq!(
            windows(json!({"five_hour": {"used_percentage": 90, "resets_at": NOW}})),
            "🔋 CW █░░░░░░░░░ 10%"
        );
    }

    /// REQ-3: a week without a usable percentage says `7D n/a`, with no battery icon and never
    /// `0%`; its countdown still follows it.
    #[test]
    fn a_week_without_a_percentage_says_n_a() {
        assert_eq!(
            item(&windows(json!({"seven_day": {}})), "7D").as_deref(),
            Some("7D n/a")
        );
        assert_eq!(week(json!("21"), None).as_deref(), Some("7D n/a"));
        assert_eq!(week(json!(null), None).as_deref(), Some("7D n/a"));
        assert_eq!(
            week(json!(null), Some(NOW + 6 * DAY + 21 * HOUR)).as_deref(),
            Some("7D n/a (6d 21h)")
        );
        assert_eq!(item(&windows(json!({})), "7D"), None, "an absent week");
    }

    /// REQ-3 of SPEC-20261122-statusline-limits-left: the five-hour window shows whenever the
    /// input carries its percentage, at any level, with the share left, the icon of the low-left
    /// rule and its countdown, between the context window and the week — 25 percent used is
    /// `🔋 5H 75%`. Without a percentage or absent it is left out with its separator. A fraction
    /// is read by the whole used percentage, so 70.9 is 70 used.
    #[test]
    fn the_five_hour_window_shows_whenever_the_input_carries_it() {
        let five = |pct: Value| {
            item(
                &windows(
                    json!({"five_hour": {"used_percentage": pct, "resets_at": NOW + 2 * HOUR + 10 * MINUTE}}),
                ),
                "5H",
            )
        };
        assert_eq!(
            five(json!(25)).as_deref(),
            Some("🔋 5H ██░░░░░░░░ 25% (2h 10m)")
        );
        assert_eq!(
            five(json!(0)).as_deref(),
            Some("🔋 5H ░░░░░░░░░░ 0% (2h 10m)")
        );
        assert_eq!(
            five(json!(70.9)).as_deref(),
            Some("🔋 5H ███████░░░ 70% (2h 10m)")
        );
        assert_eq!(
            five(json!(71)).as_deref(),
            Some("🪫 5H ███████░░░ 71% (2h 10m)")
        );
        assert_eq!(
            five(json!(100)).as_deref(),
            Some("🪫 5H ██████████ 100% (2h 10m)")
        );
        assert_eq!(
            five(json!(150)).as_deref(),
            Some("🪫 5H ██████████ 100% (2h 10m)")
        );
        assert_eq!(five(json!(null)), None);
        assert_eq!(five(json!("25")), None);
        // A reset time is not needed: the item stands without a countdown.
        assert_eq!(
            windows(json!({"five_hour": {"used_percentage": 25}})),
            "🔋 CW █░░░░░░░░░ 10% | 🔋 5H ██░░░░░░░░ 25%"
        );
        for absent in [
            json!({}),
            json!({"five_hour": {}}),
            json!({"five_hour": {"resets_at": NOW + HOUR}}),
            json!({"five_hour": null}),
        ] {
            assert_eq!(
                windows(absent.clone()),
                "🔋 CW █░░░░░░░░░ 10%",
                "{absent}: no item and no separator"
            );
        }
        assert_eq!(
            windows(json!({
                "five_hour": {"used_percentage": 85, "resets_at": NOW + HOUR + 10 * MINUTE},
                "seven_day": {"used_percentage": 21, "resets_at": NOW + 6 * DAY + 21 * HOUR},
            })),
            "🔋 CW █░░░░░░░░░ 10% | 🪫 5H ████████░░ 85% (1h 10m) | 🔋 7D ██░░░░░░░░ 21% (6d 21h)"
        );
    }
}

/// AC-3 of SPEC-20261122-statusline-limits (REQ-5, D-3): the cost shows only when the payload
/// carries neither a five-hour nor a seven-day window — a payload with only a spend limit
/// included — and is written to two decimals.
#[cfg(test)]
mod cost_tests {
    use super::line_three_support::{item, line, NOW};
    use serde_json::{json, Value};

    /// The context window at 10 percent used: one cell of the bar.
    const CW: &str = "🔋 CW █░░░░░░░░░ 10%";

    /// Line 3 of a payload with the context window at 10 percent, `cost` as given and these
    /// `rate_limits` when given.
    fn with_cost(cost: Value, rate_limits: Option<Value>) -> String {
        let mut payload = json!({"context_window": {"used_percentage": 10}, "cost": cost});
        if let Some(limits) = rate_limits {
            payload["rate_limits"] = limits;
        }
        line(&payload)
    }

    fn usd(total: Value) -> Value {
        json!({"total_cost_usd": total, "total_duration_ms": 45_000})
    }

    #[test]
    fn the_cost_shows_without_a_window() {
        assert_eq!(with_cost(usd(json!(3.87)), None), format!("{CW} | 💰 3.87"));
        assert_eq!(
            with_cost(usd(json!(3.87)), Some(json!({}))),
            format!("{CW} | 💰 3.87"),
            "rate_limits without a window"
        );
    }

    /// D-3: a gateway user whose payload has only a spend limit sees the cost.
    #[test]
    fn the_cost_shows_with_only_a_spend_limit() {
        let spend = json!({"spend_limit": {"used_percentage": 62.8, "resets_at": NOW + 3_600, "period": "monthly"}});
        assert_eq!(
            with_cost(usd(json!(3.87)), Some(spend)),
            format!("{CW} | 💰 3.87")
        );
    }

    /// A payload with either window shows no cost, also when the window itself is not drawn (a
    /// window whose reset has passed).
    #[test]
    fn the_cost_is_hidden_with_either_window() {
        for limits in [
            json!({"five_hour": {"used_percentage": 23, "resets_at": NOW + 3_600}}),
            json!({"five_hour": {"used_percentage": 85, "resets_at": NOW + 3_600}}),
            json!({"seven_day": {"used_percentage": 21, "resets_at": NOW + 3_600}}),
            json!({"seven_day": {"used_percentage": 21, "resets_at": NOW - 1}}),
            json!({"five_hour": {"used_percentage": 23}, "seven_day": {"used_percentage": 21}, "spend_limit": {"used_percentage": 1}}),
        ] {
            let out = with_cost(usd(json!(3.87)), Some(limits.clone()));
            assert_eq!(item(&out, "💰"), None, "{limits}: {out}");
        }
    }

    #[test]
    fn the_cost_rounds_to_two_decimals() {
        for (total, want) in [
            (json!(3.866), "💰 3.87"),
            (json!(3.864), "💰 3.86"),
            (json!(0.01234), "💰 0.01"),
            (json!(0), "💰 0.00"),
            (json!(12.5), "💰 12.50"),
            (json!(1234.5), "💰 1234.50"),
        ] {
            let out = with_cost(usd(total.clone()), None);
            assert_eq!(item(&out, "💰").as_deref(), Some(want), "{total}: {out}");
        }
    }

    /// A guard for the missing cases, red for the mistyped one only if a wrong type ever showed:
    /// a cost that is absent or not a number is left out with its separator.
    #[test]
    fn a_missing_or_mistyped_cost_is_left_out() {
        for cost in [
            json!(null),
            json!({}),
            usd(json!("3.87")),
            usd(json!(null)),
            json!(5),
        ] {
            assert_eq!(with_cost(cost.clone(), None), CW, "{cost}");
        }
    }

    /// A negative cost and one too large for the line are left out, a negative zero prints
    /// `0.00`; the
    /// largest shown value has two decimals.
    #[test]
    fn a_negative_or_huge_cost_is_left_out() {
        for total in [json!(-0.001), json!(-3.5), json!(1_000_000.0), json!(1e308)] {
            assert_eq!(with_cost(usd(total.clone()), None), CW, "{total}");
        }
        let out = with_cost(usd(json!(999_999.99)), None);
        assert_eq!(item(&out, "💰").as_deref(), Some("💰 999999.99"), "{out}");
        let out = with_cost(usd(json!(-0.0)), None);
        assert_eq!(item(&out, "💰").as_deref(), Some("💰 0.00"), "{out}");
    }

    /// The `cost` option off leaves the cost out of a payload that would show it.
    #[test]
    fn the_cost_option_off_leaves_the_cost_out() {
        use crate::statusline::render::Renderer;
        use crate::statusline::state::DisplayConfig;
        let mut status = crate::statusline::state::Status {
            context: Some(10.0),
            cost: Some(3.87),
            ..Default::default()
        };
        let line = |status: &crate::statusline::state::Status, off: &[&str]| {
            let config = DisplayConfig {
                segments: off.iter().map(|k| ((*k).to_string(), false)).collect(),
                ..Default::default()
            };
            Renderer::new(&config, None).render(status)
        };
        assert!(line(&status, &[]).contains("💰 3.87"));
        assert!(!line(&status, &["cost"]).contains('💰'));
        status.cost = None;
        assert!(!line(&status, &[]).contains('💰'));
    }
}

/// AC-4 of SPEC-20261122-statusline-limits (REQ-6), moved to the second line by
/// SPEC-20261123-statusline-order-bars REQ-1: `🤖 <version>` from the stdin, then `🗿 <version>` of
/// this build, or `🗿 <template> -> <build>` when the project's template is older, in front of
/// the flow; a missing version is left out. The limits line closes the screen.
#[cfg(test)]
mod version_tests {
    use super::line_three_support::{line, lines, lines_in};
    use serde_json::json;

    const BUILD: &str = crate::BUILD_VERSION;
    /// The context window at 10 percent used, as the limits line writes it.
    const CW: &str = "🔋 CW █░░░░░░░░░ 10%";

    #[test]
    fn both_versions_open_the_second_line_without_a_v() {
        let out = lines(&json!({"context_window": {"used_percentage": 10}, "version": "2.1.330"}));
        assert_eq!(out, [format!("🤖 2.1.330 | 🗿 {BUILD}"), CW.to_string()]);
        // The limits line no longer holds them.
        let limits =
            line(&json!({"context_window": {"used_percentage": 10}, "version": "2.1.330"}));
        assert_eq!(limits, CW);
        assert!(!limits.contains('🤖') && !limits.contains('🗿'), "{limits}");
    }

    /// The version comes from stdin; an escape, a newline, a direction mark or a zero-width
    /// character in it never reaches the screen.
    #[test]
    fn control_characters_in_the_claude_code_version_are_dropped() {
        let out = lines(
            &json!({"context_window": {"used_percentage": 10}, "version": "2.1\u{1b}[31m.330\n"}),
        );
        assert_eq!(out[0], format!("🤖 2.1[31m.330 | 🗿 {BUILD}"));
        let hidden = "\u{202E}2.1\u{200B}.3\u{2066}30\u{FEFF}\u{00AD}\u{061C}";
        let out = lines(&json!({"context_window": {"used_percentage": 10}, "version": hidden}));
        assert_eq!(out[0], format!("🤖 2.1.330 | 🗿 {BUILD}"));
    }

    #[test]
    fn a_missing_claude_code_version_is_left_out() {
        for version in [json!(null), json!(""), json!(5)] {
            let out =
                lines(&json!({"context_window": {"used_percentage": 10}, "version": version}));
            assert_eq!(out, [format!("🗿 {BUILD}"), CW.to_string()], "{version}");
        }
    }

    /// A project whose template is older than this build shows the update arrow, the template
    /// first; a template equal to the build shows the build alone. The template is a HeadVer
    /// version of an earlier week, so it stays older than any build of this head.
    #[test]
    fn an_older_template_shows_the_update_arrow() {
        let home = tempfile::tempdir().unwrap();
        let project = home.path().join("project");
        let sections = project
            .join(crate::state_root::MOAI_DIR)
            .join("config/sections");
        std::fs::create_dir_all(&sections).unwrap();
        let payload = json!({"context_window": {"used_percentage": 10}, "version": "2.1.330"});
        for (template, want) in [
            ("4.2600.15", format!("🗿 4.2600.15 -> {BUILD}")),
            ("v4.2600.15", format!("🗿 4.2600.15 -> {BUILD}")),
            (BUILD, format!("🗿 {BUILD}")),
        ] {
            std::fs::write(
                sections.join("system.yaml"),
                format!("moai:\n  template_version: {template}\n"),
            )
            .unwrap();
            let out = lines_in(&payload, home.path(), Some(&project));
            assert_eq!(
                out,
                [format!("🤖 2.1.330 | {want}"), CW.to_string()],
                "template {template}"
            );
        }
    }
}
