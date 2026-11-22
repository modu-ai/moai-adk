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
        self.run_exe(EXE, args, stdin)
    }

    /// [`Fx::run`] with another executable (a copy inside a fake install folder).
    fn run_exe(&self, exe: &str, args: &[&str], stdin: &[u8]) -> Value {
        let mut child = Command::new(exe)
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

    /// A fake install folder beside the project and HOME: a copy of this binary at
    /// `install/bin/<name>` with the skill package and the agent sources next to it, the layout
    /// `update::installed_package_dir` and `installed_agents_dir` look for. Returns the copy.
    fn install(&self) -> String {
        let repo = Path::new(env!("CARGO_MANIFEST_DIR"));
        let dir = self.home.parent().unwrap().join("install");
        fs::create_dir_all(dir.join("bin")).unwrap();
        fs::create_dir_all(dir.join("skills")).unwrap();
        let exe = dir.join("bin").join(Path::new(EXE).file_name().unwrap());
        common::copy_executable(Path::new(EXE), &exe);
        for (from, to) in [("skills/moai", "skills/moai"), ("agents", "agents")] {
            copy_tree(&repo.join(from), &dir.join(to));
        }
        exe.to_str().unwrap().to_string()
    }

    /// [`Fx::plan`] run by the copy [`Fx::install`] made.
    fn plan_installed(&self, exe: &str) -> Value {
        let before = self.tree();
        let v = self.run_exe(exe, &["project-update-plan", self.p()], b"");
        assert_eq!(v["ok"], json!(true), "{v}");
        assert_eq!(before, self.tree(), "project-update-plan wrote something");
        v["result"].clone()
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
            "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude", "codex", "omp"],
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

const INSTALL: [&str; 11] = [
    "profile",
    "generation",
    "claude-settings",
    "hooks",
    "github",
    "codex-hooks",
    "codex-config",
    "codex-skill",
    "omp-hooks",
    "omp-config",
    "omp-agents",
];

fn copy_tree(from: &Path, to: &Path) {
    fs::create_dir_all(to).unwrap();
    for e in fs::read_dir(from).unwrap() {
        let e = e.unwrap();
        let target = to.join(e.file_name());
        if e.file_type().unwrap().is_dir() {
            copy_tree(&e.path(), &target);
        } else {
            fs::copy(e.path(), target).unwrap();
        }
    }
}

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
    // #2936 R10: the merge phase is a real phase; with no AGENTS.local.md it has nothing to do.
    let merge = phase(&plan, "merge");
    assert_eq!(merge["state"], "skip", "{merge}");
    assert_eq!(
        step(&plan, "agents-local-merge")["state"],
        "skip",
        "{merge}"
    );
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
    // #2936 R8: a project without a status line gets MoAI's, so the owned keys include one.
    let desired = br#"{"outputStyle":"Explanatory","statusLine":{"type":"command","command":"moai statusline"}}"#;
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

/// #2936 R8 (F3): a settings file whose keys are all the user's is left alone, and the step says
/// that MoAI owns nothing there instead of calling the owned keys "current".
#[test]
fn a_settings_file_with_only_the_users_keys_says_moai_owns_none() {
    let fx = Fx::new();
    fx.with_profile(false);
    fs::create_dir_all(fx.root.join(".claude")).unwrap();
    let own = br#"{"statusLine":{"type":"command","command":"my-status"}}"#;
    fs::write(fx.root.join(".claude/settings.json"), own).unwrap();
    let plan = fx.plan();
    let settings = step(&plan, "claude-settings");
    assert_eq!(settings["state"], "unchanged", "{settings}");
    let reason = settings["reason"].as_str().unwrap();
    assert!(reason.contains("owns no key"), "{reason}");
    assert!(!reason.contains("are current"), "{reason}");
    assert_eq!(
        fs::read(fx.root.join(".claude/settings.json")).unwrap(),
        own
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

/// AC-20 (#2936 S1c-1): the backup phase's own text no longer says `.codex` is not covered, and
/// its go-state step shows how many `.codex` items fall in each class. Nothing moves yet.
#[test]
fn backup_step_counts_codex_classes_and_does_not_call_codex_uncovered() {
    let fx = Fx::new();
    fx.go_project();
    let manifest: Value =
        serde_json::from_slice(&fs::read(fx.root.join(".moai/manifest.json")).unwrap()).unwrap();
    let mut files = manifest["files"].as_object().unwrap().clone();
    let agent = "name = \"expert\"\n";
    let hooks = "{\"hooks\":{}}\n";
    fx.write(".codex/agents/moai/expert.toml", agent);
    fx.write(".codex/hooks.json", hooks);
    let h = format!("{:x}", Sha256::digest(agent.as_bytes()));
    files.insert(
        ".codex/agents/moai/expert.toml".into(),
        json!({"provenance": "template_managed", "template_hash": format!("sha256:{h}"),
               "deployed_hash": format!("sha256:{h}"), "current_hash": format!("sha256:{h}")}),
    );
    // The wiring records a bare hex hash.
    let bare = format!("{:x}", Sha256::digest(hooks.as_bytes()));
    files.insert(
        ".codex/hooks.json".into(),
        json!({"provenance": "generated_managed", "template_hash": "",
               "deployed_hash": bare, "current_hash": bare,
               "parts": [{"kind": "whole-file", "origin": "created"}]}),
    );
    fx.write(
        ".moai/manifest.json",
        &json!({"version": "v3.2.1-rc.6", "deployed_at": "2026-11-01T00:00:00Z", "files": files})
            .to_string(),
    );
    let plan = fx.plan();
    let go = step(&plan, "go-state");
    // `.claude/workflows/w.md` of the fixture is the first tool asset, the agent file the second.
    assert_eq!(go["preview"]["classes"]["tool_asset"], json!(2), "{go}");
    assert_eq!(go["preview"]["classes"]["go_hooks"], json!(1), "{go}");
    assert!(
        go["preview"]["warnings"]
            .to_string()
            .contains("running Codex session"),
        "{go}"
    );
    let reason = phase(&plan, "backup")["reason"].as_str().unwrap();
    assert!(reason.contains(".codex"), "{reason}");
    assert!(!reason.contains("not covered"), "{reason}");
    assert!(reason.contains(".agents"), "{reason}");
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

/// #2936 L1: a Go project whose settings the manifest does not record. The settings files
/// name `.moai/` in the keys the tests give.
fn go_project_with_unrecorded_settings(settings: &str, local: &str) -> Fx {
    let fx = Fx::new();
    let body = "a: 1\n";
    let h = format!("sha256:{:x}", Sha256::digest(body.as_bytes()));
    fx.write(".moai/config/sections/a.yaml", body);
    // An earlier `.moai-cli/project.yaml` that is no Rust profile: the install steps are then
    // planned from the files (as in the project the reproduction in #2936 follows) and not
    // all waiting for `/moai project`.
    fx.write(".moai-cli/project.yaml", "project_id: demo\n");
    fx.write(
        ".moai/manifest.json",
        &json!({"version": "v3.2.1-rc.6", "files": {".moai/config/sections/a.yaml": {
            "provenance": "template_managed", "template_hash": h,
            "deployed_hash": h, "current_hash": h}}})
        .to_string(),
    );
    fx.write(".claude/settings.json", settings);
    fx.write(".claude/settings.local.json", local);
    fx
}

/// #2936 L1: the backup step's `settings reference .moai/` hold names the file and the keys
/// (never their values) and says what to do next; the settings step does not call such a file
/// "current" and lists the same keys in its preview.
#[test]
fn the_settings_hold_names_the_keys_and_the_settings_step_does_not_say_current() {
    let fx = go_project_with_unrecorded_settings(
        // `.moai/plans/` is one character off the value the Go tool wrote, so it is the user's
        // (#2936 L2a): the go-settings step has nothing to clean and this note stands.
        r#"{"model":"opus","statusLine":{"type":"command","command":".moai/secret-name.sh"},"plansDirectory":".moai/plans/"}"#,
        r#"{"env":{"A":".moai/x"},"other":1}"#,
    );
    let plan = fx.plan();
    assert_eq!(step(&plan, "go-settings")["state"], "unchanged", "{plan}");
    assert_tied(
        &plan,
        "blocked",
        "held by .claude/settings.json (plansDirectory, statusLine); \
         .claude/settings.local.json (env); this update does not change those keys, so change \
         or remove them yourself and run the update again; this update does not edit \
         .claude/settings.local.json; edit or remove these keys yourself (env), then run \
         `moai project-update-plan` again",
    );
    assert!(
        !plan.to_string().contains("secret-name"),
        "a value is never printed: {plan}"
    );
    let settings = step(&plan, "claude-settings");
    assert_eq!(settings["state"], "unchanged", "{settings}");
    let reason = settings["reason"].as_str().unwrap();
    assert!(
        reason.contains("still naming .moai/ and holding the Go .moai/ folder (go-state step)")
            && reason.contains(".claude/settings.json (plansDirectory, statusLine)"),
        "{settings}"
    );
    assert_eq!(
        settings["preview"]["go_references"],
        json!([
            {"file": ".claude/settings.json", "keys": ["plansDirectory", "statusLine"]},
            {"file": ".claude/settings.local.json", "keys": ["env"]}
        ]),
        "{settings}"
    );
}

/// #2936 L1 (audit mutant M8): the "held by" note belongs to the `settings reference .moai/`
/// block only. A Go project whose `.moai/` stays for another reason (here: no usable Go
/// manifest) keeps that reason as it is, although a settings key names `.moai/`.
#[test]
fn the_held_by_note_is_only_for_the_settings_reference_block() {
    let fx = Fx::new();
    fx.write(".moai/config/config.yaml", "user: {}\n");
    fx.write(
        ".claude/settings.json",
        r#"{"statusLine":{"type":"command","command":".moai/status.sh"}}"#,
    );
    let plan = fx.plan();
    assert_tied(&plan, "blocked", "no usable Go manifest");
    let reason = step(&plan, "go-state")["reason"].as_str().unwrap();
    assert!(!reason.contains("held by"), "{reason}");
    assert!(!reason.contains("statusLine"), "{reason}");
}

/// #2936 L1 (audit mutant M9): the "still naming" note replaces the word "current" only on a
/// step that reads `unchanged`. A settings step that has MoAI's status line to write (`ready`)
/// in a Go project whose settings name `.moai/` carries neither the note nor `go_references`.
#[test]
fn a_settings_step_that_is_not_unchanged_gets_no_go_reference_note() {
    // The file has no status line, so the step would create MoAI's own; `plansDirectory` is
    // one character off the Go tool's value, so the go-settings step leaves it alone.
    let fx = go_project_with_unrecorded_settings(
        r#"{"model":"opus","plansDirectory":".moai/plans/"}"#,
        r#"{"other":1}"#,
    );
    let plan = fx.plan();
    assert_tied(
        &plan,
        "blocked",
        "held by .claude/settings.json (plansDirectory)",
    );
    let settings = step(&plan, "claude-settings");
    assert_eq!(settings["state"], "ready", "{settings}");
    assert!(
        !settings["reason"]
            .as_str()
            .unwrap()
            .contains("still naming"),
        "{settings}"
    );
    assert!(
        settings["preview"].get("go_references").is_none(),
        "{settings}"
    );
}

/// #2936 L1 (audit mutant M11): a Go project whose settings files name `.moai/` in no key value
/// gets no note anywhere. The text of the file still holds the folder (a key name does), so
/// go-state keeps its own reason, without "held by"; the settings step stays plainly unchanged.
#[test]
fn zero_references_give_no_note() {
    let fx = go_project_with_unrecorded_settings(
        r#"{"statusLine":{"type":"command","command":"moai statusline"},".moai/x":1}"#,
        r#"{"other":1}"#,
    );
    let plan = fx.plan();
    assert_tied(&plan, "blocked", "settings reference .moai/");
    let reason = step(&plan, "go-state")["reason"].as_str().unwrap();
    assert!(!reason.contains("held by"), "{reason}");
    let settings = step(&plan, "claude-settings");
    assert_eq!(settings["state"], "unchanged", "{settings}");
    assert!(
        !settings["reason"]
            .as_str()
            .unwrap()
            .contains("still naming"),
        "{settings}"
    );
    assert!(
        settings["preview"].get("go_references").is_none(),
        "{settings}"
    );
}

/// #2936 L1: the note is for the earlier generation's folder only. A settings file with the
/// same keys in a project that is not Go-generation keeps the plain "current" reason.
#[test]
fn a_project_without_go_state_gets_no_go_reference_note() {
    let fx = Fx::new();
    fx.with_profile(false);
    // MoAI's own status line is already there, so the step has nothing to write.
    fx.write(
        ".claude/settings.json",
        r#"{"statusLine":{"type":"command","command":"moai statusline"},"plansDirectory":".moai/plans"}"#,
    );
    let plan = fx.plan();
    let settings = step(&plan, "claude-settings");
    assert_eq!(settings["state"], "unchanged", "{settings}");
    assert!(
        !settings["reason"]
            .as_str()
            .unwrap()
            .contains("still naming"),
        "{settings}"
    );
    assert!(
        settings["preview"].get("go_references").is_none(),
        "{settings}"
    );
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
    // #3009 S3b: the managed `.gitignore` block (the state-folder allow-list) is planned for
    // every choice; the personal file is planned only when the choice makes it.
    assert!(!paths.iter().any(|p| p == "AGENTS.local.md"), "{paths:?}");
    assert!(paths.iter().any(|p| p == ".gitignore"), "{paths:?}");

    for (choice, lists_local) in [("tracked", false), ("ignored", true)] {
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
        assert!(
            paths.iter().any(|p| p == ".gitignore"),
            "{choice}: {paths:?}"
        );
        // The block lists the personal file only for the `ignored` choice (condition 4).
        let ignore = fs::read_to_string(fx.root.join(".gitignore")).unwrap();
        assert_eq!(
            ignore.lines().any(|l| l == "AGENTS.local.md"),
            lists_local,
            "{choice}: {ignore}"
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

/// #3009 S3b: the block recorded before the `kind` field existed listed the personal file, so
/// a manifest row without a kind still reads as `ignored`; a row whose kind says "state only"
/// reads as `tracked` when `AGENTS.local.md` exists, and as `none` when it does not.
#[test]
fn the_recorded_block_kind_tells_the_personal_file_choice() {
    let old_block = "# moai:managed:begin id=moai-local\nAGENTS.local.md\n.moai-cli/receipts/\n# moai:managed:end id=moai-local\n";
    let sha = |b: &str| format!("sha256:{:x}", Sha256::digest(b.as_bytes()));
    // (kind to record or None to drop it, expected choice in the printed request)
    for (kind, expected) in [(None, "ignored"), (Some("agents:ignore-state"), "tracked")] {
        let fx = Fx::new();
        fx.with_profile(false);
        let mut request = step(&fx.plan(), "generation")["apply"]["stdin"]["request"].clone();
        request["local_file"] = json!("ignored");
        fx.generate(&request);
        let manifest_path = fx.generation_manifest(&fx.plan());
        let mut manifest: Value =
            serde_json::from_slice(&fs::read(&manifest_path).unwrap()).unwrap();
        let text = if kind.is_none() {
            old_block.to_string()
        } else {
            old_block.replace("AGENTS.local.md\n", "")
        };
        for file in manifest["owned_files"].as_array_mut().unwrap() {
            if file["path"] == ".gitignore" {
                file["baseline_sha256"] = json!(sha(&text));
                match kind {
                    Some(k) => file["kind"] = json!(k),
                    None => {
                        file.as_object_mut().unwrap().remove("kind");
                    }
                }
            }
        }
        fs::write(
            &manifest_path,
            serde_json::to_vec_pretty(&manifest).unwrap(),
        )
        .unwrap();
        fx.write(".gitignore", &text);
        let plan = fx.plan();
        let generation = step(&plan, "generation");
        assert_ne!(
            generation["state"], "needs_decision",
            "{kind:?}: {generation}"
        );
        assert_eq!(
            generation["apply"]["stdin"]["request"]["local_file"], expected,
            "{kind:?}: {generation}"
        );
    }
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

/// #2936 R10: a settled project with a personal `AGENTS.local.md` gets the merge as a decision
/// that prints the apply command; applying it settles the preview again.
#[test]
fn the_merge_step_offers_the_personal_file_and_settles_after_the_apply() {
    let fx = Fx::new();
    fx.settled();
    let settled = fx.plan();
    assert_eq!(
        step(&settled, "agents-local-merge")["state"],
        "skip",
        "{settled}"
    );

    let mine = "# My notes\n\nReply in Korean.\n";
    fs::write(fx.root.join("AGENTS.local.md"), mine).unwrap();
    let plan = fx.plan();
    let merge = step(&plan, "agents-local-merge");
    assert_eq!(merge["state"], "needs_decision", "{merge}");
    assert_eq!(merge["preview"]["body"], json!(mine), "{merge}");
    assert!(
        merge["preview"]["commit_notice"]
            .as_str()
            .unwrap()
            .contains("committed"),
        "{merge}"
    );
    let argv: Vec<&str> = merge["apply"]["argv"]
        .as_array()
        .unwrap()
        .iter()
        .map(|w| w.as_str().unwrap())
        .collect();
    assert_eq!(&argv[..2], ["moai", "agents-merge-apply"], "{merge}");
    assert_eq!(argv.last(), Some(&"<permission-reference>"), "{merge}");
    assert_eq!(plan["all_unchanged"], json!(false));

    assert_eq!(fx.apply_step(merge)["ok"], json!(true), "{merge}");
    let after = fx.plan();
    assert_eq!(
        step(&after, "agents-local-merge")["state"],
        "unchanged",
        "{after}"
    );
    assert_eq!(after["all_unchanged"], json!(true), "{after}");
}

/// #2936 R10: while the generation step still has files to write, the merge waits: that step's
/// digest binds the `AGENTS.md` bytes the merge would change.
#[test]
fn the_merge_waits_for_the_install_phase() {
    let fx = Fx::new();
    fx.with_profile(false);
    fs::write(fx.root.join("AGENTS.local.md"), "# mine\n\ntext\n").unwrap();
    let plan = fx.plan();
    assert_eq!(step(&plan, "generation")["state"], "ready", "{plan}");
    let merge = step(&plan, "agents-local-merge");
    assert_eq!(merge["state"], "blocked", "{merge}");
    assert_eq!(merge["reason"], "re-run after the install phase", "{merge}");
    assert_eq!(merge["apply"], Value::Null, "{merge}");
}

// ---- #2936 R8: the new settings.json and the earlier values offered back ----

impl Fx {
    /// A Go project whose recorded `.claude/settings.json` (`settings`) `go-state-apply` moved
    /// into its backup (the shared settings move only with the Go folder, so `settings` must not
    /// name `.moai/`). `.claude/settings.local.json` is written too and stays in place. Returns
    /// the backup's UTC stamp.
    fn go_project_with_moved_settings(&self, settings: &Value) -> String {
        self.go_project_moved(settings, "{\"model\":\"local\"}\n", false)
    }

    /// [`Fx::go_project_with_moved_settings`] with the body of `settings.local.json` and, when
    /// `edited`, settings whose bytes differ from the hash the manifest recorded at the deploy.
    fn go_project_moved(&self, settings: &Value, local: &str, edited: bool) -> String {
        let mut files = serde_json::Map::new();
        for (rel, body) in [
            (".moai/config/sections/a.yaml", "a: 1\n".to_string()),
            (".claude/settings.json", settings.to_string()),
        ] {
            self.write(rel, &body);
            let recorded = if edited && rel == ".claude/settings.json" {
                format!("{body} ")
            } else {
                body.clone()
            };
            let h = format!("sha256:{:x}", Sha256::digest(recorded.as_bytes()));
            files.insert(
                rel.into(),
                json!({"provenance": "template_managed", "template_hash": h,
                       "deployed_hash": h, "current_hash": h}),
            );
        }
        self.write(".claude/settings.local.json", local);
        self.write(
            ".moai/manifest.json",
            &json!({"version": "v3.2.1-rc.6", "deployed_at": "2026-11-01T00:00:00Z",
                    "files": files})
            .to_string(),
        );
        let plan = self.call(&["go-state-plan", self.p()], b"");
        let digest = plan["plan_digest"].as_str().unwrap().to_string();
        let applied = self.call(
            &[
                "go-state-apply",
                self.p(),
                "--plan-digest",
                &digest,
                "--permission-reference",
                REF,
            ],
            b"",
        );
        assert!(
            !self.root.join(".claude/settings.json").exists(),
            "{applied}"
        );
        applied["backup"].as_str().unwrap().to_string()
    }

    /// Applies every `ready` step the preview prints, as the UPDATE flow does after the user's
    /// go-ahead, so what is left is the user's own decisions.
    fn apply_ready(&self) {
        for s in steps(&self.plan())
            .into_iter()
            .filter(|s| s["state"] == "ready")
        {
            assert_eq!(self.apply_step(s)["ok"], json!(true), "{s}");
        }
    }

    fn settings_json(&self) -> Value {
        serde_json::from_slice(&fs::read(self.root.join(".claude/settings.json")).unwrap()).unwrap()
    }
}

fn credential() -> String {
    format!("{}{}", "ghp_", "a".repeat(30))
}

fn user_settings() -> Value {
    json!({
        "permissions": {"allow": ["Bash(git status)"], "deny": ["Read(./.env)"],
                        "additionalDirectories": ["/tmp/elsewhere"]},
        "env": {"EDITOR": "vim", "MY_TOKEN": "plain", "OTHER": credential()},
        "model": "opus",
        "outputStyle": "Explanatory",
        "statusLine": {"type": "command", "command": "my-status"},
        "hooks": {"PreToolUse": [{"hooks": [{"type": "command", "command": "my-check"}]}]},
        "enabledPlugins": {"x@y": false},
    })
}

fn candidate<'a>(merge: &'a Value, key: &str) -> &'a Value {
    merge["preview"]["candidates"]
        .as_array()
        .unwrap()
        .iter()
        .find(|c| c["key"] == key)
        .unwrap_or_else(|| panic!("no candidate {key}: {merge}"))
}

/// R8-1, R8-3: a project with no status line gets MoAI's and nothing else; a second preview
/// reads `unchanged`.
#[test]
fn a_project_without_a_status_line_gets_the_moai_one_and_nothing_else() {
    let fx = Fx::new();
    fx.with_profile(false);
    let plan = fx.plan();
    let settings = step(&plan, "claude-settings");
    assert_eq!(settings["state"], "ready", "{settings}");
    assert_eq!(
        settings["apply"]["stdin"],
        json!({"statusLine": {"type": "command", "command": "moai statusline"}}),
        "{settings}"
    );
    fx.apply_ready();
    assert_eq!(
        fx.settings_json(),
        json!({"statusLine": {"type": "command", "command": "moai statusline"}})
    );
    let after = fx.plan();
    assert_eq!(
        step(&after, "claude-settings")["state"],
        "unchanged",
        "{after}"
    );
    assert_eq!(step(&after, "claude-settings")["apply"], Value::Null);
}

/// R8-2: a status line the user has stays; the step has nothing to do and writes nothing.
#[test]
fn a_status_line_the_user_has_is_left_alone() {
    let fx = Fx::new();
    fx.with_profile(false);
    let mine = "{\"statusLine\":{\"type\":\"command\",\"command\":\"my-status\"}}\n";
    fx.write(".claude/settings.json", mine);
    let plan = fx.plan();
    let settings = step(&plan, "claude-settings");
    assert_eq!(settings["state"], "unchanged", "{settings}");
    assert_eq!(settings["apply"], Value::Null);
    assert_eq!(
        fs::read_to_string(fx.root.join(".claude/settings.json")).unwrap(),
        mine
    );
}

/// R8-4: with nothing moved into a backup there is nothing to merge back.
#[test]
fn without_a_moved_settings_file_the_settings_merge_skips() {
    let fx = Fx::new();
    fx.settled();
    let plan = fx.plan();
    let merge = step(&plan, "claude-settings-merge");
    assert_eq!(merge["state"], "skip", "{merge}");
    assert_eq!(merge["apply"], Value::Null);
}

/// R8-5 to R8-8, R8-10: the earlier values come back on the user's go-ahead and nothing else.
#[test]
fn the_earlier_values_are_offered_and_merged_back_without_the_rest() {
    let fx = Fx::new();
    let backup = fx.go_project_with_moved_settings(&user_settings());
    fx.with_profile(false);
    let local_before = fs::read(fx.root.join(".claude/settings.local.json")).unwrap();

    // The install step comes first: the merge waits for it.
    let plan = fx.plan();
    let waiting = step(&plan, "claude-settings-merge");
    assert_eq!(waiting["state"], "blocked", "{waiting}");
    assert_eq!(
        waiting["reason"], "re-run after the install phase",
        "{waiting}"
    );
    assert_eq!(waiting["apply"], Value::Null);

    fx.apply_ready();
    let plan = fx.plan();
    let merge = step(&plan, "claude-settings-merge");
    assert_eq!(merge["state"], "needs_decision", "{merge}");
    assert_eq!(
        merge["preview"]["backup"],
        json!(format!(
            ".moai/backups/{backup}/go-state/.claude/settings.json"
        )),
        "{merge}"
    );
    for key in ["permissions", "env", "model", "outputStyle", "statusLine"] {
        assert_eq!(
            candidate(merge, key)["state"],
            "mergeable",
            "{key}: {merge}"
        );
    }
    for key in ["hooks", "enabledPlugins"] {
        assert_eq!(
            candidate(merge, key)["state"],
            "kept_in_backup",
            "{key}: {merge}"
        );
    }
    assert!(
        merge["preview"]["commit_notice"]
            .as_str()
            .unwrap()
            .contains("committed"),
        "{merge}"
    );
    // The credential-shaped value and the credential-named entry are named, never shown.
    let text = plan.to_string();
    assert!(
        !text.contains(&credential()),
        "the plan prints a credential"
    );
    assert!(text.contains("MY_TOKEN") && text.contains("OTHER"));
    assert_eq!(
        merge["apply"]["stdin"]["env"],
        json!({"EDITOR": "vim"}),
        "{merge}"
    );
    assert_eq!(
        merge["apply"]["stdin"]["permissions"],
        json!({"allow": ["Bash(git status)"], "deny": ["Read(./.env)"]}),
        "{merge}"
    );
    assert_eq!(plan["all_unchanged"], json!(false));

    assert_eq!(fx.apply_step(merge)["ok"], json!(true), "{merge}");
    let written = fx.settings_json();
    assert_eq!(written["env"], json!({"EDITOR": "vim"}));
    assert_eq!(written["model"], json!("opus"));
    assert_eq!(written["outputStyle"], json!("Explanatory"));
    assert_eq!(
        written["statusLine"],
        json!({"type": "command", "command": "my-status"})
    );
    assert_eq!(
        written["permissions"],
        json!({"allow": ["Bash(git status)"], "deny": ["Read(./.env)"]})
    );
    for key in ["hooks", "enabledPlugins"] {
        assert!(written.get(key).is_none(), "{key} was merged: {written}");
    }
    assert!(!written.to_string().contains(&credential()));
    assert_eq!(
        fs::read(fx.root.join(".claude/settings.local.json")).unwrap(),
        local_before,
        "the local settings are never touched"
    );
    // The values are MoAI-owned now: a later preview asks nothing and offers nothing to write.
    let after = fx.plan();
    assert_eq!(
        step(&after, "claude-settings")["state"],
        "unchanged",
        "{after}"
    );
    assert_eq!(
        step(&after, "claude-settings-merge")["state"],
        "unchanged",
        "{after}"
    );
}

/// Audit 6888156222 B1: a credential-shaped value outside `env` blocks its key. The key is named,
/// the value is neither printed in the plan nor written to the committed file, and a clean
/// sibling key still merges. Each key is tried alone with the credential and once clean.
#[test]
fn a_credential_outside_env_blocks_its_key_and_is_never_shown_or_written() {
    let secret = credential();
    let cases: [(&str, Value, Value); 4] = [
        (
            "permissions",
            json!({"allow": ["Bash(git status)", format!("Bash(curl -H 'Authorization: Bearer {secret}':*)")]}),
            json!({"allow": ["Bash(git status)", "Bash(curl -H 'Accept: json':*)"]}),
        ),
        (
            "statusLine",
            json!({"type": "command", "command": format!("my-status --token {secret}")}),
            json!({"type": "command", "command": "my-status --compact"}),
        ),
        ("model", json!(format!("opus-{secret}")), json!("opus")),
        (
            "outputStyle",
            json!(format!("Style-{secret}")),
            json!("Explanatory"),
        ),
    ];
    for (key, bad, clean) in cases {
        // The credential: the key is blocked, the clean sibling (`env`) still merges.
        let fx = Fx::new();
        fx.go_project_with_moved_settings(&json!({ key: bad, "env": {"EDITOR": "vim"} }));
        fx.with_profile(false);
        fx.apply_ready();
        let plan = fx.plan();
        let merge = step(&plan, "claude-settings-merge");
        assert_eq!(merge["state"], "needs_decision", "{key}: {merge}");
        assert_eq!(candidate(merge, key)["state"], "blocked", "{key}: {merge}");
        assert_eq!(candidate(merge, "env")["state"], "mergeable", "{key}");
        // The printed command carries no value of the blocked key (for `statusLine` it still
        // carries MoAI's own, which the install step wrote).
        let stdin = &merge["apply"]["stdin"];
        assert!(!stdin.to_string().contains(&secret), "{key}: {merge}");
        if key != "statusLine" {
            assert!(stdin.get(key).is_none(), "{key}: {merge}");
        }
        assert!(
            !plan.to_string().contains(&secret),
            "{key}: the plan prints it"
        );
        assert_eq!(fx.apply_step(merge)["ok"], json!(true), "{key}");
        let written = fx.settings_json();
        assert!(!written.to_string().contains(&secret), "{key}: written");
        assert_eq!(written["env"], json!({"EDITOR": "vim"}), "{key}");

        // The same key without a credential merges.
        let fx = Fx::new();
        fx.go_project_with_moved_settings(&json!({ key: clean.clone() }));
        fx.with_profile(false);
        fx.apply_ready();
        let plan = fx.plan();
        let merge = step(&plan, "claude-settings-merge");
        assert_eq!(
            candidate(merge, key)["state"],
            "mergeable",
            "{key}: {merge}"
        );
        assert_eq!(merge["apply"]["stdin"][key], clean, "{key}: {merge}");
    }
}

/// R8-5: what the Go generation set up itself is not the user's, so nothing is offered.
#[test]
fn only_go_made_values_leave_nothing_to_merge() {
    let fx = Fx::new();
    fx.go_project_with_moved_settings(&json!({
        "statusLine": {"type": "command", "command": "bash $CLAUDE_PROJECT_DIR/status_line.sh"},
        "outputStyle": "MoAI-Easy",
        "hooks": {"Stop": [{"hooks": [{"command": "bash",
            "args": ["${CLAUDE_PROJECT_DIR}/.claude/hooks/moai/handle-stop.sh"]}]}]},
    }));
    fx.with_profile(false);
    fx.apply_ready();
    let plan = fx.plan();
    let merge = step(&plan, "claude-settings-merge");
    assert_eq!(merge["state"], "skip", "{merge}");
    assert_eq!(merge["apply"], Value::Null);
    for key in ["statusLine", "outputStyle", "hooks"] {
        assert_eq!(
            candidate(merge, key)["state"],
            "go_generated",
            "{key}: {merge}"
        );
    }
}

/// R8-4: a backup copy changed after the move is not trusted.
#[test]
fn a_backup_copy_changed_after_the_move_blocks_the_merge() {
    let fx = Fx::new();
    let backup = fx.go_project_with_moved_settings(&user_settings());
    fx.with_profile(false);
    fx.apply_ready();
    fx.write(
        &format!(".moai/backups/{backup}/go-state/.claude/settings.json"),
        "{\"model\":\"tampered\"}\n",
    );
    let plan = fx.plan();
    let merge = step(&plan, "claude-settings-merge");
    assert_eq!(merge["state"], "blocked", "{merge}");
    assert!(
        merge["reason"].as_str().unwrap().contains("changed since"),
        "{merge}"
    );
    assert_eq!(merge["apply"], Value::Null);
}

// ---- #2936 R9 S5a: the host hook steps ----

/// The whole plan as text, for the words no step may print (Q-H2: no trust bypass, no
/// approval mode, no `features.` key, no `trust_level`).
fn assert_no_forbidden_words(plan: &Value) {
    let text = plan.to_string();
    for word in [
        "trust_level",
        "--dangerously-bypass-hook-trust",
        "approvalMode",
        "features.",
    ] {
        assert!(!text.contains(word), "the plan prints {word}: {text}");
    }
}

#[test]
fn the_host_hook_steps_print_their_apply_and_settle() {
    let fx = Fx::new();
    fx.with_profile(false);
    let plan = fx.plan();
    for (name, host, file) in [
        ("codex-hooks", "codex", ".codex/hooks.json"),
        ("omp-hooks", "omp", ".omp/extensions/"),
    ] {
        let s = step(&plan, name);
        assert_eq!(s["state"], "ready", "{name}: {s}");
        let argv: Vec<&str> = s["apply"]["argv"]
            .as_array()
            .unwrap()
            .iter()
            .map(|w| w.as_str().unwrap())
            .collect();
        assert_eq!(&argv[..2], ["moai", "hooks-apply"], "{argv:?}");
        assert_eq!(argv[2], fx.p(), "{argv:?}");
        let at = argv.iter().position(|w| *w == "--host").unwrap();
        assert_eq!(argv[at + 1], host, "{argv:?}");
        assert!(
            s["preview"]["path"].as_str().unwrap().starts_with(file),
            "{name}: {s}"
        );
        assert!(s["plan_digest"].is_string(), "{name}: {s}");
    }
    for name in ["codex-hooks", "omp-hooks"] {
        let out = fx.apply_step(step(&plan, name));
        assert_eq!(out["ok"], json!(true), "{name}: {out}");
    }
    assert!(fx.root.join(".codex/hooks.json").is_file());
    let again = fx.plan();
    for name in ["codex-hooks", "omp-hooks"] {
        let s = step(&again, name);
        assert_eq!(s["state"], "unchanged", "{name}: {s}");
        assert_eq!(s["apply"], Value::Null, "{name}: {s}");
    }
}

#[test]
fn the_codex_hook_step_says_the_user_trusts_the_hooks_in_codex() {
    let fx = Fx::new();
    fx.with_profile(false);
    let plan = fx.plan();
    let notice = step(&plan, "codex-hooks")["preview"]["trust_required"]
        .as_str()
        .unwrap_or_else(|| panic!("no trust notice: {plan}"))
        .to_string();
    assert!(notice.contains("/hooks"), "{notice}");
    assert_no_forbidden_words(&plan);
}

#[test]
fn an_edited_codex_hook_file_is_a_conflict_and_stays_byte_identical() {
    let fx = Fx::new();
    fx.with_profile(false);
    let mine = "{\"hooks\": {\"Notification\": []}}\n";
    fx.write(".codex/hooks.json", mine);
    let before = fx.tree();
    let plan = fx.plan();
    let s = step(&plan, "codex-hooks");
    assert_eq!(s["state"], "conflict", "{s}");
    assert_eq!(s["apply"], Value::Null, "{s}");
    assert!(
        s["reason"]
            .as_str()
            .unwrap()
            .contains("preservation review"),
        "{s}"
    );
    assert_eq!(before, fx.tree(), "the plan changed the user's file");
    assert_eq!(
        fs::read_to_string(fx.root.join(".codex/hooks.json")).unwrap(),
        mine
    );
}

const CODEX_STEPS: [&str; 3] = ["codex-hooks", "codex-config", "codex-skill"];
const OMP_STEPS: [&str; 3] = ["omp-hooks", "omp-config", "omp-agents"];

/// SPEC-20261118-profile-host-omp AC-3: a host's steps are planned only when the profile lists
/// that host; an unlisted host's steps are `skip` with a reason naming it and no command.
#[test]
fn host_steps_follow_the_profile_hosts() {
    for (hosts, codex_picked, omp_picked) in [
        (json!(["claude", "codex", "omp"]), true, true),
        (json!(["claude", "codex"]), true, false),
        (json!(["claude", "omp"]), false, true),
        (json!(["claude"]), false, false),
    ] {
        let fx = Fx::new();
        fx.with_profile_extra(json!({"hosts": hosts}));
        let plan = fx.plan();
        for (names, host, picked) in [
            (CODEX_STEPS, "codex", codex_picked),
            (OMP_STEPS, "omp", omp_picked),
        ] {
            for name in names {
                let s = step(&plan, name);
                if !picked {
                    assert_eq!(s["state"], "skip", "{hosts}: {name}: {s}");
                    assert_eq!(s["apply"], Value::Null, "{hosts}: {name}: {s}");
                    assert!(
                        s["reason"].as_str().unwrap().contains(host),
                        "{hosts}: {name}: {s}"
                    );
                }
            }
        }
        for (name, picked) in [("codex-hooks", codex_picked), ("omp-hooks", omp_picked)] {
            let s = step(&plan, name);
            assert_eq!(s["state"] == "ready", picked, "{hosts}: {name}: {s}");
            assert_eq!(
                s["apply"]["argv"].is_array(),
                picked,
                "{hosts}: {name}: {s}"
            );
        }
    }
}

/// SPEC-20261118-profile-host-omp AC-3: `project-plan` takes a profile with all three hosts and
/// still refuses a fourth entry or an unknown name, writing nothing.
#[test]
fn a_profile_with_omp_passes_project_plan() {
    let fx = Fx::new();
    let detect = fx.call(&["project-detect", fx.p()], b"");
    let profile = |hosts: Value| {
        json!({
            "schema_version": 2, "project_id": "demo", "purpose": "Build a small editor",
            "tools": ["Rust"], "development_methods": ["TDD"], "hosts": hosts,
            "detected": detect["detected"],
        })
    };
    for ok in [json!(["omp"]), json!(["claude", "codex", "omp"])] {
        let plan = fx.call(
            &["project-plan", fx.p()],
            &serde_json::to_vec(&profile(ok.clone())).unwrap(),
        );
        assert!(plan["plan_digest"].is_string(), "{ok}: {plan}");
    }
    for bad in [
        json!(["claude", "codex", "omp", "claude"]),
        json!(["claude", "codex", "omp", "kiro"]),
        json!(["zcode"]),
    ] {
        let before = fx.tree();
        let v = fx.run_exe(
            EXE,
            &["project-plan", fx.p()],
            &serde_json::to_vec(&profile(bad.clone())).unwrap(),
        );
        assert_eq!(v["ok"], json!(false), "{bad}: {v}");
        assert_eq!(before, fx.tree(), "{bad}: project-plan wrote something");
    }
}

#[test]
fn the_omp_config_step_writes_nothing_and_prints_no_command() {
    let fx = Fx::new();
    fx.with_profile(false);
    let plan = fx.plan();
    let s = step(&plan, "omp-config");
    assert_eq!(s["state"], "skip", "{s}");
    assert_eq!(s["apply"], Value::Null, "{s}");
    assert!(
        s["reason"].as_str().unwrap().contains(".omp/config.yml"),
        "{s}"
    );
    for ready in steps(&plan).into_iter().filter(|x| x["state"] == "ready") {
        assert_eq!(fx.apply_step(ready)["ok"], json!(true), "{ready}");
    }
    assert!(
        !fx.root.join(".omp/config.yml").exists(),
        "the install phase created .omp/config.yml"
    );
}

#[test]
fn a_broken_profile_blocks_its_steps_and_not_the_host_hook_steps() {
    let fx = Fx::new();
    fx.with_profile(false);
    // The profile lives in the state folder; find it where PROJECT wrote it.
    let stored = [".moai/project.yaml", ".moai-cli/project.yaml"]
        .into_iter()
        .find(|rel| fx.root.join(rel).is_file())
        .expect("the profile file");
    fx.write(stored, "{not a profile\n");
    let plan = fx.plan();
    assert_eq!(step(&plan, "profile")["state"], "blocked", "{plan}");
    assert_eq!(step(&plan, "generation")["state"], "blocked", "{plan}");
    for name in ["codex-hooks", "codex-config", "omp-hooks"] {
        let s = step(&plan, name);
        assert_eq!(s["state"], "ready", "{name}: {s}");
        assert!(s["apply"]["argv"].is_array(), "{name}: {s}");
    }
    assert_eq!(step(&plan, "omp-config")["state"], "skip", "{plan}");
}

// ---- SPEC-20261121-remove-codex-project-layer: the retired hosts get no step of their own ----

/// The six steps earlier versions planned for the Codex and OMP hosts.
const RETIRED_STEPS: [&str; 6] = [
    "codex-hooks",
    "codex-config",
    "codex-skill",
    "omp-hooks",
    "omp-config",
    "omp-agents",
];

/// The files and folders earlier versions wrote for those hosts, as `retired-hosts` lists them.
const RETIRED_FILES: [(&str, &str); 6] = [
    (".codex/hooks.json", "{\"hooks\": {\"Notification\": []}}\n"),
    (".codex/config.toml", "[tui]\nstatus_line = [\"model\"]\n"),
    (".codex/agents/mine.toml", "name = \"mine\"\n"),
    (".agents/skills/moai/SKILL.md", "my own skill\n"),
    (".omp/extensions/moai-hooks.ts", "// mine\n"),
    (".omp/agents/mine.md", "mine\n"),
];

impl Fx {
    /// The entries under the project's `.codex`, `.omp` and `.agents` folders, bytes included.
    fn host_folders(&self) -> Vec<(PathBuf, Node)> {
        self.tree()
            .into_iter()
            .filter(|(p, _)| {
                [".codex", ".omp", ".agents"]
                    .iter()
                    .any(|dir| p.starts_with(self.root.join(dir)))
            })
            .collect()
    }
}

/// REQ-1 (AC-1): a full profile that still lists codex and omp plans the Claude steps and none
/// of the six host steps, and the printed undo order names none of them either.
#[test]
fn no_codex_or_omp_step_is_planned() {
    let fx = Fx::new();
    fx.with_profile(false);
    let plan = fx.plan();
    let names: Vec<&str> = steps(&plan)
        .iter()
        .map(|s| s["step"].as_str().unwrap())
        .collect();
    for retired in RETIRED_STEPS {
        assert!(!names.contains(&retired), "{retired} is planned: {names:?}");
    }
    for claude in [
        "profile",
        "generation",
        "claude-settings",
        "hooks",
        "github",
    ] {
        assert!(names.contains(&claude), "{claude} is missing: {names:?}");
    }
    let undo = plan["undo_order"].to_string();
    for retired in RETIRED_STEPS {
        assert!(
            !undo.contains(retired),
            "the undo order names {retired}: {undo}"
        );
    }
}

/// REQ-2 (AC-2): a profile that lists codex or omp loads and plans without refusal, and the
/// plan leaves its bytes alone (the profile step has nothing to write).
#[test]
fn a_profile_that_lists_codex_and_omp_still_loads() {
    for hosts in [
        json!(["claude", "codex", "omp"]),
        json!(["codex"]),
        json!(["omp"]),
    ] {
        let fx = Fx::new();
        fx.with_profile_extra(json!({"hosts": hosts}));
        let stored = [".moai/project.yaml", ".moai-cli/project.yaml"]
            .into_iter()
            .find(|rel| fx.root.join(rel).is_file())
            .expect("the profile file");
        let before = fs::read(fx.root.join(stored)).unwrap();
        let plan = fx.plan();
        let profile = step(&plan, "profile");
        assert_eq!(profile["state"], "unchanged", "{hosts}: {profile}");
        assert_eq!(profile["apply"], Value::Null, "{hosts}: {profile}");
        assert_eq!(
            fs::read(fx.root.join(stored)).unwrap(),
            before,
            "{hosts}: the plan rewrote the profile"
        );
        let text = String::from_utf8(before).unwrap();
        assert!(
            hosts
                .as_array()
                .unwrap()
                .iter()
                .all(|h| text.contains(h.as_str().unwrap())),
            "{hosts}: the stored profile lost a host: {text}"
        );
        assert_eq!(
            step(&plan, "retired-hosts")["state"],
            "skip",
            "{hosts}: {plan}"
        );
    }
}

/// REQ-3 (AC-3): the retired-hosts step is in every preview, with state skip and action info; it
/// lists the sorted paths that exist, prints no command and carries no digest.
#[test]
fn the_retired_hosts_step_only_informs() {
    // No profile and nothing found: the step is still there, with an empty list, and it is not
    // one of the steps that wait for the profile.
    let fx = Fx::new();
    let plan = fx.plan();
    let s = step(&plan, "retired-hosts");
    assert_eq!(s["state"], "skip", "{s}");
    assert_eq!(s["preview"], json!({"action": "info", "files": []}), "{s}");
    assert_eq!(s["apply"], Value::Null, "{s}");
    assert_eq!(s["plan_digest"], Value::Null, "{s}");
    assert_eq!(
        step(&plan, "profile")["state"],
        "needs_decision",
        "the other install steps still wait: {plan}"
    );

    // A full profile and every retired path present: the list is sorted and complete.
    let fx = Fx::new();
    fx.with_profile(false);
    for (rel, body) in RETIRED_FILES {
        fx.write(rel, body);
    }
    let plan = fx.plan();
    let s = step(&plan, "retired-hosts");
    assert_eq!(s["state"], "skip", "{s}");
    assert_eq!(
        s["preview"],
        json!({"action": "info", "files": [
            ".agents/skills/moai",
            ".codex/agents",
            ".codex/config.toml",
            ".codex/hooks.json",
            ".omp/agents",
            ".omp/extensions/moai-hooks.ts",
        ]}),
        "{s}"
    );
    assert_eq!(s["apply"], Value::Null, "{s}");
    assert_eq!(s["plan_digest"], Value::Null, "{s}");
    assert!(
        phase(&plan, "install")["steps"]
            .as_array()
            .unwrap()
            .iter()
            .any(|st| st["step"] == "retired-hosts"),
        "the step belongs to the install phase: {plan}"
    );
}

/// REQ-4 (AC-4): every file under `.codex/`, `.omp/` and the project skill copy stays byte for
/// byte through the plan and through the apply of every printed command.
#[test]
fn the_codex_and_omp_folders_stay_byte_identical() {
    let fx = Fx::new();
    fx.with_profile(true);
    // The two hook files stay absent: earlier versions would have written them, so a write
    // there shows in the listing below.
    for (rel, body) in RETIRED_FILES {
        if rel != ".codex/hooks.json" && rel != ".omp/extensions/moai-hooks.ts" {
            fx.write(rel, body);
        }
    }
    // A folder of the user's own, so a new file beside theirs would show in the listing.
    fx.write(".codex/notes.md", "mine\n");
    fx.write(".omp/notes.md", "mine\n");
    let before = fx.host_folders();
    assert!(before.len() >= RETIRED_FILES.len(), "{before:?}");

    let plan = fx.plan();
    assert_eq!(before, fx.host_folders(), "the plan changed a host folder");
    let mut applied = 0;
    for s in steps(&plan).into_iter().filter(|s| s["state"] == "ready") {
        let out = fx.apply_step(s);
        assert_eq!(out["ok"], json!(true), "{}: {out}", s["step"]);
        applied += 1;
    }
    assert!(applied >= 2, "the Claude steps still apply: {plan}");
    assert_eq!(before, fx.host_folders(), "an apply changed a host folder");
    let again = fx.plan();
    assert_eq!(
        before,
        fx.host_folders(),
        "the second plan changed a host folder"
    );
    assert_eq!(step(&again, "retired-hosts")["state"], "skip", "{again}");
}

fn argv_of(step: &Value) -> Vec<&str> {
    step["apply"]["argv"]
        .as_array()
        .unwrap_or_else(|| panic!("no printed command: {step}"))
        .iter()
        .map(|w| w.as_str().unwrap())
        .collect()
}

impl Fx {
    /// The entries under HOME only.
    fn home_tree(&self) -> Vec<(PathBuf, Node)> {
        self.tree()
            .into_iter()
            .filter(|(p, _)| p.starts_with(&self.home))
            .collect()
    }
}

#[test]
fn the_codex_config_step_prints_its_apply_with_the_status_line_on_stdin_and_settles() {
    let fx = Fx::new();
    fx.with_profile(false);
    let plan = fx.plan();
    let s = step(&plan, "codex-config");
    assert_eq!(s["state"], "ready", "{s}");
    let argv = argv_of(s);
    assert_eq!(&argv[..2], ["moai", "codex-config-apply"], "{argv:?}");
    assert_eq!(argv[2], fx.p(), "{argv:?}");
    assert_eq!(
        s["apply"]["stdin"],
        json!({"tui.status_line": "moai"}),
        "{s}"
    );
    assert!(s["plan_digest"].is_string(), "{s}");
    assert!(
        s["preview"]["config_path"]
            .as_str()
            .unwrap()
            .ends_with("config.toml"),
        "{s}"
    );
    assert_no_forbidden_words(&plan);
    let home = fx.home_tree();
    let out = fx.apply_step(s);
    assert_eq!(out["ok"], json!(true), "{out}");
    assert!(fx.root.join(".codex/config.toml").is_file());
    assert_eq!(home, fx.home_tree(), "the apply wrote into HOME");
    let again = fx.plan();
    let s = step(&again, "codex-config");
    assert_eq!(s["state"], "unchanged", "{s}");
    assert_eq!(s["apply"], Value::Null, "{s}");
}

#[test]
fn the_omp_agents_step_is_skipped_when_this_binary_has_no_install() {
    let fx = Fx::new();
    fx.with_profile(false);
    let plan = fx.plan();
    let s = step(&plan, "omp-agents");
    assert_eq!(s["state"], "skip", "{s}");
    assert_eq!(s["apply"], Value::Null, "{s}");
    assert!(s["reason"].as_str().unwrap().contains("installed"), "{s}");
}

#[test]
fn the_omp_agents_step_prints_its_apply_and_settles() {
    let fx = Fx::new();
    fx.with_profile(false);
    let exe = fx.install();
    let plan = fx.plan_installed(&exe);
    let s = step(&plan, "omp-agents");
    assert_eq!(s["state"], "ready", "{s}");
    let argv = argv_of(s);
    assert_eq!(&argv[..2], ["moai", "agents-project-apply"], "{argv:?}");
    assert_eq!(argv[2], fx.p(), "{argv:?}");
    assert!(argv[3].ends_with("install/skills/moai"), "{argv:?}");
    assert!(argv[4].ends_with("install/agents"), "{argv:?}");
    assert!(argv.contains(&"--plan-digest"), "{argv:?}");
    assert_eq!(s["apply"]["stdin"], Value::Null, "{s}");
    assert!(
        s["preview"]["shadows_user_agents"]
            .as_str()
            .unwrap()
            .contains("shadows"),
        "{s}"
    );
    let files = s["preview"]["files"].as_array().unwrap();
    assert!(!files.is_empty(), "{s}");
    assert!(files.iter().all(|f| f["action"] == "create"), "{s}");
    assert_no_forbidden_words(&plan);
    let home = fx.home_tree();
    let out = fx.apply_step(s);
    assert_eq!(out["ok"], json!(true), "{out}");
    let written = fs::read_dir(fx.root.join(".omp/agents")).unwrap().count();
    assert_eq!(written, files.len(), "{out}");
    assert_eq!(home, fx.home_tree(), "the apply wrote into HOME");
    let again = fx.plan_installed(&exe);
    let s = step(&again, "omp-agents");
    assert_eq!(s["state"], "unchanged", "{s}");
    assert_eq!(s["apply"], Value::Null, "{s}");
}

#[test]
fn an_agent_file_the_user_edited_stays_and_the_step_has_nothing_to_do() {
    let fx = Fx::new();
    fx.with_profile(false);
    let exe = fx.install();
    let plan = fx.plan_installed(&exe);
    assert_eq!(fx.apply_step(step(&plan, "omp-agents"))["ok"], json!(true));
    let file = fs::read_dir(fx.root.join(".omp/agents"))
        .unwrap()
        .next()
        .unwrap()
        .unwrap()
        .path();
    fs::write(&file, "mine\n").unwrap();
    let again = fx.plan_installed(&exe);
    let s = step(&again, "omp-agents");
    assert_eq!(s["state"], "unchanged", "{s}");
    assert_eq!(s["apply"], Value::Null, "{s}");
    assert!(
        s["preview"]["files"]
            .as_array()
            .unwrap()
            .iter()
            .any(|f| f["action"] == "preserved"),
        "{s}"
    );
    assert_eq!(fs::read_to_string(&file).unwrap(), "mine\n");
}

#[test]
fn a_linked_omp_folder_blocks_the_agents_step_and_nothing_lands_beyond_it() {
    let fx = Fx::new();
    fx.with_profile(false);
    let exe = fx.install();
    let beyond = fx.home.parent().unwrap().join("beyond");
    fs::create_dir_all(&beyond).unwrap();
    std::os::unix::fs::symlink(&beyond, fx.root.join(".omp")).unwrap();
    let plan = fx.plan_installed(&exe);
    let s = step(&plan, "omp-agents");
    assert_eq!(s["state"], "blocked", "{s}");
    assert_eq!(s["apply"], Value::Null, "{s}");
    assert!(fs::read_dir(&beyond).unwrap().next().is_none());
}

#[test]
fn a_status_line_the_user_changed_after_the_apply_is_kept_and_stays_byte_identical() {
    let fx = Fx::new();
    fx.with_profile(false);
    let plan = fx.plan();
    assert_eq!(
        fx.apply_step(step(&plan, "codex-config"))["ok"],
        json!(true)
    );
    let mine = "[tui]\nstatus_line = [\"model\"]\n";
    fx.write(".codex/config.toml", mine);
    let before = fx.tree();
    let again = fx.plan();
    let s = step(&again, "codex-config");
    assert_eq!(s["state"], "unchanged", "{s}");
    assert_eq!(s["apply"], Value::Null, "{s}");
    assert_eq!(before, fx.tree(), "the plan changed the user's file");
    assert_eq!(
        fs::read_to_string(fx.root.join(".codex/config.toml")).unwrap(),
        mine
    );
}

#[test]
fn the_codex_skill_step_is_skipped_when_this_binary_has_no_install() {
    let fx = Fx::new();
    fx.with_profile(false);
    let plan = fx.plan();
    let s = step(&plan, "codex-skill");
    assert_eq!(s["state"], "skip", "{s}");
    assert_eq!(s["apply"], Value::Null, "{s}");
    assert!(s["reason"].as_str().unwrap().contains("installed"), "{s}");
}

#[test]
fn the_codex_skill_step_prints_its_apply_with_both_flags_and_settles() {
    let fx = Fx::new();
    fx.with_profile(false);
    let exe = fx.install();
    let plan = fx.plan_installed(&exe);
    let s = step(&plan, "codex-skill");
    assert_eq!(s["state"], "ready", "{s}");
    let argv = argv_of(s);
    assert_eq!(
        &argv[..3],
        ["moai", "project-skill-apply", fx.p()],
        "{argv:?}"
    );
    assert_eq!(argv[3], "--plan-digest", "{argv:?}");
    assert_eq!(argv[5], "--permission-reference", "{argv:?}");
    assert_eq!(argv[6], PLACEHOLDER, "{argv:?}");
    assert_eq!(argv.len(), 7, "{argv:?}");
    let request = &s["apply"]["stdin"];
    assert_eq!(request["opt_in"], json!(true), "{s}");
    assert_eq!(request["hosts"], json!(["codex"]), "{s}");
    assert_eq!(request["home"], json!(fx.home.to_str().unwrap()), "{s}");
    assert!(
        request["source_dir"]
            .as_str()
            .unwrap()
            .ends_with("install/skills/moai"),
        "{s}"
    );
    // The step carries the digest its printed command names, and the install steps come in the
    // order of `INSTALL`.
    assert_eq!(s["plan_digest"], json!(argv[4]), "{s}");
    let order: Vec<&str> = phase(&plan, "install")["steps"]
        .as_array()
        .unwrap()
        .iter()
        .map(|st| st["step"].as_str().unwrap())
        .collect();
    assert_eq!(order, INSTALL, "{plan}");
    assert_eq!(s["preview"]["path"], ".agents/skills/moai", "{s}");
    assert_eq!(s["preview"]["action"], "create", "{s}");
    assert_no_forbidden_words(&plan);
    let home = fx.home_tree();
    let out = fx.apply_step(s);
    assert_eq!(out["ok"], json!(true), "{out}");
    assert!(fx.root.join(".agents/skills/moai/SKILL.md").is_file());
    assert_eq!(home, fx.home_tree(), "the apply wrote into HOME");
    let again = fx.plan_installed(&exe);
    let s = step(&again, "codex-skill");
    assert_eq!(s["state"], "unchanged", "{s}");
    assert_eq!(s["apply"], Value::Null, "{s}");
}

#[test]
fn a_moai_skill_in_the_users_codex_home_skips_the_project_copy() {
    let fx = Fx::new();
    fx.with_profile(false);
    let exe = fx.install();
    let skill = fx.home.join(".agents/skills/moai");
    fs::create_dir_all(&skill).unwrap();
    fs::write(skill.join("SKILL.md"), "mine\n").unwrap();
    let plan = fx.plan_installed(&exe);
    let s = step(&plan, "codex-skill");
    assert_eq!(s["state"], "skip", "{s}");
    assert_eq!(s["apply"], Value::Null, "{s}");
    assert!(s["reason"].as_str().unwrap().contains("HOME"), "{s}");
    assert!(!fx.root.join(".agents").exists());
    assert_eq!(
        fs::read_to_string(skill.join("SKILL.md")).unwrap(),
        "mine\n"
    );
}

#[test]
fn a_linked_agents_folder_blocks_the_codex_skill_step_and_nothing_lands_beyond_it() {
    let fx = Fx::new();
    fx.with_profile(false);
    let exe = fx.install();
    let beyond = fx.home.parent().unwrap().join("beyond-skill");
    fs::create_dir_all(&beyond).unwrap();
    std::os::unix::fs::symlink(&beyond, fx.root.join(".agents")).unwrap();
    let plan = fx.plan_installed(&exe);
    let s = step(&plan, "codex-skill");
    assert_eq!(s["state"], "blocked", "{s}");
    assert_eq!(s["apply"], Value::Null, "{s}");
    assert!(fs::read_dir(&beyond).unwrap().next().is_none());
}

/// Audit #3025 B2: a project that is the user's HOME, or lies in a host's configuration folder
/// under it, gets no `project-skill-apply` command: following it would create the user's own
/// `~/.agents/skills/moai`. The plan still writes nothing (`plan_installed` compares the trees).
#[test]
fn a_project_that_is_the_home_or_inside_its_agents_folder_gets_no_skill_command() {
    for inside in [None, Some(".agents/work"), Some(".codex/work")] {
        let base = Fx::new();
        let (tmp, home) = (base._tmp, base.home);
        let root = match inside {
            None => home.clone(),
            Some(rel) => {
                let p = home.join(rel);
                fs::create_dir_all(&p).unwrap();
                p
            }
        };
        let fx = Fx {
            _tmp: tmp,
            home,
            root,
        };
        fx.with_profile(false);
        let exe = fx.install();
        let plan = fx.plan_installed(&exe);
        let s = step(&plan, "codex-skill");
        assert_eq!(s["state"], "blocked", "{inside:?}: {s}");
        assert_eq!(s["apply"], Value::Null, "{inside:?}: {s}");
        assert!(
            !fx.home.join(".agents/skills/moai").exists(),
            "{inside:?}: the user's skill folder was created"
        );
    }
}

#[test]
fn a_project_skill_folder_that_is_not_ours_is_blocked_and_kept() {
    let fx = Fx::new();
    fx.with_profile(false);
    let exe = fx.install();
    fx.write(".agents/skills/moai/SKILL.md", "my own skill\n");
    let before = fx.tree();
    let plan = fx.plan_installed(&exe);
    let s = step(&plan, "codex-skill");
    assert_eq!(s["state"], "blocked", "{s}");
    assert_eq!(s["apply"], Value::Null, "{s}");
    assert_eq!(before, fx.tree(), "the plan changed the tree");
}

// ---- #2936 S-CYCLE: the hooks step after the settings moved, and the backup step's block ----

/// The words of a printed command.
fn words(step: &Value) -> Vec<&str> {
    step["apply"]["argv"]
        .as_array()
        .unwrap_or_else(|| panic!("no command: {step}"))
        .iter()
        .map(|w| w.as_str().unwrap())
        .collect()
}

/// AC 13: the Go hook entries went with the settings, so the hooks step offers the MoAI hook set
/// (`hooks-apply --host claude`), but only in a project whose settings a committed go-state apply
/// moved. A project that never had Go settings gets no such offer, and a reverted move takes
/// the offer back.
#[test]
fn the_hooks_step_needs_a_settings_move_receipt() {
    // No receipt: nothing to fill in, before and after the settings step.
    let plain = Fx::new();
    plain.with_profile(false);
    for round in ["before", "after"] {
        let plan = plain.plan();
        let hooks = step(&plan, "hooks");
        assert_eq!(hooks["state"], "unchanged", "{round}: {hooks}");
        assert_eq!(hooks["apply"], Value::Null, "{round}: {hooks}");
        plain.apply_ready();
    }

    // A committed move: the offer comes once the settings step has written the file.
    let fx = Fx::new();
    fx.go_project_with_moved_settings(&json!({"permissions": {"deny": ["Read(./.env)"]}}));
    fx.with_profile(false);
    let plan = fx.plan();
    let hooks = step(&plan, "hooks");
    assert_eq!(hooks["state"], "blocked", "{hooks}");
    assert_eq!(hooks["reason"], "re-run after the claude-settings apply");
    fx.apply_ready();
    let plan = fx.plan();
    let hooks = step(&plan, "hooks");
    assert_eq!(hooks["state"], "ready", "{hooks}");
    let argv = words(hooks);
    assert_eq!(&argv[..2], ["moai", "hooks-apply"], "{argv:?}");
    assert!(
        argv.windows(2).any(|w| w == ["--host", "claude"]),
        "{argv:?}"
    );
    assert_eq!(fx.apply_step(hooks)["ok"], json!(true), "{hooks}");
    assert!(fx.settings_json()["hooks"].is_object());
    let plan = fx.plan();
    assert_eq!(step(&plan, "hooks")["state"], "unchanged");

    // A reverted move offers nothing: the Go settings are back, so the backup phase has its
    // move to make again and every install step waits for it.
    let fx = Fx::new();
    let backup = fx.go_project_with_moved_settings(&json!({"permissions": {"deny": ["x"]}}));
    let rp = fx.call(&["go-state-revert-plan", fx.p(), "--backup", &backup], b"");
    let digest = rp["plan_digest"].as_str().unwrap().to_string();
    fx.call(
        &[
            "go-state-revert",
            fx.p(),
            "--backup",
            &backup,
            "--plan-digest",
            &digest,
            "--permission-reference",
            REF,
        ],
        b"",
    );
    assert!(fx.root.join(".claude/settings.json").is_file());
    fx.with_profile(false);
    let plan = fx.plan();
    let hooks = step(&plan, "hooks");
    assert_eq!(hooks["state"], "blocked", "{hooks}");
    assert_eq!(hooks["reason"], "re-run after the backup phase", "{hooks}");
    assert_eq!(hooks["apply"], Value::Null, "{hooks}");
}

/// AC 13 (edited): settings the user changed since the deploy move too, and the offer follows.
#[test]
fn the_hooks_offer_follows_edited_settings_too() {
    let fx = Fx::new();
    fx.go_project_moved(
        &json!({"permissions": {"deny": ["Read(./.env)"]}}),
        "{\"model\":\"local\"}\n",
        true,
    );
    fx.with_profile(false);
    fx.apply_ready();
    let plan = fx.plan();
    let hooks = step(&plan, "hooks");
    assert_eq!(hooks["state"], "ready", "{hooks}");
    assert_eq!(&words(hooks)[..2], ["moai", "hooks-apply"], "{hooks}");
}

/// AC 13 (reverted, folder stays): a reverted move takes the offer back even when the receipt
/// is still on disk. The Go folder stayed (local settings name it), so the stay-layout receipt
/// lives under `.moai-cli/backups/`; once the Go settings are gone again nothing else holds
/// the hooks step, so only the receipt state can.
#[test]
fn a_reverted_staying_folder_move_offers_no_hooks() {
    let fx = Fx::new();
    let backup = fx.go_project_moved(
        &json!({"permissions": {"deny": ["x"]}}),
        "{\"statusLine\":{\"type\":\"command\",\"command\":\".moai/status.sh\"}}\n",
        false,
    );
    assert!(
        fx.root.join(".moai-cli/backups").join(&backup).is_dir(),
        "the stay-layout backup is under .moai-cli/backups"
    );
    let rp = fx.call(&["go-state-revert-plan", fx.p(), "--backup", &backup], b"");
    let digest = rp["plan_digest"].as_str().unwrap().to_string();
    fx.call(
        &[
            "go-state-revert",
            fx.p(),
            "--backup",
            &backup,
            "--plan-digest",
            &digest,
            "--permission-reference",
            REF,
        ],
        b"",
    );
    assert!(fx.root.join(".moai-cli/backups").join(&backup).is_dir());
    fs::remove_file(fx.root.join(".claude/settings.json")).unwrap();
    fx.with_profile(false);
    let plan = fx.plan();
    let hooks = step(&plan, "hooks");
    assert_eq!(hooks["state"], "unchanged", "{hooks}");
    assert_eq!(hooks["apply"], Value::Null, "{hooks}");
}

/// AC 13 (no second move): the Go manifest keeps recording `.claude/settings.json` while the
/// Go folder stays, but once the install phase wrote this generation's settings the file is
/// not the Go tool's any more: a later preview offers no second move, and the hooks step is
/// the next thing to do.
#[test]
fn the_settings_the_install_phase_wrote_are_not_moved_again() {
    let fx = Fx::new();
    fx.go_project_moved(
        &json!({"permissions": {"deny": ["x"]}}),
        "{\"statusLine\":{\"type\":\"command\",\"command\":\".moai/status.sh\"}}\n",
        false,
    );
    fx.with_profile(false);
    fx.apply_ready();
    assert!(fx.root.join(".claude/settings.json").is_file());
    let after = fx.plan();
    let go = step(&after, "go-state");
    // The local settings still hold the Go folder (`blocked`, as before this change); what the
    // install phase wrote is not offered for a second move.
    assert_ne!(go["state"], "needs_decision", "{go}");
    let settings = go["preview"]["items"]
        .as_array()
        .unwrap()
        .iter()
        .find(|i| i["path"] == ".claude/settings.json")
        .unwrap();
    assert_eq!(settings["action"], "leave", "{settings}");
    assert!(go["preview"].get("settings_move").is_none(), "{go}");
    assert_eq!(
        step(&after, "hooks")["state"],
        "ready",
        "{}",
        step(&after, "hooks")
    );
}

/// AC 13 (not committed): only a committed apply's receipt offers the hooks. A receipt that
/// never reached `committed` (the move may have stopped midway) offers nothing.
#[test]
fn an_uncommitted_receipt_offers_no_hooks() {
    let fx = Fx::new();
    let backup = fx.go_project_moved(
        &json!({"permissions": {"deny": ["x"]}}),
        "{\"statusLine\":{\"type\":\"command\",\"command\":\".moai/status.sh\"}}\n",
        false,
    );
    // The settings step goes first: with a committed receipt the hooks step is then `ready`.
    fx.with_profile(false);
    fx.apply_ready();
    let ready = fx.plan();
    let hooks = step(&ready, "hooks");
    assert_eq!(hooks["state"], "ready", "{hooks}");
    let receipt = fx
        .root
        .join(".moai-cli/backups")
        .join(&backup)
        .join("go-state-receipt.json");
    let text = fs::read_to_string(&receipt).unwrap();
    let committed = "\"state\": \"committed\"";
    assert_eq!(text.matches(committed).count(), 1, "{text}");
    fs::write(&receipt, text.replace(committed, "\"state\": \"pending\"")).unwrap();
    let plan = fx.plan();
    let hooks = step(&plan, "hooks");
    assert_ne!(hooks["state"], "ready", "{hooks}");
    assert_eq!(hooks["apply"], Value::Null, "{hooks}");
}

/// AC 16: settings the Go manifest proves no longer block the backup step with
/// `settings reference .moai/`: the step has the move to offer. The same text in settings the
/// manifest does not record still blocks it, as before.
#[test]
fn settings_the_manifest_proves_do_not_hold_the_backup_step() {
    for recorded in [false, true] {
        let fx = Fx::new();
        let settings = r#"{"statusLine":{"type":"command","command":".moai/status.sh"}}"#;
        let hash = |t: &str| format!("sha256:{:x}", Sha256::digest(t.as_bytes()));
        fx.write(".moai/config/sections/a.yaml", "a: 1\n");
        fx.write(".claude/settings.json", settings);
        let mut files = serde_json::Map::new();
        for (rel, body) in [
            (".moai/config/sections/a.yaml", "a: 1\n"),
            (".claude/settings.json", settings),
        ] {
            if rel == ".claude/settings.json" && !recorded {
                continue;
            }
            let h = hash(body);
            files.insert(
                rel.into(),
                json!({"provenance": "template_managed", "template_hash": h,
                       "deployed_hash": h, "current_hash": h}),
            );
        }
        fx.write(
            ".moai/manifest.json",
            &json!({"version": "v3.2.1-rc.6", "files": files}).to_string(),
        );
        let plan = fx.plan();
        if recorded {
            let go = step(&plan, "go-state");
            assert_eq!(go["state"], "needs_decision", "{go}");
            assert_eq!(words(go)[1], "go-state-apply", "{go}");
            assert_eq!(go["preview"]["layout"], "folder_moves", "{go}");
            assert_eq!(
                go["preview"]["settings_move"]["path"], ".claude/settings.json",
                "{go}"
            );
        } else {
            assert_tied(&plan, "blocked", "settings reference .moai/");
            assert!(step(&plan, "go-state")["preview"]
                .get("settings_move")
                .is_none());
        }
    }
}

/// #2936 L2a: a Go project whose settings carry a Go-made `plansDirectory` (no launcher is
/// needed to clean that key); `recorded` says whether the Go manifest records the settings file.
fn go_project_with_go_made_plans_key(recorded: bool) -> Fx {
    let fx = Fx::new();
    let settings = r#"{"model":"opus","plansDirectory":".moai/plans"}"#;
    let hash = |t: &str| format!("sha256:{:x}", Sha256::digest(t.as_bytes()));
    fx.write(".moai/config/sections/a.yaml", "a: 1\n");
    fx.write(".claude/settings.json", settings);
    let mut files = serde_json::Map::new();
    for (rel, body) in [
        (".moai/config/sections/a.yaml", "a: 1\n"),
        (".claude/settings.json", settings),
    ] {
        if rel == ".claude/settings.json" && !recorded {
            continue;
        }
        let h = hash(body);
        files.insert(
            rel.into(),
            json!({"provenance": "template_managed", "template_hash": h,
                   "deployed_hash": h, "current_hash": h}),
        );
    }
    fx.write(
        ".moai/manifest.json",
        &json!({"version": "v3.2.1-rc.6", "files": files}).to_string(),
    );
    fx
}

/// #2936 L2a: the backup phase lists `go-settings` before `go-state`; with no Go-made key it is
/// `unchanged` and prints no command, and it counts in `all_unchanged`.
#[test]
fn the_backup_phase_lists_go_settings_before_go_state() {
    let fx = Fx::new();
    let plan = fx.plan();
    let names: Vec<&str> = phase(&plan, "backup")["steps"]
        .as_array()
        .unwrap()
        .iter()
        .map(|s| s["step"].as_str().unwrap())
        .collect();
    assert_eq!(names, ["go-settings", "go-state", "reports"], "{plan}");
    let gs = step(&plan, "go-settings");
    assert_eq!(gs["state"], "unchanged", "{gs}");
    assert_eq!(gs["apply"], Value::Null, "{gs}");
    assert_eq!(gs["preview"]["action"], "none", "{gs}");

    // A step that is not `unchanged` or `skip` keeps `all_unchanged` false.
    let fx = go_project_with_go_made_plans_key(false);
    let plan = fx.plan();
    assert_eq!(step(&plan, "go-settings")["state"], "ready", "{plan}");
    assert_eq!(plan["all_unchanged"], json!(false), "{plan}");
}

/// #2936 L2a: while the go-settings step has a command, the go-state step and every install
/// and merge step wait for a new preview (their digests would be stale after the settings are
/// rewritten); the go-state reason names the step; after the apply go-state offers its move.
#[test]
fn a_ready_go_settings_step_holds_the_later_steps_until_it_is_applied() {
    let fx = go_project_with_go_made_plans_key(false);
    let plan = fx.plan();
    let gs = step(&plan, "go-settings");
    assert_eq!(gs["state"], "ready", "{gs}");
    assert_eq!(words(gs)[1], "go-settings-apply", "{gs}");
    assert_eq!(
        gs["preview"]["go_keys"],
        json!([{"key": "plansDirectory", "change": "remove"}])
    );
    assert_eq!(gs["preview"]["launcher"]["state"], "not_needed", "{gs}");
    let own = fx.call(&["go-settings-plan", fx.p()], b"");
    assert_eq!(gs["plan_digest"], own["plan_digest"], "{gs}");
    let go = step(&plan, "go-state");
    assert_eq!(go["state"], "blocked", "{go}");
    assert!(
        go["reason"]
            .as_str()
            .unwrap()
            .contains("the go-settings step above clears the Go-made keys (plansDirectory)"),
        "{go}"
    );
    assert_eq!(go["apply"], Value::Null, "{go}");
    for name in INSTALL {
        let s = step(&plan, name);
        assert_eq!(s["reason"], BLOCKED_BY_BACKUP, "{name}: {s}");
        assert_eq!(s["apply"], Value::Null, "{name}: {s}");
    }
    for name in ["agents-local-merge", "claude-settings-merge"] {
        assert_eq!(step(&plan, name)["reason"], BLOCKED_BY_BACKUP, "{name}");
    }

    let out = fx.apply_step(gs);
    assert_eq!(out["ok"], json!(true), "{out}");
    let again = fx.plan();
    assert_eq!(step(&again, "go-settings")["state"], "unchanged", "{again}");
    let go = step(&again, "go-state");
    assert_eq!(go["state"], "needs_decision", "{go}");
    assert_eq!(words(go)[1], "go-state-apply", "{go}");
}

/// Audit B3 (reports step): moving the untracked report files out changes the report listing the
/// go-state preview reads, so while the reports step has a command the go-state step waits for a
/// new preview and names the reports step; once the reports are moved go-state offers its own.
#[test]
fn a_ready_reports_step_holds_the_go_state_command_until_it_is_applied() {
    let fx = go_project_with_go_made_plans_key(false);
    // The reports step asks git which files are tracked: the project is a repository.
    common::git::git_ok(&fx.root, &["init", "-q"]);
    let gs = step(&fx.plan(), "go-settings").clone();
    assert_eq!(fx.apply_step(&gs)["ok"], json!(true));
    let before = fx.plan();
    let go = step(&before, "go-state");
    assert_eq!(go["state"], "needs_decision", "{go}");
    assert_eq!(words(go)[1], "go-state-apply", "{go}");
    assert_eq!(step(&before, "reports")["state"], "unchanged", "{before}");

    fx.write(".moai/reports/earlier.md", "an earlier report");
    let plan = fx.plan();
    let reports = step(&plan, "reports");
    assert_eq!(reports["state"], "ready", "{reports}");
    assert_eq!(words(reports)[1], "reports-move-apply", "{reports}");
    let go = step(&plan, "go-state");
    assert_eq!(go["state"], "blocked", "{go}");
    assert_eq!(go["reason"], "re-run after the reports apply", "{go}");
    assert_eq!(go["apply"], Value::Null, "{go}");

    let out = fx.apply_step(reports);
    assert_eq!(out["ok"], json!(true), "{out}");
    let again = fx.plan();
    assert_eq!(step(&again, "reports")["state"], "unchanged", "{again}");
    let go = step(&again, "go-state");
    assert_ne!(go["reason"], "re-run after the reports apply", "{go}");
}

/// #2936 L2a: when the Go manifest proves the settings file the earlier generation's, the file
/// moves into the backup with the go-state step, so the go-settings step skips and leaves it
/// alone; the go-state step keeps its own command.
#[test]
fn go_settings_skips_a_settings_file_the_manifest_proves() {
    let fx = go_project_with_go_made_plans_key(true);
    let plan = fx.plan();
    let gs = step(&plan, "go-settings");
    assert_eq!(gs["state"], "skip", "{gs}");
    assert_eq!(gs["apply"], Value::Null, "{gs}");
    let go = step(&plan, "go-state");
    assert_eq!(go["state"], "needs_decision", "{go}");
    assert_eq!(words(go)[1], "go-state-apply", "{go}");
    assert_eq!(
        go["preview"]["settings_move"]["path"], ".claude/settings.json",
        "{go}"
    );
}
