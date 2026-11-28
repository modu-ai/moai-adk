//! The capability table of `moai doctor` (SPEC-20261127-doctor-moai-layer REQ-12): which feature
//! each host offers, from the documentation the SPEC's research read on 2026-11-27 (the evidence
//! table "능력 표" of the SPEC). A cell is `yes`, `no`, `partial` or `unverified`; a cell whose
//! basis the research only inferred, or could not read, is `unverified` and never `yes`. `docs`
//! names the page, `line` the quoted line or the section it came from (or says that the SPEC
//! kept only the row), and `read_at` the day of the reading. The eleven cells that were
//! `unverified` in slice 2 had their pages read again on 2026-11-28 (slice 3): none of the seven
//! pages states them, so they stay `unverified` and `line` says what the pages say instead.

use super::{Check, State};
use serde_json::{json, Map, Value};

/// The day the pages of the SPEC's evidence table were read.
const READ_AT: &str = "2026-11-27";
/// The day the pages behind the cells that stayed `unverified` were read again (slice 3).
const RE_READ_AT: &str = "2026-11-28";

/// The hosts, in the column order of the table.
pub(super) const COLUMNS: [&str; 5] = ["cli", "gateway", "desktop", "vscode", "cloud"];

/// One cell: its value, the page and the line (or section) it came from, and the day it was read.
struct Cell {
    value: &'static str,
    docs: &'static str,
    line: &'static str,
    read_at: &'static str,
}

const fn cell(value: &'static str, docs: &'static str, line: &'static str) -> Cell {
    Cell {
        value,
        docs,
        line,
        read_at: READ_AT,
    }
}

/// A cell whose pages were read again and still do not state it: `unverified`, and `line` says
/// what the pages say instead (SPEC-20261127-doctor-moai-layer slice 3, REQ-12).
const fn not_stated(docs: &'static str, line: &'static str) -> Cell {
    Cell {
        value: "unverified",
        docs,
        line,
        read_at: RE_READ_AT,
    }
}

/// For a cell whose page section the research did not keep.
const ROW_ONLY: &str = "the SPEC evidence table kept this row; the page section was not kept";

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
                not_stated(
                    "cross-session-messaging.md",
                    "Availability says \"available on every provider, including Amazon Bedrock, Claude Platform on AWS, Google Cloud's Agent Platform, and Microsoft Foundry\" and names no gateway; llm-gateway-connect.md never mentions cross-session messaging",
                ),
                cell(
                    "partial",
                    "cross-session-messaging.md",
                    "\"local, SSH, and WSL sessions\"",
                ),
                cell("yes", "cross-session-messaging.md", "Availability"),
                not_stated(
                    "cross-session-messaging.md",
                    "no line covers a cloud session messaging a session on its own machine; the page only says a cloud session can't use SendMessage to answer a message from another machine",
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
                not_stated(
                    "cross-session-messaging.md",
                    "no line gives the desktop app availability for sessions beyond this machine; the only desktop line is about the approval dialog (\"A session in the VS Code extension or the Desktop app can't show the dialog\"), and desktop.md names cross-session messaging without saying which machines",
                ),
                not_stated(
                    "cross-session-messaging.md",
                    "no line gives the VS Code extension availability for sessions beyond this machine; the only such line is about the approval dialog (\"A session in the VS Code extension or the Desktop app can't show the dialog\"), and vs-code.md has no line on messaging another machine",
                ),
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
                not_stated(
                    "output-styles.md",
                    "the page names no gateway or ANTHROPIC_BASE_URL, and llm-gateway-connect.md never mentions output styles",
                ),
                cell("partial", "output-styles.md", ROW_ONLY),
                cell("yes", "output-styles.md", "since 2.1.257"),
                not_stated(
                    "output-styles.md",
                    "the page has no line about Claude Code on the web; the nearest line is about Remote Control from the mobile app or web, where only the built-in styles can be listed and selected (/output-style, v2.1.269 or later)",
                ),
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
                not_stated(
                    "statusline.md",
                    "the page never mentions the desktop app, and desktop.md names no status line or 5-hour or 7-day limit (its \"Check usage\" section describes a usage ring)",
                ),
                not_stated(
                    "statusline.md",
                    "the page never mentions VS Code; vs-code.md shows plan usage bars in the /usage dialog, not in a status line",
                ),
                not_stated(
                    "statusline.md",
                    "the page never mentions Claude Code on the web or cloud sessions; it says rate_limits appears only for claude.ai Pro, Max and Ultra subscribers or behind a Claude apps gateway that sets a spend limit",
                ),
            ],
        },
        Row {
            feature: "artifacts",
            cells: [
                cell("yes", "artifacts.md", "Availability"),
                cell("no", "artifacts.md", "Availability"),
                cell("yes", "artifacts.md", "Availability"),
                not_stated(
                    "artifacts.md",
                    "the Availability \"Surface\" row names the Claude Code CLI, the Claude desktop app, Claude Tag and claude remote-control sessions, and the words VS Code appear nowhere in the page; a surface that is not listed is not stated as unsupported; vs-code.md has no line on artifacts",
                ),
                not_stated(
                    "artifacts.md",
                    "the page names a cloud session only for deleting (\"outside a cloud session\") and for reading a shared artifact, and no line says a cloud session can publish; the Availability \"Surface\" row does not list web or cloud",
                ),
            ],
        },
    ]
}

fn cell_json(c: &Cell) -> Value {
    json!({ "value": c.value, "docs": c.docs, "line": c.line, "read_at": c.read_at })
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
