//! `session-register` (#2592 S-C): the lane name registry; the logic is
//! `moai_cli::lane_registry`.

use std::path::PathBuf;

/// Register this session's lane name
///
/// Register this session's lane name in the clone's registry (`<clone common dir>/
/// .moai-cli/sessions/<name>.json`) before claiming (#2592 S-C). A name another instance holds
/// under a live lease is refused (`conflict`); the first call returns a fresh `instance`, and
/// later calls with `--instance` renew the lease. `--release` frees the name before the
/// session is cleared. `--show` reads the holder without writing and may leave PROJECT out (the
/// clone of the current folder is read; #3015 C3); the leader then releases
/// that instance with `--release --by leader` (#2902 O3a). `--mode leader|agent` records the
/// session's mode for the statusline tag (#2928).
#[derive(clap::Args)]
pub struct Args {
    /// Any checkout of the repository (the primary one or a worktree). Only `--show` may leave
    /// it out: it then reads the clone of the current folder.
    project: Option<PathBuf>,
    /// The lane's name, as in its claim comments (`agent-8`). Required unless `--liveness-of`
    /// is given (the lookup answers a session id and needs no name).
    #[arg(long)]
    name: Option<String>,
    /// A holder session id to answer liveness for (read-only, BOARD 03 (가),
    /// SPEC-20261120-launcher-liveness REQ-2): `alive` (the name's `.live` lock is held),
    /// `launcher-exited` (the lock is free) or `unknown` (no entry names the id, or the lock
    /// cannot be probed). A leader entry with no `.live` lock file is `alive` while its lease
    /// runs (evidence basis `registry_lease`). Writes nothing. Reads the clone of the current
    /// folder unless PROJECT is given, exactly as `--show` does.
    #[arg(long, conflicts_with_all = ["instance", "release", "show", "mode", "name"])]
    liveness_of: Option<String>,
    /// The `instance` a previous registration of this session returned; absent on the first.
    #[arg(long)]
    instance: Option<String>,
    /// Lease length in seconds, 60 to 86400 (default 14400, four hours).
    #[arg(long, default_value_t = moai_cli::lane_registry::DEFAULT_LEASE_SECS)]
    lease_secs: u64,
    /// Free the name held by `--instance` instead of registering it.
    #[arg(long, requires = "instance")]
    release: bool,
    /// Who releases for the lane: only `leader`, with the instance `--show` printed.
    #[arg(long, requires = "release", value_parser = ["leader"])]
    by: Option<String>,
    /// Print the current holder's `instance`, lease end and host `session` (when it has one), or
    /// `free`; writes nothing.
    #[arg(long, conflicts_with_all = ["instance", "release"])]
    show: bool,
    /// The session's mode, `leader` or `agent`, kept in the entry for the statusline (#2928);
    /// stored only when given, so pass it again with every renewal. `leader` is the coordination
    /// window: it needs the host session id (refused without), is refused while another leader
    /// entry holds a live lease, and is renewed only from the session it names.
    #[arg(long, conflicts_with_all = ["show", "release"], value_parser = moai_cli::lane_registry::MODES.to_vec())]
    mode: Option<String>,
    /// After a `/clear` or a restart: take over the entry of this name when it names another
    /// host session in the same git worktree. The entry then lists the replaced session id first
    /// in `previous_sessions` and gets a fresh `instance` (`continued`); `card adopt` reads that
    /// list. Refused, with the missing proof named and the entry left as it was, when the entry
    /// names no session or no worktree, names another worktree, or this session has no host id or
    /// runs outside a git worktree. With no entry it is an ordinary registration.
    #[arg(long = "continue", conflicts_with_all = ["instance", "release", "show", "liveness_of"])]
    continue_session: bool,
    /// `YYYY-MM-DDTHH:MM:SSZ` to judge leases against; the system clock when absent.
    #[arg(long)]
    now: Option<String>,
}

pub const OPERATION: &str = "session.register";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let now_text = args.now.unwrap_or_else(moai_cli::audit::utc_now);
    let now = moai_cli::forge::policy::parse_utc(&now_text)
        .ok_or_else(|| format!("invalid --now `{now_text}`: expected YYYY-MM-DDTHH:MM:SSZ"))?;
    // BOARD 03 (가), SPEC-20261120-launcher-liveness REQ-2: a read-only liveness answer for a
    // holder session id; nothing is written. It needs no --name and reads the clone of the
    // current folder unless PROJECT is given (B1 of the audit: the workflow document spells
    // the command without either).
    if let Some(id) = &args.liveness_of {
        let project = args.project.clone().unwrap_or_else(|| PathBuf::from("."));
        let (mut state, mut evidence) = moai_cli::lane_registry::liveness_of_session(&project, id);
        use moai_cli::lane_registry::Liveness;
        // SPEC-20261123-leader-registration-lease REQ-6: a coordination window opened without
        // the launcher has no `.live` lock, so a leader entry whose lease runs is the only sign
        // it is there. A lock file that exists keeps its answer.
        if state == Liveness::LauncherExited {
            if let Some(leader) =
                moai_cli::leader_round::registration::leader_alive_by_lease(&project, id, now)
            {
                state = Liveness::Alive;
                evidence = serde_json::json!({
                    "name": leader.name,
                    "lease_until": moai_cli::forge::policy::format_utc(leader.lease_until),
                    "basis": "registry_lease",
                });
            }
        }
        return Ok(serde_json::json!({
            "action": "session.liveness",
            "session": id,
            "state": match state {
                Liveness::Alive => "alive",
                Liveness::LauncherExited => "launcher-exited",
                Liveness::Unknown => "unknown",
            },
            "evidence": evidence,
            "writes": false,
        }));
    }
    // The other paths write or read a named entry: the name is required there (B1 of the
    // SPEC-20261120-launcher-liveness audit: --liveness-of answers a session id and needs no
    // name, and reads the current folder's clone unless PROJECT is given).
    let name = args.name.clone().ok_or(
        "NAME is required: pass --name <lane> (only --liveness-of answers a session id without a name)",
    )?;
    if args.show {
        let project = args.project.unwrap_or_else(|| PathBuf::from("."));
        return moai_cli::lane_registry::show(&project, &name, now);
    }
    let project = args
        .project
        .ok_or("PROJECT is required (only --show and --liveness-of may leave it out)")?;
    let instance = args.instance.as_deref();
    match (args.release, instance) {
        (true, Some(i)) => {
            moai_cli::lane_registry::release(&project, &name, i, args.by.as_deref(), now)
        }
        _ => moai_cli::lane_registry::register_with(
            &project,
            &name,
            instance,
            args.lease_secs,
            args.mode.as_deref(),
            now,
            args.continue_session,
        ),
    }
}
