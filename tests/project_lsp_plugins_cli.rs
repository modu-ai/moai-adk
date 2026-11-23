//! SPEC-20261123-lsp-official-plugins: LSP only through the official Claude Code plugins.
//! The rule text (AC-1), the PROJECT workflow steps (AC-2) and the `lsp_plugins` fact of
//! `project-detect` (AC-3 to AC-5), through the binary in temporary folders only. The PATH the
//! binary sees is always a temporary folder the test builds; the real PATH is never read.

mod common;
use common::home::TempHome;
use common::CleanEnv;
use serde_json::{json, Value};
use std::{
    collections::BTreeMap,
    ffi::OsStr,
    fs,
    path::{Path, PathBuf},
    process::{Output, Stdio},
};
use tempfile::TempDir;

fn package(rel: &str) -> String {
    fs::read_to_string(
        Path::new(env!("CARGO_MANIFEST_DIR"))
            .join("skills/moai")
            .join(rel),
    )
    .unwrap()
}

fn write(root: &Path, rel: &str, text: &str) {
    let path = root.join(rel);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, text).unwrap();
}

/// Runs `project-detect` on `root` with `PATH` set to `path` (`None` removes the variable).
fn detect_with_path(root: &Path, path: Option<&OsStr>) -> Value {
    // The clean environment of a TempHome child holds no PATH; the test adds the one it wants.
    let home = TempHome::new();
    let mut cmd = home.command(&["project-detect", root.to_str().unwrap()]);
    cmd.stdin(Stdio::null())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped());
    if let Some(p) = path {
        cmd.env("PATH", p);
    }
    let out: Output = cmd.output().unwrap();
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

/// A fake server: an executable file whose body leaves a marker when it is run. The marker is
/// written with the shell's own `:` and redirection, never an external program: the PATH the
/// binary sees holds only a temporary folder, so a `touch` there would exit 127 and leave no marker
/// even when the server did run (`a_fake_server_leaves_its_marker_with_only_its_own_folder_on_the_path`
/// pins this with that same one-folder `PATH`).
#[cfg(unix)]
fn fake_server(dir: &Path, name: &str, marker: &Path) {
    use std::os::unix::fs::PermissionsExt;
    let file = dir.join(name);
    fs::write(&file, format!("#!/bin/sh\n: > '{}'\n", marker.display())).unwrap();
    fs::set_permissions(&file, fs::Permissions::from_mode(0o755)).unwrap();
}

fn tree(root: &Path) -> BTreeMap<String, Vec<u8>> {
    fn walk(dir: &Path, base: &Path, out: &mut BTreeMap<String, Vec<u8>>) {
        for e in fs::read_dir(dir).unwrap().flatten() {
            let path: PathBuf = e.path();
            let rel = path.strip_prefix(base).unwrap().display().to_string();
            if path.is_dir() {
                out.insert(format!("{rel}/"), Vec::new());
                walk(&path, base, out);
            } else {
                out.insert(rel, fs::read(&path).unwrap());
            }
        }
    }
    let mut out = BTreeMap::new();
    walk(root, root, &mut out);
    out
}

fn row<'a>(result: &'a Value, plugin: &str) -> &'a Value {
    result["lsp_plugins"]
        .as_array()
        .unwrap()
        .iter()
        .find(|r| r["plugin"] == plugin)
        .unwrap_or_else(|| panic!("no row for {plugin}: {}", result["lsp_plugins"]))
}

/// AC-1: rule 25 of the core rules.
#[test]
fn core_rules_state_the_official_lsp_plugin_rule() {
    let rules = package("references/internal/core-rules.md");
    let rule = rules
        .lines()
        .find(|l| l.starts_with("25. "))
        .expect("rule 25");
    for phrase in [
        "official",
        "<language>-lsp@claude-plugins-official",
        "never with Homebrew",
        "README",
        "selection",
        "`guarded`",
        "/reload-plugins",
        "no LSP was used",
    ] {
        assert!(rule.contains(phrase), "rule 25 must say `{phrase}`: {rule}");
    }
    // The fixed numbers keep their place: nothing moved.
    for n in [3, 7, 19, 20, 21, 22, 23, 24] {
        assert!(
            rules.lines().any(|l| l.starts_with(&format!("{n}. "))),
            "rule {n} is gone"
        );
    }
}

/// AC-2: the PROJECT workflow step texts.
#[test]
fn project_workflow_has_the_lsp_and_recommendation_steps() {
    let project = package("workflows/project.md");
    for phrase in [
        "`lsp_plugins`",
        "claude plugin list",
        "claude plugin install <name>@claude-plugins-official",
        "one reason line",
        "reads the names of the files in the PATH folders",
    ] {
        assert!(project.contains(phrase), "project.md must say `{phrase}`");
    }
    // The old promise that detection never probes PATH is gone: it names its one exception.
    assert!(
        !project.contains("never probes PATH"),
        "project.md still says detection never probes PATH"
    );
}

/// AC-3: one row per official plugin, the PATH state, and the server names of `detected.lsp`.
#[cfg(unix)]
#[test]
fn lsp_plugins_name_the_official_plugin_and_the_path_state() {
    let project = TempDir::new().unwrap();
    write(project.path(), "Cargo.toml", "[package]\nname = \"x\"\n");
    write(project.path(), "web/app.ts", "export const a = 1;\n");
    write(project.path(), "web/app.js", "export const b = 2;\n");
    write(project.path(), "kt/Main.kt", "fun main() {}\n");
    write(project.path(), "lib/init.lua", "return {}\n");
    let bin = TempDir::new().unwrap();
    let marker = TempDir::new().unwrap();
    fake_server(bin.path(), "rust-analyzer", &marker.path().join("ra"));
    fake_server(
        bin.path(),
        "typescript-language-server",
        &marker.path().join("ts"),
    );

    let result = detect_with_path(project.path(), Some(bin.path().as_os_str()));

    let rust = row(&result, "rust-analyzer-lsp");
    assert_eq!(rust["marketplace"], "claude-plugins-official");
    assert_eq!(rust["languages"], json!(["rust"]));
    assert_eq!(rust["executable"], "rust-analyzer");
    assert_eq!(rust["installed"], json!(true));
    let ts = row(&result, "typescript-lsp");
    assert_eq!(ts["languages"], json!(["javascript", "typescript"]));
    assert_eq!(ts["executable"], "typescript-language-server");
    assert_eq!(ts["installed"], json!(true));
    let kotlin = row(&result, "kotlin-lsp");
    assert_eq!(kotlin["executable"], "kotlin-lsp");
    assert_eq!(kotlin["installed"], json!(false));
    let lua = row(&result, "lua-lsp");
    assert_eq!(lua["executable"], "lua-language-server");
    assert_eq!(lua["installed"], json!(false));
    assert_eq!(result["lsp_plugins"].as_array().unwrap().len(), 4);

    // `detected.lsp` names the server of the official plugin of each language (REQ-12).
    assert_eq!(
        result["detected"]["lsp"],
        json!([
            "kotlin-lsp",
            "lua-language-server",
            "rust-analyzer",
            "typescript-language-server"
        ])
    );
    // A project with no language reports an empty list.
    let empty = TempDir::new().unwrap();
    let none = detect_with_path(empty.path(), Some(bin.path().as_os_str()));
    assert_eq!(none["lsp_plugins"], json!([]));
}

/// AC-4: a missing PATH, a missing folder or a file without the executable bit read as not
/// installed, and `project-detect` still succeeds.
#[cfg(unix)]
#[test]
fn an_unreadable_or_missing_path_reads_as_not_installed() {
    let project = TempDir::new().unwrap();
    write(project.path(), "Cargo.toml", "[package]\nname = \"x\"\n");
    let not_exec = TempDir::new().unwrap();
    write(not_exec.path(), "rust-analyzer", "#!/bin/sh\n");
    let gone = not_exec.path().join("no-such-folder");
    let joined = std::env::join_paths([gone.as_path(), not_exec.path()]).unwrap();
    for path in [None, Some(joined.as_os_str()), Some(OsStr::new(""))] {
        let result = detect_with_path(project.path(), path);
        assert_eq!(
            row(&result, "rust-analyzer-lsp")["installed"],
            json!(false),
            "{path:?}"
        );
    }
}

/// The guard of AC-5 can fail: the fake server leaves its marker when run exactly as the real
/// check sees it, with a `PATH` that holds only the server's own temporary folder. A cleared
/// environment with no `PATH` at all is not the same: `/bin/sh` then searches its default path and
/// finds `/usr/bin/touch`, so a server that still called `touch` would pass.
#[cfg(unix)]
#[test]
fn a_fake_server_leaves_its_marker_with_only_its_own_folder_on_the_path() {
    let bin = TempDir::new().unwrap();
    let marker_dir = TempDir::new().unwrap();
    let marker = marker_dir.path().join("started");
    fake_server(bin.path(), "rust-analyzer", &marker);
    let status = std::process::Command::new(bin.path().join("rust-analyzer"))
        .clean_env()
        .env("PATH", bin.path())
        .stdin(Stdio::null())
        .status()
        .unwrap();
    assert!(status.success(), "the fake server exited with {status}");
    assert!(marker.exists(), "the fake server left no marker");
}

/// AC-5: the check reads file names only: nothing is written and no server is started.
#[cfg(unix)]
#[test]
fn the_path_check_writes_nothing_and_starts_nothing() {
    let project = TempDir::new().unwrap();
    write(project.path(), "Cargo.toml", "[package]\nname = \"x\"\n");
    let bin = TempDir::new().unwrap();
    let marker_dir = TempDir::new().unwrap();
    let marker = marker_dir.path().join("started");
    fake_server(bin.path(), "rust-analyzer", &marker);
    let project_before = tree(project.path());
    let bin_before = tree(bin.path());

    let result = detect_with_path(project.path(), Some(bin.path().as_os_str()));

    assert_eq!(row(&result, "rust-analyzer-lsp")["installed"], json!(true));
    assert!(!marker.exists(), "the fake server was started");
    assert_eq!(tree(project.path()), project_before);
    assert_eq!(tree(bin.path()), bin_before);
}
