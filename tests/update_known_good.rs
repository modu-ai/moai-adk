//! Known-good copies + `version_before` (04-B #2089). `update-apply` preserves the
//! exact bytes of every overwritten file under
//! `~/.moai-cli/installations/<id>/hosts/moai/<host>/known-good/<sha256-hex>` before the
//! apply transaction starts (REQ-04C-03 recovery input), and each receipt carries the
//! previous receipt's `package_version` as `version_before` (REQ-04C-06). New-file
//! discipline: existing test files are untouched. Packages are built at RUNTIME inside
//! temporary directories and tests never touch the real user home.

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

fn skill_md(body: &str) -> String {
    format!("---\nname: moai\ndescription: fixture package\n---\n# MoAI\n{body}\n")
}

fn write_pkg(dir: &Path, files: &[(&str, &str)]) {
    for (rel, body) in files {
        let p = dir.join(rel);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        fs::write(p, body).unwrap();
    }
}

/// v1: three files, all `create` on a fresh home. v2: changes SKILL.md and
/// workflows/update.md (`update`), keeps references/routes.json identical
/// (`preserved`), and adds workflows/new.md (`create`).
fn stage_v1() -> Vec<(&'static str, String)> {
    vec![
        ("SKILL.md", skill_md("body v1")),
        ("references/routes.json", "{\"routes\":1}\n".into()),
        ("workflows/update.md", "# UPDATE\nv1\n".into()),
    ]
}

fn stage_v2() -> Vec<(&'static str, String)> {
    vec![
        ("SKILL.md", skill_md("body v2")),
        ("references/routes.json", "{\"routes\":1}\n".into()),
        ("workflows/update.md", "# UPDATE\nv2\n".into()),
        ("workflows/new.md", "# NEW\nadded in v2\n".into()),
    ]
}

fn setup() -> (TempDir, PathBuf, PathBuf, PathBuf) {
    let root = TempDir::new().unwrap();
    let rootp = root.path().to_path_buf();
    let home = rootp.join("home");
    fs::create_dir_all(&home).unwrap();
    // Fake Go-generation state in the same home: update must never touch it.
    fs::create_dir_all(home.join(".moai/cache")).unwrap();
    fs::write(home.join(".moai/config.yaml"), "generation: go\n").unwrap();
    let stage1 = rootp.join("stage-v1");
    let stage2 = rootp.join("stage-v2");
    write_pkg(
        &stage1,
        &stage_v1()
            .iter()
            .map(|(a, b)| (*a, b.as_str()))
            .collect::<Vec<_>>(),
    );
    write_pkg(
        &stage2,
        &stage_v2()
            .iter()
            .map(|(a, b)| (*a, b.as_str()))
            .collect::<Vec<_>>(),
    );
    (root, home, stage1, stage2)
}

fn run_cli(args: &[&str], stdin: Option<&Value>) -> std::process::Output {
    let mut child = Command::new(EXE)
        .args(args)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    let pipe = child.stdin.take().unwrap();
    match stdin {
        Some(v) => {
            let mut pipe = pipe;
            let data = v.to_string();
            match pipe.write_all(data.as_bytes()) {
                Ok(()) => {}
                Err(e) if e.kind() == std::io::ErrorKind::BrokenPipe => {}
                Err(e) => panic!("stdin write failed: {e}"),
            }
        }
        None => drop(pipe),
    }
    child.wait_with_output().unwrap()
}

fn ok_result(out: &std::process::Output) -> Value {
    let v: Value = serde_json::from_slice(&out.stdout).expect("envelope JSON on stdout");
    assert_eq!(
        v["ok"],
        json!(true),
        "{}",
        String::from_utf8_lossy(&out.stdout)
    );
    v["result"].clone()
}

fn plan_via_cli(home: &Path, stage: &Path) -> Value {
    let out = run_cli(
        &[
            "update-plan",
            home.to_str().unwrap(),
            stage.to_str().unwrap(),
        ],
        None,
    );
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    ok_result(&out)
}

fn apply_via_cli(home: &Path, stage: &Path, plan: &Value, digest: &str) -> std::process::Output {
    run_cli(
        &[
            "update-apply",
            home.to_str().unwrap(),
            stage.to_str().unwrap(),
            "--plan-digest",
            digest,
            "--permission-reference",
            "fixture-approved",
        ],
        Some(plan),
    )
}

fn apply_ok(home: &Path, stage: &Path, plan: &Value, digest: &str) -> Value {
    let out = apply_via_cli(home, stage, plan, digest);
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    ok_result(&out)
}

fn host<'a>(result: &'a Value, name: &str) -> &'a Value {
    result["hosts"]
        .as_array()
        .unwrap()
        .iter()
        .find(|h| h["host"] == json!(name))
        .expect("host entry")
}

/// The single fixture installation's committed host receipt.
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

/// Bare-hex sha256 of `bytes` — the known-good file name form (REQ-04C-03 `<sha256>`).
fn file_hex(bytes: &[u8]) -> String {
    digest(bytes).trim_start_matches("sha256:").to_string()
}

#[test]
fn known_good_copy_holds_pre_overwrite_bytes() {
    let (_root, home, stage1, stage2) = setup();
    let v1_skill = stage_v1()[0].1.clone();
    let v1_update_md = stage_v1()[2].1.clone();

    let plan1 = plan_via_cli(&home, &stage1);
    apply_ok(
        &home,
        &stage1,
        &plan1,
        plan1["plan_digest"].as_str().unwrap(),
    );
    let install_id = receipt_json(&home)["install_id"]
        .as_str()
        .unwrap()
        .to_string();
    // AC-2 (first half): a fresh install has no before bytes, so no copies exist.
    assert!(!known_good_dir(&home, &install_id).exists());

    let plan2 = plan_via_cli(&home, &stage2);
    apply_ok(
        &home,
        &stage2,
        &plan2,
        plan2["plan_digest"].as_str().unwrap(),
    );

    // AC-1: each overwritten file has a copy whose bytes equal the pre-overwrite file.
    let dir = known_good_dir(&home, &install_id);
    for body in [&v1_skill, &v1_update_md] {
        let hex = file_hex(body.as_bytes());
        let copy = dir.join(&hex);
        assert!(copy.is_file(), "missing known-good copy {hex}");
        assert_eq!(
            fs::read(&copy).unwrap(),
            body.as_bytes(),
            "copy bytes differ"
        );
    }
    // AC-2 (second half): exactly the two overwritten files — no copy for the
    // created workflows/new.md and none for the preserved routes.json.
    let count = fs::read_dir(&dir).unwrap().count();
    assert_eq!(count, 2, "unexpected extra known-good copies");
    assert!(
        !dir.join(file_hex(stage_v2()[3].1.as_bytes())).exists(),
        "a created file must not get a known-good copy"
    );
    // The copies hold the PRE-overwrite bytes: SKILL.md on disk is now v2.
    assert_eq!(
        fs::read(home.join(".claude/skills/moai/SKILL.md")).unwrap(),
        stage_v2()[0].1.as_bytes()
    );
}

#[test]
fn version_before_tracks_previous_receipt() {
    let (_root, home, stage1, stage2) = setup();

    let plan1 = plan_via_cli(&home, &stage1);
    apply_ok(
        &home,
        &stage1,
        &plan1,
        plan1["plan_digest"].as_str().unwrap(),
    );
    let r1 = receipt_json(&home);
    assert_eq!(r1["status"], json!("committed"));
    assert!(
        r1["version_before"].is_null(),
        "a first apply has no previous version"
    );

    let plan2 = plan_via_cli(&home, &stage2);
    apply_ok(
        &home,
        &stage2,
        &plan2,
        plan2["plan_digest"].as_str().unwrap(),
    );
    let r2 = receipt_json(&home);
    // AC-3: the second receipt's version_before is the first apply's package_version.
    assert_eq!(
        r2["version_before"], r1["package_version"],
        "version_before must preserve the previous receipt's package_version"
    );
    assert_eq!(r2["package_version"], r1["package_version"]);
}

#[test]
fn known_good_failure_starts_nothing_for_the_host() {
    let (_root, home, stage1, stage2) = setup();
    let v1_skill = stage_v1()[0].1.clone();

    let plan1 = plan_via_cli(&home, &stage1);
    apply_ok(
        &home,
        &stage1,
        &plan1,
        plan1["plan_digest"].as_str().unwrap(),
    );
    let install_id = receipt_json(&home)["install_id"]
        .as_str()
        .unwrap()
        .to_string();

    // Inject a copy-save failure: plant a DIRECTORY at the digest path the v2 apply
    // must write (SKILL.md is the first update-action file).
    let blocker = known_good_dir(&home, &install_id).join(file_hex(v1_skill.as_bytes()));
    fs::create_dir_all(&blocker).unwrap();

    let plan2 = plan_via_cli(&home, &stage2);
    let result = apply_ok(
        &home,
        &stage2,
        &plan2,
        plan2["plan_digest"].as_str().unwrap(),
    );

    // AC-4: the host is reported failed and none of its files changed.
    let h = host(&result, HOST);
    assert_eq!(h["status"], json!("failed"), "{h}");
    assert!(
        h["reload_basis"]
            .as_str()
            .unwrap()
            .contains("known-good copy failed"),
        "{}",
        h["reload_basis"]
    );
    assert_eq!(
        fs::read(home.join(".claude/skills/moai/SKILL.md")).unwrap(),
        v1_skill.as_bytes(),
        "the host must be untouched when the copy fails"
    );
    // The pre-existing committed receipt is untouched (still v1's committed one).
    let r = receipt_json(&home);
    assert_eq!(r["status"], json!("committed"));
    assert_eq!(r["operation"], json!("update.apply"));
    assert!(r["version_before"].is_null());
}

#[test]
fn stale_or_partial_copy_is_replaced_with_true_bytes() {
    let (_root, home, stage1, stage2) = setup();
    let v1_skill = stage_v1()[0].1.clone();

    let plan1 = plan_via_cli(&home, &stage1);
    apply_ok(
        &home,
        &stage1,
        &plan1,
        plan1["plan_digest"].as_str().unwrap(),
    );
    let install_id = receipt_json(&home)["install_id"]
        .as_str()
        .unwrap()
        .to_string();

    // Plant a stale copy — e.g. the remains of a failed earlier save — at the
    // digest path the v2 apply will reuse for SKILL.md.
    let stale = known_good_dir(&home, &install_id).join(file_hex(v1_skill.as_bytes()));
    fs::create_dir_all(stale.parent().unwrap()).unwrap();
    fs::write(&stale, b"partial bytes from a failed save\n").unwrap();

    let plan2 = plan_via_cli(&home, &stage2);
    apply_ok(
        &home,
        &stage2,
        &plan2,
        plan2["plan_digest"].as_str().unwrap(),
    );

    // The reuse path verified the name against the bytes and replaced the stale
    // copy with the true pre-overwrite bytes.
    assert_eq!(fs::read(&stale).unwrap(), v1_skill.as_bytes());
    assert_eq!(receipt_json(&home)["status"], json!("committed"));
}
