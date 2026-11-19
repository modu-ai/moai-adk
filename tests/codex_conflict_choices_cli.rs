//! SPEC-20261119-codex-conflict-choices: a Codex config conflict comes with `keep`, `diff` and
//! `replace` (a replacement that backs the file up first). Every case runs the built binary in a
//! temporary project with an empty temporary HOME; the real HOME and checkout are never read or
//! written.

#![cfg(unix)]

mod common;
use common::CleanEnv;
use serde_json::{json, Value};
use std::{
    fs,
    path::PathBuf,
    process::{Command, Stdio},
};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;
const REF: &str = "test:codex-conflict-choices";
const PLACEHOLDER: &str = "<permission-reference>";
const CONFIG: &str = ".codex/config.toml";
/// The user's own file: a comment, an unrelated key with a trailing comment, their own status line.
const USER_CONFIG: &str = "# mine\nmodel = \"x\" # keep\n\n[tui]\nstatus_line = [\"model\"]\n";

struct Fx {
    _tmp: TempDir,
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
        Fx {
            _tmp: tmp,
            home,
            root,
        }
    }

    /// A project whose own Codex config already sets the status line.
    fn with_user_config() -> Fx {
        let fx = Fx::new();
        fx.write(CONFIG, USER_CONFIG);
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
        self.root.join(rel).exists()
    }

    /// Runs the binary with `stdin`; returns the exit code and the whole envelope.
    fn run(&self, args: &[&str], stdin: &[u8]) -> (i32, Value) {
        let mut child = Command::new(EXE)
            .args(args)
            .clean_env()
            .env("HOME", &self.home)
            .env("PATH", "/usr/bin:/bin")
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

    fn call(&self, args: &[&str], stdin: &[u8]) -> Value {
        let (code, v) = self.run(args, stdin);
        assert_eq!((code, &v["ok"]), (0, &json!(true)), "{args:?}: {v}");
        v["result"].clone()
    }

    /// `codex-config-plan`, with or without the replace flag.
    fn config_plan(&self, replace: bool) -> Value {
        let mut args = vec!["codex-config-plan", self.p()];
        if replace {
            args.push("--replace-with-backup");
        }
        self.call(&args, want().to_string().as_bytes())
    }

    /// `codex-config-apply` of `digest`, with or without the replace flag.
    fn config_apply(&self, digest: &str, replace: bool) -> (i32, Value) {
        let mut args = vec!["codex-config-apply", self.p()];
        if replace {
            args.push("--replace-with-backup");
        }
        args.extend(["--plan-digest", digest, "--permission-reference", REF]);
        self.run(&args, want().to_string().as_bytes())
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
                    let bytes = fs::read(&p).unwrap();
                    out.push((p, bytes));
                }
            }
        }
        out.sort();
        out
    }

    /// A profile written the way PROJECT writes it, listing the Codex host.
    fn with_profile(&self) {
        let detect = self.call(&["project-detect", self.p()], b"");
        let profile = json!({
            "schema_version": 2, "project_id": "demo", "purpose": "Build a small editor",
            "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude", "codex"],
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

fn want() -> Value {
    json!({"tui.status_line": "moai"})
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

fn digest(plan: &Value) -> String {
    plan["plan_digest"].as_str().unwrap().to_string()
}

fn backup_rel(plan: &Value) -> String {
    plan["backup"]["path"]
        .as_str()
        .unwrap_or_else(|| panic!("the plan names a backup path: {plan}"))
        .to_string()
}

// ---- plan -------------------------------------------------------------------------------

/// AC-1 (REQ-1): a conflicting key is listed with the value the file holds and the value MoAI
/// would write, and the plain plan still refuses to replace.
#[test]
fn plan_lists_each_conflicting_key_with_both_values() {
    let fx = Fx::with_user_config();
    let before = fx.tree();
    let plan = fx.config_plan(false);
    assert_eq!(plan["action"], "conflict", "{plan}");
    let diff = plan["diff"].as_array().unwrap_or_else(|| panic!("{plan}"));
    assert_eq!(diff.len(), 1, "{plan}");
    assert_eq!(diff[0]["key"], "tui.status_line", "{plan}");
    assert_eq!(diff[0]["current"], json!(["model"]), "{plan}");
    assert!(
        diff[0]["desired"].as_array().is_some_and(|d| !d.is_empty()),
        "{plan}"
    );
    assert!(
        plan.get("backup").is_none(),
        "a plain plan has no backup: {plan}"
    );
    assert_eq!(before, fx.tree(), "the plan wrote something");
}

/// AC-1 (REQ-2, REQ-3): the replace plan reports `replace`, the keys, the backup and a digest
/// of its own, and writes nothing.
#[test]
fn plan_with_the_flag_reports_replace_and_a_digest_of_its_own_and_writes_nothing() {
    let fx = Fx::with_user_config();
    let plain = fx.config_plan(false);
    let before = fx.tree();
    let replace = fx.config_plan(true);
    assert_eq!(replace["action"], "replace", "{replace}");
    assert_eq!(replace["replaced"], json!(["tui.status_line"]), "{replace}");
    assert_eq!(replace["conflicts"], json!([]), "{replace}");
    assert_eq!(replace["writes"], json!(false), "{replace}");
    assert_ne!(digest(&replace), digest(&plain), "{replace}");
    let path = backup_rel(&replace);
    assert!(path.contains("backups/"), "{path}");
    // The copy is named by the bytes it holds: the first 16 hex digits of the file's sha256.
    let sha = replace["expected_sha256"].as_str().unwrap();
    let hex = sha.trim_start_matches("sha256:");
    assert!(
        path.ends_with(&format!("codex-config-{}.toml", &hex[..16])),
        "{path}"
    );
    assert!(!fx.exists(&path), "the plan wrote the backup");
    assert_eq!(
        replace["backup"]["sha256"], replace["expected_sha256"],
        "the backup is the file as it is now: {replace}"
    );
    assert_eq!(before, fx.tree(), "the plan wrote something");
}

/// AC-1 (REQ-7): without a conflict the flag changes nothing: same action, same digest, no
/// backup.
#[test]
fn plan_with_the_flag_and_no_conflict_is_the_plain_plan() {
    let fx = Fx::new();
    let plain = fx.config_plan(false);
    let flagged = fx.config_plan(true);
    assert_eq!(plain["action"], "create", "{plain}");
    assert_eq!(flagged, plain);
    assert!(flagged.get("backup").is_none(), "{flagged}");
    assert!(flagged.get("replaced").is_none(), "{flagged}");
}

// ---- apply ------------------------------------------------------------------------------

/// AC-2 (REQ-4, REQ-5, REQ-10): the replace apply backs the original up byte for byte, sets the
/// key, keeps every other key and comment, and the next plan has nothing to do.
#[test]
fn apply_replace_backs_the_file_up_then_writes_and_settles() {
    let fx = Fx::with_user_config();
    let plan = fx.config_plan(true);
    let path = backup_rel(&plan);
    let (code, out) = fx.config_apply(&digest(&plan), true);
    assert_eq!(code, 0, "{out}");
    assert_eq!(
        fx.read(&path),
        USER_CONFIG.as_bytes(),
        "the backup is the original"
    );
    assert_eq!(out["result"]["backup"]["path"], json!(path), "{out}");
    let now = String::from_utf8(fx.read(CONFIG)).unwrap();
    assert!(now.contains("# mine\nmodel = \"x\" # keep\n"), "{now}");
    assert!(!now.contains("status_line = [\"model\"]"), "{now}");
    assert_ne!(now, USER_CONFIG);
    for flag in [false, true] {
        let again = fx.config_plan(flag);
        assert_eq!(again["action"], "unchanged", "flag {flag}: {again}");
    }
}

/// AC-2 (REQ-3): a digest taken from the plain plan cannot apply a replacement.
#[test]
fn apply_with_the_flag_refuses_the_plain_digest() {
    let fx = Fx::with_user_config();
    let plain = fx.config_plan(false);
    let before = fx.tree();
    let (code, out) = fx.config_apply(&digest(&plain), true);
    assert_eq!(code, 1, "{out}");
    assert_eq!(out["ok"], json!(false), "{out}");
    assert_eq!(before, fx.tree(), "a refused apply wrote something");
}

/// AC-2 (REQ-6): without the flag a conflict is refused and nothing is written, as before.
#[test]
fn apply_without_the_flag_refuses_a_conflict_and_writes_nothing() {
    let fx = Fx::with_user_config();
    let replace = fx.config_plan(true);
    let before = fx.tree();
    for digest in [digest(&fx.config_plan(false)), digest(&replace)] {
        let (code, out) = fx.config_apply(&digest, false);
        assert_eq!(code, 1, "{out}");
        assert_eq!(out["error"]["kind"], "conflict", "{out}");
        assert_eq!(before, fx.tree(), "a refused apply wrote something");
    }
}

/// AC-2 (REQ-5): a backup file of that name that holds other bytes is never overwritten, and the
/// config stays as it is.
#[test]
fn apply_refuses_a_backup_name_that_holds_other_bytes() {
    let fx = Fx::with_user_config();
    let plan = fx.config_plan(true);
    let path = backup_rel(&plan);
    fx.write(&path, "someone else's file\n");
    let (code, out) = fx.config_apply(&digest(&plan), true);
    assert_eq!(code, 1, "{out}");
    assert_eq!(fx.read(&path), b"someone else's file\n");
    assert_eq!(
        fx.read(CONFIG),
        USER_CONFIG.as_bytes(),
        "the config changed"
    );
}

/// AC-2 (REQ-5): a backup of the same bytes left by an interrupted run is reused, so the apply
/// can be run again.
#[test]
fn apply_reuses_a_backup_that_already_holds_the_same_bytes() {
    let fx = Fx::with_user_config();
    let plan = fx.config_plan(true);
    let path = backup_rel(&plan);
    fx.write(&path, USER_CONFIG);
    let (code, out) = fx.config_apply(&digest(&plan), true);
    assert_eq!(code, 0, "{out}");
    assert_eq!(fx.read(&path), USER_CONFIG.as_bytes());
}

/// AC-2 (REQ-7): the flag on a config without a conflict applies as usual and writes no backup.
#[test]
fn apply_with_the_flag_and_no_conflict_writes_no_backup() {
    let fx = Fx::new();
    let plan = fx.config_plan(true);
    let (code, out) = fx.config_apply(&digest(&plan), true);
    assert_eq!(code, 0, "{out}");
    assert!(out["result"].get("backup").is_none(), "{out}");
    assert!(
        !fx.exists(".moai-cli/backups"),
        "a backup folder appeared without a conflict"
    );
}

// ---- project-update-plan step -----------------------------------------------------------

fn choices(plan: &Value) -> Vec<Value> {
    step(plan, "codex-config")["preview"]["choices"]
        .as_array()
        .unwrap_or_else(|| panic!("the conflict step carries choices: {plan}"))
        .clone()
}

/// AC-3 (REQ-8, REQ-9): the conflict step lists `keep` (recommended, no command), `diff`
/// (read-only) and `replace` (the apply with the replace digest and the backup path).
#[test]
fn step_lists_keep_diff_and_replace_in_that_order() {
    let fx = Fx::with_user_config();
    fx.with_profile();
    let plan = fx.update_plan();
    let s = step(&plan, "codex-config");
    assert_eq!(s["state"], "conflict", "{s}");
    let list = choices(&plan);
    let ids: Vec<&str> = list.iter().map(|c| c["id"].as_str().unwrap()).collect();
    assert_eq!(ids, ["keep", "diff", "replace"], "{s}");
    assert_eq!(list[0]["recommended"], json!(true), "{s}");
    assert_eq!(list[1]["recommended"], json!(false), "{s}");
    assert_eq!(list[2]["recommended"], json!(false), "{s}");
    assert_eq!(list[0]["command"], Value::Null, "{s}");
    assert_eq!(list[0]["writes"], json!(false), "{s}");
    assert_eq!(list[1]["writes"], json!(false), "{s}");
    assert_eq!(list[2]["writes"], json!(true), "{s}");
    let diff = list[1]["command"]["argv"].as_array().unwrap();
    assert_eq!(
        diff.iter().map(|w| w.as_str().unwrap()).collect::<Vec<_>>(),
        ["moai", "codex-config-plan", fx.p(), "--replace-with-backup"],
        "{s}"
    );
    assert_eq!(list[1]["command"]["stdin"], want(), "{s}");
    assert!(
        list[2]["backup"].as_str().unwrap().contains("backups/"),
        "{s}"
    );
    assert!(
        s["apply"].is_null(),
        "a conflict step prints no apply of its own: {s}"
    );
}

/// AC-3 (REQ-9, REQ-10): the printed `diff` and `replace` commands run as printed, the digest
/// the step prints is the one the replace plan computes, and the project settles afterwards.
#[test]
fn step_printed_commands_run_as_printed() {
    let fx = Fx::with_user_config();
    fx.with_profile();
    let list = choices(&fx.update_plan());

    let (code, shown) = fx.run_printed(&list[1]["command"]);
    assert_eq!(code, 0, "{shown}");
    assert_eq!(shown["result"]["action"], "replace", "{shown}");
    let printed: Vec<&str> = list[2]["command"]["argv"]
        .as_array()
        .unwrap()
        .iter()
        .map(|w| w.as_str().unwrap())
        .collect();
    let at = printed.iter().position(|w| *w == "--plan-digest").unwrap();
    assert_eq!(
        printed[at + 1],
        shown["result"]["plan_digest"].as_str().unwrap(),
        "{printed:?}"
    );
    assert!(printed.contains(&"--replace-with-backup"), "{printed:?}");
    assert!(printed.contains(&PLACEHOLDER), "{printed:?}");

    let backup = list[2]["backup"].as_str().unwrap();
    assert!(!fx.exists(backup), "the diff command wrote the backup");
    let (code, applied) = fx.run_printed(&list[2]["command"]);
    assert_eq!(code, 0, "{applied}");
    assert_eq!(fx.read(backup), USER_CONFIG.as_bytes());

    let again = fx.update_plan();
    assert_eq!(
        step(&again, "codex-config")["state"],
        "unchanged",
        "{again}"
    );
    assert!(
        step(&again, "codex-config")["preview"]
            .get("choices")
            .is_none(),
        "a settled step offers no choices"
    );
}

/// AC-3 (REQ-8): only a conflict carries choices.
#[test]
fn step_without_a_conflict_carries_no_choices() {
    let fx = Fx::new();
    fx.with_profile();
    let plan = fx.update_plan();
    let s = step(&plan, "codex-config");
    assert_eq!(s["state"], "ready", "{s}");
    assert!(s["preview"].get("choices").is_none(), "{s}");
}
