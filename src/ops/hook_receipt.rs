//! `hook-receipt`: a hook handler that records one allow-listed receipt line and never fails
//! the hook. Moved out of `src/main.rs` unchanged (#2834 R2c).

use std::io::Read;

/// Record one hook receipt line
///
/// Hook handler: record one allow-listed receipt line from the host's hook stdin under
/// ./.moai-cli/receipts/; never fails the hook.
#[derive(clap::Args)]
pub struct Args {
    #[arg(long)]
    host: String,
}

pub const OPERATION: &str = "hook.receipt";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    // Hook stdin can be large (tool output); read a bounded prefix. A cut-off object
    // fails to parse and is skipped, never an error: the hook must not fail.
    let mut bytes = Vec::new();
    let _ = std::io::stdin()
        .take(4 * 1024 * 1024)
        .read_to_end(&mut bytes);
    let now = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map_or(0, |d| d.as_millis());
    Ok(match std::env::current_dir() {
        Ok(cwd) => moai_cli::hooks::receipt(&args.host, &bytes, &cwd, now),
        Err(_) => {
            serde_json::json!({"recorded": false, "reason": "working directory unavailable"})
        }
    })
}
