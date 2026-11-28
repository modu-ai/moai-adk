//! The name-switch script part of `moai update` (SPEC-20261121-go-path-removal, revision 6).
//!
//! `path/moai` of the HOME state folder is, on a machine the Go generation set up, a short `sh`
//! script that sends `moai` to the Rust build or, for the words only it has, to the Go build
//! (`moai.old`). The installer lays a link of that name and leaves a plain file alone with a
//! warning, so nobody ends such a script. This part ends the two scripts it knows, v1 and v2, by
//! their sha256: the confirmed run keeps the old script as `moai.before-<UTC>` and puts a
//! relative link `moai` to the running build's `bin/moai` in its place
//! (`../versions/<v>/bin/moai`, SPEC-20261127-installed-binary-name-moai D-4).
//!
//! Like the home part (`update_home`) it has its own token (`--confirm-switch`): the preview's
//! clock-free content digest; it is not part of the layer's `plan_digest`; a yes at the terminal
//! question never applies it. The link is offered only while the `bin/moai` of the build folder
//! the running binary sits in answers `--version` with the Rust signature; a build outside a
//! `versions/<v>/bin/` folder of the state folder, or one with no `bin/moai`, keeps the step held
//! with a sentence naming the file it looked for. The apply checks every condition before its
//! first write and replaces the script by one rename, so the name never goes missing. The revert
//! moves back only a kept copy that is byte-equal to a known script, and only over a link this
//! step made (the shape above, or the earlier `moai-cli` one). `go/bin/moai` is read and
//! reported, never written.

use super::test_fault;
use serde_json::{json, Value};
use std::fs;
use std::io::{Read, Write};
use std::path::{Path, PathBuf};

type Result<T> = std::result::Result<T, String>;

/// sha256 (hex) of the known v1 switch script, the 1747 B file a person put in `path/moai`.
pub const V1_SHA256: &str = "3c3446a3dbd7d918a7d6465c0765ae66b2dafe3231b9be11cd3181737b6658ac";
/// sha256 (hex) of the v2 switch script, which the program once carried and wrote.
pub const V2_SHA256: &str = "36dd242bf1337efa9b4759ba7166fb6adea04427800334a374027cc246eb26e4";
/// The launcher the confirmed run links `moai` to: the `bin/moai` of the running build's folder.
const LAUNCHER: &str = "moai";
/// The link an earlier version of this step made, relative to the script's folder: only the
/// revert still reads it.
const EARLIER_TARGET: &str = "moai-cli";

/// The token prefix, so a token of another part is never taken for this one.
const PREFIX: &str = "switch:";
/// Largest script read: 64 KiB is read, one byte more is "larger".
const MAX_BYTES: usize = 65536;
/// The name pattern the preview shows for the kept copy.
const SAVED_PATTERN: &str = "moai.before-<UTC>";
const NOT_REVIEWED: &str =
    "the switch token differs from the preview's; nothing applied, preview again";
const NO_LINK_HERE: &str =
    "this step makes a relative symbolic link, which is not made on this platform; nothing written";

/// A script this update knows: its name in previews and results, its sha256 and its bytes.
#[derive(Debug)]
struct Known {
    name: &'static str,
    sha: &'static str,
    bytes: Vec<u8>,
}

/// What `path/moai` is.
#[derive(Debug)]
enum Found {
    Absent,
    Link,
    Known(Known),
    /// A file this update does not know, with the reason.
    Other(String),
}

/// What the launcher beside the script is.
#[derive(Debug)]
enum Launcher {
    Missing,
    /// It started but printed no first line.
    Unprobed,
    /// The first line of its `--version`, which is not the Rust signature.
    NotRust(String),
    /// The first line of its `--version`, the Rust signature.
    Rust(String),
}

impl Launcher {
    fn from_line(line: Option<String>) -> Launcher {
        match line {
            Some(line) if crate::install::generation_of(Some(&line)) == "rust" => {
                Launcher::Rust(line)
            }
            Some(line) => Launcher::NotRust(line),
            None => Launcher::Unprobed,
        }
    }

    fn line(&self) -> Option<&str> {
        match self {
            Launcher::Rust(line) | Launcher::NotRust(line) => Some(line),
            Launcher::Missing | Launcher::Unprobed => None,
        }
    }

    fn is_rust(&self) -> bool {
        matches!(self, Launcher::Rust(_))
    }
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

/// The name and the sha256 of the known script these bytes are.
fn known_of(bytes: &[u8]) -> Option<(&'static str, &'static str)> {
    match hex_of(bytes).as_str() {
        h if h == V1_SHA256 => Some(("v1", V1_SHA256)),
        h if h == V2_SHA256 => Some(("v2", V2_SHA256)),
        _ => None,
    }
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
    match known_of(&bytes) {
        Some((name, sha)) => Found::Known(Known { name, sha, bytes }),
        None => Found::Other(format!(
            "the file is not a script this update knows (sha256 {})",
            hex_of(&bytes)
        )),
    }
}

/// Whether `path/moai` is a symbolic link whose target is exactly `target`.
fn is_link_to(file: &Path, target: &str) -> bool {
    fs::symlink_metadata(file).is_ok_and(|m| m.file_type().is_symlink())
        && fs::read_link(file).is_ok_and(|t| t == Path::new(target))
}

/// Whether `target` has a shape this step makes: `../versions/<v>/bin/moai`, or the earlier
/// `moai-cli`.
fn is_switch_target(target: &Path) -> bool {
    use std::path::Component::{Normal, ParentDir};
    if target == Path::new(EARLIER_TARGET) {
        return true;
    }
    let mut parts = target.components();
    matches!(
        (parts.next(), parts.next(), parts.next(), parts.next(), parts.next(), parts.next()),
        (Some(ParentDir), Some(Normal(versions)), Some(Normal(_)), Some(Normal(bin)), Some(Normal(name)), None)
            if versions == "versions" && bin == "bin" && name == LAUNCHER
    )
}

/// Whether `path/moai` is a link this step made, in either shape (the revert's test).
fn is_a_switch_link(file: &Path) -> bool {
    fs::symlink_metadata(file).is_ok_and(|m| m.file_type().is_symlink())
        && fs::read_link(file).is_ok_and(|t| is_switch_target(&t))
}

/// The running build as the step sees it: the launcher file it looks for and, when the binary
/// sits in `<state folder>/versions/<v>/bin/`, the relative link target to that launcher.
struct Build {
    launcher: PathBuf,
    target: Option<String>,
}

/// `p` resolved through links; a file that does not exist (yet) is resolved by its folder.
fn resolved(p: &Path) -> PathBuf {
    if let Ok(c) = fs::canonicalize(p) {
        return c;
    }
    match (
        p.parent().and_then(|d| fs::canonicalize(d).ok()),
        p.file_name(),
    ) {
        (Some(d), Some(n)) => d.join(n),
        _ => p.to_path_buf(),
    }
}

fn target_of(home: &Path, bin: &Path) -> Option<String> {
    let canonical_home = fs::canonicalize(home).ok()?;
    let versions = crate::bin_names::versions_dir_of(&canonical_home, bin)?;
    // The link lies in `<state folder>/path/`, so `..` must reach the folder that holds `versions`.
    let state = fs::canonicalize(crate::home_root::path(home)).ok()?;
    if versions.parent()? != state {
        return None;
    }
    let folder = bin.parent()?.file_name()?.to_str()?;
    Some(format!("../versions/{folder}/bin/{LAUNCHER}"))
}

fn build_of(home: &Path, exe: &Path) -> Build {
    let exe = resolved(exe);
    let bin = exe.parent().map(Path::to_path_buf).unwrap_or_default();
    Build {
        target: target_of(home, &bin),
        launcher: bin.join(LAUNCHER),
    }
}

/// The sentence for a held step: it names the launcher file it looked for.
fn held_reason(build: &Build) -> String {
    match build.target {
        Some(_) => format!(
            "the launcher `{}` (the bin/moai of the running build) is missing or does not answer `--version` with the Rust signature, so no link is offered; lay the launcher with the installer, then preview again",
            build.launcher.display()
        ),
        None => format!(
            "the running build is not in a `versions/<v>/bin/` folder of this state folder, so no relative link to `{}` can be made; run `moai update` from the installed build, then preview again",
            build.launcher.display()
        ),
    }
}

/// What the launcher `bin/moai` of the running build is: it must be a regular file, or a link
/// that resolves to one, with an execute bit, and it is then started once with `--version` by
/// the safe probe of `install` (cleared environment but `HOME`, a fixed `PATH`, a time limit). A
/// dangling link and a link to a folder read as missing, because `metadata` follows the link.
fn launcher_state(path: &Path, home: &Path) -> Launcher {
    let Ok(meta) = fs::metadata(path) else {
        return Launcher::Missing;
    };
    if !meta.is_file() {
        return Launcher::Missing;
    }
    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt as _;
        if meta.permissions().mode() & 0o111 == 0 {
            return Launcher::Missing;
        }
    }
    Launcher::from_line(crate::install::probe_version_env(path, &[("HOME", home)]))
}

/// The token of one preview: sixteen hex digits of the digest of what it is about (the script,
/// the launcher, the link target and the line the launcher answered).
fn token_for(file: &Path, script_sha: &str, launcher: &Path, target: &str, line: &str) -> String {
    let d = hex_of(
        format!(
            "{}|{script_sha}|{}|{target}|{line}",
            file.display(),
            launcher.display()
        )
        .as_bytes(),
    );
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
        "Projects the Go build set up call `moai hook ...` from their hook files; through the link that call reaches the Rust build, which answers exit 0 and runs no Go hook. Nothing in those projects is changed.",
        "Their `.moai/status_line.sh` runs `moai statusline`; through the link that is the Rust render. Nothing in those projects is changed."
    ])
}

/// The switch preview of `home` for the running binary. Writes nothing.
pub fn plan(home: &Path) -> Result<Layer> {
    let exe = std::env::current_exe().map_err(|_| "current executable unavailable".to_string())?;
    plan_for(home, &exe)
}

/// The switch preview of `home` for the build whose binary is `exe`. Writes nothing.
pub fn plan_for(home: &Path, exe: &Path) -> Result<Layer> {
    let none = Layer {
        preview: None,
        token: None,
    };
    let dir = link_dir(home);
    if !real_dir(&dir) {
        return Ok(none);
    }
    let file = dir.join("moai");
    let build = build_of(home, exe);
    let found = classify(&file);
    // A file this update does not know is never replaced, so nothing beside it is started.
    let (known, launcher) = match &found {
        Found::Absent | Found::Link => return Ok(none),
        Found::Known(known) => (Some(known), Some(launcher_state(&build.launcher, home))),
        Found::Other(why) => {
            let _ = why;
            (None, None)
        }
    };
    let launcher_line = launcher.as_ref().and_then(|l| l.line().map(str::to_string));
    let (action, reason, token) = match (known, launcher.as_ref(), build.target.as_deref()) {
        (Some(known), Some(state), Some(target)) if state.is_rust() => (
            "link",
            None,
            Some(token_for(
                &file,
                known.sha,
                &build.launcher,
                target,
                state.line().unwrap_or_default(),
            )),
        ),
        (Some(_), Some(_), _) => ("held", Some(held_reason(&build)), None),
        _ => ("left", why_of(&found), None),
    };
    let target_shown = build.target.as_deref().unwrap_or_default();
    let summary_line = match (&token, &reason, known) {
        (Some(t), _, Some(known)) => format!(
            "{} script -> a link to the Rust launcher `{target_shown}` · the old script is kept as {SAVED_PATTERN} · confirm with: moai update --confirm-switch {t}",
            known.name
        ),
        (None, Some(why), _) if action == "held" => format!("waits: {why}"),
        (None, Some(why), _) => format!("left as it is: {why}"),
        _ => String::new(),
    };
    let preview = json!({
        "action": action,
        "file": file.display().to_string(),
        "script": known.map(|k| k.name),
        "script_sha256": known.map(|k| k.sha),
        "launcher": launcher.as_ref().map(|_| build.launcher.display().to_string()),
        "launcher_version_line": launcher_line,
        "link_target": build.target,
        "saved_as": SAVED_PATTERN,
        "reason": reason,
        "summary": summary_line,
        "confirm_token": token,
        "apply_flag": token.as_ref().map(|t| format!("--confirm-switch {t}")),
        "revert_with": "moai update --revert-switch <stamp>",
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

/// Why a file this update does not know is left alone.
fn why_of(found: &Found) -> Option<String> {
    match found {
        Found::Other(why) => Some(why.clone()),
        _ => None,
    }
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

    /// Whether confirming would write a file (the preview offers `link`).
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

/// `YYYYMMDDTHHMMSSZ`: the stamp a kept copy carries.
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

/// Whether this platform makes the link this step is about.
fn links_here() -> bool {
    cfg!(unix)
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

#[cfg(unix)]
/// The link, made under a name that must not exist yet.
fn make_link(target: &str, temp: &Path) -> std::io::Result<()> {
    std::os::unix::fs::symlink(target, temp)
}

#[cfg(not(unix))]
/// The backstop behind [`links_here`]: never reached, and it writes nothing if it is.
fn make_link(_target: &str, _temp: &Path) -> std::io::Result<()> {
    Err(std::io::Error::other(NO_LINK_HERE))
}

fn absent(p: &Path) -> bool {
    matches!(fs::symlink_metadata(p), Err(e) if e.kind() == std::io::ErrorKind::NotFound)
}

fn sync_dir(dir: &Path) {
    if let Ok(d) = fs::File::open(dir) {
        let _ = d.sync_all();
    }
}

/// End the previewed script (the confirmed `token`); `now` (`YYYY-MM-DDTHH:MM:SSZ`) names the
/// kept copy. The preview is computed again and must give the same token; every condition is
/// checked before the first write; a failed step leaves the script whole.
pub fn apply(home: &Path, token: &str, now: &str) -> Result<Value> {
    let exe = std::env::current_exe().map_err(|_| "current executable unavailable".to_string())?;
    apply_for(home, &exe, token, now)
}

/// [`apply`] for the build whose binary is `exe`.
pub fn apply_for(home: &Path, exe: &Path, token: &str, now: &str) -> Result<Value> {
    if !links_here() {
        return Err(NO_LINK_HERE.into());
    }
    let stamp = compact_utc(now)
        .ok_or("the clock is not a UTC time (YYYY-MM-DDTHH:MM:SSZ); nothing written")?;
    let fresh = plan_for(home, exe)?;
    if fresh.token() != Some(token) {
        return Err(NOT_REVIEWED.into());
    }
    let build = build_of(home, exe);
    let target = build.target.clone().ok_or(NOT_REVIEWED)?;
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
    let Found::Known(script) = classify(&file) else {
        return Err(format!(
            "{} is not a script this update knows any more; nothing written, preview again",
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
    create_new(&saved, &script.bytes, mode)
        .map_err(|e| format!("saving {}: {e}; nothing replaced", saved.display()))?;
    if read_bounded(&saved).map(|b| hex_of(&b)).as_deref() != Some(script.sha) {
        let _ = fs::remove_file(&saved);
        return Err(format!(
            "the kept copy {} did not read back as the {} script; nothing replaced",
            saved.display(),
            script.name
        ));
    }
    let after_first = |e: String| {
        format!(
            "{e}; `moai` is unchanged, the old script is kept as {}",
            saved.display()
        )
    };
    test_fault("switch-after-save").map_err(after_first)?;
    if let Err(e) = make_link(&target, &temp) {
        let _ = fs::remove_file(&temp);
        return Err(after_first(format!("linking {}: {e}", temp.display())));
    }
    // Still the script the preview was about, right before the replace.
    if !matches!(classify(&file), Found::Known(again) if again.sha == script.sha) {
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
    if !is_link_to(&file, &target) {
        return Err(format!(
            "{} did not read back as the link this step makes; the old script is kept as {}",
            file.display(),
            saved.display()
        ));
    }
    Ok(json!({
        "state": "applied",
        "file": file.display().to_string(),
        "stamp": stamp,
        "saved": saved.display().to_string(),
        "script": script.name,
        "script_sha256": script.sha,
        "link": target,
        "launcher": build.launcher.display().to_string(),
        "revert": format!("moai update --revert-switch {stamp}"),
        "revert_plain": format!("mv \"{}\" \"{}\"", saved.display(), file.display()),
    }))
}

/// Put the kept copy `moai.before-<stamp>` back: only a regular file byte-equal to a known
/// script, only over the link this step makes, by one rename. Every other case changes nothing
/// and says why.
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
    let Some(bytes) = read_bounded(&saved) else {
        return Err(format!(
            "{} could not be read; left as it is, nothing changed",
            saved.display()
        ));
    };
    let Some((name, sha)) = known_of(&bytes) else {
        return Err(format!(
            "{} is not a known script; left as it is, nothing changed",
            saved.display()
        ));
    };
    if !is_a_switch_link(&file) {
        return Err(format!(
            "{} is not the link this step makes to a build's `bin/{LAUNCHER}`; left as it is, the kept copy stays",
            file.display()
        ));
    }
    test_fault("switch-revert-rename")?;
    fs::rename(&saved, &file).map_err(|e| format!("putting {} back: {e}", saved.display()))?;
    sync_dir(&dir);
    if !matches!(classify(&file), Found::Known(back) if back.sha == sha) {
        return Err(format!(
            "{} did not read back as the {name} script after the revert",
            file.display()
        ));
    }
    Ok(json!({
        "state": "reverted",
        "file": file.display().to_string(),
        "restored_from": saved.display().to_string(),
        "script": name,
        "script_sha256": sha,
    }))
}
