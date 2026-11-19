//! SPEC-20261120-hooks-conflict-choices: a `.codex/hooks.json` or `.omp/extensions/moai-hooks.ts`
//! that no receipt matches comes with `keep`, `diff` and `replace` (a replacement that backs the
//! file up first). Every case runs the built binary in a temporary project with an empty temporary
//! HOME; the real HOME, PATH and checkout are never read or written.

#![cfg(unix)]

mod common;
use common::CleanEnv;
use serde_json::{json, Value};
use std::{
    fs,
    os::unix::fs::PermissionsExt,
    path::{Path, PathBuf},
    process::{Command, Stdio},
};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;
const REF: &str = "test:hooks-conflict-choices";
const PLACEHOLDER: &str = "<permission-reference>";
const SYSTEM_PATH: &str = "/usr/bin:/bin";
const RUST_MOAI: &str = "moai 1.2.3 (rust)";
const GO_MOAI: &str = "moai-adk v3.2.1-rc.6";

/// One host: its hook file, the file's extension and a file of the user's own.
#[derive(Clone, Copy)]
struct H {
    name: &'static str,
    step: &'static str,
    file: &'static str,
    ext: &'static str,
    user: &'static str,
}

const CODEX: H = H {
    name: "codex",
    step: "codex-hooks",
    file: ".codex/hooks.json",
    ext: "json",
    user: "{\n  \"hooks\": {\n    \"Stop\": [\n      {\n        \"hooks\": [{\"type\": \"command\", \"command\": \"echo mine\"}]\n      }\n    ]\n  }\n}\n",
};
const OMP: H = H {
    name: "omp",
    step: "omp-hooks",
    file: ".omp/extensions/moai-hooks.ts",
    ext: "ts",
    user: "// mine\nexport default function () {}\n",
};
const BOTH: [H; 2] = [CODEX, OMP];
const BACKUPS: &str = ".moai-cli/backups";

struct Fx {
    _tmp: TempDir,
    base: PathBuf,
    home: PathBuf,
    root: PathBuf,
}

impl Fx {
    fn new() -> Fx {
        let tmp = TempDir::new().unwrap();
        let base = fs::canonicalize(tmp.path()).unwrap();
        let (home, root) = (base.join("home"), base.join("project"));
        fs::create_dir_all(&home).unwrap();
        fs::create_dir_all(&root).unwrap();
        for (dir, body) in [("rust", RUST_MOAI), ("go", GO_MOAI)] {
            let d = base.join(dir);
            fs::create_dir_all(&d).unwrap();
            common::write_executable(
                &d.join("moai"),
                &format!("#!/bin/sh\nprintf '%s\\n' '{body}'\n"),
            );
        }
        Fx {
            _tmp: tmp,
            base,
            home,
            root,
        }
    }

    /// A project whose own hook file for `h` is already there and owned by no receipt.
    fn with_user_file(h: H) -> Fx {
        let fx = Fx::new();
        fx.write(h.file, h.user);
        fx
    }

    fn p(&self) -> &str {
        self.root.to_str().unwrap()
    }

    fn write(&self, rel: &str, body: &str) {
        let p = self.root.join(rel);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        fs::write(p, body).unwrap();
    }

    fn read(&self, rel: &str) -> Vec<u8> {
        fs::read(self.root.join(rel)).unwrap()
    }

    fn exists(&self, rel: &str) -> bool {
        fs::symlink_metadata(self.root.join(rel)).is_ok()
    }

    /// The PATH whose first `moai` is the Rust build (`rust`) or the earlier build (`go`).
    fn path_with(&self, dir: &str) -> String {
        format!("{}:{SYSTEM_PATH}", self.base.join(dir).display())
    }

    /// Runs the binary with `stdin` and PATH `path`; returns the exit code and the envelope.
    fn run_on(&self, args: &[&str], stdin: &[u8], path: &str) -> (i32, Value) {
        let mut child = Command::new(EXE)
            .args(args)
            .clean_env()
            .env("HOME", &self.home)
            .env("PATH", path)
            .env("LC_ALL", "C")
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
            .unwrap();
        common::feed_stdin(&mut child, stdin);
        let out = child.wait_with_output().unwrap();
        let value = serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
            panic!(
                "{args:?}: {}{}",
                String::from_utf8_lossy(&out.stdout),
                String::from_utf8_lossy(&out.stderr)
            )
        });
        (out.status.code().unwrap(), value)
    }

    fn run(&self, args: &[&str], stdin: &[u8]) -> (i32, Value) {
        self.run_on(args, stdin, SYSTEM_PATH)
    }

    fn call(&self, args: &[&str], stdin: &[u8]) -> Value {
        let (code, v) = self.run(args, stdin);
        assert_eq!((code, &v["ok"]), (0, &json!(true)), "{args:?}: {v}");
        v["result"].clone()
    }

    /// `hooks-plan` for `h`; the whole envelope.
    fn plan_env(&self, h: H, replace: bool, extra: &[&str], path: &str) -> (i32, Value) {
        let mut args = vec!["hooks-plan", self.p(), "--host", h.name];
        if replace {
            args.push("--replace-with-backup");
        }
        args.extend(extra);
        self.run_on(&args, b"", path)
    }

    fn plan(&self, h: H, replace: bool) -> Value {
        let (code, v) = self.plan_env(h, replace, &[], SYSTEM_PATH);
        assert_eq!(
            (code, &v["ok"]),
            (0, &json!(true)),
            "{h:?} {replace}: {v}",
            h = h.name
        );
        v["result"].clone()
    }

    /// `hooks-apply` of `digest` for `h`; the whole envelope.
    fn apply_env(
        &self,
        h: H,
        digest: &str,
        replace: bool,
        extra: &[&str],
        path: &str,
    ) -> (i32, Value) {
        let mut args = vec!["hooks-apply", self.p(), "--host", h.name];
        if replace {
            args.push("--replace-with-backup");
        }
        args.extend(extra);
        args.extend(["--plan-digest", digest, "--permission-reference", REF]);
        self.run_on(&args, b"", path)
    }

    fn apply(&self, h: H, digest: &str, replace: bool) -> (i32, Value) {
        self.apply_env(h, digest, replace, &[], SYSTEM_PATH)
    }

    /// Installs MoAI's own hook set for `h` the plain way, so a receipt owns the file.
    fn install(&self, h: H) {
        let plan = self.plan(h, false);
        let (code, v) = self.apply(h, &digest(&plan), false);
        assert_eq!(code, 0, "{v}");
    }

    /// Every file under the project and HOME with its bytes, in path order.
    fn tree(&self) -> Vec<(PathBuf, Vec<u8>)> {
        let mut out = Vec::new();
        let mut stack = vec![self.root.clone(), self.home.clone()];
        while let Some(d) = stack.pop() {
            for e in fs::read_dir(&d).unwrap() {
                let p = e.unwrap().path();
                if p.is_dir() {
                    stack.push(p);
                } else {
                    out.push((p.clone(), fs::read(&p).unwrap()));
                }
            }
        }
        out.sort();
        out
    }

    /// The backups in the state folder: file names with their bytes, in name order.
    fn backups(&self) -> Vec<(String, Vec<u8>)> {
        let dir = self.root.join(BACKUPS);
        let Ok(rd) = fs::read_dir(&dir) else {
            return Vec::new();
        };
        let mut out: Vec<(String, Vec<u8>)> = rd
            .map(|e| {
                let p = e.unwrap().path();
                (
                    p.file_name().unwrap().to_str().unwrap().to_string(),
                    fs::read(&p).unwrap(),
                )
            })
            .collect();
        out.sort();
        out
    }

    /// A profile written the way PROJECT writes it, listing the Codex and omp hosts.
    fn with_profile(&self) {
        let detect = self.call(&["project-detect", self.p()], b"");
        let profile = json!({
            "schema_version": 2, "project_id": "demo", "purpose": "Build a small editor",
            "tools": ["Rust"], "development_methods": ["TDD"],
            "hosts": ["claude", "codex", "omp"],
            "detected": detect["detected"],
        });
        let plan = self.call(
            &["project-plan", self.p()],
            &serde_json::to_vec(&profile).unwrap(),
        );
        let digest = plan["plan_digest"].as_str().unwrap().to_string();
        self.call(
            &[
                "project-apply",
                self.p(),
                "--plan-digest",
                &digest,
                "--permission-reference",
                REF,
            ],
            &serde_json::to_vec(&plan).unwrap(),
        );
    }

    /// `project-update-plan`, with the whole tree compared before and after.
    fn update_plan(&self) -> Value {
        let before = self.tree();
        let plan = self.call(&["project-update-plan", self.p()], b"");
        assert_eq!(before, self.tree(), "project-update-plan wrote something");
        plan
    }

    /// Runs one printed command (`{"argv": [...], "stdin": ...}`): `moai` is this binary and the
    /// placeholder is a permission reference.
    fn run_printed(&self, command: &Value) -> (i32, Value) {
        let argv: Vec<String> = command["argv"]
            .as_array()
            .unwrap_or_else(|| panic!("a choice with a command prints its argv: {command}"))
            .iter()
            .map(|w| w.as_str().unwrap().to_string())
            .collect();
        assert_eq!(argv[0], "moai", "{argv:?}");
        let args: Vec<&str> = argv[1..]
            .iter()
            .map(|w| if w == PLACEHOLDER { REF } else { w.as_str() })
            .collect();
        let stdin = match &command["stdin"] {
            Value::Null => Vec::new(),
            v => serde_json::to_vec(v).unwrap(),
        };
        self.run(&args, &stdin)
    }
}

fn digest(plan: &Value) -> String {
    plan["plan_digest"]
        .as_str()
        .unwrap_or_else(|| panic!("the plan has a digest: {plan}"))
        .to_string()
}

fn step<'a>(plan: &'a Value, name: &str) -> &'a Value {
    plan["phases"]
        .as_array()
        .unwrap()
        .iter()
        .flat_map(|p| p["steps"].as_array().into_iter().flatten())
        .find(|s| s["step"] == name)
        .unwrap_or_else(|| panic!("no step {name}: {plan}"))
}

fn backup_rel(plan: &Value) -> String {
    plan["backup"]["path"]
        .as_str()
        .unwrap_or_else(|| panic!("the plan names a backup path: {plan}"))
        .to_string()
}

fn mode(p: &Path) -> u32 {
    fs::metadata(p).unwrap().permissions().mode() & 0o777
}

fn ids(choices: &Value) -> Vec<&str> {
    choices
        .as_array()
        .unwrap_or_else(|| panic!("the step has choices: {choices}"))
        .iter()
        .map(|c| c["id"].as_str().unwrap())
        .collect()
}

// ---- plan -------------------------------------------------------------------------------

/// AC-6 (REQ-8): without the flag the plan still refuses a file no receipt matches, and writes
/// nothing.
#[test]
fn plan_without_the_flag_still_refuses_an_unowned_hook_file() {
    for h in BOTH {
        let fx = Fx::with_user_file(h);
        let before = fx.tree();
        let (code, v) = fx.plan_env(h, false, &[], SYSTEM_PATH);
        assert_ne!(code, 0, "{}: {v}", h.name);
        assert_eq!(v["error"]["kind"], json!("conflict"), "{v}");
        assert!(
            v.to_string().contains("requires preservation review"),
            "{v}"
        );
        assert_eq!(before, fx.tree(), "the plan wrote something");
    }
}

/// AC-1 (REQ-1, REQ-2): the replace plan reports `replace`, the file, the backup, a diff and a
/// digest, and writes nothing.
#[test]
fn plan_replace_reports_the_replacement_and_writes_nothing() {
    let fx = Fx::with_user_file(CODEX);
    let before = fx.tree();
    let plan = fx.plan(CODEX, true);
    assert_eq!(plan["action"], json!("replace"), "{plan}");
    assert_eq!(plan["path"], json!(CODEX.file), "{plan}");
    assert_eq!(plan["conflicts"], json!([]), "{plan}");
    assert_eq!(plan["replaced"], json!([CODEX.file]), "{plan}");
    assert_eq!(plan["writes"], json!(false), "{plan}");
    assert!(plan["planned_sha256"].is_string(), "{plan}");
    assert_ne!(plan["expected_sha256"], plan["planned_sha256"], "{plan}");
    let path = backup_rel(&plan);
    assert!(path.starts_with(BACKUPS), "{path}");
    // The copy is named by the bytes it holds: the first 16 hex digits of the file's sha256.
    let hex = plan["expected_sha256"]
        .as_str()
        .unwrap()
        .trim_start_matches("sha256:");
    assert!(
        path.ends_with(&format!("codex-hooks-{}.json", &hex[..16])),
        "{path}"
    );
    assert_eq!(plan["backup"]["sha256"], plan["expected_sha256"], "{plan}");
    assert_eq!(plan["backup"]["bytes"], json!(CODEX.user.len()), "{plan}");
    assert!(!fx.exists(&path), "the plan wrote the backup");
    let diff = &plan["diff"];
    assert_eq!(diff["path"], json!(CODEX.file), "{plan}");
    assert_eq!(diff["current"]["bytes"], json!(CODEX.user.len()), "{plan}");
    assert!(
        diff["removed"]
            .as_array()
            .is_some_and(|r| r.iter().any(|l| l.as_str().unwrap().contains("echo mine"))),
        "the lines only the file has are listed: {plan}"
    );
    assert!(
        diff["added"].as_array().is_some_and(|a| !a.is_empty()),
        "{plan}"
    );
    assert_eq!(before, fx.tree(), "the plan wrote something");
}

/// AC-1 (REQ-1): omp takes the same path and names a `.ts` backup.
#[test]
fn plan_replace_for_omp_names_the_ts_backup() {
    let fx = Fx::with_user_file(OMP);
    let plan = fx.plan(OMP, true);
    assert_eq!(plan["action"], json!("replace"), "{plan}");
    assert_eq!(plan["replaced"], json!([OMP.file]), "{plan}");
    let path = backup_rel(&plan);
    let hex = plan["expected_sha256"]
        .as_str()
        .unwrap()
        .trim_start_matches("sha256:");
    assert!(
        path.ends_with(&format!("omp-hooks-{}.ts", &hex[..16])),
        "{path}"
    );
}

/// AC-1 (REQ-1): a file MoAI wrote and the user then edited is a replacement too.
#[test]
fn plan_replace_after_an_edit_of_an_owned_file() {
    for h in BOTH {
        let fx = Fx::new();
        fx.install(h);
        fx.write(h.file, h.user);
        let (code, v) = fx.plan_env(h, false, &[], SYSTEM_PATH);
        assert_ne!(code, 0, "{}: {v}", h.name);
        let plan = fx.plan(h, true);
        assert_eq!(plan["action"], json!("replace"), "{plan}");
        assert_eq!(plan["backup"]["bytes"], json!(h.user.len()), "{plan}");
    }
}

/// AC-1 (REQ-9): with no file to replace the flag changes nothing: same action, same digest, no
/// backup, for an absent file and for an owned one.
#[test]
fn plan_replace_without_a_conflict_is_the_plain_plan() {
    for h in BOTH {
        let fx = Fx::new();
        let (plain, flagged) = (fx.plan(h, false), fx.plan(h, true));
        assert_eq!(plain["action"], json!("create"), "{plain}");
        assert_eq!(flagged, plain, "{}", h.name);
        fx.install(h);
        let (plain, flagged) = (fx.plan(h, false), fx.plan(h, true));
        assert_eq!(plain["action"], json!("unchanged"), "{plain}");
        assert_eq!(flagged, plain, "{}", h.name);
        for p in [&plain, &flagged] {
            assert!(
                p.get("backup").is_none() && p.get("replaced").is_none(),
                "{p}"
            );
        }
    }
}

/// AC-1 (REQ-10): the flag is for `codex` and `omp`; `claude` is refused and nothing is written.
#[test]
fn plan_replace_refuses_the_claude_host() {
    let fx = Fx::new();
    fx.write(".claude/settings.json", "{}\n");
    let before = fx.tree();
    let args = [
        "hooks-plan",
        fx.p(),
        "--host",
        "claude",
        "--replace-with-backup",
    ];
    let (code, v) = fx.run(&args, b"");
    assert_ne!(code, 0, "{v}");
    assert_eq!(v["error"]["kind"], json!("invalid_input"), "{v}");
    assert!(v.to_string().contains("--host codex or omp"), "{v}");
    let digest = "sha256:00";
    let args = [
        "hooks-apply",
        fx.p(),
        "--host",
        "claude",
        "--replace-with-backup",
        "--plan-digest",
        digest,
        "--permission-reference",
        REF,
    ];
    let (code, v) = fx.run(&args, b"");
    assert_ne!(code, 0, "{v}");
    assert_eq!(v["error"]["kind"], json!("invalid_input"), "{v}");
    assert_eq!(before, fx.tree());
}

/// AC-1 (REQ-2): a line shaped like a secret is not printed in the diff.
#[test]
fn plan_replace_withholds_a_secret_shaped_line() {
    let fx = Fx::new();
    // Built at run time so this file holds no secret-shaped text of its own.
    let secret = format!("{}{}", "pass", "word=hunter2hunter2");
    let body = format!("// mine\n// {secret}\nexport default function () {{}}\n");
    fx.write(OMP.file, &body);
    let plan = fx.plan(OMP, true);
    let all = plan.to_string();
    assert!(
        !all.contains("hunter2"),
        "the diff printed the secret: {all}"
    );
    assert!(
        plan["diff"]["removed"]
            .as_array()
            .unwrap()
            .iter()
            .any(|l| l.as_str().unwrap().contains("withheld")),
        "{plan}"
    );
    // The line before and the line after are still listed.
    assert!(all.contains("export default function"), "{all}");
}

/// AC-1 (REQ-2): at most 40 lines a side are listed, and the diff says it stopped.
#[test]
fn plan_replace_diff_stops_at_forty_lines() {
    let fx = Fx::new();
    let body: String = (0..100).map(|i| format!("// line {i:03}\n")).collect();
    fx.write(OMP.file, &body);
    let plan = fx.plan(OMP, true);
    let diff = &plan["diff"];
    assert_eq!(diff["removed"].as_array().unwrap().len(), 40, "{plan}");
    assert!(diff["added"].as_array().unwrap().len() <= 40, "{plan}");
    assert_eq!(diff["truncated"], json!(true), "{plan}");
    let small = Fx::with_user_file(OMP);
    assert_eq!(small.plan(OMP, true)["diff"]["truncated"], json!(false));
}

// ---- apply ------------------------------------------------------------------------------

/// AC-2 (REQ-3, REQ-7): the replace apply backs the original up byte for byte (mode 0600), writes
/// the MoAI file and the receipt, and the next plain plan has nothing to do.
#[test]
fn apply_replace_writes_backup_then_file_then_receipt_and_settles() {
    let fx = Fx::with_user_file(CODEX);
    let plan = fx.plan(CODEX, true);
    let path = backup_rel(&plan);
    let (code, out) = fx.apply(CODEX, &digest(&plan), true);
    assert_eq!(code, 0, "{out}");
    assert_eq!(
        fx.read(&path),
        CODEX.user.as_bytes(),
        "the backup is the original"
    );
    assert_eq!(mode(&fx.root.join(&path)), 0o600);
    assert_eq!(out["result"]["backup"]["path"], json!(path), "{out}");
    assert_eq!(out["result"]["state"], json!("applied"), "{out}");
    let file: Value = serde_json::from_slice(&fx.read(CODEX.file)).unwrap();
    assert!(
        file["hooks"].is_object() && file["description"].is_string(),
        "{file}"
    );
    let receipt: Value =
        serde_json::from_slice(&fx.read(".moai-cli/manifests/codex-hooks.json")).unwrap();
    assert_eq!(receipt["sha256"], plan["planned_sha256"], "{receipt}");
    assert!(receipt.get("pending_sha256").is_none(), "{receipt}");
    assert_eq!(fx.plan(CODEX, false)["action"], json!("unchanged"));
    assert_eq!(fx.backups().len(), 1);
}

/// AC-2 (REQ-3, REQ-7): the same for omp.
#[test]
fn apply_replace_for_omp_settles() {
    let fx = Fx::with_user_file(OMP);
    let plan = fx.plan(OMP, true);
    let path = backup_rel(&plan);
    let (code, out) = fx.apply(OMP, &digest(&plan), true);
    assert_eq!(code, 0, "{out}");
    assert_eq!(fx.read(&path), OMP.user.as_bytes());
    assert_eq!(mode(&fx.root.join(&path)), 0o600);
    assert_ne!(fx.read(OMP.file), OMP.user.as_bytes());
    assert_eq!(fx.plan(OMP, false)["action"], json!("unchanged"));
}

/// AC-2 (REQ-8): an apply without the flag refuses the digest of a replace plan as a conflict and
/// writes nothing.
#[test]
fn apply_without_the_flag_refuses_the_replace_digest() {
    for h in BOTH {
        let fx = Fx::with_user_file(h);
        let plan = fx.plan(h, true);
        let before = fx.tree();
        let (code, v) = fx.apply(h, &digest(&plan), false);
        assert_ne!(code, 0, "{}: {v}", h.name);
        assert_eq!(v["error"]["kind"], json!("conflict"), "{v}");
        assert_eq!(before, fx.tree(), "the refusal wrote something");
        assert!(fx.backups().is_empty());
    }
}

/// AC-2 (REQ-4): a digest of a file that changed since is refused before any backup exists.
#[test]
fn apply_refuses_a_stale_replace_digest_after_the_file_changed() {
    let fx = Fx::with_user_file(OMP);
    let plan = fx.plan(OMP, true);
    let changed = format!("{}// changed since the plan\n", OMP.user);
    fx.write(OMP.file, &changed);
    let before = fx.tree();
    let (code, v) = fx.apply(OMP, &digest(&plan), true);
    assert_ne!(code, 0, "{v}");
    assert!(v.to_string().contains("stale or altered plan"), "{v}");
    assert_eq!(before, fx.tree());
    assert!(
        fx.backups().is_empty(),
        "no backup before the digest is checked"
    );
}

/// AC-2 (REQ-5): a backup of the planned name holding other bytes is refused, and the hook file
/// and the receipt are not touched.
#[test]
fn apply_refuses_a_backup_name_that_holds_other_bytes() {
    let fx = Fx::with_user_file(CODEX);
    let plan = fx.plan(CODEX, true);
    let path = backup_rel(&plan);
    fx.write(&path, "not the bytes the plan names\n");
    let before = fx.tree();
    let (code, v) = fx.apply(CODEX, &digest(&plan), true);
    assert_ne!(code, 0, "{v}");
    assert_eq!(before, fx.tree(), "the refusal wrote something");
    assert_eq!(fx.read(CODEX.file), CODEX.user.as_bytes());
    assert!(!fx.exists(".moai-cli/manifests/codex-hooks.json"));
}

/// AC-2 (REQ-5): a backup of the same bytes already there is reused, not refused.
#[test]
fn apply_reuses_a_backup_of_the_same_bytes() {
    let fx = Fx::with_user_file(CODEX);
    let plan = fx.plan(CODEX, true);
    let path = backup_rel(&plan);
    fx.write(&path, CODEX.user);
    let (code, v) = fx.apply(CODEX, &digest(&plan), true);
    assert_eq!(code, 0, "{v}");
    assert_eq!(fx.backups().len(), 1);
    assert_ne!(fx.read(CODEX.file), CODEX.user.as_bytes());
}

/// AC-2 (REQ-9): with nothing to replace the flagged apply is the plain apply: no backup.
#[test]
fn apply_replace_without_a_conflict_writes_no_backup() {
    for h in BOTH {
        let fx = Fx::new();
        let plan = fx.plan(h, true);
        let (code, v) = fx.apply(h, &digest(&plan), true);
        assert_eq!(code, 0, "{}: {v}", h.name);
        assert!(!fx.exists(BACKUPS), "{}: a backup folder was made", h.name);
        assert_eq!(fx.plan(h, false)["action"], json!("unchanged"));
    }
}

/// AC-2 (REQ-7): two replacements in a row keep two backups, each the bytes it replaced.
#[test]
fn apply_two_replaces_in_a_row_keep_both_backups() {
    let fx = Fx::with_user_file(CODEX);
    let first = fx.plan(CODEX, true);
    assert_eq!(fx.apply(CODEX, &digest(&first), true).0, 0);
    let second_user = "{\n  \"hooks\": {},\n  \"note\": \"my second edit\"\n}\n";
    fx.write(CODEX.file, second_user);
    let second = fx.plan(CODEX, true);
    assert_eq!(fx.apply(CODEX, &digest(&second), true).0, 0);
    let backups = fx.backups();
    assert_eq!(backups.len(), 2, "{backups:?}");
    let held: Vec<&[u8]> = backups.iter().map(|(_, b)| b.as_slice()).collect();
    assert!(
        held.contains(&CODEX.user.as_bytes()),
        "the first original is kept"
    );
    assert!(
        held.contains(&second_user.as_bytes()),
        "the second original is kept"
    );
}

/// AC-2 (REQ-1, REQ-3): a file that already holds MoAI's bytes but has no receipt is a
/// replacement: the receipt adopts it, after a backup of the same bytes.
#[test]
fn apply_replace_of_identical_bytes_adopts_the_file() {
    let fx = Fx::new();
    fx.install(CODEX);
    let ours = fx.read(CODEX.file);
    fs::remove_file(fx.root.join(".moai-cli/manifests/codex-hooks.json")).unwrap();
    let (code, v) = fx.plan_env(CODEX, false, &[], SYSTEM_PATH);
    assert_ne!(code, 0, "{v}");
    let plan = fx.plan(CODEX, true);
    assert_eq!(plan["action"], json!("replace"), "{plan}");
    let (code, v) = fx.apply(CODEX, &digest(&plan), true);
    assert_eq!(code, 0, "{v}");
    let backups = fx.backups();
    assert_eq!(backups.len(), 1, "{backups:?}");
    assert_eq!(backups[0].1, ours);
    assert_eq!(fx.plan(CODEX, false)["action"], json!("unchanged"));
}

/// AC-2 (REQ-3): the lane Stop hook the plan was made with is what the replacement writes and
/// the receipt records.
#[test]
fn apply_replace_keeps_the_lane_the_receipt_records() {
    let fx = Fx::with_user_file(CODEX);
    let lane = ["--lane-loop", "agent-1", "--repo", "o/r"];
    let (code, v) = fx.plan_env(CODEX, true, &lane, SYSTEM_PATH);
    assert_eq!(code, 0, "{v}");
    let plan = v["result"].clone();
    assert_eq!(plan["action"], json!("replace"), "{plan}");
    let (code, v) = fx.apply_env(CODEX, &digest(&plan), true, &lane, SYSTEM_PATH);
    assert_eq!(code, 0, "{v}");
    let receipt: Value =
        serde_json::from_slice(&fx.read(".moai-cli/manifests/codex-hooks.json")).unwrap();
    assert_eq!(receipt["lane_loop"]["lane"], json!("agent-1"), "{receipt}");
    // A replace plan made without the lane has another digest: the lane is bound.
    let other = Fx::with_user_file(CODEX);
    assert_ne!(digest(&other.plan(CODEX, true)), digest(&plan));
}

// ---- linked backups ---------------------------------------------------------------------

/// AC-3 (REQ-6): a `backups` folder that is a link refuses the replace plan and the apply, and
/// writes nothing through it; a plain install that needs no copy is not blocked.
#[test]
fn linked_backup_folder_blocks_the_replacement_only() {
    let fx = Fx::with_user_file(CODEX);
    let elsewhere = fx.base.join("elsewhere");
    fs::create_dir_all(&elsewhere).unwrap();
    fs::create_dir_all(fx.root.join(".moai-cli")).unwrap();
    std::os::unix::fs::symlink(&elsewhere, fx.root.join(BACKUPS)).unwrap();
    let before = fx.tree();
    let (code, v) = fx.plan_env(CODEX, true, &[], SYSTEM_PATH);
    assert_ne!(code, 0, "{v}");
    assert!(v.to_string().contains("backups"), "{v}");
    let (code, v) = fx.apply(CODEX, "sha256:00", true);
    assert_ne!(code, 0, "{v}");
    assert_eq!(before, fx.tree());
    assert_eq!(
        fs::read_dir(&elsewhere).unwrap().count(),
        0,
        "written through the link"
    );

    let plain = Fx::new();
    fs::create_dir_all(plain.root.join(".moai-cli")).unwrap();
    std::os::unix::fs::symlink(&elsewhere, plain.root.join(BACKUPS)).unwrap();
    let plan = plain.plan(CODEX, false);
    assert_eq!(plain.apply(CODEX, &digest(&plan), false).0, 0);
}

// ---- gate -------------------------------------------------------------------------------

/// AC-5 (REQ-3): the gate that refuses a plain write while the first `moai` on PATH is the
/// earlier build refuses a replacement too, at plan time and at apply time, and no backup is made.
#[test]
fn gate_refuses_a_replace_when_moai_on_path_is_the_go_build() {
    let fx = Fx::with_user_file(CODEX);
    let (go, rust) = (fx.path_with("go"), fx.path_with("rust"));
    let (code, v) = fx.plan_env(CODEX, true, &[], &go);
    assert_eq!(code, 0, "{v}");
    let plan = v["result"].clone();
    assert_eq!(plan["gate"]["allowed"], json!(false), "{plan}");
    assert!(
        plan["gate"]["reason"]
            .as_str()
            .unwrap()
            .contains("wrong_generation"),
        "{plan}"
    );
    let (_, v) = fx.plan_env(CODEX, true, &[], &rust);
    assert_eq!(v["result"]["gate"]["allowed"], json!(true), "{v}");
    let before = fx.tree();
    let (code, v) = fx.apply_env(CODEX, &digest(&plan), true, &[], &go);
    assert_ne!(code, 0, "{v}");
    assert!(v.to_string().contains("wrong_generation"), "{v}");
    assert_eq!(before, fx.tree(), "the refusal wrote something");
    assert!(fx.backups().is_empty());
    let (code, v) = fx.apply_env(CODEX, &digest(&plan), true, &[], &rust);
    assert_eq!(code, 0, "{v}");
}

// ---- step -------------------------------------------------------------------------------

/// AC-4 (REQ-11): both conflicting hook steps list `keep`, `diff`, `replace` in that order:
/// `keep` recommended with no command, `diff` the read-only replace plan, `replace` the apply of
/// that plan's digest.
#[test]
fn step_lists_keep_diff_and_replace_in_that_order() {
    let fx = Fx::new();
    fx.with_profile();
    for h in BOTH {
        fx.write(h.file, h.user);
    }
    let plan = fx.update_plan();
    for h in BOTH {
        let s = step(&plan, h.step);
        assert_eq!(s["state"], json!("conflict"), "{s}");
        assert!(
            s["reason"]
                .as_str()
                .unwrap()
                .contains("preservation review"),
            "{s}"
        );
        assert!(s["plan_digest"].is_null() && s["apply"].is_null(), "{s}");
        let choices = &s["preview"]["choices"];
        assert_eq!(ids(choices), ["keep", "diff", "replace"], "{s}");
        let c = choices.as_array().unwrap();
        assert_eq!(c[0]["recommended"], json!(true), "{s}");
        assert!(c[0]["command"].is_null(), "{s}");
        assert_eq!(
            c.iter().filter(|x| x["recommended"] == json!(true)).count(),
            1,
            "{s}"
        );
        let writes: Vec<&str> = c
            .iter()
            .filter(|x| x["writes"] == json!(true))
            .map(|x| x["id"].as_str().unwrap())
            .collect();
        assert_eq!(writes, ["replace"], "{s}");
        let argv = |i: usize| -> Vec<&str> {
            c[i]["command"]["argv"]
                .as_array()
                .unwrap()
                .iter()
                .map(|w| w.as_str().unwrap())
                .collect()
        };
        let (diff, replace) = (argv(1), argv(2));
        assert_eq!(&diff[..2], ["moai", "hooks-plan"], "{diff:?}");
        assert!(
            diff.contains(&"--replace-with-backup") && diff.contains(&h.name),
            "{diff:?}"
        );
        assert_eq!(&replace[..2], ["moai", "hooks-apply"], "{replace:?}");
        assert!(replace.contains(&"--replace-with-backup") && replace.contains(&h.name));
        assert!(replace.contains(&PLACEHOLDER), "{replace:?}");
        let at = replace.iter().position(|w| *w == "--plan-digest").unwrap();
        let replace_plan = fx.plan(h, true);
        assert_eq!(replace[at + 1], digest(&replace_plan), "{s}");
        assert_eq!(c[2]["backup"], replace_plan["backup"]["path"], "{s}");
        // The two sides are in the step itself: reading the diff needs no command.
        assert!(s["preview"]["diff"]["removed"].is_array(), "{s}");
    }
}

/// AC-4 (REQ-11): the printed `diff` and `replace` commands run as printed; the `diff` writes
/// nothing, the `replace` settles the step.
#[test]
fn step_commands_run_as_printed_and_settle() {
    let fx = Fx::new();
    fx.with_profile();
    for h in BOTH {
        fx.write(h.file, h.user);
    }
    let plan = fx.update_plan();
    for h in BOTH {
        let choices = step(&plan, h.step)["preview"]["choices"].clone();
        let before = fx.tree();
        let (code, v) = fx.run_printed(&choices[1]["command"]);
        assert_eq!(code, 0, "{}: {v}", h.name);
        assert_eq!(v["result"]["action"], json!("replace"), "{v}");
        assert_eq!(before, fx.tree(), "the diff command wrote something");
        let (code, v) = fx.run_printed(&choices[2]["command"]);
        assert_eq!(code, 0, "{}: {v}", h.name);
        assert_eq!(v["result"]["state"], json!("applied"), "{v}");
    }
    assert_eq!(fx.backups().len(), 2);
    let after = fx.update_plan();
    for h in BOTH {
        let s = step(&after, h.step);
        assert_eq!(s["state"], json!("unchanged"), "{s}");
        assert!(s["preview"]["choices"].is_null(), "{s}");
    }
}

/// AC-4 (REQ-12): when the replace plan cannot be made (a linked `backups` folder) the step stays
/// a conflict with no choices, and the plain steps keep their shape.
#[test]
fn step_without_a_replace_plan_stays_a_bare_conflict() {
    let fx = Fx::new();
    fx.with_profile();
    fx.write(CODEX.file, CODEX.user);
    let elsewhere = fx.base.join("elsewhere");
    fs::create_dir_all(&elsewhere).unwrap();
    fs::create_dir_all(fx.root.join(".moai-cli")).unwrap();
    std::os::unix::fs::symlink(&elsewhere, fx.root.join(BACKUPS)).unwrap();
    let plan = fx.update_plan();
    let s = step(&plan, CODEX.step);
    assert_eq!(s["state"], json!("conflict"), "{s}");
    assert!(
        s["reason"]
            .as_str()
            .unwrap()
            .contains("preservation review"),
        "{s}"
    );
    assert!(s["preview"]["choices"].is_null(), "{s}");
    assert!(s["plan_digest"].is_null() && s["apply"].is_null(), "{s}");
    // The omp step has no file of the user's: it is a plain create with its own command.
    let omp = step(&plan, OMP.step);
    assert_eq!(omp["state"], json!("ready"), "{omp}");
    assert!(omp["apply"]["argv"].is_array(), "{omp}");
}
