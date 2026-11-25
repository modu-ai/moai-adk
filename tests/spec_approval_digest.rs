//! SPEC-20261125-approval-digest-requirements-only, AC-1 to AC-5: a SPEC approval binds the
//! requirement sentences with their links, the criteria (id, result text, verification kind,
//! links) and the `forbidden:` entries of `## Scope`, and nothing else. `parse` carries a second
//! digest, `approval_digest`, next to the legacy `digest`. The field is read through the
//! serialized `SpecDoc`, so a build without it fails at run time instead of at compile time.
//! SPEC texts are built in memory; the command-line cases run `moai` in a temporary HOME with a
//! temporary project and ledger, never the real ones.

mod common;

use common::home::TempHome;
use moai_cli::ledger;
use moai_cli::spec_doc::{parse, Options, Severity, SpecDoc};
use serde_json::Value;
use std::fs;
use std::path::{Path, PathBuf};
use std::process::Output;

const ID: &str = "SPEC-20261117-login-fix";

/// Everything after the title heading: every section the format knows, one criterion of each
/// verification kind a completed SPEC accepts, and one entry of each `## Scope` kind.
const BODY: &str = r#"## Interview

- Q1: Which logins fail? → A: Logins with a valid password.

## Intent

Logins fail for valid passwords.

## Non-goals

- No change to the sign-up flow.

## Requirements

- REQ-1: When a user sends a valid password, the server shall accept the login.
- REQ-2 (REQ-1): IF the password check fails, THEN the server shall refuse the login.

## Acceptance

- AC-1 (REQ-1): A valid password logs in · verify: `cargo test --test login`
- AC-2 (REQ-2): A wrong password is refused · verify: `cargo test --test login_refused`
- AC-3 (REQ-1): The login page reads well · manual: a person reads the login page

## Decisions

- D-1: Keep the current hash.

## Scope

- write: `src/login.rs`
- read-only: `src/hash.rs`
- forbidden: `src/session.rs`

## Plan

1. Fix the comparison.

## Risks

- R-1: A slow hash.

## Result

- AC-1: pending
- AC-2: pending
- AC-3: pending
"#;

/// An edit of the text as `(label, from, to)`: the first `from` becomes `to`.
type Edit = (&'static str, &'static str, &'static str);

/// Edits of `BODY` that no part of the approval digest covers (AC-1).
#[rustfmt::skip]
const UNCOVERED: [Edit; 11] = [
    ("Intent", "Logins fail for valid passwords.", "Logins fail for every password."),
    ("Non-goals", "No change to the sign-up flow.", "No change to the sign-in flow."),
    ("Risks", "A slow hash.", "A slow hash and a lock-out."),
    ("Decisions", "Keep the current hash.", "Replace the current hash."),
    ("Plan", "Fix the comparison.", "Fix the comparison first."),
    ("Result", "- AC-1: pending", "- AC-1: met — a test run"),
    ("Interview", "with a valid password.", "with a valid password on a phone."),
    ("write entry changed", "- write: `src/login.rs`", "- write: `src/auth/login.rs`"),
    ("write entry added", "- write: `src/login.rs`", "- write: `src/login.rs`, `tests/login.rs`"),
    ("read-only entry changed", "- read-only: `src/hash.rs`", "- read-only: `src/crypto/hash.rs`"),
    ("verify command", "verify: `cargo test --test login`", "verify: `cargo test --test login_moved`"),
];

/// Edits of the approved text's front matter that the approval digest does not cover (AC-1).
#[rustfmt::skip]
const FRONT: [Edit; 3] = [
    ("title", "title: \"Fix login\"", "title: \"Fix the login\""),
    ("status", "status: approved\n", "status: completed\n"),
    ("revision", "revision: 1\n", "revision: 2\n"),
];

/// Edits of `BODY` that the approval digest covers (AC-2).
#[rustfmt::skip]
const COVERED: [Edit; 9] = [
    ("requirement sentence, one character", "accept the login.", "accept the logins."),
    ("criterion result text, one character", "A valid password logs in", "A valid password logs on"),
    ("verify changed to manual", "verify: `cargo test --test login`", "manual: `cargo test --test login`"),
    ("manual changed to unknown", "manual: a person reads", "unknown: a person reads"),
    ("manual text changed", "reads the login page", "reads the sign-in page"),
    ("link of a criterion", "- AC-3 (REQ-1)", "- AC-3 (REQ-2)"),
    ("forbidden entry changed", "forbidden: `src/session.rs`", "forbidden: `src/sessions.rs`"),
    ("forbidden entry added", "- forbidden: `src/session.rs`\n", "- forbidden: `src/session.rs`\n- forbidden: `src/token.rs`\n"),
    ("forbidden entry removed", "- forbidden: `src/session.rs`\n", ""),
];

fn front(status: &str, extra: &str) -> String {
    format!(
        "---\nid: {ID}\ntitle: \"Fix login\"\nstatus: {status}\nrevision: 1\ncreated: 2026-11-17\n\
         {extra}---\n\n# Fix login\n\n"
    )
}

fn read(text: &str) -> SpecDoc {
    parse(text, &Options::default())
}

fn draft_text(body: &str) -> String {
    format!("{}{body}", front("draft", ""))
}

/// The text of a SPEC approved with `digest`.
fn approved(body: &str, digest: &str) -> String {
    let block = format!(
        "approval:\n  by: user\n  at: 2026-11-17\n  basis: test selection\n  digest: {digest}\n"
    );
    format!("{}{body}", front("approved", &block))
}

fn error_codes(doc: &SpecDoc) -> Vec<&'static str> {
    let errors = doc
        .findings
        .iter()
        .filter(|f| f.severity == Severity::Error);
    errors.map(|f| f.code).collect()
}

fn is_digest(s: &str) -> bool {
    s.strip_prefix("sha256:")
        .is_some_and(|h| h.len() == 64 && h.bytes().all(|b| b.is_ascii_hexdigit()))
}

/// The approval digest of a parsed SPEC, read from its JSON form.
fn approval_digest(doc: &SpecDoc) -> String {
    let value = serde_json::to_value(doc).expect("a SpecDoc serializes");
    let digest = value["approval_digest"].as_str().unwrap_or_else(|| {
        panic!(
            "the SpecDoc JSON has no string `approval_digest`: {}",
            value["approval_digest"]
        )
    });
    assert!(is_digest(digest), "not a sha256 digest: {digest}");
    digest.to_string()
}

/// `text` with the first `from` replaced by `to`; both the match and a real change are required.
fn swap(text: &str, (_, from, to): &Edit) -> String {
    assert!(text.contains(from), "the fixture lacks {from:?}");
    assert_ne!(from, to);
    text.replacen(from, to, 1)
}

/// The approval digest of the unchanged SPEC and its legacy digest, both from the draft parse.
fn base_digests() -> (String, String) {
    let doc = read(&draft_text(BODY));
    assert!(error_codes(&doc).is_empty(), "{:?}", doc.findings);
    (
        approval_digest(&doc),
        doc.digest.clone().expect("a GEARS draft has a digest"),
    )
}

/// AC-1: each full text keeps the approval digest of `BODY` and passes the check.
fn assert_kept(cases: Vec<(&str, String)>) {
    let (digest, _) = base_digests();
    let mut broken = Vec::new();
    for (label, text) in cases {
        let doc = read(&text);
        let (errors, value) = (error_codes(&doc), serde_json::to_value(&doc).unwrap());
        if !errors.is_empty() || value["approval_digest"] != digest.as_str() {
            let got = &value["approval_digest"];
            broken.push(format!("{label}: errors {errors:?}, approval_digest {got}"));
        }
    }
    assert!(broken.is_empty(), "{}", broken.join("\n"));
}

#[test]
fn ac1_the_unchanged_approved_spec_passes_with_its_approval_digest() {
    let (digest, _) = base_digests();
    let doc = read(&approved(BODY, &digest));
    assert!(error_codes(&doc).is_empty(), "{:?}", doc.findings);
}

#[test]
fn ac1_sections_scope_write_entries_and_verify_command_keep_the_approval() {
    let (digest, _) = base_digests();
    assert_kept(
        UNCOVERED
            .iter()
            .map(|e| (e.0, approved(&swap(BODY, e), &digest)))
            .collect(),
    );
}

#[test]
fn ac1_front_matter_lines_keep_the_approval() {
    let (digest, _) = base_digests();
    let text = approved(BODY, &digest);
    assert_kept(FRONT.iter().map(|e| (e.0, swap(&text, e))).collect());
}

#[test]
fn ac2_each_covered_change_moves_the_approval_digest_and_the_approval_refuses() {
    let (digest, _) = base_digests();
    let mut broken = Vec::new();
    for edit in &COVERED {
        let (label, changed) = (edit.0, swap(BODY, edit));
        let moved = read(&draft_text(&changed));
        if !error_codes(&moved).is_empty() {
            broken.push(format!(
                "{label}: the changed draft has errors {:?}",
                moved.findings
            ));
            continue;
        }
        if approval_digest(&moved) == digest {
            broken.push(format!("{label}: the approval digest did not change"));
        }
        let errors = error_codes(&read(&approved(&changed, &digest)));
        if errors != vec!["approval_digest"] {
            broken.push(format!("{label}: the approved SPEC reports {errors:?}"));
        }
    }
    assert!(broken.is_empty(), "{}", broken.join("\n"));
}

/// A temporary HOME with a registered project ledger, for the command-line cases.
struct World {
    home: TempHome,
    project: PathBuf,
}

impl World {
    fn new() -> Self {
        let home = TempHome::new();
        let project = home.project("demo");
        ledger::register(home.path(), "demo").unwrap();
        World { home, project }
    }

    /// `moai <args> --project <project>` with the arguments split at spaces, and its JSON
    /// envelope (null when stdout is not JSON).
    fn run(&self, args: &str) -> (Output, Value) {
        let mut all: Vec<&str> = args.split(' ').collect();
        all.extend_from_slice(&["--project", self.project.to_str().unwrap()]);
        let out = self.home.command(&all).output().expect("run moai");
        let json = serde_json::from_slice(&out.stdout).unwrap_or(Value::Null);
        (out, json)
    }

    /// The `result` of a command that must exit 0.
    fn ok(&self, args: &str) -> Value {
        let (out, json) = self.run(args);
        let (stdout, stderr) = (
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr),
        );
        assert_eq!(out.status.code(), Some(0), "{args}: {stdout} {stderr}");
        json["result"].clone()
    }

    /// The `(kind, message)` of a command that must exit 1.
    fn refused(&self, args: &str) -> (String, String) {
        let (out, json) = self.run(args);
        assert_eq!(out.status.code(), Some(1), "{args}: {json}");
        let text = |k: &str| json["error"][k].as_str().unwrap_or_default().to_string();
        (text("kind"), text("message"))
    }

    fn write_spec(&self, text: &str) {
        let dir = self.project.join(".moai/specs").join(ID);
        fs::create_dir_all(&dir).unwrap();
        fs::write(dir.join("spec.md"), text).unwrap();
    }

    /// The digest `spec show` prints under `key` for the SPEC on disk.
    fn shown_digest(&self, key: &str) -> String {
        let shown = self.ok(&format!("spec show {ID}"));
        let digest = shown["spec"][key].as_str();
        digest
            .unwrap_or_else(|| panic!("`spec show` has no string {key}: {shown}"))
            .to_string()
    }
}

#[test]
fn ac2_card_spec_and_spec_check_refuse_a_changed_requirement_after_approval() {
    let w = World::new();
    let added = w.ok("card add --title t --kind feature --scope src/login.rs --ac fixture");
    let card = added["name"].as_str().unwrap().to_string();
    w.ok(&format!("card ready {card}"));
    w.write_spec(&draft_text(BODY));
    let digest = w.shown_digest("approval_digest");

    // The approved SPEC as written passes the check and links to the card.
    w.write_spec(&approved(BODY, &digest));
    w.ok(&format!("spec check {ID}"));
    w.ok(&format!("card spec {card} {ID}"));
    w.ok(&format!("card spec {card} --unlink"));

    // One changed character of a requirement sentence is refused by both commands.
    w.write_spec(&approved(&swap(BODY, &COVERED[0]), &digest));
    let (_, message) = w.refused(&format!("spec check {ID}"));
    assert!(message.contains("approval_digest"), "{message}");
    let (kind, message) = w.refused(&format!("card spec {card} {ID}"));
    assert_eq!(kind, "invalid_input", "{message}");
    assert!(
        message.contains(ID) && message.contains("error"),
        "{message}"
    );
    assert_eq!(w.ok(&format!("card show {card}"))["spec"], Value::Null);
}

#[test]
fn ac3_a_legacy_digest_approval_passes_while_the_document_is_unchanged() {
    let (approval, legacy) = base_digests();
    // The SPEC has a `write:` entry, so the two digests are different computations.
    assert_ne!(approval, legacy);
    let doc = read(&approved(BODY, &legacy));
    assert!(error_codes(&doc).is_empty(), "{:?}", doc.findings);
}

#[test]
fn ac3_a_legacy_approval_still_refuses_a_changed_write_entry_or_verify_command() {
    let (approval, legacy) = base_digests();
    assert_ne!(approval, legacy);
    for edit in [&UNCOVERED[7], &UNCOVERED[10]] {
        let doc = read(&approved(&swap(BODY, edit), &legacy));
        assert_eq!(error_codes(&doc), vec!["approval_digest"], "{}", edit.0);
    }
}

#[test]
fn ac3_a_recorded_digest_that_matches_neither_computation_is_refused() {
    let (approval, legacy) = base_digests();
    let other = format!("sha256:{}", "0".repeat(64));
    assert!(other != approval && other != legacy);
    let doc = read(&approved(BODY, &other));
    assert_eq!(error_codes(&doc), vec!["approval_digest"]);
}

#[test]
fn ac4_spec_show_prints_both_digests_and_they_differ_for_a_scope_with_write() {
    let w = World::new();
    w.write_spec(&draft_text(BODY));
    let (legacy, approval) = (w.shown_digest("digest"), w.shown_digest("approval_digest"));
    assert!(is_digest(&legacy), "{legacy}");
    assert!(is_digest(&approval), "{approval}");
    assert_ne!(legacy, approval, "the Scope has a write: entry");
}

const SDD: &str = "skills/moai/references/internal/sdd.md";
const PLAN: &str = "skills/moai/workflows/plan.md";

fn skill_text(relative: &str) -> String {
    let path = Path::new(env!("CARGO_MANIFEST_DIR")).join(relative);
    fs::read_to_string(&path).unwrap_or_else(|e| panic!("{}: {e}", path.display()))
}

#[test]
fn ac5_sdd_states_that_an_approval_stays_valid_while_the_approval_digest_matches() {
    let sdd = skill_text(SDD);
    let sentence = "An approval stays valid while the approval digest matches";
    assert!(
        sdd.contains(sentence),
        "sdd.md lacks the approval-digest sentence"
    );
    assert!(
        sdd.contains("approval_digest"),
        "sdd.md never names `approval_digest`"
    );
}

#[test]
fn ac5_plan_names_the_approval_digest() {
    let plan = skill_text(PLAN);
    assert!(
        plan.contains("approval_digest"),
        "plan.md never names `approval_digest`"
    );
}

#[test]
fn ac5_neither_text_says_that_every_change_of_an_approved_spec_needs_a_new_approval() {
    let old = [
        "Changing an approved SPEC raises `revision` and needs a new approval",
        "does not cover a changed SPEC",
    ];
    for (name, path) in [("sdd.md", SDD), ("plan.md", PLAN)] {
        let text = skill_text(path);
        for sentence in old {
            assert!(!text.contains(sentence), "{name} still says {sentence:?}");
        }
    }
}
