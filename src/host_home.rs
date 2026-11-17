//! The one place that says where a host keeps its user-level folder (card C-79,
//! SPEC-20261117-host-home-resolver). Routing, agents, the output style, the skill copy, the
//! Codex doctor and the user-scope guards all ask here, so they agree on one folder.
//!
//! - Claude Code: `CLAUDE_CONFIG_DIR`, else `~/.claude`. "If you set `CLAUDE_CONFIG_DIR`, every
//!   `~/.claude` path on this page lives under that directory instead"
//!   (https://code.claude.com/docs/en/claude-directory.md, read 2026-11-17T09:33Z): `CLAUDE.md`,
//!   `settings.json`, `agents/`, `skills/` and `output-styles/` move with it.
//! - Codex: `CODEX_HOME`, else `~/.codex`. The global `AGENTS.md` "defaults to `~/.codex`,
//!   unless you set `CODEX_HOME`" (https://learn.chatgpt.com/docs/agent-configuration/agents-md,
//!   read 2026-11-17T09:33Z). Codex's user skills are `~/.agents/skills`, not under
//!   `CODEX_HOME`, so the skill copy for Codex does not ask here.
//! - OMP: `~/.omp`; no relocation variable is documented, so none is read.
//!
//! **Whose variable.** The variables relocate the folders of the account that runs the host,
//! so they count only for that account's own home: the home in the password database (Unix),
//! else [`crate::key_vault::home`] (`HOME`, `USERPROFILE` on Windows). A home that is not it — a temporary HOME of a test, another `--home` — keeps
//! its folders under it, whatever the variables say. So a process started with a temporary
//! HOME but the caller's `CLAUDE_CONFIG_DIR` can never write into the caller's real config
//! folder. `MOAI_ACCOUNT_HOME` names the account home instead (tests set it to their
//! temporary HOME); it can only make the user's own variables count, never move a folder by
//! itself.
//!
//! A variable that counts must be an absolute path: an empty or relative value is refused with
//! a message naming the variable (never its value), because the host and this tool could read
//! a relative path against different working folders. The path is used as given, not
//! canonicalized: like `~/.claude` today, the folder itself must be a real folder or absent
//! where the callers check it, and a link there is refused, never followed. Reads the
//! environment, the password database and the two homes' real paths; writes nothing.

use std::ffi::OsString;
use std::path::{Path, PathBuf};

/// The variable that names the account home instead of the password database.
pub const ACCOUNT_HOME_VAR: &str = "MOAI_ACCOUNT_HOME";

/// A host with a user-level folder.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Host {
    Claude,
    Codex,
    Omp,
}

impl Host {
    /// The variable that relocates the folder, if the host documents one.
    pub fn variable(self) -> Option<&'static str> {
        match self {
            Host::Claude => Some("CLAUDE_CONFIG_DIR"),
            Host::Codex => Some("CODEX_HOME"),
            Host::Omp => None,
        }
    }

    /// The folder under HOME when the variable does not count.
    pub fn default_dir(self) -> &'static str {
        match self {
            Host::Claude => ".claude",
            Host::Codex => ".codex",
            Host::Omp => ".omp",
        }
    }

    /// `claude`, `codex` or `omp`; any other name is no host here.
    pub fn from_name(name: &str) -> Option<Host> {
        match name {
            "claude" => Some(Host::Claude),
            "codex" => Some(Host::Codex),
            "omp" => Some(Host::Omp),
            _ => None,
        }
    }
}

/// The host's variable as a folder: `None` when unset, the path when absolute, a refusal
/// naming the variable when empty or relative. Whose home it belongs to is not asked here;
/// the user-scope guards use it to refuse writing into whatever folder it names.
pub fn variable_dir(
    host: Host,
    get: impl Fn(&str) -> Option<OsString>,
) -> Result<Option<PathBuf>, String> {
    let Some(var) = host.variable() else {
        return Ok(None);
    };
    match get(var) {
        None => Ok(None),
        Some(v) if v.is_empty() => Err(format!(
            "{var} is set but empty; set it to an absolute folder or unset it"
        )),
        Some(v) => {
            let path = PathBuf::from(v);
            if path.is_absolute() {
                Ok(Some(path))
            } else {
                Err(format!(
                    "{var} must be an absolute path; a relative value is refused because the host \
                     and moai could read it against different folders"
                ))
            }
        }
    }
}

/// The host's user-level folder for `home`, from this process's environment.
pub fn resolve(host: Host, home: &Path) -> Result<PathBuf, String> {
    let get = |k: &str| std::env::var_os(k);
    let account = get(ACCOUNT_HOME_VAR)
        .filter(|v| !v.is_empty())
        .map(PathBuf::from)
        .or_else(account_home)
        // Without a password database: the one HOME rule (`HOME`, `USERPROFILE` on Windows).
        .or_else(|| crate::key_vault::home().ok());
    resolve_with(host, home, account.as_deref(), get)
}

/// [`resolve`] with the account home and the environment given (tests pass a map). The
/// variable counts only when `home` is `account` (compared by real path); otherwise, and when
/// it is unset, the folder is the default under `home`.
pub fn resolve_with(
    host: Host,
    home: &Path,
    account: Option<&Path>,
    get: impl Fn(&str) -> Option<OsString>,
) -> Result<PathBuf, String> {
    let default = || home.join(host.default_dir());
    if !account.is_some_and(|a| same_place(a, home)) {
        return Ok(default());
    }
    Ok(variable_dir(host, get)?.unwrap_or_else(default))
}

/// Where a path below a host folder really is: `base` is the folder the link checks start at
/// (it must itself be a real folder or absent) and `rel` the path below it.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Place {
    pub base: PathBuf,
    pub rel: PathBuf,
}

impl Place {
    pub fn path(&self) -> PathBuf {
        self.base.join(&self.rel)
    }
}

/// `rel` (for example `skills/moai`, `agents`, `output-styles/MoAI-Easy.md`) below the host
/// folder of `home`. When the host folder is relocated and the first part of `rel` is a link to
/// a folder there (a dangling link counts), the link is followed only when it resolves to that same folder under `home`'s
/// default host folder (both real paths equal): the place is then that default folder (user
/// decision 2026-11-17, selection "기본 폴더 링크만 허용"). A link to anything else, or to a default
/// folder that is not there, is refused naming the link and what to do. Everything else is the
/// path as it is; the callers check the folders for links as before.
pub fn place(host: Host, home: &Path, rel: &str) -> Result<Place, String> {
    let root = resolve(host, home)?;
    place_in(host, home, &root, rel)
}

/// [`place`] below a host folder `root` already resolved for `home`.
pub fn place_in(host: Host, home: &Path, root: &Path, rel: &str) -> Result<Place, String> {
    let default_root = home.join(host.default_dir());
    let plain = || Place {
        base: root.to_path_buf(),
        rel: PathBuf::from(rel),
    };
    let Some(first) = Path::new(rel).components().next() else {
        return Ok(plain());
    };
    if root == default_root {
        return Ok(plain());
    }
    let link = root.join(first);
    // Only a link to a folder is a projection folder; a linked file (`settings.json`) is left
    // to the callers, which refuse it.
    if !std::fs::symlink_metadata(&link).is_ok_and(|m| m.file_type().is_symlink())
        || std::fs::metadata(&link).is_ok_and(|m| !m.is_dir())
    {
        return Ok(plain());
    }
    let want = default_root.join(first);
    match (std::fs::canonicalize(&link), std::fs::canonicalize(&want)) {
        (Ok(a), Ok(b)) if a == b => Ok(Place {
            base: default_root,
            rel: PathBuf::from(rel),
        }),
        _ => Err(format!(
            "{} is a link to a folder other than {}; make it a real folder, or a link to {}",
            link.display(),
            want.display(),
            want.display()
        )),
    }
}

/// [`resolve`] for a host named `claude`, `codex` or `omp`.
pub fn resolve_named(host: &str, home: &Path) -> Result<PathBuf, String> {
    let h = Host::from_name(host).ok_or_else(|| format!("unknown host {host:?}"))?;
    resolve(h, home)
}

/// Two paths name the same folder: equal real paths, or equal as written when either cannot
/// be resolved.
fn same_place(a: &Path, b: &Path) -> bool {
    match (std::fs::canonicalize(a), std::fs::canonicalize(b)) {
        (Ok(x), Ok(y)) => x == y,
        _ => a == b,
    }
}

/// The account's home from the password database.
#[cfg(unix)]
fn account_home() -> Option<PathBuf> {
    use std::os::unix::ffi::OsStrExt;
    let mut pwd: libc::passwd = unsafe { std::mem::zeroed() };
    let mut buf = vec![0u8; 16 * 1024];
    let mut found: *mut libc::passwd = std::ptr::null_mut();
    // SAFETY: `pwd`, `buf` and `found` outlive the call; `buf.len()` is the buffer's size.
    let rc = unsafe {
        libc::getpwuid_r(
            libc::getuid(),
            &mut pwd,
            buf.as_mut_ptr().cast(),
            buf.len(),
            &mut found,
        )
    };
    if rc != 0 || found.is_null() || pwd.pw_dir.is_null() {
        return None;
    }
    // SAFETY: on success `pw_dir` points at a NUL-terminated string inside `buf`.
    let dir = unsafe { std::ffi::CStr::from_ptr(pwd.pw_dir) };
    let dir = std::ffi::OsStr::from_bytes(dir.to_bytes());
    (!dir.is_empty()).then(|| PathBuf::from(dir))
}

#[cfg(not(unix))]
fn account_home() -> Option<PathBuf> {
    None
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::collections::HashMap;

    fn env(pairs: &[(&str, &str)]) -> impl Fn(&str) -> Option<OsString> {
        let m: HashMap<String, OsString> = pairs
            .iter()
            .map(|(k, v)| (k.to_string(), OsString::from(v)))
            .collect();
        move |k| m.get(k).cloned()
    }

    const H: &str = "/nonexistent-moai-test/h";

    #[test]
    fn unset_variables_mean_the_folder_under_home() {
        let home = Path::new(H);
        for (host, dir) in [
            (Host::Claude, ".claude"),
            (Host::Codex, ".codex"),
            (Host::Omp, ".omp"),
        ] {
            assert_eq!(
                resolve_with(host, home, Some(home), env(&[])).unwrap(),
                home.join(dir)
            );
        }
    }

    #[test]
    fn an_absolute_variable_of_the_account_home_is_the_folder_as_given() {
        let home = Path::new(H);
        let e = env(&[("CLAUDE_CONFIG_DIR", "/c/conf"), ("CODEX_HOME", "/x/codex")]);
        assert_eq!(
            resolve_with(Host::Claude, home, Some(home), &e).unwrap(),
            Path::new("/c/conf")
        );
        assert_eq!(
            resolve_with(Host::Codex, home, Some(home), &e).unwrap(),
            Path::new("/x/codex")
        );
        // Not canonicalized: `..` stays as written.
        let e = env(&[("CODEX_HOME", "/x/../y")]);
        assert_eq!(
            resolve_with(Host::Codex, home, Some(home), e).unwrap(),
            Path::new("/x/../y")
        );
        // OMP reads no variable.
        let e = env(&[("CLAUDE_CONFIG_DIR", "/c"), ("CODEX_HOME", "/x")]);
        assert_eq!(
            resolve_with(Host::Omp, home, Some(home), e).unwrap(),
            home.join(".omp")
        );
    }

    #[test]
    fn another_home_keeps_its_folders_whatever_the_variables_say() {
        let home = Path::new(H);
        let e = env(&[("CLAUDE_CONFIG_DIR", "/c/conf"), ("CODEX_HOME", "rel")]);
        for account in [Some(Path::new("/nonexistent-moai-test/other")), None] {
            assert_eq!(
                resolve_with(Host::Claude, home, account, &e).unwrap(),
                home.join(".claude")
            );
            // Not its variable: a relative value is not even read.
            assert_eq!(
                resolve_with(Host::Codex, home, account, &e).unwrap(),
                home.join(".codex")
            );
        }
    }

    #[test]
    fn an_empty_or_relative_variable_is_refused_by_name_without_its_value() {
        let home = Path::new(H);
        for (host, var) in [
            (Host::Claude, "CLAUDE_CONFIG_DIR"),
            (Host::Codex, "CODEX_HOME"),
        ] {
            for value in ["", "c", "./c", "~/c", "secret-word/x"] {
                let err = resolve_with(host, home, Some(home), env(&[(var, value)])).unwrap_err();
                assert!(err.starts_with(var), "{err}");
                assert!(err.contains("absolute"), "{var}={value:?}: {err}");
                if value.len() > 1 {
                    assert!(!err.contains(value), "the value is never echoed: {err}");
                }
                assert_eq!(variable_dir(host, env(&[(var, value)])).unwrap_err(), err);
            }
        }
        assert_eq!(variable_dir(Host::Claude, env(&[])).unwrap(), None);
        assert_eq!(
            variable_dir(Host::Omp, env(&[("CLAUDE_CONFIG_DIR", "x")])).unwrap(),
            None
        );
    }

    #[test]
    fn host_names_map_to_hosts() {
        assert_eq!(Host::from_name("claude"), Some(Host::Claude));
        assert_eq!(Host::from_name("codex"), Some(Host::Codex));
        assert_eq!(Host::from_name("omp"), Some(Host::Omp));
        assert_eq!(Host::from_name("cursor"), None);
        assert!(resolve_named("cursor", Path::new(H)).is_err());
    }

    #[cfg(unix)]
    #[test]
    fn the_account_home_comes_from_the_password_database() {
        let dir = account_home().expect("this account has a home");
        assert!(dir.is_absolute(), "{}", dir.display());
    }
}
