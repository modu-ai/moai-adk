//! #2723: `moai-cli home-state-plan` on temporary HOME folders only. The real HOME is never
//! read or written here: every case passes `--home <tempdir>`.
#![cfg(unix)]

mod common;
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};

const BIN: &str = crate::common::MOAI_BIN;
const NOW: &str = "2026-11-13T18:05:09Z";

fn write(root: &Path, rel: &str, body: &str) {
    let p = root.join(rel);
    fs::create_dir_all(p.parent().unwrap()).unwrap();
    fs::write(p, body).unwrap();
}

/// A temporary HOME whose `.moai` looks like the Go generation's.
fn go_home() -> (tempfile::TempDir, PathBuf) {
    let dir = tempfile::tempdir().unwrap();
    let home = fs::canonicalize(dir.path()).unwrap();
    write(
        &home,
        ".moai/user-assets.json",
        r#"{"schema_version": 1, "files": {}}"#,
    );
    write(&home, ".moai/state/active-sessions.json", "[]");
    write(&home, ".moai/credentials.yaml", "token: PLACEHOLDER\n");
    write(&home, ".moai/.env.glm", "PLACEHOLDER=1\n");
    write(&home, ".moai/config/participation.yaml", "x: 1\n");
    (dir, home)
}

fn plan_with(home: &Path, moai_home: Option<&str>) -> Value {
    let mut c = Command::new(BIN);
    c.args([
        "home-state-plan",
        "--home",
        home.to_str().unwrap(),
        "--now",
        NOW,
    ])
    .env_remove("MOAI_HOME");
    if let Some(m) = moai_home {
        c.env("MOAI_HOME", m);
    }
    let out = c.output().unwrap();
    let v: Value = serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("{}", String::from_utf8_lossy(&out.stderr)));
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn plan(home: &Path) -> Value {
    plan_with(home, None)
}

fn tree(root: &Path) -> Vec<(String, Vec<u8>)> {
    let mut out = Vec::new();
    let mut stack = vec![root.to_path_buf()];
    while let Some(d) = stack.pop() {
        for e in fs::read_dir(&d).unwrap().flatten() {
            let p = e.path();
            let m = fs::symlink_metadata(&p).unwrap();
            if m.is_dir() {
                stack.push(p);
            } else {
                out.push((p.display().to_string(), fs::read(&p).unwrap_or_default()));
            }
        }
    }
    out.sort();
    out
}

#[test]
fn a_go_home_offers_the_move_and_writes_nothing() {
    let (_d, home) = go_home();
    let before = tree(&home);
    let p = plan(&home);
    assert_eq!(p["state"]["generation"], json!("go"), "{p}");
    assert_eq!(
        p["state"]["go_markers"],
        json!([
            ".moai/user-assets.json",
            ".moai/state/active-sessions.json",
            ".moai/credentials.yaml"
        ])
    );
    assert_eq!(p["action"], json!("migrate"), "{p}");
    assert_eq!(p["blocks"], json!([]));
    assert_eq!(
        p["backup"],
        json!(home.join(".moai-backup").display().to_string())
    );
    assert_eq!(p["copy"]["default"], json!("none"));
    let cands = p["copy"]["candidates"].as_array().unwrap();
    assert!(cands.iter().all(|c| c["selected"] == json!(false)));
    let kind = |n: &str| {
        cands
            .iter()
            .find(|c| c["name"] == json!(n))
            .unwrap_or_else(|| panic!("{n}"))["kind"]
            .clone()
    };
    assert_eq!(kind(".env.glm"), json!("secret"));
    assert_eq!(kind("credentials.yaml"), json!("secret"));
    assert_eq!(kind("config"), json!("entry"));
    assert_eq!(p["writes"], json!(false));
    assert_eq!(tree(&home), before, "the preview writes nothing");
}

#[test]
fn an_existing_backup_is_never_overwritten() {
    let (_d, home) = go_home();
    fs::create_dir_all(home.join(".moai-backup")).unwrap();
    let p = plan(&home);
    assert_eq!(
        p["backup"],
        json!(home
            .join(".moai-backup-20261113T180509Z")
            .display()
            .to_string())
    );
    assert_eq!(p["action"], json!("migrate"));
    fs::create_dir_all(home.join(".moai-backup-20261113T180509Z")).unwrap();
    let p = plan(&home);
    assert_eq!(p["backup"], Value::Null);
    assert_eq!(p["action"], json!("blocked"));
    assert_eq!(p["blocks"], json!(["backup_name_taken"]));
}

#[test]
fn verdicts_other_than_go_never_offer_the_move() {
    // none, empty, rust: nothing to do.
    let dir = tempfile::tempdir().unwrap();
    let home = fs::canonicalize(dir.path()).unwrap();
    assert_eq!(plan(&home)["state"]["generation"], json!("none"));
    assert_eq!(plan(&home)["action"], json!("none"));
    fs::create_dir_all(home.join(".moai")).unwrap();
    assert_eq!(plan(&home)["state"]["generation"], json!("empty"));
    write(
        &home,
        ".moai/moai-cli.json",
        r#"{"schema_version":1,"generation":"moai-cli"}"#,
    );
    let p = plan(&home);
    assert_eq!(
        (p["state"]["generation"].clone(), p["action"].clone()),
        (json!("rust"), json!("none"))
    );
    // mixed and unknown: blocked, nothing moves.
    write(&home, ".moai/credentials.yaml", "t: x\n");
    let p = plan(&home);
    assert_eq!(
        (p["state"]["generation"].clone(), p["action"].clone()),
        (json!("mixed"), json!("blocked"))
    );
    let dir2 = tempfile::tempdir().unwrap();
    let home2 = fs::canonicalize(dir2.path()).unwrap();
    write(&home2, ".moai/notes.txt", "n\n");
    let p = plan(&home2);
    assert_eq!(
        (p["state"]["generation"].clone(), p["action"].clone()),
        (json!("unknown"), json!("blocked"))
    );
    // A user-assets.json that is not the Go manifest is not a marker.
    write(&home2, ".moai/user-assets.json", r#"{"files": []}"#);
    assert_eq!(plan(&home2)["state"]["go_markers"], json!([]));
    // A linked .moai is refused, never followed.
    let dir3 = tempfile::tempdir().unwrap();
    let home3 = fs::canonicalize(dir3.path()).unwrap();
    let (_g, gohome) = go_home();
    std::os::unix::fs::symlink(gohome.join(".moai"), home3.join(".moai")).unwrap();
    let p = plan(&home3);
    assert_eq!(
        (p["state"]["generation"].clone(), p["action"].clone()),
        (json!("refuse"), json!("blocked"))
    );
}

#[test]
fn a_live_go_session_blocks_the_move() {
    let (_d, home) = go_home();
    let lock = home.join(".moai/state/active-sessions.json.lock");
    fs::write(&lock, "").unwrap();
    let ready = home.join("locked");
    let script = format!(
        "import fcntl,time,pathlib\nf=open({:?},'a')\nfcntl.flock(f,fcntl.LOCK_EX)\npathlib.Path({:?}).write_text('1')\ntime.sleep(30)\n",
        lock.display().to_string(),
        ready.display().to_string()
    );
    let mut child = Command::new("python3")
        .args(["-I", "-c", &script])
        .current_dir(std::env::temp_dir())
        .stdout(Stdio::null())
        .spawn()
        .unwrap();
    let start = std::time::Instant::now();
    while !ready.exists() && start.elapsed().as_secs() < 20 {
        std::thread::sleep(std::time::Duration::from_millis(50));
    }
    let p = plan(&home);
    let _ = child.kill();
    let _ = child.wait();
    assert_eq!(p["action"], json!("blocked"), "{p}");
    // The holder also keeps the lock file open, so the open-file check may name it too.
    let blocks = p["blocks"].as_array().unwrap();
    assert!(blocks.contains(&json!("go_session_live")), "{p}");
    assert!(
        blocks
            .iter()
            .all(|b| b == "go_session_live" || b == "file_open_in_use"),
        "{p}"
    );
    // Released: the move is offered again.
    assert_eq!(plan(&home)["action"], json!("migrate"));
}

#[test]
fn a_process_inside_or_a_nested_checkout_blocks_the_move() {
    let (_d, home) = go_home();
    let mut child = Command::new("sleep")
        .arg("30")
        .current_dir(home.join(".moai/config"))
        .spawn()
        .unwrap();
    let p = plan(&home);
    let _ = child.kill();
    let _ = child.wait();
    // `lsof` lists a working folder among open files, so both checks may name it.
    let blocks = p["blocks"].as_array().unwrap();
    assert!(blocks.contains(&json!("in_use")), "{p}");
    assert!(
        blocks
            .iter()
            .all(|b| b == "in_use" || b == "file_open_in_use"),
        "{p}"
    );

    // A checkout outside the work folders, so this case names only the nested checkout
    // (`worktrees/` itself blocks on its own since #2811).
    write(&home, ".moai/clone/x/.git", "gitdir: /elsewhere\n");
    let p = plan(&home);
    assert_eq!(p["action"], json!("blocked"));
    assert_eq!(p["blocks"], json!(["nested_checkouts:1"]));
}

#[test]
fn moai_home_elsewhere_is_reported() {
    let (_d, home) = go_home();
    let p = plan_with(&home, Some("/somewhere/else/.moai"));
    assert!(
        p["warnings"].to_string().contains("MOAI_HOME is set to"),
        "{p}"
    );
    let same = home.join(".moai").display().to_string();
    assert_eq!(plan_with(&home, Some(&same))["warnings"], json!([]));
}

#[test]
fn a_relative_home_or_a_bad_clock_is_refused() {
    let out = Command::new(BIN)
        .args(["home-state-plan", "--home", "relative/home", "--now", NOW])
        .output()
        .unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], json!(false), "{v}");
    let (_d, home) = go_home();
    let out = Command::new(BIN)
        .args([
            "home-state-plan",
            "--home",
            home.to_str().unwrap(),
            "--now",
            "yesterday",
        ])
        .output()
        .unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], json!(false), "{v}");
}

/// #2771 audit F1 (6861649871): the Go tool holds the registry lock only while it rewrites the
/// file, so a live session shows as a row whose `pid` is alive while the lock is free. A dead
/// pid does not block; a registry that cannot be read as rows of pids blocks (fail-closed).
#[test]
fn a_registered_live_pid_blocks_and_a_dead_one_does_not() {
    let (_d, home) = go_home();
    let reg = home.join(".moai/state/active-sessions.json");
    let mut child = Command::new("sleep")
        .arg("30")
        .current_dir(&home)
        .spawn()
        .unwrap();
    let live = child.id();
    fs::write(&reg, format!(r#"[{{"session_id":"s","pid":{live}}}]"#)).unwrap();
    let p = plan(&home);
    assert_eq!(p["blocks"], json!(["go_session_live"]), "{p}");
    let _ = child.kill();
    let _ = child.wait();
    // The same pid, now ended: nothing blocks.
    assert_eq!(plan(&home)["action"], json!("migrate"));
    for bad in [
        r#"{"pid": 1}"#,
        r#"[{"session_id":"s"}]"#,
        r#"[{"pid": -4}]"#,
        "not json",
    ] {
        fs::write(&reg, bad).unwrap();
        let p = plan(&home);
        assert_eq!(
            p["blocks"],
            json!(["session_registry_unknown"]),
            "{bad}: {p}"
        );
    }
}

/// #2771 audit F2: a file inside `~/.moai` held open by a process whose working folder is
/// elsewhere blocks the move.
#[test]
fn a_file_held_open_inside_blocks_the_move() {
    let (_d, home) = go_home();
    write(&home, ".moai/logs/run.log", "x\n");
    let held = fs::File::open(home.join(".moai/logs/run.log")).unwrap();
    let mut child = Command::new("sleep")
        .arg("30")
        .current_dir(std::env::temp_dir())
        .stdin(held)
        .spawn()
        .unwrap();
    let p = plan(&home);
    let _ = child.kill();
    let _ = child.wait();
    assert_eq!(p["blocks"], json!(["file_open_in_use"]), "{p}");
}

/// #2811 D7/D8a: `worktrees/` and `projects/` are user work in the same `~/.moai`. They are
/// never copy candidates, and while either is present the move of the whole folder is
/// blocked (a rename would carry them into the backup). An empty Go HOME without them still
/// offers the move.
#[test]
fn work_folders_are_never_copied_and_block_the_move() {
    for name in ["worktrees", "projects"] {
        let (_d, home) = go_home();
        write(&home, &format!(".moai/{name}/keep.txt"), "work\n");
        let before = tree(&home);
        let p = plan(&home);
        assert_eq!(p["action"], json!("blocked"), "{name}: {p}");
        assert!(
            p["blocks"]
                .as_array()
                .unwrap()
                .contains(&json!(format!("work_folder_present:{name}"))),
            "{name}: {p}"
        );
        let c = p["copy"]["candidates"]
            .as_array()
            .unwrap()
            .iter()
            .find(|c| c["name"] == json!(name))
            .unwrap()
            .clone();
        assert_eq!(c["kind"], json!("work_folder"), "{c}");
        assert_eq!(c["selected"], json!(false), "{c}");
        assert_eq!(tree(&home), before, "the plan writes nothing");
    }
    let (_d, home) = go_home();
    assert_eq!(plan(&home)["action"], json!("migrate"));
}
