//! SPEC-20261120-generation-request-record: `generation-apply` records the project agent and
//! skill request in `generation.json`, so the generation step of `project-update-plan` rebuilds
//! it instead of stopping at `needs_decision`. Every case runs the built binary in a temporary
//! project with a temporary HOME; the real checkout and HOME are never read.

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
const REF: &str = "test:generation-request";
const PLACEHOLDER: &str = "<permission-reference>";

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

    /// A profile written the way PROJECT writes it, with `extra` keys merged in.
    fn with_profile(&self, extra: Value) {
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

    fn plan(&self) -> Value {
        self.call(&["project-update-plan", self.p()], b"")
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

    /// Runs one step's printed command with the permission reference filled in.
    fn apply_step(&self, step: &Value) -> Value {
        let argv: Vec<String> = step["apply"]["argv"]
            .as_array()
            .unwrap_or_else(|| panic!("a ready step prints its command: {step}"))
            .iter()
            .map(|w| w.as_str().unwrap().to_string())
            .collect();
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

    /// The generation manifest's path, spelled by the state-folder resolver.
    fn manifest_path(&self) -> PathBuf {
        self.root.join(moai_cli::state_root::rel(
            &self.root,
            "manifests/generation.json",
        ))
    }

    fn manifest(&self) -> Value {
        let path = self.manifest_path();
        serde_json::from_slice(&fs::read(&path).unwrap_or_else(|e| panic!("{path:?}: {e}")))
            .unwrap()
    }
}

fn steps(plan: &Value) -> Vec<&Value> {
    plan["phases"]
        .as_array()
        .unwrap()
        .iter()
        .flat_map(|p| p["steps"].as_array().into_iter().flatten())
        .collect()
}

fn step<'a>(plan: &'a Value, name: &str) -> &'a Value {
    steps(plan)
        .into_iter()
        .find(|s| s["step"] == name)
        .unwrap_or_else(|| panic!("no step {name}: {plan}"))
}

/// Written out in full: the record holds the request as its own types render it, defaults
/// included, so a request that already spells them reads back as the same JSON.
fn role() -> Value {
    json!({"role_id": "prj-test-runner", "host": "claude",
           "purpose": "Runs the test command and reports what failed",
           "read_only_intent": false, "tools": []})
}

fn skill() -> Value {
    json!({"id": "prj-checks", "hosts": [],
           "purpose": "The checks to run before a change is done"})
}

/// The request the generation step prints for a project whose profile is applied, with the
/// given project agents and skills added.
fn request_with(fx: &Fx, roles: Vec<Value>, skills: Vec<Value>) -> Value {
    let mut request = step(&fx.plan(), "generation")["apply"]["stdin"]["request"].clone();
    assert!(request.is_object(), "no request printed: {request}");
    request["native_roles"] = Value::Array(roles);
    if !skills.is_empty() {
        request["project_skills"] = Value::Array(skills);
    }
    request
}

/// A project with one project agent and one project skill generated, then every other printed
/// `ready` command applied, so the update preview is settled.
fn generated_project() -> Fx {
    settled_project().0
}

/// `generated_project` and the request that built it.
fn settled_project() -> (Fx, Value) {
    let fx = Fx::new();
    fx.with_profile(json!({}));
    let request = request_with(&fx, vec![role()], vec![skill()]);
    fx.generate(&request);
    for s in steps(&fx.plan())
        .into_iter()
        .filter(|s| s["state"] == "ready")
    {
        assert_eq!(fx.apply_step(s)["ok"], json!(true), "{s}");
    }
    (fx, request)
}

#[test]
fn the_applied_request_is_recorded_as_the_request_itself() {
    let fx = Fx::new();
    fx.with_profile(json!({}));
    let request = request_with(&fx, vec![role()], vec![skill()]);
    fx.generate(&request);
    let record = &fx.manifest()["generation_request"];
    assert_eq!(record["native_roles"], request["native_roles"], "{record}");
    assert_eq!(
        record["project_skills"], request["project_skills"],
        "{record}"
    );
    assert_eq!(
        record["skill_projections"], request["skill_projections"],
        "{record}"
    );

    // An empty request leaves no key, so a manifest without agents or skills is as before.
    let plain = Fx::new();
    plain.with_profile(json!({}));
    let empty = request_with(&plain, Vec::new(), Vec::new());
    plain.generate(&empty);
    assert!(
        plain.manifest().get("generation_request").is_none(),
        "{}",
        plain.manifest()
    );
}

#[test]
fn a_generated_agent_and_skill_read_unchanged_on_the_next_plan() {
    let fx = generated_project();
    let plan = fx.plan();
    let generation = step(&plan, "generation");
    assert_eq!(generation["state"], "unchanged", "{generation}");
    assert_eq!(generation["apply"], Value::Null, "{generation}");
    let files = generation["preview"]["files"].to_string();
    assert!(
        files.contains("prj-test-runner") && files.contains("prj-checks"),
        "the plan covers the agent and the skill: {files}"
    );
    assert_eq!(plan["all_unchanged"], json!(true), "{plan}");
}

#[test]
fn a_profile_change_plans_the_rebuild_of_a_recorded_skill() {
    let fx = generated_project();
    fx.with_profile(json!({"decisions": [
        {"id": "test_command", "topic": "test_command", "value": "cargo test",
         "source": "user", "scope": "project"}
    ]}));
    let plan = fx.plan();
    let generation = step(&plan, "generation");
    assert_eq!(generation["state"], "ready", "{generation}");
    assert!(
        generation["apply"]["argv"].is_array(),
        "the rebuild prints its command: {generation}"
    );
    let skill_row = generation["preview"]["files"]
        .as_array()
        .unwrap()
        .iter()
        .find(|f| f["path"].as_str().unwrap().contains("prj-checks"))
        .unwrap_or_else(|| panic!("no prj-checks row: {generation}"));
    assert_ne!(skill_row["action"], "remove", "{skill_row}");
    assert_eq!(fx.apply_step(generation)["ok"], json!(true));
    let again = fx.plan();
    assert_eq!(step(&again, "generation")["state"], "unchanged", "{again}");
}

#[test]
fn a_later_apply_replaces_the_record_and_an_agents_apply_keeps_it() {
    let fx = Fx::new();
    fx.with_profile(json!({}));
    let request = request_with(&fx, vec![role()], vec![skill()]);
    fx.generate(&request);
    let first = fx.manifest()["generation_request"].clone();
    assert_eq!(
        first["native_roles"].as_array().unwrap().len(),
        1,
        "{first}"
    );

    // An agents-only apply carries the record over unchanged.
    let agents = json!({
        "protocol_version": 1, "language": request["language"],
        "workflow_reference": request["workflow_reference"],
        "local_file": request["local_file"],
    });
    agents_apply(&fx, &agents);
    assert_eq!(fx.manifest()["generation_request"], first);

    // A later generation apply with one role fewer replaces the record whole.
    let mut fewer = request.clone();
    fewer["native_roles"] = json!([]);
    fx.generate(&fewer);
    let second = fx.manifest();
    let record = &second["generation_request"];
    assert!(
        record.get("native_roles").is_none()
            || record["native_roles"].as_array().is_some_and(Vec::is_empty),
        "{record}"
    );
    assert_eq!(
        record["project_skills"], request["project_skills"],
        "{record}"
    );

    // A request with nothing left drops the record, so the key never outlives its files.
    let mut none = request.clone();
    none["native_roles"] = json!([]);
    none["project_skills"] = json!([]);
    fx.generate(&none);
    assert!(
        fx.manifest().get("generation_request").is_none(),
        "{}",
        fx.manifest()
    );

    // An agents-only apply of a manifest with no record writes no key for it, not even a null.
    agents_apply(&fx, &agents);
    assert!(
        fx.manifest().get("generation_request").is_none(),
        "{}",
        fx.manifest()
    );
}

/// `agents-plan` with `agents`, then `agents-apply` of that plan.
fn agents_apply(fx: &Fx, agents: &Value) {
    let plan = fx.call(
        &["agents-plan", fx.p()],
        &serde_json::to_vec(agents).unwrap(),
    );
    fx.call(
        &[
            "agents-apply",
            fx.p(),
            "--plan-digest",
            plan["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            REF,
        ],
        &serde_json::to_vec(&plan).unwrap(),
    );
}

/// The manifest of `fx` with its `generation_request` replaced by `record`.
fn set_record(fx: &Fx, record: Value) {
    let mut manifest = fx.manifest();
    manifest["generation_request"] = record;
    let mut bytes = serde_json::to_vec_pretty(&manifest).unwrap();
    bytes.push(b'\n');
    fs::write(fx.manifest_path(), bytes).unwrap();
}

/// The bytes of every file the settled project's agent and skill live in.
fn owned_bytes(fx: &Fx) -> Vec<(PathBuf, Vec<u8>)> {
    fx.manifest()["owned_files"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|f| {
            f["section_id"]
                .as_str()
                .is_some_and(|s| s.starts_with("generation:"))
        })
        .map(|f| {
            let path = fx.root.join(f["path"].as_str().unwrap());
            let bytes = fs::read(&path).unwrap();
            (path, bytes)
        })
        .collect()
}

#[test]
fn a_record_that_is_only_partly_there_is_a_question_and_plans_no_removal() {
    let (fx, _) = settled_project();
    let full = fx.manifest()["generation_request"].clone();
    let before = owned_bytes(&fx);
    assert!(!before.is_empty());
    let mut extra_key = full.clone();
    extra_key["native_rolez"] = json!([]);
    let shapes = [
        ("an empty object", json!({})),
        (
            "only the skills",
            json!({"project_skills": full["project_skills"]}),
        ),
        (
            "one list missing",
            json!({"native_roles": full["native_roles"],
                   "project_skills": full["project_skills"]}),
        ),
        ("a key nothing writes", extra_key),
        (
            "all three lists empty",
            json!({"native_roles": [], "skill_projections": [], "project_skills": []}),
        ),
    ];
    for (name, record) in shapes {
        set_record(&fx, record);
        let plan = fx.plan();
        let generation = step(&plan, "generation");
        assert_eq!(
            generation["state"], "needs_decision",
            "{name}: {generation}"
        );
        assert_eq!(generation["apply"], Value::Null, "{name}: {generation}");
        assert!(
            !generation.to_string().contains("\"remove\""),
            "{name}: no removal is planned: {generation}"
        );
        for (path, bytes) in &before {
            assert_eq!(&fs::read(path).unwrap(), bytes, "{name}: {path:?} changed");
        }
    }
    // The whole record, put back, reads again.
    set_record(&fx, full);
    assert_eq!(step(&fx.plan(), "generation")["state"], "unchanged");
}

#[test]
fn a_record_of_the_right_shape_with_a_value_the_request_refuses_is_blocked() {
    let (fx, _) = settled_project();
    let mut record = fx.manifest()["generation_request"].clone();
    record["native_roles"][0]["role_id"] = json!("moai-audit");
    set_record(&fx, record);
    let plan = fx.plan();
    let generation = step(&plan, "generation");
    assert_eq!(generation["state"], "blocked", "{generation}");
    assert_eq!(generation["apply"], Value::Null, "{generation}");
}

#[test]
fn an_apply_with_every_file_unchanged_still_records_the_request_and_clears_the_decision() {
    let (fx, request) = settled_project();

    // A manifest written before the record existed: the agent and skill files are owned and
    // current, but no request is recorded, so the generation step asks the user to decide.
    let mut manifest = fx.manifest();
    manifest
        .as_object_mut()
        .unwrap()
        .remove("generation_request")
        .expect("the settled project recorded its request");
    let mut bytes = serde_json::to_vec_pretty(&manifest).unwrap();
    bytes.push(b'\n');
    fs::write(fx.manifest_path(), bytes).unwrap();
    let stuck = fx.plan();
    assert_eq!(
        step(&stuck, "generation")["state"],
        "needs_decision",
        "{stuck}"
    );

    // The files the request owns, read before the apply.
    let owned: Vec<(PathBuf, Vec<u8>)> = manifest["owned_files"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|f| {
            f["section_id"]
                .as_str()
                .is_some_and(|s| s.starts_with("generation:"))
        })
        .map(|f| {
            let path = fx.root.join(f["path"].as_str().unwrap());
            let bytes = fs::read(&path).unwrap();
            (path, bytes)
        })
        .collect();
    assert!(!owned.is_empty(), "{manifest}");

    // The same request again: every managed file is unchanged, and the apply still records.
    fx.generate(&request);
    let after = fx.manifest();
    assert_eq!(
        after["generation_request"]["native_roles"], request["native_roles"],
        "{after}"
    );
    assert_eq!(
        after["generation_request"]["project_skills"], request["project_skills"],
        "{after}"
    );
    for (path, before) in &owned {
        assert_eq!(&fs::read(path).unwrap(), before, "{path:?} changed");
    }
    let plan = fx.plan();
    assert_eq!(step(&plan, "generation")["state"], "unchanged", "{plan}");
}
