//! SPEC-20261117-ac-verify-required (card C-86, slice 2; user decision 2026-11-17 "권고 말고
//! 강제", rule "completion refused while AC verify runs are missing"):
//! - `moai spec check` refuses a `completed` SPEC with an acceptance criterion that has no way
//!   to verify it yet (`unknown:`), naming the criteria;
//! - `moai card done` refuses a card linked to a SPEC that fails `spec check` or still has an
//!   `unknown:` criterion, naming the criteria, and changes nothing.
//!
//! SPEC texts are built in memory; the card cases run against a temporary HOME and project and a
//! fake host connector, so nothing reads the real `~/.moai` or the network.
#![cfg(unix)]

mod common;

use common::home::TempHome;
use moai_cli::card::work;
use moai_cli::forge::connector::{FakeConnector, HostKind, PrPhase, PrState};
use moai_cli::ledger;
use moai_cli::spec_doc::{parse, Options, Severity, SpecDoc};
use serde_json::{json, Value};
use std::fs;
use std::path::PathBuf;
use std::process::Output;

const ID: &str = "SPEC-20261117-login-fix";
const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
const S_A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const MERGE: &str = "0123456789abcdef0123456789abcdef01234567";
const HEAD: &str = "89abcdef0123456789abcdef0123456789abcdef";

fn front(id: &str, status: &str, extra: &str) -> String {
    format!(
        "---\nid: {id}\ntitle: \"Fix login\"\nstatus: {status}\nrevision: 1\ncreated: 2026-11-17\n\
         {extra}---\n\n# Fix login\n\n"
    )
}

/// A minimal body whose `## Acceptance` holds `acs` (one `- AC-…` line each).
fn body(acs: &[&str]) -> String {
    format!(
        "## Intent\n\nLogins fail.\n\n## Requirements\n\n- REQ-1: When a user sends a valid \
         password, the server shall accept the login. · failure: none\n\n## Acceptance\n\n{}\n\n\
         ## Scope\n\n- write: `src/login.rs`\n\n## Plan\n\n1. Fix it.\n",
        acs.join("\n")
    )
}

fn approval_for(id: &str, body: &str) -> String {
    let draft = parse(
        &format!("{}{body}", front(id, "draft", "")),
        &Options::default(),
    );
    let digest = draft.digest.expect("a GEARS draft has a digest");
    format!(
        "approval:\n  by: user\n  at: 2026-11-17\n  basis: test selection\n  digest: {digest}\n"
    )
}

fn spec_text(id: &str, status: &str, acs: &[&str]) -> String {
    let b = body(acs);
    let approval = if status == "draft" {
        String::new()
    } else {
        approval_for(id, &b)
    };
    format!("{}{b}", front(id, status, &approval))
}

fn errors(d: &SpecDoc) -> Vec<(&'static str, String)> {
    d.findings
        .iter()
        .filter(|f| f.severity == Severity::Error)
        .map(|f| (f.code, f.message.clone()))
        .collect()
}

const VERIFIED: &str = "- AC-1 (REQ-1): it logs in · verify: `cargo test --test login`";
const MANUAL: &str = "- AC-2 (REQ-1): the page reads well · manual: a person reads the page";
const UNKNOWN_2: &str = "- AC-2 (REQ-1): it is fast · unknown: no benchmark exists yet";
const UNKNOWN_3: &str = "- AC-3 (REQ-1): it is safe · unknown: no probe exists yet";

/// AC-1 (N and N+1): a completed SPEC whose every criterion has `verify:` or `manual:` passes;
/// one `unknown:` criterion more is an error that names it, and two name both.
#[test]
fn a_completed_spec_refuses_a_criterion_without_a_way_to_verify_it() {
    let opts = Options::default();
    for acs in [vec![VERIFIED], vec![VERIFIED, MANUAL]] {
        let d = parse(&spec_text(ID, "completed", &acs), &opts);
        assert!(errors(&d).is_empty(), "{acs:?}: {:?}", d.findings);
    }
    let one = parse(&spec_text(ID, "completed", &[VERIFIED, UNKNOWN_2]), &opts);
    let found = errors(&one);
    assert_eq!(found.len(), 1, "{found:?}");
    assert_eq!(found[0].0, "ac_unverified");
    assert!(
        found[0].1.contains("AC-2") && !found[0].1.contains("AC-1"),
        "{found:?}"
    );
    let two = parse(
        &spec_text(ID, "completed", &[VERIFIED, UNKNOWN_2, UNKNOWN_3]),
        &opts,
    );
    let found = errors(&two);
    assert_eq!(found.len(), 1, "{found:?}");
    assert!(found[0].1.contains("AC-2, AC-3"), "{found:?}");
    // The closure check reads the same rule.
    let closure = Options {
        folder: None,
        closure: true,
    };
    let closed = parse(
        &format!(
            "{}## Result\n\n- AC-1: met — exit 0\n- AC-2: met — looked\n",
            spec_text(ID, "completed", &[VERIFIED, UNKNOWN_2])
        ),
        &closure,
    );
    assert!(
        errors(&closed).iter().any(|(c, _)| *c == "ac_unverified"),
        "{:?}",
        closed.findings
    );
}

/// AC-2: before completion an `unknown:` criterion stays a warning, as it was.
#[test]
fn before_completion_an_unknown_criterion_is_only_a_warning() {
    for status in ["draft", "approved"] {
        let d = parse(
            &spec_text(ID, status, &[VERIFIED, UNKNOWN_2]),
            &Options::default(),
        );
        assert!(errors(&d).is_empty(), "{status}: {:?}", d.findings);
        assert!(
            d.findings.iter().any(|f| f.code == "ac_unknown"),
            "{status}: {:?}",
            d.findings
        );
    }
}

struct World {
    home: TempHome,
    project: PathBuf,
}

fn envelope(out: &Output) -> Value {
    assert_eq!(
        out.status.code(),
        Some(0),
        "{} {}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

impl World {
    fn new() -> Self {
        let home = TempHome::new();
        let project = home.project("demo");
        ledger::register(home.path(), "demo").unwrap();
        World { home, project }
    }

    fn card(&self, session: Option<&str>, args: &[&str]) -> Value {
        let mut all = vec!["card"];
        all.extend_from_slice(args);
        all.extend_from_slice(&["--project", self.project.to_str().unwrap()]);
        let mut cmd = self.home.command(&all);
        if let Some(s) = session {
            cmd.env(SESSION_VAR, s);
        }
        envelope(&cmd.output().unwrap())
    }

    fn write_spec(&self, id: &str, text: &str) -> PathBuf {
        let dir = self.project.join(".moai/specs").join(id);
        fs::create_dir_all(&dir).unwrap();
        let path = dir.join("spec.md");
        fs::write(&path, text).unwrap();
        path
    }

    /// A card linked to `id`, taken, in review, its criterion met and a merged pull request
    /// linked: everything `card done` asks for except what the SPEC rule adds. Returns the
    /// generation.
    fn closable(&self, id: &str) -> i64 {
        let added = self.card(
            None,
            &[
                "add", "--title", "t", "--scope", "src/a/", "--ac", "it works",
            ],
        );
        assert_eq!(added["name"], json!("C-1"));
        self.card(None, &["ready", "C-1"]);
        self.card(None, &["spec", "C-1", id]);
        let taken = self.card(Some(S_A), &["take", "C-1"]);
        let g = taken["reservation"]["generation"].as_i64().unwrap();
        let gs = g.to_string();
        self.card(Some(S_A), &["review", "C-1", "--generation", &gs]);
        self.card(
            Some(S_A),
            &[
                "criterion",
                "C-1",
                "1",
                "--generation",
                &gs,
                "--met",
                "--evidence",
                "e",
            ],
        );
        let fake = FakeConnector::new(HostKind::Github).with_pr(PrState {
            number: 9,
            phase: PrPhase::Merged,
            head: HEAD.to_string(),
            merge_commit: Some(MERGE.to_string()),
        });
        let request = work::PrRequest {
            holder: S_A,
            generation: g,
            number: 9,
            from: None,
            repo: Some("octo/demo"),
        };
        work::pr(&self.project, self.home.path(), "C-1", &request, &fake).unwrap();
        g
    }

    fn done(&self, g: i64) -> Result<Value, String> {
        work::done(&self.project, self.home.path(), "C-1", S_A, g)
    }

    fn state(&self) -> String {
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        l.conn()
            .query_row("SELECT state FROM card WHERE card_id = 1", [], |r| r.get(0))
            .unwrap()
    }
}

/// AC-3 (N and N+1): `card done` on a card whose SPEC has an `unknown:` criterion is refused,
/// names the criterion and leaves the card in review; with every criterion verifiable it closes.
#[test]
fn card_done_waits_until_every_criterion_of_its_spec_can_be_verified() {
    let w = World::new();
    let path = w.write_spec(
        ID,
        &spec_text(ID, "draft", &[VERIFIED, UNKNOWN_2, UNKNOWN_3]),
    );
    let g = w.closable(ID);
    let refused = w.done(g).unwrap_err();
    assert!(refused.starts_with("card done refused:"), "{refused}");
    assert!(
        refused.contains(ID) && refused.contains("AC-2, AC-3"),
        "{refused}"
    );
    assert!(refused.contains("nothing was changed"), "{refused}");
    assert_eq!(w.state(), "review");
    fs::write(&path, spec_text(ID, "draft", &[VERIFIED, UNKNOWN_2])).unwrap();
    let refused = w.done(g).unwrap_err();
    assert!(
        refused.contains("AC-2") && !refused.contains("AC-3"),
        "{refused}"
    );
    fs::write(&path, spec_text(ID, "draft", &[VERIFIED, MANUAL])).unwrap();
    let closed = w.done(g).unwrap();
    assert_eq!(closed["state"], json!("done"));
    assert_eq!(w.state(), "done");
}

/// AC-4: a linked SPEC that is gone or fails `spec check` refuses `card done` too.
#[test]
fn card_done_refuses_a_linked_spec_that_is_gone_or_broken() {
    let w = World::new();
    let path = w.write_spec(ID, &spec_text(ID, "draft", &[VERIFIED]));
    let g = w.closable(ID);
    fs::write(&path, "---\nid: x\n---\n").unwrap();
    let refused = w.done(g).unwrap_err();
    assert!(refused.contains("spec check"), "{refused}");
    // Without spec.md the folder is read as a SPEC with an error; without the folder there is
    // no SPEC to read.
    fs::remove_file(&path).unwrap();
    let refused = w.done(g).unwrap_err();
    assert!(refused.contains("spec check"), "{refused}");
    fs::remove_dir(path.parent().unwrap()).unwrap();
    let refused = w.done(g).unwrap_err();
    assert!(refused.contains("cannot be read"), "{refused}");
    assert_eq!(w.state(), "review");
    w.write_spec(ID, &spec_text(ID, "draft", &[VERIFIED]));
    assert_eq!(w.done(g).unwrap()["state"], json!("done"));
}

/// #3174 audit N1 (SPEC-20261117-card-done-refusal-pin AC-1): "nothing was changed" holds byte
/// for byte. Every path, kind and file content under the temporary HOME (the project lies under
/// it) is the same before and after a refused `card done`; modification times are not compared
/// (decision D1). The control: once the SPEC is fixed, the same comparison differs.
#[test]
fn a_refused_card_done_changes_no_byte() {
    let w = World::new();
    let path = w.write_spec(ID, &spec_text(ID, "draft", &[VERIFIED, UNKNOWN_2]));
    let g = w.closable(ID);
    let bytes = || -> Vec<(String, &'static str, Vec<u8>)> {
        w.home
            .snapshot()
            .into_iter()
            .map(|e| (e.path, e.kind, e.bytes))
            .collect()
    };
    let before = bytes();
    assert!(
        before.iter().any(|(p, _, _)| p.starts_with("work/demo/")),
        "the project is not under the HOME"
    );
    let refused = w.done(g).unwrap_err();
    assert!(refused.contains("nothing was changed"), "{refused}");
    assert_eq!(bytes(), before, "a refused card done changed a byte");
    fs::write(&path, spec_text(ID, "draft", &[VERIFIED, MANUAL])).unwrap();
    let fixed = bytes();
    assert_eq!(w.done(g).unwrap()["state"], json!("done"));
    assert_ne!(
        bytes(),
        fixed,
        "closing the card must show in the comparison"
    );
}

/// A card that names no SPEC closes as before (its own refusals are `tests/card_work_cli.rs`).
#[test]
fn a_card_without_a_spec_closes_as_before() {
    let w = World::new();
    w.write_spec(ID, &spec_text(ID, "draft", &[VERIFIED]));
    let g = w.closable(ID);
    w.card(None, &["spec", "C-1", "--unlink"]);
    assert_eq!(w.done(g).unwrap()["state"], json!("done"));
}
