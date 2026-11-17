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

/// REQ-1, REQ-2: the routing section goes to `$CLAUDE_CONFIG_DIR/CLAUDE.md` and
/// `$CODEX_HOME/AGENTS.md`, and revert takes it back from there; nothing under `~/.claude` or
/// `~/.codex`.
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
    assert_eq!(
        host(&plan, "codex")["file"],
        json!(s(&cx.join("AGENTS.md")))
    );
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
    assert!(cx.join("AGENTS.md").is_file());
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
    assert!(!cx.join("AGENTS.md").exists());
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
    assert_eq!(
        host(&plan, "codex")["file"],
        json!(s(&w.base.join(".codex/AGENTS.md")))
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
        assert_eq!(host(&plan, "codex")["preflight"], json!("ok"));
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
        assert_eq!(host(&plan, "codex")["preflight"], json!("ok"));
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
    // The Codex doctor refuses a relative CODEX_HOME.
    let project = w.base.join("repo");
    fs::create_dir_all(&project).unwrap();
    let out = w.run(
        &[("CODEX_HOME", "cx")],
        true,
        &["codex-doctor", &s(&project), "--home", w.h()],
    );
    assert_eq!(
        out.status.code(),
        Some(1),
        "{}",
        String::from_utf8_lossy(&out.stdout)
    );
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    let m = v["error"]["message"].as_str().unwrap();
    assert!(m.contains("CODEX_HOME") && m.contains("absolute"), "{m}");
}

/// REQ-1: agents and the output style follow the variables; OMP stays under HOME.
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
    assert_eq!(host(&plan, "codex")["dir"], json!(s(&cx.join("agents"))));
    if let Some(omp) = plan["hosts"]
        .as_array()
        .unwrap()
        .iter()
        .find(|h| h["host"] == json!("omp"))
    {
        assert_eq!(omp["dir"], json!(s(&w.base.join(".omp/agent/agents"))));
    }
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
            ("block".to_string(), s(&cx.join("AGENTS.md"))),
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

/// REQ-1: the Claude skill copy goes to `$CLAUDE_CONFIG_DIR/skills/moai`; Codex's stays in
/// `~/.agents/skills/moai` (not under `CODEX_HOME`).
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
    assert_eq!(
        host(&plan, "codex")["projection_root"],
        json!(s(&w.base.join(".agents/skills/moai")))
    );
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
}

/// REQ-1: the Codex doctor reads the global AGENTS file from `CODEX_HOME`; `--codex-home`
/// still wins.
#[test]
fn the_codex_doctor_reads_codex_home() {
    let w = World::new();
    let cx = w.outside("codex-home");
    fs::create_dir_all(&cx).unwrap();
    fs::write(cx.join("AGENTS.md"), "relocated codex notes\n").unwrap();
    let project = w.base.join("repo");
    fs::create_dir_all(&project).unwrap();
    let x = s(&cx);
    let v = ok(&w.run(
        &[("CODEX_HOME", x.as_str())],
        true,
        &["codex-doctor", &s(&project), "--home", w.h()],
    ));
    assert!(v.to_string().contains(&s(&cx.join("AGENTS.md"))), "{v}");
    let flag = w.outside("flag-home");
    fs::create_dir_all(&flag).unwrap();
    let v = ok(&w.run(
        &[("CODEX_HOME", x.as_str())],
        true,
        &[
            "codex-doctor",
            &s(&project),
            "--home",
            w.h(),
            "--codex-home",
            &s(&flag),
        ],
    ));
    assert!(!v.to_string().contains(&s(&cx)), "{v}");
}
