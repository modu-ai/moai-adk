//! `lane-screen` (card C-120, SPEC-20261119-lane-screen-guard): the state of a lane's rendered
//! screen; the logic is `moai_cli::lane_status::lane_screen`.

use std::io::Read;
use std::path::PathBuf;

/// Read a lane's screen into a state
///
/// Reads one lane screen on standard input, either an Orca `terminal read --screen --json`
/// answer or the screen as plain text, and prints its state (`working`, `idle`,
/// `permission_prompt`, `input_pending`, `usage_limit`, `unreachable` or `unknown`) with the one
/// screen row it was read from. For a usage limit it prints the reset time in UTC: from the
/// lane's `usage-limits.json` when `--home` is given and a window is at its limit, otherwise
/// from the screen text; a time with no zone or an unknown zone is `unknown` with its text.
/// Only the five states `lane-status` names carry a `pane` object. Reads standard input and,
/// with `--home`, one record; writes nothing and sends nothing.
#[derive(clap::Args)]
pub struct Args {
    /// The lane's HOME, to read its `usage-limits.json` for the reset time of a usage limit.
    #[arg(long)]
    home: Option<PathBuf>,
    /// The time to take as now, `YYYY-MM-DDTHH:MM:SSZ` (default: the clock).
    #[arg(long)]
    now: Option<String>,
}

pub const OPERATION: &str = "lane.screen";

/// Standard input is read up to this many bytes; a screen is a few kilobytes.
const INPUT_LIMIT: usize = 65_536;

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let mut bytes = Vec::new();
    std::io::stdin()
        .take(INPUT_LIMIT as u64 + 1)
        .read_to_end(&mut bytes)
        .map_err(|_| "stdin read failed")?;
    if bytes.len() > INPUT_LIMIT {
        return Err("input exceeds 64 KiB".into());
    }
    let input = String::from_utf8_lossy(&bytes);
    let now = match args.now.as_deref() {
        Some(s) => moai_cli::forge::policy::parse_utc(s)
            .ok_or("--now must be YYYY-MM-DDTHH:MM:SSZ")?,
        None => std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .map_err(|_| "the clock is before 1970")?
            .as_secs() as i64,
    };
    Ok(moai_cli::lane_status::lane_screen::report(
        &input,
        args.home.as_deref(),
        now,
    ))
}
