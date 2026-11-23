//! SPEC-20261123-home-state-retire: `home-archive-plan` / `-apply` / `-revert` archive the
//! earlier HOME state folder `~/.moai-cli` whole under `~/.moai/archive/` by one rename. Every
//! case runs on temporary HOME folders passed with `--home`; the real HOME is never read or
//! written, secrets are synthetic names with placeholder content.
#![cfg(unix)]

mod common;
#[path = "common/spawn_cap.rs"]
mod spawn_cap;
use serde_json::{json, Value};
use std::fs;
use std::os::unix::fs::{symlink, PermissionsExt};
use std::path::{Path, PathBuf};
use std::process::{Child, Command, Stdio};

const BIN: &str = crate::common::MOAI_BIN;
const NOW: &str = "2026-11-23T12:00:00Z";
const STAMP: &str = "20261123T120000Z";
const LATER: &str = "2026-11-23T13:00:00Z";
const PERM: &str = "user approved in the fixture";
const FAULT: &str = "MOAI_CLI_TEST_STATE_RETIRE_FAULT";

type Tree = Vec<(String, String, u32, Vec<u8>)>;

fn write(root: &Path, rel: &str, body: &str) {
    let p = root.join(rel);
    fs::create_dir_all(p.parent().unwrap()).unwrap();
    fs::write(p, body).unwrap();
}

fn chmod(p: &Path, mode: u32) {
    fs::set_permissions(p, fs::Permissions::from_mode(mode)).unwrap();
}

/// A HOME with the earlier folder (a vault with owner-only modes, a version folder, a PATH link,
/// receipts, state, a loose file and a top-level link) beside a Go-generation `~/.moai`.
fn earlier_home() -> (tempfile::TempDir, PathBuf) {
    let dir = tempfile::tempdir().unwrap();
    let home = fs::canonicalize(dir.path()).unwrap();
    write(&home, ".moai-cli/vault/_shared/glm.key", "PLACEHOLDER\n");
    chmod(&home.join(".moai-cli/vault"), 0o700);
    chmod(&home.join(".moai-cli/vault/_shared"), 0o700);
    chmod(&home.join(".moai-cli/vault/_shared/glm.key"), 0o600);
    write(
        &home,
        ".moai-cli/versions/4.2647.1/bin/moai-cli",
        "#!/bin/sh\n",
    );
    chmod(
        &home.join(".moai-cli/versions/4.2647.1/bin/moai-cli"),
        0o755,
    );
    fs::create_dir_all(home.join(".moai-cli/path")).unwrap();
    symlink(
        "../versions/4.2647.1/bin/moai-cli",
        home.join(".moai-cli/path/moai-cli"),
    )
    .unwrap();
    write(&home, ".moai-cli/receipts/r.json", "{}\n");
    write(&home, ".moai-cli/state/usage-limits.json", "{}\n");
    write(&home, ".moai-cli/note.txt", "a loose file\n");
    symlink("versions", home.join(".moai-cli/latest")).unwrap();
    write(&home, ".moai/state/active-sessions.json", "[]\n");
    write(&home, ".moai/.env.glm", "GLM_API_KEY=PLACEHOLDER\n");
    write(
        &home,
        ".moai/config/sections/statusline.yaml",
        "preset: a\n",
    );
    write(
        &home,
        ".zshrc",
        "export PATH=\"$HOME/.moai-cli/path:$PATH\"\n",
    );
    (dir, home)
}

fn call(args: &[&str], envs: &[(&str, &str)]) -> Value {
    let _permit = spawn_cap::acquire();
    let out = Command::new(BIN)
        .args(args)
        .env_remove("MOAI_HOME")
        .envs([crate::common::process_table()])
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
            "home-archive-plan",
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

fn apply_with(
    home: &Path,
    now: &str,
    key: &str,
    digest: &str,
    perm: &str,
    envs: &[(&str, &str)],
) -> Value {
    call(
        &[
            "home-archive-apply",
            "--home",
            home.to_str().unwrap(),
            "--now",
            now,
            key,
            digest,
            "--permission-reference",
            perm,
        ],
        envs,
    )
}

fn apply(home: &Path, digest: &str) -> Value {
    apply_with(home, NOW, "--plan-digest", digest, PERM, &[])
}

fn revert_with(home: &Path, perm: &str) -> Value {
    call(
        &[
            "home-archive-revert",
            "--home",
            home.to_str().unwrap(),
            "--permission-reference",
            perm,
        ],
        &[],
    )
}

fn revert(home: &Path) -> Value {
    revert_with(home, PERM)
}

fn digest_of(p: &Value) -> String {
    p["plan_digest"].as_str().unwrap().to_string()
}

/// Every entry under `root`: relative path, kind, mode and, for a file, its bytes. Links are
/// listed by target, never followed.
fn tree(root: &Path) -> Tree {
    fn walk(dir: &Path, rel: &str, out: &mut Tree) {
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

fn exists(p: &Path) -> bool {
    fs::symlink_metadata(p).is_ok()
}

/// The earlier state folder of `home`: the one name the tests below spell through one place.
fn older(home: &Path) -> PathBuf {
    home.join(".moai-cli")
}

/// The Go-generation `~/.moai` of `home`.
fn moai(home: &Path) -> PathBuf {
    home.join(".moai")
}

fn archive(home: &Path) -> PathBuf {
    home.join(".moai/archive").join(format!("moai-cli-{STAMP}"))
}

struct Kill(Child);
impl Drop for Kill {
    fn drop(&mut self) {
        let _ = self.0.kill();
        let _ = self.0.wait();
    }
}

/// A program that keeps `dir` in use until the test drops it: `cat` started in `dir`, waiting
/// on a pipe nobody writes to (no clock decides how long it lasts).
fn in_use(dir: PathBuf) -> Kill {
    crate::common::use_real_process_table();
    Kill(
        Command::new("cat")
            .stdin(Stdio::piped())
            .stdout(Stdio::null())
            .stderr(Stdio::null())
            .current_dir(dir)
            .spawn()
            .unwrap(),
    )
}

/// A program that holds `file` open (descriptor 3) while its working folder lies elsewhere, so
/// only the open-file check can see it. The test reads the program's `ready` line before it goes
/// on, so no clock decides whether the file is open yet.
fn holding_open(file: &Path) -> Kill {
    use std::io::{BufRead, BufReader};
    crate::common::use_real_process_table();
    let mut child = Command::new("sh")
        .args(["-c", "exec 3<\"$1\"; echo ready; exec cat", "sh"])
        .arg(file)
        .current_dir(std::env::temp_dir())
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::null())
        .spawn()
        .unwrap();
    let mut line = String::new();
    BufReader::new(child.stdout.take().unwrap())
        .read_line(&mut line)
        .unwrap();
    assert_eq!(line.trim(), "ready");
    Kill(child)
}

/// The earlier home with `extra` more loose files, so the folder holds exactly `total` entries.
fn home_with_entries(total: usize) -> (tempfile::TempDir, PathBuf) {
    let (d, home) = earlier_home();
    let have = fs::read_dir(older(&home)).unwrap().count();
    for i in 0..total - have {
        write(&home, &format!(".moai-cli/bulk-{i:04}.txt"), "x\n");
    }
    assert_eq!(fs::read_dir(older(&home)).unwrap().count(), total);
    (d, home)
}

fn manifest(home: &Path) -> PathBuf {
    moai(home).join(format!("archive/state-retire-manifest-{STAMP}.json"))
}

fn journal(home: &Path) -> PathBuf {
    moai(home).join(format!("archive/move-journal-{STAMP}.json"))
}

fn blocks(p: &Value) -> Vec<String> {
    p["blocks"]
        .as_array()
        .unwrap_or_else(|| panic!("no blocks in {p}"))
        .iter()
        .map(|b| b.as_str().unwrap().to_string())
        .collect()
}

fn refused(v: &Value) -> String {
    assert_eq!(v["ok"], json!(false), "expected a refusal: {v}");
    v["error"]["message"]
        .as_str()
        .unwrap_or_default()
        .to_string()
}

// ---- the preview -----------------------------------------------------------------------------

#[test]
fn the_preview_lists_every_entry_plans_one_move_and_writes_nothing() {
    let (_d, home) = earlier_home();
    let before = tree(&home);
    let p = plan(&home);
    assert_eq!(p["action"], json!("retire"), "{p}");
    assert_eq!(p["writes"], json!(false));
    assert_eq!(blocks(&p), Vec::<String>::new());
    assert_eq!(
        p["destination"],
        json!(archive(&home).display().to_string()),
        "{p}"
    );
    let entries = p["entries"].as_array().unwrap();
    let names: Vec<&str> = entries
        .iter()
        .map(|e| e["name"].as_str().unwrap())
        .collect();
    assert_eq!(
        names,
        ["latest", "note.txt", "path", "receipts", "state", "vault", "versions"]
    );
    let kind = |n: &str| {
        entries.iter().find(|e| e["name"] == n).unwrap()["kind"]
            .as_str()
            .unwrap()
            .to_string()
    };
    assert_eq!(kind("latest"), "link");
    assert_eq!(kind("note.txt"), "file");
    assert_eq!(kind("vault"), "folder");
    let vault = entries.iter().find(|e| e["name"] == "vault").unwrap();
    assert_eq!(
        vault["mode"],
        json!("700"),
        "the mode is read by lstat: {vault}"
    );
    assert_eq!(before, tree(&home), "a preview changes nothing");
    assert!(!exists(&home.join(".moai/archive")));
}

#[test]
fn the_preview_of_a_home_without_the_earlier_folder_is_none() {
    let dir = tempfile::tempdir().unwrap();
    let home = fs::canonicalize(dir.path()).unwrap();
    write(&home, ".moai/state/active-sessions.json", "[]\n");
    let p = plan(&home);
    assert_eq!(p["action"], json!("none"), "{p}");
    let v = apply(&home, &digest_of(&p));
    refused(&v);
    assert!(!exists(&home.join(".moai/archive")));
}

// ---- blocked cases ---------------------------------------------------------------------------

#[test]
fn blocked_when_the_earlier_folder_is_a_link_or_a_plain_file() {
    let dir = tempfile::tempdir().unwrap();
    let home = fs::canonicalize(dir.path()).unwrap();
    fs::create_dir_all(home.join("elsewhere")).unwrap();
    symlink(home.join("elsewhere"), older(&home)).unwrap();
    let p = plan(&home);
    assert_eq!(p["action"], json!("blocked"), "{p}");
    assert_eq!(blocks(&p), ["earlier_folder_is_a_link"]);
    refused(&apply(&home, &digest_of(&p)));
    assert!(exists(&home.join("elsewhere")) && !exists(&home.join(".moai/archive")));

    let dir = tempfile::tempdir().unwrap();
    let home = fs::canonicalize(dir.path()).unwrap();
    write(&home, ".moai-cli", "not a folder\n");
    let p = plan(&home);
    assert_eq!(blocks(&p), ["earlier_folder_not_a_folder"], "{p}");
}

#[test]
fn blocked_when_moai_or_its_archive_folder_is_not_a_folder_or_the_name_is_taken() {
    let (_d, home) = earlier_home();
    fs::remove_dir_all(moai(&home)).unwrap();
    fs::create_dir_all(home.join("elsewhere")).unwrap();
    symlink(home.join("elsewhere"), moai(&home)).unwrap();
    assert_eq!(blocks(&plan(&home)), ["moai_folder_not_a_folder"]);

    let (_d, home) = earlier_home();
    write(&home, ".moai/archive", "a plain file\n");
    assert_eq!(blocks(&plan(&home)), ["archive_not_a_folder"]);

    let (_d, home) = earlier_home();
    fs::create_dir_all(archive(&home)).unwrap();
    assert_eq!(blocks(&plan(&home)), ["stamp_taken"]);
}

#[test]
fn blocked_while_a_program_uses_the_earlier_folder_and_the_apply_refuses() {
    let (_d, home) = earlier_home();
    let before = tree(&home);
    let _cwd = in_use(home.join(".moai-cli/state"));
    let p = plan(&home);
    assert_eq!(p["action"], json!("blocked"), "{p}");
    assert_eq!(blocks(&p), ["in_use:process_cwd"]);
    refused(&apply(&home, &digest_of(&p)));
    assert_eq!(before, tree(&home), "a blocked apply writes nothing");
}

#[test]
fn blocked_while_a_program_holds_a_file_of_the_earlier_folder_open() {
    let (_d, home) = earlier_home();
    let before = tree(&home);
    let _open = holding_open(&home.join(".moai-cli/receipts/r.json"));
    let p = plan(&home);
    assert_eq!(p["action"], json!("blocked"), "{p}");
    assert_eq!(blocks(&p), ["in_use:file_open"]);
    refused(&apply(&home, &digest_of(&p)));
    assert_eq!(before, tree(&home), "a blocked apply writes nothing");
}

#[test]
fn blocked_at_501_entries_and_planned_at_500_entries() {
    // N and N+1 of the listing limit: 500 entries are all listed and move, 501 are not fully
    // listed and the preview refuses to plan a move it could not show.
    let (_d, home) = home_with_entries(500);
    let p = plan(&home);
    assert_eq!(p["action"], json!("retire"), "{p}");
    assert_eq!(p["total"], json!(500));
    assert_eq!(p["entries"].as_array().unwrap().len(), 500);
    assert_eq!(blocks(&p), Vec::<String>::new());
    let applied = apply(&home, &digest_of(&p));
    assert_eq!(applied["ok"], json!(true), "{applied}");
    assert!(!exists(&older(&home)));

    let (_d, home) = home_with_entries(501);
    let before = tree(&home);
    let p = plan(&home);
    assert_eq!(p["action"], json!("blocked"), "{p}");
    assert_eq!(p["total"], json!(501));
    assert_eq!(p["entries"].as_array().unwrap().len(), 500, "the first 500");
    assert_eq!(blocks(&p), ["not_fully_listed"]);
    refused(&apply(&home, &digest_of(&p)));
    assert_eq!(before, tree(&home), "a blocked apply writes nothing");
}

#[test]
fn blocked_when_a_manifest_of_this_clock_is_not_the_one_this_step_wrote() {
    for foreign in ["not json at all", "{\"kind\":\"another-step\"}"] {
        let (_d, home) = earlier_home();
        write(
            &home,
            &format!(".moai/archive/state-retire-manifest-{STAMP}.json"),
            foreign,
        );
        let before = tree(&home);
        let p = plan(&home);
        assert_eq!(blocks(&p), ["stamp_taken"], "{foreign}: {p}");
        refused(&apply(&home, &digest_of(&p)));
        assert_eq!(before, tree(&home), "{foreign}");
    }
}

#[test]
fn blocked_when_the_process_table_cannot_be_read() {
    let (_d, home) = earlier_home();
    let p = plan_with(
        &home,
        NOW,
        &[("MOAI_CLI_TEST_PROCESS_CHECK", "unavailable")],
    );
    assert_eq!(p["action"], json!("blocked"), "{p}");
    assert_eq!(blocks(&p), ["in_use:process_check_unavailable"]);
}

#[test]
fn blocked_when_the_running_binary_lies_inside_the_earlier_folder() {
    let (_d, home) = earlier_home();
    let inside = home.join(".moai-cli/versions/4.2647.1/bin/moai-cli");
    fs::remove_file(&inside).unwrap();
    if fs::hard_link(BIN, &inside).is_err() {
        fs::copy(BIN, &inside).unwrap();
    }
    chmod(&inside, 0o755);
    let _permit = spawn_cap::acquire();
    let out = Command::new(&inside)
        .args([
            "home-archive-plan",
            "--home",
            home.to_str().unwrap(),
            "--now",
            NOW,
        ])
        .env_remove("MOAI_HOME")
        .envs([crate::common::process_table()])
        .stdin(Stdio::null())
        .output()
        .unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["result"]["action"], json!("blocked"), "{v}");
    assert_eq!(blocks(&v["result"]), ["running_binary_inside"]);
}

// ---- PATH guidance ---------------------------------------------------------------------------

#[test]
fn path_guidance_lists_the_entries_that_name_the_earlier_folder_and_touches_no_shell_file() {
    let (_d, home) = earlier_home();
    let rc = fs::read(home.join(".zshrc")).unwrap();
    let named = format!("{}/.moai-cli/path", home.display());
    let path = format!("{named}:/usr/sbin:/usr/bin:/bin");
    let p = plan_with(&home, NOW, &[("PATH", &path)]);
    let entries: Vec<&str> = p["path_entries"]
        .as_array()
        .unwrap()
        .iter()
        .map(|e| e.as_str().unwrap())
        .collect();
    assert_eq!(entries, [named.as_str()], "{p}");
    let guidance = p["path_guidance"].as_str().unwrap();
    assert!(guidance.contains("moai update"), "{guidance}");
    let applied = apply_with(
        &home,
        NOW,
        "--plan-digest",
        &digest_of(&p),
        PERM,
        &[("PATH", &path)],
    );
    assert_eq!(applied["ok"], json!(true), "{applied}");
    assert_eq!(
        rc,
        fs::read(home.join(".zshrc")).unwrap(),
        "no shell file is edited"
    );
}

// ---- the apply -------------------------------------------------------------------------------

#[test]
fn the_apply_moves_the_folder_whole_and_keeps_bytes_modes_and_go_files() {
    let (_d, home) = earlier_home();
    let old_before = tree(&older(&home));
    let moai_before: Tree = tree(&moai(&home));
    let p = plan(&home);
    let v = apply(&home, &digest_of(&p));
    assert_eq!(v["ok"], json!(true), "{v}");
    assert_eq!(v["result"]["state"], json!("applied"));
    assert_eq!(v["result"]["moved"], json!(1));
    assert!(!exists(&older(&home)), "no earlier folder is left");
    assert_eq!(
        old_before,
        tree(&archive(&home)),
        "same bytes, kinds and modes"
    );
    let vault = archive(&home).join("vault");
    assert_eq!(
        fs::metadata(&vault).unwrap().permissions().mode() & 0o777,
        0o700
    );
    assert_eq!(
        fs::metadata(vault.join("_shared/glm.key"))
            .unwrap()
            .permissions()
            .mode()
            & 0o777,
        0o600
    );
    let moai_after: Tree = tree(&moai(&home))
        .into_iter()
        .filter(|(r, ..)| !r.starts_with("/archive"))
        .collect();
    assert_eq!(
        moai_before, moai_after,
        "every Go name of ~/.moai is as it was"
    );
    assert!(exists(Path::new(v["result"]["manifest"].as_str().unwrap())));
    assert!(exists(Path::new(v["result"]["journal"].as_str().unwrap())));
}

#[test]
fn the_apply_accepts_the_content_digest_at_another_clock() {
    let (_d, home) = earlier_home();
    let p = plan(&home);
    let content = p["content_digest"].as_str().unwrap().to_string();
    let v = apply_with(&home, LATER, "--content-digest", &content, PERM, &[]);
    assert_eq!(v["ok"], json!(true), "{v}");
    assert!(exists(
        &home.join(".moai/archive/moai-cli-20261123T130000Z")
    ));
}

#[test]
fn the_apply_refuses_a_stale_token_a_missing_reference_and_writes_nothing() {
    let (_d, home) = earlier_home();
    let p = plan(&home);
    write(
        &home,
        ".moai-cli/late-arrival.txt",
        "added after the preview\n",
    );
    let before = tree(&home);
    let stale = refused(&apply(&home, &digest_of(&p)));
    assert!(stale.contains("stale or altered"), "{stale}");
    let fresh = plan(&home);
    let none = refused(&apply_with(
        &home,
        NOW,
        "--plan-digest",
        &digest_of(&fresh),
        "",
        &[],
    ));
    assert!(none.contains("permission reference"), "{none}");
    assert_eq!(before, tree(&home));
}

// ---- the revert ------------------------------------------------------------------------------

#[test]
fn the_revert_restores_the_tree_and_a_second_revert_changes_nothing() {
    let (_d, home) = earlier_home();
    let before = tree(&older(&home));
    let p = plan(&home);
    assert_eq!(apply(&home, &digest_of(&p))["ok"], json!(true));
    let r = revert(&home);
    assert_eq!(r["ok"], json!(true), "{r}");
    assert_eq!(r["result"]["state"], json!("reverted"), "{r}");
    assert_eq!(before, tree(&older(&home)), "bytes and modes restored");
    assert!(exists(&home.join(".moai/archive")), "nothing is deleted");
    let again = revert(&home);
    assert_eq!(again["result"]["changed"], json!(false), "{again}");
}

#[test]
fn the_revert_refuses_when_the_earlier_folder_came_back_or_the_entry_was_replaced() {
    let (_d, home) = earlier_home();
    let p = plan(&home);
    assert_eq!(apply(&home, &digest_of(&p))["ok"], json!(true));
    write(&home, ".moai-cli/new/f.txt", "made after the apply\n");
    let r = revert(&home);
    assert_eq!(r["result"]["state"], json!("revert_partial"), "{r}");
    assert!(
        exists(&archive(&home).join("vault")),
        "the entry stays in the archive"
    );
    assert_eq!(
        fs::read(home.join(".moai-cli/new/f.txt")).unwrap(),
        b"made after the apply\n"
    );

    let (_d, home) = earlier_home();
    let p = plan(&home);
    assert_eq!(apply(&home, &digest_of(&p))["ok"], json!(true));
    fs::rename(archive(&home), home.join(".moai/archive/moved-away")).unwrap();
    write(
        &home,
        ".moai/archive/moai-cli-20261123T120000Z/other.txt",
        "another entry\n",
    );
    let r = revert(&home);
    assert_eq!(r["result"]["state"], json!("revert_partial"), "{r}");
    assert!(exists(&archive(&home).join("other.txt")));
    assert!(!exists(&older(&home)));
}

#[test]
fn the_revert_refuses_an_empty_or_placeholder_reference_and_leaves_the_archive_as_it_is() {
    let (_d, home) = earlier_home();
    let before = tree(&older(&home));
    let p = plan(&home);
    assert_eq!(apply(&home, &digest_of(&p))["ok"], json!(true));
    let archived = tree(&home);
    for bad in ["", "   ", "null", "placeholder", " placeholder "] {
        let v = revert_with(&home, bad);
        let why = refused(&v);
        assert!(why.contains("permission reference"), "`{bad}`: {why}");
        assert_eq!(
            archived,
            tree(&home),
            "`{bad}`: the archive stays where it is"
        );
        assert!(!exists(&older(&home)), "`{bad}`");
    }
    let r = revert(&home);
    assert_eq!(r["ok"], json!(true), "{r}");
    assert_eq!(before, tree(&older(&home)));
}

// ---- crashes ---------------------------------------------------------------------------------

#[test]
fn a_crash_between_the_manifest_and_the_journal_is_finished_by_a_rerun_of_the_same_clock() {
    let (_d, home) = earlier_home();
    let before = tree(&older(&home));
    let p = plan(&home);
    let digest = digest_of(&p);
    let crashed = apply_with(
        &home,
        NOW,
        "--plan-digest",
        &digest,
        PERM,
        &[(FAULT, "after_manifest")],
    );
    assert!(
        refused(&crashed).contains("injected crash"),
        "after_manifest: {crashed}"
    );
    assert!(exists(&manifest(&home)), "the manifest was written");
    assert!(!exists(&journal(&home)), "the journal was not");
    assert_eq!(before, tree(&older(&home)), "nothing moved");

    // The manifest left behind is this step's own: the same clock is not `stamp_taken`.
    let again = plan(&home);
    assert_eq!(again["action"], json!("retire"), "{again}");
    assert_eq!(
        digest_of(&again),
        digest,
        "the same preview as before the crash"
    );

    let done = apply(&home, &digest);
    assert_eq!(done["ok"], json!(true), "{done}");
    assert_eq!(done["result"]["state"], json!("applied"), "{done}");
    assert!(!exists(&older(&home)));
    assert_eq!(before, tree(&archive(&home)));
    let m: Value = serde_json::from_slice(&fs::read(manifest(&home)).unwrap()).unwrap();
    assert_eq!(
        m["plan_digest"],
        json!(digest),
        "the manifest names the finished plan"
    );
    assert!(exists(&journal(&home)));

    let r = revert(&home);
    assert_eq!(r["ok"], json!(true), "{r}");
    assert_eq!(before, tree(&older(&home)));
}

fn temporary_manifest(home: &Path) -> PathBuf {
    moai(home).join(format!("archive/state-retire-manifest-{STAMP}.json.tmp"))
}

#[test]
fn a_link_planted_at_the_temporary_manifest_name_is_replaced_and_never_followed() {
    let (_d, home) = earlier_home();
    let before = tree(&older(&home));
    let digest = digest_of(&plan(&home));
    let crashed = apply_with(
        &home,
        NOW,
        "--plan-digest",
        &digest,
        PERM,
        &[(FAULT, "after_manifest")],
    );
    assert!(refused(&crashed).contains("injected crash"), "{crashed}");

    // A file outside HOME, and a link to it at the name the rerun writes its manifest under.
    let outside = tempfile::tempdir().unwrap();
    let victim = outside.path().join("victim.txt");
    fs::write(&victim, "precious outside file\n").unwrap();
    chmod(&victim, 0o600);
    symlink(&victim, temporary_manifest(&home)).unwrap();

    let done = apply(&home, &digest);
    assert_eq!(done["ok"], json!(true), "{done}");
    assert_eq!(done["result"]["state"], json!("applied"), "{done}");
    assert_eq!(
        fs::read(&victim).unwrap(),
        b"precious outside file\n",
        "the file the link pointed at is untouched"
    );
    assert_eq!(
        fs::metadata(&victim).unwrap().permissions().mode() & 0o777,
        0o600
    );
    let name = fs::symlink_metadata(manifest(&home)).unwrap();
    assert!(
        name.is_file() && !name.file_type().is_symlink(),
        "the manifest is a plain file of its own"
    );
    let m: Value = serde_json::from_slice(&fs::read(manifest(&home)).unwrap()).unwrap();
    assert_eq!(m["plan_digest"], json!(digest));
    assert!(
        !exists(&temporary_manifest(&home)),
        "the link is gone, not left"
    );
    assert_eq!(before, tree(&archive(&home)));
}

#[test]
fn the_first_manifest_is_written_whole_under_a_temporary_name_and_published_by_one_step() {
    let (_d, home) = earlier_home();
    let before = tree(&older(&home));
    let digest = digest_of(&plan(&home));
    let crashed = apply_with(
        &home,
        NOW,
        "--plan-digest",
        &digest,
        PERM,
        &[(FAULT, "after_manifest_tmp")],
    );
    assert!(refused(&crashed).contains("injected crash"), "{crashed}");
    // A crash here leaves no file at the manifest name, so no empty or half-written manifest.
    assert!(!exists(&manifest(&home)), "no file at the manifest name");
    let tmp: Value = serde_json::from_slice(&fs::read(temporary_manifest(&home)).unwrap())
        .expect("the temporary file holds the whole manifest");
    assert_eq!(tmp["plan_digest"], json!(digest));
    assert!(!exists(&journal(&home)));
    assert_eq!(before, tree(&older(&home)), "nothing moved");

    // The leftover name does not block the clock, and a rerun replaces whatever sits there.
    fs::write(temporary_manifest(&home), "junk of an earlier crash").unwrap();
    let again = plan(&home);
    assert_eq!(again["action"], json!("retire"), "{again}");
    assert_eq!(digest_of(&again), digest);
    let done = apply(&home, &digest);
    assert_eq!(done["ok"], json!(true), "{done}");
    assert!(!exists(&temporary_manifest(&home)));
    let m: Value = serde_json::from_slice(&fs::read(manifest(&home)).unwrap()).unwrap();
    assert_eq!(m["plan_digest"], json!(digest));
    assert_eq!(before, tree(&archive(&home)));
}

#[test]
fn a_crash_at_any_journal_step_is_finished_by_a_rerun_or_undone_by_the_revert() {
    for point in [
        "after_journal",
        "after_mark_moving",
        "after_rename",
        "before_commit",
    ] {
        let (_d, home) = earlier_home();
        let before = tree(&older(&home));
        let p = plan(&home);
        let digest = digest_of(&p);
        let crashed = apply_with(
            &home,
            NOW,
            "--plan-digest",
            &digest,
            PERM,
            &[(FAULT, point)],
        );
        assert!(
            refused(&crashed).contains("injected crash"),
            "{point}: {crashed}"
        );
        if point == "after_journal" {
            // Nothing moved yet: the rerun with the same clock and token finishes the move.
            let again = apply(&home, &digest);
            assert_eq!(again["ok"], json!(true), "{point}: {again}");
            assert!(!exists(&older(&home)), "{point}");
            assert_eq!(before, tree(&archive(&home)), "{point}");
        } else {
            let r = revert(&home);
            assert_eq!(r["ok"], json!(true), "{point}: {r}");
            assert_eq!(before, tree(&older(&home)), "{point}: undone");
            assert!(
                !exists(&archive(&home)),
                "{point}: nothing left in the archive"
            );
        }
    }
}
