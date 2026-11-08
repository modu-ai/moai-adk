//! Codex adapter fixtures (02-C #1993, Issue body r30). Every tree is built in a TempDir at
//! run time; nothing touches the real home, `~/.codex` or `~/.agents`.
use serde_json::Value;
use std::{fs, path::Path, process::Command};
use tempfile::TempDir;

fn run(args: &[&str]) -> (bool, Value) {
    let out = Command::new(env!("CARGO_BIN_EXE_moai-cli"))
        .args(args)
        .env_remove("CODEX_HOME")
        .output()
        .unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap_or(Value::Null);
    (out.status.success(), v)
}
fn write(path: &Path, text: &str) {
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, text).unwrap();
}
fn s(p: &Path) -> &str {
    p.to_str().unwrap()
}
fn package() -> String {
    Path::new(env!("CARGO_MANIFEST_DIR"))
        .join("skills/moai")
        .to_str()
        .unwrap()
        .to_string()
}

/// A project root, a nested working directory, an empty Codex home and an empty user home.
struct Fixture {
    _dir: TempDir,
    root: std::path::PathBuf,
    cwd: std::path::PathBuf,
    codex_home: std::path::PathBuf,
    home: std::path::PathBuf,
}
fn fixture() -> Fixture {
    let dir = TempDir::new().unwrap();
    let root = dir.path().join("repo");
    let cwd = root.join("a/b");
    let codex_home = dir.path().join("codex-home");
    let home = dir.path().join("home");
    for d in [&cwd, &codex_home, &home] {
        fs::create_dir_all(d).unwrap();
    }
    Fixture {
        root: root.canonicalize().unwrap(),
        cwd: cwd.canonicalize().unwrap(),
        codex_home,
        home,
        _dir: dir,
    }
}
fn doctor(f: &Fixture, extra: &[&str]) -> (bool, Value) {
    let mut args = vec![
        "codex-doctor",
        s(&f.root),
        "--cwd",
        s(&f.cwd),
        "--codex-home",
        s(&f.codex_home),
        "--home",
        s(&f.home),
    ];
    args.extend_from_slice(extra);
    run(&args)
}
fn included(v: &Value) -> Vec<String> {
    v["result"]["agents"]["included"]
        .as_array()
        .unwrap()
        .iter()
        .map(|e| e["path"].as_str().unwrap().to_string())
        .collect()
}

#[test]
fn codex_agents_chain_plain() {
    let f = fixture();
    write(&f.codex_home.join("AGENTS.md"), "global\n");
    write(&f.root.join("AGENTS.md"), "root\n");
    write(&f.root.join("a/b/AGENTS.md"), "leaf\n");
    let (ok, v) = doctor(&f, &[]);
    assert!(ok, "{v}");
    assert_eq!(v["operation"], "codex.doctor");
    let r = &v["result"]["agents"];
    assert_eq!(
        r["global"]["included"].as_str().unwrap(),
        s(&f.codex_home.join("AGENTS.md"))
    );
    // Root first, then down to the working directory; `a/` has no file.
    assert_eq!(
        included(&v),
        vec![
            s(&f.root.join("AGENTS.md")).to_string(),
            s(&f.root.join("a/b/AGENTS.md")).to_string()
        ]
    );
    assert_eq!(r["project_chain"].as_array().unwrap().len(), 3);
    assert_eq!(r["project_bytes"], 5 + 2 + 5);
    assert_eq!(r["max_bytes"], 32768);
    assert_eq!(r["over_budget"], false);
    assert_eq!(r["agents_override_created"], false);
}

#[test]
fn codex_agents_chain_override() {
    let f = fixture();
    // An empty global override is skipped: the first non-empty file wins.
    write(&f.codex_home.join("AGENTS.override.md"), "  \n");
    write(&f.codex_home.join("AGENTS.md"), "global\n");
    write(&f.root.join("AGENTS.override.md"), "override\n");
    write(&f.root.join("AGENTS.md"), "shadowed\n");
    let (ok, v) = doctor(&f, &[]);
    assert!(ok, "{v}");
    assert_eq!(
        v["result"]["agents"]["global"]["included"]
            .as_str()
            .unwrap(),
        s(&f.codex_home.join("AGENTS.md"))
    );
    // At most one file per directory, the override first.
    assert_eq!(
        included(&v),
        vec![s(&f.root.join("AGENTS.override.md")).to_string()]
    );
    // The doctor never creates an override file.
    assert!(!f.cwd.join("AGENTS.override.md").exists());
    assert!(!f.home.join("AGENTS.override.md").exists());
}

#[test]
fn codex_agents_chain_skips_empty_files() {
    // "Codex skips empty files" (AGENTS.md page). Falling through to the next name in the
    // same directory is inferred; the page does not say it for project directories.
    let f = fixture();
    write(&f.root.join("AGENTS.override.md"), "\n  \n");
    write(&f.root.join("AGENTS.md"), "root\n");
    write(&f.root.join("a/AGENTS.md"), "");
    let (ok, v) = doctor(&f, &[]);
    assert!(ok, "{v}");
    assert_eq!(included(&v), vec![s(&f.root.join("AGENTS.md")).to_string()]);
    assert_eq!(v["result"]["agents"]["project_bytes"], 5);
    assert!(v["result"]["agents"]["project_chain"][1]["included"].is_null());
}

#[test]
fn codex_agents_chain_fallback() {
    let f = fixture();
    write(&f.root.join("TEAM.md"), "team\n");
    let (_, v) = doctor(&f, &[]);
    assert_eq!(
        included(&v),
        Vec::<String>::new(),
        "no fallback without the argument"
    );
    assert!(v["result"]["agents"]["fallback_source"]
        .as_str()
        .unwrap()
        .contains("TOML"));
    let (ok, v) = doctor(&f, &["--fallback", "TEAM.md"]);
    assert!(ok, "{v}");
    assert_eq!(included(&v), vec![s(&f.root.join("TEAM.md")).to_string()]);
    assert_eq!(v["result"]["agents"]["fallback_source"], "argument");
    // AGENTS.md still wins over a fallback name in the same directory.
    write(&f.root.join("AGENTS.md"), "agents\n");
    let (_, v) = doctor(&f, &["--fallback", "TEAM.md"]);
    assert_eq!(included(&v), vec![s(&f.root.join("AGENTS.md")).to_string()]);
    // A fallback name is a plain file name.
    let (ok, v) = doctor(&f, &["--fallback", "../x.md"]);
    assert!(!ok);
    assert_eq!(v["error"]["kind"], "invalid_input");
}

#[test]
fn codex_agents_chain_over_budget() {
    let f = fixture();
    write(&f.root.join("AGENTS.md"), &"r".repeat(30));
    write(&f.root.join("a/AGENTS.md"), &"a".repeat(30));
    let (ok, v) = doctor(&f, &["--max-bytes", "40"]);
    assert!(ok, "{v}");
    let r = &v["result"]["agents"];
    assert_eq!(r["max_bytes"], 40);
    assert_eq!(r["max_bytes_source"], "argument");
    assert_eq!(r["over_budget"], true);
    assert_eq!(
        r["over_budget_at"].as_str().unwrap(),
        s(&f.root.join("a/AGENTS.md"))
    );
    let (_, v) = doctor(&f, &[]);
    assert_eq!(v["result"]["agents"]["over_budget"], false);
    let (ok, v) = doctor(&f, &["--max-bytes", "0"]);
    assert!(!ok);
    assert_eq!(v["error"]["kind"], "invalid_input");
}

#[test]
fn codex_doctor_never_claims_trust() {
    let f = fixture();
    let config = "[projects.\"/x\"]\ntrust_level = \"trusted\"\n[hooks]\n# user\n";
    write(&f.root.join(".codex/config.toml"), config);
    write(&f.root.join(".codex/hooks.json"), "{\"hooks\":{}}\n");
    write(&f.codex_home.join("config.toml"), config);
    let (ok, v) = doctor(&f, &[]);
    assert!(ok, "{v}");
    let r = &v["result"];
    // Even a config that names the project trusted is not read as trust.
    assert_eq!(r["trust"]["status"], "not_evaluated");
    assert!(!r.to_string().contains("\"trusted\""));
    assert!(r["trust"]["reason"].as_str().unwrap().contains("02-B"));
    assert_eq!(r["codex_layer"]["config_toml"], true);
    assert_eq!(r["codex_layer"]["hooks_json"], true);
    assert_eq!(r["writes"], serde_json::json!([]));
    assert_eq!(
        fs::read_to_string(f.root.join(".codex/config.toml")).unwrap(),
        config
    );
    assert_eq!(
        fs::read_to_string(f.codex_home.join("config.toml")).unwrap(),
        config
    );
}

#[test]
fn codex_config_requests_report_hold_without_writes() {
    let f = fixture();
    let target = f.home.join(".agents/skills/moai");
    let (ok, v) = run(&[
        "codex-projection-plan",
        &package(),
        "--target",
        s(&target),
        "--disable-bundled-skills",
        "--status-line",
        "model",
    ]);
    assert!(ok, "{v}");
    let c = &v["result"]["config_changes"];
    assert_eq!(c["status"], "hold");
    assert_eq!(c["requested"]["skills.bundled.enabled"], false);
    assert_eq!(c["writes"], serde_json::json!([]));
    assert!(!target.exists(), "plan only");
    assert!(!f.home.join(".codex").exists());
    let (_, v) = run(&["codex-projection-plan", &package(), "--target", s(&target)]);
    assert_eq!(v["result"]["config_changes"]["status"], "none_requested");
}

#[test]
fn codex_statusline_items_are_unverified() {
    let f = fixture();
    let target = f.home.join(".agents/skills/moai");
    let (ok, v) = run(&[
        "codex-projection-plan",
        &package(),
        "--target",
        s(&target),
        "--status-line",
        "model",
        "--status-line",
        "git-branch",
    ]);
    assert!(ok, "{v}");
    let c = &v["result"]["config_changes"];
    assert_eq!(
        c["unverified_items"],
        serde_json::json!(["model", "git-branch"])
    );
    assert!(!v["result"]["files"].to_string().contains("config.toml"));
    let (_, d) = doctor(&f, &[]);
    assert_eq!(d["result"]["codex_layer"]["status_line"], "not_evaluated");
}

#[test]
fn codex_hooks_not_generated() {
    let f = fixture();
    let target = f.home.join(".agents/skills/moai");
    let (_, v) = run(&["codex-projection-plan", &package(), "--target", s(&target)]);
    assert_eq!(v["result"]["hooks"]["generated"], serde_json::json!([]));
    assert!(!v["result"]["files"].to_string().contains("hooks"));
    let (_, d) = doctor(&f, &[]);
    assert_eq!(d["result"]["codex_layer"]["inline_hooks"], "not_evaluated");
    assert_eq!(d["result"]["codex_layer"]["hooks_json"], false);
}

#[test]
fn codex_openai_yaml_only_explicit_fields() {
    let f = fixture();
    let target = f.home.join(".agents/skills/moai");
    let pkg = package();
    let yaml = |extra: &[&str]| {
        let mut args = vec![
            "codex-projection-plan",
            pkg.as_str(),
            "--target",
            s(&target),
        ];
        args.extend_from_slice(extra);
        let (ok, v) = run(&args);
        assert!(ok, "{v}");
        let files = v["result"]["files"].as_array().unwrap().clone();
        files
            .iter()
            .find(|e| e["path"] == "agents/openai.yaml")
            .unwrap()["content"]
            .as_str()
            .unwrap()
            .to_string()
    };
    let plain = yaml(&[]);
    let keys: Vec<&str> = plain
        .lines()
        .map(|l| l.split(':').next().unwrap().trim())
        .collect();
    assert_eq!(
        keys,
        [
            "interface",
            "display_name",
            "short_description",
            "default_prompt"
        ]
    );
    assert!(
        !plain.contains("allow_implicit_invocation"),
        "default kept by omission"
    );
    let off = yaml(&["--allow-implicit-invocation", "false"]);
    assert!(off.ends_with("policy:\n  allow_implicit_invocation: false\n"));
    // Quotes are escaped, control characters refused.
    let quoted = yaml(&["--display-name", "Mo\"AI"]);
    assert!(quoted.contains(r#"display_name: "Mo\"AI""#));
    let (ok, v) = run(&[
        "codex-projection-plan",
        &package(),
        "--target",
        s(&target),
        "--display-name",
        "a\nb",
    ]);
    assert!(!ok);
    assert_eq!(v["error"]["kind"], "invalid_input");
}

#[test]
fn codex_projection_plan_writes_nothing() {
    let f = fixture();
    // `--target` is required: no default projection location is chosen (HOLD).
    let (ok, v) = run(&["codex-projection-plan", &package()]);
    assert!(!ok);
    assert_eq!(v["error"]["kind"], "usage");
    let target = f.home.join(".agents/skills/moai");
    let (ok, v) = run(&["codex-projection-plan", &package(), "--target", s(&target)]);
    assert!(ok, "{v}");
    let r = &v["result"];
    assert_eq!(r["files_created"], serde_json::json!([]));
    assert_eq!(r["host_loaded"], false);
    assert_eq!(r["execution_verified"], false);
    // Every package file is a byte-identical copy (REQ-02C-01).
    let pkg = Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai");
    let copies: Vec<&Value> = r["files"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|e| e["action"] == "copy")
        .collect();
    assert!(copies.iter().any(|e| e["path"] == "SKILL.md"));
    for e in &copies {
        let bytes = fs::read(pkg.join(e["path"].as_str().unwrap())).unwrap();
        assert_eq!(e["sha256"].as_str().unwrap(), moai_cli::digest(&bytes));
    }
    assert!(!f.home.join(".agents").exists());
    // Relative targets, targets inside the package and non-`moai` names are refused.
    for bad in ["rel/moai", "/tmp/x/other"] {
        let (ok, v) = run(&["codex-projection-plan", &package(), "--target", bad]);
        assert!(!ok, "{bad}");
        assert_eq!(v["error"]["kind"], "invalid_input");
    }
    let inside = pkg.join("moai");
    let (ok, _) = run(&["codex-projection-plan", &package(), "--target", s(&inside)]);
    assert!(!ok);
}

#[test]
fn codex_messages_have_a_deliberate_kind() {
    use moai_cli::contract::{classify, Kind::*};
    for (message, kind) in [
        ("invalid target: must be an absolute path", InvalidInput),
        ("invalid target: must be outside the package", InvalidInput),
        (
            "invalid target: the skill directory must be named moai",
            InvalidInput,
        ),
        (
            "invalid package: agents/ belongs to the Codex projection, not the package",
            InvalidInput,
        ),
        (
            "invalid display_name: 1-256 characters without control characters",
            InvalidInput,
        ),
        ("invalid cwd: must be inside the project", InvalidInput),
        (
            "invalid fallback name: a plain file name is required",
            InvalidInput,
        ),
        ("invalid max-bytes: must be positive", InvalidInput),
        ("home directory missing; pass --home", InvalidInput),
        ("package read failed", Io),
        ("project directory unavailable", NotFound),
        ("path unavailable", NotFound),
    ] {
        assert_eq!(classify(message), kind, "{message}");
    }
}

#[test]
fn codex_duplicate_moai_skill_is_warned() {
    let f = fixture();
    let (_, v) = doctor(&f, &[]);
    assert!(v["result"]["skills"]["duplicate_warning"].is_null());
    write(
        &f.root.join("a/.agents/skills/moai/SKILL.md"),
        "---\nname: moai\n---\n",
    );
    write(
        &f.home.join(".agents/skills/moai/SKILL.md"),
        "---\nname: moai\n---\n",
    );
    // A folder named moai outside a skills root is not a skill.
    write(&f.root.join("moai/SKILL.md"), "x");
    let (ok, v) = doctor(&f, &[]);
    assert!(ok, "{v}");
    let found = v["result"]["skills"]["moai"].as_array().unwrap();
    let scopes: Vec<&str> = found.iter().map(|e| e["scope"].as_str().unwrap()).collect();
    assert_eq!(scopes, ["repository", "user"]);
    assert!(v["result"]["skills"]["duplicate_warning"]
        .as_str()
        .unwrap()
        .contains("does not merge"));
}
