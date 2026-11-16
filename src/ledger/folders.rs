//! The three folders above a ledger (Issue #3068 hardening, C3 and C4):
//! `<home>/.moai`, `<home>/.moai/projects` and `<home>/.moai/projects/<id>`.
//!
//! - **Owner only (C3).** A folder this code creates gets mode 0700, whatever the umask: it is
//!   made with 0700 and then set to exactly 0700, level by level. `<home>/.moai` is only
//!   created when absent and never changed when it exists (other tools own it). An existing
//!   `projects` or `projects/<id>` that this user owns and that any group or other bit opens is
//!   tightened to 0700 by `register` (and listed by the plan first); one that another user owns
//!   is never touched and `register` refuses it (`folder_not_ours`).
//! - **No links (C4).** A symbolic link at any of those three components, a dangling one too, is
//!   refused (`path_link`) by every ledger call before anything is read or written through it.
//!   The HOME folder and its ancestors may be links (`/tmp` and `/var` on macOS): they are not
//!   examined.
//!
//! The check is `lstat` on each component. For the steps that create or replace things
//! ([`create_chain`] and [`verify_open`]) each existing component is also opened with
//! `O_NOFOLLOW | O_DIRECTORY` and its `fstat` identity (device and inode) compared with a fresh
//! `lstat` of the path, so a component swapped for a link or another folder between the check
//! and the open is noticed. **Residual race:** SQLite opens the database by path, so a swap
//! that happens after the last verification and before SQLite's own `open` (or the `link` that
//! puts the new file in place) is not seen. The folders are 0700, so only the same user (or
//! root) can make that swap.

use super::LedgerError;
use std::path::{Path, PathBuf};

/// The label of a component in an error text and in the plan.
const ROOT: &str = ".moai";
const PROJECTS: &str = "projects";
const PROJECT: &str = "projects/<id>";

/// What [`create_chain`] did.
#[derive(Debug, Clone, Default, PartialEq, Eq)]
pub struct FolderReport {
    /// Folders this call made (mode 0700).
    pub created: Vec<PathBuf>,
    /// Existing folders of this user that were opened to group or other and are now 0700.
    pub tightened: Vec<PathBuf>,
}

/// What would change, read-only (the register plan).
#[derive(Debug, Clone, Default, PartialEq, Eq)]
pub struct FolderPlan {
    /// Existing `projects` / `projects/<id>` of this user with a group or other bit set.
    pub would_tighten: Vec<PathBuf>,
    /// Existing `projects` / `projects/<id>` that another user owns.
    pub not_ours: Vec<PathBuf>,
}

fn components(home: &Path, project_id: &str) -> Result<[(&'static str, PathBuf); 3], LedgerError> {
    let project = super::location::project_dir(home, project_id)?;
    let projects = crate::home_workdir::projects_root(home);
    let root = crate::home_workdir::root(home);
    Ok([(ROOT, root), (PROJECTS, projects), (PROJECT, project)])
}

fn link_error(label: &str, path: &Path) -> LedgerError {
    LedgerError::PathLink(format!("{label} ({})", path.display()))
}

/// Refuses a symbolic link at `.moai`, `projects` or `<id>`. Reads nothing but the link status
/// of the three paths; an absent component is fine.
pub fn refuse_links(home: &Path, project_id: &str) -> Result<(), LedgerError> {
    for (label, path) in components(home, project_id)? {
        match std::fs::symlink_metadata(&path) {
            Ok(m) if m.file_type().is_symlink() => return Err(link_error(label, &path)),
            Ok(_) => {}
            // A parent that is a file (ENOTDIR) or absent is not a link; the caller meets it.
            Err(_) => {}
        }
    }
    Ok(())
}

/// The user id this process runs as.
#[cfg(unix)]
fn my_uid() -> u32 {
    // SAFETY: geteuid has no preconditions and cannot fail.
    unsafe { libc::geteuid() }
}

#[cfg(not(unix))]
fn my_uid() -> u32 {
    0
}

#[cfg(unix)]
fn owned_by(m: &std::fs::Metadata, uid: u32) -> bool {
    use std::os::unix::fs::MetadataExt;
    m.uid() == uid
}

#[cfg(not(unix))]
fn owned_by(_m: &std::fs::Metadata, _uid: u32) -> bool {
    true
}

#[cfg(unix)]
fn mode_of(m: &std::fs::Metadata) -> u32 {
    use std::os::unix::fs::PermissionsExt;
    m.permissions().mode() & 0o777
}

#[cfg(not(unix))]
fn mode_of(_m: &std::fs::Metadata) -> u32 {
    0o700
}

/// What a plan would tighten or refuse. Read-only. A link is an error.
pub fn inspect(home: &Path, project_id: &str) -> Result<FolderPlan, LedgerError> {
    inspect_as(home, project_id, my_uid())
}

fn inspect_as(home: &Path, project_id: &str, uid: u32) -> Result<FolderPlan, LedgerError> {
    refuse_links(home, project_id)?;
    let mut plan = FolderPlan::default();
    for (label, path) in components(home, project_id)? {
        if label == ROOT {
            continue;
        }
        let Ok(m) = std::fs::symlink_metadata(&path) else {
            continue;
        };
        if !m.is_dir() {
            continue;
        }
        if !owned_by(&m, uid) {
            plan.not_ours.push(path);
        } else if mode_of(&m) & 0o077 != 0 {
            plan.would_tighten.push(path);
        }
    }
    Ok(plan)
}

/// The octal modes of `projects` and `projects/<id>` for `ledger-status`: `None` when absent.
/// Reads only.
pub fn modes(
    home: &Path,
    project_id: &str,
) -> Result<(Option<String>, Option<String>), LedgerError> {
    let show = |path: &Path| {
        std::fs::symlink_metadata(path)
            .ok()
            .filter(|m| m.is_dir())
            .map(|m| format!("{:04o}", mode_of(&m)))
    };
    let projects = crate::home_workdir::projects_root(home);
    let project = super::location::project_dir(home, project_id)?;
    Ok((show(&projects), show(&project)))
}

#[cfg(unix)]
fn set_private(path: &Path) -> Result<(), LedgerError> {
    use std::os::unix::fs::PermissionsExt;
    std::fs::set_permissions(path, std::fs::Permissions::from_mode(0o700))
        .map_err(|e| LedgerError::Io(format!("{}: {e}", path.display())))
}

#[cfg(not(unix))]
fn set_private(_path: &Path) -> Result<(), LedgerError> {
    Ok(())
}

fn make_private_dir(path: &Path) -> Result<bool, LedgerError> {
    #[cfg(unix)]
    let builder = {
        use std::os::unix::fs::DirBuilderExt;
        let mut builder = std::fs::DirBuilder::new();
        builder.mode(0o700);
        builder
    };
    #[cfg(not(unix))]
    let builder = std::fs::DirBuilder::new();
    match builder.create(path) {
        Ok(()) => {
            // The mode given to mkdir is cut by the umask; set it to exactly 0700.
            set_private(path)?;
            Ok(true)
        }
        Err(e) if e.kind() == std::io::ErrorKind::AlreadyExists => Ok(false),
        Err(e) => Err(LedgerError::Io(format!("{}: {e}", path.display()))),
    }
}

/// Makes the three folders (each 0700, one level at a time) and tightens an existing
/// `projects` / `<id>` of this user, then verifies the chain ([`verify_open`]). A link or a
/// folder of another user is an error and nothing below it is made.
pub fn create_chain(home: &Path, project_id: &str) -> Result<FolderReport, LedgerError> {
    create_chain_as(home, project_id, my_uid())
}

fn create_chain_as(home: &Path, project_id: &str, uid: u32) -> Result<FolderReport, LedgerError> {
    let mut report = FolderReport::default();
    for (label, path) in components(home, project_id)? {
        match std::fs::symlink_metadata(&path) {
            Ok(m) if m.file_type().is_symlink() => return Err(link_error(label, &path)),
            Ok(m) if !m.is_dir() => {
                return Err(LedgerError::Io(format!(
                    "{} is not a folder",
                    path.display()
                )))
            }
            Ok(m) => {
                if label != ROOT {
                    if !owned_by(&m, uid) {
                        return Err(LedgerError::FolderNotOurs(format!(
                            "{label} ({})",
                            path.display()
                        )));
                    }
                    if mode_of(&m) & 0o077 != 0 {
                        set_private(&path)?;
                        report.tightened.push(path);
                    }
                }
            }
            Err(e) if e.kind() == std::io::ErrorKind::NotFound => {
                if make_private_dir(&path)? {
                    report.created.push(path);
                }
            }
            Err(e) => return Err(LedgerError::Io(format!("{}: {e}", path.display()))),
        }
    }
    verify_open(home, project_id)?;
    Ok(report)
}

/// Tightens the existing folders of this user without creating any (used when the ledger is
/// already there). Same refusals as [`create_chain`].
pub fn tighten_existing(home: &Path, project_id: &str) -> Result<FolderReport, LedgerError> {
    tighten_existing_as(home, project_id, my_uid())
}

fn tighten_existing_as(
    home: &Path,
    project_id: &str,
    uid: u32,
) -> Result<FolderReport, LedgerError> {
    refuse_links(home, project_id)?;
    let mut report = FolderReport::default();
    for (label, path) in components(home, project_id)? {
        if label == ROOT {
            continue;
        }
        let Ok(m) = std::fs::symlink_metadata(&path) else {
            continue;
        };
        if !m.is_dir() {
            continue;
        }
        if !owned_by(&m, uid) {
            return Err(LedgerError::FolderNotOurs(format!(
                "{label} ({})",
                path.display()
            )));
        }
        if mode_of(&m) & 0o077 != 0 {
            set_private(&path)?;
            report.tightened.push(path);
        }
    }
    Ok(report)
}

/// Opens every existing component with `O_NOFOLLOW | O_DIRECTORY` and checks its identity
/// against a fresh `lstat` of the path: a link, or a folder swapped since the check, is
/// `path_link`. Call it right before the step that uses the path (the residual race is in the
/// module comment).
#[cfg(unix)]
pub fn verify_open(home: &Path, project_id: &str) -> Result<(), LedgerError> {
    use std::os::unix::fs::{MetadataExt, OpenOptionsExt};
    for (label, path) in components(home, project_id)? {
        let before = match std::fs::symlink_metadata(&path) {
            Ok(m) if m.file_type().is_symlink() => return Err(link_error(label, &path)),
            Ok(m) => m,
            Err(_) => continue,
        };
        let opened = std::fs::OpenOptions::new()
            .read(true)
            .custom_flags(libc::O_NOFOLLOW | libc::O_DIRECTORY | libc::O_CLOEXEC)
            .open(&path)
            .map_err(|e| match e.raw_os_error() {
                // A link (ELOOP with O_NOFOLLOW) or something that is no longer a folder.
                Some(libc::ELOOP) | Some(libc::ENOTDIR) => link_error(label, &path),
                _ => LedgerError::Io(format!("{}: {e}", path.display())),
            })?;
        let inside = opened
            .metadata()
            .map_err(|e| LedgerError::Io(e.to_string()))?;
        let after = std::fs::symlink_metadata(&path).map_err(|_| link_error(label, &path))?;
        let same =
            |a: &std::fs::Metadata, b: &std::fs::Metadata| a.dev() == b.dev() && a.ino() == b.ino();
        if after.file_type().is_symlink() || !same(&before, &inside) || !same(&inside, &after) {
            return Err(link_error(label, &path));
        }
    }
    Ok(())
}

#[cfg(not(unix))]
pub fn verify_open(home: &Path, project_id: &str) -> Result<(), LedgerError> {
    refuse_links(home, project_id)
}

#[cfg(all(test, unix))]
mod tests {
    use super::*;
    use std::os::unix::fs::PermissionsExt;

    fn mode(path: &Path) -> u32 {
        std::fs::metadata(path).unwrap().permissions().mode() & 0o777
    }

    fn loose_tree(home: &Path) -> (PathBuf, PathBuf) {
        let projects = home.join(".moai/projects");
        let project = projects.join("demo");
        std::fs::create_dir_all(&project).unwrap();
        for p in [&projects, &project] {
            std::fs::set_permissions(p, std::fs::Permissions::from_mode(0o755)).unwrap();
        }
        (projects, project)
    }

    /// A folder another user owns is never changed and never registered into (the user id is
    /// injected: a real foreign owner needs root to set up).
    #[test]
    fn a_folder_of_another_user_is_reported_and_never_chmodded() {
        let dir = tempfile::tempdir().unwrap();
        let (projects, project) = loose_tree(dir.path());
        let stranger = my_uid() + 1;
        let root_mode = mode(&dir.path().join(".moai"));
        let plan = inspect_as(dir.path(), "demo", stranger).unwrap();
        assert_eq!(plan.not_ours, vec![projects.clone(), project.clone()]);
        assert!(plan.would_tighten.is_empty());
        for result in [
            create_chain_as(dir.path(), "demo", stranger).map(|_| ()),
            tighten_existing_as(dir.path(), "demo", stranger).map(|_| ()),
        ] {
            assert!(
                matches!(result, Err(LedgerError::FolderNotOurs(_))),
                "{result:?}"
            );
        }
        assert_eq!(
            (mode(&projects), mode(&project)),
            (0o755, 0o755),
            "no chmod"
        );
        // The same tree, seen by its owner, is tightened.
        let plan = inspect_as(dir.path(), "demo", my_uid()).unwrap();
        assert_eq!(plan.would_tighten, vec![projects.clone(), project.clone()]);
        let report = create_chain_as(dir.path(), "demo", my_uid()).unwrap();
        assert_eq!(report.tightened, vec![projects.clone(), project.clone()]);
        assert_eq!((mode(&projects), mode(&project)), (0o700, 0o700));
        assert_eq!(
            mode(&dir.path().join(".moai")),
            root_mode,
            ".moai is not ours to change"
        );
    }

    /// The identity check notices a component that was swapped for a link after it was made.
    #[test]
    fn a_component_swapped_for_a_link_after_creation_is_noticed() {
        let dir = tempfile::tempdir().unwrap();
        let report = create_chain(dir.path(), "demo").unwrap();
        assert_eq!(report.created.len(), 3);
        verify_open(dir.path(), "demo").unwrap();
        let project = dir.path().join(".moai/projects/demo");
        let elsewhere = dir.path().join("elsewhere");
        std::fs::create_dir(&elsewhere).unwrap();
        std::fs::remove_dir(&project).unwrap();
        std::os::unix::fs::symlink(&elsewhere, &project).unwrap();
        assert!(matches!(
            verify_open(dir.path(), "demo"),
            Err(LedgerError::PathLink(_))
        ));
        assert!(matches!(
            refuse_links(dir.path(), "demo"),
            Err(LedgerError::PathLink(_))
        ));
    }
}
