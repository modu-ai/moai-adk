//! SPEC-20261127-review-term-rename AC-3 (slice 1, the gate's readers): a record written under the
//! review name is read exactly as the same record under the earlier audit name, the two names
//! mix, and a line neither name reads is refused closed as before. The gate is driven through
//! `audit-gate` with a recorded input on stdin (no network), like `audit_gate_cli`.

mod common;

use serde_json::{json, Value};
use std::process::{Command, Output, Stdio};

const H: &str = "1111111111111111111111111111111111111111";
const OLD: &str = "2222222222222222222222222222222222222222";
const B: &str = "3333333333333333333333333333333333333333";

/// The two spellings of the verdict word and of the PR body tier marker.
const AUDIT: (&str, &str) = ("AUDIT", "audit-tier");
const REVIEW: (&str, &str) = ("REVIEW", "review-tier");

/// A documentation path of the fixture, built at run time so the docs-path scan of
/// `tools/ci_scope.py` does not read it as a repository document.
fn doc() -> String {
    format!("{}/n.md", "docs")
}

fn run(input: &Value) -> Output {
    let home = tempfile::TempDir::new().unwrap();
    let mut child = Command::new(crate::common::MOAI_BIN)
        .arg("audit-gate")
        .env("HOME", home.path())
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    common::feed_stdin(&mut child, input.to_string().as_bytes());
    child.wait_with_output().unwrap()
}

/// The fields of the decision that say what the gate did; the texts that echo a comment are left
/// out because they carry the word the comment was written with.
fn decide(input: &Value) -> Value {
    let out = run(input);
    let v: Value = serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("{}", String::from_utf8_lossy(&out.stderr)));
    assert_eq!(v["ok"], json!(true), "{v}");
    let r = &v["result"];
    json!({"state": r["state"], "rule": r["rule"], "tier": r["tier"],
           "cites_run_checked": r["cites_run_checked"]})
}

fn verdict(names: (&str, &str), id: u64, head: &str, word: &str, lane: &str) -> Value {
    json!({"id": id, "author_association": "MEMBER",
           "body": format!("{} #9 (x) head {head} base {B} — verdict: {word}\n\n감사: {lane}\n", names.0)})
}

fn input(
    names: (&str, &str),
    tier: u8,
    files: &[String],
    sessions: &[&str],
    comments: Vec<Value>,
) -> Value {
    let messages: Vec<String> = sessions
        .iter()
        .map(|s| format!("feat: x\n\nMoai-Spec: SPEC-1\nMoai-Session: {s}\n"))
        .collect();
    json!({"schema_version": 1, "pr": 9, "head": H,
           "body": format!("{}: {tier} reason\n\nbody", names.1),
           "files": files, "commit_messages": messages, "comments": comments})
}

/// The same scenario under the audit names and under the review names: the decision must be the
/// same, and it must be the one expected, so two wrong answers cannot pass as equal.
fn same_decision(build: impl Fn((&str, &str)) -> Value, state: &str, rule: Option<&str>) {
    let old = decide(&build(AUDIT));
    let new = decide(&build(REVIEW));
    assert_eq!(
        old, new,
        "the review names read differently from the audit names"
    );
    assert_eq!(old["state"], json!(state), "{old}");
    if let Some(rule) = rule {
        assert_eq!(old["rule"], json!(rule), "{old}");
    }
}

#[test]
fn a_review_verdict_is_read_like_an_audit_verdict_at_each_tier() {
    let docs = vec![doc()];
    let src = vec!["src/a.rs".to_string()];
    // Tier 0: no verdict passes; a posted FAIL stops it until the same reviewer passes.
    same_decision(|n| input(n, 0, &docs, &[], vec![]), "pass", None);
    same_decision(
        |n| {
            input(
                n,
                0,
                &docs,
                &[],
                vec![verdict(n, 5, OLD, "FAIL", "agent-7")],
            )
        },
        "fail",
        Some("open-fail"),
    );
    same_decision(
        |n| {
            let comments = vec![
                verdict(n, 5, OLD, "FAIL", "agent-7"),
                verdict(n, 6, H, "PASS", "agent-7"),
            ];
            input(n, 0, &docs, &[], comments)
        },
        "pass",
        None,
    );
    same_decision(
        |n| {
            let comments = vec![
                verdict(n, 5, OLD, "FAIL", "agent-7"),
                verdict(n, 6, H, "PASS", "agent-8"),
            ];
            input(n, 0, &docs, &[], comments)
        },
        "fail",
        None,
    );
    // Tier 1 and 2: a PASS on exactly this head by a lane that wrote no commit.
    for tier in [1, 2] {
        same_decision(
            |n| {
                input(
                    n,
                    tier,
                    &src,
                    &["agent-1"],
                    vec![verdict(n, 5, H, "PASS", "agent-7")],
                )
            },
            "pass",
            None,
        );
        same_decision(
            |n| {
                input(
                    n,
                    tier,
                    &src,
                    &["agent-7"],
                    vec![verdict(n, 5, H, "PASS", "agent-7")],
                )
            },
            "missing",
            None,
        );
        same_decision(
            |n| {
                input(
                    n,
                    tier,
                    &src,
                    &[],
                    vec![verdict(n, 5, H, "PASS", "agent-7")],
                )
            },
            "missing",
            Some("authors"),
        );
        same_decision(
            |n| {
                input(
                    n,
                    tier,
                    &src,
                    &["agent-1"],
                    vec![verdict(n, 5, OLD, "PASS", "agent-7")],
                )
            },
            "missing",
            Some("pass-on-head"),
        );
    }
}

#[test]
fn the_two_names_mix_and_the_newest_verdict_of_a_reviewer_decides() {
    let docs = vec![doc()];
    // An audit FAIL then a review PASS from one session clears it, the other way round too.
    let fail_then_pass = vec![
        verdict(AUDIT, 5, OLD, "FAIL", "agent-7"),
        verdict(REVIEW, 6, H, "PASS", "agent-7"),
    ];
    let r = decide(&input(AUDIT, 0, &docs, &[], fail_then_pass));
    assert_eq!(r["state"], json!("pass"), "{r}");
    let fail_then_pass = vec![
        verdict(REVIEW, 5, OLD, "FAIL", "agent-7"),
        verdict(AUDIT, 6, H, "PASS", "agent-7"),
    ];
    let r = decide(&input(REVIEW, 0, &docs, &[], fail_then_pass));
    assert_eq!(r["state"], json!("pass"), "{r}");
    // A PASS then a later FAIL under the other name stops it.
    let pass_then_fail = vec![
        verdict(AUDIT, 5, H, "PASS", "agent-7"),
        verdict(REVIEW, 6, H, "FAIL", "agent-7"),
    ];
    let r = decide(&input(AUDIT, 0, &docs, &[], pass_then_fail));
    assert_eq!(
        (r["state"].clone(), r["rule"].clone()),
        (json!("fail"), json!("open-fail")),
        "{r}"
    );
}

#[test]
fn the_tier_marker_of_the_body_is_read_under_both_names() {
    let docs = vec![doc()];
    // Declared tier 0 on a documentation-only change passes; declared tier 2 asks for a verdict.
    same_decision(|n| input(n, 0, &docs, &[], vec![]), "pass", None);
    same_decision(|n| input(n, 2, &docs, &[], vec![]), "missing", None);
    // A protected path is tier 2 whatever the body declares, under either marker.
    for names in [AUDIT, REVIEW] {
        let r = decide(&input(
            names,
            0,
            &[".github/workflows/x.yml".into()],
            &[],
            vec![],
        ));
        assert_eq!(
            (r["state"].clone(), r["tier"].clone()),
            (json!("missing"), json!(2)),
            "{r}"
        );
    }
}

#[test]
fn a_line_neither_name_reads_is_refused_closed_as_before() {
    let src = vec!["src/a.rs".to_string()];
    // A verdict line whose head is not 40 hex is no verdict, whichever word starts it: the gate
    // still has no PASS and says so, it never reads it as a pass.
    for names in [AUDIT, REVIEW] {
        let garbled = json!({"id": 5, "author_association": "MEMBER",
            "body": format!("{} #9 (x) head zz — verdict: PASS\nagent-7", names.0)});
        let r = decide(&input(names, 1, &src, &["agent-1"], vec![garbled]));
        assert_ne!(r["state"], json!("pass"), "{r}");
    }
    // A word that is neither name starts no verdict.
    let other = json!({"id": 5, "author_association": "MEMBER",
        "body": format!("INSPECT #9 (x) head {H} base {B} — verdict: PASS\n\n감사: agent-7\n")});
    let r = decide(&input(REVIEW, 1, &src, &["agent-1"], vec![other]));
    assert_eq!(r["state"], json!("missing"), "{r}");
    // A review verdict quoted further down a comment is never a verdict.
    let quoted = json!({"id": 5, "author_association": "MEMBER",
        "body": format!("note\nREVIEW #9 (x) head {H} base {B} — verdict: PASS\n\n감사: agent-7\n")});
    let r = decide(&input(REVIEW, 1, &src, &["agent-1"], vec![quoted]));
    assert_eq!(r["state"], json!("missing"), "{r}");
}

// ---- The merge gate (`gate-check`): reservations, withdrawals and non-member lines ----

/// The spellings of the reservation line and its two whole-lane withdrawals.
const CLAIM_AUDIT: (&str, &str, &str) = (
    "moai-audit-claim",
    "moai-audit-withdraw",
    "moai-audit-claim-withdraw",
);
const CLAIM_REVIEW: (&str, &str, &str) = (
    "moai-review-claim",
    "moai-review-withdraw",
    "moai-review-claim-withdraw",
);

/// Every pairing of a verdict word with a reservation word: the two names mix freely.
fn pairings() -> Vec<(
    (&'static str, &'static str),
    (&'static str, &'static str, &'static str),
)> {
    vec![
        (AUDIT, CLAIM_AUDIT),
        (REVIEW, CLAIM_REVIEW),
        (REVIEW, CLAIM_AUDIT),
        (AUDIT, CLAIM_REVIEW),
    ]
}

fn gate_check(snapshot: &Value) -> Value {
    let mut child = Command::new(crate::common::MOAI_BIN)
        .arg("gate-check")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    common::feed_stdin(&mut child, snapshot.to_string().as_bytes());
    let out = child.wait_with_output().unwrap();
    serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("{}", String::from_utf8_lossy(&out.stderr)))
}

fn claim(words: (&str, &str, &str), id: u64, lane: &str) -> Value {
    json!({"id": id, "author_association": "MEMBER",
           "body": format!("{}: lane={lane} head={H}\n", words.0)})
}

fn gate_verdict(names: (&str, &str), id: u64, lane: &str) -> Value {
    json!({"id": id, "author_association": "MEMBER",
           "body": format!("{} #9 (x) head {H} base {B} — verdict: PASS\n\n감사: {lane} (예약 100).\n", names.0)})
}

/// A complete snapshot that passes: one reservation, one PASS by the same lane, a local check.
fn passing(names: (&str, &str), words: (&str, &str, &str)) -> Value {
    let local = json!({"id": 900, "author_association": "MEMBER",
        "body": format!("local-check: head={H} result=pass cmd=tools/local_check.sh at=2026-11-14T12:00:00Z\n")});
    json!({
        "schema_version": 1, "pr": 9, "head": H,
        "comments": [claim(words, 100, "agent-10"), gate_verdict(names, 200, "agent-10"), local],
        "checks": [
            {"name": "ci", "conclusion": "success", "head_sha": H},
            {"name": "ci / linux-x86_64", "conclusion": "success", "head_sha": H},
            {"name": "windows-build", "conclusion": "skipped", "head_sha": H}
        ],
        "pr_files": ["src/a.rs"], "tip_files_since_verdict_base": ["src/b.rs"],
        "merge_clean": true, "body_check_ok": true, "verdict_base_known": true,
        "verdict_base_in_tip": true, "pr_authors": ["agent-26"],
        "tip_ci": {"state": "green", "tip": OLD, "decided_by": OLD, "detail": "green"},
        "commits_without_claim": [], "body_claim": true, "stacked_below": []
    })
}

fn stopped(v: &Value) -> (bool, String) {
    let r = &v["result"];
    (
        r["pass"].as_bool().unwrap(),
        r["rule"].as_str().unwrap_or("").to_string(),
    )
}

#[test]
fn a_complete_snapshot_passes_under_every_pairing_of_the_names() {
    for (names, words) in pairings() {
        let v = gate_check(&passing(names, words));
        assert_eq!(v["result"]["pass"], json!(true), "{names:?} {words:?}: {v}");
        assert_eq!(
            v["result"]["verdict_base"],
            json!(B),
            "{names:?} {words:?}: {v}"
        );
    }
}

#[test]
fn a_lower_reservation_under_either_name_takes_the_review() {
    for (names, words) in pairings() {
        for other in [CLAIM_AUDIT, CLAIM_REVIEW] {
            let mut s = passing(names, words);
            s["comments"]
                .as_array_mut()
                .unwrap()
                .insert(0, claim(other, 50, "agent-99"));
            let v = gate_check(&s);
            assert_eq!(
                stopped(&v),
                (false, "G2".into()),
                "{names:?} {words:?} {other:?}: {v}"
            );
        }
    }
}

#[test]
fn a_withdrawal_under_either_name_gives_the_reservation_up_and_only_its_own() {
    for (names, words) in pairings() {
        for other in [CLAIM_AUDIT, CLAIM_REVIEW] {
            let lower = |s: &mut Value| {
                s["comments"]
                    .as_array_mut()
                    .unwrap()
                    .insert(0, claim(other, 50, "agent-99"));
            };
            // `<claim> 철회: <lane> (<claim id>)` withdraws that claim of that lane.
            let mut s = passing(names, words);
            lower(&mut s);
            let line = format!("{} 철회: agent-99 (50)\n", other.0);
            s["comments"].as_array_mut().unwrap().insert(
                1,
                json!({"id": 60, "author_association": "MEMBER", "body": line}),
            );
            let v = gate_check(&s);
            assert_eq!(
                v["result"]["pass"],
                json!(true),
                "{names:?} {other:?} withdrawn: {v}"
            );
            // The withdrawal names another lane: the reservation keeps counting.
            let mut s = passing(names, words);
            lower(&mut s);
            let line = format!("{} 철회: agent-1 (50)\n", other.0);
            s["comments"].as_array_mut().unwrap().insert(
                1,
                json!({"id": 60, "author_association": "MEMBER", "body": line}),
            );
            let v = gate_check(&s);
            assert_eq!(
                stopped(&v),
                (false, "G2".into()),
                "{names:?} {other:?} wrong lane: {v}"
            );
            // The whole-lane withdrawals, under both of their spellings, posted after the claim.
            for withdraw in [other.1, other.2] {
                let mut s = passing(names, words);
                lower(&mut s);
                let line = format!("{withdraw}: lane=agent-99\n");
                s["comments"].as_array_mut().unwrap().insert(
                    1,
                    json!({"id": 60, "author_association": "MEMBER", "body": line.clone()}),
                );
                let v = gate_check(&s);
                assert_eq!(v["result"]["pass"], json!(true), "{names:?} {line}: {v}");
                // Posted before the claim it gives nothing up.
                let mut s = passing(names, words);
                lower(&mut s);
                s["comments"].as_array_mut().unwrap().insert(
                    0,
                    json!({"id": 40, "author_association": "MEMBER", "body": line.clone()}),
                );
                let v = gate_check(&s);
                assert_eq!(
                    stopped(&v),
                    (false, "G2".into()),
                    "{names:?} early {line}: {v}"
                );
            }
        }
    }
}

#[test]
fn a_non_member_verdict_or_reservation_under_the_review_names_is_ignored_and_named() {
    // The verdict from someone who is no member stops the gate, and the answer names its line.
    for (names, words) in pairings() {
        let mut s = passing(names, words);
        s["comments"][1]["author_association"] = json!("NONE");
        let v = gate_check(&s);
        assert_eq!(v["result"]["pass"], json!(false), "{names:?}: {v}");
        assert_eq!(
            v["result"]["ignored_control_lines"]["items"][0]["comment_id"],
            json!(200),
            "{names:?}: {v}"
        );
        assert_eq!(
            v["result"]["ignored_control_lines"]["items"][0]["kind"],
            json!("audit_verdict"),
            "{names:?}: {v}"
        );
    }
    // A forged lower reservation takes nothing, and the answer names it with the same kind as
    // the earlier name.
    for other in [CLAIM_AUDIT, CLAIM_REVIEW] {
        let mut s = passing(REVIEW, CLAIM_REVIEW);
        let mut forged = claim(other, 50, "agent-99");
        forged["author_association"] = json!("NONE");
        s["comments"].as_array_mut().unwrap().insert(0, forged);
        let v = gate_check(&s);
        assert_eq!(v["result"]["pass"], json!(true), "{other:?}: {v}");
        assert_eq!(
            v["result"]["ignored_control_lines"]["items"][0]["comment_id"],
            json!(50),
            "{other:?}: {v}"
        );
        assert_eq!(
            v["result"]["ignored_control_lines"]["items"][0]["kind"],
            json!("audit_claim"),
            "{other:?}: {v}"
        );
    }
}
