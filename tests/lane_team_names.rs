//! SPEC-20261125-agent-team-session-names (REQ-15 to REQ-27): the skill package says that a
//! session enters a role with `/moai agent` or `/moai leader`, that these two words set a mode and
//! take no instruction text, that a session started without a role flag renames itself with one
//! line and falls back to the short id, how a message is addressed, and that every step of the
//! flow goes by a message table whose rules (notice, agent to agent, channel order) live in
//! `references/flows/team-messages.md`. The tests read the package files of this checkout and pin
//! sentences whole; whitespace is collapsed so a re-wrapped line does not break a pin. Nothing is
//! run and no HOME is touched.
use std::{fs, path::Path};

fn raw(rel: &str) -> String {
    fs::read_to_string(
        Path::new(env!("CARGO_MANIFEST_DIR"))
            .join("skills/moai")
            .join(rel),
    )
    .unwrap_or_else(|e| panic!("{rel}: {e}"))
}

fn package(rel: &str) -> String {
    raw(rel).split_whitespace().collect::<Vec<_>>().join(" ")
}

fn assert_has(file: &str, text: &str, needle: &str) {
    assert!(
        text.contains(needle),
        "{file} lacks the sentence:\n  {needle}"
    );
}

const WORK: &str = "workflows/work.md";
const LEADER: &str = "workflows/leader.md";
const ENTRY: &str = "SKILL.md";
const TABLE: &str = "references/flows/team-messages.md";
const FLOW: &str = "references/flows/next-flow.md";

/// REQ-24: the first sentence of the agent workflow tells this agent from the base agents and from
/// a host subagent.
#[test]
fn the_agent_workflow_tells_this_agent_from_base_agents_and_subagents() {
    let work = raw(WORK);
    let first = work
        .lines()
        .find(|l| !(l.trim().is_empty() || l.starts_with('#')))
        .unwrap()
        .split_whitespace()
        .collect::<Vec<_>>()
        .join(" ");
    let sentence = first.split(". ").next().unwrap();
    for needle in [
        "`/moai agent` makes this session an agent",
        "not one of the skill's base agents",
        "not a host subagent",
    ] {
        assert!(
            sentence.contains(needle),
            "the first sentence of {WORK} lacks `{needle}`: {sentence}"
        );
    }
}

/// REQ-15, REQ-27: the skill entry and the route table pick WORK for `agent` and LEADER for
/// `leader`; `work` and `worker` are not route words; no `/moai work` form is left in the package.
#[test]
fn the_entry_and_the_route_table_name_agent_and_leader_and_not_work() {
    let entry = package(ENTRY);
    for needle in [
        "`/moai agent` picks WORK",
        "A request whose first word after the skill name is `agent` picks WORK, and one whose first word is `leader` picks LEADER",
        "`work` and `worker` are not route words, not even as an alias",
    ] {
        assert_has(ENTRY, &entry, needle);
    }
    let routes: serde_json::Value = serde_json::from_str(&raw("references/routes.json")).unwrap();
    assert_eq!(routes["routes"]["WORK"]["entry_word"], "agent");
    assert_eq!(routes["routes"]["LEADER"]["entry_word"], "leader");
    let mut checked = 0;
    for rel in [
        ENTRY,
        WORK,
        LEADER,
        TABLE,
        FLOW,
        "references/internal/git-safety.md",
        "references/internal/loop.md",
        "references/chains/index.md",
        "references/home/flow.md",
        "references/home/routing.md",
    ] {
        let text = package(rel);
        for form in ["/moai work", "$moai work", "/skill:moai work"] {
            assert!(!text.contains(form), "{rel} still has the form `{form}`");
        }
        checked += 1;
    }
    assert_eq!(checked, 10);
}

/// REQ-25, REQ-26: the mode words take no instruction text, and an agent waits for the registered
/// leader's message once its mode is set.
#[test]
fn the_mode_words_take_no_instruction_text_and_an_agent_waits_for_the_leader() {
    let entry = package(ENTRY);
    assert_has(
        ENTRY,
        &entry,
        "These two words set a mode and are not commands: when text follows one of them, set the mode, run nothing from that text and say in one line, in the user's conversation language, that instructions come as a message",
    );
    let work = package(WORK);
    assert_has(
        WORK,
        &work,
        "`/moai agent` sets a mode and is not a command: text after it is never run, and the session says in one line, in the user's conversation language, that instructions come as a message",
    );
    assert_has(
        WORK,
        &work,
        "While a leader is registered, wait for that leader's message once the mode is set",
    );
    let leader = package(LEADER);
    assert_has(
        LEADER,
        &leader,
        "`/moai leader` sets a mode and is not a command: text after it is never run",
    );
}

/// REQ-16, REQ-17: a session whose name was not set by a role flag gives the rename line once and
/// reads its name from the first line of `ListAgents`; the fallback is the short id and one ledger
/// note.
#[test]
fn a_plain_session_gets_the_rename_line_and_the_short_id_fallback() {
    let cases = [
        (
            WORK,
            "does not already start with `<project name>-agent-`",
            "`/rename <project name>-agent-<english name>`",
        ),
        (
            LEADER,
            "does not already end with `-leader`",
            "`/rename <project name>-leader`",
        ),
    ];
    for (file, start, line) in cases {
        let text = package(file);
        for needle in [
            start,
            "gives the user the one line",
            line,
            "once, in the user's conversation language",
            "first line of `ListAgents`",
            "go on under the name the",
            "short id",
            "write one ledger note that says so",
        ] {
            assert_has(file, &text, needle);
        }
    }
}

/// REQ-18: the leader is addressed by the session id of `moai session-register --show` matched to
/// a row of `ListAgents`; any other session by its listed name, or by the short id when two rows
/// share a name.
#[test]
fn the_address_rule_names_the_registered_id_and_the_short_id() {
    for (file, text) in [
        (WORK, package(WORK)),
        (TABLE, package(TABLE)),
        (FLOW, package(FLOW)),
    ] {
        assert_has(file, &text, "moai session-register --show --name leader");
    }
    let table = package(TABLE);
    assert_has(TABLE, &table, "matched to a row of `ListAgents`");
}

/// SPEC-20261125-session-names-project-prefix REQ-9, REQ-10: before a message goes to a session
/// other than the registered leader, the leader runs the read-only lane check with the listed name
/// and sends only on `lane`; two rows with one name get nothing. The table, the leader workflow and
/// the agent workflow say it the same way.
#[test]
fn a_message_goes_only_to_a_session_the_registry_calls_a_lane() {
    let table = package(TABLE);
    for needle in [
        "Before you send a message to any session other than the registered leader, run `moai session-register --lane-check <listed name>`",
        "send only when it answers `lane`",
        "on `not_lane` send nothing and write one ledger note that names the row and the answer",
        "When two rows of `ListAgents` carry the same name, send to neither and write one ledger note",
    ] {
        assert_has(TABLE, &table, needle);
    }
    let leader = package(LEADER);
    for needle in [
        "Before every send run `moai session-register --lane-check <listed name>` with the listed name of the receiving session and send only when it answers `lane`",
        "when two rows share the name send to neither",
    ] {
        assert_has(LEADER, &leader, needle);
    }
    assert_has(
        WORK,
        &package(WORK),
        "Any other session is addressed by its listed name once `moai session-register --lane-check <listed name>` has answered `lane`; when two rows share the name, send nothing and write one ledger note",
    );
}

/// REQ-11, REQ-12, decision D-8: the first action of a session that enters `/moai agent` or
/// `/moai leader` is `pwd` and the project name; an agent registers its own name when the registry
/// has no entry for it.
#[test]
fn the_first_action_of_a_role_is_the_folder_and_the_project_name() {
    for file in [WORK, LEADER] {
        let text = package(file);
        for needle in [
            "Before any other action run `pwd` and `moai session-register --names`",
            "`another repository`",
            "takes and writes nothing",
        ] {
            assert_has(file, &text, needle);
        }
    }
    let work = package(WORK);
    for needle in [
        "`moai session-register --show --name <its name>` answers `free`",
        "`moai session-register . --name <its name> --mode agent`",
    ] {
        assert_has(WORK, &work, needle);
    }
}

/// REQ-19, REQ-20, REQ-21, REQ-22: the message table names who sends and receives each message
/// with its one-line head, a message is a notice and the ledger is the record, an agent may ask
/// another agent, and the channel order is host message, Orca messaging, ledger note.
#[test]
fn the_message_table_and_its_rules_are_in_the_reference_file() {
    let table = package(TABLE);
    for kind in [
        "| Dispatch | leader | agent |",
        "| Question | agent | leader |",
        "| Answer | leader | agent |",
        "| Audit request | leader | auditor agent |",
        "| Verdict | auditor agent | leader |",
        "| Repair request | leader | author agent |",
        "| Landing notice | leader | agents concerned |",
        "| Completion | agent | leader |",
        "| Next card | leader | agent |",
        "`<KIND> <card> <state> head=<10 characters>`",
    ] {
        assert_has(TABLE, &table, kind);
    }
    for needle in [
        "A message is a notice and the ledger is the record. The receiver reads the ledger before it acts",
        "An agent may send a message to another agent to check a fact",
        "A verdict, a decision and a repair request go to the ledger as well",
        "Use the host's cross-session message first",
        "use the lane-terminal manager's own messaging (its adapter reference under `references/adapters/` names the send command, and the check command for the leader)",
        "write the ledger note alone",
        "`/clear` and screen reading stay with the lane-terminal manager",
    ] {
        assert_has(TABLE, &table, needle);
    }
    let (a, b) = (
        table
            .find("Use the host's cross-session message first")
            .unwrap(),
        table.find("write the ledger note alone").unwrap(),
    );
    assert!(a < b, "the channel order puts the host message first");
    // The one place that names the terminal manager's commands is its adapter reference
    // (`orca_optional_3005` keeps the word out of every other package file).
    let adapter = package("references/adapters/orca.md");
    for needle in ["`orca orchestration send`", "`orca orchestration check`"] {
        assert_has("references/adapters/orca.md", &adapter, needle);
    }
}

/// REQ-23: the flow reference points to the message table and stays within 8192 bytes, and the
/// source manifest lists the new reference file.
#[test]
fn the_flow_points_to_the_table_within_its_byte_budget() {
    assert_has(FLOW, &package(FLOW), "`flows/team-messages`");
    let bytes = fs::metadata(
        Path::new(env!("CARGO_MANIFEST_DIR"))
            .join("skills/moai")
            .join(FLOW),
    )
    .unwrap()
    .len();
    assert!(bytes <= 8192, "{FLOW} is {bytes} bytes");
    let manifest: serde_json::Value =
        serde_json::from_str(&raw("references/source-manifest.json")).unwrap();
    assert!(
        manifest["files"][TABLE]["sha256"].is_string(),
        "the source manifest lacks {TABLE}"
    );
}

/// REQ-27: the dispatch the leader workflow tells a leader to send is the dispatch line of the
/// message table.
#[test]
fn the_leader_dispatch_is_the_dispatch_line_of_the_table() {
    assert_has(
        LEADER,
        &package(LEADER),
        "The dispatch is the `DISPATCH <card> ready head=<tip>` line and the scope (`flows/team-messages`)",
    );
    assert_has(
        TABLE,
        &package(TABLE),
        "`DISPATCH <card> ready head=<integration tip>` and the scope",
    );
}
