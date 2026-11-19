//! BOARD 04 link view (SPEC-20261120-board-bundle-link): joins a bundle's run record, its source
//! and target SHAs and the pull request's exact head from stored observation rows, and shows the
//! two audit layers (the SPEC audit of the lane, the bundle pull request audit) apart. "No longer
//! in effect" is computed here at read time and never stored: a pass is in effect only while the
//! head it names is the current head of its own layer.
//!
//! Pure functions over rows already read: no file, no database, no process, no network. It is
//! written with standard-library and serde paths only (no `crate::`) so that the slice can be
//! built and tested before the module is declared from `src/lib.rs`.

use serde::Serialize;
use serde_json::Value;
use std::collections::BTreeMap;

/// A stored observation row as the read side hands it over (the fields this view needs).
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct StoredRow {
    pub seq: i64,
    pub kind: String,
    pub scope: String,
    pub observed_at: i64,
    pub content: String,
}

/// A verdict word of an audit.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize)]
#[serde(rename_all = "UPPERCASE")]
pub enum Verdict {
    Pass,
    Fail,
}

/// One lane audit read from a verdict note of the ledger (read-only, by the caller).
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct LaneAudit {
    pub spec: String,
    pub head: String,
    pub verdict: Verdict,
}

/// What the lane layer is compared with: the current head of each SPEC's lane, and the audits.
#[derive(Debug, Clone, Default, PartialEq, Eq)]
pub struct LaneInput {
    pub heads: BTreeMap<String, String>,
    pub audits: Vec<LaneAudit>,
}

/// One verdict and the head it names; `in_effect` only while that head is the layer's current one.
#[derive(Debug, Clone, PartialEq, Eq, Serialize)]
pub struct AuditEntry {
    pub verdict: Verdict,
    /// `None` when the head could not be read: such a pass is never in effect.
    pub head: Option<String>,
    pub in_effect: bool,
}

/// What a stored run record says about its own command (exit code only, no log text).
#[derive(Debug, Clone, PartialEq, Eq, Serialize)]
pub struct CommandResult {
    pub check: String,
    pub exit: i64,
}

/// A run's standing against the pull request's current head.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize)]
#[serde(rename_all = "snake_case")]
pub enum Effect {
    /// The pull request still shows the run's source SHA as its head.
    Standing,
    /// The pull request's head moved on: the run is history, no longer in effect.
    Lapsed,
    NoPullRequest,
}

/// The pull request side of one bundle: its current values and the bundle pull request audit layer.
#[derive(Debug, Clone, PartialEq, Eq, Serialize)]
pub struct PrView {
    pub number: u64,
    pub state: String,
    pub head: String,
    pub audit: Vec<AuditEntry>,
}

/// The lane audit layer of one SPEC of the bundle.
#[derive(Debug, Clone, PartialEq, Eq, Serialize)]
pub struct LaneLayer {
    pub spec: String,
    pub head: Option<String>,
    pub audit: Vec<AuditEntry>,
}

/// One bundle: run record, SHAs, command results, pull request and the two audit layers.
#[derive(Debug, Clone, PartialEq, Eq, Serialize)]
pub struct Bundle {
    pub branch: String,
    pub source_sha: String,
    pub target_sha: String,
    pub commands: Vec<CommandResult>,
    pub not_run: Vec<String>,
    pub specs: Vec<String>,
    pub run_effect: Effect,
    pub pr: Option<PrView>,
    pub lane: Vec<LaneLayer>,
}

/// Whether the link could be read.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize)]
#[serde(rename_all = "snake_case")]
pub enum State {
    Known,
    Unknown,
}

/// The whole view.
#[derive(Debug, Clone, PartialEq, Eq, Serialize)]
pub struct LinkView {
    pub state: State,
    pub reason: Option<String>,
    pub unreadable: Vec<String>,
    pub bundles: Vec<Bundle>,
}

/// 40 lowercase hex characters.
fn is_sha(s: &str) -> bool {
    s.len() == 40
        && s.bytes()
            .all(|b| b.is_ascii_digit() || (b'a'..=b'f').contains(&b))
}

/// A head that reads as a SHA, else `None`.
fn read_head(s: &str) -> Option<String> {
    is_sha(s).then(|| s.to_string())
}

struct Run {
    branch: String,
    source_sha: String,
    target_sha: String,
    commands: Vec<CommandResult>,
    not_run: Vec<String>,
    specs: Vec<String>,
}

struct Pr {
    seq: i64,
    number: u64,
    state: String,
    head: String,
    /// `(verdict, the head the verdict names)`; `None` when the row has no verdict.
    review: Option<(Verdict, Option<String>)>,
}

fn text(v: &Value, key: &str) -> Option<String> {
    v.get(key)?.as_str().map(str::to_string)
}

fn texts(v: &Value, key: &str) -> Option<Vec<String>> {
    v.get(key)?
        .as_array()?
        .iter()
        .map(|x| x.as_str().map(str::to_string))
        .collect()
}

fn parse_run(content: &str) -> Option<Run> {
    let v: Value = serde_json::from_str(content).ok()?;
    let source_sha = text(&v, "source_sha").filter(|s| is_sha(s))?;
    let target_sha = text(&v, "target_sha").filter(|s| is_sha(s))?;
    let commands = v
        .get("commands")?
        .as_array()?
        .iter()
        .map(|c| {
            Some(CommandResult {
                check: text(c, "check")?,
                exit: c.get("exit")?.as_i64()?,
            })
        })
        .collect::<Option<Vec<_>>>()?;
    Some(Run {
        branch: text(&v, "branch")?,
        source_sha,
        target_sha,
        commands,
        not_run: texts(&v, "not_run")?,
        specs: texts(&v, "specs")?,
    })
}

fn parse_pr(seq: i64, content: &str) -> Option<Pr> {
    let v: Value = serde_json::from_str(content).ok()?;
    let review = match v.get("review") {
        None | Some(Value::Null) => None,
        Some(r) => {
            let verdict = match r.get("verdict")?.as_str()? {
                "PASS" => Verdict::Pass,
                "FAIL" => Verdict::Fail,
                _ => return None,
            };
            let head = r.get("head").and_then(Value::as_str).and_then(read_head);
            Some((verdict, head))
        }
    };
    Some(Pr {
        seq,
        number: v.get("number")?.as_u64()?,
        state: text(&v, "state")?,
        head: text(&v, "head").filter(|s| is_sha(s))?,
        review,
    })
}

/// Adds `(verdict, head)` once: the same verdict seen again for the same head is one entry.
fn push_entry(
    entries: &mut Vec<AuditEntry>,
    verdict: Verdict,
    head: Option<String>,
    current: Option<&str>,
) {
    if entries
        .iter()
        .any(|e| e.verdict == verdict && e.head == head)
    {
        return;
    }
    let in_effect = head.is_some() && head.as_deref() == current;
    entries.push(AuditEntry {
        verdict,
        head,
        in_effect,
    });
}

/// Builds the view. `rows` is `None` when the observation store is missing or cannot be opened.
/// Rows of kinds other than `bundle-run@1` and `pr-state@1` are not this view's business.
pub fn link(rows: Option<&[StoredRow]>, lane: &LaneInput) -> LinkView {
    let Some(rows) = rows else {
        return LinkView {
            state: State::Unknown,
            reason: Some("the observation store is missing or cannot be opened".into()),
            unreadable: Vec::new(),
            bundles: Vec::new(),
        };
    };
    let mut ordered: Vec<&StoredRow> = rows.iter().collect();
    ordered.sort_by_key(|r| r.seq);
    let (mut runs, mut prs): (Vec<Run>, Vec<Pr>) = (Vec::new(), Vec::new());
    let mut unreadable = Vec::new();
    for r in ordered {
        let parsed = match r.kind.as_str() {
            "bundle-run@1" => parse_run(&r.content).map(|x| runs.push(x)),
            "pr-state@1" => parse_pr(r.seq, &r.content).map(|x| prs.push(x)),
            _ => continue,
        };
        if parsed.is_none() {
            unreadable.push(format!("row {} ({})", r.seq, r.kind));
        }
    }
    // One bundle per (branch, source SHA): a later run record of the same pair replaces the earlier.
    let mut latest: Vec<Run> = Vec::new();
    for run in runs {
        match latest
            .iter_mut()
            .find(|x| x.branch == run.branch && x.source_sha == run.source_sha)
        {
            Some(slot) => *slot = run,
            None => latest.push(run),
        }
    }
    let bundles = latest
        .into_iter()
        .map(|run| bundle(run, &prs, lane))
        .collect();
    let (state, reason) = if unreadable.is_empty() {
        (State::Known, None)
    } else {
        (
            State::Unknown,
            Some(format!("{} stored row(s) cannot be read", unreadable.len())),
        )
    };
    LinkView {
        state,
        reason,
        unreadable,
        bundles,
    }
}

fn bundle(run: Run, prs: &[Pr], lane: &LaneInput) -> Bundle {
    // The pull request of a bundle is one that showed the bundle's source SHA as its head. When
    // several did, the one that still shows it now comes first; none does, the latest to have
    // shown it (the run is then no longer in effect).
    let now = |n: u64| {
        prs.iter()
            .filter(|p| p.number == n)
            .max_by_key(|p| p.seq)
            .map(|p| p.head.as_str())
    };
    let shown: Vec<&Pr> = prs.iter().filter(|p| p.head == run.source_sha).collect();
    let number = shown
        .iter()
        .filter(|p| now(p.number) == Some(run.source_sha.as_str()))
        .max_by_key(|p| p.seq)
        .or_else(|| shown.iter().max_by_key(|p| p.seq))
        .map(|p| p.number);
    let pr = number.and_then(|n| {
        let of_n: Vec<&Pr> = prs.iter().filter(|p| p.number == n).collect();
        let current = of_n.iter().max_by_key(|p| p.seq)?;
        let mut audit = Vec::new();
        for p in &of_n {
            if let Some((verdict, head)) = &p.review {
                push_entry(&mut audit, *verdict, head.clone(), Some(&current.head));
            }
        }
        Some(PrView {
            number: n,
            state: current.state.clone(),
            head: current.head.clone(),
            audit,
        })
    });
    let run_effect = match &pr {
        None => Effect::NoPullRequest,
        Some(p) if p.head == run.source_sha => Effect::Standing,
        Some(_) => Effect::Lapsed,
    };
    let lane_layers = run
        .specs
        .iter()
        .map(|spec| {
            let head = lane.heads.get(spec).and_then(|h| read_head(h));
            let mut audit = Vec::new();
            for a in lane.audits.iter().filter(|a| &a.spec == spec) {
                push_entry(&mut audit, a.verdict, read_head(&a.head), head.as_deref());
            }
            LaneLayer {
                spec: spec.clone(),
                head,
                audit,
            }
        })
        .collect();
    Bundle {
        branch: run.branch,
        source_sha: run.source_sha,
        target_sha: run.target_sha,
        commands: run.commands,
        not_run: run.not_run,
        specs: run.specs,
        run_effect,
        pr,
        lane: lane_layers,
    }
}
