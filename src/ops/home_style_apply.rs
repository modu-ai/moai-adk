//! `home-style-apply` (#2875): apply a previewed MoAI-Easy output style.

use std::path::PathBuf;

/// Write a previewed output style into HOME
///
/// Apply a previewed MoAI-Easy output style (#2875): the style file, the `outputStyle` key and
/// the managed block change only as previewed, with a receipt holding the before bytes. The
/// digest must match a fresh preview.
#[derive(clap::Args)]
pub struct Args {
    /// User home root.
    home: PathBuf,
    /// Skill package folder carrying `references/home/output-style.md`.
    package: PathBuf,
    /// The preview's `plan_digest`.
    #[arg(long)]
    plan_digest: String,
    /// Caller-supplied permission reference (recorded; not an authorization verifier).
    #[arg(long)]
    permission_reference: String,
}

pub const OPERATION: &str = "home.style_apply";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::home_style::apply(
        &args.home,
        &args.package,
        &args.plan_digest,
        &args.permission_reference,
    )
}
