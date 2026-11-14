//! #2936 R3 (live run 6879901297): a profile that is still the placeholder a first launch
//! writes must not generate a managed AGENTS.md section. `agents-plan`, `agents-apply`,
//! `generation-plan` and `generation-apply` refuse it and name `/moai project`, and
//! `project-update-plan` shows its generation step as waiting for the interview instead of a
//! step whose apply would refuse. Temporary HOME and project only.
#![cfg(unix)]

mod common;
use common::CleanEnv;
use serde_json::{json, Value};
use std::{
    fs,
    io::Write,
    path::PathBuf,
    process::{Command, Stdio},
};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;
const REF: &str = "test:2936-r3";
const WORKFLOW: &str = ".moai/workflows/review-flow/";

struct Fx {
    _tmp: TempDir,
    home: PathBuf,
    root: PathBuf,
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

    /// The whole envelope; an error is data here, not a panic.
    fn run(&self, args: &[&str], stdin: &Value) -> Value {
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
        match pipe.write_all(stdin.to_string().as_bytes()) {
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

    fn call(&self, args: &[&str], stdin: &Value) -> Value {
        let v = self.run(args, stdin);
        assert_eq!(v["ok"], json!(true), "{args:?}: {v}");
        v["result"].clone()
    }

    /// `<op>-plan` then `<op>-apply` on the plan's digest.
    fn plan_apply(&self, op: &str, input: &Value) -> Value {
        let plan = self.call(&[&format!("{op}-plan"), self.p()], input);
        let digest = plan["plan_digest"].as_str().unwrap().to_string();
        self.call(
            &[
                &format!("{op}-apply"),
                self.p(),
                "--plan-digest",
                &digest,
                "--permission-reference",
                REF,
            ],
            &plan,
        )
    }

    fn set_profile(&self, profile: &Value) {
        self.plan_apply("project", profile);
    }

    /// [`Fx::set_profile`] with the facts `project-detect` finds, as PROJECT writes it, so
    /// `project-update-plan` has no profile step left to apply before its later steps.
    fn set_detected_profile(&self, profile: &Value) {
        let detect = self.call(&["project-detect", self.p()], &json!(null));
        let mut profile = profile.clone();
        profile["detected"] = detect["detected"].clone();
        self.set_profile(&profile);
    }

    fn agents_md(&self) -> Option<Vec<u8>> {
        fs::read(self.root.join("AGENTS.md")).ok()
    }
}

/// The shape `launcher_init::default_profile` writes on a first launch.
fn placeholder() -> Value {
    json!({
        "schema_version": 2, "project_id": "demo",
        "purpose": "Created by moai-cli on first launch; refine it in the project interview",
        "tools": ["unspecified"], "development_methods": ["unspecified"], "hosts": ["claude"],
        "decisions": [
            {"id": "auto-init-tools", "topic": "tools", "value": "unspecified", "source": "default", "scope": "project"},
            {"id": "auto-init-methods", "topic": "development_methods", "value": "unspecified", "source": "default", "scope": "project"},
            {"id": "auto-init-hosts", "topic": "hosts", "value": "claude", "source": "default", "scope": "project"}
        ]
    })
}

fn interviewed() -> Value {
    json!({
        "schema_version": 2, "project_id": "demo", "purpose": "Build a small editor",
        "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude"],
    })
}

fn generation() -> Value {
    json!({
        "protocol_version": 1, "language": "en", "workflow_reference": WORKFLOW,
        "local_file": "none", "workflow_id": "review-flow",
        "native_roles": [], "skill_projections": []
    })
}

fn agents() -> Value {
    json!({"protocol_version": 1, "language": "en", "workflow_reference": WORKFLOW, "local_file": "none"})
}

/// A refusal that names the way out: kind `blocked`, `/moai project` in the message.
fn assert_blocked(v: &Value, what: &str) {
    assert_eq!(v["ok"], json!(false), "{what}: {v}");
    assert_eq!(v["error"]["kind"], json!("blocked"), "{what}: {v}");
    let message = v["error"]["message"].as_str().unwrap_or("");
    assert!(message.contains("/moai project"), "{what}: {v}");
    assert!(message.contains("placeholder"), "{what}: {v}");
}

/// The four commands refuse a placeholder profile and write nothing. The two applies are given
/// plans made while the profile was still interviewed, so only the guard can stop them.
#[test]
fn a_placeholder_profile_blocks_agents_and_generation_plan_and_apply() {
    let fx = Fx::new();
    fx.set_profile(&interviewed());
    fx.plan_apply("generation", &generation());
    let agents_plan = fx.call(&["agents-plan", fx.p()], &agents());
    let generation_plan = fx.call(&["generation-plan", fx.p()], &generation());
    fx.set_profile(&placeholder());
    let before = fx.agents_md();
    assert!(before.is_some(), "the interviewed run wrote AGENTS.md");

    assert_blocked(&fx.run(&["agents-plan", fx.p()], &agents()), "agents-plan");
    assert_blocked(
        &fx.run(&["generation-plan", fx.p()], &generation()),
        "generation-plan",
    );
    let agents_digest = agents_plan["plan_digest"].as_str().unwrap();
    assert_blocked(
        &fx.run(
            &[
                "agents-apply",
                fx.p(),
                "--plan-digest",
                agents_digest,
                "--permission-reference",
                REF,
            ],
            &agents_plan,
        ),
        "agents-apply",
    );
    let generation_digest = generation_plan["plan_digest"].as_str().unwrap();
    assert_blocked(
        &fx.run(
            &[
                "generation-apply",
                fx.p(),
                "--plan-digest",
                generation_digest,
                "--permission-reference",
                REF,
            ],
            &generation_plan,
        ),
        "generation-apply",
    );
    assert_eq!(fx.agents_md(), before, "AGENTS.md changed");
}

/// N/N+1: one answered axis makes the profile an interviewed one. Tools or methods that are
/// not `unspecified`, or any decision the user or the repository gave, lets the commands run.
#[test]
fn one_answered_axis_or_decision_is_enough_to_generate() {
    let mut tools_only = placeholder();
    tools_only["tools"] = json!(["Rust"]);
    let mut methods_only = placeholder();
    methods_only["development_methods"] = json!(["TDD"]);
    let mut user_decision = placeholder();
    user_decision["decisions"][0]["source"] = json!("user");
    for (name, profile) in [
        ("tools answered", tools_only),
        ("methods answered", methods_only),
        ("a user decision", user_decision),
    ] {
        let fx = Fx::new();
        fx.set_profile(&profile);
        let v = fx.run(&["generation-plan", fx.p()], &generation());
        assert_eq!(v["ok"], json!(true), "{name}: {v}");
    }
}

/// The preview step for generation waits for the interview and offers no apply, so the
/// printed plan never lists a command that would refuse.
#[test]
fn project_update_plan_sends_a_placeholder_profile_to_the_interview() {
    let fx = Fx::new();
    fx.set_detected_profile(&placeholder());
    let plan = fx.call(&["project-update-plan", fx.p()], &json!(null));
    let step = plan["phases"]
        .as_array()
        .unwrap()
        .iter()
        .flat_map(|p| p["steps"].as_array().into_iter().flatten())
        .find(|s| s["step"] == "generation")
        .unwrap_or_else(|| panic!("no generation step: {plan}"));
    assert_eq!(step["state"], json!("needs_decision"), "{step}");
    assert!(
        step["reason"].as_str().unwrap().contains("/moai project"),
        "{step}"
    );
    assert!(step["apply"].is_null(), "{step}");
    assert!(step["plan_digest"].is_null(), "{step}");
    // An interviewed profile still plans the generation step as before.
    let fx = Fx::new();
    fx.set_detected_profile(&interviewed());
    let plan = fx.call(&["project-update-plan", fx.p()], &json!(null));
    let step = plan["phases"]
        .as_array()
        .unwrap()
        .iter()
        .flat_map(|p| p["steps"].as_array().into_iter().flatten())
        .find(|s| s["step"] == "generation")
        .unwrap();
    assert_ne!(step["state"], json!("needs_decision"), "{step}");
}
