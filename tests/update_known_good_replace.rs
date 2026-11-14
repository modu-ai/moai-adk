//! #2103: a mismatched known-good copy is replaced through the 01-B engine's
//! temp+rename (`apply::publish_replace`), so replacing it swaps the inode instead of
//! truncating in place — a hard link that shares the corrupt copy's inode keeps its
//! own bytes. Mirrors tests/update_known_good.rs helpers; that file and
//! tests/update_cli.rs are untouched.

#![cfg(unix)]

mod common;
use moai_cli::digest;
use serde_json::{json, Value};
use std::{
    fs,
    io::Write as _,
    path::{Path, PathBuf},
    process::{Command, Stdio},
};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;
const HOST: &str = "claude";
const SKILL_V1: &str = "---\nname: moai\ndescription: fixture package\n---\n# MoAI\nbody v1\n";
const SKILL_V2: &str = "---\nname: moai\ndescription: fixture package\n---\n# MoAI\nbody v2\n";
const UPDATE_V1: &str = "# UPDATE\nv1\n";
const UPDATE_V2: &str = "# UPDATE\nv2\n";

fn file_hex(bytes: &[u8]) -> String {
    digest(bytes).trim_start_matches("sha256:").to_string()
}

fn write_pkg(dir: &Path, files: &[(&str, &str)]) {
    for (rel, body) in files {
        let p = dir.join(rel);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        fs::write(p, body).unwrap();
    }
}

fn setup() -> (TempDir, PathBuf, PathBuf, PathBuf) {
    let root = TempDir::new().unwrap();
    let rootp = root.path().to_path_buf();
    let home = rootp.join("home");
    fs::create_dir_all(&home).unwrap();
    // An earlier-root HOME: its state stays in `~/.moai-cli` (#2696 C4b2 gives a HOME with no
    // state `~/.moai`; that layout is pinned by home_root tests and update_cli).
    fs::create_dir_all(home.join(".moai-cli")).unwrap();
    let stage1 = rootp.join("stage-v1");
    let stage2 = rootp.join("stage-v2");
    write_pkg(
        &stage1,
        &[("SKILL.md", SKILL_V1), ("workflows/update.md", UPDATE_V1)],
    );
    write_pkg(
        &stage2,
        &[
            ("SKILL.md", SKILL_V2),
            ("workflows/update.md", UPDATE_V2),
            ("workflows/new.md", "# NEW\nadded in v2\n"),
        ],
    );
    (root, home, stage1, stage2)
}

fn run(args: &[&str], stdin: Option<&Value>, home: &Path) -> (bool, Value) {
    let mut child = Command::new(EXE)
        .args(args)
        .env("HOME", home)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_NOSYSTEM", "1")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    let mut pipe = child.stdin.take().unwrap();
    if let Some(v) = stdin {
        match pipe.write_all(v.to_string().as_bytes()) {
            Ok(()) => {}
            Err(e) if e.kind() == std::io::ErrorKind::BrokenPipe => {}
            Err(e) => panic!("stdin write failed: {e}"),
        }
    }
    drop(pipe);
    let out = child.wait_with_output().unwrap();
    let v: Value = serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("no envelope: {}", String::from_utf8_lossy(&out.stderr)));
    (out.status.success(), v)
}

fn plan_via_cli(home: &Path, stage: &Path) -> Value {
    let (ok, v) = run(
        &[
            "update-plan",
            home.to_str().unwrap(),
            stage.to_str().unwrap(),
        ],
        None,
        home,
    );
    assert!(ok, "{v}");
    v["result"].clone()
}

fn apply_ok(home: &Path, stage: &Path, plan: &Value) -> Value {
    let (ok, v) = run(
        &[
            "update-apply",
            home.to_str().unwrap(),
            stage.to_str().unwrap(),
            "--plan-digest",
            plan["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "fixture-approved",
        ],
        Some(plan),
        home,
    );
    assert!(ok, "{v}");
    v["result"].clone()
}

fn receipt_json(home: &Path) -> Value {
    let installations = home.join(".moai-cli").join("installations");
    let mut ids = fs::read_dir(&installations).unwrap();
    let id = ids.next().expect("one installation").unwrap().file_name();
    let bytes = fs::read(
        installations
            .join(id)
            .join("hosts")
            .join("moai")
            .join(format!("{HOST}.json")),
    )
    .unwrap();
    serde_json::from_slice(&bytes).unwrap()
}

fn known_good_dir(home: &Path, install_id: &str) -> PathBuf {
    home.join(".moai-cli")
        .join("installations")
        .join(install_id)
        .join("hosts")
        .join("moai")
        .join(HOST)
        .join("known-good")
}

/// AC-1: the corrupt copy is hard-linked to a file elsewhere; replacing the copy
/// swaps the digest path's inode (temp+rename) and the linked file keeps the old
/// bytes. The copy itself ends up holding the true pre-overwrite bytes.
#[test]
fn replacing_a_mismatched_copy_leaves_hard_linked_bytes_intact() {
    let (root, home, stage1, stage2) = setup();
    let v1_skill = SKILL_V1;
    let corrupt = b"corrupted bytes from a failed save\n";

    let plan1 = plan_via_cli(&home, &stage1);
    apply_ok(&home, &stage1, &plan1);
    let install_id = receipt_json(&home)["install_id"]
        .as_str()
        .unwrap()
        .to_string();

    // Plant a corrupt copy under the digest name and hard-link a file OUTSIDE the
    // home to the same inode, so an in-place truncate would corrupt it too.
    let copy = known_good_dir(&home, &install_id).join(file_hex(v1_skill.as_bytes()));
    fs::create_dir_all(copy.parent().unwrap()).unwrap();
    fs::write(&copy, corrupt).unwrap();
    let linked = root.path().join("linked-elsewhere.bin");
    fs::hard_link(&copy, &linked).unwrap();
    assert_eq!(fs::read(&linked).unwrap(), corrupt);

    let plan2 = plan_via_cli(&home, &stage2);
    let result = apply_ok(&home, &stage2, &plan2);
    assert_eq!(result["status"], json!("applied"), "{result}");

    // The digest path now holds the true bytes on a fresh inode; the hard-linked
    // file keeps the old inode's bytes untouched.
    assert_eq!(fs::read(&copy).unwrap(), v1_skill.as_bytes());
    assert_eq!(fs::read(&linked).unwrap(), corrupt);
    assert_eq!(receipt_json(&home)["status"], json!("committed"));
}
