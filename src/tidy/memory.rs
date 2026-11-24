//! The size line of the automatic memory index (SPEC-20261124-home-tidy-slice2, REQ-9 and
//! REQ-10): one file named by the caller is measured against the host's load cap and a notice
//! is raised above 80 percent. The file is only read, never written, and nothing else under the
//! Claude folders is walked.

use serde_json::{json, Value};
use std::fs;
use std::io::Read;
use std::path::Path;

/// The host loads the first 200 lines or 25 KB of the index, whichever comes first.
pub const CAP_LINES: u64 = 200;
pub const CAP_BYTES: u64 = 25 * 1024;
/// The share of the cap above which the notice is raised.
pub const NOTICE_PERCENT: f64 = 80.0;
/// The targets of the leader's note (A50): under 100 lines and 12 KB, 120 bytes a line.
pub const TARGET_LINES: u64 = 100;
pub const TARGET_BYTES: u64 = 12 * 1024;
pub const TARGET_LINE_BYTES: u64 = 120;
/// An index larger than this is not read (the host cap is far below it).
pub const READ_LIMIT: u64 = 1024 * 1024;

/// One row for `--memory-index <path>`; every failure is a row with a reason, never an error.
pub fn row(path: &Path) -> Value {
    let shown = path.display().to_string();
    let md = match fs::symlink_metadata(path) {
        Ok(m) => m,
        Err(e) => {
            return json!({"path": shown, "found": false, "reason": format!("missing: {}", e.kind())})
        }
    };
    if md.file_type().is_symlink() {
        // A link is followed by no part of tidy; the caller names the real file.
        return json!({"path": shown, "found": true, "reason": "symlink_not_followed"});
    }
    if !md.is_file() {
        return json!({"path": shown, "found": true, "reason": "not_a_regular_file"});
    }
    if md.len() > READ_LIMIT {
        return json!({"path": shown, "found": true, "bytes": md.len(), "reason": "over_limit: not read"});
    }
    let mut bytes = Vec::new();
    let read = fs::File::open(path).and_then(|f| f.take(READ_LIMIT + 1).read_to_end(&mut bytes));
    if let Err(e) = read {
        return json!({"path": shown, "found": true, "reason": format!("unreadable: {}", e.kind())});
    }
    let lines = bytes.iter().filter(|b| **b == b'\n').count() as u64
        + u64::from(bytes.last().is_some_and(|b| *b != b'\n'));
    let longest = bytes
        .split(|b| *b == b'\n')
        .map(|l| l.len() as u64 + 1)
        .max()
        .unwrap_or(0);
    let size = bytes.len() as u64;
    let percent =
        (lines as f64 * 100.0 / CAP_LINES as f64).max(size as f64 * 100.0 / CAP_BYTES as f64);
    json!({
        "path": shown,
        "found": true,
        "lines": lines,
        "bytes": size,
        "longest_line_bytes": longest,
        "percent": percent,
        "over_80": percent > NOTICE_PERCENT,
        "over_target": lines > TARGET_LINES || size > TARGET_BYTES || longest > TARGET_LINE_BYTES,
        "cap": {"lines": CAP_LINES, "bytes": CAP_BYTES},
        "targets": {"lines": TARGET_LINES, "bytes": TARGET_BYTES, "line_bytes": TARGET_LINE_BYTES},
    })
}

/// The rows of every `--memory-index` path.
pub fn rows(paths: &[std::path::PathBuf]) -> Vec<Value> {
    paths.iter().map(|p| row(p)).collect()
}

/// The text for the user to run in a Claude session, when any index is over 80 percent. moai
/// never edits the index and never runs `claude purge`; that line is a reference only.
pub fn tidy_text(rows: &[Value]) -> Option<Value> {
    if !rows.iter().any(|r| r["over_80"] == true) {
        return None;
    }
    Some(json!({
        "groups": ["rules_to_instructions", "retired_items", "merges"],
        "run_in_claude_session": "In a Claude session, ask for a plan that sorts the memory index \
            into three groups: (1) rules that belong in the instruction files, (2) items that are \
            retired and can go, (3) entries that can be merged. Read the plan, then pick what \
            runs. moai never edits MEMORY.md.",
        "purge_reference": {
            "command": "claude purge --all",
            "never_run": true,
            "note": "reference only: it deletes the Claude Code data of every project; read from `claude purge --help` on 2026-11-24",
        },
    }))
}
