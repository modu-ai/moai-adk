//! The people-facing help layout (#2802 S1; design #2799 D2 section 9, leader D3 6865936070).
//!
//! `moai --help` lists the commands a person types, in five groups; `moai help --all` lists
//! every command in eight groups. The JSON envelope on stdout is not touched by anything here:
//! help is text for people. clap has no way to put subcommands under several headings, so the
//! root help body is drawn from [`GROUPS`] and handed to clap as the root's help template.

use clap::Command;
use unicode_width::UnicodeWidthStr;

/// One help group: its heading and its commands, each with whether the root help shows it.
pub struct Group {
    pub heading: &'static str,
    pub commands: &'static [(&'static str, bool)],
}

/// Every command word in exactly one group (pinned by `help_table_names_every_command_once`).
/// `true` marks the commands the root help shows (#2799 B 6865555998, root draft 6865796537);
/// the rest appear only in `help --all`.
pub const GROUPS: &[Group] = &[
    Group {
        heading: "Start",
        commands: &[("cc", true), ("codex", true), ("glm", true)],
    },
    Group {
        heading: "Project",
        commands: &[
            ("project-detect", true),
            ("project-plan", true),
            ("project-apply", true),
            ("language-setup", true),
            ("index-build", true),
            ("agents-plan", false),
            ("agents-apply", false),
            ("generation-plan", false),
            ("generation-apply", false),
            ("github-plan", false),
            ("github-apply", false),
            ("labels-plan", false),
            ("labels-apply", false),
            ("claude-settings-plan", false),
            ("claude-settings-apply", false),
            ("codex-config-plan", false),
            ("codex-config-apply", false),
            ("hooks-plan", false),
            ("hooks-apply", false),
            ("hooks-revert", false),
        ],
    },
    Group {
        heading: "Update and move",
        commands: &[
            ("update", true),
            ("go-state-plan", true),
            ("go-state-apply", true),
            ("go-state-revert-plan", true),
            ("go-state-revert", true),
            ("home-state-plan", true),
            ("home-state-apply", true),
            ("home-state-revert", true),
            ("worktree-move-plan", true),
            ("worktree-move-apply", true),
            ("update-plan", false),
            ("update-apply", false),
            ("recovery-plan", false),
            ("recovery-apply", false),
            ("apply-recover", false),
            ("layout-plan", false),
            ("layout-apply", false),
            ("migrate-plan", false),
            ("migrate-apply", false),
            ("migrate-recovery-plan", false),
            ("migrate-recover", false),
            ("skills-legacy", false),
            ("bin-names-plan", false),
            ("bin-names-apply", false),
            ("agents-home-plan", false),
            ("agents-home-apply", false),
            ("agents-home-revert", false),
            ("home-routing-plan", false),
            ("home-routing-apply", false),
            ("home-routing-revert", false),
        ],
    },
    Group {
        heading: "Keys",
        commands: &[
            ("key-set", true),
            ("key-status", true),
            ("key-remove", true),
            ("key-import-plan", true),
            ("key-import-apply", true),
        ],
    },
    Group {
        heading: "Diagnose",
        commands: &[
            ("claude-doctor", true),
            ("codex-doctor", true),
            ("install-inspect", true),
            ("identity", true),
            ("worktree-report", true),
            ("optimize-plan", true),
            ("baseline-inventory", true),
        ],
    },
    Group {
        heading: "Lane and leader",
        commands: &[
            ("claim-acquire", false),
            ("claim-verify", false),
            ("claim-show", false),
            ("claim-release", false),
            ("claim-handoff", false),
            ("queue-next", false),
            ("queue-free", false),
            ("queue-query", false),
            ("leader-tick", false),
            ("lane-status", false),
            ("lane-handoff-check", false),
            ("lane-safe-to-reset", false),
            ("session-register", false),
            ("usage-limit", false),
            ("heavy-slot-acquire", false),
            ("heavy-slot-status", false),
            ("goal-show", false),
            ("goal-plan", false),
            ("goal-apply", false),
            ("goal-block", false),
            ("gate-check", false),
            ("gate-merge", false),
            ("gate-loop", false),
            ("wait", false),
            ("scope-check", false),
            ("worktree-plan", false),
            ("worktree-apply", false),
            ("pr-body-render", false),
            ("pr-confirm", false),
            ("audit-claim", false),
            ("audit-clone-plan", false),
            ("labels-check", false),
            ("decision-due", false),
            ("decision-file-check", false),
            ("decision-append", false),
            ("flow-wip-check", false),
            ("dispatch-record", false),
            ("dispatch-ack", false),
            ("dispatch-check", false),
            ("message-record", false),
            ("message-ack", false),
            ("message-reconcile", false),
            ("effect-intend", false),
            ("effect-confirm", false),
            ("effects-reconcile", false),
        ],
    },
    Group {
        heading: "Checks",
        commands: &[
            ("review-package", false),
            ("review-check", false),
            ("repo-scans", false),
            ("pr-self-audit-check", false),
            ("flake-classify", false),
            ("flake-record", false),
            ("audit-delta", false),
            ("audit-reduce", false),
            ("parity-check", false),
            ("plan-admit", false),
            ("run-verify", false),
        ],
    },
    Group {
        heading: "Internal",
        commands: &[
            ("statusline", false),
            ("hook-receipt", false),
            ("hook-dispatch", false),
            ("forge-read", false),
            ("forge-links-probe", false),
            ("forge-report", false),
            ("forge-audit-status", false),
            ("index-query", false),
            ("atlas-map", false),
            ("atlas-graph", false),
            ("atlas-tag-scan", false),
            ("search", false),
            ("context-bundle", false),
            ("context-verify", false),
            ("decision-record", false),
            ("cost-report", false),
            ("snapshot-describe", false),
            ("package-inspect", false),
            ("package-resolve", false),
            ("codex-projection-plan", false),
            ("codex-projection-apply", false),
            ("project-skill-plan", false),
            ("project-skill-apply", false),
            ("project-skill-revert", false),
            ("archive-plan", false),
            ("archive-apply", false),
            ("archive-restore", false),
            ("archive-search", false),
            ("worktree-gc-plan", false),
            ("worktree-gc-apply", false),
            ("session", false),
        ],
    },
];

/// The launchers the root help shows on one line.
const LAUNCHERS: [&str; 3] = ["cc", "codex", "glm"];
const LAUNCHER_LINE: (&str, &str) = (
    "cc, codex, glm",
    "Start a Claude Code, Codex or GLM session",
);

/// The group heading of a command word, or `None` for a word in no group.
pub fn group_of(word: &str) -> Option<&'static str> {
    GROUPS
        .iter()
        .find(|g| g.commands.iter().any(|(w, _)| *w == word))
        .map(|g| g.heading)
}

/// The name a person typed: `moai` or `moai-cli` from `argv[0]` (#2802, D2 section 9); any
/// other name (a renamed copy) reads as `moai-cli`, the canonical one.
pub fn bin_name(argv0: &str) -> &'static str {
    let stem = std::path::Path::new(argv0)
        .file_stem()
        .and_then(|s| s.to_str())
        .unwrap_or("");
    if stem == "moai" {
        "moai"
    } else {
        "moai-cli"
    }
}

/// The width help lines wrap at: `COLUMNS` when set and sane, else 100 (clap's own cap).
fn width() -> usize {
    std::env::var("COLUMNS")
        .ok()
        .and_then(|c| c.trim().parse::<usize>().ok())
        .filter(|w| (20..=1000).contains(w))
        .unwrap_or(100)
        .min(100)
}

/// `  name  about` rows under each heading, the about column wrapped to `width` with a hanging
/// indent; `all` lists every command, otherwise only the root ones (launchers on one line).
pub fn render_groups(cmd: &Command, all: bool, width: usize) -> String {
    let about = |word: &str| -> String {
        cmd.find_subcommand(word)
            .and_then(|s| s.get_about())
            .map(|a| a.to_string())
            .unwrap_or_default()
    };
    let mut rows: Vec<(&str, Vec<(String, String)>)> = Vec::new();
    for g in GROUPS {
        let mut list = Vec::new();
        for (word, root) in g.commands {
            if !all && !root {
                continue;
            }
            if !all && LAUNCHERS.contains(word) {
                if *word == LAUNCHERS[0] {
                    list.push((LAUNCHER_LINE.0.to_string(), LAUNCHER_LINE.1.to_string()));
                }
                continue;
            }
            list.push((word.to_string(), about(word)));
        }
        if !list.is_empty() {
            rows.push((g.heading, list));
        }
    }
    let name_w = rows
        .iter()
        .flat_map(|(_, l)| l.iter().map(|(n, _)| n.width()))
        .max()
        .unwrap_or(0);
    let indent = 2 + name_w + 2;
    let mut out = String::new();
    for (heading, list) in rows {
        out.push_str(heading);
        out.push('\n');
        for (name, text) in list {
            let pad = name_w - name.width();
            let row = format!(
                "  {name}{}  {}",
                " ".repeat(pad),
                wrap(&text, indent, width)
            );
            // A command without an about leaves no padding at the line end.
            out.push_str(row.trim_end());
            out.push('\n');
        }
    }
    out
}

/// `text` broken at spaces so no line passes `width` display cells (a wide character counts
/// two), later lines indented by `indent`.
fn wrap(text: &str, indent: usize, width: usize) -> String {
    let room = width.saturating_sub(indent).max(20);
    let mut lines: Vec<String> = Vec::new();
    let mut line = String::new();
    for word in text.split_whitespace() {
        if !line.is_empty() && line.width() + 1 + word.width() > room {
            lines.push(std::mem::take(&mut line));
        }
        if !line.is_empty() {
            line.push(' ');
        }
        line.push_str(word);
    }
    if !line.is_empty() || lines.is_empty() {
        lines.push(line);
    }
    lines.join(&format!("\n{}", " ".repeat(indent)))
}

/// The root command with the people-facing layout: usage and help name the typed `bin`, and the
/// root help body is the grouped list (`all` for `help --all`).
pub fn configure(cmd: Command, bin: &'static str, version: &str, all: bool) -> Command {
    let groups = render_groups(&cmd, all, width());
    let footer = if all {
        format!("Run '{bin} <command> --help' for details and examples.")
    } else {
        format!(
            "Run '{bin} help --all' for every command.\n\
             Run '{bin} <command> --help' for details and examples."
        )
    };
    // The template is literal text plus clap's `{usage-heading} {usage}` and, for `--all`,
    // `{options}`; braces in the group text would read as template tags, and command abouts
    // carry none (pinned by `help_text_has_no_template_braces`).
    let options = if all { "\nOptions:\n{options}\n" } else { "" };
    let template = format!(
        "MoAI-ADK {version} — agentic development kit\n\n{{usage-heading}} {{usage}}\n\n{groups}{options}\n{footer}\n"
    );
    cmd.bin_name(bin)
        .override_usage(format!(
            "{bin} <command> [options]\n       {bin} cc|codex|glm -l|-a    lead or agent session"
        ))
        .help_template(template)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn bin_name_reads_the_typed_name() {
        assert_eq!(bin_name("/usr/local/bin/moai"), "moai");
        assert_eq!(bin_name("moai"), "moai");
        assert_eq!(bin_name("/x/moai-cli"), "moai-cli");
        assert_eq!(bin_name("/x/moai-copy"), "moai-cli");
        assert_eq!(bin_name(""), "moai-cli");
    }

    #[test]
    fn wrap_keeps_lines_inside_the_width() {
        let text = "Preview moving the Go-era .moai folder and everything it holds";
        let out = wrap(text, 24, 60);
        for line in out.lines() {
            assert!(line.chars().count() <= 60, "{line:?}");
        }
        assert_eq!(out.split_whitespace().collect::<Vec<_>>().join(" "), text);
        assert_eq!(wrap("short", 24, 60), "short");
    }

    /// Hangul counts two cells, so a Korean about wraps at half the characters (S6 will need it).
    #[test]
    fn wrap_counts_wide_characters_as_two_cells() {
        let text = "한글 설명 문장은 두 칸씩 차지해요 그래서 더 일찍 줄을 바꿔요";
        let out = wrap(text, 4, 30);
        for line in out.lines() {
            assert!(UnicodeWidthStr::width(line) <= 30, "{line:?}");
        }
        assert!(out.lines().count() >= 3, "{out}");
    }

    #[test]
    fn group_headings_are_unique() {
        let mut seen = std::collections::BTreeSet::new();
        for g in GROUPS {
            assert!(seen.insert(g.heading), "{}", g.heading);
        }
    }
}
