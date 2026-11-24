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

/// The extension file with a valid body padded by trailing blanks to exactly `len` bytes.
fn extension_of_len(body: &str, len: usize) -> String {
    assert!(body.len() <= len, "the body is longer than {len} bytes");
    format!("{body}{}", " ".repeat(len - body.len()))
}

/// REQ-6 (repair): the file limit holds at both edges — 262144 bytes are read, 262145 are
/// refused — for the extension file.
#[test]
fn the_extension_file_limit_holds_at_both_edges() {
    let root = TempDir::new().unwrap();
    let r = root.path();
    write(r, "apps/svc/mix.exs", "defmodule X do end\n");
    let body = r#"{"schema": 1, "manifests": [{"ecosystem": "hex", "language": "elixir", "files": ["mix.exs"]}]}"#;

    write(r, &extension(), &extension_of_len(body, 262_144));
    let result = detect(&root);
    assert_eq!(manifests(&result), ["apps/svc/mix.exs"], "{result}");
    assert_eq!(
        result["manifest_catalog"],
        json!({"extension": "read", "rows": 1, "problem": null}),
        "{result}"
    );

    write(r, &extension(), &extension_of_len(body, 262_145));
    let result = detect(&root);
    assert!(manifests(&result).is_empty(), "{result}");
    assert_eq!(
        result["manifest_catalog"]["extension"],
        json!("refused"),
        "{result}"
    );
    assert!(
        result["manifest_catalog"]["problem"].is_string(),
        "{result}"
    );
}

/// The extension file with `rows` rows, each naming one file of its own.
fn rows_json(rows: usize) -> String {
    let rows: Vec<String> = (0..rows)
        .map(|i| format!(r#"{{"ecosystem": "e{i}", "language": "l{i}", "files": ["f{i}.mk"]}}"#))
        .collect();
    format!(r#"{{"schema": 1, "manifests": [{}]}}"#, rows.join(","))
}

/// One row whose `key` list holds `n` names built by `name`, beside the file name every row
/// needs.
fn one_row_with(key: &str, n: usize, name: fn(usize) -> String) -> String {
    let list: Vec<String> = (0..n).map(|i| format!("\"{}\"", name(i))).collect();
    let list = list.join(",");
    let row = match key {
        "files" => format!(r#"{{"ecosystem": "e", "language": "l", "files": [{list}]}}"#),
        "suffixes" => format!(
            r#"{{"ecosystem": "e", "language": "l", "files": ["a.mk"], "suffixes": [{list}]}}"#
        ),
        _ => format!(
            r#"{{"ecosystem": "e", "language": "l", "files": ["a.mk"], "tools": [{list}]}}"#
        ),
    };
    format!(r#"{{"schema": 1, "manifests": [{row}]}}"#)
}

/// The state of the extension file in a project that holds `text` and the number of rows read.
fn extension_state(text: &str) -> (String, i64) {
    let root = TempDir::new().unwrap();
    write(root.path(), "package.json", "{}");
    write(root.path(), &extension(), text);
    let result = detect(&root);
    (
        result["manifest_catalog"]["extension"]
            .as_str()
            .unwrap()
            .to_string(),
        result["manifest_catalog"]["rows"].as_i64().unwrap(),
    )
}

/// The row and list limit of the extension file (64) holds at both edges: 64 rows, 64 folders
/// to skip and 64 names in one row's list are read, 65 are refused.
#[test]
fn the_extension_row_and_list_limits_hold_at_both_edges() {
    assert_eq!(extension_state(&rows_json(64)), ("read".to_string(), 64));
    assert_eq!(extension_state(&rows_json(65)).0, "refused");

    let skip = |n: usize| {
        let dirs: Vec<String> = (0..n).map(|i| format!("\"out{i}\"")).collect();
        format!(
            r#"{{"schema": 1, "manifests": [], "skip_dirs": [{}]}}"#,
            dirs.join(",")
        )
    };
    assert_eq!(extension_state(&skip(64)).0, "read");
    assert_eq!(extension_state(&skip(65)).0, "refused");

    let lists: [(&str, fn(usize) -> String); 3] = [
        ("files", |i| format!("f{i}.mk")),
        ("suffixes", |i| format!(".s{i}")),
        ("tools", |i| format!("t{i}")),
    ];
    for (key, name) in lists {
        assert_eq!(
            extension_state(&one_row_with(key, 64, name)),
            ("read".to_string(), 1),
            "{key} 64"
        );
        assert_eq!(
            extension_state(&one_row_with(key, 65, name)).0,
            "refused",
            "{key} 65"
        );
    }
}

/// A name of dots alone (`.` or `..`) is not a name that can be recorded, in any field of an
/// extension row; a name that merely holds a dot still is.
#[test]
fn a_name_of_only_dots_is_refused_in_the_extension_file() {
    let rows = [
        r#"{"ecosystem": "e", "language": "l", "files": [".."]}"#,
        r#"{"ecosystem": "e", "language": "l", "files": ["a.mk"], "suffixes": [".."]}"#,
        r#"{"ecosystem": "e", "language": "l", "files": ["a.mk"], "tools": ["."]}"#,
        r#"{"ecosystem": "..", "language": "l", "files": ["a.mk"]}"#,
        r#"{"ecosystem": "e", "language": "...", "files": ["a.mk"]}"#,
    ];
    for row in rows {
        let text = format!(r#"{{"schema": 1, "manifests": [{row}]}}"#);
        assert_eq!(extension_state(&text).0, "refused", "{row}");
    }
    let skip = r#"{"schema": 1, "manifests": [], "skip_dirs": [".."]}"#;
    assert_eq!(extension_state(skip).0, "refused");
    let fine = r#"{"schema": 1, "manifests": [{"ecosystem": "e", "language": "l", "files": ["a.mk"], "suffixes": [".b.mk"], "tools": ["t.1"]}]}"#;
    assert_eq!(extension_state(fine), ("read".to_string(), 1));
}

/// An ignore file of exactly 262144 bytes is read and one of 262145 bytes is not, for the root
/// list and for a list in a sub-folder; a list that is not read counts as no lines.
#[test]
fn the_ignore_file_limit_holds_at_both_edges() {
    let list = |len: usize| extension_of_len("skipme/\n", len);
    for path in [".gitignore", "apps/.gitignore"] {
        let root = TempDir::new().unwrap();
        write(root.path(), "apps/skipme/package.json", "{}");
        write(root.path(), "apps/a/package.json", "{}");
        write(root.path(), path, &list(262_144));
        let result = detect(&root);
        assert_eq!(
            manifests(&result),
            ["apps/a/package.json"],
            "{path} {result}"
        );
        write(root.path(), path, &list(262_145));
        let result = detect(&root);
        assert_eq!(
            manifests(&result),
            ["apps/a/package.json", "apps/skipme/package.json"],
            "{path} {result}"
        );
    }
}

/// The ignore lists the walk reads are capped in total (1 MiB), so a project with hundreds of
/// large `.gitignore` files cannot hold them all in memory. The lists are read in the walk's
/// order; a list that would pass the total counts as no lines, so it can only let more
/// manifests in, never hide one.
#[test]
fn the_ignore_files_the_walk_reads_have_a_total_size_cap() {
    let root = TempDir::new().unwrap();
    let list = format!("skipme/\n{}", "#\n".repeat(100_000));
    for i in 1..=7 {
        write(root.path(), &format!("d{i}/skipme/package.json"), "{}");
        write(root.path(), &format!("d{i}/.gitignore"), &list);
    }
    let result = detect(&root);
    // Five lists of about 200 KB fit in 1 MiB, the sixth and seventh do not.
    assert_eq!(
        manifests(&result),
        ["d6/skipme/package.json", "d7/skipme/package.json"],
        "{result}"
    );
}

/// Restores a folder's mode when dropped, so a failed assertion leaves a removable tree.
#[cfg(unix)]
struct Mode(std::path::PathBuf, u32);

#[cfg(unix)]
impl Drop for Mode {
    fn drop(&mut self) {
        use std::os::unix::fs::PermissionsExt;
        let _ = fs::set_permissions(&self.0, fs::Permissions::from_mode(self.1));
    }
}

/// REQ-3, REQ-7, REQ-8 (repair): a folder the walk cannot read is left out and the walk says
/// so — `complete` is false and one next action says the list is partial — while the readable
/// folders are still listed.
#[cfg(unix)]
#[test]
fn a_folder_that_cannot_be_read_makes_the_walk_incomplete() {
    use std::os::unix::fs::PermissionsExt;
    let root = TempDir::new().unwrap();
    let r = root.path();
    write(r, "apps/a/package.json", "{}");
    write(r, "apps/locked/package.json", "{}");
    let locked = r.join("apps/locked");
    let _restore = Mode(locked.clone(), 0o755);
    fs::set_permissions(&locked, fs::Permissions::from_mode(0o000)).unwrap();
    if fs::read_dir(&locked).is_ok() {
        eprintln!("skipped: this user reads a folder of mode 000");
        return;
    }
    let result = detect(&root);
    assert_eq!(manifests(&result), ["apps/a/package.json"], "{result}");
    assert_eq!(
        result["manifest_walk"],
        json!({"complete": false, "over_cap": 0}),
        "{result}"
    );
    assert_eq!(partial_lines(&result), 1, "{result}");

    // The same tree with the folder readable again is complete.
    fs::set_permissions(&locked, fs::Permissions::from_mode(0o755)).unwrap();
    let result = detect(&root);
    assert_eq!(
        manifests(&result),
        ["apps/a/package.json", "apps/locked/package.json"],
        "{result}"
    );
    assert_eq!(result["manifest_walk"]["complete"], json!(true), "{result}");
}
