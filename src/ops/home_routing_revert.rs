//! `home-routing-revert` (#2638): undo one routing section apply.

use std::path::PathBuf;

/// Undo one routing section apply
///
/// Undo one routing section apply (#2638) for the files that still hold the bytes it wrote; a
/// file changed since is skipped and reported.
#[derive(clap::Args)]
pub struct Args {
    /// User home root.
    home: PathBuf,
    /// The receipt id a `home-routing-apply` returned (`routing-…`).
    #[arg(long)]
    receipt: String,
    /// Caller-supplied permission reference (recorded; not an authorization verifier).
    #[arg(long)]
    permission_reference: String,
}

pub const OPERATION: &str = "home.routing_revert";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::home_routing::revert(&args.home, &args.receipt, &args.permission_reference)
}
