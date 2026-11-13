//! 04-C (#2000) UPDATE partial-failure recovery. Every case runs in a temporary HOME; the
//! pending state is built from a REAL v1 install (update-apply on both hosts) and then
//! either rewritten by hand to a chosen crash point (the end state a crash would leave:
//! files on disk + a pending receipt) or produced by 04-B's own sabotage pattern.
//! No injection hook is added to `apply_update`.

#![cfg(unix)]

mod common;
use moai_cli::{contract, tree_digest};
use serde_json::{json, Value};
use std::{
    fs,
    io::Write as _,
    path::{Path, PathBuf},
    process::{Command, Stdio},
};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;
const SKILL_V1: &str = "---\nname: moai\ndescription: fixture package\n---\n# MoAI\nbody v1\n";
const SKILL_V2: &str = "---\nname: moai\ndescription: fixture package\n---\n# MoAI\nbody v2\n";
const UPDATE_V1: &str = "# UPDATE\nv1\n";
const UPDATE_V2: &str = "# UPDATE\nv2\n";
const NEW_V2: &str = "# NEW\nadded in v2\n";

fn sha(b: &[u8]) -> String {
    moai_cli::digest(b)
}

fn write_pkg(dir: &Path, files: &[(&str, &str)]) {
    for (rel, body) in files {
        let p = dir.join(rel);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        fs::write(p, body).unwrap();
    }
}

fn run(args: &[&str], stdin: Option<&Value>) -> (bool, Value) {
    let mut child = Command::new(EXE)
        .args(args)
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

/// A HOME with both hosts committed at v1, and a v2 stage.
struct Fixture {
    _root: TempDir,
    home: PathBuf,
    stage2: PathBuf,
}

fn fixture() -> Fixture {
    let root = TempDir::new().unwrap();
    let home = root.path().join("home");
    fs::create_dir_all(home.join(".moai/cache")).unwrap();
    fs::write(home.join(".moai/config.yaml"), "generation: go\n").unwrap();
    let stage1 = root.path().join("stage-v1");
    let stage2 = root.path().join("stage-v2");
    write_pkg(
        &stage1,
        &[("SKILL.md", SKILL_V1), ("workflows/update.md", UPDATE_V1)],
    );
    write_pkg(
        &stage2,
        &[
            ("SKILL.md", SKILL_V2),
            ("workflows/update.md", UPDATE_V2),
            ("workflows/new.md", NEW_V2),
        ],
    );
    let (ok, v) = run(
        &[
            "update-plan",
            home.to_str().unwrap(),
            stage1.to_str().unwrap(),
        ],
        None,
    );
    assert!(ok, "{v}");
    let plan = v["result"].clone();
    let (ok, v) = run(
        &[
            "update-apply",
            home.to_str().unwrap(),
            stage1.to_str().unwrap(),
            "--plan-digest",
            plan["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "fixture-approved",
        ],
        Some(&plan),
    );
    assert!(ok, "{v}");
    assert_eq!(v["result"]["status"], json!("applied"), "{v}");
    Fixture {
        _root: root,
        home,
        stage2,
    }
}

impl Fixture {
    fn codex(&self) -> PathBuf {
        self.home.join(".agents/skills/moai")
    }
    fn claude(&self) -> PathBuf {
        self.home.join(".claude/skills/moai")
    }
    fn hosts_dir(&self) -> PathBuf {
        let inst = self.home.join(".moai-cli/installations");
        let dirs: Vec<PathBuf> = fs::read_dir(&inst)
            .unwrap()
            .map(|e| e.unwrap().path())
            .filter(|p| p.is_dir())
            .collect();
        assert_eq!(dirs.len(), 1);
        dirs[0].join("hosts/moai")
    }
    fn receipt(&self, host: &str) -> Value {
        serde_json::from_slice(&fs::read(self.hosts_dir().join(format!("{host}.json"))).unwrap())
            .unwrap()
    }
    /// The end state of a codex apply to v2 that crashed after `new.md` was created and
    /// `SKILL.md` replaced, before `update.md` and before the receipt commit: both files
    /// already carry v2 but the receipt still says pending (published left false for
    /// SKILL.md to show that classification ignores the flag).
    fn crash_codex(&self) {
        let root = self.codex();
        fs::write(root.join("workflows/new.md"), NEW_V2).unwrap();
        fs::write(root.join("SKILL.md"), SKILL_V2).unwrap();
        let mut r = self.receipt("codex");
        r["status"] = json!("pending");
        r["operation"] = json!("update.apply");
        r["finished_at"] = Value::Null;
        // What 04-B writes since #2091: the version this apply replaces.
        r["version_before"] = r["package_version"].clone();
        r["entries"] = json!([
            {"path": "workflows/new.md", "action": "create", "before_sha256": null,
             "after_sha256": sha(NEW_V2.as_bytes()), "published": true, "error": null},
            {"path": "SKILL.md", "action": "update", "before_sha256": sha(SKILL_V1.as_bytes()),
             "after_sha256": sha(SKILL_V2.as_bytes()), "published": false, "error": null},
            {"path": "workflows/update.md", "action": "update",
             "before_sha256": sha(UPDATE_V1.as_bytes()), "after_sha256": sha(UPDATE_V2.as_bytes()),
             "published": false, "error": null},
        ]);
        fs::write(
            self.hosts_dir().join("codex.json"),
            serde_json::to_vec_pretty(&r).unwrap(),
        )
        .unwrap();
    }
    /// Places the known-good copy of `bytes` for `host` (what 04-B stores at apply since #2091).
    fn known_good(&self, host: &str, bytes: &str) {
        let s = sha(bytes.as_bytes());
        let dir = self.hosts_dir().join(host).join("known-good");
        fs::create_dir_all(&dir).unwrap();
        fs::write(dir.join(s.strip_prefix("sha256:").unwrap()), bytes).unwrap();
    }
    fn plan(&self) -> Value {
        let (ok, v) = run(&["recovery-plan", self.home.to_str().unwrap()], None);
        assert!(ok, "{v}");
        assert_eq!(v["operation"], json!("recovery.plan"));
        v["result"].clone()
    }
    fn apply(&self, plan: &Value) -> (bool, Value) {
        run(
            &[
                "recovery-apply",
                self.home.to_str().unwrap(),
                "--plan-digest",
                plan["plan_digest"].as_str().unwrap(),
                "--permission-reference",
                "fixture-approved",
            ],
            Some(plan),
        )
    }
    fn digest_of(&self, dir: &Path) -> String {
        tree_digest(dir, &[]).unwrap()
    }
}

fn host<'a>(v: &'a Value, name: &str) -> &'a Value {
    v["hosts"]
        .as_array()
        .unwrap()
        .iter()
        .find(|h| h["host"] == json!(name))
        .unwrap()
}
fn file<'a>(h: &'a Value, path: &str) -> &'a Value {
    h["files"]
        .as_array()
        .unwrap()
        .iter()
        .find(|f| f["path"] == json!(path))
        .unwrap()
}
fn kind_of(v: &Value) -> String {
    let msg = v["error"]["message"].as_str().unwrap();
    assert_eq!(
        contract::classify(msg).as_str(),
        v["error"]["kind"].as_str().unwrap(),
        "{msg}"
    );
    v["error"]["kind"].as_str().unwrap().to_string()
}

/// AC-04C-01: host1 (claude) committed, host2 (codex) pending after a crash; files 1·2 are
/// `matches_after`, file 3 `matches_before`; the plan writes nothing.
#[test]
fn recovery_plan_classifies_partial_host_without_writes() {
    let f = fixture();
    f.crash_codex();
    f.known_good("codex", SKILL_V1);
    let before = f.digest_of(&f.home);
    let plan = f.plan();
    assert_eq!(f.digest_of(&f.home), before, "recovery-plan writes nothing");
    assert_eq!(host(&plan, "claude")["receipt_status"], json!("committed"));
    assert_eq!(host(&plan, "claude")["files"], json!([]));
    let codex = host(&plan, "codex");
    assert_eq!(codex["receipt_status"], json!("pending"));
    let new = file(codex, "workflows/new.md");
    assert_eq!(
        (new["class"].as_str(), new["action"].as_str()),
        (Some("matches_after"), Some("remove"))
    );
    let skill = file(codex, "SKILL.md");
    assert_eq!(
        (skill["class"].as_str(), skill["action"].as_str()),
        (Some("matches_after"), Some("restore"))
    );
    let upd = file(codex, "workflows/update.md");
    assert_eq!(
        (upd["class"].as_str(), upd["action"].as_str()),
        (Some("matches_before"), Some("none"))
    );
    // The plan is stable: a second read has the same digest.
    assert_eq!(f.plan()["plan_digest"], plan["plan_digest"]);
}

/// AC-04C-02: the pending host's `matches_after` files return to their before bytes (the
/// created file is removed, the replaced one restored from known-good); the committed
/// host's tree is byte-identical; the receipt becomes `recovered`.
#[test]
fn recovery_restores_only_this_request_changes_and_keeps_other_host() {
    let f = fixture();
    f.crash_codex();
    f.known_good("codex", SKILL_V1);
    let claude_before = f.digest_of(&f.claude());
    let claude_receipt = f.receipt("claude");
    let plan = f.plan();
    let (ok, v) = f.apply(&plan);
    assert!(ok, "{v}");
    let r = &v["result"];
    assert_eq!(r["status"], json!("success"));
    let codex = host(r, "codex");
    assert_eq!(codex["receipt_status"], json!("recovered"));
    assert_eq!(codex["recovered"], json!(["workflows/new.md", "SKILL.md"]));
    assert_eq!(codex["conflicts"], json!([]));
    // REQ-04C-06: the receipt's version_before reaches the report (#2091).
    assert_eq!(
        codex["version_before"],
        f.receipt("claude")["package_version"]
    );
    assert!(codex["version_before"].is_string());
    assert_eq!(codex["version_before_basis"], Value::Null);
    assert_eq!(
        codex["version_after"],
        f.receipt("claude")["package_version"]
    );
    assert!(!f.codex().join("workflows/new.md").exists());
    assert_eq!(
        sha(&fs::read(f.codex().join("SKILL.md")).unwrap()),
        sha(SKILL_V1.as_bytes())
    );
    assert_eq!(
        fs::read_to_string(f.codex().join("workflows/update.md")).unwrap(),
        UPDATE_V1
    );
    assert_eq!(f.receipt("codex")["status"], json!("recovered"));
    let record: Value =
        serde_json::from_slice(&fs::read(f.hosts_dir().join("codex.recovery.json")).unwrap())
            .unwrap();
    assert_eq!(record["status"], json!("recovered"));
    // The committed host: no write at all.
    assert_eq!(host(r, "claude")["written"], json!(false));
    assert_eq!(f.digest_of(&f.claude()), claude_before);
    assert_eq!(f.receipt("claude"), claude_receipt);
    // A stale or altered plan is refused before any write.
    let (ok, v) = f.apply(&plan);
    assert!(!ok);
    assert_eq!(kind_of(&v), "stale_plan");
}

/// AC-04C-03: a user edit made after the failure is preserved as `conflict`; only the
/// other file is reverted; the run is `partial` with the conflict listed; still pending.
#[test]
fn recovery_preserves_user_edit_made_after_failure() {
    let f = fixture();
    f.crash_codex();
    f.known_good("codex", SKILL_V1);
    fs::write(f.codex().join("SKILL.md"), "my own edit\n").unwrap();
    let plan = f.plan();
    let skill = file(host(&plan, "codex"), "SKILL.md");
    assert_eq!(skill["class"], json!("conflict"));
    assert_eq!(skill["reason"], json!("user_changed"));
    let (ok, v) = f.apply(&plan);
    assert!(ok, "{v}");
    let r = &v["result"];
    assert_eq!(
        r["status"],
        json!("partial"),
        "a remaining conflict is never summarised as success"
    );
    let codex = host(r, "codex");
    assert_eq!(codex["recovered"], json!(["workflows/new.md"]));
    assert_eq!(codex["conflicts"][0]["path"], json!("SKILL.md"));
    assert_eq!(codex["receipt_status"], json!("pending"));
    assert_eq!(
        fs::read_to_string(f.codex().join("SKILL.md")).unwrap(),
        "my own edit\n"
    );
    assert!(!f.codex().join("workflows/new.md").exists());
    assert_eq!(f.receipt("codex")["status"], json!("pending"));
}

/// AC-04C-04: no known-good copy, a damaged receipt, a symlink substitution and a
/// read-only directory each leave the files as they are, report conflict/blocked, and
/// never retry.
#[test]
fn recovery_refuses_without_known_good_or_on_tamper() {
    // (a) No known-good copy (the hand-built crash stores none, as for a receipt written
    // before #2091 or a copy since lost; 04-B now stores one per replaced file): the
    // replaced file is conflict(no_known_good); the created file is still removed.
    let f = fixture();
    f.crash_codex();
    let plan = f.plan();
    let skill = file(host(&plan, "codex"), "SKILL.md");
    assert_eq!(
        (skill["class"].as_str(), skill["reason"].as_str()),
        (Some("conflict"), Some("no_known_good"))
    );
    let (ok, v) = f.apply(&plan);
    assert!(ok, "{v}");
    assert_eq!(v["result"]["status"], json!("partial"));
    assert_eq!(
        fs::read_to_string(f.codex().join("SKILL.md")).unwrap(),
        SKILL_V2
    );
    // A known-good file whose bytes do not hash to the name is not a copy.
    let f = fixture();
    f.crash_codex();
    let dir = f.hosts_dir().join("codex/known-good");
    fs::create_dir_all(&dir).unwrap();
    let name = sha(SKILL_V1.as_bytes());
    fs::write(
        dir.join(name.strip_prefix("sha256:").unwrap()),
        "tampered\n",
    )
    .unwrap();
    assert_eq!(
        file(host(&f.plan(), "codex"), "SKILL.md")["reason"],
        json!("no_known_good")
    );

    // (b) Damaged receipt: the host is blocked and nothing is written.
    let f = fixture();
    f.crash_codex();
    fs::write(
        f.hosts_dir().join("codex.json"),
        b"{\"schema_version\": 1, \"surprise\": true}",
    )
    .unwrap();
    let before = f.digest_of(&f.home);
    let plan = f.plan();
    assert_eq!(host(&plan, "codex")["receipt_status"], json!("blocked"));
    let (ok, v) = f.apply(&plan);
    assert!(ok, "{v}");
    assert_eq!(v["result"]["status"], json!("partial"));
    assert_eq!(
        f.digest_of(&f.home),
        before,
        "a blocked host is not written"
    );

    // (c) Symlink substitution: the created file became a link; it is not followed.
    let f = fixture();
    f.crash_codex();
    let outside = f.home.join("outside.md");
    fs::write(&outside, NEW_V2).unwrap();
    fs::remove_file(f.codex().join("workflows/new.md")).unwrap();
    std::os::unix::fs::symlink(&outside, f.codex().join("workflows/new.md")).unwrap();
    let plan = f.plan();
    let new = file(host(&plan, "codex"), "workflows/new.md");
    assert_eq!(
        (new["class"].as_str(), new["reason"].as_str()),
        (Some("conflict"), Some("symlink"))
    );
    let (ok, _) = f.apply(&plan);
    assert!(ok);
    assert!(fs::symlink_metadata(f.codex().join("workflows/new.md"))
        .unwrap()
        .file_type()
        .is_symlink());
    assert_eq!(
        fs::read_to_string(&outside).unwrap(),
        NEW_V2,
        "the link target is untouched"
    );

    // (d) Read-only directory: the remove fails; nothing is reverted, the receipt stays
    // pending with the failure recorded, and the command does not retry.
    let f = fixture();
    f.crash_codex();
    let wf = f.codex().join("workflows");
    let mut perm = fs::metadata(&wf).unwrap().permissions();
    std::os::unix::fs::PermissionsExt::set_mode(&mut perm, 0o555);
    fs::set_permissions(&wf, perm.clone()).unwrap();
    let plan = f.plan();
    let (ok, v) = f.apply(&plan);
    std::os::unix::fs::PermissionsExt::set_mode(&mut perm, 0o755);
    fs::set_permissions(&wf, perm).unwrap();
    assert!(!ok);
    assert_eq!(kind_of(&v), "conflict", "{v}");
    assert!(f.codex().join("workflows/new.md").exists());
    assert_eq!(f.receipt("codex")["status"], json!("pending"));
    let record: Value =
        serde_json::from_slice(&fs::read(f.hosts_dir().join("codex.recovery.json")).unwrap())
            .unwrap();
    assert_eq!(record["status"], json!("failed"));
    assert_eq!(record["failed"]["path"], json!("workflows/new.md"));
    assert_eq!(record["reverted"], json!([]));

    // (e) A real pending receipt from 04-B's own failure path (sabotaged codex tree).
    let f = fixture();
    fs::remove_dir_all(f.codex().join("workflows")).unwrap();
    fs::write(f.codex().join("workflows"), b"not a directory\n").unwrap();
    let (ok, v) = run(
        &[
            "update-plan",
            f.home.to_str().unwrap(),
            f.stage2.to_str().unwrap(),
        ],
        None,
    );
    assert!(ok, "{v}");
    let p = v["result"].clone();
    let (ok, v) = run(
        &[
            "update-apply",
            f.home.to_str().unwrap(),
            f.stage2.to_str().unwrap(),
            "--plan-digest",
            p["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "fixture-approved",
        ],
        Some(&p),
    );
    assert!(ok, "{v}");
    assert_eq!(f.receipt("codex")["status"], json!("pending"));
    let plan = f.plan();
    let codex = host(&plan, "codex");
    assert_eq!(codex["receipt_status"], json!("pending"));
    // SKILL.md was replaced before the failure; 04-B (#2091) kept its known-good copy, so
    // it is revertible. The files under the sabotaged `workflows` path are not regular
    // files and are preserved as conflicts.
    let skill = file(codex, "SKILL.md");
    assert_eq!(
        (skill["class"].as_str(), skill["action"].as_str()),
        (Some("matches_after"), Some("restore"))
    );
    assert_eq!(
        file(codex, "workflows/new.md")["reason"],
        json!("not_a_regular_file")
    );
}

/// AC-04C-02 against 04-B's real known-good copy and REQ-04C-06 `version_before`: the
/// receipt 04-B itself wrote (sabotaged mid-apply) is recovered from the copy 04-B stored,
/// and the report carries the version that failed apply replaced.
#[test]
fn recovery_uses_real_known_good_and_reports_version_before() {
    let f = fixture();
    let claude_before = f.digest_of(&f.claude());
    fs::remove_dir_all(f.codex().join("workflows")).unwrap();
    fs::write(f.codex().join("workflows"), b"not a directory\n").unwrap();
    let home = f.home.to_str().unwrap();
    let stage = f.stage2.to_str().unwrap();
    let (ok, v) = run(&["update-plan", home, stage], None);
    assert!(ok, "{v}");
    let p = v["result"].clone();
    let (ok, v) = run(
        &[
            "update-apply",
            home,
            stage,
            "--plan-digest",
            p["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "fixture-approved",
        ],
        Some(&p),
    );
    assert!(ok, "{v}");
    let receipt = f.receipt("codex");
    assert_eq!(receipt["status"], json!("pending"));
    assert!(
        receipt["version_before"].is_string(),
        "04-B records version_before: {receipt}"
    );
    // The copy 04-B stored, under the REQ-04C-03 path.
    let hex = sha(SKILL_V1.as_bytes());
    assert!(f
        .hosts_dir()
        .join("codex/known-good")
        .join(hex.strip_prefix("sha256:").unwrap())
        .is_file());
    assert_eq!(
        fs::read_to_string(f.codex().join("SKILL.md")).unwrap(),
        SKILL_V2
    );
    // claude committed v2 in the same run; it must stay as it is now.
    let claude_after_update = f.digest_of(&f.claude());
    assert_ne!(claude_after_update, claude_before);

    let plan = f.plan();
    assert_eq!(
        host(&plan, "codex")["version_before"],
        receipt["version_before"]
    );
    let (ok, v) = f.apply(&plan);
    assert!(ok, "{v}");
    let codex = host(&v["result"], "codex");
    assert_eq!(codex["recovered"], json!(["SKILL.md"]));
    assert_eq!(codex["version_before"], receipt["version_before"]);
    assert_eq!(codex["version_before_basis"], Value::Null);
    // Conflicts remain under the sabotaged path, so the host stays pending and the run is
    // partial; SKILL.md is back to its v1 bytes from the real copy.
    assert_eq!(v["result"]["status"], json!("partial"));
    assert_eq!(codex["receipt_status"], json!("pending"));
    assert_eq!(
        fs::read_to_string(f.codex().join("SKILL.md")).unwrap(),
        SKILL_V1
    );
    assert_eq!(
        f.digest_of(&f.claude()),
        claude_after_update,
        "committed host untouched"
    );
    // claude's committed v2 receipt also records the version it replaced: no basis needed.
    assert!(host(&v["result"], "claude")["version_before"].is_string());
    assert_eq!(
        host(&v["result"], "claude")["version_before_basis"],
        Value::Null
    );
}

/// AC-04C-05: a pending host refuses a new update-apply until recovered; the committed
/// host plans and applies as usual; after recovery the host applies normally.
#[test]
fn pending_host_blocks_reapply_until_recovered() {
    let f = fixture();
    f.crash_codex();
    f.known_good("codex", SKILL_V1);
    let home = f.home.to_str().unwrap();
    let stage = f.stage2.to_str().unwrap();
    let (ok, v) = run(&["update-plan", home, stage], None);
    assert!(ok, "{v}");
    let p = v["result"].clone();
    let codex = host(&p, "codex");
    assert_eq!(codex["preflight"], json!("blocked"));
    assert_eq!(
        codex["blocked_reason"],
        json!("host receipt is pending; run recovery-plan first")
    );
    assert_eq!(host(&p, "claude")["preflight"], json!("ok"));
    let codex_before = f.digest_of(&f.codex());
    let (ok, v) = run(
        &[
            "update-apply",
            home,
            stage,
            "--plan-digest",
            p["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "fixture-approved",
        ],
        Some(&p),
    );
    assert!(ok, "{v}");
    assert_eq!(host(&v["result"], "codex")["status"], json!("blocked"));
    assert_eq!(host(&v["result"], "claude")["status"], json!("applied"));
    assert_eq!(
        f.digest_of(&f.codex()),
        codex_before,
        "pending host not written"
    );

    // Recover, then the host plans and applies again.
    let plan = f.plan();
    assert!(f.apply(&plan).0);
    let (ok, v) = run(&["update-plan", home, stage], None);
    assert!(ok, "{v}");
    let p = v["result"].clone();
    assert_eq!(host(&p, "codex")["preflight"], json!("ok"));
    let (ok, v) = run(
        &[
            "update-apply",
            home,
            stage,
            "--plan-digest",
            p["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "fixture-approved",
        ],
        Some(&p),
    );
    assert!(ok, "{v}");
    assert_eq!(host(&v["result"], "codex")["status"], json!("applied"));
    assert_eq!(f.receipt("codex")["status"], json!("committed"));
    assert_eq!(
        fs::read_to_string(f.codex().join("SKILL.md")).unwrap(),
        SKILL_V2
    );
}

/// AC-04C-06: a second recovery finds nothing to revert and changes nothing.
#[test]
fn recovery_rerun_is_noop() {
    let f = fixture();
    f.crash_codex();
    f.known_good("codex", SKILL_V1);
    assert!(f.apply(&f.plan()).0);
    let before = f.digest_of(&f.home);
    let plan = f.plan();
    assert_eq!(host(&plan, "codex")["receipt_status"], json!("recovered"));
    assert_eq!(host(&plan, "codex")["files"], json!([]));
    let (ok, v) = f.apply(&plan);
    assert!(ok, "{v}");
    assert_eq!(v["result"]["status"], json!("success"));
    assert!(v["result"]["hosts"]
        .as_array()
        .unwrap()
        .iter()
        .all(|h| h["written"] == json!(false)));
    assert_eq!(f.digest_of(&f.home), before);
    // The fake Go state in the same HOME is untouched throughout (AC-04C-07 fixture half).
    assert_eq!(
        fs::read_to_string(f.home.join(".moai/config.yaml")).unwrap(),
        "generation: go\n"
    );
}
