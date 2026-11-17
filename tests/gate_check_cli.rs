//! Issue #2674 S1: `gate-check` decides the leader merge gate (rules G1–G8, Issue comment
//! 6849084101) from a snapshot on stdin. One passing snapshot; every test changes one thing and
//! sees exactly the rule that stops it, and the boundary that still passes.

mod common;

use serde_json::{json, Value};
use std::process::{Command, Stdio};

const H: &str = "1111111111111111111111111111111111111111";
const B: &str = "2222222222222222222222222222222222222222";
const OTHER: &str = "3333333333333333333333333333333333333333";

/// A documentation file of the fixture, built at run time so the docs-path scan of
/// `tools/ci_scope.py` does not read it as a repository document.
fn doc() -> String {
    format!("{}/n.md", "docs")
}

fn gate(snapshot: &Value) -> Value {
    let mut child = Command::new(crate::common::MOAI_BIN)
        .arg("gate-check")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    common::feed_stdin(&mut child, snapshot.to_string().as_bytes());
    let out = child.wait_with_output().unwrap();
    let v: Value = serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("{}", String::from_utf8_lossy(&out.stderr)));
    v
}

fn verdict(id: u64, lane: &str, head: &str, v: &str) -> Value {
    json!({"id": id, "author_association": "MEMBER", "body": format!("AUDIT #9 (x) head {head} base {B} — verdict: {v}\n\n감사: {lane} (예약 100).\n")})
}

fn claim(id: u64, lane: &str, head: &str) -> Value {
    json!({"id": id, "author_association": "MEMBER", "body": format!("moai-audit-claim: lane={lane} head={head}\n")})
}

/// The local check pass line for the head `H` (#2906 S5b, G5), posted after the verdict.
fn lc() -> Value {
    local_check(900, H, "pass")
}

fn passing() -> Value {
    json!({
        "schema_version": 1,
        "pr": 9,
        "head": H,
        "comments": [claim(100, "agent-10", H), verdict(200, "agent-10", H, "PASS"), lc()],
        "checks": [
            {"name": "ci", "conclusion": "success", "head_sha": H},
            {"name": "ci / linux-x86_64", "conclusion": "success", "head_sha": H},
            {"name": "windows-build", "conclusion": "skipped", "head_sha": H},
            {"name": "Vercel Preview Comments", "conclusion": "failure", "head_sha": H}
        ],
        "pr_files": ["src/a.rs", doc()],
        "tip_files_since_verdict_base": ["src/b.rs"],
        "merge_clean": true,
        "body_check_ok": true,
        "verdict_base_known": true,
        "verdict_base_in_tip": true,
        "pr_authors": ["agent-26"],
        "tip_ci": {"state": "green", "tip": OTHER, "decided_by": OTHER, "detail": "green"},
        "commits_without_claim": [],
        "body_claim": true,
        "stacked_below": []
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
fn a_complete_snapshot_passes_and_names_its_base() {
    let v = gate(&passing());
    assert_eq!(v["ok"], json!(true), "{v}");
    assert_eq!(v["result"]["pass"], json!(true), "{v}");
    assert_eq!(v["result"]["verdict_base"], json!(B));
    assert_eq!(v["result"]["overlap"], json!([]));
}

/// Issue #3052: the verdict and the local check count only when a repository member wrote
/// them. Each of the two, written by someone else, is named in the answer; the verdict then
/// stops the gate, while the local check — optional evidence since C-84, G5 reads `ci` — is
/// only dropped. The same comments from an owner, a member or a collaborator pass.
#[test]
fn a_verdict_or_local_check_from_a_non_member_does_not_pass_the_gate() {
    for index in [1, 2] {
        // C-84: comment 2 is the local check; without it the gate still passes on `ci`.
        let stops = index == 1;
        for association in [
            Some("NONE"),
            Some("CONTRIBUTOR"),
            Some("FIRST_TIME_CONTRIBUTOR"),
            Some(""),
            None,
        ] {
            let mut s = passing();
            let c = &mut s["comments"][index];
            match association {
                Some(a) => c["author_association"] = json!(a),
                None => {
                    c.as_object_mut().unwrap().remove("author_association");
                }
            }
            let id = c["id"].clone();
            let v = gate(&s);
            assert_eq!(
                v["result"]["pass"],
                json!(!stops),
                "comment {index} {association:?}: {v}"
            );
            assert_eq!(
                v["result"]["ignored_control_lines"]["items"][0]["comment_id"], id,
                "comment {index} {association:?}: {v}"
            );
        }
    }
    for association in ["OWNER", "MEMBER", "COLLABORATOR"] {
        let mut s = passing();
        for c in s["comments"].as_array_mut().unwrap() {
            c["author_association"] = json!(association);
        }
        let v = gate(&s);
        assert_eq!(v["result"]["pass"], json!(true), "{association}: {v}");
        assert!(v["result"].get("ignored_control_lines").is_none(), "{v}");
    }
}

/// Issue #3052: an audit reservation with a lower id than the auditor's takes the audit from
/// the auditor when a member wrote it (G2). Written by someone else it takes nothing, and the
/// answer names it.
#[test]
fn a_lower_audit_reservation_from_a_non_member_does_not_take_the_audit() {
    for association in [Some("NONE"), Some("CONTRIBUTOR"), Some(""), None] {
        let mut s = passing();
        let mut forged = claim(50, "agent-99", H);
        match association {
            Some(a) => forged["author_association"] = json!(a),
            None => {
                forged.as_object_mut().unwrap().remove("author_association");
            }
        }
        s["comments"].as_array_mut().unwrap().insert(0, forged);
        let v = gate(&s);
        assert_eq!(v["result"]["pass"], json!(true), "{association:?}: {v}");
        assert_eq!(
            v["result"]["ignored_control_lines"]["items"][0]["comment_id"],
            json!(50),
            "{association:?}: {v}"
        );
    }
    // The same reservation from a member does take it: the gate stops on G2.
    let mut s = passing();
    s["comments"]
        .as_array_mut()
        .unwrap()
        .insert(0, claim(50, "agent-99", H));
    let v = gate(&s);
    assert_eq!(stopped(&v), (false, "G2".into()), "{v}");
}

/// An old snapshot file, written before comments carried an association, is read as having no
/// member comment at all: it does not pass, and the answer says how to fix the file.
#[test]
fn a_snapshot_without_associations_fails_closed_and_says_why() {
    let mut s = passing();
    for c in s["comments"].as_array_mut().unwrap() {
        c.as_object_mut().unwrap().remove("author_association");
    }
    let v = gate(&s);
    assert_eq!(v["result"]["pass"], json!(false), "{v}");
    assert_eq!(
        v["result"]["ignored_control_lines"]["count"],
        json!(3),
        "{v}"
    );
    assert!(
        v["result"]["ignored_control_lines"]["hint"]
            .as_str()
            .is_some_and(|h| h.contains("author_association")),
        "{v}"
    );
}

#[test]
fn g1_the_last_verdict_must_be_a_pass_on_this_head() {
    let mut s = passing();
    s["comments"] = json!([claim(100, "agent-10", H), lc()]);
    assert_eq!(stopped(&gate(&s)), (false, "G1".into()));
    s["comments"] = json!([
        claim(100, "agent-10", H),
        verdict(200, "agent-10", OTHER, "PASS"),
        lc()
    ]);
    assert_eq!(stopped(&gate(&s)), (false, "G1".into()));
    s["comments"] = json!([
        claim(100, "agent-10", H),
        verdict(200, "agent-10", H, "FAIL"),
        lc()
    ]);
    assert_eq!(stopped(&gate(&s)), (false, "G1".into()));
    // A short base is not a verdict line.
    s["comments"] = json!([claim(100, "agent-10", H), {"id": 200, "author_association": "MEMBER", "body": format!("AUDIT #9 head {H} base 2222 — verdict: PASS\nagent-10")}, lc()]);
    assert_eq!(stopped(&gate(&s)), (false, "G1".into()));
    // The latest verdict counts: an older PASS on this head does not outlive a newer one elsewhere.
    s["comments"] = json!([
        claim(100, "agent-10", H),
        verdict(200, "agent-10", H, "PASS"),
        verdict(300, "agent-10", OTHER, "PASS"),
        lc()
    ]);
    assert_eq!(stopped(&gate(&s)), (false, "G1".into()));
}

#[test]
fn g2_the_auditor_holds_the_lowest_claim_or_the_lowest_on_this_head() {
    let mut s = passing();
    // A lane with no claim wrote the PASS.
    s["comments"] = json!([
        claim(100, "agent-10", H),
        verdict(200, "agent-27", H, "PASS"),
        lc()
    ]);
    assert_eq!(stopped(&gate(&s)), (false, "G2".into()));
    // The PR's lowest claim (on an older head) still owns the re-audit.
    s["comments"] = json!([
        claim(100, "agent-10", OTHER),
        claim(150, "agent-27", H),
        verdict(200, "agent-10", H, "PASS"),
        lc()
    ]);
    assert!(stopped(&gate(&s)).0);
    // The lowest claim on the current head is allowed too.
    s["comments"] = json!([
        claim(100, "agent-10", OTHER),
        claim(150, "agent-27", H),
        verdict(200, "agent-27", H, "PASS"),
        lc()
    ]);
    assert!(stopped(&gate(&s)).0);
    // Only the lowest on this head: a second claim on the head is not allowed.
    s["comments"] = json!([
        claim(100, "agent-10", OTHER),
        claim(150, "agent-27", H),
        claim(160, "agent-8", H),
        verdict(200, "agent-8", H, "PASS"),
        lc()
    ]);
    assert_eq!(stopped(&gate(&s)), (false, "G2".into()));
    // The auditor is the first lane after the AUDIT line, not one mentioned later.
    s["comments"] = json!([claim(100, "agent-10", H), {"id": 200, "author_association": "MEMBER", "body": format!("AUDIT #9 head {H} base {B} — verdict: PASS\n\n감사: agent-27, 저자 agent-10.\n")}, lc()]);
    assert_eq!(stopped(&gate(&s)), (false, "G2".into()));
    // A verdict naming no lane stops.
    s["comments"] = json!([claim(100, "agent-10", H), {"id": 200, "author_association": "MEMBER", "body": format!("AUDIT #9 head {H} base {B} — verdict: PASS\n\nok\n")}, lc()]);
    assert_eq!(stopped(&gate(&s)), (false, "G2".into()));
    // A lane name runs to its end: `agent-10x` is not `agent-10`.
    s["comments"] = json!([claim(100, "agent-10", H), {"id": 200, "author_association": "MEMBER", "body": format!("AUDIT #9 head {H} base {B} — verdict: PASS\n\n감사: agent-10x.\n")}, lc()]);
    assert_eq!(stopped(&gate(&s)), (false, "G2".into()));
    // `agent-100` is not `agent-10`.
    s["comments"] = json!([
        claim(100, "agent-10", H),
        verdict(200, "agent-100", H, "PASS"),
        lc()
    ]);
    assert_eq!(stopped(&gate(&s)), (false, "G2".into()));
    // No claim at all: nothing to check.
    s["comments"] = json!([verdict(200, "agent-27", H, "PASS"), lc()]);
    assert!(stopped(&gate(&s)).0);
}

#[test]
fn g3_a_fail_on_this_head_needs_the_same_auditor_to_reverse_it() {
    let mut s = passing();
    s["comments"] = json!([
        claim(100, "agent-10", H),
        verdict(150, "agent-10", H, "FAIL"),
        verdict(200, "agent-10", H, "PASS"),
        lc()
    ]);
    assert!(stopped(&gate(&s)).0);
    s["comments"] = json!([
        claim(100, "agent-10", H),
        claim(120, "agent-27", H),
        verdict(150, "agent-27", H, "FAIL"),
        verdict(200, "agent-10", H, "PASS"),
        lc()
    ]);
    assert_eq!(stopped(&gate(&s)), (false, "G3".into()));
    // A FAIL on an older head does not block the new head.
    s["comments"] = json!([
        claim(100, "agent-10", H),
        verdict(150, "agent-27", OTHER, "FAIL"),
        verdict(200, "agent-10", H, "PASS"),
        lc()
    ]);
    assert!(stopped(&gate(&s)).0);
}

#[test]
fn g4_the_self_audit_checker_must_pass() {
    let mut s = passing();
    s["body_check_ok"] = json!(false);
    assert_eq!(stopped(&gate(&s)), (false, "G4".into()));
}

/// C-84 (was #2906 S5b "only inform"): the check-run cases that only informed since #2906 S5b
/// stop G5 again on this head — no aggregate, a running job, a skipped or failed job, a failed
/// `windows-build` — and `ci_info` still names them. Only `windows-build` may be skipped;
/// unrelated checks are not read.
#[test]
fn g5_check_runs_without_the_aggregate_or_with_a_bad_job_stop_it() {
    let mut s = passing();
    for (checks, info, stops) in [
        (
            json!([run("ci / linux-x86_64", Some("success"), H)]),
            "ci=pending",
            true,
        ),
        (
            json!([
                run("ci", Some("success"), H),
                run("ci / linux-x86_64", None, H)
            ]),
            "ci=pending",
            true,
        ),
        (
            json!([
                run("ci", Some("success"), H),
                run("ci / linux-x86_64", Some("skipped"), H)
            ]),
            "ci=failure",
            true,
        ),
        (
            json!([
                run("ci", Some("success"), H),
                run("windows-build", Some("failure"), H)
            ]),
            "ci=failure",
            true,
        ),
        (
            json!([
                run("ci", Some("success"), H),
                run("windows-build", Some("skipped"), H),
                run("lint", Some("failure"), H)
            ]),
            "ci=success",
            false,
        ),
    ] {
        s["checks"] = checks;
        let v = gate(&s);
        let want = if stops {
            (false, "G5".to_string())
        } else {
            (true, String::new())
        };
        assert_eq!(stopped(&v), want, "{v}");
        assert_eq!(v["result"]["ci_info"], json!(info), "{v}");
    }
}

/// The `local-check:` line `tools/local_check.sh` prints last (#2906 S5a).
fn local_check_line(head: &str, result: &str) -> String {
    format!(
        "local-check: head={head} result={result} cmd=tools/local_check.sh at=2026-11-14T12:00:00Z"
    )
}

fn local_check(id: u64, head: &str, result: &str) -> Value {
    json!({"id": id, "author_association": "MEMBER", "body": format!("{}\n", local_check_line(head, result))})
}

/// `passing()` with its local-check comment replaced by `extra` (ids above the verdict's).
fn with_local_checks(extra: &[Value]) -> Value {
    let mut s = passing();
    let mut comments = vec![
        claim(100, "agent-10", H),
        verdict(200, "agent-10", H, "PASS"),
    ];
    comments.extend_from_slice(extra);
    s["comments"] = Value::Array(comments);
    s
}

fn g5_reason(s: &Value) -> String {
    let v = gate(s);
    assert_eq!(stopped(&v), (false, "G5".into()), "{v}");
    v["result"]["reason"].as_str().unwrap().to_string()
}

/// C-84 (was #2906 S5b AC21, "the last local-check line for this head decides"): no form of
/// the `local-check:` line — none, another head's, a fail, a malformed one, a pass, any order —
/// decides G5. With the head's `ci` success each case passes; without any `ci` each stops G5
/// as "no `ci` check run". The line's own parsing stays pinned by the unit tests in
/// `src/gate.rs` (`a_local_check_head_is_exactly_40_lowercase_hex` and the long-form ones).
#[test]
fn g5_no_local_check_line_decides_it() {
    let both = |a: &str, b: &str| json!({"id": 300, "author_association": "MEMBER", "body": format!("{}\n{}\n", local_check_line(H, a), local_check_line(H, b))});
    let at = "at=2026-11-14T12:00:00Z";
    let cmd = "cmd=tools/local_check.sh";
    let mut cases: Vec<Vec<Value>> = vec![
        vec![],
        vec![local_check(300, OTHER, "pass")],
        vec![local_check(300, H, "fail")],
        vec![local_check(300, H, "pass")],
        vec![local_check(300, H, "pass"), local_check(310, H, "fail")],
        vec![local_check(310, H, "fail"), local_check(300, H, "pass")],
        vec![both("pass", "fail")],
        vec![both("fail", "pass")],
    ];
    for line in [
        format!("local-check: head={} result=pass {cmd} {at}", &H[..39]),
        format!("local-check: head={H}1 result=pass {cmd} {at}"),
        format!(
            "local-check: head={} result=pass {cmd} {at}",
            "A".repeat(40)
        ),
        format!(
            "local-check: head={} result=pass {cmd} {at}",
            "g".repeat(40)
        ),
        format!("local-check: head={H} result=pass {cmd}"),
        format!("local-check: head={H} result=pass cmd=tools/other.sh {at}"),
        format!("local-check: head={H} result=pass dirty=1 {cmd} {at}"),
        format!("local-check: head={H} result=pass {cmd} at=2026-11-14 12:00:00Z"),
        format!("local-check: head={H} result=pass {cmd} at=2026-11-14T12:00:00"),
        format!("local-check: head={H} result=ok {cmd} {at}"),
        format!("local-check: head={H} result=pass {cmd} {at} extra"),
        format!("local-check:  head={H} result=pass {cmd} {at}"),
        format!(" {}", local_check_line(H, "pass")),
        format!("> {}", local_check_line(H, "pass")),
    ] {
        cases.push(vec![
            json!({"id": 300, "author_association": "MEMBER", "body": line}),
        ]);
    }
    for extra in cases {
        let s = with_local_checks(&extra);
        assert_eq!(stopped(&gate(&s)), (true, String::new()), "{extra:?}");
        let mut s = s;
        s["checks"] = json!([]);
        assert!(g5_reason(&s).contains("no `ci` check run"), "{extra:?}");
    }
}

const TREE: &str = "4444444444444444444444444444444444444444";
const SCRIPT: &str = "5555555555555555555555555555555555555555";
const RUN: &str = "0123456789ab";

/// The fields of the longer `local-check:` line (#3066 P1), in order, for the head `H`: a
/// passing run with a real base and a real tree.
fn evidence_fields(result: &str) -> Vec<(&'static str, String)> {
    [
        ("head", H),
        ("result", result),
        ("base", B),
        ("tree", TREE),
        ("script", SCRIPT),
        ("rustc", "1.91.0"),
        ("run", RUN),
        ("cmd", "tools/local_check.sh"),
        ("at", "2026-11-14T12:00:00Z"),
    ]
    .into_iter()
    .map(|(k, v)| (k, v.to_string()))
    .collect()
}

fn evidence_join(fields: &[(&str, String)]) -> String {
    let parts: Vec<String> = fields.iter().map(|(k, v)| format!("{k}={v}")).collect();
    format!("local-check: {}", parts.join(" "))
}

/// The new-form line for `head` with this `result` and the default evidence fields.
fn evidence_line(head: &str, result: &str) -> String {
    let mut f = evidence_fields(result);
    f[0].1 = head.to_string();
    evidence_join(&f)
}

/// The passing new-form line for `H` with one field's value replaced.
fn evidence_set(key: &str, value: &str) -> String {
    let mut f = evidence_fields("pass");
    f.iter_mut().find(|(k, _)| *k == key).unwrap().1 = value.to_string();
    evidence_join(&f)
}

/// The passing new-form line for `H` with one field removed.
fn evidence_drop(key: &str) -> String {
    let mut f = evidence_fields("pass");
    f.retain(|(k, _)| *k != key);
    evidence_join(&f)
}

fn evidence_comment(id: u64, line: &str) -> Value {
    json!({"id": id, "author_association": "MEMBER", "body": format!("{line}\n")})
}

fn raw_comment(id: u64, line: &str) -> Value {
    json!({"id": id, "author_association": "MEMBER", "body": line})
}

/// C-84 (was #3066 P1, where these lines decided G5): G5 ignores the `local-check:` line in
/// `s` — the snapshot passes with the head's `ci` success and, with no check run, stops G5 as
/// "no `ci` check run" whatever the line says. The line's parsing (both forms, every edge)
/// stays pinned by the unit tests in `src/gate.rs`.
fn g5_ignores(s: &Value) -> bool {
    assert_eq!(stopped(&gate(s)), (true, String::new()), "{}", gate(s));
    let mut without = s.clone();
    without["checks"] = json!([]);
    assert!(
        g5_reason(&without).contains("no `ci` check run"),
        "{without}"
    );
    true
}

fn assert_g5_ignores(extra: &[Value]) {
    assert!(g5_ignores(&with_local_checks(extra)));
}

/// C-84 (was #3066 P1, where the longer line decided G5 like the old one): G5 ignores the
/// longer line too, pass or fail, on any line of a comment.
#[test]
fn g5_ignores_the_new_local_check_line() {
    assert_g5_ignores(&[evidence_comment(300, &evidence_line(H, "pass"))]);
    // On any line of a comment, as with the old form.
    assert_g5_ignores(&[raw_comment(
        300,
        &format!("ran it\n\n{}\n", evidence_line(H, "pass")),
    )]);
    let s = with_local_checks(&[evidence_comment(300, &evidence_line(H, "fail"))]);
    assert!(g5_ignores(&s));
}

/// C-84 (was #3066 P1, the edges of the new line deciding G5): G5 ignores every edge of the
/// longer line, rejected or accepted; the edges themselves stay pinned in `src/gate.rs`.
#[test]
fn g5_ignores_every_edge_of_the_new_local_check_line() {
    let hex39 = &B[..39];
    let hex41 = format!("{B}2");
    let upper = "A".repeat(40);
    let nonhex = "g".repeat(40);
    let none_pass_fail = |line: String| {
        let s = with_local_checks(&[evidence_comment(300, &line)]);
        assert!(g5_ignores(&s), "{line}");
    };
    let mut rejected: Vec<String> = Vec::new();
    // Hex fields: 39 and 41 digits, uppercase, a non-hex letter.
    for key in ["head", "base", "tree", "script"] {
        for bad in [hex39, hex41.as_str(), upper.as_str(), nonhex.as_str(), ""] {
            rejected.push(evidence_set(key, bad));
        }
    }
    for bad in [
        &RUN[..11],
        "0123456789abc",
        "0123456789AB",
        "0123456789ag",
        "",
    ] {
        rejected.push(evidence_set("run", bad));
    }
    // `rustc`: 33 characters, empty, a space, `=`, `/`, and other characters.
    for bad in [
        "a".repeat(33),
        String::new(),
        "1.91 0".into(),
        "1=2".into(),
        "1/2".into(),
        "1.91.0:x".into(),
        "\u{e9}".into(),
    ] {
        rejected.push(evidence_set("rustc", &bad));
    }
    // `none` only on a failed run, and `base=none` needs `tree=none`.
    for (base, tree, result) in [
        ("none", "none", "pass"),
        (B, "none", "pass"),
        ("none", TREE, "fail"),
        ("none", TREE, "pass"),
        ("NONE", "none", "fail"),
        ("none", "NONE", "fail"),
        ("None", "None", "fail"),
    ] {
        let mut f = evidence_fields(result);
        f[2].1 = base.to_string();
        f[3].1 = tree.to_string();
        rejected.push(evidence_join(&f));
    }
    // Another result word.
    rejected.push(evidence_set("result", "ok"));
    rejected.push(evidence_set("result", ""));
    // Fields in another order.
    for (a, b) in [(2, 3), (4, 5), (5, 6), (0, 1), (6, 7), (7, 8)] {
        let mut f = evidence_fields("pass");
        f.swap(a, b);
        rejected.push(evidence_join(&f));
    }
    // A missing field, one case per field.
    for (key, _) in evidence_fields("pass") {
        rejected.push(evidence_drop(key));
    }
    // Another command or time.
    rejected.push(evidence_set("cmd", "tools/other.sh"));
    rejected.push(evidence_set("at", "2026-11-14T12:00:00"));
    rejected.push(evidence_set("at", "2026-11-14 12:00:00Z"));
    // A `dirty=1` field anywhere, extra trailing text, a second space, a leading space or
    // quote, no `local-check: ` prefix spacing.
    let good = evidence_line(H, "pass");
    rejected.push(good.replacen(" cmd=", " dirty=1 cmd=", 1));
    rejected.push(good.replacen(" base=", " dirty=1 base=", 1));
    rejected.push(good.replacen(" run=", " dirty=1 run=", 1));
    rejected.push(format!("{good} dirty=1"));
    rejected.push(format!("{good} extra"));
    rejected.push(format!("{good} "));
    rejected.push(good.replacen(" base=", "  base=", 1));
    rejected.push(good.replacen(" run=", "  run=", 1));
    rejected.push(good.replacen(" cmd=", "  cmd=", 1));
    rejected.push(good.replacen("local-check: ", "local-check:  ", 1));
    rejected.push(format!(" {good}"));
    rejected.push(format!("> {good}"));
    for line in rejected {
        none_pass_fail(line);
    }

    // Accepted edges. `rustc` of 32 characters from the whole allowed set passes.
    let rustc32 = "aZ09._+-".repeat(4);
    assert_eq!(rustc32.len(), 32);
    assert_g5_ignores(&[evidence_comment(300, &evidence_set("rustc", &rustc32))]);
    // One character is enough.
    assert_g5_ignores(&[evidence_comment(300, &evidence_set("rustc", "x"))]);
    // A failed run may have no tree (a conflict: real base) or no base and no tree; both are
    // read, and both stop G5 with the failure reason rather than "no line".
    for (base, tree) in [(B, "none"), ("none", "none")] {
        let mut f = evidence_fields("fail");
        f[2].1 = base.to_string();
        f[3].1 = tree.to_string();
        let s = with_local_checks(&[evidence_comment(300, &evidence_join(&f))]);
        assert!(g5_ignores(&s), "{base} {tree}");
    }
}

/// C-84 (was #3066 P1, "the last line for the head decides"): no order of old and new lines
/// decides G5; which line `gate-merge` records is pinned by `local_check_of`'s unit test.
#[test]
fn g5_ignores_old_and_new_lines_in_any_order() {
    let old = |r: &str| local_check_line(H, r);
    let new = |r: &str| evidence_line(H, r);
    for (low, high, passes) in [
        (old("pass"), new("fail"), false),
        (old("fail"), new("pass"), true),
        (new("pass"), old("fail"), false),
        (new("fail"), old("pass"), true),
    ] {
        // Listed in both orders: the comment id orders them, not the position in the list.
        for flip in [false, true] {
            let mut extra = vec![evidence_comment(300, &low), evidence_comment(310, &high)];
            if flip {
                extra.reverse();
            }
            if passes {
                assert_g5_ignores(&extra);
            } else {
                let s = with_local_checks(&extra);
                assert!(g5_ignores(&s), "{low} / {high}");
            }
        }
        // Inside one comment the later line wins.
        let both = [raw_comment(300, &format!("{low}\n{high}\n"))];
        if passes {
            assert_g5_ignores(&both);
        } else {
            let s = with_local_checks(&both);
            assert!(g5_ignores(&s), "{low} / {high}");
        }
    }
}

/// #3066 P1 (C-84: ignored by G5 like every other local-check line): a new-form line for
/// another head is ignored, pass or fail, like an old one.
#[test]
fn g5_a_new_local_check_line_for_another_head_is_ignored() {
    let s = with_local_checks(&[evidence_comment(300, &evidence_line(OTHER, "pass"))]);
    assert!(g5_ignores(&s));
    for r in ["pass", "fail"] {
        assert_g5_ignores(&[
            evidence_comment(300, &evidence_line(H, "pass")),
            evidence_comment(310, &evidence_line(OTHER, r)),
        ]);
    }
    // The head field is the head of the line; the base is not mistaken for it.
    let mut f = evidence_fields("pass");
    f[0].1 = OTHER.to_string();
    f[2].1 = H.to_string();
    let s = with_local_checks(&[evidence_comment(300, &evidence_join(&f))]);
    assert!(g5_ignores(&s));
}

/// C-84 (was #2906 S5b "check runs are information only"): the head's check runs decide G5
/// again, even with a local-check pass line, and every answer — pass or stop — carries
/// `ci_info`.
#[test]
fn g5_check_runs_decide_and_every_answer_carries_ci_info() {
    let mut s = with_local_checks(&[local_check(300, H, "pass")]);
    let v = gate(&s);
    assert_eq!(stopped(&v), (true, String::new()), "{v}");
    assert_eq!(v["result"]["ci_info"], json!("ci=success"), "{v}");
    for (checks, info) in [
        (
            json!([
                run("ci", Some("failure"), H),
                run("ci / linux-x86_64", Some("failure"), H)
            ]),
            "ci=failure",
        ),
        (json!([run("ci", None, H)]), "ci=pending"),
        (json!([run("lint", Some("failure"), H)]), "ci=none"),
        (json!([]), "ci=none"),
    ] {
        s["checks"] = checks;
        let v = gate(&s);
        assert_eq!(stopped(&v), (false, "G5".into()), "{v}");
        assert_eq!(v["result"]["ci_info"], json!(info), "{v}");
    }
}

#[test]
fn g6_overlap_is_allowed_only_for_kept_docs_and_a_matching_manifest() {
    let manifest = "skills/moai/references/source-manifest.json";
    let mut s = passing();
    s["merge_clean"] = json!(false);
    assert_eq!(stopped(&gate(&s)), (false, "G6".into()));
    let mut s = passing();
    s["pr_files"] = json!([]);
    assert_eq!(stopped(&gate(&s)), (false, "G6".into()));
    let mut s = passing();
    s["tip_files_since_verdict_base"] = json!(["src/a.rs"]);
    assert_eq!(stopped(&gate(&s)), (false, "G6".into()));
    s["tip_files_since_verdict_base"] = json!([doc()]);
    assert_eq!(stopped(&gate(&s)), (false, "G6".into()));
    s["docs_lines_kept"] = json!([doc()]);
    let v = gate(&s);
    assert_eq!(v["result"]["pass"], json!(true), "{v}");
    assert_eq!(v["result"]["overlap"], json!([doc()]));
    let mut s = passing();
    s["pr_files"] = json!([manifest]);
    s["tip_files_since_verdict_base"] = json!([manifest]);
    assert_eq!(stopped(&gate(&s)), (false, "G6".into()));
    s["merged_manifest_ok"] = json!(true);
    assert!(stopped(&gate(&s)).0);
}

#[test]
fn g7_a_review_file_waits_for_the_user() {
    let mut s = passing();
    s["pr_files"] = json!(["REVIEW.md"]);
    assert_eq!(stopped(&gate(&s)), (false, "G7".into()));
    s["review_confirmed_prs"] = json!([8, 10]);
    assert_eq!(stopped(&gate(&s)), (false, "G7".into()));
    s["review_confirmed_prs"] = json!([9]);
    assert!(stopped(&gate(&s)).0);
    let mut s = passing();
    s["pr_files"] = json!(["src/x/REVIEW.md"]);
    assert_eq!(stopped(&gate(&s)), (false, "G7".into()));
    // A file that only ends in the same letters is not a REVIEW.md.
    s["pr_files"] = json!(["src/x/MY_REVIEW.md"]);
    assert!(stopped(&gate(&s)).0);
}

#[test]
fn a_bad_snapshot_is_refused() {
    let mut s = passing();
    s["head"] = json!("abc");
    assert_eq!(stopped(&gate(&s)), (false, "input".into()));
    let mut s = passing();
    s["extra"] = json!(1);
    assert_eq!(gate(&s)["ok"], json!(false));
}

/// Audit 6850181707 B1: only a comment's first line is a verdict. A later comment quoting a
/// PASS line further down (a repair note, a merge record) neither becomes the last verdict nor
/// reverses a FAIL.
#[test]
fn a_quoted_verdict_line_is_not_a_verdict() {
    let quoted = json!({"id": 300, "author_association": "MEMBER", "body": format!("수리 push — agent-10에게\nAUDIT #9 (x) head {H} base {B} — verdict: PASS\nagent-10\n")});
    let mut s = passing();
    s["comments"] = json!([
        claim(100, "agent-10", H),
        verdict(200, "agent-10", H, "FAIL"),
        quoted,
        lc()
    ]);
    assert_eq!(stopped(&gate(&s)), (false, "G1".into()));
    // With the FAIL reversed by a real first-line PASS, a quote after it changes nothing.
    s["comments"] = json!([
        claim(100, "agent-10", H),
        verdict(150, "agent-10", H, "FAIL"),
        verdict(200, "agent-10", H, "PASS"),
        quoted,
        lc()
    ]);
    assert!(stopped(&gate(&s)).0);
    // A quoted FAIL does not block either.
    let quoted_fail = json!({"id": 300, "author_association": "MEMBER", "body": format!("note\nAUDIT #9 (x) head {H} base {B} — verdict: FAIL\nagent-27\n")});
    s["comments"] = json!([
        claim(100, "agent-10", H),
        verdict(200, "agent-10", H, "PASS"),
        quoted_fail,
        lc()
    ]);
    assert!(stopped(&gate(&s)).0);
}

/// Audit N1–N4: a title containing `head` still parses; `rc-ci` counts as the aggregate; a
/// withdrawn lowest claim (machine form) hands G2 to the next one, a free-text one does not;
/// an unknown verdict base stops G1.
#[test]
fn audit_notes_title_head_rc_ci_withdrawal_and_known_base() {
    let mut s = passing();
    s["comments"] = json!([claim(100, "agent-10", H), {"id": 200, "author_association": "MEMBER", "body": format!("AUDIT #9 (fix head recheck) head {H} base {B} — verdict: PASS\n\n감사: agent-10.\n")}, lc()]);
    assert!(stopped(&gate(&s)).0);

    let mut s = passing();
    s["checks"] = json!([
        run("rc-ci", Some("success"), H),
        run("ci / linux-x86_64", Some("success"), H)
    ]);
    assert_eq!(gate(&s)["result"]["ci_info"], json!("ci=success"));
    // C-84: a failed `rc-ci` on this head stops G5 (was #2906 S5b: reported, not a stop).
    s["checks"] = json!([
        run("ci", Some("success"), H),
        run("rc-ci", Some("failure"), H)
    ]);
    let v = gate(&s);
    assert_eq!(stopped(&v), (false, "G5".into()), "{v}");
    assert_eq!(v["result"]["ci_info"], json!("ci=failure"), "{v}");

    let mut s = passing();
    let withdraw = json!({"id": 120, "author_association": "MEMBER", "body": "moai-audit-claim 철회: agent-10 (100) — agent-27 takes it\n"});
    s["comments"] = json!([
        claim(100, "agent-10", H),
        withdraw,
        claim(130, "agent-27", H),
        verdict(200, "agent-27", H, "PASS"),
        lc()
    ]);
    assert!(stopped(&gate(&s)).0);
    let free_text = json!({"id": 120, "author_association": "MEMBER", "body": "agent-10: 감사 예약 100 철회\n"});
    s["comments"] = json!([
        claim(100, "agent-10", H),
        free_text,
        claim(130, "agent-27", OTHER),
        verdict(200, "agent-27", H, "PASS"),
        lc()
    ]);
    assert_eq!(stopped(&gate(&s)), (false, "G2".into()));

    let mut s = passing();
    s["verdict_base_known"] = json!(false);
    assert_eq!(stopped(&gate(&s)), (false, "G1".into()));
    s.as_object_mut().unwrap().remove("verdict_base_known");
    assert_eq!(stopped(&gate(&s)), (false, "G1".into()));
}

/// #3066 P3: a verdict base that exists but is not an ancestor of the base-branch tip says
/// nothing about the tip the PR merges onto, so G1 stops with its own reason. Known and in
/// the tip passes; an unknown base keeps the old reason; a snapshot without the
/// `verdict_base_in_tip` key reads as false (fail-closed).
#[test]
fn g1_stops_a_verdict_base_that_is_not_an_ancestor_of_the_tip() {
    let reason = |v: &Value| v["result"]["reason"].as_str().unwrap().to_string();

    // known + in the tip: passes.
    let v = gate(&passing());
    assert_eq!(stopped(&v), (true, "".into()), "{v}");

    // known, not in the tip: G1 with the new reason.
    let mut s = passing();
    s["verdict_base_in_tip"] = json!(false);
    let v = gate(&s);
    assert_eq!(stopped(&v), (false, "G1".into()), "{v}");
    assert!(
        reason(&v).contains("not an ancestor of the base-branch tip"),
        "{v}"
    );
    assert!(reason(&v).contains(B), "{v}");
    assert!(!reason(&v).contains("not a known commit"), "{v}");

    // The key removed: fail-closed, same stop.
    let mut s = passing();
    s.as_object_mut().unwrap().remove("verdict_base_in_tip");
    let v = gate(&s);
    assert_eq!(stopped(&v), (false, "G1".into()), "{v}");
    assert!(
        reason(&v).contains("not an ancestor of the base-branch tip"),
        "{v}"
    );

    // unknown base: the old reason stays, whatever `in_tip` says.
    for in_tip in [true, false] {
        let mut s = passing();
        s["verdict_base_known"] = json!(false);
        s["verdict_base_in_tip"] = json!(in_tip);
        let v = gate(&s);
        assert_eq!(stopped(&v), (false, "G1".into()), "{v}");
        assert!(reason(&v).contains("is not a known commit"), "{v}");
        assert!(!reason(&v).contains("ancestor"), "{v}");
    }
}

/// Audit 6850638994 B2: a withdrawal removes a claim only when its lane matches that claim's
/// lane; a mistyped id naming another lane's claim leaves it in place and G2 stops. N5: a
/// failed `rc-ci / …` job counts like a `ci / …` one (since #2906 S5b in `ci_info`, not G5).
#[test]
fn a_withdrawal_must_name_the_claims_own_lane_and_rc_ci_jobs_count() {
    let mut s = passing();
    let wrong = json!({"id": 3, "author_association": "MEMBER", "body": "moai-audit-claim 철회: agent-10 (1)\n"});
    s["comments"] = json!([
        claim(1, "agent-22", H),
        claim(2, "agent-10", H),
        wrong,
        verdict(200, "agent-10", H, "PASS"),
        lc()
    ]);
    assert_eq!(stopped(&gate(&s)), (false, "G2".into()));
    let right = json!({"id": 3, "author_association": "MEMBER", "body": "moai-audit-claim 철회: agent-22 (1)\n"});
    s["comments"] = json!([
        claim(1, "agent-22", H),
        claim(2, "agent-10", H),
        right,
        verdict(200, "agent-10", H, "PASS"),
        lc()
    ]);
    assert!(stopped(&gate(&s)).0);

    let mut s = passing();
    // C-84: a failed `rc-ci / …` job on this head stops G5 like a `ci / …` one (was: reported
    // in `ci_info` only).
    s["checks"] = json!([
        run("rc-ci", Some("success"), H),
        run("rc-ci / linux-x86_64", Some("failure"), H)
    ]);
    let v = gate(&s);
    assert_eq!(stopped(&v), (false, "G5".into()), "{v}");
    assert_eq!(v["result"]["ci_info"], json!("ci=failure"), "{v}");
}

/// #2777 G8: the tip's push CI. Green and flaky pass; red stops every PR but one whose body
/// carries `moai-tip-repair:`; unknown (or no `tip_ci` at all) stops every PR, a repair too.
#[test]
fn a_red_tip_stops_all_but_its_repair_and_unknown_stops_everything() {
    let tip = |state: &str| json!({"state": state, "tip": OTHER, "decided_by": OTHER, "detail": "ci failure run 7"});
    for (state, repair, want) in [
        ("green", false, (true, String::new())),
        ("flaky", false, (true, String::new())),
        ("red", false, (false, "G8".to_string())),
        ("red", true, (true, String::new())),
        ("unknown", false, (false, "G8".to_string())),
        ("unknown", true, (false, "G8".to_string())),
        ("pending", false, (false, "G8".to_string())),
    ] {
        let mut s = passing();
        s["tip_ci"] = tip(state);
        s["tip_repair"] = json!(repair);
        assert_eq!(stopped(&gate(&s)), want, "{state} repair={repair}");
    }
    let mut s = passing();
    s.as_object_mut().unwrap().remove("tip_ci");
    let v = gate(&s);
    assert_eq!(stopped(&v), (false, "G8".into()), "{v}");
    assert!(
        v["result"]["reason"].as_str().unwrap().contains("not read"),
        "{v}"
    );
}

/// G8 comes first: on a red tip the stop names G8 even when a later rule would also stop, so
/// the merge loop acts on nothing else (no update-branch) while the tip is red. The reason
/// names the earlier commit whose completed CI decided.
#[test]
fn a_red_tip_is_the_first_stop_and_names_the_deciding_commit() {
    let mut s = passing();
    s["merge_clean"] = json!(false);
    s["tip_ci"] =
        json!({"state": "red", "tip": OTHER, "decided_by": B, "detail": "ci failure run 7"});
    let v = gate(&s);
    assert_eq!(stopped(&v), (false, "G8".into()), "{v}");
    let reason = v["result"]["reason"].as_str().unwrap();
    assert!(
        reason.contains("last completed CI on 2222222222"),
        "{reason}"
    );
    assert!(reason.contains("moai-tip-repair:"), "{reason}");
    s["tip_ci"]["state"] = json!("green");
    assert_eq!(stopped(&gate(&s)), (false, "G6".into()));
}

/// #2674 R1 (leader scope 6850215811): an author never audits its own PR — the verdict's
/// auditor must not be a session whose claim the PR's commits name; unknown authors stop.
#[test]
fn g9_the_auditor_is_not_an_author() {
    let mut s = passing();
    s["pr_authors"] = json!(["agent-26", "agent-10"]);
    assert_eq!(stopped(&gate(&s)), (false, "G9".into()));
    // Lane names compare without case.
    s["pr_authors"] = json!(["Agent-10"]);
    assert_eq!(stopped(&gate(&s)), (false, "G9".into()));
    // A lane whose name only starts like the auditor's is someone else.
    s["pr_authors"] = json!(["agent-1", "agent-100"]);
    assert_eq!(stopped(&gate(&s)), (true, String::new()));
    // No trailer in the PR: nobody to compare, passes.
    s["pr_authors"] = json!([]);
    assert_eq!(stopped(&gate(&s)), (true, String::new()));
    // Unknown authors stop rather than pass.
    s.as_object_mut().unwrap().remove("pr_authors");
    let v = gate(&s);
    assert_eq!(stopped(&v), (false, "G9".into()), "{v}");
    assert!(
        v["result"]["reason"].as_str().unwrap().contains("unknown"),
        "{v}"
    );
}

/// Audit 6863303949 B2: G9 reads the auditor of the last verdict, not of an earlier one.
#[test]
fn g9_reads_the_last_verdict() {
    let mut s = passing();
    // agent-26 holds the lowest claim, so G2 lets its verdict through; it is also an author.
    s["comments"] = json!([
        claim(100, "agent-26", H),
        verdict(200, "agent-10", H, "FAIL"),
        verdict(300, "agent-26", H, "PASS"),
        lc()
    ]);
    s["pr_authors"] = json!(["agent-26"]);
    assert_eq!(stopped(&gate(&s)), (false, "G9".into()));
}

/// Audit 6864692054 B1: G10 carries over `flow-gate-verdict`'s `commit_claim_lines` — every
/// commit of the PR that is not a merge has a `Moai-Claim` trailer. 0 missing passes, 1 stops,
/// unknown stops.
#[test]
fn g10_every_commit_of_the_pr_carries_a_claim_trailer() {
    let mut s = passing();
    s["commits_without_claim"] = json!([OTHER]);
    let v = gate(&s);
    assert_eq!(stopped(&v), (false, "G10".into()), "{v}");
    let reason = v["result"]["reason"].as_str().unwrap();
    assert!(
        reason.contains("1 commit(s)") && reason.contains(&OTHER[..10]),
        "{reason}"
    );
    s["commits_without_claim"] = json!([OTHER, H]);
    let v = gate(&s);
    assert!(
        v["result"]["reason"]
            .as_str()
            .unwrap()
            .contains("2 commit(s)"),
        "{v}"
    );
    s["commits_without_claim"] = json!([]);
    assert_eq!(stopped(&gate(&s)), (true, String::new()));
    s.as_object_mut().unwrap().remove("commits_without_claim");
    let v = gate(&s);
    assert_eq!(stopped(&v), (false, "G10".into()), "{v}");
    assert!(
        v["result"]["reason"]
            .as_str()
            .unwrap()
            .contains("not checked"),
        "{v}"
    );
}

/// Audit 6864692054 B1: G11 carries over `flow-gate-verdict`'s `pr_body_claim` — the PR body
/// names its claim. Absent reads as no.
#[test]
fn g11_the_pr_body_carries_a_claim_line() {
    let mut s = passing();
    s["body_claim"] = json!(false);
    assert_eq!(stopped(&gate(&s)), (false, "G11".into()));
    s.as_object_mut().unwrap().remove("body_claim");
    assert_eq!(stopped(&gate(&s)), (false, "G11".into()));
    s["body_claim"] = json!(true);
    assert_eq!(stopped(&gate(&s)), (true, String::new()));
}

/// #2833: the last PASS is on an earlier head (OTHER) and the current head (H) only merged the
/// base branch into it — the collector's finding plus equal patch-ids per file.
fn on_delta() -> Value {
    let mut s = passing();
    s["comments"] = json!([
        claim(100, "agent-10", OTHER),
        verdict(200, "agent-10", OTHER, "PASS"),
        lc()
    ]);
    s["auto_delta"] = json!({
        "from": OTHER, "to": H, "ok": true, "detail": "merges only the base branch",
        "files": [
            {"path": "src/a.rs", "patch_id_from": "aa11", "patch_id_to": "aa11"},
            {"path": doc(), "patch_id_from": "bb22", "patch_id_to": "bb22"}
        ]
    });
    s
}

/// #2833 D1: a mechanical delta carries the PASS on its verdict head to the current head, and
/// the verdict's base is the one reported.
#[test]
fn a_mechanical_delta_carries_the_pass_to_the_current_head() {
    let v = gate(&on_delta());
    assert_eq!(stopped(&v), (true, String::new()), "{v}");
    assert_eq!(v["result"]["verdict_base"], json!(B));
    assert_eq!(v["result"]["auto_delta"]["from"], json!(OTHER), "{v}");
    assert_eq!(v["result"]["auto_delta"]["to"], json!(H), "{v}");
    // Without a delta the same comments stop at G1, as before.
    let mut s = on_delta();
    s.as_object_mut().unwrap().remove("auto_delta");
    let v = gate(&s);
    assert_eq!(stopped(&v), (false, "G1".into()));
    assert!(v["result"].get("auto_delta").is_none());
}

/// #2833 AC-1/AC-2 at the gate: every way the delta does not hold stops at G1 with its reason;
/// the gate checks the patch-ids itself, so a snapshot claiming `ok` with unequal ids stops.
#[test]
fn a_delta_that_does_not_hold_stops_at_g1() {
    let reason = |s: &Value| {
        let v = gate(s);
        assert_eq!(stopped(&v), (false, "G1".into()), "{v}");
        v["result"]["reason"].as_str().unwrap().to_string()
    };
    let mut s = on_delta();
    s["auto_delta"]["ok"] = json!(false);
    s["auto_delta"]["detail"] = json!("1 commit(s) after the verdict head are not merges");
    assert!(reason(&s).contains("not a mechanical delta: 1 commit(s)"));

    let mut s = on_delta();
    s["auto_delta"]["files"][1]["patch_id_to"] = json!("bb23");
    assert!(reason(&s).contains("patch-ids or heads do not match"));

    let mut s = on_delta();
    s["auto_delta"]["files"][0]["patch_id_from"] = json!("");
    s["auto_delta"]["files"][0]["patch_id_to"] = json!("");
    reason(&s);

    let mut s = on_delta();
    s["auto_delta"]["files"] = json!([]);
    reason(&s);

    // The delta must start at the last verdict's head and end at the current head.
    let mut s = on_delta();
    s["auto_delta"]["from"] = json!(B);
    reason(&s);
    let mut s = on_delta();
    s["auto_delta"]["to"] = json!(OTHER);
    reason(&s);
    // A third head: neither the verdict head nor the current one.
    let mut s = on_delta();
    s["auto_delta"]["to"] = json!(B);
    reason(&s);

    // A FAIL is never carried: the delta counts only a PASS.
    let mut s = on_delta();
    s["comments"] = json!([
        claim(100, "agent-10", OTHER),
        verdict(200, "agent-10", OTHER, "FAIL"),
        lc()
    ]);
    assert_eq!(stopped(&gate(&s)), (false, "G1".into()));
}

/// #2833 scope 4: the current head's own check still decides (G5) — since C-84 its GitHub
/// `ci` (was #2906 S5b: its `local-check:` line): the verdict head's `ci` success does not
/// carry over to the merged head, and a failed `ci` on the current head stops it.
#[test]
fn a_mechanical_delta_still_needs_the_current_heads_ci() {
    let mut s = on_delta();
    s["checks"] = json!([run("ci", Some("success"), OTHER)]);
    assert_eq!(stopped(&gate(&s)), (false, "G5".into()));
    let mut s = on_delta();
    s["checks"] = json!([
        run("ci", Some("failure"), H),
        run("ci / linux-x86_64", Some("failure"), H)
    ]);
    assert_eq!(stopped(&gate(&s)), (false, "G5".into()));
    // The current head's success passes, a local-check line or not.
    let mut s = on_delta();
    s["checks"] = json!([run("ci", Some("success"), H)]);
    assert_eq!(stopped(&gate(&s)), (true, String::new()));
}

/// #2833 D2: the audit rules read the verdict head. An unresolved FAIL there still blocks
/// (G3), and the auditor's claim on that head counts (G2).
#[test]
fn the_audit_rules_apply_to_the_verdict_head_of_a_delta() {
    let mut s = on_delta();
    s["comments"] = json!([
        claim(100, "agent-10", OTHER),
        verdict(150, "agent-27", OTHER, "FAIL"),
        verdict(200, "agent-10", OTHER, "PASS"),
        lc()
    ]);
    assert_eq!(stopped(&gate(&s)), (false, "G3".into()));
    // The same auditor reversing itself on the verdict head passes, as on a plain head.
    s["comments"] = json!([
        claim(100, "agent-10", OTHER),
        verdict(150, "agent-10", OTHER, "FAIL"),
        verdict(200, "agent-10", OTHER, "PASS"),
        lc()
    ]);
    assert_eq!(stopped(&gate(&s)), (true, String::new()));
    // G2: the lowest claim is another lane's on an older head; the claim on the verdict head
    // is the auditor's, so it is allowed.
    s["comments"] = json!([
        claim(90, "agent-27", B),
        claim(100, "agent-10", OTHER),
        verdict(200, "agent-10", OTHER, "PASS"),
        lc()
    ]);
    assert_eq!(stopped(&gate(&s)), (true, String::new()));
}

fn note(id: u64, body: &str) -> Value {
    json!({"id": id, "author_association": "MEMBER", "body": body})
}

/// #2860 rule 1 (shell owner_ok.py): a leader handoff whose first line names an allowed claim
/// moves it to the named lane; the handoff must come after the claim, name an allowed claim,
/// say `by=leader`, and sit on the first line. The latest handoff wins.
#[test]
fn g2_a_leader_handoff_moves_an_allowed_claim() {
    let mut s = passing();
    let handoff = "moai-claim-handoff: comment=100 to=agent-27 by=leader\n\nleader";
    s["comments"] = json!([
        claim(100, "agent-10", H),
        note(150, handoff),
        verdict(200, "agent-27", H, "PASS"),
        lc()
    ]);
    assert_eq!(stopped(&gate(&s)), (true, "".into()), "{}", gate(&s));
    // Not on the first line, not by the leader, before the claim, or naming another claim:
    // the claim stays with agent-10 and agent-27's verdict stops at G2.
    for (id, body) in [
        (
            150,
            "leader note\nmoai-claim-handoff: comment=100 to=agent-27 by=leader",
        ),
        (
            150,
            "moai-claim-handoff: comment=100 to=agent-27 by=agent-8",
        ),
        (90, handoff),
        (150, "moai-claim-handoff: comment=101 to=agent-27 by=leader"),
        (150, "moai-claim-handoff: comment=100 to=agent 27 by=leader"),
    ] {
        s["comments"] = json!([
            claim(100, "agent-10", H),
            note(id, body),
            verdict(200, "agent-27", H, "PASS"),
            lc()
        ]);
        assert_eq!(stopped(&gate(&s)), (false, "G2".into()), "{body}");
    }
    // The latest handoff wins: moved to agent-27, then on to agent-9.
    s["comments"] = json!([
        claim(100, "agent-10", H),
        note(150, handoff),
        note(160, "moai-claim-handoff: comment=100 to=agent-9 by=leader"),
        verdict(200, "agent-27", H, "PASS"),
        lc()
    ]);
    assert_eq!(stopped(&gate(&s)), (false, "G2".into()));
    s["comments"] = json!([
        claim(100, "agent-10", H),
        note(150, handoff),
        note(160, "moai-claim-handoff: comment=100 to=agent-9 by=leader"),
        verdict(200, "agent-9", H, "PASS"),
        lc()
    ]);
    assert_eq!(stopped(&gate(&s)), (true, "".into()));
}

/// #2860 rule 3 (shell gate_merge.sh): a verdict whose base is a short SHA passes only when the
/// collector resolved it to exactly one commit (`verdict_base_full`) that starts with it. Too
/// short, unresolved, or resolved to another commit stops at G1; the head stays 40 hex.
#[test]
fn g1_a_short_base_counts_only_when_resolved_to_one_commit() {
    let short = &B[..12];
    let line = |base: &str, head: &str| json!({"id": 200, "author_association": "MEMBER", "body": format!("AUDIT #9 (x) head {head} base {base} — verdict: PASS\n\n감사: agent-10 (예약 100).\n")});
    let mut s = passing();
    s["comments"] = json!([claim(100, "agent-10", H), line(short, H), lc()]);
    s["verdict_base_full"] = json!(B);
    assert_eq!(stopped(&gate(&s)), (true, "".into()), "{}", gate(&s));
    // Not resolved by the collector.
    s["verdict_base_full"] = Value::Null;
    assert_eq!(stopped(&gate(&s)), (false, "G1".into()));
    // Resolved to a commit the short base does not name.
    s["verdict_base_full"] = json!(OTHER);
    assert_eq!(stopped(&gate(&s)), (false, "G1".into()));
    // Six hex digits are too short even when "resolved"; a short head never counts.
    s["verdict_base_full"] = json!(B);
    s["comments"] = json!([claim(100, "agent-10", H), line(&B[..6], H), lc()]);
    assert_eq!(stopped(&gate(&s)), (false, "G1".into()));
    s["comments"] = json!([claim(100, "agent-10", H), line(B, &H[..12]), lc()]);
    assert_eq!(stopped(&gate(&s)), (false, "G1".into()));
}

/// #2860 scope add 6872037464 (shell owner_ok.py, #2831): a first-line withdrawal gives the PR
/// up for that lane; its claims before the withdrawal no longer count, so another lane's
/// claim on the current head and its PASS pass G2. A withdrawal of another lane, one not on the
/// first line, or one posted before the claim changes nothing.
#[test]
fn g2_a_withdrawn_lane_no_longer_holds_the_audit() {
    let mut s = passing();
    for line in [
        "moai-audit-withdraw: lane=agent-10",
        "moai-audit-claim-withdraw: lane=agent-10",
    ] {
        // agent-27's claim names an older head, so only the withdrawal makes it the lowest
        // claim (without it agent-10 holds the audit; mutant W1).
        s["comments"] = json!([
            claim(100, "agent-10", OTHER),
            note(150, &format!("{line}\n\nhanding the re-audit over")),
            claim(160, "agent-27", OTHER),
            verdict(200, "agent-27", H, "PASS"),
            lc()
        ]);
        assert_eq!(
            stopped(&gate(&s)),
            (true, "".into()),
            "{line}: {}",
            gate(&s)
        );
    }
    for (id, body) in [
        (150, "moai-audit-withdraw: lane=agent-9"),
        (150, "note\nmoai-audit-withdraw: lane=agent-10"),
        (90, "moai-audit-withdraw: lane=agent-10"),
    ] {
        s["comments"] = json!([
            claim(100, "agent-10", OTHER),
            note(id, body),
            claim(160, "agent-27", OTHER),
            verdict(200, "agent-27", H, "PASS"),
            lc()
        ]);
        assert_eq!(stopped(&gate(&s)), (false, "G2".into()), "{body}");
    }
}

// #2906 T (user decision 6878414880, Issue #1975): audit by risk tier. Tier 0 (docs, wording,
// tests) merges on its local-check line without a verdict; tiers 1 and 2 need the PASS as
// before; the path rule raises the tier, never lowers it. Paths are joined at run time so the
// repository's path scans do not read them as files.

const MB: &str = "4444444444444444444444444444444444444444";

fn p(parts: &[&str]) -> String {
    parts.join("/")
}

/// `passing()` as a tier-0 pull request: the body says tier 0, the merge base is known, it
/// changes one document and one test, the head's `ci` succeeded and there is no comment (C-84:
/// no local-check line either; was "the comments hold only the local-check pass line").
fn tier0() -> Value {
    let mut s = passing();
    s["body_tier"] = json!(0);
    s["merge_base"] = json!(MB);
    s["pr_files"] = json!([doc(), p(&["tests", "x.rs"])]);
    s["comments"] = json!([]);
    s
}

fn tier_of(v: &Value) -> Value {
    v["result"]["effective_tier"].clone()
}

/// AC25 (C-84: on the head's `ci` success; was "a local-check pass"): a tier-0 body with only
/// document and test files merges with no verdict and no claim; the base it names is the merge
/// base. Without a merge base it stops.
#[test]
fn tier0_docs_and_tests_merge_on_the_ci_success_alone() {
    let v = gate(&tier0());
    assert_eq!(v["result"]["pass"], json!(true), "{v}");
    assert_eq!(tier_of(&v), json!(0), "{v}");
    assert_eq!(v["result"]["body_tier"], json!(0), "{v}");
    assert_eq!(v["result"]["verdict_base"], json!(MB), "{v}");
    // A root Markdown file is tier-0-safe too.
    let mut s = tier0();
    s["pr_files"] = json!(["README.md"]);
    assert_eq!(gate(&s)["result"]["pass"], json!(true), "{}", gate(&s));
    let mut s = tier0();
    s.as_object_mut().unwrap().remove("merge_base");
    let v = gate(&s);
    assert_eq!(stopped(&v), (false, "tier".into()), "{v}");
    assert!(
        v["result"]["reason"]
            .as_str()
            .unwrap()
            .contains("merge base unknown"),
        "{v}"
    );
}

/// AC25b: a tier-0 body that changes a code file is tier 1 and needs its PASS (G1). Skill
/// Markdown is product instructions, and agent instruction files too: never tier 0.
#[test]
fn tier0_with_code_or_skill_text_is_tier1_and_needs_a_pass() {
    for extra in [
        p(&["src", "foo.rs"]),
        p(&["skills", "x.md"]),
        p(&["skills", "moai", "workflows", "a.md"]),
        "CLAUDE.md".to_string(),
        p(&["sub", "AGENTS.md"]),
        p(&["tools", "x.py"]),
    ] {
        let mut s = tier0();
        s["pr_files"] = json!([doc(), extra.clone()]);
        let v = gate(&s);
        assert_eq!(stopped(&v), (false, "G1".into()), "{extra}: {v}");
        assert_eq!(tier_of(&v), json!(1), "{extra}: {v}");
    }
    // `skills/x.md` alone is not tier 0 either.
    let mut s = tier0();
    s["pr_files"] = json!([p(&["skills", "x.md"])]);
    assert_eq!(tier_of(&gate(&s)), json!(1));
    // With its PASS (the passing() comments) a tier-1 PR merges as before.
    let mut s = passing();
    s["body_tier"] = json!(0);
    s["merge_base"] = json!(MB);
    let v = gate(&s);
    assert_eq!(v["result"]["pass"], json!(true), "{v}");
    assert_eq!(tier_of(&v), json!(1), "{v}");
    assert_eq!(v["result"]["verdict_base"], json!(B), "{v}");
}

/// AC26: a merge gate, install, hook, update, HOME, migration, key, CI or host settings path
/// forces tier 2 whatever the body says; the PASS is still needed.
#[test]
fn the_path_rule_forces_tier2() {
    for path in [
        p(&["src", "gate.rs"]),
        p(&["src", "gate_collect.rs"]),
        p(&["src", "ops", "gate_merge.rs"]),
        p(&["src", "install.rs"]),
        p(&["src", "agent_install.rs"]),
        p(&["src", "hook_engine.rs"]),
        p(&["src", "hooks.rs"]),
        p(&["src", "update.rs"]),
        p(&["src", "self_update.rs"]),
        p(&["src", "home_state.rs"]),
        p(&["src", "migrate.rs"]),
        p(&["src", "go_state.rs"]),
        p(&["src", "key_vault.rs"]),
        p(&["src", "worktree_include.rs"]),
        // The secret-copy rule file itself (#2967 audit B2): joined here as in src/gate.rs.
        format!(".worktree{}", "include"),
        p(&["tools", "local_check.sh"]),
        p(&[".github", "workflows", "ci.yml"]),
        p(&[".claude", "settings.json"]),
        p(&[".codex", "config.toml"]),
        "install.sh".to_string(),
        p(&["scripts", "install.sh"]),
    ] {
        for body in [0, 1] {
            let mut s = tier0();
            s["body_tier"] = json!(body);
            s["pr_files"] = json!([doc(), path.clone()]);
            let v = gate(&s);
            assert_eq!(stopped(&v), (false, "G1".into()), "{path} body {body}: {v}");
            assert_eq!(tier_of(&v), json!(2), "{path} body {body}: {v}");
        }
    }
    // Documents and tests of those areas are not the code: tier 0 stays 0.
    let mut s = tier0();
    s["pr_files"] = json!([p(&["tests", "gate_check_cli.rs"]), doc()]);
    assert_eq!(tier_of(&gate(&s)), json!(0));
    // With its PASS a tier-2 PR merges as before.
    let mut s = passing();
    s["body_tier"] = json!(0);
    s["pr_files"] = json!([p(&["src", "gate.rs"])]);
    let v = gate(&s);
    assert_eq!(v["result"]["pass"], json!(true), "{v}");
    assert_eq!(tier_of(&v), json!(2), "{v}");
}

/// AC27: no tier from the body (the collector found no well-formed line) or a value outside
/// 0–2 is tier 2: unknown is audited.
#[test]
fn an_unknown_body_tier_is_tier2() {
    for body in [None, Some(json!(3)), Some(json!(255))] {
        let mut s = tier0();
        match &body {
            Some(b) => s["body_tier"] = b.clone(),
            None => {
                s.as_object_mut().unwrap().remove("body_tier");
            }
        }
        let v = gate(&s);
        assert_eq!(stopped(&v), (false, "G1".into()), "{body:?}: {v}");
        assert_eq!(tier_of(&v), json!(2), "{body:?}: {v}");
    }
    // passing() has no body tier: tier 2, and it passes with its PASS.
    let v = gate(&passing());
    assert_eq!(tier_of(&v), json!(2), "{v}");
    assert_eq!(v["result"]["body_tier"], json!(null), "{v}");
}

/// Tier 0 skips only the audit rules: the head's `ci` (G5, C-84), the overlap with the tip
/// (G6), a red tip (G8), the body checker (G4), the claim rules (G10, G11), unknown authors
/// (G9) and an open FAIL (G3) still stop it.
#[test]
fn tier0_still_meets_every_other_rule() {
    let mut s = tier0();
    s["checks"] = json!([]);
    assert_eq!(stopped(&gate(&s)), (false, "G5".into()));
    let mut s = tier0();
    s["checks"] = json!([run("ci", Some("failure"), H)]);
    assert_eq!(stopped(&gate(&s)), (false, "G5".into()));
    let mut s = tier0();
    s["tip_files_since_verdict_base"] = json!([p(&["tests", "x.rs"])]);
    let v = gate(&s);
    assert_eq!(stopped(&v), (false, "G6".into()), "{v}");
    assert!(v["result"]["reason"]
        .as_str()
        .unwrap()
        .contains("overlap in non-doc file"));
    let mut s = tier0();
    s["merge_clean"] = json!(false);
    assert_eq!(stopped(&gate(&s)), (false, "G6".into()));
    let mut s = tier0();
    s["tip_ci"] =
        json!({"state": "red", "tip": OTHER, "decided_by": OTHER, "detail": "ci failure run 7"});
    let v = gate(&s);
    assert_eq!(stopped(&v), (false, "G8".into()), "{v}");
    assert_eq!(tier_of(&v), json!(0), "{v}");
    let mut s = tier0();
    s["body_check_ok"] = json!(false);
    assert_eq!(stopped(&gate(&s)), (false, "G4".into()));
    let mut s = tier0();
    s["commits_without_claim"] = json!([OTHER]);
    assert_eq!(stopped(&gate(&s)), (false, "G10".into()));
    let mut s = tier0();
    s["body_claim"] = json!(false);
    assert_eq!(stopped(&gate(&s)), (false, "G11".into()));
    let mut s = tier0();
    s.as_object_mut().unwrap().remove("pr_authors");
    assert_eq!(stopped(&gate(&s)), (false, "G9".into()));
    let mut s = tier0();
    s["stacked_below"] = json!(null);
    assert_eq!(stopped(&gate(&s)), (false, "G12".into()));
    // An auditor's FAIL on this head is not ignored because the body says tier 0.
    let mut s = tier0();
    s["comments"] = json!([verdict(200, "agent-10", H, "FAIL"), lc()]);
    assert_eq!(stopped(&gate(&s)), (false, "G3".into()), "{}", gate(&s));
    // C-84 (leader decision 2026-11-17: a posted FAIL stops the merge at every tier): a FAIL
    // on an older head blocks a tier-0 merge too, until the same auditor posts a later PASS
    // (was: "A FAIL on an older head does not block the current one").
    s["comments"] = json!([verdict(200, "agent-10", OTHER, "FAIL"), lc()]);
    let v = gate(&s);
    assert_eq!(stopped(&v), (false, "G3".into()), "{v}");
}

// ---- C-84 (user decision 2026-11-17 "CI 성공 + PASS"): G5 reads the head's GitHub `ci` ----

/// One check run of the snapshot: name, conclusion (`None` while it runs) and the head it ran on.
fn run(name: &str, conclusion: Option<&str>, head: &str) -> Value {
    json!({"name": name, "conclusion": conclusion, "head_sha": head})
}

/// `passing()` with only these check runs and without its local-check line.
fn with_checks(checks: Value) -> Value {
    let mut s = passing();
    s["comments"] = json!([
        claim(100, "agent-10", H),
        verdict(200, "agent-10", H, "PASS")
    ]);
    s["checks"] = checks;
    s
}

/// C-84 AC: an exact-head `ci` success passes G5 with no `local-check:` line anywhere; the
/// earlier name of the aggregate, `rc-ci`, counts the same.
#[test]
fn g5_passes_on_an_exact_head_ci_success_without_a_local_check_line() {
    for aggregate in ["ci", "rc-ci"] {
        let s = with_checks(json!([run(aggregate, Some("success"), H)]));
        let v = gate(&s);
        assert_eq!(stopped(&v), (true, String::new()), "{aggregate}: {v}");
    }
}

/// C-84 N/N+1 on the head: the success on this exact head passes (N); the same success on the
/// head before it (an older push) stops G5 (N+1), and so does a run that names no head.
#[test]
fn g5_needs_the_ci_success_on_this_exact_head() {
    let s = with_checks(json!([run("ci", Some("success"), H)]));
    assert_eq!(stopped(&gate(&s)), (true, String::new()));
    let s = with_checks(json!([run("ci", Some("success"), OTHER)]));
    assert!(g5_reason(&s).contains("not on this head"), "{}", gate(&s));
    let s = with_checks(json!([{"name": "ci", "conclusion": "success"}]));
    assert!(g5_reason(&s).contains("names no head"), "{}", gate(&s));
    // An older head's success never outweighs this head's failure.
    let s = with_checks(json!([
        run("ci", Some("success"), OTHER),
        run("ci", Some("failure"), H)
    ]));
    assert!(
        g5_reason(&s).contains("failed on this head"),
        "{}",
        gate(&s)
    );
}

/// C-84: a failed, cancelled or timed-out `ci` (or one of its jobs, or `windows-build`) stops
/// G5 and names the run; a running one stops it as still running; no `ci` aggregate at all, or
/// only unrelated checks, stops it as missing. A skipped `windows-build` is allowed.
#[test]
fn g5_stops_on_a_failed_cancelled_running_or_missing_ci() {
    for (checks, reason) in [
        (
            json!([run("ci", Some("failure"), H)]),
            "failed on this head",
        ),
        (
            json!([run("ci", Some("cancelled"), H)]),
            "failed on this head",
        ),
        (
            json!([run("ci", Some("timed_out"), H)]),
            "failed on this head",
        ),
        (
            json!([
                run("ci", Some("success"), H),
                run("ci / linux-x86_64", Some("failure"), H)
            ]),
            "failed on this head",
        ),
        (
            json!([
                run("ci", Some("success"), H),
                run("windows-build", Some("failure"), H)
            ]),
            "failed on this head",
        ),
        (json!([run("ci", None, H)]), "still running"),
        (
            json!([run("ci / linux-x86_64", Some("success"), H)]),
            "no `ci` check run",
        ),
        (
            json!([run("lint", Some("success"), H)]),
            "no `ci` check run",
        ),
        (json!([]), "no `ci` check run"),
    ] {
        let s = with_checks(checks.clone());
        assert!(g5_reason(&s).contains(reason), "{checks}: {}", gate(&s));
    }
    let s = with_checks(json!([
        run("ci", Some("success"), H),
        run("windows-build", Some("skipped"), H),
        run("lint", Some("failure"), H)
    ]));
    assert_eq!(stopped(&gate(&s)), (true, String::new()));
}

/// C-84: the `local-check:` line is optional evidence and never decides G5: a pass line does
/// not make up for a missing or failed `ci`, and a fail line does not stop a `ci` success.
#[test]
fn a_local_check_line_never_decides_g5() {
    let mut s = with_checks(json!([]));
    s["comments"].as_array_mut().unwrap().push(lc());
    assert!(g5_reason(&s).contains("no `ci` check run"), "{}", gate(&s));
    let mut s = with_checks(json!([run("ci", Some("failure"), H)]));
    s["comments"].as_array_mut().unwrap().push(lc());
    assert!(
        g5_reason(&s).contains("failed on this head"),
        "{}",
        gate(&s)
    );
    let mut s = with_checks(json!([run("ci", Some("success"), H)]));
    s["comments"]
        .as_array_mut()
        .unwrap()
        .push(local_check(900, H, "fail"));
    assert_eq!(stopped(&gate(&s)), (true, String::new()));
}

/// C-84 (leader decision 2026-11-17): at tier 0 a posted FAIL on any head stops the merge
/// (G3) until the same auditor posts a later PASS; with that PASS, or with no FAIL, the `ci`
/// success alone merges it.
#[test]
fn a_posted_fail_stops_a_tier_0_merge_until_the_same_auditor_passes() {
    let mut s = tier0();
    s["comments"] = json!([]);
    assert_eq!(stopped(&gate(&s)), (true, String::new()), "{}", gate(&s));
    s["comments"] = json!([verdict(200, "agent-10", OTHER, "FAIL")]);
    assert_eq!(stopped(&gate(&s)), (false, "G3".into()), "{}", gate(&s));
    // Another auditor's later PASS does not lift it; the same auditor's does.
    s["comments"] = json!([
        verdict(200, "agent-10", OTHER, "FAIL"),
        verdict(300, "agent-27", H, "PASS")
    ]);
    assert_eq!(stopped(&gate(&s)), (false, "G3".into()), "{}", gate(&s));
    s["comments"] = json!([
        verdict(200, "agent-10", OTHER, "FAIL"),
        verdict(300, "agent-10", H, "PASS")
    ]);
    assert_eq!(stopped(&gate(&s)), (true, String::new()), "{}", gate(&s));
    // An earlier PASS does not lift a later FAIL.
    s["comments"] = json!([
        verdict(200, "agent-10", H, "PASS"),
        verdict(300, "agent-10", OTHER, "FAIL")
    ]);
    assert_eq!(stopped(&gate(&s)), (false, "G3".into()), "{}", gate(&s));
}
