//! #2936 L2a (leader decisions D1 = A and D2 = C, comment 6904376186): `go-settings-plan`,
//! `go-settings-apply` and `go-settings-revert` clean the Go-made `.moai/` references out of
//! `.claude/settings.json`, and `project-update-plan` lists them as the `go-settings` step of
//! the backup phase, before `go-state`.
//!
//! Every case builds a temporary Go-generation project with an empty temporary HOME and a fake
//! Rust launcher at `HOME/.moai-cli/path/moai-cli` (a shell script); nothing outside the
//! temporary folders is read or written. The shapes are the reproduction's G (the Go manifest
//! does not record the settings file) and H (it does, and a Rust settings receipt owns a key).
//! A preview leaves the whole tree byte for byte as it was.

#![cfg(unix)]

mod common;
#[path = "common/go_project_real_run.rs"]
mod go_project_real_run;

use go_project_real_run::{step_in, GoFx, REF};
use moai_cli::go_settings_clean::{count_plans, PLANS_WALK_LIMIT};
use moai_cli::status_line::RUST_HELP_LINE;
use serde_json::{json, Value};
use std::fs;
use std::os::unix::fs::PermissionsExt as _;
use std::path::Path;

const SETTINGS: &str = ".claude/settings.json";
const LOCAL: &str = ".claude/settings.local.json";
const LAUNCHER: &str = ".moai-cli/path/moai-cli";
const SHOWN: &str = "$HOME/.moai-cli/path/moai-cli";
const PLACEHOLDER: &str = "<permission-reference>";

/// The probe limit the hanging launcher is cut at (debug builds only).
const SHORT_PROBE: [(&str, &str); 1] = [("MOAI_CLI_TEST_PROBE_TIMEOUT_MS", "500")];

/// A launcher that answers `statusline --help` the way the Rust build does.
fn rust_launcher() -> String {
    format!(
        "#!/bin/sh\nif [ \"$1\" = statusline ] && [ \"$2\" = --help ]; then\n  echo '{RUST_HELP_LINE}'\n  echo 'Usage: moai statusline'\nfi\n"
    )
}

/// The Go tool's settings: three Go-made keys, and keys that are not its business.
fn go_settings() -> Value {
    json!({
        "permissions": {"deny": ["Read(./.env)"], "allow": ["Bash(ls:*)"]},
        "statusLine": {"type": "command", "command": "$CLAUDE_PROJECT_DIR/.moai/status_line.sh"},
        "model": "opus",
        "subagentStatusLine": {"type": "command",
                               "command": "\"${CLAUDE_PROJECT_DIR}\"/.moai/status_line.sh"},
        "plansDirectory": ".moai/plans",
        "env": {"FOO": "1"},
    })
}

/// The local file of most cases: the user's own, naming nothing under `.moai/`.
const LOCAL_BODY: &str = "{\"env\":{\"LOCAL\":\"1\"}}\n";

/// The local file that holds Go references (D1 condition 3: it is never written).
const LOCAL_WITH_GO_REFS: &str = "{\"statusLine\":{\"type\":\"command\",\"command\":\"$CLAUDE_PROJECT_DIR/.moai/status_line.sh\"},\"plansDirectory\":\".moai/plans\"}\n";

fn put_launcher(fx: &GoFx, body: &str) {
    let path = fx.home.join(LAUNCHER);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    common::write_executable(&path, body);
}

fn base_fx() -> GoFx {
    let mut fx = GoFx::new();
    fx.deployed(".moai/status_line.sh", "#!/bin/sh\necho status\n");
    fx.write(".moai/plans/a.md", "# a\n");
    fx.write(".moai/plans/sub/b.md", "# b\n");
    fx.write(LOCAL, LOCAL_BODY);
    put_launcher(&fx, &rust_launcher());
    fx
}

/// Shape G: the Go manifest does not record the settings file.
fn shape_g() -> GoFx {
    let fx = base_fx();
    fx.manifest();
    fx.write(SETTINGS, &go_settings().to_string());
    fx
}

/// Shape H: the manifest records the settings file and a Rust settings receipt owns a key.
fn shape_h() -> GoFx {
    let mut fx = base_fx();
    fx.deployed(SETTINGS, &go_settings().to_string());
    fx.manifest();
    fx.write(
        ".moai-cli/manifests/claude-settings.json",
        &json!({"schema_version": 1, "owned": {"permissions": go_settings()["permissions"]}})
            .to_string(),
    );
    fx
}

fn plan_env(fx: &GoFx, env: &[(&str, &str)]) -> Value {
    let before = fx.tree();
    let v = fx.run_env(&["go-settings-plan", fx.p()], env);
    assert_eq!(v["ok"], json!(true), "{v}");
    assert_eq!(fx.tree(), before, "a preview writes nothing");
    v["result"].clone()
}

fn plan(fx: &GoFx) -> Value {
    plan_env(fx, &[])
}

fn apply_with(fx: &GoFx, digest: &str, reference: &str, env: &[(&str, &str)]) -> Value {
    fx.run_env(
        &[
            "go-settings-apply",
            fx.p(),
            "--plan-digest",
            digest,
            "--permission-reference",
            reference,
        ],
        env,
    )
}

fn apply(fx: &GoFx, plan: &Value) -> Value {
    apply_with(fx, plan["plan_digest"].as_str().unwrap(), REF, &[])
}

fn revert(fx: &GoFx, id: &str) -> Value {
    fx.run(&[
        "go-settings-revert",
        fx.p(),
        "--receipt",
        id,
        "--permission-reference",
        REF,
    ])
}

fn read(fx: &GoFx, rel: &str) -> Vec<u8> {
    fs::read(fx.root.join(rel)).unwrap()
}

fn settings(fx: &GoFx) -> Value {
    serde_json::from_slice(&read(fx, SETTINGS)).unwrap()
}

fn keys_of(v: &Value) -> Vec<String> {
    v.as_array()
        .unwrap()
        .iter()
        .map(|k| k.as_str().unwrap().to_string())
        .collect()
}

fn go_keys(plan: &Value) -> Vec<(String, String)> {
    plan["go_keys"]
        .as_array()
        .unwrap()
        .iter()
        .map(|k| {
            (
                k["key"].as_str().unwrap().to_string(),
                k["change"].as_str().unwrap().to_string(),
            )
        })
        .collect()
}

fn error_text(v: &Value) -> String {
    assert_eq!(v["ok"], json!(false), "{v}");
    v.to_string()
}

fn the_three() -> Vec<(String, String)> {
    [
        ("plansDirectory", "remove"),
        ("statusLine", "replace"),
        ("subagentStatusLine", "remove"),
    ]
    .map(|(k, c)| (k.to_string(), c.to_string()))
    .to_vec()
}

/// The receipt folders of the go-settings clean-ups in the project.
fn receipts(fx: &GoFx) -> Vec<String> {
    let dir = fx.root.join(".moai-cli/receipts");
    let Ok(read) = fs::read_dir(dir) else {
        return Vec::new();
    };
    let mut out: Vec<String> = read
        .flatten()
        .map(|e| e.file_name().to_string_lossy().into_owned())
        .filter(|n| n.starts_with("go-settings-"))
        .collect();
    out.sort();
    out
}

fn receipt_status(fx: &GoFx, id: &str) -> String {
    let v: Value =
        serde_json::from_slice(&read(fx, &format!(".moai-cli/receipts/{id}/receipt.json")))
            .unwrap();
    v["status"].as_str().unwrap().to_string()
}

/// The reproduction's G and H, end to end: the preview names the keys and the command, the
/// apply cleans exactly them, and the next preview no longer finds the go-state step held by
/// the settings. Every other key keeps its value and its place; the local file is untouched.
fn plan_apply_replan(fx: GoFx) {
    let local = read(&fx, LOCAL);
    let before = settings(&fx);

    // The reproduction's loop: go-state is held by the settings, and says whose keys.
    let up = fx.update_plan();
    let held = step_in(&up, 0, "go-state");
    assert_eq!(held["state"], "blocked", "{held}");
    let why = held["reason"].as_str().unwrap();
    assert!(why.contains("settings reference .moai/"), "{why}");

    let p = plan(&fx);
    assert_eq!(p["action"], json!("replace"), "{p}");
    assert_eq!(go_keys(&p), the_three(), "{p}");
    assert_eq!(p["kept"], json!([]), "{p}");
    assert_eq!(
        p["launcher"],
        json!({"path": SHOWN, "state": "rust"}),
        "{p}"
    );
    assert_eq!(p["writes"], json!(false), "{p}");
    assert_eq!(p["local_settings"], json!("never written"), "{p}");
    // D1 condition 2: the files under `.moai/plans` are counted and not deleted.
    assert_eq!(p["plans_files"]["count"], json!(2), "{p}");
    assert_eq!(p["plans_files"]["more_than_limit"], json!(false), "{p}");
    assert_eq!(p["plans_files"]["deleted"], json!(false), "{p}");
    assert!(
        p["plans_files"]["note"]
            .as_str()
            .unwrap()
            .contains("not deleted"),
        "{p}"
    );

    // The update plan lists the step first and prints the command like the hooks step does.
    let up = fx.update_plan();
    let backup: Vec<&str> = up["phases"][0]["steps"]
        .as_array()
        .unwrap()
        .iter()
        .map(|s| s["step"].as_str().unwrap())
        .collect();
    assert_eq!(backup, ["go-settings", "go-state"], "{up}");
    let step = step_in(&up, 0, "go-settings");
    assert_eq!(step["state"], "ready", "{step}");
    assert_eq!(step["plan_digest"], p["plan_digest"], "{step}");
    assert_eq!(
        step["apply"]["argv"],
        json!([
            "moai",
            "go-settings-apply",
            fx.p(),
            "--plan-digest",
            p["plan_digest"],
            "--permission-reference",
            PLACEHOLDER
        ]),
        "{step}"
    );
    assert_eq!(step["apply"]["stdin"], Value::Null, "{step}");
    // go-state stays blocked, but says which step clears the keys and that a new preview follows.
    let held = step_in(&up, 0, "go-state");
    assert_eq!(held["state"], "blocked", "{held}");
    let why = held["reason"].as_str().unwrap();
    assert!(
        why.contains("the go-settings step above clears the Go-made keys")
            && why.contains("plansDirectory, statusLine, subagentStatusLine")
            && why.contains("preview again"),
        "{why}"
    );
    assert_eq!(held["apply"], Value::Null, "{held}");

    let out = fx.apply_step(step);
    assert_eq!(out["ok"], json!(true), "{out}");
    assert_eq!(out["result"]["state"], json!("applied"), "{out}");
    assert_eq!(
        out["result"]["removed"],
        json!(["plansDirectory", "subagentStatusLine"])
    );
    assert_eq!(out["result"]["replaced"], json!(["statusLine"]));

    let after = settings(&fx);
    assert_eq!(
        after["statusLine"],
        json!({"type": "command", "command": format!("\"{SHOWN}\" statusline")}),
        "{after}"
    );
    assert!(after.get("plansDirectory").is_none(), "{after}");
    assert!(after.get("subagentStatusLine").is_none(), "{after}");
    // Every other key: same value, same place.
    let names = |v: &Value| -> Vec<String> {
        v.as_object()
            .unwrap()
            .keys()
            .filter(|k| {
                !["statusLine", "plansDirectory", "subagentStatusLine"].contains(&k.as_str())
            })
            .cloned()
            .collect()
    };
    assert_eq!(names(&after), names(&before));
    for k in names(&before) {
        assert_eq!(after[&k].to_string(), before[&k].to_string(), "{k}");
    }
    assert_eq!(
        read(&fx, LOCAL),
        local,
        "the local settings are never written"
    );

    // The next preview: nothing left to clean, and go-state is no longer held by the settings.
    let again = fx.update_plan();
    assert_eq!(
        step_in(&again, 0, "go-settings")["state"],
        "unchanged",
        "{again}"
    );
    let state = step_in(&again, 0, "go-state");
    let why = state["reason"].as_str().unwrap();
    assert!(!why.contains("settings reference"), "{state}");
    // The `.moai/` folder now moves: a decision with its command, not a hold.
    assert_eq!(state["state"], "needs_decision", "{state}");
    assert_eq!(
        state["apply"]["argv"][1],
        json!("go-state-apply"),
        "{state}"
    );
    assert_eq!(plan(&fx)["action"], json!("none"));
}

#[test]
fn shape_g_plan_apply_replan() {
    plan_apply_replan(shape_g());
}

#[test]
fn shape_h_plan_apply_replan() {
    plan_apply_replan(shape_h());
}

/// The whole chain on shape G: settings clean-up, then the go-state move, then the install
/// phase: the `claude-settings` step must not try to overwrite the status line the clean-up
/// wrote.
#[test]
fn after_the_go_state_move_claude_settings_leaves_the_new_status_line() {
    let fx = shape_g();
    fx.with_profile();
    let up = fx.update_plan();
    // The install steps wait behind the backup phase while a clean-up is to be applied.
    assert_eq!(
        step_in(&up, 1, "claude-settings")["reason"],
        json!("re-run after the backup phase")
    );
    let out = fx.apply_step(step_in(&up, 0, "go-settings"));
    assert_eq!(out["ok"], json!(true), "{out}");
    let written = read(&fx, SETTINGS);

    let up = fx.update_plan();
    let state = step_in(&up, 0, "go-state");
    assert_eq!(state["state"], "needs_decision", "{state}");
    let out = fx.apply_step(state);
    assert_eq!(out["ok"], json!(true), "{out}");

    let up = fx.update_plan();
    assert_eq!(step_in(&up, 0, "go-settings")["state"], "unchanged", "{up}");
    let cs = step_in(&up, 1, "claude-settings");
    assert_ne!(cs["state"], "blocked", "{cs}");
    // The status line is not among the keys the step would write, and it says why.
    assert!(cs["apply"]["stdin"].get("statusLine").is_none(), "{cs}");
    assert_eq!(
        cs["preview"]["status_line_migration"]["reason"],
        json!("foreign"),
        "{cs}"
    );
    assert_eq!(
        cs["preview"]["status_line_migration"]["state"],
        json!("none"),
        "{cs}"
    );
    assert_eq!(
        read(&fx, SETTINGS),
        written,
        "the later previews wrote nothing"
    );
    // Applying what the step prints (when it prints anything) keeps the line as it is.
    if !cs["apply"].is_null() {
        let out = fx.apply_step(cs);
        assert_eq!(out["ok"], json!(true), "{out}");
    }
    assert_eq!(
        settings(&fx)["statusLine"],
        json!({"type": "command", "command": format!("\"{SHOWN}\" statusline")})
    );
}

/// D1 condition 1: a value that differs by even one character is the user's. It stays, and the
/// preview names its key (never the value).
#[test]
fn a_value_one_character_off_is_kept_and_named_by_key() {
    let fx = base_fx();
    fx.manifest();
    let mut s = go_settings();
    // One character more in the user's status line and plans folder, a private value elsewhere.
    s["statusLine"]["command"] = json!("$CLAUDE_PROJECT_DIR/.moai/status_line.sh ");
    s["plansDirectory"] = json!(".moai/plans/");
    s["customWrapper"] = json!({"command": "my-private-wrapper .moai/mine.sh"});
    fx.write(SETTINGS, &s.to_string());
    let p = plan(&fx);
    assert_eq!(p["action"], json!("replace"), "{p}");
    // Only the subagent line is Go-made; the status line is the user's, so no launcher is needed.
    assert_eq!(
        go_keys(&p),
        [("subagentStatusLine".to_string(), "remove".to_string())],
        "{p}"
    );
    assert_eq!(
        keys_of(&p["kept"]),
        ["customWrapper", "plansDirectory", "statusLine"],
        "{p}"
    );
    assert_eq!(p["launcher"]["state"], json!("not_needed"), "{p}");
    assert_eq!(p["plans_files"], Value::Null, "{p}");
    let text = p.to_string();
    for value in [
        "my-private-wrapper",
        "mine.sh",
        "status_line.sh",
        "CLAUDE_PROJECT_DIR",
    ] {
        assert!(
            !text.contains(value),
            "a value is never printed: {value}: {p}"
        );
    }
    // The update plan names the kept keys too, without their values.
    let up = fx.update_plan();
    let step = step_in(&up, 0, "go-settings");
    assert_eq!(step["state"], "ready", "{step}");
    assert_eq!(
        keys_of(&step["preview"]["kept"]),
        ["customWrapper", "plansDirectory", "statusLine"],
        "{step}"
    );
    assert!(!up.to_string().contains("my-private-wrapper"), "{up}");
    let why = step_in(&up, 0, "go-state")["reason"]
        .as_str()
        .unwrap()
        .to_string();
    assert!(why.contains("your values"), "{why}");

    let before = settings(&fx);
    assert_eq!(apply(&fx, &p)["ok"], json!(true));
    let after = settings(&fx);
    assert!(after.get("subagentStatusLine").is_none(), "{after}");
    for k in [
        "statusLine",
        "plansDirectory",
        "customWrapper",
        "model",
        "env",
        "permissions",
    ] {
        assert_eq!(after[k], before[k], "{k}");
    }
    // Nothing is Go-made any more; the kept keys still hold the folder, and go-state says so.
    let p = plan(&fx);
    assert_eq!(p["action"], json!("none"), "{p}");
    assert_eq!(
        keys_of(&p["kept"]),
        ["customWrapper", "plansDirectory", "statusLine"]
    );
    let up = fx.update_plan();
    assert_eq!(step_in(&up, 0, "go-settings")["state"], "unchanged", "{up}");
    let why = step_in(&up, 0, "go-state")["reason"]
        .as_str()
        .unwrap()
        .to_string();
    assert!(
        why.contains("settings reference .moai/") && why.contains("change or remove them yourself"),
        "{why}"
    );
}

/// Each spelling the Go tool wrote is Go-made; one character off is not.
#[test]
fn only_exact_go_values_are_cleaned() {
    let ok = "$CLAUDE_PROJECT_DIR/.moai/status_line.sh";
    for off in [
        json!({"type": "command", "command": format!("{ok}2")}),
        json!({"type": "command", "command": format!("bash {ok}")}),
        json!({"type": "command", "command": ok, "padding": 0}),
        json!({"type": "Command", "command": ok}),
        json!(ok),
    ] {
        let fx = base_fx();
        fx.manifest();
        fx.write(
            SETTINGS,
            &json!({"statusLine": off, "model": "x"}).to_string(),
        );
        let p = plan(&fx);
        assert_eq!(p["action"], json!("none"), "{off}: {p}");
        assert_eq!(keys_of(&p["kept"]), ["statusLine"], "{off}: {p}");
    }
    for spelling in [
        "$CLAUDE_PROJECT_DIR/.moai/status_line.sh",
        "${CLAUDE_PROJECT_DIR}/.moai/status_line.sh",
        "\"$CLAUDE_PROJECT_DIR\"/.moai/status_line.sh",
        "\"${CLAUDE_PROJECT_DIR}\"/.moai/status_line.sh",
    ] {
        let fx = base_fx();
        fx.manifest();
        fx.write(
            SETTINGS,
            &json!({"statusLine": {"type": "command", "command": spelling}}).to_string(),
        );
        let p = plan(&fx);
        assert_eq!(
            go_keys(&p),
            [("statusLine".to_string(), "replace".to_string())],
            "{spelling}"
        );
    }
}

/// D1 condition 3: the local file is never written, even when it holds Go references.
#[test]
fn the_local_settings_are_byte_identical_after_apply_and_revert() {
    let fx = shape_g();
    fx.write(LOCAL, LOCAL_WITH_GO_REFS);
    let local = read(&fx, LOCAL);
    let p = plan(&fx);
    // The step says the local file's keys are the user's to change, before the apply.
    let up = fx.update_plan();
    let why = step_in(&up, 0, "go-state")["reason"]
        .as_str()
        .unwrap()
        .to_string();
    assert!(
        why.contains(".claude/settings.local.json are your values"),
        "{why}"
    );
    let out = apply(&fx, &p);
    assert_eq!(out["ok"], json!(true), "{out}");
    assert_eq!(read(&fx, LOCAL), local);
    let id = out["result"]["receipt"].as_str().unwrap().to_string();
    assert_eq!(revert(&fx, &id)["ok"], json!(true));
    assert_eq!(read(&fx, LOCAL), local);
    // The local file still names `.moai/`, so go-state keeps waiting on it after the clean-up.
    let out = apply(&fx, &plan(&fx));
    assert_eq!(out["ok"], json!(true), "{out}");
    let up = fx.update_plan();
    let state = step_in(&up, 0, "go-state");
    assert_eq!(state["state"], "blocked", "{state}");
    assert!(
        state["reason"]
            .as_str()
            .unwrap()
            .contains(".claude/settings.local.json (plansDirectory, statusLine)"),
        "{state}"
    );
    assert_eq!(read(&fx, LOCAL), local);
}

#[test]
fn a_stale_digest_writes_nothing() {
    let fx = shape_g();
    let p = plan(&fx);
    let digest = p["plan_digest"].as_str().unwrap().to_string();
    let before = fx.tree();
    for wrong in ["", "sha256:00", "not-a-digest"] {
        let text = error_text(&apply_with(&fx, wrong, REF, &[]));
        assert!(text.contains("stale or altered plan"), "{wrong}: {text}");
        assert_eq!(fx.tree(), before, "{wrong}");
    }
    // The settings changed after the preview.
    let mut s = go_settings();
    s["model"] = json!("sonnet");
    fx.write(SETTINGS, &s.to_string());
    let before = fx.tree();
    let text = error_text(&apply_with(&fx, &digest, REF, &[]));
    assert!(text.contains("stale or altered plan"), "{text}");
    assert_eq!(fx.tree(), before);
    assert!(receipts(&fx).is_empty());
    // The launcher changed after the preview: the digest binds its state.
    let fresh = plan(&fx);
    put_launcher(&fx, "#!/bin/sh\necho 'something else'\n");
    let before = fx.tree();
    let text = error_text(&apply_with(
        &fx,
        fresh["plan_digest"].as_str().unwrap(),
        REF,
        &[],
    ));
    assert!(
        text.contains("stale or altered plan") || text.contains("held"),
        "{text}"
    );
    assert_eq!(fx.tree(), before);
}

#[test]
fn a_missing_permission_reference_writes_nothing() {
    let fx = shape_g();
    let p = plan(&fx);
    let digest = p["plan_digest"].as_str().unwrap();
    let before = fx.tree();
    for reference in ["", "placeholder", "null", "  "] {
        let text = error_text(&apply_with(&fx, digest, reference, &[]));
        assert!(
            text.contains("permission reference is required"),
            "{reference:?}: {text}"
        );
        assert_eq!(fx.tree(), before, "{reference:?}");
    }
    let text = error_text(&fx.run(&[
        "go-settings-revert",
        fx.p(),
        "--receipt",
        "go-settings-1-0",
        "--permission-reference",
        "placeholder",
    ]));
    assert!(text.contains("permission reference is required"), "{text}");
}

#[test]
fn revert_restores_the_exact_bytes_and_refuses_after_a_change() {
    let fx = shape_g();
    let original = read(&fx, SETTINGS);
    let p = plan(&fx);
    let out = apply(&fx, &p);
    let id = out["result"]["receipt"].as_str().unwrap().to_string();
    assert!(id.starts_with("go-settings-"), "{id}");
    assert_eq!(receipt_status(&fx, &id), "applied");
    // The receipt keeps the exact old bytes.
    assert_eq!(
        read(
            &fx,
            &format!(".moai-cli/receipts/{id}/before/settings.json")
        ),
        original
    );
    assert_ne!(read(&fx, SETTINGS), original);

    // The user edits the file after the apply: the revert refuses and writes nothing.
    let mut s = settings(&fx);
    s["model"] = json!("user-changed");
    fx.write(SETTINGS, &serde_json::to_string_pretty(&s).unwrap());
    let edited = read(&fx, SETTINGS);
    let before = fx.tree();
    let text = error_text(&revert(&fx, &id));
    assert!(text.contains("changed since the apply"), "{text}");
    assert_eq!(fx.tree(), before);
    assert_eq!(read(&fx, SETTINGS), edited);
    assert_eq!(receipt_status(&fx, &id), "applied");
    assert!(error_text(&revert(&fx, "../escape")).contains("unknown receipt id"));
    assert!(error_text(&revert(&fx, "go-hooks-1-0")).contains("unknown receipt id"));

    // A second clean-up on a fresh project reverts byte for byte, and only once.
    let fx = shape_h();
    let original = read(&fx, SETTINGS);
    let out = apply(&fx, &plan(&fx));
    let id = out["result"]["receipt"].as_str().unwrap().to_string();
    let r = revert(&fx, &id);
    assert_eq!(r["result"]["state"], json!("reverted"), "{r}");
    assert_eq!(read(&fx, SETTINGS), original);
    assert_eq!(receipt_status(&fx, &id), "reverted");
    let text = error_text(&revert(&fx, &id));
    assert!(text.contains("nothing to revert"), "{text}");
    assert_eq!(read(&fx, SETTINGS), original);
    // After the revert the same plan applies again.
    assert_eq!(plan(&fx)["action"], json!("replace"));
}

/// The launcher cannot be shown to be the Rust build: nothing is written, and the plan says why
/// and what to run (leader's option A as the fallback of D2 = C).
#[test]
fn a_launcher_that_is_not_shown_to_be_rust_holds_everything() {
    struct Case {
        name: &'static str,
        state: &'static str,
        setup: fn(&GoFx),
        env: &'static [(&'static str, &'static str)],
    }
    let cases = [
        Case {
            name: "missing",
            state: "missing",
            setup: |fx| fs::remove_file(fx.home.join(LAUNCHER)).unwrap(),
            env: &[],
        },
        Case {
            name: "not executable",
            state: "missing",
            setup: |fx| {
                fs::set_permissions(fx.home.join(LAUNCHER), fs::Permissions::from_mode(0o644))
                    .unwrap()
            },
            env: &[],
        },
        Case {
            name: "a folder",
            state: "missing",
            setup: |fx| {
                fs::remove_file(fx.home.join(LAUNCHER)).unwrap();
                fs::create_dir(fx.home.join(LAUNCHER)).unwrap();
            },
            env: &[],
        },
        Case {
            name: "a dangling link",
            state: "missing",
            setup: |fx| {
                fs::remove_file(fx.home.join(LAUNCHER)).unwrap();
                std::os::unix::fs::symlink("nowhere", fx.home.join(LAUNCHER)).unwrap();
            },
            env: &[],
        },
        Case {
            name: "another first line",
            state: "not_rust",
            setup: |fx| put_launcher(fx, "#!/bin/sh\necho 'moai 3.2.1 (go)'\n"),
            env: &[],
        },
        Case {
            name: "the right line second",
            state: "not_rust",
            setup: |fx| {
                put_launcher(
                    fx,
                    &format!("#!/bin/sh\necho 'Go moai'\necho '{RUST_HELP_LINE}'\n"),
                )
            },
            env: &[],
        },
        Case {
            name: "prints nothing",
            state: "unprobed",
            setup: |fx| put_launcher(fx, "#!/bin/sh\nexit 0\n"),
            env: &[],
        },
        Case {
            name: "hangs",
            state: "unprobed",
            setup: |fx| put_launcher(fx, "#!/bin/sh\nsleep 30\n"),
            env: &SHORT_PROBE,
        },
    ];
    for case in cases {
        let fx = shape_g();
        (case.setup)(&fx);
        let p = plan_env(&fx, case.env);
        let name = case.name;
        assert_eq!(p["action"], json!("held"), "{name}: {p}");
        assert_eq!(
            p["launcher"],
            json!({"path": SHOWN, "state": case.state}),
            "{name}: {p}"
        );
        // The keys that would change are still named; the reason and the next command are given.
        assert_eq!(go_keys(&p), the_three(), "{name}: {p}");
        assert!(p["reason"].as_str().unwrap().contains(SHOWN), "{name}: {p}");
        let next = p["next"].as_str().unwrap();
        assert!(
            next.contains(SHOWN)
                && next.contains("statusline --help")
                && next.contains(RUST_HELP_LINE)
                && next.contains("nothing was written"),
            "{name}: {p}"
        );
        // An apply with the held plan's digest writes nothing at all.
        let before = fx.tree();
        let text = error_text(&apply_with(
            &fx,
            p["plan_digest"].as_str().unwrap(),
            REF,
            case.env,
        ));
        assert!(text.contains("the plan is held"), "{name}: {text}");
        assert_eq!(fx.tree(), before, "{name}");
        assert!(receipts(&fx).is_empty(), "{name}");

        // The update plan: a decision without a command; go-state says the step could not clear.
        let up = fx.update_plan_env(case.env);
        let step = step_in(&up, 0, "go-settings");
        assert_eq!(step["state"], "needs_decision", "{name}: {step}");
        assert_eq!(step["apply"], Value::Null, "{name}: {step}");
        assert!(
            step["reason"].as_str().unwrap().contains(SHOWN),
            "{name}: {step}"
        );
        let state = step_in(&up, 0, "go-state");
        assert_eq!(state["state"], "blocked", "{name}: {state}");
        assert!(
            state["reason"]
                .as_str()
                .unwrap()
                .contains("the go-settings step above could not clear the Go-made keys"),
            "{name}: {state}"
        );
        assert_eq!(settings(&fx), go_settings(), "{name}");
    }
}

/// A launcher that is a link to an executable regular file is shown like the file itself.
#[test]
fn a_linked_launcher_that_runs_is_accepted() {
    let fx = shape_g();
    let real = fx.home.join("real-launcher");
    common::write_executable(&real, &rust_launcher());
    fs::remove_file(fx.home.join(LAUNCHER)).unwrap();
    std::os::unix::fs::symlink(&real, fx.home.join(LAUNCHER)).unwrap();
    let p = plan(&fx);
    assert_eq!(p["action"], json!("replace"), "{p}");
    assert_eq!(p["launcher"]["state"], json!("rust"), "{p}");
}

/// The launcher is only looked at when a Go-made `statusLine` is to be replaced.
#[test]
fn no_launcher_is_needed_to_remove_the_other_keys() {
    let fx = base_fx();
    fx.manifest();
    fs::remove_file(fx.home.join(LAUNCHER)).unwrap();
    let mut s = go_settings();
    s.as_object_mut().unwrap().remove("statusLine");
    fx.write(SETTINGS, &s.to_string());
    let p = plan(&fx);
    assert_eq!(p["action"], json!("replace"), "{p}");
    assert_eq!(p["launcher"]["state"], json!("not_needed"), "{p}");
    let out = apply(&fx, &p);
    assert_eq!(out["ok"], json!(true), "{out}");
    let after = settings(&fx);
    assert!(after.get("statusLine").is_none(), "{after}");
    assert!(after.get("plansDirectory").is_none(), "{after}");
}

/// Nothing Go-made, no file, or a project that is not Go-generation: `none`, and an apply with
/// that digest changes nothing.
#[test]
fn nothing_to_clean_is_none() {
    // A settings file without a Go-made key.
    let fx = base_fx();
    fx.manifest();
    fx.write(SETTINGS, "{\"model\":\"opus\"}\n");
    let p = plan(&fx);
    assert_eq!(p["action"], json!("none"), "{p}");
    assert_eq!(go_keys(&p), Vec::<(String, String)>::new());
    let before = fx.tree();
    let out = apply(&fx, &p);
    assert_eq!(out["result"]["state"], json!("unchanged"), "{out}");
    assert_eq!(fx.tree(), before);
    let up = fx.update_plan();
    assert_eq!(step_in(&up, 0, "go-settings")["state"], "unchanged", "{up}");
    assert_eq!(step_in(&up, 0, "go-settings")["apply"], Value::Null);

    // No settings file at all.
    let fx = base_fx();
    fx.manifest();
    let p = plan(&fx);
    assert_eq!(p["action"], json!("none"), "{p}");
    assert_eq!(p["settings_present"], json!(false), "{p}");

    // The same keys in a project that is not Go-generation: the folder is this generation's own.
    let fx = GoFx::new();
    fx.write(SETTINGS, &go_settings().to_string());
    let p = plan(&fx);
    assert_eq!(p["action"], json!("none"), "{p}");
    assert_ne!(p["state_root"], json!("go"), "{p}");

    // Broken settings are refused, not read as empty.
    let fx = base_fx();
    fx.manifest();
    fx.write(SETTINGS, "[1, 2]");
    let text = error_text(&fx.run(&["go-settings-plan", fx.p()]));
    assert!(text.contains("not a JSON object"), "{text}");
    let up = fx.update_plan();
    assert_eq!(step_in(&up, 0, "go-settings")["state"], "blocked", "{up}");
}

/// A settings file that is a link is refused, never followed or replaced.
#[test]
fn a_linked_settings_file_is_refused() {
    let fx = base_fx();
    fx.manifest();
    fx.write("elsewhere.json", &go_settings().to_string());
    fx.link(SETTINGS, "../elsewhere.json");
    let before = fx.tree();
    let text = error_text(&fx.run(&["go-settings-plan", fx.p()]));
    assert!(text.contains("symlink"), "{text}");
    assert_eq!(fx.tree(), before);
}

/// A failed replace puts everything back: the file keeps its bytes, the receipt says so, and
/// there is nothing to revert.
#[test]
fn a_failed_apply_leaves_the_settings_as_they_were() {
    let fx = shape_g();
    let original = read(&fx, SETTINGS);
    let p = plan(&fx);
    let claude = fx.root.join(".claude");
    fs::set_permissions(&claude, fs::Permissions::from_mode(0o555)).unwrap();
    let out = apply(&fx, &p);
    fs::set_permissions(&claude, fs::Permissions::from_mode(0o755)).unwrap();
    let text = error_text(&out);
    assert!(text.contains("everything was put back"), "{text}");
    assert_eq!(read(&fx, SETTINGS), original);
    let ids = receipts(&fx);
    assert_eq!(ids.len(), 1, "{ids:?}");
    assert_eq!(receipt_status(&fx, &ids[0]), "rolled_back");
    let text = error_text(&revert(&fx, &ids[0]));
    assert!(text.contains("nothing to revert"), "{text}");
    // With the folder writable again the same plan applies.
    let out = apply(&fx, &plan(&fx));
    assert_eq!(out["ok"], json!(true), "{out}");
}

/// An interrupted apply leaves a pending receipt that a revert refuses with a hand-recovery note.
#[test]
fn a_pending_receipt_is_not_reverted() {
    let fx = shape_g();
    let out = apply(&fx, &plan(&fx));
    let id = out["result"]["receipt"].as_str().unwrap().to_string();
    let path = format!(".moai-cli/receipts/{id}/receipt.json");
    let mut r: Value = serde_json::from_slice(&read(&fx, &path)).unwrap();
    r["status"] = json!("pending");
    fx.write(&path, &r.to_string());
    let before = fx.tree();
    let text = error_text(&revert(&fx, &id));
    assert!(
        text.contains("pending") && text.contains("by hand"),
        "{text}"
    );
    assert_eq!(fx.tree(), before);
}

fn plans_fixture(files: usize, dirs: usize) -> tempfile::TempDir {
    let tmp = tempfile::tempdir().unwrap();
    let plans = tmp.path().join(".moai/plans");
    fs::create_dir_all(&plans).unwrap();
    for i in 0..files {
        fs::write(plans.join(format!("f{i}.md")), "x").unwrap();
    }
    for i in 0..dirs {
        fs::create_dir(plans.join(format!("d{i}"))).unwrap();
    }
    tmp
}

/// The `.moai/plans` count looks at no more than `PLANS_WALK_LIMIT` entries: N is counted whole,
/// N + 1 is reported as more than the limit.
#[test]
fn the_plans_count_is_bounded_at_the_limit_and_one_past_it() {
    assert_eq!(PLANS_WALK_LIMIT, 1000);
    let at = plans_fixture(PLANS_WALK_LIMIT, 0);
    let c = count_plans(at.path());
    assert_eq!(
        (c.exists, c.files, c.more_than_limit),
        (true, PLANS_WALK_LIMIT, false)
    );
    let past = plans_fixture(PLANS_WALK_LIMIT + 1, 0);
    let c = count_plans(past.path());
    assert_eq!(
        (c.exists, c.files, c.more_than_limit),
        (true, PLANS_WALK_LIMIT, true)
    );
    // Folders count as entries too: 999 files and one folder is N, 1000 and one is N + 1.
    let at = plans_fixture(PLANS_WALK_LIMIT - 1, 1);
    let c = count_plans(at.path());
    assert_eq!((c.files, c.more_than_limit), (PLANS_WALK_LIMIT - 1, false));
    let past = plans_fixture(PLANS_WALK_LIMIT, 1);
    assert!(count_plans(past.path()).more_than_limit);
    // The empty case and an absent folder.
    let empty = plans_fixture(0, 0);
    let c = count_plans(empty.path());
    assert_eq!((c.exists, c.files, c.more_than_limit), (true, 0, false));
    let none = tempfile::tempdir().unwrap();
    assert!(!count_plans(none.path()).exists);
}

/// Links are neither followed nor counted, and a linked `.moai` or `.moai/plans` is no folder.
#[test]
fn the_plans_count_never_follows_a_link() {
    let tmp = plans_fixture(2, 1);
    let plans = tmp.path().join(".moai/plans");
    let outside = tempfile::tempdir().unwrap();
    fs::write(outside.path().join("secret.md"), "x").unwrap();
    std::os::unix::fs::symlink(outside.path(), plans.join("link-to-folder")).unwrap();
    std::os::unix::fs::symlink(outside.path().join("secret.md"), plans.join("link-to-file"))
        .unwrap();
    fs::write(plans.join("d0/inner.md"), "x").unwrap();
    let c = count_plans(tmp.path());
    assert_eq!((c.files, c.more_than_limit), (3, false));

    let linked = tempfile::tempdir().unwrap();
    fs::create_dir_all(linked.path().join(".moai")).unwrap();
    std::os::unix::fs::symlink(outside.path(), linked.path().join(".moai/plans")).unwrap();
    assert!(!count_plans(linked.path()).exists);
    let moai_link = tempfile::tempdir().unwrap();
    let real = plans_fixture(3, 0);
    std::os::unix::fs::symlink(real.path().join(".moai"), moai_link.path().join(".moai")).unwrap();
    assert!(!count_plans(moai_link.path()).exists);
}

/// A big plans folder is reported as more than the limit, in the plan, and still not deleted.
#[test]
fn a_big_plans_folder_is_reported_as_more_than_the_limit() {
    let fx = shape_g();
    for i in 0..=PLANS_WALK_LIMIT {
        fx.write(&format!(".moai/plans/many/p{i}.md"), "x");
    }
    let p = plan(&fx);
    assert_eq!(p["plans_files"]["more_than_limit"], json!(true), "{p}");
    assert_eq!(p["plans_files"]["limit"], json!(PLANS_WALK_LIMIT), "{p}");
    let before: Vec<_> = (0..=PLANS_WALK_LIMIT)
        .map(|i| fx.root.join(format!(".moai/plans/many/p{i}.md")).is_file())
        .collect();
    assert_eq!(apply(&fx, &p)["ok"], json!(true));
    let after: Vec<_> = (0..=PLANS_WALK_LIMIT)
        .map(|i| fx.root.join(format!(".moai/plans/many/p{i}.md")).is_file())
        .collect();
    assert_eq!(before, after, "no plans file is deleted");
    assert!(Path::new(&fx.root.join(".moai/plans/a.md")).is_file());
}
