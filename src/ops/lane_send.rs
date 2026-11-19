//! `lane-send` (card C-120 slice B, SPEC-20261119-lane-send-guard): send a text into a lane
//! terminal only when its rendered screen says `working` or `idle`.

/// Send a text to a lane terminal if its screen allows it
///
/// Reads the rendered screen of the Orca terminal `--terminal`, classifies it like `lane-screen`
/// and sends `--text` with `orca terminal send` only when the lane is `working` or `idle`. A
/// confirmation or choice dialog, a usage limit, an unknown screen or an unreachable terminal
/// sends nothing and exits non-zero naming the state and the one screen row it was read from;
/// a usage limit also names its reset time in UTC. The screen is read again just before the
/// send and a different state sends nothing. Enter is sent only with `--enter`. A text with a
/// control, line-separator or bidirectional character, an empty one, one starting with `-` or
/// one over 2000 characters is refused before Orca is called. Runs `orca` (`MOAI_ORCA_BIN`
/// overrides the program, `MOAI_ORCA_TIMEOUT_SECS` the 15 second limit per call).
#[derive(clap::Args)]
pub struct Args {
    /// The Orca terminal handle of the lane, such as `term_123`.
    #[arg(long, allow_hyphen_values = true)]
    terminal: String,
    /// The text to send, one line.
    #[arg(long, allow_hyphen_values = true)]
    text: String,
    /// Send Enter after the text.
    #[arg(long)]
    enter: bool,
    /// The time to take as now, `YYYY-MM-DDTHH:MM:SSZ` (default: the clock).
    #[arg(long)]
    now: Option<String>,
}

pub const OPERATION: &str = "lane.send";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    use moai_cli::lane_status::lane_send::{self, ProcessOrca, DEFAULT_TIMEOUT_SECS};

    let now = match args.now.as_deref() {
        Some(s) => {
            moai_cli::forge::policy::parse_utc(s).ok_or("--now must be YYYY-MM-DDTHH:MM:SSZ")?
        }
        None => std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .map_err(|_| "the clock is before 1970")?
            .as_secs() as i64,
    };
    let secs = std::env::var("MOAI_ORCA_TIMEOUT_SECS")
        .ok()
        .and_then(|v| v.parse::<u64>().ok())
        .filter(|s| (1..=300).contains(s))
        .unwrap_or(DEFAULT_TIMEOUT_SECS);
    let orca = ProcessOrca {
        program: std::env::var_os("MOAI_ORCA_BIN").unwrap_or_else(|| "orca".into()),
        timeout: std::time::Duration::from_secs(secs),
    };
    lane_send::send(&orca, &args.terminal, &args.text, args.enter, now)
}
