//! Card C-79 (SPEC-20261117-host-home-resolver): routing, agents, the output style, the skill
//! copy and the Codex doctor find the same host folder. `CLAUDE_CONFIG_DIR` moves `~/.claude`
//! (code.claude.com/docs/en/claude-directory.md) and `CODEX_HOME` moves `~/.codex`
//! (learn.chatgpt.com/docs/agent-configuration/agents-md); an empty or relative value is
//! refused. The variables count only for the account's own home: `MOAI_ACCOUNT_HOME` names it
//! here, so a temporary HOME never reaches the real config folder. Every command runs in a
//! clean environment with a temporary HOME; the real HOME is not read or written.
#![cfg(unix)]

mod common;

use common::home::TempHome;
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::Output;

const PERM: &str = "test-permission";

struct World {
    home: TempHome,
    base: PathBuf,
    package: PathBuf,
}

impl World {
    fn new() -> Self {
        let home = TempHome::new();
        let base = fs::canonicalize(home.path()).unwrap();
        let repo = Path::new(env!("CARGO_MANIFEST_DIR"));
        World {
            home,
            base,
            package: repo.join("skills/moai"),
        }
    }

    fn h(&self) -> &str {
        self.base.to_str().unwrap()
    }

    fn pkg(&self) -> &str {
        self.package.to_str().unwrap()
    }

    /// A folder outside HOME's host folders.
    fn outside(&self, name: &str) -> PathBuf {
        self.base.join("elsewhere").join(name)
    }

    /// `moai <args>` with `vars` set; `account` names the account home.
    fn run(&self, vars: &[(&str, &str)], account: bool, args: &[&str]) -> Output {
        let mut cmd = self.home.command(args);
        cmd.env("HOME", &self.base);
        if account {
            cmd.env("MOAI_ACCOUNT_HOME", &self.base);
        }
        for (k, v) in vars {
            cmd.env(k, v);
        }
        cmd.output().expect("run moai")
    }

    /// [`World::run`] with values that need not be UTF-8 and `stdin` fed to the command.
    fn run_os(
        &self,
        vars: &[(&str, &std::ffi::OsStr)],
        args: &[&str],
        stdin: Option<&[u8]>,
    ) -> Output {
        let mut cmd = self.home.command(args);
        cmd.env("HOME", &self.base)
            .env("MOAI_ACCOUNT_HOME", &self.base);
        for (k, v) in vars {
            cmd.env(k, v);
        }
        let mut child = cmd
            .stdin(std::process::Stdio::piped())
            .stdout(std::process::Stdio::piped())
            .stderr(std::process::Stdio::piped())
            .spawn()
            .expect("run moai");
        common::feed_stdin(&mut child, stdin.unwrap_or(b""));
        child.wait_with_output().unwrap()
    }
}

/// The refusal: exit 1 and its message.
fn refused(out: &Output) -> String {
    let v: Value = serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
        panic!(
            "stdout is not JSON ({e}): {} {}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    });
    assert_eq!(out.status.code(), Some(1), "{v}");
    v["error"]["message"].as_str().unwrap().to_string()
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

fn s(p: &Path) -> String {
    p.to_str().unwrap().to_string()
}

/// REQ-1, REQ-2: the routing section goes to `$CLAUDE_CONFIG_DIR/CLAUDE.md`, and revert takes it
/// back from there; nothing under `~/.claude`, and no file in the folder `CODEX_HOME` names (moai
/// no longer writes there).
#[test]
fn routing_writes_and_reverts_in_the_relocated_folders() {
    let w = World::new();
    let conf = w.outside("claude-conf");
    let cx = w.outside("codex-home");
    let vars = [("CLAUDE_CONFIG_DIR", s(&conf)), ("CODEX_HOME", s(&cx))];
    let vars: Vec<(&str, &str)> = vars.iter().map(|(k, v)| (*k, v.as_str())).collect();
    let plan = ok(&w.run(&vars, true, &["home-routing-plan", w.h(), w.pkg()]));
    assert_eq!(
        host(&plan, "claude")["file"],
        json!(s(&conf.join("CLAUDE.md"))),
        "{plan}"
    );
    assert_eq!(plan["hosts"].as_array().unwrap().len(), 1, "{plan}");
    let applied = ok(&w.run(
        &vars,
        true,
        &[
            "home-routing-apply",
            w.h(),
            w.pkg(),
            "--plan-digest",
            plan["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            PERM,
        ],
    ));
    assert!(conf.join("CLAUDE.md").is_file(), "{applied}");
    assert!(!cx.join("AGENTS.md").exists());
    assert!(!w.base.join(".claude").exists());
    assert!(!w.base.join(".codex").exists());
    let receipt = applied["receipt"].as_str().unwrap().to_string();
    ok(&w.run(
        &vars,
        true,
        &[
            "home-routing-revert",
            w.h(),
            "--receipt",
            &receipt,
            "--permission-reference",
            PERM,
        ],
    ));
    assert!(!conf.join("CLAUDE.md").exists());
}

/// REQ-5: a home that is not the account's own home keeps the folders under it, whatever the
/// variables say: a temporary HOME never reaches the real config folder.
#[test]
fn the_variables_count_only_for_the_account_home() {
    let w = World::new();
    let conf = w.outside("claude-conf");
    let cx = w.outside("codex-home");
    let (c, x) = (s(&conf), s(&cx));
    let vars = [
        ("CLAUDE_CONFIG_DIR", c.as_str()),
        ("CODEX_HOME", x.as_str()),
    ];
    let plan = ok(&w.run(&vars, false, &["home-routing-plan", w.h(), w.pkg()]));
    assert_eq!(
        host(&plan, "claude")["file"],
        json!(s(&w.base.join(".claude/CLAUDE.md"))),
        "{plan}"
    );
    // Another account home named: the same.
    let other = w.outside("someone");
    let mut cmd = w.home.command(&["home-routing-plan", w.h(), w.pkg()]);
    cmd.env("HOME", &w.base)
        .env("MOAI_ACCOUNT_HOME", &other)
        .env("CLAUDE_CONFIG_DIR", &conf);
    let plan = ok(&cmd.output().unwrap());
    assert_eq!(
        host(&plan, "claude")["file"],
        json!(s(&w.base.join(".claude/CLAUDE.md")))
    );
}

/// REQ-3: an empty or relative value is refused by name, for that host only.
#[test]
fn an_empty_or_relative_variable_blocks_its_host_by_name() {
    let w = World::new();
    for (value, says) in [("", "set but empty"), ("conf", "must be an absolute path")] {
        let plan = ok(&w.run(
            &[("CLAUDE_CONFIG_DIR", value)],
            true,
            &["home-routing-plan", w.h(), w.pkg()],
        ));
        let claude = host(&plan, "claude");
        assert_eq!(claude["preflight"], json!("blocked"), "{plan}");
        let why = claude["blocked_reason"].as_str().unwrap();
        assert!(
            why.contains("CLAUDE_CONFIG_DIR") && why.contains(says),
            "{why}"
        );
        // Agents and the output style block the Claude host the same way.
        let repo = Path::new(env!("CARGO_MANIFEST_DIR"));
        let plan = ok(&w.run(
            &[("CLAUDE_CONFIG_DIR", value)],
            true,
            &["agents-home-plan", w.h(), w.pkg(), &s(&repo.join("agents"))],
        ));
        assert_eq!(
            host(&plan, "claude")["preflight"],
            json!("blocked"),
            "{plan}"
        );
        let style = ok(&w.run(
            &[("CLAUDE_CONFIG_DIR", value)],
            true,
            &["home-style-plan", w.h(), w.pkg()],
        ));
        let blocked: Vec<&Value> = style["targets"]
            .as_array()
            .unwrap()
            .iter()
            .filter(|t| t["host"] == json!("claude"))
            .collect();
        assert!(!blocked.is_empty(), "{style}");
        assert!(
            blocked.iter().all(|t| t["action"] == json!("blocked")
                && t["detail"].as_str().unwrap().contains("CLAUDE_CONFIG_DIR")),
            "{style}"
        );
    }
}

/// REQ-1: agents and the output style follow `CLAUDE_CONFIG_DIR`; no other host is planned.
#[test]
fn agents_and_the_output_style_follow_the_variables() {
    let w = World::new();
    let conf = w.outside("claude-conf");
    let cx = w.outside("codex-home");
    let (c, x) = (s(&conf), s(&cx));
    let vars = [
        ("CLAUDE_CONFIG_DIR", c.as_str()),
        ("CODEX_HOME", x.as_str()),
    ];
    let repo = Path::new(env!("CARGO_MANIFEST_DIR"));
    let agents = s(&repo.join("agents"));
    let plan = ok(&w.run(&vars, true, &["agents-home-plan", w.h(), w.pkg(), &agents]));
    assert_eq!(
        host(&plan, "claude")["dir"],
        json!(s(&conf.join("agents"))),
        "{plan}"
    );
    assert_eq!(plan["hosts"].as_array().unwrap().len(), 1, "{plan}");
    let applied = ok(&w.run(
        &vars,
        true,
        &[
            "agents-home-apply",
            w.h(),
            w.pkg(),
            &agents,
            "--plan-digest",
            plan["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            PERM,
        ],
    ));
    assert!(
        fs::read_dir(conf.join("agents")).unwrap().count() > 0,
        "{applied}"
    );
    assert!(!w.base.join(".claude").exists());

    let style = ok(&w.run(&vars, true, &["home-style-plan", w.h(), w.pkg()]));
    let paths: Vec<(String, String)> = style["targets"]
        .as_array()
        .unwrap()
        .iter()
        .map(|t| {
            (
                t["kind"].as_str().unwrap().to_string(),
                t["path"].as_str().unwrap().to_string(),
            )
        })
        .collect();
    assert_eq!(
        paths,
        vec![
            (
                "style_file".to_string(),
                s(&conf.join("output-styles/MoAI-Easy.md"))
            ),
            ("setting".to_string(), s(&conf.join("settings.json"))),
        ],
        "{style}"
    );
}

/// A minimal skill package for `update-plan`.
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

/// REQ-1: the Claude skill copy goes to `$CLAUDE_CONFIG_DIR/skills/moai`; no copy is planned
/// for `~/.agents/skills/moai` or under `CODEX_HOME`.
#[test]
fn the_skill_copy_follows_claude_config_dir() {
    let w = World::new();
    let conf = w.outside("claude-conf");
    let cx = w.outside("codex-home");
    let (c, x) = (s(&conf), s(&cx));
    let vars = [
        ("CLAUDE_CONFIG_DIR", c.as_str()),
        ("CODEX_HOME", x.as_str()),
    ];
    let pkg = w.base.join("stage");
    stage(&pkg);
    let plan = ok(&w.run(&vars, true, &["update-plan", w.h(), &s(&pkg)]));
    assert_eq!(
        host(&plan, "claude")["projection_root"],
        json!(s(&conf.join("skills/moai"))),
        "{plan}"
    );
    assert_eq!(plan["hosts"].as_array().unwrap().len(), 1, "{plan}");
    // A relative value blocks the Claude host's copy by name.
    let plan = ok(&w.run(
        &[("CLAUDE_CONFIG_DIR", "conf")],
        true,
        &["update-plan", w.h(), &s(&pkg)],
    ));
    let claude = host(&plan, "claude");
    assert!(claude.to_string().contains("CLAUDE_CONFIG_DIR"), "{plan}");
}

/// REQ-4: a relocated folder that is a link is refused, as a linked `~/.claude` is.
#[test]
fn a_linked_config_folder_is_refused_not_followed() {
    let w = World::new();
    let real = w.outside("real-conf");
    fs::create_dir_all(&real).unwrap();
    let link = w.outside("linked-conf");
    std::os::unix::fs::symlink(&real, &link).unwrap();
    let l = s(&link);
    let plan = ok(&w.run(
        &[("CLAUDE_CONFIG_DIR", l.as_str())],
        true,
        &["home-routing-plan", w.h(), w.pkg()],
    ));
    let claude = host(&plan, "claude");
    assert_eq!(claude["preflight"], json!("blocked"), "{plan}");
    assert!(
        claude["blocked_reason"]
            .as_str()
            .unwrap()
            .contains("symlink"),
        "{plan}"
    );
    let repo = Path::new(env!("CARGO_MANIFEST_DIR"));
    let plan = ok(&w.run(
        &[("CLAUDE_CONFIG_DIR", l.as_str())],
        true,
        &["agents-home-plan", w.h(), w.pkg(), &s(&repo.join("agents"))],
    ));
    assert_eq!(
        host(&plan, "claude")["preflight"],
        json!("blocked"),
        "{plan}"
    );
    let pkg = w.base.join("stage");
    stage(&pkg);
    let plan = ok(&w.run(
        &[("CLAUDE_CONFIG_DIR", l.as_str())],
        true,
        &["update-plan", w.h(), &s(&pkg)],
    ));
    assert!(
        host(&plan, "claude").to_string().contains("symlink"),
        "{plan}"
    );
    assert_eq!(
        fs::read_dir(&real).unwrap().count(),
        0,
        "nothing written through the link"
    );

    // The same with the linked folder outside HOME: the check starts at the relocated
    // folder, not at HOME.
    let away = tempfile::tempdir().unwrap();
    let away_base = fs::canonicalize(away.path()).unwrap();
    let far_real = away_base.join("real-conf");
    fs::create_dir_all(&far_real).unwrap();
    let far_link = away_base.join("linked-conf");
    std::os::unix::fs::symlink(&far_real, &far_link).unwrap();
    assert!(!far_link.starts_with(&w.base));
    let fl = s(&far_link);
    let plan = ok(&w.run(
        &[("CLAUDE_CONFIG_DIR", fl.as_str())],
        true,
        &["update-plan", w.h(), &s(&pkg)],
    ));
    assert!(
        host(&plan, "claude").to_string().contains("symlink"),
        "{plan}"
    );
    let plan = ok(&w.run(
        &[("CLAUDE_CONFIG_DIR", fl.as_str())],
        true,
        &["home-routing-plan", w.h(), w.pkg()],
    ));
    assert_eq!(
        host(&plan, "claude")["preflight"],
        json!("blocked"),
        "{plan}"
    );
    assert_eq!(fs::read_dir(&far_real).unwrap().count(), 0);
}

/// REQ-7 (audit B1): the hook writer refuses a project whose `.claude` is the folder
/// `CLAUDE_CONFIG_DIR` names, or a project inside that folder; nothing is written. A project
/// elsewhere still plans.
#[test]
fn hooks_refuse_the_relocated_user_folders() {
    let w = World::new();
    let (host, var, dir) = ("claude", "CLAUDE_CONFIG_DIR", ".claude");
    let project = w.outside(&format!("{host}-project"));
    fs::create_dir_all(project.join(dir)).unwrap();
    let user = project.join(dir);
    let inside = user.join("sub");
    fs::create_dir_all(&inside).unwrap();
    let u = std::ffi::OsString::from(&user);
    for target in [&project, &inside] {
        let m = refused(&w.run_os(
            &[(var, u.as_os_str())],
            &["hooks-plan", &s(target), "--host", host],
            None,
        ));
        assert!(m.contains(var), "{host} {}: {m}", target.display());
        let m = refused(&w.run_os(
            &[(var, u.as_os_str())],
            &[
                "hooks-apply",
                &s(target),
                "--host",
                host,
                "--plan-digest",
                "sha256:00",
                "--permission-reference",
                PERM,
            ],
            None,
        ));
        assert!(m.contains(var), "{host} apply: {m}");
    }
    assert!(!user.join("hooks.json").exists());
    assert!(!user.join("settings.json").exists());
    // An empty or relative value is refused by name too.
    for bad in ["", "rel"] {
        let m = refused(&w.run_os(
            &[(var, std::ffi::OsStr::new(bad))],
            &["hooks-plan", &s(&project), "--host", host],
            None,
        ));
        assert!(m.contains(var), "{host} {bad:?}: {m}");
    }
    // A project elsewhere plans with the variable set.
    let other = w.outside(&format!("{host}-other"));
    fs::create_dir_all(&other).unwrap();
    ok(&w.run_os(
        &[(var, u.as_os_str())],
        &["hooks-plan", &s(&other), "--host", host],
        None,
    ));
}

/// REQ-5, D10 (user decision 2026-11-17, "기본 폴더 링크만 허용"): a folder in the relocated host
/// folder that is a link is followed only when it resolves to the same folder under the account
/// home's default host folder; preview and apply then use that folder. A link anywhere else
/// blocks the host, and the reason names the link.
#[test]
fn a_link_to_the_default_folder_is_followed_and_any_other_link_is_blocked() {
    let w = World::new();
    let repo = Path::new(env!("CARGO_MANIFEST_DIR"));
    let agents_src = s(&repo.join("agents"));
    let pkg = w.base.join("stage");
    stage(&pkg);
    // Allowed: conf/skills → ~/.claude/skills, conf/agents → ~/.claude/agents.
    let conf = w.outside("claude-conf");
    fs::create_dir_all(&conf).unwrap();
    for (link, target) in [
        (conf.join("skills"), w.base.join(".claude/skills")),
        (conf.join("agents"), w.base.join(".claude/agents")),
    ] {
        fs::create_dir_all(&target).unwrap();
        std::os::unix::fs::symlink(&target, &link).unwrap();
    }
    let c = std::ffi::OsString::from(&conf);
    let vars = [("CLAUDE_CONFIG_DIR", c.as_os_str())];
    let plan = ok(&w.run_os(&vars, &["update-plan", w.h(), &s(&pkg)], None));
    let claude = host(&plan, "claude");
    assert_eq!(
        claude["projection_root"],
        json!(s(&w.base.join(".claude/skills/moai"))),
        "{plan}"
    );
    assert!(claude["blocked_reason"].is_null(), "{plan}");
    let out = w.run_os(
        &vars,
        &[
            "update-apply",
            w.h(),
            &s(&pkg),
            "--plan-digest",
            plan["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            PERM,
        ],
        Some(plan.to_string().as_bytes()),
    );
    ok(&out);
    assert!(w.base.join(".claude/skills/moai/SKILL.md").is_file());

    let plan = ok(&w.run_os(
        &vars,
        &["agents-home-plan", w.h(), w.pkg(), &agents_src],
        None,
    ));
    assert_eq!(
        host(&plan, "claude")["dir"],
        json!(s(&w.base.join(".claude/agents"))),
        "{plan}"
    );
    assert_eq!(host(&plan, "claude")["preflight"], json!("ok"));
    ok(&w.run_os(
        &vars,
        &[
            "agents-home-apply",
            w.h(),
            w.pkg(),
            &agents_src,
            "--plan-digest",
            plan["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            PERM,
        ],
        None,
    ));
    assert!(fs::read_dir(w.base.join(".claude/agents")).unwrap().count() > 0);
    assert!(!w.base.join(".codex").exists());

    // Blocked: links to any other folder, or to a default folder that is not there.
    let conf2 = w.outside("claude-conf-2");
    let elsewhere = w.outside("other-skills");
    fs::create_dir_all(&conf2).unwrap();
    fs::create_dir_all(&elsewhere).unwrap();
    std::os::unix::fs::symlink(&elsewhere, conf2.join("skills")).unwrap();
    std::os::unix::fs::symlink(&elsewhere, conf2.join("agents")).unwrap();
    let c2 = std::ffi::OsString::from(&conf2);
    let vars2 = [("CLAUDE_CONFIG_DIR", c2.as_os_str())];
    let plan = ok(&w.run_os(&vars2, &["update-plan", w.h(), &s(&pkg)], None));
    let why = host(&plan, "claude")["blocked_reason"]
        .as_str()
        .unwrap_or_default()
        .to_string();
    assert!(why.contains(&s(&conf2.join("skills"))), "{plan}");
    let plan = ok(&w.run_os(
        &vars2,
        &["agents-home-plan", w.h(), w.pkg(), &agents_src],
        None,
    ));
    let claude = host(&plan, "claude");
    assert_eq!(claude["preflight"], json!("blocked"), "{plan}");
    assert!(
        claude["blocked_reason"]
            .as_str()
            .unwrap()
            .contains(&s(&conf2.join("agents"))),
        "{plan}"
    );
    assert_eq!(fs::read_dir(&elsewhere).unwrap().count(), 0);
    // A link to a default folder that is not there blocks the Claude host too.
    let missing = w.outside("claude-conf-3");
    fs::create_dir_all(&missing).unwrap();
    std::os::unix::fs::symlink(w.base.join(".claude/nowhere"), missing.join("agents")).unwrap();
    let x3 = std::ffi::OsString::from(&missing);
    let plan = ok(&w.run_os(
        &[("CLAUDE_CONFIG_DIR", x3.as_os_str())],
        &["agents-home-plan", w.h(), w.pkg(), &agents_src],
        None,
    ));
    assert_eq!(
        host(&plan, "claude")["preflight"],
        json!("blocked"),
        "{plan}"
    );
}

/// Audit N1 (M5): the output style refuses a relocated Claude folder that is a link.
#[test]
fn the_output_style_refuses_a_linked_config_folder() {
    let w = World::new();
    let real = w.outside("real-conf");
    fs::create_dir_all(&real).unwrap();
    let link = w.outside("linked-conf");
    std::os::unix::fs::symlink(&real, &link).unwrap();
    let l = s(&link);
    let style = ok(&w.run(
        &[("CLAUDE_CONFIG_DIR", l.as_str())],
        true,
        &["home-style-plan", w.h(), w.pkg()],
    ));
    let claude: Vec<&Value> = style["targets"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|t| t["host"] == json!("claude"))
        .collect();
    assert!(!claude.is_empty(), "{style}");
    assert!(
        claude.iter().all(|t| t["action"] == json!("blocked")
            && t["detail"].as_str().unwrap().contains("symlink")),
        "{style}"
    );
}

/// Audit N1 (M6): the Claude settings guard refuses an empty or relative variable by name and
/// writes nothing.
#[test]
fn the_user_scope_guards_refuse_empty_and_relative_values() {
    let w = World::new();
    let project = w.outside("project");
    fs::create_dir_all(project.join(".claude")).unwrap();
    for bad in ["", "rel"] {
        let m = refused(&w.run_os(
            &[("CLAUDE_CONFIG_DIR", std::ffi::OsStr::new(bad))],
            &["claude-settings-plan", &s(&project)],
            Some(br#"{"outputStyle":"MoAI-Easy"}"#),
        ));
        assert!(m.contains("CLAUDE_CONFIG_DIR"), "{bad:?}: {m}");
    }
    assert!(!project.join(".claude/settings.json").exists());
}

/// Audit N3: a variable that is not UTF-8 blocks only that host in `update-plan`.
#[test]
fn a_variable_that_is_not_utf8_blocks_only_its_host() {
    use std::os::unix::ffi::OsStrExt;
    let w = World::new();
    let pkg = w.base.join("stage");
    stage(&pkg);
    let bad = std::ffi::OsStr::from_bytes(b"/nonexistent-moai-test/conf-\xff");
    let plan = ok(&w.run_os(
        &[("CLAUDE_CONFIG_DIR", bad)],
        &["update-plan", w.h(), &s(&pkg)],
        None,
    ));
    let claude = host(&plan, "claude");
    assert!(
        claude["blocked_reason"]
            .as_str()
            .is_some_and(|r| r.contains("UTF-8")),
        "{plan}"
    );
    assert_eq!(plan["hosts"].as_array().unwrap().len(), 1, "{plan}");
}
