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
    Writes,
    /// Reads only (a legacy source).
    Reads,
    None,
}

/// One row of the HOME name table.
#[derive(Clone, Copy, Debug)]
pub struct Entry {
    /// A top-level name under `~/.moai`; one `*` stands for any run of characters.
    pub name: &'static str,
    pub rust: RustUse,
    /// The Go generation uses this name in its `~/.moai`.
    pub go: bool,
    pub rule: Rule,
    /// Who writes it (file:line at base `ba5ef23bab`) and why the rule.
    pub why: &'static str,
}

const fn row(name: &'static str, rust: RustUse, go: bool, rule: Rule, why: &'static str) -> Entry {
    Entry {
        name,
        rust,
        go,
        rule,
        why,
    }
}

use Rule::*;
use RustUse::{Reads, Writes};

/// The HOME name table. Names are unique ignoring case (the default macOS file system does not
/// tell them apart).
pub const HOME_TABLE: &[Entry] = &[
    // This generation's names (PLAN 1-A).
    row("agents", Writes, false, RustName, "agent_install.rs:28,44"),
    row("backups", Writes, true, GoRetires, "update.rs:1213,2398, recovery.rs:100; Go update backups (clean --home removes backups/removed-*, clean_home.go:147-153)"),
    row("bin", RustUse::None, true, UnknownKeep, "no writer in src (only versions/<v>/bin); Go EnsureHomeLayout makes it (clean_home.go:38); owner unknown"),
    row("cleanup", Writes, false, RustName, "update_cleanup.rs:39,894"),
    row("config", Writes, true, GoRetires, "session.rs:209 config/launcher.yaml; Go config/sections/*.yaml; the hook TOML file moves to hook-config.toml"),
    row("hook-config.toml", RustUse::None, false, RustName, "new name for the hook TOML file now at `config` (hook_config.rs:112), which clashes with the config/ folder of both generations; rule only, nothing renamed yet"),
    row("home-routing", Writes, false, RustName, "home_routing.rs:30,134"),
    row("home-style", Writes, false, RustName, "home_style.rs:31,135"),
    row("installations", Writes, false, RustName, "install.rs:368,445, update.rs:960"),
    row("path", Writes, false, RustName, "update.rs:509,600, bin_names.rs:95; on PATH via shims"),
    row("projects", Writes, false, RustName, "layout.rs:46,122,254 (home project store) and home_workdir.rs:30 (#2811 new-project parent)"),
    row("receipts", Writes, false, RustName, "key_vault.rs:1001"),
    row("state", Writes, true, GoRetires, "usage_limit.rs:50, statusline/state.rs:88; Go state/active-sessions.json (a live registry: skip-in-use)"),
    row("vault", Writes, false, SecretRenameOnly, "key_vault.rs:37,239 (.master.key and entries; 0700/0600)"),
    row("versions", Writes, false, RustName, "update.rs:901, bin_names.rs:95"),
    row("worktrees", Writes, true, UserDataKeep, "worktree_move.rs:175, home_workdir.rs (#2811 D8: shared; Go worktrees/projects never moved)"),
    row("moai-cli.json", Writes, false, RustName, "state_root.rs:35 marker"),
    row("home-state-receipt.json", Writes, false, RustName, "home_state.rs:677"),
    // Go names (survey §1, PLAN 1-A list).
    row(".env*", Reads, true, SecretRenameOnly, "Go key files; key_vault.rs:773 reads .env.glm as a legacy source"),
    row("credentials*", RustUse::None, true, SecretRenameOnly, "Go credentials.yaml(.bak)"),
    row("secrets", RustUse::None, true, SecretRenameOnly, "Go secrets"),
    row("gateway-auth", RustUse::None, true, SecretRenameOnly, "Go gateway auth"),
    row("claude-profiles", RustUse::None, true, SecretRenameOnly, "Go Claude profile homes (sign-ins inside)"),
    row("cache", RustUse::None, true, GoRetires, "Go cache"),
    row("run", RustUse::None, true, GoRetires, "Go run state (sqlite; skip-in-use)"),
    row("db", RustUse::None, true, GoRetires, "Go databases (sqlite; skip-in-use)"),
    row("logs", RustUse::None, true, GoRetires, "Go logs (Go update recreates it)"),
    row("releases", RustUse::None, true, GoRetires, "Go releases"),
    row("integrations", RustUse::None, true, GoRetires, "Go integrations"),
    row("user-assets*.json", RustUse::None, true, GoRetires, "Go asset manifest (a Go HOME marker)"),
    row("user-assets.acquire-guard.guard", RustUse::None, true, GoRetires, "Go asset lock"),
    row(".migrate-tx-*.json", RustUse::None, true, GoRetires, "Go migration journals"),
    row("specs", RustUse::None, true, UserDataKeep, "user SPECs (D-M1 item 3)"),
    row("reports", RustUse::None, true, UserDataKeep, "user reports (D-M1 item 3)"),
    row("handoffs", RustUse::None, true, UserDataKeep, "session handoff notes"),
    row("todo", RustUse::None, true, UserDataKeep, "user todo"),
    row("kanban", RustUse::None, true, UserDataKeep, "user board"),
    row("backup", RustUse::None, true, UnknownKeep, "owner unknown"),
    row("dev-login.sh", RustUse::None, true, UnknownKeep, "owner unknown"),
    row("gpt-appserver", RustUse::None, true, UnknownKeep, "owner unknown"),
    row("local-plugins", RustUse::None, true, UnknownKeep, "owner unknown"),
    row("mcp", RustUse::None, true, UnknownKeep, "owner unknown"),
    row("search", RustUse::None, true, UnknownKeep, "owner unknown"),
    row("studio", RustUse::None, true, UnknownKeep, "owner unknown"),
    row("holding", RustUse::None, true, UnknownKeep, "owner unknown"),
    row("locks", RustUse::None, true, UnknownKeep, "owner unknown"),
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

/// The row for a top-level name, ignoring case; `None` for a name the table does not know.
pub fn lookup(name: &str) -> Option<&'static Entry> {
    HOME_TABLE.iter().find(|e| matches(e.name, name))
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
        "path",
        "projects",
        "receipts",
        "state",
        "vault",
        "versions",
        "worktrees",
        "moai-cli.json",
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
            if name != ".env.glm" {
                assert_eq!(e.rust, RustUse::Writes, "{name}");
            }
        }
        assert!(lookup("not-a-moai-name").is_none());
    }

    #[test]
    fn table_names_are_unique_ignoring_case() {
        for (i, a) in HOME_TABLE.iter().enumerate() {
            for b in &HOME_TABLE[i + 1..] {
                let lit = |e: &Entry| e.name.replace('*', "x");
                assert!(
                    !matches(a.name, &lit(b)) && !matches(b.name, &lit(a)),
                    "{} and {} overlap",
                    a.name,
                    b.name
                );
            }
        }
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
            "logs",
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
