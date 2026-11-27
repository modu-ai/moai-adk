//! #2483: the confirmed backup-then-replace procedure for an earlier-generation (Go-era)
//! `moai` skill folder (decision row 22). New-file discipline: the helpers are a minimal
//! copy of update_unmanaged_failclosed.rs's.
//!
//! Every fixture runs in a temporary HOME with a synthetic Go-type folder — the real
//! `~/.claude`, `~/.agents` and `~/.moai-cli` are never read or written.
//!
//! AC-1 preview without writes; AC-2 no token → nothing changes, a matching token → the
//! backup MANIFEST and the moved tree equal the original bytes and modes, the projection
//! equals the package; AC-3 recovery restores the original bytes and modes and records
//! the restore; AC-4 injected failures (after the backup, mid replace) keep the original
//! or the backup intact and surface as pending; AC-6 the Claude host location, with the shared
//! agents folder left alone.

#![cfg(unix)]

mod common;
use common::CleanEnv;
use serde_json::{json, Value};
use std::{
    collections::BTreeMap,
    fs,
    io::Write as _,
    os::unix::fs::PermissionsExt,
    path::{Path, PathBuf},
    process::{Command, Stdio},
    thread::sleep,
    time::Duration,
};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;
/// The HOME state folder of the new name, where every backup is made.
const NEWER: &str = ".moai";
/// The state folder the fixture HOME starts with, so its receipts and version folders sit there.
const EARLIER_STATE: &str = ".moai-cli";
const ROOTS: [(&str, &str); 1] = [("claude", ".claude/skills/moai")];

fn write_pkg(dir: &Path) {
    let files = [
        (
            "SKILL.md",
            "---\nname: moai\ndescription: fixture package\n---\n# MoAI\nrust\n",
        ),
        ("references/routes.json", "{\"routes\":1}\n"),
        ("workflows/update.md", "# UPDATE\nrust\n"),
    ];
    for (rel, body) in files {
        let p = dir.join(rel);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        fs::write(p, body).unwrap();
    }
}

/// A synthetic Go-type skill: other files than the package, an executable and a
/// private file, a nested directory.
fn write_go_skill(root: &Path, tag: &str) {
    let files: [(&str, String, u32); 4] = [
        (
            "SKILL.md",
            format!("---\nname: moai\n---\n# Go-era MoAI {tag}\n"),
            0o644,
        ),
        ("modules/run.md", format!("go run module {tag}\n"), 0o644),
        ("bin/hook.sh", "#!/bin/sh\nexit 0\n".into(), 0o755),
        ("notes/private.txt", format!("user note {tag}\n"), 0o600),
    ];
    for (rel, body, mode) in files {
        let p = root.join(rel);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        fs::write(&p, body).unwrap();
        fs::set_permissions(&p, fs::Permissions::from_mode(mode)).unwrap();
    }
}

struct Fixture {
    _tmp: TempDir,
    home: PathBuf,
    stage: PathBuf,
}

fn fixture() -> Fixture {
    let tmp = TempDir::new().unwrap();
    let home = tmp.path().join("home");
    let stage = tmp.path().join("stage");
    write_pkg(&stage);
    for (host, rel) in ROOTS {
        write_go_skill(&home.join(rel), host);
    }
    // An earlier-root HOME: its state stays in `~/.moai-cli` (#2696 C4b2 puts a HOME with no
    // state on the new state folder; that layout is pinned by home_root tests and update_cli).
    fs::create_dir_all(home.join(EARLIER_STATE)).unwrap();
    let home = fs::canonicalize(&home).unwrap();
    Fixture {
        _tmp: tmp,
        home,
        stage,
    }
}

/// rel path -> (bytes, mode & 0o7777) for every regular file under `root`.
fn snapshot(root: &Path) -> BTreeMap<String, (Vec<u8>, u32)> {
    fn walk(root: &Path, dir: &Path, out: &mut BTreeMap<String, (Vec<u8>, u32)>) {
        for e in fs::read_dir(dir).unwrap() {
            let p = e.unwrap().path();
            let m = fs::symlink_metadata(&p).unwrap();
            if m.is_dir() {
                walk(root, &p, out);
            } else {
                let rel = p.strip_prefix(root).unwrap().to_str().unwrap().to_string();
                out.insert(
                    rel,
                    (fs::read(&p).unwrap(), m.permissions().mode() & 0o7777),
                );
            }
        }
    }
    let mut out = BTreeMap::new();
    if root.exists() {
        walk(root, root, &mut out);
    }
    out
}

fn run_cli(
    args: &[&str],
    stdin: Option<&Value>,
    home: &Path,
    fault: Option<&str>,
) -> std::process::Output {
    let mut cmd = Command::new(EXE);
    cmd.args(args)
        .env("HOME", home)
        .env_remove("MOAI_TEST_FAULT")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped());
    if let Some(f) = fault {
        cmd.env("MOAI_TEST_FAULT", f);
    }
    let mut child = cmd.spawn().unwrap();
    let mut pipe = child.stdin.take().unwrap();
    if let Some(v) = stdin {
        let _ = pipe.write_all(v.to_string().as_bytes());
    }
    drop(pipe);
    child.wait_with_output().unwrap()
}

fn envelope(out: &std::process::Output) -> Value {
    serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
        panic!(
            "envelope JSON on stdout; stderr: {}",
            String::from_utf8_lossy(&out.stderr)
        )
    })
}

fn result_of(out: &std::process::Output) -> Value {
    let v = envelope(out);
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn plan(f: &Fixture) -> Value {
    let out = run_cli(
        &[
            "update-plan",
            f.home.to_str().unwrap(),
            f.stage.to_str().unwrap(),
        ],
        None,
        &f.home,
        None,
    );
    result_of(&out)
}

fn apply(f: &Fixture, plan: &Value, tokens: &[&str], fault: Option<&str>) -> std::process::Output {
    let digest = plan["plan_digest"].as_str().unwrap().to_string();
    let mut args = vec![
        "update-apply".to_string(),
        f.home.to_str().unwrap().to_string(),
        f.stage.to_str().unwrap().to_string(),
        "--plan-digest".into(),
        digest,
        "--permission-reference".into(),
        "fixture-approved".into(),
    ];
    for t in tokens {
        args.push("--confirm-backup-replace".into());
        args.push((*t).to_string());
    }
    let refs: Vec<&str> = args.iter().map(String::as_str).collect();
    run_cli(&refs, Some(plan), &f.home, fault)
}

fn recovery_plan(f: &Fixture) -> Value {
    result_of(&run_cli(
        &["recovery-plan", f.home.to_str().unwrap()],
        None,
        &f.home,
        None,
    ))
}

fn recovery_apply(f: &Fixture, plan: &Value) -> std::process::Output {
    let digest = plan["plan_digest"].as_str().unwrap();
    run_cli(
        &[
            "recovery-apply",
            f.home.to_str().unwrap(),
            "--plan-digest",
            digest,
            "--permission-reference",
            "fixture-approved",
        ],
        Some(plan),
        &f.home,
        None,
    )
}

fn host<'a>(v: &'a Value, name: &str) -> &'a Value {
    v["hosts"]
        .as_array()
        .unwrap()
        .iter()
        .find(|h| h["host"] == json!(name))
        .expect("host row")
}

fn token(plan: &Value, name: &str) -> String {
    host(plan, name)["backup_then_replace"]["confirm_token"]
        .as_str()
        .unwrap()
        .to_string()
}

fn backups(home: &Path) -> Vec<PathBuf> {
    let dir = home.join(NEWER).join("backups");
    if !dir.exists() {
        return Vec::new();
    }
    let mut v: Vec<PathBuf> = fs::read_dir(dir)
        .unwrap()
        .map(|e| e.unwrap().path())
        .collect();
    v.sort();
    v
}

fn backup_of(home: &Path, host: &str) -> PathBuf {
    let found: Vec<PathBuf> = backups(home)
        .into_iter()
        .filter(|p| {
            p.file_name()
                .unwrap()
                .to_str()
                .unwrap()
                .ends_with(&format!("-{host}"))
        })
        .collect();
    assert_eq!(found.len(), 1, "one backup folder for {host}: {found:?}");
    found[0].clone()
}

fn originals(f: &Fixture) -> Vec<BTreeMap<String, (Vec<u8>, u32)>> {
    ROOTS
        .iter()
        .map(|(_, rel)| snapshot(&f.home.join(rel)))
        .collect()
}

fn assert_package_installed(root: &Path, stage: &Path) {
    let installed: BTreeMap<String, Vec<u8>> = snapshot(root)
        .into_iter()
        .map(|(k, (b, _))| (k, b))
        .collect();
    let package: BTreeMap<String, Vec<u8>> = snapshot(stage)
        .into_iter()
        .map(|(k, (b, _))| (k, b))
        .collect();
    assert_eq!(
        installed, package,
        "the projection holds exactly the package"
    );
}

/// AC-1 (+AC-6): the plan previews backup-then-replace for the Claude host — what, where,
/// how many files, which hashes — and writes nothing.
#[test]
fn plan_previews_backup_then_replace_and_writes_nothing() {
    let f = fixture();
    let before = originals(&f);
    let p = plan(&f);
    for ((name, rel), orig) in ROOTS.iter().zip(&before) {
        let h = host(&p, name);
        assert_eq!(
            h["preflight"],
            json!("blocked"),
            "still blocked without confirmation"
        );
        let b = &h["backup_then_replace"];
        assert_eq!(b["source_root"], json!(f.home.join(rel).to_str().unwrap()));
        assert_eq!(b["source_file_count"], json!(orig.len()));
        let files = b["source_files"].as_array().unwrap();
        assert_eq!(files.len(), orig.len());
        for e in files {
            let (bytes, mode) = &orig[e["path"].as_str().unwrap()];
            let sha = format!("sha256:{}", sha256_hex(bytes));
            assert_eq!(e["sha256"], json!(sha), "{e}");
            assert_eq!(e["mode"], json!(mode), "{e}");
        }
        assert!(b["source_tree_sha256"]
            .as_str()
            .unwrap()
            .starts_with("sha256:"));
        assert_eq!(
            b["backup_root"],
            json!(f.home.join(NEWER).join("backups").to_str().unwrap())
        );
        assert_eq!(b["replace_with"]["file_count"], json!(3));
        assert_eq!(b["replace_with"]["package_digest"], p["package_digest"]);
        let tok = b["confirm_token"].as_str().unwrap();
        assert!(tok.starts_with(&format!("backup-replace:{name}:")), "{tok}");
    }
    assert_eq!(originals(&f), before, "plan wrote nothing");
    assert!(
        !f.home.join(NEWER).join("backups").exists(),
        "no backup folder from a plan"
    );
}

/// A managed or absent projection gets no preview (the field is left out).
#[test]
fn no_preview_without_an_unmanaged_package() {
    let tmp = TempDir::new().unwrap();
    let home = tmp.path().join("home");
    fs::create_dir_all(&home).unwrap();
    let stage = tmp.path().join("stage");
    write_pkg(&stage);
    let f = Fixture {
        home: fs::canonicalize(&home).unwrap(),
        stage,
        _tmp: tmp,
    };
    let p = plan(&f);
    for (name, _) in ROOTS {
        assert_eq!(host(&p, name)["preflight"], json!("ok"));
        assert!(host(&p, name).get("backup_then_replace").is_none(), "{p}");
    }
}

/// AC-2 (first half): without a confirmation token the apply stays blocked and nothing
/// changes; an unknown token is refused before any write.
#[test]
fn apply_without_token_or_with_a_wrong_token_changes_nothing() {
    let f = fixture();
    let before = originals(&f);
    let p = plan(&f);
    let out = apply(&f, &p, &[], None);
    let r = result_of(&out);
    assert_eq!(r["status"], json!("blocked"), "{r}");
    assert_eq!(originals(&f), before);
    assert!(backups(&f.home).is_empty());

    let out = apply(&f, &p, &["backup-replace:claude:0000000000000000"], None);
    let v = envelope(&out);
    assert_eq!(
        v["ok"],
        json!(false),
        "a token the plan did not offer is refused: {v}"
    );
    assert_eq!(originals(&f), before);
    assert!(backups(&f.home).is_empty());
}

/// AC-2 (second half) + AC-6: with the token, the host's original folder moves into
/// a backup whose MANIFEST and tree equal the original bytes and modes, and the
/// projection holds exactly the package.
#[test]
fn confirmed_apply_backs_up_and_replaces_the_host() {
    let f = fixture();
    let before = originals(&f);
    let p = plan(&f);
    let toks = [token(&p, "claude")];
    let refs: Vec<&str> = toks.iter().map(String::as_str).collect();
    let r = result_of(&apply(&f, &p, &refs, None));
    assert_eq!(r["status"], json!("applied"), "{r}");
    for ((name, rel), orig) in ROOTS.iter().zip(&before) {
        let h = host(&r, name);
        assert_eq!(h["status"], json!("replaced"), "{h}");
        assert_eq!(h["receipt_status"], json!("committed"));
        let dir = backup_of(&f.home, name);
        assert_eq!(h["backup_dir"], json!(dir.to_str().unwrap()));
        assert_eq!(
            &snapshot(&dir.join("moai")),
            orig,
            "{name}: backup tree = original"
        );
        let manifest: Value =
            serde_json::from_slice(&fs::read(dir.join("MANIFEST.json")).unwrap()).unwrap();
        let listed: BTreeMap<String, (String, u32)> = manifest["files"]
            .as_array()
            .unwrap()
            .iter()
            .map(|e| {
                (
                    e["path"].as_str().unwrap().to_string(),
                    (
                        e["sha256"].as_str().unwrap().to_string(),
                        e["mode"].as_u64().unwrap() as u32,
                    ),
                )
            })
            .collect();
        let expected: BTreeMap<String, (String, u32)> = orig
            .iter()
            .map(|(k, (b, m))| (k.clone(), (format!("sha256:{}", sha256_hex(b)), *m)))
            .collect();
        assert_eq!(
            listed, expected,
            "{name}: MANIFEST lists the original bytes and modes"
        );
        assert_package_installed(&f.home.join(rel), &f.stage);
    }
    // A rerun is an ordinary managed update now: ok and unchanged.
    let again = plan(&f);
    for (name, _) in ROOTS {
        assert_eq!(host(&again, name)["preflight"], json!("ok"), "{again}");
    }
}

/// A Go-type folder in the shared agents folder is not planned, backed up or replaced: moai no
/// longer manages it, so the plan has no row for it and the apply leaves it byte for byte.
#[test]
fn the_shared_agents_folder_is_not_planned_or_replaced() {
    let f = fixture();
    let shared = f.home.join(".agents/skills/moai");
    write_go_skill(&shared, "shared");
    let shared_before = snapshot(&shared);
    let p = plan(&f);
    let hosts: Vec<&str> = p["hosts"]
        .as_array()
        .unwrap()
        .iter()
        .map(|h| h["host"].as_str().unwrap())
        .collect();
    assert_eq!(hosts, ["claude"], "{p}");
    let r = result_of(&apply(&f, &p, &[&token(&p, "claude")], None));
    assert_eq!(host(&r, "claude")["status"], json!("replaced"), "{r}");
    assert_eq!(snapshot(&shared), shared_before, "shared folder untouched");
    assert_eq!(backups(&f.home).len(), 1);
}

/// AC-3: recovery restores the original folder byte and mode identical and records the
/// use of the backup; the next plan offers the preview again.
#[test]
fn recovery_restores_the_original_folder() {
    let f = fixture();
    let before = originals(&f);
    let p = plan(&f);
    let toks = [token(&p, "claude")];
    let refs: Vec<&str> = toks.iter().map(String::as_str).collect();
    result_of(&apply(&f, &p, &refs, None));

    let rp = recovery_plan(&f);
    for (name, _) in ROOTS {
        let rb = &host(&rp, name)["restore_backup"];
        assert_eq!(rb["original_at"], json!("backup"), "{rp}");
        assert_eq!(rb["action"], json!("restore_backup"), "{rp}");
    }
    let r = result_of(&recovery_apply(&f, &rp));
    assert_eq!(r["status"], json!("success"), "{r}");
    for ((name, rel), orig) in ROOTS.iter().zip(&before) {
        assert_eq!(
            &snapshot(&f.home.join(rel)),
            orig,
            "{name}: original restored"
        );
        let dir = backup_of(&f.home, name);
        let restored: Value =
            serde_json::from_slice(&fs::read(dir.join("RESTORED.json")).unwrap()).unwrap();
        assert_eq!(restored["host"], json!(name));
        assert_eq!(restored["plan_digest"], rp["plan_digest"]);
        // The replaced package was moved aside, never deleted.
        assert_package_installed(&dir.join("replaced"), &f.stage);
        assert_eq!(host(&r, name)["receipt_status"], json!("recovered"));
    }
    let again = plan(&f);
    for (name, _) in ROOTS {
        assert_eq!(host(&again, name)["preflight"], json!("blocked"));
        assert!(
            host(&again, name)["backup_then_replace"].is_object(),
            "{again}"
        );
    }
}

/// AC-4: a failure after the backup but before the replacement leaves the original in
/// place; the host is pending and recovery releases it without moving anything.
#[test]
fn a_failure_after_the_backup_keeps_the_original_and_recovers() {
    let f = fixture();
    let before = originals(&f);
    let p = plan(&f);
    let out = apply(&f, &p, &[&token(&p, "claude")], Some("after_backup"));
    let r = result_of(&out);
    assert_eq!(r["status"], json!("partial"), "{r}");
    assert_eq!(host(&r, "claude")["receipt_status"], json!("pending"));
    assert_eq!(
        snapshot(&f.home.join(ROOTS[0].1)),
        before[0],
        "original intact"
    );
    assert_eq!(
        host(&plan(&f), "claude")["blocked_reason"],
        json!("host receipt is pending; run recovery-plan first")
    );

    let rp = recovery_plan(&f);
    let rb = &host(&rp, "claude")["restore_backup"];
    assert_eq!(rb["original_at"], json!("root"), "{rp}");
    assert_eq!(rb["action"], json!("release"), "{rp}");
    result_of(&recovery_apply(&f, &rp));
    assert_eq!(snapshot(&f.home.join(ROOTS[0].1)), before[0]);
    assert_eq!(host(&plan(&f), "claude")["preflight"], json!("blocked"));
    assert!(host(&plan(&f), "claude")["backup_then_replace"].is_object());
}

/// AC-4: a failure part way through the replacement leaves the original intact in the
/// backup; recovery puts it back byte and mode identical.
#[test]
fn a_failure_mid_replace_keeps_the_backup_and_recovers() {
    let f = fixture();
    let before = originals(&f);
    let p = plan(&f);
    let r = result_of(&apply(&f, &p, &[&token(&p, "claude")], Some("mid_replace")));
    assert_eq!(r["status"], json!("partial"), "{r}");
    assert_eq!(host(&r, "claude")["receipt_status"], json!("pending"));
    let dir = backup_of(&f.home, "claude");
    assert_eq!(snapshot(&dir.join("moai")), before[0], "backup intact");
    assert_ne!(
        snapshot(&f.home.join(ROOTS[0].1)),
        before[0],
        "replacement started"
    );

    let rp = recovery_plan(&f);
    let rb = &host(&rp, "claude")["restore_backup"];
    assert_eq!(rb["original_at"], json!("backup"), "{rp}");
    assert_eq!(rb["action"], json!("restore_backup"), "{rp}");
    result_of(&recovery_apply(&f, &rp));
    assert_eq!(
        snapshot(&f.home.join(ROOTS[0].1)),
        before[0],
        "original restored"
    );
}

/// AC-2/AC-5: when the moved folder no longer matches the manifest (a writer changed it
/// between the move and the check — injected), the apply moves it back, installs
/// nothing and leaves the host pending; recovery releases it.
#[test]
fn a_moved_folder_that_differs_from_the_manifest_is_moved_back() {
    let f = fixture();
    let before = originals(&f);
    let p = plan(&f);
    let r = result_of(&apply(
        &f,
        &p,
        &[&token(&p, "claude")],
        Some("corrupt_after_move"),
    ));
    let h = host(&r, "claude");
    assert_eq!(h["status"], json!("failed"), "{r}");
    assert_eq!(h["receipt_status"], json!("pending"));
    assert!(
        h["reload_basis"]
            .as_str()
            .unwrap()
            .contains("does not match the manifest"),
        "{h}"
    );
    let root = f.home.join(ROOTS[0].1);
    let now = snapshot(&root);
    assert_eq!(
        now.keys().collect::<Vec<_>>(),
        before[0].keys().collect::<Vec<_>>(),
        "moved back; no package file installed"
    );
    assert!(!backup_of(&f.home, "claude").join("moai").exists());
}

/// A changed backup is never put back: recovery reports a conflict and writes nothing.
#[test]
fn a_tampered_backup_is_not_restored() {
    let f = fixture();
    let p = plan(&f);
    result_of(&apply(&f, &p, &[&token(&p, "claude")], None));
    let dir = backup_of(&f.home, "claude");
    fs::write(dir.join("moai/modules/run.md"), "tampered\n").unwrap();
    let installed = snapshot(&f.home.join(ROOTS[0].1));

    let rp = recovery_plan(&f);
    let rb = &host(&rp, "claude")["restore_backup"];
    assert_eq!(rb["action"], json!("none"), "{rp}");
    assert_eq!(rb["reason"], json!("backup_mismatch"), "{rp}");
    let out = recovery_apply(&f, &rp);
    let _ = envelope(&out);
    assert_eq!(
        snapshot(&f.home.join(ROOTS[0].1)),
        installed,
        "projection untouched"
    );
    assert!(!dir.join("RESTORED.json").exists());
}

/// Every entry (recursive) under `dir`, for "nothing was written there" checks.
fn entries(dir: &Path) -> Vec<PathBuf> {
    let mut out = Vec::new();
    if let Ok(rd) = fs::read_dir(dir) {
        for e in rd {
            let p = e.unwrap().path();
            out.push(p.clone());
            if fs::symlink_metadata(&p).unwrap().is_dir() {
                out.extend(entries(&p));
            }
        }
    }
    out
}

/// Confirmed apply when `link` (a state path) is a symlink to a directory outside HOME:
/// the host fails before anything moves, the original folder keeps its bytes and modes,
/// and the link target receives no write.
fn assert_symlinked_state_path_is_refused(link_rel: &str) {
    let f = fixture();
    let outside = f._tmp.path().join("outside");
    fs::create_dir_all(&outside).unwrap();
    let link = f.home.join(link_rel);
    fs::create_dir_all(link.parent().unwrap()).unwrap();
    // The fixture's empty earlier root gives way to the link under test.
    let _ = fs::remove_dir(&link);
    std::os::unix::fs::symlink(&outside, &link).unwrap();
    let before = originals(&f);
    let p = plan(&f);
    let r = result_of(&apply(&f, &p, &[&token(&p, "claude")], None));
    let h = host(&r, "claude");
    assert_eq!(h["status"], json!("failed"), "{link_rel}: {r}");
    assert_eq!(
        h["receipt_status"],
        Value::Null,
        "{link_rel}: no pending receipt: {h}"
    );
    assert!(
        h["reload_basis"]
            .as_str()
            .unwrap()
            .contains("is a symlink; refused"),
        "{link_rel}: {h}"
    );
    assert_eq!(
        originals(&f),
        before,
        "{link_rel}: original folders unchanged"
    );
    assert_eq!(
        entries(&outside),
        Vec::<PathBuf>::new(),
        "{link_rel}: no write through the link"
    );
}

/// B1 (audit 6812528075): the backups folder (now under the new state folder, revision 3 of the update
/// baseline SPEC) is a symlink leading out of HOME.
#[test]
fn a_symlinked_backups_directory_is_refused() {
    assert_symlinked_state_path_is_refused(&format!("{NEWER}/backups"));
}

/// B1 (audit 6812528075): `~/.moai-cli` itself is a symlink leading out of HOME.
#[test]
fn a_symlinked_state_directory_is_refused() {
    assert_symlinked_state_path_is_refused(".moai-cli");
}

/// A user edit to the installed package blocks the restore (never overwritten).
#[test]
fn a_user_edit_after_the_replace_blocks_the_restore() {
    let f = fixture();
    let p = plan(&f);
    result_of(&apply(&f, &p, &[&token(&p, "claude")], None));
    let root = f.home.join(ROOTS[0].1);
    fs::write(root.join("workflows/update.md"), "user edit\n").unwrap();
    let rp = recovery_plan(&f);
    let rb = &host(&rp, "claude")["restore_backup"];
    assert_eq!(rb["action"], json!("none"), "{rp}");
    assert_eq!(rb["reason"], json!("user_changed"), "{rp}");
}

fn sha256_hex(bytes: &[u8]) -> String {
    use sha2::{Digest, Sha256};
    format!("{:x}", Sha256::digest(bytes))
}

// ---- #2505: same-second reruns, the replace window, the restore re-check ----

/// Test seam (#2505): debug builds take the backup folder's time stamp from this variable,
/// so two runs "in the same second" are reproducible.
const STAMP_VAR: &str = "MOAI_TEST_BACKUP_STAMP";
const STAMP: &str = "20260101T000000Z";

fn run_cli_env(
    args: &[&str],
    stdin: Option<&Value>,
    home: &Path,
    envs: &[(&str, &str)],
) -> std::process::Output {
    let mut cmd = Command::new(EXE);
    cmd.args(args)
        .env("HOME", home)
        .env_remove("MOAI_TEST_FAULT")
        .env_remove(STAMP_VAR)
        .envs(envs.iter().copied())
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped());
    let mut child = cmd.spawn().unwrap();
    let mut pipe = child.stdin.take().unwrap();
    if let Some(v) = stdin {
        let _ = pipe.write_all(v.to_string().as_bytes());
    }
    drop(pipe);
    child.wait_with_output().unwrap()
}

fn apply_env(
    f: &Fixture,
    plan: &Value,
    tokens: &[&str],
    envs: &[(&str, &str)],
) -> std::process::Output {
    let digest = plan["plan_digest"].as_str().unwrap().to_string();
    let mut args = vec![
        "update-apply".to_string(),
        f.home.to_str().unwrap().to_string(),
        f.stage.to_str().unwrap().to_string(),
        "--plan-digest".into(),
        digest,
        "--permission-reference".into(),
        "fixture-approved".into(),
    ];
    for t in tokens {
        args.push("--confirm-backup-replace".into());
        args.push((*t).to_string());
    }
    let refs: Vec<&str> = args.iter().map(String::as_str).collect();
    run_cli_env(&refs, Some(plan), &f.home, envs)
}

fn backups_dir(f: &Fixture) -> PathBuf {
    f.home.join(NEWER).join("backups")
}

/// AC-1: replace, restore, replace again within the same (pinned) second. The second run
/// gets `<stamp>.2-<host>` instead of failing, and the first backup is left as it was.
#[test]
fn a_rerun_in_the_same_second_gets_its_own_backup_folder() {
    let f = fixture();
    let p = plan(&f);
    let r = result_of(&apply_env(
        &f,
        &p,
        &[&token(&p, "claude")],
        &[(STAMP_VAR, STAMP)],
    ));
    let first = backups_dir(&f).join(format!("{STAMP}-claude"));
    assert_eq!(
        host(&r, "claude")["backup_dir"],
        json!(first.to_str().unwrap()),
        "{r}"
    );
    result_of(&recovery_apply(&f, &recovery_plan(&f)));
    let kept = snapshot(&first);

    let p2 = plan(&f);
    let r2 = result_of(&apply_env(
        &f,
        &p2,
        &[&token(&p2, "claude")],
        &[(STAMP_VAR, STAMP)],
    ));
    let h = host(&r2, "claude");
    assert_eq!(h["status"], json!("replaced"), "{r2}");
    let second = backups_dir(&f).join(format!("{STAMP}.2-claude"));
    assert_eq!(h["backup_dir"], json!(second.to_str().unwrap()), "{h}");
    assert_eq!(snapshot(&first), kept, "the earlier backup is untouched");
    assert_package_installed(&f.home.join(ROOTS[0].1), &f.stage);
}

/// AC-1 boundary: 16 names are tried (`<stamp>-<host>`, then `.2` … `.16`). With 15 taken
/// the run uses `.16`; with all 16 taken it fails before anything moves, and the message
/// names the cause.
#[test]
fn the_sixteenth_backup_name_is_the_last_one_tried() {
    for taken in [15usize, 16] {
        let f = fixture();
        let before = originals(&f);
        fs::create_dir_all(backups_dir(&f)).unwrap();
        fs::create_dir(backups_dir(&f).join(format!("{STAMP}-claude"))).unwrap();
        for n in 2..=taken {
            fs::create_dir(backups_dir(&f).join(format!("{STAMP}.{n}-claude"))).unwrap();
        }
        let p = plan(&f);
        let r = result_of(&apply_env(
            &f,
            &p,
            &[&token(&p, "claude")],
            &[(STAMP_VAR, STAMP)],
        ));
        let h = host(&r, "claude");
        if taken == 15 {
            assert_eq!(h["status"], json!("replaced"), "{taken}: {h}");
            let last = backups_dir(&f).join(format!("{STAMP}.16-claude"));
            assert_eq!(h["backup_dir"], json!(last.to_str().unwrap()), "{h}");
        } else {
            assert_eq!(h["status"], json!("failed"), "{taken}: {h}");
            assert_eq!(h["receipt_status"], Value::Null, "{h}");
            let why = h["reload_basis"].as_str().unwrap();
            assert!(
                why.contains("16 backup folder names") && why.contains("already exist"),
                "{why}"
            );
            assert!(why.contains(STAMP), "{why}");
            assert_eq!(
                snapshot(&f.home.join(ROOTS[0].1)),
                before[0],
                "nothing moved"
            );
            assert_eq!(backups(&f.home).len(), 16, "no folder created");
        }
    }
}

/// AC-2: the package is written and verified in `<backup>/incoming` before the original
/// moves, so a failure while staging leaves the original folder exactly in place.
#[test]
fn a_failure_while_staging_leaves_the_original_in_place() {
    let f = fixture();
    let before = originals(&f);
    let p = plan(&f);
    let r = result_of(&apply(&f, &p, &[&token(&p, "claude")], Some("mid_stage")));
    assert_eq!(
        host(&r, "claude")["receipt_status"],
        json!("pending"),
        "{r}"
    );
    assert_eq!(
        snapshot(&f.home.join(ROOTS[0].1)),
        before[0],
        "original in place"
    );
    let rp = recovery_plan(&f);
    let rb = &host(&rp, "claude")["restore_backup"];
    assert_eq!(rb["original_at"], json!("root"), "{rb}");
    assert_eq!(rb["action"], json!("release"), "{rb}");
}

/// AC-2: between moving the original out and moving the staged package in, the projection
/// is absent — never a partly written folder — and recovery restores the original.
#[test]
fn the_projection_is_absent_never_partial_between_the_two_moves() {
    let f = fixture();
    let before = originals(&f);
    let p = plan(&f);
    let r = result_of(&apply(&f, &p, &[&token(&p, "claude")], Some("mid_replace")));
    assert_eq!(
        host(&r, "claude")["receipt_status"],
        json!("pending"),
        "{r}"
    );
    let root = f.home.join(ROOTS[0].1);
    assert!(
        fs::symlink_metadata(&root).is_err(),
        "absent, not partial: {:?}",
        snapshot(&root).keys().collect::<Vec<_>>()
    );
    let dir = backup_of(&f.home, "claude");
    assert_package_installed(&dir.join("incoming"), &f.stage);
    result_of(&recovery_apply(&f, &recovery_plan(&f)));
    assert_eq!(snapshot(&root), before[0], "original restored");
}

/// AC-2: a successful replace leaves no staging folder behind.
#[test]
fn a_successful_replace_leaves_no_staging_folder() {
    let f = fixture();
    let p = plan(&f);
    result_of(&apply(&f, &p, &[&token(&p, "claude")], None));
    let dir = backup_of(&f.home, "claude");
    assert!(!dir.join("incoming").exists(), "{:?}", entries(&dir));
    assert_package_installed(&f.home.join(ROOTS[0].1), &f.stage);
}

/// Audit note X6 (#2486): the manifest re-check right after a restore. A seam changes one
/// restored file; the recovery must fail, write no RESTORED.json and leave the receipt
/// unrecovered.
#[test]
fn a_restored_folder_that_differs_from_the_manifest_is_not_marked_restored() {
    let f = fixture();
    let p = plan(&f);
    result_of(&apply(&f, &p, &[&token(&p, "claude")], None));
    let rp = recovery_plan(&f);
    let digest = rp["plan_digest"].as_str().unwrap();
    let out = run_cli(
        &[
            "recovery-apply",
            f.home.to_str().unwrap(),
            "--plan-digest",
            digest,
            "--permission-reference",
            "fixture-approved",
        ],
        Some(&rp),
        &f.home,
        Some("corrupt_after_restore"),
    );
    let v = envelope(&out);
    assert_eq!(v["ok"], json!(false), "{v}");
    assert!(v.to_string().contains("does not match the manifest"), "{v}");
    assert!(!backup_of(&f.home, "claude").join("RESTORED.json").exists());
    assert_ne!(
        host(&recovery_plan(&f), "claude")["receipt_status"],
        json!("recovered")
    );
}

/// #2544 (audit note C3 on #2505): when moving the staged package into place fails, the
/// original folder is moved back — whole, byte for byte — and the backup no longer holds it.
/// A debug-only seam makes that one rename fail; the move back runs for real.
#[test]
fn a_failed_move_into_place_puts_the_original_back() {
    let f = fixture();
    let before = originals(&f);
    let p = plan(&f);
    let r = result_of(&apply(
        &f,
        &p,
        &[&token(&p, "claude")],
        Some("stage_rename"),
    ));
    let h = host(&r, "claude");
    assert_eq!(h["receipt_status"], json!("pending"), "{r}");
    let why = h["reload_basis"].as_str().unwrap_or_default();
    assert!(
        why.contains("moving the staged package into place failed")
            && why.contains("the original was moved back"),
        "{why}"
    );
    let root = f.home.join(ROOTS[0].1);
    assert_eq!(snapshot(&root), before[0], "original back in place");
    let dir = backup_of(&f.home, "claude");
    assert!(
        fs::symlink_metadata(dir.join("moai")).is_err(),
        "{:?}",
        entries(&dir)
    );
    let rp = recovery_plan(&f);
    let rb = &host(&rp, "claude")["restore_backup"];
    assert_eq!(rb["original_at"], json!("root"), "{rb}");
    assert_eq!(rb["action"], json!("release"), "{rb}");
}

// ---- SPEC-20261125-update-installed-from-link: an apply that writes no file ----
//
// These cases run copies of the built binary from version folders inside the temporary HOME, so
// each copy has its own installation id.

/// A version an earlier build would have written into a receipt.
const EARLIER_VERSION: &str = "2026.11.0-rc.2+local.4";

fn state_dir(home: &Path) -> PathBuf {
    home.join(EARLIER_STATE)
}

fn version_dir(home: &Path, version: &str) -> PathBuf {
    state_dir(home).join("versions").join(version)
}

/// A copy of the built binary at `versions/<version>/bin/moai-cli`.
fn put_build(home: &Path, version: &str) -> PathBuf {
    let bin = version_dir(home, version).join("bin");
    fs::create_dir_all(&bin).unwrap();
    let exe = bin.join("moai-cli");
    common::copy_executable(Path::new(EXE), &exe);
    exe
}

/// Runs a copy with a cleared environment; the temporary HOME is also the current folder.
fn run_build(
    exe: &Path,
    home: &Path,
    args: &[&str],
    stdin: Option<&Value>,
) -> std::process::Output {
    let mut child = Command::new(exe)
        .args(args)
        .clean_env()
        .env("HOME", home)
        .env("PATH", "/usr/bin:/bin")
        .env("LC_ALL", "C")
        .current_dir(home)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    let mut pipe = child.stdin.take().unwrap();
    if let Some(v) = stdin {
        let _ = pipe.write_all(v.to_string().as_bytes());
    }
    drop(pipe);
    child.wait_with_output().unwrap()
}

fn plan_by(exe: &Path, home: &Path, pkg: &Path) -> Value {
    result_of(&run_build(
        exe,
        home,
        &["update-plan", home.to_str().unwrap(), pkg.to_str().unwrap()],
        None,
    ))
}

/// `update-apply` of `plan` by the copy, with the backup-then-replace token when one is given.
fn apply_by(exe: &Path, home: &Path, pkg: &Path, plan: &Value, token: Option<&str>) -> Value {
    let mut args = vec![
        "update-apply",
        home.to_str().unwrap(),
        pkg.to_str().unwrap(),
        "--plan-digest",
        plan["plan_digest"].as_str().unwrap(),
        "--permission-reference",
        "fixture-approved",
    ];
    if let Some(t) = token {
        args.extend(["--confirm-backup-replace", t]);
    }
    result_of(&run_build(exe, home, &args, Some(plan)))
}

/// Moves the Go-era skill aside and installs `pkg` as the copy.
fn replace_by(exe: &Path, home: &Path, pkg: &Path) {
    let p = plan_by(exe, home, pkg);
    let r = apply_by(exe, home, pkg, &p, Some(&token(&p, "claude")));
    assert_eq!(host(&r, "claude")["status"], json!("replaced"), "{r}");
    // Receipts carry second-resolution timestamps; keep the order of applies observable.
    sleep(Duration::from_millis(1100));
}

fn recovery_plan_by(exe: &Path, home: &Path) -> Value {
    result_of(&run_build(
        exe,
        home,
        &["recovery-plan", home.to_str().unwrap()],
        None,
    ))
}

fn recovery_apply_by(exe: &Path, home: &Path, plan: &Value) -> Value {
    result_of(&run_build(
        exe,
        home,
        &[
            "recovery-apply",
            home.to_str().unwrap(),
            "--plan-digest",
            plan["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "fixture-approved",
        ],
        Some(plan),
    ))
}

/// The installation id of a copy: the hash of its canonical path.
fn id_of(exe: &Path) -> String {
    sha256_hex(fs::canonicalize(exe).unwrap().to_str().unwrap().as_bytes())
}

fn receipt_file(home: &Path, id: &str) -> PathBuf {
    state_dir(home)
        .join("installations")
        .join(id)
        .join("hosts/moai/claude.json")
}

fn read_receipt(home: &Path, id: &str) -> Value {
    serde_json::from_slice(&fs::read(receipt_file(home, id)).unwrap()).unwrap()
}

fn set_receipt_field(home: &Path, id: &str, field: &str, value: &str) {
    let mut r = read_receipt(home, id);
    r[field] = json!(value);
    let bytes = serde_json::to_vec_pretty(&r).unwrap();
    fs::write(receipt_file(home, id), bytes).unwrap();
}

/// The receipt of a backup-then-replace apply holds what a restore needs, and `recovery-plan`
/// offers it: the state a case starts from, before its no-write apply.
fn restorable_receipt(exe: &Path, home: &Path) -> Value {
    let r = read_receipt(home, &id_of(exe));
    assert!(r["backup"].is_object(), "{r}");
    assert!(
        r["entries"].as_array().is_some_and(|e| !e.is_empty()),
        "{r}"
    );
    let rp = recovery_plan_by(exe, home);
    assert_eq!(
        host(&rp, "claude")["restore_backup"]["action"],
        json!("restore_backup"),
        "{rp}"
    );
    r
}

/// What a no-write apply must not take from the receipt that made the backup: `recovery-plan`
/// still offers the restore, and the fields that describe that apply are as they were. Returns
/// the recovery plan.
fn assert_own_backup_kept(exe: &Path, home: &Path, before: &Value) -> Value {
    let rp = recovery_plan_by(exe, home);
    assert_eq!(
        host(&rp, "claude")["restore_backup"]["action"],
        json!("restore_backup"),
        "the restore of the original folder is still offered: {rp}"
    );
    let after = read_receipt(home, &id_of(exe));
    for field in ["backup", "entries", "operation", "version_before"] {
        assert_eq!(after[field], before[field], "{field}: {after}");
    }
    assert_eq!(
        after["package_version"],
        json!(moai_cli::BUILD_VERSION),
        "{after}"
    );
    assert_eq!(after["status"], json!("committed"), "{after}");
    rp
}

// REQ-4: the receipt that moved the user's folder aside takes this build's version from a
// no-write apply and keeps everything a restore reads.
#[test]
fn a_no_write_apply_keeps_the_own_backup_that_recovery_restores() {
    let f = fixture();
    let original = originals(&f);
    let a = put_build(&f.home, "a");
    replace_by(&a, &f.home, &f.stage);
    set_receipt_field(&f.home, &id_of(&a), "package_version", EARLIER_VERSION);
    let before = restorable_receipt(&a, &f.home);

    let p = plan_by(&a, &f.home, &f.stage);
    let r = apply_by(&a, &f.home, &f.stage, &p, None);
    assert_eq!(host(&r, "claude")["status"], json!("unchanged"), "{r}");

    let rp = assert_own_backup_kept(&a, &f.home, &before);
    let restored = recovery_apply_by(&a, &f.home, &rp);
    assert_eq!(restored["status"], json!("success"), "{restored}");
    assert_eq!(snapshot(&f.home.join(ROOTS[0].1)), original[0]);
}

/// `update` from the copy's own folder (its package and agents beside it); the whole envelope.
fn update_by(exe: &Path, home: &Path, args: &[&str]) -> Value {
    let mut all = vec!["update"];
    all.extend_from_slice(args);
    envelope(&run_build(exe, home, &all, None))
}

// REQ-4 going back: a newer build's no-write receipt is the baseline, and the older build's
// confirmed `update --allow-older` writes no file. Its own receipt keeps the backup it made.
#[test]
fn a_no_write_apply_going_back_keeps_the_own_backup_and_refreshes_the_receipt() {
    let f = fixture();
    let original = originals(&f);
    let a = put_build(&f.home, "a");
    let folder = version_dir(&f.home, "a");
    // The version folder gets what an installed one has: a package and an agents folder.
    let pkg = folder.join("skills/moai");
    write_pkg(&pkg);
    fs::create_dir_all(folder.join("agents")).unwrap();
    replace_by(&a, &f.home, &pkg);
    set_receipt_field(&f.home, &id_of(&a), "package_version", EARLIER_VERSION);

    // A second build applies the same package: no file moves, and it gets a receipt of its own.
    let b = put_build(&f.home, "b");
    let p = plan_by(&b, &f.home, &pkg);
    let r = apply_by(&b, &f.home, &pkg, &p, None);
    assert_eq!(host(&r, "claude")["status"], json!("unchanged"), "{r}");
    // That receipt names a version one head above this build.
    let head: u32 = moai_cli::BUILD_VERSION
        .split('.')
        .next()
        .unwrap()
        .parse()
        .unwrap();
    let newer = format!("{}.0001.1", head + 1);
    set_receipt_field(&f.home, &id_of(&b), "package_version", &newer);
    // Its owned list also gets a file that neither the package nor the disk holds (it sorts
    // after every package path): the plan has no row for it, so nothing is written, and this
    // baseline's list differs from the older build's own. Another digest for a package file
    // would instead be a baseline move, which is not the no-write path.
    let mut theirs = read_receipt(&f.home, &id_of(&b));
    theirs["owned"].as_array_mut().unwrap().push(json!({
        "host": "claude",
        "path": "zz-retired/note.md",
        "sha256": format!("sha256:{}", sha256_hex(b"retired")),
    }));
    fs::write(
        receipt_file(&f.home, &id_of(&b)),
        serde_json::to_vec_pretty(&theirs).unwrap(),
    )
    .unwrap();
    sleep(Duration::from_millis(1100));

    // Without the flag the older build is blocked: the going-back path is the one under test.
    let preview = update_by(&a, &f.home, &[]);
    let blocked = host(&preview["result"]["plan"], "claude");
    assert_eq!(blocked["preflight"], json!("blocked"), "{preview}");
    assert!(
        blocked["blocked_reason"]
            .as_str()
            .is_some_and(|why| why.contains(&newer)),
        "{preview}"
    );

    let before = restorable_receipt(&a, &f.home);
    assert_ne!(before["owned"], theirs["owned"]);
    let preview = update_by(&a, &f.home, &["--allow-older"]);
    let digest = preview["result"]["plan_digest"]
        .as_str()
        .unwrap_or_else(|| panic!("{preview}"));
    let back = update_by(&a, &f.home, &["--allow-older", "--confirm", digest]);
    assert_eq!(back["ok"], json!(true), "{back}");
    assert_eq!(
        host(&back["result"]["result"], "claude")["status"],
        json!("unchanged"),
        "{back}"
    );

    let rp = assert_own_backup_kept(&a, &f.home, &before);
    let after = read_receipt(&f.home, &id_of(&a));
    assert_eq!(after["owned"], theirs["owned"], "{after}");
    assert_ne!(after["package_version"], before["package_version"]);
    assert_ne!(after["finished_at"], before["finished_at"], "{after}");
    // The refreshed finish time makes this receipt the baseline again.
    let p = plan_by(&a, &f.home, &pkg);
    assert_eq!(host(&p, "claude")["preflight"], json!("ok"), "{p}");
    let restored = recovery_apply_by(&a, &f.home, &rp);
    assert_eq!(restored["status"], json!("success"), "{restored}");
    assert_eq!(snapshot(&f.home.join(ROOTS[0].1)), original[0]);
}

// REQ-5: an own receipt that is neither committed nor recovered is left byte for byte, even when
// the baseline is a newer folder's committed receipt of another version.
#[test]
fn a_no_write_apply_never_rewrites_an_own_failed_receipt() {
    let f = fixture();
    let a = put_build(&f.home, "a");
    replace_by(&a, &f.home, &f.stage);

    // A second build really writes a changed file, so it commits a receipt of its own.
    fs::write(f.stage.join("workflows/update.md"), "# UPDATE\nrust v2\n").unwrap();
    let b = put_build(&f.home, "b");
    let p = plan_by(&b, &f.home, &f.stage);
    let r = apply_by(&b, &f.home, &f.stage, &p, None);
    assert_eq!(host(&r, "claude")["status"], json!("applied"), "{r}");
    set_receipt_field(&f.home, &id_of(&b), "package_version", EARLIER_VERSION);
    set_receipt_field(&f.home, &id_of(&a), "status", "failed");
    let mine = read_receipt(&f.home, &id_of(&a));
    let theirs = read_receipt(&f.home, &id_of(&b));
    assert_eq!(theirs["status"], json!("committed"), "{theirs}");
    assert!(
        mine["finished_at"].as_str().unwrap() < theirs["finished_at"].as_str().unwrap(),
        "{mine} {theirs}"
    );
    let bytes = fs::read(receipt_file(&f.home, &id_of(&a))).unwrap();

    // The baseline is the second build's receipt, and nothing is left to write.
    let p = plan_by(&a, &f.home, &f.stage);
    assert_eq!(p["source_version"], json!(EARLIER_VERSION), "{p}");
    assert!(
        host(&p, "claude")["receipt_from"]
            .as_str()
            .is_some_and(|from| from.contains(&id_of(&b))),
        "{p}"
    );
    let r = apply_by(&a, &f.home, &f.stage, &p, None);
    assert_eq!(host(&r, "claude")["status"], json!("unchanged"), "{r}");
    assert_eq!(
        fs::read(receipt_file(&f.home, &id_of(&a))).unwrap(),
        bytes,
        "the failed receipt is not rewritten"
    );
}

// ---- SPEC-20261126-keep-restore-backup: an ordinary apply keeps the restore offer ----

fn put_file(dir: &Path, rel: &str, body: &str) {
    let p = dir.join(rel);
    fs::create_dir_all(p.parent().unwrap()).unwrap();
    fs::write(p, body).unwrap();
}

/// A plan and an apply of it by the copy, with no confirmation token.
fn ordinary_apply(exe: &Path, home: &Path, pkg: &Path) -> (Value, Value) {
    let p = plan_by(exe, home, pkg);
    let r = apply_by(exe, home, pkg, &p, None);
    (p, r)
}

/// The Claude host's plan rows the apply writes, as (path, action), in write order.
fn plan_writes(plan: &Value) -> Vec<(String, String)> {
    host(plan, "claude")["files"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|f| f["action"] == json!("create") || f["action"] == json!("update"))
        .map(|f| {
            (
                f["path"].as_str().unwrap().to_string(),
                f["action"].as_str().unwrap().to_string(),
            )
        })
        .collect()
}

fn plan_row<'a>(plan: &'a Value, path: &str) -> &'a Value {
    host(plan, "claude")["files"]
        .as_array()
        .unwrap()
        .iter()
        .find(|f| f["path"] == json!(path))
        .unwrap_or_else(|| panic!("no plan row for {path}: {plan}"))
}

/// The receipt entries as (path, action), in receipt order.
fn entry_rows(receipt: &Value) -> Vec<(String, String)> {
    receipt["entries"]
        .as_array()
        .unwrap()
        .iter()
        .map(|e| {
            (
                e["path"].as_str().unwrap().to_string(),
                e["action"].as_str().unwrap().to_string(),
            )
        })
        .collect()
}

/// The receipt holds no backup and lists the writes of `plan` only.
fn assert_writes_only(exe: &Path, home: &Path, plan: &Value) {
    let receipt = read_receipt(home, &id_of(exe));
    assert_eq!(receipt["status"], json!("committed"), "{receipt}");
    assert!(receipt["backup"].is_null(), "no backup: {receipt}");
    let writes = plan_writes(plan);
    assert!(!writes.is_empty(), "the apply under test writes: {plan}");
    assert_eq!(
        entry_rows(&receipt),
        writes,
        "entries are the writes: {receipt}"
    );
}

/// The receipt holds no backup and lists the writes of `plan` only, and no restore row is shown.
fn assert_writes_only_and_no_offer(exe: &Path, home: &Path, plan: &Value) {
    assert_writes_only(exe, home, plan);
    let rp = recovery_plan_by(exe, home);
    assert!(
        host(&rp, "claude")["restore_backup"].is_null(),
        "no restore row: {rp}"
    );
}

/// The restore is offered, the receipt carries the replace receipt's backup, and its entries are
/// the writes of `plan` followed by one for every other installed file. Returns the recovery plan.
fn assert_offer_covers_root(exe: &Path, home: &Path, replaced: &Value, plan: &Value) -> Value {
    let rp = recovery_plan_by(exe, home);
    let offer = &host(&rp, "claude")["restore_backup"];
    assert_eq!(
        offer["action"],
        json!("restore_backup"),
        "the restore of the original folder is still offered: {rp}"
    );
    assert_eq!(offer["original_at"], json!("backup"), "{rp}");

    let receipt = read_receipt(home, &id_of(exe));
    assert_eq!(receipt["status"], json!("committed"), "{receipt}");
    assert_eq!(receipt["operation"], json!("update.apply"), "{receipt}");
    assert_eq!(receipt["backup"], replaced["backup"], "{receipt}");

    let writes = plan_writes(plan);
    let rows = entry_rows(&receipt);
    assert!(rows.len() >= writes.len(), "{receipt}");
    assert_eq!(
        rows[..writes.len()],
        writes[..],
        "writes come first, in write order: {receipt}"
    );

    let owned: BTreeMap<&str, &Value> = receipt["owned"]
        .as_array()
        .unwrap()
        .iter()
        .map(|o| (o["path"].as_str().unwrap(), &o["sha256"]))
        .collect();
    let entries = receipt["entries"].as_array().unwrap();
    for e in &entries[writes.len()..] {
        let path = e["path"].as_str().unwrap();
        let digest = owned
            .get(path)
            .unwrap_or_else(|| panic!("{path} is not owned: {receipt}"));
        assert_eq!(e["action"], json!("unchanged"), "{e}");
        assert_eq!(&e["before_sha256"], *digest, "{e}");
        assert_eq!(&e["after_sha256"], *digest, "{e}");
        assert_eq!(e["published"], json!(false), "{e}");
        assert_eq!(e["error"], Value::Null, "{e}");
    }

    let installed = snapshot(&home.join(ROOTS[0].1));
    let mut listed: Vec<&str> = entries
        .iter()
        .map(|e| e["path"].as_str().unwrap())
        .collect();
    listed.sort_unstable();
    assert_eq!(
        listed,
        installed.keys().map(String::as_str).collect::<Vec<_>>(),
        "one entry for each installed file: {receipt}"
    );
    for e in entries {
        let (bytes, _) = &installed[e["path"].as_str().unwrap()];
        assert_eq!(
            e["after_sha256"],
            json!(format!("sha256:{}", sha256_hex(bytes))),
            "{e}"
        );
    }
    rp
}

/// Recovery puts the original folder back, bytes and modes, and keeps what was installed in
/// `<backup>/replaced`.
fn assert_restores_original(
    exe: &Path,
    home: &Path,
    rp: &Value,
    original: &BTreeMap<String, (Vec<u8>, u32)>,
) {
    let root = home.join(ROOTS[0].1);
    let installed = snapshot(&root);
    let r = recovery_apply_by(exe, home, rp);
    assert_eq!(r["status"], json!("success"), "{r}");
    assert_eq!(host(&r, "claude")["backup_restored"], json!(true), "{r}");
    assert_eq!(
        host(&r, "claude")["receipt_status"],
        json!("recovered"),
        "{r}"
    );
    assert_eq!(&snapshot(&root), original, "original restored");
    let dir = backup_of(home, "claude");
    assert_eq!(
        snapshot(&dir.join("replaced")),
        installed,
        "the installed package is kept"
    );
    assert!(dir.join("RESTORED.json").is_file());
    let receipt = read_receipt(home, &id_of(exe));
    assert_eq!(receipt["status"], json!("recovered"), "{receipt}");
}

// SPEC-20261126-keep-restore-backup AC-1
#[test]
fn keep_restore_an_ordinary_apply_keeps_the_backup_that_recovery_restores() {
    let f = fixture();
    let original = originals(&f);
    let a = put_build(&f.home, "a");
    replace_by(&a, &f.home, &f.stage);
    let replaced = restorable_receipt(&a, &f.home);

    put_file(&f.stage, "workflows/update.md", "# UPDATE\nrust v2\n");
    put_file(&f.stage, "references/new.md", "new v2\n");
    let (p, r) = ordinary_apply(&a, &f.home, &f.stage);
    assert_eq!(host(&r, "claude")["status"], json!("applied"), "{r}");
    assert_offer_covers_root(&a, &f.home, &replaced, &p);

    put_file(&f.stage, "workflows/update.md", "# UPDATE\nrust v3\n");
    let (p, r) = ordinary_apply(&a, &f.home, &f.stage);
    assert_eq!(host(&r, "claude")["status"], json!("applied"), "{r}");
    let rp = assert_offer_covers_root(&a, &f.home, &replaced, &p);

    assert_package_installed(&f.home.join(ROOTS[0].1), &f.stage);
    assert_restores_original(&a, &f.home, &rp, &original[0]);
}

// SPEC-20261126-keep-restore-backup AC-2
#[test]
fn keep_restore_adds_nothing_without_an_earlier_backup() {
    let f = fixture();
    // With the unmanaged folder out of the way, nothing is ever replaced.
    fs::rename(f.home.join(ROOTS[0].1), f.home.join("go-skill-aside")).unwrap();
    let a = put_build(&f.home, "a");

    let (p, r) = ordinary_apply(&a, &f.home, &f.stage);
    assert_eq!(host(&r, "claude")["status"], json!("applied"), "{r}");
    assert_writes_only_and_no_offer(&a, &f.home, &p);

    put_file(&f.stage, "workflows/update.md", "# UPDATE\nrust v2\n");
    let (p, r) = ordinary_apply(&a, &f.home, &f.stage);
    assert_eq!(host(&r, "claude")["status"], json!("applied"), "{r}");
    assert_writes_only_and_no_offer(&a, &f.home, &p);
    assert!(backups(&f.home).is_empty());
}

// SPEC-20261126-keep-restore-backup AC-3
#[test]
fn keep_restore_covers_a_baseline_move() {
    let f = fixture();
    let original = originals(&f);
    let a = put_build(&f.home, "a");
    replace_by(&a, &f.home, &f.stage);
    let replaced = restorable_receipt(&a, &f.home);

    // The package and the installed file change to the same bytes: only the baseline moves.
    let moved = "# UPDATE\nalready in place\n";
    put_file(&f.stage, "workflows/update.md", moved);
    put_file(&f.home.join(ROOTS[0].1), "workflows/update.md", moved);
    let (p, r) = ordinary_apply(&a, &f.home, &f.stage);
    let row = plan_row(&p, "workflows/update.md");
    assert_eq!(row["action"], json!("unchanged"), "{p}");
    assert_ne!(row["baseline_sha256"], row["current_sha256"], "{p}");
    assert!(plan_writes(&p).is_empty(), "{p}");
    assert_eq!(host(&r, "claude")["status"], json!("unchanged"), "{r}");
    assert_eq!(
        host(&r, "claude")["receipt_status"],
        json!("committed"),
        "{r}"
    );

    let rp = assert_offer_covers_root(&a, &f.home, &replaced, &p);
    assert_restores_original(&a, &f.home, &rp, &original[0]);
}

// SPEC-20261126-keep-restore-backup AC-4
#[test]
fn keep_restore_is_blocked_by_a_kept_user_edit() {
    let f = fixture();
    let original = originals(&f);
    let a = put_build(&f.home, "a");
    replace_by(&a, &f.home, &f.stage);

    let root = f.home.join(ROOTS[0].1);
    // Edited where the package did not change (kept), edited where it did (conflict), and one
    // new package file so that the apply writes.
    put_file(&root, "references/routes.json", "{\"routes\":\"mine\"}\n");
    put_file(&root, "workflows/update.md", "user edit\n");
    put_file(&f.stage, "workflows/update.md", "# UPDATE\nrust v2\n");
    put_file(&f.stage, "references/new.md", "new v2\n");
    let (p, r) = ordinary_apply(&a, &f.home, &f.stage);
    for (path, action) in [
        ("references/routes.json", "preserved"),
        ("workflows/update.md", "conflict"),
        ("references/new.md", "create"),
    ] {
        assert_eq!(plan_row(&p, path)["action"], json!(action), "{p}");
    }
    assert_eq!(host(&r, "claude")["status"], json!("applied"), "{r}");

    let rp = recovery_plan_by(&a, &f.home);
    let offer = &host(&rp, "claude")["restore_backup"];
    assert_eq!(offer["action"], json!("none"), "{rp}");
    assert_eq!(offer["reason"], json!("user_changed"), "{rp}");

    let installed = snapshot(&root);
    let dir = backup_of(&f.home, "claude");
    assert_eq!(snapshot(&dir.join("moai")), original[0]);
    let out = recovery_apply_by(&a, &f.home, &rp);
    assert_eq!(out["status"], json!("partial"), "{out}");
    assert_eq!(snapshot(&root), installed, "installed folder unchanged");
    assert_eq!(snapshot(&dir.join("moai")), original[0], "backup unchanged");
    assert!(!dir.join("RESTORED.json").exists());
}

// SPEC-20261126-keep-restore-backup AC-5
#[test]
fn keep_restore_leaves_a_failed_apply_to_per_file_recovery() {
    let f = fixture();
    let a = put_build(&f.home, "a");
    replace_by(&a, &f.home, &f.stage);
    restorable_receipt(&a, &f.home);

    let root = f.home.join(ROOTS[0].1);
    let skill_before = fs::read(root.join("SKILL.md")).unwrap();
    put_file(
        &f.stage,
        "SKILL.md",
        "---\nname: moai\ndescription: fixture package\n---\n# MoAI\nrust v2\n",
    );
    put_file(&f.stage, "workflows/new.md", "# NEW\n");
    // A plain file where the folder was: the apply fails after SKILL.md.
    fs::remove_dir_all(root.join("workflows")).unwrap();
    fs::write(root.join("workflows"), "not a directory\n").unwrap();
    let (p, r) = ordinary_apply(&a, &f.home, &f.stage);
    assert_eq!(host(&r, "claude")["status"], json!("failed"), "{r}");
    assert_eq!(
        host(&r, "claude")["receipt_status"],
        json!("pending"),
        "{r}"
    );

    let writes = plan_writes(&p);
    assert_eq!(writes[0].0, "SKILL.md", "{p}");
    assert!(writes.iter().any(|w| w.0 == "workflows/new.md"), "{p}");
    let receipt = read_receipt(&f.home, &id_of(&a));
    assert_eq!(receipt["status"], json!("pending"), "{receipt}");
    assert_eq!(receipt["operation"], json!("update.apply"), "{receipt}");
    assert!(receipt["backup"].is_null(), "no backup: {receipt}");
    assert_eq!(
        entry_rows(&receipt),
        writes,
        "entries are the writes: {receipt}"
    );
    assert_eq!(receipt["entries"][0]["published"], json!(true), "{receipt}");

    let rp = recovery_plan_by(&a, &f.home);
    let h = host(&rp, "claude");
    assert_eq!(h["receipt_status"], json!("pending"), "{rp}");
    assert!(h["restore_backup"].is_null(), "no restore row: {rp}");
    let files = h["files"].as_array().unwrap();
    let listed: Vec<&str> = files.iter().map(|x| x["path"].as_str().unwrap()).collect();
    assert_eq!(
        listed,
        writes.iter().map(|w| w.0.as_str()).collect::<Vec<_>>(),
        "{rp}"
    );
    let row = |path: &str| files.iter().find(|x| x["path"] == json!(path)).unwrap();
    assert_eq!(row("SKILL.md")["class"], json!("matches_after"), "{rp}");
    assert_eq!(row("SKILL.md")["action"], json!("restore"), "{rp}");
    assert_eq!(
        row("workflows/new.md")["reason"],
        json!("not_a_regular_file"),
        "{rp}"
    );
    let out = recovery_apply_by(&a, &f.home, &rp);
    assert_eq!(out["status"], json!("partial"), "{out}");
    assert_eq!(fs::read(root.join("SKILL.md")).unwrap(), skill_before);
}

// SPEC-20261126-keep-restore-backup AC-6
#[test]
fn keep_restore_is_not_taken_from_another_installation() {
    let f = fixture();
    let a = put_build(&f.home, "a");
    let b = put_build(&f.home, "b");
    replace_by(&a, &f.home, &f.stage);
    restorable_receipt(&a, &f.home);
    let bytes = fs::read(receipt_file(&f.home, &id_of(&a))).unwrap();

    put_file(&f.stage, "workflows/update.md", "# UPDATE\nrust v2\n");
    let (p, r) = ordinary_apply(&b, &f.home, &f.stage);
    assert!(
        host(&p, "claude")["receipt_from"]
            .as_str()
            .is_some_and(|from| from.contains(&id_of(&a))),
        "the replacing build's receipt is the baseline: {p}"
    );
    assert_eq!(host(&r, "claude")["status"], json!("applied"), "{r}");
    assert_writes_only(&b, &f.home, &p);
    assert_eq!(
        fs::read(receipt_file(&f.home, &id_of(&a))).unwrap(),
        bytes,
        "the other installation's receipt is untouched"
    );
}

// SPEC-20261126-keep-restore-backup AC-7
#[test]
fn keep_restore_never_carries_a_used_backup() {
    let f = fixture();
    let a = put_build(&f.home, "a");
    replace_by(&a, &f.home, &f.stage);
    restorable_receipt(&a, &f.home);
    let rp = recovery_plan_by(&a, &f.home);
    let restored = recovery_apply_by(&a, &f.home, &rp);
    assert_eq!(restored["status"], json!("success"), "{restored}");
    let used = read_receipt(&f.home, &id_of(&a));
    assert_eq!(used["status"], json!("recovered"), "{used}");
    assert!(used["backup"].is_object(), "{used}");

    // The restored folder is unmanaged again; out of the way, the apply installs the package.
    fs::rename(f.home.join(ROOTS[0].1), f.home.join("go-skill-aside")).unwrap();
    let (p, r) = ordinary_apply(&a, &f.home, &f.stage);
    assert_eq!(host(&r, "claude")["status"], json!("applied"), "{r}");
    assert_writes_only_and_no_offer(&a, &f.home, &p);
}

// SPEC-20261126-keep-restore-backup AC-10
#[test]
fn keep_restore_covers_what_another_build_wrote() {
    let f = fixture();
    let original = originals(&f);
    let a = put_build(&f.home, "a");
    let b = put_build(&f.home, "b");
    replace_by(&a, &f.home, &f.stage);
    let replaced = restorable_receipt(&a, &f.home);

    let theirs = f._tmp.path().join("stage-b");
    write_pkg(&theirs);
    put_file(&theirs, "workflows/update.md", "# UPDATE\nfrom b\n");
    put_file(&theirs, "references/new.md", "added by b\n");
    let (_, r) = ordinary_apply(&b, &f.home, &theirs);
    assert_eq!(host(&r, "claude")["status"], json!("applied"), "{r}");
    sleep(Duration::from_millis(1100));

    let (p, r) = ordinary_apply(&a, &f.home, &f.stage);
    assert!(
        host(&p, "claude")["receipt_from"]
            .as_str()
            .is_some_and(|from| from.contains(&id_of(&b))),
        "the other build's receipt is the baseline: {p}"
    );
    assert_eq!(host(&r, "claude")["status"], json!("applied"), "{r}");

    let rp = assert_offer_covers_root(&a, &f.home, &replaced, &p);
    assert_restores_original(&a, &f.home, &rp, &original[0]);
    let kept = backup_of(&f.home, "claude").join("replaced/references/new.md");
    assert_eq!(fs::read_to_string(kept).unwrap(), "added by b\n");
}

// ---- SPEC-20261127-recovery-prior-install: the restore offer survives a version upgrade ----

type Snapshot = BTreeMap<String, (Vec<u8>, u32)>;

/// `a` replaced the Go-era folder and holds its backup. With `applied`, `b` (another build, so
/// another installation id) then took over with an ordinary apply; without it `b` is only
/// installed. Returns the fixture, the original folder, `a`, `b` and the plan of `b`'s apply
/// (`null` without one).
fn upgraded(applied: bool) -> (Fixture, Snapshot, PathBuf, PathBuf, Value) {
    let f = fixture();
    let original = originals(&f).remove(0);
    let a = put_build(&f.home, "a");
    let b = put_build(&f.home, "b");
    replace_by(&a, &f.home, &f.stage);
    restorable_receipt(&a, &f.home);
    let mut plan = Value::Null;
    if applied {
        put_file(&f.stage, "workflows/update.md", "# UPDATE\nrust v2\n");
        put_file(&f.stage, "references/new.md", "new v2\n");
        let (p, r) = ordinary_apply(&b, &f.home, &f.stage);
        assert_eq!(host(&r, "claude")["status"], json!("applied"), "{r}");
        plan = p;
    }
    (f, original, a, b, plan)
}

/// The plan offers the restore of the original folder out of the earlier installation `a`.
fn assert_offers_prior(rp: &Value, a: &Path) {
    let offer = &host(rp, "claude")["restore_backup"];
    assert_eq!(offer["action"], json!("restore_backup"), "{rp}");
    assert_eq!(offer["original_at"], json!("backup"), "{rp}");
    assert_eq!(offer["from_install"], json!(id_of(a)), "{rp}");
}

/// Every regular file under `home` by content digest and count, leaving out what a recovery
/// rewrites on purpose (the host receipts and recovery records) and the build copies.
fn home_digests(home: &Path) -> BTreeMap<String, usize> {
    fn walk(dir: &Path, skip: &Path, out: &mut BTreeMap<String, usize>) {
        for e in fs::read_dir(dir).unwrap() {
            let p = e.unwrap().path();
            let m = fs::symlink_metadata(&p).unwrap();
            if m.is_dir() {
                if p != skip {
                    walk(&p, skip, out);
                }
            } else if m.is_file() {
                let name = p.file_name().unwrap().to_str().unwrap();
                let rewritten = p.components().any(|c| c.as_os_str() == "installations")
                    && (name == "claude.json" || name == "claude.recovery.json");
                if !rewritten {
                    *out.entry(sha256_hex(&fs::read(&p).unwrap())).or_default() += 1;
                }
            }
        }
    }
    let mut out = BTreeMap::new();
    walk(home, &state_dir(home).join("versions"), &mut out);
    out
}

// SPEC-20261127-recovery-prior-install AC-1
#[test]
fn prior_install_offers_the_restore_after_an_ordinary_apply() {
    let (f, _, a, b, _) = upgraded(false);
    let a_bytes = fs::read(receipt_file(&f.home, &id_of(&a))).unwrap();
    put_file(&f.stage, "workflows/update.md", "# UPDATE\nrust v2\n");
    put_file(&f.stage, "references/new.md", "new v2\n");
    let (p, r) = ordinary_apply(&b, &f.home, &f.stage);
    assert_eq!(host(&r, "claude")["status"], json!("applied"), "{r}");

    assert_writes_only(&b, &f.home, &p);
    assert_offers_prior(&recovery_plan_by(&b, &f.home), &a);
    assert_eq!(
        fs::read(receipt_file(&f.home, &id_of(&a))).unwrap(),
        a_bytes,
        "the earlier installation's receipt is untouched by the apply"
    );
}

// SPEC-20261127-recovery-prior-install AC-2
#[test]
fn prior_install_offers_the_restore_before_the_new_build_applies() {
    let (f, _, a, b, _) = upgraded(false);
    assert!(
        !receipt_file(&f.home, &id_of(&b)).exists(),
        "the new build has applied nothing yet"
    );
    assert_offers_prior(&recovery_plan_by(&b, &f.home), &a);
}

/// Restores across installations and checks that nothing in the temporary HOME is lost.
fn restore_across(applied: bool) {
    let (f, original, a, b, _) = upgraded(applied);
    let a_before = read_receipt(&f.home, &id_of(&a));
    let rp = recovery_plan_by(&b, &f.home);
    assert_offers_prior(&rp, &a);

    let root = f.home.join(ROOTS[0].1);
    let installed = snapshot(&root);
    let before = home_digests(&f.home);
    let out = recovery_apply_by(&b, &f.home, &rp);
    assert_eq!(out["status"], json!("success"), "{out}");
    assert_eq!(
        host(&out, "claude")["backup_restored"],
        json!(true),
        "{out}"
    );
    let after = home_digests(&f.home);
    let lost: Vec<&String> = before
        .iter()
        .filter(|(d, n)| after.get(*d).copied().unwrap_or(0) < **n)
        .map(|(d, _)| d)
        .collect();
    eprintln!(
        "prior install restore (applied={applied}): files before {}, after {}, lost {}",
        before.values().sum::<usize>(),
        after.values().sum::<usize>(),
        lost.len()
    );
    assert!(lost.is_empty(), "files lost: {lost:?}");

    assert_eq!(
        snapshot(&root),
        original,
        "original restored, bytes and modes"
    );
    let dir = backup_of(&f.home, "claude");
    assert_eq!(
        snapshot(&dir.join("replaced")),
        installed,
        "what was installed is kept in the backup"
    );
    assert!(dir.join("RESTORED.json").is_file());

    let a_after = read_receipt(&f.home, &id_of(&a));
    assert_eq!(a_after["status"], json!("recovered"), "{a_after}");
    assert_eq!(a_after["owned"], json!([]), "{a_after}");
    assert_eq!(a_after["backup"], a_before["backup"], "{a_after}");
    assert_eq!(
        a_after["finished_at"], a_before["finished_at"],
        "the earlier receipt keeps its finish time, so no two receipts finish together"
    );
    if applied {
        let b_after = read_receipt(&f.home, &id_of(&b));
        assert_eq!(b_after["status"], json!("recovered"), "{b_after}");
        assert_eq!(b_after["owned"], json!([]), "{b_after}");
    } else {
        assert!(
            !receipt_file(&f.home, &id_of(&b)).exists(),
            "a build that applied nothing gets no receipt"
        );
    }
}

// SPEC-20261127-recovery-prior-install AC-3
#[test]
fn prior_install_restore_moves_and_deletes_nothing() {
    restore_across(true);
    restore_across(false);
}

// SPEC-20261127-recovery-prior-install AC-4
#[test]
fn prior_install_offer_is_blocked_by_a_user_edit() {
    let (f, original, a, b, _) = upgraded(true);
    let root = f.home.join(ROOTS[0].1);
    put_file(&root, "notes/mine.txt", "mine\n");
    let rp = recovery_plan_by(&b, &f.home);
    assert!(
        host(&rp, "claude")["restore_backup"].is_null(),
        "no offer over a file neither receipt knows: {rp}"
    );

    let installed = snapshot(&root);
    let dir = backup_of(&f.home, "claude");
    let _ = recovery_apply_by(&b, &f.home, &rp);
    assert_eq!(snapshot(&root), installed, "installed folder unchanged");
    assert_eq!(snapshot(&dir.join("moai")), original, "backup unchanged");
    assert!(!dir.join("RESTORED.json").exists());
    assert_eq!(
        read_receipt(&f.home, &id_of(&a))["status"],
        json!("committed")
    );
}

// SPEC-20261127-recovery-prior-install AC-5
#[test]
fn prior_install_offer_waits_for_own_pending_recovery() {
    let f = fixture();
    let a = put_build(&f.home, "a");
    let b = put_build(&f.home, "b");
    replace_by(&a, &f.home, &f.stage);
    restorable_receipt(&a, &f.home);

    let root = f.home.join(ROOTS[0].1);
    put_file(
        &f.stage,
        "SKILL.md",
        "---\nname: moai\ndescription: fixture package\n---\n# MoAI\nrust v2\n",
    );
    put_file(&f.stage, "workflows/new.md", "# NEW\n");
    // A plain file where the folder was: the apply fails after SKILL.md.
    fs::remove_dir_all(root.join("workflows")).unwrap();
    fs::write(root.join("workflows"), "not a directory\n").unwrap();
    let (_, r) = ordinary_apply(&b, &f.home, &f.stage);
    assert_eq!(host(&r, "claude")["status"], json!("failed"), "{r}");

    let rp = recovery_plan_by(&b, &f.home);
    let h = host(&rp, "claude");
    assert_eq!(h["receipt_status"], json!("pending"), "{rp}");
    assert!(h["restore_backup"].is_null(), "no offer yet: {rp}");
    assert!(
        !h["files"].as_array().unwrap().is_empty(),
        "per-file recovery stays first: {rp}"
    );
}

// SPEC-20261127-recovery-prior-install AC-6
#[test]
fn prior_install_offer_is_not_guessed_between_two_backups() {
    let (f, _, a, b, _) = upgraded(false);
    // A second earlier installation that finished in the same second and holds a backup too.
    let other = sha256_hex(b"fixture-second-earlier-installation");
    let mut copy = read_receipt(&f.home, &id_of(&a));
    copy["install_id"] = json!(other);
    let path = receipt_file(&f.home, &other);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(&path, serde_json::to_vec_pretty(&copy).unwrap()).unwrap();
    let rp = recovery_plan_by(&b, &f.home);
    assert!(
        host(&rp, "claude")["restore_backup"].is_null(),
        "two backups that finished together: {rp}"
    );

    // An earlier installation whose receipt cannot be read is not guessed past either.
    fs::write(&path, "not json\n").unwrap();
    let rp = recovery_plan_by(&b, &f.home);
    assert!(
        host(&rp, "claude")["restore_backup"].is_null(),
        "an unreadable receipt: {rp}"
    );

    // Control: without the second installation the offer is there again.
    fs::remove_dir_all(state_dir(&f.home).join("installations").join(&other)).unwrap();
    assert_offers_prior(&recovery_plan_by(&b, &f.home), &a);
}

// SPEC-20261127-recovery-prior-install AC-7
#[test]
fn prior_install_used_backup_is_not_offered_again() {
    let (f, original, a, b, _) = upgraded(true);
    let rp = recovery_plan_by(&b, &f.home);
    assert_offers_prior(&rp, &a);
    let out = recovery_apply_by(&b, &f.home, &rp);
    assert_eq!(out["status"], json!("success"), "{out}");

    let again = recovery_plan_by(&b, &f.home);
    assert!(
        host(&again, "claude")["restore_backup"].is_null(),
        "{again}"
    );
    let c = put_build(&f.home, "c");
    let from_c = recovery_plan_by(&c, &f.home);
    assert!(
        host(&from_c, "claude")["restore_backup"].is_null(),
        "{from_c}"
    );

    // The restored folder is unmanaged again, and the two closed receipts do not make a third
    // build's next update ambiguous.
    let p = plan_by(&c, &f.home, &f.stage);
    let h = host(&p, "claude");
    assert!(h["backup_then_replace"].is_object(), "{p}");
    assert!(
        !h["blocked_reason"].to_string().contains("same second"),
        "{p}"
    );
    assert_eq!(snapshot(&f.home.join(ROOTS[0].1)), original);
}

// Characterization, not an acceptance criterion: an apply by the new build that finds nothing
// to write keeps the receipt's entries, so the offer stays.
#[test]
fn prior_install_offer_stays_after_a_no_write_apply() {
    let (f, _, a, b, _) = upgraded(true);
    sleep(Duration::from_millis(1100));
    let (_, r) = ordinary_apply(&b, &f.home, &f.stage);
    assert_eq!(host(&r, "claude")["status"], json!("unchanged"), "{r}");
    assert_offers_prior(&recovery_plan_by(&b, &f.home), &a);
}

// Characterization, not an acceptance criterion, and a known limit of the approved REQ-2: the
// offer counts the earlier receipt's entries and the entries of this installation's last apply
// only, so once the new build writes a second time the files its first apply wrote belong to
// neither list and the offer is no longer shown (`user_changed`, which shows no row).
#[test]
fn prior_install_offer_ends_after_a_second_writing_apply() {
    let (f, _, _, b, _) = upgraded(true);
    sleep(Duration::from_millis(1100));
    put_file(&f.stage, "workflows/update.md", "# UPDATE\nrust v3\n");
    let (_, r) = ordinary_apply(&b, &f.home, &f.stage);
    assert_eq!(host(&r, "claude")["status"], json!("applied"), "{r}");
    let rp = recovery_plan_by(&b, &f.home);
    assert!(host(&rp, "claude")["restore_backup"].is_null(), "{rp}");
}

// Guards of the candidate choice beyond SPEC-20261127-recovery-prior-install AC-1..7: an earlier
// receipt that is still pending, or that names another installation, lends no offer.
#[test]
fn prior_install_offer_is_not_taken_from_an_unsettled_receipt() {
    let (f, _, a, b, _) = upgraded(false);
    let a_id = id_of(&a);

    set_receipt_field(&f.home, &a_id, "status", "pending");
    let rp = recovery_plan_by(&b, &f.home);
    assert!(
        host(&rp, "claude")["restore_backup"].is_null(),
        "an earlier receipt that is pending: {rp}"
    );
    set_receipt_field(&f.home, &a_id, "status", "committed");

    set_receipt_field(&f.home, &a_id, "install_id", &sha256_hex(b"fixture-other"));
    let rp = recovery_plan_by(&b, &f.home);
    assert!(
        host(&rp, "claude")["restore_backup"].is_null(),
        "a receipt that names another installation: {rp}"
    );
    set_receipt_field(&f.home, &a_id, "install_id", &a_id);

    // Control: with both fields as they were the offer is there.
    assert_offers_prior(&recovery_plan_by(&b, &f.home), &a);
}
