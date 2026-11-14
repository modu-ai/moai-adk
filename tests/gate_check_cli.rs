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

fn passing() -> Value {
    json!({
        "schema_version": 1,
        "pr": 9,
        "head": H,
        "comments": [claim(100, "agent-10", H), verdict(200, "agent-10", H, "PASS")],
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
        "body_claim": true
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
    s["comments"] = json!([claim(100, "agent-10", H)]);
    assert_eq!(stopped(&gate(&s)), (false, "G1".into()));
    s["comments"] = json!([
        claim(100, "agent-10", H),
        verdict(200, "agent-10", OTHER, "PASS")
    ]);
    assert_eq!(stopped(&gate(&s)), (false, "G1".into()));
    s["comments"] = json!([
        claim(100, "agent-10", H),
        verdict(200, "agent-10", H, "FAIL")
    ]);
    assert_eq!(stopped(&gate(&s)), (false, "G1".into()));
    // A short base is not a verdict line.
    s["comments"] = json!([claim(100, "agent-10", H), {"id": 200, "body": format!("AUDIT #9 head {H} base 2222 — verdict: PASS\nagent-10")}]);
    assert_eq!(stopped(&gate(&s)), (false, "G1".into()));
    // The latest verdict counts: an older PASS on this head does not outlive a newer one elsewhere.
    s["comments"] = json!([
        claim(100, "agent-10", H),
        verdict(200, "agent-10", H, "PASS"),
        verdict(300, "agent-10", OTHER, "PASS")
    ]);
    assert_eq!(stopped(&gate(&s)), (false, "G1".into()));
}

#[test]
fn g2_the_auditor_holds_the_lowest_claim_or_the_lowest_on_this_head() {
    let mut s = passing();
    // A lane with no claim wrote the PASS.
    s["comments"] = json!([
        claim(100, "agent-10", H),
        verdict(200, "agent-27", H, "PASS")
    ]);
    assert_eq!(stopped(&gate(&s)), (false, "G2".into()));
    // The PR's lowest claim (on an older head) still owns the re-audit.
    s["comments"] = json!([
        claim(100, "agent-10", OTHER),
        claim(150, "agent-27", H),
        verdict(200, "agent-10", H, "PASS")
    ]);
    assert!(stopped(&gate(&s)).0);
    // The lowest claim on the current head is allowed too.
    s["comments"] = json!([
        claim(100, "agent-10", OTHER),
        claim(150, "agent-27", H),
        verdict(200, "agent-27", H, "PASS")
    ]);
    assert!(stopped(&gate(&s)).0);
    // Only the lowest on this head: a second claim on the head is not allowed.
    s["comments"] = json!([
        claim(100, "agent-10", OTHER),
        claim(150, "agent-27", H),
        claim(160, "agent-8", H),
        verdict(200, "agent-8", H, "PASS")
    ]);
    assert_eq!(stopped(&gate(&s)), (false, "G2".into()));
    // The auditor is the first lane after the AUDIT line, not one mentioned later.
    s["comments"] = json!([claim(100, "agent-10", H), {"id": 200, "body": format!("AUDIT #9 head {H} base {B} — verdict: PASS\n\n감사: agent-27, 저자 agent-10.\n")}]);
    assert_eq!(stopped(&gate(&s)), (false, "G2".into()));
    // A verdict naming no lane stops.
    s["comments"] = json!([claim(100, "agent-10", H), {"id": 200, "body": format!("AUDIT #9 head {H} base {B} — verdict: PASS\n\nok\n")}]);
    assert_eq!(stopped(&gate(&s)), (false, "G2".into()));
    // A lane name runs to its end: `agent-10x` is not `agent-10`.
    s["comments"] = json!([claim(100, "agent-10", H), {"id": 200, "body": format!("AUDIT #9 head {H} base {B} — verdict: PASS\n\n감사: agent-10x.\n")}]);
    assert_eq!(stopped(&gate(&s)), (false, "G2".into()));
    // `agent-100` is not `agent-10`.
    s["comments"] = json!([
        claim(100, "agent-10", H),
        verdict(200, "agent-100", H, "PASS")
    ]);
    assert_eq!(stopped(&gate(&s)), (false, "G2".into()));
    // No claim at all: nothing to check.
    s["comments"] = json!([verdict(200, "agent-27", H, "PASS")]);
    assert!(stopped(&gate(&s)).0);
}

#[test]
fn g3_a_fail_on_this_head_needs_the_same_auditor_to_reverse_it() {
    let mut s = passing();
    s["comments"] = json!([
        claim(100, "agent-10", H),
        verdict(150, "agent-10", H, "FAIL"),
        verdict(200, "agent-10", H, "PASS")
    ]);
    assert!(stopped(&gate(&s)).0);
    s["comments"] = json!([
        claim(100, "agent-10", H),
        claim(120, "agent-27", H),
        verdict(150, "agent-27", H, "FAIL"),
        verdict(200, "agent-10", H, "PASS")
    ]);
    assert_eq!(stopped(&gate(&s)), (false, "G3".into()));
    // A FAIL on an older head does not block the new head.
    s["comments"] = json!([
        claim(100, "agent-10", H),
        verdict(150, "agent-27", OTHER, "FAIL"),
        verdict(200, "agent-10", H, "PASS")
    ]);
    assert!(stopped(&gate(&s)).0);
}

#[test]
fn g4_the_self_audit_checker_must_pass() {
    let mut s = passing();
    s["body_check_ok"] = json!(false);
    assert_eq!(stopped(&gate(&s)), (false, "G4".into()));
}

#[test]
fn g5_checks_need_the_aggregate_and_no_other_outcome() {
    let mut s = passing();
    s["checks"] = json!([{"name": "ci / linux-x86_64", "conclusion": "success"}]);
    assert_eq!(stopped(&gate(&s)), (false, "G5".into()));
    s["checks"] = json!([{"name": "ci", "conclusion": "success"}, {"name": "ci / linux-x86_64", "conclusion": null}]);
    assert_eq!(stopped(&gate(&s)), (false, "G5".into()));
    s["checks"] = json!([{"name": "ci", "conclusion": "success"}, {"name": "ci / linux-x86_64", "conclusion": "skipped"}]);
    assert_eq!(stopped(&gate(&s)), (false, "G5".into()));
    s["checks"] = json!([{"name": "ci", "conclusion": "success"}, {"name": "windows-build", "conclusion": "failure"}]);
    assert_eq!(stopped(&gate(&s)), (false, "G5".into()));
    // Only `windows-build` may be skipped; unrelated checks are not read.
    s["checks"] = json!([{"name": "ci", "conclusion": "success"}, {"name": "windows-build", "conclusion": "skipped"}, {"name": "lint", "conclusion": "failure"}]);
    assert!(stopped(&gate(&s)).0);
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
        quoted
    ]);
    assert_eq!(stopped(&gate(&s)), (false, "G1".into()));
    // With the FAIL reversed by a real first-line PASS, a quote after it changes nothing.
    s["comments"] = json!([
        claim(100, "agent-10", H),
        verdict(150, "agent-10", H, "FAIL"),
        verdict(200, "agent-10", H, "PASS"),
        quoted
    ]);
    assert!(stopped(&gate(&s)).0);
    // A quoted FAIL does not block either.
    let quoted_fail = json!({"id": 300, "body": format!("note\nAUDIT #9 (x) head {H} base {B} — verdict: FAIL\nagent-27\n")});
    s["comments"] = json!([
        claim(100, "agent-10", H),
        verdict(200, "agent-10", H, "PASS"),
        quoted_fail
    ]);
    assert!(stopped(&gate(&s)).0);
}

/// Audit N1–N4: a title containing `head` still parses; `rc-ci` counts as the aggregate; a
/// withdrawn lowest claim (machine form) hands G2 to the next one, a free-text one does not;
/// an unknown verdict base stops G1.
#[test]
fn audit_notes_title_head_rc_ci_withdrawal_and_known_base() {
    let mut s = passing();
    s["comments"] = json!([claim(100, "agent-10", H), {"id": 200, "body": format!("AUDIT #9 (fix head recheck) head {H} base {B} — verdict: PASS\n\n감사: agent-10.\n")}]);
    assert!(stopped(&gate(&s)).0);

    let mut s = passing();
    s["checks"] = json!([{"name": "rc-ci", "conclusion": "success"}, {"name": "ci / linux-x86_64", "conclusion": "success"}]);
    assert!(stopped(&gate(&s)).0);
    s["checks"] = json!([{"name": "ci", "conclusion": "success"}, {"name": "rc-ci", "conclusion": "failure"}]);
    assert_eq!(stopped(&gate(&s)), (false, "G5".into()));

    let mut s = passing();
    let withdraw =
        json!({"id": 120, "body": "moai-audit-claim 철회: agent-10 (100) — agent-27 takes it\n"});
    s["comments"] = json!([
        claim(100, "agent-10", H),
        withdraw,
        claim(130, "agent-27", H),
        verdict(200, "agent-27", H, "PASS")
    ]);
    assert!(stopped(&gate(&s)).0);
    let free_text = json!({"id": 120, "body": "agent-10: 감사 예약 100 철회\n"});
    s["comments"] = json!([
        claim(100, "agent-10", H),
        free_text,
        claim(130, "agent-27", OTHER),
        verdict(200, "agent-27", H, "PASS")
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
/// failed `rc-ci / …` job stops G5 like a `ci / …` one.
#[test]
fn a_withdrawal_must_name_the_claims_own_lane_and_rc_ci_jobs_count() {
    let mut s = passing();
    let wrong = json!({"id": 3, "body": "moai-audit-claim 철회: agent-10 (1)\n"});
    s["comments"] = json!([
        claim(1, "agent-22", H),
        claim(2, "agent-10", H),
        wrong,
        verdict(200, "agent-10", H, "PASS")
    ]);
    assert_eq!(stopped(&gate(&s)), (false, "G2".into()));
    let right = json!({"id": 3, "body": "moai-audit-claim 철회: agent-22 (1)\n"});
    s["comments"] = json!([
        claim(1, "agent-22", H),
        claim(2, "agent-10", H),
        right,
        verdict(200, "agent-10", H, "PASS")
    ]);
    assert!(stopped(&gate(&s)).0);

    let mut s = passing();
    s["checks"] = json!([{"name": "rc-ci", "conclusion": "success"}, {"name": "rc-ci / linux-x86_64", "conclusion": "failure"}]);
    assert_eq!(stopped(&gate(&s)), (false, "G5".into()));
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
        verdict(300, "agent-26", H, "PASS")
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
        verdict(200, "agent-10", OTHER, "PASS")
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
        verdict(200, "agent-10", OTHER, "FAIL")
    ]);
    assert_eq!(stopped(&gate(&s)), (false, "G1".into()));
}

/// #2833 scope 4: the CI on the current head still decides (G5).
#[test]
fn a_mechanical_delta_still_needs_the_current_heads_ci() {
    let mut s = on_delta();
    s["checks"] = json!([
        {"name": "ci", "conclusion": "failure"},
        {"name": "ci / linux-x86_64", "conclusion": "failure"}
    ]);
    assert_eq!(stopped(&gate(&s)), (false, "G5".into()));
}

/// #2833 D2: the audit rules read the verdict head. An unresolved FAIL there still blocks
/// (G3), and the auditor's claim on that head counts (G2).
#[test]
fn the_audit_rules_apply_to_the_verdict_head_of_a_delta() {
    let mut s = on_delta();
    s["comments"] = json!([
        claim(100, "agent-10", OTHER),
        verdict(150, "agent-27", OTHER, "FAIL"),
        verdict(200, "agent-10", OTHER, "PASS")
    ]);
    assert_eq!(stopped(&gate(&s)), (false, "G3".into()));
    // The same auditor reversing itself on the verdict head passes, as on a plain head.
    s["comments"] = json!([
        claim(100, "agent-10", OTHER),
        verdict(150, "agent-10", OTHER, "FAIL"),
        verdict(200, "agent-10", OTHER, "PASS")
    ]);
    assert_eq!(stopped(&gate(&s)), (true, String::new()));
    // G2: the lowest claim is another lane's on an older head; the claim on the verdict head
    // is the auditor's, so it is allowed.
    s["comments"] = json!([
        claim(90, "agent-27", B),
        claim(100, "agent-10", OTHER),
        verdict(200, "agent-10", OTHER, "PASS")
    ]);
    assert_eq!(stopped(&gate(&s)), (true, String::new()));
}
