//! A private temporary folder for the `test` step of `check affected`
//! (SPEC-20261124-home-guard-flake revision 4).
//!
//! A hand-picked private `TMPDIR` kept failing for one of three reasons: it lay inside the
//! worktree or a git repository, or its path held a session id. This module makes the choice by
//! rule: a root that meets the three conditions, one fresh folder under it that this process
//! created, and a removal that can reach that folder and nothing else. No shell and no `rm` is
//! involved; the one `remove_dir_all` is called with the recorded canonical path after it has
//! been checked again.

use serde_json::{json, Value};
use std::ffi::OsString;
use std::path::{Path, PathBuf};
use std::process::Command;

/// How many names are tried before the folder is given up on.
const ATTEMPTS: usize = 4;

/// The folder this run created for its test step, as recorded when it was made.
#[derive(Debug)]
pub struct Private {
    /// The canonical path of the folder.
    pub path: PathBuf,
    /// The canonical path of the root it was made under.
    root: PathBuf,
}

/// The temporary roots to try, in order, each with the label that names it in a reason: the
/// caller's `TMPDIR`, the macOS user temporary folder and `/tmp`.
pub fn candidates(caller: Option<OsString>) -> Vec<(&'static str, PathBuf)> {
    let mut found = Vec::new();
    if let Some(c) = caller.filter(|c| !c.is_empty()) {
        found.push(("the caller's TMPDIR", PathBuf::from(c)));
    }
    if cfg!(target_os = "macos") {
        if let Ok(out) = Command::new("getconf").arg("DARWIN_USER_TEMP_DIR").output() {
            let text = String::from_utf8_lossy(&out.stdout).trim().to_string();
            if out.status.success() && !text.is_empty() {
                found.push(("the macOS user temporary folder", PathBuf::from(text)));
            }
        }
    }
    if cfg!(unix) {
        found.push(("/tmp", PathBuf::from("/tmp")));
    } else {
        found.push(("the system temporary folder", std::env::temp_dir()));
    }
    found
}

/// Whether `path` holds a session id shape (hex groups 8-4-4-4-12), the shape the repository
/// scans and `bundle` refuse in a text that leaves the repository.
fn session_shaped(path: &Path) -> bool {
    crate::check_tree::uuid_hits(path.as_os_str().as_encoded_bytes()) > 0
}

/// The first folder above `path` (or `path` itself) that holds a `.git` entry.
fn git_root(path: &Path) -> Option<PathBuf> {
    path.ancestors()
        .find(|a| a.join(".git").symlink_metadata().is_ok())
        .map(Path::to_path_buf)
}

/// The canonical `root` when it meets the three conditions of REQ-14, else the reason it does
/// not: outside the worktree `work` and the repository it belongs to, inside no git repository
/// at all, and with no session id shape in its path.
pub fn fits(root: &Path, work: &Path) -> Result<PathBuf, String> {
    let canon = root
        .canonicalize()
        .map_err(|e| format!("it cannot be resolved ({})", e.kind()))?;
    if !canon.is_dir() {
        return Err("it is not a folder".to_string());
    }
    let work = work.canonicalize().unwrap_or_else(|_| work.to_path_buf());
    if canon.starts_with(&work) {
        return Err("it lies inside the worktree".to_string());
    }
    if let Some(repo) = git_root(&work) {
        if canon.starts_with(&repo) {
            return Err("it lies inside the repository of the worktree".to_string());
        }
    }
    if git_root(&canon).is_some() {
        return Err("it lies inside a git repository".to_string());
    }
    if session_shaped(root) || session_shaped(&canon) {
        return Err("its path holds a session id shape".to_string());
    }
    Ok(canon)
}

/// `mkdir` with mode 0700 that fails when the name exists, whatever it is (a folder, a file or a
/// link, a dangling one too).
fn make_dir(path: &Path) -> std::io::Result<()> {
    let mut b = std::fs::DirBuilder::new();
    #[cfg(unix)]
    {
        use std::os::unix::fs::DirBuilderExt;
        b.mode(0o700);
    }
    b.create(path)
}

/// Create one fresh folder under the canonical `root`, trying `names` in turn. A name that is
/// taken is skipped and what carries it is never opened, removed or followed.
fn create_in(root: &Path, names: &mut dyn Iterator<Item = String>) -> Result<Private, String> {
    for name in names.take(ATTEMPTS) {
        let path = root.join(&name);
        match make_dir(&path) {
            Ok(()) => {
                let canon = match path.canonicalize() {
                    Ok(c) if c == path && c.parent() == Some(root) => c,
                    _ => {
                        // Only the folder just made is empty and removable by `remove_dir`;
                        // a link or a file at the name makes this call fail and nothing else.
                        let _ = std::fs::remove_dir(&path);
                        return Err("the new folder did not resolve to its own path".to_string());
                    }
                };
                return Ok(Private {
                    path: canon,
                    root: root.to_path_buf(),
                });
            }
            Err(e) if e.kind() == std::io::ErrorKind::AlreadyExists => continue,
            Err(e) => return Err(format!("the folder cannot be created ({})", e.kind())),
        }
    }
    Err("every name tried was taken".to_string())
}

/// Names that no one can predict and that carry no session id shape: the process id and the
/// clock in nanoseconds, plus the attempt.
fn fresh_names() -> impl Iterator<Item = String> {
    let pid = std::process::id();
    let nanos = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map_or(0, |d| d.as_nanos());
    (0u128..).map(move |n| format!("moai-affected-{pid}-{}", nanos + n))
}

/// Make the private folder under the first candidate that fits and can hold it. `Err` carries
/// one reason per candidate.
pub fn prepare(work: &Path, from: &[(&'static str, PathBuf)]) -> Result<Private, String> {
    let mut why = Vec::new();
    for (label, candidate) in from {
        match fits(candidate, work) {
            Ok(root) => match create_in(&root, &mut fresh_names()) {
                Ok(private) => return Ok(private),
                Err(e) => why.push(format!("{label}: {e}")),
            },
            Err(e) => why.push(format!("{label}: {e}")),
        }
    }
    Err(if why.is_empty() {
        "no temporary root to try".to_string()
    } else {
        why.join("; ")
    })
}

/// Remove the folder this run created, and nothing else: it must still be a folder (no link)
/// whose canonical path is the one recorded and whose parent is the recorded root. Anything
/// else is left alone and named in the `Err`.
pub fn remove_own(p: &Private) -> Result<(), String> {
    match p.path.symlink_metadata() {
        Ok(m) if m.is_dir() && !m.file_type().is_symlink() => {}
        Ok(_) => return Err("it is no longer a folder".to_string()),
        Err(e) => return Err(format!("it cannot be read any more ({})", e.kind())),
    }
    match p.path.canonicalize() {
        Ok(c) if c == p.path => {}
        _ => return Err("its path no longer resolves to itself".to_string()),
    }
    if p.path.parent() != Some(p.root.as_path()) {
        return Err("its parent is not the root it was made under".to_string());
    }
    std::fs::remove_dir_all(&p.path).map_err(|e| format!("it could not be removed ({})", e.kind()))
}

/// What the answer's `test_tmpdir` says once the test step has exited with `code`: a passed run
/// removes the folder, a failed one keeps it as the evidence, and a folder that no longer fits is
/// left with the reason.
pub fn finish(p: Private, code: i32) -> Value {
    if code != 0 {
        return json!({
            "state": "kept",
            "path": p.path.to_string_lossy(),
            "reason": "the test step did not pass; the folder holds what its tests wrote",
        });
    }
    match remove_own(&p) {
        Ok(()) => json!({"state": "removed"}),
        Err(reason) => json!({"state": "left", "path": p.path.to_string_lossy(), "reason": reason}),
    }
}

/// The answer's `test_tmpdir` when no private folder could be made: the step keeps the caller's
/// `TMPDIR` and the reason says why.
pub fn inherited(reason: &str) -> Value {
    json!({"state": "inherited", "reason": reason})
}

/// The answer's `test_tmpdir` when the test step did not run.
pub fn not_run() -> Value {
    json!({"state": "not_run"})
}

#[cfg(test)]
fn scratch() -> tempfile::TempDir {
    tempfile::tempdir().unwrap()
}

/// A folder that is outside everything fits and comes back canonical.
#[test]
fn a_folder_outside_the_worktree_and_every_repository_fits() {
    let (root, work) = (scratch(), scratch());
    let got = fits(root.path(), work.path()).unwrap();
    assert_eq!(got, root.path().canonicalize().unwrap());
}

#[test]
fn a_session_id_shape_in_the_path_is_refused() {
    let (root, work) = (scratch(), scratch());
    let shaped = root.path().join("11111111-2222-3333-4444-555555555555");
    std::fs::create_dir(&shaped).unwrap();
    let why = fits(&shaped, work.path()).unwrap_err();
    assert!(why.contains("session id"), "{why}");
}

#[test]
fn a_git_ancestor_is_refused() {
    let (root, work) = (scratch(), scratch());
    std::fs::create_dir(root.path().join(".git")).unwrap();
    let inner = root.path().join("t");
    std::fs::create_dir(&inner).unwrap();
    let why = fits(&inner, work.path()).unwrap_err();
    assert!(why.contains("git repository"), "{why}");
    // A `.git` file (a linked worktree) counts the same.
    let (root, work) = (scratch(), scratch());
    std::fs::write(root.path().join(".git"), "gitdir: elsewhere\n").unwrap();
    let why = fits(root.path(), work.path()).unwrap_err();
    assert!(why.contains("git repository"), "{why}");
}

#[test]
fn the_worktree_and_what_lies_inside_it_are_refused() {
    let work = scratch();
    std::fs::write(work.path().join(".git"), "gitdir: elsewhere\n").unwrap();
    let inner = work.path().join("target").join("t");
    std::fs::create_dir_all(&inner).unwrap();
    // The reason names the worktree itself, not the repository check that comes after it.
    let why = fits(&inner, work.path()).unwrap_err();
    assert_eq!(why, "it lies inside the worktree");
    let why = fits(work.path(), work.path()).unwrap_err();
    assert_eq!(why, "it lies inside the worktree");
}

/// AC-16: a case a test cannot build with the real `/tmp`: every candidate fails.
#[test]
fn no_candidate_root_that_fits_gives_the_reason() {
    let (shaped_parent, repo, work) = (scratch(), scratch(), scratch());
    let shaped = shaped_parent
        .path()
        .join("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee");
    std::fs::create_dir(&shaped).unwrap();
    std::fs::create_dir(repo.path().join(".git")).unwrap();
    let from = vec![
        ("the caller's TMPDIR", shaped),
        ("/tmp", repo.path().to_path_buf()),
    ];
    let why = prepare(work.path(), &from).unwrap_err();
    assert!(
        why.contains("the caller's TMPDIR: its path holds a session id shape"),
        "{why}"
    );
    assert!(
        why.contains("/tmp: it lies inside a git repository"),
        "{why}"
    );
    // Without a candidate there is still a reason.
    assert!(prepare(work.path(), &[]).is_err());
}

#[test]
fn the_next_candidate_is_used_when_the_first_does_not_fit() {
    let (shaped_parent, good, work) = (scratch(), scratch(), scratch());
    let shaped = shaped_parent
        .path()
        .join("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee");
    std::fs::create_dir(&shaped).unwrap();
    let from = vec![("first", shaped), ("second", good.path().to_path_buf())];
    let private = prepare(work.path(), &from).unwrap();
    assert_eq!(private.root, good.path().canonicalize().unwrap());
    assert!(private.path.is_dir());
    let name = private
        .path
        .file_name()
        .unwrap()
        .to_string_lossy()
        .to_string();
    assert!(name.starts_with("moai-affected-"), "{name}");
    assert!(!session_shaped(&private.path), "{name}");
    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt;
        let mode = std::fs::metadata(&private.path)
            .unwrap()
            .permissions()
            .mode();
        assert_eq!(mode & 0o777, 0o700);
    }
}

/// AC-17: a folder, a file and a link that carry a name are never adopted, removed or followed.
#[cfg(unix)]
#[test]
fn a_taken_name_is_never_adopted() {
    let (root, other) = (scratch(), scratch());
    let root_path = root.path().canonicalize().unwrap();
    std::fs::create_dir(root_path.join("n1")).unwrap();
    std::fs::write(root_path.join("n1").join("keep"), "folder\n").unwrap();
    std::fs::write(root_path.join("n2"), "file\n").unwrap();
    std::fs::write(other.path().join("keep"), "target\n").unwrap();
    std::os::unix::fs::symlink(other.path(), root_path.join("n3")).unwrap();
    std::os::unix::fs::symlink(root_path.join("missing"), root_path.join("n4")).unwrap();

    // Every name taken: nothing is made and nothing taken is touched.
    let mut taken = ["n1", "n2", "n3", "n4"].into_iter().map(String::from);
    let why = create_in(&root_path, &mut taken).unwrap_err();
    assert!(why.contains("taken"), "{why}");
    assert_eq!(
        std::fs::read_to_string(root_path.join("n1/keep")).unwrap(),
        "folder\n"
    );
    assert_eq!(
        std::fs::read_to_string(root_path.join("n2")).unwrap(),
        "file\n"
    );
    assert_eq!(
        std::fs::read_to_string(other.path().join("keep")).unwrap(),
        "target\n"
    );
    assert!(root_path
        .join("n3")
        .symlink_metadata()
        .unwrap()
        .file_type()
        .is_symlink());
    assert!(!root_path.join("missing").exists());

    // A free name after the taken ones is taken instead.
    let mut names = ["n1", "n2", "n3", "free"].into_iter().map(String::from);
    let made = create_in(&root_path, &mut names).unwrap();
    assert_eq!(made.path, root_path.join("free"));
}

/// AC-18: a removal takes the created folder and its content, and not the root or a sibling.
#[test]
fn a_passed_run_removes_only_its_own_folder() {
    let (root, work) = (scratch(), scratch());
    let from = vec![("root", root.path().to_path_buf())];
    let private = prepare(work.path(), &from).unwrap();
    std::fs::create_dir_all(private.path.join("deep/er")).unwrap();
    std::fs::write(private.path.join("deep/er/file"), "x\n").unwrap();
    let sibling = root.path().join("sibling");
    std::fs::create_dir(&sibling).unwrap();
    std::fs::write(sibling.join("file"), "keep\n").unwrap();
    let path = private.path.clone();
    let answer = finish(private, 0);
    assert_eq!(answer["state"], "removed", "{answer}");
    assert!(!path.exists());
    assert_eq!(
        std::fs::read_to_string(sibling.join("file")).unwrap(),
        "keep\n"
    );
    assert!(root.path().is_dir());
}

#[test]
fn a_failed_run_keeps_the_folder_and_names_it() {
    let (root, work) = (scratch(), scratch());
    let from = vec![("root", root.path().to_path_buf())];
    let private = prepare(work.path(), &from).unwrap();
    std::fs::write(private.path.join("file"), "x\n").unwrap();
    let path = private.path.clone();
    let answer = finish(private, 100);
    assert_eq!(answer["state"], "kept", "{answer}");
    assert_eq!(answer["path"], path.to_string_lossy().as_ref());
    assert!(path.join("file").is_file());
}

/// AC-18: a folder replaced by a link is never followed or removed.
#[cfg(unix)]
#[test]
fn a_swapped_link_is_never_followed_or_removed() {
    let (root, work, other) = (scratch(), scratch(), scratch());
    let from = vec![("root", root.path().to_path_buf())];
    let private = prepare(work.path(), &from).unwrap();
    std::fs::write(other.path().join("keep"), "target\n").unwrap();
    std::fs::remove_dir(&private.path).unwrap();
    std::os::unix::fs::symlink(other.path(), &private.path).unwrap();
    let path = private.path.clone();
    let answer = finish(private, 0);
    assert_eq!(answer["state"], "left", "{answer}");
    assert!(
        answer["reason"]
            .as_str()
            .unwrap()
            .contains("no longer a folder"),
        "{answer}"
    );
    assert_eq!(
        std::fs::read_to_string(other.path().join("keep")).unwrap(),
        "target\n"
    );
    assert!(path.symlink_metadata().unwrap().file_type().is_symlink());
}

#[test]
fn a_file_in_place_of_the_folder_is_left() {
    let (root, work) = (scratch(), scratch());
    let from = vec![("root", root.path().to_path_buf())];
    let private = prepare(work.path(), &from).unwrap();
    std::fs::remove_dir(&private.path).unwrap();
    std::fs::write(&private.path, "now a file\n").unwrap();
    let path = private.path.clone();
    let answer = finish(private, 0);
    assert_eq!(answer["state"], "left", "{answer}");
    assert!(path.is_file());
}

/// A folder whose parent is not the recorded root is left, even when it is a folder.
#[test]
fn a_folder_under_another_parent_is_left() {
    let (root, work, elsewhere) = (scratch(), scratch(), scratch());
    let from = vec![("root", root.path().to_path_buf())];
    let mut private = prepare(work.path(), &from).unwrap();
    std::fs::create_dir(elsewhere.path().join("x")).unwrap();
    private.path = elsewhere.path().canonicalize().unwrap().join("x");
    let why = remove_own(&private).unwrap_err();
    assert!(why.contains("parent"), "{why}");
    assert!(elsewhere.path().join("x").is_dir());
}

/// A folder that is no link itself but whose path now runs through a link resolves elsewhere:
/// it is left, and what the link leads to keeps its content.
#[cfg(unix)]
#[test]
fn a_path_that_now_runs_through_a_link_is_left() {
    let (holder, work) = (scratch(), scratch());
    let root = holder.path().canonicalize().unwrap().join("root");
    std::fs::create_dir(&root).unwrap();
    let from = vec![("root", root.clone())];
    let private = prepare(work.path(), &from).unwrap();
    std::fs::write(private.path.join("keep"), "kept\n").unwrap();
    // The root moves away and a link takes its place.
    let moved = holder.path().canonicalize().unwrap().join("moved");
    std::fs::rename(&root, &moved).unwrap();
    std::os::unix::fs::symlink(&moved, &root).unwrap();
    let answer = finish(private, 0);
    assert_eq!(answer["state"], "left", "{answer}");
    assert!(
        answer["reason"].as_str().unwrap().contains("resolves"),
        "{answer}"
    );
    let name = moved
        .read_dir()
        .unwrap()
        .next()
        .unwrap()
        .unwrap()
        .file_name();
    assert_eq!(
        std::fs::read_to_string(moved.join(name).join("keep")).unwrap(),
        "kept\n"
    );
}
