//! Opt-in hook rules from `hook-config.toml` in the HOME state folder, or from the older file
//! name `config` there (Issue #2391 stage 7; user decision, Issue #2391 comment 6779829826,
//! design 6777574227; the new name is #2871 decision 6887464608: `config` is also the folder
//! of `config/launcher.yaml`, and a name cannot be both a file and a folder). The new name
//! wins when it exists; the old name is read only when it is not a folder. Both rules are off
//! unless the user turns them on:
//!
//! ```toml
//! [hooks.outside_project_writes]   # B9
//! enabled = true
//! allow = ["~/.claude/", "~/.codex/", "~/.moai-cli/", "/tmp/", "/private/tmp/"]
//!
//! [hooks.primary_checkout]         # B3
//! enabled = true
//! repos = ["/Users/me/src/moai-adk"]
//! allow_commands = ["git switch v2026.11.0"]
//! deny_commits_on = ["main"]       # #2878: no commit/revert/cherry-pick on these branches
//!
//! [hooks.commit_identity]          # #2878: no commit under a test identity
//! enabled = true
//! deny_emails = ["ci@example.com"] # added to the built-in test addresses
//!
//! [hooks.edit_diagnostics]         # SPEC-20261117-edit-diagnostics S1 (Claude Code)
//! enabled = true
//! timeout_secs = 10                # 1-30
//! max_lines = 20                   # 1-200
//! max_chars = 4000                 # 200-9000
//! checkers = { py = ["ruff", "check", "--output-format", "concise"] }  # file appended
//!
//! [hooks.modes]                    # card C-86, leader decision note 2346
//! merge_confirm = "enforce"        # off | observe | enforce; absent or anything else: observe
//! ```
//!
//! `[hooks.modes]` sets, per rule id, how a refusing rule acts ([`RuleMode`]): `off` does
//! nothing, `observe` (the default, also for a value that is not one of the three words) logs
//! what the rule would have done and lets the call through, `enforce` acts. It never turns a
//! rule that has its own `enabled` switch on or off.
//!
//! A malformed `[hooks.edit_diagnostics]` turns only that feature off, named by
//! [`ConfigProblem::InvalidEditDiagnostics`]; the guard rules keep their own settings.
//!
//! A config that is absent, not a regular file (a symlink too), over 64 KiB or not valid TOML
//! turns both rules off: a broken config must not block every write (the engine's fail-open
//! rule, D1). A rule whose own table is malformed (a non-boolean `enabled`, a non-string list
//! entry) is off as well. An unusable file (not a regular file, too large, unreadable, not
//! valid TOML) is also named by a [`ConfigProblem`], which the engine logs as one line (#2391
//! N1, design 6777574227) so a rule the user believes on is not silently off. An absent file
//! is not a problem.
//!
//! The file is opened once without following a symlink (on unix by the open itself; elsewhere
//! by a look just before it), checked on that handle, and read
//! through a `CONFIG_MAX + 1` bound, so a file swapped or grown after the check is still
//! refused (#2391 N2).

use std::fs;
use std::io::Read;
use std::path::{Component, Path, PathBuf};

/// Largest config file read.
pub const CONFIG_MAX: u64 = 65_536;

/// B9: writes outside the project are refused unless under one of `allow`.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct OutsideRule {
    /// Absolute, normalized folder prefixes (`~` expanded).
    pub allow: Vec<PathBuf>,
}

/// B3: switching, resetting, stashing or rebasing a shared primary checkout is refused.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct PrimaryRule {
    /// Primary worktree folders the rule covers (absolute, normalized, `~` expanded).
    pub repos: Vec<PathBuf>,
    /// Exact command lines that are let through.
    pub allow_commands: Vec<String>,
    /// Branches on which a primary checkout takes no new commit, revert or cherry-pick
    /// (#2878 E2, the Go generation's `deny_commits_on`). Empty: none.
    pub deny_commits_on: Vec<String>,
}

/// The commit-identity rule (#2878 E3, the Go generation's `commit_identity_guard`).
#[derive(Clone, Debug, Default, PartialEq, Eq)]
pub struct IdentityRule {
    /// Extra author or committer emails to refuse, besides the built-in test addresses;
    /// compared trimmed and without case.
    pub deny_emails: Vec<String>,
}

/// SPEC-20261117-edit-diagnostics S1: after a file edit, run the checker the user named for the
/// file's extension and hand its diagnostics back as context. Off unless enabled.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct EditRule {
    /// Seconds the checker may run, [`EDIT_TIMEOUT_SECS`] by default.
    pub timeout_secs: u64,
    /// Diagnostic lines returned at most.
    pub max_lines: usize,
    /// Characters returned at most (summary included).
    pub max_chars: usize,
    /// Lower-case file extension -> checker argv (the edited file is appended as the last
    /// argument; no shell).
    pub checkers: Vec<(String, Vec<String>)>,
}

/// The defaults and bounds of [`EditRule`] (SPEC-20261117-edit-diagnostics D3, D4).
pub const EDIT_TIMEOUT_SECS: (u64, u64, u64) = (10, 1, 30);
pub const EDIT_MAX_LINES: (usize, usize, usize) = (20, 1, 200);
pub const EDIT_MAX_CHARS: (usize, usize, usize) = (4_000, 200, 9_000);

/// How a refusing hook rule acts (card C-86, leader decision note 2346): `Observe` records
/// what it would have refused and lets the call through, and is the default.
#[derive(Clone, Copy, Debug, Default, PartialEq, Eq)]
pub enum RuleMode {
    Off,
    #[default]
    Observe,
    Enforce,
}

impl RuleMode {
    /// The mode a setting names, or `None` for anything but `off`, `observe` and `enforce`.
    pub fn parse(word: &str) -> Option<RuleMode> {
        match word {
            "off" => Some(RuleMode::Off),
            "observe" => Some(RuleMode::Observe),
            "enforce" => Some(RuleMode::Enforce),
            _ => None,
        }
    }
}

#[derive(Clone, Debug, Default, PartialEq, Eq)]
pub struct HookConfig {
    /// `[hooks.modes]`: rule id → mode. A rule without an entry is in [`RuleMode::Observe`].
    pub modes: std::collections::BTreeMap<String, RuleMode>,
    pub outside: Option<OutsideRule>,
    pub primary: Option<PrimaryRule>,
    pub identity: Option<IdentityRule>,
    /// `[hooks.edit_diagnostics]` when it is enabled and well formed.
    pub edit: Option<EditRule>,
    /// `[hooks.edit_diagnostics]` is enabled but malformed: the feature is off and the engine
    /// logs one `config_error` line ([`ConfigProblem::InvalidEditDiagnostics`]).
    pub edit_invalid: bool,
}

/// Why a present config file was not used (both rules are then off).
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum ConfigProblem {
    /// A symlink, a directory, a FIFO or any other non-regular file.
    NotRegularFile,
    /// More than [`CONFIG_MAX`] bytes.
    TooLarge,
    /// The file could not be opened or read.
    Unreadable,
    /// Not UTF-8 or not valid TOML.
    InvalidToml,
    /// The file is usable but `[hooks.edit_diagnostics]` is enabled and malformed; only that
    /// feature is off.
    InvalidEditDiagnostics,
}

impl ConfigProblem {
    /// The fixed reason word written to the engine log.
    pub fn reason(self) -> &'static str {
        match self {
            ConfigProblem::NotRegularFile => "not_regular_file",
            ConfigProblem::TooLarge => "too_large",
            ConfigProblem::Unreadable => "unreadable",
            ConfigProblem::InvalidToml => "invalid_toml",
            ConfigProblem::InvalidEditDiagnostics => "invalid_edit_diagnostics",
        }
    }
}

/// A loaded config and, when a present file was unusable, why.
#[derive(Clone, Debug, Default, PartialEq, Eq)]
pub struct Loaded {
    pub config: HookConfig,
    pub problem: Option<ConfigProblem>,
}

/// The file name in the HOME state folder (#2871 decision 6887464608).
pub const HOOK_CONFIG_FILE: &str = "hook-config.toml";
/// The name before it; read only while it is not a folder.
const OLD_HOOK_CONFIG_NAME: &str = "config";

/// The config of `HOME` (both rules off when `HOME` is unset or the file is unusable).
pub fn load() -> Loaded {
    let Some(home) = std::env::var_os("HOME").filter(|h| !h.is_empty()) else {
        return Loaded::default();
    };
    let home = PathBuf::from(home);
    let root = crate::home_root::path(&home);
    let new = root.join(HOOK_CONFIG_FILE);
    // Present in any form (a link or a folder too): `load_path` names the problem.
    if fs::symlink_metadata(&new).is_ok() {
        return load_path(&new, &home);
    }
    let old = root.join(OLD_HOOK_CONFIG_NAME);
    // The folder `config/` (`config/launcher.yaml`, both generations) is not the old file.
    if fs::symlink_metadata(&old).is_ok_and(|m| m.is_dir()) {
        return Loaded::default();
    }
    load_path(&old, &home)
}

/// The config at `path`; `home` expands a leading `~`.
pub fn load_path(path: &Path, home: &Path) -> Loaded {
    let off = |problem| Loaded {
        config: HookConfig::default(),
        problem: Some(problem),
    };
    let file = match open_nofollow(path) {
        Ok(f) => f,
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => return Loaded::default(),
        // The open refused a symlink (O_NOFOLLOW); the extra look only names the reason.
        Err(_) if fs::symlink_metadata(path).is_ok_and(|m| m.file_type().is_symlink()) => {
            return off(ConfigProblem::NotRegularFile)
        }
        Err(_) => return off(ConfigProblem::Unreadable),
    };
    match file.metadata() {
        Ok(m) if !m.is_file() => return off(ConfigProblem::NotRegularFile),
        Ok(m) if m.len() > CONFIG_MAX => return off(ConfigProblem::TooLarge),
        Ok(_) => {}
        Err(_) => return off(ConfigProblem::Unreadable),
    }
    let text = match read_bounded(file) {
        Ok(Some(text)) => text,
        Ok(None) => return off(ConfigProblem::TooLarge),
        Err(e) if e.kind() == std::io::ErrorKind::InvalidData => {
            return off(ConfigProblem::InvalidToml)
        }
        Err(_) => return off(ConfigProblem::Unreadable),
    };
    match parse_checked(&text, home) {
        Some(config) => Loaded {
            config,
            problem: None,
        },
        None => off(ConfigProblem::InvalidToml),
    }
}

/// Opens a file for reading without following a symlink. On unix the open itself refuses a
/// symlink (`O_NOFOLLOW`) and does not block on a FIFO (`O_NONBLOCK`). Elsewhere a symlink is
/// refused by a look before the open, as before N2; that look and the open are two steps.
fn open_nofollow(path: &Path) -> std::io::Result<fs::File> {
    let mut options = fs::OpenOptions::new();
    options.read(true);
    #[cfg(unix)]
    {
        use std::os::unix::fs::OpenOptionsExt;
        options.custom_flags(libc::O_NOFOLLOW | libc::O_NONBLOCK);
    }
    #[cfg(not(unix))]
    if fs::symlink_metadata(path)?.file_type().is_symlink() {
        return Err(std::io::Error::other("config is a symlink"));
    }
    options.open(path)
}

/// Reads at most [`CONFIG_MAX`] bytes as UTF-8: `Ok(None)` when there are more, whatever size
/// was seen before the read; `InvalidData` when the bytes are not UTF-8.
fn read_bounded(reader: impl Read) -> std::io::Result<Option<String>> {
    let mut bytes = Vec::new();
    reader.take(CONFIG_MAX + 1).read_to_end(&mut bytes)?;
    if bytes.len() as u64 > CONFIG_MAX {
        return Ok(None);
    }
    String::from_utf8(bytes)
        .map(Some)
        .map_err(|_| std::io::Error::from(std::io::ErrorKind::InvalidData))
}

/// Parses the config text; `home` expands a leading `~`.
pub fn parse(text: &str, home: &Path) -> HookConfig {
    parse_checked(text, home).unwrap_or_default()
}

/// [`parse`], with `None` when the text is not valid TOML.
fn parse_checked(text: &str, home: &Path) -> Option<HookConfig> {
    let doc = text.parse::<toml_edit::DocumentMut>().ok()?;
    let hooks = doc.get("hooks").and_then(|h| h.as_table_like());
    let table = |name: &str| {
        hooks
            .and_then(|h| h.get(name))
            .and_then(|t| t.as_table_like())
    };
    let enabled =
        |t: &dyn toml_edit::TableLike| t.get("enabled").and_then(|e| e.as_bool()) == Some(true);
    let strings = |t: &dyn toml_edit::TableLike, key: &str| -> Option<Vec<String>> {
        match t.get(key) {
            None => Some(Vec::new()),
            Some(item) => item
                .as_array()?
                .iter()
                .map(|v| v.as_str().map(str::to_string))
                .collect(),
        }
    };
    let outside = table("outside_project_writes").and_then(|t| {
        if !enabled(t) {
            return None;
        }
        let allow = strings(t, "allow")?;
        Some(OutsideRule {
            allow: allow.iter().filter_map(|a| absolute(a, home)).collect(),
        })
    });
    let primary = table("primary_checkout").and_then(|t| {
        if !enabled(t) {
            return None;
        }
        Some(PrimaryRule {
            repos: strings(t, "repos")?
                .iter()
                .filter_map(|r| absolute(r, home))
                .map(|p| canonical(&p))
                .collect(),
            allow_commands: strings(t, "allow_commands")?
                .into_iter()
                .map(|c| c.trim().to_string())
                .collect(),
            deny_commits_on: strings(t, "deny_commits_on")?
                .into_iter()
                .map(|b| b.trim().to_string())
                .filter(|b| !b.is_empty())
                .collect(),
        })
    });
    let identity = table("commit_identity").and_then(|t| {
        if !enabled(t) {
            return None;
        }
        Some(IdentityRule {
            deny_emails: strings(t, "deny_emails")?
                .into_iter()
                .map(|e| e.trim().to_ascii_lowercase())
                .filter(|e| !e.is_empty())
                .collect(),
        })
    });
    let (edit, edit_invalid) = match table("edit_diagnostics") {
        Some(t) if enabled(t) => match edit_rule(t) {
            Some(rule) => (Some(rule), false),
            None => (None, true),
        },
        // Present but `enabled` not `true`: off. A non-boolean `enabled` is malformed.
        Some(t) => (
            None,
            t.get("enabled").is_some_and(|e| e.as_bool().is_none()),
        ),
        None => (None, false),
    };
    // A word that is not one of the three modes leaves the rule in observe.
    let modes = table("modes")
        .map(|t| {
            t.iter()
                .filter_map(|(rule, item)| {
                    let mode = RuleMode::parse(item.as_str()?)?;
                    Some((rule.to_string(), mode))
                })
                .collect()
        })
        .unwrap_or_default();
    Some(HookConfig {
        modes,
        outside,
        primary,
        identity,
        edit,
        edit_invalid,
    })
}

impl HookConfig {
    /// The mode of rule `rule`: its `[hooks.modes]` entry, or observe.
    pub fn mode_of(&self, rule: &str) -> RuleMode {
        self.modes.get(rule).copied().unwrap_or_default()
    }
}

/// The keys `[hooks.edit_diagnostics]` takes; any other key makes it malformed.
const EDIT_KEYS: [&str; 6] = [
    "enabled",
    "mode",
    "timeout_secs",
    "max_lines",
    "max_chars",
    "checkers",
];

/// An enabled `[hooks.edit_diagnostics]` table, or `None` when anything in it is malformed:
/// an unknown key, a `mode` other than `context` (blocking is slice S2), a number outside its
/// bounds, a checker key that is not 1-16 lower-case letters or digits, or a checker that is
/// not a non-empty list of non-empty strings without NUL.
fn edit_rule(t: &dyn toml_edit::TableLike) -> Option<EditRule> {
    if t.iter().any(|(k, _)| !EDIT_KEYS.contains(&k)) {
        return None;
    }
    match t.get("mode") {
        None => {}
        Some(m) if m.as_str() == Some("context") => {}
        Some(_) => return None,
    }
    let number = |key: &str, (default, low, high): (u64, u64, u64)| -> Option<u64> {
        match t.get(key) {
            None => Some(default),
            Some(v) => u64::try_from(v.as_integer()?)
                .ok()
                .filter(|n| (low..=high).contains(n)),
        }
    };
    let widen = |(a, b, c): (usize, usize, usize)| (a as u64, b as u64, c as u64);
    let timeout_secs = number("timeout_secs", EDIT_TIMEOUT_SECS)?;
    let max_lines = number("max_lines", widen(EDIT_MAX_LINES))? as usize;
    let max_chars = number("max_chars", widen(EDIT_MAX_CHARS))? as usize;
    let mut checkers = Vec::new();
    if let Some(item) = t.get("checkers") {
        for (ext, argv) in item.as_table_like()?.iter() {
            let ext_ok = (1..=16).contains(&ext.len())
                && ext
                    .bytes()
                    .all(|b| b.is_ascii_lowercase() || b.is_ascii_digit());
            let argv: Vec<String> = argv
                .as_array()?
                .iter()
                .map(|v| v.as_str().map(str::to_string))
                .collect::<Option<_>>()?;
            if !ext_ok || argv.is_empty() || argv.iter().any(|a| a.is_empty() || a.contains('\0')) {
                return None;
            }
            checkers.push((ext.to_string(), argv));
        }
    }
    Some(EditRule {
        timeout_secs,
        max_lines,
        max_chars,
        checkers,
    })
}

/// `~` or `~/…` expanded against `home`; a relative path is ignored (None).
fn absolute(text: &str, home: &Path) -> Option<PathBuf> {
    let text = text.trim();
    let path = if text == "~" {
        home.to_path_buf()
    } else if let Some(rest) = text.strip_prefix("~/") {
        home.join(rest)
    } else {
        PathBuf::from(text)
    };
    path.is_absolute().then(|| normalize(&path))
}

/// `.` dropped and `..` applied lexically (never above the root).
pub fn normalize(path: &Path) -> PathBuf {
    let mut out = PathBuf::new();
    for c in path.components() {
        match c {
            Component::CurDir => {}
            Component::ParentDir => {
                out.pop();
            }
            other => out.push(other.as_os_str()),
        }
    }
    out
}

/// The path with its longest existing prefix resolved through symlinks, then normalized, so a
/// symlink out of a folder and a `..` in a part that does not exist yet are both seen.
pub fn canonical(path: &Path) -> PathBuf {
    let path = normalize(path);
    let mut existing = path.clone();
    let mut tail = Vec::new();
    while fs::symlink_metadata(&existing).is_err() {
        match (
            existing.file_name().map(|n| n.to_os_string()),
            existing.parent(),
        ) {
            (Some(name), Some(parent)) => {
                tail.push(name);
                existing = parent.to_path_buf();
            }
            _ => return path,
        }
    }
    let mut out = fs::canonicalize(&existing).unwrap_or(existing);
    for name in tail.into_iter().rev() {
        out.push(name);
    }
    // The tail was normalized above and canonicalize leaves no `..`.
    out
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn rules_are_off_unless_enabled_and_well_formed() {
        let home = Path::new("/home/u");
        for text in [
            "",
            "not toml [",
            "[hooks]\n",
            "[hooks.outside_project_writes]\nallow = [\"/tmp/\"]\n",
            "[hooks.outside_project_writes]\nenabled = \"yes\"\n",
            "[hooks.outside_project_writes]\nenabled = true\nallow = [1]\n",
            "[hooks.primary_checkout]\nenabled = false\nrepos = [\"/r\"]\n",
            "[hooks.primary_checkout]\nenabled = true\nrepos = \"/r\"\n",
        ] {
            let c = parse(text, home);
            assert_eq!(c, HookConfig::default(), "{text:?}");
        }
    }

    /// Card C-86 (note 2346): `[hooks.modes]` names a mode per rule; a missing entry, a word that
    /// is not off, observe or enforce, and a value that is not a string all leave observe.
    #[test]
    fn rule_modes_default_to_observe() {
        let home = Path::new("/home/u");
        let rule = "merge_confirm";
        for (text, want) in [
            ("", RuleMode::Observe),
            ("[hooks.modes]\n", RuleMode::Observe),
            ("[hooks.modes]\nmerge_confirm = \"off\"\n", RuleMode::Off),
            (
                "[hooks.modes]\nmerge_confirm = \"observe\"\n",
                RuleMode::Observe,
            ),
            (
                "[hooks.modes]\nmerge_confirm = \"enforce\"\n",
                RuleMode::Enforce,
            ),
            (
                "[hooks.modes]\nmerge_confirm = \"Enforce\"\n",
                RuleMode::Observe,
            ),
            (
                "[hooks.modes]\nmerge_confirm = \"yes\"\n",
                RuleMode::Observe,
            ),
            ("[hooks.modes]\nmerge_confirm = 3\n", RuleMode::Observe),
            ("[hooks.modes]\nother = \"enforce\"\n", RuleMode::Observe),
        ] {
            assert_eq!(parse(text, home).mode_of(rule), want, "{text:?}");
        }
        // A bad word for one rule leaves the other rules' words alone.
        let c = parse(
            "[hooks.modes]\nmerge_confirm = \"x\"\nother = \"off\"\n",
            home,
        );
        assert_eq!(
            (c.mode_of(rule), c.mode_of("other")),
            (RuleMode::Observe, RuleMode::Off)
        );
    }

    #[test]
    fn lists_are_expanded_and_normalized() {
        let home = Path::new("/home/u");
        let c = parse(
            "[hooks.outside_project_writes]\nenabled = true\nallow = [\"~/.claude/\", \"~\", \"/tmp/x/../y\", \"relative/dir\"]\n\
             [hooks.primary_checkout]\nenabled = true\nrepos = [\"/no/such/repo/./a\"]\nallow_commands = [\"  git switch main  \"]\n",
            home,
        );
        assert_eq!(
            c.outside.unwrap().allow,
            vec![
                PathBuf::from("/home/u/.claude"),
                PathBuf::from("/home/u"),
                PathBuf::from("/tmp/y")
            ]
        );
        let p = c.primary.unwrap();
        assert_eq!(p.repos, vec![PathBuf::from("/no/such/repo/a")]);
        assert_eq!(p.allow_commands, vec!["git switch main".to_string()]);
        // #2878 E2: absent means none.
        assert!(p.deny_commits_on.is_empty());
    }

    /// #2878 E3: `[hooks.commit_identity]` is off unless enabled; `deny_emails` are trimmed,
    /// lowercased and blanks dropped; a malformed list turns the rule off.
    #[test]
    fn commit_identity_is_opt_in() {
        let home = Path::new("/home/u");
        assert_eq!(parse("[hooks.commit_identity]\n", home).identity, None);
        assert_eq!(
            parse("[hooks.commit_identity]\nenabled = false\n", home).identity,
            None
        );
        let on = parse(
            "[hooks.commit_identity]\nenabled = true\ndeny_emails = [\" CI@Corp.Example \", \"\"]\n",
            home,
        );
        assert_eq!(on.identity.unwrap().deny_emails, ["ci@corp.example"]);
        assert_eq!(
            parse("[hooks.commit_identity]\nenabled = true\n", home)
                .identity
                .unwrap()
                .deny_emails,
            Vec::<String>::new()
        );
        assert_eq!(
            parse(
                "[hooks.commit_identity]\nenabled = true\ndeny_emails = \"x\"\n",
                home
            )
            .identity,
            None
        );
    }

    /// #2878 E2: `deny_commits_on` is a list of branch names, trimmed, blanks dropped; a
    /// non-list or a non-string entry turns the rule off like any malformed list.
    #[test]
    fn deny_commits_on_is_read_as_branch_names() {
        let home = Path::new("/home/u");
        let c = parse(
            "[hooks.primary_checkout]\nenabled = true\nrepos = [\"/r\"]\ndeny_commits_on = [\" main \", \"\", \"release\"]\n",
            home,
        );
        assert_eq!(c.primary.unwrap().deny_commits_on, ["main", "release"]);
        for bad in ["deny_commits_on = \"main\"", "deny_commits_on = [1]"] {
            let text =
                format!("[hooks.primary_checkout]\nenabled = true\nrepos = [\"/r\"]\n{bad}\n");
            assert_eq!(parse(&text, home).primary, None, "{bad}");
        }
    }

    /// N2: the bound is on the bytes actually read, so a file that grows after its size was
    /// seen is still cut off; 65,536 bytes are read whole, 65,537 are refused.
    #[test]
    fn the_read_bound_is_on_the_bytes_read() {
        let at = vec![b'#'; 65_536];
        let over = vec![b'#'; 65_537];
        assert_eq!(
            read_bounded(&at[..]).unwrap().map(|s| s.len()),
            Some(65_536)
        );
        assert_eq!(read_bounded(&over[..]).unwrap(), None);
        assert!(read_bounded(&b"\xff"[..]).is_err(), "not UTF-8");
    }

    /// N2 (pre-audit 6812063964 B2): the read itself stops after `CONFIG_MAX + 1` bytes, so a
    /// file that keeps growing is never read whole. The reader is endless and panics once more
    /// than that bound has been asked of it; the post-read length check alone cannot pass this.
    #[test]
    fn the_read_stops_at_the_bound_on_an_endless_reader() {
        struct Endless {
            served: u64,
        }
        impl Read for Endless {
            fn read(&mut self, buf: &mut [u8]) -> std::io::Result<usize> {
                self.served += buf.len() as u64;
                assert!(
                    self.served <= CONFIG_MAX + 1,
                    "read past the bound: {} bytes asked",
                    self.served
                );
                buf.fill(b'#');
                Ok(buf.len())
            }
        }
        let mut reader = Endless { served: 0 };
        assert_eq!(read_bounded(&mut reader).unwrap(), None);
        assert_eq!(reader.served, CONFIG_MAX + 1);
    }

    /// N2 (pre-audit 6812063964 B1): the file is opened before its type is checked, so a FIFO
    /// must not hold the open (`O_NONBLOCK`); it is named `NotRegularFile` within a bound.
    #[cfg(unix)]
    #[test]
    fn a_fifo_config_returns_at_once_as_not_regular() {
        let dir = tempfile::TempDir::new().unwrap();
        let home = dir.path().to_path_buf();
        let path = home.join("config");
        let c = std::ffi::CString::new(path.as_os_str().as_encoded_bytes()).unwrap();
        // SAFETY: `c` is a valid NUL-terminated path that outlives the call.
        assert_eq!(unsafe { libc::mkfifo(c.as_ptr(), 0o600) }, 0, "mkfifo");
        let (tx, rx) = std::sync::mpsc::channel();
        std::thread::spawn(move || {
            let _ = tx.send(load_path(&path, &home).problem);
        });
        let got = rx
            .recv_timeout(std::time::Duration::from_secs(10))
            .expect("load_path blocked on a FIFO config");
        assert_eq!(got, Some(ConfigProblem::NotRegularFile));
    }

    /// N1/N2: each unusable file has its fixed reason; an absent file has none.
    #[cfg(unix)]
    #[test]
    fn load_path_names_the_problem() {
        let dir = tempfile::TempDir::new().unwrap();
        let home = dir.path();
        let path = home.join("config");
        let problem = |p: &Path| load_path(p, home).problem;
        assert_eq!(problem(&path), None);
        fs::write(&path, "[hooks]\n").unwrap();
        assert_eq!(problem(&path), None);
        fs::write(&path, "not toml [").unwrap();
        assert_eq!(problem(&path), Some(ConfigProblem::InvalidToml));
        fs::write(&path, vec![b'#'; 65_537]).unwrap();
        assert_eq!(problem(&path), Some(ConfigProblem::TooLarge));
        let link = home.join("link");
        std::os::unix::fs::symlink(&path, &link).unwrap();
        assert_eq!(problem(&link), Some(ConfigProblem::NotRegularFile));
        assert_eq!(problem(home), Some(ConfigProblem::NotRegularFile));
    }

    #[test]
    fn normalize_applies_parent_steps_lexically() {
        for (input, want) in [
            ("/repo/../etc/x", "/etc/x"),
            ("/repo/new/../../etc", "/etc"),
            ("/../../x", "/x"),
            ("/a/./b/", "/a/b"),
        ] {
            assert_eq!(normalize(Path::new(input)), PathBuf::from(want), "{input}");
        }
    }
}
