//! The check-affected receipt a READY or REPAIRED note needs (SPEC-20261124-ready-receipt-gate):
//! one `summary.txt` line in `<worktree>/target/check-affected-<head12>/`. A test whose subject is
//! not the receipt gate writes the receipt of the head it names with this instead of running
//! `moai check affected`. Included by a test file with
//! `#[path = "common/ready_receipt.rs"] mod ready_receipt;`.

use std::path::{Path, PathBuf};

/// The folder `moai check affected` writes its logs and `summary.txt` into for `head`.
#[allow(dead_code)]
pub fn receipt_dir(worktree: &Path, head: &str) -> PathBuf {
    worktree
        .join("target")
        .join(format!("check-affected-{}", &head[..12]))
}

/// A summary line in the shape `moai check affected` answers, for `head` and `result`
/// (`PASS` or `FAIL`).
#[allow(dead_code)]
pub fn summary_line(head: &str, result: &str) -> String {
    let code = if result == "PASS" { 0 } else { 1 };
    format!(
        "check-affected base=000000000000 head={} targets=1 scans=listed:0 test-threads=4 \
         diff_check=0 fmt=0 clippy=0 test={code} result={result}",
        &head[..12]
    )
}

/// Adds `text` as a `report` note of `card` (`C-<n>`) signed `lane`, through the ledger the way
/// `card note` stores it, without the receipt gate; returns the note's number. For a fixture
/// whose project folder is not the repository that holds the heads (the bundle tests), so there
/// is no worktree to put a receipt in, and whose subject is not the gate.
#[allow(dead_code)]
pub fn add_note_unchecked(home: &Path, project_id: &str, card: &str, text: &str) -> i64 {
    let id: i64 = card.strip_prefix("C-").unwrap().parse().unwrap();
    let mut w = moai_cli::ledger::open_write(home, project_id).unwrap();
    moai_cli::ledger::workflow::add_note(&mut w, id, "report", "lane", None, text.trim_end(), None)
        .unwrap()
        .note_id
}

/// Writes the receipt of `head` under `worktree` and returns its path.
#[allow(dead_code)]
pub fn write_receipt(worktree: &Path, head: &str, result: &str) -> PathBuf {
    let dir = receipt_dir(worktree, head);
    std::fs::create_dir_all(&dir).unwrap();
    let path = dir.join("summary.txt");
    std::fs::write(&path, format!("{}\n", summary_line(head, result))).unwrap();
    path
}
