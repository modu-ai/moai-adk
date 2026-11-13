//! #2592 S2: base agent install into a temporary HOME — preview, apply, re-run unchanged,
//! the user's edit preserved, a foreign file never overwritten, symlinks refused, revert.
//! Temporary folders only; the real HOME is never read or written.

mod common;
use common::CleanEnv;
use serde_json::Value;
use std::{fs, path::Path, process::Command};
use tempfile::TempDir;

const EXE: &str = env!("CARGO_BIN_EXE_moai-cli");
const PERMISSION: &str = "fixture-approved";

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
    /// The agent source folder, outside the skill package (#2721).
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
    fn agent(&self, host_dir: &str, file: &str) -> std::path::PathBuf {
        Path::new(&self.home)
            .join(host_dir)
            .join("agents")
            .join(file)
    }
}

fn actions(plan: &Value, host: &str) -> Vec<(String, String)> {
    plan["hosts"]
        .as_array()
        .unwrap()
        .iter()
        .find(|h| h["host"] == host)
        .unwrap()["files"]
        .as_array()
        .unwrap()
        .iter()
        .map(|f| {
            (
                f["file"].as_str().unwrap().to_string(),
                f["action"].as_str().unwrap().to_string(),
            )
        })
        .collect()
}

fn action(plan: &Value, host: &str, file: &str) -> String {
    actions(plan, host)
        .into_iter()
        .find(|(f, _)| f == file)
        .map(|(_, a)| a)
        .unwrap_or_default()
}

/// AC-2: install into an empty HOME, then a re-run is unchanged everywhere.
#[test]
fn install_then_rerun_is_unchanged() {
    let fx = fx();
    let plan = fx.plan();
    for host in ["claude", "codex", "omp"] {
        let acts = actions(&plan, host);
        assert_eq!(acts.len(), 5, "{host}: {acts:?}");
        assert!(acts.iter().all(|(_, a)| a == "create"), "{acts:?}");
    }
    // #2721: the plan names the source folder's digest beside the skill package's.
    assert!(
        plan["agents_digest"]
            .as_str()
            .is_some_and(|d| d.starts_with("sha256:") && d.len() == 71),
        "{plan}"
    );
    assert_ne!(plan["agents_digest"], plan["package_digest"]);
    let v = fx.apply(&plan);
    assert_eq!(v["result"]["state"], "applied", "{v}");
    assert_eq!(v["result"]["written"].as_array().unwrap().len(), 15);
    // #2633: OMP reads its agents from ~/.omp/agent/agents/, with an explicit tool list.
    let omp_audit = fs::read_to_string(fx.agent(".omp/agent", "moai-audit.md")).unwrap();
    assert!(
        omp_audit.starts_with("---\nname: moai-audit\n"),
        "{omp_audit}"
    );
    assert!(omp_audit.contains("\ntools: read, grep, glob, find, bash, web_search\n"));
    let impl_md = fs::read_to_string(fx.agent(".claude", "moai-impl.md")).unwrap();
    assert!(impl_md.starts_with("---\nname: moai-impl\n"), "{impl_md}");
    let audit_toml = fs::read_to_string(fx.agent(".codex", "moai-audit.toml")).unwrap();
    assert!(
        audit_toml.starts_with("name = \"moai-audit\"\n"),
        "{audit_toml}"
    );
    let again = fx.plan();
    for host in ["claude", "codex", "omp"] {
        assert!(actions(&again, host).iter().all(|(_, a)| a == "unchanged"));
    }
    let v = fx.apply(&again);
    assert_eq!(v["result"]["state"], "unchanged", "{v}");
}

/// AC-2: the user's edit survives a re-run; when the package also changes that file it is a
/// conflict (never overwritten), while an untouched file is updated.
#[test]
fn a_user_edit_is_preserved_and_a_changed_package_conflicts_only_there() {
    let fx = fx();
    fx.apply(&fx.plan());
    let mine = fx.agent(".claude", "moai-impl.md");
    fs::write(&mine, "my own impl agent\n").unwrap();
    let plan = fx.plan();
    assert_eq!(action(&plan, "claude", "moai-impl.md"), "preserved");
    fx.apply(&plan);
    assert_eq!(fs::read_to_string(&mine).unwrap(), "my own impl agent\n");
    for role in ["impl", "audit"] {
        let src = Path::new(&fx.agents).join(format!("{role}.md"));
        let text = fs::read_to_string(&src).unwrap() + "\nOne more line.\n";
        fs::write(&src, text).unwrap();
    }
    let plan = fx.plan();
    assert_eq!(action(&plan, "claude", "moai-impl.md"), "conflict");
    assert_eq!(action(&plan, "claude", "moai-audit.md"), "update");
    let v = fx.apply(&plan);
    assert_eq!(v["result"]["state"], "applied", "{v}");
    assert_eq!(fs::read_to_string(&mine).unwrap(), "my own impl agent\n");
    assert!(fs::read_to_string(fx.agent(".claude", "moai-audit.md"))
        .unwrap()
        .contains("One more line."));
}

/// A file with our name that we never wrote is a conflict and stays; other files in the
/// folder are neither listed nor touched.
#[test]
fn a_foreign_file_is_never_overwritten_and_others_are_ignored() {
    let fx = fx();
    let dir = Path::new(&fx.home).join(".claude/agents");
    fs::create_dir_all(&dir).unwrap();
    fs::write(dir.join("moai-plan.md"), "someone else's plan agent\n").unwrap();
    fs::write(dir.join("helper.md"), "a user agent\n").unwrap();
    let plan = fx.plan();
    assert_eq!(action(&plan, "claude", "moai-plan.md"), "conflict");
    assert!(actions(&plan, "claude")
        .iter()
        .all(|(f, _)| f != "helper.md"));
    fx.apply(&plan);
    assert_eq!(
        fs::read_to_string(dir.join("moai-plan.md")).unwrap(),
        "someone else's plan agent\n"
    );
    assert_eq!(
        fs::read_to_string(dir.join("helper.md")).unwrap(),
        "a user agent\n"
    );
}

/// A symlinked agent folder blocks that host; the other host still installs.
#[cfg(unix)]
#[test]
fn a_symlinked_agent_folder_blocks_only_that_host() {
    let fx = fx();
    let outside = Path::new(&fx.home).parent().unwrap().join("outside");
    fs::create_dir_all(&outside).unwrap();
    fs::create_dir_all(Path::new(&fx.home).join(".claude")).unwrap();
    std::os::unix::fs::symlink(&outside, Path::new(&fx.home).join(".claude/agents")).unwrap();
    let plan = fx.plan();
    let claude = plan["hosts"]
        .as_array()
        .unwrap()
        .iter()
        .find(|h| h["host"] == "claude")
        .unwrap();
    assert_eq!(claude["preflight"], "blocked", "{plan}");
    let v = fx.apply(&plan);
    assert_eq!(v["result"]["written"].as_array().unwrap().len(), 10, "{v}");
    assert!(
        fs::read_dir(&outside).unwrap().next().is_none(),
        "wrote through the link"
    );
    assert!(fx.agent(".codex", "moai-impl.toml").is_file());
}

/// #2633: OMP's agent folder is two levels below `~/.omp`; a link at either level above it
/// blocks OMP alone and nothing is written through the link.
#[cfg(unix)]
#[test]
fn a_symlink_anywhere_above_the_omp_agent_folder_blocks_only_omp() {
    for linked in [".omp", ".omp/agent"] {
        let fx = fx();
        let outside = Path::new(&fx.home).parent().unwrap().join("outside");
        fs::create_dir_all(&outside).unwrap();
        let link = Path::new(&fx.home).join(linked);
        fs::create_dir_all(link.parent().unwrap()).unwrap();
        std::os::unix::fs::symlink(&outside, &link).unwrap();
        let plan = fx.plan();
        let omp = plan["hosts"]
            .as_array()
            .unwrap()
            .iter()
            .find(|h| h["host"] == "omp")
            .unwrap();
        assert_eq!(omp["preflight"], "blocked", "{linked}: {plan}");
        let v = fx.apply(&plan);
        assert_eq!(
            v["result"]["written"].as_array().unwrap().len(),
            10,
            "{linked}: {v}"
        );
        assert!(
            fs::read_dir(&outside).unwrap().next().is_none(),
            "{linked}: wrote through the link"
        );
    }
}

/// Revert removes created files that are unchanged, skips the ones the user changed, and
/// refuses a second revert; a stale plan is refused before any write.
#[test]
fn revert_undoes_only_unchanged_files_and_stale_plans_are_refused() {
    let fx = fx();
    let plan = fx.plan();
    fs::create_dir_all(Path::new(&fx.home).join(".claude/agents")).unwrap();
    fs::write(
        fx.agent(".claude", "moai-docs.md"),
        "appeared after the plan\n",
    )
    .unwrap();
    let stale = fx.apply(&plan);
    assert_eq!(stale["ok"], Value::Bool(false), "{stale}");
    assert!(!fx.agent(".codex", "moai-docs.toml").exists());
    fs::remove_file(fx.agent(".claude", "moai-docs.md")).unwrap();

    let v = fx.apply(&fx.plan());
    let id = v["result"]["receipt"].as_str().unwrap().to_string();
    fs::write(fx.agent(".codex", "moai-plan.toml"), "edited after apply\n").unwrap();
    let r = call(&[
        "agents-home-revert",
        &fx.home,
        "--receipt",
        &id,
        "--permission-reference",
        PERMISSION,
    ]);
    assert_eq!(r["result"]["state"], "reverted", "{r}");
    assert_eq!(r["result"]["reverted"].as_array().unwrap().len(), 14);
    assert_eq!(r["result"]["skipped"][0]["file"], "moai-plan.toml");
    assert!(!fx.agent(".claude", "moai-impl.md").exists());
    assert!(!fx.agent(".omp/agent", "moai-impl.md").exists());
    assert_eq!(
        fs::read_to_string(fx.agent(".codex", "moai-plan.toml")).unwrap(),
        "edited after apply\n"
    );
    let again = call(&[
        "agents-home-revert",
        &fx.home,
        "--receipt",
        &id,
        "--permission-reference",
        PERMISSION,
    ]);
    assert_eq!(again["ok"], Value::Bool(false), "{again}");
    // After the revert our state no longer owns the removed files: a new plan creates them.
    assert_eq!(action(&fx.plan(), "claude", "moai-impl.md"), "create");
}

/// Audit FAIL 6842034981 B1: after an apply the agent folder is moved outside HOME and a link
/// put in its place. Revert must not follow the link: the files outside stay, and every entry
/// of that host is reported as skipped.
#[cfg(unix)]
#[test]
fn revert_never_writes_through_a_folder_that_became_a_link() {
    let fx = fx();
    let v = fx.apply(&fx.plan());
    let id = v["result"]["receipt"].as_str().unwrap().to_string();
    let agents = Path::new(&fx.home).join(".claude/agents");
    let outside = Path::new(&fx.home).parent().unwrap().join("outside");
    fs::rename(&agents, &outside).unwrap();
    std::os::unix::fs::symlink(&outside, &agents).unwrap();
    let r = call(&[
        "agents-home-revert",
        &fx.home,
        "--receipt",
        &id,
        "--permission-reference",
        PERMISSION,
    ]);
    assert_eq!(r["ok"], Value::Bool(true), "{r}");
    let blocked = r["result"]["skipped"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|s| {
            s["host"] == "claude"
                && s["reason"]
                    .as_str()
                    .unwrap()
                    .starts_with("agent_folder_not_plain")
        })
        .count();
    assert_eq!(blocked, 5, "{r}");
    for role in ["impl", "audit", "research", "docs", "plan"] {
        assert!(
            outside.join(format!("moai-{role}.md")).is_file(),
            "{role} removed outside HOME"
        );
    }
    // The other host was not linked and is reverted as usual.
    assert!(!fx.agent(".codex", "moai-impl.toml").exists());
}
