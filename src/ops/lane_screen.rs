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
///
/// With `--scan` it reads no standard input: it runs `orca terminal list --json`, reads the
/// screen of every listed terminal one at a time (`orca terminal read`, at most 200) and prints
/// one row per terminal (handle, title, branch, state, reason, evidence row, reset time in UTC
/// of a usage limit) with a count by state and a table. A terminal that cannot be read is
/// `unreachable` and the scan goes on; a list Orca reports as cut short is said so (`n of m`).
/// The reset time comes from the screen text only, so `--scan` cannot be combined with `--home`.
/// Runs `orca` (`MOAI_ORCA_BIN` overrides the program, `MOAI_ORCA_TIMEOUT_SECS` the 15 second
/// limit per call); exits non-zero when the list cannot be read.
#[derive(clap::Args)]
pub struct Args {
    /// The lane's HOME, to read its `usage-limits.json` for the reset time of a usage limit.
    #[arg(long)]
    home: Option<PathBuf>,
    /// Scan every terminal `orca terminal list` reports instead of reading one screen.
    #[arg(long, conflicts_with = "home")]
    scan: bool,
    /// The time to take as now, `YYYY-MM-DDTHH:MM:SSZ` (default: the clock).
    #[arg(long)]
    now: Option<String>,
}

pub const OPERATION: &str = "lane.screen";

/// Standard input is read up to this many bytes; a screen is a few kilobytes.
const INPUT_LIMIT: usize = 65_536;

fn clock(now: Option<&str>) -> Result<i64, String> {
    match now {
        Some(s) => Ok(
            moai_cli::forge::policy::parse_utc(s).ok_or("--now must be YYYY-MM-DDTHH:MM:SSZ")?
        ),
        None => Ok(std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .map_err(|_| "the clock is before 1970")?
            .as_secs() as i64),
    }
}

/// `--scan`: no standard input; the list and each screen come from `orca`.
fn scan(args: &Args) -> Result<serde_json::Value, String> {
    use moai_cli::lane_status::lane_screen as screen;
    use moai_cli::lane_status::lane_send::{read_checked, Orca, ProcessOrca};

    let now = clock(args.now.as_deref())?;
    let orca = ProcessOrca::from_env();
    let list = orca.list().map_err(|d| screen::list_error(&d))?;
    screen::scan(&list, &mut |handle| read_checked(&orca, handle), now)
}

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    if args.scan {
        return scan(&args);
    }
    let mut bytes = Vec::new();
    std::io::stdin()
        .take(INPUT_LIMIT as u64 + 1)
        .read_to_end(&mut bytes)
        .map_err(|_| "stdin read failed")?;
    if bytes.len() > INPUT_LIMIT {
        return Err("input exceeds 64 KiB".into());
    }
    let input = String::from_utf8_lossy(&bytes);
    let now = clock(args.now.as_deref())?;
    Ok(moai_cli::lane_status::lane_screen::report(
        &input,
        args.home.as_deref(),
        now,
    ))
}
