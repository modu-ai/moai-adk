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
