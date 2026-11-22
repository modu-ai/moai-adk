//! Card C-53: MODE-AGENT and MODE-LEADER ran their round on the local card ledger. Both left the
//! package with their routes (SPEC-20261122-next-flow-skills-agents REQ-2); the file stays because
//! `skills/moai/references/rule-history.md` names it (D-11). Pinned here now: every
//! `moai card <action> --flag` the WORK workflow shows, checked against `--help` of this build.
//!
//! The binary runs with a temporary HOME; only `--help` is called and no ledger is made.

mod common;

use common::home::TempHome;
use std::{fs, path::Path};

/// The card actions the WORK workflow uses.
const ACTIONS: [&str; 6] = ["next", "take", "check", "criterion", "land", "note"];

fn workflow(name: &str) -> String {
    fs::read_to_string(
        Path::new(env!("CARGO_MANIFEST_DIR"))
            .join("skills/moai/workflows")
            .join(name),
    )
    .unwrap()
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

/// Re-pointed to WORK (SPEC-20261122-next-flow-skills-agents REQ-3); the name stays.
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
    // Re-pointed to WORK with the retired modes (SPEC-20261122-next-flow-skills-agents REQ-3).
    {
        let name = "work.md";
        let text = workflow(name);
        assert!(
            card_spans(&text).len() >= ACTIONS.len(),
            "{name}: the file shows its commands as code spans"
        );
        assert_eq!(
            unknown(&home, &text),
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
             `card check C-1 --generation 1 --lease-secs 60`; `moai card-import-plan --from x`"
        ),
        [
            ("take".to_string(), "--holder".to_string()),
            ("steal".to_string(), String::new()),
            ("check".to_string(), "--lease-secs".to_string()),
        ]
    );
}

// Retired with the MODE-AGENT route (SPEC-20261122-next-flow-skills-agents REQ-2): `both_modes_carry_one_card_section_that_wins_where_a_ledger_is_registered` pinned the card section of both mode files.

// Retired with the MODE-AGENT route (SPEC-20261122-next-flow-skills-agents REQ-2): `both_modes_name_every_card_action_of_the_round` pinned every card action of the lane and leader rounds.

// Retired with the MODE-AGENT route (SPEC-20261122-next-flow-skills-agents REQ-2): `the_agent_section_maps_each_step_of_the_round_to_a_card_command` pinned the lane round step by step; "never your conclusions or self-assessment" returns with the audit contract.

// Retired with the MODE-LEADER route (SPEC-20261122-next-flow-skills-agents REQ-2): `the_leader_section_makes_cards_ready_assigns_audits_and_names_the_transition` pinned the leader's card steps.

// Retired with the MODE-AGENT route (SPEC-20261122-next-flow-skills-agents REQ-2): `the_round_keeps_audits_with_the_leader_and_moves_on_from_a_card_in_review` pinned the lane's audit tiers and review waits.

// Retired with the MODE-AGENT route (SPEC-20261122-next-flow-skills-agents REQ-2): `a_merge_needs_ci_success_and_a_pass_and_a_local_check_line_never_replaces_ci` pinned the pull-request merge gate of both modes, and WORK lands with `moai card land`.

// Retired with the MODE-AGENT route (SPEC-20261122-next-flow-skills-agents REQ-2): `both_mode_files_stay_within_the_package_file_bound` bounded the two mode files, which are gone.

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
