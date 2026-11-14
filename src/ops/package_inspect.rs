//! `package-inspect` (01-A): the digest of a package tree; reads files, writes nothing.

use std::path::PathBuf;

/// Inspect a skill package's digests
#[derive(clap::Args)]
pub struct Args {
    package: PathBuf,
}

pub const OPERATION: &str = "package.inspect";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    Ok(serde_json::json!({
        "digest": moai_cli::package_digest(&args.package)?,
        "digest_algorithm": "moai-package-tree-v1",
        "publisher_trust": "not_verified",
        "host_loaded": false
    }))
}
