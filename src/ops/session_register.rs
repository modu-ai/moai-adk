//! `session-register` (#2592 S-C): the lane name registry; the logic is
//! `moai_cli::lane_registry`.

use std::path::PathBuf;

/// Register this session's lane name in the clone's registry (`<clone common dir>/
/// .moai-cli/sessions/<name>.json`) before claiming (#2592 S-C). A name another instance holds
/// under a live lease is refused (`conflict`); the first call returns a fresh `instance`, and
/// later calls with `--instance` renew the lease. `--release` frees the name before the
/// session is cleared.
#[derive(clap::Args)]
pub struct Args {
    /// Any checkout of the repository (the primary one or a worktree).
    project: PathBuf,
    /// The lane's name, as in its claim comments (`agent-8`).
    #[arg(long)]
    name: String,
    /// The `instance` a previous registration of this session returned; absent on the first.
    #[arg(long)]
    instance: Option<String>,
    /// Lease length in seconds, 60 to 86400 (default 3600).
    #[arg(long, default_value_t = moai_cli::lane_registry::DEFAULT_LEASE_SECS)]
    lease_secs: u64,
    /// Free the name held by `--instance` instead of registering it.
    #[arg(long, requires = "instance")]
    release: bool,
    /// `YYYY-MM-DDTHH:MM:SSZ` to judge leases against; the system clock when absent.
    #[arg(long)]
    now: Option<String>,
}

pub const OPERATION: &str = "session.register";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let now_text = args.now.unwrap_or_else(moai_cli::audit::utc_now);
    let now = moai_cli::forge::policy::parse_utc(&now_text)
        .ok_or_else(|| format!("invalid --now `{now_text}`: expected YYYY-MM-DDTHH:MM:SSZ"))?;
    let instance = args.instance.as_deref();
    match (args.release, instance) {
        (true, Some(i)) => moai_cli::lane_registry::release(&args.project, &args.name, i, now),
        _ => moai_cli::lane_registry::register(
            &args.project,
            &args.name,
            instance,
            args.lease_secs,
            now,
        ),
    }
}
