//! SPEC-20261121-orca-claude-config-dir: an Orca profile starts Claude Code with the settings
//! folder moved by `CLAUDE_CONFIG_DIR`. `update-plan` and `update-apply` must put the `moai` skill
//! where that session reads it, create nothing under the real `~/.claude`, and leave nothing to
//! apply the second time; without the variable the skill goes to `~/.claude/skills/moai`. Every
//! command runs in a clean environment with a temporary HOME that is also the account home
//! (`MOAI_ACCOUNT_HOME`), so the real HOME is neither read nor written.
#![cfg(unix)]

mod common;

use common::home::TempHome;
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Output, Stdio};

const PERM: &str = "test-permission";

struct World {
    home: TempHome,
    base: PathBuf,
    package: PathBuf,
    profile: PathBuf,
}

impl World {
    fn new() -> Self {
        let home = TempHome::new();
        let base = fs::canonicalize(home.path()).unwrap();
        let package = base.join("package");
        stage(&package);
        // An Orca-style profile folder, outside `~/.claude`.
        let profile = base.join("orca-profile").join("claude");
        World {
            home,
            base,
            package,
            profile,
        }
    }

    /// `moai <args>` as the account's own home, with `CLAUDE_CONFIG_DIR` set when `moved`.
    fn run(&self, moved: bool, args: &[&str], stdin: Option<&[u8]>) -> Output {
        let mut cmd = self.home.command(args);
        cmd.env("HOME", &self.base)
            .env("MOAI_ACCOUNT_HOME", &self.base);
        if moved {
            cmd.env("CLAUDE_CONFIG_DIR", &self.profile);
        }
        let mut child = cmd
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
            .expect("run moai");
        common::feed_stdin(&mut child, stdin.unwrap_or(b""));
        child.wait_with_output().unwrap()
    }

    fn plan(&self, moved: bool) -> Value {
        ok(&self.run(
            moved,
            &["update-plan", self.base.to_str().unwrap(), self.pkg()],
            None,
        ))
    }

    fn apply(&self, moved: bool, plan: &Value) {
        ok(&self.run(
            moved,
            &[
                "update-apply",
                self.base.to_str().unwrap(),
                self.pkg(),
                "--plan-digest",
                plan["plan_digest"].as_str().unwrap(),
                "--permission-reference",
                PERM,
            ],
            Some(plan.to_string().as_bytes()),
        ));
    }

    fn pkg(&self) -> &str {
        self.package.to_str().unwrap()
    }
}

fn stage(dir: &Path) {
    for (rel, body) in [
        (
            "SKILL.md",
            "---\nname: moai\ndescription: fixture package\n---\n# MoAI\n",
        ),
        ("references/routes.json", "{\"routes\":1}\n"),
    ] {
        let p = dir.join(rel);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        fs::write(p, body).unwrap();
    }
}

fn ok(out: &Output) -> Value {
    let v: Value = serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
        panic!(
            "stdout is not JSON ({e}): {} {}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    });
    assert_eq!(out.status.code(), Some(0), "{v}");
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn host<'a>(v: &'a Value, name: &str) -> &'a Value {
    v["hosts"]
        .as_array()
        .unwrap()
        .iter()
        .find(|h| h["host"] == json!(name))
        .unwrap_or_else(|| panic!("no host {name}: {v}"))
}

/// REQ-1: the preview names the moved folder for Claude Code and is not blocked; Codex keeps
/// `~/.agents/skills/moai`.
#[test]
fn profile_plan_targets_the_moved_folder() {
    let w = World::new();
    let plan = w.plan(true);
    let claude = host(&plan, "claude");
    assert_eq!(
        claude["projection_root"],
        json!(w.profile.join("skills/moai").to_str().unwrap()),
        "{plan}"
    );
    assert!(claude["blocked_reason"].is_null(), "{plan}");
    assert_eq!(claude["preflight"], json!("ok"), "{plan}");
    assert_eq!(
        host(&plan, "codex")["projection_root"],
        json!(w.base.join(".agents/skills/moai").to_str().unwrap())
    );
}

/// REQ-2: applying writes `SKILL.md` with the name `moai` where an Orca-profile session reads
/// skills, and creates nothing under the real `~/.claude`.
#[test]
fn profile_apply_puts_the_skill_where_the_session_reads_it() {
    let w = World::new();
    let plan = w.plan(true);
    w.apply(true, &plan);
    let skill = w.profile.join("skills/moai/SKILL.md");
    let text = fs::read_to_string(&skill).unwrap_or_else(|e| panic!("{}: {e}", skill.display()));
    assert!(text.starts_with("---\nname: moai\n"), "{text}");
    assert!(
        !w.base.join(".claude").exists(),
        "nothing may be created under ~/.claude"
    );
}

/// REQ-3: after the apply the next preview has nothing left to change for Claude Code.
#[test]
fn profile_second_plan_has_nothing_to_apply() {
    let w = World::new();
    let first = w.plan(true);
    w.apply(true, &first);
    let second = w.plan(true);
    let files = host(&second, "claude")["files"].as_array().unwrap().clone();
    assert!(!files.is_empty(), "{second}");
    let open: Vec<&Value> = files
        .iter()
        .filter(|f| f["action"] != json!("unchanged"))
        .collect();
    assert!(open.is_empty(), "still to apply: {open:?}");
}

/// REQ-4, control: without the variable the skill goes to `~/.claude/skills/moai`, so the
/// variable is what moves it.
#[test]
fn without_the_variable_the_skill_goes_to_home_claude() {
    let w = World::new();
    let plan = w.plan(false);
    assert_eq!(
        host(&plan, "claude")["projection_root"],
        json!(w.base.join(".claude/skills/moai").to_str().unwrap()),
        "{plan}"
    );
    w.apply(false, &plan);
    assert!(w.base.join(".claude/skills/moai/SKILL.md").is_file());
    assert!(!w.profile.exists(), "the profile folder stays untouched");
}

const ADAPTER: &str = "skills/moai/references/adapters/orca.md";

fn adapter() -> String {
    std::fs::read_to_string(Path::new(env!("CARGO_MANIFEST_DIR")).join(ADAPTER))
        .unwrap_or_else(|e| panic!("{ADAPTER}: {e}"))
}

fn flat(text: &str) -> String {
    text.split_whitespace().collect::<Vec<_>>().join(" ")
}

fn table_cells(text: &str, row: &str) -> Vec<String> {
    let line = text
        .lines()
        .find(|l| l.starts_with(&format!("| {row} |")))
        .unwrap_or_else(|| panic!("no table row {row}"));
    line.trim_matches('|')
        .split('|')
        .map(|c| c.trim().to_string())
        .collect()
}

/// REQ-5: SETUP and UPDATE work with Orca because the skill copy follows `CLAUDE_CONFIG_DIR`.
#[test]
fn adapter_rows_setup_and_update_work_with_orca_through_the_moved_folder() {
    let text = adapter();
    for row in ["SETUP", "UPDATE"] {
        let cells = table_cells(&text, row);
        assert_eq!(cells.len(), 4, "row {row}: {cells:?}");
        assert!(
            cells[1].starts_with("works"),
            "row {row}, With Orca must start with works: {}",
            cells[1]
        );
        assert!(
            cells[1].contains("CLAUDE_CONFIG_DIR"),
            "row {row} names the variable: {}",
            cells[1]
        );
        assert!(
            !cells[1].contains("unsupported"),
            "row {row} still says unsupported: {}",
            cells[1]
        );
    }
}

/// REQ-6: the settings folder of an Orca profile is not an unsupported item any more.
#[test]
fn adapter_rows_unsupported_list_drops_the_profile_settings_folder() {
    let text = adapter();
    let list = text
        .lines()
        .find(|l| l.starts_with("Unsupported today"))
        .expect("the section has an `Unsupported today` paragraph");
    for gone in [
        "the settings folder of an Orca profile (`CLAUDE_CONFIG_DIR`), which `moai update` does not see",
        "SETUP and UPDATE are unsupported with Orca",
    ] {
        assert!(!list.contains(gone), "the list still says `{gone}`");
    }
    assert!(
        list.contains("a Codex or OMP leader's periodic run"),
        "the other unsupported item stays: {list}"
    );
}

/// REQ-7: what the measurement did not verify is named, with the pages it rests on: Claude Code's
/// folder page for the variable, Orca's settings page for the per-account folder.
#[test]
fn adapter_unverified_names_the_live_profile_session_and_its_page() {
    let text = flat(&adapter());
    for phrase in [
        "no live Orca profile session was run",
        "https://code.claude.com/docs/en/claude-directory.md",
        "Orca's pages say each Claude account keeps its own config folder but never name `CLAUDE_CONFIG_DIR`",
        "that Orca sets it is unverified",
        "https://orca.dev/docs/settings",
    ] {
        assert!(text.contains(phrase), "the adapter must say {phrase:?}");
    }
}

/// REQ-8: reuse is not a single-leader guarantee; what the automations page leaves out is named;
/// a leader checks for a running leader before it starts one.
#[test]
fn adapter_automations_reuse_is_not_a_single_leader_guarantee() {
    let text = adapter();
    let line = text
        .lines()
        .find(|l| l.starts_with("Automations."))
        .expect("the adapter has an `Automations.` paragraph");
    let line = flat(line);
    for phrase in [
        "`--reuse-session` only continues in the previous live automation terminal",
        "https://orca.dev/docs/cli/automations, read 2026-11-21T01:31:19Z",
        "says nothing about a manual run that overlaps a running one, a restart, a sleeping machine or the missed-run grace window",
        "a leader checks whether a leader already runs before it starts one by a manual run or after a restart",
    ] {
        assert!(line.contains(phrase), "the paragraph must say {phrase:?}");
    }
}
