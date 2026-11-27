//! `improve-note` (SPEC-20261121-self-improvement-loop, slice S2): appends one observed problem
//! to the project's `improve/notes.jsonl` (the file `improve-signals` reads). It is the one write
//! of the slice, and it never follows a link. The logic is in this file with its unit tests; the
//! command line tests are `tests/improve_note_cli.rs`.

use moai_cli::forge::policy::format_utc;
use moai_cli::redact;
use moai_cli::state_root;
use serde_json::{json, Value};
use std::io::Write;
use std::path::{Path, PathBuf};

/// Append one observed problem to the project's notes file
///
/// Adds one line of the D7 shape to `improve/notes.jsonl` below the project's state folder,
/// creating the folder and the file when they are absent: `{"v":1,"at":…,"kind":…,"key":…,
/// "text":…,"by":…}`. `improve-signals` reads it back as the class `<kind>:<key>`. A kind outside
/// `correction`, `failure` and `false_claim`, a key over 80 characters, a text over 200, a
/// control character or an append that takes the file over 1 MiB is refused; a secret-shaped word
/// is written as `[redacted]`. A symbolic link or anything but a regular file on the way is
/// refused and never followed.
#[derive(clap::Args)]
pub struct Args {
    /// The project folder.
    #[arg(default_value = ".")]
    project: PathBuf,
    /// What was observed: `correction`, `failure` or `false_claim`.
    #[arg(long)]
    kind: String,
    /// The topic or command the note is about (at most 80 characters).
    #[arg(long)]
    key: String,
    /// What happened, in a short plain sentence (at most 200 characters).
    #[arg(long)]
    text: String,
    /// Who observed it (at most 80 characters).
    #[arg(long, default_value = "agent")]
    by: String,
}

pub const OPERATION: &str = "improve.note";

/// The read limit of `improve-signals` for the notes file.
const FILE_BYTES: u64 = 1024 * 1024;
const KEY_CHARS: usize = 80;
const TEXT_CHARS: usize = 200;
const BY_CHARS: usize = 80;
/// The kinds a line may carry (the same three `improve-signals` reads).
const NOTE_KINDS: [&str; 3] = ["correction", "failure", "false_claim"];

fn refused(why: &str) -> String {
    format!("improve-note refused: {why}")
}

pub fn run(args: Args) -> Result<Value, String> {
    check_shape(&args.kind, &args.key, &args.text, &args.by)?;
    let root = std::fs::canonicalize(&args.project)
        .map_err(|_| moai_cli::contract::PROJECT_PATH_UNAVAILABLE.to_string())?;
    refuse_state_link(&root)?;
    let notes = state_root::path(&root, "improve/notes.jsonl");
    let rel = notes
        .strip_prefix(&root)
        .map_err(|_| refused("the notes file is not below the project folder"))?
        .to_path_buf();

    let key = masked(&args.key);
    let line = note_line(
        &now_utc()?,
        &args.kind,
        &key,
        &masked(&args.text),
        &masked(&args.by),
    );
    let file_bytes = append(&root, &rel, &line)?;
    Ok(json!({
        "action": "note", "path": rel.to_string_lossy(), "kind": args.kind, "key": key,
        "appended": 1, "file_bytes": file_bytes, "writes": true,
    }))
}

/// Refuses a project whose state folder is a symbolic link, under the name this generation
/// writes and under the name the resolver answers. The resolver reads a link as "not ours" and
/// can fall back to the other name, so the path it answers would not show the link and a write
/// would create a second state folder beside it. The names come from `state_root`.
fn refuse_state_link(root: &Path) -> Result<(), String> {
    for name in [state_root::MOAI_DIR, state_root::dir(root)] {
        if std::fs::symlink_metadata(root.join(name)).is_ok_and(|m| m.file_type().is_symlink()) {
            return Err(refused(&format!(
                "the state folder {name} is a symbolic link; it is not followed"
            )));
        }
    }
    Ok(())
}

/// The kind, the lengths and the characters of a note, before anything is read or written.
fn check_shape(kind: &str, key: &str, text: &str, by: &str) -> Result<(), String> {
    if !NOTE_KINDS.contains(&kind) {
        return Err(refused("the kind is correction, failure or false_claim"));
    }
    let n = key.chars().count();
    if n == 0 || n > KEY_CHARS {
        return Err(refused("the key is 1 to 80 characters"));
    }
    if text.chars().count() > TEXT_CHARS {
        return Err(refused("the text is at most 200 characters"));
    }
    if by.chars().count() > BY_CHARS {
        return Err(refused("the observer is at most 80 characters"));
    }
    // A control character in the key would make `improve-signals` count the line as a bad line.
    if [key, text, by]
        .iter()
        .any(|s| s.chars().any(char::is_control))
    {
        return Err(refused(
            "a control character is not allowed in the key, the text or the observer",
        ));
    }
    Ok(())
}

/// A text as it is written: a secret-shaped word becomes `[redacted]`.
fn masked(text: &str) -> String {
    redact::redact_outgoing(&redact::redact_secret_shapes(text)).0
}

fn now_utc() -> Result<String, String> {
    let secs = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map_err(|_| refused("the clock is before 1970"))?
        .as_secs();
    Ok(format_utc(secs as i64))
}

/// The line, without its newline, in the D7 order.
fn note_line(at: &str, kind: &str, key: &str, text: &str, by: &str) -> String {
    json!({"v": 1, "at": at, "kind": kind, "key": key, "text": text, "by": by}).to_string()
}

/// Appends `line` and a newline to `root/rel`, one folder at a time and without following a
/// link; returns the size of the file after the write.
fn append(root: &Path, rel: &Path, line: &str) -> Result<u64, String> {
    let parts: Vec<&std::ffi::OsStr> = rel.iter().collect();
    let (file_name, folders) = parts
        .split_last()
        .ok_or_else(|| refused("the notes path is empty"))?;
    let mut cur = root.to_path_buf();
    for part in folders {
        cur.push(part);
        match std::fs::symlink_metadata(&cur) {
            Ok(m) if m.file_type().is_symlink() => {
                return Err(refused(&format!(
                    "{} is a symbolic link; it is not followed",
                    shown(root, &cur)
                )))
            }
            Ok(m) if !m.is_dir() => {
                return Err(refused(&format!("{} is not a folder", shown(root, &cur))))
            }
            Ok(_) => {}
            Err(e) if e.kind() == std::io::ErrorKind::NotFound => {
                std::fs::create_dir(&cur)
                    .map_err(|_| refused(&format!("{} cannot be created", shown(root, &cur))))?;
            }
            Err(_) => {
                return Err(refused(&format!(
                    "{} cannot be inspected",
                    shown(root, &cur)
                )))
            }
        }
    }
    let path = cur.join(file_name);
    match std::fs::symlink_metadata(&path) {
        Ok(m) if m.file_type().is_symlink() => {
            return Err(refused(
                "the notes file is a symbolic link; it is not followed",
            ))
        }
        Ok(m) if !m.is_file() => return Err(refused("the notes file is not a regular file")),
        Ok(_) => {}
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => {}
        Err(_) => return Err(refused("the notes file cannot be inspected")),
    }
    write_line(&path, line)
}

/// The folder as the project-relative path the refusal names.
fn shown(root: &Path, path: &Path) -> String {
    path.strip_prefix(root)
        .unwrap_or(path)
        .to_string_lossy()
        .into_owned()
}

/// Opens the file for the append without following a link and without waiting on a pipe, checks
/// that it is a regular file and that the line fits under the limit, then writes the line (and a
/// newline before it when the file does not end with one) in one call.
#[cfg(unix)]
fn write_line(path: &Path, line: &str) -> Result<u64, String> {
    use std::os::unix::fs::{MetadataExt, OpenOptionsExt};
    let mut file = std::fs::OpenOptions::new()
        .read(true)
        .append(true)
        .create(true)
        .mode(0o600)
        .custom_flags(libc::O_NOFOLLOW | libc::O_NONBLOCK)
        .open(path)
        .map_err(|e| match e.raw_os_error() {
            Some(code) if code == libc::ELOOP => {
                refused("the notes file is a symbolic link; it is not followed")
            }
            _ => refused("the notes file cannot be opened"),
        })?;
    let meta = file
        .metadata()
        .map_err(|_| refused("the notes file cannot be inspected"))?;
    if !meta.is_file() {
        return Err(refused("the notes file is not a regular file"));
    }
    let mut data = Vec::with_capacity(line.len() + 2);
    if meta.size() > 0 && !ends_with_newline(&file)? {
        data.push(b'\n');
    }
    data.extend_from_slice(line.as_bytes());
    data.push(b'\n');
    if !fits(meta.size(), data.len() as u64) {
        return Err(refused("the notes file would pass 1 MiB"));
    }
    file.write_all(&data)
        .map_err(|_| refused("the notes file cannot be written"))?;
    Ok(meta.size() + data.len() as u64)
}

#[cfg(not(unix))]
fn write_line(_path: &Path, _line: &str) -> Result<u64, String> {
    Err(refused("this platform is not supported"))
}

/// Whether the last byte of the file is a newline (read through a second handle at the end).
#[cfg(unix)]
fn ends_with_newline(file: &std::fs::File) -> Result<bool, String> {
    use std::os::unix::fs::FileExt;
    let len = file
        .metadata()
        .map_err(|_| refused("the notes file cannot be inspected"))?
        .len();
    let mut last = [0u8; 1];
    file.read_exact_at(&mut last, len - 1)
        .map_err(|_| refused("the notes file cannot be read"))?;
    Ok(last[0] == b'\n')
}

/// A file of `size` bytes takes `add` more bytes and stays at or under the limit.
fn fits(size: u64, add: u64) -> bool {
    size.checked_add(add)
        .is_some_and(|total| total <= FILE_BYTES)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn improve_note_the_shape_is_held_at_n_and_refused_at_n_plus_one() {
        let ok = |k: &str, key: &str, text: &str, by: &str| check_shape(k, key, text, by);
        assert!(ok(
            "failure",
            &"k".repeat(80),
            &"t".repeat(200),
            &"b".repeat(80)
        )
        .is_ok());
        assert!(ok("failure", &"k".repeat(81), "t", "b").is_err());
        assert!(ok("failure", "k", &"t".repeat(201), "b").is_err());
        assert!(ok("failure", "k", "t", &"b".repeat(81)).is_err());
        assert!(ok("failure", "", "t", "b").is_err());
        assert!(ok("praise", "k", "t", "b").is_err());
        assert!(ok("failure", "k\u{7}", "t", "b").is_err());
        assert!(ok(
            "failure",
            "한".repeat(80).as_str(),
            "글".repeat(200).as_str(),
            "b"
        )
        .is_ok());
    }

    #[test]
    fn improve_note_the_size_limit_is_exact() {
        assert!(
            fits(FILE_BYTES - 10, 10),
            "ending exactly at the limit is accepted"
        );
        assert!(!fits(FILE_BYTES - 10, 11), "one byte over is refused");
        assert!(!fits(FILE_BYTES + 1, 0));
        assert!(!fits(u64::MAX, 1));
    }

    #[test]
    fn improve_note_the_line_is_the_d7_shape_in_order() {
        let line = note_line("2026-11-27T00:00:00Z", "failure", "cargo", "red", "agent");
        assert_eq!(
            line,
            r#"{"v":1,"at":"2026-11-27T00:00:00Z","kind":"failure","key":"cargo","text":"red","by":"agent"}"#
        );
        let v: Value = serde_json::from_str(&line).unwrap();
        assert_eq!(v["v"], json!(1));
    }

    #[test]
    fn improve_note_a_secret_shaped_word_is_masked_before_the_write() {
        let shown = masked("use password=hunter2 now");
        assert!(!shown.contains("hunter2"), "{shown}");
        assert!(shown.contains("[redacted]"), "{shown}");
        assert_eq!(masked("plain words"), "plain words");
    }

    #[cfg(unix)]
    #[test]
    fn improve_note_a_state_folder_link_is_refused_under_either_name() {
        for name in [state_root::MOAI_DIR, moai_cli::PROJECT_STATE_DIR] {
            let dir = tempfile::TempDir::new().unwrap();
            assert!(refuse_state_link(dir.path()).is_ok());
            std::os::unix::fs::symlink(dir.path(), dir.path().join(name)).unwrap();
            let err = refuse_state_link(dir.path()).unwrap_err();
            assert!(err.contains("symbolic link") && err.contains(name), "{err}");
        }
    }

    #[cfg(unix)]
    #[test]
    fn improve_note_a_link_and_a_pipe_are_refused_by_name() {
        let dir = tempfile::TempDir::new().unwrap();
        let target = dir.path().join("t.jsonl");
        std::fs::write(&target, "x\n").unwrap();
        let link = dir.path().join("notes.jsonl");
        std::os::unix::fs::symlink(&target, &link).unwrap();
        let err = write_line(&link, "{}").unwrap_err();
        assert!(err.contains("symbolic link"), "{err}");
        assert_eq!(std::fs::read_to_string(&target).unwrap(), "x\n");
        let err = append(dir.path(), Path::new("nothing/notes.jsonl"), "{}").map(|_| ());
        assert!(err.is_ok(), "an absent folder is created: {err:?}");
    }
}
