//! Independent AUDIT and completion admission (Issue #2004, 05-D).
//!
//! `audit` is the auditor's own run: it observes the candidate in PROJECT (commit, tree digest,
//! workspace id, toolchain), re-runs the approved verification through 05-C, and hands
//! everything to `reduce`, a pure function that decides admission. Receipts that arrive as JSON
//! are self-reports and never count; only the receipt the auditor's own run produced does
//! (#2004 leader decision D1). Admission is `admitted` only when every required check passed on
//! a fresh runner receipt of exactly this candidate, the reviewer was independent, and no
//! reviewer opinion contradicts the evidence.

use crate::digest;
use crate::plan::{build_plan, current_revision, Decision};
use crate::run::{self, ProducerKind, Receipt, RunOutcome, RunRequest, Summary};
use crate::verify::{self, CheckResult, Toolchain};
use serde::{Deserialize, Serialize};
use serde_json::Value;
use std::path::Path;

pub type Result<T> = std::result::Result<T, String>;

/// What the author presents for audit.
#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct Candidate {
    pub commit: String,
    pub tree_digest: String,
    /// 01-A `identity` workspace id of the worktree the work was done in.
    pub workspace_id: String,
}

#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct Finding {
    pub rule_id: String,
    pub severity: String,
    pub message: String,
}

/// The reviewer's free-text judgement: `pass` or `fail`. Never evidence by itself.
#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct Opinion {
    pub verdict: String,
    pub summary: String,
}

/// Self-declared by the caller; recorded and judged by the table, not verified.
#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct Reviewer {
    /// `claude`, `codex`, …
    pub host: String,
    /// Subagent or session identifier.
    pub identity: String,
    /// `fresh`, `fork`, `separate-session` or `same-session`.
    pub context_sharing: String,
    pub read_only: bool,
}

#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct Waiver {
    /// `check:<index>` of a required check.
    pub gate: String,
    /// `quality`, or one of the never-waivable `security`, `secret`, `permission`.
    pub gate_class: String,
    pub reason: String,
    pub approver: String,
    pub scope: String,
    /// `YYYY-MM-DDTHH:MM:SSZ`.
    pub expires_at: String,
}

#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct AuditInput {
    pub protocol_version: u32,
    pub request_id: String,
    pub candidate: Candidate,
    /// The 05-C request the auditor runs again.
    pub run: RunRequest,
    /// Indexes into the snapshot's verification commands that must pass.
    pub required_checks: Vec<usize>,
    /// Receipts the author produced; recorded as self-reports, never used for must-pass.
    #[serde(default)]
    pub self_reports: Vec<Value>,
    #[serde(default)]
    pub findings: Vec<Finding>,
    pub opinion: Option<Opinion>,
    pub reviewer: Reviewer,
    #[serde(default)]
    pub waivers_allowed: bool,
    #[serde(default)]
    pub waivers: Vec<Waiver>,
    /// `gh api repos/<o>/<r>/pulls/<n>` JSON when there is a PR; fetched by the caller.
    pub pr: Option<Value>,
    #[serde(default)]
    pub pr_lookup_failed: bool,
}

/// What the auditor saw in PROJECT itself.
#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct Observed {
    pub commit: String,
    pub tree_digest: String,
    pub workspace_id: Option<String>,
    pub toolchain: Toolchain,
}

/// The 00-D verdict vocabulary.
#[derive(Clone, Copy, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum Status {
    Passed,
    Failed,
    NotRun,
    Blocked,
    Stale,
    Unknown,
}

#[derive(Clone, Copy, Debug, PartialEq, Eq, Serialize)]
#[serde(rename_all = "snake_case")]
pub enum Admission {
    Admitted,
    Rejected,
    Held,
    Blocked,
}

#[derive(Clone, Copy, Debug, PartialEq, Eq, Serialize)]
#[serde(rename_all = "snake_case")]
pub enum PrBinding {
    NotApplicable,
    Ok,
    Stale,
    Unknown,
}

#[derive(Clone, Debug, PartialEq, Eq, Serialize)]
pub struct Independence {
    pub accepted: bool,
    pub limitation: Option<&'static str>,
}

#[derive(Clone, Debug, PartialEq, Eq, Serialize)]
pub struct MustPass {
    pub command_ref: usize,
    pub command: String,
    /// The original result; a waiver never rewrites it.
    pub status: Status,
    pub waived: bool,
}

#[derive(Clone, Debug, PartialEq, Eq, Serialize)]
pub struct StaleReceipt {
    pub attempt_id: String,
    pub reasons: Vec<&'static str>,
}

#[derive(Clone, Debug, PartialEq, Eq, Serialize)]
pub struct WaiverRecord {
    #[serde(flatten)]
    pub waiver: Waiver,
    pub original: Option<Status>,
    /// `waived` or `refused`.
    pub outcome: &'static str,
    pub refusal: Option<&'static str>,
}

#[derive(Clone, Debug, PartialEq, Eq, Serialize)]
pub struct Verdict {
    pub request_id: String,
    pub status: Status,
    pub admission: Admission,
    pub reasons: Vec<&'static str>,
    pub candidate: Candidate,
    pub observed: Observed,
    pub reviewer: Reviewer,
    pub independence: Independence,
    /// What `independence` was judged from: `self-declared` while the reviewer fields are the
    /// caller's own words (#2392 decision 6774899518 row 14, stage 1).
    pub independence_basis: &'static str,
    pub pr_binding: PrBinding,
    pub run_decision: Decision,
    pub run_reasons: Vec<String>,
    pub must_pass: Vec<MustPass>,
    pub stale_receipts: Vec<StaleReceipt>,
    pub excluded_self_reports: Vec<String>,
    pub findings: Vec<Finding>,
    pub waivers: Vec<WaiverRecord>,
    pub receipt: Option<Receipt>,
}

/// Independence rules (#2004 leader decision H1, conservative until 02-B/02-C measure more).
/// A Claude subagent with a fresh context or any separate session is independent; a fork, and a
/// Codex subagent whose context inheritance is undocumented, proceed with a recorded limitation;
/// the same session, an unknown host or an unknown sharing value is not independent.
const INDEPENDENCE: [(&str, &str, Option<&str>); 6] = [
    ("claude", "fresh", None),
    ("claude", "separate-session", None),
    (
        "claude",
        "fork",
        Some("fork shares the parent conversation and the files it read"),
    ),
    ("codex", "separate-session", None),
    (
        "codex",
        "fresh",
        Some("codex subagent context inheritance is undocumented"),
    ),
    (
        "codex",
        "fork",
        Some("fork shares the parent conversation and the files it read"),
    ),
];

/// The reviewer fields are taken as the caller states them; stage 2 of #2392 D2 derives them
/// from the candidate's claim trailer and the host session id instead.
const INDEPENDENCE_BASIS: &str = "self-declared";

pub fn independence(r: &Reviewer) -> Independence {
    match INDEPENDENCE
        .iter()
        .find(|(h, s, _)| *h == r.host && *s == r.context_sharing)
    {
        Some(&(_, _, limitation)) => Independence {
            accepted: true,
            limitation:
                limitation
                    .or((!r.read_only)
                        .then_some("reviewer could write; an edit makes a new candidate")),
        },
        None => Independence {
            accepted: false,
            limitation: None,
        },
    }
}

/// Severities that do not hold admission; any other value — `critical`, `high`, or a word this
/// table does not know — does (#2392 decision 6774899518 row 17 = A). Case-insensitive.
const OPEN_SEVERITIES: [&str; 3] = ["medium", "low", "info"];

/// Whether an open finding of this severity holds admission.
pub fn blocking_severity(severity: &str) -> bool {
    let s = severity.trim().to_ascii_lowercase();
    !OPEN_SEVERITIES.contains(&s.as_str())
}

/// The PR's `head.sha` against the audited commit; a failed lookup is `unknown`.
pub fn pr_binding(commit: &str, pr: Option<&Value>, lookup_failed: bool) -> PrBinding {
    if lookup_failed {
        return PrBinding::Unknown;
    }
    let Some(pr) = pr else {
        return PrBinding::NotApplicable;
    };
    match pr["head"]["sha"].as_str() {
        Some(sha) if sha == commit => PrBinding::Ok,
        Some(_) => PrBinding::Stale,
        None => PrBinding::Unknown,
    }
}

/// The auditor's own view of PROJECT.
pub fn observe(project: &Path) -> Result<Observed> {
    let identity = crate::identity::identity(project)?;
    Ok(Observed {
        commit: current_revision(project)?,
        tree_digest: crate::tree_digest(project, &run::TREE_EXCLUDES)?,
        workspace_id: identity["workspace_id"].as_str().map(String::from),
        toolchain: verify::toolchain(project),
    })
}

/// `YYYY-MM-DDTHH:MM:SSZ` now (civil-from-days, H. Hinnant), for waiver expiry.
pub fn utc_now() -> String {
    let secs = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map_or(0, |d| d.as_secs()) as i64;
    let (days, rem) = (secs.div_euclid(86_400), secs.rem_euclid(86_400));
    let z = days + 719_468;
    let era = z.div_euclid(146_097);
    let doe = z - era * 146_097;
    let yoe = (doe - doe / 1_460 + doe / 36_524 - doe / 146_096) / 365;
    let doy = doe - (365 * yoe + yoe / 4 - yoe / 100);
    let mp = (5 * doy + 2) / 153;
    let d = doy - (153 * mp + 2) / 5 + 1;
    let m = if mp < 10 { mp + 3 } else { mp - 9 };
    let y = yoe + era * 400 + i64::from(m <= 2);
    format!(
        "{y:04}-{m:02}-{d:02}T{:02}:{:02}:{:02}Z",
        rem / 3_600,
        rem % 3_600 / 60,
        rem % 60
    )
}

fn valid_timestamp(s: &str) -> bool {
    let b = s.as_bytes();
    b.len() == 20
        && b.iter().enumerate().all(|(i, c)| match i {
            4 | 7 => *c == b'-',
            10 => *c == b'T',
            13 | 16 => *c == b':',
            19 => *c == b'Z',
            _ => c.is_ascii_digit(),
        })
}

/// REQ-05D-01: identity, attempt and timestamps present.
fn receipt_well_formed(r: &Receipt) -> bool {
    !r.task_id.is_empty()
        && !r.attempt_id.is_empty()
        && r.checks.iter().all(|c| {
            c.result == CheckResult::NotRun
                || (c.started_unix_ms.is_some() && c.finished_unix_ms.is_some())
        })
}

/// REQ-05D-02: the receipt against the approved contracts and what the auditor observed.
fn receipt_stale_reasons(
    r: &Receipt,
    input: &AuditInput,
    observed: &Observed,
) -> Vec<&'static str> {
    let spec = &input.run.admission.spec;
    let spec_digest = spec
        .snapshot
        .as_ref()
        .map(|s| s.digest.clone())
        .unwrap_or_default();
    let plan_digest = build_plan(
        spec,
        &input.run.admission.tasks,
        &input.run.admission.expected_revision,
    )
    .plan_digest;
    let policy = spec.policy_revision.as_ref().map(|p| digest(p.as_bytes()));
    let mut out = Vec::new();
    if r.tree_digest != observed.tree_digest {
        out.push("tree-changed");
    }
    if r.spec_digest != spec_digest {
        out.push("spec-changed");
    }
    if r.plan_digest != plan_digest {
        out.push("plan-changed");
    }
    if r.policy_digest != policy {
        out.push("policy-changed");
    }
    if r.toolchain != observed.toolchain {
        out.push("toolchain-changed");
    }
    out
}

/// Decides admission. Pure: no files, no processes, no clock (`now` is passed in).
pub fn reduce(input: &AuditInput, observed: &Observed, outcome: &RunOutcome, now: &str) -> Verdict {
    let commands = &input.run.admission.spec.verification_commands;
    let independence = independence(&input.reviewer);
    let pr = pr_binding(
        &input.candidate.commit,
        input.pr.as_ref(),
        input.pr_lookup_failed,
    );

    let excluded_self_reports: Vec<String> = input
        .self_reports
        .iter()
        .map(|v| match run::ingest_self_report(v.clone()) {
            Ok(r) => r.attempt_id,
            Err(_) => "malformed".into(),
        })
        .collect();

    // Which receipt may count, and why another may not.
    let mut stale_receipts = Vec::new();
    let usable: Option<&Receipt> = match &outcome.receipt {
        Some(r) if r.producer.kind == ProducerKind::RunnerReceipt => {
            let reasons = receipt_stale_reasons(r, input, observed);
            if reasons.is_empty() {
                Some(r)
            } else {
                stale_receipts.push(StaleReceipt {
                    attempt_id: r.attempt_id.clone(),
                    reasons,
                });
                None
            }
        }
        _ => None,
    };
    let well_formed = usable.is_none_or(receipt_well_formed);
    let status_of = |i: usize| -> Status {
        match usable {
            Some(_) if !well_formed => Status::Unknown,
            Some(r) => match r
                .checks
                .iter()
                .find(|c| c.command_ref == i)
                .map(|c| c.result)
            {
                Some(CheckResult::Passed) => Status::Passed,
                Some(CheckResult::Failed) => Status::Failed,
                Some(CheckResult::NotRun) => Status::NotRun,
                None => Status::Unknown,
            },
            // A runner receipt was set aside as stale; anything else that is not one proves
            // nothing; no receipt at all means the run itself stopped.
            None => match &outcome.receipt {
                Some(r) if r.producer.kind == ProducerKind::RunnerReceipt => Status::Stale,
                Some(_) => Status::Unknown,
                None if outcome.decision == Decision::Stale => Status::Stale,
                None => Status::Blocked,
            },
        }
    };
    let mut must_pass: Vec<MustPass> = input
        .required_checks
        .iter()
        .map(|&i| MustPass {
            command_ref: i,
            command: commands.get(i).cloned().unwrap_or_default(),
            status: if i < commands.len() {
                status_of(i)
            } else {
                Status::Unknown
            },
            waived: false,
        })
        .collect();

    // Waivers: quality gates only, allowed by policy, not expired; the original stays.
    let waivers: Vec<WaiverRecord> = input
        .waivers
        .iter()
        .map(|w| {
            let target = w
                .gate
                .strip_prefix("check:")
                .and_then(|n| n.parse::<usize>().ok())
                .and_then(|n| must_pass.iter().position(|m| m.command_ref == n));
            let refusal = if !input.waivers_allowed {
                Some("waivers-not-allowed")
            } else if w.gate_class != "quality" {
                Some("gate-not-waivable")
            } else if !valid_timestamp(&w.expires_at) {
                Some("invalid-expiry")
            } else if w.expires_at.as_str() <= now {
                Some("expired")
            } else if target.is_none() {
                Some("unknown-gate")
            } else {
                None
            };
            let original = target.map(|t| must_pass[t].status);
            if let (None, Some(t)) = (refusal, target) {
                must_pass[t].waived = true;
            }
            WaiverRecord {
                waiver: w.clone(),
                original,
                outcome: if refusal.is_none() {
                    "waived"
                } else {
                    "refused"
                },
                refusal,
            }
        })
        .collect();

    let evidence_ok = !must_pass.is_empty()
        && must_pass
            .iter()
            .all(|m| m.status == Status::Passed || m.waived)
        && usable.is_some_and(|r| r.result != Summary::Unknown);
    let mut findings = input.findings.clone();
    let mut reasons: Vec<&'static str> = Vec::new();
    let worst = || {
        [
            Status::Failed,
            Status::Blocked,
            Status::Stale,
            Status::NotRun,
            Status::Unknown,
        ]
        .into_iter()
        .find(|s| must_pass.iter().any(|m| m.status == *s && !m.waived))
        .unwrap_or(Status::Unknown)
    };
    let (status, admission) = if !independence.accepted {
        reasons.push("independence-missing");
        (Status::Blocked, Admission::Blocked)
    } else if observed.workspace_id.is_none() {
        reasons.push("workspace-unknown");
        (Status::Unknown, Admission::Held)
    } else if observed.workspace_id.as_deref() != Some(input.candidate.workspace_id.as_str()) {
        reasons.push("foreign-workspace");
        (Status::Failed, Admission::Rejected)
    } else if observed.commit != input.candidate.commit
        || observed.tree_digest != input.candidate.tree_digest
    {
        reasons.push("candidate-changed");
        (Status::Stale, Admission::Rejected)
    } else if pr == PrBinding::Stale {
        reasons.push("pr-head-mismatch");
        (Status::Stale, Admission::Rejected)
    } else if pr == PrBinding::Unknown {
        reasons.push("pr-lookup-failed");
        (Status::Unknown, Admission::Held)
    } else if let Some(o) = input
        .opinion
        .as_ref()
        .filter(|o| (o.verdict == "pass") != evidence_ok)
    {
        reasons.push("opinion-conflicts-with-evidence");
        findings.push(Finding {
            rule_id: "opinion-conflicts-with-evidence".into(),
            severity: "high".into(),
            message: format!(
                "reviewer said {} while the evidence {}; the evidence decides and the verdict is held",
                o.verdict,
                if evidence_ok { "passes" } else { "does not pass" }
            ),
        });
        (Status::Unknown, Admission::Held)
    } else if usable.is_some_and(|r| r.result == Summary::Unknown) {
        reasons.push("oracle-changed");
        (Status::Unknown, Admission::Held)
    } else if !evidence_ok {
        reasons.push("must-pass-not-met");
        (worst(), Admission::Rejected)
    } else if input
        .findings
        .iter()
        .any(|f| blocking_severity(&f.severity))
    {
        reasons.push("blocking-finding-open");
        (Status::Unknown, Admission::Held)
    } else {
        (Status::Passed, Admission::Admitted)
    };

    Verdict {
        request_id: input.request_id.clone(),
        status,
        admission,
        reasons,
        candidate: input.candidate.clone(),
        observed: observed.clone(),
        reviewer: input.reviewer.clone(),
        independence,
        independence_basis: INDEPENDENCE_BASIS,
        pr_binding: pr,
        run_decision: outcome.decision,
        run_reasons: outcome.reasons.clone(),
        must_pass,
        stale_receipts,
        excluded_self_reports,
        findings,
        waivers,
        receipt: outcome.receipt.clone(),
    }
}

/// The auditor's run: observe PROJECT, re-run the approved verification, reduce.
pub fn audit(project: &Path, input: &AuditInput) -> Result<Verdict> {
    if input.required_checks.is_empty() {
        return Err("invalid audit request: required_checks is empty".into());
    }
    let observed = observe(project)?;
    let outcome = run::run(project, &input.run, &observed.commit)?;
    // #2392 F8: judge the project as it is after the checks ran. A check that changed the
    // tree (or the commit) while it ran then meets the same `candidate-changed` rejection as
    // a change made before the audit, instead of an admission of a tree nobody checked.
    let after = observe(project)?;
    Ok(reduce(input, &after, &outcome, &utc_now()))
}
