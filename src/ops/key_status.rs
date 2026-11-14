//! `key-status` (#2393): whether a provider's vault entry exists, is safe and decrypts; never
//! the value. Moved out of `src/main.rs` unchanged (#2834 R3).

/// Check a provider key without showing it
///
/// Whether a provider's entry (current project, or `--shared`) exists, is safe and
/// decrypts (#2393). Never the value, its length or a fingerprint.
#[derive(clap::Args)]
pub struct Args {
    provider: String,
    /// The key every project uses (`~/.moai-cli/vault/_shared/`).
    #[arg(long)]
    shared: bool,
}

pub const OPERATION: &str = "key.status";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let scope = moai_cli::key_vault::scope_here(args.shared)?;
    moai_cli::key_vault::status(&moai_cli::key_vault::home()?, &scope, &args.provider)
}
