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
    /// The `PATH` the binary runs with: a stub `git` folder can come first, or the only entry.
    path: std::cell::RefCell<String>,
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
            path: std::cell::RefCell::new("/usr/bin:/bin".to_string()),
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
            .env("PATH", self.path.borrow().as_str())
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

/// Paths the user rules of the audit's B1 table act on (audit 6897773189).
const RULE_PROBES: [&str; 11] = [
    ".moai-cli/config/.env",
    ".moai-cli/config/credentials.json",
    ".moai-cli/decisions/id_rsa",
    ".moai-cli/docs/private-notes.md",
    ".moai-cli/project.yaml",
    ".moai-cli/moai-cli.json",
    ".moai-cli/moai.json",
    ".moai-cli/manifests/generation.json",
    ".moai-cli/workflows/run.md",
    ".moai-cli/config/sections/language.yaml",
    ".moai-cli/specs/SPEC-1/spec.md",
];

const USER_RULES: [&str; 6] = [
    ".moai-cli/*",
    ".moai-cli/config/",
    ".moai-cli/docs/",
    "*.json",
    "*.yaml",
    ".moai-cli/**",
];

fn ignored_probes(fx: &Fx) -> Vec<&'static str> {
    RULE_PROBES.into_iter().filter(|p| fx.ignored(p)).collect()
}

/// B1 (leader decision 6897797810, B): a rule the user wrote outside the block keeps its effect.
/// Every path one of the six rule shapes ignored before the apply is ignored after it, and the
/// preview names the rule beside the path it left alone.
#[test]
fn a_user_rule_keeps_its_effect_and_the_preview_names_it() {
    for rule in USER_RULES {
        let fx = Fx::new(None);
        fs::write(fx.root.join(".gitignore"), format!("{rule}\n")).unwrap();
        let before = ignored_probes(&fx);
        assert!(!before.is_empty(), "{rule}: the rule ignores nothing here");

        let plan = fx.plan("none");
        let row = Fx::ignore_row(&plan);
        assert_eq!(row["action"], "append", "{rule}: {row}");
        let detail = row["detail"].as_str().unwrap();
        let named = format!("left ignored by your rule `{rule}` in .gitignore: .moai-cli/");
        assert!(
            detail.lines().any(|l| l.starts_with(&named)),
            "{rule}: the preview does not name the rule and a path:\n{detail}"
        );
        assert_eq!(fx.ignore(), format!("{rule}\n"), "the preview wrote");

        fx.apply_ok("none");
        let after = ignored_probes(&fx);
        for path in &before {
            assert!(
                after.contains(path),
                "{rule}: {path} stopped being ignored\n{}",
                fx.ignore()
            );
        }
        assert!(
            fx.ignore().starts_with(&format!("{rule}\n")),
            "{rule}: the user's rule moved"
        );
    }
}

/// A rule that ignores nothing in the state folder changes nothing about the block.
#[test]
fn a_user_rule_elsewhere_leaves_the_block_as_it_is() {
    let plain = Fx::new(None);
    plain.apply_ok("none");
    let fx = Fx::new(None);
    fs::write(
        fx.root.join(".gitignore"),
        "node_modules/\n# note\n!keep.log\n",
    )
    .unwrap();
    let plan = fx.plan("none");
    let detail = Fx::ignore_row(&plan)["detail"]
        .as_str()
        .unwrap()
        .to_string();
    assert!(!detail.contains("left ignored"), "{detail}");
    // No global excludes file is set here and the default one does not exist: nothing to say.
    assert!(!detail.contains("global excludes file"), "{detail}");
    fx.apply_ok("none");
    let block = |t: String| t[t.find(BEGIN).unwrap()..].to_string();
    assert_eq!(block(fx.ignore()), block(plain.ignore()));
}

/// A user rule that re-includes a path (`!rule`) does not ignore it: that path is no "left
/// ignored" hit and keeps its allow line, while the folder-wide rule still names the others.
#[test]
fn a_user_negation_is_not_an_ignoring_rule() {
    let fx = Fx::new(None);
    fs::write(
        fx.root.join(".gitignore"),
        ".moai-cli/*\n!.moai-cli/project.yaml\n",
    )
    .unwrap();
    let plan = fx.plan("none");
    let detail = Fx::ignore_row(&plan)["detail"]
        .as_str()
        .unwrap()
        .to_string();
    // git answers a file path matched by a `!` rule with that rule; it is not an ignoring one.
    assert!(!detail.contains("your rule `!"), "{detail}");
    assert!(
        !detail.contains("in .gitignore: .moai-cli/project.yaml"),
        "{detail}"
    );
    assert!(
        detail
            .lines()
            .any(|l| l == "left ignored by your rule `.moai-cli/*` in .gitignore: .moai-cli/docs/"),
        "{detail}"
    );
    fx.apply_ok("none");
    let text = fx.ignore();
    let block = &text[text.find(BEGIN).unwrap()..];
    assert!(block.contains("\n!.moai-cli/project.yaml\n"), "{text}");
}

/// Like [`Fx::ignored`], with the global excludes file given; the repository's own
/// `.git/info/exclude` always applies.
fn ignored_with(fx: &Fx, rel: &str, global: &Path) -> bool {
    let setting = format!("core.excludesFile={}", global.display());
    git_raw(
        &fx.root,
        &["-c", &setting, "check-ignore", "-q", "--no-index", rel],
    )
    .success
}

fn ignored_probes_with(fx: &Fx, global: &Path) -> Vec<&'static str> {
    RULE_PROBES
        .into_iter()
        .filter(|p| ignored_with(fx, p, global))
        .collect()
}

/// B1 scope (leader decision 6897897560, 5 and 6): rules in `.git/info/exclude` keep their effect
/// too, and the preview names the file they stand in.
#[test]
fn rules_in_info_exclude_keep_their_effect_and_the_preview_names_the_file() {
    let fx = Fx::new(None);
    let empty = fx.home.join("empty-global");
    fs::write(&empty, "").unwrap();
    fs::write(
        fx.root.join(".git/info/exclude"),
        ".moai-cli/config/\n*.yaml\n",
    )
    .unwrap();
    let before = ignored_probes_with(&fx, &empty);
    assert!(before.contains(&".moai-cli/config/.env"), "{before:?}");
    assert!(before.contains(&".moai-cli/project.yaml"), "{before:?}");

    let plan = fx.plan("none");
    let detail = Fx::ignore_row(&plan)["detail"]
        .as_str()
        .unwrap()
        .to_string();
    for line in [
        "left ignored by your rule `.moai-cli/config/` in .git/info/exclude: .moai-cli/config/",
        "left ignored by your rule `*.yaml` in .git/info/exclude: .moai-cli/project.yaml",
    ] {
        assert!(detail.lines().any(|l| l == line), "{line:?} in:\n{detail}");
    }
    fx.apply_ok("none");
    let after = ignored_probes_with(&fx, &empty);
    for path in &before {
        assert!(
            after.contains(path),
            "{path} stopped being ignored\n{}",
            fx.ignore()
        );
    }
    assert!(
        !fx.ignore().contains("!.moai-cli/config/"),
        "{}",
        fx.ignore()
    );
}

/// B1 scope (5, 6): rules in the global excludes file git reads keep their effect; the preview
/// names that file with the home folder written as `~`.
#[test]
fn rules_in_the_global_excludes_file_keep_their_effect_and_the_preview_names_the_file() {
    let fx = Fx::new(None);
    let global = fx.home.join("global-ignore");
    fs::write(&global, ".moai-cli/docs/\n*.json\nid_rsa\n").unwrap();
    fs::write(
        fx.home.join(".gitconfig"),
        "[core]\n\texcludesFile = ~/global-ignore\n",
    )
    .unwrap();
    let before = ignored_probes_with(&fx, &global);
    assert!(
        before.contains(&".moai-cli/docs/private-notes.md"),
        "{before:?}"
    );
    assert!(before.contains(&".moai-cli/moai.json"), "{before:?}");

    let plan = fx.plan("none");
    let detail = Fx::ignore_row(&plan)["detail"]
        .as_str()
        .unwrap()
        .to_string();
    for line in [
        "left ignored by your rule `.moai-cli/docs/` in ~/global-ignore: .moai-cli/docs/",
        "left ignored by your rule `*.json` in ~/global-ignore: .moai-cli/moai.json",
    ] {
        assert!(detail.lines().any(|l| l == line), "{line:?} in:\n{detail}");
    }
    assert!(
        !detail.contains(fx.home.to_str().unwrap()),
        "the home folder is written out:\n{detail}"
    );
    fx.apply_ok("none");
    let after = ignored_probes_with(&fx, &global);
    for path in &before {
        assert!(
            after.contains(path),
            "{path} stopped being ignored\n{}",
            fx.ignore()
        );
    }
}

/// B1 scope (7): a global excludes file that is set but missing does not fail the plan; the
/// preview says so in one line and the block is written without it.
#[test]
fn a_global_excludes_file_that_is_missing_does_not_fail_the_plan() {
    let fx = Fx::new(None);
    fs::write(
        fx.home.join(".gitconfig"),
        "[core]\n\texcludesFile = ~/no-such-ignore\n",
    )
    .unwrap();
    let plan = fx.plan("none");
    let row = Fx::ignore_row(&plan);
    assert_eq!(row["action"], "create", "{row}");
    let detail = row["detail"].as_str().unwrap();
    let line = "global excludes file ~/no-such-ignore is missing or unreadable; its rules are not considered";
    assert_eq!(detail.lines().filter(|l| *l == line).count(), 1, "{detail}");
    fx.apply_ok("none");
    assert!(
        fx.ignore().contains("!.moai-cli/config/\n"),
        "{}",
        fx.ignore()
    );
}

/// B1 condition 4: when the user later removes the rule, the next plan offers the allow lines
/// as a `merge`, shown line by line first.
#[test]
fn removing_a_user_rule_lets_the_next_plan_offer_the_allow_line() {
    let fx = Fx::new(None);
    fs::write(
        fx.root.join(".gitignore"),
        ".moai-cli/config/\nnode_modules/\n",
    )
    .unwrap();
    fx.apply_ok("none");
    assert!(fx.ignored(".moai-cli/config/.env"), "{}", fx.ignore());
    assert!(
        !fx.ignore().contains("!.moai-cli/config/"),
        "{}",
        fx.ignore()
    );

    let without = fx
        .ignore()
        .strip_prefix(".moai-cli/config/\n")
        .unwrap()
        .to_string();
    fs::write(fx.root.join(".gitignore"), &without).unwrap();
    let plan = fx.plan("none");
    let row = Fx::ignore_row(&plan);
    assert_eq!(row["action"], "merge", "{row}");
    let detail = row["detail"].as_str().unwrap();
    assert!(
        detail.lines().any(|l| l == "+ !.moai-cli/config/"),
        "{detail}"
    );
    assert!(!detail.contains("left ignored"), "{detail}");
    assert_eq!(fx.ignore(), without, "the preview wrote");

    fx.apply_ok("none");
    assert!(fx.ignore().contains("!.moai-cli/config/\n"));
    assert!(!fx.ignored(".moai-cli/config/.env"), "{}", fx.ignore());
}

/// B3: an update moves `.moai-cli/` to `.moai/`. What git tracks in the earlier folder keeps its
/// place in the new one, so a later `git add -A` does not take it out of git.
#[test]
fn what_git_tracks_in_the_earlier_folder_keeps_its_place_in_the_new_one() {
    let fx = Fx::new(Some(true));
    let tracked = [
        ".moai-cli/specs/SPEC-1/spec.md",
        ".moai-cli/state/lane.json",
        ".moai-cli/evidence/run.json",
    ];
    for rel in tracked {
        let file = fx.root.join(rel);
        fs::create_dir_all(file.parent().unwrap()).unwrap();
        fs::write(&file, "{}\n").unwrap();
        git_ok(&fx.root, &["add", "-f", rel]);
    }
    git_ok(
        &fx.root,
        &[
            "-c",
            "user.name=fixture",
            "-c",
            "user.email=fixture@example.invalid",
            "commit",
            "-q",
            "-m",
            "state",
        ],
    );
    // The update: the folder moves and the new generation's marker is written last.
    fs::rename(fx.root.join(".moai-cli"), fx.root.join(".moai")).unwrap();
    fs::write(
        fx.root.join(".moai/moai-cli.json"),
        r#"{"schema_version":1,"generation":"moai-cli"}"#,
    )
    .unwrap();

    let plan = fx.plan("none");
    let row = Fx::ignore_row(&plan);
    let detail = row["detail"].as_str().unwrap();
    for name in ["specs/", "state/", "evidence/"] {
        let note = format!(
            "kept because git tracks .moai-cli/{name}, which the update moves to .moai/{name}"
        );
        assert!(detail.lines().any(|l| l == note), "{name}:\n{detail}");
        assert!(
            detail.lines().any(|l| l == format!("+ !.moai/{name}")),
            "{name}:\n{detail}"
        );
    }
    fx.apply_ok("none");
    git_ok(&fx.root, &["add", "-A"]);
    let listed = git(
        &fx.root,
        &[
            "ls-files",
            "--",
            ".moai/specs",
            ".moai/state",
            ".moai/evidence",
        ],
    );
    for rel in tracked {
        let moved = rel.replacen(".moai-cli/", ".moai/", 1);
        assert!(
            listed.lines().any(|l| l == moved),
            "{moved} left git after `git add -A`:\n{listed}"
        );
    }
    // Machine output nobody tracked is still ignored in the new folder.
    assert!(fx.ignored(".moai/index/atlas.db"), "{}", fx.ignore());
}

/// B2: a `.gitignore` without a block gets one appended, and the plan lists every line.
#[test]
fn the_plan_shows_the_lines_added_to_a_gitignore_without_a_block() {
    let fx = Fx::new(None);
    fs::write(fx.root.join(".gitignore"), "node_modules/\n").unwrap();
    let plan = fx.plan("ignored");
    let row = Fx::ignore_row(&plan);
    assert_eq!(row["action"], "append", "{row}");
    let detail = row["detail"].as_str().unwrap();
    for line in [
        "managed block lines: 12 added, 0 removed",
        "+ AGENTS.local.md",
        "+ .moai-cli/*",
        "+ !.moai-cli/specs/",
    ] {
        assert!(detail.lines().any(|l| l == line), "{line:?} in:\n{detail}");
    }
    assert_eq!(fx.ignore(), "node_modules/\n", "the preview wrote");
}

/// B2: a block that loses a line shows the line it loses.
#[test]
fn the_plan_shows_the_lines_a_changed_block_loses() {
    let fx = Fx::new(None);
    fx.apply_ok("ignored");
    let plan = fx.plan("none");
    let row = Fx::ignore_row(&plan);
    assert_eq!(row["action"], "merge", "{row}");
    let detail = row["detail"].as_str().unwrap();
    assert!(
        detail
            .lines()
            .any(|l| l == "managed block lines: 0 added, 1 removed"),
        "{detail}"
    );
    assert!(detail.lines().any(|l| l == "- AGENTS.local.md"), "{detail}");
}

/// B4: the block's own `AGENTS.local.md` line is not a user rule. With the generation manifest
/// gone, the same choice plans `unchanged` for the file, applies, and records the block again
/// with the kind its text says.
#[test]
fn a_block_listing_the_personal_file_is_not_a_user_rule_without_its_manifest() {
    for (choice, kind) in [
        ("ignored", "agents:ignore-with-local"),
        ("tracked", "agents:ignore-state"),
    ] {
        let fx = Fx::new(None);
        fx.apply_ok(choice);
        let text = fx.ignore();
        fs::remove_file(fx.root.join(".moai-cli/manifests/generation.json")).unwrap();

        let plan = fx.plan(choice);
        assert_eq!(
            Fx::ignore_row(&plan)["action"],
            "unchanged",
            "{choice}: {plan}"
        );
        let out = fx.apply(&plan);
        assert_eq!(out["ok"], json!(true), "{choice}: {out}");
        assert_eq!(fx.ignore(), text, "{choice}: the file changed");
        let row = fx.manifest()["owned_files"]
            .as_array()
            .unwrap()
            .iter()
            .find(|f| f["path"] == ".gitignore")
            .cloned()
            .expect("the block is owned again");
        assert_eq!(row["kind"], json!(kind), "{choice}: {row}");
        // And it stays quiet afterwards.
        let again = fx.plan(choice);
        assert_eq!(
            Fx::ignore_row(&again)["action"],
            "unchanged",
            "{choice}: {again}"
        );
    }
}

/// B4: a user rule `AGENTS.local.md` stands outside the block; the second plan has nothing to do.
#[test]
fn a_user_rule_for_the_personal_file_does_not_make_the_second_plan_add_a_line() {
    let fx = Fx::new(None);
    fs::write(fx.root.join(".gitignore"), "AGENTS.local.md\n").unwrap();
    fx.apply_ok("ignored");
    let text = fx.ignore();
    assert_eq!(text.matches("AGENTS.local.md").count(), 1, "{text}");
    let again = fx.plan("ignored");
    assert_eq!(Fx::ignore_row(&again)["action"], "unchanged", "{again}");
    fx.apply_ok("ignored");
    assert_eq!(fx.ignore(), text);
}

impl Fx {
    /// A project whose update has moved `.moai-cli/` to `.moai/`; `files` (relative to the root,
    /// under `.moai-cli/`) were committed before the move. `rules` is the user's `.gitignore`.
    fn moved(github_issues: Option<bool>, files: &[&str], rules: &str) -> Fx {
        let fx = Fx::new(github_issues);
        fs::write(fx.root.join(".gitignore"), rules).unwrap();
        for rel in files {
            let file = fx.root.join(rel);
            if file.is_dir() {
                // The project setup made this folder; the case wants a file of that name.
                fs::remove_dir_all(&file).unwrap();
            }
            fs::create_dir_all(file.parent().unwrap()).unwrap();
            fs::write(&file, "{}\n").unwrap();
            git_ok(&fx.root, &["add", "-f", "--", rel]);
        }
        git_ok(
            &fx.root,
            &[
                "-c",
                "user.name=fixture",
                "-c",
                "user.email=fixture@example.invalid",
                "commit",
                "-q",
                "--allow-empty",
                "-m",
                "state",
            ],
        );
        fs::rename(fx.root.join(".moai-cli"), fx.root.join(".moai")).unwrap();
        fs::write(
            fx.root.join(".moai/moai-cli.json"),
            r#"{"schema_version":1,"generation":"moai-cli"}"#,
        )
        .unwrap();
        fx
    }

    /// Puts a `git` first on `PATH` that refuses `word` (a subcommand or option) with exit
    /// `code` and a message, and is the real git for everything else.
    fn git_fails_on(&self, word: &str, code: i32) {
        let dir = self.root.parent().unwrap().join("stub-bin");
        fs::create_dir_all(&dir).unwrap();
        let script = format!(
            "#!/bin/sh\nfor a in \"$@\"; do\n  if [ \"$a\" = '{word}' ]; then\n    echo 'fatal: stub refuses {word}' >&2\n    exit {code}\n  fi\ndone\nexec /usr/bin/git \"$@\"\n"
        );
        common::write_executable(&dir.join("git"), &script);
        *self.path.borrow_mut() = format!("{}:/usr/bin:/bin", dir.display());
    }

    /// A `PATH` with no `git` on it.
    fn without_git(&self) {
        let dir = self.root.parent().unwrap().join("empty-bin");
        fs::create_dir_all(&dir).unwrap();
        *self.path.borrow_mut() = dir.display().to_string();
    }

    /// `agents-plan` answered with its envelope, ok or not.
    fn plan_envelope(&self, local_file: &str) -> Value {
        self.run(&["agents-plan"], &Fx::request(local_file))
    }
}

/// N1 (audit 6899590801, leader decision): a user rule wins over a name git tracked in the
/// earlier folder. The new folder keeps ignoring what the rule ignores, no `!` line is written
/// for it, and the preview names the rule, the path and the earlier name whose tracking ends.
#[test]
fn a_user_rule_beats_a_tracked_name_in_the_earlier_folder() {
    // (rule, tracked file, the new-folder name the rule ignores, a path probed there)
    let cases = [
        (
            "state/",
            ".moai-cli/state/lane.json",
            "state/",
            ".moai/state/session-token.json",
        ),
        (
            "config/",
            ".moai-cli/config/sections/language.yaml",
            "config/",
            ".moai/config/credentials.json",
        ),
        (
            "evidence",
            ".moai-cli/evidence/run.json",
            "evidence/",
            ".moai/evidence/run.json",
        ),
    ];
    for (rule, tracked, name, probe) in cases {
        let fx = Fx::moved(Some(true), &[tracked], &format!("{rule}\n"));
        let before = fx.ignored(probe);
        assert!(before, "{rule}: the rule ignores {probe} on its own");

        let plan = fx.plan("none");
        let detail = Fx::ignore_row(&plan)["detail"]
            .as_str()
            .unwrap()
            .to_string();
        let line = format!(
            "tracking ends: your rule `{rule}` in .gitignore ignores .moai/{name}, which git tracks as .moai-cli/{name}"
        );
        assert!(detail.lines().any(|l| l == line), "{rule}:\n{detail}");
        assert!(
            !detail.contains(&format!("kept because git tracks .moai-cli/{name}")),
            "{rule}:\n{detail}"
        );
        assert!(
            !detail.lines().any(|l| l == format!("+ !.moai/{name}")),
            "{rule}:\n{detail}"
        );

        fx.apply_ok("none");
        let text = fx.ignore();
        let block = &text[text.find(BEGIN).unwrap()..];
        assert!(
            !block.contains(&format!("!.moai/{name}\n")),
            "{rule}:\n{text}"
        );
        assert!(
            fx.ignored(probe),
            "{rule}: {probe} stopped being ignored\n{text}"
        );
    }
}

/// N1 boundary: a rule that ignores nothing a tracked name needs leaves the keep line in place.
#[test]
fn a_user_rule_elsewhere_does_not_cost_a_tracked_name_its_line() {
    let fx = Fx::moved(
        Some(true),
        &[".moai-cli/state/lane.json"],
        "node_modules/\n",
    );
    let plan = fx.plan("none");
    let detail = Fx::ignore_row(&plan)["detail"]
        .as_str()
        .unwrap()
        .to_string();
    assert!(detail.lines().any(|l| l == "+ !.moai/state/"), "{detail}");
    assert!(!detail.contains("tracking ends"), "{detail}");
}

/// N2: git failing is not "nothing is tracked". Each question the plan asks git stops it with a
/// message when git answers with an error.
#[test]
fn a_failing_git_stops_the_plan_instead_of_reading_as_nothing() {
    // (word the stub refuses, exit code, rules in .gitignore, text the message must carry)
    let cases = [
        ("ls-files", 128, "", "which files git tracks"),
        ("--git-path", 128, "", ".git/info/exclude is"),
        ("--type=path", 128, "", "core.excludesFile"),
        ("--type=path", 2, "", "core.excludesFile"),
        (
            "check-ignore",
            128,
            "state/\n",
            "check-ignore did not answer",
        ),
        ("check-ignore", 2, "state/\n", "check-ignore did not answer"),
    ];
    for (word, code, rules, expected) in cases {
        let fx = Fx::moved(Some(true), &[".moai-cli/state/lane.json"], rules);
        fx.git_fails_on(word, code);
        let envelope = fx.plan_envelope("none");
        assert_eq!(envelope["ok"], json!(false), "{word} {code}: {envelope}");
        let text = envelope.to_string();
        assert!(text.contains(expected), "{word} {code}: {text}");
        assert!(
            text.contains("stub refuses") || word == "check-ignore",
            "{word}: {text}"
        );
    }
}

/// N2: the two answers that do mean "nothing to read" still give a plan. A folder that is not a
/// git repository has nothing tracked; git missing from `PATH` is said in the preview.
#[test]
fn no_repository_and_no_git_are_not_failures() {
    let fx = Fx::moved(Some(true), &[".moai-cli/state/lane.json"], "");
    fs::remove_dir_all(fx.root.join(".git")).unwrap();
    let plan = fx.plan("none");
    let detail = Fx::ignore_row(&plan)["detail"]
        .as_str()
        .unwrap()
        .to_string();
    assert!(!detail.contains("cannot tell"), "{detail}");
    assert!(!detail.contains("kept because git tracks"), "{detail}");
    assert!(!detail.contains("git is not available"), "{detail}");

    // No git on PATH, no rule anywhere: a plan with one explanatory line.
    let fx = Fx::new(None);
    fx.without_git();
    let plan = fx.plan("none");
    let detail = Fx::ignore_row(&plan)["detail"]
        .as_str()
        .unwrap()
        .to_string();
    let note =
        "git is not available, so your ignore rules and the files git tracks were not considered";
    assert_eq!(detail.lines().filter(|l| *l == note).count(), 1, "{detail}");

    // Comments and blank lines are not rules: with only those, no git is still just a note.
    let fx = Fx::new(None);
    fs::write(
        fx.root.join(".gitignore"),
        "# nothing to ignore here\n\n   \n",
    )
    .unwrap();
    fx.without_git();
    let envelope = fx.plan_envelope("none");
    assert_eq!(envelope["ok"], json!(true), "{envelope}");

    // With a rule to apply, git is needed to read it: the plan stops, and says why.
    let fx = Fx::new(None);
    fs::write(fx.root.join(".gitignore"), "*.json\n").unwrap();
    fx.without_git();
    let envelope = fx.plan_envelope("none");
    assert_eq!(envelope["ok"], json!(false), "{envelope}");
    assert!(
        envelope.to_string().contains("git unavailable"),
        "{envelope}"
    );
}

/// N2: a `.git/info/exclude` that is there and cannot be read stops the plan; one that is not
/// there does not.
#[test]
fn an_unreadable_info_exclude_stops_the_plan() {
    let fx = Fx::new(None);
    let exclude = fx.root.join(".git/info/exclude");
    fs::write(&exclude, [0xff, 0xfe, b'\n']).unwrap();
    let envelope = fx.plan_envelope("none");
    assert_eq!(envelope["ok"], json!(false), "{envelope}");
    assert!(envelope.to_string().contains("not UTF-8"), "{envelope}");

    fs::remove_file(&exclude).unwrap();
    assert_eq!(fx.plan_envelope("none")["ok"], json!(true));
}

/// Git's default global excludes file (`$HOME/.config/git/ignore`, no setting needed) is a
/// source of rules, and the preview writes its home folder as `~`.
#[test]
fn the_default_global_excludes_file_is_read() {
    let fx = Fx::new(None);
    fs::create_dir_all(fx.home.join(".config/git")).unwrap();
    fs::write(fx.home.join(".config/git/ignore"), "*.json\n").unwrap();
    let plan = fx.plan("none");
    let detail = Fx::ignore_row(&plan)["detail"]
        .as_str()
        .unwrap()
        .to_string();
    let line = "left ignored by your rule `*.json` in ~/.config/git/ignore: .moai-cli/moai.json";
    assert!(detail.lines().any(|l| l == line), "{detail}");
}

/// The receipts folder has its own line in the block; a tracked receipts name does not get a
/// second, `!`, line, whether git tracks it as a folder or as a file.
#[test]
fn a_tracked_receipts_name_gets_no_keep_line() {
    for tracked in [".moai-cli/receipts/r.json", ".moai-cli/receipts"] {
        let fx = Fx::moved(Some(true), &[tracked], "");
        let plan = fx.plan("none");
        let detail = Fx::ignore_row(&plan)["detail"]
            .as_str()
            .unwrap()
            .to_string();
        assert!(
            detail.lines().any(|l| l == "+ .moai/receipts/"),
            "{tracked}: the block's own line stays\n{detail}"
        );
        assert!(!detail.contains("!.moai/receipts"), "{tracked}:\n{detail}");
        assert!(!detail.contains("kept because"), "{tracked}:\n{detail}");
        // A file named `receipts` stops the apply itself (no receipts folder); the plan is the
        // evidence for that shape, the apply for the folder shape.
        if tracked.ends_with("r.json") {
            fx.apply_ok("none");
            let text = fx.ignore();
            let block = &text[text.find(BEGIN).unwrap()..];
            assert!(!block.contains("!.moai/receipts"), "{tracked}:\n{text}");
        }
    }
}

/// A tracked name with characters a pattern reads as syntax is written escaped, so the line
/// matches that name and no other; a name with a line break is refused with the reason.
#[test]
fn tracked_names_are_written_as_patterns_that_match_exactly_them() {
    let fx = Fx::moved(
        Some(true),
        &[".moai-cli/we*ird[1].txt", ".moai-cli/trail "],
        "",
    );
    let plan = fx.plan("none");
    let detail = Fx::ignore_row(&plan)["detail"]
        .as_str()
        .unwrap()
        .to_string();
    assert!(
        detail.lines().any(|l| l == r"+ !.moai/we\*ird\[1].txt"),
        "{detail}"
    );
    assert!(detail.lines().any(|l| l == r"+ !.moai/trail\ "), "{detail}");
    fx.apply_ok("none");
    assert!(!fx.ignored(".moai/we*ird[1].txt"), "{}", fx.ignore());
    assert!(!fx.ignored(".moai/trail "), "{}", fx.ignore());
    assert!(
        fx.ignored(".moai/weird1.txt"),
        "the escaped line stays exact\n{}",
        fx.ignore()
    );

    let fx = Fx::moved(Some(true), &[".moai-cli/line\nbreak"], "");
    let envelope = fx.plan_envelope("none");
    assert_eq!(envelope["ok"], json!(false), "{envelope}");
    assert!(envelope.to_string().contains("line break"), "{envelope}");
}
