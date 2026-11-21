//! SPEC-20261121-claude-agents-md-import: Claude Code reads AGENTS.md directly only when no
//! CLAUDE.md exists and never reads AGENTS.local.md; Codex has no import syntax. The opt-in pick
//! `claude-agents-import` plans one managed CLAUDE.md block holding the `@AGENTS.md` line, and a
//! project that uses a personal file gets, in the AGENTS.md block, the instruction to read it
//! plus an `@AGENTS.local.md` line. Nothing outside a managed block is touched.

mod common;

use common::cli::{create, ok_result, profile};

use serde_json::{json, Value};
use std::fs;
use std::io::Write;
use std::path::Path;
use std::process::{Command, Stdio};
use tempfile::TempDir;

const CLAUDE: &str = "CLAUDE.md";
const AGENTS: &str = "AGENTS.md";
const PICK: &str = "claude-agents-import";

/// Runs `moai ARGS` with `input` (JSON) on stdin. The child's HOME is a temporary folder, so no
/// command of this file can read or write the real one (tests/home_isolation_scan.rs).
fn cli(args: &[&str], input: &Value) -> std::process::Output {
    let home = TempDir::new().unwrap();
    let mut child = Command::new(common::MOAI_BIN)
        .args(args)
        .env("HOME", home.path())
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    match child
        .stdin
        .take()
        .unwrap()
        .write_all(input.to_string().as_bytes())
    {
        Ok(()) => {}
        Err(e) if e.kind() == std::io::ErrorKind::BrokenPipe => {}
        Err(e) => panic!("stdin write failed: {e}"),
    }
    child.wait_with_output().unwrap()
}

fn request(language: &str, local_file: &str, picks: &[&str]) -> Value {
    let mut r = json!({
        "protocol_version": 1,
        "language": language,
        "workflow_reference": "skills/moai/workflows/",
        "local_file": local_file,
    });
    if !picks.is_empty() {
        r["picks"] = json!(picks);
    }
    r
}

fn plan_ok(root: &TempDir, request: &Value) -> Value {
    let out = cli(&["agents-plan", root.path().to_str().unwrap()], request);
    assert!(
        out.status.success(),
        "{:?} {:?}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    ok_result(&out)
}

fn apply_ok(root: &TempDir, plan: &Value) -> Value {
    let digest = plan["plan_digest"].as_str().unwrap();
    let out = cli(
        &[
            "agents-apply",
            root.path().to_str().unwrap(),
            "--plan-digest",
            digest,
            "--permission-reference",
            "fixture-approved",
        ],
        plan,
    );
    assert!(
        out.status.success(),
        "{:?} {:?}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    ok_result(&out)
}

/// A project with a profile that targets `claude` and `codex`, and the workflow folder.
fn project() -> TempDir {
    project_with(&profile(None))
}

fn project_with(profile: &Value) -> TempDir {
    let root = TempDir::new().unwrap();
    fs::create_dir_all(root.path().join("skills/moai/workflows")).unwrap();
    create(&root, profile);
    root
}

fn row<'a>(plan: &'a Value, path: &str) -> Option<&'a Value> {
    plan["files"]
        .as_array()
        .unwrap()
        .iter()
        .find(|f| f["path"] == path)
}

fn read(root: &TempDir, path: &str) -> String {
    fs::read_to_string(root.path().join(path)).unwrap()
}

/// The text between the managed markers of section `id` in `text` (marker lines excluded).
fn block(text: &str, id: &str) -> Option<String> {
    let begin = format!("<!-- moai:managed:begin id={id} -->");
    let end = format!("<!-- moai:managed:end id={id} -->");
    let s = text.find(&begin)? + begin.len();
    let e = text.find(&end)?;
    Some(text[s..e].trim_matches('\n').to_string())
}

/// The lines of `text` the host would read as text, not as a code span or a fenced block.
fn prose_lines(text: &str) -> Vec<String> {
    let mut fenced = false;
    let mut out = Vec::new();
    for line in text.lines() {
        if line.trim_start().starts_with("```") {
            fenced = !fenced;
        } else if !fenced {
            out.push(line.split('`').step_by(2).collect::<Vec<_>>().join(" "));
        }
    }
    out
}

/// AC-1: with the pick and no CLAUDE.md, the plan creates one holding only the managed block with
/// the import line, shown in the preview; the plan writes nothing.
#[test]
fn the_pick_plans_a_claude_md_block_with_only_the_import_line() {
    let root = project();
    let plan = plan_ok(&root, &request("en", "none", &[PICK]));
    let r = row(&plan, CLAUDE).unwrap_or_else(|| panic!("no row for {CLAUDE}: {plan}"));
    assert_eq!(r["action"], json!("create"), "{r}");
    let detail = r["detail"].as_str().unwrap_or_default();
    assert!(detail.contains("+ @AGENTS.md"), "{detail}");
    assert!(!root.path().join(CLAUDE).exists(), "a plan writes nothing");

    apply_ok(&root, &plan);
    let written = read(&root, CLAUDE);
    assert_eq!(
        block(&written, "moai-claude-import").as_deref(),
        Some("@AGENTS.md")
    );
    assert!(
        written.trim_start().starts_with("<!-- moai:managed:begin"),
        "only the block: {written:?}"
    );
}

/// AC-2: applying appends the block after the user's text byte for byte, once, and a second plan
/// finds nothing to do.
#[test]
fn applying_appends_after_the_user_text_and_a_second_plan_is_unchanged() {
    let root = project();
    // No final newline and a CRLF line: both stay as they are.
    let user = "# My Claude notes\r\nUse tabs\r\nkeep this last line";
    fs::write(root.path().join(CLAUDE), user).unwrap();
    let plan = plan_ok(&root, &request("en", "none", &[PICK]));
    assert_eq!(
        row(&plan, CLAUDE).unwrap()["action"],
        json!("append"),
        "{plan}"
    );
    apply_ok(&root, &plan);
    let after = read(&root, CLAUDE);
    assert!(after.starts_with(user), "user bytes first: {after:?}");
    assert_eq!(after.matches("moai:managed:begin").count(), 1, "{after:?}");
    assert_eq!(
        block(&after, "moai-claude-import").as_deref(),
        Some("@AGENTS.md")
    );

    let again = plan_ok(&root, &request("en", "none", &[PICK]));
    assert_eq!(
        row(&again, CLAUDE).unwrap()["action"],
        json!("unchanged"),
        "{again}"
    );
}

/// AC-3: a line outside the block that already imports AGENTS.md wins: no row, nothing written.
/// An `@AGENTS.md` inside a code span is not an import, so it does not count.
#[test]
fn a_claude_md_that_already_imports_agents_md_gets_no_block() {
    for user in ["@AGENTS.md\n# mine\n", "See @AGENTS.md for the rules\n"] {
        let root = project();
        fs::write(root.path().join(CLAUDE), user).unwrap();
        let plan = plan_ok(&root, &request("en", "none", &[PICK]));
        assert!(row(&plan, CLAUDE).is_none(), "{user:?}: {plan}");
        assert_eq!(read(&root, CLAUDE), user);
    }
    let root = project();
    fs::write(
        root.path().join(CLAUDE),
        "Mention `@AGENTS.md` only in prose\n",
    )
    .unwrap();
    let plan = plan_ok(&root, &request("en", "none", &[PICK]));
    assert_eq!(
        row(&plan, CLAUDE).unwrap()["action"],
        json!("append"),
        "{plan}"
    );
}

/// AC-4: no pick, or no `claude` host, plans no CLAUDE.md row.
#[test]
fn without_the_pick_or_the_claude_host_there_is_no_claude_md_row() {
    let root = project();
    let plan = plan_ok(&root, &request("en", "none", &[]));
    assert!(row(&plan, CLAUDE).is_none(), "no pick: {plan}");

    let mut codex_only = profile(None);
    codex_only["hosts"] = json!(["codex"]);
    let root = project_with(&codex_only);
    let plan = plan_ok(&root, &request("en", "none", &[PICK]));
    assert!(row(&plan, CLAUDE).is_none(), "codex only: {plan}");
}

/// AC-5: a project with a personal file gets the read instruction and the import line in the
/// AGENTS.md block, in English and in Korean.
#[test]
fn agents_md_tells_every_agent_to_read_the_local_file_and_imports_it() {
    for (language, phrase) in [
        ("en", "read it at session start"),
        ("ko", "세션을 시작할 때"),
    ] {
        for local in ["ignored", "tracked"] {
            let root = project();
            let plan = plan_ok(&root, &request(language, local, &[]));
            apply_ok(&root, &plan);
            let text = block(&read(&root, AGENTS), "moai-workflow").expect("managed block");
            assert!(text.contains(phrase), "{language}/{local}: {text}");
            let imports: Vec<&str> = text
                .lines()
                .filter(|l| l.trim() == "@AGENTS.local.md")
                .collect();
            assert_eq!(imports.len(), 1, "{language}/{local}: {text}");
        }
    }
}

/// AC-6: without a personal file the block is the one before this change: the `ignored` block
/// with the two added lines removed equals the `none` block.
#[test]
fn agents_md_is_unchanged_when_no_local_file_is_chosen() {
    let none = project();
    apply_ok(&none, &plan_ok(&none, &request("en", "none", &[])));
    let none_block = block(&read(&none, AGENTS), "moai-workflow").unwrap();
    assert!(!none_block.contains("@AGENTS.local.md"), "{none_block}");
    assert!(!none_block.contains("session start"), "{none_block}");

    let ignored = project();
    apply_ok(&ignored, &plan_ok(&ignored, &request("en", "ignored", &[])));
    let ignored_block = block(&read(&ignored, AGENTS), "moai-workflow").unwrap();
    let without_added: Vec<&str> = ignored_block
        .lines()
        .filter(|l| !l.contains("session start") && l.trim() != "@AGENTS.local.md")
        .collect();
    assert_eq!(without_added.join("\n"), none_block);
}

/// AC-7: the host imports exactly the lines meant to be imported: outside code spans and fences
/// the only lines that start with `@` are the two import lines, and no code span holds an `@`.
#[test]
fn the_rendered_blocks_have_no_at_sign_inside_a_code_span() {
    let root = project();
    apply_ok(&root, &plan_ok(&root, &request("en", "ignored", &[PICK])));
    for (file, id) in [(CLAUDE, "moai-claude-import"), (AGENTS, "moai-workflow")] {
        let text = block(&read(&root, file), id).unwrap();
        for line in prose_lines(&text) {
            let at: Vec<&str> = line
                .split_whitespace()
                .filter(|t| t.starts_with('@'))
                .collect();
            for t in at {
                assert!(
                    t == "@AGENTS.md" || t == "@AGENTS.local.md",
                    "{file}: unexpected import {t}"
                );
            }
        }
        let spans: Vec<&str> = text.split('`').skip(1).step_by(2).collect();
        assert!(spans.iter().all(|s| !s.contains('@')), "{file}: {spans:?}");
    }
}

/// AC-8: the plan warns when the projected AGENTS.md passes 200 lines (the length Claude Code
/// advises): 199, 200 and 201 lines are the edges.
#[test]
fn the_plan_warns_when_agents_md_passes_two_hundred_lines() {
    let user_file = |user_lines: usize| -> TempDir {
        let root = project();
        fs::write(root.path().join(AGENTS), "line\n".repeat(user_lines)).unwrap();
        root
    };
    let warns = |user_lines: usize| -> bool {
        let root = user_file(user_lines);
        let plan = plan_ok(&root, &request("en", "none", &[]));
        plan["warnings"].to_string().contains("200 lines")
    };
    // The managed block and its separator add a fixed number of lines to the user's: apply once
    // to learn the real total, then pick the user-line counts that give 199, 200 and 201 lines.
    let root = user_file(150);
    apply_ok(&root, &plan_ok(&root, &request("en", "none", &[])));
    let total = read(&root, AGENTS).lines().count();
    let user_for = |lines: usize| 150 + lines - total;
    assert!(!warns(user_for(199)), "199 lines must not warn");
    assert!(!warns(user_for(200)), "200 lines must not warn");
    assert!(warns(user_for(201)), "201 lines must warn");
}

/// AC-9: the new pick is a known id, an unknown one is refused, and the schema lists the same
/// ids as the code.
#[test]
fn the_pick_is_a_known_id_and_an_unknown_one_is_refused() {
    let root = project();
    let out = cli(
        &["agents-plan", root.path().to_str().unwrap()],
        &request("en", "none", &["no-such-pick"]),
    );
    assert!(!out.status.success());
    let ok = plan_ok(&root, &request("en", "none", &[PICK]));
    assert!(ok["plan_digest"].is_string());

    let schema: Value = serde_json::from_str(
        &fs::read_to_string(
            Path::new(env!("CARGO_MANIFEST_DIR"))
                .join("schemas/protocol/v1/agents-request.schema.json"),
        )
        .unwrap(),
    )
    .unwrap();
    let mut in_schema: Vec<String> = schema["properties"]["picks"]["items"]["enum"]
        .as_array()
        .expect("the schema lists the picks")
        .iter()
        .map(|v| v.as_str().unwrap().to_string())
        .collect();
    let mut in_code: Vec<String> = moai_cli::agents::PICKS
        .iter()
        .map(|s| s.to_string())
        .collect();
    in_schema.sort();
    in_code.sort();
    assert_eq!(in_schema, in_code);
    assert!(in_code.contains(&PICK.to_string()));
}
