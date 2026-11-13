//! `gate-check` (#2674 S1): the leader merge gate rules over one pull request snapshot.

use moai_cli::gate::{self, Snapshot};
use std::io::Read;

/// The leader merge gate (#2674 S1), read-only: decides from a pull request snapshot on stdin
/// (comments, check runs, changed files, merge result, Self-Audit result, user-confirmed
/// REVIEW.md PRs) whether every rule G1–G7 holds. Prints `pass`, or the rule that stopped and
/// why. Reads nothing else, posts nothing, merges nothing.
#[derive(clap::Args)]
pub struct Args {}

pub const OPERATION: &str = "gate.check";

pub fn run(_args: Args) -> Result<serde_json::Value, String> {
    let mut text = String::new();
    std::io::stdin()
        .take(4 * 1024 * 1024 + 1)
        .read_to_string(&mut text)
        .map_err(|_| "stdin is not readable UTF-8".to_string())?;
    if text.len() > 4 * 1024 * 1024 {
        return Err("snapshot larger than 4 MiB; refused".into());
    }
    let snapshot: Snapshot =
        serde_json::from_str(&text).map_err(|e| format!("invalid gate snapshot: {e}"))?;
    Ok(match gate::check(&snapshot) {
        Ok((base, overlap)) => serde_json::json!({
            "pr": snapshot.pr,
            "head": snapshot.head,
            "pass": true,
            "verdict_base": base,
            "overlap": overlap,
        }),
        Err(stop) => serde_json::json!({
            "pr": snapshot.pr,
            "head": snapshot.head,
            "pass": false,
            "rule": stop.rule,
            "reason": stop.reason,
        }),
    })
}
