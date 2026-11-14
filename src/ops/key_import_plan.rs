//! `key-import-plan` (#2393 S3): preview of moving a legacy key file into the vault. Moved out
//! of `src/main.rs` unchanged (#2834 R3).

/// Preview importing an old key file
///
/// Preview of moving a legacy key file into the encrypted vault (#2393 S3): which file,
/// which destination folder, never the value. Writes nothing.
#[derive(clap::Args)]
pub struct Args {
    provider: String,
    /// Import into `~/.moai-cli/vault/_shared/` (every project) instead of the project.
    #[arg(long)]
    shared: bool,
}

pub const OPERATION: &str = "key.import_plan";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let scope = moai_cli::key_vault::scope_here(args.shared)?;
    moai_cli::key_vault::import_plan(&moai_cli::key_vault::home()?, &scope, &args.provider)
}
