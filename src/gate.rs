//! Leader merge gate rules (Issue #2674 S1): the checks `target/leader/gate_merge.sh`,
//! `owner_ok.py` and `fail_ok.py` made, as pure functions over a snapshot of one pull request.
//! Nothing here reads the network or git; the caller gathers the snapshot and this module
//! decides. Every rule names itself in the result so a stop says which rule stopped it.

use serde::{Deserialize, Serialize};

/// One pull request comment: its id (ascending ids are posting order) and body.
#[derive(Clone, Debug, Deserialize, Serialize)]
#[serde(deny_unknown_fields)]
pub struct Comment {
    pub id: u64,
    pub body: String,
}

/// One check run on the head: its name and conclusion (`null` while it runs).
#[derive(Clone, Debug, Deserialize, Serialize)]
#[serde(deny_unknown_fields)]
pub struct CheckRun {
    pub name: String,
    pub conclusion: Option<String>,
}

/// Everything the gate decides from. File lists are repository paths.
#[derive(Clone, Debug, Deserialize, Serialize)]
#[serde(deny_unknown_fields)]
pub struct Snapshot {
    pub schema_version: u32,
    pub pr: u64,
    /// The PR head, 40 lowercase hex.
    pub head: String,
    pub comments: Vec<Comment>,
    pub checks: Vec<CheckRun>,
    /// Files the PR changes (merge base to head).
    pub pr_files: Vec<String>,
    /// Files the tip changed since the verdict's base.
    pub tip_files_since_verdict_base: Vec<String>,
    /// Whether `git merge-tree` of the tip and the head is clean.
    pub merge_clean: bool,
    /// Overlapping `docs/` files whose merged text keeps every line each side added.
    #[serde(default)]
    pub docs_lines_kept: Vec<String>,
    /// Whether every skill file in the merged tree matches its source-manifest sha256.
    #[serde(default)]
    pub merged_manifest_ok: bool,
    /// PR numbers whose REVIEW.md draft the user confirmed (#2614 decision).
    #[serde(default)]
    pub review_confirmed_prs: Vec<u64>,
    /// Exit status of the PR body Self-Audit checker.
    pub body_check_ok: bool,
    /// Whether the verdict's base is a commit the repository has (`git cat-file -e`). Absent
    /// means unknown, and the gate stops (audit 6850181707 N4).
    #[serde(default)]
    pub verdict_base_known: bool,
    /// The sessions of the claims the PR's own commits name in `Moai-Claim: <issue>#<id>`
    /// trailers (merge base to head, merges skipped). Absent means unknown, and G9 stops.
    #[serde(default)]
    pub pr_authors: Option<Vec<String>>,
    /// The integration tip's push CI as G8 reads it (#2777). Absent means it was not read,
    /// and the gate stops (unknown is never a pass).
    #[serde(default)]
    pub tip_ci: Option<TipCi>,
    /// Whether the PR body carries a `moai-tip-repair:` line (#2665): the one kind of PR a red
    /// tip lets through.
    #[serde(default)]
    pub tip_repair: bool,
    /// The PR's own commits (merge base to head, merges skipped) with no `Moai-Claim` trailer
    /// (G10, the `commit_claim_lines` check `flow-gate-verdict` made). Absent means unknown,
    /// and the gate stops.
    #[serde(default)]
    pub commits_without_claim: Option<Vec<String>>,
    /// Whether the PR body carries a `Moai-Claim: ` line (G11, the `pr_body_claim` check
    /// `flow-gate-verdict` made). Absent means no.
    #[serde(default)]
    pub body_claim: bool,
    /// The mechanical delta from the last verdict's head to the current head (#2833): absent
    /// when they are the same head or the last verdict is not a PASS.
    #[serde(default)]
    pub auto_delta: Option<AutoDelta>,
}

/// The mechanical delta (#2833): whether the current head `to` is the verdict head `from` with
/// only the base branch merged in. `ok` is the collector's finding; the gate also checks that
/// every file's patch-id is the same on both sides before it counts the PASS on `from` for `to`.
#[derive(Clone, Debug, Deserialize, Serialize)]
#[serde(deny_unknown_fields)]
pub struct AutoDelta {
    pub from: String,
    pub to: String,
    pub ok: bool,
    /// Why it is not a mechanical delta, or a summary when it is.
    #[serde(default)]
    pub detail: String,
    /// The PR's files, each with its change's patch-id against its merge base on both heads.
    #[serde(default)]
    pub files: Vec<DeltaFile>,
}

/// One PR file of a mechanical delta: `git patch-id --stable` of its change on each head.
#[derive(Clone, Debug, Deserialize, Serialize)]
#[serde(deny_unknown_fields)]
pub struct DeltaFile {
    pub path: String,
    pub patch_id_from: String,
    pub patch_id_to: String,
}

/// The tip's push CI for G8: `state` is `green`, `flaky` or `red` from the last completed CI
/// (`decided_by` names the commit it ran on), or `unknown` when none could be read.
#[derive(Clone, Debug, Deserialize, Serialize)]
#[serde(deny_unknown_fields)]
pub struct TipCi {
    pub state: String,
    /// The tip commit the gate judged, 40 hex.
    pub tip: String,
    /// The commit whose completed CI decided `state`: the tip, or an earlier tip commit when
    /// the tip's own CI is still running or was only cancelled.
    #[serde(default)]
    pub decided_by: Option<String>,
    /// What was found, for the stop reason (failed checks with run ids, or why unknown).
    #[serde(default)]
    pub detail: String,
}

const MANIFEST: &str = "skills/moai/references/source-manifest.json";

/// A stop: the rule id (G1…G8 of Issue #2674 comment 6849084101, G8 is #2777; G9 of the
/// leader scope 6850215811; G10 and G11 carry over the claim-line checks of the removed
/// `flow-gate-verdict`, audit 6864692054 B1) and why.
#[derive(Clone, Debug, PartialEq, Serialize)]
pub struct Stop {
    pub rule: &'static str,
    pub reason: String,
}

/// Whether `s` is a full commit id as verdict lines spell it: 40 lowercase hex.
pub fn is_full_sha(s: &str) -> bool {
    is_hex40(s)
}

fn is_hex40(s: &str) -> bool {
    s.len() == 40
        && s.bytes()
            .all(|b| b.is_ascii_digit() || (b'a'..=b'f').contains(&b))
}

/// A verdict is a comment whose first line starts with `AUDIT #`; a line quoted further down
/// (a repair note, a merge record) is never a verdict (audit 6850181707 B1).
fn verdict_line(body: &str) -> Option<&str> {
    body.lines().next().filter(|l| l.starts_with("AUDIT #"))
}

/// `(head, base, verdict)` from a verdict line `… head <40hex> base <40hex> — verdict: <V>`.
pub fn parse_verdict(line: &str) -> Option<(String, String, String)> {
    // The last ` head ` is the verdict's own: a title may contain the word (audit N3).
    let h = line.rfind(" head ")? + 6;
    let head = line.get(h..h + 40)?;
    let rest = line.get(h + 40..)?;
    let rest = rest.strip_prefix(" base ")?;
    let base = rest.get(..40)?;
    let verdict = rest.get(40..)?.strip_prefix(" — verdict: ")?;
    if !is_hex40(head) || !is_hex40(base) || verdict.is_empty() || verdict.contains(' ') {
        return None;
    }
    Some((head.into(), base.into(), verdict.into()))
}

/// The first line of a verdict comment (#2833 D4): `AUDIT #<pr> (<title>) head <head> base
/// <base> — verdict: <verdict>`, refused unless [`parse_verdict`] reads back exactly these
/// values — the line the gate's G1 reads is the line the helper prints.
pub fn format_verdict_line(
    pr: u64,
    title: &str,
    head: &str,
    base: &str,
    verdict: &str,
) -> Result<String, String> {
    if !is_hex40(head) || !is_hex40(base) {
        return Err("head and base must be 40 lowercase hex".into());
    }
    if !matches!(verdict, "PASS" | "FAIL") {
        return Err("verdict must be PASS or FAIL".into());
    }
    if title.trim().is_empty() || title.contains(['\n', '\r']) {
        return Err("title must be one non-empty line".into());
    }
    let line = format!(
        "AUDIT #{pr} ({}) head {head} base {base} — verdict: {verdict}",
        title.trim()
    );
    match parse_verdict(&line) {
        Some((h, b, v)) if h == head && b == base && v == verdict => Ok(line),
        _ => Err("the line would not read back as this head, base and verdict".into()),
    }
}

/// The auditor of a verdict: the first lane name (`agent-N`, `pr-N`, with an optional
/// `-suffix`) after the `AUDIT #` line. A lane named further on is not the author (owner_ok).
pub fn auditor(body: &str) -> Option<String> {
    let mut lines = body.lines();
    lines.next().filter(|l| l.starts_with("AUDIT #"))?;
    let rest: String = lines.collect::<Vec<_>>().join("\n");
    let b = rest.as_bytes();
    let mut i = 0;
    while i < b.len() {
        // Byte offsets inside a multibyte character (Korean text around the lane name) are
        // skipped; every slice below starts on an ASCII byte.
        let starts = rest.is_char_boundary(i)
            && (rest[i..].starts_with("agent-") || rest[i..].starts_with("pr-"))
            && (i == 0 || !(b[i - 1].is_ascii_alphanumeric() || b[i - 1] == b'-'));
        if starts {
            let dash = i + rest[i..].find('-').unwrap() + 1;
            let digits = rest[dash..].bytes().take_while(u8::is_ascii_digit).count();
            if digits > 0 {
                let mut end = dash + digits;
                // An optional `-suffix` of letters and digits (`agent-4-hB`).
                if b.get(end) == Some(&b'-') {
                    let n = rest[end + 1..]
                        .bytes()
                        .take_while(u8::is_ascii_alphanumeric)
                        .count();
                    if n > 0 {
                        end += 1 + n;
                    }
                }
                let next_ok = b
                    .get(end)
                    .is_none_or(|c| !(c.is_ascii_alphanumeric() || *c == b'-'));
                if next_ok {
                    return Some(rest[i..end].to_string());
                }
            }
        }
        i += 1;
    }
    None
}

/// `(comment id, lane, head)` of every `moai-audit-claim: lane=<lane> [head=<40hex>]` comment.
pub fn audit_claims(comments: &[Comment]) -> Vec<(u64, String, String)> {
    // `moai-audit-claim 철회: <lane> (<claim id>)` withdraws that claim (audit N2). A free-text
    // withdrawal is not read, so the claim keeps counting (the gate stops rather than guesses).
    // `(lane, claim id)`: the withdrawal names both, and only a claim whose lane matches is
    // withdrawn — a mistyped id never removes another lane's claim (audit 6850638994 B2).
    let withdrawn: Vec<(String, u64)> = comments
        .iter()
        .filter_map(|c| {
            let first = c.body.lines().next()?;
            let rest = first.strip_prefix("moai-audit-claim 철회: ")?;
            let open = rest.find('(')?;
            let close = rest[open..].find(')')? + open;
            let lane = rest[..open].trim().to_string();
            Some((lane, rest[open + 1..close].parse().ok()?))
        })
        .collect();
    let mut out: Vec<_> = comments
        .iter()
        .filter_map(|c| {
            let first = c.body.lines().next()?;
            let rest = first.strip_prefix("moai-audit-claim: lane=")?;
            let lane: String = rest
                .chars()
                .take_while(|ch| ch.is_ascii_alphanumeric() || *ch == '-')
                .collect();
            if lane.is_empty() {
                return None;
            }
            let after = &rest[lane.len()..];
            let head = match after.strip_prefix(" head=") {
                Some(h) if h.get(..40).is_some_and(is_hex40) => h[..40].to_string(),
                Some(_) => return None,
                None if after.is_empty() || after.starts_with(' ') => String::new(),
                None => return None,
            };
            Some((c.id, lane, head))
        })
        .collect();
    out.retain(|c| {
        !withdrawn
            .iter()
            .any(|(lane, id)| *id == c.0 && *lane == c.1)
    });
    out.sort();
    out
}

/// The mechanical delta the gate accepts (#2833 D1): the last verdict is a PASS on `from`, the
/// current head is `to`, the collector found only base-branch merges in between, and every PR
/// file's patch-id is the same on both heads (checked here again — a snapshot may come from
/// stdin). `None` otherwise.
pub fn accepted_auto_delta(s: &Snapshot) -> Option<&AutoDelta> {
    let d = s.auto_delta.as_ref()?;
    let line = s
        .comments
        .iter()
        .filter_map(|c| verdict_line(&c.body).map(|l| (c.id, l)))
        .max_by_key(|(id, _)| *id)?
        .1;
    let (head, _, verdict) = parse_verdict(line)?;
    let same = |f: &DeltaFile| !f.patch_id_from.is_empty() && f.patch_id_from == f.patch_id_to;
    (d.ok
        && verdict == "PASS"
        && head == d.from
        && d.to == s.head
        && d.from != d.to
        && is_hex40(&d.from)
        && !d.files.is_empty()
        && d.files.iter().all(same))
    .then_some(d)
}

/// G1: the last verdict is a PASS on the current head with a 40-hex base, or a PASS on an
/// earlier head the current one is a mechanical delta of (#2833). Returns the base and the
/// head the verdict judged.
fn g1(s: &Snapshot) -> Result<(String, String), Stop> {
    let last = s
        .comments
        .iter()
        .filter_map(|c| verdict_line(&c.body).map(|l| (c.id, l)))
        .max_by_key(|(id, _)| *id);
    let stop = |reason: String| Stop { rule: "G1", reason };
    let (_, line) = last.ok_or_else(|| stop("no AUDIT verdict".into()))?;
    let (head, base, verdict) = parse_verdict(line).ok_or_else(|| {
        stop(
            "the last verdict line does not read `head <40hex> base <40hex> — verdict: <V>`".into(),
        )
    })?;
    if head != s.head && accepted_auto_delta(s).is_none() {
        let why = match &s.auto_delta {
            Some(d) if !d.ok => format!("; not a mechanical delta: {}", d.detail),
            Some(_) => "; the delta's patch-ids or heads do not match".to_string(),
            None => String::new(),
        };
        return Err(stop(format!(
            "the last verdict is on {head}, not the current head{why}"
        )));
    }
    if verdict != "PASS" {
        return Err(stop(format!("the last verdict on this head is {verdict}")));
    }
    if !s.verdict_base_known {
        return Err(stop(format!(
            "the verdict base {base} is not a known commit"
        )));
    }
    Ok((base, head))
}

/// `(claim id, lane)` from a comment whose first line is a leader handoff,
/// `moai-claim-handoff: comment=<id> to=<lane> by=leader` (the lane is letters, digits and
/// `-`). Any other first line, or the line later in the body, is not a handoff.
pub fn leader_handoff(body: &str) -> Option<(u64, String)> {
    let first = body.lines().next()?;
    let rest = first.strip_prefix("moai-claim-handoff: comment=")?;
    let (id, rest) = rest.split_once(' ')?;
    let id: u64 = id.parse().ok()?;
    let rest = rest.strip_prefix("to=")?;
    let (lane, rest) = rest.split_once(' ')?;
    if lane.is_empty()
        || !lane
            .chars()
            .all(|ch| ch.is_ascii_alphanumeric() || ch == '-')
    {
        return None;
    }
    // The shell rule matched the line's start; anything after `by=leader` is ignored.
    rest.starts_with("by=leader")
        .then(|| (id, lane.to_string()))
}

/// G2: the last verdict's auditor holds the PR's lowest audit claim or the lowest claim on the
/// current head, or the lane a leader handoff moved one of those to. No claim at all passes.
fn g2(s: &Snapshot, judged: &str) -> Result<(), Stop> {
    let claims = audit_claims(&s.comments);
    if claims.is_empty() {
        return Ok(());
    }
    let stop = |reason: String| Stop { rule: "G2", reason };
    let (_, body) = s
        .comments
        .iter()
        .filter(|c| verdict_line(&c.body).is_some())
        .map(|c| (c.id, &c.body))
        .max_by_key(|(id, _)| *id)
        .ok_or_else(|| stop("no verdict".into()))?;
    let mut allowed: Vec<(u64, String, String)> = vec![claims[0].clone()];
    if let Some(on_head) = claims.iter().find(|c| c.2 == judged) {
        if on_head.0 != claims[0].0 {
            allowed.push(on_head.clone());
        }
    }
    // #2860 rule 1 (shell owner_ok.py): a leader handoff moves an allowed claim to another
    // lane. Only a comment whose first line is `moai-claim-handoff: comment=<id> to=<lane>
    // by=leader`, that names an allowed claim and comes after it, counts; the latest one wins.
    let mut by_id: Vec<&Comment> = s.comments.iter().collect();
    by_id.sort_by_key(|c| c.id);
    for c in by_id {
        let Some((claim, lane)) = leader_handoff(&c.body) else {
            continue;
        };
        if let Some(a) = allowed.iter_mut().find(|a| a.0 == claim && c.id > claim) {
            a.1 = lane;
        }
    }
    let names = allowed
        .iter()
        .map(|c| format!("{}#{}", c.1, c.0))
        .collect::<Vec<_>>()
        .join(", ");
    let who = auditor(body).ok_or_else(|| {
        stop(format!(
            "the last verdict names no auditor lane (allowed: {names})"
        ))
    })?;
    if allowed.iter().any(|c| c.1 == who) {
        Ok(())
    } else {
        Err(stop(format!(
            "the last verdict is by {who}, which holds no allowed audit claim ({names})"
        )))
    }
}

/// G9 (#2674 R1, leader scope 6850215811): the last verdict's auditor is not an author — no
/// session whose claim a commit of this PR names in its `Moai-Claim` trailer. Unknown authors
/// stop the gate rather than pass it.
fn g9(s: &Snapshot) -> Result<(), Stop> {
    let stop = |reason: String| Stop { rule: "G9", reason };
    let authors = s
        .pr_authors
        .as_ref()
        .ok_or_else(|| stop("the authors of the PR's commits are unknown".into()))?;
    let body = s
        .comments
        .iter()
        .filter(|c| verdict_line(&c.body).is_some())
        .max_by_key(|c| c.id)
        .map(|c| &c.body)
        .ok_or_else(|| stop("no verdict".into()))?;
    let who = auditor(body).ok_or_else(|| stop("the last verdict names no auditor lane".into()))?;
    match authors.iter().find(|a| a.eq_ignore_ascii_case(&who)) {
        Some(a) => Err(stop(format!(
            "the last verdict is by {who}, and {a} wrote commits of this PR (Moai-Claim trailer); an author never audits its own PR"
        ))),
        None => Ok(()),
    }
}

/// G10 (audit 6864692054 B1, was `flow-gate-verdict` `commit_claim_lines`): every commit of
/// the PR that is not a merge carries a `Moai-Claim` trailer. A commit without one hides its
/// author from G9. Unknown stops the gate rather than passes it.
fn g10(s: &Snapshot) -> Result<(), Stop> {
    let stop = |reason: String| Stop {
        rule: "G10",
        reason,
    };
    let missing = s
        .commits_without_claim
        .as_ref()
        .ok_or_else(|| stop("the PR's commits were not checked for a claim trailer".into()))?;
    if missing.is_empty() {
        return Ok(());
    }
    let short: Vec<&str> = missing
        .iter()
        .map(|c| c.get(..10).unwrap_or(c.as_str()))
        .collect();
    Err(stop(format!(
        "{} commit(s) of this PR carry no `Moai-Claim` trailer: {}",
        missing.len(),
        short.join(", ")
    )))
}

/// G11 (audit 6864692054 B1, was `flow-gate-verdict` `pr_body_claim`): the PR body carries a
/// `Moai-Claim: ` line.
fn g11(s: &Snapshot) -> Result<(), Stop> {
    if s.body_claim {
        Ok(())
    } else {
        Err(Stop {
            rule: "G11",
            reason: "the PR body carries no `Moai-Claim: ` line".into(),
        })
    }
}

/// G3: a FAIL on the current head blocks unless every FAIL there and the later last PASS
/// come from the same auditor (it reversed itself).
fn g3(s: &Snapshot, judged: &str) -> Result<(), Stop> {
    let mut on_head: Vec<(u64, String, String)> = s
        .comments
        .iter()
        .filter_map(|c| {
            let (h, _, v) = parse_verdict(verdict_line(&c.body)?)?;
            (h == judged).then(|| (c.id, v, auditor(&c.body).unwrap_or_default()))
        })
        .collect();
    on_head.sort();
    let fails: Vec<_> = on_head.iter().filter(|v| v.1 == "FAIL").collect();
    if fails.is_empty() {
        return Ok(());
    }
    let last = on_head.last().unwrap();
    let reversed = last.1 == "PASS"
        && !last.2.is_empty()
        && fails.iter().all(|f| f.2 == last.2 && f.0 < last.0);
    if reversed {
        Ok(())
    } else {
        Err(Stop {
            rule: "G3",
            reason: format!(
                "{} FAIL verdict(s) on this head and no later PASS by the same auditor",
                fails.len()
            ),
        })
    }
}

/// G5: the `ci` aggregate succeeded, and every `ci / …` or `windows-build` run succeeded
/// (`windows-build` may be `skipped` on a pull request: it runs after the merge, U10).
fn g5(s: &Snapshot) -> Result<(), Stop> {
    let watched = |n: &str| {
        n == "ci"
            || n == "rc-ci"
            || n.starts_with("ci / ")
            || n.starts_with("rc-ci / ")
            || n.starts_with("windows-build")
    };
    let mut bad = Vec::new();
    let mut aggregate = false;
    for c in s.checks.iter().filter(|c| watched(&c.name)) {
        let concl = c.conclusion.as_deref().unwrap_or("pending");
        // `rc-ci` is the aggregate's earlier name (flow.rs `CI_CHECK_NAMES`, audit N1).
        if (c.name == "ci" || c.name == "rc-ci") && concl == "success" {
            aggregate = true;
        }
        let skipped_windows = c.name.starts_with("windows-build") && concl == "skipped";
        if concl != "success" && !skipped_windows {
            bad.push(format!("{}={concl}", c.name));
        }
    }
    if !bad.is_empty() {
        return Err(Stop {
            rule: "G5",
            reason: format!("check runs not successful: {}", bad.join(", ")),
        });
    }
    if !aggregate {
        return Err(Stop {
            rule: "G5",
            reason: "no successful `ci` aggregate on this head".into(),
        });
    }
    Ok(())
}

/// G6: a clean merge with the tip; files both the PR and the tip (since the verdict's base)
/// changed are allowed only as `docs/` files that keep both sides' lines, or the source manifest
/// when the merged tree matches it. Returns the overlap.
fn g6(s: &Snapshot) -> Result<Vec<String>, Stop> {
    let stop = |reason: String| Stop { rule: "G6", reason };
    if !s.merge_clean {
        return Err(stop("merge-tree conflict with the tip".into()));
    }
    if s.pr_files.is_empty() {
        return Err(stop("the PR changes no files".into()));
    }
    let overlap: Vec<String> = s
        .pr_files
        .iter()
        .filter(|f| s.tip_files_since_verdict_base.contains(f))
        .cloned()
        .collect();
    for f in &overlap {
        if f == MANIFEST {
            if !s.merged_manifest_ok {
                return Err(stop(
                    "the merged source manifest does not match the skill files".into(),
                ));
            }
        } else if f.starts_with("docs/") {
            if !s.docs_lines_kept.contains(f) {
                return Err(stop(format!("merged {f} lost lines one side added")));
            }
        } else {
            return Err(stop(format!("overlap in non-doc file: {f}")));
        }
    }
    Ok(overlap)
}

/// G7: a PR that adds or changes a REVIEW.md merges only after the user confirmed it.
fn g7(s: &Snapshot) -> Result<(), Stop> {
    let touches = s
        .pr_files
        .iter()
        .any(|f| f == "REVIEW.md" || f.ends_with("/REVIEW.md"));
    if touches && !s.review_confirmed_prs.contains(&s.pr) {
        return Err(Stop {
            rule: "G7",
            reason: "this PR changes REVIEW.md; it waits for the user's confirmation (#2614)"
                .into(),
        });
    }
    Ok(())
}

/// G8 (#2665, #2777): a red tip stops every PR but its repair. The tip's state comes from its
/// last completed push CI; `unknown` (nothing read, or no completed CI) stops every PR,
/// a repair included — the gate stops rather than guesses. `flaky` (a re-run turned a failure
/// green on the same commit) is not a stop, as [`crate::flow::tip_ci_judgment`] decides.
fn g8(s: &Snapshot) -> Result<(), Stop> {
    let stop = |reason: String| Stop { rule: "G8", reason };
    let Some(ci) = &s.tip_ci else {
        return Err(stop(
            "the tip's push CI was not read; the gate stops".into(),
        ));
    };
    let short = |x: &str| x.get(..10).unwrap_or(x).to_string();
    let decided = ci
        .decided_by
        .as_deref()
        .filter(|d| *d != ci.tip)
        .map(|d| format!(" (last completed CI on {})", short(d)))
        .unwrap_or_default();
    match ci.state.as_str() {
        "green" | "flaky" => Ok(()),
        "red" if s.tip_repair => Ok(()),
        "red" => Err(stop(format!(
            "tip {} is red{decided}: {}; only a PR with a `{}` line passes",
            short(&ci.tip),
            ci.detail,
            crate::flow::TIP_REPAIR_MARKER
        ))),
        _ => Err(stop(format!(
            "tip {} CI unknown: {}; the gate stops",
            short(&ci.tip),
            ci.detail
        ))),
    }
}

/// The gate: every rule in order, stopping at the first that fails. G8 (a red tip) comes
/// first: it stops every PR alike, so nothing else is acted on while the tip is red. On
/// success returns the verdict base and the accepted overlap.
pub fn check(s: &Snapshot) -> Result<(String, Vec<String>), Stop> {
    if s.schema_version != 1 || !is_hex40(&s.head) {
        return Err(Stop {
            rule: "input",
            reason: "schema_version 1 and a 40-hex head are required".into(),
        });
    }
    g8(s)?;
    // `judged` is the head the last verdict read: the current head, or the earlier head of an
    // accepted mechanical delta (#2833). The audit rules (G2 claims, G3 open FAILs) apply to
    // it, so a delta never carries a verdict past an unresolved FAIL on that head (D2).
    let (base, judged) = g1(s)?;
    g2(s, &judged)?;
    g9(s)?;
    g10(s)?;
    g3(s, &judged)?;
    if !s.body_check_ok {
        return Err(Stop {
            rule: "G4",
            reason: "the PR body Self-Audit checker did not pass".into(),
        });
    }
    g11(s)?;
    g5(s)?;
    let overlap = g6(s)?;
    g7(s)?;
    Ok((base, overlap))
}
