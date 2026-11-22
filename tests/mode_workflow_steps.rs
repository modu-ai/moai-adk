//! #2497: the MODE-AGENT and MODE-LEADER workflows as testable step lists. Both left the package
//! with their routes (SPEC-20261122-next-flow-skills-agents REQ-2); the file stays because
//! `skills/moai/references/rule-history.md` names it (D-11). What still holds is pinned here: the
//! WORK workflow's card take and blocked note, and the AUDIT, SKILL.md, core-rules, PROJECT and
//! UPDATE sentences the old tests also read.

use std::{fs, path::Path};

fn workflow(name: &str) -> String {
    fs::read_to_string(
        Path::new(env!("CARGO_MANIFEST_DIR"))
            .join("skills/moai/workflows")
            .join(name),
    )
    .unwrap()
}

// Retired with the MODE-AGENT route (SPEC-20261122-next-flow-skills-agents REQ-2): `mode_agent_first_run_steps_are_fixed` pinned the lane procedure keys; WORK pins its own in tests/next_flow_skill_scan.rs.

// Retired with the MODE-LEADER route (SPEC-20261122-next-flow-skills-agents REQ-2): `mode_leader_first_run_steps_are_fixed` pinned the leader procedure keys.

// Retired with the MODE-AGENT route (SPEC-20261122-next-flow-skills-agents REQ-2): `ops_named_by_the_mode_workflows_exist_with_their_flags` checked the ops both mode files named; the card commands WORK names are checked in tests/mode_cards_53.rs.

// Retired with the MODE-AGENT route (SPEC-20261122-next-flow-skills-agents REQ-2): `spans_parse_ops_and_flags` tested the span reader only the retired op check used.

// Retired with the MODE-LEADER route (SPEC-20261122-next-flow-skills-agents REQ-2): `workflows_record_a_dispatch_and_its_receipt_as_card_notes` pinned the leader dispatch and the lane receipt.

// Retired with the MODE-LEADER route (SPEC-20261122-next-flow-skills-agents REQ-2): `the_leader_registers_its_mode_for_the_status_line` pinned the leader session registration.

// Retired with the MODE-AGENT route (SPEC-20261122-next-flow-skills-agents REQ-2): `self_opening_rules_are_pinned_in_both_mode_workflows` pinned the pull-request opening gate and the merge train.

// Retired with the MODE-AGENT route (SPEC-20261122-next-flow-skills-agents REQ-2): `the_round_keeps_going_until_its_end_condition` pinned the keep-going, wake and supervise steps; WORK repeats until no card is left.

// Retired with the MODE-AGENT route (SPEC-20261122-next-flow-skills-agents REQ-2): `bottleneck_rules_are_pinned_in_both_mode_workflows` pinned the lane and leader bottleneck rules.

// Retired with the MODE-AGENT route (SPEC-20261122-next-flow-skills-agents REQ-2): `session_id_registers_renews_and_stops_on_a_held_name` pinned the session-id register and renew step.

// Retired with the MODE-AGENT route (SPEC-20261122-next-flow-skills-agents REQ-2): `claim_verify_names_the_worktree_session_without_project_hooks` pinned the lane claim-verify step.

// Retired with the MODE-AGENT route (SPEC-20261122-next-flow-skills-agents REQ-2): `claim_verify_names_the_codex_hook_trust_steps` pinned the Codex hook trust steps of the lane.

/// SPEC-20261118-skill-audit-followups REQ-2 (SDD-6 audit N2): the auditor works in its own
/// worktree or a clone inside it, never in the implementer's, and the reason is the worktree
/// key of `Moai-Session`.
#[test]
fn the_auditor_works_outside_the_implementers_worktree() {
    let audit = workflow("audit.md");
    let line = audit
        .lines()
        .find(|l| l.contains("never in the implementer's worktree"))
        .expect("audit.md says where the auditor works");
    for text in [
        "own worktree or a clone inside it",
        "worktree key",
        "`Moai-Session`",
        "`internal/commit-trailers`",
    ] {
        assert!(line.contains(text), "audit.md line lacks {text:?}: {line}");
    }
}

/// #2680 AC-2/AC-3: a gate that could not run is recorded with substitute evidence. The leader
/// half (merges wait, ask once per gate) left with the MODE-LEADER route
/// (SPEC-20261122-next-flow-skills-agents REQ-2).
#[test]
fn gate_unavailable_rules_are_written_where_they_are_used() {
    let audit = workflow("audit.md");
    for text in [
        "ci: gate_unavailable(<cause> since <since>)",
        "substitute: <command> exit <n> <log>",
        "Substitute evidence is a record, not a merge",
        "is not a gate",
    ] {
        assert!(audit.contains(text), "audit.md lacks {text:?}");
    }
}

// Retired with the MODE-AGENT route (SPEC-20261122-next-flow-skills-agents REQ-2): `usage_limit_rules_are_pinned_in_both_mode_workflows` pinned the usage-limit rules of the lane and the leader.

/// #2592 Codex parity: Codex spawns agents on a skill instruction that requests delegation, so
/// SKILL.md requests it in words and names each host's own subagent tool.
#[test]
fn skill_requests_delegation_with_each_hosts_subagent_tool() {
    let skill =
        fs::read_to_string(Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai/SKILL.md"))
            .unwrap();
    for needle in [
        "This skill requests delegation",
        "`## Base agents`",
        "`spawn_agent` and then `wait_agent` in Codex CLI",
        "the agent tool in Claude Code",
        "do the step in this session and say so",
    ] {
        assert!(skill.contains(needle), "SKILL.md lacks {needle}");
    }
}

/// #2675, re-pointed by C-53 S5 and then to WORK with the retired MODE-AGENT route
/// (SPEC-20261122-next-flow-skills-agents REQ-3): the card to take comes from one read-only
/// `moai card next`, and an empty or failed output is never read as a count; the name stays.
#[test]
fn state_and_claim_read_the_free_card_list() {
    let work = workflow("work.md");
    for p in [
        "Run `moai card next` and take the card its `next` names: a card in `mine` first (it is already yours), else the first `ready` one",
        "an empty or failed output is not a count",
    ] {
        assert!(work.contains(p), "work.md must say `{p}`");
    }
    assert!(!work.contains("queue-free"), "the Issue buckets are gone");
}

// Retired with the MODE-AGENT route (SPEC-20261122-next-flow-skills-agents REQ-2): `blocked_lines_name_kind_owner_and_next_check` pinned the lane report step and its AGENTS.md pointer, which goes stale.

// Retired with the MODE-LEADER route (SPEC-20261122-next-flow-skills-agents REQ-2): `codex_leader_timer_is_the_measured_orca_automation` pinned the Codex leader timer.

/// #2664 (inspection P2 items): E2 a rule change names the record it replaces and a quick
/// reversal is asked once; E3 the final report ends with the person's check list and the
/// decisions made for them. E1 (external pull requests) and E4 (the reduced mode) left with the
/// MODE-AGENT and MODE-LEADER routes (SPEC-20261122-next-flow-skills-agents REQ-2).
#[test]
fn inspection_p2_rules_are_written_where_they_are_used() {
    let rules = fs::read_to_string(
        Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai/references/internal/core-rules.md"),
    )
    .unwrap();
    for p in [
        "A rule change is a decision record that names the record it replaces",
        "reverses a rule recorded less than an hour earlier, ask the person who made it once, before applying it",
        "the checks to try themselves, every item touching a HOLD category or `guarded` setting first",
        "the decisions made for them, each with how to undo it",
        "Mark test data as test data",
    ] {
        assert!(rules.contains(p), "core-rules.md must say `{p}`");
    }
}

// Retired with the MODE-LEADER route (SPEC-20261122-next-flow-skills-agents REQ-2): `timeout_takeover_without_a_terminal_list_uses_claim_timeout_check` pinned the gate schedule and the reclaim wording of both modes.

// Retired with the MODE-AGENT route (SPEC-20261122-next-flow-skills-agents REQ-2): `agent_workflow_opens_with_the_leader_busy_summary_and_reads_screens_only_with_orca` pinned the leader-busy summary and the Orca screen reads.

// Retired with the MODE-AGENT route (SPEC-20261122-next-flow-skills-agents REQ-2): `waiting_is_a_line_and_the_next_job_and_an_armed_watch_is_idle` pinned the waiting line and the wake ladder.

// Retired with the MODE-AGENT route (SPEC-20261122-next-flow-skills-agents REQ-2): `a_lane_does_not_stop_before_its_first_claim_and_knows_where_to_report` pinned the lane first turn.

// Retired with the MODE-LEADER route (SPEC-20261122-next-flow-skills-agents REQ-2): `an_unrendered_screen_stops_the_typing_never_the_dispatch` pinned the Orca screen dispatch rules.

// Retired with the MODE-LEADER route (SPEC-20261122-next-flow-skills-agents REQ-2): `a_dispatch_says_to_claim_names_the_report_issue_and_forbids_nothing_extra` pinned the leader dispatch text.

// Retired with the MODE-LEADER route (SPEC-20261122-next-flow-skills-agents REQ-2): `a_leader_that_only_waits_arms_a_wake_and_never_marks_its_goal_blocked` pinned the leader wake.

// Retired with the MODE-AGENT route (SPEC-20261122-next-flow-skills-agents REQ-2): `codex_and_omp_lanes_must_print_their_wake_line_and_the_stop_hook_notice` pinned the Codex and OMP lane wake line.

// Retired with the MODE-LEADER route (SPEC-20261122-next-flow-skills-agents REQ-2): `a_dispatch_is_a_card_and_semi_keeps_only_four_things_for_the_person` pinned the leader dispatch and the autonomy levels.

/// #3005 N6, re-pointed to WORK with the retired modes (SPEC-20261122-next-flow-skills-agents
/// REQ-3): a card that needs a decision ends in a `LANE-BLOCKED kind=decision` note on the card,
/// never a question to the user and never a non-blocking question tool; the name stays.
#[test]
fn the_modes_never_ask_through_a_non_blocking_question_tool() {
    let skill =
        fs::read_to_string(Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai/SKILL.md"))
            .unwrap();
    for p in [
        "In WORK a card that needs a decision ends in `LANE-BLOCKED kind=decision …` as a note on the card",
        "never a question to the user and never a question through a non-blocking question tool",
    ] {
        assert!(skill.contains(p), "SKILL.md must say `{p}`");
    }
    let work = workflow("work.md");
    for p in [
        "write `LANE-BLOCKED kind=decision <card>",
        "as a note on the card, never a question to the user",
    ] {
        assert!(work.contains(p), "work.md must say `{p}`");
    }
    for tool in ["request_user_input", "AskUserQuestion"] {
        assert!(!work.contains(tool), "work.md names {tool}");
    }
}

// Retired with the MODE-LEADER route (SPEC-20261122-next-flow-skills-agents REQ-2): `a_leader_after_a_compaction_checks_before_it_repeats_a_user_instruction` pinned the leader compaction rule.

// Retired with the MODE-AGENT route (SPEC-20261122-next-flow-skills-agents REQ-2): `a_message_the_person_typed_into_a_lane_outranks_the_leaders_dispatch` pinned the lane rule against a leader dispatch.

// Retired with the MODE-AGENT route (SPEC-20261122-next-flow-skills-agents REQ-2): `a_lane_without_a_leader_watches_by_a_ladder_and_stops_at_the_idle_limit` pinned the wake ladder and the idle limit.

// Retired with the MODE-LEADER route (SPEC-20261122-next-flow-skills-agents REQ-2): `dispatch_is_batched_per_lane_without_sleep_turns_and_a_lost_card_moves_on` pinned the leader batched dispatch.

// Retired with the MODE-LEADER route (SPEC-20261122-next-flow-skills-agents REQ-2): `the_leader_looks_every_round_for_merges_that_skipped_the_gate` pinned the leader merge-gate sweep.

// Retired with the MODE-AGENT route (SPEC-20261122-next-flow-skills-agents REQ-2): `a_lane_registers_in_its_first_turn_and_renews_while_it_waits` pinned the session register and renew of the lane.

// Retired with the MODE-AGENT route (SPEC-20261122-next-flow-skills-agents REQ-2): `orca_mail_is_checked_for_delivery_and_focus_is_never_switched` pinned the Orca mail rules of both modes.

// Retired with the MODE-LEADER route (SPEC-20261122-next-flow-skills-agents REQ-2): `a_leader_with_orca_dispatches_through_a_run_and_detects_the_channel` pinned the leader Orca Run dispatch.

/// #3005 P10 and P11: OMP enters a skill with `/skill:moai <text>`. The mode entries and the OMP
/// lane wake left with the MODE-AGENT and MODE-LEADER routes
/// (SPEC-20261122-next-flow-skills-agents REQ-2); the PROJECT and UPDATE entries stay.
#[test]
fn omp_has_an_entry_spelling_and_a_wake_that_a_lane_can_end() {
    let skill =
        fs::read_to_string(Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai/SKILL.md"))
            .unwrap();
    // SKILL.md no longer names OMP (the 2026-11-20 guidance sweep).
    for p in ["in OMP", "the task tool in OMP"] {
        assert!(!skill.contains(p), "SKILL.md must not say `{p}` any more");
    }
    for (file, key) in [
        ("project.md", "`/skill:moai project` in OMP"),
        ("update.md", "`/skill:moai update` in OMP"),
    ] {
        assert!(workflow(file).contains(key), "{file} must say `{key}`");
    }
}

// Retired with the MODE-AGENT route (SPEC-20261122-next-flow-skills-agents REQ-2): `a_lane_with_an_orca_dispatch_reports_by_orca_as_a_second_path` pinned the lane Orca report path.

// Retired with the MODE-LEADER route (SPEC-20261122-next-flow-skills-agents REQ-2): `the_observations_example_is_a_file_lane_safe_to_reset_reads` pinned the leader observations example.

// Retired with the MODE-LEADER route (SPEC-20261122-next-flow-skills-agents REQ-2): `the_leader_pulse_is_built_by_the_cli_and_sent_by_the_skill` pinned the leader pulse.

// Retired with the MODE-AGENT route (SPEC-20261122-next-flow-skills-agents REQ-2): `a_message_is_received_when_its_ack_line_is_read_and_orca_abilities_are_judged_apart` pinned the Orca acknowledgement rules of both modes.
