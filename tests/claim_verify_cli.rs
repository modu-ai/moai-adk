//! rc.2 A9 #2255: claim ownership re-verification. The evaluator is pure, so every AC-1
//! scenario is a fixture of comments; the op and the hook run over recorded exchanges or a
//! stub `gh` in temporary repositories. No live GitHub call is made.
mod common;

use moai_cli::claim::{
    parse_claim_line, render_claim_line, resolve_winner, ClaimComment, ClaimFields, ClaimRecord,
    Journal, Owner, SpecRef, TaskRef, WorktreeId,
};
use moai_cli::claim_verify::{
    classify_comment, evaluate, hook_cli, is_commit_or_push, refusal, render_handoff_line,
    render_release_line, verify_cli, CommentKind, IssueComment, Verifier, VerifyRequest,
};
use moai_cli::contract::{classify, Kind};
use moai_cli::forge::policy::{parse_utc, Clock};
use moai_cli::forge::transport::{Exchange, RecordedTransport};
use serde_json::{json, Value};
use std::fs;
use std::path::Path;
use std::process::{Command, Stdio};

const NOW: &str = "2026-11-09T05:00:00Z";
const LIVE: &str = "2026-11-09T09:00:00Z";
const EXPIRED: &str = "2026-11-09T04:00:00Z";
const URL: &str = "https://api.github.com/repos/acme/widgets/issues/7/comments?per_page=100";

struct FixedClock(i64);
impl Clock for FixedClock {
    fn now(&self) -> i64 {
        self.0
    }
    fn sleep(&self, _: u64) {}
}

fn now() -> i64 {
    parse_utc(NOW).unwrap()
}

fn claim(id: i64, session: &str, lease: &str) -> IssueComment {
    IssueComment {
        id,
        author: "GoosLab".into(),
        body: format!("moai-claim: session={session} epoch=1 lease_until={lease}"),
    }
}

fn comment(id: i64, author: &str, body: &str) -> IssueComment {
    IssueComment {
        id,
        author: author.into(),
        body: body.into(),
    }
}

fn verdict(comments: &[IssueComment], session: &str, comment_id: Option<i64>) -> Value {
    evaluate(
        comments,
        &Verifier {
            session,
            comment_id,
            allowed_authors: &["GoosLab".to_string()],
            now: now(),
        },
    )
}

/// The generator writes the RC1 line first, so the RC1 parser reads it unchanged, and the
/// appended fields round-trip through the new classifier.
#[test]
fn render_claim_line_keeps_the_rc1_prefix_and_round_trips() {
    let line = render_claim_line(&ClaimFields {
        session: "agent-1-e29c15".into(),
        epoch: 2,
        lease_until: LIVE.into(),
        lane: Some("agent-1".into()),
        issue: Some(2255),
        branch: Some("issue/2255-rc2-a9-claim-enforce".into()),
        at: Some(NOW.into()),
        relay: true,
    });
    assert!(line.starts_with(&format!(
        "moai-claim: session=agent-1-e29c15 epoch=2 lease_until={LIVE} "
    )));
    for part in [
        "lane=agent-1",
        "issue=2255",
        "branch=issue/2255-rc2-a9-claim-enforce",
        &format!("at={NOW}"),
        "relay=true",
    ] {
        assert!(line.contains(part), "{line}");
    }
    let rc1 = parse_claim_line(&line).unwrap();
    assert_eq!((rc1.session.as_str(), rc1.epoch), ("agent-1-e29c15", 2));
    assert_eq!(
        classify_comment(&line),
        CommentKind::Claim {
            session: "agent-1-e29c15".into(),
            epoch: 2,
            lease_until: LIVE.into()
        }
    );
    let unverified = render_claim_line(&ClaimFields {
        session: "unverified".into(),
        epoch: 1,
        lease_until: LIVE.into(),
        ..Default::default()
    });
    assert_eq!(
        unverified,
        format!("moai-claim: session=unverified epoch=1 lease_until={LIVE}")
    );
}

/// AC-1 concurrent posts: the lowest open claim id holds; the other claimant is refused.
#[test]
fn concurrent_claims_lowest_id_holds() {
    let c = [claim(100, "a", LIVE), claim(200, "b", LIVE)];
    assert_eq!(verdict(&c, "a", Some(100))["status"], json!("owner"));
    let v = verdict(&c, "b", Some(200));
    assert_eq!(v["status"], json!("not_owner"));
    assert_eq!(v["lost_to"], json!(100));
    let msg = refusal(&v).unwrap();
    assert_eq!(classify(&msg), Kind::Conflict, "{msg}");
    assert!(refusal(&verdict(&c, "a", Some(100))).is_none());
}

/// Decision D1 (#1975 comment 6840449133, Issue #2612): a new claim line of the same session
/// renews without a handoff line when every open claim before it is that session's; another
/// session's claim in between keeps the lowest-open rule. Editing the own lease still works.
#[test]
fn same_session_new_line_renews_without_handoff() {
    let epoch = |id: i64, n: u64, lease: &str| IssueComment {
        id,
        author: "GoosLab".into(),
        body: format!("moai-claim: session=a epoch={n} lease_until={lease}"),
    };
    // Both live: the new line holds, the old id reads closed by renewal.
    let both_live = [claim(100, "a", LIVE), epoch(200, 2, LIVE)];
    let v = verdict(&both_live, "a", Some(200));
    assert_eq!(v["status"], json!("owner"));
    assert_eq!(v["holder"], json!(200));
    assert_eq!(v["closed"][0]["comment_id"], json!(100));
    assert_eq!(v["closed"][0]["by"], json!("renewal"));
    assert_eq!(
        verdict(&both_live, "a", Some(100))["status"],
        json!("closed")
    );
    // The first lapsed: still owner, no handoff line needed.
    let lapsed = [claim(100, "a", EXPIRED), epoch(200, 2, LIVE)];
    assert_eq!(verdict(&lapsed, "a", Some(200))["status"], json!("owner"));
    // Without a known id, the session's latest open claim speaks for it.
    assert_eq!(verdict(&lapsed, "a", None)["status"], json!("owner"));
    // A chain of renewals ends on the last line.
    let chain = [
        claim(100, "a", EXPIRED),
        epoch(200, 2, EXPIRED),
        epoch(300, 3, LIVE),
    ];
    let v = verdict(&chain, "a", Some(300));
    assert_eq!(v["status"], json!("owner"));
    assert_eq!(v["closed"].as_array().unwrap().len(), 2);
    // The renewing line itself lapsed: stale, never owner through an older lease.
    let renewed_lapsed = [claim(100, "a", LIVE), epoch(200, 2, EXPIRED)];
    assert_eq!(
        verdict(&renewed_lapsed, "a", Some(200))["status"],
        json!("stale")
    );
    // Another session's claim before the new line: no renewal, whether live or not.
    for lease in [LIVE, EXPIRED] {
        let between = [
            claim(100, "a", EXPIRED),
            claim(150, "b", lease),
            epoch(200, 2, LIVE),
        ];
        let v = verdict(&between, "a", Some(200));
        assert_ne!(v["status"], json!("owner"), "b lease {lease}");
        assert_eq!(v["holder"], json!(100), "b lease {lease}");
    }
    let other_first = [
        claim(50, "b", LIVE),
        claim(100, "a", LIVE),
        epoch(200, 2, LIVE),
    ];
    let v = verdict(&other_first, "a", Some(200));
    assert_eq!(v["status"], json!("not_owner"));
    assert_eq!(v["lost_to"], json!(50));
    // Editing the own comment's lease (same id, later time) keeps owner.
    assert_eq!(
        verdict(&[claim(100, "a", LIVE)], "a", Some(100))["status"],
        json!("owner")
    );
    assert_eq!(
        verdict(&[claim(100, "a", EXPIRED)], "a", Some(100))["status"],
        json!("stale")
    );
}

/// Audit FAIL 6840678988 B1 (#2621): once a renewing line is itself released or handed off,
/// the line it renewed stays closed — the next session's claim owns.
#[test]
fn renewed_line_stays_closed_after_the_renewing_line_is_released_or_handed_off() {
    let epoch2 = |lease: &str| IssueComment {
        id: 200,
        author: "GoosLab".into(),
        body: format!("moai-claim: session=a epoch=2 lease_until={lease}"),
    };
    let released = [
        claim(100, "a", LIVE),
        epoch2(LIVE),
        comment(250, "GoosLab", "moai-claim-release: comment=200 by=a"),
        claim(300, "c", LIVE),
    ];
    let v = verdict(&released, "c", Some(300));
    assert_eq!(v["status"], json!("owner"), "{v}");
    assert_eq!(v["holder"], json!(300));
    let handed = [
        claim(100, "a", EXPIRED),
        epoch2(LIVE),
        comment(
            250,
            "GoosLab",
            "moai-claim-handoff: comment=200 to=c by=leader",
        ),
        claim(300, "c", LIVE),
    ];
    let v = verdict(&handed, "c", Some(300));
    assert_eq!(v["status"], json!("owner"), "{v}");
    assert_eq!(v["holder"], json!(300));
}

/// Decision D2 (#1975 comment 6840449133): a lane's `by=timeout` handoff closes an expired
/// lease and the next claim owns; on a live lease the line is ignored and reported.
#[test]
fn timeout_handoff_closes_only_an_expired_lease() {
    let timeout = |id: i64| {
        comment(
            id,
            "GoosLab",
            "moai-claim-handoff: comment=100 to=b by=timeout",
        )
    };
    let expired = [
        claim(100, "a", EXPIRED),
        timeout(150),
        claim(200, "b", LIVE),
    ];
    let v = verdict(&expired, "b", Some(200));
    assert_eq!(v["status"], json!("owner"));
    assert_eq!(v["closed"][0]["by"], json!("handoff"));
    let live = [claim(100, "a", LIVE), timeout(150), claim(200, "b", LIVE)];
    let v = verdict(&live, "b", Some(200));
    assert_eq!(v["status"], json!("not_owner"));
    assert_eq!(v["lost_to"], json!(100));
    assert_eq!(v["invalid"][0]["comment_id"], json!(150));
    assert_eq!(
        v["invalid"][0]["detail"],
        json!("timeout handoff closes a live lease")
    );
    // A leader's handoff on a live lease still closes it.
    let leader = [
        claim(100, "a", LIVE),
        comment(
            150,
            "GoosLab",
            "moai-claim-handoff: comment=100 to=b by=leader",
        ),
        claim(200, "b", LIVE),
    ];
    assert_eq!(verdict(&leader, "b", Some(200))["status"], json!("owner"));
}

/// AC-1 late-visible competitor: the first read showed only our claim; the re-read shows a
/// lower id and the verdict flips to not_owner.
#[test]
fn late_visible_competitor_is_judged_on_reread() {
    let first = [claim(200, "b", LIVE)];
    assert_eq!(verdict(&first, "b", Some(200))["status"], json!("owner"));
    let reread = [claim(200, "b", LIVE), claim(150, "a", LIVE)];
    let v = verdict(&reread, "b", Some(200));
    assert_eq!(v["status"], json!("not_owner"));
    assert_eq!(v["lost_to"], json!(150));
}

/// Claim default (user instruction 2026-11-09, #1975): after a full re-read of every page, the
/// earliest server-registered VALID claim holds — an earlier malformed or foreign-author line
/// does not — and an existing live owner blocks a later competing claim. Through the op: the
/// competitor's claim is on the second page, so a single-page read would wrongly pass it.
#[test]
fn full_reread_earliest_valid_claim_wins_and_blocks_later_claims() {
    let page2 = "https://api.github.com/repos/acme/widgets/issues/7/comments?per_page=100&page=2";
    let mut first = get(json!([
        raw(
            40,
            "mallory",
            &format!("moai-claim: session=m epoch=1 lease_until={LIVE}")
        ),
        raw(
            50,
            "GoosLab",
            "moai-claim: session=x epoch=1 lease_until=tomorrow"
        ),
        raw(
            100,
            "GoosLab",
            &format!("moai-claim: session=a epoch=1 lease_until={LIVE}")
        ),
    ]));
    first
        .response
        .headers
        .insert("link".into(), format!("<{page2}>; rel=\"next\""));
    let mut second = get(json!([raw(
        300,
        "GoosLab",
        &format!("moai-claim: session=b epoch=1 lease_until={LIVE}")
    )]));
    second.request.url = page2.into();
    let clock = FixedClock(now());

    // The later claimant (300, page 2) is refused; the holder is 100, not 40 or 50.
    let t = RecordedTransport::new(vec![first.clone(), second.clone()]);
    let err = verify_cli(&request("b", Some(300)), &t, &clock).unwrap_err();
    assert_eq!(classify(&err), Kind::Conflict, "{err}");
    assert!(err.contains("comment 100"), "{err}");
    assert_eq!(t.sent.borrow().len(), 2, "both pages read");

    // The existing owner keeps the Issue on the same full read.
    let t = RecordedTransport::new(vec![first, second]);
    let v = verify_cli(&request("a", Some(100)), &t, &clock).unwrap();
    assert_eq!(v["status"], json!("owner"));
    assert_eq!(v["holder"], json!(100));
    let reasons: Vec<&str> = v["invalid"]
        .as_array()
        .unwrap()
        .iter()
        .map(|i| i["reason"].as_str().unwrap())
        .collect();
    assert_eq!(reasons, ["author_mismatch", "malformed"]);
}

/// AC-1 post failure / uncertain: our recorded comment id is not on the Issue.
#[test]
fn missing_own_comment_is_uncertain() {
    let c = [claim(100, "a", LIVE)];
    let v = verdict(&c, "b", Some(300));
    assert_eq!(v["status"], json!("uncertain"));
    assert_eq!(classify(&refusal(&v).unwrap()), Kind::Blocked);
}

/// AC-1 duplicate: the same session and epoch claimed twice — the first stands, the second
/// is reported and ignored.
#[test]
fn duplicate_claim_is_reported_and_ignored() {
    let c = [claim(100, "a", LIVE), claim(120, "a", LIVE)];
    let v = verdict(&c, "a", None);
    assert_eq!(v["status"], json!("owner"));
    assert_eq!(v["mine"], json!(100));
    assert_eq!(
        v["invalid"],
        json!([{"comment_id": 120, "reason": "duplicate"}])
    );
}

/// AC-1 forged: a claim line from an author outside allowed_authors is ignored, so a lower
/// forged id does not beat the real claim (REQ-A9-02).
#[test]
fn forged_author_is_ignored() {
    let forged = comment(
        50,
        "mallory",
        &format!("moai-claim: session=b epoch=1 lease_until={LIVE}"),
    );
    let c = [forged, claim(100, "a", LIVE)];
    let v = verdict(&c, "a", Some(100));
    assert_eq!(v["status"], json!("owner"));
    assert_eq!(
        v["invalid"],
        json!([{"comment_id": 50, "reason": "author_mismatch"}])
    );
}

/// AC-1 handoff and release: a closing line opens the Issue to the next claim; the closed
/// claimant itself is refused as `closed`.
#[test]
fn handoff_and_release_close_the_holder() {
    for closing in [
        render_handoff_line(100, "b", "leader"),
        render_release_line(100, "leader"),
    ] {
        let c = [
            claim(100, "a", LIVE),
            comment(150, "GoosLab", &format!("relay note\n{closing}")),
            claim(200, "b", LIVE),
        ];
        let v = verdict(&c, "b", Some(200));
        assert_eq!(v["status"], json!("owner"), "{closing}");
        assert_eq!(v["closed"][0]["comment_id"], json!(100));
        let old = verdict(&c, "a", Some(100));
        assert_eq!(old["status"], json!("closed"));
        assert_eq!(classify(&refusal(&old).unwrap()), Kind::Conflict);
    }
    assert!(matches!(
        classify_comment(&render_handoff_line(1, "x", "y")),
        CommentKind::Handoff { target: 1, .. }
    ));
}

/// AC-1 stale (REQ-A9-03): the holder's lease expired and no handoff exists — the verdict is
/// `stale` for everyone, nobody is assigned, and the holder's own commit is refused too.
#[test]
fn expired_holder_is_stale_and_never_reassigned() {
    let c = [claim(100, "a", EXPIRED), claim(200, "b", LIVE)];
    for (session, id) in [("b", 200), ("a", 100)] {
        let v = verdict(&c, session, Some(id));
        assert_eq!(v["status"], json!("stale"), "{session}");
        assert_eq!(v["holder"], json!(100));
        assert_eq!(v["auto_assigned"], json!(false));
        assert_eq!(classify(&refusal(&v).unwrap()), Kind::Blocked);
    }
}

/// AC-1 restart: a restarted session finds its claim through the journal comment id or its
/// session id; a session that restarted under a new id owns nothing.
#[test]
fn restart_keeps_ownership_only_for_the_same_claim() {
    let c = [claim(100, "a", LIVE), claim(200, "b", LIVE)];
    assert_eq!(verdict(&c, "a", Some(100))["status"], json!("owner"));
    assert_eq!(verdict(&c, "a", None)["status"], json!("owner"));
    let v = verdict(&c, "a-restarted", None);
    assert_eq!(v["status"], json!("not_owner"));
    assert_eq!(v["mine"], json!(null));
}

/// Malformed lines of each family are reported and never counted as claims.
#[test]
fn malformed_lines_are_reported() {
    let c = [
        comment(
            10,
            "GoosLab",
            "moai-claim: session=a epoch=1 lease_until=tomorrow",
        ),
        comment(11, "GoosLab", "moai-claim: session=a"),
        comment(12, "GoosLab", "moai-claim-release: by=leader"),
        comment(13, "GoosLab", "moai-claim-handoff: comment=10"),
        comment(14, "GoosLab", "moai-claim-release: comment=999 by=leader"),
        comment(15, "GoosLab", "just a note"),
        claim(100, "a", LIVE),
    ];
    let v = verdict(&c, "a", None);
    assert_eq!(v["status"], json!("owner"));
    let ids: Vec<i64> = v["invalid"]
        .as_array()
        .unwrap()
        .iter()
        .map(|i| {
            assert_eq!(i["reason"], json!("malformed"));
            i["comment_id"].as_i64().unwrap()
        })
        .collect();
    assert_eq!(ids, [10, 11, 12, 13, 14]);
    assert_eq!(v["claim_lines"], json!(6));
}

/// Where every claim is live and valid, the holder equals RC1 `resolve_winner`.
#[test]
fn agrees_with_resolve_winner_when_all_claims_are_live() {
    let sets: [&[i64]; 4] = [&[5], &[9, 3], &[300, 200, 100], &[42, 7, 99, 8]];
    for ids in sets {
        let comments: Vec<IssueComment> = ids
            .iter()
            .map(|&id| claim(id, &format!("s{id}"), LIVE))
            .collect();
        let rc1: Vec<ClaimComment> = ids
            .iter()
            .map(|&id| ClaimComment {
                comment_id: id,
                session: format!("s{id}"),
                epoch: 1,
                lease_until: LIVE.into(),
            })
            .collect();
        let winner = resolve_winner(&rc1).unwrap().comment_id;
        assert_eq!(verdict(&comments, "nobody", None)["holder"], json!(winner));
        let v = verdict(&comments, &format!("s{winner}"), None);
        assert_eq!(v["status"], json!("owner"));
    }
}

#[test]
fn commit_or_push_detection() {
    for yes in [
        "git commit -m x",
        "git -C repo push origin HEAD",
        "cargo test && git commit -F msg.txt",
        "/usr/bin/git push",
        "(cd x; git commit)",
        "GIT_EDITOR=true git commit --amend",
    ] {
        assert!(is_commit_or_push(yes), "{yes}");
    }
    // `git` only as an argument is not a git run (the #2270 audit's non-blocking (2)).
    for no in [
        "git status",
        "git log --grep=commit",
        "echo commit",
        "echo git commit",
        "grep -n 'git push' notes.md",
        "gh pr create",
        "",
    ] {
        assert!(!is_commit_or_push(no), "{no}");
    }
    assert!(!is_commit_or_push("git status && echo push"));
}

/// #2399 review (pr-2, PR comment 6770138445): only git's subcommand position counts. A read
/// command whose arguments contain the word is not a commit or push; global options that take
/// a value are skipped before the subcommand.
#[test]
fn commit_or_push_is_read_at_the_subcommand_position_only() {
    for no in [
        "git log --grep commit",
        "git log -- commit",
        "git show commit-ish",
        "git log --grep push",
        "git branch push",
        "git -C commit status",
        "git -c user.name=commit log",
    ] {
        assert!(!is_commit_or_push(no), "{no}");
    }
    for yes in [
        "git -C dir commit -m x",
        "git --no-pager commit -m x",
        "git --git-dir .git --work-tree . commit -m x",
        "git -c user.name=x -C dir push origin HEAD",
        "git -c alias.ci=commit ci -m x",
        "git -calias.ci=commit ci -m x",
    ] {
        assert!(is_commit_or_push(yes), "{yes}");
    }
}

/// #2399 re-audit B5: every git(1) global option that takes a separate value is skipped with
/// its value, and an option the list does not know makes both its next word and the following
/// subcommand candidates — a real commit or push is never missed behind it.
#[test]
fn global_options_with_values_and_unknown_options_never_hide_a_commit_or_push() {
    for yes in [
        "git --attr-source HEAD commit -m x",
        "git --attr-source HEAD push",
        "git --attr-source=HEAD commit -m x",
        "git --config-env k=E commit -m x",
        "git --namespace ns push",
        "git --list-cmds main commit -m x",
        "git --future-option val commit -m x",
        "git --future-option commit -m x",
        "git --future-option val --no-pager push",
    ] {
        assert!(is_commit_or_push(yes), "{yes}");
    }
    for no in [
        // The listed option's value is skipped, even a tree-ish spelled like a verb.
        "git --attr-source commit log",
        "git --namespace push status",
        "git --attr-source HEAD log --grep commit",
        "git --future-option log --grep commit",
        "git --future-option val log -- push",
    ] {
        assert!(!is_commit_or_push(no), "{no}");
    }
}

/// #2306: the bypass spellings the #2255 audit (#2270, comment 6738997664) found undetected —
/// a wrapper word before `git`, a command inside `sh -c`/`bash -c`, a `git -c alias.*` verb,
/// and a quoted or backslash-escaped `git`.
#[test]
fn commit_or_push_detection_sees_through_wrappers_and_quoting() {
    // The auditor's eight measured spellings, verbatim (all returned false at 4749bfe97).
    for yes in [
        "env git commit -m x",
        "command git commit -m x",
        "sh -c 'git commit -m x'",
        "nice git push",
        "git -c alias.ci=commit ci -m x",
        "\\git commit -m x",
        "\"git\" commit -m x",
        "xargs git push < /dev/null",
    ] {
        assert!(is_commit_or_push(yes), "{yes}");
    }
    for yes in [
        // Wrapper words (with their options and assignments) before `git`.
        "env -i git push",
        "time -p git commit -m x",
        "sudo -k git push",
        "env -i PATH=/usr/bin git push origin HEAD",
        "command git push origin HEAD",
        "nice -n 10 git commit -m x",
        "sudo -u lane git push",
        "time git commit -m x",
        "nohup git push origin HEAD",
        "timeout 30 git push",
        "echo HEAD | xargs git push origin",
        "xargs -I{} git commit -m {} < msgs",
        // A command line handed to a shell.
        "sh -c 'git push origin HEAD'",
        "bash -lc \"cargo test && git commit -m x\"",
        "/bin/sh -c 'env git push'",
        // Value-taking shell options before `-c` (rc.2 A9 follow-up 2 #2323, AC-1).
        "bash -o pipefail -c 'git push'",
        "sh -O extglob -c 'git commit -m x'",
        "bash +o posix -c 'git push'",
        // Bundled short options ending in o/O take a value (rc.2 A9 follow-up 3 #2330, AC-1).
        "bash -eo pipefail -c 'git push'",
        "bash -euo pipefail -c 'git commit -m x'",
        // An alias defined on the command line for commit or push.
        "git -c alias.ci=commit ci -m x",
        "git -c 'alias.p=push --force' p",
        // A quoted or escaped `git`.
        "\\git commit -m x",
        "'git' push origin HEAD",
        "\"git\" commit -m x",
        "g\\it push",
    ] {
        assert!(is_commit_or_push(yes), "{yes}");
    }
    // Mentions inside strings and wrappers running something else stay quiet.
    for no in [
        "sh -c 'echo git commit'",
        "bash -c \"git status\"",
        // Value-taking options reach the script now — non-git scripts stay quiet (AC-3).
        "bash -o pipefail -c 'cargo test'",
        "bash -o pipefail -c 'git status'",
        // Bundled options too: non-git scripts and a script file stay quiet (follow-up 3).
        "bash -eo pipefail -c 'cargo test'",
        "bash -eo pipefail script.sh",
        "env FOO=1 echo git push",
        "command -v git",
        "xargs echo git push",
        "git -c alias.st=status st",
        "echo '\\git commit'",
        "nice -n 5 cargo test",
        "grep -n 'a; git push' notes.md",
    ] {
        assert!(!is_commit_or_push(no), "{no}");
    }
    // Documented limit (#2306 scope): a shell's `-c` script is read one level deep only.
    assert!(!is_commit_or_push("sh -c 'sh -c \"git push\"'"));
    // Documented limit (rc.2 A9 follow-up 2 #2323): ANSI-C quoting `$'…'` is `$` expansion to
    // this text detector and stays invisible.
    assert!(!is_commit_or_push("$'git' push"));
    assert!(!is_commit_or_push("bash -c \"$'git' push\""));
}

#[test]
fn branch_remote_and_token_parsing() {
    use moai_cli::claim_verify::{claim_tokens, github_owner_repo, issue_from_branch};
    assert_eq!(
        issue_from_branch("issue/2255-rc2-a9-claim-enforce"),
        Some(2255)
    );
    assert_eq!(issue_from_branch("issue/7"), Some(7));
    for no in [
        "issue/x-1",
        "issue/7x",
        "issue/",
        "issue/0-a",
        "feature/7-a",
        "main",
    ] {
        assert_eq!(issue_from_branch(no), None, "{no}");
    }
    let pair = Some(("modu-ai".to_string(), "moai-adk".to_string()));
    for url in [
        "https://github.com/modu-ai/moai-adk.git",
        "https://github.com/modu-ai/moai-adk",
        "git@github.com:modu-ai/moai-adk.git",
        "ssh://git@github.com/modu-ai/moai-adk.git\n",
    ] {
        assert_eq!(github_owner_repo(url), pair, "{url}");
    }
    for no in [
        "https://gitlab.com/a/b.git",
        "https://github.com/a",
        "https://github.com/a/b/c",
        "/srv/repo.git",
    ] {
        assert_eq!(github_owner_repo(no), None, "{no}");
    }
    assert_eq!(
        claim_tokens("git commit -m \"fix\n\nMoai-Claim: 7#100\nMoai-Claim: 7#250\""),
        vec![(7, 100), (7, 250)]
    );
    assert_eq!(claim_tokens("Moai-Claim: 7#"), vec![]);
    assert_eq!(claim_tokens("Moai-Claim: #5"), vec![]);
}

fn get(body: Value) -> Exchange {
    common::exchange("GET", URL, common::json_response(200, body))
}

fn raw(id: i64, login: &str, body: &str) -> Value {
    json!({"id": id, "user": {"login": login}, "body": body})
}

fn request(session: &str, comment_id: Option<i64>) -> VerifyRequest {
    VerifyRequest {
        schema_version: 1,
        owner: "acme".into(),
        repo: "widgets".into(),
        issue_number: 7,
        session: session.into(),
        comment_id,
        allowed_authors: vec!["GoosLab".into()],
    }
}

/// REQ-A9-01 through the op: one GET, the owner passes, a non-owner is refused (Conflict),
/// and nothing is written to GitHub.
#[test]
fn verify_op_refuses_non_owner_with_one_get() {
    let body = json!([
        raw(
            100,
            "GoosLab",
            &format!("moai-claim: session=a epoch=1 lease_until={LIVE}")
        ),
        raw(
            200,
            "GoosLab",
            &format!("moai-claim: session=b epoch=1 lease_until={LIVE}")
        ),
    ]);
    let clock = FixedClock(now());
    let t = RecordedTransport::new(vec![get(body.clone())]);
    let ok = verify_cli(&request("a", Some(100)), &t, &clock).unwrap();
    assert_eq!(ok["status"], json!("owner"));
    assert!(t.sent.borrow().iter().all(|r| r.method == "GET"));

    let t = RecordedTransport::new(vec![get(body)]);
    let err = verify_cli(&request("b", Some(200)), &t, &clock).unwrap_err();
    assert_eq!(classify(&err), Kind::Conflict, "{err}");
    assert_eq!(t.sent.borrow().len(), 1);

    let mut bad = request("a", None);
    bad.owner = "acme/../x".into();
    let t = RecordedTransport::new(vec![]);
    let err = verify_cli(&bad, &t, &clock).unwrap_err();
    assert_eq!(classify(&err), Kind::InvalidInput);
    assert!(t.sent.borrow().is_empty());
}

fn git(dir: &Path, args: &[&str]) -> String {
    let full = [
        ["-c", "user.name=t", "-c", "user.email=t@example.invalid"].as_slice(),
        args,
    ]
    .concat();
    common::git::git(dir, &full)
}

/// A temporary repository with one active claim record whose worktree is the repo itself.
fn claimed_repo(comment_id: i64, session: &str) -> tempfile::TempDir {
    let dir = tempfile::TempDir::new().unwrap();
    git(dir.path(), &["init", "-q", "-b", "main"]);
    git(dir.path(), &["commit", "-q", "--allow-empty", "-m", "base"]);
    let base = git(dir.path(), &["rev-parse", "HEAD"]);
    let root = fs::canonicalize(dir.path()).unwrap();
    let record = ClaimRecord {
        schema_version: 1,
        task_ref: TaskRef::new("github", "github.com", "acme", "widgets", 1, 70, 7),
        task_key: "0123456789abcdef".into(),
        logical_key: "A9".into(),
        owner: Owner::new("session", session),
        assignee: None,
        provider_session: None,
        integration_owner: None,
        attempt: None,
        epoch: 1,
        base_revision: base,
        worktree: WorktreeId::new(
            root.join(".git").display().to_string(),
            root.join(".git").display().to_string(),
            root.display().to_string(),
            "b".repeat(40),
        ),
        write_scope: vec!["src/x.rs".into()],
        spec_ref: SpecRef {
            id: "SPEC-A9".into(),
            revision: "r1".into(),
            digest: None,
        },
        comment_id: Some(comment_id),
        claim_token: Some(format!("Moai-Claim: 7#{comment_id}")),
        lease_until: Some(LIVE.into()),
        state: "assigned".into(),
        effects: vec![],
        last_observed_at: NOW.into(),
    };
    let journal = Journal::open(&moai_cli::claim::journal_root(&root).unwrap()).unwrap();
    assert!(journal.try_lock().unwrap());
    journal.commit("acquire", &record).unwrap();
    dir
}

fn hook_input(command: &str, cwd: &Path) -> Vec<u8> {
    json!({"hook_event_name": "PreToolUse", "tool_name": "Bash",
           "tool_input": {"command": command}, "cwd": cwd})
    .to_string()
    .into_bytes()
}

/// AC-2 (library path): the guard lets everything through except a commit/push inside a
/// claimed worktree, which it verifies against the Issue.
#[test]
fn hook_guards_commit_in_claimed_worktree_only() {
    let clock = FixedClock(now());
    let foreign_lower = json!([
        raw(
            100,
            "GoosLab",
            &format!("moai-claim: session=a epoch=1 lease_until={LIVE}")
        ),
        raw(
            200,
            "GoosLab",
            &format!("moai-claim: session=b epoch=1 lease_until={LIVE}")
        ),
    ]);
    let repo = claimed_repo(200, "b");
    let root = fs::canonicalize(repo.path()).unwrap();

    // Not a commit: allowed, no GitHub read.
    let t = RecordedTransport::new(vec![]);
    let v = hook_cli(&root, &hook_input("cargo test", &root), &t, &clock).unwrap();
    assert_eq!(v["allowed"], json!(true));
    assert!(t.sent.borrow().is_empty());
    let v = hook_cli(&root, br#"{"tool_name":"Edit"}"#, &t, &clock).unwrap();
    assert_eq!(v["allowed"], json!(true));

    // Commit in the claimed worktree while a lower claim holds: refused.
    let t = RecordedTransport::new(vec![get(foreign_lower.clone())]);
    let err = hook_cli(&root, &hook_input("git commit -m x", &root), &t, &clock).unwrap_err();
    assert_eq!(classify(&err), Kind::Conflict, "{err}");

    // The owner's commit passes.
    let owner = claimed_repo(100, "a");
    let oroot = fs::canonicalize(owner.path()).unwrap();
    let t = RecordedTransport::new(vec![get(foreign_lower)]);
    let v = hook_cli(&oroot, &hook_input("git push", &oroot), &t, &clock).unwrap();
    assert_eq!(v["status"], json!("owner"));

    // A commit in a tree with neither a claim record nor an issue/<n> branch is not guarded.
    let plain = tempfile::TempDir::new().unwrap();
    git(plain.path(), &["init", "-q", "-b", "main"]);
    let proot = fs::canonicalize(plain.path()).unwrap();
    let t = RecordedTransport::new(vec![]);
    let v = hook_cli(&proot, &hook_input("git commit", &proot), &t, &clock).unwrap();
    assert_eq!(
        v["reason"],
        json!("no claim record and the branch is not issue/<n>; not a lane worktree")
    );
    assert!(t.sent.borrow().is_empty());
}

/// #2336: a lane branch carries its Issue number first (`issue/<n>-<key>`), the shape every
/// lane uses and `worktree-plan` now emits. The two card cases — the old `worktree-plan`
/// shape `issue/<key>-<n>-e<epoch>` with a digit-led key, and a letter-led key — are reported
/// as unparseable instead of reading a wrong Issue (6) or passing as "not a lane branch".
#[test]
fn lane_branch_reads_issue_first_and_reports_unparseable() {
    use moai_cli::claim_verify::{issue_from_branch, lane_branch, LaneBranch};
    assert_eq!(issue_from_branch("issue/06-C-2007-e1"), None);
    assert_eq!(lane_branch("issue/06-C-2007-e1"), LaneBranch::Unparseable);
    assert_eq!(lane_branch("issue/A9-x-2306-e1"), LaneBranch::Unparseable);
    assert_eq!(lane_branch("issue/2007-06-C-e1"), LaneBranch::Issue(2007));
    assert_eq!(lane_branch("issue/2306-A9-x-e1"), LaneBranch::Issue(2306));
    assert_eq!(
        lane_branch("issue/2255-rc2-a9-claim-enforce"),
        LaneBranch::Issue(2255)
    );
    assert_eq!(lane_branch("issue/7"), LaneBranch::Issue(7));
    for unparseable in ["issue/", "issue/x-1", "issue/7x", "issue/0-a", "issue/06-C"] {
        assert_eq!(
            lane_branch(unparseable),
            LaneBranch::Unparseable,
            "{unparseable}"
        );
    }
    for not_lane in ["main", "feature/7-a", "lane/2007-06-C-e1", "issues/7-a"] {
        assert_eq!(lane_branch(not_lane), LaneBranch::NotLane, "{not_lane}");
    }
}

/// #2336: with no claim record, a commit on a lane-looking branch whose Issue number cannot be
/// read is still allowed (fail-open stays: decision D5 HOLD), but the verdict says so in a
/// field of its own instead of the plain "not a lane worktree" reason. No GitHub read.
#[test]
fn hook_reports_unparseable_lane_branch_without_blocking() {
    let clock = FixedClock(now());
    let (_dir, root) = lane_repo("issue/A9-x-2306-e1");
    let t = RecordedTransport::new(vec![]);
    let v = hook_cli(&root, &hook_input("git commit -m x", &root), &t, &clock).unwrap();
    assert_eq!(v["allowed"], json!(true));
    assert_eq!(v["branch"], json!("unparseable"));
    let reason = v["reason"].as_str().unwrap();
    assert!(reason.contains("issue/A9-x-2306-e1"), "{reason}");
    assert!(t.sent.borrow().is_empty());
}

fn gh_stub(dir: &Path, json_body: &str) -> std::path::PathBuf {
    let path = dir.join("gh-stub");
    fs::write(
        &path,
        "#!/bin/sh\nprintf 'HTTP/2.0 200 OK\\r\\n\\r\\n%s' \"$STUB_JSON\"\n",
    )
    .unwrap();
    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt;
        fs::set_permissions(&path, fs::Permissions::from_mode(0o755)).unwrap();
    }
    let _ = json_body;
    path
}

/// The binary: `claim-verify --hook` exits 2 on a refused commit (the code a host hook
/// needs to block) and 0 when the commit is allowed.
#[cfg(unix)]
#[test]
fn hook_binary_exits_2_on_refusal() {
    let lease = moai_cli::forge::policy::format_utc(
        std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap()
            .as_secs() as i64
            + 3600,
    );
    let repo = claimed_repo(200, "b");
    let root = fs::canonicalize(repo.path()).unwrap();
    let stub_dir = tempfile::TempDir::new().unwrap();
    let body = json!([
        raw(
            100,
            "GoosLab",
            &format!("moai-claim: session=a epoch=1 lease_until={lease}")
        ),
        raw(
            200,
            "GoosLab",
            &format!("moai-claim: session=b epoch=1 lease_until={lease}")
        ),
    ])
    .to_string();
    let gh = gh_stub(stub_dir.path(), &body);
    let run = |command: &str| {
        let mut child = Command::new(env!("CARGO_BIN_EXE_moai-cli"))
            .args(["claim-verify", root.to_str().unwrap(), "--hook"])
            .current_dir(&root)
            .env("MOAI_FORGE_GH_BIN", &gh)
            .env("STUB_JSON", &body)
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_SYSTEM", "/dev/null")
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
            .unwrap();
        common::feed_stdin(&mut child, &hook_input(command, &root));
        child.wait_with_output().unwrap()
    };
    let refused = run("git commit -m x");
    assert_eq!(
        refused.status.code(),
        Some(2),
        "{}",
        String::from_utf8_lossy(&refused.stdout)
    );
    assert!(String::from_utf8_lossy(&refused.stderr).contains("claim verify: not owner"));
    let allowed = run("cargo test");
    assert_eq!(allowed.status.code(), Some(0));
}

/// Before a push, every lane-written commit on the branch's own first-parent line since the
/// claim base must carry the claim token. Merge commits are not checked (decision 2026-11-09
/// 06:02, #1975: the claim line is required on lane commits, a merge commit is only a record;
/// this reverses the earlier #2226 follow-up that checked merges), and neither are the commits
/// of other cards a merge brings in.
#[test]
fn hook_refuses_push_when_a_branch_commit_lacks_the_trailer() {
    let clock = FixedClock(now());
    let owner_only = json!([raw(
        100,
        "GoosLab",
        &format!("moai-claim: session=a epoch=1 lease_until={LIVE}")
    )]);
    let repo = claimed_repo(100, "a");
    let root = fs::canonicalize(repo.path()).unwrap();
    let token = "Moai-Claim: 7#100";
    git(
        &root,
        &[
            "commit",
            "-q",
            "--allow-empty",
            "-m",
            &format!("work\n\n{token}"),
        ],
    );
    // Another card's commit on a side branch, merged in without the trailer.
    git(&root, &["checkout", "-q", "-b", "tip"]);
    git(
        &root,
        &[
            "commit",
            "-q",
            "--allow-empty",
            "-m",
            "other card (no token)",
        ],
    );
    git(&root, &["checkout", "-q", "main"]);

    let push = || {
        let t = RecordedTransport::new(vec![get(owner_only.clone())]);
        hook_cli(&root, &hook_input("git push", &root), &t, &clock)
    };
    let v = push().unwrap();
    assert_eq!(v["trailer_checked"], json!(1));

    // A tip merge without the token passes: neither it nor the merged-in commit is checked.
    git(&root, &["merge", "-q", "--no-ff", "-m", "Merge tip", "tip"]);
    let v = push().unwrap();
    assert_eq!(v["trailer_checked"], json!(1));

    // A lane commit without the token refuses the push.
    git(
        &root,
        &["commit", "-q", "--allow-empty", "-m", "forgot the trailer"],
    );
    let err = push().unwrap_err();
    assert!(err.starts_with("claim verify: trailer missing"), "{err}");
    assert_eq!(classify(&err), Kind::InvalidInput);

    // A commit (not a push) is not trailer-checked here.
    let t = RecordedTransport::new(vec![get(owner_only.clone())]);
    let v = hook_cli(&root, &hook_input("git commit -m x", &root), &t, &clock).unwrap();
    assert!(v.get("trailer_checked").is_none());
}

/// After a handoff, commits written under the previous epoch's token stay valid: the push
/// accepts the own claim's token and every earlier claim closed by a handoff line (the #2270
/// audit's non-blocking (1): a new epoch's base no longer has to be the hand-over tip).
#[test]
fn push_accepts_tokens_of_claims_closed_by_handoff() {
    let clock = FixedClock(now());
    let repo = claimed_repo(300, "b");
    let root = fs::canonicalize(repo.path()).unwrap();
    git(
        &root,
        &[
            "commit",
            "-q",
            "--allow-empty",
            "-m",
            "epoch 1\n\nMoai-Claim: 7#100",
        ],
    );
    git(
        &root,
        &[
            "commit",
            "-q",
            "--allow-empty",
            "-m",
            "epoch 2\n\nMoai-Claim: 7#300",
        ],
    );
    let claims = |with_handoff: bool| {
        let mut c = vec![raw(
            100,
            "GoosLab",
            &format!("moai-claim: session=a epoch=1 lease_until={LIVE}"),
        )];
        if with_handoff {
            c.push(raw(
                250,
                "GoosLab",
                "moai-claim-handoff: comment=100 to=b by=leader",
            ));
        }
        c.push(raw(
            300,
            "GoosLab",
            &format!("moai-claim: session=b epoch=2 lease_until={LIVE}"),
        ));
        Value::Array(c)
    };
    let t = RecordedTransport::new(vec![get(claims(true))]);
    let v = hook_cli(&root, &hook_input("git push", &root), &t, &clock).unwrap();
    assert_eq!(v["status"], json!("owner"));
    assert_eq!(v["trailer_checked"], json!(2));

    // Without the handoff line the epoch-1 claim still holds: epoch 2 is not the owner.
    let t = RecordedTransport::new(vec![get(claims(false))]);
    let err = hook_cli(&root, &hook_input("git push", &root), &t, &clock).unwrap_err();
    assert_eq!(classify(&err), Kind::Conflict, "{err}");

    // Decision D1: the same session's epoch 1 renewed by its epoch 2 line, no handoff line;
    // the epoch-1 commit's token is still accepted on push.
    let renewed = json!([
        raw(
            100,
            "GoosLab",
            &format!("moai-claim: session=b epoch=1 lease_until={EXPIRED}"),
        ),
        raw(
            300,
            "GoosLab",
            &format!("moai-claim: session=b epoch=2 lease_until={LIVE}"),
        ),
    ]);
    let t = RecordedTransport::new(vec![get(renewed)]);
    let v = hook_cli(&root, &hook_input("git push", &root), &t, &clock).unwrap();
    assert_eq!(v["status"], json!("owner"));
    assert_eq!(v["trailer_checked"], json!(2));
}

/// A lane worktree with no claim record: branch `issue/7-x`, origin on GitHub, the default
/// branch at `refs/remotes/origin/HEAD`. No journal is written, as when the leader posts
/// the claim for the lane.
fn lane_repo(branch: &str) -> (tempfile::TempDir, std::path::PathBuf) {
    let dir = tempfile::TempDir::new().unwrap();
    let root = fs::canonicalize(dir.path()).unwrap();
    git(&root, &["init", "-q", "-b", "main"]);
    git(&root, &["commit", "-q", "--allow-empty", "-m", "base"]);
    git(
        &root,
        &[
            "remote",
            "add",
            "origin",
            "https://github.com/acme/widgets.git",
        ],
    );
    git(&root, &["update-ref", "refs/remotes/origin/main", "HEAD"]);
    git(
        &root,
        &[
            "symbolic-ref",
            "refs/remotes/origin/HEAD",
            "refs/remotes/origin/main",
        ],
    );
    git(&root, &["checkout", "-q", "-b", branch]);
    (dir, root)
}

fn owner_100() -> Value {
    json!([raw(
        100,
        "GoosLab",
        &format!("moai-claim: session=agent-1-abc epoch=1 lease_until={LIVE}")
    )])
}

/// AC-2 (operating path, #2270 audit FAIL): with no claim record the guard still verifies a
/// commit in an `issue/<n>` worktree, from the branch's Issue and the `Moai-Claim` token in
/// the commit message; without a usable token it refuses before any GitHub read
/// (fail-closed, with the next action in the message).
#[test]
fn hook_guards_a_lane_branch_without_a_claim_record() {
    let clock = FixedClock(now());
    let (_d, root) = lane_repo("issue/7-widget");

    // The owner's commit, token in the message: verified against the Issue, allowed.
    let t = RecordedTransport::new(vec![get(owner_100())]);
    let cmd = "git commit -m \"work\n\nMoai-Claim: 7#100\"";
    let v = hook_cli(&root, &hook_input(cmd, &root), &t, &clock).unwrap();
    assert_eq!(v["status"], json!("owner"));
    assert_eq!(v["source"], json!("branch"));
    assert_eq!(t.sent.borrow().len(), 1);
    assert_eq!(t.sent.borrow()[0].url, URL);

    // A lower live claim holds: the same commit is refused (Conflict).
    let lower = json!([
        raw(
            50,
            "GoosLab",
            &format!("moai-claim: session=x epoch=1 lease_until={LIVE}")
        ),
        raw(
            100,
            "GoosLab",
            &format!("moai-claim: session=y epoch=1 lease_until={LIVE}")
        ),
    ]);
    let t = RecordedTransport::new(vec![get(lower)]);
    let err = hook_cli(&root, &hook_input(cmd, &root), &t, &clock).unwrap_err();
    assert_eq!(classify(&err), Kind::Conflict, "{err}");

    // No token (the first commit before the claim id arrived): uncertain, no GitHub read.
    let t = RecordedTransport::new(vec![]);
    let err = hook_cli(&root, &hook_input("git commit -m work", &root), &t, &clock).unwrap_err();
    assert!(err.starts_with("claim verify: uncertain;"), "{err}");
    assert!(
        err.contains("ask the leader for the claim comment id"),
        "{err}"
    );
    assert!(err.contains("`Moai-Claim: 7#<id>`"), "{err}");
    assert_eq!(classify(&err), Kind::Blocked);
    assert!(t.sent.borrow().is_empty());

    // A token for another Issue only: uncertain.
    let err = hook_cli(
        &root,
        &hook_input("git commit -m \"x\n\nMoai-Claim: 8#100\"", &root),
        &t,
        &clock,
    )
    .unwrap_err();
    assert!(err.contains("names another Issue"), "{err}");
    assert_eq!(classify(&err), Kind::Blocked);

    // A message from -F is not read: uncertain, not allowed.
    let err = hook_cli(
        &root,
        &hook_input("git commit -F msg.txt", &root),
        &t,
        &clock,
    )
    .unwrap_err();
    assert_eq!(classify(&err), Kind::Blocked, "{err}");

    // An origin that is not GitHub: uncertain.
    git(&root, &["remote", "set-url", "origin", "/srv/repo.git"]);
    let err = hook_cli(&root, &hook_input(cmd, &root), &t, &clock).unwrap_err();
    assert!(err.contains("not a GitHub repository"), "{err}");
    assert!(t.sent.borrow().is_empty());

    // Other branches and a detached HEAD are not lane worktrees: allowed, reason given.
    let (_e, other) = lane_repo("leader/docs");
    let v = hook_cli(&other, &hook_input("git commit -m x", &other), &t, &clock).unwrap();
    assert_eq!(v["allowed"], json!(true));
    assert!(v["reason"].as_str().unwrap().contains("not issue/<n>"));
    git(&other, &["checkout", "-q", "--detach"]);
    let v = hook_cli(&other, &hook_input("git commit -m x", &other), &t, &clock).unwrap();
    assert!(v["reason"].as_str().unwrap().contains("detached HEAD"));
    assert!(t.sent.borrow().is_empty());
}

/// The push half of the operating path: the comment id comes from the newest own commit's
/// token, the range is the commits on no other origin branch, tip merges are not checked.
#[test]
fn hook_guards_a_lane_push_without_a_claim_record() {
    let clock = FixedClock(now());
    let (_d, root) = lane_repo("issue/7-widget");
    let push = || {
        let t = RecordedTransport::new(vec![get(owner_100())]);
        hook_cli(
            &root,
            &hook_input("git push -u origin HEAD", &root),
            &t,
            &clock,
        )
    };

    // No commit carries a token yet: uncertain.
    let err = push().unwrap_err();
    assert!(err.contains("no commit on the branch carries"), "{err}");
    assert_eq!(classify(&err), Kind::Blocked);

    git(
        &root,
        &[
            "commit",
            "-q",
            "--allow-empty",
            "-m",
            "work\n\nMoai-Claim: 7#100",
        ],
    );
    // A tip merge (the default branch moved on) without the token, as decision 06:02 allows.
    git(&root, &["checkout", "-q", "main"]);
    git(
        &root,
        &["commit", "-q", "--allow-empty", "-m", "other card"],
    );
    git(&root, &["update-ref", "refs/remotes/origin/main", "HEAD"]);
    git(&root, &["checkout", "-q", "issue/7-widget"]);
    git(
        &root,
        &["merge", "-q", "--no-ff", "-m", "Merge main", "main"],
    );
    let v = push().unwrap();
    assert_eq!(v["status"], json!("owner"));
    assert_eq!(v["trailer_checked"], json!(1));

    // A lane commit without the token: the push is refused.
    git(
        &root,
        &["commit", "-q", "--allow-empty", "-m", "no trailer"],
    );
    let err = push().unwrap_err();
    assert!(err.starts_with("claim verify: trailer missing"), "{err}");

    git(
        &root,
        &[
            "commit",
            "-q",
            "--amend",
            "--allow-empty",
            "-m",
            "fixed\n\nMoai-Claim: 7#100",
        ],
    );
    // Two lane commits now ("work" and "fixed"); the tip merge between them is not checked.
    assert_eq!(push().unwrap()["trailer_checked"], json!(2));

    // The branch's own pushed commits are still checked: excluding origin/<own branch>
    // from "other remote branches" keeps them in range.
    git(
        &root,
        &["update-ref", "refs/remotes/origin/issue/7-widget", "HEAD"],
    );
    assert_eq!(push().unwrap()["trailer_checked"], json!(2));

    // With no other remote branch, every commit is checked: the base commit lacks the
    // token, so the push is refused rather than passed unchecked.
    git(
        &root,
        &["symbolic-ref", "--delete", "refs/remotes/origin/HEAD"],
    );
    git(&root, &["update-ref", "-d", "refs/remotes/origin/main"]);
    let err = push().unwrap_err();
    assert!(err.starts_with("claim verify: trailer missing"), "{err}");
}

/// After a renewal (handoff + new epoch) the new token may sit only on a merge commit: the
/// lane's last own commit still carries the handed-off epoch's token. The push must verify
/// the newest claim (highest comment id among the branch's tokens), not the closed one —
/// found by running the guard on #2270's own push (it refused the owner as `closed`).
#[test]
fn lane_push_after_renewal_verifies_the_newest_epoch() {
    let clock = FixedClock(now());
    let (_d, root) = lane_repo("issue/7-widget");
    git(
        &root,
        &[
            "commit",
            "-q",
            "--allow-empty",
            "-m",
            "epoch 1\n\nMoai-Claim: 7#100",
        ],
    );
    git(&root, &["checkout", "-q", "main"]);
    git(
        &root,
        &["commit", "-q", "--allow-empty", "-m", "other card"],
    );
    git(&root, &["update-ref", "refs/remotes/origin/main", "HEAD"]);
    git(&root, &["checkout", "-q", "issue/7-widget"]);
    git(
        &root,
        &[
            "merge",
            "-q",
            "--no-ff",
            "-m",
            "tip sync\n\nMoai-Claim: 7#200",
            "main",
        ],
    );
    let renewed = json!([
        raw(
            100,
            "GoosLab",
            &format!("moai-claim: session=a epoch=1 lease_until={EXPIRED}")
        ),
        raw(
            150,
            "GoosLab",
            "moai-claim-handoff: comment=100 to=a2 by=leader"
        ),
        raw(
            200,
            "GoosLab",
            &format!("moai-claim: session=a2 epoch=2 lease_until={LIVE}")
        ),
    ]);
    let t = RecordedTransport::new(vec![get(renewed)]);
    let v = hook_cli(&root, &hook_input("git push", &root), &t, &clock).unwrap();
    assert_eq!(v["status"], json!("owner"));
    assert_eq!(v["mine"], json!(200));
    // The epoch-1 lane commit is checked and accepted: its claim is handoff-closed.
    assert_eq!(v["trailer_checked"], json!(1));
}

/// This repository's shape: the integration branch shares no history with the default
/// branch `origin/HEAD` points to, so no merge-base with it exists. The push check still
/// scopes to the lane's own commits (those on no other origin branch).
#[test]
fn lane_push_works_when_the_integration_branch_is_unrelated_to_the_default_branch() {
    let clock = FixedClock(now());
    let (_d, root) = lane_repo("issue/7-widget");
    git(&root, &["checkout", "-q", "--orphan", "integration"]);
    git(
        &root,
        &["commit", "-q", "--allow-empty", "-m", "integration root"],
    );
    git(
        &root,
        &["update-ref", "refs/remotes/origin/integration", "HEAD"],
    );
    git(&root, &["branch", "-q", "-D", "issue/7-widget"]);
    git(&root, &["checkout", "-q", "-b", "issue/7-widget"]);
    git(
        &root,
        &[
            "commit",
            "-q",
            "--allow-empty",
            "-m",
            "work\n\nMoai-Claim: 7#100",
        ],
    );
    let t = RecordedTransport::new(vec![get(owner_100())]);
    let v = hook_cli(&root, &hook_input("git push", &root), &t, &clock).unwrap();
    assert_eq!(v["status"], json!("owner"));
    assert_eq!(v["trailer_checked"], json!(1));
}

/// The binary on the operating path: a tokenless commit in an `issue/<n>` worktree without a
/// claim record exits 2 (the host blocks it); a non-commit exits 0.
#[cfg(unix)]
#[test]
fn hook_binary_exits_2_on_a_lane_branch_without_a_token() {
    let (_d, root) = lane_repo("issue/7-widget");
    let run = |command: &str| {
        let mut child = Command::new(env!("CARGO_BIN_EXE_moai-cli"))
            .args(["claim-verify", ".", "--hook"])
            .current_dir(&root)
            // The refusal comes before any GitHub read, so no `gh` stub is needed.
            .env_remove("MOAI_FORGE_GH_BIN")
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_SYSTEM", "/dev/null")
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
            .unwrap();
        common::feed_stdin(&mut child, &hook_input(command, &root));
        child.wait_with_output().unwrap()
    };
    let refused = run("git commit -m work");
    assert_eq!(refused.status.code(), Some(2));
    assert!(String::from_utf8_lossy(&refused.stderr).contains("claim verify: uncertain"));
    assert_eq!(run("cargo test").status.code(), Some(0));
}

/// AC-2 (workflow text): the MODE-AGENT workflow names the boundary before commit and
/// push, and the guard that enforces it; the binary test above is the mechanical half.
#[test]
fn mode_agent_workflow_names_claim_verify_before_writes() {
    let text = fs::read_to_string(
        Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai/workflows/mode-agent.md"),
    )
    .unwrap();
    let step = text
        .lines()
        .find(|l| l.contains("moai claim-verify"))
        .expect("mode-agent names claim-verify");
    assert!(
        step.contains("Before every commit and every push"),
        "{step}"
    );
    assert!(step.contains("claim-verify --hook"), "{step}");
    assert!(
        step.contains("never take the Issue over by time alone"),
        "{step}"
    );
}

/// #2673 AC-3: `claim-verify --hook` reads the branch, the claim record and the commits in the
/// folder `git -C` names. Before, a commit sent from the primary checkout with
/// `git -C <lane worktree>` read the primary's branch and passed as "not a lane worktree".
#[test]
fn hook_reads_the_folder_git_dash_c_names() {
    let clock = FixedClock(now());
    let (_d, root) = lane_repo("leader/docs");
    git(
        &root,
        &["worktree", "add", "-q", "-b", "issue/7-widget", "wt7"],
    );

    // No token, commit in the lane worktree from the primary checkout: uncertain, refused
    // before any GitHub read.
    let t = RecordedTransport::new(vec![]);
    let err = hook_cli(
        &root,
        &hook_input("git -C wt7 commit -m work", &root),
        &t,
        &clock,
    )
    .unwrap_err();
    assert!(err.contains("`Moai-Claim: 7#<id>`"), "{err}");
    assert_eq!(classify(&err), Kind::Blocked);
    assert!(t.sent.borrow().is_empty());

    // With the token: verified against Issue 7, the lane worktree's branch.
    let t = RecordedTransport::new(vec![get(owner_100())]);
    let cmd = "git -C wt7 commit -m \"work\n\nMoai-Claim: 7#100\"";
    let v = hook_cli(&root, &hook_input(cmd, &root), &t, &clock).unwrap();
    assert_eq!(v["status"], json!("owner"));
    assert_eq!(v["source"], json!("branch"));
    assert_eq!(v["issue"], json!(7));

    // Two folders on one line: uncertain, nothing read.
    let t = RecordedTransport::new(vec![]);
    let err = hook_cli(
        &root,
        &hook_input("git -C wt7 commit -m a && git commit -m b", &root),
        &t,
        &clock,
    )
    .unwrap_err();
    assert!(err.contains("more than one folder"), "{err}");
    assert_eq!(classify(&err), Kind::Blocked);
    assert!(t.sent.borrow().is_empty());

    // From the lane worktree, `-C ..` names the primary checkout: not a lane branch, allowed.
    let wt = root.join("wt7");
    let v = hook_cli(&root, &hook_input("git -C .. commit -m x", &wt), &t, &clock).unwrap();
    assert_eq!(v["allowed"], json!(true));
    assert!(
        v["reason"].as_str().unwrap().contains("not issue/<n>"),
        "{v}"
    );
}
