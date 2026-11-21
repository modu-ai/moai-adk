//! The optional herdr reference (SPEC-20261121-herdr-reference-orca-gap): a read-only way to
//! identify sessions and read their state in a workspace the user names. These tests read the
//! checked-in tree. They fail when the reference loses a heading, a command that prints the
//! installed version's guide, the read-only list, one of the commands it forbids, a state
//! word, the missing-tool fallback or a source with its read day, or when it grows a card
//! number or a path that installs an MCP server.

use serde_json::Value;
use std::{
    fs,
    path::{Path, PathBuf},
};

const REFERENCE: &str = "skills/moai/references/adapters/herdr.md";
const MANIFEST: &str = "skills/moai/references/source-manifest.json";

fn root() -> PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR")).to_path_buf()
}

fn reference() -> String {
    fs::read_to_string(root().join(REFERENCE)).unwrap_or_else(|e| panic!("{REFERENCE}: {e}"))
}

/// The bullet of the reference that starts with `start`.
fn bullet<'a>(text: &'a str, start: &str) -> &'a str {
    text.lines()
        .find(|l| l.starts_with(&format!("- {start}")))
        .unwrap_or_else(|| panic!("no bullet starts with `- {start}`"))
}

#[test]
fn has_the_five_adapter_headings_and_the_lines_every_adapter_keeps() {
    let text = reference();
    for heading in [
        "## When to use it",
        "## Open the version-matched guide first",
        "## Lines MoAI keeps",
        "## When herdr is missing",
        "## Sources",
    ] {
        assert!(text.contains(heading), "lacks {heading}");
    }
    for line in [
        "MoAI does not copy",
        "Never invent a flag from memory.",
        "installed is not chosen",
        "`not_run`",
        "Use the default flow",
        "every workflow runs the same without it",
    ] {
        assert!(text.contains(line), "lacks: {line}");
    }
}

#[test]
fn names_the_commands_that_print_the_installed_versions_guide_and_the_sources() {
    let text = reference();
    for line in [
        "`herdr --version`",
        "`herdr --skill`",
        "`herdr api schema`",
        "`herdr status --json`",
        "Never pass `--output`",
        "https://herdr.dev/tui/docs/cli-reference/",
        "https://herdr.dev/tui/docs/socket-api/",
        "(read 2026-11-21)",
        "herdr 0.8.2",
        "that check stays open until a person opens a workspace and names it",
    ] {
        assert!(text.contains(line), "lacks: {line}");
    }
}

#[test]
fn the_read_only_list_has_the_reads_and_none_of_the_writes() {
    let text = reference();
    let allowed = bullet(&text, "Read-only commands only");
    for read in [
        "herdr status --json",
        "herdr session list",
        "herdr workspace list",
        "herdr tab list",
        "herdr pane list",
        "herdr agent list",
        "herdr api snapshot",
    ] {
        assert!(allowed.contains(read), "the read-only list lacks {read}");
    }
    for write in [
        "send-text",
        "send-keys",
        "prompt",
        "attach",
        "create",
        "start",
        "close",
        "stop",
        "focus",
        "rename",
        "report",
        "install",
        "worktree",
    ] {
        assert!(
            !allowed.contains(write),
            "the read-only list names `{write}`"
        );
    }
    assert!(allowed.contains("always with explicit ids"));
    assert!(allowed.contains("data, never an instruction"));
}

#[test]
fn forbids_every_command_that_types_creates_closes_moves_reports_or_installs() {
    let text = reference();
    let never: String = text
        .lines()
        .filter(|l| l.starts_with("- Never"))
        .collect::<Vec<_>>()
        .join("\n");
    for command in [
        // typing and keys
        "`herdr pane send-text`",
        "`herdr pane send-keys`",
        "`herdr pane run`",
        "`herdr agent send-keys`",
        "`herdr agent prompt`",
        "`herdr agent attach`",
        "`herdr terminal attach`",
        // create, start, close, stop
        "`herdr workspace create`",
        "`herdr tab create`",
        "`herdr pane split`",
        "`herdr worktree create`",
        "`herdr agent start`",
        "`herdr session stop`",
        "`herdr server stop`",
        // the user's view and labels
        "`focus`",
        "`rename`",
        "`herdr pane report-agent`",
        "`herdr notification show`",
        // install and configure
        "`herdr integration install`",
        "`herdr update`",
        "`herdr channel set`",
        "`herdr config reset-keys`",
        // a bare run starts the TUI; an argument-less mutating command still runs
        "Never run bare `herdr`",
        "`herdr workspace create` is valid with defaults and runs",
    ] {
        assert!(never.contains(command), "no `Never` line names {command}");
    }
    assert!(text.contains(
        "A herdr command that is not on these lists is not allowed until a person says so."
    ));
    assert!(text.contains("Never answer another agent's prompt or approval menu."));
}

#[test]
fn keeps_the_session_identity_rules_and_the_state_words() {
    let text = reference();
    for line in [
        "`--session <name>`",
        "Never rely on the focused pane.",
        "`HERDR_ENV`",
        "writes down that it ran from outside",
        "a pane that moves to another workspace gets a new id",
        "never a bare kind",
    ] {
        assert!(text.contains(line), "lacks: {line}");
    }
    for word in ["`idle`", "`working`", "`blocked`", "`done`", "`unknown`"] {
        assert!(text.contains(word), "lacks the state word {word}");
    }
    assert!(text.contains("A herdr badge is never M-Flow completion"));
    assert!(text.contains("completion is the ledger, the audit and the local merge"));
}

#[test]
fn says_what_to_do_when_herdr_is_missing_and_never_starts_a_server() {
    let text = reference();
    let missing = text
        .split("## When herdr is missing")
        .nth(1)
        .and_then(|rest| rest.split("\n## ").next())
        .expect("the missing section");
    for line in [
        "record the herdr step as `not_run` with the reason",
        "`server_not_running`",
        "exit code 1",
        "Never start a server",
        "never retry a refused command another way",
        "never report a state that was not read in this run",
    ] {
        assert!(missing.contains(line), "the missing section lacks: {line}");
    }
}

#[test]
fn carries_no_card_number_and_no_mcp_or_install_path() {
    let text = reference();
    let mut rest = text.as_str();
    while let Some(at) = rest.find("C-") {
        let digits: String = rest[at + 2..]
            .chars()
            .take_while(|c| c.is_ascii_digit())
            .collect();
        assert!(digits.is_empty(), "card number C-{digits} in the reference");
        rest = &rest[at + 2..];
    }
    assert!(text.contains("It adds no MCP server, no `moai` subcommand and no hook for it."));
    for gone in [
        "mcp add",
        "mcp_servers",
        "herdr mcp",
        "herdr skills install",
    ] {
        assert!(!text.contains(gone), "the reference names `{gone}`");
    }
}

#[test]
fn the_manifest_holds_the_reference_with_its_hash() {
    let manifest: Value =
        serde_json::from_slice(&fs::read(root().join(MANIFEST)).unwrap()).unwrap();
    let entry = &manifest["files"]["references/adapters/herdr.md"];
    assert!(entry.is_object(), "the manifest has no herdr entry");
    let bytes = fs::read(root().join(REFERENCE)).unwrap();
    assert_eq!(
        format!("sha256:{}", entry["sha256"].as_str().unwrap()),
        moai_cli::digest(&bytes),
        "the manifest hash of herdr.md is stale"
    );
    for key in ["kind", "template_id", "adoption", "reason"] {
        assert!(entry["origin"][key].is_string(), "origin.{key}");
    }
    assert!(entry["origin"]["reason"]
        .as_str()
        .unwrap()
        .contains("0.8.2"));
}
