//! SPEC-20261126-check-affected-trailers-flaky (AC-4, AC-5): the flaky ledger is written under the
//! `local` folder of the project's current state folder and never into the earlier state folder,
//! the answer of `state_root::dir` is the same after a record, and `check flaky` counts the new
//! ledger and an existing earlier ledger together while only reading the earlier one. Temporary
//! projects only; the folder names come from the `state_root` constants.
#![cfg(unix)]
mod common;

use common::git::git_ok;
use moai_cli::check_flaky::{ledger_path, record, report, TestName};
use moai_cli::state_root::{classify, dir, EARLIER_PROJECT_DIR, LOCAL_DIR, MOAI_DIR};
use std::collections::BTreeMap;
use std::fs;
use std::path::{Path, PathBuf};

const LEDGER: &str = "check-flaky.jsonl";
const HEAD: &str = "0123456789abcdef0123456789abcdef01234567";

/// The two project shapes of AC-4.
#[derive(Clone, Copy, Debug)]
enum Layout {
    /// No state folder of either name.
    Empty,
    /// The earlier folder holds files; the new folder holds only a profile file and no marker (the
    /// shape of a worktree of a repository that tracks its profile and SPECs).
    EarlierPresent,
}

fn write(path: &Path, bytes: &[u8]) {
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, bytes).unwrap();
}

fn project(layout: Layout) -> tempfile::TempDir {
    let tmp = tempfile::tempdir().unwrap();
    if let Layout::EarlierPresent = layout {
        let earlier = tmp.path().join(EARLIER_PROJECT_DIR);
        write(&earlier.join("project.yaml"), b"name: earlier\n");
        write(&earlier.join("old/data.bin"), &[0, 1, 2, 255, 254, 10, 13]);
        write(
            &tmp.path().join(MOAI_DIR).join("project.yaml"),
            b"name: new\n",
        );
    }
    tmp
}

fn flaky_name() -> TestName {
    TestName {
        binary: "moai-cli::demo_cli".to_string(),
        name: "t_flaky".to_string(),
    }
}

/// Every entry under `root` (folders as `None`), by relative path, with the bytes of each file.
fn snapshot(root: &Path) -> BTreeMap<String, Option<Vec<u8>>> {
    let mut out = BTreeMap::new();
    let mut stack = vec![root.to_path_buf()];
    while let Some(folder) = stack.pop() {
        let Ok(entries) = fs::read_dir(&folder) else {
            continue;
        };
        for e in entries.flatten() {
            let path = e.path();
            let rel = path
                .strip_prefix(root)
                .unwrap()
                .to_string_lossy()
                .into_owned();
            if e.file_type().unwrap().is_dir() {
                out.insert(rel, None);
                stack.push(path);
            } else {
                out.insert(rel, Some(fs::read(&path).unwrap()));
            }
        }
    }
    out
}

/// Where the new ledger belongs, built from the constants (never from `ledger_path`).
fn new_ledger(root: &Path) -> PathBuf {
    root.join(MOAI_DIR).join(LOCAL_DIR).join(LEDGER)
}

/// AC-4 (REQ-5, REQ-6): the row goes to the `local` folder of the new state folder; the earlier
/// folder keeps every byte and gains no entry; `ledger_path` names the new place. The problems of
/// both layouts are collected, so one run shows all of them.
#[test]
fn the_flaky_ledger_is_written_under_the_local_folder_and_the_earlier_folder_stays_untouched() {
    let mut problems: Vec<String> = Vec::new();
    for layout in [Layout::EarlierPresent, Layout::Empty] {
        let tmp = project(layout);
        let root = tmp.path();
        if let Layout::EarlierPresent = layout {
            // The premise: this is the shape in which the earlier folder is picked today.
            assert_eq!(dir(root), EARLIER_PROJECT_DIR, "{layout:?}");
        }
        let earlier = root.join(EARLIER_PROJECT_DIR);
        let before = snapshot(&earlier);
        let out = tempfile::tempdir().unwrap();
        let recorded = record(root, out.path(), HEAD, &[flaky_name()]);
        if let Err(e) = &recorded {
            problems.push(format!("{layout:?}: record failed: {e}"));
        }
        if snapshot(&earlier) != before {
            problems.push(format!(
                "{layout:?}: the earlier folder changed: before {:?}, after {:?}",
                before.keys().collect::<Vec<_>>(),
                snapshot(&earlier).keys().collect::<Vec<_>>()
            ));
        }
        let expected = new_ledger(root);
        if ledger_path(root) != expected {
            problems.push(format!(
                "{layout:?}: ledger_path is {}, want {}",
                ledger_path(root).display(),
                expected.display()
            ));
        }
        let text = fs::read_to_string(&expected).unwrap_or_default();
        let rows: Vec<serde_json::Value> = text
            .lines()
            .filter_map(|l| serde_json::from_str(l).ok())
            .collect();
        if rows.len() != 1 || rows[0]["name"] != flaky_name().text().as_str() {
            problems.push(format!(
                "{layout:?}: want one row for {} in {}, found {text:?}",
                flaky_name().text(),
                expected.display()
            ));
        }
        // A second record appends to the same ledger.
        let _ = record(root, out.path(), HEAD, &[flaky_name()]);
        let again = fs::read_to_string(&expected).unwrap_or_default();
        if again.lines().count() != 2 {
            problems.push(format!(
                "{layout:?}: want two rows after a second record, found {again:?}"
            ));
        }
        if let Layout::Empty = layout {
            if root.join(EARLIER_PROJECT_DIR).exists() {
                problems.push(format!("{layout:?}: the earlier folder was created"));
            }
        }
    }
    assert!(problems.is_empty(), "{}", problems.join("\n"));
}

/// AC-4 (REQ-5, REQ-6): recording a flaky name does not change which folder `state_root::dir`
/// names. The generation `classify` reports is printed before and after, not compared: the new
/// folder can turn a no-marker `none` into `unknown` (SPEC Risks), and that is a recorded fact.
#[test]
fn the_state_folder_answer_is_the_same_after_a_flaky_record() {
    for layout in [Layout::EarlierPresent, Layout::Empty] {
        let tmp = project(layout);
        let root = tmp.path();
        let answer_before = dir(root);
        let generation_before = classify(root).generation;
        let out = tempfile::tempdir().unwrap();
        record(root, out.path(), HEAD, &[flaky_name()]).unwrap();
        let answer_after = dir(root);
        let generation_after = classify(root).generation;
        eprintln!(
            "{layout:?}: dir {answer_before} -> {answer_after}; classify.generation {generation_before} -> {generation_after}"
        );
        assert_eq!(answer_before, answer_after, "{layout:?}");
    }
}

/// One ledger row in the format `record` writes.
fn row(name: &str) -> String {
    format!(
        "{}\n",
        serde_json::json!({
            "name": name,
            "time": "2026-11-20T00:00:00Z",
            "head": "0123456789ab",
            "folder": "elsewhere",
        })
    )
}

/// AC-5 (REQ-7, REQ-8): a name seen twice in the earlier ledger and once in the new one counts
/// three times and is marked `fix`; the earlier ledger keeps every byte through the record and
/// the report.
#[test]
fn the_flaky_report_counts_both_ledgers_and_reads_the_earlier_one_only() {
    let tmp = project(Layout::EarlierPresent);
    let root = tmp.path();
    // The report finds the project root through git; make this folder its own root.
    git_ok(root, &["init", "-q"]);
    let name = flaky_name().text();
    let earlier_ledger = root.join(EARLIER_PROJECT_DIR).join(LEDGER);
    write(
        &earlier_ledger,
        format!("{}{}", row(&name), row(&name)).as_bytes(),
    );
    let earlier_bytes = fs::read(&earlier_ledger).unwrap();
    let earlier_folder = snapshot(&root.join(EARLIER_PROJECT_DIR));

    let out = tempfile::tempdir().unwrap();
    record(root, out.path(), HEAD, &[flaky_name()]).unwrap();
    assert_eq!(
        snapshot(&root.join(EARLIER_PROJECT_DIR)),
        earlier_folder,
        "the record changed the earlier folder"
    );
    let v = report(root).unwrap();
    let found = v["flaky"]
        .as_array()
        .unwrap()
        .iter()
        .find(|r| r["name"] == name.as_str())
        .unwrap_or_else(|| panic!("{name} is not in the report: {v}"));
    assert_eq!(found["count"], 3, "{v}");
    assert_eq!(found["status"], "fix", "{v}");
    assert_eq!(
        fs::read(&earlier_ledger).unwrap(),
        earlier_bytes,
        "the report changed the earlier ledger"
    );
    assert_eq!(snapshot(&root.join(EARLIER_PROJECT_DIR)), earlier_folder);
}
