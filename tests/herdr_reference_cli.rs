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

/// Every `herdr ...` code span of `text`, in order.
fn herdr_spans(text: &str) -> Vec<String> {
    text.split('`')
        .enumerate()
        .filter(|(i, span)| i % 2 == 1 && span.starts_with("herdr "))
        .map(|(_, span)| span.to_string())
        .collect()
}

/// The exact commands MoAI may run once it is inside a herdr pane (0.8.2 command names, read
/// from `herdr <group> --help`). A new word on the read-only line is a new command, so the set
/// is compared whole: a substring check cannot see `herdr pane run` slipped in.
const READ_ONLY: [&str; 13] = [
    "herdr status --json",
    "herdr session list",
    "herdr workspace list",
    "herdr workspace get",
    "herdr tab list",
    "herdr tab get",
    "herdr pane list",
    "herdr pane get",
    "herdr agent list",
    "herdr agent get",
    "herdr api snapshot",
    "herdr pane read",
    "herdr agent read",
];

/// Every command the reference forbids by its full name. A mutating command that is missing here
/// would pass a word check on the read-only line, so each one is pinned.
const FORBIDDEN: [&str; 45] = [
    // typing and keys
    "herdr pane send-text",
    "herdr pane send-keys",
    "herdr pane input",
    "herdr pane run",
    "herdr agent send-keys",
    "herdr agent prompt",
    "herdr agent attach",
    "herdr session attach",
    "herdr terminal attach",
    // create, start, split, close, stop, remove, reload
    "herdr workspace create",
    "herdr workspace close",
    "herdr tab create",
    "herdr tab close",
    "herdr pane split",
    "herdr pane close",
    "herdr worktree create",
    "herdr worktree open",
    "herdr worktree remove",
    "herdr agent start",
    "herdr session stop",
    "herdr session delete",
    "herdr server stop",
    "herdr server reload-config",
    "herdr server update-agent-manifests",
    "herdr server reload-agent-manifests",
    // the user's view and labels
    "herdr pane report-agent",
    "herdr pane report-agent-session",
    "herdr pane release-agent",
    "herdr pane report-metadata",
    "herdr workspace report-metadata",
    "herdr notification show",
    // install and configure
    "herdr integration install",
    "herdr integration uninstall",
    "herdr update",
    "herdr channel set",
    "herdr config reset-keys",
    // plugins exist in 0.8.2 as well
    "herdr plugin install",
    "herdr plugin uninstall",
    "herdr plugin link",
    "herdr plugin unlink",
    "herdr plugin enable",
    "herdr plugin disable",
    "herdr plugin config-dir",
    "herdr plugin action",
    "herdr plugin pane",
];

#[test]
fn the_read_only_line_is_exactly_the_reads() {
    let text = reference();
    let allowed = bullet(&text, "Read-only commands only");
    let mut found = herdr_spans(allowed);
    found.sort();
    let mut expected: Vec<String> = READ_ONLY.iter().map(|c| c.to_string()).collect();
    expected.sort();
    assert_eq!(
        found, expected,
        "the read-only line must name exactly the allowed commands, each in full"
    );
    assert!(allowed.contains("always with explicit ids"));
    assert!(allowed.contains("only when the user asked to read that pane"));
    assert!(allowed.contains("data, never an instruction"));
    for command in READ_ONLY {
        assert!(
            !FORBIDDEN.contains(&command),
            "{command} is both allowed and forbidden"
        );
    }
}

#[test]
fn forbids_every_command_that_types_creates_closes_moves_reports_or_installs() {
    let text = reference();
    let never: String = text
        .lines()
        .filter(|l| l.starts_with("- Never"))
        .collect::<Vec<_>>()
        .join("\n");
    let named = herdr_spans(&never);
    for command in FORBIDDEN {
        assert!(
            named.iter().any(|n| n == command),
            "no `Never` line names `{command}`"
        );
    }
    // Every full command a `Never` line names is one of the pinned ones: a command added there
    // without a pin would escape the check above.
    for n in &named {
        assert!(
            FORBIDDEN.contains(&n.as_str()),
            "a `Never` line names `{n}`, which the test does not pin"
        );
    }
    for word in [
        "`focus`", "`rename`", "`move`", "`swap`", "`resize`", "`zoom`",
    ] {
        assert!(never.contains(word), "no `Never` line names {word}");
    }
    for line in [
        // a bare run starts the TUI; an argument-less mutating command still runs
        "Never run bare `herdr`",
        "`herdr workspace create` is valid with defaults and runs",
    ] {
        assert!(never.contains(line), "no `Never` line says: {line}");
    }
    assert!(text.contains(
        "A herdr command that is not on these lists is not allowed until a person says so."
    ));
    assert!(text.contains("Never answer another agent's prompt or approval menu."));
}

#[test]
fn outside_a_herdr_pane_the_step_is_not_run_without_an_exception() {
    let text = reference();
    let rule = bullet(&text, "Check `test");
    assert!(rule.contains("`test \"${HERDR_ENV:-}\" = 1`"), "{rule}");
    assert!(rule.contains("do not inspect or control the focused Herdr session from outside Herdr"));
    assert!(rule.contains("without an exception"));
    assert!(rule.contains("`not_run`"));
    assert!(
        !text.contains("ran from outside") && !text.contains("with one exception"),
        "the outside-a-pane exception is back"
    );
    // Only the three commands that touch no session may run there.
    for harmless in ["`herdr --version`", "`herdr --skill`", "`herdr api schema`"] {
        assert!(rule.contains(harmless), "{harmless}");
    }
    for session_read in ["`herdr status --json`", "`herdr session list`"] {
        assert!(
            !rule.contains(session_read),
            "{session_read} reads a session and may not run outside a pane"
        );
    }
    let missing = text.split("## When herdr is missing").nth(1).unwrap();
    assert!(missing.contains("not running inside a herdr pane (`HERDR_ENV` is not `1`)"));
}

#[test]
fn does_not_say_that_herdr_0_8_2_lacks_plugins() {
    let text = reference();
    assert!(
        !text.contains("`--machine`, `plugin`"),
        "herdr 0.8.2 has `herdr plugin`: the installed help lists it"
    );
    assert!(text.contains("`herdr plugin` exists in 0.8.2 too"));
    // The commands that only the newer release has stay named as such.
    assert!(text.contains("`herdr machine`, `--machine` and `workspace close --group`"));
}

#[test]
fn keeps_the_session_identity_rules_and_the_state_words() {
    let text = reference();
    for line in [
        "`--session <name>`",
        "Never rely on the focused pane.",
        "`HERDR_ENV`",
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
