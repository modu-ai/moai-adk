//! C-33 S3 (SPEC-20261119-update-switch-layer): the switch part of `moai update` at library level.
//! Temporary HOME folders only; the real HOME is never read or written. The v1 script is read
//! from the S2a preview at build time; the v2 script is the one built into the program.
#![cfg(unix)]

use moai_cli::update::switch::{self, V1_SHA256, V2, V2_SHA256};
use serde_json::{json, Value};
use std::fs;
use std::os::unix::fs::{symlink, PermissionsExt};
use std::path::PathBuf;
use std::sync::{Mutex, MutexGuard};
use tempfile::TempDir;

const V1: &[u8] = include_bytes!("../.moai/specs/SPEC-20261119-switch-script-v2/preview/moai.v1");
const NOW: &str = "2026-11-19T07:47:22Z";
const STAMP: &str = "20261119T074722Z";

/// The fault seam (`MOAI_CLI_TEST_FAULT`) is a process-wide variable: every test that applies
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

struct Home {
    _t: TempDir,
    root: PathBuf,
}

impl Home {
    /// A HOME whose state folder holds `path/moai` = the v1 script at mode 755.
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

    fn saved(&self, stamp: &str) -> PathBuf {
        self.dir().join(format!("moai.before-{stamp}"))
    }

    fn put(&self, bytes: &[u8]) {
        let _ = fs::remove_file(self.moai());
        fs::write(self.moai(), bytes).unwrap();
        fs::set_permissions(self.moai(), fs::Permissions::from_mode(0o755)).unwrap();
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
}

// ---- AC-5 -----------------------------------------------------------------

#[test]
fn the_pinned_digests_are_the_s2a_scripts() {
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
    assert_eq!(V1.len(), 1747, "the installed v1 script is 1747 B");
}

// ---- AC-1 -----------------------------------------------------------------

#[test]
fn plan_offers_v2_for_the_known_v1_script() {
    let _g = gate();
    let h = Home::new();
    let before = h.names();
    let layer = switch::plan(&h.root).unwrap();
    let p = layer.to_json().expect("a v1 script has a switch part");
    assert_eq!(p["action"], json!("replace"), "{p}");
    assert_eq!(p["file"], json!(h.moai().display().to_string()));
    assert_eq!(p["saved_as"], json!("moai.before-<UTC>"));
    assert_eq!(p["v1_sha256"], json!(V1_SHA256));
    assert_eq!(p["v2_sha256"], json!(V2_SHA256));
    assert_eq!(p["writes"], json!(false));
    let t = layer.token().expect("a token");
    assert!(t.starts_with("switch:") && t.len() == 7 + 16, "{t}");
    assert_eq!(p["apply_flag"], json!(format!("--confirm-switch {t}")));
    let line = layer.summary();
    assert!(line.contains("moai.before-<UTC>"), "{line}");
    assert!(line.contains(&format!("--confirm-switch {t}")), "{line}");
    // The token is the content's digest, not a clock: a second preview gives the same one.
    assert_eq!(h.token(), t);
    // A preview writes nothing.
    assert_eq!(h.names(), before);
    assert_eq!(h.read(), V1);
}

#[test]
fn plan_has_no_part_for_a_link_an_absent_file_or_v2() {
    let _g = gate();
    // v2 already there.
    let h = Home::new();
    h.put(V2);
    assert!(h.part().is_none() && switch::plan(&h.root).unwrap().token().is_none());
    assert_eq!(switch::plan(&h.root).unwrap().summary(), "");
    // absent
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
    let h = Home::new();
    let mut edited = V1.to_vec();
    edited[40] ^= 1;
    let other: [Vec<u8>; 3] = [
        b"#!/bin/sh\necho mine\n".to_vec(),
        edited,
        vec![b'x'; 70_000],
    ];
    for bytes in other {
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
        assert_eq!(h.names(), vec!["moai"]);
    }
}

#[test]
fn go_bin_is_reported_and_never_written() {
    let _g = gate();
    let h = Home::new();
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

// ---- AC-2 -----------------------------------------------------------------

#[test]
fn apply_saves_v1_and_installs_v2() {
    let _g = gate();
    let h = Home::new();
    let token = h.token();
    let r = switch::apply(&h.root, &token, NOW).unwrap();
    assert_eq!(r["state"], json!("applied"), "{r}");
    assert_eq!(r["stamp"], json!(STAMP));
    let saved = h.saved(STAMP);
    assert_eq!(r["saved"], json!(saved.display().to_string()));
    assert_eq!(
        fs::read(&saved).unwrap(),
        V1,
        "the old script is kept byte for byte"
    );
    assert_eq!(
        fs::metadata(&saved).unwrap().permissions().mode() & 0o777,
        0o755
    );
    assert_eq!(h.read(), V2);
    assert_eq!(
        fs::metadata(h.moai()).unwrap().permissions().mode() & 0o777,
        0o755
    );
    assert_eq!(
        r["revert"],
        json!(format!("moai-cli update --revert-switch {STAMP}"))
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
        vec!["moai", format!("moai.before-{STAMP}").as_str()]
    );
    // Nothing is left to offer afterwards.
    assert!(h.part().is_none());
}

// ---- AC-3 -----------------------------------------------------------------

#[test]
fn apply_refuses_a_wrong_or_stale_token() {
    let _g = gate();
    let h = Home::new();
    let layer = switch::plan(&h.root).unwrap();
    let token = layer.token().unwrap().to_string();
    for bad in ["switch:0000000000000000", "home:aaaaaaaaaaaaaaaa", ""] {
        let e = switch::apply(&h.root, bad, NOW).unwrap_err();
        assert!(e.contains("token") && e.contains("preview again"), "{e}");
        assert_eq!(h.read(), V1);
        assert_eq!(h.names(), vec!["moai"]);
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
    h.put(V2);
    let none = switch::plan(&h.root).unwrap();
    assert!(switch::check_token(&none, Some(token.as_str())).is_err());
    // Stale: the file changed after the preview.
    h.put(b"#!/bin/sh\necho edited after the preview\n");
    let e = switch::apply(&h.root, &token, NOW).unwrap_err();
    assert!(e.contains("preview again"), "{e}");
    assert_eq!(h.read(), b"#!/bin/sh\necho edited after the preview\n");
    assert_eq!(h.names(), vec!["moai"]);
}

#[test]
fn apply_checks_everything_before_the_first_write() {
    let _g = gate();
    // The saved-copy name is taken: nothing is written, the taken name is not touched.
    let h = Home::new();
    let token = h.token();
    fs::write(h.saved(STAMP), b"keep me").unwrap();
    let e = switch::apply(&h.root, &token, NOW).unwrap_err();
    assert!(e.contains("already exists"), "{e}");
    assert_eq!(fs::read(h.saved(STAMP)).unwrap(), b"keep me");
    assert_eq!(h.read(), V1);
    assert_eq!(h.names(), vec!["moai", &format!("moai.before-{STAMP}")[..]]);

    // The temporary name is taken: nothing is written, no saved copy appears.
    let h = Home::new();
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
    let token = h.token();
    let real = h.root.join(".moai-cli/path-real");
    fs::rename(h.dir(), &real).unwrap();
    symlink(&real, h.dir()).unwrap();
    let e = switch::apply(&h.root, &token, NOW).unwrap_err();
    assert!(e.contains("preview again") || e.contains("link"), "{e}");
    assert_eq!(fs::read(real.join("moai")).unwrap(), V1);
    assert_eq!(fs::read_dir(&real).unwrap().count(), 1);

    // The file became a link after the preview.
    let h = Home::new();
    let token = h.token();
    let target = h.root.join("elsewhere");
    fs::write(&target, V1).unwrap();
    fs::remove_file(h.moai()).unwrap();
    symlink(&target, h.moai()).unwrap();
    assert!(switch::apply(&h.root, &token, NOW).is_err());
    assert_eq!(fs::read(&target).unwrap(), V1);
    assert_eq!(h.names(), vec!["moai"]);

    // A clock that is not a UTC time names no file.
    let h = Home::new();
    let token = h.token();
    for bad in ["not a time", "../../x", "", "2026-11-19 07:47:22"] {
        let e = switch::apply(&h.root, &token, bad).unwrap_err();
        assert!(e.contains("UTC"), "{bad}: {e}");
        assert_eq!(h.names(), vec!["moai"], "{bad}");
        assert_eq!(h.read(), V1);
    }
}

#[test]
fn a_failed_step_leaves_a_whole_file() {
    let _g = gate();
    // A folder that cannot be written (skipped when the process may write anyway, e.g. root).
    let h = Home::new();
    let token = h.token();
    fs::set_permissions(h.dir(), fs::Permissions::from_mode(0o555)).unwrap();
    let probe = fs::write(h.dir().join("probe"), b"x");
    if probe.is_err() {
        let e = switch::apply(&h.root, &token, NOW).unwrap_err();
        assert!(!e.is_empty());
        fs::set_permissions(h.dir(), fs::Permissions::from_mode(0o755)).unwrap();
        assert_eq!(h.read(), V1, "still the whole v1");
        assert_eq!(h.names(), vec!["moai"], "no saved copy, no temporary file");
    } else {
        let _ = fs::remove_file(h.dir().join("probe"));
        fs::set_permissions(h.dir(), fs::Permissions::from_mode(0o755)).unwrap();
    }
    // Faults after the first write (debug builds read the seam): the file stays v1, the
    // temporary file is gone, the saved copy is kept and named in the message.
    for point in ["switch-after-save", "switch-rename"] {
        let h = Home::new();
        let token = h.token();
        std::env::set_var("MOAI_CLI_TEST_FAULT", point);
        let r = switch::apply(&h.root, &token, NOW);
        std::env::remove_var("MOAI_CLI_TEST_FAULT");
        let e = r.unwrap_err();
        assert!(e.contains(point), "{point}: {e}");
        assert_eq!(h.read(), V1, "{point}: still the whole v1");
        assert_eq!(
            h.names(),
            vec!["moai", &format!("moai.before-{STAMP}")[..]],
            "{point}: no temporary file left"
        );
        assert!(
            e.contains(&h.saved(STAMP).display().to_string()),
            "{point}: names the saved copy: {e}"
        );
        assert_eq!(fs::read(h.saved(STAMP)).unwrap(), V1);
        // The next apply, with a new stamp, still works.
        let r = switch::apply(&h.root, &h.token(), "2026-11-19T07:48:00Z").unwrap();
        assert_eq!(r["state"], json!("applied"), "{point}: {r}");
        assert_eq!(h.read(), V2);
    }
}

// ---- AC-4 -----------------------------------------------------------------

#[test]
fn revert_puts_v1_back_and_refuses_anything_else() {
    let _g = gate();
    let h = Home::new();
    switch::apply(&h.root, &h.token(), NOW).unwrap();
    // Every refusal changes nothing.
    let state = |h: &Home| (h.read(), h.names());
    let before = state(&h);
    for bad in [
        "",
        "2026",
        "../x",
        "20261119T074722",
        "20261119T074722Zx",
        "20261119T0747227Z",
    ] {
        let e = switch::revert(&h.root, bad).unwrap_err();
        assert!(e.contains("stamp"), "{bad}: {e}");
        assert_eq!(state(&h), before, "{bad}");
    }
    let e = switch::revert(&h.root, "20250101T000000Z").unwrap_err();
    assert!(e.contains("no saved copy"), "{e}");
    // The saved copy is not v1 (changed by hand): refused, left as it is.
    fs::write(h.saved(STAMP), b"#!/bin/sh\nnot v1\n").unwrap();
    let e = switch::revert(&h.root, STAMP).unwrap_err();
    assert!(e.contains("not the v1"), "{e}");
    assert_eq!(fs::read(h.saved(STAMP)).unwrap(), b"#!/bin/sh\nnot v1\n");
    assert_eq!(h.read(), V2);
    // The saved copy is a link: refused.
    fs::remove_file(h.saved(STAMP)).unwrap();
    let real = h.root.join("v1-elsewhere");
    fs::write(&real, V1).unwrap();
    symlink(&real, h.saved(STAMP)).unwrap();
    assert!(switch::revert(&h.root, STAMP).is_err());
    assert_eq!(h.read(), V2);
    fs::remove_file(h.saved(STAMP)).unwrap();
    fs::write(h.saved(STAMP), V1).unwrap();
    fs::set_permissions(h.saved(STAMP), fs::Permissions::from_mode(0o755)).unwrap();
    // `moai` is not v2 any more (edited since): refused, the saved copy stays.
    h.put(b"#!/bin/sh\nmy own edit\n");
    let e = switch::revert(&h.root, STAMP).unwrap_err();
    assert!(e.contains("not the v2"), "{e}");
    assert_eq!(h.read(), b"#!/bin/sh\nmy own edit\n");
    assert_eq!(fs::read(h.saved(STAMP)).unwrap(), V1);
    // The good case: v2 back to v1 by one rename; the saved name is gone.
    h.put(V2);
    let r = switch::revert(&h.root, STAMP).unwrap();
    assert_eq!(r["state"], json!("reverted"), "{r}");
    assert_eq!(h.read(), V1);
    assert_eq!(
        fs::metadata(h.moai()).unwrap().permissions().mode() & 0o777,
        0o755
    );
    assert_eq!(h.names(), vec!["moai"]);
    // A second revert finds nothing.
    let e = switch::revert(&h.root, STAMP).unwrap_err();
    assert!(e.contains("no saved copy"), "{e}");
    assert_eq!(h.read(), V1);
    // And the preview offers v2 again.
    assert_eq!(h.part().unwrap()["action"], json!("replace"));
}

// ---- AC-6, library part (REQ-10) --------------------------------------------

#[test]
fn revert_switch_runs_alone_and_is_refused_beside_a_confirming_flag() {
    use moai_cli::update::update_command_switch;
    let _g = gate();
    let h = Home::new();
    switch::apply(&h.root, &h.token(), NOW).unwrap();
    let ask = |_: &str, _: &Value| Ok(false);
    let flags: [[Option<&str>; 4]; 4] = [
        [Some("x"), None, None, None],
        [None, Some("x"), None, None],
        [None, None, Some("x"), None],
        [None, None, None, Some("x")],
    ];
    for [c, cc, ch, cs] in flags {
        let e =
            update_command_switch(&h.root, c, cc, ch, cs, None, Some(STAMP), false, ask)
                .unwrap_err();
        assert!(e.contains("cannot be combined"), "{e}");
        assert_eq!(h.read(), V2, "nothing changed");
        assert!(h.saved(STAMP).exists());
    }
    // Alone it runs without any installed package: only the saved copy and `moai` are read.
    let r = update_command_switch(&h.root, None, None, None, None, None, Some(STAMP), false, ask)
        .unwrap();
    assert_eq!(r["mode"], json!("revert_switch"), "{r}");
    assert_eq!(r["switch_result"]["state"], json!("reverted"), "{r}");
    assert_eq!(h.read(), V1);
    // An apply result's `revert` line comes first in the revert order.
    let order = moai_cli::update::revert_order(&json!({
        "switch": "moai-cli update --revert-switch X", "home": "h", "agents": "a"
    }));
    assert_eq!(order, vec!["switch", "home", "agents"]);
}
