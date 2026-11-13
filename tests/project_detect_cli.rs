//! `project-detect` for the `/moai project` route (Issue #2496; user decisions 16 = B and 17 =
//! all three, #1975 comment 6815357071), through the binary in temporary folders only:
//! new/existing detection (AC-2), LSP·MCP·DB facts recorded in `project.yaml` with no MCP
//! configuration created and no secret leaked (AC-3), and `index-build` suggested but not run
//! (AC-4).

use serde_json::{json, Value};
use std::{
    collections::BTreeMap,
    fs,
    io::Write,
    path::Path,
    process::{Command, Output, Stdio},
};
use tempfile::TempDir;

const EXE: &str = env!("CARGO_BIN_EXE_moai-cli");
const PROFILE: &str = ".moai-cli/project.yaml";
const TOKEN: &str = "fixture-token-value-not-real-7f3a";

fn run(args: &[&str], stdin: &[u8]) -> Output {
    let mut child = Command::new(EXE)
        .args(args)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    let mut pipe = child.stdin.take().unwrap();
    match pipe.write_all(stdin) {
        Ok(()) => {}
        Err(e) if e.kind() == std::io::ErrorKind::BrokenPipe => {}
        Err(e) => panic!("stdin write failed: {e}"),
    }
    drop(pipe);
    child.wait_with_output().unwrap()
}

fn ok(out: &Output) -> Value {
    assert!(
        out.status.success(),
        "{}{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn p(root: &TempDir) -> &str {
    root.path().to_str().unwrap()
}

fn write(root: &Path, rel: &str, text: &str) {
    let path = root.join(rel);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, text).unwrap();
}

fn detect(root: &TempDir) -> Value {
    ok(&run(&["project-detect", p(root)], b""))
}

/// Every path below `root` with its bytes, so a test can show nothing was created or changed.
fn tree(root: &Path) -> BTreeMap<String, Vec<u8>> {
    fn walk(dir: &Path, base: &Path, out: &mut BTreeMap<String, Vec<u8>>) {
        for e in fs::read_dir(dir).unwrap().flatten() {
            let path = e.path();
            let rel = path.strip_prefix(base).unwrap().display().to_string();
            let kind = fs::symlink_metadata(&path).unwrap().file_type();
            if kind.is_dir() {
                out.insert(format!("{rel}/"), Vec::new());
                walk(&path, base, out);
            } else if kind.is_file() {
                out.insert(rel, fs::read(&path).unwrap());
            } else {
                out.insert(format!("{rel}@"), Vec::new());
            }
        }
    }
    let mut out = BTreeMap::new();
    walk(root, root, &mut out);
    out
}

fn with_profile() -> TempDir {
    let root = TempDir::new().unwrap();
    let profile = json!({"schema_version":2,"project_id":"demo","purpose":"Build a small editor","tools":["Rust"],"development_methods":["TDD"],"hosts":["claude"]});
    let mut bytes = serde_json::to_vec_pretty(&profile).unwrap();
    bytes.push(b'\n');
    write(root.path(), PROFILE, std::str::from_utf8(&bytes).unwrap());
    root
}

fn plan_apply(root: &TempDir, profile: &Value) -> Value {
    let plan = ok(&run(
        &["project-plan", p(root)],
        profile.to_string().as_bytes(),
    ));
    ok(&run(
        &[
            "project-apply",
            p(root),
            "--plan-digest",
            plan["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "fixture-approved",
        ],
        plan.to_string().as_bytes(),
    ));
    plan
}

/// AC-2: an empty folder and a folder a first launch initialized (#2481) are new; a manifest
/// or a code file makes a folder existing, with the language and its language server.
#[test]
fn new_and_existing_folders_are_told_apart() {
    let empty = TempDir::new().unwrap();
    let d = detect(&empty)["detected"].clone();
    assert_eq!(d["project_kind"], "new");
    assert_eq!(d["languages"], json!([]));

    let initialized = with_profile();
    fs::create_dir(initialized.path().join(".git")).unwrap();
    write(initialized.path(), "AGENTS.md", "# demo\n");
    write(
        initialized.path(),
        ".moai-cli/workflows/moai/read.md",
        "# READ\n",
    );
    write(
        initialized.path(),
        "target/debug/build.rs",
        "fn main() {}\n",
    );
    assert_eq!(detect(&initialized)["detected"]["project_kind"], "new");

    let manifest = TempDir::new().unwrap();
    write(manifest.path(), "Cargo.toml", "[package]\nname = \"x\"\n");
    let d = detect(&manifest)["detected"].clone();
    assert_eq!(d["project_kind"], "existing");
    assert_eq!(d["manifests"], json!(["Cargo.toml"]));
    assert_eq!(d["languages"], json!(["rust"]));
    assert_eq!(d["lsp"], json!(["rust-analyzer"]));

    let code = TempDir::new().unwrap();
    write(code.path(), "app/main.py", "print(1)\n");
    let d = detect(&code)["detected"].clone();
    assert_eq!(d["project_kind"], "existing");
    assert_eq!(d["languages"], json!(["python"]));
    assert_eq!(d["lsp"], json!(["pyright"]));
}

/// Symbolic links are never followed: a linked manifest or a linked code folder counts for
/// nothing.
#[cfg(unix)]
#[test]
fn symbolic_links_are_not_followed() {
    let outside = TempDir::new().unwrap();
    write(outside.path(), "Cargo.toml", "[package]\nname = \"x\"\n");
    write(outside.path(), "src/lib.rs", "pub fn f() {}\n");
    let root = TempDir::new().unwrap();
    std::os::unix::fs::symlink(
        outside.path().join("Cargo.toml"),
        root.path().join("Cargo.toml"),
    )
    .unwrap();
    std::os::unix::fs::symlink(outside.path().join("src"), root.path().join("src")).unwrap();
    let d = detect(&root)["detected"].clone();
    assert_eq!(d["project_kind"], "new", "{d}");
}

/// #2496 pre-audit B1/B2: no read leaves the project — a symlinked middle folder (`prisma`,
/// `.codex`), a symlinked file (`Cargo.toml`, `.mcp.json`, `requirements.txt`) and a symlink
/// loop all count for nothing, while the same files in real folders are found.
#[cfg(unix)]
#[test]
fn symlinked_folders_and_files_never_lead_outside() {
    use std::os::unix::fs::symlink;
    let outside = TempDir::new().unwrap();
    let o = outside.path();
    write(
        o,
        "prisma/schema.prisma",
        "datasource db {\n  provider = \"mysql\"\n}\n",
    );
    write(
        o,
        ".codex/config.toml",
        "[mcp_servers.leaked]\ncommand = \"x\"\n",
    );
    write(
        o,
        "Cargo.toml",
        "[package]\nname = \"x\"\n\n[dependencies]\nrusqlite = \"0.31\"\n",
    );
    write(
        o,
        "mcp.json",
        &json!({"mcpServers": {"leaked2": {}}}).to_string(),
    );
    write(o, "requirements.txt", "psycopg\n");

    let root = TempDir::new().unwrap();
    let r = root.path();
    symlink(o.join("prisma"), r.join("prisma")).unwrap();
    symlink(o.join(".codex"), r.join(".codex")).unwrap();
    symlink(o.join("Cargo.toml"), r.join("Cargo.toml")).unwrap();
    symlink(o.join("mcp.json"), r.join(".mcp.json")).unwrap();
    symlink(o.join("requirements.txt"), r.join("requirements.txt")).unwrap();
    fs::create_dir(r.join("a")).unwrap();
    symlink("..", r.join("a").join("up")).unwrap();
    symlink(".", r.join("loop")).unwrap();
    let d = detect(&root)["detected"].clone();
    assert_eq!(d["databases"], json!([]), "{d}");
    assert_eq!(d["mcp_servers"], json!([]), "{d}");
    assert_eq!(d["manifests"], json!([]), "{d}");
    assert_eq!(d["project_kind"], "new", "{d}");

    // Positive control: the same files inside the project are found.
    let real = TempDir::new().unwrap();
    for rel in [
        "prisma/schema.prisma",
        ".codex/config.toml",
        "Cargo.toml",
        "requirements.txt",
    ] {
        write(real.path(), rel, &fs::read_to_string(o.join(rel)).unwrap());
    }
    write(
        real.path(),
        ".mcp.json",
        &fs::read_to_string(o.join("mcp.json")).unwrap(),
    );
    let d = detect(&real)["detected"].clone();
    assert_eq!(
        d["databases"],
        json!(["mysql", "postgres", "sqlite"]),
        "{d}"
    );
    assert_eq!(d["mcp_servers"], json!(["leaked", "leaked2"]), "{d}");
}

/// AC-3: MCP server names and database kinds are recorded in `project.yaml`; MCP files are
/// read for names only — no token reaches the output or the profile, the files stay byte
/// for byte, and no MCP configuration is created.
#[test]
fn detected_facts_are_recorded_without_mcp_writes_or_secrets() {
    let root = with_profile();
    write(
        root.path(),
        ".mcp.json",
        &json!({"mcpServers": {"github": {"command": "npx", "env": {"GITHUB_TOKEN": TOKEN}}}})
            .to_string(),
    );
    write(
        root.path(),
        ".codex/config.toml",
        &format!(
            "[mcp_servers.docs]\ncommand = \"docs-server\"\nenv = {{ API_KEY = \"{TOKEN}\" }}\n"
        ),
    );
    write(
        root.path(),
        "Cargo.toml",
        "[package]\nname = \"x\"\n\n[dependencies]\nrusqlite = \"0.31\"\ntokio-postgres = \"0.7\"\n",
    );
    write(
        root.path(),
        "package.json",
        &json!({"dependencies": {"mongoose": "8"}, "devDependencies": {"left-pad": "1"}})
            .to_string(),
    );
    write(
        root.path(),
        "prisma/schema.prisma",
        "datasource db {\n  provider = \"mysql\"\n  url = env(\"DATABASE_URL\")\n}\n",
    );
    let before = tree(root.path());
    let out = run(&["project-detect", p(&root)], b"");
    assert!(!String::from_utf8_lossy(&out.stdout).contains(TOKEN));
    assert!(!String::from_utf8_lossy(&out.stderr).contains(TOKEN));
    let result = ok(&out);
    assert_eq!(tree(root.path()), before, "project-detect writes nothing");
    let d = &result["detected"];
    assert_eq!(d["mcp_servers"], json!(["docs", "github"]));
    assert_eq!(
        d["databases"],
        json!(["mongodb", "mysql", "postgres", "sqlite"])
    );
    assert_eq!(d["languages"], json!(["javascript", "rust"]));
    assert_eq!(result["profile_written"], json!(false));

    let plan = plan_apply(&root, &result["profile"]);
    assert_eq!(plan["delta"]["added"], json!(["detected"]));
    let yaml = fs::read_to_string(root.path().join(PROFILE)).unwrap();
    assert!(!yaml.contains(TOKEN));
    let stored: Value = serde_json::from_str(&yaml).unwrap();
    assert_eq!(&stored["detected"], d);
    // The MCP files are unchanged and no new MCP configuration exists.
    for rel in [".mcp.json", ".codex/config.toml"] {
        assert_eq!(
            fs::read(root.path().join(rel)).unwrap(),
            before[rel],
            "{rel}"
        );
    }
    let after = tree(root.path());
    let added: Vec<&String> = after.keys().filter(|k| !before.contains_key(*k)).collect();
    assert!(
        added.iter().all(|k| k.starts_with(".moai-cli/")),
        "only MoAI state is written by project-apply: {added:?}"
    );
    assert!(!added.iter().any(|k| k.contains("mcp")), "{added:?}");

    // A re-run on the same files plans `unchanged`.
    let again = detect(&root);
    let replan = ok(&run(
        &["project-plan", p(&root)],
        again["profile"].to_string().as_bytes(),
    ));
    assert_eq!(replan["action"], json!("unchanged"));
}

/// AC-4: `index-build` is suggested in `next_actions` and never run — detection leaves the
/// folder exactly as it was, so no index exists afterwards.
#[test]
fn index_build_is_suggested_but_not_run() {
    let root = with_profile();
    write(root.path(), "src/lib.rs", "pub fn f() {}\n");
    let before = tree(root.path());
    let result = detect(&root);
    assert_eq!(tree(root.path()), before);
    let actions = result["next_actions"].to_string();
    assert!(actions.contains("moai-cli index-build"), "{actions}");
    assert!(actions.contains("after the user agrees"), "{actions}");
    assert!(!root.path().join(".moai-cli/index").exists());
}

/// Without a profile the facts are still reported, and the step says how to get a profile.
#[test]
fn a_folder_without_a_profile_gets_the_facts_and_no_profile() {
    let root = TempDir::new().unwrap();
    write(
        root.path(),
        "go.mod",
        "module example.com/x\n\nrequire github.com/lib/pq v1.10.9\n",
    );
    let result = detect(&root);
    assert_eq!(result["profile"], Value::Null);
    assert_eq!(result["detected"]["databases"], json!(["postgres"]));
    assert_eq!(result["detected"]["lsp"], json!(["gopls"]));
    assert!(result["next_actions"]
        .to_string()
        .contains("create the profile first"));
}

/// A hand-edited `detected` that `project-detect` would never write is refused.
#[test]
fn a_malformed_detected_value_is_refused() {
    let root = with_profile();
    let mut profile: Value =
        serde_json::from_slice(&fs::read(root.path().join(PROFILE)).unwrap()).unwrap();
    for bad in [
        json!({"project_kind": "maybe", "manifests": [], "languages": [], "lsp": [], "mcp_servers": [], "databases": []}),
        json!({"project_kind": "new", "manifests": [], "languages": ["rust", "go"], "lsp": [], "mcp_servers": [], "databases": []}),
        json!({"project_kind": "new", "manifests": [], "languages": [], "lsp": [], "mcp_servers": ["../x"], "databases": []}),
        json!({"project_kind": "new", "manifests": [], "languages": [], "lsp": [], "mcp_servers": [], "databases": [], "extra": 1}),
    ] {
        profile["detected"] = bad.clone();
        let out = run(&["project-plan", p(&root)], profile.to_string().as_bytes());
        assert!(!out.status.success(), "{bad}");
    }
}

/// #2535 AC-2 (item 2): manifests, dependency databases and compose images are read at the
/// root only; a subfolder's code still counts for `languages`. The root control shows the same
/// dependency is found when it sits at the root.
#[test]
fn detection_reads_manifests_and_databases_at_the_root_only() {
    let root = TempDir::new().unwrap();
    write(
        root.path(),
        "web/package.json",
        r#"{"dependencies":{"pg":"8"}}"#,
    );
    write(root.path(), "web/index.js", "console.log(1)\n");
    write(
        root.path(),
        "docker-compose.yml",
        "services:\n  db:\n    image: postgres:16\n",
    );
    let d = detect(&root)["detected"].clone();
    assert_eq!(d["project_kind"], "existing");
    assert_eq!(d["manifests"], json!([]));
    assert_eq!(d["languages"], json!(["javascript"]));
    assert_eq!(d["databases"], json!([]));

    write(
        root.path(),
        "package.json",
        r#"{"dependencies":{"pg":"8"}}"#,
    );
    let d = detect(&root)["detected"].clone();
    assert_eq!(d["manifests"], json!(["package.json"]));
    assert_eq!(d["databases"], json!(["postgres"]));
}

/// #2535 AC-2 (item 4): `project_kind` speaks of code only. A folder holding nothing but a
/// Prisma schema and a Codex MCP configuration is `new`, and its database and server names are
/// still recorded as configuration facts.
#[test]
fn a_configuration_only_folder_is_new_with_its_facts() {
    let root = TempDir::new().unwrap();
    write(
        root.path(),
        "prisma/schema.prisma",
        "datasource db {\n  provider = \"postgresql\"\n  url = env(\"DATABASE_URL\")\n}\n",
    );
    write(
        root.path(),
        ".codex/config.toml",
        "[mcp_servers.docs]\ncommand = \"x\"\n",
    );
    let d = detect(&root)["detected"].clone();
    assert_eq!(d["project_kind"], "new");
    assert_eq!(d["manifests"], json!([]));
    assert_eq!(d["languages"], json!([]));
    assert_eq!(d["databases"], json!(["postgres"]));
    assert_eq!(d["mcp_servers"], json!(["docs"]));
}

/// #2535 AC-1 (item 1): `project-apply` takes the plan output's `result` object on stdin. The
/// whole response envelope is refused and writes nothing; the `result` object applies.
#[test]
fn project_apply_takes_the_plan_result_not_the_envelope() {
    let root = with_profile();
    let profile = detect(&root)["profile"].clone();
    let out = run(&["project-plan", p(&root)], profile.to_string().as_bytes());
    let envelope: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(envelope["ok"], json!(true), "{envelope}");
    let digest = envelope["result"]["plan_digest"]
        .as_str()
        .unwrap()
        .to_string();
    let apply = |stdin: &Value| {
        run(
            &[
                "project-apply",
                p(&root),
                "--plan-digest",
                &digest,
                "--permission-reference",
                "fixture-approved",
            ],
            stdin.to_string().as_bytes(),
        )
    };
    let before = tree(root.path());
    let refused = apply(&envelope);
    assert!(!refused.status.success());
    let v: Value = serde_json::from_slice(&refused.stdout).unwrap();
    assert_eq!(v["error"]["message"], json!("invalid typed JSON input"));
    assert_eq!(tree(root.path()), before, "a refused apply writes nothing");
    let applied = ok(&apply(&envelope["result"]));
    assert_eq!(applied["state"], json!("updated"));
}

/// #2535 item 7 (leader host observation 6826590481): a profile with no baseline holds every
/// changed key as a conflict, so an interview answer that changes `purpose` cannot apply. The
/// documented two steps get past it without a hand edit: apply the detected facts with every
/// existing value kept (that writes the baseline), then plan the answer again.
#[test]
fn a_profile_without_a_baseline_takes_the_answers_in_two_steps() {
    let root = with_profile();
    let baseline = root
        .path()
        .join(".moai-cli/manifests/profile-baseline.json");
    assert!(!baseline.exists());
    let facts = detect(&root)["profile"].clone();
    let mut answered = facts.clone();
    answered["purpose"] = json!("MoAI-ADK, an agentic development kit");

    let before = tree(root.path());
    let held = ok(&run(
        &["project-plan", p(&root)],
        answered.to_string().as_bytes(),
    ));
    assert_eq!(held["action"], json!("conflict"), "{held}");
    assert_eq!(held["delta"]["conflict"], json!(["purpose"]), "{held}");
    assert_eq!(tree(root.path()), before, "planning writes nothing");

    let first = plan_apply(&root, &facts);
    assert_eq!(first["action"], json!("update"), "{first}");
    assert!(first["delta"]["conflict"].as_array().unwrap().is_empty());
    assert!(baseline.is_file(), "the first apply writes the baseline");

    let second = plan_apply(&root, &answered);
    assert_eq!(second["action"], json!("update"), "{second}");
    assert_eq!(second["delta"]["changed"], json!(["purpose"]), "{second}");
    let written: Value =
        serde_json::from_slice(&fs::read(root.path().join(PROFILE)).unwrap()).unwrap();
    assert_eq!(written["purpose"], answered["purpose"]);
}

/// #2561: Go-generation project state is reported as a fact, read only, beside `detected`
/// (never inside it, so the profile and its digest do not change). Only when `.moai/` is a
/// real folder does the step suggest the read-only `go-state-plan` preview, and only to run
/// after the user agrees; nothing is moved or created by detection.
#[test]
fn go_generation_state_is_reported_and_its_preview_only_suggested() {
    // No `.moai/`: absent, no suggestion.
    let root = with_profile();
    let result = detect(&root);
    assert_eq!(
        result["go_state"],
        json!({"moai_dir": "absent", "go_manifest": false}),
        "{result}"
    );
    assert!(!result["next_actions"].to_string().contains("go-state-plan"));
    assert!(result["detected"].get("go_state").is_none(), "{result}");

    // A Go `.moai/` with its manifest: reported, and the preview is suggested after consent.
    let root = with_profile();
    write(
        root.path(),
        ".moai/manifest.json",
        "{\"version\":\"1.0.0\",\"files\":{}}\n",
    );
    write(root.path(), ".moai/specs/SPEC-1/spec.md", "# spec\n");
    let before = tree(root.path());
    let result = detect(&root);
    assert_eq!(tree(root.path()), before, "detection wrote nothing");
    assert_eq!(
        result["go_state"],
        json!({"moai_dir": "folder", "go_manifest": true}),
        "{result}"
    );
    let actions = result["next_actions"].to_string();
    assert!(actions.contains("moai-cli go-state-plan"), "{actions}");
    assert!(actions.contains("after the user agrees"), "{actions}");
    assert!(result["profile"].get("go_state").is_none(), "{result}");

    // A `.moai/` without a manifest is still a folder fact, with go_manifest false.
    let root = with_profile();
    write(root.path(), ".moai/config/x.yaml", "a: 1\n");
    let result = detect(&root);
    assert_eq!(
        result["go_state"],
        json!({"moai_dir": "folder", "go_manifest": false}),
        "{result}"
    );
}

/// #2561: a linked `.moai` is reported as a link and never followed (its target's manifest is
/// not read), and no preview is suggested for it.
#[cfg(unix)]
#[test]
fn a_linked_moai_folder_is_reported_not_followed() {
    let root = with_profile();
    let outside = TempDir::new().unwrap();
    write(
        outside.path(),
        "manifest.json",
        "{\"version\":\"1.0.0\",\"files\":{}}\n",
    );
    std::os::unix::fs::symlink(outside.path(), root.path().join(".moai")).unwrap();
    let result = detect(&root);
    assert_eq!(
        result["go_state"],
        json!({"moai_dir": "link", "go_manifest": false}),
        "{result}"
    );
    assert!(!result["next_actions"].to_string().contains("go-state-plan"));

    // A real `.moai/` whose manifest.json is a link: the folder is reported, the linked
    // manifest is not counted as the Go manifest.
    let root = with_profile();
    fs::create_dir_all(root.path().join(".moai")).unwrap();
    std::os::unix::fs::symlink(
        outside.path().join("manifest.json"),
        root.path().join(".moai/manifest.json"),
    )
    .unwrap();
    let result = detect(&root);
    assert_eq!(
        result["go_state"],
        json!({"moai_dir": "folder", "go_manifest": false}),
        "{result}"
    );
}

/// #2639 (read before asking): build and test command candidates and CI file names are
/// reported beside `detected` — fixed templates chosen by what the root manifests declare,
/// never a script body, never run — so the interview confirms them instead of asking.
#[test]
fn build_and_test_commands_and_ci_files_are_reported_beside_the_facts() {
    let empty = TempDir::new().unwrap();
    assert_eq!(
        detect(&empty)["commands"],
        json!({"build": [], "test": [], "ci_files": []})
    );

    let root = TempDir::new().unwrap();
    let r = root.path();
    write(r, "Cargo.toml", "[package]\nname = \"x\"\n");
    write(
        r,
        "package.json",
        &format!(
            r#"{{"scripts": {{"test": "echo {TOKEN}", "build": "tsc"}}, "devDependencies": {{}}}}"#
        ),
    );
    write(r, "pnpm-lock.yaml", "");
    write(r, "go.mod", "module example.com/x\n");
    write(r, "requirements.txt", "pytest==8.0\n");
    write(
        r,
        "Makefile",
        "build:\n\techo\ntest:\n\techo\nlint:\n\techo\n",
    );
    write(r, "pom.xml", "<project/>");
    write(r, "build.gradle", "");
    write(r, "gradlew", "");
    write(r, ".github/workflows/ci.yml", "on: push\n");
    write(r, ".github/workflows/release.yaml", "on: push\n");
    write(r, ".github/workflows/notes.txt", "not a workflow\n");
    let before = tree(r);
    let result = detect(&root);
    assert_eq!(
        result["commands"],
        json!({
            "build": ["./gradlew build", "cargo build", "go build ./...", "make build", "mvn package", "pnpm run build"],
            "test": ["./gradlew test", "cargo test", "go test ./...", "make test", "mvn test", "pnpm test", "pytest"],
            "ci_files": ["ci.yml", "release.yaml"],
        })
    );
    // A script body never reaches the output, nothing is written, and the profile is unchanged.
    assert!(!result.to_string().contains(TOKEN), "{result}");
    assert_eq!(tree(r), before);
    assert!(result["detected"].get("commands").is_none());

    // The package manager follows the lock file; without one it is npm, and a package.json
    // without the script names no command.
    let npm = TempDir::new().unwrap();
    write(
        npm.path(),
        "package.json",
        r#"{"scripts": {"test": "jest"}}"#,
    );
    assert_eq!(detect(&npm)["commands"]["test"], json!(["npm test"]));
    let yarn = TempDir::new().unwrap();
    write(yarn.path(), "package.json", r#"{"scripts": {"lint": "x"}}"#);
    write(yarn.path(), "yarn.lock", "");
    assert_eq!(
        detect(&yarn)["commands"],
        json!({"build": [], "test": [], "ci_files": []})
    );
}

#[cfg(unix)]
#[test]
fn a_symlinked_workflows_folder_is_not_read() {
    let outside = TempDir::new().unwrap();
    write(outside.path(), "workflows/leak.yml", "on: push\n");
    let root = TempDir::new().unwrap();
    std::os::unix::fs::symlink(outside.path(), root.path().join(".github")).unwrap();
    assert_eq!(detect(&root)["commands"]["ci_files"], json!([]));
    let inner = TempDir::new().unwrap();
    fs::create_dir_all(inner.path().join(".github")).unwrap();
    std::os::unix::fs::symlink(
        outside.path().join("workflows"),
        inner.path().join(".github/workflows"),
    )
    .unwrap();
    assert_eq!(detect(&inner)["commands"]["ci_files"], json!([]));
}
