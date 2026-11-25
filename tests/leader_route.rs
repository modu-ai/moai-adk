//! SPEC-20261124-moai-leader-route: the skill has a ninth route, LEADER, whose workflow makes the
//! window that types `/moai leader` the registered coordination window of the project: it reads
//! the registry entry first, registers on `free`, stops on a held entry, renews once per round,
//! reads the ledger, sends audits to another session, lands only after a `PASS` for the exact head
//! and releases the registration when it ends. The tests read the package files of this checkout
//! and pin each sentence whole; whitespace is collapsed so a re-wrapped line does not break a pin.
//! Nothing is run and no HOME is touched.
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

const LEADER: &str = "workflows/leader.md";

fn leader() -> String {
    package(LEADER)
}

fn assert_has(text: &str, needle: &str) {
    assert!(
        text.contains(needle),
        "{LEADER} lacks the sentence:\n  {needle}"
    );
}

fn assert_before(text: &str, first: &str, second: &str) {
    let a = text
        .find(first)
        .unwrap_or_else(|| panic!("{LEADER} lacks: {first}"));
    let b = text
        .find(second)
        .unwrap_or_else(|| panic!("{LEADER} lacks: {second}"));
    assert!(a < b, "{LEADER}: `{first}` must come before `{second}`");
}

/// The text of the failure table's section (`## When a head does not land`), up to the next one.
fn table_section(text: &str) -> &str {
    let start = text
        .find("## When a head does not land")
        .expect("the table section");
    let rest = &text[start..];
    &rest[..rest.find("\n## ").map_or(rest.len(), |i| i + 1).max(1)]
}

/// The cells of every body row of the failure table (the header and the rule line are left out).
fn table_rows(section: &str) -> Vec<Vec<&str>> {
    section
        .lines()
        .filter(|l| l.starts_with("| ") && !l.starts_with("| What") && !l.starts_with("| ---"))
        .map(|l| l.trim_matches('|').split('|').map(str::trim).collect())
        .collect()
}

/// REQ-10: what the window is, and that it holds no standing role.
const INTRO: &str = "`/moai leader` makes the window that types it the registered coordination window of the project; the window holds no standing role, and when it ends the lanes go on from the ledger alone.";
/// REQ-3: the registry is read first.
const READ_FIRST: &str = "Read the entry with `moai session-register --show --name leader` first.";
/// REQ-3: the registration, on `free` only.
const REGISTER: &str = "When it answers `free`, register this session with `moai session-register . --name leader --mode leader` and keep the `instance` the answer returns.";
/// REQ-3: an entry this window already holds is renewed, never registered a second time.
const OWN_ENTRY: &str = "When it answers `held` and its `session` is this window's own host session id, register nothing again: renew that entry as step 2 says, with the `instance` the answer shows.";
/// REQ-4: a held entry or a conflict stops the route; any other refusal is an error.
const STOP: &str = "When another session holds the entry, or the registration is refused with `conflict`, stop with one line that names the holder's lease end and register nothing; a refusal of any other kind is an error to report, not an answer to who leads.";
/// REQ-5: the renewal, once per round, before the ledger is read; the mode is stored only when
/// given, so every renewal passes it again (D-7).
const RENEW: &str = "Once per round, before you read the ledger, renew the registration with `moai session-register . --name leader --mode leader --instance <instance>`; the mode is stored only when given, so pass `--mode leader` with every renewal.";
/// REQ-6: the ledger is the input, a completion message a notice.
const INPUT: &str = "Read the ledger with `moai card inbox`, `moai card wait` and `moai card next`; a lane's completion message is a notice that you check against the ledger before you act, as the \"Completion message\" and \"Transition\" sections of `flows/next-flow` say.";
/// D-3: the dispatch channel.
const DISPATCH: &str = "Send the next card's dispatch with the host's message tool; no `moai` command sends a cross-session message yet.";
/// C-521: the host's message comes first; `moai lane-send` and the lane-terminal manager are the
/// fallback when that channel is missing, refused or held, and otherwise stay for `/clear` and
/// screen reads.
const DISPATCH_FALLBACK: &str = "Only when that channel is missing, refused or held, send with `moai lane-send` (it reads the screen first and sends only to a lane that is working or idle) or with the lane-terminal manager's own messaging";
const DISPATCH_CLEAR: &str =
    "The lane-terminal manager otherwise stays for `/clear` and for reading a screen.";
/// C-521: each step names its command and what comes next.
const PLAN: &str = "Then run `moai card orchestrate plan`: it reads the ledger and the registry and changes nothing";
const ASSIGN: &str =
    "Name the auditor with `moai card assign <card> --auditor <session id> --by leader`";
const AUDIT_REQUEST: &str = "then send `AUDIT-REQUEST <card> ready head=<head to audit>` with the caps and the receipt to cite";
const VERDICT_OWN: &str = "The auditor records the verdict with `moai card verdict` from its own session; you read it in the ledger and never write it for the auditor.";
const BATCH: &str = "Use `moai card land-batch C-n=<head> ... --by leader` for 2 to 4 audited heads, which runs the wide check once on their join, and `moai card land` alone only for one head with a risk or an exception.";
const LANDED: &str = "send `LANDED <card> landed head=<10 hex>` to the lane concerned";
/// SPEC-20261125-land-sync-merge REQ-8: a branch that is not a card head enters only through
/// `moai card land-sync`, and the sentence says who calls it, what it makes and refuses.
const LAND_SYNC: &str = "A branch that is not a card head (such as `main` coming back after a release) enters only through `moai card land-sync <branch> --spec <SPEC-ID> --by leader`, called from this leader session for a branch that `work_flow.sync_sources` lists and an approved SPEC (an empty list refuses all): the command makes the merge commit, runs the wide check itself, moves the branch by one compare-and-swap and refuses `stale_base` and the check codes without moving;";
/// SPEC-20261125-land-sync-merge REQ-8: the first cell of the failure table's land-sync row.
const LAND_SYNC_CMD: &str = "`moai card land-sync`";
/// REQ-7: the audit goes to a session that wrote none of the change.
const AUDIT: &str = "Send the audit of a head to a session that wrote none of the change (the AUDIT route); the author of a change never audits it.";
/// REQ-7: landing after a `PASS` for the exact head, by the leader.
const LAND: &str = "Land that head with `moai card land <card> --head <40 hex> --by leader` only after a verdict `PASS` for that exact head; the command accepts only `--by leader` today, so this window is the caller.";
/// REQ-8: a decision note is sorted in the order of the skill entry's paragraph, which the text
/// points to and does not copy (D-8).
const DECIDE: &str = "A `LANE-BLOCKED kind=decision` note is sorted in the order of the skill entry's paragraph that begins \"Sort an open question\", case by case: this window decides and records as a note on the card what that order leaves to the leader, and what it sends to the user goes to the user as a selection with the recommended option first and waits for the pick.";
/// REQ-9: the release at the end.
const RELEASE: &str = "When the route ends, release the registration with `moai session-register . --name leader --release --instance <instance>`.";
/// REQ-9: a closed window releases nothing.
const CLOSED: &str = "A window that is closed releases nothing: its lease ends by itself (four hours by default), and the lanes then write the ledger note only.";
/// REQ-9: the user can free the entry of a closed window before its lease ends.
const FREED_BY_LEADER: &str = "Until then the entry can be freed with `moai session-register . --name leader --release --by leader --instance <instance>`, taking the `instance` from `moai session-register --show --name leader`.";

/// The procedure's step keys, in order.
const STEPS: [&str; 6] = ["register", "round", "dispatch", "audit", "land", "end"];

#[test]
fn register_reads_the_entry_first_and_registers_on_free() {
    let t = leader();
    for s in [READ_FIRST, REGISTER] {
        assert_has(&t, s);
    }
    assert_before(&t, READ_FIRST, REGISTER);
}

#[test]
fn register_stops_on_a_held_entry_or_a_conflict() {
    assert_has(&leader(), STOP);
}

#[test]
fn register_renews_its_own_held_entry_instead_of_registering_again() {
    let t = leader();
    for s in [REGISTER, OWN_ENTRY, STOP] {
        assert_has(&t, s);
    }
    // The own-entry case is read before the stop on a held entry, which names another session.
    assert_before(&t, REGISTER, OWN_ENTRY);
    assert_before(&t, OWN_ENTRY, STOP);
}

#[test]
fn round_renews_once_before_it_reads_the_ledger() {
    let t = leader();
    for s in [RENEW, INPUT] {
        assert_has(&t, s);
    }
    assert_before(&t, RENEW, INPUT);
}

#[test]
fn round_points_to_the_flow_reference_and_repeats_none_of_it() {
    let t = leader();
    assert_has(
        &t,
        "Read the flow reference `flows/next-flow` before the first round",
    );
    let flow = package("references/flows/next-flow.md");
    // Whole sentences of the flow reference's two sections stay there (parent REQ-11).
    for s in [
        "When the leader gets the message, it reads the ledger first",
        "A message the ledger does not back starts nothing",
        "A lane never sends `/clear` to itself",
        "The leader sends `/clear` to the exact existing window only after",
    ] {
        assert!(flow.contains(s), "next-flow lost the sentence: {s}");
        assert!(!t.contains(s), "{LEADER} repeats a flow sentence: {s}");
    }
}

#[test]
fn round_sends_the_dispatch_by_a_channel_that_exists() {
    let t = leader();
    for s in [DISPATCH, DISPATCH_FALLBACK, DISPATCH_CLEAR] {
        assert_has(&t, s);
    }
    // The host's message is named first; the terminal manager is the fallback, never the lead.
    assert_before(&t, DISPATCH, DISPATCH_FALLBACK);
    assert!(
        !t.contains("or with `moai lane-send` when the lanes run in terminals"),
        "{LEADER} puts `moai lane-send` on a par with the host's message"
    );
}

/// C-521 criteria 1 and 4: a round names its plan command, the auditor assignment, the audit
/// request, the one landing sentence for a single head and for a batch, and the landing notice,
/// in step order, and every step but the last says where it goes next.
#[test]
fn round_names_one_command_per_step_and_the_next_action() {
    let t = leader();
    for s in [PLAN, ASSIGN, AUDIT_REQUEST, VERDICT_OWN, BATCH, LANDED] {
        assert_has(&t, s);
    }
    assert_before(&t, PLAN, DISPATCH);
    assert_before(&t, DISPATCH, ASSIGN);
    assert_before(&t, ASSIGN, LAND);
    assert_before(&t, LAND, BATCH);
    assert_before(&t, BATCH, LANDED);
    let text = raw(LEADER);
    let start = text.find("## Procedure").unwrap();
    let rest = &text[start..];
    let body = &rest[..rest.find("\n## When a head").unwrap()];
    let steps: Vec<&str> = body
        .lines()
        .filter(|l| {
            l.split_once(". **")
                .is_some_and(|(n, _)| n.trim().parse::<usize>().is_ok())
        })
        .collect();
    for line in &steps[1..5] {
        assert!(
            line.contains("Next: ") || line.contains("Then do steps"),
            "a step names no next action: {line}"
        );
    }
}

/// C-521 criterion 2: the branches after a `fail`, a second `fail`, `no_pass`, `stale_base` and
/// `tip_moved` are a table, each row with a next action and no row without one.
#[test]
fn a_table_gives_the_next_action_of_each_failed_landing() {
    let text = raw(LEADER);
    let section = table_section(&text);
    let rows = table_rows(section);
    for key in [
        "a first `fail` verdict",
        "a second `fail` of the same card",
        "`no_pass` from `moai card land`",
        "`stale_base`",
        "`tip_moved`",
    ] {
        let row = rows
            .iter()
            .find(|r| r[0] == key)
            .unwrap_or_else(|| panic!("the table has no row `{key}`"));
        assert_eq!(row.len(), 3, "{key}");
        assert!(!row[2].is_empty(), "{key} has no next action");
    }
    for r in &rows {
        assert!(r.iter().all(|c| !c.is_empty()), "an empty cell: {r:?}");
    }
    // A repair goes by `moai card reopen` of the author, never of the leader.
    assert!(section.contains(
        "the author takes the card back with `moai card reopen <card> --generation <g>`"
    ));
}

#[test]
fn round_sends_the_audit_to_another_session_and_lands_only_after_a_pass() {
    let t = leader();
    for s in [AUDIT, LAND] {
        assert_has(&t, s);
    }
    assert_before(&t, AUDIT, LAND);
}

#[test]
fn round_lands_through_one_command_and_never_without_a_pass() {
    let t = leader();
    // The landing instruction is the command with its arguments; the Completion Check only names
    // the command to say when a head counts as landed.
    assert_eq!(
        t.matches("moai card land <card>").count(),
        1,
        "{LEADER} holds exactly one landing command"
    );
    // Every sentence that gives the landing command carries the `PASS` condition.
    for sentence in t
        .split(". ")
        .filter(|s| s.contains("moai card land <card>"))
    {
        assert!(
            sentence.contains("only after a verdict `PASS` for that exact head"),
            "{LEADER}: a landing sentence without the PASS condition: {sentence}"
        );
    }
    // No sentence lands a head by another condition.
    for opposite in [
        "without a verdict",
        "without waiting",
        "before the verdict",
        "no verdict",
        "regardless of the verdict",
    ] {
        assert!(!t.contains(opposite), "{LEADER} says `{opposite}`");
    }
}

#[test]
fn round_sorts_a_decision_note_by_pointing_to_the_open_question_order() {
    let t = leader();
    assert_has(&t, DECIDE);
    // The skill entry still holds the paragraph the text points to (a pointer must not dangle).
    assert!(
        package("SKILL.md").contains("Sort an open question before asking anyone, in this order"),
        "SKILL.md lost the open-question paragraph that {LEADER} points to"
    );
    // The revision 1 summary is not copied back (D-8).
    assert!(
        !t.contains("a reversible choice of category none is decided here"),
        "{LEADER} copies the open-question cases"
    );
}

#[test]
fn round_steps_run_in_order() {
    let text = raw(LEADER);
    let start = text.find("## Procedure").expect("a Procedure section");
    let rest = &text[start + "## Procedure".len()..];
    let body = &rest[..rest.find("\n## ").unwrap_or(rest.len())];
    let mut keys: Vec<(usize, String)> = Vec::new();
    for line in body.lines() {
        let Some((num, tail)) = line.split_once(". **") else {
            continue;
        };
        let Ok(n) = num.trim().parse::<usize>() else {
            continue;
        };
        let key = tail.split("**").next().unwrap_or("").to_string();
        keys.push((n, key));
    }
    let numbers: Vec<usize> = keys.iter().map(|(n, _)| *n).collect();
    let names: Vec<&str> = keys.iter().map(|(_, k)| k.as_str()).collect();
    assert_eq!(names, STEPS, "the LEADER steps");
    assert_eq!(numbers, (1..=STEPS.len()).collect::<Vec<_>>());
}

#[test]
fn end_releases_the_registration() {
    assert_has(&leader(), RELEASE);
}

#[test]
fn end_a_closed_window_releases_nothing() {
    assert_has(&leader(), CLOSED);
}

#[test]
fn end_a_closed_window_can_be_freed_by_the_leader() {
    let t = leader();
    for s in [CLOSED, FREED_BY_LEADER] {
        assert_has(&t, s);
    }
    assert_before(&t, CLOSED, FREED_BY_LEADER);
}

#[test]
fn end_the_window_holds_no_standing_role() {
    assert_has(&leader(), INTRO);
}

/// SPEC-20261125-land-sync-merge AC-8: step 5 holds the land-sync sentence exactly once.
#[test]
fn land_sync_sentence_is_in_step_5_exactly_once() {
    let t = leader();
    assert_has(&t, LAND_SYNC);
    assert_eq!(
        t.matches(LAND_SYNC).count(),
        1,
        "{LEADER} holds the land-sync sentence more than once"
    );
}

/// SPEC-20261125-land-sync-merge AC-8: the sentence follows the single-head landing and the batch
/// landing, and comes before the landing notice.
#[test]
fn land_sync_sentence_comes_after_land_and_batch_and_before_landed() {
    let t = leader();
    assert_before(&t, LAND, BATCH);
    assert_before(&t, BATCH, LAND_SYNC);
    assert_before(&t, LAND_SYNC, LANDED);
}

/// SPEC-20261125-land-sync-merge AC-8: the failure table has one land-sync row, as wide as the
/// header, with no empty cell.
#[test]
fn land_sync_has_one_complete_row_in_the_failure_table() {
    let text = raw(LEADER);
    let section = table_section(&text);
    let header: Vec<&str> = section
        .lines()
        .find(|l| l.starts_with("| What"))
        .expect("the table header")
        .trim_matches('|')
        .split('|')
        .map(str::trim)
        .collect();
    let rows = table_rows(section);
    let sync: Vec<&Vec<&str>> = rows
        .iter()
        .filter(|r| r[0].contains(LAND_SYNC_CMD))
        .collect();
    assert_eq!(sync.len(), 1, "the table needs one row for {LAND_SYNC_CMD}");
    let row = sync[0];
    assert_eq!(row.len(), header.len(), "the land-sync row's cell count");
    assert_eq!(row.len(), 3, "the land-sync row's cell count");
    assert!(row.iter().all(|c| !c.is_empty()), "an empty cell: {row:?}");
}

#[test]
fn end_workflow_stays_within_its_byte_budget() {
    let n = raw(LEADER).len();
    assert!(n <= 10240, "{LEADER} is {n} bytes, over 10240");
}
