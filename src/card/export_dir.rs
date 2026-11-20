//! The export folder on disk (#3070 K0, acceptance criteria AC6 and AC8).
//!
//! [`write`] puts the files of an [`Export`] into a folder and [`read`] gets them back. Neither
//! opens a ledger. The rules are the ones a user's own folder needs:
//! - `write` goes only into a folder that does not exist yet or is empty, and it makes every
//!   file with `create_new`, so nothing that was already there is ever replaced or removed.
//! - What is written is private to its owner, like the ledger it comes from (unix): a new folder
//!   is made 0700 and every file 0600; an existing empty folder that group or others can open is
//!   refused ([`DirError::TooOpen`]), not narrowed, because it is not ours to change.
//! - [`resolve_out`] says where the folder really is before anything is written: links in the
//!   parent folders are followed and the real path is what is used and shown, and a folder
//!   inside the MoAI state folder (which holds the ledgers) is refused.
//! - `index.json` is written last. It names every other file with a digest, so a folder that
//!   was cut short (disk full, killed) has no index and `read` refuses it; nothing is deleted
//!   to hide the half-written files.
//! - `read` takes only the files the format writes (`index.json`, `events.jsonl`, `C-<n>.md`).
//!   A link, a folder, a stray file such as `.DS_Store`, a file over [`MAX_FILE_BYTES`] or text
//!   that is not UTF-8 is refused by name, never skipped.

use super::export_format::{
    card_file_name, parse, Export, ExportError, Snapshot, EVENTS_FILE, INDEX_FILE,
};
use std::collections::BTreeSet;
use std::ffi::OsString;
use std::fmt;
use std::fs::{self, OpenOptions};
use std::io::{Read, Write};
use std::path::{Path, PathBuf};

/// The largest file `read` takes. A ledger of many thousand cards stays far below it.
pub const MAX_FILE_BYTES: u64 = 64 * 1024 * 1024;
/// The most card files `read` takes.
pub const MAX_CARD_FILES: usize = 100_000;

/// Why a folder cannot be written or read.
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum DirError {
    /// The folder does not exist (`read`).
    Missing(PathBuf),
    /// The path is not a folder.
    NotADirectory(PathBuf),
    /// The path is a link. Neither `write` nor `read` follows a link to the folder.
    Link(PathBuf),
    /// `write` was given a folder that already holds something.
    NotEmpty(PathBuf),
    /// `write` was given an existing empty folder that group or others can open (unix).
    TooOpen(PathBuf),
    /// The path names no folder (it ends in `..` or is a root).
    NoName(PathBuf),
    /// The folder would be inside the MoAI state folder, where the ledgers live.
    InStateFolder { path: PathBuf, state: PathBuf },
    /// A file name of the export is not one the format writes.
    BadName(String),
    /// An entry of the folder is not one of the format's files, or is not a plain file.
    Unexpected(String),
    /// A file is over the size limit.
    TooLarge { name: String, limit: u64 },
    /// The folder holds more card files than the limit.
    TooMany(usize),
    /// A file is not UTF-8 text.
    NotUtf8(String),
    /// The file system said no.
    Io { path: PathBuf, message: String },
    /// The files are read but are not a valid export.
    Export(ExportError),
}

impl fmt::Display for DirError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::Missing(p) => write!(f, "{} does not exist", p.display()),
            Self::NotADirectory(p) => write!(f, "{} is not a folder", p.display()),
            Self::Link(p) => write!(f, "{} is a link; give the folder itself", p.display()),
            Self::NotEmpty(p) => write!(
                f,
                "{} already holds files; the export goes into an empty or new folder",
                p.display()
            ),
            Self::TooOpen(p) => write!(
                f,
                "{} is open to other users; run `chmod 700` on it first, or give a new folder \
                 (an export holds card text and session ids)",
                p.display()
            ),
            Self::NoName(p) => write!(f, "{} does not name a folder", p.display()),
            Self::InStateFolder { path, state } => write!(
                f,
                "{} is inside the MoAI state folder {}, which holds the ledgers; \
                 give a folder outside it",
                path.display(),
                state.display()
            ),
            Self::BadName(n) => write!(f, "{n:?} is not an export file name"),
            Self::Unexpected(n) => write!(
                f,
                "{n:?} is not a file of the export (a link, a folder or another name)"
            ),
            Self::TooLarge { name, limit } => {
                write!(f, "export file {name} is over {limit} bytes")
            }
            Self::TooMany(n) => write!(f, "the folder holds {n} entries, more than an export does"),
            Self::NotUtf8(n) => write!(f, "export file {n} is not UTF-8 text"),
            Self::Io { path, message } => write!(f, "{}: {message}", path.display()),
            Self::Export(e) => e.fmt(f),
        }
    }
}

impl std::error::Error for DirError {}

impl From<ExportError> for DirError {
    fn from(e: ExportError) -> Self {
        Self::Export(e)
    }
}

fn io(path: &Path) -> impl FnOnce(std::io::Error) -> DirError + '_ {
    move |e| DirError::Io {
        path: path.to_path_buf(),
        message: e.to_string(),
    }
}

/// `C-<n>.md` exactly as `card_file_name` writes it (digits only, no leading zero).
fn card_number(name: &str) -> Option<i64> {
    let digits = name.strip_prefix("C-")?.strip_suffix(".md")?;
    if digits.is_empty() || !digits.bytes().all(|b| b.is_ascii_digit()) {
        return None;
    }
    let id: i64 = digits.parse().ok()?;
    (card_file_name(id) == name).then_some(id)
}

/// Every name of `export` is one the format writes, and no name repeats.
fn check_names(export: &Export) -> Result<(), DirError> {
    let mut seen = BTreeSet::new();
    for (name, _) in &export.cards {
        if card_number(name).is_none() || !seen.insert(name.as_str()) {
            return Err(DirError::BadName(name.clone()));
        }
    }
    Ok(())
}

/// One new folder, 0700 on unix. Only the last step is made: the parent must exist.
fn make_private_dir(dir: &Path) -> std::io::Result<()> {
    let builder = fs::DirBuilder::new();
    #[cfg(unix)]
    let builder = {
        use std::os::unix::fs::DirBuilderExt;
        let mut builder = builder;
        builder.mode(0o700);
        builder
    };
    builder.create(dir)
}

/// The folder is new (made here, private) or empty and closed to other users. The parent folder
/// must exist.
fn prepare(dir: &Path) -> Result<(), DirError> {
    match fs::symlink_metadata(dir) {
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => {
            make_private_dir(dir).map_err(io(dir))
        }
        Err(e) => Err(io(dir)(e)),
        Ok(m) if m.file_type().is_symlink() => Err(DirError::Link(dir.to_path_buf())),
        Ok(m) if !m.is_dir() => Err(DirError::NotADirectory(dir.to_path_buf())),
        Ok(m) => {
            let mut entries = fs::read_dir(dir).map_err(io(dir))?;
            match entries.next() {
                None => {}
                Some(Err(e)) => return Err(io(dir)(e)),
                Some(Ok(_)) => return Err(DirError::NotEmpty(dir.to_path_buf())),
            }
            #[cfg(unix)]
            {
                use std::os::unix::fs::PermissionsExt;
                if m.permissions().mode() & 0o077 != 0 {
                    return Err(DirError::TooOpen(dir.to_path_buf()));
                }
            }
            #[cfg(not(unix))]
            let _ = m;
            Ok(())
        }
    }
}

/// Where `out` really is, before anything is written. The folder's parent must exist and is
/// resolved (a link in it, such as `/var` on macOS, is followed), the last name is kept: whether
/// that name is a link is `prepare`'s to refuse. A folder inside `state_root` (the MoAI state
/// folder `<HOME>/.moai`, which holds every ledger) is refused, also when it is reached through a
/// link or by another letter case on a case-insensitive volume. The returned path is the one to
/// write to and to show.
pub fn resolve_out(out: &Path, state_root: &Path) -> Result<PathBuf, DirError> {
    let name = out
        .file_name()
        .ok_or_else(|| DirError::NoName(out.to_path_buf()))?;
    let parent = match out.parent() {
        Some(p) if !p.as_os_str().is_empty() => p,
        _ => Path::new("."),
    };
    let real_parent = fs::canonicalize(parent).map_err(io(parent))?;
    let target = real_parent.join(name);
    let state = fs::canonicalize(state_root).unwrap_or_else(|_| state_root.to_path_buf());
    if target.starts_with(&state) {
        return Err(DirError::InStateFolder {
            path: target,
            state,
        });
    }
    Ok(target)
}

/// One new file: it fails if the name is taken, whatever it is (file, folder, link).
fn put(dir: &Path, name: &str, text: &str) -> Result<(), DirError> {
    let path = dir.join(name);
    let mut options = OpenOptions::new();
    options.write(true).create_new(true);
    #[cfg(unix)]
    {
        use std::os::unix::fs::OpenOptionsExt;
        options.mode(0o600);
    }
    let mut file = options.open(&path).map_err(io(&path))?;
    file.write_all(text.as_bytes()).map_err(io(&path))?;
    file.sync_all().map_err(io(&path))
}

/// The files in the order that keeps a cut-short folder recognisable: cards, events, index.
fn put_all(dir: &Path, export: &Export) -> Result<(), DirError> {
    for (name, text) in &export.cards {
        put(dir, name, text)?;
    }
    put(dir, EVENTS_FILE, &export.events)?;
    put(dir, INDEX_FILE, &export.index)?;
    // Making the new names durable is a courtesy: some file systems cannot sync a folder.
    if let Ok(d) = fs::File::open(dir) {
        let _ = d.sync_all();
    }
    Ok(())
}

/// Writes `export` into `dir`, which is made if it does not exist (its parent must) or must be
/// empty. Nothing already in `dir` is replaced; on an error the files written so far stay and
/// there is no `index.json`.
pub fn write(dir: &Path, export: &Export) -> Result<(), DirError> {
    check_names(export)?;
    prepare(dir)?;
    put_all(dir, export)
}

/// A file name as text. A name that is not UTF-8 is refused and shown with replacement marks.
fn utf8_name(name: OsString) -> Result<String, DirError> {
    name.into_string()
        .map_err(|n| DirError::NotUtf8(n.to_string_lossy().into_owned()))
}

/// One file's text. The file is opened without following a link and must be a plain file.
fn read_text(dir: &Path, name: &str, max: u64) -> Result<String, DirError> {
    let path = dir.join(name);
    let mut options = OpenOptions::new();
    options.read(true);
    #[cfg(unix)]
    {
        use std::os::unix::fs::OpenOptionsExt;
        // O_NONBLOCK: a pipe swapped in after the listing must not hang the open.
        options.custom_flags(libc::O_NOFOLLOW | libc::O_NONBLOCK);
    }
    let file = options.open(&path).map_err(io(&path))?;
    if !file.metadata().map_err(io(&path))?.is_file() {
        return Err(DirError::Unexpected(name.to_string()));
    }
    let mut bytes = Vec::new();
    file.take(max + 1)
        .read_to_end(&mut bytes)
        .map_err(io(&path))?;
    if bytes.len() as u64 > max {
        return Err(DirError::TooLarge {
            name: name.to_string(),
            limit: max,
        });
    }
    String::from_utf8(bytes).map_err(|_| DirError::NotUtf8(name.to_string()))
}

fn read_with_limit(dir: &Path, max: u64) -> Result<Export, DirError> {
    match fs::symlink_metadata(dir) {
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => {
            return Err(DirError::Missing(dir.to_path_buf()))
        }
        Err(e) => return Err(io(dir)(e)),
        Ok(m) if m.file_type().is_symlink() => return Err(DirError::Link(dir.to_path_buf())),
        Ok(m) if !m.is_dir() => return Err(DirError::NotADirectory(dir.to_path_buf())),
        Ok(_) => {}
    }
    let mut entries = Vec::new();
    for entry in fs::read_dir(dir).map_err(io(dir))? {
        let entry = entry.map_err(io(dir))?;
        let name = utf8_name(entry.file_name())?;
        let plain = entry.file_type().map_err(io(dir))?.is_file();
        entries.push((name, plain));
        if entries.len() > MAX_CARD_FILES + 2 {
            return Err(DirError::TooMany(entries.len()));
        }
    }
    entries.sort();
    let (mut has_index, mut has_events) = (false, false);
    let mut cards = Vec::new();
    for (name, plain) in entries {
        if !plain {
            return Err(DirError::Unexpected(name));
        }
        if name == INDEX_FILE {
            has_index = true;
        } else if name == EVENTS_FILE {
            has_events = true;
        } else if let Some(id) = card_number(&name) {
            cards.push((id, name));
        } else {
            return Err(DirError::Unexpected(name));
        }
    }
    if !has_index {
        return Err(ExportError::MissingFile(INDEX_FILE.into()).into());
    }
    if !has_events {
        return Err(ExportError::MissingFile(EVENTS_FILE.into()).into());
    }
    // By card number, as `render` lists them: `C-2.md` comes before `C-10.md`.
    cards.sort();
    let mut files = Vec::new();
    for (_, name) in cards {
        let text = read_text(dir, &name, max)?;
        files.push((name, text));
    }
    Ok(Export {
        index: read_text(dir, INDEX_FILE, max)?,
        cards: files,
        events: read_text(dir, EVENTS_FILE, max)?,
    })
}

/// The files of an export folder, as `write` made them. Whether they are a valid export is
/// for `parse` (or [`load`]) to say.
pub fn read(dir: &Path) -> Result<Export, DirError> {
    read_with_limit(dir, MAX_FILE_BYTES)
}

/// The rows of an export folder: [`read`] and `parse`.
pub fn load(dir: &Path) -> Result<Snapshot, DirError> {
    Ok(parse(&read(dir)?)?)
}

#[cfg(test)]
mod tests {
    use super::super::export_format::{canonical, render, tests::sample};
    use super::*;

    fn export() -> Export {
        render(&sample()).unwrap()
    }

    /// An empty folder the way `write` takes it: closed to other users (unix).
    fn private_dir(path: &Path) {
        fs::create_dir(path).unwrap();
        #[cfg(unix)]
        {
            use std::os::unix::fs::PermissionsExt;
            fs::set_permissions(path, fs::Permissions::from_mode(0o700)).unwrap();
        }
    }

    fn names(dir: &Path) -> Vec<String> {
        let mut v: Vec<String> = fs::read_dir(dir)
            .unwrap()
            .map(|e| e.unwrap().file_name().into_string().unwrap())
            .collect();
        v.sort();
        v
    }

    #[test]
    fn a_folder_round_trips_to_the_same_rows_and_the_same_bytes() {
        let tmp = tempfile::tempdir().unwrap();
        let dir = tmp.path().join("out");
        let e = export();
        write(&dir, &e).unwrap();
        assert_eq!(
            names(&dir),
            ["C-1.md", "C-2.md", "C-5.md", "events.jsonl", "index.json"]
        );
        assert_eq!(read(&dir).unwrap(), e);
        assert_eq!(load(&dir).unwrap(), canonical(&sample()).unwrap());
        // Reading and writing again gives the same bytes in another folder.
        let again = tmp.path().join("again");
        write(&again, &read(&dir).unwrap()).unwrap();
        for n in names(&dir) {
            assert_eq!(
                fs::read(dir.join(&n)).unwrap(),
                fs::read(again.join(&n)).unwrap(),
                "{n}"
            );
        }
    }

    #[test]
    fn a_ledger_without_cards_writes_and_reads_back() {
        let tmp = tempfile::tempdir().unwrap();
        let mut s = sample();
        s.cards.clear();
        s.acs.clear();
        s.notes.clear();
        s.links.clear();
        s.events.clear();
        let e = render(&s).unwrap();
        let dir = tmp.path().join("empty-ledger");
        write(&dir, &e).unwrap();
        assert_eq!(names(&dir), ["events.jsonl", "index.json"]);
        assert_eq!(load(&dir).unwrap(), canonical(&s).unwrap());
    }

    #[test]
    fn card_files_come_back_in_card_number_order_not_name_order() {
        let tmp = tempfile::tempdir().unwrap();
        let mut s = sample();
        let (mut c2, mut c10) = (s.cards[1].clone(), s.cards[1].clone());
        c2.card_id = 2;
        c10.card_id = 10;
        s.cards = vec![s.cards[0].clone(), c2, c10];
        s.acs.clear();
        s.notes.clear();
        s.links.clear();
        s.events.clear();
        let e = render(&s).unwrap();
        let dir = tmp.path().join("o");
        write(&dir, &e).unwrap();
        let got = read(&dir).unwrap();
        let order: Vec<&str> = got.cards.iter().map(|(n, _)| n.as_str()).collect();
        assert_eq!(order, ["C-1.md", "C-2.md", "C-10.md"]);
        assert_eq!(got, e);
    }

    #[test]
    fn an_empty_existing_folder_is_taken_and_a_folder_with_anything_in_it_is_not() {
        let tmp = tempfile::tempdir().unwrap();
        let dir = tmp.path().join("o");
        private_dir(&dir);
        write(&dir, &export()).unwrap();
        // Written again into the folder that now holds an export: refused, nothing changed.
        let before: Vec<Vec<u8>> = names(&dir)
            .iter()
            .map(|n| fs::read(dir.join(n)).unwrap())
            .collect();
        assert_eq!(write(&dir, &export()), Err(DirError::NotEmpty(dir.clone())));
        let after: Vec<Vec<u8>> = names(&dir)
            .iter()
            .map(|n| fs::read(dir.join(n)).unwrap())
            .collect();
        assert_eq!(before, after);
        // Any single entry counts, even a hidden one or a folder.
        for stray in [".hidden", "sub"] {
            let d = tmp.path().join(format!("with-{stray}"));
            fs::create_dir(&d).unwrap();
            if stray == "sub" {
                fs::create_dir(d.join(stray)).unwrap();
            } else {
                fs::write(d.join(stray), "x").unwrap();
            }
            assert_eq!(write(&d, &export()), Err(DirError::NotEmpty(d.clone())));
            assert_eq!(names(&d), [stray]);
        }
    }

    #[test]
    fn the_path_must_be_a_real_folder_or_nothing() {
        let tmp = tempfile::tempdir().unwrap();
        let file = tmp.path().join("f");
        fs::write(&file, "keep").unwrap();
        assert_eq!(
            write(&file, &export()),
            Err(DirError::NotADirectory(file.clone()))
        );
        assert_eq!(fs::read_to_string(&file).unwrap(), "keep");
        assert_eq!(read(&file), Err(DirError::NotADirectory(file.clone())));
        // The parent is not made: a typo in the path does not build a tree.
        let deep = tmp.path().join("no-such").join("out");
        assert!(matches!(write(&deep, &export()), Err(DirError::Io { .. })));
        assert!(!tmp.path().join("no-such").exists());
        let gone = tmp.path().join("gone");
        assert_eq!(read(&gone), Err(DirError::Missing(gone.clone())));
        assert!(!gone.exists());
    }

    #[cfg(unix)]
    #[test]
    fn a_link_to_the_folder_is_refused_both_ways() {
        let tmp = tempfile::tempdir().unwrap();
        let real = tmp.path().join("real");
        private_dir(&real);
        let link = tmp.path().join("link");
        std::os::unix::fs::symlink(&real, &link).unwrap();
        assert_eq!(write(&link, &export()), Err(DirError::Link(link.clone())));
        assert!(names(&real).is_empty());
        write(&real, &export()).unwrap();
        assert_eq!(read(&link), Err(DirError::Link(link.clone())));
        // A dangling link is a link too, and nothing is made behind it.
        let dangling = tmp.path().join("dangling");
        std::os::unix::fs::symlink(tmp.path().join("nowhere"), &dangling).unwrap();
        assert_eq!(
            write(&dangling, &export()),
            Err(DirError::Link(dangling.clone()))
        );
        assert!(!tmp.path().join("nowhere").exists());
    }

    #[test]
    fn export_names_that_are_not_the_formats_are_refused_before_anything_is_written() {
        let tmp = tempfile::tempdir().unwrap();
        for bad in [
            "../escape.md",
            "sub/C-1.md",
            "C-01.md",
            "C-.md",
            "C--1.md",
            "C-1.MD",
            "index.json",
            "events.jsonl",
            "",
            "C-99999999999999999999.md",
        ] {
            let mut e = export();
            e.cards.push((bad.to_string(), "x".into()));
            let dir = tmp.path().join("o");
            assert_eq!(
                write(&dir, &e),
                Err(DirError::BadName(bad.to_string())),
                "{bad:?}"
            );
            assert!(!dir.exists(), "{bad:?} left a folder behind");
        }
        let mut twice = export();
        twice.cards.push(twice.cards[0].clone());
        assert_eq!(
            write(&tmp.path().join("o"), &twice),
            Err(DirError::BadName("C-1.md".into()))
        );
    }

    #[test]
    fn a_name_that_is_taken_is_never_replaced() {
        let tmp = tempfile::tempdir().unwrap();
        fs::write(tmp.path().join("C-1.md"), "keep").unwrap();
        assert!(matches!(
            put(tmp.path(), "C-1.md", "new"),
            Err(DirError::Io { .. })
        ));
        assert_eq!(
            fs::read_to_string(tmp.path().join("C-1.md")).unwrap(),
            "keep"
        );
        fs::create_dir(tmp.path().join("d")).unwrap();
        assert!(put(tmp.path(), "d", "new").is_err());
    }

    #[test]
    fn a_write_that_stops_midway_leaves_no_index() {
        let tmp = tempfile::tempdir().unwrap();
        // The events file cannot be made: a folder holds its name. The cards come first.
        fs::create_dir(tmp.path().join(EVENTS_FILE)).unwrap();
        let err = put_all(tmp.path(), &export()).unwrap_err();
        assert!(matches!(err, DirError::Io { .. }), "{err:?}");
        assert!(tmp.path().join("C-1.md").is_file());
        assert!(!tmp.path().join(INDEX_FILE).exists());
        // And such a folder is not readable as an export: its index is missing.
        let half = tmp.path().join("half");
        fs::create_dir(&half).unwrap();
        for (n, t) in &export().cards {
            fs::write(half.join(n), t).unwrap();
        }
        fs::write(half.join(EVENTS_FILE), export().events).unwrap();
        assert_eq!(
            read(&half),
            Err(DirError::Export(ExportError::MissingFile(
                INDEX_FILE.into()
            )))
        );
    }

    fn written(tmp: &Path) -> PathBuf {
        let dir = tmp.join("o");
        write(&dir, &export()).unwrap();
        dir
    }

    #[test]
    fn a_missing_events_file_or_card_file_is_named() {
        let tmp = tempfile::tempdir().unwrap();
        let dir = written(tmp.path());
        fs::remove_file(dir.join(EVENTS_FILE)).unwrap();
        assert_eq!(
            read(&dir),
            Err(DirError::Export(ExportError::MissingFile(
                EVENTS_FILE.into()
            )))
        );
        fs::write(dir.join(EVENTS_FILE), export().events).unwrap();
        fs::remove_file(dir.join("C-2.md")).unwrap();
        // `read` hands over the files; the index check that a card file is missing is `parse`'s.
        assert_eq!(
            load(&dir),
            Err(DirError::Export(ExportError::MissingFile("C-2.md".into())))
        );
    }

    // No name here differs from a card file only by case: on a case-insensitive volume (the
    // default on macOS) that is the same file.
    #[test]
    fn a_stray_file_or_folder_is_refused_by_name() {
        let tmp = tempfile::tempdir().unwrap();
        for stray in [".DS_Store", "notes.txt", "C-01.md", "C-1.txt", "C-1.md.bak"] {
            let dir = tmp.path().join(stray);
            private_dir(&dir);
            write(&dir, &export()).unwrap();
            fs::write(dir.join(stray), "x").unwrap();
            assert_eq!(
                read(&dir),
                Err(DirError::Unexpected(stray.to_string())),
                "{stray}"
            );
        }
        let dir = written(&tmp.path().join("."));
        fs::create_dir(dir.join("sub")).unwrap();
        assert_eq!(read(&dir), Err(DirError::Unexpected("sub".into())));
        // A card file that is a folder is not a plain file either.
        let dir2 = tmp.path().join("o2");
        write(&dir2, &export()).unwrap();
        fs::remove_file(dir2.join("C-5.md")).unwrap();
        fs::create_dir(dir2.join("C-5.md")).unwrap();
        assert_eq!(read(&dir2), Err(DirError::Unexpected("C-5.md".into())));
    }

    #[cfg(unix)]
    #[test]
    fn a_link_in_place_of_a_file_is_refused_and_not_followed() {
        let tmp = tempfile::tempdir().unwrap();
        let outside = tmp.path().join("outside.txt");
        fs::write(&outside, "secret").unwrap();
        for target in ["C-1.md", EVENTS_FILE, INDEX_FILE] {
            let dir = tmp.path().join(format!("o-{target}"));
            write(&dir, &export()).unwrap();
            fs::remove_file(dir.join(target)).unwrap();
            std::os::unix::fs::symlink(&outside, dir.join(target)).unwrap();
            assert_eq!(
                read(&dir),
                Err(DirError::Unexpected(target.to_string())),
                "{target}"
            );
        }
    }

    #[cfg(unix)]
    #[test]
    fn a_file_swapped_for_a_link_after_the_listing_is_still_not_followed() {
        let tmp = tempfile::tempdir().unwrap();
        let outside = tmp.path().join("outside.txt");
        fs::write(&outside, "secret").unwrap();
        let dir = tmp.path().join("o");
        fs::create_dir(&dir).unwrap();
        std::os::unix::fs::symlink(&outside, dir.join("C-1.md")).unwrap();
        // `read_text` is what runs after the listing said "plain file".
        let err = read_text(&dir, "C-1.md", MAX_FILE_BYTES).unwrap_err();
        assert!(matches!(err, DirError::Io { .. }), "{err:?}");
    }

    #[cfg(unix)]
    #[test]
    fn a_file_name_that_is_not_utf8_is_refused_and_shown_with_replacement_marks() {
        use std::os::unix::ffi::OsStringExt;
        assert_eq!(
            utf8_name(OsString::from("C-1.md")),
            Ok("C-1.md".to_string())
        );
        assert_eq!(
            utf8_name(OsString::from_vec(vec![b'C', 0xff, b'.', b'm', b'd'])),
            Err(DirError::NotUtf8("C\u{fffd}.md".into()))
        );
    }

    #[cfg(unix)]
    #[test]
    fn a_pipe_in_place_of_a_file_is_refused_and_does_not_wait_for_a_writer() {
        use std::os::unix::ffi::OsStrExt;
        let tmp = tempfile::tempdir().unwrap();
        let c = std::ffi::CString::new(tmp.path().join("C-1.md").as_os_str().as_bytes()).unwrap();
        assert_eq!(unsafe { libc::mkfifo(c.as_ptr(), 0o600) }, 0, "mkfifo");
        // The listing would already say "not a plain file"; this is the file swapped in later.
        assert_eq!(
            read_text(tmp.path(), "C-1.md", MAX_FILE_BYTES),
            Err(DirError::Unexpected("C-1.md".into()))
        );
    }

    #[test]
    fn a_file_over_the_limit_or_not_text_is_refused_by_name() {
        let tmp = tempfile::tempdir().unwrap();
        let dir = written(tmp.path());
        let index_len = fs::metadata(dir.join(INDEX_FILE)).unwrap().len();
        // The limit is on bytes read: one byte under passes, exactly at it passes, one over fails.
        assert!(read_with_limit(&dir, 1 << 20).is_ok());
        let longest = names(&dir)
            .iter()
            .map(|n| fs::metadata(dir.join(n)).unwrap().len())
            .max()
            .unwrap();
        assert!(longest >= index_len);
        assert!(read_with_limit(&dir, longest).is_ok());
        match read_with_limit(&dir, longest - 1) {
            Err(DirError::TooLarge { limit, .. }) => assert_eq!(limit, longest - 1),
            other => panic!("{other:?}"),
        }
        fs::write(dir.join(EVENTS_FILE), [0xff, 0xfe, b'\n']).unwrap();
        assert_eq!(read(&dir), Err(DirError::NotUtf8(EVENTS_FILE.into())));
    }

    #[test]
    fn a_folder_that_is_not_the_export_it_names_is_refused_by_parse_through_load() {
        let tmp = tempfile::tempdir().unwrap();
        let dir = written(tmp.path());
        // Same length, so only the digest can tell.
        let mut text = fs::read_to_string(dir.join("C-1.md")).unwrap();
        text = text.replacen("Card 1", "Card X", 1);
        fs::write(dir.join("C-1.md"), text).unwrap();
        assert_eq!(
            load(&dir),
            Err(DirError::Export(ExportError::Digest("C-1.md".into())))
        );
    }

    #[cfg(unix)]
    fn mode_of(path: &Path) -> u32 {
        use std::os::unix::fs::PermissionsExt;
        fs::metadata(path).unwrap().permissions().mode() & 0o777
    }

    #[cfg(unix)]
    #[test]
    fn a_new_folder_is_0700_and_every_file_in_it_0600() {
        let tmp = tempfile::tempdir().unwrap();
        let dir = tmp.path().join("out");
        write(&dir, &export()).unwrap();
        assert_eq!(mode_of(&dir), 0o700);
        for n in names(&dir) {
            assert_eq!(mode_of(&dir.join(&n)), 0o600, "{n}");
        }
        // The parent is not touched.
        assert_ne!(mode_of(tmp.path()) & 0o700, 0);
    }

    #[cfg(unix)]
    #[test]
    fn an_existing_empty_folder_that_others_can_open_is_refused_and_left_as_it_is() {
        use std::os::unix::fs::PermissionsExt;
        let tmp = tempfile::tempdir().unwrap();
        for open in [0o755, 0o750, 0o705, 0o770, 0o777] {
            let dir = tmp.path().join(format!("o{open:o}"));
            fs::create_dir(&dir).unwrap();
            fs::set_permissions(&dir, fs::Permissions::from_mode(open)).unwrap();
            assert_eq!(
                write(&dir, &export()),
                Err(DirError::TooOpen(dir.clone())),
                "{open:o}"
            );
            assert!(names(&dir).is_empty(), "{open:o}");
            assert_eq!(
                mode_of(&dir),
                open,
                "the folder's mode is not ours to change"
            );
        }
        // Closed to others, whatever the owner bits: taken.
        let ok = tmp.path().join("ok");
        fs::create_dir(&ok).unwrap();
        fs::set_permissions(&ok, fs::Permissions::from_mode(0o700)).unwrap();
        write(&ok, &export()).unwrap();
    }

    #[test]
    fn out_with_no_folder_name_or_a_missing_parent_is_refused() {
        let tmp = tempfile::tempdir().unwrap();
        let state = tmp.path().join(".moai");
        fs::create_dir(&state).unwrap();
        for bad in [tmp.path().join("a").join(".."), PathBuf::from("/")] {
            assert!(
                matches!(resolve_out(&bad, &state), Err(DirError::NoName(_))),
                "{bad:?}"
            );
        }
        let missing = tmp.path().join("no-such").join("out");
        assert!(matches!(
            resolve_out(&missing, &state),
            Err(DirError::Io { .. })
        ));
        assert!(!tmp.path().join("no-such").exists());
    }

    #[test]
    fn a_folder_inside_the_state_folder_is_refused_at_any_depth() {
        let tmp = tempfile::tempdir().unwrap();
        let state = tmp.path().join(".moai");
        let ledgers = state.join("projects").join("demo");
        fs::create_dir_all(&ledgers).unwrap();
        for inside in [
            state.join("x"),
            state.join("projects").join("x"),
            ledgers.join("x"),
        ] {
            match resolve_out(&inside, &state) {
                Err(DirError::InStateFolder { path, state: s }) => {
                    assert_eq!(path.file_name(), inside.file_name());
                    assert!(path.starts_with(&s));
                }
                other => panic!("{inside:?}: {other:?}"),
            }
        }
        // Beside it is fine, including a name that only starts like the state folder's.
        for outside in [tmp.path().join("x"), tmp.path().join(".moai-exports")] {
            assert_eq!(
                resolve_out(&outside, &state).unwrap().file_name(),
                outside.file_name()
            );
        }
    }

    #[cfg(unix)]
    #[test]
    fn a_link_in_a_parent_is_followed_and_the_real_path_is_what_is_checked() {
        let tmp = tempfile::tempdir().unwrap();
        let state = tmp.path().join(".moai");
        fs::create_dir_all(state.join("projects")).unwrap();
        let elsewhere = tmp.path().join("elsewhere");
        fs::create_dir(&elsewhere).unwrap();
        // A link that leads out of nowhere into the state folder.
        let into_state = elsewhere.join("looks-harmless");
        std::os::unix::fs::symlink(state.join("projects"), &into_state).unwrap();
        assert!(matches!(
            resolve_out(&into_state.join("x"), &state),
            Err(DirError::InStateFolder { .. })
        ));
        // A link to an ordinary folder: followed, and the real path comes back.
        let real = tmp.path().join("real");
        fs::create_dir(&real).unwrap();
        let link = elsewhere.join("to-real");
        std::os::unix::fs::symlink(&real, &link).unwrap();
        let got = resolve_out(&link.join("x"), &state).unwrap();
        assert_eq!(got, fs::canonicalize(&real).unwrap().join("x"));
        // The state folder itself behind a link is the state folder.
        let state_link = tmp.path().join("state-link");
        std::os::unix::fs::symlink(&state, &state_link).unwrap();
        assert!(matches!(
            resolve_out(&state.join("x"), &state_link),
            Err(DirError::InStateFolder { .. })
        ));
    }
}
