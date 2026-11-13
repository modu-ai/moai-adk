//! `bin-names-apply` (#2791): make the previewed `moai` hard link.

use std::path::PathBuf;

/// Make the `moai` hard link `bin-names-plan` previewed (#2791 AC-1). The digest must match a
/// fresh preview; only `create` and `replace_copy` write. A separate copy is moved aside to
/// `moai.replaced-<UTC>`, never deleted, and moved back if the link fails.
#[derive(clap::Args)]
pub struct Args {
    /// The HOME to read (default `$HOME`). Tests pass a temporary folder.
    #[arg(long)]
    home: Option<PathBuf>,
    /// The preview's `plan_digest`.
    #[arg(long)]
    plan_digest: String,
    /// Caller-supplied permission reference (recorded; not an authorization verifier).
    #[arg(long)]
    permission_reference: String,
    /// `YYYY-MM-DDTHH:MM:SSZ` for the aside name; the system clock when absent.
    #[arg(long)]
    now: Option<String>,
}

pub const OPERATION: &str = "bin_names.apply";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let home = match args.home {
        Some(h) => h,
        None => std::env::var_os("HOME")
            .filter(|h| !h.is_empty())
            .map(PathBuf::from)
            .ok_or("HOME is not set; pass --home")?,
    };
    let exe = std::env::current_exe().map_err(|_| "current executable unavailable")?;
    let now = args.now.unwrap_or_else(moai_cli::audit::utc_now);
    moai_cli::bin_names::apply_for(
        &home,
        &exe,
        &args.plan_digest,
        &args.permission_reference,
        &now,
    )
}
