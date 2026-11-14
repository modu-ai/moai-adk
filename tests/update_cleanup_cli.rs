//! #2872 stage 1: `cleanup-plan`, `cleanup-apply`, `cleanup-revert-plan` and `cleanup-revert`
//! on temporary HOME folders only. Every call passes the temporary HOME as its argument and as
//! `HOME`; the real HOME is never read or written. The Rust skill is installed with
//! `update-plan`/`update-apply`, so the host receipt that proves projection extras is real.
#![cfg(unix)]

mod common;
use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::collections::BTreeMap;
use std::fs;
use std::os::unix::fs::PermissionsExt;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};

const BIN: &str = crate::common::MOAI_BIN;
const NOW: &str = "2026-11-14T01:02:03Z";
const STAMP: &str = "20261114T010203Z";
const REF: &str = "fixture-approved-2872";

fn sha(b: &[u8]) -> String {
    format!("{:x}", Sha256::digest(b))
}

fn write(root: &Path, rel: &str, body: &str) {
    let p = root.join(rel);
    fs::create_dir_all(p.parent().unwrap()).unwrap();
    fs::write(p, body).unwrap();
}

fn chmod(p: &Path, mode: u32) {
    fs::set_permissions(p, fs::Permissions::from_mode(mode)).unwrap();
}

fn mkfifo(p: &Path) {
    let st = Command::new("/usr/bin/mkfifo").arg(p).status().unwrap();
    assert!(st.success());
}

/// Every entry under `root` with its bytes (regular files that can be read), its link target
/// (links, never followed) or a kind tag; never opens a FIFO.
fn tree(root: &Path, skip: &[&str]) -> BTreeMap<String, String> {
    let mut out = BTreeMap::new();
    let mut stack = vec![root.to_path_buf()];
    while let Some(d) = stack.pop() {
        for e in fs::read_dir(&d).unwrap().flatten() {
            let p = e.path();
            let rel = p.strip_prefix(root).unwrap().display().to_string();
            if skip
                .iter()
                .any(|s| rel == *s || rel.starts_with(&format!("{s}/")))
            {
                continue;
            }
            let m = fs::symlink_metadata(&p).unwrap();
            let v = if m.file_type().is_symlink() {
                format!("link:{}", fs::read_link(&p).unwrap().display())
            } else if m.is_dir() {
                stack.push(p);
                "dir".to_string()
            } else if m.is_file() {
                match fs::read(&p) {
                    Ok(b) => format!("file:{}", sha(&b)),
                    Err(_) => format!("unreadable:{}", m.len()),
                }
            } else {
                "special".to_string()
            };
            out.insert(rel, v);
        }
    }
    out
}

fn run(home: &Path, args: &[&str], stdin: Option<&str>, env: &[(&str, &str)]) -> (Value, bool) {
    let mut cmd = Command::new(BIN);
    cmd.args(args)
        .env_remove("MOAI_HOME")
        .env("HOME", home)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped());
    for (k, v) in env {
        cmd.env(k, v);
    }
    let mut child = cmd.spawn().unwrap();
    let pipe = child.stdin.take().unwrap();
    if let Some(s) = stdin {
        common::write_stdin(pipe, s.as_bytes());
    } else {
        drop(pipe);
    }
    let out = child.wait_with_output().unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap_or_else(
        |_| json!({"ok": false, "stderr": String::from_utf8_lossy(&out.stderr).to_string()}),
    );
    (v, out.status.success())
}

fn ok(home: &Path, args: &[&str], stdin: Option<&str>) -> Value {
    let (v, success) = run(home, args, stdin, &[]);
    assert!(success && v["ok"] == json!(true), "{args:?}: {v}");
    v["result"].clone()
}

struct Fixture {
    _dir: tempfile::TempDir,
    home: PathBuf,
    stage: PathBuf,
}

impl Fixture {
    fn h(&self) -> &str {
        self.home.to_str().unwrap()
    }
    fn s(&self) -> &str {
        self.stage.to_str().unwrap()
    }
    fn plan(&self) -> Value {
        ok(&self.home, &["cleanup-plan", self.h(), self.s()], None)
    }
    fn apply(&self, token: &str, env: &[(&str, &str)]) -> (Value, bool) {
        run(
            &self.home,
            &[
                "cleanup-apply",
                self.h(),
                self.s(),
                "--confirm-cleanup",
                token,
                "--permission-reference",
                REF,
                "--now",
                NOW,
            ],
            None,
            env,
        )
    }
    fn revert_plan(&self, stamp: &str) -> Value {
        ok(
            &self.home,
            &["cleanup-revert-plan", self.h(), "--stamp", stamp],
            None,
        )
    }
    fn revert(&self, stamp: &str) -> (Value, bool) {
        run(
            &self.home,
            &[
                "cleanup-revert",
                self.h(),
                "--stamp",
                stamp,
                "--permission-reference",
                REF,
            ],
            None,
            &[],
        )
    }
    fn snapshot(&self) -> BTreeMap<String, String> {
        tree(&self.home, &[".moai-cli/cleanup"])
    }
}

fn skill_md(body: &str) -> String {
    format!("---\nname: moai\ndescription: fixture package\n---\n# MoAI\n{body}\n")
}

/// A HOME with the Rust skill installed from a four-file package, then staged down to two
/// files, plus a Go generation: `~/.moai` with a user-asset manifest, sibling skills, agents,
/// secrets, a large run-data folder, a link and a FIFO.
fn fixture() -> Fixture {
    let dir = tempfile::tempdir().unwrap();
    let root = fs::canonicalize(dir.path()).unwrap();
    let home = root.join("home");
    fs::create_dir_all(&home).unwrap();
    let v1 = root.join("stage-v1");
    for (rel, body) in [
        ("SKILL.md", skill_md("v1")),
        ("a.md", "a\n".into()),
        ("b.md", "b\n".into()),
        ("c.md", "c\n".into()),
    ] {
        write(&v1, rel, &body);
    }
    let stage = root.join("stage-v2");
    write(&stage, "SKILL.md", &skill_md("v1"));
    write(&stage, "a.md", "a\n");

    // A Go `~/.moai` first: on a HOME without one the install adopts `~/.moai` for Rust.
    write(&home, ".moai/state/active-sessions.json", "[]");
    let (h, s1) = (home.to_str().unwrap(), v1.to_str().unwrap());
    let plan = ok(&home, &["update-plan", h, s1], None);
    let digest = plan["plan_digest"].as_str().unwrap().to_string();
    ok(
        &home,
        &[
            "update-apply",
            h,
            s1,
            "--plan-digest",
            &digest,
            "--permission-reference",
            REF,
        ],
        Some(&plan.to_string()),
    );
    assert!(home.join(".claude/skills/moai/b.md").is_file());
    assert!(home.join(".agents/skills/moai/c.md").is_file());

    // Projection: an edit, an unrecorded file, a Go-recorded file and a secret.
    write(&home, ".claude/skills/moai/c.md", "c edited by the user\n");
    write(&home, ".claude/skills/moai/user-note.md", "mine\n");
    write(&home, ".claude/skills/moai/legacy.md", "go legacy\n");
    write(
        &home,
        ".claude/skills/moai/.env",
        "TOKEN=projection-secret\n",
    );
    chmod(&home.join(".claude/skills/moai/.env"), 0o000);

    // Go sibling skills and agents.
    write(&home, ".claude/skills/moai-proven/SKILL.md", "proven\n");
    write(&home, ".claude/skills/moai-proven/refs/x.md", "x\n");
    write(&home, ".claude/skills/moai-extra/SKILL.md", "extra\n");
    write(&home, ".claude/skills/moai-extra/mine.md", "user file\n");
    write(&home, ".agents/skills/moai-mismatch/SKILL.md", "changed\n");
    write(&home, ".claude/skills/other-skill/SKILL.md", "not moai\n");
    write(&home, "elsewhere/SKILL.md", "outside\n");
    std::os::unix::fs::symlink(
        home.join("elsewhere"),
        home.join(".claude/skills/moai-link"),
    )
    .unwrap();
    write(&home, ".claude/agents/moai-a.md", "agent a\n");
    write(&home, ".claude/agents/moai-b.md", "agent b edited\n");
    write(&home, ".claude/agents/mine.md", "my agent\n");
    write(&home, ".codex/agents/moai-c.toml", "agent c\n");

    let files = json!({
        "claude-skills/moai/legacy.md": {"sha256": sha(b"go legacy\n"), "bundle": "core"},
        "claude-skills/moai-proven/SKILL.md": {"sha256": sha(b"proven\n"), "bundle": "core"},
        "claude-skills/moai-proven/refs/x.md": {"sha256": sha(b"x\n"), "bundle": "core"},
        "claude-skills/moai-extra/SKILL.md": {"sha256": sha(b"extra\n"), "bundle": "core"},
        "agents-skills/moai-mismatch/SKILL.md": {"sha256": sha(b"original\n"), "bundle": "core"},
        "claude-agents/moai-a.md": {"sha256": sha(b"agent a\n"), "bundle": "core"},
        "claude-agents/moai-b.md": {"sha256": sha(b"agent b\n"), "bundle": "core"},
        "codex-agents/moai-c.toml": {"sha256": sha(b"agent c\n"), "bundle": "core"},
        "claude-agents@work/moai-a.md": {"sha256": sha(b"agent a\n"), "bundle": "core"},
    });
    write(
        &home,
        ".moai/user-assets.json",
        &json!({"schema_version": 1, "bundles": {}, "files": files, "collisions": []}).to_string(),
    );
    // `~/.moai`: Go generation markers, secrets, run data and leftovers.
    assert!(!home.join(".moai/moai-cli.json").exists());
    for i in 0..300 {
        write(&home, &format!(".moai/state/runs/r{i}/log.txt"), "run\n");
    }
    write(&home, ".moai/credentials.yaml", "token: go-secret\n");
    chmod(&home.join(".moai/credentials.yaml"), 0o000);
    write(&home, ".moai/.env", "KEY=go-secret\n");
    chmod(&home.join(".moai/.env"), 0o000);
    write(&home, ".moai/cache/blob", "cached\n");
    write(&home, ".moai/tmp-123/part", "partial\n");
    fs::create_dir_all(home.join(".moai/logs")).unwrap();
    write(&home, ".moai/run/pid", "1\n");
    write(&home, ".moai/memories/a.md", "remember\n");
    mkfifo(&home.join(".moai/tmp-fifo"));
    std::os::unix::fs::symlink(home.join("elsewhere"), home.join(".moai/tmp-link")).unwrap();
    Fixture {
        _dir: dir,
        home,
        stage,
    }
}

fn actions(plan: &Value) -> BTreeMap<String, (String, String)> {
    plan["items"]
        .as_array()
        .unwrap()
        .iter()
        .map(|i| {
            (
                i["path"].as_str().unwrap().to_string(),
                (
                    i["action"].as_str().unwrap().to_string(),
                    i["reason"].as_str().unwrap().to_string(),
                ),
            )
        })
        .collect()
}

const ARCHIVED: [&str; 10] = [
    ".agents/skills/moai/b.md",
    ".agents/skills/moai/c.md",
    ".claude/agents/moai-a.md",
    ".claude/skills/moai-proven",
    ".claude/skills/moai/b.md",
    ".claude/skills/moai/legacy.md",
    ".codex/agents/moai-c.toml",
    ".moai/cache",
    ".moai/logs",
    ".moai/tmp-123",
];

#[test]
fn the_plan_archives_only_proven_items_and_writes_nothing() {
    let f = fixture();
    let before = tree(&f.home, &[]);
    let plan = f.plan();
    assert_eq!(tree(&f.home, &[]), before, "the plan wrote something");
    assert_eq!(plan["writes"], json!(false));
    assert_eq!(plan["generation"], json!("go"));
    assert_eq!(plan["go_manifest"], json!("valid"));
    let a = actions(&plan);
    let archived: Vec<&str> = a
        .iter()
        .filter(|(_, (act, _))| act == "archive")
        .map(|(p, _)| p.as_str())
        .collect();
    assert_eq!(archived, ARCHIVED.to_vec(), "{plan:#}");
    let want_keep = [
        (".claude/skills/moai/c.md", "user_edited"),
        (".claude/skills/moai/user-note.md", "unrecorded"),
        (".claude/skills/moai/.env", "secret"),
        (".claude/skills/moai-extra", "unproven_files"),
        (".agents/skills/moai-mismatch", "unproven_files"),
        (".claude/skills/moai-link", "symlink"),
        (".claude/agents/moai-b.md", "user_edited"),
        (".moai/credentials.yaml", "secret"),
        (".moai/.env", "secret"),
        (".moai/state", "go_runtime"),
        (".moai/user-assets.json", "not_leftover"),
        (".moai/run", "not_leftover"),
        (".moai/memories", "not_in_table"),
        (".moai/tmp-fifo", "special_file"),
        (".moai/tmp-link", "symlink"),
    ];
    for (path, reason) in want_keep {
        assert_eq!(
            a.get(path),
            Some(&("keep".to_string(), reason.to_string())),
            "{path}: {plan:#}"
        );
    }
    for absent in [
        ".claude/agents/mine.md",
        ".claude/skills/other-skill",
        ".claude/skills/moai/a.md",
        ".claude/skills/moai/SKILL.md",
    ] {
        assert!(!a.contains_key(absent), "{absent} is listed");
    }
    // Secrets are never hashed.
    for i in plan["items"].as_array().unwrap() {
        if i["reason"] == "secret" {
            assert!(
                i.get("sha256").is_none() && i.get("tree_digest").is_none(),
                "{i}"
            );
        }
    }
    let item = |p: &str| {
        plan["items"]
            .as_array()
            .unwrap()
            .iter()
            .find(|i| i["path"] == p)
            .unwrap()
            .clone()
    };
    assert_eq!(
        item(".claude/skills/moai/b.md")["sha256"],
        json!(sha(b"b\n"))
    );
    assert!(item(".claude/skills/moai-proven")["tree_digest"].is_string());
    assert_eq!(plan["counts"]["archive"], json!(ARCHIVED.len()));
    let token = plan["confirm_token"].as_str().unwrap();
    assert!(
        token.starts_with("cleanup:") && token.len() == 24,
        "{token}"
    );
    // The same HOME plans the same token.
    assert_eq!(f.plan()["confirm_token"], plan["confirm_token"]);
}

#[test]
fn apply_needs_the_fresh_token_and_a_permission_reference() {
    let f = fixture();
    let before = tree(&f.home, &[]);
    let token = f.plan()["confirm_token"].as_str().unwrap().to_string();
    let (v, success) = f.apply("cleanup:0000000000000000", &[]);
    assert!(!success && v["ok"] == json!(false), "{v}");
    assert_eq!(tree(&f.home, &[]), before);
    // Without the flag the command line is refused.
    let (_, success) = run(
        &f.home,
        &["cleanup-apply", f.h(), f.s(), "--permission-reference", REF],
        None,
        &[],
    );
    assert!(!success);
    // A placeholder permission reference refuses.
    let (v, success) = run(
        &f.home,
        &[
            "cleanup-apply",
            f.h(),
            f.s(),
            "--confirm-cleanup",
            &token,
            "--permission-reference",
            "placeholder",
        ],
        None,
        &[],
    );
    assert!(!success && v["ok"] == json!(false), "{v}");
    // A change after the plan changes the token: the old one refuses.
    write(&f.home, ".moai/tmp-123/part", "partial, later\n");
    let before = tree(&f.home, &[]);
    let (v, success) = f.apply(&token, &[]);
    assert!(!success, "{v}");
    assert_eq!(tree(&f.home, &[]), before);
    assert!(!f.home.join(".moai-cli/cleanup").exists());
}

#[test]
fn apply_refuses_across_file_systems_before_moving() {
    let f = fixture();
    let before = tree(&f.home, &[]);
    let token = f.plan()["confirm_token"].as_str().unwrap().to_string();
    let (v, success) = f.apply(&token, &[("MOAI_CLI_TEST_CLEANUP_DEVICE", "other")]);
    assert!(!success, "{v}");
    assert!(
        v["error"].to_string().contains("another file system"),
        "{v}"
    );
    assert_eq!(tree(&f.home, &[]), before);
}

#[test]
fn apply_moves_the_archive_items_and_revert_restores_them_byte_for_byte() {
    let f = fixture();
    let before = f.snapshot();
    let plan = f.plan();
    let token = plan["confirm_token"].as_str().unwrap();
    let (v, success) = f.apply(token, &[]);
    assert!(success, "{v}");
    let r = &v["result"];
    assert_eq!(r["state"], json!("committed"));
    assert_eq!(r["stamp"], json!(STAMP));
    assert_eq!(r["moved"], json!(ARCHIVED.len()));
    let archive = f.home.join(".moai-cli/cleanup").join(STAMP);
    let manifest = fs::read_to_string(archive.join("manifest.tsv")).unwrap();
    let mut want: Vec<String> = ARCHIVED
        .iter()
        .map(|p| {
            format!(
                "{}\t{}",
                f.home.join(p).display(),
                archive.join(p).display()
            )
        })
        .collect();
    want.sort();
    let mut got: Vec<String> = manifest.lines().map(str::to_string).collect();
    got.sort();
    assert_eq!(got, want);
    let receipt: Value =
        serde_json::from_slice(&fs::read(archive.join("receipt.json")).unwrap()).unwrap();
    assert_eq!(receipt["status"], json!("committed"));
    assert!(receipt["items"]
        .as_array()
        .unwrap()
        .iter()
        .all(|i| i["moved"] == json!(true)));
    for p in ARCHIVED {
        assert!(!f.home.join(p).exists(), "{p} still in place");
        assert!(
            fs::symlink_metadata(archive.join(p)).is_ok(),
            "{p} not archived"
        );
    }
    // Everything else is where it was.
    let after = f.snapshot();
    let moved = |k: &String| {
        ARCHIVED
            .iter()
            .any(|p| k == p || k.starts_with(&format!("{p}/")))
    };
    let kept: BTreeMap<_, _> = before.iter().filter(|(k, _)| !moved(k)).collect();
    let after_kept: BTreeMap<_, _> = after.iter().collect();
    for (k, v) in &kept {
        assert_eq!(after_kept.get(k), Some(v), "{k} changed");
    }
    // A fresh plan archives nothing.
    let again = f.plan();
    assert_eq!(again["counts"]["archive"], json!(0), "{again:#}");
    assert_eq!(again["confirm_token"], Value::Null);

    let rp = f.revert_plan(STAMP);
    assert_eq!(rp["writes"], json!(false));
    assert_eq!(rp["counts"]["restore"], json!(ARCHIVED.len()));
    assert_eq!(rp["action"], json!("restore"));
    let (v, success) = f.revert(STAMP);
    assert!(success, "{v}");
    assert_eq!(v["result"]["restored"], json!(ARCHIVED.len()));
    assert_eq!(f.snapshot(), before, "the revert is not byte-identical");
    let receipt: Value =
        serde_json::from_slice(&fs::read(archive.join("receipt.json")).unwrap()).unwrap();
    assert_eq!(receipt["status"], json!("reverted"));
    // A rerun moves nothing and succeeds.
    let (v, success) = f.revert(STAMP);
    assert!(success, "{v}");
    assert_eq!(v["result"]["restored"], json!(0));
    assert_eq!(v["result"]["counts"]["done"], json!(ARCHIVED.len()));
    assert_eq!(f.snapshot(), before);
}

#[test]
fn revert_refuses_as_a_whole_on_a_conflict_or_a_missing_item() {
    let f = fixture();
    let token = f.plan()["confirm_token"].as_str().unwrap().to_string();
    assert!(f.apply(&token, &[]).1);
    let archive = f.home.join(".moai-cli/cleanup").join(STAMP);

    // Something new at an original path.
    write(&f.home, ".claude/agents/moai-a.md", "a new agent\n");
    let rp = f.revert_plan(STAMP);
    assert_eq!(rp["action"], json!("blocked"));
    assert_eq!(rp["counts"]["conflict"], json!(1));
    let before = tree(&f.home, &[]);
    let (v, success) = f.revert(STAMP);
    assert!(!success && v["ok"] == json!(false), "{v}");
    assert_eq!(
        tree(&f.home, &[]),
        before,
        "a refused revert moved something"
    );
    fs::remove_file(f.home.join(".claude/agents/moai-a.md")).unwrap();

    // An archived copy changed.
    write(&archive, ".moai/tmp-123/part", "tampered\n");
    let rp = f.revert_plan(STAMP);
    assert_eq!(rp["counts"]["missing"], json!(1), "{rp:#}");
    let before = tree(&f.home, &[]);
    let (_, success) = f.revert(STAMP);
    assert!(!success);
    assert_eq!(tree(&f.home, &[]), before);

    // A bad stamp and a linked receipt are refused.
    let (_, success) = run(
        &f.home,
        &["cleanup-revert-plan", f.h(), "--stamp", "../x"],
        None,
        &[],
    );
    assert!(!success);
    let receipt = archive.join("receipt.json");
    let real = archive.join("receipt.real");
    fs::rename(&receipt, &real).unwrap();
    std::os::unix::fs::symlink(&real, &receipt).unwrap();
    let (_, success) = run(
        &f.home,
        &["cleanup-revert-plan", f.h(), "--stamp", STAMP],
        None,
        &[],
    );
    assert!(!success);
}

#[test]
fn revert_refuses_when_a_parent_became_a_link_outside_home() {
    let f = fixture();
    let token = f.plan()["confirm_token"].as_str().unwrap().to_string();
    assert!(f.apply(&token, &[]).1);
    // After the apply, `.claude/agents` is swapped for a link to a folder outside HOME.
    let outside = f.home.parent().unwrap().join("outside");
    fs::create_dir_all(&outside).unwrap();
    let agents = f.home.join(".claude/agents");
    fs::rename(&agents, f.home.join(".claude/agents.real")).unwrap();
    std::os::unix::fs::symlink(&outside, &agents).unwrap();

    let rp = f.revert_plan(STAMP);
    assert_eq!(rp["action"], json!("blocked"), "{rp:#}");
    let linked: Vec<&Value> = rp["items"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|i| i["path"].as_str().unwrap().starts_with(".claude/agents/"))
        .collect();
    assert!(!linked.is_empty(), "{rp:#}");
    assert!(
        linked.iter().all(|i| i["status"] == json!("conflict")),
        "{rp:#}"
    );
    let (_, success) = f.revert(STAMP);
    assert!(!success);
    assert_eq!(
        fs::read_dir(&outside).unwrap().count(),
        0,
        "a refused revert wrote outside HOME"
    );
}

#[test]
fn a_second_apply_in_the_same_second_takes_the_next_folder() {
    let f = fixture();
    let token = f.plan()["confirm_token"].as_str().unwrap().to_string();
    assert!(f.apply(&token, &[]).1);
    write(&f.home, ".moai/tmp-new/x", "new\n");
    let plan = f.plan();
    assert_eq!(plan["counts"]["archive"], json!(1), "{plan:#}");
    let (v, success) = f.apply(plan["confirm_token"].as_str().unwrap(), &[]);
    assert!(success, "{v}");
    let second = format!("{STAMP}.2");
    assert_eq!(v["result"]["stamp"], json!(second));
    assert!(f
        .home
        .join(".moai-cli/cleanup")
        .join(&second)
        .join(".moai/tmp-new/x")
        .is_file());
}

#[test]
fn a_live_go_session_keeps_every_go_item() {
    let f = fixture();
    write(
        &f.home,
        ".moai/state/active-sessions.json",
        &json!([{"pid": std::process::id()}]).to_string(),
    );
    let a = actions(&f.plan());
    for p in [
        ".claude/skills/moai-proven",
        ".claude/agents/moai-a.md",
        ".moai/cache",
        ".claude/skills/moai/legacy.md",
    ] {
        assert_eq!(a[p], ("keep".into(), "in_use".into()), "{p}");
    }
    // What the Rust host receipt proves is not Go state.
    assert_eq!(a[".claude/skills/moai/b.md"].0, "archive");
}

#[test]
fn an_unreadable_go_manifest_proves_nothing() {
    let f = fixture();
    write(&f.home, ".moai/user-assets.json", "{not json");
    let plan = f.plan();
    assert_eq!(plan["go_manifest"], json!("unreadable"));
    let a = actions(&plan);
    assert_eq!(
        a[".claude/skills/moai-proven"],
        ("keep".into(), "go_manifest_unreadable".into())
    );
    assert_eq!(
        a[".claude/skills/moai/legacy.md"],
        ("keep".into(), "go_manifest_unreadable".into())
    );
    assert!(!a.contains_key(".claude/agents/moai-a.md"));
    assert_eq!(a[".claude/skills/moai/b.md"].0, "archive");
}

#[test]
fn a_linked_skills_folder_is_kept_and_never_followed() {
    let f = fixture();
    let skills = f.home.join(".agents/skills");
    fs::rename(&skills, f.home.join("agents-skills-real")).unwrap();
    std::os::unix::fs::symlink(f.home.join("agents-skills-real"), &skills).unwrap();
    let a = actions(&f.plan());
    assert_eq!(a[".agents/skills"], ("keep".into(), "symlink".into()));
    assert_eq!(a[".agents/skills/moai"], ("keep".into(), "symlink".into()));
    assert!(!a.keys().any(|k| k.starts_with(".agents/skills/moai/")));
}
