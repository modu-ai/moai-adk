//! SPEC-20261121-worktreeinclude-personal-rules: a worktree a host makes starts from tracked
//! files only, so the personal files are missing there. The `worktree-include` pick plans one
//! managed block in the project's `.worktreeinclude` whose lines are one fixed list, read from a
//! packaged template and the same for every project. Nothing is planned without the pick, and the
//! user's own lines stay as they are.

mod common;

use common::cli::{create, ok_result, profile};

use serde_json::{json, Value};
use std::fs;
use std::io::Write;
use std::path::Path;
use std::process::{Command, Stdio};
use tempfile::TempDir;

const FILE: &str = ".worktreeinclude";
const TEMPLATE: &str = "templates/worktreeinclude/personal.txt";

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

fn request(local_file: &str, picks: &[&str]) -> Value {
    let mut r = json!({
        "protocol_version": 1,
        "language": "en",
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

/// A project with a profile and the workflow folder, nothing applied yet.
fn project() -> TempDir {
    let root = TempDir::new().unwrap();
    fs::create_dir_all(root.path().join("skills/moai/workflows")).unwrap();
    create(&root, &profile(None));
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

/// The template's file lines: every line that is neither blank nor a comment.
fn template_paths() -> Vec<String> {
    let text = fs::read_to_string(Path::new(env!("CARGO_MANIFEST_DIR")).join(TEMPLATE))
        .expect("the packaged template exists");
    file_lines(&text)
}

fn file_lines(text: &str) -> Vec<String> {
    text.lines()
        .map(str::trim)
        .filter(|l| !l.is_empty() && !l.starts_with('#'))
        .map(String::from)
        .collect()
}

/// The file lines between the managed markers of `text`.
fn block_paths(text: &str) -> Vec<String> {
    let mut inside = false;
    let mut out = String::new();
    for l in text.lines() {
        if l.starts_with("# moai:managed:begin") {
            inside = true;
        } else if l.starts_with("# moai:managed:end") {
            inside = false;
        } else if inside {
            out.push_str(l);
            out.push('\n');
        }
    }
    file_lines(&out)
}

/// AC-1: with the pick, the plan carries one managed block with the template's lines, shown in
/// the row's preview; the plan writes nothing.
#[test]
fn the_pick_plans_one_managed_block_with_the_template_lines() {
    let root = project();
    let plan = plan_ok(&root, &request("ignored", &["worktree-include"]));
    let r = row(&plan, FILE).unwrap_or_else(|| panic!("no row for {FILE}: {plan}"));
    assert_eq!(r["action"], json!("create"), "{r}");
    let detail = r["detail"].as_str().unwrap_or_default();
    let paths = template_paths();
    assert!(!paths.is_empty());
    for p in &paths {
        assert!(detail.contains(&format!("+ {p}")), "{p} in {detail}");
    }
    assert!(!root.path().join(FILE).exists(), "a plan writes nothing");
}

/// AC-2: the block is the same for every project: whether the personal file is ignored, tracked
/// or not created, the applied block holds the template's lines and nothing else.
#[test]
fn the_block_is_the_same_whatever_the_local_file_setting() {
    let expected = template_paths();
    for local in ["ignored", "tracked", "none"] {
        let root = project();
        let plan = plan_ok(&root, &request(local, &["worktree-include"]));
        apply_ok(&root, &plan);
        assert_eq!(block_paths(&read(&root, FILE)), expected, "{local}");
    }
}

/// AC-3: the template names exactly the personal paths the code names, and no secret file.
#[test]
fn the_template_lists_exactly_the_personal_paths_and_no_secret() {
    let mut in_template = template_paths();
    let mut in_code: Vec<String> = moai_cli::agents::PERSONAL_PATHS
        .iter()
        .map(|s| s.to_string())
        .collect();
    in_template.sort();
    in_code.sort();
    assert_eq!(in_template, in_code, "template and code lists differ");
    assert!(in_code.contains(&"AGENTS.local.md".to_string()));
    assert!(in_code.contains(&".claude/settings.local.json".to_string()));

    let text = fs::read_to_string(Path::new(env!("CARGO_MANIFEST_DIR")).join(TEMPLATE)).unwrap();
    let secret = moai_cli::worktree_include::secret_lines(&text);
    assert!(secret.is_empty(), "{secret:?}");
}

/// AC-4: applying writes the block; the user's own lines stay byte for byte, and a second plan
/// finds nothing to do.
#[test]
fn applying_keeps_every_user_line_and_a_second_plan_is_unchanged() {
    let root = project();
    // The user's file has no final newline and a CRLF line: both stay as they are.
    let user = "# mine\r\n.env.example\r\nbuild/cache/keep.txt";
    fs::write(root.path().join(FILE), user).unwrap();
    let plan = plan_ok(&root, &request("ignored", &["worktree-include"]));
    assert_eq!(
        row(&plan, FILE).unwrap()["action"],
        json!("append"),
        "{plan}"
    );
    apply_ok(&root, &plan);
    let after = read(&root, FILE);
    assert!(after.starts_with(user), "user bytes first: {after:?}");
    assert_eq!(block_paths(&after), template_paths(), "{after:?}");
    assert_eq!(
        after.matches("# moai:managed:begin").count(),
        1,
        "one block: {after:?}"
    );

    let again = plan_ok(&root, &request("ignored", &["worktree-include"]));
    assert_eq!(
        row(&again, FILE).unwrap()["action"],
        json!("unchanged"),
        "{again}"
    );
}

/// AC-5: the pick is opt-in: without it nothing is planned.
#[test]
fn without_the_pick_nothing_is_planned() {
    for local in ["ignored", "tracked", "none"] {
        let root = project();
        let plan = plan_ok(&root, &request(local, &[]));
        assert!(row(&plan, FILE).is_none(), "{local}: {plan}");
    }
}

/// AC-6: the apply keeps the file as it was before, beside the receipt, and the result names it.
#[test]
fn the_apply_receipt_keeps_the_file_as_it_was_and_names_it() {
    let root = project();
    let user = "# mine\n.env.example\n";
    fs::write(root.path().join(FILE), user).unwrap();
    let plan = plan_ok(&root, &request("ignored", &["worktree-include"]));
    let applied = apply_ok(&root, &plan);
    assert!(applied.to_string().contains(FILE), "{applied}");
    // The receipt names its request; the before bytes sit beside it, under the receipts folder.
    let id = applied["receipt"]
        .as_str()
        .expect("the result names its receipt");
    let kept = root
        .path()
        .join(moai_cli::apply::receipts_rel(root.path()))
        .join(format!("{id}.before"))
        .join(FILE);
    assert_eq!(
        fs::read_to_string(&kept).unwrap(),
        user,
        "{}",
        kept.display()
    );
}

/// AC-7: the new pick is a known id, an unknown one is refused, and the schema lists the same
/// ids as the code.
#[test]
fn the_pick_is_a_known_id_and_an_unknown_one_is_refused() {
    let root = project();
    let out = cli(
        &["agents-plan", root.path().to_str().unwrap()],
        &request("ignored", &["no-such-pick"]),
    );
    assert!(!out.status.success());
    let ok = plan_ok(&root, &request("ignored", &["worktree-include"]));
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
    assert!(in_code.contains(&"worktree-include".to_string()));
}

/// AC-9: the applied file names no secret, so the read-only secret warning of the file stays
/// quiet.
#[test]
fn the_applied_file_raises_no_secret_warning() {
    let root = project();
    let plan = plan_ok(&root, &request("ignored", &["worktree-include"]));
    apply_ok(&root, &plan);
    let warnings = moai_cli::worktree_include::warnings(root.path());
    assert!(warnings.is_empty(), "{warnings:?}");
}
