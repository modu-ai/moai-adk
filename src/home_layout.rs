//! #2871 I2 PR-A: the HOME name table as code, and where this generation's HOME items go.
//!
//! User decision D-M1 (#2871 comment 6876633358): this generation owns `~/.moai` at its top
//! level, with no subfolder of its own. When a name is taken by both generations, this
//! generation keeps the name and the Go item retires to a dated backup with a manifest; user
//! data, secrets and items of unknown owner are kept. This module only records those rules and
//! resolves paths: nothing here reads a file's content, moves or writes anything. The moves are
//! later steps of #2871.
//!
//! The table lists every top-level name this generation writes or reads under its HOME state
//! folder (PLAN 6876533944 table 1-A, leader survey 6876590509 §1) and every name the Go
//! generation keeps there (the reserved names). A HOME uses this layout only once its marker
//! has schema 2 ([`crate::state_root::classify_home`]); every other HOME keeps the paths it had.

use std::path::{Path, PathBuf};

/// Where this generation's HOME items live inside `~/.moai`.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Layout {
    /// Directly in `~/.moai/<name>`.
    TopLevel,
    /// In one subfolder, `~/.moai/<sub>/<name>`.
    Subdir(&'static str),
}

/// The one place the layout is chosen. User decision D-M1 (#2871 comment 6876633358, 2026-11-14):
/// top level — this generation owns `~/.moai`. `Layout::Subdir` stays only so the rule can
/// change here alone.
pub const RUST_HOME_LAYOUT: Layout = Layout::TopLevel;

/// What happens to the item under a name (data only in this step; nothing moves yet).
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Rule {
    /// This generation's name; on a clash it keeps the name.
    RustName,
    /// The Go item under this name retires to a dated backup with a manifest (a later step);
    /// when this generation also uses the name, it takes the name afterwards.
    GoRetires,
    /// User data (specs, reports, notes, shared worktrees): kept in place, bytes unchanged.
    UserDataKeep,
    /// A secret: never read, opened or copied; only renamed as a whole.
    SecretRenameOnly,
    /// Owner not known: kept as it is.
    UnknownKeep,
}

/// How this generation uses a name.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum RustUse {
    /// This generation owns the name: it writes it, or reads a file the user writes there and
    /// carries it with the HOME move.
    Writes,
    /// Reads only (a legacy source).
    Reads,
    None,
}

/// What kind of item a name holds (PLAN v2 6876772276 table A, column "분류").
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Class {
    /// This generation's own item.
    Rust,
    /// A Go tool asset: backed up when Go retires.
    ToolAsset,
    /// Go settings or instructions: backed up; some values are merge candidates.
    ConfigGuide,
    /// A secret: renamed as a whole, never read.
    Secret,
    /// User data: kept in place.
    UserData,
    /// Owner not known: kept in place and reported.
    Unknown,
}

/// One row of the HOME name table.
#[derive(Clone, Copy, Debug)]
pub struct Entry {
    /// A top-level name under `~/.moai`; one `*` stands for any run of characters.
    pub name: &'static str,
    pub rust: RustUse,
    /// The Go generation uses this name in its `~/.moai`.
    pub go: bool,
    pub class: Class,
    pub rule: Rule,
    /// A merge candidate (PLAN v2 table A, column "병합 후보"): the U4 step offers this item's
    /// user values from the backup manifest; only the rows marked here are offered.
    pub merge: bool,
    /// Who writes it (file:line at base `ba5ef23bab`, Go paths from PLAN v2 table A) and why the
    /// rule.
    pub why: &'static str,
}

const fn row(
    name: &'static str,
    rust: RustUse,
    go: bool,
    class: Class,
    rule: Rule,
    merge: bool,
    why: &'static str,
) -> Entry {
    Entry {
        name,
        rust,
        go,
        class,
        rule,
        merge,
        why,
    }
}

use Class::{ConfigGuide, Secret, ToolAsset, Unknown, UserData};
use Rule::*;
use RustUse::{Reads, Writes};

const MERGE: bool = true;
const NO: bool = false;

/// The HOME name table, row values from PLAN v2 6876772276 table A (class, action, merge
/// candidate; its open choices Z2, Z3, Z5 and Z8 at their defaults). Names are unique ignoring
/// case (the default macOS file system does not tell them apart), except that an exact name may
/// sit under a pattern with the same class and rule; [`lookup`] takes the exact name first.
pub const HOME_TABLE: &[Entry] = &[
    // This generation's names (PLAN 1-A).
    row("agents", Writes, false, Class::Rust, RustName, NO, "agent_install.rs:28,44"),
    row("backups", Writes, true, ToolAsset, GoRetires, NO, "update.rs:1213,2398, recovery.rs:100; Go homestate/paths.go:463 update backups"),
    row("bin", RustUse::None, true, ToolAsset, GoRetires, NO, "Go paths.go:461 (table A: tool asset, backed up). No Rust writer: the installed command is `path/` (update.rs:509, bin_names.rs:95) and `versions/<v>/bin`; design #3009 §2 (마) lists a device-global `bin`, which no code makes yet, so the Go folder retires and the name stays free for it"),
    row("cleanup", Writes, false, Class::Rust, RustName, NO, "update_cleanup.rs:39,894"),
    row("config", Writes, true, ConfigGuide, GoRetires, MERGE, "session.rs:209 config/launcher.yaml (a folder); Go paths.go:450 config/sections/*.yaml (merge: sections/statusline.yaml user values). The hook TOML file no longer lives at `config` (it is `hook-config.toml`, hook_config.rs); design #3009 §2 (마) keeps `config` as the device-global settings place"),
    row("hook-config.toml", Writes, false, Class::Rust, RustName, NO, "hook_config.rs: the hook TOML file; read here first, and from the old file name `config` only while that is a regular file (the folder `config/` of both generations cannot also be a file). The user writes it, but this generation owns the name, so the HOME move carries it (Codex review of #3019: after the marker the loader reads only `~/.moai`)"),
    row("home-routing", Writes, false, Class::Rust, RustName, NO, "home_routing.rs:30,134"),
    row("home-style", Writes, false, Class::Rust, RustName, NO, "home_style.rs:31,135"),
    row("installations", Writes, false, Class::Rust, RustName, NO, "install.rs:368,445, update.rs:960"),
    row("path", Writes, false, Class::Rust, RustName, NO, "update.rs:509,600, bin_names.rs:95; on PATH via shims"),
    row("projects", Writes, true, UserData, UserDataKeep, NO, "design #3009 §2 (나): `projects/<project_id>/` holds this generation's machine-written data (layout.rs:46,122,254); the same folder is the #2811 new-project parent `projects/<name>` (home_workdir.rs:30); Go carve-out clean_home.go:33 with another meaning (Z3: kept in place, never moved or retired). A child is never a table name"),
    row("receipts", Writes, false, Class::Rust, RustName, NO, "key_vault.rs:1001"),
    row("schedule", Writes, false, Class::Rust, RustName, NO, "gate_schedule.rs:372 schedule_dir (unit files, lock, receipts); spelled through a constant, so the src scan below does not see it"),
    row("state", Writes, true, ToolAsset, GoRetires, NO, "usage_limit.rs:50, statusline/state.rs:88; Go paths.go:456 state/active-sessions.json (a live registry: skip-in-use)"),
    row("vault", Writes, false, Class::Rust, SecretRenameOnly, NO, "key_vault.rs:37,239 (.master.key and entries; 0700/0600)"),
    row("versions", Writes, false, Class::Rust, RustName, NO, "update.rs:901, bin_names.rs:95"),
    row("worktrees", Writes, true, UserData, UserDataKeep, NO, "design #3009 §2 (라): `worktrees/<project_id>/<issue>-<key>/` (worktree_move.rs:175, home_workdir.rs); Go worktreeroot.go:70 shares the folder (#2811 D8: never moved or retired)"),
    row("moai-cli.json", Reads, false, Class::Rust, RustName, NO, "state_root.rs LEGACY_MARKER: the marker's earlier name (#2871, decision 6886824285 item 2); readers still accept it, and the only writer is home_install_apply.rs for an install an earlier build began (it ends under the name that goes with the saved marker bytes; its revert moves that file into the journal folder)"),
    row("moai.json", Writes, false, Class::Rust, RustName, NO, "state_root.rs NEW_MARKER / RUST_MARKER: the marker every writer writes (#2871, decision 6886824285 item 2)"),
    row("home-state-receipt.json", Writes, false, Class::Rust, RustName, NO, "home_state.rs:677"),
    // Go names (survey §1, PLAN v2 table A).
    row("settings.json", RustUse::None, true, ConfigGuide, GoRetires, MERGE, "Go paths.go:130. No Rust key receives it in rc.2 (leader decision 6903832699 item 1): home-retire moves the file to the Go backup and the manifest names it, so a user restores it by hand; the merge flag only marks it as a candidate in the preview"),
    // The pattern stands before the exact name on purpose: `lookup` must still pick the exact row.
    row(".env*", RustUse::None, true, Secret, SecretRenameOnly, NO, "Go key files such as .env.typesafe (defs/dirs.go:404)"),
    row(".env.glm", Reads, true, Secret, SecretRenameOnly, MERGE, "Go defs/dirs.go:401; key_vault.rs:773 reads it as a legacy source (merge: vault import only, the value never shown)"),
    row("credentials*", RustUse::None, true, Secret, SecretRenameOnly, NO, "Go paths.go:451 credentials.yaml(.bak)"),
    row("secrets", RustUse::None, true, Secret, SecretRenameOnly, NO, "Go secrets (no current writer, home_state.rs:71)"),
    row("gateway-auth", RustUse::None, true, Secret, SecretRenameOnly, NO, "Go gateway auth (no current writer)"),
    row("claude-profiles", RustUse::None, true, Unknown, UnknownKeep, NO, "Go paths.go:449 (Z2: owner unknown, kept and reported)"),
    row("cache", RustUse::None, true, ToolAsset, GoRetires, NO, "Go paths.go:459"),
    row("run", RustUse::None, true, ToolAsset, GoRetires, NO, "Go paths.go:322 run state (sqlite; skip-in-use)"),
    row("db", RustUse::None, true, ToolAsset, GoRetires, NO, "Go paths.go:286 databases (sqlite; skip-in-use)"),
    row("logs", Writes, true, ToolAsset, GoRetires, NO, "gate_schedule.rs:40,418 `logs/gate-loop.log` and its folder; Go paths.go:452 logs (Go update recreates it). Design #3009 §2 (나) moves machine logs under `projects/<project_id>/logs` in a later step; until then this name is shared like `backups` and `state`"),
    row("releases", RustUse::None, true, ToolAsset, GoRetires, NO, "Go paths.go:436"),
    row("evidence", RustUse::None, true, ToolAsset, GoRetires, NO, "Go paths.go:457"),
    row("rescue", RustUse::None, true, ToolAsset, GoRetires, NO, "home_state.rs SELECTIVE_TABLE `go_run_data` (the update step moves it to the backup; this table now agrees)"),
    row("scratch", RustUse::None, true, ToolAsset, GoRetires, NO, "home_state.rs SELECTIVE_TABLE `go_run_data` (same)"),
    row("tmp-*", RustUse::None, true, ToolAsset, GoRetires, NO, "home_state.rs SELECTIVE_TABLE `go_run_data` (same)"),
    row("todo", RustUse::None, true, ToolAsset, GoRetires, NO, "Go factory/state_dir.go:43-56"),
    row("integrations", RustUse::None, true, Unknown, UnknownKeep, NO, "Go paths.go:460 (Z2: owner unknown, kept and reported)"),
    row("user-assets*.json", RustUse::None, true, ToolAsset, GoRetires, NO, "Go userassets/paths.go:192 asset manifest (a Go HOME marker; read before the backup)"),
    row("user-assets.acquire-guard.guard", RustUse::None, true, ToolAsset, GoRetires, NO, "Go lock_guard_unix.go:24 asset lock"),
    row(".migrate-tx-*.json", RustUse::None, true, ToolAsset, GoRetires, NO, "Go migrate_home_state.go:331 migration journals"),
    row("specs", RustUse::None, true, UserData, UserDataKeep, NO, "user SPECs (D-M1 item 3)"),
    row("reports", RustUse::None, true, Unknown, UnknownKeep, NO, "Go paths.go:464 (Z2: owner unknown, kept and reported)"),
    row("handoffs", RustUse::None, true, Unknown, UnknownKeep, NO, "no current writer"),
    row("kanban", RustUse::None, true, Unknown, UnknownKeep, NO, "no current writer"),
    row("backup", RustUse::None, true, Unknown, UnknownKeep, NO, "no current writer"),
    row("dev-login.sh", RustUse::None, true, Unknown, UnknownKeep, NO, "no current writer"),
    row("gpt-appserver", RustUse::None, true, Unknown, UnknownKeep, NO, "no current writer"),
    row("local-plugins", RustUse::None, true, Unknown, UnknownKeep, NO, "no current writer"),
    row("mcp", RustUse::None, true, Unknown, UnknownKeep, NO, "no current writer"),
    row("search", RustUse::None, true, Unknown, UnknownKeep, NO, "no current writer"),
    row("studio", RustUse::None, true, Unknown, UnknownKeep, NO, "no current writer"),
    row("holding", Writes, false, Class::Rust, RustName, NO, "hygiene.rs:38 `holding/<UTC date>/` (held worktrees and build folders, purged after 3 days); another tool (lsof: hided) also holds the folder open (Z8): nothing here retires or moves it"),
    row("locks", RustUse::None, true, Unknown, UnknownKeep, NO, "no current writer"),
];

/// `name` matches `pattern` ignoring ASCII case; one `*` in the pattern matches any run.
pub fn matches(pattern: &str, name: &str) -> bool {
    let (p, n) = (pattern.to_ascii_lowercase(), name.to_ascii_lowercase());
    match p.split_once('*') {
        None => p == n,
        Some((head, tail)) => {
            n.len() >= head.len() + tail.len() && n.starts_with(head) && n.ends_with(tail)
        }
    }
}

/// The row for a top-level name, ignoring case, an exact name before a pattern; `None` for a
/// name the table does not know.
pub fn lookup(name: &str) -> Option<&'static Entry> {
    HOME_TABLE
        .iter()
        .find(|e| !e.name.contains('*') && matches(e.name, name))
        .or_else(|| HOME_TABLE.iter().find(|e| matches(e.name, name)))
}

/// The folder this generation's HOME items live in under `layout`. Pure.
pub fn root_in(layout: Layout, home: &Path) -> PathBuf {
    let moai = home.join(crate::state_root::MOAI_DIR);
    match layout {
        Layout::TopLevel => moai,
        Layout::Subdir(sub) => moai.join(sub),
    }
}

/// `root_in(layout, home)/name`; an empty `name` is the folder itself. Pure.
pub fn rust_path_in(layout: Layout, home: &Path, name: &str) -> PathBuf {
    let root = root_in(layout, home);
    if name.is_empty() {
        root
    } else {
        root.join(name)
    }
}

/// Where this generation's HOME item `name` lives on a HOME that uses this layout
/// ([`RUST_HOME_LAYOUT`]). Pure.
pub fn rust_path(home: &Path, name: &str) -> PathBuf {
    rust_path_in(RUST_HOME_LAYOUT, home, name)
}

/// The folders on the way to this generation's HOME root are real folders (or absent), never
/// links. Reads metadata only.
pub fn check_root_in(layout: Layout, home: &Path) -> Result<(), String> {
    let moai = crate::state_root::MOAI_DIR;
    let mut steps = vec![(home.join(moai), moai.to_string())];
    if let Layout::Subdir(sub) = layout {
        steps.push((root_in(layout, home), format!("{moai}/{sub}")));
    }
    for (path, name) in steps {
        match std::fs::symlink_metadata(&path) {
            Err(e) if e.kind() == std::io::ErrorKind::NotFound => return Ok(()),
            Err(_) => return Err(format!("~/{name} cannot be read")),
            // `symlink_metadata` describes a link as a link, so `is_dir` refuses one.
            Ok(m) if !m.is_dir() => {
                return Err(format!(
                    "~/{name} is not a folder (a link is never followed)"
                ))
            }
            Ok(_) => {}
        }
    }
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::fs;

    /// Top-level names this generation spells through constants or helpers today (the src scan
    /// below sees only literals joined to `home_root::path(…)` on one line).
    const RUST_TOP_NAMES: &[&str] = &[
        "agents",
        "backups",
        "cleanup",
        "config",
        "home-routing",
        "home-style",
        "installations",
        "holding",
        "logs",
        "path",
        "projects",
        "receipts",
        "schedule",
        "state",
        "vault",
        "versions",
        "worktrees",
        "moai-cli.json",
        "moai.json",
        "home-state-receipt.json",
        ".env.glm",
    ];

    /// Every `home_root::path(…).join("<lit>")` in src, first path component of `<lit>`.
    fn scanned_names() -> std::collections::BTreeSet<String> {
        let root = Path::new(env!("CARGO_MANIFEST_DIR")).join("src");
        let mut stack = vec![root];
        let mut out = std::collections::BTreeSet::new();
        let mut files = 0;
        while let Some(d) = stack.pop() {
            for e in fs::read_dir(&d).unwrap() {
                let p = e.unwrap().path();
                if p.is_dir() {
                    stack.push(p);
                    continue;
                }
                if p.extension().is_none_or(|x| x != "rs") {
                    continue;
                }
                files += 1;
                for line in fs::read_to_string(&p).unwrap().lines() {
                    if line.trim_start().starts_with("//") {
                        continue;
                    }
                    let Some(at) = line.find("home_root::path(") else {
                        continue;
                    };
                    // The join right after the call's closing parenthesis (arguments hold none).
                    let rest = &line[at + "home_root::path(".len()..];
                    let Some(close) = rest.find(')') else {
                        continue;
                    };
                    let Some(lit) = rest[close + 1..].strip_prefix(".join(\"") else {
                        continue;
                    };
                    let lit = &lit[..lit.find('"').unwrap()];
                    out.insert(lit.split('/').next().unwrap().to_string());
                }
            }
        }
        assert!(files > 100, "the walk saw {files} files; src moved?");
        out
    }

    /// Both a src scan (literals) and an explicit list (names behind constants).
    #[test]
    fn every_rust_written_home_name_is_in_the_table() {
        let scanned = scanned_names();
        for want in ["backups", "config", "installations", "state", "versions"] {
            assert!(scanned.contains(want), "the scan lost {want}: {scanned:?}");
        }
        for name in scanned
            .iter()
            .map(String::as_str)
            .chain(RUST_TOP_NAMES.iter().copied())
        {
            let e = lookup(name).unwrap_or_else(|| panic!("{name} is not in HOME_TABLE"));
            // The two read-only names: the legacy key file and the marker's earlier name.
            if name != ".env.glm" && name != "moai-cli.json" {
                assert_eq!(e.rust, RustUse::Writes, "{name}");
            }
        }
        assert_eq!(lookup("moai.json").unwrap().rust, RustUse::Writes);
        assert_eq!(lookup("moai-cli.json").unwrap().rust, RustUse::Reads);
        assert_eq!(lookup("MOAI.JSON").unwrap().rust, RustUse::Writes);
        assert!(lookup("not-a-moai-name").is_none());
    }

    #[test]
    fn table_names_are_unique_ignoring_case() {
        for (i, a) in HOME_TABLE.iter().enumerate() {
            for b in &HOME_TABLE[i + 1..] {
                let lit = |e: &Entry| e.name.replace('*', "x");
                if !matches(a.name, &lit(b)) && !matches(b.name, &lit(a)) {
                    continue;
                }
                // Only an exact name under a pattern of the same class and rule.
                let one_pattern = a.name.contains('*') != b.name.contains('*');
                assert!(
                    one_pattern && a.class == b.class && a.rule == b.rule,
                    "{} and {} overlap",
                    a.name,
                    b.name
                );
            }
        }
        // The exact name wins over the pattern it sits under.
        assert!(lookup(".ENV.GLM").unwrap().merge);
        assert!(!lookup(".env.typesafe").unwrap().merge);
        assert_eq!(lookup(".env.typesafe").unwrap().name, ".env*");
    }

    /// #2871 (decision 6887464608): the update step's table (`home_state::SELECTIVE_TABLE`) and
    /// this one name the same folders; a name in the first has a row here with the matching
    /// action (a secret is renamed, a work folder stays, run data retires to the backup).
    #[test]
    fn the_update_table_and_the_home_table_agree() {
        use crate::home_state::{EntryClass, SELECTIVE_TABLE};
        for (pattern, class, _) in SELECTIVE_TABLE {
            let sample = pattern.replace('*', "x");
            let e = lookup(&sample).unwrap_or_else(|| panic!("{pattern} has no HOME_TABLE row"));
            let want = match class {
                EntryClass::SecretManual => SecretRenameOnly,
                EntryClass::KeepInPlace => UserDataKeep,
                EntryClass::MoveToBackup => GoRetires,
                EntryClass::Unknown => UnknownKeep,
            };
            assert_eq!(e.rule, want, "{pattern} (as {sample})");
        }
    }

    /// The work folders are written by this generation and never moved or retired (design #3009
    /// §2 (나)/(라)); the other mismatch names say what the code does.
    #[test]
    fn the_names_of_design_3009_match_what_the_code_does() {
        for name in ["projects", "worktrees"] {
            let e = lookup(name).unwrap();
            assert_eq!(
                (e.rust, e.go, e.class, e.rule),
                (Writes, true, UserData, UserDataKeep),
                "{name}"
            );
        }
        // `logs`: gate_schedule writes `logs/gate-loop.log`; Go's folder of that name retires.
        let logs = lookup("logs").unwrap();
        assert_eq!((logs.rust, logs.rule), (Writes, GoRetires));
        // `holding`: hygiene writes `holding/<UTC date>/`; nothing retires or moves it.
        let holding = lookup("holding").unwrap();
        assert_eq!(
            (holding.rust, holding.go, holding.class, holding.rule),
            (Writes, false, Class::Rust, RustName)
        );
        // `config` is the launcher folder; the hook file has its own name and is only read.
        assert_eq!(lookup("config").unwrap().rule, GoRetires);
        let hook = lookup(crate::hook_config::HOOK_CONFIG_FILE).unwrap();
        assert_eq!((hook.rust, hook.rule), (Writes, RustName));
        // `bin`: no Rust writer (the command lives in `path/`), so the Go folder retires.
        assert_eq!(lookup("bin").unwrap().rust, RustUse::None);
    }

    /// The class decides the action (PLAN v2 table A): a secret is only renamed, user data and
    /// unknown items stay, Go tool assets and settings retire to the backup, and this
    /// generation's own rows are never Go's.
    #[test]
    fn every_class_has_its_rule() {
        for e in HOME_TABLE {
            let ok = match e.class {
                Secret => e.rule == SecretRenameOnly && e.go,
                UserData => e.rule == UserDataKeep,
                Unknown => e.rule == UnknownKeep && e.go,
                ToolAsset | ConfigGuide => e.rule == GoRetires && e.go,
                Class::Rust => {
                    !e.go
                        && (e.rule == RustName || (e.name == "vault" && e.rule == SecretRenameOnly))
                }
            };
            assert!(ok, "{}: {:?} with {:?}", e.name, e.class, e.rule);
        }
    }

    /// Row values pinned to PLAN v2 6876772276 table A: (name looked up, class, rule, Go uses it).
    #[test]
    fn rows_follow_plan_v2_table_a() {
        let want: &[(&str, Class, Rule, bool)] = &[
            ("backups", ToolAsset, GoRetires, true),
            ("bin", ToolAsset, GoRetires, true),
            ("config", ConfigGuide, GoRetires, true),
            ("state", ToolAsset, GoRetires, true),
            ("db", ToolAsset, GoRetires, true),
            ("run", ToolAsset, GoRetires, true),
            // `logs` is written by gate_schedule too (#2871 decision 6887464608): shared like
            // `backups` and `state`, so Go's folder retires and this generation takes the name.
            ("logs", ToolAsset, GoRetires, true),
            ("releases", ToolAsset, GoRetires, true),
            ("cache", ToolAsset, GoRetires, true),
            ("evidence", ToolAsset, GoRetires, true),
            ("todo", ToolAsset, GoRetires, true),
            (
                "user-assets.acquire-guard.guard",
                ToolAsset,
                GoRetires,
                true,
            ),
            ("user-assets.json", ToolAsset, GoRetires, true),
            ("settings.json", ConfigGuide, GoRetires, true),
            (".env.glm", Secret, SecretRenameOnly, true),
            (".env.typesafe", Secret, SecretRenameOnly, true),
            ("credentials", Secret, SecretRenameOnly, true),
            ("credentials.yaml.bak", Secret, SecretRenameOnly, true),
            ("secrets", Secret, SecretRenameOnly, true),
            ("gateway-auth", Secret, SecretRenameOnly, true),
            ("worktrees", UserData, UserDataKeep, true),
            ("projects", UserData, UserDataKeep, true),
            ("claude-profiles", Unknown, UnknownKeep, true),
            ("integrations", Unknown, UnknownKeep, true),
            ("reports", Unknown, UnknownKeep, true),
            // hygiene.rs writes `holding/<UTC date>/`; the name is this generation's.
            ("holding", Class::Rust, RustName, false),
        ];
        for &(name, class, rule, go) in want {
            let e = lookup(name).unwrap_or_else(|| panic!("{name} is not in HOME_TABLE"));
            assert_eq!((e.class, e.rule, e.go), (class, rule, go), "{name}");
        }
        for name in [
            "kanban",
            "mcp",
            "search",
            "studio",
            "backup",
            "locks",
            "gpt-appserver",
            "handoffs",
            "local-plugins",
            "dev-login.sh",
        ] {
            assert_eq!(lookup(name).unwrap().class, Unknown, "{name}");
        }
        for name in [
            "agents",
            "cleanup",
            "home-routing",
            "home-style",
            "installations",
            "path",
            "receipts",
            "vault",
            "versions",
            "moai-cli.json",
            "moai.json",
        ] {
            assert_eq!(lookup(name).unwrap().class, Class::Rust, "{name}");
        }
    }

    /// The merge candidates are exactly the "예" rows of PLAN v2 table A.
    #[test]
    fn merge_candidates_are_the_plan_rows() {
        let merge: Vec<&str> = HOME_TABLE
            .iter()
            .filter(|e| e.merge)
            .map(|e| e.name)
            .collect();
        assert_eq!(merge, ["config", "settings.json", ".env.glm"]);
    }

    /// A name both generations use is declared as a clash with a rule that says who yields;
    /// this generation never takes a Go secret, and its own names never collide with a Go name
    /// it did not declare, whatever the case.
    #[test]
    fn rust_names_never_collide_with_go_names_undeclared() {
        for e in HOME_TABLE.iter().filter(|e| e.rust == RustUse::Writes) {
            for spelled in [e.name.to_string(), e.name.to_ascii_uppercase()] {
                let hits: Vec<_> = HOME_TABLE
                    .iter()
                    .filter(|g| g.go && matches(g.name, &spelled))
                    .collect();
                assert!(hits.len() <= 1, "{spelled}: {hits:?}");
                if let Some(g) = hits.first() {
                    assert_eq!(g.name, e.name, "{spelled} collides with Go {}", g.name);
                    assert!(
                        matches!(e.rule, GoRetires | UserDataKeep | UnknownKeep),
                        "{}: a clash needs a yielding rule",
                        e.name
                    );
                }
            }
        }
        for secret in [
            ".env.glm",
            "Credentials.yaml",
            "SECRETS",
            "gateway-auth",
            "vault",
        ] {
            assert_eq!(lookup(secret).unwrap().rule, SecretRenameOnly, "{secret}");
        }
        for g in [
            "cache",
            "run",
            "db",
            "releases",
            "integrations",
            "user-assets.json",
        ] {
            let e = lookup(g).unwrap();
            assert!(e.go && e.rust != RustUse::Writes, "{g}");
        }
        assert_eq!(lookup("user-assets-2.json").unwrap().rule, GoRetires);
        assert_eq!(lookup(".migrate-tx-20261114.json").unwrap().rule, GoRetires);
        assert!(lookup(".migrate-tx-.jsonx").is_none());
    }

    #[test]
    fn layout_is_top_level_and_paths_follow_it() {
        let h = Path::new("/h");
        assert_eq!(RUST_HOME_LAYOUT, Layout::TopLevel);
        assert_eq!(rust_path(h, "vault"), PathBuf::from("/h/.moai/vault"));
        assert_eq!(rust_path(h, ""), PathBuf::from("/h/.moai"));
        assert_eq!(
            rust_path_in(Layout::Subdir("cli"), h, "vault"),
            PathBuf::from("/h/.moai/cli/vault")
        );
        assert_eq!(
            root_in(Layout::Subdir("cli"), h),
            PathBuf::from("/h/.moai/cli")
        );
        assert_eq!(
            rust_path_in(Layout::TopLevel, h, "state"),
            PathBuf::from("/h/.moai/state")
        );
    }

    #[cfg(unix)]
    #[test]
    fn a_linked_rust_root_is_refused() {
        use std::os::unix::fs::symlink;
        let home = tempfile::tempdir().unwrap();
        let h = home.path();
        // Absent and real folders pass.
        assert_eq!(check_root_in(Layout::Subdir("cli"), h), Ok(()));
        fs::create_dir_all(h.join(".moai/cli")).unwrap();
        assert_eq!(check_root_in(Layout::Subdir("cli"), h), Ok(()));
        assert_eq!(check_root_in(Layout::TopLevel, h), Ok(()));
        // A linked subfolder, a file in its place, and a linked `.moai`.
        fs::remove_dir(h.join(".moai/cli")).unwrap();
        fs::create_dir(h.join("elsewhere")).unwrap();
        symlink(h.join("elsewhere"), h.join(".moai/cli")).unwrap();
        assert!(check_root_in(Layout::Subdir("cli"), h).is_err());
        fs::remove_file(h.join(".moai/cli")).unwrap();
        fs::write(h.join(".moai/cli"), "x").unwrap();
        assert!(check_root_in(Layout::Subdir("cli"), h).is_err());
        let linked = tempfile::tempdir().unwrap();
        symlink(h.join("elsewhere"), linked.path().join(".moai")).unwrap();
        assert!(check_root_in(Layout::TopLevel, linked.path()).is_err());
        assert!(check_root_in(Layout::Subdir("cli"), linked.path()).is_err());
    }
}
