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
    os::unix::fs::PermissionsExt,
    path::PathBuf,
    process::{Command, Stdio},
};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;
const REF: &str = "test:codex-conflict-choices";
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
        self.config_plan_for(&want(), replace)
    }

    /// [`Fx::config_plan`] for another request.
    fn config_plan_for(&self, request: &Value, replace: bool) -> Value {
        let mut args = vec!["codex-config-plan", self.p()];
        if replace {
            args.push("--replace-with-backup");
        }
        self.call(&args, request.to_string().as_bytes())
    }

    /// `codex-config-apply` of `digest`, with or without the replace flag.
    fn config_apply(&self, digest: &str, replace: bool) -> (i32, Value) {
        self.config_apply_for(&want(), digest, replace)
    }

    /// [`Fx::config_apply`] for another request.
    fn config_apply_for(&self, request: &Value, digest: &str, replace: bool) -> (i32, Value) {
        let mut args = vec!["codex-config-apply", self.p()];
        if replace {
            args.push("--replace-with-backup");
        }
        args.extend(["--plan-digest", digest, "--permission-reference", REF]);
        self.run(&args, request.to_string().as_bytes())
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
}

fn want() -> Value {
    json!({"tui.status_line": "moai"})
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
    // The value MoAI would write is the one the replace plan says the key holds afterwards.
    let replace = fx.config_plan(true);
    assert!(
        diff[0]["desired"].as_array().is_some_and(|d| !d.is_empty()),
        "{plan}"
    );
    assert_eq!(
        diff[0]["desired"], replace["status_line"]["after"],
        "{plan} / {replace}"
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

/// AC-1 (REQ-8): without a conflict the flag changes nothing: same action, same digest, no
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

/// AC-2 (REQ-4, REQ-6, REQ-12): the replace apply backs the original up byte for byte, sets the
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
    // Private whatever the original's mode was: a config can hold secrets.
    let mode = fs::metadata(fx.root.join(&path))
        .unwrap()
        .permissions()
        .mode();
    assert_eq!(mode & 0o777, 0o600, "backup mode {mode:o}");
    // Everything before the replaced value is the user's text, byte for byte.
    let now = String::from_utf8(fx.read(CONFIG)).unwrap();
    let kept = "# mine\nmodel = \"x\" # keep\n\n[tui]\nstatus_line = [";
    assert!(now.starts_with(kept), "{now}");
    assert!(USER_CONFIG.starts_with(kept), "{USER_CONFIG}");
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

/// AC-2 (REQ-7): without the flag a conflict is refused and nothing is written, as before.
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

/// AC-2 (REQ-4, REQ-6): a backup of the same bytes left by an interrupted run is reused, so the apply
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

/// AC-2 (REQ-8): the flag on a config without a conflict applies as usual and writes no backup.
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

/// AC-2 (REQ-1, REQ-4): every conflicting key is listed and replaced together, in one backup.
#[test]
fn apply_replaces_every_conflicting_key_with_one_backup() {
    let user = "project_doc_max_bytes = 65536 # mine\n\n[tui]\nstatus_line = [\"model\"]\n";
    let fx = Fx::new();
    fx.write(CONFIG, user);
    let two = json!({"tui.status_line": "moai", "project_doc_max_bytes": 32768});
    let plain = fx.config_plan_for(&two, false);
    let keys: Vec<&str> = plain["diff"]
        .as_array()
        .unwrap()
        .iter()
        .map(|d| d["key"].as_str().unwrap())
        .collect();
    assert_eq!(
        keys,
        ["project_doc_max_bytes", "tui.status_line"],
        "{plain}"
    );
    let plan = fx.config_plan_for(&two, true);
    assert_eq!(
        plan["replaced"],
        json!(["project_doc_max_bytes", "tui.status_line"]),
        "{plan}"
    );
    let (code, out) = fx.config_apply_for(&two, &digest(&plan), true);
    assert_eq!(code, 0, "{out}");
    assert_eq!(fx.read(&backup_rel(&plan)), user.as_bytes());
    let now = String::from_utf8(fx.read(CONFIG)).unwrap();
    assert!(now.contains("project_doc_max_bytes = 32768"), "{now}");
    assert!(!now.contains("65536"), "{now}");
    assert_eq!(fx.config_plan_for(&two, false)["action"], "unchanged");
}

/// AC-2 (REQ-5): a backup folder that is a link is refused for a replacement, and a plain plan,
/// which keeps no copy, is not blocked by it.
#[test]
fn a_linked_backup_folder_blocks_the_replacement_only() {
    let fx = Fx::with_user_config();
    let rel = backup_rel(&fx.config_plan(true));
    let folder = std::path::Path::new(&rel).parent().unwrap().to_path_buf();
    let elsewhere = fx.root.join("elsewhere");
    fs::create_dir_all(&elsewhere).unwrap();
    fs::create_dir_all(fx.root.join(folder.parent().unwrap())).unwrap();
    std::os::unix::fs::symlink(&elsewhere, fx.root.join(&folder)).unwrap();

    assert_eq!(fx.config_plan(false)["action"], "conflict");
    let (code, out) = fx.run(
        &["codex-config-plan", fx.p(), "--replace-with-backup"],
        want().to_string().as_bytes(),
    );
    assert_eq!((code, &out["ok"]), (1, &json!(false)), "{out}");
    assert_eq!(
        fs::read_dir(&elsewhere).unwrap().count(),
        0,
        "a copy went through the link"
    );
    assert_eq!(fx.read(CONFIG), USER_CONFIG.as_bytes());
}
