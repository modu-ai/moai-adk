//! The two commands that bring a host's open issues into the local ledger (#3071 H3b):
//! `card-import-plan` (a preview that writes nothing) and `card-import-apply` (writes only the
//! preview the user was shown). The `src/ops/card_import_*.rs` files only parse arguments, build
//! the GitHub connector and call these; the functions take the connector as a trait object, so
//! the tests of this slice use [`super::connector::FakeConnector`] or a stub `gh`.
//!
//! The ledger is never created here: `ledger-register-apply` makes it. A preview says whether the
//! ledger is ready; an apply without a ledger stops before it reads the host. The text of an
//! issue is data: the preview lists what would be kept and what was dropped, and acts on none of
//! it.
use super::connector::{HostConnector, HostKind, ProjectRef};
use super::import_apply::{apply_import, ImportReport};
use super::import_plan::{plan_import, AcReading, ImportPlan, PlannedCard};
use crate::ledger::summary::{sources_in_ledger, SourceCount};
use crate::ledger::{self, import::SqliteImport, LedgerError};
use serde_json::{json, Value};
use std::path::Path;

type Result<T> = std::result::Result<T, String>;

/// The host project named by `--from` and `--repo`. A host whose transport is not written yet
/// (GitLab) is refused here, before any connector exists, so nothing is sent.
pub fn project_ref(from: &str, repo: &str) -> Result<ProjectRef> {
    let host = HostKind::parse(from)
        .and_then(HostKind::require_implemented)
        .map_err(|e| e.to_string())?;
    ProjectRef::parse(host, repo).map_err(|e| e.to_string())
}

fn card_json(card: &PlannedCard) -> Value {
    let (criteria, criteria_problem) = match &card.ac {
        AcReading::Read { rows, skipped } => (
            json!({"rows": rows.len(), "skipped_rows": skipped}),
            Value::Null,
        ),
        AcReading::Unreadable(problem) => (Value::Null, json!(format!("{problem:?}"))),
    };
    let mut notes = serde_json::Map::new();
    for note in &card.notes {
        let count = notes.entry(note.kind.as_str()).or_insert(json!(0));
        *count = json!(count.as_u64().unwrap_or(0) + 1);
    }
    json!({
        "number": card.source_number,
        "source": format!(
            "{} {}#{}",
            card.source_host, card.source_project, card.source_number
        ),
        "title": card.title,
        "kind": card.kind,
        "kind_source": format!("{:?}", card.kind_source).to_lowercase(),
        "criteria": criteria,
        "criteria_unreadable": criteria_problem,
        "notes": notes,
        "member_control_lines": card.member_control_lines,
        "dropped": {
            "control_lines": card.dropped.lines.len(),
            "other": card.dropped.other,
            "unread": card.dropped.unread,
        },
    })
}

fn plan_json(plan: &ImportPlan) -> Value {
    json!({
        "source": {"host": plan.source_host, "project": plan.source_project},
        "plan_digest": plan.digest,
        "cards": plan.cards.iter().map(card_json).collect::<Vec<_>>(),
        "card_count": plan.cards.len(),
        "criteria_unreadable": plan.unreadable_ac(),
        "duplicate_numbers": plan.duplicate_numbers,
        "pull_requests_skipped": plan.pull_requests_skipped,
    })
}

/// The host projects a ledger already holds cards from, as plan JSON, and a one-line notice when
/// that is more than the one project `host`/`project` (leader decision on #3119: a ledger may
/// hold several repositories, and the plan says so instead of refusing).
pub(super) fn sources_json(sources: &[SourceCount], host: &str, project: &str) -> (Value, Value) {
    let list: Vec<Value> = sources
        .iter()
        .map(|s| json!({"host": s.host, "project": s.project, "cards": s.cards}))
        .collect();
    let others: Vec<String> = sources
        .iter()
        .filter(|s| !(s.host == host && s.project == project))
        .map(|s| format!("{} {} ({} cards)", s.host, s.project, s.cards))
        .collect();
    let notice = if others.is_empty() {
        Value::Null
    } else {
        json!(format!(
            "this ledger already holds cards from other repositories: {}; {host} {project} is a different source",
            others.join(", ")
        ))
    };
    (json!(list), notice)
}

/// What the ledger of the project at `home` can take. Opens it read-only: nothing is created.
fn ledger_json(home: &Path, project_id: &str, host: &str, project: &str) -> Value {
    match ledger::open_read_only(home, project_id) {
        Ok(read) => match sources_in_ledger(&read) {
            Ok(sources) => {
                let (list, notice) = sources_json(&sources, host, project);
                json!({"state": "ready", "problem": null, "sources": list, "notice": notice})
            }
            Err(error) => json!({"state": "unavailable", "problem": error.code()}),
        },
        Err(LedgerError::Absent) => json!({"state": "absent", "problem": null}),
        Err(error) => json!({"state": "unavailable", "problem": error.code()}),
    }
}

/// `card-import-plan`: reads every page of the open issues and says what cards they would make.
/// Writes nothing, creates no folder and no file.
pub fn plan(
    project: &Path,
    home: &Path,
    connector: &dyn HostConnector,
    host_project: &ProjectRef,
) -> Result<Value> {
    let project_id = ledger::command::project_id_of(project)?;
    if connector.host() != host_project.host() {
        return Err("the connector is for another host than the project".into());
    }
    let issues = connector
        .open_issues(host_project)
        .map_err(|e| e.to_string())?;
    let plan = plan_import(&issues);
    let mut out = plan_json(&plan);
    out["project_id"] = json!(project_id);
    out["ledger"] = ledger_json(
        home,
        &project_id,
        host_project.host().as_str(),
        &host_project.source_project(),
    );
    out["writes"] = json!(false);
    out["network"] = json!("read-only (GET)");
    out["apply"] = json!(
        "card-import-apply with the same project, --from, --repo and home, this plan_digest and a permission reference, only after the user approves; the ledger must exist (ledger-register-apply)"
    );
    Ok(out)
}

/// `card-import-apply`: writes the preview the caller holds, only when the host still gives the
/// same issues (the digest is recomputed from a fresh read). The ledger must already exist.
pub fn apply(
    project: &Path,
    home: &Path,
    connector: &dyn HostConnector,
    host_project: &ProjectRef,
    plan_digest: &str,
    permission_reference: &str,
) -> Result<Value> {
    if !crate::contract::permission_reference_ok(permission_reference) {
        return Err("permission reference required".into());
    }
    let project_id = ledger::command::project_id_of(project)?;
    // The ledger first: without one nothing is read from the host.
    let mut write = ledger::open_write(home, &project_id).map_err(|error| match error {
        LedgerError::Absent => "ledger not found: register it first (ledger-register-plan, then ledger-register-apply); nothing was read or written".to_string(),
        other => other.to_string(),
    })?;
    let report = apply_import(
        connector,
        &mut SqliteImport::new(&mut write),
        host_project,
        plan_digest,
    )
    .map_err(|e| e.to_string())?;
    Ok(report_json(&report, &project_id))
}

fn report_json(report: &ImportReport, project_id: &str) -> Value {
    json!({
        "project_id": project_id,
        "state": if report.created.is_empty() { "unchanged" } else { "applied" },
        "plan_digest": report.digest,
        "created": report.created,
        "already_there": report.already_there,
        "duplicate_numbers": report.duplicate_numbers,
        "created_without_criteria": report.created_without_criteria,
        "notes_written": report.notes_written,
        "pull_requests_skipped": report.pull_requests_skipped,
        "network": "read-only (GET)",
    })
}
