//! #2497: the MODE-AGENT and MODE-LEADER workflows as testable step lists, and every
//! `moai-cli <op> [--flag]` they name checked against the built binary's `--help`.
use std::{fs, path::Path, process::Command};

fn workflow(name: &str) -> String {
    fs::read_to_string(
        Path::new(env!("CARGO_MANIFEST_DIR"))
            .join("skills/moai/workflows")
            .join(name),
    )
    .unwrap()
}

/// The fixed keys of the `## Procedure` steps (`N. **key** — …`), in order; the step numbers
/// must run 1, 2, 3, … without a gap.
fn procedure_keys(text: &str) -> Vec<String> {
    let start = text.find("\n## Procedure\n").expect("a Procedure section");
    let body = &text[start + 1..];
    let body = &body[..body[3..].find("\n## ").map_or(body.len(), |i| i + 3)];
    let mut keys = Vec::new();
    for line in body.lines() {
        let Some((num, rest)) = line.split_once(". ") else {
            continue;
        };
        if num.is_empty() || !num.bytes().all(|b| b.is_ascii_digit()) {
            continue;
        }
        assert_eq!(
            num,
            (keys.len() + 1).to_string(),
            "step numbers run without a gap: {line}"
        );
        let key = rest
            .strip_prefix("**")
            .and_then(|r| r.split_once("** — "))
            .map(|(k, _)| k)
            .unwrap_or_else(|| panic!("step {num} starts with its **key** — : {line}"));
        assert!(
            !key.is_empty() && key.bytes().all(|b| b.is_ascii_lowercase() || b == b'-'),
            "{key}"
        );
        keys.push(key.to_string());
    }
    keys
}

#[test]
fn mode_agent_first_run_steps_are_fixed() {
    assert_eq!(
        procedure_keys(&workflow("mode-agent.md")),
        [
            "mode-only",
            "banner",
            "state",
            "session-id",
            "claim",
            "claim-verify",
            "heavy-checks",
            "pr-gate",
            "merge-train",
            "hand-off",
            "goal-lines",
            "keep-going",
            "self-pick",
            "wake",
            "report",
        ]
    );
}

#[test]
fn mode_leader_first_run_steps_are_fixed() {
    assert_eq!(
        procedure_keys(&workflow("mode-leader.md")),
        [
            "codex-refusal",
            "mode-only",
            "banner",
            "state",
            "answer-requests",
            "goal-line",
            "supervise",
            "pr-gate",
            "merge-train",
            "handoff-line",
            "lane-done",
        ]
    );
}

/// Each `` `moai-cli <op> [--flag …]` `` span in a workflow: (op, flags).
fn cli_spans(text: &str) -> Vec<(String, Vec<String>)> {
    let mut out = Vec::new();
    for (i, _) in text.match_indices("`moai-cli ") {
        let rest = &text[i + 1..];
        let span = &rest[..rest.find('`').expect("closed code span")];
        let mut words = span.split_whitespace().skip(1);
        let op = words.next().expect("an op after moai-cli").to_string();
        let flags = words
            .take_while(|w| w.starts_with("--"))
            .map(str::to_string)
            .collect();
        out.push((op, flags));
    }
    out
}

fn help(args: &[&str]) -> String {
    let out = Command::new(env!("CARGO_BIN_EXE_moai-cli"))
        .args(args)
        .arg("--help")
        .output()
        .unwrap();
    assert!(
        out.status.success(),
        "`moai-cli {} --help` failed: {}",
        args.join(" "),
        String::from_utf8_lossy(&out.stderr)
    );
    String::from_utf8_lossy(&out.stdout).into_owned()
}

/// AC-3: every op a mode workflow names exists in this binary, with every flag it shows.
#[test]
fn ops_named_by_the_mode_workflows_exist_with_their_flags() {
    let top = help(&[]);
    let mut seen = Vec::new();
    for name in ["mode-agent.md", "mode-leader.md"] {
        for (op, flags) in cli_spans(&workflow(name)) {
            assert!(
                top.lines()
                    .any(|l| l.split_whitespace().next() == Some(op.as_str())),
                "{name}: `moai-cli {op}` is not a subcommand of this binary"
            );
            let sub = help(&[op.as_str()]);
            for flag in &flags {
                assert!(
                    sub.contains(flag.as_str()),
                    "{name}: `moai-cli {op}` has no {flag}"
                );
            }
            seen.push(format!(
                "{op}{}",
                flags.iter().map(|f| format!(" {f}")).collect::<String>()
            ));
        }
    }
    seen.sort();
    seen.dedup();
    assert_eq!(
        seen,
        [
            "claim-verify",
            "claim-verify --hook",
            "goal-block",
            "goal-plan",
            "goal-show",
            "heavy-slot-acquire",
            "labels-apply",
            "labels-plan",
            "pr-confirm",
            "queue-free",
            "queue-next",
            "scope-check",
            "usage-limit"
        ],
        "the op list the workflows name"
    );
}

#[test]
fn spans_parse_ops_and_flags() {
    assert_eq!(
        cli_spans("run `moai-cli claim-verify --hook` and `moai-cli queue-next`"),
        [
            ("claim-verify".to_string(), vec!["--hook".to_string()]),
            ("queue-next".to_string(), vec![])
        ]
    );
}

/// The text of one `## Procedure` step (`N. **key** — …` up to the next step or section).
fn step(text: &str, key: &str) -> String {
    let head = format!("**{key}** — ");
    let start = text.find(&head).unwrap_or_else(|| panic!("step {key}"));
    let rest = &text[start..];
    let end = rest[head.len()..]
        .find("\n1")
        .into_iter()
        .chain(
            [
                "\n2", "\n3", "\n4", "\n5", "\n6", "\n7", "\n8", "\n9", "\n## ",
            ]
            .iter()
            .filter_map(|m| rest[head.len()..].find(m)),
        )
        .min()
        .map_or(rest.len(), |i| i + head.len());
    rest[..end].to_string()
}

fn assert_says(text: &str, key: &str, phrases: &[&str]) {
    let body = step(text, key);
    for p in phrases {
        assert!(body.contains(p), "{key} must say `{p}`:\n{body}");
    }
}

// #2515: lanes open their own pull requests, settle a simultaneous opening themselves,
// run one round as a cycle, name their waiting branches, and follow the U6 gate (no tip
// chasing); the leader workflow says the same.
#[test]
fn self_opening_rules_are_pinned_in_both_mode_workflows() {
    let agent = workflow("mode-agent.md");
    assert_says(
        &agent,
        "pr-gate",
        &[
            // #2612: the limit is N, read from decision D7 (until recorded, the leader's
            // latest announcement), no longer a literal four.
            "Below N, open one yourself; no leader nomination is needed",
            "`git merge-tree`",
            "among the N lowest open non-draft numbers",
            "`gh pr ready --undo`",
            "the author of the lowest-numbered draft marks it ready",
        ],
    );
    assert_says(
        &agent,
        "self-pick",
        &["One round runs `claim → worktree work → verify → push → pull request (when the pr-gate count allows) → report → next round`"],
    );
    assert_says(&agent, "report", &["· waiting for PR slot: #<issue> …"]);
    assert_says(
        &agent,
        "merge-train",
        &[
            "the latest-tip requirement is off",
            "Do not merge the tip to follow it",
        ],
    );
    let leader = workflow("mode-leader.md");
    assert_says(
        &leader,
        "pr-gate",
        &[
            "lanes open their own pull requests",
            "you no longer name the next lane by default",
        ],
    );
    assert_says(
        &leader,
        "merge-train",
        &[
            "the latest-tip requirement is off",
            "no conflict and no file both sides changed",
        ],
    );
    for (name, text) in [("mode-agent.md", &agent), ("mode-leader.md", &leader)] {
        for stale in ["names the one lane", "name the one lane", "at its own turn"] {
            assert!(!text.contains(stale), "{name} still says `{stale}`");
        }
    }
}

// #2515 (comment 6823975793): ten of twelve lanes ended their turn after `LANE-DONE` and
// stayed stopped. The round only ends on its end condition; the open-PR limit stops only
// the opening; a wait is not work; `LANE-DONE` is a progress record; a turn that must end
// arms a wake first; the leader restarts a stopped lane. Each sentence is pinned.
#[test]
fn the_round_keeps_going_until_its_end_condition() {
    let agent = workflow("mode-agent.md");
    assert_says(
        &agent,
        "keep-going",
        &[
            "A turn ends only when (i) the repository has no open pull request and no startable Issue (not on HOLD, not `needs:decision`, prerequisites closed)",
            "(ii) the user says stop",
            "In every other case do not end the turn",
            "A card's completion report closes the card, never the round.",
        ],
    );
    // (i) the limit does not block implementation.
    assert_says(
        &agent,
        "pr-gate",
        &["The limit stops only the opening — go on with the round and claim the next Issue"],
    );
    assert_says(
        &agent,
        "self-pick",
        &["with nothing above left and whatever the open-PR count, self-claim a ready Issue"],
    );
    assert!(!agent.contains("below the open-PR limit"));
    // A wait is not work, for CI and for a leader reply alike.
    assert_says(&agent, "heavy-checks", &["a wait is not work"]);
    assert_says(
        &agent,
        "claim",
        &["nor on a requested decision or handoff line"],
    );
    // (ii) LANE-DONE continues in the same turn.
    assert_says(
        &agent,
        "report",
        &["A `LANE-DONE` line is a progress record, not a stop: after writing it, go to the next pick in the same turn"],
    );
    // (iii) arm a wake before ending a turn.
    assert_says(
        &agent,
        "wake",
        &[
            "arm a wake before ending the turn, at most 300 seconds away",
            "`ScheduleWakeup`",
            "`decision: \"block\"`",
            "`LANE-BLOCKED <lane> wake no wake mechanism on this host`",
        ],
    );
    assert_says(
        &agent,
        "goal-lines",
        &["The goal is the round, not one card", "`loop_line`"],
    );
    // (iv) the leader restarts a stopped lane and answers requests within the round.
    let leader = workflow("mode-leader.md");
    assert_says(
        &leader,
        "supervise",
        &[
            "Restart, do not dispatch",
            "`순환 계속 — self-pick (1)부터`",
            "only when its input box is empty",
        ],
    );
    assert_says(
        &leader,
        "handoff-line",
        &[
            "Answer handoff requests, decision requests and `LANE-BLOCKED` lines within the supervision round",
            // Decision D1 (#1975 comment 6840449133) replaced the leader's renewal handoff.
            "A same-session renewal needs no handoff line",
        ],
    );
    // Audit 6824297146 F1/N1: self-pick states the same end condition, exclusions included.
    assert_says(
        &agent,
        "self-pick",
        &["no startable Issue (not on HOLD, not `needs:decision`, prerequisites closed) — the same end condition as the `keep-going` step"],
    );
    // Cause 8 (6824158448): no bare "no startable Issue"; expired foreign claims go up.
    assert_says(&agent, "state", &["expired claims of other sessions"]);
    assert_says(
        &agent,
        "self-pick",
        &[
            "Never report \"no startable Issue\" bare",
            "When an expired claim of another session blocks an otherwise ready Issue, send `LANE-BLOCKED <lane> #<n> expired claim of <session> — leader handoff needed` for each such Issue and take other work",
        ],
    );
    assert_says(
        &leader,
        "supervise",
        &[
            "ownerless — no claim, or an expired claim of a session that is not running",
            "hand it over with your handoff line",
        ],
    );
    let skill =
        fs::read_to_string(Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai/SKILL.md"))
            .unwrap();
    assert!(skill.contains("reporting a finished card and then stopping is not completion"));
}

/// #2612: the lane and leader rules from the bottleneck study (items 1-16). Each sentence is
/// pinned so deleting it fails this test. The decided values of D1, D2, D5, D6 and D7
/// (Issue #1975 comment 6840449133) replace the earlier "pending" wording.
#[test]
fn bottleneck_rules_are_pinned_in_both_mode_workflows() {
    let agent = workflow("mode-agent.md");
    // 1. no-op wake backoff (D6), unchanged lines are not reprinted.
    assert_says(
        &agent,
        "wake",
        &[
            "After three no-op wakes in a row on the same state, stop polling at the short interval",
            "arm one longer fallback wake at most 1,800 seconds away (decision D6",
            "without a running background watch, the wake stays at most 300 seconds away",
            "Do not print a `LANE-BLOCKED` or `LANE-DONE` line whose content is unchanged",
        ],
    );
    // 2. renewal by editing the own claim comment, or by a new line of the same session (D1).
    assert_says(
        &agent,
        "claim",
        &[
            "edit the lease time on your own claim comment before it expires",
            "A new claim line of the same session also renews, without a handoff line (decision D1",
            "when another session's claim stands before it, the new line does not renew",
            "Before arming a wake, read the lease of every claim you hold and extend each one that ends before the next wake",
        ],
    );
    assert!(!agent.contains("this renews the lease and is not a takeover"));
    // D2: the lane's timeout takeover and its three conditions.
    assert_says(
        &agent,
        "claim-verify",
        &[
            "A timeout takeover (decision D2",
            "the earlier holder's lease has expired",
            "your handoff request on the Issue has had no answer for 30 minutes",
            "the earlier holder's session is not in the terminal list",
            "`moai-claim-handoff: comment=<earlier claim id> to=<your session> by=timeout`",
            "`claim-verify` ignores a `by=timeout` line whose target lease is still live",
            "except the `by=timeout` line of decision D2 above",
        ],
    );
    // 3. the oldest pr-wait across lanes opens first; N read from one place (D7).
    assert_says(
        &agent,
        "pr-gate",
        &[
            "The open-PR limit N is read from one place: decision D7 (Issue #1975 comment 6840449133) sets it to eight",
            "post one line `pr-wait: <UTC time>` on its Issue",
            "when another lane's line is older, let that branch open first",
        ],
    );
    // 4. a slot wait is not a wake reason.
    assert_says(
        &agent,
        "self-pick",
        &["Waiting for a pull request slot alone is never a reason to arm a wake"],
    );
    // 5. the poll interval covers only unchanged state.
    assert!(agent.contains(
        "That interval applies only to re-reading a state that has not changed: claim, `claim-verify`, push and pull request follow one another in the same turn"
    ));
    // 6. a lost reservation posts no verdict.
    assert_says(
        &agent,
        "self-pick",
        &["when an earlier reservation names the same head, that lane holds it — withdraw and post no verdict"],
    );
    // 7. pre-push checks that only CI caught.
    assert_says(
        &agent,
        "pr-gate",
        &[
            "`python3 -m unittest tests.ci_scope_test`",
            "`cargo test --offline --locked --test ac_command_lint`",
            "`git interpret-trailers --parse` on each new commit message",
        ],
    );
    // 8. an unread trailer is fixed by a fresh branch, not by the leader.
    assert_says(
        &agent,
        "claim-verify",
        &[
            "commit the same change again on a new branch with a readable trailer and open a new pull request",
            "\"the same diff plus the trailer\"",
        ],
    );
    // 9. a handoff request names every open claim id before yours.
    assert_says(
        &agent,
        "claim-verify",
        &["A handoff request names every open claim id that stands before yours on the Issue"],
    );
    let leader = workflow("mode-leader.md");
    // 10, 11, 13, 14, 16 in supervise.
    assert_says(
        &leader,
        "supervise",
        &[
            "a lane that armed a wake has not ended its round, so send it nothing",
            "a lane whose last line is `LANE-BLOCKED` gets an answer to that line, never the restart line",
            "never send the same restart sentence twice in a row",
            "report it as `UNREACHABLE`",
            "record \"found = answered\" with both counts",
            "Before you open a question to the user, finish what is ready",
            "ask it in writing, on the Issue and on the affected lane's screen, and keep supervising",
            "open the question window only for an urgent matter that touches money, law, an external contract, secrets or real data (decision D5",
            "A decision you record goes both to the Issue or pull request comment and to the affected lane's screen",
        ],
    );
    // 12. one handoff line per claim id, announced only after `owner`.
    assert_says(
        &leader,
        "handoff-line",
        &[
            "post one handoff line per claim id, each as the first line of its own comment",
            "announce the handoff only when it answers `owner`",
            "except the `by=timeout` line of decision D2",
            "check each such line against those three conditions",
        ],
    );
    // 15. operating numbers change together; N from D7.
    assert_says(
        &leader,
        "pr-gate",
        &[
            "Keep the open-PR limit N, read from one place: decision D7 (Issue #1975 comment 6840449133) sets it to eight",
            "change the MODE-AGENT and MODE-LEADER text and your announcement in the same round",
        ],
    );
    // No pending wording left once the decisions are recorded.
    for (name, text) in [("mode-agent.md", &agent), ("mode-leader.md", &leader)] {
        for pending in [
            "6839380081",
            "until D6 is recorded",
            "until D7 is recorded",
            "not yet recorded",
            "four when none was made",
        ] {
            assert!(
                !text.contains(pending),
                "{name} still says a decision is pending: `{pending}`"
            );
        }
    }
}

/// #2581: project hooks are read from the session's working folder, so a lane session in a
/// linked worktree without the project settings file runs no project hook; the manual
/// `claim-verify` step is then the only boundary. The claim-verify step says so.
#[test]
fn claim_verify_names_the_worktree_session_without_project_hooks() {
    assert_says(
        &workflow("mode-agent.md"),
        "claim-verify",
        &[
            "The host reads project hooks from the session's working folder",
            "runs none of the project's hooks, so there this step is the only boundary",
            "`worktree_sessions: not_reached`",
        ],
    );
}

/// #2680 AC-2/AC-3: a gate that could not run is recorded with substitute evidence,
/// merges wait by default, and the merge-on-substitute question is asked once per gate.
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
    let leader = workflow("mode-leader.md");
    for text in [
        "merges wait for the gate by default",
        "a user decision for that one gate",
        "Ask it once per gate and cause",
    ] {
        assert!(leader.contains(text), "mode-leader.md lacks {text:?}");
    }
}

// #2668 S3: a lane renews its claims before a claude.ai usage limit stops it, leaves the
// resume to the host's own automatic continue, and the leader treats a `usage_limit` lane as
// held, not stopped.
#[test]
fn usage_limit_rules_are_pinned_in_both_mode_workflows() {
    let agent = workflow("mode-agent.md");
    assert_says(
        &agent,
        "wake",
        &[
            "run `moai-cli usage-limit \"$HOME\"`",
            "before any other write, renew every claim you hold to its `renew_until`",
            "`LANE-WAIT <lane> usage-limit until <renew_until>`",
            "Keep Claude Code's own automatic continue on (`autoContinueAtUsageLimit`",
            "arm no wake of your own for it",
            "Codex CLI documents neither a usage record nor an automatic continue: report this step `not_run` there",
        ],
    );
    let leader = workflow("mode-leader.md");
    assert_says(
        &leader,
        "handoff-line",
        &[
            "reports idle with cause `usage_limit`",
            "post no handoff line for its cards before that time",
        ],
    );
}

/// #2675: the state summary and the self-claim candidate come from `moai-cli queue-free`, and
/// an empty answer is said as `0건`, never inferred from missing output.
#[test]
fn state_and_claim_read_the_free_card_list() {
    let agent = workflow("mode-agent.md");
    assert_says(
        &agent,
        "state",
        &[
            "`moai-cli queue-free` returns these buckets in one read-only call (Issue #2675)",
            "when `free_count` is 0 say `0건`; an empty or failed output is not a count",
        ],
    );
    assert_says(
        &agent,
        "claim",
        &["Take the self-claim candidate from the `free` list of `moai-cli queue-free`"],
    );
}
