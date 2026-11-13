//! SPEC auto-archive: the eligibility planner (Issue #2187, D1; REQ-D1-01) and the
//! verified move, restore and search (Issue #2188, D2; REQ-D2-01 — see the D2 block
//! further down).
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
//! complete`, …) excludes as `ambiguous_hold` / `ambiguous_evidence`, and any form
//! the flat parser cannot read makes the whole block untrusted
//! (`ambiguous_frontmatter`): an indented key, a line without `key:`, a marker key
//! not spelled exactly `key: value` (`hold : true`, `"hold": true`, `hold:true`), the
//! same marker key twice, or a flow mapping naming a marker key — a marker the parser
//! cannot read is never treated as "no marker" (design §12: ambiguous states are excluded; #2187
//! audit fixes). A readable non-terminal status is recorded as `not_terminal`.
//! Missing time information is excluded as `missing_time_info`, never guessed.
//! Last activity is ONE git pass over the specs tree (newest commit time per SPEC
//! directory); outside git, or without history for a directory, the newest file mtime
//! of that directory is the activity source; with neither the SPEC is excluded as
//! `missing_time_info`.

use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::collections::BTreeMap;
use std::fs;
use std::path::{Path, PathBuf};
use std::process::Command;
use std::time::{SystemTime, UNIX_EPOCH};

pub const SCHEMA_VERSION: u32 = 1;
/// The inherited inactivity window (design §12: 90 days by default — Go `DefaultArchiveGraceDays`).
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
    /// `missing_time_info` | `not_terminal`
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
    /// Digest of `eligible` only (D2 #2188): what `archive-apply --plan-digest`
    /// binds to. `now` and `cutoff` are left out because they change every run.
    pub plan_digest: String,
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
        // Blank lines and whole-line comments carry no marker.
        if line.trim().is_empty() || line.trim_start().starts_with('#') {
            continue;
        }
        // Only column-0 keys are this contract's vocabulary. An indented key means
        // nested YAML the parser does not read, so the block is recorded untrusted
        // and no marker from it may vouch for eligibility (#2187 audit fix).
        if line.starts_with([' ', '\t']) {
            fm.untrusted = true;
            continue;
        }
        // A line that is not `key: value` is a form this parser cannot read.
        let Some((raw_key, value)) = line.split_once(':') else {
            fm.untrusted = true;
            continue;
        };
        let key = key_name(raw_key);
        if !MARKER_KEYS.contains(&key.as_str()) {
            // A flow mapping naming a marker key (`meta: {hold: true}`) is nested
            // structure this flat contract cannot vouch for.
            let v = value.trim_start();
            if (v.starts_with('{') || v.starts_with('['))
                && MARKER_KEYS.iter().any(|k| v.contains(&format!("{k}:")))
            {
                fm.untrusted = true;
            }
            continue;
        }
        // A marker key in any form other than exactly `key: value` — `hold : true`,
        // a quoted key, `hold:true` — is a form this parser does not trust (#2187
        // epoch-3 audit fix F1); YAML itself may read such lines as the marker.
        if raw_key != key || !(value.is_empty() || value.starts_with([' ', '\t'])) {
            fm.untrusted = true;
            continue;
        }
        let slot = match key.as_str() {
            "status" => &mut fm.status,
            "evidence" => &mut fm.evidence,
            _ => &mut fm.hold,
        };
        // The same marker twice is ambiguous whatever the values: YAML keeps the
        // last one, which would let a later line silently clear a hold (F2).
        if slot.is_some() {
            fm.untrusted = true;
            continue;
        }
        *slot = Some(unquote(value));
    }
    Ok(fm)
}

/// The frontmatter keys this contract reads; any other column-0 key is ignored.
const MARKER_KEYS: [&str; 3] = ["status", "evidence", "hold"];

/// A key's name for the vocabulary comparison: surrounding whitespace and one layer
/// of double or single quotes removed. Comparing the result with the raw key tells
/// an exact `key:` apart from every other spelling of the same name.
fn key_name(raw: &str) -> String {
    let t = raw.trim();
    let unq = t
        .strip_prefix('"')
        .and_then(|s| s.strip_suffix('"'))
        .or_else(|| t.strip_prefix('\'').and_then(|s| s.strip_suffix('\'')))
        .unwrap_or(t);
    unq.trim().to_string()
}

/// The frontmatter markers a `spec.md` carries, as raw value strings.
struct Frontmatter {
    status: Option<String>,
    evidence: Option<String>,
    hold: Option<String>,
    /// The block held a form this flat contract cannot vouch for — an indented key,
    /// a line without `key:`, a marker key not spelled exactly `key: value`, a
    /// repeated marker key or a flow mapping naming one — so every marker read from
    /// it is untrusted.
    untrusted: bool,
}

impl Frontmatter {
    fn absent() -> Self {
        Self {
            status: None,
            evidence: None,
            hold: None,
            untrusted: false,
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
/// (design §12: ambiguous states are excluded; #2187 audit fix).
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
    let mut cmd = Command::new("git");
    crate::identity::scrub_git_env(&mut cmd);
    let out = cmd
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
        plan_digest: eligible_digest(&[]),
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
        // Every moved id must also be accepted by restore; staging names stay put.
        if !spec_id_ok(&spec_id) {
            plan.excluded.push(ExcludedSpec {
                spec_id,
                status: None,
                reason: "invalid_spec_id".into(),
            });
            continue;
        }
        let dir = specs.join(&spec_id);
        let Ok(fm) = frontmatter(&dir) else {
            plan.excluded.push(ExcludedSpec {
                spec_id: spec_id.clone(),
                status: None,
                reason: "ambiguous_status".into(),
            });
            continue;
        };
        if fm.untrusted {
            // The block carried a form the parser cannot read; no marker from it may
            // vouch for eligibility, even when the status line itself parsed.
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
            // Not a candidate — recorded with its status, so "looked at and not
            // terminal" is distinguishable from "not looked at" (audit note 3).
            plan.excluded.push(ExcludedSpec {
                spec_id: spec_id.clone(),
                status: Some(status.clone()),
                reason: "not_terminal".into(),
            });
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
    plan.plan_digest = eligible_digest(&plan.eligible);
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
        "plan_digest": plan.plan_digest,
    })
}

// ---------------------------------------------------------------------------
// D2 #2188 (REQ-D2-01): automatic archive move, restore and search.
//
// Move order (decisions 2026-11-09 00:5x — automatic move allowed, hash/readback and
// restore required):
//  1. hash before the move: the source directory's manifest (sha256 and size per
//     relative path, plus subdirectories)
//  2. copy into a unique staging directory under `.moai-cli/archive/specs/`
//  3. readback: re-read the staging copy and compare its manifest with step 1
//  4. source re-check: the source must not have changed during the copy
//  5. rename staging to the final place, write the manifest file, update the index
//     (the pointer)
//  6. only then remove the source
// If step 3 or 4 disagrees, the staging copy is removed and the original is kept.

/// Archive root, project-relative.
pub const ARCHIVE_DIR: &str = crate::project_state!("archive");
const INDEX_SCHEMA_VERSION: u32 = 1;
/// Maximum search query length, in characters.
pub const SEARCH_QUERY_CHARS_MAX: usize = 256;
/// Maximum number of search hits; more sets `truncated: true`.
pub const SEARCH_RESULTS_MAX: usize = 200;

/// One manifest file: path relative to the directory (`/`-separated), sha256, size.
#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct ManifestFile {
    pub path: String,
    pub sha256: String,
    pub bytes: u64,
}

/// Content fingerprint of one directory: files and subdirectories in path order.
#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct Manifest {
    pub files: Vec<ManifestFile>,
    pub dirs: Vec<String>,
}

impl Manifest {
    /// Digest of the whole manifest — recorded in the index, checked on restore.
    pub fn digest(&self) -> String {
        crate::digest(&serde_json::to_vec(self).unwrap_or_default())
    }
}

/// Builds the manifest of `dir` recursively. Symlinks and anything that is not a
/// regular file or directory are refused, so nothing behind a link is copied or removed.
pub fn manifest(dir: &Path) -> Result<Manifest> {
    let meta = fs::symlink_metadata(dir).map_err(|_| "SPEC directory unavailable".to_string())?;
    if meta.file_type().is_symlink() {
        return Err("symlink in SPEC directory refused".into());
    }
    if !meta.is_dir() {
        return Err("SPEC path is not a directory".into());
    }
    let mut m = Manifest {
        files: Vec::new(),
        dirs: Vec::new(),
    };
    walk(dir, "", &mut m)?;
    m.files.sort_by(|a, b| a.path.cmp(&b.path));
    m.dirs.sort();
    Ok(m)
}

fn walk(dir: &Path, prefix: &str, m: &mut Manifest) -> Result<()> {
    let entries = fs::read_dir(dir).map_err(|_| "SPEC directory unreadable".to_string())?;
    for entry in entries {
        let entry = entry.map_err(|_| "SPEC directory unreadable".to_string())?;
        let name = entry
            .file_name()
            .into_string()
            .map_err(|_| "SPEC file name is not UTF-8".to_string())?;
        let rel = if prefix.is_empty() {
            name.clone()
        } else {
            format!("{prefix}/{name}")
        };
        let ft = entry
            .file_type()
            .map_err(|_| "SPEC directory unreadable".to_string())?;
        let path = entry.path();
        if ft.is_symlink() {
            return Err(format!("symlink in SPEC directory refused: {rel}"));
        } else if ft.is_dir() {
            m.dirs.push(rel.clone());
            walk(&path, &rel, m)?;
        } else if ft.is_file() {
            let bytes = fs::read(&path).map_err(|_| format!("SPEC file read failed: {rel}"))?;
            m.files.push(ManifestFile {
                path: rel,
                sha256: crate::digest(&bytes),
                bytes: bytes.len() as u64,
            });
        } else {
            return Err(format!("special file in SPEC directory refused: {rel}"));
        }
    }
    Ok(())
}

/// Copies `src` to `dst` and publishes `dst` only after the readback matches. The
/// source is never removed here (the caller removes it after the index write). An
/// existing `dst` is refused.
pub fn copy_verified(src: &Path, dst: &Path, expected: &Manifest) -> Result<()> {
    copy_verified_with(src, dst, expected, |_| {})
}

/// As [`copy_verified`], calling `after_copy(staging)` right after the copy and before
/// the readback — the seam a test uses to corrupt the staged copy and exercise the
/// readback check directly.
#[doc(hidden)]
pub fn copy_verified_with(
    src: &Path,
    dst: &Path,
    expected: &Manifest,
    after_copy: impl FnOnce(&Path),
) -> Result<()> {
    if fs::symlink_metadata(dst).is_ok() {
        return Err("archive destination already exists; exclusive create refused".into());
    }
    let parent = dst
        .parent()
        .ok_or("archive destination parent unavailable")?;
    let name = dst
        .file_name()
        .and_then(|n| n.to_str())
        .ok_or("archive destination name invalid")?;
    fs::create_dir_all(parent).map_err(|_| "archive directory unavailable".to_string())?;
    let nanos = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .map(|d| d.as_nanos())
        .unwrap_or(0);
    let staging = parent.join(format!(".{name}.staging-{}-{nanos}", std::process::id()));
    fs::create_dir(&staging).map_err(|_| "archive staging unavailable".to_string())?;
    let outcome = (|| {
        for d in &expected.dirs {
            fs::create_dir_all(staging.join(d))
                .map_err(|_| "archive copy write failed".to_string())?;
        }
        for f in &expected.files {
            fs::copy(src.join(&f.path), staging.join(&f.path))
                .map_err(|_| format!("archive copy write failed: {}", f.path))?;
        }
        after_copy(&staging);
        // Readback: re-read the written copy and compare it with the pre-move hash (REQ-D2-01).
        if manifest(&staging)? != *expected {
            return Err("readback digest mismatch; move cancelled, original kept".to_string());
        }
        // A source changed mid-copy would lose that change once the source is removed.
        if manifest(src)? != *expected {
            return Err("source changed during apply; move cancelled, original kept".to_string());
        }
        fs::rename(&staging, dst).map_err(|_| "archive publish write failed".to_string())
    })();
    if outcome.is_err() {
        let _ = fs::remove_dir_all(&staging);
    }
    outcome
}

/// One archive index entry — the pointer to an archived SPEC.
#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct IndexEntry {
    pub spec_id: String,
    pub status: String,
    pub archived_at: String,
    /// Project-relative path.
    pub path: String,
    pub manifest_digest: String,
    pub file_count: usize,
}

#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
struct Index {
    schema_version: u32,
    entries: Vec<IndexEntry>,
}

fn index_path(project: &Path) -> PathBuf {
    project.join(ARCHIVE_DIR).join("index.json")
}

/// The index and its raw bytes (the publish re-checks them). Missing means empty; an
/// unreadable or malformed index is an error, never overwritten on a guess.
fn read_index(project: &Path) -> Result<(Index, Option<Vec<u8>>)> {
    let before = crate::apply::read_opt(&index_path(project))?;
    let index = match &before {
        None => Index {
            schema_version: INDEX_SCHEMA_VERSION,
            entries: Vec::new(),
        },
        Some(bytes) => {
            let idx: Index =
                serde_json::from_slice(bytes).map_err(|_| "archive index invalid".to_string())?;
            if idx.schema_version != INDEX_SCHEMA_VERSION {
                return Err("archive index invalid: unknown schema_version".into());
            }
            idx
        }
    };
    Ok((index, before))
}

fn write_index(project: &Path, index: &Index, before: Option<&[u8]>) -> Result<()> {
    let bytes = serde_json::to_vec_pretty(index).map_err(|_| "encode failed".to_string())?;
    crate::apply::publish_replace(&index_path(project), before, &bytes)
}

fn manifest_path(project: &Path, spec_id: &str) -> PathBuf {
    project
        .join(ARCHIVE_DIR)
        .join("specs")
        .join(format!("{spec_id}.manifest.json"))
}

fn permission_ok(permission: &str) -> Result<()> {
    // Authorization stays with the host; the reference is caller metadata, not a verifier.
    if !crate::contract::permission_reference_ok(permission) {
        return Err("permission reference required".into());
    }
    Ok(())
}

/// Digest of the plan's eligible list only; `now` and `cutoff` change on every run.
fn eligible_digest(eligible: &[EligibleSpec]) -> String {
    crate::digest(&serde_json::to_vec(eligible).unwrap_or_default())
}

/// REQ-D2-01: recomputes the plan and, only when it matches `plan_digest`, moves each
/// eligible SPEC as hash → copy → readback → manifest and index → source removal. A
/// failing SPEC keeps its original and stops the run (earlier moves are reported as
/// `partial apply`).
pub fn apply(
    project: &Path,
    now: SystemTime,
    grace_request: Option<u64>,
    plan_digest: &str,
    permission: &str,
) -> Result<Value> {
    apply_with_before_remove(project, now, grace_request, plan_digest, permission, |_| {})
}

/// Test seam after index publication and before the final source check.
#[doc(hidden)]
pub fn apply_with_before_remove(
    project: &Path,
    now: SystemTime,
    grace_request: Option<u64>,
    plan_digest: &str,
    permission: &str,
    mut before_remove: impl FnMut(&Path),
) -> Result<Value> {
    permission_ok(permission)?;
    let current = plan(project, now, grace_request)?;
    if current.plan_digest != plan_digest {
        return Err("stale or altered plan; re-preview required".into());
    }
    let specs = project.join(crate::PROJECT_STATE_DIR).join("specs");
    let mut archived: Vec<Value> = Vec::new();
    let mut done: Vec<String> = Vec::new();
    for spec in &current.eligible {
        let id = &spec.spec_id;
        let wrap = |done: &[String], e: String| {
            if done.is_empty() {
                e
            } else {
                format!(
                    "partial apply: archived [{}]; failed at {id}: {e}",
                    done.join(", ")
                )
            }
        };
        let step = (|| -> Result<(IndexEntry, Manifest)> {
            let src = specs.join(id);
            let rel = format!("{ARCHIVE_DIR}/specs/{id}");
            let dst = project.join(&rel);
            let (index, before) = read_index(project)?;
            if index.entries.iter().any(|e| &e.spec_id == id) {
                return Err("archive index entry already exists; exclusive create refused".into());
            }
            // 1) hash before the move
            let m = manifest(&src)?;
            // 2)-4) copy, readback, source re-check
            copy_verified(&src, &dst, &m)?;
            // 5) manifest file and index (the pointer)
            let m_bytes = serde_json::to_vec_pretty(&m).map_err(|_| "encode failed".to_string())?;
            crate::apply::publish_exclusive(&manifest_path(project, id), &m_bytes)?;
            let entry = IndexEntry {
                spec_id: id.clone(),
                status: spec.status.clone(),
                archived_at: rfc3339(
                    SystemTime::now()
                        .duration_since(UNIX_EPOCH)
                        .map(|d| d.as_secs())
                        .unwrap_or(0),
                ),
                path: rel,
                manifest_digest: m.digest(),
                file_count: m.files.len(),
            };
            let mut next = index.clone();
            next.entries.push(entry.clone());
            next.entries.sort_by(|a, b| a.spec_id.cmp(&b.spec_id));
            write_index(project, &next, before.as_deref())?;
            Ok((entry, m))
        })();
        let (entry, expected) = step.map_err(|e| wrap(&done, e))?;
        let src = specs.join(id);
        before_remove(&src);
        if manifest(&src).ok().as_ref() != Some(&expected) {
            done.push(id.clone());
            return Err(format!(
                "partial apply: archived [{}]; failed at {id}: source changed during apply; source removal refused; archived copy and index kept",
                done.join(", ")
            ));
        }
        // 6) remove the source only after the verified copy and its pointer exist.
        // This final check narrows the race; concurrent writers are not locked out.
        if fs::remove_dir_all(&src).is_err() {
            done.push(id.clone());
            return Err(format!(
                "partial apply: archived [{}]; failed at {id}: source directory removal write failed; archived copy and index kept",
                done.join(", ")
            ));
        }
        done.push(id.clone());
        archived.push(serde_json::to_value(&entry).map_err(|_| "encode failed".to_string())?);
    }
    Ok(json!({
        "state": if archived.is_empty() { "unchanged" } else { "applied" },
        "archived": archived,
        "index": format!("{ARCHIVE_DIR}/index.json"),
        "network": "disabled",
    }))
}

/// A SPEC id must be one path component (no traversal).
fn spec_id_ok(id: &str) -> bool {
    !id.is_empty()
        && !id.starts_with('.')
        && id
            .bytes()
            .all(|b| b.is_ascii_alphanumeric() || matches!(b, b'-' | b'_' | b'.'))
}

/// Restores one archived SPEC to `.moai-cli/specs/<id>`. The manifest file must match
/// the index digest and the restored copy's readback must match the manifest before
/// it is published; then the index entry and the archived copy are removed.
pub fn restore(project: &Path, spec_id: &str, permission: &str) -> Result<Value> {
    permission_ok(permission)?;
    if !spec_id_ok(spec_id) {
        return Err(format!("invalid spec id: {spec_id}"));
    }
    let (index, before) = read_index(project)?;
    let entry = index
        .entries
        .iter()
        .find(|e| e.spec_id == spec_id)
        .cloned()
        .ok_or_else(|| format!("spec not found in archive index: {spec_id}"))?;
    let src = project.join(ARCHIVE_DIR).join("specs").join(spec_id);
    let dst = project
        .join(crate::PROJECT_STATE_DIR)
        .join("specs")
        .join(spec_id);
    if fs::symlink_metadata(&dst).is_ok() {
        return Err("restore destination already exists; exclusive create refused".into());
    }
    let m_bytes = crate::apply::read_opt(&manifest_path(project, spec_id))?
        .ok_or("archive manifest not found")?;
    let m: Manifest =
        serde_json::from_slice(&m_bytes).map_err(|_| "archive manifest invalid".to_string())?;
    if m.digest() != entry.manifest_digest {
        return Err("archive manifest digest mismatch; restore refused".into());
    }
    // An archived copy that no longer matches its manifest (tampered or damaged) is not restored.
    if manifest(&src)? != m {
        return Err("archived copy digest mismatch; restore refused".into());
    }
    copy_verified(&src, &dst, &m)?;
    let mut next = index.clone();
    next.entries.retain(|e| e.spec_id != spec_id);
    let cleanup = write_index(project, &next, before.as_deref())
        .and_then(|_| {
            fs::remove_file(manifest_path(project, spec_id))
                .map_err(|_| "archive manifest removal write failed".to_string())
        })
        .and_then(|_| {
            fs::remove_dir_all(&src).map_err(|_| "archived copy removal write failed".to_string())
        });
    if let Err(e) = cleanup {
        return Err(format!(
            "partial apply: restored [{spec_id}]; failed at archive cleanup: {e}"
        ));
    }
    Ok(json!({
        "state": "restored",
        "spec_id": spec_id,
        "path": format!("{}/specs/{spec_id}", crate::PROJECT_STATE_DIR),
        "manifest_digest": entry.manifest_digest,
        "file_count": entry.file_count,
    }))
}

/// Non-vector archive search: case-insensitive substring match over the index's SPEC
/// id and status and over the lines of archived text files. Writes nothing.
pub fn search(project: &Path, query: &str) -> Result<Value> {
    let n = query.chars().count();
    if n == 0 || n > SEARCH_QUERY_CHARS_MAX {
        return Err(format!(
            "search query must be 1 to {SEARCH_QUERY_CHARS_MAX} characters"
        ));
    }
    let needle = query.to_lowercase();
    let (index, _) = read_index(project)?;
    let mut hits: Vec<Value> = Vec::new();
    let mut truncated = false;
    let mut push = |hit: Value, hits: &mut Vec<Value>| -> bool {
        if hits.len() == SEARCH_RESULTS_MAX {
            truncated = true;
            return false;
        }
        hits.push(hit);
        true
    };
    'outer: for e in &index.entries {
        if !spec_id_ok(&e.spec_id) {
            return Err("archive index invalid: invalid spec id".into());
        }
        let rel = format!("{ARCHIVE_DIR}/specs/{}", e.spec_id);
        if e.path != rel {
            return Err("archive index invalid: spec path mismatch".into());
        }
        if e.spec_id.to_lowercase().contains(&needle) || e.status.to_lowercase().contains(&needle) {
            let hit = json!({"spec_id": e.spec_id, "source": "index", "path": null, "line": null});
            if !push(hit, &mut hits) {
                break 'outer;
            }
        }
        let dir = project.join(ARCHIVE_DIR).join("specs").join(&e.spec_id);
        let Ok(m) = manifest(&dir) else {
            continue;
        };
        for f in &m.files {
            let Ok(text) = fs::read_to_string(dir.join(&f.path)) else {
                // Non-text files are not searched.
                continue;
            };
            for (i, line) in text.lines().enumerate() {
                if line.to_lowercase().contains(&needle) {
                    let hit = json!({"spec_id": e.spec_id, "source": "file", "path": f.path, "line": i + 1});
                    if !push(hit, &mut hits) {
                        break 'outer;
                    }
                }
            }
        }
    }
    Ok(json!({
        "query": query,
        "hits": hits,
        "truncated": truncated,
        "writes": "none",
        "network": "disabled",
    }))
}
