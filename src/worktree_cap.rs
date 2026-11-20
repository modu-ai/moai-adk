//! SPEC-20261120-worktree-cap-and-cleanup: the worktree cap and the cleanup of finished
//! worktrees, behind `optimize-plan` (preview) and `optimize-apply` (removal).
//!
//! The cap is only shown (D-7): the preview says how many worktrees the project has against it.
//! A worktree is a cleanup candidate when its branch is merged into the integration ref or the
//! card that reserved it is `done`, it has no tracked change, no untracked file, no ignored
//! file that is not regenerable, no git lock, no live card lease or session, and no process
//! working inside it. The facts come from `worktree_gc::plan_paced` (the one place that knows
//! how to read a worktree); this layer re-judges its rows because that plan keeps the old
//! `.claude/worktrees` folder (`preserved_location`) and does not know cards or nesting. Codex
//! worktrees and the Go generation's folder stay kept.
//!
//! Applying removes one picked worktree at a time with `git worktree remove` (no `--force`, so
//! git refuses a changed worktree itself), deepest first, after judging the item again.

use crate::identity::scrub_git_env;
use crate::worktree_gc::{self, Pace};
use crate::worktree_report::{disk_bytes, parse_worktree_list};
use serde_json::{json, Value};
use std::collections::{BTreeMap, BTreeSet};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::Command;
use std::time::{Duration, Instant};

type Result<T> = std::result::Result<T, String>;

/// Worktrees per project before the preview says the project is over the cap.
pub const DEFAULT_CAP: u64 = 40;
/// The HOME settings file (`~/.moai/hook-config.toml`) and the key the cap is read from.
pub const CONFIG_FILE: &str = "hook-config.toml";
pub const CONFIG_KEY: &str = "[worktrees] cap";
const CONFIG_MAX: u64 = 65_536;

/// What the sizes mean and what the freed space means (REQ-4, REQ-9).
pub const SPACE_BASIS: &str = "sizes are allocated blocks (st_blocks x 512), not du; APFS clones make their sum larger than what is freed, so the space freed is only the change of the df free bytes after an apply";

/// The cap in force and where it came from.
#[derive(Debug, PartialEq, Eq)]
pub struct Cap {
    pub value: u64,
    /// `flag`, `config` or `default`.
    pub source: &'static str,
}

/// A whole number of at least 1; everything else (0, a sign, a fraction, text, a number past
/// `u64`) is refused.
pub fn parse_cap(raw: &str) -> Option<u64> {
    if raw.is_empty() || !raw.bytes().all(|b| b.is_ascii_digit()) {
        return None;
    }
    raw.parse::<u64>().ok().filter(|n| *n >= 1)
}

/// The cap: `--worktree-cap`, else `[worktrees] cap` of the HOME `hook-config.toml`, else
/// [`DEFAULT_CAP`]. A value that is set and unusable is refused with its name, never skipped.
pub fn resolve_cap(flag: Option<&str>, home: Option<&Path>) -> Result<Cap> {
    if let Some(raw) = flag {
        return parse_cap(raw)
            .map(|value| Cap {
                value,
                source: "flag",
            })
            .ok_or_else(|| {
                format!(
                    "invalid --worktree-cap: must be a whole number of at least 1 (got {raw:?})"
                )
            });
    }
    if let Some(cap) = home.and_then(config_cap) {
        return cap.map(|value| Cap {
            value,
            source: "config",
        });
    }
    Ok(Cap {
        value: DEFAULT_CAP,
        source: "default",
    })
}

/// `Some(Ok(n))` / `Some(Err)` when the HOME file sets the key, `None` when it does not (no
/// file, an unreadable or oversized file and a file that is not TOML count as not set: the hook
/// engine reports those itself).
fn config_cap(home: &Path) -> Option<Result<u64>> {
    let path = crate::home_root::path(home).join(CONFIG_FILE);
    let meta = fs::symlink_metadata(&path).ok()?;
    if !meta.is_file() || meta.len() > CONFIG_MAX {
        return None;
    }
    let doc = fs::read_to_string(&path)
        .ok()?
        .parse::<toml_edit::DocumentMut>()
        .ok()?;
    let item = doc.get("worktrees")?.as_table_like()?.get("cap")?;
    let bad =
        || format!("invalid {CONFIG_KEY} in {CONFIG_FILE}: must be a whole number of at least 1");
    Some(match item.as_integer() {
        Some(n) if n >= 1 => Ok(n as u64),
        _ => Err(bad()),
    })
}

/// The `worktree_cap` object of the preview.
pub fn cap_report(count: usize, cap: &Cap) -> Value {
    let over_by = (count as u64).saturating_sub(cap.value);
    json!({
        "count": count,
        "cap": cap.value,
        "source": cap.source,
        "over": over_by > 0,
        "over_by": over_by,
        "note": "the cap is shown, never enforced: it blocks nothing and removes nothing",
    })
}

fn git(dir: &Path, args: &[&str]) -> Result<String> {
    let mut cmd = Command::new("git");
    scrub_git_env(&mut cmd);
    let out = cmd
        .arg("--no-optional-locks")
        .arg("-C")
        .arg(dir)
        .args(args)
        .output()
        .map_err(|e| format!("git could not start: {e}"))?;
    if out.status.success() {
        Ok(String::from_utf8_lossy(&out.stdout).into_owned())
    } else {
        let err = String::from_utf8_lossy(&out.stderr);
        Err(err
            .lines()
            .next()
            .unwrap_or("git failed")
            .trim()
            .to_string())
    }
}

fn canon(p: &Path) -> PathBuf {
    fs::canonicalize(p).unwrap_or_else(|_| p.to_path_buf())
}

/// Bytes allocated on disk (`st_blocks` x 512) under `root`, symlinks not followed, a hard
/// linked file once, the folders in `skip` (worktrees nested inside) left out; `true` when the
/// walk was cut short by `deadline` (the figure is then a lower bound).
pub fn allocated_bytes(root: &Path, skip: &BTreeSet<PathBuf>, deadline: Instant) -> (u64, bool) {
    let (mut total, mut partial) = (0u64, false);
    let mut seen = BTreeSet::new();
    let mut stack = vec![root.to_path_buf()];
    while let Some(dir) = stack.pop() {
        if Instant::now() >= deadline {
            return (total, true);
        }
        let Ok(rd) = fs::read_dir(&dir) else {
            partial = true;
            continue;
        };
        for entry in rd.flatten() {
            let path = entry.path();
            let Ok(meta) = fs::symlink_metadata(&path) else {
                partial = true;
                continue;
            };
            if meta.is_dir() {
                if !skip.contains(&path) {
                    stack.push(path);
                }
                total += blocks(&meta);
            } else {
                #[cfg(unix)]
                {
                    use std::os::unix::fs::MetadataExt;
                    if meta.nlink() > 1 && !seen.insert((meta.dev(), meta.ino())) {
                        continue;
                    }
                }
                let _ = &mut seen;
                total += blocks(&meta);
            }
        }
    }
    (total, partial)
}

#[cfg(unix)]
fn blocks(meta: &fs::Metadata) -> u64 {
    use std::os::unix::fs::MetadataExt;
    meta.blocks() * 512
}

#[cfg(not(unix))]
fn blocks(meta: &fs::Metadata) -> u64 {
    meta.len()
}

/// What the card ledger says about one worktree.
#[derive(Debug, Clone)]
struct Reservation {
    card_state: String,
    status: String,
    lease_until: i64,
    holder: String,
}

/// Reservations by worktree path and by branch; `None` when no ledger could be asked (no HOME,
/// no profile, no ledger yet), which the preview says (`ledger_observed`).
struct Cards {
    by_path: BTreeMap<PathBuf, Vec<Reservation>>,
    by_branch: BTreeMap<String, Vec<Reservation>>,
}

fn read_cards(project: &Path, home: Option<&Path>) -> Result<Option<Cards>> {
    let Some(home) = home else {
        return Ok(None);
    };
    let Ok(Some(id)) = crate::project_id::read(&canon(project)) else {
        return Ok(None);
    };
    let ledger = match crate::ledger::open_read_only(home, &id) {
        Ok(l) => l,
        Err(crate::ledger::LedgerError::Absent) => return Ok(None),
        // a ledger that is there and cannot be read: no candidate, never a guess
        Err(e) => return Err(format!("card ledger unreadable: {e}")),
    };
    let mut cards = Cards {
        by_path: BTreeMap::new(),
        by_branch: BTreeMap::new(),
    };
    let unreadable = |e: rusqlite::Error| format!("card ledger unreadable: {e}");
    let mut stmt = ledger
        .conn()
        .prepare(
            "SELECT c.state, r.status, r.lease_until, r.holder, r.worktree, r.branch \
             FROM reservation r JOIN card c ON c.card_id = r.card_id \
             WHERE r.worktree IS NOT NULL OR r.branch IS NOT NULL ORDER BY r.reservation_id",
        )
        .map_err(unreadable)?;
    let rows = stmt
        .query_map([], |r| {
            Ok((
                Reservation {
                    card_state: r.get(0)?,
                    status: r.get(1)?,
                    lease_until: r.get(2)?,
                    holder: r.get(3)?,
                },
                r.get::<_, Option<String>>(4)?,
                r.get::<_, Option<String>>(5)?,
            ))
        })
        .map_err(unreadable)?;
    for row in rows {
        let (res, worktree, branch) = row.map_err(unreadable)?;
        if let Some(w) = worktree {
            cards
                .by_path
                .entry(canon(Path::new(&w)))
                .or_default()
                .push(res.clone());
        }
        if let Some(b) = branch {
            cards.by_branch.entry(b).or_default().push(res);
        }
    }
    Ok(Some(cards))
}

const ACTIVE: [&str; 3] = ["held", "expired_candidate", "held_for_review"];

impl Cards {
    fn facts(&self, path: &Path, branch: Option<&str>) -> Vec<&Reservation> {
        let mut out: Vec<&Reservation> = self.by_path.get(path).into_iter().flatten().collect();
        if let Some(b) = branch {
            out.extend(self.by_branch.get(b).into_iter().flatten());
        }
        out
    }
}

/// One cleanup candidate or kept worktree, before it becomes JSON.
struct Judged {
    path: PathBuf,
    shown: String,
    branch: Option<String>,
    head: Option<String>,
    location: String,
    reasons: Vec<String>,
    because: Option<&'static str>,
    children: Vec<PathBuf>,
}

/// Reasons the gc plan computes after its first hard stop; this layer recomputes them from the
/// row's census fields so a row that stopped early is still judged on every fact.
const CENSUS: [&str; 6] = [
    "untracked_files",
    "ignored_files_outside_build_output",
    "file_census_unknown",
    "no_build_output",
    "empty_worktree",
    "install_folder",
];

fn str_list(v: &Value) -> Vec<String> {
    v.as_array()
        .map(|a| {
            a.iter()
                .filter_map(|x| x.as_str().map(String::from))
                .collect()
        })
        .unwrap_or_default()
}

fn judge_row(
    row: &Value,
    abs: &Path,
    registered: &BTreeSet<PathBuf>,
    cards: Option<&Cards>,
    project: &Path,
    now: i64,
) -> Judged {
    let branch = row["branch"].as_str().map(String::from);
    let location = row["location"].as_str().unwrap_or("other").to_string();
    let merged = !row["merged_by"].is_null();
    let facts = cards
        .map(|c| c.facts(abs, branch.as_deref()))
        .unwrap_or_default();
    // the newest reservation names the card's state
    let card_done = facts.last().is_some_and(|r| r.card_state == "done");
    let mut reasons: Vec<String> = Vec::new();
    for r in str_list(&row["reasons"]) {
        let drop = (r == "preserved_location" && location == "claude")
            || ((r == "not_merged" || r == "merged_unknown") && card_done)
            || CENSUS.contains(&r.as_str());
        if !drop {
            reasons.push(r);
        }
    }
    // nested worktrees: every registered worktree below this one
    let children: Vec<PathBuf> = registered
        .iter()
        .filter(|p| p.as_path() != abs && p.starts_with(abs))
        .cloned()
        .collect();
    let direct: BTreeSet<String> = children
        .iter()
        .filter(|c| c.parent() == Some(abs.join("target").as_path()))
        .filter_map(|c| c.file_name().map(|n| n.to_string_lossy().into_owned()))
        .collect();
    let in_target: BTreeSet<String> = str_list(&row["preserved_in_target"]).into_iter().collect();
    match row["untracked"].as_u64() {
        Some(0) => {}
        Some(_) => reasons.push("untracked_files".into()),
        None => reasons.push("file_census_unknown".into()),
    }
    match row["ignored_outside_build"].as_u64() {
        Some(0) => {}
        // the only things outside the build output are worktrees of their own (checked below)
        Some(n)
            if n as usize == in_target.len() && !in_target.is_empty() && in_target == direct => {}
        Some(_) => reasons.push("ignored_files_outside_build_output".into()),
        None => {
            if !reasons.contains(&"file_census_unknown".to_string()) {
                reasons.push("file_census_unknown".into());
            }
        }
    }
    if fs::symlink_metadata(abs.join("node_modules")).is_ok_and(|m| m.is_dir()) {
        reasons.push("install_folder".into());
    }
    if cards.is_some() {
        for r in facts.iter().filter(|r| ACTIVE.contains(&r.status.as_str())) {
            if r.lease_until > now {
                reasons.push("live_claim".into());
            } else {
                match crate::lane_registry::session_live(project, &r.holder, now) {
                    Some(false) => {}
                    Some(true) => reasons.push("live_session".into()),
                    None => reasons.push("session_unobserved".into()),
                }
            }
        }
    }
    reasons.sort();
    reasons.dedup();
    let because = if !reasons.is_empty() {
        None
    } else if merged {
        Some("merged")
    } else if card_done {
        Some("card_done")
    } else {
        None
    };
    Judged {
        path: abs.to_path_buf(),
        shown: row["path"].as_str().unwrap_or_default().to_string(),
        branch,
        head: row["head"].as_str().map(String::from),
        location,
        reasons,
        because,
        children,
    }
}

fn item_digest(j: &Judged) -> String {
    // children are not in the digest: removing them first must not change the parent's
    let bound = json!({
        "path": j.path, "branch": j.branch, "head": j.head,
        "because": j.because, "location": j.location,
    });
    crate::digest(serde_json::to_string(&bound).unwrap_or_default().as_bytes())
}

/// What a preview or one fresh check produces.
struct Assessment {
    judged: Vec<Judged>,
    primary: PathBuf,
    registered: BTreeSet<PathBuf>,
    count: usize,
    ledger_observed: bool,
}

/// Reads the worktrees of `project` and judges each linked one (`only`: just those). A worktree
/// with worktrees below it stays kept unless `child_ok` says each of them goes too.
#[allow(clippy::too_many_arguments)]
fn assess(
    project: &Path,
    integration_ref: &str,
    recent: Duration,
    budget: Duration,
    now_unix: Option<u64>,
    home: Option<&Path>,
    only: &[PathBuf],
    all_children_go: bool,
) -> Result<Assessment> {
    let listing = git(project, &["worktree", "list", "--porcelain"])?;
    let entries = parse_worktree_list(&listing);
    let primary = entries
        .first()
        .map(|e| canon(&e.path))
        .ok_or("git worktree list is empty")?;
    let registered: BTreeSet<PathBuf> = entries
        .iter()
        .filter(|e| !e.bare)
        .map(|e| canon(&e.path))
        .collect();
    let count = entries.iter().filter(|e| !e.bare).count();
    let cards = read_cards(project, home)?;
    let gc = worktree_gc::plan_paced(
        project,
        integration_ref,
        recent,
        budget,
        now_unix,
        true,
        &Pace {
            only: only.to_vec(),
            ..Pace::default()
        },
    )?;
    let now = gc["now_unix"].as_u64().unwrap_or_default() as i64;
    let mut judged: Vec<Judged> = Vec::new();
    for row in gc["candidates"].as_array().into_iter().flatten() {
        let rel = row["path"].as_str().unwrap_or_default();
        if rel == "." {
            continue;
        }
        let abs = if Path::new(rel).is_absolute() {
            PathBuf::from(rel)
        } else {
            primary.join(rel)
        };
        judged.push(judge_row(
            row,
            &abs,
            &registered,
            cards.as_ref(),
            project,
            now,
        ));
    }
    // inside out: a worktree with worktrees below it is a candidate only when each of them is
    judged.sort_by_key(|j| std::cmp::Reverse(j.path.components().count()));
    let mut going: BTreeSet<PathBuf> = BTreeSet::new();
    for j in &mut judged {
        if j.because.is_some() && !j.children.is_empty() {
            let ok = all_children_go && j.children.iter().all(|c| going.contains(c));
            if !ok {
                j.reasons.push("nested_worktree_kept".into());
                j.because = None;
            }
        }
        if j.because.is_some() {
            going.insert(j.path.clone());
        }
    }
    judged.sort_by(|a, b| a.shown.cmp(&b.shown));
    Ok(Assessment {
        judged,
        primary,
        registered,
        count,
        ledger_observed: cards.is_some(),
    })
}

/// The `worktree_cap` and `cleanup` objects of `optimize-plan`. Reads only.
pub fn preview(
    project: &Path,
    integration_ref: &str,
    recent: Duration,
    budget: Duration,
    now_unix: Option<u64>,
    home: Option<&Path>,
    cap: &Cap,
) -> Result<(Value, Value)> {
    let deadline = Instant::now() + budget;
    let a = assess(
        project,
        integration_ref,
        recent,
        budget,
        now_unix,
        home,
        &[],
        true,
    )?;
    let mut candidates = Vec::new();
    let mut kept = Vec::new();
    let mut total = 0u64;
    let mut any_partial = false;
    for j in &a.judged {
        if let Some(because) = j.because {
            let skip: BTreeSet<PathBuf> = a
                .registered
                .iter()
                .filter(|p| p.as_path() != j.path && p.starts_with(&j.path))
                .cloned()
                .collect();
            let (bytes, partial) = allocated_bytes(&j.path, &skip, deadline);
            total += bytes;
            any_partial |= partial;
            let nested_in = a
                .judged
                .iter()
                .filter(|p| p.path != j.path && j.path.starts_with(&p.path))
                .max_by_key(|p| p.path.components().count())
                .map(|p| p.shown.clone());
            candidates.push(json!({
                "path": j.shown,
                "branch": j.branch,
                "head": j.head,
                "location": j.location,
                "remove_because": because,
                "allocated_bytes": bytes,
                "bytes_partial": partial,
                "item_digest": item_digest(j),
                "nested_in": nested_in,
                "nested_children": j.children.iter().map(|c| worktree_shown(c, &a.primary)).collect::<Vec<_>>(),
                "restore": restore_command(&j.path, j.branch.as_deref(), j.head.as_deref()),
            }));
        } else {
            kept.push(json!({
                "path": j.shown,
                "branch": j.branch,
                "location": j.location,
                "reasons": j.reasons,
            }));
        }
    }
    let cleanup = json!({
        "candidates": candidates,
        "kept": kept,
        "reclaimable_allocated_bytes": total,
        "bytes_partial": any_partial,
        "space_basis": SPACE_BASIS,
        "before_free_bytes": disk_bytes(&a.primary).map(|(free, _)| free),
        "ledger_observed": a.ledger_observed,
        "apply": "optimize-apply PROJECT --integration-ref REF --pick PATH@item_digest (repeat) --permission-reference R, after the user picks; each pick is judged again and removed with git worktree remove (no --force), the deepest first; branches are kept",
        "network": "disabled",
        "writes": false,
    });
    Ok((cap_report(a.count, cap), cleanup))
}

fn worktree_shown(path: &Path, primary: &Path) -> String {
    path.strip_prefix(primary)
        .map(|r| r.to_string_lossy().into_owned())
        .unwrap_or_else(|_| path.to_string_lossy().into_owned())
}

fn quote(s: &str) -> String {
    format!("'{}'", s.replace('\'', "'\\''"))
}

/// The command that brings a removed worktree back on its branch (ignored files and build
/// output are not restored).
fn restore_command(path: &Path, branch: Option<&str>, head: Option<&str>) -> String {
    let p = quote(&path.to_string_lossy());
    match (branch, head) {
        (Some(b), _) => format!("git worktree add {p} {}", quote(b)),
        (None, Some(h)) => format!("git worktree add --detach {p} {h}"),
        (None, None) => format!("git worktree add {p}"),
    }
}

/// Removes the picked worktrees (`path@item_digest`). Nothing is removed unless the permission
/// reference is usable; every pick is judged again just before its removal.
pub fn apply(
    project: &Path,
    integration_ref: &str,
    recent: Duration,
    budget: Duration,
    now_unix: Option<u64>,
    home: Option<&Path>,
    picks: &[String],
    permission_reference: &str,
) -> Result<Value> {
    if !crate::contract::permission_reference_ok(permission_reference) {
        return Err("invalid permission reference; nothing was removed".into());
    }
    let mut parsed: Vec<(String, String)> = Vec::new();
    for p in picks {
        let (path, digest) = p
            .rsplit_once('@')
            .filter(|(a, b)| !a.is_empty() && !b.is_empty())
            .ok_or_else(|| format!("invalid --pick {p:?}: expected PATH@item_digest"))?;
        parsed.push((path.to_string(), digest.to_string()));
    }
    let listing = git(project, &["worktree", "list", "--porcelain"])?;
    let primary = parse_worktree_list(&listing)
        .first()
        .map(|e| canon(&e.path))
        .ok_or("git worktree list is empty")?;
    let common = git(project, &["rev-parse", "--git-common-dir"])?;
    let common = PathBuf::from(common.trim());
    let common = canon(&if common.is_absolute() {
        common
    } else {
        project.join(common)
    });
    // the deepest first, so a worktree nested in another goes before it
    let mut order: Vec<(PathBuf, String, String)> = parsed
        .into_iter()
        .map(|(rel, d)| {
            let abs = if Path::new(&rel).is_absolute() {
                PathBuf::from(&rel)
            } else {
                primary.join(&rel)
            };
            (canon(&abs), rel, d)
        })
        .collect();
    order.sort_by_key(|(abs, _, _)| std::cmp::Reverse(abs.components().count()));

    let before = disk_bytes(&primary).map(|(free, _)| free);
    let (mut removed, mut refused) = (Vec::new(), Vec::new());
    for (abs, rel, digest) in &order {
        let refuse =
            |reason: &str, detail: Value| json!({"path": rel, "reason": reason, "detail": detail});
        let fresh = match assess(
            project,
            integration_ref,
            recent,
            budget,
            now_unix,
            home,
            std::slice::from_ref(abs),
            false,
        ) {
            Ok(a) => a,
            Err(e) => {
                refused.push(refuse("not_a_candidate", json!(e)));
                continue;
            }
        };
        let Some(j) = fresh.judged.iter().find(|j| &j.path == abs) else {
            refused.push(refuse(
                "not_a_candidate",
                json!("not a linked worktree of this clone"),
            ));
            continue;
        };
        if j.because.is_none() {
            refused.push(refuse("not_a_candidate", json!(j.reasons)));
            continue;
        }
        if &item_digest(j) != digest {
            refused.push(refuse("digest_mismatch", Value::Null));
            continue;
        }
        let restore = restore_command(&j.path, j.branch.as_deref(), j.head.as_deref());
        match git(&primary, &["worktree", "remove", &abs.to_string_lossy()]) {
            Ok(_) => removed.push(json!({
                "path": j.shown,
                "branch": j.branch,
                "head": j.head,
                "because": j.because,
                "restore": restore,
            })),
            Err(e) => refused.push(refuse("remove_failed", json!(e))),
        }
    }
    let after = disk_bytes(&primary).map(|(free, _)| free);
    let delta = match (before, after) {
        (Some(b), Some(a)) => json!(a as i64 - b as i64),
        _ => Value::Null,
    };
    let now = now_unix.unwrap_or_else(|| {
        std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .map(|d| d.as_secs())
            .unwrap_or_default()
    });
    let record = if removed.is_empty() {
        Value::Null
    } else {
        let names: Vec<&str> = removed.iter().filter_map(|r| r["path"].as_str()).collect();
        let body = json!({
            "schema_version": 1,
            "now_unix": now,
            "permission_reference": permission_reference,
            "items": removed,
            "refused": refused,
            "free_bytes_before": before,
            "free_bytes_after": after,
        });
        match write_record(&common, now, &body) {
            Ok(path) => json!(path),
            Err(e) => {
                return Err(format!(
                    "removed {} worktree(s) but the record could not be written ({e}); removed: {}",
                    names.len(),
                    names.join(", ")
                ))
            }
        }
    };
    Ok(json!({
        "removed": removed,
        "refused": refused,
        "record": record,
        "free_bytes_before": before,
        "free_bytes_after": after,
        "free_bytes_delta": delta,
        "space_basis": SPACE_BASIS,
        "branches": "kept",
        "writes": !removed.is_empty(),
    }))
}

/// `<git folder>/.moai-cli/worktree-cleanup/<now>-<digest12>.json`, created new; a link in the
/// way is refused.
fn write_record(common: &Path, now: u64, body: &Value) -> Result<String> {
    let text = serde_json::to_string_pretty(body).map_err(|e| e.to_string())?;
    let dir = common.join(".moai-cli").join("worktree-cleanup");
    for d in [common.join(".moai-cli"), dir.clone()] {
        if fs::symlink_metadata(&d).is_ok_and(|m| m.file_type().is_symlink()) {
            return Err(format!("{} is a link", d.display()));
        }
    }
    fs::create_dir_all(&dir).map_err(|e| format!("{}: {e}", dir.display()))?;
    let name = format!("{now}-{}.json", &crate::digest(text.as_bytes())[..12]);
    let path = dir.join(name);
    {
        use std::io::Write;
        let mut f = fs::OpenOptions::new()
            .write(true)
            .create_new(true)
            .open(&path)
            .map_err(|e| format!("{}: {e}", path.display()))?;
        f.write_all(text.as_bytes())
            .map_err(|e| format!("{}: {e}", path.display()))?;
    }
    Ok(path.to_string_lossy().into_owned())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn the_cap_is_a_whole_number_of_at_least_one() {
        assert_eq!(parse_cap("1"), Some(1));
        assert_eq!(parse_cap("40"), Some(40));
        for bad in [
            "0",
            "-1",
            "+1",
            "1.5",
            "",
            " 1",
            "abc",
            "99999999999999999999999",
        ] {
            assert_eq!(parse_cap(bad), None, "{bad:?}");
        }
    }

    #[test]
    fn at_the_cap_is_not_over_and_one_past_it_is() {
        let cap = Cap {
            value: 40,
            source: "default",
        };
        assert_eq!(cap_report(40, &cap)["over"], false);
        assert_eq!(cap_report(41, &cap)["over_by"], 1);
        assert_eq!(cap_report(0, &cap)["over_by"], 0);
    }

    #[test]
    fn the_flag_beats_the_default_and_a_bad_flag_is_named() {
        let c = resolve_cap(Some("7"), None).unwrap();
        assert_eq!((c.value, c.source), (7, "flag"));
        let d = resolve_cap(None, None).unwrap();
        assert_eq!((d.value, d.source), (DEFAULT_CAP, "default"));
        assert!(resolve_cap(Some("0"), None)
            .unwrap_err()
            .contains("--worktree-cap"));
    }

    #[test]
    fn the_home_file_sets_the_cap_and_a_bad_value_is_named() {
        let home = tempfile::TempDir::new().unwrap();
        let root = crate::home_root::path(home.path());
        fs::create_dir_all(&root).unwrap();
        fs::write(root.join(CONFIG_FILE), "[worktrees]\ncap = 12\n").unwrap();
        let c = resolve_cap(None, Some(home.path())).unwrap();
        assert_eq!((c.value, c.source), (12, "config"));
        // the flag still wins
        assert_eq!(resolve_cap(Some("3"), Some(home.path())).unwrap().value, 3);
        for bad in ["cap = 0", "cap = -2", "cap = \"x\"", "cap = 1.5"] {
            fs::write(root.join(CONFIG_FILE), format!("[worktrees]\n{bad}\n")).unwrap();
            let e = resolve_cap(None, Some(home.path())).unwrap_err();
            assert!(e.contains(CONFIG_KEY), "{bad}: {e}");
        }
        // a file that sets nothing, or is not TOML, leaves the default
        fs::write(root.join(CONFIG_FILE), "[hooks]\n").unwrap();
        assert_eq!(
            resolve_cap(None, Some(home.path())).unwrap().source,
            "default"
        );
        fs::write(root.join(CONFIG_FILE), "not = [toml").unwrap();
        assert_eq!(
            resolve_cap(None, Some(home.path())).unwrap().source,
            "default"
        );
    }

    #[test]
    fn the_restore_command_quotes_the_path_and_names_the_branch() {
        let c = restore_command(Path::new("/tmp/a b/it's"), Some("issue/1-x"), None);
        assert_eq!(c, "git worktree add '/tmp/a b/it'\\''s' 'issue/1-x'");
        assert!(restore_command(Path::new("/x"), None, Some("abc")).contains("--detach"));
    }
}
