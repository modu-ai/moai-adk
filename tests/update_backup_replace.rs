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
//! or the backup intact and surface as pending; AC-6 both host locations.

#![cfg(unix)]

use serde_json::{json, Value};
use std::{
    collections::BTreeMap,
    fs,
    io::Write as _,
    os::unix::fs::PermissionsExt,
    path::{Path, PathBuf},
    process::{Command, Stdio},
};
use tempfile::TempDir;

const EXE: &str = env!("CARGO_BIN_EXE_moai-cli");
const ROOTS: [(&str, &str); 2] = [
    ("claude", ".claude/skills/moai"),
    ("codex", ".agents/skills/moai"),
];

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
        .env_remove("MOAI_CLI_TEST_FAULT")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped());
    if let Some(f) = fault {
        cmd.env("MOAI_CLI_TEST_FAULT", f);
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
    let dir = home.join(".moai-cli/backups");
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

/// AC-1 (+AC-6): the plan previews backup-then-replace for both hosts — what, where,
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
            json!(f.home.join(".moai-cli/backups").to_str().unwrap())
        );
        assert_eq!(b["replace_with"]["file_count"], json!(3));
        assert_eq!(b["replace_with"]["package_digest"], p["package_digest"]);
        let tok = b["confirm_token"].as_str().unwrap();
        assert!(tok.starts_with(&format!("backup-replace:{name}:")), "{tok}");
    }
    assert_eq!(originals(&f), before, "plan wrote nothing");
    assert!(
        !f.home.join(".moai-cli/backups").exists(),
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

/// AC-2 (second half) + AC-6: with the tokens, each host's original folder moves into
/// a backup whose MANIFEST and tree equal the original bytes and modes, and the
/// projection holds exactly the package.
#[test]
fn confirmed_apply_backs_up_and_replaces_both_hosts() {
    let f = fixture();
    let before = originals(&f);
    let p = plan(&f);
    let toks = [token(&p, "claude"), token(&p, "codex")];
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

/// Only the confirmed host is replaced; the other stays blocked and untouched.
#[test]
fn a_token_replaces_only_its_host() {
    let f = fixture();
    let before = originals(&f);
    let p = plan(&f);
    let r = result_of(&apply(&f, &p, &[&token(&p, "codex")], None));
    assert_eq!(host(&r, "claude")["status"], json!("blocked"));
    assert_eq!(host(&r, "codex")["status"], json!("replaced"));
    assert_eq!(
        snapshot(&f.home.join(ROOTS[0].1)),
        before[0],
        "claude untouched"
    );
    assert_eq!(backups(&f.home).len(), 1);
}

/// AC-3: recovery restores the original folder byte and mode identical and records the
/// use of the backup; the next plan offers the preview again.
#[test]
fn recovery_restores_the_original_folder() {
    let f = fixture();
    let before = originals(&f);
    let p = plan(&f);
    let toks = [token(&p, "claude"), token(&p, "codex")];
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

/// B1 (audit 6812528075): `~/.moai-cli/backups` is a symlink leading out of HOME.
#[test]
fn a_symlinked_backups_directory_is_refused() {
    assert_symlinked_state_path_is_refused(".moai-cli/backups");
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
