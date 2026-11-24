//! `project-detect` finds the sub-package manifests of a monorepo in any language and says when
//! its list is partial (SPEC-20261124-project-detect-monorepo REQ-1 to REQ-8), through the
//! binary in temporary folders and a temporary HOME only.

mod common;
use common::home::TempHome;
use serde_json::{json, Value};
use std::{
    collections::BTreeMap,
    fs,
    io::Write,
    path::Path,
    process::{Output, Stdio},
};
use tempfile::TempDir;

const PARTIAL: &str = "manifest list is partial";
/// The project's own catalog rows: one file under the `config/` folder of the state folder.
fn extension() -> String {
    // Built at run time: the state-folder scan lists test files that spell the folder name.
    format!(".{}/config/detect-catalog.json", "moai")
}

fn run(args: &[&str]) -> Output {
    // A temporary HOME, so the run never reads the real one.
    let home = TempHome::new();
    let mut child = home
        .command(args)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    let mut pipe = child.stdin.take().unwrap();
    match pipe.write_all(b"") {
        Ok(()) => {}
        Err(e) if e.kind() == std::io::ErrorKind::BrokenPipe => {}
        Err(e) => panic!("stdin write failed: {e}"),
    }
    drop(pipe);
    child.wait_with_output().unwrap()
}

fn detect(root: &TempDir) -> Value {
    let out = run(&["project-detect", root.path().to_str().unwrap()]);
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

fn write(root: &Path, rel: &str, text: &str) {
    let path = root.join(rel);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, text).unwrap();
}

fn strings(value: &Value) -> Vec<&str> {
    value
        .as_array()
        .unwrap()
        .iter()
        .map(|m| m.as_str().unwrap())
        .collect()
}

fn manifests(result: &Value) -> Vec<&str> {
    strings(&result["detected"]["manifests"])
}

fn partial_lines(result: &Value) -> usize {
    result["next_actions"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|l| l.as_str().unwrap().contains(PARTIAL))
        .count()
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

/// More entries than the walk budget (5000) sit in one folder three levels down that sorts
/// before `apps/`.
fn a_project_with_a_large_early_folder() -> TempDir {
    let root = TempDir::new().unwrap();
    let big = root.path().join("aaa/b/c");
    fs::create_dir_all(&big).unwrap();
    for i in 0..6000 {
        fs::write(big.join(format!("f{i:05}.txt")), "x").unwrap();
    }
    write(root.path(), "apps/x/package.json", "{}");
    write(root.path(), "apps/id/go.mod", "module x\n");
    root
}

/// REQ-1: the walk reads shallow folders first, so a large folder that sorts early does not
/// use up the budget before `apps/` is read; the list is reported as partial because the
/// budget did run out.
#[test]
fn a_large_early_folder_does_not_hide_a_later_packages_manifest() {
    let root = a_project_with_a_large_early_folder();
    let result = detect(&root);
    assert_eq!(
        manifests(&result),
        ["apps/id/go.mod", "apps/x/package.json"],
        "{result}"
    );
    assert_eq!(
        result["manifest_walk"]["complete"],
        json!(false),
        "{result}"
    );
    assert_eq!(result["manifest_walk"]["over_cap"], json!(0), "{result}");
}

/// REQ-2: tool, cache and backup folders (names starting with a dot) are not walked for
/// manifests; a root manifest is read whatever the folder names are.
#[test]
fn hidden_folders_are_not_walked_for_manifests() {
    let root = TempDir::new().unwrap();
    write(root.path(), "package.json", "{}");
    write(root.path(), ".cache/pkg/package.json", "{}");
    write(root.path(), ".vercel/x/package.json", "{}");
    write(root.path(), ".moai-backup/y/go.mod", "module y\n");
    write(root.path(), "apps/x/package.json", "{}");
    let result = detect(&root);
    assert_eq!(
        manifests(&result),
        ["apps/x/package.json", "package.json"],
        "{result}"
    );
    assert_eq!(result["manifest_walk"]["complete"], json!(true), "{result}");
}

/// REQ-3, REQ-4 (the card's temporary monorepo): the manifests of three languages below the
/// root are all listed, the tools and languages they imply are the ones the earlier tables
/// gave, and no manifest content is read.
#[test]
fn a_temporary_monorepo_lists_every_sub_manifest() {
    let root = TempDir::new().unwrap();
    let r = root.path();
    write(r, "package.json", "{}");
    write(r, "apps/a/package.json", "{ not json");
    write(r, "apps/b/tsconfig.json", "{ not json");
    write(r, "apps/id/go.mod", "module id\n");
    let result = detect(&root);
    assert_eq!(
        manifests(&result),
        [
            "apps/a/package.json",
            "apps/b/tsconfig.json",
            "apps/id/go.mod",
            "package.json"
        ],
        "{result}"
    );
    assert_eq!(
        result["tools"],
        json!({
            "go": ["apps/id/go.mod"],
            "node": ["apps/a/package.json", "package.json"],
            "npm": ["apps/a/package.json", "package.json"],
            "typescript": ["apps/b/tsconfig.json"],
        }),
        "{result}"
    );
    assert_eq!(
        result["detected"]["languages"],
        json!(["go", "javascript", "typescript"]),
        "{result}"
    );
    assert_eq!(
        result["manifest_catalog"],
        json!({"extension": "absent", "rows": 0, "problem": null}),
        "{result}"
    );
    let before = tree(r);
    detect(&root);
    assert_eq!(tree(r), before);
}

/// REQ-2: a folder the shipped catalog lists is not walked, and a folder the project's own
/// catalog rows list is not walked either; a folder neither lists is.
#[test]
fn folders_the_catalog_lists_are_not_walked() {
    let root = TempDir::new().unwrap();
    let r = root.path();
    write(r, "node_modules/x/package.json", "{}");
    write(r, "target/x/Cargo.toml", "[package]\n");
    write(r, "vendor/x/go.mod", "module x\n");
    write(r, "generated-out/p/package.json", "{}");
    write(r, "kept/p/package.json", "{}");
    let result = detect(&root);
    assert_eq!(
        manifests(&result),
        ["generated-out/p/package.json", "kept/p/package.json"],
        "{result}"
    );

    write(
        r,
        &extension(),
        r#"{"schema": 1, "skip_dirs": ["generated-out"]}"#,
    );
    let result = detect(&root);
    assert_eq!(manifests(&result), ["kept/p/package.json"], "{result}");
    assert_eq!(
        result["manifest_catalog"]["problem"],
        json!(null),
        "{result}"
    );
}

/// REQ-3, REQ-5: a language the shipped catalog does not know is found through one row of the
/// project's own catalog, by a file name and by a suffix, and the row's language and tools
/// reach the output; without the file the manifests are not listed. No engine code names it.
#[test]
fn a_row_in_the_extension_file_adds_a_language_without_code() {
    let root = TempDir::new().unwrap();
    let r = root.path();
    write(r, "apps/svc/mix.exs", "defmodule X do end\n");
    write(r, "apps/web/App.csproj", "<Project />\n");
    let result = detect(&root);
    assert!(manifests(&result).is_empty(), "{result}");

    write(
        r,
        &extension(),
        r#"{"schema": 1, "manifests": [
            {"ecosystem": "hex", "language": "elixir", "files": ["mix.exs"], "tools": ["mix"]},
            {"ecosystem": "nuget", "language": "csharp", "suffixes": [".csproj"], "tools": ["dotnet"]}
        ]}"#,
    );
    let result = detect(&root);
    assert_eq!(
        manifests(&result),
        ["apps/svc/mix.exs", "apps/web/App.csproj"],
        "{result}"
    );
    assert_eq!(
        result["detected"]["languages"],
        json!(["csharp", "elixir"]),
        "{result}"
    );
    assert_eq!(
        result["tools"],
        json!({"dotnet": ["apps/web/App.csproj"], "mix": ["apps/svc/mix.exs"]}),
        "{result}"
    );
    assert_eq!(
        result["manifest_catalog"],
        json!({"extension": "read", "rows": 2, "problem": null}),
        "{result}"
    );
    let before = tree(r);
    detect(&root);
    assert_eq!(tree(r), before);

    fs::remove_file(r.join(extension())).unwrap();
    let result = detect(&root);
    assert!(manifests(&result).is_empty(), "{result}");
    assert_eq!(result["detected"]["languages"], json!([]), "{result}");
}

/// REQ-5, REQ-6: a project row never replaces a shipped row, and an extension file that is a
/// link, too large, not valid or holds a value that cannot be recorded is reported and left
/// out whole, the shipped rows alone are used and nothing is written.
#[test]
fn an_extension_file_adds_rows_and_a_bad_one_is_reported_and_ignored() {
    let root = TempDir::new().unwrap();
    let r = root.path();
    write(r, "apps/a/package.json", "{}");
    write(r, "apps/svc/mix.exs", "defmodule X do end\n");

    // A row for a name the shipped catalog already has changes nothing of that name.
    write(
        r,
        &extension(),
        r#"{"schema": 1, "manifests": [
            {"ecosystem": "mine", "language": "elixir", "files": ["package.json"], "tools": ["mix"]},
            {"ecosystem": "hex", "language": "elixir", "files": ["mix.exs"], "tools": ["mix"]}
        ]}"#,
    );
    let result = detect(&root);
    assert_eq!(
        manifests(&result),
        ["apps/a/package.json", "apps/svc/mix.exs"],
        "{result}"
    );
    assert_eq!(
        result["tools"],
        json!({
            "mix": ["apps/svc/mix.exs"],
            "node": ["apps/a/package.json"],
            "npm": ["apps/a/package.json"],
        }),
        "{result}"
    );
    assert_eq!(
        result["detected"]["languages"],
        json!(["elixir", "javascript"]),
        "{result}"
    );

    // Each bad file is reported and none of its rows is used.
    let good_row = r#"{"ecosystem": "hex", "language": "elixir", "files": ["mix.exs"]}"#;
    let bad_files = [
        "{ not json".to_string(),
        format!(r#"{{"schema": 1, "manifests": [{good_row}], "surprise": 1}}"#),
        r#"{"schema": 1, "manifests": [{"ecosystem": "x y", "language": "elixir", "files": ["mix.exs"]}]}"#
            .to_string(),
        r#"{"schema": 1, "manifests": [{"ecosystem": "hex", "language": "elixir", "files": ["../mix.exs"]}]}"#
            .to_string(),
        format!(
            r#"{{"schema": 1, "manifests": [{good_row}]}}{}"#,
            " ".repeat(300_000)
        ),
    ];
    for bad in &bad_files {
        write(r, &extension(), bad);
        let before = tree(r);
        let result = detect(&root);
        assert_eq!(manifests(&result), ["apps/a/package.json"], "{result}");
        assert_eq!(
            result["manifest_catalog"]["extension"],
            json!("refused"),
            "{result}"
        );
        assert!(
            result["manifest_catalog"]["problem"].is_string(),
            "{result}"
        );
        assert_eq!(tree(r), before, "a read wrote something");
    }

    #[cfg(unix)]
    {
        // A link to a valid file is not followed.
        let elsewhere = TempDir::new().unwrap();
        write(
            elsewhere.path(),
            "rows.json",
            &format!(r#"{{"schema": 1, "manifests": [{good_row}]}}"#),
        );
        fs::remove_file(r.join(extension())).unwrap();
        std::os::unix::fs::symlink(elsewhere.path().join("rows.json"), r.join(extension()))
            .unwrap();
        let result = detect(&root);
        assert_eq!(manifests(&result), ["apps/a/package.json"], "{result}");
        assert!(
            result["manifest_catalog"]["problem"].is_string(),
            "{result}"
        );
    }
}

/// REQ-7: a small project reports a complete walk and nothing over the cap; 70 manifests are
/// 64 listed and 6 over the cap.
#[test]
fn the_manifest_walk_fact_reports_completeness_and_the_cap() {
    let small = TempDir::new().unwrap();
    write(small.path(), "package.json", "{}");
    write(small.path(), "apps/a/package.json", "{}");
    let result = detect(&small);
    assert_eq!(
        result["manifest_walk"],
        json!({"complete": true, "over_cap": 0}),
        "{result}"
    );

    let many = TempDir::new().unwrap();
    for i in 0..70 {
        write(many.path(), &format!("pkg{i:02}/package.json"), "{}");
    }
    let result = detect(&many);
    assert_eq!(manifests(&result).len(), 64, "{result}");
    assert_eq!(
        result["manifest_walk"],
        json!({"complete": true, "over_cap": 6}),
        "{result}"
    );
}

/// REQ-8: one `next_actions` line says the list is partial when the walk stopped or the cap cut
/// the list, and none when the list is complete.
#[test]
fn a_partial_list_adds_one_next_action_and_a_complete_one_adds_none() {
    let stopped = a_project_with_a_large_early_folder();
    assert_eq!(partial_lines(&detect(&stopped)), 1);

    let many = TempDir::new().unwrap();
    for i in 0..65 {
        write(many.path(), &format!("pkg{i:02}/package.json"), "{}");
    }
    assert_eq!(partial_lines(&detect(&many)), 1);

    let small = TempDir::new().unwrap();
    write(small.path(), "apps/a/package.json", "{}");
    assert_eq!(partial_lines(&detect(&small)), 0);
}
