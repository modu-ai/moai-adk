//! #2871 U2 (PLAN comment 6878852377, user decision D-M1 6876633358): retire the Go
//! generation's items in `~/.moai`. Each top-level entry the HOME name table
//! ([`crate::home_layout::HOME_TABLE`]) names as a Go item moves, by one rename, into a dated
//! folder `~/.moai-go-retired/<UTC>/` outside the active `~/.moai` (same file system), with a
//! manifest beside it. User data, items of unknown owner, names not in the table, this
//! generation's own names, a key file that is still a merge candidate and anything a running
//! program uses stay where they are. Nothing is deleted or copied; a secret is renamed by name
//! and never opened or walked.
//!
//! The moves are the per-item journal of [`crate::move_journal`]: written before the first
//! rename, resumable, and reversible only for the entry that was moved.
//!
//! Only a HOME the verdict table calls `go`, `mixed` or `rust-first` has anything to retire. A
//! name both generations write (`state`, `config`, `backups`) retires only on a pure `go` HOME:
//! elsewhere the folder may hold this generation's files too.

use crate::home_layout::{self, Class, Entry, Rule, RustUse};
use crate::home_state::{live_sessions, session_lock_held, stamp, InUse, CANDIDATE_LIMIT};
use crate::move_journal::{self, MoveEnv, PlannedMove};
use crate::state_root::{classify_home, MOAI_DIR};
use serde_json::{json, Value};
use std::collections::BTreeMap;
use std::fs;
use std::path::{Path, PathBuf};

/// The folder beside `~/.moai` that holds the dated backups.
pub const RETIRED_DIR: &str = ".moai-go-retired";
/// The journal `kind` of a retire.
pub const KIND: &str = "go-retire";
const MANIFEST_PREFIX: &str = "retire-manifest-";

/// The kind of a top-level entry, from its own metadata (a link is never followed).
fn kind_of(p: &Path) -> &'static str {
    match fs::symlink_metadata(p) {
        Ok(m) if m.file_type().is_symlink() => "link",
        Ok(m) if m.is_dir() => "folder",
        Ok(m) if m.is_file() => "file",
        _ => "special",
    }
}

fn class_str(c: Class) -> &'static str {
    match c {
        Class::Rust => "rust",
        Class::ToolAsset => "tool-asset",
        Class::ConfigGuide => "config-guide",
        Class::Secret => "secret",
        Class::UserData => "user-data",
        Class::Unknown => "unknown",
    }
}

fn rule_str(r: Rule) -> &'static str {
    match r {
        Rule::RustName => "rust-name",
        Rule::GoRetires => "go-retires",
        Rule::UserDataKeep => "user-data-keep",
        Rule::SecretRenameOnly => "secret-rename-only",
        Rule::UnknownKeep => "unknown-keep",
    }
}

/// What happens to the top-level entry of `kind` that the table row `entry` names, on a HOME
/// with verdict `generation`: `("retire" | "keep", reason)`. Pure; whether a program uses the
/// entry is judged afterwards.
pub fn decide(entry: Option<&Entry>, kind: &str, generation: &str) -> (&'static str, &'static str) {
    let Some(e) = entry else {
        return ("keep", "not_in_table");
    };
    match kind {
        "file" | "folder" => {}
        "link" => return ("keep", "link"),
        _ => return ("keep", "special_file"),
    }
    if !e.go {
        return ("keep", "this_generation");
    }
    match e.rule {
        Rule::UserDataKeep => ("keep", "user_data"),
        Rule::UnknownKeep => ("keep", "owner_unknown"),
        Rule::RustName => ("keep", "this_generation"),
        Rule::GoRetires | Rule::SecretRenameOnly => {
            // The key layer reads `.env.glm` where it is; the way back to it from a backup is a
            // later step, so it stays until then.
            if e.merge && e.class == Class::Secret {
                ("keep", "merge_candidate_kept")
            } else if e.rust != RustUse::None && generation != "go" {
                ("keep", "shared_name")
            } else {
                ("retire", "go_item")
            }
        }
    }
}

fn canon(p: &Path) -> PathBuf {
    fs::canonicalize(p).unwrap_or_else(|_| p.to_path_buf())
}

/// The preview for the HOME at `home`. `moai_home` is the `MOAI_HOME` value, `now` the clock
/// as `YYYY-MM-DDTHH:MM:SSZ` (names the dated folder). Writes nothing.
pub fn plan(home: &Path, moai_home: Option<&str>, now: &str) -> Result<Value, String> {
    if !home.is_absolute() {
        return Err("HOME is not an absolute path".into());
    }
    let stamp = stamp(now)?;
    let moai = home.join(MOAI_DIR);
    let c = classify_home(home);
    let root = home.join(RETIRED_DIR);
    let folder = root.join(&stamp);
    let active = matches!(c.generation, "go" | "mixed" | "rust-first");

    let mut blocks: Vec<String> = Vec::new();
    let mut names: Vec<String> = Vec::new();
    let mut unreadable = 0usize;
    if active {
        if let Ok(rd) = fs::read_dir(&moai) {
            for e in rd.flatten() {
                match e.file_name().into_string() {
                    Ok(n) => names.push(n),
                    Err(_) => unreadable += 1,
                }
            }
        }
        names.sort();
        match fs::symlink_metadata(&root) {
            Ok(m) if m.is_dir() => {}
            Ok(_) => blocks.push("retired_root_not_a_folder".into()),
            Err(_) => {}
        }
        let taken = [
            folder.clone(),
            move_journal::journal_path(&root, &stamp),
            root.join(format!("{MANIFEST_PREFIX}{stamp}.json")),
        ];
        if taken.iter().any(|p| fs::symlink_metadata(p).is_ok()) {
            blocks.push("stamp_taken".into());
        }
        let lock = session_lock_held(&moai);
        let registered = live_sessions(&moai);
        if lock == Some(true) || registered.is_some_and(|n| n > 0) {
            blocks.push("go_session_live".into());
        } else if lock.is_none() {
            blocks.push("session_lock_unknown".into());
        } else if registered.is_none() {
            blocks.push("session_registry_unknown".into());
        }
        if moai_home.is_some_and(|m| !m.is_empty() && Path::new(m) != moai) {
            blocks.push("moai_home_elsewhere".into());
        }
        if names.len() + unreadable > CANDIDATE_LIMIT {
            blocks.push("not_fully_listed".into());
        }
        if unreadable > 0 {
            blocks.push(format!("unreadable_names:{unreadable}"));
        }
    } else if matches!(c.generation, "interrupted" | "refuse") {
        blocks.push(format!("generation_{}", c.generation));
    }
    let total = names.len() + unreadable;

    let in_use = (active && blocks.is_empty()).then(InUse::read);
    let canon_moai = canon(&moai);
    let limit = crate::go_state::unrecorded_limit();
    let entries: Vec<Value> = names
        .iter()
        .take(CANDIDATE_LIMIT)
        .map(|n| {
            let raw = moai.join(n);
            let kind = kind_of(&raw);
            let row = home_layout::lookup(n);
            let (mut action, mut reason) = decide(row, kind, c.generation);
            let secret = row.is_some_and(|e| e.class == Class::Secret);
            if action == "retire" {
                if let Some(in_use) = &in_use {
                    match in_use.uses(&raw, &canon_moai.join(n)) {
                        Ok(None) => {}
                        Ok(Some(why)) | Err(why) => {
                            action = "skip-in-use";
                            reason = why;
                        }
                    }
                }
                // A nested checkout would lose its registration if renamed. A secret is judged
                // by its name and kind alone: it is not walked.
                if action == "retire" && !secret && kind == "folder" {
                    match crate::go_state::nested_checkouts_under(
                        home,
                        &format!("{MOAI_DIR}/{n}"),
                        limit,
                    ) {
                        Some(0) => {}
                        Some(_) => {
                            action = "keep";
                            reason = "nested_checkout";
                        }
                        None => {
                            action = "keep";
                            reason = "not_fully_checked";
                        }
                    }
                }
            }
            let destination = if action == "retire" {
                Value::from(folder.join(n).display().to_string())
            } else {
                Value::Null
            };
            json!({
                "name": n,
                "kind": kind,
                "class": row.map_or("unknown", |e| class_str(e.class)),
                "rule": row.map_or("none", |e| rule_str(e.rule)),
                "merge_candidate": row.is_some_and(|e| e.merge),
                "action": action,
                "reason": reason,
                "destination": destination,
            })
        })
        .collect();

    let count = |a: &str| entries.iter().filter(|e| e["action"] == a).count();
    let counts = json!({
        "retire": count("retire"),
        "keep": count("keep"),
        "skip-in-use": count("skip-in-use"),
    });
    let action = if !blocks.is_empty() {
        "blocked"
    } else if count("retire") > 0 {
        "retire"
    } else {
        "none"
    };

    let mut warnings = Vec::new();
    if total > CANDIDATE_LIMIT {
        warnings.push(format!(
            "{total} entries in {}; the first {CANDIDATE_LIMIT} are listed",
            moai.display()
        ));
    }
    if entries
        .iter()
        .any(|e| e["reason"] == "merge_candidate_kept")
    {
        warnings.push("a key file stays in place until the merge step has offered it".to_string());
    }

    let folder_shown = active.then(|| folder.display().to_string());
    let bound = json!({
        "mode": "retire",
        "generation": c.generation,
        "action": action,
        "folder": folder_shown,
        "blocks": blocks,
        "entries": entries.iter().map(|e| json!({
            "name": e["name"], "kind": e["kind"], "action": e["action"], "reason": e["reason"],
        })).collect::<Vec<_>>(),
    });
    let plan_digest = crate::digest(serde_json::to_string(&bound).unwrap_or_default().as_bytes());
    // What the confirmation binds, without the clock: the HOME it was made for, the generation,
    // the action, the blocks and the retirement set (the entries that retire). The dated folder
    // is out, so `moai update` can hand the digest out as a token for a later run. An entry
    // that stays or is in use is out too: the layer's own writes (a store like `.moai/agents`)
    // add some, and they confirm nothing. An entry that starts or stops being retired moves
    // the digest, so a program that frees (or takes) an item between preview and confirmation
    // makes the token stale.
    let stable = json!({
        "mode": "retire",
        "home": home.display().to_string(),
        "generation": c.generation,
        "action": action,
        "blocks": blocks,
        "entries": entries
            .iter()
            .filter(|e| e["action"] == "retire")
            .map(|e| json!({
                "name": e["name"], "kind": e["kind"], "action": e["action"], "reason": e["reason"],
            }))
            .collect::<Vec<_>>(),
    });
    let content_digest = crate::digest(
        serde_json::to_string(&stable)
            .unwrap_or_default()
            .as_bytes(),
    );
    Ok(json!({
        "schema_version": 1,
        "mode": "retire",
        "home": home.display().to_string(),
        "state": c.to_json(),
        "action": action,
        "folder": folder_shown,
        "blocks": blocks,
        "entries": entries,
        "counts": counts,
        "total": total,
        "warnings": warnings,
        "plan_digest": plan_digest,
        "content_digest": content_digest,
        "apply": "home-retire-apply with the same --now, this plan_digest and a permission reference; on a real HOME it runs only after the user confirms",
        "writes": false,
    }))
}

fn make_dir(p: &Path) -> Result<(), String> {
    let mut b = fs::DirBuilder::new();
    #[cfg(unix)]
    std::os::unix::fs::DirBuilderExt::mode(&mut b, 0o700);
    b.create(p).map_err(|e| format!("{}: {e}", p.display()))
}

/// Retire the entries of an identical fresh preview whose action is `retire` (see the module
/// comment). Writes the dated folder, the manifest and the journal, then renames each entry.
pub fn apply(
    home: &Path,
    moai_home: Option<&str>,
    now: &str,
    plan_digest: &str,
    permission_reference: &str,
) -> Result<Value, String> {
    apply_checked(
        home,
        moai_home,
        now,
        "plan_digest",
        plan_digest,
        permission_reference,
    )
}

/// [`apply`] confirmed by the preview's `content_digest` (no dated folder in it) instead of its
/// `plan_digest`: the preview and the apply may run at different clocks. `moai update` uses it.
pub fn apply_confirmed(
    home: &Path,
    moai_home: Option<&str>,
    now: &str,
    content_digest: &str,
    permission_reference: &str,
) -> Result<Value, String> {
    apply_checked(
        home,
        moai_home,
        now,
        "content_digest",
        content_digest,
        permission_reference,
    )
}

fn apply_checked(
    home: &Path,
    moai_home: Option<&str>,
    now: &str,
    digest_key: &str,
    expected: &str,
    permission_reference: &str,
) -> Result<Value, String> {
    if !crate::contract::permission_reference_ok(permission_reference) {
        return Err("permission reference required".into());
    }
    let fresh = plan(home, moai_home, now)?;
    if fresh[digest_key].as_str() != Some(expected) {
        return Err("stale or altered plan: the recomputed digest differs; re-preview required, nothing was written".into());
    }
    let plan_digest = fresh["plan_digest"]
        .as_str()
        .unwrap_or_default()
        .to_string();
    let plan_digest = plan_digest.as_str();
    if fresh["action"] != "retire" {
        return Err(format!(
            "the preview's action is `{}`, not `retire`; nothing was written",
            fresh["action"].as_str().unwrap_or("?")
        ));
    }
    let stamp = stamp(now)?;
    let moai = home.join(MOAI_DIR);
    let root = home.join(RETIRED_DIR);
    let folder = root.join(&stamp);
    let rows: Vec<&Value> = fresh["entries"]
        .as_array()
        .map(|a| a.iter().filter(|e| e["action"] == "retire").collect())
        .unwrap_or_default();
    let items: Vec<PlannedMove> = rows
        .iter()
        .filter_map(|e| e["name"].as_str())
        .map(|n| PlannedMove {
            from: moai.join(n),
            to: folder.join(n),
        })
        .collect();

    match fs::symlink_metadata(&root) {
        Ok(m) if m.is_dir() => {}
        Ok(_) => {
            return Err(format!(
                "{} is not a folder; nothing was written",
                root.display()
            ))
        }
        Err(_) => make_dir(&root).map_err(|e| format!("{e}; nothing was written"))?,
    }
    make_dir(&folder).map_err(|e| format!("{e}; nothing was written"))?;

    let manifest_path = root.join(format!("{MANIFEST_PREFIX}{stamp}.json"));
    let kept: Vec<Value> = fresh["entries"]
        .as_array()
        .map(|a| {
            a.iter()
                .filter(|e| e["action"] != "retire")
                .cloned()
                .collect()
        })
        .unwrap_or_default();
    let manifest = json!({
        "schema_version": 1,
        "kind": KIND,
        "stamp": stamp,
        "home": home.display().to_string(),
        "created_at": now,
        "plan_digest": plan_digest,
        "permission_reference": permission_reference,
        "journal": move_journal::journal_path(&root, &stamp).display().to_string(),
        "items": rows.iter().map(|e| json!({
            "name": e["name"], "kind": e["kind"], "class": e["class"], "rule": e["rule"],
            "merge_candidate": e["merge_candidate"],
            "from": moai.join(e["name"].as_str().unwrap_or_default()).display().to_string(),
            "to": e["destination"],
        })).collect::<Vec<_>>(),
        "kept": kept,
    });
    write_new(&manifest_path, format!("{manifest:#}\n").as_bytes())?;

    // The programs that run now decide again: a name a program started to use after the
    // preview is skipped by the journal, not moved.
    let procs = InUse::read();
    let check = |p: &Path| match procs.uses(p, &canon(p)) {
        Ok(None) => move_journal::InUse::No,
        Ok(Some(why)) => move_journal::InUse::Yes(why.to_string()),
        Err(_) => move_journal::InUse::Unknown,
    };
    let env = MoveEnv::new(&check);
    let report =
        move_journal::apply(&root, KIND, &items, &BTreeMap::new(), now, &env).map_err(|e| {
            format!(
                "{e}; the journal in {} lets home-retire-revert put back what moved",
                root.display()
            )
        })?;
    let outcome_of = |o: &str| report.items.iter().filter(|i| i.outcome == o).count();
    let moved = outcome_of("moved");
    Ok(json!({
        "schema_version": 1,
        "state": "applied",
        "stamp": stamp,
        "home": home.display().to_string(),
        "folder": folder.display().to_string(),
        "manifest": manifest_path.display().to_string(),
        "journal": report.journal,
        "journal_state": report.state,
        "moved": moved,
        "not_moved": report.items.len() - moved,
        "items": report.items.iter().map(|i| json!({
            "name": Path::new(&i.from).file_name().map(|n| n.to_string_lossy().into_owned()),
            "outcome": i.outcome,
        })).collect::<Vec<_>>(),
        "revert": format!("home-retire-revert --home {} --stamp {stamp} --permission-reference <ref>", home.display()),
        "rerun": "home-retire-plan again: with no Go marker left the action is `none`",
    }))
}

/// Create `path` exclusively and write `bytes` into it, fsynced.
fn write_new(path: &Path, bytes: &[u8]) -> Result<(), String> {
    use std::io::Write;
    let mut f = fs::OpenOptions::new()
        .write(true)
        .create_new(true)
        .open(path)
        .map_err(|e| format!("{}: {e}", path.display()))?;
    f.write_all(bytes)
        .and_then(|()| f.sync_all())
        .map_err(|e| format!("{}: {e}", path.display()))
}

/// The newest retire journal in `root` (by its stamp), or the one of `stamp`.
fn find_journal(root: &Path, stamp: Option<&str>) -> Result<PathBuf, String> {
    if let Some(s) = stamp {
        if !crate::state_root::utc_stamp(s) {
            return Err(format!("stamp `{s}` is not YYYYMMDDTHHMMSSZ"));
        }
        return Ok(move_journal::journal_path(root, s));
    }
    let mut found: Vec<String> = Vec::new();
    for e in fs::read_dir(root)
        .map_err(|e| format!("{}: {e}", root.display()))?
        .flatten()
    {
        let name = e.file_name().to_string_lossy().into_owned();
        let Some(s) = name
            .strip_prefix("move-journal-")
            .and_then(|n| n.strip_suffix(".json"))
        else {
            continue;
        };
        if crate::state_root::utc_stamp(s) {
            found.push(s.to_string());
        }
    }
    found.sort();
    found
        .iter()
        .rev()
        .map(|s| move_journal::journal_path(root, s))
        .find(|p| move_journal::read_journal(p).is_ok_and(|j| j.kind == KIND))
        .ok_or_else(|| format!("no retire journal in {}; nothing was moved", root.display()))
}

/// Put back what a retire moved: the journal of `stamp` (the newest when `None`) is replayed
/// in reverse, each entry only while it is still the one that was moved and its old place is
/// free. Deletes nothing.
pub fn revert(
    home: &Path,
    stamp: Option<&str>,
    permission_reference: &str,
) -> Result<Value, String> {
    if !crate::contract::permission_reference_ok(permission_reference) {
        return Err("permission reference required".into());
    }
    if !home.is_absolute() {
        return Err("HOME is not an absolute path".into());
    }
    let root = home.join(RETIRED_DIR);
    match fs::symlink_metadata(&root) {
        Ok(m) if m.is_dir() => {}
        _ => {
            return Err(format!(
                "{} is not a folder; nothing was moved",
                root.display()
            ))
        }
    }
    let path = find_journal(&root, stamp)?;
    let journal = move_journal::read_journal(&path)?;
    if journal.kind != KIND {
        return Err(format!(
            "{} is not a retire journal; nothing was moved",
            path.display()
        ));
    }
    let report = move_journal::revert(&path)?;
    Ok(json!({
        "schema_version": 1,
        "state": report.state,
        "home": home.display().to_string(),
        "journal": report.journal,
        "changed": report.changed,
        "items": report.items.iter().map(|i| json!({
            "name": Path::new(&i.from).file_name().map(|n| n.to_string_lossy().into_owned()),
            "outcome": i.outcome,
        })).collect::<Vec<_>>(),
    }))
}

#[cfg(test)]
mod tests {
    use super::*;
    use home_layout::lookup;

    fn d(name: &str, kind: &str, generation: &str) -> (&'static str, &'static str) {
        decide(lookup(name), kind, generation)
    }

    #[test]
    fn the_decision_follows_the_table_row() {
        // Go tool assets, settings and plain secrets retire, whatever the case of the name.
        for n in [
            "cache",
            "CACHE",
            "user-assets.json",
            "credentials.yaml",
            ".env.typesafe",
            "config",
        ] {
            assert_eq!(d(n, "folder", "go"), ("retire", "go_item"), "{n}");
        }
        assert_eq!(d("specs", "folder", "go"), ("keep", "user_data"));
        assert_eq!(d("projects", "folder", "go"), ("keep", "user_data"));
        assert_eq!(d("reports", "folder", "go"), ("keep", "owner_unknown"));
        assert_eq!(
            d("claude-profiles", "folder", "go"),
            ("keep", "owner_unknown")
        );
        assert_eq!(
            d("nothing-like-it", "folder", "go"),
            ("keep", "not_in_table")
        );
        // This generation's own names never retire.
        for n in ["vault", "versions", "receipts", "moai-cli.json"] {
            assert_eq!(d(n, "folder", "go"), ("keep", "this_generation"), "{n}");
        }
        // The key file stays until the merge step.
        assert_eq!(
            d(".env.glm", "file", "go"),
            ("keep", "merge_candidate_kept")
        );
        // A link or a special file is never moved, a secret included.
        assert_eq!(d("credentials.yaml", "link", "go"), ("keep", "link"));
        assert_eq!(d(".env", "special", "go"), ("keep", "special_file"));
    }

    #[test]
    fn a_name_both_generations_write_retires_only_on_a_pure_go_home() {
        for n in ["state", "config", "backups"] {
            assert_eq!(d(n, "folder", "go"), ("retire", "go_item"), "{n}");
            for g in ["mixed", "rust-first"] {
                assert_eq!(d(n, "folder", g), ("keep", "shared_name"), "{n} on {g}");
            }
        }
        // A name only Go writes retires on every generation that has Go items.
        for g in ["go", "mixed", "rust-first"] {
            assert_eq!(d("cache", "folder", g), ("retire", "go_item"), "{g}");
        }
    }

    /// The `content_digest` holds no clock: two previews at different times give the same one
    /// (the `plan_digest` names the dated folder, so it differs). It follows the retirement set
    /// (what retires or is in use) and the HOME it was made for, and nothing else: an entry
    /// that stays (a name this generation writes, user data, an unknown name) does not move it,
    /// so the layer's own writes do not make a confirmed token stale.
    #[test]
    fn the_content_digest_is_clock_free_and_follows_the_entries() {
        let dir = tempfile::tempdir().unwrap();
        let home = fs::canonicalize(dir.path()).unwrap();
        let moai = home.join(".moai");
        fs::create_dir_all(moai.join("state")).unwrap();
        fs::write(moai.join("user-assets.json"), "{}").unwrap();
        fs::write(moai.join("state/active-sessions.json"), "[]").unwrap();
        fs::create_dir_all(moai.join("cache")).unwrap();
        let a = plan(&home, None, "2026-11-14T12:00:00Z").unwrap();
        let b = plan(&home, None, "2026-11-14T13:30:00Z").unwrap();
        assert_eq!(a["action"], json!("retire"), "{a}");
        assert_eq!(a["content_digest"], b["content_digest"]);
        assert_ne!(a["plan_digest"], b["plan_digest"]);
        // Entries that stay do not move the digest.
        fs::write(moai.join("appeared-later"), "x").unwrap();
        fs::create_dir_all(moai.join("agents")).unwrap();
        let kept = plan(&home, None, "2026-11-14T12:00:00Z").unwrap();
        assert_eq!(a["content_digest"], kept["content_digest"], "{kept}");
        // A new retire candidate does.
        fs::create_dir_all(moai.join("run")).unwrap();
        let c = plan(&home, None, "2026-11-14T12:00:00Z").unwrap();
        assert_ne!(a["content_digest"], c["content_digest"]);
    }

    /// Two HOMEs with the same layout do not share a `content_digest`.
    #[test]
    fn the_content_digest_names_its_home() {
        let make = || {
            let dir = tempfile::tempdir().unwrap();
            let home = fs::canonicalize(dir.path()).unwrap();
            let moai = home.join(".moai");
            fs::create_dir_all(moai.join("state")).unwrap();
            fs::write(moai.join("user-assets.json"), "{}").unwrap();
            fs::write(moai.join("state/active-sessions.json"), "[]").unwrap();
            fs::create_dir_all(moai.join("cache")).unwrap();
            (dir, home)
        };
        let (_a, ha) = make();
        let (_b, hb) = make();
        let now = "2026-11-14T12:00:00Z";
        let (pa, pb) = (plan(&ha, None, now).unwrap(), plan(&hb, None, now).unwrap());
        assert_eq!(pa["action"], pb["action"]);
        assert_ne!(pa["content_digest"], pb["content_digest"]);
    }

    #[test]
    fn every_go_row_has_a_decision_and_no_rust_row_retires() {
        for e in home_layout::HOME_TABLE {
            let name = e.name.replace('*', "x");
            let (action, _) = decide(Some(e), "folder", "go");
            if !e.go {
                assert_eq!(action, "keep", "{name}");
            }
            if matches!(e.rule, Rule::UserDataKeep | Rule::UnknownKeep) {
                assert_eq!(action, "keep", "{name}");
            }
        }
    }
}
