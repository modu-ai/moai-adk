//! SPEC-20261119-spec-branch-name-first (card C-112; user decision 2026-11-19): a SPEC work
//! branch is named `spec/<slug>-<yyyymmdd>` (name first) and every reader still resolves the old
//! `spec/<yyyymmdd>-<slug>` to the same SPEC id, with or without a repair suffix `-r<N>`; a name
//! that reads as two SPEC ids is refused with its reason, never guessed. Every case runs the real
//! binary in a temporary repository with a temporary HOME; no real session, ledger or network.
#![cfg(unix)]
mod common;

use common::home::TempHome;
use moai_cli::ledger;
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Output, Stdio};

const BIN: &str = crate::common::MOAI_BIN;
const ID: &str = "SPEC-20261117-demo";
/// The name-first branch of `ID`, its old form, and the repair-suffix variants of both.
const NEW: &str = "spec/demo-20261117";
const OLD: &str = "spec/20261117-demo";
const NEW_R2: &str = "spec/demo-20261117-r2";
const OLD_R2: &str = "spec/20261117-demo-r2";
/// Reads date first as `SPEC-20261117-demo-20261118` and name first as
/// `SPEC-20261118-20261117-demo`: two SPEC ids.
const TWO_WAYS: &str = "spec/20261117-demo-20261118";
const TWO_IDS: &str = "reads as two SPEC ids";

// ---- hook (tool-preflight) ----

struct Run {
    code: i32,
    stdout: String,
    stderr: String,
}

fn preflight(cwd: &Path, input: &Value) -> Run {
    let mut cmd = Command::new(BIN);
    cmd.args([
        "hook-dispatch",
        "--host",
        "claude",
        "--event",
        "tool-preflight",
    ])
    .current_dir(cwd)
    .env("HOME", cwd.join("home"))
    .env_remove("MOAI_FORGE_GH_BIN")
    .env("GIT_CONFIG_GLOBAL", "/dev/null")
    .env("GIT_CONFIG_SYSTEM", "/dev/null")
    .stdin(Stdio::piped())
    .stdout(Stdio::piped())
    .stderr(Stdio::piped());
    let mut child = cmd.spawn().unwrap();
    common::feed_stdin(&mut child, input.to_string().as_bytes());
    let out = child.wait_with_output().unwrap();
    Run {
        code: out.status.code().unwrap(),
        stdout: String::from_utf8_lossy(&out.stdout).into_owned(),
        stderr: String::from_utf8_lossy(&out.stderr).into_owned(),
    }
}

fn git(cwd: &Path, args: &[&str]) {
    let out = common::git::command()
        .args(args)
        .current_dir(cwd)
        .env("HOME", cwd.join("home"))
        .output()
        .unwrap();
    assert!(
        out.status.success(),
        "git {args:?}: {}",
        String::from_utf8_lossy(&out.stderr)
    );
}

/// A temporary MoAI repository (a committed `.moai/` folder) on `branch` with one commit.
fn repo(branch: &str) -> (tempfile::TempDir, PathBuf) {
    let dir = tempfile::TempDir::new().unwrap();
    let root = fs::canonicalize(dir.path()).unwrap();
    fs::create_dir_all(root.join("home")).unwrap();
    git(&root, &["init", "-q", "-b", branch]);
    git(&root, &["config", "user.name", "fixture"]);
    git(&root, &["config", "user.email", "fixture@example.invalid"]);
    git(&root, &["config", "commit.gpgsign", "false"]);
    fs::create_dir_all(root.join(".moai")).unwrap();
    fs::write(root.join(".moai/.keep"), "").unwrap();
    fs::write(root.join("a.txt"), "a\n").unwrap();
    fs::write(root.join(".gitignore"), "home/\n").unwrap();
    git(&root, &["add", "-A"]);
    git(&root, &["commit", "-q", "-m", "init"]);
    (dir, root)
}

fn spec_text(id: &str) -> String {
    format!("---\nid: {id}\ntitle: \"demo\"\nstatus: draft\nrevision: 1\ncreated: 2026-11-17\n---\n\n## Intent\n\nx\n")
}

/// The SPEC file of `id` exists in the work tree (not committed).
fn write_spec(root: &Path, id: &str) {
    let folder = root.join(".moai/specs").join(id);
    fs::create_dir_all(&folder).unwrap();
    fs::write(folder.join("spec.md"), spec_text(id)).unwrap();
}

fn write(path: &str, cwd: &Path) -> Value {
    json!({"hook_event_name": "PreToolUse", "tool_name": "Write",
           "tool_input": {"file_path": path, "content": "fn main() {}\n"}, "cwd": cwd})
}

fn deny_reason(r: &Run) -> Option<String> {
    assert_eq!(r.code, 0, "exit code: {}", r.stderr);
    if r.stdout.is_empty() {
        return None;
    }
    let v: Value = serde_json::from_str(r.stdout.trim()).expect("stdout is one JSON object");
    assert_eq!(v["hookSpecificOutput"]["permissionDecision"], "deny", "{v}");
    Some(
        v["hookSpecificOutput"]["permissionDecisionReason"]
            .as_str()
            .unwrap()
            .to_string(),
    )
}

/// AC-2 (REQ-3, REQ-8): the SPEC-first rule reads both forms, with and without `-r<N>`, as the
/// same SPEC: the code write waits while `SPEC-20261117-demo` is missing and goes ahead once it
/// exists.
#[test]
fn name_first_hook_resolves_both_forms_and_repair_suffixes_to_one_spec() {
    for branch in [NEW, OLD, NEW_R2, OLD_R2] {
        let (_d, root) = repo(branch);
        let waiting = deny_reason(&preflight(&root, &write("src/main.rs", &root)));
        let reason = waiting.unwrap_or_else(|| panic!("{branch}: a code write must wait"));
        assert!(
            reason.starts_with("MoAI tool-preflight:"),
            "{branch}: {reason}"
        );
        // The reason is a fixed string: no branch from the input is echoed, and it names the
        // new form.
        assert!(!reason.contains("demo"), "{branch}: {reason}");
        assert!(
            reason.contains("spec/<slug>-<yyyymmdd>"),
            "{branch}: {reason}"
        );
        write_spec(&root, ID);
        assert_eq!(
            deny_reason(&preflight(&root, &write("src/main.rs", &root))),
            None,
            "{branch}: the SPEC exists now"
        );
    }
}

/// AC-2 (REQ-4): a name that is not a SPEC work branch in either form is left alone.
#[test]
fn name_first_hook_leaves_other_names_alone() {
    for branch in [
        "main",
        "spec/Demo-20261117",
        "spec/demo-2026111",
        "spec/demo-",
        "spec/-20261117",
        "spec/demo",
        "spec/demo-20261117/sub",
        "spec/demo_20261117",
        "issue/demo-20261117",
    ] {
        let (_d, root) = repo(branch);
        assert_eq!(
            deny_reason(&preflight(&root, &write("src/main.rs", &root))),
            None,
            "{branch} is not a SPEC work branch"
        );
    }
}

/// AC-5 (REQ-5): a name that reads as two SPEC ids is refused with the reason; a SPEC folder of
/// exactly one of them is evidence that settles it; folders of both do not.
#[test]
fn name_first_ambiguous_hook_refuses_with_the_reason() {
    let (_d, root) = repo(TWO_WAYS);
    let reason = deny_reason(&preflight(&root, &write("src/main.rs", &root)))
        .expect("an ambiguous name refuses");
    assert!(reason.starts_with("MoAI tool-preflight:"), "{reason}");
    assert!(reason.contains(TWO_IDS), "{reason}");
    // A write of the SPEC itself is never refused by this rule.
    assert_eq!(
        deny_reason(&preflight(
            &root,
            &write(".moai/specs/SPEC-20261117-demo-20261118/spec.md", &root)
        )),
        None
    );
    // One of the two readings has a SPEC folder: that is the one.
    write_spec(&root, "SPEC-20261117-demo-20261118");
    assert_eq!(
        deny_reason(&preflight(&root, &write("src/main.rs", &root))),
        None
    );
    // Both readings have one: refused again.
    write_spec(&root, "SPEC-20261118-20261117-demo");
    let reason = deny_reason(&preflight(&root, &write("src/main.rs", &root)))
        .expect("two SPEC folders refuse");
    assert!(reason.contains(TWO_IDS), "{reason}");
}

// ---- tdd ----

const CHECK: &str = "sh check.sh";
const CHECK_SH: &str =
    "if grep -q fixed src/lib.txt; then echo 'test result: ok. 1 passed; 0 failed'; \
                        else echo 'test result: FAILED. 0 passed; 1 failed'; exit 101; fi\n";

/// A repository on `branch` with a failing check, a SPEC folder for `ID` and a RED commit.
fn tdd_repo(branch: &str) -> (tempfile::TempDir, PathBuf) {
    let (d, root) = repo(branch);
    fs::write(root.join("check.sh"), CHECK_SH).unwrap();
    fs::create_dir_all(root.join("src")).unwrap();
    fs::write(root.join("src/lib.txt"), "broken\n").unwrap();
    write_spec(&root, ID);
    git(&root, &["add", "-A"]);
    git(&root, &["commit", "-q", "-m", "check"]);
    fs::create_dir_all(root.join("tests")).unwrap();
    fs::write(root.join("tests/login_test.rs"), "assert fixed\n").unwrap();
    git(&root, &["add", "-A"]);
    git(
        &root,
        &[
            "commit",
            "-q",
            "-m",
            "test: login fails first\n\nMoai-Phase: red",
        ],
    );
    (d, root)
}

fn tdd_red(root: &Path) -> (i32, Value) {
    let out = Command::new(BIN)
        .args(["tdd", "red", "--check", CHECK])
        .current_dir(root)
        .env("HOME", root.join("home"))
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .stdin(Stdio::null())
        .output()
        .unwrap();
    let v: Value = serde_json::from_slice(out.stdout.split(|b| *b == b'\n').next().unwrap())
        .unwrap_or(json!(null));
    (out.status.code().unwrap(), v)
}

/// AC-3 (REQ-3): `moai tdd red` finds the SPEC of the branch in both forms and with a repair
/// suffix.
#[test]
fn name_first_tdd_resolves_both_forms_and_repair_suffixes_to_one_spec() {
    for branch in [NEW, OLD, NEW_R2, OLD_R2] {
        let (_d, root) = tdd_repo(branch);
        let (code, v) = tdd_red(&root);
        assert_eq!((code, &v["ok"]), (0, &json!(true)), "{branch}: {v}");
        assert_eq!(v["result"]["record"]["spec"], json!(ID), "{branch}: {v}");
    }
}

/// AC-5 (REQ-5): `moai tdd red` refuses a name that reads as two SPEC ids, with the reason.
#[test]
fn name_first_ambiguous_tdd_refuses_with_the_reason() {
    let (_d, root) = tdd_repo(TWO_WAYS);
    let (code, v) = tdd_red(&root);
    assert!(code != 0 && v["ok"] == json!(false), "{code} {v}");
    let message = v["error"]["message"].as_str().unwrap_or_default();
    assert!(message.contains(TWO_IDS), "{message}");
    assert!(
        message.contains("SPEC-20261117-demo-20261118")
            && message.contains("SPEC-20261118-20261117-demo"),
        "{message}"
    );
}

// ---- trailer-check ----

fn trailer_check(dir: &Path, home: &Path) -> Value {
    let out = Command::new(BIN)
        .args(["trailer-check", "HEAD~1..HEAD"])
        .current_dir(dir)
        .env("HOME", home)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .env("LC_ALL", "C")
        .output()
        .unwrap();
    serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("{}", String::from_utf8_lossy(&out.stderr)))
}

/// AC-4 (REQ-3): a SPEC work branch in either form, with a repair suffix, needs the
/// `Moai-Spec` trailer; another branch does not. (Already true on the `spec/` prefix; kept as a
/// pin so the new reader cannot lose it.)
#[test]
fn name_first_trailer_check_knows_both_forms_as_spec_branches() {
    for (branch, spec_branch) in [
        (NEW, true),
        (OLD, true),
        (NEW_R2, true),
        (OLD_R2, true),
        ("demo-20261117", false),
        ("issue/demo-20261117", false),
    ] {
        let (_d, root) = repo(branch);
        fs::write(root.join("a.txt"), "b\n").unwrap();
        git(&root, &["add", "-A"]);
        git(
            &root,
            &["commit", "-q", "-m", "change without a SPEC trailer"],
        );
        let v = trailer_check(&root, &root.join("home"));
        assert_eq!(
            v["result"]["spec_branch"],
            json!(spec_branch),
            "{branch}: {v}"
        );
        assert_eq!(v["result"]["ok"], json!(!spec_branch), "{branch}: {v}");
    }
}

// ---- spec take ----

fn envelope(out: &Output) -> Value {
    serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
        panic!(
            "{} / {}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    })
}

fn ok(out: &Output) -> Value {
    let v = envelope(out);
    assert_eq!(out.status.code(), Some(0), "{v}");
    v["result"].clone()
}

fn refused(out: &Output) -> String {
    let v = envelope(out);
    assert_eq!(out.status.code(), Some(1), "{v}");
    v["error"]["message"]
        .as_str()
        .unwrap_or_default()
        .to_string()
}

/// A HOME, a registered ledger and a repository `demo` on `main` with one commit.
struct World {
    home: TempHome,
    project: PathBuf,
}

impl World {
    fn new() -> Self {
        let home = TempHome::new();
        let project = home.project("demo");
        ledger::register(home.path(), "demo").unwrap();
        let w = World { home, project };
        w.git(&["init", "-q", "-b", "main"]);
        fs::write(w.project.join("a.txt"), "a\n").unwrap();
        w.git(&["add", "-A"]);
        common::git::git_ok(
            &w.project,
            &[
                "-c",
                "user.name=fixture",
                "-c",
                "user.email=fixture@example.invalid",
                "-c",
                "commit.gpgsign=false",
                "commit",
                "-q",
                "-m",
                "init",
            ],
        );
        w
    }

    fn p(&self) -> &str {
        self.project.to_str().unwrap()
    }

    fn git(&self, args: &[&str]) -> String {
        common::git::git(&self.project, args)
    }

    fn run(&self, args: &[&str]) -> Output {
        self.home.command(args).output().expect("run moai")
    }

    fn take(&self, id: &str) -> Output {
        self.run(&["spec", "take", id, "--project", self.p()])
    }

    /// A draft SPEC `SPEC-20261117-<slug>` that passes `spec check` (placeholders filled).
    fn draft_spec(&self, slug: &str) -> String {
        let r = ok(&self.run(&[
            "spec",
            "new",
            slug,
            "--title",
            "Login fix",
            "--minimal",
            "--date",
            "2026-11-17",
            "--project",
            self.p(),
        ]));
        let id = r["id"].as_str().unwrap().to_string();
        let path = self.project.join(".moai/specs").join(&id).join("spec.md");
        fs::write(&path, filled(&fs::read_to_string(&path).unwrap())).unwrap();
        id
    }

    fn branch(&self) -> String {
        self.git(&["rev-parse", "--abbrev-ref", "HEAD"])
    }

    /// Branches, HEAD and the status: what a refusal must leave as it was.
    fn state(&self) -> (String, String, String) {
        (
            self.git(&[
                "branch",
                "--list",
                "--format=%(refname:short) %(objectname)",
            ]),
            self.git(&["rev-parse", "HEAD"]),
            self.git(&["status", "--porcelain"]),
        )
    }
}

/// A template's text with every `<…>` placeholder filled with the word `filled`.
fn filled(text: &str) -> String {
    let mut out = String::new();
    let mut rest = text;
    while let Some(open) = rest.find('<') {
        let Some(close) = rest[open..].find('>') else {
            break;
        };
        out.push_str(&rest[..open]);
        out.push_str("filled");
        rest = &rest[open + close + 1..];
    }
    out.push_str(rest);
    out
}

/// AC-1 (REQ-1): `spec take` creates `spec/<slug>-<yyyymmdd>`, records the same SPEC id, and
/// prints the next commands with that branch.
#[test]
fn name_first_take_creates_the_name_first_branch() {
    let w = World::new();
    let id = w.draft_spec("login-fix");
    assert_eq!(id, "SPEC-20261117-login-fix");
    let r = ok(&w.take(&id));
    assert_eq!(r["spec"], json!(id), "{r}");
    assert_eq!(r["branch"], json!("spec/login-fix-20261117"), "{r}");
    assert_eq!(r["created"], json!(true), "{r}");
    assert_eq!(w.branch(), "spec/login-fix-20261117");
    let next: Vec<&str> = r["next"]
        .as_array()
        .unwrap()
        .iter()
        .map(|v| v.as_str().unwrap())
        .collect();
    assert!(
        next.contains(&"git push -u origin spec/login-fix-20261117"),
        "{next:?}"
    );
    assert!(
        next.iter()
            .any(|c| c.contains("--head spec/login-fix-20261117")),
        "{next:?}"
    );
    assert!(
        w.git(&["branch", "--list", "spec/20261117-login-fix"])
            .is_empty(),
        "no old-form branch is made"
    );
    // Again on the branch: nothing to create or switch.
    let again = ok(&w.take(&id));
    assert_eq!(
        (again["created"].clone(), again["switched"].clone()),
        (json!(false), json!(false))
    );
}

/// AC-1 (REQ-2): a SPEC whose branch already exists in the old form is switched to, not given a
/// second branch.
#[test]
fn name_first_take_switches_to_an_existing_old_form_branch() {
    let w = World::new();
    let id = w.draft_spec("login-fix");
    w.git(&["branch", "spec/20261117-login-fix"]);
    let r = ok(&w.take(&id));
    assert_eq!(r["branch"], json!("spec/20261117-login-fix"), "{r}");
    assert_eq!(r["created"], json!(false), "{r}");
    assert_eq!(r["switched"], json!(true), "{r}");
    assert_eq!(w.branch(), "spec/20261117-login-fix");
    assert!(
        w.git(&["branch", "--list", "spec/login-fix-20261117"])
            .is_empty(),
        "no name-first branch is made beside it"
    );
}

/// AC-5 (REQ-6): a SPEC with a branch in both forms is refused, naming both, and nothing changes.
#[test]
fn name_first_ambiguous_take_refuses_two_branches_of_one_spec() {
    let w = World::new();
    let id = w.draft_spec("login-fix");
    w.git(&["branch", "spec/20261117-login-fix"]);
    w.git(&["branch", "spec/login-fix-20261117"]);
    let before = w.state();
    let message = refused(&w.take(&id));
    assert!(
        message.contains("spec/20261117-login-fix") && message.contains("spec/login-fix-20261117"),
        "{message}"
    );
    assert_eq!(w.state(), before, "a refusal changes nothing");
}

/// AC-5 (REQ-6): a SPEC whose new branch would read as two SPEC ids (the slug starts with a
/// date) is refused before any branch is made.
#[test]
fn name_first_ambiguous_take_refuses_a_branch_that_reads_two_ways() {
    let w = World::new();
    let id = w.draft_spec("20260101-foo");
    assert_eq!(id, "SPEC-20261117-20260101-foo");
    let before = w.state();
    let message = refused(&w.take(&id));
    assert!(message.contains(TWO_IDS), "{message}");
    assert!(
        message.contains("SPEC-20261117-20260101-foo")
            && message.contains("SPEC-20260101-foo-20261117"),
        "{message}"
    );
    assert_eq!(w.state(), before, "a refusal changes nothing");
}

// ---- skills ----

/// AC-6 (REQ-7): the skill references that define the work flow show the name-first branch.
#[test]
fn name_first_skills_show_the_new_form() {
    let root = Path::new(env!("CARGO_MANIFEST_DIR"));
    for file in [
        "skills/moai/references/internal/sdd.md",
        "skills/moai/references/internal/interview.md",
    ] {
        let text = fs::read_to_string(root.join(file)).unwrap();
        assert!(text.contains("`spec/<slug>-<yyyymmdd>`"), "{file}");
        assert!(
            !text.contains("branch `spec/<yyyymmdd>-<slug>`"),
            "{file} still defines the branch date first"
        );
    }
}
