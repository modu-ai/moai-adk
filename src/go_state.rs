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
//! `.moai/moai.json` (the name every writer writes; readers also accept the earlier
//! `moai-cli.json`, #2871) is written last. The stage-2 section below has the layout.
//!
//! #2936 S1c-1 (PLAN 6878897673): the preview also classifies the Go generation's `.codex/`
//! files ([`classify_codex`]) and moves none of them: every `.codex` item is `leave`. A
//! `.codex` path is never one the receipts may name ([`ROOTS`] does not list it).

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
/// The one size limit of a go-state receipt: [`save_receipt`] refuses a receipt above it and
/// [`read_receipt`] reads up to it. The 256 KiB `apply::RECEIPT_LIMIT` that every other target
/// keeps does not apply: a project with 1,319 restored SPEC copies has a receipt of 262,362
/// bytes, and the receipt grows with every copy the apply makes.
const RECEIPT_MAX: u64 = crate::state_root::GO_STATE_RECEIPT_MAX;
const FILE_LIMIT: u64 = 64 * 1024 * 1024;
/// #2936 S1b: `.agents/` holds the Go skill mirror and the published command skills.
const ROOTS: [&str; 3] = [".moai", ".claude", ".agents"];
/// #2936 S1c-1: the folders the `unrecorded` count walks. `.codex` is counted but is not in
/// [`ROOTS`], so a receipt still refuses a `.codex` path (nothing under it moves yet).
const COUNTED_ROOTS: [&str; 4] = [".moai", ".claude", ".agents", ".codex"];
/// #2936 S1c: the Go generation's Codex wiring files (`internal/codexwiring`), and the folders
/// the deployer fills under `.codex/`.
const CODEX_DIR: &str = ".codex";
const CODEX_HOOKS: &str = ".codex/hooks.json";
const CODEX_CONFIG: &str = ".codex/config.toml";
const CODEX_GO_AGENTS: &str = ".codex/agents/moai/";
const CODEX_GO_HOOK_SCRIPTS: &str = ".codex/hooks/moai/";
/// The Go wiring's journal (`JournalRelPath`) and the prefix of its temp files (`tempPrefix`).
const GO_WIRING_JOURNAL: &str = ".moai/state/codex-wiring-journal.json";
const WIRING_TEMP: &str = ".codexwiring-";
/// The reason of an item that S1c-2 will move; S1c-1 only classifies it.
const MOVES_LATER: &str = "classified; moves in S1c-2";
/// The size bound of a Codex config file.
const CODEX_READ_LIMIT: u64 = crate::hosts::config_backup::CONFIG_LIMIT;
/// The handler commands the Go wiring owns (`moaiHandlerPrefix`, `isMoaiNamespaceCommand`).
const GO_HANDLER_PREFIX: &str = "moai hook ";
const GO_HANDLER_NAMESPACE: &str = "/.codex/hooks/moai/";
/// The values of the Go `[mcp_servers.moai]` table (`configtoml.go`).
const GO_MCP_ARGS: [&str; 1] = ["mcp-server"];
const GO_MCP_ENV_VARS: [&str; 9] = [
    "MOAI_HOME",
    "MOAI_KANBAN_ID",
    "MOAI_SESSION_PID",
    "MOAI_KANBAN_BACKEND",
    "MOAI_FACTORY_ROLE",
    "MOAI_FACTORY_WORKER",
    "MOAI_FACTORY_WORKERS",
    "CLAUDE_PROJECT_DIR",
    "CLAUDE_CODE_SESSION_ID",
];
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
/// builds take `MOAI_TEST_MIRROR_FILE_LIMIT` so a test can reach it with a few files.
fn mirror_file_limit() -> u64 {
    #[cfg(debug_assertions)]
    if let Some(n) = std::env::var("MOAI_TEST_MIRROR_FILE_LIMIT")
        .ok()
        .and_then(|v| v.parse().ok())
    {
        return n;
    }
    10_000
}

/// #2936 O6: a name that holds secrets by convention. Such an entry under `.claude/` or
/// `.agents/` is never read, hashed, copied or moved; the preview only counts it. The home tidy
/// operations protect by the same rule (`tidy::rules::class_of_name`).
pub(crate) fn secret_name(name: &str) -> bool {
    crate::home_state::is_secret(name) || crate::worktree_include::secret_name(name)
}

pub(crate) fn has_secret_component(rel: &str) -> bool {
    Path::new(rel)
        .components()
        .any(|c| c.as_os_str().to_str().is_some_and(secret_name))
}
/// Decision 1 (#2561 comment 6831345102): `.moai/` moves as a whole folder.
const MOAI_DIR: &str = ".moai";
/// Entries the `unrecorded` count reads before it stops (#2561 F1: a project's `.moai/` held
/// unpacked review trees with `node_modules`, and the preview read for over 14 minutes). The
/// count only informs; nothing moves on its account. Debug builds take
/// `MOAI_TEST_UNRECORDED_LIMIT` so a test can reach the limit with a few files.
pub(crate) fn unrecorded_limit() -> u64 {
    #[cfg(debug_assertions)]
    if let Some(n) = std::env::var("MOAI_TEST_UNRECORDED_LIMIT")
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
const STATE_ENTRIES: [&str; 18] = [
    "archive",
    "config",
    "decisions",
    "evidence",
    "forge",
    "imports",
    "index",
    // `lane_loop` writes `lane-stop-hook/<lane>.json` through `state_root::path` (#2871, leader
    // decision 6887464608): it is product state of the earlier root like its siblings.
    "lane-stop-hook",
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
/// O4); the install step writes a new `CLAUDE.md`, and the CLAUDE files take the place of the
/// older two. `CLAUDE.local.md` is the user's and stays.
const ROOT_FILES: [&str; 3] = ["CLAUDE.md", "AGENTS.md", "AGENTS.local.md"];
/// The folder of an apply's backup area that takes the earlier entries the user picked for
/// `--set-aside`.
const SET_ASIDE: &str = "set-aside";
/// The name a revert gives this generation's `.moai/` before the Go folder comes back.
const RUST_ASIDE: &str = ".moai.rust-";
/// The folder of an apply's backup area where a revert keeps a `.claude/settings.json` that was
/// written after the move, under `<UTC>/<same path>`, before the earlier one comes back.
const REPLACED_BY_REVERT: &str = "replaced-by-revert";
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
    /// #2936 S1c: the ownership parts of a `generated_managed` file (`manifest.Part`). Read
    /// leniently (any shape): only the count of parts that are not `created` is shown.
    #[serde(default)]
    parts: serde_json::Value,
}

impl Entry {
    /// (all parts, parts whose origin is not `created`).
    fn part_counts(&self) -> (usize, usize) {
        let parts = self.parts.as_array().map(Vec::as_slice).unwrap_or(&[]);
        let other = parts
            .iter()
            .filter(|p| p["origin"].as_str() != Some("created"))
            .count();
        (parts.len(), other)
    }
}

/// Whether `actual` (`sha256:<hex>` of the bytes on disk) is the hash the manifest recorded.
/// The Go deployer records `sha256:<hex>`; the Codex wiring records its `generated_managed`
/// files with bare hex (`internal/codexwiring/ownership.go`, `intendedEntry`). Only that
/// provenance is normalized; every other record must match exactly.
fn same_hash(provenance: &str, recorded: &str, actual: &str) -> bool {
    if recorded.is_empty() {
        return false;
    }
    if provenance == "generated_managed" {
        let bare = |h: &str| h.strip_prefix("sha256:").unwrap_or(h).to_string();
        return bare(recorded) == bare(actual);
    }
    recorded == actual
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
    /// #2936 S-CYCLE: on a moving item that the Go `.claude/settings.json` named, the path of
    /// those settings: the item stays in place only while settings that stay name it, and these
    /// settings move, so it moves with them. Empty for every other item. Preview-only.
    #[serde(skip_serializing_if = "String::is_empty")]
    pub freed_by: String,
    /// The manifest's `deployed_hash`, kept for the receipt of an edited file (O2).
    #[serde(skip)]
    pub(crate) deployed_hash: String,
    /// S-DIGEST (#2936): what the moving item holds, as the preview read it — a file's
    /// `sha256:<hex>`, a mirror link's body digest, a copy folder's digest over its files;
    /// empty for the folders `.moai` and `.moai-cli` (one rename, content not pinned) and for
    /// every item that stays. The digest binds it and the receipt takes it, so the preview
    /// and the apply read each moving file once.
    #[serde(skip)]
    pub(crate) content: String,
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
    /// Regular files under `.moai/`, `.claude/`, `.agents/` and `.codex/` the manifest does not record, counted per
    /// second-level folder (`.moai/specs`, `.claude/worktrees`, …). They are never moved.
    pub unrecorded: BTreeMap<String, u64>,
    /// AC-5: files copied from the Go backup's `specs/` into `.moai/specs` (`copy`), or not
    /// (`conflict`, `conflict_link`, `excluded_secret`, `excluded_unclassifiable`,
    /// `leave_link`). Empty unless the folder moves.
    pub restores: Vec<Restore>,
    /// #2696: the name the Go `.moai/` gets — `.moai-backup`, or `.moai-backup-<UTC>` when
    /// that is taken. `None` unless the folder moves.
    pub go_backup: Option<String>,
    /// #2936 S0: what the move does with the Go `.moai/` folder, in one word — `folder_moves`
    /// (the folder becomes the Go backup, see `go_backup`), `folder_stays` (the folder stays
    /// where it is: settings name it, a worktree or nested checkout is inside, …),
    /// `legacy_into_new` (no Go folder; the earlier root's product state moves into a new
    /// `.moai/`) or `nothing`.
    pub layout: &'static str,
    /// #2936 S0: where the dated backup of the moving files would go, `<UTC>` filled in by the
    /// apply: `.moai/backups/<UTC>` when a new `.moai/` is made, else the current state
    /// folder's `backups/<UTC>`. `None` when no file moves or is copied — then no backup is
    /// made. A `folder_stays` plan can still move proven `.claude/` files, and then names it.
    pub backup_to: Option<String>,
    /// #2696: the earlier root's entries, each `move` (into `.moai/`) or `leave`.
    pub state_moves: Vec<StateMove>,
    /// #2696, #2936 O4: `CLAUDE.md`, `AGENTS.md` and `AGENTS.local.md`, moved into the backup
    /// (`move`: gone from the project root; the install step writes a new `CLAUDE.md`, and the
    /// older files are not written again) or left
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
    /// #2936 S-CYCLE: set when the Go `.claude/settings.json` moves: what that frees. Omitted
    /// otherwise, so a project the rule does not touch previews as before. Preview-only.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub settings_move: Option<SettingsMove>,
    /// SPEC-20261126-go-local-settings-reference: what the local settings name under `.moai/`
    /// while they hold the Go folder, and whether [`REWRITE_PICK`] can rewrite them. Omitted
    /// otherwise, so a project without the hold previews as before. Preview-only.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub local_references: Option<LocalReferences>,
    /// #2936 S-AGENTS-STAY U1: set when the Go folder stays, files move, and the backup area's
    /// existing `.gitignore` does not keep the copies out of git (or git cannot say). The apply
    /// stops before it writes anything. Omitted otherwise, so a project without the problem
    /// previews as before. Preview-only.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub backup_ignore: Option<BackupIgnore>,
    /// SPEC-20261126-tracked-state-move: set when the earlier root's state would move into a
    /// `.moai/` that is absent or holds only the SPECs and the local folder, and git ignores
    /// where a tracked file or the marker would go, or cannot say. The apply stops before it
    /// writes anything. Omitted otherwise. Preview-only: the apply judges again.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub ignored_tracked: Option<IgnoredTracked>,
    /// #2936 O6, Q6: secret-named entries under `.claude/`, `.agents/` and `.codex/`; only counted,
    /// never listed, read or moved.
    pub secrets: u64,
    pub warnings: Vec<String>,
    pub plan_digest: String,
}

/// The hold of the local settings on the Go folder: places and path texts only, never any
/// other part of a value.
#[derive(Clone, Debug, Serialize, PartialEq, Eq)]
pub struct LocalReferences {
    pub file: String,
    /// The pick of `go-state-plan` and `go-state-apply` that rewrites the file.
    pub pick: &'static str,
    /// This preview was made with the pick.
    pub picked: bool,
    pub references: Vec<LocalReference>,
    /// Why the pick cannot move the folder for these settings; empty when it can.
    pub blocked: Vec<String>,
}

#[derive(Clone, Debug, Serialize, PartialEq, Eq)]
pub struct LocalReference {
    /// The JSON pointer of the value (of the object, for a name inside a key).
    pub pointer: String,
    /// The `.moai/…` path text alone.
    pub path: String,
    /// `project` (the pick rewrites it) or `other`.
    pub class: &'static str,
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
    /// ignore list was not read) or the moving settings file could not be read as a JSON
    /// object (`wiring_lost` is then `None`): the lists above may be incomplete.
    pub unknown: bool,
}

/// #2936 S-CYCLE: the Go settings the manifest proves move even when the Go folder stays, and
/// what they hold is no longer held. Counts and paths only: never a rule, a command or a value.
#[derive(Clone, Debug, Serialize, PartialEq, Eq)]
pub struct SettingsMove {
    pub path: String,
    /// The manifest proof, as on the item.
    pub proof: String,
    /// The user changed the file since the Go deploy; the receipt keeps both hashes.
    pub edited: bool,
    /// The plan's `layout`: `folder_moves` or `folder_stays`.
    pub layout: String,
    /// The moving files that only these settings named (each carries `freed_by`).
    pub freed: Vec<String>,
    /// The settings named a path under `.moai/` and the Go folder moves now.
    pub folder_freed: bool,
    /// The settings files that still name a path under `.moai/` and keep the folder.
    pub still_holding: Vec<String>,
}

/// #2936 S-AGENTS-STAY U1: the backup area's `.gitignore` exists and git does not ignore the
/// copies the apply would make there. Paths only: never a line of the file.
#[derive(Clone, Debug, Serialize, PartialEq, Eq)]
pub struct BackupIgnore {
    /// `not_ignored` (git would offer a backup path) or `unknown` (git could not say; the check
    /// stops closed).
    pub state: &'static str,
    /// The ignore file that exists and is not enough, e.g. `.moai-cli/backups/.gitignore`.
    pub file: String,
    /// Backup paths git does not ignore, `<UTC>` standing for the apply's stamp; at most
    /// [`IGNORE_EXAMPLES`]. Empty for `unknown`.
    pub examples: Vec<String>,
    pub reason: String,
}

/// SPEC-20261126-tracked-state-move: tracked files of the earlier root would leave version
/// control with the move. Paths and `.gitignore` rules only: never a file's content.
#[derive(Clone, Debug, Serialize, PartialEq, Eq)]
pub struct IgnoredTracked {
    /// `ignored` (git ignores a destination) or `unknown` (git could not say; the check stops
    /// closed).
    pub state: &'static str,
    /// The tracked files whose destination git ignores, at most [`IGNORED_FILES`]. Empty for
    /// `unknown`, and when only the marker is hidden.
    pub files: Vec<HiddenFile>,
    /// How many tracked files have an ignored destination, the named ones included.
    pub total: usize,
    /// The rule that hides the marker, when git ignores it.
    pub marker: Option<String>,
    pub reason: String,
}

/// One tracked file the move would put where git ignores it.
#[derive(Clone, Debug, Serialize, PartialEq, Eq)]
pub struct HiddenFile {
    /// The tracked file, relative to the project.
    pub file: String,
    /// Where the move would put it.
    pub to: String,
    /// `<source>:<line>:<pattern>` of the rule that hides `to`.
    pub rule: String,
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
    /// `move`, `leave`, or (a `.moai/` that already holds the team's kept names) `set-aside`:
    /// the earlier entry moves into the apply's backup area because the user picked it.
    pub action: String,
    pub reason: String,
    /// How many files git tracks under `from` (preview-only; shown for a `.moai/` that holds
    /// kept names, never when it is zero or git cannot say).
    #[serde(skip_serializing_if = "Option::is_none")]
    pub tracked: Option<u64>,
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

/// The restore candidates of the Go `.moai/` folder: every regular file of the user-data folders
/// ([`USER_DATA_FROM`], #2936 O1; `.moai/specs` was the first, AC-5), and, outside them, every
/// file the Go manifest does not record (a recorded file is Go's own and stays in the backup;
/// the markers and the Go backups folder are never restored, [`restore_excluded`],
/// SPEC-20261126-go-state-restore-rule). Regular files are copied unless the target exists
/// (never overwritten) or the name or content is or may be a secret (a secret name is
/// excluded before any read). A document (`.md`, `.txt`, anything but a config file) passes on
/// its name and size; a config-shaped file (`.json`, `.toml`, `.yaml`, `.yml`) and anything
/// under `specs/` goes through the `migrate` filter, fail-closed, and a refusal carries its
/// reason. Links, a linked user-data folder and special files are never followed or opened. The copy reads the file from the Go
/// backup after the rename. A folder the earlier root also moves into `.moai/` (only `specs/`)
/// lands there first, so a file already at `.moai-cli/<folder>/<same path>` is the conflict.
fn plan_restores(
    project: &Path,
    go_backup: &str,
    recorded: &HashMap<String, Entry>,
) -> Vec<Restore> {
    let mut out = Vec::new();
    let moai = project.join(MOAI_DIR);
    {
        {
            let mut stack = vec![moai.clone()];
            while let Some(dir) = stack.pop() {
                let Ok(entries) = fs::read_dir(&dir) else {
                    continue;
                };
                for e in entries.flatten() {
                    let p = e.path();
                    let Some(rel) = p
                        .strip_prefix(&moai)
                        .ok()
                        .and_then(|r| r.to_str())
                        .map(|r| r.replace('\\', "/"))
                    else {
                        continue;
                    };
                    let folder = rel.split('/').next().unwrap_or_default().to_string();
                    let folder = folder.as_str();
                    let user_data = USER_DATA_FROM.contains(&folder);
                    // Outside the ten folders only what the Go manifest does not record is
                    // the user's: a recorded file is Go's own and stays in the backup, and the
                    // markers and the Go backups are never restored
                    // (SPEC-20261126-go-state-restore-rule REQ-2, REQ-3).
                    if !user_data
                        && (restore_excluded(&rel)
                            || recorded.contains_key(&format!("{MOAI_DIR}/{rel}")))
                    {
                        continue;
                    }
                    // A user-data folder that is itself a link is never entered: the apply
                    // refuses to read through it (after the Go folder was renamed), so the
                    // preview leaves it where it is.
                    if user_data
                        && rel == folder
                        && fs::symlink_metadata(&p).is_ok_and(|m| m.file_type().is_symlink())
                    {
                        out.push(Restore {
                            from: format!("{go_backup}/{folder}"),
                            to: format!("{MOAI_DIR}/{folder}"),
                            sha256: String::new(),
                            action: "leave_link".into(),
                            reason: "the folder is a link; never followed, it stays in the backup"
                                .into(),
                        });
                        continue;
                    }
                    let from = format!("{go_backup}/{rel}");
                    let to = format!("{MOAI_DIR}/{rel}");
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
                        .then(|| format!("{LEGACY_ROOT}/{rel}"));
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
                                        "the content looks like a credential; not copied"
                                            .to_string(),
                                    ),
                                    Err(why) => ("excluded_unclassifiable", why),
                                    Ok(false) => (
                                        match &lands {
                                            Some(l) if linked_component(project, l) => {
                                                "conflict_link"
                                            }
                                            Some(l)
                                                if fs::symlink_metadata(project.join(l))
                                                    .is_ok() =>
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
    }
    out.sort_by(|a, b| a.from.cmp(&b.from));
    out
}

/// A path under the Go `.moai/` (relative to it) that is never restored, whoever wrote it: the
/// three Go markers, whose restore would make the new `.moai/` read as the Go generation again
/// (`state_root`), the Go tool's update lock (its process lock: the update reports a dead one as a
/// leftover fact and expects it to stay in the Go backup), and the Go tool's own backups folder.
fn restore_excluded(rel: &str) -> bool {
    const NEVER_RESTORED: [&str; 4] = [
        "manifest.json",
        "config/config.yaml",
        "config/sections/system.yaml",
        ".update.lock",
    ];
    NEVER_RESTORED.contains(&rel) || rel.split('/').next() == Some("backups")
}

pub(crate) fn sha256_file(path: &Path) -> Result<String> {
    let meta = fs::symlink_metadata(path).map_err(|_| "file metadata unavailable")?;
    if !meta.is_file() || meta.len() > FILE_LIMIT {
        return Err("not a bounded regular file".into());
    }
    let bytes = fs::read(path).map_err(|_| "file read failed")?;
    Ok(format!("sha256:{:x}", Sha256::digest(bytes)))
}

/// A plain relative key: normal components only, no backslash, at least one component.
pub(crate) fn plain_relative(key: &str) -> bool {
    !key.is_empty()
        && !key.contains('\\')
        && Path::new(key)
            .components()
            .all(|c| matches!(c, Component::Normal(_)))
}

/// Whether any component from the project root down to `rel` is a symlink.
pub(crate) fn linked_component(project: &Path, rel: &str) -> bool {
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

/// #2936 S-CYCLE: what proves `.claude/settings.json` the Go tool's file.
struct SettingsProof {
    /// The user changed the file since the Go deploy (or the manifest says `user_modified`).
    edited: bool,
    /// The manifest proof, as the item shows it.
    proof: String,
    /// The file's `sha256:<hex>`, read once here.
    content: String,
}

/// Why the settings are not proven: the (reason, class) the item shows.
type NotProven = (&'static str, &'static str);

/// #2936 S-CYCLE: the one place that decides whether the shared settings are the Go tool's.
/// Both [`policy`] (the settings item) and [`plan`] (which settings bodies hold files and the
/// folder) use this verdict, so they cannot disagree. Proven means: the manifest records the
/// file as `template_managed` or `user_modified`, no component of the path is a link, it is a
/// regular file, and its hash can be read. Anything else (`user_created`, another provenance,
/// a link, a folder, an unreadable file) is not proven and stays.
fn settings_proof(project: &Path, entry: &Entry) -> std::result::Result<SettingsProof, NotProven> {
    if linked_component(project, SETTINGS) {
        return Err(("a path component is a symlink", "unknown"));
    }
    let path = project.join(SETTINGS);
    match fs::symlink_metadata(&path) {
        Err(_) => return Err(("missing", "unknown")),
        Ok(m) if !m.is_file() => return Err(("not a regular file", "unknown")),
        Ok(_) => {}
    }
    match entry.provenance.as_str() {
        "user_created" => return Err(("user_created: the user's own file", "user_data")),
        "template_managed" | "user_modified" => {}
        _ => return Err(("not template_managed", "unknown")),
    }
    // The Go manifest keeps recording this path after the move (the folder stays), so a file
    // the install phase wrote since is this generation's, not the Go tool's: when the Rust
    // settings receipt owns a key in it, it is not proven again (no second move) — unless a
    // revert put these very bytes back (S-AGENTS-STAY U2, below).
    let rust_owns =
        match rust_record(project, "manifests/claude-settings.json") {
            Ok(None) => false,
            Ok(Some(v)) if v["schema_version"].as_u64() == Some(1) && v["owned"].is_object() => {
                let owns = |o: &serde_json::Value| o.as_object().is_some_and(|m| !m.is_empty());
                owns(&v["owned"]) || owns(&v["pending"]["owned"])
            }
            _ => return Err((
                "a Rust settings receipt cannot be read, so these settings cannot be told apart \
                 from the Go tool's",
                "unknown",
            )),
        };
    let content = sha256_file(&path);
    if rust_owns
        && !content
            .as_ref()
            .is_ok_and(|h| restored_by_revert(project, h))
    {
        return Err((
            "this generation's settings: the Rust settings receipt owns keys in it",
            "rust",
        ));
    }
    let content = content.map_err(|_| ("not a bounded regular file", "unknown"))?;
    let unchanged = entry.provenance == "template_managed"
        && same_hash(&entry.provenance, &entry.deployed_hash, &content);
    // After a committed move a settings file someone wrote later is read here again (the
    // manifest still records the path): the text says what is known, not that the bytes are
    // the Go tool's.
    let proof = if unchanged {
        "manifest; sha256 matches the deploy"
    } else if entry.provenance == "user_modified" {
        "manifest; recorded user_modified"
    } else {
        "manifest records this path; content differs from what the Go tool deployed"
    }
    .to_string();
    Ok(SettingsProof {
        edited: !unchanged,
        proof,
        content,
    })
}

/// #2936 S-AGENTS-STAY U2: whether the settings file on disk (`content`, its `sha256:<hex>`) is
/// byte for byte the one an earlier apply moved and a revert then put back. The receipt of a
/// reverted apply records the hash of the settings it moved (`sha256`) and that they are back
/// (`moved: false`), so a file with that hash is the Go file again although the Rust settings
/// receipt, which the revert leaves alone, still owns keys. One changed character makes the
/// hashes differ, so a file the user edited since stays this generation's. Reads receipts only
/// (the newest [`RECEIPT_SCAN`] under either backup folder); a receipt that cannot be read, or
/// anything else unexpected, answers `false`.
fn restored_by_revert(project: &Path, content: &str) -> bool {
    const RECEIPT_SCAN: usize = 64;
    if content.is_empty() {
        return false;
    }
    let mut stamps: Vec<String> = Vec::new();
    for base in [MOAI_DIR, LEGACY_ROOT] {
        let backups = format!("{base}/backups");
        if linked_component(project, &backups) {
            continue;
        }
        if let Ok(entries) = fs::read_dir(project.join(&backups)) {
            stamps.extend(
                entries
                    .filter_map(|e| e.ok())
                    .filter_map(|e| e.file_name().into_string().ok())
                    .filter(|n| valid_stamp(n)),
            );
        }
    }
    stamps.sort_unstable_by(|a, b| b.cmp(a));
    stamps.dedup();
    stamps.into_iter().take(RECEIPT_SCAN).any(|stamp| {
        read_receipt(project, &stamp).is_ok_and(|(r, _, _)| {
            r.state == "reverted"
                && r.moves
                    .iter()
                    .any(|m| m.path == SETTINGS && !m.moved && m.sha256 == content)
        })
    })
}

/// The decision for one recorded path, as (action, reason, class, proof). The single place
/// the move rule lives. `content` is set to the file's `sha256:<hex>` once the policy has read
/// it (S-DIGEST: the digest binds it and the apply takes it, so the file is read once).
/// `settings` are the settings bodies that hold files: without the proven shared settings,
/// which move themselves (S-CYCLE). `verdict` is [`settings_proof`] for the recorded
/// `.claude/settings.json`, `None` when the manifest does not record it.
fn policy(
    project: &Path,
    key: &str,
    entry: &Entry,
    settings: &[(&'static str, String)],
    verdict: Option<&std::result::Result<SettingsProof, NotProven>>,
    folder_moves: bool,
    content: &mut String,
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
    // #2936 S-CYCLE (leader decision 6886519352): settings the manifest proves are the Go
    // tool's move whether or not the Go folder stays; what the install steps put in their place
    // (status line, hook set, the deny rules offered back) is in the same update. Settings that
    // are not proven stay (AUDIT #2584 B1) and keep holding what they name.
    if key == SETTINGS {
        return match verdict {
            Some(Ok(p)) => {
                *content = p.content.clone();
                let reason = match (p.edited, folder_moves) {
                    (false, true) => {
                        "the Go settings move to the backup; the install step writes new ones (O3)"
                    }
                    (false, false) => {
                        "the Go settings move to the backup although the Go folder stays; the \
                         install steps write new ones (S-CYCLE)"
                    }
                    // Not unchanged: the manifest records the path, but the file is not the
                    // unchanged Go deploy; it still moves, byte for byte.
                    (true, true) => {
                        "the manifest records this settings file, which is not the unchanged Go \
                         deploy; it moves to the backup as it is; the install step writes new \
                         ones (O3)"
                    }
                    (true, false) => {
                        "the manifest records this settings file, which is not the unchanged Go \
                         deploy; it moves to the backup as it is although the Go folder stays; \
                         the install steps write new ones (S-CYCLE)"
                    }
                };
                ("move", reason, "settings", p.proof.clone())
            }
            Some(Err(_)) | None if !folder_moves => leave(
                "live host settings; moving them drops permissions, hooks and the status line",
                "settings",
            ),
            Some(Err((reason, class))) => leave(reason, class),
            None => leave("missing", "unknown"),
        };
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
            let same = entry.provenance == "template_managed"
                && same_hash(&entry.provenance, &entry.deployed_hash, &h);
            *content = h;
            same
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
                i.content = digest(body.as_bytes());
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
        // `rel\0sha256` per file, in the sorted order of `files`: what the digest binds.
        let mut listing = String::new();
        for inner in &files {
            let rec = recorded.get(&format!("{canon}{inner}"));
            let h = sha256_file(&project.join(&rel).join(inner));
            if let Ok(h) = &h {
                listing.push_str(&format!("{inner}\0{h}\n"));
            }
            match (rec, h) {
                (Some(e), Ok(h)) if same_hash(&e.provenance, &e.deployed_hash, &h) => same += 1,
                _ => differ += 1,
            }
        }
        out.push(if files.is_empty() {
            new_item(&rel, "leave", "an empty folder".into(), "unknown")
        } else if differ == 0 {
            let mut i = new_item(&rel, "leave", String::new(), "go_mirror_copy");
            i.proof = format!("every file matches the deployed_hash under {canon}");
            i.content = digest(listing.as_bytes());
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

// ---- #2936 S1c-1: the Go generation's `.codex/` files (classify and preview only) ----

/// What this generation's own records say about `.codex/`, read once per preview.
struct CodexRecords {
    /// A record of this generation exists but cannot be read: Rust-made files cannot be told
    /// apart from Go's, so nothing under `.codex` is proven.
    broken: Option<&'static str>,
    /// `.codex/**` paths of the Rust generation manifest's `owned_files`.
    owned_paths: Vec<String>,
    /// The `sha256` and `pending_sha256` of the Rust hooks receipt.
    hooks_sha: Vec<String>,
    /// Whether the Rust config receipt owns any key.
    config_owned: bool,
    /// Why the Go wiring looks interrupted (a temp file or an unfinished journal entry).
    interrupted: Option<&'static str>,
}

/// One JSON record under this generation's state folder: `Ok(None)` when absent, `Err` when it
/// is a link, not a bounded regular file, or not JSON.
fn rust_record(project: &Path, sub: &str) -> std::result::Result<Option<serde_json::Value>, ()> {
    let rel = crate::state_root::rel(project, sub);
    match read_bounded_text(project, &rel, CODEX_READ_LIMIT) {
        Ok(None) => Ok(None),
        Ok(Some(text)) => serde_json::from_str(&text).map(Some).map_err(|_| ()),
        Err(_) => Err(()),
    }
}

fn codex_records(project: &Path) -> CodexRecords {
    let mut r = CodexRecords {
        broken: None,
        owned_paths: Vec::new(),
        hooks_sha: Vec::new(),
        config_owned: false,
        interrupted: wiring_interruption(project),
    };
    const BROKEN: &str = "a Rust record under the state folder cannot be read, so Rust-made files \
                          cannot be told apart from the Go tool's; it stays";
    match rust_record(project, "manifests/generation.json") {
        Ok(None) => {}
        Ok(Some(v)) => match v.get("owned_files") {
            None => {}
            Some(list) => match list.as_array() {
                Some(list) => {
                    r.owned_paths = list
                        .iter()
                        .filter_map(|f| f["path"].as_str())
                        .filter(|p| top_of(p) == CODEX_DIR)
                        .map(str::to_string)
                        .collect();
                }
                None => r.broken = Some(BROKEN),
            },
        },
        Err(()) => r.broken = Some(BROKEN),
    }
    match rust_record(project, "manifests/codex-hooks.json") {
        Ok(None) => {}
        Ok(Some(v)) if v["schema_version"].as_u64() == Some(1) => {
            r.hooks_sha = ["sha256", "pending_sha256"]
                .iter()
                .filter_map(|k| v[*k].as_str().map(str::to_string))
                .collect();
        }
        _ => r.broken = Some(BROKEN),
    }
    match rust_record(project, "manifests/codex-config.json") {
        Ok(None) => {}
        Ok(Some(v)) if v["schema_version"].as_u64() == Some(1) && v["owned"].is_object() => {
            let owns = |o: &serde_json::Value| o.as_object().is_some_and(|m| !m.is_empty());
            r.config_owned = owns(&v["owned"]) || owns(&v["pending"]["owned"]);
        }
        _ => r.broken = Some(BROKEN),
    }
    r
}

/// Q5: why the Go wiring looks interrupted, if it does. A temp file of the wiring's atomic
/// write sits in `.codex/`, or the Go journal holds an entry that is neither `complete` nor
/// `discarded` (or cannot be read). A finished journal stays on disk and is not a sign.
fn wiring_interruption(project: &Path) -> Option<&'static str> {
    if fs::symlink_metadata(project.join(CODEX_DIR)).is_ok_and(|m| m.is_dir()) {
        let temp = fs::read_dir(project.join(CODEX_DIR))
            .map(|rd| {
                rd.flatten()
                    .any(|e| e.file_name().to_string_lossy().starts_with(WIRING_TEMP))
            })
            .unwrap_or(false);
        if temp {
            return Some("a .codexwiring-* temp file is in .codex/");
        }
    }
    match read_bounded_text(project, GO_WIRING_JOURNAL, MANIFEST_LIMIT) {
        Ok(None) => None,
        Err(_) => Some("the Go wiring journal cannot be read"),
        Ok(Some(text)) => {
            let Ok(doc) = serde_json::from_str::<serde_json::Value>(&text) else {
                return Some("the Go wiring journal cannot be read");
            };
            match doc.get("entries") {
                None => None,
                Some(entries) => match entries.as_array() {
                    None => Some("the Go wiring journal cannot be read"),
                    Some(entries) => entries
                        .iter()
                        .any(|e| !matches!(e["state"].as_str(), Some("complete" | "discarded")))
                        .then_some("the Go wiring journal has an unfinished entry"),
                },
            }
        }
    }
}

/// Dotted names (at most `MAX_NAMES`, each cut to `MAX_NAME` characters) for the proof text:
/// key names are shown, never values.
fn key_names(names: &[String]) -> String {
    const MAX_NAMES: usize = 8;
    const MAX_NAME: usize = 64;
    let mut sorted: Vec<String> = names
        .iter()
        .map(|n| n.chars().take(MAX_NAME).collect())
        .collect();
    sorted.sort();
    sorted.dedup();
    let more = sorted.len().saturating_sub(MAX_NAMES);
    sorted.truncate(MAX_NAMES);
    if more > 0 {
        format!("{} and {more} more", sorted.join(", "))
    } else {
        sorted.join(", ")
    }
}

/// Content proof for `.codex/hooks.json` (PLAN S1c, key point b): the top-level keys are only
/// `description` and `hooks`, and every handler command is the Go wiring's own (it starts with
/// `moai hook ` or names a script under `/.codex/hooks/moai/`). `Err` says why, with key names
/// and counts only; a command is never shown.
fn hooks_proof(bytes: &[u8]) -> std::result::Result<String, String> {
    let Ok(serde_json::Value::Object(top)) = serde_json::from_slice::<serde_json::Value>(bytes)
    else {
        return Err("hooks.json does not parse as a JSON object".into());
    };
    let foreign: Vec<String> = top
        .keys()
        .filter(|k| !matches!(k.as_str(), "description" | "hooks"))
        .cloned()
        .collect();
    if !foreign.is_empty() {
        return Err(format!(
            "keys outside the Go wiring: {}",
            key_names(&foreign)
        ));
    }
    let (mut handlers, mut foreign_handlers) = (0usize, 0usize);
    if let Some(events) = top.get("hooks") {
        let shape = "hooks.json holds a hooks table the Go wiring does not write";
        let events = events.as_object().ok_or(shape)?;
        for entries in events.values() {
            for entry in entries.as_array().ok_or(shape)? {
                for handler in entry["hooks"].as_array().ok_or(shape)? {
                    handlers += 1;
                    let own = handler["command"].as_str().is_some_and(|c| {
                        c.starts_with(GO_HANDLER_PREFIX) || c.contains(GO_HANDLER_NAMESPACE)
                    });
                    if !own {
                        foreign_handlers += 1;
                    }
                }
            }
        }
    }
    if foreign_handlers > 0 {
        return Err(format!("{foreign_handlers} handler(s) are not MoAI's"));
    }
    let keys: Vec<String> = top.keys().cloned().collect();
    Ok(format!(
        "content: keys {}; {handlers} handler(s), all MoAI's",
        key_names(&keys)
    ))
}

/// The dotted names of every value under `table`, tables entered, arrays and values leaves.
fn toml_key_paths(table: &dyn toml_edit::TableLike, prefix: &str, out: &mut Vec<String>) {
    for (key, item) in table.iter() {
        let path = if prefix.is_empty() {
            key.to_string()
        } else {
            format!("{prefix}.{key}")
        };
        match item {
            toml_edit::Item::None => {}
            toml_edit::Item::Table(t) => toml_key_paths(t, &path, out),
            toml_edit::Item::Value(toml_edit::Value::InlineTable(t)) => {
                toml_key_paths(t, &path, out)
            }
            _ => out.push(path),
        }
    }
}

fn toml_strings(item: Option<&toml_edit::Item>) -> Option<Vec<&str>> {
    item?.as_array()?.iter().map(|v| v.as_str()).collect()
}

/// Content proof for `.codex/config.toml` (PLAN S1c, key point b): the only keys are the Go
/// `[mcp_servers.moai]` table with exactly the Go values, and `tui.status_line` (a list of
/// strings, any items). `Err` says why with key names only, never a value.
fn config_proof(bytes: &[u8]) -> std::result::Result<String, String> {
    let parse = "config.toml does not parse as TOML";
    let text = std::str::from_utf8(bytes).map_err(|_| parse)?;
    let doc: toml_edit::DocumentMut = text.parse().map_err(|_| parse)?;
    let mut paths = Vec::new();
    toml_key_paths(doc.as_table(), "", &mut paths);
    let mcp_key = |k: &str| format!("mcp_servers.moai.{k}");
    let mcp_keys = ["command", "args", "env_vars", "default_tools_approval_mode"];
    let foreign: Vec<String> = paths
        .iter()
        .filter(|p| p.as_str() != "tui.status_line" && !mcp_keys.iter().any(|k| **p == mcp_key(k)))
        .cloned()
        .collect();
    if !foreign.is_empty() {
        return Err(format!(
            "keys outside the Go wiring: {}",
            key_names(&foreign)
        ));
    }
    let moai = doc
        .get("mcp_servers")
        .and_then(|m| m.as_table_like())
        .and_then(|m| m.get("moai"))
        .and_then(|m| m.as_table_like());
    if let Some(t) = moai {
        let same = t.get("command").and_then(|v| v.as_str()) == Some("moai")
            && toml_strings(t.get("args")).as_deref() == Some(&GO_MCP_ARGS[..])
            && toml_strings(t.get("env_vars")).as_deref() == Some(&GO_MCP_ENV_VARS[..])
            && t.get("default_tools_approval_mode")
                .and_then(|v| v.as_str())
                == Some("writes");
        if !same {
            return Err("mcp_servers.moai is not the table the Go wiring writes".into());
        }
    }
    let line = doc
        .get("tui")
        .and_then(|t| t.as_table_like())
        .and_then(|t| t.get("status_line"));
    if line.is_some() && toml_strings(line).is_none() {
        return Err("tui.status_line is not a list of strings".into());
    }
    let mut keys = Vec::new();
    if moai.is_some() {
        keys.push("mcp_servers.moai".to_string());
    }
    if line.is_some() {
        keys.push("tui.status_line".to_string());
    }
    Ok(format!("content: keys {}", key_names(&keys)))
}

/// A bounded regular file's bytes for the content proof; `Err` is the reason it is not read.
fn read_wiring_file(path: &Path) -> std::result::Result<Vec<u8>, String> {
    let m = fs::symlink_metadata(path).map_err(|_| "file metadata unavailable".to_string())?;
    if !m.is_file() {
        return Err("not a regular file".into());
    }
    if m.len() > CODEX_READ_LIMIT {
        return Err(format!("larger than {CODEX_READ_LIMIT} bytes; not read"));
    }
    let bytes = fs::read(path).map_err(|_| "file read failed".to_string())?;
    if bytes.len() as u64 > CODEX_READ_LIMIT {
        return Err(format!("larger than {CODEX_READ_LIMIT} bytes; not read"));
    }
    Ok(bytes)
}

/// Why a `.codex` path is Rust-made, if it is: listed in the Rust generation manifest, the
/// bytes of `hooks.json` the Rust hooks receipt owns, or a `config.toml` whose keys the Rust
/// config receipt owns. Rust-made files are never classified as Go's.
fn rust_proof(project: &Path, path: &str, rec: &CodexRecords) -> Option<&'static str> {
    if rec.owned_paths.iter().any(|p| p == path) {
        return Some("listed in the Rust generation manifest");
    }
    if path == CODEX_HOOKS && !rec.hooks_sha.is_empty() {
        let bytes = read_wiring_file(&project.join(path)).ok()?;
        if rec.hooks_sha.contains(&crate::digest(&bytes)) {
            return Some("sha256 equals the Rust hooks receipt");
        }
    }
    if path == CODEX_CONFIG && rec.config_owned {
        return Some("the Rust config receipt owns keys in it");
    }
    None
}

/// The item for one `.codex` path; `None` for a path only the Rust side named that is not
/// Rust's after all (or is not there). Always `leave`: S1c-1 moves nothing.
fn codex_item(
    project: &Path,
    path: &str,
    entry: Option<&Entry>,
    rec: &CodexRecords,
) -> Option<Item> {
    let item = |class: &str, reason: &str, proof: String| {
        let mut i = new_item(path, "leave", reason.to_string(), class);
        i.proof = proof;
        i.deployed_hash = entry.map(|e| e.deployed_hash.clone()).unwrap_or_default();
        Some(i)
    };
    if !plain_relative(path) {
        return item("outside", "not a plain relative path", String::new());
    }
    if linked_component(project, path) {
        return item(
            "unknown",
            "a path component is a symlink; it stays",
            String::new(),
        );
    }
    match fs::symlink_metadata(project.join(path)) {
        Err(_) if entry.is_some() => return item("unknown", "missing", String::new()),
        Err(_) => return None,
        Ok(m) if !m.is_file() => {
            return item("unknown", "not a regular file", String::new());
        }
        Ok(_) => {}
    }
    if let Some(why) = rec.broken {
        return item("unknown", why, String::new());
    }
    if let Some(why) = rust_proof(project, path, rec) {
        return item("rust", "this generation's file; it stays", why.to_string());
    }
    let entry = entry?;
    if let Some(why) = rec.interrupted {
        return item(
            "interrupted",
            "interrupted Codex wiring; everything under .codex stays (Q5)",
            why.to_string(),
        );
    }
    if entry.provenance == "user_created" {
        return item(
            "user_data",
            "user_created: the user's own file",
            String::new(),
        );
    }
    let full = project.join(path);
    if path == CODEX_HOOKS || path == CODEX_CONFIG {
        return wiring_item(path, &full, entry, &item);
    }
    if path.starts_with(CODEX_GO_AGENTS) {
        return agent_item(&full, entry, &item);
    }
    if path.starts_with(CODEX_GO_HOOK_SCRIPTS) {
        return item(
            "go_hook_script",
            "the Go hook scripts stay in place (Q4)",
            String::new(),
        );
    }
    item(
        "unknown",
        "no rule moves this .codex path; it stays",
        String::new(),
    )
}

/// `.codex/agents/moai/*.toml`: a `template_managed` file with its deployed hash is the tool's;
/// an edited or `user_modified` one moves later with both hashes recorded (O2, Q3).
fn agent_item(
    full: &Path,
    entry: &Entry,
    item: &dyn Fn(&str, &str, String) -> Option<Item>,
) -> Option<Item> {
    if !matches!(
        entry.provenance.as_str(),
        "template_managed" | "user_modified"
    ) {
        return item("unknown", "not template_managed", String::new());
    }
    let Ok(h) = sha256_file(full) else {
        return item("unknown", "not a bounded regular file", String::new());
    };
    let unchanged = entry.provenance == "template_managed"
        && same_hash(&entry.provenance, &entry.deployed_hash, &h);
    if unchanged {
        item(
            "tool_asset",
            MOVES_LATER,
            "manifest; sha256 matches the deploy".to_string(),
        )
    } else if entry.provenance == "user_modified" {
        item(
            "user_modified",
            MOVES_LATER,
            "manifest; recorded user_modified".to_string(),
        )
    } else {
        item(
            "user_modified",
            MOVES_LATER,
            "manifest; sha256 differs from the deploy".to_string(),
        )
    }
}

/// `.codex/hooks.json` and `.codex/config.toml`: the manifest's `generated_managed` record, the
/// hash and then the content must all prove the file is only the Go wiring's (hash alone is not
/// proof: the wiring re-records the hash after the user edits the file).
fn wiring_item(
    path: &str,
    full: &Path,
    entry: &Entry,
    item: &dyn Fn(&str, &str, String) -> Option<Item>,
) -> Option<Item> {
    if entry.provenance != "generated_managed" {
        return item("unknown", "not generated_managed", String::new());
    }
    let bytes = match read_wiring_file(full) {
        Ok(b) => b,
        Err(why) => return item("unknown", &why, String::new()),
    };
    let (all, other) = entry.part_counts();
    let record = format!("manifest generated_managed; {all} part(s), {other} not created");
    if !same_hash(
        &entry.provenance,
        &entry.deployed_hash,
        &crate::digest(&bytes),
    ) {
        return item(
            "user_modified",
            "sha256 differs from the wiring record; it stays in place",
            format!("{record}; sha256 differs from the wiring record"),
        );
    }
    let hooks = path == CODEX_HOOKS;
    match if hooks {
        hooks_proof(&bytes)
    } else {
        config_proof(&bytes)
    } {
        Ok(content) => item(
            if hooks { "go_hooks" } else { "go_config" },
            MOVES_LATER,
            format!("{record}; sha256 matches; {content}"),
        ),
        Err(why) => item(
            "user_modified",
            "content is not proven to be only the Go wiring's; it stays in place",
            format!("{record}; sha256 matches; {why}"),
        ),
    }
}

/// #2936 S1c-1: the items for `.codex/`, always `leave`, and their warnings. Candidates are the
/// Go manifest's `.codex` records and the `.codex` files this generation made; a secret-named
/// path is skipped before any read (the `secrets` count has it).
fn classify_codex(project: &Path, recorded: &HashMap<String, Entry>) -> (Vec<Item>, Vec<String>) {
    let rec = codex_records(project);
    let mut warnings = Vec::new();
    let mut paths: BTreeMap<String, Option<&Entry>> = BTreeMap::new();
    for (key, entry) in recorded {
        if top_of(key) == CODEX_DIR && !has_secret_component(key) {
            paths.insert(key.clone(), Some(entry));
        }
    }
    let rust_named = rec
        .owned_paths
        .iter()
        .map(String::as_str)
        .chain((!rec.hooks_sha.is_empty()).then_some(CODEX_HOOKS))
        .chain(rec.config_owned.then_some(CODEX_CONFIG));
    for path in rust_named {
        if !has_secret_component(path) {
            paths.entry(path.to_string()).or_insert(None);
        }
    }
    let items: Vec<Item> = paths
        .iter()
        .filter_map(|(path, entry)| codex_item(project, path, *entry, &rec))
        .collect();
    if let Some(why) = rec.interrupted {
        if items.iter().any(|i| i.class == "interrupted") {
            warnings.push(format!(
                "interrupted Codex wiring ({why}): everything the Go wiring wrote under .codex \
                 stays in place until it is recovered (the Go tool: `moai tool enable codex`)"
            ));
        }
    }
    (items, warnings)
}

/// O6 (and Q6 for `.codex/`): secret-named entries under `.claude/`, `.agents/` and `.codex/` (a secret-named folder counts
/// once and is not entered), without reading any; `true` when the walk passed `limit`.
fn count_secrets(project: &Path, limit: u64) -> (u64, bool) {
    let (mut found, mut seen) = (0u64, 0u64);
    for root in [".claude", ".agents", CODEX_DIR] {
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

pub(crate) fn read_bounded_text(project: &Path, rel: &str, limit: u64) -> Result<Option<String>> {
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

/// Folders a stopped-walk reason names.
const NAMED_FOLDERS: usize = 5;

/// The first-level folders under `<project>/<rel>` with how many entries (files and folders at
/// any depth, a nested checkout counted but not entered) each holds, counted up to `limit`: a
/// folder over it shows `limit + 1`. Largest first, equal counts by name, at most
/// [`NAMED_FOLDERS`]. Only the stopped-walk reason asks for it, so no preview pays for it.
fn largest_entry_folders(project: &Path, rel: &str, limit: u64) -> Vec<(String, u64)> {
    let Ok(entries) = fs::read_dir(project.join(rel)) else {
        return Vec::new();
    };
    let mut folders: Vec<(String, u64)> = Vec::new();
    for e in entries.flatten() {
        let Ok(name) = e.file_name().into_string() else {
            continue;
        };
        if !fs::symlink_metadata(e.path()).is_ok_and(|m| m.is_dir()) {
            continue;
        }
        let mut count = 0u64;
        let mut stack = vec![e.path()];
        'walk: while let Some(dir) = stack.pop() {
            let Ok(inner) = fs::read_dir(&dir) else {
                continue;
            };
            for f in inner.flatten() {
                count += 1;
                if count > limit {
                    break 'walk;
                }
                let p = f.path();
                if fs::symlink_metadata(&p).is_ok_and(|m| m.is_dir())
                    && fs::symlink_metadata(p.join(".git")).is_err()
                {
                    stack.push(p);
                }
            }
        }
        folders.push((name, count));
    }
    folders.sort_by(|a, b| b.1.cmp(&a.1).then_with(|| a.0.cmp(&b.0)));
    folders.truncate(NAMED_FOLDERS);
    folders
}

/// Why `.moai/` stays when its nested-checkout check passed the entry limit: the largest
/// folders by name, and, when `reports` is one of them, the command that moves reports out.
fn stopped_walk_reason(project: &Path, limit: u64) -> String {
    let mut reason = format!(
        "the .moai/ folder holds more than {limit} entries, so it was not fully checked for \
         nested checkouts; the folder stays"
    );
    let folders = largest_entry_folders(project, MOAI_DIR, limit);
    if folders.is_empty() {
        return reason;
    }
    let named: Vec<String> = folders
        .iter()
        .map(|(name, n)| match n {
            n if *n > limit => format!("{name} (more than {limit} entries)"),
            1 => format!("{name} (1 entry)"),
            n => format!("{name} ({n} entries)"),
        })
        .collect();
    reason.push_str(&format!(
        ". Largest folders under .moai/: {}.",
        named.join(", ")
    ));
    if folders.iter().any(|(name, _)| name == "reports") {
        reason.push_str(
            " To make room, move the report files out of the project: `moai reports-move-plan .` \
             previews moving the ones git does not track to ~/.moai/reports/<project>/; apply \
             it, then run this preview again.",
        );
    }
    reason
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
    for root in COUNTED_ROOTS {
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
fn plan_state_moves(project: &Path, picks: &[String]) -> Result<(Vec<StateMove>, Option<String>)> {
    let root = project.join(LEGACY_ROOT);
    match fs::symlink_metadata(&root) {
        Err(_) => {
            unused_picks(picks, &[])?;
            return Ok((Vec::new(), None));
        }
        Ok(m) if !m.is_dir() => {
            unused_picks(picks, &[])?;
            return Ok((
                Vec::new(),
                Some(format!(
                    "{LEGACY_ROOT} is a link or not a folder; its state cannot move into the new \
                     .moai/, so the Go folder stays"
                )),
            ));
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
    // SDD B1/RB1: a `.moai/` that holds only the project's tracked SPECs is the destination,
    // so the earlier root's `specs/` cannot move as one folder. Each SPEC moves into
    // `.moai/specs/` on its own; an id on both sides is a conflict that keeps everything.
    let moai = project.join(MOAI_DIR);
    let into_kept = crate::state_root::kept_only(&moai);
    let into_specs = crate::state_root::specs_only(&moai)
        || (into_kept
            && fs::symlink_metadata(moai.join(crate::state_root::SPECS_DIR))
                .is_ok_and(|m| m.is_dir()));
    let mut used: Vec<&str> = Vec::new();
    for name in names {
        let from = format!("{LEGACY_ROOT}/{name}");
        let linked = fs::symlink_metadata(root.join(&name))
            .map(|m| m.file_type().is_symlink())
            .unwrap_or(true);
        if into_specs && name == crate::state_root::SPECS_DIR && !linked {
            plan_spec_merge(project, &mut out, &mut block);
            continue;
        }
        let mut to = String::new();
        let (action, reason) = if STATE_ENTRIES.contains(&name.as_str()) {
            if linked {
                block.get_or_insert_with(|| {
                    format!(
                        "{from} is a link; product state is never moved through a link, so the \
                         Go folder stays"
                    )
                });
                ("leave", "a link; never followed".to_string())
            } else if into_kept && fs::symlink_metadata(moai.join(&name)).is_ok() {
                // The same name stands on both sides: the earlier entry stays unless the user
                // picked it, and then it goes aside, never over the `.moai/` entry.
                if let Some(pick) = picks.iter().find(|p| **p == name) {
                    used.push(pick.as_str());
                    to = format!("{MOAI_DIR}/backups/<UTC>/set-aside/{name}");
                    (
                        "set-aside",
                        format!(
                            "picked: moves into the backup area; {MOAI_DIR}/{name} is never \
                             overwritten"
                        ),
                    )
                } else {
                    (
                        "leave",
                        format!(
                            "{MOAI_DIR}/{name} already exists; both stay (to move the earlier \
                             one aside pick it with `--set-aside {name}`)"
                        ),
                    )
                }
            } else {
                to = format!("{MOAI_DIR}/{name}");
                ("move", "this generation's product state".to_string())
            }
        } else {
            let why = match name.as_str() {
                "worktrees" => {
                    "lane worktrees stay until worktree-move-plan moves them (#2723, D6)"
                }
                "reports" => "reports stay until `moai reports-move-plan .` moves them (#2725 D5)",
                "backups" => "earlier go-state backups stay where their receipts expect them",
                _ => "not product state this build moves; it stays",
            };
            ("leave", why.to_string())
        };
        out.push(StateMove {
            to,
            from,
            action: action.into(),
            reason,
            tracked: None,
        });
    }
    unused_picks(picks, &used)?;
    Ok((out, block))
}

/// A `--set-aside` pick that named no colliding entry is a typo, not a request: the plan is
/// refused, so a different pick set can never silently plan the same thing.
fn unused_picks(picks: &[String], used: &[&str]) -> Result<()> {
    match picks.iter().find(|p| !used.contains(&p.as_str())) {
        None => Ok(()),
        Some(p) => Err(format!(
            "--set-aside {p}: no entry of {LEGACY_ROOT}/ with that name also exists in {MOAI_DIR}/ \
             beside the team's kept names; nothing planned"
        )),
    }
}

/// The tracked files of `.moai-cli/` per planned row and the `.gitignore` check (SPEC
/// REQ-5). One read-only `git ls-files` counts the tracked files under each row; then the target
/// of every `move` row whose source is tracked is asked of `git check-ignore`, and so is the
/// marker whenever the source of any `move` or `set-aside` row is tracked. A target ignored
/// while its source is tracked blocks the whole move (the returned reason names the entry and
/// the rule). A `set-aside` target is the backup area, which git ignores by design, so only its
/// count and a warning are given. Without a git work tree the counts are unknown and nothing
/// blocks; git failing to answer blocks, because the answer is then unknown too.
fn kept_git_checks(project: &Path, moves: &mut [StateMove]) -> (Option<String>, Option<String>) {
    if fs::symlink_metadata(project.join(".git")).is_err() {
        return (
            None,
            Some(
                "not a git repository: the tracked file counts and the .gitignore check of the \
                 targets are unknown and skipped"
                    .into(),
            ),
        );
    }
    let Ok(out) = git_in(project, &["ls-files", "-z", "--", LEGACY_ROOT], &[]) else {
        return (
            Some("git could not be run; nothing moves until it can".into()),
            None,
        );
    };
    if !out.status.success() {
        return (
            Some(format!(
                "git could not list the tracked files ({}); nothing moves until it can",
                crate::identity::failure_summary(&out.stderr)
            )),
            None,
        );
    }
    let listed = String::from_utf8_lossy(&out.stdout).into_owned();
    let tracked: Vec<&str> = listed.split('\0').filter(|p| !p.is_empty()).collect();
    for m in moves.iter_mut() {
        let under = format!("{}/", m.from);
        let n = tracked
            .iter()
            .filter(|p| **p == m.from || p.starts_with(&under))
            .count() as u64;
        m.tracked = (n > 0).then_some(n);
    }
    // A set-aside entry goes to the backup area, which `.moai/backups/.gitignore` hides by
    // design: no rule of the team's `.gitignore` is asked about it (a stamp that does not exist
    // yet cannot be allowed with a `!` line). The plan only says that git sees them leave.
    let aside: Vec<String> = moves
        .iter()
        .filter(|m| m.action == "set-aside")
        .filter_map(|m| {
            m.tracked
                .map(|n| format!("{} ({n} tracked file(s))", m.from))
        })
        .collect();
    let aside_note = (!aside.is_empty()).then(|| {
        format!(
            "set aside into the backup area, which git ignores by design: {}; git shows these \
             tracked files as deleted until they are committed or restored",
            aside.join(", ")
        )
    });
    // A folder entry is asked as a folder (a trailing `/`): the target does not exist yet, and
    // git matches a rule that names only folders (`!/.moai/manifests/`) only for a folder.
    let sources: Vec<(String, String)> = moves
        .iter()
        .filter(|m| m.tracked.is_some() && m.action == "move")
        .map(|m| {
            let to = m.to.clone();
            let folder = fs::symlink_metadata(project.join(&m.from)).is_ok_and(|s| s.is_dir());
            (m.from.clone(), if folder { format!("{to}/") } else { to })
        })
        .collect();
    // The marker is written whenever anything moves, so a tracked `set-aside` source asks for it
    // too, not only a tracked `move` source.
    let marker_matters = moves
        .iter()
        .any(|m| m.tracked.is_some() && matches!(m.action.as_str(), "move" | "set-aside"));
    if sources.is_empty() && !marker_matters {
        return (None, aside_note);
    }
    let marker = format!("{MOAI_DIR}/{}", crate::state_root::RUST_MARKER);
    let mut paths: Vec<String> = sources.iter().map(|(_, to)| to.clone()).collect();
    paths.push(marker.clone());
    let mut rules = Vec::new();
    for batch in paths.chunks(IGNORE_BATCH) {
        let out =
            match git_in(
                project,
                &["check-ignore", "--no-index", "-v", "-n", "--"],
                batch,
            ) {
                Ok(o) if matches!(o.status.code(), Some(0 | 1)) => o,
                _ => return (
                    Some(
                        "git could not say whether it ignores the targets; nothing moves until \
                         it can"
                            .into(),
                    ),
                    None,
                ),
            };
        let text = String::from_utf8_lossy(&out.stdout).into_owned();
        let lines: Vec<&str> = text.lines().collect();
        if lines.len() != batch.len() {
            return (
                Some("git answered for a different number of targets; nothing moves".into()),
                None,
            );
        }
        for line in lines {
            rules.push(ignoring_rule(line));
        }
    }
    for (i, rule) in rules.iter().enumerate() {
        let Some(rule) = rule else { continue };
        let (what, from) = match sources.get(i) {
            Some((from, to)) => (
                to.trim_end_matches('/').to_string(),
                format!("the {} tracked file(s) of {from}", {
                    moves
                        .iter()
                        .find(|m| &m.from == from)
                        .and_then(|m| m.tracked)
                        .unwrap_or(0)
                }),
            ),
            None => (
                marker.clone(),
                "the marker the team keeps in git".to_string(),
            ),
        };
        return (
            Some(format!(
                "{what} is ignored by the .gitignore rule `{rule}` while {from} would move \
                 there; nothing moves until the rule allows it (a `!/{what}` line)"
            )),
            None,
        );
    }
    (None, aside_note)
}

/// The rule text of one `git check-ignore -v -n` line when it ignores the path
/// (`<source>:<line>:<pattern>`), `None` for a miss or for an un-ignoring `!` pattern.
fn ignoring_rule(line: &str) -> Option<String> {
    (check_ignore_line(line) == Some(true)).then(|| {
        let head = line.split_once('\t').map(|(h, _)| h).unwrap_or(line);
        head.to_string()
    })
}

/// The revert-plan action for `.moai/specs`: set aside with `.moai/`, then moved straight back.
const KEEP_SPECS: &str = "restored_to_moai_specs";

/// The names a `.moai/` the apply did not create already holds and the revert leaves, sorted:
/// the receipt's `kept_before`. Everything else the folder may hold (`backups/`, Finder's
/// `.DS_Store`) is no such name.
fn kept_names_in(moai: &Path) -> Vec<String> {
    let mut names: Vec<String> = fs::read_dir(moai)
        .map(|rd| {
            rd.filter_map(|e| e.ok())
                .filter_map(|e| e.file_name().into_string().ok())
                .filter(|n| kept_name(n))
                .collect()
        })
        .unwrap_or_default();
    names.sort();
    names
}

/// Whether the `.moai/` a state move would reuse holds anything beyond the SPECs and the
/// machine's local folder, so that [`kept_git_checks`] judges the move; otherwise
/// [`ignored_tracked`] does. The local folder is not counted: a flaky record makes it, and a
/// record changes no answer of the plan. An unreadable folder counts as holding more.
fn holds_more_than_specs_and_local(moai: &Path) -> bool {
    fs::read_dir(moai)
        .map(|rd| {
            rd.filter_map(|e| e.ok()).any(|e| {
                let name = e.file_name();
                name != ".DS_Store"
                    && name != crate::state_root::SPECS_DIR
                    && name != crate::state_root::LOCAL_DIR
            })
        })
        .unwrap_or(true)
}

/// A first-level name the team keeps in git under `.moai/` ([`crate::state_root::KEPT`]) that is
/// not a generation marker, or the machine's local folder
/// ([`crate::state_root::LOCAL_DIR`]): both were there before an apply that reuses the folder.
fn kept_name(name: &str) -> bool {
    name == crate::state_root::LOCAL_DIR
        || (name != crate::state_root::RUST_MARKER
            && name != crate::state_root::LEGACY_MARKER
            && crate::state_root::KEPT
                .iter()
                .any(|k| k.trim_end_matches('/') == name))
}

/// `.moai/specs`, the project's tracked SPECs.
fn specs_rel() -> String {
    format!("{MOAI_DIR}/{}", crate::state_root::SPECS_DIR)
}

/// `specs/<one name>`: a state entry of the SPEC merge ([`plan_spec_merge`]).
fn is_spec_entry(e: &str) -> bool {
    e.strip_prefix(crate::state_root::SPECS_DIR)
        .and_then(|r| r.strip_prefix('/'))
        .is_some_and(|n| !n.is_empty() && !n.contains(['/', '\\']) && n != "." && n != "..")
}

/// The earlier root's `specs/` merged into a `.moai/specs/` that already holds the project's
/// tracked SPECs (RB1): one `move` per entry, `.moai-cli/specs/<name>` → `.moai/specs/<name>`.
/// A name on both sides is a `conflict`, and a linked entry stays; either keeps the whole
/// earlier root where it is (`block`), so nothing moves.
fn plan_spec_merge(project: &Path, out: &mut Vec<StateMove>, block: &mut Option<String>) {
    let specs = crate::state_root::SPECS_DIR;
    let from_dir = project.join(LEGACY_ROOT).join(specs);
    let to_dir = project.join(MOAI_DIR).join(specs);
    let mut names: Vec<String> = fs::read_dir(&from_dir)
        .map(|rd| {
            rd.filter_map(|e| e.ok())
                .filter_map(|e| e.file_name().into_string().ok())
                .collect()
        })
        .unwrap_or_default();
    names.sort();
    if names.is_empty() {
        out.push(StateMove {
            from: format!("{LEGACY_ROOT}/{specs}"),
            to: String::new(),
            action: "leave".into(),
            reason: format!("empty; {MOAI_DIR}/{specs} already holds the project's SPECs"),
            tracked: None,
        });
    }
    for name in names {
        let from = format!("{LEGACY_ROOT}/{specs}/{name}");
        let to = format!("{MOAI_DIR}/{specs}/{name}");
        let linked = fs::symlink_metadata(from_dir.join(&name))
            .map(|m| m.file_type().is_symlink())
            .unwrap_or(true);
        let (action, reason) = if linked {
            block.get_or_insert_with(|| {
                format!(
                    "{from} is a link; product state is never moved through a link, so it stays"
                )
            });
            ("leave", "a link; never followed".to_string())
        } else if fs::symlink_metadata(to_dir.join(&name)).is_ok() {
            block.get_or_insert_with(|| {
                format!(
                    "conflict: {name} is in both {LEGACY_ROOT}/{specs}/ and {MOAI_DIR}/{specs}/; \
                     nothing moves until one of them is renamed or removed"
                )
            });
            ("conflict", format!("{to} already exists; both copies stay"))
        } else {
            (
                "move",
                format!("a SPEC joins the project's tracked SPECs in {MOAI_DIR}/{specs}/"),
            )
        };
        out.push(StateMove {
            to: if action == "move" { to } else { String::new() },
            from,
            action: action.into(),
            reason,
            tracked: None,
        });
    }
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

/// Why a root instruction file moves into the backup (SPEC-20261122-claude-md-text-and-ignore
/// REQ-13). `CLAUDE.md` is written again by the install step; an older file is not: the CLAUDE
/// file of its kind takes its place and the moved copy stays in the backup.
fn moved_reason(name: &str, no_go_proof: bool) -> String {
    let lead = if no_go_proof {
        "no Go proof; moves to the backup as D-M1 says (O4)"
    } else {
        "moves to the backup"
    };
    let instead = match name {
        "AGENTS.md" => "CLAUDE.md",
        "AGENTS.local.md" => "CLAUDE.local.md",
        _ => return format!("{lead}; the install step writes a new one"),
    };
    format!(
        "{lead}; {instead} takes its place, {name} is not written again, and the moved copy \
         stays in the backup"
    )
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
        let moved = moved_reason(name, proof == "none");
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
        } else {
            ("move", moved.as_str(), "instructions")
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

// ---- SPEC-20261126-go-local-settings-reference: the local settings that name `.moai/` ----
//
// A Go project whose local settings name a path under `.moai/` keeps its Go folder, and a
// session never edits that file by hand. The pick backs the file up (the root-files copy),
// replaces the folder name in each project reference with the Go backup's and moves the
// folder. Only the bytes of those names change; each permission deny or ask entry also stays
// as it was.

/// The pick of `go-state-plan` and `go-state-apply` that rewrites the local settings.
pub const REWRITE_PICK: &str = "--rewrite-local-references";
/// What a hold of the folder by settings looks for, in the preview and the pick alike.
const FOLDER_REFERENCE: &str = ".moai/";
/// The Claude Code project folder variable, in both of its shell spellings.
const PROJECT_VARIABLES: [&str; 2] = ["$CLAUDE_PROJECT_DIR", "${CLAUDE_PROJECT_DIR}"];
/// The permission lists that restrict: the original entry stays and its rewritten twin joins
/// it.
const KEPT_LISTS: [&str; 2] = ["/permissions/deny", "/permissions/ask"];
/// A backup name as long as the longest one, for the size check of the rewritten file.
const LONGEST_BACKUP: &str = ".moai-backup-00000000T000000Z";
/// The most characters of one path text the preview prints.
const SHOWN_PATH: usize = 200;

/// One string of the settings file, as it is spelled and where it stands.
struct JsonString {
    /// Byte range of the token in the file, quotes included.
    start: usize,
    end: usize,
    value: String,
    /// The value's JSON pointer; for a key, the pointer of its object.
    pointer: String,
    key: bool,
}

/// Every string token of a JSON text the caller has already parsed (so the nesting is within
/// the parser's depth limit).
struct JsonStrings<'a> {
    text: &'a str,
    at: usize,
    out: Vec<JsonString>,
}

impl JsonStrings<'_> {
    fn peek(&mut self) -> Option<u8> {
        let bytes = self.text.as_bytes();
        while bytes.get(self.at).is_some_and(u8::is_ascii_whitespace) {
            self.at += 1;
        }
        bytes.get(self.at).copied()
    }

    fn string(&mut self, pointer: &str, key: bool) -> Option<String> {
        let bytes = self.text.as_bytes();
        let start = self.at;
        let mut i = start + 1;
        while *bytes.get(i)? != b'"' {
            i += if bytes[i] == b'\\' { 2 } else { 1 };
        }
        let end = i + 1;
        let value: String = serde_json::from_str(self.text.get(start..end)?).ok()?;
        self.at = end;
        self.out.push(JsonString {
            start,
            end,
            value: value.clone(),
            pointer: pointer.to_string(),
            key,
        });
        Some(value)
    }

    fn value(&mut self, pointer: &str) -> Option<()> {
        match self.peek()? {
            b'"' => {
                self.string(pointer, false)?;
            }
            open @ (b'{' | b'[') => {
                self.at += 1;
                let close = if open == b'{' { b'}' } else { b']' };
                let mut index = 0usize;
                while self.peek()? != close {
                    let child = if open == b'{' {
                        let name = self.string(pointer, true)?;
                        // The `:`.
                        self.peek()?;
                        self.at += 1;
                        format!("{pointer}/{}", name.replace('~', "~0").replace('/', "~1"))
                    } else {
                        format!("{pointer}/{index}")
                    };
                    self.value(&child)?;
                    index += 1;
                    if self.peek()? == b',' {
                        self.at += 1;
                    }
                }
                self.at += 1;
            }
            _ => {
                let bytes = self.text.as_bytes();
                while bytes
                    .get(self.at)
                    .is_some_and(|b| !b.is_ascii_whitespace() && !matches!(b, b',' | b']' | b'}'))
                {
                    self.at += 1;
                }
            }
        }
        Some(())
    }
}

fn json_strings(text: &str) -> Option<Vec<JsonString>> {
    let mut walk = JsonStrings {
        text,
        at: 0,
        out: Vec::new(),
    };
    walk.value("")?;
    Some(walk.out)
}

/// A shell or rule separator: what may stand right before a path.
fn separates(c: char) -> bool {
    c.is_whitespace() || "=(;|&<>:`,".contains(c)
}

/// Whether a path may start after `before`: the start of the value, a separator, or one
/// quote that itself opens there (a quote that closes another word does not).
fn path_starts_after(before: &str) -> bool {
    let mut back = before.chars().rev();
    match back.next() {
        None => true,
        Some('"' | '\'') => back.next().is_none_or(separates),
        Some(c) => separates(c),
    }
}

/// Whether the name at the end of `before` is the project's own folder: relative to the
/// project, after the project folder variable in any quoting, or after the absolute root.
/// Everything else (HOME, a parent or nested folder, another variable) is not.
fn project_reference(before: &str, root: &str, changes_folder: bool) -> bool {
    let relative = before.strip_suffix("./").unwrap_or(before);
    if !changes_folder && path_starts_after(relative) {
        return true;
    }
    let Some(head) = before.strip_suffix('/') else {
        return false;
    };
    let unquoted = head.strip_suffix(['"', '\'']).unwrap_or(head);
    if PROJECT_VARIABLES
        .iter()
        .filter_map(|v| unquoted.strip_suffix(v))
        .any(path_starts_after)
    {
        return true;
    }
    // `//<root>/…` is how a permission rule writes an absolute path.
    head.strip_suffix(root).is_some_and(|rest| {
        path_starts_after(rest) || rest.strip_suffix('/').is_some_and(path_starts_after)
    })
}

/// The `.moai/…` path text at the start of `from`, and nothing after it.
fn shown_path(from: &str) -> String {
    from.chars()
        .take_while(|c| c.is_ascii_alphanumeric() || "._-/*".contains(*c))
        .take(SHOWN_PATH)
        .collect()
}

/// The references of one settings text and why a rewrite of it is not possible.
fn scan_local(text: &str, root: &str) -> (Vec<LocalReference>, Vec<String>) {
    let object = serde_json::from_str::<serde_json::Value>(text).is_ok_and(|v| v.is_object());
    let Some(strings) = object.then(|| json_strings(text)).flatten() else {
        return (
            Vec::new(),
            vec![format!(
                "{SETTINGS_LOCAL} is not a readable JSON object, so the folder stays"
            )],
        );
    };
    let mut references = Vec::new();
    let mut blocked = Vec::new();
    let mut escaped = false;
    for s in &strings {
        let found: Vec<usize> = s
            .value
            .match_indices(FOLDER_REFERENCE)
            .map(|m| m.0)
            .collect();
        // The rewrite changes the bytes of the file, so every name must be spelled plainly.
        escaped |= text[s.start..s.end].matches(FOLDER_REFERENCE).count() != found.len();
        let changes_folder = s
            .value
            .split(|c: char| !(c.is_ascii_alphanumeric() || c == '_'))
            .any(|word| word == "cd" || word == "pushd");
        for at in found {
            let project = !s.key && project_reference(&s.value[..at], root, changes_folder);
            references.push(LocalReference {
                pointer: s.pointer.clone(),
                path: shown_path(&s.value[at..]),
                class: if project { "project" } else { "other" },
            });
        }
    }
    let others: Vec<&str> = references
        .iter()
        .filter(|r| r.class == "other")
        .map(|r| r.pointer.as_str())
        .collect();
    if !others.is_empty() {
        blocked.push(format!(
            "{SETTINGS_LOCAL} names the folder in a form the pick cannot rewrite; edit these \
             values yourself: {}",
            others.join(", ")
        ));
    }
    if escaped {
        blocked.push(format!(
            "{SETTINGS_LOCAL} spells the folder name with an escape; write it plainly yourself"
        ));
    }
    if blocked.is_empty() {
        let after = rewritten_text(text, &strings, LONGEST_BACKUP);
        if after.len() as u64 > crate::apply::RECEIPT_LIMIT {
            blocked.push(format!(
                "{SETTINGS_LOCAL} is or would become larger than {} KiB; the pick does not \
                 rewrite it",
                crate::apply::RECEIPT_LIMIT / 1024
            ));
        }
    }
    (references, blocked)
}

/// `text` with the folder name of every reference replaced by `go_backup`, and no other byte
/// changed; an entry of a [`KEPT_LISTS`] list stays, followed by its rewritten twin.
fn rewritten_text(text: &str, strings: &[JsonString], go_backup: &str) -> String {
    let moved = format!("{go_backup}/");
    let mut out = String::new();
    let mut done = 0;
    for s in strings {
        let token = &text[s.start..s.end];
        if s.key || !token.contains(FOLDER_REFERENCE) {
            continue;
        }
        let twin = token.replace(FOLDER_REFERENCE, &moved);
        out.push_str(&text[done..s.start]);
        let kept = s.pointer.rsplit_once('/').is_some_and(|(list, last)| {
            KEPT_LISTS.contains(&list) && last.parse::<usize>().is_ok()
        });
        if kept {
            out.push_str(token);
            out.push_str(", ");
        }
        out.push_str(&twin);
        done = s.end;
    }
    out.push_str(&text[done..]);
    out
}

/// The refusal of an apply with the pick whose fresh preview does not rewrite and move.
fn refuse_unrewritable(fresh: &GoStatePlan) -> Result<()> {
    let refused = |why: String| {
        Err(format!(
            "{REWRITE_PICK}: refused; nothing was written; {why}"
        ))
    };
    let Some(local) = &fresh.local_references else {
        return refused(format!(
            "{SETTINGS_LOCAL} holds no Go folder here; run go-state-plan without the pick"
        ));
    };
    if !local.blocked.is_empty() {
        return refused(local.blocked.join("; "));
    }
    match fresh.items.iter().find(|i| i.path == MOAI_DIR) {
        Some(folder) if folder.action == "move" => Ok(()),
        Some(folder) => refused(folder.reason.clone()),
        None => refused(format!("the preview moves no {MOAI_DIR}/ folder")),
    }
}

/// Replace the local settings with their rewritten text: atomically, with the mode they had.
/// `before` must still be what the file holds.
fn replace_local(project: &Path, before: &[u8], after: &[u8]) -> Result<()> {
    refuse_linked(project, SETTINGS_LOCAL)?;
    crate::apply::publish_replace_limited(
        &project.join(SETTINGS_LOCAL),
        Some(before),
        after,
        crate::apply::RECEIPT_LIMIT,
    )
}

/// How far an apply with the folder move got, for the message and the undo of a pick that
/// fails after its first write.
#[derive(Default)]
struct Progress {
    step: &'static str,
    /// The local settings before and after the rewrite, from just before the replace.
    rewrite: Option<(Vec<u8>, Vec<u8>)>,
}

impl Progress {
    /// REQ-8: the error of the failed apply, after putting the original settings back.
    fn undo_pick(self, project: &Path, error: &str, stamp: &str) -> String {
        let copy = format!("{MOAI_DIR}/backups/{stamp}/root-files/{SETTINGS_LOCAL}");
        let settings = match self.rewrite {
            None => format!("{SETTINGS_LOCAL} was not rewritten and holds its original bytes"),
            Some((before, after)) => {
                let now = read_bounded_file(&project.join(SETTINGS_LOCAL));
                if now.as_deref() == Some(before.as_slice()) {
                    format!("{SETTINGS_LOCAL} holds its original bytes")
                } else if now.as_deref() != Some(after.as_slice()) {
                    format!(
                        "{SETTINGS_LOCAL} changed meanwhile and was not touched; its original \
                         bytes are in {copy}"
                    )
                } else {
                    match replace_local(project, &after, &before) {
                        Ok(()) => format!("{SETTINGS_LOCAL} is back to its original bytes"),
                        Err(e) => format!(
                            "{SETTINGS_LOCAL} could not be put back ({e}); its original bytes \
                             are in {copy}"
                        ),
                    }
                }
            }
        };
        format!(
            "{error}; the apply with {REWRITE_PICK} stopped at the step \"{}\"; {settings}; see \
             `moai go-state-revert-plan <project> --backup {stamp}`",
            self.step
        )
    }
}

/// #2936 L2b: the Go proof of the root `CLAUDE.md`, for the leftovers `project-update-plan`
/// states as facts. `Some("manifest")` when the Go manifest records it, `Some("marker")` when it
/// holds a Go curator block; `None` for no file, a link or another non-file, a file over the
/// read limit, one with this generation's managed section, or one without either proof.
/// Read only; a manifest that cannot be read is no proof.
pub fn go_claude_md_proof(project: &Path) -> Option<&'static str> {
    const NAME: &str = "CLAUDE.md";
    if linked_component(project, NAME) {
        return None;
    }
    let bytes = read_bounded_file(&project.join(NAME))?;
    let text = String::from_utf8_lossy(&bytes);
    if text.contains(RUST_MARKER_TEXT) {
        return None;
    }
    let recorded = read_bounded_text(project, MANIFEST, MANIFEST_LIMIT)
        .ok()
        .flatten()
        .and_then(|t| serde_json::from_str::<Manifest>(&t).ok())
        .is_some_and(|m| m.files.contains_key(NAME));
    if recorded {
        Some("manifest")
    } else if GO_MARKERS.iter().any(|m| text.contains(m)) {
        Some("marker")
    } else {
        None
    }
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

const GONE: &str = "no counterpart \u{2014} will be gone";

/// #2936 R-A: what this generation puts in place of one kind of moved file, and a note. The
/// user layer is named, never read (O-3: this preview does not read HOME).
fn counterpart_of(kind: &str) -> (&'static str, &'static str) {
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
            "the claude-settings step writes the keys MoAI owns (a status line when there is \
             none) and the hooks step the MoAI hook set; permissions.deny and the other keys you \
             set come back through claude-settings-merge, key by key, with your go-ahead",
        ),
        _ => (
            GONE,
            "this generation installs nothing in its place; the files stay in the backup",
        ),
    }
}

/// #2936 R-A: the moving items grouped by kind, and the counterpart of each kind.
fn plan_counterparts(
    project: &Path,
    items: &[Item],
    settings: &[(&'static str, String)],
) -> (BTreeMap<String, u64>, Vec<Counterpart>) {
    let mut kinds = BTreeMap::new();
    for i in items
        .iter()
        .filter(|i| moving_file(i) && !i.kind.is_empty())
    {
        *kinds.entry(i.kind.clone()).or_insert(0u64) += 1;
    }
    let hooks_replace = kinds.contains_key("hook") && hooks_step_has_work(project, items, settings);
    let rows = kinds
        .iter()
        .map(|(kind, moved)| {
            let (counterpart, note) = match kind.as_str() {
                "hook" if !hooks_replace => (
                    GONE,
                    "after this move no Go hook entry or script is left, so the hooks step \
                     installs nothing; write the MoAI hook set with hooks-apply",
                ),
                _ => counterpart_of(kind),
            };
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

/// #2936 F-A: whether the `hooks` install step still finds a Go hook to replace once the moving
/// files are gone (`go_hooks::plan`: `replace` rather than `none`). A Go hook entry counts
/// when the settings file stays and carries one; a script counts when it stays in the hook
/// folder. A settings file that stays but is not a JSON object, or a hook folder that cannot
/// be listed, makes that step refuse rather than report `unchanged`, so both count as work.
fn hooks_step_has_work(
    project: &Path,
    items: &[Item],
    settings: &[(&'static str, String)],
) -> bool {
    let moving = |path: &str| items.iter().any(|i| moving_file(i) && i.path == path);
    if !moving(SETTINGS) {
        let text = settings
            .iter()
            .find(|(f, _)| *f == SETTINGS)
            .map_or("", |(_, text)| text.as_str());
        // `settings` holds "" both for a missing file and for an empty one, so the file's
        // presence is read here: an empty file that stays is not a JSON object either, and
        // the step refuses it just as it refuses any other non-object.
        if fs::symlink_metadata(project.join(SETTINGS)).is_ok() {
            match serde_json::from_str::<serde_json::Value>(text) {
                Ok(serde_json::Value::Object(map)) => {
                    let entries = map
                        .get("hooks")
                        .and_then(serde_json::Value::as_object)
                        .map_or(0, |h| crate::go_hooks::split(h).0.len());
                    if entries > 0 {
                        return true;
                    }
                }
                _ => return true,
            }
        }
    }
    match crate::go_hooks::scripts(project) {
        Ok(names) => names
            .iter()
            .any(|n| !moving(&format!("{}/{n}", crate::go_hooks::GO_HOOK_DIR))),
        Err(_) => true,
    }
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
    let settings_move = moving.contains(SETTINGS);
    let wiring_lost = settings_move
        .then(|| {
            settings
                .iter()
                .find(|(f, _)| *f == SETTINGS)
                .and_then(|(_, text)| wiring_of(text))
        })
        .flatten();
    // F-F: a moving settings file that is not a JSON object (invalid JSON, or another JSON
    // shape) has wiring the preview could not read; it says so instead of showing nothing.
    let wiring_unread = settings_move && wiring_lost.is_none();
    AfterMove {
        empty_folders: empty.into_iter().collect(),
        dangling,
        wiring_lost,
        unknown: walk_stopped || unread || wiring_unread,
    }
}

/// #2936 S0: the plan's `layout` and `backup_to`, from what the plan itself decided. A file
/// that moves or is copied needs a backup; the Go folder and the earlier root go with a new
/// `.moai/`, so their backup area lives there, any other move uses the current state folder.
fn layout_and_backup(
    project: &Path,
    items: &[Item],
    state_moves: &[StateMove],
    root_files: &[RootFile],
    restores: &[Restore],
) -> (&'static str, Option<String>) {
    let action_of = |path: &str| {
        items
            .iter()
            .find(|i| i.path == path)
            .map(|i| i.action.as_str())
    };
    let pending = items.iter().any(|i| i.action == "move")
        || state_moves
            .iter()
            .any(|m| m.action == "move" || m.action == "set-aside")
        || root_files.iter().any(|f| f.action != "leave")
        || restores.iter().any(|r| r.action == "copy");
    let layout = match (action_of(MOAI_DIR), action_of(LEGACY_ROOT)) {
        (Some("move"), _) => "folder_moves",
        (_, Some("move")) => "legacy_into_new",
        (Some(_), _) => "folder_stays",
        // No Go folder to judge: files that move anyway are still a stay, never `nothing`.
        _ if pending => "folder_stays",
        _ => "nothing",
    };
    let backup_to = pending.then(|| match layout {
        "folder_moves" | "legacy_into_new" => format!("{MOAI_DIR}/backups/<UTC>"),
        _ => format!("{}/backups/<UTC>", crate::state_root::dir(project)),
    });
    (layout, backup_to)
}

/// The read-only preview over one project folder.
pub fn plan(project: &Path) -> Result<GoStatePlan> {
    plan_with(project, &[])
}

/// [`plan`] with the entries the user picked for `--set-aside`. The picks are part of the
/// plan: a set-aside row is in the digest, so an apply made with other picks is stale.
pub fn plan_with(project: &Path, picks: &[String]) -> Result<GoStatePlan> {
    plan_picked(project, picks, false)
}

/// [`plan_with`] and the [`REWRITE_PICK`]. Where the pick can rewrite the local settings,
/// the plan moves the folder and its root row says `rewrite`, so the digest binds the pick.
pub fn plan_picked(project: &Path, picks: &[String], rewrite_local: bool) -> Result<GoStatePlan> {
    let project =
        fs::canonicalize(project).map_err(|_| crate::contract::PROJECT_PATH_UNAVAILABLE)?;
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
    // S1c: Codex reads these two files while a session runs (the Go wiring's trust layer).
    if [CODEX_HOOKS, CODEX_CONFIG]
        .iter()
        .any(|f| fs::symlink_metadata(project.join(f)).is_ok())
    {
        warnings.push(
            "A running Codex session reads .codex/hooks.json and .codex/config.toml; stop it \
             before a later step moves them (this preview cannot tell whether one is running)."
                .to_string(),
        );
    }
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
    // #2936 S-CYCLE: the shared settings the manifest proves are the Go tool's move themselves,
    // so what they name is not theirs to hold: the judgement "referenced by settings" is made
    // without their body. The local settings never move and keep holding.
    let verdict = recorded.get(SETTINGS).map(|e| settings_proof(&project, e));
    let proven = verdict.as_ref().and_then(|v| v.as_ref().ok());
    let holding: Vec<(&'static str, String)> = settings
        .iter()
        .filter(|(f, _)| !(*f == SETTINGS && proven.is_some()))
        .cloned()
        .collect();
    // Moving `.moai/` also moves every git worktree registered under it, and git then loses
    // them (their recorded paths no longer exist). Unreadable means "maybe": stay.
    let checked = version.is_some() && folder_is_dir;
    let local_references = settings
        .iter()
        .find(|(f, text)| checked && *f == SETTINGS_LOCAL && text.contains(FOLDER_REFERENCE))
        .map(|(_, text)| {
            let (references, mut blocked) = scan_local(text, &project.to_string_lossy());
            blocked.extend(
                holding
                    .iter()
                    .filter(|(f, t)| *f != SETTINGS_LOCAL && t.contains(FOLDER_REFERENCE))
                    .map(|(f, _)| format!("{f} also names the folder and keeps it")),
            );
            LocalReferences {
                file: SETTINGS_LOCAL.into(),
                pick: REWRITE_PICK,
                picked: rewrite_local,
                references,
                blocked,
            }
        });
    let rewrites = rewrite_local
        && local_references
            .as_ref()
            .is_some_and(|l| l.blocked.is_empty());
    // The local settings the pick rewrites no longer hold the folder; they still hold every
    // other file they name.
    let holds_folder = |file: &str, text: &str| {
        text.contains(FOLDER_REFERENCE) && !(rewrites && file == SETTINGS_LOCAL)
    };
    let referenced_by_settings = holding.iter().any(|(f, text)| holds_folder(f, text));
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
        plan_state_moves(&project, &[])?
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
                let way = match &local_references {
                    None => String::new(),
                    Some(l) if l.blocked.is_empty() => format!(
                        "; `{REWRITE_PICK}` on go-state-plan and go-state-apply backs up \
                         {SETTINGS_LOCAL} and points it at the Go backup folder"
                    ),
                    Some(_) => format!(
                        "; `{REWRITE_PICK}` cannot rewrite it yet (see local_references.blocked)"
                    ),
                };
                format!(
                    "settings reference .moai/; the folder stays until they are cleaned \
                     (#2581){way}"
                )
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
                         their repositories off from them, so the .moai/ folder stays where it \
                         is and is not moved into a backup until they are moved out"
                    ),
                    (_, None) => stopped_walk_reason(&project, unrecorded_limit()),
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
            "Stop moai commands and agent sessions in this project before the apply: the \
             Go .moai/ becomes {go_backup}, a new .moai/ is made and the {LEGACY_ROOT} product \
             state moves into it."
        ));
    }
    let mut keys: Vec<&String> = recorded.keys().collect();
    keys.sort();
    for key in keys {
        // S1c-1: the `.codex` records are classified together with the files this generation
        // made there (see `classify_codex` below).
        if top_of(key) == CODEX_DIR {
            continue;
        }
        // O6: a secret name under `.claude/` or `.agents/` is never listed, hashed or moved;
        // `secrets` counts what is on disk.
        if matches!(top_of(key), ".claude" | ".agents") && has_secret_component(key) {
            continue;
        }
        let mut content = String::new();
        let (action, reason, class, proof) = policy(
            &project,
            key,
            &recorded[key],
            &holding,
            verdict.as_ref().filter(|_| key == SETTINGS),
            folder_moves,
            &mut content,
        );
        let mut item = new_item(key, action, reason.into(), class);
        item.proof = proof;
        item.deployed_hash = recorded[key].deployed_hash.clone();
        item.content = content;
        // A file only the moving settings named: said once the gate below has the last word.
        if action == "move" && key != SETTINGS {
            if let Some(body) = proven.and(settings.iter().find(|(f, _)| *f == SETTINGS)) {
                if referenced(&project, &body.1, key).is_some() {
                    item.freed_by = SETTINGS.into();
                }
            }
        }
        items.push(item);
    }
    if version.is_some() {
        items.extend(classify_agents_skills(&project, &recorded, folder_moves));
        let (codex_items, codex_warnings) = classify_codex(&project, &recorded);
        items.extend(codex_items);
        warnings.extend(codex_warnings);
    }
    let (secrets, secrets_stopped) = count_secrets(&project, unrecorded_limit());
    if secrets_stopped {
        warnings.push(format!(
            "the secret-name count under .claude/, .agents/ and .codex/ is a lower bound: its walk \
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
                 {limit} entries under .moai/, .claude/, .agents/ and .codex/; the numbers in \
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
        plan_restores(&project, go_backup, &recorded)
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
    let mut hidden = None;
    let legacy_moves = if state.generation == "none" {
        let (mut moves, mut block) = plan_state_moves(&project, picks)?;
        // A `.moai/` that holds the team's kept names (beyond the SPECs alone): the tracked
        // file counts and the `.gitignore` check of the targets (REQ-5).
        let moai = project.join(MOAI_DIR);
        let beside_kept =
            crate::state_root::kept_only(&moai) && !crate::state_root::specs_only(&moai);
        if beside_kept && holds_more_than_specs_and_local(&moai) {
            let (ignored, note) = kept_git_checks(&project, &mut moves);
            warnings.extend(note);
            if let Some(why) = ignored {
                block.get_or_insert(why);
            }
        } else if block.is_none() {
            // An absent `.moai/`, or one with only the SPECs and the local folder: the moves
            // stay in the preview and the user decides (SPEC-20261126-tracked-state-move).
            hidden = ignored_tracked(&project, &moves);
        }
        let any = moves
            .iter()
            .any(|m| m.action == "move" || m.action == "set-aside");
        // A SPEC conflict (RB1) is shown even when nothing else would move.
        let conflict = moves.iter().any(|m| m.action == "conflict");
        if let (true, Some(block)) = (any || conflict, &block) {
            items.push(new_item(
                LEGACY_ROOT,
                "leave",
                block.replace("so the Go folder stays", "so it stays where it is"),
                "legacy_state",
            ));
            Vec::new()
        } else if any {
            // The plan's words stay the same with or without a SPECs-only `.moai/` (SDD B1:
            // `spec new` changes no plan answer); the apply's receipt message says which.
            if beside_kept {
                // The `.moai/` stays as it is: the earlier state moves beside the names it holds
                // and the marker is written last (REQ-6).
                items.push(new_item(
                    LEGACY_ROOT,
                    "move",
                    "the earlier root's product state moves beside the kept names of the \
                     existing .moai/, which stay as they are (REQ-6)"
                        .into(),
                    "legacy_state",
                ));
                warnings.push(format!(
                    "Stop moai commands and agent sessions in this project before the apply: \
                     the {LEGACY_ROOT} product state moves beside the names the existing .moai/ \
                     holds; no new .moai/ is made."
                ));
            } else {
                items.push(new_item(
                    LEGACY_ROOT,
                    "move",
                    "the earlier root's product state moves into a new .moai/ (#2696 C4a)".into(),
                    "legacy_state",
                ));
                warnings.push(format!(
                    "Stop moai commands and agent sessions in this project before the apply: a \
                     new .moai/ is made and the {LEGACY_ROOT} product state moves into it."
                ));
            }
            moves
        } else {
            Vec::new()
        }
    } else if picks.is_empty() {
        Vec::new()
    } else {
        return Err(format!(
            "--set-aside is for a move of the earlier {LEGACY_ROOT}/ into a {MOAI_DIR}/ that \
             holds only the team's kept names; state_root is `{}` here; nothing planned",
            state.generation
        ));
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
    if rewrites {
        for row in root_files
            .iter_mut()
            .filter(|f| f.path == SETTINGS_LOCAL && f.action == "copy")
        {
            row.action = "rewrite".into();
            row.reason = format!(
                "{REWRITE_PICK}: its original bytes go into the backup, then it is rewritten in \
                 place"
            );
            warnings.push(format!(
                "{SETTINGS_LOCAL} will name {} instead of {MOAI_DIR}: its hooks and status \
                 line then run the files in that folder, so keep the folder; its deny and ask \
                 entries stay too and then cover the new {MOAI_DIR}/",
                go_backup.as_deref().unwrap_or(crate::state_root::GO_BACKUP)
            ));
        }
    }
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
    // S-CYCLE: `freed_by` stands only for an item that still moves, under settings that move
    // (the `state_root` gate above may have held everything).
    let settings_moves = items
        .iter()
        .any(|i| i.path == SETTINGS && i.action == "move");
    for i in items.iter_mut() {
        i.kind = kind_of(&i.path).into();
        if i.action != "move" || !settings_moves {
            i.freed_by.clear();
        }
    }
    let (kinds, counterparts) = plan_counterparts(&project, &items, &settings);
    let after_move = plan_after_move(&project, &items, &settings);
    // O-2 (#2936): the skill mirror folders may end up empty; the apply deletes no folder, so
    // the preview says so and the user decides.
    for f in after_move
        .empty_folders
        .iter()
        .filter(|f| matches!(f.as_str(), ".agents/skills" | ".agents"))
    {
        warnings.push(format!(
            "{f} holds nothing once the move is done; the apply deletes no folder (O-2), remove \
             it yourself if you do not want it"
        ));
    }
    let (layout, backup_to) =
        layout_and_backup(&project, &items, &state_moves, &root_files, &restores);
    let settings_move = settings_moves
        .then(|| {
            let item = items.iter().find(|i| i.path == SETTINGS)?;
            let body = settings
                .iter()
                .find(|(f, _)| *f == SETTINGS)
                .map_or("", |(_, t)| t.as_str());
            let mut freed: Vec<String> = items
                .iter()
                .filter(|i| i.freed_by == SETTINGS)
                .map(|i| i.path.clone())
                .collect();
            freed.sort();
            let folder_moves_now = items
                .iter()
                .any(|i| i.path == MOAI_DIR && i.action == "move");
            Some(SettingsMove {
                path: SETTINGS.into(),
                proof: item.proof.clone(),
                edited: proven.is_some_and(|p| p.edited),
                layout: layout.into(),
                freed,
                folder_freed: folder_moves_now && body.contains(".moai/"),
                still_holding: holding
                    .iter()
                    .filter(|(f, t)| holds_folder(f, t))
                    .map(|(f, _)| f.to_string())
                    .collect(),
            })
        })
        .flatten();
    if let Some(sm) = &settings_move {
        warnings.push(settings_move_warning(sm, after_move.wiring_lost.as_ref()));
    }
    // S-AGENTS-STAY U1: with the Go folder staying, the moved files go to the current state
    // folder's backup area, which a `.gitignore` already there must keep out of git.
    let moves: Vec<&str> = items
        .iter()
        .filter(|i| moving_path(i))
        .map(|i| i.path.as_str())
        .collect();
    // Nothing to move means the apply writes no backup (it answers `unchanged` first), so
    // there is nothing for git to offer and nothing to stop.
    let backup_ignore = if layout == "folder_stays" && !moves.is_empty() {
        backup_ignore(
            &project,
            crate::state_root::dir(&project),
            &moves,
            EXAMPLE_STAMP,
        )
        .map(|mut b| {
            for e in &mut b.examples {
                *e = e.replace(EXAMPLE_STAMP, "<UTC>");
            }
            b
        })
    } else {
        None
    };
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
        layout,
        backup_to,
        state_moves,
        root_files,
        left_in_backup,
        kinds,
        counterparts,
        after_move,
        settings_move,
        local_references,
        backup_ignore,
        ignored_tracked: hidden,
        secrets,
        warnings,
        plan_digest: String::new(),
    };
    plan.plan_digest = bound_digest(&plan)?;
    Ok(plan)
}

/// #2936 S-CYCLE: the warning that comes with a settings move. Counts only: never a deny rule,
/// a command or a value of the file.
fn settings_move_warning(sm: &SettingsMove, wiring: Option<&WiringLost>) -> String {
    let mut carried = Vec::new();
    if let Some(w) = wiring {
        if w.go_hook_entries > 0 {
            carried.push(format!("{} Go hook entries", w.go_hook_entries));
        }
        if w.deny_rules > 0 {
            carried.push(format!("{} deny rules", w.deny_rules));
        }
        if w.keys.iter().any(|k| k == "statusLine") {
            carried.push("a status line".to_string());
        }
    }
    let with = if carried.is_empty() {
        String::new()
    } else {
        format!(" with {}", carried.join(", "))
    };
    let freed = if sm.freed.is_empty() {
        String::new()
    } else {
        format!(
            " {} file(s) that only these settings named move with it.",
            sm.freed.len()
        )
    };
    let folder = if sm.folder_freed {
        " The Go .moai/ folder, which they named, moves too."
    } else {
        ""
    };
    format!(
        "{} moves into the backup{with}; the install phase writes MoAI's status line \
         (claude-settings) and hook set (hooks), and claude-settings-merge offers the deny \
         rules and your other keys back with your go-ahead.{freed}{folder}",
        sm.path
    )
}

// ---- S-DIGEST: what `plan_digest` binds (#2936, plan 6887264857 §3) ----
//
// The digest covers what the apply moves, copies or writes, and nothing the preview only
// shows. Hashing the whole preview made any harmless change between preview and apply (a new
// log file, a field another lane added) stale, while the content of a moving file was read
// again at apply time, so an edit of an already edited file was not caught. The allow-list
// below fixes both: each moving file's content hash is bound, and the apply refuses when a
// fresh preview binds anything else.

/// Version of the [`Bound`] layout. The field order below is the hash order: never reorder or
/// add a field without bumping this, so an older digest is refused as a stale plan instead of
/// matching by luck.
const BOUND_VERSION: u32 = 2;

/// What `plan_digest` hashes, as JSON.
#[derive(Serialize)]
struct Bound {
    bound_version: u32,
    project: String,
    /// The Go backup name, or the stamped placeholder when the plain name is taken.
    go_backup: Option<String>,
    /// The items with `action` `move`, by path.
    moves: Vec<BoundMove>,
    /// The restore rows the apply copies (`copy`) or stops on (`conflict*`), by `from`.
    restores: Vec<BoundRestore>,
    /// The earlier root's entries that move, by `from`.
    state_moves: Vec<BoundState>,
    /// The root files the apply moves or copies, by path.
    root_files: Vec<BoundRoot>,
}

#[derive(Serialize)]
struct BoundMove {
    path: String,
    action: String,
    class: String,
    deployed_hash: String,
    /// [`Item::content`]: what the move carries.
    content: String,
}

#[derive(Serialize)]
struct BoundRestore {
    from: String,
    to: String,
    action: String,
    sha256: String,
}

#[derive(Serialize)]
struct BoundState {
    from: String,
    to: String,
    /// `move` or `set-aside`: a pick changes the row, so the digest changes with the picks.
    action: String,
}

#[derive(Serialize)]
struct BoundRoot {
    path: String,
    action: String,
    sha256: String,
}

/// Picks the bound part of a plan. Every plan type is destructured without `..`, so a field
/// added to one of them does not compile until it is placed here (bound) or in the
/// preview-only list beside it.
fn bound_of(plan: &GoStatePlan) -> Result<Bound> {
    let GoStatePlan {
        // Bound: what the apply acts on.
        project,
        items,
        restores,
        go_backup,
        state_moves,
        root_files,
        // Preview-only: the apply recomputes them (`backup_to` from `state_root::dir`) or never
        // reads them.
        schema_version: _,
        manifest_version: _,
        unrecorded: _,
        layout: _,
        backup_to: _,
        left_in_backup: _,
        kinds: _,
        counterparts: _,
        after_move: _,
        settings_move: _,
        local_references: _,
        backup_ignore: _,
        ignored_tracked: _,
        secrets: _,
        warnings: _,
        plan_digest: _,
    } = plan;
    let mut moves = Vec::new();
    for item in items {
        let Item {
            // Bound.
            path,
            action,
            class,
            deployed_hash,
            content,
            // Preview-only: why and how the preview shows the path.
            reason: _,
            kind: _,
            proof: _,
            git: _,
            freed_by: _,
        } = item;
        if action != "move" {
            continue;
        }
        // The two folders move by one rename and their content cannot be pinned (the receipt
        // holds the tree digest taken after the move). Any other move must carry the hash of
        // what it moves: fail closed, so a future kind of move cannot slip in unbound.
        if content.is_empty() && path != MOAI_DIR && path != LEGACY_ROOT {
            return Err(format!(
                "{path}: a moving path without a content hash; refused"
            ));
        }
        moves.push(BoundMove {
            path: path.clone(),
            action: action.clone(),
            class: class.clone(),
            deployed_hash: deployed_hash.clone(),
            content: content.clone(),
        });
    }
    moves.sort_by(|a, b| a.path.cmp(&b.path));
    let mut bound_restores = Vec::new();
    for row in restores {
        let Restore {
            from,
            to,
            sha256,
            action,
            // Preview-only: why a row is not copied.
            reason: _,
        } = row;
        // `excluded_*`, `leave_link` and the like are never acted on.
        if action == "copy" || action.starts_with("conflict") {
            bound_restores.push(BoundRestore {
                from: from.clone(),
                to: to.clone(),
                action: action.clone(),
                sha256: sha256.clone(),
            });
        }
    }
    bound_restores.sort_by(|a, b| a.from.cmp(&b.from));
    let mut bound_state = Vec::new();
    for row in state_moves {
        let StateMove {
            from,
            to,
            action,
            // Preview-only.
            reason: _,
            tracked: _,
        } = row;
        if action == "move" || action == "set-aside" {
            bound_state.push(BoundState {
                from: from.clone(),
                to: to.clone(),
                action: action.clone(),
            });
        }
    }
    bound_state.sort_by(|a, b| a.from.cmp(&b.from));
    let mut bound_roots = Vec::new();
    for row in root_files {
        let RootFile {
            path,
            sha256,
            action,
            // Preview-only: where the copy goes (the apply derives it), the later choice, why,
            // what the file is and what git says.
            backup_copy: _,
            user_choice: _,
            reason: _,
            class: _,
            go_proof: _,
            git: _,
        } = row;
        if action != "leave" {
            bound_roots.push(BoundRoot {
                path: path.clone(),
                action: action.clone(),
                sha256: sha256.clone(),
            });
        }
    }
    bound_roots.sort_by(|a, b| a.path.cmp(&b.path));
    Ok(Bound {
        bound_version: BOUND_VERSION,
        project: project.clone(),
        go_backup: go_backup.clone(),
        moves,
        restores: bound_restores,
        state_moves: bound_state,
        root_files: bound_roots,
    })
}

/// The `plan_digest` of a plan: `sha256:<hex>` over the [`Bound`] part.
fn bound_digest(plan: &GoStatePlan) -> Result<String> {
    let bytes = serde_json::to_vec(&bound_of(plan)?).map_err(|_| "encode failed")?;
    Ok(digest(&bytes))
}

// ---- stage 2: apply and revert ----
//
// Two layouts, chosen by the preview's `.moai` item:
//
// * The Go folder moves (#2696). In this order, against one plan digest:
//   1. `.moai` → the Go backup (`.moai-backup`, or `.moai-backup-<UTC>` when taken), one
//      rename that never overwrites;
//   2. a new `.moai/` with the receipt `.moai/backups/<UTC>/go-state-receipt.json`, `pending`
//      (when this step fails, the rename is undone at once; the ignore file of the backup
//      area, written right after it, is outside that undo: its failure says `partial apply`
//      and names the revert plan);
//   3. the proven `.claude/` files move to `.moai/backups/<UTC>/go-state/<same path>`;
//   4. the earlier root's product state moves, `.moai-cli/<entry>` → `.moai/<entry>`;
//   5. SPECs are copied from the backup into `.moai/specs`;
//   6. `CLAUDE.md`/`AGENTS.md`/`AGENTS.local.md` move to `.moai/backups/<UTC>/root-files/`, a
//      0600 copy of the local settings goes there too;
//   7. the marker `.moai/moai.json` (the write name; see `state_root::NEW_MARKER`) is
//      written last, then the receipt says `committed`.
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
            // The hash the preview took and the digest bound; `still_as_previewed` checks the
            // file again right before it moves.
            m.sha256 = item.content.clone();
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
    /// SPEC-20261121-project-state-move D-7: the names a `.moai/` that held only the team's kept
    /// names had before the apply. The apply kept that folder and moved the state in beside
    /// them, so the revert leaves these names where they are and sets nothing aside. Empty
    /// (and absent) for every earlier receipt, which reads as it always did.
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub kept_before: Vec<String>,
    /// SPEC-20261126-go-local-settings-reference: the local settings the apply rewrote, from
    /// just before the replace; the original is the `copied` entry of the same sha256.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub rewritten: Option<Rewritten>,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
#[serde(deny_unknown_fields)]
pub struct Rewritten {
    pub path: String,
    pub original_sha256: String,
    pub rewritten_sha256: String,
    /// The permission bits the file had, for a revert that finds it gone.
    pub mode: u32,
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
    // The writer keeps the limit its reader uses, before anything is created: a receipt that
    // could not be read back is never written.
    if bytes.len() as u64 > RECEIPT_MAX {
        return Err(format!(
            "receipt larger than {} MiB; refused; nothing was written",
            RECEIPT_MAX / (1024 * 1024)
        ));
    }
    let path = project.join(&rel);
    match before {
        None => crate::apply::publish_exclusive_mode(&path, &bytes, Some(PRIVATE_FILE))?,
        // The replace re-reads the old receipt: with the receipt's own limit, not the 256 KiB
        // of an ordinary target (which reported a size as "target changed during apply").
        Some(b) => crate::apply::publish_replace_limited(&path, Some(b), &bytes, RECEIPT_MAX)?,
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
pub(crate) fn make_dirs(project: &Path, rel: &str) -> Result<()> {
    make_dirs_with(project, rel, false)
}

/// [`make_dirs`] for a backup area: every folder it creates is 0700.
pub(crate) fn make_private_dirs(project: &Path, rel: &str) -> Result<()> {
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

/// [`make_dirs`] for a restore (#2966 F3): `to_dir` is under the new `.moai/`, `from_dir` is the
/// same folder in the Go backup, with as many components. A folder that has to be created is
/// 0700 when the folder it comes from has no group or other bits, and an ordinary one
/// otherwise; a folder that exists is left as it is.
fn make_dirs_like(project: &Path, to_dir: &str, from_dir: &str) -> Result<()> {
    let from: Vec<_> = Path::new(from_dir).components().collect();
    let mut cur = project.to_path_buf();
    let mut src = project.to_path_buf();
    for (i, part) in Path::new(to_dir).components().enumerate() {
        cur.push(part);
        if let Some(s) = from.get(i) {
            src.push(s);
        }
        match fs::symlink_metadata(&cur) {
            Ok(m) if m.file_type().is_symlink() => {
                return Err(format!("{to_dir}: a path component is a symlink; refused"))
            }
            Ok(m) if m.is_dir() => {}
            Ok(_) => return Err(format!("{to_dir}: a path component is not a folder")),
            Err(_) => {
                let private = from.len() == Path::new(to_dir).components().count()
                    && mode_bits(&src).is_some_and(|m| m & 0o077 == 0);
                create_dir_with(&cur, private)
                    .map_err(|_| format!("{to_dir}: folder create failed"))?
            }
        }
    }
    Ok(())
}

/// `<base>/backups/.gitignore` holding `*`: nothing under the backup area is ever offered to
/// git, whose copies may hold tokens. `base` is the folder that keeps the area: `.moai` in the
/// moving layouts, the current state folder when the Go folder stays (AUDIT #3077 B1: the
/// proven settings move there too). An existing file is left as it is.
fn keep_backups_out_of_git(project: &Path, base: &str) -> Result<()> {
    let rel = format!("{base}/backups/.gitignore");
    refuse_linked(project, &rel)?;
    match fs::symlink_metadata(project.join(&rel)) {
        Ok(_) => Ok(()),
        Err(_) => crate::apply::publish_exclusive(&project.join(&rel), b"*\n"),
    }
}

/// A stamp of the real shape, for the preview's git question before any apply has a stamp.
const EXAMPLE_STAMP: &str = "19700101T000000Z";
/// How many unignored backup paths the preview names.
const IGNORE_EXAMPLES: usize = 3;
/// Paths per `git check-ignore` call, so a long move list never reaches the argument limit.
const IGNORE_BATCH: usize = 64;
/// How many hidden tracked files an [`IgnoredTracked`] names (decision D-5).
const IGNORED_FILES: usize = 20;

/// What git says of a set of backup paths.
#[derive(Debug, PartialEq, Eq)]
enum IgnoreVerdict {
    /// Every path is ignored, or the project is not a git work tree (git offers nothing).
    Covered,
    /// These paths git would offer.
    Offered(Vec<String>),
    /// Git could not say; the reason holds no path.
    Unknown(String),
}

/// Runs `git --no-optional-locks -C <project> <args…> <paths…>` with the shared environment
/// scrub and no stdin.
fn git_in(
    project: &Path,
    args: &[&str],
    paths: &[String],
) -> std::io::Result<std::process::Output> {
    let mut cmd = std::process::Command::new("git");
    crate::identity::scrub_git_env(&mut cmd);
    cmd.arg("--no-optional-locks")
        .arg("-C")
        .arg(project)
        .args(args)
        .args(paths)
        .env("GIT_OPTIONAL_LOCKS", "0")
        .stdin(std::process::Stdio::null())
        .output()
}

/// One line of `git check-ignore --no-index -v -n` as "is this path ignored"; `None` when the
/// line is not in that shape. A match reads `<source>:<line>:<pattern>\t<path>` (the source can
/// hold a colon, so the line number is the first `:<digits>:`), a miss reads `::\t<path>`, and
/// a pattern that starts with `!` is the rule that un-ignores the path.
fn check_ignore_line(line: &str) -> Option<bool> {
    let (head, _path) = line.split_once('\t')?;
    if head == "::" {
        return Some(false);
    }
    for (i, _) in head.match_indices(':') {
        let rest = &head[i + 1..];
        let digits = rest.bytes().take_while(u8::is_ascii_digit).count();
        if digits > 0 && rest.as_bytes().get(digits) == Some(&b':') {
            let pattern = &rest[digits + 1..];
            return (!pattern.is_empty()).then(|| !pattern.starts_with('!'));
        }
    }
    None
}

/// Asks git whether each of `paths` (relative to `project`) is ignored. The paths need not
/// exist. Nothing is written and no file is read here; the rules are git's own, so a literal
/// look at the `.gitignore` text is never the judge.
fn ignored_by_git(project: &Path, paths: &[String]) -> IgnoreVerdict {
    let inside = match git_in(project, &["rev-parse", "--is-inside-work-tree"], &[]) {
        Err(_) => return IgnoreVerdict::Unknown("git could not be run".into()),
        Ok(o) if o.status.success() => String::from_utf8_lossy(&o.stdout).trim().to_string(),
        Ok(o) => {
            let why = crate::identity::failure_summary(&o.stderr);
            return if why.starts_with(crate::identity::NOT_A_REPOSITORY) {
                IgnoreVerdict::Covered
            } else {
                IgnoreVerdict::Unknown(why)
            };
        }
    };
    match inside.as_str() {
        "true" => {}
        // A bare repository or the inside of `.git`: git has no work tree to offer files from.
        "false" => return IgnoreVerdict::Covered,
        _ => return IgnoreVerdict::Unknown("git gave an unexpected answer".into()),
    }
    let mut offered = Vec::new();
    for batch in paths.chunks(IGNORE_BATCH) {
        let out = match git_in(
            project,
            &["check-ignore", "--no-index", "-v", "-n", "--"],
            batch,
        ) {
            Err(_) => return IgnoreVerdict::Unknown("git could not be run".into()),
            Ok(o) => o,
        };
        // 0: some path is ignored, 1: none is; anything else is git failing.
        if !matches!(out.status.code(), Some(0 | 1)) {
            return IgnoreVerdict::Unknown(crate::identity::failure_summary(&out.stderr));
        }
        let Ok(text) = String::from_utf8(out.stdout) else {
            return IgnoreVerdict::Unknown("git output was not UTF-8".into());
        };
        let lines: Vec<&str> = text.lines().collect();
        if lines.len() != batch.len() {
            return IgnoreVerdict::Unknown("git answered for a different number of paths".into());
        }
        for (path, line) in batch.iter().zip(lines) {
            match check_ignore_line(line) {
                Some(true) => {}
                Some(false) => offered.push(path.clone()),
                None => return IgnoreVerdict::Unknown("git gave an unexpected answer".into()),
            }
        }
    }
    if offered.is_empty() {
        IgnoreVerdict::Covered
    } else {
        IgnoreVerdict::Offered(offered)
    }
}

/// SPEC-20261126-tracked-state-move REQ-1, REQ-3: the git check of a move into a `.moai/` that
/// is absent or holds only the SPECs and the local folder. `None` when the move may be offered:
/// every destination is allowed, no moving entry holds a tracked file, or the project is in no
/// git work tree. Git being unable to answer is `unknown`: the check stops closed. Reads only.
fn ignored_tracked(project: &Path, moves: &[StateMove]) -> Option<IgnoredTracked> {
    let moving: Vec<&StateMove> = moves.iter().filter(|m| m.action == "move").collect();
    if moving.is_empty() {
        return None;
    }
    hidden_by_move(project, &moving).unwrap_or_else(|why| {
        Some(IgnoredTracked {
            state: "unknown",
            files: Vec::new(),
            total: 0,
            marker: None,
            reason: format!(
                "git could not say whether the move would put tracked state where it is ignored \
                 ({why}); nothing moves until it can; then preview again"
            ),
        })
    })
}

/// Each tracked file under a moving entry is asked of `git check-ignore` at the path it would
/// get, and so is the marker when there is such a file. An error is git failing to answer.
fn hidden_by_move(project: &Path, moving: &[&StateMove]) -> Result<Option<IgnoredTracked>> {
    let no_git = |_: std::io::Error| "git could not be run".to_string();
    match git_in(project, &["rev-parse", "--is-inside-work-tree"], &[]) {
        Ok(o) if o.status.success() => match String::from_utf8_lossy(&o.stdout).trim() {
            "true" => {}
            // A bare repository or the inside of `.git`: no work tree holds these files.
            "false" => return Ok(None),
            _ => return Err("git gave an unexpected answer".into()),
        },
        Ok(o) => {
            let why = crate::identity::failure_summary(&o.stderr);
            return if why.starts_with(crate::identity::NOT_A_REPOSITORY) {
                Ok(None)
            } else {
                Err(why)
            };
        }
        // Without a git program, a folder with no `.git` here or above is in no work tree all
        // the same.
        Err(e) => {
            let in_work_tree = project
                .ancestors()
                .any(|dir| fs::symlink_metadata(dir.join(".git")).is_ok());
            return if in_work_tree {
                Err(no_git(e))
            } else {
                Ok(None)
            };
        }
    }
    let out = git_in(project, &["ls-files", "-z", "--", LEGACY_ROOT], &[]).map_err(no_git)?;
    if !out.status.success() {
        return Err(format!(
            "git could not list the tracked files: {}",
            crate::identity::failure_summary(&out.stderr)
        ));
    }
    let listed = String::from_utf8(out.stdout)
        .map_err(|_| "git listed a file name that is not UTF-8".to_string())?;
    // Each tracked file under a moving entry, with the path the move would give it.
    let files: Vec<(&str, String)> = listed
        .split('\0')
        .filter_map(|file| {
            moving.iter().find_map(|m| {
                let rest = file.strip_prefix(m.from.as_str())?;
                (rest.is_empty() || rest.starts_with('/'))
                    .then(|| (file, format!("{}{rest}", m.to)))
            })
        })
        .collect();
    if files.is_empty() {
        return Ok(None);
    }
    let marker = format!("{MOAI_DIR}/{}", crate::state_root::RUST_MARKER);
    let mut paths: Vec<String> = files.iter().map(|(_, to)| to.clone()).collect();
    paths.push(marker.clone());
    let mut rules = Vec::with_capacity(paths.len());
    for batch in paths.chunks(IGNORE_BATCH) {
        let out = git_in(
            project,
            &["check-ignore", "--no-index", "-v", "-n", "--"],
            batch,
        )
        .map_err(no_git)?;
        // 0: some path is ignored, 1: none is; anything else is git failing.
        if !matches!(out.status.code(), Some(0 | 1)) {
            return Err(format!(
                "git could not judge the destinations: {}",
                crate::identity::failure_summary(&out.stderr)
            ));
        }
        let text = String::from_utf8_lossy(&out.stdout);
        let lines: Vec<&str> = text.lines().collect();
        if lines.len() != batch.len() {
            return Err("git answered for a different number of destinations".into());
        }
        for line in lines {
            if check_ignore_line(line).is_none() {
                return Err("git gave an unexpected answer".into());
            }
            rules.push(ignoring_rule(line));
        }
    }
    let marker_rule = rules.pop().flatten();
    let mut hidden: Vec<HiddenFile> = files
        .into_iter()
        .zip(rules)
        .filter_map(|((file, to), rule)| {
            rule.map(|rule| HiddenFile {
                file: file.to_string(),
                to,
                rule,
            })
        })
        .collect();
    let total = hidden.len();
    if total == 0 && marker_rule.is_none() {
        return Ok(None);
    }
    hidden.truncate(IGNORED_FILES);
    let mut found = Vec::new();
    if total > 0 {
        let named: Vec<String> = hidden
            .iter()
            .map(|h| format!("{} -> {} (rule `{}`)", h.file, h.to, h.rule))
            .collect();
        let cut = if total > hidden.len() {
            format!("; the first {} are named", hidden.len())
        } else {
            String::new()
        };
        found.push(format!(
            "{total} tracked file(s) would move where git ignores them: {}{cut}",
            named.join(", ")
        ));
    }
    if let Some(rule) = &marker_rule {
        found.push(format!(
            "the marker {marker} the team keeps in git is ignored by the rule `{rule}`"
        ));
    }
    Ok(Some(IgnoredTracked {
        state: "ignored",
        files: hidden,
        total,
        marker: marker_rule,
        reason: format!(
            "{}; nothing moves until the rules allow these paths (a `!` line) or git stops \
             tracking the files under {LEGACY_ROOT}/; then preview again",
            found.join("; ")
        ),
    }))
}

/// #2936 S-AGENTS-STAY U1: whether the `<base>/backups/.gitignore` that is already there keeps
/// the staying layout's backup copies out of git. They are the receipt and one copy per moving
/// file under `<base>/backups/<stamp>/`, and the settings among them may hold a token.
/// `None` when nothing needs deciding: no such file (the apply writes `*`), a link on the way
/// (the apply refuses it as before), git ignores every path, or the project is not a git work
/// tree. Git being unable to answer is `unknown`: the check stops closed.
fn backup_ignore(project: &Path, base: &str, moves: &[&str], stamp: &str) -> Option<BackupIgnore> {
    let file = format!("{base}/backups/.gitignore");
    if linked_component(project, &file) || fs::symlink_metadata(project.join(&file)).is_err() {
        return None;
    }
    let area = format!("{base}/backups/{stamp}");
    let mut paths = vec![format!("{area}/{RECEIPT}")];
    paths.extend(moves.iter().map(|m| format!("{area}/go-state/{m}")));
    match ignored_by_git(project, &paths) {
        IgnoreVerdict::Covered => None,
        IgnoreVerdict::Offered(offered) => Some(BackupIgnore {
            state: "not_ignored",
            reason: format!(
                "{file} exists and git does not ignore the backup copies under {base}/backups/: \
                 they may hold tokens, so nothing moves until that file ignores them; MoAI never \
                 edits it (add a line such as `*` to it, or remove it and the apply writes one)"
            ),
            file,
            examples: offered.into_iter().take(IGNORE_EXAMPLES).collect(),
        }),
        IgnoreVerdict::Unknown(why) => Some(BackupIgnore {
            state: "unknown",
            reason: format!(
                "{file} exists and git could not say whether it ignores the backup copies \
                 ({why}); nothing moves until it can; MoAI never edits that file"
            ),
            file,
            examples: Vec::new(),
        }),
    }
}

/// The apply's refusal for a [`BackupIgnore`]. It carries the file name and a path, never a line
/// of the file; the classifier reads `refused`.
fn backup_ignore_refusal(b: &BackupIgnore) -> String {
    let named = if b.examples.is_empty() {
        String::new()
    } else {
        format!(" (for example {})", b.examples.join(", "))
    };
    format!(
        "backup area not kept out of git; refused: {}{named}; nothing written",
        b.reason
    )
}

/// F-I (#2936): the ignore file is written after the pending receipt, so a failure there is
/// outside the step-2 undo. `done` says what is on disk (the receipt is `pending`, so
/// `state_root` says `interrupted`); the revert plan takes the project back. The phrase
/// `partial apply` makes `contract::classify` say `partial`.
pub(crate) fn ignore_file_failed(e: &str, done: &str, stamp: &str) -> String {
    format!(
        "partial apply: {done}; failed at {MOAI_DIR}/backups/.gitignore: {e}; no file has moved \
         yet; run `moai go-state-revert-plan <project> --backup {stamp}` to put the project back"
    )
}

pub(crate) fn parent_of(rel: &str) -> Result<String> {
    Path::new(rel)
        .parent()
        .and_then(|p| p.to_str())
        .map(str::to_string)
        .ok_or_else(|| format!("{rel}: no parent folder"))
}

/// Move one file without overwriting: hard link, then remove the source.
pub(crate) fn move_exclusive(from: &Path, to: &Path) -> Result<()> {
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

/// S-DIGEST: a file moves as a hard link plus an unlink, so a write between the hash the
/// receipt holds and the move would leave a receipt that names bytes the backup does not have.
/// The file is hashed once more right before it moves; a different hash stops the move.
pub(crate) fn still_as_previewed(from: &Path, rel: &str, sha256: &str) -> Result<()> {
    if sha256_file(from).is_ok_and(|h| h == sha256) {
        return Ok(());
    }
    Err(format!("{rel} changed after the preview; move stopped"))
}

/// Test seam (S-DIGEST): `MOAI_TEST_FAULT=go_state_edit_after_plan` appends one byte to the
/// first file the apply is about to move, standing in for a writer that edits it after the
/// fresh preview and before the move. Debug builds only.
fn test_edit_first_moving_file(project: &Path, moves: &[Moved]) {
    #[cfg(debug_assertions)]
    if std::env::var("MOAI_TEST_FAULT").as_deref() == Ok("go_state_edit_after_plan") {
        if let Some(first) = moves
            .iter()
            .find(|m| matches!(m.kind.as_str(), "" | "edited"))
        {
            if let Ok(mut f) = fs::OpenOptions::new()
                .append(true)
                .open(project.join(&first.path))
            {
                use std::io::Write as _;
                let _ = f.write_all(b"x");
            }
        }
    }
    let _ = (project, moves);
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

/// The apply's refusal when the confirmed digest is not the fresh preview's (F17, #2936: the
/// phrase `stale or altered plan` makes `contract::classify` say `stale_plan`, and the run
/// command is named so the reader knows the next step).
pub(crate) const STALE_PLAN: &str = "stale or altered plan: what go-state-apply would move, copy \
     or write no longer matches the preview that was confirmed (a moving file, a state entry, a \
     restored copy, a root file or the backup name changed, or the digest is from an older \
     version); nothing was written; run `moai go-state-plan <project>` again, read the new \
     preview and confirm again with its plan_digest";

/// The revert's refusal when the confirmed digest is not the fresh revert plan's.
pub(crate) const STALE_REVERT_PLAN: &str = "stale or altered plan: the revert plan no longer \
     matches the confirmed go-state-revert-plan (a path changed or the receipt moved on); \
     nothing was written; run `moai go-state-revert-plan <project> --backup <UTC>` again and \
     confirm again with its plan_digest";

/// #2936 S-CYCLE: the shared settings move last among the files, in both layouts: a stop
/// before them leaves the live host settings in place (the revert brings the rest back).
fn settings_last(moves: &mut Vec<Moved>) {
    if let Some(at) = moves.iter().position(|m| m.path == SETTINGS) {
        let settings = moves.remove(at);
        moves.push(settings);
    }
}

/// #2936 S-CYCLE: what a staying-folder apply may carry. The #2936 kinds (links, copy folders,
/// edited files) go only with the Go folder, except the proven settings, which are `edited`
/// whenever the user changed them since the deploy.
fn stay_kinds_ok(moves: &[Moved]) -> bool {
    moves
        .iter()
        .all(|m| m.kind.is_empty() || (m.kind == "edited" && m.path == SETTINGS))
}

/// Apply the preview whose digest the caller confirmed. Nothing is written unless a fresh
/// preview has exactly that digest.
pub fn apply_with_stamp(
    project: &Path,
    plan_digest: &str,
    permission_reference: &str,
    stamp: &str,
) -> Result<Receipt> {
    apply_picked_with_stamp(project, plan_digest, permission_reference, stamp, &[])
}

/// [`apply_with_stamp`] with the `--set-aside` picks the digest was made with: the fresh
/// preview is made with the same picks, so another set of picks fails the digest check.
pub fn apply_picked_with_stamp(
    project: &Path,
    plan_digest: &str,
    permission_reference: &str,
    stamp: &str,
    picks: &[String],
) -> Result<Receipt> {
    apply_all_with_stamp(
        project,
        plan_digest,
        permission_reference,
        stamp,
        picks,
        false,
    )
}

/// [`apply_picked_with_stamp`] and the [`REWRITE_PICK`]: with it, a fresh preview that cannot
/// rewrite the local settings and move the folder refuses before the first write.
pub fn apply_all_with_stamp(
    project: &Path,
    plan_digest: &str,
    permission_reference: &str,
    stamp: &str,
    picks: &[String],
    rewrite_local: bool,
) -> Result<Receipt> {
    let fresh = plan_picked(project, picks, rewrite_local)?;
    if fresh.plan_digest != plan_digest {
        return Err(STALE_PLAN.into());
    }
    if permission_reference.trim().is_empty() {
        return Err("a permission reference is required".into());
    }
    if !valid_stamp(stamp) {
        return Err("invalid backup stamp".into());
    }
    if rewrite_local {
        refuse_unrewritable(&fresh)?;
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
        kept_before: Vec::new(),
        rewritten: None,
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
    settings_last(&mut receipt.moves);
    test_edit_first_moving_file(&project, &receipt.moves);
    if planned.iter().any(|i| i.path == MOAI_DIR) {
        let mut progress = Progress::default();
        return apply_moving_folder(&project, &fresh, receipt, stamp, &mut progress).map_err(|e| {
            if rewrite_local {
                progress.undo_pick(&project, &e, stamp)
            } else {
                e
            }
        });
    }
    // The policy plans the #2936 kinds only with the Go folder, except the proven settings
    // (S-CYCLE); the stage-2 layout moves unchanged files and those settings alone.
    if !stay_kinds_ok(&receipt.moves) {
        return Err(
            "a staying-folder plan with a link, copy folder or edited file; refused".into(),
        );
    }
    // The Go folder stays: the #2561 stage-2 layout under the current state folder.
    let base = crate::state_root::dir(&project);
    let backup_rel = format!("{base}/backups/{stamp}");
    // S-AGENTS-STAY U1: an ignore file that is already there is never edited, so one that
    // leaves the copies to git (or that git cannot judge) stops the apply before anything is
    // written, `backups/` included. Asked again here with the real stamp and moves.
    let moving: Vec<&str> = receipt.moves.iter().map(|m| m.path.as_str()).collect();
    if let Some(b) = backup_ignore(&project, base, &moving, stamp) {
        return Err(backup_ignore_refusal(&b));
    }
    // `make_dirs` refuses a linked state folder or `backups`; the stamp folder itself is
    // created exclusively, so a link already at that name fails the create.
    make_private_dirs(&project, &format!("{base}/backups"))?;
    // The proven settings may hold a token (`env`): keep the area out of git before the stamp
    // folder, the receipt or any file exists, so a failure here leaves nothing moved (B1).
    keep_backups_out_of_git(&project, base)?;
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
        if rel == SETTINGS {
            crate::update::test_fault("go_state_before_settings")?;
        }
        refuse_linked(&project, &rel)?;
        let dest_rel = format!("{backup_rel}/go-state/{rel}");
        make_private_dirs(&project, &parent_of(&dest_rel)?)?;
        still_as_previewed(&project.join(&rel), &rel, &receipt.moves[i].sha256)?;
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
    progress: &mut Progress,
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
    progress.step = "move the Go folder aside";
    move_folder(&project.join(MOAI_DIR), &project.join(&go_backup))?;
    receipt.moves[0].moved = true;
    // 2. The new folder and the pending receipt. Should either fail, undo the rename now:
    //    nothing else has changed yet.
    progress.step = "make the new folder and its receipt";
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
    crate::update::test_fault("go_state_before_ignore")
        .and_then(|_| keep_backups_out_of_git(project, MOAI_DIR))
        .map_err(|e| {
            ignore_file_failed(
                &e,
                &format!(
                    "the Go folder was renamed to {go_backup} and a new {MOAI_DIR}/ holds the \
                     pending receipt"
                ),
                stamp,
            )
        })?;
    receipt.moves[0].sha256 = crate::baseline::tree_digest(&project.join(&go_backup), &[])?;
    saved = save_receipt(project, MOAI_DIR, &receipt, Some(&saved))?;
    // 3. The proven `.claude/` and `.agents/` entries: files, the mirror links (as links) and
    //    the mirror copy folders (one rename each).
    progress.step = "move the proven files";
    for i in 0..receipt.moves.len() {
        let kind = receipt.moves[i].kind.clone();
        if !matches!(kind.as_str(), "" | "edited" | "link" | "mirror_copy") {
            continue;
        }
        let rel = receipt.moves[i].path.clone();
        if rel == SETTINGS {
            crate::update::test_fault("go_state_before_settings")?;
        }
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
            _ => {
                still_as_previewed(&from, &rel, &receipt.moves[i].sha256)?;
                move_exclusive(&from, &to)?
            }
        }
        receipt.moves[i].moved = true;
        saved = save_receipt(project, MOAI_DIR, &receipt, Some(&saved))?;
    }
    // 4. The earlier root's product state.
    progress.step = "move the earlier state";
    saved = move_state_entries(project, &mut receipt, saved)?;
    // 5. SPECs from the backup (the original stays there), never overwriting.
    progress.step = "restore the SPECs";
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
        make_dirs_like(project, &parent_of(&r.to)?, &parent_of(&src_rel)?)?;
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
    //    and a copy goes there (O3). The ignore file of the backup area was written in step 2,
    //    before anything that may hold a token (test seam: a stop right here).
    progress.step = "back up the root files";
    crate::update::test_fault("go_state_before_root_files")?;
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
        // The pick: only after the original is in the backup, and recorded before the
        // replace so a revert finds it whatever happens next.
        if f.action == "rewrite" {
            progress.step = "rewrite the local settings";
            let text = std::str::from_utf8(&bytes).map_err(|_| "local settings are not UTF-8")?;
            let strings = json_strings(text).ok_or("local settings are not a JSON object")?;
            let after = rewritten_text(text, &strings, &go_backup).into_bytes();
            receipt.rewritten = Some(Rewritten {
                path: f.path.clone(),
                original_sha256: f.sha256.clone(),
                rewritten_sha256: digest(&after),
                mode: mode_bits(&project.join(&f.path)).unwrap_or(PRIVATE_FILE),
            });
            saved = save_receipt(project, MOAI_DIR, &receipt, Some(&saved))?;
            crate::update::test_fault("go_state_before_rewrite")?;
            progress.rewrite = Some((bytes.clone(), after.clone()));
            replace_local(project, &bytes, &after)?;
            crate::update::test_fault("go_state_after_rewrite")?;
        }
    }
    // 7. The marker, last: from here `state_root` says `rust`.
    progress.step = "write the marker";
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
        let aside = receipt.moves[i].kind == "aside";
        if receipt.moves[i].kind != "state" && !aside {
            continue;
        }
        let from = receipt.moves[i].path.clone();
        let entry = from
            .strip_prefix(&format!("{LEGACY_ROOT}/"))
            .ok_or("state entry outside the earlier root")?;
        refuse_linked(project, &from)?;
        // A picked entry goes into the apply's backup area, never over the `.moai/` entry of the
        // same name; a state entry goes beside the kept names.
        let to = if aside {
            let dir = format!("{MOAI_DIR}/backups/{}/{SET_ASIDE}", receipt.backup);
            make_private_dirs(project, &dir)?;
            project.join(dir).join(entry)
        } else {
            project.join(MOAI_DIR).join(entry)
        };
        move_entry(&project.join(&from), &to).map_err(|e| format!("{from}: {e}"))?;
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
    // @cc:careful keep this before the first folder and the receipt: `fresh` judged the
    // destinations again, and a refusal leaves every file and folder as it was
    // (SPEC-20261126-tracked-state-move REQ-4, REQ-5).
    if let Some(found) = &fresh.ignored_tracked {
        return Err(format!(
            "tracked state would leave version control; refused: {}; nothing written",
            found.reason
        ));
    }
    let area = format!("{MOAI_DIR}/backups/{stamp}");
    receipt.legacy_move = true;
    // A C4a receipt carries no Go backup; one that also names a Go backup fails the
    // folder-entry check in `read_receipt` (a Go backup always goes with a folder move).
    for s in fresh
        .state_moves
        .iter()
        .filter(|s| s.action == "move" || s.action == "set-aside")
    {
        receipt.moves.push(Moved {
            path: s.from.clone(),
            kind: if s.action == "move" { "state" } else { "aside" }.into(),
            ..Moved::default()
        });
    }
    receipt.state = "pending".into();
    // A `.moai/` that holds only the project's tracked SPECs is kept: the state moves in beside
    // them (SDD B1). On a failure below, `remove_dir` leaves that folder as it was.
    let moai = project.join(MOAI_DIR);
    let specs_alone = crate::state_root::specs_only(&moai);
    let kept = !specs_alone && crate::state_root::kept_only(&moai);
    let reused = specs_alone || kept;
    // A `.moai/` that holds more kept names than the SPECs is kept too, and the receipt records
    // which names it held, so the revert leaves them (D-7).
    if kept {
        receipt.kept_before = kept_names_in(&moai);
    }
    let made = (if reused {
        Ok(())
    } else {
        fs::create_dir(project.join(MOAI_DIR))
            .map_err(|_| format!("{MOAI_DIR}: folder create failed"))
    })
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
    crate::update::test_fault("go_state_legacy_before_ignore")
        .and_then(|_| keep_backups_out_of_git(project, MOAI_DIR))
        .map_err(|e| {
            // N-f: say which `.moai/` holds the receipt.
            let holder = if reused {
                format!("the {MOAI_DIR}/ beside the project's SPECs holds the pending receipt")
            } else {
                format!("a new {MOAI_DIR}/ holds the pending receipt")
            };
            ignore_file_failed(&e, &holder, stamp)
        })?;
    let saved = move_state_entries(project, &mut receipt, saved)?;
    let marker = crate::state_root::marker_bytes("legacy-move", &[]);
    finish_with_marker(project, receipt, saved, &marker)
}

pub fn apply(project: &Path, plan_digest: &str, permission_reference: &str) -> Result<Receipt> {
    apply_with_stamp(project, plan_digest, permission_reference, &stamp_now())
}

/// [`apply`] with the picks of the preview whose digest is given: `--set-aside` entries and
/// the [`REWRITE_PICK`].
pub fn apply_picked(
    project: &Path,
    plan_digest: &str,
    permission_reference: &str,
    picks: &[String],
    rewrite_local: bool,
) -> Result<Receipt> {
    apply_all_with_stamp(
        project,
        plan_digest,
        permission_reference,
        &stamp_now(),
        picks,
        rewrite_local,
    )
}

#[derive(Clone, Debug, Serialize)]
pub struct RevertItem {
    pub path: String,
    /// `restore`, `set_aside_then_restore` (only `.claude/settings.json`: a regular file is at
    /// the path again; the revert moves it into `<backup area>/replaced-by-revert/<UTC>/` and
    /// then restores the earlier one), `restore_original` /
    /// `set_aside_then_restore_original` (the local settings an apply rewrote: its original
    /// bytes come back from the backup copy, a file at the path going aside first),
    /// `conflict` (something else is at the original path
    /// again), `missing` (the backup copy is gone or changed), `skip` (never moved),
    /// `remove_restored` /
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

/// Whether a receipt may name `path` as a restored copy: a plain relative path under a user-data
/// folder of the new `.moai/` (the moving layouts), or under the earlier root's `specs/`.
fn restored_path_allowed(path: &str, moving: bool) -> bool {
    if !plain_relative(path) {
        return false;
    }
    if moving {
        path.strip_prefix(&format!("{MOAI_DIR}/"))
            .is_some_and(|rel| !rel.is_empty() && !restore_excluded(rel))
    } else {
        path.starts_with(&format!("{LEGACY_ROOT}/specs/"))
    }
}

fn read_receipt(project: &Path, backup: &str) -> Result<(Receipt, Vec<u8>, String)> {
    if !valid_stamp(backup) {
        return Err("invalid backup name: expected the apply's UTC stamp YYYYMMDDTHHMMSSZ".into());
    }
    let base = find_receipt_base(project, backup)
        .ok_or_else(|| format!("no go-state receipt for {backup}"))?;
    let rel = receipt_rel(&base, backup);
    let text = read_bounded_text(project, &rel, RECEIPT_MAX)?
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
            && m.kind != "aside"
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
        // S-CYCLE: the proven settings may be edited in the staying layout too.
        "edited" => (!go_folder && m.path != SETTINGS) || m.path == MOAI_DIR,
        "link" | "mirror_copy" => !go_folder || !mirror_entry(&m.path),
        "root_file" => !go_folder || !ROOT_FILES.contains(&m.path.as_str()),
        "state" => {
            !moving
                || !m
                    .path
                    .strip_prefix(&format!("{LEGACY_ROOT}/"))
                    .is_some_and(|e| STATE_ENTRIES.contains(&e) || is_spec_entry(e))
        }
        // A picked entry set aside beside kept names: one known earlier-root entry, only in
        // the earlier-root move.
        "aside" => {
            !receipt.legacy_move
                || go_folder
                || !m
                    .path
                    .strip_prefix(&format!("{LEGACY_ROOT}/"))
                    .is_some_and(|e| STATE_ENTRIES.contains(&e))
        }
        _ => true,
    }) {
        return Err("receipt holds an entry of the wrong kind; refused".into());
    }
    // The names a kept `.moai/` held go only with the earlier-root move, and only kept names.
    if !receipt.kept_before.is_empty()
        && (!receipt.legacy_move || go_folder || receipt.kept_before.iter().any(|n| !kept_name(n)))
    {
        return Err("receipt names kept folder entries it cannot have; refused".into());
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
    if receipt
        .restored
        .iter()
        .any(|r| !restored_path_allowed(&r.path, moving))
    {
        return Err(
            "receipt holds a restored path outside the folders a restore may write; refused".into(),
        );
    }
    let copies = format!("{MOAI_DIR}/backups/{backup}/root-files/");
    if receipt
        .copied
        .iter()
        .any(|c| !go_folder || !plain_relative(&c.path) || !c.path.starts_with(&copies))
    {
        return Err(format!("receipt holds a copy outside {copies}; refused"));
    }
    // A rewrite is of the local settings alone, with its original among the copies.
    if receipt.rewritten.as_ref().is_some_and(|r| {
        let original = format!("{copies}{SETTINGS_LOCAL}");
        r.path != SETTINGS_LOCAL
            || !receipt
                .copied
                .iter()
                .any(|c| c.path == original && c.sha256 == r.original_sha256)
    }) {
        return Err("receipt names a rewrite it cannot have; refused".into());
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

/// Where a committed apply put one root instruction file (#2936 R10): the project-relative path
/// of the copy in the backup, the apply's UTC stamp and the sha256 the receipt recorded.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct BackedUpRootFile {
    pub backup: String,
    pub rel: String,
    pub sha256: String,
}

/// The newest committed apply that moved the root instruction file `name` into its backup's
/// `root-files/` and has not been reverted. Reads only; `Ok(None)` when no apply did (also
/// when an apply's receipt cannot be read: it names nothing this function may trust). A link
/// on the way to the backups is an error.
pub fn backed_up_root_file(project: &Path, name: &str) -> Result<Option<BackedUpRootFile>> {
    if !ROOT_FILES.contains(&name) {
        return Err(format!("{name}: not a root instruction file"));
    }
    newest_backed_up(project, |m| m.kind == "root_file" && m.path == name)
}

/// Where a committed apply put the shared Claude settings (#2936 R8): the copy in the backup's
/// `go-state/`, as [`backed_up_root_file`] reports a root instruction file. The file moves when
/// the manifest proves it the Go tool's (`policy`, S-CYCLE), with the Go folder or without it,
/// so `Ok(None)` is the ordinary answer for a project whose settings stayed. Reads only.
pub fn backed_up_settings(project: &Path) -> Result<Option<BackedUpRootFile>> {
    newest_backed_up(project, |m| {
        m.path == SETTINGS && (m.kind.is_empty() || m.kind == "edited")
    })
}

/// The newest committed, not reverted apply whose receipt has a moved entry `wanted` accepts.
fn newest_backed_up(
    project: &Path,
    wanted: impl Fn(&Moved) -> bool,
) -> Result<Option<BackedUpRootFile>> {
    let project =
        fs::canonicalize(project).map_err(|_| crate::contract::PROJECT_PATH_UNAVAILABLE)?;
    // An apply that moved the Go folder keeps its backup area under the new `.moai/`; one that
    // left the folder in place (S-CYCLE: the proven settings still move) under the earlier
    // root. Both are read.
    let mut stamps: Vec<String> = Vec::new();
    for base in [MOAI_DIR, LEGACY_ROOT] {
        let backups = format!("{base}/backups");
        refuse_linked(&project, &backups)?;
        if let Ok(entries) = fs::read_dir(project.join(&backups)) {
            stamps.extend(
                entries
                    .filter_map(|e| e.ok())
                    .filter_map(|e| e.file_name().into_string().ok())
                    .filter(|n| valid_stamp(n)),
            );
        }
    }
    // The stamp is a fixed-width UTC time, so the greatest name is the newest apply.
    stamps.sort_unstable_by(|a, b| b.cmp(a));
    stamps.dedup();
    for stamp in stamps {
        let Ok((receipt, _, base)) = read_receipt(&project, &stamp) else {
            continue;
        };
        if receipt.state != "committed" {
            continue;
        }
        if let Some(m) = receipt.moves.iter().find(|m| m.moved && wanted(m)) {
            return Ok(Some(BackedUpRootFile {
                rel: backup_rel_of(&format!("{base}/backups/{stamp}"), m),
                backup: stamp,
                sha256: m.sha256.clone(),
            }));
        }
    }
    Ok(None)
}

/// Read-only: what a revert of one apply would do.
pub fn revert_plan(project: &Path, backup: &str) -> Result<RevertPlan> {
    let project =
        fs::canonicalize(project).map_err(|_| crate::contract::PROJECT_PATH_UNAVAILABLE)?;
    let (receipt, _, base) = read_receipt(&project, backup)?;
    let moving = !receipt.go_backup.is_empty();
    // Either #2696 layout: the marker and this generation's `.moai/` are undone the same way.
    let new_layout = moving || receipt.legacy_move;
    // This generation's `.moai/` is still in place (not yet set aside by a revert).
    let at_moai = base == MOAI_DIR;
    let area = format!("{base}/backups/{backup}");
    let mut items = Vec::new();
    if new_layout {
        // The marker files actually present, one item each (#2871 marker rename, decision
        // 6886824285 item 2: readers accept both names, writers write the new one). The name
        // this build writes always has an item; the earlier name has one only when the file is
        // there (an apply of an earlier build wrote it). A file whose bytes are not the ones the
        // apply recorded (`receipt.marker`) is a `conflict`, so the revert stops before any
        // removal and never deletes what the apply did not write. A file of the name this build
        // writes beside an empty record is the apply that stopped between publishing and
        // recording the marker: it is the apply's own and goes as before. The earlier name
        // beside an empty record is not known to be this apply's, so it is a `conflict`.
        for name in [
            crate::state_root::LEGACY_MARKER,
            crate::state_root::RUST_MARKER,
        ] {
            let marker = format!("{MOAI_DIR}/{name}");
            let present = fs::symlink_metadata(project.join(&marker));
            let written_by_this_build = name == crate::state_root::RUST_MARKER;
            if !written_by_this_build && present.is_err() {
                continue;
            }
            let action = if !at_moai {
                "skip"
            } else {
                match present {
                    Err(_) => "skip",
                    Ok(m) if m.is_file() => {
                        let recorded = receipt.marker.as_str();
                        if written_by_this_build && recorded.is_empty() {
                            "remove"
                        } else {
                            match sha256_file(&project.join(&marker)) {
                                Ok(h) if h == recorded => "remove",
                                _ => "conflict",
                            }
                        }
                    }
                    Ok(_) => "conflict",
                }
            };
            items.push(RevertItem {
                path: marker,
                action: action.into(),
            });
        }
    }
    if let Some(r) = &receipt.rewritten {
        items.push(RevertItem {
            path: r.path.clone(),
            action: rewritten_action(&project, &area, r).into(),
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
            "state" | "aside" => {
                let entry = m
                    .path
                    .strip_prefix(&format!("{LEGACY_ROOT}/"))
                    .unwrap_or_default();
                // A state entry stands beside the kept names; a picked one is in the backup
                // area of the apply.
                let src = if m.kind == "aside" {
                    format!("{area}/{SET_ASIDE}/{entry}")
                } else {
                    format!("{base}/{entry}")
                };
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
                let here = fs::symlink_metadata(project.join(&m.path));
                let linked = linked_component(&project, &m.path);
                // AUDIT #3077 B2: an install step may have written a new `.claude/settings.json`
                // after the move. A regular file there (no link on the way, not a folder) is
                // moved aside into the backup area and the earlier one comes back; anything
                // else at the path, and every other path, is still a `conflict`.
                let replaceable =
                    m.path == SETTINGS && !linked && here.as_ref().is_ok_and(|h| h.is_file());
                if !m.moved {
                    "skip"
                } else if (here.is_ok() || linked) && !replaceable {
                    "conflict"
                } else {
                    let dest_rel = backup_rel_of(&area, m);
                    let dest = project.join(&dest_rel);
                    let ready = match m.kind.as_str() {
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
                    };
                    if replaceable && ready == "restore" {
                        "set_aside_then_restore"
                    } else {
                        ready
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
    // Not when the apply kept a `.moai/` that held kept names (D-7): that folder was never made
    // by the apply, so the revert leaves every name it held and sets nothing aside.
    if receipt.legacy_move && receipt.kept_before.is_empty() {
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
        // The project's tracked SPECs go aside with `.moai/` and come straight back to
        // `.moai/specs` (SDD B1, RB1 N-e): the plan says so.
        let whole = format!("{LEGACY_ROOT}/{}", crate::state_root::SPECS_DIR);
        let specs_moved_whole = receipt
            .moves
            .iter()
            .any(|m| m.kind == "state" && m.path == whole);
        if action == "set_aside" && !specs_moved_whole && real_dir(&project, &specs_rel()) {
            items.push(RevertItem {
                path: specs_rel(),
                action: KEEP_SPECS.into(),
            });
        }
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

/// The backup copy of a rewritten file, in the backup area `area`.
fn original_copy(area: &str, r: &Rewritten) -> String {
    format!("{area}/root-files/{}", r.path)
}

/// What a revert does with the file an apply rewrote.
fn rewritten_action(project: &Path, area: &str, r: &Rewritten) -> &'static str {
    let is = |rel: &str, sha: &str| sha256_file(&project.join(rel)).is_ok_and(|h| h == sha);
    if linked_component(project, &r.path) {
        return "conflict";
    }
    let here = fs::symlink_metadata(project.join(&r.path));
    if here.as_ref().is_ok_and(|m| !m.is_file()) {
        return "conflict";
    }
    // Already the original: never rewritten, or put back by the apply or an earlier revert.
    if here.is_ok() && is(&r.path, &r.original_sha256) {
        return "skip";
    }
    let copy = original_copy(area, r);
    if linked_component(project, &copy) || !is(&copy, &r.original_sha256) {
        return "missing";
    }
    if here.is_ok() {
        "set_aside_then_restore_original"
    } else {
        "restore_original"
    }
}

/// Put the original bytes of a rewritten file back: a file at the path moves into
/// `replaced-by-revert/<UTC>/` of the backup area first, and nothing is overwritten.
fn restore_rewritten(project: &Path, area: &str, r: &Rewritten, aside: bool) -> Result<()> {
    let copy = original_copy(area, r);
    refuse_linked(project, &copy)?;
    let bytes = fs::read(project.join(&copy)).map_err(|_| "backup copy read failed")?;
    if digest(&bytes) != r.original_sha256 {
        return Err(format!("{copy} changed after the revert plan; stopped"));
    }
    refuse_linked(project, &r.path)?;
    if aside {
        if !fs::symlink_metadata(project.join(&r.path)).is_ok_and(|m| m.is_file()) {
            return Err(format!(
                "{}: changed after the revert plan; nothing moved aside",
                r.path
            ));
        }
        let to = format!("{area}/{REPLACED_BY_REVERT}/{}/{}", stamp_now(), r.path);
        make_private_dirs(project, &parent_of(&to)?)?;
        move_exclusive(&project.join(&r.path), &project.join(&to))
            .map_err(|e| format!("{}: {e}", r.path))?;
    }
    make_dirs(project, &parent_of(&r.path)?)?;
    crate::apply::publish_exclusive_mode(&project.join(&r.path), &bytes, Some(r.mode & 0o777))
        .map_err(|e| format!("{}: {e}", r.path))
}

/// Undo one apply, all or nothing up front: any `conflict` or `missing` item refuses the whole
/// revert before the first write. Never overwrites, never deletes what the apply did not write.
/// A regular `.claude/settings.json` that is there again is not a conflict: it is moved aside
/// first, into `replaced-by-revert/<UTC>/` of the backup area (`set_aside_then_restore`).
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
        return Err(STALE_REVERT_PLAN.into());
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
    // Before the first write: prove the receipt can be saved by saving the unchanged receipt
    // through the same replace every later step uses. A size, permission or space problem
    // refuses here, with nothing moved or removed; a stop after this point is recorded in the
    // receipt and the revert is run again (an already removed copy reads as `skip`).
    saved = save_receipt(&project, &base, &receipt, Some(&saved))?;
    // AUDIT #3077 B2: a settings file written after the move is set aside first, before any
    // other write, into `<area>/replaced-by-revert/<UTC>/` (never overwritten, never deleted);
    // a stop later on reads the item as a plain `restore` on the next plan.
    for item in fresh
        .items
        .iter()
        .filter(|i| i.action == "set_aside_then_restore")
    {
        if item.path != SETTINGS {
            return Err(format!(
                "{}: only the shared settings are set aside; stopped",
                item.path
            ));
        }
        refuse_linked(&project, &item.path)?;
        if !fs::symlink_metadata(project.join(&item.path)).is_ok_and(|m| m.is_file()) {
            return Err(format!(
                "{}: changed after the revert plan; nothing moved aside",
                item.path
            ));
        }
        let to = format!(
            "{base}/backups/{backup}/{REPLACED_BY_REVERT}/{}/{}",
            stamp_now(),
            item.path
        );
        make_private_dirs(&project, &parent_of(&to)?)?;
        move_exclusive(&project.join(&item.path), &project.join(&to))
            .map_err(|e| format!("{}: {e}", item.path))?;
    }
    // The rewritten local settings come back before the folder moves again. A stop here
    // leaves the file at its path or in `replaced-by-revert/`, and the next plan reads it.
    if let Some(r) = receipt.rewritten.clone() {
        let action = fresh
            .items
            .iter()
            .find(|i| i.path == r.path)
            .map(|i| i.action.as_str());
        if let Some(action @ ("restore_original" | "set_aside_then_restore_original")) = action {
            restore_rewritten(
                &project,
                &format!("{base}/backups/{backup}"),
                &r,
                action == "set_aside_then_restore_original",
            )?;
        }
    }
    // The recorded marker digest stays until every marker file is gone: with both names present,
    // clearing it after the first removal would leave a stopped revert comparing the other file
    // against an empty record (`conflict`) on the next run (Codex review of #3047).
    let mut marker_removed = false;
    for item in fresh.items.iter().filter(|i| i.action == "remove") {
        refuse_linked(&project, &item.path)?;
        fs::remove_file(project.join(&item.path)).map_err(|_| "marker not removed")?;
        marker_removed = true;
    }
    if marker_removed {
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
        .filter(|i| i.action == "restore" || i.action == "set_aside_then_restore")
        .map(|i| (i.path.clone(), entry_of(&i.path)))
        .collect();
    // Entries and files first, the folder last.
    for (path, entry) in restores.iter().filter(|(_, m)| m.kind != "folder") {
        let kind = &entry.kind;
        if kind == "state" || kind == "aside" {
            let entry = path
                .strip_prefix(&format!("{LEGACY_ROOT}/"))
                .ok_or("state entry outside the earlier root")?;
            // A merged SPEC goes back into `.moai-cli/specs/` (RB1).
            make_dirs(&project, &parent_of(path)?)?;
            let src = if kind == "aside" {
                format!("{base}/backups/{backup}/{SET_ASIDE}/{entry}")
            } else {
                format!("{base}/{entry}")
            };
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
        // The project's tracked SPECs stay where git has them (SDD B1): only they come back,
        // as the plan's `restored_to_moai_specs` item says.
        let specs = project
            .join(&receipt.rust_aside)
            .join(crate::state_root::SPECS_DIR);
        let planned = fresh
            .items
            .iter()
            .any(|i| i.path == specs_rel() && i.action == KEEP_SPECS);
        if planned
            && fs::symlink_metadata(&specs).is_ok_and(|m| m.is_dir())
            && fs::symlink_metadata(project.join(MOAI_DIR)).is_err()
        {
            fs::create_dir(project.join(MOAI_DIR))
                .map_err(|_| format!("{MOAI_DIR}: folder create failed"))?;
            move_folder(
                &specs,
                &project.join(MOAI_DIR).join(crate::state_root::SPECS_DIR),
            )?;
        }
    }
    receipt.state = "reverted".into();
    save_receipt(&project, &base, &receipt, Some(&saved))?;
    Ok(fresh)
}

#[cfg(test)]
mod tests {
    use super::*;

    fn file_item(path: &str, content: &str) -> Item {
        let mut i = new_item(
            path,
            "move",
            "proven Go deploy, unchanged".into(),
            "tool_asset",
        );
        i.proof = "manifest; sha256 matches the deploy".into();
        i.deployed_hash = "sha256:aa".into();
        i.content = content.into();
        i.kind = "rule".into();
        i.git = "clean".into();
        i
    }

    fn restore(from: &str, action: &str, sha: &str) -> Restore {
        Restore {
            from: from.into(),
            to: from.replace(".moai-backup", ".moai"),
            sha256: sha.into(),
            action: action.into(),
            reason: String::new(),
        }
    }

    fn root_file(path: &str, action: &str, sha: &str) -> RootFile {
        RootFile {
            path: path.into(),
            sha256: sha.into(),
            action: action.into(),
            backup_copy: format!(".moai/backups/<UTC>/root-files/{path}"),
            user_choice: "keep".into(),
            reason: "why".into(),
            class: "instructions".into(),
            go_proof: "none".into(),
            git: String::new(),
        }
    }

    fn state_move(from: &str, to: &str, action: &str) -> StateMove {
        StateMove {
            from: from.into(),
            to: to.into(),
            action: action.into(),
            reason: "state".into(),
            tracked: None,
        }
    }

    /// One row of every shape the apply acts on and of every shape it ignores.
    fn sample() -> GoStatePlan {
        GoStatePlan {
            schema_version: SCHEMA_VERSION,
            project: "/p".into(),
            manifest_version: Some("v3".into()),
            items: vec![
                new_item(MOAI_DIR, "move", "the folder".into(), "go_folder"),
                file_item(".claude/a.md", "sha256:01"),
                file_item(".claude/b.md", "sha256:02"),
                new_item(".claude/c.md", "leave", "referenced".into(), "tool_asset"),
                new_item(".moai/x", "with_folder", "with it".into(), "go_folder"),
            ],
            unrecorded: BTreeMap::from([(".claude".to_string(), 1)]),
            restores: vec![
                restore(".moai-backup/specs/a.md", "copy", "sha256:03"),
                restore(".moai-backup/specs/c.md", "conflict", "sha256:04"),
                restore(".moai-backup/specs/s.md", "excluded_secret", ""),
                restore(".moai-backup/specs", "leave_link", ""),
            ],
            go_backup: Some(".moai-backup".into()),
            layout: "folder_moves",
            backup_to: Some(".moai/backups/<UTC>".into()),
            state_moves: vec![
                state_move("earlier/index", ".moai/index", "move"),
                state_move("earlier/notes", "", "leave"),
            ],
            root_files: vec![
                root_file("CLAUDE.md", "move", "sha256:05"),
                root_file(".claude/settings.local.json", "copy", "sha256:06"),
                root_file("AGENTS.md", "leave", ""),
            ],
            left_in_backup: vec!["config".into()],
            kinds: BTreeMap::from([("rule".to_string(), 2)]),
            counterparts: Vec::new(),
            after_move: AfterMove::default(),
            settings_move: None,
            local_references: None,
            backup_ignore: None,
            ignored_tracked: None,
            secrets: 0,
            warnings: vec!["w".into()],
            plan_digest: String::new(),
        }
    }

    fn moved_entry(path: &str, kind: &str) -> Moved {
        Moved {
            path: path.into(),
            kind: kind.into(),
            ..Moved::default()
        }
    }

    fn digest_of(plan: &GoStatePlan) -> String {
        bound_digest(plan).unwrap()
    }

    type Change = fn(&mut GoStatePlan);

    /// F7: the preview shows more than the apply acts on; none of that data changes the digest.
    #[test]
    fn preview_only_data_does_not_change_the_digest() {
        let base = digest_of(&sample());
        let same: Vec<(&str, Change)> = vec![
            ("schema_version", |p| p.schema_version += 1),
            ("manifest_version", |p| p.manifest_version = None),
            ("unrecorded", |p| {
                p.unrecorded.insert(".moai/logs".into(), 9);
            }),
            ("layout", |p| p.layout = "folder_stays"),
            ("backup_to", |p| p.backup_to = None),
            ("left_in_backup", |p| p.left_in_backup.push("logs".into())),
            ("kinds", |p| {
                p.kinds.insert("hook".into(), 1);
            }),
            ("counterparts", |p| {
                p.counterparts.push(Counterpart {
                    kind: "rule".into(),
                    moved: 2,
                    counterpart: "c".into(),
                    note: "n".into(),
                })
            }),
            ("after_move", |p| {
                p.after_move.unknown = true;
                p.after_move.empty_folders.push(".claude".into());
            }),
            ("settings_move", |p| {
                p.settings_move = Some(SettingsMove {
                    path: SETTINGS.into(),
                    proof: "manifest".into(),
                    edited: true,
                    layout: "folder_stays".into(),
                    freed: vec![".claude/a.md".into()],
                    folder_freed: true,
                    still_holding: vec![SETTINGS_LOCAL.into()],
                })
            }),
            ("backup_ignore", |p| {
                p.backup_ignore = Some(BackupIgnore {
                    state: "not_ignored",
                    file: crate::project_state!("backups/.gitignore").into(),
                    examples: vec![
                        crate::project_state!("backups/<UTC>/go-state-receipt.json").into()
                    ],
                    reason: "r".into(),
                })
            }),
            ("item freed_by", |p| p.items[1].freed_by = SETTINGS.into()),
            ("secrets", |p| p.secrets = 7),
            ("warnings", |p| p.warnings.push("another".into())),
            ("plan_digest", |p| p.plan_digest = "sha256:ff".into()),
            ("item reason", |p| p.items[1].reason = "other".into()),
            ("item proof", |p| p.items[1].proof = "other".into()),
            ("item git", |p| p.items[1].git = "modified".into()),
            ("item kind", |p| p.items[1].kind = "hook".into()),
            ("item that stays", |p| {
                p.items[3].class = "user_modified".into();
                p.items[3].content = "sha256:99".into();
                p.items[3].deployed_hash = "sha256:98".into();
            }),
            ("new item that stays", |p| {
                p.items
                    .push(new_item(".claude/new.md", "leave", "r".into(), "unknown"))
            }),
            ("item with the folder", |p| {
                p.items[4].reason = "other".into()
            }),
            ("restore reason", |p| p.restores[0].reason = "other".into()),
            ("excluded restore", |p| {
                p.restores[2].sha256 = "sha256:77".into();
                p.restores[2].reason = "other".into();
            }),
            ("new excluded restore", |p| {
                p.restores
                    .push(restore(".moai-backup/docs/.env", "excluded_secret", ""))
            }),
            ("new linked restore", |p| {
                p.restores
                    .push(restore(".moai-backup/docs", "leave_link", ""))
            }),
            ("state reason", |p| p.state_moves[0].reason = "other".into()),
            ("state that stays", |p| {
                p.state_moves[1].reason = "other".into();
                p.state_moves.push(state_move("earlier/other", "", "leave"));
            }),
            ("root backup_copy", |p| {
                p.root_files[0].backup_copy = "x".into()
            }),
            ("root user_choice", |p| {
                p.root_files[0].user_choice = "x".into()
            }),
            ("root reason", |p| p.root_files[0].reason = "x".into()),
            ("root class", |p| p.root_files[0].class = "conflict".into()),
            ("root go_proof", |p| {
                p.root_files[0].go_proof = "marker".into()
            }),
            ("root git", |p| p.root_files[0].git = "modified".into()),
            ("root that stays", |p| {
                p.root_files[2].sha256 = "sha256:88".into();
                p.root_files.push(root_file("AGENTS.local.md", "leave", ""));
            }),
        ];
        for (what, change) in same {
            let mut p = sample();
            change(&mut p);
            assert_eq!(digest_of(&p), base, "{what} must not change the digest");
        }
    }

    /// Everything the apply moves, copies or writes changes the digest.
    #[test]
    fn bound_data_changes_the_digest() {
        let base = digest_of(&sample());
        let differ: Vec<(&str, Change)> = vec![
            ("project", |p| p.project = "/q".into()),
            ("go_backup name", |p| {
                p.go_backup = Some(".moai-backup-<UTC>".into())
            }),
            ("go_backup gone", |p| p.go_backup = None),
            ("item path", |p| p.items[1].path = ".claude/z.md".into()),
            ("item action", |p| p.items[1].action = "leave".into()),
            ("item class", |p| p.items[1].class = "user_modified".into()),
            ("item deployed_hash", |p| {
                p.items[1].deployed_hash = "sha256:bb".into()
            }),
            ("item content", |p| p.items[1].content = "sha256:ee".into()),
            ("new move", |p| {
                p.items.push(file_item(".claude/n.md", "sha256:0a"))
            }),
            ("moving folder stays", |p| {
                p.items[0].action = "leave".into()
            }),
            ("restore copy sha256", |p| {
                p.restores[0].sha256 = "sha256:ee".into()
            }),
            ("restore copy to", |p| {
                p.restores[0].to = ".moai/other".into()
            }),
            ("restore copy action", |p| {
                p.restores[0].action = "conflict".into()
            }),
            ("restore conflict sha256", |p| {
                p.restores[1].sha256 = "sha256:ee".into()
            }),
            ("new restore copy", |p| {
                p.restores
                    .push(restore(".moai-backup/specs/n.md", "copy", "sha256:0b"))
            }),
            ("new restore conflict", |p| {
                p.restores.push(restore(
                    ".moai-backup/specs/n.md",
                    "conflict_link",
                    "sha256:0b",
                ))
            }),
            ("restore copy gone", |p| {
                p.restores.remove(0);
            }),
            ("state to", |p| p.state_moves[0].to = ".moai/other".into()),
            ("new state move", |p| {
                p.state_moves
                    .push(state_move("earlier/decisions", ".moai/decisions", "move"))
            }),
            ("state move now stays", |p| {
                p.state_moves[0].action = "leave".into()
            }),
            ("root sha256", |p| {
                p.root_files[0].sha256 = "sha256:ee".into()
            }),
            ("root action", |p| p.root_files[0].action = "copy".into()),
            ("root copy sha256", |p| {
                p.root_files[1].sha256 = "sha256:ee".into()
            }),
            ("new root move", |p| {
                p.root_files
                    .push(root_file("AGENTS.local.md", "move", "sha256:0c"))
            }),
            ("root now stays", |p| {
                p.root_files[0].action = "leave".into()
            }),
        ];
        for (what, change) in differ {
            let mut p = sample();
            change(&mut p);
            assert_ne!(digest_of(&p), base, "{what} must change the digest");
        }
    }

    /// The digest does not depend on the order the preview lists its rows in.
    #[test]
    fn row_order_does_not_change_the_digest() {
        let base = digest_of(&sample());
        let mut p = sample();
        p.items.reverse();
        p.restores.reverse();
        p.state_moves.reverse();
        p.root_files.reverse();
        assert_eq!(digest_of(&p), base);
    }

    /// A moving path other than the two folders must carry the hash of what it moves (fail
    /// closed); the folders, a leaving path and a path under the folder need none.
    #[test]
    fn a_move_without_content_is_refused_except_the_two_folders() {
        let legacy = new_item(LEGACY_ROOT, "move", "r".into(), "legacy_state");
        let mut p = sample();
        p.items.push(legacy);
        assert!(bound_digest(&p).is_ok(), "{LEGACY_ROOT}");
        assert!(bound_digest(&sample()).is_ok(), "{MOAI_DIR}");

        let mut p = sample();
        p.items.push(new_item(
            ".codex/hooks.json",
            "move",
            "r".into(),
            "go_hooks",
        ));
        let err = bound_digest(&p).unwrap_err();
        assert!(err.contains(".codex/hooks.json"), "{err}");
        assert_eq!(
            crate::contract::classify(&err),
            crate::contract::Kind::InvalidInput
        );
        let mut p = sample();
        p.items[1].content.clear();
        assert!(bound_digest(&p).is_err(), "a file move with an empty hash");
        // Not a move: nothing to bind.
        let mut p = sample();
        p.items.push(new_item(
            ".codex/hooks.json",
            "leave",
            "r".into(),
            "go_hooks",
        ));
        assert!(bound_digest(&p).is_ok());
    }

    /// The hash order is part of the contract: a golden text and digest for a small plan. A
    /// change here needs a `BOUND_VERSION` bump.
    #[test]
    fn the_bound_layout_is_pinned() {
        assert_eq!(BOUND_VERSION, 2);
        let mut p = sample();
        p.items.truncate(2);
        p.restores.truncate(1);
        p.state_moves.truncate(1);
        p.root_files.truncate(1);
        let text = serde_json::to_string(&bound_of(&p).unwrap()).unwrap();
        assert_eq!(
            text,
            concat!(
                r#"{"bound_version":2,"project":"/p","go_backup":".moai-backup","#,
                r#""moves":[{"path":".claude/a.md","action":"move","class":"tool_asset","#,
                r#""deployed_hash":"sha256:aa","content":"sha256:01"},"#,
                r#"{"path":".moai","action":"move","class":"go_folder","#,
                r#""deployed_hash":"","content":""}],"#,
                r#""restores":[{"from":".moai-backup/specs/a.md","to":".moai/specs/a.md","#,
                r#""action":"copy","sha256":"sha256:03"}],"#,
                r#""state_moves":[{"from":"earlier/index","to":".moai/index","action":"move"}],"#,
                r#""root_files":[{"path":"CLAUDE.md","action":"move","sha256":"sha256:05"}]}"#
            )
        );
        assert_eq!(
            digest_of(&p),
            "sha256:8927c17cb1ed999ad2ab58405a88b8e4b9084671a5001ec90a6224d622b6d4c9"
        );
        // The empty plan, the smallest `Bound`.
        let mut empty = sample();
        empty.items.clear();
        empty.restores.clear();
        empty.state_moves.clear();
        empty.root_files.clear();
        empty.go_backup = None;
        assert_eq!(
            serde_json::to_string(&bound_of(&empty).unwrap()).unwrap(),
            r#"{"bound_version":2,"project":"/p","go_backup":null,"moves":[],"restores":[],"state_moves":[],"root_files":[]}"#
        );
    }

    /// AUDIT #3077 B3: a move never overwrites. The proven settings now travel this path, so
    /// the pin: a destination that exists (a file, or a hard link to the source itself) is
    /// refused with its own message and both names keep their bytes. A plain `fs::rename` would
    /// replace the file, and would report success for the hard link.
    #[test]
    #[cfg(unix)]
    fn move_exclusive_never_overwrites_an_existing_destination() {
        use std::os::unix::fs::MetadataExt;
        let tmp = tempfile::TempDir::new().unwrap();
        let (from, to) = (tmp.path().join("from.json"), tmp.path().join("to.json"));
        fs::write(&from, b"source bytes").unwrap();
        fs::write(&to, b"destination bytes").unwrap();
        let e = move_exclusive(&from, &to).unwrap_err();
        assert!(e.contains("the destination already exists"), "{e}");
        assert_eq!(fs::read(&from).unwrap(), b"source bytes");
        assert_eq!(fs::read(&to).unwrap(), b"destination bytes");

        // The destination is a second name of the source.
        let alias = tmp.path().join("alias.json");
        fs::hard_link(&from, &alias).unwrap();
        let e = move_exclusive(&from, &alias).unwrap_err();
        assert!(e.contains("the destination already exists"), "{e}");
        assert_eq!(fs::read(&from).unwrap(), b"source bytes");
        assert_eq!(fs::read(&alias).unwrap(), b"source bytes");
        assert_eq!(fs::metadata(&from).unwrap().nlink(), 2, "both names remain");

        // A free destination takes the file, and the source name is gone.
        fs::remove_file(&alias).unwrap();
        let free = tmp.path().join("free.json");
        move_exclusive(&from, &free).unwrap();
        assert!(!from.exists());
        assert_eq!(fs::read(&free).unwrap(), b"source bytes");
    }

    /// F17: both refusals map to `stale_plan` and name the command to run next.
    #[test]
    fn the_stale_messages_have_the_stale_plan_kind_and_name_their_next_step() {
        use crate::contract::{classify, Kind};
        for (m, command) in [
            (STALE_PLAN, "go-state-plan"),
            (STALE_REVERT_PLAN, "go-state-revert-plan"),
        ] {
            assert_eq!(classify(m), Kind::StalePlan, "{m}");
            assert!(crate::error_next::names_next_step(m), "{m}");
            assert!(m.contains(&format!("`moai {command} ")), "{m}");
            assert!(m.contains("nothing was written; run `"), "{m}");
        }
    }

    /// S-CYCLE: the settings are the last file to move, whatever order the plan listed them in.
    #[test]
    fn the_settings_move_last() {
        let mut moves = vec![
            moved_entry(".claude/a.md", ""),
            moved_entry(SETTINGS, "edited"),
            moved_entry(".claude/skills/z.md", ""),
        ];
        settings_last(&mut moves);
        let paths: Vec<&str> = moves.iter().map(|m| m.path.as_str()).collect();
        assert_eq!(paths, [".claude/a.md", ".claude/skills/z.md", SETTINGS]);
        // Without settings the order is untouched.
        let mut other = vec![moved_entry("b", ""), moved_entry("a", "")];
        settings_last(&mut other);
        assert_eq!(other[0].path, "b");
    }

    /// S-CYCLE: a staying-folder apply carries plain files, and an edited file only when it is
    /// the proven settings; every other kind, and an edited file anywhere else, is refused.
    #[test]
    fn the_staying_layout_carries_edited_settings_and_nothing_else_edited() {
        assert!(stay_kinds_ok(&[]));
        assert!(stay_kinds_ok(&[moved_entry(".claude/a.md", "")]));
        assert!(stay_kinds_ok(&[moved_entry(SETTINGS, "edited")]));
        assert!(stay_kinds_ok(&[moved_entry(SETTINGS, "")]));
        for kind in [
            "edited",
            "link",
            "mirror_copy",
            "root_file",
            "state",
            "folder",
        ] {
            assert!(
                !stay_kinds_ok(&[moved_entry(".claude/rules/x.md", kind)]),
                "{kind}"
            );
        }
        for kind in ["link", "mirror_copy", "root_file", "state", "folder"] {
            assert!(!stay_kinds_ok(&[moved_entry(SETTINGS, kind)]), "{kind}");
        }
    }

    fn entry(provenance: &str, deployed: &str) -> Entry {
        Entry {
            provenance: provenance.into(),
            deployed_hash: deployed.into(),
            parts: serde_json::Value::Null,
        }
    }

    /// S-CYCLE: what proves the settings, one condition at a time, in both directions.
    #[test]
    fn settings_the_rust_receipt_owns_are_never_proven_again() {
        let tmp = tempfile::TempDir::new().unwrap();
        let project = fs::canonicalize(tmp.path()).unwrap();
        fs::create_dir_all(project.join(".claude")).unwrap();
        let body = b"{\"a\":1}";
        fs::write(project.join(SETTINGS), body).unwrap();
        let e = entry("template_managed", &digest(body));
        let receipt = project.join(crate::state_root::rel(
            &project,
            "manifests/claude-settings.json",
        ));
        fs::create_dir_all(receipt.parent().unwrap()).unwrap();
        for (text, want) in [
            // No owned key: the file is still the Go tool's.
            (r#"{"schema_version":1,"owned":{}}"#, None),
            (
                r#"{"schema_version":1,"owned":{"statusLine":{"type":"command"}}}"#,
                Some("rust"),
            ),
            // An interrupted write that was about to own a key counts as owned.
            (
                r#"{"schema_version":1,"owned":{},"pending":{"owned":{"hooks":{}}}}"#,
                Some("rust"),
            ),
            // A receipt that cannot be read leaves the question open: not proven.
            ("not json", Some("unknown")),
            (r#"{"schema_version":2,"owned":{}}"#, Some("unknown")),
        ] {
            fs::write(&receipt, text).unwrap();
            let got = settings_proof(&project, &e).err().map(|(_, class)| class);
            assert_eq!(got, want, "{text}");
        }
    }

    /// S-AGENTS-STAY U1: the `git check-ignore -v -n` lines as git 2.54 prints them, in both
    /// directions: a match is ignored, a miss and a `!` rule are not, and anything else is not
    /// guessed.
    #[test]
    fn check_ignore_lines_are_read_in_both_directions() {
        for (line, want) in [
            (
                "state/backups/.gitignore:1:*\tstate/backups/x/a.json",
                Some(true),
            ),
            (".gitignore:3:.moai-cli/backups/\t\"a b\"", Some(true)),
            // The source may hold a colon; the line number is the first `:<digits>:`.
            ("/odd:dir/ignore:12:go-state/\tp", Some(true)),
            ("::\t.moai-cli/backups/x/a.json", Some(false)),
            // A negated pattern is the rule that un-ignores the path.
            ("state/backups/.gitignore:2:!keep-me\tp", Some(false)),
            // Not in the shape: no tab, no line number, an empty pattern.
            ("", None),
            ("::", None),
            ("src:x:*\tp", None),
            ("src:1:\tp", None),
            ("warning: unable to access\tp", None),
        ] {
            assert_eq!(check_ignore_line(line), want, "{line:?}");
        }
    }

    #[test]
    fn the_settings_are_proven_only_by_the_manifest_and_a_plain_file() {
        let tmp = tempfile::TempDir::new().unwrap();
        let project = fs::canonicalize(tmp.path()).unwrap();
        fs::create_dir_all(project.join(".claude")).unwrap();
        let body = b"{\"a\":1}";
        fs::write(project.join(SETTINGS), body).unwrap();
        let hash = digest(body);

        let p = settings_proof(&project, &entry("template_managed", &hash)).unwrap();
        assert!(!p.edited);
        assert_eq!(p.proof, "manifest; sha256 matches the deploy");
        assert_eq!(p.content, hash);
        let p = settings_proof(&project, &entry("template_managed", "sha256:00")).unwrap();
        assert!(p.edited);
        assert_eq!(
            p.proof,
            "manifest records this path; content differs from what the Go tool deployed"
        );
        let p = settings_proof(&project, &entry("user_modified", &hash)).unwrap();
        assert!(p.edited);
        assert_eq!(p.proof, "manifest; recorded user_modified");

        for (provenance, why) in [
            ("user_created", "user_created: the user's own file"),
            ("generated_managed", "not template_managed"),
            ("", "not template_managed"),
        ] {
            let e = settings_proof(&project, &entry(provenance, &hash)).err();
            assert_eq!(e.map(|(r, _)| r), Some(why), "{provenance}");
        }
        // A folder, a link, a missing file.
        fs::remove_file(project.join(SETTINGS)).unwrap();
        assert_eq!(
            settings_proof(&project, &entry("template_managed", &hash))
                .err()
                .map(|(r, _)| r),
            Some("missing")
        );
        fs::create_dir(project.join(SETTINGS)).unwrap();
        assert_eq!(
            settings_proof(&project, &entry("template_managed", &hash))
                .err()
                .map(|(r, _)| r),
            Some("not a regular file")
        );
        fs::remove_dir(project.join(SETTINGS)).unwrap();
        #[cfg(unix)]
        {
            fs::write(project.join("real.json"), body).unwrap();
            std::os::unix::fs::symlink(project.join("real.json"), project.join(SETTINGS)).unwrap();
            assert_eq!(
                settings_proof(&project, &entry("template_managed", &hash))
                    .err()
                    .map(|(r, _)| r),
                Some("a path component is a symlink")
            );
        }
    }

    #[test]
    fn a_file_that_changed_since_the_hash_stops_the_move() {
        let dir = tempfile::tempdir().unwrap();
        let file = dir.path().join("f.md");
        fs::write(&file, "one\n").unwrap();
        let sha = sha256_file(&file).unwrap();
        assert!(still_as_previewed(&file, "f.md", &sha).is_ok());
        fs::write(&file, "one\nx").unwrap();
        let err = still_as_previewed(&file, "f.md", &sha).unwrap_err();
        assert_eq!(err, "f.md changed after the preview; move stopped");
        // Gone, or a link in its place: never moved either.
        fs::remove_file(&file).unwrap();
        assert!(still_as_previewed(&file, "f.md", &sha).is_err());
        #[cfg(unix)]
        {
            fs::write(dir.path().join("t.md"), "one\n").unwrap();
            std::os::unix::fs::symlink(dir.path().join("t.md"), &file).unwrap();
            assert!(still_as_previewed(&file, "f.md", &sha).is_err());
        }
    }

    /// #3004 P4-2 AC-11: the hooks the Rust build generates call `moai`, and none of them is
    /// taken for the Go wiring's `moai hook <event>`; a real Go handler still is.
    #[test]
    fn the_rust_generated_hooks_are_not_the_go_wiring() {
        let go = br#"{"hooks":{"SessionStart":[{"hooks":[{"type":"command","command":"moai hook session-start"}]}]}}"#;
        assert!(hooks_proof(go).is_ok());
        let set = crate::hooks::current_hooks("claude").unwrap();
        let doc = serde_json::json!({ "hooks": set });
        assert!(doc.to_string().contains("\"moai "), "{doc}");
        let proof = hooks_proof(doc.to_string().as_bytes());
        assert!(proof.is_err(), "{proof:?}");
        for command in [
            "moai hook-dispatch --host codex --event stop",
            "moai hook-receipt --host codex",
            "moai claim-verify .",
            "moai lane-stop-hook --lane a --repo o/r",
        ] {
            let doc = serde_json::json!({"hooks": {"Stop": [{"hooks": [
                {"type": "command", "command": command}]}]}});
            assert!(
                hooks_proof(doc.to_string().as_bytes()).is_err(),
                "{command}"
            );
        }
    }

    /// #2936 R-fix (audit mutant M03): a folder that is a link to a folder is not emptied by
    /// the move, however many of the files that would move sit inside its target. The
    /// classification refuses a path with a linked component, so only this function's own
    /// check stands between a link and the list of emptied folders.
    #[cfg(unix)]
    #[test]
    fn a_linked_folder_is_not_listed_as_emptied() {
        use std::collections::BTreeSet;
        let tmp = tempfile::TempDir::new().unwrap();
        let project = tmp.path().join("project");
        fs::create_dir_all(project.join(".claude")).unwrap();
        fs::create_dir_all(project.join("store/workflows")).unwrap();
        fs::write(project.join("store/workflows/w1.md"), "# w1\n").unwrap();
        std::os::unix::fs::symlink("../store/workflows", project.join(".claude/workflows"))
            .unwrap();
        // The same shape with a real folder, to show the walk lists what it should.
        fs::create_dir_all(project.join(".claude/rules")).unwrap();
        fs::write(project.join(".claude/rules/r.md"), "# r\n").unwrap();

        let linked = BTreeSet::from([".claude/workflows/w1.md"]);
        let (empty, stopped) = empty_after(&project, &linked, 100);
        assert!(!stopped);
        assert!(empty.is_empty(), "a link was listed as emptied: {empty:?}");

        let real = BTreeSet::from([".claude/rules/r.md"]);
        let (empty, stopped) = empty_after(&project, &real, 100);
        assert!(!stopped);
        assert_eq!(
            empty,
            BTreeSet::from([".claude/rules".to_string()]),
            "`.claude` also holds the link, so it stays"
        );
    }

    /// A receipt of exactly the shared limit is written; one byte more is refused before any
    /// file or folder is created, and the message names the limit.
    #[test]
    fn a_receipt_over_the_limit_is_refused_before_any_write() {
        let receipt = |pad: usize| Receipt {
            schema_version: 1,
            project: "/p".into(),
            backup: "20261120T000000Z".into(),
            plan_digest: "sha256:x".into(),
            permission_reference: "r".repeat(pad),
            state: "pending".into(),
            moves: Vec::new(),
            restored: Vec::new(),
            go_backup: String::new(),
            copied: Vec::new(),
            marker: String::new(),
            rust_aside: String::new(),
            legacy_move: false,
            kept_before: Vec::new(),
            rewritten: None,
        };
        let fixed = serde_json::to_vec_pretty(&receipt(0)).unwrap().len();
        let pad = RECEIPT_MAX as usize - fixed;

        let at_limit = tempfile::TempDir::new().unwrap();
        let bytes = save_receipt(at_limit.path(), MOAI_DIR, &receipt(pad), None).unwrap();
        assert_eq!(bytes.len() as u64, RECEIPT_MAX);

        let over = tempfile::TempDir::new().unwrap();
        let err = save_receipt(over.path(), MOAI_DIR, &receipt(pad + 1), None).unwrap_err();
        assert_eq!(
            err,
            "receipt larger than 16 MiB; refused; nothing was written"
        );
        assert!(
            fs::read_dir(over.path()).unwrap().next().is_none(),
            "nothing was created"
        );
    }

    /// SPEC-20261126-go-state-restore-rule AC-6 (REQ-7): a restored path may lie under any
    /// folder of the new `.moai/`, not under its backups folder and not on a Go marker.
    #[test]
    fn restore_rule_allows_any_folder_but_not_backups_or_a_marker() {
        for ok in [
            ".moai/specs/SPEC-1/spec.md",
            ".moai/design/x.md",
            ".moai/decisions/user.md",
            ".moai/school/a/b/c.md",
            ".moai/config/sections/b.yaml",
        ] {
            assert!(restored_path_allowed(ok, true), "{ok}");
        }
        for refused in [
            ".moai/backups/x/y.md",
            ".moai/backups",
            ".moai/manifest.json",
            ".moai/config/config.yaml",
            ".moai/config/sections/system.yaml",
            ".moai/.update.lock",
            ".moai",
            ".moai/",
            ".claude/rules/a.md",
            "design/x.md",
        ] {
            assert!(!restored_path_allowed(refused, true), "{refused}");
        }
        // The earlier root is not a place the moving layout restores into.
        assert!(!restored_path_allowed(
            &format!("{LEGACY_ROOT}/specs/a.md"),
            true
        ));
    }

    /// SPEC-20261126-go-state-restore-rule AC-6: the earlier-root layout restores into its
    /// `specs/` only, as before.
    #[test]
    fn restore_rule_keeps_the_earlier_root_to_its_specs() {
        let specs = format!("{LEGACY_ROOT}/specs/a/spec.md");
        assert!(restored_path_allowed(&specs, false));
        assert!(!restored_path_allowed(".moai/design/x.md", false));
        assert!(!restored_path_allowed(
            &format!("{LEGACY_ROOT}/design/x.md"),
            false
        ));
    }
}
