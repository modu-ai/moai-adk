//! 08-C (#2014) T1 deterministic host-fixture E2E: SETUP → UPDATE → interrupt → recovery,
//! digest-mismatch fail-closed, re-run diff 0, real-HOME write denial, and clean-clone
//! portability — all in temporary HOME and temporary project directories, never the real
//! user HOME or the Go generation's state. The interrupt is real (a read-only projection
//! directory fails the apply mid-transaction after an earlier file published), matching
//! the end state a crash leaves; no injection hook exists in `apply_update` and none is
//! added here. T2 (real host loading, AC-08C-3·4·6) is out of this file's scope:
//! `tests/e2e/host-claude.sh` and `tests/e2e/host-codex.sh` carry those commands for the
//! operator's measured run, recorded as not_run in `docs/e2e-matrix.md`.
//! AC command names must match this file: `cargo test --offline --locked --test
//! e2e_host_fixture`.

#![cfg(unix)]

mod common;

use moai_cli::tree_digest;
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
const NEW_V2: &str = "# NEW\nadded in v2\n";
const ROUTES: &str = "{\"routes\":1}\n";
const USER_ROUTES: &str = "{\"routes\":1,\"user\":true}\n";

fn sha_hex(bytes: &[u8]) -> String {
    moai_cli::digest(bytes)
        .trim_start_matches("sha256:")
        .to_string()
}

fn write_pkg(dir: &Path, files: &[(&str, &str)]) {
    for (rel, body) in files {
        let p = dir.join(rel);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        fs::write(p, body).unwrap();
    }
}

fn stage_v1() -> Vec<(&'static str, &'static str)> {
    vec![
        ("SKILL.md", SKILL_V1),
        ("references/routes.json", ROUTES),
        ("workflows/update.md", UPDATE_V1),
    ]
}

fn stage_v2() -> Vec<(&'static str, &'static str)> {
    vec![
        ("SKILL.md", SKILL_V2),
        ("references/routes.json", ROUTES), // byte-identical: preserved unless a user edit
        ("workflows/update.md", UPDATE_V2),
        ("workflows/new.md", NEW_V2),
    ]
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

fn plan_and_apply(home: &Path, stage: &Path) -> Value {
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
    let plan = v["result"].clone();
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
        Some(&plan),
        home,
    );
    assert!(ok, "{v}");
    v["result"].clone()
}

/// The 08-C SETUP leg on a temporary project (project-plan/apply, 01-B) plus the
/// package install into the temporary HOME (update-apply v1, 04-B) — both hosts, one
/// installation, fake Go-generation state in the same HOME that must stay untouched.
struct Fixture {
    _root: TempDir,
    home: PathBuf,
    stage2: PathBuf,
}

fn fixture() -> Fixture {
    let root = TempDir::new().unwrap();
    let rootp = root.path().to_path_buf();
    let home = rootp.join("home");
    fs::create_dir_all(home.join(".moai/cache")).unwrap();
    fs::write(home.join(".moai/config.yaml"), "generation: go\n").unwrap();
    fs::write(home.join(".moai/cache/blob"), b"go state bytes").unwrap();
    let project = rootp.join("project");
    fs::create_dir_all(&project).unwrap();
    // SETUP (01-B): a v2 profile previewed and applied into the temporary project.
    let input = json!({"schema_version":2,"project_id":"demo","purpose":"Build a small editor","tools":["Rust"],"development_methods":["TDD"],"hosts":["claude","codex"],"interview_revision":"rc1-fixture-1"});
    let (ok, v) = run(
        &["project-plan", project.to_str().unwrap()],
        Some(&input),
        &home,
    );
    assert!(ok, "{v}");
    let plan = v["result"].clone();
    let (ok, v) = run(
        &[
            "project-apply",
            project.to_str().unwrap(),
            "--plan-digest",
            plan["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "fixture-approved",
        ],
        Some(&plan),
        &home,
    );
    assert!(ok, "{v}");
    // INSTALL (04-B): v1 into both host projections of the temporary HOME.
    let stage1 = rootp.join("stage-v1");
    write_pkg(&stage1, &stage_v1());
    let result = plan_and_apply(&home, &stage1);
    assert_eq!(result["status"], json!("applied"), "{result}");
    let stage2 = rootp.join("stage-v2");
    write_pkg(&stage2, &stage_v2());
    Fixture {
        _root: root,
        home,
        stage2,
    }
}

impl Fixture {
    fn claude(&self) -> PathBuf {
        self.home.join(".claude/skills/moai")
    }

    fn codex(&self) -> PathBuf {
        self.home.join(".agents/skills/moai")
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

    fn go_digest(&self) -> String {
        tree_digest(&self.home.join(".moai"), &[]).unwrap()
    }

    fn plan_v2(&self) -> Value {
        let (ok, v) = run(
            &[
                "update-plan",
                self.home.to_str().unwrap(),
                self.stage2.to_str().unwrap(),
            ],
            None,
            &self.home,
        );
        assert!(ok, "{v}");
        v["result"].clone()
    }

    fn apply_v2(&self, plan: &Value) -> (bool, Value) {
        run(
            &[
                "update-apply",
                self.home.to_str().unwrap(),
                self.stage2.to_str().unwrap(),
                "--plan-digest",
                plan["plan_digest"].as_str().unwrap(),
                "--permission-reference",
                "fixture-approved",
            ],
            Some(plan),
            &self.home,
        )
    }

    fn recovery_plan(&self) -> Value {
        let (ok, v) = run(
            &["recovery-plan", self.home.to_str().unwrap()],
            None,
            &self.home,
        );
        assert!(ok, "{v}");
        assert_eq!(v["operation"], json!("recovery.plan"));
        v["result"].clone()
    }

    fn recovery_apply(&self, plan: &Value) -> Value {
        let (ok, v) = run(
            &[
                "recovery-apply",
                self.home.to_str().unwrap(),
                "--plan-digest",
                plan["plan_digest"].as_str().unwrap(),
                "--permission-reference",
                "fixture-approved",
            ],
            Some(plan),
            &self.home,
        );
        assert!(ok, "{v}");
        v["result"].clone()
    }
}

/// AC-08C-1 (HA-02/RL-03): SETUP → UPDATE with a real mid-apply interruption →
/// recovery restores the replaced file from its known-good copy, the preserved user
/// edit survives byte for byte, the other host stays committed, and the fake
/// Go-generation state in the same HOME never moves.
#[test]
fn e2e_setup_update_interrupt_restore_preserves_user_edits() {
    let f = fixture();
    // A user edit in the claude projection before the update: v2's routes.json is
    // byte-identical to v1's, so the plan must classify the edited file as a conflict
    // and never overwrite it.
    let user_routes = f.claude().join("references/routes.json");
    fs::write(&user_routes, USER_ROUTES).unwrap();
    let go_before = f.go_digest();

    // UPDATE v2 with a real interruption: the workflows directory becomes unwritable,
    // so the apply fails mid-transaction after SKILL.md published. The known-good copy
    // of the overwritten SKILL.md was saved at transaction start (REQ-04C-03).
    let workflows = f.claude().join("workflows");
    let original = fs::metadata(&workflows).unwrap().permissions();
    let mut locked = original.clone();
    locked.set_readonly(true);
    fs::set_permissions(&workflows, locked).unwrap();
    let plan = f.plan_v2();
    let (ok, v) = f.apply_v2(&plan);
    // The operation itself succeeds and reports "partial": the interrupted host failed
    // mid-transaction (SKILL.md published; the next write refused), the other host
    // applied.
    assert!(ok, "the operation envelope must succeed: {v}");
    assert_eq!(v["result"]["status"], json!("partial"), "{v}");
    fs::set_permissions(&workflows, original).unwrap();

    let receipt = f.receipt(HOST);
    assert_eq!(receipt["status"], json!("pending"), "{receipt}");
    assert_eq!(
        receipt["version_before"], receipt["package_version"],
        "#2091 version_before must be recorded on the pending receipt"
    );
    let skill_copy = f
        .hosts_dir()
        .join(HOST)
        .join("known-good")
        .join(sha_hex(SKILL_V1.as_bytes()));
    assert!(
        skill_copy.is_file(),
        "known-good copy must exist: {skill_copy:?}"
    );

    // RECOVERY: the replaced SKILL.md is matches_after (revertible from known-good).
    let rplan = f.recovery_plan();
    let result = f.recovery_apply(&rplan);
    let claude = result["hosts"]
        .as_array()
        .unwrap()
        .iter()
        .find(|h| h["host"] == json!(HOST))
        .unwrap();
    assert_eq!(claude["receipt_status"], json!("recovered"), "{claude}");

    // The interrupted host is back at v1 exactly; the user edit survived byte for byte;
    // the other host keeps its committed v2.
    assert_eq!(
        fs::read(f.claude().join("SKILL.md")).unwrap(),
        SKILL_V1.as_bytes()
    );
    assert_eq!(fs::read(&user_routes).unwrap(), USER_ROUTES.as_bytes());
    assert_eq!(
        fs::read(f.claude().join("workflows/update.md")).unwrap(),
        UPDATE_V1.as_bytes()
    );
    assert!(!f.claude().join("workflows/new.md").exists());
    assert_eq!(
        fs::read(f.codex().join("SKILL.md")).unwrap(),
        SKILL_V2.as_bytes(),
        "the committed host must keep the applied v2"
    );
    assert_eq!(f.receipt("codex")["status"], json!("committed"));
    assert_eq!(
        f.go_digest(),
        go_before,
        "Go-generation state must not move"
    );
    // The updated package still reads as fresh for the user-edited file: v2's template
    // did not change it, so the plan keeps the user's edit as "preserved"
    // (AC-04B-03) — a rerun never overwrites user bytes.
    let plan_again = f.plan_v2();
    let routes = plan_again["hosts"]
        .as_array()
        .unwrap()
        .iter()
        .find(|h| h["host"] == json!(HOST))
        .unwrap()["files"]
        .as_array()
        .unwrap()
        .iter()
        .find(|x| x["path"] == json!("references/routes.json"))
        .unwrap();
    assert_eq!(routes["action"], json!("preserved"), "{routes}");
}

/// AC-08C-2 (RL-07/HA-06): staged bytes changed after the plan digest was taken → the
/// apply is refused before anything is written; the whole HOME tree is byte-identical.
#[test]
fn e2e_digest_mismatch_is_fail_closed() {
    let f = fixture();
    let before = tree_digest(&f.home, &[]).unwrap();
    let plan = f.plan_v2();
    // Tamper with the staged package after the digest was computed.
    fs::write(f.stage2.join("workflows/new.md"), "# NEW\ntampered\n").unwrap();
    let (ok, v) = f.apply_v2(&plan);
    assert!(!ok, "a digest mismatch must be refused: {v}");
    assert_eq!(
        tree_digest(&f.home, &[]).unwrap(),
        before,
        "a refused apply must change nothing"
    );
}

/// HA-02: re-running the same update is a diff of zero — every action comes back
/// unchanged/preserved and the apply writes nothing.
#[test]
fn e2e_second_update_is_diff_zero() {
    let f = fixture();
    plan_and_apply(&f.home, &f.stage2);
    let after_first = tree_digest(&f.home, &[]).unwrap();
    let plan = f.plan_v2();
    for h in plan["hosts"].as_array().unwrap() {
        for file in h["files"].as_array().unwrap() {
            assert_ne!(file["action"], json!("update"));
            assert_ne!(file["action"], json!("create"));
        }
    }
    let result = plan_and_apply(&f.home, &f.stage2);
    assert_eq!(result["status"], json!("unchanged"), "{result}");
    assert_eq!(
        tree_digest(&f.home, &[]).unwrap(),
        after_first,
        "a re-run must be a diff of zero"
    );
}

/// AC-08C-7 (RL-GENERATION-ISOLATION, deny-home sandbox form allowed by the leader's
/// scope comment): the whole SETUP+UPDATE scenario runs under a sandbox that denies
/// file writes anywhere under the caller's REAL home, so a lost HOME redirection fails
/// loudly instead of writing user state. Also asserts the fixture's fake Go-generation
/// digest and that the receipt/known-good state stays inside the temporary HOME.
/// Measured on macOS only: `sandbox-exec` does not exist on Linux, so on that platform
/// this test is not compiled (an explicit cfg gate, not a hidden skip) and
/// docs/e2e-matrix.md records AC-08C-7 as not_run there — a green Linux CI is NOT an
/// AC-08C-7 measurement.
///
/// The denied subtree is the account's home directory from the password database, not
/// `$HOME`: the documented T1 command runs cargo with `HOME=$(mktemp -d)`, and a profile
/// built from `$HOME` would then deny only that throwaway directory (Issue #2014,
/// agent-8 N1). Only the path is read; nothing under the real home is read or written.
#[cfg(target_os = "macos")]
#[test]
fn e2e_never_writes_outside_temp_home() {
    let real_home = passwd_home();
    let dir = TempDir::new().unwrap();
    let profile = dir.path().join("deny-home.sb");
    fs::write(
        &profile,
        format!(
            "(version 1)\n(allow default)\n(deny file-write* (subpath \"{}\"))\n",
            real_home.display()
        ),
    )
    .unwrap();
    // Probe with the same profile on a no-op command: distinguishes "sandbox-exec
    // works here" from "this machine cannot run the denial check".
    let probe = Command::new("sandbox-exec")
        .arg("-f")
        .arg(&profile)
        .arg("/usr/bin/true")
        .output();
    match probe {
        Ok(o) if o.status.success() => {}
        Ok(o) => panic!(
            "sandbox-exec refused its init (exit {:?}); the denial test cannot run \
             honestly: {}",
            o.status.code(),
            String::from_utf8_lossy(&o.stderr)
        ),
        Err(e) => panic!("sandbox-exec not found; cannot run the denial test: {e}"),
    }

    let f = fixture();
    // The fixture HOME must lie outside the denied subtree, or a pass would only mean
    // the sandbox denied nothing the apply touched.
    assert!(
        !f.home.starts_with(&real_home),
        "the fixture HOME {:?} lies under the denied account home {real_home:?}; set \
         TMPDIR outside it to measure AC-08C-7",
        f.home
    );
    let go_before = f.go_digest();
    let plan = f.plan_v2();
    let args = [
        EXE.to_string(),
        "update-apply".into(),
        f.home.to_str().unwrap().into(),
        f.stage2.to_str().unwrap().into(),
        "--plan-digest".into(),
        plan["plan_digest"].as_str().unwrap().into(),
        "--permission-reference".into(),
        "fixture-approved".into(),
    ];
    let out = Command::new("sandbox-exec")
        .arg("-f")
        .arg(&profile)
        .args(&args)
        .env("HOME", &f.home)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_NOSYSTEM", "1")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    let mut child = out;
    common::feed_stdin(&mut child, plan.to_string().as_bytes());
    let out = child.wait_with_output().unwrap();
    assert!(
        out.status.success(),
        "the sandboxed apply must succeed with HOME redirected: {} {}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    assert_eq!(f.receipt(HOST)["status"], json!("committed"));
    assert_eq!(f.go_digest(), go_before);
}

/// The current account's home directory from the password database (`getpwuid_r`),
/// independent of `$HOME`. Panics when the entry cannot be read: without the real path
/// the denial check would measure nothing.
#[cfg(target_os = "macos")]
fn passwd_home() -> PathBuf {
    use std::ffi::CStr;
    use std::os::unix::ffi::OsStrExt as _;
    let mut buf = vec![0 as libc::c_char; 16 * 1024];
    let mut pwd: libc::passwd = unsafe { std::mem::zeroed() };
    let mut found: *mut libc::passwd = std::ptr::null_mut();
    // SAFETY: every pointer refers to a live local of the stated size; on success
    // `pw_dir` points into `buf`, which outlives the copy below.
    let rc = unsafe {
        libc::getpwuid_r(
            libc::getuid(),
            &mut pwd,
            buf.as_mut_ptr(),
            buf.len(),
            &mut found,
        )
    };
    assert!(
        rc == 0 && !found.is_null() && !pwd.pw_dir.is_null(),
        "cannot read this account's passwd entry (rc {rc}); the deny-home check \
         cannot run honestly"
    );
    // SAFETY: checked non-null above; NUL-terminated by getpwuid_r.
    let dir = unsafe { CStr::from_ptr(pwd.pw_dir) };
    let home = PathBuf::from(std::ffi::OsStr::from_bytes(dir.to_bytes()));
    assert!(
        home.is_absolute() && home != Path::new("/"),
        "unusable passwd home {home:?}"
    );
    home
}

/// AC-08C-8 (HA-04, T1 half): a clean clone carries the full portable asset set for
/// READ/PLAN — the skill, its routes, and every workflow the routes name — without any
/// moai binary present. The host-level READ/PLAN run (the AC's `claude --bare` command)
/// is the operator's T2 measurement, recorded not_run in docs/e2e-matrix.md.
#[test]
fn e2e_clean_clone_carries_portable_read_plan_assets() {
    let manifest = PathBuf::from(env!("CARGO_MANIFEST_DIR"));
    let dest = TempDir::new().unwrap();
    let clone = dest.path().join("clean");
    clone_checkout(&manifest, &clone);

    // The portable assets are all plain files in the clone — no binary, no state dir.
    let skill = clone.join("skills/moai");
    let skill_md = fs::read_to_string(skill.join("SKILL.md")).unwrap();
    assert!(
        skill_md.starts_with("---\nname: moai\ndescription: "),
        "SKILL.md frontmatter must name the skill: {skill_md}"
    );
    let routes: Value =
        serde_json::from_slice(&fs::read(skill.join("references/routes.json")).unwrap()).unwrap();
    // READ and PLAN are the routes AC-08C-8 needs without a binary: the routes file
    // itself declares requires_binary=false for them, and the clone carries the
    // workflow files they reference.
    for route in ["READ", "PLAN"] {
        let r = &routes["routes"][route];
        assert!(
            r["requires_binary"] == json!(false),
            "{route} must be a no-binary route: {r}"
        );
        let rel = r["reference"].as_str().unwrap();
        let resolved = skill.join("references").join(rel);
        assert!(
            resolved.is_file(),
            "{route} references {rel} but the clean clone lacks it"
        );
    }
    // Every other referenced workflow file must exist in the clone too.
    let refs = &routes["routes"];
    for (name, r) in refs.as_object().unwrap() {
        let rel = r["reference"].as_str().unwrap();
        let resolved = skill.join("references").join(rel);
        assert!(
            resolved.is_file(),
            "route {name} references {rel} but the clean clone lacks it"
        );
    }
    // The clone carries no installed user state (only repo content such as the
    // tracked .moai-cli manifests): nothing under the clone references a real HOME.
}

fn clone_checkout(manifest: &Path, clone: &Path) {
    let discovery = common::git::command()
        .args([
            "-C",
            manifest.to_str().unwrap(),
            "rev-parse",
            "--git-common-dir",
        ])
        .output()
        .unwrap();
    let discovered = String::from_utf8_lossy(&discovery.stdout)
        .trim()
        .to_string();
    if !discovery.status.success() || discovered.is_empty() {
        // Explicit not_run that cargo cannot count as ok (Issue #2104): this tree
        // carries no git metadata (e.g. an exported source archive), so there is no
        // repository to clean-clone. A silent return read as a vacuous pass — the
        // one thing RL-10 forbids — so the unmeasurable condition surfaces as a
        // failure naming its reason. docs/e2e-matrix.md records the combination as
        // not_run; inside a git checkout this path is never taken.
        panic!(
            "NOT RUN: e2e_clean_clone_carries_portable_read_plan_assets — no git \
             metadata in this tree ({}); the clean-clone portability check was not \
             performed. Run from a git checkout, or record the combination as \
             not_run in docs/e2e-matrix.md",
            String::from_utf8_lossy(&discovery.stderr).trim()
        );
    }
    let gitdir = PathBuf::from(discovered);
    let gitdir = if gitdir.is_absolute() {
        gitdir
    } else {
        manifest.join(gitdir)
    };
    // Pin this checkout's commit: a linked worktree may differ from the common-dir
    // default branch, and CI checks out a detached merge commit with no branch name.
    let revision = {
        let out = common::git::command()
            .args([
                "-C",
                manifest.to_str().unwrap(),
                "rev-parse",
                "--verify",
                "HEAD^{commit}",
            ])
            .output()
            .unwrap();
        assert!(out.status.success(), "cannot resolve the tested commit");
        String::from_utf8(out.stdout).unwrap().trim().to_string()
    };
    // A local copy includes the detached worktree's object even if no branch names
    // it. Do not use a shallow transport clone, which only fetches advertised refs.
    let out = common::git::command()
        .args([
            "clone",
            "--no-hardlinks",
            "--no-checkout",
            gitdir.to_str().unwrap(),
            clone.to_str().unwrap(),
        ])
        .output()
        .unwrap();
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    let checkout = common::git::command()
        .arg("-C")
        .arg(clone)
        .args(["checkout", "--detach", &revision])
        .output()
        .unwrap();
    assert!(
        checkout.status.success(),
        "{}",
        String::from_utf8_lossy(&checkout.stderr)
    );
}

/// The clone must select the tested revision even when the common-dir HEAD differs.
#[test]
fn e2e_clean_clone_keeps_exact_attached_and_detached_revision() {
    fn git(root: &Path, args: &[&str]) -> String {
        let out = common::git::command()
            .arg("-C")
            .arg(root)
            .args([
                "-c",
                "user.name=Fixture",
                "-c",
                "user.email=fixture@example.invalid",
                "-c",
                "commit.gpgsign=false",
                "-c",
                "core.hooksPath=/dev/null",
            ])
            .args(args)
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_NOSYSTEM", "1")
            .output()
            .unwrap();
        assert!(
            out.status.success(),
            "{}",
            String::from_utf8_lossy(&out.stderr)
        );
        String::from_utf8(out.stdout).unwrap().trim().to_string()
    }
    let temp = TempDir::new().unwrap();
    let source = temp.path().join("source");
    fs::create_dir(&source).unwrap();
    git(&source, &["init", "--initial-branch=default"]);
    fs::write(source.join("asset"), "default").unwrap();
    git(&source, &["add", "asset"]);
    git(&source, &["commit", "-m", "default assets"]);
    let lane = temp.path().join("lane");
    git(
        &source,
        &["worktree", "add", "-b", "lane", lane.to_str().unwrap()],
    );
    fs::write(lane.join("asset"), "attached lane").unwrap();
    git(&lane, &["add", "asset"]);
    git(&lane, &["commit", "-m", "lane assets"]);
    let attached = temp.path().join("attached-clone");
    clone_checkout(&lane, &attached);
    assert_eq!(
        git(&attached, &["rev-parse", "HEAD"]),
        git(&lane, &["rev-parse", "HEAD"])
    );
    assert_eq!(
        fs::read_to_string(attached.join("asset")).unwrap(),
        "attached lane"
    );

    git(&lane, &["checkout", "--detach"]);
    fs::write(lane.join("asset"), "detached merge assets").unwrap();
    git(&lane, &["add", "asset"]);
    git(&lane, &["commit", "-m", "detached assets"]);
    let detached = temp.path().join("detached-clone");
    clone_checkout(&lane, &detached);
    assert_eq!(
        git(&detached, &["rev-parse", "HEAD"]),
        git(&lane, &["rev-parse", "HEAD"])
    );
    assert_eq!(
        fs::read_to_string(detached.join("asset")).unwrap(),
        "detached merge assets"
    );
    assert_eq!(fs::read_to_string(source.join("asset")).unwrap(), "default");
}
