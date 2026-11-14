//! #2561: Go-generation project state — the read-only preview (`go-state-plan`).
//!
//! The Go tool records what it deployed into a project in `.moai/manifest.json`
//! (`internal/manifest/types.go`: provenance `template_managed` means "deployed from a
//! template and not modified by the user"; hashes are `sha256:<hex>`). That file is read
//! as untrusted input. A path is offered for moving only when every check holds:
//! a plain relative key (no root, no `..`, no backslash), under `.moai/` or `.claude/`,
//! no path component is a symlink, the path is a regular file, the manifest records it as
//! `template_managed`, its bytes still hash to the recorded `deployed_hash`, and
//! neither `.claude/settings.json` nor `.claude/settings.local.json` references it, and it
//! is not one of those two live settings files. Anything else stays, with its reason;
//! files the manifest does not record are never offered and are counted per folder.
//! Without a usable manifest nothing moves. Which of these rules apply to `.moai/` as a
//! whole is a pending decision (#2561 comment 6827973256); [`policy`] is the one place
//! that choice lives. The preview writes nothing.
//!
//! #2696 (#2725 S2c, user RC2 decision via #2725 rev2 6859340307): when the Go `.moai/` moves,
//! it is kept whole under a new name — `.moai-backup`, or `.moai-backup-<UTC>` when that name
//! is taken, never overwritten — and a new `.moai/` becomes this generation's state folder:
//! the earlier `.moai-cli/` product state moves into it, SPECs are copied from the backup,
//! `CLAUDE.md`/`AGENTS.md`/`AGENTS.local.md` are moved into the backup's `root-files/` (#2936
//! O4; the local settings are copied there, 0600), and the marker
//! `.moai/moai-cli.json` is written last. The stage-2 section below has the layout.

use serde::{Deserialize, Serialize};
// Stage 2 (`go-state-apply`, `go-state-revert-plan`, `go-state-revert`) is below `plan`.
use sha2::{Digest, Sha256};
use std::{
    collections::{BTreeMap, HashMap},
    fs,
    path::{Component, Path, PathBuf},
};

type Result<T> = std::result::Result<T, String>;

const SCHEMA_VERSION: u32 = 1;
const MANIFEST: &str = ".moai/manifest.json";
const SETTINGS: &str = ".claude/settings.json";
const SETTINGS_LOCAL: &str = ".claude/settings.local.json";
/// The Go manifest of this checkout is about 180 KiB; this only refuses absurd inputs.
const MANIFEST_LIMIT: u64 = 16 * 1024 * 1024;
const FILE_LIMIT: u64 = 64 * 1024 * 1024;
/// #2936 S1b: `.agents/` holds the Go skill mirror and the published command skills.
const ROOTS: [&str; 3] = [".moai", ".claude", ".agents"];
/// The Go skill mirror (`internal/template/skill_mirror.go`): `.agents/skills/<n>` is a link
/// whose body is exactly `../../.claude/skills/<n>`, or a real copy of that folder. Mirror
/// entries are never in the Go manifest; the canonical files are.
const AGENTS_SKILLS: &str = ".agents/skills";
const CANONICAL_SKILLS: &str = ".claude/skills";
/// #2936 O1: the Go `.moai/` folders that hold the user's own data, restored from the backup.
const USER_DATA_FROM: [&str; 10] = [
    "specs",
    "project",
    "plans",
    "reports",
    "docs",
    "research",
    "memory",
    "proposals",
    "todo",
    "release-notes",
];
/// This generation's managed section (`agents::markers`).
const RUST_MARKER_TEXT: &str = "<!-- moai:managed:begin";
/// The Go curator's block markers (`internal/harness/curator/writer.go`).
const GO_MARKERS: [&str; 3] = [
    "<!-- moai:learned-start",
    "<!-- moai:harness-start",
    "<!-- moai:learned-local-start",
];
/// Files a Go skill-mirror copy folder may hold before the preview stops checking it. Debug
/// builds take `MOAI_CLI_TEST_MIRROR_FILE_LIMIT` so a test can reach it with a few files.
fn mirror_file_limit() -> u64 {
    #[cfg(debug_assertions)]
    if let Some(n) = std::env::var("MOAI_CLI_TEST_MIRROR_FILE_LIMIT")
        .ok()
        .and_then(|v| v.parse().ok())
    {
        return n;
    }
    10_000
}

/// #2936 O6: a name that holds secrets by convention. Such an entry under `.claude/` or
/// `.agents/` is never read, hashed, copied or moved; the preview only counts it.
fn secret_name(name: &str) -> bool {
    crate::home_state::is_secret(name) || crate::worktree_include::secret_name(name)
}

fn has_secret_component(rel: &str) -> bool {
    Path::new(rel)
        .components()
        .any(|c| c.as_os_str().to_str().is_some_and(secret_name))
}
/// Decision 1 (#2561 comment 6831345102): `.moai/` moves as a whole folder.
const MOAI_DIR: &str = ".moai";
/// Entries the `unrecorded` count reads before it stops (#2561 F1: a project's `.moai/` held
/// unpacked review trees with `node_modules`, and the preview read for over 14 minutes). The
/// count only informs; nothing moves on its account. Debug builds take
/// `MOAI_CLI_TEST_UNRECORDED_LIMIT` so a test can reach the limit with a few files.
pub(crate) fn unrecorded_limit() -> u64 {
    #[cfg(debug_assertions)]
    if let Some(n) = std::env::var("MOAI_CLI_TEST_UNRECORDED_LIMIT")
        .ok()
        .and_then(|v| v.parse().ok())
    {
        return n;
    }
    200_000
}
/// The earlier Rust project root whose product state moves into `.moai/` (#2696).
const LEGACY_ROOT: &str = crate::PROJECT_STATE_DIR;
/// The earlier root's entries that are product state and move into `.moai/` (design
/// 6862154151). Everything else there stays: `worktrees/` until #2723 moves them (D6),
/// `reports/` (D5), `backups/` where earlier go-state receipts expect it, and anything else.
const STATE_ENTRIES: [&str; 17] = [
    "archive",
    "config",
    "decisions",
    "evidence",
    "forge",
    "imports",
    "index",
    "manifests",
    "messages",
    "project-skill-rollback",
    "project.yaml",
    "receipts",
    "settings-backups",
    "specs",
    "state",
    "statusline.local.yaml",
    "workflows",
];
/// Root instruction files that move into the backup's `root-files/` with the Go folder (#2936
/// O4); the install step writes new ones. `CLAUDE.local.md` is the user's and stays.
const ROOT_FILES: [&str; 3] = ["CLAUDE.md", "AGENTS.md", "AGENTS.local.md"];
/// The name a revert gives this generation's `.moai/` before the Go folder comes back.
const RUST_ASIDE: &str = ".moai.rust-";
/// The Go backup name when [`crate::state_root::GO_BACKUP`] is taken; the apply fills in its
/// stamp.
const GO_BACKUP_STAMPED: &str = ".moai-backup-<UTC>";

#[derive(Deserialize)]
struct Manifest {
    version: String,
    files: HashMap<String, Entry>,
}

#[derive(Deserialize)]
struct Entry {
    provenance: String,
    #[serde(default)]
    deployed_hash: String,
}

#[derive(Clone, Debug, Serialize, PartialEq, Eq, Default)]
pub struct Item {
    pub path: String,
    /// `move`, `leave` or `with_folder`.
    pub action: String,
    pub reason: String,
    /// #2936: what the path is — `go_folder`, `tool_asset`, `user_modified`, `user_data`,
    /// `go_mirror` (the mirror link), `go_mirror_copy`, `conflict`, `settings`,
    /// `local_settings`, `instructions`, `legacy_state`, `outside` or `unknown`.
    pub class: String,
    /// #2936 R-A: what the file is for under `.claude/` and `.agents/` — `hook`,
    /// `output_style`, `rule`, `command`, `workflow`, `loop`, `skill`, `agent`, `settings` or
    /// `other`; empty elsewhere. See [`kind_of`].
    #[serde(skip_serializing_if = "String::is_empty")]
    pub kind: String,
    /// What proves the class (the manifest record and hash, or the link body); may be empty.
    pub proof: String,
    /// U4: for a moving item, `modified`, `untracked`, `clean` or `unknown` (read-only
    /// `git status`).
    #[serde(skip_serializing_if = "String::is_empty")]
    pub git: String,
    /// The manifest's `deployed_hash`, kept for the receipt of an edited file (O2).
    #[serde(skip)]
    pub(crate) deployed_hash: String,
}

fn new_item(path: &str, action: &str, reason: String, class: &str) -> Item {
    Item {
        path: path.into(),
        action: action.into(),
        reason,
        class: class.into(),
        ..Item::default()
    }
}

#[derive(Clone, Debug, Serialize)]
pub struct GoStatePlan {
    pub schema_version: u32,
    pub project: String,
    pub manifest_version: Option<String>,
    pub items: Vec<Item>,
    /// Regular files under `.moai/` and `.claude/` the manifest does not record, counted per
    /// second-level folder (`.moai/specs`, `.claude/worktrees`, …). They are never moved.
    pub unrecorded: BTreeMap<String, u64>,
    /// AC-5: files copied from the Go backup's `specs/` into `.moai/specs` (`copy`), or not
    /// (`conflict`, `conflict_link`, `excluded_secret`, `excluded_unclassifiable`,
    /// `leave_link`). Empty unless the folder moves.
    pub restores: Vec<Restore>,
    /// #2696: the name the Go `.moai/` gets — `.moai-backup`, or `.moai-backup-<UTC>` when
    /// that is taken. `None` unless the folder moves.
    pub go_backup: Option<String>,
    /// #2696: the earlier root's entries, each `move` (into `.moai/`) or `leave`.
    pub state_moves: Vec<StateMove>,
    /// #2696, #2936 O4: `CLAUDE.md`, `AGENTS.md` and `AGENTS.local.md`, moved into the backup
    /// (`move`: gone from the project root until the install step writes a new one) or left
    /// (`leave`), and `.claude/settings.local.json`, copied (`copy`, the original stays).
    pub root_files: Vec<RootFile>,
    /// #2696: the Go folder's top-level entries; all of them stay in the backup.
    pub left_in_backup: Vec<String>,
    /// #2936 R-A: how many moving items each [`Item::kind`] has.
    pub kinds: BTreeMap<String, u64>,
    /// #2936 R-A: for each kind that moves, what this generation puts in its place.
    pub counterparts: Vec<Counterpart>,
    /// #2936 R-A: what the move leaves behind.
    pub after_move: AfterMove,
    /// #2936 O6: secret-named entries under `.claude/` and `.agents/`; only counted, never
    /// listed, read or moved.
    pub secrets: u64,
    pub warnings: Vec<String>,
    pub plan_digest: String,
}

/// #2936 R-A: what takes the place of one kind of moved file.
#[derive(Clone, Debug, Serialize, PartialEq, Eq)]
pub struct Counterpart {
    pub kind: String,
    /// How many files of the kind move.
    pub moved: u64,
    /// `reinstalled by <step>` (a step of `project-update-plan`), `user layer after moai
    /// update` (HOME, not read here) or `no counterpart — will be gone`.
    pub counterpart: String,
    pub note: String,
}

/// #2936 R-A: what the move leaves behind, computed from the plan alone.
#[derive(Clone, Debug, Serialize, PartialEq, Eq, Default)]
pub struct AfterMove {
    /// Folders under the project that hold nothing once the moving files are gone; a folder
    /// that keeps a script the settings name is not listed.
    pub empty_folders: Vec<String>,
    /// Include and ignore lines that select a moving path.
    pub dangling: Vec<crate::worktree_include::ListHit>,
    /// Set when the settings file moves: the hook entries and deny rules it carried.
    pub wiring_lost: Option<WiringLost>,
    /// `true` when a bounded read gave up (the folder walk passed its limit, or an include or
    /// ignore list was not read): the lists above may be incomplete.
    pub unknown: bool,
}

#[derive(Clone, Debug, Serialize, PartialEq, Eq)]
pub struct WiringLost {
    pub file: String,
    /// Go-generation hook entries in the settings.
    pub go_hook_entries: usize,
    /// `permissions.deny` rules in the settings.
    pub deny_rules: usize,
    /// The settings file's top-level keys; the install step writes only the keys MoAI owns.
    pub keys: Vec<String>,
}

#[derive(Clone, Debug, Serialize, PartialEq, Eq)]
pub struct StateMove {
    pub from: String,
    pub to: String,
    /// `move` or `leave`.
    pub action: String,
    pub reason: String,
}

#[derive(Clone, Debug, Serialize, PartialEq, Eq)]
pub struct RootFile {
    pub path: String,
    pub sha256: String,
    /// `move` (a root instruction file, #2936 O4), `copy` (`.claude/settings.local.json`,
    /// which stays in place, O3) or `leave`.
    pub action: String,
    /// Where the apply puts the file or its copy (`<UTC>` is the apply's stamp).
    pub backup_copy: String,
    /// `keep`: a later PROJECT step offers the choice; the backup stays either way.
    pub user_choice: String,
    pub reason: String,
    /// `instructions`, `rust`, `conflict`, `user_data` or `local_settings`.
    pub class: String,
    /// `manifest` (the Go manifest records it), `marker` (a Go curator marker), or `none`.
    pub go_proof: String,
    #[serde(skip_serializing_if = "String::is_empty")]
    pub git: String,
}

#[derive(Clone, Debug, Serialize, PartialEq, Eq)]
pub struct Restore {
    pub from: String,
    pub to: String,
    pub sha256: String,
    pub action: String,
    /// Why a file is not copied (`excluded_secret`, `excluded_unclassifiable`, `leave_link`);
    /// empty for `copy` and `conflict`.
    #[serde(skip_serializing_if = "String::is_empty")]
    pub reason: String,
}

/// `.json`, `.toml`, `.yaml` and `.yml` (any case): the only restore candidates the config
/// classifier of `migrate` is meant for. Every other file is a document (#2966 audit B1).
fn config_shaped(name: &str) -> bool {
    Path::new(name)
        .extension()
        .and_then(|e| e.to_str())
        .is_some_and(|e| {
            matches!(
                e.to_ascii_lowercase().as_str(),
                "json" | "toml" | "yaml" | "yml"
            )
        })
}

/// The restore candidates under the Go folder's user-data folders ([`USER_DATA_FROM`], #2936
/// O1; `.moai/specs` was the first, AC-5): regular files are copied unless the target exists
/// (never overwritten) or the name or content is or may be a secret (a secret name is
/// excluded before any read). A document (`.md`, `.txt`, anything but a config file) passes on
/// its name and size; a config-shaped file (`.json`, `.toml`, `.yaml`, `.yml`) and anything
/// under `specs/` goes through the `migrate` filter, fail-closed, and a refusal carries its
/// reason. Links, a linked user-data folder and special files are never followed or opened. The copy reads the file from the Go
/// backup after the rename. A folder the earlier root also moves into `.moai/` (only `specs/`)
/// lands there first, so a file already at `.moai-cli/<folder>/<same path>` is the conflict.
fn plan_restores(project: &Path, go_backup: &str) -> Vec<Restore> {
    let mut out = Vec::new();
    for folder in USER_DATA_FROM {
        let top = project.join(MOAI_DIR).join(folder);
        // A folder that is itself a link is never entered: the apply refuses to read through
        // it (after the Go folder was renamed), so the preview leaves it where it is.
        if fs::symlink_metadata(&top).is_ok_and(|m| m.file_type().is_symlink()) {
            out.push(Restore {
                from: format!("{go_backup}/{folder}"),
                to: format!("{MOAI_DIR}/{folder}"),
                sha256: String::new(),
                action: "leave_link".into(),
                reason: "the folder is a link; never followed, it stays in the backup".into(),
            });
            continue;
        }
        let mut stack = vec![top.clone()];
        while let Some(dir) = stack.pop() {
            let Ok(entries) = fs::read_dir(&dir) else {
                continue;
            };
            for e in entries.flatten() {
                let p = e.path();
                let Some(rel) = p
                    .strip_prefix(&top)
                    .ok()
                    .and_then(|r| r.to_str())
                    .map(|r| r.replace('\\', "/"))
                else {
                    continue;
                };
                let from = format!("{go_backup}/{folder}/{rel}");
                let to = format!("{MOAI_DIR}/{folder}/{rel}");
                if e.file_name().to_str().is_none_or(secret_name) {
                    out.push(Restore {
                        from,
                        to,
                        sha256: String::new(),
                        action: "excluded_secret".into(),
                        reason: "a secret name; never opened".into(),
                    });
                    continue;
                }
                let name = e.file_name().to_string_lossy().into_owned();
                let Ok(m) = fs::symlink_metadata(&p) else {
                    continue;
                };
                if m.is_dir() {
                    stack.push(p);
                    continue;
                }
                let lands = STATE_ENTRIES
                    .contains(&folder)
                    .then(|| format!("{LEGACY_ROOT}/{folder}/{rel}"));
                let unreadable = "larger than the file limit or not readable; not copied";
                let (sha, action, reason) = if !m.is_file() {
                    (
                        String::new(),
                        "leave_link",
                        "a link or special file; never opened or followed".to_string(),
                    )
                } else if m.len() > FILE_LIMIT {
                    (
                        String::new(),
                        "excluded_unclassifiable",
                        format!("larger than the file limit ({FILE_LIMIT} bytes); not read"),
                    )
                } else {
                    match fs::read(&p) {
                        Ok(bytes) if (bytes.len() as u64) <= FILE_LIMIT => {
                            let sha = format!("sha256:{:x}", Sha256::digest(&bytes));
                            let is_spec = folder == "specs";
                            // The config classifier is for config files. A document in a
                            // user-data folder passes on its name and size alone; SPECs keep
                            // the stricter text check they have always had.
                            let verdict = if is_spec || config_shaped(&name) {
                                crate::migrate::sensitive_content(&bytes, is_spec)
                            } else {
                                Ok(false)
                            };
                            let (action, reason) = match verdict {
                                Ok(true) => (
                                    "excluded_secret",
                                    "the content looks like a credential; not copied".to_string(),
                                ),
                                Err(why) => ("excluded_unclassifiable", why),
                                Ok(false) => (
                                    match &lands {
                                        Some(l) if linked_component(project, l) => "conflict_link",
                                        Some(l)
                                            if fs::symlink_metadata(project.join(l)).is_ok() =>
                                        {
                                            "conflict"
                                        }
                                        _ => "copy",
                                    },
                                    String::new(),
                                ),
                            };
                            (sha, action, reason)
                        }
                        _ => (
                            String::new(),
                            "excluded_unclassifiable",
                            unreadable.to_string(),
                        ),
                    }
                };
                out.push(Restore {
                    from,
                    to,
                    sha256: sha,
                    action: action.into(),
                    reason,
                });
            }
        }
    }
    out.sort_by(|a, b| a.from.cmp(&b.from));
    out
}

fn sha256_file(path: &Path) -> Result<String> {
    let meta = fs::symlink_metadata(path).map_err(|_| "file metadata unavailable")?;
    if !meta.is_file() || meta.len() > FILE_LIMIT {
        return Err("not a bounded regular file".into());
    }
    let bytes = fs::read(path).map_err(|_| "file read failed")?;
    Ok(format!("sha256:{:x}", Sha256::digest(bytes)))
}

/// A plain relative key: normal components only, no backslash, at least one component.
fn plain_relative(key: &str) -> bool {
    !key.is_empty()
        && !key.contains('\\')
        && Path::new(key)
            .components()
            .all(|c| matches!(c, Component::Normal(_)))
}

/// Whether any component from the project root down to `rel` is a symlink.
fn linked_component(project: &Path, rel: &str) -> bool {
    let mut cur = project.to_path_buf();
    for part in Path::new(rel).components() {
        cur.push(part);
        match fs::symlink_metadata(&cur) {
            Ok(m) if m.file_type().is_symlink() => return true,
            Ok(_) => {}
            Err(_) => return false,
        }
    }
    false
}

/// `.agents/skills/moai-<command>/SKILL.md`, the Go published command skills
/// (`internal/template/published_skills.go`); the Go manifest records them.
fn published_skill(key: &str) -> bool {
    let parts: Vec<&str> = key.split('/').collect();
    matches!(parts.as_slice(), [".agents", "skills", name, "SKILL.md"] if name.starts_with("moai-"))
}

/// The decision for one recorded path, as (action, reason, class, proof). The single place
/// the move rule lives.
fn policy(
    project: &Path,
    key: &str,
    entry: &Entry,
    settings: &[(&'static str, String)],
    folder_moves: bool,
) -> (&'static str, &'static str, &'static str, String) {
    let leave = |reason, class| ("leave", reason, class, String::new());
    if !plain_relative(key) {
        return leave("not a plain relative path", "outside");
    }
    let top = top_of(key);
    if top == ".codex" {
        return leave("Codex files are handled in S1c (#2936)", "outside");
    }
    if ROOT_FILES.contains(&key) {
        return leave("a root instruction file; see root_files", "instructions");
    }
    if !ROOTS.contains(&top) {
        return leave("outside .moai/, .claude/ and .agents/", "outside");
    }
    // Decision 1: everything under `.moai/` follows the folder (one rename), so no file
    // there moves on its own; when the folder stays, all of it stays.
    if top == MOAI_DIR {
        return if folder_moves {
            (
                "with_folder",
                "moves with .moai/ as a whole",
                "go_folder",
                String::new(),
            )
        } else {
            leave("the .moai/ folder stays (see the .moai item)", "go_folder")
        };
    }
    if key == MANIFEST {
        return leave("the Go manifest itself", "go_folder");
    }
    // #2936 O3 (user decision 6877929684): the local settings stay where they are, so the
    // deny rules in them keep working; the backup gets a copy (see `root_files`).
    if key == SETTINGS_LOCAL {
        return leave(
            "stays in place; a copy goes into the backup (O3, see root_files)",
            "local_settings",
        );
    }
    // AUDIT #2584 B1: the shared settings are the live host settings. O3 moves them only with
    // the Go folder, when the install step writes new ones; otherwise they stay.
    if key == SETTINGS && !folder_moves {
        return leave(
            "live host settings; moving them drops permissions, hooks and the status line",
            "settings",
        );
    }
    if linked_component(project, key) {
        return leave("a path component is a symlink", "unknown");
    }
    let path = project.join(key);
    match fs::symlink_metadata(&path) {
        Err(_) => return leave("missing", "unknown"),
        Ok(m) if !m.is_file() => return leave("not a regular file", "unknown"),
        Ok(_) => {}
    }
    match entry.provenance.as_str() {
        "user_created" => return leave("user_created: the user's own file", "user_data"),
        "template_managed" | "user_modified" => {}
        _ => return leave("not template_managed", "unknown"),
    }
    if top == ".agents" && !published_skill(key) {
        return leave(
            "under .agents/ only the published skill files are the Go tool's",
            "unknown",
        );
    }
    let unchanged = match sha256_file(&path) {
        Ok(h) => {
            entry.provenance == "template_managed"
                && !entry.deployed_hash.is_empty()
                && h == entry.deployed_hash
        }
        Err(_) => return leave("not a bounded regular file", "unknown"),
    };
    let proof = if unchanged {
        "manifest; sha256 matches the deploy"
    } else if entry.provenance == "user_modified" {
        "manifest; recorded user_modified"
    } else {
        "manifest; sha256 differs from the deploy"
    }
    .to_string();
    if key == SETTINGS {
        return (
            "move",
            "the Go settings move to the backup; the install step writes new ones (O3)",
            "settings",
            proof,
        );
    }
    let class = if unchanged {
        "tool_asset"
    } else {
        "user_modified"
    };
    for (file, text) in settings {
        if let Some(named) = referenced(project, text, key) {
            let reason = match (named, *file == SETTINGS_LOCAL) {
                (Named::Path, true) => "referenced by .claude/settings.local.json",
                (Named::Path, false) => "referenced by .claude/settings.json",
                (Named::StyleName, true) => {
                    "referenced by .claude/settings.local.json: its outputStyle names this style"
                }
                (Named::StyleName, false) => {
                    "referenced by .claude/settings.json: its outputStyle names this style"
                }
            };
            return ("leave", reason, class, proof);
        }
    }
    if unchanged {
        ("move", "proven Go deploy, unchanged", class, proof)
    } else if !folder_moves {
        ("leave", "changed since the Go deploy", class, proof)
    } else {
        (
            "move",
            "edited since the Go deploy; moves with both hashes recorded for the stage-3 merge \
             (O2)",
            class,
            proof,
        )
    }
}

/// Every file under `<project>/<rel>` (a real folder), without following a link: `Err` with
/// the reason when the walk meets a link, a special file, a secret name, or more than
/// `limit` files or folders. Nothing is opened.
fn plain_files_under(
    project: &Path,
    rel: &str,
    limit: u64,
) -> std::result::Result<Vec<String>, String> {
    let base = project.join(rel);
    let mut out = Vec::new();
    let mut dirs = 0u64;
    let mut stack = vec![base.clone()];
    while let Some(dir) = stack.pop() {
        let entries = fs::read_dir(&dir).map_err(|_| "a folder could not be read".to_string())?;
        for e in entries {
            let e = e.map_err(|_| "a folder entry could not be read".to_string())?;
            if e.file_name().to_str().is_none_or(secret_name) {
                return Err("holds a secret-named entry; never read".into());
            }
            let p = e.path();
            let m = fs::symlink_metadata(&p).map_err(|_| "metadata unavailable".to_string())?;
            if m.is_dir() {
                dirs += 1;
                if dirs > limit {
                    return Err(format!("more than {limit} folders; not checked"));
                }
                stack.push(p);
            } else if m.is_file() {
                let inner = p
                    .strip_prefix(&base)
                    .ok()
                    .and_then(|r| r.to_str())
                    .ok_or("non UTF-8 path")?
                    .replace('\\', "/");
                out.push(inner);
                if out.len() as u64 > limit {
                    return Err(format!("more than {limit} files; not checked"));
                }
            } else {
                return Err("holds a link or special file; never opened or followed".into());
            }
        }
    }
    out.sort();
    Ok(out)
}

/// #2936: the entries of `.agents/skills/` that the Go skill mirror made. A link moves as a
/// link when its body is exactly `../../.claude/skills/<n>` and the Go manifest records a file
/// under that canonical folder; a real folder moves whole when every file in it matches the
/// recorded `deployed_hash` of the same canonical path (O5: only some → the whole folder is
/// a conflict and stays). A folder holding recorded files (a published skill) is left to the
/// per-file items; anything else under `.agents/` is never listed (it is counted unrecorded).
fn classify_agents_skills(
    project: &Path,
    recorded: &HashMap<String, Entry>,
    folder_moves: bool,
) -> Vec<Item> {
    let base = project.join(AGENTS_SKILLS);
    if linked_component(project, AGENTS_SKILLS)
        || !fs::symlink_metadata(&base).is_ok_and(|m| m.is_dir())
    {
        return Vec::new();
    }
    let mut names: Vec<String> = fs::read_dir(&base)
        .map(|rd| {
            rd.filter_map(|e| e.ok())
                .filter_map(|e| e.file_name().into_string().ok())
                .collect()
        })
        .unwrap_or_default();
    names.sort();
    let limit = mirror_file_limit();
    let mut out = Vec::new();
    for name in names {
        if secret_name(&name) {
            continue;
        }
        let rel = format!("{AGENTS_SKILLS}/{name}");
        if recorded.keys().any(|k| k.starts_with(&format!("{rel}/"))) {
            continue;
        }
        let canon = format!("{CANONICAL_SKILLS}/{name}/");
        let recorded_under = recorded.keys().any(|k| k.starts_with(&canon));
        let Ok(m) = fs::symlink_metadata(project.join(&rel)) else {
            continue;
        };
        let moving = |mut i: Item, reason: &str| {
            if folder_moves {
                i.action = "move".into();
                i.reason = reason.into();
            } else {
                i.action = "leave".into();
                i.reason = "moves only with the Go .moai/ folder (see the .moai item)".into();
            }
            i
        };
        if m.file_type().is_symlink() {
            let body = fs::read_link(project.join(&rel))
                .ok()
                .and_then(|b| b.to_str().map(str::to_string))
                .unwrap_or_default();
            let want = format!("../../{CANONICAL_SKILLS}/{name}");
            let mut i = new_item(&rel, "leave", String::new(), "unknown");
            i.proof = format!("link {body}");
            out.push(if body == want && recorded_under {
                i.class = "go_mirror".into();
                moving(
                    i,
                    "the Go skill mirror link; moves as a link, never followed",
                )
            } else {
                i.reason = "a link that is not the Go skill mirror; never followed".into();
                i
            });
            continue;
        }
        if !m.is_dir() || !recorded_under {
            continue;
        }
        let files = match plain_files_under(project, &rel, limit) {
            Ok(f) => f,
            Err(why) => {
                out.push(new_item(&rel, "leave", why, "unknown"));
                continue;
            }
        };
        let (mut same, mut differ) = (0usize, 0usize);
        for inner in &files {
            let rec = recorded.get(&format!("{canon}{inner}"));
            let h = sha256_file(&project.join(&rel).join(inner));
            match (rec, h) {
                (Some(e), Ok(h)) if !e.deployed_hash.is_empty() && e.deployed_hash == h => {
                    same += 1
                }
                _ => differ += 1,
            }
        }
        out.push(if files.is_empty() {
            new_item(&rel, "leave", "an empty folder".into(), "unknown")
        } else if differ == 0 {
            let mut i = new_item(&rel, "leave", String::new(), "go_mirror_copy");
            i.proof = format!("every file matches the deployed_hash under {canon}");
            moving(
                i,
                "a copy of the Go skill mirror; every file matches the Go deploy",
            )
        } else if same > 0 {
            new_item(
                &rel,
                "leave",
                "only some files match the Go deploy; the whole folder stays (O5)".into(),
                "conflict",
            )
        } else {
            new_item(
                &rel,
                "leave",
                "no file matches the Go deploy".into(),
                "unknown",
            )
        });
    }
    out
}

/// O6: secret-named entries under `.claude/` and `.agents/` (a secret-named folder counts
/// once and is not entered), without reading any; `true` when the walk passed `limit`.
fn count_secrets(project: &Path, limit: u64) -> (u64, bool) {
    let (mut found, mut seen) = (0u64, 0u64);
    for root in [".claude", ".agents"] {
        let base = project.join(root);
        if !fs::symlink_metadata(&base).is_ok_and(|m| m.is_dir()) {
            continue;
        }
        let mut stack = vec![base];
        while let Some(dir) = stack.pop() {
            let Ok(entries) = fs::read_dir(&dir) else {
                continue;
            };
            for e in entries.flatten() {
                seen += 1;
                if seen > limit {
                    return (found, true);
                }
                if e.file_name().to_str().is_none_or(secret_name) {
                    found += 1;
                    continue;
                }
                let p = e.path();
                if fs::symlink_metadata(&p).is_ok_and(|m| m.is_dir())
                    && fs::symlink_metadata(p.join(".git")).is_err()
                {
                    stack.push(p);
                }
            }
        }
    }
    (found, false)
}

/// U4: one read-only `git status` over the paths S1b moves, as (path, state) rows; `None`
/// when the project is not a git checkout or the status cannot be read.
fn git_states(project: &Path) -> Option<Vec<(String, &'static str)>> {
    fs::symlink_metadata(project.join(".git")).ok()?;
    let mut cmd = std::process::Command::new("git");
    crate::identity::scrub_git_env(&mut cmd);
    let out = cmd
        .arg("--no-optional-locks")
        .arg("-C")
        .arg(project)
        .args([
            "status",
            "--porcelain=v1",
            "-z",
            "--untracked-files=all",
            "--ignored=matching",
            "--",
            ".claude",
            ".agents",
        ])
        .args(ROOT_FILES)
        .output()
        .ok()?;
    if !out.status.success() {
        return None;
    }
    let text = String::from_utf8_lossy(&out.stdout).into_owned();
    let mut rows = Vec::new();
    let mut fields = text.split('\0');
    while let Some(f) = fields.next() {
        if f.len() < 4 {
            continue;
        }
        let (xy, path) = (&f[..2], &f[3..]);
        if xy.starts_with('R') || xy.starts_with('C') {
            // The rename's source path follows as its own field.
            fields.next();
        }
        let state = if xy == "??" || xy == "!!" {
            "untracked"
        } else {
            "modified"
        };
        rows.push((path.to_string(), state));
    }
    Some(rows)
}

/// What git knows of one path (a folder: the worst of its files).
fn git_state(rows: &Option<Vec<(String, &'static str)>>, path: &str) -> &'static str {
    let Some(rows) = rows else {
        return "unknown";
    };
    let mut state = "clean";
    for (p, s) in rows {
        let hit = p == path
            || p.starts_with(&format!("{path}/"))
            || (p.ends_with('/') && path.starts_with(p.as_str()));
        if hit && (state == "clean" || *s == "modified") {
            state = s;
        }
    }
    state
}

/// How a settings file names a recorded file.
enum Named {
    /// By path (hooks, statusline).
    Path,
    /// An output style, by the style name in `outputStyle`.
    StyleName,
}

/// The `name:` of a Markdown file's frontmatter (the first `---` block), read from at most
/// the first 16 KiB; `None` when there is none or the file is not readable.
fn style_frontmatter_name(path: &Path) -> Option<String> {
    use std::io::Read as _;
    let m = fs::symlink_metadata(path).ok().filter(|m| m.is_file())?;
    let mut buf = Vec::new();
    fs::File::open(path)
        .ok()?
        .take(m.len().min(16 * 1024))
        .read_to_end(&mut buf)
        .ok()?;
    let text = String::from_utf8_lossy(&buf);
    let mut lines = text.lines();
    (lines.next()?.trim_end() == "---").then_some(())?;
    for line in lines {
        let line = line.trim_end();
        if line == "---" {
            return None;
        }
        if let Some(v) = line.strip_prefix("name:") {
            let v = v.trim().trim_matches(|c| c == '"' || c == '\'');
            return (!v.is_empty()).then(|| v.to_string());
        }
    }
    None
}

/// `.claude/settings.json` names the file by path (hooks, statusline) or, for an output
/// style, by the style name in `outputStyle`. The style name is compared without case with
/// the file's name and with the `name:` of its frontmatter (#2936 O-6): the settings say
/// `MoAI-Easy` while the file is `moai-easy.md`.
fn referenced(project: &Path, settings: &str, key: &str) -> Option<Named> {
    if settings.is_empty() {
        return None;
    }
    if settings.contains(key) {
        return Some(Named::Path);
    }
    let stem = key
        .strip_prefix(".claude/output-styles/")
        .and_then(|r| Path::new(r).file_stem())
        .and_then(|s| s.to_str())?;
    let style = serde_json::from_str::<serde_json::Value>(settings)
        .ok()
        .and_then(|v| v["outputStyle"].as_str().map(str::to_string))?;
    if style == stem {
        return Some(Named::Path);
    }
    let named = style.eq_ignore_ascii_case(stem)
        || style_frontmatter_name(&project.join(key))
            .is_some_and(|n| style.eq_ignore_ascii_case(&n));
    named.then_some(Named::StyleName)
}

fn read_bounded_text(project: &Path, rel: &str, limit: u64) -> Result<Option<String>> {
    if linked_component(project, rel) {
        return Err(format!("{rel}: a path component is a symlink; refused"));
    }
    let path = project.join(rel);
    match fs::symlink_metadata(&path) {
        Err(_) => Ok(None),
        Ok(m) if !m.is_file() || m.len() > limit => {
            Err(format!("{rel} is not a bounded regular file"))
        }
        Ok(_) => fs::read_to_string(&path)
            .map(Some)
            .map_err(|_| format!("{rel} read failed")),
    }
}

/// How many nested checkouts (folders holding a `.git` file or folder) are at or under
/// `<project>/<rel>`, not descending into them; `None` when the walk passes `limit` entries
/// before it ends (unknown, so the caller stays fail-closed).
pub(crate) fn nested_checkouts_under(project: &Path, rel: &str, limit: u64) -> Option<usize> {
    let base = project.join(rel);
    // The folder itself may be a checkout (another repository's worktree at `.moai`, #2753
    // re-audit F2 6859648118): that alone keeps it in place.
    if fs::symlink_metadata(base.join(".git")).is_ok() {
        return Some(1);
    }
    let mut stack = vec![base];
    let (mut seen, mut nested) = (0u64, 0usize);
    while let Some(dir) = stack.pop() {
        let Ok(entries) = fs::read_dir(&dir) else {
            continue;
        };
        for e in entries.flatten() {
            seen += 1;
            if seen > limit {
                return None;
            }
            let p = e.path();
            match fs::symlink_metadata(&p) {
                Ok(m) if m.is_dir() => {
                    if fs::symlink_metadata(p.join(".git")).is_ok() {
                        nested += 1;
                    } else {
                        stack.push(p);
                    }
                }
                _ => {}
            }
        }
    }
    Some(nested)
}

/// How many git worktrees of the project's repository are registered at or under
/// `<project>/<rel>`; `None` when the project is a git checkout whose worktree list cannot be
/// read. A folder that is not a git checkout has no registered worktrees.
fn registered_worktrees_under(project: &Path, rel: &str) -> Option<usize> {
    let mut cmd = std::process::Command::new("git");
    crate::identity::scrub_git_env(&mut cmd);
    let out = cmd
        .arg("--no-optional-locks")
        .arg("-C")
        .arg(project)
        .args(["worktree", "list", "--porcelain"])
        .output();
    let listing = match out {
        Ok(o) if o.status.success() => String::from_utf8_lossy(&o.stdout).into_owned(),
        _ if fs::symlink_metadata(project.join(".git")).is_err() => return Some(0),
        _ => return None,
    };
    let base = project.join(rel);
    Some(
        crate::worktree_report::parse_worktree_list(&listing)
            .iter()
            .map(|e| fs::canonicalize(&e.path).unwrap_or_else(|_| e.path.clone()))
            .filter(|p| p.starts_with(&base))
            .count(),
    )
}

/// Files under `.moai/` and `.claude/` the manifest does not record, per second-level folder,
/// and how many nested checkouts (a folder holding a `.git` entry, such as a worktree under
/// `.claude/worktrees/`) were skipped unread. A project's worktrees can hold hundreds of
/// thousands of files; they are never the Go tool's (#2561 F1, comment 6856196324).
fn count_unrecorded(
    project: &Path,
    recorded: &HashMap<String, Entry>,
    limit: u64,
) -> (BTreeMap<String, u64>, u64, bool) {
    let mut counts = BTreeMap::new();
    let mut nested = 0u64;
    let mut seen = 0u64;
    for root in ROOTS {
        let base = project.join(root);
        let Ok(m) = fs::symlink_metadata(&base) else {
            continue;
        };
        if !m.is_dir() {
            continue;
        }
        let mut stack: Vec<PathBuf> = vec![base];
        while let Some(dir) = stack.pop() {
            let Ok(entries) = fs::read_dir(&dir) else {
                continue;
            };
            for e in entries.flatten() {
                // Bounded by entries, not time, so one tree always gives one answer.
                seen += 1;
                if seen > limit {
                    return (counts, nested, true);
                }
                let p = e.path();
                let Ok(m) = fs::symlink_metadata(&p) else {
                    continue;
                };
                if m.is_dir() {
                    if fs::symlink_metadata(p.join(".git")).is_ok() {
                        nested += 1;
                    } else {
                        stack.push(p);
                    }
                    continue;
                }
                if !m.is_file() {
                    continue;
                }
                let Ok(rel) = p.strip_prefix(project) else {
                    continue;
                };
                let rel = rel.to_string_lossy().replace('\\', "/");
                if recorded.contains_key(&rel) || rel == MANIFEST {
                    continue;
                }
                let folder: String = rel.splitn(3, '/').take(2).collect::<Vec<_>>().join("/");
                *counts.entry(folder).or_insert(0) += 1;
            }
        }
    }
    (counts, nested, false)
}

/// The earlier root's entries and what the move does with each, plus the reason the whole
/// folder must stay when one of them cannot move safely (a link is never followed; the earlier
/// root itself must be a real folder).
fn plan_state_moves(project: &Path) -> (Vec<StateMove>, Option<String>) {
    let root = project.join(LEGACY_ROOT);
    match fs::symlink_metadata(&root) {
        Err(_) => return (Vec::new(), None),
        Ok(m) if !m.is_dir() => {
            return (
                Vec::new(),
                Some(format!(
                    "{LEGACY_ROOT} is a link or not a folder; its state cannot move into the new \
                     .moai/, so the Go folder stays"
                )),
            )
        }
        Ok(_) => {}
    }
    let mut names: Vec<String> = fs::read_dir(&root)
        .map(|rd| {
            rd.filter_map(|e| e.ok())
                .filter_map(|e| e.file_name().into_string().ok())
                .collect()
        })
        .unwrap_or_default();
    names.sort();
    let mut out = Vec::new();
    let mut block = None;
    for name in names {
        let from = format!("{LEGACY_ROOT}/{name}");
        let linked = fs::symlink_metadata(root.join(&name))
            .map(|m| m.file_type().is_symlink())
            .unwrap_or(true);
        let (action, reason) = if STATE_ENTRIES.contains(&name.as_str()) {
            if linked {
                block.get_or_insert_with(|| {
                    format!(
                        "{from} is a link; product state is never moved through a link, so the \
                         Go folder stays"
                    )
                });
                ("leave", "a link; never followed".to_string())
            } else {
                ("move", "this generation's product state".to_string())
            }
        } else {
            let why = match name.as_str() {
                "worktrees" => {
                    "lane worktrees stay until worktree-move-plan moves them (#2723, D6)"
                }
                "reports" => "reports keep the earlier root (#2725 D5)",
                "backups" => "earlier go-state backups stay where their receipts expect them",
                _ => "not product state this build moves; it stays",
            };
            ("leave", why.to_string())
        };
        out.push(StateMove {
            to: if action == "move" {
                format!("{MOAI_DIR}/{name}")
            } else {
                String::new()
            },
            from,
            action: action.into(),
            reason,
        });
    }
    (out, block)
}

/// A bounded regular file's bytes, refusing a link at the path itself; `None` otherwise.
fn read_bounded_file(path: &Path) -> Option<Vec<u8>> {
    let m = fs::symlink_metadata(path).ok()?;
    if !m.is_file() || m.len() > FILE_LIMIT {
        return None;
    }
    let bytes = fs::read(path).ok()?;
    ((bytes.len() as u64) <= FILE_LIMIT).then_some(bytes)
}

/// The root instruction files (#2936 O4, D-M1): each moves into the backup's `root-files/`
/// with the Go folder, its Go proof shown (`manifest`, a Go curator `marker`, or `none`). A
/// link stays and is never followed; this generation's file (its managed section) stays; one
/// with both a managed section and Go proof is a conflict the user decides; a file the
/// manifest records `user_created` stays. `.claude/settings.local.json` stays in place and
/// only a copy goes into the backup (O3).
fn plan_root_files(project: &Path, recorded: &HashMap<String, Entry>) -> Vec<RootFile> {
    let mut out = Vec::new();
    let row =
        |path: &str, sha: String, action: &str, reason: &str, class: &str, proof: &str| RootFile {
            path: path.into(),
            sha256: sha,
            action: action.into(),
            backup_copy: format!("{MOAI_DIR}/backups/<UTC>/root-files/{path}"),
            user_choice: "keep".into(),
            reason: reason.into(),
            class: class.into(),
            go_proof: proof.into(),
            git: String::new(),
        };
    for name in ROOT_FILES {
        let path = project.join(name);
        let Ok(m) = fs::symlink_metadata(&path) else {
            continue;
        };
        let rec = recorded.get(name);
        if m.file_type().is_symlink() {
            out.push(row(
                name,
                String::new(),
                "leave",
                "a link; not moved, never followed",
                "instructions",
                "",
            ));
            continue;
        }
        let Some(bytes) = read_bounded_file(&path) else {
            out.push(row(
                name,
                String::new(),
                "leave",
                "not a bounded regular file",
                "instructions",
                "",
            ));
            continue;
        };
        let text = String::from_utf8_lossy(&bytes);
        let rust = text.contains(RUST_MARKER_TEXT);
        let proof = if rec.is_some() {
            "manifest"
        } else if GO_MARKERS.iter().any(|m| text.contains(m)) {
            "marker"
        } else {
            "none"
        };
        let (action, reason, class) = if rec.is_some_and(|e| e.provenance == "user_created") {
            ("leave", "user_created: the user's own file", "user_data")
        } else if rust && proof != "none" {
            (
                "leave",
                "holds this generation's managed section and Go proof; the user decides",
                "conflict",
            )
        } else if rust {
            (
                "leave",
                "this generation's file (managed section); it stays",
                "rust",
            )
        } else if proof == "none" {
            (
                "move",
                "no Go proof; moves to the backup as D-M1 says (O4); the install step writes a \
                 new one",
                "instructions",
            )
        } else {
            (
                "move",
                "moves to the backup; the install step writes a new one",
                "instructions",
            )
        };
        out.push(row(name, digest(&bytes), action, reason, class, proof));
    }
    if fs::symlink_metadata(project.join(SETTINGS_LOCAL)).is_ok() {
        let proof = if recorded.contains_key(SETTINGS_LOCAL) {
            "manifest"
        } else {
            "none"
        };
        let bytes = if linked_component(project, SETTINGS_LOCAL) {
            None
        } else {
            read_bounded_file(&project.join(SETTINGS_LOCAL))
        };
        out.push(match bytes {
            Some(b) => row(
                SETTINGS_LOCAL,
                digest(&b),
                "copy",
                "stays in place; a copy goes into the backup (O3)",
                "local_settings",
                proof,
            ),
            None => row(
                SETTINGS_LOCAL,
                String::new(),
                "leave",
                "a link or not a bounded regular file; not copied",
                "local_settings",
                proof,
            ),
        });
    }
    out
}

/// #2936 R-A: what a recorded file under `.claude/` or `.agents/` is for, by its place; empty
/// for any other path.
fn kind_of(path: &str) -> &'static str {
    let top = top_of(path);
    if top != ".claude" && top != ".agents" {
        return "";
    }
    let under = |dir: &str| path.starts_with(&format!("{dir}/"));
    if path == SETTINGS {
        "settings"
    } else if path.rsplit('/').next() == Some("loop.md") {
        "loop"
    } else if under(".claude/hooks") {
        "hook"
    } else if under(".claude/output-styles") {
        "output_style"
    } else if under(".claude/rules") {
        "rule"
    } else if under(".claude/commands") {
        "command"
    } else if under(".claude/workflows") {
        "workflow"
    } else if under(".claude/skills") || under(AGENTS_SKILLS) {
        "skill"
    } else if under(".claude/agents") {
        "agent"
    } else {
        "other"
    }
}

/// #2936 R-A: what this generation puts in place of one kind of moved file, and a note. The
/// user layer is named, never read (O-3: this preview does not read HOME).
fn counterpart_of(kind: &str) -> (&'static str, &'static str) {
    const GONE: &str = "no counterpart \u{2014} will be gone";
    const USER_LAYER: &str = "user layer after moai update";
    match kind {
        "hook" => (
            "reinstalled by hooks",
            "the MoAI hook set is written into .claude/settings.json by the hooks step; the \
             moved scripts themselves are not reinstalled",
        ),
        "output_style" => (
            USER_LAYER,
            "the style name resolves from HOME's output-styles only; this project keeps no \
             file for it (HOME is not read here)",
        ),
        "skill" | "agent" => (
            USER_LAYER,
            "moai update installs it under HOME; the project keeps no copy",
        ),
        "settings" => (
            "reinstalled by claude-settings",
            "only the keys MoAI owns are written; hooks, permissions.deny and the other keys \
             come back with the merge phase (a later slice)",
        ),
        _ => (
            GONE,
            "this generation installs nothing in its place; the files stay in the backup",
        ),
    }
}

/// #2936 R-A: the moving items grouped by kind, and the counterpart of each kind.
fn plan_counterparts(items: &[Item]) -> (BTreeMap<String, u64>, Vec<Counterpart>) {
    let mut kinds = BTreeMap::new();
    for i in items
        .iter()
        .filter(|i| moving_file(i) && !i.kind.is_empty())
    {
        *kinds.entry(i.kind.clone()).or_insert(0u64) += 1;
    }
    let rows = kinds
        .iter()
        .map(|(kind, moved)| {
            let (counterpart, note) = counterpart_of(kind);
            Counterpart {
                kind: kind.clone(),
                moved: *moved,
                counterpart: counterpart.into(),
                note: note.into(),
            }
        })
        .collect();
    (kinds, rows)
}

/// An item that moves on its own (not the two folders that move whole).
fn moving_file(i: &Item) -> bool {
    i.action == "move" && i.path != MOAI_DIR && i.path != LEGACY_ROOT
}

/// #2936 R-A: the folders under the project that hold nothing once `moving` is gone, deepest
/// first so a folder whose only content is an emptied folder counts. A folder is listed only
/// when every entry in it moves or is itself listed; a script the settings name stays and
/// keeps its folders. The walk reads at most `limit` entries (all of a folder's, or none of
/// it); the second value is `true` when it stopped there.
fn empty_after(
    project: &Path,
    moving: &std::collections::BTreeSet<&str>,
    limit: u64,
) -> (std::collections::BTreeSet<String>, bool) {
    let mut candidates = std::collections::BTreeSet::new();
    for path in moving {
        let mut cur = Path::new(path).parent();
        while let Some(p) = cur.filter(|p| !p.as_os_str().is_empty()) {
            if let Some(s) = p.to_str() {
                candidates.insert(s.to_string());
            }
            cur = p.parent();
        }
    }
    let mut ordered: Vec<String> = candidates.into_iter().collect();
    ordered.sort_by(|a, b| {
        let depth = |s: &str| s.matches('/').count();
        depth(b).cmp(&depth(a)).then_with(|| a.cmp(b))
    });
    let mut empty = std::collections::BTreeSet::new();
    let mut seen = 0u64;
    for folder in ordered {
        let dir = project.join(&folder);
        if !fs::symlink_metadata(&dir).is_ok_and(|m| m.is_dir()) {
            continue;
        }
        let Ok(rd) = fs::read_dir(&dir) else {
            continue;
        };
        let names: Vec<String> = rd
            .filter_map(|e| e.ok())
            .map(|e| e.file_name().to_string_lossy().into_owned())
            .collect();
        seen += names.len() as u64;
        if seen > limit {
            return (empty, true);
        }
        let gone = |name: &String| {
            let rel = format!("{folder}/{name}");
            moving.contains(rel.as_str()) || empty.contains(&rel)
        };
        if !names.is_empty() && names.iter().all(gone) {
            empty.insert(folder);
        }
    }
    (empty, false)
}

/// #2936 R-A: the hook entries and deny rules of the settings file that is about to move.
fn wiring_of(settings_text: &str) -> Option<WiringLost> {
    let v: serde_json::Value = serde_json::from_str(settings_text).ok()?;
    let go_hook_entries = v["hooks"]
        .as_object()
        .map(|h| crate::go_hooks::split(h).0.len())
        .unwrap_or(0);
    let deny_rules = v["permissions"]["deny"]
        .as_array()
        .map(Vec::len)
        .unwrap_or(0);
    let mut keys: Vec<String> = v.as_object()?.keys().cloned().collect();
    keys.sort();
    Some(WiringLost {
        file: SETTINGS.into(),
        go_hook_entries,
        deny_rules,
        keys,
    })
}

/// #2936 R-A: what the move leaves behind. Reads only; nothing is written.
fn plan_after_move(
    project: &Path,
    items: &[Item],
    settings: &[(&'static str, String)],
) -> AfterMove {
    let moving: std::collections::BTreeSet<&str> = items
        .iter()
        .filter(|i| moving_file(i))
        .map(|i| i.path.as_str())
        .collect();
    if moving.is_empty() {
        return AfterMove::default();
    }
    let (empty, walk_stopped) = empty_after(project, &moving, unrecorded_limit());
    let paths: Vec<String> = moving.iter().map(|p| p.to_string()).collect();
    let (dangling, unread) = crate::worktree_include::lists_selecting(project, &paths);
    let wiring_lost = moving
        .contains(SETTINGS)
        .then(|| {
            settings
                .iter()
                .find(|(f, _)| *f == SETTINGS)
                .and_then(|(_, text)| wiring_of(text))
        })
        .flatten();
    AfterMove {
        empty_folders: empty.into_iter().collect(),
        dangling,
        wiring_lost,
        unknown: walk_stopped || unread,
    }
}

/// The read-only preview over one project folder.
pub fn plan(project: &Path) -> Result<GoStatePlan> {
    let project = fs::canonicalize(project).map_err(|_| "project folder unavailable")?;
    // A linked `.moai` (or any linked component above the manifest) is refused by the
    // manifest read below, with or without a manifest file.
    let mut warnings = vec![
        "Running Claude Code sessions in this project may read .claude/hooks, rules and output \
         styles; apply only while none is running (this preview cannot tell)."
            .to_string(),
    ];
    let manifest = match read_bounded_text(&project, MANIFEST, MANIFEST_LIMIT)? {
        None => None,
        Some(text) => serde_json::from_str::<Manifest>(&text).ok(),
    };
    // #2670: the Go generation's worktree include list copies `.env*` into every new Claude
    // Code worktree. Reported, never edited (a line may be deliberate).
    warnings.extend(crate::worktree_include::warnings(&project));
    let mut settings = Vec::new();
    for file in [SETTINGS, SETTINGS_LOCAL] {
        let text = read_bounded_text(&project, file, MANIFEST_LIMIT)?.unwrap_or_default();
        settings.push((file, text));
    }
    let mut items = Vec::new();
    let (version, recorded) = match manifest {
        Some(m) => (Some(m.version), m.files),
        None => {
            warnings.push(format!(
                "no usable Go manifest at {MANIFEST}: nothing is proven tool-made, so nothing moves"
            ));
            (None, HashMap::new())
        }
    };
    // The `.moai/` folder item. Only with a usable manifest (nothing is proven otherwise),
    // only a real folder (a linked `.moai` was refused by the manifest read above), and
    // never while either settings file names a path under `.moai/` (decision 2: what the
    // settings reference stays, fail-closed; cleanup is #2581).
    let folder_is_dir = fs::symlink_metadata(project.join(MOAI_DIR))
        .map(|m| m.is_dir())
        .unwrap_or(false);
    let referenced_by_settings = settings.iter().any(|(_, text)| text.contains(".moai/"));
    // Moving `.moai/` also moves every git worktree registered under it, and git then loses
    // them (their recorded paths no longer exist). Unreadable means "maybe": stay.
    let checked = version.is_some() && folder_is_dir;
    let worktrees_inside = if checked {
        registered_worktrees_under(&project, MOAI_DIR)
    } else {
        Some(0)
    };
    // A worktree of another repository (or any nested checkout) under `.moai/` is not in this
    // repository's list, yet moving the folder cuts it off too (#2753 audit F1 6858965696).
    let nested_inside = if checked {
        nested_checkouts_under(&project, MOAI_DIR, unrecorded_limit())
    } else {
        Some(0)
    };
    let (state_moves, state_block) = if checked {
        plan_state_moves(&project)
    } else {
        (Vec::new(), None)
    };
    let folder_moves = checked
        && !referenced_by_settings
        && worktrees_inside == Some(0)
        && nested_inside == Some(0)
        && state_block.is_none();
    if version.is_some() && folder_is_dir {
        items.push(new_item(
            MOAI_DIR,
            if folder_moves { "move" } else { "leave" },
            if folder_moves {
                "the Go .moai/ folder as a whole (decision 1, #2561 comment 6831345102)".into()
            } else if referenced_by_settings {
                "settings reference .moai/; the folder stays until they are cleaned (#2581)".into()
            } else if let Some(block) = &state_block {
                block.clone()
            } else {
                match (worktrees_inside, nested_inside) {
                    (Some(n), _) if n > 0 => format!(
                        "{n} git worktree(s) are registered under .moai/; moving the folder would \
                         cut git off from them, so it stays until they are moved out \
                         (worktree-move-plan, #2723)"
                    ),
                    (None, _) => "the git worktree list could not be read, so it is unknown \
                                  whether a worktree lives under .moai/; the folder stays"
                        .into(),
                    (_, Some(n)) => format!(
                        "{n} nested checkout(s) (folders holding a .git entry, such as worktrees \
                         of another repository) are under .moai/; moving the folder would cut \
                         their repositories off from them, so it stays until they are moved out"
                    ),
                    (_, None) => format!(
                        "the .moai/ folder holds more than {} entries, so it was not fully \
                         checked for nested checkouts; the folder stays",
                        unrecorded_limit()
                    ),
                }
            },
            "go_folder",
        ));
    }
    let go_backup = if fs::symlink_metadata(project.join(crate::state_root::GO_BACKUP)).is_err() {
        crate::state_root::GO_BACKUP
    } else {
        GO_BACKUP_STAMPED
    };
    if folder_moves {
        warnings.push(format!(
            "Stop moai-cli commands and agent sessions in this project before the apply: the \
             Go .moai/ becomes {go_backup}, a new .moai/ is made and the {LEGACY_ROOT} product \
             state moves into it."
        ));
    }
    let mut keys: Vec<&String> = recorded.keys().collect();
    keys.sort();
    for key in keys {
        // O6: a secret name under `.claude/` or `.agents/` is never listed, hashed or moved;
        // `secrets` counts what is on disk.
        if matches!(top_of(key), ".claude" | ".agents") && has_secret_component(key) {
            continue;
        }
        let (action, reason, class, proof) =
            policy(&project, key, &recorded[key], &settings, folder_moves);
        let mut item = new_item(key, action, reason.into(), class);
        item.proof = proof;
        item.deployed_hash = recorded[key].deployed_hash.clone();
        items.push(item);
    }
    if version.is_some() {
        items.extend(classify_agents_skills(&project, &recorded, folder_moves));
    }
    let (secrets, secrets_stopped) = count_secrets(&project, unrecorded_limit());
    if secrets_stopped {
        warnings.push(format!(
            "the secret-name count under .claude/ and .agents/ is a lower bound: its walk \
             passed {} entries",
            unrecorded_limit()
        ));
    }
    let unrecorded = if version.is_some() {
        let limit = unrecorded_limit();
        let (counts, nested, stopped) = count_unrecorded(&project, &recorded, limit);
        if stopped {
            warnings.push(format!(
                "the count of files the manifest does not record stopped after \
                 {limit} entries under .moai/, .claude/ and .agents/; the numbers in \
                 `unrecorded` are lower bounds (no file is moved on their account)"
            ));
        }
        if nested > 0 {
            warnings.push(format!(
                "{nested} nested checkout(s) under .moai/ or .claude/ (folders holding a .git \
                 entry, such as worktrees) were not read; their files are not counted and \
                 none of them is offered for moving"
            ));
        }
        counts
    } else {
        BTreeMap::new()
    };
    let restores = if folder_moves {
        plan_restores(&project, go_backup)
    } else {
        Vec::new()
    };
    // #2725 (leader condition 6854854683): only a `.moai/` that `state_root` classifies as
    // the Go generation's may move. Any other verdict (`rust`, `mixed`, `unknown`, `refuse`,
    // `interrupted`, `empty`) moves nothing, and the preview says why.
    let state = crate::state_root::classify(&project);
    let (items, restores) = if matches!(state.generation, "go" | "none") {
        (items, restores)
    } else {
        warnings.push(format!(
            "state_root is `{}`: {}; nothing moves",
            state.generation, state.reason
        ));
        let held = format!("state_root `{}`: nothing moves (#2725)", state.generation);
        let items = items
            .into_iter()
            .map(|mut i| {
                if i.action != "leave" {
                    i.action = "leave".into();
                    i.reason = held.clone();
                }
                i
            })
            .collect();
        (items, Vec::new())
    };
    // #2696 C4a: a project whose only state is the earlier `.moai-cli/` (verdict `none`, no
    // `.moai` of any kind) is offered the same move into a new `.moai/`, without a Go folder.
    let mut items = items;
    let legacy_moves = if state.generation == "none" {
        let (moves, block) = plan_state_moves(&project);
        let any = moves.iter().any(|m| m.action == "move");
        if let (true, Some(block)) = (any, &block) {
            items.push(new_item(
                LEGACY_ROOT,
                "leave",
                block.replace("so the Go folder stays", "so it stays where it is"),
                "legacy_state",
            ));
            Vec::new()
        } else if any {
            items.push(new_item(
                LEGACY_ROOT,
                "move",
                "the earlier root's product state moves into a new .moai/ (#2696 C4a)".into(),
                "legacy_state",
            ));
            warnings.push(format!(
                "Stop moai-cli commands and agent sessions in this project before the apply: a \
                 new .moai/ is made and the {LEGACY_ROOT} product state moves into it."
            ));
            moves
        } else {
            Vec::new()
        }
    } else {
        Vec::new()
    };
    let moves_folder = items
        .iter()
        .any(|i| i.path == MOAI_DIR && i.action == "move");
    let (go_backup, state_moves, root_files, left_in_backup) = if moves_folder {
        let mut left: Vec<String> = fs::read_dir(project.join(MOAI_DIR))
            .map(|rd| {
                rd.filter_map(|e| e.ok())
                    .filter_map(|e| e.file_name().into_string().ok())
                    .collect()
            })
            .unwrap_or_default();
        left.sort();
        (
            Some(go_backup.to_string()),
            state_moves,
            plan_root_files(&project, &recorded),
            left,
        )
    } else {
        (None, legacy_moves, Vec::new(), Vec::new())
    };
    // U4: what git knows of each moving path, from one read-only status (only when
    // something moves).
    let (mut items, mut root_files) = (items, root_files);
    let moving_path = |i: &Item| i.action == "move" && i.path != MOAI_DIR && i.path != LEGACY_ROOT;
    if items.iter().any(moving_path) || root_files.iter().any(|f| f.action != "leave") {
        let rows = git_states(&project);
        let mut hint = |path: &str, state: &str| {
            if state == "modified" {
                warnings.push(format!(
                    "{path}: git says modified; the committed version is `git show HEAD:{path}`"
                ));
            }
        };
        for i in items.iter_mut().filter(|i| moving_path(i)) {
            i.git = git_state(&rows, &i.path).into();
            hint(&i.path, &i.git);
        }
        for f in root_files.iter_mut().filter(|f| f.action != "leave") {
            f.git = git_state(&rows, &f.path).into();
            hint(&f.path, &f.git);
        }
    }
    for i in items.iter_mut() {
        i.kind = kind_of(&i.path).into();
    }
    let (kinds, counterparts) = plan_counterparts(&items);
    let after_move = plan_after_move(&project, &items, &settings);
    let mut plan = GoStatePlan {
        schema_version: SCHEMA_VERSION,
        project: project
            .to_str()
            .ok_or("non UTF-8 project path")?
            .to_string(),
        manifest_version: version,
        items,
        unrecorded,
        restores,
        go_backup,
        state_moves,
        root_files,
        left_in_backup,
        kinds,
        counterparts,
        after_move,
        secrets,
        warnings,
        plan_digest: String::new(),
    };
    let bytes = serde_json::to_vec(&plan).map_err(|_| "encode failed")?;
    plan.plan_digest = format!("sha256:{:x}", Sha256::digest(bytes));
    Ok(plan)
}

// ---- stage 2: apply and revert ----
//
// Two layouts, chosen by the preview's `.moai` item:
//
// * The Go folder moves (#2696). In this order, against one plan digest:
//   1. `.moai` → the Go backup (`.moai-backup`, or `.moai-backup-<UTC>` when taken), one
//      rename that never overwrites;
//   2. a new `.moai/` with the receipt `.moai/backups/<UTC>/go-state-receipt.json`, `pending`
//      (when this step fails, the rename is undone at once);
//   3. the proven `.claude/` files move to `.moai/backups/<UTC>/go-state/<same path>`;
//   4. the earlier root's product state moves, `.moai-cli/<entry>` → `.moai/<entry>`;
//   5. SPECs are copied from the backup into `.moai/specs`;
//   6. `CLAUDE.md`/`AGENTS.md`/`AGENTS.local.md` move to `.moai/backups/<UTC>/root-files/`, a
//      0600 copy of the local settings goes there too;
//   7. the marker `.moai/moai-cli.json` is written last, then the receipt says `committed`.
//   Stopped anywhere after step 1, `state_root` says `interrupted` and the revert takes over.
// * The Go folder stays (settings name it, a worktree is inside, …): only the proven
//   `.claude/` files move, into `<state folder>/backups/<UTC>/go-state/`, with the receipt
//   beside them — the #2561 stage-2 layout. Revert also reads receipts of the #2561 stage-3
//   layout, where the folder itself went to `.moai-cli/backups/<UTC>/go-state/.moai`.
//
// A file moves as a hard link to its new path followed by removing the source: the link fails
// when the destination exists (no overwrite) or lies on another filesystem (no copy fallback).
// Every step is recorded in the receipt as it happens, so a revert works after a failure too.

use crate::state_root::{utc_stamp as valid_stamp, GO_BACKUP, GO_STATE_RECEIPT as RECEIPT};

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq, Default)]
#[serde(deny_unknown_fields)]
pub struct Moved {
    pub path: String,
    /// A file's sha256, or for the folder its `baseline::tree_digest` taken right after
    /// the move (empty until then; empty for a `state` entry). A `link`: the sha256 of its
    /// body; a `mirror_copy`: the folder's `tree_digest`.
    pub sha256: String,
    pub moved: bool,
    /// `folder` for the `.moai/` rename, `state` for an earlier-root entry moving into the new
    /// `.moai/` (#2696); #2936: `edited` (a file the user changed, O2), `link` (the Go skill
    /// mirror link, moved as a link), `mirror_copy` (a copy folder of it), `root_file` (a root
    /// instruction file, into `root-files/`, O4); absent for an unchanged file.
    #[serde(default, skip_serializing_if = "String::is_empty")]
    pub kind: String,
    /// #2936 O2: for `edited`, the manifest's `deployed_hash` (`sha256` is the current one),
    /// for the stage-3 merge.
    #[serde(default, skip_serializing_if = "String::is_empty")]
    pub deployed_hash: String,
    /// For `link`: the link body, never followed.
    #[serde(default, skip_serializing_if = "String::is_empty")]
    pub target: String,
}

/// The receipt entry for one planned move other than the folder.
fn planned_move(project: &Path, item: &Item) -> Result<Moved> {
    let path = project.join(&item.path);
    let mut m = Moved {
        path: item.path.clone(),
        ..Moved::default()
    };
    match item.class.as_str() {
        "go_mirror" => {
            let body = fs::read_link(&path)
                .ok()
                .and_then(|b| b.to_str().map(str::to_string))
                .ok_or_else(|| format!("{}: not a link; refused", item.path))?;
            m.sha256 = digest(body.as_bytes());
            m.kind = "link".into();
            m.target = body;
        }
        "go_mirror_copy" => {
            refuse_linked(project, &item.path)?;
            m.sha256 = crate::baseline::tree_digest(&path, &[])?;
            m.kind = "mirror_copy".into();
        }
        _ => {
            m.sha256 = sha256_file(&path)?;
            if item.class == "user_modified" || m.sha256 != item.deployed_hash {
                m.kind = "edited".into();
                m.deployed_hash = item.deployed_hash.clone();
            }
        }
    }
    Ok(m)
}

/// Move a link as a link, never following it: its body must still be `body`; a new link
/// with that body is made at `to` (failing when anything is there), checked, and only then
/// is the source removed. A link is never hard-linked.
fn move_link(from: &Path, to: &Path, body: &str) -> Result<()> {
    let now = fs::read_link(from).map_err(|_| "link move refused: the source is not a link")?;
    if now.to_str() != Some(body) {
        return Err("link move refused: the link changed after the preview".into());
    }
    make_link(body, to).map_err(|e| {
        if e.kind() == std::io::ErrorKind::AlreadyExists {
            "move refused: the destination already exists".to_string()
        } else {
            "link move failed (symlink)".to_string()
        }
    })?;
    if fs::read_link(to).ok().as_deref().and_then(Path::to_str) != Some(body) {
        let _ = fs::remove_file(to);
        return Err("link move failed: the new link does not read back".into());
    }
    fs::remove_file(from).map_err(|_| "move failed: source not removed (both links kept)".into())
}

#[cfg(unix)]
fn make_link(body: &str, to: &Path) -> std::io::Result<()> {
    std::os::unix::fs::symlink(body, to)
}

#[cfg(not(unix))]
fn make_link(_body: &str, _to: &Path) -> std::io::Result<()> {
    Err(std::io::Error::other("links are moved only on Unix"))
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
#[serde(deny_unknown_fields)]
pub struct Receipt {
    pub schema_version: u32,
    pub project: String,
    pub backup: String,
    pub plan_digest: String,
    /// The caller's `--permission-reference`, recorded as given (not an authorization).
    pub permission_reference: String,
    /// `pending` → `committed` → `reverted`.
    pub state: String,
    pub moves: Vec<Moved>,
    /// AC-5: the SPEC copies written into the state folder, each with its sha256.
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub restored: Vec<Restored>,
    /// #2696: the Go folder's new name; empty for an apply that left the Go folder in place.
    #[serde(default, skip_serializing_if = "String::is_empty")]
    pub go_backup: String,
    /// #2696: the `CLAUDE.md`/`AGENTS.md` copies, each with its sha256.
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub copied: Vec<Restored>,
    /// #2696: the sha256 of the marker the apply wrote (empty before it, and after a revert
    /// removed it).
    #[serde(default, skip_serializing_if = "String::is_empty")]
    pub marker: String,
    /// #2696: the name a revert gave this generation's `.moai/` before the Go folder came back.
    #[serde(default, skip_serializing_if = "String::is_empty")]
    pub rust_aside: String,
    /// #2696 C4a: the apply moved only the earlier `.moai-cli/` state into a new `.moai/`
    /// (no Go folder).
    #[serde(default, skip_serializing_if = "std::ops::Not::not")]
    pub legacy_move: bool,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
#[serde(deny_unknown_fields)]
pub struct Restored {
    pub path: String,
    pub sha256: String,
}

fn stamp_now() -> String {
    crate::baseline::utc_now().replace(['-', ':'], "")
}

fn digest(bytes: &[u8]) -> String {
    format!("sha256:{:x}", Sha256::digest(bytes))
}

/// The receipt of the apply `stamp` when its backup area is under `base`.
fn receipt_rel(base: &str, stamp: &str) -> String {
    format!("{base}/backups/{stamp}/{RECEIPT}")
}

/// Every existing component of `rel` under `project` must be a real folder (not a link).
fn refuse_linked(project: &Path, rel: &str) -> Result<()> {
    if linked_component(project, rel) {
        return Err(format!("{rel}: a path component is a symlink; refused"));
    }
    Ok(())
}

fn save_receipt(
    project: &Path,
    base: &str,
    receipt: &Receipt,
    before: Option<&[u8]>,
) -> Result<Vec<u8>> {
    let rel = receipt_rel(base, &receipt.backup);
    refuse_linked(project, &rel)?;
    let bytes = serde_json::to_vec_pretty(receipt).map_err(|_| "encode failed")?;
    let path = project.join(&rel);
    match before {
        None => crate::apply::publish_exclusive_mode(&path, &bytes, Some(PRIVATE_FILE))?,
        Some(b) => crate::apply::publish_replace(&path, Some(b), &bytes)?,
    }
    Ok(bytes)
}

/// Owner-only bits for a backup folder, and for a receipt or a copy that may hold a token
/// (#2966 audit B3).
const PRIVATE_DIR: u32 = 0o700;
const PRIVATE_FILE: u32 = 0o600;

/// Create one folder (not its parents); a private one is 0700 from its creation.
#[cfg(unix)]
fn create_dir_with(path: &Path, private: bool) -> std::io::Result<()> {
    use std::os::unix::fs::DirBuilderExt;
    let mut builder = fs::DirBuilder::new();
    if private {
        builder.mode(PRIVATE_DIR);
    }
    builder.create(path)
}

#[cfg(not(unix))]
fn create_dir_with(path: &Path, _private: bool) -> std::io::Result<()> {
    fs::create_dir(path)
}

/// The permission bits of a file, never following a link; `None` where there are none.
#[cfg(unix)]
fn mode_bits(path: &Path) -> Option<u32> {
    use std::os::unix::fs::PermissionsExt;
    fs::symlink_metadata(path)
        .ok()
        .map(|m| m.permissions().mode() & 0o777)
}

#[cfg(not(unix))]
fn mode_bits(_path: &Path) -> Option<u32> {
    None
}

/// Create the folders of `rel` (relative to `project`) one by one, refusing a link.
fn make_dirs(project: &Path, rel: &str) -> Result<()> {
    make_dirs_with(project, rel, false)
}

/// [`make_dirs`] for a backup area: every folder it creates is 0700.
fn make_private_dirs(project: &Path, rel: &str) -> Result<()> {
    make_dirs_with(project, rel, true)
}

fn make_dirs_with(project: &Path, rel: &str, private: bool) -> Result<()> {
    let mut cur = project.to_path_buf();
    for part in Path::new(rel).components() {
        cur.push(part);
        match fs::symlink_metadata(&cur) {
            Ok(m) if m.file_type().is_symlink() => {
                return Err(format!("{rel}: a path component is a symlink; refused"))
            }
            Ok(m) if m.is_dir() => {}
            Ok(_) => return Err(format!("{rel}: a path component is not a folder")),
            Err(_) => create_dir_with(&cur, private)
                .map_err(|_| format!("{rel}: folder create failed"))?,
        }
    }
    Ok(())
}

/// `.moai/backups/.gitignore` holding `*`: nothing under the backup area is ever offered to
/// git, whose copies may hold tokens. An existing file is left as it is.
fn keep_backups_out_of_git(project: &Path) -> Result<()> {
    let rel = format!("{MOAI_DIR}/backups/.gitignore");
    refuse_linked(project, &rel)?;
    match fs::symlink_metadata(project.join(&rel)) {
        Ok(_) => Ok(()),
        Err(_) => crate::apply::publish_exclusive(&project.join(&rel), b"*\n"),
    }
}

fn parent_of(rel: &str) -> Result<String> {
    Path::new(rel)
        .parent()
        .and_then(|p| p.to_str())
        .map(str::to_string)
        .ok_or_else(|| format!("{rel}: no parent folder"))
}

/// Move one file without overwriting: hard link, then remove the source.
fn move_exclusive(from: &Path, to: &Path) -> Result<()> {
    fs::hard_link(from, to).map_err(|e| {
        if e.kind() == std::io::ErrorKind::AlreadyExists {
            "move refused: the destination already exists".to_string()
        } else if e.raw_os_error() == Some(18) {
            // EXDEV on Linux and macOS.
            "cross-filesystem move refused; the project and its backup must share one filesystem"
                .to_string()
        } else {
            "move failed (hard link)".to_string()
        }
    })?;
    fs::remove_file(from).map_err(|_| "move failed: source not removed (both copies kept)".into())
}

/// Move one folder with a single rename: the source must be a real folder (not a link),
/// the destination must not exist, and another filesystem is refused (no copy fallback).
fn move_folder(from: &Path, to: &Path) -> Result<()> {
    match fs::symlink_metadata(from) {
        Ok(m) if m.is_dir() => {}
        Ok(_) => return Err("folder move refused: the source is not a real folder".into()),
        Err(_) => return Err("folder move refused: the source is missing".into()),
    }
    if fs::symlink_metadata(to).is_ok() {
        return Err("folder move refused: the destination already exists".into());
    }
    fs::rename(from, to).map_err(|e| {
        if e.raw_os_error() == Some(18) {
            "cross-filesystem move refused; the project and its backup must share one filesystem"
                .to_string()
        } else {
            "folder move failed (rename)".to_string()
        }
    })
}

/// Move an earlier-root entry: a folder by one rename, a file by link-and-remove.
fn move_entry(from: &Path, to: &Path) -> Result<()> {
    match fs::symlink_metadata(from) {
        Ok(m) if m.is_dir() => move_folder(from, to),
        Ok(m) if m.is_file() => move_exclusive(from, to),
        Ok(_) => Err("move refused: neither a real folder nor a regular file".into()),
        Err(_) => Err("move refused: the source is missing".into()),
    }
}

/// The marker written last (decision D1): what `state_root` accepts, plus where the Go
/// folder went.
fn marker_bytes(go_backup: &str) -> Result<Vec<u8>> {
    Ok(crate::state_root::marker_bytes(
        "go-state-apply",
        &[("go_backup", go_backup)],
    ))
}

/// Apply the preview whose digest the caller confirmed. Nothing is written unless a fresh
/// preview has exactly that digest.
pub fn apply_with_stamp(
    project: &Path,
    plan_digest: &str,
    permission_reference: &str,
    stamp: &str,
) -> Result<Receipt> {
    let fresh = plan(project)?;
    if fresh.plan_digest != plan_digest {
        return Err("stale or unconfirmed plan: the digest does not match a fresh go-state-plan; nothing written".into());
    }
    if permission_reference.trim().is_empty() {
        return Err("a permission reference is required".into());
    }
    if !valid_stamp(stamp) {
        return Err("invalid backup stamp".into());
    }
    let project = PathBuf::from(&fresh.project);
    let planned: Vec<&Item> = fresh.items.iter().filter(|i| i.action == "move").collect();
    let mut receipt = Receipt {
        schema_version: SCHEMA_VERSION,
        project: fresh.project.clone(),
        backup: stamp.to_string(),
        plan_digest: fresh.plan_digest.clone(),
        permission_reference: permission_reference.to_string(),
        state: "unchanged".into(),
        moves: Vec::new(),
        restored: Vec::new(),
        go_backup: String::new(),
        copied: Vec::new(),
        marker: String::new(),
        rust_aside: String::new(),
        legacy_move: false,
    };
    if planned.is_empty() {
        return Ok(receipt);
    }
    if let Some(c) = fresh
        .restores
        .iter()
        .find(|r| r.action.starts_with("conflict"))
    {
        return Err(if c.action == "conflict_link" {
            format!(
                "{}: a path component is a symlink; refused; nothing written",
                c.to
            )
        } else {
            format!(
                "restore conflict: {} already exists and is never overwritten; nothing written",
                c.to
            )
        });
    }
    if planned.iter().any(|i| i.path == LEGACY_ROOT) {
        return apply_legacy_move(&project, &fresh, receipt, stamp);
    }
    for item in planned.iter().filter(|i| i.path != MOAI_DIR) {
        receipt.moves.push(planned_move(&project, item)?);
    }
    if planned.iter().any(|i| i.path == MOAI_DIR) {
        return apply_moving_folder(&project, &fresh, receipt, stamp);
    }
    // The policy plans the #2936 kinds only with the Go folder; the stage-2 layout moves
    // unchanged files alone.
    if receipt.moves.iter().any(|m| !m.kind.is_empty()) {
        return Err(
            "a staying-folder plan with a link, copy folder or edited file; refused".into(),
        );
    }
    // The Go folder stays: the #2561 stage-2 layout under the current state folder.
    let base = crate::state_root::dir(&project);
    let backup_rel = format!("{base}/backups/{stamp}");
    // `make_dirs` refuses a linked state folder or `backups`; the stamp folder itself is
    // created exclusively, so a link already at that name fails the create.
    make_private_dirs(&project, &format!("{base}/backups"))?;
    create_dir_with(&project.join(&backup_rel), true).map_err(|e| {
        if e.kind() == std::io::ErrorKind::AlreadyExists {
            "exclusive create refused: backup already exists".to_string()
        } else {
            "backup folder create failed".to_string()
        }
    })?;
    receipt.state = "pending".into();
    let mut saved = save_receipt(&project, base, &receipt, None)?;
    for i in 0..receipt.moves.len() {
        let rel = receipt.moves[i].path.clone();
        refuse_linked(&project, &rel)?;
        let dest_rel = format!("{backup_rel}/go-state/{rel}");
        make_private_dirs(&project, &parent_of(&dest_rel)?)?;
        move_exclusive(&project.join(&rel), &project.join(&dest_rel))?;
        receipt.moves[i].moved = true;
        saved = save_receipt(&project, base, &receipt, Some(&saved))?;
    }
    receipt.state = "committed".into();
    save_receipt(&project, base, &receipt, Some(&saved))?;
    Ok(receipt)
}

/// The #2696 layout (see the section comment): steps 1 to 7.
fn apply_moving_folder(
    project: &Path,
    fresh: &GoStatePlan,
    mut receipt: Receipt,
    stamp: &str,
) -> Result<Receipt> {
    let planned_backup = fresh
        .go_backup
        .as_deref()
        .ok_or("a moving plan without a Go backup name")?;
    let go_backup = if planned_backup == GO_BACKUP {
        GO_BACKUP.to_string()
    } else {
        format!("{GO_BACKUP}-{stamp}")
    };
    let area = format!("{MOAI_DIR}/backups/{stamp}");
    receipt.go_backup = go_backup.clone();
    receipt.moves.insert(
        0,
        Moved {
            path: MOAI_DIR.into(),
            kind: "folder".into(),
            ..Moved::default()
        },
    );
    for s in fresh.state_moves.iter().filter(|s| s.action == "move") {
        receipt.moves.push(Moved {
            path: s.from.clone(),
            kind: "state".into(),
            ..Moved::default()
        });
    }
    for f in fresh.root_files.iter().filter(|f| f.action == "move") {
        receipt.moves.push(Moved {
            path: f.path.clone(),
            sha256: f.sha256.clone(),
            kind: "root_file".into(),
            ..Moved::default()
        });
    }
    receipt.state = "pending".into();
    // 1. One rename; the backup name is never overwritten.
    move_folder(&project.join(MOAI_DIR), &project.join(&go_backup))?;
    receipt.moves[0].moved = true;
    // 2. The new folder and the pending receipt. Should either fail, undo the rename now:
    //    nothing else has changed yet.
    let made = fs::create_dir(project.join(MOAI_DIR))
        .map_err(|_| format!("{MOAI_DIR}: folder create failed"))
        .and_then(|_| make_private_dirs(project, &area))
        .and_then(|_| crate::update::test_fault("go_state_after_rename"))
        .and_then(|_| save_receipt(project, MOAI_DIR, &receipt, None));
    let mut saved = match made {
        Ok(bytes) => bytes,
        Err(e) => {
            // Only empty folders this step made; `remove_dir` keeps anything else.
            for rel in [area.clone(), format!("{MOAI_DIR}/backups"), MOAI_DIR.into()] {
                let _ = fs::remove_dir(project.join(rel));
            }
            return Err(
                match move_folder(&project.join(&go_backup), &project.join(MOAI_DIR)) {
                    Ok(()) => format!("{e}; the Go folder was renamed back, nothing else changed"),
                    Err(back) => format!(
                        "{e}; renaming {go_backup} back to {MOAI_DIR} also failed ({back}); \
                         nothing else changed"
                    ),
                },
            );
        }
    };
    keep_backups_out_of_git(project)?;
    receipt.moves[0].sha256 = crate::baseline::tree_digest(&project.join(&go_backup), &[])?;
    saved = save_receipt(project, MOAI_DIR, &receipt, Some(&saved))?;
    // 3. The proven `.claude/` and `.agents/` entries: files, the mirror links (as links) and
    //    the mirror copy folders (one rename each).
    for i in 0..receipt.moves.len() {
        let kind = receipt.moves[i].kind.clone();
        if !matches!(kind.as_str(), "" | "edited" | "link" | "mirror_copy") {
            continue;
        }
        let rel = receipt.moves[i].path.clone();
        let (from, to) = (
            project.join(&rel),
            project.join(format!("{area}/go-state/{rel}")),
        );
        if kind == "link" {
            refuse_linked(project, &parent_of(&rel)?)?;
        } else {
            refuse_linked(project, &rel)?;
        }
        make_private_dirs(project, &parent_of(&format!("{area}/go-state/{rel}"))?)?;
        match kind.as_str() {
            "link" => move_link(&from, &to, &receipt.moves[i].target)?,
            "mirror_copy" => {
                if crate::baseline::tree_digest(&from, &[])? != receipt.moves[i].sha256 {
                    return Err(format!("{rel} changed after the preview; move stopped"));
                }
                move_folder(&from, &to)?
            }
            _ => move_exclusive(&from, &to)?,
        }
        receipt.moves[i].moved = true;
        saved = save_receipt(project, MOAI_DIR, &receipt, Some(&saved))?;
    }
    // 4. The earlier root's product state.
    saved = move_state_entries(project, &mut receipt, saved)?;
    // 5. SPECs from the backup (the original stays there), never overwriting.
    for r in fresh.restores.iter().filter(|r| r.action == "copy") {
        let inner = r
            .from
            .strip_prefix(&format!("{planned_backup}/"))
            .ok_or("restore source outside the Go backup")?;
        let src_rel = format!("{go_backup}/{inner}");
        refuse_linked(project, &src_rel)?;
        let bytes = fs::read(project.join(&src_rel)).map_err(|_| "restore read failed")?;
        if digest(&bytes) != r.sha256 {
            return Err(format!(
                "{} changed after the preview; restore stopped",
                r.from
            ));
        }
        make_dirs(project, &parent_of(&r.to)?)?;
        // The restored copy keeps the mode of the file in the Go backup.
        let mode = mode_bits(&project.join(&src_rel));
        crate::apply::publish_exclusive_mode(&project.join(&r.to), &bytes, mode)?;
        receipt.restored.push(Restored {
            path: r.to.clone(),
            sha256: r.sha256.clone(),
        });
        saved = save_receipt(project, MOAI_DIR, &receipt, Some(&saved))?;
    }
    // 6. The root instruction files move into `root-files/` (O4); the local settings stay
    //    and a copy goes there (O3).
    for f in fresh.root_files.iter().filter(|f| f.action != "leave") {
        refuse_linked(project, &f.path)?;
        match fs::symlink_metadata(project.join(&f.path)) {
            Ok(m) if m.is_file() && m.len() <= FILE_LIMIT => {}
            _ => return Err(format!("{}: not a bounded regular file; stopped", f.path)),
        }
        let bytes = fs::read(project.join(&f.path)).map_err(|_| "root file read failed")?;
        if digest(&bytes) != f.sha256 {
            return Err(format!("{} changed after the preview; stopped", f.path));
        }
        let to = format!("{area}/root-files/{}", f.path);
        make_private_dirs(project, &parent_of(&to)?)?;
        if f.action == "move" {
            let i = receipt
                .moves
                .iter()
                .position(|m| m.kind == "root_file" && m.path == f.path)
                .ok_or("a root file move the receipt does not hold")?;
            move_exclusive(&project.join(&f.path), &project.join(&to))?;
            receipt.moves[i].moved = true;
        } else {
            // The copy of the local settings may hold tokens: owner-only, whatever the
            // original's mode (the original stays as it is).
            crate::apply::publish_exclusive_mode(&project.join(&to), &bytes, Some(PRIVATE_FILE))?;
            receipt.copied.push(Restored {
                path: to,
                sha256: f.sha256.clone(),
            });
        }
        saved = save_receipt(project, MOAI_DIR, &receipt, Some(&saved))?;
    }
    // 7. The marker, last: from here `state_root` says `rust`.
    let marker = marker_bytes(&go_backup)?;
    finish_with_marker(project, receipt, saved, &marker)
}

/// Step 4 of both moving layouts: each recorded `.moai-cli/<entry>` into `.moai/<entry>`.
fn move_state_entries(
    project: &Path,
    receipt: &mut Receipt,
    mut saved: Vec<u8>,
) -> Result<Vec<u8>> {
    for i in 0..receipt.moves.len() {
        if receipt.moves[i].kind != "state" {
            continue;
        }
        let from = receipt.moves[i].path.clone();
        let entry = from
            .strip_prefix(&format!("{LEGACY_ROOT}/"))
            .ok_or("state entry outside the earlier root")?;
        refuse_linked(project, &from)?;
        move_entry(&project.join(&from), &project.join(MOAI_DIR).join(entry))
            .map_err(|e| format!("{from}: {e}"))?;
        receipt.moves[i].moved = true;
        saved = save_receipt(project, MOAI_DIR, receipt, Some(&saved))?;
    }
    Ok(saved)
}

/// The last step of both moving layouts: publish the marker, record it, commit.
fn finish_with_marker(
    project: &Path,
    mut receipt: Receipt,
    saved: Vec<u8>,
    marker: &[u8],
) -> Result<Receipt> {
    crate::apply::publish_exclusive(
        &project.join(MOAI_DIR).join(crate::state_root::RUST_MARKER),
        marker,
    )?;
    receipt.marker = digest(marker);
    let saved = save_receipt(project, MOAI_DIR, &receipt, Some(&saved))?;
    receipt.state = "committed".into();
    save_receipt(project, MOAI_DIR, &receipt, Some(&saved))?;
    Ok(receipt)
}

/// #2696 C4a: the earlier `.moai-cli/` state alone into a new `.moai/` — the moving layout
/// without a Go folder: new `.moai/` with the pending receipt, the state entries, the marker
/// last. A failure before the receipt removes the empty folders it made.
fn apply_legacy_move(
    project: &Path,
    fresh: &GoStatePlan,
    mut receipt: Receipt,
    stamp: &str,
) -> Result<Receipt> {
    let area = format!("{MOAI_DIR}/backups/{stamp}");
    receipt.legacy_move = true;
    // A C4a receipt carries no Go backup; one that also names a Go backup fails the
    // folder-entry check in `read_receipt` (a Go backup always goes with a folder move).
    for s in fresh.state_moves.iter().filter(|s| s.action == "move") {
        receipt.moves.push(Moved {
            path: s.from.clone(),
            kind: "state".into(),
            ..Moved::default()
        });
    }
    receipt.state = "pending".into();
    let made = fs::create_dir(project.join(MOAI_DIR))
        .map_err(|_| format!("{MOAI_DIR}: folder create failed"))
        .and_then(|_| make_private_dirs(project, &area))
        .and_then(|_| crate::update::test_fault("go_state_legacy_after_mkdir"))
        .and_then(|_| save_receipt(project, MOAI_DIR, &receipt, None));
    let saved = match made {
        Ok(bytes) => bytes,
        Err(e) => {
            for rel in [area.clone(), format!("{MOAI_DIR}/backups"), MOAI_DIR.into()] {
                let _ = fs::remove_dir(project.join(rel));
            }
            return Err(format!("{e}; nothing else changed"));
        }
    };
    keep_backups_out_of_git(project)?;
    let saved = move_state_entries(project, &mut receipt, saved)?;
    let marker = crate::state_root::marker_bytes("legacy-move", &[]);
    finish_with_marker(project, receipt, saved, &marker)
}

pub fn apply(project: &Path, plan_digest: &str, permission_reference: &str) -> Result<Receipt> {
    apply_with_stamp(project, plan_digest, permission_reference, &stamp_now())
}

#[derive(Clone, Debug, Serialize)]
pub struct RevertItem {
    pub path: String,
    /// `restore`, `conflict` (something is at the original path again), `missing` (the
    /// backup copy is gone or changed), `skip` (never moved), `remove_restored` /
    /// `keep_changed` (a SPEC copy), or `remove` (the marker, #2696).
    pub action: String,
}

#[derive(Clone, Debug, Serialize)]
pub struct RevertPlan {
    pub schema_version: u32,
    pub project: String,
    pub backup: String,
    pub receipt_state: String,
    pub items: Vec<RevertItem>,
    pub plan_digest: String,
}

/// `.moai.rust-<UTC>`, the only name a revert sets this generation's `.moai/` aside under.
fn rust_aside_name(name: &str) -> bool {
    name.strip_prefix(RUST_ASIDE).is_some_and(valid_stamp)
}

/// Where the receipt of apply `stamp` lives: `.moai` (an apply that moved the Go folder),
/// the earlier root (one that left it in place), or a `.moai.rust-<UTC>` a revert set aside.
/// The first candidate with anything at the receipt path (or a link on the way) is read.
fn find_receipt_base(project: &Path, stamp: &str) -> Option<String> {
    let mut asides: Vec<String> = fs::read_dir(project)
        .map(|rd| {
            rd.filter_map(|e| e.ok())
                .filter_map(|e| e.file_name().into_string().ok())
                .filter(|n| rust_aside_name(n))
                .collect()
        })
        .unwrap_or_default();
    asides.sort();
    let mut bases = vec![MOAI_DIR.to_string(), LEGACY_ROOT.to_string()];
    bases.extend(asides);
    bases.into_iter().find(|base| {
        let rel = receipt_rel(base, stamp);
        linked_component(project, &rel) || fs::symlink_metadata(project.join(rel)).is_ok()
    })
}

fn top_of(path: &str) -> &str {
    Path::new(path)
        .components()
        .next()
        .and_then(|c| c.as_os_str().to_str())
        .unwrap_or("")
}

fn read_receipt(project: &Path, backup: &str) -> Result<(Receipt, Vec<u8>, String)> {
    if !valid_stamp(backup) {
        return Err("invalid backup name: expected the apply's UTC stamp YYYYMMDDTHHMMSSZ".into());
    }
    let base = find_receipt_base(project, backup)
        .ok_or_else(|| format!("no go-state receipt for {backup}"))?;
    let rel = receipt_rel(&base, backup);
    let text = read_bounded_text(project, &rel, MANIFEST_LIMIT)?
        .ok_or_else(|| format!("{rel}: no go-state receipt"))?;
    let receipt: Receipt =
        serde_json::from_str(&text).map_err(|_| format!("{rel}: unreadable receipt"))?;
    if receipt.backup != backup || receipt.project != project.to_string_lossy() {
        return Err("receipt scope mismatch".into());
    }
    // The receipt is not its own authority: it may only name what an apply can do, and only
    // from the place that layout keeps it.
    let go_folder = !receipt.go_backup.is_empty();
    // Both #2696 layouts keep the receipt under `.moai`: the Go-folder move and the
    // earlier-root move (C4a); never both at once.
    let moving = go_folder || receipt.legacy_move;
    if moving != (base != LEGACY_ROOT)
        || (go_folder && !crate::state_root::go_backup_name(&receipt.go_backup))
    {
        return Err("receipt layout does not match where it was found; refused".into());
    }
    if !(receipt.rust_aside.is_empty() || moving && rust_aside_name(&receipt.rust_aside))
        || (base != MOAI_DIR && base != LEGACY_ROOT && base != receipt.rust_aside)
    {
        return Err("receipt names an aside folder it was not found in; refused".into());
    }
    if receipt.moves.iter().any(|m| !plain_relative(&m.path)) {
        return Err("receipt holds a path that is not plain relative; refused".into());
    }
    if receipt.moves.iter().any(|m| {
        m.kind != "state"
            && m.kind != "root_file"
            && (!ROOTS.contains(&top_of(&m.path)) || m.path == MANIFEST)
    }) {
        return Err("receipt holds a path outside .moai/, .claude/ and .agents/; refused".into());
    }
    // A folder entry is exactly `.moai`; a file entry is never the folder itself; a state
    // entry is one known earlier-root entry, and only in the moving layout. The #2936 kinds
    // only go with a Go folder move: an edited file, a mirror entry exactly
    // `.agents/skills/<n>`, a root file one of [`ROOT_FILES`].
    let mirror_entry = |p: &str| {
        p.strip_prefix(&format!("{AGENTS_SKILLS}/"))
            .is_some_and(|n| !n.is_empty() && !n.contains('/'))
    };
    if receipt.moves.iter().any(|m| match m.kind.as_str() {
        "folder" => m.path != MOAI_DIR,
        "" => m.path == MOAI_DIR,
        "edited" => !go_folder || m.path == MOAI_DIR,
        "link" | "mirror_copy" => !go_folder || !mirror_entry(&m.path),
        "root_file" => !go_folder || !ROOT_FILES.contains(&m.path.as_str()),
        "state" => {
            !moving
                || !m
                    .path
                    .strip_prefix(&format!("{LEGACY_ROOT}/"))
                    .is_some_and(|e| STATE_ENTRIES.contains(&e))
        }
        _ => true,
    }) {
        return Err("receipt holds an entry of the wrong kind; refused".into());
    }
    // Under `.moai` a folder move goes with a Go backup name and only then; under the earlier
    // root a folder entry is the #2561 stage-3 layout (`.moai-cli/backups/<UTC>/go-state/.moai`).
    if moving && go_folder != receipt.moves.iter().any(|m| m.kind == "folder") {
        return Err(
            "receipt has a folder move without a Go backup, or the reverse; refused".into(),
        );
    }
    // Restored copies land in a user-data folder of the new `.moai/` (#2936 O1), or in the
    // earlier root's `specs/` (the #2561 layouts).
    let allowed: Vec<String> = if moving {
        USER_DATA_FROM
            .iter()
            .map(|f| format!("{MOAI_DIR}/{f}/"))
            .collect()
    } else {
        vec![format!("{LEGACY_ROOT}/specs/")]
    };
    if receipt.restored.iter().any(|r| {
        !plain_relative(&r.path) || !allowed.iter().any(|a| r.path.starts_with(a.as_str()))
    }) {
        return Err(format!(
            "receipt holds a restored path outside {}; refused",
            allowed.join(", ")
        ));
    }
    let copies = format!("{MOAI_DIR}/backups/{backup}/root-files/");
    if receipt
        .copied
        .iter()
        .any(|c| !go_folder || !plain_relative(&c.path) || !c.path.starts_with(&copies))
    {
        return Err(format!("receipt holds a copy outside {copies}; refused"));
    }
    Ok((receipt, text.into_bytes(), base))
}

/// Where the apply put one moved file or mirror entry, under the backup area `area`: a root
/// file in `root-files/`, everything else in `go-state/` at its own path.
fn backup_rel_of(area: &str, m: &Moved) -> String {
    if m.kind == "root_file" {
        format!("{area}/root-files/{}", m.path)
    } else {
        format!("{area}/go-state/{}", m.path)
    }
}

/// A real folder at `rel`, reached without a link.
fn real_dir(project: &Path, rel: &str) -> bool {
    !linked_component(project, rel)
        && fs::symlink_metadata(project.join(rel))
            .map(|m| m.is_dir())
            .unwrap_or(false)
}

/// Read-only: what a revert of one apply would do.
pub fn revert_plan(project: &Path, backup: &str) -> Result<RevertPlan> {
    let project = fs::canonicalize(project).map_err(|_| "project folder unavailable")?;
    let (receipt, _, base) = read_receipt(&project, backup)?;
    let moving = !receipt.go_backup.is_empty();
    // Either #2696 layout: the marker and this generation's `.moai/` are undone the same way.
    let new_layout = moving || receipt.legacy_move;
    // This generation's `.moai/` is still in place (not yet set aside by a revert).
    let at_moai = base == MOAI_DIR;
    let area = format!("{base}/backups/{backup}");
    let mut items = Vec::new();
    if new_layout {
        let marker = format!("{MOAI_DIR}/{}", crate::state_root::RUST_MARKER);
        let action = if !at_moai {
            "skip"
        } else {
            match fs::symlink_metadata(project.join(&marker)) {
                Err(_) => "skip",
                Ok(m) if m.is_file() => "remove",
                Ok(_) => "conflict",
            }
        };
        items.push(RevertItem {
            path: marker,
            action: action.into(),
        });
    }
    for r in &receipt.restored {
        // Read where they were written. Once this generation's `.moai/` is set aside, `.moai`
        // is either missing (the read gives `skip`) or something that came back, which the
        // folder item below reports as `conflict`, so the revert stops before any removal.
        let action = if linked_component(&project, &r.path) {
            "keep_changed"
        } else {
            match sha256_file(&project.join(&r.path)) {
                Ok(h) if h == r.sha256 => "remove_restored",
                Ok(_) => "keep_changed",
                Err(_) => "skip",
            }
        };
        items.push(RevertItem {
            path: r.path.clone(),
            action: action.into(),
        });
    }
    for m in &receipt.moves {
        let action = match m.kind.as_str() {
            "folder" if moving => {
                let backup_ok = real_dir(&project, &receipt.go_backup);
                if !at_moai && fs::symlink_metadata(project.join(MOAI_DIR)).is_ok() {
                    // Set aside already, and something is at `.moai` again: never merged.
                    "conflict"
                } else if !backup_ok {
                    if m.moved {
                        "missing"
                    } else {
                        "skip"
                    }
                } else if m.sha256.is_empty() {
                    // Renamed but the receipt write did not follow: the disk says where it is.
                    "restore"
                } else {
                    match crate::baseline::tree_digest(&project.join(&receipt.go_backup), &[]) {
                        Ok(d) if d == m.sha256 => "restore",
                        _ => "missing",
                    }
                }
            }
            "folder" => {
                let dest_rel = format!("{area}/go-state/{MOAI_DIR}");
                let source_present = fs::symlink_metadata(project.join(MOAI_DIR)).is_ok();
                if source_present {
                    // Never merged: `.moai/` came back (a live hook) or was never moved.
                    if m.moved {
                        "conflict"
                    } else {
                        "skip"
                    }
                } else if !real_dir(&project, &dest_rel) {
                    "missing"
                } else if m.sha256.is_empty() {
                    "restore"
                } else {
                    match crate::baseline::tree_digest(&project.join(&dest_rel), &[]) {
                        Ok(d) if d == m.sha256 => "restore",
                        _ => "missing",
                    }
                }
            }
            "state" => {
                let entry = m
                    .path
                    .strip_prefix(&format!("{LEGACY_ROOT}/"))
                    .unwrap_or_default();
                let src = format!("{base}/{entry}");
                let back = fs::symlink_metadata(project.join(&m.path)).is_ok()
                    || linked_component(&project, &m.path);
                let here = !linked_component(&project, &src)
                    && fs::symlink_metadata(project.join(&src)).is_ok();
                if back {
                    if m.moved {
                        "conflict"
                    } else {
                        "skip"
                    }
                } else if here {
                    // Also when the receipt missed the move: the disk says where it is.
                    "restore"
                } else if m.moved {
                    "missing"
                } else {
                    "skip"
                }
            }
            _ => {
                if !m.moved {
                    "skip"
                } else if fs::symlink_metadata(project.join(&m.path)).is_ok()
                    || linked_component(&project, &m.path)
                {
                    "conflict"
                } else {
                    let dest_rel = backup_rel_of(&area, m);
                    let dest = project.join(&dest_rel);
                    match m.kind.as_str() {
                        // The backup link must still carry the body the apply moved.
                        "link" => {
                            let body = fs::read_link(&dest).ok();
                            if parent_of(&dest_rel).is_ok_and(|p| !linked_component(&project, &p))
                                && body.as_deref().and_then(Path::to_str) == Some(m.target.as_str())
                                && digest(m.target.as_bytes()) == m.sha256
                            {
                                "restore"
                            } else {
                                "missing"
                            }
                        }
                        "mirror_copy" => {
                            if real_dir(&project, &dest_rel)
                                && crate::baseline::tree_digest(&dest, &[])
                                    .is_ok_and(|d| d == m.sha256)
                            {
                                "restore"
                            } else {
                                "missing"
                            }
                        }
                        _ if linked_component(&project, &dest_rel) => "missing",
                        _ => match sha256_file(&dest) {
                            Ok(h) if h == m.sha256 => "restore",
                            _ => "missing",
                        },
                    }
                }
            }
        };
        items.push(RevertItem {
            path: m.path.clone(),
            action: action.into(),
        });
    }
    // C4a: without a Go folder to bring back, this generation's `.moai/` is set aside last.
    if receipt.legacy_move {
        let action = if !at_moai {
            "skip"
        } else if real_dir(&project, MOAI_DIR) {
            "set_aside"
        } else {
            "missing"
        };
        items.push(RevertItem {
            path: MOAI_DIR.into(),
            action: action.into(),
        });
    }
    let mut plan = RevertPlan {
        schema_version: SCHEMA_VERSION,
        project: receipt.project.clone(),
        backup: backup.to_string(),
        receipt_state: receipt.state.clone(),
        items,
        plan_digest: String::new(),
    };
    let bytes = serde_json::to_vec(&plan).map_err(|_| "encode failed")?;
    plan.plan_digest = format!("sha256:{:x}", Sha256::digest(bytes));
    Ok(plan)
}

/// Undo one apply, all or nothing up front: any `conflict` or `missing` item refuses the whole
/// revert before the first write. Never overwrites, never deletes what the apply did not write.
/// For the #2696 layout, in this order: remove the marker (from then on `state_root` no longer
/// says `rust`), remove unchanged SPEC copies, move the earlier-root entries and the `.claude/`
/// files back, set this generation's `.moai/` aside as `.moai.rust-<UTC>` (kept, with the
/// receipt and the root-file copies), and rename the Go backup back to `.moai`. Each step is
/// recorded, so running the revert again after a failure continues where it stopped.
pub fn revert(
    project: &Path,
    backup: &str,
    plan_digest: &str,
    permission_reference: &str,
) -> Result<RevertPlan> {
    let fresh = revert_plan(project, backup)?;
    if fresh.plan_digest != plan_digest {
        return Err("stale or unconfirmed revert plan: the digest does not match a fresh go-state-revert-plan; nothing written".into());
    }
    if permission_reference.trim().is_empty() {
        return Err("a permission reference is required".into());
    }
    if fresh.receipt_state == "reverted" {
        return Err("this apply was already reverted".into());
    }
    if let Some(bad) = fresh
        .items
        .iter()
        .find(|i| i.action == "conflict" || i.action == "missing")
    {
        return Err(format!(
            "revert refused: {} is {}; nothing written",
            bad.path, bad.action
        ));
    }
    let project = PathBuf::from(&fresh.project);
    let (mut receipt, mut saved, mut base) = read_receipt(&project, backup)?;
    let moving = !receipt.go_backup.is_empty();
    for item in fresh.items.iter().filter(|i| i.action == "remove") {
        refuse_linked(&project, &item.path)?;
        fs::remove_file(project.join(&item.path)).map_err(|_| "marker not removed")?;
        receipt.marker.clear();
        saved = save_receipt(&project, &base, &receipt, Some(&saved))?;
    }
    // Unchanged restored copies go; a changed one stays and is reported (`keep_changed`).
    for item in fresh.items.iter().filter(|i| i.action == "remove_restored") {
        refuse_linked(&project, &item.path)?;
        fs::remove_file(project.join(&item.path)).map_err(|_| "restored copy not removed")?;
        receipt.restored.retain(|r| r.path != item.path);
        saved = save_receipt(&project, &base, &receipt, Some(&saved))?;
    }
    let entry_of = |path: &str| {
        receipt
            .moves
            .iter()
            .find(|m| m.path == path)
            .cloned()
            .unwrap_or_default()
    };
    let restores: Vec<(String, Moved)> = fresh
        .items
        .iter()
        .filter(|i| i.action == "restore")
        .map(|i| (i.path.clone(), entry_of(&i.path)))
        .collect();
    // Entries and files first, the folder last.
    for (path, entry) in restores.iter().filter(|(_, m)| m.kind != "folder") {
        let kind = &entry.kind;
        if kind == "state" {
            let entry = path
                .strip_prefix(&format!("{LEGACY_ROOT}/"))
                .ok_or("state entry outside the earlier root")?;
            make_dirs(&project, LEGACY_ROOT)?;
            let src = format!("{base}/{entry}");
            refuse_linked(&project, &src)?;
            move_entry(&project.join(&src), &project.join(path))
                .map_err(|e| format!("{path}: {e}"))?;
        } else {
            let parent = parent_of(path).unwrap_or_default();
            if !parent.is_empty() {
                make_dirs(&project, &parent)?;
            }
            let src = project.join(backup_rel_of(&format!("{base}/backups/{backup}"), entry));
            match kind.as_str() {
                "link" => move_link(&src, &project.join(path), &entry.target)?,
                "mirror_copy" => move_folder(&src, &project.join(path))?,
                _ => move_exclusive(&src, &project.join(path))?,
            }
        }
        if let Some(m) = receipt.moves.iter_mut().find(|m| &m.path == path) {
            m.moved = false;
        }
        saved = save_receipt(&project, &base, &receipt, Some(&saved))?;
    }
    if restores.iter().any(|(_, m)| m.kind == "folder") {
        if moving {
            if base == MOAI_DIR {
                if receipt.rust_aside.is_empty() {
                    receipt.rust_aside = format!("{RUST_ASIDE}{}", stamp_now());
                    saved = save_receipt(&project, &base, &receipt, Some(&saved))?;
                }
                move_folder(&project.join(MOAI_DIR), &project.join(&receipt.rust_aside))?;
                base = receipt.rust_aside.clone();
            }
            move_folder(&project.join(&receipt.go_backup), &project.join(MOAI_DIR))?;
        } else {
            let src = format!("{base}/backups/{backup}/go-state/{MOAI_DIR}");
            move_folder(&project.join(src), &project.join(MOAI_DIR))?;
        }
        if let Some(m) = receipt.moves.iter_mut().find(|m| m.kind == "folder") {
            m.moved = false;
        }
        saved = save_receipt(&project, &base, &receipt, Some(&saved))?;
    }
    if fresh.items.iter().any(|i| i.action == "set_aside") {
        if receipt.rust_aside.is_empty() {
            receipt.rust_aside = format!("{RUST_ASIDE}{}", stamp_now());
            saved = save_receipt(&project, &base, &receipt, Some(&saved))?;
        }
        move_folder(&project.join(MOAI_DIR), &project.join(&receipt.rust_aside))?;
        base = receipt.rust_aside.clone();
    }
    receipt.state = "reverted".into();
    save_receipt(&project, &base, &receipt, Some(&saved))?;
    Ok(fresh)
}
