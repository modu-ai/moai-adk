//! #2871 U3a (PLAN comment 6881562064, user decision D-M1 6876633358): the read-only preview of
//! moving this generation's items from `~/.moai-cli` to the `~/.moai` top level. Each top-level
//! entry of `~/.moai-cli` that the HOME name table ([`crate::home_layout::HOME_TABLE`]) says this
//! generation writes would move, by one rename, to `~/.moai/<name>`. User data, the old shared
//! worktree root, names the table does not know, links and the marker stay. Nothing here writes,
//! opens a secret or follows a link: the vault is judged by `lstat` alone.
//!
//! The preview also gathers what the later pieces need and cannot recompute safely: the install
//! id of every version folder before and after the move (the id is a hash of the binary's
//! canonical path, so it changes with the folder), how each `path` entry would be re-pointed,
//! and the modes of the vault. A HOME whose marker already has schema 2 has moved; what an older
//! build wrote back into `~/.moai-cli` since is listed as `late_arrivals` and nothing else.
//!
//! The user's order is retire the Go items first, then install: while Go items remain in
//! `~/.moai` the preview is blocked (`go_items_not_retired`), and a name that is still taken
//! there blocks the move (`destination_taken`), ignoring case.

use crate::home_layout::{self, Entry, Rule, RustUse};
use crate::home_retire::{class_str, kind_of, rule_str};
use crate::home_state::{stamp, CANDIDATE_LIMIT};
use crate::move_journal;
use crate::state_root::{classify_home, MOAI_DIR, RUST_MARKER};
use serde_json::{json, Value};
use std::collections::{BTreeMap, BTreeSet};
use std::fs;
use std::path::Path;

/// The folder inside `~/.moai-cli` that will hold the install journal and manifest. It is not an
/// entry to move: it stays behind with the old folder.
pub const JOURNAL_DIR: &str = "home-install";
/// The journal `kind` of an install.
pub const KIND: &str = "home-install";
/// The prefix of the confirmation token, apart from the other layers' (`home:` is the Go retire).
pub const TOKEN_PREFIX: &str = "install:";
const MANIFEST_PREFIX: &str = "install-manifest-";
/// The `path` entries listed in a stub plan, and the longest link target shown.
const STUB_LIMIT: usize = 50;
/// The binary every version folder holds, relative to the folder: `bin/moai-cli` with the
/// platform's executable suffix (`.exe` on Windows, where the installer places it).
fn version_bin(suffix: &str) -> String {
    format!("bin/moai-cli{suffix}")
}

/// What happens to the top-level entry `name` of `kind` that the table row `entry` names:
/// `("move" | "keep", reason)`. Pure; whether the destination is free is judged afterwards.
pub fn decide(entry: Option<&Entry>, name: &str, kind: &str) -> (&'static str, &'static str) {
    if name.eq_ignore_ascii_case(JOURNAL_DIR) {
        return ("keep", "install_journal");
    }
    match kind {
        "file" | "folder" => {}
        "link" => return ("keep", "link"),
        _ => return ("keep", "special_file"),
    }
    let Some(e) = entry else {
        return ("keep", "not_in_table");
    };
    if e.name == RUST_MARKER {
        // The schema-2 marker is the last thing written; whatever sits under this name now is
        // not it.
        return ("keep", "marker_written_last");
    }
    if e.name == "worktrees" {
        // The old shared worktree root: moving it would break the checkouts registered in it.
        return ("keep", "worktrees_stay");
    }
    if e.rule == Rule::UserDataKeep {
        return ("keep", "user_data");
    }
    if e.rust != RustUse::Writes {
        return ("keep", "not_written_by_this_generation");
    }
    ("move", "rust_item")
}

/// `<digest>` as the token: `install:` and sixteen hex digits of the `content_digest`.
pub fn token_of(content_digest: &str) -> String {
    let hex = content_digest
        .strip_prefix("sha256:")
        .unwrap_or(content_digest);
    format!("{TOKEN_PREFIX}{}", hex.chars().take(16).collect::<String>())
}

/// `dev:inode` of an entry, from its own metadata (a link is not followed).
#[cfg(unix)]
fn identity(p: &Path) -> String {
    use std::os::unix::fs::MetadataExt;
    fs::symlink_metadata(p).map_or(String::new(), |m| format!("{}:{}", m.dev(), m.ino()))
}

#[cfg(not(unix))]
fn identity(_p: &Path) -> String {
    String::new()
}

#[cfg(unix)]
fn device(p: &Path) -> Option<u64> {
    use std::os::unix::fs::MetadataExt;
    fs::symlink_metadata(p).ok().map(|m| m.dev())
}

#[cfg(not(unix))]
fn device(_p: &Path) -> Option<u64> {
    None
}

/// The permission bits of `p` as octal text, from its own metadata.
#[cfg(unix)]
fn mode_of(p: &Path) -> Option<u32> {
    use std::os::unix::fs::PermissionsExt;
    fs::symlink_metadata(p)
        .ok()
        .map(|m| m.permissions().mode() & 0o777)
}

#[cfg(not(unix))]
fn mode_of(_p: &Path) -> Option<u32> {
    None
}

#[cfg(unix)]
fn link_count(p: &Path) -> u64 {
    use std::os::unix::fs::MetadataExt;
    fs::symlink_metadata(p).map_or(1, |m| m.nlink())
}

#[cfg(not(unix))]
fn link_count(_p: &Path) -> u64 {
    1
}

fn octal(mode: Option<u32>) -> Value {
    mode.map_or(Value::Null, |m| Value::from(format!("{m:03o}")))
}

fn canon(p: &Path) -> std::path::PathBuf {
    fs::canonicalize(p).unwrap_or_else(|_| p.to_path_buf())
}

/// The names under `dir`, sorted; `unreadable` counts names that are not UTF-8 and entries the
/// listing could not read. A folder that cannot be listed at all gives nothing here: callers that
/// must not take that for "empty" ask [`listable`] first.
fn names_of(dir: &Path) -> (Vec<String>, usize) {
    let mut names = Vec::new();
    let mut unreadable = 0;
    if let Ok(rd) = fs::read_dir(dir) {
        for e in rd {
            match e.map(|e| e.file_name().into_string()) {
                Ok(Ok(n)) => names.push(n),
                Ok(Err(_)) | Err(_) => unreadable += 1,
            }
        }
    }
    names.sort();
    (names, unreadable)
}

/// Whether the entries of `dir` can be listed.
fn listable(dir: &Path) -> bool {
    fs::read_dir(dir).is_ok()
}

fn install_id_of(path: &Path) -> Option<String> {
    crate::install::install_id(path).ok()
}

/// The install id of every version folder before and after the move. A folder without a regular
/// `bin/moai-cli` has no id and is listed apart.
fn alias_table(home: &Path, old: &Path) -> (BTreeMap<String, String>, Vec<String>) {
    let versions = old.join("versions");
    let mut table = BTreeMap::new();
    let mut skipped = Vec::new();
    if !fs::symlink_metadata(&versions).is_ok_and(|m| m.is_dir()) {
        return (table, skipped);
    }
    let chome = canon(home);
    let (names, _) = names_of(&versions);
    let rel = version_bin(std::env::consts::EXE_SUFFIX);
    for v in names.iter().take(CANDIDATE_LIMIT) {
        // Every component is judged by `lstat`: a version folder or its `bin` that is a link
        // would make the canonical path (the old identity) another one than the moved path (the
        // new identity), so such a folder gets no alias.
        let folder = versions.join(v);
        let real_dir = |p: &Path| fs::symlink_metadata(p).is_ok_and(|m| m.is_dir());
        let bin = folder.join(&rel);
        if !real_dir(&folder)
            || !real_dir(&folder.join("bin"))
            || !fs::symlink_metadata(&bin).is_ok_and(|m| m.is_file())
        {
            skipped.push(v.clone());
            continue;
        }
        let new_bin = home_layout::rust_path(&chome, &format!("versions/{v}/{rel}"));
        match (install_id_of(&canon(&bin)), install_id_of(&new_bin)) {
            (Some(a), Some(b)) => {
                table.insert(a, b);
            }
            _ => skipped.push(v.clone()),
        }
    }
    (table, skipped)
}

/// How each `path` entry would follow the move: a link into the old `versions` folder must be
/// pointed at the new one; a file moves as it is.
fn stub_plan(old: &Path) -> Vec<Value> {
    let dir = old.join("path");
    if !fs::symlink_metadata(&dir).is_ok_and(|m| m.is_dir()) {
        return Vec::new();
    }
    let old_versions = old.join("versions");
    let (names, _) = names_of(&dir);
    names
        .iter()
        .take(STUB_LIMIT)
        .map(|n| {
            let p = dir.join(n);
            let kind = kind_of(&p);
            let target = if kind == "link" {
                fs::read_link(&p)
                    .ok()
                    .map(|t| t.display().to_string())
                    .map_or(Value::Null, Value::from)
            } else {
                Value::Null
            };
            let repoint = target
                .as_str()
                .is_some_and(|t| Path::new(t).starts_with(&old_versions));
            json!({"name": n, "kind": kind, "target": target, "repoint": repoint})
        })
        .collect()
}

/// The vault's own metadata: its mode, the master key's mode, how many entries it holds and how
/// many of those are links or have more than one name. Nothing inside is opened.
fn vault_facts(old: &Path) -> Value {
    let vault = old.join("vault");
    let Ok(meta) = fs::symlink_metadata(&vault) else {
        return json!({"kind": "absent"});
    };
    let kind = kind_of(&vault);
    if !meta.is_dir() {
        return json!({"kind": kind, "mode": octal(mode_of(&vault))});
    }
    let (names, unreadable) = names_of(&vault);
    let master = vault.join(".master.key");
    let master_mode = if fs::symlink_metadata(&master).is_ok() {
        mode_of(&master)
    } else {
        None
    };
    let links = names
        .iter()
        .filter(|n| kind_of(&vault.join(n)) == "link")
        .count();
    let multi_link = names
        .iter()
        // A folder's link count grows with its sub-folders; only a file with several names counts.
        .filter(|n| kind_of(&vault.join(n)) != "folder" && link_count(&vault.join(n)) > 1)
        .count();
    let loose = [mode_of(&vault), master_mode]
        .into_iter()
        .flatten()
        .any(|m| m & 0o077 != 0);
    json!({
        "kind": kind,
        "mode": octal(mode_of(&vault)),
        "master_key_mode": octal(master_mode),
        "entries": names.len() + unreadable,
        "links": links,
        "multi_link": multi_link,
        "loose": loose,
    })
}

/// The preview for the HOME at `home`. `moai_home` is the `MOAI_HOME` value, `now` the clock as
/// `YYYY-MM-DDTHH:MM:SSZ` (names the journal). Writes nothing.
pub fn plan(home: &Path, moai_home: Option<&str>, now: &str) -> Result<Value, String> {
    if !home.is_absolute() {
        return Err("HOME is not an absolute path".into());
    }
    let stamp = stamp(now)?;
    let old = home.join(crate::HOME_STATE_DIR);
    let moai = home.join(MOAI_DIR);
    let c = classify_home(home);
    let old_kind = match fs::symlink_metadata(&old) {
        Err(_) => "absent",
        Ok(m) if m.file_type().is_symlink() => "link",
        Ok(m) if m.is_dir() => "folder",
        Ok(_) => "other",
    };
    let moved = matches!(c.generation, "rust" | "rust-first");
    let movable = matches!(c.generation, "none" | "empty" | "go" | "mixed");

    let mut blocks: Vec<String> = Vec::new();
    if old_kind != "absent" {
        if !moved && !movable {
            blocks.push(format!("generation_{}", c.generation));
        }
        if old_kind != "folder" {
            blocks.push("state_dir_not_a_folder".into());
        }
    }

    let listed = old_kind == "folder";
    if listed && !listable(&old) {
        // A folder that cannot be listed is not an empty one: say so instead of "nothing to do".
        blocks.push("state_dir_unreadable".into());
    }
    let (names, unreadable) = if listed {
        names_of(&old)
    } else {
        (Vec::new(), 0)
    };
    let total = names.len() + unreadable;
    let journal_root = old.join(JOURNAL_DIR);
    let mut entries: Vec<Value> = Vec::new();
    let mut late: Vec<Value> = Vec::new();

    if listed && moved {
        for n in names
            .iter()
            .filter(|n| !n.eq_ignore_ascii_case(JOURNAL_DIR))
            .take(CANDIDATE_LIMIT)
        {
            late.push(json!({"name": n, "kind": kind_of(&old.join(n))}));
        }
    }

    if listed && movable {
        let taken: BTreeSet<String> = if fs::symlink_metadata(&moai).is_ok_and(|m| m.is_dir()) {
            if !listable(&moai) {
                blocks.push("destination_unreadable".into());
            }
            names_of(&moai)
                .0
                .iter()
                .map(|n| n.to_ascii_lowercase())
                .collect()
        } else {
            BTreeSet::new()
        };
        if moai_home.is_some_and(|m| !m.is_empty() && Path::new(m) != moai) {
            blocks.push("moai_home_elsewhere".into());
        }
        match fs::symlink_metadata(&journal_root) {
            Ok(m) if !m.is_dir() => blocks.push("journal_root_not_a_folder".into()),
            _ => {}
        }
        let stamp_files = [
            move_journal::journal_path(&journal_root, &stamp),
            journal_root.join(format!("{MANIFEST_PREFIX}{stamp}.json")),
        ];
        if stamp_files.iter().any(|p| fs::symlink_metadata(p).is_ok()) {
            blocks.push("stamp_taken".into());
        }
        let under = if taken.is_empty() && !moai.exists() {
            home
        } else {
            &moai
        };
        if let (Some(a), Some(b)) = (device(&old), device(under)) {
            if a != b {
                blocks.push("different_file_system".into());
            }
        }
        if total > CANDIDATE_LIMIT {
            blocks.push("not_fully_listed".into());
        }
        if unreadable > 0 {
            blocks.push(format!("unreadable_names:{unreadable}"));
        }
        if matches!(c.generation, "go" | "mixed") {
            let retire = crate::home_retire::plan(home, moai_home, now)?;
            let pending = ["retire", "skip-in-use"]
                .iter()
                .map(|k| retire["counts"][*k].as_u64().unwrap_or(0))
                .sum::<u64>();
            if pending > 0 {
                blocks.push(format!("go_items_not_retired:{pending}"));
            } else if retire["action"] == "blocked" {
                // The retirement preview could not list or judge everything (it lists at most
                // 500 names), so zero pending items proves nothing about the Go items.
                blocks.push("go_retire_preview_blocked".into());
            }
        }

        for n in names.iter().take(CANDIDATE_LIMIT) {
            let raw = old.join(n);
            let kind = kind_of(&raw);
            let row = home_layout::lookup(n);
            let (mut action, mut reason) = decide(row, n, kind);
            // The listing compare folds ASCII case only; the file system is asked as well, so a
            // volume that folds more (case-insensitive APFS, Unicode forms) still reports the clash.
            let clash = taken.contains(&n.to_ascii_lowercase())
                || fs::symlink_metadata(moai.join(n)).is_ok();
            if action == "move" && clash {
                action = "blocked";
                reason = "destination_taken";
                blocks.push(format!("destination_taken:{n}"));
            }
            let destination = if action == "keep" {
                Value::Null
            } else {
                Value::from(home_layout::rust_path(home, n).display().to_string())
            };
            entries.push(json!({
                "name": n,
                "kind": kind,
                "class": row.map_or("unknown", |e| class_str(e.class)),
                "rule": row.map_or("none", |e| rule_str(e.rule)),
                "action": action,
                "reason": reason,
                "destination": destination,
                "identity": identity(&raw),
            }));
        }
    }

    if listed && movable {
        // The alias table and the stub plan are bounded lists the later steps rely on; a cut one
        // would leave a version folder without an alias or a `path` link pointing nowhere.
        // A link or a plain file there is not a list the later steps walk (`alias_table` and the
        // stub plan skip it), so what it points at must not count against the bound.
        let listed_in = |sub: &str| {
            let dir = old.join(sub);
            if fs::symlink_metadata(&dir).is_ok_and(|m| m.is_dir()) {
                names_of(&dir).0.len()
            } else {
                0
            }
        };
        if listed_in("versions") > CANDIDATE_LIMIT {
            blocks.push("versions_not_fully_listed".into());
        }
        if listed_in("path") > STUB_LIMIT {
            blocks.push("path_not_fully_listed".into());
        }
    }

    let count = |a: &str| entries.iter().filter(|e| e["action"] == a).count();
    let counts = json!({
        "move": count("move"),
        "keep": count("keep"),
        "blocked": count("blocked"),
    });
    let action = if old_kind == "absent" {
        "none"
    } else if !blocks.is_empty() {
        "blocked"
    } else if moved || count("move") == 0 {
        "none"
    } else {
        "install"
    };

    let (aliases, alias_skipped) = if listed && !moved {
        alias_table(home, &old)
    } else {
        (BTreeMap::new(), Vec::new())
    };
    let stub = if listed && !moved {
        stub_plan(&old)
    } else {
        Vec::new()
    };
    let vault = if listed && !moved {
        vault_facts(&old)
    } else {
        Value::Null
    };

    let mut warnings = Vec::new();
    if total > CANDIDATE_LIMIT {
        warnings.push(format!(
            "{total} entries in {}; the first {CANDIDATE_LIMIT} are listed",
            old.display()
        ));
    }
    if vault["loose"] == json!(true) {
        warnings.push(
            "the vault or its master key has modes looser than 0700/0600; a move keeps the modes as they are"
                .to_string(),
        );
    }
    if vault["multi_link"].as_u64().unwrap_or(0) > 0 {
        warnings.push("a vault entry has more than one name; a move keeps it as it is".to_string());
    }
    if !late.is_empty() {
        warnings.push(format!(
            "{} entries sit in {} although this HOME has moved; merging them is a later step",
            late.len(),
            old.display()
        ));
    }

    // The digest of what the preview decided, without the clock; the confirmation token is its
    // first sixteen hex digits, and the apply recomputes it before it moves anything.
    let identities = entries
        .iter()
        .map(|e| {
            json!({
                "name": e["name"], "kind": e["kind"], "action": e["action"],
                "reason": e["reason"], "identity": e["identity"],
            })
        })
        .collect::<Vec<_>>();
    let content = json!({
        "mode": "install",
        "generation": c.generation,
        "state_dir": old_kind,
        "action": action,
        "blocks": blocks,
        "entries": identities,
        "late_arrivals": late,
        "alias_table": aliases,
        "alias_skipped": alias_skipped,
        "stub": stub,
        "vault": vault,
    });
    let content_digest = crate::digest(
        serde_json::to_string(&content)
            .unwrap_or_default()
            .as_bytes(),
    );
    let plan_bound = json!({"content": content_digest, "stamp": stamp});
    let plan_digest = crate::digest(
        serde_json::to_string(&plan_bound)
            .unwrap_or_default()
            .as_bytes(),
    );
    let confirm_token = (action == "install").then(|| token_of(&content_digest));

    Ok(json!({
        "schema_version": 1,
        "mode": "install",
        "home": home.display().to_string(),
        "state": c.to_json(),
        "state_dir": old_kind,
        "action": action,
        "journal": journal_root.display().to_string(),
        "blocks": blocks,
        "entries": entries,
        "counts": counts,
        "total": total,
        "late_arrivals": late,
        "alias_table": aliases,
        "alias_skipped": alias_skipped,
        "stub": stub,
        "vault": vault,
        "warnings": warnings,
        "plan_digest": plan_digest,
        "content_digest": content_digest,
        "confirm_token": confirm_token,
        "apply": "not in this build yet (#2871 U3c); on a real HOME it will run only after the user confirms",
        "writes": false,
    }))
}

/// The most install journals read; a journal folder with more is read newest-name last up to
/// this many and the rest ignored (the reader is a lookup, never a gate).
const ALIAS_JOURNAL_LIMIT: usize = 64;

/// An install id as `install::install_id` makes it: the lower-case hex of a SHA-256, 64 digits.
/// A journal is a file in the user's HOME and its aliases become path parts under
/// `installations/`, so a name of any other shape (`../x`, `a/b`, empty, long) is never used
/// (audit 6884416142 B1).
pub(crate) fn install_id_shaped(s: &str) -> bool {
    s.len() == 64 && s.bytes().all(|b| matches!(b, b'0'..=b'9' | b'a'..=b'f'))
}

/// R5 (#2871 U3b): old install id to new install id, from the `committed` install journals in
/// `<home>/<HOME_STATE_DIR>/home-install/`.
///
/// The install id is a hash of the binary's canonical path, so moving a version folder to
/// `~/.moai/versions` changes it, while the receipts keep the id they were written under. A
/// reader that wants the receipt of a build asks for its earlier ids too. Only a committed
/// journal counts: a pending or reverted one lends nothing, so a revert takes the ids back. The
/// journal folder and every file are read by `lstat` first (a link lends nothing); a file that
/// is not a valid journal, is of another `kind` or is unreadable is skipped, never an error.
/// With no journal this is empty and every caller behaves as before. A later journal (by name)
/// wins over an earlier one for the same old id. Reads names and JSON only.
pub fn committed_aliases(home: &Path) -> BTreeMap<String, String> {
    let dir = home.join(crate::HOME_STATE_DIR).join(JOURNAL_DIR);
    let mut out = BTreeMap::new();
    if !fs::symlink_metadata(&dir).is_ok_and(|m| m.is_dir()) {
        return out;
    }
    let (names, _) = names_of(&dir);
    let mut names: Vec<String> = names
        .into_iter()
        .filter(|n| n.starts_with(move_journal::PREFIX) && n.ends_with(".json"))
        .collect();
    names.sort();
    for n in names.iter().take(ALIAS_JOURNAL_LIMIT) {
        let Ok(j) = move_journal::read_journal(&dir.join(n)) else {
            continue;
        };
        if j.kind == KIND && j.state == move_journal::JournalState::Committed {
            out.extend(
                j.id_aliases
                    .into_iter()
                    .filter(|(old, new)| install_id_shaped(old) && install_id_shaped(new)),
            );
        }
    }
    out
}

/// The earlier install ids that `new_id` replaced, sorted. Empty with no committed journal.
pub fn earlier_ids_of(home: &Path, new_id: &str) -> Vec<String> {
    committed_aliases(home)
        .into_iter()
        .filter(|(_, new)| new == new_id)
        .map(|(old, _)| old)
        .collect()
}

#[cfg(test)]
mod tests {
    use super::*;

    /// Codex review P2 on #2987: the installer places `bin/moai-cli.exe` on Windows.
    #[test]
    fn the_version_binary_carries_the_platform_suffix() {
        assert_eq!(version_bin(""), "bin/moai-cli");
        assert_eq!(version_bin(".exe"), "bin/moai-cli.exe");
    }

    fn row(name: &str) -> Option<&'static Entry> {
        home_layout::lookup(name)
    }

    /// The earlier state folder, whatever generation the HOME is in.
    fn old_dir(home: &Path) -> std::path::PathBuf {
        home.join(crate::HOME_STATE_DIR)
    }

    #[test]
    fn decide_names_the_reason_of_every_keep() {
        for (name, kind, want) in [
            ("agents", "folder", ("move", "rust_item")),
            ("vault", "folder", ("move", "rust_item")),
            ("schedule", "folder", ("move", "rust_item")),
            ("home-state-receipt.json", "file", ("move", "rust_item")),
            ("agents", "link", ("keep", "link")),
            ("agents", "special", ("keep", "special_file")),
            ("home-install", "folder", ("keep", "install_journal")),
            ("Home-Install", "folder", ("keep", "install_journal")),
            ("worktrees", "folder", ("keep", "worktrees_stay")),
            ("projects", "folder", ("keep", "user_data")),
            ("bin", "folder", ("keep", "not_written_by_this_generation")),
            (
                ".env.glm",
                "file",
                ("keep", "not_written_by_this_generation"),
            ),
            ("moai-cli.json", "file", ("keep", "marker_written_last")),
            (
                "hook-config.toml",
                "file",
                ("keep", "not_written_by_this_generation"),
            ),
        ] {
            assert_eq!(decide(row(name), name, kind), want, "{name} {kind}");
        }
        assert_eq!(decide(None, "mystery", "folder"), ("keep", "not_in_table"));
    }

    #[test]
    fn a_token_is_sixteen_hex_digits_with_its_own_prefix() {
        assert_eq!(
            token_of("sha256:0123456789abcdef0123456789abcdef"),
            "install:0123456789abcdef"
        );
        assert_eq!(token_of("0123456789abcdef9"), "install:0123456789abcdef");
    }

    /// The digest the token comes from follows the decision and never the clock: it moves with
    /// each field a later apply relies on, and with nothing else.
    #[test]
    fn the_content_digest_inputs_are_pinned() {
        let dir = tempfile::tempdir().unwrap();
        let home = fs::canonicalize(dir.path()).unwrap();
        let old = old_dir(&home);
        for (rel, body) in [
            ("agents/a.json", "{}"),
            ("vault/.master.key", "k"),
            ("path/moai", "#!/bin/sh\n"),
            ("versions/1.0.0/bin/moai-cli", "bin"),
        ] {
            let p = old.join(rel);
            fs::create_dir_all(p.parent().unwrap()).unwrap();
            fs::write(p, body).unwrap();
        }
        let digest = |now: &str| {
            plan(&home, None, now).unwrap()["content_digest"]
                .as_str()
                .unwrap()
                .to_string()
        };
        let base = digest("2026-11-14T12:00:00Z");
        assert_eq!(
            base,
            digest("2026-11-14T13:30:00Z"),
            "the clock is not an input"
        );
        // A new entry, a new alias version, a different vault mode: each moves the digest.
        fs::create_dir_all(old.join("receipts")).unwrap();
        let with_entry = digest("2026-11-14T12:00:00Z");
        assert_ne!(base, with_entry);
        let bin = old.join("versions/2.0.0/bin/moai-cli");
        fs::create_dir_all(bin.parent().unwrap()).unwrap();
        fs::write(&bin, "bin2").unwrap();
        let with_alias = digest("2026-11-14T12:00:00Z");
        assert_ne!(with_entry, with_alias);
        #[cfg(unix)]
        {
            use std::os::unix::fs::PermissionsExt;
            // The folder was made 0755; 0700 is a different vault mode.
            fs::set_permissions(old.join("vault"), fs::Permissions::from_mode(0o700)).unwrap();
            assert_ne!(with_alias, digest("2026-11-14T12:00:00Z"));
        }
    }

    #[test]
    fn a_bad_clock_is_refused_and_a_relative_home_too() {
        assert!(plan(Path::new("relative"), None, "2026-11-14T12:00:00Z").is_err());
        let dir = tempfile::tempdir().unwrap();
        assert!(plan(dir.path(), None, "yesterday").is_err());
    }

    /// `class_str` and `rule_str` are shared with the retire preview and must stay total.
    #[test]
    fn the_shared_labels_cover_every_table_row() {
        for e in home_layout::HOME_TABLE {
            assert!(!class_str(e.class).is_empty() && !rule_str(e.rule).is_empty());
        }
    }

    fn journal(dir: &Path, stamp: &str, state: move_journal::JournalState, pairs: &[(&str, &str)]) {
        let j = move_journal::Journal {
            schema_version: move_journal::SCHEMA_VERSION,
            kind: KIND.to_string(),
            stamp: stamp.to_string(),
            state,
            id_aliases: pairs
                .iter()
                .map(|(a, b)| (a.to_string(), b.to_string()))
                .collect(),
            items: Vec::new(),
        };
        fs::create_dir_all(dir).unwrap();
        fs::write(
            move_journal::journal_path(dir, stamp),
            serde_json::to_vec(&j).unwrap(),
        )
        .unwrap();
    }

    /// An install id as the real one is made: the SHA-256 of a tag, in lower-case hex.
    fn hid(tag: &str) -> String {
        use sha2::{Digest, Sha256};
        format!("{:x}", Sha256::digest(tag.as_bytes()))
    }

    /// R5 (#2871 U3b): only committed journals lend ids; a later one wins for the same old id.
    #[test]
    fn committed_aliases_read_committed_journals_and_the_later_one_wins() {
        use move_journal::JournalState::{Committed, Pending, Reverted};
        let t = tempfile::tempdir().unwrap();
        let home = t.path();
        assert!(committed_aliases(home).is_empty(), "no folder, no ids");
        let dir = old_dir(home).join(JOURNAL_DIR);
        let (a, a1, a2, b, b1) = (hid("a"), hid("a1"), hid("a2"), hid("b"), hid("b1"));
        let (c, c1, d, d1) = (hid("c"), hid("c1"), hid("d"), hid("d1"));
        journal(&dir, "20261114T100000Z", Committed, &[(&a, &a1), (&b, &b1)]);
        journal(&dir, "20261114T110000Z", Committed, &[(&a, &a2)]);
        journal(&dir, "20261114T120000Z", Pending, &[(&c, &c1)]);
        journal(&dir, "20261114T130000Z", Reverted, &[(&d, &d1)]);
        let got = committed_aliases(home);
        assert_eq!(got.get(&a), Some(&a2), "{got:?}");
        assert_eq!(got.get(&b), Some(&b1), "{got:?}");
        assert!(!got.contains_key(&c) && !got.contains_key(&d), "{got:?}");
        assert_eq!(earlier_ids_of(home, &a2), vec![a.clone()]);
        assert!(
            earlier_ids_of(home, &a1).is_empty(),
            "replaced by a later journal"
        );
    }

    /// Audit 6884416142 B1: a journal is a file in the HOME, and an alias becomes a path part
    /// under `installations/`. Only a SHA-256 in lower-case hex (64 digits) is an install id;
    /// anything else lends nothing, key or value, and never fails the lookup.
    #[test]
    fn committed_aliases_lend_only_ids_shaped_like_an_install_id() {
        use move_journal::JournalState::Committed;
        assert!(install_id_shaped(&hid("x")));
        let good = hid("good");
        let upper = good.to_ascii_uppercase();
        let long = format!("{good}0");
        let short = &good[..63];
        let non_hex = format!("{}g", &good[..63]);
        let bad: Vec<&str> = vec![
            "../../elsewhere",
            "a/b",
            "..",
            ".",
            "",
            "old",
            &upper,
            &long,
            short,
            &non_hex,
        ];
        let huge = "a".repeat(100_000);
        let t = tempfile::tempdir().unwrap();
        let dir = old_dir(t.path()).join(JOURNAL_DIR);
        let mut pairs: Vec<(&str, &str)> = Vec::new();
        for b in &bad {
            pairs.push((b, &good));
            pairs.push((&good, b));
        }
        pairs.push((&huge, &good));
        let (old_id, new_id) = (hid("old"), hid("new"));
        pairs.push((&old_id, &new_id));
        journal(&dir, "20261114T100000Z", Committed, &pairs);
        let got = committed_aliases(t.path());
        assert_eq!(
            got.into_iter().collect::<Vec<_>>(),
            vec![(old_id.clone(), new_id.clone())],
            "only the well-formed pair stands"
        );
        assert!(earlier_ids_of(t.path(), &good).is_empty());
    }

    /// A file that is not a journal, a journal folder that is a link and a bad file lend nothing
    /// and never fail the lookup.
    #[cfg(unix)]
    #[test]
    fn committed_aliases_skip_what_is_not_a_plain_journal() {
        use move_journal::JournalState::Committed;
        let t = tempfile::tempdir().unwrap();
        let home = t.path();
        let dir = old_dir(home).join(JOURNAL_DIR);
        journal(
            &dir,
            "20261114T100000Z",
            Committed,
            &[(&hid("a"), &hid("a1"))],
        );
        fs::write(dir.join("move-journal-20261114T110000Z.json"), "{not json").unwrap();
        fs::write(dir.join("notes.txt"), "x").unwrap();
        assert_eq!(committed_aliases(home).len(), 1);
        // The same journal folder reached through a link lends nothing.
        let linked = tempfile::tempdir().unwrap();
        fs::create_dir_all(old_dir(linked.path())).unwrap();
        std::os::unix::fs::symlink(&dir, old_dir(linked.path()).join(JOURNAL_DIR)).unwrap();
        assert!(committed_aliases(linked.path()).is_empty());
    }

    /// The reader is a lookup with a bound: past the limit, the rest are ignored.
    #[test]
    fn committed_aliases_read_at_most_the_journal_limit() {
        use move_journal::JournalState::Committed;
        let t = tempfile::tempdir().unwrap();
        let dir = old_dir(t.path()).join(JOURNAL_DIR);
        for i in 0..=ALIAS_JOURNAL_LIMIT {
            let stamp = format!("20261114T{:02}{:02}00Z", i / 60, i % 60);
            journal(
                &dir,
                &stamp,
                Committed,
                &[(&hid(&format!("old{i:03}")), &hid("new"))],
            );
        }
        let got = committed_aliases(t.path());
        assert_eq!(
            got.len(),
            ALIAS_JOURNAL_LIMIT,
            "the {}th is ignored",
            ALIAS_JOURNAL_LIMIT + 1
        );
        assert!(got.contains_key(&hid("old063")) && !got.contains_key(&hid("old064")));
    }
}
