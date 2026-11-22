//! `host-wake`: one flow wake through the host adapter (SPEC-20261122-session-messages).
//! Lists the config root's session registry — a missing or unreadable registry
//! folder is the no-channel refusal, never an internal error — refuses with
//! the reason when the addressee is unreachable, and answers the result
//! shape: queued (one line written to the addressee's inbox socket), held
//! with a notice, or refused (an unknown name, no channel, no inbox, an
//! ambiguous name, the idle flag, or a socket path outside its guarded
//! folder).

use std::path::PathBuf;

/// One wake through the host adapter
#[derive(clap::Args)]
pub struct Args {
    /// The addressee's session name
    name: String,
    /// The one-line text naming the ledger change to read
    #[arg(long)]
    text: String,
    /// Attach the one-shot idle notice instead of a polling loop
    #[arg(long)]
    idle: bool,
    /// The config root whose session registry the wake lists
    #[arg(long)]
    config_dir: Option<PathBuf>,
}

pub const OPERATION: &str = "host.wake";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let config_dir = args.config_dir.unwrap_or_else(default_config_dir);
    let wake = moai_cli::host_wake::Wake {
        addressee: args.name,
        text: args.text,
        notify_when_idle: args.idle,
    };
    // B1 of the N-slice audit: a missing or unreadable registry folder is the
    // no-channel refusal, never an internal error; the refused wrapper below
    // answers it in the same shape as every other refusal.
    let answer = match moai_cli::host_wake::RegistryListing::read(&config_dir) {
        Ok(listing) => wake.send_once(&listing),
        Err(_) => Ok(moai_cli::host_wake::SendShape::Refused(
            moai_cli::host_wake::Refusal::NoChannel,
        )),
    };
    // N3 and N9 of the first audit: the answer says queued, never delivered
    // (the status notice goes to a sender socket this command does not have),
    // and every failure answers the refused shape with its one-line reason.
    let answer = match answer {
        Ok(shape) => shape,
        Err(reason) => {
            return Ok(serde_json::json!({
                "schema_version": 1,
                "addressee": wake.addressee,
                "shape": "refused",
                "reason": reason,
                "socket_write": false,
            }));
        }
    };
    Ok(serde_json::json!({
        "schema_version": 1,
        "addressee": wake.addressee,
        "shape": shape_name(&answer),
        "reason": reason_line(&answer),
        // N4: a queued wake wrote one line to the addressee's inbox socket;
        // nothing else was written.
        "socket_write": matches!(answer, moai_cli::host_wake::SendShape::Queued),
    }))
}

/// The default config root: `CLAUDE_CONFIG_DIR` when set, else `~/.claude`
/// (the runtime's own rule; the account separation is the config dir).
fn default_config_dir() -> PathBuf {
    if let Ok(dir) = std::env::var("CLAUDE_CONFIG_DIR") {
        return PathBuf::from(dir);
    }
    let home = std::env::var("HOME").unwrap_or_default();
    PathBuf::from(home).join(".claude")
}

fn shape_name(shape: &moai_cli::host_wake::SendShape) -> &'static str {
    match shape {
        moai_cli::host_wake::SendShape::Queued => "queued",
        moai_cli::host_wake::SendShape::HeldWithNotice => "held_with_notice",
        moai_cli::host_wake::SendShape::Refused(_) => "refused",
    }
}

fn reason_line(shape: &moai_cli::host_wake::SendShape) -> Option<String> {
    match shape {
        moai_cli::host_wake::SendShape::Refused(refusal) => Some(refusal.line().to_string()),
        _ => None,
    }
}
