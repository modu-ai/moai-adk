//! The capability table of `moai doctor` (SPEC-20261127-doctor-moai-layer REQ-12): which feature
//! each host offers, from the documentation the SPEC's research read on 2026-11-27 (the evidence
//! table "능력 표" of the SPEC). A cell is `yes`, `no`, `partial` or `unverified`; a cell whose
//! basis the research only inferred, or could not read, is `unverified` and never `yes`. `docs`
//! names the page, `line` the quoted line or the section it came from (or says that the SPEC
//! kept only the row), and `read_at` the day of the reading.

use super::{Check, State};
use serde_json::{json, Map, Value};

/// The day the pages were read (the SPEC's evidence table).
const READ_AT: &str = "2026-11-27";

/// The hosts, in the column order of the table.
pub(super) const COLUMNS: [&str; 5] = ["cli", "gateway", "desktop", "vscode", "cloud"];

/// One cell: its value, the page and the line (or section) it came from.
struct Cell {
    value: &'static str,
    docs: &'static str,
    line: &'static str,
}

const fn cell(value: &'static str, docs: &'static str, line: &'static str) -> Cell {
    Cell { value, docs, line }
}

/// For a cell whose page section the research did not keep.
const ROW_ONLY: &str = "the SPEC evidence table kept this row; the page section was not kept";
/// For a cell the docs do not state.
const NOT_STATED: &str = "not stated in the pages that were read";

struct Row {
    feature: &'static str,
    cells: [Cell; 5],
}

/// The eight rows of the SPEC's evidence table, one cell per host.
fn rows() -> [Row; 8] {
    [
        Row {
            feature: "remote_control",
            cells: [
                cell("yes", "remote-control.md", ROW_ONLY),
                cell(
                    "no",
                    "feature-availability.md",
                    "Features that require a Claude subscription",
                ),
                cell("yes", "remote-control.md", ROW_ONLY),
                cell("yes", "remote-control.md", ROW_ONLY),
                cell("no", "remote-control.md", ROW_ONLY),
            ],
        },
        Row {
            feature: "cross_session_same_machine",
            cells: [
                cell("yes", "cross-session-messaging.md", "Availability"),
                cell(
                    "unverified",
                    "cross-session-messaging.md",
                    "\"available on every provider\" was read, but not for the gateway column itself",
                ),
                cell(
                    "partial",
                    "cross-session-messaging.md",
                    "\"local, SSH, and WSL sessions\"",
                ),
                cell("yes", "cross-session-messaging.md", "Availability"),
                cell(
                    "unverified",
                    "cross-session-messaging.md",
                    "only inferred, not stated",
                ),
            ],
        },
        Row {
            feature: "cross_session_other_machine",
            cells: [
                cell(
                    "yes",
                    "cross-session-messaging.md",
                    "needs Remote Control (Availability)",
                ),
                cell("no", "cross-session-messaging.md", ROW_ONLY),
                cell("unverified", "cross-session-messaging.md", NOT_STATED),
                cell("unverified", "cross-session-messaging.md", NOT_STATED),
                cell(
                    "partial",
                    "cross-session-messaging.md",
                    "receiving only (Availability)",
                ),
            ],
        },
        Row {
            feature: "mods",
            cells: [
                cell("yes", "plugins/mods/overview.md", ROW_ONLY),
                cell(
                    "yes",
                    "plugins/mods/overview.md",
                    "\"whichever provider serves the model\"",
                ),
                cell("yes", "plugins/mods/overview.md", "since 2.1.286"),
                cell("yes", "plugins/mods/overview.md", ROW_ONLY),
                cell("partial", "plugins/mods/overview.md", ROW_ONLY),
            ],
        },
        Row {
            feature: "plugins",
            cells: [
                cell("yes", "plugins/mods/overview.md", ROW_ONLY),
                cell("yes", "plugins/mods/overview.md", ROW_ONLY),
                cell("yes", "plugins/mods/overview.md", ROW_ONLY),
                cell("yes", "plugins/mods/overview.md", ROW_ONLY),
                cell("partial", "plugins/mods/overview.md", ROW_ONLY),
            ],
        },
        Row {
            feature: "output_styles",
            cells: [
                cell("yes", "output-styles.md", ROW_ONLY),
                cell("unverified", "output-styles.md", NOT_STATED),
                cell("partial", "output-styles.md", ROW_ONLY),
                cell("yes", "output-styles.md", "since 2.1.257"),
                cell("unverified", "output-styles.md", NOT_STATED),
            ],
        },
        Row {
            feature: "rate_limit_status_line",
            cells: [
                cell("yes", "statusline.md", "Rate limit usage (Pro, Max, Ultra)"),
                cell(
                    "no",
                    "feature-availability.md",
                    "read as an exclusion in Features that require a Claude subscription",
                ),
                cell("unverified", "statusline.md", NOT_STATED),
                cell("unverified", "statusline.md", NOT_STATED),
                cell("unverified", "statusline.md", NOT_STATED),
            ],
        },
        Row {
            feature: "artifacts",
            cells: [
                cell("yes", "artifacts.md", "Availability"),
                cell("no", "artifacts.md", "Availability"),
                cell("yes", "artifacts.md", "Availability"),
                cell("unverified", "artifacts.md", NOT_STATED),
                cell("unverified", "artifacts.md", NOT_STATED),
            ],
        },
    ]
}

fn cell_json(c: &Cell) -> Value {
    json!({ "value": c.value, "docs": c.docs, "line": c.line, "read_at": READ_AT })
}

/// The host column the process runs under: `gateway` when `ANTHROPIC_BASE_URL` names a host other
/// than `api.anthropic.com`, otherwise `cli` (REQ-12).
pub(super) fn host_of(base_url: Option<&str>) -> &'static str {
    let Some(url) = base_url else {
        return "cli";
    };
    let Some((_, rest)) = url.split_once("://") else {
        return "cli";
    };
    let authority = rest.split(['/', '?', '#']).next().unwrap_or("");
    // Drop any user information, then the port.
    let host = authority.rsplit('@').next().unwrap_or("");
    let host = if host.starts_with('[') {
        host.split(']').next().unwrap_or("")
    } else {
        host.split(':').next().unwrap_or("")
    };
    if host.is_empty() || host.eq_ignore_ascii_case("api.anthropic.com") {
        "cli"
    } else {
        "gateway"
    }
}

/// The cells of the detected host: one per row (REQ-12).
pub(super) fn host_check(host: &'static str) -> Check {
    let column = COLUMNS.iter().position(|c| *c == host).unwrap_or(0);
    let mut cells = Map::new();
    let mut unverified = 0;
    for row in rows() {
        let c = &row.cells[column];
        if c.value == "unverified" {
            unverified += 1;
        }
        cells.insert(row.feature.to_string(), cell_json(c));
    }
    Check {
        id: "host_capabilities",
        state: State::Ok,
        summary: format!(
            "the host reads as {host}; {unverified} of 8 features are unverified for it in the docs that were read"
        ),
        evidence: json!({ "host": host, "cells": cells }),
        next: "`moai doctor --table` prints the whole table".into(),
    }
}

/// The whole table, printed with `--table`.
pub(super) fn table() -> Value {
    json!({
        "columns": COLUMNS,
        "read_at": READ_AT,
        "rows": rows()
            .iter()
            .map(|row| {
                let mut cells = Map::new();
                for (column, c) in COLUMNS.iter().zip(row.cells.iter()) {
                    cells.insert((*column).to_string(), cell_json(c));
                }
                json!({ "feature": row.feature, "cells": cells })
            })
            .collect::<Vec<_>>(),
    })
}
