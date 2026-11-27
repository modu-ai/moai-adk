//! SPEC-20261127-update-old-manifest-way: a generation manifest that owns project agent and
//! skill files but records no `generation_request` gets a way forward in `project-update-plan`:
//! the choice `keep-project-files` plans the rest of the generation (the managed block, the
//! workflow files) and leaves those files and their manifest lines as they are. Every case runs
//! the built binary in a temporary project with a temporary HOME; the real checkout and HOME are
//! never read.

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
const REF: &str = "test:old-manifest-way";
const PLACEHOLDER: &str = "<permission-reference>";
const KEEP: &str = "keep-project-files";

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

    /// A profile written the way PROJECT writes it.
    fn with_profile(&self) {
        let detect = self.call(&["project-detect", self.p()], b"");
        let profile = json!({
            "schema_version": 2, "project_id": "demo", "purpose": "Build a small editor",
            "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude", "codex"],
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

    fn plan(&self) -> Value {
        self.call(&["project-update-plan", self.p()], b"")
    }

    /// `generation-plan` with `request`, as the whole answer.
    fn generation_plan(&self, request: &Value) -> Value {
        self.call(
            &["generation-plan", self.p()],
            &serde_json::to_vec(request).unwrap(),
        )
    }

    /// `generation-plan` with `request`, then `generation-apply` of that plan.
    fn generate(&self, request: &Value) {
        let plan = self.generation_plan(request);
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

    /// Runs a printed command (a step or one of its choices) with the permission reference
    /// filled in.
    fn apply(&self, printed: &Value) -> Value {
        let argv: Vec<String> = printed["apply"]["argv"]
            .as_array()
            .unwrap_or_else(|| panic!("a printed command is expected: {printed}"))
            .iter()
            .map(|w| w.as_str().unwrap().to_string())
            .collect();
        let args: Vec<&str> = argv[1..]
            .iter()
            .map(|w| if w == PLACEHOLDER { REF } else { w.as_str() })
            .collect();
        let stdin = match &printed["apply"]["stdin"] {
            Value::Null => Vec::new(),
            v => serde_json::to_vec(v).unwrap(),
        };
        self.run(&args, &stdin)
    }

    fn manifest_path(&self) -> PathBuf {
        self.root.join(moai_cli::state_root::rel(
            &self.root,
            "manifests/generation.json",
        ))
    }

    fn manifest(&self) -> Value {
        serde_json::from_slice(&fs::read(self.manifest_path()).unwrap()).unwrap()
    }

    fn write_manifest(&self, manifest: &Value) {
        let mut bytes = serde_json::to_vec_pretty(manifest).unwrap();
        bytes.push(b'\n');
        fs::write(self.manifest_path(), bytes).unwrap();
    }

    /// An older manifest that never recorded the managed block: its `CLAUDE.md` entry is gone and
    /// the file is deleted, so the generation has a block to create.
    fn forget_block(&self) {
        let mut manifest = self.manifest();
        let list = manifest["owned_files"].as_array_mut().unwrap();
        let before = list.len();
        list.retain(|f| f["path"] != "CLAUDE.md");
        assert!(list.len() < before, "the manifest owns the CLAUDE.md block");
        self.write_manifest(&manifest);
        fs::remove_file(self.root.join("CLAUDE.md")).unwrap();
    }

    /// An older manifest whose block baseline is not what is rendered now, over a `CLAUDE.md` the
    /// user rewrote with no block: the missing block is a conflict.
    fn stale_block(&self, text: &str) {
        let mut manifest = self.manifest();
        let entry = manifest["owned_files"]
            .as_array_mut()
            .unwrap()
            .iter_mut()
            .find(|f| f["path"] == "CLAUDE.md")
            .expect("the manifest owns the CLAUDE.md block");
        entry["baseline_sha256"] = json!(format!("sha256:{}", "a".repeat(64)));
        self.write_manifest(&manifest);
        fs::write(self.root.join("CLAUDE.md"), text).unwrap();
    }

    /// The owned entries of the three project kinds, as the manifest holds them.
    fn kept_entries(&self) -> Vec<Value> {
        self.manifest()["owned_files"]
            .as_array()
            .unwrap()
            .iter()
            .filter(|f| {
                f["section_id"].as_str().is_some_and(|s| {
                    ["native-role", "project-skill", "skill-projection"]
                        .iter()
                        .any(|k| s.starts_with(&format!("generation:{k}:")))
                })
            })
            .cloned()
            .collect()
    }

    /// The bytes of the files the kept entries name, in entry order.
    fn kept_bytes(&self) -> Vec<(String, Vec<u8>)> {
        self.kept_entries()
            .iter()
            .map(|f| {
                let rel = f["path"].as_str().unwrap().to_string();
                let bytes = fs::read(self.root.join(&rel)).unwrap_or_default();
                (rel, bytes)
            })
            .collect()
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

fn role() -> Value {
    json!({"role_id": "prj-test-runner", "host": "claude",
           "purpose": "Runs the test command and reports what failed",
           "read_only_intent": false, "tools": []})
}

fn skill() -> Value {
    json!({"id": "prj-checks", "hosts": [],
           "purpose": "The checks to run before a change is done"})
}

/// The request the generation step prints for a project whose profile is applied.
fn base_request(fx: &Fx) -> Value {
    let request = step(&fx.plan(), "generation")["apply"]["stdin"]["request"].clone();
    assert!(request.is_object(), "no request printed: {request}");
    request
}

/// A project with one project agent and one project skill generated and every other printed
/// `ready` command applied, whose manifest was written before the request record existed: the
/// files are owned and current, no `generation_request` is recorded.
fn old_manifest_project() -> Fx {
    let fx = Fx::new();
    fx.with_profile();
    let mut request = base_request(&fx);
    request["native_roles"] = json!([role()]);
    request["project_skills"] = json!([skill()]);
    fx.generate(&request);
    for s in steps(&fx.plan())
        .into_iter()
        .filter(|s| s["state"] == "ready")
    {
        assert_eq!(fx.apply(s)["ok"], json!(true), "{s}");
    }
    let mut manifest = fx.manifest();
    manifest
        .as_object_mut()
        .unwrap()
        .remove("generation_request")
        .expect("the generated project recorded its request");
    fx.write_manifest(&manifest);
    fx
}

/// The rows of a plan or of a step preview, as (path, action).
fn rows(files: &Value) -> Vec<(String, String)> {
    files
        .as_array()
        .into_iter()
        .flatten()
        .map(|f| {
            (
                f["path"].as_str().unwrap().to_string(),
                f["action"].as_str().unwrap().to_string(),
            )
        })
        .collect()
}

fn choice<'a>(generation: &'a Value, id: &str) -> &'a Value {
    generation["preview"]["choices"]
        .as_array()
        .into_iter()
        .flatten()
        .find(|c| c["id"] == id)
        .unwrap_or_else(|| panic!("no choice {id}: {generation}"))
}

fn names_a_kept_file(generation: &Value, kept: &[(String, Vec<u8>)]) {
    let listed: Vec<&str> = generation["preview"]["not_rebuilt"]
        .as_array()
        .into_iter()
        .flatten()
        .map(|p| p.as_str().unwrap())
        .collect();
    assert!(!listed.is_empty(), "{generation}");
    for path in &listed {
        assert!(kept.iter().any(|(rel, _)| rel == path), "{path}: {kept:?}");
    }
}

#[test]
fn old_manifest_offers_keep_and_plans_the_managed_block() {
    let fx = old_manifest_project();
    let kept = fx.kept_bytes();
    // The managed block was never recorded: the generation has a row to write besides the kept
    // files.
    fx.forget_block();
    let plan = fx.plan();
    let generation = step(&plan, "generation");
    assert_eq!(generation["state"], "needs_decision", "{generation}");
    assert_eq!(generation["apply"], Value::Null, "{generation}");
    names_a_kept_file(generation, &kept);
    let shown = rows(&generation["preview"]["files"]);
    assert!(
        shown.iter().any(|(p, a)| p == "CLAUDE.md" && a == "create"),
        "the managed block is planned: {generation}"
    );
    assert!(
        shown.iter().all(|(_, a)| a != "remove" && a != "conflict"),
        "{generation}"
    );
    let keep = choice(generation, KEEP);
    assert_eq!(keep["recommended"], json!(true), "{keep}");
    assert!(keep["apply"]["argv"].is_array(), "{keep}");
    assert_eq!(keep["picks"], json!(["keep-harness"]), "{keep}");
    let planned = rows(&keep["apply"]["stdin"]["files"]);
    assert!(
        planned.iter().all(|(_, a)| a != "remove"),
        "the choice removes nothing: {keep}"
    );
}

#[test]
fn keep_apply_writes_the_rest_and_keeps_the_project_files() {
    let fx = old_manifest_project();
    let kept = fx.kept_bytes();
    let entries = fx.kept_entries();
    fx.forget_block();
    let plan = fx.plan();
    let keep = choice(step(&plan, "generation"), KEEP).clone();
    let applied = fx.apply(&keep);
    assert_eq!(applied["ok"], json!(true), "{applied}");
    assert!(fx.root.join("CLAUDE.md").is_file(), "the block is written");
    assert_eq!(fx.kept_bytes(), kept, "the project files keep their bytes");
    assert_eq!(fx.kept_entries(), entries, "their manifest lines stay");
    assert!(
        fx.manifest().get("generation_request").is_none(),
        "nothing is recorded for files the request does not name: {}",
        fx.manifest()
    );
}

#[test]
fn an_old_manifest_with_every_row_current_reads_unchanged() {
    let fx = old_manifest_project();
    let kept = fx.kept_bytes();
    let plan = fx.plan();
    let generation = step(&plan, "generation");
    assert_eq!(generation["state"], "unchanged", "{generation}");
    names_a_kept_file(generation, &kept);
}

#[test]
fn after_keep_the_step_reads_unchanged() {
    let fx = old_manifest_project();
    let kept = fx.kept_bytes();
    fx.forget_block();
    let keep = choice(step(&fx.plan(), "generation"), KEEP).clone();
    assert_eq!(fx.apply(&keep)["ok"], json!(true));
    let plan = fx.plan();
    let generation = step(&plan, "generation");
    assert_eq!(generation["state"], "unchanged", "{generation}");
    names_a_kept_file(generation, &kept);
    assert!(
        generation["preview"]["choices"]
            .as_array()
            .is_none_or(|c| c.is_empty()),
        "nothing is left to choose: {generation}"
    );
}

#[test]
fn a_conflict_keeps_its_choices_and_the_project_files() {
    let fx = old_manifest_project();
    let kept = fx.kept_bytes();
    // A CLAUDE.md the user wrote with no managed block, over a stale baseline: the missing block
    // is a conflict row.
    fx.stale_block("# mine\n");
    let plan = fx.plan();
    let generation = step(&plan, "generation");
    assert_eq!(generation["state"], "conflict", "{generation}");
    let append = choice(generation, "append-managed").clone();
    assert_eq!(append["picks"], json!(["append-managed", "keep-harness"]));
    assert_eq!(fx.apply(&append)["ok"], json!(true));
    let text = fs::read_to_string(fx.root.join("CLAUDE.md")).unwrap();
    assert!(text.starts_with("# mine\n"), "{text}");
    assert!(text.len() > "# mine\n".len(), "the block is appended");
    assert_eq!(fx.kept_bytes(), kept, "the project files keep their bytes");
}

#[test]
fn the_pick_stops_removals_and_still_plans_the_named_files() {
    let fx = Fx::new();
    fx.with_profile();
    let mut request = base_request(&fx);
    request["native_roles"] = json!([role()]);
    request["project_skills"] = json!([skill()]);
    fx.generate(&request);

    // A request that names neither: without the pick both owned files are planned `remove`.
    let mut none = request.clone();
    none["native_roles"] = json!([]);
    none["project_skills"] = json!([]);
    let without = fx.generation_plan(&none);
    let removed: Vec<&str> = without["files"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|f| f["action"] == "remove")
        .map(|f| f["kind"].as_str().unwrap())
        .collect();
    assert_eq!(removed.len(), 2, "{without}");

    // With the pick they have no row at all.
    none["picks"] = json!(["keep-harness"]);
    let kept = fx.generation_plan(&none);
    assert!(
        kept["files"]
            .as_array()
            .unwrap()
            .iter()
            .all(|f| f["action"] != "remove" && f["action"] != "conflict"),
        "{kept}"
    );

    // A skill the request does name is planned as it is without the pick.
    let mut named = none.clone();
    named["project_skills"] = json!([{"id": "prj-new", "hosts": [],
                                     "purpose": "A skill the request names"}]);
    let plan = fx.generation_plan(&named);
    let create = plan["files"]
        .as_array()
        .unwrap()
        .iter()
        .find(|f| {
            f["kind"]
                .as_str()
                .is_some_and(|k| k.starts_with("generation:project-skill:claude:prj-new"))
        })
        .unwrap_or_else(|| panic!("the named skill is planned: {plan}"));
    assert_eq!(create["action"], "create", "{create}");
    assert!(
        plan["files"]
            .as_array()
            .unwrap()
            .iter()
            .all(|f| f["action"] != "remove"),
        "{plan}"
    );
}
