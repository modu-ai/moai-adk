//! Card C-53: MODE-AGENT and MODE-LEADER run their round on the local card ledger. S2 added one
//! section that mapped the Issue-based steps to cards; S5 rewrote the steps themselves on cards
//! and replaced that section with a short `Cards are the work ledger` section. Pinned here: the
//! section and its sentence, every card action named in both files, every
//! `moai card <action> --flag` either whole file shows checked against `--help` of this build,
//! the steps of the round, and the package's per-file bound.
//!
//! The binary runs with a temporary HOME; only `--help` is called and no ledger is made.

mod common;

use common::home::TempHome;
use std::{fs, path::Path};

const TITLE: &str = "Cards are the work ledger";
const WINS: &str = "GitHub Issues are no longer the work ledger; the project's local cards are \
                    (user decision 2026-11-16, card C-53)";
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

/// Actions only the leader workflow shows (the lanes never need them): allowed in a code span,
/// not required of both files.
const LEADER_ONLY: [&str; 1] = ["timeline"];

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
        if !ACTIONS.contains(&action.as_str()) && !LEADER_ONLY.contains(&action.as_str()) {
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

/// S5 (was S2's "this section wins over every Issue-based step"): the steps themselves are on
/// cards, so no step is left for an Issue section to override.
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
            "Pull requests, CI, the audit and the merge gate stay on GitHub",
        ] {
            assert!(section.contains(phrase), "{name} must say {phrase}");
        }
        let text = workflow(name);
        for gone in [
            "Issue-based",
            "Cards replace the Issue ledger",
            "lane-next",
            "queue-free",
            "moai-claim-handoff",
            "claim-acquire",
            "CLOSE-TABLE",
        ] {
            assert!(!text.contains(gone), "{name} still says `{gone}`");
        }
    }
}

#[test]
fn both_modes_name_every_card_action_of_the_round() {
    for name in ["mode-agent.md", "mode-leader.md"] {
        let named: Vec<String> = card_spans(&workflow(name))
            .into_iter()
            .map(|(action, _)| action)
            .collect();
        let missing: Vec<&str> = ACTIONS
            .iter()
            .filter(|a| !named.iter().any(|n| n == *a))
            .copied()
            .collect();
        assert!(missing.is_empty(), "{name} names no `card {missing:?}`");
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
    // S5: the whole file, since the card commands now sit in the steps (S2 checked one section).
    for name in ["mode-agent.md", "mode-leader.md"] {
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
             `card release C-1 --generation 1 --lease-secs 60`; `moai card-import-plan --from x`"
        ),
        [
            ("take".to_string(), "--holder".to_string()),
            ("steal".to_string(), String::new()),
            ("release".to_string(), "--lease-secs".to_string()),
        ]
    );
}

/// S5: the phrases are pinned in the whole file, where the steps now say them (S2 pinned them in
/// its one section).
fn assert_says(name: &str, phrases: &[&str]) {
    let text = workflow(name);
    for phrase in phrases {
        assert!(text.contains(phrase), "{name} must say `{phrase}`");
    }
}

/// The line of a numbered step, found by its key (`6. **claim-verify**`).
fn step(name: &str, key: &str) -> String {
    let text = workflow(name);
    let mark = format!(". **{key}**");
    text.lines()
        .find(|l| l.contains(&mark) && l.starts_with(|c: char| c.is_ascii_digit()))
        .unwrap_or_else(|| panic!("{name}: no step `{key}`"))
        .to_string()
}

/// A numbered step with its indented items (`self-pick` lists its order on lines of their
/// own), up to the next numbered step or section.
fn step_block(name: &str, key: &str) -> String {
    let text = workflow(name);
    let head = step(name, key);
    let start = text.find(&head).unwrap();
    let mut out = String::new();
    for (i, line) in text[start..].lines().enumerate() {
        let next_step = line.starts_with(|c: char| c.is_ascii_digit()) && line.contains(". **");
        if i > 0 && (next_step || line.starts_with("## ")) {
            break;
        }
        out.push_str(line);
        out.push('\n');
    }
    out
}

#[test]
fn the_agent_section_maps_each_step_of_the_round_to_a_card_command() {
    assert_says(
        "mode-agent.md",
        &[
            // Find, take, keep (S2: `each writer then takes`, `card renew` before the lease
            // ends, `card check … before each write`).
            "`moai card next`",
            "`moai card take <card> --scope <p> --worktree <w> --branch card/<n>-<slug>`",
            "under `~/.moai/worktrees/<project>-<n>` on branch `card/<n>-<slug>`",
            "Every writer takes the `--generation <g>` that `card take` answered",
            "run `card renew <card> --generation <g>` before the lease ends",
            "run `moai card check <card> --generation <g> --path <p>` with every path the change \
             touches; write only when it answers `allowed: true`",
            // Plan, reports, decisions (S2: `card note --kind report < file`, a decision note
            // needs `--by`: the lane reads decisions, the leader writes them).
            "`card note <card> --kind report --ref plan-C-<n>-<UTC> < plan.md` whose first line \
             is `PLAN C-<n>`",
            "RUN starts with the test that fails first (RED)",
            "The answer is a `--kind decision` note on the card naming your request (decision L3)",
            "`card show` prints each note's author and session — read both before you follow it",
            // Pull request, criteria, review, close. A local card names its host project
            // (audit N2).
            "first line `audit-tier: 0|1|2`",
            "Second line `card: C-<n>`",
            "Then a `## Self-Audit` section",
            "`card criterion <card> <n> --generation <g> --met --evidence <ref>`",
            "`--withdraw --reason <text>`",
            "`card pr <card> <pr> --generation <g> --repo <owner/repo>` (a card with no host \
             project needs `--repo`)",
            // User decision 2026-11-17: the lane starts its own audit agent at tier 0 and 1; at
            // tier 2 `review` is the request to the leader (was: always the leader's).
            "run `card review <card> --generation <g>`, then start your audit agent (`self-pick` \
             (2)); at tier 2 the card in `review` with an open pull request is your audit \
             request to the leader",
            // Repair N9: the criteria carry the new head's evidence before the next review.
            "On a FAIL: `card reopen <card> --generation <g>`, repair every blocking finding, \
             push, re-mark each criterion with the new head's evidence, and run `card review` \
             again",
            "run `card pr` again (it records the merge) and `card done <card> --generation <g>`",
            // Giving a card back or on.
            "`card release <card> --generation <g>` (never in `review`: `card reopen` first)",
            "a card reaches you only through its holder's `card handoff --to <your session>`",
            // Nothing to take, and a review card nobody holds.
            "`LANE-IDLE lane=<lane>",
            "never take a GitHub Issue instead",
            "`review_no_holder`",
            "Never edit the ledger file",
            // What a lane does not do.
            "`card add`, `card scope` and `card ready` are the leader's",
            // Until S3 the gate still asks for an Issue claim trailer: one sentence.
            "Until C-53 S3 the merge gate still asks for an Issue claim trailer, so the leader \
             merges your card's pull request under the transition exception (MODE-LEADER)",
        ],
    );
    // The parts that still read Issue claims: the commit guard lets a card branch through as
    // "not a lane worktree" (`claim_verify.rs`, audit N7). The guard answers; it is not not_run.
    let guard = step("mode-agent.md", "claim-verify");
    assert!(
        guard.contains("the commit guard (`moai claim-verify --hook`) answers `allowed` on a card branch, checking nothing")
            && guard.contains("C-53 S3"),
        "{guard}"
    );
    assert!(
        !guard.contains("`not_run`"),
        "the guard answers, it is not not_run"
    );
    // One sentence names the transition exception in MODE-AGENT.
    assert_eq!(
        workflow("mode-agent.md")
            .matches("transition exception")
            .count(),
        1
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
            // Audits (user decision 2026-11-17): tier 2 stays with the leader's own read-only
            // agent, never a lane and never the implementer; tier 0 and 1 are the lane's own
            // agent, whose verdict the leader checks before merging. The S5 head widened the
            // leader's assignment to "or to a lane that did not write it" and the old pin
            // ("read-only audit agent") still passed.
            "At audit-tier 2, when a card is `review` and its pull request is open, assign the \
             audit to one read-only audit agent you start (`moai-audit`), never to a lane and \
             never the session that implemented it",
            "at audit-tier 0 and 1 the lane starts a fresh read-only `moai-audit` agent for its \
             own pull request and posts its verdict",
            "before merging, check that the verdict line has the AUDIT format, names that agent, \
             cites the current head and covers the card's criteria",
            "No lane audits another lane's pull request.",
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
    let leader = workflow("mode-leader.md");
    for widened in [
        "or to a lane that did not write it",
        "starting an audit and changing code",
        "Before dispatching a card or an audit",
    ] {
        assert!(
            !leader.contains(widened),
            "mode-leader.md hands a formal audit to a lane: `{widened}`"
        );
    }
}

/// Repair of the C-53 S5 audit: B1 as changed by the user decision of 2026-11-17 (the lane
/// starts its own fresh `moai-audit` agent at tier 0 and 1, the same agent re-audits, tier 2
/// stays with the leader, no lane audits another lane's pull request), B2 (a `mine` card in
/// `review` is not the next card: `card next` answers `next_is: "mine"` for it), N1 (no session
/// id on OMP), N2 (the PLAN ref date), N4 (the lane's limits), N5 (non-Issue rules restored) and
/// N6 (host smoke without the lifted HOLD).
#[test]
fn the_round_keeps_audits_with_the_leader_and_moves_on_from_a_card_in_review() {
    // B1 / user decision 2026-11-17: (a) own agent at tier 0 and 1, (b) the same agent
    // re-audits, (c) tier 2 is the leader's, (d) never another lane's pull request.
    let pick = step_block("mode-agent.md", "self-pick");
    assert!(
        pick.contains("    - (1) ") && pick.contains("    - (4) "),
        "{pick}"
    );
    for p in [
        // (a)
        // Core rule 17: the decision names the record it replaces (11-16, #1975 6904941797).
        "(2) Your own pull request's audit (user decision 2026-11-17, replacing #1975 \
         6904941797): at audit-tier 0 or 1 start one fresh-context, read-only `moai-audit` \
         subagent",
        "never your conclusions or self-assessment",
        "your own context never audits",
        // (b)
        "A re-audit after FAIL goes to the same agent (SendMessage or resume), never a new one",
        // (c)
        "Tier 2 audits are the leader's",
        // (d), and N4's "you did not author or repair".
        "Never audit another lane's pull request: one you did not author or repair is not yours \
         to audit or change",
    ] {
        assert!(pick.contains(p), "self-pick must say `{p}`");
    }
    assert!(!pick.contains("Formal audits go to the leader's audit agent"));
    let agent = workflow("mode-agent.md");
    for gone in [
        "An audit the leader dispatched to you",
        "audit worktree",
        "except an audit reservation",
        "When you audit",
    ] {
        assert!(!agent.contains(gone), "mode-agent.md still says `{gone}`");
    }
    // B2: in the claim step and in the self-pick order, and the hand-off goes on.
    const REVIEW_WAITS: &str =
        "a `mine` card in `review` waits on its audit; take the first `ready` card instead";
    assert!(step("mode-agent.md", "claim").contains(REVIEW_WAITS));
    assert!(pick.contains(REVIEW_WAITS));
    assert!(step("mode-agent.md", "hand-off")
        .contains("Then do not wait for the merge: follow `Waiting is switching`"));
    // N1 (re-audit N-r1: the line has the `owner=` and `next=` the report step asks for).
    assert!(step("mode-agent.md", "session-id").contains(
        "refuses a session without one: there (OMP) report `LANE-BLOCKED <lane> card-take no \
         session id (OMP) · kind=host owner=user next=card C-62` once"
    ));
    // N2: the date gives the seven digits in a row `pr-self-audit-check` wants on `plan:`.
    assert!(step("mode-agent.md", "pr-gate").contains(
        "the PLAN note, by its `ref`, whose `<UTC>` (`date -u +%Y%m%dT%H%M%SZ`) gives the seven \
         digits the check wants"
    ));
    // N4: the lane's limits.
    let limits = section_of_agent("Exceptions and Permissions");
    assert!(limits.contains(
        "Never decide a HOLD, merge, or post an audit status; those belong to the leader and the \
         audit gate"
    ));
    // N5: restored without the Issue words.
    for (key, p) in [
        (
            "session-id",
            "a duplicate id stops the round (SPEC §4, item 6)",
        ),
        (
            "pr-gate",
            "an auditor or the leader may raise it, and a missing or malformed line counts as 2",
        ),
        (
            "pr-gate",
            "a shared file only under `docs/**` that merges cleanly, keeping both sides' lines, \
             does not count",
        ),
        (
            "heavy-checks",
            // C-82: the worktree's build folder (was: the lane's one build folder).
            "Every build, a self-audit's too, goes to the worktree's build folder",
        ),
        (
            "report",
            "Independent audit, required CI and HOLD stay as they are.",
        ),
    ] {
        assert!(
            step("mode-agent.md", key).contains(p),
            "{key} must say `{p}`"
        );
    }
    assert!(agent.contains("re-post nothing and do not rewrite a pull request body"));
    // N6: both host smokes are `not_run` without the lifted HOLD of user decision 1.
    for name in ["mode-agent.md", "mode-leader.md"] {
        let text = workflow(name);
        assert!(
            text.contains("\nNot run yet (`not_run`): in a clean clone with a registered ledger"),
            "{name}"
        );
        assert!(!text.contains("user decision 1, HOLD"), "{name}");
    }
}

/// N3, user decision 2026-11-17 ("CI 성공 + PASS"), which replaces user decisions 6877084436
/// and 6878414880: a merge needs the head's GitHub `ci` success and, at audit-tier 1 or 2, an
/// independent PASS; a `local-check:` line is optional and never replaces `ci`. Since card C-84
/// rule G5 of `moai gate-check` reads the exact head's `ci`; both modes say so (was: "until
/// card C-84 changes G5 … still posted").
#[test]
fn a_merge_needs_ci_success_and_a_pass_and_a_local_check_line_never_replaces_ci() {
    let train = step("mode-leader.md", "merge-train");
    for p in [
        "A merge needs the pull request head's GitHub `ci` success and, at effective audit tier \
         1 or 2, a non-author PASS that cites that run",
        // Re-audit N-r3 (leader decision 2026-11-17): the whole tier-0 sentence, so an added
        // alternative ("… or a `ci` run", the auditor's M5) fails; a FAIL stops every tier.
        "at effective tier 0 (documents, wording and tests only, after the gate's path rule) the \
         `ci` success alone, when no FAIL is posted. A posted FAIL verdict stops the merge at \
         every tier (leader decision 2026-11-17).",
        "A `local-check:` line is optional and never replaces `ci` (user decision 2026-11-17 \"CI \
         성공 + PASS\", which replaces user decisions 6877084436 and 6878414880)",
        // C-84 landed (was: "Until card C-84 changes rule G5, … still stops on a missing
        // local-check line").
        "`moai gate-check` rule G5 reads that `ci` result on the exact head (card C-84).",
    ] {
        assert!(train.contains(p), "mode-leader merge-train must say `{p}`");
    }
    assert!(!train.contains("GitHub CI on the pull request does not block the merge"));
    // The transition exception follows the same rule.
    assert!(workflow("mode-leader.md").contains(
        "merge a card's pull request under the `merge-train` rule (the head's GitHub `ci` \
         success and, at audit-tier 1 or 2, an independent PASS)"
    ));
    assert!(step("mode-agent.md", "merge-train").contains(
        "A merge needs the head's GitHub `ci` success and, at audit-tier 1 or 2, a non-author \
         PASS that cites that run; a `local-check:` line never replaces `ci` (user decision \
         2026-11-17)"
    ));
    let gate = step("mode-agent.md", "pr-gate");
    // N-r2, C-84: the same rule as mode-leader — the line is optional, G5 reads the head's `ci`.
    assert!(gate.contains(
        "as a pull request comment: optional evidence; `moai gate-check` (G5) reads the head's \
         `ci` (C-84)."
    ));
    for text in [workflow("mode-agent.md"), workflow("mode-leader.md")] {
        assert!(!text.contains("stops on a missing local-check line"));
    }
    // N-r3: the lane's tier-0 sentence, whole.
    assert!(gate.contains(
        "0 for documents, wording and tests only, which merge on the `ci` success alone when no \
         FAIL is posted (a posted FAIL stops a merge at every tier);"
    ));
    assert!(!gate.contains("the merge gate reads that line, not GitHub CI"));
}

/// The text of one `## ` section of MODE-AGENT.
fn section_of_agent(title: &str) -> String {
    let text = workflow("mode-agent.md");
    let head = format!("\n## {title}\n");
    let start = text.find(&head).expect("the section") + head.len();
    let rest = &text[start..];
    rest[..rest.find("\n## ").unwrap_or(rest.len())].to_string()
}

/// Every package file is read with the bound `profile::LIMIT` (`profile::read`, used by
/// `package_digest` for install, update and `package-inspect`). MODE-AGENT was 64,181 bytes
/// before its S2 card section and 65,509 after it; S5 rewrote the steps on cards and removed
/// the Issue text. A file over the bound makes the whole package unreadable ("target is not a
/// bounded regular file").
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
