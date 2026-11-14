//! #2936 R-fix: a Go-generation project built at test run time, and the runners the tests
//! share. A test crate pulls it in with
//! `#[path = "common/go_project_real_run.rs"] mod go_project_real_run;` beside `mod common;`.
//!
//! The builder writes files into a temporary project with an empty temporary HOME; it reads
//! nothing from the real checkout or the real HOME. Every file a test records in the Go
//! manifest is written first and hashed from its bytes, so the recorded `deployed_hash` is
//! the real one.

#![allow(dead_code)]

use crate::common::{CleanEnv, MOAI_BIN};
use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::{
    fs,
    path::{Path, PathBuf},
    process::{Command, Stdio},
};
use tempfile::TempDir;

pub const REF: &str = "test:2936-rfix";

pub fn sha(bytes: &[u8]) -> String {
    format!("sha256:{:x}", Sha256::digest(bytes))
}

/// One tree entry, never followed: a file's bytes, a link's target, or a folder.
#[derive(Debug, PartialEq, Eq, PartialOrd, Ord, Clone)]
pub enum Node {
    File(Vec<u8>),
    Link(PathBuf),
    Dir,
}

pub struct GoFx {
    _tmp: TempDir,
    pub home: PathBuf,
    pub root: PathBuf,
    files: serde_json::Map<String, Value>,
}

impl GoFx {
    pub fn new() -> GoFx {
        let tmp = TempDir::new().unwrap();
        let base = fs::canonicalize(tmp.path()).unwrap();
        let (home, root) = (base.join("home"), base.join("project"));
        fs::create_dir_all(&home).unwrap();
        fs::create_dir_all(&root).unwrap();
        GoFx {
            _tmp: tmp,
            home,
            root,
            files: serde_json::Map::new(),
        }
    }

    pub fn p(&self) -> &str {
        self.root.to_str().unwrap()
    }

    pub fn write(&self, rel: &str, body: &str) {
        let p = self.root.join(rel);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        fs::write(p, body).unwrap();
    }

    /// A symbolic link at `rel` whose body is `body`.
    pub fn link(&self, rel: &str, body: &str) {
        let p = self.root.join(rel);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        std::os::unix::fs::symlink(body, p).unwrap();
    }

    /// A file the Go deployer wrote and recorded as `template_managed`.
    pub fn deployed(&mut self, rel: &str, body: &str) {
        self.write(rel, body);
        let h = sha(body.as_bytes());
        self.files.insert(
            rel.into(),
            json!({"provenance": "template_managed", "template_hash": h,
                   "deployed_hash": h, "current_hash": h}),
        );
    }

    /// Writes `.moai/manifest.json` from everything recorded so far.
    pub fn manifest(&self) {
        self.write(
            ".moai/manifest.json",
            &json!({"version": "v3.2.1-rc.6", "deployed_at": "2026-11-01T00:00:00Z",
                    "files": self.files})
            .to_string(),
        );
    }

    /// Every entry under the project and HOME; links read as links, never followed.
    pub fn tree(&self) -> Vec<(PathBuf, Node)> {
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

    pub fn run_env(&self, args: &[&str], env: &[(&str, &str)]) -> Value {
        let mut cmd = Command::new(MOAI_BIN);
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
        let out = cmd.output().unwrap();
        serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
            panic!(
                "{args:?}: {}{}",
                String::from_utf8_lossy(&out.stdout),
                String::from_utf8_lossy(&out.stderr)
            )
        })
    }

    pub fn run(&self, args: &[&str]) -> Value {
        self.run_env(args, &[])
    }

    /// `go-state-plan` (the `result`), with the whole tree compared before and after.
    pub fn plan_env(&self, env: &[(&str, &str)]) -> Value {
        let before = self.tree();
        let v = self.run_env(&["go-state-plan", self.p()], env);
        assert_eq!(v["ok"], json!(true), "{v}");
        assert_eq!(self.tree(), before, "a preview writes nothing");
        v["result"].clone()
    }

    pub fn plan(&self) -> Value {
        self.plan_env(&[])
    }

    /// `project-update-plan` (the `result`), with the whole tree compared before and after.
    pub fn update_plan_env(&self, env: &[(&str, &str)]) -> Value {
        let before = self.tree();
        let v = self.run_env(&["project-update-plan", self.p()], env);
        assert_eq!(v["ok"], json!(true), "{v}");
        assert_eq!(self.tree(), before, "a preview writes nothing");
        v["result"].clone()
    }

    pub fn update_plan(&self) -> Value {
        self.update_plan_env(&[])
    }

    /// Runs the binary with `stdin` and returns the whole envelope.
    pub fn run_stdin(&self, args: &[&str], stdin: &[u8]) -> Value {
        use std::io::Write as _;
        let mut child = Command::new(MOAI_BIN)
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

    /// A call that must succeed; returns its `result`.
    pub fn call(&self, args: &[&str], stdin: &[u8]) -> Value {
        let v = self.run_stdin(args, stdin);
        assert_eq!(v["ok"], json!(true), "{args:?}: {v}");
        v["result"].clone()
    }

    /// A profile written the way PROJECT writes it: the detected facts, then `project-plan`
    /// and `project-apply`.
    pub fn with_profile(&self) {
        let detect = self.call(&["project-detect", self.p()], b"");
        let profile = json!({
            "schema_version": 2, "project_id": "demo", "purpose": "Build a small editor",
            "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude"],
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

    /// Runs the command a step of `project-update-plan` prints, exactly as printed, with the
    /// permission reference in place of the one placeholder.
    pub fn apply_step(&self, step: &Value) -> Value {
        let argv: Vec<String> = step["apply"]["argv"]
            .as_array()
            .unwrap_or_else(|| panic!("the step prints a command: {step}"))
            .iter()
            .map(|w| w.as_str().unwrap().to_string())
            .collect();
        assert_eq!(argv[0], "moai", "{argv:?}");
        let args: Vec<&str> = argv[1..]
            .iter()
            .map(|w| {
                if w == "<permission-reference>" {
                    REF
                } else {
                    w.as_str()
                }
            })
            .collect();
        let stdin = match &step["apply"]["stdin"] {
            Value::Null => Vec::new(),
            v => serde_json::to_vec(v).unwrap(),
        };
        self.run_stdin(&args, &stdin)
    }

    pub fn apply(&self, digest: &str) -> Value {
        self.run(&[
            "go-state-apply",
            self.p(),
            "--plan-digest",
            digest,
            "--permission-reference",
            REF,
        ])
    }
}

/// The item for `rel` in a `go-state-plan` result.
pub fn item<'a>(plan: &'a Value, rel: &str) -> &'a Value {
    plan["items"]
        .as_array()
        .unwrap()
        .iter()
        .find(|i| i["path"] == rel)
        .unwrap_or_else(|| panic!("{rel} not in plan: {plan}"))
}

pub fn action<'a>(plan: &'a Value, rel: &str) -> &'a str {
    item(plan, rel)["action"].as_str().unwrap()
}

/// The `preview` of the go-state step in a `project-update-plan` result.
pub fn go_state_preview(update: &Value) -> &Value {
    let backup = &update["phases"][0];
    assert_eq!(backup["phase"], json!("backup"), "{update}");
    &backup["steps"][0]["preview"]
}

/// The step named `name` in the install phase of a `project-update-plan` result.
pub fn install_step<'a>(update: &'a Value, name: &str) -> Option<&'a Value> {
    update["phases"][1]["steps"]
        .as_array()?
        .iter()
        .find(|s| s["step"] == name)
}

pub fn path_in(root: &Path, rel: &str) -> PathBuf {
    root.join(rel)
}
