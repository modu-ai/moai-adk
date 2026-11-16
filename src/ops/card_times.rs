//! `card-times` (#3067 S6): when each stage of a card and of its pull requests happened. The
//! logic is `moai_cli::card_times`.

use moai_cli::forge::policy::SystemClock;
use moai_cli::forge::transport::GhTransport;
use std::io::Read;

/// Read the server times of cards and their pull requests
///
/// Read the stage times of 1 to 50 cards from the request on stdin
/// (`{schema_version:1, owner, repo, issues:[…]}`): per card the dispatch times (sent, received,
/// decided, started), the first claim, the `pr-wait:` lines, and per pull request when it opened,
/// became ready, queued and ran CI, was reserved and judged for audit, was repaired and merged.
/// Each stage is `observed` (a server time), `absent` (the read finished with no record) or
/// `unknown` (a cut read, a time in the future, a spent budget); a span has a length only when
/// both ends are observed. Comments of non-members are ignored and listed. GraphQL reads only;
/// writes nothing.
#[derive(clap::Args)]
pub struct Args {}

pub const OPERATION: &str = "card.times";

pub fn run(_args: Args) -> Result<serde_json::Value, String> {
    let mut text = String::new();
    std::io::stdin()
        .take(1024 * 1024 + 1)
        .read_to_string(&mut text)
        .map_err(|_| "invalid request JSON: stdin is not readable UTF-8".to_string())?;
    if text.len() > 1024 * 1024 {
        return Err("invalid request JSON: larger than 1 MiB; refused".into());
    }
    let request: moai_cli::card_times::Request =
        serde_json::from_str(&text).map_err(|e| format!("invalid request JSON: {e}"))?;
    request.validate()?;
    moai_cli::card_times::run(&request, &GhTransport::from_env()?, &SystemClock)
}
