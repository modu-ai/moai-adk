//! `tidy-apply` (SPEC-20261126-home-tidy-apply): act on the picked items of a freshly
//! recomputed plan. An item is removed only after its archive was read back and matched both
//! the list and the source; the removal takes the listed files one by one and a folder only
//! when it is empty.

use super::archive::{self, Kind, Row};
use super::{plan, Options};
use crate::home_state;
use cap_std::fs::Dir;
use serde_json::{json, Value};
use std::collections::BTreeMap;
use std::io::{self, Write};
use std::path::Path;

/// The receipt of one run, in its archive folder.
const RECEIPT: &str = "receipt.json";
const RECEIPT_NEXT: &str = "receipt.json.next";

/// What the caller picked and approved.
pub struct Request<'a> {
    /// Plan item paths, relative to the moai folder.
    pub picks: &'a [String],
    /// The `digest` of the plan the picks were made from.
    pub confirm: &'a str,
    /// The caller's reference for the user's approval.
    pub permission: &'a str,
}

/// A point of a run a test may act at (SPEC D-4); the operation does nothing there.
pub enum Step<'a> {
    /// The archive and the list of `item` are written in `dir`; nothing is read back yet.
    Written { item: &'a str, dir: &'a Path },
    /// The checks of `item` passed and its first removal comes next; `run` holds the receipt.
    Removing { item: &'a str, run: &'a Path },
}

/// One picked item and what became of it.
struct Done {
    path: String,
    kind: String,
    action: String,
    state: &'static str,
    reason: Option<String>,
    archive_item: Option<String>,
    left: Vec<String>,
}

impl Done {
    fn json(&self) -> Value {
        let mut v = json!({
            "path": self.path,
            "kind": self.kind,
            "action": self.action,
            "state": self.state,
        });
        if let Some(reason) = &self.reason {
            v["reason"] = json!(reason);
        }
        if let Some(folder) = &self.archive_item {
            v["archive_item"] = json!(folder);
        }
        if !self.left.is_empty() {
            v["left"] = json!(self.left);
        }
        v
    }

    fn whole(&self) -> bool {
        matches!(self.state, "archived" | "deleted") && self.left.is_empty()
    }

    fn end(&mut self, state: &'static str, reason: impl Into<String>) {
        self.state = state;
        self.reason = Some(reason.into());
    }
}

/// The run folder, its receipt and where both are.
struct Run {
    dir: Dir,
    path: std::path::PathBuf,
    name: String,
    head: Value,
}

impl Run {
    /// Replaces the receipt in one step.
    fn receipt(&self, status: &str, items: &[Done]) -> io::Result<()> {
        let mut body = self.head.clone();
        body["status"] = json!(status);
        body["items"] = json!(items.iter().map(Done::json).collect::<Vec<_>>());
        let text = serde_json::to_vec_pretty(&body).map_err(io::Error::other)?;
        // A file left by a failed earlier write of this run.
        let _ = self.dir.remove_file(RECEIPT_NEXT);
        let mut file = archive::create_private(&self.dir, RECEIPT_NEXT)?;
        file.write_all(&text)?;
        file.sync_all()?;
        self.dir.rename(RECEIPT_NEXT, &self.dir, RECEIPT)?;
        archive::own_handle(&self.dir)?.sync_all()
    }
}

/// Applies the picked items of the plan of `<home>/.moai`.
pub fn apply(home: &Path, opts: &Options, request: &Request) -> Result<Value, String> {
    apply_with(home, opts, request, &mut |_| {})
}

/// [`apply`] with a hook at each [`Step`].
#[doc(hidden)]
pub fn apply_with(
    home: &Path,
    opts: &Options,
    request: &Request,
    at: &mut dyn FnMut(Step<'_>),
) -> Result<Value, String> {
    // Every refusal comes before the first write.
    if !crate::contract::permission_reference_ok(request.permission) {
        return Err("tidy-apply needs a usable permission reference for the user's pick".into());
    }
    if request.picks.is_empty() {
        return Err("no item is picked; name each with --pick <plan item path>".into());
    }
    let scan = super::scan(home, opts)?;
    let (moai_path, moai) = super::open_moai(home)?;
    let archive_path = moai_path.join("archive");
    match moai.symlink_metadata("archive") {
        Err(e) if e.kind() == io::ErrorKind::NotFound => {}
        Ok(md) if md.is_dir() => {}
        _ => {
            return Err(format!(
                "`{}` is a link, not a folder or not readable; nothing was written",
                archive_path.display()
            ));
        }
    }
    let (planned, digest) = plan::build(&scan, &opts.thresholds);
    if request.confirm != digest {
        return Err(format!(
            "--confirm differs from the digest of the plan now ({digest}); read tidy-plan again \
             and pick from it"
        ));
    }
    let mut items = picked(&planned, request.picks)?;

    let free_before = free_bytes(home);
    let run = start(&moai, &archive_path, home, &digest, request.permission)?;
    if let Err(e) = run.receipt("pending", &items) {
        return Err(format!(
            "the receipt in `{}` could not be written ({}); nothing was removed",
            run.path.display(),
            e.kind()
        ));
    }
    let mut receipt_failed = false;
    for i in 0..items.len() {
        if receipt_failed {
            items[i].end("skipped", "receipt_not_written");
            continue;
        }
        one(&moai, &moai_path, &run, opts, &mut items, i, at);
        receipt_failed = run.receipt("pending", &items).is_err();
    }
    let state = if items.iter().all(Done::whole) {
        "committed"
    } else {
        "partial"
    };
    let receipt_failed = run.receipt(state, &items).is_err() || receipt_failed;
    let mut out = json!({
        "home": home.display().to_string(),
        "archive": run.name,
        "archive_root": run.path.display().to_string(),
        "digest": digest,
        "state": if receipt_failed { "partial" } else { state },
        "items": items.iter().map(Done::json).collect::<Vec<_>>(),
        "free_bytes": { "before": free_before, "after": free_bytes(home) },
        "writes": true,
    });
    if receipt_failed {
        out["receipt"] = json!("not_written");
    }
    Ok(out)
}

/// The plan items the picks name, in plan order. A pick that names none, or a kept one, is
/// refused.
fn picked(planned: &[Value], picks: &[String]) -> Result<Vec<Done>, String> {
    let text = |v: &Value, k: &str| v[k].as_str().unwrap_or("").to_string();
    for pick in picks {
        let Some(item) = planned.iter().find(|i| i["path"] == pick.as_str()) else {
            return Err(format!("`{pick}` is no item of the plan"));
        };
        if !matches!(item["action"].as_str(), Some("archive" | "delete")) {
            return Err(format!(
                "`{pick}` is kept by the plan: {}",
                text(item, "reason")
            ));
        }
    }
    Ok(planned
        .iter()
        .filter(|i| picks.iter().any(|p| i["path"] == p.as_str()))
        .map(|i| Done {
            path: text(i, "path"),
            kind: text(i, "kind"),
            action: text(i, "action"),
            state: "pending",
            reason: None,
            archive_item: None,
            left: Vec::new(),
        })
        .collect())
}

/// Creates the archive folder of this run: `archive/<UTC>`, with `.2`, `.3` … in one second.
fn start(
    moai: &Dir,
    archive_path: &Path,
    home: &Path,
    digest: &str,
    permission: &str,
) -> Result<Run, String> {
    let fail = |e: io::Error| {
        format!(
            "no archive folder could be made under `{}`: {}",
            archive_path.display(),
            e.kind()
        )
    };
    match archive::private_dir(moai, "archive") {
        Ok(_) => {}
        Err(e) if e.kind() == io::ErrorKind::AlreadyExists => {}
        Err(e) => return Err(fail(e)),
    }
    let all = archive::sub_dir(moai, "archive", "archive")?;
    let stamp = home_state::stamp(&crate::audit::utc_now())?;
    for n in 1..1000u32 {
        let name = if n == 1 {
            stamp.clone()
        } else {
            format!("{stamp}.{n}")
        };
        match archive::private_dir(&all, &name) {
            Ok(dir) => {
                return Ok(Run {
                    dir,
                    path: archive_path.join(&name),
                    head: json!({
                        "schema_version": 1,
                        "operation": "tidy.apply",
                        "created": name,
                        "home": home.display().to_string(),
                        "digest": digest,
                        "permission_reference": permission,
                    }),
                    name,
                });
            }
            Err(e) if e.kind() == io::ErrorKind::AlreadyExists => {}
            Err(e) => return Err(fail(e)),
        }
    }
    Err(fail(io::Error::from(io::ErrorKind::AlreadyExists)))
}

/// Whether a Go session may be live: its lock is held, a registered process is alive, or
/// either cannot be told.
fn in_use(moai: &Path) -> Option<&'static str> {
    let state = moai.join("state");
    match std::fs::symlink_metadata(&state) {
        Err(e) if e.kind() == io::ErrorKind::NotFound => return None,
        // A link is not followed, and a folder that cannot be listed tells nothing.
        Ok(md) if md.is_dir() && std::fs::read_dir(&state).is_ok() => {}
        _ => return Some("in_use_unknown"),
    }
    let lock = home_state::session_lock_held(moai);
    let live = home_state::live_sessions(moai);
    if lock == Some(true) || live.is_some_and(|n| n > 0) {
        Some("in_use")
    } else if lock.is_none() || live.is_none() {
        Some("in_use_unknown")
    } else {
        None
    }
}

/// Runs one picked item and records in it what happened.
fn one(
    moai: &Dir,
    moai_path: &Path,
    run: &Run,
    opts: &Options,
    items: &mut [Done],
    i: usize,
    at: &mut dyn FnMut(Step<'_>),
) {
    let path = items[i].path.clone();
    let retired = items[i].kind == "retired_tool";
    let busy = || if retired { in_use(moai_path) } else { None };
    if let Some(reason) = busy() {
        return items[i].end("skipped", reason);
    }
    if items[i].action == "delete" {
        at(Step::Removing {
            item: &path,
            run: &run.path,
        });
        if let Some(reason) = busy() {
            return items[i].end("skipped", reason);
        }
        return match delete_empty(moai, &path) {
            Ok(()) => items[i].state = "deleted",
            Err(reason) => items[i].end("kept", reason),
        };
    }
    let rows = match archive::list(moai, &path, opts.limits) {
        Ok(rows) => rows,
        Err(reason) => return items[i].end("skipped", reason),
    };
    let folder = format!("item-{:03}", i + 1);
    let dir = match archive::private_dir(&run.dir, &folder) {
        Ok(dir) => dir,
        Err(e) => return items[i].end("failed", format!("archive_not_written: {}", e.kind())),
    };
    let checked = archive::write(moai, &dir, &path, &rows).and_then(|()| {
        at(Step::Written {
            item: &path,
            dir: &run.path.join(&folder),
        });
        verify(moai, &dir, &path, &rows, opts)
    });
    if let Err(reason) = checked {
        discard(&run.dir, &dir, &folder);
        return items[i].end("failed", reason);
    }
    items[i].archive_item = Some(folder.clone());
    // The receipt names the archive before the first file of the original goes.
    if run.receipt("pending", items).is_err() {
        return items[i].end("failed", "receipt_not_written");
    }
    at(Step::Removing {
        item: &path,
        run: &run.path,
    });
    // Asked again next to the removal: archiving a large item takes a while.
    if let Some(reason) = busy() {
        discard(&run.dir, &dir, &folder);
        items[i].archive_item = None;
        return items[i].end("skipped", reason);
    }
    let trouble = remove(moai, &rows);
    items[i].left = remaining(moai, &path);
    match trouble {
        None => items[i].state = "archived",
        Some(reason) => items[i].end("partly_removed", reason),
    }
}

/// The archive and its list read back, and the source still as listed. A file that appeared
/// since is not a change of a listed one: it is not in the archive and it will stay.
fn verify(moai: &Dir, dir: &Dir, item: &str, rows: &[Row], opts: &Options) -> Result<(), String> {
    let list = archive::read_list(dir)?;
    if list.item != item || list.entries != rows {
        return Err("the list read back differs from the files listed".into());
    }
    archive::check(dir, rows)?;
    let now: BTreeMap<String, Row> = archive::list(moai, item, opts.limits)
        .map_err(|reason| format!("source_changed: {reason}"))?
        .into_iter()
        .map(|r| (r.path.clone(), r))
        .collect();
    for row in rows {
        if !now.get(&row.path).is_some_and(|n| row.same_as(n)) {
            return Err(format!("source_changed: {}", row.path));
        }
    }
    Ok(())
}

/// Removes the unfinished files this run created for one item, and their folder.
fn discard(run: &Dir, dir: &Dir, folder: &str) {
    for name in [archive::DATA, archive::LIST] {
        let _ = dir.remove_file(name);
    }
    let _ = run.remove_dir(folder);
}

/// Removes the listed files, then the listed folders that are empty, children first. Says what
/// first stopped a removal; a folder that still holds something is not a trouble.
fn remove(moai: &Dir, rows: &[Row]) -> Option<String> {
    let mut trouble = None;
    for row in rows.iter().filter(|r| r.kind == Kind::File) {
        if let Err(reason) = remove_file(moai, row) {
            trouble.get_or_insert(reason);
        }
    }
    for row in rows.iter().rev().filter(|r| r.kind == Kind::Dir) {
        if let Err(reason) = remove_folder(moai, &row.path) {
            trouble.get_or_insert(reason);
        }
    }
    trouble
}

/// Removes one listed file, if it is still the file that was archived.
fn remove_file(moai: &Dir, row: &Row) -> Result<(), String> {
    let names = archive::parts(&row.path)?;
    let dir = archive::parent(moai, &names)?;
    let name = names[names.len() - 1];
    let unreadable = |e: io::Error| format!("unreadable: {}: {}", row.path, e.kind());
    let (file, md) = archive::open_file(&dir, name).map_err(unreadable)?;
    let (sum, bytes) = archive::hash(file).map_err(unreadable)?;
    use std::os::unix::fs::MetadataExt;
    let same = bytes == row.bytes
        && Some(sum) == row.sha256
        && md.mode() & 0o7777 == row.mode
        && md.mtime() == row.mtime_secs
        && md.mtime_nsec() == i64::from(row.mtime_nanos);
    if !same {
        return Err(format!("source_changed: {}", row.path));
    }
    dir.remove_file(name)
        .map_err(|e| format!("not_removed: {}: {}", row.path, e.kind()))
}

/// Removes the folder at `path` if it is empty; one that is not stays.
fn remove_folder(moai: &Dir, path: &str) -> Result<(), String> {
    let names = archive::parts(path)?;
    let dir = archive::parent(moai, &names)?;
    let name = names[names.len() - 1];
    archive::sub_dir(&dir, name, path)?;
    match dir.remove_dir(name) {
        Err(e) if e.kind() != io::ErrorKind::DirectoryNotEmpty => {
            Err(format!("not_removed: {path}: {}", e.kind()))
        }
        _ => Ok(()),
    }
}

/// The `delete` action: only an empty folder goes, without any recursive removal.
fn delete_empty(moai: &Dir, path: &str) -> Result<(), String> {
    let names = archive::parts(path)?;
    let dir = archive::parent(moai, &names)?;
    let name = names[names.len() - 1];
    archive::sub_dir(&dir, name, path)?;
    dir.remove_dir(name).map_err(|e| match e.kind() {
        io::ErrorKind::DirectoryNotEmpty => format!("not_empty: {path}"),
        kind => format!("not_removed: {path}: {kind}"),
    })
}

/// What is still at and below `path`, by names only; a link is named and not followed.
fn remaining(moai: &Dir, path: &str) -> Vec<String> {
    fn walk(dir: &Dir, name: &str, path: &str, out: &mut Vec<String>) {
        let Ok(md) = dir.symlink_metadata(name) else {
            return;
        };
        out.push(path.to_string());
        if !md.is_dir() {
            return;
        }
        let Ok(below) = archive::sub_dir(dir, name, path) else {
            return;
        };
        for child in archive::names_of(&below).unwrap_or_default() {
            walk(&below, &child, &format!("{path}/{child}"), out);
        }
    }
    let mut out = Vec::new();
    let names: Vec<&str> = path.split('/').collect();
    match archive::parent(moai, &names) {
        Ok(dir) => walk(&dir, names[names.len() - 1], path, &mut out),
        Err(_) => out.push(path.to_string()),
    }
    out
}

/// Free bytes of the file system of `path`, or `not_checked`.
fn free_bytes(path: &Path) -> Value {
    use std::os::unix::ffi::OsStrExt;
    let Ok(c) = std::ffi::CString::new(path.as_os_str().as_bytes()) else {
        return json!("not_checked");
    };
    // SAFETY: statvfs fills the zeroed struct for a NUL-terminated path.
    let mut s: libc::statvfs = unsafe { std::mem::zeroed() };
    if unsafe { libc::statvfs(c.as_ptr(), &mut s) } != 0 {
        return json!("not_checked");
    }
    #[allow(clippy::unnecessary_cast)]
    let free = s.f_bavail as u64 * s.f_frsize as u64;
    json!(free)
}
