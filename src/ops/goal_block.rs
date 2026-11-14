//! `goal-block` (#2708, formerly tools/gen_goal_block.py): the `/goal` paste block.

use std::io::Read;

/// Build the goal paste block from the state
///
/// Build the `/goal` paste block from the state on stdin (#2163; #2708 moved it from
/// tools/gen_goal_block.py): `{milestone, issues, prs, keep_set}` in, the leader line, the
/// agent's lane-wide loop line and card lines, and the Codex line out. Reads only stdin.
#[derive(clap::Args)]
pub struct Args {}

pub const OPERATION: &str = "goal.block";

pub fn run(_args: Args) -> Result<serde_json::Value, String> {
    let mut text = String::new();
    std::io::stdin()
        .take(1024 * 1024 + 1)
        .read_to_string(&mut text)
        .map_err(|_| "invalid fixture JSON: stdin is not readable UTF-8".to_string())?;
    if text.len() > 1024 * 1024 {
        return Err("invalid fixture JSON: larger than 1 MiB; refused".into());
    }
    let fixture: serde_json::Value =
        serde_json::from_str(&text).map_err(|e| format!("invalid fixture JSON: {e}"))?;
    moai_cli::goal_block::build(&fixture)
}
