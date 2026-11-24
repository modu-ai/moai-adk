//! SPEC-20261118-model-effort-questions (card C-47, slice 2a): the tool pick, the model and effort
//! selections, the plan of the `launcher.models` key and its digest-bound apply to the launcher
//! settings file given by path. Every file lives in a temporary folder.
//!
//! SPEC-20261124-launcher-retire-role-flags (REQ-6, AC-9): one model and one effort question per
//! host, saved as the flat `launcher.models.<tool>` entry; a saved `leader` or `agent` entry is
//! listed as replaced by that write.
//!
//! SPEC-20261121-codex-launcher-and-models-removed: the tool question offers Claude and GLM only;
//! a `codex` value saved by an earlier version is ignored by the plan and left out of the file
//! the next save writes (leader decision D2).
#![cfg(unix)]
use serde_json::{json, Map, Value};
use std::fs;
use std::io::Cursor;
use std::os::unix::fs::{symlink, PermissionsExt};
use std::path::Path;

use moai_cli::model_effort_questions::{
    self, apply, effort_question, model_question, plan, plan_file, read_answer, run, tool_question,
    Answer, Kind, Outcome, Pick,
};
use moai_cli::model_effort_settings::{self, resolve, Choice, Source};

const SECRET: &str = "SECRET-answer-or-saved-text";

fn object(value: Value) -> Map<String, Value> {
    match value {
        Value::Object(map) => map,
        other => panic!("fixture must be an object, got {other}"),
    }
}

fn pick(tool: &str, model: &str, effort: &str) -> Pick {
    Pick {
        tool: tool.into(),
        model: model.into(),
        effort: effort.into(),
    }
}

fn values(q: &model_effort_questions::Question) -> Vec<&str> {
    q.options.iter().map(|o| o.value.as_str()).collect()
}

#[test]
fn the_tool_question_lists_two_tools_and_defaults_to_the_installed_ones() {
    let q = tool_question(&["glm", "claude"]);
    assert_eq!(q.kind, Kind::Many);
    assert_eq!(values(&q), ["claude", "glm"]);
    assert_eq!(q.defaults, [0, 1], "installed tools, in option order");
    assert!(tool_question(&[]).defaults.is_empty());
    // glm is pickable but never auto-defaulted (no folder probe finds it): it stays a
    // default only when the caller names it, here at its option index.
    assert_eq!(tool_question(&["glm", "nope"]).defaults, [1]);
    assert_eq!(tool_question(&["nope"]).defaults, Vec::<usize>::new());
    // Codex is not offered, so naming it defaults nothing.
    assert_eq!(tool_question(&["codex"]).defaults, Vec::<usize>::new());
    assert_eq!(tool_question(&["codex", "claude"]).defaults, [0]);
}

#[test]
fn model_and_effort_questions_offer_only_table_values_with_the_default_first() {
    for (tool, default_model, default_effort) in [
        ("claude", "claude-opus-5-5", "high"),
        // SPEC-20261123-glm-54-followups D-3: the glm default is 5.4.
        ("glm", "zai/glm-5.4[1m]", "high"),
    ] {
        let q = model_question(tool).expect("known tool");
        assert_eq!(q.kind, Kind::One);
        assert_eq!(
            q.id,
            format!("{tool}.model"),
            "one question per host, no role"
        );
        assert_eq!(q.options[0].value, default_model, "{tool}: default first");
        assert_eq!(q.defaults, [0]);
        let table = model_effort_settings::supported(tool).unwrap();
        assert_eq!(q.options.len(), table.len());
        for option in &q.options {
            assert!(table.iter().any(|row| row.id == option.value));
        }
        let e = effort_question(tool, default_model).unwrap();
        assert_eq!(e.id, format!("{tool}.effort"));
        assert_eq!(e.options[e.defaults[0]].value, default_effort);
        let row = table.iter().find(|r| r.id == default_model).unwrap();
        assert_eq!(values(&e), row.efforts);
    }
    // A model that is not the default still gets an effort default the model offers.
    let e = effort_question("glm", "zai/glm-5.3").unwrap();
    assert_eq!(e.options[e.defaults[0]].value, "high");
    let e = effort_question("claude", "claude-haiku-5-5").unwrap();
    assert_eq!(e.options[e.defaults[0]].value, "high");
    // The GLM default effort is high, which flash also offers.
    let e = effort_question("glm", "zai/glm-5.3-flash").unwrap();
    assert_eq!(e.options[e.defaults[0]].value, "high");
    // No question for a name the table does not know, or a model the tool does not list.
    assert!(model_question("nope").is_none());
    assert!(effort_question("glm", "claude-opus-5-5").is_none());
    assert!(effort_question("nope", "gpt-6.1-sol").is_none());
    // Codex has no question any more.
    assert!(model_question("codex").is_none());
    assert!(effort_question("codex", "gpt-6.2-sol").is_none());
}

#[test]
fn an_answer_line_is_read_and_a_bad_one_is_never_echoed() {
    let many = tool_question(&["claude"]);
    let one = model_question("claude").unwrap();
    let v = |a: Answer| match a {
        Answer::Values(v) => v,
        other => panic!("expected values, got {other:?}"),
    };
    // Empty and blank lines are the default.
    assert_eq!(v(read_answer(&many, Some(""))), ["claude"]);
    assert_eq!(v(read_answer(&many, Some("  \n"))), ["claude"]);
    assert_eq!(v(read_answer(&one, Some("\n"))), ["claude-opus-5-5"]);
    // Numbers: 1-based, in option order, duplicates collapsed, comma or space separated.
    assert_eq!(v(read_answer(&many, Some("2,1"))), ["claude", "glm"]);
    assert_eq!(v(read_answer(&many, Some("2 2 2"))), ["glm"]);
    assert_eq!(v(read_answer(&one, Some(" 2 "))), ["claude-sonnet-5-5"]);
    // There is no third tool to pick any more.
    assert_eq!(read_answer(&many, Some("3")), Answer::Invalid);
    // A single choice takes exactly one number.
    assert_eq!(read_answer(&one, Some("1,2")), Answer::Invalid);
    // Anything else is invalid, and end of input is cancelled.
    for bad in ["0", "9", "-1", "x", SECRET, "1,x", "1.5", "１"] {
        assert_eq!(read_answer(&many, Some(bad)), Answer::Invalid, "{bad:?}");
        assert_eq!(read_answer(&one, Some(bad)), Answer::Invalid, "{bad:?}");
    }
    assert_eq!(read_answer(&many, None), Answer::Cancelled);
    assert_eq!(read_answer(&one, None), Answer::Cancelled);
    // No result prints the line back.
    assert!(!format!("{:?}", read_answer(&many, Some(SECRET))).contains(SECRET));
}

#[test]
fn the_loop_asks_the_tool_pick_first_and_stops_on_an_empty_or_cancelled_pick() {
    // An empty pick (nothing installed, Enter) asks no model question.
    let mut out = Vec::new();
    let outcome = run(&mut Cursor::new("\n"), &mut out, &[]).unwrap();
    assert_eq!(outcome, Outcome::Picks(vec![]));
    let text = String::from_utf8(out).unwrap();
    assert!(text.contains("Claude") && text.contains("GLM"), "{text}");
    assert!(!text.contains("OMP"), "no OMP offer: {text}");
    assert!(
        !text.contains("Codex") && !text.contains("GPT"),
        "no Codex offer: {text}"
    );
    assert!(
        !text.contains("model"),
        "no model question after an empty pick: {text}"
    );

    // The tool question comes first, then the model of the first picked tool.
    let mut out = Vec::new();
    let outcome = run(&mut Cursor::new("1\n"), &mut out, &[]).unwrap();
    assert_eq!(
        outcome,
        Outcome::Cancelled,
        "end of input at the first model question"
    );
    let text = String::from_utf8(out).unwrap();
    let tools_at = text.find("GLM").unwrap();
    let model_at = text.find("claude-opus-5-5").unwrap();
    assert!(tools_at < model_at, "tools first: {text}");
    assert!(
        !text.contains("zai/glm"),
        "only the picked tool is asked: {text}"
    );

    // End of input at the tool question is cancelled.
    let outcome = run(&mut Cursor::new(""), &mut Vec::new(), &["claude"]).unwrap();
    assert_eq!(outcome, Outcome::Cancelled);

    // Three bad answers to one question end the loop; the line is never echoed.
    let mut out = Vec::new();
    // A fourth try would read the `1` and then complete with the empty lines after it.
    let bad = format!("{SECRET}\n{SECRET}\n{SECRET}\n1\n{}", "\n".repeat(2));
    let outcome = run(&mut Cursor::new(bad), &mut out, &["claude"]).unwrap();
    assert_eq!(outcome, Outcome::Cancelled);
    assert!(!String::from_utf8(out).unwrap().contains(SECRET));

    // Two bad answers and then a good one still work.
    let mut out = Vec::new();
    let input = "9\nx\n\n".to_string() + &"\n".repeat(2);
    let outcome = run(&mut Cursor::new(input), &mut out, &["claude"]).unwrap();
    assert!(
        matches!(outcome, Outcome::Picks(ref p) if p.len() == 1),
        "{outcome:?}"
    );
}

#[test]
fn enter_everywhere_picks_the_defaults_of_the_installed_tools() {
    // One line for the tool pick, then one model and one effort per tool.
    let input = "\n".repeat(1 + 2 + 2);
    let outcome = run(&mut Cursor::new(input), &mut Vec::new(), &["claude", "glm"]).unwrap();
    assert_eq!(
        outcome,
        Outcome::Picks(vec![
            pick("claude", "claude-opus-5-5", "high"),
            pick("glm", "zai/glm-5.4[1m]", "high"),
        ])
    );
    // A chosen model changes the effort question: the glm pick takes its default model
    // explicitly, then Enter takes the default effort.
    let input = "2\n1\n\n".to_string() + &"\n".repeat(2);
    let outcome = run(&mut Cursor::new(input), &mut Vec::new(), &[]).unwrap();
    match outcome {
        Outcome::Picks(p) => {
            assert_eq!(p[0], pick("glm", "zai/glm-5.4[1m]", "high"));
        }
        other => panic!("{other:?}"),
    }
    // AC-9: the screen asks exactly one model question and one effort question per host, and
    // none of them names a role.
    let mut out = Vec::new();
    run(
        &mut Cursor::new("1,2\n".to_string() + &"\n".repeat(4)),
        &mut out,
        &[],
    )
    .unwrap();
    let text = String::from_utf8(out).unwrap();
    assert_eq!(text.matches("which model?").count(), 2, "{text}");
    assert_eq!(text.matches("which effort").count(), 2, "{text}");
    assert!(
        !text.contains("leader") && !text.contains("agent"),
        "{text}"
    );
}

/// SPEC-20261123-glm-54-followups REQ-1..REQ-3 / AC-1: the glm model question offers the
/// 5.4 id first as the default and keeps both 5.3 rows after it; a 5.3 choice saved earlier is a
/// supported value, so Enter keeps it as the default and its effort question still answers.
#[test]
fn the_glm_model_question_offers_54_first_and_keeps_a_saved_53() {
    let q = model_question("glm").unwrap();
    assert_eq!(
        values(&q),
        ["zai/glm-5.4[1m]", "zai/glm-5.3", "zai/glm-5.3-flash"]
    );
    assert_eq!(q.defaults, [0]);
    // The 5.4 row offers the efforts the table lists, with high as the default.
    let e = effort_question("glm", "zai/glm-5.4[1m]").unwrap();
    assert_eq!(values(&e), ["low", "high", "max"]);
    assert_eq!(e.options[e.defaults[0]].value, "high");
    // A saved 5.3 choice: it is the default option, and the other two follow.
    let saved = Choice {
        model: "zai/glm-5.3".into(),
        effort: "low".into(),
    };
    let q = model_effort_questions::model_question_with("glm", Some(&saved)).unwrap();
    assert_eq!(
        values(&q),
        ["zai/glm-5.3", "zai/glm-5.4[1m]", "zai/glm-5.3-flash"]
    );
    let e =
        model_effort_questions::effort_question_with("glm", "zai/glm-5.3", Some(&saved)).unwrap();
    assert_eq!(e.options[e.defaults[0]].value, "low");
}

#[test]
fn the_plan_changes_only_the_picked_tools_and_lists_unsupported_ones() {
    let launcher = object(json!({
        "role_access": "prompt",
        "models": {
            "codex": {"agent": {"model": "gpt-6.2-sol", "effort": "max"}},
            "claude": {
                "leader": {"model": SECRET, "effort": "high"},
                "agent": {"model": "claude-sonnet-5-5", "effort": "high"}
            },
            "glm": {"model": "zai/glm-5.3-flash"},
            "omp": {"agent": {"model": "zai/glm-5.3-flash", "effort": "high"}}
        }
    }));
    let before = launcher.clone();
    let picks = [pick("claude", "claude-opus-5-5", "low")];
    let p = plan(&launcher, &picks).unwrap();
    assert_eq!(launcher, before, "the input is not changed");
    assert_eq!(p.changes.len(), 1);
    assert_eq!(p.changes[0].tool, "claude");
    assert_eq!(
        p.changes[0].from, None,
        "a saved role entry is not a flat choice, so there is no from"
    );
    assert_eq!(p.changes[0].to.model, "claude-opus-5-5");
    // AC-9: the write replaces the whole claude object, so both saved role entries are listed as
    // replaced (their values are never quoted), and the claude entry is the flat pair alone.
    assert_eq!(
        p.replaced,
        [
            "launcher.models.claude.leader",
            "launcher.models.claude.agent"
        ]
    );
    assert_eq!(
        p.after["models"]["claude"],
        json!({"model": "claude-opus-5-5", "effort": "low"})
    );
    assert!(!format!("{:?}", p.replaced).contains(SECRET));
    // The unsupported flat glm entry is not picked: it stays and is listed; the replaced claude
    // role entries are not listed as unsupported.
    assert_eq!(p.kept_unsupported, ["launcher.models.glm"]);
    assert_eq!(p.after["role_access"], json!("prompt"));
    assert_eq!(p.after["models"]["glm"], before["models"]["glm"]);
    // The saved Codex and OMP values are not a choice of any tool: they are not listed, and the
    // written file leaves them out (leader decision D2).
    assert!(p.after["models"].get("codex").is_none(), "{:?}", p.after);
    assert!(p.after["models"].get("omp").is_none(), "{:?}", p.after);
    // Nothing in the plan quotes the saved text.
    assert!(!format!("{:?}", p.kept_unsupported).contains(SECRET));

    // A glm pick replaces the unsupported glm entry: it is no longer listed. The unsupported
    // claude role entry is not picked, so it stays listed.
    let glm = plan(&launcher, &[pick("glm", "zai/glm-5.3", "high")]).unwrap();
    assert_eq!(glm.kept_unsupported, ["launcher.models.claude.leader"]);
    assert!(glm.replaced.is_empty(), "glm held no role entry");
    assert_eq!(
        glm.changes[0].from, None,
        "an unsupported value is not a from"
    );

    // The digest follows the content, not the key order of the input.
    let reordered = object(json!({
        "models": {
            "glm": {"model": "zai/glm-5.3-flash"},
            "claude": {
                "agent": {"effort": "high", "model": "claude-sonnet-5-5"},
                "leader": {"effort": "high", "model": SECRET}
            },
            "codex": {"agent": {"effort": "max", "model": "gpt-6.2-sol"}},
            "omp": {"agent": {"effort": "high", "model": "zai/glm-5.3-flash"}}
        },
        "role_access": "prompt"
    }));
    assert_eq!(plan(&reordered, &picks).unwrap().digest, p.digest);
    let other = [pick("claude", "claude-opus-5-5", "high")];
    assert_ne!(plan(&launcher, &other).unwrap().digest, p.digest);
    assert_eq!(p.digest.len(), 64);

    // An unsupported pick is refused.
    assert!(plan(&launcher, &[pick("glm", "zai/glm-5.3", "ultra")]).is_err());
    // A pick for OMP is refused, whatever it holds.
    assert!(plan(&launcher, &[pick("omp", "zai/glm-5.3", "high")]).is_err());
    assert!(plan(&launcher, &[pick("nope", "gpt-6.1-sol", "high")]).is_err());
    // A pick for Codex is refused, whatever it holds.
    assert!(plan(&launcher, &[pick("codex", "gpt-6.2-sol", "high")]).is_err());
}

#[test]
fn a_pick_writes_only_its_own_tool() {
    // A glm pick is its own pick now (the earlier OMP pick also set glm; OMP is not a tool any
    // more, so this plan API refuses an omp pick).
    let p = plan(&Map::new(), &[pick("glm", "zai/glm-5.3", "low")]).unwrap();
    let tools: Vec<&str> = p.changes.iter().map(|c| c.tool.as_str()).collect();
    assert_eq!(tools, ["glm"]);
    assert_eq!(
        model_effort_settings::flat_choice(&p.after, "glm"),
        Some(Choice {
            model: "zai/glm-5.3".into(),
            effort: "low".into()
        })
    );
    // Claude was not picked and keeps nothing saved, and omp stays untouched.
    assert_eq!(model_effort_settings::flat_choice(&p.after, "claude"), None);
    assert_eq!(p.after.get("models").and_then(|m| m.get("omp")), None);
    assert!(plan(&Map::new(), &[pick("omp", "zai/glm-5.3", "low")]).is_err());
}

#[test]
fn a_structurally_broken_models_key_is_refused_without_quoting_it() {
    let pick = [pick("claude", "claude-opus-5-5", "high")];
    for launcher in [
        json!({"models": SECRET}),
        json!({"models": {SECRET: {"leader": {}}}}),
        json!({"models": {"claude": SECRET}}),
        json!({"models": {"claude": {SECRET: {}}}}),
        json!({"models": {"claude": {"agent": SECRET}}}),
    ] {
        let message = plan(&object(launcher), &pick).unwrap_err();
        assert!(message.contains("launcher.models"), "{message}");
        assert!(!message.contains(SECRET), "{message}");
    }
}

#[test]
fn a_saved_codex_value_of_any_shape_does_not_block_a_plan_and_is_dropped() {
    let pick = [pick("claude", "claude-opus-5-5", "high")];
    for codex in [
        json!({"agent": {"model": "gpt-6.2-sol", "effort": "max"}}),
        json!(SECRET),
        json!({SECRET: {}}),
        json!({"agent": SECRET}),
        json!({"leader": {"model": SECRET, "effort": SECRET}}),
    ] {
        let launcher = object(json!({"models": {"codex": codex.clone()}}));
        let p = plan(&launcher, &pick).unwrap_or_else(|e| panic!("{codex}: {e}"));
        assert!(
            p.after["models"].get("codex").is_none(),
            "{codex}: left out of the written file"
        );
        assert!(p.kept_unsupported.is_empty(), "{codex}: not listed");
        assert_eq!(p.changes.len(), 1);
    }
}

fn launcher_file(dir: &Path, body: &str, mode: u32) -> std::path::PathBuf {
    let path = dir.join("config").join("launcher.yaml");
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(&path, body).unwrap();
    fs::set_permissions(&path, fs::Permissions::from_mode(mode)).unwrap();
    path
}

fn read_json(path: &Path) -> Value {
    serde_json::from_slice(&fs::read(path).unwrap()).unwrap()
}

fn entries(dir: &Path) -> Vec<String> {
    let mut names: Vec<String> = fs::read_dir(dir)
        .unwrap()
        .map(|e| e.unwrap().file_name().to_string_lossy().into_owned())
        .collect();
    names.sort();
    names
}

#[test]
fn apply_writes_through_the_digest_and_reads_back() {
    let home = tempfile::tempdir().unwrap();
    let path = launcher_file(
        home.path(),
        r#"{"launcher":{"role_access":"prompt","advisor":"fable"},"other":[1,{"a":null}]}"#,
        0o644,
    );
    let picks = [
        pick("claude", "claude-opus-5-5", "high"),
        pick("glm", "zai/glm-5.3-flash", "max"),
    ];
    let previewed = plan_file(&path, &picks).unwrap();
    let before = fs::read(&path).unwrap();
    assert_eq!(
        fs::read(&path).unwrap(),
        before,
        "the preview writes nothing"
    );

    let applied = apply(&path, &picks, &previewed.digest).unwrap();
    assert_eq!(applied.digest, previewed.digest);
    let after = read_json(&path);
    assert_eq!(after["other"], json!([1, {"a": null}]));
    assert_eq!(after["launcher"]["role_access"], json!("prompt"));
    assert_eq!(after["launcher"]["advisor"], json!("fable"));
    assert_eq!(after["launcher"]["models"], json!(applied.after["models"]));
    assert_eq!(
        after["launcher"]["models"]["glm"],
        json!({"model": "zai/glm-5.3-flash", "effort": "max"})
    );
    assert_eq!(
        fs::metadata(&path).unwrap().permissions().mode() & 0o777,
        0o644,
        "mode kept"
    );
    assert_eq!(
        entries(path.parent().unwrap()),
        ["launcher.yaml"],
        "no temp file left"
    );

    // The same digest cannot be applied twice for a changed file: the plan now reads the saved
    // choice as `from`, so the digest differs.
    assert!(apply(&path, &picks, &previewed.digest).is_err());

    // An absent file (and folder) is created with only the launcher key, private.
    let fresh = tempfile::tempdir().unwrap();
    let new_path = fresh.path().join("config").join("launcher.yaml");
    let previewed = plan_file(&new_path, &picks).unwrap();
    apply(&new_path, &picks, &previewed.digest).unwrap();
    let created = read_json(&new_path);
    assert_eq!(
        created
            .as_object()
            .unwrap()
            .keys()
            .map(String::as_str)
            .collect::<Vec<_>>(),
        ["launcher"]
    );
    assert_eq!(
        created["launcher"]["models"]["claude"]["model"],
        json!("claude-opus-5-5")
    );
    assert_eq!(
        fs::metadata(&new_path).unwrap().permissions().mode() & 0o777,
        0o600
    );
}

#[test]
fn apply_refuses_a_stale_digest_and_an_unsafe_target_and_writes_nothing() {
    let picks = [pick("claude", "claude-sonnet-5-5", "max")];

    // A wrong digest and a file that changed after the preview.
    let home = tempfile::tempdir().unwrap();
    let body = r#"{"launcher":{"role_access":"prompt"}}"#;
    let path = launcher_file(home.path(), body, 0o600);
    assert!(apply(&path, &picks, &"0".repeat(64)).is_err());
    assert_eq!(fs::read_to_string(&path).unwrap(), body);
    let previewed = plan_file(&path, &picks).unwrap();
    let changed = r#"{"launcher":{"role_access":"full"}}"#;
    fs::write(&path, changed).unwrap();
    assert!(apply(&path, &picks, &previewed.digest).is_err());
    assert_eq!(fs::read_to_string(&path).unwrap(), changed);

    // A link, a directory, a special file and an oversize file; and a broken file whose text must
    // not reach the message.
    let home = tempfile::tempdir().unwrap();
    let target = home.path().join("target.txt");
    fs::write(&target, body).unwrap();
    let config = home.path().join("config");
    fs::create_dir_all(&config).unwrap();
    let link = config.join("launcher.yaml");
    symlink(&target, &link).unwrap();
    let digest = plan(&Map::new(), &picks).unwrap().digest;
    assert!(plan_file(&link, &picks).is_err());
    assert!(apply(&link, &picks, &digest).is_err());
    assert_eq!(fs::read_to_string(&target).unwrap(), body);

    let dir_home = tempfile::tempdir().unwrap();
    let dir_path = dir_home.path().join("config").join("launcher.yaml");
    fs::create_dir_all(&dir_path).unwrap();
    assert!(plan_file(&dir_path, &picks).is_err());
    assert!(apply(&dir_path, &picks, &digest).is_err());
    assert!(dir_path.is_dir());

    let fifo_home = tempfile::tempdir().unwrap();
    let fifo = fifo_home.path().join("launcher.yaml");
    let made = std::process::Command::new("mkfifo")
        .arg(&fifo)
        .status()
        .unwrap();
    assert!(made.success());
    assert!(plan_file(&fifo, &picks).is_err(), "a FIFO is never opened");
    assert!(apply(&fifo, &picks, &digest).is_err());

    let big_home = tempfile::tempdir().unwrap();
    let big = launcher_file(big_home.path(), &" ".repeat(1024 * 1024 + 1), 0o600);
    assert!(plan_file(&big, &picks).is_err());
    assert!(apply(&big, &picks, &digest).is_err());
    assert_eq!(fs::metadata(&big).unwrap().len(), 1024 * 1024 + 1);

    let broken_home = tempfile::tempdir().unwrap();
    let broken_body = format!("{{ {SECRET}");
    let broken = launcher_file(broken_home.path(), &broken_body, 0o600);
    let message = plan_file(&broken, &picks).unwrap_err();
    assert!(!message.contains(SECRET), "{message}");
    let message = apply(&broken, &picks, &digest).unwrap_err();
    assert!(!message.contains(SECRET), "{message}");
    assert_eq!(fs::read_to_string(&broken).unwrap(), broken_body);
}

#[test]
fn the_apply_guards_hold_at_their_edges() {
    let picks = [pick("claude", "claude-sonnet-5-5", "max")];
    let digest = plan(&Map::new(), &picks).unwrap().digest;

    // The folder of the file is a link: refused, and nothing appears behind the link.
    let home = tempfile::tempdir().unwrap();
    let real = home.path().join("real");
    fs::create_dir_all(&real).unwrap();
    let linked = home.path().join("config");
    symlink(&real, &linked).unwrap();
    assert!(apply(&linked.join("launcher.yaml"), &picks, &digest).is_err());
    assert!(entries(&real).is_empty());

    // A mode the umask would change is kept (0664 becomes 0644 without the explicit permission call).
    let home = tempfile::tempdir().unwrap();
    let path = launcher_file(home.path(), r#"{"launcher":{}}"#, 0o664);
    let preview = plan_file(&path, &picks).unwrap();
    apply(&path, &picks, &preview.digest).unwrap();
    assert_eq!(
        fs::metadata(&path).unwrap().permissions().mode() & 0o777,
        0o664
    );

    // A temporary-looking file of someone else is never touched.
    let dir = path.parent().unwrap();
    let foreign = dir.join(format!(".launcher.tmp-{}", std::process::id()));
    fs::write(&foreign, "foreign").unwrap();
    let other = [pick("claude", "claude-haiku-5-5", "low")];
    let preview = plan_file(&path, &other).unwrap();
    apply(&path, &other, &preview.digest).unwrap();
    assert_eq!(fs::read_to_string(&foreign).unwrap(), "foreign");
    assert_eq!(
        entries(dir).len(),
        2,
        "only the foreign file and the settings file"
    );

    // A folder that cannot be written fails clean: the file is as it was and nothing is left.
    let home = tempfile::tempdir().unwrap();
    let body = r#"{"launcher":{"role_access":"prompt"}}"#;
    let path = launcher_file(home.path(), body, 0o600);
    let dir = path.parent().unwrap();
    let preview = plan_file(&path, &picks).unwrap();
    fs::set_permissions(dir, fs::Permissions::from_mode(0o555)).unwrap();
    let probe = dir.join(".probe");
    let can_write_anyway = fs::write(&probe, "").is_ok(); // running as the superuser
    let _ = fs::remove_file(&probe);
    let result = apply(&path, &picks, &preview.digest);
    fs::set_permissions(dir, fs::Permissions::from_mode(0o755)).unwrap();
    if !can_write_anyway {
        assert!(result.is_err());
        assert_eq!(fs::read_to_string(&path).unwrap(), body);
        assert_eq!(entries(dir), ["launcher.yaml"]);
    }

    // The size limit: exactly one MiB is read, one byte more is refused.
    let home = tempfile::tempdir().unwrap();
    let head = r#"{"launcher":{}}"#;
    let exact = format!("{head}{}", " ".repeat(1024 * 1024 - head.len()));
    assert_eq!(exact.len(), 1024 * 1024);
    let path = launcher_file(home.path(), &exact, 0o600);
    assert!(plan_file(&path, &picks).is_ok());
    fs::write(&path, format!("{exact} ")).unwrap();
    assert!(plan_file(&path, &picks).is_err());

    // Nothing to change writes nothing: the same pick again leaves odd formatting as it is, and
    // an empty pick list leaves an absent file (and its folder) absent.
    let home = tempfile::tempdir().unwrap();
    let odd = r#"{ "z" : 1 ,"launcher":{"models":{"claude":{"model":"claude-sonnet-5-5","effort":"max"}}} }"#;
    let path = launcher_file(home.path(), odd, 0o600);
    let preview = plan_file(&path, &picks).unwrap();
    apply(&path, &picks, &preview.digest).unwrap();
    assert_eq!(fs::read_to_string(&path).unwrap(), odd);
    let fresh = tempfile::tempdir().unwrap();
    let absent = fresh.path().join("config").join("launcher.yaml");
    let empty = plan_file(&absent, &[]).unwrap();
    apply(&absent, &[], &empty.digest).unwrap();
    assert!(!fresh.path().join("config").exists());
}

#[test]
fn an_overlong_answer_line_is_one_invalid_answer() {
    // The long line counts as one bad answer, its rest is dropped, and the next line is read.
    let long = "a".repeat(10_000);
    let input = format!("{long}\n\n{}", "\n".repeat(2));
    let mut out = Vec::new();
    let outcome = run(&mut Cursor::new(input), &mut out, &["claude"]).unwrap();
    assert!(
        matches!(outcome, Outcome::Picks(ref p) if p.len() == 1),
        "{outcome:?}"
    );
    assert!(!String::from_utf8(out).unwrap().contains("aaaa"));
    // A long line at the end of input, with no newline, ends as cancelled.
    let outcome = run(
        &mut Cursor::new("b".repeat(5000)),
        &mut Vec::new(),
        &["claude"],
    )
    .unwrap();
    assert_eq!(outcome, Outcome::Cancelled);
}

#[test]
fn the_answer_line_cap_is_exact_and_a_long_tail_never_leaks() {
    let bad = |out: &[u8]| {
        String::from_utf8_lossy(out)
            .matches("not one of the numbers")
            .count()
    };
    let two_questions = "\n".repeat(2);
    let one_pick = |outcome: &Outcome| matches!(outcome, Outcome::Picks(p) if p.len() == 1);

    // 4096 content bytes are an ordinary line: the answer `1` is read.
    let line = format!("1{}\n", " ".repeat(4095));
    let mut out = Vec::new();
    let outcome = run(
        &mut Cursor::new(format!("{line}{two_questions}")),
        &mut out,
        &[],
    )
    .unwrap();
    assert!(one_pick(&outcome), "{outcome:?}");
    assert_eq!(bad(&out), 0);

    // 4097 content bytes are one invalid answer; the next line answers.
    let line = format!("1{}\n", " ".repeat(4096));
    let mut out = Vec::new();
    let input = format!("{line}1\n{two_questions}");
    let outcome = run(&mut Cursor::new(input), &mut out, &[]).unwrap();
    assert!(one_pick(&outcome), "{outcome:?}");
    assert_eq!(bad(&out), 1);

    // A long line whose tail looks like a valid answer does not leak into the next question.
    let line = format!("{}1\n", "a".repeat(8192));
    let mut out = Vec::new();
    let input = format!("{line}1\n{two_questions}");
    let outcome = run(&mut Cursor::new(input), &mut out, &[]).unwrap();
    assert!(one_pick(&outcome), "{outcome:?}");
    assert_eq!(bad(&out), 1, "the dropped tail is not a second answer");
}

#[test]
fn a_failed_rename_removes_its_temporary_file_and_a_growing_file_is_refused() {
    let picks = [pick("claude", "claude-sonnet-5-5", "max")];

    // Renaming onto `..` fails after the temporary file was created: it must be removed.
    let home = tempfile::tempdir().unwrap();
    let path = home.path().join("newfolder").join("..");
    let preview = plan_file(&path, &picks).unwrap();
    assert!(apply(&path, &picks, &preview.digest).is_err());
    assert!(
        entries(&home.path().join("newfolder")).is_empty(),
        "the temporary file is removed after a failed rename"
    );

    // A file that the rewrite would grow past one MiB is refused before anything is written.
    let home = tempfile::tempdir().unwrap();
    let body = format!(
        r#"{{"launcher":{{}},"pad":[{}]}}"#,
        vec!["1"; 480_000].join(",")
    );
    assert!(
        (body.len() as u64) < 1024 * 1024,
        "the file itself is under the limit"
    );
    let path = launcher_file(home.path(), &body, 0o600);
    let preview = plan_file(&path, &picks).unwrap();
    let message = apply(&path, &picks, &preview.digest).unwrap_err();
    assert!(message.contains("larger than one MiB"), "{message}");
    assert_eq!(fs::read_to_string(&path).unwrap(), body);
    assert_eq!(entries(path.parent().unwrap()), ["launcher.yaml"]);
}

#[test]
fn a_saved_value_is_the_default_of_its_questions_and_enter_keeps_it() {
    use moai_cli::model_effort_questions::{effort_question_with, model_question_with, run_with};
    let saved = Choice {
        model: "claude-sonnet-5-5".into(),
        effort: "max".into(),
    };
    // The saved model comes first and is the default; the saved effort is the default effort of
    // that model, and another model keeps the table default.
    let models = model_question_with("claude", Some(&saved)).unwrap();
    assert_eq!(models.options[0].value, "claude-sonnet-5-5");
    assert_eq!(models.defaults, vec![0]);
    assert_eq!(models.options.len(), 4, "every model once: {models:?}");
    let efforts = effort_question_with("claude", "claude-sonnet-5-5", Some(&saved)).unwrap();
    assert_eq!(efforts.options[efforts.defaults[0]].value, "max");
    let other = effort_question_with("claude", "claude-opus-5-5", Some(&saved)).unwrap();
    assert_eq!(other.options[other.defaults[0]].value, "high");

    // Enter at every question keeps the saved flat entry of claude and gives the table default to
    // glm, which nobody saved; a saved role entry beside it is not a default of any question.
    let launcher = object(json!({"models": {
        "claude": {"model": "claude-sonnet-5-5", "effort": "max"},
        "glm": {"leader": {"model": "zai/glm-5.3", "effort": "low"}}
    }}));
    let outcome = run_with(
        &mut Cursor::new("1,2\n\n\n\n\n"),
        &mut Vec::new(),
        &[],
        &launcher,
    )
    .unwrap();
    let Outcome::Picks(picks) = outcome else {
        panic!("picks expected");
    };
    let claude = picks.iter().find(|p| p.tool == "claude").unwrap();
    assert_eq!(
        (claude.model.as_str(), claude.effort.as_str()),
        ("claude-sonnet-5-5", "max")
    );
    let glm = picks.iter().find(|p| p.tool == "glm").unwrap();
    assert_eq!(
        (glm.model.as_str(), glm.effort.as_str()),
        ("zai/glm-5.4[1m]", "high")
    );
}

/// A temporary folder for the key-vault scenes: unique per test and removed when the test
/// ends, so no folder with a fake key stays in the system temporary folder. It derefs to the
/// folder's path.
struct KeyHome(tempfile::TempDir);

impl std::ops::Deref for KeyHome {
    type Target = Path;

    fn deref(&self) -> &Path {
        self.0.path()
    }
}

fn key_temp_home(tag: &str) -> KeyHome {
    KeyHome(
        tempfile::Builder::new()
            .prefix(&format!("glm-key-{tag}-"))
            .tempdir()
            .unwrap(),
    )
}

/// The project scope of the test's own folder, resolved with the temporary `home` (the plain
/// `scope_here` would read the process HOME).
fn glm_scope(home: &Path) -> moai_cli::key_vault::Scope {
    let cwd = std::env::current_dir().unwrap();
    model_effort_questions::glm_key_scope(&cwd, home).expect("the tests run inside a git work tree")
}

#[test]
fn a_glm_pick_asks_for_the_key_once_and_stores_it_in_the_vault() {
    let home = key_temp_home("store");
    let scope = glm_scope(&home);
    let input = format!("{SECRET}\n");
    let mut out = Vec::new();
    model_effort_questions::glm_key_step(
        &mut Cursor::new(input),
        &mut out,
        &home,
        &[pick("glm", "glm-4", "high")],
        false,
    )
    .unwrap();
    let text = String::from_utf8(out).unwrap();
    assert!(
        text.contains("hidden"),
        "the question says the input is hidden: {text}"
    );
    assert!(
        text.contains("skip"),
        "the question says it is skippable: {text}"
    );
    assert!(
        !text.contains(SECRET),
        "the value appears in no output: {text}"
    );
    let stored = moai_cli::key_vault::read_entry(&home, &scope, "glm").unwrap();
    assert_eq!(
        stored.as_deref(),
        Some(SECRET),
        "the vault holds the typed value"
    );
    // The value is encrypted at rest: no file under the temporary home holds it as text, so
    // it is in no log, no receipt and no settings file either.
    assert_eq!(
        files_holding(&home, SECRET),
        Vec::<String>::new(),
        "the plain value is in no file under the home"
    );
}

/// The files under `dir` (relative names) whose bytes contain `needle`.
fn files_holding(dir: &Path, needle: &str) -> Vec<String> {
    let mut found = Vec::new();
    let mut folders = vec![dir.to_path_buf()];
    while let Some(folder) = folders.pop() {
        for entry in fs::read_dir(&folder).unwrap().flatten() {
            let path = entry.path();
            let kind = entry.file_type().unwrap();
            if kind.is_dir() {
                folders.push(path);
            } else if kind.is_file() {
                let bytes = fs::read(&path).unwrap();
                if bytes
                    .windows(needle.len())
                    .any(|window| window == needle.as_bytes())
                {
                    found.push(path.strip_prefix(dir).unwrap().display().to_string());
                }
            }
        }
    }
    found
}

#[test]
fn an_empty_answer_skips_the_key_and_keeps_the_pick_flow() {
    let home = key_temp_home("skip");
    let scope = glm_scope(&home);
    let mut out = Vec::new();
    model_effort_questions::glm_key_step(
        &mut Cursor::new("\n"),
        &mut out,
        &home,
        &[pick("glm", "glm-4", "high")],
        false,
    )
    .unwrap();
    let text = String::from_utf8(out).unwrap();
    assert!(text.contains("hidden"), "the question still asks: {text}");
    assert!(
        moai_cli::key_vault::read_entry(&home, &scope, "glm")
            .unwrap()
            .is_none(),
        "nothing was stored"
    );
}

#[test]
fn an_existing_key_is_kept_by_default_without_retyping() {
    let home = key_temp_home("keep");
    let scope = glm_scope(&home);
    moai_cli::key_vault::set(&home, &scope, "glm", "SECRET-existing").unwrap();
    let mut out = Vec::new();
    model_effort_questions::glm_key_step(
        &mut Cursor::new("\n"),
        &mut out,
        &home,
        &[pick("glm", "glm-4", "high")],
        false,
    )
    .unwrap();
    let text = String::from_utf8(out).unwrap();
    assert!(
        text.contains("already in the key vault"),
        "the existing key is named: {text}"
    );
    assert!(
        !text.contains("SECRET-existing"),
        "the stored value is not shown: {text}"
    );
    let stored = moai_cli::key_vault::read_entry(&home, &scope, "glm").unwrap();
    assert_eq!(
        stored.as_deref(),
        Some("SECRET-existing"),
        "the stored key is kept"
    );
}

#[test]
fn a_new_answer_replaces_an_existing_key() {
    let home = key_temp_home("new");
    let scope = glm_scope(&home);
    moai_cli::key_vault::set(&home, &scope, "glm", "SECRET-existing").unwrap();
    let mut out = Vec::new();
    model_effort_questions::glm_key_step(
        &mut Cursor::new("n\nSECRET-replacement\n"),
        &mut out,
        &home,
        &[pick("glm", "glm-4", "high")],
        false,
    )
    .unwrap();
    assert!(
        !String::from_utf8(out)
            .unwrap()
            .contains("SECRET-replacement"),
        "no echo of the new value"
    );
    let stored = moai_cli::key_vault::read_entry(&home, &scope, "glm").unwrap();
    assert_eq!(
        stored.as_deref(),
        Some("SECRET-replacement"),
        "the new value is stored"
    );
}

#[test]
fn the_word_new_also_asks_for_a_new_key() {
    let home = key_temp_home("new-word");
    let scope = glm_scope(&home);
    moai_cli::key_vault::set(&home, &scope, "glm", "SECRET-existing").unwrap();
    model_effort_questions::glm_key_step(
        &mut Cursor::new("new\nSECRET-replacement\n"),
        &mut Vec::new(),
        &home,
        &[pick("glm", "glm-4", "high")],
        false,
    )
    .unwrap();
    let stored = moai_cli::key_vault::read_entry(&home, &scope, "glm").unwrap();
    assert_eq!(stored.as_deref(), Some("SECRET-replacement"));
}

#[cfg(unix)]
#[test]
fn a_vault_that_cannot_be_written_is_told_on_the_screen_and_the_step_goes_on() {
    use std::os::unix::fs::PermissionsExt;
    let home = key_temp_home("read-only");
    // The scope is resolved before the folder is closed: the identity reads the home state.
    let scope = glm_scope(&home);
    fs::set_permissions(&*home, fs::Permissions::from_mode(0o500)).unwrap();
    let mut out = Vec::new();
    let result = model_effort_questions::glm_key_step(
        &mut Cursor::new(format!("{SECRET}\n")),
        &mut out,
        &home,
        &[pick("glm", "glm-4", "high")],
        false,
    );
    fs::set_permissions(&*home, fs::Permissions::from_mode(0o700)).unwrap();
    result.expect("a refused write does not end the step");
    let text = String::from_utf8(out).unwrap();
    assert!(
        text.contains("The key could not be stored:"),
        "the screen says the write failed: {text}"
    );
    assert!(!text.contains(SECRET), "the value is not echoed: {text}");
    assert!(
        moai_cli::key_vault::read_entry(&home, &scope, "glm")
            .unwrap()
            .is_none(),
        "nothing was stored"
    );
}

#[test]
fn a_broken_project_entry_is_not_hidden_by_a_shared_key() {
    // The launcher stops at a project entry that exists but cannot be read; it does not go on to
    // the shared one. The question therefore does not claim a key exists.
    let home = key_temp_home("broken");
    let scope = glm_scope(&home);
    moai_cli::key_vault::set(
        &home,
        &moai_cli::key_vault::Scope::Shared,
        "glm",
        "SECRET-shared",
    )
    .unwrap();
    let entry = moai_cli::key_vault::entry_path(&home, &scope, "glm");
    fs::create_dir_all(entry.parent().unwrap()).unwrap();
    fs::write(&entry, b"not a sealed key").unwrap();
    let mut out = Vec::new();
    model_effort_questions::glm_key_step(
        &mut Cursor::new("\n"),
        &mut out,
        &home,
        &[pick("glm", "glm-4", "high")],
        false,
    )
    .unwrap();
    let text = String::from_utf8(out).unwrap();
    assert!(
        !text.contains("already in the key vault"),
        "a broken project entry is no existing key: {text}"
    );
    assert!(text.contains("hidden"), "the key is asked for: {text}");
}

#[test]
fn a_claude_pick_asks_nothing_about_a_key() {
    let home = key_temp_home("claude");
    let mut out = Vec::new();
    model_effort_questions::glm_key_step(
        &mut Cursor::new(""),
        &mut out,
        &home,
        &[pick("claude", "claude-sonnet-5-5", "high")],
        false,
    )
    .unwrap();
    assert!(
        out.is_empty(),
        "no key question for another provider: {}",
        String::from_utf8_lossy(&out)
    );
}

#[test]
fn a_shared_key_counts_as_an_existing_key_and_is_kept_without_a_project_copy() {
    // The launcher finds the project's entry, then the shared one; a shared key is a key.
    let home = key_temp_home("shared");
    let scope = glm_scope(&home);
    moai_cli::key_vault::set(
        &home,
        &moai_cli::key_vault::Scope::Shared,
        "glm",
        "SECRET-shared",
    )
    .unwrap();
    let mut out = Vec::new();
    model_effort_questions::glm_key_step(
        &mut Cursor::new("\n"),
        &mut out,
        &home,
        &[pick("glm", "glm-4", "high")],
        false,
    )
    .unwrap();
    let text = String::from_utf8(out).unwrap();
    assert!(text.contains("already in the key vault"), "{text}");
    assert!(!text.contains("SECRET-shared"), "{text}");
    assert!(
        moai_cli::key_vault::read_entry(&home, &scope, "glm")
            .unwrap()
            .is_none(),
        "keeping stores no project copy"
    );
}

#[test]
fn a_piped_answer_that_is_a_new_line_after_the_key_is_not_read_as_the_key() {
    // One answer line is the key; the line after it stays for the next question (the save
    // question of the stage), so the key step reads exactly one line.
    let home = key_temp_home("one-line");
    let scope = glm_scope(&home);
    let mut input = Cursor::new(format!("{SECRET}\ny\n"));
    let mut out = Vec::new();
    model_effort_questions::glm_key_step(
        &mut input,
        &mut out,
        &home,
        &[pick("glm", "glm-4", "high")],
        false,
    )
    .unwrap();
    let mut rest = String::new();
    std::io::BufRead::read_line(&mut input, &mut rest).unwrap();
    assert_eq!(rest, "y\n", "the next line is left for the next question");
    let stored = moai_cli::key_vault::read_entry(&home, &scope, "glm").unwrap();
    assert_eq!(stored.as_deref(), Some(SECRET));
}

// SPEC-20261122-model-fast-pair REQ-4: the API refuses an Opus 4.8 request with fast mode from
// 2026-11-30, so no model question offers an id of that family.
#[test]
fn fast_pair_no_model_question_offers_an_opus_48_id() {
    for tool in model_effort_settings::TOOLS {
        let q = model_question(tool).expect("known tool");
        assert!(!q.options.is_empty(), "{tool}");
        for value in values(&q) {
            // The match ends there or goes on with `-`, `[` or `@`; a digit after it is
            // another model, the same shape as `is_opus_48` in `claude_settings_merge`.
            let lower = value.to_ascii_lowercase();
            let names_family = ["opus-4-8", "opus-4.8"].iter().any(|family| {
                lower.match_indices(family).any(|(at, _)| {
                    matches!(
                        lower[at + family.len()..].chars().next(),
                        None | Some('-' | '[' | '@')
                    )
                })
            });
            assert!(!names_family, "{tool}: {value}");
        }
    }
}
