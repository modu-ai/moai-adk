//! SPEC-20261126-project-next-flow-setup: picking NeXT-Flow in the project setup makes the flow
//! run. The setup step takes the integration branch, the card ledger and the protection hook in
//! that order, asks two plain follow-up questions and ends the report with the line that starts
//! the leader and the lanes. The tests read the PROJECT workflow and the question bank as text and
//! run the commands that text names with the built binary and a real git, in temporary
//! repositories and a temporary HOME; nothing here touches a real repository, hook or ledger.

#![cfg(unix)]

mod common;

use common::home::TempHome;
use common::CleanEnv;
use moai_cli::card::guard;
use moai_cli::ledger;
use moai_cli::profile::LAND_VERDICT_MODES;
use serde_json::{json, Value};
use std::collections::BTreeSet;
use std::fs;
use std::os::unix::fs::PermissionsExt;
use std::path::PathBuf;
use std::process::{Command, Output, Stdio};

const PROJECT: &str = "skills/moai/workflows/project.md";
const BANK: &str = "skills/moai/references/internal/interview.md";
/// The integration branch of the fixtures: a new dedicated branch beside `main`.
const BRANCH: &str = "integration";
const PROJECT_ID: &str = "demo";
const STEP_9: &str = "9. **Set up NeXT-Flow.**";
/// The rule that lists the new dedicated branch first, as both texts carry it.
const DEDICATED_RULE: &str = "while the profile names no integration branch and the current branch is the default branch, a new dedicated integration branch comes first";

fn read(rel: &str) -> String {
    fs::read_to_string(format!("{}/{rel}", env!("CARGO_MANIFEST_DIR")))
        .unwrap_or_else(|e| panic!("{rel}: {e}"))
}

fn position(text: &str, needle: &str) -> usize {
    text.find(needle)
        .unwrap_or_else(|| panic!("{needle:?} is missing from: {text}"))
}

/// The first line of `text` that starts with `head`.
fn line_starting<'a>(text: &'a str, head: &str) -> &'a str {
    text.lines()
        .find(|l| l.starts_with(head))
        .unwrap_or_else(|| panic!("no line starting {head:?}"))
}

/// The `## Procedure` section of the PROJECT workflow, up to the next `## ` heading.
fn procedure() -> String {
    let text = read(PROJECT);
    let rest = &text[position(&text, "\n## Procedure\n") + 1..];
    let end = rest[3..].find("\n## ").map_or(rest.len(), |at| at + 3);
    rest[..end].to_string()
}

/// Step 9 of the procedure: from its first line up to the blank line that ends the list item.
fn step_9() -> String {
    let procedure = procedure();
    let rest = &procedure[position(&procedure, STEP_9)..];
    rest[..position(rest, "\n\n")].to_string()
}

/// A bullet of step 9 (three spaces of indent) with its sub-bullets.
fn part(step: &str, head: &str) -> String {
    let head = format!("   - {head}");
    let mut out = String::new();
    let mut inside = false;
    for line in step.lines() {
        if line.starts_with(&head) {
            inside = true;
        } else if inside && !line.starts_with("     ") {
            break;
        }
        if inside {
            out.push_str(line);
            out.push('\n');
        }
    }
    assert!(!out.is_empty(), "step 9 has no bullet {head:?}: {step}");
    out
}

fn git_executable() -> PathBuf {
    let path = std::env::var_os("PATH").expect("PATH");
    std::env::split_paths(&path)
        .map(|dir| dir.join("git"))
        .find(|p| p.is_file())
        .expect("a git on PATH")
}

/// A temporary repository with one commit on `main`, a temporary HOME and, outside both, a folder
/// that holds the `moai` of this build for the hook to find on PATH.
struct Fx {
    home: TempHome,
    dir: tempfile::TempDir,
    repo: PathBuf,
    /// The one commit of `main`.
    tip: String,
}

impl Fx {
    fn new() -> Self {
        let home = TempHome::new();
        let dir = tempfile::Builder::new()
            .prefix("moai-test-flow-setup-")
            .tempdir()
            .unwrap();
        let bin = dir.path().join("bin");
        fs::create_dir_all(&bin).unwrap();
        std::os::unix::fs::symlink(common::MOAI_BIN, bin.join("moai")).unwrap();
        let repo = dir.path().join("r");
        fs::create_dir_all(&repo).unwrap();
        // The resolved path, as `work-flow-plan` compares worktree paths with it.
        let repo = repo.canonicalize().unwrap();
        let mut fx = Fx {
            home,
            dir,
            repo,
            tip: String::new(),
        };
        fx.git(&["init", "-q", "-b", "main"]);
        fs::write(fx.repo.join("a.txt"), "a\n").unwrap();
        fx.git(&["add", "a.txt"]);
        fx.git(&["commit", "-q", "-m", "init"]);
        fx.tip = fx.git(&["rev-parse", "HEAD"]);
        fx
    }

    fn path_var(&self) -> String {
        std::env::join_paths([
            self.dir.path().join("bin"),
            git_executable().parent().unwrap().to_path_buf(),
            PathBuf::from("/usr/bin"),
            PathBuf::from("/bin"),
        ])
        .unwrap()
        .to_string_lossy()
        .into_owned()
    }

    fn try_git(&self, args: &[&str]) -> Output {
        Command::new(git_executable())
            .arg("-C")
            .arg(&self.repo)
            .args([
                "-c",
                "user.name=fixture",
                "-c",
                "user.email=fixture@example.invalid",
                "-c",
                "commit.gpgsign=false",
                "-c",
                "gc.auto=0",
                "-c",
                "maintenance.auto=false",
            ])
            .args(args)
            .clean_env()
            .env("HOME", self.home.path())
            .env("PATH", self.path_var())
            .env("LC_ALL", "C")
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_SYSTEM", "/dev/null")
            .stdin(Stdio::null())
            .output()
            .unwrap()
    }

    fn git(&self, args: &[&str]) -> String {
        let out = self.try_git(args);
        assert!(
            out.status.success(),
            "git {args:?}: {}",
            String::from_utf8_lossy(&out.stderr)
        );
        String::from_utf8_lossy(&out.stdout).trim().to_string()
    }

    fn branches(&self) -> String {
        self.git(&[
            "for-each-ref",
            "refs/heads/",
            "--format=%(refname:short) %(objectname)",
        ])
    }

    fn status(&self) -> String {
        self.git(&["status", "--porcelain=v1", "-z", "--untracked-files=all"])
    }

    /// `moai <args>` in the repository with the HOME of this fixture and a clean environment;
    /// `stdin` is the request, when the command reads one. Answers the `result` of a success.
    fn moai(&self, args: &[&str], stdin: Option<&Value>) -> Value {
        let mut cmd = self.home.command(args);
        cmd.current_dir(&self.repo)
            .env("PATH", self.path_var())
            .env("MOAI_GIT_BIN", git_executable())
            .env("LC_ALL", "C")
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_SYSTEM", "/dev/null")
            .stdin(if stdin.is_some() {
                Stdio::piped()
            } else {
                Stdio::null()
            })
            .stdout(Stdio::piped())
            .stderr(Stdio::piped());
        let mut child = cmd.spawn().expect("run moai");
        if let Some(request) = stdin {
            common::feed_stdin(&mut child, request.to_string().as_bytes());
        }
        let out = child.wait_with_output().expect("wait for moai");
        let envelope: Value = serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
            panic!(
                "{args:?}: no envelope: {} / {}",
                String::from_utf8_lossy(&out.stdout),
                String::from_utf8_lossy(&out.stderr)
            )
        });
        assert_eq!(envelope["ok"], json!(true), "{args:?}: {envelope}");
        envelope["result"].clone()
    }

    fn repo_arg(&self) -> &str {
        self.repo.to_str().unwrap()
    }

    /// Step 4 of the route: `project-plan`, then `project-apply` of the plan as printed.
    fn apply_profile(&self, profile: &Value) {
        let plan = self.moai(&["project-plan", self.repo_arg()], Some(profile));
        let digest = plan["plan_digest"].as_str().expect("a digest").to_string();
        self.moai(
            &[
                "project-apply",
                self.repo_arg(),
                "--plan-digest",
                &digest,
                "--permission-reference",
                "approved-by-test",
            ],
            Some(&plan),
        );
    }

    fn work_flow_plan(&self) -> Value {
        self.moai(&["work-flow-plan", "--project", self.repo_arg()], None)
    }

    /// The hook preview and its apply, as the step names them; answers the apply's result.
    fn install_hook(&self) -> Value {
        let plan = self.moai(&["git-hooks-plan", self.repo_arg()], None);
        let token = plan["confirm_token"].as_str().expect("a token").to_string();
        assert!(
            plan["install"]
                .as_array()
                .unwrap()
                .contains(&json!("reference-transaction")),
            "{plan}"
        );
        self.moai(
            &[
                "git-hooks-apply",
                self.repo_arg(),
                "--confirm-git-hooks",
                &token,
            ],
            None,
        )
    }

    fn hook(&self) -> PathBuf {
        self.repo.join(".git/hooks/reference-transaction")
    }
}

fn profile(branch: Option<&str>) -> Value {
    let mut work_flow = json!({"flow": "next-flow"});
    if let Some(branch) = branch {
        work_flow["integration_branch"] = json!(branch);
    }
    json!({
        "schema_version": 2, "project_id": PROJECT_ID, "purpose": "Fixture",
        "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude"],
        "work_flow": work_flow,
    })
}

/// A refused hand move: a non-zero exit with the guard's words, and the branches where they were.
fn refused(fx: &Fx, args: &[&str]) {
    let before = fx.branches();
    let out = fx.try_git(args);
    let stderr = String::from_utf8_lossy(&out.stderr);
    assert!(!out.status.success(), "git {args:?} was not refused");
    assert!(stderr.contains("moai ref guard"), "git {args:?}: {stderr}");
    assert_eq!(fx.branches(), before, "git {args:?} moved a branch");
}

/// AC-1 (REQ-1, REQ-4): the order step 9 names (integration branch, card ledger, protection
/// hook), run for real, leaves the branch record, an executable hook and the ledger file; the hand
/// moves are then refused, nothing of it enters the tracked project tree, and the way back the
/// apply names takes the hook out.
#[test]
fn setup_makes_record_hook_ledger_and_refuses_hand_moves() {
    let step = step_9();
    assert!(
        position(&step, "`moai work-flow-plan") < position(&step, "`moai ledger-register-apply")
            && position(&step, "`moai ledger-register-apply")
                < position(&step, "`moai git-hooks-apply"),
        "step 9 takes the branch, then the ledger, then the hook: {step}"
    );

    let fx = Fx::new();
    fx.apply_profile(&profile(Some(BRANCH)));
    let status = fx.status();

    let first = fx.work_flow_plan();
    assert_eq!(first["state"]["kind"], "missing-target", "{first}");
    fx.git(&["branch", BRANCH, &fx.tip]);
    let second = fx.work_flow_plan();
    assert_eq!(second["state"]["kind"], "ok", "{second}");

    let ledger_file = ledger::ledger_path(fx.home.path(), PROJECT_ID).unwrap();
    assert!(!ledger_file.exists());
    let preview = fx.moai(&["ledger-register-plan", fx.repo_arg()], None);
    assert_eq!(preview["action"], "register", "{preview}");
    let digest = preview["plan_digest"].as_str().unwrap().to_string();
    fx.moai(
        &[
            "ledger-register-apply",
            fx.repo_arg(),
            "--plan-digest",
            &digest,
            "--permission-reference",
            "approved-by-test",
        ],
        None,
    );
    assert!(ledger_file.is_file(), "{}", ledger_file.display());

    let applied = fx.install_hook();
    assert_eq!(applied["guard"]["record"], "written", "{applied}");
    assert_eq!(applied["guard"]["branch"], BRANCH, "{applied}");
    assert_eq!(
        fs::read_to_string(fx.repo.join(".git").join(guard::RECORD_FILE)).unwrap(),
        format!("refs/heads/{BRANCH}\n")
    );
    let mode = fs::metadata(fx.hook()).unwrap().permissions().mode();
    assert_eq!(mode & 0o111, 0o111, "the hook is executable: {mode:o}");
    assert_eq!(
        fx.status(),
        status,
        "the ledger and the hook write nothing the closing commit would miss"
    );

    let other = fx.git(&[
        "commit-tree",
        &format!("{}^{{tree}}", fx.tip),
        "-p",
        &fx.tip,
        "-m",
        "a head no landing made",
    ]);
    refused(
        &fx,
        &["update-ref", &format!("refs/heads/{BRANCH}"), &other],
    );
    refused(&fx, &["branch", "-f", BRANCH, &other]);
    refused(&fx, &["branch", "-D", BRANCH]);

    let stamp = applied["stamp"].as_str().unwrap().to_string();
    assert_eq!(
        applied["revert"],
        json!(format!("moai git-hooks-revert {stamp}")),
        "{applied}"
    );
    fx.moai(&["git-hooks-revert", &stamp, fx.repo_arg()], None);
    assert!(!fx.hook().exists(), "the way back takes the hook out");
    fx.git(&["branch", "-f", BRANCH, &other]);
}

/// AC-1, the reason for the order: with the hook installed first, the integration branch cannot
/// be created by hand.
#[test]
fn hook_first_then_branch_create_is_refused() {
    let fx = Fx::new();
    fx.apply_profile(&profile(Some(BRANCH)));
    let applied = fx.install_hook();
    assert_eq!(applied["guard"]["record"], "written", "{applied}");
    refused(&fx, &["branch", BRANCH, &fx.tip]);
    assert!(
        !fx.branches()
            .lines()
            .any(|l| l.starts_with(&format!("{BRANCH} "))),
        "the branch was not made"
    );

    // Step 9 gives that as the reason for its order.
    let step = step_9();
    let head = line_starting(&step, STEP_9);
    assert!(
        head.contains("every hand move of the integration branch is refused, creating it included"),
        "{head}"
    );
}

/// REQ-4, the case the hook cannot cover: with `core.hooksPath` set, git runs no hook from
/// `.git/hooks`, so a hand move goes through although the apply wrote the hook and the record.
/// The apply answers `not_active` with the folder git reads, and step 9 reads the same setting
/// and never calls the branch protected then.
#[test]
fn hooks_path_leaves_the_hook_unused_and_step_9_says_so() {
    let fx = Fx::new();
    fx.apply_profile(&profile(Some(BRANCH)));
    fx.git(&["branch", BRANCH, &fx.tip]);
    fx.git(&["config", "core.hooksPath", ".husky"]);
    let applied = fx.install_hook();
    assert_eq!(applied["guard"]["record"], "not_active", "{applied}");
    assert!(
        applied["guard"]["hooks_path"]
            .as_str()
            .is_some_and(|p| p.ends_with(".husky")),
        "the answer names the folder git reads: {applied}"
    );
    assert!(applied["guard"]["why"].is_string(), "{applied}");
    assert!(fx.hook().exists(), "{applied}");
    let other = fx.git(&[
        "commit-tree",
        &format!("{}^{{tree}}", fx.tip),
        "-p",
        &fx.tip,
        "-m",
        "a head no landing made",
    ]);
    fx.git(&["branch", "-f", BRANCH, &other]);
    assert_eq!(
        fx.git(&["rev-parse", BRANCH]),
        other,
        "the hand move went through"
    );

    let step = step_9();
    let hook_part = part(&step, "Protection hook.");
    for text in [
        "`git config --get core.hooksPath`",
        "other than the repository's `.git/hooks`",
        "`not applied: git runs hooks from <that folder>`",
        "never call the branch protected",
    ] {
        assert!(hook_part.contains(text), "{text:?}: {hook_part}");
    }
    assert!(
        position(&hook_part, "`git config --get core.hooksPath`")
            < position(&hook_part, "`moai git-hooks-apply"),
        "the setting is read before the apply: {hook_part}"
    );
}

/// The cells of the follow-up table of the question bank: name, question, options, goes to.
fn follow_up_rows(section: &str) -> Vec<Vec<String>> {
    section
        .lines()
        .filter(|l| {
            l.starts_with("| ") && !l.starts_with("| Follow-up ") && !l.starts_with("| --- ")
        })
        .map(|l| {
            l.trim_matches('|')
                .split(" | ")
                .map(|c| c.trim().to_string())
                .collect()
        })
        .collect()
}

/// The `## NeXT-Flow follow-ups` section of the question bank.
fn follow_up_section() -> String {
    let bank = read(BANK);
    let rest = &bank[position(&bank, "\n## NeXT-Flow follow-ups\n") + 1..];
    let end = rest[3..].find("\n## ").map_or(rest.len(), |at| at + 3);
    rest[..end].to_string()
}

/// AC-2 (REQ-5, REQ-6, REQ-7, REQ-9, REQ-11): the two follow-up questions speak plain words, the
/// recommended option is first, nothing is preselected, the asked table still holds the purpose
/// alone, and both answers are read back from the profile.
#[test]
fn follow_up_rows_use_plain_words_and_answers_reach_the_profile() {
    let section = follow_up_section();
    let rows = follow_up_rows(&section);
    let names: Vec<&str> = rows.iter().map(|r| r[0].as_str()).collect();
    assert_eq!(names, ["Remote copy", "Landing check"], "{section}");
    let mut banned = vec![
        "spec",
        "worktree",
        "audit",
        "leader",
        "card",
        "remote_copy",
        "land_verdict",
        "work_flow",
        "integration_branch",
        "keep-open",
        "keep_open",
        "`",
    ];
    banned.extend_from_slice(LAND_VERDICT_MODES);
    for row in &rows {
        assert_eq!(row.len(), 4, "{row:?}");
        assert!(row[1].contains('?'), "the question cell asks: {row:?}");
        for cell in [&row[1], &row[2]] {
            let lower = cell.to_lowercase();
            for word in &banned {
                assert!(
                    !lower.contains(word),
                    "the user would read `{word}`: {cell}"
                );
            }
        }
        let options = &row[2];
        assert_eq!(
            options.matches("(recommended)").count(),
            1,
            "one recommended option: {options}"
        );
        assert!(
            position(options, "(recommended)") < position(options, "; "),
            "the recommended option is first: {options}"
        );
    }
    let (remote, landing) = (&rows[0], &rows[1]);
    assert!(
        remote[2].starts_with("No, keep it local (recommended)"),
        "{remote:?}"
    );
    assert!(
        landing[2].starts_with("Turn it on (recommended)")
            && position(&landing[2], "Turn it on") < position(&landing[2], "Leave it off"),
        "{landing:?}"
    );
    // A closing landing needs the passed check in both settings, and the question says so.
    assert!(
        landing[1].contains("needs a passed check by another agent whichever you pick"),
        "{landing:?}"
    );
    for text in [
        "No option is preselected",
        "an unanswered or cancelled question records nothing",
        "one follow-up call of three questions or fewer",
    ] {
        assert!(
            section.contains(text),
            "{text:?} is missing from: {section}"
        );
    }

    // The asked table keeps the purpose as its only row.
    let bank = read(BANK);
    let asked = &bank[position(&bank, "\n## Asked\n")..position(&bank, "\n## Confirmed\n")];
    let ids: Vec<&str> = asked.lines().filter(|l| l.starts_with("| `")).collect();
    assert_eq!(ids.len(), 1, "{ids:?}");
    assert!(ids[0].starts_with("| `purpose` |"), "{ids:?}");

    let project = read(PROJECT);
    let follow_up = line_starting(&project, "   - NeXT-Flow follow-up");
    for text in [
        "one follow-up call of three questions or fewer",
        "`NeXT-Flow follow-ups`",
    ] {
        assert!(
            follow_up.contains(text),
            "{text:?} is missing from: {follow_up}"
        );
    }
    let step_4 = line_starting(&project, "4. **Plan and apply.**");
    let fields = &step_4[position(step_4, "`result.profile` (")..];
    let fields = &fields[..position(fields, ")")];
    assert!(fields.contains("remote_copy"), "{fields}");
    assert!(
        step_4.contains("an unanswered follow-up records nothing"),
        "{step_4}"
    );

    // The keys and words the bank says it records reach the profile through the plan and apply.
    for text in ["profile `remote_copy`", "`true` for yes", "`false` for no"] {
        assert!(remote[3].contains(text), "{text:?}: {remote:?}");
    }
    for text in [
        "profile `work_flow.land_verdict`",
        "`enforce` for on",
        "`observe` for off",
    ] {
        assert!(landing[3].contains(text), "{text:?}: {landing:?}");
    }
    for (copy, verdict) in [(false, "enforce"), (true, "observe")] {
        let fx = Fx::new();
        let mut answers = profile(Some(BRANCH));
        answers["remote_copy"] = json!(copy);
        answers["work_flow"]["land_verdict"] = json!(verdict);
        fx.apply_profile(&answers);
        let (read_back, _) = moai_cli::read_profile(&fx.repo).unwrap();
        assert_eq!(read_back.remote_copy, Some(copy));
        assert_eq!(
            read_back.work_flow.unwrap().land_verdict.as_deref(),
            Some(verdict)
        );
    }
    let fx = Fx::new();
    fx.apply_profile(&profile(Some(BRANCH)));
    let (unanswered, _) = moai_cli::read_profile(&fx.repo).unwrap();
    assert_eq!(unanswered.remote_copy, None);
    assert_eq!(unanswered.work_flow.unwrap().land_verdict, None);
}

/// Every state `work-flow-plan` can answer, read from the code that prints it.
fn state_kinds() -> BTreeSet<String> {
    let source = read("src/work_flow.rs");
    source
        .split("\"kind\": \"")
        .skip(1)
        .map(|rest| rest.split('"').next().unwrap().to_string())
        .collect()
}

/// AC-3 (REQ-2, REQ-3): the guidance covers every state `work-flow-plan` answers, relays its
/// message as printed, offers the creation only for a missing target after the pick, and creates
/// nothing for an occupied or unreadable one.
#[test]
fn setup_guidance_is_work_flow_plans_answer() {
    // The real answers, in temporary repositories; the command creates no branch in any of them.
    let mut answered = BTreeSet::new();
    let mut ask = |fx: &Fx| {
        let before = fx.branches();
        let v = fx.work_flow_plan();
        assert_eq!(fx.branches(), before, "work-flow-plan created a branch");
        answered.insert(v["state"]["kind"].as_str().unwrap().to_string());
        v
    };
    let missing = Fx::new();
    missing.apply_profile(&profile(Some(BRANCH)));
    let v = ask(&missing);
    assert_eq!(v["state"]["kind"], "missing-target", "{v}");
    assert!(v["state"]["message"].is_string(), "{v}");
    missing.git(&[
        "update-ref",
        &format!("refs/remotes/origin/{BRANCH}"),
        "HEAD",
    ]);
    let v = ask(&missing);
    assert!(
        v["state"]["message"]
            .as_str()
            .unwrap()
            .contains("exists on origin"),
        "{v}"
    );
    missing.git(&["branch", BRANCH, &missing.tip]);
    let v = ask(&missing);
    assert_eq!(v["state"]["kind"], "ok", "{v}");
    assert_eq!(v["profile"]["integration_branch"], BRANCH, "{v}");

    let occupied = Fx::new();
    occupied.apply_profile(&profile(Some(BRANCH)));
    let other = occupied.dir.path().join("other");
    occupied.git(&[
        "worktree",
        "add",
        "-q",
        "-b",
        BRANCH,
        other.to_str().unwrap(),
    ]);
    let v = ask(&occupied);
    assert_eq!(v["state"]["kind"], "occupied-worktree", "{v}");

    let unreadable = Fx::new();
    let file = moai_cli::state_root::path(&unreadable.repo, "project.yaml");
    fs::create_dir_all(file.parent().unwrap()).unwrap();
    fs::write(&file, "{not a profile").unwrap();
    let v = ask(&unreadable);
    assert_eq!(v["state"]["kind"], "profile-unreadable", "{v}");

    // `ok` with no recorded branch: the step has nothing to set up.
    let unnamed = Fx::new();
    unnamed.apply_profile(&profile(None));
    let v = ask(&unnamed);
    assert_eq!(v["state"]["kind"], "ok", "{v}");
    assert_eq!(v["profile"]["integration_branch"], Value::Null, "{v}");

    assert_eq!(answered, state_kinds(), "every state was produced here");

    let step = step_9();
    let branch_part = part(&step, "Integration branch.");
    for text in [
        "`moai work-flow-plan --project <project>`",
        "`result.state.kind`",
        "only when the answer is `ok` and `result.profile.integration_branch` names a branch",
        "`not applied: <reason>`",
    ] {
        assert!(branch_part.contains(text), "{text:?}: {branch_part}");
    }
    let case = |head: &str| line_starting(&branch_part, &format!("     - {head}")).to_string();
    for kind in &answered {
        assert!(
            branch_part
                .lines()
                .any(|l| l.starts_with("     - ") && l.contains(&format!("`{kind}`"))),
            "no case for `{kind}`: {branch_part}"
        );
    }
    let missing = case("`missing-target`");
    for text in [
        "`state.message` as printed",
        "ask one selection",
        "`git branch <branch> <tip>`",
        "only after the user picks it",
        "no option is preselected",
        "exists on origin",
        "run `work-flow-plan` again",
        "create the branch from the current tip, or leave it",
        "also create it from `origin/<branch>`",
        "will not hold the setup files",
    ] {
        assert!(missing.contains(text), "{text:?}: {missing}");
    }
    assert_eq!(
        step.matches("`git branch ").count(),
        missing.matches("`git branch ").count(),
        "a branch is created for a missing target only: {step}"
    );
    let stop = case("`occupied-worktree` or `profile-unreadable`");
    for text in ["create nothing", "`state.message` as printed"] {
        assert!(stop.contains(text), "{text:?}: {stop}");
    }
}

/// AC-4 (REQ-1, REQ-8, REQ-10, REQ-12): step 9 stands after step 8 and before the final report,
/// takes its parts in order with nothing preselected, the report ends with the line that starts
/// the leader and the lanes, nothing tells the session to push, and a new dedicated branch is the
/// first recommendation only for a first setup on the default branch.
#[test]
fn step_9_order_and_closing_line() {
    let project = read(PROJECT);
    let procedure = procedure();
    let step_8 = line_starting(&procedure, "8. **Offer the closing commit");
    assert!(
        position(&procedure, "run `moai project-apply") < position(&procedure, STEP_9)
            && position(&procedure, "8. **Offer the closing commit") < position(&procedure, STEP_9)
            && position(&procedure, STEP_9) < position(&procedure, "\nFinal report:"),
        "step 9 follows the profile apply and step 8, before the final report"
    );
    assert!(
        step_8.contains("just before step 9 and the final report")
            && !step_8.contains("just before the final report"),
        "step 8 agrees with step 9: {step_8}"
    );
    let step = step_9();
    let head = line_starting(&step, STEP_9);
    for text in [
        "`next-flow`",
        "skip it for every other flow",
        "after step 8 and before the final report",
        "primary working folder",
        "outside a git repository",
    ] {
        assert!(head.contains(text), "{text:?}: {head}");
    }
    let order = [
        "`moai work-flow-plan",
        "`moai ledger-register-plan",
        "`moai ledger-register-apply",
        "`moai git-hooks-plan",
        "`moai git-hooks-apply",
    ];
    for pair in order.windows(2) {
        assert!(
            position(&step, pair[0]) < position(&step, pair[1]),
            "{} comes before {}: {step}",
            pair[0],
            pair[1]
        );
    }
    let ledger_part = part(&step, "Card ledger.");
    for text in [
        "only after the user picks it",
        "no option is preselected",
        "--plan-digest <result.plan_digest> --permission-reference <ref>",
    ] {
        assert!(ledger_part.contains(text), "{text:?}: {ledger_part}");
    }
    let hook_part = part(&step, "Protection hook.");
    for text in [
        "only after the user picks it",
        "no option is preselected",
        "--confirm-git-hooks <result.confirm_token>",
        "`result.guard.record`",
        "`not_written`",
        "`not_active`",
        "`result.guard.why`",
        "`moai git-hooks-revert <result.stamp> <project>`",
        "only the landing command (`moai card land`) moves",
    ] {
        assert!(hook_part.contains(text), "{text:?}: {hook_part}");
    }
    let ledger_pick = line_starting(&project, "**The card-ledger registration pick.**");
    assert!(
        ledger_pick.contains("A NeXT-Flow project makes this pick in step 9"),
        "the ledger is offered once: {ledger_pick}"
    );

    // The follow-up bullet of step 3 and the bank: the new branch first, only for a first setup.
    let follow_up = line_starting(&project, "   - NeXT-Flow follow-up");
    assert!(follow_up.contains(DEDICATED_RULE), "{follow_up}");
    assert!(
        follow_up.contains("otherwise the profile value stays first"),
        "{follow_up}"
    );
    let bank = read(BANK);
    assert!(bank.contains(DEDICATED_RULE), "the bank says the same");

    // The report: the NeXT-Flow rows and, as its last line, how to start the team.
    let table = line_starting(&procedure, "Final report:");
    for row in [
        "the integration branch",
        "the protection hook",
        "the remote copy",
        "the landing check",
        "the card ledger",
    ] {
        assert!(table.contains(row), "the report lacks {row:?}: {table}");
    }
    let last = procedure
        .lines()
        .rev()
        .find(|l| !l.trim().is_empty())
        .unwrap();
    assert!(
        last.contains("last line of the final report")
            && last.contains("`moai cc -l`")
            && last.contains("`moai cc -a`"),
        "the report ends with the starting line: {last}"
    );
    let home = TempHome::new();
    let help = home.command(&["cc", "--help"]).output().expect("run moai");
    let help = String::from_utf8_lossy(&help.stdout);
    for flag in ["-l, --leader", "-a, --agent"] {
        assert!(help.contains(flag), "`moai cc` has no {flag}: {help}");
    }
    for retired in ["/moai -l ", "/moai -a "] {
        assert!(!project.contains(retired), "project.md says {retired:?}");
    }
    let from_step_8 = &procedure[position(&procedure, "8. **Offer the closing commit")..];
    assert!(
        !from_step_8.contains("git push"),
        "nothing from step 8 on tells the session to push: {from_step_8}"
    );
}
