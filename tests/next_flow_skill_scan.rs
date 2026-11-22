//! SPEC-20261122-next-flow-skills-agents, text slice A (AC-1, AC-2, AC-5): the `/moai` skill
//! package has eight routes, `/moai work` takes a card through the whole flow, and the retired
//! SETUP, VERIFY and OPTIMIZE routes leave their rules in PROJECT, RUN and UPDATE.
//!
//! The tests read the skill files of this checkout; nothing is run and no HOME is touched.

use serde_json::Value;
use std::{
    fs,
    path::{Path, PathBuf},
};

fn package() -> PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai")
}

fn read(rel: &str) -> String {
    fs::read_to_string(package().join(rel)).unwrap_or_else(|e| panic!("{rel}: {e}"))
}

fn routes() -> Value {
    serde_json::from_str(&read("references/routes.json")).expect("routes.json is JSON")
}

fn route_keys() -> Vec<String> {
    let mut keys: Vec<String> = routes()["routes"]
        .as_object()
        .expect("routes.json has a `routes` object")
        .keys()
        .cloned()
        .collect();
    keys.sort();
    keys
}

/// REQ-1: the eight routes, sorted.
const ROUTES: [&str; 8] = [
    "AUDIT", "PLAN", "PROJECT", "READ", "RUN", "SYNC", "UPDATE", "WORK",
];

/// REQ-2, the files text slice A removes. The list grows: slice B adds
/// `references/internal/build-folder.md` and `references/internal/review.md`, and Plan step 6
/// adds `references/rule-history.md` when the rule history moves to `docs/`.
const REMOVED: [&str; 6] = [
    "workflows/mode-agent.md",
    "workflows/mode-leader.md",
    "workflows/setup.md",
    "workflows/verify.md",
    "workflows/optimize.md",
    "references/adapters/herdr.md",
];

/// REQ-4: the mode spellings SKILL.md no longer names.
const MODE_SPELLINGS: [&str; 5] = [
    "--mode-agent",
    "--mode-leader",
    "--autonomy",
    "MODE-AGENT",
    "MODE-LEADER",
];

/// REQ-4: the one sentence of SKILL.md that makes a dash option plain request text. It names
/// no retired flag (REQ-8).
const DASH_SENTENCE: &str =
    "A word of the request that is not a route, a dash option included, is request text and picks no route.";

/// REQ-3: the sentence of SKILL.md that picks WORK by the leading word.
const WORK_SENTENCE: &str = "A request whose first word after the skill name is `work` picks WORK";

/// REQ-3: the WORK procedure's step keys, in order.
const WORK_STEPS: [&str; 7] = ["take", "plan", "run", "sync", "audit", "land", "repeat"];

/// The mode spellings `text` holds, in the order of `MODE_SPELLINGS`.
fn mode_spellings(text: &str) -> Vec<&'static str> {
    MODE_SPELLINGS
        .iter()
        .copied()
        .filter(|s| text.contains(s))
        .collect()
}

/// The body of the `## <name>` section of `text`, up to the next `## ` heading.
fn section<'a>(text: &'a str, name: &str) -> &'a str {
    let head = format!("\n## {name}\n");
    let at = text
        .find(&head)
        .unwrap_or_else(|| panic!("no `## {name}` section"));
    let body = &text[at + head.len()..];
    &body[..body.find("\n## ").unwrap_or(body.len())]
}

/// The numbered steps of a procedure written as `N. **key** — …`: (number, key) per step.
fn step_keys(procedure: &str) -> Vec<(usize, String)> {
    procedure
        .lines()
        .filter_map(|line| {
            let (number, rest) = line.split_once(". **")?;
            let number: usize = number.parse().ok()?;
            let (key, after) = rest.split_once("** — ")?;
            (!after.is_empty()).then(|| (number, key.to_string()))
        })
        .collect()
}

/// The route cells of the chain table, one list per row.
fn chain_rows(text: &str) -> Vec<Vec<String>> {
    text.lines()
        .filter(|l| l.starts_with("| ") && !l.starts_with("| Request") && !l.starts_with("| ---"))
        .map(|l| {
            let cells: Vec<&str> = l.trim_matches('|').split('|').map(str::trim).collect();
            assert_eq!(cells.len(), 4, "a chain row has four cells: {l}");
            cells[1].split(", ").map(str::to_string).collect()
        })
        .collect()
}

#[test]
fn routes_table_names_exactly_the_eight_routes() {
    assert_eq!(route_keys(), ROUTES, "the route table");
}

#[test]
fn routes_every_reference_is_a_shipped_file() {
    let routes = routes();
    for key in route_keys() {
        let want = format!("../workflows/{}.md", key.to_lowercase());
        assert_eq!(
            routes["routes"][&key]["reference"],
            Value::String(want.clone()),
            "{key}"
        );
        let rel = want.strip_prefix("../").unwrap();
        assert!(package().join(rel).is_file(), "{key}: {rel} is not shipped");
    }
}

#[test]
fn routes_removed_files_are_gone() {
    let left: Vec<&str> = REMOVED
        .iter()
        .copied()
        .filter(|rel| package().join(rel).exists())
        .collect();
    assert!(left.is_empty(), "still in the package: {left:?}");
}

#[test]
fn routes_skill_entry_names_no_mode_flag() {
    // The checker finds a planted spelling, so an empty result below means none is there.
    assert_eq!(
        mode_spellings("pick `/moai --mode-agent` or MODE-LEADER"),
        ["--mode-agent", "MODE-LEADER"]
    );
    assert_eq!(
        mode_spellings(&read("SKILL.md")),
        Vec::<&str>::new(),
        "SKILL.md"
    );
    let modes: Vec<String> = route_keys()
        .into_iter()
        .filter(|k| k.starts_with("MODE-"))
        .collect();
    assert!(modes.is_empty(), "mode routes: {modes:?}");
}

#[test]
fn routes_dash_options_are_plain_text() {
    assert_eq!(
        read("SKILL.md").matches(DASH_SENTENCE).count(),
        1,
        "SKILL.md holds the dash-option sentence once"
    );
}

#[test]
fn routes_chain_table_uses_only_table_routes() {
    let rows = chain_rows(&read("references/chains/index.md"));
    assert!(!rows.is_empty(), "the chain table has rows");
    for row in &rows {
        for route in row {
            assert!(
                ROUTES.contains(&route.as_str()),
                "the chain table names {route}"
            );
        }
    }
    assert!(
        rows.iter().any(|r| r.iter().any(|route| route == "WORK")),
        "the chain table has a WORK row"
    );
}

#[test]
fn work_steps_run_in_order() {
    let work = read("workflows/work.md");
    let steps = step_keys(section(&work, "Procedure"));
    let numbers: Vec<usize> = steps.iter().map(|(n, _)| *n).collect();
    let keys: Vec<&str> = steps.iter().map(|(_, k)| k.as_str()).collect();
    assert_eq!(keys, WORK_STEPS, "the WORK steps");
    assert_eq!(numbers, (1..=WORK_STEPS.len()).collect::<Vec<_>>());
}

#[test]
fn work_takes_with_card_next_and_lands_with_card_land() {
    let work = read("workflows/work.md");
    for command in [
        "`moai card next`",
        "`moai card take",
        "`moai card check <card> --generation <g> --path <p>`",
        "`moai card land <card> --head <40 hex>`",
    ] {
        assert!(work.contains(command), "work.md lacks {command}");
    }
}

#[test]
fn work_repeats_until_no_card_is_left() {
    let work = read("workflows/work.md");
    let repeat = section(&work, "Procedure")
        .lines()
        .find(|l| l.contains(". **repeat** — "))
        .expect("work.md has a repeat step");
    assert!(
        repeat.contains("until no card is left"),
        "the repeat step: {repeat}"
    );
}

#[test]
fn work_decision_becomes_a_blocked_note_on_the_card() {
    let work = read("workflows/work.md");
    for phrase in [
        "`LANE-BLOCKED kind=decision",
        "as a note on the card",
        "never a question to the user",
    ] {
        assert!(work.contains(phrase), "work.md lacks {phrase:?}");
    }
    for tool in ["AskUserQuestion", "request_user_input", "question tool"] {
        assert!(!work.contains(tool), "work.md names {tool}");
    }
}

#[test]
fn work_is_picked_by_the_leading_word() {
    let skill = read("SKILL.md");
    assert!(
        skill.contains("`/moai work`"),
        "SKILL.md names `/moai work`"
    );
    assert!(
        skill.contains(WORK_SENTENCE),
        "SKILL.md lacks {WORK_SENTENCE:?}"
    );
}

#[test]
fn absorb_project_carries_the_setup_steps() {
    let project = read("workflows/project.md");
    for phrase in [
        "record the first work item with `moai card add`",
        "creating it is a write under the real HOME",
        "run `moai ledger-register-apply` only after its own pick in a selection",
        "Allowed initial local commit",
        "it never implies push or Issue publication",
        "existing commit is never amended",
    ] {
        assert!(project.contains(phrase), "project.md lacks {phrase:?}");
    }
}

#[test]
fn absorb_project_asks_no_language_the_user_scope_holds() {
    let project = read("workflows/project.md");
    for phrase in [
        "the local tool language question is the instruction-file language",
        "asks no conversation, document or commit language that the user scope already holds",
    ] {
        assert!(project.contains(phrase), "project.md lacks {phrase:?}");
    }
}

#[test]
fn absorb_update_carries_a_leftover_preview_that_deletes_nothing() {
    let update = read("workflows/update.md");
    for phrase in [
        "`moai optimize-plan",
        "deletes, moves and writes nothing",
        "Deleting is a `guarded` choice",
    ] {
        assert!(update.contains(phrase), "update.md lacks {phrase:?}");
    }
}

#[test]
fn absorb_run_carries_the_verify_sections() {
    let run = read("workflows/run.md");
    for name in [
        "Check-run rules",
        "Lane worktrees and build cache",
        "Change-specific rules",
        "Shell and helper rules",
    ] {
        assert_eq!(
            run.matches(&format!("\n## {name}\n")).count(),
            1,
            "run.md holds `## {name}` once"
        );
    }
}

#[test]
fn absorb_entry_names_no_retired_route() {
    for rel in ["SKILL.md", "references/chains/index.md"] {
        let text = read(rel);
        for word in ["SETUP", "VERIFY", "OPTIMIZE", "/moai optimize"] {
            assert!(!text.contains(word), "{rel} names {word}");
        }
    }
}

#[test]
fn step_keys_and_sections_read_what_they_say() {
    let procedure =
        "1. **take** — a\n   - 2. **no** — nested\n2. **plan** — b\n3. **x**: no dash\n";
    assert_eq!(
        step_keys(procedure),
        [(1, "take".to_string()), (2, "plan".to_string())]
    );
    let text = "# T\n\n## Procedure\n\n1. **take** — a\n\n## Next\n\nz\n";
    assert_eq!(section(text, "Procedure"), "\n1. **take** — a\n");
}
