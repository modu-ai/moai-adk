//! `cleanup-plan`, `cleanup-apply`, `cleanup-revert-plan` and `cleanup-revert` (#2872 stage 1,
//! design 6872036842, stage note 6872171920): move what an update or the Go generation left in
//! HOME into one archive folder, only where it is proven to be a tool's own unchanged file, and
//! move it back on request. Nothing is deleted.
//!
//! The plan lists candidates with an action, `archive` or `keep`, and a reason:
//!
//! - projection extras: files in `~/.claude/skills/moai` and `~/.agents/skills/moai` the staged
//!   package does not carry. Archived when the host receipt owns the file with the bytes on disk
//!   (update-plan's `remove_blocked` with no edit) or the Go user-asset manifest proves it.
//! - Go sibling skills `~/.claude/skills/moai-*` and `~/.agents/skills/moai-*`: one folder
//!   each, archived only when every file inside is proven by the Go manifest
//!   (`~/.moai/user-assets.json`) and nothing inside is a link, a special file or a secret.
//! - Go agents directly in `~/.claude/agents` and `~/.codex/agents` whose name the manifest
//!   records; archived when the bytes match.
//! - `~/.moai` leftovers, only while `~/.moai` is the Go generation: a top-level entry of the
//!   `move-to-backup` class of [`crate::home_state::SELECTIVE_TABLE`] that is an empty folder or
//!   named `tmp-*`, `scratch` or `cache`. Every other top-level entry is listed as `keep`.
//!
//! Never archived: a path with a secret name (listed by name, never read or hashed), Go run data
//! (`claude-profiles`, `backups`, `state`, `releases`, `db`, `handoffs`, `gpt-appserver`,
//! `backup`), links and special files (never followed), and every Go item while a Go session is
//! live or that cannot be told (`in_use`, `in_use_unknown`). A walk over a limit keeps its group
//! (`over_limit`). The apply recomputes the plan and runs only with its `confirm_token`.

use crate::home_state::{is_secret, table_class, EntryClass};
use crate::state_root::{classify_with, read_limited, MOAI_DIR};
use crate::update::StagedPackage;
use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::collections::BTreeMap;
use std::fs;
use std::io::Write;
use std::path::{Path, PathBuf};

type Result<T> = std::result::Result<T, String>;

/// The archive folders live under `<state>/cleanup/<YYYYMMDDTHHMMSSZ>`.
pub const ARCHIVE_DIR: &str = "cleanup";
/// `<absolute original>\t<absolute archived>` per moved item.
pub const MANIFEST: &str = "manifest.tsv";
/// The apply's record: status `pending`, `committed` or `reverted`, and every item.
pub const RECEIPT: &str = "receipt.json";
const USER_ASSETS: &str = "user-assets.json";
/// Same bound as the home-state reader of the Go manifest.
const USER_ASSETS_LIMIT: u64 = 16 * 1024 * 1024;
const RECEIPT_LIMIT: u64 = 16 * 1024 * 1024;
/// `<stamp>`, then `<stamp>.2` … `<stamp>.64` when the second is taken.
const STAMP_TRIES: usize = 64;
/// Go run data under `~/.moai`: never archived.
const GO_RUNTIME: [&str; 8] = [
    "claude-profiles",
    "backups",
    "state",
    "releases",
    "db",
    "handoffs",
    "gpt-appserver",
    "backup",
];
/// The Go manifest's root slugs and the HOME folders they stand for.
const GO_ROOTS: [(&str, &str); 4] = [
    ("claude-skills", ".claude/skills"),
    ("claude-agents", ".claude/agents"),
    ("agents-skills", ".agents/skills"),
    ("codex-agents", ".codex/agents"),
];
const SIBLING_ROOTS: [&str; 2] = [".claude/skills", ".agents/skills"];
const AGENT_ROOTS: [&str; 2] = [".claude/agents", ".codex/agents"];

/// The plan's bounds; over one, the affected group is kept (fail-closed).
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub(crate) struct Limits {
    /// Items one plan lists.
    pub items: usize,
    /// Entries one walked folder may hold.
    pub tree_entries: usize,
    /// Bytes one walked folder (or one file) may hold.
    pub tree_bytes: u64,
    /// Folder depth of one walk.
    pub depth: usize,
}

pub(crate) const LIMITS: Limits = Limits {
    items: 2048,
    tree_entries: 4096,
    tree_bytes: 64 * 1024 * 1024,
    depth: 32,
};

/// A name that holds secrets: the home-state rule plus the selective table's secret rows
/// (`gateway-auth`).
fn secret_name(name: &str) -> bool {
    is_secret(name) || table_class(name).0 == EntryClass::SecretManual
}

fn hex(bytes: &[u8]) -> String {
    format!("{:x}", Sha256::digest(bytes))
}

fn exists(p: &Path) -> bool {
    fs::symlink_metadata(p).is_ok()
}

// ---- walking and digests ----

#[derive(Debug, PartialEq, Eq)]
enum Seen {
    File(String),
    Dir(String),
    Link(String),
    Special(String),
    /// A secret name: never opened, walked or hashed.
    Secret(String),
}

impl Seen {
    fn rel(&self) -> &str {
        match self {
            Seen::File(r) | Seen::Dir(r) | Seen::Link(r) | Seen::Special(r) | Seen::Secret(r) => r,
        }
    }
}

/// Every entry under `root` (relative paths, sorted), never following a link and never
/// descending into a secret. `Err` is a reason: `over_limit` or `unreadable`.
fn list(root: &Path, limits: &Limits) -> std::result::Result<Vec<Seen>, &'static str> {
    fn walk(
        root: &Path,
        dir: &Path,
        depth: usize,
        limits: &Limits,
        out: &mut Vec<Seen>,
        bytes: &mut u64,
    ) -> std::result::Result<(), &'static str> {
        if depth > limits.depth {
            return Err("over_limit");
        }
        for e in fs::read_dir(dir).map_err(|_| "unreadable")? {
            let e = e.map_err(|_| "unreadable")?;
            let path = e.path();
            let rel = path
                .strip_prefix(root)
                .ok()
                .and_then(|r| r.to_str())
                .ok_or("unreadable")?
                .to_string();
            if out.len() >= limits.tree_entries {
                return Err("over_limit");
            }
            let name = e.file_name();
            let name = name.to_str().ok_or("unreadable")?;
            if secret_name(name) {
                out.push(Seen::Secret(rel));
                continue;
            }
            let m = fs::symlink_metadata(&path).map_err(|_| "unreadable")?;
            if m.file_type().is_symlink() {
                out.push(Seen::Link(rel));
            } else if m.is_dir() {
                out.push(Seen::Dir(rel));
                walk(root, &path, depth + 1, limits, out, bytes)?;
            } else if m.is_file() {
                *bytes += m.len();
                if *bytes > limits.tree_bytes {
                    return Err("over_limit");
                }
                out.push(Seen::File(rel));
            } else {
                out.push(Seen::Special(rel));
            }
        }
        Ok(())
    }
    let mut out = Vec::new();
    let mut bytes = 0u64;
    walk(root, root, 0, limits, &mut out, &mut bytes)?;
    out.sort_by(|a, b| a.rel().cmp(b.rel()));
    Ok(out)
}

/// The bare-hex sha256 of a regular file; never follows a link.
fn file_sha(path: &Path, limits: &Limits) -> std::result::Result<String, &'static str> {
    let m = fs::symlink_metadata(path).map_err(|_| "unreadable")?;
    if m.file_type().is_symlink() {
        return Err("symlink");
    }
    if !m.is_file() {
        return Err("special_file");
    }
    if m.len() > limits.tree_bytes {
        return Err("over_limit");
    }
    fs::read(path).map(|b| hex(&b)).map_err(|_| "unreadable")
}

/// A folder's digest and its files as `(rel, sha)`.
type Tree = (String, Vec<(String, String)>);

/// A folder's digest over its sorted entries (`rel/` per folder, `rel\t<sha>` per file) and
/// its files as `(rel, sha)`. A link, special file or secret inside refuses it.
fn tree(path: &Path, limits: &Limits) -> std::result::Result<Tree, &'static str> {
    let seen = list(path, limits)?;
    let mut text = String::new();
    let mut files = Vec::new();
    for s in &seen {
        match s {
            Seen::Dir(r) => text.push_str(&format!("{r}/\n")),
            Seen::File(r) => {
                let sha = file_sha(&path.join(r), limits)?;
                text.push_str(&format!("{r}\t{sha}\n"));
                files.push((r.clone(), sha));
            }
            Seen::Link(_) => return Err("symlink"),
            Seen::Special(_) => return Err("special_file"),
            Seen::Secret(_) => return Err("secret"),
        }
    }
    Ok((hex(text.as_bytes()), files))
}

/// The digest an item was planned with: a file's sha or a folder's tree digest.
fn digest_of(path: &Path, dir: bool, limits: &Limits) -> std::result::Result<String, &'static str> {
    let m = fs::symlink_metadata(path).map_err(|_| "unreadable")?;
    if m.file_type().is_symlink() {
        return Err("symlink");
    }
    match (dir, m.is_dir()) {
        (true, true) => tree(path, limits).map(|(d, _)| d),
        (false, false) => file_sha(path, limits),
        _ => Err("kind_changed"),
    }
}

/// `Ok(true)` when every component of `home/rel` is a real folder, `Ok(false)` when one is
/// missing, `Err` when one is a link or not a folder.
fn real_dirs(home: &Path, rel: &str) -> std::result::Result<bool, &'static str> {
    let mut p = home.to_path_buf();
    for part in rel.split('/') {
        p.push(part);
        match fs::symlink_metadata(&p) {
            Err(_) => return Ok(false),
            Ok(m) if m.file_type().is_symlink() => return Err("symlink"),
            Ok(m) if !m.is_dir() => return Err("not_a_folder"),
            Ok(_) => {}
        }
    }
    Ok(true)
}

/// The HOME-relative parent of `rel`; `None` for a top-level name.
fn parent_rel(rel: &str) -> Option<&str> {
    rel.rsplit_once('/').map(|(p, _)| p)
}

/// Sorted UTF-8 names in `dir`; other names are skipped (they are never archived).
fn names(dir: &Path) -> Vec<String> {
    let mut v: Vec<String> = fs::read_dir(dir)
        .map(|rd| {
            rd.filter_map(|e| e.ok())
                .filter_map(|e| e.file_name().into_string().ok())
                .collect()
        })
        .unwrap_or_default();
    v.sort();
    v
}

// ---- the Go manifest and the session check ----

enum GoManifest {
    Absent,
    /// HOME-relative path → bare-hex sha256.
    Valid(BTreeMap<String, String>),
    Unreadable,
}

impl GoManifest {
    fn state(&self) -> &'static str {
        match self {
            GoManifest::Absent => "absent",
            GoManifest::Valid(_) => "valid",
            GoManifest::Unreadable => "unreadable",
        }
    }
    fn proves(&self, path: &str, sha: &str) -> bool {
        matches!(self, GoManifest::Valid(m) if m.get(path).is_some_and(|s| s == sha))
    }
    /// The keep reason when the manifest cannot prove anything more.
    fn keep_reason(&self) -> &'static str {
        match self {
            GoManifest::Unreadable => "go_manifest_unreadable",
            _ => "unrecorded",
        }
    }
}

fn sha_ok(s: &str) -> bool {
    s.len() == 64
        && s.bytes()
            .all(|b| b.is_ascii_digit() || (b'a'..=b'f').contains(&b))
}

/// A HOME-relative path with no empty, `.` or `..` part.
fn rel_ok(s: &str) -> bool {
    !s.is_empty()
        && !s.starts_with('/')
        && !s.contains('\\')
        && !s.chars().any(char::is_control)
        && s.split('/').all(|p| !p.is_empty() && p != "." && p != "..")
}

fn go_manifest(home: &Path) -> GoManifest {
    let moai = home.join(MOAI_DIR);
    match fs::symlink_metadata(&moai) {
        Err(_) => return GoManifest::Absent,
        Ok(m) if !m.is_dir() => return GoManifest::Unreadable,
        Ok(_) => {}
    }
    if !exists(&moai.join(USER_ASSETS)) {
        return GoManifest::Absent;
    }
    let Some(v) = read_limited(&moai, &[USER_ASSETS], USER_ASSETS_LIMIT)
        .and_then(|b| serde_json::from_slice::<Value>(&b).ok())
    else {
        return GoManifest::Unreadable;
    };
    let (true, Some(files)) = (v["schema_version"].is_u64(), v["files"].as_object()) else {
        return GoManifest::Unreadable;
    };
    let mut map = BTreeMap::new();
    for (key, entry) in files {
        let Some((slug, rel)) = key.split_once('/') else {
            continue;
        };
        let Some((_, root)) = GO_ROOTS.iter().find(|(s, _)| *s == slug) else {
            continue;
        };
        let Some(sha) = entry["sha256"].as_str().filter(|s| sha_ok(s)) else {
            continue;
        };
        if rel_ok(rel) {
            map.insert(format!("{root}/{rel}"), sha.to_string());
        }
    }
    GoManifest::Valid(map)
}

/// `Some(reason)` while a Go session may be live: the lock is held, a registered pid is alive,
/// or either cannot be told.
fn in_use(home: &Path) -> Option<&'static str> {
    let moai = home.join(MOAI_DIR);
    let lock = crate::home_state::session_lock_held(&moai);
    let live = crate::home_state::live_sessions(&moai);
    if lock == Some(true) || live.is_some_and(|n| n > 0) {
        Some("in_use")
    } else if lock.is_none() || live.is_none() {
        Some("in_use_unknown")
    } else {
        None
    }
}

// ---- the plan ----

#[derive(Clone, Debug, PartialEq, Eq)]
struct Item {
    /// HOME-relative.
    path: String,
    class: &'static str,
    archive: bool,
    reason: &'static str,
    dir: bool,
    digest: Option<String>,
}

impl Item {
    fn keep(path: &str, class: &'static str, reason: &'static str) -> Item {
        Item {
            path: path.to_string(),
            class,
            archive: false,
            reason,
            dir: false,
            digest: None,
        }
    }

    fn to_json(&self) -> Value {
        let mut v = json!({
            "path": self.path,
            "class": self.class,
            "action": if self.archive { "archive" } else { "keep" },
            "reason": self.reason,
        });
        if let Some(d) = &self.digest {
            v[if self.dir { "tree_digest" } else { "sha256" }] = json!(d);
        }
        v
    }
}

struct Plan {
    items: Vec<Item>,
    truncated: bool,
    manifest: &'static str,
    in_use: Option<&'static str>,
    generation: &'static str,
}

impl Plan {
    fn archive(&self) -> impl Iterator<Item = &Item> {
        self.items.iter().filter(|i| i.archive)
    }

    /// `cleanup:` and the first 16 hex of the sha256 over `path\tdigest\n` of every archive
    /// item in path order; `None` when nothing is archived.
    fn token(&self) -> Option<String> {
        let mut rows: Vec<(&str, &str)> = self
            .archive()
            .map(|i| (i.path.as_str(), i.digest.as_deref().unwrap_or("")))
            .collect();
        if rows.is_empty() {
            return None;
        }
        rows.sort();
        let text: String = rows.iter().map(|(p, d)| format!("{p}\t{d}\n")).collect();
        Some(format!("cleanup:{}", &hex(text.as_bytes())[..16]))
    }
}

const PROJECTION: &str = "projection_extra";
const GO_SKILL: &str = "go_skill";
const GO_AGENT: &str = "go_agent";
const MOAI_LEFTOVER: &str = "moai_leftover";

fn projection_items(
    home: &Path,
    staged: &StagedPackage,
    manifest: &GoManifest,
    busy: Option<&'static str>,
    limits: &Limits,
    items: &mut Vec<Item>,
) {
    let install_id = crate::update::current_install_id(home);
    for host in crate::update::HOSTS {
        let root = crate::update::projection_root(home, host);
        let Some(rel_root) = root
            .strip_prefix(home)
            .ok()
            .and_then(|r| r.to_str())
            .map(str::to_string)
        else {
            continue;
        };
        match real_dirs(home, &rel_root) {
            Ok(true) => {}
            Ok(false) => continue,
            Err(r) => {
                items.push(Item::keep(&rel_root, PROJECTION, r));
                continue;
            }
        }
        let owned: std::result::Result<BTreeMap<String, String>, &'static str> = match install_id
            .as_ref()
            .map_err(|_| "receipt_unavailable")
            .and_then(|id| {
                crate::update::resolve_host_receipt(home, id, host)
                    .map_err(|_| "receipt_unavailable")
            }) {
            Err(r) => Err(r),
            Ok(s) if s.blocked.is_some() => Err("receipt_unavailable"),
            Ok(s) if s.receipt.as_ref().is_some_and(|r| r.status == "pending") => {
                Err("receipt_pending")
            }
            Ok(s) => Ok(crate::update::owned_map(s.receipt.as_ref())
                .into_iter()
                .map(|(k, v)| {
                    let bare = v.strip_prefix("sha256:").unwrap_or(&v).to_string();
                    (k, bare)
                })
                .collect()),
        };
        let seen = match list(&root, limits) {
            Ok(s) => s,
            Err(r) => {
                items.push(Item::keep(&rel_root, PROJECTION, r));
                continue;
            }
        };
        for s in seen {
            if staged.files.contains_key(s.rel()) {
                continue;
            }
            let path = format!("{rel_root}/{}", s.rel());
            let rel = match s {
                Seen::Dir(_) => continue,
                Seen::Secret(_) => {
                    items.push(Item::keep(&path, PROJECTION, "secret"));
                    continue;
                }
                Seen::Link(_) => {
                    items.push(Item::keep(&path, PROJECTION, "symlink"));
                    continue;
                }
                Seen::Special(_) => {
                    items.push(Item::keep(&path, PROJECTION, "special_file"));
                    continue;
                }
                Seen::File(rel) => rel,
            };
            let sha = match file_sha(&root.join(&rel), limits) {
                Ok(s) => s,
                Err(r) => {
                    items.push(Item::keep(&path, PROJECTION, r));
                    continue;
                }
            };
            let owned_sha = owned.as_ref().ok().and_then(|m| m.get(&rel));
            let (archive, reason) = if owned_sha == Some(&sha) {
                (true, "owned_unchanged")
            } else if manifest.proves(&path, &sha) {
                match busy {
                    Some(r) => (false, r),
                    None => (true, "go_manifest"),
                }
            } else if owned_sha.is_some() {
                (false, "user_edited")
            } else if let Err(r) = &owned {
                (false, *r)
            } else {
                (false, manifest.keep_reason())
            };
            items.push(Item {
                path,
                class: PROJECTION,
                archive,
                reason,
                dir: false,
                digest: Some(sha),
            });
        }
    }
}

fn sibling_item(
    home: &Path,
    path: &str,
    manifest: &GoManifest,
    busy: Option<&'static str>,
    limits: &Limits,
) -> Item {
    let p = home.join(path);
    let Ok(m) = fs::symlink_metadata(&p) else {
        return Item::keep(path, GO_SKILL, "unreadable");
    };
    if m.file_type().is_symlink() {
        return Item::keep(path, GO_SKILL, "symlink");
    }
    let (dir, digest, files) = if m.is_dir() {
        match tree(&p, limits) {
            Ok((d, files)) => (
                true,
                d,
                files
                    .into_iter()
                    .map(|(r, s)| (format!("{path}/{r}"), s))
                    .collect::<Vec<_>>(),
            ),
            Err(r) => return Item::keep(path, GO_SKILL, r),
        }
    } else if m.is_file() {
        match file_sha(&p, limits) {
            Ok(s) => (false, s.clone(), vec![(path.to_string(), s)]),
            Err(r) => return Item::keep(path, GO_SKILL, r),
        }
    } else {
        return Item::keep(path, GO_SKILL, "special_file");
    };
    let proven = !files.is_empty() && files.iter().all(|(f, s)| manifest.proves(f, s));
    let (archive, reason) = match (proven, busy) {
        (true, None) => (true, "go_manifest"),
        (true, Some(r)) => (false, r),
        (false, _) if matches!(manifest, GoManifest::Valid(_)) && !files.is_empty() => {
            (false, "unproven_files")
        }
        (false, _) => (false, manifest.keep_reason()),
    };
    Item {
        path: path.to_string(),
        class: GO_SKILL,
        archive,
        reason,
        dir,
        digest: Some(digest),
    }
}

fn go_items(
    home: &Path,
    manifest: &GoManifest,
    busy: Option<&'static str>,
    limits: &Limits,
    items: &mut Vec<Item>,
) {
    for root in SIBLING_ROOTS {
        match real_dirs(home, root) {
            Ok(true) => {}
            Ok(false) => continue,
            Err(r) => {
                items.push(Item::keep(root, GO_SKILL, r));
                continue;
            }
        }
        for name in names(&home.join(root)) {
            if name.starts_with("moai-") {
                let path = format!("{root}/{name}");
                items.push(sibling_item(home, &path, manifest, busy, limits));
            }
        }
    }
    let GoManifest::Valid(map) = manifest else {
        return;
    };
    for root in AGENT_ROOTS {
        match real_dirs(home, root) {
            Ok(true) => {}
            Ok(false) => continue,
            Err(r) => {
                items.push(Item::keep(root, GO_AGENT, r));
                continue;
            }
        }
        for name in names(&home.join(root)) {
            let path = format!("{root}/{name}");
            let Some(want) = map.get(&path) else {
                continue;
            };
            if secret_name(&name) {
                items.push(Item::keep(&path, GO_AGENT, "secret"));
                continue;
            }
            let sha = match file_sha(&home.join(&path), limits) {
                Ok(s) => s,
                Err(r) => {
                    items.push(Item::keep(&path, GO_AGENT, r));
                    continue;
                }
            };
            let (archive, reason) = match (sha == *want, busy) {
                (true, None) => (true, "go_manifest"),
                (true, Some(r)) => (false, r),
                (false, _) => (false, "user_edited"),
            };
            items.push(Item {
                path,
                class: GO_AGENT,
                archive,
                reason,
                dir: false,
                digest: Some(sha),
            });
        }
    }
}

fn leftover_name(name: &str) -> bool {
    let n = name.to_ascii_lowercase();
    n.starts_with("tmp-") || n == "scratch" || n == "cache"
}

fn moai_items(
    home: &Path,
    generation: &str,
    busy: Option<&'static str>,
    limits: &Limits,
    items: &mut Vec<Item>,
) {
    if generation != "go" || real_dirs(home, MOAI_DIR) != Ok(true) {
        return;
    }
    for name in names(&home.join(MOAI_DIR)) {
        let path = format!("{MOAI_DIR}/{name}");
        if secret_name(&name) {
            items.push(Item::keep(&path, MOAI_LEFTOVER, "secret"));
            continue;
        }
        if GO_RUNTIME.iter().any(|g| g.eq_ignore_ascii_case(&name)) {
            items.push(Item::keep(&path, MOAI_LEFTOVER, "go_runtime"));
            continue;
        }
        let (class, reason) = table_class(&name);
        if class != EntryClass::MoveToBackup {
            items.push(Item::keep(&path, MOAI_LEFTOVER, reason));
            continue;
        }
        let p = home.join(&path);
        let Ok(m) = fs::symlink_metadata(&p) else {
            items.push(Item::keep(&path, MOAI_LEFTOVER, "unreadable"));
            continue;
        };
        if m.file_type().is_symlink() {
            items.push(Item::keep(&path, MOAI_LEFTOVER, "symlink"));
            continue;
        }
        let empty_dir = m.is_dir()
            && fs::read_dir(&p)
                .map(|mut rd| rd.next().is_none())
                .unwrap_or(false);
        if !(leftover_name(&name) || empty_dir) {
            items.push(Item::keep(&path, MOAI_LEFTOVER, "not_leftover"));
            continue;
        }
        let digest = match digest_of(&p, m.is_dir(), limits) {
            Ok(d) => d,
            Err(r) => {
                items.push(Item::keep(&path, MOAI_LEFTOVER, r));
                continue;
            }
        };
        let (archive, reason) = match busy {
            Some(r) => (false, r),
            None => (true, "go_leftover"),
        };
        items.push(Item {
            path,
            class: MOAI_LEFTOVER,
            archive,
            reason,
            dir: m.is_dir(),
            digest: Some(digest),
        });
    }
}

fn canonical_home(home: &Path) -> Result<PathBuf> {
    if !home.is_absolute() {
        return Err("HOME is not an absolute path".into());
    }
    fs::canonicalize(home).map_err(|_| "home directory unavailable".into())
}

fn compute(home: &Path, staged: &StagedPackage, limits: &Limits) -> Plan {
    let manifest = go_manifest(home);
    let busy = in_use(home);
    let generation = classify_with(home, crate::home_state::home_go_markers).generation;
    let mut items = Vec::new();
    projection_items(home, staged, &manifest, busy, limits, &mut items);
    go_items(home, &manifest, busy, limits, &mut items);
    moai_items(home, generation, busy, limits, &mut items);
    cap(&mut items, limits.items);
    let truncated = items
        .iter()
        .any(|i| i.reason == "over_limit" && i.class == "all");
    items.retain(|i| i.class != "all");
    Plan {
        items,
        truncated,
        manifest: manifest.state(),
        in_use: busy,
        generation,
    }
}

/// More than `max` items: the list is cut to `max` and nothing is archived (fail-closed). A
/// marker item of class `all` says so and is taken out by the caller.
fn cap(items: &mut Vec<Item>, max: usize) {
    if items.len() <= max {
        return;
    }
    items.truncate(max);
    for i in items.iter_mut() {
        i.archive = false;
        i.reason = "over_limit";
    }
    items.push(Item::keep("", "all", "over_limit"));
}

fn plan_json(home: &Path, staged: &StagedPackage, p: &Plan) -> Value {
    let archive = p.archive().count();
    let token = p.token();
    json!({
        "schema_version": 1,
        "home": home.display().to_string(),
        "package_digest": staged.digest,
        "generation": p.generation,
        "go_manifest": p.manifest,
        "in_use": p.in_use,
        "items": p.items.iter().map(Item::to_json).collect::<Vec<_>>(),
        "counts": {"archive": archive, "keep": p.items.len() - archive, "total": p.items.len()},
        "truncated": p.truncated,
        "archive_root": crate::home_root::path(home).join(ARCHIVE_DIR).join("<YYYYMMDDTHHMMSSZ>").display().to_string(),
        "apply_flag": token.as_ref().map(|t| format!("--confirm-cleanup {t}")),
        "confirm_token": token,
        "writes": false,
    })
}

/// The read-only preview (see the module comment).
pub fn plan(home: &Path, staged: &StagedPackage) -> Result<Value> {
    let home = canonical_home(home)?;
    let p = compute(&home, staged, &LIMITS);
    Ok(plan_json(&home, staged, &p))
}

// ---- apply ----

#[cfg(unix)]
fn device(p: &Path) -> Option<u64> {
    use std::os::unix::fs::MetadataExt;
    fs::symlink_metadata(p).ok().map(|m| m.dev())
}

/// Without device ids the same-filesystem rule cannot be checked: the apply refuses.
#[cfg(not(unix))]
fn device(_p: &Path) -> Option<u64> {
    None
}

/// The device of the archive base: its nearest existing ancestor (test builds can pretend
/// another file system, so the refusal is pinned by a test).
fn archive_device(base: &Path) -> Option<u64> {
    let mut p = base.to_path_buf();
    while !exists(&p) {
        p = p.parent()?.to_path_buf();
    }
    let dev = device(&p)?;
    #[cfg(debug_assertions)]
    if std::env::var_os("MOAI_CLI_TEST_CLEANUP_DEVICE").is_some_and(|v| v == "other") {
        return Some(dev.wrapping_add(1));
    }
    Some(dev)
}

/// temp (create_new) + write + fsync + rename + parent fsync.
fn write_atomic(path: &Path, bytes: &[u8]) -> Result<()> {
    let tmp = path.with_extension("tmp-write");
    let mut f = fs::OpenOptions::new()
        .write(true)
        .create_new(true)
        .open(&tmp)
        .map_err(|e| format!("{}: {e}", tmp.display()))?;
    let written = f.write_all(bytes).and_then(|()| f.sync_all());
    if let Err(e) = written {
        let _ = fs::remove_file(&tmp);
        return Err(format!("{}: {e}", tmp.display()));
    }
    fs::rename(&tmp, path).map_err(|e| format!("{}: {e}", path.display()))?;
    if let Some(dir) = path.parent().and_then(|d| fs::File::open(d).ok()) {
        let _ = dir.sync_all();
    }
    Ok(())
}

fn stamp_ok(s: &str) -> bool {
    let (base, n) = match s.split_once('.') {
        Some((b, n)) => (b, Some(n)),
        None => (s, None),
    };
    let b = base.as_bytes();
    let base_ok = b.len() == 16
        && b[8] == b'T'
        && b[15] == b'Z'
        && b[..8].iter().chain(&b[9..15]).all(u8::is_ascii_digit);
    let n_ok = n.is_none_or(|n| {
        n.parse::<usize>()
            .is_ok_and(|v| (2..=STAMP_TRIES).contains(&v) && n == v.to_string())
    });
    base_ok && n_ok
}

/// Creates `<base>/<stamp>`, or `<stamp>.2` … when taken; never reuses a folder.
fn create_archive(base: &Path, stamp: &str) -> Result<(PathBuf, String)> {
    for n in 1..=STAMP_TRIES {
        let name = if n == 1 {
            stamp.to_string()
        } else {
            format!("{stamp}.{n}")
        };
        let dir = base.join(&name);
        match fs::create_dir(&dir) {
            Ok(()) => return Ok((dir, name)),
            Err(e) if e.kind() == std::io::ErrorKind::AlreadyExists => continue,
            Err(e) => return Err(format!("{}: {e}; nothing was moved", dir.display())),
        }
    }
    Err(format!(
        "{STAMP_TRIES} archive folders for {stamp} exist; nothing was moved"
    ))
}

/// The archive base `<state>/cleanup`; a link on the way is refused.
fn archive_base(home: &Path) -> Result<PathBuf> {
    let state = crate::home_root::path(home);
    for p in [&state, &state.join(ARCHIVE_DIR)] {
        if let Ok(m) = fs::symlink_metadata(p) {
            if !m.is_dir() || m.file_type().is_symlink() {
                return Err(format!(
                    "{} is a link or not a folder; nothing was moved",
                    p.display()
                ));
            }
        }
    }
    Ok(state.join(ARCHIVE_DIR))
}

fn receipt_bytes(r: &Value) -> Vec<u8> {
    format!("{r:#}\n").into_bytes()
}

/// Move every archive item of a fresh plan into a new archive folder (see the module comment).
pub fn apply(
    home: &Path,
    staged: &StagedPackage,
    confirm: &str,
    permission_reference: &str,
    now: &str,
) -> Result<Value> {
    if !crate::contract::permission_reference_ok(permission_reference) {
        return Err("permission reference required".into());
    }
    let stamp = crate::home_state::stamp(now)?;
    let home = canonical_home(home)?;
    let p = compute(&home, staged, &LIMITS);
    let Some(token) = p.token() else {
        return Err("the plan archives nothing; nothing was moved".into());
    };
    if token != confirm {
        return Err(
            "the confirm token differs from a fresh plan's; re-run cleanup-plan. Nothing was moved"
                .into(),
        );
    }
    let base = archive_base(&home)?;
    let dev = archive_device(&base)
        .ok_or("the archive's file system cannot be told; nothing was moved")?;
    for i in p.archive() {
        let parent = home.join(&i.path);
        let parent = parent.parent().ok_or("item path invalid")?;
        if device(parent) != Some(dev) {
            return Err(format!(
                "{} is on another file system than {}; nothing was moved",
                i.path,
                base.display()
            ));
        }
    }
    fs::create_dir_all(&base).map_err(|e| format!("{}: {e}; nothing was moved", base.display()))?;
    let (root, name) = create_archive(&base, &stamp)?;

    let mut manifest = String::new();
    let mut rows = Vec::new();
    for i in p.archive() {
        let original = home.join(&i.path);
        let archived = root.join(&i.path);
        manifest.push_str(&format!("{}\t{}\n", original.display(), archived.display()));
        rows.push(json!({
            "path": i.path,
            "class": i.class,
            "kind": if i.dir { "dir" } else { "file" },
            "digest": i.digest,
            "original": original.display().to_string(),
            "archived": archived.display().to_string(),
            "moved": false,
        }));
    }
    write_atomic(&root.join(MANIFEST), manifest.as_bytes())?;
    let receipt_path = root.join(RECEIPT);
    let mut receipt = json!({
        "schema_version": 1,
        "operation": "cleanup",
        "home": home.display().to_string(),
        "stamp": name,
        "archive_root": root.display().to_string(),
        "status": "pending",
        "created_at": now,
        "finished_at": null,
        "permission_reference": permission_reference,
        "confirm_token": token,
        "items": rows,
    });
    write_atomic(&receipt_path, &receipt_bytes(&receipt))?;

    let todo: Vec<Item> = p.archive().cloned().collect();
    for (n, i) in todo.iter().enumerate() {
        let original = home.join(&i.path);
        let archived = root.join(&i.path);
        let now_digest = digest_of(&original, i.dir, &LIMITS).ok();
        let parent_real = parent_rel(&i.path).is_none_or(|r| real_dirs(&home, r) == Ok(true));
        if !parent_real || now_digest.is_none() || now_digest != i.digest {
            receipt["items"][n]["changed"] = json!(true);
            write_atomic(&receipt_path, &receipt_bytes(&receipt))?;
            return Err(format!(
                "{} changed since the plan; stopped after {n} moved, the receipt {} stays pending",
                i.path,
                receipt_path.display()
            ));
        }
        let moved = archived
            .parent()
            .ok_or_else(|| "archive path invalid".to_string())
            .and_then(|d| fs::create_dir_all(d).map_err(|e| e.to_string()))
            .and_then(|()| fs::rename(&original, &archived).map_err(|e| e.to_string()));
        if let Err(e) = moved {
            receipt["items"][n]["error"] = json!(e);
            write_atomic(&receipt_path, &receipt_bytes(&receipt))?;
            return Err(format!(
                "moving {} failed: {e}; stopped after {n} moved, the receipt {} stays pending",
                i.path,
                receipt_path.display()
            ));
        }
        receipt["items"][n]["moved"] = json!(true);
        write_atomic(&receipt_path, &receipt_bytes(&receipt))?;
    }
    receipt["status"] = json!("committed");
    receipt["finished_at"] = json!(now);
    write_atomic(&receipt_path, &receipt_bytes(&receipt))?;
    Ok(json!({
        "schema_version": 1,
        "state": "committed",
        "home": home.display().to_string(),
        "stamp": name,
        "archive_root": root.display().to_string(),
        "moved": todo.len(),
        "manifest": root.join(MANIFEST).display().to_string(),
        "receipt": receipt_path.display().to_string(),
        "revert": format!("cleanup-revert-plan {} --stamp {name}", home.display()),
    }))
}

// ---- revert ----

struct RevertItem {
    path: String,
    dir: bool,
    digest: String,
    status: &'static str,
}

fn load_receipt(home: &Path, stamp: &str) -> Result<(PathBuf, Value)> {
    if !stamp_ok(stamp) {
        return Err(format!(
            "stamp `{stamp}` is not YYYYMMDDTHHMMSSZ or YYYYMMDDTHHMMSSZ.N"
        ));
    }
    let state = crate::home_root::path(home);
    match fs::symlink_metadata(&state) {
        Ok(m) if m.is_dir() => {}
        _ => return Err(format!("{} is not a folder", state.display())),
    }
    let bytes =
        read_limited(&state, &[ARCHIVE_DIR, stamp, RECEIPT], RECEIPT_LIMIT).ok_or_else(|| {
            format!(
                "no readable cleanup receipt for {stamp} (a link or an oversized file is refused)"
            )
        })?;
    let v: Value = serde_json::from_slice(&bytes)
        .map_err(|_| "the cleanup receipt is not JSON".to_string())?;
    if v["operation"] != "cleanup" || !v["items"].is_array() {
        return Err("the file is not a cleanup receipt".into());
    }
    Ok((state.join(ARCHIVE_DIR).join(stamp), v))
}

fn revert_items(home: &Path, root: &Path, receipt: &Value) -> Result<Vec<RevertItem>> {
    let mut out = Vec::new();
    for i in receipt["items"].as_array().into_iter().flatten() {
        let (Some(path), Some(kind), Some(digest)) =
            (i["path"].as_str(), i["kind"].as_str(), i["digest"].as_str())
        else {
            return Err("a receipt item lacks path, kind or digest".into());
        };
        if !rel_ok(path) || !matches!(kind, "file" | "dir") {
            return Err(format!("receipt item `{path}` is not a HOME-relative path"));
        }
        let dir = kind == "dir";
        let original = home.join(path);
        let archived = root.join(path);
        let same = |p: &Path| digest_of(p, dir, &LIMITS).ok().as_deref() == Some(digest);
        // A parent swapped for a link since the apply would carry the restore outside HOME.
        let parent_linked = parent_rel(path).is_some_and(|r| real_dirs(home, r).is_err());
        let status = if parent_linked {
            "conflict"
        } else if exists(&original) {
            if !exists(&archived) && same(&original) {
                "done"
            } else {
                "conflict"
            }
        } else if exists(&archived) && same(&archived) {
            "restore"
        } else {
            "missing"
        };
        out.push(RevertItem {
            path: path.to_string(),
            dir,
            digest: digest.to_string(),
            status,
        });
    }
    Ok(out)
}

fn revert_json(home: &Path, stamp: &str, receipt: &Value, items: &[RevertItem]) -> Value {
    let count = |s: &str| items.iter().filter(|i| i.status == s).count();
    let blocked = count("conflict") + count("missing") > 0;
    json!({
        "schema_version": 1,
        "home": home.display().to_string(),
        "stamp": stamp,
        "receipt_status": receipt["status"],
        "items": items.iter().map(|i| json!({
            "path": i.path,
            "kind": if i.dir { "dir" } else { "file" },
            "digest": i.digest,
            "status": i.status,
        })).collect::<Vec<_>>(),
        "counts": {
            "restore": count("restore"),
            "done": count("done"),
            "conflict": count("conflict"),
            "missing": count("missing"),
        },
        "action": if blocked { "blocked" } else if count("restore") > 0 { "restore" } else { "none" },
    })
}

/// The read-only revert preview of the apply `stamp`.
pub fn revert_plan(home: &Path, stamp: &str) -> Result<Value> {
    let home = canonical_home(home)?;
    let (root, receipt) = load_receipt(&home, stamp)?;
    let items = revert_items(&home, &root, &receipt)?;
    let mut v = revert_json(&home, stamp, &receipt, &items);
    v["writes"] = json!(false);
    Ok(v)
}

/// Move every `restore` item back; refuses as a whole on any `conflict` or `missing`. Safe to
/// rerun: restored items read `done`.
pub fn revert(home: &Path, stamp: &str, permission_reference: &str) -> Result<Value> {
    if !crate::contract::permission_reference_ok(permission_reference) {
        return Err("permission reference required".into());
    }
    let home = canonical_home(home)?;
    let (root, mut receipt) = load_receipt(&home, stamp)?;
    let items = revert_items(&home, &root, &receipt)?;
    let bad: Vec<String> = items
        .iter()
        .filter(|i| matches!(i.status, "conflict" | "missing"))
        .map(|i| format!("{} ({})", i.path, i.status))
        .collect();
    if !bad.is_empty() {
        return Err(format!(
            "the revert is refused as a whole: {}; nothing was moved",
            bad.join(", ")
        ));
    }
    let mut restored = 0usize;
    for i in items.iter().filter(|i| i.status == "restore") {
        let original = home.join(&i.path);
        if let Some(d) = original.parent() {
            fs::create_dir_all(d).map_err(|e| {
                format!("{}: {e}; restored {restored} before stopping", d.display())
            })?;
        }
        fs::rename(root.join(&i.path), &original).map_err(|e| {
            format!(
                "restoring {} failed: {e}; restored {restored} before stopping (rerun is safe)",
                i.path
            )
        })?;
        restored += 1;
    }
    if receipt["status"] != "reverted" {
        receipt["status"] = json!("reverted");
        write_atomic(&root.join(RECEIPT), &receipt_bytes(&receipt))?;
    }
    let after = revert_items(&home, &root, &receipt)?;
    let mut v = revert_json(&home, stamp, &receipt, &after);
    v["state"] = json!("reverted");
    v["restored"] = json!(restored);
    Ok(v)
}

#[cfg(test)]
mod tests {
    use super::*;

    fn item(path: &str, digest: &str) -> Item {
        Item {
            path: path.into(),
            class: GO_SKILL,
            archive: true,
            reason: "go_manifest",
            dir: false,
            digest: Some(digest.into()),
        }
    }

    fn plan_of(items: Vec<Item>) -> Plan {
        Plan {
            items,
            truncated: false,
            manifest: "valid",
            in_use: None,
            generation: "go",
        }
    }

    #[test]
    fn the_token_binds_paths_and_digests_in_any_order() {
        let a = plan_of(vec![item("a", "1"), item("b", "2")]);
        let b = plan_of(vec![item("b", "2"), item("a", "1")]);
        let t = a.token().unwrap();
        assert_eq!(Some(t.clone()), b.token());
        assert!(
            t.starts_with("cleanup:") && t.len() == "cleanup:".len() + 16,
            "{t}"
        );
        assert_ne!(
            Some(t.clone()),
            plan_of(vec![item("a", "1"), item("b", "3")]).token()
        );
        assert_ne!(
            Some(t),
            plan_of(vec![item("a", "1"), item("c", "2")]).token()
        );
        let mut keep = item("a", "1");
        keep.archive = false;
        assert_eq!(plan_of(vec![keep]).token(), None);
        assert_eq!(plan_of(Vec::new()).token(), None);
    }

    #[test]
    fn the_item_cap_keeps_everything_past_n() {
        let make = |n: usize| {
            (0..n)
                .map(|i| item(&format!("p{i}"), "d"))
                .collect::<Vec<_>>()
        };
        for (n, max) in [(0, 0), (3, 3), (4, 3)] {
            let mut items = make(n);
            cap(&mut items, max);
            let over = n > max;
            assert_eq!(items.iter().any(|i| i.class == "all"), over, "{n}/{max}");
            items.retain(|i| i.class != "all");
            assert_eq!(items.len(), n.min(max));
            assert_eq!(items.iter().all(|i| i.archive), !over, "{n}/{max}");
        }
    }

    fn limits(entries: usize, bytes: u64, depth: usize) -> Limits {
        Limits {
            items: 10,
            tree_entries: entries,
            tree_bytes: bytes,
            depth,
        }
    }

    #[test]
    fn tree_walk_limits_hold_at_n_and_refuse_n_plus_1() {
        let d = tempfile::tempdir().unwrap();
        let t = d.path().join("t");
        fs::create_dir_all(t.join("sub")).unwrap();
        fs::write(t.join("a"), b"12345").unwrap();
        fs::write(t.join("sub/b"), b"67890").unwrap();
        // Three entries (a, sub, sub/b), ten bytes, depth 1.
        assert!(tree(&t, &limits(3, 10, 1)).is_ok());
        assert_eq!(tree(&t, &limits(2, 10, 1)), Err("over_limit"));
        assert_eq!(tree(&t, &limits(3, 9, 1)), Err("over_limit"));
        assert_eq!(tree(&t, &limits(3, 10, 0)), Err("over_limit"));
        // A file alone: its size against the byte bound.
        assert!(file_sha(&t.join("a"), &limits(1, 5, 0)).is_ok());
        assert_eq!(file_sha(&t.join("a"), &limits(1, 4, 0)), Err("over_limit"));
        // Empty folder: a digest over nothing.
        let e = d.path().join("e");
        fs::create_dir(&e).unwrap();
        assert_eq!(tree(&e, &limits(0, 0, 0)).unwrap().0, hex(b""));
    }

    #[test]
    fn a_tree_with_a_link_special_file_or_secret_is_refused_unread() {
        let d = tempfile::tempdir().unwrap();
        let t = d.path().join("t");
        fs::create_dir_all(&t).unwrap();
        fs::write(t.join("a"), b"x").unwrap();
        let ok = tree(&t, &LIMITS).unwrap();
        fs::write(t.join(".env"), b"SECRET").unwrap();
        #[cfg(unix)]
        fs::set_permissions(
            t.join(".env"),
            std::os::unix::fs::PermissionsExt::from_mode(0o000),
        )
        .unwrap();
        assert_eq!(tree(&t, &LIMITS), Err("secret"));
        fs::remove_file(t.join(".env")).unwrap();
        assert_eq!(tree(&t, &LIMITS).unwrap(), ok);
        #[cfg(unix)]
        {
            std::os::unix::fs::symlink("/etc", t.join("l")).unwrap();
            assert_eq!(tree(&t, &LIMITS), Err("symlink"));
        }
    }

    #[test]
    fn stamps_are_checked() {
        for ok in [
            "20261114T010203Z",
            "20261114T010203Z.2",
            "20261114T010203Z.64",
        ] {
            assert!(stamp_ok(ok), "{ok}");
        }
        for bad in [
            "20261114T010203Z.1",
            "20261114T010203Z.65",
            "20261114T010203Z.02",
            "20261114T010203",
            "../x",
            "2026111XT010203Z",
            "",
        ] {
            assert!(!stamp_ok(bad), "{bad}");
        }
    }

    #[test]
    fn manifest_paths_and_hashes_are_checked() {
        assert!(rel_ok(".claude/skills/moai-x/a.md"));
        for bad in ["", "/a", "a//b", "a/../b", "./a", "a\\b"] {
            assert!(!rel_ok(bad), "{bad}");
        }
        assert!(sha_ok(&"a".repeat(64)));
        assert!(!sha_ok(&"A".repeat(64)));
        assert!(!sha_ok(&"a".repeat(63)));
        assert!(!sha_ok(&format!("sha256:{}", "a".repeat(57))));
    }
}
