//! #2936 S1: `moai project-update-plan <project>` — one read-only preview of the project
//! update, grouped into the three phases of the user's UPDATE flow (backup, install, merge).
//! Every case runs the built binary in a temporary project with an empty temporary HOME, and
//! every plan leaves the project tree byte-identical; the real checkout and HOME are never read.

#![cfg(unix)]

mod common;
use common::CleanEnv;
use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::{
    fs,
    io::Write,
    path::{Path, PathBuf},
    process::{Command, Stdio},
};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;
const REF: &str = "test:2936";
const PLACEHOLDER: &str = "<permission-reference>";
const BLOCKED_BY_BACKUP: &str = "re-run after the backup phase";

struct Fx {
    _tmp: TempDir,
    home: PathBuf,
    root: PathBuf,
}

/// One tree entry: a file's bytes, a link's target, or a folder.
#[derive(Debug, PartialEq, Eq, PartialOrd, Ord)]
enum Node {
    File(Vec<u8>),
    Link(PathBuf),
    Dir,
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

    fn p(&self) -> &str {
        self.root.to_str().unwrap()
    }

    fn write(&self, rel: &str, body: &str) {
        let p = self.root.join(rel);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        fs::write(p, body).unwrap();
    }

    /// Runs the binary with `stdin` and returns the whole envelope.
    fn run(&self, args: &[&str], stdin: &[u8]) -> Value {
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
        let mut pipe = child.stdin.take().unwrap();
        match pipe.write_all(stdin) {
            Ok(()) => {}
            Err(e) if e.kind() == std::io::ErrorKind::BrokenPipe => {}
            Err(e) => panic!("stdin write failed: {e}"),
        }
        drop(pipe);
        let out = child.wait_with_output().unwrap();
        serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
            panic!(
                "{args:?}: {}{}",
                String::from_utf8_lossy(&out.stdout),
                String::from_utf8_lossy(&out.stderr)
            )
        })
    }

    fn call(&self, args: &[&str], stdin: &[u8]) -> Value {
        let v = self.run(args, stdin);
        assert_eq!(v["ok"], json!(true), "{args:?}: {v}");
        v["result"].clone()
    }

    /// Every entry under the project and HOME, links read as links, never followed.
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
                    out.push((p.clone(), Node::Dir));
                    stack.push(p);
                } else {
                    out.push((p.clone(), Node::File(fs::read(&p).unwrap())));
                }
            }
        }
        out.sort();
        out
    }

    /// `project-update-plan`, with the whole tree compared before and after.
    fn plan(&self) -> Value {
        let before = self.tree();
        let plan = self.call(&["project-update-plan", self.p()], b"");
        assert_eq!(before, self.tree(), "project-update-plan wrote something");
        plan
    }

    /// A profile written the way PROJECT writes it: the detected facts, then
    /// `project-plan` and `project-apply`.
    fn with_profile(&self, github_issues: bool) {
        let detect = self.call(&["project-detect", self.p()], b"");
        let mut profile = json!({
            "schema_version": 2, "project_id": "demo", "purpose": "Build a small editor",
            "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude", "codex"],
            "detected": detect["detected"],
        });
        if github_issues {
            profile["github_issues"] = json!(true);
        }
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

    /// A Go-generation project (the `tests/update_route.rs` fixture): recorded files the
    /// Go deployer left unchanged, one the user edited, and a SPEC folder.
    fn go_project(&self) {
        let mut files = serde_json::Map::new();
        for (rel, body) in [
            (".moai/config/sections/a.yaml", "a: 1\n"),
            (".claude/workflows/w.md", "# w\n"),
            (".moai/config/sections/b.yaml", "b: 1\n"),
        ] {
            self.write(rel, body);
            let h = format!("sha256:{:x}", Sha256::digest(body.as_bytes()));
            files.insert(
                rel.into(),
                json!({"provenance": "template_managed", "template_hash": h,
                       "deployed_hash": h, "current_hash": h}),
            );
        }
        self.write(".moai/config/sections/b.yaml", "b: 2 # edited\n");
        self.write(".moai/specs/SPEC-1/spec.md", "# spec\n");
        self.write(
            ".moai/manifest.json",
            &json!({"version": "v3.2.1-rc.6", "deployed_at": "2026-11-01T00:00:00Z",
                    "files": files})
            .to_string(),
        );
    }

    /// Runs one step's printed command: `moai` is this binary and the placeholder is a
    /// permission reference; the printed stdin goes on stdin as it is.
    fn apply_step(&self, step: &Value) -> Value {
        let argv: Vec<String> = step["apply"]["argv"]
            .as_array()
            .unwrap_or_else(|| panic!("a ready step prints its command: {step}"))
            .iter()
            .map(|w| w.as_str().unwrap().to_string())
            .collect();
        assert_eq!(argv[0], "moai", "{argv:?}");
        assert_eq!(
            argv.iter().filter(|w| w.contains('<')).collect::<Vec<_>>(),
            vec![PLACEHOLDER],
            "the only placeholder is the permission reference: {argv:?}"
        );
        let args: Vec<&str> = argv[1..]
            .iter()
            .map(|w| if w == PLACEHOLDER { REF } else { w.as_str() })
            .collect();
        let stdin = match &step["apply"]["stdin"] {
            Value::Null => Vec::new(),
            v => serde_json::to_vec(v).unwrap(),
        };
        self.run(&args, &stdin)
    }
}

fn phase<'a>(plan: &'a Value, name: &str) -> &'a Value {
    plan["phases"]
        .as_array()
        .unwrap()
        .iter()
        .find(|p| p["phase"] == name)
        .unwrap_or_else(|| panic!("no phase {name}: {plan}"))
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

fn steps(plan: &Value) -> Vec<&Value> {
    plan["phases"]
        .as_array()
        .unwrap()
        .iter()
        .flat_map(|p| p["steps"].as_array().into_iter().flatten())
        .collect()
}

const INSTALL: [&str; 4] = ["profile", "generation", "claude-settings", "github"];

#[test]
fn empty_project() {
    let fx = Fx::new();
    let plan = fx.plan();
    let names: Vec<&str> = plan["phases"]
        .as_array()
        .unwrap()
        .iter()
        .map(|p| p["phase"].as_str().unwrap())
        .collect();
    assert_eq!(names, ["backup", "install", "merge"], "{plan}");
    for s in steps(&plan) {
        for key in ["step", "state", "reason", "preview", "plan_digest", "apply"] {
            assert!(s.get(key).is_some(), "step lacks {key}: {s}");
        }
    }
    assert_eq!(step(&plan, "go-state")["state"], "unchanged", "{plan}");
    for name in INSTALL {
        let s = step(&plan, name);
        assert_eq!(s["state"], "needs_decision", "{name}: {s}");
        assert_eq!(s["reason"], "run /moai project", "{name}: {s}");
        assert_eq!(s["apply"], Value::Null, "{name}: {s}");
    }
    assert_eq!(phase(&plan, "install")["state"], "needs_decision");
    let merge = phase(&plan, "merge");
    assert_eq!(merge["state"], "not_run", "{merge}");
    assert_eq!(merge["candidates"], json!([]), "{merge}");
    assert_eq!(plan["all_unchanged"], json!(false));
}

#[test]
fn printed_commands_apply() {
    let fx = Fx::new();
    fx.with_profile(true);
    let plan = fx.plan();
    assert_eq!(step(&plan, "profile")["state"], "unchanged", "{plan}");
    assert_eq!(step(&plan, "generation")["state"], "ready", "{plan}");
    assert_eq!(step(&plan, "github")["state"], "ready", "{plan}");
    let mut applied = 0;
    for s in steps(&plan).into_iter().filter(|s| s["state"] == "ready") {
        let out = fx.apply_step(s);
        assert_eq!(out["ok"], json!(true), "{}: {out}", s["step"]);
        applied += 1;
    }
    assert!(applied >= 2, "{plan}");
}

#[test]
fn rerun_unchanged() {
    let fx = Fx::new();
    fx.with_profile(true);
    let plan = fx.plan();
    for s in steps(&plan).into_iter().filter(|s| s["state"] == "ready") {
        assert_eq!(fx.apply_step(s)["ok"], json!(true), "{s}");
    }
    let again = fx.plan();
    for s in steps(&again) {
        assert!(
            s["state"] == "unchanged" || s["state"] == "skip",
            "{}: {s}",
            s["step"]
        );
        assert_eq!(s["apply"], Value::Null, "{s}");
    }
    // #2936 Q3 B: the generation manifest reads `unchanged` when its bytes would not move.
    let generation = step(&again, "generation");
    assert!(
        generation["preview"]["files"]
            .as_array()
            .unwrap()
            .iter()
            .any(|f| f["path"]
                .as_str()
                .unwrap()
                .ends_with("manifests/generation.json")
                && f["action"] == "unchanged"),
        "{generation}"
    );
    assert_eq!(again["all_unchanged"], json!(true), "{again}");
}

#[test]
fn one_blocked_step() {
    let fx = Fx::new();
    fx.with_profile(false);
    let outside = fx.home.parent().unwrap().join("elsewhere.json");
    fs::write(&outside, "{}\n").unwrap();
    fs::create_dir_all(fx.root.join(".claude")).unwrap();
    std::os::unix::fs::symlink(&outside, fx.root.join(".claude/settings.json")).unwrap();
    let plan = fx.plan();
    let settings = step(&plan, "claude-settings");
    assert_eq!(settings["state"], "blocked", "{settings}");
    assert!(
        !settings["reason"].as_str().unwrap().is_empty(),
        "{settings}"
    );
    assert_eq!(settings["apply"], Value::Null);
    assert_eq!(step(&plan, "generation")["state"], "ready", "{plan}");
    assert_eq!(step(&plan, "profile")["state"], "unchanged", "{plan}");
    assert_eq!(phase(&plan, "install")["state"], "blocked", "{plan}");
    assert_eq!(plan["all_unchanged"], json!(false));
}

#[test]
fn owned_settings_kept() {
    let fx = Fx::new();
    fx.with_profile(false);
    let desired = br#"{"outputStyle":"Explanatory"}"#;
    let plan = fx.call(&["claude-settings-plan", fx.p()], desired);
    let digest = plan["plan_digest"].as_str().unwrap().to_string();
    fx.call(
        &[
            "claude-settings-apply",
            fx.p(),
            "--plan-digest",
            &digest,
            "--permission-reference",
            REF,
        ],
        desired,
    );
    let settings_before = fs::read(fx.root.join(".claude/settings.json")).unwrap();
    let plan = fx.plan();
    let settings = step(&plan, "claude-settings");
    assert_eq!(settings["state"], "unchanged", "{settings}");
    assert_eq!(
        fs::read(fx.root.join(".claude/settings.json")).unwrap(),
        settings_before
    );
}

#[test]
fn backup_phase_first() {
    let fx = Fx::new();
    fx.go_project();
    let plan = fx.plan();
    let go = step(&plan, "go-state");
    // A move asks the user first (PROJECT step 2), so it is a decision with its command.
    assert_eq!(go["state"], "needs_decision", "{go}");
    assert_eq!(phase(&plan, "backup")["state"], "needs_decision", "{plan}");
    assert!(
        phase(&plan, "backup")["reason"]
            .as_str()
            .unwrap()
            .contains(".agents"),
        "{plan}"
    );
    assert_eq!(go["apply"]["argv"][1], "go-state-apply", "{go}");
    assert_eq!(go["apply"]["stdin"], Value::Null, "{go}");
    for name in INSTALL {
        let s = step(&plan, name);
        assert_eq!(s["state"], "blocked", "{name}: {s}");
        assert_eq!(s["reason"], BLOCKED_BY_BACKUP, "{name}: {s}");
        assert_eq!(s["apply"], Value::Null, "{name}: {s}");
    }
    assert_eq!(plan["all_unchanged"], json!(false));

    // Applying the printed backup command, then planning again: the backup is done and the
    // install phase is no longer held back by it.
    let out = fx.apply_step(go);
    assert_eq!(out["ok"], json!(true), "{out}");
    let again = fx.plan();
    assert_eq!(step(&again, "go-state")["state"], "unchanged", "{again}");
    for name in INSTALL {
        assert_ne!(
            step(&again, name)["reason"],
            BLOCKED_BY_BACKUP,
            "{name}: {again}"
        );
    }
}

/// The printed argv is a word list that names the project as one word (its full path).
#[test]
fn printed_argv_keeps_the_project_one_word() {
    let fx = Fx::new();
    fx.with_profile(false);
    let plan = fx.plan();
    let generation = step(&plan, "generation");
    let argv = generation["apply"]["argv"].as_array().unwrap();
    let project = fs::canonicalize(Path::new(fx.p())).unwrap();
    assert!(
        argv.iter().any(|w| w == project.to_str().unwrap()),
        "{argv:?}"
    );
}
