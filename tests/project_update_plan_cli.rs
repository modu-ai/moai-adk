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
        let extra = if github_issues {
            json!({"github_issues": true})
        } else {
            json!({})
        };
        self.with_profile_extra(extra);
    }

    /// [`Fx::with_profile`] with `extra` keys merged into the profile.
    fn with_profile_extra(&self, extra: Value) {
        let detect = self.call(&["project-detect", self.p()], b"");
        let mut profile = json!({
            "schema_version": 2, "project_id": "demo", "purpose": "Build a small editor",
            "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude", "codex"],
            "detected": detect["detected"],
        });
        for (k, v) in extra.as_object().unwrap() {
            profile[k] = v.clone();
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

    /// `generation-plan` with `request`, then `generation-apply` of that plan.
    fn generate(&self, request: &Value) {
        let plan = self.call(
            &["generation-plan", self.p()],
            &serde_json::to_vec(request).unwrap(),
        );
        let digest = plan["plan_digest"].as_str().unwrap().to_string();
        self.call(
            &[
                "generation-apply",
                self.p(),
                "--plan-digest",
                &digest,
                "--permission-reference",
                REF,
            ],
            &serde_json::to_vec(&plan).unwrap(),
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

impl Fx {
    /// A project whose update preview reads `all_unchanged`: a profile, then every printed
    /// `ready` command applied.
    fn settled(&self) {
        self.with_profile(false);
        for s in steps(&self.plan())
            .into_iter()
            .filter(|s| s["state"] == "ready")
        {
            assert_eq!(self.apply_step(s)["ok"], json!(true), "{s}");
        }
        let plan = self.plan();
        assert_eq!(plan["all_unchanged"], json!(true), "{plan}");
    }

    /// The generation manifest's path, from the generation step's files.
    fn generation_manifest(&self, plan: &Value) -> PathBuf {
        let rel = generation_paths(plan)
            .into_iter()
            .find(|p| p.ends_with("manifests/generation.json"))
            .unwrap_or_else(|| panic!("no manifest row: {plan}"));
        self.root.join(rel)
    }
}

/// The paths the generation step plans.
fn generation_paths(plan: &Value) -> Vec<String> {
    step(plan, "generation")["preview"]["files"]
        .as_array()
        .into_iter()
        .flatten()
        .map(|f| f["path"].as_str().unwrap().to_string())
        .collect()
}

/// #2953 F1: earlier state that stays where it is reads `state` with a reason, prints no
/// command, and keeps `all_unchanged` false.
fn assert_tied(plan: &Value, state: &str, needle: &str) {
    let go = step(plan, "go-state");
    assert_eq!(go["state"], state, "{go}");
    assert!(
        go["reason"].as_str().unwrap().contains(needle),
        "{needle:?}: {go}"
    );
    assert_eq!(go["apply"], Value::Null, "{go}");
    assert_eq!(phase(plan, "backup")["state"], state, "{plan}");
    assert_eq!(plan["all_unchanged"], json!(false), "{plan}");
}

/// #2953 F1: a go-state migration stopped midway (a Go backup and no `.moai`) waits for the
/// user's revert decision.
#[test]
fn go_state_interrupted() {
    let fx = Fx::new();
    fs::create_dir_all(fx.root.join(".moai-backup/config")).unwrap();
    assert_tied(&fx.plan(), "needs_decision", "state_root `interrupted`");
}

/// #2953 F1: every other kind of earlier state that cannot move is `blocked` with its reason.
#[test]
fn go_state_tied_blocked() {
    fn newer(fx: &Fx) {
        fx.write(
            ".moai/moai-cli.json",
            r#"{"generation":"moai-cli","schema_version":99}"#,
        );
    }
    fn linked(fx: &Fx) {
        let target = fx.home.parent().unwrap().join("elsewhere");
        fs::create_dir_all(&target).unwrap();
        std::os::unix::fs::symlink(&target, fx.root.join(".moai")).unwrap();
    }
    fn mixed(fx: &Fx) {
        fx.write(
            ".moai/moai-cli.json",
            r#"{"generation":"moai-cli","schema_version":1}"#,
        );
        fx.write(".moai/manifest.json", "{}");
    }
    fn unknown(fx: &Fx) {
        fx.write(".moai/notes.txt", "mine\n");
    }
    fn no_manifest(fx: &Fx) {
        fx.write(".moai/config/config.yaml", "user: {}\n");
    }
    fn settings_reference(fx: &Fx) {
        let body = "a: 1\n";
        let h = format!("sha256:{:x}", Sha256::digest(body.as_bytes()));
        fx.write(".moai/config/sections/a.yaml", body);
        fx.write(
            ".moai/manifest.json",
            &json!({"version": "v3.2.1-rc.6", "files": {".moai/config/sections/a.yaml":
                {"provenance": "template_managed", "template_hash": h,
                 "deployed_hash": h, "current_hash": h}}})
            .to_string(),
        );
        fx.write(
            ".claude/settings.json",
            r#"{"statusLine":{"type":"command","command":".moai/status.sh"}}"#,
        );
    }
    type Make = fn(&Fx);
    let cases: [(Make, &str); 6] = [
        (newer, "state_root `refuse`"),
        // A linked `.moai` is refused before any verdict; its reason is the refusal.
        (linked, ""),
        (mixed, "state_root `mixed`"),
        (unknown, "state_root `unknown`"),
        (no_manifest, "no usable Go manifest"),
        (settings_reference, "settings reference .moai/"),
    ];
    for (make, needle) in cases {
        let fx = Fx::new();
        make(&fx);
        assert_tied(&fx.plan(), "blocked", needle);
    }
}

/// #2953 F2: the personal-file choice comes from the manifest — `none`, `tracked` and
/// `ignored` — and a record that cannot tell asks the user.
#[test]
fn local_file_from_manifest() {
    let fx = Fx::new();
    fx.with_profile(false);
    let plan = fx.plan();
    let request = &step(&plan, "generation")["apply"]["stdin"]["request"];
    assert_eq!(request["local_file"], "none", "{plan}");
    let paths = generation_paths(&plan);
    assert!(
        !paths
            .iter()
            .any(|p| p == "AGENTS.local.md" || p == ".gitignore"),
        "{paths:?}"
    );

    for (choice, gitignore) in [("tracked", false), ("ignored", true)] {
        let fx = Fx::new();
        fx.with_profile(false);
        let mut request = step(&fx.plan(), "generation")["apply"]["stdin"]["request"].clone();
        request["local_file"] = json!(choice);
        fx.generate(&request);
        let plan = fx.plan();
        let generation = step(&plan, "generation");
        assert_eq!(generation["state"], "unchanged", "{choice}: {generation}");
        let paths = generation_paths(&plan);
        assert!(
            paths.iter().any(|p| p == "AGENTS.local.md"),
            "{choice}: {paths:?}"
        );
        assert_eq!(
            paths.iter().any(|p| p == ".gitignore"),
            gitignore,
            "{choice}: {paths:?}"
        );
    }

    // `ignored` over the user's own rule records no section: tracked or ignored is unknown.
    let fx = Fx::new();
    fx.with_profile(false);
    fx.write(".gitignore", "AGENTS.local.md\n");
    let mut request = step(&fx.plan(), "generation")["apply"]["stdin"]["request"].clone();
    request["local_file"] = json!("ignored");
    fx.generate(&request);
    let plan = fx.plan();
    let generation = step(&plan, "generation");
    assert_eq!(generation["state"], "needs_decision", "{generation}");
    assert!(
        generation["reason"]
            .as_str()
            .unwrap()
            .contains("tracked or ignored"),
        "{generation}"
    );
    assert_eq!(generation["apply"], Value::Null, "{generation}");
    assert_eq!(plan["all_unchanged"], json!(false), "{plan}");
}

/// #2953 F3: a `ready` profile step holds the generation step until the preview is re-run.
#[test]
fn profile_first() {
    let fx = Fx::new();
    fx.with_profile(false);
    fx.write(
        "Cargo.toml",
        "[package]\nname = \"demo\"\nversion = \"0.1.0\"\nedition = \"2021\"\n",
    );
    let plan = fx.plan();
    let profile = step(&plan, "profile");
    assert_eq!(profile["state"], "ready", "{profile}");
    let generation = step(&plan, "generation");
    assert_eq!(generation["state"], "blocked", "{generation}");
    assert_eq!(
        generation["reason"], "re-run after the profile apply",
        "{generation}"
    );
    assert_eq!(generation["apply"], Value::Null, "{generation}");
    assert_eq!(fx.apply_step(profile)["ok"], json!(true));
    let again = fx.plan();
    assert_eq!(step(&again, "profile")["state"], "unchanged", "{again}");
    assert_ne!(
        step(&again, "generation")["reason"],
        "re-run after the profile apply",
        "{again}"
    );
}

/// #2953 F4 (Q1 A): a manifest owning a project skill the stored state cannot rebuild asks
/// the user, names the file, and plans no removal.
#[test]
fn unrebuildable_not_removed() {
    let fx = Fx::new();
    fx.settled();
    let manifest_path = fx.generation_manifest(&fx.plan());
    let skill = ".claude/skills/prj-checks/SKILL.md";
    let body = "# checks\n";
    fx.write(skill, body);
    let mut manifest: Value = serde_json::from_slice(&fs::read(&manifest_path).unwrap()).unwrap();
    manifest["owned_files"].as_array_mut().unwrap().push(json!({
        "path": skill,
        "section_id": "generation:project-skill:claude:prj-checks",
        "baseline_sha256": format!("sha256:{:x}", Sha256::digest(body.as_bytes())),
    }));
    fs::write(
        &manifest_path,
        serde_json::to_vec_pretty(&manifest).unwrap(),
    )
    .unwrap();
    let plan = fx.plan();
    let generation = step(&plan, "generation");
    assert_eq!(generation["state"], "needs_decision", "{generation}");
    assert_eq!(
        generation["preview"]["not_rebuilt"],
        json!([skill]),
        "{generation}"
    );
    assert_eq!(generation["apply"], Value::Null, "{generation}");
    assert!(
        !generation["preview"]["files"]
            .as_array()
            .into_iter()
            .flatten()
            .any(|f| f["action"] == "remove"),
        "{generation}"
    );
    assert_eq!(plan["all_unchanged"], json!(false), "{plan}");
}

/// #2953 I1 (Q3 B, other direction): when the manifest bytes an apply would publish differ
/// from the file, the manifest row is `merge` and the preview is not `all_unchanged`.
#[test]
fn manifest_bytes_differ() {
    let fx = Fx::new();
    fx.settled();
    let manifest_path = fx.generation_manifest(&fx.plan());
    let manifest: Value = serde_json::from_slice(&fs::read(&manifest_path).unwrap()).unwrap();
    // The same record written compact: the bytes an apply publishes are pretty JSON.
    fs::write(&manifest_path, serde_json::to_vec(&manifest).unwrap()).unwrap();
    let plan = fx.plan();
    let generation = step(&plan, "generation");
    let row = generation["preview"]["files"]
        .as_array()
        .unwrap()
        .iter()
        .find(|f| {
            f["path"]
                .as_str()
                .unwrap()
                .ends_with("manifests/generation.json")
        })
        .unwrap();
    assert_eq!(row["action"], "merge", "{generation}");
    assert_eq!(generation["state"], "ready", "{generation}");
    assert_eq!(plan["all_unchanged"], json!(false), "{plan}");

    // The agents plan decides the same row on its own (`agents-plan`): `unchanged` after its
    // apply, `merge` once the bytes differ.
    let fx = Fx::new();
    fx.with_profile(false);
    let request = step(&fx.plan(), "generation")["apply"]["stdin"]["request"].clone();
    let agents_request = serde_json::to_vec(&json!({
        "protocol_version": 1, "language": request["language"],
        "workflow_reference": request["workflow_reference"], "local_file": request["local_file"],
    }))
    .unwrap();
    fs::create_dir_all(
        fx.root
            .join(request["workflow_reference"].as_str().unwrap()),
    )
    .unwrap();
    let agents_row = |fx: &Fx| {
        let plan = fx.call(&["agents-plan", fx.p()], &agents_request);
        let row = plan["files"]
            .as_array()
            .unwrap()
            .iter()
            .find(|f| {
                f["path"]
                    .as_str()
                    .unwrap()
                    .ends_with("manifests/generation.json")
            })
            .unwrap()
            .clone();
        (plan, row)
    };
    let (plan, _) = agents_row(&fx);
    let digest = plan["plan_digest"].as_str().unwrap().to_string();
    fx.call(
        &[
            "agents-apply",
            fx.p(),
            "--plan-digest",
            &digest,
            "--permission-reference",
            REF,
        ],
        &serde_json::to_vec(&plan).unwrap(),
    );
    assert_eq!(agents_row(&fx).1["action"], "unchanged");
    let path = fx.root.join(".moai/manifests/generation.json");
    let manifest: Value = serde_json::from_slice(&fs::read(&path).unwrap()).unwrap();
    fs::write(&path, serde_json::to_vec(&manifest).unwrap()).unwrap();
    assert_eq!(agents_row(&fx).1["action"], "merge");
}

/// #2953 F5: the request defaults — instruction language from the profile, `local_file`
/// from the manifest, workflow values from the manifest. (The settings stdin without `hooks`
/// is the unit test `the_settings_stdin_has_no_hooks`: this path never plans settings `ready`.)
#[test]
fn request_defaults() {
    let fx = Fx::new();
    fx.with_profile_extra(json!({"language": {"instruction_files": "ko"}}));
    let plan = fx.plan();
    let request = step(&plan, "generation")["apply"]["stdin"]["request"].clone();
    assert_eq!(request["language"], "ko", "{plan}");
    assert_eq!(request["local_file"], "none", "{plan}");
    assert_eq!(request["workflow_id"], "project", "{plan}");
    assert_eq!(
        request["workflow_reference"], ".moai/workflows/project/",
        "{plan}"
    );

    // The workflow values a generation apply recorded are what the next preview plans with.
    let mut custom = request;
    custom["workflow_id"] = json!("review-flow");
    custom["workflow_reference"] = json!(".moai/workflows/review-flow/");
    fx.generate(&custom);
    let again = fx.plan();
    let generation = step(&again, "generation");
    assert_eq!(generation["state"], "unchanged", "{generation}");
    assert!(
        generation_paths(&again)
            .iter()
            .any(|p| p.contains("review-flow")),
        "{generation}"
    );
}
