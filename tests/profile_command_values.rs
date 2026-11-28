//! SPEC-20261126-profile-command-values: a decision value may hold a command with "/", every
//! other profile text keeps the earlier screen, and a refusal names the field without repeating
//! a value. Every command runs under a temporary HOME against a temporary project.
mod common;

use common::cli::{ok_result, profile, tree_state};
use common::feed_stdin;
use common::home::TempHome;
use serde_json::{json, Value};
use std::process::{Output, Stdio};
use tempfile::TempDir;

/// The refusal that names no field.
const PLAIN: &str = "invalid or nonportable profile; no values were logged";

/// The refusal that names `field`.
fn naming(field: &str) -> String {
    format!("invalid or nonportable profile field {field}; no values were logged")
}

fn decision(id: &str, value: &str) -> Value {
    json!({"id":id,"topic":"tools","value":value,"source":"user","scope":"project"})
}

/// A valid plan input with these decision records.
fn with_decisions(records: Vec<Value>) -> Value {
    let mut input = profile(None);
    input["decisions"] = Value::Array(records);
    input
}

/// A valid plan input with `key` set to `value`.
fn with_key(key: &str, value: Value) -> Value {
    let mut input = profile(None);
    input[key] = value;
    input
}

/// Runs `moai ARGS` under `home` with `input` on stdin.
fn run(home: &TempHome, args: &[&str], input: &Value) -> Output {
    let mut child = home
        .command(args)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    feed_stdin(&mut child, input.to_string().as_bytes());
    child.wait_with_output().unwrap()
}

fn project_plan(home: &TempHome, project: &TempDir, input: &Value) -> Output {
    run(
        home,
        &["project-plan", project.path().to_str().unwrap()],
        input,
    )
}

/// `project-plan` refuses `input` with `message` and the kind `invalid_input`, repeats none of
/// `hidden` on either stream (as typed or as JSON escapes it) and leaves the project untouched.
fn assert_refused(input: &Value, message: &str, hidden: &[&str]) {
    let home = TempHome::new();
    let project = TempDir::new().unwrap();
    let before = tree_state(project.path());
    let out = project_plan(&home, &project, input);
    let stdout = String::from_utf8_lossy(&out.stdout);
    let stderr = String::from_utf8_lossy(&out.stderr);
    assert_eq!(out.status.code(), Some(1), "{input}: {stdout}");
    let failure: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(failure["ok"], json!(false), "{input}");
    assert_eq!(failure["error"]["message"], json!(message), "{input}");
    assert_eq!(failure["error"]["kind"], json!("invalid_input"), "{input}");
    for value in hidden {
        let quoted = serde_json::to_string(value).unwrap();
        for form in [*value, &quoted[1..quoted.len() - 1]] {
            assert!(!stdout.contains(form), "stdout repeats {form:?}");
            assert!(!stderr.contains(form), "stderr repeats {form:?}");
        }
    }
    assert_eq!(tree_state(project.path()), before, "{input}");
}

#[test]
fn accept_commands_survive_plan_apply_and_readback() {
    let home = TempHome::new();
    let project = TempDir::new().unwrap();
    let commands = [
        ("test_command", "go test ./..."),
        ("build_command", "./gradlew build"),
    ];
    let input = with_decisions(
        commands
            .iter()
            .map(|(id, value)| decision(id, value))
            .collect(),
    );
    let out = project_plan(&home, &project, &input);
    let plan = ok_result(&out);
    assert_eq!(plan["action"], "create");
    let applied = ok_result(&run(
        &home,
        &[
            "project-apply",
            project.path().to_str().unwrap(),
            "--plan-digest",
            plan["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "fixture-approved",
        ],
        &plan,
    ));
    assert_eq!(applied["state"], "created");
    let (saved, _) = moai_cli::read_profile(project.path()).unwrap();
    let saved: Vec<(&str, &str)> = saved
        .decisions
        .iter()
        .map(|d| (d.id.as_str(), d.value.as_str()))
        .collect();
    assert_eq!(saved, commands);
    let again = ok_result(&project_plan(&home, &project, &input));
    assert_eq!(again["action"], "unchanged");
}

#[test]
fn accept_ordinary_command_values() {
    for value in [
        "cargo test",
        "go test ./...",
        "./gradlew build",
        "test -z \"$(gofmt -l .)\"",
        "npm run build",
        "ruff check .",
        "../scripts/run.sh",
        // A "/" after each kind of name character.
        "src/*/mod_/a-/b9/c//d",
        // A "-" inside a word starts no flag, and a flag may name a relative path after "=".
        "run pre-/post --out=./dist -- -v",
        // A "~" that leads to no "/", or that follows a name character.
        "git diff HEAD~1..HEAD",
        "echo ~ done/ok",
        "echo ~example",
        "echo backup~name/file",
    ] {
        let home = TempHome::new();
        let project = TempDir::new().unwrap();
        let input = with_decisions(vec![decision("test_command", value)]);
        let out = project_plan(&home, &project, &input);
        assert!(
            out.status.success(),
            "{value}: {}",
            String::from_utf8_lossy(&out.stdout)
        );
        assert_eq!(
            ok_result(&out)["profile"]["decisions"][0]["value"],
            json!(value)
        );
    }
}

#[test]
fn accept_values_up_to_the_byte_limit() {
    // 256 bytes pass, as ASCII and as text that ends in three-byte characters.
    for value in ["x".repeat(256), format!("{}x", "가".repeat(85))] {
        assert_eq!(value.len(), 256);
        let home = TempHome::new();
        let project = TempDir::new().unwrap();
        let input = with_decisions(vec![decision("test_command", &value)]);
        let out = project_plan(&home, &project, &input);
        assert!(
            out.status.success(),
            "{}",
            String::from_utf8_lossy(&out.stdout)
        );
    }
}

#[test]
fn refused_decision_values_are_named_and_write_nothing() {
    let long = "x".repeat(257);
    let long_wide = "가".repeat(86);
    let message = naming("decisions[test_command].value");
    for value in [
        "",
        "   ",
        "/Users/example/x",
        "go test /Users/example/x",
        "~/x",
        "cd ~/x",
        "a~/x",
        "~example/x",
        "run ~a.b+c-d_e/x",
        "-I~example/x",
        "-~example/x",
        "$HOME/x",
        "$Home/x",
        "x$HOME/x",
        "${HOME}/x",
        "x${home}/x",
        "${HOME:-}/x",
        "$(echo ~)/x",
        "-/x",
        "-I/usr/include",
        "cc (-I/x)",
        "-isystem/Users/example",
        "--out=/x",
        "a:/usr/bin",
        "file:///x",
        "2>/dev/null",
        "go test \u{3164}/Users/example",
        "C:\\x",
        "a\\b",
        "C:/x",
        "'/x'",
        "\"/x\"",
        "(/x)",
        "cargo test\u{7}",
        "cargo test\nrm x",
        long.as_str(),
        long_wide.as_str(),
    ] {
        let hidden: &[&str] = if value.trim().is_empty() {
            &[]
        } else {
            &[value]
        };
        assert_refused(
            &with_decisions(vec![decision("test_command", value)]),
            &message,
            hidden,
        );
    }
}

#[test]
fn refused_secret_values_are_named_and_write_nothing() {
    // Built at run time, so no secret shape is written in this file.
    let secrets = [
        format!("{}{}", "pass", "word=fixture-only"),
        format!("{}{}", "s", "k-fixture-only-value"),
        format!("{}{}", "gh", "p_fixtureonly"),
        format!("{}{}{}", "AK", "IA", "A1".repeat(8)),
        // Known to the outgoing redaction screen only.
        format!("{}{}", "xo", "xb-0123456789abcdef"),
    ];
    let message = naming("decisions[test_command].value");
    for secret in &secrets {
        for value in [secret.clone(), format!("./run {secret}")] {
            assert_refused(
                &with_decisions(vec![decision("test_command", &value)]),
                &message,
                &[&value, secret],
            );
        }
    }
}

/// `project-plan` refuses each value as a test command, names the field and repeats neither the
/// value nor `hidden`.
fn assert_value_refused(value: &str, hidden: &[&str]) {
    let mut unseen = vec![value];
    unseen.extend_from_slice(hidden);
    assert_refused(
        &with_decisions(vec![decision("test_command", value)]),
        &naming("decisions[test_command].value"),
        &unseen,
    );
}

/// `project-plan` takes each value as a test command and keeps it as typed.
fn assert_values_accepted(values: &[&str]) {
    for value in values {
        let home = TempHome::new();
        let project = TempDir::new().unwrap();
        let input = with_decisions(vec![decision("test_command", value)]);
        let out = project_plan(&home, &project, &input);
        assert!(
            out.status.success(),
            "{value}: {}",
            String::from_utf8_lossy(&out.stdout)
        );
        assert_eq!(
            ok_result(&out)["profile"]["decisions"][0]["value"],
            json!(value)
        );
    }
}

// SPEC-20261126-profile-command-values AC-2
#[test]
fn refused_secret_shapes_behind_a_path_or_a_flag_are_named_and_write_nothing() {
    // Built at run time, so no secret shape is written in this file.
    let shapes = [
        format!("{}{}{}", "AK", "IA", "A1".repeat(8)),
        format!("{}{}{}", "AS", "IA", "A1".repeat(8)),
        format!("{}{}", "xo", "xb-0123456789abcdef"),
        format!("{}{}{}", "AI", "za", "a1".repeat(18)),
        format!("{}{}.{}.{}", "ey", "J0ZXN0", "a1".repeat(8), "b2".repeat(8)),
    ];
    for shape in &shapes {
        // The outgoing screen knows each one when it stands alone.
        assert_eq!(moai_cli::redact::redact_outgoing(shape).1, 1);
        for prefix in ["x/", "./", "a/b/", "x//", "./run x/", "./run ./x --key="] {
            assert_value_refused(&format!("{prefix}{shape}"), &[shape]);
        }
    }
}

// SPEC-20261126-profile-command-values AC-2
#[test]
fn refused_short_secret_shape_that_only_the_profile_screen_knows() {
    // Too short for the outgoing screen, whole or in pieces.
    let short = format!("{}{}", "s", "k-short");
    for value in [format!("x/{short}"), format!("./run x/{short}")] {
        for text in [value.as_str(), short.as_str()] {
            assert_eq!(moai_cli::redact::redact_outgoing(text).1, 0);
        }
        assert_value_refused(&value, &[&short]);
    }
}

// SPEC-20261126-profile-command-values AC-2
#[test]
fn refused_paths_built_on_the_home_variable_are_named_and_write_nothing() {
    for value in [
        "$HOME$X/x",
        "$HOME.d/x",
        "${HOME}$X/x",
        "${HOME}.d/x",
        "${HOME:-x}.d/y",
        "cat $hOmE-old/x",
        "x${Home}y/z",
        "go test $HOME.d/x ./...",
        // A later use counts after one that builds no path or names another variable.
        "echo $HOME $HOME.d/x",
        "$HOMEBREW_PREFIX:$HOME.d/x",
        "PATH=$HOME:./bin go test",
        // A word ends at ASCII white space only, where a shell splits it.
        "$HOME\u{a0}x/y",
    ] {
        assert_value_refused(value, &[]);
    }
}

// SPEC-20261126-profile-command-values AC-2
#[test]
fn refused_paths_built_on_a_windows_home_variable_are_named_and_write_nothing() {
    for value in [
        "$USERPROFILE/x",
        "${USERPROFILE}.d/x",
        "$HOMEDRIVE.d/x",
        "${HomeDrive}x/y",
        "$homepath/x",
        "${HOMEPATH:-x}.d/y",
        "$HOMEDRIVE$HOMEPATH/x",
    ] {
        assert_value_refused(value, &[]);
    }
}

// SPEC-20261126-profile-command-values AC-1
#[test]
fn accept_variable_uses_that_build_no_home_path() {
    assert_values_accepted(&[
        "PATH=$HOME:$PATH go test ./...",
        "echo $HOME && go test ./...",
        "echo ${HOME} ./x",
        "echo $USERPROFILE $HOMEDRIVE$HOMEPATH && ./run x/y",
        // Other variables, whose names only start like a home variable's.
        "$HOMEBREW_PREFIX/bin/x",
        "${HOMEBREW_PREFIX}x/y",
        "$HOME_DIR/x",
        "$HOME2/x",
        "$USERPROFILES/x",
        "$HOMEDRIVE_X/x",
        "${HOMEPATH2}x/y",
    ]);
}

// SPEC-20261126-profile-command-values AC-2
#[test]
fn refused_passwords_before_an_at_sign_are_named_and_write_nothing() {
    // Built at run time, so no address with a password is written in this file.
    let pair = format!("{}:{}", "user", "fixture-pw");
    for value in [
        format!("psql {pair}@host.invalid/db"),
        format!("x/{pair}@host.invalid/db"),
        format!("run --url={pair}@host.invalid/db -v"),
        // Every "@" of a word is read, the first and a later one.
        format!("psql {pair}@mid@host.invalid/db"),
        format!("psql a@{pair}@host.invalid/db"),
        // One character on each side of the ":" is enough, of any alphabet.
        format!("psql {}:{}@h/x", "a", "b"),
        format!("psql {}:{}@h/x", "가", "나"),
        // A cost the SPEC states (R-1): a package specifier of the same form.
        "npm:cowsay@1.5.0/cli".to_string(),
    ] {
        assert_value_refused(&value, &[&pair]);
    }
}

// SPEC-20261126-profile-command-values AC-1
#[test]
fn accept_at_signs_without_both_a_password_and_a_path() {
    assert_values_accepted(&[
        "docker pull ghcr.io/org/img:1.2@sha256:abc",
        "com.example:lib:1.0@aar",
        "go test github.com/org/repo/...",
        "docker run -u 1000:1000 -v ./x img",
        // No character on one side of the ":".
        "psql :b@h/x",
        "psql a:@h/x",
        "npm i alias@npm:@scope/pkg",
        // A "/" between the ":" and the "@", and a "/" in a later word only.
        "run a:b/c@d/e",
        "run a:b@c d/e",
    ]);
}

// SPEC-20261126-profile-command-values AC-1
#[test]
fn accept_the_known_limits_of_the_spec() {
    // Known limits of the SPEC (R-2): no rule of the screen reads these, so they are taken.
    // The last four are built at run time, so no credential shape is written in this file: a
    // webhook path without a scheme, a password given to a flag, a secret key with two "/" in the
    // shape of the AWS documentation's example, and a key shape behind a ":".
    let key = format!("{}{}", "AK", "IA") + &"A1".repeat(8);
    let built = [
        format!("hooks.example.invalid/services/T0/B0/{}", "x".repeat(24)),
        format!("curl -u {}:{} host.invalid/path", "user", "fixture-pw"),
        format!("{}/{}/{}", "wJalrXUtnFEMI", "K7MDENG", "bPxRfiCY".repeat(2)),
        format!("go test x:{key} ./..."),
    ];
    let mut values = vec![
        "../../Users/name",
        "x/Users/name",
        "./../../../Users/name",
        "host.invalid/path",
    ];
    values.extend(built.iter().map(String::as_str));
    assert_values_accepted(&values);
}

#[test]
fn other_fields_still_refuse_a_slash() {
    let slashed = "go test ./fixture-slash";
    for (key, value, field) in [
        ("project_id", json!(slashed), "project_id"),
        ("purpose", json!(slashed), "purpose"),
        ("tools", json!(["Rust", slashed]), "tools"),
        (
            "development_methods",
            json!([slashed]),
            "development_methods",
        ),
        ("hosts", json!(["claude", slashed]), "hosts"),
        ("package_pin", json!(slashed), "package_pin"),
        ("interview_revision", json!(slashed), "interview_revision"),
        (
            "versioning",
            json!({"scheme":"custom","format":slashed}),
            "versioning.format",
        ),
    ] {
        assert_refused(&with_key(key, value), &naming(field), &[slashed]);
    }
    for (part, field) in [
        ("id", "decisions[#0].id"),
        ("topic", "decisions[test_command].topic"),
        ("source", "decisions[test_command].source"),
        ("scope", "decisions[test_command].scope"),
    ] {
        let mut record = decision("test_command", "cargo test");
        record[part] = json!(slashed);
        assert_refused(&with_decisions(vec![record]), &naming(field), &[slashed]);
    }
}

#[test]
fn field_names_a_plain_key() {
    let many_tools: Vec<String> = (0..33).map(|i| format!("tool{i}")).collect();
    let many_records: Vec<Value> = (0..33)
        .map(|i| decision(&format!("d{i}"), "cargo test"))
        .collect();
    for (key, value) in [
        ("schema_version", json!(1)),
        ("schema_version", json!(3)),
        ("purpose", json!(" ")),
        ("tools", json!([])),
        ("tools", json!(many_tools)),
        ("development_methods", json!([])),
        ("development_methods", json!(many_tools)),
        ("hosts", json!([])),
        ("hosts", json!(["claude", "codex", "omp", "claude"])),
        ("interview_revision", json!("r".repeat(129))),
        ("decisions", json!(many_records)),
        ("context_max_bytes", json!(0)),
        ("archive_grace_days", json!(0)),
        ("stale_draft_days", json!(0)),
        ("lease_sweep_hours", json!(0)),
        ("shared_files", json!(["../fixture-up"])),
    ] {
        assert_refused(&with_key(key, value), &naming(key), &["fixture-up"]);
    }
}

#[test]
fn field_names_the_first_failing_key_in_table_order() {
    let mut input = with_key("purpose", json!("a/b"));
    input["tools"] = json!(["a/b"]);
    input["decisions"] = json!([decision("test_command", "/x")]);
    assert_refused(&input, &naming("purpose"), &[]);
    // Within a record the parts are checked in the record's order.
    let mut record = decision("test_command", "/x");
    record["scope"] = json!("a/b");
    assert_refused(
        &with_decisions(vec![record]),
        &naming("decisions[test_command].value"),
        &[],
    );
}

#[test]
fn field_names_a_decision_by_its_id() {
    let longest = "i".repeat(80);
    for id in ["build_command", "build-list", longest.as_str()] {
        assert_refused(
            &with_decisions(vec![
                decision("test_command", "cargo test"),
                decision(id, "/Users/example/x"),
            ]),
            &naming(&format!("decisions[{id}].value")),
            &["/Users/example/x"],
        );
    }
}

#[test]
fn field_names_a_decision_by_position_when_its_id_cannot_be_shown() {
    let too_long = "i".repeat(81);
    for id in ["fixture id", "fixture/id", too_long.as_str()] {
        assert_refused(
            &with_decisions(vec![
                decision("test_command", "cargo test"),
                decision(id, "cargo test"),
            ]),
            &naming("decisions[#1].id"),
            &[id],
        );
    }
    // Ids that are portable but match a secret screen: the profile screen's and the outgoing one.
    for id in [
        format!("{}{}", "gh", "p_fixtureonly"),
        format!("{}{}{}", "AK", "IA", "A1".repeat(8)),
    ] {
        assert_refused(
            &with_decisions(vec![decision(&id, "/x")]),
            &naming("decisions[#0].value"),
            &[&id],
        );
    }
}

#[test]
fn field_ids_that_read_like_other_messages_stay_invalid_input() {
    for id in [
        "temporary",
        "exited",
        "unavailable",
        "unreadable",
        "disappeared",
    ] {
        assert_refused(
            &with_decisions(vec![decision(id, "/x")]),
            &naming(&format!("decisions[{id}].value")),
            &[],
        );
    }
}

#[test]
fn shape_names_a_missing_or_wrong_typed_key() {
    for key in [
        "schema_version",
        "purpose",
        "tools",
        "development_methods",
        "hosts",
    ] {
        let mut input = profile(None);
        input.as_object_mut().unwrap().remove(key);
        assert_refused(&input, &naming(key), &[]);
    }
    let wrong = "fixture-wrong-type";
    for (key, value) in [
        ("schema_version", json!(wrong)),
        ("project_id", json!([wrong])),
        ("purpose", json!([wrong])),
        ("tools", json!(wrong)),
        ("decisions", json!(wrong)),
        (
            "decisions",
            json!([{"id":"d1","topic":"tools","value":"x","source":"user","scope":"project","extra":wrong}]),
        ),
        ("github_issues", json!(wrong)),
        ("versioning", json!(wrong)),
        ("shared_files", json!(wrong)),
    ] {
        assert_refused(&with_key(key, value), &naming(key), &[wrong]);
    }
    // The first key of the profile's table is named, whatever the order of the input.
    let mut input = with_key("tools", json!(wrong));
    input["purpose"] = json!(7);
    assert_refused(&input, &naming("purpose"), &[wrong]);
}

#[test]
fn shape_keeps_the_earlier_sentence_without_a_known_cause() {
    let unknown = "fixture_unknown_key";
    let input = with_key(unknown, json!(1));
    assert_refused(&input, PLAIN, &[unknown]);
    // An unknown key wins over a known key of the wrong type.
    let mut both = input;
    both["purpose"] = json!(7);
    assert_refused(&both, PLAIN, &[unknown]);
    for input in [
        json!([]),
        json!("fixture-not-a-profile"),
        json!(7),
        Value::Null,
    ] {
        assert_refused(&input, PLAIN, &["fixture-not-a-profile"]);
    }
}
