//! #3009 S3b: the managed `.gitignore` block (`moai-local`) through the binary. It keeps the
//! state folder's team files in git and ignores its machine output, shows its lines in the
//! plan before anything is written, widens an older block without touching the user's own rules,
//! and never untracks a file. Every case runs in a temporary git repository with an empty
//! temporary HOME; git answers `check-ignore`, so the lists are tested as git reads them.

mod common;
use common::git::{git, git_ok, git_raw};
use common::CleanEnv;
use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::{
    fs,
    io::Write,
    path::{Path, PathBuf},
    process::{Command, Stdio},
};
use tempfile::TempDir;

const PERMISSION: &str = "fixture-approved";
const BEGIN: &str = "# moai:managed:begin id=moai-local";
const END: &str = "# moai:managed:end id=moai-local";

struct Fx {
    _tmp: TempDir,
    root: PathBuf,
    home: PathBuf,
}

impl Fx {
    /// A git repository with a project profile (`github_issues` as given) applied.
    fn new(github_issues: Option<bool>) -> Fx {
        let tmp = TempDir::new().unwrap();
        let base = fs::canonicalize(tmp.path()).unwrap();
        let (root, home) = (base.join("project"), base.join("home"));
        fs::create_dir_all(root.join(".moai-cli")).unwrap();
        fs::create_dir_all(root.join("skills/moai/workflows")).unwrap();
        fs::create_dir_all(&home).unwrap();
        let fx = Fx {
            _tmp: tmp,
            root,
            home,
        };
        git_ok(&fx.root, &["init", "-q", "-b", "main"]);
        let mut profile = json!({
            "schema_version": 2, "project_id": "demo", "purpose": "Build a small editor",
            "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude"],
            "interview_revision": "rc1-fixture-1",
        });
        if let Some(on) = github_issues {
            profile["github_issues"] = json!(on);
        }
        let plan = fx.call(&["project-plan"], &profile);
        let digest = plan["plan_digest"].as_str().unwrap().to_string();
        fx.call(
            &[
                "project-apply",
                "--plan-digest",
                &digest,
                "--permission-reference",
                PERMISSION,
            ],
            &plan,
        );
        fx
    }

    fn p(&self) -> &str {
        self.root.to_str().unwrap()
    }

    fn run(&self, args: &[&str], stdin: &Value) -> Value {
        // The project path is the first argument after the subcommand name.
        let mut full: Vec<&str> = vec![args[0], self.p()];
        full.extend(&args[1..]);
        let mut child = Command::new(common::MOAI_BIN)
            .args(&full)
            .clean_env()
            .env("HOME", &self.home)
            .env("PATH", "/usr/bin:/bin")
            .env("LC_ALL", "C")
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
            .unwrap();
        let mut pipe = child.stdin.take().unwrap();
        match pipe.write_all(stdin.to_string().as_bytes()) {
            Ok(()) => {}
            Err(e) if e.kind() == std::io::ErrorKind::BrokenPipe => {}
            Err(e) => panic!("stdin write failed: {e}"),
        }
        drop(pipe);
        let out = child.wait_with_output().unwrap();
        serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
            panic!(
                "{full:?}: {}{}",
                String::from_utf8_lossy(&out.stdout),
                String::from_utf8_lossy(&out.stderr)
            )
        })
    }

    fn call(&self, args: &[&str], stdin: &Value) -> Value {
        let v = self.run(args, stdin);
        assert_eq!(v["ok"], json!(true), "{args:?}: {v}");
        v["result"].clone()
    }

    fn request(local_file: &str) -> Value {
        json!({"protocol_version": 1, "language": "en",
               "workflow_reference": "skills/moai/workflows/", "local_file": local_file})
    }

    fn plan(&self, local_file: &str) -> Value {
        self.call(&["agents-plan"], &Fx::request(local_file))
    }

    fn apply(&self, plan: &Value) -> Value {
        self.run(
            &[
                "agents-apply",
                "--plan-digest",
                plan["plan_digest"].as_str().unwrap(),
                "--permission-reference",
                PERMISSION,
            ],
            plan,
        )
    }

    fn apply_ok(&self, local_file: &str) -> Value {
        let out = self.apply(&self.plan(local_file));
        assert_eq!(out["ok"], json!(true), "{out}");
        out
    }

    fn ignore(&self) -> String {
        fs::read_to_string(self.root.join(".gitignore")).unwrap()
    }

    fn ignored(&self, rel: &str) -> bool {
        // `--no-index`: the rules alone decide, not whether the file happens to be tracked.
        let out = git_raw(
            &self.root,
            &[
                "-c",
                "core.excludesFile=/dev/null",
                "check-ignore",
                "-q",
                "--no-index",
                rel,
            ],
        );
        out.success
    }

    fn manifest(&self) -> Value {
        let text = fs::read_to_string(self.root.join(".moai-cli/manifests/generation.json"))
            .expect("generation manifest");
        serde_json::from_str(&text).unwrap()
    }

    fn ignore_row(plan: &Value) -> &Value {
        plan["files"]
            .as_array()
            .unwrap()
            .iter()
            .find(|f| f["path"] == ".gitignore")
            .unwrap_or_else(|| panic!("no .gitignore row: {plan}"))
    }
}

fn sha(bytes: &[u8]) -> String {
    format!("sha256:{:x}", Sha256::digest(bytes))
}

/// What the block was before S3b: the personal file and the receipts line, nothing else.
fn old_block() -> String {
    format!("{BEGIN}\nAGENTS.local.md\n.moai-cli/receipts/\n{END}\n")
}

/// The project as an earlier tool left it: the old block between the user's own rules, the
/// manifest baseline of that block, no recorded kind.
fn write_old_state(fx: &Fx, before: &str, after: &str) {
    fx.apply_ok("ignored");
    fs::write(
        fx.root.join(".gitignore"),
        format!("{before}{}{after}", old_block()),
    )
    .unwrap();
    let path = fx.root.join(".moai-cli/manifests/generation.json");
    let mut manifest = fx.manifest();
    for file in manifest["owned_files"].as_array_mut().unwrap() {
        if file["path"] == ".gitignore" {
            file["baseline_sha256"] = json!(sha(old_block().as_bytes()));
            file.as_object_mut().unwrap().remove("kind");
        }
    }
    fs::write(path, serde_json::to_vec_pretty(&manifest).unwrap()).unwrap();
}

const TEAM_FILES: [&str; 11] = [
    ".moai-cli/project.yaml",
    ".moai-cli/moai-cli.json",
    ".moai-cli/moai.json",
    ".moai-cli/manifests/generation.json",
    ".moai-cli/workflows/run.md",
    ".moai-cli/config/sections/language.yaml",
    ".moai-cli/decisions/D-1.md",
    ".moai-cli/docs/design.md",
    ".moai-cli/specs/SPEC-1/spec.md",
    "src/main.rs",
    "AGENTS.md",
];

const MACHINE_FILES: [&str; 6] = [
    ".moai-cli/state/lane.json",
    ".moai-cli/index/atlas.db",
    ".moai-cli/receipts/r1.json",
    ".moai-cli/worktrees/w1/file",
    ".moai-cli/cache/blob",
    ".moai-cli/logs/run.log",
];

#[test]
fn git_keeps_the_team_files_and_ignores_the_machine_output() {
    let fx = Fx::new(None);
    fx.apply_ok("ignored");
    for rel in TEAM_FILES {
        assert!(!fx.ignored(rel), "{rel} is ignored:\n{}", fx.ignore());
    }
    for rel in MACHINE_FILES.iter().chain(&["AGENTS.local.md"]) {
        assert!(fx.ignored(rel), "{rel} is not ignored:\n{}", fx.ignore());
    }
}

#[test]
fn the_personal_file_is_listed_only_when_it_is_ignored() {
    for (choice, listed, kind) in [
        ("ignored", true, "agents:ignore-with-local"),
        ("tracked", false, "agents:ignore-state"),
        ("none", false, "agents:ignore-state"),
    ] {
        let fx = Fx::new(None);
        fx.apply_ok(choice);
        let text = fx.ignore();
        assert_eq!(
            text.lines().any(|l| l == "AGENTS.local.md"),
            listed,
            "{choice}:\n{text}"
        );
        assert_eq!(fx.ignored("AGENTS.local.md"), listed, "{choice}");
        // The state-folder rules do not depend on the choice.
        for rel in MACHINE_FILES {
            assert!(fx.ignored(rel), "{choice}: {rel}");
        }
        let manifest = fx.manifest();
        let row = manifest["owned_files"]
            .as_array()
            .unwrap()
            .iter()
            .find(|f| f["path"] == ".gitignore")
            .unwrap();
        assert_eq!(row["kind"], json!(kind), "{choice}: {row}");
    }
}

#[test]
fn specs_stay_in_git_only_for_a_project_without_github_issues() {
    for (issues, kept) in [(None, true), (Some(false), true), (Some(true), false)] {
        let fx = Fx::new(issues);
        fx.apply_ok("none");
        assert_eq!(
            !fx.ignored(".moai-cli/specs/SPEC-1/spec.md"),
            kept,
            "github_issues {issues:?}:\n{}",
            fx.ignore()
        );
        // The rest of the allow-list does not move with the setting.
        assert!(!fx.ignored(".moai-cli/decisions/D-1.md"), "{issues:?}");
        assert!(fx.ignored(".moai-cli/state/lane.json"), "{issues:?}");
    }
}

#[test]
fn the_plan_shows_the_new_block_line_by_line_and_writes_nothing() {
    let fx = Fx::new(None);
    let plan = fx.plan("ignored");
    let row = Fx::ignore_row(&plan);
    assert_eq!(row["action"], "create", "{row}");
    let detail = row["detail"].as_str().unwrap();
    for line in [
        "managed block lines: 12 added, 0 removed",
        "+ AGENTS.local.md",
        "+ .moai-cli/receipts/",
        "+ .moai-cli/*",
        "+ !.moai-cli/project.yaml",
        "+ !.moai-cli/manifests/",
        "+ !.moai-cli/specs/",
    ] {
        assert!(detail.lines().any(|l| l == line), "{line:?} in:\n{detail}");
    }
    assert!(!fx.root.join(".gitignore").exists(), "the preview wrote");
    // Without the personal file the line is not offered.
    let none = fx.plan("none");
    let detail = Fx::ignore_row(&none)["detail"]
        .as_str()
        .unwrap()
        .to_string();
    assert!(!detail.contains("AGENTS.local.md"), "{detail}");
}

#[test]
fn widening_an_older_block_is_shown_first_and_keeps_the_users_rules() {
    let fx = Fx::new(None);
    let (before, after) = ("node_modules/\n# mine\n", "*.log\n!keep.log\n");
    write_old_state(&fx, before, after);

    let plan = fx.plan("ignored");
    let row = Fx::ignore_row(&plan);
    assert_eq!(row["action"], "merge", "{row}");
    let detail = row["detail"].as_str().unwrap();
    assert!(
        detail.lines().any(|l| l == "+ .moai-cli/*"),
        "the widened lines are listed:\n{detail}"
    );
    assert!(
        detail.lines().any(|l| l == "+ !.moai-cli/config/"),
        "{detail}"
    );
    // The two lines the old block had are kept, so neither is added nor removed.
    assert!(!detail.contains("AGENTS.local.md"), "{detail}");
    assert!(!detail.contains(".moai-cli/receipts/"), "{detail}");
    assert!(
        !detail.lines().any(|l| l.starts_with("- ")),
        "nothing is removed:\n{detail}"
    );
    // The preview wrote nothing.
    assert_eq!(fx.ignore(), format!("{before}{}{after}", old_block()));

    fx.apply_ok("ignored");
    let text = fx.ignore();
    assert!(text.starts_with(before), "{text}");
    assert!(text.ends_with(after), "{text}");
    assert!(text.contains("!.moai-cli/project.yaml\n"), "{text}");
    assert_eq!(text.matches(BEGIN).count(), 1, "{text}");
    // Once widened, the next preview has nothing to do.
    let again = fx.plan("ignored");
    assert_eq!(Fx::ignore_row(&again)["action"], "unchanged", "{again}");
}

#[test]
fn a_block_the_user_edited_is_kept() {
    let fx = Fx::new(None);
    fx.apply_ok("ignored");
    let edited = fx.ignore().replace(END, &format!("my-extra-rule\n{END}"));
    fs::write(fx.root.join(".gitignore"), &edited).unwrap();

    // Nothing new to write: the edit stays and the preview has nothing to do for the file.
    let plan = fx.plan("ignored");
    assert_eq!(Fx::ignore_row(&plan)["action"], "unchanged", "{plan}");
    fx.apply_ok("ignored");
    assert_eq!(fx.ignore(), edited, "the edited block changed");
}

#[test]
fn a_block_the_user_edited_blocks_a_change_to_it() {
    let fx = Fx::new(None);
    fx.apply_ok("none");
    let edited = fx.ignore().replace(END, &format!("my-extra-rule\n{END}"));
    fs::write(fx.root.join(".gitignore"), &edited).unwrap();

    // Choosing the personal file now would change the block: the edit makes it a conflict.
    let plan = fx.plan("ignored");
    assert_eq!(Fx::ignore_row(&plan)["action"], "conflict", "{plan}");
    let out = fx.apply(&plan);
    assert_eq!(out["ok"], json!(false), "{out}");
    assert_eq!(fx.ignore(), edited, "the edited block changed");
}

#[test]
fn an_apply_twice_changes_nothing_the_second_time() {
    let fx = Fx::new(None);
    fx.apply_ok("ignored");
    let (text, manifest) = (fx.ignore(), fx.manifest());
    let plan = fx.plan("ignored");
    assert_eq!(Fx::ignore_row(&plan)["action"], "unchanged", "{plan}");
    fx.apply_ok("ignored");
    assert_eq!(fx.ignore(), text);
    assert_eq!(fx.manifest(), manifest);
}

#[test]
fn a_file_git_already_tracks_stays_tracked() {
    let fx = Fx::new(None);
    // A machine-output file somebody committed earlier; the new rules ignore its name.
    let tracked = ".moai-cli/state/lane.json";
    let file = Path::new(&fx.root).join(tracked);
    fs::create_dir_all(file.parent().unwrap()).unwrap();
    fs::write(&file, "{}\n").unwrap();
    git_ok(&fx.root, &["add", "-f", tracked]);
    fx.apply_ok("ignored");
    assert!(fx.ignored(tracked), "the rule should name it");
    let listed = git(&fx.root, &["ls-files", "--", tracked]);
    assert_eq!(listed, tracked, "the file left the index");
    assert_eq!(fs::read_to_string(&file).unwrap(), "{}\n");
}
