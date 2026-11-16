//! The `moai` link in `<state-home>/path/` (#3004 P6-1): a read-only plan.
//!
//! `install/install.sh` links `<state-home>/path/moai` to `<state-home>/versions/<v>/bin/moai`
//! (the name the user types, #3004 decision "one name: moai"). This module only *reports* what
//! `moai update` would do about that link; it writes nothing, so a second piece can apply it.
//! The running build's `bin/moai` is the target (the same path the installer links), and the
//! link is judged by what is on disk, never by the fact that something is there:
//!
//! * `create` — no link, and the target is a Rust executable;
//! * `unchanged` — the link already reaches that target;
//! * `repoint` — the link reaches another `versions/<other>/bin/moai` of this installer;
//! * `blocked` — the target is missing, not a regular file, not executable, a `#!` script, or
//!   does not say `(rust)` for `--version` (a link is not proof of Rust, so a link to a Go
//!   binary is `blocked`, never `unchanged`);
//! * `foreign` — the link path holds a regular file or folder (a shim, say), or a link that does
//!   not reach `versions/<v>/bin/moai` under the same state home; it is left as it is;
//! * `outside_layout` — the running binary is not in `<state-home>/versions/<v>/bin/`;
//! * `unsupported` — Windows: `install.ps1` copies the binary instead of linking.
//!
//! No link is followed out of the state home: a link text is resolved (a relative one against
//! the link's folder) and, when it exists, canonicalized before it is compared with the
//! canonical `versions/` folder.

use crate::install::{generation_of, probe_version};
use serde_json::{json, Value};
use std::fs;
use std::io::Read as _;
use std::path::{Component, Path, PathBuf};

type Probe<'a> = &'a dyn Fn(&Path) -> Option<String>;

/// What the plan found; nothing here is applied by this piece.
#[derive(Debug, Clone)]
pub struct Plan {
    /// `create`, `unchanged`, `repoint`, `blocked`, `foreign`, `outside_layout` or `unsupported`.
    pub state: &'static str,
    /// Why nothing would be written, for every state but `create`, `unchanged` and `repoint`.
    pub reason: Option<String>,
    /// `<state-home>/path/moai`.
    pub link: Option<PathBuf>,
    /// `bin/moai` of the version folder this build runs from: where the link should point.
    pub target: Option<PathBuf>,
    /// The text the link holds now, when the link path is a link.
    pub current: Option<String>,
}

impl Plan {
    pub fn to_json(&self) -> Value {
        let text = |p: &Option<PathBuf>| p.as_ref().map(|p| p.to_string_lossy().into_owned());
        json!({
            "state": self.state,
            "reason": self.reason,
            "link": text(&self.link),
            "target": text(&self.target),
            "current": self.current,
        })
    }
}

fn plan(
    state: &'static str,
    reason: Option<String>,
    link: Option<PathBuf>,
    target: Option<PathBuf>,
    current: Option<String>,
) -> Plan {
    Plan {
        state,
        reason,
        link,
        target,
        current,
    }
}

/// Preview the `moai` link for the executable `exe` (the running binary in the CLI). Reads
/// only; the target is run once with `--version` under the version probe's cleared environment.
pub fn plan_for(home: &Path, exe: &Path) -> Plan {
    plan_with(home, exe, cfg!(windows), &probe_version)
}

fn moai_name() -> String {
    format!("moai{}", std::env::consts::EXE_SUFFIX)
}

fn plan_with(home: &Path, exe: &Path, windows: bool, probe: Probe) -> Plan {
    if windows {
        return plan(
            "unsupported",
            Some(
                "on Windows the installer copies moai instead of linking it; nothing is linked"
                    .into(),
            ),
            None,
            None,
            None,
        );
    }
    let Some(bin) = crate::bin_names::layout_bin(home, exe) else {
        return plan(
            "outside_layout",
            Some(format!(
                "this binary is not in the version folder layout ~/{}/versions/<version>/bin/; nothing is linked",
                crate::home_root::dir(home)
            )),
            None,
            None,
            None,
        );
    };
    // `layout_bin` canonicalized the home, so the state home below is built from the same root.
    let canonical_home = fs::canonicalize(home).unwrap_or_else(|_| home.to_path_buf());
    let state = crate::home_root::path(&canonical_home);
    let link_dir = state.join("path");
    let link = link_dir.join(moai_name());
    let target = bin.join(moai_name());
    let found = |state, reason, current| {
        plan(
            state,
            reason,
            Some(link.clone()),
            Some(target.clone()),
            current,
        )
    };

    if fs::symlink_metadata(&link_dir).is_ok_and(|m| m.file_type().is_symlink()) {
        return found(
            "foreign",
            Some("path/ is a link; it is left as it is".into()),
            None,
        );
    }
    let (kind, current) = match fs::symlink_metadata(&link) {
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => (Kind::Absent, None),
        Err(e) => {
            return found(
                "foreign",
                Some(format!("path/moai cannot be inspected ({})", e.kind())),
                None,
            )
        }
        Ok(m) if m.file_type().is_symlink() => {
            let Ok(text) = fs::read_link(&link) else {
                return found(
                    "foreign",
                    Some("the link path/moai cannot be read".into()),
                    None,
                );
            };
            let kind = classify_link(
                &link_dir,
                &text,
                &state.join("versions"),
                &crate::home_root::path(home).join("versions"),
                &target,
            );
            (kind, Some(text.display().to_string()))
        }
        Ok(_) => {
            return found(
                "foreign",
                Some("path/moai is a file or folder, not a link made by the installer; it is left as it is".into()),
                None,
            )
        }
    };
    if let Kind::Foreign(reason) = kind {
        return found("foreign", Some(reason), current);
    }
    // A link is not proof of Rust: the target is judged whatever the link says.
    if let Err(reason) = rust_executable(&target, probe) {
        return found("blocked", Some(reason), current);
    }
    let state = match kind {
        Kind::Absent => "create",
        Kind::Ours => "unchanged",
        _ => "repoint",
    };
    found(state, None, current)
}

/// What an existing link reaches.
enum Kind {
    Absent,
    /// The link reaches this build's `bin/moai`.
    Ours,
    /// The link reaches another version folder's `bin/moai` under the same state home.
    OtherVersion,
    Foreign(String),
}

/// A path with `.` and `..` folded away by name, without touching the file system.
fn lexical(path: &Path) -> PathBuf {
    let mut out = PathBuf::new();
    for c in path.components() {
        match c {
            Component::CurDir => {}
            Component::ParentDir => {
                if !out.pop() {
                    out.push("..");
                }
            }
            other => out.push(other.as_os_str()),
        }
    }
    out
}

/// Judge a link by where its text leads. `canonical_versions` is `<canonical state>/versions`,
/// `plain_versions` the same folder spelled from the HOME the caller gave (the installer writes
/// links from `$HOME`, which may not be canonical); `target` is this build's `bin/moai`.
fn classify_link(
    link_dir: &Path,
    text: &Path,
    canonical_versions: &Path,
    plain_versions: &Path,
    target: &Path,
) -> Kind {
    let resolved = lexical(&if text.is_absolute() {
        text.to_path_buf()
    } else {
        link_dir.join(text)
    });
    // An existing path is judged by where it really is: a version folder that is itself a link
    // out of the state home must not pass.
    let real = fs::canonicalize(&resolved).ok();
    let rel = match &real {
        Some(r) => r.strip_prefix(canonical_versions).ok(),
        None => resolved
            .strip_prefix(canonical_versions)
            .ok()
            .or_else(|| resolved.strip_prefix(plain_versions).ok()),
    };
    let Some(rel) = rel else {
        return Kind::Foreign(format!(
            "path/moai points to {}, outside the version folders of this installer; it is left as it is",
            text.display()
        ));
    };
    let parts: Vec<_> = rel.components().collect();
    let is_bin_moai = parts.len() == 3
        && parts[1].as_os_str() == "bin"
        && parts[2].as_os_str() == moai_name().as_str();
    if !is_bin_moai {
        return Kind::Foreign(format!(
            "path/moai points to {}, which is not versions/<version>/bin/moai; it is left as it is",
            text.display()
        ));
    }
    if real.as_deref().unwrap_or(&resolved) == target {
        Kind::Ours
    } else {
        Kind::OtherVersion
    }
}

/// `target` is a regular file (not a link), executable, not a `#!` script, and its `--version`
/// first line is the Rust signature. The `Err` is the reason it is not.
fn rust_executable(target: &Path, probe: Probe) -> Result<(), String> {
    let meta = fs::symlink_metadata(target)
        .map_err(|_| "bin/moai is missing in this version folder".to_string())?;
    if !meta.file_type().is_file() {
        return Err("bin/moai is a link or a folder, not a regular file".into());
    }
    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt as _;
        if meta.permissions().mode() & 0o111 == 0 {
            return Err("bin/moai is not executable".into());
        }
    }
    let mut head = Vec::new();
    fs::File::open(target)
        .and_then(|f| f.take(2).read_to_end(&mut head))
        .map_err(|_| "bin/moai cannot be read".to_string())?;
    if head == b"#!" {
        return Err("bin/moai is a script (it starts with #!), not a Rust executable".into());
    }
    let line = probe(target);
    if generation_of(line.as_deref()) == "rust" {
        Ok(())
    } else {
        Err(format!(
            "bin/moai is not the Rust build (--version: {})",
            line.as_deref().unwrap_or("no output")
        ))
    }
}

#[cfg(all(test, unix))]
mod tests {
    use super::*;
    use std::cell::Cell;
    use std::os::unix::fs::{symlink, PermissionsExt as _};
    use tempfile::TempDir;

    const RUST: &str = "moai-cli 1.2.3 (rust)";

    struct Fx {
        _tmp: TempDir,
        home: PathBuf,
        /// The state home of the empty HOME (`home_root::path`).
        state: PathBuf,
        exe: PathBuf,
        target: PathBuf,
        link_dir: PathBuf,
        link: PathBuf,
    }

    fn put(path: &Path, bytes: &[u8], mode: u32) {
        fs::create_dir_all(path.parent().unwrap()).unwrap();
        let _ = fs::remove_file(path);
        fs::write(path, bytes).unwrap();
        fs::set_permissions(path, fs::Permissions::from_mode(mode)).unwrap();
    }

    /// A temporary HOME holding `~/.moai-cli/versions/a/bin/{moai-cli,moai}` (two regular files,
    /// the second standing for the hard link) and no `path/`.
    fn fx() -> Fx {
        let tmp = TempDir::new().unwrap();
        let home = fs::canonicalize(tmp.path()).unwrap();
        let state = crate::home_root::path(&home);
        let bin = state.join("versions/a/bin");
        let exe = bin.join("moai-cli");
        put(&exe, b"\x7fELF-fake", 0o755);
        let target = bin.join("moai");
        put(&target, b"\x7fELF-fake", 0o755);
        let link_dir = state.join("path");
        let link = link_dir.join("moai");
        Fx {
            _tmp: tmp,
            home,
            state,
            exe,
            target,
            link_dir,
            link,
        }
    }

    fn rust(_: &Path) -> Option<String> {
        Some(RUST.into())
    }

    fn run(f: &Fx) -> Plan {
        plan_with(&f.home, &f.exe, false, &rust)
    }

    fn run_probe(f: &Fx, line: Option<&'static str>) -> Plan {
        plan_with(&f.home, &f.exe, false, &move |_| line.map(String::from))
    }

    fn reason(p: &Plan) -> &str {
        p.reason.as_deref().unwrap_or("")
    }

    #[test]
    fn no_link_and_a_rust_target_is_create() {
        let f = fx();
        let p = run(&f);
        assert_eq!(p.state, "create", "{p:?}");
        assert_eq!(p.link.as_deref(), Some(f.link.as_path()));
        assert_eq!(p.target.as_deref(), Some(f.target.as_path()));
        assert_eq!(p.current, None);
        assert_eq!(p.reason, None);
        assert_eq!(
            p.to_json(),
            json!({
                "state": "create",
                "reason": null,
                "link": f.link.to_string_lossy(),
                "target": f.target.to_string_lossy(),
                "current": null,
            })
        );
    }

    #[test]
    fn a_link_to_this_target_is_unchanged_absolute_or_relative() {
        let f = fx();
        fs::create_dir_all(&f.link_dir).unwrap();
        symlink(&f.target, &f.link).unwrap();
        let p = run(&f);
        assert_eq!(p.state, "unchanged", "{p:?}");
        assert_eq!(p.current.as_deref(), Some(f.target.to_str().unwrap()));
        // A relative text resolves against the link's folder, not the working folder.
        fs::remove_file(&f.link).unwrap();
        symlink("../versions/a/bin/moai", &f.link).unwrap();
        let p = run(&f);
        assert_eq!(p.state, "unchanged", "{p:?}");
        assert_eq!(p.current.as_deref(), Some("../versions/a/bin/moai"));
    }

    #[test]
    fn a_link_to_another_version_of_this_installer_is_repoint() {
        let f = fx();
        let other = f.state.join("versions/b/bin/moai");
        put(&other, b"\x7fELF-old", 0o755);
        fs::create_dir_all(&f.link_dir).unwrap();
        symlink(&other, &f.link).unwrap();
        let p = run(&f);
        assert_eq!(p.state, "repoint", "{p:?}");
        assert_eq!(p.target.as_deref(), Some(f.target.as_path()));
        // The old version folder may be gone: a dangling link into versions is still ours.
        fs::remove_file(&other).unwrap();
        let p = run(&f);
        assert_eq!(p.state, "repoint", "{p:?}");
        // So may a relative one.
        fs::remove_file(&f.link).unwrap();
        symlink("../versions/b/bin/moai", &f.link).unwrap();
        assert_eq!(run(&f).state, "repoint");
    }

    #[test]
    fn a_link_to_a_non_rust_target_is_never_unchanged() {
        let f = fx();
        fs::create_dir_all(&f.link_dir).unwrap();
        symlink(&f.target, &f.link).unwrap();
        for line in [
            Some("moai 3.2.0"),
            Some("moai-adk 3.2.0"),
            Some("other"),
            None,
        ] {
            let p = run_probe(&f, line);
            assert_eq!(p.state, "blocked", "{line:?} {p:?}");
            assert!(reason(&p).contains("not the Rust build"), "{p:?}");
            assert!(p.current.is_some(), "the link text is still reported");
        }
        // The same without a link.
        fs::remove_file(&f.link).unwrap();
        assert_eq!(run_probe(&f, Some("moai 3.2.0")).state, "blocked");
    }

    #[test]
    fn the_rust_signature_of_either_name_passes() {
        let f = fx();
        for line in ["moai-cli 0.1.0 (rust)", "moai 2026.11.0-rc.2 (rust)"] {
            assert_eq!(run_probe(&f, Some(line)).state, "create", "{line}");
        }
        // The suffix alone is not enough.
        assert_eq!(run_probe(&f, Some("moai 3.2.0 (go)")).state, "blocked");
    }

    #[test]
    fn a_script_target_is_blocked_at_exactly_the_shebang() {
        let f = fx();
        put(&f.target, b"#!/bin/sh\nexec moai-cli \"$@\"\n", 0o755);
        let p = run(&f);
        assert_eq!(p.state, "blocked", "{p:?}");
        assert!(reason(&p).contains("#!"), "{p:?}");
        // N: the two bytes `#!` alone are a script.
        put(&f.target, b"#!", 0o755);
        assert_eq!(run(&f).state, "blocked");
        // N-1 and a near miss are not: the probe decides.
        put(&f.target, b"#", 0o755);
        assert_eq!(run(&f).state, "create");
        put(&f.target, b"!#", 0o755);
        assert_eq!(run(&f).state, "create");
        put(&f.target, b" #!", 0o755);
        assert_eq!(run(&f).state, "create");
    }

    #[test]
    fn the_probe_does_not_run_for_a_script() {
        let f = fx();
        put(&f.target, b"#!/bin/sh\n", 0o755);
        let calls = Cell::new(0);
        let p = plan_with(&f.home, &f.exe, false, &|_| {
            calls.set(calls.get() + 1);
            Some(RUST.into())
        });
        assert_eq!(p.state, "blocked");
        assert_eq!(calls.get(), 0, "a script is never run to ask its version");
    }

    #[test]
    fn a_missing_unexecutable_or_linked_target_is_blocked() {
        let f = fx();
        fs::remove_file(&f.target).unwrap();
        let p = run(&f);
        assert_eq!(p.state, "blocked", "{p:?}");
        assert!(reason(&p).contains("missing"), "{p:?}");
        put(&f.target, b"\x7fELF-fake", 0o644);
        let p = run(&f);
        assert_eq!(p.state, "blocked", "{p:?}");
        assert!(reason(&p).contains("not executable"), "{p:?}");
        // One execute bit is enough (the limit is `any x`: r-----r-x passes, r--r--r-- does not).
        put(&f.target, b"\x7fELF-fake", 0o405);
        assert_eq!(run(&f).state, "create");
        put(&f.target, b"\x7fELF-fake", 0o444);
        assert_eq!(run(&f).state, "blocked");
        // A symlink in bin/ is not the regular file the installer makes.
        fs::remove_file(&f.target).unwrap();
        symlink(&f.exe, &f.target).unwrap();
        let p = run(&f);
        assert_eq!(p.state, "blocked", "{p:?}");
        assert!(reason(&p).contains("not a regular file"), "{p:?}");
    }

    #[test]
    fn a_file_or_folder_at_the_link_path_is_foreign_and_never_run() {
        let f = fx();
        let calls = Cell::new(0);
        let probe = |_: &Path| {
            calls.set(calls.get() + 1);
            Some(RUST.to_string())
        };
        put(&f.link, b"#!/bin/sh\nexec something\n", 0o755);
        let p = plan_with(&f.home, &f.exe, false, &probe);
        assert_eq!(p.state, "foreign", "{p:?}");
        assert_eq!(p.current, None);
        fs::remove_file(&f.link).unwrap();
        fs::create_dir(&f.link).unwrap();
        assert_eq!(plan_with(&f.home, &f.exe, false, &probe).state, "foreign");
        assert_eq!(calls.get(), 0, "a foreign link is reported, nothing is run");
    }

    #[test]
    fn a_link_that_leaves_the_versions_folder_is_foreign() {
        let f = fx();
        fs::create_dir_all(&f.link_dir).unwrap();
        let outside = f.home.join("go/bin/moai");
        put(&outside, b"\x7fELF-go", 0o755);
        let cases: Vec<(&str, PathBuf)> = vec![
            ("an absolute path elsewhere", outside.clone()),
            ("/usr/bin/true", PathBuf::from("/usr/bin/true")),
            (
                "a relative path out of the state home",
                "../../go/bin/moai".into(),
            ),
            (
                "`..` that climbs out of versions",
                "../versions/../../go/bin/moai".into(),
            ),
            ("a dangling path elsewhere", f.home.join("gone/moai")),
            (
                "the other name in this version",
                f.target.with_file_name("moai-cli"),
            ),
            ("the versions folder itself", f.state.join("versions")),
        ];
        for (what, text) in cases {
            let _ = fs::remove_file(&f.link);
            symlink(&text, &f.link).unwrap();
            let p = run(&f);
            assert_eq!(p.state, "foreign", "{what}: {p:?}");
            assert!(reason(&p).contains("left as it is"), "{what}: {p:?}");
            assert_eq!(p.current.as_deref(), Some(text.to_str().unwrap()), "{what}");
        }
    }

    /// The depth of the link target under `versions/` is a limit: exactly
    /// `<version>/bin/moai` is ours (N), one folder more or less is not.
    #[test]
    fn only_version_bin_moai_is_a_link_of_this_installer() {
        let f = fx();
        fs::create_dir_all(&f.link_dir).unwrap();
        let v = f.state.join("versions");
        put(&v.join("b/bin/moai"), b"x", 0o755);
        put(&v.join("b/moai"), b"x", 0o755); // N-1 components
        put(&v.join("b/bin/sub/moai"), b"x", 0o755); // N+1 components
        put(&v.join("moai"), b"x", 0o755); // N-2
        let want = [
            ("b/bin/moai", "repoint"),
            ("b/moai", "foreign"),
            ("b/bin/sub/moai", "foreign"),
            ("moai", "foreign"),
        ];
        for (rel, state) in want {
            let _ = fs::remove_file(&f.link);
            symlink(v.join(rel), &f.link).unwrap();
            assert_eq!(run(&f).state, state, "{rel}");
        }
    }

    #[test]
    fn a_version_folder_that_is_a_link_out_of_the_state_home_is_foreign() {
        let f = fx();
        let elsewhere = f.home.join("elsewhere/bin");
        put(&elsewhere.join("moai"), b"\x7fELF-x", 0o755);
        symlink(f.home.join("elsewhere"), f.state.join("versions/c")).unwrap();
        fs::create_dir_all(&f.link_dir).unwrap();
        // By its text the link is inside versions/c/bin/moai; really it is not.
        symlink(f.state.join("versions/c/bin/moai"), &f.link).unwrap();
        let p = run(&f);
        assert_eq!(p.state, "foreign", "{p:?}");
    }

    #[test]
    fn a_path_folder_that_is_a_link_is_foreign() {
        let f = fx();
        let real = f.home.join("real-path");
        fs::create_dir_all(&real).unwrap();
        symlink(&real, &f.link_dir).unwrap();
        let p = run(&f);
        assert_eq!(p.state, "foreign", "{p:?}");
        assert!(reason(&p).contains("path/ is a link"), "{p:?}");
        assert!(!real.join("moai").exists());
    }

    #[test]
    fn a_path_entry_that_is_a_file_is_foreign() {
        let f = fx();
        put(&f.link_dir, b"not a folder", 0o644);
        assert_eq!(run(&f).state, "foreign");
    }

    #[test]
    fn a_binary_outside_the_version_layout_is_outside_layout() {
        let f = fx();
        let stray = f.home.join("somewhere/moai-cli");
        put(&stray, b"\x7fELF", 0o755);
        let p = plan_with(&f.home, &stray, false, &rust);
        assert_eq!(p.state, "outside_layout", "{p:?}");
        assert!(p.link.is_none() && p.target.is_none());
        // One folder too deep is outside too (bin/ must be directly under versions/<v>/).
        let deep = f.state.join("versions/a/bin/x/moai-cli");
        put(&deep, b"\x7fELF", 0o755);
        assert_eq!(
            plan_with(&f.home, &deep, false, &rust).state,
            "outside_layout"
        );
        // A binary that does not exist cannot be placed.
        let gone = f.state.join("versions/a/bin/none");
        assert_eq!(
            plan_with(&f.home, &gone, false, &rust).state,
            "outside_layout"
        );
    }

    #[test]
    fn windows_is_reported_unsupported_before_anything_is_read() {
        let f = fx();
        let p = plan_with(&f.home, &f.exe, true, &|_| panic!("nothing is run"));
        assert_eq!(p.state, "unsupported", "{p:?}");
        assert!(reason(&p).contains("Windows"), "{p:?}");
    }

    #[test]
    fn the_state_home_follows_the_home_marker() {
        // A HOME with the Rust marker keeps its state in ~/.moai, and the plan follows it.
        let tmp = TempDir::new().unwrap();
        let home = fs::canonicalize(tmp.path()).unwrap();
        fs::create_dir_all(home.join(".moai")).unwrap();
        fs::write(
            home.join(".moai").join(crate::state_root::RUST_MARKER),
            crate::state_root::marker_bytes("test", &[]),
        )
        .unwrap();
        let exe = home.join(".moai/versions/a/bin/moai-cli");
        put(&exe, b"\x7fELF", 0o755);
        put(&exe.with_file_name("moai"), b"\x7fELF", 0o755);
        let p = plan_with(&home, &exe, false, &rust);
        assert_eq!(p.state, "create", "{p:?}");
        assert_eq!(p.link, Some(home.join(".moai/path/moai")));
    }

    #[test]
    fn planning_writes_nothing() {
        let f = fx();
        let list = || {
            let mut v = Vec::new();
            let mut stack = vec![f.home.clone()];
            while let Some(d) = stack.pop() {
                for e in fs::read_dir(&d).unwrap().flatten() {
                    let p = e.path();
                    let m = fs::symlink_metadata(&p).unwrap();
                    v.push((p.clone(), m.len(), m.modified().unwrap()));
                    if m.is_dir() {
                        stack.push(p);
                    }
                }
            }
            v.sort();
            v
        };
        let before = list();
        assert_eq!(run(&f).state, "create");
        assert_eq!(list(), before);
        assert!(!f.link_dir.exists(), "path/ is not created by a preview");
    }
}
