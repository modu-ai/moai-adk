//! #2561: Go-generation project state — the read-only preview. The Go manifest
//! (`.moai/manifest.json`) is untrusted input: only a regular, unlinked file it records as
//! `template_managed` whose bytes still hash to its `deployed_hash` is proven tool-made and
//! offered for moving; everything else stays, with a reason. Every case builds a temporary
//! project; the real checkout is never read.

#![cfg(unix)]

use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::{
    fs,
    os::unix::fs::symlink,
    path::{Path, PathBuf},
    process::{Command, Output},
};
use tempfile::TempDir;

const EXE: &str = env!("CARGO_BIN_EXE_moai-cli");

fn sha(bytes: &[u8]) -> String {
    format!("sha256:{:x}", Sha256::digest(bytes))
}

struct Fx {
    _tmp: TempDir,
    root: PathBuf,
    files: serde_json::Map<String, Value>,
}

impl Fx {
    fn new() -> Fx {
        let tmp = TempDir::new().unwrap();
        let root = fs::canonicalize(tmp.path()).unwrap().join("project");
        fs::create_dir_all(&root).unwrap();
        Fx {
            _tmp: tmp,
            root,
            files: serde_json::Map::new(),
        }
    }

    fn write(&self, rel: &str, body: &str) {
        let p = self.root.join(rel);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        fs::write(p, body).unwrap();
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

    fn tree(&self) -> Vec<(PathBuf, Vec<u8>)> {
        let mut out = Vec::new();
        let mut stack = vec![self.root.clone()];
        while let Some(d) = stack.pop() {
            for e in fs::read_dir(&d).unwrap() {
                let p = e.unwrap().path();
                let m = fs::symlink_metadata(&p).unwrap();
                if m.is_dir() {
                    stack.push(p);
                } else if m.file_type().is_symlink() {
                    out.push((
                        p.clone(),
                        fs::read_link(&p)
                            .unwrap()
                            .into_os_string()
                            .into_encoded_bytes(),
                    ));
                } else {
                    out.push((p.clone(), fs::read(&p).unwrap()));
                }
            }
        }
        out.sort();
        out
    }
}

fn run(args: &[&str]) -> Output {
    Command::new(EXE)
        .args(args)
        .env_clear()
        .env("PATH", "/usr/bin:/bin")
        .env("LC_ALL", "C")
        .output()
        .unwrap()
}

fn plan(fx: &Fx) -> Value {
    let before = fx.tree();
    let out = run(&["go-state-plan", fx.root.to_str().unwrap()]);
    let v: Value = serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
        panic!(
            "{}{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    });
    assert_eq!(v["ok"], json!(true), "{v}");
    assert_eq!(fx.tree(), before, "a preview writes nothing");
    v["result"].clone()
}

fn action<'a>(plan: &'a Value, rel: &str) -> (&'a str, &'a str) {
    let item = plan["items"]
        .as_array()
        .unwrap()
        .iter()
        .find(|i| i["path"] == rel)
        .unwrap_or_else(|| panic!("{rel} not in plan: {plan}"));
    (
        item["action"].as_str().unwrap(),
        item["reason"].as_str().unwrap(),
    )
}

#[test]
fn only_unchanged_template_files_are_offered_for_moving() {
    let mut fx = Fx::new();
    fx.deployed(".moai/config/sections/a.yaml", "a: 1\n");
    fx.deployed(".claude/workflows/w.md", "# w\n");
    fx.deployed(".moai/config/sections/b.yaml", "b: 1\n");
    fx.write(".moai/config/sections/b.yaml", "b: 2 # edited\n");
    fx.write("AGENTS.md", "mine\n");
    fx.record("AGENTS.md", "user_created", &sha(b"mine\n"));
    fx.write(".moai/config/sections/u.yaml", "u: 1\n");
    fx.record(
        ".moai/config/sections/u.yaml",
        "user_modified",
        &sha(b"u: 1\n"),
    );
    fx.write(".moai/specs/SPEC-1/spec.md", "# spec\n");
    fx.manifest();
    let p = plan(&fx);
    assert_eq!(p["manifest_version"], json!("v3.2.1-rc.6"));
    assert_eq!(action(&p, ".moai/config/sections/a.yaml").0, "move");
    assert_eq!(action(&p, ".claude/workflows/w.md").0, "move");
    assert_eq!(
        action(&p, ".moai/config/sections/b.yaml"),
        ("leave", "changed since the Go deploy")
    );
    assert_eq!(
        action(&p, ".moai/config/sections/u.yaml"),
        ("leave", "not template_managed")
    );
    assert_eq!(
        action(&p, "AGENTS.md"),
        ("leave", "outside .moai/ and .claude/")
    );
    // Not recorded by the Go manifest: never moved, counted under its folder.
    assert_eq!(p["unrecorded"][".moai/specs"], json!(1));
    assert!(p["plan_digest"].as_str().unwrap().starts_with("sha256:"));
}

#[test]
fn links_and_escaping_keys_are_never_moved() {
    let mut fx = Fx::new();
    fs::create_dir_all(fx.root.join("elsewhere")).unwrap();
    fs::write(fx.root.join("elsewhere/h.sh"), "#!/bin/sh\n").unwrap();
    fs::create_dir_all(fx.root.join(".claude")).unwrap();
    symlink(fx.root.join("elsewhere"), fx.root.join(".claude/hooks")).unwrap();
    fx.record(
        ".claude/hooks/h.sh",
        "template_managed",
        &sha(b"#!/bin/sh\n"),
    );
    fx.write(".moai/config/real.yaml", "x\n");
    symlink(
        fx.root.join(".moai/config/real.yaml"),
        fx.root.join(".moai/config/link.yaml"),
    )
    .unwrap();
    fx.record(".moai/config/link.yaml", "template_managed", &sha(b"x\n"));
    fx.record(".moai/../escape.txt", "template_managed", &sha(b""));
    fx.record("/etc/passwd", "template_managed", &sha(b""));
    fx.manifest();
    let p = plan(&fx);
    assert_eq!(
        action(&p, ".claude/hooks/h.sh"),
        ("leave", "a path component is a symlink")
    );
    assert_eq!(
        action(&p, ".moai/config/link.yaml"),
        ("leave", "a path component is a symlink")
    );
    assert_eq!(
        action(&p, ".moai/../escape.txt"),
        ("leave", "not a plain relative path")
    );
    assert_eq!(
        action(&p, "/etc/passwd"),
        ("leave", "not a plain relative path")
    );
}

#[test]
fn files_settings_still_reference_stay() {
    let mut fx = Fx::new();
    fx.deployed(".claude/hooks/moai/handle-stop.sh", "#!/bin/sh\n");
    fx.deployed(
        ".claude/output-styles/moai/MoAI-Easy.md",
        "---\nname: MoAI-Easy\n---\n",
    );
    fx.deployed(".claude/workflows/free.md", "# free\n");
    fx.write(
        ".claude/settings.json",
        &json!({"hooks": {"Stop": [{"hooks": [{"type": "command",
            "command": "\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-stop.sh\""}]}]},
            "outputStyle": "MoAI-Easy"})
        .to_string(),
    );
    fx.manifest();
    let p = plan(&fx);
    assert_eq!(
        action(&p, ".claude/hooks/moai/handle-stop.sh"),
        ("leave", "referenced by .claude/settings.json")
    );
    assert_eq!(
        action(&p, ".claude/output-styles/moai/MoAI-Easy.md"),
        ("leave", "referenced by .claude/settings.json")
    );
    assert_eq!(action(&p, ".claude/workflows/free.md").0, "move");
}

#[test]
fn without_a_valid_manifest_nothing_moves() {
    for body in [
        None,
        Some("not json"),
        Some(r#"{"version":"v3","files":[]}"#),
    ] {
        let fx = Fx::new();
        fx.write(".moai/config/a.yaml", "a\n");
        if let Some(b) = body {
            fx.write(".moai/manifest.json", b);
        }
        let p = plan(&fx);
        assert_eq!(p["items"], json!([]), "{body:?}: {p}");
        assert!(
            p["warnings"]
                .as_array()
                .unwrap()
                .iter()
                .any(|w| w.as_str().unwrap().contains("no usable Go manifest")),
            "{p}"
        );
    }
}

#[test]
fn a_linked_moai_folder_or_manifest_is_refused() {
    let mut fx = Fx::new();
    fx.deployed(".moai/config/a.yaml", "a\n");
    fx.manifest();
    let real = fx.root.join("real-moai");
    fs::rename(fx.root.join(".moai"), &real).unwrap();
    symlink(&real, fx.root.join(".moai")).unwrap();
    let out = run(&["go-state-plan", fx.root.to_str().unwrap()]);
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], json!(false), "{v}");
    assert!(v.to_string().contains("symlink"), "{v}");
    let _ = Path::new("");
}
