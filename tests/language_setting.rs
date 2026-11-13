//! Per-project writing language by item (Issue #2293), end to end through the binary in
//! temporary projects only: profile validation (AC-1), the setup input step feeding
//! `project-plan`/`project-apply` (AC-2) and the AGENTS.md managed section rendered from the
//! setting through `agents-*` and `generation-*` (AC-3).

use serde_json::{json, Value};
use std::{
    fs,
    io::Write,
    path::Path,
    process::{Command, Output, Stdio},
};
use tempfile::TempDir;

const EXE: &str = env!("CARGO_BIN_EXE_moai-cli");
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
    let mut child = Command::new(EXE)
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

fn agents_md(root: &TempDir) -> String {
    fs::read_to_string(root.path().join("AGENTS.md")).unwrap()
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
/// with no answers keeps every value (plan `unchanged`, file bytes identical).
#[test]
fn setup_answers_reach_the_profile_and_rerun_keeps_them() {
    let root = project(None);
    let before = fs::read(root.path().join(PROFILE)).unwrap();
    let out = setup_answers(
        &root,
        &json!({"code_comments": "ja", "user_chat": "ko"}),
        &["--conversation", "ko"],
    );
    // The step itself writes nothing.
    assert_eq!(fs::read(root.path().join(PROFILE)).unwrap(), before);
    assert_eq!(out["profile_written"], json!(false));
    let expected = json!({
        "agent_messages": "en", "code_comments": "ja", "commit_messages": "ko",
        "instruction_files": "en", "issue_pr_records": "ko", "reports": "ko",
        "skill_instructions": "en", "user_chat": "ko"
    });
    assert_eq!(out["language"], expected);
    let plan = apply_profile(&root, &out["profile"]);
    assert_eq!(plan["action"], json!("update"));
    assert_eq!(plan["delta"]["added"], json!(["language"]));
    assert_eq!(profile_value(root.path())["language"], expected);
    let applied = fs::read(root.path().join(PROFILE)).unwrap();

    // Re-run: no answers, no --conversation. Every value is kept.
    let again = setup_answers(&root, &json!({}), &[]);
    assert_eq!(again["changed_items"], json!([]));
    assert_eq!(again["language"], expected);
    let replan = ok(&run(
        &["project-plan", path(&root)],
        again["profile"].to_string().as_bytes(),
    ));
    assert_eq!(replan["action"], json!("unchanged"));
    apply_profile(&root, &again["profile"]);
    assert_eq!(fs::read(root.path().join(PROFILE)).unwrap(), applied);

    // The question set shows the stored values as current and suggested.
    let q = ok(&run(&["language-setup", path(&root), "--questions"], b""));
    let items = q["questions"]["items"].as_array().unwrap();
    assert_eq!(items.len(), 8);
    let comments = items.iter().find(|i| i["id"] == "code_comments").unwrap();
    assert_eq!(comments["current"], json!("ja"));
    assert_eq!(comments["suggested"], json!("ja"));
    assert_eq!(comments["default"], json!("en"));

    // One changed answer changes exactly that item.
    let third = setup_answers(&root, &json!({"reports": "en"}), &[]);
    assert_eq!(third["changed_items"], json!(["reports"]));
    let plan = apply_profile(&root, &third["profile"]);
    assert_eq!(plan["delta"]["changed"], json!(["language"]));
    assert_eq!(
        profile_value(root.path())["language"]["reports"],
        json!("en")
    );
}

/// AC-2: without a question tool the step asks on the terminal: an empty line keeps the
/// suggested value, end of input keeps the rest, a bad code is asked again.
#[test]
fn interactive_setup_reads_one_answer_per_item() {
    let root = project(Some(json!({"commit_messages": "zh"})));
    let out = ok(&run(&["language-setup", path(&root)], b"ko\n\nfr\nja\n"));
    let l = &out["language"];
    assert_eq!(l["user_chat"], json!("ko"));
    // The conversation language is fixed before asking (no flag, no stored user_chat: en).
    assert_eq!(l["reports"], json!("en"));
    assert_eq!(l["issue_pr_records"], json!("ja"));
    assert_eq!(l["commit_messages"], json!("zh"));
    assert_eq!(l["code_comments"], json!("en"));
    let stderr = String::from_utf8(run(&["language-setup", path(&root)], b"fr\n").stderr).unwrap();
    assert!(stderr.contains("unsupported language code"), "{stderr}");
    let refused = run(&["language-setup", path(&root)], b"fr\nde\nes\n");
    assert!(failed(&refused).contains("too many unsupported answers"));
    assert!(!String::from_utf8_lossy(&refused.stdout).contains("\"de\""));
}

/// The managed block of a profile without `language`, pinned: the #2293 change must leave
/// existing projects' AGENTS.md byte-identical. #2385 deliberately adds one line (the
/// pointer to the VERIFY workflow's "Check-run rules"), so existing projects get it as an
/// `update` on their next plan. #2619 replaces the line on deciding and asking with the sorting order of the
/// skill (facts answered, one-way and guarded choices to the user, the rest decided and recorded).
const PLAIN_EN: &str = "<!-- moai:managed:begin id=moai-workflow -->
## MoAI workflow (managed section)

Project `demo` · tools: Rust · methods: TDD · hosts: claude, codex.
Purpose: Build a small editor

Confirm the approved SPEC and Issue scope, then proceed plan → implement → real verification → independent audit → report.
Answer a factual question yourself and cite the evidence. A choice that cannot be undone, that goes beyond the allowed scope (cost, permissions, external transmission), or that touches money, legal terms, external contracts, secrets, real data or a guarded setting (permissions, keys, the real HOME and global settings, deletions, protection settings) waits for the user; decide any other choice within the allowed scope and record it.
Never report an unexecuted check as PASS. Distinguish passed, failed, not_run, blocked and stale.
Which checks run locally, and when an exact-head CI result may be cited instead, follow the \"Check-run rules\" section of the VERIFY workflow (the single source).
Without the MoAI binary, do what native tools allow. When a required gate has no verified substitute, block the dependent action and report blocked.
Read only the workflow you need from `skills/moai/workflows/`.
Never share the personal `AGENTS.local.md`, credentials or raw secret logs in the repo or in Issues.
<!-- moai:managed:end id=moai-workflow -->
";

#[test]
fn absent_language_renders_the_earlier_managed_block() {
    let root = project(None);
    plan_then_apply(&root, "agents", &agents_request("en"));
    assert_eq!(agents_md(&root), format!("# demo\n\n{PLAIN_EN}"));
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
    let text = agents_md(&korean);
    let block = "항목별 작성 언어(`.moai-cli/project.yaml`의 `language`):
- 사용자 대화: 한국어 (`ko`)
- 사용자 보고: 한국어 (`ko`)
- Issue·PR 기록: 한국어 (`ko`)
- 커밋 메시지: 한국어 (`ko`)
- 에이전트 간 메시지: 영어 (`en`)
- 지침 파일(AGENTS.md, 규칙): 한국어 (`ko`)
- skill 지침: 영어 (`en`)
- 코드 주석: 영어 (`en`)

승인한 SPEC과";
    assert!(text.contains(block), "{text}");
    // #2619: the Korean section carries the same sorting order as the English one.
    assert!(
        text.contains("사실 확인은 스스로 하고 근거를 남긴다. 되돌릴 수 없거나, 허용 범위를 넘거나(비용·권한·외부 전송), 돈·법·외부 계약·비밀·실데이터·보호 대상 설정(권한·키·실제 HOME과 전역 설정·삭제·보호 설정)에 닿는 선택은 사용자를 기다리고, 나머지는 허용 범위 안에서 정한 뒤 기록한다."),
        "{text}"
    );
    // The managed section follows instruction_files, so nothing is reported.
    assert!(!plan["warnings"].to_string().contains("instruction_files"));

    // A second project, partial setting: unset items fall back to their defaults.
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
    let text = agents_md(&japanese);
    let block = "Purpose: Build a small editor

Writing language by item (`language` in `.moai-cli/project.yaml`):
- Conversation with the user: Japanese (`ja`)
- Reports to the user: English (`en`)
- Issue and PR records: English (`en`)
- Commit messages: English (`en`)
- Messages between agents: English (`en`)
- Instruction files (AGENTS.md, rules): English (`en`)
- Skill instructions: English (`en`)
- Code comments: Chinese (`zh`)

Confirm the approved";
    assert!(text.contains(block), "{text}");
    assert!(!agents_md(&korean).contains("Writing language by item"));

    // Changing the setting and re-applying rewrites only the managed lines.
    let mut profile = profile_value(japanese.path());
    profile["language"]["code_comments"] = json!("ja");
    apply_profile(&japanese, &profile);
    plan_then_apply(&japanese, "agents", &agents_request("en"));
    assert!(agents_md(&japanese).contains("- Code comments: Japanese (`ja`)\n"));
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
    let text = agents_md(&root);
    assert!(text.contains("## MoAI 작업 흐름 (관리 구간)"), "{text}");
    assert!(!text.contains("## MoAI workflow (managed section)"));
    // The writing-language lines speak the section's language too (mutation G3).
    assert!(
        text.contains("- 지침 파일(AGENTS.md, 규칙): 한국어 (`ko`)"),
        "{text}"
    );
    assert!(!text.contains("Writing language by item"));
    // User text outside the markers survives a re-run, and the re-run plans unchanged.
    let with_note = format!("{text}\n## Our notes\n\nkeep me\n");
    std::fs::write(root.path().join("AGENTS.md"), &with_note).unwrap();
    let again = plan_then_apply(&root, "agents", &agents_request("en"));
    assert!(
        again["files"]
            .as_array()
            .unwrap()
            .iter()
            .any(|f| f["path"] == json!("AGENTS.md") && f["action"] == json!("unchanged")),
        "{again}"
    );
    assert_eq!(agents_md(&root), with_note);

    // Changing the stored language rewrites only the managed section.
    let mut profile = profile_value(root.path());
    profile["language"]["instruction_files"] = json!("en");
    apply_profile(&root, &profile);
    plan_then_apply(&root, "agents", &agents_request("ko"));
    let text = agents_md(&root);
    assert!(
        text.contains("## MoAI workflow (managed section)"),
        "{text}"
    );
    assert!(text.ends_with("## Our notes\n\nkeep me\n"));

    // `ja` has no managed-section text: the request language, with a warning.
    let japanese = project(Some(json!({"instruction_files": "ja"})));
    let plan = plan_then_apply(&japanese, "agents", &agents_request("ko"));
    assert!(agents_md(&japanese).contains("## MoAI 작업 흐름 (관리 구간)"));
    assert!(
        plan["warnings"]
            .to_string()
            .contains("no managed-section text in profile language.instruction_files `ja`"),
        "{plan}"
    );
}

/// #2696 C2: in a migrated project (the profile under `.moai/`, marker present) the managed
/// section's heading names `.moai/project.yaml`, where the setting actually lives.
#[test]
fn a_migrated_project_names_its_own_profile_path() {
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
    let text = agents_md(&root);
    assert!(
        text.contains("Writing language by item (`language` in `.moai/project.yaml`):"),
        "{text}"
    );
    assert!(!text.contains(".moai-cli/project.yaml"), "{text}");
    // AUDIT #2794 B2: the preview renders the same section the apply wrote — a re-plan right
    // after the apply finds AGENTS.md `unchanged` (a plan that named another profile path
    // would render a different block and plan an update).
    let again = ok(&run(
        &["agents-plan", path(&root)],
        agents_request("en").to_string().as_bytes(),
    ));
    let agents = again["files"]
        .as_array()
        .unwrap()
        .iter()
        .find(|f| f["path"] == "AGENTS.md")
        .unwrap_or_else(|| panic!("{again}"))
        .clone();
    assert_eq!(agents["action"], json!("unchanged"), "{again}");
}
