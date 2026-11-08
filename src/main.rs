use clap::{error::ErrorKind, ArgGroup, Args, Parser, Subcommand};
use moai_cli::{
    agents_apply, agents_plan, apply, contract, plan, resolve, session_launch, AgentsPlan,
    AgentsRequest, Plan, Profile, SessionRequest,
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

/// `moai-cli cc|glm|codex -l|-a [options] [PROMPT]`.
#[derive(Args)]
#[command(group(ArgGroup::new("role").required(true).args(["leader", "agent"])))]
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
    /// Read-only clone/workspace identity and state roots of a path (typed JSON on stdout).
    Identity {
        path: PathBuf,
    },
    /// Read-only Claude Code check: whether a session started in PROJECT reads AGENTS.md.
    ClaudeDoctor {
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

/// Builds the session request for the launcher syntax (REQ-00F-01..04).
fn launch(
    host: &str,
    leader: bool,
    o: LaunchOpts,
    prompt: Option<String>,
) -> Result<serde_json::Value, String> {
    let role = if leader { "leader" } else { "agent" };
    let name_given = o.name.is_some();
    let name = o
        .name
        .unwrap_or_else(|| if leader { "leader" } else { "agent-1" }.into());
    let request = SessionRequest {
        role: role.into(),
        host: host.into(),
        name,
        model: o.model,
        effort: o.effort,
        // Claude defaults (operator 2026-11-08): advisor fable, permission skip (interim).
        advisor: (host != "codex" && !o.no_advisor).then(|| "fable".into()),
        skip_permissions: host != "codex" && !o.permissions,
        prompt,
        claude_bin: "claude".into(),
        codex_bin: "codex".into(),
    };
    let mut v = session_launch(&request, o.dry_run)?;
    if !leader && !name_given {
        // HOLD (00-F): automatic agent numbering is undecided; RC1 uses a fixed default.
        if let Some(notes) = v["notes"].as_array_mut() {
            notes.push("agent name defaulted to agent-1 (automatic numbering is HOLD); pass --name agent-<n>".into());
        }
    }
    Ok(v)
}

/// Internal operations (01-C #1990): subcommand word and operation name. The launchers are
/// developer tools and stay outside the envelope contract.
const OPERATIONS: [(&str, &str); 9] = [
    ("package-inspect", "package.inspect"),
    ("project-plan", "project.plan"),
    ("project-apply", "project.apply"),
    ("package-resolve", "package.resolve"),
    ("agents-plan", "agents.plan"),
    ("agents-apply", "agents.apply"),
    ("baseline-inventory", "baseline.inventory"),
    ("identity", "identity"),
    ("claude-doctor", "claude.doctor"),
];
fn operation_of(word: &str) -> Option<&'static str> {
    OPERATIONS
        .iter()
        .find(|(w, _)| *w == word)
        .map(|&(_, op)| op)
}

/// Launcher paths (`-l`, `-a`, `cc|glm|codex`, `session`): plain JSON or a plain error line.
fn run_launcher(cli: Cli) -> Result<serde_json::Value, String> {
    let Some(command) = cli.command else {
        return match (cli.leader, cli.agent) {
            (true, _) => launch("claude", true, cli.launch, None),
            (false, Some(word)) => launch(host_of(&word)?, false, cli.launch, None),
            (false, None) => Err(
                "usage: moai-cli -l | -a [cc|glm|codex] | moai-cli cc|glm|codex -l|-a | moai-cli <command>; see --help".into(),
            ),
        };
    };
    match command {
        Command::Cc(h) => launch("claude", h.leader, h.launch, h.prompt),
        Command::Glm(h) => launch("glm", h.leader, h.launch, h.prompt),
        Command::Codex(h) => launch("codex", h.leader, h.launch, h.prompt),
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
                    role,
                    host,
                    name,
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
            serde_json::to_value(plan(&project, input::<Profile>()?)?)
                .map_err(|_| "encode failed".into())
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
        Command::ClaudeDoctor { project } => moai_cli::claude::doctor(&project),
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
