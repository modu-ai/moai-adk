//! Opt-in hook rules from `~/.moai-cli/config` (Issue #2391 stage 7; user decision, Issue #2391
//! comment 6779829826, design 6777574227). Both rules are off unless the user turns them on:
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
//! ```
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
}

#[derive(Clone, Debug, Default, PartialEq, Eq)]
pub struct HookConfig {
    pub outside: Option<OutsideRule>,
    pub primary: Option<PrimaryRule>,
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
}

impl ConfigProblem {
    /// The fixed reason word written to the engine log.
    pub fn reason(self) -> &'static str {
        match self {
            ConfigProblem::NotRegularFile => "not_regular_file",
            ConfigProblem::TooLarge => "too_large",
            ConfigProblem::Unreadable => "unreadable",
            ConfigProblem::InvalidToml => "invalid_toml",
        }
    }
}

/// A loaded config and, when a present file was unusable, why.
#[derive(Clone, Debug, Default, PartialEq, Eq)]
pub struct Loaded {
    pub config: HookConfig,
    pub problem: Option<ConfigProblem>,
}

/// The config of `HOME` (both rules off when `HOME` is unset or the file is unusable).
pub fn load() -> Loaded {
    let Some(home) = std::env::var_os("HOME").filter(|h| !h.is_empty()) else {
        return Loaded::default();
    };
    let home = PathBuf::from(home);
    let path = crate::home_root::path(&home).join("config");
    load_path(&path, &home)
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
        })
    });
    Some(HookConfig { outside, primary })
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
