//! UPDATE fixtures (04-B #1999). New-file discipline: existing test files are untouched.
//! Packages and projections are built at RUNTIME inside temporary directories — a
//! committed fixture directory would be dropped by the repo `.gitignore` (which excludes
//! `.claude/` and `.moai/` at every depth) and tests never touch the real user home.

#![cfg(unix)]

use moai_cli::tree_digest;
use serde_json::{json, Value};
use std::{
    fs,
    io::Write as _,
    path::{Path, PathBuf},
    process::{Command, Stdio},
};
use tempfile::TempDir;

const EXE: &str = env!("CARGO_BIN_EXE_moai-cli");

fn write_pkg(dir: &Path, files: &[(&str, &str)]) {
    for (rel, body) in files {
        let p = dir.join(rel);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        fs::write(p, body).unwrap();
    }
}

fn skill_md(body: &str) -> String {
    format!("---\nname: moai\ndescription: fixture package\n---\n# MoAI\n{body}\n")
}

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
        ("references/routes.json", "{\"routes\":1}\n".into()), // unchanged template
        ("workflows/update.md", "# UPDATE\nv2\n".into()),
        ("workflows/new.md", "# NEW\nadded in v2\n".into()),
    ]
}

fn stage_v3() -> Vec<(&'static str, String)> {
    // Drops both workflows files (remove_blocked) and keeps routes.json byte-identical
    // to v1/v2 (preserved), changing only SKILL.md.
    vec![
        ("SKILL.md", skill_md("body v3")),
        ("references/routes.json", "{\"routes\":1}\n".into()),
    ]
}

fn setup() -> (TempDir, PathBuf, PathBuf, PathBuf, PathBuf) {
    let root = TempDir::new().unwrap();
    let rootp = root.path().to_path_buf();
    let home = rootp.join("home");
    fs::create_dir_all(&home).unwrap();
    // Fake Go-generation state in the same home: update must never touch it.
    fs::create_dir_all(home.join(".moai/cache")).unwrap();
    fs::write(home.join(".moai/config.yaml"), "generation: go\n").unwrap();
    fs::write(home.join(".moai/cache/blob"), b"go state bytes").unwrap();
    let stage1 = rootp.join("stage-v1");
    let stage2 = rootp.join("stage-v2");
    let stage3 = rootp.join("stage-v3");
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
    write_pkg(
        &stage3,
        &stage_v3()
            .iter()
            .map(|(a, b)| (*a, b.as_str()))
            .collect::<Vec<_>>(),
    );
    (root, home, stage1, stage2, stage3)
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

fn envelope(out: &std::process::Output) -> Value {
    serde_json::from_slice(&out.stdout).expect("envelope JSON on stdout")
}

fn plan_via_cli(home: &Path, stage: &Path) -> (Value, std::process::Output) {
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
    let result = ok_result(&out);
    (result.clone(), out)
}

fn apply_via_cli(
    home: &Path,
    stage: &Path,
    plan: &Value,
    digest: &str,
    permission: &str,
) -> std::process::Output {
    run_cli(
        &[
            "update-apply",
            home.to_str().unwrap(),
            stage.to_str().unwrap(),
            "--plan-digest",
            digest,
            "--permission-reference",
            permission,
        ],
        Some(plan),
    )
}

fn apply_ok(home: &Path, stage: &Path, plan: &Value, digest: &str) -> Value {
    let out = apply_via_cli(home, stage, plan, digest, "fixture-approved");
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

fn plan_action(plan: &Value, host_name: &str, path: &str) -> String {
    plan["hosts"]
        .as_array()
        .unwrap()
        .iter()
        .find(|h| h["host"] == json!(host_name))
        .unwrap()["files"]
        .as_array()
        .unwrap()
        .iter()
        .find(|f| f["path"] == json!(path))
        .map(|f| f["action"].as_str().unwrap().to_string())
        .unwrap_or_else(|| format!("missing {host_name}/{path}"))
}

fn go_digest(home: &Path) -> String {
    tree_digest(&home.join(".moai"), &[]).unwrap()
}

fn projection_digest(home: &Path, host: &str) -> String {
    let dir = if host == "claude" {
        home.join(".claude/skills/moai")
    } else {
        home.join(".agents/skills/moai")
    };
    moai_cli::package_digest(&dir).unwrap()
}

fn receipts_dir(home: &Path) -> PathBuf {
    let installations = home.join(".moai-cli/installations");
    let mut entries: Vec<PathBuf> = fs::read_dir(&installations)
        .unwrap()
        .map(|e| e.unwrap().path())
        .filter(|p| p.is_dir())
        .collect();
    assert_eq!(entries.len(), 1, "one installation id directory");
    entries.pop().unwrap().join("hosts/moai")
}

fn receipt_value(home: &Path, host: &str) -> Value {
    let path = receipts_dir(home).join(format!("{host}.json"));
    serde_json::from_slice(&fs::read(&path).unwrap()).unwrap()
}

/// AC-04B-01: `update-plan` computes per-file three-way actions over BOTH hosts and
/// writes nothing (tree digest identical); the fake Go state is untouched.
#[test]
fn update_plan_is_read_only_and_covers_all_hosts() {
    let (_root, home, stage1, _s2, _s3) = setup();
    // Pre-existing projections: claude has one user-modified file and the rest of v1;
    // codex matches v1 byte for byte. No receipts exist yet, so the user-modified
    // file (baseline=None, current≠new) plans as a conflict.
    let claude = home.join(".claude/skills/moai");
    let codex = home.join(".agents/skills/moai");
    let v1: Vec<(&str, String)> = stage_v1();
    write_pkg(&claude, &[("SKILL.md", &skill_md("user edited locally"))]);
    write_pkg(
        &claude,
        &v1.iter()
            .filter(|(a, _)| *a != "SKILL.md")
            .map(|(a, b)| (*a, b.as_str()))
            .collect::<Vec<_>>(),
    );
    write_pkg(
        &codex,
        &v1.iter().map(|(a, b)| (*a, b.as_str())).collect::<Vec<_>>(),
    );

    let before = tree_digest(home.as_path(), &[]).unwrap();
    let (plan, out) = plan_via_cli(&home, &stage1);
    assert_eq!(envelope(&out)["operation"], json!("update.plan"));
    assert_eq!(envelope(&out)["protocol_version"], json!(1));

    assert_eq!(plan["hosts"].as_array().unwrap().len(), 2);
    assert_eq!(plan["scope"], json!(["claude", "codex"]));
    assert_eq!(plan["protocol_range"], json!({"min": 1, "max": 1}));
    assert_eq!(plan["source_version"], json!(null));
    assert!(!plan["plan_digest"].as_str().unwrap().is_empty());

    // claude: user-edited SKILL.md with no receipt baseline is a conflict; the
    // byte-identical routes.json is unchanged.
    assert_eq!(plan_action(&plan, "claude", "SKILL.md"), "conflict");
    assert_eq!(
        plan_action(&plan, "claude", "references/routes.json"),
        "unchanged"
    );
    // codex matches v1 byte for byte: unchanged everywhere.
    assert_eq!(plan_action(&plan, "codex", "SKILL.md"), "unchanged");
    assert_eq!(
        plan_action(&plan, "codex", "workflows/update.md"),
        "unchanged"
    );
    // Nothing to write anywhere: no expected hashes, no rollback manifest.
    assert_eq!(plan["expected_file_hashes"], json!([]));
    assert_eq!(plan["rollback_manifest"], json!([]));

    assert_eq!(
        tree_digest(home.as_path(), &[]).unwrap(),
        before,
        "update-plan must be read-only"
    );
}

/// AC-04B-02: applying to both hosts makes each projection's package digest equal the
/// staged package digest and leaves a committed receipt per host.
#[test]
fn update_apply_both_hosts_match_package_digest() {
    let (_root, home, stage1, _s2, _s3) = setup();
    let go_before = go_digest(&home);
    let (plan, _) = plan_via_cli(&home, &stage1);
    let digest = plan["plan_digest"].as_str().unwrap().to_string();
    let result = apply_ok(&home, &stage1, &plan, &digest);

    assert_eq!(result["status"], json!("applied"));
    assert_eq!(result["package_digest"], plan["package_digest"]);
    assert_eq!(result["host_loaded"], json!(false));
    assert_eq!(result["execution_verified"], json!("not_run"));
    for name in ["claude", "codex"] {
        let h = host(&result, name);
        assert_eq!(h["status"], json!("applied"));
        assert_eq!(h["receipt_status"], json!("committed"));
        assert!(h["files"]
            .as_array()
            .unwrap()
            .iter()
            .all(|f| f["published"] == json!(true)));
        assert_eq!(
            projection_digest(&home, name),
            moai_cli::package_digest(&stage1).unwrap(),
            "{name} projection digest must match the staged package"
        );
        // Independent copies, not symlinks (REQ-04B-08).
        let skill = if name == "claude" {
            home.join(".claude/skills/moai/SKILL.md")
        } else {
            home.join(".agents/skills/moai/SKILL.md")
        };
        assert!(fs::symlink_metadata(&skill).unwrap().is_file());
    }
    // reload_required with its documented basis (REQ-04B-06).
    let claude = host(&result, "claude");
    assert_eq!(claude["reload_required"], json!("yes"));
    assert!(claude["reload_basis"]
        .as_str()
        .unwrap()
        .contains("/reload-skills"));
    let codex = host(&result, "codex");
    assert_eq!(codex["reload_required"], json!("unknown"));
    assert!(codex["reload_basis"].as_str().unwrap().contains("restart"));

    // Committed receipts per host under the install id.
    let dir = receipts_dir(&home);
    for name in ["claude", "codex"] {
        let receipt: Value =
            serde_json::from_slice(&fs::read(dir.join(format!("{name}.json"))).unwrap()).unwrap();
        assert_eq!(receipt["status"], json!("committed"));
        assert_eq!(receipt["operation"], json!("update.apply"));
        assert_eq!(receipt["package_digest"], plan["package_digest"]);
        assert!(receipt["owned"].as_array().unwrap().len() >= 3);
    }
    assert_eq!(go_digest(&home), go_before, "Go state untouched");
}

/// AC-04B-03: a user edit is preserved per file — conflict when the template also
/// changed, preserved when it did not; user-added files are never touched; a file the
/// new package drops is remove_blocked, never deleted. Other files and hosts apply.
#[test]
fn update_preserves_user_edits_and_reports_conflict_per_file() {
    let (_root, home, stage1, stage2, stage3) = setup();
    // First apply establishes receipts (the baseline).
    let (p1, _) = plan_via_cli(&home, &stage1);
    apply_ok(&home, &stage1, &p1, p1["plan_digest"].as_str().unwrap());

    // User edits one projected file and adds a private file on the claude host.
    let update_md = home.join(".claude/skills/moai/workflows/update.md");
    fs::write(&update_md, "# UPDATE\nv1\nUSER EDIT\n").unwrap();
    fs::write(
        home.join(".claude/skills/moai/references/extra.md"),
        "mine\n",
    )
    .unwrap();

    let (p2, _) = plan_via_cli(&home, &stage2);
    assert_eq!(plan_action(&p2, "claude", "SKILL.md"), "update");
    assert_eq!(
        plan_action(&p2, "claude", "workflows/update.md"),
        "conflict"
    );
    assert_eq!(
        plan_action(&p2, "claude", "references/routes.json"),
        "unchanged"
    );
    assert_eq!(
        plan_action(&p2, "claude", "references/extra.md"),
        "preserved"
    );
    assert_eq!(plan_action(&p2, "claude", "workflows/new.md"), "create");
    assert_eq!(plan_action(&p2, "codex", "workflows/update.md"), "update");

    let r2 = apply_ok(&home, &stage2, &p2, p2["plan_digest"].as_str().unwrap());
    assert_eq!(r2["status"], json!("applied"));
    let claude_update_md = fs::read_to_string(&update_md).unwrap();
    assert!(claude_update_md.contains("USER EDIT"), "user edit kept");
    assert!(
        !claude_update_md.contains("v2"),
        "conflicted file not overwritten"
    );
    assert_eq!(
        fs::read_to_string(home.join(".claude/skills/moai/SKILL.md")).unwrap(),
        stage_v2()[0].1,
        "non-conflicted file applied"
    );
    assert_eq!(
        fs::read_to_string(home.join(".claude/skills/moai/references/extra.md")).unwrap(),
        "mine\n",
        "user file preserved"
    );
    assert_eq!(
        fs::read_to_string(home.join(".agents/skills/moai/workflows/update.md")).unwrap(),
        "# UPDATE\nv2\n",
        "other host applied"
    );

    // Template unchanged + user edited: preserved (routes.json user edit, v3 keeps it).
    fs::write(
        home.join(".claude/skills/moai/references/routes.json"),
        "{\"routes\": mine}\n",
    )
    .unwrap();
    let (p3, _) = plan_via_cli(&home, &stage3);
    assert_eq!(
        plan_action(&p3, "claude", "references/routes.json"),
        "preserved"
    );
    assert_eq!(
        plan_action(&p3, "claude", "workflows/update.md"),
        "remove_blocked"
    );
    assert_eq!(
        plan_action(&p3, "claude", "workflows/new.md"),
        "remove_blocked"
    );
    assert_eq!(plan_action(&p3, "claude", "SKILL.md"), "update");
    let r3 = apply_ok(&home, &stage3, &p3, p3["plan_digest"].as_str().unwrap());
    assert_eq!(r3["status"], json!("applied"));
    assert_eq!(
        fs::read_to_string(home.join(".claude/skills/moai/references/routes.json")).unwrap(),
        "{\"routes\": mine}\n",
        "template-unchanged user edit preserved"
    );
    assert!(
        home.join(".claude/skills/moai/workflows/update.md")
            .exists(),
        "dropped file kept"
    );
    assert!(
        home.join(".claude/skills/moai/workflows/new.md").exists(),
        "dropped file kept"
    );
    assert_eq!(
        fs::read_to_string(home.join(".claude/skills/moai/SKILL.md")).unwrap(),
        stage_v3()[0].1,
    );
}

/// AC-04B-04: re-running the same package reports every host unchanged with a
/// byte-identical projection tree.
#[test]
fn update_rerun_is_unchanged_diff_zero() {
    let (_root, home, stage1, _s2, _s3) = setup();
    let (p1, _) = plan_via_cli(&home, &stage1);
    apply_ok(&home, &stage1, &p1, p1["plan_digest"].as_str().unwrap());

    let claude_before = tree_digest(&home.join(".claude"), &[]).unwrap();
    let codex_before = tree_digest(&home.join(".agents"), &[]).unwrap();
    let (p2, _) = plan_via_cli(&home, &stage1);
    for name in ["claude", "codex"] {
        assert_eq!(plan_action(&p2, name, "SKILL.md"), "unchanged");
    }
    assert_eq!(p2["expected_file_hashes"], json!([]));
    let r2 = apply_ok(&home, &stage1, &p2, p2["plan_digest"].as_str().unwrap());
    assert_eq!(r2["status"], json!("unchanged"));
    for name in ["claude", "codex"] {
        let h = host(&r2, name);
        assert_eq!(h["status"], json!("unchanged"));
        assert_eq!(h["reload_required"], json!("no"));
    }
    assert_eq!(
        tree_digest(&home.join(".claude"), &[]).unwrap(),
        claude_before
    );
    assert_eq!(
        tree_digest(&home.join(".agents"), &[]).unwrap(),
        codex_before
    );
}

/// AC-04B-05: stale digests, tampered plans, replays under another HOME and symlinked
/// projections are refused with zero writes.
#[test]
fn update_apply_rejects_stale_tampered_replayed_and_symlinked() {
    let (_root, home, stage1, stage2, _s3) = setup();
    // Establish a committed state first.
    let (p1, _) = plan_via_cli(&home, &stage1);
    apply_ok(&home, &stage1, &p1, p1["plan_digest"].as_str().unwrap());
    let before = tree_digest(&home.join(".claude"), &[]).unwrap();

    // (a) wrong plan digest.
    let (pv2, _) = plan_via_cli(&home, &stage2);
    let out = apply_via_cli(&home, &stage2, &pv2, "sha256:deadbeef", "fixture-approved");
    assert!(!out.status.success());
    assert_eq!(envelope(&out)["error"]["kind"], json!("stale_plan"));
    assert_eq!(tree_digest(&home.join(".claude"), &[]).unwrap(), before);

    // (b) tampered plan body.
    let mut tampered = pv2.clone();
    tampered["target_version"] = json!("9.9.9");
    let out = apply_via_cli(
        &home,
        &stage2,
        &tampered,
        pv2["plan_digest"].as_str().unwrap(),
        "fixture-approved",
    );
    assert!(!out.status.success());
    assert_eq!(envelope(&out)["error"]["kind"], json!("stale_plan"));
    assert_eq!(tree_digest(&home.join(".claude"), &[]).unwrap(), before);

    // (c) missing permission reference.
    let out = apply_via_cli(
        &home,
        &stage2,
        &pv2,
        pv2["plan_digest"].as_str().unwrap(),
        "placeholder",
    );
    assert!(!out.status.success());
    assert_eq!(
        envelope(&out)["error"]["kind"],
        json!("permission_reference_missing")
    );

    // (d) replay under another HOME.
    let other = TempDir::new().unwrap();
    let other_home = other.path().join("home");
    fs::create_dir_all(&other_home).unwrap();
    let out = apply_via_cli(
        &other_home,
        &stage2,
        &pv2,
        pv2["plan_digest"].as_str().unwrap(),
        "fixture-approved",
    );
    assert!(
        !out.status.success(),
        "a plan bound to one HOME cannot be replayed in another"
    );
    assert!(
        !other_home.join(".claude").exists(),
        "no writes in the replayed HOME"
    );

    // (e) symlinked projection: blocked, never followed; the other host still applies.
    let third = TempDir::new().unwrap();
    let home3 = third.path().join("home");
    fs::create_dir_all(home3.join(".claude")).unwrap();
    let outside = third.path().join("outside");
    fs::create_dir_all(&outside).unwrap();
    #[cfg(unix)]
    std::os::unix::fs::symlink(&outside, home3.join(".claude/skills")).unwrap();
    fs::write(outside.join("keep.txt"), b"do not touch").unwrap();
    let (p3, _) = plan_via_cli(&home3, &stage1);
    let claude = p3["hosts"]
        .as_array()
        .unwrap()
        .iter()
        .find(|h| h["host"] == json!("claude"))
        .unwrap();
    assert_eq!(claude["preflight"], json!("blocked"));
    assert!(claude["blocked_reason"]
        .as_str()
        .unwrap()
        .contains("symlink"));
    assert_eq!(
        p3["hosts"]
            .as_array()
            .unwrap()
            .iter()
            .find(|h| h["host"] == json!("codex"))
            .unwrap()["preflight"],
        json!("ok")
    );
    let r3 = apply_ok(&home3, &stage1, &p3, p3["plan_digest"].as_str().unwrap());
    assert_eq!(host(&r3, "claude")["status"], json!("blocked"));
    assert_eq!(host(&r3, "claude")["reload_required"], json!("no"));
    assert_eq!(
        fs::read(outside.join("keep.txt")).unwrap(),
        b"do not touch",
        "symlink target untouched"
    );
    assert_eq!(
        host(&r3, "codex")["status"],
        json!("applied"),
        "the other host proceeds"
    );
}

/// AC-04B-06: per-host results carry reload_required, host_loaded=false and
/// execution_verified=not_run; a mid-apply failure on one host leaves the whole run
/// `partial` with the failed receipt pending and the committed host untouched.
#[test]
fn update_result_reports_per_host_and_partial() {
    let (_root, home, stage1, stage2, _s3) = setup();
    let (p1, _) = plan_via_cli(&home, &stage1);
    apply_ok(&home, &stage1, &p1, p1["plan_digest"].as_str().unwrap());

    // Sabotage the codex projection for a deterministic mid-apply failure: the
    // `workflows` path becomes a plain file, so creating workflows/new.md fails inside
    // step (b) — after the pending receipt was written.
    let codex = home.join(".agents/skills/moai");
    fs::remove_dir_all(codex.join("workflows")).unwrap();
    fs::write(codex.join("workflows"), b"not a directory\n").unwrap();

    let (p2, _) = plan_via_cli(&home, &stage2);
    let r2 = apply_ok(&home, &stage2, &p2, p2["plan_digest"].as_str().unwrap());
    assert_eq!(
        r2["status"],
        json!("partial"),
        "one failed host makes the run partial"
    );
    assert_eq!(r2["host_loaded"], json!(false));
    assert_eq!(r2["execution_verified"], json!("not_run"));

    let claude = host(&r2, "claude");
    assert_eq!(claude["status"], json!("applied"));
    assert_eq!(
        claude["receipt_status"],
        json!("committed"),
        "committed host not rolled back"
    );
    let codex = host(&r2, "codex");
    assert_eq!(codex["status"], json!("failed"));
    assert_eq!(
        codex["receipt_status"],
        json!("pending"),
        "failed host keeps the pending receipt"
    );
    assert_eq!(codex["reload_required"], json!("unknown"));
    let errored = codex["files"]
        .as_array()
        .unwrap()
        .iter()
        .find(|f| f["error"] != json!(null))
        .expect("the failing file carries the error");
    assert_eq!(errored["published"], json!(false));

    let receipt = receipt_value(&home, "codex");
    assert_eq!(receipt["status"], json!("pending"));
    // The pending receipt keeps the PRE-apply ownership manifest, so the partially
    // published state surfaces as a per-file conflict on the next plan, never as a
    // silent overwrite.
    let (p3, _) = plan_via_cli(&home, &stage2);
    let codex_plan = p3["hosts"]
        .as_array()
        .unwrap()
        .iter()
        .find(|h| h["host"] == json!("codex"))
        .unwrap();
    // REQ-04C-04 (04-C #2000): the pending host takes no new apply until recovery-plan /
    // recovery-apply resolves it; it is blocked, never silently overwritten.
    assert_eq!(codex_plan["preflight"], json!("blocked"));
    assert_eq!(
        codex_plan["blocked_reason"],
        json!("host receipt is pending; run recovery-plan first")
    );
    // claude, committed, plans cleanly against the new stage.
    assert_eq!(plan_action(&p3, "claude", "SKILL.md"), "unchanged");
}

/// AC-04B-07 fixture half: the fake Go-generation `.moai/` tree in the same home is
/// byte-identical across a full plan+apply cycle. (The machine-level wrap around the
/// suite measures the real `~/.moai` and is recorded in the owned report.)
#[test]
fn update_fixture_preserves_fake_go_state() {
    let (_root, home, stage1, _s2, _s3) = setup();
    let before = go_digest(&home);
    let (p1, _) = plan_via_cli(&home, &stage1);
    apply_ok(&home, &stage1, &p1, p1["plan_digest"].as_str().unwrap());
    let (p2, _) = plan_via_cli(&home, &stage1);
    apply_ok(&home, &stage1, &p2, p2["plan_digest"].as_str().unwrap());
    assert_eq!(go_digest(&home), before);
}

/// #2286 (decision row 22): the skill package is named `moai`, the same folder name an
/// earlier-generation (Go-era) MoAI skill uses. A Go-type `skills/moai` the install
/// receipt does not own blocks the host in update-plan (never an automatic replace),
/// and update-apply leaves both Go trees byte-identical (tree digests unchanged).
#[test]
fn go_type_moai_blocks_the_install_and_stays_untouched() {
    let (_root, home, stage1, _s2, _s3) = setup();
    for host_dir in [".claude/skills", ".agents/skills"] {
        write_pkg(
            &home.join(host_dir).join("moai"),
            &[("SKILL.md", "---\nname: moai\n---\nGo body\n")],
        );
    }
    let go_before: Vec<String> = [".claude/skills/moai", ".agents/skills/moai"]
        .iter()
        .map(|d| tree_digest(&home.join(d), &[]).unwrap())
        .collect();

    let (plan, _) = plan_via_cli(&home, &stage1);
    for host_name in ["claude", "codex"] {
        let h = host(&plan, host_name);
        assert_eq!(h["preflight"], json!("blocked"), "{host_name}: {h}");
        assert!(
            h["blocked_reason"]
                .as_str()
                .unwrap()
                .contains("unmanaged existing package"),
            "{h}"
        );
    }

    let v = apply_ok(&home, &stage1, &plan, plan["plan_digest"].as_str().unwrap());
    assert_eq!(v["status"], json!("blocked"), "{v}");
    let go_after: Vec<String> = [".claude/skills/moai", ".agents/skills/moai"]
        .iter()
        .map(|d| tree_digest(&home.join(d), &[]).unwrap())
        .collect();
    assert_eq!(
        go_after, go_before,
        "the Go-type moai trees must be byte-identical"
    );
}

/// Moves a fresh install into the rc-era layout (#2286 before the `moai` rename): the
/// projection roots at `skills/moai-cli` and the host receipts, marked pending, at
/// `hosts/moai-cli/`. Returns each moved root and receipt with its digest/bytes.
fn into_rc_era_layout(home: &Path, stage: &Path) -> Vec<(PathBuf, String, PathBuf, Vec<u8>)> {
    let (initial, _) = plan_via_cli(home, stage);
    apply_ok(
        home,
        stage,
        &initial,
        initial["plan_digest"].as_str().unwrap(),
    );
    let current_receipts = receipts_dir(home);
    let legacy_receipts = current_receipts.parent().unwrap().join("moai-cli");
    fs::create_dir_all(&legacy_receipts).unwrap();
    let mut legacy_state = Vec::new();
    for (name, parent) in [("claude", ".claude/skills"), ("codex", ".agents/skills")] {
        let old_root = home.join(parent).join("moai-cli");
        fs::rename(home.join(parent).join("moai"), &old_root).unwrap();
        let current = current_receipts.join(format!("{name}.json"));
        let legacy = legacy_receipts.join(format!("{name}.json"));
        let mut receipt: Value = serde_json::from_slice(&fs::read(&current).unwrap()).unwrap();
        // Pending legacy work must not be recovered against the new, absent root.
        receipt["status"] = json!("pending");
        fs::write(&legacy, serde_json::to_vec(&receipt).unwrap()).unwrap();
        fs::remove_file(&current).unwrap();
        let old_digest = tree_digest(&old_root, &[]).unwrap();
        let legacy_bytes = fs::read(&legacy).unwrap();
        legacy_state.push((old_root, old_digest, legacy, legacy_bytes));
    }
    legacy_state
}

/// Receipts written for the earlier `skills/moai-cli` root (namespace `hosts/moai-cli/`)
/// never supply a baseline or recovery input for the `skills/moai` root.
#[test]
fn legacy_receipts_cannot_supply_the_new_skill_baseline_or_recovery() {
    let (_root, home, stage1, stage2, _s3) = setup();
    let go_before = go_digest(&home);
    let legacy_state = into_rc_era_layout(&home, &stage1);
    let recovery = ok_result(&run_cli(&["recovery-plan", home.to_str().unwrap()], None));
    for name in ["claude", "codex"] {
        assert_eq!(host(&recovery, name)["receipt_status"], "none");
    }
    let (plan, _) = plan_via_cli(&home, &stage2);
    for name in ["claude", "codex"] {
        assert_eq!(host(&plan, name)["blocked_reason"], Value::Null);
        for (path, _) in stage_v2() {
            assert_eq!(plan_action(&plan, name, path), "create");
        }
    }
    apply_ok(&home, &stage2, &plan, plan["plan_digest"].as_str().unwrap());
    for (name, parent) in [("claude", ".claude/skills"), ("codex", ".agents/skills")] {
        for (path, bytes) in stage_v2() {
            assert_eq!(
                fs::read(home.join(parent).join("moai").join(path)).unwrap(),
                bytes.as_bytes()
            );
        }
        assert_eq!(receipt_value(&home, name)["status"], "committed");
    }
    for (root, digest, receipt, bytes) in legacy_state {
        assert_eq!(tree_digest(&root, &[]).unwrap(), digest);
        assert_eq!(fs::read(receipt).unwrap(), bytes);
    }
    assert_eq!(go_digest(&home), go_before);
}

/// A legacy `hosts/moai-cli/` receipt that lists `SKILL.md` grants no ownership over a
/// Go-era `skills/moai`: the host is blocked as an unmanaged existing package.
#[test]
fn legacy_receipts_do_not_own_a_go_era_moai_tree() {
    let (_root, home, stage1, stage2, _s3) = setup();
    let legacy_state = into_rc_era_layout(&home, &stage1);
    for parent in [".claude/skills", ".agents/skills"] {
        write_pkg(
            &home.join(parent).join("moai"),
            &[(
                "SKILL.md",
                "---\nname: moai\ndescription: go\n---\nGo body\n",
            )],
        );
    }
    let (plan, _) = plan_via_cli(&home, &stage2);
    for name in ["claude", "codex"] {
        let h = host(&plan, name);
        assert_eq!(h["preflight"], json!("blocked"), "{name}: {h}");
        assert!(h["blocked_reason"]
            .as_str()
            .unwrap()
            .contains("unmanaged existing package"));
    }
    for (root, digest, receipt, bytes) in legacy_state {
        assert_eq!(tree_digest(&root, &[]).unwrap(), digest);
        assert_eq!(fs::read(receipt).unwrap(), bytes);
    }
}
