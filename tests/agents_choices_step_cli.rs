//! SPEC-20261119-agents-choices-step (card C-57, slice 2b): the `generation` step of
//! `project-update-plan` offers the append, skip and CLAUDE.md import choices of slice 2a, and
//! `generation-plan`/`generation-apply` carry the pick in the request's `picks`.
//!
//! Every case runs the built binary in a temporary project with an empty temporary HOME; the
//! real checkout and HOME are never read.

#![cfg(unix)]

mod common;
use common::CleanEnv;
use serde_json::{json, Value};
use std::{
    fs,
    io::Write,
    path::{Path, PathBuf},
    process::{Command, Stdio},
};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;
const REF: &str = "test:c57-2b";
const PLACEHOLDER: &str = "<permission-reference>";
const HUMAN: &str = "# mine\n\nHand-written rules.\nNo final newline";

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

    fn write(&self, rel: &str, body: &str) {
        let p = self.root.join(rel);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        fs::write(p, body).unwrap();
    }

    fn read(&self, rel: &str) -> String {
        fs::read_to_string(self.root.join(rel)).unwrap()
    }

    /// The whole envelope of one run with `stdin`; the child's HOME is this fixture's.
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

    /// A profile for a project that holds the given hosts.
    fn with_profile(&self, hosts: &[&str]) {
        self.with_language(hosts, None);
    }

    /// [`Fx::with_profile`] with the language of the instruction files set, so a later profile
    /// in another language renders AGENTS.md differently from what the first apply recorded.
    fn with_language(&self, hosts: &[&str], instruction_files: Option<&str>) {
        let detect = self.call(&["project-detect", self.p()], b"");
        let mut profile = json!({
            "schema_version": 2, "project_id": "demo", "purpose": "Build a small editor",
            "tools": ["Rust"], "development_methods": ["TDD"], "hosts": hosts,
            "detected": detect["detected"],
        });
        if let Some(code) = instruction_files {
            profile["language"] = json!({
                "agent_messages": "en", "code_comments": "en", "commit_messages": "en",
                "instruction_files": code, "issue_pr_records": "en", "reports": "en",
                "skill_instructions": "en", "user_chat": "en"
            });
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

    fn update_plan(&self) -> Value {
        self.call(&["project-update-plan", self.p()], b"")
    }

    /// A project whose generation was applied once in English, whose AGENTS.md was then replaced
    /// by a hand-written text without managed markers, and whose profile then asks for Korean
    /// instruction files, so the rendered block differs from the recorded baseline and the
    /// plan conflicts (the same way `agents_conflict_choices_cli.rs` makes it). Returns the
    /// request the update preview builds now: the first apply's, in Korean.
    fn conflict(hosts: &[&str]) -> (Fx, Value) {
        let fx = Fx::new();
        fx.with_language(hosts, Some("en"));
        let plan = fx.update_plan();
        let generation = step(&plan, "generation");
        assert_eq!(generation["state"], "ready", "{generation}");
        let mut request = generation["apply"]["stdin"]["request"].clone();
        request["language"] = json!("ko");
        assert_eq!(fx.apply(generation)["ok"], json!(true));
        fx.write("AGENTS.md", HUMAN);
        fx.with_language(hosts, Some("ko"));
        (fx, request)
    }

    /// A project whose generation was applied through the step's own command.
    fn settled(hosts: &[&str]) -> Fx {
        Fx::settled_with_request(hosts).0
    }

    /// [`Fx::settled`] with the request of that apply (a settled step prints none any more).
    fn settled_with_request(hosts: &[&str]) -> (Fx, Value) {
        let fx = Fx::new();
        fx.with_profile(hosts);
        let plan = fx.update_plan();
        let generation = step(&plan, "generation");
        assert_eq!(generation["state"], "ready", "{generation}");
        let request = generation["apply"]["stdin"]["request"].clone();
        assert_eq!(fx.apply(generation)["ok"], json!(true));
        (fx, request)
    }

    /// Runs the printed command of a step or of one choice, the way a session does.
    fn apply(&self, with_apply: &Value) -> Value {
        let argv: Vec<String> = with_apply["apply"]["argv"]
            .as_array()
            .unwrap_or_else(|| panic!("no apply command: {with_apply}"))
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
        self.run(
            &args,
            &serde_json::to_vec(&with_apply["apply"]["stdin"]).unwrap(),
        )
    }
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

fn with_picks(request: &Value, picks: &[&str]) -> Value {
    let mut r = request.clone();
    r["picks"] = json!(picks);
    r
}

fn row<'a>(plan: &'a Value, path: &str) -> Option<&'a Value> {
    plan["files"]
        .as_array()
        .unwrap()
        .iter()
        .find(|f| f["path"] == path)
}

fn action(plan: &Value, path: &str) -> String {
    row(plan, path).unwrap_or_else(|| panic!("no row {path}: {plan}"))["action"]
        .as_str()
        .unwrap()
        .to_string()
}

fn choice_ids(generation: &Value) -> Vec<String> {
    generation["preview"]["choices"]
        .as_array()
        .unwrap_or_else(|| panic!("no choices: {generation}"))
        .iter()
        .map(|c| c["id"].as_str().unwrap().to_string())
        .collect()
}

fn choice<'a>(generation: &'a Value, id: &str) -> &'a Value {
    generation["preview"]["choices"]
        .as_array()
        .unwrap()
        .iter()
        .find(|c| c["id"] == id)
        .unwrap_or_else(|| panic!("no choice {id}: {generation}"))
}

fn plan_digest(fx: &Fx, request: &Value) -> (Value, String) {
    let plan = fx.call(
        &["generation-plan", fx.p()],
        &serde_json::to_vec(request).unwrap(),
    );
    let digest = plan["plan_digest"].as_str().unwrap().to_string();
    (plan, digest)
}

// REQ-1, REQ-2: the request carries the picks, and a request without picks is as before.
#[test]
fn generation_picks_absent_or_empty_leave_the_request_json_and_the_digest_as_before() {
    let (fx, request) = Fx::conflict(&["claude", "codex"]);
    assert!(request.get("picks").is_none(), "{request}");
    let (plain, plain_digest) = plan_digest(&fx, &request);
    let (empty, empty_digest) = plan_digest(&fx, &with_picks(&request, &[]));
    assert_eq!(plain_digest, empty_digest);
    assert!(plain["request"].get("picks").is_none(), "{plain}");
    assert!(empty["request"].get("picks").is_none(), "{empty}");
    assert_eq!(action(&plain, "AGENTS.md"), "conflict", "{plain}");
}

#[test]
fn generation_plan_with_the_append_pick_plans_agents_md_as_append_and_writes_nothing() {
    let (fx, request) = Fx::conflict(&["claude", "codex"]);
    let (plain, plain_digest) = plan_digest(&fx, &request);
    let (picked, picked_digest) = plan_digest(&fx, &with_picks(&request, &["append-managed"]));
    assert_eq!(action(&picked, "AGENTS.md"), "append", "{picked}");
    assert_eq!(picked["request"]["picks"], json!(["append-managed"]));
    assert_ne!(plain_digest, picked_digest);
    assert_eq!(action(&plain, "AGENTS.md"), "conflict");
    assert_eq!(fx.read("AGENTS.md"), HUMAN, "a plan writes nothing");
}

// REQ-3: the apply re-plans and applies with the picks of the plan it is given.
#[test]
fn generation_apply_applies_the_pick_the_plan_carries() {
    let (fx, request) = Fx::conflict(&["claude", "codex"]);
    let (plan, digest) = plan_digest(&fx, &with_picks(&request, &["append-managed"]));
    let out = fx.run(
        &[
            "generation-apply",
            fx.p(),
            "--plan-digest",
            &digest,
            "--permission-reference",
            REF,
        ],
        &serde_json::to_vec(&plan).unwrap(),
    );
    assert_eq!(out["ok"], json!(true), "{out}");
    let text = fx.read("AGENTS.md");
    assert!(text.starts_with(HUMAN), "{text:?}");
    assert_eq!(text.matches("moai:managed:begin").count(), 1, "{text:?}");
}

// REQ-4: an unknown id is refused before anything is written.
#[test]
fn generation_an_unknown_pick_is_refused_as_invalid_input_and_writes_nothing() {
    let (fx, request) = Fx::conflict(&["claude", "codex"]);
    let before = fx.read("AGENTS.md");
    let out = fx.run(
        &["generation-plan", fx.p()],
        &serde_json::to_vec(&with_picks(&request, &["no-such-pick"])).unwrap(),
    );
    assert_eq!(out["ok"], json!(false), "{out}");
    assert_eq!(out["error"]["kind"], "invalid_input", "{out}");
    assert_eq!(fx.read("AGENTS.md"), before);
}

// REQ-5: the shadow rides the plan and stays out of the digest.
#[test]
fn generation_plan_carries_the_claude_shadow_outside_the_digest() {
    // A CLAUDE.md is already named in `host_notes` (and so in the digest) whatever it holds;
    // what the shadow adds is only whether it imports AGENTS.md, and that must not move it.
    let (fx, request) = Fx::settled_with_request(&["claude", "codex"]);
    fx.write("CLAUDE.md", "@AGENTS.md\n# mine\n");
    let (without, without_digest) = plan_digest(&fx, &request);
    assert!(without.get("claude_shadow").is_none(), "{without}");
    fx.write("CLAUDE.md", "# mine\n");
    let (with, with_digest) = plan_digest(&fx, &request);
    assert_eq!(with["claude_shadow"]["path"], json!("CLAUDE.md"), "{with}");
    assert_eq!(without_digest, with_digest);
}

// REQ-6, REQ-7, REQ-10: the conflict step lists the choices and each command works.
#[test]
fn choices_a_missing_block_conflict_lists_append_then_skip_and_keeps_its_state() {
    let (fx, _) = Fx::conflict(&["claude", "codex"]);
    let plan = fx.update_plan();
    let generation = step(&plan, "generation");
    assert_eq!(generation["state"], "conflict", "{generation}");
    assert!(generation["apply"].is_null(), "{generation}");
    assert_eq!(choice_ids(generation), ["append-managed", "skip-conflicts"]);
    assert_eq!(
        generation["preview"]["choices"][0]["recommended"],
        json!(true)
    );
    assert_ne!(
        generation["preview"]["choices"][1]["recommended"],
        json!(true)
    );
    for c in generation["preview"]["choices"].as_array().unwrap() {
        let picks = c["picks"].clone();
        assert!(picks.as_array().is_some_and(|p| !p.is_empty()), "{c}");
        let digest = c["plan_digest"].as_str().unwrap();
        assert!(digest.starts_with("sha256:"), "{c}");
        let argv = c["apply"]["argv"].as_array().unwrap();
        assert_eq!(argv[1], "generation-apply", "{c}");
        assert!(argv.iter().any(|w| w == digest), "{c}");
        assert_eq!(c["apply"]["stdin"]["request"]["picks"], picks, "{c}");
        assert_eq!(c["apply"]["stdin"]["plan_digest"], json!(digest), "{c}");
    }
    assert_eq!(fx.read("AGENTS.md"), HUMAN, "a preview writes nothing");
}

#[test]
fn choices_the_append_command_keeps_every_byte_and_settles_the_next_plan() {
    let (fx, _) = Fx::conflict(&["claude", "codex"]);
    let generation = step(&fx.update_plan(), "generation").clone();
    let out = fx.apply(choice(&generation, "append-managed"));
    assert_eq!(out["ok"], json!(true), "{out}");
    let text = fx.read("AGENTS.md");
    assert!(text.starts_with(HUMAN), "{text:?}");
    assert_eq!(text.matches("moai:managed:begin").count(), 1, "{text:?}");
    let after = fx.update_plan();
    let after = step(&after, "generation");
    assert_eq!(after["state"], "unchanged", "{after}");
    assert!(after["preview"].get("choices").is_none(), "{after}");
}

#[test]
fn choices_the_skip_command_keeps_agents_md_and_applies_the_rest() {
    let (fx, _) = Fx::conflict(&["claude", "codex"]);
    let generation = step(&fx.update_plan(), "generation").clone();
    // The Korean profile changes the workflow file the conflict does not touch.
    let workflow = generation["preview"]["files"]
        .as_array()
        .unwrap()
        .iter()
        .find(|f| f["path"].as_str().unwrap().ends_with("/SKILL.md") && f["action"] == "merge")
        .unwrap_or_else(|| panic!("no workflow row: {generation}"))["path"]
        .as_str()
        .unwrap()
        .to_string();
    let before = fx.read(&workflow);
    let out = fx.apply(choice(&generation, "skip-conflicts"));
    assert_eq!(out["ok"], json!(true), "{out}");
    assert_eq!(fx.read("AGENTS.md"), HUMAN);
    assert_ne!(
        fx.read(&workflow),
        before,
        "the other files are still written"
    );
    // The skip records no baseline: the plain plan conflicts again and offers the choices again.
    let again = fx.update_plan();
    let again = step(&again, "generation");
    assert_eq!(again["state"], "conflict", "{again}");
    assert_eq!(choice_ids(again), ["append-managed", "skip-conflicts"]);
}

// REQ-6: only a missing block is offered the append and the skip; an edited block is not.
#[test]
fn choices_an_edited_block_conflict_offers_no_append_or_skip() {
    let fx = Fx::new();
    fx.with_language(&["claude", "codex"], Some("en"));
    let plan = fx.update_plan();
    assert_eq!(fx.apply(step(&plan, "generation"))["ok"], json!(true));
    // A line typed inside the managed block, then a profile that renders it differently.
    let text = fx.read("AGENTS.md");
    let end = "<!-- moai:managed:end";
    assert_eq!(text.matches(end).count(), 1, "{text}");
    fx.write(
        "AGENTS.md",
        &text.replacen(end, &format!("hand edit\n{end}"), 1),
    );
    fx.with_language(&["claude", "codex"], Some("ko"));
    let plan = fx.update_plan();
    let generation = step(&plan, "generation");
    assert_eq!(generation["state"], "conflict", "{generation}");
    let detail = generation["preview"]["files"]
        .as_array()
        .unwrap()
        .iter()
        .find(|f| f["path"] == "AGENTS.md")
        .unwrap()["detail"]
        .as_str()
        .unwrap_or_default()
        .to_string();
    assert!(!detail.contains("missing"), "{detail}");
    assert!(
        generation["preview"].get("choices").is_none(),
        "{generation}"
    );
}

// REQ-12: a choice that could not be applied is not offered.
#[test]
fn choices_a_choice_whose_own_plan_still_conflicts_is_left_out() {
    let (fx, _) = Fx::conflict(&["claude", "codex"]);
    // A hand edit of a generated workflow file: its own conflict stays whatever is picked for
    // AGENTS.md, so no command printed here could be applied.
    let workflow = ".moai/workflows/project/SKILL.md";
    let edited = format!("{}\nhand edit\n", fx.read(workflow));
    fx.write(workflow, &edited);
    let plan = fx.update_plan();
    let generation = step(&plan, "generation");
    assert_eq!(generation["state"], "conflict", "{generation}");
    let conflicts: Vec<&str> = generation["preview"]["files"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|f| f["action"] == "conflict")
        .map(|f| f["path"].as_str().unwrap())
        .collect();
    assert_eq!(conflicts, ["AGENTS.md", workflow], "{generation}");
    assert!(
        generation["preview"].get("choices").is_none(),
        "{generation}"
    );
    assert_eq!(fx.read(workflow), edited, "a preview writes nothing");
}

// REQ-9: nothing to choose, no key.
#[test]
fn choices_are_absent_without_a_conflict_and_without_a_shadow() {
    let fresh = Fx::new();
    fresh.with_profile(&["claude", "codex"]);
    let plan = fresh.update_plan();
    let generation = step(&plan, "generation");
    assert_eq!(generation["state"], "ready", "{generation}");
    assert!(
        generation["preview"].get("choices").is_none(),
        "{generation}"
    );

    let settled = Fx::settled(&["claude", "codex"]);
    let plan = settled.update_plan();
    let generation = step(&plan, "generation");
    assert_eq!(generation["state"], "unchanged", "{generation}");
    assert!(
        generation["preview"].get("choices").is_none(),
        "{generation}"
    );
}

// REQ-8: the CLAUDE.md import.
#[test]
fn claude_the_import_choice_stands_alone_while_agents_md_has_no_conflict() {
    let fx = Fx::settled(&["claude", "codex"]);
    fx.write("CLAUDE.md", "# mine\n");
    let plan = fx.update_plan();
    let generation = step(&plan, "generation");
    assert_eq!(generation["state"], "unchanged", "{generation}");
    assert_eq!(choice_ids(generation), ["claude-import"]);
    assert_eq!(
        generation["preview"]["claude_shadow"]["path"],
        json!("CLAUDE.md")
    );
    let out = fx.apply(choice(generation, "claude-import"));
    assert_eq!(out["ok"], json!(true), "{out}");
    assert_eq!(fx.read("CLAUDE.md"), "# mine\n@AGENTS.md\n");
    // Imported: the next preview has nothing to offer.
    let after = fx.update_plan();
    let after = step(&after, "generation");
    assert!(after["preview"].get("choices").is_none(), "{after}");
}

#[test]
fn claude_the_import_is_combined_with_the_append_while_agents_md_conflicts() {
    let (fx, _) = Fx::conflict(&["claude", "codex"]);
    fx.write("CLAUDE.md", "# mine\n");
    let generation = step(&fx.update_plan(), "generation").clone();
    assert_eq!(
        choice_ids(&generation),
        [
            "append-managed",
            "skip-conflicts",
            "append-managed+claude-import"
        ]
    );
    let combined = choice(&generation, "append-managed+claude-import");
    assert_eq!(
        combined["picks"],
        json!(["append-managed", "claude-import"])
    );
    let out = fx.apply(combined);
    assert_eq!(out["ok"], json!(true), "{out}");
    assert!(fx.read("AGENTS.md").starts_with(HUMAN));
    assert_eq!(fx.read("CLAUDE.md"), "# mine\n@AGENTS.md\n");
}

// REQ-11: the schemas say what the code accepts and prints.
fn schema(name: &str) -> Value {
    let path = Path::new(env!("CARGO_MANIFEST_DIR"))
        .join("schemas/protocol/v1")
        .join(name);
    serde_json::from_slice(&fs::read(path).unwrap()).unwrap()
}

#[test]
fn schema_describes_picks_claude_shadow_and_the_skip_action() {
    let request = schema("agents-request.schema.json");
    let ids: Vec<&str> = request["properties"]["picks"]["items"]["enum"]
        .as_array()
        .unwrap_or_else(|| panic!("no picks enum: {request}"))
        .iter()
        .map(|v| v.as_str().unwrap())
        .collect();
    assert_eq!(ids, ["append-managed", "skip-conflicts", "claude-import"]);
    let plan = schema("agents-plan.schema.json");
    assert!(
        plan["properties"].get("claude_shadow").is_some(),
        "claude_shadow is not described: {plan}"
    );
    let actions = plan["$defs"]["file_plan"]["properties"]["action"]["enum"]
        .as_array()
        .unwrap();
    assert!(actions.iter().any(|a| a == "skip"), "{actions:?}");
}

#[test]
fn schema_lists_every_key_of_a_real_plan_made_with_picks_and_a_shadow() {
    let (fx, request) = Fx::conflict(&["claude", "codex"]);
    fx.write("CLAUDE.md", "# mine\n");
    fs::create_dir_all(
        fx.root
            .join(request["workflow_reference"].as_str().unwrap()),
    )
    .unwrap();
    let agents_request = json!({
        "protocol_version": 1, "language": request["language"],
        "workflow_reference": request["workflow_reference"], "local_file": request["local_file"],
        "picks": ["skip-conflicts", "claude-import"],
    });
    let plan = fx.call(
        &["agents-plan", fx.p()],
        &serde_json::to_vec(&agents_request).unwrap(),
    );
    let keys_in = |value: &Value, schema: &Value, what: &str| {
        for key in value.as_object().unwrap().keys() {
            assert!(
                schema["properties"].get(key).is_some(),
                "{what}: `{key}` is not in the schema"
            );
        }
    };
    let plan_schema = schema("agents-plan.schema.json");
    keys_in(&plan, &plan_schema, "plan");
    keys_in(
        &plan["request"],
        &schema("agents-request.schema.json"),
        "request",
    );
    let file_schema = &plan_schema["$defs"]["file_plan"];
    let actions = file_schema["properties"]["action"]["enum"]
        .as_array()
        .unwrap();
    for f in plan["files"].as_array().unwrap() {
        keys_in(f, file_schema, "file");
        assert!(actions.contains(&f["action"]), "{f}");
    }
    assert!(
        plan["files"]
            .as_array()
            .unwrap()
            .iter()
            .any(|f| f["action"] == "skip"),
        "the plan has a skip row: {plan}"
    );
}
