//! SPEC auto-archive eligibility planner (Issue #2187, D1; REQ-D1-01) — dry-run only.
//!
//! `archive-plan` scans `<project>/.moai-cli/specs/<SPEC-ID>/` directories and reports,
//! as JSON, which SPECs are archive-eligible and why every other scanned SPEC is
//! excluded. It writes nothing (observation-only by construction) and never touches the
//! network.
//!
//! Eligibility (REQ-D1-01; decisions 2026-11-09 00:5x — terminal + 90-day inactivity,
//! Go-inherited): a SPEC is eligible iff
//!  1. its `spec.md` frontmatter status is terminal — `completed`, `superseded`,
//!     `archived` or `rejected` (the inherited terminal set);
//!  2. its last activity is strictly before `now − grace_days` (default 90 days; the
//!     profile key `archive_grace_days` overrides, CLI `--grace-days` overrides the
//!     profile); and
//!  3. neither pending evidence (`evidence: pending`) nor a preservation hold
//!     (`hold: true`) is recorded — such candidates are excluded with the reason.
//!
//! Marker values are normalized before the comparison (quotes stripped, case
//! folded); a value outside the marker contract (`hold: yes`, `evidence:
//! complete`, …) excludes as `ambiguous_hold` / `ambiguous_evidence`, and an
//! indented key inside the block makes the whole block untrusted
//! (`ambiguous_frontmatter`) — a marker the parser cannot read is never treated
//! as "no marker" (design §12: 애매한 상태 제외; #2187 audit fix). Missing time
//! information is excluded as `missing_time_info`, never guessed.
//! Last activity is ONE git pass over the specs tree (newest commit time per SPEC
//! directory); outside git, or without history for a directory, the newest file mtime
//! of that directory is the activity source; with neither the SPEC is excluded as
//! `missing_time_info`.

use serde::Serialize;
use serde_json::json;
use std::collections::BTreeMap;
use std::path::Path;
use std::process::Command;
use std::time::{SystemTime, UNIX_EPOCH};

pub const SCHEMA_VERSION: u32 = 1;
/// The inherited inactivity window (design §12: 기본 90일 — Go `DefaultArchiveGraceDays`).
pub const DEFAULT_ARCHIVE_GRACE_DAYS: u64 = 90;
/// The statuses the archive policy treats as terminal (Go `archiveTerminalStatuses`,
/// inherited by decision 2026-11-09 00:5x). Deliberately narrower than any other
/// terminal reading elsewhere.
const TERMINAL_STATUSES: [&str; 4] = ["completed", "superseded", "archived", "rejected"];
const DAY_SECS: u64 = 86_400;

/// One archive-eligible SPEC directory.
#[derive(Clone, Debug, PartialEq, Eq, Serialize)]
pub struct EligibleSpec {
    pub spec_id: String,
    pub status: String,
    /// RFC3339 UTC of the last activity (commit time, else newest mtime).
    pub last_activity: String,
    /// Where the activity came from: `git` or `mtime`.
    pub activity_source: String,
}

/// One scanned SPEC directory that is excluded from eligibility, with the reason.
#[derive(Clone, Debug, PartialEq, Eq, Serialize)]
pub struct ExcludedSpec {
    pub spec_id: String,
    /// The parsed status when one could be read.
    pub status: Option<String>,
    /// `pending_evidence` | `hold` | `in_grace_window` | `ambiguous_status` |
    /// `ambiguous_evidence` | `ambiguous_hold` | `ambiguous_frontmatter` |
    /// `missing_time_info`
    pub reason: String,
}

/// The dry-run plan: everything the move step (a later issue) would need, without
/// moving anything.
#[derive(Clone, Debug, PartialEq, Eq, Serialize)]
pub struct ArchivePlan {
    pub protocol_version: u32,
    pub project: String,
    pub grace_days: u64,
    /// `request` | `profile` | `default`
    pub grace_source: String,
    /// RFC3339 UTC instant the evaluation ran at.
    pub now: String,
    /// `now − grace_days` — activity at or after this keeps a SPEC put.
    pub cutoff: String,
    pub scanned: usize,
    pub eligible: Vec<EligibleSpec>,
    pub excluded: Vec<ExcludedSpec>,
    pub network: String,
    pub writes: String,
}

type Result<T> = std::result::Result<T, String>;

fn rfc3339(secs: u64) -> String {
    let days = (secs / 86_400) as i64;
    let rem = (secs % 86_400) as u32;
    let z = days + 719_468;
    let era = z.div_euclid(146_097);
    let doe = z.rem_euclid(146_097);
    let yoe = (doe - doe / 1460 + doe / 36_524 - doe / 146_096) / 365;
    let y = yoe + era * 400;
    let doy = doe - (365 * yoe + yoe / 4 - yoe / 146_096);
    let mp = (5 * doy + 2) / 153;
    let d = doy - (153 * mp + 2) / 5 + 1;
    let m = if mp < 10 { mp + 3 } else { mp - 9 };
    let day = if m <= 2 { y + 1 } else { y };
    format!(
        "{:04}-{:02}-{:02}T{:02}:{:02}:{:02}Z",
        day,
        m,
        d,
        rem / 3600,
        (rem % 3600) / 60,
        rem % 60
    )
}

/// The grace window and where it came from: the CLI request, then the profile key
/// `archive_grace_days`, then [`DEFAULT_ARCHIVE_GRACE_DAYS`] (REQ-D1-01 — the value is
/// profile-overridable). A missing profile means the default; an unreadable profile is
/// an error, never a guess.
pub fn resolve_grace_days(project: &Path, request: Option<u64>) -> Result<(u64, &'static str)> {
    if let Some(d) = request {
        return Ok((d, "request"));
    }
    let profile_path = project.join(crate::PROJECT_STATE_DIR).join("project.yaml");
    if crate::apply::read_opt(&profile_path)?.is_none() {
        return Ok((DEFAULT_ARCHIVE_GRACE_DAYS, "default"));
    }
    let (profile, _) = crate::profile::read_profile(project)?;
    Ok(match profile.archive_grace_days {
        Some(d) if d > 0 => (d, "profile"),
        _ => (DEFAULT_ARCHIVE_GRACE_DAYS, "default"),
    })
}

/// Minimal frontmatter scan of `spec.md`: the raw `status`, `evidence` and `hold`
/// values from the leading `---` block. Missing file or missing block reads as all
/// `None` — the caller decides what absence means.
fn frontmatter(dir: &Path) -> Result<Frontmatter> {
    let bytes = match crate::apply::read_opt(&dir.join("spec.md"))? {
        Some(b) => b,
        None => return Ok(Frontmatter::absent()),
    };
    let text = String::from_utf8(bytes).map_err(|_| "spec.md is not UTF-8".to_string())?;
    let mut lines = text.lines();
    if lines.next().ok_or("empty spec.md")?.trim() != "---" {
        return Ok(Frontmatter::absent());
    }
    let mut fm = Frontmatter::absent();
    for line in lines {
        if line.trim() == "---" {
            break;
        }
        // Only column-0 keys are this contract's vocabulary. An indented key means
        // nested YAML the parser does not read, so the block is recorded untrusted
        // and no marker from it may vouch for eligibility (#2187 audit fix).
        if !line.trim().is_empty() && line.starts_with([' ', '\t']) {
            fm.indented = true;
            continue;
        }
        if let Some(v) = line.strip_prefix("status:") {
            fm.status = Some(unquote(v));
        } else if let Some(v) = line.strip_prefix("evidence:") {
            fm.evidence = Some(unquote(v));
        } else if let Some(v) = line.strip_prefix("hold:") {
            fm.hold = Some(unquote(v));
        }
    }
    Ok(fm)
}

/// The frontmatter markers a `spec.md` carries, as raw value strings.
struct Frontmatter {
    status: Option<String>,
    evidence: Option<String>,
    hold: Option<String>,
    /// The block held an indented key: nested structure this flat contract cannot
    /// vouch for, so every marker read from it is untrusted.
    indented: bool,
}

impl Frontmatter {
    fn absent() -> Self {
        Self {
            status: None,
            evidence: None,
            hold: None,
            indented: false,
        }
    }
}

/// Strips the surrounding whitespace and quotes of a frontmatter value, case
/// preserved (the status echo stays verbatim).
fn unquote(v: &str) -> String {
    v.trim().trim_matches('"').to_string()
}

/// The three readings of a hold·evidence marker: the signal, an explicit
/// no-signal, or a value outside the contract — which excludes, never eligible
/// (design §12: 애매한 상태 제외; #2187 audit fix).
enum Marker {
    Signal,
    Clear,
    Ambiguous,
}

/// Reads the `hold` value: `true` (any case, quoting already stripped) signals a
/// hold, `false` an explicit no-hold, anything else is ambiguous.
fn hold_marker(value: Option<&str>) -> Marker {
    match value {
        None | Some("") => Marker::Clear,
        Some(v) => match v.to_ascii_lowercase().as_str() {
            "true" => Marker::Signal,
            "false" => Marker::Clear,
            _ => Marker::Ambiguous,
        },
    }
}

/// Reads the `evidence` value: `pending` (any case, quoting already stripped)
/// signals pending evidence; this contract knows no other value, so anything
/// else is ambiguous.
fn evidence_marker(value: Option<&str>) -> Marker {
    match value {
        None | Some("") => Marker::Clear,
        Some(v) => match v.to_ascii_lowercase().as_str() {
            "pending" => Marker::Signal,
            _ => Marker::Ambiguous,
        },
    }
}

/// One git pass over the specs tree: per-SPEC-directory newest commit time (Unix
/// seconds). Outside git, or with no history, the map is empty and callers fall back
/// to mtimes.
fn git_activity(project: &Path, specs: &Path) -> BTreeMap<String, u64> {
    let mut map = BTreeMap::new();
    let out = Command::new("git")
        .arg("-C")
        .arg(project)
        .args([
            "log",
            "--no-renames",
            "--name-only",
            "--pretty=format:@%ct",
            "--",
        ])
        .arg(specs)
        .output();
    let Ok(out) = out else {
        return map;
    };
    if !out.status.success() {
        return map;
    }
    let specs_prefix = specs.strip_prefix(project).unwrap_or(specs).to_path_buf();
    let mut current_time: Option<u64> = None;
    for line in String::from_utf8_lossy(&out.stdout).lines() {
        if let Some(t) = line.strip_prefix('@') {
            current_time = t.trim().parse::<u64>().ok();
            continue;
        }
        let (Some(secs), path) = (current_time, line.trim()) else {
            continue;
        };
        if path.is_empty() {
            continue;
        }
        let rel = Path::new(path);
        let Ok(stripped) = rel.strip_prefix(&specs_prefix) else {
            continue;
        };
        if let Some(id) = stripped.components().next() {
            let id = id.as_os_str().to_string_lossy().into_owned();
            map.entry(id)
                .and_modify(|e| *e = (*e).max(secs))
                .or_insert(secs);
        }
    }
    map
}

/// Newest file mtime in `dir` (shallow), as Unix seconds. `None` when nothing is
/// readable.
fn newest_mtime(dir: &Path) -> Option<u64> {
    let mut newest = None;
    for entry in std::fs::read_dir(dir).ok()? {
        let entry = entry.ok()?;
        let meta = entry.metadata().ok()?;
        if meta.is_file() {
            if let Ok(m) = meta.modified() {
                if let Ok(secs) = m.duration_since(UNIX_EPOCH) {
                    newest = Some(newest.unwrap_or(0).max(secs.as_secs()));
                }
            }
        }
    }
    newest
}

/// Computes the dry-run archive plan for `project` (REQ-D1-01). Writes nothing.
pub fn plan(project: &Path, now: SystemTime, grace_request: Option<u64>) -> Result<ArchivePlan> {
    let now_secs = now
        .duration_since(UNIX_EPOCH)
        .map_err(|_| "system clock is before the Unix epoch".to_string())?
        .as_secs();
    let (grace_days, grace_source) = resolve_grace_days(project, grace_request)?;
    let cutoff_secs = now_secs.saturating_sub(grace_days.saturating_mul(DAY_SECS));
    let specs = project.join(crate::PROJECT_STATE_DIR).join("specs");

    let mut plan = ArchivePlan {
        protocol_version: SCHEMA_VERSION,
        project: project.to_string_lossy().into_owned(),
        grace_days,
        grace_source: grace_source.to_string(),
        now: rfc3339(now_secs),
        cutoff: rfc3339(cutoff_secs),
        scanned: 0,
        eligible: Vec::new(),
        excluded: Vec::new(),
        network: "disabled".into(),
        writes: "none".into(),
    };
    if !specs.is_dir() {
        // No SPECs to archive is a clean empty plan, not an error.
        return Ok(plan);
    }

    let activity = git_activity(project, &specs);
    let mut ids: Vec<String> = std::fs::read_dir(&specs)
        .map_err(|e| format!("specs directory unreadable: {e}"))?
        .filter_map(|e| e.ok())
        .filter(|e| e.file_type().map(|t| t.is_dir()).unwrap_or(false))
        .map(|e| e.file_name().to_string_lossy().into_owned())
        .collect();
    ids.sort();

    for spec_id in ids {
        plan.scanned += 1;
        let dir = specs.join(&spec_id);
        let Ok(fm) = frontmatter(&dir) else {
            plan.excluded.push(ExcludedSpec {
                spec_id: spec_id.clone(),
                status: None,
                reason: "ambiguous_status".into(),
            });
            continue;
        };
        if fm.indented {
            // The block carried nested structure; no marker from it may vouch for
            // eligibility, even when the status line itself parsed.
            plan.excluded.push(ExcludedSpec {
                spec_id: spec_id.clone(),
                status: fm.status.clone(),
                reason: "ambiguous_frontmatter".into(),
            });
            continue;
        }
        let Some(status) = fm.status else {
            plan.excluded.push(ExcludedSpec {
                spec_id: spec_id.clone(),
                status: None,
                reason: "ambiguous_status".into(),
            });
            continue;
        };
        if !TERMINAL_STATUSES.contains(&status.as_str()) {
            // Non-terminal SPECs are simply not candidates; the scanned count
            // records that they were looked at.
            continue;
        }
        match evidence_marker(fm.evidence.as_deref()) {
            Marker::Signal => {
                plan.excluded.push(ExcludedSpec {
                    spec_id: spec_id.clone(),
                    status: Some(status.clone()),
                    reason: "pending_evidence".into(),
                });
                continue;
            }
            Marker::Ambiguous => {
                plan.excluded.push(ExcludedSpec {
                    spec_id: spec_id.clone(),
                    status: Some(status.clone()),
                    reason: "ambiguous_evidence".into(),
                });
                continue;
            }
            Marker::Clear => {}
        }
        match hold_marker(fm.hold.as_deref()) {
            Marker::Signal => {
                plan.excluded.push(ExcludedSpec {
                    spec_id: spec_id.clone(),
                    status: Some(status.clone()),
                    reason: "hold".into(),
                });
                continue;
            }
            Marker::Ambiguous => {
                plan.excluded.push(ExcludedSpec {
                    spec_id: spec_id.clone(),
                    status: Some(status.clone()),
                    reason: "ambiguous_hold".into(),
                });
                continue;
            }
            Marker::Clear => {}
        }
        let (last_activity, source) = match activity.get(&spec_id) {
            Some(secs) => (*secs, "git"),
            None => match newest_mtime(&dir) {
                Some(secs) => (secs, "mtime"),
                None => {
                    plan.excluded.push(ExcludedSpec {
                        spec_id: spec_id.clone(),
                        status: Some(status.clone()),
                        reason: "missing_time_info".into(),
                    });
                    continue;
                }
            },
        };
        if last_activity >= cutoff_secs {
            plan.excluded.push(ExcludedSpec {
                spec_id: spec_id.clone(),
                status: Some(status.clone()),
                reason: "in_grace_window".into(),
            });
            continue;
        }
        plan.eligible.push(EligibleSpec {
            spec_id: spec_id.clone(),
            status: status.clone(),
            last_activity: rfc3339(last_activity),
            activity_source: source.into(),
        });
    }
    Ok(plan)
}

/// JSON value used by the CLI envelope, so callers see the same field set the docs
/// record.
pub fn to_value(plan: &ArchivePlan) -> serde_json::Value {
    json!({
        "protocol_version": plan.protocol_version,
        "project": plan.project,
        "grace_days": plan.grace_days,
        "grace_source": plan.grace_source,
        "now": plan.now,
        "cutoff": plan.cutoff,
        "scanned": plan.scanned,
        "eligible": plan.eligible,
        "excluded": plan.excluded,
        "network": plan.network,
        "writes": plan.writes,
    })
}

// Unused-today placeholder for the move step (a later issue): the plan's eligible
// list is the input contract, and the move must copy → hash → read back → pointer
// before it may remove anything (decisions 2026-11-09 00:5x).
// @MX:DEBT [P09 archive move] only the dry-run planner exists; the automatic move is
// decided (allowed with hash/readback/restore) but not implemented.
// @MX:CEILING RC1 archive eligibility is terminal + 90-day inactivity with the two
// exclusions; a SPEC dir missing both time sources is excluded, never guessed.
// @MX:UPGRADE implement ExecuteArchive against ArchivePlan.eligible when the storage
// EP returns to this area.
