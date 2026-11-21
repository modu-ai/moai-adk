//! The MoAI work folders under the user's home (#2811 W2, user instruction 2026-11-13; leader
//! decisions D1–D8 in the Issue body and its latest leader comment).
//!
//! Everything lives under the HOME state folder `<HOME>/.moai` — no visible `~/moai` folder:
//! - `<HOME>/.moai/worktrees/<project>-<id>/` — where worktrees made before the worktree-place
//!   decision live; a new worktree is created in the project's `.claude/worktrees/<name>/`
//!   instead;
//! - `<HOME>/.moai/projects/` — a reserved name ([`projects_root`]); nothing creates projects
//!   there since `moai init` was removed (#2935 L3);
//! - `<HOME>/.moai/<service>/` — the default folder of a MoAI service such as `moai-stock`.
//!
//! `<HOME>` is `$HOME` on Unix and `%USERPROFILE%` on Windows, with the same folder name on
//! every OS (D1; no XDG or AppData, no new environment variable). The Go generation already
//! keeps its worktrees in `<HOME>/.moai/worktrees/projects/<project>/<name>` (D8, measured by
//! the leader): that folder is recognised, never moved or deleted, so `projects` is a reserved
//! name under `worktrees/` and never a Rust worktree name (D8b). Worktrees made before this
//! change stay where they are (`<primary>/.moai-cli/worktrees/*`, `.claude/worktrees/*`, D3);
//! [`location_of`] tells all of them apart. Nothing here creates a folder: the callers create
//! `worktrees/` or `projects/` when they first need it (D5).

use std::ffi::OsString;
use std::path::{Path, PathBuf};

type Result<T> = std::result::Result<T, String>;

/// The HOME state folder every OS uses (D1).
pub const STATE_DIR: &str = ".moai";
/// New worktrees, under [`STATE_DIR`].
pub const WORKTREES: &str = "worktrees";
/// The default parent of new projects, under [`STATE_DIR`]; also the Go generation's folder
/// under [`WORKTREES`] (D8).
pub const PROJECTS: &str = "projects";
/// Reports, under [`STATE_DIR`]: `<HOME>/.moai/reports/<project>/` (C-126, user instruction
/// 2026-11-19). A report is a local record that may hold private text, so it is never saved inside
/// the repository.
pub const REPORTS: &str = "reports";
/// Names that are never a worktree folder or a service folder: the work folders themselves.
const RESERVED: &[&str] = &[WORKTREES, PROJECTS, REPORTS];

/// The user's home from the given environment values (D1): `HOME` on Unix, `USERPROFILE` on
/// Windows. An empty value counts as unset.
pub fn home_from(
    home: Option<OsString>,
    userprofile: Option<OsString>,
    windows: bool,
) -> Result<PathBuf> {
    let (value, name) = if windows {
        (userprofile, "USERPROFILE")
    } else {
        (home, "HOME")
    };
    value
        .filter(|v| !v.is_empty())
        .map(PathBuf::from)
        .ok_or_else(|| format!("home directory missing: {name} is not set"))
}

/// The user's home from this process's environment (D1).
pub fn home() -> Result<PathBuf> {
    home_from(
        std::env::var_os("HOME"),
        std::env::var_os("USERPROFILE"),
        cfg!(windows),
    )
}

/// `<HOME>/.moai`.
pub fn root(home: &Path) -> PathBuf {
    home.join(STATE_DIR)
}

/// `<HOME>/.moai/worktrees`.
pub fn worktrees_root(home: &Path) -> PathBuf {
    root(home).join(WORKTREES)
}

/// `<HOME>/.moai/projects`.
pub fn projects_root(home: &Path) -> PathBuf {
    root(home).join(PROJECTS)
}

/// The Go generation's worktree folder `<HOME>/.moai/worktrees/projects` (D8): recognised only.
pub fn go_worktrees_root(home: &Path) -> PathBuf {
    worktrees_root(home).join(PROJECTS)
}

/// `<HOME>/.moai/reports`.
pub fn reports_root(home: &Path) -> PathBuf {
    root(home).join(REPORTS)
}

/// `<HOME>/.moai/reports/<project>`: where a project's reports are saved (C-126). `<project>`
/// is [`project_name`], the part `worktree_name` also uses. Creates nothing.
pub fn reports_dir(home: &Path, project_root: &Path) -> Result<PathBuf> {
    Ok(reports_root(home).join(project_name(project_root)?))
}

/// A name made safe for one path component (D2): every character other than ASCII letters,
/// digits, `.`, `_` and `-` becomes `-`, runs of `-` collapse, and `-` and `.` are trimmed at
/// both ends. `None` when nothing is left.
pub fn safe_name(raw: &str) -> Option<String> {
    let mut out = String::new();
    for c in raw.chars() {
        let c = if c.is_ascii_alphanumeric() || c == '.' || c == '_' || c == '-' {
            c
        } else {
            '-'
        };
        if c == '-' && out.ends_with('-') {
            continue;
        }
        out.push(c);
    }
    let out = out.trim_matches(|c| c == '-' || c == '.').to_string();
    (!out.is_empty()).then_some(out)
}

fn reserved(name: &str) -> bool {
    RESERVED.iter().any(|r| r.eq_ignore_ascii_case(name))
}

/// `<HOME>/.moai/<service>` for a MoAI service such as `moai-stock`. The name must already be
/// a safe component and not a reserved work folder name.
pub fn service_dir(home: &Path, service: &str) -> Result<PathBuf> {
    if safe_name(service).as_deref() != Some(service) || reserved(service) {
        return Err(format!(
            "invalid service name {service:?}: [A-Za-z0-9._-], not worktrees, projects or reports"
        ));
    }
    Ok(root(home).join(service))
}

/// The project's name as one path component: the safe form of the repository's top-level folder
/// name (D2). The worktree folder name (`worktree_name`) and the reports folder (`reports_dir`)
/// both start from it, so a project has one name under the MoAI home.
pub fn project_name(project_root: &Path) -> Result<String> {
    project_root
        .file_name()
        .and_then(|n| n.to_str())
        .and_then(safe_name)
        .ok_or_else(|| "project folder name has no usable characters".to_string())
}

/// The folder name of a new worktree (D2): `<project>-<id>`, where `<project>` is the safe
/// form of the repository's top-level folder name and `<id>` the existing worktree naming
/// (`<issue>-<logical_key>`), also made safe. Never a reserved name (D8b).
pub fn worktree_name(project_root: &Path, id: &str) -> Result<String> {
    let project = project_name(project_root)?;
    let id = safe_name(id).ok_or("worktree id has no usable characters")?;
    let name = format!("{project}-{id}");
    if reserved(&name) {
        return Err(format!("worktree name {name} is reserved"));
    }
    Ok(name)
}

/// The folder of a new worktree: `<project_root>/.claude/worktrees/<name>` — the one place
/// a new worktree is created (the worktree-place decision). The old places keep their names
/// in [`location_of`] and [`location_in`], and nothing here creates the folder: the callers
/// create it when they first need it.
pub fn worktree_dir(project_root: &Path, id: &str) -> Result<PathBuf> {
    Ok(project_root
        .join(".claude")
        .join("worktrees")
        .join(worktree_name(project_root, id)?))
}

/// Makes sure the repository's local exclude file lists the project's worktree folder, so a
/// worktree made there does not show in the main tree's status (REQ-10). Placeholder: the
/// body comes with the change that makes the tests pass.
pub fn ensure_place_ignored(_project: &Path) -> Result<()> {
    Ok(())
}

/// Where a worktree path sits (D3, D8): `home` (`<HOME>/.moai/worktrees/<name>`), `go`
/// (`<HOME>/.moai/worktrees/projects/<project>/<name>`), `project` (`<primary>/.moai-cli/
/// worktrees/<name>`), `claude` (`<primary>/.claude/worktrees/<name>`) or `other`. Compares
/// path components as given; the caller passes canonical paths when links may be involved.
pub fn location_of(path: &Path, home: &Path, primary: &Path) -> &'static str {
    location_in(
        path,
        &[worktrees_root(home)],
        &[go_worktrees_root(home)],
        primary,
    )
}

/// [`location_of`] against given spellings of the work folders: each root may be listed both
/// as written and resolved (`fs::canonicalize`), so a `~/.moai` or `worktrees/projects` that
/// is a symbolic link still names its worktrees `home` or `go` (#2557 lesson).
pub fn location_in(
    path: &Path,
    worktrees_roots: &[PathBuf],
    go_roots: &[PathBuf],
    primary: &Path,
) -> &'static str {
    let under = |base: &Path, depth: usize| {
        path.strip_prefix(base)
            .is_ok_and(|rest| rest.components().count() == depth)
    };
    if go_roots.iter().any(|r| under(r, 2)) {
        "go"
    } else if worktrees_roots.iter().any(|r| under(r, 1))
        && path
            .file_name()
            .and_then(|n| n.to_str())
            .is_some_and(|n| !reserved(n))
    {
        "home"
    } else if under(&primary.join(crate::WORKTREES_DIR), 1) {
        "project"
    } else if under(&primary.join(".claude").join("worktrees"), 1) {
        "claude"
    } else {
        "other"
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn os(s: &str) -> Option<OsString> {
        Some(OsString::from(s))
    }

    #[test]
    fn home_is_home_on_unix_and_userprofile_on_windows() {
        // Linux and macOS read HOME and ignore USERPROFILE.
        assert_eq!(
            home_from(os("/home/u"), os("C:\\Users\\u"), false).unwrap(),
            PathBuf::from("/home/u")
        );
        assert_eq!(
            home_from(os("/Users/u"), None, false).unwrap(),
            PathBuf::from("/Users/u")
        );
        // Windows reads USERPROFILE and ignores HOME (D1).
        assert_eq!(
            home_from(os("/c/msys/home/u"), os("C:\\Users\\u"), true).unwrap(),
            PathBuf::from("C:\\Users\\u")
        );
        // Missing or empty is refused, naming the variable.
        assert_eq!(
            home_from(None, os("C:\\Users\\u"), false).unwrap_err(),
            "home directory missing: HOME is not set"
        );
        assert_eq!(
            home_from(os("/home/u"), os(""), true).unwrap_err(),
            "home directory missing: USERPROFILE is not set"
        );
        assert!(home_from(os(""), None, false).is_err());
    }

    #[test]
    fn work_folders_sit_under_the_home_state_folder_on_every_os() {
        for home in ["/home/u", "/Users/u", "C:\\Users\\u"] {
            let h = PathBuf::from(home);
            assert_eq!(root(&h), h.join(".moai"));
            assert_eq!(worktrees_root(&h), h.join(".moai").join("worktrees"));
            assert_eq!(projects_root(&h), h.join(".moai").join("projects"));
            assert_eq!(
                go_worktrees_root(&h),
                h.join(".moai").join("worktrees").join("projects")
            );
            assert_eq!(
                service_dir(&h, "moai-stock").unwrap(),
                h.join(".moai").join("moai-stock")
            );
        }
    }

    #[test]
    fn worktree_dir_is_in_the_project_place_with_the_kept_name() {
        assert_eq!(
            worktree_dir(Path::new("/src/moai-adk"), "2811-home-workdir").unwrap(),
            PathBuf::from("/src/moai-adk/.claude/worktrees/moai-adk-2811-home-workdir")
        );
        assert_eq!(
            worktree_dir(Path::new("/src/My Repo (copy)"), "12-a b").unwrap(),
            PathBuf::from("/src/My Repo (copy)/.claude/worktrees/My-Repo-copy-12-a-b")
        );
    }

    #[test]
    fn worktree_names_are_project_dash_id_and_safe() {
        assert_eq!(
            worktree_dir(Path::new("/src/moai-adk"), "2811-home-workdir").unwrap(),
            PathBuf::from("/src/moai-adk/.claude/worktrees/moai-adk-2811-home-workdir")
        );
        assert_eq!(
            worktree_name(Path::new("/src/My Repo (copy)"), "12-a b").unwrap(),
            "My-Repo-copy-12-a-b"
        );
        assert_eq!(
            worktree_name(Path::new("/src/한글repo"), "1-k").unwrap(),
            "repo-1-k"
        );
        assert!(worktree_name(Path::new("/src/한글"), "1-k").is_err());
        assert!(worktree_name(Path::new("/src/repo"), "..").is_err());
        assert!(worktree_name(Path::new("/"), "1-k").is_err());
        // A repository called `projects` still yields `projects-<id>`, never `projects`.
        assert_eq!(
            worktree_name(Path::new("/src/projects"), "1-k").unwrap(),
            "projects-1-k"
        );
    }

    #[test]
    fn reports_dir_is_home_reports_project_and_safe() {
        let h = Path::new("/home/u");
        assert_eq!(reports_root(h), PathBuf::from("/home/u/.moai/reports"));
        assert_eq!(
            reports_dir(h, Path::new("/src/moai-adk")).unwrap(),
            PathBuf::from("/home/u/.moai/reports/moai-adk")
        );
        assert_eq!(
            reports_dir(h, Path::new("/src/My Repo (copy)")).unwrap(),
            PathBuf::from("/home/u/.moai/reports/My-Repo-copy")
        );
        // The project part is the one `worktree_name` uses, so a project has one name.
        for root in ["/src/moai-adk", "/src/My Repo (copy)", "/src/한글repo"] {
            let root = Path::new(root);
            let project = project_name(root).unwrap();
            assert!(worktree_name(root, "1-k")
                .unwrap()
                .starts_with(&format!("{project}-")));
            assert_eq!(
                reports_dir(h, root).unwrap(),
                reports_root(h).join(&project)
            );
        }
        // No usable character in the folder name is refused with `worktree_name`'s message.
        assert_eq!(
            reports_dir(h, Path::new("/src/한글")).unwrap_err(),
            worktree_name(Path::new("/src/한글"), "1-k").unwrap_err()
        );
        assert!(reports_dir(h, Path::new("/")).is_err());
    }

    #[test]
    fn reports_dir_creates_nothing_and_a_service_cannot_take_the_name() {
        let tmp = tempfile::TempDir::new().unwrap();
        let before = std::fs::read_dir(tmp.path()).unwrap().count();
        assert!(reports_dir(tmp.path(), Path::new("/src/한글")).is_err());
        assert!(reports_dir(tmp.path(), Path::new("/src/repo")).is_ok());
        assert_eq!(std::fs::read_dir(tmp.path()).unwrap().count(), before);
        // The reports folder is a work folder name now: no service may be called `reports`.
        assert!(service_dir(tmp.path(), "reports").is_err());
        assert!(service_dir(tmp.path(), "Reports").is_err());
        assert!(service_dir(tmp.path(), "moai-stock").is_ok());
    }

    #[test]
    fn safe_name_edges() {
        assert_eq!(safe_name("a//b"), Some("a-b".into()));
        assert_eq!(safe_name("-.a.-"), Some("a".into()));
        assert_eq!(safe_name("a_b.c-d"), Some("a_b.c-d".into()));
        assert_eq!(safe_name("..."), None);
        assert_eq!(safe_name(""), None);
    }

    #[test]
    fn reserved_and_unsafe_service_names_are_refused() {
        let h = Path::new("/home/u");
        for bad in [
            "projects",
            "Projects",
            "worktrees",
            "",
            "a/b",
            "..",
            "-x",
            "a b",
        ] {
            assert!(service_dir(h, bad).is_err(), "{bad:?}");
        }
    }

    #[test]
    fn locations_tell_new_old_go_and_claude_worktrees_apart() {
        let h = Path::new("/home/u");
        let p = Path::new("/src/repo");
        let loc = |s: &str| location_of(Path::new(s), h, p);
        assert_eq!(loc("/home/u/.moai/worktrees/repo-1-k"), "home");
        assert_eq!(loc("/home/u/.moai/worktrees/projects/repo/feat"), "go");
        // The reserved Go folder itself, and anything deeper or shallower, is not a worktree.
        assert_eq!(loc("/home/u/.moai/worktrees/projects"), "other");
        assert_eq!(loc("/home/u/.moai/worktrees/projects/repo"), "other");
        assert_eq!(loc("/home/u/.moai/worktrees/projects/repo/feat/x"), "other");
        assert_eq!(loc("/home/u/.moai/worktrees/repo-1-k/sub"), "other");
        assert_eq!(loc("/home/u/.moai/worktrees"), "other");
        assert_eq!(
            loc(&format!("/src/repo/{}/2811-x", crate::WORKTREES_DIR)),
            "project"
        );
        assert_eq!(loc("/src/repo/.claude/worktrees/agent-1"), "claude");
        assert_eq!(loc("/elsewhere/wt"), "other");
        assert_eq!(loc("/home/u/moai/worktrees/repo-1-k"), "other");
    }
}
