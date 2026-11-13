//! Codex adapter fixtures (02-C #1993, Issue body r30). Every tree is built in a TempDir at
//! run time; nothing touches the real home, `~/.codex` or `~/.agents`.
mod common;
use serde_json::Value;
use std::{fs, path::Path, process::Command};
use tempfile::TempDir;

fn run(args: &[&str]) -> (bool, Value) {
    let out = Command::new(crate::common::MOAI_BIN)
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
fn codex_projection_requires_host_neutral_frontmatter() {
    // REQ-02C-01: the projected SKILL.md is the package's bytes, with only name and description.
    let f = fixture();
    let target = f.home.join(".agents/skills/moai");
    let (ok, v) = run(&["codex-projection-plan", &package(), "--target", s(&target)]);
    assert!(ok, "{v}");
    assert_eq!(v["result"]["skill_md_identical"], true);
    let bad = TempDir::new().unwrap();
    for (front, label) in [
        (
            "name: moai\ndescription: d\nallowed-tools: Bash\n",
            "extra key",
        ),
        ("name: other\ndescription: d\n", "wrong name"),
        ("name: moai\n", "no description"),
    ] {
        write(
            &bad.path().join("SKILL.md"),
            &format!("---\n{front}---\n\nbody\n"),
        );
        let (ok, v) = run(&[
            "codex-projection-plan",
            s(bad.path()),
            "--target",
            s(&target),
        ]);
        assert!(!ok, "{label}");
        assert_eq!(v["error"]["kind"], "invalid_input", "{label}");
    }
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
        ("invalid package: SKILL.md is not UTF-8", InvalidInput),
        (
            "invalid package: SKILL.md frontmatter missing",
            InvalidInput,
        ),
        (
            "invalid package: SKILL.md frontmatter line without a key",
            InvalidInput,
        ),
        (
            "invalid package: SKILL.md frontmatter must hold only name and description",
            InvalidInput,
        ),
        ("invalid package: SKILL.md name must be moai", InvalidInput),
        ("package SKILL.md missing", NotFound),
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

// --- I4 #2241: projection apply ------------------------------------------------------
// Both install locations are fixtures only (the choice is a P04 / C5 HOLD): a home-shaped
// `<tmp>/home/.agents/skills/moai` and a project-shaped `<tmp>/repo/.agents/skills/moai`.

fn plan(package: &str, target: &Path) -> Value {
    let (ok, v) = run(&["codex-projection-plan", package, "--target", s(target)]);
    assert!(ok, "{v}");
    v["result"].clone()
}
fn apply(package: &str, target: &Path, digest: &str, permission: &str) -> (bool, Value) {
    run(&[
        "codex-projection-apply",
        package,
        "--target",
        s(target),
        "--plan-digest",
        digest,
        "--permission-reference",
        permission,
    ])
}
fn actions(plan: &Value) -> Vec<String> {
    plan["files"]
        .as_array()
        .unwrap()
        .iter()
        .map(|e| e["target_action"].as_str().unwrap().to_string())
        .collect()
}
/// Path → (bytes, modified time) of every file under `dir`; equal maps mean no write.
fn snapshot(dir: &Path) -> std::collections::BTreeMap<String, (Vec<u8>, std::time::SystemTime)> {
    let mut out = std::collections::BTreeMap::new();
    let mut stack = vec![dir.to_path_buf()];
    while let Some(d) = stack.pop() {
        for e in fs::read_dir(&d).unwrap() {
            let p = e.unwrap().path();
            if p.is_dir() {
                stack.push(p);
            } else {
                let rel = p.strip_prefix(dir).unwrap().to_str().unwrap().to_string();
                let modified = fs::metadata(&p).unwrap().modified().unwrap();
                out.insert(rel, (fs::read(&p).unwrap(), modified));
            }
        }
    }
    out
}
fn copy_tree(from: &Path, to: &Path) {
    for e in fs::read_dir(from).unwrap() {
        let p = e.unwrap().path();
        let dest = to.join(p.file_name().unwrap());
        if p.is_dir() {
            fs::create_dir_all(&dest).unwrap();
            copy_tree(&p, &dest);
        } else {
            fs::create_dir_all(to).unwrap();
            fs::copy(&p, &dest).unwrap();
        }
    }
}
const PERMISSION: &str = "test-approval-2241";

#[test]
fn codex_projection_apply_installs_verifies_and_reinstalls_as_noop() {
    let f = fixture();
    for target in [
        f.home.join(".agents/skills/moai"),
        f.root.join(".agents/skills/moai"),
    ] {
        let first = plan(&package(), &target);
        assert_eq!(first["source_manifest"]["status"], "verified", "{first}");
        assert!(actions(&first).iter().all(|a| a == "create"));
        let total = first["files"].as_array().unwrap().len();
        let (ok, v) = apply(
            &package(),
            &target,
            first["plan_digest"].as_str().unwrap(),
            PERMISSION,
        );
        assert!(ok, "{v}");
        assert_eq!(v["operation"], "codex.projection_apply");
        let r = &v["result"];
        assert_eq!(r["files_created"].as_array().unwrap().len(), total);
        assert_eq!(r["verified"], true);
        assert_eq!(r["host_loaded"], false);
        // The projection is the package's bytes plus the adapter's agents/openai.yaml.
        let pkg = Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai");
        assert_eq!(
            fs::read(target.join("SKILL.md")).unwrap(),
            fs::read(pkg.join("SKILL.md")).unwrap()
        );
        let yaml = first["files"]
            .as_array()
            .unwrap()
            .iter()
            .find(|e| e["path"] == "agents/openai.yaml")
            .unwrap();
        assert_eq!(
            fs::read_to_string(target.join("agents/openai.yaml")).unwrap(),
            yaml["content"].as_str().unwrap()
        );
        // Verification: a fresh plan sees every file unchanged; the first digest is now stale.
        let second = plan(&package(), &target);
        assert!(
            actions(&second).iter().all(|a| a == "unchanged"),
            "{second}"
        );
        assert_ne!(second["plan_digest"], first["plan_digest"]);
        let (ok, v) = apply(
            &package(),
            &target,
            first["plan_digest"].as_str().unwrap(),
            PERMISSION,
        );
        assert!(!ok);
        assert_eq!(v["error"]["kind"], "stale_plan");
        // Re-install: no byte and no modification time changes.
        let before = snapshot(&target);
        let (ok, v) = apply(
            &package(),
            &target,
            second["plan_digest"].as_str().unwrap(),
            PERMISSION,
        );
        assert!(ok, "{v}");
        assert_eq!(v["result"]["files_created"], serde_json::json!([]));
        assert_eq!(
            v["result"]["files_unchanged"].as_array().unwrap().len(),
            total
        );
        assert_eq!(snapshot(&target), before);
    }
    // Each target received only its own install; the other location stays as the test made it.
    assert!(f.home.join(".agents/skills/moai/SKILL.md").is_file());
    assert!(f.root.join(".agents/skills/moai/SKILL.md").is_file());
}

#[test]
fn codex_projection_apply_refuses_a_manifest_mismatch_and_writes_nothing() {
    // REQ-I4-01: a copy whose digest differs from source-manifest.json is never written.
    let f = fixture();
    let pkg = Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai");
    type Mutate = fn(&Path);
    let cases: [(&str, Mutate, &str); 4] = [
        (
            "edited file",
            |p| fs::write(p.join("references/routes.json"), "{}\n").unwrap(),
            "references/routes.json",
        ),
        (
            "unlisted file",
            |p| fs::write(p.join("workflows/extra.md"), "x\n").unwrap(),
            "workflows/extra.md",
        ),
        (
            "listed but absent",
            |p| fs::remove_file(p.join("references/routes.json")).unwrap(),
            "references/routes.json",
        ),
        (
            "manifest absent",
            |p| fs::remove_file(p.join("references/source-manifest.json")).unwrap(),
            "missing",
        ),
    ];
    for (label, mutate, needle) in cases {
        let copy = TempDir::new().unwrap();
        let package = copy.path().join("moai");
        copy_tree(&pkg, &package);
        mutate(&package);
        let target = f.home.join(".agents/skills/moai");
        let p = plan(s(&package), &target);
        assert_ne!(p["source_manifest"]["status"], "verified", "{label}");
        assert!(p.to_string().contains(needle), "{label}: {p}");
        let (ok, v) = apply(
            s(&package),
            &target,
            p["plan_digest"].as_str().unwrap(),
            PERMISSION,
        );
        assert!(!ok, "{label}");
        assert_eq!(v["error"]["kind"], "conflict", "{label}: {v}");
        assert!(
            v["error"]["message"]
                .as_str()
                .unwrap()
                .contains("nothing applied"),
            "{label}"
        );
        assert!(!target.exists(), "{label}: something was written");
    }
    // The unmodified copy of the same package installs: the refusal came from the mismatch.
    let copy = TempDir::new().unwrap();
    let package = copy.path().join("moai");
    copy_tree(&pkg, &package);
    let target = f.root.join(".agents/skills/moai");
    let p = plan(s(&package), &target);
    let (ok, v) = apply(
        s(&package),
        &target,
        p["plan_digest"].as_str().unwrap(),
        PERMISSION,
    );
    assert!(ok, "{v}");
}

#[test]
fn codex_projection_apply_never_overwrites_and_keeps_unowned_files() {
    let f = fixture();
    let target = f.home.join(".agents/skills/moai");
    // A user's own file at the target is preserved and listed, never touched.
    write(&target.join("NOTES.md"), "mine\n");
    let p = plan(&package(), &target);
    assert_eq!(p["preserved"], serde_json::json!(["NOTES.md"]));
    let (ok, v) = apply(
        &package(),
        &target,
        p["plan_digest"].as_str().unwrap(),
        PERMISSION,
    );
    assert!(ok, "{v}");
    assert_eq!(
        fs::read_to_string(target.join("NOTES.md")).unwrap(),
        "mine\n"
    );
    // A planned path holding different bytes refuses the whole apply before any write.
    let other = f.root.join(".agents/skills/moai");
    write(
        &other.join("SKILL.md"),
        "---\nname: moai\ndescription: edited\n---\n",
    );
    let p = plan(&package(), &other);
    let skill = p["files"]
        .as_array()
        .unwrap()
        .iter()
        .find(|e| e["path"] == "SKILL.md")
        .unwrap();
    assert_eq!(skill["target_action"], "conflict");
    let (ok, v) = apply(
        &package(),
        &other,
        p["plan_digest"].as_str().unwrap(),
        PERMISSION,
    );
    assert!(!ok);
    assert_eq!(v["error"]["kind"], "conflict", "{v}");
    assert!(!other.join("agents/openai.yaml").exists());
    assert!(!other.join("references").exists());
    assert_eq!(
        fs::read_to_string(other.join("SKILL.md")).unwrap(),
        "---\nname: moai\ndescription: edited\n---\n"
    );
}

#[test]
fn codex_projection_apply_gates() {
    let f = fixture();
    let target = f.home.join(".agents/skills/moai");
    let p = plan(&package(), &target);
    let digest = p["plan_digest"].as_str().unwrap();
    for bad in ["", "placeholder", "null"] {
        let (ok, v) = apply(&package(), &target, digest, bad);
        assert!(!ok);
        assert_eq!(
            v["error"]["kind"], "permission_reference_missing",
            "{bad:?}"
        );
    }
    let (ok, v) = apply(&package(), &target, "sha256:0000", PERMISSION);
    assert!(!ok);
    assert_eq!(v["error"]["kind"], "stale_plan");
    // Changing the openai.yaml choice changes the plan, so the old digest is stale.
    let (ok, v) = run(&[
        "codex-projection-apply",
        &package(),
        "--target",
        s(&target),
        "--display-name",
        "Other",
        "--plan-digest",
        digest,
        "--permission-reference",
        PERMISSION,
    ]);
    assert!(!ok);
    assert_eq!(v["error"]["kind"], "stale_plan");
    assert!(!target.exists());
    // A target that is a file, or holds a symlink, is refused.
    let file_target = f.root.join("x/moai");
    write(&file_target, "x");
    let (ok, v) = run(&[
        "codex-projection-plan",
        &package(),
        "--target",
        s(&file_target),
    ]);
    assert!(!ok);
    assert_eq!(v["error"]["kind"], "invalid_input");
    #[cfg(unix)]
    {
        let linked = f.root.join("y/moai");
        fs::create_dir_all(&linked).unwrap();
        std::os::unix::fs::symlink(&f.home, linked.join("escape")).unwrap();
        let (ok, v) = run(&["codex-projection-plan", &package(), "--target", s(&linked)]);
        assert!(!ok);
        assert_eq!(v["error"]["kind"], "invalid_input", "{v}");
    }
}

#[test]
fn codex_projection_apply_messages_have_a_deliberate_kind() {
    use moai_cli::contract::{classify, Kind::*};
    for (message, kind) in [
        (
            "source manifest digest mismatch (mismatch): references/routes.json; nothing applied",
            Conflict,
        ),
        (
            "source manifest digest mismatch (missing): ; nothing applied",
            Conflict,
        ),
        (
            "conflict in codex projection: SKILL.md differ from the planned bytes; nothing applied",
            Conflict,
        ),
        (
            "codex projection digest mismatch after write: SKILL.md",
            Conflict,
        ),
        (
            "partial apply: created 2 file(s) before x; failed at x: exclusive create refused",
            Partial,
        ),
        (
            "exclusive create refused (SKILL.md); nothing applied",
            Conflict,
        ),
        (
            "package changed since plan (SKILL.md); nothing applied",
            StalePlan,
        ),
        (
            "target directory write failed (SKILL.md); nothing applied",
            Io,
        ),
        (
            "invalid target: must be a directory, not a file or symlink",
            InvalidInput,
        ),
        (
            "invalid target: a symlink under the target is refused",
            InvalidInput,
        ),
        (
            "invalid target: '..' components are not allowed",
            InvalidInput,
        ),
        ("invalid target: path cannot be resolved", InvalidInput),
        ("invalid target: must be outside the package", InvalidInput),
        ("target file limit exceeded", InvalidInput),
        ("target metadata unavailable", Io),
        ("stale or altered plan; re-preview required", StalePlan),
        ("permission reference required", PermissionReferenceMissing),
    ] {
        assert_eq!(classify(message), kind, "{message}");
    }
}

/// #2241 audit FAIL: the outside-the-package check compared a lexical target with the
/// canonical package, so a target written through a symlinked parent (macOS `/tmp` →
/// `/private/tmp`) passed. This builds that shape explicitly, so it fails on every host.
#[cfg(unix)]
#[test]
fn codex_projection_refuses_a_package_target_reached_through_a_symlink() {
    let dir = TempDir::new().unwrap();
    let real = dir.path().join("real");
    let pkg = real.join("pkg");
    fs::create_dir_all(&pkg).unwrap();
    copy_tree(
        &Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai"),
        &pkg,
    );
    let link = dir.path().join("link");
    std::os::unix::fs::symlink(&real, &link).unwrap();
    let pkg_s = s(&pkg).to_string();
    let link_pkg = s(&link.join("pkg")).to_string();
    for (label, package, target) in [
        (
            "target through symlink",
            pkg_s.clone(),
            link.join("pkg/moai"),
        ),
        (
            "missing tail through symlink",
            pkg_s.clone(),
            link.join("pkg/a/b/moai"),
        ),
        ("package through symlink", link_pkg, pkg.join("moai")),
        ("dot-dot", pkg_s.clone(), real.join("x/../pkg/moai")),
    ] {
        let (ok, v) = run(&["codex-projection-plan", &package, "--target", s(&target)]);
        assert!(!ok, "plan {label}: {v}");
        assert_eq!(v["error"]["kind"], "invalid_input", "plan {label}: {v}");
        let (ok, v) = apply(&package, &target, "sha256:00", "fixture-approved");
        assert!(!ok, "apply {label}: {v}");
        assert_eq!(v["error"]["kind"], "invalid_input", "apply {label}: {v}");
    }
    assert!(!pkg.join("moai").exists());
    assert!(!pkg.join("a").exists());
    // Control: a target outside the package, reached through the same symlink, plans.
    let outside = link.join("out/moai");
    let (ok, v) = run(&["codex-projection-plan", &pkg_s, "--target", s(&outside)]);
    assert!(ok, "{v}");
}
