//! #2016 CLI wiring on a synthetic HOME (user decision 6784520717): the real binary runs
//! in a child process whose `HOME` and `TMPDIR` point at temporary directories, so no test
//! here ever names or reads the user's own home.

mod common;
use moai_cli::migrate::{hash, SYNTHETIC_MARKER, SYNTHETIC_MARKER_TEXT};
use serde_json::{json, Value};
use std::{
    fs,
    io::Write,
    path::Path,
    process::{Command, Stdio},
};
use tempfile::TempDir;

const CONFIG: &[u8] = b"{\"language\":\"ko\"}";

struct Env {
    /// Stands in for the user's home in the child (`HOME`).
    user: TempDir,
    /// The child's `TMPDIR`.
    tmp: TempDir,
}

impl Env {
    fn new() -> Self {
        Self {
            user: TempDir::new().unwrap(),
            tmp: TempDir::new().unwrap(),
        }
    }
    /// A prepared synthetic home under the child's temporary directory.
    fn home(&self, name: &str) -> std::path::PathBuf {
        let h = self.tmp.path().join(name);
        fs::create_dir_all(h.join(".moai/config")).unwrap();
        fs::write(h.join(".moai/config/language.json"), CONFIG).unwrap();
        fs::write(h.join(SYNTHETIC_MARKER), SYNTHETIC_MARKER_TEXT).unwrap();
        h
    }
    fn run(&self, args: &[&str], stdin: &Value) -> (i32, Value) {
        self.run_with(args, stdin, self.user.path(), self.tmp.path())
    }
    fn run_with(&self, args: &[&str], stdin: &Value, home: &Path, tmp: &Path) -> (i32, Value) {
        let mut child = Command::new(crate::common::MOAI_BIN)
            .args(args)
            .env("HOME", home)
            .env("TMPDIR", tmp)
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
            .unwrap();
        child
            .stdin
            .take()
            .unwrap()
            .write_all(&serde_json::to_vec(stdin).unwrap())
            .unwrap();
        let out = child.wait_with_output().unwrap();
        let v: Value = serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
            panic!(
                "non-JSON stdout for {args:?}: {} / {}",
                String::from_utf8_lossy(&out.stdout),
                String::from_utf8_lossy(&out.stderr)
            )
        });
        (out.status.code().unwrap_or(-1), v)
    }
}

fn request(release: &[String], project_current: Option<&str>) -> Value {
    json!({
        "hosts": ["Claude", "Codex"],
        "mode": "Default",
        "go_release": {"version": "v3.2.1", "sha256": release},
        "selections": [{
            "source": ".moai/config/language.json",
            "target": ".moai-cli/imports/language.json",
            "recorded_sha256": hash(CONFIG),
            "project_current": project_current,
        }],
    })
}

fn digest(p: &Path) -> String {
    moai_cli::tree_digest(p, &[]).unwrap()
}

fn s(p: &Path) -> &str {
    p.to_str().unwrap()
}

#[test]
fn plan_apply_recover_round_trip_through_the_cli() {
    let env = Env::new();
    let home = env.home("synthetic");
    let h = s(&home);

    let before = digest(&home);
    let (code, v) = env.run(&["migrate-plan", h], &request(&[hash(CONFIG)], None));
    assert_eq!(code, 0, "{v}");
    assert_eq!(v["operation"], "migrate.plan");
    assert_eq!(v["result"]["writes"], 0);
    assert_eq!(v["result"]["applicable"], true);
    assert_eq!(v["result"]["conflict_preview"], json!([]));
    assert_eq!(digest(&home), before, "preview wrote");
    let plan = v["result"]["plan"].clone();
    let d = v["result"]["plan_digest"].as_str().unwrap().to_string();

    let (code, v) = env.run(
        &[
            "migrate-apply",
            h,
            "--plan-digest",
            &d,
            "--permission-reference",
            "decision 6784520717",
        ],
        &plan,
    );
    assert_eq!(code, 0, "{v}");
    assert_eq!(
        v["result"]["imports"],
        json!([".moai-cli/imports/language.json"])
    );
    assert_eq!(v["result"]["backup"], d.as_str());
    assert_eq!(
        fs::read(home.join(".moai-cli/imports/language.json")).unwrap(),
        CONFIG
    );
    assert_eq!(
        fs::read(home.join(".moai/config/language.json")).unwrap(),
        CONFIG
    );
    let manifest: Value = serde_json::from_slice(
        &fs::read(home.join(".moai-go-backup").join(&d).join("manifest.json")).unwrap(),
    )
    .unwrap();
    assert_eq!(manifest["permission_reference"], "decision 6784520717");

    let (code, v) = env.run(&["migrate-recovery-plan", h, "--backup", &d], &json!(null));
    assert_eq!(code, 0, "{v}");
    assert_eq!(v["result"]["actions"], json!(["removed_import"]));
    let rplan = v["result"]["plan"].clone();
    let rd = v["result"]["plan_digest"].as_str().unwrap().to_string();

    // A wrong recovery digest is stale and removes nothing.
    let (code, v) = env.run(
        &[
            "migrate-recover",
            h,
            "--backup",
            &d,
            "--plan-digest",
            &"0".repeat(64),
            "--permission-reference",
            "r",
        ],
        &rplan,
    );
    assert_ne!(code, 0);
    assert_eq!(v["error"]["kind"], "stale_plan", "{v}");
    assert!(home.join(".moai-cli/imports/language.json").is_file());

    let (code, v) = env.run(
        &[
            "migrate-recover",
            h,
            "--backup",
            &d,
            "--plan-digest",
            &rd,
            "--permission-reference",
            "r",
        ],
        &rplan,
    );
    assert_eq!(code, 0, "{v}");
    assert_eq!(v["result"]["actions"], json!(["removed_import"]));
    assert!(!home.join(".moai-cli/imports/language.json").exists());
    assert!(home.join(".moai/config/language.json").is_file());
    assert!(home.join(".moai-go-backup").join(&d).join("0").is_file());
}

#[test]
fn the_confirmation_gate_refuses_stale_altered_or_unreferenced_plans() {
    let env = Env::new();
    let home = env.home("synthetic");
    let h = s(&home);
    let (_, v) = env.run(&["migrate-plan", h], &request(&[hash(CONFIG)], None));
    let plan = v["result"]["plan"].clone();
    let d = v["result"]["plan_digest"].as_str().unwrap().to_string();
    let before = digest(&home);
    let apply = |digest_arg: &str, reference: &str, body: &Value| {
        env.run(
            &[
                "migrate-apply",
                h,
                "--plan-digest",
                digest_arg,
                "--permission-reference",
                reference,
            ],
            body,
        )
    };

    let (code, v) = apply(&"0".repeat(64), "r", &plan);
    assert_ne!(code, 0);
    assert_eq!(v["error"]["kind"], "stale_plan", "{v}");

    let mut altered = plan.clone();
    altered["mode"] = json!("Automatic");
    let (_, v) = apply(&d, "r", &altered);
    assert_eq!(v["error"]["kind"], "stale_plan", "{v}");

    // A plan without the release list (library rehearsal shape) is not accepted.
    let mut no_release = plan.clone();
    no_release.as_object_mut().unwrap().remove("go_release");
    let (_, v) = apply(&d, "r", &no_release);
    assert_eq!(v["error"]["kind"], "stale_plan", "{v}");
    // Even one whose digest is recomputed to be self-consistent without the list.
    let mut forged: moai_cli::migrate::Plan = serde_json::from_value(plan.clone()).unwrap();
    forged.go_release = None;
    forged.digest.clear();
    forged.digest = hash(&serde_json::to_vec(&forged).unwrap());
    let forged_digest = forged.digest.clone();
    let (_, v) = apply(&forged_digest, "r", &serde_json::to_value(&forged).unwrap());
    assert_eq!(v["error"]["kind"], "stale_plan", "{v}");

    for reference in ["", "placeholder", "null"] {
        let (_, v) = apply(&d, reference, &plan);
        assert_eq!(v["error"]["kind"], "permission_reference_missing", "{v}");
    }

    // The source changed after the preview: the recompute differs.
    fs::write(
        home.join(".moai/config/language.json"),
        b"{\"language\":\"en\"}",
    )
    .unwrap();
    let changed = digest(&home);
    let (_, v) = apply(&d, "r", &plan);
    assert_eq!(v["error"]["kind"], "stale_plan", "{v}");
    assert_eq!(digest(&home), changed);
    assert_ne!(before, changed);
    assert!(!home.join(".moai-go-backup").exists());
    assert!(!home.join(".moai-cli").exists());
}

#[test]
fn update_conflict_preview_shows_project_and_target_conflicts_and_blocks_apply() {
    let env = Env::new();
    let home = env.home("synthetic");
    let h = s(&home);
    fs::create_dir_all(home.join("project/.moai/config")).unwrap();
    fs::write(
        home.join("project/.moai/config/language.json"),
        b"{\"language\":\"en\"}",
    )
    .unwrap();
    let before = digest(&home);
    let (code, v) = env.run(
        &["migrate-plan", h],
        &request(&[hash(CONFIG)], Some("project/.moai/config/language.json")),
    );
    assert_eq!(code, 0, "{v}");
    assert_eq!(v["result"]["applicable"], false);
    assert_eq!(
        v["result"]["conflict_preview"],
        json!([{
            "source": ".moai/config/language.json",
            "target": ".moai-cli/imports/language.json",
            "project_current": "project/.moai/config/language.json",
            "reason": "project_differs",
        }])
    );
    let plan = v["result"]["plan"].clone();
    let d = v["result"]["plan_digest"].as_str().unwrap().to_string();
    let (code, v) = env.run(
        &[
            "migrate-apply",
            h,
            "--plan-digest",
            &d,
            "--permission-reference",
            "r",
        ],
        &plan,
    );
    assert_ne!(code, 0);
    assert_eq!(v["error"]["kind"], "conflict", "{v}");
    assert_eq!(digest(&home), before);
    // A project comparison outside config or SPEC is a refused request.
    let (code, v) = env.run(
        &["migrate-plan", h],
        &request(&[hash(CONFIG)], Some("project/notes.txt")),
    );
    assert_ne!(code, 0);
    assert_eq!(v["error"]["kind"], "invalid_input", "{v}");

    fs::create_dir_all(home.join(".moai-cli/imports")).unwrap();
    fs::write(home.join(".moai-cli/imports/language.json"), b"existing").unwrap();
    let (_, v) = env.run(&["migrate-plan", h], &request(&[hash(CONFIG)], None));
    assert_eq!(
        v["result"]["conflict_preview"][0]["reason"],
        "target_exists"
    );
}

#[test]
fn ownership_needs_the_release_list_and_unreleased_files_are_preserved() {
    let env = Env::new();
    let home = env.home("synthetic");
    let h = s(&home);
    let (_, v) = env.run(
        &["migrate-plan", h],
        &request(&[hash(b"other release file")], None),
    );
    assert_eq!(v["result"]["plan"]["entries"][0]["ownership"], "unreleased");
    assert_eq!(v["result"]["plan"]["entries"][0]["action"], "preserved");
    assert_eq!(v["result"]["applicable"], false);
    // The release list is required in the request.
    let mut missing = request(&[hash(CONFIG)], None);
    missing.as_object_mut().unwrap().remove("go_release");
    let (code, v) = env.run(&["migrate-plan", h], &missing);
    assert_ne!(code, 0);
    assert_eq!(v["error"]["kind"], "invalid_input", "{v}");
}

#[test]
fn onprem_host_old_glm_key_file_and_secret_sources_are_refused() {
    let env = Env::new();
    let home = env.home("synthetic");
    let h = s(&home);
    fs::write(home.join(".moai/.env.glm"), b"SYNTHETIC=1").unwrap();
    let before = digest(&home);
    let mut onprem = request(&[hash(CONFIG)], None);
    onprem["hosts"] = json!(["OnPrem"]);
    let (code, v) = env.run(&["migrate-plan", h], &onprem);
    assert_ne!(code, 0);
    assert_eq!(v["error"]["kind"], "invalid_input", "{v}");
    let mut dup = request(&[hash(CONFIG)], None);
    dup["hosts"] = json!(["Glm", "Glm"]);
    assert_ne!(env.run(&["migrate-plan", h], &dup).0, 0);
    for source in [
        ".moai/.env.glm",
        ".moai/config/token.json",
        ".moai/worktrees/x",
    ] {
        let mut r = request(&[hash(CONFIG)], None);
        r["selections"][0]["source"] = json!(source);
        let (code, v) = env.run(&["migrate-plan", h], &r);
        assert_ne!(code, 0, "{source}: {v}");
        assert_eq!(v["error"]["kind"], "invalid_input", "{source}: {v}");
        assert_eq!(v["error"]["retryable"], false, "{source}: {v}");
        assert!(!v.to_string().contains("SYNTHETIC=1"));
    }
    assert_eq!(digest(&home), before);
}

#[test]
fn a_home_that_is_not_a_prepared_synthetic_home_is_refused() {
    let env = Env::new();
    let req = request(&[hash(CONFIG)], None);

    // Under the child's TMPDIR but without the marker: a redirected TMPDIR admits nothing.
    let bare = env.tmp.path().join("bare");
    fs::create_dir_all(bare.join(".moai/config")).unwrap();
    fs::write(bare.join(".moai/config/language.json"), CONFIG).unwrap();
    let before = digest(&bare);
    let (code, v) = env.run(&["migrate-plan", s(&bare)], &req);
    assert_ne!(code, 0);
    assert!(
        v["error"]["message"].as_str().unwrap().contains("marker"),
        "{v}"
    );
    // A marker with other bytes, or a link to a valid marker, is not the marker.
    fs::write(bare.join(SYNTHETIC_MARKER), b"something else\n").unwrap();
    assert_ne!(env.run(&["migrate-plan", s(&bare)], &req).0, 0);
    #[cfg(unix)]
    {
        let good = env.home("good");
        fs::remove_file(bare.join(SYNTHETIC_MARKER)).unwrap();
        std::os::unix::fs::symlink(good.join(SYNTHETIC_MARKER), bare.join(SYNTHETIC_MARKER))
            .unwrap();
        assert_ne!(env.run(&["migrate-plan", s(&bare)], &req).0, 0);
        fs::remove_file(bare.join(SYNTHETIC_MARKER)).unwrap();
    }
    assert_eq!(digest(&bare), before);

    // Marked but outside the child's temporary directory.
    let outside = env.user.path().join("outside");
    fs::create_dir_all(outside.join(".moai/config")).unwrap();
    fs::write(outside.join(".moai/config/language.json"), CONFIG).unwrap();
    fs::write(outside.join(SYNTHETIC_MARKER), SYNTHETIC_MARKER_TEXT).unwrap();
    let narrow_tmp = env.tmp.path().join("narrow");
    fs::create_dir_all(&narrow_tmp).unwrap();
    let (code, v) = env.run_with(
        &["migrate-plan", s(&outside)],
        &req,
        env.user.path(),
        &narrow_tmp,
    );
    assert_ne!(code, 0);
    assert_eq!(v["error"]["kind"], "invalid_input", "{v}");
    assert_eq!(v["error"]["retryable"], false, "{v}");

    // Marked, under TMPDIR, but it is the child's HOME, or an ancestor of it.
    let marked = env.home("as-home");
    let (code, v) = env.run_with(&["migrate-plan", s(&marked)], &req, &marked, env.tmp.path());
    assert_ne!(code, 0);
    assert!(
        v["error"]["message"].as_str().unwrap().contains("HOME"),
        "{v}"
    );
    let inner = marked.join("user");
    fs::create_dir_all(&inner).unwrap();
    let (code, _) = env.run_with(&["migrate-plan", s(&marked)], &req, &inner, env.tmp.path());
    assert_ne!(code, 0);
    // The same marked home is admitted when HOME is elsewhere (control).
    let (code, v) = env.run(&["migrate-plan", s(&marked)], &req);
    assert_eq!(code, 0, "{v}");
}

/// #2449 audit F3: the release-list bound is reachable through `migrate-plan` inside the
/// 64 KiB request: 768 distinct hashes are admitted, 769 are refused by the list check
/// (not by the input size). Literal values, not `RELEASE_HASH_LIMIT`.
#[test]
fn the_release_list_bound_is_reachable_through_the_cli() {
    let env = Env::new();
    let home = env.home("synthetic");
    let h = s(&home);
    let mut list: Vec<String> = vec![hash(CONFIG)];
    list.extend((0..768u32).map(|i| hash(&i.to_le_bytes())));
    assert_eq!(list.len(), 769);

    let at = request(&list[..768], None);
    assert!(serde_json::to_vec(&at).unwrap().len() <= 65536);
    let (code, v) = env.run(&["migrate-plan", h], &at);
    assert_eq!(code, 0, "{v}");
    assert_eq!(v["result"]["plan"]["entries"][0]["ownership"], "owned");

    let over = request(&list, None);
    assert!(serde_json::to_vec(&over).unwrap().len() <= 65536);
    let (code, v) = env.run(&["migrate-plan", h], &over);
    assert_ne!(code, 0);
    assert_eq!(v["error"]["kind"], "invalid_input", "{v}");
    assert!(
        v["error"]["message"]
            .as_str()
            .unwrap()
            .contains("bounded Go release hash list"),
        "{v}"
    );
}
