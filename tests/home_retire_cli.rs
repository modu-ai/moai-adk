//! #2871 U2: `home-retire-plan` / `home-retire-apply` / `home-retire-revert` on temporary HOME
//! folders only. The real HOME is never read or written: every case passes `--home <tempdir>`,
//! secrets are synthetic names with placeholder or no content.
#![cfg(unix)]

mod common;
use serde_json::{json, Value};
use std::fs;
use std::os::unix::fs::{MetadataExt, PermissionsExt};
use std::path::{Path, PathBuf};
use std::process::{Child, Command, Stdio};

const BIN: &str = crate::common::MOAI_BIN;
const NOW: &str = "2026-11-14T12:00:00Z";
const STAMP: &str = "20261114T120000Z";
const LATER: &str = "2026-11-14T13:00:00Z";
const PERM: &str = "user approved in the fixture";

fn write(root: &Path, rel: &str, body: &str) {
    let p = root.join(rel);
    fs::create_dir_all(p.parent().unwrap()).unwrap();
    fs::write(p, body).unwrap();
}

/// A Go-generation `~/.moai` with an entry of every kind the table names.
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
    write(&home, ".moai/.env.glm", "GLM_API_KEY=PLACEHOLDER\n");
    write(&home, ".moai/.env.typesafe", "KEY=PLACEHOLDER\n");
    write(&home, ".moai/cache/a.bin", "x");
    write(&home, ".moai/Logs/go.log", "x\n");
    write(&home, ".moai/run/x.pid", "1\n");
    write(
        &home,
        ".moai/config/sections/statusline.yaml",
        "preset: a\n",
    );
    write(&home, ".moai/specs/SPEC-1/spec.md", "# user spec\n");
    write(&home, ".moai/worktrees/projects/other/wt1/README", "x\n");
    write(&home, ".moai/my notes/n.txt", "spaces in the name\n");
    write(&home, ".moai/reports/r.md", "# report\n");
    (dir, home)
}

fn call(args: &[&str], envs: &[(&str, &str)]) -> Value {
    let out = Command::new(BIN)
        .args(args)
        .env_remove("MOAI_HOME")
        .envs(envs.iter().copied())
        .stdin(Stdio::null())
        .output()
        .unwrap();
    serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("{}", String::from_utf8_lossy(&out.stderr)))
}

fn plan_with(home: &Path, now: &str, envs: &[(&str, &str)]) -> Value {
    let v = call(
        &[
            "home-retire-plan",
            "--home",
            home.to_str().unwrap(),
            "--now",
            now,
        ],
        envs,
    );
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn plan(home: &Path) -> Value {
    plan_with(home, NOW, &[])
}

fn apply_with(home: &Path, now: &str, digest: &str, perm: &str, envs: &[(&str, &str)]) -> Value {
    call(
        &[
            "home-retire-apply",
            "--home",
            home.to_str().unwrap(),
            "--now",
            now,
            "--plan-digest",
            digest,
            "--permission-reference",
            perm,
        ],
        envs,
    )
}

fn apply(home: &Path, digest: &str) -> Value {
    apply_with(home, NOW, digest, PERM, &[])
}

fn revert(home: &Path, extra: &[&str]) -> Value {
    let mut args = vec![
        "home-retire-revert",
        "--home",
        home.to_str().unwrap(),
        "--permission-reference",
        PERM,
    ];
    args.extend_from_slice(extra);
    call(&args, &[])
}

fn entry<'a>(plan: &'a Value, name: &str) -> &'a Value {
    plan["entries"]
        .as_array()
        .unwrap()
        .iter()
        .find(|e| e["name"] == name)
        .unwrap_or_else(|| panic!("{name} missing: {plan}"))
}

fn act(plan: &Value, name: &str) -> (String, String) {
    let e = entry(plan, name);
    (
        e["action"].as_str().unwrap().to_string(),
        e["reason"].as_str().unwrap().to_string(),
    )
}

/// Every entry under `root`: relative path, kind, mode and, for a file, its bytes. Links are
/// listed by target, never followed.
fn tree(root: &Path) -> Vec<(String, String, u32, Vec<u8>)> {
    fn walk(dir: &Path, rel: &str, out: &mut Vec<(String, String, u32, Vec<u8>)>) {
        let mut names: Vec<_> = fs::read_dir(dir)
            .unwrap()
            .map(|e| e.unwrap().file_name())
            .collect();
        names.sort();
        for n in names {
            let p = dir.join(&n);
            let r = format!("{rel}/{}", n.to_string_lossy());
            let m = fs::symlink_metadata(&p).unwrap();
            let mode = m.permissions().mode();
            if m.file_type().is_symlink() {
                let t = fs::read_link(&p).unwrap();
                out.push((
                    r,
                    "link".into(),
                    mode,
                    t.to_string_lossy().as_bytes().to_vec(),
                ));
            } else if m.is_dir() {
                out.push((r.clone(), "dir".into(), mode, Vec::new()));
                walk(&p, &r, out);
            } else {
                out.push((r, "file".into(), mode, fs::read(&p).unwrap()));
            }
        }
    }
    let mut out = Vec::new();
    walk(root, "", &mut out);
    out
}

struct Kill(Child);
impl Drop for Kill {
    fn drop(&mut self) {
        let _ = self.0.kill();
        let _ = self.0.wait();
    }
}

fn retired(home: &Path) -> PathBuf {
    home.join(".moai-go-retired")
}

#[test]
fn retire_preview_classifies_every_entry_by_the_table_and_writes_nothing() {
    let (_d, home) = go_home();
    let before = tree(&home);
    let p = plan(&home);
    assert_eq!(p["action"], json!("retire"), "{p}");
    assert_eq!(p["writes"], json!(false));
    assert_eq!(p["state"]["generation"], json!("go"));
    // Go tool assets, settings and plain secrets retire; the name `Logs` matches `logs`
    // whatever its case.
    for name in [
        "user-assets.json",
        "state",
        "cache",
        "Logs",
        "run",
        "config",
        "credentials.yaml",
        ".env.typesafe",
    ] {
        assert_eq!(act(&p, name).0, "retire", "{name}: {p}");
    }
    // The destination is inside the dated folder outside `~/.moai`.
    let dest = home.join(format!(".moai-go-retired/{STAMP}/cache"));
    assert_eq!(
        entry(&p, "cache")["destination"],
        json!(dest.to_str().unwrap())
    );
    // User data, unknown owners, the unknown name and the merge-candidate key stay.
    assert_eq!(act(&p, "specs"), ("keep".into(), "user_data".into()));
    assert_eq!(act(&p, "worktrees"), ("keep".into(), "user_data".into()));
    assert_eq!(act(&p, "reports"), ("keep".into(), "owner_unknown".into()));
    assert_eq!(act(&p, "my notes"), ("keep".into(), "not_in_table".into()));
    assert_eq!(
        act(&p, ".env.glm"),
        ("keep".into(), "merge_candidate_kept".into())
    );
    assert_eq!(entry(&p, "config")["merge_candidate"], json!(true));
    assert_eq!(entry(&p, "specs")["destination"], Value::Null);
    assert_eq!(tree(&home), before, "the preview wrote something");
}

#[test]
fn retire_apply_moves_only_the_retire_entries_and_changes_no_byte() {
    let (_d, home) = go_home();
    // A secret nobody can open: the rename must not need to read it.
    let cred = home.join(".moai/credentials.yaml");
    fs::set_permissions(&cred, fs::Permissions::from_mode(0o000)).unwrap();
    let ino = fs::metadata(&cred).unwrap().ino();
    let p = plan(&home);
    let a = apply(&home, p["plan_digest"].as_str().unwrap());
    assert_eq!(a["ok"], json!(true), "{a}");
    let r = &a["result"];
    assert_eq!(r["state"], json!("applied"));
    assert_eq!(r["journal_state"], json!("committed"));
    let folder = retired(&home).join(STAMP);
    let mut kept: Vec<String> = fs::read_dir(home.join(".moai"))
        .unwrap()
        .map(|e| e.unwrap().file_name().into_string().unwrap())
        .collect();
    kept.sort();
    assert_eq!(
        kept,
        [".env.glm", "my notes", "reports", "specs", "worktrees"],
        "{r}"
    );
    let mut moved: Vec<String> = fs::read_dir(&folder)
        .unwrap()
        .map(|e| e.unwrap().file_name().into_string().unwrap())
        .collect();
    moved.sort();
    assert_eq!(
        moved,
        [
            ".env.typesafe",
            "Logs",
            "cache",
            "config",
            "credentials.yaml",
            "run",
            "state",
            "user-assets.json"
        ]
    );
    // Same file, same mode, same bytes: a rename.
    let moved_cred = folder.join("credentials.yaml");
    assert_eq!(fs::metadata(&moved_cred).unwrap().ino(), ino);
    assert_eq!(
        fs::metadata(&moved_cred).unwrap().permissions().mode() & 0o777,
        0o000
    );
    assert_eq!(
        fs::read(folder.join("config/sections/statusline.yaml")).unwrap(),
        b"preset: a\n"
    );
    // The manifest names every moved item and its merge flag; the journal is committed.
    let manifest: Value = serde_json::from_slice(
        &fs::read(retired(&home).join(format!("retire-manifest-{STAMP}.json"))).unwrap(),
    )
    .unwrap();
    let items = manifest["items"].as_array().unwrap();
    assert_eq!(items.len(), 8, "{manifest}");
    let config = items.iter().find(|i| i["name"] == "config").unwrap();
    assert_eq!(config["merge_candidate"], json!(true));
    assert_eq!(config["class"], json!("config-guide"));
    assert!(retired(&home)
        .join(format!("move-journal-{STAMP}.json"))
        .is_file());
    fs::set_permissions(&moved_cred, fs::Permissions::from_mode(0o600)).unwrap();
}

#[test]
fn retire_rerun_after_apply_is_none_and_a_new_go_folder_retires_again() {
    let (_d, home) = go_home();
    let p = plan(&home);
    assert_eq!(
        apply(&home, p["plan_digest"].as_str().unwrap())["ok"],
        json!(true)
    );
    let again = plan_with(&home, LATER, &[]);
    assert_eq!(again["action"], json!("none"), "{again}");
    assert_eq!(again["counts"]["retire"], json!(0));
    // The Go generation recreates its run data (EnsureHomeLayout): it retires once more, into
    // its own dated folder, and the first backup stays.
    write(
        &home,
        ".moai/user-assets.json",
        r#"{"schema_version": 1, "files": {}}"#,
    );
    write(&home, ".moai/cache/b.bin", "y");
    let p2 = plan_with(&home, LATER, &[]);
    assert_eq!(p2["action"], json!("retire"), "{p2}");
    let a2 = apply_with(&home, LATER, p2["plan_digest"].as_str().unwrap(), PERM, &[]);
    assert_eq!(a2["ok"], json!(true), "{a2}");
    assert!(retired(&home).join(STAMP).join("cache/a.bin").is_file());
    assert!(retired(&home)
        .join("20261114T130000Z/cache/b.bin")
        .is_file());
}

#[test]
fn retire_apply_refuses_a_stale_digest_a_missing_reference_and_another_clock() {
    let (_d, home) = go_home();
    let p = plan(&home);
    let digest = p["plan_digest"].as_str().unwrap();
    let before = tree(&home);
    for (why, v) in [
        ("wrong digest", apply(&home, "0000")),
        ("no reference", apply_with(&home, NOW, digest, "", &[])),
        (
            "placeholder reference",
            apply_with(&home, NOW, digest, "placeholder", &[]),
        ),
        // Another `--now` names another dated folder: the preview's digest no longer holds.
        ("another clock", apply_with(&home, LATER, digest, PERM, &[])),
    ] {
        assert_eq!(v["ok"], json!(false), "{why}: {v}");
    }
    assert_eq!(tree(&home), before, "a refused apply wrote something");
    // A top-level entry appeared after the preview: the digest is stale.
    write(&home, ".moai/appeared-later.bin", "z");
    let stale = apply(&home, digest);
    assert_eq!(stale["ok"], json!(false), "{stale}");
    assert!(!retired(&home).exists());
}

#[test]
fn retire_in_use_entries_stay_and_an_unreadable_process_table_moves_nothing() {
    let (_d, home) = go_home();
    let _cwd = Kill(
        Command::new("sleep")
            .arg("60")
            .current_dir(home.join(".moai/Logs"))
            .spawn()
            .unwrap(),
    );
    let p = plan(&home);
    assert_eq!(
        act(&p, "Logs"),
        ("skip-in-use".into(), "process_cwd".into())
    );
    assert_eq!(act(&p, "cache").0, "retire");
    assert_eq!(entry(&p, "Logs")["destination"], Value::Null);
    // The apply re-checks: a stop-and-go program that appeared after the preview is skipped
    // by the journal, not moved.
    let a = apply(&home, p["plan_digest"].as_str().unwrap());
    assert_eq!(a["ok"], json!(true), "{a}");
    assert!(home.join(".moai/Logs/go.log").is_file());
    assert!(!retired(&home).join(STAMP).join("Logs").exists());
    // A machine whose process table cannot be read moves nothing (fail-closed).
    let (_d2, home2) = go_home();
    let p2 = plan_with(
        &home2,
        NOW,
        &[("MOAI_CLI_TEST_PROCESS_CHECK", "unavailable")],
    );
    assert_eq!(p2["counts"]["retire"], json!(0), "{p2}");
    assert_eq!(act(&p2, "cache").0, "skip-in-use");
}

#[test]
fn retire_blocks_when_the_machine_or_the_folders_are_not_safe() {
    // MOAI_HOME points elsewhere: the Go state may live there.
    let (_d, home) = go_home();
    let p = plan_with(&home, NOW, &[("MOAI_HOME", "/somewhere/else")]);
    assert_eq!(p["action"], json!("blocked"));
    assert!(
        p["blocks"].to_string().contains("moai_home_elsewhere"),
        "{p}"
    );
    // A live Go session.
    let (_d, home) = go_home();
    let me = std::process::id();
    write(
        &home,
        ".moai/state/active-sessions.json",
        &format!(r#"[{{"pid": {me}}}]"#),
    );
    let p = plan(&home);
    assert_eq!(p["action"], json!("blocked"));
    assert!(p["blocks"].to_string().contains("go_session_live"), "{p}");
    // The retired folder is a link.
    let (_d, home) = go_home();
    fs::create_dir(home.join("elsewhere")).unwrap();
    std::os::unix::fs::symlink(home.join("elsewhere"), retired(&home)).unwrap();
    let p = plan(&home);
    assert_eq!(p["action"], json!("blocked"));
    assert!(
        p["blocks"]
            .to_string()
            .contains("retired_root_not_a_folder"),
        "{p}"
    );
    // The dated folder is taken.
    let (_d, home) = go_home();
    fs::create_dir_all(retired(&home).join(STAMP)).unwrap();
    let p = plan(&home);
    assert_eq!(p["action"], json!("blocked"));
    assert!(p["blocks"].to_string().contains("stamp_taken"), "{p}");
    // Other generations have nothing to retire.
    let dir = tempfile::tempdir().unwrap();
    let home = fs::canonicalize(dir.path()).unwrap();
    assert_eq!(plan(&home)["action"], json!("none"));
}

/// Boundary: 500 entries are listed, 501 are not fully listed and block.
#[test]
fn retire_entry_limit_is_500() {
    let (_d, home) = go_home();
    let moai = home.join(".moai");
    let have = fs::read_dir(&moai).unwrap().count();
    for i in 0..(500 - have) {
        fs::write(moai.join(format!("f{i:04}")), "x").unwrap();
    }
    let p = plan(&home);
    assert_eq!(p["total"], json!(500));
    assert_eq!(p["action"], json!("retire"), "{p}");
    fs::write(moai.join("one-more"), "x").unwrap();
    let p = plan(&home);
    assert_eq!(p["total"], json!(501));
    assert_eq!(p["action"], json!("blocked"));
    assert!(p["blocks"].to_string().contains("not_fully_listed"));
}

#[test]
fn retire_revert_restores_the_tree_byte_for_byte_and_a_second_one_changes_nothing() {
    let (_d, home) = go_home();
    let before = tree(&home);
    let p = plan(&home);
    assert_eq!(
        apply(&home, p["plan_digest"].as_str().unwrap())["ok"],
        json!(true)
    );
    assert_ne!(tree(&home), before);
    let r = revert(&home, &[]);
    assert_eq!(r["ok"], json!(true), "{r}");
    assert_eq!(r["result"]["state"], json!("reverted"));
    // Everything but the empty dated folder and the two records is back.
    let after: Vec<_> = tree(&home)
        .into_iter()
        .filter(|(n, ..)| !n.starts_with("/.moai-go-retired"))
        .collect();
    assert_eq!(after, before.to_vec());
    let again = revert(&home, &[]);
    assert_eq!(again["ok"], json!(true), "{again}");
    assert_eq!(again["result"]["changed"], json!(false));
    // A refused reference writes nothing.
    assert_eq!(
        call(
            &[
                "home-retire-revert",
                "--home",
                home.to_str().unwrap(),
                "--permission-reference",
                ""
            ],
            &[]
        )["ok"],
        json!(false)
    );
}

#[test]
fn retire_revert_leaves_an_entry_that_was_replaced_since() {
    let (_d, home) = go_home();
    let p = plan(&home);
    assert_eq!(
        apply(&home, p["plan_digest"].as_str().unwrap())["ok"],
        json!(true)
    );
    // Go made a new `cache` meanwhile: the retired one must not land on top of it.
    write(&home, ".moai/cache/new.bin", "n");
    let r = revert(&home, &["--stamp", STAMP]);
    assert_eq!(r["ok"], json!(true), "{r}");
    assert_eq!(r["result"]["state"], json!("revert_partial"), "{r}");
    assert!(home.join(".moai/cache/new.bin").is_file());
    assert!(!home.join(".moai/cache/a.bin").exists());
    assert!(retired(&home).join(STAMP).join("cache/a.bin").is_file());
    assert!(
        home.join(".moai/run/x.pid").is_file(),
        "the others came back"
    );
}

/// A preview that is blocked has a digest too; the apply must still refuse to move anything.
#[test]
fn retire_apply_refuses_a_blocked_preview_even_with_its_own_digest() {
    let (_d, home) = go_home();
    let me = std::process::id();
    write(
        &home,
        ".moai/state/active-sessions.json",
        &format!(r#"[{{"pid": {me}}}]"#),
    );
    let p = plan(&home);
    assert_eq!(p["action"], json!("blocked"), "{p}");
    let before = tree(&home);
    let a = apply(&home, p["plan_digest"].as_str().unwrap());
    assert_eq!(a["ok"], json!(false), "{a}");
    assert_eq!(tree(&home), before, "a blocked apply wrote something");
    assert!(!retired(&home).exists());
}

/// A folder that holds another checkout stays: renaming it would orphan the registration.
#[test]
fn retire_keeps_a_folder_that_holds_a_nested_checkout() {
    let (_d, home) = go_home();
    fs::create_dir_all(home.join(".moai/cache/proj/.git")).unwrap();
    let p = plan(&home);
    assert_eq!(act(&p, "cache"), ("keep".into(), "nested_checkout".into()));
    assert_eq!(act(&p, "run").0, "retire");
    let a = apply(&home, p["plan_digest"].as_str().unwrap());
    assert_eq!(a["ok"], json!(true), "{a}");
    assert!(home.join(".moai/cache/proj/.git").is_dir());
}

/// Two retires, then a revert without `--stamp`: the newest one comes back, the older stays.
#[test]
fn retire_revert_without_a_stamp_undoes_the_newest_retire_only() {
    let (_d, home) = go_home();
    let p = plan(&home);
    assert_eq!(
        apply(&home, p["plan_digest"].as_str().unwrap())["ok"],
        json!(true)
    );
    write(
        &home,
        ".moai/user-assets.json",
        r#"{"schema_version": 1, "files": {}}"#,
    );
    write(&home, ".moai/cache/b.bin", "y");
    let p2 = plan_with(&home, LATER, &[]);
    let a2 = apply_with(&home, LATER, p2["plan_digest"].as_str().unwrap(), PERM, &[]);
    assert_eq!(a2["ok"], json!(true), "{a2}");
    let r = revert(&home, &[]);
    assert_eq!(r["ok"], json!(true), "{r}");
    assert_eq!(r["result"]["state"], json!("reverted"), "{r}");
    assert!(
        home.join(".moai/cache/b.bin").is_file(),
        "the newest came back"
    );
    assert!(
        retired(&home).join(STAMP).join("cache/a.bin").is_file(),
        "the older stays retired"
    );
    // Naming the older one brings back what it can; `cache` is occupied again, so it stays.
    let r = revert(&home, &["--stamp", STAMP]);
    assert_eq!(r["result"]["state"], json!("revert_partial"), "{r}");
    assert!(home.join(".moai/run/x.pid").is_file());
    assert!(home.join(".moai/cache/b.bin").is_file());
}

/// The dated backup folder is private to the user.
#[test]
fn retire_makes_the_backup_folders_private() {
    let (_d, home) = go_home();
    let p = plan(&home);
    assert_eq!(
        apply(&home, p["plan_digest"].as_str().unwrap())["ok"],
        json!(true)
    );
    for dir in [retired(&home), retired(&home).join(STAMP)] {
        assert_eq!(
            fs::metadata(&dir).unwrap().permissions().mode() & 0o777,
            0o700,
            "{}",
            dir.display()
        );
    }
}

/// A name that is not valid UTF-8 cannot be classified: the whole preview is blocked. (A
/// file system that refuses such a name, like APFS, cannot hold the case; the Linux CI does.)
#[test]
fn retire_blocks_on_a_name_it_cannot_read() {
    use std::os::unix::ffi::OsStrExt;
    let (_d, home) = go_home();
    let bad = std::ffi::OsStr::from_bytes(b"bad-\xff-name");
    if fs::write(home.join(".moai").join(bad), "x").is_err() {
        return;
    }
    let p = plan(&home);
    assert_eq!(p["action"], json!("blocked"), "{p}");
    assert!(p["blocks"].to_string().contains("unreadable_names"), "{p}");
}

/// A HOME with this generation's marker and Go items left over (`rust-first`) retires those
/// Go items too, but keeps the names both generations write.
#[test]
fn retire_acts_on_a_rust_first_home_and_keeps_shared_names() {
    let (_d, home) = go_home();
    write(
        &home,
        ".moai/moai-cli.json",
        r#"{"schema_version":2,"generation":"moai-cli"}"#,
    );
    let p = plan(&home);
    assert_eq!(p["state"]["generation"], json!("rust-first"), "{p}");
    assert_eq!(p["action"], json!("retire"), "{p}");
    assert_eq!(act(&p, "cache").0, "retire");
    assert_eq!(act(&p, "state"), ("keep".into(), "shared_name".into()));
    assert_eq!(act(&p, "config"), ("keep".into(), "shared_name".into()));
    assert_eq!(
        act(&p, "moai-cli.json"),
        ("keep".into(), "this_generation".into())
    );
}

/// A session lock that cannot be judged (here a folder where the file should be) blocks.
#[test]
fn retire_blocks_when_the_session_lock_cannot_be_judged() {
    let (_d, home) = go_home();
    fs::create_dir_all(home.join(".moai/state/active-sessions.json.lock")).unwrap();
    let p = plan(&home);
    assert_eq!(p["action"], json!("blocked"), "{p}");
    assert!(
        p["blocks"].to_string().contains("session_lock_unknown"),
        "{p}"
    );
}
