//! `host-wake`: one flow wake through the host adapter (SPEC-20261122-session-messages).
//! Lists the config root's session registry, refuses with the reason when the
//! addressee is unreachable, and answers the result shape; the send itself
//! rides the D-8 seam (the documented wake path today is a Claude session's
//! SendMessage tool, so the CLI sender answers the open-item line).

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
    let listing = moai_cli::host_wake::RegistryListing::read(&config_dir)?;
    let answer = wake.send_once(&listing)?;
    Ok(serde_json::json!({
        "schema_version": 1,
        "addressee": wake.addressee,
        "shape": shape_name(&answer),
        "reason": reason_line(&answer),
        "writes": false,
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
        moai_cli::host_wake::SendShape::Delivered => "delivered",
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
