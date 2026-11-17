//! The `audit-gate` CI job's decision (SPEC-20261117-audit-gate, CI redesign D2): whether a pull
//! request's head has the audit its tier needs, read from the PR's verdict comments. It reuses
//! the merge gate's readers (`gate::verdict_line`, `gate::parse_verdict`, `gate::auditor`,
//! `gate::parse_body_tier`, `gate::tier_of`) and decides from GitHub data alone — no git, no tip
//! CI — so a job can run it on any runner. The leader's full merge gate (`gate-check`) is
//! unchanged.
//!
//! The rules, in order:
//! 1. Only a repository member's comment counts (`comment_authority`, Issue #3052).
//! 2. A posted FAIL stops every tier until the same auditor posts a later PASS. A first line that
//!    starts with `AUDIT #` but does not read as a verdict counts as a FAIL (it may be one).
//! 3. Tier 0: no open FAIL is enough.
//! 4. Tier 1 and 2: a PASS on exactly this head, with a full 40-hex base, by an auditor lane that
//!    is not an author. Authors are the `Moai-Session:` trailer values of the PR's commits
//!    (decision U-B, 2026-11-17); when no commit carries one, the authors are unknown and the
//!    head has no qualifying PASS (fail closed, like G9).
//!
//! Not checked in this slice: that the PASS cites the run (AGENTS.md workflow 7) — a later slice
//! adds it; the verdict base's ancestry and the mechanical delta (both need git).

use crate::forge::transport::Transport;
use crate::gate::{self, Comment};
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};

type Result<T> = std::result::Result<T, String>;

/// Everything the decision reads: the PR's head, body, files, commit messages and comments.
#[derive(Clone, Debug, Deserialize, Serialize)]
#[serde(deny_unknown_fields)]
pub struct Input {
    pub schema_version: u32,
    pub pr: u64,
    /// The head the check is for, 40 lowercase hex.
    pub head: String,
    pub body: String,
    /// The PR's changed paths; `None` when they could not be read (tier 2 then).
    #[serde(default)]
    pub files: Option<Vec<String>>,
    /// The PR's commit messages; `None` when they could not all be read (authors unknown).
    #[serde(default)]
    pub commit_messages: Option<Vec<String>>,
    pub comments: Vec<Comment>,
}

/// One verdict as the gate reads it.
#[derive(Clone, Debug, PartialEq)]
struct Verdict {
    id: u64,
    /// `None` for a first line that starts with `AUDIT #` but does not read.
    head: Option<String>,
    /// `PASS`, `FAIL`, or `unreadable`.
    word: String,
    /// The base reads as 40 hex (a short base is read only for its word).
    full_base: bool,
    auditor: Option<String>,
}

fn verdicts(comments: &[Comment]) -> Vec<Verdict> {
    let mut out: Vec<Verdict> = comments
        .iter()
        .filter_map(|c| {
            let line = gate::verdict_line(&c.body)?;
            let auditor = gate::auditor(&c.body);
            let (head, word, full_base) = match gate::parse_verdict(line) {
                Some((h, _, v)) => (Some(h), v, true),
                None => match gate::parse_verdict_short_base(line) {
                    Some((h, _, v)) => (Some(h), v, false),
                    None => (None, "unreadable".to_string(), false),
                },
            };
            Some(Verdict {
                id: c.id,
                head,
                word,
                full_base,
                auditor,
            })
        })
        .collect();
    out.sort_by_key(|v| v.id);
    out
}

/// The `Moai-Session:` values of one commit message's trailer block (its last paragraph).
pub fn sessions_of(message: &str) -> Vec<String> {
    let text = message.replace("\r\n", "\n");
    let last = text
        .trim_end()
        .rsplit("\n\n")
        .next()
        .unwrap_or_default()
        .to_string();
    last.lines()
        .filter_map(|l| l.strip_prefix("Moai-Session:"))
        .map(|v| v.trim().to_string())
        .filter(|v| !v.is_empty())
        .collect()
}

/// The decision: `pass`, `fail` (an open FAIL) or `missing` (no qualifying PASS yet), with the
/// tier it judged by, the rule that decided and why.
pub fn decide(input: &Input) -> Value {
    let out = |state: &str, tier: Option<u8>, rule: &str, reason: String| {
        json!({"state": state, "pr": input.pr, "head": input.head, "tier": tier,
               "rule": rule, "reason": reason, "cites_run_checked": false})
    };
    if input.schema_version != 1 || !gate::is_full_sha(&input.head) {
        return out(
            "fail",
            None,
            "input",
            "schema_version 1 and a 40-hex head are required".into(),
        );
    }
    let tier = gate::tier_of(
        gate::parse_body_tier(&input.body),
        input.files.as_deref().unwrap_or(&[]),
    );
    let tier = if input.files.is_none() { 2 } else { tier };
    let mut comments = input.comments.clone();
    crate::comment_authority::retain_members(&mut comments);
    let all = verdicts(&comments);

    // Rule 2: every FAIL (or unreadable verdict) needs a later PASS by the same auditor.
    let open: Vec<&Verdict> = all
        .iter()
        .filter(|f| f.word != "PASS")
        .filter(|f| {
            !all.iter().any(|p| {
                p.word == "PASS" && p.id > f.id && p.auditor.is_some() && p.auditor == f.auditor
            })
        })
        .collect();
    if let Some(f) = open.first() {
        let who = f.auditor.as_deref().unwrap_or("no auditor lane");
        return out(
            "fail",
            Some(tier),
            "open-fail",
            format!(
                "{} open FAIL verdict(s); comment {} ({}, {who}) has no later PASS by the same auditor",
                open.len(),
                f.id,
                f.word
            ),
        );
    }
    if tier == 0 {
        return out(
            "pass",
            Some(0),
            "tier-0",
            "audit tier 0 and no open FAIL".into(),
        );
    }

    // Rule 4: a non-author PASS on exactly this head.
    let authors: Vec<String> = input
        .commit_messages
        .iter()
        .flatten()
        .flat_map(|m| sessions_of(m))
        .collect();
    let on_head: Vec<&Verdict> = all
        .iter()
        .filter(|v| v.word == "PASS" && v.full_base && v.head.as_deref() == Some(&input.head))
        .collect();
    if on_head.is_empty() {
        return out(
            "missing",
            Some(tier),
            "pass-on-head",
            format!(
                "audit tier {tier}: no PASS on head {} with a 40-hex base",
                input.head
            ),
        );
    }
    if input.commit_messages.is_none() || authors.is_empty() {
        return out(
            "missing",
            Some(tier),
            "authors",
            "the PR's authors are unknown: no commit carries a `Moai-Session:` trailer".into(),
        );
    }
    let is_author = |who: &str| authors.iter().any(|a| a.eq_ignore_ascii_case(who));
    match on_head
        .iter()
        .rev()
        .find(|v| v.auditor.as_deref().is_some_and(|w| !is_author(w)))
    {
        Some(v) => out(
            "pass",
            Some(tier),
            "non-author-pass",
            format!(
                "audit tier {tier}: PASS on this head by {} (comment {}), not an author",
                v.auditor.as_deref().unwrap_or_default(),
                v.id
            ),
        ),
        None => out(
            "missing",
            Some(tier),
            "non-author-pass",
            format!(
                "audit tier {tier}: every PASS on this head names no auditor lane or one that wrote commits of this PR"
            ),
        ),
    }
}

/// Commits GitHub lists for one pull request at most; a longer PR's authors are unknown.
const MAX_LISTED_COMMITS: u64 = 250;

/// Reads one pull request into an [`Input`] (GET only). `head` is the commit the check is for;
/// `None` takes the PR's current head.
pub fn collect<T: Transport>(
    owner: &str,
    repo: &str,
    pr: u64,
    head: Option<&str>,
    t: &T,
) -> Result<Input> {
    use crate::gate_collect::{get_all, get_json};
    let base = format!("https://api.github.com/repos/{owner}/{repo}");
    let (p, _) = get_json(t, &format!("{base}/pulls/{pr}"))?;
    let current = p["head"]["sha"]
        .as_str()
        .ok_or("pull request has no head sha")?;
    let head = head.unwrap_or(current).to_string();
    let body = p["body"].as_str().unwrap_or("").to_string();
    let files = get_all(t, &format!("{base}/pulls/{pr}/files?per_page=100"), 30)
        .ok()
        .map(|list| {
            list.iter()
                .filter_map(|f| f["filename"].as_str().map(String::from))
                .collect()
        });
    let commit_messages = if p["commits"].as_u64().unwrap_or(u64::MAX) > MAX_LISTED_COMMITS {
        None
    } else {
        get_all(t, &format!("{base}/pulls/{pr}/commits?per_page=100"), 3)
            .ok()
            .map(|list| {
                list.iter()
                    .map(|c| c["commit"]["message"].as_str().unwrap_or("").to_string())
                    .collect()
            })
    };
    let comments = get_all(t, &format!("{base}/issues/{pr}/comments?per_page=100"), 30)?
        .iter()
        .map(|c| {
            Ok(Comment {
                id: c["id"].as_u64().ok_or("comment without id")?,
                body: c["body"].as_str().unwrap_or("").to_string(),
                author_association: crate::comment_authority::rest_association(c),
            })
        })
        .collect::<Result<Vec<_>>>()?;
    Ok(Input {
        schema_version: 1,
        pr,
        head,
        body,
        files,
        commit_messages,
        comments,
    })
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::forge::transport::{Exchange, HttpRequest, HttpResponse, RecordedTransport};

    const H: &str = "1111111111111111111111111111111111111111";
    const OLD: &str = "2222222222222222222222222222222222222222";
    const B: &str = "3333333333333333333333333333333333333333";

    fn verdict(id: u64, head: &str, word: &str, lane: &str) -> Comment {
        Comment {
            id,
            body: format!("AUDIT #9 (x) head {head} base {B} — verdict: {word}\nauditor: {lane}"),
            author_association: "MEMBER".into(),
        }
    }

    fn input(tier: &str, files: &[&str], sessions: &[&str], comments: Vec<Comment>) -> Input {
        Input {
            schema_version: 1,
            pr: 9,
            head: H.into(),
            body: format!("audit-tier: {tier}\nbody"),
            files: Some(files.iter().map(|s| s.to_string()).collect()),
            commit_messages: Some(
                sessions
                    .iter()
                    .map(|s| format!("feat: x\n\nbody\n\nMoai-Session: {s}\nCo-Authored-By: a <b>"))
                    .collect(),
            ),
            comments,
        }
    }

    fn state(i: &Input) -> (String, String) {
        let v = decide(i);
        (
            v["state"].as_str().unwrap().into(),
            v["rule"].as_str().unwrap().into(),
        )
    }

    #[test]
    fn trailers_are_read_from_the_last_paragraph_only() {
        assert_eq!(
            sessions_of("x\n\nMoai-Session: agent-3\nMoai-Spec: S\n"),
            vec!["agent-3"]
        );
        assert!(sessions_of("x\n\nMoai-Session: agent-3\n\nfooter line").is_empty());
        assert!(sessions_of("Moai-Session: \n").is_empty());
        assert_eq!(sessions_of("x\r\n\r\nMoai-Session: a\r\n"), vec!["a"]);
    }

    #[test]
    fn tier_0_passes_without_a_verdict_and_stops_on_an_open_fail() {
        let docs = ["docs/a.md"];
        assert_eq!(state(&input("0", &docs, &[], vec![])).0, "pass");
        let failed = vec![verdict(5, OLD, "FAIL", "agent-7")];
        assert_eq!(
            state(&input("0", &docs, &[], failed.clone())),
            ("fail".into(), "open-fail".into())
        );
        // A later PASS by another auditor does not clear it; one by the same auditor does.
        let mut other = failed.clone();
        other.push(verdict(6, H, "PASS", "agent-8"));
        assert_eq!(state(&input("0", &docs, &[], other)).0, "fail");
        let mut same = failed;
        same.push(verdict(6, H, "PASS", "agent-7"));
        assert_eq!(state(&input("0", &docs, &[], same)).0, "pass");
        // Audit B2: only a LATER PASS clears it; a PASS posted before the FAIL does not.
        let earlier = vec![
            verdict(200, H, "PASS", "agent-7"),
            verdict(300, H, "FAIL", "agent-7"),
        ];
        assert_eq!(
            state(&input("0", &docs, &[], earlier.clone())),
            ("fail".into(), "open-fail".into())
        );
        assert_eq!(
            state(&input("1", &["src/a.rs"], &["agent-1"], earlier)).0,
            "fail"
        );
        // A source path raises a declared tier 0 to 1: no verdict is then missing.
        assert_eq!(
            state(&input("0", &["src/a.rs"], &["agent-1"], vec![])),
            ("missing".into(), "pass-on-head".into())
        );
    }

    #[test]
    fn tier_1_needs_a_non_author_pass_on_this_head() {
        let src = ["src/a.rs"];
        let ok = vec![verdict(5, H, "PASS", "agent-7")];
        assert_eq!(
            state(&input("1", &src, &["agent-1"], ok.clone())),
            ("pass".into(), "non-author-pass".into())
        );
        // An author's PASS is not one.
        assert_eq!(
            state(&input("1", &src, &["agent-7"], ok.clone())),
            ("missing".into(), "non-author-pass".into())
        );
        assert_eq!(
            state(&input("1", &src, &["AGENT-7"], ok.clone())).0,
            "missing"
        );
        // A PASS on an older head is not one for this head.
        let old = vec![verdict(5, OLD, "PASS", "agent-7")];
        assert_eq!(
            state(&input("1", &src, &["agent-1"], old)),
            ("missing".into(), "pass-on-head".into())
        );
        // No Moai-Session trailer: authors unknown.
        assert_eq!(
            state(&input("1", &src, &[], ok.clone())),
            ("missing".into(), "authors".into())
        );
        // A FAIL after the PASS on the same head reopens it.
        let mut reopened = ok;
        reopened.push(verdict(6, H, "FAIL", "agent-9"));
        assert_eq!(state(&input("1", &src, &["agent-1"], reopened)).0, "fail");
    }

    #[test]
    fn only_member_comments_and_readable_verdicts_count() {
        let src = ["src/a.rs"];
        let mut outsider = verdict(5, H, "PASS", "agent-7");
        outsider.author_association = "NONE".into();
        assert_eq!(
            state(&input("1", &src, &["agent-1"], vec![outsider])).0,
            "missing"
        );
        let mut garbled = verdict(5, H, "PASS", "agent-7");
        garbled.body = "AUDIT #9 (x) head zz — verdict: PASS\nagent-7".into();
        assert_eq!(
            state(&input("0", &["docs/a.md"], &[], vec![garbled])),
            ("fail".into(), "open-fail".into())
        );
        // A short base reads for its word: a FAIL counts, a PASS is not enough.
        let short = Comment {
            id: 5,
            body: format!("AUDIT #9 (x) head {H} base 3333333 — verdict: PASS\nagent-7"),
            author_association: "MEMBER".into(),
        };
        assert_eq!(
            state(&input("1", &src, &["agent-1"], vec![short])).0,
            "missing"
        );
        // A tier the body does not declare, or unreadable files, is tier 2.
        let mut none = input("1", &["docs/a.md"], &[], vec![]);
        none.body = "no tier".into();
        assert_eq!(decide(&none)["tier"], 2);
        let mut unread = input("0", &[], &[], vec![]);
        unread.files = None;
        assert_eq!(decide(&unread)["tier"], 2);
    }

    #[test]
    fn collect_reads_the_pr_its_files_commits_and_comments() {
        let api = "https://api.github.com/repos/o/r";
        let ex = |url: String, body: Value| Exchange {
            request: HttpRequest::new("GET", url),
            response: HttpResponse::new(200).with_body(body),
        };
        let t = RecordedTransport::new(vec![
            ex(
                format!("{api}/pulls/9"),
                json!({"head": {"sha": H}, "body": "audit-tier: 1", "commits": 1}),
            ),
            ex(
                format!("{api}/pulls/9/files?per_page=100"),
                json!([{"filename": "src/a.rs"}]),
            ),
            ex(
                format!("{api}/pulls/9/commits?per_page=100"),
                json!([{"commit": {"message": "x\n\nMoai-Session: agent-1"}}]),
            ),
            ex(
                format!("{api}/issues/9/comments?per_page=100"),
                json!([{"id": 5, "body": verdict(5, H, "PASS", "agent-7").body,
                        "author_association": "MEMBER"}]),
            ),
        ]);
        let i = collect("o", "r", 9, None, &t).unwrap();
        assert_eq!(t.remaining(), 0);
        assert_eq!(i.head, H);
        assert_eq!(decide(&i)["state"], "pass");
    }
}
