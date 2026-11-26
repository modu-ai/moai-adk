//! SPEC-20261125-generation-request-record: `generation-apply` records the project agent and
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

    // A request with nothing left drops the record. The files it no longer names are planned
    // `remove` (SPEC-20261125-generation-empty-request), so after the apply of that plan no
    // owned file is left for a missing record to leave undecided; the tests below pin that.
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
    // current, but no request is recorded. Every row is current, so the generation step reads
    // `unchanged` and names the files it keeps (SPEC-20261127-update-old-manifest-way REQ-3).
    let mut manifest = fx.manifest();
    manifest
        .as_object_mut()
        .unwrap()
        .remove("generation_request")
        .expect("the settled project recorded its request");
    let mut bytes = serde_json::to_vec_pretty(&manifest).unwrap();
    bytes.push(b'\n');
    fs::write(fx.manifest_path(), bytes).unwrap();
    let before_apply = fx.plan();
    assert_eq!(
        step(&before_apply, "generation")["state"],
        "unchanged",
        "{before_apply}"
    );
    assert!(
        step(&before_apply, "generation")["preview"]["not_rebuilt"]
            .as_array()
            .is_some_and(|p| !p.is_empty()),
        "{before_apply}"
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

// SPEC-20261125-generation-empty-request: a request that no longer names an owned project agent,
// skill or skill projection plans the removal of its file, so an empty request leaves no
// `needs_decision` behind.

const ROLE_FILE: &str = ".claude/agents/moai/prj-test-runner.md";
const SKILL_FILE: &str = ".claude/skills/moai-prj-checks/SKILL.md";
const PROJECTION_FILE: &str = ".claude/skills/moai-ref-b/SKILL.md";
const GENERATED_KINDS: [&str; 3] = [
    "generation:native-role:",
    "generation:project-skill:",
    "generation:skill-projection:",
];

/// `request` with no project agent and no project skill.
fn emptied(request: &Value) -> Value {
    let mut empty = request.clone();
    empty["native_roles"] = json!([]);
    empty["project_skills"] = json!([]);
    empty
}

fn plan_of(fx: &Fx, request: &Value) -> Value {
    fx.call(
        &["generation-plan", fx.p()],
        &serde_json::to_vec(request).unwrap(),
    )
}

/// The answer of `generation-apply` for `plan`, ok or not.
fn apply_plan(fx: &Fx, plan: &Value) -> Value {
    fx.run(
        &[
            "generation-apply",
            fx.p(),
            "--plan-digest",
            plan["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            REF,
        ],
        &serde_json::to_vec(plan).unwrap(),
    )
}

fn row<'a>(plan: &'a Value, path: &str) -> &'a Value {
    plan["files"]
        .as_array()
        .unwrap()
        .iter()
        .find(|f| f["path"] == path)
        .unwrap_or_else(|| panic!("no row for {path}: {plan}"))
}

/// The manifest's owned entries of the kinds a request builds, as `section_id`s.
fn owned_generated(fx: &Fx) -> Vec<String> {
    fx.manifest()["owned_files"]
        .as_array()
        .unwrap()
        .iter()
        .filter_map(|f| f["section_id"].as_str())
        .filter(|s| GENERATED_KINDS.iter().any(|k| s.starts_with(k)))
        .map(String::from)
        .collect()
}

#[test]
fn an_empty_request_previews_the_removal_of_the_agent_and_skill_files() {
    let (fx, request) = settled_project();
    let role_before = fs::read(fx.root.join(ROLE_FILE)).unwrap();
    let skill_before = fs::read(fx.root.join(SKILL_FILE)).unwrap();
    let manifest_before = fs::read(fx.manifest_path()).unwrap();

    let plan = plan_of(&fx, &emptied(&request));
    assert_eq!(row(&plan, ROLE_FILE)["action"], "remove", "{plan}");
    assert_eq!(row(&plan, SKILL_FILE)["action"], "remove", "{plan}");

    // A plan writes nothing: the files and the manifest keep their bytes until the apply.
    assert_eq!(fs::read(fx.root.join(ROLE_FILE)).unwrap(), role_before);
    assert_eq!(fs::read(fx.root.join(SKILL_FILE)).unwrap(), skill_before);
    assert_eq!(fs::read(fx.manifest_path()).unwrap(), manifest_before);
}

#[test]
fn an_empty_request_clears_the_decision_and_the_removal_is_recoverable() {
    let (fx, request) = settled_project();
    let role_before = fs::read(fx.root.join(ROLE_FILE)).unwrap();
    let skill_before = fs::read(fx.root.join(SKILL_FILE)).unwrap();

    let plan = plan_of(&fx, &emptied(&request));
    let applied = apply_plan(&fx, &plan);
    assert_eq!(applied["ok"], json!(true), "{applied}");
    assert!(!fx.root.join(ROLE_FILE).exists(), "the agent file is gone");
    assert!(!fx.root.join(SKILL_FILE).exists(), "the skill file is gone");
    assert_eq!(owned_generated(&fx), Vec::<String>::new());
    assert!(fx.manifest().get("generation_request").is_none());

    // The next preview has nothing left to decide.
    let next = fx.plan();
    let generation = step(&next, "generation");
    assert_ne!(generation["state"], "needs_decision", "{generation}");
    assert_eq!(generation["state"], "unchanged", "{generation}");

    // The bytes went to the receipt, so the removal can be recovered.
    let receipt = applied["result"]["receipt"]
        .as_str()
        .unwrap_or_else(|| panic!("the apply names its receipt: {applied}"));
    let recovered = fx.run(
        &[
            "apply-recover",
            fx.p(),
            "--receipt",
            receipt,
            "--permission-reference",
            REF,
        ],
        b"{}",
    );
    assert_eq!(recovered["ok"], json!(true), "{recovered}");
    assert_eq!(fs::read(fx.root.join(ROLE_FILE)).unwrap(), role_before);
    assert_eq!(fs::read(fx.root.join(SKILL_FILE)).unwrap(), skill_before);
}

#[test]
fn an_edited_or_misplaced_owned_agent_is_a_conflict_and_is_kept() {
    let (fx, request) = settled_project();
    let empty = emptied(&request);

    // An agent file edited after generation is kept, and the apply stops.
    let original = fs::read(fx.root.join(ROLE_FILE)).unwrap();
    let mut edited = original.clone();
    edited.extend_from_slice(b"\nA line the user added.\n");
    fs::write(fx.root.join(ROLE_FILE), &edited).unwrap();
    let plan = plan_of(&fx, &empty);
    let role = row(&plan, ROLE_FILE);
    assert_eq!(role["action"], "conflict", "{plan}");
    assert!(
        role["detail"].as_str().unwrap().contains("edited"),
        "{role}"
    );
    let refused = apply_plan(&fx, &plan);
    assert_eq!(refused["ok"], json!(false), "{refused}");
    assert_eq!(fs::read(fx.root.join(ROLE_FILE)).unwrap(), edited);
    assert!(fx.root.join(SKILL_FILE).exists(), "nothing was applied");
    fs::write(fx.root.join(ROLE_FILE), &original).unwrap();

    // An owned entry whose recorded path is not a path its kind names is never removed through.
    let other = ".claude/agents/moai/not-the-role-path.md";
    set_role_path(&fx, other);
    let plan = plan_of(&fx, &empty);
    assert_eq!(row(&plan, other)["action"], "conflict", "{plan}");
    assert!(fx.root.join(ROLE_FILE).exists(), "the real file is kept");

    // Under a request that still names the role, that entry is not judged at all.
    let named = plan_of(&fx, &request);
    assert!(
        named["files"]
            .as_array()
            .unwrap()
            .iter()
            .all(|f| f["path"] != other),
        "{named}"
    );
}

/// The manifest's native-role entry moved to `path`.
fn set_role_path(fx: &Fx, path: &str) {
    let mut manifest = fx.manifest();
    for entry in manifest["owned_files"].as_array_mut().unwrap() {
        if entry["section_id"]
            .as_str()
            .is_some_and(|s| s.starts_with("generation:native-role:"))
        {
            entry["path"] = json!(path);
        }
    }
    let mut bytes = serde_json::to_vec_pretty(&manifest).unwrap();
    bytes.push(b'\n');
    fs::write(fx.manifest_path(), bytes).unwrap();
}

#[test]
fn an_unrequested_role_at_its_earlier_flat_path_is_removed() {
    let (fx, request) = settled_project();
    // An earlier version wrote the role flat; its manifest entry names that path.
    let flat = ".claude/agents/prj-test-runner.md";
    fs::rename(fx.root.join(ROLE_FILE), fx.root.join(flat)).unwrap();
    set_role_path(&fx, flat);
    let plan = plan_of(&fx, &emptied(&request));
    assert_eq!(row(&plan, flat)["action"], "remove", "{plan}");
    assert!(fx.root.join(flat).exists(), "a plan removes nothing");
}

#[test]
fn an_owned_agent_behind_a_link_is_a_conflict_and_nothing_outside_is_removed() {
    let (fx, request) = settled_project();
    let outside = TempDir::new().unwrap();
    let moved = outside.path().join("moai");
    fs::rename(fx.root.join(".claude/agents/moai"), &moved).unwrap();
    std::os::unix::fs::symlink(&moved, fx.root.join(".claude/agents/moai")).unwrap();
    let plan = plan_of(&fx, &emptied(&request));
    let role = row(&plan, ROLE_FILE);
    assert_eq!(role["action"], "conflict", "{plan}");
    assert!(
        role["detail"].as_str().unwrap().contains("symbolic link"),
        "{role}"
    );
    let refused = apply_plan(&fx, &plan);
    assert_eq!(refused["ok"], json!(false), "{refused}");
    assert!(
        moved.join("prj-test-runner.md").exists(),
        "nothing outside the project was removed"
    );
}

/// A project whose one skill projection `ref-b` is generated from a license-reviewed source,
/// and the request that built it.
fn projection_project() -> (Fx, Value) {
    let fx = Fx::new();
    fx.with_profile(json!({}));
    let sources = json!({"schema_version": 1, "sources": [
        {"source_url": "fixtures/skill-b", "license": {"status": "reviewed", "spdx": "MIT"},
         "resolved_revision": null}]});
    let at = sources_path(&fx);
    fs::create_dir_all(at.parent().unwrap()).unwrap();
    fs::write(at, serde_json::to_vec_pretty(&sources).unwrap()).unwrap();
    let mut request = request_with(&fx, Vec::new(), Vec::new());
    request["skill_projections"] = json!([
        {"id": "ref-b", "source_url": "fixtures/skill-b", "host_targets": ["claude"]}
    ]);
    fx.generate(&request);
    assert!(fx.root.join(PROJECTION_FILE).exists());
    assert_eq!(
        owned_generated(&fx),
        vec!["generation:skill-projection:claude:ref-b".to_string()]
    );
    (fx, request)
}

fn sources_path(fx: &Fx) -> PathBuf {
    moai_cli::state_root::path(&fx.root, "manifests/sources.json")
}

#[test]
fn a_projection_the_request_names_but_cannot_plan_is_not_removed() {
    let (fx, request) = projection_project();
    // The source is no longer license-reviewed, so the projection is not planned (a warning);
    // the request still names it, so its file stays owned and stays.
    fs::remove_file(sources_path(&fx)).unwrap();
    let plan = plan_of(&fx, &request);
    assert!(
        plan["warnings"]
            .to_string()
            .contains("not license-reviewed"),
        "{plan}"
    );
    let removals: Vec<&Value> = plan["files"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|f| f["action"] == "remove" || f["path"] == PROJECTION_FILE)
        .collect();
    assert!(removals.is_empty(), "{removals:?}");
    assert!(fx.root.join(PROJECTION_FILE).exists());
}

#[test]
fn an_unrequested_skill_projection_is_removed_like_a_project_skill() {
    let (fx, request) = projection_project();

    let mut without = request.clone();
    without["skill_projections"] = json!([]);
    let plan = plan_of(&fx, &without);
    assert_eq!(row(&plan, PROJECTION_FILE)["action"], "remove", "{plan}");
    assert!(
        fx.root.join(PROJECTION_FILE).exists(),
        "a plan removes nothing"
    );
    let applied = apply_plan(&fx, &plan);
    assert_eq!(applied["ok"], json!(true), "{applied}");
    assert!(!fx.root.join(PROJECTION_FILE).exists());
    assert_eq!(owned_generated(&fx), Vec::<String>::new());
}

#[test]
fn a_harness_only_request_plans_no_removal() {
    let (fx, request) = settled_project();
    let mut harness = emptied(&request);
    harness["picks"] = json!(["harness-only"]);
    let plan = plan_of(&fx, &harness);
    let removals: Vec<&Value> = plan["files"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|f| f["action"] == "remove")
        .collect();
    assert!(removals.is_empty(), "{removals:?}");
    assert!(fx.root.join(ROLE_FILE).exists() && fx.root.join(SKILL_FILE).exists());
}

/// A request for the harness alone that names a project agent records it like any other apply,
/// so the next plan reads `unchanged` instead of asking to rebuild it every time.
#[test]
fn a_harness_only_apply_records_its_agents_and_the_next_plan_reads_unchanged() {
    let fx = Fx::new();
    fx.with_profile(json!({}));
    // The harness is in place first; the agents are then made on their own.
    let base = request_with(&fx, Vec::new(), Vec::new());
    fx.generate(&base);
    let mut request = base.clone();
    request["native_roles"] = json!([role()]);
    request["picks"] = json!(["harness-only"]);
    fx.generate(&request);

    let record = &fx.manifest()["generation_request"];
    assert_eq!(record["native_roles"], request["native_roles"], "{record}");
    for s in steps(&fx.plan())
        .into_iter()
        .filter(|s| s["state"] == "ready")
    {
        assert_eq!(fx.apply_step(s)["ok"], json!(true), "{s}");
    }
    let plan = fx.plan();
    let generation = step(&plan, "generation");
    assert_eq!(generation["state"], "unchanged", "{generation}");
    assert_eq!(generation["apply"], Value::Null, "{generation}");
}

/// An agent made by a harness-only apply is recorded, so a later profile change that reaches it
/// plans the rebuild (a `ready` step that prints its command) and applying it settles the step.
#[test]
fn a_profile_change_after_a_harness_only_agent_plans_its_rebuild() {
    let fx = Fx::new();
    fx.with_profile(json!({}));
    let base = request_with(&fx, Vec::new(), Vec::new());
    fx.generate(&base);
    let mut request = base.clone();
    request["native_roles"] = json!([role()]);
    request["picks"] = json!(["harness-only"]);
    fx.generate(&request);
    for s in steps(&fx.plan())
        .into_iter()
        .filter(|s| s["state"] == "ready")
    {
        assert_eq!(fx.apply_step(s)["ok"], json!(true), "{s}");
    }
    let settled = fx.plan();
    assert_eq!(
        step(&settled, "generation")["state"],
        "unchanged",
        "{settled}"
    );

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
    let agent_row = generation["preview"]["files"]
        .as_array()
        .unwrap()
        .iter()
        .find(|f| f["path"].as_str().unwrap().contains("prj-test-runner"))
        .unwrap_or_else(|| panic!("no prj-test-runner row: {generation}"));
    assert_ne!(agent_row["action"], "remove", "{agent_row}");
    assert_eq!(fx.apply_step(generation)["ok"], json!(true));
    let again = fx.plan();
    assert_eq!(step(&again, "generation")["state"], "unchanged", "{again}");
}

/// A request for the harness alone that names no agent says nothing about the ones an earlier
/// apply recorded: the record stays as it was.
#[test]
fn a_harness_only_apply_without_agents_keeps_the_earlier_record() {
    let (fx, request) = settled_project();
    let before = fx.manifest()["generation_request"].clone();
    assert!(before.is_object(), "{before}");
    let mut harness = emptied(&request);
    harness["picks"] = json!(["harness-only"]);
    let applied = apply_plan(&fx, &plan_of(&fx, &harness));
    assert_eq!(applied["ok"], json!(true), "{applied}");
    assert_eq!(fx.manifest()["generation_request"], before);
}

/// The text `moai <command> --help` prints.
fn help_text(fx: &Fx, command: &str) -> String {
    let out = Command::new(EXE)
        .args([command, "--help"])
        .clean_env()
        .env("HOME", &fx.home)
        .env("PATH", "/usr/bin:/bin")
        .env("LC_ALL", "C")
        .stdin(Stdio::null())
        .output()
        .unwrap();
    format!(
        "{}{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    )
}

#[test]
fn the_help_examples_name_the_input_each_command_reads() {
    let fx = Fx::new();
    for command in ["generation-plan", "generation-apply"] {
        let help = help_text(&fx, command);
        let examples: Vec<&str> = help
            .lines()
            .filter(|l| {
                (l.contains("generation-apply ") || l.contains("generation-plan "))
                    && l.contains(" < ")
            })
            .collect();
        assert!(!examples.is_empty(), "{command} prints examples: {help}");
        for line in examples {
            // The plan reads a request; the apply reads the plan the plan answered.
            let wants = if line.contains("generation-apply ") {
                "< plan.json"
            } else {
                "< request.json"
            };
            assert!(
                line.contains(wants),
                "{command}: `{}` should read `{wants}`",
                line.trim()
            );
        }
    }
}

/// The first JSON object that starts a request (`{"protocol_version"`) in `text`.
fn request_example(text: &str) -> Value {
    let key = text
        .find("\"protocol_version\"")
        .unwrap_or_else(|| panic!("no request example in: {text}"));
    let at = text[..key]
        .rfind('{')
        .expect("the object opens before its first key");
    serde_json::Deserializer::from_str(&text[at..])
        .into_iter::<Value>()
        .next()
        .unwrap()
        .unwrap_or_else(|e| panic!("the example is not JSON ({e}): {}", &text[at..]))
}

/// The complete request example of `generation-plan --help` and of the project workflow's step 6
/// is one and the same object, `generation-plan` accepts it, and the plan it answers is the
/// input `generation-apply` accepts: the documented path works as written.
#[test]
fn the_documented_request_example_is_accepted_and_its_plan_applies() {
    let fx = Fx::new();
    fx.with_profile(json!({}));
    let from_help = request_example(&help_text(&fx, "generation-plan"));
    let workflow = fs::read_to_string(
        std::path::Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai/workflows/project.md"),
    )
    .unwrap();
    assert_eq!(
        from_help,
        request_example(&workflow),
        "help and workflow differ"
    );

    let plan = fx.call(
        &["generation-plan", fx.p()],
        &serde_json::to_vec(&from_help).unwrap(),
    );
    let files = plan["files"].to_string();
    assert!(
        files.contains("prj-test-runner") && files.contains("prj-checks"),
        "the plan covers the documented agent and skill: {files}"
    );
    let applied = apply_plan(&fx, &plan);
    assert_eq!(applied["ok"], json!(true), "{applied}");
    assert!(
        fx.root.join(ROLE_FILE).exists() && fx.root.join(SKILL_FILE).exists(),
        "the documented agent and skill were written"
    );
}

/// A manifest entry of `kind` (its bare kind and its `section_id`) naming `path`, whose recorded
/// baseline is the digest of `bytes`, so the file there reads as never edited.
fn plant_entry(fx: &Fx, section: &str, path: &str, bytes: &[u8]) {
    let bare = section
        .strip_prefix("generation:")
        .and_then(|rest| rest.split(':').next())
        .unwrap();
    let mut manifest = fx.manifest();
    manifest["owned_files"].as_array_mut().unwrap().push(json!({
        "path": path, "kind": bare, "section_id": section,
        "baseline_sha256": moai_cli::digest(bytes),
    }));
    let mut out = serde_json::to_vec_pretty(&manifest).unwrap();
    out.push(b'\n');
    fs::write(fx.manifest_path(), out).unwrap();
}

#[test]
fn an_owned_entry_whose_id_leaves_its_folder_is_a_conflict_and_the_file_stays() {
    let (fx, request) = settled_project();
    let outside_dir = fx.root.parent().unwrap().join("outside");
    fs::create_dir_all(&outside_dir).unwrap();
    // The three shapes of the audit: a file of the project, a file outside it, and the file
    // named SKILL.md that a project-skill entry reaches. Kind and path are planted together.
    let cases = [
        (
            "generation:native-role:claude:../../../README",
            ".claude/agents/moai/../../../README.md",
            fx.root.join("README.md"),
        ),
        (
            "generation:native-role:claude:../../../../outside/secret",
            ".claude/agents/moai/../../../../outside/secret.md",
            outside_dir.join("secret.md"),
        ),
        (
            "generation:project-skill:claude:../..",
            ".claude/skills/../../SKILL.md",
            fx.root.join("SKILL.md"),
        ),
        // A separator in the id, with a path that stays inside the folder.
        (
            "generation:native-role:claude:sub/dir",
            ".claude/agents/moai/sub/dir.md",
            fx.root.join(".claude/agents/moai/sub/dir.md"),
        ),
    ];
    for (section, path, target) in &cases {
        fs::create_dir_all(target.parent().unwrap()).unwrap();
        fs::write(target, b"keep me\n").unwrap();
        plant_entry(&fx, section, path, b"keep me\n");
    }
    let plan = plan_of(&fx, &emptied(&request));
    for (section, path, _) in &cases {
        let planned = row(&plan, path);
        assert_eq!(planned["action"], "conflict", "{section}: {planned}");
    }
    // The apply of that plan stops, and then so does the plan after it: nothing is deleted.
    let refused = apply_plan(&fx, &plan);
    assert_eq!(refused["ok"], json!(false), "{refused}");
    for (section, _, target) in &cases {
        assert_eq!(fs::read(target).unwrap(), b"keep me\n", "{section}");
    }
    // The update preview does not offer the removal either.
    let update = fx.plan();
    let generation = step(&update, "generation");
    assert!(
        !generation.to_string().contains("\"remove\""),
        "the update preview offers no removal: {generation}"
    );
}

/// `moai cc` in `root`, with a host stub on PATH that exits at once; the combined output.
fn first_launch(fx: &Fx) -> String {
    use std::os::unix::fs::PermissionsExt;
    let bin = fx.root.parent().unwrap().join("bin");
    fs::create_dir_all(&bin).unwrap();
    let host = bin.join("claude");
    fs::write(&host, "#!/bin/sh\nexit 0\n").unwrap();
    fs::set_permissions(&host, fs::Permissions::from_mode(0o755)).unwrap();
    let out = Command::new(EXE)
        .args(["cc", "--name", "lane-1"])
        .current_dir(&fx.root)
        .env("PATH", format!("{}:/usr/bin:/bin", bin.display()))
        .env("HOME", &fx.home)
        .env("USERPROFILE", &fx.home)
        .env("MOAI_TEST_HOST_MARKER", fx.home.join("host-started"))
        .env("ANTHROPIC_AUTH_TOKEN", "moai-test-dummy-not-real")
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .stdin(Stdio::null())
        .output()
        .unwrap();
    format!(
        "{}{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    )
}

#[test]
fn a_first_launch_never_applies_the_removal_of_an_owned_file() {
    // A project set up by the launcher (its project id is the one a first launch derives from
    // the folder name) and generated with the pick `harness-only`, so the manifest owns an agent
    // and a skill but no instruction file; its settings are then lost. The first launch plans
    // its empty request over those owned files.
    let fx = Fx::new();
    let named = moai_cli::key_vault::folder_name(None, "project", "");
    fx.with_profile(json!({"project_id": named.strip_suffix('-').unwrap_or(&named)}));
    let mut request = request_with(&fx, vec![role()], vec![skill()]);
    request["picks"] = json!(["harness-only"]);
    fx.generate(&request);
    for state in ["project.yaml", "manifests/profile-baseline.json"] {
        fs::remove_file(fx.root.join(moai_cli::state_root::rel(&fx.root, state))).unwrap();
    }
    for name in ["CLAUDE.md", "AGENTS.md"] {
        let _ = fs::remove_file(fx.root.join(name));
    }
    let role = fs::read(fx.root.join(ROLE_FILE)).unwrap();
    let skill = fs::read(fx.root.join(SKILL_FILE)).unwrap();
    let said = first_launch(&fx);
    assert!(
        said.contains("not generated"),
        "the launch says it left CLAUDE.md out: {said}"
    );
    assert_eq!(fs::read(fx.root.join(ROLE_FILE)).unwrap(), role, "{said}");
    assert_eq!(fs::read(fx.root.join(SKILL_FILE)).unwrap(), skill, "{said}");
}
