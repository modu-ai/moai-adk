//! SPEC-20261121-go-residue-detect: the read-only `go-residue` step of
//! `moai project-update-plan` finds the files the Go generation left behind, proves them
//! against the first manifest it can find (`.moai/manifest.json`, a `.moai-backup*` manifest,
//! a `.moai/backups/go-markers-*` manifest, or the backup the last committed go-state receipt
//! names), sorts them into `go_same` / `go_changed` / `user_data`, counts secret-shaped names
//! without opening them, never follows a link, and writes nothing. The frame is the
//! Go-generation survey's section 8 temp project; the real checkout is never read.

#![cfg(unix)]

mod common;
use common::{CleanEnv, MOAI_BIN};
use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::{
    fs,
    path::{Path, PathBuf},
    process::{Command, Stdio},
};
use tempfile::TempDir;

fn sha(body: &str) -> String {
    format!("sha256:{:x}", Sha256::digest(body.as_bytes()))
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
        let fx = Fx {
            _tmp: tmp,
            home,
            root,
            files: serde_json::Map::new(),
        };
        // A valid Rust state: the marker and the minimal profile.
        fx.write(
            ".moai/moai.json",
            &json!({"generation": "moai", "created_by": "moai-cli test",
                    "go_backup": ".moai-backup"})
            .to_string(),
        );
        fx.write(".moai/project.yaml", "schema_version: 2\n");
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

    /// A link inside the state folder; the target stays as written, never touched.
    fn link(&self, rel: &str, target: &str) {
        let p = self.root.join(rel);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        std::os::unix::fs::symlink(target, p).unwrap();
    }

    /// A file the Go deployer wrote and recorded in the manifest under construction.
    fn deployed(&mut self, rel: &str, body: &str) {
        self.write(rel, body);
        self.files.insert(
            rel.into(),
            json!({"provenance": "template_managed", "template_hash": sha(body),
                   "deployed_hash": sha(body), "current_hash": sha(body)}),
        );
    }

    /// Recorded but rewritten by the user: the manifest still names the old bytes.
    fn changed(&mut self, rel: &str, recorded: &str, on_disk: &str) {
        self.write(rel, on_disk);
        self.files.insert(
            rel.into(),
            json!({"provenance": "template_managed", "template_hash": sha(recorded),
                   "deployed_hash": sha(recorded), "current_hash": sha(recorded)}),
        );
    }

    /// Writes the manifest under `rel` (the live one, a backup, or a go-markers copy).
    fn manifest_at(&self, rel: &str) {
        self.write(
            rel,
            &json!({"version": "v3.2.1-rc.6", "deployed_at": "2026-11-01T00:00:00Z",
                    "files": self.files})
            .to_string(),
        );
    }

    fn git_commit_all(&self) {
        common::git::git_ok(&self.root, &["init", "-q"]);
        common::git::git_ok(&self.root, &["add", "-A"]);
        common::git::git_ok(
            &self.root,
            &[
                "-c",
                "user.name=fixture",
                "-c",
                "user.email=fixture@example.invalid",
                "commit",
                "-q",
                "-m",
                "fixture",
            ],
        );
    }

    /// Every path under the project and HOME, links named as links, never followed.
    fn tree(&self) -> Vec<(PathBuf, String)> {
        fn node(p: &Path) -> String {
            let m = fs::symlink_metadata(p).unwrap();
            if m.file_type().is_symlink() {
                format!("link:{}", fs::read_link(p).unwrap().display())
            } else if m.is_dir() {
                "dir".into()
            } else {
                format!("file:{}", m.len())
            }
        }
        let mut out = Vec::new();
        let mut stack = vec![self.root.clone(), self.home.clone()];
        while let Some(d) = stack.pop() {
            for e in fs::read_dir(&d).unwrap() {
                let e = e.unwrap();
                let p = e.path();
                if p.is_dir() && !p.is_symlink() {
                    stack.push(p);
                } else {
                    let kind = node(&p);
                    out.push((p, kind));
                }
            }
        }
        out.sort();
        out
    }

    fn run(&self, args: &[&str]) -> Value {
        let out = Command::new(MOAI_BIN)
            .args(args)
            .clean_env()
            .env("HOME", &self.home)
            .env("PATH", "/usr/bin:/bin")
            .env("LC_ALL", "C")
            .stdin(Stdio::null())
            .output()
            .expect("spawn the binary");
        serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
            panic!(
                "{args:?}: {}{}",
                String::from_utf8_lossy(&out.stdout),
                String::from_utf8_lossy(&out.stderr)
            )
        })
    }

    /// `project-update-plan`, with the whole tree compared before and after.
    fn update_plan(&self) -> Value {
        let before = self.tree();
        let v = self.run(&["project-update-plan", self.p()]);
        assert_eq!(self.tree(), before, "project-update-plan wrote something");
        v
    }

    /// The item of the `go-residue` step whose `path` ends with `name`.
    fn residue_item<'v>(plan: &'v Value, name: &str) -> &'v Value {
        step(plan, "go-residue")["preview"]["items"]
            .as_array()
            .expect("go-residue step has items")
            .iter()
            .find(|i| i["path"].as_str().is_some_and(|p| p.ends_with(name)))
            .unwrap_or_else(|| panic!("no go-residue item for {name}"))
    }
}

fn step<'v>(plan: &'v Value, name: &str) -> &'v Value {
    // The command answers its envelope; the plan body lives under `result`.
    let body = plan.get("result").unwrap_or(plan);
    body["phases"]
        .as_array()
        .expect("plan has phases")
        .iter()
        .flat_map(|p| p["steps"].as_array().into_iter().flatten())
        .find(|s| s["step"] == json!(name))
        .unwrap_or_else(|| {
            panic!(
                "no `{name}` step in the plan: {:?}",
                plan["phases"].as_array().map(|a| {
                    a.iter()
                        .flat_map(|p| p["steps"].as_array().into_iter().flatten())
                        .filter_map(|s| s["step"].as_str())
                        .collect::<Vec<_>>()
                })
            )
        })
}

/// The survey's section 8 frame: a git-tracked project with a valid Rust state, one live Go
/// manifest, byte-identical deployed files, one user-changed file, one unmanifested user
/// file, one secret-shaped name, and an output style a settings file references.
fn survey_frame() -> Fx {
    let mut fx = Fx::new();
    fx.deployed(".moai/policies/core/a.md", "# core policy\n");
    fx.deployed(".moai/workflows/gtd.md", "# gtd workflow\n");
    fx.deployed(
        ".moai/output-styles/moai/moai-easy.md",
        "# moai easy style\n",
    );
    fx.changed(
        ".moai/config/sections/user.yaml",
        "tone: formal\n",
        "tone: casual\n",
    );
    fx.write(".moai/design/x.md", "# user design notes\n");
    fx.write(".moai/.env.local", "KEY=value\n");
    // A link whose target sits outside the state folder: following it would read content
    // the survey never walks, so its digest must stay empty.
    fx.write("outside.txt", "outside content\n");
    fx.link(".moai/design/link.md", "../outside.txt");
    fx.write(
        ".claude/settings.json",
        "{\"outputStyle\": \"moai/moai-easy\"}\n",
    );
    fx.manifest_at(".moai/manifest.json");
    fx.git_commit_all();
    fx
}

/// AC-1 (REQ-1, REQ-7): the step sorts the frame — `go_same` for the byte-identical files,
/// `go_changed` for the rewritten one, `user_data` for the unmanifested one — counts the one
/// secret-shaped name without opening it, marks tracked files, carries the item fields, and a
/// second plan run reads the same.
#[test]
fn the_go_residue_step_classifies_the_survey_frame() {
    let fx = survey_frame();
    let first = fx.update_plan();
    let found = step(&first, "go-residue");
    for name in [
        "policies/core/a.md",
        "workflows/gtd.md",
        "output-styles/moai/moai-easy.md",
    ] {
        let item = Fx::residue_item(&first, name.split('/').next_back().unwrap());
        assert_eq!(item["class"], json!("go_same"), "{name}");
        assert_eq!(item["tracked"], json!(true), "{name}");
    }
    let changed = Fx::residue_item(&first, "user.yaml");
    assert_eq!(changed["class"], json!("go_changed"), "the rewritten file");
    assert_eq!(changed["tracked"], json!(true));
    let user = Fx::residue_item(&first, "x.md");
    assert_eq!(user["class"], json!("user_data"), "the unmanifested file");
    assert_eq!(user["tracked"], json!(true));
    // The secret-shaped name is counted, never opened: no item carries it or its digest.
    let items = found["preview"]["items"].as_array().unwrap();
    assert_eq!(
        found["preview"]["secret_names"],
        json!(1),
        "one secret-shaped name"
    );
    assert!(
        items
            .iter()
            .all(|i| !i["path"].as_str().is_some_and(|p| p.contains(".env"))),
        "no item names the secret file: {items:?}"
    );
    // A link is named but never read: its own digest stays empty, and the outside file's
    // digest appears nowhere in the step.
    let linked = Fx::residue_item(&first, "link.md");
    assert_eq!(linked["class"], json!("user_data"));
    assert_eq!(linked["sha256"], json!(null), "a link's digest stays empty");
    let outside = json!(sha("outside content\n"));
    assert!(
        items.iter().all(|i| i["sha256"] != outside),
        "the outside content leaked into a digest: {items:?}"
    );
    // Every item carries the fields the SPEC requires.
    for i in items {
        for field in [
            "path",
            "class",
            "sha256",
            "tracked",
            "resurrected",
            "suggested_action",
        ] {
            assert!(i.get(field).is_some(), "item {:?} lacks {field}", i["path"]);
        }
    }
    // A second run reads the same (REQ-7).
    let second = fx.update_plan();
    assert_eq!(
        step(&second, "go-residue")["preview"]["items"],
        step(&first, "go-residue")["preview"]["items"]
    );
}

/// AC-2 (REQ-2): with the live `.moai/manifest.json` gone, the step still proves the files
/// from the backup manifests (first a `.moai-backup*` manifest, then a go-markers one).
#[test]
fn the_go_residue_step_proves_from_the_backup_manifests() {
    let fx = survey_frame();
    // The live manifest moves into a go-markers copy; a `.moai-backup` manifest remains.
    fs::create_dir_all(fx.root.join(".moai/backups/go-markers-20261120T234501Z")).unwrap();
    fs::rename(
        fx.root.join(".moai/manifest.json"),
        fx.root
            .join(".moai/backups/go-markers-20261120T234501Z/manifest.json"),
    )
    .unwrap();
    fx.manifest_at(".moai-backup/manifest.json");
    let plan = fx.update_plan();
    let found = step(&plan, "go-residue");
    assert!(
        found["preview"]["proof"]
            .as_str()
            .is_some_and(|p| p.contains(".moai-backup/manifest.json")),
        "the proof names the backup manifest: {:?}",
        found["preview"]["proof"]
    );
    for name in ["a.md", "gtd.md"] {
        assert_eq!(
            Fx::residue_item(&plan, name)["class"],
            json!("go_same"),
            "{name}"
        );
    }
}

/// REQ-3, REQ-4: files only the user (or this generation) wrote are listed as `user_data`
/// with the action `keep`; they alone never ask for a decision, so the step reads `unchanged`
/// and the plan's `all_unchanged` is not forced false by it.
#[test]
fn user_data_alone_leaves_the_go_residue_step_unchanged() {
    let fx = Fx::new();
    // A backup manifest that names no file: it proves nothing for the surveyed ones.
    fx.manifest_at(".moai-backup/manifest.json");
    fx.write(".moai/design/x.md", "# user design notes\n");
    fx.git_commit_all();
    let plan = fx.update_plan();
    let found = step(&plan, "go-residue");
    let user = Fx::residue_item(&plan, "x.md");
    assert_eq!(user["class"], json!("user_data"), "the unmanifested file");
    assert_eq!(user["tracked"], json!(true));
    assert_eq!(user["suggested_action"], json!("keep"));
    assert_eq!(
        found["state"],
        json!("unchanged"),
        "user data alone is not a decision: {found:?}"
    );
    // Every surveyed item is kept as it is, so nothing here could ask for a decision.
    let items = found["preview"]["items"].as_array().unwrap();
    assert!(
        items.iter().all(|i| i["suggested_action"] == json!("keep")),
        "only keep items in this frame: {items:?}"
    );
    // The step takes no part in the other steps' state: the plan's `all_unchanged` is the
    // AND of every step, so the residue step is the one that must read `unchanged` here.
    let body = plan.get("result").unwrap_or(&plan);
    let others_unchanged = body["phases"]
        .as_array()
        .unwrap()
        .iter()
        .flat_map(|p| p["steps"].as_array().into_iter().flatten())
        .filter(|s| s["step"] != json!("go-residue"))
        .all(|s| matches!(s["state"].as_str(), Some("unchanged" | "skip" | "optional")));
    assert_eq!(
        body["all_unchanged"],
        json!(others_unchanged),
        "all_unchanged follows the other steps alone: {:?}",
        body["all_unchanged"]
    );
}

/// AC-3 (REQ-6): with no manifest anywhere the step reports no proof with zero candidates
/// and the plan still succeeds.
#[test]
fn the_go_residue_step_reports_no_proof_when_no_manifest_exists() {
    let fx = survey_frame();
    fs::remove_file(fx.root.join(".moai/manifest.json")).unwrap();
    let plan = fx.update_plan();
    let found = step(&plan, "go-residue");
    assert_eq!(
        found["preview"]["proof"],
        json!("none"),
        "no proof anywhere"
    );
    assert_eq!(found["preview"]["items"], json!([]), "zero candidates");
    assert_eq!(
        found["preview"]["secret_names"],
        json!(0),
        "nothing was surveyed as a candidate"
    );
}
