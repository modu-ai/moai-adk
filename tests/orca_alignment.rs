//! SPEC-20261120-orca-1-4-267-alignment: the Orca rules of the adapter text follow what Orca's
//! official pages and its installed CLI say, and every rule cites its page or local help. The
//! MODE-AGENT and MODE-LEADER halves left with their routes (SPEC-20261122-next-flow-skills-agents
//! REQ-2). Reads the shipped text; writes nothing.

use std::path::Path;

const ADAPTER: &str = "skills/moai/references/adapters/orca.md";

fn read(rel: &str) -> String {
    std::fs::read_to_string(Path::new(env!("CARGO_MANIFEST_DIR")).join(rel))
        .unwrap_or_else(|e| panic!("{rel}: {e}"))
}

/// Collapse every run of whitespace to one space, so a phrase may wrap across lines.
fn flat(text: &str) -> String {
    text.split_whitespace().collect::<Vec<_>>().join(" ")
}

fn must_say(file: &str, phrases: &[&str]) {
    let text = flat(&read(file));
    for phrase in phrases {
        assert!(text.contains(phrase), "{file} must say {phrase:?}");
    }
}

// Retired with the MODE-AGENT route (SPEC-20261122-next-flow-skills-agents REQ-2): `resend_rule_is_gone_from_both_mode_texts` scanned the two mode files for the old resend rule.

// Retired with the MODE-AGENT route (SPEC-20261122-next-flow-skills-agents REQ-2): `resend_a_null_delivered_at_is_read_as_not_yet_known` pinned the delivered_at rule of the two mode files.

// Retired with the MODE-AGENT route (SPEC-20261122-next-flow-skills-agents REQ-2): `rules_mail_states_are_written_as_shown_and_the_rest_as_unconfirmed` pinned the mail-state rule of the two mode files.

// Retired with the MODE-AGENT route (SPEC-20261122-next-flow-skills-agents REQ-2): `rules_a_watcher_peeks_and_leaves_the_default_check_to_the_inbox_owner` pinned the peek rule of the two mode files.

// Retired with the MODE-AGENT route (SPEC-20261122-next-flow-skills-agents REQ-2): `rules_supervised_work_is_run_task_dispatch_and_a_notice_is_mail` pinned the Run, Task and Dispatch rule of the two mode files.

// Retired with the MODE-AGENT route (SPEC-20261122-next-flow-skills-agents REQ-2): `rules_an_unclear_send_is_followed_by_its_request_id` pinned the request-id rule of the two mode files.

#[test]
fn adapter_names_the_pages_the_local_version_and_what_is_not_documented() {
    must_say(
        ADAPTER,
        &[
            "https://orca.dev/docs/cli/orchestration",
            "https://orca.dev/docs/cli/reference",
            "https://orca.dev/docs/agents/native-chat",
            "https://orca.dev/docs/cli/automations",
            "`orca --version` printed 1.4.267",
            "The pages do not document",
            "`delivered_at`",
            "any mail state",
        ],
    );
    let text = read(ADAPTER);
    let times = text
        .split(|c: char| !(c.is_ascii_digit() || c == ':' || c == 'Z' || c == '-' || c == 'T'))
        .filter(|t| t.len() == 20 && t.starts_with("2026-11-20T22:4") && t.ends_with('Z'))
        .count();
    assert!(
        times >= 4,
        "the adapter names the read times of the pages: {times}"
    );
}

#[test]
fn adapter_marks_what_the_pages_leave_out_unverified_and_keeps_ownership() {
    must_say(
        ADAPTER,
        &[
            "A missing `hostScope`, native chat and sleep or restore are unverified",
            "a lane keeps its card and its reservation while a fact is unverified",
        ],
    );
}

#[test]
fn adapter_says_what_a_screen_read_and_a_cursor_are() {
    must_say(
        ADAPTER,
        &[
            "a screen read is the frame that is rendered now",
            "a cursor of `terminal read` is a position in the output stream",
        ],
    );
}

#[test]
fn cites_the_new_rules_name_a_page_or_the_local_help_and_no_card_or_issue() {
    // The two mode files left with their routes (SPEC-20261122-next-flow-skills-agents REQ-2).
    for file in [ADAPTER] {
        let text = read(file);
        let mut seen = 0;
        for line in text.lines().filter(|l| {
            l.contains("--peek") || l.contains("null` `delivered_at`") || l.contains("hostScope")
        }) {
            seen += 1;
            let words: Vec<&str> = line.split_whitespace().collect();
            let hash = words.iter().any(|w| {
                w.trim_start_matches(['(', '`'])
                    .strip_prefix('#')
                    .is_some_and(|rest| rest.chars().take(3).all(|c| c.is_ascii_digit()))
            });
            let card = words.windows(2).any(|w| {
                (w[0] == "card" || w[0] == "Issue" || w[0] == "decision")
                    && w[1]
                        .trim_matches(['(', ')', ',', '.'])
                        .chars()
                        .any(|c| c.is_ascii_digit())
            });
            assert!(!hash && !card, "{file} cites a card or an Issue: {line}");
            assert!(
                line.contains("orca.dev/docs")
                    || line.contains("--help")
                    || line.contains("orca skills get orchestration")
                    || line.contains("the pages")
                    || line.contains("The pages"),
                "{file} names no page or local help here: {line}"
            );
        }
        assert!(seen >= 1, "{file} has no new rule line");
    }
}
