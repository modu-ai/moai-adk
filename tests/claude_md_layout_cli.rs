//! The files `agents-plan` and `agents-apply` own in a user's project: the managed block in
//! `CLAUDE.md` and the personal file `CLAUDE.local.md` on the user's pick. Never `AGENTS.md`,
//! `AGENTS.local.md` or anything under the project's `.claude` folders (moai's common content is
//! installed in the config folder). Every case runs through the binary in a temporary git
//! repository with an empty temporary HOME.
//!
//! The test names start with `blocks_`, `local_` or `flow_`: the acceptance criteria of
//! `SPEC-20261122-claude-md-layout` filter on those prefixes, and each prefix has tests.

mod common;
use common::git::git_ok;
use common::CleanEnv;
use serde_json::{json, Value};
use std::{
    fs,
    io::Write,
    path::PathBuf,
    process::{Command, Stdio},
};
use tempfile::TempDir;

const PERMISSION: &str = "fixture-approved";
const BEGIN: &str = "<!-- moai:managed:begin id=moai-workflow -->";
const END: &str = "<!-- moai:managed:end id=moai-workflow -->";
const FLOWS: [&str; 4] = ["m-flow", "github-flow", "trunk-based", "git-flow"];

struct Fx {
    _tmp: TempDir,
    root: PathBuf,
    home: PathBuf,
}

impl Fx {
    /// A git repository with a project profile applied; `work_flow` is the profile's flow.
    fn new(work_flow: Option<&str>) -> Fx {
        let tmp = TempDir::new().unwrap();
        let base = fs::canonicalize(tmp.path()).unwrap();
        let (root, home) = (base.join("project"), base.join("home"));
        fs::create_dir_all(root.join("skills/moai/workflows")).unwrap();
        fs::create_dir_all(&home).unwrap();
        let fx = Fx {
            _tmp: tmp,
            root,
            home,
        };
        git_ok(&fx.root, &["init", "-q", "-b", "main"]);
        fx.set_profile(work_flow);
        fx
    }

    fn profile(work_flow: Option<&str>) -> Value {
        let mut profile = json!({
            "schema_version": 2, "project_id": "demo", "purpose": "Build a small editor",
            "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude"],
            "interview_revision": "rc1-fixture-1",
        });
        if let Some(flow) = work_flow {
            profile["work_flow"] = json!({"flow": flow, "integration_branch": "main"});
        }
        profile
    }

    /// Applies a project profile with the given flow (a second call replaces the first).
    fn set_profile(&self, work_flow: Option<&str>) {
        let plan = self.call(&["project-plan"], &Fx::profile(work_flow));
        let digest = plan["plan_digest"].as_str().unwrap().to_string();
        self.call(
            &[
                "project-apply",
                "--plan-digest",
                &digest,
                "--permission-reference",
                PERMISSION,
            ],
            &plan,
        );
    }

    fn p(&self) -> &str {
        self.root.to_str().unwrap()
    }

    fn path(&self, rel: &str) -> PathBuf {
        self.root.join(rel)
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

    fn request(local_file: &str, picks: &[&str]) -> Value {
        let mut r = json!({"protocol_version": 1, "language": "en",
               "workflow_reference": "skills/moai/workflows/", "local_file": local_file});
        if !picks.is_empty() {
            r["picks"] = json!(picks);
        }
        r
    }

    fn plan(&self, local_file: &str, picks: &[&str]) -> Value {
        self.call(&["agents-plan"], &Fx::request(local_file, picks))
    }

    /// The raw answer of an apply of `plan` (an error answer is returned, not asserted away).
    fn apply_raw(&self, plan: &Value) -> Value {
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

    fn apply_ok(&self, local_file: &str, picks: &[&str]) -> Value {
        let out = self.apply_raw(&self.plan(local_file, picks));
        assert_eq!(out["ok"], json!(true), "{out}");
        out
    }

    fn read(&self, rel: &str) -> String {
        fs::read_to_string(self.path(rel)).unwrap()
    }

    fn exists(&self, rel: &str) -> bool {
        fs::symlink_metadata(self.path(rel)).is_ok()
    }
}

fn row<'a>(plan: &'a Value, path: &str) -> Option<&'a Value> {
    plan["files"]
        .as_array()
        .unwrap()
        .iter()
        .find(|f| f["path"] == path)
}

fn action(plan: &Value, path: &str) -> String {
    row(plan, path)
        .map(|f| f["action"].as_str().unwrap().to_string())
        .unwrap_or_else(|| "absent".into())
}

fn paths(plan: &Value) -> Vec<String> {
    plan["files"]
        .as_array()
        .unwrap()
        .iter()
        .map(|f| f["path"].as_str().unwrap().to_string())
        .collect()
}

/// The lines between the two marker lines of the managed block in `text`.
fn block_lines(text: &str) -> Vec<&str> {
    let start = text.find(BEGIN).expect("begin marker") + BEGIN.len();
    let end = text.find(END).expect("end marker");
    text[start..end]
        .lines()
        .skip(1)
        .collect::<Vec<_>>()
        .into_iter()
        .collect()
}

fn warnings(plan: &Value) -> String {
    plan["warnings"].to_string()
}

// --- the managed block lives in CLAUDE.md ---------------------------------------------------------

#[test]
fn blocks_go_into_claude_md_and_never_into_agents_md() {
    let fx = Fx::new(None);
    let plan = fx.plan("none", &[]);
    assert_eq!(action(&plan, "CLAUDE.md"), "create", "{plan}");
    let all = paths(&plan);
    assert!(
        !all.iter()
            .any(|p| p == "AGENTS.md" || p == "AGENTS.local.md"),
        "{all:?}"
    );
    fx.apply_ok("none", &[]);
    let text = fx.read("CLAUDE.md");
    assert!(text.starts_with("# demo\n\n"), "{text}");
    assert!(text.contains(BEGIN) && text.contains(END), "{text}");
    assert!(!fx.exists("AGENTS.md") && !fx.exists("AGENTS.local.md"));
    // A second plan over the written project has nothing left to do for the file.
    assert_eq!(action(&fx.plan("none", &[]), "CLAUDE.md"), "unchanged");
}

#[test]
fn blocks_hold_facts_only_within_fifteen_lines() {
    let fx = Fx::new(Some("m-flow"));
    fx.apply_ok("none", &[]);
    let text = fx.read("CLAUDE.md");
    let lines = block_lines(&text);
    assert!(
        lines.len() <= 15,
        "{} lines:\n{}",
        lines.len(),
        lines.join("\n")
    );
    let joined = lines.join("\n");
    for fact in ["demo", "Rust", "m-flow", "main", ".moai/specs"] {
        assert!(joined.contains(fact), "missing {fact:?} in:\n{joined}");
    }
    // Only the facts REQ-3 lists: no purpose, no development methods, no hosts.
    for left_out in ["Build a small editor", "TDD", "hosts"] {
        assert!(!joined.contains(left_out), "{left_out:?} in:\n{joined}");
    }
    assert!(
        joined.contains("moai check"),
        "the check command is stated:\n{joined}"
    );
    // Issue-era wording and instructions about the process are gone from the block.
    for word in ["Issue", "Issues", "independent audit", "PASS"] {
        assert!(!joined.contains(word), "{word:?} in:\n{joined}");
    }
}

#[test]
fn blocks_wait_for_a_pick_when_agents_md_exists() {
    let fx = Fx::new(None);
    fs::write(fx.path("AGENTS.md"), "# Team rules\n\nRun tests first.\n").unwrap();
    let plan = fx.plan("none", &[]);
    assert_eq!(action(&plan, "CLAUDE.md"), "absent", "{plan}");
    let w = warnings(&plan);
    for word in ["AGENTS.md", "pick", "import", "move"] {
        assert!(w.contains(word), "{word:?} not in {w}");
    }
    // The picks are named in the order the picks step offers them, `move` first and recommended.
    let (at_move, at_import, at_leave) = (
        w.find("`move`").expect("move"),
        w.find("`import`").expect("import"),
        w.find("`leave`").expect("leave"),
    );
    assert!(at_move < at_import && at_import < at_leave, "{w}");
    assert!(w[at_move..at_import].contains("recommended"), "{w}");
    fx.apply_ok("none", &[]);
    assert!(!fx.exists("CLAUDE.md"), "no CLAUDE file without a pick");
    assert_eq!(
        fx.read("AGENTS.md"),
        "# Team rules\n\nRun tests first.\n",
        "AGENTS.md is never changed"
    );
}

/// The writer alone never carries `import` or `move` out: a `CLAUDE.md` holding only the managed
/// block would stop Claude Code reading the rules in `AGENTS.md`. The project picks step
/// (`agents-merge-plan --pick`) writes the block together with the import line or the copied text.
#[test]
fn blocks_are_not_written_by_the_writer_for_the_pick_import_or_move() {
    for pick in ["import", "move"] {
        let fx = Fx::new(None);
        let rules = "# Team rules\n\nAlways run the linter.\n";
        fs::write(fx.path("AGENTS.md"), rules).unwrap();
        let plan = fx.plan("none", &[pick]);
        assert_eq!(action(&plan, "CLAUDE.md"), "absent", "{pick}: {plan}");
        let w = warnings(&plan);
        for word in ["AGENTS.md", "agents-merge", "--pick", pick] {
            assert!(w.contains(word), "{pick}: {word:?} not in {w}");
        }
        fx.apply_ok("none", &[pick]);
        assert!(!fx.exists("CLAUDE.md"), "{pick}: a CLAUDE.md was written");
        assert_eq!(fx.read("AGENTS.md"), rules, "{pick}: AGENTS.md changed");
    }
}

/// Settled projects keep their block refreshed whatever the request picks: the file already holds
/// the block, so it already says how Claude Code reads `AGENTS.md`.
#[test]
fn blocks_are_refreshed_for_a_settled_project_even_with_the_pick_import() {
    let fx = Fx::new(None);
    fx.apply_ok("none", &[]);
    fs::write(fx.path("AGENTS.md"), "# Team rules\n").unwrap();
    assert_eq!(
        action(&fx.plan("none", &["import"]), "CLAUDE.md"),
        "unchanged"
    );
}

#[test]
fn blocks_keep_refreshing_a_settled_project() {
    // A CLAUDE.md that already holds the block keeps getting it refreshed although AGENTS.md exists.
    let fx = Fx::new(None);
    fx.apply_ok("none", &[]);
    fs::write(fx.path("AGENTS.md"), "# Team rules\n").unwrap();
    assert_eq!(action(&fx.plan("none", &[]), "CLAUDE.md"), "unchanged");
    // A CLAUDE.md with the import line and no block gets the block.
    let fx = Fx::new(None);
    fs::write(fx.path("AGENTS.md"), "# Team rules\n").unwrap();
    fs::write(fx.path("CLAUDE.md"), "Our notes\n@AGENTS.md\n").unwrap();
    let plan = fx.plan("none", &[]);
    assert_eq!(action(&plan, "CLAUDE.md"), "append", "{plan}");
    fx.apply_ok("none", &[]);
    let text = fx.read("CLAUDE.md");
    assert!(text.starts_with("Our notes\n@AGENTS.md\n"), "{text}");
    assert!(text.contains(BEGIN));
}

#[test]
fn blocks_append_on_a_new_line_and_keep_every_earlier_byte() {
    let fx = Fx::new(None);
    let earlier = "# Mine\nno final newline";
    fs::write(fx.path("CLAUDE.md"), earlier).unwrap();
    let plan = fx.plan("none", &[]);
    assert_eq!(action(&plan, "CLAUDE.md"), "append", "{plan}");
    fx.apply_ok("none", &[]);
    let text = fx.read("CLAUDE.md");
    assert!(text.starts_with(earlier), "{text}");
    let begin_at = text.find(BEGIN).unwrap();
    assert!(
        text[..begin_at].ends_with('\n'),
        "the block starts on a new line: {text:?}"
    );
    assert_eq!(text.matches(BEGIN).count(), 1);
}

#[test]
fn blocks_edited_after_the_baseline_are_a_conflict_that_writes_nothing() {
    let fx = Fx::new(Some("github-flow"));
    fx.apply_ok("none", &[]);
    let edited = fx.read("CLAUDE.md").replace("Rust", "Rust and Zig");
    fs::write(fx.path("CLAUDE.md"), &edited).unwrap();
    // Nothing new to write over the edit: the plan leaves the file as it is.
    assert_eq!(action(&fx.plan("none", &[]), "CLAUDE.md"), "unchanged");
    // The profile now renders another block: the edit stands in its way, so it is a conflict.
    fx.set_profile(Some("m-flow"));
    let plan = fx.plan("none", &[]);
    assert_eq!(action(&plan, "CLAUDE.md"), "conflict", "{plan}");
    let out = fx.apply_raw(&plan);
    assert_eq!(out["ok"], json!(false), "{out}");
    assert_eq!(fx.read("CLAUDE.md"), edited, "the edited file is kept");
}

#[cfg(unix)]
#[test]
fn blocks_report_a_link_or_non_utf8_file_as_a_conflict_row() {
    // CLAUDE.md is a link: a conflict row, not a failed plan, and nothing is written through it.
    let fx = Fx::new(None);
    fs::write(fx.path("elsewhere.md"), "kept\n").unwrap();
    std::os::unix::fs::symlink("elsewhere.md", fx.path("CLAUDE.md")).unwrap();
    let plan = fx.plan("none", &[]);
    assert_eq!(action(&plan, "CLAUDE.md"), "conflict", "{plan}");
    assert!(row(&plan, "CLAUDE.md").unwrap()["detail"]
        .as_str()
        .unwrap()
        .contains("link"));
    assert_eq!(fx.apply_raw(&plan)["ok"], json!(false));
    assert_eq!(fx.read("elsewhere.md"), "kept\n");

    // CLAUDE.md is not UTF-8 text.
    let fx = Fx::new(None);
    fs::write(fx.path("CLAUDE.md"), [0xff, 0xfe, b'\n']).unwrap();
    let plan = fx.plan("none", &[]);
    assert_eq!(action(&plan, "CLAUDE.md"), "conflict", "{plan}");
    assert_eq!(fx.apply_raw(&plan)["ok"], json!(false));
    assert_eq!(fs::read(fx.path("CLAUDE.md")).unwrap(), [0xff, 0xfe, b'\n']);

    // The personal file is a link and the user picked it: a conflict row, nothing written through it.
    let fx = Fx::new(None);
    fs::write(fx.path("elsewhere.md"), "kept\n").unwrap();
    std::os::unix::fs::symlink("elsewhere.md", fx.path("CLAUDE.local.md")).unwrap();
    let plan = fx.plan("ignored", &["personal-file"]);
    assert_eq!(action(&plan, "CLAUDE.local.md"), "conflict", "{plan}");
    assert_eq!(fx.apply_raw(&plan)["ok"], json!(false));
    assert_eq!(fx.read("elsewhere.md"), "kept\n");
}

// --- the personal file --------------------------------------------------------------------------

#[test]
fn local_file_is_created_only_on_the_users_pick() {
    let fx = Fx::new(None);
    let plan = fx.plan("ignored", &[]);
    assert_eq!(action(&plan, "CLAUDE.local.md"), "absent", "{plan}");
    fx.apply_ok("ignored", &[]);
    assert!(!fx.exists("CLAUDE.local.md") && !fx.exists("AGENTS.local.md"));

    let plan = fx.plan("ignored", &["personal-file"]);
    assert_eq!(action(&plan, "CLAUDE.local.md"), "create", "{plan}");
    fx.apply_ok("ignored", &["personal-file"]);
    assert!(fx.read("CLAUDE.local.md").contains("Personal supplement"));
    // It is created once: a later plan leaves it alone, and so does a deletion.
    assert_eq!(
        action(&fx.plan("ignored", &["personal-file"]), "CLAUDE.local.md"),
        "unchanged"
    );
    fs::remove_file(fx.path("CLAUDE.local.md")).unwrap();
    assert_eq!(
        action(&fx.plan("ignored", &["personal-file"]), "CLAUDE.local.md"),
        "unchanged"
    );
}

#[test]
fn local_file_is_listed_for_git_to_ignore_under_its_new_name() {
    let fx = Fx::new(None);
    fx.apply_ok("ignored", &["personal-file"]);
    let ignore = fx.read(".gitignore");
    assert!(ignore.lines().any(|l| l == "CLAUDE.local.md"), "{ignore}");
    assert!(!ignore.lines().any(|l| l == "AGENTS.local.md"), "{ignore}");
}

/// A line `!CLAUDE.local.md` in the user's own `.gitignore` text re-includes the file, so the
/// managed block must not list it again below that line (git lets the later rule win).
#[test]
fn local_file_is_not_listed_below_a_user_negation_of_its_name() {
    for negation in ["!CLAUDE.local.md", "!/CLAUDE.local.md"] {
        let fx = Fx::new(None);
        let mine = format!("target/\n{negation}\n");
        fs::write(fx.path(".gitignore"), &mine).unwrap();
        fx.apply_ok("ignored", &["personal-file"]);
        let ignore = fx.read(".gitignore");
        assert!(ignore.starts_with(&mine), "{negation}: {ignore}");
        assert!(
            !ignore.lines().any(|l| l == "CLAUDE.local.md"),
            "{negation}: the block lists the file the user re-included: {ignore}"
        );
    }
}

#[test]
fn local_tracked_plans_like_ignored_with_a_preview_note() {
    let fx = Fx::new(None);
    let ignored = fx.plan("ignored", &["personal-file"]);
    let tracked = fx.plan("tracked", &["personal-file"]);
    assert_eq!(paths(&ignored), paths(&tracked));
    for p in paths(&ignored) {
        assert_eq!(action(&ignored, &p), action(&tracked, &p), "{p}");
    }
    assert!(
        warnings(&tracked).contains("tracked"),
        "the preview says tracked is read as ignored: {}",
        warnings(&tracked)
    );
    assert!(!warnings(&ignored).contains("tracked"));
    fx.apply_ok("tracked", &["personal-file"]);
    let ignore = fx.read(".gitignore");
    assert!(ignore.lines().any(|l| l == "CLAUDE.local.md"), "{ignore}");
}

// --- the flow: named in the block, written nowhere in the project's .claude folders -------------------

#[test]
fn flow_is_named_with_its_branch_and_no_path_for_every_flow() {
    for flow in FLOWS {
        let fx = Fx::new(Some(flow));
        fx.apply_ok("none", &[]);
        let text = fx.read("CLAUDE.md");
        let joined = block_lines(&text).join("\n");
        assert!(
            joined.contains(&format!("Work flow: {flow} (integration branch `main`)")),
            "{flow}:\n{joined}"
        );
        for path_word in [".claude", "rules/", "flow.md"] {
            assert!(
                !joined.contains(path_word),
                "{flow}: {path_word:?} in:\n{joined}"
            );
        }
    }
}

#[test]
fn flow_profile_without_a_flow_has_no_flow_line() {
    let fx = Fx::new(None);
    fx.apply_ok("none", &[]);
    let text = fx.read("CLAUDE.md");
    assert!(
        !block_lines(&text).join("\n").contains("Work flow"),
        "{text}"
    );
}

#[test]
fn flow_project_plan_writes_nothing_under_the_projects_claude_folders() {
    let picks: [&[&str]; 3] = [
        &[],
        &["import", "personal-file"],
        &["move", "personal-file"],
    ];
    for flow in FLOWS {
        for chosen in picks {
            let fx = Fx::new(Some(flow));
            fs::write(fx.path("AGENTS.md"), "# Team rules\n").unwrap();
            let plan = fx.plan("ignored", chosen);
            for p in paths(&plan) {
                assert!(!p.starts_with(".claude/"), "{flow} {chosen:?}: plans {p}");
            }
            fx.apply_ok("ignored", chosen);
            assert!(
                !fx.exists(".claude"),
                "{flow} {chosen:?}: the apply made a .claude folder"
            );
        }
    }
}
