//! The name-switch script part of `moai update` (card C-33 S3, SPEC-20261119-update-switch-layer).
//!
//! `path/moai` of the HOME state folder is, on a machine that was switched by hand, a short `sh`
//! script that sends `moai` to the Rust build or, for the words only it has, to the Go build
//! (`moai.old`). The installer leaves a `moai` that is not a link alone, and nothing else writes
//! it, so nobody updates that script. This part previews the step that replaces a known older
//! script (v1, or v2 of card C-33 S2a) with the v3 one built into the program (card C-136: `bundle`
//! is split by action), and applies it only with its own token.
//!
//! Like the home part (`update_home`) it has its own token (`--confirm-switch`): the preview's
//! clock-free content digest; it is not part of the layer's `plan_digest`; a yes at the terminal
//! question never applies it. The apply checks every condition before its first write, keeps
//! the old file as `moai.before-<UTC>`, and replaces `moai` by one rename, so the file is never
//! partial. The revert moves back only a saved copy that is byte-equal to v1 or v2, and only over
//! a `moai` that is still v2 or v3 (a script this update installed). `go/bin/moai` is read and
//! reported, never written.

use super::test_fault;
use serde_json::{json, Value};
use std::fs;
use std::io::{Read, Write};
use std::path::{Path, PathBuf};

type Result<T> = std::result::Result<T, String>;

/// sha256 (hex) of the known v1 switch script, the 1747 B file a person put in `path/moai`.
pub const V1_SHA256: &str = "3c3446a3dbd7d918a7d6465c0765ae66b2dafe3231b9be11cd3181737b6658ac";
/// sha256 (hex) of the v2 script (`preview/moai.v2` of SPEC-20261119-switch-script-v2): a known
/// older script now, and what an earlier build of this update installed.
pub const V2_SHA256: &str = "36dd242bf1337efa9b4759ba7166fb6adea04427800334a374027cc246eb26e4";
/// sha256 (hex) of the v3 script built into the program.
pub const V3_SHA256: &str = "c9a4791683d1c941adef25be281d7846f60638c5dc0443bd113a9f01f0535498";
/// The v3 script (SPEC-20261119-switch-script-v3): `install/moai-switch.sh`.
pub const V3: &[u8] = include_bytes!("../../install/moai-switch.sh");

/// The token prefix, so a token of another part is never taken for this one.
const PREFIX: &str = "switch:";
/// Largest script read: 64 KiB is read, one byte more is "larger".
const MAX_BYTES: usize = 65536;
/// The name pattern the preview shows for the saved copy.
const SAVED_PATTERN: &str = "moai.before-<UTC>";
const NOT_REVIEWED: &str =
    "the switch token differs from the preview's; nothing applied, preview again";

/// Which known older script `path/moai` holds.
#[derive(Clone, Copy, PartialEq, Eq, Debug)]
enum Old {
    V1,
    V2,
}

impl Old {
    fn name(self) -> &'static str {
        match self {
            Old::V1 => "v1",
            Old::V2 => "v2",
        }
    }

    fn sha256(self) -> &'static str {
        match self {
            Old::V1 => V1_SHA256,
            Old::V2 => V2_SHA256,
        }
    }
}

/// What `path/moai` is.
enum Found {
    Absent,
    Link,
    /// A known older script (offered for replacement), with its bytes.
    Older(Old, Vec<u8>),
    /// The v3 script of this program.
    Current,
    /// A file this update does not know, with the reason.
    Other(String),
}

/// The switch part of one `moai update` preview.
#[derive(Debug)]
pub struct Layer {
    /// The part's preview JSON; `None` when there is nothing to say.
    preview: Option<Value>,
    /// The confirming token; `None` unless confirming would write something.
    token: Option<String>,
}

fn link_dir(home: &Path) -> PathBuf {
    crate::home_root::path(home).join("path")
}

/// The folder is a real folder: judged by `lstat`, a link is not followed.
fn real_dir(p: &Path) -> bool {
    fs::symlink_metadata(p).is_ok_and(|m| m.is_dir())
}

fn hex_of(bytes: &[u8]) -> String {
    crate::digest(bytes)
        .strip_prefix("sha256:")
        .unwrap_or_default()
        .to_string()
}

/// The first `MAX_BYTES + 1` bytes of a regular file; `None` when it does not open.
fn read_bounded(p: &Path) -> Option<Vec<u8>> {
    let mut bytes = Vec::new();
    fs::File::open(p)
        .ok()?
        .take(MAX_BYTES as u64 + 1)
        .read_to_end(&mut bytes)
        .ok()?;
    Some(bytes)
}

fn classify(file: &Path) -> Found {
    let Ok(meta) = fs::symlink_metadata(file) else {
        return Found::Absent;
    };
    if meta.file_type().is_symlink() {
        return Found::Link;
    }
    if !meta.is_file() {
        return Found::Other("the path is not a regular file".into());
    }
    let Some(bytes) = read_bounded(file) else {
        return Found::Other(
            "the file could not be read, so it is not known to be the old script".into(),
        );
    };
    if bytes.len() > MAX_BYTES {
        return Found::Other(
            "the file is larger than 64 KiB, so it is not the script this update knows".into(),
        );
    }
    match hex_of(&bytes) {
        h if h == V1_SHA256 => Found::Older(Old::V1, bytes),
        h if h == V2_SHA256 => Found::Older(Old::V2, bytes),
        h if h == V3_SHA256 => Found::Current,
        h => Found::Other(format!(
            "the file is not a script this update knows (sha256 {h})"
        )),
    }
}

/// The token of one preview: sixteen hex digits of the digest of what it is about: the file, the
/// old script it holds (so a v1 preview's token is never a v2 preview's) and the script it gets.
fn token_for(file: &Path, from: Old) -> String {
    let d = hex_of(format!("{}|{}|{V3_SHA256}", file.display(), from.sha256()).as_bytes());
    format!("{PREFIX}{}", &d[..16])
}

/// `go/bin/moai` of the HOME: read, never written.
fn go_bin_report(home: &Path) -> Value {
    let p = home.join("go").join("bin").join("moai");
    let meta = fs::symlink_metadata(&p).ok();
    let exists = meta.is_some();
    let same = meta.as_ref().is_some_and(|m| m.is_file())
        && read_bounded(&p).is_some_and(|b| b.len() <= MAX_BYTES && hex_of(&b) == V1_SHA256);
    let note = match (exists, same) {
        (false, _) => "no file there; nothing to do".to_string(),
        (true, true) => format!(
            "{} holds the same old script. `moai update` does not touch it; once the preview has been applied, copy the new script there yourself if that name should follow",
            p.display()
        ),
        (true, false) => format!(
            "{} is a different file; `moai update` leaves it as it is",
            p.display()
        ),
    };
    json!({
        "path": p.display().to_string(),
        "exists": exists,
        "same_as_v1": same,
        "note": note,
    })
}

fn notes() -> Value {
    json!([
        "Projects the Go build set up call `moai hook ...` from their hook files; after the switch `moai hook` still goes to the Go build (moai.old). Nothing in those projects is changed.",
        "Their `.moai/status_line.sh` runs `moai statusline`; after the switch that is the Rust three-line render, not the Go one. Nothing in those projects is changed.",
        "`moai bundle add`, `remove` and `list` (the opt-in skill bundles) still go to the Go build; every other `moai bundle` action (`ready`, `check`, `body`, `build`, `verify`) reaches the Rust M-Flow bundle."
    ])
}

/// The switch preview of `home`. Writes nothing.
pub fn plan(home: &Path) -> Result<Layer> {
    let none = Layer {
        preview: None,
        token: None,
    };
    let dir = link_dir(home);
    if !real_dir(&dir) {
        return Ok(none);
    }
    let file = dir.join("moai");
    let (action, reason, token, from) = match classify(&file) {
        Found::Absent | Found::Link | Found::Current => return Ok(none),
        Found::Older(old, _) => ("replace", None, Some(token_for(&file, old)), Some(old)),
        Found::Other(why) => ("left", Some(why), None, None),
    };
    let summary_line = match (&token, &reason, from) {
        (Some(t), _, Some(old)) => {
            let what = match old {
                Old::V1 => "statusline to Rust; spec, session and bundle by action",
                Old::V2 => "bundle by action: ready, check, body, build and verify to Rust; add, remove and list stay Go",
            };
            format!(
                "{} -> v3 ({what}) · the old file is kept as {SAVED_PATTERN} · confirm with: moai update --confirm-switch {t}",
                old.name()
            )
        }
        (_, why, _) => format!("left as it is: {}", why.as_deref().unwrap_or("")),
    };
    let preview = json!({
        "action": action,
        "file": file.display().to_string(),
        "saved_as": SAVED_PATTERN,
        "from": from.map(Old::name),
        "v1_sha256": V1_SHA256,
        "v2_sha256": V2_SHA256,
        "v3_sha256": V3_SHA256,
        "reason": reason,
        "summary": summary_line,
        "confirm_token": token,
        "apply_flag": token.as_ref().map(|t| format!("--confirm-switch {t}")),
        "revert_with": "moai-cli update --revert-switch <stamp>",
        "go_bin": go_bin_report(home),
        "notes": notes(),
        "original_kept": true,
        "writes": false,
    });
    Ok(Layer {
        preview: Some(preview),
        token,
    })
}

impl Layer {
    /// The part's preview, when there is one.
    pub fn to_json(&self) -> Option<&Value> {
        self.preview.as_ref()
    }

    /// The token `--confirm-switch` takes, when confirming would write something.
    pub fn token(&self) -> Option<&str> {
        self.token.as_deref()
    }

    /// Whether confirming would write a file (the preview offers `replace`).
    pub fn writes(&self) -> bool {
        self.token.is_some()
    }

    /// The terminal question's lines for this part; empty when there is no part.
    pub fn summary(&self) -> String {
        match &self.preview {
            Some(p) => format!(
                "moai name-switch script\n  {}\n",
                p["summary"].as_str().unwrap_or("")
            ),
            None => String::new(),
        }
    }
}

/// Refuse a `--confirm-switch` token that is not the preview's, before anything is written.
pub fn check_token<'a>(layer: &Layer, given: Option<&'a str>) -> Result<Option<&'a str>> {
    let Some(t) = given else {
        return Ok(None);
    };
    if layer.token() != Some(t) {
        return Err(NOT_REVIEWED.into());
    }
    Ok(Some(t))
}

/// `YYYY-MM-DDTHH:MM:SSZ` -> `YYYYMMDDTHHMMSSZ`, or `None` when `now` is not that shape.
fn compact_utc(now: &str) -> Option<String> {
    let b = now.as_bytes();
    let shape = b.len() == 20
        && b.iter().enumerate().all(|(i, c)| match i {
            4 | 7 => *c == b'-',
            10 => *c == b'T',
            13 | 16 => *c == b':',
            19 => *c == b'Z',
            _ => c.is_ascii_digit(),
        });
    shape.then(|| now.chars().filter(|c| c.is_ascii_alphanumeric()).collect())
}

/// `YYYYMMDDTHHMMSSZ`: the stamp a saved copy carries.
fn stamp_ok(stamp: &str) -> bool {
    let b = stamp.as_bytes();
    b.len() == 16
        && b.iter().enumerate().all(|(i, c)| match i {
            8 => *c == b'T',
            15 => *c == b'Z',
            _ => c.is_ascii_digit(),
        })
}

#[cfg(unix)]
fn mode_of(meta: &fs::Metadata) -> u32 {
    use std::os::unix::fs::PermissionsExt;
    meta.permissions().mode() & 0o777
}

#[cfg(not(unix))]
fn mode_of(_meta: &fs::Metadata) -> u32 {
    0o755
}

/// A new file that must not exist yet, written whole, mode set exactly (the umask is not
/// trusted), flushed to disk.
fn create_new(path: &Path, bytes: &[u8], mode: u32) -> std::io::Result<()> {
    let mut f = fs::OpenOptions::new()
        .write(true)
        .create_new(true)
        .open(path)?;
    f.write_all(bytes)?;
    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt;
        f.set_permissions(fs::Permissions::from_mode(mode))?;
    }
    #[cfg(not(unix))]
    let _ = mode;
    f.sync_all()
}

fn absent(p: &Path) -> bool {
    matches!(fs::symlink_metadata(p), Err(e) if e.kind() == std::io::ErrorKind::NotFound)
}

fn sync_dir(dir: &Path) {
    if let Ok(d) = fs::File::open(dir) {
        let _ = d.sync_all();
    }
}

/// Replace the previewed script (the confirmed `token`); `now` (`YYYY-MM-DDTHH:MM:SSZ`) names
/// the saved copy. The preview is computed again and must give the same token; every condition
/// is checked before the first write; a failed step leaves `moai` whole.
pub fn apply(home: &Path, token: &str, now: &str) -> Result<Value> {
    let stamp = compact_utc(now)
        .ok_or("the clock is not a UTC time (YYYY-MM-DDTHH:MM:SSZ); nothing written")?;
    let fresh = plan(home)?;
    if fresh.token() != Some(token) {
        return Err(NOT_REVIEWED.into());
    }
    let dir = link_dir(home);
    let file = dir.join("moai");
    let saved = dir.join(format!("moai.before-{stamp}"));
    let temp = dir.join(format!(".moai.switch-{stamp}"));
    // Every check, before any write.
    if !real_dir(&dir) {
        return Err(format!(
            "{} is a link or not a folder; nothing written",
            dir.display()
        ));
    }
    let Found::Older(old, old_bytes) = classify(&file) else {
        return Err(format!(
            "{} is not a known older script any more; nothing written, preview again",
            file.display()
        ));
    };
    for p in [&saved, &temp] {
        if !absent(p) {
            return Err(format!("{} already exists; nothing written", p.display()));
        }
    }
    let mode = fs::symlink_metadata(&file)
        .map(|m| mode_of(&m))
        .map_err(|e| format!("{}: {e}", file.display()))?;
    // The first write: the old script, kept byte for byte.
    create_new(&saved, &old_bytes, mode)
        .map_err(|e| format!("saving {}: {e}; nothing replaced", saved.display()))?;
    if read_bounded(&saved).map(|b| hex_of(&b)).as_deref() != Some(old.sha256()) {
        let _ = fs::remove_file(&saved);
        return Err(format!(
            "the saved copy {} did not read back as {}; nothing replaced",
            saved.display(),
            old.name()
        ));
    }
    let after_first = |e: String| {
        format!(
            "{e}; `moai` is unchanged, the old script is saved as {}",
            saved.display()
        )
    };
    test_fault("switch-after-save").map_err(after_first)?;
    if let Err(e) = create_new(&temp, V3, 0o755) {
        let _ = fs::remove_file(&temp);
        return Err(after_first(format!("writing {}: {e}", temp.display())));
    }
    // Still the script the preview was about, right before the replace.
    if !matches!(classify(&file), Found::Older(o, _) if o == old) {
        let _ = fs::remove_file(&temp);
        return Err(after_first(
            "`moai` changed while the apply ran, so it was not replaced".into(),
        ));
    }
    if let Err(e) = test_fault("switch-rename").and_then(|()| {
        fs::rename(&temp, &file).map_err(|e| format!("replacing {}: {e}", file.display()))
    }) {
        let _ = fs::remove_file(&temp);
        return Err(after_first(e));
    }
    sync_dir(&dir);
    if !matches!(classify(&file), Found::Current) {
        return Err(format!(
            "{} did not read back as v3 after the replace; the old script is saved as {}",
            file.display(),
            saved.display()
        ));
    }
    Ok(json!({
        "state": "applied",
        "file": file.display().to_string(),
        "stamp": stamp,
        "saved": saved.display().to_string(),
        "from": old.name(),
        "v3_sha256": V3_SHA256,
        "revert": format!("moai-cli update --revert-switch {stamp}"),
        "revert_plain": format!("mv \"{}\" \"{}\"", saved.display(), file.display()),
    }))
}

/// Put the saved copy `moai.before-<stamp>` back: only a copy byte-equal to v1, only over a
/// `moai` that is still v2, by one rename. Every other case changes nothing and says why.
pub fn revert(home: &Path, stamp: &str) -> Result<Value> {
    if !stamp_ok(stamp) {
        return Err("the stamp must look like 20261119T074722Z; nothing changed".into());
    }
    let dir = link_dir(home);
    let file = dir.join("moai");
    let saved = dir.join(format!("moai.before-{stamp}"));
    if !real_dir(&dir) {
        return Err(format!(
            "{} is a link or not a folder; nothing changed",
            dir.display()
        ));
    }
    match fs::symlink_metadata(&saved) {
        Err(_) => {
            return Err(format!(
                "no saved copy {} here; nothing changed",
                saved.display()
            ))
        }
        Ok(m) if !m.is_file() => {
            return Err(format!(
                "{} is a link or not a regular file; left as it is",
                saved.display()
            ))
        }
        Ok(_) => {}
    }
    let saved_hash = read_bounded(&saved).map(|b| hex_of(&b));
    let Some(restored) = [Old::V1, Old::V2]
        .into_iter()
        .find(|o| saved_hash.as_deref() == Some(o.sha256()))
    else {
        return Err(format!(
            "{} is not the v1 or v2 script; left as it is, nothing changed",
            saved.display()
        ));
    };
    if !matches!(classify(&file), Found::Current | Found::Older(Old::V2, _)) {
        return Err(format!(
            "{} is not a v2 or v3 script this update installs; left as it is, the saved copy stays",
            file.display()
        ));
    }
    test_fault("switch-revert-rename")?;
    fs::rename(&saved, &file).map_err(|e| format!("putting {} back: {e}", saved.display()))?;
    sync_dir(&dir);
    if read_bounded(&file).map(|b| hex_of(&b)).as_deref() != Some(restored.sha256()) {
        return Err(format!(
            "{} did not read back as {} after the revert",
            file.display(),
            restored.name()
        ));
    }
    Ok(json!({
        "state": "reverted",
        "file": file.display().to_string(),
        "restored_from": saved.display().to_string(),
        "restored": restored.name(),
        "v1_sha256": V1_SHA256,
        "v2_sha256": V2_SHA256,
    }))
}
