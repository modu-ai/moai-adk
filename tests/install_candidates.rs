//! #2791 AC-2: `install-inspect` lists every `moai` on PATH and in the Go install places and
//! tells the Rust CLI from the Go `moai` by the `--version` first line — report only.
//! Temporary HOME, PATH and project folders only; the real ones are never read.

#![cfg(unix)]

mod common;

use moai_cli::tree_digest;
use serde_json::{json, Value};
use std::{
    fs,
    path::{Path, PathBuf},
    process::{Command, Stdio},
};
use tempfile::TempDir;

const EXE: &str = common::MOAI_BIN;

fn script(dir: &Path, body: &str) -> PathBuf {
    fs::create_dir_all(dir).unwrap();
    let p = dir.join("moai");
    common::write_executable(&p, body);
    p
}

fn inspect(home: &Path, path_env: &str, envs: &[(&str, &Path)]) -> Value {
    let mut cmd = Command::new(EXE);
    cmd.args(["install-inspect", "--json"])
        .env("HOME", home)
        .env("PATH", path_env)
        .env_remove("CLAUDE_PROJECT_DIR")
        .env_remove("GOBIN")
        .current_dir(home)
        .stdin(Stdio::null());
    for (k, v) in envs {
        cmd.env(k, v);
    }
    let out = cmd.output().unwrap();
    let v: Value = serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("{}", String::from_utf8_lossy(&out.stderr)));
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"]["moai_candidates"].clone()
}

fn join(dirs: &[&Path]) -> String {
    std::env::join_paths(dirs).unwrap().into_string().unwrap()
}

#[test]
fn every_moai_is_listed_in_order_and_told_apart_without_writing() {
    let d = TempDir::new().unwrap();
    let root = fs::canonicalize(d.path()).unwrap();
    let home = root.join("home");
    let (p1, p2, proj) = (root.join("p1"), root.join("p2"), root.join("proj"));
    script(&p1, "#!/bin/sh\necho 'moai-adk 3.2.0'\n");
    fs::create_dir_all(&p2).unwrap();
    common::copy_executable(Path::new(EXE), &p2.join("moai"));
    script(&proj.join(".moai/bin"), "#!/bin/sh\necho 'moai 1.4.0'\n");
    script(&home.join("go/bin"), "#!/bin/sh\nexit 0\n");
    script(
        &home.join(".local/bin"),
        "#!/bin/sh\necho 'something else 1.0'\n",
    );
    // A `moai` that is not executable is no candidate.
    let p3 = root.join("p3");
    fs::create_dir_all(&p3).unwrap();
    fs::write(p3.join("moai"), "#!/bin/sh\necho 'moai-adk 9'\n").unwrap();

    let before = tree_digest(&root, &[]).unwrap();
    // GOBIN names p1 again: one item, both sources.
    let c = inspect(
        &home,
        &join(&[&p3, &p1, &p2]),
        &[("CLAUDE_PROJECT_DIR", &proj), ("GOBIN", &p1)],
    );
    assert_eq!(tree_digest(&root, &[]).unwrap(), before, "report only");

    let items = c["items"].as_array().unwrap();
    let rows: Vec<(String, Value, Value, Value, String)> = items
        .iter()
        .map(|i| {
            let path = i["path"].as_str().unwrap();
            let rel = Path::new(path).strip_prefix(&root).unwrap();
            (
                rel.display().to_string(),
                i["sources"].clone(),
                i["path_order"].clone(),
                i["first_on_path"].clone(),
                i["generation"].as_str().unwrap().to_string(),
            )
        })
        .collect();
    assert_eq!(
        rows,
        vec![
            (
                "p1/moai".into(),
                json!(["path", "gobin"]),
                json!(1),
                json!(true),
                "go".into()
            ),
            (
                "p2/moai".into(),
                json!(["path"]),
                json!(2),
                json!(false),
                "rust".into()
            ),
            (
                "proj/.moai/bin/moai".into(),
                json!(["claude_project_dir"]),
                json!(null),
                json!(false),
                "not_probed".into()
            ),
            (
                "home/go/bin/moai".into(),
                json!(["go_bin"]),
                json!(null),
                json!(false),
                "unknown".into()
            ),
            (
                "home/.local/bin/moai".into(),
                json!(["local_bin"]),
                json!(null),
                json!(false),
                "unknown".into()
            ),
        ]
    );
    assert_eq!(items[0]["version_line"], json!("moai-adk 3.2.0"));
    assert!(items[1]["version_line"]
        .as_str()
        .unwrap()
        .starts_with("moai-cli "));
    assert_eq!(items[3]["version_line"], json!(null), "no output");
    assert!(items[0]["sha256"].as_str().unwrap().starts_with("sha256:"));
    assert_eq!(c["more"], json!(0));
    assert_eq!(c["first_on_path_generation"], json!("go"));

    // With the Rust copy first on PATH the summary follows it.
    let c = inspect(&home, &join(&[&p2, &p1]), &[]);
    assert_eq!(c["first_on_path_generation"], json!("rust"));
    // Nothing anywhere: an empty list, no summary.
    let empty = root.join("empty");
    fs::create_dir_all(&empty).unwrap();
    let bare = root.join("bare");
    fs::create_dir_all(&bare).unwrap();
    let c = inspect(&bare, &join(&[&empty]), &[]);
    assert_eq!(c["items"], json!([]));
    assert_eq!(c["first_on_path_generation"], json!(null));
}

/// Bounds as literals: the first line is cut at 200 characters (200 kept, 201 cut), only
/// the first line is read, and candidates beyond 64 are counted, not probed.
#[test]
fn the_line_and_candidate_bounds() {
    let d = TempDir::new().unwrap();
    let root = fs::canonicalize(d.path()).unwrap();
    let home = root.join("home");
    fs::create_dir_all(&home).unwrap();
    let a = root.join("a");
    let b = root.join("b");
    let line200 = format!("moai-adk {}", "x".repeat(191));
    assert_eq!(line200.chars().count(), 200);
    script(
        &a,
        &format!("#!/bin/sh\necho '{line200}'\necho 'moai-cli second'\n"),
    );
    script(&b, &format!("#!/bin/sh\necho '{line200}y'\n"));
    let c = inspect(&home, &join(&[&a, &b]), &[]);
    assert_eq!(c["items"][0]["version_line"], json!(line200));
    assert_eq!(c["items"][0]["generation"], json!("go"), "first line only");
    assert_eq!(c["items"][1]["version_line"], json!(line200), "201st cut");

    for n in [64usize, 65] {
        let base = root.join(format!("many{n}"));
        let dirs: Vec<PathBuf> = (0..n).map(|i| base.join(i.to_string())).collect();
        for dir in &dirs {
            script(dir, "#!/bin/sh\nexit 0\n");
        }
        let refs: Vec<&Path> = dirs.iter().map(PathBuf::as_path).collect();
        let c = inspect(&home, &join(&refs), &[]);
        assert_eq!(c["items"].as_array().unwrap().len(), 64, "{n}");
        assert_eq!(c["more"], json!(n - 64), "{n}");
    }
}

/// #2800 audit B1 (leader decision on #2791): a `moai` inside the project tree is repository
/// content — it is reported with its path and sha256 but never run.
#[test]
fn a_project_tree_moai_is_reported_never_run() {
    let d = TempDir::new().unwrap();
    let root = fs::canonicalize(d.path()).unwrap();
    let home = root.join("home");
    let proj = root.join("proj");
    fs::create_dir_all(&home).unwrap();
    let marker = root.join("EXECUTED");
    script(
        &proj.join(".moai/bin"),
        &format!(
            "#!/bin/sh\nenv > '{}'\necho 'moai-adk 9.9.9'\n",
            marker.display()
        ),
    );
    let empty = root.join("empty");
    fs::create_dir_all(&empty).unwrap();
    let c = inspect(&home, &join(&[&empty]), &[("CLAUDE_PROJECT_DIR", &proj)]);
    assert!(!marker.exists(), "the project's moai ran");
    let item = &c["items"][0];
    assert_eq!(item["sources"], json!(["claude_project_dir"]), "{c}");
    assert_eq!(item["probed"], json!(false), "{item}");
    assert_eq!(item["version_line"], json!(null));
    assert_eq!(item["generation"], json!("not_probed"));
    assert!(item["sha256"].as_str().unwrap().starts_with("sha256:"));
}

/// #2800 audit B1: a probed `moai` runs with a cleared environment — no token variable of
/// this process reaches it — and outside the project and HOME folders.
#[test]
fn a_probe_gets_no_token_variable_and_runs_outside_the_project() {
    let d = TempDir::new().unwrap();
    let root = fs::canonicalize(d.path()).unwrap();
    let home = root.join("home");
    let pathbin = root.join("pathbin");
    fs::create_dir_all(&home).unwrap();
    script(
        &pathbin,
        "#!/bin/sh\necho \"moai-adk 1.0 token=[${ANTHROPIC_AUTH_TOKEN}${PROBE_TEST_TOKEN}] cwd=[$(pwd)]\"\n",
    );
    let c = inspect(
        &home,
        &join(&[&pathbin, Path::new("/usr/bin"), Path::new("/bin")]),
        &[
            ("ANTHROPIC_AUTH_TOKEN", Path::new("dummy-not-a-real-key")),
            ("PROBE_TEST_TOKEN", Path::new("dummy-not-a-real-key")),
        ],
    );
    let item = &c["items"][0];
    assert_eq!(item["probed"], json!(true), "{item}");
    let line = item["version_line"].as_str().unwrap();
    assert!(
        line.contains("token=[]"),
        "a token reached the probe: {line}"
    );
    assert!(
        !line.contains(home.to_str().unwrap()),
        "ran in HOME: {line}"
    );
    assert_eq!(item["generation"], json!("go"));
}

/// #2791 pre-audit 6865400825 B1: the 5 s limit holds even when a shell wrapper's own child
/// keeps stdout open — a wrapper stuck in `sleep 60`, and one that prints its line and exits
/// while a background child still holds the pipe. Both probes end near the limit (not after
/// 60 s); the first reports nothing, the second keeps the line printed before the limit.
#[test]
fn the_probe_limit_holds_when_a_wrapper_child_keeps_stdout_open() {
    let d = TempDir::new().unwrap();
    let root = fs::canonicalize(d.path()).unwrap();
    let home = root.join("home");
    fs::create_dir_all(&home).unwrap();
    let stuck = root.join("stuck");
    let left = root.join("left");
    script(&stuck, "#!/bin/sh\nsleep 60\n");
    script(&left, "#!/bin/sh\n(sleep 60) &\necho 'moai-adk 1.0'\n");
    let started = std::time::Instant::now();
    let c = inspect(&home, &join(&[&stuck, &left]), &[]);
    let took = started.elapsed();
    // Three probes of at most 5 s each (`stuck` is also the first PATH `moai` of
    // `path_resolution`), plus start-up; far below the 60 s each child would hold the pipe.
    assert!(took < std::time::Duration::from_secs(25), "took {took:?}");
    assert_eq!(c["items"][0]["version_line"], json!(null), "{c}");
    assert_eq!(c["items"][0]["generation"], json!("unknown"));
    assert_eq!(c["items"][1]["version_line"], json!("moai-adk 1.0"), "{c}");
    assert_eq!(c["items"][1]["generation"], json!("go"));
}
