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
    /// GitHub's `author_association` of the writer. A snapshot that does not say is read as
    /// "not a member" (Issue #3052), so an old snapshot file fails closed.
    #[serde(default)]
    pub author_association: String,
}

impl crate::comment_authority::Authored for Comment {
    fn comment_id(&self) -> i64 {
        i64::try_from(self.id).unwrap_or(i64::MAX)
    }
    fn body(&self) -> &str {
        &self.body
    }
    fn association(&self) -> &str {
        &self.author_association
    }
}

/// One check run on the head: its name, conclusion (`null` while it runs) and the commit it ran
/// on (C-84: G5 counts only a run on the exact PR head; the collector reads `head_sha` from
/// GitHub, or the head it asked for).
#[derive(Clone, Debug, Deserialize, Serialize)]
#[serde(deny_unknown_fields)]
pub struct CheckRun {
    pub name: String,
    pub conclusion: Option<String>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub head_sha: Option<String>,
    /// C-99 slice 3C: the attempt of the workflow run behind a `ci` aggregate that succeeded
    /// (`run_attempt` of GitHub's workflow run; one for the first run, more for each re-run).
    /// Information only: no rule reads it, [`ci_info`] shows it from two on. Absent means one.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub attempt: Option<u32>,
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
    /// What [`Snapshot::keep_member_comments`] and the collector left out (Issue #3052). It is
    /// never part of a snapshot file.
    #[serde(skip)]
    pub ignored: crate::comment_authority::Dropped,
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
    /// Whether the PR's own change to `src/lib.rs` is only added module lines that the merged
    /// tree keeps (#2860 rule 2); then a tip change to that file is no code overlap.
    #[serde(default)]
    pub lib_mod_only_kept: bool,
    /// PR numbers whose REVIEW.md draft the user confirmed (#2614 decision).
    #[serde(default)]
    pub review_confirmed_prs: Vec<u64>,
    /// Exit status of the PR body Self-Audit checker.
    pub body_check_ok: bool,
    /// Whether the verdict's base is a commit the repository has (`git cat-file -e`). Absent
    /// means unknown, and the gate stops (audit 6850181707 N4).
    #[serde(default)]
    pub verdict_base_known: bool,
    /// Whether the verdict's base is the base-branch tip or one of its ancestors
    /// (`git merge-base --is-ancestor <base> <tip>`, #3066 P3). Collected only when the base is
    /// a known commit. Absent reads as false, and G1 stops: a base off the tip (a side branch,
    /// the PR's own old head) says nothing about the tip the PR merges onto.
    #[serde(default)]
    pub verdict_base_in_tip: bool,
    /// The full SHA of the last verdict's short base, when the collector resolved it to exactly
    /// one commit (#2860 rule 3). Absent for a full-length base or one that did not resolve.
    #[serde(default)]
    pub verdict_base_full: Option<String>,
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
    /// Other open pull requests whose own commits this one would bring into the tip (#2859),
    /// each with what G12 needs to see that a PASS covers them. Absent means not read, and the
    /// gate stops (unknown is never a pass).
    #[serde(default)]
    pub stacked_below: Option<Vec<StackedBelow>>,
    /// The audit tier the PR body declares (#2906 T, [`parse_body_tier`]). Absent means none
    /// was read, and [`effective_tier`] counts it as tier 2.
    #[serde(default)]
    pub body_tier: Option<u8>,
    /// Where the head meets the tip, 40 hex: the base a tier-0 pass names. Absent means
    /// unknown, and a tier-0 PR stops.
    #[serde(default)]
    pub merge_base: Option<String>,
}

impl Snapshot {
    /// Keeps only the comments a repository member wrote and remembers what it removed
    /// (Issue #3052). Applying it twice is harmless.
    pub fn keep_member_comments(&mut self) {
        let dropped = crate::comment_authority::retain_members(&mut self.comments);
        self.ignored.merge(dropped);
    }
}

/// One pull request below a stacked one (#2859): another open PR whose commits this PR's head
/// carries. The collector fills it; G12 decides from it.
#[derive(Clone, Debug, Deserialize, Serialize)]
#[serde(deny_unknown_fields)]
pub struct StackedBelow {
    pub pr: u64,
    /// That PR's commits this PR brings in that no other brought commit of it descends from
    /// (the newest ones), 40 hex: what a PASS on that PR must cover.
    pub brought: Vec<String>,
    /// Head and verdict word of that PR's last verdict line; absent when it has none.
    #[serde(default)]
    pub verdict_head: Option<String>,
    #[serde(default)]
    pub verdict: Option<String>,
    /// The auditor lane that last verdict names (`auditor`); absent when it names none.
    #[serde(default)]
    pub auditor: Option<String>,
    /// Sessions of the claims the brought commits name in `Moai-Claim` trailers (G9 for that
    /// PR's part).
    #[serde(default)]
    pub authors: Vec<String>,
    /// Brought commits, merges skipped, without a `Moai-Claim` trailer (G10 for that part).
    #[serde(default)]
    pub commits_without_claim: Vec<String>,
    /// Whether every `brought` commit is the verdict head or one of its ancestors; false when
    /// the verdict head is absent or not a local commit.
    pub covered: bool,
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

/// Paths that force audit tier 2 (#2906 T; user decision 6878414880, Issue #1975: "the path
/// rule forces tier 2" for HOME/project migration, secrets and keys, install scripts, the merge
/// gate and hooks). Each entry is a path prefix. Besides the decision's list this holds the
/// modules that do that work here: install (`src/install.rs`, `src/agent_install.rs`), update
/// (`src/update*.rs`, `src/self_update.rs`), HOME and Go-state migration (`src/home_*.rs`,
/// `src/migrate.rs`, `src/go_state.rs`, `src/go_settings.rs`, `src/state_root.rs` and their
/// `src/ops/` commands), keys and tokens (`src/key_vault.rs`, `src/redact.rs`,
/// `src/forge/audit_status.rs`, `src/forge/transport.rs`), hooks (`src/hook*.rs`,
/// `src/go_hooks.rs`), the gate's tip judgment (`src/flow.rs`) and the worktree secret-copy
/// rule (`src/worktree_include.rs` and the file that module reads; that file's name is joined
/// at compile time because `tests/worktree_include_guard.rs` lets only that module spell it).
/// The files that judge a pull request are here too, one entry per entry of
/// `tools/protected_checks.json` (the list `check-guard` reads from the base branch):
/// `tests/protected_paths_one_list.rs` fails when an entry of that list is not judged tier 2.
pub const TIER2_PREFIXES: &[&str] = &[
    "src/gate",
    "src/ops/gate",
    "src/flow.rs",
    "src/hook",
    "src/ops/hook",
    "src/go_hooks",
    "src/ops/go_hooks",
    ".github/",
    "tools/local_check.sh",
    "install",
    "scripts/install",
    "src/install",
    "src/agent_install",
    "src/update",
    "src/self_update",
    "src/home",
    "src/ops/home",
    "src/ops/agents_home",
    "src/migrat",
    "src/go_state",
    "src/go_settings",
    "src/ops/go_settings",
    "src/state_root",
    "src/secret",
    "src/vault",
    "src/keys",
    "src/key_vault",
    "src/redact",
    "src/forge/audit_status",
    "src/forge/transport",
    "src/worktree_include",
    concat!(".worktree", "include"),
    ".claude/",
    ".codex/",
    "tools/ci_scope.py",
    "tools/check_guard.py",
    "tools/protected_checks.json",
    "tools/repo_scans.json",
    "tools/self_check.sh",
    "tools/review_checklist.json",
    "src/audit_gate.rs",
    "REVIEW.md",
    "agents/audit.md",
    "skills/moai/workflows/audit.md",
    "AGENTS.md",
    "CLAUDE.md",
];

/// Whether a changed path may be tier 0 (#2906 T, user decision 6878414880): `docs/**`,
/// `tests/**`, or a Markdown file outside `skills/` that is not an agent instruction file
/// (`AGENTS.md`, `CLAUDE.md`). Skill Markdown is product instructions, so it is tier 1 or more.
fn tier0_safe(path: &str) -> bool {
    let name = path.rsplit('/').next().unwrap_or(path);
    path.starts_with("docs/")
        || path.starts_with("tests/")
        || (path.ends_with(".md")
            && !path.starts_with("skills/")
            && name != "AGENTS.md"
            && name != "CLAUDE.md")
}

/// The audit tier a PR body declares (#2906 T): among its first 10 lines, the first that
/// starts with `audit-tier:` decides, and it counts only as exactly `audit-tier: 0`, `1` or
/// `2` followed by the end of the line or a space (the reason). Anything else is `None`.
pub fn parse_body_tier(body: &str) -> Option<u8> {
    let line = body
        .lines()
        .take(10)
        .find(|l| l.starts_with("audit-tier:"))?;
    let line = line.strip_suffix('\r').unwrap_or(line);
    let rest = line.strip_prefix("audit-tier: ")?;
    let mut chars = rest.chars();
    let tier = match chars.next()? {
        '0' => 0,
        '1' => 1,
        '2' => 2,
        _ => return None,
    };
    matches!(chars.next(), None | Some(' ')).then_some(tier)
}

/// The tier the gate judges by (#2906 T, user decision 6878414880): the body's tier, or 2 when
/// it declared none (unknown is audited); a [`TIER2_PREFIXES`] path raises it to 2, and any
/// path that is not [`tier0_safe`] to at least 1. The path rule only raises.
pub fn effective_tier(s: &Snapshot) -> u8 {
    tier_of(s.body_tier, &s.pr_files)
}

/// [`effective_tier`] from its two inputs: the tier the body declares and the PR's files (the
/// `audit-gate` job reads them without a snapshot, SPEC-20261117-audit-gate).
pub fn tier_of(body_tier: Option<u8>, pr_files: &[String]) -> u8 {
    let mut tier = body_tier.filter(|t| *t <= 2).unwrap_or(2);
    for f in pr_files {
        if TIER2_PREFIXES.iter().any(|p| f.starts_with(p)) {
            return 2;
        }
        if !tier0_safe(f) {
            tier = tier.max(1);
        }
    }
    tier
}

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
pub fn verdict_line(body: &str) -> Option<&str> {
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

/// Like [`parse_verdict`], but the base may be a short SHA of 7 to 39 lowercase hex digits
/// (#2860 rule 3, shell gate_merge.sh). The head stays 40 hex. Only a base the collector
/// resolved to exactly one commit is accepted, through [`verdict_of`].
pub fn parse_verdict_short_base(line: &str) -> Option<(String, String, String)> {
    let h = line.rfind(" head ")? + 6;
    let head = line.get(h..h + 40)?;
    let rest = line.get(h + 40..)?.strip_prefix(" base ")?;
    let (base, verdict) = rest.split_once(" — verdict: ")?;
    let short = (7..40).contains(&base.len())
        && base
            .bytes()
            .all(|b| b.is_ascii_digit() || (b'a'..=b'f').contains(&b));
    if !is_hex40(head) || !short || verdict.is_empty() || verdict.contains(' ') {
        return None;
    }
    Some((head.into(), base.into(), verdict.into()))
}

/// The verdict a line carries for this snapshot: a full line, or one whose short base the
/// collector resolved to exactly one commit (`verdict_base_full`, #2860 rule 3), read with
/// that full SHA as its base.
pub fn verdict_of(s: &Snapshot, line: &str) -> Option<(String, String, String)> {
    if let Some(v) = parse_verdict(line) {
        return Some(v);
    }
    let (head, short, verdict) = parse_verdict_short_base(line)?;
    let full = s.verdict_base_full.as_ref()?;
    (is_hex40(full) && full.starts_with(&short)).then(|| (head, full.clone(), verdict))
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
    // #2860 scope add 6872037464 (shell owner_ok.py, #2831): a comment whose first line is
    // `moai-audit-withdraw: lane=<lane>` or `moai-audit-claim-withdraw: lane=<lane>` gives the
    // PR up for that lane — its claims posted before that comment no longer count.
    for c in comments {
        let Some(lane) = lane_withdrawal(&c.body) else {
            continue;
        };
        out.retain(|cl| !(cl.1 == lane && cl.0 < c.id));
    }
    out.sort();
    out
}

/// The lane a whole-lane withdrawal names, from a comment's first line
/// `moai-audit-withdraw: lane=<lane>` or `moai-audit-claim-withdraw: lane=<lane>`.
pub fn lane_withdrawal(body: &str) -> Option<String> {
    let first = body.lines().next()?;
    let rest = first
        .strip_prefix("moai-audit-withdraw: lane=")
        .or_else(|| first.strip_prefix("moai-audit-claim-withdraw: lane="))?;
    let lane: String = rest
        .chars()
        .take_while(|ch| ch.is_ascii_alphanumeric() || *ch == '-')
        .collect();
    (!lane.is_empty()).then_some(lane)
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
    let (head, _, verdict) = verdict_of(s, line)?;
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
    let (head, base, verdict) = verdict_of(s, line).ok_or_else(|| {
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
    // #3066 P3: a known base off the tip (a side branch, the PR's own old head) is no verdict
    // about the tip the PR merges onto.
    if !s.verdict_base_in_tip {
        return Err(stop(format!(
            "the verdict base {base} is not an ancestor of the base-branch tip"
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

/// G12 (#2859): a pull request stacked on another open one merges only when the PR below has
/// a PASS on the commits this one would bring in — the head it carries or a newer head of that
/// PR — by an auditor who wrote none of them, and every brought commit names its claim (D1:
/// the same verdict, G9 and G10 rules). A PR below that already merged is not open and not
/// listed (D2). Unknown stops the gate rather than passes it.
fn g12(s: &Snapshot) -> Result<(), Stop> {
    let stop = |reason: String| Stop {
        rule: "G12",
        reason,
    };
    let below = s
        .stacked_below
        .as_ref()
        .ok_or_else(|| stop("the open pull requests below this one were not read".into()))?;
    for b in below {
        let carried = b
            .brought
            .iter()
            .map(|c| c.get(..10).unwrap_or(c.as_str()))
            .collect::<Vec<_>>()
            .join(", ");
        let at = |why: String| stop(format!("PR #{} below carries {carried}: {why}", b.pr));
        let head = b
            .verdict_head
            .as_deref()
            .ok_or_else(|| at("it has no AUDIT verdict".into()))?;
        if b.verdict.as_deref() != Some("PASS") {
            return Err(at(format!(
                "its last verdict, on {}, is {}",
                head.get(..10).unwrap_or(head),
                b.verdict.as_deref().unwrap_or("unreadable")
            )));
        }
        if b.brought.is_empty() || !b.covered {
            return Err(at(format!(
                "its PASS is on {}, which does not hold the commits this PR brings in",
                head.get(..10).unwrap_or(head)
            )));
        }
        let who = b
            .auditor
            .as_deref()
            .ok_or_else(|| at("its last verdict names no auditor lane".into()))?;
        if let Some(a) = b.authors.iter().find(|a| a.eq_ignore_ascii_case(who)) {
            return Err(at(format!(
                "its PASS is by {who}, and {a} wrote commits of it (Moai-Claim trailer)"
            )));
        }
        if !b.commits_without_claim.is_empty() {
            let short: Vec<&str> = b
                .commits_without_claim
                .iter()
                .map(|c| c.get(..10).unwrap_or(c.as_str()))
                .collect();
            return Err(at(format!(
                "{} of its commits carry no `Moai-Claim` trailer: {}",
                short.len(),
                short.join(", ")
            )));
        }
    }
    Ok(())
}

/// G3: a FAIL on the current head blocks unless every FAIL there and the later last PASS
/// come from the same auditor (it reversed itself).
fn g3(s: &Snapshot, judged: &str) -> Result<(), Stop> {
    let mut on_head: Vec<(u64, String, String)> = s
        .comments
        .iter()
        .filter_map(|c| {
            let (h, _, v) = verdict_of(s, verdict_line(&c.body)?)?;
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

/// G3 at tier 0 (C-84, leader decision 2026-11-17: a posted FAIL stops the merge at every
/// tier): a tier-0 PR has no PASS that supersedes an older head's FAIL, so a FAIL verdict on
/// any head stops it until the same auditor posts a later PASS (a higher comment id).
fn g3_tier0(s: &Snapshot) -> Result<(), Stop> {
    let verdicts: Vec<(u64, String, String)> = s
        .comments
        .iter()
        .filter_map(|c| {
            let (_, _, v) = verdict_of(s, verdict_line(&c.body)?)?;
            Some((c.id, v, auditor(&c.body).unwrap_or_default()))
        })
        .collect();
    let open = verdicts
        .iter()
        .filter(|f| f.1 == "FAIL")
        .filter(|f| {
            !verdicts
                .iter()
                .any(|p| p.1 == "PASS" && !f.2.is_empty() && p.2 == f.2 && p.0 > f.0)
        })
        .count();
    if open == 0 {
        Ok(())
    } else {
        Err(Stop {
            rule: "G3",
            reason: format!(
                "audit tier 0: {open} FAIL verdict(s) on this pull request and no later PASS by the same auditor"
            ),
        })
    }
}

/// `YYYY-MM-DDTHH:MM:SSZ`: digits and separators only, 20 bytes, nothing after the `Z`.
fn is_utc_seconds(at: &str) -> bool {
    at.len() == 20
        && at.bytes().enumerate().all(|(i, b)| match i {
            4 | 7 => b == b'-',
            10 => b == b'T',
            13 | 16 => b == b':',
            19 => b == b'Z',
            _ => b.is_ascii_digit(),
        })
}

/// `(head, result)` from the line `tools/local_check.sh` prints last, in either of its two
/// forms (#2906 S5a, #3066 P1), from column 0, fields in this order:
///
/// - the short form `local-check: head=<40 lowercase hex> result=<pass|fail>
///   cmd=tools/local_check.sh at=<YYYY-MM-DDTHH:MM:SSZ>` (what the script prints today);
/// - the long form read by [`parse_local_check_evidence`] (accepted here so the gate keeps
///   working when the script starts to print it).
///
/// Anything else (a short or long head, uppercase, no `at=`, another command, a `dirty=1`
/// field) is `None`.
pub fn parse_local_check(line: &str) -> Option<(&str, &str)> {
    if let Some(short) = parse_local_check_short(line) {
        return Some(short);
    }
    let e = parse_local_check_evidence(line)?;
    Some((e.head, e.result))
}

fn parse_local_check_short(line: &str) -> Option<(&str, &str)> {
    let rest = line.strip_prefix("local-check: head=")?;
    let (head, rest) = rest.split_once(' ')?;
    let (result, rest) = rest.strip_prefix("result=")?.split_once(' ')?;
    let at = rest.strip_prefix("cmd=tools/local_check.sh at=")?;
    (is_hex40(head) && matches!(result, "pass" | "fail") && is_utc_seconds(at))
        .then_some((head, result))
}

/// The fields of a long-form `local-check:` line (#3066 P1), all already validated by
/// [`parse_local_check_evidence`].
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct LocalCheckEvidence<'a> {
    /// The checked commit: 40 lowercase hex.
    pub head: &'a str,
    /// `pass` or `fail`.
    pub result: &'a str,
    /// The base the run merged the head onto: 40 lowercase hex, or `None` when the line says
    /// `base=none` (no base was known; only on a `fail`).
    pub base: Option<&'a str>,
    /// The tree of that merge: 40 lowercase hex, or `None` when the line says `tree=none` (no
    /// tree, for example a merge conflict; only on a `fail`).
    pub tree: Option<&'a str>,
    /// The checking script's identity: 40 lowercase hex.
    pub script: &'a str,
    /// The toolchain token: 1 to 32 characters from `[A-Za-z0-9._+-]`.
    pub rustc: &'a str,
    /// The run id: 12 lowercase hex.
    pub run: &'a str,
    /// The UTC time, `YYYY-MM-DDTHH:MM:SSZ`.
    pub at: &'a str,
}

/// `Some(None)` for the word `none`, `Some(Some(hex))` for 40 lowercase hex, else `None`.
fn hex40_or_none(s: &str) -> Option<Option<&str>> {
    if s == "none" {
        Some(None)
    } else {
        is_hex40(s).then_some(Some(s))
    }
}

/// One `key=value` field of a long-form line: the value up to the next single space, and
/// what follows that space.
fn local_check_field<'a>(rest: &'a str, key: &str) -> Option<(&'a str, &'a str)> {
    rest.strip_prefix(key)?.split_once(' ')
}

/// Reads the long form of the `local-check:` line (#3066 P1), from column 0, single spaces,
/// no trailing text, fields in exactly this order: `local-check: head=<40hex> result=<pass|fail>
/// base=<40hex|none> tree=<40hex|none> script=<40hex> rustc=<token> run=<12hex>
/// cmd=tools/local_check.sh at=<YYYY-MM-DDTHH:MM:SSZ>`. All hex is lowercase. A `pass` line
/// has a real base and a real tree; `tree=none` and `base=none` only on a `fail`, and
/// `base=none` needs `tree=none`. Anything else, including the short form and a `dirty=1`
/// field, is `None`.
pub fn parse_local_check_evidence(line: &str) -> Option<LocalCheckEvidence<'_>> {
    let rest = line.strip_prefix("local-check: head=")?;
    let (head, rest) = rest.split_once(' ')?;
    let (result, rest) = local_check_field(rest, "result=")?;
    let (base, rest) = local_check_field(rest, "base=")?;
    let (tree, rest) = local_check_field(rest, "tree=")?;
    let (script, rest) = local_check_field(rest, "script=")?;
    let (rustc, rest) = local_check_field(rest, "rustc=")?;
    let (run, rest) = local_check_field(rest, "run=")?;
    let at = rest.strip_prefix("cmd=tools/local_check.sh at=")?;
    let (base, tree) = (hex40_or_none(base)?, hex40_or_none(tree)?);
    let rustc_ok = (1..=32).contains(&rustc.len())
        && rustc
            .bytes()
            .all(|b| b.is_ascii_alphanumeric() || matches!(b, b'.' | b'_' | b'+' | b'-'));
    let run_ok = run.len() == 12
        && run
            .bytes()
            .all(|b| b.is_ascii_digit() || (b'a'..=b'f').contains(&b));
    let evidence_ok = match (base, tree) {
        (Some(_), Some(_)) => true,
        // A conflict: the base is known, the merge has no tree.
        (Some(_), None) => result == "fail",
        (None, None) => result == "fail",
        (None, Some(_)) => false,
    };
    (is_hex40(head)
        && matches!(result, "pass" | "fail")
        && evidence_ok
        && is_hex40(script)
        && rustc_ok
        && run_ok
        && is_utc_seconds(at))
    .then_some(LocalCheckEvidence {
        head,
        result,
        base,
        tree,
        script,
        rustc,
        run,
        at,
    })
}

/// The optional `local-check:` line for this head (C-84: evidence `gate-merge` records; G5
/// reads `ci`): `(comment id, line, result)` of the last line for this head (highest comment
/// id, last line within it), or `None` when there is none.
pub fn local_check_of(s: &Snapshot) -> Option<(u64, &str, &str)> {
    s.comments
        .iter()
        .flat_map(|c| {
            c.body.lines().enumerate().filter_map(move |(i, l)| {
                let (head, result) = parse_local_check(l)?;
                (head == s.head).then_some((c.id, i, l, result))
            })
        })
        .max_by_key(|&(id, i, _, _)| (id, i))
        .map(|(id, _, l, r)| (id, l, r))
}

/// The check runs the merge reads: the `ci` aggregate (`rc-ci` is its earlier name), its jobs
/// (`ci / …`, `rc-ci / …`) and `windows-build`.
fn watched_check(name: &str) -> bool {
    name == "ci"
        || name == "rc-ci"
        || name.starts_with("ci / ")
        || name.starts_with("rc-ci / ")
        || name.starts_with("windows-build")
}

/// Whether a finished watched run counts as failed: anything but `success`, except a skipped
/// `windows-build` (it runs after the merge, U10).
fn check_failed(c: &CheckRun) -> bool {
    match c.conclusion.as_deref() {
        Some("success") | None => false,
        Some("skipped") => !c.name.starts_with("windows-build"),
        Some(_) => true,
    }
}

/// G5 (C-84, user decision 2026-11-17 "CI 성공 + PASS", which replaces user decisions
/// 6877084436 and 6878414880): the PR head's GitHub `ci` succeeded on this exact head. A
/// watched run on this head that failed (or was cancelled, timed out, …) stops it, so does a
/// `ci` still running, a `ci` that ran only on another head or names none, and no `ci` at all.
/// A `local-check:` line is optional evidence ([`local_check_of`]) and never decides G5.
fn g5(s: &Snapshot) -> Result<(), Stop> {
    let stop = |reason: String| Stop { rule: "G5", reason };
    let on_head: Vec<&CheckRun> = s
        .checks
        .iter()
        .filter(|c| watched_check(&c.name) && c.head_sha.as_deref() == Some(s.head.as_str()))
        .collect();
    if let Some(bad) = on_head.iter().find(|c| check_failed(c)) {
        return Err(stop(format!(
            "`{}` failed on this head ({})",
            bad.name,
            bad.conclusion.as_deref().unwrap_or("")
        )));
    }
    if let Some(c) = on_head.iter().find(|c| c.conclusion.is_none()) {
        return Err(stop(format!("`{}` is still running on this head", c.name)));
    }
    let aggregate = |c: &&CheckRun| c.name == "ci" || c.name == "rc-ci";
    if on_head
        .iter()
        .any(|c| aggregate(c) && c.conclusion.as_deref() == Some("success"))
    {
        return Ok(());
    }
    let elsewhere: Vec<&CheckRun> = s.checks.iter().filter(aggregate).collect();
    if let Some(c) = elsewhere.iter().find(|c| c.head_sha.is_none()) {
        return Err(stop(format!(
            "the `{}` check run names no head; the gate cannot tell it is this one",
            c.name
        )));
    }
    if let Some(c) = elsewhere.first() {
        let other = c.head_sha.as_deref().unwrap_or("");
        return Err(stop(format!(
            "`{}` ran on {}, not on this head {}",
            c.name,
            other.get(..10).unwrap_or(other),
            &s.head[..10]
        )));
    }
    Err(stop("no `ci` check run on this head".into()))
}

/// The head's GitHub CI, as information only (#2906 S5b): `ci=failure` when a `ci`, `rc-ci`,
/// `ci / …`, `rc-ci / …` or `windows-build` run ended other than success (`windows-build` may
/// be `skipped`: it runs after the merge, U10), else `ci=pending` while one runs or before a
/// `ci`/`rc-ci` aggregate succeeded, `ci=success`, or `ci=none` when there is no such run.
pub fn ci_info(s: &Snapshot) -> String {
    let runs: Vec<&CheckRun> = s.checks.iter().filter(|c| watched_check(&c.name)).collect();
    let failed = runs.iter().any(|c| check_failed(c));
    // `rc-ci` is the aggregate's earlier name (flow.rs `CI_CHECK_NAMES`, audit N1).
    let aggregate = runs.iter().any(|c| {
        (c.name == "ci" || c.name == "rc-ci") && c.conclusion.as_deref() == Some("success")
    });
    let state = if runs.is_empty() {
        "none"
    } else if failed {
        "failure"
    } else if runs.iter().any(|c| c.conclusion.is_none()) || !aggregate {
        "pending"
    } else {
        "success"
    };
    // C-99 slice 3C: a `ci` that went green only on a re-run says so, so a person reading the line
    // does not take the pass for a first-run pass. G5 does not read this.
    let rerun = runs
        .iter()
        .filter(|c| {
            (c.name == "ci" || c.name == "rc-ci") && c.conclusion.as_deref() == Some("success")
        })
        .filter_map(|c| c.attempt)
        .find(|n| *n >= 2);
    match (state, rerun) {
        ("success", Some(n)) => format!("ci={state} attempt={n}"),
        _ => format!("ci={state}"),
    }
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
        } else if f == "src/lib.rs" && s.lib_mod_only_kept {
            // #2860 rule 2: only module lines added on the PR side, all kept in the merge.
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
///
/// The audit tier (#2906 T, user decision 6878414880) is decided before any rule
/// ([`effective_tier`]). Tier 0 needs no audit: G1 (the PASS) and G2 (the audit claim) are not
/// applied, G9 only needs the authors to be known (there is no auditor to compare), and the
/// base returned is the merge base. G3 still runs on the current head, so an auditor's open
/// FAIL there is not passed over. Every other rule runs as for tiers 1 and 2, which are
/// unchanged.
pub fn check(s: &Snapshot) -> Result<(String, Vec<String>), Stop> {
    if s.schema_version != 1 || !is_hex40(&s.head) {
        return Err(Stop {
            rule: "input",
            reason: "schema_version 1 and a 40-hex head are required".into(),
        });
    }
    let tier = effective_tier(s);
    g8(s)?;
    let base = if tier == 0 {
        let base = s.merge_base.clone().filter(|b| is_hex40(b)).ok_or(Stop {
            rule: "tier",
            reason: "audit tier 0: merge base unknown; the gate stops".into(),
        })?;
        if s.pr_authors.is_none() {
            return Err(Stop {
                rule: "G9",
                reason: "the authors of the PR's commits are unknown".into(),
            });
        }
        g10(s)?;
        g3(s, &s.head)?;
        g3_tier0(s)?;
        base
    } else {
        // `judged` is the head the last verdict read: the current head, or the earlier head of
        // an accepted mechanical delta (#2833). The audit rules (G2 claims, G3 open FAILs)
        // apply to it, so a delta never carries a verdict past an unresolved FAIL on that head
        // (D2).
        let (base, judged) = g1(s)?;
        g2(s, &judged)?;
        g9(s)?;
        g10(s)?;
        g3(s, &judged)?;
        base
    };
    if !s.body_check_ok {
        return Err(Stop {
            rule: "G4",
            reason: "the PR body Self-Audit checker did not pass".into(),
        });
    }
    g11(s)?;
    g12(s)?;
    g5(s)?;
    let overlap = g6(s)?;
    g7(s)?;
    Ok((base, overlap))
}

#[cfg(test)]
mod tests {
    use super::{
        parse_body_tier, parse_local_check, parse_local_check_evidence, tier0_safe,
        LocalCheckEvidence,
    };

    /// #2906 T: `audit-tier: <0|1|2>` at column 0, then the end of the line or a space.
    #[test]
    fn a_body_tier_is_exactly_one_digit_after_the_prefix() {
        for (body, want) in [
            ("audit-tier: 0", Some(0)),
            ("audit-tier: 1\nrest", Some(1)),
            ("audit-tier: 2 src/gate.rs changed", Some(2)),
            ("audit-tier: 0 docs only\r\nbody", Some(0)),
            ("audit-tier: 0\r\n", Some(0)),
            ("title\naudit-tier: 0 docs", Some(0)),
            ("audit-tier: 3", None),
            ("audit-tier: 9 x", None),
            ("audit-tier:0", None),
            ("audit-tier:  0", None),
            ("audit-tier: 01", None),
            ("audit-tier: 0x", None),
            ("audit-tier: 0,docs", None),
            ("audit-tier: ", None),
            ("audit-tier:", None),
            ("Audit-tier: 0", None),
            (" audit-tier: 0", None),
            ("> audit-tier: 0", None),
            ("audit-tier 0", None),
            ("", None),
        ] {
            assert_eq!(parse_body_tier(body), want, "{body:?}");
        }
    }

    /// The first `audit-tier:` line decides, even when it is malformed and a later one is
    /// not; only the first 10 lines are read (line 10 is read, line 11 is not).
    #[test]
    fn the_first_tier_line_in_the_first_ten_lines_decides() {
        assert_eq!(parse_body_tier("audit-tier: 2\naudit-tier: 0"), Some(2));
        assert_eq!(parse_body_tier("audit-tier: 3\naudit-tier: 0"), None);
        assert_eq!(parse_body_tier("audit-tier:0\naudit-tier: 0"), None);
        let at = |n: usize| format!("{}audit-tier: 0\n", "x\n".repeat(n - 1));
        assert_eq!(parse_body_tier(&at(10)), Some(0));
        assert_eq!(parse_body_tier(&at(11)), None);
        // In the middle of a line it is not a tier line.
        assert_eq!(parse_body_tier("see audit-tier: 0"), None);
    }

    /// Tier-0-safe paths: documents, tests, Markdown outside `skills/` but for agent
    /// instruction files.
    #[test]
    fn tier0_safe_paths() {
        for p in [
            "docs/a.md",
            "docs/img/y.png",
            "tests/a.rs",
            "README.md",
            "a/b.md",
        ] {
            assert!(tier0_safe(p), "{p}");
        }
        for p in [
            "skills/a.md",
            "skills/moai/SKILL.md",
            "CLAUDE.md",
            "AGENTS.md",
            "x/CLAUDE.md",
            "src/a.rs",
            "docs",
            "tests",
            "a.mdx",
            "Cargo.toml",
        ] {
            assert!(!tier0_safe(p), "{p}");
        }
    }

    const AT: &str = "at=2026-11-14T12:00:00Z";

    fn line(head: &str, result: &str) -> String {
        format!("local-check: head={head} result={result} cmd=tools/local_check.sh {AT}")
    }

    /// #2906 S5b: 40 lowercase hex is read (N); 39 and 41 digits, uppercase or a non-hex
    /// letter are not, whatever head the snapshot has.
    #[test]
    fn a_local_check_head_is_exactly_40_lowercase_hex() {
        let h = "0123456789abcdef0123456789abcdef01234567";
        assert_eq!(parse_local_check(&line(h, "pass")), Some((h, "pass")));
        assert_eq!(parse_local_check(&line(h, "fail")), Some((h, "fail")));
        assert_eq!(parse_local_check(&line(&h[..39], "pass")), None);
        assert_eq!(parse_local_check(&line(&format!("{h}8"), "pass")), None);
        assert_eq!(parse_local_check(&line(&h.to_uppercase(), "pass")), None);
        assert_eq!(parse_local_check(&line(&h.replace('a', "g"), "pass")), None);
        assert_eq!(parse_local_check(&line(h, "ok")), None);
        assert_eq!(parse_local_check(&line(h, "")), None);
    }

    /// The time is `YYYY-MM-DDTHH:MM:SSZ` and nothing follows it.
    /// C-84: the local-check line `gate-merge` records is the last one for this head — the
    /// highest comment id wins, then the later line inside a comment; another head's line and a
    /// malformed one are not read. (G5 no longer reads it; this pins what the record cites.)
    #[test]
    fn the_recorded_local_check_line_is_the_last_one_for_this_head() {
        let h = "0123456789abcdef0123456789abcdef01234567";
        let other = "fedcba9876543210fedcba9876543210fedcba98";
        let snap = |comments: serde_json::Value| -> super::Snapshot {
            serde_json::from_value(serde_json::json!({
                "schema_version": 1, "pr": 9, "head": h, "comments": comments, "checks": [],
                "pr_files": [], "tip_files_since_verdict_base": [], "merge_clean": true,
                "body_check_ok": true
            }))
            .unwrap()
        };
        let c = |id: u64, body: String| serde_json::json!({"id": id, "author_association": "MEMBER", "body": body});
        let s = snap(serde_json::json!([
            c(310, line(h, "fail")),
            c(300, line(h, "pass")),
            c(320, line(other, "pass"))
        ]));
        assert_eq!(
            super::local_check_of(&s).map(|x| (x.0, x.2)),
            Some((310, "fail"))
        );
        let s = snap(serde_json::json!([c(
            300,
            format!("{}\n{}", line(h, "fail"), line(h, "pass"))
        )]));
        assert_eq!(super::local_check_of(&s).map(|x| x.2), Some("pass"));
        let s = snap(serde_json::json!([c(300, format!(" {}", line(h, "pass")))]));
        assert!(super::local_check_of(&s).is_none());
    }

    #[test]
    fn a_local_check_time_is_utc_seconds() {
        let h = "0123456789abcdef0123456789abcdef01234567";
        let with_at = |at: &str| {
            format!("local-check: head={h} result=pass cmd=tools/local_check.sh at={at}")
        };
        assert!(parse_local_check(&with_at("2026-11-14T12:00:00Z")).is_some());
        for bad in [
            "",
            "2026-11-14T12:00:00",
            "2026-11-14T12:00:00Z ",
            "2026-11-14T12:00:0Z",
            "2026-11-14 12:00:00Z",
            "2026/11/14T12:00:00Z",
            "2026-11-14T12-00-00Z",
            "2026-11-14T12:00:0aZ",
        ] {
            assert_eq!(parse_local_check(&with_at(bad)), None, "{bad:?}");
        }
        assert_eq!(
            parse_local_check(&format!(
                "local-check: head={h} result=pass cmd=tools/local_check.sh"
            )),
            None
        );
    }

    const HEAD: &str = "0123456789abcdef0123456789abcdef01234567";
    const BASE: &str = "1123456789abcdef0123456789abcdef01234567";
    const TREE: &str = "2123456789abcdef0123456789abcdef01234567";
    const SCRIPT: &str = "3123456789abcdef0123456789abcdef01234567";

    /// A long-form line from its variable fields.
    fn long(
        head: &str,
        result: &str,
        base: &str,
        tree: &str,
        script: &str,
        rustc: &str,
        run: &str,
    ) -> String {
        format!(
            "local-check: head={head} result={result} base={base} tree={tree} script={script} rustc={rustc} run={run} cmd=tools/local_check.sh {AT}"
        )
    }

    fn good(result: &str) -> String {
        long(HEAD, result, BASE, TREE, SCRIPT, "1.91.0", "0123456789ab")
    }

    /// #3066 P1: every field of a long-form line is returned as written, and
    /// `parse_local_check` gives `(head, result)` for it, not another field.
    #[test]
    fn a_long_local_check_line_returns_its_fields() {
        let l = good("pass");
        assert_eq!(
            parse_local_check_evidence(&l),
            Some(LocalCheckEvidence {
                head: HEAD,
                result: "pass",
                base: Some(BASE),
                tree: Some(TREE),
                script: SCRIPT,
                rustc: "1.91.0",
                run: "0123456789ab",
                at: "2026-11-14T12:00:00Z",
            })
        );
        assert_eq!(parse_local_check(&l), Some((HEAD, "pass")));
        assert_eq!(parse_local_check(&good("fail")), Some((HEAD, "fail")));
        // The short form is not a long-form line, and the long form is not a short one.
        assert_eq!(parse_local_check_evidence(&line(HEAD, "pass")), None);
    }

    /// #3066 P1: `none` is read as no value, only on a failed run.
    #[test]
    fn a_long_local_check_none_is_only_for_a_failed_run() {
        let ok = |base: &str, tree: &str, result: &str| {
            parse_local_check_evidence(&long(HEAD, result, base, tree, SCRIPT, "x", "0123456789ab"))
                .map(|e| (e.base.map(str::to_string), e.tree.map(str::to_string)))
        };
        assert_eq!(ok(BASE, "none", "fail"), Some((Some(BASE.into()), None)));
        assert_eq!(ok("none", "none", "fail"), Some((None, None)));
        assert_eq!(
            ok(BASE, TREE, "pass"),
            Some((Some(BASE.into()), Some(TREE.into())))
        );
        assert_eq!(ok(BASE, "none", "pass"), None);
        assert_eq!(ok("none", "none", "pass"), None);
        assert_eq!(ok("none", TREE, "fail"), None);
        assert_eq!(ok("none", TREE, "pass"), None);
        assert_eq!(ok("NONE", "none", "fail"), None);
        assert_eq!(ok(BASE, "None", "fail"), None);
    }

    /// #3066 P1: each hex field is exactly its length (N passes, N-1 and N+1 do not), lowercase
    /// hex only.
    #[test]
    fn a_long_local_check_hex_fields_have_an_exact_length() {
        let run = "0123456789ab";
        let h = |s: &str, n: usize| -> String { s.chars().cycle().take(n).collect() };
        for n in [0usize, 39, 41, 80] {
            let v = h(HEAD, n);
            assert_eq!(
                parse_local_check_evidence(&long(&v, "pass", BASE, TREE, SCRIPT, "x", run)),
                None,
                "head {n}"
            );
            assert_eq!(
                parse_local_check_evidence(&long(HEAD, "pass", &v, TREE, SCRIPT, "x", run)),
                None,
                "base {n}"
            );
            assert_eq!(
                parse_local_check_evidence(&long(HEAD, "pass", BASE, &v, SCRIPT, "x", run)),
                None,
                "tree {n}"
            );
            assert_eq!(
                parse_local_check_evidence(&long(HEAD, "pass", BASE, TREE, &v, "x", run)),
                None,
                "script {n}"
            );
        }
        for n in [0usize, 11, 13, 40] {
            let v = h(run, n);
            assert_eq!(
                parse_local_check_evidence(&long(HEAD, "pass", BASE, TREE, SCRIPT, "x", &v)),
                None,
                "run {n}"
            );
        }
        for bad in ["A", "g", "G"] {
            let up = format!("{}{bad}", &HEAD[1..]);
            assert_eq!(
                parse_local_check_evidence(&long(&up, "pass", BASE, TREE, SCRIPT, "x", run)),
                None,
                "{up}"
            );
            assert_eq!(
                parse_local_check_evidence(&long(HEAD, "pass", &up, TREE, SCRIPT, "x", run)),
                None,
                "{up}"
            );
            assert_eq!(
                parse_local_check_evidence(&long(HEAD, "pass", BASE, &up, SCRIPT, "x", run)),
                None,
                "{up}"
            );
            assert_eq!(
                parse_local_check_evidence(&long(HEAD, "pass", BASE, TREE, &up, "x", run)),
                None,
                "{up}"
            );
            let bad_run = format!("{}{bad}", &run[1..]);
            assert_eq!(
                parse_local_check_evidence(&long(HEAD, "pass", BASE, TREE, SCRIPT, "x", &bad_run)),
                None,
                "{bad_run}"
            );
        }
    }

    /// #3066 P1: `rustc` is 1 to 32 characters from `[A-Za-z0-9._+-]`.
    #[test]
    fn a_long_local_check_rustc_is_a_short_token() {
        let with = |r: &str| {
            parse_local_check_evidence(&long(HEAD, "pass", BASE, TREE, SCRIPT, r, "0123456789ab"))
                .map(|e| e.rustc.to_string())
        };
        assert_eq!(
            with("1.91.0-nightly+Ab_9").as_deref(),
            Some("1.91.0-nightly+Ab_9")
        );
        assert!(with("x").is_some());
        assert!(with(&"a".repeat(32)).is_some());
        assert!(with(&"a".repeat(33)).is_none());
        assert!(with("").is_none());
        for bad in ["a b", "a=b", "a/b", "a:b", "a\u{e9}", "a,b", "a\tb"] {
            assert!(with(bad).is_none(), "{bad:?}");
        }
    }

    /// #3066 P1: no `dirty=1`, no extra or reordered or missing field, no second space, no
    /// trailing text, and the time rule is the short form's.
    #[test]
    fn a_long_local_check_line_is_strict() {
        let l = good("pass");
        assert!(parse_local_check_evidence(&l).is_some());
        for bad in [
            l.replacen(" cmd=", " dirty=1 cmd=", 1),
            l.replacen(" base=", " dirty=1 base=", 1),
            format!("{l} dirty=1"),
            format!("{l} extra"),
            format!("{l} "),
            l.replacen(" tree=", "  tree=", 1),
            l.replacen("local-check: ", "local-check:  ", 1),
            format!(" {l}"),
            l.replacen("cmd=tools/local_check.sh", "cmd=tools/other.sh", 1),
            l.replacen("at=2026-11-14T12:00:00Z", "at=2026-11-14T12:00:00", 1),
            l.replacen("at=2026-11-14T12:00:00Z", "at=2026-11-14 12:00:00Z", 1),
            l.replacen("result=pass", "result=ok", 1),
            l.replacen(
                &format!("base={BASE} tree={TREE}"),
                &format!("tree={TREE} base={BASE}"),
                1,
            ),
        ] {
            assert_eq!(parse_local_check_evidence(&bad), None, "{bad}");
            assert_eq!(parse_local_check(&bad), None, "{bad}");
        }
        // One field missing at a time.
        for key in [
            "result", "base", "tree", "script", "rustc", "run", "cmd", "at",
        ] {
            let dropped: Vec<&str> = l
                .split(' ')
                .filter(|p| !p.starts_with(&format!("{key}=")))
                .collect();
            assert_eq!(
                parse_local_check_evidence(&dropped.join(" ")),
                None,
                "{key}"
            );
        }
    }
}
