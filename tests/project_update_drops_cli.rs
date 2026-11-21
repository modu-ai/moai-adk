//! The profile plan names the recorded manifests that applying it would drop
//! (SPEC-20261121-plan-shows-dropped-manifests), through the binary in temporary folders and a
//! temporary HOME only.

mod common;
use common::home::TempHome;
use serde_json::{json, Value};
use std::{fs, io::Write, path::PathBuf, process::Stdio};

const REF: &str = "test:2936";

struct Fx {
    home: TempHome,
    root: PathBuf,
}

impl Fx {
    fn new() -> Fx {
        let home = TempHome::new();
        let root = home.path().join("work").join("demo");
        fs::create_dir_all(&root).unwrap();
        Fx { home, root }
    }

    fn p(&self) -> &str {
        self.root.to_str().unwrap()
    }

    fn write(&self, rel: &str, body: &str) {
        let path = self.root.join(rel);
        fs::create_dir_all(path.parent().unwrap()).unwrap();
        fs::write(path, body).unwrap();
    }

    fn run(&self, args: &[&str], stdin: &[u8]) -> Value {
        let mut child = self
            .home
            .command(args)
            .current_dir(&self.root)
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

    /// The profile PROJECT writes for the manifests now in the project: detect, plan, apply.
    fn profile_for_now(&self) -> Value {
        let detect = self.call(&["project-detect", self.p()], b"");
        json!({
            "schema_version": 2, "project_id": "demo", "purpose": "Build a small editor",
            "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude", "codex", "omp"],
            "detected": detect["detected"],
        })
    }

    fn plan_of(&self, profile: &Value) -> Value {
        self.call(
            &["project-plan", self.p()],
            &serde_json::to_vec(profile).unwrap(),
        )
    }

    fn apply(&self, plan: &Value) -> Value {
        let digest = plan["plan_digest"].as_str().unwrap().to_string();
        self.run(
            &[
                "project-apply",
                self.p(),
                "--plan-digest",
                &digest,
                "--permission-reference",
                REF,
            ],
            &serde_json::to_vec(plan).unwrap(),
        )
    }

    /// A project with three manifests, recorded in its profile.
    fn recorded(&self) {
        self.write("package.json", "{}");
        self.write("apps/a/package.json", "{}");
        self.write("apps/b/go.mod", "module b\n");
        let plan = self.plan_of(&self.profile_for_now());
        assert_eq!(self.apply(&plan)["ok"], json!(true));
    }

    /// The two nested manifests are gone (or, in a bad run, not found).
    fn shrink(&self) {
        fs::remove_dir_all(self.root.join("apps")).unwrap();
    }

    fn profile_path(&self) -> PathBuf {
        [".moai/project.yaml", ".moai-cli/project.yaml"]
            .iter()
            .map(|rel| self.root.join(rel))
            .find(|p| p.is_file())
            .expect("the applied profile")
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

fn names(v: &Value) -> Vec<&str> {
    v.as_array()
        .unwrap()
        .iter()
        .map(|m| m.as_str().unwrap())
        .collect()
}

/// SPEC-20261121-plan-shows-dropped-manifests REQ-1: a plan that replaces a longer recorded
/// manifest list with a shorter one names what drops, sorted.
#[test]
fn the_project_plan_names_the_recorded_manifests_a_smaller_detected_list_drops() {
    let fx = Fx::new();
    fx.recorded();
    fx.shrink();
    let plan = fx.plan_of(&fx.profile_for_now());
    assert_eq!(plan["action"], "update", "{plan}");
    assert_eq!(
        names(&plan["dropped_manifests"]),
        ["apps/a/package.json", "apps/b/go.mod"],
        "{plan}"
    );
}

/// REQ-3: the update plan's profile step says how many would drop and carries the list.
#[test]
fn the_update_plan_profile_step_says_how_many_manifests_would_drop() {
    let fx = Fx::new();
    fx.recorded();
    fx.shrink();
    let plan = fx.call(&["project-update-plan", fx.p()], b"");
    let profile = step(&plan, "profile");
    assert_eq!(profile["state"], "ready", "{profile}");
    let reason = profile["reason"].as_str().unwrap();
    assert!(
        reason.contains("2 recorded manifests would drop"),
        "{reason}"
    );
    assert_eq!(
        names(&profile["preview"]["dropped_manifests"]),
        ["apps/a/package.json", "apps/b/go.mod"],
        "{profile}"
    );
}

/// REQ-1: nothing to say when the list stays or only grows.
#[test]
fn a_list_that_only_grows_or_stays_drops_nothing() {
    let fx = Fx::new();
    fx.recorded();
    let same = fx.plan_of(&fx.profile_for_now());
    assert!(same.get("dropped_manifests").is_none(), "{same}");

    fx.write("apps/c/package.json", "{}");
    let grown = fx.plan_of(&fx.profile_for_now());
    assert_eq!(grown["action"], "update", "{grown}");
    assert!(grown.get("dropped_manifests").is_none(), "{grown}");
    let update = fx.call(&["project-update-plan", fx.p()], b"");
    let profile = step(&update, "profile");
    assert!(
        !profile["reason"].as_str().unwrap().contains("would drop"),
        "{profile}"
    );
    assert!(
        profile["preview"].get("dropped_manifests").is_none(),
        "{profile}"
    );
}

/// REQ-2, REQ-4: the digest covers the list, so a plan with it removed is not applied.
#[test]
fn a_plan_with_the_dropped_list_removed_is_refused_at_apply() {
    let fx = Fx::new();
    fx.recorded();
    fx.shrink();
    let before = fs::read(fx.profile_path()).unwrap();
    let mut plan = fx.plan_of(&fx.profile_for_now());
    assert!(plan["dropped_manifests"].is_array(), "{plan}");
    plan.as_object_mut().unwrap().remove("dropped_manifests");
    let refused = fx.apply(&plan);
    assert_eq!(refused["ok"], json!(false), "{refused}");
    assert!(
        refused.to_string().contains("stale or altered plan"),
        "{refused}"
    );
    assert_eq!(fs::read(fx.profile_path()).unwrap(), before);
}

/// REQ-1: a `detected` key the user changed by hand is kept, so nothing of it drops.
#[test]
fn a_hand_edited_detected_key_drops_nothing() {
    let fx = Fx::new();
    fx.recorded();
    let path = fx.profile_path();
    let mut recorded: Value = serde_json::from_slice(&fs::read(&path).unwrap()).unwrap();
    let manifests = recorded["detected"]["manifests"].as_array_mut().unwrap();
    manifests.push(json!("tools/extra/package.json"));
    manifests.sort_by(|a, b| a.as_str().cmp(&b.as_str()));
    fs::write(&path, serde_json::to_vec_pretty(&recorded).unwrap()).unwrap();
    fx.shrink();
    let plan = fx.plan_of(&fx.profile_for_now());
    assert_eq!(plan["action"], "conflict", "{plan}");
    assert!(plan.get("dropped_manifests").is_none(), "{plan}");
}
