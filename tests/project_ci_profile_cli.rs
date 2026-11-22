//! The two profile keys of the CI design (SPEC-20261122-project-ci-design, slice A): `ci`, the
//! saved answers of the plain questions, and `checks`, the ordered commands the project
//! declares. Both are additive keys that go through `project-plan` and `project-apply` like
//! every other key, through the binary in temporary folders only.
//!
//! The declared list is the one source of the generated workflow's run steps and of the
//! project's local check (REQ-4, REQ-5), so its order is kept as declared, each command is one
//! line a person confirmed, and nothing that could carry event data or a secret into a
//! workflow is accepted.

mod common;
use common::home::TempHome;
use serde_json::{json, Value};
use std::{
    fs,
    io::Write,
    path::Path,
    process::{Output, Stdio},
};
use tempfile::TempDir;

/// Runs the binary with a temporary HOME of its own, so no test reads or writes the real one.
fn run(args: &[&str], stdin: &[u8]) -> Output {
    let home = TempHome::new();
    let mut child = home
        .command(args)
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
    child.wait_with_output().unwrap()
}

fn text(out: &Output) -> String {
    format!(
        "{}{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    )
}

fn ok(out: &Output) -> Value {
    assert!(out.status.success(), "{}", text(out));
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn p(root: &TempDir) -> &str {
    root.path().to_str().unwrap()
}

fn base() -> Value {
    json!({
        "schema_version": 2, "project_id": "demo", "purpose": "Build a small editor",
        "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude"]
    })
}

fn ci() -> Value {
    json!({
        "kind": "cli",
        "systems": ["linux", "windows"],
        "deploy": {"enabled": false},
        "visibility": "public",
        "strict": "report"
    })
}

/// Declared in an order that is not alphabetical: the list is kept as declared, never sorted.
fn checks() -> Value {
    json!([
        {"id": "lint", "command": "cargo clippy --all-targets -- -D warnings"},
        {"id": "build", "command": "cargo build"},
        {"id": "test", "command": "cargo test"}
    ])
}

fn with(key: &str, value: Value) -> Value {
    let mut profile = base();
    profile[key] = value;
    profile
}

fn plan(root: &TempDir, profile: &Value) -> Output {
    run(&["project-plan", p(root)], profile.to_string().as_bytes())
}

fn apply(root: &TempDir, plan: &Value) -> Output {
    run(
        &[
            "project-apply",
            p(root),
            "--plan-digest",
            plan["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "fixture-approved",
        ],
        plan.to_string().as_bytes(),
    )
}

fn plan_apply(root: &TempDir, profile: &Value) -> Value {
    let planned = ok(&plan(root, profile));
    ok(&apply(root, &planned));
    planned
}

/// The profile file the apply wrote: a project with no state gets `.moai/` (#2696).
fn profile_file(root: &TempDir) -> std::path::PathBuf {
    for dir in [".moai", ".moai-cli"] {
        let path = root.path().join(dir).join("project.yaml");
        if path.is_file() {
            return path;
        }
    }
    panic!("no profile file was written");
}

fn written(root: &TempDir) -> Value {
    serde_json::from_slice(&fs::read(profile_file(root)).unwrap()).unwrap()
}

/// A plan that is refused fails the command and says what kind of refusal it is without
/// echoing the offending value back.
fn refused(profile: &Value, why: &str) {
    let root = TempDir::new().unwrap();
    let out = plan(&root, profile);
    assert!(!out.status.success(), "{why}: the plan was accepted");
    let shown = text(&out);
    assert!(
        shown.contains("invalid") || shown.contains("unknown field"),
        "{why}: {shown}"
    );
    assert!(
        !root.path().join(".moai").join("project.yaml").exists(),
        "{why}: a refused plan wrote a file"
    );
}

#[test]
fn saved_answers_and_declared_checks_round_trip_and_keep_their_order() {
    let root = TempDir::new().unwrap();
    let mut profile = base();
    profile["ci"] = ci();
    profile["checks"] = checks();
    plan_apply(&root, &profile);
    let file = written(&root);
    assert_eq!(file["ci"], ci());
    assert_eq!(file["checks"], checks());
    let ids: Vec<&str> = file["checks"]
        .as_array()
        .unwrap()
        .iter()
        .map(|c| c["id"].as_str().unwrap())
        .collect();
    assert_eq!(ids, ["lint", "build", "test"], "declared order, not sorted");

    // The same input a second time changes nothing.
    let again = ok(&plan(&root, &profile));
    assert_eq!(again["action"], "unchanged", "{again}");

    // project-detect hands the saved keys back in `profile` for the next plan.
    let detected = ok(&run(&["project-detect", p(&root)], b""));
    assert_eq!(detected["profile"]["ci"], ci(), "{detected}");
    assert_eq!(detected["profile"]["checks"], checks(), "{detected}");
}

#[test]
fn a_profile_without_the_keys_plans_as_before_and_gains_them_as_added_keys() {
    let root = TempDir::new().unwrap();
    let first = ok(&plan(&root, &base()));
    let listed = |plan: &Value| plan["delta"].to_string();
    assert!(!listed(&first).contains("\"ci\""), "{first}");
    assert!(!listed(&first).contains("\"checks\""), "{first}");
    ok(&apply(&root, &first));

    let mut profile = base();
    profile["ci"] = ci();
    profile["checks"] = checks();
    let second = ok(&plan(&root, &profile));
    assert_eq!(second["action"], "update", "{second}");
    assert_eq!(
        second["delta"]["added"],
        json!(["checks", "ci"]),
        "{second}"
    );
}

#[test]
fn a_hand_edit_of_the_declared_checks_is_kept_and_a_different_input_is_a_conflict() {
    let root = TempDir::new().unwrap();
    plan_apply(&root, &with("checks", checks()));
    // The user edits the list by hand after the apply.
    let path = profile_file(&root);
    let mut file = written(&root);
    file["checks"] = json!([{"id": "test", "command": "make test"}]);
    fs::write(&path, serde_json::to_vec_pretty(&file).unwrap()).unwrap();

    // The input still carries the applied list: the edit is kept and the plan changes nothing.
    let same = ok(&plan(&root, &with("checks", checks())));
    assert_eq!(same["action"], "unchanged", "{same}");
    assert_eq!(same["drift"], json!(["checks"]), "{same}");

    // A third list is neither the baseline nor the user's edit: nothing is overwritten.
    let other = json!([{"id": "test", "command": "cargo test --all"}]);
    let conflict = ok(&plan(&root, &with("checks", other)));
    assert_eq!(conflict["action"], "conflict", "{conflict}");
    assert_eq!(
        conflict["delta"]["conflict"],
        json!(["checks"]),
        "{conflict}"
    );
    let out = apply(&root, &conflict);
    assert!(!out.status.success(), "{}", text(&out));
    assert_eq!(
        written(&root)["checks"],
        json!([{"id": "test", "command": "make test"}]),
        "the edited file was overwritten"
    );
}

#[test]
fn an_answer_set_with_a_wrong_value_is_refused() {
    let mut cases: Vec<(&str, Value)> = Vec::new();
    let mutate = |f: &dyn Fn(&mut Value)| {
        let mut v = ci();
        f(&mut v);
        v
    };
    cases.push(("unknown kind", mutate(&|v| v["kind"] = json!("game"))));
    cases.push(("kind is lower-case", mutate(&|v| v["kind"] = json!("CLI"))));
    cases.push((
        "unknown system",
        mutate(&|v| v["systems"] = json!(["linux", "beos"])),
    ));
    cases.push(("no systems", mutate(&|v| v["systems"] = json!([]))));
    cases.push((
        "linux missing",
        mutate(&|v| v["systems"] = json!(["windows"])),
    ));
    cases.push((
        "systems out of order",
        mutate(&|v| v["systems"] = json!(["windows", "linux"])),
    ));
    cases.push((
        "systems twice",
        mutate(&|v| v["systems"] = json!(["linux", "linux"])),
    ));
    cases.push((
        "macos without the price yes",
        mutate(&|v| v["systems"] = json!(["linux", "macos"])),
    ));
    cases.push((
        "macos with a no",
        mutate(&|v| {
            v["systems"] = json!(["linux", "macos"]);
            v["macos_price_ack"] = json!(false);
        }),
    ));
    cases.push((
        "deploy yes without a place",
        mutate(&|v| v["deploy"] = json!({"enabled": true})),
    ));
    cases.push((
        "deploy no with a place",
        mutate(&|v| v["deploy"] = json!({"enabled": false, "target": "production"})),
    ));
    cases.push((
        "deploy place with a space",
        mutate(&|v| v["deploy"] = json!({"enabled": true, "target": "my site"})),
    ));
    cases.push((
        "deploy place is a path",
        mutate(&|v| v["deploy"] = json!({"enabled": true, "target": "a/b"})),
    ));
    cases.push((
        "deploy missing",
        mutate(&|v| {
            v.as_object_mut().unwrap().remove("deploy");
        }),
    ));
    cases.push((
        "unknown visibility",
        mutate(&|v| v["visibility"] = json!("internal")),
    ));
    cases.push((
        "a budget on a public repository",
        mutate(&|v| v["monthly_budget_usd"] = json!(10)),
    ));
    cases.push((
        "a negative budget",
        mutate(&|v| {
            v["visibility"] = json!("private");
            v["monthly_budget_usd"] = json!(-1);
        }),
    ));
    cases.push((
        "a budget as text",
        mutate(&|v| {
            v["visibility"] = json!("private");
            v["monthly_budget_usd"] = json!("10");
        }),
    ));
    cases.push((
        "unknown strictness",
        mutate(&|v| v["strict"] = json!("strict")),
    ));
    cases.push(("unknown field", mutate(&|v| v["extra"] = json!(1))));
    cases.push(("an answer set that is a list", json!(["cli"])));
    for (why, value) in cases {
        refused(&with("ci", value), why);
    }
}

#[test]
fn a_declared_check_with_a_wrong_value_is_refused() {
    let one = |id: &str, command: &str| json!([{"id": id, "command": command}]);
    let long = "x".repeat(201);
    let cases: Vec<(&str, Value)> = vec![
        ("an empty list", json!([])),
        ("not a list", json!({"id": "test", "command": "cargo test"})),
        ("an empty command", one("test", "")),
        ("a blank command", one("test", "   ")),
        ("leading space", one("test", " cargo test")),
        ("trailing space", one("test", "cargo test ")),
        ("a new line", one("test", "cargo test\nrm -rf x")),
        ("a tab", one("test", "cargo\ttest")),
        ("a carriage return", one("test", "cargo test\r")),
        (
            "an expression in the command",
            one("test", "echo ${{ github.event.pull_request.title }}"),
        ),
        (
            "an expression, spaced out",
            one("test", "echo ${{secrets.NAME}}"),
        ),
        ("a key shape", one("test", "deploy --key sk-abc123")),
        ("a token assignment", one("test", "run token=abc")),
        ("201 bytes", one("test", &long)),
        ("a missing command", json!([{"id": "test"}])),
        ("a missing id", json!([{"command": "cargo test"}])),
        ("an id with a space", one("my test", "cargo test")),
        ("an id with a slash", one("a/b", "cargo test")),
        ("an empty id", one("", "cargo test")),
        ("an id of 41 bytes", one(&"i".repeat(41), "cargo test")),
        (
            "the same id twice",
            json!([
                {"id": "test", "command": "cargo test"},
                {"id": "test", "command": "cargo test --all"}
            ]),
        ),
        (
            "an unknown field",
            json!([{"id": "test", "command": "cargo test", "x": 1}]),
        ),
        (
            "a command that is a list",
            json!([{"id": "test", "command": ["cargo", "test"]}]),
        ),
        (
            "17 checks",
            Value::Array(
                (0..17)
                    .map(|i| json!({"id": format!("c{i}"), "command": "true"}))
                    .collect(),
            ),
        ),
    ];
    for (why, value) in cases {
        refused(&with("checks", value), why);
    }
}

/// N and N+1 of every stated limit: the value at the limit passes.
#[test]
fn the_limits_are_met_exactly_at_the_boundary() {
    let root = TempDir::new().unwrap();
    let sixteen = Value::Array(
        (0..16)
            .map(|i| json!({"id": format!("c{i}"), "command": "true"}))
            .collect(),
    );
    let mut profile = base();
    profile["checks"] = sixteen.clone();
    ok(&plan(&root, &profile));

    let edge = json!([
        {"id": "i".repeat(40), "command": "x".repeat(200)},
        {"id": "p-q_9", "command": "go build ./..."},
        {"id": "gofmt", "command": "test -z \"$(gofmt -l .)\""}
    ]);
    ok(&plan(&root, &with("checks", edge)));

    let mut answers = ci();
    answers["visibility"] = json!("private");
    answers["monthly_budget_usd"] = json!(10_000);
    answers["deploy"] = json!({"enabled": true, "target": "p".repeat(40)});
    answers["systems"] = json!(["linux", "macos", "windows"]);
    answers["macos_price_ack"] = json!(true);
    answers["dependabot"] = json!(true);
    ok(&plan(&root, &with("ci", answers.clone())));

    let mut over = answers.clone();
    over["monthly_budget_usd"] = json!(10_001);
    refused(&with("ci", over), "a budget of 10001");
    let mut long_place = answers;
    long_place["deploy"] = json!({"enabled": true, "target": "p".repeat(41)});
    refused(&with("ci", long_place), "a deploy place of 41 bytes");
    let mut zero = ci();
    zero["visibility"] = json!("private");
    zero["monthly_budget_usd"] = json!(0);
    ok(&plan(&root, &with("ci", zero)));
}

/// A profile edited by hand is read with the same rules: a bad key stops every reader, a good
/// one passes, and nothing is written by the read.
#[test]
fn a_hand_edited_profile_is_read_with_the_same_rules() {
    let root = TempDir::new().unwrap();
    let dir = root.path().join(".moai-cli");
    fs::create_dir_all(&dir).unwrap();
    let path = dir.join("project.yaml");

    let mut file = base();
    file["ci"] = ci();
    file["checks"] = checks();
    fs::write(&path, serde_json::to_vec_pretty(&file).unwrap()).unwrap();
    let before = fs::read(&path).unwrap();
    let detected = ok(&run(&["project-detect", p(&root)], b""));
    assert_eq!(detected["profile"]["ci"], ci());
    assert_eq!(fs::read(&path).unwrap(), before);

    for bad in [
        ("ci", json!({"kind": "game"})),
        ("checks", json!([{"id": "t", "command": "echo ${{ x }}"}])),
        ("checks", json!([])),
    ] {
        let mut file = base();
        file[bad.0] = bad.1;
        fs::write(&path, serde_json::to_vec_pretty(&file).unwrap()).unwrap();
        let out = run(&["project-detect", p(&root)], b"");
        assert!(!out.status.success(), "{}", text(&out));
    }
}

/// A refusal says what is wrong with the key, never what the value was.
#[test]
fn a_refusal_does_not_echo_the_offending_value() {
    let root = TempDir::new().unwrap();
    let secret = "deploy --key sk-fixture-not-a-real-key-1234";
    let out = plan(
        &root,
        &with("checks", json!([{"id": "d", "command": secret}])),
    );
    assert!(!out.status.success());
    assert!(!text(&out).contains("sk-fixture"), "{}", text(&out));
    assert!(Path::new(p(&root)).exists());
}

/// Audit N2: the same holds for a refused `ci` value; no value of the answers is repeated.
#[test]
fn a_refusal_of_the_answers_does_not_echo_the_offending_value() {
    let marker = "zz-fixture-marker-9183";
    let mut cases = Vec::new();
    for (field, value) in [
        ("kind", json!(marker)),
        ("visibility", json!(marker)),
        ("strict", json!(marker)),
        ("systems", json!(["linux", marker])),
        (
            "deploy",
            json!({"enabled": true, "target": format!("{marker} x")}),
        ),
    ] {
        let mut answers = ci();
        answers[field] = value;
        cases.push(answers);
    }
    for answers in cases {
        let root = TempDir::new().unwrap();
        let out = plan(&root, &with("ci", answers.clone()));
        assert!(!out.status.success(), "{answers}");
        assert!(!text(&out).contains(marker), "{answers}: {}", text(&out));
    }
}

/// Audit N1: a check id and a deploy place are names, but a name that has the shape of a key
/// is refused like any other profile text, and the value is not repeated.
#[test]
fn a_check_id_or_a_deploy_place_in_the_shape_of_a_key_is_refused() {
    for name in ["sk-live-4242", "ghp_abcdef1234", "github_pat_x1"] {
        let root = TempDir::new().unwrap();
        let out = plan(
            &root,
            &with("checks", json!([{"id": name, "command": "true"}])),
        );
        assert!(!out.status.success(), "id {name} was accepted");
        assert!(!text(&out).contains(name), "{}", text(&out));

        let mut answers = ci();
        answers["deploy"] = json!({"enabled": true, "target": name});
        let out = plan(&root, &with("ci", answers));
        assert!(!out.status.success(), "deploy place {name} was accepted");
        assert!(!text(&out).contains(name), "{}", text(&out));
    }
    // A word that only contains the letters is a plain name.
    let root = TempDir::new().unwrap();
    let mut answers = ci();
    answers["deploy"] = json!({"enabled": true, "target": "task-based"});
    ok(&plan(&root, &with("ci", answers)));
    ok(&plan(
        &root,
        &with("checks", json!([{"id": "risk-free", "command": "true"}])),
    ));
}
