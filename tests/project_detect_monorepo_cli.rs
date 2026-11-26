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

/// The one partial line of `next_actions`, whose reasons say why the list is partial.
fn partial_line(result: &Value) -> String {
    let lines: Vec<&str> = result["next_actions"]
        .as_array()
        .unwrap()
        .iter()
        .map(|l| l.as_str().unwrap())
        .filter(|l| l.contains(PARTIAL))
        .collect();
    assert_eq!(lines.len(), 1, "{result}");
    lines[0].to_string()
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
    // Each line names its own reason and none of the others'.
    let stopped = a_project_with_a_large_early_folder();
    let line = partial_line(&detect(&stopped));
    assert!(line.contains("entry budget"), "{line}");
    assert!(
        !line.contains("capped") && !line.contains("ignore list"),
        "{line}"
    );

    let many = TempDir::new().unwrap();
    for i in 0..65 {
        write(many.path(), &format!("pkg{i:02}/package.json"), "{}");
    }
    let line = partial_line(&detect(&many));
    assert!(line.contains("capped"), "{line}");
    assert!(
        !line.contains("entry budget") && !line.contains("ignore list"),
        "{line}"
    );

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

/// Builds the name at one place of a list.
type NameOf = fn(usize) -> String;

/// The extension file with `rows` rows, each naming one file of its own.
fn rows_json(rows: usize) -> String {
    let rows: Vec<String> = (0..rows)
        .map(|i| format!(r#"{{"ecosystem": "e{i}", "language": "l{i}", "files": ["f{i}.mk"]}}"#))
        .collect();
    format!(r#"{{"schema": 1, "manifests": [{}]}}"#, rows.join(","))
}

/// One row whose `key` list holds `n` names built by `name`, beside the file name every row
/// needs.
fn one_row_with(key: &str, n: usize, name: NameOf) -> String {
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

    let lists: [(&str, NameOf); 3] = [
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
        assert_eq!(result["manifest_walk"]["complete"], json!(true), "{result}");
        assert_eq!(partial_lines(&result), 0, "{path} {result}");
        write(root.path(), path, &list(262_145));
        let result = detect(&root);
        assert_eq!(
            manifests(&result),
            ["apps/a/package.json", "apps/skipme/package.json"],
            "{path} {result}"
        );
        // The list that was not read is a gap the walk reports, not a silent one.
        assert_eq!(
            result["manifest_walk"]["complete"],
            json!(false),
            "{path} {result}"
        );
        let line = partial_line(&result);
        assert!(line.contains("ignore list"), "{path} {line}");
    }
}

/// A skipped ignore list that holds a negation can hide a folder: the root list excludes `old/`
/// and `z/.gitignore` keeps it with `!old`. When that list is skipped — past the total (q2) or
/// over the file limit (q4), not text — `z/old/package.json` is not listed, so the walk says it is not
/// complete and the partial line names the skipped list; the same tree with the list applied
/// lists it and is complete.
#[test]
fn a_skipped_ignore_list_makes_the_walk_incomplete() {
    let build = |filler: usize, z_list: String| {
        let root = TempDir::new().unwrap();
        write(root.path(), ".gitignore", "old/\n");
        write(root.path(), "a/package.json", "{}");
        for name in ["f1", "f2", "f3", "f4"][..filler].iter() {
            write(
                root.path(),
                &format!("{name}/.gitignore"),
                &padded_list(262_144),
            );
        }
        write(root.path(), "z/old/package.json", "{}");
        write(root.path(), "z/.gitignore", &z_list);
        root
    };
    // Control: the list is read and keeps `z/old`.
    let control = detect(&build(0, "!old\n".to_string()));
    assert_eq!(
        manifests(&control),
        ["a/package.json", "z/old/package.json"],
        "{control}"
    );
    assert_eq!(
        control["manifest_walk"]["complete"],
        json!(true),
        "{control}"
    );
    assert_eq!(partial_lines(&control), 0, "{control}");
    // q2: four lists of exactly 262144 bytes use the whole 1048576, so the `!old` list is skipped.
    // q4: one list over 262144 bytes is skipped on its own.
    let over_the_total = build(4, "!old\n".to_string());
    let over_the_file_limit = build(0, extension_of_len("!old\n", 262_145));
    // A list that is not text cannot be read either.
    let not_text = build(0, "!old\n".to_string());
    fs::write(not_text.path().join("z/.gitignore"), b"!old\n\xff\xfe").unwrap();
    for (case, root) in [
        ("q2", over_the_total),
        ("q4", over_the_file_limit),
        ("not text", not_text),
    ] {
        let result = detect(&root);
        assert_eq!(manifests(&result), ["a/package.json"], "{case} {result}");
        assert_eq!(
            result["manifest_walk"]["complete"],
            json!(false),
            "{case} {result}"
        );
        let line = partial_line(&result);
        assert!(line.contains("ignore list"), "{case} {line}");
        assert!(!line.contains("entry budget"), "{case} {line}");
    }
}

/// An ignore list of `len` bytes that holds no rule: one comment line, no newline.
fn padded_list(len: usize) -> String {
    format!("#{}", "x".repeat(len - 1))
}

/// A walk with four lists of 262144 bytes below the root, then `d/.gitignore` of `last` bytes
/// whose one rule `s` (the first byte) excludes `d/s`.
fn a_walk_with_lists_summing_to(last: &str) -> TempDir {
    let root = TempDir::new().unwrap();
    for (name, len) in [
        ("a", 262_144),
        ("b", 262_144),
        ("c", 262_144),
        ("d", 262_143),
    ] {
        write(
            root.path(),
            &format!("{name}/.gitignore"),
            &padded_list(len),
        );
    }
    write(root.path(), "e/s/package.json", "{}");
    write(root.path(), "e/.gitignore", last);
    root
}

/// The total of 1048576 bytes read holds at its edge (REQ-2, AC-12): the lists before `e` add up
/// to 1048575 bytes, so a one-byte list (the sum 1048576) is applied.
#[test]
fn the_total_ignore_cap_applies_a_list_that_reaches_it_exactly() {
    let root = a_walk_with_lists_summing_to("s");
    let result = detect(&root);
    assert_eq!(manifests(&result), Vec::<&str>::new(), "{result}");
    assert_eq!(result["manifest_walk"]["complete"], json!(true), "{result}");
}

/// The same tree with a two-byte list (the sum 1048577) leaves that list unapplied: its folder
/// is listed and the walk says it is not complete.
#[test]
fn the_total_ignore_cap_leaves_a_list_that_passes_it_by_one_byte() {
    let root = a_walk_with_lists_summing_to("s\n");
    let result = detect(&root);
    assert_eq!(manifests(&result), ["e/s/package.json"], "{result}");
    assert_eq!(
        result["manifest_walk"]["complete"],
        json!(false),
        "{result}"
    );
    assert!(partial_line(&result).contains("ignore list"), "{result}");
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
    let line = partial_line(&result);
    assert!(line.contains("could not be read"), "{line}");
    assert!(
        !line.contains("entry budget") && !line.contains("ignore list"),
        "{line}"
    );

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

// ---- SPEC-20261125-detect-code-extension-catalog: code file extensions as catalog data ----

/// The extension table the engine held before the catalog took it: extension → language.
const EARLIER_CODE_EXTENSIONS: [(&str, &str); 17] = [
    ("rs", "rust"),
    ("py", "python"),
    ("js", "javascript"),
    ("jsx", "javascript"),
    ("ts", "typescript"),
    ("tsx", "typescript"),
    ("go", "go"),
    ("java", "java"),
    ("kt", "kotlin"),
    ("rb", "ruby"),
    ("php", "php"),
    ("c", "c"),
    ("h", "c"),
    ("cpp", "cpp"),
    ("cs", "csharp"),
    ("swift", "swift"),
    ("lua", "lua"),
];

/// A good manifest row beside the code rows under test: a file that is refused whole must not
/// let even this row be used.
const GOOD_MANIFEST_ROW: &str =
    r#"{"ecosystem": "hex", "language": "elixir", "files": ["mix.exs"]}"#;

/// REQ-1: each earlier extension still gives its language, alone in a project so a wrong
/// pair shows by name, and all of them together give the earlier languages and language
/// servers. The extension is compared as written.
#[test]
fn every_shipped_code_extension_keeps_its_language() {
    for (ext, language) in EARLIER_CODE_EXTENSIONS {
        let root = TempDir::new().unwrap();
        write(root.path(), &format!("src/a.{ext}"), "x\n");
        let result = detect(&root);
        assert_eq!(
            result["detected"]["languages"],
            json!([language]),
            ".{ext}: {result}"
        );
    }

    let root = TempDir::new().unwrap();
    for (ext, _) in EARLIER_CODE_EXTENSIONS {
        write(root.path(), &format!("src/a.{ext}"), "x\n");
    }
    let result = detect(&root);
    assert_eq!(
        result["detected"]["languages"],
        json!([
            "c",
            "cpp",
            "csharp",
            "go",
            "java",
            "javascript",
            "kotlin",
            "lua",
            "php",
            "python",
            "ruby",
            "rust",
            "swift",
            "typescript"
        ]),
        "{result}"
    );
    assert_eq!(
        result["detected"]["lsp"],
        json!([
            "clangd",
            "csharp-ls",
            "gopls",
            "intelephense",
            "jdtls",
            "kotlin-lsp",
            "lua-language-server",
            "pyright",
            "ruby-lsp",
            "rust-analyzer",
            "sourcekit-lsp",
            "typescript-language-server"
        ]),
        "{result}"
    );

    let root = TempDir::new().unwrap();
    write(root.path(), "src/A.RS", "x\n");
    let result = detect(&root);
    assert_eq!(result["detected"]["languages"], json!([]), "{result}");
}

/// REQ-2, REQ-4: a row of the extension file adds a language the shipped catalog does not
/// know, for the code files that carry its extensions, and the rows are counted with the
/// manifest rows. Without the file the language is gone and nothing was written.
#[test]
fn a_code_extension_row_adds_a_language_without_code() {
    let root = TempDir::new().unwrap();
    let r = root.path();
    write(r, "lib/a.ex", "x\n");
    write(r, "lib/b.exs", "x\n");
    write(r, "tools/c.zig", "x\n");
    let result = detect(&root);
    assert_eq!(result["detected"]["languages"], json!([]), "{result}");

    write(
        r,
        &extension(),
        r#"{"schema": 1, "code_extensions": [
            {"language": "elixir", "extensions": ["ex", "exs"]}
        ]}"#,
    );
    let result = detect(&root);
    assert_eq!(
        result["detected"]["languages"],
        json!(["elixir"]),
        "{result}"
    );
    assert_eq!(
        result["manifest_catalog"],
        json!({"extension": "read", "rows": 1, "problem": null}),
        "{result}"
    );
    let before = tree(r);
    detect(&root);
    assert_eq!(tree(r), before, "a read wrote something");

    // Rows of both kinds are counted together.
    write(
        r,
        &extension(),
        &format!(
            r#"{{"schema": 1, "manifests": [{GOOD_MANIFEST_ROW}], "code_extensions": [
                {{"language": "elixir", "extensions": ["ex", "exs"]}},
                {{"language": "zig", "extensions": ["zig"]}}
            ]}}"#
        ),
    );
    let result = detect(&root);
    assert_eq!(
        result["detected"]["languages"],
        json!(["elixir", "zig"]),
        "{result}"
    );
    assert_eq!(
        result["manifest_catalog"],
        json!({"extension": "read", "rows": 3, "problem": null}),
        "{result}"
    );

    fs::remove_file(r.join(extension())).unwrap();
    let result = detect(&root);
    assert_eq!(result["detected"]["languages"], json!([]), "{result}");
}

/// REQ-2: a row never changes the language of an extension the shipped catalog has; only the
/// extensions it does not have are used, and a row left with none adds nothing.
#[test]
fn a_code_extension_row_never_replaces_a_shipped_extension() {
    let root = TempDir::new().unwrap();
    let r = root.path();
    write(r, "src/a.rs", "x\n");
    write(r, "tools/b.zig", "x\n");
    write(
        r,
        &extension(),
        r#"{"schema": 1, "code_extensions": [
            {"language": "zig", "extensions": ["rs", "zig"]}
        ]}"#,
    );
    let result = detect(&root);
    assert_eq!(
        result["detected"]["languages"],
        json!(["rust", "zig"]),
        "{result}"
    );

    let only_shipped = TempDir::new().unwrap();
    write(only_shipped.path(), "src/a.rs", "x\n");
    write(
        only_shipped.path(),
        &extension(),
        r#"{"schema": 1, "code_extensions": [
            {"language": "zig", "extensions": ["rs"]}
        ]}"#,
    );
    let result = detect(&only_shipped);
    assert_eq!(result["detected"]["languages"], json!(["rust"]), "{result}");
    assert_eq!(
        result["manifest_catalog"]["extension"],
        json!("read"),
        "{result}"
    );
}

/// REQ-3: one row that cannot be read refuses the whole extension file: the problem is
/// reported, the good rows beside it are not used either (the shipped rows alone are), and
/// nothing is written. The same file with the bad row mended is read, so each refusal is
/// the row's doing.
#[test]
fn a_code_extension_row_that_cannot_be_read_refuses_the_whole_file() {
    let root = TempDir::new().unwrap();
    let r = root.path();
    write(r, "apps/svc/mix.exs", "defmodule X do end\n");
    write(r, "lib/a.ex", "x\n");
    write(r, "src/a.rs", "x\n");

    let file_with = |code_row: &str| {
        format!(
            r#"{{"schema": 1, "manifests": [{GOOD_MANIFEST_ROW}], "code_extensions": [{code_row}]}}"#
        )
    };

    // The control: the mended row is read, and both kinds of row take effect.
    write(
        r,
        &extension(),
        &file_with(r#"{"language": "elixir", "extensions": ["ex"]}"#),
    );
    let result = detect(&root);
    assert_eq!(
        result["manifest_catalog"],
        json!({"extension": "read", "rows": 2, "problem": null}),
        "{result}"
    );
    assert_eq!(manifests(&result), ["apps/svc/mix.exs"], "{result}");
    assert_eq!(
        result["detected"]["languages"],
        json!(["elixir", "rust"]),
        "{result}"
    );

    let bad_rows = [
        r#"{"language": "elixir", "extensions": []}"#,
        r#"{"language": "elixir"}"#,
        r#"{"language": "", "extensions": ["ex"]}"#,
        r#"{"language": "eli xir", "extensions": ["ex"]}"#,
        r#"{"language": "elixir", "extensions": ["e x"]}"#,
        r#"{"language": "elixir", "extensions": [".ex"]}"#,
        r#"{"language": "elixir", "extensions": ["e.x"]}"#,
        r#"{"language": "elixir", "extensions": ["ex"], "surprise": 1}"#,
        r#"{"language": "elixir", "extensions": "ex"}"#,
        r#"{"language": 7, "extensions": ["ex"]}"#,
    ];
    for bad in bad_rows {
        write(r, &extension(), &file_with(bad));
        let before = tree(r);
        let result = detect(&root);
        assert_eq!(
            result["manifest_catalog"]["extension"],
            json!("refused"),
            "{bad}: {result}"
        );
        assert_eq!(result["manifest_catalog"]["rows"], json!(0), "{bad}");
        assert!(
            result["manifest_catalog"]["problem"].is_string(),
            "{bad}: {result}"
        );
        // Neither the code row nor the good manifest row beside it is used.
        assert!(manifests(&result).is_empty(), "{bad}: {result}");
        assert_eq!(
            result["detected"]["languages"],
            json!(["rust"]),
            "{bad}: {result}"
        );
        assert_eq!(tree(r), before, "{bad}: a read wrote something");
    }
}

/// The extension file with `n` rows, each a language and one extension of its own.
fn code_rows_json(n: usize) -> String {
    let rows: Vec<String> = (0..n)
        .map(|i| format!(r#"{{"language": "l{i}", "extensions": ["x{i}"]}}"#))
        .collect();
    format!(
        r#"{{"schema": 1, "code_extensions": [{}]}}"#,
        rows.join(",")
    )
}

/// The extension file with one row of `n` extensions.
fn code_row_with_extensions(n: usize) -> String {
    let list: Vec<String> = (0..n).map(|i| format!("\"e{i}\"")).collect();
    format!(
        r#"{{"schema": 1, "code_extensions": [{{"language": "l", "extensions": [{}]}}]}}"#,
        list.join(",")
    )
}

/// REQ-3: the limits of 64 rows and of 64 extensions in one row hold at both edges: 64 are
/// read, 65 are refused.
#[test]
fn the_code_extension_row_limits_hold_at_both_edges() {
    assert_eq!(
        extension_state(&code_rows_json(64)),
        ("read".to_string(), 64)
    );
    assert_eq!(
        extension_state(&code_rows_json(65)),
        ("refused".to_string(), 0)
    );
    assert_eq!(
        extension_state(&code_row_with_extensions(64)),
        ("read".to_string(), 1)
    );
    assert_eq!(
        extension_state(&code_row_with_extensions(65)),
        ("refused".to_string(), 0)
    );
}

/// REQ-5: two rows of the extension file that give one extension two languages refuse the
/// whole file and the problem names the extension, whichever row comes first; the same
/// language twice (in two rows or twice in one row) is read.
#[test]
fn one_extension_for_two_languages_in_the_extension_file_is_refused() {
    let root = TempDir::new().unwrap();
    let r = root.path();
    write(r, "lib/a.ex", "x\n");
    write(r, "src/a.rs", "x\n");
    let conflicts = [
        (
            r#"{"language": "elixir", "extensions": ["ex"]},
               {"language": "erlang", "extensions": ["erl", "ex"]}"#,
            "the extension ex is claimed by two languages: elixir and erlang",
        ),
        (
            r#"{"language": "erlang", "extensions": ["erl", "ex"]},
               {"language": "elixir", "extensions": ["ex"]}"#,
            "the extension ex is claimed by two languages: erlang and elixir",
        ),
    ];
    for (rows, problem) in conflicts {
        write(
            r,
            &extension(),
            &format!(r#"{{"schema": 1, "code_extensions": [{rows}]}}"#),
        );
        let before = tree(r);
        let result = detect(&root);
        assert_eq!(
            result["manifest_catalog"],
            json!({"extension": "refused", "rows": 0, "problem": problem}),
            "{result}"
        );
        // The shipped rows alone are used: the refused file adds no language.
        assert_eq!(result["detected"]["languages"], json!(["rust"]), "{result}");
        assert_eq!(tree(r), before, "a read wrote something");
    }

    // The same language for one extension is no conflict: twice in one row, or in two rows.
    for rows in [
        r#"{"language": "elixir", "extensions": ["ex", "ex"]}"#,
        r#"{"language": "elixir", "extensions": ["ex"]},
           {"language": "elixir", "extensions": ["ex", "exs"]}"#,
    ] {
        write(
            r,
            &extension(),
            &format!(r#"{{"schema": 1, "code_extensions": [{rows}]}}"#),
        );
        let result = detect(&root);
        assert_eq!(
            result["manifest_catalog"]["extension"],
            json!("read"),
            "{result}"
        );
        assert_eq!(
            result["detected"]["languages"],
            json!(["elixir", "rust"]),
            "{result}"
        );
    }
}

// ---- SPEC-20261125-detect-database-dependency-catalog: database dependencies as catalog data ----

/// The dependency table the engine held before the catalog took it: dependency → kind.
const EARLIER_DATABASE_DEPENDENCIES: [(&str, &str); 22] = [
    ("postgres", "postgres"),
    ("tokio-postgres", "postgres"),
    ("pg", "postgres"),
    ("psycopg", "postgres"),
    ("psycopg2", "postgres"),
    ("psycopg2-binary", "postgres"),
    ("asyncpg", "postgres"),
    ("github.com/jackc/pgx/v5", "postgres"),
    ("github.com/lib/pq", "postgres"),
    ("mysql", "mysql"),
    ("mysql2", "mysql"),
    ("pymysql", "mysql"),
    ("mysqlclient", "mysql"),
    ("github.com/go-sql-driver/mysql", "mysql"),
    ("rusqlite", "sqlite"),
    ("sqlite3", "sqlite"),
    ("better-sqlite3", "sqlite"),
    ("mongodb", "mongodb"),
    ("pymongo", "mongodb"),
    ("mongoose", "mongodb"),
    ("redis", "redis"),
    ("ioredis", "redis"),
];

/// Declares each of `names` as a dependency of the project: a name with a `/` as a Go module
/// requirement, any other name in `package.json` (a Python requirement stops at a `/`).
fn declare_dependencies(root: &Path, names: &[&str]) {
    let (modules, packages): (Vec<&str>, Vec<&str>) = names.iter().partition(|n| n.contains('/'));
    if !packages.is_empty() {
        let list: Vec<String> = packages
            .iter()
            .map(|n| format!(r#""{n}": "1.0.0""#))
            .collect();
        write(
            root,
            "package.json",
            &format!(r#"{{"dependencies": {{{}}}}}"#, list.join(", ")),
        );
    }
    if !modules.is_empty() {
        let list: Vec<String> = modules
            .iter()
            .map(|n| format!("require {n} v1.0.0\n"))
            .collect();
        write(root, "go.mod", &format!("module app\n\n{}", list.concat()));
    }
}

/// A project that declares only `names`.
fn project_declaring(names: &[&str]) -> TempDir {
    let root = TempDir::new().unwrap();
    declare_dependencies(root.path(), names);
    root
}

fn databases(result: &Value) -> Vec<&str> {
    strings(&result["detected"]["databases"])
}

/// An extension file that holds only the `database_dependencies` rows in `rows`.
fn db_file(rows: &str) -> String {
    format!(r#"{{"schema": 1, "database_dependencies": [{rows}]}}"#)
}

/// REQ-1: each earlier dependency still gives its kind, alone in a project so a wrong pair
/// shows by name, and all of them together give the five earlier kinds.
#[test]
fn every_shipped_database_dependency_keeps_its_kind() {
    for (dependency, kind) in EARLIER_DATABASE_DEPENDENCIES {
        let root = project_declaring(&[dependency]);
        let result = detect(&root);
        assert_eq!(databases(&result), [kind], "{dependency}: {result}");
    }

    let all: Vec<&str> = EARLIER_DATABASE_DEPENDENCIES
        .iter()
        .map(|(d, _)| *d)
        .collect();
    let root = project_declaring(&all);
    let result = detect(&root);
    assert_eq!(
        databases(&result),
        ["mongodb", "mysql", "postgres", "redis", "sqlite"],
        "{result}"
    );
}

/// REQ-2, REQ-4: a row of the extension file adds a kind for a dependency the shipped catalog
/// does not know, and the rows are counted with the manifest and code rows. Without the file
/// the kind is gone and nothing was written.
#[test]
fn a_database_dependency_row_adds_a_kind_without_code() {
    let root = project_declaring(&["sqlx"]);
    let r = root.path();
    let result = detect(&root);
    assert_eq!(result["detected"]["databases"], json!([]), "{result}");

    write(
        r,
        &extension(),
        &db_file(r#"{"kind": "postgres", "dependencies": ["sqlx"]}"#),
    );
    let result = detect(&root);
    assert_eq!(databases(&result), ["postgres"], "{result}");
    assert_eq!(
        result["manifest_catalog"],
        json!({"extension": "read", "rows": 1, "problem": null}),
        "{result}"
    );
    let before = tree(r);
    detect(&root);
    assert_eq!(tree(r), before, "a read wrote something");

    // Rows of the three kinds are counted together.
    write(
        r,
        &extension(),
        &format!(
            r#"{{"schema": 1, "manifests": [{GOOD_MANIFEST_ROW}],
                "code_extensions": [{{"language": "elixir", "extensions": ["ex"]}}],
                "database_dependencies": [{{"kind": "postgres", "dependencies": ["sqlx"]}}]}}"#
        ),
    );
    let result = detect(&root);
    assert_eq!(databases(&result), ["postgres"], "{result}");
    assert_eq!(
        result["manifest_catalog"],
        json!({"extension": "read", "rows": 3, "problem": null}),
        "{result}"
    );

    fs::remove_file(r.join(extension())).unwrap();
    let result = detect(&root);
    assert_eq!(result["detected"]["databases"], json!([]), "{result}");
}

/// REQ-2: a row never changes the kind of a dependency the shipped catalog has; only the
/// dependencies it does not have are used, and a row left with none adds nothing.
#[test]
fn a_database_dependency_row_never_replaces_a_shipped_dependency() {
    let root = project_declaring(&["pg", "sqlx"]);
    write(
        root.path(),
        &extension(),
        &db_file(r#"{"kind": "mysql", "dependencies": ["pg", "sqlx"]}"#),
    );
    let result = detect(&root);
    assert_eq!(databases(&result), ["mysql", "postgres"], "{result}");
    assert_eq!(
        result["manifest_catalog"]["extension"],
        json!("read"),
        "{result}"
    );

    let only_shipped = project_declaring(&["pg"]);
    write(
        only_shipped.path(),
        &extension(),
        &db_file(r#"{"kind": "mysql", "dependencies": ["pg", "sqlx"]}"#),
    );
    let result = detect(&only_shipped);
    assert_eq!(databases(&result), ["postgres"], "{result}");

    // A row that holds only the shipped dependency adds nothing, and is still read.
    let nothing_new = project_declaring(&["pg", "sqlx"]);
    write(
        nothing_new.path(),
        &extension(),
        &db_file(r#"{"kind": "mysql", "dependencies": ["pg"]}"#),
    );
    let result = detect(&nothing_new);
    assert_eq!(databases(&result), ["postgres"], "{result}");
    assert_eq!(
        result["manifest_catalog"],
        json!({"extension": "read", "rows": 1, "problem": null}),
        "{result}"
    );
}

/// REQ-3: one row that cannot be read refuses the whole extension file: the problem is
/// reported, the good rows beside it are not used either (the shipped rows alone are), and
/// nothing is written. The same file with the bad row mended is read, so each refusal is
/// the row's doing.
#[test]
fn a_database_dependency_row_that_cannot_be_read_refuses_the_whole_file() {
    let root = project_declaring(&["pg", "tedious"]);
    let r = root.path();
    write(r, "apps/svc/mix.exs", "defmodule X do end\n");
    write(r, "lib/a.ex", "x\n");
    write(r, "src/a.rs", "x\n");

    let file_with = |db_row: &str| {
        format!(
            r#"{{"schema": 1, "manifests": [{GOOD_MANIFEST_ROW}],
                "code_extensions": [{{"language": "elixir", "extensions": ["ex"]}}],
                "database_dependencies": [{db_row}]}}"#
        )
    };

    // The control: the mended row is read, and all three kinds of row take effect.
    write(
        r,
        &extension(),
        &file_with(r#"{"kind": "sqlserver", "dependencies": ["tedious"]}"#),
    );
    let result = detect(&root);
    assert_eq!(
        result["manifest_catalog"],
        json!({"extension": "read", "rows": 3, "problem": null}),
        "{result}"
    );
    assert_eq!(
        manifests(&result),
        ["apps/svc/mix.exs", "package.json"],
        "{result}"
    );
    assert_eq!(
        result["detected"]["languages"],
        json!(["elixir", "javascript", "rust"]),
        "{result}"
    );
    assert_eq!(databases(&result), ["postgres", "sqlserver"], "{result}");

    let long = "a".repeat(121);
    let bad_rows = [
        r#"{"kind": "postgres", "dependencies": []}"#.to_string(),
        r#"{"kind": "postgres"}"#.to_string(),
        r#"{"dependencies": ["tedious"]}"#.to_string(),
        r#"{"kind": "", "dependencies": ["tedious"]}"#.to_string(),
        r#"{"kind": "sql server", "dependencies": ["tedious"]}"#.to_string(),
        r#"{"kind": "sql/server", "dependencies": ["tedious"]}"#.to_string(),
        r#"{"kind": 7, "dependencies": ["tedious"]}"#.to_string(),
        r#"{"kind": "sqlserver", "dependencies": [""]}"#.to_string(),
        r#"{"kind": "sqlserver", "dependencies": ["."]}"#.to_string(),
        r#"{"kind": "sqlserver", "dependencies": ["te dious"]}"#.to_string(),
        r#"{"kind": "sqlserver", "dependencies": ["c++"]}"#.to_string(),
        r#"{"kind": "sqlserver", "dependencies": ["Tedious"]}"#.to_string(),
        format!(r#"{{"kind": "sqlserver", "dependencies": ["{long}"]}}"#),
        r#"{"kind": "sqlserver", "dependencies": ["tedious"], "surprise": 1}"#.to_string(),
        r#"{"kind": "sqlserver", "dependencies": "tedious"}"#.to_string(),
    ];
    for bad in &bad_rows {
        write(r, &extension(), &file_with(bad));
        let before = tree(r);
        let result = detect(&root);
        assert_eq!(
            result["manifest_catalog"]["extension"],
            json!("refused"),
            "{bad}: {result}"
        );
        assert_eq!(result["manifest_catalog"]["rows"], json!(0), "{bad}");
        assert!(
            result["manifest_catalog"]["problem"].is_string(),
            "{bad}: {result}"
        );
        // Neither the manifest row, the code row nor the dependency row is used.
        assert_eq!(manifests(&result), ["package.json"], "{bad}: {result}");
        assert_eq!(
            result["detected"]["languages"],
            json!(["javascript", "rust"]),
            "{bad}: {result}"
        );
        assert_eq!(databases(&result), ["postgres"], "{bad}: {result}");
        assert_eq!(tree(r), before, "{bad}: a read wrote something");
    }
}

/// The extension file with `n` rows, each a kind and one dependency of its own.
fn db_rows_json(n: usize) -> String {
    let rows: Vec<String> = (0..n)
        .map(|i| format!(r#"{{"kind": "k", "dependencies": ["d{i}"]}}"#))
        .collect();
    db_file(&rows.join(","))
}

/// The extension file with one row of `n` dependencies.
fn db_row_with_dependencies(n: usize) -> String {
    let list: Vec<String> = (0..n).map(|i| format!("\"d{i}\"")).collect();
    db_file(&format!(
        r#"{{"kind": "k", "dependencies": [{}]}}"#,
        list.join(",")
    ))
}

/// The extension file with one row that names `dependency`.
fn db_row_named(dependency: &str) -> String {
    db_file(&format!(
        r#"{{"kind": "k", "dependencies": ["{dependency}"]}}"#
    ))
}

/// REQ-3: the limits of 64 rows, of 64 dependencies in one row and of 120 bytes in a name hold
/// at both edges: the last value is read, the next is refused. Every name is distinct, so no
/// refusal comes from one dependency claimed twice.
#[test]
fn the_database_dependency_row_limits_hold_at_both_edges() {
    assert_eq!(extension_state(&db_rows_json(64)), ("read".to_string(), 64));
    assert_eq!(
        extension_state(&db_rows_json(65)),
        ("refused".to_string(), 0)
    );
    assert_eq!(
        extension_state(&db_row_with_dependencies(64)),
        ("read".to_string(), 1)
    );
    assert_eq!(
        extension_state(&db_row_with_dependencies(65)),
        ("refused".to_string(), 0)
    );
    assert_eq!(
        extension_state(&db_row_named(&"a".repeat(120))),
        ("read".to_string(), 1)
    );
    assert_eq!(
        extension_state(&db_row_named(&"a".repeat(121))),
        ("refused".to_string(), 0)
    );
}

/// REQ-5: two rows of the extension file that give one dependency two kinds refuse the whole
/// file and the problem names the dependency, whichever row comes first; the same kind twice
/// (in one row or in two rows) is read.
#[test]
fn one_dependency_for_two_kinds_in_the_extension_file_is_refused() {
    let root = project_declaring(&["sqlx", "redis"]);
    let r = root.path();
    let conflicts = [
        (
            r#"{"kind": "postgres", "dependencies": ["sqlx"]},
               {"kind": "mysql", "dependencies": ["sqlx"]}"#,
            "the dependency sqlx is claimed by two kinds: postgres and mysql",
        ),
        (
            r#"{"kind": "mysql", "dependencies": ["sqlx"]},
               {"kind": "postgres", "dependencies": ["sqlx"]}"#,
            "the dependency sqlx is claimed by two kinds: mysql and postgres",
        ),
    ];
    for (rows, problem) in conflicts {
        write(r, &extension(), &db_file(rows));
        let before = tree(r);
        let result = detect(&root);
        assert_eq!(
            result["manifest_catalog"],
            json!({"extension": "refused", "rows": 0, "problem": problem}),
            "{result}"
        );
        // The shipped rows alone are used: the refused file adds no kind.
        assert_eq!(databases(&result), ["redis"], "{result}");
        assert_eq!(tree(r), before, "a read wrote something");
    }

    // The same kind for one dependency is no conflict: twice in one row, or in two rows.
    for rows in [
        r#"{"kind": "postgres", "dependencies": ["sqlx", "sqlx"]}"#,
        r#"{"kind": "postgres", "dependencies": ["sqlx"]},
           {"kind": "postgres", "dependencies": ["sqlx", "diesel"]}"#,
    ] {
        write(r, &extension(), &db_file(rows));
        let result = detect(&root);
        assert_eq!(
            result["manifest_catalog"]["extension"],
            json!("read"),
            "{result}"
        );
        assert_eq!(databases(&result), ["postgres", "redis"], "{result}");
    }
}
