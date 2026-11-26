//! SPEC-20261121-resurrected-marker: a project whose `moai.json` in the state folder is valid
//! stays Rust-owned when git brings the Go markers back (`manifest.json`,
//! `config/sections/system.yaml`), so the card commands read the project profile from the state
//! folder and `project-update-plan` offers to keep the markers instead of stopping. The project
//! has the shape of the field case: a git repository with every file committed and no earlier
//! state folder. Everything runs in a temporary HOME; nothing moves or is deleted. The folder
//! name comes from `moai_cli::state_root::MOAI_DIR`, never from a literal.

mod common;

use common::home::TempHome;
use moai_cli::ledger;
use serde_json::{json, Value};
use std::fs;
use std::path::Path;
use std::process::{Output, Stdio};

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
const SESSION: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const MARKER: &str =
    r#"{"schema_version":1,"generation":"moai","created_by":"moai test","project_id":"demo"}"#;
const PROFILE: &str = r#"{"schema_version":2,"project_id":"demo"}"#;
const SYSTEM_YAML: &str = "# System\nmoai:\n  version: \"2.4.0\"\n";

fn git(dir: &Path, args: &[&str]) -> String {
    let mut all = vec![
        "-c",
        "user.name=fixture",
        "-c",
        "user.email=fixture@example.invalid",
    ];
    all.extend_from_slice(args);
    common::git::git(dir, &all)
}

/// `rel` inside this generation's state folder, as a project-relative path.
fn state(rel: &str) -> String {
    format!("{}/{rel}", moai_cli::state_root::MOAI_DIR)
}

fn put(root: &Path, rel: &str, text: &str) {
    let p = root.join(rel);
    fs::create_dir_all(p.parent().unwrap()).unwrap();
    fs::write(p, text).unwrap();
}

struct Fx {
    home: TempHome,
    p: String,
}

impl Fx {
    /// The field shape: the profile only in the state folder, the Go markers committed, and the
    /// Rust marker too when `marker` is true. Nothing is left in the earlier state folder.
    fn new(marker: bool) -> Fx {
        let home = TempHome::new();
        let project = home.project("demo");
        fs::remove_dir_all(project.join(moai_cli::PROJECT_STATE_DIR)).unwrap();
        put(&project, &state("project.yaml"), PROFILE);
        put(&project, &state("manifest.json"), "{}");
        put(&project, &state("config/sections/system.yaml"), SYSTEM_YAML);
        if marker {
            put(&project, &state("moai.json"), MARKER);
        }
        git(&project, &["init", "-q", "-b", "main"]);
        git(&project, &["add", "-A"]);
        git(&project, &["commit", "-q", "-m", "init"]);
        ledger::register(home.path(), "demo").unwrap();
        let p = project.to_str().unwrap().to_string();
        Fx { home, p }
    }

    fn run(&self, args: &[&str]) -> Value {
        let out: Output = self
            .home
            .command(args)
            .env(SESSION_VAR, SESSION)
            .stdin(Stdio::null())
            .output()
            .expect("run moai");
        serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
            panic!(
                "{args:?}: {} / {}",
                String::from_utf8_lossy(&out.stdout),
                String::from_utf8_lossy(&out.stderr)
            )
        })
    }

    fn ok(&self, args: &[&str]) -> Value {
        let v = self.run(args);
        assert_eq!(v["ok"], json!(true), "{args:?}: {v}");
        v["result"].clone()
    }

    fn plan(&self) -> Value {
        self.ok(&["project-update-plan", &self.p])
    }
}

fn step<'a>(plan: &'a Value, name: &str) -> Option<&'a Value> {
    plan["phases"]
        .as_array()
        .unwrap()
        .iter()
        .flat_map(|ph| ph["steps"].as_array().unwrap())
        .find(|s| s["step"] == json!(name))
}

#[test]
fn card_commands_work_when_go_markers_came_back_beside_a_valid_marker() {
    let fx = Fx::new(true);
    let added = fx.ok(&[
        "card",
        "add",
        "--title",
        "t",
        "--scope",
        "apps/demo/",
        "--project",
        &fx.p,
    ]);
    let name = added["name"].as_str().unwrap().to_string();
    let listed = fx.ok(&["card", "list", &fx.p]);
    assert!(listed.to_string().contains(&name), "{listed}");
    let shown = fx.ok(&["card", "show", &name, "--project", &fx.p]);
    assert!(shown.to_string().contains("\"t\""), "{shown}");
}

#[test]
fn go_markers_without_a_valid_marker_keep_failing_the_way_they_do() {
    let fx = Fx::new(false);
    let v = fx.run(&["card", "list", &fx.p]);
    assert_eq!(v["ok"], json!(false), "{v}");
    let message = v["error"]["message"].as_str().unwrap();
    assert!(message.contains("project profile missing"), "{v}");
    let plan = fx.plan();
    assert_eq!(plan["state_root"]["generation"], json!("go"), "{plan}");
    // The read-only `go-residue` survey (SPEC-20261121-go-residue-detect) is in every plan. With
    // no valid marker and no usable manifest it proves nothing, offers no command and no keeping
    // of the markers.
    let residue = step(&plan, "go-residue").unwrap_or_else(|| panic!("no go-residue: {plan}"));
    assert_eq!(residue["state"], json!("unchanged"), "{residue}");
    assert!(residue["apply"].is_null(), "{residue}");
    assert_eq!(residue["preview"]["items"], json!([]), "{residue}");
    assert_eq!(residue["preview"]["proof"], json!("none"), "{residue}");
    assert!(residue["preview"]["markers"].is_null(), "{residue}");
    assert!(residue["preview"]["keep_in"].is_null(), "{residue}");
    let go_state = step(&plan, "go-state").unwrap();
    assert!(
        !go_state["reason"].as_str().unwrap().contains("go-residue"),
        "{go_state}"
    );
}

// The second half of SPEC-20261121-resurrected-marker: this test is RED on purpose until
// `project-update-plan` offers to keep the resurrected markers (REQ-4, REQ-8). The `go-residue`
// step exists as a read-only survey and reports `unchanged` here, so no offer is made yet. It is
// ignored so the suite stays green and is run by name (`-- --ignored`) when the offer is written.
#[test]
#[ignore = "RED until the plan offers to keep the resurrected markers (second half of the SPEC)"]
fn the_plan_offers_to_keep_the_resurrected_markers_instead_of_stopping() {
    let fx = Fx::new(true);
    let plan = fx.plan();
    assert_eq!(plan["state_root"]["generation"], json!("mixed"), "{plan}");
    let residue = step(&plan, "go-residue").unwrap_or_else(|| panic!("no go-residue: {plan}"));
    assert_eq!(residue["state"], json!("needs_decision"), "{residue}");
    assert!(residue["apply"].is_null(), "no command yet: {residue}");
    let markers = residue["preview"]["markers"].as_array().unwrap();
    let listed: Vec<(&str, bool)> = markers
        .iter()
        .map(|m| (m["path"].as_str().unwrap(), m["tracked"].as_bool().unwrap()))
        .collect();
    let (system_yaml, manifest) = (state("config/sections/system.yaml"), state("manifest.json"));
    assert_eq!(
        listed,
        vec![(system_yaml.as_str(), true), (manifest.as_str(), true)],
        "{residue}"
    );
    assert!(
        residue["preview"]["keep_in"]
            .as_str()
            .unwrap()
            .starts_with(&state("backups/go-markers-")),
        "{residue}"
    );
    let go_state = step(&plan, "go-state").unwrap();
    assert_eq!(go_state["state"], json!("skip"), "{go_state}");
    assert!(
        go_state["reason"].as_str().unwrap().contains("go-residue"),
        "{go_state}"
    );
    assert_eq!(plan["all_unchanged"], json!(false), "{plan}");
}

#[test]
fn the_plan_writes_nothing() {
    let fx = Fx::new(true);
    let before = fx.home.snapshot();
    fx.plan();
    assert_eq!(
        fx.home.snapshot(),
        before,
        "no file, folder or time changed"
    );
}
