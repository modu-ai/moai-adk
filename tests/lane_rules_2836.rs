//! #2836 (leader decision 6869212685): the host's worktree guard (Claude Code, not a MoAI hook)
//! and the shapes it accepts are pinned in the shipped MODE-AGENT workflow, with the rule that a
//! check and the push or comment it guards are never sent as parallel calls.
use std::{fs, path::Path};

fn workflow(name: &str) -> String {
    fs::read_to_string(
        Path::new(env!("CARGO_MANIFEST_DIR"))
            .join("skills/moai/workflows")
            .join(name),
    )
    .unwrap()
}

fn says(text: &str, file: &str, phrases: &[&str]) {
    for p in phrases {
        assert!(text.contains(p), "{file} must say `{p}`");
    }
}

/// Every refusal kind measured in the lane sessions has a row, and the table keeps the line
/// between spelling a permitted effect differently and reaching a refused one.
#[test]
fn mode_agent_lists_every_host_guard_refusal_with_its_accepted_shape() {
    let agent = workflow("mode-agent.md");
    says(
        &agent,
        "mode-agent.md",
        &[
            "Host worktree guard (Issue #2836",
            "not a MoAI hook, whose refusals start with `MoAI tool-preflight:`",
            // The four refusal kinds of the #2836 sample, as the host words them.
            "| `names git in a form too complex to verify`",
            "| `gives worktree add a path that can't be verified`",
            "| `must target its own worktree`",
            "| `sets HOME, injecting git configuration`",
            // The accepted shapes.
            "One plain `git -C <absolute worktree path> <subcommand>` per command",
            "run as `zsh <absolute script path>`",
            "<absolute literal path>` — no variable, `~`, `..` or relative path",
            "`git -C <own worktree> show <ref>:<path>`",
            "`GIT_CONFIG_GLOBAL=/dev/null`, `GIT_CONFIG_SYSTEM=/dev/null`",
            // Never a bypass.
            "a refusal of the effect itself — a write in another worktree, the real HOME — stays a refusal",
        ],
    );
}

/// D4: a refused call does not stop the other calls of the same response, so a push, a comment
/// or a leader message that reports a check waits for the check's exit.
#[test]
fn mode_agent_keeps_a_check_apart_from_the_push_and_the_comments_it_guards() {
    let agent = workflow("mode-agent.md");
    says(
        &agent,
        "mode-agent.md",
        &[
            "Never send a check and the push it guards as parallel calls in one response",
            "The same holds for an Issue or pull request comment and a leader message that report the check",
            "a guard refusal stops only the refused call, so the other calls of the same response still run",
        ],
    );
    let verify = workflow("verify.md");
    says(
        &verify,
        "verify.md",
        &["the shapes it accepts are listed in the MODE-AGENT workflow (\"Host worktree guard\")"],
    );
}
