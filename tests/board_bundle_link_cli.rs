//! SPEC-20261120-board-bundle-link (BOARD 04, slice 2): the link view over stored observation
//! rows. The module is pure, so the cases build rows by hand; it is included by path because the
//! `src/board/` module is declared from `src/lib.rs` in another slice.
#![cfg(unix)]

#[allow(dead_code)]
#[path = "../src/board/link.rs"]
mod link;

use link::*;
use serde_json::json;
use std::collections::BTreeMap;

const SPEC: &str = "SPEC-20261120-demo-feature";

fn sha(c: char) -> String {
    // A 40-character lowercase hex stand-in: the digit repeated.
    let d = if c.is_ascii_hexdigit() { c } else { 'a' };
    std::iter::repeat_n(d, 40).collect()
}

fn row(seq: i64, kind: &str, scope: &str, content: serde_json::Value) -> StoredRow {
    StoredRow {
        seq,
        kind: kind.into(),
        scope: scope.into(),
        observed_at: 1_700_000_000 + seq,
        content: content.to_string(),
    }
}

fn target() -> String {
    sha('0')
}

fn run_row(seq: i64, source: &str, exit: i64) -> StoredRow {
    row(
        seq,
        "bundle-run@1",
        "bundle:m-flow/bundle-1",
        json!({"branch": "m-flow/bundle-1", "source_sha": source, "target_sha": target(),
            "commands": [{"check": "fmt", "exit": 0}, {"check": "clippy", "exit": exit}],
            "not_run": ["python"], "run_digest": "sha256:00", "specs": [SPEC]}),
    )
}

fn pr_row(seq: i64, number: u64, head: &str, review: serde_json::Value) -> StoredRow {
    row(
        seq,
        "pr-state@1",
        &format!("pr:{number}"),
        json!({"number": number, "state": "open", "head": head, "review": review}),
    )
}

fn pass(head: &str) -> serde_json::Value {
    json!({"verdict": "PASS", "head": head})
}

fn no_lane() -> LaneInput {
    LaneInput::default()
}

fn one_bundle(view: &LinkView) -> &Bundle {
    assert_eq!(view.state, State::Known, "{view:?}");
    assert_eq!(view.bundles.len(), 1, "{view:?}");
    &view.bundles[0]
}

/// AC-1 (REQ-1): the run record, both SHAs, the command results and the pull request's exact
/// head come out together for one bundle.
#[test]
fn the_link_view_shows_the_run_the_shas_and_the_pr_head_together() {
    let a = sha('a');
    let rows = vec![run_row(1, &a, 0), pr_row(2, 7, &a, serde_json::Value::Null)];
    let view = link(Some(&rows), &no_lane());
    let b = one_bundle(&view);
    assert_eq!(b.source_sha, a);
    assert_eq!(b.target_sha, target());
    assert_eq!(b.branch, "m-flow/bundle-1");
    assert_eq!(
        b.commands,
        vec![
            CommandResult {
                check: "fmt".into(),
                exit: 0
            },
            CommandResult {
                check: "clippy".into(),
                exit: 0
            }
        ]
    );
    assert_eq!(b.not_run, vec!["python".to_string()]);
    assert_eq!(b.specs, vec![SPEC.to_string()]);
    let pr =
        b.pr.as_ref()
            .expect("the pull request is attached by the SHA");
    assert_eq!(
        (pr.number, pr.head.as_str(), pr.state.as_str()),
        (7, a.as_str(), "open")
    );
    assert_eq!(b.run_effect, Effect::Standing);
    // A bundle with no pull request yet is not stale: it simply has none.
    let early = link(Some(&rows[..1]), &no_lane());
    let b = one_bundle(&early);
    assert!(b.pr.is_none());
    assert_eq!(b.run_effect, Effect::NoPullRequest);
}

/// AC-2 (REQ-2, REQ-8): a new head on the pull request makes the earlier pass history that is no
/// longer in effect; the pass for the new head is in effect; a pass whose head cannot be read is
/// never in effect.
#[test]
fn a_head_change_marks_the_earlier_pass_as_no_longer_in_effect() {
    let (a, b2) = (sha('a'), sha('b'));
    let before = vec![run_row(1, &a, 0), pr_row(2, 7, &a, pass(&a))];
    let view = link(Some(&before), &no_lane());
    let b = one_bundle(&view);
    assert_eq!(b.run_effect, Effect::Standing);
    let audit = &b.pr.as_ref().unwrap().audit;
    assert_eq!(audit.len(), 1);
    assert!(audit[0].in_effect && audit[0].verdict == Verdict::Pass);

    // The repair changes the head; the old verdict still names the old head.
    let mut after = before.clone();
    after.push(pr_row(3, 7, &b2, pass(&a)));
    let view = link(Some(&after), &no_lane());
    let b = one_bundle(&view);
    assert_eq!(b.run_effect, Effect::Lapsed, "{view:?}");
    let pr = b.pr.as_ref().unwrap();
    assert_eq!(pr.head, b2);
    assert_eq!(
        pr.audit.len(),
        1,
        "the earlier pass stays as history: {pr:?}"
    );
    assert_eq!(pr.audit[0].head.as_deref(), Some(a.as_str()));
    assert!(!pr.audit[0].in_effect, "{pr:?}");

    // A verdict for the new head is in effect and the old one stays out.
    after.push(pr_row(4, 7, &b2, pass(&b2)));
    let view = link(Some(&after), &no_lane());
    let pr = one_bundle(&view).pr.clone().unwrap();
    let in_effect: Vec<_> = pr.audit.iter().filter(|e| e.in_effect).collect();
    assert_eq!(in_effect.len(), 1, "{pr:?}");
    assert_eq!(in_effect[0].head.as_deref(), Some(b2.as_str()));
    assert!(pr
        .audit
        .iter()
        .any(|e| e.head.as_deref() == Some(a.as_str()) && !e.in_effect));

    // A pass whose head cannot be read is shown, and is not in effect.
    let bad = vec![
        run_row(1, &a, 0),
        pr_row(2, 7, &a, json!({"verdict": "PASS", "head": "not-a-sha"})),
    ];
    let view = link(Some(&bad), &no_lane());
    let pr = one_bundle(&view).pr.clone().unwrap();
    assert_eq!(pr.audit.len(), 1, "{pr:?}");
    assert_eq!(pr.audit[0].head, None);
    assert!(!pr.audit[0].in_effect);
}

/// AC-3 (REQ-3): the SPEC audit of the lane and the bundle pull request audit are two layers,
/// each with its own head; a pass in one never stands for the other.
#[test]
fn the_two_audit_layers_stay_apart() {
    let (a, l) = (sha('a'), sha('c'));
    let lane = LaneInput {
        heads: BTreeMap::from([(SPEC.to_string(), l.clone())]),
        audits: vec![LaneAudit {
            spec: SPEC.into(),
            head: l.clone(),
            verdict: Verdict::Pass,
        }],
    };
    // A lane pass, no pull request verdict: the bundle layer has none.
    let rows = vec![run_row(1, &a, 0), pr_row(2, 7, &a, serde_json::Value::Null)];
    let view = link(Some(&rows), &lane);
    let b = one_bundle(&view);
    assert_eq!(b.lane.len(), 1);
    assert_eq!(b.lane[0].spec, SPEC);
    assert_eq!(b.lane[0].head.as_deref(), Some(l.as_str()));
    assert!(b.lane[0].audit[0].in_effect);
    assert!(
        b.pr.as_ref().unwrap().audit.is_empty(),
        "no verdict is shown as none"
    );

    // A pull request pass, and the lane has moved on: the lane pass is out, the PR pass is in.
    let rows = vec![run_row(1, &a, 0), pr_row(2, 7, &a, pass(&a))];
    let moved = LaneInput {
        heads: BTreeMap::from([(SPEC.to_string(), sha('d'))]),
        audits: lane.audits.clone(),
    };
    let view = link(Some(&rows), &moved);
    let b = one_bundle(&view);
    assert!(!b.lane[0].audit[0].in_effect, "{b:?}");
    assert!(b.pr.as_ref().unwrap().audit[0].in_effect, "{b:?}");
    // The layers carry their own heads: they are different SHAs here.
    assert_ne!(
        b.lane[0].audit[0].head,
        b.pr.as_ref().unwrap().audit[0].head
    );
}

/// AC-5 (REQ-6): the view reads and writes nothing outside its arguments: the module names no
/// file, database, process or network path, and calling it leaves its inputs as they were and
/// gives the same answer twice. (The byte-for-byte check of the real files comes with the
/// observation store in a later slice.)
#[test]
fn the_link_view_writes_nothing() {
    let source = std::fs::read_to_string(concat!(env!("CARGO_MANIFEST_DIR"), "/src/board/link.rs"))
        .expect("read the module");
    let code: String = source
        .lines()
        .filter(|l| !l.trim_start().starts_with("//"))
        .collect::<Vec<_>>()
        .join("\n");
    for banned in [
        "std::fs",
        "std::process",
        "std::net",
        "std::env",
        "File::",
        "OpenOptions",
        "rusqlite",
        "Command::",
        "TcpStream",
    ] {
        assert!(!code.contains(banned), "the pure module names `{banned}`");
    }
    let a = sha('a');
    let rows = vec![run_row(1, &a, 0), pr_row(2, 7, &a, pass(&a))];
    let lane = no_lane();
    let (rows_before, lane_before) = (rows.clone(), lane.clone());
    let first = serde_json::to_string(&link(Some(&rows), &lane)).unwrap();
    let second = serde_json::to_string(&link(Some(&rows), &lane)).unwrap();
    assert_eq!(first, second);
    assert_eq!((rows, lane), (rows_before, lane_before));
}

/// AC-6 (REQ-7): a missing store, or a row that cannot be read, makes the link unknown and the
/// unreadable row is named.
#[test]
fn a_missing_store_or_an_unreadable_row_is_unknown() {
    let view = link(None, &no_lane());
    assert_eq!(view.state, State::Unknown);
    assert!(
        view.reason.as_deref().unwrap_or("").contains("store"),
        "{view:?}"
    );
    assert!(view.bundles.is_empty());

    let a = sha('a');
    let mut rows = vec![run_row(1, &a, 0), pr_row(2, 7, &a, pass(&a))];
    rows.push(StoredRow {
        seq: 5,
        kind: "pr-state@1".into(),
        scope: "pr:8".into(),
        observed_at: 0,
        content: "{not json".into(),
    });
    rows.push(row(
        6,
        "bundle-run@1",
        "bundle:x",
        json!({"branch": "x", "source_sha": "short", "target_sha": target(),
            "commands": [], "not_run": [], "run_digest": "sha256:00", "specs": []}),
    ));
    let view = link(Some(&rows), &no_lane());
    assert_eq!(view.state, State::Unknown, "{view:?}");
    assert_eq!(
        view.unreadable,
        vec![
            "row 5 (pr-state@1)".to_string(),
            "row 6 (bundle-run@1)".to_string()
        ]
    );
    // Rows of other kinds are not this view's business and are never named.
    let other = vec![row(1, "gap@1", "x", json!({"anything": 1}))];
    let view = link(Some(&other), &no_lane());
    assert_eq!(view.state, State::Known);
    assert!(view.unreadable.is_empty() && view.bundles.is_empty());
}
