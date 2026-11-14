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
    json!({"id": id, "body": format!("AUDIT #9 (x) head {head} base {B} — verdict: {v}\n\n감사: {lane} (예약 100).\n")})
}

fn claim(id: u64, lane: &str, head: &str) -> Value {
    json!({"id": id, "body": format!("moai-audit-claim: lane={lane} head={head}\n")})
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
            {"name": "ci", "conclusion": "success"},
            {"name": "ci / linux-x86_64", "conclusion": "success"},
            {"name": "windows-build", "conclusion": "skipped"},
            {"name": "Vercel Preview Comments", "conclusion": "failure"}
        ],
        "pr_files": ["src/a.rs", doc()],
        "tip_files_since_verdict_base": ["src/b.rs"],
        "merge_clean": true,
        "body_check_ok": true,
        "verdict_base_known": true,
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
    s["comments"] = json!([claim(100, "agent-10", H), {"id": 200, "body": format!("AUDIT #9 head {H} base 2222 — verdict: PASS\nagent-10")}, lc()]);
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
    s["comments"] = json!([claim(100, "agent-10", H), {"id": 200, "body": format!("AUDIT #9 head {H} base {B} — verdict: PASS\n\n감사: agent-27, 저자 agent-10.\n")}, lc()]);
    assert_eq!(stopped(&gate(&s)), (false, "G2".into()));
    // A verdict naming no lane stops.
    s["comments"] = json!([claim(100, "agent-10", H), {"id": 200, "body": format!("AUDIT #9 head {H} base {B} — verdict: PASS\n\nok\n")}, lc()]);
    assert_eq!(stopped(&gate(&s)), (false, "G2".into()));
    // A lane name runs to its end: `agent-10x` is not `agent-10`.
    s["comments"] = json!([claim(100, "agent-10", H), {"id": 200, "body": format!("AUDIT #9 head {H} base {B} — verdict: PASS\n\n감사: agent-10x.\n")}, lc()]);
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

/// #2906 S5b: the check-run cases that stopped G5 before (no aggregate, a running, skipped or
/// failed job) no longer stop the merge; `ci_info` names them. Only `windows-build` may be
/// skipped; unrelated checks are not read.
#[test]
fn g5_check_runs_without_the_aggregate_or_with_a_bad_job_only_inform() {
    let mut s = passing();
    for (checks, info) in [
        (
            json!([{"name": "ci / linux-x86_64", "conclusion": "success"}]),
            "ci=pending",
        ),
        (
            json!([{"name": "ci", "conclusion": "success"}, {"name": "ci / linux-x86_64", "conclusion": null}]),
            "ci=pending",
        ),
        (
            json!([{"name": "ci", "conclusion": "success"}, {"name": "ci / linux-x86_64", "conclusion": "skipped"}]),
            "ci=failure",
        ),
        (
            json!([{"name": "ci", "conclusion": "success"}, {"name": "windows-build", "conclusion": "failure"}]),
            "ci=failure",
        ),
        (
            json!([{"name": "ci", "conclusion": "success"}, {"name": "windows-build", "conclusion": "skipped"}, {"name": "lint", "conclusion": "failure"}]),
            "ci=success",
        ),
    ] {
        s["checks"] = checks;
        let v = gate(&s);
        assert_eq!(stopped(&v), (true, String::new()), "{v}");
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
    json!({"id": id, "body": format!("{}\n", local_check_line(head, result))})
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

/// #2906 S5b AC21: G5 passes only on a well-formed `local-check:` line for this head whose
/// result is `pass`; the last such line (highest comment id, last line in it) decides.
#[test]
fn g5_the_last_local_check_line_for_this_head_decides() {
    const NONE: &str = "no local-check line for this head";
    const FAILED: &str = "local-check result=fail on this head";
    // No line at all, or one for another head.
    assert!(g5_reason(&with_local_checks(&[])).contains(NONE));
    assert!(g5_reason(&with_local_checks(&[local_check(300, OTHER, "pass")])).contains(NONE));
    // A failed run on this head.
    assert!(g5_reason(&with_local_checks(&[local_check(300, H, "fail")])).contains(FAILED));
    // Malformed lines count as absent: 39 or 41 hex, uppercase, no `at=`, another command, an
    // extra field, a bad time, another result word, not at column 0, quoted.
    let at = "at=2026-11-14T12:00:00Z";
    let cmd = "cmd=tools/local_check.sh";
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
        let s = with_local_checks(&[json!({"id": 300, "body": line.clone()})]);
        assert!(g5_reason(&s).contains(NONE), "{line}");
    }
    // The well-formed line passes (N), on any line of a comment.
    let s = with_local_checks(&[
        json!({"id": 300, "body": format!("ran it\n\n{}\n", local_check_line(H, "pass"))}),
    ]);
    assert_eq!(stopped(&gate(&s)), (true, String::new()), "{}", gate(&s));
    // The last line for this head decides, across comments and within one.
    let s = with_local_checks(&[local_check(300, H, "pass"), local_check(310, H, "fail")]);
    assert!(g5_reason(&s).contains(FAILED));
    let s = with_local_checks(&[local_check(300, H, "fail"), local_check(310, H, "pass")]);
    assert_eq!(stopped(&gate(&s)), (true, String::new()), "{}", gate(&s));
    let both = |a: &str, b: &str| json!({"id": 300, "body": format!("{}\n{}\n", local_check_line(H, a), local_check_line(H, b))});
    assert!(g5_reason(&with_local_checks(&[both("pass", "fail")])).contains(FAILED));
    let s = with_local_checks(&[both("fail", "pass")]);
    assert_eq!(stopped(&gate(&s)), (true, String::new()), "{}", gate(&s));
    // Comment ids order the lines, not their place in the list.
    let s = with_local_checks(&[local_check(310, H, "fail"), local_check(300, H, "pass")]);
    assert!(g5_reason(&s).contains(FAILED));
    // A line for another head after this head's pass changes nothing.
    for r in ["pass", "fail"] {
        let s = with_local_checks(&[local_check(300, H, "pass"), local_check(310, OTHER, r)]);
        assert_eq!(stopped(&gate(&s)), (true, String::new()), "{}", gate(&s));
    }
}

/// #2906 S5b: GitHub check runs no longer block the merge; `ci_info` reports them.
#[test]
fn g5_check_runs_are_information_only() {
    let mut s = with_local_checks(&[local_check(300, H, "pass")]);
    let v = gate(&s);
    assert_eq!(stopped(&v), (true, String::new()), "{v}");
    assert_eq!(v["result"]["ci_info"], json!("ci=success"), "{v}");
    for (checks, info) in [
        (
            json!([{"name": "ci", "conclusion": "failure"}, {"name": "ci / linux-x86_64", "conclusion": "failure"}]),
            "ci=failure",
        ),
        (json!([{"name": "ci", "conclusion": null}]), "ci=pending"),
        (
            json!([{"name": "lint", "conclusion": "failure"}]),
            "ci=none",
        ),
        (json!([]), "ci=none"),
    ] {
        s["checks"] = checks;
        let v = gate(&s);
        assert_eq!(stopped(&v), (true, String::new()), "{v}");
        assert_eq!(v["result"]["ci_info"], json!(info), "{v}");
    }
    // A stop carries it too.
    s["comments"] = json!([
        claim(100, "agent-10", H),
        verdict(200, "agent-10", H, "PASS")
    ]);
    let v = gate(&s);
    assert_eq!(stopped(&v), (false, "G5".into()), "{v}");
    assert_eq!(v["result"]["ci_info"], json!("ci=none"), "{v}");
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
    let quoted = json!({"id": 300, "body": format!("수리 push — agent-10에게\nAUDIT #9 (x) head {H} base {B} — verdict: PASS\nagent-10\n")});
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
    let quoted_fail = json!({"id": 300, "body": format!("note\nAUDIT #9 (x) head {H} base {B} — verdict: FAIL\nagent-27\n")});
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
    s["comments"] = json!([claim(100, "agent-10", H), {"id": 200, "body": format!("AUDIT #9 (fix head recheck) head {H} base {B} — verdict: PASS\n\n감사: agent-10.\n")}, lc()]);
    assert!(stopped(&gate(&s)).0);

    let mut s = passing();
    s["checks"] = json!([{"name": "rc-ci", "conclusion": "success"}, {"name": "ci / linux-x86_64", "conclusion": "success"}]);
    assert_eq!(gate(&s)["result"]["ci_info"], json!("ci=success"));
    // #2906 S5b: a failed `rc-ci` is reported, not a stop.
    s["checks"] = json!([{"name": "ci", "conclusion": "success"}, {"name": "rc-ci", "conclusion": "failure"}]);
    let v = gate(&s);
    assert!(stopped(&v).0, "{v}");
    assert_eq!(v["result"]["ci_info"], json!("ci=failure"), "{v}");

    let mut s = passing();
    let withdraw =
        json!({"id": 120, "body": "moai-audit-claim 철회: agent-10 (100) — agent-27 takes it\n"});
    s["comments"] = json!([
        claim(100, "agent-10", H),
        withdraw,
        claim(130, "agent-27", H),
        verdict(200, "agent-27", H, "PASS"),
        lc()
    ]);
    assert!(stopped(&gate(&s)).0);
    let free_text = json!({"id": 120, "body": "agent-10: 감사 예약 100 철회\n"});
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

/// Audit 6850638994 B2: a withdrawal removes a claim only when its lane matches that claim's
/// lane; a mistyped id naming another lane's claim leaves it in place and G2 stops. N5: a
/// failed `rc-ci / …` job counts like a `ci / …` one (since #2906 S5b in `ci_info`, not G5).
#[test]
fn a_withdrawal_must_name_the_claims_own_lane_and_rc_ci_jobs_count() {
    let mut s = passing();
    let wrong = json!({"id": 3, "body": "moai-audit-claim 철회: agent-10 (1)\n"});
    s["comments"] = json!([
        claim(1, "agent-22", H),
        claim(2, "agent-10", H),
        wrong,
        verdict(200, "agent-10", H, "PASS"),
        lc()
    ]);
    assert_eq!(stopped(&gate(&s)), (false, "G2".into()));
    let right = json!({"id": 3, "body": "moai-audit-claim 철회: agent-22 (1)\n"});
    s["comments"] = json!([
        claim(1, "agent-22", H),
        claim(2, "agent-10", H),
        right,
        verdict(200, "agent-10", H, "PASS"),
        lc()
    ]);
    assert!(stopped(&gate(&s)).0);

    let mut s = passing();
    s["checks"] = json!([{"name": "rc-ci", "conclusion": "success"}, {"name": "rc-ci / linux-x86_64", "conclusion": "failure"}]);
    let v = gate(&s);
    assert!(stopped(&v).0, "{v}");
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

/// #2833 scope 4: the current head's own check still decides (G5) — since #2906 S5b its
/// `local-check:` line, not its CI: a pass line on the verdict head does not carry over, and
/// failed check runs alone do not stop.
#[test]
fn a_mechanical_delta_still_needs_the_current_heads_local_check() {
    let mut s = on_delta();
    s["comments"] = json!([
        claim(100, "agent-10", OTHER),
        verdict(200, "agent-10", OTHER, "PASS"),
        local_check(900, OTHER, "pass")
    ]);
    assert_eq!(stopped(&gate(&s)), (false, "G5".into()));
    let mut s = on_delta();
    s["checks"] = json!([
        {"name": "ci", "conclusion": "failure"},
        {"name": "ci / linux-x86_64", "conclusion": "failure"}
    ]);
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
    json!({"id": id, "body": body})
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
    let line = |base: &str, head: &str| json!({"id": 200, "body": format!("AUDIT #9 (x) head {head} base {base} — verdict: PASS\n\n감사: agent-10 (예약 100).\n")});
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
