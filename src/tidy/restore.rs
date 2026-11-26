//! `tidy-restore` (SPEC-20261126-home-tidy-apply, REQ-9, REQ-10): put the files of an archive
//! back where nothing exists. An item is checked whole against its list before the first of its
//! files is written; nothing is overwritten and the archive is kept.

use super::archive::{self, Kind, List, Row};
use cap_std::fs::Dir;
use serde_json::{json, Value};
use std::collections::BTreeMap;
use std::fs::{FileTimes, Permissions};
use std::io::{self, Read};
use std::os::unix::fs::PermissionsExt;
use std::path::Path;

/// One item folder of the archive: the item it names and its list, or why it cannot be used.
struct Held {
    item: String,
    dir: Result<Dir, String>,
}

/// Restores the picked items (all of them without a pick) of the archive `stamp`.
pub fn restore(
    home: &Path,
    stamp: &str,
    picks: &[String],
    permission: &str,
) -> Result<Value, String> {
    if !crate::contract::permission_reference_ok(permission) {
        return Err("tidy-restore needs a usable permission reference".into());
    }
    if !is_stamp(stamp) {
        return Err(format!(
            "`{stamp}` is not the name of an archive (YYYYMMDDTHHMMSSZ, with .N)"
        ));
    }
    let (_, moai) = super::open_moai(home)?;
    let run = archive::sub_dir(&moai, "archive", "archive")
        .and_then(|all| archive::sub_dir(&all, stamp, stamp))
        .map_err(|reason| format!("no archive `{stamp}` to read: {reason}"))?;
    let held = held(&run)?;
    for pick in picks {
        if !held.iter().any(|h| &h.item == pick) {
            return Err(format!("`{pick}` is no item of the archive `{stamp}`"));
        }
    }
    let mut items = Vec::new();
    let mut whole = true;
    for one in held
        .iter()
        .filter(|h| picks.is_empty() || picks.contains(&h.item))
    {
        let mut existing = Vec::new();
        let done = one
            .dir
            .as_ref()
            .map_err(Clone::clone)
            .and_then(|dir| item(&moai, dir, &one.item, &mut existing));
        let mut v = json!({ "path": one.item, "state": "restored" });
        if let Err(reason) = done {
            v["state"] = json!("failed");
            v["reason"] = json!(reason);
            whole = false;
        }
        if !existing.is_empty() {
            existing.sort();
            v["existing"] = json!(existing);
            whole = false;
        }
        items.push(v);
    }
    Ok(json!({
        "home": home.display().to_string(),
        "archive": stamp,
        "state": if whole { "restored" } else { "partial" },
        "items": items,
        "writes": true,
    }))
}

/// `YYYYMMDDTHHMMSSZ`, optionally followed by `.N`.
fn is_stamp(s: &str) -> bool {
    let b = s.as_bytes();
    let digits = |r: &[u8]| !r.is_empty() && r.iter().all(u8::is_ascii_digit);
    b.len() >= 16
        && digits(&b[..8])
        && b[8] == b'T'
        && digits(&b[9..15])
        && b[15] == b'Z'
        && (b.len() == 16 || (b[16] == b'.' && digits(&b[17..])))
}

/// The item folders of a run. The receipt gives the name of an item whose list cannot.
fn held(run: &Dir) -> Result<Vec<Held>, String> {
    let named = receipt_names(run);
    let mut out = Vec::new();
    for folder in archive::names_of(run).map_err(|e| format!("the archive is not readable: {e}"))? {
        if !folder.starts_with("item-") {
            continue;
        }
        let read = archive::sub_dir(run, &folder, &folder)
            .and_then(|dir| archive::read_list(&dir).map(|list| (dir, list.item)));
        let known = named.get(&folder).cloned();
        out.push(match read {
            Ok((dir, item)) => Held {
                dir: match &known {
                    Some(k) if *k != item => {
                        Err("the list names another item than the receipt".into())
                    }
                    _ => Ok(dir),
                },
                item: known.unwrap_or(item),
            },
            Err(reason) => Held {
                item: known.unwrap_or(folder),
                dir: Err(reason),
            },
        });
    }
    Ok(out)
}

/// `item folder → item path` of the receipt, as far as it can be read.
fn receipt_names(run: &Dir) -> BTreeMap<String, String> {
    let mut text = Vec::new();
    let read = archive::open_file(run, "receipt.json")
        .and_then(|(f, _)| f.take(64 * 1024 * 1024).read_to_end(&mut text));
    let body: Value = match read {
        Ok(_) => serde_json::from_slice(&text).unwrap_or(Value::Null),
        Err(_) => Value::Null,
    };
    body["items"]
        .as_array()
        .into_iter()
        .flatten()
        .filter_map(|i| {
            Some((
                i["archive_item"].as_str()?.to_string(),
                i["path"].as_str()?.to_string(),
            ))
        })
        .collect()
}

/// Checks one archived item whole, then creates what is missing of it.
fn item(moai: &Dir, dir: &Dir, name: &str, existing: &mut Vec<String>) -> Result<(), String> {
    let list: List = archive::read_list(dir)?;
    if list.item != name {
        return Err("the list changed while it was read".into());
    }
    archive::check_rows(&list)?;
    archive::check(dir, &list.entries)?;
    let mut rows: Vec<&Row> = list.entries.iter().collect();
    rows.sort_by(|a, b| a.path.cmp(&b.path));
    make_parents(moai, &list.item)?;

    // Folders first, parents before children; one that cannot be entered blocks what is below.
    let mut made: Vec<&Row> = Vec::new();
    let mut blocked: Vec<String> = Vec::new();
    let under = |blocked: &[String], path: &str| blocked.iter().any(|b| path.starts_with(b));
    for row in rows.iter().filter(|r| r.kind == Kind::Dir) {
        if under(&blocked, &row.path) {
            continue;
        }
        let names = archive::parts(&row.path)?;
        let parent = archive::parent(moai, &names)?;
        let last = names[names.len() - 1];
        match parent.symlink_metadata(last) {
            Ok(md) => {
                existing.push(row.path.clone());
                if !md.is_dir() {
                    blocked.push(format!("{}/", row.path));
                }
            }
            Err(e) if e.kind() == io::ErrorKind::NotFound => {
                archive::private_dir(&parent, last).map_err(|e| not_written(&row.path, &e))?;
                made.push(row);
            }
            Err(e) => return Err(not_written(&row.path, &e)),
        }
    }
    let written = archive::each_entry(dir, &list.entries, &mut |row, content| {
        if under(&blocked, &row.path) {
            return archive::hash(content).map_err(|e| not_written(&row.path, &e));
        }
        file(moai, row, content, existing)
    });
    // Modes and times last, children first (writing a child changes its folder's time), also
    // after a failed write: the folders that were created stay.
    for row in made.iter().rev() {
        let names = archive::parts(&row.path)?;
        let parent = archive::parent(moai, &names)?;
        let own = archive::sub_dir(&parent, names[names.len() - 1], &row.path)?;
        archive::own_handle(&own)
            .and_then(|f| {
                f.set_times(FileTimes::new().set_modified(row.modified()))?;
                f.set_permissions(Permissions::from_mode(row.mode))
            })
            .map_err(|e| not_written(&row.path, &e))?;
    }
    written
}

fn not_written(path: &str, e: &io::Error) -> String {
    format!("not_written: {path}: {}", e.kind())
}

/// Opens the folders above the item without following a link, creating the missing ones.
fn make_parents(moai: &Dir, item: &str) -> Result<(), String> {
    let names = archive::parts(item)?;
    let mut dir = moai.try_clone().map_err(|e| not_written(item, &e))?;
    for (i, name) in names[..names.len() - 1].iter().enumerate() {
        let shown = names[..=i].join("/");
        match dir.symlink_metadata(name) {
            Err(e) if e.kind() == io::ErrorKind::NotFound => match dir.create_dir(name) {
                Ok(()) => {}
                Err(e) => return Err(not_written(&shown, &e)),
            },
            _ => {}
        }
        dir = archive::sub_dir(&dir, name, &shown)?;
    }
    Ok(())
}

/// Creates one file where nothing exists and says what was read from the archive for it.
fn file(
    moai: &Dir,
    row: &Row,
    content: &mut dyn Read,
    existing: &mut Vec<String>,
) -> Result<(String, u64), String> {
    let names = archive::parts(&row.path)?;
    let parent = archive::parent(moai, &names)?;
    let name = names[names.len() - 1];
    let failed = |e: io::Error| not_written(&row.path, &e);
    let mut out = match archive::create_private(&parent, name) {
        Ok(out) => out,
        Err(e) if e.kind() == io::ErrorKind::AlreadyExists => {
            existing.push(row.path.clone());
            return archive::hash(content).map_err(failed);
        }
        Err(e) => return Err(failed(e)),
    };
    let mut content = content;
    let written = archive::copy_hashed(&mut content, &mut out).and_then(|got| {
        out.set_times(FileTimes::new().set_modified(row.modified()))?;
        out.set_permissions(Permissions::from_mode(row.mode))?;
        out.sync_all()?;
        Ok(got)
    });
    match written {
        Ok((sum, bytes)) if bytes == row.bytes && Some(&sum) == row.sha256.as_ref() => {
            Ok((sum, bytes))
        }
        other => {
            // Only the file this call just created.
            let _ = parent.remove_file(name);
            match other {
                Ok(_) => Err(format!("`{}` differs from the list", row.path)),
                Err(e) => Err(failed(e)),
            }
        }
    }
}
