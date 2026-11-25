//! SPEC-20261125-run-sync-procedure-receipt: `workflows/run.md` is one numbered procedure and a
//! failure table, `workflows/sync.md` keeps the lane part, the long rules live in four internal
//! references and one flow reference, the receipt rule is stated once, and the gate sentences stay.
//!
//! The texts are read from the checked-in tree. The only process started is the binary's own
//! `--help`, with a temporary HOME and a clean environment (`TempHome::command`). Nothing depends
//! on the clock or the network.

mod common;

use common::home::TempHome;
use sha2::{Digest, Sha256};
use std::collections::BTreeMap;
use std::path::{Path, PathBuf};

const RUN: &str = "workflows/run.md";
const SYNC: &str = "workflows/sync.md";
const SELF_CHECK: &str = "references/internal/self-check.md";
const TEST_DESIGN: &str = "references/internal/test-design.md";
const CHECK_RULES: &str = "references/internal/check-rules.md";
const BUILD_SLOTS: &str = "references/internal/build-slots.md";
const SHELL_GUARDS: &str = "references/internal/shell-guards.md";
const WEB_PAGES: &str = "references/internal/web-pages.md";
const LANDING: &str = "references/flows/landing-release.md";
const AUDIT_CONTRACT: &str = "references/audit.md";

/// The files the SPEC edits (paths under `skills/moai/`).
const EDITED: [&str; 9] = [
    RUN,
    SYNC,
    SELF_CHECK,
    TEST_DESIGN,
    "references/internal/tdd-results.md",
    "references/internal/check-completeness.md",
    "references/internal/doc-claims.md",
    "references/internal/boundary-checks.md",
    "references/internal/security-checklist.md",
];

/// The files the SPEC adds (paths under `skills/moai/`).
const NEW: [&str; 5] = [CHECK_RULES, BUILD_SLOTS, SHELL_GUARDS, WEB_PAGES, LANDING];

/// The gate sentences of `## Exceptions and Permissions` in `workflows/run.md`, word for word.
const RUN_GATES: [&str; 4] = [
    "The session that wrote a change never audits it.",
    "The auditing session records its own verdict with `moai card verdict`; the author never records it.",
    "A command that was not run is `not_run`, never a pass.",
    "A HOLD or a `guarded` choice is reported as blocked and never decided by a lane.",
];

/// The gate sentence of `references/internal/self-check.md` (tests are never Python).
const SELF_CHECK_GATE: &str = "Tests are Rust only.";

/// The one receipt sentence fragment: it must stand once, in the receipt step of `run.md`.
const RECEIPT_PHRASE: &str = "never run it again on a tree that already passed it";

fn skill_path(rel: &str) -> PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR"))
        .join("skills/moai")
        .join(rel)
}

/// The text of a file under `skills/moai/`, or `None` while the file does not exist yet.
fn read_opt(rel: &str) -> Option<String> {
    std::fs::read_to_string(skill_path(rel)).ok()
}

fn text(rel: &str) -> String {
    read_opt(rel).unwrap_or_else(|| panic!("skills/moai/{rel} is missing or not UTF-8"))
}

/// Every text file under `skills/` and `agents/`, with its path relative to the repository root.
fn shipped_texts() -> Vec<(String, String)> {
    fn walk(dir: &Path, out: &mut Vec<PathBuf>) {
        let Ok(entries) = std::fs::read_dir(dir) else {
            return;
        };
        for entry in entries.flatten() {
            let Ok(kind) = entry.file_type() else {
                continue;
            };
            if kind.is_dir() {
                walk(&entry.path(), out);
            } else if kind.is_file() {
                out.push(entry.path());
            }
        }
    }
    let root = Path::new(env!("CARGO_MANIFEST_DIR"));
    let mut paths = Vec::new();
    for top in ["skills", "agents"] {
        walk(&root.join(top), &mut paths);
    }
    paths.sort();
    paths
        .into_iter()
        .filter_map(|p| {
            let body = std::fs::read_to_string(&p).ok()?;
            let rel = p.strip_prefix(root).ok()?.to_string_lossy().into_owned();
            Some((rel, body))
        })
        .collect()
}

/// The lines of `text` that are not inside a fenced code block (the fence lines are dropped too).
fn outside_fences(text: &str) -> Vec<&str> {
    let mut inside = false;
    let mut out = Vec::new();
    for line in text.lines() {
        if line.trim_start().starts_with("```") {
            inside = !inside;
            continue;
        }
        if !inside {
            out.push(line);
        }
    }
    out
}

/// The lines of the `## <heading>` section, up to the next `## ` heading.
fn section<'a>(lines: &[&'a str], heading: &str) -> Vec<&'a str> {
    let mut out = Vec::new();
    let mut inside = false;
    for line in lines {
        if let Some(h) = line.strip_prefix("## ") {
            inside = h.trim() == heading;
            continue;
        }
        if inside {
            out.push(*line);
        }
    }
    out
}

fn section_text(text: &str, heading: &str) -> String {
    section(&outside_fences(text), heading).join("\n")
}

/// The number of a numbered list item line (`1. `, `10. `) at the margin.
fn item_number(line: &str) -> Option<usize> {
    let digits = line.chars().take_while(char::is_ascii_digit).count();
    if digits > 0 && line[digits..].starts_with(". ") {
        line[..digits].parse().ok()
    } else {
        None
    }
}

/// Collapses every run of whitespace, so a sentence matches when the text wraps it.
fn squash(s: &str) -> String {
    s.split_whitespace().collect::<Vec<_>>().join(" ")
}

/// `needle` stands in `hay` with a non-alphanumeric character (or an edge) on both sides.
fn has_word(hay: &str, needle: &str) -> bool {
    let bytes = hay.as_bytes();
    hay.match_indices(needle).any(|(i, _)| {
        let before = i == 0 || !bytes[i - 1].is_ascii_alphanumeric();
        let end = i + needle.len();
        let after = end >= bytes.len() || !bytes[end].is_ascii_alphanumeric();
        before && after
    })
}

/// AC-1 (REQ-1, REQ-7): the Procedure is one numbered list whose steps name the actions in order.
#[test]
fn ac1_the_procedure_is_one_numbered_list_in_work_order() {
    let run = text(RUN);
    let lines = outside_fences(&run);
    let procedure = section(&lines, "Procedure");

    // Group the Procedure into items: a numbered line starts one, later lines continue it.
    let mut items: Vec<(usize, String)> = Vec::new();
    for line in &procedure {
        if let Some(n) = item_number(line) {
            items.push((n, (*line).to_string()));
        } else if let Some(last) = items.last_mut() {
            last.1.push('\n');
            last.1.push_str(line);
        }
    }
    let numbers: Vec<usize> = items.iter().map(|(n, _)| *n).collect();
    let expected: Vec<usize> = (1..=items.len()).collect();
    assert!(
        items.len() >= 8 && numbers == expected,
        "the Procedure must be one list numbered 1..n with at least 8 steps, found {numbers:?}"
    );
    let in_file = lines.iter().filter(|l| item_number(l).is_some()).count();
    assert_eq!(
        in_file,
        items.len(),
        "every numbered item of run.md stands in the Procedure"
    );

    // Each key names its action in a step at or after the step of the key before it.
    let keys = [
        "moai card check",
        "internal/test-design",
        "internal/tdd-results",
        "moai-impl",
        "verify:",
        "internal/self-check",
        "moai check affected --card",
        "SYNC",
    ];
    let mut at = 0;
    for key in keys {
        match items
            .iter()
            .skip(at)
            .position(|(_, body)| body.contains(key))
        {
            Some(offset) => at += offset,
            None => panic!(
                "no Procedure step at or after step {} names `{key}` (order: {keys:?})",
                at + 1
            ),
        }
    }

    // The conditional references stand inside the Procedure, in the step they belong to.
    let whole = procedure.join("\n");
    for reference in [
        "internal/check-completeness",
        "internal/boundary-checks",
        "internal/security-checklist",
    ] {
        assert!(
            whole.contains(reference),
            "the Procedure does not name {reference}"
        );
    }
    assert!(
        text(SYNC).contains("internal/doc-claims"),
        "sync.md does not name internal/doc-claims"
    );
}

/// AC-2 (REQ-2, REQ-3): at most 1,200 words, exactly eight sections, nine failure rows.
#[test]
fn ac2_run_is_short_with_eight_sections_and_a_nine_row_table() {
    let run = text(RUN);
    let words = run.split_whitespace().count();
    assert!(
        words <= 1200,
        "run.md holds {words} words, the limit is 1200"
    );

    let lines = outside_fences(&run);
    let headings: Vec<&str> = lines
        .iter()
        .filter_map(|l| l.strip_prefix("## "))
        .map(str::trim)
        .collect();
    assert_eq!(
        headings,
        [
            "Input",
            "Output",
            "Procedure",
            "When a step stops",
            "Base agents",
            "Resume",
            "Exceptions and Permissions",
            "Completion Check",
        ],
        "the second-level headings of run.md"
    );

    let table = section(&lines, "When a step stops");
    let rows: Vec<&str> = table
        .iter()
        .copied()
        .filter(|l| l.trim_start().starts_with('|'))
        .collect();
    let separators = rows
        .iter()
        .filter(|l| l.chars().all(|c| matches!(c, '|' | '-' | ':' | ' ')))
        .count();
    let body = rows.len().saturating_sub(separators + 1);
    assert_eq!(
        body, 9,
        "body rows of the table under `## When a step stops`"
    );
}

/// AC-3 (REQ-4): the receipt sentence stands once, in the Procedure of `run.md`; the rest point to it.
#[test]
fn ac3_the_receipt_rule_is_stated_once_and_pointed_to() {
    let scanned: Vec<&str> = [RUN, SYNC, SELF_CHECK, TEST_DESIGN]
        .into_iter()
        .chain(NEW)
        .collect();
    let bodies: BTreeMap<&str, String> = scanned
        .iter()
        .map(|rel| (*rel, read_opt(rel).unwrap_or_default()))
        .collect();

    let total: usize = bodies
        .values()
        .map(|b| b.matches(RECEIPT_PHRASE).count())
        .sum();
    assert_eq!(
        total, 1,
        "`{RECEIPT_PHRASE}` must appear exactly once across {scanned:?}"
    );
    assert!(
        text(RUN)
            .lines()
            .any(|l| l.contains(RECEIPT_PHRASE) && l.contains("moai check affected --card")),
        "the receipt sentence is not on one line with `moai check affected --card` in run.md"
    );
    assert!(
        section_text(&text(RUN), "Procedure").contains(RECEIPT_PHRASE),
        "the receipt sentence is not in the Procedure of run.md"
    );

    for (rel, body) in &bodies {
        let lower = body.to_lowercase();
        for banned in ["never narrowed", "full check set"] {
            assert!(!lower.contains(banned), "{rel} still says {banned:?}");
        }
        let wide = lower.matches("wide set").count();
        match *rel {
            RUN => {
                for line in body
                    .lines()
                    .filter(|l| l.to_lowercase().contains("wide set"))
                {
                    assert!(
                        line.contains(RECEIPT_PHRASE),
                        "run.md says `wide set` outside the receipt sentence: {line}"
                    );
                }
            }
            CHECK_RULES => assert!(wide <= 1, "check-rules.md says `wide set` {wide} times"),
            _ => assert_eq!(
                wide, 0,
                "{rel} says `wide set`; it points to the receipt step"
            ),
        }
    }

    for rel in [SELF_CHECK, SYNC, CHECK_RULES] {
        let body = &bodies[rel];
        assert!(
            body.contains("step 7") || body.contains("receipt step"),
            "{rel} does not point to the receipt step (`step 7`) of the RUN workflow"
        );
    }
}

/// AC-4 (REQ-5): four internal references and the flow reference exist, are in the manifest with an
/// origin, hold their probes, run.md names each, and no old section name or tier line stays behind.
#[test]
fn ac4_the_long_rules_have_their_new_homes() {
    let mut problems: Vec<String> = Vec::new();

    let manifest: serde_json::Value =
        serde_json::from_str(&text("references/source-manifest.json"))
            .expect("source-manifest.json is JSON");
    for rel in NEW {
        if read_opt(rel).is_none() {
            problems.push(format!("{rel} does not exist"));
        }
        match manifest["files"].get(rel) {
            Some(entry) if entry.get("origin").is_some() => {}
            Some(_) => problems.push(format!("manifest entry of {rel} has no `origin`")),
            None => problems.push(format!(
                "{rel} is not listed in source-manifest.json `files`"
            )),
        }
    }

    for (rel, probe) in [
        (CHECK_RULES, "moai flake-classify"),
        (BUILD_SLOTS, "limits.json"),
        (BUILD_SLOTS, "moai heavy-slot-acquire"),
        (SHELL_GUARDS, "names git in a form too complex to verify"),
        (WEB_PAGES, "Chrome"),
    ] {
        match read_opt(rel) {
            Some(body) if body.contains(probe) => {}
            Some(_) => problems.push(format!("{rel} lacks the rule text `{probe}`")),
            None => problems.push(format!("{rel} is missing, so `{probe}` has no home")),
        }
    }

    let run = text(RUN);
    for name in [
        "internal/check-rules",
        "internal/build-slots",
        "internal/shell-guards",
        "internal/web-pages",
    ] {
        if !run.contains(name) {
            problems.push(format!("run.md does not name {name}"));
        }
    }

    let shipped = shipped_texts();
    for (rel, body) in &shipped {
        if body.contains("Check-run rules") {
            problems.push(format!("{rel} still names `Check-run rules`"));
        }
    }

    let contract = format!("skills/moai/{AUDIT_CONTRACT}");
    for prefix in [
        "- docs-only (tier",
        "- test-only (tier",
        "- code (tier",
        "- high-risk (tier",
    ] {
        let holders: Vec<&str> = shipped
            .iter()
            .filter(|(_, body)| body.contains(prefix))
            .map(|(rel, _)| rel.as_str())
            .collect();
        // The audit agent file carries a byte copy of the four lines that other tests pin
        // (`audit_text_facts`, `next_flow_skill_scan`), so it is the one allowed second holder.
        let allowed = [contract.as_str(), "agents/audit.md"];
        if holders.iter().any(|h| !allowed.contains(h)) || !holders.contains(&contract.as_str()) {
            problems.push(format!(
                "the tier line `{prefix}` must stand only in {contract} (and the audit agent's copy), found in {holders:?}"
            ));
        }
    }

    assert!(
        problems.is_empty(),
        "AC-4 problems:\n{}",
        problems.join("\n")
    );
}

/// What AC-5 forbids in a text: one entry per kind found (internal numbers, tool paths, Python
/// steps and retired hosts).
fn forbidden_in(body: &str) -> Vec<String> {
    let mut found = Vec::new();
    for needle in [
        "Issue #",
        "tools/",
        "python3",
        ".py",
        "unittest",
        "Go generation",
    ] {
        if body.contains(needle) {
            found.push(format!("`{needle}`"));
        }
    }
    // Names of retired hosts, as whole words: `OMP` must not match the middle of `COMPLETE`.
    for word in ["Kiro", "Cursor", "Codex", "OMP"] {
        if has_word(body, word) {
            found.push(format!("`{word}`"));
        }
    }
    let bytes = body.as_bytes();
    for i in 0..bytes.len() {
        // A `#` followed by four digits: an Issue or pull request number.
        if bytes[i] == b'#'
            && bytes.len() >= i + 5
            && bytes[i + 1..i + 5].iter().all(u8::is_ascii_digit)
        {
            found.push(format!("a `#` and four digits at byte {i}"));
        }
        // A card id `C-<digit>` that starts a word, so `SPEC-2026…` and `ASYNC-1` do not match.
        if bytes[i] == b'C'
            && bytes.get(i + 1) == Some(&b'-')
            && bytes.get(i + 2).is_some_and(u8::is_ascii_digit)
            && (i == 0 || !bytes[i - 1].is_ascii_alphanumeric())
        {
            found.push(format!("a card id `C-<digits>` at byte {i}"));
        }
    }
    found
}

/// The `moai <command>` mentions in `body`: the first word and, when one follows, the second.
/// Only a lowercase word with dashes counts, so `<card>`, `--head` and `moai-impl` are skipped.
fn command_mentions(body: &str) -> Vec<(String, Option<String>)> {
    fn is_word(w: &str) -> bool {
        w.chars().next().is_some_and(|c| c.is_ascii_lowercase())
            && w.chars().all(|c| c.is_ascii_lowercase() || c == '-')
    }
    let mut out = Vec::new();
    for part in body.split("`moai ").skip(1) {
        let mut tokens = part.split_whitespace();
        let Some(first) = tokens.next() else { continue };
        let closed = first.ends_with('`');
        let first = first.trim_end_matches('`');
        if !is_word(first) {
            continue;
        }
        let second = if closed {
            None
        } else {
            tokens
                .next()
                .map(|s| s.trim_end_matches('`'))
                .filter(|s| is_word(s))
                .map(str::to_string)
        };
        out.push((first.to_string(), second));
    }
    out
}

/// Runs `moai <args> --help` with a temporary HOME; the exit code and the standard output.
fn help(home: &TempHome, args: &[&str]) -> (Option<i32>, String) {
    let mut all: Vec<&str> = args.to_vec();
    all.push("--help");
    let out = home.command(&all).output().expect("run moai --help");
    (
        out.status.code(),
        String::from_utf8_lossy(&out.stdout).into_owned(),
    )
}

/// AC-5 (REQ-6): no internal number, tool path, Python step or retired host, and every `moai`
/// command the texts name answers `--help` with exit 0.
///
/// A name is the first word after `` `moai ``; a second word counts only when the first word's
/// help says `Usage: moai <first> <COMMAND>` (a command group such as `card` or `check`).
#[test]
fn ac5_texts_name_no_internal_things_and_only_real_commands() {
    let mut problems: Vec<String> = Vec::new();
    let mut mentions: Vec<(String, Option<String>)> = Vec::new();
    let mut named_in: BTreeMap<String, String> = BTreeMap::new();
    for rel in EDITED.iter().chain(NEW.iter()) {
        let Some(body) = read_opt(rel) else {
            problems.push(format!("{rel} does not exist"));
            continue;
        };
        for hit in forbidden_in(&body) {
            problems.push(format!("{rel} contains {hit}"));
        }
        for (first, second) in command_mentions(&body) {
            let name = second
                .as_ref()
                .map_or(first.clone(), |s| format!("{first} {s}"));
            named_in.entry(name).or_insert_with(|| (*rel).to_string());
            mentions.push((first, second));
        }
    }

    // The commands the SPEC makes these texts name must be among the ones scanned.
    for required in [
        "card check",
        "check affected",
        "heavy-slot-acquire",
        "flake-classify",
        "card land",
        "pr-body-render",
    ] {
        if !named_in.contains_key(required) {
            problems.push(format!("no text names `moai {required}`"));
        }
    }

    let home = TempHome::new();
    let mut groups: BTreeMap<String, bool> = BTreeMap::new();
    for (first, second) in &mentions {
        if !groups.contains_key(first) {
            let (code, out) = help(&home, &[first]);
            if code != Some(0) {
                problems.push(format!("`moai {first} --help` exits {code:?}"));
            }
            groups.insert(
                first.clone(),
                out.contains(&format!("Usage: moai {first} <COMMAND>")),
            );
        }
        if let (Some(second), true) = (second, groups[first]) {
            let key = format!("{first} {second}");
            // `insert` answers `None` the first time a pair is seen, so each pair runs once.
            if groups.insert(key.clone(), false).is_none() {
                let (code, _) = help(&home, &[first, second]);
                if code != Some(0) {
                    problems.push(format!("`moai {key} --help` exits {code:?}"));
                }
            }
        }
    }

    assert!(
        problems.is_empty(),
        "AC-5 problems:\n{}",
        problems.join("\n")
    );
}

/// AC-6 (REQ-8, REQ-9): `sync.md` keeps the lane part and runs no check set; the lander and
/// release part lives in `flows/landing-release`.
#[test]
fn ac6_sync_keeps_the_lane_part_and_the_rest_moved() {
    let sync = text(SYNC);
    for banned in ["cargo test", "cargo clippy", "check affected"] {
        assert!(
            !sync.contains(banned),
            "sync.md still runs or names a check set: `{banned}`"
        );
    }
    assert_eq!(
        sync.matches("flows/landing-release").count(),
        1,
        "sync.md names `flows/landing-release` exactly once"
    );
    assert!(
        sync.lines().any(|l| l == "## Before the audit"),
        "sync.md lost `## Before the audit`"
    );
    assert!(
        sync.lines()
            .any(|l| l.starts_with("3. **Classify differences**")),
        "sync.md lost the procedure step `3. **Classify differences**`"
    );

    let landing = text(LANDING);
    for needle in [
        "effect-intend",
        "moai card land",
        "moai pr-body-render",
        "effects-reconcile",
    ] {
        assert!(
            landing.contains(needle),
            "landing-release.md lacks `{needle}`"
        );
    }
    assert!(
        landing.to_lowercase().contains("integration receipt"),
        "landing-release.md lacks the integration receipt"
    );
}

/// Checks the sentences REQ-10 keeps: the four of `run.md`'s `## Exceptions and Permissions` and the
/// one of `self-check.md`. Whitespace is collapsed, so a wrapped sentence still matches. The error
/// names the sentence that is missing.
fn gates_present(run: &str, self_check: &str) -> Result<(), String> {
    let exceptions = squash(&section_text(run, "Exceptions and Permissions"));
    for sentence in RUN_GATES {
        if !exceptions.contains(&squash(sentence)) {
            return Err(format!(
                "run.md `## Exceptions and Permissions` lacks the gate sentence: {sentence}"
            ));
        }
    }
    if !squash(self_check).contains(SELF_CHECK_GATE) {
        return Err(format!(
            "self-check.md lacks the gate sentence: {SELF_CHECK_GATE}"
        ));
    }
    Ok(())
}

/// A run text and a self-check text holding the gate sentences except `skip`, an index into the
/// five (the four of `RUN_GATES`, then `SELF_CHECK_GATE`).
fn gate_fixture(skip: Option<usize>) -> (String, String) {
    let all: Vec<&str> = RUN_GATES.iter().copied().chain([SELF_CHECK_GATE]).collect();
    let kept = |i: usize| (Some(i) != skip).then_some(all[i]);
    let run_body: Vec<&str> = (0..4).filter_map(kept).collect();
    let run = format!(
        "## Resume\n\nRead the state afresh.\n\n## Exceptions and Permissions\n\n{}\n\n## Completion Check\n\nDone.\n",
        run_body.join("\n")
    );
    let self_check = format!("# Internal: self-check\n\n{}\n", kept(4).unwrap_or(""));
    (run, self_check)
}

/// AC-7 (REQ-10): the five gate sentences are present in the shipped texts.
#[test]
fn ac7_gate_sentences_present() {
    if let Err(why) = gates_present(&text(RUN), &text(SELF_CHECK)) {
        panic!("{why}");
    }
}

/// AC-7 (REQ-10): the guard mutant: removing any one of the five sentences makes `gates_present`
/// fail and name it. Runs on an in-memory copy, so it does not depend on the shipped texts.
#[test]
fn ac7_removing_a_gate_sentence_fails() {
    let (run, self_check) = gate_fixture(None);
    assert_eq!(
        gates_present(&run, &self_check),
        Ok(()),
        "the control holds all five"
    );

    let all: Vec<&str> = RUN_GATES.iter().copied().chain([SELF_CHECK_GATE]).collect();
    for (i, sentence) in all.iter().enumerate() {
        let (run, self_check) = gate_fixture(Some(i));
        let err = gates_present(&run, &self_check)
            .expect_err(&format!("removing {sentence:?} must fail the check"));
        assert!(
            err.contains(sentence),
            "the error for a removed sentence names it: {err}"
        );
    }

    // A gate sentence outside `## Exceptions and Permissions` does not count for run.md.
    let (run, self_check) = gate_fixture(None);
    let moved = run.replace("## Exceptions and Permissions", "## Elsewhere");
    assert!(gates_present(&moved, &self_check).is_err());
}

/// AC-8 (REQ-11): self-check has no tool path or Python item and names guard mutants only at tier 2;
/// test-design names the project's test language and a mutation proof only at tier 2.
#[test]
fn ac8_self_check_and_test_design_state_the_decisions() {
    let self_check = text(SELF_CHECK);
    let lower = self_check.to_lowercase();
    for banned in ["tools/", "--full-python", "python"] {
        assert!(
            !lower.contains(banned),
            "self-check.md still says {banned:?}"
        );
    }
    let mutant_lines: Vec<&str> = self_check
        .lines()
        .filter(|l| l.contains("guard-mutant"))
        .collect();
    assert!(
        !mutant_lines.is_empty(),
        "self-check.md lost its `guard-mutant` line"
    );
    for line in mutant_lines {
        assert!(
            line.contains("tier 2"),
            "a `guard-mutant` line without `tier 2`: {line}"
        );
    }

    let design = text(TEST_DESIGN);
    let lower = design.to_lowercase();
    assert!(!has_word(&lower, "rust"), "test-design.md still names Rust");
    assert!(
        !has_word(&lower, "python"),
        "test-design.md still names Python"
    );
    assert!(
        design.contains("test language"),
        "test-design.md does not name the project's test language"
    );
    let proof_lines: Vec<&str> = design
        .lines()
        .filter(|l| l.to_lowercase().contains("mutation"))
        .collect();
    assert!(
        !proof_lines.is_empty(),
        "test-design.md lost the mutation proof"
    );
    for line in proof_lines {
        assert!(
            line.contains("tier 2"),
            "a mutation line without `tier 2`: {line}"
        );
    }
}

/// AC-9 (REQ-12): the manifest holds each edited and new file's real SHA-256, and no shipped text
/// points to a section name that no longer exists.
#[test]
fn ac9_manifest_hashes_match_and_no_pointer_dangles() {
    let mut problems: Vec<String> = Vec::new();

    let manifest: serde_json::Value =
        serde_json::from_str(&text("references/source-manifest.json"))
            .expect("source-manifest.json is JSON");
    for rel in EDITED.iter().chain(NEW.iter()) {
        let Ok(bytes) = std::fs::read(skill_path(rel)) else {
            problems.push(format!("{rel} does not exist"));
            continue;
        };
        let actual = format!("{:x}", Sha256::digest(&bytes));
        match manifest["files"][*rel]["sha256"].as_str() {
            Some(listed) if listed == actual => {}
            Some(listed) => problems.push(format!(
                "manifest sha256 of {rel} is {listed}, the file hashes to {actual}"
            )),
            None => problems.push(format!("{rel} has no sha256 entry in source-manifest.json")),
        }
    }

    let shipped = shipped_texts();
    for gone in [
        "Check-run rules",
        "Heavy checks are serialised",
        "Lane worktrees and build cache",
    ] {
        for (rel, body) in &shipped {
            if body.contains(gone) {
                problems.push(format!(
                    "{rel} still points to the removed section `{gone}`"
                ));
            }
        }
    }

    assert!(
        problems.is_empty(),
        "AC-9 problems:\n{}",
        problems.join("\n")
    );
}
