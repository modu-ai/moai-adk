//! SPEC-20261121-base-agents-user-folder (REQ-1 .. REQ-7, REQ-10 .. REQ-13): the five base
//! agents of Claude Code live only in `~/.claude/agents/moai/`; a flat copy from the earlier
//! layout moves there (nested copy first, flat file removed, bytes kept in the receipt) when it
//! is still ours and unedited, and stays untouched when it is edited or not ours. Every case
//! runs the built binary on a temporary HOME; the real HOME is never read or written. The last
//! case asks the installed Claude Code which agents it finds (skipped when `claude` is absent).
#![cfg(unix)]

mod common;
use common::CleanEnv;
use serde_json::Value;
use std::{
    fs,
    path::{Path, PathBuf},
    process::Command,
};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;
const PERMISSION: &str = "fixture-approved";
const BASE: [&str; 5] = ["audit", "docs", "impl", "plan", "research"];

fn call(args: &[&str]) -> Value {
    let out = Command::new(EXE)
        .args(args)
        .clean_env()
        .env("PATH", "/usr/bin:/bin")
        .env("LC_ALL", "C")
        .output()
        .unwrap();
    serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
        panic!(
            "{args:?}: {}{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    })
}

fn copy_tree(from: &Path, to: &Path) {
    fs::create_dir_all(to).unwrap();
    for e in fs::read_dir(from).unwrap() {
        let e = e.unwrap();
        let target = to.join(e.file_name());
        if e.file_type().unwrap().is_dir() {
            copy_tree(&e.path(), &target);
        } else {
            fs::copy(e.path(), target).unwrap();
        }
    }
}

struct Fx {
    _tmp: TempDir,
    home: String,
    package: String,
    agents: String,
}

fn fx() -> Fx {
    let tmp = TempDir::new().unwrap();
    let base = fs::canonicalize(tmp.path()).unwrap();
    let home = base.join("home");
    fs::create_dir_all(&home).unwrap();
    let repo = Path::new(env!("CARGO_MANIFEST_DIR"));
    let package = base.join("package");
    copy_tree(&repo.join("skills/moai"), &package);
    let agents = base.join("agents");
    copy_tree(&repo.join("agents"), &agents);
    Fx {
        _tmp: tmp,
        home: home.to_str().unwrap().into(),
        package: package.to_str().unwrap().into(),
        agents: agents.to_str().unwrap().into(),
    }
}

impl Fx {
    fn plan(&self) -> Value {
        let v = call(&["agents-home-plan", &self.home, &self.package, &self.agents]);
        assert_eq!(v["ok"], Value::Bool(true), "{v}");
        v["result"].clone()
    }
    fn apply(&self, plan: &Value) -> Value {
        call(&[
            "agents-home-apply",
            &self.home,
            &self.package,
            &self.agents,
            "--plan-digest",
            plan["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            PERMISSION,
        ])
    }
    fn revert(&self, receipt: &str) -> Value {
        call(&[
            "agents-home-revert",
            &self.home,
            "--receipt",
            receipt,
            "--permission-reference",
            PERMISSION,
        ])
    }
    fn claude_agents(&self) -> PathBuf {
        Path::new(&self.home).join(".claude/agents")
    }
    fn nested(&self, role: &str) -> PathBuf {
        self.claude_agents().join(format!("moai/moai-{role}.md"))
    }
    fn flat(&self, role: &str) -> PathBuf {
        self.claude_agents().join(format!("moai-{role}.md"))
    }
}

fn claude_files(plan: &Value) -> Vec<Value> {
    plan["hosts"]
        .as_array()
        .unwrap()
        .iter()
        .find(|h| h["host"] == "claude")
        .unwrap()["files"]
        .as_array()
        .unwrap()
        .clone()
}

fn claude_file(plan: &Value, role: &str) -> Value {
    let name = format!("moai-{role}.md");
    claude_files(plan)
        .into_iter()
        .find(|f| f["file"] == name.as_str())
        .unwrap_or_else(|| panic!("no plan line for {name}: {plan}"))
}

/// The earlier layout: the same install, then every Claude Code file moved up to `agents/`
/// (the state still records the names and digests this tool wrote, as an earlier apply left it).
fn earlier_layout(fx: &Fx) -> Vec<(String, Vec<u8>)> {
    let first = fx.apply(&fx.plan());
    assert_eq!(first["result"]["state"], "applied", "{first}");
    let mut kept = Vec::new();
    for role in BASE {
        let flat = fx.flat(role);
        let nested = fx.nested(role);
        if nested.is_file() {
            fs::rename(&nested, &flat).unwrap();
        }
        kept.push((format!("moai-{role}.md"), fs::read(&flat).unwrap()));
    }
    let moai = fx.claude_agents().join("moai");
    if moai.is_dir() {
        assert!(
            fs::read_dir(&moai).unwrap().next().is_none(),
            "the nested folder still holds files"
        );
        fs::remove_dir(&moai).unwrap();
    }
    kept
}

/// REQ-1, REQ-2, REQ-7: the five base agents land only in `agents/moai/`, keep their names and
/// leave the user's own agents as they are.
#[test]
fn base_agents_install_only_inside_the_moai_folder() {
    let fx = fx();
    fs::create_dir_all(fx.claude_agents().join("moai")).unwrap();
    fs::write(fx.claude_agents().join("mine.md"), "my agent\n").unwrap();
    fs::write(fx.claude_agents().join("moai/mine.md"), "my other agent\n").unwrap();
    let plan = fx.plan();
    let v = fx.apply(&plan);
    assert_eq!(v["result"]["state"], "applied", "{v}");

    let mut top: Vec<String> = fs::read_dir(fx.claude_agents())
        .unwrap()
        .map(|e| e.unwrap().file_name().to_string_lossy().into_owned())
        .collect();
    top.sort();
    assert_eq!(top, ["mine.md", "moai"], "the top level of agents/");
    let mut inside: Vec<String> = fs::read_dir(fx.claude_agents().join("moai"))
        .unwrap()
        .map(|e| e.unwrap().file_name().to_string_lossy().into_owned())
        .collect();
    inside.sort();
    assert_eq!(
        inside,
        [
            "mine.md",
            "moai-audit.md",
            "moai-docs.md",
            "moai-impl.md",
            "moai-plan.md",
            "moai-research.md"
        ]
    );
    for role in BASE {
        let text = fs::read_to_string(fx.nested(role)).unwrap();
        assert!(
            text.contains(&format!("\nname: moai-{role}\n")),
            "moai-{role}: {text}"
        );
    }
    assert_eq!(
        fs::read(fx.claude_agents().join("mine.md")).unwrap(),
        b"my agent\n"
    );
    assert_eq!(
        fs::read(fx.claude_agents().join("moai/mine.md")).unwrap(),
        b"my other agent\n"
    );
}

/// REQ-3, REQ-13, REQ-6: a flat set this tool wrote and nobody edited is moved: the preview
/// shows the removal, the nested copy is written, the flat file is gone, a second run is
/// unchanged, and the user's own `agents/mine.md` is still there.
#[test]
fn an_unedited_flat_install_of_the_earlier_layout_moves_into_the_moai_folder() {
    let fx = fx();
    let kept = earlier_layout(&fx);
    fs::write(fx.claude_agents().join("mine.md"), "my agent\n").unwrap();

    let plan = fx.plan();
    for role in BASE {
        let line = claude_file(&plan, role);
        assert_eq!(line["legacy"], "remove", "{role}: {line}");
        assert_eq!(line["action"], "create", "{role}: {line}");
    }
    let v = fx.apply(&plan);
    assert_eq!(v["result"]["state"], "applied", "{v}");
    for (name, bytes) in &kept {
        let nested = fx.claude_agents().join("moai").join(name);
        assert_eq!(&fs::read(&nested).unwrap(), bytes, "{name}: nested copy");
        assert!(!fx.claude_agents().join(name).exists(), "{name}: flat file");
    }
    assert_eq!(
        fs::read(fx.claude_agents().join("mine.md")).unwrap(),
        b"my agent\n"
    );
    let again = fx.apply(&fx.plan());
    assert_eq!(again["result"]["state"], "unchanged", "{again}");
}

/// REQ-4: a flat file that was edited, or that this tool never wrote, stays as it is, its
/// nested copy is not written, and the plan lists it as a conflict naming the flat file.
#[test]
fn an_edited_or_foreign_flat_file_stays_and_blocks_its_nested_copy() {
    let fx = fx();
    earlier_layout(&fx);
    fs::write(fx.flat("impl"), "I edited this one\n").unwrap();
    // A flat file the state does not own: forget the record of one agent.
    let state = [".moai/agents/claude.json", ".moai-cli/agents/claude.json"]
        .iter()
        .map(|rel| Path::new(&fx.home).join(rel))
        .find(|p| p.is_file())
        .expect("the agent state file of the claude host");
    let mut json: Value = serde_json::from_slice(&fs::read(&state).unwrap()).unwrap();
    json["owned"].as_object_mut().unwrap().remove("moai-plan.md");
    fs::write(&state, serde_json::to_vec_pretty(&json).unwrap()).unwrap();
    fs::write(fx.flat("plan"), "someone else's plan agent\n").unwrap();

    let plan = fx.plan();
    let edited = claude_file(&plan, "impl");
    assert_eq!(edited["action"], "conflict", "{edited}");
    assert_eq!(edited["legacy"], "kept_edited", "{edited}");
    let foreign = claude_file(&plan, "plan");
    assert_eq!(foreign["action"], "conflict", "{foreign}");
    assert_eq!(foreign["legacy"], "kept_foreign", "{foreign}");
    // The others move as usual.
    assert_eq!(claude_file(&plan, "audit")["legacy"], "remove");

    fx.apply(&plan);
    assert_eq!(fs::read(fx.flat("impl")).unwrap(), b"I edited this one\n");
    assert_eq!(
        fs::read(fx.flat("plan")).unwrap(),
        b"someone else's plan agent\n"
    );
    assert!(!fx.nested("impl").exists(), "nested copy of the edited file");
    assert!(!fx.nested("plan").exists(), "nested copy of the foreign file");
    assert!(fx.nested("audit").is_file());
    assert!(!fx.flat("audit").exists());
}

/// REQ-5, REQ-12: the revert of a move puts the flat files back with their bytes and removes
/// the nested copies; a nested file changed since the apply is skipped and reported.
#[test]
fn the_revert_of_a_move_restores_the_flat_files_and_skips_a_changed_one() {
    let fx = fx();
    let kept = earlier_layout(&fx);
    let v = fx.apply(&fx.plan());
    let id = v["result"]["receipt"].as_str().unwrap().to_string();
    fs::write(fx.nested("docs"), "edited after the move\n").unwrap();

    let r = fx.revert(&id);
    assert_eq!(r["result"]["state"], "reverted", "{r}");
    let skipped = r["result"]["skipped"].to_string();
    assert!(skipped.contains("moai-docs.md"), "{skipped}");
    for (name, bytes) in &kept {
        if name == "moai-docs.md" {
            continue;
        }
        assert_eq!(
            &fs::read(fx.claude_agents().join(name)).unwrap(),
            bytes,
            "{name}: flat file"
        );
        assert!(
            !fx.claude_agents().join("moai").join(name).exists(),
            "{name}: nested copy"
        );
    }
    assert_eq!(
        fs::read(fx.nested("docs")).unwrap(),
        b"edited after the move\n"
    );
}

/// REQ-10: Claude Code itself finds the five base agents and a project role. There is no
/// documented command that lists definitions without a model run; an unknown `--agent` name
/// exits before any model call and prints the list it found (measured on Claude Code 2.1.330).
#[test]
fn claude_code_finds_both_kinds() {
    let Some(claude) = std::env::var_os("PATH").and_then(|paths| {
        std::env::split_paths(&paths)
            .map(|d| d.join("claude"))
            .find(|p| p.is_file())
    }) else {
        eprintln!("skipped: `claude` is not on the PATH");
        return;
    };
    let fx = fx();
    let v = fx.apply(&fx.plan());
    assert_eq!(v["result"]["state"], "applied", "{v}");
    for role in BASE {
        assert!(fx.nested(role).is_file(), "moai-{role} is not nested");
        assert!(!fx.flat(role).exists(), "moai-{role} is flat too");
    }
    let project = Path::new(&fx.home).join("project");
    fs::create_dir_all(project.join(".claude/agents/moai")).unwrap();
    fs::write(
        project.join(".claude/agents/moai/prj-probe.md"),
        "---\nname: prj-probe\ndescription: probe role of the project\ntools: Read\n---\n\nProbe.\n",
    )
    .unwrap();
    let out = Command::new(&claude)
        .clean_env()
        .args(["--agent", "no-such-agent-for-the-list", "-p", "hi"])
        .current_dir(&project)
        .env("PATH", std::env::var_os("PATH").unwrap())
        .env("HOME", &fx.home)
        .env("CLAUDE_CONFIG_DIR", Path::new(&fx.home).join(".claude"))
        .output()
        .unwrap();
    let text = format!(
        "{}{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    assert!(text.contains("Available agents:"), "{text}");
    for role in BASE {
        assert!(text.contains(&format!("moai-{role}")), "moai-{role}: {text}");
    }
    assert!(text.contains("prj-probe"), "{text}");
}
