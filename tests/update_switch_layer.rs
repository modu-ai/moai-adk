//! SPEC-20261121-go-path-removal revision 6 (REQ-13 to REQ-18): the switch part of `moai update`
//! at library level. A known routing script (the v1 or the v2 one) at `<state folder>/path/moai`
//! is kept as `moai.before-<UTC>` and replaced by a relative link to the Rust launcher
//! `path/moai-cli`; the revert puts the kept copy back. Temporary HOME folders only; the real HOME
//! is never read or written. The two scripts are read from the preserved copies of the v2 SPEC at
//! build time: the product file that carried the v2 script left the tree.
#![cfg(unix)]

use moai_cli::update::switch::{self, V1_SHA256, V2_SHA256};
use serde_json::{json, Value};
use std::fs;
use std::os::unix::fs::{symlink, PermissionsExt};
use std::path::{Path, PathBuf};
use std::process::Command;
use std::sync::{Mutex, MutexGuard};
use tempfile::TempDir;

const V1: &[u8] = include_bytes!("../.moai/specs/SPEC-20261119-switch-script-v2/preview/moai.v1");
const V2: &[u8] = include_bytes!("../.moai/specs/SPEC-20261119-switch-script-v2/preview/moai.v2");
const NOW: &str = "2026-11-24T07:47:22Z";
const STAMP: &str = "20261124T074722Z";
const LINK: &str = "moai-cli";

/// The fault seam (`MOAI_TEST_FAULT`) is a process-wide variable: every test that applies
/// takes this gate, so a fault set by one test is never seen by another.
static GATE: Mutex<()> = Mutex::new(());

fn gate() -> MutexGuard<'static, ()> {
    GATE.lock().unwrap_or_else(|e| e.into_inner())
}

fn sha(bytes: &[u8]) -> String {
    moai_cli::digest(bytes)
        .strip_prefix("sha256:")
        .unwrap()
        .to_string()
}

/// The first line `moai --version` prints: the build's own version line, read from the constant.
fn rust_line() -> String {
    format!("moai {} (rust)", moai_cli::BUILD_VERSION)
}

fn write_executable(path: &Path, text: &str) {
    fs::write(path, text).unwrap();
    fs::set_permissions(path, fs::Permissions::from_mode(0o755)).unwrap();
}

fn mode_of(path: &Path) -> u32 {
    fs::metadata(path).unwrap().permissions().mode() & 0o777
}

struct Home {
    _t: TempDir,
    root: PathBuf,
}

impl Home {
    /// A HOME whose state folder holds `path/moai` = the v1 script at mode 755, and no launcher.
    fn new() -> Home {
        let t = TempDir::new().unwrap();
        let root = fs::canonicalize(t.path()).unwrap();
        let h = Home { _t: t, root };
        fs::create_dir_all(h.dir()).unwrap();
        h.put(V1);
        h
    }

    fn dir(&self) -> PathBuf {
        self.root.join(".moai-cli/path")
    }

    fn moai(&self) -> PathBuf {
        self.dir().join("moai")
    }

    fn launcher_file(&self) -> PathBuf {
        self.dir().join(LINK)
    }

    fn saved(&self, stamp: &str) -> PathBuf {
        self.dir().join(format!("moai.before-{stamp}"))
    }

    fn put(&self, bytes: &[u8]) {
        self.put_mode(bytes, 0o755);
    }

    fn put_mode(&self, bytes: &[u8], mode: u32) {
        let _ = fs::remove_file(self.moai());
        fs::write(self.moai(), bytes).unwrap();
        fs::set_permissions(self.moai(), fs::Permissions::from_mode(mode)).unwrap();
    }

    fn read(&self) -> Vec<u8> {
        fs::read(self.moai()).unwrap()
    }

    fn names(&self) -> Vec<String> {
        let mut v: Vec<String> = fs::read_dir(self.dir())
            .unwrap()
            .map(|e| e.unwrap().file_name().to_string_lossy().into_owned())
            .collect();
        v.sort();
        v
    }

    /// A launcher script that records every call under the HOME and prints the build's Rust
    /// version line first, the way the Rust build answers `--version`.
    fn rust_launcher(&self) -> PathBuf {
        let log = self.root.join("launcher-calls");
        self.launcher(&format!(
            "#!/bin/sh\necho \"$*\" >> '{}'\necho '{}'\n",
            log.display(),
            rust_line()
        ))
    }

    fn launcher(&self, text: &str) -> PathBuf {
        let p = self.launcher_file();
        let _ = fs::remove_file(&p);
        write_executable(&p, text);
        p
    }

    /// A `go/bin/moai.old` that records a start and fails.
    fn moai_old(&self) -> PathBuf {
        let log = self.root.join("moai-old-calls");
        let p = self.root.join("go/bin/moai.old");
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        write_executable(
            &p,
            &format!("#!/bin/sh\necho \"$*\" >> '{}'\nexit 1\n", log.display()),
        );
        p
    }

    fn calls(&self, name: &str) -> Vec<String> {
        fs::read_to_string(self.root.join(name))
            .map(|t| t.lines().map(str::to_string).collect())
            .unwrap_or_default()
    }

    fn token(&self) -> String {
        switch::plan(&self.root)
            .unwrap()
            .token()
            .expect("a token")
            .to_string()
    }

    fn part(&self) -> Option<Value> {
        switch::plan(&self.root).unwrap().to_json().cloned()
    }

    fn go_bin(&self, bytes: &[u8]) -> PathBuf {
        let dir = self.root.join("go/bin");
        fs::create_dir_all(&dir).unwrap();
        let p = dir.join("moai");
        fs::write(&p, bytes).unwrap();
        fs::set_permissions(&p, fs::Permissions::from_mode(0o755)).unwrap();
        p
    }

    /// `path/moai` started by its path with these words: the exit code and stdout.
    fn through_the_link(&self, args: &[&str]) -> (Option<i32>, String) {
        let out = Command::new(self.moai())
            .args(args)
            .env("HOME", &self.root)
            .env("PATH", "/usr/bin:/bin")
            .output()
            .unwrap();
        (
            out.status.code(),
            String::from_utf8_lossy(&out.stdout).into_owned(),
        )
    }

    /// Whether `path/moai` is the relative link this step makes.
    fn is_the_link(&self) -> bool {
        fs::read_link(self.moai()).is_ok_and(|t| t == Path::new(LINK))
    }
}

// ---- the scripts the step knows ---------------------------------------------

#[test]
fn the_pinned_digests_are_the_preserved_scripts() {
    assert_eq!(sha(V1), V1_SHA256);
    assert_eq!(sha(V2), V2_SHA256);
    assert_eq!(
        V1_SHA256,
        "3c3446a3dbd7d918a7d6465c0765ae66b2dafe3231b9be11cd3181737b6658ac"
    );
    assert_eq!(
        V2_SHA256,
        "36dd242bf1337efa9b4759ba7166fb6adea04427800334a374027cc246eb26e4"
    );
    assert_ne!(V1, V2, "the two known scripts differ");
}

// ---- the preview (REQ-13, REQ-15) -------------------------------------------

#[test]
fn plan_offers_link_for_the_known_v1_script() {
    let _g = gate();
    let h = Home::new();
    let launcher = h.rust_launcher();
    let before = h.names();
    let layer = switch::plan(&h.root).unwrap();
    let p = layer
        .to_json()
        .expect("a known script beside a Rust launcher has a switch part");
    assert_eq!(p["action"], json!("link"), "{p}");
    assert_eq!(p["file"], json!(h.moai().display().to_string()));
    assert_eq!(p["script"], json!("v1"));
    assert_eq!(p["script_sha256"], json!(V1_SHA256));
    assert_eq!(p["launcher"], json!(launcher.display().to_string()), "{p}");
    assert_eq!(p["launcher_version_line"], json!(rust_line()), "{p}");
    assert_eq!(p["link_target"], json!(LINK));
    assert_eq!(p["saved_as"], json!("moai.before-<UTC>"));
    assert_eq!(p["writes"], json!(false));
    assert_eq!(p["original_kept"], json!(true));
    let t = layer.token().expect("a token");
    assert!(t.starts_with("switch:") && t.len() == 7 + 16, "{t}");
    assert_eq!(p["confirm_token"], json!(t));
    assert_eq!(p["apply_flag"], json!(format!("--confirm-switch {t}")));
    assert_eq!(
        p["revert_with"],
        json!("moai update --revert-switch <stamp>")
    );
    let line = layer.summary();
    assert!(
        line.contains("moai.before-<UTC>")
            && line.contains(&format!("--confirm-switch {t}"))
            && line.contains("link"),
        "{line}"
    );
    // The token is the content's digest, not a clock: a second preview gives the same one.
    assert_eq!(h.token(), t);
    // A preview writes nothing.
    assert_eq!(h.names(), before);
    assert_eq!(h.read(), V1);
}

#[test]
fn plan_offers_link_for_v2_too() {
    let _g = gate();
    let h = Home::new();
    h.put(V2);
    h.rust_launcher();
    let layer = switch::plan(&h.root).unwrap();
    let p = layer.to_json().expect("v2 is a known script too");
    assert_eq!(p["action"], json!("link"), "{p}");
    assert_eq!(p["script"], json!("v2"));
    assert_eq!(p["script_sha256"], json!(V2_SHA256));
    assert!(layer.token().is_some(), "{p}");
    assert_eq!(h.read(), V2, "a preview writes nothing");
}

#[test]
fn plan_has_no_part_for_a_link_an_absent_file_or_no_state_folder() {
    let _g = gate();
    // absent, beside a Rust launcher
    let h = Home::new();
    h.rust_launcher();
    fs::remove_file(h.moai()).unwrap();
    assert!(h.part().is_none());
    // a link (the installer's own kind) to a real file, and a dangling one
    let real = h.root.join("real-moai-cli");
    fs::write(&real, V1).unwrap();
    symlink(&real, h.moai()).unwrap();
    assert!(
        h.part().is_none(),
        "a link is the installer's, never offered"
    );
    fs::remove_file(h.moai()).unwrap();
    symlink(h.root.join("nowhere"), h.moai()).unwrap();
    assert!(h.part().is_none());
    // no state folder at all
    let t = TempDir::new().unwrap();
    let empty = fs::canonicalize(t.path()).unwrap();
    assert!(switch::plan(&empty).unwrap().to_json().is_none());
}

#[test]
fn plan_leaves_an_unknown_file_alone() {
    let _g = gate();
    let mut edited = V1.to_vec();
    edited[40] ^= 1;
    let other: [Vec<u8>; 4] = [
        b"#!/bin/sh\necho mine\n".to_vec(),
        edited,
        b"#!/bin/sh\necho '# MoAI name-switch shim'\nexit 0\n".to_vec(),
        vec![b'x'; 70_000],
    ];
    for bytes in other {
        let h = Home::new();
        // A launcher beside it does not make an unknown file the step's.
        h.rust_launcher();
        h.put(&bytes);
        let layer = switch::plan(&h.root).unwrap();
        let p = layer.to_json().expect("an unknown file is shown");
        assert_eq!(p["action"], json!("left"), "{p}");
        assert!(p["reason"].as_str().unwrap().contains("not"), "{p}");
        assert!(
            layer.token().is_none() && p["confirm_token"].is_null(),
            "{p}"
        );
        assert!(p["apply_flag"].is_null(), "{p}");
        assert!(switch::apply(&h.root, "switch:0000000000000000", NOW).is_err());
        assert_eq!(h.read(), bytes, "the file stays as it is");
        assert!(!h.saved(STAMP).exists());
    }
}

/// A launcher that is not shown to be the Rust build holds the step: the script stays, no token
/// is given, and a token that is given writes nothing (REQ-15).
#[test]
fn plan_is_held_while_the_launcher_is_not_shown_to_be_rust() {
    let _g = gate();
    let mut held = 0;
    let mut check = |h: &Home, what: &str| {
        let layer = switch::plan(&h.root).unwrap();
        let p = layer.to_json().expect("the known script is still shown");
        assert_eq!(p["action"], json!("held"), "{what}: {p}");
        assert!(layer.token().is_none(), "{what}: {p}");
        assert!(
            p["confirm_token"].is_null() && p["apply_flag"].is_null(),
            "{what}: {p}"
        );
        let reason = p["reason"].as_str().unwrap();
        assert!(reason.contains(LINK), "{what}: {reason}");
        assert!(!layer.summary().is_empty(), "{what}");
        let e = switch::apply(&h.root, "switch:0000000000000000", NOW).unwrap_err();
        assert!(e.contains("preview again"), "{what}: {e}");
        assert_eq!(h.read(), V1, "{what}: the script stays");
        assert!(!h.saved(STAMP).exists(), "{what}: nothing is written");
        held += 1;
    };
    // no launcher at all
    let h = Home::new();
    check(&h, "absent");
    // a launcher without an execute bit
    let h = Home::new();
    fs::write(h.launcher_file(), "#!/bin/sh\necho 'moai 1.0.0 (rust)'\n").unwrap();
    check(&h, "not executable");
    // one that answers with a Go line
    let h = Home::new();
    h.launcher("#!/bin/sh\necho 'moai-adk v3.2.0-rc.23'\n");
    check(&h, "a Go line");
    // one that prints nothing
    let h = Home::new();
    h.launcher("#!/bin/sh\nexit 1\n");
    check(&h, "no line");
    // one that prints the Rust words in another shape
    let h = Home::new();
    h.launcher("#!/bin/sh\necho 'moai 1.0.0 rust'\n");
    check(&h, "not the signature");
    // a folder of that name
    let h = Home::new();
    fs::create_dir(h.launcher_file()).unwrap();
    check(&h, "a folder");
    // a dangling link of that name, and one to a folder
    let h = Home::new();
    symlink(h.root.join("nowhere"), h.launcher_file()).unwrap();
    check(&h, "a dangling link");
    let h = Home::new();
    symlink(&h.root, h.launcher_file()).unwrap();
    check(&h, "a link to a folder");
    // A launcher the preview did not see makes an older token stale.
    let h = Home::new();
    let first = h.token_with(|| {
        h.launcher(&format!("#!/bin/sh\necho '{}'\n", rust_line()));
    });
    h.launcher("#!/bin/sh\necho 'moai 9.9.9 (rust)'\n");
    let e = switch::apply(&h.root, &first, NOW).unwrap_err();
    assert!(e.contains("preview again"), "{e}");
    assert_eq!(h.read(), V1);
    assert_eq!(held, 8);
}

impl Home {
    /// The preview's token after `make` laid the launcher.
    fn token_with(&self, make: impl FnOnce()) -> String {
        make();
        self.token()
    }
}

#[test]
fn go_bin_is_reported_and_never_written() {
    let _g = gate();
    let h = Home::new();
    h.rust_launcher();
    // absent
    let p = h.part().unwrap();
    assert_eq!(p["go_bin"]["exists"], json!(false), "{p}");
    assert_eq!(p["go_bin"]["same_as_v1"], json!(false));
    // the same old script
    let g = h.go_bin(V1);
    let before = fs::metadata(&g).unwrap();
    let p = h.part().unwrap();
    assert_eq!(p["go_bin"]["path"], json!(g.display().to_string()), "{p}");
    assert_eq!(p["go_bin"]["exists"], json!(true));
    assert_eq!(p["go_bin"]["same_as_v1"], json!(true));
    assert!(!p["go_bin"]["note"].as_str().unwrap().is_empty());
    // notes for the Go-generation projects
    let notes = p["notes"].as_array().unwrap();
    let all = notes
        .iter()
        .map(|n| n.as_str().unwrap())
        .collect::<Vec<_>>()
        .join("\n");
    assert!(
        all.contains("moai hook") && all.contains("moai statusline"),
        "{all}"
    );
    // apply leaves it untouched: bytes, mode and modification time
    let t = h.token();
    switch::apply(&h.root, &t, NOW).unwrap();
    let after = fs::metadata(&g).unwrap();
    assert_eq!(fs::read(&g).unwrap(), V1);
    assert_eq!(after.permissions().mode(), before.permissions().mode());
    assert_eq!(after.modified().unwrap(), before.modified().unwrap());
    // other bytes
    let h2 = Home::new();
    h2.go_bin(b"#!/bin/sh\necho other\n");
    let p = h2.part().unwrap();
    assert_eq!(p["go_bin"]["exists"], json!(true));
    assert_eq!(p["go_bin"]["same_as_v1"], json!(false));
}

// ---- the apply (REQ-14, REQ-15, REQ-18) --------------------------------------

#[test]
fn apply_keeps_v1_and_makes_the_link() {
    let _g = gate();
    let h = Home::new();
    h.rust_launcher();
    h.moai_old();
    let token = h.token();
    let r = switch::apply(&h.root, &token, NOW).unwrap();
    assert_eq!(r["state"], json!("applied"), "{r}");
    assert_eq!(r["stamp"], json!(STAMP));
    assert_eq!(r["script"], json!("v1"), "{r}");
    let saved = h.saved(STAMP);
    assert_eq!(r["saved"], json!(saved.display().to_string()));
    assert_eq!(
        fs::read(&saved).unwrap(),
        V1,
        "the old script is kept byte for byte"
    );
    assert_eq!(mode_of(&saved), 0o755);
    // `moai` is the relative link to the launcher beside it.
    assert!(h.is_the_link(), "{:?}", fs::symlink_metadata(h.moai()));
    assert_eq!(r["link"], json!(LINK), "{r}");
    assert_eq!(
        r["revert"],
        json!(format!("moai update --revert-switch {STAMP}"))
    );
    assert_eq!(
        r["revert_plain"],
        json!(format!(
            "mv \"{}\" \"{}\"",
            saved.display(),
            h.moai().display()
        ))
    );
    assert_eq!(
        h.names(),
        vec!["moai", LINK, &format!("moai.before-{STAMP}")[..]]
    );
    // Through the link a word only the Go build had reaches the launcher, and
    // `moai.old` is not started. Before that the product itself started the launcher only to
    // ask its version: once for the preview that gave the token and once inside the apply.
    let (code, out) = h.through_the_link(&["hook", "pre-tool"]);
    assert_eq!(code, Some(0));
    assert_eq!(out.lines().next(), Some(rust_line().as_str()));
    let calls = h.calls("launcher-calls");
    assert_eq!(
        calls.last().map(String::as_str),
        Some("hook pre-tool"),
        "{calls:?}"
    );
    assert!(
        calls[..calls.len() - 1]
            .iter()
            .all(|call| call == "--version"),
        "the product asks the launcher only for its version: {calls:?}"
    );
    assert!(
        calls.len() >= 2,
        "the preview and the apply each asked once: {calls:?}"
    );
    assert!(h.calls("moai-old-calls").is_empty());
    // Nothing is left to offer afterwards.
    assert!(h.part().is_none());
}

#[test]
fn apply_keeps_v2_and_makes_the_link() {
    let _g = gate();
    let h = Home::new();
    h.put(V2);
    h.rust_launcher();
    let r = switch::apply(&h.root, &h.token(), NOW).unwrap();
    assert_eq!(r["script"], json!("v2"), "{r}");
    assert_eq!(fs::read(h.saved(STAMP)).unwrap(), V2);
    assert!(h.is_the_link());
    assert!(h.part().is_none());
}

#[test]
fn apply_keeps_the_mode_of_the_old_script() {
    let _g = gate();
    let h = Home::new();
    h.put_mode(V1, 0o700);
    h.rust_launcher();
    switch::apply(&h.root, &h.token(), NOW).unwrap();
    assert_eq!(mode_of(&h.saved(STAMP)), 0o700);
}

#[test]
fn apply_refuses_a_wrong_or_stale_token() {
    let _g = gate();
    let h = Home::new();
    h.rust_launcher();
    let layer = switch::plan(&h.root).unwrap();
    let token = layer.token().unwrap().to_string();
    for bad in ["switch:0000000000000000", "home:aaaaaaaaaaaaaaaa", ""] {
        let e = switch::apply(&h.root, bad, NOW).unwrap_err();
        assert!(e.contains("token") && e.contains("preview again"), "{e}");
        assert_eq!(h.read(), V1);
        assert!(!h.saved(STAMP).exists());
    }
    // check_token: its own token and nothing else.
    assert_eq!(switch::check_token(&layer, None), Ok(None));
    assert_eq!(
        switch::check_token(&layer, Some(token.as_str())),
        Ok(Some(token.as_str()))
    );
    for bad in ["switch:bbbbbbbbbbbbbbbb", "home:aaaaaaaaaaaaaaaa", ""] {
        assert!(switch::check_token(&layer, Some(bad)).is_err(), "{bad}");
    }
    // No token where the step is held or has nothing to do.
    let held = Home::new();
    assert!(switch::check_token(&switch::plan(&held.root).unwrap(), Some(token.as_str())).is_err());
    // Stale: the script changed after the preview.
    let h = Home::new();
    h.rust_launcher();
    let token = h.token();
    h.put(b"#!/bin/sh\necho edited after the preview\n");
    let e = switch::apply(&h.root, &token, NOW).unwrap_err();
    assert!(e.contains("preview again"), "{e}");
    assert_eq!(h.read(), b"#!/bin/sh\necho edited after the preview\n");
    assert!(!h.saved(STAMP).exists());
    // Stale: the other known script stands there now.
    let h = Home::new();
    h.rust_launcher();
    let token = h.token();
    h.put(V2);
    assert!(switch::apply(&h.root, &token, NOW).is_err());
    assert_eq!(h.read(), V2);
    assert!(!h.saved(STAMP).exists());
}

#[test]
fn apply_checks_everything_before_the_first_write() {
    let _g = gate();
    // The kept-copy name is taken: nothing is written, the taken name is not touched.
    let h = Home::new();
    h.rust_launcher();
    let token = h.token();
    fs::write(h.saved(STAMP), b"keep me").unwrap();
    let e = switch::apply(&h.root, &token, NOW).unwrap_err();
    assert!(e.contains("already exists"), "{e}");
    assert_eq!(fs::read(h.saved(STAMP)).unwrap(), b"keep me");
    assert_eq!(h.read(), V1);
    assert!(!h.is_the_link());

    // The temporary name is taken: nothing is written, no kept copy appears.
    let h = Home::new();
    h.rust_launcher();
    let token = h.token();
    let temp = h.dir().join(format!(".moai.switch-{STAMP}"));
    fs::write(&temp, b"taken").unwrap();
    let e = switch::apply(&h.root, &token, NOW).unwrap_err();
    assert!(e.contains("already exists"), "{e}");
    assert_eq!(fs::read(&temp).unwrap(), b"taken");
    assert!(!h.saved(STAMP).exists());
    assert_eq!(h.read(), V1);

    // The folder became a link after the preview.
    let h = Home::new();
    h.rust_launcher();
    let token = h.token();
    let real = h.root.join(".moai-cli/path-real");
    fs::rename(h.dir(), &real).unwrap();
    symlink(&real, h.dir()).unwrap();
    let e = switch::apply(&h.root, &token, NOW).unwrap_err();
    assert!(e.contains("preview again") || e.contains("link"), "{e}");
    assert_eq!(fs::read(real.join("moai")).unwrap(), V1);
    assert_eq!(
        fs::read_dir(&real).unwrap().count(),
        2,
        "the script and the launcher"
    );

    // The file became a link after the preview.
    let h = Home::new();
    h.rust_launcher();
    let token = h.token();
    let target = h.root.join("elsewhere");
    fs::write(&target, V1).unwrap();
    fs::remove_file(h.moai()).unwrap();
    symlink(&target, h.moai()).unwrap();
    assert!(switch::apply(&h.root, &token, NOW).is_err());
    assert_eq!(fs::read(&target).unwrap(), V1);
    assert_eq!(h.names(), vec!["moai", LINK]);
    assert!(!h.saved(STAMP).exists());

    // The launcher went away after the preview.
    let h = Home::new();
    h.rust_launcher();
    let token = h.token();
    fs::remove_file(h.launcher_file()).unwrap();
    assert!(switch::apply(&h.root, &token, NOW).is_err());
    assert_eq!(h.read(), V1);
    assert!(!h.saved(STAMP).exists());

    // A clock that is not a UTC time names no file.
    let h = Home::new();
    h.rust_launcher();
    let token = h.token();
    for bad in ["not a time", "../../x", "", "2026-11-24 07:47:22"] {
        let e = switch::apply(&h.root, &token, bad).unwrap_err();
        assert!(e.contains("UTC"), "{bad}: {e}");
        assert!(!h.saved(STAMP).exists(), "{bad}");
        assert_eq!(h.read(), V1, "{bad}");
    }
}

#[test]
fn a_failed_step_leaves_a_whole_script() {
    let _g = gate();
    // A folder that cannot be written (skipped when the process may write anyway, e.g. root).
    let h = Home::new();
    h.rust_launcher();
    let token = h.token();
    fs::set_permissions(h.dir(), fs::Permissions::from_mode(0o555)).unwrap();
    let probe = fs::write(h.dir().join("probe"), b"x");
    if probe.is_err() {
        let e = switch::apply(&h.root, &token, NOW).unwrap_err();
        assert!(!e.is_empty());
        fs::set_permissions(h.dir(), fs::Permissions::from_mode(0o755)).unwrap();
        assert_eq!(h.read(), V1, "still the whole v1");
        assert!(!h.saved(STAMP).exists(), "no kept copy, no temporary file");
    } else {
        let _ = fs::remove_file(h.dir().join("probe"));
        fs::set_permissions(h.dir(), fs::Permissions::from_mode(0o755)).unwrap();
    }
    // Faults after the first write (debug builds read the seam): the script stays whole, the
    // temporary name is gone, the kept copy is kept and named in the message.
    for point in ["switch-after-save", "switch-rename"] {
        for (bytes, name) in [(V1, "v1"), (V2, "v2")] {
            let h = Home::new();
            h.put(bytes);
            h.rust_launcher();
            let token = h.token();
            std::env::set_var("MOAI_TEST_FAULT", point);
            let r = switch::apply(&h.root, &token, NOW);
            std::env::remove_var("MOAI_TEST_FAULT");
            let e = r.unwrap_err();
            assert!(e.contains(point), "{name} {point}: {e}");
            assert_eq!(h.read(), bytes, "{name} {point}: still the whole script");
            assert!(!h.is_the_link(), "{name} {point}");
            assert_eq!(
                h.names(),
                vec!["moai", LINK, &format!("moai.before-{STAMP}")[..]],
                "{name} {point}: no temporary file left"
            );
            assert!(
                e.contains(&h.saved(STAMP).display().to_string()),
                "{name} {point}: names the kept copy: {e}"
            );
            assert_eq!(fs::read(h.saved(STAMP)).unwrap(), bytes);
            // The next apply, with a new stamp, still works and keeps the first copy.
            let r = switch::apply(&h.root, &h.token(), "2026-11-24T07:48:00Z").unwrap();
            assert_eq!(r["state"], json!("applied"), "{name} {point}: {r}");
            assert!(h.is_the_link());
            assert_eq!(fs::read(h.saved(STAMP)).unwrap(), bytes);
        }
    }
}

// ---- the revert (REQ-16, REQ-17) ---------------------------------------------

#[test]
fn revert_puts_the_kept_script_back_and_refuses_anything_else() {
    let _g = gate();
    let h = Home::new();
    h.rust_launcher();
    switch::apply(&h.root, &h.token(), NOW).unwrap();
    // Every refusal changes nothing.
    let state = |h: &Home| (h.is_the_link(), fs::read(h.saved(STAMP)).unwrap());
    let before = state(&h);
    for bad in [
        "",
        "2026",
        "../x",
        "20261124T074722",
        "20261124T074722Zx",
        "20261124T0747227Z",
    ] {
        let e = switch::revert(&h.root, bad).unwrap_err();
        assert!(e.contains("stamp"), "{bad}: {e}");
        assert_eq!(state(&h), before, "{bad}");
    }
    let e = switch::revert(&h.root, "20250101T000000Z").unwrap_err();
    assert!(e.contains("no saved copy"), "{e}");
    // The saved copy is not a known script (changed by hand): refused, left as it is.
    fs::write(h.saved(STAMP), b"#!/bin/sh\nnot a known script\n").unwrap();
    let e = switch::revert(&h.root, STAMP).unwrap_err();
    assert!(e.contains("known script"), "{e}");
    assert_eq!(
        fs::read(h.saved(STAMP)).unwrap(),
        b"#!/bin/sh\nnot a known script\n"
    );
    assert!(h.is_the_link());
    // The saved copy is a link: refused.
    fs::remove_file(h.saved(STAMP)).unwrap();
    let real = h.root.join("v1-elsewhere");
    fs::write(&real, V1).unwrap();
    symlink(&real, h.saved(STAMP)).unwrap();
    assert!(switch::revert(&h.root, STAMP).is_err());
    assert!(h.is_the_link());
    fs::remove_file(h.saved(STAMP)).unwrap();
    fs::write(h.saved(STAMP), V1).unwrap();
    fs::set_permissions(h.saved(STAMP), fs::Permissions::from_mode(0o755)).unwrap();
    // `moai` is not the link this step made: a file edited by hand, a link elsewhere, absent.
    h.put(b"#!/bin/sh\nmy own edit\n");
    let e = switch::revert(&h.root, STAMP).unwrap_err();
    assert!(e.contains("link"), "{e}");
    assert_eq!(h.read(), b"#!/bin/sh\nmy own edit\n");
    assert_eq!(fs::read(h.saved(STAMP)).unwrap(), V1);
    fs::remove_file(h.moai()).unwrap();
    symlink(&real, h.moai()).unwrap();
    let e = switch::revert(&h.root, STAMP).unwrap_err();
    assert!(e.contains("link"), "{e}");
    assert_eq!(fs::read(&real).unwrap(), V1);
    fs::remove_file(h.moai()).unwrap();
    assert!(switch::revert(&h.root, STAMP).is_err());
    assert_eq!(fs::read(h.saved(STAMP)).unwrap(), V1, "the kept copy stays");
    // The good case: back to v1 by one rename; the kept name is gone.
    symlink(Path::new(LINK), h.moai()).unwrap();
    let r = switch::revert(&h.root, STAMP).unwrap();
    assert_eq!(r["state"], json!("reverted"), "{r}");
    assert_eq!(r["script"], json!("v1"));
    assert_eq!(h.read(), V1);
    assert_eq!(mode_of(&h.moai()), 0o755);
    assert!(!fs::symlink_metadata(h.moai())
        .unwrap()
        .file_type()
        .is_symlink());
    assert_eq!(h.names(), vec!["moai", LINK]);
    // A second revert finds nothing.
    let e = switch::revert(&h.root, STAMP).unwrap_err();
    assert!(e.contains("no saved copy"), "{e}");
    assert_eq!(h.read(), V1);
    // And the preview offers the link again.
    assert_eq!(h.part().unwrap()["action"], json!("link"));
}

#[test]
fn revert_puts_v2_back_too_and_refuses_a_copy_of_the_old_step() {
    let _g = gate();
    // v2 was kept: it comes back.
    let h = Home::new();
    h.put(V2);
    h.rust_launcher();
    switch::apply(&h.root, &h.token(), NOW).unwrap();
    let r = switch::revert(&h.root, STAMP).unwrap();
    assert_eq!(r["script"], json!("v2"), "{r}");
    assert_eq!(h.read(), V2);
    assert!(!h.is_the_link());
    // A copy the earlier v1-to-v2 step kept, beside a plain v2 file it wrote: not this step's
    // link, so nothing moves and the copy stays.
    let h = Home::new();
    h.put(V2);
    h.rust_launcher();
    fs::write(h.saved(STAMP), V1).unwrap();
    fs::set_permissions(h.saved(STAMP), fs::Permissions::from_mode(0o755)).unwrap();
    let e = switch::revert(&h.root, STAMP).unwrap_err();
    assert!(e.contains("link"), "{e}");
    assert_eq!(h.read(), V2);
    assert_eq!(fs::read(h.saved(STAMP)).unwrap(), V1);
}

// ---- the update command, library part (REQ-16) ------------------------------

#[test]
fn revert_switch_runs_alone_and_is_refused_beside_a_confirming_flag() {
    use moai_cli::update::update_command_switch;
    let _g = gate();
    let h = Home::new();
    h.rust_launcher();
    switch::apply(&h.root, &h.token(), NOW).unwrap();
    let ask = |_: &str, _: &Value| Ok(false);
    let flags: [[Option<&str>; 4]; 4] = [
        [Some("x"), None, None, None],
        [None, Some("x"), None, None],
        [None, None, Some("x"), None],
        [None, None, None, Some("x")],
    ];
    for [c, cc, ch, cs] in flags {
        let e = update_command_switch(&h.root, c, cc, ch, cs, None, Some(STAMP), false, ask)
            .unwrap_err();
        assert!(e.contains("cannot be combined"), "{e}");
        assert!(h.is_the_link(), "nothing changed");
        assert!(h.saved(STAMP).exists());
    }
    // Alone it runs without any installed package: only the kept copy and `moai` are read.
    let r = update_command_switch(
        &h.root,
        None,
        None,
        None,
        None,
        None,
        Some(STAMP),
        false,
        ask,
    )
    .unwrap();
    assert_eq!(r["mode"], json!("revert_switch"), "{r}");
    assert_eq!(r["switch_result"]["state"], json!("reverted"), "{r}");
    assert_eq!(h.read(), V1);
    // An apply result's `revert` line comes first in the revert order.
    let order = moai_cli::update::revert_order(&json!({
        "switch": "moai update --revert-switch X", "home": "h", "agents": "a"
    }));
    assert_eq!(order, vec!["switch", "home", "agents"]);
}
