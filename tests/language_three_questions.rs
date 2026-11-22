//! SPEC-20261119-language-three-questions: the language setup asks three questions
//! (conversation, commit messages, documents) and writes everything else in English.
//!
//! End to end through the binary, in temporary projects and temporary HOME folders only; the real
//! HOME is never read or written. The installer's saved conversation language is a file in the
//! temporary HOME (`.moai-cli/config/language.yaml`).

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

/// A project whose profile has review answers (so REVIEW.md is planned too) and, when given, a
/// `language` value.
fn project(language: Option<Value>) -> TempDir {
    let root = TempDir::new().unwrap();
    fs::create_dir_all(root.path().join("skills/moai/workflows")).unwrap();
    fs::create_dir_all(root.path().join(".moai-cli")).unwrap();
    let decision = |id: &str, value: &str| json!({"id": id, "topic": id, "value": value, "source": "user", "scope": "project"});
    let mut profile = json!({"schema_version": 2, "project_id": "demo", "purpose": "Build a small editor",
        "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude"],
        "interview_revision": "r1",
        "decisions": [decision("review_blocks", "a change without a test"),
                      decision("protected", "the billing module")]});
    if let Some(l) = language {
        profile["language"] = l;
    }
    let mut bytes = serde_json::to_vec_pretty(&profile).unwrap();
    bytes.push(b'\n');
    fs::write(root.path().join(PROFILE), bytes).unwrap();
    root
}

/// A temporary HOME; with `saved`, the installer's `language.yaml` holds that value.
fn home(saved: Option<&str>) -> TempDir {
    let home = TempDir::new().unwrap();
    if let Some(code) = saved {
        let dir = home.path().join(".moai-cli/config");
        fs::create_dir_all(&dir).unwrap();
        fs::write(
            dir.join("language.yaml"),
            format!("language:\n  conversation_language: {code}\n"),
        )
        .unwrap();
    }
    home
}

fn run_in(home: &Path, args: &[&str], stdin: &[u8]) -> Output {
    let mut child = Command::new(EXE)
        .env("HOME", home)
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

fn run(args: &[&str], stdin: &[u8]) -> Output {
    run_in(home(None).path(), args, stdin)
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

fn path(root: &TempDir) -> &str {
    root.path().to_str().unwrap()
}

fn questions(root: &TempDir, home: &Path, extra: &[&str]) -> Value {
    let mut args = vec!["language-setup", path(root), "--questions"];
    args.extend_from_slice(extra);
    ok(&run_in(home, &args, b""))["questions"].clone()
}

fn ids(q: &Value) -> Vec<String> {
    q["items"]
        .as_array()
        .unwrap()
        .iter()
        .map(|i| i["id"].as_str().unwrap().to_string())
        .collect()
}

fn item<'a>(q: &'a Value, id: &str) -> &'a Value {
    q["items"]
        .as_array()
        .unwrap()
        .iter()
        .find(|i| i["id"] == id)
        .unwrap_or_else(|| panic!("no item {id}: {q}"))
}

fn setup_answers(root: &TempDir, answers: &Value, extra: &[&str]) -> Value {
    let mut args = vec!["language-setup", path(root), "--answers"];
    args.extend_from_slice(extra);
    ok(&run(&args, answers.to_string().as_bytes()))
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

fn profile_value(root: &Path) -> Value {
    serde_json::from_slice(&fs::read(root.join(PROFILE)).unwrap()).unwrap()
}

fn agents_request(language: &str) -> Value {
    json!({"protocol_version": 1, "language": language,
        "workflow_reference": "skills/moai/workflows/", "local_file": "none"})
}

fn agents_plan(root: &TempDir, language: &str) -> Value {
    ok(&run(
        &["agents-plan", path(root)],
        agents_request(language).to_string().as_bytes(),
    ))
}

fn agents_apply(root: &TempDir, plan: &Value) {
    ok(&run(
        &[
            "agents-apply",
            path(root),
            "--plan-digest",
            plan["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "fixture-approved",
        ],
        plan.to_string().as_bytes(),
    ));
}

/// Plan and apply AGENTS.md and REVIEW.md for the request language; returns the plan.
fn render(root: &TempDir, language: &str) -> Value {
    let plan = agents_plan(root, language);
    agents_apply(root, &plan);
    plan
}

fn read(root: &TempDir, rel: &str) -> String {
    fs::read_to_string(root.path().join(rel)).unwrap()
}

const KO_AGENTS: &str = "## MoAI 작업 흐름 (관리 구간)";
const EN_AGENTS: &str = "## MoAI workflow (managed section)";
const KO_REVIEW: &str = "# 검토 지침";
const EN_REVIEW: &str = "# Review instructions";

/// AC-1 (REQ-1): the question set holds exactly the three items, in order, each with the
/// current value, the suggested value and the default.
#[test]
fn questions_list_exactly_three_items_with_their_values() {
    let root = project(None);
    let h = home(None);
    let q = questions(&root, h.path(), &[]);
    assert_eq!(
        ids(&q),
        ["commit_messages", "instruction_files", "documents"],
        "{q}"
    );
    for id in ["commit_messages", "instruction_files", "documents"] {
        let i = item(&q, id);
        for key in ["current", "suggested", "default"] {
            assert!(i.get(key).is_some(), "{id} has no {key}: {i}");
        }
        assert_eq!(i["current"], json!(null), "{id}");
        assert_eq!(
            i["suggested"],
            json!("en"),
            "{id}: no conversation language yet"
        );
    }
    // The conversation language is never asked here: it comes from the install step, and the
    // items above are the whole set.
    assert_eq!(q["items"].as_array().unwrap().len(), 3, "{q}");
    // With a conversation language the three items suggest it; the item whose default is
    // English keeps that default.
    let q = questions(&root, h.path(), &["--conversation", "ko"]);
    for id in ["commit_messages", "documents"] {
        assert_eq!(item(&q, id)["suggested"], json!("ko"), "{id}");
        assert_eq!(item(&q, id)["default"], json!("ko"), "{id}");
    }
    assert_eq!(item(&q, "instruction_files")["suggested"], json!("en"));
    assert_eq!(item(&q, "instruction_files")["default"], json!("en"));
    // A stored value is the item's current and suggested value; the stored `user_chat` does
    // not feed the conversation default any more, so the English default stands.
    let stored = project(Some(json!({"documents": "ja", "user_chat": "ko"})));
    let q = questions(&stored, h.path(), &[]);
    assert_eq!(item(&q, "documents")["current"], json!("ja"));
    assert_eq!(item(&q, "documents")["suggested"], json!("ja"));
    assert_eq!(item(&q, "commit_messages")["current"], json!(null));
    assert_eq!(item(&q, "commit_messages")["suggested"], json!("en"));
}

/// AC-2 (REQ-2): the language the installer saved is the conversation value with its source,
/// and is not asked again from scratch. A profile's earlier `user_chat` is not a source any
/// more: with nothing in HOME the conversation resolves to English, and the profile's value
/// stays as recorded data.
#[test]
fn the_installer_value_is_the_confirmed_conversation_default() {
    let root = project(None);
    let h = home(Some("ko"));
    let q = questions(&root, h.path(), &[]);
    assert_eq!(q["conversation_language"], json!("ko"), "{q}");
    assert_eq!(q["conversation_source"], json!("installer"), "{q}");
    assert_eq!(item(&q, "documents")["suggested"], json!("ko"));
    // A profile's earlier user_chat is not a source: English without an install value.
    let stored = project(Some(json!({"user_chat": "ja"})));
    let none = home(None);
    let q = questions(&stored, none.path(), &[]);
    assert_eq!(q["conversation_language"], json!("en"), "{q}");
    assert_eq!(q["conversation_source"], json!("default"), "{q}");
    // The value the installer saved in HOME decides over anything the profile holds.
    let q = questions(&stored, h.path(), &[]);
    assert_eq!(q["conversation_language"], json!("ko"), "{q}");
    assert_eq!(q["conversation_source"], json!("installer"), "{q}");
    // The explicit flag is ahead of both.
    let q = questions(&stored, h.path(), &["--conversation", "zh"]);
    assert_eq!(q["conversation_source"], json!("flag"), "{q}");
}

/// AC-2 (REQ-3): a saved value that is missing or has no name falls to the next source, named.
#[test]
fn a_bad_installer_value_falls_to_the_next_source() {
    let root = project(None);
    let none = home(None);
    let q = questions(&root, none.path(), &[]);
    assert_eq!(q["conversation_source"], json!("default"), "{q}");
    assert_eq!(q["conversation_language"], json!("en"), "{q}");
    // A language the setup has no name for (`fr`) is skipped.
    let french = home(Some("fr"));
    let q = questions(&root, french.path(), &[]);
    assert_eq!(q["conversation_source"], json!("default"), "{q}");
    assert_eq!(q["conversation_language"], json!("en"), "{q}");
    // The previous generation's backup is the next source after the installer's value.
    let go = project(None);
    let dir = go.path().join(".moai-backup/config/sections");
    fs::create_dir_all(&dir).unwrap();
    fs::write(
        dir.join("language.yaml"),
        "language:\n  conversation_language: \"ja\"\n",
    )
    .unwrap();
    let q = questions(&go, french.path(), &[]);
    assert_eq!(q["conversation_source"], json!("go_backup"), "{q}");
    assert_eq!(q["conversation_language"], json!("ja"), "{q}");
}

/// AC-3 (REQ-4): answers name the three items and keep every other key the profile had; a fresh
/// profile gets no key for an item that is not asked.
#[test]
fn answers_name_the_three_items_and_keep_old_keys() {
    let fresh = project(None);
    let out = setup_answers(
        &fresh,
        &json!({"documents": "en"}),
        &["--conversation", "ko"],
    );
    assert_eq!(
        out["language"],
        json!({"commit_messages": "ko", "documents": "en", "instruction_files": "en",
               "user_chat": "ko"}),
        "{out}"
    );
    assert_eq!(out["profile_written"], json!(false));
    let old = project(Some(json!({"code_comments": "ja", "reports": "ko"})));
    let out = setup_answers(&old, &json!({}), &["--conversation", "ko"]);
    assert_eq!(
        out["language"],
        json!({"code_comments": "ja", "commit_messages": "ko", "documents": "ko",
               "instruction_files": "en", "reports": "ko", "user_chat": "ko"}),
        "{out}"
    );
    // The old reports value is the suggestion for documents, not a copy of the key's meaning.
    let old = project(Some(json!({"user_chat": "ko", "reports": "ja"})));
    let q = questions(&old, home(None).path(), &[]);
    assert_eq!(item(&q, "documents")["current"], json!(null), "{q}");
    assert_eq!(item(&q, "documents")["suggested"], json!("ja"), "{q}");
    // A re-run with no answers and no HOME record moves the resolved user_chat to English once
    // (the ask-once rule: the profile's stored value is not a source), and every re-run after
    // that keeps every value.
    let first = setup_answers(&fresh, &json!({}), &["--conversation", "ko"]);
    apply_profile(&fresh, &first["profile"]);
    let again = setup_answers(&fresh, &json!({}), &[]);
    assert_eq!(again["changed_items"], json!(["user_chat"]), "{again}");
    let mut moved = first["language"].clone();
    moved["user_chat"] = json!("en");
    assert_eq!(again["language"], moved);
    apply_profile(&fresh, &again["profile"]);
    let stable = setup_answers(&fresh, &json!({}), &[]);
    assert_eq!(stable["changed_items"], json!([]), "{stable}");
    assert_eq!(stable["language"], moved);
    // An answer for an item that is neither asked nor an old key is still refused.
    let refused = run(
        &["language-setup", path(&fresh), "--answers"],
        json!({"emoji": "en"}).to_string().as_bytes(),
    );
    assert!(!refused.status.success());
}

/// AC-4 (REQ-5): the terminal step asks three prompts, one answer line each.
#[test]
fn the_terminal_step_asks_three_prompts() {
    let root = project(None);
    let out = run(&["language-setup", path(&root)], b"ko\n\nja\nzh\nzh\n");
    let l = ok(&out)["language"].clone();
    assert_eq!(
        l,
        json!({"commit_messages": "ko", "documents": "ja", "instruction_files": "en",
               "user_chat": "en"}),
        "{l}"
    );
    let stderr = String::from_utf8_lossy(&out.stderr).into_owned();
    assert_eq!(
        stderr.matches("(en/ko/ja/zh)").count(),
        3,
        "three prompts: {stderr}"
    );
}

/// AC-5 (REQ-6): the managed section lists `documents` and says the rest is English.
#[test]
fn the_managed_section_names_documents_and_the_english_rest() {
    let root = project(Some(
        json!({"user_chat": "en", "commit_messages": "ko", "documents": "ko"}),
    ));
    render(&root, "en");
    let text = read(&root, "AGENTS.md");
    assert!(
        text.contains("- Documents (reports and SPEC documents): Korean (`ko`)"),
        "{text}"
    );
    assert!(text.contains("- Commit messages: Korean (`ko`)"), "{text}");
    assert!(
        text.contains("instruction files, skills and rules, messages between agents, code comments, issue and PR records")
            && text.contains("English unless a line above says otherwise"),
        "{text}"
    );
    // The same in Korean when the conversation language is Korean.
    let ko = project(Some(json!({"user_chat": "ko", "documents": "ko"})));
    render(&ko, "en");
    let text = read(&ko, "AGENTS.md");
    assert!(
        text.contains("- 문서(보고서·SPEC): 한국어 (`ko`)"),
        "{text}"
    );
    assert!(
        text.contains("지침 파일, 스킬·규칙, 에이전트 간 메시지, 코드 주석, Issue·PR 기록")
            && text.contains("영어"),
        "{text}"
    );
    // A profile without `language` still renders no language lines at all.
    let plain = project(None);
    render(&plain, "en");
    let text = read(&plain, "AGENTS.md");
    assert!(
        !text.contains("English unless a line above says otherwise"),
        "{text}"
    );
}

/// AC-6 (REQ-7): AGENTS.md and REVIEW.md follow `user_chat` for `en` and `ko`, whatever the
/// request language is.
#[test]
fn agents_and_review_follow_the_conversation_language() {
    let ko = project(Some(json!({"user_chat": "ko"})));
    render(&ko, "en");
    let (agents, review) = (read(&ko, "AGENTS.md"), read(&ko, "REVIEW.md"));
    assert!(
        agents.contains(KO_AGENTS) && !agents.contains(EN_AGENTS),
        "{agents}"
    );
    assert!(review.starts_with(KO_REVIEW), "{review}");
    let en = project(Some(json!({"user_chat": "en"})));
    render(&en, "ko");
    let (agents, review) = (read(&en, "AGENTS.md"), read(&en, "REVIEW.md"));
    assert!(
        agents.contains(EN_AGENTS) && !agents.contains(KO_AGENTS),
        "{agents}"
    );
    assert!(review.starts_with(EN_REVIEW), "{review}");
    // A re-run is unchanged.
    let again = agents_plan(&en, "ko");
    assert!(
        again["files"]
            .as_array()
            .unwrap()
            .iter()
            .all(|f| f["action"] == json!("unchanged")),
        "{again}"
    );
}

/// AC-6 (REQ-9): an explicit `instruction_files` still decides, ahead of `user_chat`.
#[test]
fn an_explicit_instruction_files_still_wins() {
    let root = project(Some(json!({"user_chat": "ko", "instruction_files": "en"})));
    let plan = render(&root, "ko");
    let agents = read(&root, "AGENTS.md");
    assert!(
        agents.contains(EN_AGENTS) && !agents.contains(KO_AGENTS),
        "{agents}"
    );
    assert!(read(&root, "REVIEW.md").starts_with(EN_REVIEW));
    assert!(
        !plan["warnings"].to_string().contains("language.user_chat"),
        "{plan}"
    );
}

/// AC-7 (REQ-8): another conversation language writes in the request language and warns.
#[test]
fn another_conversation_language_falls_back_with_a_warning() {
    let root = project(Some(json!({"user_chat": "ja"})));
    let plan = render(&root, "ko");
    assert!(read(&root, "AGENTS.md").contains(KO_AGENTS));
    assert!(read(&root, "REVIEW.md").starts_with(KO_REVIEW));
    assert!(
        plan["warnings"]
            .to_string()
            .contains("no managed-section text in profile language.user_chat `ja`"),
        "{plan}"
    );
}

/// AC-12 (REQ-13): an `instruction_files` code with no managed-section text hands the choice to
/// the conversation language and warns, naming `instruction_files` and its code.
#[test]
fn an_unsupported_instruction_files_code_hands_over_to_the_conversation_language() {
    let root = project(Some(json!({"user_chat": "ko", "instruction_files": "ja"})));
    let plan = render(&root, "en");
    assert!(read(&root, "AGENTS.md").contains(KO_AGENTS));
    assert!(read(&root, "REVIEW.md").starts_with(KO_REVIEW));
    let warnings = plan["warnings"].to_string();
    assert!(
        warnings.contains("rendered in `ko`")
            && warnings.contains("language.instruction_files `ja`"),
        "{plan}"
    );
    // Without a usable conversation language the request language is used.
    let root = project(Some(json!({"user_chat": "zh", "instruction_files": "ja"})));
    render(&root, "en");
    assert!(read(&root, "AGENTS.md").contains(EN_AGENTS));
}

/// AC-8 (REQ-10): a profile with the old eight items plans and applies without a conflict and
/// keeps its own values.
#[test]
fn a_profile_with_the_old_eight_items_plans_and_applies_without_conflict() {
    let eight = json!({"user_chat": "ko", "reports": "ko", "issue_pr_records": "ko",
        "commit_messages": "ko", "agent_messages": "en", "instruction_files": "en",
        "skill_instructions": "en", "code_comments": "en"});
    let root = project(None);
    let mut profile = profile_value(root.path());
    profile["language"] = eight.clone();
    let plan = apply_profile(&root, &profile);
    assert_ne!(plan["action"], json!("conflict"), "{plan}");
    assert_eq!(profile_value(root.path())["language"], eight);
    let plan = render(&root, "ko");
    assert!(
        plan["files"]
            .as_array()
            .unwrap()
            .iter()
            .all(|f| f["action"] != json!("conflict")),
        "{plan}"
    );
    // The explicit `instruction_files: en` is still honored; every old key shows as its line.
    let agents = read(&root, "AGENTS.md");
    assert!(agents.contains(EN_AGENTS), "{agents}");
    assert!(
        agents.contains("- Reports to the user: Korean (`ko`)"),
        "{agents}"
    );
    assert!(
        agents.contains("- Code comments: English (`en`)"),
        "{agents}"
    );
    // The setup step over it keeps every old key and adds `documents` (the old `reports` value
    // is its suggestion). The stored user_chat is not a source of the conversation any more, so
    // the resolved English is written over it.
    let out = setup_answers(&root, &json!({}), &[]);
    for (k, v) in eight.as_object().unwrap() {
        if k == "user_chat" {
            assert_eq!(out["language"][k], json!("en"), "{k}: {out}");
        } else {
            assert_eq!(out["language"][k], *v, "{k}: {out}");
        }
    }
    assert_eq!(
        out["language"]["documents"],
        json!("ko"),
        "{out}: the old reports value is its suggestion"
    );
    assert_eq!(
        out["changed_items"],
        json!(["user_chat", "documents"]),
        "{out}"
    );
}

fn github_notes(root: &Path, request: Value) -> String {
    let out = Command::new(EXE)
        .args(["github-plan", root.to_str().unwrap()])
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .and_then(|mut c| {
            c.stdin
                .take()
                .unwrap()
                .write_all(request.to_string().as_bytes())?;
            c.wait_with_output()
        })
        .unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"]["notes"].to_string()
}

/// AC-9 (REQ-11): when the profile sets `language`, the form language is English unless
/// `issue_pr_records` is set; a profile with no `language` keeps the Korean forms it always had.
#[test]
fn the_form_language_is_english_unless_issue_pr_records_is_set() {
    let root = project(None);
    let mut profile = profile_value(root.path());
    profile["github_issues"] = json!(true);
    fs::write(
        root.path().join(PROFILE),
        serde_json::to_vec_pretty(&profile).unwrap(),
    )
    .unwrap();
    let notes = github_notes(root.path(), json!({"protocol_version": 1}));
    assert!(
        notes.contains("form language ko"),
        "no language set: {notes}"
    );
    for (language, expected) in [
        (json!({"user_chat": "ko"}), "en"),
        (json!({"user_chat": "ko", "documents": "ko"}), "en"),
        (json!({"user_chat": "ko", "issue_pr_records": "ko"}), "ko"),
        (json!({"user_chat": "en", "issue_pr_records": "ko"}), "ko"),
    ] {
        let root = project(Some(language.clone()));
        let mut profile = profile_value(root.path());
        profile["github_issues"] = json!(true);
        fs::write(
            root.path().join(PROFILE),
            serde_json::to_vec_pretty(&profile).unwrap(),
        )
        .unwrap();
        let notes = github_notes(root.path(), json!({"protocol_version": 1}));
        assert!(
            notes.contains(&format!("form language {expected}")),
            "{language}: {notes}"
        );
    }
}
