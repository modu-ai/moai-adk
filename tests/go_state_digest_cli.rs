//! #2936 S-DIGEST (plan 6887264857 §3): `plan_digest` binds what `go-state-apply` acts on and
//! nothing else. A change to preview-only data between the preview and the apply (a new log
//! file, a new unrecorded file, a new `leave` row, an excluded secret) keeps the confirmed
//! digest valid; a change to anything the apply moves, copies or writes (a moving file's
//! bytes, its recorded `deployed_hash`, a restore, a state move, a root file, the Go backup
//! name) makes it stale, with the error kind `stale_plan`. Every case builds a temporary
//! project and HOME; the real checkout and HOME are never read.

#![cfg(unix)]

mod common;
use common::CleanEnv;
use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::{
    fs,
    os::unix::fs::symlink,
    path::PathBuf,
    process::{Command, Stdio},
};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;
const REF: &str = "test-permission-ref";
/// `go_state::FILE_LIMIT`.
const FILE_LIMIT: u64 = 64 * 1024 * 1024;

fn sha(bytes: &[u8]) -> String {
    format!("sha256:{:x}", Sha256::digest(bytes))
}

/// A Markdown document as people write it (the restore filter refuses one-line titles).
fn prose(title: &str) -> String {
    format!(
        "# {title}\n\n이번 주에 한 일을 적어 둔 글이에요.\n\n- 첫째 항목\n- 둘째 항목\n\n\
         ```\ncargo test\n```\n\n끝.\n"
    )
}

struct Fx {
    _tmp: TempDir,
    home: PathBuf,
    root: PathBuf,
    files: serde_json::Map<String, Value>,
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
            files: serde_json::Map::new(),
        }
    }

    fn p(&self) -> &str {
        self.root.to_str().unwrap()
    }

    fn write(&self, rel: &str, body: &str) {
        let p = self.root.join(rel);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        fs::write(p, body).unwrap();
    }

    fn deployed(&mut self, rel: &str, body: &str) {
        self.write(rel, body);
        self.record(rel, "template_managed", &sha(body.as_bytes()));
    }

    fn record(&mut self, rel: &str, provenance: &str, deployed: &str) {
        self.files.insert(
            rel.into(),
            json!({"provenance": provenance, "template_hash": deployed,
                   "deployed_hash": deployed, "current_hash": deployed}),
        );
    }

    fn manifest(&self) {
        self.write(
            ".moai/manifest.json",
            &json!({"version": "v3.2.1-rc.6", "deployed_at": "2026-11-01T00:00:00Z",
                    "files": self.files})
            .to_string(),
        );
    }

    /// Every entry under the project and HOME; links read as links.
    fn tree(&self) -> Vec<(PathBuf, Vec<u8>)> {
        let mut out = Vec::new();
        let mut stack = vec![self.root.clone(), self.home.clone()];
        while let Some(d) = stack.pop() {
            for e in fs::read_dir(&d).unwrap() {
                let p = e.unwrap().path();
                let m = fs::symlink_metadata(&p).unwrap();
                if m.file_type().is_symlink() {
                    let body = fs::read_link(&p).unwrap();
                    out.push((p, body.into_os_string().into_encoded_bytes()));
                } else if m.is_dir() {
                    stack.push(p);
                } else {
                    out.push((p.clone(), fs::read(&p).unwrap()));
                }
            }
        }
        out.sort();
        out
    }

    fn run_env(&self, args: &[&str], env: &[(&str, &str)]) -> Value {
        let mut cmd = Command::new(EXE);
        cmd.args(args)
            .clean_env()
            .env("HOME", &self.home)
            .env("PATH", "/usr/bin:/bin")
            .env("LC_ALL", "C")
            .stdin(Stdio::null());
        for (k, v) in env {
            cmd.env(k, v);
        }
        let out = cmd.output().unwrap();
        serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
            panic!(
                "{args:?}: {}{}",
                String::from_utf8_lossy(&out.stdout),
                String::from_utf8_lossy(&out.stderr)
            )
        })
    }

    fn plan_env(&self, env: &[(&str, &str)]) -> Value {
        let before = self.tree();
        let v = self.run_env(&["go-state-plan", self.p()], env);
        assert_eq!(v["ok"], json!(true), "{v}");
        assert_eq!(self.tree(), before, "a preview writes nothing");
        v["result"].clone()
    }

    fn plan(&self) -> Value {
        self.plan_env(&[])
    }

    fn apply_env(&self, digest: &str, env: &[(&str, &str)]) -> Value {
        self.run_env(
            &[
                "go-state-apply",
                self.p(),
                "--plan-digest",
                digest,
                "--permission-reference",
                REF,
            ],
            env,
        )
    }

    fn apply(&self, digest: &str) -> Value {
        self.apply_env(digest, &[])
    }

    fn revert_plan(&self, backup: &str) -> Value {
        self.run_env(&["go-state-revert-plan", self.p(), "--backup", backup], &[])
    }

    fn area(&self, backup: &str) -> PathBuf {
        self.root.join(".moai/backups").join(backup)
    }

    fn receipt(&self, backup: &str) -> Value {
        serde_json::from_slice(&fs::read(self.area(backup).join("go-state-receipt.json")).unwrap())
            .unwrap()
    }
}

fn item<'a>(plan: &'a Value, rel: &str) -> &'a Value {
    plan["items"]
        .as_array()
        .unwrap()
        .iter()
        .find(|i| i["path"] == rel)
        .unwrap_or_else(|| panic!("{rel} not in plan: {plan}"))
}

/// (action, class) of one item.
fn ac<'a>(plan: &'a Value, rel: &str) -> (&'a str, &'a str) {
    let i = item(plan, rel);
    (i["action"].as_str().unwrap(), i["class"].as_str().unwrap())
}

/// The Go folder moves; a unchanged tool file, an edited tool file, a SPEC to restore and an
/// earlier-root state entry come along.
fn base() -> Fx {
    let mut fx = Fx::new();
    fx.deployed(".moai/config/sections/a.yaml", "a: 1\n");
    fx.deployed(".claude/workflows/w.md", "# w\n");
    fx.deployed(".claude/rules/b.md", "b: 1\n");
    fx.write(".claude/rules/b.md", "b: 2 # edited\n");
    fx.write(".moai/specs/SPEC-1/spec.md", &prose("spec"));
    fx.write(".moai-cli/index/a.json", "{}");
    fx.manifest();
    fx
}

/// The preview of `fx` is the one the apply must accept: `digest` is the old preview's, the
/// fresh preview must have the same one, and the apply must go through.
fn assert_accepted(fx: &Fx, old: &Value) -> Value {
    let fresh = fx.plan();
    assert_eq!(
        fresh["plan_digest"], old["plan_digest"],
        "preview-only data changed the digest: {fresh}"
    );
    let v = fx.apply(old["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], json!(true), "{v}");
    assert_eq!(v["result"]["state"], json!("committed"), "{v}");
    v
}

/// The fresh preview differs, the apply with the old digest is refused as `stale_plan`, and
/// the tree is exactly as it was.
fn assert_refused(fx: &Fx, old: &Value) {
    let fresh = fx.plan();
    assert_ne!(
        fresh["plan_digest"], old["plan_digest"],
        "the digest ignored a change the apply acts on: {fresh}"
    );
    let before = fx.tree();
    let v = fx.apply(old["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], json!(false), "{v}");
    assert_eq!(v["error"]["kind"], json!("stale_plan"), "{v}");
    assert_eq!(fx.tree(), before, "a refused apply writes nothing");
}

// ---- D1, D2: preview-only data does not make the digest stale ----

#[test]
fn a_new_log_file_under_moai_keeps_the_digest_valid() {
    let fx = base();
    let old = fx.plan();
    fx.write(".moai/logs/x.log", "line\n");
    let fresh = fx.plan();
    assert_ne!(
        fresh["left_in_backup"], old["left_in_backup"],
        "the log file shows in the preview"
    );
    let v = assert_accepted(&fx, &old);
    assert_eq!(
        fs::read_to_string(fx.root.join(".moai-backup/logs/x.log")).unwrap(),
        "line\n",
        "{v}"
    );
}

#[test]
fn a_new_unrecorded_file_keeps_the_digest_valid() {
    let fx = base();
    let old = fx.plan();
    fx.write(".claude/notes.md", "notes\n");
    let fresh = fx.plan();
    assert_ne!(
        fresh["unrecorded"], old["unrecorded"],
        "the file shows in the preview"
    );
    assert_accepted(&fx, &old);
    assert!(fx.root.join(".claude/notes.md").is_file(), "never moved");
}

// ---- D3, D4: what a moving file is ----

#[test]
fn an_unchanged_moving_file_edited_after_the_preview_is_refused() {
    let fx = base();
    let old = fx.plan();
    fx.write(".claude/workflows/w.md", "# w, edited after the preview\n");
    assert_refused(&fx, &old);
}

/// An edited file's class, reason and proof do not change when it is edited again, so only
/// its bytes tell the preview from the apply.
#[test]
fn an_already_edited_moving_file_edited_again_after_the_preview_is_refused() {
    let fx = base();
    let old = fx.plan();
    fx.write(".claude/rules/b.md", "b: 3 # edited again\n");
    let fresh = fx.plan();
    assert_eq!(ac(&fresh, ".claude/rules/b.md"), ("move", "user_modified"));
    assert_eq!(
        item(&fresh, ".claude/rules/b.md")["reason"],
        item(&old, ".claude/rules/b.md")["reason"]
    );
    assert_eq!(
        item(&fresh, ".claude/rules/b.md")["proof"],
        item(&old, ".claude/rules/b.md")["proof"]
    );
    assert_refused(&fx, &old);
}

/// The manifest's `deployed_hash` of an edited file goes into the receipt; the item keeps
/// its class, reason and proof when only that hash changes.
#[test]
fn a_changed_deployed_hash_of_a_moving_file_is_refused() {
    let mut fx = base();
    let old = fx.plan();
    fx.record(".claude/rules/b.md", "template_managed", &sha(b"b: 0\n"));
    fx.manifest();
    let fresh = fx.plan();
    assert_eq!(ac(&fresh, ".claude/rules/b.md"), ("move", "user_modified"));
    assert_eq!(
        item(&fresh, ".claude/rules/b.md")["reason"],
        item(&old, ".claude/rules/b.md")["reason"]
    );
    assert_refused(&fx, &old);
}

// ---- D5, D6: the earlier root's entries ----

#[test]
fn a_new_state_entry_that_moves_is_refused_and_one_that_stays_is_not() {
    let fx = base();
    let old = fx.plan();
    fx.write(".moai-cli/decisions/d.json", "{}");
    assert_refused(&fx, &old);

    let fx = base();
    let old = fx.plan();
    fx.write(".moai-cli/other/o.txt", "mine\n");
    let fresh = fx.plan();
    let stays = fresh["state_moves"]
        .as_array()
        .unwrap()
        .iter()
        .find(|s| s["from"] == ".moai-cli/other")
        .unwrap_or_else(|| panic!("no state row for .moai-cli/other: {fresh}"));
    assert_eq!(stays["action"], json!("leave"), "{fresh}");
    assert_accepted(&fx, &old);
    assert!(fx.root.join(".moai-cli/other/o.txt").is_file());
}

// ---- D7: restores ----

#[test]
fn a_new_restore_copy_is_refused_and_an_excluded_secret_is_not() {
    let fx = base();
    let old = fx.plan();
    fx.write(".moai/specs/new.md", &prose("new"));
    assert_refused(&fx, &old);

    let fx = base();
    let old = fx.plan();
    fx.write(".moai/docs/.env", "TOKEN=placeholder\n");
    let fresh = fx.plan();
    let row = fresh["restores"]
        .as_array()
        .unwrap()
        .iter()
        .find(|r| r["from"] == ".moai-backup/docs/.env")
        .unwrap_or_else(|| panic!("no restore row for the secret: {fresh}"));
    assert_eq!(row["action"], json!("excluded_secret"), "{fresh}");
    assert_accepted(&fx, &old);
    assert!(!fx.root.join(".moai/docs/.env").exists(), "never copied");
}

// ---- D8: root files ----

#[test]
fn a_root_file_edited_or_added_after_the_preview_is_refused() {
    let fx = base();
    fx.write("CLAUDE.md", "# project rules\n");
    let old = fx.plan();
    fx.write("CLAUDE.md", "# project rules, edited\n");
    assert_refused(&fx, &old);

    let fx = base();
    fx.write("CLAUDE.md", "# project rules\n");
    let old = fx.plan();
    fx.write("AGENTS.local.md", "# mine\n");
    assert_refused(&fx, &old);
}

// ---- D9: the copy folder, the link and the Go backup name ----

#[test]
fn a_copy_folder_with_other_content_and_the_same_file_set_is_refused() {
    let mut fx = Fx::new();
    fx.deployed(".moai/config/a.yaml", "a\n");
    // The canonical file is recorded but absent, so only the copy folder carries the bytes.
    fx.record(
        ".claude/skills/moai-b/ref/x.md",
        "template_managed",
        &sha(b"x\n"),
    );
    fx.write(".agents/skills/moai-b/ref/x.md", "x\n");
    fx.manifest();
    let old = fx.plan();
    assert_eq!(
        ac(&old, ".agents/skills/moai-b"),
        ("move", "go_mirror_copy")
    );
    // The file and its record change together: the item stays a matching copy.
    fx.write(".agents/skills/moai-b/ref/x.md", "y\n");
    fx.record(
        ".claude/skills/moai-b/ref/x.md",
        "template_managed",
        &sha(b"y\n"),
    );
    fx.manifest();
    let fresh = fx.plan();
    assert_eq!(
        ac(&fresh, ".agents/skills/moai-b"),
        ("move", "go_mirror_copy")
    );
    assert_refused(&fx, &old);
}

#[test]
fn a_link_with_another_body_is_refused() {
    let mut fx = Fx::new();
    fx.deployed(".moai/config/a.yaml", "a\n");
    fx.deployed(".claude/skills/moai-a/SKILL.md", "# a\n");
    fx.deployed(".claude/skills/moai-a/ref/x.md", "x\n");
    fs::create_dir_all(fx.root.join(".agents/skills")).unwrap();
    symlink(
        "../../.claude/skills/moai-a",
        fx.root.join(".agents/skills/moai-a"),
    )
    .unwrap();
    fx.manifest();
    let old = fx.plan();
    assert_eq!(ac(&old, ".agents/skills/moai-a"), ("move", "go_mirror"));
    fs::remove_file(fx.root.join(".agents/skills/moai-a")).unwrap();
    symlink(
        "../../.claude/skills/moai-a/ref",
        fx.root.join(".agents/skills/moai-a"),
    )
    .unwrap();
    assert_refused(&fx, &old);
}

#[test]
fn a_go_backup_name_that_became_taken_is_refused() {
    let fx = base();
    let old = fx.plan();
    assert_eq!(old["go_backup"], json!(".moai-backup"), "{old}");
    fs::create_dir(fx.root.join(".moai-backup")).unwrap();
    let fresh = fx.plan();
    assert_eq!(fresh["go_backup"], json!(".moai-backup-<UTC>"), "{fresh}");
    assert_eq!(ac(&fresh, ".moai"), ("move", "go_folder"), "{fresh}");
    assert_refused(&fx, &old);
}

/// With nothing to restore, no state to move and no root file, the backup name is the only
/// bound datum that differs when `.moai-backup` appears after the preview.
#[test]
fn only_the_go_backup_name_changing_is_refused() {
    let mut fx = Fx::new();
    fx.deployed(".moai/config/a.yaml", "a\n");
    fx.deployed(".claude/workflows/w.md", "# w\n");
    fx.manifest();
    let old = fx.plan();
    assert_eq!(old["restores"], json!([]), "{old}");
    assert_eq!(old["state_moves"], json!([]), "{old}");
    assert_eq!(old["root_files"], json!([]), "{old}");
    fs::create_dir(fx.root.join(".moai-backup")).unwrap();
    let fresh = fx.plan();
    assert_eq!(fresh["go_backup"], json!(".moai-backup-<UTC>"), "{fresh}");
    assert_eq!(fresh["items"], old["items"], "{fresh}");
    assert_refused(&fx, &old);
}

// ---- the error contract (F17) ----

#[test]
fn a_wrong_digest_is_a_stale_plan_that_names_the_plan_command() {
    let fx = base();
    let p = fx.plan();
    for wrong in ["", "sha256:0000", "not-a-digest"] {
        let before = fx.tree();
        let v = fx.apply(wrong);
        assert_eq!(v["ok"], json!(false), "{wrong}: {v}");
        assert_eq!(v["error"]["kind"], json!("stale_plan"), "{wrong}: {v}");
        assert!(
            v["error"]["message"]
                .as_str()
                .unwrap()
                .contains("go-state-plan"),
            "{v}"
        );
        assert_eq!(
            v["error"]["next_actions"],
            json!(["go_state.plan"]),
            "{wrong}: {v}"
        );
        assert_eq!(fx.tree(), before, "{wrong}: nothing written");
    }
    let ok = fx.apply(p["plan_digest"].as_str().unwrap());
    assert_eq!(ok["ok"], json!(true), "{ok}");
}

#[test]
fn a_wrong_revert_digest_is_a_stale_plan_that_names_the_revert_plan_command() {
    let fx = base();
    let p = fx.plan();
    let v = fx.apply(p["plan_digest"].as_str().unwrap());
    let backup = v["result"]["backup"].as_str().unwrap().to_string();
    let rp = fx.revert_plan(&backup);
    assert_eq!(rp["ok"], json!(true), "{rp}");
    let before = fx.tree();
    let v = fx.run_env(
        &[
            "go-state-revert",
            fx.p(),
            "--backup",
            &backup,
            "--plan-digest",
            "sha256:0000",
            "--permission-reference",
            REF,
        ],
        &[],
    );
    assert_eq!(v["ok"], json!(false), "{v}");
    assert_eq!(v["error"]["kind"], json!("stale_plan"), "{v}");
    assert!(
        v["error"]["message"]
            .as_str()
            .unwrap()
            .contains("go-state-revert-plan"),
        "{v}"
    );
    assert_eq!(fx.tree(), before, "nothing written");
}

// ---- boundaries: N and N+1 of the enforced limits ----

/// A sparse file of `len` bytes, recorded as the Go deploy of `recorded_len` zero bytes.
fn zeros(fx: &mut Fx, rel: &str, len: u64, recorded_len: u64) {
    let p = fx.root.join(rel);
    fs::create_dir_all(p.parent().unwrap()).unwrap();
    fs::File::create(&p).unwrap().set_len(len).unwrap();
    let mut h = Sha256::new();
    let chunk = vec![0u8; 1 << 20];
    let mut left = recorded_len;
    while left > 0 {
        let n = left.min(chunk.len() as u64) as usize;
        h.update(&chunk[..n]);
        left -= n as u64;
    }
    fx.record(
        rel,
        "template_managed",
        &format!("sha256:{:x}", h.finalize()),
    );
}

fn sparse(fx: &Fx, rel: &str, len: u64) {
    let p = fx.root.join(rel);
    fs::create_dir_all(p.parent().unwrap()).unwrap();
    fs::File::create(&p).unwrap().set_len(len).unwrap();
}

/// A moving file of exactly `FILE_LIMIT` bytes is bound by its content; one byte more is a
/// `leave` row with nothing to bind; growing from N to N+1 after the preview is refused.
#[test]
fn a_moving_file_at_the_file_limit_is_bound_and_one_byte_more_is_not_a_move() {
    let mut fx = Fx::new();
    fx.deployed(".moai/config/a.yaml", "a\n");
    zeros(&mut fx, ".claude/big/at.bin", FILE_LIMIT, FILE_LIMIT);
    zeros(
        &mut fx,
        ".claude/big/over.bin",
        FILE_LIMIT + 1,
        FILE_LIMIT + 1,
    );
    fx.manifest();
    let old = fx.plan();
    assert_eq!(ac(&old, ".claude/big/at.bin"), ("move", "tool_asset"));
    assert_eq!(ac(&old, ".claude/big/over.bin").0, "leave");
    // N+1 stays a `leave` row: changing it does not touch the digest.
    sparse(&fx, ".claude/big/over.bin", FILE_LIMIT + 2);
    assert_eq!(fx.plan()["plan_digest"], old["plan_digest"]);
    // N becomes N+1: the move row is gone.
    sparse(&fx, ".claude/big/at.bin", FILE_LIMIT + 1);
    let fresh = fx.plan();
    assert_eq!(ac(&fresh, ".claude/big/at.bin").0, "leave", "{fresh}");
    assert_refused(&fx, &old);
}

#[test]
fn a_restore_at_the_file_limit_is_bound_and_one_byte_more_is_not_a_copy() {
    let fx = base();
    sparse(&fx, ".moai/docs/at.md", FILE_LIMIT);
    sparse(&fx, ".moai/docs/over.md", FILE_LIMIT + 1);
    let old = fx.plan();
    let action = |p: &Value, from: &str| {
        p["restores"]
            .as_array()
            .unwrap()
            .iter()
            .find(|r| r["from"] == from)
            .unwrap_or_else(|| panic!("{from} not in restores: {p}"))["action"]
            .clone()
    };
    assert_eq!(action(&old, ".moai-backup/docs/at.md"), json!("copy"));
    assert_eq!(
        action(&old, ".moai-backup/docs/over.md"),
        json!("excluded_unclassifiable")
    );
    // The over-sized row is preview-only.
    sparse(&fx, ".moai/docs/over.md", FILE_LIMIT + 2);
    assert_eq!(fx.plan()["plan_digest"], old["plan_digest"]);
    // N becomes N+1: the copy row is gone.
    sparse(&fx, ".moai/docs/at.md", FILE_LIMIT + 1);
    assert_eq!(
        action(&fx.plan(), ".moai-backup/docs/at.md"),
        json!("excluded_unclassifiable")
    );
    assert_refused(&fx, &old);
}

#[test]
fn a_root_file_at_the_file_limit_is_bound_and_one_byte_more_is_not_a_move() {
    let fx = base();
    sparse(&fx, "CLAUDE.md", FILE_LIMIT);
    sparse(&fx, "AGENTS.md", FILE_LIMIT + 1);
    let old = fx.plan();
    let action = |p: &Value, path: &str| {
        p["root_files"]
            .as_array()
            .unwrap()
            .iter()
            .find(|f| f["path"] == path)
            .unwrap_or_else(|| panic!("{path} not in root_files: {p}"))["action"]
            .clone()
    };
    assert_eq!(action(&old, "CLAUDE.md"), json!("move"));
    assert_eq!(action(&old, "AGENTS.md"), json!("leave"));
    sparse(&fx, "AGENTS.md", FILE_LIMIT + 2);
    assert_eq!(fx.plan()["plan_digest"], old["plan_digest"]);
    sparse(&fx, "CLAUDE.md", FILE_LIMIT + 1);
    assert_eq!(action(&fx.plan(), "CLAUDE.md"), json!("leave"));
    assert_refused(&fx, &old);
}

/// A copy folder of N files is bound by its content; the (N+1)th file turns the row into a
/// `leave`, and the apply with the old digest is refused under the same limit.
#[test]
fn a_copy_folder_at_the_mirror_file_limit_is_bound_and_one_more_file_is_not_a_move() {
    let mut fx = Fx::new();
    fx.deployed(".moai/config/a.yaml", "a\n");
    fx.deployed(".claude/skills/moai-a/SKILL.md", "# a\n");
    fx.deployed(".claude/skills/moai-a/ref/x.md", "x\n");
    fx.deployed(".claude/skills/moai-a/ref/y.md", "y\n");
    fx.write(".agents/skills/moai-a/SKILL.md", "# a\n");
    fx.write(".agents/skills/moai-a/ref/x.md", "x\n");
    fx.manifest();
    let env = [("MOAI_CLI_TEST_MIRROR_FILE_LIMIT", "2")];
    let old = fx.plan_env(&env);
    assert_eq!(
        ac(&old, ".agents/skills/moai-a"),
        ("move", "go_mirror_copy"),
        "{old}"
    );
    fx.write(".agents/skills/moai-a/ref/y.md", "y\n");
    let fresh = fx.plan_env(&env);
    assert_eq!(ac(&fresh, ".agents/skills/moai-a").0, "leave", "{fresh}");
    assert_ne!(fresh["plan_digest"], old["plan_digest"]);
    let before = fx.tree();
    let v = fx.apply_env(old["plan_digest"].as_str().unwrap(), &env);
    assert_eq!(v["ok"], json!(false), "{v}");
    assert_eq!(v["error"]["kind"], json!("stale_plan"), "{v}");
    assert_eq!(fx.tree(), before);
}

/// The unrecorded and secret counts stop at the entry limit and are preview-only: the
/// digest is the same at any limit that still lets the `.moai/` walk finish. A limit below
/// the `.moai/` walk keeps the folder in place, which the apply acts on.
#[test]
fn the_unrecorded_limit_changes_the_counts_but_not_the_digest_until_it_stops_the_folder() {
    let mut fx = Fx::new();
    fx.deployed(".moai/config/a.yaml", "a\n");
    for n in 0..12 {
        fx.write(&format!(".claude/extra/f{n}.md"), "e\n");
    }
    fx.write(".claude/.env", "not read\n");
    fx.manifest();
    // `.moai/` holds 3 entries (manifest.json, config, config/a.yaml).
    let wide = fx.plan();
    let narrow = fx.plan_env(&[("MOAI_CLI_TEST_UNRECORDED_LIMIT", "8")]);
    assert_eq!(ac(&narrow, ".moai"), ("move", "go_folder"), "{narrow}");
    assert_ne!(narrow["unrecorded"], wide["unrecorded"], "{narrow}");
    assert!(
        narrow["warnings"].to_string().contains("stopped after"),
        "{narrow}"
    );
    assert_eq!(narrow["plan_digest"], wide["plan_digest"], "{narrow}");
    // The apply under the narrow limit accepts the digest of the wide preview.
    let v = fx.apply_env(
        wide["plan_digest"].as_str().unwrap(),
        &[("MOAI_CLI_TEST_UNRECORDED_LIMIT", "8")],
    );
    assert_eq!(v["ok"], json!(true), "{v}");
    // Below the `.moai/` walk the folder stays: a different plan, a different digest.
    let fx = {
        let mut fx = Fx::new();
        fx.deployed(".moai/config/a.yaml", "a\n");
        fx.manifest();
        fx
    };
    let wide = fx.plan();
    let tiny = fx.plan_env(&[("MOAI_CLI_TEST_UNRECORDED_LIMIT", "2")]);
    assert_eq!(ac(&tiny, ".moai"), ("leave", "go_folder"), "{tiny}");
    assert_ne!(tiny["plan_digest"], wide["plan_digest"]);
    let before = fx.tree();
    let v = fx.apply_env(
        wide["plan_digest"].as_str().unwrap(),
        &[("MOAI_CLI_TEST_UNRECORDED_LIMIT", "2")],
    );
    assert_eq!(v["ok"], json!(false), "{v}");
    assert_eq!(v["error"]["kind"], json!("stale_plan"), "{v}");
    assert_eq!(fx.tree(), before);
}

// ---- the race and the link probe ----

/// A writer that changes a moving file after the apply's fresh preview and before the move
/// (debug-build fault seam): the move stops, the receipt does not mark the file moved, and
/// the edited file stays where it is.
#[test]
fn a_file_edited_after_the_fresh_preview_stops_the_move() {
    let fx = base();
    let p = fx.plan();
    let v = fx.apply_env(
        p["plan_digest"].as_str().unwrap(),
        &[("MOAI_CLI_TEST_FAULT", "go_state_edit_after_plan")],
    );
    assert_eq!(v["ok"], json!(false), "{v}");
    assert!(
        v.to_string()
            .contains("changed after the preview; move stopped"),
        "{v}"
    );
    // The first moving file is `.claude/rules/b.md` (the plan's items are sorted by path).
    assert_eq!(
        fs::read_to_string(fx.root.join(".claude/rules/b.md")).unwrap(),
        "b: 2 # edited\nx",
        "the file the writer changed stays in place"
    );
    assert_eq!(
        fs::read_to_string(fx.root.join(".claude/workflows/w.md")).unwrap(),
        "# w\n",
        "no other file moved"
    );
    let stamps: Vec<_> = fs::read_dir(fx.root.join(".moai/backups"))
        .unwrap()
        .map(|e| e.unwrap().file_name().into_string().unwrap())
        .filter(|name| name != ".gitignore")
        .collect();
    assert_eq!(stamps.len(), 1, "{stamps:?}");
    let receipt = fx.receipt(&stamps[0]);
    assert_eq!(receipt["state"], json!("pending"), "{receipt}");
    let moves = receipt["moves"].as_array().unwrap();
    for m in moves.iter().filter(|m| m["path"] != ".moai") {
        assert_eq!(m["moved"], json!(false), "{receipt}");
    }
    assert!(!fx
        .area(&stamps[0])
        .join("go-state/.claude/rules/b.md")
        .exists());
}

/// The same race in the layout where the Go folder stays: only the proven `.claude/` file
/// moves, and the tree is otherwise as it was.
#[test]
fn a_file_edited_after_the_fresh_preview_stops_the_move_in_the_staying_layout() {
    let mut fx = Fx::new();
    fx.deployed(".moai/config/a.yaml", "a\n");
    fx.deployed(".claude/workflows/w.md", "# w\n");
    // The settings name `.moai/`, so the folder stays.
    fx.write(".claude/settings.json", r#"{"note":".moai/config"}"#);
    fx.manifest();
    let p = fx.plan();
    assert_eq!(p["layout"], json!("folder_stays"), "{p}");
    assert_eq!(ac(&p, ".claude/workflows/w.md").0, "move", "{p}");
    let v = fx.apply_env(
        p["plan_digest"].as_str().unwrap(),
        &[("MOAI_CLI_TEST_FAULT", "go_state_edit_after_plan")],
    );
    assert_eq!(v["ok"], json!(false), "{v}");
    assert!(
        v.to_string()
            .contains("changed after the preview; move stopped"),
        "{v}"
    );
    assert_eq!(
        fs::read_to_string(fx.root.join(".claude/workflows/w.md")).unwrap(),
        "# w\nx"
    );
    assert!(fx.root.join(".moai/manifest.json").is_file());
    let backups = fx.root.join(".moai-cli/backups");
    let stamp = fs::read_dir(&backups).unwrap().next().unwrap().unwrap();
    assert!(!stamp
        .path()
        .join("go-state/.claude/workflows/w.md")
        .exists());
}

/// A moving file replaced by a link to a file outside the project after the preview is
/// refused (the preview no longer lists it), and neither file is touched.
#[test]
fn a_moving_file_replaced_by_a_link_after_the_preview_is_refused() {
    let fx = base();
    let old = fx.plan();
    let outside = fx.root.parent().unwrap().join("outside.md");
    fs::write(&outside, "# outside\n").unwrap();
    fs::remove_file(fx.root.join(".claude/workflows/w.md")).unwrap();
    symlink(&outside, fx.root.join(".claude/workflows/w.md")).unwrap();
    let fresh = fx.plan();
    assert_eq!(
        item(&fresh, ".claude/workflows/w.md")["action"],
        json!("leave"),
        "{fresh}"
    );
    assert_refused(&fx, &old);
    assert_eq!(fs::read_to_string(&outside).unwrap(), "# outside\n");
}
