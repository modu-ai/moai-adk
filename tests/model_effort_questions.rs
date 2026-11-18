//! SPEC-20261118-model-effort-questions (card C-47, slice 2a): the tool pick, the model and effort
//! selections, the plan of the `launcher.models` key and its digest-bound apply to the launcher
//! settings file given by path. Every file lives in a temporary folder.
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

fn pick(tool: &str, role: &str, model: &str, effort: &str) -> Pick {
    Pick {
        tool: tool.into(),
        role: role.into(),
        model: model.into(),
        effort: effort.into(),
    }
}

fn values(q: &model_effort_questions::Question) -> Vec<&str> {
    q.options.iter().map(|o| o.value.as_str()).collect()
}

#[test]
fn the_tool_question_lists_three_tools_and_defaults_to_the_installed_ones() {
    let q = tool_question(&["codex", "claude"]);
    assert_eq!(q.kind, Kind::Many);
    assert_eq!(values(&q), ["claude", "codex", "omp"]);
    assert_eq!(q.defaults, [0, 1], "installed tools, in option order");
    assert!(tool_question(&[]).defaults.is_empty());
    // A name that is not a pickable tool is not a default.
    assert_eq!(
        tool_question(&["glm", "nope"]).defaults,
        Vec::<usize>::new()
    );
}

#[test]
fn model_and_effort_questions_offer_only_table_values_with_the_default_first() {
    for (tool, role, default_model, default_effort) in [
        ("claude", "leader", "claude-opus-5-5", "medium"),
        ("codex", "agent", "gpt-6.1-terra", "xhigh"),
        ("omp", "agent", "zai/glm-5.3-flash", "max"),
    ] {
        let q = model_question(tool, role).expect("known tool and role");
        assert_eq!(q.kind, Kind::One);
        assert_eq!(
            q.options[0].value, default_model,
            "{tool} {role}: default first"
        );
        assert_eq!(q.defaults, [0]);
        let table = model_effort_settings::supported(tool).unwrap();
        assert_eq!(q.options.len(), table.len());
        for option in &q.options {
            assert!(table.iter().any(|row| row.id == option.value));
        }
        let e = effort_question(tool, role, default_model).unwrap();
        assert_eq!(e.options[e.defaults[0]].value, default_effort);
        let row = table.iter().find(|r| r.id == default_model).unwrap();
        assert_eq!(values(&e), row.efforts);
    }
    // A model that is not the default still gets an effort default the model offers.
    let e = effort_question("codex", "agent", "gpt-6-luna").unwrap();
    assert_eq!(e.options[e.defaults[0]].value, "xhigh");
    let e = effort_question("claude", "leader", "claude-haiku-5-5").unwrap();
    assert_eq!(e.options[e.defaults[0]].value, "medium");
    // The GLM default effort for the leader is high, which flash also offers.
    let e = effort_question("glm", "leader", "zai/glm-5.3-flash").unwrap();
    assert_eq!(e.options[e.defaults[0]].value, "high");
    // No question for a name the table does not know, or a model the tool does not list.
    assert!(model_question("nope", "leader").is_none());
    assert!(model_question("codex", "nope").is_none());
    assert!(effort_question("codex", "leader", "claude-opus-5-5").is_none());
    assert!(effort_question("nope", "leader", "gpt-6.1-sol").is_none());
}

#[test]
fn an_answer_line_is_read_and_a_bad_one_is_never_echoed() {
    let many = tool_question(&["claude"]);
    let one = model_question("codex", "leader").unwrap();
    let v = |a: Answer| match a {
        Answer::Values(v) => v,
        other => panic!("expected values, got {other:?}"),
    };
    // Empty and blank lines are the default.
    assert_eq!(v(read_answer(&many, Some(""))), ["claude"]);
    assert_eq!(v(read_answer(&many, Some("  \n"))), ["claude"]);
    assert_eq!(v(read_answer(&one, Some("\n"))), ["gpt-6.1-sol"]);
    // Numbers: 1-based, in option order, duplicates collapsed, comma or space separated.
    assert_eq!(v(read_answer(&many, Some("3,1"))), ["claude", "omp"]);
    assert_eq!(v(read_answer(&many, Some("2 2 2"))), ["codex"]);
    assert_eq!(v(read_answer(&one, Some(" 2 "))), ["gpt-6.1-terra"]);
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
    assert!(text.contains("Claude Code") && text.contains("Codex") && text.contains("OMP"));
    assert!(
        !text.contains("model"),
        "no model question after an empty pick: {text}"
    );

    // The tool question comes first, then the first picked tool, leader, model.
    let mut out = Vec::new();
    let outcome = run(&mut Cursor::new("2\n"), &mut out, &[]).unwrap();
    assert_eq!(
        outcome,
        Outcome::Cancelled,
        "end of input at the first model question"
    );
    let text = String::from_utf8(out).unwrap();
    let tools_at = text.find("OMP").unwrap();
    let model_at = text.find("gpt-6.1-sol").unwrap();
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
    let bad = format!("{SECRET}\n{SECRET}\n{SECRET}\n1\n{}", "\n".repeat(4));
    let outcome = run(&mut Cursor::new(bad), &mut out, &["claude"]).unwrap();
    assert_eq!(outcome, Outcome::Cancelled);
    assert!(!String::from_utf8(out).unwrap().contains(SECRET));

    // Two bad answers and then a good one still work.
    let mut out = Vec::new();
    let input = "9\nx\n\n".to_string() + &"\n".repeat(4);
    let outcome = run(&mut Cursor::new(input), &mut out, &["claude"]).unwrap();
    assert!(
        matches!(outcome, Outcome::Picks(ref p) if p.len() == 2),
        "{outcome:?}"
    );
}

#[test]
fn enter_everywhere_picks_the_defaults_of_the_installed_tools() {
    // One line for the tool pick, then model and effort for leader and agent of each tool.
    let input = "\n".repeat(1 + 4 + 4);
    let outcome = run(
        &mut Cursor::new(input),
        &mut Vec::new(),
        &["claude", "codex"],
    )
    .unwrap();
    assert_eq!(
        outcome,
        Outcome::Picks(vec![
            pick("claude", "leader", "claude-opus-5-5", "medium"),
            pick("claude", "agent", "claude-sonnet-5-5", "high"),
            pick("codex", "leader", "gpt-6.1-sol", "high"),
            pick("codex", "agent", "gpt-6.1-terra", "xhigh"),
        ])
    );
    // A chosen model changes the effort question: codex leader picks the third model, then Enter.
    let input = "2\n3\n\n".to_string() + &"\n".repeat(2);
    let outcome = run(&mut Cursor::new(input), &mut Vec::new(), &[]).unwrap();
    match outcome {
        Outcome::Picks(p) => {
            assert_eq!(p[0], pick("codex", "leader", "gpt-6.2-sol", "high"));
        }
        other => panic!("{other:?}"),
    }
}

#[test]
fn the_plan_changes_only_the_picked_entries_and_lists_unsupported_ones() {
    let launcher = object(json!({
        "role_access": "prompt",
        "models": {
            "codex": {"agent": {"model": "gpt-6.2-sol", "effort": "max"}},
            "claude": {"leader": {"model": SECRET, "effort": "high"}},
            "glm": {"agent": {"model": "zai/glm-5.3-flash"}}
        }
    }));
    let before = launcher.clone();
    let picks = [
        pick("codex", "leader", "gpt-6.1-sol", "high"),
        pick("claude", "leader", "claude-opus-5-5", "low"),
        pick("codex", "agent", "gpt-6.1-terra", "xhigh"),
    ];
    let p = plan(&launcher, &picks).unwrap();
    assert_eq!(launcher, before, "the input is not changed");
    assert_eq!(p.changes.len(), 3);
    let change = |tool: &str, role: &str| {
        p.changes
            .iter()
            .find(|c| c.tool == tool && c.role == role)
            .unwrap_or_else(|| panic!("{tool} {role}"))
    };
    assert_eq!(change("codex", "leader").from, None);
    assert_eq!(
        change("codex", "agent").from,
        Some(Choice {
            model: "gpt-6.2-sol".into(),
            effort: "max".into()
        })
    );
    assert_eq!(
        change("claude", "leader").from,
        None,
        "an unsupported value is not a from"
    );
    assert_eq!(change("claude", "leader").to.model, "claude-opus-5-5");
    // The unsupported glm.agent stays and is listed; the replaced claude.leader is not listed.
    assert_eq!(p.kept_unsupported, ["launcher.models.glm.agent"]);
    assert_eq!(p.after["role_access"], json!("prompt"));
    assert_eq!(p.after["models"]["glm"], before["models"]["glm"]);
    assert_eq!(
        p.after["models"]["claude"]["leader"],
        json!({"model": "claude-opus-5-5", "effort": "low"})
    );
    // Nothing in the plan quotes the saved text.
    assert!(!format!("{:?}", p.kept_unsupported).contains(SECRET));

    // The digest follows the content, not the key order of the input.
    let reordered = object(json!({
        "models": {
            "glm": {"agent": {"model": "zai/glm-5.3-flash"}},
            "claude": {"leader": {"effort": "high", "model": SECRET}},
            "codex": {"agent": {"effort": "max", "model": "gpt-6.2-sol"}}
        },
        "role_access": "prompt"
    }));
    assert_eq!(plan(&reordered, &picks).unwrap().digest, p.digest);
    let other = [pick("codex", "leader", "gpt-6.1-sol", "low")];
    assert_ne!(plan(&launcher, &other).unwrap().digest, p.digest);
    assert_eq!(p.digest.len(), 64);

    // An unsupported pick is refused.
    assert!(plan(
        &launcher,
        &[pick("codex", "leader", "gpt-6.1-sol", "ultra")]
    )
    .is_err());
    assert!(plan(&launcher, &[pick("nope", "leader", "gpt-6.1-sol", "high")]).is_err());
}

#[test]
fn an_omp_pick_sets_omp_and_glm() {
    let p = plan(&Map::new(), &[pick("omp", "leader", "zai/glm-5.3", "low")]).unwrap();
    let mut tools: Vec<&str> = p.changes.iter().map(|c| c.tool.as_str()).collect();
    tools.sort_unstable();
    assert_eq!(tools, ["glm", "omp"]);
    for tool in ["omp", "glm"] {
        assert_eq!(
            resolve(tool, "leader", Some(&p.after)).unwrap(),
            (
                Choice {
                    model: "zai/glm-5.3".into(),
                    effort: "low".into()
                },
                Source::Saved
            )
        );
    }
    // The agent role was not picked and keeps its default.
    assert_eq!(
        resolve("glm", "agent", Some(&p.after)).unwrap().1,
        Source::Default
    );
}

#[test]
fn a_structurally_broken_models_key_is_refused_without_quoting_it() {
    let pick = [pick("codex", "leader", "gpt-6.1-sol", "high")];
    for launcher in [
        json!({"models": SECRET}),
        json!({"models": {SECRET: {"leader": {}}}}),
        json!({"models": {"codex": SECRET}}),
        json!({"models": {"codex": {SECRET: {}}}}),
        json!({"models": {"codex": {"agent": SECRET}}}),
    ] {
        let message = plan(&object(launcher), &pick).unwrap_err();
        assert!(message.contains("launcher.models"), "{message}");
        assert!(!message.contains(SECRET), "{message}");
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
        pick("codex", "leader", "gpt-6.1-sol", "high"),
        pick("omp", "agent", "zai/glm-5.3-flash", "max"),
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
        after["launcher"]["models"]["glm"]["agent"],
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
        created["launcher"]["models"]["codex"]["leader"]["model"],
        json!("gpt-6.1-sol")
    );
    assert_eq!(
        fs::metadata(&new_path).unwrap().permissions().mode() & 0o777,
        0o600
    );
}

#[test]
fn apply_refuses_a_stale_digest_and_an_unsafe_target_and_writes_nothing() {
    let picks = [pick("claude", "agent", "claude-sonnet-5-5", "max")];

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
    let picks = [pick("claude", "agent", "claude-sonnet-5-5", "max")];
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
    let other = [pick("codex", "leader", "gpt-6.1-sol", "low")];
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
    let odd = r#"{ "z" : 1 ,"launcher":{"models":{"claude":{"agent":{"model":"claude-sonnet-5-5","effort":"max"}}}} }"#;
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
    let input = format!("{long}\n\n{}", "\n".repeat(4));
    let mut out = Vec::new();
    let outcome = run(&mut Cursor::new(input), &mut out, &["claude"]).unwrap();
    assert!(
        matches!(outcome, Outcome::Picks(ref p) if p.len() == 2),
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
    let four_questions = "\n".repeat(4);
    let two_picks = |outcome: &Outcome| matches!(outcome, Outcome::Picks(p) if p.len() == 2);

    // 4096 content bytes are an ordinary line: the answer `1` is read.
    let line = format!("1{}\n", " ".repeat(4095));
    let mut out = Vec::new();
    let outcome = run(
        &mut Cursor::new(format!("{line}{four_questions}")),
        &mut out,
        &[],
    )
    .unwrap();
    assert!(two_picks(&outcome), "{outcome:?}");
    assert_eq!(bad(&out), 0);

    // 4097 content bytes are one invalid answer; the next line answers.
    let line = format!("1{}\n", " ".repeat(4096));
    let mut out = Vec::new();
    let input = format!("{line}1\n{four_questions}");
    let outcome = run(&mut Cursor::new(input), &mut out, &[]).unwrap();
    assert!(two_picks(&outcome), "{outcome:?}");
    assert_eq!(bad(&out), 1);

    // A long line whose tail looks like a valid answer does not leak into the next question.
    let line = format!("{}1\n", "a".repeat(8192));
    let mut out = Vec::new();
    let input = format!("{line}1\n{four_questions}");
    let outcome = run(&mut Cursor::new(input), &mut out, &[]).unwrap();
    assert!(two_picks(&outcome), "{outcome:?}");
    assert_eq!(bad(&out), 1, "the dropped tail is not a second answer");
}

#[test]
fn a_failed_rename_removes_its_temporary_file_and_a_growing_file_is_refused() {
    let picks = [pick("claude", "agent", "claude-sonnet-5-5", "max")];

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
