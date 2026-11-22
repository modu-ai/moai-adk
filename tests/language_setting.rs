//! Per-project writing language by item (Issue #2293), end to end through the binary in
//! temporary projects only: profile validation (AC-1), the setup input step feeding
//! `project-plan`/`project-apply` (AC-2) and the CLAUDE.md managed block rendered from the
//! setting through `agents-*` and `generation-*` (AC-3).

mod common;
use serde_json::{json, Value};
use std::{
    fs,
    io::Write,
    path::Path,
    process::{Command, Output, Stdio},
};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;
const PROFILE: &str = ".moai-cli/project.yaml";

fn project(language: Option<Value>) -> TempDir {
    let root = TempDir::new().unwrap();
    fs::create_dir_all(root.path().join("skills/moai/workflows")).unwrap();
    fs::create_dir_all(root.path().join(".moai-cli")).unwrap();
    let mut profile = json!({"schema_version":2,"project_id":"demo","purpose":"Build a small editor","tools":["Rust"],"development_methods":["TDD"],"hosts":["claude","codex"]});
    if let Some(l) = language {
        profile["language"] = l;
    }
    let mut bytes = serde_json::to_vec_pretty(&profile).unwrap();
    bytes.push(b'\n');
    fs::write(root.path().join(PROFILE), &bytes).unwrap();
    root
}

fn run(args: &[&str], stdin: &[u8]) -> Output {
    // #2935 L1: `language-setup` reads a conversation default from HOME; an empty temporary HOME
    // keeps these cases off the developer's real one.
    let home = TempDir::new().unwrap();
    let mut command = Command::new(EXE);
    if args.first() == Some(&"language-setup") {
        command.env("HOME", home.path());
    }
    let mut child = command
        .args(args)
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

fn ok(out: &Output) -> Value {
    assert!(
        out.status.success(),
        "{}{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn failed(out: &Output) -> String {
    assert!(!out.status.success());
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    v["error"]["message"].as_str().unwrap().to_owned()
}

fn path(root: &TempDir) -> &str {
    root.path().to_str().unwrap()
}

fn plan_then_apply(root: &TempDir, op: &str, request: &Value) -> Value {
    let plan = ok(&run(
        &[&format!("{op}-plan"), path(root)],
        request.to_string().as_bytes(),
    ));
    ok(&run(
        &[
            &format!("{op}-apply"),
            path(root),
            "--plan-digest",
            plan["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "fixture-approved",
        ],
        plan.to_string().as_bytes(),
    ));
    plan
}

fn profile_value(root: &Path) -> Value {
    serde_json::from_slice(&fs::read(root.join(PROFILE)).unwrap()).unwrap()
}

fn agents_request(language: &str) -> Value {
    json!({"protocol_version":1,"language":language,"workflow_reference":"skills/moai/workflows/","local_file":"none"})
}

fn claude_md(root: &TempDir) -> String {
    fs::read_to_string(root.path().join("CLAUDE.md")).unwrap()
}

/// AC-1: `project-plan` refuses an unknown item and an unsupported code without echoing
/// the value, and a hand-edited profile with either is refused by every reader.
#[test]
fn unknown_items_and_codes_are_refused() {
    let root = project(None);
    let base = profile_value(root.path());
    for bad in [
        json!({"code_comments": "en", "slogans": "en"}),
        json!({"code_comments": "klingon"}),
        json!({}),
    ] {
        let mut profile = base.clone();
        profile["language"] = bad;
        let message = failed(&run(
            &["project-plan", path(&root)],
            profile.to_string().as_bytes(),
        ));
        assert!(
            !message.contains("klingon") && !message.contains("slogans"),
            "{message}"
        );
    }
    let edited = project(Some(json!({"code_comments": "xx"})));
    assert!(
        failed(&run(&["language-setup", path(&edited), "--questions"], b"")).contains("invalid")
    );
    assert!(!run(
        &["agents-plan", path(&edited)],
        agents_request("en").to_string().as_bytes()
    )
    .status
    .success());
}

fn setup_answers(root: &TempDir, answers: &Value, extra: &[&str]) -> Value {
    let mut args = vec!["language-setup", path(root), "--answers"];
    args.extend_from_slice(extra);
    ok(&run(&args, answers.to_string().as_bytes()))
}

/// `run` with a given HOME: the installed flow, where the HOME state folder holds the
/// conversation value `language-set --home` recorded.
fn run_with_home(args: &[&str], stdin: &[u8], home: &Path) -> Output {
    let mut command = Command::new(EXE);
    if args.first().is_some_and(|a| a.starts_with("language-")) {
        command.env("HOME", home);
    }
    let mut child = command
        .args(args)
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

fn setup_answers_in(root: &TempDir, answers: &Value, extra: &[&str], home: &Path) -> Value {
    let mut args = vec!["language-setup", path(root), "--answers"];
    args.extend_from_slice(extra);
    ok(&run_with_home(&args, answers.to_string().as_bytes(), home))
}

fn apply_profile(root: &TempDir, profile: &Value) -> Value {
    let plan = ok(&run(
        &["project-plan", path(root)],
        profile.to_string().as_bytes(),
    ));
    ok(&run(
        &[
            "project-apply",
            path(root),
            "--plan-digest",
            plan["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "fixture-approved",
        ],
        plan.to_string().as_bytes(),
    ));
    plan
}

/// AC-2: answers become the profile's `language` through project-plan/apply only; a re-run
/// with no answers keeps every value (plan `unchanged`, file bytes identical). The HOME here
/// is the installed flow's: it holds the conversation value `language-set --home` recorded.
#[test]
fn setup_answers_reach_the_profile_and_rerun_keeps_them() {
    let root = project(None);
    let home = tempfile::tempdir().unwrap();
    fs::create_dir(home.path().join(".moai-cli")).unwrap();
    let recorded = ok(&run_with_home(
        &["language-set", "--home", "ko"],
        b"",
        home.path(),
    ));
    assert_eq!(recorded["written"], json!(true), "{recorded}");
    let before = fs::read(root.path().join(PROFILE)).unwrap();
    let out = setup_answers_in(
        &root,
        &json!({"commit_messages": "ja"}),
        &["--conversation", "ko"],
        home.path(),
    );
    // The step itself writes nothing.
    assert_eq!(fs::read(root.path().join(PROFILE)).unwrap(), before);
    assert_eq!(out["profile_written"], json!(false));
    // The three asked items; `documents` was not answered, so it takes the conversation
    // language, and the saved conversation fills the profile's user_chat.
    let expected = json!({"commit_messages": "ja", "documents": "ko", "instruction_files": "en", "user_chat": "ko"});
    assert_eq!(out["language"], expected);
    let plan = apply_profile(&root, &out["profile"]);
    assert_eq!(plan["action"], json!("update"));
    assert_eq!(plan["delta"]["added"], json!(["language"]));
    assert_eq!(profile_value(root.path())["language"], expected);
    let applied = fs::read(root.path().join(PROFILE)).unwrap();

    // Re-run: no answers, no --conversation. Every value is kept: the HOME record answers.
    let again = setup_answers_in(&root, &json!({}), &[], home.path());
    assert_eq!(again["changed_items"], json!([]));
    assert_eq!(again["language"], expected);
    let replan = ok(&run(
        &["project-plan", path(&root)],
        again["profile"].to_string().as_bytes(),
    ));
    assert_eq!(replan["action"], json!("unchanged"));
    apply_profile(&root, &again["profile"]);
    assert_eq!(fs::read(root.path().join(PROFILE)).unwrap(), applied);

    // The question set lists the three items, with the stored values as current and suggested.
    let q = ok(&run_with_home(
        &["language-setup", path(&root), "--questions"],
        b"",
        home.path(),
    ));
    let items = q["questions"]["items"].as_array().unwrap();
    let ids: Vec<&str> = items.iter().map(|i| i["id"].as_str().unwrap()).collect();
    assert_eq!(ids, ["commit_messages", "instruction_files", "documents"]);
    let commits = items.iter().find(|i| i["id"] == "commit_messages").unwrap();
    assert_eq!(commits["current"], json!("ja"));
    assert_eq!(commits["suggested"], json!("ja"));

    // One changed answer changes exactly that item.
    let third = setup_answers_in(&root, &json!({"documents": "en"}), &[], home.path());
    assert_eq!(third["changed_items"], json!(["documents"]));
    let plan = apply_profile(&root, &third["profile"]);
    assert_eq!(plan["delta"]["changed"], json!(["language"]));
    assert_eq!(
        profile_value(root.path())["language"]["documents"],
        json!("en")
    );
}

/// The ask-once rule's migration: with no flag and no HOME record the conversation resolves to
/// English, so a re-run moves a stored `user_chat` ko → en once (the stored item answers stay
/// sticky), and every re-run after that keeps every value.
#[test]
fn without_a_home_record_the_rerun_moves_the_conversation_to_english() {
    let root = project(None);
    let out = setup_answers(
        &root,
        &json!({"commit_messages": "ja"}),
        &["--conversation", "ko"],
    );
    let ko = json!({"commit_messages": "ja", "documents": "ko", "instruction_files": "en", "user_chat": "ko"});
    assert_eq!(out["language"], ko);
    apply_profile(&root, &out["profile"]);

    let again = setup_answers(&root, &json!({}), &[]);
    // The stored `documents` answer is sticky (a stored value is its own suggestion); only the
    // resolved `user_chat` moves.
    let en = json!({"commit_messages": "ja", "documents": "ko", "instruction_files": "en", "user_chat": "en"});
    assert_eq!(again["changed_items"], json!(["user_chat"]));
    assert_eq!(again["language"], en);
    apply_profile(&root, &again["profile"]);
    let stable = setup_answers(&root, &json!({}), &[]);
    assert_eq!(stable["changed_items"], json!([]));
    assert_eq!(stable["language"], en);
}

/// AC-2: without a question tool the step asks on the terminal: an empty line keeps the
/// suggested value, end of input keeps the rest, a bad code is asked again.
#[test]
fn interactive_setup_reads_one_answer_per_item() {
    let root = project(Some(json!({"commit_messages": "zh"})));
    // Three prompts, in order: commit_messages, instruction_files, documents.
    let out = ok(&run(&["language-setup", path(&root)], b"ja\n\nfr\nko\n"));
    let l = &out["language"];
    assert_eq!(l["commit_messages"], json!("ja"));
    // An empty line keeps the stored value; instruction_files has none, so it is English.
    assert_eq!(l["instruction_files"], json!("en"));
    // A bad code is asked again; the next line answers the third item.
    assert_eq!(l["documents"], json!("ko"));
    // Nothing else is asked; the resolved conversation fills the profile's user_chat.
    assert_eq!(l.as_object().unwrap().len(), 4, "{l}");
    let stderr = String::from_utf8(run(&["language-setup", path(&root)], b"fr\n").stderr).unwrap();
    assert!(stderr.contains("unsupported language code"), "{stderr}");
    let refused = run(&["language-setup", path(&root)], b"fr\nde\nes\n");
    assert!(failed(&refused).contains("too many unsupported answers"));
    assert!(!String::from_utf8_lossy(&refused.stdout).contains("\"de\""));
}

/// The managed block of a profile without `language`, pinned: no writing-language line. The
/// block lives in CLAUDE.md and states facts only (SPEC-20261122-claude-md-layout REQ-3): the
/// project id and tools, the check command, and where the SPECs and cards are.
fn plain_en(root: &Path) -> String {
    format!(
        "<!-- moai:managed:begin id=moai-workflow -->
## MoAI project facts (managed block)

Project `demo` · tools: Rust.
Checks: `moai check affected`.
SPECs: `{}/specs/` · cards: `moai card list`.
<!-- moai:managed:end id=moai-workflow -->
",
        moai_cli::state_root::dir(root)
    )
}

#[test]
fn absent_language_renders_the_block_without_a_language_line() {
    let root = project(None);
    plan_then_apply(&root, "agents", &agents_request("en"));
    assert_eq!(
        claude_md(&root),
        format!("# demo\n\n{}", plain_en(root.path()))
    );
}

/// AC-3: two projects with different settings get managed text that states exactly their
/// settings, through `agents-*` and through `generation-*`.
#[test]
fn managed_section_states_each_projects_setting() {
    let korean = project(Some(json!({
        "user_chat": "ko", "reports": "ko", "issue_pr_records": "ko", "commit_messages": "ko",
        "agent_messages": "en", "instruction_files": "ko", "skill_instructions": "en",
        "code_comments": "en"
    })));
    let plan = plan_then_apply(&korean, "agents", &agents_request("ko"));
    let text = claude_md(&korean);
    // The CLAUDE.md block states the writing languages on one line (SPEC-20261122-claude-md-layout
    // REQ-3): every item the profile sets, in item order, then that the rest is English. The
    // instruction-files label is matched around its parenthesis, which names the instruction
    // file and follows the layout of the text slice.
    let line = text
        .lines()
        .find(|l| l.starts_with("작성 언어: "))
        .unwrap_or_else(|| panic!("no writing-language line:\n{text}"));
    let (head, tail) = line
        .split_once("; 지침 파일(")
        .unwrap_or_else(|| panic!("{line}"));
    assert_eq!(
        head,
        "작성 언어: 사용자 대화: 한국어 (`ko`); 사용자 보고: 한국어 (`ko`); Issue·PR 기록: 한국어 (`ko`); 커밋 메시지: 한국어 (`ko`); 에이전트 간 메시지: 영어 (`en`)"
    );
    assert!(
        tail.ends_with(
            " 규칙): 한국어 (`ko`); skill 지침: 영어 (`en`); 코드 주석: 영어 (`en`); 그 밖은 영어."
        ),
        "{line}"
    );
    // The managed section follows instruction_files, so nothing is reported.
    assert!(!plan["warnings"].to_string().contains("instruction_files"));

    // A second project, partial setting (card C-87 S1): only the items it sets get a line; an
    // unset item is never filled with a guessed default (before, `reports` read English here
    // although the conversation is Japanese).
    // Applied through project-apply so the profile has a baseline for the later change.
    let japanese = project(None);
    let mut profile = profile_value(japanese.path());
    profile["language"] = json!({"user_chat": "ja", "code_comments": "zh"});
    apply_profile(&japanese, &profile);
    let generation = json!({
        "protocol_version": 1, "language": "en",
        "workflow_reference": "skills/moai/workflows/", "local_file": "none",
        "workflow_id": "review-flow", "native_roles": [], "skill_projections": []
    });
    plan_then_apply(&japanese, "generation", &generation);
    let text = claude_md(&japanese);
    let block = "Checks: `moai check affected`.
Writing language: Conversation with the user: Japanese (`ja`); Code comments: Chinese (`zh`); everything else English.
";
    assert!(text.contains(block), "{text}");
    assert!(!claude_md(&korean).contains("Writing language"));

    // Changing the setting and re-applying rewrites only the managed lines.
    let mut profile = profile_value(japanese.path());
    profile["language"]["code_comments"] = json!("ja");
    apply_profile(&japanese, &profile);
    plan_then_apply(&japanese, "agents", &agents_request("en"));
    assert!(claude_md(&japanese)
        .contains("; Code comments: Japanese (`ja`); everything else English.\n"));
}

/// #2706: the stored `instruction_files` language decides the managed section, ahead of the
/// request language; a re-run with the stored setting is unchanged and text outside the
/// markers survives; a code without managed-section text (`ja`) falls back to the request
/// language with a warning.
#[test]
fn the_stored_instruction_language_writes_the_managed_section() {
    // Stored the way `/moai project` stores it: through project-apply (a baseline exists).
    let root = project(None);
    let mut profile = profile_value(root.path());
    profile["language"] = json!({"instruction_files": "ko"});
    apply_profile(&root, &profile);
    plan_then_apply(&root, "agents", &agents_request("en"));
    let text = claude_md(&root);
    assert!(text.contains("## MoAI 프로젝트 사실 (관리 구간)"), "{text}");
    assert!(!text.contains("## MoAI project facts (managed block)"));
    // The writing-language line speaks the section's language too (mutation G3); the label's
    // parenthesis names the instruction file and is not pinned here.
    assert!(text.contains("작성 언어: 지침 파일("), "{text}");
    assert!(
        text.contains(" 규칙): 한국어 (`ko`); 그 밖은 영어.\n"),
        "{text}"
    );
    assert!(!text.contains("Writing language"));
    // User text outside the markers survives a re-run, and the re-run plans unchanged.
    let with_note = format!("{text}\n## Our notes\n\nkeep me\n");
    std::fs::write(root.path().join("CLAUDE.md"), &with_note).unwrap();
    let again = plan_then_apply(&root, "agents", &agents_request("en"));
    assert!(
        again["files"]
            .as_array()
            .unwrap()
            .iter()
            .any(|f| f["path"] == json!("CLAUDE.md") && f["action"] == json!("unchanged")),
        "{again}"
    );
    assert_eq!(claude_md(&root), with_note);

    // Changing the stored language rewrites only the managed section.
    let mut profile = profile_value(root.path());
    profile["language"]["instruction_files"] = json!("en");
    apply_profile(&root, &profile);
    plan_then_apply(&root, "agents", &agents_request("ko"));
    let text = claude_md(&root);
    assert!(
        text.contains("## MoAI project facts (managed block)"),
        "{text}"
    );
    assert!(text.ends_with("## Our notes\n\nkeep me\n"));

    // `ja` has no managed-section text: the request language, with a warning.
    let japanese = project(Some(json!({"instruction_files": "ja"})));
    let plan = plan_then_apply(&japanese, "agents", &agents_request("ko"));
    assert!(claude_md(&japanese).contains("## MoAI 프로젝트 사실 (관리 구간)"));
    assert!(
        plan["warnings"]
            .to_string()
            .contains("no managed-section text in profile language.instruction_files `ja`"),
        "{plan}"
    );
}

/// #2696 C2: in a migrated project (the profile under `.moai/`, marker present) the managed
/// block names the folder the project's state actually lives in. The CLAUDE.md block names no
/// profile path any more (SPEC-20261122-claude-md-layout REQ-3); its SPEC place carries the
/// folder.
#[test]
fn a_migrated_project_names_its_own_state_folder() {
    let root = project(Some(json!({"user_chat": "en", "code_comments": "en"})));
    fs::create_dir(root.path().join(".moai")).unwrap();
    fs::rename(
        root.path().join(PROFILE),
        root.path().join(".moai/project.yaml"),
    )
    .unwrap();
    fs::write(
        root.path().join(".moai/moai-cli.json"),
        r#"{"schema_version":1,"generation":"moai-cli","created_by":"moai-cli test"}"#,
    )
    .unwrap();
    plan_then_apply(&root, "agents", &agents_request("en"));
    let text = claude_md(&root);
    let state = moai_cli::state_root::dir(root.path());
    assert_ne!(
        state,
        moai_cli::PROJECT_STATE_DIR,
        "the project is migrated"
    );
    assert!(text.contains(&format!("SPECs: `{state}/specs/`")), "{text}");
    assert!(
        !text.contains(&format!("{}/", moai_cli::PROJECT_STATE_DIR)),
        "{text}"
    );
    // AUDIT #2794 B2: the preview renders the same section the apply wrote — a re-plan right
    // after the apply finds CLAUDE.md `unchanged` (a plan that named another state folder
    // would render a different block and plan an update).
    let again = ok(&run(
        &["agents-plan", path(&root)],
        agents_request("en").to_string().as_bytes(),
    ));
    let agents = again["files"]
        .as_array()
        .unwrap()
        .iter()
        .find(|f| f["path"] == "CLAUDE.md")
        .unwrap_or_else(|| panic!("{again}"))
        .clone();
    assert_eq!(agents["action"], json!("unchanged"), "{again}");
}

/// The explicit `--conversation` flag is checked like every other entry into the language
/// settings: a code outside the list is refused with nothing offered or written, and a listed
/// code answers as the flag source.
#[test]
fn a_conversation_flag_outside_the_list_is_refused() {
    let root = project(None);
    let before = fs::read(root.path().join(PROFILE)).unwrap();
    let message = failed(&run(
        &[
            "language-setup",
            path(&root),
            "--questions",
            "--conversation",
            "xx",
        ],
        b"",
    ));
    assert!(
        message.contains("invalid language setting"),
        "the flag is checked: {message}"
    );
    assert!(!message.contains("xx"), "{message}");
    assert_eq!(fs::read(root.path().join(PROFILE)).unwrap(), before);
    let q = ok(&run(
        &[
            "language-setup",
            path(&root),
            "--questions",
            "--conversation",
            "ko",
        ],
        b"",
    ));
    assert_eq!(q["questions"]["conversation_language"], json!("ko"));
    assert_eq!(q["questions"]["conversation_source"], json!("flag"));
}
