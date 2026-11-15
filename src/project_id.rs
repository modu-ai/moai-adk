//! The project id that does not move (#3009 S1, design comment 6887461482 §3).
//!
//! One random value, written once into the tracked file `<project>/.moai/project.yaml`, so
//! every clone, every worktree and a renamed folder read the same id. A folder path or a
//! remote URL names a place; this names the project. Nothing here touches the HOME or any
//! other file of the project: the Go generation's `.moai/config/sections/project.yaml` is a
//! different file and is neither read nor changed.
//!
//! An existing file is never rewritten. A file that is there but cannot be read as exactly
//! one valid id is refused and left as it is, so a person's edit is not replaced by a new id
//! that would split the project in two.
use chacha20poly1305::aead::{rand_core::RngCore, OsRng};
use std::{
    fs,
    io::Read,
    path::{Path, PathBuf},
};

type Result<T> = std::result::Result<T, String>;

/// Where the id lives, relative to the project folder.
pub const PROJECT_FILE: &str = ".moai/project.yaml";
/// Characters of an id: 128 random bits as lower-case base32 (RFC 4648 alphabet).
pub const ID_LEN: usize = 26;
/// The most this reader takes from the file; a bigger file is not an identity file.
pub const FILE_LIMIT: u64 = 64 * 1024;
/// Tries of the write while the file is still absent after a failed one.
const PUBLISH_ATTEMPTS: usize = 3;

const ALPHABET: &[u8; 32] = b"abcdefghijklmnopqrstuvwxyz234567";
const KEY: &str = "project_id";
const HEADER: &str =
    "# MoAI project identity (#3009). Commit this file: every clone, worktree and\n\
# renamed folder of this project reads the same id from it. Do not edit the id.\n";

/// The id and whether this call wrote it.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Issued {
    pub project_id: String,
    /// `true` for the one call that created the file; `false` when it was already there
    /// (including a call that lost a race to a concurrent one).
    pub created: bool,
}

/// An id is exactly [`ID_LEN`] characters of `a-z` and `2-7`: it is safe as a folder name on
/// a case-insensitive filesystem and carries nothing a path could.
pub fn valid(id: &str) -> bool {
    id.len() == ID_LEN && id.bytes().all(|b| ALPHABET.contains(&b))
}

fn dir_of(project: &Path) -> PathBuf {
    project.join(".moai")
}

fn file_of(project: &Path) -> PathBuf {
    project.join(PROJECT_FILE)
}

/// `.moai` or the file being a link is refused before anything is read or written through it.
fn refuse_links(project: &Path) -> Result<()> {
    for path in [dir_of(project), file_of(project)] {
        if let Ok(m) = fs::symlink_metadata(&path) {
            if m.file_type().is_symlink() {
                return Err(format!(
                    "{}: is a symbolic link; refused",
                    rel_of(project, &path)
                ));
            }
        }
    }
    Ok(())
}

fn rel_of(project: &Path, path: &Path) -> String {
    path.strip_prefix(project)
        .unwrap_or(path)
        .to_string_lossy()
        .replace('\\', "/")
}

/// The id of `project`: `Ok(None)` when there is no `.moai/project.yaml`, `Ok(Some(id))`
/// when it holds exactly one valid id, and an error that names the file when it is there
/// but is a link, is not a regular file, is over [`FILE_LIMIT`], or does not hold one id.
/// Reads only.
pub fn read(project: &Path) -> Result<Option<String>> {
    refuse_links(project)?;
    let path = file_of(project);
    let meta = match fs::symlink_metadata(&path) {
        Ok(m) => m,
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => return Ok(None),
        Err(e) => return Err(format!("{PROJECT_FILE}: {e}")),
    };
    if !meta.is_file() {
        return Err(format!("{PROJECT_FILE}: not a regular file"));
    }
    if meta.len() > FILE_LIMIT {
        return Err(format!("{PROJECT_FILE}: over {FILE_LIMIT} bytes"));
    }
    let mut bytes = Vec::new();
    fs::File::open(&path)
        .and_then(|f| f.take(FILE_LIMIT + 1).read_to_end(&mut bytes))
        .map_err(|e| format!("{PROJECT_FILE}: {e}"))?;
    if bytes.len() as u64 > FILE_LIMIT {
        return Err(format!("{PROJECT_FILE}: over {FILE_LIMIT} bytes"));
    }
    let text = String::from_utf8(bytes).map_err(|_| format!("{PROJECT_FILE}: not UTF-8"))?;
    parse(&text).map(Some)
}

/// The id in `text`: one top-level `project_id:` line, a plain or quoted value, an optional
/// trailing ` # comment`. Other keys, comments and blank lines are ignored.
fn parse(text: &str) -> Result<String> {
    let mut found: Option<String> = None;
    for line in text.lines() {
        let Some(rest) = line.strip_prefix(KEY).and_then(|r| r.strip_prefix(':')) else {
            continue;
        };
        if found.is_some() {
            return Err(format!("{PROJECT_FILE}: `{KEY}` appears more than once"));
        }
        let mut value = rest.trim();
        if let Some(at) = value.find(" #") {
            value = value[..at].trim_end();
        }
        let value = value
            .strip_prefix('"')
            .and_then(|v| v.strip_suffix('"'))
            .or_else(|| value.strip_prefix('\'').and_then(|v| v.strip_suffix('\'')))
            .unwrap_or(value);
        if !valid(value) {
            return Err(format!(
                "{PROJECT_FILE}: `{KEY}` is not {ID_LEN} characters of a-z and 2-7"
            ));
        }
        found = Some(value.to_string());
    }
    found.ok_or_else(|| format!("{PROJECT_FILE}: no `{KEY}` line"))
}

fn new_id() -> Result<String> {
    let mut random = [0u8; 16];
    OsRng
        .try_fill_bytes(&mut random)
        .map_err(|_| "the system random source is unavailable".to_string())?;
    let mut id = String::with_capacity(ID_LEN);
    let (mut buffer, mut bits) = (0u32, 0u32);
    for byte in random {
        buffer = (buffer << 8) | u32::from(byte);
        bits += 8;
        while bits >= 5 {
            bits -= 5;
            id.push(char::from(ALPHABET[((buffer >> bits) & 31) as usize]));
        }
    }
    if bits > 0 {
        id.push(char::from(ALPHABET[((buffer << (5 - bits)) & 31) as usize]));
    }
    Ok(id)
}

/// The id of `project`, written first if the file is not there yet. An existing file is
/// returned as it is (`created: false`) and never rewritten; one that cannot be read as a
/// single valid id is an error and stays untouched. The write fails rather than replaces
/// when the file appeared meanwhile, and that caller reads the winner's id.
pub fn issue(project: &Path) -> Result<Issued> {
    if let Some(project_id) = read(project)? {
        return Ok(Issued {
            project_id,
            created: false,
        });
    }
    let dir = dir_of(project);
    match fs::create_dir(&dir) {
        Ok(()) => {}
        Err(e) if e.kind() == std::io::ErrorKind::AlreadyExists => {}
        Err(e) => return Err(format!(".moai: {e}")),
    }
    // A link or a plain file may have taken the place of `.moai` after the read above.
    refuse_links(project)?;
    if !fs::symlink_metadata(&dir).is_ok_and(|m| m.is_dir()) {
        return Err(".moai: not a folder".into());
    }
    let mut last = String::new();
    for _ in 0..PUBLISH_ATTEMPTS {
        let id = new_id()?;
        let body = format!("{HEADER}{KEY}: {id}\n");
        match crate::apply::publish_exclusive(&file_of(project), body.as_bytes()) {
            Ok(()) => {
                return Ok(Issued {
                    project_id: id,
                    created: true,
                })
            }
            // The publish names the loss of a race in more than one way (the link was
            // refused, or the target changed while the temporary file was written). Whatever
            // it said, a file that now holds a valid id is the winner's, and it is the id.
            Err(e) => match read(project)? {
                Some(project_id) => {
                    return Ok(Issued {
                        project_id,
                        created: false,
                    })
                }
                None => last = e,
            },
        }
    }
    Err(format!("{PROJECT_FILE}: {last}"))
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::collections::HashSet;

    fn project() -> tempfile::TempDir {
        tempfile::tempdir().expect("temp project")
    }

    fn put(project: &Path, text: &str) {
        fs::create_dir_all(project.join(".moai")).unwrap();
        fs::write(project.join(PROJECT_FILE), text).unwrap();
    }

    const GOOD: &str = "abcdefghijklmnopqrstuvwxyz";

    #[test]
    fn a_new_project_gets_one_id_and_keeps_it() {
        let p = project();
        let first = issue(p.path()).unwrap();
        assert!(first.created);
        assert!(valid(&first.project_id), "{}", first.project_id);
        assert_eq!(read(p.path()).unwrap().as_deref(), Some(&*first.project_id));
        let bytes = fs::read(p.path().join(PROJECT_FILE)).unwrap();
        let second = issue(p.path()).unwrap();
        assert!(!second.created);
        assert_eq!(second.project_id, first.project_id);
        assert_eq!(fs::read(p.path().join(PROJECT_FILE)).unwrap(), bytes);
    }

    #[test]
    fn reading_a_project_without_the_file_writes_nothing() {
        let p = project();
        assert_eq!(read(p.path()).unwrap(), None);
        assert!(!p.path().join(".moai").exists());
    }

    #[test]
    fn a_renamed_folder_keeps_its_id() {
        let parent = project();
        let before = parent.path().join("before");
        fs::create_dir(&before).unwrap();
        let id = issue(&before).unwrap().project_id;
        let after = parent.path().join("after");
        fs::rename(&before, &after).unwrap();
        assert_eq!(read(&after).unwrap().as_deref(), Some(&*id));
        assert!(!issue(&after).unwrap().created);
    }

    #[test]
    fn a_file_that_does_not_hold_one_valid_id_is_refused_and_left_alone() {
        let cases: Vec<(&str, String)> = vec![
            ("empty", String::new()),
            ("no key", "name: demo\n".into()),
            (
                "duplicate",
                format!("project_id: {GOOD}\nproject_id: {GOOD}\n"),
            ),
            ("25 characters", format!("project_id: {}\n", &GOOD[..25])),
            ("27 characters", format!("project_id: {GOOD}a\n")),
            (
                "upper case",
                format!("project_id: {}\n", GOOD.to_uppercase()),
            ),
            ("a slash", format!("project_id: {}/\n", &GOOD[..25])),
            ("digit 1", format!("project_id: {}1\n", &GOOD[..25])),
            ("empty value", "project_id:\n".into()),
            ("an indented key only", format!("  project_id: {GOOD}\n")),
        ];
        for (name, text) in cases {
            let p = project();
            put(p.path(), &text);
            let err = read(p.path()).expect_err(name);
            assert!(err.starts_with(PROJECT_FILE), "{name}: {err}");
            assert!(issue(p.path()).is_err(), "{name}: issue must refuse");
            assert_eq!(
                fs::read_to_string(p.path().join(PROJECT_FILE)).unwrap(),
                text,
                "{name}: the file must be left as it was"
            );
        }
    }

    #[test]
    fn id_length_edges_are_25_no_26_yes_27_no() {
        assert!(!valid(&GOOD[..25]));
        assert!(valid(GOOD));
        assert!(!valid(&format!("{GOOD}a")));
        assert!(!valid(""));
    }

    #[test]
    fn the_file_size_limit_passes_at_n_and_refuses_at_n_plus_1() {
        let line = format!("project_id: {GOOD}\n");
        let filler = |total: u64| {
            let pad = total as usize - line.len();
            format!("{line}{}", "#".repeat(pad))
        };
        let p = project();
        put(p.path(), &filler(FILE_LIMIT));
        assert_eq!(read(p.path()).unwrap().as_deref(), Some(GOOD));
        let q = project();
        put(q.path(), &filler(FILE_LIMIT + 1));
        let err = read(q.path()).unwrap_err();
        assert!(err.contains("over"), "{err}");
        assert!(issue(q.path()).is_err());
    }

    #[test]
    fn comments_quotes_other_keys_and_crlf_are_read() {
        for text in [
            format!("# note\nname: demo\nproject_id: {GOOD}\n"),
            format!("project_id: \"{GOOD}\"\n"),
            format!("project_id: '{GOOD}'\n"),
            format!("project_id: {GOOD} # the id\nother: 1\n"),
            format!("name: demo\r\nproject_id: {GOOD}\r\n"),
        ] {
            let p = project();
            put(p.path(), &text);
            assert_eq!(read(p.path()).unwrap().as_deref(), Some(GOOD), "{text:?}");
        }
    }

    #[test]
    fn the_go_generation_project_file_is_not_the_identity_file() {
        let p = project();
        let go = p.path().join(".moai/config/sections");
        fs::create_dir_all(&go).unwrap();
        fs::write(go.join("project.yaml"), format!("project_id: {GOOD}\n")).unwrap();
        assert_eq!(read(p.path()).unwrap(), None);
        let issued = issue(p.path()).unwrap();
        assert!(issued.created);
        assert_ne!(issued.project_id, GOOD);
        assert_eq!(
            fs::read_to_string(go.join("project.yaml")).unwrap(),
            format!("project_id: {GOOD}\n")
        );
    }

    #[test]
    fn ids_differ_and_the_file_carries_no_machine_detail() {
        let mut seen = HashSet::new();
        for _ in 0..20 {
            let p = project();
            let id = issue(p.path()).unwrap().project_id;
            let text = fs::read_to_string(p.path().join(PROJECT_FILE)).unwrap();
            assert!(!text.contains(&*p.path().to_string_lossy()), "{text}");
            assert!(
                !text.contains("/Users/") && !text.contains("/home/"),
                "{text}"
            );
            assert!(seen.insert(id), "two projects got the same id");
        }
    }

    #[test]
    fn eight_concurrent_issuers_agree_on_one_id() {
        let p = project();
        let root = p.path().to_path_buf();
        let barrier = std::sync::Arc::new(std::sync::Barrier::new(8));
        let handles: Vec<_> = (0..8)
            .map(|_| {
                let (root, barrier) = (root.clone(), barrier.clone());
                std::thread::spawn(move || {
                    barrier.wait();
                    issue(&root).unwrap()
                })
            })
            .collect();
        let results: Vec<Issued> = handles.into_iter().map(|h| h.join().unwrap()).collect();
        let ids: HashSet<&str> = results.iter().map(|r| r.project_id.as_str()).collect();
        assert_eq!(ids.len(), 1, "{results:?}");
        assert_eq!(
            results.iter().filter(|r| r.created).count(),
            1,
            "{results:?}"
        );
        assert_eq!(read(&root).unwrap().as_deref(), ids.into_iter().next());
    }

    #[test]
    fn a_plain_file_in_place_of_the_folder_is_refused() {
        let p = project();
        fs::write(p.path().join(".moai"), "not a folder").unwrap();
        assert!(issue(p.path()).is_err());
        assert_eq!(fs::read(p.path().join(".moai")).unwrap(), b"not a folder");
    }

    #[cfg(unix)]
    #[test]
    fn a_link_for_the_folder_or_the_file_is_refused_and_nothing_is_written_through_it() {
        use std::os::unix::fs::symlink;
        // `.moai` is a link to a folder elsewhere.
        let p = project();
        let elsewhere = project();
        symlink(elsewhere.path(), p.path().join(".moai")).unwrap();
        assert!(read(p.path()).unwrap_err().contains("symbolic link"));
        assert!(issue(p.path()).unwrap_err().contains("symbolic link"));
        assert_eq!(fs::read_dir(elsewhere.path()).unwrap().count(), 0);
        // `.moai/project.yaml` is a link to a file elsewhere, with a valid id in it.
        let q = project();
        let target = project();
        let real = target.path().join("real.yaml");
        fs::write(&real, format!("project_id: {GOOD}\n")).unwrap();
        fs::create_dir(q.path().join(".moai")).unwrap();
        symlink(&real, q.path().join(PROJECT_FILE)).unwrap();
        assert!(read(q.path()).unwrap_err().contains("symbolic link"));
        assert!(issue(q.path()).is_err());
        assert_eq!(
            fs::read_to_string(&real).unwrap(),
            format!("project_id: {GOOD}\n")
        );
        // A link that points nowhere is a link too: nothing is created behind it.
        let r = project();
        fs::create_dir(r.path().join(".moai")).unwrap();
        symlink(r.path().join("nowhere"), r.path().join(PROJECT_FILE)).unwrap();
        assert!(issue(r.path()).is_err());
        assert!(!r.path().join("nowhere").exists());
    }
}
