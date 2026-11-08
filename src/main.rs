use clap::{error::ErrorKind, ArgGroup, Args, Parser, Subcommand};
use moai_cli::{
    agents_apply, agents_plan, apply, contract, hosts::codex, plan, resolve, session_launch,
    AgentsPlan, AgentsRequest, Plan, Profile, SessionRequest,
};
use std::{
    io::{self, Read},
    path::PathBuf,
    process::ExitCode,
};

#[derive(Parser)]
#[command(
    name = "moai-cli",
    version,
    about = "MoAI RC internal execution tool",
    args_conflicts_with_subcommands = true
)]
struct Cli {
    /// Start a leader session on the default host (cc).
    #[arg(short = 'l', long = "leader", conflicts_with = "agent")]
    leader: bool,
    /// Start an agent session on HOST (cc|glm|codex; default cc).
    #[arg(short = 'a', long = "agent", value_name = "HOST", num_args = 0..=1, default_missing_value = "cc")]
    agent: Option<String>,
    #[command(flatten)]
    launch: LaunchOpts,
    #[command(subcommand)]
    command: Option<Command>,
}

/// Options shared by `moai-cli -l|-a` and `moai-cli cc|glm|codex -l|-a` (00-F #2028).
#[derive(Args)]
struct LaunchOpts {
    /// Session name (default `leader` for -l, `agent-1` for -a).
    #[arg(long)]
    name: Option<String>,
    /// Main model (Claude hosts default `opus`; codex uses its own default unless set).
    #[arg(long)]
    model: Option<String>,
    /// Claude effort (default `medium`).
    #[arg(long)]
    effort: Option<String>,
    /// Do not pass `--advisor fable` (Claude Code host).
    #[arg(long)]
    no_advisor: bool,
    /// Keep permission prompts: do not pass `--dangerously-skip-permissions`.
    #[arg(long)]
    permissions: bool,
    /// Print the composed command as JSON instead of starting the host.
    #[arg(long)]
    dry_run: bool,
}

/// `moai-cli cc|glm|codex [-l|-a] [options] [PROMPT]`. Without `-l|-a` it is a plain host
/// session: no role is guessed (Issue #2053).
#[derive(Args)]
#[command(group(ArgGroup::new("role").args(["leader", "agent"])))]
struct HostLaunch {
    /// Leader session.
    #[arg(short = 'l', long = "leader")]
    leader: bool,
    /// Agent session.
    #[arg(short = 'a', long = "agent")]
    agent: bool,
    #[command(flatten)]
    launch: LaunchOpts,
    /// Extra first-prompt text (codex only).
    prompt: Option<String>,
}
#[derive(Subcommand)]
enum Command {
    PackageInspect {
        package: PathBuf,
    },
    ProjectPlan {
        project: PathBuf,
    },
    ProjectApply {
        project: PathBuf,
        #[arg(long)]
        plan_digest: String,
        #[arg(long)]
        permission_reference: String,
    },
    PackageResolve {
        project: PathBuf,
        #[arg(long)]
        repo_package: Option<PathBuf>,
        #[arg(long)]
        user_package: Option<PathBuf>,
    },
    AgentsPlan {
        project: PathBuf,
    },
    AgentsApply {
        project: PathBuf,
        #[arg(long)]
        plan_digest: String,
        #[arg(long)]
        permission_reference: String,
    },
    /// Revert exactly the files one apply request published, from its receipt
    /// (Issue #1989, 01-B). Files changed since the apply are skipped, never overwritten.
    ApplyRecover {
        project: PathBuf,
        /// Receipt id printed by the apply result (`receipt` field) or the file stem
        /// under `<PROJECT>/.moai-cli/receipts/`.
        #[arg(long)]
        receipt: String,
        #[arg(long)]
        permission_reference: String,
    },
    /// Read-only baseline inventory of a checkout (typed JSON on stdout).
    BaselineInventory {
        project: PathBuf,
        /// Go-generation baseline commit whose `.claude` trees are compared (read with ls-tree only).
        #[arg(long)]
        go_baseline_commit: Option<String>,
        /// Extra repository to look up the baseline commit in when the project lacks it.
        #[arg(long)]
        go_repo: Option<PathBuf>,
    },
    /// Read GitHub Issues and comments as TODO (GET only; typed JSON request on stdin,
    /// envelope on stdout). Poll state: `<PROJECT>/.moai-cli/forge/poll-state.json`.
    ForgeRead {
        project: PathBuf,
    },
    /// Create or update the one report comment of a run on its Issue (typed JSON payload on
    /// stdin; envelope on stdout). Large output goes to artifact references, not the body.
    ForgeReport {
        project: PathBuf,
        #[arg(long)]
        permission_reference: String,
    },
    /// Preview promotion of selected local TODOs to GitHub Issues (reads only; typed JSON
    /// request on stdin naming `tasks_path`, `target` and `selected`; envelope on stdout).
    TodoPromotePreview {
        project: PathBuf,
    },
    /// Promote the previewed TODOs: refuses a different `--preview-digest`, records an outbox
    /// intent before each create and `published` only after a confirming GET.
    TodoPromoteApply {
        project: PathBuf,
        #[arg(long)]
        preview_digest: String,
        #[arg(long)]
        permission_reference: String,
    },
    /// Read-only clone/workspace identity and state roots of a path (typed JSON on stdout).
    Identity {
        path: PathBuf,
    },
    /// Claim a task (06-A #2005): typed JSON request on stdin; posts exactly one claim
    /// comment (⑤) behind the same-host journal serialization (①) and records the record.
    ClaimAcquire {
        project: PathBuf,
    },
    /// Read the current claim record of a task key; byte-identical when repeated.
    ClaimShow {
        project: PathBuf,
        /// The 16-hex task key an acquire returned.
        #[arg(long)]
        task: String,
    },
    /// Close the active claim as `cancelled`; the epoch must match the current record.
    ClaimRelease {
        project: PathBuf,
        #[arg(long)]
        task: String,
        #[arg(long)]
        epoch: u64,
        /// Recorded in the release effect.
        #[arg(long)]
        reason: Option<String>,
    },
    /// Read-only Claude Code check: whether a session started in PROJECT reads AGENTS.md.
    ClaudeDoctor {
        project: PathBuf,
    },
    /// Plan a Codex projection of the common package into an explicit target; writes nothing.
    CodexProjectionPlan {
        package: PathBuf,
        /// Absolute skill directory named `moai`, e.g. ~/.agents/skills/moai (no default).
        #[arg(long)]
        target: PathBuf,
        #[arg(long, default_value = "MoAI")]
        display_name: String,
        #[arg(
            long,
            default_value = "Set up, plan, implement, verify, audit and sync through one workflow"
        )]
        short_description: String,
        #[arg(long, default_value = "Use $moai to read this project's status")]
        default_prompt: String,
        /// Write `policy.allow_implicit_invocation`; omitted keeps Codex's default.
        #[arg(long)]
        allow_implicit_invocation: Option<bool>,
        /// Requested `tui.status_line` item (repeatable); reported as HOLD, never written.
        #[arg(long = "status-line")]
        status_line: Vec<String>,
        /// Requested `[skills.bundled] enabled = false`; reported as HOLD, never written.
        #[arg(long)]
        disable_bundled_skills: bool,
    },
    /// Read-only Codex doctor: AGENTS chain and budget, skill duplicates, `.codex/` layer.
    CodexDoctor {
        /// Project root (Codex walks from here down to the working directory).
        project: PathBuf,
        /// Working directory inside the project (default: the project root).
        #[arg(long)]
        cwd: Option<PathBuf>,
        /// Codex home holding the global AGENTS files (default: $CODEX_HOME, else ~/.codex).
        #[arg(long)]
        codex_home: Option<PathBuf>,
        /// Home holding `.agents/skills` (default: $HOME).
        #[arg(long)]
        home: Option<PathBuf>,
        /// `project_doc_fallback_filenames` entry (repeatable; config.toml is not read).
        #[arg(long)]
        fallback: Vec<String>,
        /// `project_doc_max_bytes` (default 32768; config.toml is not read).
        #[arg(long)]
        max_bytes: Option<u64>,
    },
    /// Read-only installation inspection (04-A #1998): PATH generations, receipts under
    /// `~/.moai-cli/installations/`, Go-generation path existence. Writes nothing.
    InstallInspect {
        /// Accepted for the REQ-04A-02 invocation form; output is always the JSON envelope.
        #[arg(long)]
        json: bool,
    },
    /// Temporary (05-B #2002; the name is fixed after P06): RUN admission of an approved plan.
    /// Typed request on stdin; reads the project's `HEAD^{commit}` with git, writes nothing.
    PlanAdmit {
        project: PathBuf,
    },
    /// Temporary (05-C #2003; the name is fixed after P06): re-check the pinned contract and
    /// run the approved verification commands in PROJECT as argv; returns a receipt.
    RunVerify {
        project: PathBuf,
    },
    /// Read-only: PROJECT's HEAD commit and a digest of its tracked changes (07-D #2011).
    SnapshotDescribe {
        project: PathBuf,
    },
    /// Read-only: requested file ranges with digests and a budget (typed request on stdin).
    ContextBundle {
        project: PathBuf,
    },
    /// Read-only: re-check a bundle's ranges against the files now (typed request on stdin).
    ContextVerify {
        project: PathBuf,
    },
    /// Append one native decision record (stdin) to `.moai-cli/decisions/<RUN>.jsonl`.
    DecisionRecord {
        project: PathBuf,
        #[arg(long)]
        run: String,
        #[arg(long)]
        permission_reference: String,
    },
    /// Temporary (05-D #2004; the name is fixed after P06): independent audit of PROJECT. Re-runs
    /// the approved verification itself and decides admission; receipts on stdin never count.
    AuditReduce {
        project: PathBuf,
    },
    /// Claude Code session launcher: `cc -l|-a`.
    Cc(HostLaunch),
    /// Claude Code on z.ai GLM: `glm -l|-a` (needs ANTHROPIC_AUTH_TOKEN in the environment).
    Glm(HostLaunch),
    /// Codex CLI session launcher: `codex -l|-a`.
    Codex(HostLaunch),
    /// Legacy Claude-only launcher kept for compatibility; prefer `-l|-a` or `cc|glm|codex`.
    Session {
        /// leader | agent
        #[arg(long, default_value = "agent")]
        role: String,
        #[arg(long, default_value = "claude")]
        host: String,
        /// Session display name, e.g. leader or agent-3 (passed as `--name`).
        #[arg(long)]
        name: String,
        #[arg(long, default_value = "opus")]
        model: String,
        #[arg(long, default_value = "medium")]
        effort: String,
        /// Advisor model for hard reasoning (fable|opus|sonnet or full id); omit to disable.
        #[arg(long)]
        advisor: Option<String>,
        #[arg(long, default_value = "claude")]
        claude_bin: String,
        /// Print the composed command as JSON instead of starting the host.
        #[arg(long)]
        dry_run: bool,
    },
}
fn input<T: serde::de::DeserializeOwned>() -> Result<T, String> {
    let mut bytes = Vec::new();
    io::stdin()
        .take(65537)
        .read_to_end(&mut bytes)
        .map_err(|_| "stdin read failed")?;
    if bytes.len() > 65536 {
        return Err("input exceeds 64 KiB".into());
    }
    serde_json::from_slice(&bytes).map_err(|_| "invalid typed JSON input".into())
}
/// Maps the launcher's host word to the session host.
fn host_of(word: &str) -> Result<&'static str, String> {
    match word {
        "cc" => Ok("claude"),
        "glm" => Ok("glm"),
        "codex" => Ok("codex"),
        other => Err(format!("unknown host {other} (cc|glm|codex)")),
    }
}

/// Builds the session request for the launcher syntax (REQ-00F-01..04). `leader` is
/// `Some(true)` for `-l`, `Some(false)` for `-a` and `None` without a role (#2053).
fn launch(
    host: &str,
    leader: Option<bool>,
    o: LaunchOpts,
    prompt: Option<String>,
) -> Result<serde_json::Value, String> {
    let role = leader.map(|l| if l { "leader" } else { "agent" });
    let name_given = o.name.is_some();
    // A role brings its default name; a plain session gets none.
    let name = o
        .name
        .or_else(|| leader.map(|l| if l { "leader" } else { "agent-1" }.into()));
    let request = SessionRequest {
        role: role.map(str::to_string),
        host: host.into(),
        name,
        model: o.model,
        effort: o.effort,
        // Claude defaults (operator 2026-11-08): advisor fable; without a role only under
        // HOLD-L1 option A. The permission skip (interim) applies to `-l|-a` only (#2053).
        advisor: (host != "codex" && !o.no_advisor && moai_cli::session::launcher_defaults(role))
            .then(|| "fable".into()),
        skip_permissions: role.is_some() && host != "codex" && !o.permissions,
        prompt,
        claude_bin: "claude".into(),
        codex_bin: "codex".into(),
    };
    let mut v = session_launch(&request, o.dry_run)?;
    if leader == Some(false) && !name_given {
        // HOLD (00-F): automatic agent numbering is undecided; RC1 uses a fixed default.
        if let Some(notes) = v["notes"].as_array_mut() {
            notes.push("agent name defaulted to agent-1 (automatic numbering is HOLD); pass --name agent-<n>".into());
        }
    }
    Ok(v)
}

/// Internal operations (01-C #1990): subcommand word and operation name. The launchers are
/// developer tools and stay outside the envelope contract. A slice, not a fixed-length array:
/// two branches each bumping the length to the same number merged silently (#2044).
const OPERATIONS: &[(&str, &str)] = &[
    ("package-inspect", "package.inspect"),
    ("project-plan", "project.plan"),
    ("project-apply", "project.apply"),
    ("package-resolve", "package.resolve"),
    ("agents-plan", "agents.plan"),
    ("agents-apply", "agents.apply"),
    ("apply-recover", "apply.recover"),
    ("baseline-inventory", "baseline.inventory"),
    ("identity", "identity"),
    ("forge-read", "forge.read"),
    ("forge-report", "forge.report"),
    ("plan-admit", "plan.admit"),
    ("claude-doctor", "claude.doctor"),
    ("codex-projection-plan", "codex.projection_plan"),
    ("codex-doctor", "codex.doctor"),
    ("claim-acquire", "claim.acquire"),
    ("claim-show", "claim.show"),
    ("claim-release", "claim.release"),
    ("todo-promote-preview", "todo.promote.plan"),
    ("todo-promote-apply", "todo.promote.apply"),
    ("run-verify", "run.verify"),
    ("snapshot-describe", "snapshot.describe"),
    ("context-bundle", "context.bundle"),
    ("context-verify", "context.verify"),
    ("decision-record", "decision.record"),
    ("audit-reduce", "audit.reduce"),
    ("install-inspect", "install.inspect"),
];
fn operation_of(word: &str) -> Option<&'static str> {
    OPERATIONS
        .iter()
        .find(|(w, _)| *w == word)
        .map(|&(_, op)| op)
}

/// `-l` → leader, `-a` → agent, neither → no role (never guessed).
fn role_of(h: &HostLaunch) -> Option<bool> {
    match (h.leader, h.agent) {
        (true, _) => Some(true),
        (false, true) => Some(false),
        (false, false) => None,
    }
}

/// Launcher paths (`-l`, `-a`, `cc|glm|codex`, `session`): plain JSON or a plain error line.
fn run_launcher(cli: Cli) -> Result<serde_json::Value, String> {
    let Some(command) = cli.command else {
        return match (cli.leader, cli.agent) {
            (true, _) => launch("claude", Some(true), cli.launch, None),
            (false, Some(word)) => launch(host_of(&word)?, Some(false), cli.launch, None),
            (false, None) => Err(
                "usage: moai-cli -l | -a [cc|glm|codex] | moai-cli cc|glm|codex -l|-a | moai-cli <command>; see --help".into(),
            ),
        };
    };
    match command {
        Command::Cc(h) => launch("claude", role_of(&h), h.launch, h.prompt),
        Command::Glm(h) => launch("glm", role_of(&h), h.launch, h.prompt),
        Command::Codex(h) => launch("codex", role_of(&h), h.launch, h.prompt),
        Command::Session {
            role,
            host,
            name,
            model,
            effort,
            advisor,
            claude_bin,
            dry_run,
        } => {
            // Legacy behaviour (REQ-00F-06): Claude only, no permission skip.
            if host != "claude" {
                return Err(format!(
                    "host {host} is not available on `session`; use `moai-cli cc|glm|codex -l|-a`"
                ));
            }
            session_launch(
                &SessionRequest {
                    role: Some(role),
                    host,
                    name: Some(name),
                    model: Some(model),
                    effort: Some(effort),
                    advisor,
                    skip_permissions: false,
                    prompt: None,
                    claude_bin,
                    codex_bin: "codex".into(),
                },
                dry_run,
            )
        }
        _ => unreachable!("internal operations run through run_internal"),
    }
}

/// Internal operations; `main` wraps the result in the envelope.
fn run_internal(command: Command) -> Result<serde_json::Value, String> {
    match command {
        Command::PackageInspect { package } => Ok(
            serde_json::json!({"digest":moai_cli::package_digest(&package)?,"digest_algorithm":"moai-package-tree-v1","publisher_trust":"not_verified","host_loaded":false}),
        ),
        Command::ProjectPlan { project } => {
            // REQ-03A-08: an unknown input key or a non-v2 schema_version is refused with
            // the profile message; no input value is echoed.
            let profile: Profile = serde_json::from_value(input::<serde_json::Value>()?)
                .map_err(|_| "invalid or nonportable profile; no values were logged".to_string())?;
            serde_json::to_value(plan(&project, profile)?).map_err(|_| "encode failed".into())
        }
        Command::ProjectApply {
            project,
            plan_digest,
            permission_reference,
        } => apply(
            &project,
            input::<Plan>()?,
            &plan_digest,
            &permission_reference,
        ),
        Command::PackageResolve {
            project,
            repo_package,
            user_package,
        } => resolve(&project, repo_package.as_deref(), user_package.as_deref()),
        Command::AgentsPlan { project } => {
            serde_json::to_value(agents_plan(&project, input::<AgentsRequest>()?)?)
                .map_err(|_| "encode failed".into())
        }
        Command::AgentsApply {
            project,
            plan_digest,
            permission_reference,
        } => agents_apply(
            &project,
            input::<AgentsPlan>()?,
            &plan_digest,
            &permission_reference,
        ),
        Command::ApplyRecover {
            project,
            receipt,
            permission_reference,
        } => moai_cli::apply::recover(&project, &receipt, &permission_reference),
        Command::ForgeReport {
            project,
            permission_reference,
        } => {
            let mut bytes = Vec::new();
            io::stdin()
                .take(65537)
                .read_to_end(&mut bytes)
                .map_err(|_| "stdin read failed")?;
            if bytes.len() > 65536 {
                return Err(
                    "input exceeds 64 KiB; put large output in an artifact reference, not the payload"
                        .into(),
                );
            }
            let payload: moai_cli::forge::report::ReportPayload =
                serde_json::from_slice(&bytes).map_err(|_| "invalid typed JSON input")?;
            moai_cli::forge::report::report_cli(
                &project,
                payload,
                &permission_reference,
                &moai_cli::forge::transport::GhTransport::from_env()?,
                &moai_cli::forge::policy::SystemClock,
            )
        }
        Command::ForgeRead { project } => moai_cli::forge::forge_read_cli(
            &project,
            input::<moai_cli::forge::ReadRequest>()?,
            &moai_cli::forge::transport::GhTransport::from_env()?,
            &moai_cli::forge::policy::SystemClock,
        ),
        Command::TodoPromotePreview { project } => moai_cli::forge::promote::preview_cli(
            &project,
            input::<moai_cli::forge::promote::PromoteRequest>()?,
            &moai_cli::forge::transport::GhTransport::from_env()?,
            &moai_cli::forge::policy::SystemClock,
        ),
        Command::TodoPromoteApply {
            project,
            preview_digest,
            permission_reference,
        } => moai_cli::forge::promote::apply_cli(
            &project,
            input::<moai_cli::forge::promote::PromoteRequest>()?,
            &preview_digest,
            &permission_reference,
            &moai_cli::forge::transport::GhTransport::from_env()?,
            &moai_cli::forge::policy::SystemClock,
        ),
        Command::BaselineInventory {
            project,
            go_baseline_commit,
            go_repo,
        } => moai_cli::baseline::inventory(
            &project,
            go_baseline_commit.as_deref(),
            go_repo.as_deref(),
        ),
        Command::Identity { path } => moai_cli::identity::identity(&path),
        Command::ClaimAcquire { project } => moai_cli::claim::acquire_cli(
            &project,
            input::<moai_cli::claim::ClaimRequest>()?,
            &moai_cli::forge::transport::GhTransport::from_env()?,
            &moai_cli::forge::policy::SystemClock,
        ),
        Command::ClaimShow { project, task } => moai_cli::claim::show_cli(&project, &task),
        Command::ClaimRelease {
            project,
            task,
            epoch,
            reason,
        } => moai_cli::claim::release_cli(
            &project,
            &task,
            epoch,
            reason.as_deref().unwrap_or("unspecified"),
            &moai_cli::forge::policy::SystemClock,
        ),
        Command::ClaudeDoctor { project } => moai_cli::claude::doctor(&project),
        Command::CodexProjectionPlan {
            package,
            target,
            display_name,
            short_description,
            default_prompt,
            allow_implicit_invocation,
            status_line,
            disable_bundled_skills,
        } => codex::projection_plan(
            &package,
            &target,
            &codex::OpenAiYaml {
                display_name,
                short_description,
                default_prompt,
                allow_implicit_invocation,
            },
            &status_line,
            disable_bundled_skills,
        ),
        Command::CodexDoctor {
            project,
            cwd,
            codex_home,
            home,
            fallback,
            max_bytes,
        } => {
            let home = home
                .or_else(|| std::env::var_os("HOME").map(PathBuf::from))
                .ok_or("home directory missing; pass --home")?;
            let codex_home = codex_home
                .or_else(|| std::env::var_os("CODEX_HOME").map(PathBuf::from))
                .unwrap_or_else(|| home.join(".codex"));
            let cwd = cwd.unwrap_or_else(|| project.clone());
            codex::doctor(&project, &cwd, &codex_home, &home, &fallback, max_bytes)
        }
        Command::InstallInspect { json: _ } => {
            let home = std::env::var_os("HOME")
                .filter(|h| !h.is_empty())
                .map(PathBuf::from)
                .ok_or("home directory missing; install-inspect requires HOME")?;
            let path_env = std::env::var("PATH").unwrap_or_default();
            moai_cli::install::inspect(&home, &path_env)
        }
        Command::PlanAdmit { project } => {
            let request = input::<moai_cli::plan::AdmissionRequest>()?;
            if request.protocol_version != contract::PROTOCOL_VERSION {
                return Err("unsupported protocol_version; supported: 1".into());
            }
            let current = moai_cli::plan::current_revision(&project)?;
            serde_json::to_value(moai_cli::plan::admit(&request, &current))
                .map_err(|_| "encode failed".into())
        }
        Command::RunVerify { project } => {
            let request = input::<moai_cli::run::RunRequest>()?;
            if request.protocol_version != contract::PROTOCOL_VERSION {
                return Err("unsupported protocol_version; supported: 1".into());
            }
            let current = moai_cli::plan::current_revision(&project)?;
            serde_json::to_value(moai_cli::run::run(&project, &request, &current)?)
                .map_err(|_| "encode failed".into())
        }
        Command::SnapshotDescribe { project } => moai_cli::context::snapshot(&project),
        Command::ContextBundle { project } => moai_cli::context::bundle(&project, &input()?),
        Command::ContextVerify { project } => moai_cli::context::verify(&project, &input()?),
        Command::DecisionRecord {
            project,
            run,
            permission_reference,
        } => moai_cli::decision::append(&project, &run, &input()?, &permission_reference),
        Command::AuditReduce { project } => {
            let request = input::<moai_cli::audit::AuditInput>()?;
            if request.protocol_version != contract::PROTOCOL_VERSION {
                return Err("unsupported protocol_version; supported: 1".into());
            }
            serde_json::to_value(moai_cli::audit::audit(&project, &request)?)
                .map_err(|_| "encode failed".into())
        }
        _ => unreachable!("launchers run through run_launcher"),
    }
}

/// A command-line error on an internal operation becomes a `usage` envelope (exit 2) with a
/// message that echoes no argument; stderr keeps clap's first line for people. Help, version,
/// the launchers and words that name no operation keep clap's own output.
fn usage_error(e: clap::Error) -> ExitCode {
    let op = std::env::args().nth(1).and_then(|w| operation_of(&w));
    let Some(op) =
        op.filter(|_| !matches!(e.kind(), ErrorKind::DisplayHelp | ErrorKind::DisplayVersion))
    else {
        e.exit()
    };
    let message = match e.kind() {
        ErrorKind::MissingRequiredArgument => "required arguments were not provided; see --help",
        ErrorKind::UnknownArgument => "unexpected argument; see --help",
        ErrorKind::InvalidValue | ErrorKind::ValueValidation => {
            "invalid argument value; see --help"
        }
        _ => "invalid command line; see --help",
    };
    let rendered = e.render().to_string();
    eprintln!("{}", rendered.lines().next().unwrap_or(message));
    println!("{}", contract::failure(op, contract::Kind::Usage, message));
    ExitCode::from(2)
}

fn main() -> ExitCode {
    let cli = match Cli::try_parse() {
        Ok(cli) => cli,
        Err(e) => return usage_error(e),
    };
    let op = match &cli.command {
        Some(Command::Cc(_) | Command::Glm(_) | Command::Codex(_) | Command::Session { .. })
        | None => None,
        Some(_) => std::env::args().nth(1).and_then(|w| operation_of(&w)),
    };
    let Some(op) = op else {
        return match run_launcher(cli) {
            Ok(value) => {
                println!("{value}");
                ExitCode::SUCCESS
            }
            Err(message) => {
                eprintln!("{message}");
                ExitCode::from(1)
            }
        };
    };
    match run_internal(cli.command.expect("internal operation")) {
        Ok(result) => {
            println!("{}", contract::success(op, result));
            ExitCode::SUCCESS
        }
        Err(message) => {
            println!(
                "{}",
                contract::failure(op, contract::classify(&message), &message)
            );
            eprintln!("{message}");
            ExitCode::from(1)
        }
    }
}
