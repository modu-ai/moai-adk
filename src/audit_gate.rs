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
//!    head has no qualifying PASS (fail closed, like G9). A verdict names its auditor's session
//!    with a `Moai-Session:` line below its first line (SDD-6, `internal/commit-trailers`); the
//!    auditor is an author when that session or its lane name is one of the authors, and two
//!    verdicts come from the same auditor when their sessions match (else their lane names).
//!    A leader-assigned audit agent (audit centralization, 2026-11-16) names itself with the
//!    fixed line `auditor: leader-audit <token>` below the first line; the token counts as an
//!    author too when it equals one of the `Moai-Session:` values (card C-95).
//!
//! 5. A release audit summary (SPEC-20261125-release-audit-summary): a member's comment whose
//!    first line starts `RELEASE-AUDIT-SUMMARY ` is read by [`summary::parse`], the grammar the
//!    command that builds it shares. A summary that does not read, or that names this head and
//!    has a `FAIL` line, is an open FAIL at every tier (`summary-malformed`, `summary-fail`),
//!    decided after the `AUDIT #` open FAILs. At tier 1 and 2 the summaries that name this head
//!    must all pass and the old path is not consulted: a base other than `Input::release_base`
//!    is `missing` (`summary-base`), a landing with no passing verdict is `missing`
//!    (`summary-landings`), else `pass` (`release-summary`). A summary of another head is not
//!    read. Without a summary rules 1 to 4 decide exactly as before.
//!
//! Not checked in this slice: that the PASS cites the run (AGENTS.md workflow 7) — a later slice
//! adds it; the verdict base's ancestry and the mechanical delta (both need git).

use crate::forge::transport::Transport;
use crate::gate::{self, Comment};
use crate::release_audit_summary::{self as summary, Kind, Line, Summary};
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
    /// The scan registry's text at the merge base and at the head, when the PR changes it
    /// (SPEC-20261124-registry-line-add-tier REQ-8). An input without the field judges the
    /// registry path by its name, tier 2.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub registry: Option<RegistryTexts>,
    /// The merge base of the PR's base and head (40 lowercase hex), read only when a member's
    /// comment is a release audit summary (SPEC-20261125-release-audit-summary REQ-11); `None`
    /// when no summary is present or the merge base could not be read.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub release_base: Option<String>,
}

/// The two sides of the scan registry's change; a side that could not be read is `None`.
#[derive(Clone, Debug, Default, PartialEq, Eq, Deserialize, Serialize)]
#[serde(deny_unknown_fields)]
pub struct RegistryTexts {
    #[serde(default)]
    pub old: Option<String>,
    #[serde(default)]
    pub new: Option<String>,
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
    /// The auditor's session id from a `Moai-Session:` line below the first (SDD-6), the same
    /// key and shape a commit's trailer has.
    session: Option<String>,
    /// The token of a leader-assigned auditor (`auditor: leader-audit <token>`, card C-95),
    /// when the auditor came from that line — compared against the PR's authors too.
    leader_token: Option<String>,
}

impl Verdict {
    /// The same auditor: the same session id when both name one, else the same lane name.
    fn same_auditor(&self, other: &Verdict) -> bool {
        match (&self.session, &other.session) {
            (Some(a), Some(b)) => crate::trailers::same_identity(a, b),
            _ => self.auditor.is_some() && self.auditor == other.auditor,
        }
    }
}

/// The auditor a verdict body names: a lane (`gate::auditor`), or — for a leader-assigned audit
/// agent (card C-95, SPEC-20261119-leader-auditor-verdict) — the fixed line
/// `auditor: leader-audit <token>` below the first line. The token is 1 to 64 characters of
/// `A-Z a-z 0-9 . _ -` with no spaces; an empty, longer or invalid token reads no auditor. `gate.rs` is
/// shared with the merge gate and stays unchanged; the recognition is local to this gate (D-1).
fn verdict_auditor(body: &str) -> Option<String> {
    if let Some(lane) = gate::auditor(body) {
        return Some(lane);
    }
    body.lines()
        .skip(1)
        .filter_map(|l| l.trim().strip_prefix("auditor: leader-audit "))
        .map(str::trim)
        .find(|t| {
            // The strip needs the trailing space, so a hit is never empty (audit N6: the check
            // was dead); the cap keeps a whole paragraph from riding in as one token.
            t.len() <= 64
                && t.bytes()
                    .all(|b| b.is_ascii_alphanumeric() || matches!(b, b'.' | b'_' | b'-'))
        })
        .map(|t| format!("leader-audit {t}"))
}

/// The `Moai-Session:` value of a verdict body below its first line, when it has the id shape.
fn verdict_session(body: &str) -> Option<String> {
    body.lines()
        .skip(1)
        .filter_map(|l| l.trim().strip_prefix("Moai-Session:"))
        .map(str::trim)
        .find(|v| crate::trailers::value_shape(v) != crate::trailers::Shape::Bad)
        .map(String::from)
}

fn verdicts(comments: &[Comment]) -> Vec<Verdict> {
    let mut out: Vec<Verdict> = comments
        .iter()
        .filter_map(|c| {
            let line = gate::verdict_line(&c.body)?;
            let auditor = verdict_auditor(&c.body);
            let session = verdict_session(&c.body);
            let leader_token = auditor
                .as_deref()
                .filter(|a| a.starts_with("leader-audit "))
                .map(|a| a["leader-audit ".len()..].to_string());
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
                session,
                leader_token,
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

/// The PR's changed files for the tier: each path, and for the scan registry the texts of both
/// sides when the input carries them, so a registry change that only grows it is read as
/// `moai bundle check` reads it (SPEC-20261124-registry-line-add-tier REQ-8). Without texts the
/// registry is judged by its name, tier 2.
fn changed_files(input: &Input) -> Vec<gate::ChangedFile> {
    input
        .files
        .iter()
        .flatten()
        .map(|path| {
            let mut file = gate::ChangedFile::named(path);
            if let Some(texts) = input
                .registry
                .as_ref()
                .filter(|_| path == gate::SCAN_REGISTRY)
            {
                file.old = texts.old.clone();
                file.new = texts.new.clone();
            }
            file
        })
        .collect()
}

/// The most landings a `summary-landings` reason names; it counts all of them.
const MAX_NAMED_LANDINGS: usize = 10;

/// Whether a comment is a release audit summary: its first line starts with the header word and
/// a space. `collect` reads the merge base on the same test, so the two cannot disagree.
fn is_summary(body: &str) -> bool {
    body.lines()
        .next()
        .and_then(|first| first.strip_prefix(summary::HEADER))
        .is_some_and(|rest| rest.starts_with(' '))
}

/// Summaries that name the head and read, by comment id.
type Named = Vec<(u64, Summary)>;
/// Summaries that do not read, by comment id, with why.
type Unread = Vec<(u64, String)>;

/// The member summaries of a PR by comment id: the ones that name `head` and read, and the ones
/// that do not read (with why). A summary that reads but names another head is left out (REQ-11).
fn summaries(comments: &[Comment], head: &str) -> (Named, Unread) {
    let mut found: Vec<&Comment> = comments.iter().filter(|c| is_summary(&c.body)).collect();
    found.sort_by_key(|c| c.id);
    let (mut named, mut unread) = (Vec::new(), Vec::new());
    for c in found {
        match summary::parse(&c.body) {
            Ok(s) if s.head == head => named.push((c.id, s)),
            Ok(_) => {}
            Err(why) => unread.push((c.id, why)),
        }
    }
    (named, unread)
}

/// The cards a landing line names, for the reason of a refusal.
fn cards_text(cards: &[i64]) -> String {
    if cards.is_empty() {
        "no card".to_string()
    } else {
        format!(
            "cards {}",
            cards
                .iter()
                .map(i64::to_string)
                .collect::<Vec<_>>()
                .join(",")
        )
    }
}

/// The first [`MAX_NAMED_LANDINGS`] of `items`, and how many more there are.
fn named_list(items: &[String]) -> String {
    let shown = items
        .iter()
        .take(MAX_NAMED_LANDINGS)
        .cloned()
        .collect::<Vec<_>>()
        .join(", ");
    match items.len().checked_sub(MAX_NAMED_LANDINGS) {
        Some(rest) if rest > 0 => format!("{shown} (and {rest} more)"),
        _ => shown,
    }
}

/// The landings of `s` the gate does not count as passed, each as `sha12 (cards …)`: a line with
/// no passing verdict, and a `PASS` line with no verdict or one the gate does not count
/// ([`summary::Verdict::counts`], REQ-14). `ABS` lines belong to a later line and add nothing.
fn uncounted(s: &Summary) -> Vec<String> {
    let name = |line: &Line, cards: &[i64]| format!("{} ({})", &line.sha[..12], cards_text(cards));
    s.lines
        .iter()
        .filter_map(|line| match &line.kind {
            Kind::Abs { .. } => None,
            Kind::None { cards, .. } | Kind::Fail { cards, .. } => Some(name(line, cards)),
            Kind::Pass(verdicts) => {
                let bad: Vec<i64> = verdicts
                    .iter()
                    .filter(|v| !v.counts(s))
                    .map(|v| v.card)
                    .collect();
                (verdicts.is_empty() || !bad.is_empty()).then(|| name(line, &bad))
            }
        })
        .collect()
}

/// The decision: `pass`, `fail` (an open FAIL) or `missing` (no qualifying PASS yet), with the
/// tier it judged by, the rule that decided and why.
pub fn decide(input: &Input) -> Value {
    let out = |state: &str, tier: Option<u8>, rule: &str, reason: String| {
        json!({"state": state, "pr": input.pr, "head": input.head, "tier": tier,
               "rule": rule, "reason": reason, "comment": Value::Null,
               "cites_run_checked": false})
    };
    if input.schema_version != 1 || !gate::is_full_sha(&input.head) {
        return out(
            "fail",
            None,
            "input",
            "schema_version 1 and a 40-hex head are required".into(),
        );
    }
    let tier = gate::tier_of_changes(gate::parse_body_tier(&input.body), &changed_files(input));
    let tier = if input.files.is_none() { 2 } else { tier };
    let mut comments = input.comments.clone();
    crate::comment_authority::retain_members(&mut comments);
    let all = verdicts(&comments);

    // Rule 2: every FAIL (or unreadable verdict) needs a later PASS by the same auditor.
    let open: Vec<&Verdict> = all
        .iter()
        .filter(|f| f.word != "PASS")
        .filter(|f| {
            !all.iter()
                .any(|p| p.word == "PASS" && p.id > f.id && p.same_auditor(f))
        })
        .collect();
    if let Some(f) = open.first() {
        let who = f.auditor.as_deref().unwrap_or("no auditor lane");
        let mut r = out(
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
        r["comment"] = json!(f.id);
        return r;
    }

    // Rule 5: a member's release audit summary. One that does not read, or that names this head
    // and has a FAIL line, is an open FAIL at every tier, whatever the old path would say.
    let (named, unread) = summaries(&comments, &input.head);
    if let Some((id, why)) = unread.first() {
        let mut r = out(
            "fail",
            Some(tier),
            "summary-malformed",
            format!(
                "{} release audit summary comment(s) do not read; comment {id}: {why}",
                unread.len()
            ),
        );
        r["comment"] = json!(id);
        return r;
    }
    for (id, s) in &named {
        let failed: Vec<String> = s
            .lines
            .iter()
            .filter_map(|line| match &line.kind {
                Kind::Fail { why, cards, .. } => Some(format!(
                    "{} ({why}, {})",
                    &line.sha[..12],
                    cards_text(cards)
                )),
                _ => None,
            })
            .collect();
        if !failed.is_empty() {
            let mut r = out(
                "fail",
                Some(tier),
                "summary-fail",
                format!(
                    "release audit summary comment {id} has {} FAIL line(s): {}",
                    failed.len(),
                    named_list(&failed)
                ),
            );
            r["comment"] = json!(id);
            return r;
        }
    }
    if tier == 0 {
        return out(
            "pass",
            Some(0),
            "tier-0",
            "audit tier 0 and no open FAIL".into(),
        );
    }

    // Tier 1 and 2 with a summary of this head: every such summary must pass and the old path is
    // not consulted, so it does not need the PR's commit list (a release PR has over 250).
    for (id, s) in &named {
        if input.release_base.as_deref() != Some(s.base.as_str()) {
            let mut r = out(
                "missing",
                Some(tier),
                "summary-base",
                format!(
                    "release audit summary comment {id} names base {} but the PR's merge base is {}",
                    s.base,
                    input.release_base.as_deref().unwrap_or("unread")
                ),
            );
            r["comment"] = json!(id);
            return r;
        }
        let open = uncounted(s);
        if !open.is_empty() {
            let mut r = out(
                "missing",
                Some(tier),
                "summary-landings",
                format!(
                    "{} landing(s) of release audit summary comment {id} have no passing verdict: {}",
                    open.len(),
                    named_list(&open)
                ),
            );
            r["comment"] = json!(id);
            return r;
        }
    }
    if let Some((id, s)) = named.last() {
        let mut r = out(
            "pass",
            Some(tier),
            "release-summary",
            format!(
                "audit tier {tier}: release audit summary comment {id} covers {} landing(s), each with a passing verdict by an auditor that wrote or held none of its cards",
                s.lines.len()
            ),
        );
        r["comment"] = json!(id);
        return r;
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
    let is_author = |who: &str| {
        authors
            .iter()
            .any(|a| a.eq_ignore_ascii_case(who) || crate::trailers::same_identity(a, who))
    };
    match on_head
        .iter()
        .rev()
        .find(|v| {
            (v.auditor.is_some() || v.session.is_some())
                && !v.auditor.as_deref().is_some_and(is_author)
                && !v.session.as_deref().is_some_and(is_author)
                && !v.leader_token.as_deref().is_some_and(is_author)
        })
    {
        Some(v) => {
            let mut r = out(
                "pass",
                Some(tier),
                "non-author-pass",
                format!(
                    "audit tier {tier}: PASS on this head by {} (comment {}), not an author",
                    v.auditor
                        .as_deref()
                        .or(v.session.as_deref())
                        .unwrap_or_default(),
                    v.id
                ),
            );
            r["comment"] = json!(v.id);
            r
        }
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
/// `None` takes the PR's current head. A renamed file counts by both of its names — `filename`
/// and, when present, `previous_filename` — like check-guard reads them, so a rename cannot
/// walk a protected path out of the tier judgment.
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
    let files: Option<Vec<String>> =
        get_all(t, &format!("{base}/pulls/{pr}/files?per_page=100"), 30)
            .ok()
            .map(|list| {
                list.iter()
                    .flat_map(|f| {
                        ["filename", "previous_filename"]
                            .iter()
                            .filter_map(|k| f[*k].as_str().map(String::from))
                            .collect::<Vec<_>>()
                    })
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
    // Read last, so the four reads every pull request needs keep their order. The merge base of
    // the PR's base and head is read once, and only for a reader that needs it: the scan registry
    // when the PR changes it, a release audit summary when a member posted one (REQ-11).
    let pr_base = p["base"]["sha"].as_str().filter(|s| gate::is_full_sha(s));
    let wants_registry = files
        .as_ref()
        .is_some_and(|f| f.iter().any(|p| p == gate::SCAN_REGISTRY));
    let wants_summary = {
        let mut members = comments.clone();
        crate::comment_authority::retain_members(&mut members);
        members.iter().any(|c| is_summary(&c.body))
    };
    let merge_base = pr_base
        .filter(|_| wants_registry || wants_summary)
        .map(|pr_base| merge_base_of(t, &base, pr_base, &head));
    let registry = merge_base
        .as_ref()
        .filter(|_| wants_registry)
        .map(|merge_base| read_registry(t, &base, merge_base.as_deref(), &head));
    let release_base = merge_base.flatten().filter(|_| wants_summary);
    Ok(Input {
        schema_version: 1,
        pr,
        head,
        body,
        files,
        commit_messages,
        comments,
        registry,
        release_base,
    })
}

/// The merge base of the PR's base and `head`, from the compare answer; `None` when GitHub does
/// not answer or the answer holds no 40-hex commit.
fn merge_base_of<T: Transport>(t: &T, base_url: &str, pr_base: &str, head: &str) -> Option<String> {
    use crate::gate_collect::get_json;
    get_json(
        t,
        &format!("{base_url}/compare/{pr_base}...{head}?per_page=1"),
    )
    .ok()
    .and_then(|(v, _)| {
        v["merge_base_commit"]["sha"]
            .as_str()
            .filter(|s| gate::is_full_sha(s))
            .map(String::from)
    })
}

/// The scan registry at the merge base and at `head` (REQ-8), read only when the PR changes it:
/// the file at both commits through the Git Trees and Blobs APIs ([`registry_text`]), the merge
/// base being the compare answer [`merge_base_of`] gave. A side that cannot be read (no answer,
/// a registry that is no regular file, text that is no base64) is left out, and the decision
/// then judges the registry by its name, tier 2 (REQ-9).
fn read_registry<T: Transport>(
    t: &T,
    base_url: &str,
    merge_base: Option<&str>,
    head: &str,
) -> RegistryTexts {
    let Some(merge_base) = merge_base else {
        return RegistryTexts::default();
    };
    RegistryTexts {
        old: registry_text(t, base_url, merge_base),
        new: registry_text(t, base_url, head),
    }
}

/// The registry text at `rev`, read only when the registry is a regular file there. The contents
/// API answers a symbolic link that points at a normal file of the repository with that file's
/// content, and no documented field tells the two answers apart (SPEC-20261124-registry-line-add-
/// tier, Appendix C), so the path is not read by name. The Git Trees API names each entry's mode,
/// and 120000 is a link: the path is walked one folder at a time from the root tree of `rev`, every
/// step must be a tree (mode 040000) and the last a blob of mode 100644 or 100755, and the text is
/// that blob read by its id. A tree that GitHub says is truncated, a link, a directory, a submodule
/// or any answer that does not match leaves the text out (`None`), so the registry is judged by
/// its name, tier 2.
fn registry_text<T: Transport>(t: &T, base_url: &str, rev: &str) -> Option<String> {
    use crate::gate_collect::get_json;
    let parts: Vec<&str> = gate::SCAN_REGISTRY.split('/').collect();
    let mut at = rev.to_string();
    for (i, part) in parts.iter().enumerate() {
        let (tree, _) = get_json(t, &format!("{base_url}/git/trees/{at}")).ok()?;
        if tree["truncated"] != false {
            return None;
        }
        let entry = tree["tree"]
            .as_array()?
            .iter()
            .find(|e| e["path"] == *part)?;
        let (mode, kind) = (entry["mode"].as_str()?, entry["type"].as_str()?);
        let id = entry["sha"].as_str().filter(|s| gate::is_full_sha(s))?;
        if i + 1 < parts.len() {
            if kind != "tree" || mode != "040000" {
                return None;
            }
            at = id.to_string();
            continue;
        }
        if kind != "blob" || !(mode == "100644" || mode == "100755") {
            return None;
        }
        let (blob, _) = get_json(t, &format!("{base_url}/git/blobs/{id}")).ok()?;
        return blob_text(&blob, id);
    }
    None
}

/// The text of the blob answer for `id`: `content` is base64 when `encoding` says so, and the
/// answer must name the blob that was asked for.
fn blob_text(answer: &Value, id: &str) -> Option<String> {
    if answer["sha"] != id || answer["encoding"] != "base64" {
        return None;
    }
    String::from_utf8(base64_decode(answer["content"].as_str()?)?).ok()
}

/// Standard base64 with the line breaks GitHub puts in; anything else, a wrong length or data
/// after the padding is `None`.
fn base64_decode(text: &str) -> Option<Vec<u8>> {
    let (mut out, mut acc, mut bits) = (Vec::new(), 0u32, 0u32);
    let (mut data, mut pad) = (0usize, 0usize);
    for c in text.bytes() {
        let v = match c {
            b'A'..=b'Z' => c - b'A',
            b'a'..=b'z' => c - b'a' + 26,
            b'0'..=b'9' => c - b'0' + 52,
            b'+' => 62,
            b'/' => 63,
            b'=' => {
                pad += 1;
                continue;
            }
            b'\n' | b'\r' => continue,
            _ => return None,
        };
        if pad > 0 {
            return None;
        }
        data += 1;
        acc = (acc << 6) | u32::from(v);
        bits += 6;
        if bits >= 8 {
            bits -= 8;
            out.push((acc >> bits) as u8);
            acc &= (1 << bits) - 1;
        }
    }
    (pad <= 2 && (data + pad) % 4 == 0).then_some(out)
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
            registry: None,
            release_base: None,
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

    /// SDD-6: commits carry the author's `Moai-Session:` (a session id); a verdict that names
    /// its auditor's session the same way is compared by that id, and its lane name too.
    /// Re-audit B3: a subagent auditor inherits its parent's session id; the worktree key tells
    /// them apart. A legacy bare author id still counts every value of its session as an author.
    #[test]
    fn the_worktree_key_tells_a_subagent_auditor_from_its_parent() {
        const S: &str = "428e289a-1c67-492e-a23b-a5aca15b4d1f";
        let author = format!("{S}:aaaaaaaaaaaa");
        let src = ["src/a.rs"];
        let pass = |session: &str| {
            Comment {
            id: 5,
            body: format!(
                "AUDIT #9 (x) head {H} base {B} — verdict: PASS\nauditor: agent-7\nMoai-Session: {session}"
            ),
            author_association: "MEMBER".into(),
        }
        };
        let own = format!("{S}:bbbbbbbbbbbb");
        assert_eq!(
            state(&input("1", &src, &[author.as_str()], vec![pass(&own)])).0,
            "pass",
            "another worktree of the same host session"
        );
        assert_eq!(
            state(&input("1", &src, &[author.as_str()], vec![pass(&author)])).0,
            "missing",
            "the author's own worktree"
        );
        assert_eq!(
            state(&input("1", &src, &[S], vec![pass(&own)])).0,
            "missing",
            "a legacy bare author id covers its whole session"
        );
    }

    #[test]
    fn a_verdict_names_its_auditors_session_like_a_commit_does() {
        const AUTHOR: &str = "428e289a-1c67-492e-a23b-a5aca15b4d1f";
        const AUDITOR: &str = "4d5550e0-88b3-40ca-8c46-07425eedc812";
        let src = ["src/a.rs"];
        let pass = |session: &str| {
            Comment {
            id: 5,
            body: format!(
                "AUDIT #9 (x) head {H} base {B} — verdict: PASS\nauditor: agent-7\nMoai-Session: {session}"
            ),
            author_association: "MEMBER".into(),
        }
        };
        assert_eq!(
            state(&input("1", &src, &[AUTHOR], vec![pass(AUDITOR)])),
            ("pass".into(), "non-author-pass".into())
        );
        // The auditor's session wrote commits of this PR: not an independent PASS.
        assert_eq!(
            state(&input("1", &src, &[AUTHOR], vec![pass(AUTHOR)])).0,
            "missing"
        );
        // A FAIL is cleared by a later PASS of the same session, not of another one.
        let fail = |id: u64, session: &str| {
            Comment {
            id,
            body: format!(
                "AUDIT #9 (x) head {H} base {B} — verdict: FAIL\nauditor: agent-9\nMoai-Session: {session}"
            ),
            author_association: "MEMBER".into(),
        }
        };
        let mut again = pass(AUDITOR);
        again.id = 7;
        again.body = again.body.replace("agent-7", "agent-8");
        assert_eq!(
            state(&input(
                "1",
                &src,
                &[AUTHOR],
                vec![fail(6, AUDITOR), again.clone()]
            ))
            .0,
            "pass"
        );
        let other = "11111111-2222-3333-4444-555555555555";
        assert_eq!(
            state(&input("1", &src, &[AUTHOR], vec![fail(6, other), again])).0,
            "fail"
        );
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

    /// Card C-95 (SPEC-20261119-leader-auditor-verdict): a leader-assigned audit agent names
    /// itself with the fixed line `auditor: leader-audit <token>`; the gate reads it as the
    /// auditor, a non-member's line is dropped with the whole comment, and a token that equals a
    /// `Moai-Session:` author reads as an author's PASS.
    #[test]
    fn the_leader_assigned_auditor_line_is_a_verdict() {
        let src = ["src/a.rs"];
        let pass = |token: &str, association: &str| Comment {
            id: 5,
            body: format!(
                "AUDIT #9 (x) head {H} base {B} — verdict: PASS\nauditor: leader-audit {token}"
            ),
            author_association: association.into(),
        };
        assert_eq!(
            state(&input(
                "1",
                &src,
                &["agent-1"],
                vec![pass("audit-3218-1", "MEMBER")]
            )),
            ("pass".into(), "non-author-pass".into()),
            "the fixed line names the auditor"
        );
        // The cap's near edge: 64 characters are a token.
        assert_eq!(
            state(&input(
                "1",
                &src,
                &["agent-1"],
                vec![pass(&"a".repeat(64), "MEMBER")]
            )),
            ("pass".into(), "non-author-pass".into()),
            "64 characters are a token"
        );
        // An empty token or one with a character outside the set reads no auditor; a token is
        // capped at 64 characters (audit N6: the dead emptiness check became a length cap).
        for bad in ["", "au@dit", "two tokens", &"a".repeat(65)] {
            assert_eq!(
                state(&input("1", &src, &["agent-1"], vec![pass(bad, "MEMBER")])).0,
                "missing",
                "token {bad:?}"
            );
        }
        // A forged line by a non-member is dropped with the comment (REQ-5).
        assert_eq!(
            state(&input(
                "1",
                &src,
                &["agent-1"],
                vec![pass("audit-3218-1", "NONE")]
            ))
            .0,
            "missing"
        );
        // A token equal to a Moai-Session author reads as an author's PASS (REQ-6). The token
        // charset has no `:`, so the bare session id is the shape that reaches the author check:
        // a legacy bare author id covers its whole session (same_identity).
        const AUTHOR: &str = "428e289a-1c67-492e-a23b-a5aca15b4d1f";
        let own = format!("{AUTHOR}:aaaaaaaaaaaa");
        // The token holds the worktree key too (`id:key`): the `:` leaves the charset, so this
        // shape is refused by the charset alone — the author check behind it is never reached.
        assert_eq!(
            state(&input(
                "1",
                &src,
                &[own.as_str()],
                vec![pass(own.as_str(), "MEMBER")]
            ))
            .0,
            "missing"
        );
        // A bare session id is a charset-valid token that names the author's whole session
        // (legacy shape): the leader_token author check is what refuses it (mutant M1's gap).
        assert_eq!(
            state(&input(
                "1",
                &src,
                &[own.as_str()],
                vec![pass(AUTHOR, "MEMBER")]
            ))
            .0,
            "missing"
        );
    }

    /// Card C-95 REQ-3: the pass and fail decisions name the comment that decided them.
    #[test]
    fn the_decision_names_the_comment_that_decided_it() {
        let src = ["src/a.rs"];
        let v = decide(&input(
            "1",
            &src,
            &["agent-1"],
            vec![verdict(5, H, "PASS", "agent-7")],
        ));
        assert_eq!(v["comment"], 5, "{v}");
        // Tier 0 passes with no comment at all: the field is null.
        let v = decide(&input("0", &["docs/a.md"], &[], vec![]));
        assert!(v["comment"].is_null(), "{v}");
        // An open FAIL names the failing comment.
        let v = decide(&input(
            "0",
            &["docs/a.md"],
            &[],
            vec![verdict(6, OLD, "FAIL", "agent-7")],
        ));
        assert_eq!(v["comment"], 6, "{v}");
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

    /// A PR that renames a tier 2 rule file under docs/ is judged by its old name too,
    /// like check-guard reads `previous_filename` — the rename cannot lower the tier.
    #[test]
    fn collect_judges_a_rename_by_its_previous_filename_too() {
        let api = "https://api.github.com/repos/o/r";
        let ex = |url: String, body: Value| Exchange {
            request: HttpRequest::new("GET", url),
            response: HttpResponse::new(200).with_body(body),
        };
        // Built at run time so the CI scope's docs-path scan does not read the fixture's
        // new name as a reference to a repository document (the file does not exist).
        let new_name = format!("{}/local-check.md", "docs");
        let t = RecordedTransport::new(vec![
            ex(
                format!("{api}/pulls/9"),
                json!({"head": {"sha": H}, "body": "audit-tier: 0", "commits": 1}),
            ),
            ex(
                format!("{api}/pulls/9/files?per_page=100"),
                json!([{"filename": new_name.clone(),
                        "previous_filename": "tools/local_check.sh"}]),
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
        assert_eq!(
            i.files,
            Some(vec![new_name, "tools/local_check.sh".into()]),
            "a renamed entry counts by both of its names"
        );
        let v = decide(&i);
        assert_eq!(v["tier"], 2, "the rename cannot lower the tier: {v}");
    }

    // ---- the release audit summary (SPEC-20261125-release-audit-summary) ---------------------

    use crate::release_audit_summary::{self as summary, Kind, Line, Summary};

    /// The merge base of the release PR, and another commit that is none of the fixtures' ids.
    const MB: &str = "4444444444444444444444444444444444444444";
    const OTHER_HEAD: &str = "6666666666666666666666666666666666666666";
    const S_AUTH: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
    const S_REV: &str = "cccccccc-0000-4000-8000-00000000000c";

    /// The commit id of landing `i` before the last one (the last is the summary's head).
    fn landing_sha(i: usize) -> String {
        format!("{:02x}", 0x50 + i).repeat(20)
    }

    fn pass_verdict(card: i64, reviewer: usize, excluded: Vec<usize>) -> summary::Verdict {
        summary::Verdict {
            card,
            head: "abcdef012345".into(),
            tier: 0,
            reviewer,
            excluded,
            event: 7,
        }
    }

    /// A passing line: card 1 audited by session s2, written or held by session s1.
    fn pass() -> Kind {
        Kind::Pass(vec![pass_verdict(1, 1, vec![0])])
    }

    fn unverdicted(card: i64) -> Kind {
        Kind::None {
            cards: vec![card],
            title: "a direct commit".into(),
        }
    }

    /// A summary of `kinds`, oldest first, as an unbroken chain from `base` to `head`.
    fn built(base: &str, head: &str, sessions: &[&str], kinds: Vec<Kind>) -> Summary {
        let n = kinds.len();
        let mut previous = base.to_string();
        let mut lines = Vec::new();
        for (i, kind) in kinds.into_iter().enumerate() {
            let sha = if i + 1 == n {
                head.to_string()
            } else {
                landing_sha(i)
            };
            lines.push(Line {
                sha: sha.clone(),
                parent: previous[..12].to_string(),
                kind,
            });
            previous = sha;
        }
        Summary {
            head: head.into(),
            base: base.into(),
            sessions: sessions.iter().map(|s| s.to_string()).collect(),
            lines,
        }
    }

    /// The summary of the release PR: base [`MB`], head [`H`], sessions s1 = author, s2 = auditor.
    fn standard(kinds: Vec<Kind>) -> Summary {
        built(MB, H, &[S_AUTH, S_REV], kinds)
    }

    fn member(id: u64, body: String) -> Comment {
        Comment {
            id,
            body,
            author_association: "MEMBER".into(),
        }
    }

    fn posted(id: u64, s: &Summary) -> Comment {
        member(id, summary::comment(s).unwrap())
    }

    /// A release PR: more than 250 commits and no file list, so the authors are unknown and the
    /// tier is 2 — the summary has to decide without either.
    fn release_input(base: Option<&str>, comments: Vec<Comment>) -> Input {
        Input {
            schema_version: 1,
            pr: 9,
            head: H.into(),
            body: "audit-tier: 2\nbody".into(),
            files: None,
            commit_messages: None,
            comments,
            registry: None,
            release_base: base.map(String::from),
        }
    }

    /// A tier 1 PR with one known author, so an `AUDIT #` PASS by agent-7 passes the old path.
    fn tier_1(comments: Vec<Comment>) -> Input {
        let mut i = input("1", &["src/a.rs"], &["agent-1"], comments);
        i.release_base = Some(MB.into());
        i
    }

    fn reason(i: &Input) -> String {
        decide(i)["reason"].as_str().unwrap().to_string()
    }

    fn edit(text: &str, from: &str, to: &str) -> String {
        assert!(
            text.contains(from),
            "`{from}` is not in the comment: {text}"
        );
        text.replacen(from, to, 1)
    }

    fn missing(rule: &str) -> (String, String) {
        ("missing".into(), rule.into())
    }

    #[test]
    fn a_complete_summary_with_non_author_passes_passes_tier_2() {
        let absorbed = Kind::Abs {
            target: H[..12].to_string(),
        };
        let s = standard(vec![pass(), absorbed, pass()]);
        // Without the summary the old path finds no PASS on this head.
        assert_eq!(
            state(&release_input(Some(MB), vec![])),
            missing("pass-on-head")
        );
        let v = decide(&release_input(Some(MB), vec![posted(21, &s)]));
        assert_eq!(v["state"], "pass", "{v}");
        assert_eq!(v["rule"], "release-summary", "{v}");
        assert_eq!(v["tier"], 2, "{v}");
        assert_eq!(v["comment"], 21, "{v}");
        // The tier 1 path, with known authors, reads the summary the same way.
        assert_eq!(
            state(&tier_1(vec![posted(21, &s)])),
            ("pass".into(), "release-summary".into())
        );
        // One landing is the smallest summary; a line that closed two cards counts both.
        let one = standard(vec![pass()]);
        assert_eq!(
            state(&release_input(Some(MB), vec![posted(5, &one)])).0,
            "pass"
        );
        let two = standard(vec![Kind::Pass(vec![
            pass_verdict(1, 1, vec![0]),
            pass_verdict(2, 1, vec![0]),
        ])]);
        assert_eq!(
            state(&release_input(Some(MB), vec![posted(5, &two)])).0,
            "pass"
        );
    }

    #[test]
    fn a_landing_without_a_verdict_fails_the_gate_by_name() {
        let line = Kind::None {
            cards: vec![3, 4],
            title: "a direct commit".into(),
        };
        let s = standard(vec![pass(), line, pass()]);
        // The old path passes on its own (tier 1, a non-author AUDIT # PASS on this head)...
        let old = vec![verdict(5, H, "PASS", "agent-7")];
        assert_eq!(state(&tier_1(old.clone())).0, "pass");
        // ...and a landing without a verdict still stops it, by name.
        let mut with_summary = old;
        with_summary.push(posted(21, &s));
        let i = tier_1(with_summary);
        assert_eq!(state(&i), missing("summary-landings"));
        let why = reason(&i);
        assert!(why.contains(&landing_sha(1)[..12]), "{why}");
        assert!(why.contains("cards 3,4"), "{why}");
        assert!(why.contains("1 landing"), "{why}");
        assert_eq!(decide(&i)["comment"], 21);
        // A long list names the first ten landings and counts them all.
        for (count, named) in [(10usize, 10usize), (11, 10)] {
            let all = standard((0..count).map(|n| unverdicted(n as i64 + 1)).collect());
            let i = release_input(Some(MB), vec![posted(21, &all)]);
            assert_eq!(state(&i), missing("summary-landings"));
            let why = reason(&i);
            assert!(why.contains(&format!("{count} landing")), "{why}");
            let shas: Vec<String> = all.lines.iter().map(|l| l.sha[..12].to_string()).collect();
            for (n, sha) in shas.iter().enumerate() {
                assert_eq!(why.contains(sha.as_str()), n < named, "{n}: {why}");
            }
            assert_eq!(why.contains("more"), count > named, "{why}");
        }
    }

    #[test]
    fn a_reviewer_among_the_excluded_sessions_is_not_a_pass() {
        let counted = |s: Summary| state(&release_input(Some(MB), vec![posted(21, &s)]));
        let pass_line = |reviewer: usize, excluded: Vec<usize>| {
            Kind::Pass(vec![pass_verdict(1, reviewer, excluded)])
        };
        // The accepted case: an auditor that is none of the excluded sessions.
        assert_eq!(
            counted(standard(vec![pass_line(1, vec![0])])),
            ("pass".into(), "release-summary".into())
        );
        // The auditor wrote or held the card: it is one of the excluded sessions.
        assert_eq!(
            counted(standard(vec![pass_line(1, vec![0, 1])])),
            missing("summary-landings")
        );
        // An empty list means the writers are unknown: closed.
        assert_eq!(
            counted(standard(vec![pass_line(1, vec![])])),
            missing("summary-landings")
        );
        // A line that closed two cards is a pass only when both verdicts count.
        let mixed = Kind::Pass(vec![
            pass_verdict(1, 1, vec![0]),
            pass_verdict(2, 1, vec![0, 1]),
        ]);
        assert_eq!(counted(standard(vec![mixed])), missing("summary-landings"));
        // A legacy bare id names every worktree of its session, so it is the same identity as
        // a full id of that session; another worktree of the session is independent.
        let full = |key: &str| format!("{S_AUTH}:{key}");
        let (aaa, bbb) = (full("aaaaaaaaaaaa"), full("bbbbbbbbbbbb"));
        assert_eq!(
            counted(built(MB, H, &[S_AUTH, &bbb], vec![pass_line(1, vec![0])])),
            missing("summary-landings"),
            "bare id against a full id of the same session"
        );
        assert_eq!(
            counted(built(MB, H, &[&aaa, &bbb], vec![pass_line(1, vec![0])])),
            ("pass".into(), "release-summary".into()),
            "two full ids of one session are two worktrees"
        );
    }

    #[test]
    fn a_summary_for_another_head_or_base_does_not_count() {
        let s_other_head = built(MB, OTHER_HEAD, &[S_AUTH, S_REV], vec![pass(), pass()]);
        // A summary of another head is no summary of this PR: the old path decides, whether
        // that is missing or an AUDIT # PASS (a FAIL line in it blocks nothing here).
        let failing_other = built(
            MB,
            OTHER_HEAD,
            &[S_AUTH, S_REV],
            vec![
                pass(),
                Kind::Fail {
                    why: "late_fail".into(),
                    cards: vec![2],
                    title: "x".into(),
                },
            ],
        );
        for stale in [&s_other_head, &failing_other] {
            assert_eq!(
                state(&release_input(Some(MB), vec![posted(21, stale)])),
                missing("pass-on-head")
            );
            let old = vec![verdict(5, H, "PASS", "agent-7"), posted(21, stale)];
            assert_eq!(
                state(&tier_1(old)),
                ("pass".into(), "non-author-pass".into())
            );
        }
        // This head with another base than the PR's merge base: missing, by the base.
        let other_base = built(OLD, H, &[S_AUTH, S_REV], vec![pass(), pass()]);
        let i = release_input(Some(MB), vec![posted(21, &other_base)]);
        assert_eq!(state(&i), missing("summary-base"));
        let old = vec![verdict(5, H, "PASS", "agent-7"), posted(21, &other_base)];
        assert_eq!(state(&tier_1(old)), missing("summary-base"));
        // A merge base that could not be read is not a match either.
        let right = standard(vec![pass(), pass()]);
        assert_eq!(
            state(&release_input(None, vec![posted(21, &right)])),
            missing("summary-base")
        );
        // The control: the same summary with the merge base it names passes.
        assert_eq!(
            state(&release_input(Some(MB), vec![posted(21, &right)])).0,
            "pass"
        );
    }

    #[test]
    fn a_gap_in_the_chain_or_a_wrong_count_is_refused() {
        let s = standard(vec![pass(), pass(), pass()]);
        let text = summary::comment(&s).unwrap();
        let sha1 = landing_sha(1);
        let refused = |body: String| {
            // Whatever else the PR holds, an unreadable summary is a FAIL.
            let old = vec![verdict(5, H, "PASS", "agent-7"), member(21, body)];
            let r = decide(&tier_1(old));
            (
                r["state"].as_str().unwrap().to_string(),
                r["rule"].as_str().unwrap().to_string(),
                r["comment"].clone(),
            )
        };
        let malformed = (
            "fail".to_string(),
            "summary-malformed".to_string(),
            json!(21),
        );
        // The control: the unedited comment reads and passes.
        assert_eq!(
            state(&release_input(Some(MB), vec![member(21, text.clone())])).0,
            "pass"
        );
        // A count that is too high or too low.
        assert_eq!(refused(edit(&text, "landings=3", "landings=4")), malformed);
        assert_eq!(refused(edit(&text, "landings=3", "landings=2")), malformed);
        // A gap: the second landing's first parent is not the first landing.
        assert_eq!(
            refused(edit(&text, "p=505050505050", "p=777777777777")),
            malformed
        );
        // A line cut out, with the count left as it was or corrected to match.
        let without: String = text
            .lines()
            .filter(|l| !l.starts_with(&format!("L {sha1}")))
            .collect::<Vec<_>>()
            .join("\n");
        assert_eq!(refused(without.clone()), malformed);
        assert_eq!(
            refused(edit(&without, "landings=3", "landings=2")),
            malformed
        );
        // A PASS line without a verdict, and a verdict naming a session the table lacks.
        let verdict_text = " v=1:abcdef012345:0:s2:s1:7";
        assert_eq!(refused(edit(&text, verdict_text, "")), malformed);
        assert_eq!(refused(edit(&text, ":s2:s1:7", ":s9:s1:7")), malformed);
        // The first line alone, a summary of another base's chain, and a comment over the cap.
        assert_eq!(
            refused("RELEASE-AUDIT-SUMMARY head=zz".to_string()),
            malformed
        );
        assert_eq!(
            refused(edit(&text, &format!("base={MB}"), &format!("base={OLD}"))),
            malformed
        );
        assert_eq!(
            refused(format!("{text}{}", "\n".repeat(summary::MAX_COMMENT_BYTES))),
            malformed
        );
        // Only a first line that starts `RELEASE-AUDIT-SUMMARY ` makes a comment a summary.
        for not_a_summary in [
            "RELEASE-AUDIT-SUMMARY".to_string(),
            "RELEASE-AUDIT-SUMMARYX head=zz".to_string(),
            format!("hello\n{text}"),
        ] {
            assert_eq!(
                state(&release_input(
                    Some(MB),
                    vec![member(21, not_a_summary.clone())]
                )),
                missing("pass-on-head"),
                "{not_a_summary}"
            );
        }
    }

    #[test]
    fn a_non_member_summary_does_not_count() {
        let good = standard(vec![pass(), pass()]);
        let failing = standard(vec![
            pass(),
            Kind::Fail {
                why: "late_fail".into(),
                cards: vec![2],
                title: "x".into(),
            },
        ]);
        let by = |association: &str, s: &Summary| Comment {
            author_association: association.into(),
            ..posted(21, s)
        };
        // The control: a member's comment of the same text passes.
        assert_eq!(
            state(&release_input(Some(MB), vec![by("MEMBER", &good)])).0,
            "pass"
        );
        for association in ["NONE", "CONTRIBUTOR", "FIRST_TIME_CONTRIBUTOR", ""] {
            // A forged pass is no pass...
            assert_eq!(
                state(&release_input(Some(MB), vec![by(association, &good)])),
                missing("pass-on-head"),
                "{association:?}"
            );
            // ...and a forged FAIL, or text that does not read, blocks nothing.
            let docs = input("0", &["docs/a.md"], &[], vec![by(association, &failing)]);
            assert_eq!(state(&docs).0, "pass", "{association:?}");
            let mut garbled = by(association, &good);
            garbled.body = "RELEASE-AUDIT-SUMMARY head=zz".into();
            let docs = input("0", &["docs/a.md"], &[], vec![garbled]);
            assert_eq!(state(&docs).0, "pass", "{association:?}");
        }
    }

    #[test]
    fn an_open_fail_still_blocks_a_valid_summary() {
        let s = standard(vec![pass(), pass()]);
        // The control: the summary alone passes.
        assert_eq!(
            state(&release_input(Some(MB), vec![posted(21, &s)])).0,
            "pass"
        );
        // An AUDIT # FAIL stands against the head until its auditor passes it later.
        let fail = verdict(5, OLD, "FAIL", "agent-7");
        let i = release_input(Some(MB), vec![fail.clone(), posted(21, &s)]);
        let v = decide(&i);
        assert_eq!(v["state"], "fail", "{v}");
        assert_eq!(v["rule"], "open-fail", "{v}");
        assert_eq!(v["comment"], 5, "{v}");
        // A later PASS by another auditor clears nothing; one by the same auditor does.
        let other = vec![
            fail.clone(),
            verdict(6, H, "PASS", "agent-8"),
            posted(21, &s),
        ];
        assert_eq!(state(&release_input(Some(MB), other)).0, "fail");
        let same = vec![fail, verdict(6, H, "PASS", "agent-7"), posted(21, &s)];
        assert_eq!(
            state(&release_input(Some(MB), same)),
            ("pass".into(), "release-summary".into())
        );
    }

    #[test]
    fn a_failing_summary_blocks_even_when_the_old_path_passes() {
        let failing = standard(vec![
            pass(),
            Kind::Fail {
                why: "late_fail".into(),
                cards: vec![2],
                title: "a late FAIL".into(),
            },
            pass(),
        ]);
        let old = vec![verdict(5, H, "PASS", "agent-7")];
        // The control: the old path alone passes.
        assert_eq!(
            state(&tier_1(old.clone())),
            ("pass".into(), "non-author-pass".into())
        );
        let mut both = old;
        both.push(posted(21, &failing));
        let v = decide(&tier_1(both));
        assert_eq!(v["state"], "fail", "{v}");
        assert_eq!(v["rule"], "summary-fail", "{v}");
        assert_eq!(v["comment"], 21, "{v}");
        // Tier 0 asks only for no open FAIL, and a failing summary is one.
        let docs = input("0", &["docs/a.md"], &[], vec![posted(21, &failing)]);
        assert_eq!(state(&docs), ("fail".into(), "summary-fail".into()));
        // A summary with no FAIL line asks tier 0 for nothing, whatever its landings are.
        let open = standard(vec![pass(), unverdicted(3)]);
        let docs = input("0", &["docs/a.md"], &[], vec![posted(21, &open)]);
        assert_eq!(state(&docs).0, "pass");
    }

    #[test]
    fn collect_reads_the_merge_base_only_when_a_summary_is_present() {
        let api = "https://api.github.com/repos/o/r";
        let compare = format!("{api}/compare/{B}...{H}?per_page=1");
        let ex = |url: String, body: Value| Exchange {
            request: HttpRequest::new("GET", url),
            response: HttpResponse::new(200).with_body(body),
        };
        let s = standard(vec![pass(), pass()]);
        let text = summary::comment(&s).unwrap();
        let comment = |id: u64, body: &str, association: &str| json!({"id": id, "body": body, "author_association": association});
        // The four reads every PR needs, then the comments given, then `more`.
        let transport = |comments: Vec<Value>, files: Value, more: Vec<Exchange>| {
            let mut all = vec![
                ex(
                    format!("{api}/pulls/9"),
                    json!({"head": {"sha": H}, "base": {"sha": B}, "body": "audit-tier: 1",
                           "commits": 1}),
                ),
                ex(format!("{api}/pulls/9/files?per_page=100"), files),
                ex(
                    format!("{api}/pulls/9/commits?per_page=100"),
                    json!([{"commit": {"message": "x\n\nMoai-Session: agent-1"}}]),
                ),
                ex(
                    format!("{api}/issues/9/comments?per_page=100"),
                    Value::Array(comments),
                ),
            ];
            all.extend(more);
            RecordedTransport::new(all)
        };
        let compares = |t: &RecordedTransport| {
            t.sent
                .borrow()
                .iter()
                .filter(|r| r.url.contains("/compare/"))
                .count()
        };
        let src = || json!([{"filename": "src/a.rs"}]);
        let answer = || ex(compare.clone(), json!({"merge_base_commit": {"sha": MB}}));

        // A member's summary: the merge base is read once, after the comments, and it decides.
        let t = transport(vec![comment(21, &text, "MEMBER")], src(), vec![answer()]);
        let i = collect("o", "r", 9, None, &t).unwrap();
        assert_eq!(t.remaining(), 0);
        assert_eq!(compares(&t), 1);
        assert_eq!(i.release_base.as_deref(), Some(MB));
        assert_eq!(state(&i), ("pass".into(), "release-summary".into()));

        // No summary, a non-member's summary, and a header that is not the first line: no read.
        let verdict_body = verdict(5, H, "PASS", "agent-7").body;
        let later_line = format!("hello\n{text}");
        for (what, comments) in [
            ("no summary", vec![comment(5, &verdict_body, "MEMBER")]),
            ("a non-member's summary", vec![comment(21, &text, "NONE")]),
            (
                "a header on a later line",
                vec![comment(21, &later_line, "MEMBER")],
            ),
        ] {
            let t = transport(comments, src(), vec![]);
            let i = collect("o", "r", 9, None, &t).unwrap();
            assert_eq!(t.remaining(), 0, "{what}");
            assert_eq!(compares(&t), 0, "{what}: no compare request");
            assert_eq!(i.release_base, None, "{what}");
        }

        // A merge base that cannot be read leaves the field empty, and the gate is missing.
        let t = transport(vec![comment(21, &text, "MEMBER")], src(), vec![]);
        let i = collect("o", "r", 9, None, &t).unwrap();
        assert_eq!(compares(&t), 1);
        assert_eq!(i.release_base, None);
        assert_eq!(state(&i), missing("summary-base"));

        // A PR that also changes the scan registry reads the merge base once, for both.
        let files = json!([{"filename": "src/a.rs"}, {"filename": gate::SCAN_REGISTRY}]);
        let t = transport(vec![comment(21, &text, "MEMBER")], files, vec![answer()]);
        let i = collect("o", "r", 9, None, &t).unwrap();
        assert_eq!(compares(&t), 1, "one compare request for both readers");
        assert_eq!(i.release_base.as_deref(), Some(MB));
        assert_eq!(t.remaining(), 0);

        // A registry-only PR reads it for the registry, and the input carries no release base.
        let files = json!([{"filename": gate::SCAN_REGISTRY}]);
        let t = transport(
            vec![comment(5, &verdict_body, "MEMBER")],
            files,
            vec![answer()],
        );
        let i = collect("o", "r", 9, None, &t).unwrap();
        assert_eq!(compares(&t), 1);
        assert_eq!(i.release_base, None);
        assert!(i.registry.is_some());
    }

    /// The grammar refuses a `PASS` line with no verdict; the gate does not count one either,
    /// so a summary built another way cannot pass a landing without a verdict.
    #[test]
    fn a_pass_line_without_a_verdict_is_not_counted() {
        let empty = standard(vec![Kind::Pass(vec![]), pass()]);
        let open = uncounted(&empty);
        assert_eq!(open.len(), 1, "{open:?}");
        assert!(open[0].starts_with(&landing_sha(0)[..12]), "{open:?}");
        assert!(uncounted(&standard(vec![pass(), pass()])).is_empty());
        let absorbed = Kind::Abs {
            target: H[..12].to_string(),
        };
        assert!(uncounted(&standard(vec![absorbed, pass()])).is_empty());
    }
}
