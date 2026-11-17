//! Card C-53 S2: MODE-AGENT and MODE-LEADER each carry one section that maps the round to the
//! local card ledger and says it wins where a ledger is registered. Pinned here: the section and
//! its sentence, every card action named in both files, and every `moai card <action> --flag`
//! the two sections show checked against `--help` of this build.
//!
//! The binary runs with a temporary HOME; only `--help` is called and no ledger is made.

mod common;

use common::home::TempHome;
use std::{fs, path::Path};

const TITLE: &str = "Cards replace the Issue ledger";
const WINS: &str = "Where the project has a registered card ledger, this section wins over every \
                    Issue-based step of this workflow";
/// The actions a round uses. `export` is not one of them.
const ACTIONS: [&str; 17] = [
    "list",
    "show",
    "add",
    "scope",
    "ready",
    "take",
    "release",
    "handoff",
    "next",
    "note",
    "criterion",
    "pr",
    "review",
    "reopen",
    "done",
    "renew",
    "check",
];

fn workflow(name: &str) -> String {
    fs::read_to_string(
        Path::new(env!("CARGO_MANIFEST_DIR"))
            .join("skills/moai/workflows")
            .join(name),
    )
    .unwrap()
}

/// The text of the card section of a workflow, which the file holds exactly once.
fn card_section(name: &str) -> String {
    let text = workflow(name);
    let head = format!("\n## {TITLE}\n");
    assert_eq!(
        text.matches(&head).count(),
        1,
        "{name}: one `## {TITLE}` section"
    );
    let start = text.find(&head).unwrap() + head.len();
    let rest = &text[start..];
    rest[..rest.find("\n## ").unwrap_or(rest.len())].to_string()
}

/// Every code span that starts with `moai card ` or `card `: its action and the options it
/// shows (`--flag`, without a value).
fn card_spans(text: &str) -> Vec<(String, Vec<String>)> {
    let mut out = Vec::new();
    for (i, span) in text.split('`').enumerate() {
        // Odd pieces are inside a pair of backticks.
        if i % 2 == 0 {
            continue;
        }
        let Some(rest) = span
            .strip_prefix("moai card ")
            .or_else(|| span.strip_prefix("card "))
        else {
            continue;
        };
        let mut words = rest.split_whitespace();
        let action = words.next().unwrap_or_default().to_string();
        let flags = words
            .filter(|w| w.starts_with("--"))
            .map(|w| {
                let name: String = w[2..]
                    .chars()
                    .take_while(|c| c.is_ascii_alphanumeric() || *c == '-')
                    .collect();
                format!("--{name}")
            })
            .collect();
        out.push((action, flags));
    }
    out
}

fn card_help(home: &TempHome, action: &str) -> String {
    let out = home
        .command(&["card", action, "--help"])
        .output()
        .expect("run moai");
    assert!(
        out.status.success(),
        "`moai card {action} --help` failed: {}",
        String::from_utf8_lossy(&out.stderr)
    );
    String::from_utf8_lossy(&out.stdout).into_owned()
}

/// True when `help` lists the option `flag` (an option line starts with it, after a short
/// form if there is one).
fn help_has(help: &str, flag: &str) -> bool {
    help.lines().any(|line| {
        let line = line.trim_start();
        let line = line
            .strip_prefix("-h, ")
            .or_else(|| line.strip_prefix("-V, "))
            .unwrap_or(line);
        line.strip_prefix(flag)
            .is_some_and(|rest| rest.is_empty() || rest.starts_with(' '))
    })
}

/// What a section shows that this build does not have: `(action, flag)`; an unknown action has
/// an empty flag.
fn unknown(home: &TempHome, text: &str) -> Vec<(String, String)> {
    let mut bad = Vec::new();
    for (action, flags) in card_spans(text) {
        if !ACTIONS.contains(&action.as_str()) {
            bad.push((action, String::new()));
            continue;
        }
        let help = card_help(home, &action);
        for flag in flags {
            if !help_has(&help, &flag) {
                bad.push((action.clone(), flag));
            }
        }
    }
    bad
}

#[test]
fn both_modes_carry_one_card_section_that_wins_where_a_ledger_is_registered() {
    for name in ["mode-agent.md", "mode-leader.md"] {
        let section = card_section(name);
        assert!(section.contains(WINS), "{name} must say: {WINS}");
        // How a session knows: the read-only list answers `ledger: "ready"`.
        for phrase in [
            "`moai card list`",
            "`ledger: \"ready\"`",
            "`ledger: \"absent\"`",
            "user decision 2026-11-16",
        ] {
            assert!(section.contains(phrase), "{name} must say {phrase}");
        }
    }
}

#[test]
fn both_modes_name_every_card_action_of_the_round() {
    for name in ["mode-agent.md", "mode-leader.md"] {
        let named: Vec<String> = card_spans(&card_section(name))
            .into_iter()
            .map(|(action, _)| action)
            .collect();
        let missing: Vec<&str> = ACTIONS
            .iter()
            .filter(|a| !named.iter().any(|n| n == *a))
            .copied()
            .collect();
        assert!(
            missing.is_empty(),
            "{name} names no `card {missing:?}` in its card section"
        );
    }
}

#[test]
fn the_card_sections_show_only_actions_and_options_this_build_has() {
    let home = TempHome::new();
    let out = home
        .command(&["card", "--help"])
        .output()
        .expect("run moai");
    assert!(out.status.success(), "`moai card --help` failed");
    let root = String::from_utf8_lossy(&out.stdout).into_owned();
    for action in ACTIONS {
        assert!(
            root.lines()
                .any(|l| l.split_whitespace().next() == Some(action)),
            "`moai card` has no action `{action}`"
        );
    }
    for name in ["mode-agent.md", "mode-leader.md"] {
        let section = card_section(name);
        assert!(
            card_spans(&section).len() >= ACTIONS.len(),
            "{name}: the section shows its commands as code spans"
        );
        assert_eq!(
            unknown(&home, &section),
            Vec::<(String, String)>::new(),
            "{name} shows a card action or option that `--help` lacks"
        );
    }
    // The check does refuse: an option no action has, an action that does not exist, and an
    // option of another action.
    assert_eq!(
        unknown(
            &home,
            "`moai card take C-1 --holder x --scope src/`, `card steal C-1` and \
             `card release C-1 --generation 1 --lease-secs 60`; `moai card-import-plan --from x`"
        ),
        [
            ("take".to_string(), "--holder".to_string()),
            ("steal".to_string(), String::new()),
            ("release".to_string(), "--lease-secs".to_string()),
        ]
    );
}

fn assert_says(name: &str, phrases: &[&str]) {
    let section = card_section(name);
    for phrase in phrases {
        assert!(section.contains(phrase), "{name} must say `{phrase}`");
    }
}

#[test]
fn the_agent_section_maps_each_step_of_the_round_to_a_card_command() {
    assert_says(
        "mode-agent.md",
        &[
            // Find, take, keep. The section is short because the file sits near the package's
            // per-file limit (see the size test below).
            "`moai card next`",
            "`moai card take <card> --scope <p> --worktree <w> --branch card/<n>-<slug>`",
            "each writer then takes `--generation <g>`",
            "`card renew` before the lease ends",
            "`card check <card> --generation <g> --path <p>` before each write",
            // Reports, decisions, criteria.
            "`card note --kind report < file`",
            "`PLAN C-<n>`",
            "`--kind decision` note needs `--by`",
            "`card criterion <card> <n> --met --evidence <ref>`",
            "`--withdraw --reason <text>`",
            // Pull request, review, close. A local card names its host project (audit N2).
            "`card pr <card> <pr> --repo <o/r>` (local card), then `card review` once the audit \
             is asked",
            "on FAIL `card reopen`",
            "after the merge `card pr` again and `card done`",
            // Giving a card back or on.
            "`card release` (never in `review`)",
            "`card handoff --to <session>`",
            // Nothing to take, and a review card nobody holds.
            "`LANE-IDLE lane=<lane>`",
            "never take a GitHub Issue instead",
            "`review_no_holder`",
            "never edit the ledger file",
            // What a lane does not do.
            "`card add`, `card scope` and `card ready` are the leader's",
            // The parts that still read Issue claims: the commit guard lets a card branch
            // through as "not a lane worktree" (`claim_verify.rs`, audit N7).
            "the commit guard answers `allowed` on a card branch, checking nothing",
            "C-53 S3",
        ],
    );
    let agent = card_section("mode-agent.md");
    assert!(
        !agent.contains("`not_run`"),
        "the guard answers, it is not not_run"
    );
}

#[test]
fn the_leader_section_makes_cards_ready_assigns_audits_and_names_the_transition() {
    assert_says(
        "mode-leader.md",
        &[
            "`moai card add --title <title> --ac <criterion> --scope <path> --priority <p0..p3>`",
            // A card added without a scope is skipped by `card next` until the leader sets one.
            "`moai card scope <card> --scope <path> --by <leader>`",
            "`empty_scope`",
            "`moai card ready <card> --by <leader>`",
            "a hint, never a grant",
            // Audits: one read-only agent per pull request, never the implementer.
            "read-only audit agent",
            "never the session that implemented it",
            "`moai card note <card> --kind report --by <leader> < file`",
            "`moai card note <card> --kind decision --by <leader> < file`",
            // The holder closes; the leader merges through the gate.
            "`card done` is the holder's",
            // The transition exception ends with the gate change.
            "transition exception: C-<n>",
            "only as long as `moai gate-check` still asks for an Issue claim",
            "C-53 S3",
            "`review_no_holder`",
        ],
    );
}

/// Every package file is read with the bound `profile::LIMIT` (`profile::read`, used by
/// `package_digest` for install, update and `package-inspect`). MODE-AGENT was 64,181 bytes
/// before its card section, which is why that section is one paragraph; a file over the bound
/// makes the whole package unreadable ("target is not a bounded regular file").
#[test]
fn both_mode_files_stay_within_the_package_file_bound() {
    let bound = moai_cli::profile::LIMIT as usize;
    for name in ["mode-agent.md", "mode-leader.md"] {
        let bytes = workflow(name).len();
        assert!(
            bytes <= bound,
            "{name} is {bytes} bytes, over the {bound}-byte bound of a package file"
        );
    }
}

#[test]
fn spans_and_help_lines_are_read_as_written() {
    assert_eq!(
        card_spans(
            "`moai card take <card> --scope <p> --branch card/<n>-<slug>`, `card note C-1 --kind \
             report|decision`, `moai card-import-plan --from github` and `cardinal --x`"
        ),
        [
            (
                "take".to_string(),
                vec!["--scope".to_string(), "--branch".to_string()]
            ),
            ("note".to_string(), vec!["--kind".to_string()]),
        ]
    );
    let help = "Options:\n      --generation <GENERATION>\n          The generation\n      --met\n  -h, --help\n";
    assert!(help_has(help, "--generation"));
    assert!(help_has(help, "--met"));
    assert!(help_has(help, "--help"));
    assert!(!help_has(help, "--gen"), "a prefix is not the option");
    assert!(!help_has(help, "--evidence"));
}
