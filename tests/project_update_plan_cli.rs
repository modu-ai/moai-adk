//! #2936 S1: `moai project-update-plan <project>` — one read-only preview of the project
//! update, grouped into the three phases of the user's UPDATE flow (backup, install, merge).
//! Every case runs the built binary in a temporary project with an empty temporary HOME, and
//! every plan leaves the project tree byte-identical; the real checkout and HOME are never read.

#![cfg(unix)]

mod common;
use common::CleanEnv;
use moai_cli::{launcher_init::STATE_STOP, state_root::EARLIER_PROJECT_DIR};
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
const CONFIRM_PLACEHOLDER: &str = "<warning-confirmation>";
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
        let (digest, plan) = self.profile_plan(extra);
        self.call(&self.profile_apply(&digest), &plan);
    }

    /// [`Fx::with_profile`] for a Go project that has no earlier state folder yet. The apply
    /// is refused with nothing written (REQ-1 of the state-write stop), and with the folder
    /// there the same profile is written as before (REQ-3).
    fn with_go_profile(&self) {
        let (digest, plan) = self.profile_plan(json!({}));
        let before = self.tree();
        let v = self.run(&self.profile_apply(&digest), &plan);
        assert_eq!(v["ok"], json!(false), "{v}");
        assert_eq!(v["error"]["kind"], json!("invalid_input"), "{v}");
        let message = v["error"]["message"].as_str().unwrap();
        assert!(
            message.starts_with(STATE_STOP) && message.contains("moai go-state-plan"),
            "{message}"
        );
        assert_eq!(self.tree(), before, "a refused apply writes nothing");
        fs::create_dir(self.root.join(EARLIER_PROJECT_DIR)).unwrap();
        self.with_profile(false);
    }

    /// The `project-apply` words for the plan with this digest.
    fn profile_apply<'a>(&'a self, digest: &'a str) -> [&'a str; 6] {
        [
            "project-apply",
            self.p(),
            "--plan-digest",
            digest,
            "--permission-reference",
            REF,
        ]
    }

    /// The detected facts, with `extra` keys merged, through `project-plan`: the plan's
    /// digest and the bytes `project-apply` reads.
    fn profile_plan(&self, extra: Value) -> (String, Vec<u8>) {
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
        (digest, serde_json::to_vec(&plan).unwrap())
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
        // The permission reference is the only placeholder, plus the confirmation token of a
        // plan that carries a warning (the step then says so in its preview).
        let wants_token = step["preview"]["needs_confirmation"] == json!(true);
        let expected: Vec<&str> = if wants_token {
            vec![CONFIRM_PLACEHOLDER, PLACEHOLDER]
        } else {
            vec![PLACEHOLDER]
        };
        assert_eq!(
            argv.iter()
                .filter(|w| w.contains('<'))
                .map(String::as_str)
                .collect::<Vec<_>>(),
            expected,
            "placeholders of the printed command: {argv:?}"
        );
        let token = step["preview"]["confirm_token"]
            .as_str()
            .unwrap_or_default();
        let args: Vec<&str> = argv[1..]
            .iter()
            .map(|w| match w.as_str() {
                PLACEHOLDER => REF,
                CONFIRM_PLACEHOLDER => token,
                other => other,
            })
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

/// The install steps that wait for the profile or the backup phase.
const INSTALL: [&str; 5] = [
    "profile",
    "generation",
    "claude-settings",
    "hooks",
    "github",
];

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

/// SPEC-20261121-project-state-move REQ-8: a tree whose `.moai/` holds only the names the team
/// keeps in git, beside an earlier `.moai-cli/`, gets the same go-state step a plain earlier-root
/// tree gets: waiting for the user's pick with its apply command, never "decide outside this
/// route".
#[test]
fn a_kept_names_moai_beside_the_earlier_root_gets_the_go_state_move_step() {
    let fx = Fx::new();
    fx.write(".moai/project.yaml", "{}\n");
    fx.write(".moai/evals/e/case.json", "{}\n");
    fx.write(".moai/specs/SPEC-20261121-a/spec.md", "# a\n");
    fx.write(".moai-cli/manifests/generation.json", "{}\n");
    fx.write(".moai-cli/state/a.json", "{}\n");
    let plan = fx.plan();
    assert_eq!(plan["state_root"]["generation"], json!("none"), "{plan}");
    let go = step(&plan, "go-state");
    assert_eq!(go["state"], json!("needs_decision"), "{go}");
    assert_eq!(go["preview"]["layout"], json!("legacy_into_new"), "{go}");
    let argv = go["apply"]["argv"]
        .as_array()
        .unwrap_or_else(|| panic!("the step prints its apply command: {go}"));
    assert_eq!(argv[1], json!("go-state-apply"), "{go}");
    for s in steps(&plan) {
        assert!(
            !s["reason"]
                .as_str()
                .unwrap_or_default()
                .contains("decides outside this route"),
            "{s}"
        );
    }
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
    assert!(!paths.iter().any(|p| p == "CLAUDE.local.md"), "{paths:?}");
    assert!(paths.iter().any(|p| p == ".gitignore"), "{paths:?}");

    // A personal file is never shared, so `tracked` plans like `ignored` (the layout's REQ-7).
    for (choice, lists_local) in [("tracked", true), ("ignored", true)] {
        let fx = Fx::new();
        fx.with_profile(false);
        let mut request = step(&fx.plan(), "generation")["apply"]["stdin"]["request"].clone();
        request["local_file"] = json!(choice);
        request["picks"] = json!(["personal-file"]);
        fx.generate(&request);
        let plan = fx.plan();
        let generation = step(&plan, "generation");
        assert_eq!(generation["state"], "unchanged", "{choice}: {generation}");
        let paths = generation_paths(&plan);
        // The personal file is planned on the pick only, and the file stays as it was written.
        assert!(
            !paths.iter().any(|p| p == "CLAUDE.local.md"),
            "{choice}: {paths:?}"
        );
        assert!(fx.root.join("CLAUDE.local.md").is_file(), "{choice}");
        assert!(
            paths.iter().any(|p| p == ".gitignore"),
            "{choice}: {paths:?}"
        );
        // The block lists the personal file whenever the choice is not `none`, `tracked` too: a
        // personal file is never shared, so both `ignored` and `tracked` keep it out of git.
        let ignore = fs::read_to_string(fx.root.join(".gitignore")).unwrap();
        assert_eq!(
            ignore.lines().any(|l| l == "CLAUDE.local.md"),
            lists_local,
            "{choice}: {ignore}"
        );
    }

    // `ignored` over the user's own rule records no section: tracked or ignored is unknown.
    let fx = Fx::new();
    fx.with_profile(false);
    fx.write(".gitignore", "CLAUDE.local.md\n");
    let mut request = step(&fx.plan(), "generation")["apply"]["stdin"]["request"].clone();
    request["local_file"] = json!("ignored");
    request["picks"] = json!(["personal-file"]);
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
        request["picks"] = json!(["personal-file"]);
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

/// Generation moved to `CLAUDE.local.md`: a manifest of an earlier release that records the
/// personal file under the old name reads like one that records it under the new name. Without
/// that the record "file seen once, block lists it" is half a record and the update asks the user
/// again.
#[test]
fn a_manifest_that_names_agents_local_md_gives_the_same_choice() {
    let fx = Fx::new();
    fx.with_profile(false);
    let mut request = step(&fx.plan(), "generation")["apply"]["stdin"]["request"].clone();
    request["local_file"] = json!("ignored");
    request["picks"] = json!(["personal-file"]);
    fx.generate(&request);
    let manifest_path = fx.generation_manifest(&fx.plan());
    let mut manifest: Value = serde_json::from_slice(&fs::read(&manifest_path).unwrap()).unwrap();
    let recorded = manifest["created_once"].as_array_mut().unwrap();
    assert!(
        recorded.iter().any(|p| p == "CLAUDE.local.md"),
        "the apply records the personal file: {recorded:?}"
    );
    // A manifest of an earlier release records it under the old name: it reads the same.
    for p in recorded.iter_mut() {
        if *p == "CLAUDE.local.md" {
            *p = json!("AGENTS.local.md");
        }
    }
    fs::write(
        &manifest_path,
        serde_json::to_vec_pretty(&manifest).unwrap(),
    )
    .unwrap();
    let plan = fx.plan();
    let generation = step(&plan, "generation");
    assert_ne!(generation["state"], "needs_decision", "{generation}");
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

impl Fx {
    /// Makes the settled project's `CLAUDE.md` import `AGENTS.md`, below its managed block. A
    /// project like that is offered no pick (SPEC-20261122-claude-md-project-picks), so the
    /// personal-file merge step stays the one that offers an `AGENTS.local.md`.
    fn import_agents_md(&self) {
        let path = self.root.join("CLAUDE.md");
        let text = fs::read_to_string(&path).unwrap();
        fs::write(&path, format!("{text}\n@AGENTS.md\n")).unwrap();
    }
}

/// #2936 R10: a settled project with a personal `AGENTS.local.md` gets the merge as a decision
/// that prints the apply command; applying it settles the preview again. A project whose
/// `CLAUDE.md` does not import `AGENTS.md` is offered the picks instead, and the merge step skips
/// while they are pending.
#[test]
fn the_merge_step_offers_the_personal_file_and_settles_after_the_apply() {
    let fx = Fx::new();
    fx.settled();
    let pending = fx.root.join("AGENTS.local.md");
    fs::write(&pending, "# My notes\n\nReply in Korean.\n").unwrap();
    let plan = fx.plan();
    assert_eq!(
        step(&plan, "agents-picks")["state"],
        "needs_decision",
        "{plan}"
    );
    assert_eq!(step(&plan, "agents-local-merge")["state"], "skip", "{plan}");
    fs::remove_file(&pending).unwrap();
    fx.import_agents_md();
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

/// The audit of the personal-file move: a first `CLAUDE.local.md` over an `AGENTS.md` would hide
/// that file from Claude Code, so the merge on its own warns in plain words and writes nothing
/// until the confirmation token comes with it. The update preview offers the picks first for such
/// a project: they write `CLAUDE.md` before the personal file, so their preview carries no such
/// warning, and the merge step skips while they are pending.
#[test]
fn the_merge_preview_warns_and_the_apply_waits_for_the_confirmation() {
    let fx = Fx::new();
    // The project holds an AGENTS.md of its own and no CLAUDE file, so generation writes none.
    fx.write("AGENTS.md", "# rules\n");
    fx.write("AGENTS.local.md", "# My notes\n\nReply in Korean.\n");
    fx.with_profile(false);
    for s in steps(&fx.plan())
        .into_iter()
        .filter(|s| s["state"] == "ready")
    {
        assert_eq!(fx.apply_step(s)["ok"], json!(true), "{s}");
    }
    assert!(!fx.root.join("CLAUDE.md").exists());
    let plan = fx.plan();
    let picks = step(&plan, "agents-picks");
    assert_eq!(picks["state"], "needs_decision", "{picks}");
    assert!(
        !picks.to_string().contains("claude-md-and-agents-md"),
        "{picks}"
    );
    let merge = step(&plan, "agents-local-merge");
    assert_eq!(merge["state"], "skip", "{merge}");
    assert_eq!(merge["apply"], Value::Null, "{merge}");

    // The merge on its own: the plan warns and carries the token that answers the warning.
    let own = fx.call(&["agents-merge-plan", fx.p()], b"");
    let warnings = own["warnings"].to_string();
    assert!(
        warnings.contains("AGENTS.md") && warnings.contains("claude-md-and-agents-md"),
        "the warning names the file and the setting that keeps both read: {own}"
    );
    assert_eq!(own["needs_confirmation"], json!(true), "{own}");
    let token = own["confirm_token"].as_str().expect("a token").to_string();
    let digest = own["plan_digest"].as_str().unwrap().to_string();

    // Without the token: refused, the project is unchanged.
    let command = |token: Option<&str>| {
        let mut args = vec![
            "agents-merge-apply",
            fx.p(),
            "--plan-digest",
            digest.as_str(),
            "--permission-reference",
            REF,
        ];
        if let Some(token) = token {
            args.extend(["--confirm-warnings", token]);
        }
        fx.run(&args, b"")
    };
    let before = fx.tree();
    let refused = command(None);
    assert_eq!(refused["ok"], json!(false), "{refused}");
    assert_eq!(before, fx.tree(), "a refused apply wrote something");
    assert!(!fx.root.join("CLAUDE.local.md").exists());

    // With the token the merge is applied.
    assert_eq!(command(Some(&token))["ok"], json!(true));
    assert!(fx.root.join("CLAUDE.local.md").exists());
}

/// The audit of the personal-file move: in a git project the preview shows the one `.gitignore`
/// line and the apply writes it.
#[test]
fn the_merge_step_adds_the_ignore_line_in_a_git_project() {
    let fx = Fx::new();
    fx.settled();
    let init = Command::new("/usr/bin/git")
        .args(["init", "-q"])
        .arg(&fx.root)
        .clean_env()
        .status()
        .unwrap();
    assert!(init.success());
    fx.import_agents_md();
    fx.write("AGENTS.local.md", "# My notes\n\nReply in Korean.\n");
    let plan = fx.plan();
    let merge = step(&plan, "agents-local-merge");
    assert_eq!(merge["preview"]["gitignore"]["action"], "add", "{merge}");
    assert_eq!(fx.apply_step(merge)["ok"], json!(true), "{merge}");
    let ignore = fs::read_to_string(fx.root.join(".gitignore")).unwrap();
    assert!(ignore.lines().any(|l| l == "CLAUDE.local.md"), "{ignore}");
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

// ---- the install phase after the Codex and OMP steps were removed ----

/// The profile's `hosts` no longer change the steps: whatever it lists, the install phase holds
/// the five Claude-side steps and then the retired-hosts step, and no step prints a command that
/// writes for codex or omp.
#[test]
fn the_install_steps_do_not_depend_on_the_profile_hosts() {
    let mut want: Vec<&str> = INSTALL.to_vec();
    want.push("retired-hosts");
    for hosts in [
        json!(["claude", "codex", "omp"]),
        json!(["claude", "codex"]),
        json!(["claude", "omp"]),
        json!(["claude"]),
    ] {
        let fx = Fx::new();
        fx.with_profile_extra(json!({"hosts": hosts}));
        let plan = fx.plan();
        let names: Vec<&str> = phase(&plan, "install")["steps"]
            .as_array()
            .unwrap()
            .iter()
            .map(|s| s["step"].as_str().unwrap())
            .collect();
        assert_eq!(names, want, "{hosts}: {plan}");
        for s in steps(&plan) {
            for word in s["apply"]["argv"].as_array().into_iter().flatten() {
                let word = word.as_str().unwrap();
                assert!(
                    !matches!(word, "codex" | "omp")
                        && !word.starts_with("codex-")
                        && !word.starts_with("omp-"),
                    "{hosts}: {}: {s}",
                    s["step"]
                );
            }
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
fn a_broken_profile_blocks_its_steps_and_not_the_retired_hosts_step() {
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
    let retired = step(&plan, "retired-hosts");
    assert_eq!(retired["state"], "skip", "{retired}");
    assert_eq!(retired["apply"], Value::Null, "{retired}");
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
    fx.with_go_profile();
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

// ---- SPEC-20261125-update-declined-steps: a declined choice step stays declined ----

const DECLINE_OP: &str = "project-update-decline";
/// `sub` inside the project's state folder, spelled through the resolver.
fn state_rel(root: &Path, sub: &str) -> String {
    moai_cli::state_root::rel(root, sub)
}

/// The decline record of the project, found through the resolver.
fn declined_file(root: &Path) -> PathBuf {
    root.join(state_rel(root, "manifests/update-declined.json"))
}
const LOCAL_TEXT: &str = "# My notes\n\nReply in Korean.\n";

/// The paths whose entry differs between two trees (added, removed or changed).
fn changed_paths(before: &[(PathBuf, Node)], after: &[(PathBuf, Node)]) -> Vec<PathBuf> {
    let b: std::collections::BTreeMap<&PathBuf, &Node> =
        before.iter().map(|(p, n)| (p, n)).collect();
    let a: std::collections::BTreeMap<&PathBuf, &Node> =
        after.iter().map(|(p, n)| (p, n)).collect();
    let mut out: Vec<PathBuf> = b
        .keys()
        .chain(a.keys())
        .filter(|p| b.get(*p) != a.get(*p))
        .map(|p| (*p).clone())
        .collect();
    out.sort();
    out.dedup();
    out
}

impl Fx {
    /// A Go project whose personal `AGENTS.local.md` (and, with `settings`, its recorded
    /// `.claude/settings.json`) `go-state-apply` moved into its backup. Returns the backup stamp.
    fn go_project_with_backed_up_local(&self, settings: Option<&Value>) -> String {
        let mut files = serde_json::Map::new();
        let mut record = |rel: &str, body: String| {
            self.write(rel, &body);
            let h = format!("sha256:{:x}", Sha256::digest(body.as_bytes()));
            files.insert(
                rel.into(),
                json!({"provenance": "template_managed", "template_hash": h,
                       "deployed_hash": h, "current_hash": h}),
            );
        };
        record(".moai/config/sections/a.yaml", "a: 1\n".to_string());
        if let Some(s) = settings {
            record(".claude/settings.json", s.to_string());
        }
        self.write("AGENTS.local.md", LOCAL_TEXT);
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
        assert!(!self.root.join("AGENTS.local.md").exists(), "{applied}");
        applied["backup"].as_str().unwrap().to_string()
    }

    /// The profile, then every printed `ready` command until none is left.
    fn settle_install(&self) {
        self.with_profile(false);
        for _ in 0..4 {
            if !steps(&self.plan()).iter().any(|s| s["state"] == "ready") {
                return;
            }
            self.apply_ready();
        }
    }

    /// Runs a printed command (`argv` and `stdin` of `printed`): `moai` is this binary and the
    /// placeholder is a permission reference.
    fn run_printed(&self, printed: &Value) -> Value {
        let words: Vec<&str> = printed["argv"]
            .as_array()
            .unwrap_or_else(|| panic!("no printed command: {printed}"))
            .iter()
            .map(|w| w.as_str().unwrap())
            .collect();
        assert_eq!(words[0], "moai", "{words:?}");
        let args: Vec<&str> = words[1..]
            .iter()
            .map(|w| if *w == PLACEHOLDER { REF } else { *w })
            .collect();
        let stdin = match &printed["stdin"] {
            Value::Null => Vec::new(),
            v => serde_json::to_vec(v).unwrap(),
        };
        self.run(&args, &stdin)
    }

    /// The decline command with its own arguments, as a caller types it.
    fn decline_with(&self, step: &str, sha: &str, reference: &str) -> Value {
        self.run(
            &[
                DECLINE_OP,
                self.p(),
                "--step",
                step,
                "--source-sha256",
                sha,
                "--permission-reference",
                reference,
            ],
            b"",
        )
    }

    /// A project whose merge step wants a decision about the backed-up personal file.
    fn with_backed_up_local_decision(&self) -> (String, Value) {
        let backup = self.go_project_with_backed_up_local(None);
        self.settle_install();
        let plan = self.plan();
        let merge = step(&plan, "agents-local-merge").clone();
        assert_eq!(merge["state"], "needs_decision", "{merge}");
        (backup, merge)
    }

    /// Declines the merge step the way the preview prints it; returns its source sha256.
    fn decline_merge(&self, merge: &Value) -> String {
        let out = self.run_printed(&merge["preview"]["decline"]);
        assert_eq!(out["ok"], json!(true), "{out}");
        merge["preview"]["decline"]["source_sha256"]
            .as_str()
            .unwrap()
            .to_string()
    }

    /// Rewrites the backed-up personal file and the receipt's sha256 of it, so the copy is
    /// still the one the go-state apply recorded: only its bytes (the source sha256) change.
    fn rewrite_backed_up_local(&self, backup: &str, text: &str) {
        self.write(
            &state_rel(
                &self.root,
                &format!("backups/{backup}/root-files/AGENTS.local.md"),
            ),
            text,
        );
        let dir = self
            .root
            .join(state_rel(&self.root, &format!("backups/{backup}")));
        for entry in fs::read_dir(&dir).unwrap() {
            let path = entry.unwrap().path();
            let Ok(bytes) = fs::read(&path) else { continue };
            let Ok(mut receipt) = serde_json::from_slice::<Value>(&bytes) else {
                continue;
            };
            let Some(moves) = receipt["moves"].as_array_mut() else {
                continue;
            };
            for m in moves.iter_mut() {
                if m["path"] == "AGENTS.local.md" {
                    m["sha256"] = json!(format!("sha256:{:x}", Sha256::digest(text.as_bytes())));
                }
            }
            fs::write(&path, serde_json::to_vec_pretty(&receipt).unwrap()).unwrap();
            return;
        }
        panic!("no go-state receipt under {}", dir.display());
    }
}

/// REQ-1, REQ-2, REQ-14, REQ-15: the decline command records the step and the sha256 the preview
/// printed and writes nothing else; a stale sha256, an unknown step, a placeholder permission
/// reference and a link for the folder or the file are refused with nothing written.
#[test]
fn declined_record_is_written_for_the_printed_sha_and_nothing_else() {
    let fx = Fx::new();
    let (_backup, merge) = fx.with_backed_up_local_decision();
    let decline = &merge["preview"]["decline"];
    let sha = decline["source_sha256"].as_str().expect("a source sha256");
    assert_eq!(sha, merge["preview"]["source"]["sha256"], "{merge}");
    assert!(
        decline["effect"].as_str().is_some_and(|e| !e.is_empty()),
        "{merge}"
    );
    assert_eq!(decline["argv"][0], "moai", "{merge}");
    assert_eq!(decline["argv"][1], DECLINE_OP, "{merge}");
    assert_eq!(decline["ask_again_argv"], Value::Null, "{merge}");
    assert!(fx.root.join(state_rel(&fx.root, "manifests")).is_dir());
    assert!(!declined_file(&fx.root).exists());

    // Refused, each with the tree (project and HOME) byte-identical.
    let stale = format!("sha256:{}", "0".repeat(64));
    let refused = |what: &str, out: Value, before: &[(PathBuf, Node)], needle: &str| {
        assert_eq!(out["ok"], json!(false), "{what}: {out}");
        let message = out["error"]["message"].as_str().unwrap_or_default();
        assert!(message.contains(needle), "{what}: {out}");
        assert_eq!(
            before,
            fx.tree(),
            "{what}: a refused decline wrote something"
        );
    };
    let before = fx.tree();
    refused(
        "stale sha256",
        fx.decline_with("agents-local-merge", &stale, REF),
        &before,
        "source",
    );
    refused(
        "unknown step",
        fx.decline_with("go-state", sha, REF),
        &before,
        "go-state",
    );
    refused(
        "placeholder reference",
        fx.decline_with("agents-local-merge", sha, "placeholder"),
        &before,
        "permission",
    );
    // A link in place of the record file.
    fs::write(fx.root.join(state_rel(&fx.root, "elsewhere.txt")), "x\n").unwrap();
    std::os::unix::fs::symlink("../elsewhere.txt", declined_file(&fx.root)).unwrap();
    let before = fx.tree();
    refused(
        "linked record file",
        fx.decline_with("agents-local-merge", sha, REF),
        &before,
        "symbolic link",
    );
    fs::remove_file(declined_file(&fx.root)).unwrap();
    // A link in place of the manifests folder.
    fs::rename(
        fx.root.join(state_rel(&fx.root, "manifests")),
        fx.root.join(state_rel(&fx.root, "manifests-real")),
    )
    .unwrap();
    std::os::unix::fs::symlink(
        "manifests-real",
        fx.root.join(state_rel(&fx.root, "manifests")),
    )
    .unwrap();
    let before = fx.tree();
    refused(
        "linked folder",
        fx.decline_with("agents-local-merge", sha, REF),
        &before,
        "symbolic link",
    );
    fs::remove_file(fx.root.join(state_rel(&fx.root, "manifests"))).unwrap();
    fs::rename(
        fx.root.join(state_rel(&fx.root, "manifests-real")),
        fx.root.join(state_rel(&fx.root, "manifests")),
    )
    .unwrap();

    // The printed command records, and only the record file changes.
    let before = fx.tree();
    let out = fx.run_printed(decline);
    assert_eq!(out["ok"], json!(true), "{out}");
    let after = fx.tree();
    assert_eq!(
        changed_paths(&before, &after),
        [declined_file(&fx.root)],
        "{out}"
    );
    let record = fs::read_to_string(declined_file(&fx.root)).unwrap();
    assert!(
        record.contains("agents-local-merge") && record.contains(sha),
        "{record}"
    );
}

/// REQ-3, REQ-4, REQ-7: a declined merge reads `skip` with a reason that says so and names the
/// command that asks again, and the preview is `all_unchanged` again.
#[test]
fn declined_skip_reads_skip_and_names_the_way_back() {
    let fx = Fx::new();
    let (_backup, merge) = fx.with_backed_up_local_decision();
    assert_eq!(fx.plan()["all_unchanged"], json!(false));
    fx.decline_merge(&merge);
    let plan = fx.plan();
    let merge = step(&plan, "agents-local-merge");
    assert_eq!(merge["state"], "skip", "{merge}");
    let reason = merge["reason"].as_str().unwrap();
    assert!(
        reason.contains("declined")
            && reason.contains(DECLINE_OP)
            && reason.contains("--ask-again"),
        "{merge}"
    );
    assert_eq!(merge["apply"], Value::Null, "{merge}");
    let again: Vec<&str> = merge["preview"]["decline"]["ask_again_argv"]
        .as_array()
        .unwrap_or_else(|| panic!("no ask_again_argv: {merge}"))
        .iter()
        .map(|w| w.as_str().unwrap())
        .collect();
    assert_eq!(&again[..2], ["moai", DECLINE_OP], "{merge}");
    assert!(again.contains(&"--ask-again"), "{merge}");
    assert_eq!(plan["all_unchanged"], json!(true), "{plan}");
}

/// REQ-5: a record of another source sha256 is not read: the step asks again, and the record
/// stays in its file.
#[test]
fn declined_source_changes_ask_again() {
    let fx = Fx::new();
    let (backup, merge) = fx.with_backed_up_local_decision();
    let old = fx.decline_merge(&merge);
    assert_eq!(
        step(&fx.plan(), "agents-local-merge")["state"],
        "skip",
        "declined"
    );
    let record = fs::read(declined_file(&fx.root)).unwrap();
    fx.rewrite_backed_up_local(&backup, "# My notes\n\nReply in English.\n");
    let plan = fx.plan();
    let merge = step(&plan, "agents-local-merge");
    assert_eq!(merge["state"], "needs_decision", "{merge}");
    assert_ne!(merge["preview"]["decline"]["source_sha256"], json!(old));
    assert_eq!(merge["preview"]["decline"]["ask_again_argv"], Value::Null);
    assert_eq!(plan["all_unchanged"], json!(false));
    assert_eq!(fs::read(declined_file(&fx.root)).unwrap(), record);
}

/// REQ-6: `--ask-again` removes the one step's record and leaves the others.
#[test]
fn declined_ask_again_removes_only_that_step() {
    let fx = Fx::new();
    let settings = json!({"env": {"EDITOR": "vim"}, "model": "opus"});
    fx.go_project_with_backed_up_local(Some(&settings));
    fx.settle_install();
    let plan = fx.plan();
    let (local, merge) = (
        step(&plan, "agents-local-merge").clone(),
        step(&plan, "claude-settings-merge").clone(),
    );
    assert_eq!(local["state"], "needs_decision", "{local}");
    assert_eq!(merge["state"], "needs_decision", "{merge}");
    for s in [&local, &merge] {
        let out = fx.run_printed(&s["preview"]["decline"]);
        assert_eq!(out["ok"], json!(true), "{out}");
    }
    let plan = fx.plan();
    assert_eq!(step(&plan, "agents-local-merge")["state"], "skip");
    assert_eq!(step(&plan, "claude-settings-merge")["state"], "skip");

    let out = fx.run_printed(&json!({
        "argv": step(&plan, "agents-local-merge")["preview"]["decline"]["ask_again_argv"]
    }));
    assert_eq!(out["ok"], json!(true), "{out}");
    let plan = fx.plan();
    assert_eq!(
        step(&plan, "agents-local-merge")["state"],
        "needs_decision",
        "{plan}"
    );
    assert_eq!(step(&plan, "claude-settings-merge")["state"], "skip");
    let record = fs::read_to_string(declined_file(&fx.root)).unwrap();
    assert!(
        record.contains("claude-settings-merge") && !record.contains("agents-local-merge"),
        "{record}"
    );

    // A step with no record is left as it is: nothing is written.
    let before = fx.tree();
    let out = fx.run_printed(&json!({
        "argv": ["moai", DECLINE_OP, fx.p(), "--step", "agents-local-merge", "--ask-again",
                 "--permission-reference", PLACEHOLDER]
    }));
    assert_eq!(out["ok"], json!(true), "{out}");
    assert_eq!(
        before,
        fx.tree(),
        "an ask-again with no record wrote something"
    );
}

/// REQ-8: the merge step's reason names the backup path and says the file is a copy moved there.
#[test]
fn declined_reason_names_backup() {
    let fx = Fx::new();
    let (backup, merge) = fx.with_backed_up_local_decision();
    let reason = merge["reason"].as_str().unwrap();
    let path = state_rel(
        &fx.root,
        &format!("backups/{backup}/root-files/AGENTS.local.md"),
    );
    assert!(reason.contains(&path), "{merge}");
    assert!(
        reason.contains("copy moved") && reason.contains("earlier"),
        "{merge}"
    );
}

impl Fx {
    /// A Go project whose earlier settings were moved into the backup, and whose new
    /// `.claude/settings.json` already carries a status line of the user's own: the earlier
    /// `statusLine` is then a conflict, `env` and `model` merge.
    fn with_settings_conflict(&self) -> Value {
        let earlier = json!({
            "statusLine": {"type": "command", "command": "my-status"},
            "env": {"EDITOR": "vim"},
            "model": "opus",
        });
        self.go_project_with_backed_up_local(Some(&earlier));
        self.write(
            ".claude/settings.json",
            "{\"statusLine\":{\"type\":\"command\",\"command\":\"current-status\"}}\n",
        );
        self.settle_install();
        let plan = self.plan();
        step(&plan, "claude-settings-merge").clone()
    }
}

fn choice<'a>(step: &'a Value, id: &str) -> &'a Value {
    step["preview"]["choices"]
        .as_array()
        .unwrap_or_else(|| panic!("no choices: {step}"))
        .iter()
        .find(|c| c["id"] == id)
        .unwrap_or_else(|| panic!("no choice {id}: {step}"))
}

/// REQ-9, REQ-10, REQ-13: the conflict names the key and the file and lists `keep-mine` and
/// `skip`; `skip` is the decline command.
#[test]
fn declined_settings_conflict_names_the_keys_and_the_choices() {
    let fx = Fx::new();
    let merge = fx.with_settings_conflict();
    assert_eq!(merge["state"], "conflict", "{merge}");
    let reason = merge["reason"].as_str().unwrap();
    assert!(
        reason.contains("statusLine") && reason.contains(".claude/settings.json"),
        "{merge}"
    );
    assert!(
        !reason.contains("env"),
        "only the conflicting keys: {merge}"
    );
    let ids: Vec<&Value> = merge["preview"]["choices"]
        .as_array()
        .unwrap_or_else(|| panic!("no choices: {merge}"))
        .iter()
        .map(|c| &c["id"])
        .collect();
    assert_eq!(ids, ["keep-mine", "skip"], "{merge}");
    let skip = &choice(&merge, "skip")["apply"];
    assert_eq!(skip["argv"][1], DECLINE_OP, "{merge}");
    assert_eq!(
        skip["argv"], merge["preview"]["decline"]["argv"],
        "the skip choice is the decline command: {merge}"
    );
    let keep = &choice(&merge, "keep-mine")["apply"];
    assert_eq!(keep["argv"][1], "claude-settings-apply", "{merge}");
    assert_eq!(
        keep["stdin"]["statusLine"],
        json!({"type": "command", "command": "current-status"}),
        "{merge}"
    );
}

/// REQ-11: `keep-mine` run as printed leaves the conflicting key at the file's value and merges
/// the other earlier settings.
#[test]
fn declined_settings_conflict_keep_mine_leaves_the_key_and_merges_the_rest() {
    let fx = Fx::new();
    let merge = fx.with_settings_conflict();
    let out = fx.run_printed(&choice(&merge, "keep-mine")["apply"]);
    assert_eq!(out["ok"], json!(true), "{out}");
    let written = fx.settings_json();
    assert_eq!(
        written["statusLine"],
        json!({"type": "command", "command": "current-status"}),
        "{written}"
    );
    assert_eq!(written["env"], json!({"EDITOR": "vim"}), "{written}");
    assert_eq!(written["model"], json!("opus"), "{written}");
    // The settings apply does not record the file's value as MoAI's, so the earlier copy still
    // differs for that key: the next preview still reads the conflict, and the way to end it is
    // the `skip` choice, whose command declines the step.
    let plan = fx.plan();
    let merge = step(&plan, "claude-settings-merge");
    assert_eq!(merge["state"], "conflict", "{merge}");
    let skip = choice(merge, "skip");
    assert_eq!(
        skip["apply"]["argv"][1], "project-update-decline",
        "{merge}"
    );
    let out = fx.run_printed(&skip["apply"]);
    assert_eq!(out["ok"], json!(true), "{out}");
    let plan = fx.plan();
    assert_eq!(step(&plan, "claude-settings-merge")["state"], "skip");
}

/// REQ-13, REQ-3: `skip` run as printed declines the conflict; the step then reads `skip`.
#[test]
fn declined_settings_conflict_skip_declines_the_step() {
    let fx = Fx::new();
    let merge = fx.with_settings_conflict();
    let before = fx.settings_json();
    let out = fx.run_printed(&choice(&merge, "skip")["apply"]);
    assert_eq!(out["ok"], json!(true), "{out}");
    assert_eq!(
        fx.settings_json(),
        before,
        "a decline never writes the settings"
    );
    let plan = fx.plan();
    let merge = step(&plan, "claude-settings-merge");
    assert_eq!(merge["state"], "skip", "{merge}");
    assert!(
        merge["reason"].as_str().unwrap().contains("declined"),
        "{merge}"
    );
}

// ---- SPEC-20261125-update-next-say-ask-chain: `next`, `say`, `decline`, `ask`, `chain`, `finish` ----

const NEXT_WORDS: [&str; 5] = ["apply", "ask", "wait", "report", "none"];
/// REQ-18: the steps whose entries read `guarded`.
const GUARDED_STEPS: [&str; 6] = [
    "go-settings",
    "go-state",
    "reports",
    "claude-settings",
    "hooks",
    "claude-settings-merge",
];
const STATE_WORDS: [&str; 7] = [
    "unchanged",
    "skip",
    "optional",
    "ready",
    "needs_decision",
    "conflict",
    "blocked",
];
/// Every step id a plan prints (the retired-hosts ids come from the plan itself).
const STEP_IDS: [&str; 12] = [
    "go-settings",
    "go-state",
    "reports",
    "profile",
    "generation",
    "claude-settings",
    "hooks",
    "github",
    "agents-picks",
    "agents-local-merge",
    "claude-settings-merge",
    "retired-hosts",
];

/// The five projects of AC-1 and AC-3, each with its preview (taken with the tree compared
/// before and after): an empty project, a project with a profile, a Go project whose backup
/// step is open, a project whose `AGENTS.local.md` waits for the install, and a project with
/// one blocked step.
fn five_plans() -> Vec<(&'static str, Fx, Value)> {
    let empty = Fx::new();
    let profile = Fx::new();
    profile.with_profile(false);
    let go = Fx::new();
    go.go_project();
    let local = Fx::new();
    local.with_profile(false);
    local.write("AGENTS.local.md", "# mine\n\ntext\n");
    let blocked = Fx::new();
    blocked.with_profile(false);
    let outside = blocked.home.parent().unwrap().join("elsewhere.json");
    fs::write(&outside, "{}\n").unwrap();
    fs::create_dir_all(blocked.root.join(".claude")).unwrap();
    std::os::unix::fs::symlink(&outside, blocked.root.join(".claude/settings.json")).unwrap();
    [
        ("empty", empty),
        ("profile", profile),
        ("go-backup", go),
        ("local-waits-for-install", local),
        ("one-blocked-step", blocked),
    ]
    .into_iter()
    .map(|(label, fx)| {
        let plan = fx.plan();
        (label, fx, plan)
    })
    .collect()
}

/// The keys this SPEC adds: `ask` and `finish` on the plan (null or an object), `next`, `say`
/// and `decline` on every step.
fn assert_new_keys(label: &str, plan: &Value) {
    for key in ["ask", "finish"] {
        assert!(
            plan.get(key).is_some(),
            "{label}: the plan lacks {key}: {plan}"
        );
    }
    for s in steps(plan) {
        for key in ["next", "say", "decline"] {
            assert!(s.get(key).is_some(), "{label}: a step lacks {key}: {s}");
        }
    }
}

/// `plan["phases"][..]["steps"][..]["next"]` of the step `name`.
fn assert_next(plan: &Value, name: &str, want: &str) {
    let s = step(plan, name);
    assert_eq!(s["next"], json!(want), "{name}: {s}");
}

/// The entry of `list` (an array of objects with a `step`) for `id`.
fn entry<'a>(list: &'a Value, id: &str) -> &'a Value {
    list.as_array()
        .into_iter()
        .flatten()
        .find(|e| e["step"] == id)
        .unwrap_or_else(|| panic!("no entry for {id} in {list}"))
}

/// The `step` ids of an array of entries, in order.
fn entry_ids(list: &Value) -> Vec<&str> {
    list.as_array()
        .into_iter()
        .flatten()
        .filter_map(|e| e["step"].as_str())
        .collect()
}

/// The string ids of an array of strings.
fn strings(list: &Value) -> Vec<&str> {
    list.as_array()
        .into_iter()
        .flatten()
        .filter_map(Value::as_str)
        .collect()
}

/// The ids the `chain` says to preview again after `id` has run.
fn then_of<'a>(plan: &'a Value, id: &str) -> Vec<&'a str> {
    plan["chain"]["preview_again"]
        .as_array()
        .into_iter()
        .flatten()
        .filter(|e| e["after"] == id)
        .flat_map(|e| strings(&e["then"]))
        .collect()
}

/// `ask` is an object, `finish` is null.
fn assert_asks(plan: &Value) {
    assert!(plan["ask"].is_object(), "no ask: {plan}");
    assert_eq!(plan["finish"], Value::Null, "{plan}");
}

/// REQ-18: every entry of `ask` says whether its step is guarded, true for the six named steps.
fn assert_guarded(label: &str, plan: &Value) {
    for key in ["steps", "follows", "picks"] {
        for e in plan["ask"][key].as_array().into_iter().flatten() {
            let id = e["step"].as_str().unwrap_or_default();
            assert_eq!(
                e["guarded"],
                json!(GUARDED_STEPS.contains(&id)),
                "{label}: {key}: {e}"
            );
        }
    }
}

/// AC-1, REQ-1, REQ-2, REQ-3, REQ-4: every step carries `next`, `say` and `decline` next to the
/// keys it had, `next` is one of five words, the top-level keys stay, and the preview leaves the
/// project and HOME as they were (`Fx::plan` compares the whole tree before and after).
#[test]
fn next_say_ac1_every_step_carries_the_new_keys_and_keeps_the_old_ones() {
    let plans = five_plans();
    assert_eq!(plans.len(), 5);
    for (label, _fx, plan) in &plans {
        assert_new_keys(label, plan);
        for s in steps(plan) {
            let next = s["next"].as_str().unwrap_or_default();
            assert!(
                NEXT_WORDS.contains(&next),
                "{label}: next is one of five words: {s}"
            );
            assert!(s["say"].is_string(), "{label}: {s}");
            for key in ["step", "state", "reason", "preview", "plan_digest", "apply"] {
                assert!(s.get(key).is_some(), "{label}: a step lost {key}: {s}");
            }
        }
        for key in [
            "project",
            "detected",
            "state_root",
            "leftovers",
            "phases",
            "all_unchanged",
            "undo_order",
        ] {
            assert!(plan.get(key).is_some(), "{label}: the plan lost {key}");
        }
        let names: Vec<&str> = plan["phases"]
            .as_array()
            .unwrap()
            .iter()
            .map(|p| p["phase"].as_str().unwrap())
            .collect();
        assert_eq!(names, ["backup", "install", "merge"], "{label}");
    }
}

/// What REQ-5 to REQ-9 say `next` is, from what a step prints.
fn expected_next(
    state: &str,
    has_command: bool,
    needs_confirmation: bool,
    has_choices: bool,
    waits_behind: bool,
) -> &'static str {
    match state {
        "unchanged" | "skip" | "optional" => "none",
        "blocked" if waits_behind => "wait",
        "ready" | "needs_decision" | "conflict"
            if has_choices || (has_command && needs_confirmation) =>
        {
            "ask"
        }
        _ if has_command && !needs_confirmation && !has_choices => "apply",
        _ => "report",
    }
}

/// AC-2, REQ-5 to REQ-9: `next_of` agrees with the rules for every state and every combination
/// of command, confirmation word, choices and waiting that a step can print. A confirmation word
/// belongs to a printed command, and a blocked step prints neither command nor choices.
#[test]
fn next_say_ac2_the_table_holds_for_every_combination() {
    use moai_cli::project_update::guide::next_of;
    let mut checked = 0;
    for state in STATE_WORDS {
        for bits in 0..16u8 {
            let (has_command, needs_confirmation, has_choices, waits_behind) =
                (bits & 1 != 0, bits & 2 != 0, bits & 4 != 0, bits & 8 != 0);
            if needs_confirmation && !has_command {
                continue;
            }
            if state == "blocked" && (has_command || has_choices) {
                continue;
            }
            assert_eq!(
                next_of(
                    state,
                    has_command,
                    needs_confirmation,
                    has_choices,
                    waits_behind
                ),
                expected_next(
                    state,
                    has_command,
                    needs_confirmation,
                    has_choices,
                    waits_behind
                ),
                "state={state} command={has_command} confirmation={needs_confirmation} \
                 choices={has_choices} waits={waits_behind}"
            );
            checked += 1;
        }
    }
    // Six states with 12 combinations each (16 minus the 4 with a confirmation word and no
    // command) and the blocked state with 2: every combination was reached.
    assert_eq!(checked, 6 * 12 + 2, "combinations checked");
}

/// AC-2, REQ-5 to REQ-9: the named steps read as the table says, in real previews.
#[test]
fn next_say_ac2_named_steps_read_as_the_table_says() {
    // Unchanged and skip read `none`; an install step with no profile prints no command: `report`.
    let empty = Fx::new();
    let plan = empty.plan();
    assert_next(&plan, "go-state", "none");
    assert_next(&plan, "agents-local-merge", "none");
    for name in INSTALL {
        assert_next(&plan, name, "report");
    }

    // A ready generation prints a plain command: `apply`.
    let profile = Fx::new();
    profile.with_profile(false);
    let plan = profile.plan();
    assert_eq!(step(&plan, "generation")["state"], "ready", "{plan}");
    assert_next(&plan, "generation", "apply");
    assert_next(&plan, "profile", "none");

    // The backup step is a decision with a plain command; the install steps wait behind it.
    let go = Fx::new();
    go.go_project();
    let plan = go.plan();
    assert_next(&plan, "go-state", "apply");
    for name in INSTALL {
        assert_next(&plan, name, "wait");
    }

    // A settings merge conflict prints choices: `ask`.
    let conflict = Fx::new();
    let merge = conflict.with_settings_conflict();
    assert_eq!(merge["state"], "conflict", "{merge}");
    assert_eq!(merge["next"], json!("ask"), "{merge}");

    // A go-state hold with no command, and a blocked step that waits for nothing: `report`.
    let interrupted = Fx::new();
    fs::create_dir_all(interrupted.root.join(".moai-backup/config")).unwrap();
    let plan = interrupted.plan();
    let hold = step(&plan, "go-state");
    assert_eq!(hold["state"], "needs_decision", "{hold}");
    assert_eq!(hold["apply"], Value::Null, "{hold}");
    assert_next(&plan, "go-state", "report");
    let blocked = five_plans().remove(4);
    assert_eq!(step(&blocked.2, "claude-settings")["state"], "blocked");
    assert_next(&blocked.2, "claude-settings", "report");

    // A declined step reads `none`, and so does the personal-file merge that skips while the
    // picks are pending.
    let declined = Fx::new();
    let (_backup, merge) = declined.with_backed_up_local_decision();
    assert_eq!(merge["next"], json!("apply"), "{merge}");
    declined.decline_merge(&merge);
    let plan = declined.plan();
    assert_next(&plan, "agents-local-merge", "none");
    let pending = Fx::new();
    pending.settled();
    pending.write("AGENTS.local.md", LOCAL_TEXT);
    let plan = pending.plan();
    assert_eq!(step(&plan, "agents-picks")["state"], "needs_decision");
    assert_eq!(step(&plan, "agents-local-merge")["state"], "skip");
    assert_next(&plan, "agents-local-merge", "none");
}

/// The words of `text`, lower-cased, split at everything but letters, digits, `_` and `-`.
fn plain_words(text: &str) -> Vec<String> {
    text.to_lowercase()
        .split(|c: char| !(c.is_ascii_alphanumeric() || c == '_' || c == '-'))
        .map(|w| w.trim_matches('-').to_string())
        .filter(|w| !w.is_empty())
        .collect()
}

/// What breaks REQ-10 in one sentence for a person: not one short line ending in a full stop, or
/// a step id, a state word, an issue number, a SPEC id, `digest`, `argv`, `<` or a backtick.
fn sentence_problems(text: &str, ids: &[String]) -> Vec<String> {
    let mut problems = Vec::new();
    if text.is_empty() {
        problems.push("empty".to_string());
    }
    if text.contains('\n') || text.contains('\r') {
        problems.push("more than one line".into());
    }
    if text.chars().count() > 200 {
        problems.push("longer than 200 characters".into());
    }
    if !text.ends_with('.') {
        problems.push("no final full stop".into());
    }
    let words = plain_words(text);
    for w in &words {
        if ids.contains(w) {
            problems.push(format!("step id {w}"));
        }
        if STATE_WORDS.contains(&w.as_str()) {
            problems.push(format!("state word {w}"));
        }
    }
    if text
        .as_bytes()
        .windows(2)
        .any(|w| w[0] == b'#' && w[1].is_ascii_digit())
    {
        problems.push("an issue number".into());
    }
    let lower = text.to_lowercase();
    for needle in ["spec-", "digest", "argv", "<", "`"] {
        if lower.contains(needle) {
            problems.push(format!("contains {needle}"));
        }
    }
    problems
}

/// Every sentence the plan adds for a person: the `say` of each step, the question and the
/// option sentences and the entry sentences of `ask`, and the `say` of `finish`.
fn plan_sentences(plan: &Value) -> Vec<(String, String)> {
    let text = |v: &Value| v.as_str().unwrap_or_default().to_string();
    let mut out = Vec::new();
    for s in steps(plan) {
        out.push((format!("say of {}", s["step"]), text(&s["say"])));
    }
    let ask = &plan["ask"];
    if ask.is_object() {
        out.push(("ask.question".into(), text(&ask["question"])));
        for o in ask["options"].as_array().into_iter().flatten() {
            out.push((format!("option {}", o["id"]), text(&o["say"])));
        }
        for key in ["steps", "follows", "picks"] {
            for e in ask[key].as_array().into_iter().flatten() {
                out.push((format!("ask.{key} {}", e["step"]), text(&e["say"])));
            }
        }
    }
    if plan["finish"].is_object() {
        out.push(("finish.say".into(), text(&plan["finish"]["say"])));
    }
    out
}

/// AC-3, REQ-10: in every plan of AC-1 each sentence is one line of at most 200 characters ending
/// in a full stop, and names no step id, state word, issue number, SPEC id, digest, command or
/// backtick. Each plan has exactly one of `ask` and `finish`, so a sentence of each kind is read.
#[test]
fn next_say_ac3_every_sentence_is_one_plain_line() {
    let (mut asked, mut finished) = (0, 0);
    for (label, _fx, plan) in five_plans() {
        assert_new_keys(label, &plan);
        let mut ids: Vec<String> = STEP_IDS.iter().map(|s| s.to_string()).collect();
        ids.extend(
            steps(&plan)
                .iter()
                .map(|s| s["step"].as_str().unwrap().to_string()),
        );
        let (ask, finish) = (plan["ask"].is_object(), plan["finish"].is_object());
        assert!(ask != finish, "{label}: one of ask and finish: {plan}");
        asked += usize::from(ask);
        finished += usize::from(finish);
        for (what, text) in plan_sentences(&plan) {
            let problems = sentence_problems(&text, &ids);
            assert!(
                problems.is_empty(),
                "{label}: {what}: {text:?}: {problems:?}"
            );
        }
    }
    assert!(
        asked > 0 && finished > 0,
        "asked {asked}, finished {finished}"
    );
}

/// AC-4, REQ-11, REQ-12: a step's `decline` is the object of its preview, copied as it is; a
/// declined step reads `none` and its `decline` carries the command that asks again; a step with
/// no `decline` object in its preview has `decline` null.
#[test]
fn next_say_ac4_decline_copies_the_preview_and_a_declined_step_reads_none() {
    let fx = Fx::new();
    let (_backup, _decision) = fx.with_backed_up_local_decision();
    let plan = fx.plan();
    let merge = step(&plan, "agents-local-merge");
    let decline = &merge["preview"]["decline"];
    assert!(decline.is_object(), "{merge}");
    assert_eq!(merge.get("decline"), Some(decline), "{merge}");
    // Every step of the plan: a copy where the preview has one, null where it has none.
    let check_all = |label: &str, plan: &Value| {
        for s in steps(plan) {
            let want = if s["preview"]["decline"].is_object() {
                s["preview"]["decline"].clone()
            } else {
                Value::Null
            };
            assert_eq!(s.get("decline"), Some(&want), "{label}: {s}");
        }
    };
    check_all("before the decline", &plan);
    assert_eq!(
        step(&plan, "generation").get("decline"),
        Some(&Value::Null),
        "a step that cannot be declined"
    );

    let merge = merge.clone();
    fx.decline_merge(&merge);
    let after = fx.plan();
    let declined = step(&after, "agents-local-merge");
    assert_eq!(declined["state"], "skip", "{declined}");
    assert_eq!(declined["next"], json!("none"), "{declined}");
    assert!(
        declined["decline"]["ask_again_argv"].is_array(),
        "{declined}"
    );
    check_all("after the decline", &after);

    // The settings conflict prints its decline too.
    let conflict = Fx::new();
    let merge = conflict.with_settings_conflict();
    assert_eq!(
        merge.get("decline"),
        Some(&merge["preview"]["decline"]),
        "{merge}"
    );
    assert!(merge["decline"].is_object(), "{merge}");
    check_all("settings conflict", &conflict.plan());
}

/// AC-5, REQ-13 to REQ-15, REQ-17, REQ-18: with the install done and only an `AGENTS.local.md`
/// added, the picks step has a recommended choice, so it goes into `ask.steps` with its `pick`
/// `move`; the options are `apply`, `choose`, `leave` with `apply` recommended.
#[test]
fn next_say_ac5_a_recommended_pick_goes_into_steps_with_its_choice() {
    let fx = Fx::new();
    fx.settled();
    fx.write("AGENTS.local.md", LOCAL_TEXT);
    let plan = fx.plan();
    assert_eq!(step(&plan, "agents-picks")["state"], "needs_decision");
    assert_asks(&plan);
    let ask = &plan["ask"];
    assert!(ask["question"].is_string(), "{ask}");
    let picks = entry(&ask["steps"], "agents-picks");
    assert_eq!(picks["pick"], json!("move"), "{picks}");
    assert_eq!(picks["guarded"], json!(false), "{picks}");
    assert!(picks["say"].is_string(), "{picks}");
    let ids: Vec<&Value> = ask["options"]
        .as_array()
        .unwrap_or_else(|| panic!("no options: {ask}"))
        .iter()
        .map(|o| &o["id"])
        .collect();
    assert_eq!(ids, ["apply", "choose", "leave"], "{ask}");
    assert_eq!(ask["options"][0]["recommended"], json!(true), "{ask}");
    assert_ne!(ask["options"][1]["recommended"], json!(true), "{ask}");
    assert_guarded("settled + AGENTS.local.md", &plan);
}

/// AC-5, REQ-13, REQ-16, REQ-18: a project whose generation is ready has it in `steps`, and the
/// merge steps that wait behind it in `follows`, each with the `after` that names `generation`.
#[test]
fn next_say_ac5_a_waiting_step_goes_into_follows_after_the_step_it_waits_behind() {
    let fx = Fx::new();
    fx.with_profile(false);
    fx.write("AGENTS.local.md", "# mine\n\ntext\n");
    let plan = fx.plan();
    assert_eq!(step(&plan, "generation")["state"], "ready", "{plan}");
    assert_asks(&plan);
    let ask = &plan["ask"];
    let generation = entry(&ask["steps"], "generation");
    assert_eq!(generation["pick"], Value::Null, "{generation}");
    for name in ["agents-picks", "agents-local-merge"] {
        assert_next(&plan, name, "wait");
        let e = entry(&ask["follows"], name);
        assert!(e["say"].is_string(), "{e}");
        assert_eq!(e["guarded"], json!(false), "{e}");
        assert!(strings(&e["after"]).contains(&"generation"), "{e}");
    }
    assert!(
        !entry_ids(&ask["steps"]).contains(&"agents-local-merge"),
        "{ask}"
    );
    assert_guarded("profile + AGENTS.local.md", &plan);
}

/// AC-5, REQ-14, REQ-17, REQ-18: a settings merge conflict has no recommended choice, so it goes
/// into `picks`; while another step is runnable the options are `apply`, `choose`, `leave`, and
/// with `steps` empty they are `choose` (recommended) and `leave`. The guarded flag is true for
/// the settings merge and for the backup steps, false for the others.
#[test]
fn next_say_ac5_a_conflict_goes_into_picks_and_empty_steps_recommend_choose() {
    let fx = Fx::new();
    let conflict = fx.with_settings_conflict();
    assert_eq!(conflict["state"], "conflict", "{conflict}");
    let plan = fx.plan();
    assert_asks(&plan);
    let ask = &plan["ask"];
    let merge = entry(&ask["picks"], "claude-settings-merge");
    assert_eq!(merge["guarded"], json!(true), "{merge}");
    assert!(merge["say"].is_string(), "{merge}");
    assert!(
        !entry_ids(&ask["steps"]).contains(&"claude-settings-merge"),
        "{ask}"
    );
    let local = step(&plan, "agents-local-merge");
    assert_eq!(local["next"], json!("apply"), "{local}");
    assert!(
        entry_ids(&ask["steps"]).contains(&"agents-local-merge"),
        "{ask}"
    );
    let ids: Vec<&Value> = ask["options"]
        .as_array()
        .unwrap_or_else(|| panic!("no options: {ask}"))
        .iter()
        .map(|o| &o["id"])
        .collect();
    assert_eq!(ids, ["apply", "choose", "leave"], "{ask}");
    assert_guarded("settings conflict", &plan);

    // With the personal-file merge declined, nothing is left to apply: the conflict is a pick.
    fx.decline_merge(local);
    let plan = fx.plan();
    assert_asks(&plan);
    let ask = &plan["ask"];
    assert_eq!(entry_ids(&ask["steps"]), Vec::<&str>::new(), "{ask}");
    assert!(
        entry_ids(&ask["picks"]).contains(&"claude-settings-merge"),
        "{ask}"
    );
    let ids: Vec<&Value> = ask["options"]
        .as_array()
        .unwrap_or_else(|| panic!("no options: {ask}"))
        .iter()
        .map(|o| &o["id"])
        .collect();
    assert_eq!(ids, ["choose", "leave"], "{ask}");
    assert_eq!(ask["options"][0]["recommended"], json!(true), "{ask}");
    assert_guarded("settings conflict, merge declined", &plan);

    // The Go project: the move is a guarded step in `steps`; the install steps follow it.
    let go = Fx::new();
    go.go_project();
    let plan = go.plan();
    assert_asks(&plan);
    assert_eq!(
        entry(&plan["ask"]["steps"], "go-state")["guarded"],
        json!(true),
        "{plan}"
    );
    assert_guarded("go project", &plan);

    // The Go settings step is guarded too, and the profile and generation steps are not.
    let go = go_project_with_go_made_plans_key(false);
    let plan = go.plan();
    assert_asks(&plan);
    assert_eq!(
        entry(&plan["ask"]["steps"], "go-settings")["guarded"],
        json!(true),
        "{plan}"
    );
    assert_guarded("go-made key", &plan);
}

/// AC-6, REQ-19: `apply_order` is the ids of `ask.steps`, in the same order.
#[test]
fn next_say_ac6_apply_order_is_the_ids_of_the_steps() {
    let go = Fx::new();
    go.go_project();
    let profile = Fx::new();
    profile.with_profile(false);
    profile.write("AGENTS.local.md", "# mine\n\ntext\n");
    let conflict = Fx::new();
    conflict.with_settings_conflict();
    for (label, fx) in [("go", &go), ("profile", &profile), ("conflict", &conflict)] {
        let plan = fx.plan();
        assert_asks(&plan);
        let order = strings(&plan["chain"]["apply_order"]);
        assert!(!order.is_empty(), "{label}: {plan}");
        assert_eq!(order, entry_ids(&plan["ask"]["steps"]), "{label}: {plan}");
    }
}

/// AC-6, REQ-6, REQ-16, REQ-20: for each of the six reasons a step waits (the backup, the
/// go-settings step, the reports step, the profile, the settings and the install), the waiting
/// step reads `wait`, is in the `follows` of `ask` after the step its reason speaks of, and is in
/// the `then` of that step in `chain.preview_again`.
#[test]
fn next_say_ac6_each_waiting_reason_names_the_step_it_waits_behind() {
    // (project label, the plan, [(waiting step, the reason it prints, the step it waits behind)])
    type Waits<'a> = Vec<(&'a str, &'a str, &'a str)>;
    let mut cases: Vec<(&str, Value, Waits)> = Vec::new();

    // The backup: every install step waits behind the move.
    let go = Fx::new();
    go.go_project();
    cases.push((
        "backup",
        go.plan(),
        INSTALL
            .iter()
            .map(|n| (*n, BLOCKED_BY_BACKUP, "go-state"))
            .collect(),
    ));

    // The go-settings step: the install steps wait behind it (the go-state step of this project
    // is held by the Go-made keys and prints no command).
    let go_made = go_project_with_go_made_plans_key(false);
    let waits: Vec<(&str, &str, &str)> = INSTALL
        .iter()
        .map(|n| (*n, BLOCKED_BY_BACKUP, "go-settings"))
        .collect();
    cases.push(("go-settings", go_made.plan(), waits));

    // The reports step: the go-state command waits behind it.
    let reports = go_project_with_go_made_plans_key(false);
    common::git::git_ok(&reports.root, &["init", "-q"]);
    let gs = step(&reports.plan(), "go-settings").clone();
    assert_eq!(reports.apply_step(&gs)["ok"], json!(true));
    reports.write(
        &state_rel(&reports.root, "reports/earlier.md"),
        "an earlier report",
    );
    cases.push((
        "reports",
        reports.plan(),
        vec![("go-state", "re-run after the reports apply", "reports")],
    ));

    // The profile: the generation step waits behind it.
    let profile = Fx::new();
    profile.with_profile(false);
    profile.write(
        "Cargo.toml",
        "[package]\nname = \"demo\"\nversion = \"0.1.0\"\nedition = \"2021\"\n",
    );
    cases.push((
        "profile",
        profile.plan(),
        vec![("generation", "re-run after the profile apply", "profile")],
    ));

    // The settings: the hooks step waits behind them.
    let settings = Fx::new();
    settings.go_project_with_moved_settings(&json!({"permissions": {"deny": ["Read(./.env)"]}}));
    settings.with_profile(false);
    cases.push((
        "settings",
        settings.plan(),
        vec![(
            "hooks",
            "re-run after the claude-settings apply",
            "claude-settings",
        )],
    ));

    // The install: the personal-file merge waits behind the generation step, and the settings
    // merge behind the settings step.
    let install = Fx::new();
    install.with_profile(false);
    install.write("AGENTS.local.md", "# mine\n\ntext\n");
    cases.push((
        "install (generation)",
        install.plan(),
        vec![(
            "agents-local-merge",
            "re-run after the install phase",
            "generation",
        )],
    ));
    let merge = Fx::new();
    merge.go_project_with_moved_settings(&user_settings());
    merge.with_profile(false);
    cases.push((
        "install (settings)",
        merge.plan(),
        vec![(
            "claude-settings-merge",
            "re-run after the install phase",
            "claude-settings",
        )],
    ));

    for (label, plan, waits) in &cases {
        assert_asks(plan);
        for (name, reason, behind) in waits {
            let s = step(plan, name);
            assert!(
                s["reason"].as_str().unwrap_or_default().contains(reason),
                "{label}: {name} prints its reason: {s}"
            );
            assert_eq!(s["next"], json!("wait"), "{label}: {name}: {s}");
            let follows = entry(&plan["ask"]["follows"], name);
            assert!(
                strings(&follows["after"]).contains(behind),
                "{label}: {name} follows {behind}: {follows}"
            );
            assert!(
                then_of(plan, behind).contains(name),
                "{label}: {name} is in the then of {behind}: {}",
                plan["chain"]
            );
            assert!(
                strings(&plan["chain"]["apply_order"]).contains(behind),
                "{label}: {behind} is run first: {}",
                plan["chain"]
            );
        }
    }
}

/// AC-6, REQ-16: a step that waits behind a go-state step with no command is in `follows`, and
/// its `after` names only steps that can run: never the go-state step, which prints no command,
/// and `then_of` the go-state step is empty. The project holds a stopped move (the go-state step
/// wants the user's decision and prints no command) and an untracked earlier report (the reports
/// step can run), so the plan still asks; `after` is then exactly the reports step.
#[test]
fn next_say_ac6_a_wait_behind_a_step_without_a_command_is_not_after_it() {
    let fx = Fx::new();
    common::git::git_ok(&fx.root, &["init", "-q"]);
    fx.write(
        &format!(
            "{}/reports/earlier.md",
            moai_cli::state_root::EARLIER_PROJECT_DIR
        ),
        "an earlier report",
    );
    fs::create_dir_all(fx.root.join(".moai-backup/config")).unwrap();
    let plan = fx.plan();
    assert_eq!(step(&plan, "reports")["state"], "ready", "{plan}");
    let hold = step(&plan, "go-state");
    assert_eq!(hold["state"], "needs_decision", "{hold}");
    assert_eq!(hold["apply"], Value::Null, "{hold}");
    assert_next(&plan, "go-state", "report");
    assert_asks(&plan);
    let ask = &plan["ask"];
    for name in INSTALL {
        let s = step(&plan, name);
        assert_eq!(s["reason"], BLOCKED_BY_BACKUP, "{name}: {s}");
        assert_eq!(s["next"], json!("wait"), "{name}: {s}");
        let follows = entry(&ask["follows"], name);
        assert_eq!(strings(&follows["after"]), ["reports"], "{name}: {follows}");
    }
    assert!(then_of(&plan, "go-state").is_empty(), "{}", plan["chain"]);
}

/// AC-7, REQ-21, REQ-22, REQ-23: a project that is all current has `ask` null, a `finish` that
/// says nothing is left and an empty `open`, and every step reads `none`; an empty project has
/// `ask` null and the five install steps in `finish.open`, and `all_unchanged` stays false; a plan
/// with a step to run has `finish` null.
#[test]
fn next_say_ac7_nothing_to_ask_ends_with_finish() {
    let fx = Fx::new();
    fx.settled();
    let plan = fx.plan();
    assert_eq!(plan["all_unchanged"], json!(true), "{plan}");
    assert_eq!(plan["ask"], Value::Null, "{plan}");
    let finish = &plan["finish"];
    assert!(finish.is_object(), "{plan}");
    assert!(
        finish["say"]
            .as_str()
            .is_some_and(|s| s.to_lowercase().contains("nothing")),
        "{finish}"
    );
    assert_eq!(finish["open"], json!([]), "{finish}");
    for s in steps(&plan) {
        assert_eq!(s["next"], json!("none"), "{s}");
    }

    let empty = Fx::new();
    let plan = empty.plan();
    assert_eq!(plan["all_unchanged"], json!(false), "{plan}");
    assert_eq!(plan["ask"], Value::Null, "{plan}");
    let finish = &plan["finish"];
    assert!(finish["say"].is_string(), "{plan}");
    let mut open: Vec<&str> = finish["open"]
        .as_array()
        .unwrap_or_else(|| panic!("no open list: {finish}"))
        .iter()
        .map(|e| {
            e.as_str()
                .or_else(|| e["step"].as_str())
                .unwrap_or_default()
        })
        .collect();
    open.sort_unstable();
    let mut install = INSTALL.to_vec();
    install.sort_unstable();
    assert_eq!(open, install, "{finish}");

    let profile = Fx::new();
    profile.with_profile(false);
    let plan = profile.plan();
    assert_asks(&plan);
}

/// Runs the command a step prints, or the command of its recommended choice.
fn run_step_or_pick(fx: &Fx, step: &Value) -> Value {
    if step["apply"].is_object() {
        return fx.run_printed(&step["apply"]);
    }
    let pick = step["preview"]["choices"]
        .as_array()
        .into_iter()
        .flatten()
        .find(|c| c["recommended"] == json!(true))
        .unwrap_or_else(|| panic!("no command and no recommended choice: {step}"));
    fx.run_printed(&pick["apply"])
}

/// Follows the `chain` the way the update does: run `apply_order` as printed, preview again, and
/// repeat until `ask` is null. After each round the steps of `then` no longer wait. Returns the
/// last preview and the number of rounds.
fn follow_chain(fx: &Fx) -> (Value, usize) {
    for round in 1..=6 {
        let plan = fx.plan();
        if plan["ask"].is_null() {
            return (plan, round - 1);
        }
        let order: Vec<String> = strings(&plan["chain"]["apply_order"])
            .into_iter()
            .map(String::from)
            .collect();
        assert!(!order.is_empty(), "round {round}: {plan}");
        for id in &order {
            let out = run_step_or_pick(fx, step(&plan, id));
            assert_eq!(out["ok"], json!(true), "round {round}: {id}: {out}");
        }
        let again = fx.plan();
        for id in &order {
            for later in then_of(&plan, id) {
                assert_ne!(
                    step(&again, later)["next"],
                    json!("wait"),
                    "round {round}: {later} still waits after {id}: {again}"
                );
            }
        }
    }
    panic!("the chain did not end in six rounds: {}", fx.plan());
}

/// AC-8, REQ-19, REQ-20, REQ-21: following the chain ends in a plan with `ask` null, a `finish`
/// and `all_unchanged` true, for a project whose merge steps wait for the install and for a
/// settled project whose `CLAUDE.md` imports `AGENTS.md` and holds a personal file.
#[test]
fn next_say_ac8_following_the_chain_ends_in_finish() {
    let waits = Fx::new();
    waits.with_profile(false);
    waits.write("AGENTS.local.md", LOCAL_TEXT);
    let imports = Fx::new();
    imports.settled();
    imports.import_agents_md();
    imports.write("AGENTS.local.md", LOCAL_TEXT);
    for (label, fx) in [
        ("waits for the install", &waits),
        ("imports AGENTS.md", &imports),
    ] {
        let (last, rounds) = follow_chain(fx);
        assert!(rounds >= 1, "{label}: the chain ran no round");
        assert_eq!(last["ask"], Value::Null, "{label}: {last}");
        assert!(last["finish"].is_object(), "{label}: {last}");
        assert_eq!(last["all_unchanged"], json!(true), "{label}: {last}");
    }
}
