//! `host-wake`: one flow wake through the host adapter (SPEC-20261122-session-messages).
//! Lists the config root's session registry — a missing or unreadable sessions
//! folder answers with its path named, never an internal error — refuses with
//! the reason when the addressee is unreachable, and answers the result
//! shape: queued (one line written to the addressee's inbox socket), held
//! with a notice, or refused (an unknown name, the sessions folder missing,
//! no inbox, an ambiguous name, the idle flag, or a socket path outside its
//! guarded folder).

use serde_json::Value;
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
    // Follow-up 2 of the landed N slice: validation and the idle check run
    // before any registry read, so an invalid name, an empty text or the
    // idle flag answers its own reason even when the registry folder is
    // missing. Follow-up 1: a missing or unreadable registry folder — a
    // mistyped --config-dir, a file in its place — answers with the folder
    // named, never an internal error.
    let answer = match wake.validate() {
        Err(invalid) => Err(invalid.line().to_string()),
        Ok(()) if wake.notify_when_idle => Err(moai_cli::host_wake::Refusal::IdleUnsupported
            .line()
            .to_string()),
        Ok(()) => match moai_cli::host_wake::RegistryListing::read(&config_dir) {
            Err(_) => Err(format!(
                "wake refused: the sessions registry {} is missing or unreadable",
                config_dir.join("sessions").display()
            )),
            Ok(listing) => wake.send_once(&listing),
        },
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
    let reason = match reason_line(&answer) {
        Some(line) => Value::String(line),
        None => Value::Null,
    };
    Ok(serde_json::json!({
        "schema_version": 1,
        "addressee": wake.addressee,
        "shape": shape_name(&answer),
        "reason": reason,
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
