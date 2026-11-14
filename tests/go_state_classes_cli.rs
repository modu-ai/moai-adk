//! #2936 S1b: what the backup phase does with each kind of earlier-generation file (PLAN
//! 6877821298, decisions 6877929684). The preview gives every item a `class` and a `proof`;
//! the apply moves the Go skill mirror under `.agents/skills/` (a link as a link, a copy
//! folder whole), files the user edited (O2, both hashes recorded), `.claude/settings.json`
//! (O3) and the root instruction files (O4), copies `.claude/settings.local.json` aside and
//! leaves it in place (O3), restores the user's data folders from the Go backup (O1), and
//! never reads, hashes, copies or moves a secret-named file under `.claude/` or `.agents/`
//! (O6). Every case builds a temporary project; the real checkout is never read.

#![cfg(unix)]

mod common;
use common::CleanEnv;
use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::{
    ffi::CString,
    fs,
    os::unix::fs::{symlink, MetadataExt, PermissionsExt},
    path::{Path, PathBuf},
    process::{Command, Stdio},
    time::{Duration, Instant},
};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;
const REF: &str = "test-permission-ref";
/// `go_state::FILE_LIMIT`.
const FILE_LIMIT: u64 = 64 * 1024 * 1024;
/// A command that opened a FIFO would wait for a writer forever; this ends the test instead.
const DEADLINE: Duration = Duration::from_secs(120);

fn sha(bytes: &[u8]) -> String {
    format!("sha256:{:x}", Sha256::digest(bytes))
}

/// One tree entry, never followed and never opened unless it is a regular file.
#[derive(Debug, PartialEq, Eq, PartialOrd, Ord, Clone)]
enum Node {
    File(Vec<u8>),
    Link(PathBuf),
    Fifo,
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

    fn link(&self, rel: &str, body: &str) {
        let p = self.root.join(rel);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        symlink(body, p).unwrap();
    }

    fn fifo(&self, rel: &str) {
        let p = self.root.join(rel);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        let c = CString::new(p.to_str().unwrap()).unwrap();
        // SAFETY: mkfifo with a valid NUL-terminated path.
        assert_eq!(unsafe { libc::mkfifo(c.as_ptr(), 0o600) }, 0, "{rel}");
    }

    /// A file the Go deployer wrote and recorded.
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

    /// Every entry under the project (`.git` included) and HOME; links read as links, FIFOs
    /// named, never opened.
    fn tree(&self) -> Vec<(PathBuf, Node)> {
        let mut out = Vec::new();
        let mut stack = vec![self.root.clone(), self.home.clone()];
        while let Some(d) = stack.pop() {
            for e in fs::read_dir(&d).unwrap() {
                let p = e.unwrap().path();
                let m = fs::symlink_metadata(&p).unwrap();
                if m.file_type().is_symlink() {
                    out.push((p.clone(), Node::Link(fs::read_link(&p).unwrap())));
                } else if m.is_dir() {
                    stack.push(p);
                } else if m.is_file() {
                    out.push((p.clone(), Node::File(fs::read(&p).unwrap())));
                } else {
                    out.push((p.clone(), Node::Fifo));
                }
            }
        }
        out.sort();
        out
    }

    /// The project as the user sees it: without the backup areas and the set-aside
    /// `.moai.rust-<UTC>` a revert keeps as the record of both acts.
    fn project_tree(&self) -> Vec<(PathBuf, Node)> {
        let home = self.home.clone();
        self.tree()
            .into_iter()
            .filter(|(p, _)| !p.starts_with(&home))
            .filter(|(p, _)| {
                !p.strip_prefix(&self.root)
                    .unwrap()
                    .to_string_lossy()
                    .starts_with(".moai.rust-")
            })
            .collect()
    }

    fn run_env(&self, args: &[&str], env: &[(&str, &str)]) -> Value {
        let mut cmd = Command::new(EXE);
        cmd.args(args)
            .clean_env()
            .env("HOME", &self.home)
            .env("PATH", "/usr/bin:/bin")
            .env("LC_ALL", "C")
            .stdin(Stdio::null())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped());
        for (k, v) in env {
            cmd.env(k, v);
        }
        let mut child = cmd.spawn().unwrap();
        let start = Instant::now();
        while child.try_wait().unwrap().is_none() {
            if start.elapsed() > DEADLINE {
                let _ = child.kill();
                let _ = child.wait();
                panic!("{args:?} did not finish: a FIFO was opened, or the walk did not stop");
            }
            std::thread::sleep(Duration::from_millis(20));
        }
        let out = child.wait_with_output().unwrap();
        serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
            panic!(
                "{args:?}: {}{}",
                String::from_utf8_lossy(&out.stdout),
                String::from_utf8_lossy(&out.stderr)
            )
        })
    }

    fn run(&self, args: &[&str]) -> Value {
        self.run_env(args, &[])
    }

    /// `go-state-plan`, with the whole tree compared before and after.
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

    fn apply(&self, digest: &str) -> Value {
        self.run(&[
            "go-state-apply",
            self.p(),
            "--plan-digest",
            digest,
            "--permission-reference",
            REF,
        ])
    }

    /// Plan and apply; the apply must succeed. Returns the plan and the backup stamp.
    fn apply_ok(&self) -> (Value, String) {
        let p = self.plan();
        let v = self.apply(p["plan_digest"].as_str().unwrap());
        assert_eq!(v["ok"], json!(true), "{v}");
        assert_eq!(v["result"]["state"], json!("committed"), "{v}");
        let backup = v["result"]["backup"].as_str().unwrap().to_string();
        (p, backup)
    }

    fn revert_plan(&self, backup: &str) -> Value {
        self.run(&["go-state-revert-plan", self.p(), "--backup", backup])
    }

    fn revert(&self, backup: &str) -> Value {
        let rp = self.revert_plan(backup);
        assert_eq!(rp["ok"], json!(true), "{rp}");
        self.run(&[
            "go-state-revert",
            self.p(),
            "--backup",
            backup,
            "--plan-digest",
            rp["result"]["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            REF,
        ])
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

fn has_item(plan: &Value, rel: &str) -> bool {
    plan["items"]
        .as_array()
        .unwrap()
        .iter()
        .any(|i| i["path"] == rel)
}

/// (action, class) of one item.
fn ac<'a>(plan: &'a Value, rel: &str) -> (&'a str, &'a str) {
    let i = item(plan, rel);
    (i["action"].as_str().unwrap(), i["class"].as_str().unwrap())
}

fn root_file<'a>(plan: &'a Value, rel: &str) -> &'a Value {
    plan["root_files"]
        .as_array()
        .unwrap()
        .iter()
        .find(|f| f["path"] == rel)
        .unwrap_or_else(|| panic!("{rel} not in root_files: {plan}"))
}

fn moved<'a>(receipt: &'a Value, rel: &str) -> &'a Value {
    receipt["moves"]
        .as_array()
        .unwrap()
        .iter()
        .find(|m| m["path"] == rel)
        .unwrap_or_else(|| panic!("{rel} not in receipt: {receipt}"))
}

/// A Go project whose `.moai/` moves, with one canonical skill the mirror points at.
fn go_fx() -> Fx {
    let mut fx = Fx::new();
    fx.deployed(".moai/config/sections/a.yaml", "a: 1\n");
    fx.deployed(".claude/skills/moai-a/SKILL.md", "# a\n");
    fx.deployed(".claude/skills/moai-a/ref/x.md", "x\n");
    fx
}

/// `.agents/skills/<name>/` as a real copy of `.claude/skills/moai-a/`.
fn copy_folder(fx: &Fx, name: &str) {
    fx.write(&format!(".agents/skills/{name}/SKILL.md"), "# a\n");
    fx.write(&format!(".agents/skills/{name}/ref/x.md"), "x\n");
}

/// A Markdown document as people write it: Korean prose, blank lines, a list and a fenced
/// block, with lines that have no `:` in them. The config classifier refuses such text
/// (audit 6879582411 B1), so every restore test uses this and not a one-line `# title`.
fn prose(title: &str) -> String {
    format!(
        "# {title}\n\n이번 주에 한 일을 적어 둔 글이에요.\n\n- 첫째 항목\n- 둘째 항목\n\n\
         ```\ncargo test\n```\n\n끝.\n"
    )
}

/// A file of `len` bytes written sparse, not recorded in the Go manifest.
fn sparse(fx: &Fx, rel: &str, len: u64) {
    let p = fx.root.join(rel);
    fs::create_dir_all(p.parent().unwrap()).unwrap();
    fs::File::create(&p).unwrap().set_len(len).unwrap();
}

fn chmod(p: &Path, mode: u32) {
    fs::set_permissions(p, fs::Permissions::from_mode(mode)).unwrap();
}

/// The permission bits of an entry, never following a link.
fn mode(p: &Path) -> u32 {
    fs::symlink_metadata(p).unwrap().mode() & 0o777
}

/// One `restores` row of a plan, by its `from`.
fn restore<'a>(plan: &'a Value, from: &str) -> &'a Value {
    plan["restores"]
        .as_array()
        .unwrap()
        .iter()
        .find(|r| r["from"] == from)
        .unwrap_or_else(|| panic!("{from} not in restores: {plan}"))
}

// ---- S1b-1: classes and the preview ----

#[test]
fn a_mirror_link_with_the_exact_body_moves_as_a_link() {
    let fx = go_fx();
    fx.link(".agents/skills/moai-a", "../../.claude/skills/moai-a");
    fx.manifest();
    let p = fx.plan();
    assert_eq!(
        ac(&p, ".agents/skills/moai-a"),
        ("move", "go_mirror"),
        "{p}"
    );
    assert_eq!(
        item(&p, ".agents/skills/moai-a")["proof"],
        json!("link ../../.claude/skills/moai-a"),
        "{p}"
    );
    assert_eq!(
        ac(&p, ".claude/skills/moai-a/SKILL.md"),
        ("move", "tool_asset")
    );
}

#[test]
fn a_mirror_link_with_another_body_an_absolute_body_or_no_go_record_stays() {
    let mut fx = go_fx();
    fx.deployed(".claude/skills/moai-c/SKILL.md", "# c\n");
    // Points at another skill's folder.
    fx.link(".agents/skills/moai-b", "../../.claude/skills/moai-a");
    // The right folder, by an absolute body.
    let abs = fx.root.join(".claude/skills/moai-c");
    fx.link(".agents/skills/moai-c", abs.to_str().unwrap());
    // The exact shape, but the Go manifest records nothing under the canonical folder.
    fx.write(".claude/skills/mine/SKILL.md", "# mine\n");
    fx.link(".agents/skills/mine", "../../.claude/skills/mine");
    fx.manifest();
    let p = fx.plan();
    for rel in [
        ".agents/skills/moai-b",
        ".agents/skills/moai-c",
        ".agents/skills/mine",
    ] {
        assert_eq!(ac(&p, rel), ("leave", "unknown"), "{rel}: {p}");
    }
}

#[test]
fn a_copy_folder_that_matches_the_go_deploy_moves_whole() {
    let fx = go_fx();
    copy_folder(&fx, "moai-a");
    fx.manifest();
    let p = fx.plan();
    assert_eq!(
        ac(&p, ".agents/skills/moai-a"),
        ("move", "go_mirror_copy"),
        "{p}"
    );
    // The folder is one item; its files are not listed one by one.
    assert!(!has_item(&p, ".agents/skills/moai-a/SKILL.md"), "{p}");
}

#[test]
fn a_copy_folder_with_one_changed_byte_is_a_conflict_and_stays() {
    let fx = go_fx();
    copy_folder(&fx, "moai-a");
    fx.write(".agents/skills/moai-a/ref/x.md", "y\n");
    fx.manifest();
    let p = fx.plan();
    assert_eq!(
        ac(&p, ".agents/skills/moai-a"),
        ("leave", "conflict"),
        "{p}"
    );
}

#[test]
fn a_fifo_in_a_copy_folder_keeps_it_in_place_unopened() {
    let fx = go_fx();
    copy_folder(&fx, "moai-a");
    fx.fifo(".agents/skills/moai-a/pipe");
    fx.manifest();
    let p = fx.plan();
    assert_eq!(ac(&p, ".agents/skills/moai-a"), ("leave", "unknown"), "{p}");
}

#[test]
fn a_published_skill_file_is_a_tool_asset_and_others_under_agents_stay() {
    let mut fx = go_fx();
    fx.deployed(".agents/skills/moai-plan/SKILL.md", "# plan\n");
    fx.deployed(".agents/skills/moai-plan/notes.md", "n\n");
    fx.manifest();
    let p = fx.plan();
    assert_eq!(
        ac(&p, ".agents/skills/moai-plan/SKILL.md"),
        ("move", "tool_asset"),
        "{p}"
    );
    assert_eq!(
        ac(&p, ".agents/skills/moai-plan/notes.md").0,
        "leave",
        "{p}"
    );
    // A folder that holds recorded files is not taken for a mirror copy.
    assert!(!has_item(&p, ".agents/skills/moai-plan"), "{p}");
}

#[test]
fn user_created_stays_and_user_edited_files_move_with_both_hashes() {
    let mut fx = go_fx();
    fx.write(".claude/rules/mine.md", "mine\n");
    fx.record(".claude/rules/mine.md", "user_created", &sha(b"mine\n"));
    fx.deployed(".claude/rules/b.md", "b: 1\n");
    fx.write(".claude/rules/b.md", "b: 2 # edited\n");
    fx.write(".claude/rules/u.md", "u: 1\n");
    fx.record(".claude/rules/u.md", "user_modified", &sha(b"u: 1\n"));
    fx.manifest();
    let p = fx.plan();
    assert_eq!(ac(&p, ".claude/rules/mine.md"), ("leave", "user_data"));
    assert_eq!(ac(&p, ".claude/rules/b.md"), ("move", "user_modified"));
    assert_eq!(ac(&p, ".claude/rules/u.md"), ("move", "user_modified"));

    let v = fx.apply(p["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], json!(true), "{v}");
    let backup = v["result"]["backup"].as_str().unwrap();
    let r = fx.receipt(backup);
    let b = moved(&r, ".claude/rules/b.md");
    assert_eq!(b["kind"], json!("edited"), "{r}");
    assert_eq!(b["deployed_hash"], json!(sha(b"b: 1\n")), "{r}");
    assert_eq!(b["sha256"], json!(sha(b"b: 2 # edited\n")), "{r}");
    assert!(fx.root.join(".claude/rules/mine.md").is_file());
    assert_eq!(
        fs::read_to_string(fx.area(backup).join("go-state/.claude/rules/b.md")).unwrap(),
        "b: 2 # edited\n"
    );
}

/// O6: a secret-named entry under `.claude/` or `.agents/` is never opened, hashed, copied or
/// moved, recorded or not, and the preview only counts it. FIFOs make any open hang.
#[test]
fn secret_names_are_counted_and_never_opened() {
    let mut fx = go_fx();
    fx.fifo(".claude/.env");
    fx.fifo(".agents/skills/x/credentials.json");
    fx.write(".claude/hooks/server.pem", "not read\n");
    fx.record(
        ".claude/hooks/server.pem",
        "template_managed",
        &sha(b"not read\n"),
    );
    // A secret in the Go folder's user data is filtered by name before any read.
    fx.fifo(".moai/docs/.env");
    fx.write(".moai/docs/guide.md", "# guide\n");
    fx.manifest();
    let p = fx.plan();
    for rel in [
        ".claude/.env",
        ".agents/skills/x/credentials.json",
        ".claude/hooks/server.pem",
    ] {
        assert!(!has_item(&p, rel), "{rel}: {p}");
    }
    assert_eq!(p["secrets"], json!(3), "{p}");
    let docs: Vec<(&str, &str)> = p["restores"]
        .as_array()
        .unwrap()
        .iter()
        .map(|r| (r["from"].as_str().unwrap(), r["action"].as_str().unwrap()))
        .collect();
    assert!(
        docs.contains(&(".moai-backup/docs/.env", "excluded_secret")),
        "{p}"
    );
    assert!(
        docs.contains(&(".moai-backup/docs/guide.md", "copy")),
        "{p}"
    );
}

#[test]
fn root_instruction_files_carry_their_go_proof_and_markers() {
    let fx = go_fx();
    fx.write(
        "CLAUDE.md",
        "# rules\n<!-- moai:learned-start -->\nx\n<!-- moai:learned-end -->\n",
    );
    fx.write(
        "AGENTS.md",
        "# a\n<!-- moai:managed:begin id=moai-workflow -->\nm\n<!-- moai:managed:end id=moai-workflow -->\n",
    );
    fx.write(
        "AGENTS.local.md",
        "<!-- moai:learned-local-start -->\n<!-- moai:managed:begin id=x -->\n",
    );
    fx.write("CLAUDE.local.md", "# mine\n");
    fx.manifest();
    let p = fx.plan();
    let claude = root_file(&p, "CLAUDE.md");
    assert_eq!(claude["action"], json!("move"), "{p}");
    assert_eq!(claude["go_proof"], json!("marker"), "{p}");
    assert_eq!(root_file(&p, "AGENTS.md")["action"], json!("leave"), "{p}");
    assert_eq!(root_file(&p, "AGENTS.md")["class"], json!("rust"), "{p}");
    assert_eq!(
        root_file(&p, "AGENTS.local.md")["class"],
        json!("conflict"),
        "{p}"
    );
    assert_eq!(
        root_file(&p, "AGENTS.local.md")["action"],
        json!("leave"),
        "{p}"
    );
    assert!(
        !p["root_files"].to_string().contains("CLAUDE.local.md"),
        "CLAUDE.local.md stays and is not listed: {p}"
    );
}

#[test]
fn a_root_file_without_go_proof_moves_and_says_so_and_user_created_stays() {
    let mut fx = go_fx();
    fx.write("CLAUDE.md", "# plain\n");
    fx.write("AGENTS.md", "# recorded\n");
    fx.record("AGENTS.md", "template_managed", &sha(b"# recorded\n"));
    fx.write("AGENTS.local.md", "# mine\n");
    fx.record("AGENTS.local.md", "user_created", &sha(b"# mine\n"));
    fx.manifest();
    let p = fx.plan();
    assert_eq!(root_file(&p, "CLAUDE.md")["action"], json!("move"), "{p}");
    assert_eq!(root_file(&p, "CLAUDE.md")["go_proof"], json!("none"), "{p}");
    assert_eq!(
        root_file(&p, "AGENTS.md")["go_proof"],
        json!("manifest"),
        "{p}"
    );
    assert_eq!(root_file(&p, "AGENTS.md")["action"], json!("move"), "{p}");
    assert_eq!(
        root_file(&p, "AGENTS.local.md")["action"],
        json!("leave"),
        "{p}"
    );
    assert_eq!(
        root_file(&p, "AGENTS.local.md")["class"],
        json!("user_data"),
        "{p}"
    );
}

/// U4: each moving item says what git knows of it, from one read-only status.
#[test]
fn each_moving_item_says_what_git_knows_of_it() {
    let mut fx = go_fx();
    common::git::git_ok(&fx.root, &["init", "-q"]);
    fx.write(".claude/workflows/w.md", "# committed\n");
    fx.write(".claude/workflows/c.md", "# clean\n");
    common::git::git_ok(&fx.root, &["add", ".claude/workflows"]);
    common::git::git_ok(
        &fx.root,
        &[
            "-c",
            "user.name=t",
            "-c",
            "user.email=t@example.invalid",
            "commit",
            "-q",
            "-m",
            "x",
        ],
    );
    fx.deployed(".claude/workflows/w.md", "# deployed later\n");
    fx.deployed(".claude/workflows/c.md", "# clean\n");
    fx.manifest();
    let p = fx.plan();
    assert_eq!(item(&p, ".claude/workflows/w.md")["git"], json!("modified"));
    assert_eq!(item(&p, ".claude/workflows/c.md")["git"], json!("clean"));
    assert_eq!(
        item(&p, ".claude/skills/moai-a/SKILL.md")["git"],
        json!("untracked")
    );
    assert!(
        p["warnings"]
            .to_string()
            .contains("git show HEAD:.claude/workflows/w.md"),
        "{p}"
    );
}

#[test]
fn outside_a_git_checkout_the_git_column_is_unknown() {
    let fx = go_fx();
    fx.manifest();
    let p = fx.plan();
    assert_eq!(
        item(&p, ".claude/skills/moai-a/SKILL.md")["git"],
        json!("unknown")
    );
}

#[test]
fn codex_files_stay_for_s1c() {
    let mut fx = go_fx();
    fx.deployed(".codex/config.toml", "x = 1\n");
    fx.manifest();
    let p = fx.plan();
    let i = item(&p, ".codex/config.toml");
    assert_eq!(i["action"], json!("leave"), "{p}");
    assert!(i["reason"].as_str().unwrap().contains("S1c"), "{p}");
}

/// The backup step of `project-update-plan` shows how many items fall in each class.
#[test]
fn the_backup_step_counts_items_per_class() {
    let fx = go_fx();
    fx.link(".agents/skills/moai-a", "../../.claude/skills/moai-a");
    fx.fifo(".claude/.env");
    fx.manifest();
    let before = fx.tree();
    let v = fx.run(&["project-update-plan", fx.p()]);
    assert_eq!(v["ok"], json!(true), "{v}");
    assert_eq!(fx.tree(), before, "project-update-plan wrote something");
    let step = v["result"]["phases"]
        .as_array()
        .unwrap()
        .iter()
        .flat_map(|ph| ph["steps"].as_array().into_iter().flatten())
        .find(|s| s["step"] == "go-state")
        .unwrap()
        .clone();
    let classes = &step["preview"]["classes"];
    assert_eq!(classes["go_mirror"], json!(1), "{step}");
    assert_eq!(classes["tool_asset"], json!(2), "{step}");
    assert_eq!(step["preview"]["secrets"], json!(1), "{step}");
}

// ---- boundaries ----

/// Zeros of `len` bytes, written sparse, and their sha256.
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

#[test]
fn a_file_of_exactly_the_file_limit_moves_and_one_byte_more_stays() {
    let mut fx = go_fx();
    zeros(&mut fx, ".claude/big/at.bin", FILE_LIMIT, FILE_LIMIT);
    zeros(
        &mut fx,
        ".claude/big/over.bin",
        FILE_LIMIT + 1,
        FILE_LIMIT + 1,
    );
    fx.manifest();
    let p = fx.plan();
    assert_eq!(ac(&p, ".claude/big/at.bin"), ("move", "tool_asset"), "{p}");
    assert_eq!(
        item(&p, ".claude/big/over.bin")["action"],
        json!("leave"),
        "{p}"
    );
    assert_eq!(
        item(&p, ".claude/big/over.bin")["reason"],
        json!("not a bounded regular file"),
        "{p}"
    );
}

/// The entry limit now spans `.agents/` too: N entries fit a limit of N, N+1 do not.
#[test]
fn the_unrecorded_limit_counts_agents_entries_at_its_edge() {
    let mut fx = Fx::new();
    fx.deployed(".moai/config/a.yaml", "a\n");
    fx.write(".agents/skills/note.md", "n\n");
    fx.manifest();
    // .moai/manifest.json, .moai/config, .moai/config/a.yaml, .agents/skills,
    // .agents/skills/note.md: five entries.
    let stopped = |n: &str| {
        let p = fx.plan_env(&[("MOAI_CLI_TEST_UNRECORDED_LIMIT", n)]);
        p["warnings"].to_string().contains("stopped after")
    };
    assert!(!stopped("5"), "five entries fit a limit of five");
    assert!(stopped("4"), "the fifth entry passes a limit of four");
    let p = fx.plan();
    assert_eq!(p["unrecorded"][".agents/skills"], json!(1), "{p}");
}

/// A copy folder of N files fits a limit of N; one more file and it stays.
#[test]
fn a_copy_folder_at_its_file_limit_moves_and_one_more_file_stays() {
    let mut fx = go_fx();
    fx.deployed(".claude/skills/moai-a/ref/y.md", "y\n");
    copy_folder(&fx, "moai-a");
    fx.manifest();
    let env = [("MOAI_CLI_TEST_MIRROR_FILE_LIMIT", "2")];
    // Two files: at the limit.
    let p = fx.plan_env(&env);
    assert_eq!(
        ac(&p, ".agents/skills/moai-a"),
        ("move", "go_mirror_copy"),
        "{p}"
    );
    // Three files: over it.
    fx.write(".agents/skills/moai-a/ref/y.md", "y\n");
    let p = fx.plan_env(&env);
    let i = item(&p, ".agents/skills/moai-a");
    assert_eq!(i["action"], json!("leave"), "{p}");
    assert!(
        i["reason"].as_str().unwrap().contains("more than 2 files"),
        "{p}"
    );
    assert_eq!(
        ac(&fx.plan(), ".agents/skills/moai-a"),
        ("move", "go_mirror_copy"),
        "the default limit is far above three"
    );
}

// ---- S1b-2: apply and revert ----

/// Everything S1b moves, in one project: apply, then revert gives the same tree back, links
/// compared by their body.
#[test]
fn apply_then_revert_restores_the_tree_links_by_body() {
    let mut fx = go_fx();
    fx.deployed(".claude/skills/moai-b/SKILL.md", "# a\n");
    fx.deployed(".claude/skills/moai-b/ref/x.md", "x\n");
    fx.link(".agents/skills/moai-a", "../../.claude/skills/moai-a");
    copy_folder(&fx, "moai-b");
    fx.deployed(".claude/rules/b.md", "b: 1\n");
    fx.write(".claude/rules/b.md", "b: 2\n");
    fx.deployed(
        ".claude/settings.json",
        "{\"permissions\":{\"deny\":[\"Read(./.env)\"]}}",
    );
    fx.write(".claude/settings.local.json", "{\"local\":true}");
    fx.write("CLAUDE.md", "# rules\n");
    fx.write(".moai/plans/p.md", "# plan\n");
    fx.write(".moai/decisions/d.md", "# go decision\n");
    fx.manifest();
    let before = fx.project_tree();
    let (p, backup) = fx.apply_ok();
    assert_eq!(ac(&p, ".agents/skills/moai-b"), ("move", "go_mirror_copy"));
    let area = fx.area(&backup);

    // The link moved as a link, its body unchanged; the copy folder moved whole.
    assert!(fs::symlink_metadata(fx.root.join(".agents/skills/moai-a")).is_err());
    let l = area.join("go-state/.agents/skills/moai-a");
    assert!(fs::symlink_metadata(&l).unwrap().file_type().is_symlink());
    assert_eq!(
        fs::read_link(&l).unwrap(),
        PathBuf::from("../../.claude/skills/moai-a")
    );
    assert!(!fx.root.join(".agents/skills/moai-b").exists());
    assert_eq!(
        fs::read_to_string(area.join("go-state/.agents/skills/moai-b/ref/x.md")).unwrap(),
        "x\n"
    );
    // O3: settings.json moved; settings.local.json stays, a copy beside the backup.
    assert!(!fx.root.join(".claude/settings.json").exists());
    assert!(area.join("go-state/.claude/settings.json").is_file());
    assert_eq!(
        fs::read_to_string(fx.root.join(".claude/settings.local.json")).unwrap(),
        "{\"local\":true}"
    );
    assert_eq!(
        fs::read_to_string(area.join("root-files/.claude/settings.local.json")).unwrap(),
        "{\"local\":true}"
    );
    // O4: the root instruction file moved into root-files/.
    assert!(!fx.root.join("CLAUDE.md").exists());
    assert_eq!(
        fs::read_to_string(area.join("root-files/CLAUDE.md")).unwrap(),
        "# rules\n"
    );
    // O1: the user's data folder is restored; a Rust-named Go folder is not.
    assert_eq!(
        fs::read_to_string(fx.root.join(".moai/plans/p.md")).unwrap(),
        "# plan\n"
    );
    assert!(!fx.root.join(".moai/decisions/d.md").exists());
    let r = fx.receipt(&backup);
    assert_eq!(moved(&r, ".agents/skills/moai-a")["kind"], json!("link"));
    assert_eq!(
        moved(&r, ".agents/skills/moai-a")["target"],
        json!("../../.claude/skills/moai-a")
    );
    assert_eq!(
        moved(&r, ".agents/skills/moai-b")["kind"],
        json!("mirror_copy")
    );
    assert_eq!(moved(&r, "CLAUDE.md")["kind"], json!("root_file"));

    let v = fx.revert(&backup);
    assert_eq!(v["ok"], json!(true), "{v}");
    assert_eq!(fx.project_tree(), before);
}

#[test]
fn a_digest_change_after_the_preview_is_refused() {
    let fx = go_fx();
    copy_folder(&fx, "moai-a");
    fx.manifest();
    let p = fx.plan();
    assert_eq!(ac(&p, ".agents/skills/moai-a").0, "move");
    fx.write(".agents/skills/moai-a/ref/x.md", "y\n");
    let before = fx.tree();
    let v = fx.apply(p["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], json!(false), "{v}");
    assert_eq!(fx.tree(), before, "nothing written");
}

#[test]
fn a_link_body_changed_after_the_preview_is_refused() {
    let fx = go_fx();
    fx.link(".agents/skills/moai-a", "../../.claude/skills/moai-a");
    fx.manifest();
    let p = fx.plan();
    fs::remove_file(fx.root.join(".agents/skills/moai-a")).unwrap();
    fx.link(".agents/skills/moai-a", "../../.claude/skills/moai-a/ref");
    let before = fx.tree();
    let v = fx.apply(p["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], json!(false), "{v}");
    assert_eq!(fx.tree(), before, "nothing written");
}

/// A file with a second hard link elsewhere keeps its inode through apply and revert: the
/// move is a link and an unlink, never a copy.
#[test]
fn a_hard_linked_file_keeps_its_inode_through_apply_and_revert() {
    let mut fx = go_fx();
    fx.deployed(".claude/workflows/w.md", "# w\n");
    let other = fx.home.join("w.md");
    fs::hard_link(fx.root.join(".claude/workflows/w.md"), &other).unwrap();
    let ino = fs::metadata(&other).unwrap().ino();
    fx.manifest();
    let (_, backup) = fx.apply_ok();
    let moved = fs::metadata(fx.area(&backup).join("go-state/.claude/workflows/w.md")).unwrap();
    assert_eq!((moved.ino(), moved.nlink()), (ino, 2));
    let v = fx.revert(&backup);
    assert_eq!(v["ok"], json!(true), "{v}");
    let back = fs::metadata(fx.root.join(".claude/workflows/w.md")).unwrap();
    assert_eq!((back.ino(), back.nlink()), (ino, 2));
}

/// The receipt is not its own authority: a move it names under `.codex/` is refused for every
/// kind, before any write.
#[test]
fn a_receipt_naming_codex_is_refused() {
    let fx = go_fx();
    fx.link(".agents/skills/moai-a", "../../.claude/skills/moai-a");
    fx.manifest();
    let (_, backup) = fx.apply_ok();
    let path = fx.area(&backup).join("go-state-receipt.json");
    let original = fs::read(&path).unwrap();
    for kind in ["", "edited", "link", "mirror_copy", "root_file"] {
        let mut r: Value = serde_json::from_slice(&original).unwrap();
        let mut m = json!({"path": ".codex/x", "sha256": sha(b"x"), "moved": true});
        if !kind.is_empty() {
            m["kind"] = json!(kind);
        }
        r["moves"].as_array_mut().unwrap().push(m);
        fs::write(&path, serde_json::to_vec_pretty(&r).unwrap()).unwrap();
        let rp = fx.revert_plan(&backup);
        assert_eq!(rp["ok"], json!(false), "{kind}: {rp}");
    }
    fs::write(&path, &original).unwrap();
    assert_eq!(fx.revert_plan(&backup)["ok"], json!(true));
}

#[test]
fn a_regenerated_root_file_or_settings_is_a_conflict_on_revert() {
    let mut fx = go_fx();
    fx.write("AGENTS.md", "# go\n");
    fx.deployed(".claude/settings.json", "{}");
    fx.manifest();
    let (_, backup) = fx.apply_ok();
    for (rel, body) in [
        ("AGENTS.md", "# new\n"),
        (".claude/settings.json", "{\"new\":1}"),
    ] {
        fx.write(rel, body);
        let rp = fx.revert_plan(&backup);
        let i = rp["result"]["items"]
            .as_array()
            .unwrap()
            .iter()
            .find(|i| i["path"] == rel)
            .unwrap()
            .clone();
        assert_eq!(i["action"], json!("conflict"), "{rel}: {rp}");
        let before = fx.tree();
        let v = fx.revert(&backup);
        assert_eq!(v["ok"], json!(false), "{rel}: {v}");
        assert_eq!(fx.tree(), before, "{rel}: nothing overwritten");
        fs::remove_file(fx.root.join(rel)).unwrap();
    }
    assert_eq!(fx.revert(&backup)["ok"], json!(true));
}

/// O3: the local settings stay byte for byte, and the backup holds a copy.
#[test]
fn settings_local_stays_in_place_with_a_copy_in_the_backup() {
    let mut fx = go_fx();
    let local = "{\"permissions\":{\"deny\":[\"Bash(rm -rf:*)\"]}}";
    fx.deployed(".claude/settings.local.json", local);
    fx.manifest();
    let p = fx.plan();
    assert_eq!(
        root_file(&p, ".claude/settings.local.json")["action"],
        json!("copy"),
        "{p}"
    );
    assert_eq!(
        item(&p, ".claude/settings.local.json")["action"],
        json!("leave")
    );
    let ino = fs::metadata(fx.root.join(".claude/settings.local.json"))
        .unwrap()
        .ino();
    let (_, backup) = fx.apply_ok();
    let here = fx.root.join(".claude/settings.local.json");
    assert_eq!(fs::read_to_string(&here).unwrap(), local);
    assert_eq!(fs::metadata(&here).unwrap().ino(), ino, "never moved");
    assert_eq!(
        fs::read_to_string(
            fx.area(&backup)
                .join("root-files/.claude/settings.local.json")
        )
        .unwrap(),
        local
    );
}

/// O6 through the apply and the revert: secret-named entries never move, FIFO or not.
#[test]
fn secrets_stay_untouched_through_apply_and_revert() {
    let mut fx = go_fx();
    fx.fifo(".claude/.env");
    fx.write(".claude/hooks/server.pem", "not read\n");
    fx.record(
        ".claude/hooks/server.pem",
        "template_managed",
        &sha(b"not read\n"),
    );
    fx.manifest();
    let pem_ino = fs::metadata(fx.root.join(".claude/hooks/server.pem"))
        .unwrap()
        .ino();
    let (_, backup) = fx.apply_ok();
    let check = |fx: &Fx| {
        let m = fs::symlink_metadata(fx.root.join(".claude/.env")).unwrap();
        assert!(!m.is_file() && !m.is_dir(), "the FIFO is still a FIFO");
        let pem = fs::metadata(fx.root.join(".claude/hooks/server.pem")).unwrap();
        assert_eq!(pem.ino(), pem_ino);
    };
    check(&fx);
    let names = fx
        .tree()
        .into_iter()
        .filter(|(p, _)| p.starts_with(fx.area(&backup)))
        .map(|(p, _)| p.to_string_lossy().into_owned())
        .collect::<Vec<_>>();
    assert!(
        !names
            .iter()
            .any(|n| n.ends_with(".env") || n.ends_with(".pem")),
        "{names:?}"
    );
    let v = fx.revert(&backup);
    assert_eq!(v["ok"], json!(true), "{v}");
    check(&fx);
}

#[test]
fn user_data_folders_are_restored_and_removed_again_on_revert() {
    let fx = go_fx();
    fx.write(".moai/reports/r.md", &prose("보고서"));
    fx.write(".moai/research/n.md", &prose("조사 메모"));
    fx.write(".moai/state/s.json", "{}");
    fx.manifest();
    let p = fx.plan();
    let froms: Vec<&str> = p["restores"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|r| r["action"] == "copy")
        .map(|r| r["from"].as_str().unwrap())
        .collect();
    assert_eq!(
        froms,
        [".moai-backup/reports/r.md", ".moai-backup/research/n.md"],
        "{p}"
    );
    let before = fx.project_tree();
    let (_, backup) = fx.apply_ok();
    assert!(fx.root.join(".moai/reports/r.md").is_file());
    assert!(!fx.root.join(".moai/state/s.json").exists());
    let r = fx.receipt(&backup);
    assert!(
        r["restored"].to_string().contains(".moai/research/n.md"),
        "{r}"
    );
    let v = fx.revert(&backup);
    assert_eq!(v["ok"], json!(true), "{v}");
    assert_eq!(fx.project_tree(), before);
}

// ---- repair of audit 6879582411 (PR #2966): B1 documents, B3 modes and git, N5, N6 ----

/// B1: every user-data folder gets a realistic Markdown document (and a nested one); none of
/// them is a config file, so none goes through the config classifier. Restored byte for byte,
/// and the revert gives the project back.
#[test]
fn prose_markdown_in_every_user_data_folder_is_restored() {
    let fx = go_fx();
    let folders = [
        "specs",
        "project",
        "plans",
        "reports",
        "docs",
        "research",
        "memory",
        "proposals",
        "todo",
        "release-notes",
    ];
    for f in folders {
        fx.write(&format!(".moai/{f}/note.md"), &prose(f));
    }
    fx.write(".moai/research/회의.txt", "회의 메모\n\n정해진 것 없음\n");
    fx.write(".moai/plans/nested/deep/guide.md", &prose("guide"));
    fx.write(".moai/todo/README", "확장자 없는 안내문\n");
    fx.manifest();
    let before = fx.project_tree();
    let p = fx.plan();
    let mut want: Vec<String> = folders.iter().map(|f| format!("{f}/note.md")).collect();
    want.extend(
        [
            "research/회의.txt",
            "plans/nested/deep/guide.md",
            "todo/README",
        ]
        .map(String::from),
    );
    for inner in &want {
        let r = restore(&p, &format!(".moai-backup/{inner}"));
        assert_eq!(r["action"], json!("copy"), "{inner}: {r}");
    }
    let (_, backup) = fx.apply_ok();
    for inner in &want {
        let from = fx.root.join(".moai-backup").join(inner);
        let to = fx.root.join(".moai").join(inner);
        assert_eq!(fs::read(&to).unwrap(), fs::read(&from).unwrap(), "{inner}");
    }
    let v = fx.revert(&backup);
    assert_eq!(v["ok"], json!(true), "{v}");
    assert_eq!(fx.project_tree(), before);
}

/// B1: only `.json`, `.toml`, `.yaml` and `.yml` (any case) are config-shaped and go through the
/// classifier; a refusal carries its reason, so the preview never drops a file silently.
#[test]
fn only_config_shaped_files_are_classified_and_a_refusal_says_why() {
    let fx = go_fx();
    fx.write(".moai/docs/ok.json", "{\"width\": 80}");
    fx.write(".moai/docs/prose.yaml", "그냥 글이에요\n둘째 줄\n");
    fx.write(".moai/docs/Loose.YML", "이것도 글이에요\n");
    fx.write(".moai/docs/keys.json", "{\"api_token\": \"x\"}");
    fx.write(".moai/docs/.env", "not read\n");
    fx.manifest();
    let p = fx.plan();
    assert_eq!(
        restore(&p, ".moai-backup/docs/ok.json")["action"],
        json!("copy")
    );
    for name in ["prose.yaml", "Loose.YML"] {
        let r = restore(&p, &format!(".moai-backup/docs/{name}"));
        assert_eq!(r["action"], json!("excluded_unclassifiable"), "{name}: {r}");
        assert!(
            r["reason"]
                .as_str()
                .is_some_and(|s| s.contains("unsupported config format")),
            "{name}: the refusal says why: {r}"
        );
    }
    for name in ["keys.json", ".env"] {
        let r = restore(&p, &format!(".moai-backup/docs/{name}"));
        assert_eq!(r["action"], json!("excluded_secret"), "{name}: {r}");
        assert!(
            r["reason"].as_str().is_some_and(|s| !s.is_empty()),
            "{name}: {r}"
        );
    }
}

/// B1 limit: a document of exactly `FILE_LIMIT` bytes is restored, one byte more is not, and
/// says why.
#[test]
fn a_document_of_exactly_the_file_limit_is_restored_and_one_byte_more_is_not() {
    let fx = go_fx();
    sparse(&fx, ".moai/docs/at.md", FILE_LIMIT);
    sparse(&fx, ".moai/docs/over.md", FILE_LIMIT + 1);
    fx.manifest();
    let p = fx.plan();
    let at = restore(&p, ".moai-backup/docs/at.md");
    assert_eq!(at["action"], json!("copy"), "{at}");
    let over = restore(&p, ".moai-backup/docs/over.md");
    assert_eq!(over["action"], json!("excluded_unclassifiable"), "{over}");
    assert!(
        over["reason"]
            .as_str()
            .is_some_and(|s| s.contains("file limit")),
        "{over}"
    );
}

/// N6: a user-data folder that is a link is never followed. The plan leaves it with a reason
/// and the apply goes on with the rest; before, the plan said `copy` and the apply stopped
/// after `.moai` was renamed.
#[test]
fn a_linked_user_data_folder_is_left_in_the_plan_and_the_apply_goes_on() {
    let fx = go_fx();
    fx.write("outside-docs/guide.md", &prose("guide"));
    fx.link(".moai/docs", "../outside-docs");
    fx.write(".moai/plans/p.md", &prose("plan"));
    fx.manifest();
    let before = fx.project_tree();
    let p = fx.plan();
    let r = restore(&p, ".moai-backup/docs");
    assert_eq!(r["action"], json!("leave_link"), "{r}");
    assert!(
        r["reason"].as_str().is_some_and(|s| s.contains("link")),
        "{r}"
    );
    assert!(
        !p["restores"]
            .to_string()
            .contains(".moai-backup/docs/guide.md"),
        "nothing is read through the link: {p}"
    );
    assert_eq!(
        restore(&p, ".moai-backup/plans/p.md")["action"],
        json!("copy")
    );
    let (_, backup) = fx.apply_ok();
    assert!(fx.root.join(".moai/plans/p.md").is_file());
    assert!(fs::symlink_metadata(fx.root.join(".moai/docs")).is_err());
    assert_eq!(
        fs::read_link(fx.root.join(".moai-backup/docs")).unwrap(),
        PathBuf::from("../outside-docs")
    );
    let v = fx.revert(&backup);
    assert_eq!(v["ok"], json!(true), "{v}");
    assert_eq!(fx.project_tree(), before);
}

/// B3: the copy of the local settings (which may hold tokens) is owner-only even when the
/// original is wider, every backup folder is 0700 and the receipt is 0600. A moved file keeps
/// its mode (same inode), the original settings are untouched.
#[test]
fn the_backup_area_holds_owner_only_copies_in_owner_only_folders() {
    for original in [0o600, 0o644] {
        let mut fx = go_fx();
        fx.write(
            ".claude/settings.local.json",
            "{\"env\":{\"ANTHROPIC_AUTH_TOKEN\":\"not-a-real-token\"}}",
        );
        chmod(&fx.root.join(".claude/settings.local.json"), original);
        fx.write("CLAUDE.md", "# rules\n");
        chmod(&fx.root.join("CLAUDE.md"), 0o644);
        fx.write("AGENTS.md", "# agents\n");
        chmod(&fx.root.join("AGENTS.md"), 0o640);
        fx.deployed(".claude/rules/r.md", "r: 1\n");
        chmod(&fx.root.join(".claude/rules/r.md"), 0o664);
        fx.manifest();
        let (_, backup) = fx.apply_ok();
        let area = fx.area(&backup);
        let copy = area.join("root-files/.claude/settings.local.json");
        assert_eq!(mode(&copy), 0o600, "original {original:o}");
        assert_eq!(
            mode(&fx.root.join(".claude/settings.local.json")),
            original,
            "the original is untouched"
        );
        assert_eq!(mode(&area.join("go-state-receipt.json")), 0o600);
        // A moved file is the same inode: its mode is its own.
        assert_eq!(mode(&area.join("root-files/CLAUDE.md")), 0o644);
        assert_eq!(mode(&area.join("root-files/AGENTS.md")), 0o640);
        assert_eq!(mode(&area.join("go-state/.claude/rules/r.md")), 0o664);
        let mut dirs = 0;
        let mut stack = vec![fx.root.join(".moai/backups")];
        while let Some(d) = stack.pop() {
            assert_eq!(mode(&d), 0o700, "{}", d.display());
            dirs += 1;
            for e in fs::read_dir(&d).unwrap() {
                let p = e.unwrap().path();
                if fs::symlink_metadata(&p).unwrap().is_dir() {
                    stack.push(p);
                }
            }
        }
        assert!(
            dirs >= 6,
            "backups, the stamp, root-files, .claude, go-state, …: {dirs}"
        );
    }
}

/// B3: the layout that leaves the Go folder in place keeps its backup area private too.
#[test]
fn the_staying_folder_layout_keeps_its_backup_area_private() {
    let fx = go_fx();
    // A reference under `.moai/` keeps the Go folder where it is (decision 2).
    let settings = "{\"hooks\":\"node .moai/hooks/x.js\"}";
    fx.write(".claude/settings.local.json", settings);
    fx.manifest();
    let (p, backup) = fx.apply_ok();
    assert_eq!(item(&p, ".moai")["action"], json!("leave"), "{p}");
    let found: Vec<PathBuf> = [".moai", ".moai-cli"]
        .iter()
        .map(|b| fx.root.join(b).join("backups"))
        .filter(|b| b.join(&backup).is_dir())
        .collect();
    assert_eq!(found.len(), 1, "{found:?}");
    let area = found[0].join(&backup);
    assert_eq!(mode(&area.join("go-state-receipt.json")), 0o600);
    for d in [
        &found[0],
        &area,
        &area.join("go-state"),
        &area.join("go-state/.claude"),
    ] {
        assert_eq!(mode(d), 0o700, "{}", d.display());
    }
}

/// B3: the earlier-root move (C4a) makes the same private area, kept out of git.
#[test]
fn the_earlier_root_move_keeps_its_backup_area_private_and_out_of_git() {
    let fx = Fx::new();
    fx.write(".moai-cli/project.yaml", "schema_version: 2\n");
    fx.write(".moai-cli/index/a.json", "{\"a\":1}");
    common::git::git_ok(&fx.root, &["init", "-q"]);
    let (_, backup) = fx.apply_ok();
    let area = fx.area(&backup);
    assert_eq!(mode(&area.join("go-state-receipt.json")), 0o600);
    assert_eq!(mode(&area), 0o700);
    assert_eq!(mode(&fx.root.join(".moai/backups")), 0o700);
    let rel = format!(".moai/backups/{backup}/go-state-receipt.json");
    assert!(
        common::git::git_raw(&fx.root, &["check-ignore", "-q", "--", &rel]).success,
        "{rel} is ignored"
    );
}

/// B3: restored user data keeps the mode of the file in the Go backup.
#[test]
fn restored_user_data_keeps_the_mode_of_the_original() {
    let fx = go_fx();
    let cases = [
        (".moai/specs/SPEC-1/spec.md", 0o600),
        (".moai/plans/p.md", 0o640),
        (".moai/docs/run.md", 0o755),
        (".moai/memory/m.md", 0o444),
    ];
    for (rel, m) in cases {
        fx.write(rel, &prose(rel));
        chmod(&fx.root.join(rel), m);
    }
    fx.manifest();
    let (_, backup) = fx.apply_ok();
    for (rel, m) in cases {
        assert_eq!(mode(&fx.root.join(rel)), m, "{rel}");
    }
    let v = fx.revert(&backup);
    assert_eq!(v["ok"], json!(true), "{v}");
    for (rel, m) in cases {
        assert_eq!(mode(&fx.root.join(rel)), m, "{rel} after the revert");
    }
}

/// B3: the backup area is never offered to git: `.moai/backups/.gitignore` holds `*`, and
/// `git check-ignore` agrees for the copy, the receipt and a moved file.
#[test]
fn the_backup_area_is_ignored_by_git() {
    let mut fx = go_fx();
    common::git::git_ok(&fx.root, &["init", "-q"]);
    fx.write(
        ".claude/settings.local.json",
        "{\"env\":{\"ANTHROPIC_AUTH_TOKEN\":\"x\"}}",
    );
    fx.write("CLAUDE.md", "# rules\n");
    fx.deployed(".claude/rules/r.md", "r: 1\n");
    fx.manifest();
    let (_, backup) = fx.apply_ok();
    assert_eq!(
        fs::read_to_string(fx.root.join(".moai/backups/.gitignore")).unwrap(),
        "*\n"
    );
    let area = format!(".moai/backups/{backup}");
    for rel in [
        format!("{area}/root-files/.claude/settings.local.json"),
        format!("{area}/root-files/CLAUDE.md"),
        format!("{area}/go-state-receipt.json"),
        format!("{area}/go-state/.claude/rules/r.md"),
    ] {
        let out = common::git::git_raw(&fx.root, &["check-ignore", "-q", "--", &rel]);
        assert!(out.success, "{rel} must be ignored: {}", out.stderr);
    }
    let status = common::git::git(
        &fx.root,
        &["status", "--porcelain", "--untracked-files=all"],
    );
    assert!(!status.contains(".moai/backups"), "{status}");
}

/// B3 guard: apply then revert gives back every file and folder with the mode it had.
#[test]
fn apply_then_revert_keeps_every_mode() {
    let mut fx = go_fx();
    fx.deployed(".claude/rules/r.md", "r: 1\n");
    chmod(&fx.root.join(".claude/rules/r.md"), 0o664);
    fx.write(".claude/settings.local.json", "{\"local\":true}");
    chmod(&fx.root.join(".claude/settings.local.json"), 0o600);
    fx.write("CLAUDE.md", "# rules\n");
    chmod(&fx.root.join("CLAUDE.md"), 0o640);
    fx.write(".moai/specs/S/spec.md", &prose("spec"));
    chmod(&fx.root.join(".moai/specs/S/spec.md"), 0o600);
    fx.write(".moai/docs/g.md", &prose("guide"));
    fx.manifest();
    let modes = |fx: &Fx| -> Vec<(PathBuf, u32)> {
        let mut out = Vec::new();
        let mut stack = vec![fx.root.clone()];
        while let Some(d) = stack.pop() {
            for e in fs::read_dir(&d).unwrap() {
                let p = e.unwrap().path();
                let rel = p.strip_prefix(&fx.root).unwrap().to_string_lossy();
                if rel.starts_with(".moai.rust-") {
                    continue;
                }
                let m = fs::symlink_metadata(&p).unwrap();
                if m.file_type().is_symlink() {
                    continue;
                }
                out.push((p.clone(), m.mode() & 0o777));
                if m.is_dir() {
                    stack.push(p);
                }
            }
        }
        out.sort();
        out
    };
    let (before, tree) = (modes(&fx), fx.project_tree());
    let (_, backup) = fx.apply_ok();
    let v = fx.revert(&backup);
    assert_eq!(v["ok"], json!(true), "{v}");
    assert_eq!(fx.project_tree(), tree);
    assert_eq!(modes(&fx), before);
}

// ---- N5: the limits that no test reached ----

/// The secret-name count stops after `limit` entries under `.claude/` and `.agents/`: two
/// entries fit a limit of two; the second passes a limit of one and the count says it is a
/// lower bound.
#[test]
fn the_secret_count_limit_counts_entries_at_its_edge() {
    let mut fx = Fx::new();
    fx.deployed(".moai/config/a.yaml", "a: 1\n");
    fx.write(".claude/.env", "not read\n");
    fx.write(".claude/note.md", "n\n");
    fx.manifest();
    let lower_bound = |n: &str| {
        let p = fx.plan_env(&[("MOAI_CLI_TEST_UNRECORDED_LIMIT", n)]);
        (
            p["warnings"].to_string().contains("secret-name count"),
            p["secrets"].clone(),
        )
    };
    assert_eq!(
        lower_bound("2"),
        (false, json!(1)),
        "two entries fit a limit of two"
    );
    assert!(lower_bound("1").0, "the second entry passes a limit of one");
    assert_eq!(fx.plan()["secrets"], json!(1));
}

/// A copy folder with N sub-folders fits a folder limit of N; one more folder and it stays,
/// with the reason (the file count stays inside the limit throughout).
#[test]
fn a_copy_folder_at_its_folder_limit_moves_and_one_more_folder_stays() {
    let mut fx = Fx::new();
    fx.deployed(".claude/skills/moai-m/SKILL.md", "# m\n");
    fx.deployed(".claude/skills/moai-m/a/b/f.md", "f\n");
    fx.write(".agents/skills/moai-m/SKILL.md", "# m\n");
    fx.write(".agents/skills/moai-m/a/b/f.md", "f\n");
    fx.manifest();
    let env = [("MOAI_CLI_TEST_MIRROR_FILE_LIMIT", "2")];
    // Two files, two folders (`a`, `a/b`): at the limit.
    let p = fx.plan_env(&env);
    assert_eq!(
        ac(&p, ".agents/skills/moai-m"),
        ("move", "go_mirror_copy"),
        "{p}"
    );
    // A third, empty folder: over it.
    fs::create_dir(fx.root.join(".agents/skills/moai-m/e")).unwrap();
    let p = fx.plan_env(&env);
    let i = item(&p, ".agents/skills/moai-m");
    assert_eq!(i["action"], json!("leave"), "{p}");
    assert!(
        i["reason"]
            .as_str()
            .unwrap()
            .contains("more than 2 folders"),
        "{p}"
    );
    assert_eq!(
        ac(&fx.plan(), ".agents/skills/moai-m"),
        ("move", "go_mirror_copy"),
        "the default limit is far above three"
    );
}

/// A root instruction file of exactly `FILE_LIMIT` bytes is planned and applied; one byte more
/// stays where it is.
#[test]
fn a_root_file_of_exactly_the_file_limit_moves_and_one_byte_more_stays() {
    let fx = go_fx();
    sparse(&fx, "CLAUDE.md", FILE_LIMIT);
    sparse(&fx, "AGENTS.md", FILE_LIMIT + 1);
    fx.manifest();
    let p = fx.plan();
    assert_eq!(root_file(&p, "CLAUDE.md")["action"], json!("move"), "{p}");
    let over = root_file(&p, "AGENTS.md");
    assert_eq!(over["action"], json!("leave"), "{over}");
    assert_eq!(
        over["reason"],
        json!("not a bounded regular file"),
        "{over}"
    );
    // The apply reads the file again and bounds it the same way.
    fs::remove_file(fx.root.join("AGENTS.md")).unwrap();
    let (_, backup) = fx.apply_ok();
    let moved = fx.area(&backup).join("root-files/CLAUDE.md");
    assert_eq!(fs::metadata(&moved).unwrap().len(), FILE_LIMIT);
    assert!(!fx.root.join("CLAUDE.md").exists());
    let v = fx.revert(&backup);
    assert_eq!(v["ok"], json!(true), "{v}");
    assert_eq!(
        fs::metadata(fx.root.join("CLAUDE.md")).unwrap().len(),
        FILE_LIMIT
    );
}

// ---- B2: what the skill tells an agent to say ----

/// The PROJECT workflow must not tell an agent that `CLAUDE.md` and `AGENTS.md` stay in the
/// root: after the apply they are moved into the backup's `root-files/` (`action` `move`).
#[test]
fn the_project_workflow_says_the_root_instruction_files_are_moved() {
    let workflow = include_str!("../skills/moai/workflows/project.md");
    for sentence in [
        "the root instruction files `CLAUDE.md`, `AGENTS.md` and `AGENTS.local.md` are moved to `.moai/backups/<UTC>/root-files/`",
        "A `root_files` row with `action` `move` is no longer in the project root after the apply",
        "a copy of `.claude/settings.local.json` (the original settings file stays)",
        "`.moai/backups/.gitignore` keeps them out of git",
    ] {
        assert!(workflow.contains(sentence), "project.md lacks: {sentence}");
    }
    for stale in [
        "`CLAUDE.md` and `AGENTS.md` themselves stay as they are",
        "the copies of `CLAUDE.md` and `AGENTS.md` go to",
    ] {
        assert!(!workflow.contains(stale), "project.md still says: {stale}");
    }
}
