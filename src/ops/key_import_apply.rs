//! `key-import-apply` (#2393 S3): copies a legacy key file's key into the vault after the
//! preview's digest. Moved out of `src/main.rs` unchanged (#2834 R3).

/// Import a previewed old key file
///
/// Copy a legacy key file's key into the encrypted vault (#2393 S3): only with the
/// preview's digest and a permission reference; the read-back must match; the legacy file
/// is kept. Destination: the current project's folder, or `_shared` with `--shared`.
#[derive(clap::Args)]
pub struct Args {
    provider: String,
    #[arg(long)]
    plan_digest: String,
    #[arg(long)]
    permission_reference: String,
    /// Import into `~/.moai-cli/vault/_shared/` (every project) instead of the project.
    #[arg(long)]
    shared: bool,
}

pub const OPERATION: &str = "key.import_apply";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let scope = moai_cli::key_vault::scope_here(args.shared)?;
    moai_cli::key_vault::import_apply(
        &moai_cli::key_vault::home()?,
        &scope,
        &args.provider,
        &args.plan_digest,
        &args.permission_reference,
    )
}
