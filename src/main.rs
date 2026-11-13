use clap::{error::ErrorKind, ArgGroup, Args, CommandFactory, Parser, Subcommand};
use moai_cli::{
    agents_apply, agents_plan, apply, contract, hosts::codex, plan, resolve, session_launch,
    AgentsPlan, AgentsRequest, Plan, Profile, SessionRequest,
};
use std::{
    io::{self, Read},
    path::PathBuf,
    process::ExitCode,
};

mod ops;

/// Where a 06-C command reads its claim record: the local journal, or a record file that must
/// agree with the Issue claim comment it names (the file is input, not an authority).
#[derive(Args)]
struct RecordSource {
    /// The 16-hex task key of a record in the local claim journal.
    #[arg(long)]
    task: Option<String>,
    /// A ClaimRecord JSON file (a claim relayed outside the local journal).
    #[arg(long)]
    record: Option<PathBuf>,
    /// The Issue claim comment as JSON (`id`, `body`); required with --record.
    #[arg(long)]
    claim_comment: Option<PathBuf>,
}

impl RecordSource {
    fn load(&self, project: &std::path::Path) -> Result<moai_cli::claim::ClaimRecord, String> {
        moai_cli::worktree::load_record(
            project,
            self.task.as_deref(),
            self.record.as_deref(),
            self.claim_comment.as_deref(),
        )
    }
}

#[derive(Parser)]
#[command(
    name = "moai-cli",
    // One `main` serves both binary names (`moai`, `moai-cli`; #2790 R1): usage and help name
    // `moai-cli` whichever name ran, so the two binaries print the same text.
    bin_name = "moai-cli",
    version = moai_cli::VERSION_SIGNATURE,
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
    /// Session name (default `leader` for -l, the project's next `agent-N` for -a; none without a role).
    #[arg(long)]
    name: Option<String>,
    /// Main model (Claude hosts default `opus`; codex uses its own default unless set).
    #[arg(long)]
    model: Option<String>,
    /// Claude effort (default `medium`).
    #[arg(long)]
    effort: Option<String>,
    /// Advisor model on Claude hosts (fable|opus|sonnet or a full claude-* id). None by default;
    /// the user setting `launcher.advisor` can turn it on. May bill to usage credits.
    #[arg(long)]
    advisor: Option<String>,
    /// Pass no `--advisor` even when the user setting `launcher.advisor` turns it on.
    #[arg(long)]
    no_advisor: bool,
    /// Keep permission prompts: no `--dangerously-skip-permissions` (Claude hosts) and the
    /// `workspace-write` / `on-request` sandbox on codex. Without it a -l|-a session has full access;
    /// a session without a role never gets a permission flag.
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
    /// The one-file operations under `src/ops/` (#2617): add a file there, not a variant here.
    #[command(flatten)]
    Op(ops::Ops),
    ProjectPlan {
        project: PathBuf,
    },
    /// Detect the project's facts for the `/moai project` route (#2496): new or existing,
    /// manifests, languages, language servers, configured MCP server names, databases. Reads
    /// only and writes nothing; returns the profile with `detected` set for `project-plan`.
    ProjectDetect {
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
    /// Probe whether the sub-issue and dependency APIs answer for one Issue and plan the
    /// links: API write intents, or `Parent:`/`Blocked-By:` body lines (GET only; #2172).
    ForgeLinksProbe {
        project: PathBuf,
    },
    /// Create or update the one report comment of a run on its Issue (typed JSON payload on
    /// stdin; envelope on stdout). Large output goes to artifact references, not the body.
    ForgeReport {
        project: PathBuf,
        #[arg(long)]
        permission_reference: String,
    },
    /// Post the `moai/audit` commit status on an exact SHA (typed JSON on stdin). Uses the
    /// audit App token from `MOAI_AUDIT_APP_TOKEN` when set, else warns (#2173).
    ForgeAuditStatus {
        project: PathBuf,
        #[arg(long)]
        permission_reference: String,
    },
    /// Read-only clone/workspace identity and state roots of a path (typed JSON on stdout).
    Identity {
        path: PathBuf,
    },
    /// Build an incremental-index generation (C6 #2182): scan the tracked files, stage an
    /// immutable manifest directory, and publish it with an atomic pointer swap. A failed
    /// build keeps the previous generation and marks partial (REQ-C6-01).
    IndexBuild {
        project: PathBuf,
    },
    /// Query the incremental index (C6 #2182): read-only — resolves the published
    /// generation and reports one of fresh/stale/partial/unavailable with the detected
    /// changes (additions, deletions, modifications, untracked new files, branch switch).
    IndexQuery {
        project: PathBuf,
    },
    /// Copy a legacy key file's key into the encrypted vault (#2393 S3): only with the
    /// preview's digest and a permission reference; the read-back must match; the legacy file
    /// is kept. Destination: the current project's folder, or `_shared` with `--shared`.
    KeyImportApply {
        provider: String,
        #[arg(long)]
        plan_digest: String,
        #[arg(long)]
        permission_reference: String,
        /// Import into `~/.moai-cli/vault/_shared/` (every project) instead of the project.
        #[arg(long)]
        shared: bool,
    },
    /// Preview of moving a legacy key file into the encrypted vault (#2393 S3): which file,
    /// which destination folder, never the value. Writes nothing.
    KeyImportPlan {
        provider: String,
        /// Import into `~/.moai-cli/vault/_shared/` (every project) instead of the project.
        #[arg(long)]
        shared: bool,
    },
    /// Remove a provider's key from the vault (#2393): the current project's entry, or the
    /// shared one with `--shared`. Without `--yes` it only reports what would be removed. The
    /// machine key and any legacy key file are never touched.
    KeyRemove {
        provider: String,
        #[arg(long)]
        yes: bool,
        /// The key every project uses (`~/.moai-cli/vault/_shared/`).
        #[arg(long)]
        shared: bool,
    },
    /// Store a provider's key, encrypted, in the vault (#2393): for the current project, or
    /// for every project with `--shared`. The value is read from stdin only (never argv); an
    /// interactive terminal is refused. The answer never contains the value.
    KeySet {
        provider: String,
        /// The key every project uses (`~/.moai-cli/vault/_shared/`).
        #[arg(long)]
        shared: bool,
    },
    /// Whether a provider's entry (current project, or `--shared`) exists, is safe and
    /// decrypts (#2393). Never the value, its length or a fingerprint.
    KeyStatus {
        provider: String,
        /// The key every project uses (`~/.moai-cli/vault/_shared/`).
        #[arg(long)]
        shared: bool,
    },
    /// P09 opt-in home layout (#2175): measure the receipt/outbox/claim stores and bind
    /// a move plan to `~/.moai-cli/projects/<project_id>/`. Writes nothing.
    LayoutPlan {
        project: PathBuf,
    },
    /// Execute a confirmed layout plan: copy without deleting the originals, verify the
    /// copy, then switch the pointer (`layout.json`). The default stays project-side.
    LayoutApply {
        project: PathBuf,
        #[arg(long)]
        plan_digest: String,
        #[arg(long)]
        permission_reference: String,
    },
    /// Go MoAI migration preview (#2016) on a prepared synthetic HOME only: typed request
    /// on stdin (`hosts`, `mode`, `go_release`, `selections`); shows ownership, the
    /// update-conflict preview and the plan digest. Writes nothing.
    MigratePlan {
        /// Synthetic home: a temporary directory carrying `.moai-cli-synthetic-home`.
        home: PathBuf,
    },
    /// Apply a previewed migration (#2016): the exact plan on stdin must equal a fresh
    /// recompute and `--plan-digest`. Backs up, then copies only the selected owned files
    /// to `.moai-cli/imports/`; the Go source is kept.
    MigrateApply {
        home: PathBuf,
        #[arg(long)]
        plan_digest: String,
        #[arg(long)]
        permission_reference: String,
    },
    /// Read-only recovery plan for one migration apply (#2016), from its backup manifest.
    MigrateRecoveryPlan {
        home: PathBuf,
        /// The apply's backup id (its plan digest).
        #[arg(long)]
        backup: String,
    },
    /// Remove only the imports one migration apply published (#2016); changed files are
    /// kept. The recovery plan travels on stdin and must match a fresh recompute.
    MigrateRecover {
        home: PathBuf,
        #[arg(long)]
        backup: String,
        #[arg(long)]
        plan_digest: String,
        #[arg(long)]
        permission_reference: String,
    },
    /// Claim a task (06-A #2005): typed JSON request on stdin; posts exactly one claim
    /// comment (⑤) behind the same-host journal serialization (①) and records the record.
    ClaimAcquire {
        project: PathBuf,
    },
    /// Wait for one of N machine-wide heavy-run slots, run COMMAND holding it, and exit with
    /// COMMAND's code (#2677). Same folder and lock as `tools/heavy_check_slot.py`; while
    /// waiting and on failure it prints the load averages and free disk space. Unix only.
    HeavySlotAcquire {
        /// The machine's slot count; every caller passes the same value (the leader's limit).
        #[arg(long)]
        max: usize,
        /// The caller's name, shown by `heavy-slot-status`.
        #[arg(long)]
        name: String,
        /// Slot folder (default `${XDG_CACHE_HOME:-~/.cache}/moai-cli/heavy-slots`).
        #[arg(long)]
        dir: Option<PathBuf>,
        /// Start nothing (exit 75) while the working folder's file system has less free space,
        /// in decimal GB; 0 turns it off.
        #[arg(long, default_value_t = moai_cli::heavy_slot::DEFAULT_LIMITS.min_free_gb)]
        min_free_gb: u64,
        /// Start nothing (exit 75) while swap use is above this percent; 100 turns it off.
        #[arg(long, default_value_t = moai_cli::heavy_slot::DEFAULT_LIMITS.max_swap_pct)]
        max_swap_pct: u64,
        /// The heavy command and its arguments, after `--`.
        #[arg(last = true)]
        command: Vec<String>,
    },
    /// List the heavy-run slots held now, the slot paths refused, and the machine's load and
    /// free disk space (#2677). Reads only.
    HeavySlotStatus {
        #[arg(long)]
        dir: Option<PathBuf>,
    },
    /// Self-claim the next card (rc.2 A2 #2161): typed JSON request on stdin; re-reads the
    /// open Issues, skips epics, pull requests and open-blocked cards, and claims the first
    /// candidate through `claim-acquire` under one journal lock.
    QueueNext {
        project: PathBuf,
    },
    /// List the open cards a lane may take (#2675): typed JSON `{schema_version, owner, repo}`
    /// on stdin; GET only. Every open Issue lands in one bucket (free, reserved, claimed, held,
    /// epics, blocked); an empty result says `free_count: 0`.
    QueueFree {
        project: PathBuf,
    },
    /// Read the merge train in a pinned leader Issue comment; GET only, no local store.
    QueueQuery {
        project: PathBuf,
    },
    /// Leader supervision tick (rc.2 A3 #2162): typed JSON request on stdin; GET-only read
    /// of open PRs, reports only what changed since the last tick, drafts assignee hints.
    LeaderTick {
        project: PathBuf,
    },
    /// Record a dispatch revision over durable S1 messages (typed JSON stdin).
    DispatchRecord {
        project: PathBuf,
        #[arg(long)]
        permission_reference: String,
    },
    /// Record received ACK and explicit accepted/deferred/rejected decision, separately.
    DispatchAck {
        project: PathBuf,
        #[arg(long)]
        permission_reference: String,
    },
    /// Observe dispatch revisions, decisions and the five-minute ACK alert (GET only).
    /// Never authorizes a claim or an audit gate.
    DispatchCheck,
    /// Publish a durable message from typed JSON on stdin; queued is not receiver ACK
    /// or business acceptance.
    MessageRecord {
        project: PathBuf,
        #[arg(long)]
        permission_reference: String,
    },
    /// Publish an explicit receiver ACK from typed JSON on stdin; not business acceptance
    /// or completion.
    MessageAck {
        project: PathBuf,
        #[arg(long)]
        permission_reference: String,
    },
    /// Reconcile messages from typed JSON on stdin: remote GET only, with local
    /// reconciliation writes; never publishes messages or ACKs.
    MessageReconcile {
        project: PathBuf,
    },
    /// Lane handoff gate (rc.2 A10 #2264): typed JSON request on stdin; read-only
    /// verification that the worktree is safe to exit — status clean, HEAD ==
    /// observed remote == PR head, claim trailers on lane-authored non-merge commits, leader report
    /// record present, no publish_unknown outbox effects. A failed status blocks
    /// the exit (kind blocked, exit 1) with a preservation note.
    LaneHandoffCheck {
        project: PathBuf,
    },
    /// Lane status table (rc.2 A11 #2265): typed JSON request on stdin; GET-only read of the
    /// named Issues, their open PRs and the local worktrees; writes nothing.
    LaneStatus {
        project: PathBuf,
    },
    /// Per-item writing-language setup (#2293): asks one question per item and returns a
    /// complete profile for `project-plan`; writes nothing. `--questions` prints the question
    /// set for a host question tool, `--answers` reads that tool's `{item: code}` JSON on
    /// stdin, and neither flag asks on the terminal (prompts on stderr, answers on stdin).
    LanguageSetup {
        project: PathBuf,
        #[arg(long, conflicts_with = "answers")]
        questions: bool,
        #[arg(long)]
        answers: bool,
        /// Conversation language code for the conversation-default items; else the
        /// profile's `language.user_chat`, else `en`.
        #[arg(long)]
        conversation: Option<String>,
    },
    /// Re-verify claim ownership before a commit or push (rc.2 A9 #2255); GET only. Typed
    /// request on stdin, or `--hook` with a host PreToolUse input (refusal exits 2).
    ClaimVerify {
        project: PathBuf,
        #[arg(long)]
        hook: bool,
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
    /// Explicit leader handoff (06-B #2006): liveness probe, worktree evidence, close as
    /// `handed_off`; an alive or indeterminate owner needs --force-explicit.
    ClaimHandoff {
        project: PathBuf,
        /// The 16-hex task key an acquire returned.
        #[arg(long)]
        task: String,
        /// The epoch the request saw; a mismatch is a `stale_request` refusal (REQ-06B-01).
        #[arg(long = "from-epoch")]
        from_epoch: u64,
        #[arg(long)]
        reason: String,
        /// Override the alive/indeterminate gate; the reason stays required.
        #[arg(long = "force-explicit")]
        force_explicit: bool,
        /// This host's id; the liveness probe runs only for a local-process owner whose
        /// host_id matches.
        #[arg(long)]
        host_id: Option<String>,
    },
    /// Record an external-write intent before it is sent (REQ-06B-06); payload JSON on
    /// stdin gives the intent digest. A stale epoch is fenced (REQ-06B-08).
    EffectIntend {
        project: PathBuf,
        #[arg(long)]
        task: String,
        /// The epoch the caller believes it is publishing under.
        #[arg(long)]
        epoch: u64,
        /// Operation word (portable name, 07-B format).
        #[arg(long)]
        operation: String,
        #[arg(long = "project-id")]
        project_id: String,
        #[arg(long = "local-id")]
        local_id: String,
        #[arg(long)]
        attempt: u32,
        /// The publish kind, e.g. comment_post or pr_create.
        #[arg(long)]
        kind: String,
    },
    /// Confirm an intended effect after the write returned, or mark it uncertain.
    EffectConfirm {
        project: PathBuf,
        #[arg(long)]
        task: String,
        #[arg(long = "operation-id")]
        operation_id: String,
        /// What the remote returned (empty allowed for uncertain).
        #[arg(long = "response-id", default_value = "")]
        response_id: String,
        /// The observed outcome; `uncertain` marks the interrupted case.
        #[arg(long = "observed-state")]
        observed_state: String,
    },
    /// Reconcile uncertain effects against the remote by their stable marker (GET only,
    /// never re-creates). `--remote-probe` replaces the built-in comment lookup (tests).
    EffectsReconcile {
        project: PathBuf,
        #[arg(long)]
        task: String,
        /// Probe command: exit 0 + `{"remote_id":...}` found, 4 absent, 5 + seconds
        /// rate limited.
        #[arg(long = "remote-probe")]
        remote_probe: Option<String>,
    },
    /// Flow control (A13 #2279): WIP cap check over the open pull requests of the
    /// repository — the assignment-refusal signal (the refusal itself is the
    /// leader's act). GET-only.
    FlowWipCheck {
        /// GitHub owner (e.g. modu-ai).
        #[arg(long)]
        owner: String,
        /// GitHub repository name.
        #[arg(long)]
        repo: String,
        /// The lane account whose open PRs are counted (author login).
        #[arg(long)]
        lane: String,
        /// Lane open-PR cap (user-specified default 2, not a measured baseline).
        #[arg(long, default_value_t = moai_cli::flow::DEFAULT_LANE_CAP)]
        lane_cap: u64,
        /// Total open-PR cap (user-specified default 8, not a measured baseline).
        #[arg(long, default_value_t = moai_cli::flow::DEFAULT_TOTAL_CAP)]
        total_cap: u64,
    },
    /// Read-only Claude Code check: whether a session started in PROJECT reads AGENTS.md.
    ClaudeDoctor {
        project: PathBuf,
    },
    /// Preview a three-way merge of the managed keys (`outputStyle`, `statusLine`, `permissions`) on stdin
    /// into PROJECT/.claude/settings.json (#2239); other keys are refused. Writes nothing.
    ClaudeSettingsPlan {
        project: PathBuf,
        /// `project` (default): .claude/settings.json. `local`: .claude/settings.local.json, the
        /// file Claude Code also reads for sessions in a linked worktree (#2581).
        #[arg(long, default_value = "project")]
        scope: String,
    },
    /// Apply a previewed settings merge: same stdin, the preview's digest and a permission
    /// reference. Writes the project settings file and its receipt only; home is refused.
    ClaudeSettingsApply {
        project: PathBuf,
        #[arg(long)]
        plan_digest: String,
        #[arg(long)]
        permission_reference: String,
        /// The scope the plan was made for (#2581); the digest binds it.
        #[arg(long, default_value = "project")]
        scope: String,
        /// The `confirm_token` a conflict plan offered (#2383): after the user confirms, keep
        /// the existing file in `.moai-cli/settings-backups/` and replace the conflicting keys.
        #[arg(long)]
        confirm_replace: Option<String>,
    },
    /// Preview managed dotted keys from stdin in PROJECT/.codex/config.toml; never writes.
    CodexConfigPlan {
        project: PathBuf,
    },
    /// Apply a previewed project config with the same stdin, digest and permission reference.
    /// User HOME and CODEX_HOME config targets are refused; trust is never changed.
    CodexConfigApply {
        project: PathBuf,
        #[arg(long)]
        plan_digest: String,
        #[arg(long)]
        permission_reference: String,
    },
    /// Preview the MoAI hook set for one host in PROJECT (rc.2 I1 #2238); writes nothing.
    HooksPlan {
        project: PathBuf,
        /// `claude` (.claude/settings.json `hooks`) or `codex` (.codex/hooks.json).
        #[arg(long)]
        host: String,
    },
    /// Write a previewed hook set and its receipt; requires the plan digest.
    HooksApply {
        project: PathBuf,
        #[arg(long)]
        host: String,
        #[arg(long)]
        plan_digest: String,
        #[arg(long)]
        permission_reference: String,
    },
    /// Hook handler: record one allow-listed receipt line from the host's hook stdin under
    /// ./.moai-cli/receipts/; never fails the hook.
    HookReceipt {
        #[arg(long)]
        host: String,
    },
    /// Hook engine (#2391): run one MoAI hook duty for a host and print the host's native
    /// answer; reads the host's hook JSON on stdin, no network.
    HookDispatch {
        /// `claude` or `codex`.
        #[arg(long)]
        host: String,
        /// `session-check`, `tool-preflight`, `tool-observe` or `turn-validation`.
        #[arg(long)]
        event: String,
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
    /// Install a planned Codex projection into its explicit target (I4 #2241). Refuses a
    /// stale plan, a source-manifest mismatch or a conflicting target file before writing.
    CodexProjectionApply {
        package: PathBuf,
        /// Absolute skill directory named `moai`; the same target the plan was made for.
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
        #[arg(long)]
        allow_implicit_invocation: Option<bool>,
        #[arg(long)]
        plan_digest: String,
        #[arg(long)]
        permission_reference: String,
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
    /// Plan the branch and worktree of a claim (06-C #2007). Writes nothing.
    WorktreePlan {
        project: PathBuf,
        #[command(flatten)]
        source: RecordSource,
        /// Branch prefix (HOLD: no default; the user decides it).
        #[arg(long)]
        branch_prefix: String,
        /// Worktree path; default `<HOME>/.moai/worktrees/<project>-<issue>-<logical_key>` (#2811).
        #[arg(long)]
        path: Option<PathBuf>,
    },
    /// Create the planned worktree with `git worktree add --lock -b` (06-C #2007); refuses
    /// without `--force` when the branch, path or task already has a worktree.
    WorktreeApply {
        project: PathBuf,
        #[command(flatten)]
        source: RecordSource,
        #[arg(long)]
        branch_prefix: String,
        #[arg(long)]
        path: Option<PathBuf>,
        #[arg(long)]
        plan_digest: String,
        /// Test seam (WT-5): the free-space floor in bytes instead of the decided 30 GiB.
        #[arg(long, hide = true, default_value_t = moai_cli::worktree::DISK_FLOOR_BYTES)]
        disk_floor_bytes: u64,
        /// Test seam (WT-5): the project-total warning line in bytes instead of 40 GiB.
        #[arg(long, hide = true, default_value_t = moai_cli::worktree::PROJECT_WARN_BYTES)]
        warn_total_bytes: u64,
    },
    /// Read-only removal plan (WT-4, #2380): per linked worktree `remove`, `trim` (cargo
    /// build output only) or `skip` with reasons, and a digest. Deletes nothing.
    WorktreeGcPlan {
        project: PathBuf,
        /// Local ref of the integration branch, e.g. `origin/v2026.11.0`; a worktree whose
        /// HEAD it contains counts as merged. Nothing is fetched.
        #[arg(long)]
        integration_ref: String,
        /// A change newer than this many minutes keeps a worktree (design: 120).
        #[arg(long, default_value_t = moai_cli::worktree_gc::DEFAULT_RECENT_MINUTES)]
        recent_minutes: u64,
        /// Time budget in milliseconds; past it the remaining worktrees are `skip`.
        #[arg(long, default_value_t = moai_cli::worktree_gc::DEFAULT_BUDGET_MS)]
        budget_ms: u64,
        /// Test seam: the current time in Unix seconds instead of the clock.
        #[arg(long, hide = true)]
        now_unix: Option<u64>,
    },
    /// Act on a `worktree-gc-plan` (WT-9, #2498): recomputes the plan with the same inputs,
    /// refuses before any write unless its digest equals `--plan-digest`, then trims the
    /// listed cargo output and removes the listed worktrees (no `--force`; branches kept).
    WorktreeGcApply {
        project: PathBuf,
        #[arg(long)]
        integration_ref: String,
        #[arg(long, default_value_t = moai_cli::worktree_gc::DEFAULT_RECENT_MINUTES)]
        recent_minutes: u64,
        #[arg(long, default_value_t = moai_cli::worktree_gc::DEFAULT_BUDGET_MS)]
        budget_ms: u64,
        /// Test seam: the current time in Unix seconds instead of the clock.
        #[arg(long, hide = true)]
        now_unix: Option<u64>,
        #[arg(long)]
        plan_digest: String,
        #[arg(long)]
        permission_reference: String,
    },
    /// Read-only space report of the repository's worktrees, their build folders and the
    /// temporary `moai-*` folders (WT-1, #2380). Facts only; removal candidates are not
    /// evaluated. Writes nothing.
    WorktreeReport {
        project: PathBuf,
        /// Folder whose `moai-*` children are classified (default `/tmp`).
        #[arg(long)]
        tmp_root: Option<PathBuf>,
        /// Time budget in milliseconds for the size walk; past it a class is `partial`.
        #[arg(long, default_value_t = moai_cli::worktree_report::DEFAULT_BUDGET_MS)]
        budget_ms: u64,
    },
    /// Code Map query (C9 #2185): read-only — the published generation's map records
    /// (file and tag nodes, `annotated_by` edges, missing coverage) of the requested paths
    /// or directory prefixes; `unavailable` when no index records can be read.
    AtlasMap {
        project: PathBuf,
    },
    /// Code Graph query (C9 #2185): read-only — a bounded-hop walk over the same
    /// generation's map and graph records from the given node ids; `unavailable` when no
    /// index records can be read.
    AtlasGraph {
        project: PathBuf,
    },
    /// Non-vector-first search (C7 #2183): read-only — exact paths first, then literal
    /// matches over the published index generation (or the tracked files), each result with
    /// a source citation; the typed-edge step is reported unavailable (#2408).
    Search {
        project: PathBuf,
    },
    /// write_scope overlap of a claim against the other active claims (06-C #2007).
    ScopeCheck {
        project: PathBuf,
        #[command(flatten)]
        source: RecordSource,
        /// The task the leader orders first (task key, logical key or `#<issue>`).
        #[arg(long)]
        integration_order: Option<String>,
    },
    /// Semantic parity of TEMPLATE vs GENERATED (optional third DOC set, D6 #2192):
    /// per-key classification with the 00-A vocabulary (missing / extra /
    /// value_mismatch / identical / doc_only). Read-only.
    ParityCheck {
        template: PathBuf,
        generated: PathBuf,
        #[arg(long)]
        doc: Option<PathBuf>,
    },
    /// Read-only dual-read scan of `@MOAI:` / `@MX:` tag markers (C2 #2178): both
    /// namespaces normalize into one node shape, duplicates are diagnosed only,
    /// nothing is written. PATH is a file or a directory.
    AtlasTagScan {
        path: PathBuf,
    },
    /// Check a PR body's Self-Audit section (#2708, formerly tools/check_pr_self_audit.py):
    /// the body from --file or stdin; prints one line, `self-audit: pass (…)` (exit 0) or
    /// `self-audit: blocked missing=…` (exit 1); a usage error exits 2. Raw text, not the
    /// envelope, so the result and exit code are the ones the Python checker gave.
    PrSelfAuditCheck {
        /// Also require the `## 독립 review` section of a re-audit (#2388).
        #[arg(long)]
        reaudit: bool,
        /// Read the PR body from this file (UTF-8) instead of stdin.
        #[arg(long, value_name = "BODY.md")]
        file: Option<PathBuf>,
    },
    /// Render the Claude Code status line from the session JSON on stdin (#2350). The one
    /// operation that prints raw text instead of the JSON envelope: Claude Code shows the
    /// command's stdout as is. Always exits 0.
    Statusline {
        /// Test seam (#2350 leader decision): the render budget in milliseconds instead of
        /// the Go 800 ms. Hidden; Claude Code runs `statusline` without options. The golden
        /// test passes a large value so a git call slowed by machine load still lands; `0`
        /// is an already expired budget (no git call is started).
        #[arg(long, hide = true, value_name = "MS")]
        render_budget_ms: Option<u64>,
        /// Test seam (#2756): the current time in Unix seconds for the prompt cache timer,
        /// so goldens do not depend on the wall clock. Hidden; Claude Code never passes it.
        #[arg(long, hide = true, value_name = "SECS")]
        now_unix: Option<i64>,
        /// #2588: the detached counts refresh a render starts. Hidden; asks `gh` for the
        /// open Issue and PR counts of `--board-root` and rewrites its cache.
        #[arg(long, hide = true, requires = "board_root")]
        refresh_counts: bool,
        #[arg(long, hide = true, value_name = "DIR")]
        board_root: Option<PathBuf>,
    },
    /// Record the PR's ACK (06-C #2007, REQ-06C-11): confirm the `pr_create` intent and, once
    /// it is confirmed, move the claim record to `pr_ready`.
    PrConfirm {
        project: PathBuf,
        #[arg(long)]
        task: String,
        #[arg(long = "operation-id")]
        operation_id: String,
        /// What the remote returned, e.g. the PR number (empty allowed for uncertain).
        #[arg(long = "response-id", default_value = "")]
        response_id: String,
        /// The observed outcome, e.g. `opened`; `uncertain` marks a lost answer.
        #[arg(long = "observed-state")]
        observed_state: String,
    },
    /// Render the PR body of a claim (06-C #2007): Refs, exact SHAs, receipts, op marker.
    PrBodyRender {
        project: PathBuf,
        #[command(flatten)]
        source: RecordSource,
        /// JSON array of VERIFY receipts (05-C #2003); `[]` renders not_run.
        #[arg(long)]
        receipts: PathBuf,
        #[arg(long)]
        base_sha: String,
        #[arg(long)]
        base_branch: String,
        #[arg(long)]
        default_branch: String,
        /// Accepted and warned about; closing keywords are never rendered.
        #[arg(long)]
        allow_close: bool,
        #[arg(long)]
        not_run: Vec<String>,
        #[arg(long)]
        review_link: Option<String>,
        /// Extra bullet for the remaining-gap section (repeatable).
        #[arg(long)]
        gap: Vec<String>,
        /// Which PR of this claim epoch the body is for (#2329). 1 (default) keeps the
        /// original marker; N >= 2 gives a further PR its own operation and marker.
        #[arg(long, default_value_t = 1, value_parser = clap::value_parser!(u32).range(1..=999))]
        pr_seq: u32,
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
    /// Dry-run (D1 #2187, REQ-D1-01): which SPEC directories under
    /// `.moai-cli/specs/` are archive-eligible and why the rest are excluded.
    /// Writes nothing. `--grace-days` overrides the profile's inactivity window.
    ArchivePlan {
        project: PathBuf,
        #[arg(long)]
        grace_days: Option<u64>,
    },
    /// Move the archive-eligible SPECs (D2 #2188, REQ-D2-01): recomputes the plan and
    /// requires its `plan_digest`; each SPEC is hashed, copied, read back and indexed
    /// before its source is removed. Any mismatch cancels the move and keeps the source.
    ArchiveApply {
        project: PathBuf,
        #[arg(long)]
        plan_digest: String,
        #[arg(long)]
        permission_reference: String,
        #[arg(long)]
        grace_days: Option<u64>,
    },
    /// Restore one archived SPEC to `.moai-cli/specs/<id>` after verifying its manifest
    /// (D2 #2188).
    ArchiveRestore {
        project: PathBuf,
        #[arg(long)]
        spec: String,
        #[arg(long)]
        permission_reference: String,
    },
    /// Read-only, non-vector search over the archive index and archived text (D2 #2188).
    ArchiveSearch {
        project: PathBuf,
        #[arg(long)]
        query: String,
    },
    /// Plan project-unit `moai` skill copies read-only (I5 #2242, P04 HOLD):
    /// typed request on stdin; judgment envelope on stdout; writes nothing.
    ProjectSkillPlan {
        project: PathBuf,
    },
    /// Copy the staged `moai` skill package into the project host folders
    /// (I5 #2242): the exact request on stdin; requires `opt_in` and writes
    /// only under the caller-supplied project root.
    ProjectSkillApply {
        project: PathBuf,
    },
    /// Revert one applied project-skill copy from its rollback receipt
    /// (I5 #2242): typed request on stdin.
    ProjectSkillRevert {
        project: PathBuf,
    },
    /// Aggregate cost items across run/verify receipts (D5 #2191): typed request
    /// (receipt array) on stdin; breakdown envelope on stdout; writes nothing.
    /// External backend calls are structurally zero; tokens are absent, never
    /// estimated (REQ-D5-01).
    CostReport {},
    /// Plan project-specific generated files read-only (03-C #1997): typed request on
    /// stdin; judgment envelope on stdout; writes nothing.
    GenerationPlan {
        project: PathBuf,
    },
    /// Plan the GitHub templates of a project (08-B #2013): typed request on stdin
    /// (`github_issues` opts in); read-only.
    GithubPlan {
        project: PathBuf,
    },
    /// Plan a project's label sync from `.github/moai/labels.yml` (#2434 decision 40): typed
    /// request `{schema_version, owner, repo}` on stdin; GET only, never deletes.
    LabelsPlan {
        project: PathBuf,
    },
    /// Apply an approved label sync plan (#2434 decision 40): the exact plan on stdin,
    /// `--plan-digest` and `--permission-reference`; POST/PATCH only.
    LabelsApply {
        project: PathBuf,
        #[arg(long)]
        plan_digest: String,
        #[arg(long)]
        permission_reference: String,
    },
    /// Apply an approved GitHub template plan (08-B #2013): the exact plan on stdin,
    /// `--plan-digest` and `--permission-reference`.
    GithubApply {
        project: PathBuf,
        #[arg(long)]
        plan_digest: String,
        #[arg(long)]
        permission_reference: String,
    },
    /// Apply an approved generation plan (03-C #1997): the exact plan on stdin,
    /// `--plan-digest` and `--permission-reference`.
    GenerationApply {
        project: PathBuf,
        #[arg(long)]
        plan_digest: String,
        #[arg(long)]
        permission_reference: String,
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
    /// Read-only plan of an independent audit clone (WT-3, #2380): the single-branch,
    /// no-checkout clone commands for an exact PR head and where the evidence is kept. Runs
    /// nothing, no network.
    AuditClonePlan {
        project: PathBuf,
        #[arg(long)]
        pr: u64,
        /// The exact head commit (40 lowercase hex).
        #[arg(long)]
        head: String,
        #[arg(long)]
        head_branch: String,
        #[arg(long)]
        base_branch: String,
        /// Default: the project's `origin` URL.
        #[arg(long)]
        repo_url: Option<String>,
        /// Folder for the clone (default `/tmp`).
        #[arg(long)]
        tmp_root: Option<PathBuf>,
    },
    /// Compare one local two-parent merge with a controlled replay; evidence only, not AUDIT PASS.
    AuditDelta {
        #[arg(default_value = ".")]
        project: PathBuf,
        /// Previously audited commit (40 lowercase hex).
        #[arg(long)]
        pass_head: String,
        /// Integrated base tip (40 lowercase hex).
        #[arg(long)]
        tip: String,
        /// Candidate merge commit (40 lowercase hex).
        #[arg(long)]
        head: String,
    },
    /// Add rows (stdin) after the last row of the user-decisions log's single table (#2403).
    /// The log must equal HEAD in the worktree and index, and HEAD the given base; existing
    /// rows are never edited or removed. Preview by default; `--apply` writes. Rows go to the
    /// end of the table, so read the log by its time column, not by row order. The remote is
    /// not checked (`remote_checked: false`): fetch the tip first and pass it as `base`.
    DecisionAppend {
        project: PathBuf,
        /// Repository-relative path of the log, e.g. `docs/decisions-2026-11-08.md`.
        #[arg(long)]
        log: PathBuf,
        #[arg(long)]
        apply: bool,
        #[arg(long)]
        permission_reference: Option<String>,
    },
    /// Reserve the independent audit of one exact PR head (rc.2 #2403 item 1): typed JSON
    /// `{schema_version, owner, repo, pr, head, lane}` on stdin. Reads the PR, its commits and
    /// comments (GET); refuses a stale head or an author lane; posts at most one
    /// `moai-audit-claim` comment, then re-reads. The lowest reservation id on the head wins
    /// (user decision #2403 c6784520362). Start the audit only when `outcome` is `reserved`.
    AuditClaim,
    /// Temporary (05-D #2004; the name is fixed after P06): independent audit of PROJECT. Re-runs
    /// the approved verification itself and decides admission; receipts on stdin never count.
    AuditReduce {
        project: PathBuf,
    },
    /// Update the skill from the package installed with this binary (#2522): previews like
    /// update-plan and writes nothing unless confirmed — in a terminal it asks once; elsewhere
    /// pass the preview's digest with --confirm. HOME comes from the environment.
    Update {
        /// The `plan_digest` of the preview being confirmed; applies only if it still matches.
        #[arg(long, value_name = "PLAN_DIGEST")]
        confirm: Option<String>,
    },
    /// The deterministic review package for an AUDIT (#2614): the diff from the merge base to
    /// HEAD, the changed files, and the review instructions (`REVIEW.md` and its includes) and
    /// context (`AGENTS.md`, `CLAUDE.md`) read from BASE only. Reads git; writes nothing.
    ReviewPackage {
        /// Repository root.
        project: PathBuf,
        /// The base revision (the pull request's base branch tip).
        #[arg(long)]
        base: String,
        /// The head revision under review.
        #[arg(long)]
        head: String,
    },
    /// Format check of review findings before they are posted (#2614 AC-3): stdin
    /// `{"schema_version":1,"findings":[…]}`; each finding needs a severity, a `file:line` in
    /// HEAD, a title, evidence quoted from that line, and a confirmed verification. Prints the
    /// verdict (any `important` is FAIL), counts, post body and digest. Reads git only.
    ReviewCheck {
        /// Repository root.
        project: PathBuf,
        /// The head revision the findings cite.
        #[arg(long)]
        head: String,
    },
    /// Read-only preview of the Go-generation state in PROJECT (#2561): which files the Go
    /// manifest proves tool-made and unchanged, and why every other one stays. Writes nothing.
    GoStatePlan {
        /// Project root carrying the Go `.moai/manifest.json`.
        project: PathBuf,
    },
    /// Carry out a confirmed go-state preview (#2561, #2696): the Go `.moai/` becomes
    /// `.moai-backup` (or `.moai-backup-<UTC>`), a new `.moai/` takes the earlier `.moai-cli`
    /// product state and the marker last; a Go folder that stays keeps the #2561 stage-2 file
    /// moves. Applies only when a fresh preview has the given digest; never overwrites.
    GoStateApply {
        project: PathBuf,
        #[arg(long)]
        plan_digest: String,
        #[arg(long)]
        permission_reference: String,
    },
    /// Read-only: what reverting one go-state apply would do (#2561 stage 2).
    GoStateRevertPlan {
        project: PathBuf,
        /// The apply's backup stamp (`YYYYMMDDTHHMMSSZ`).
        #[arg(long)]
        backup: String,
    },
    /// Undo one go-state apply (#2561, #2696): the new `.moai/` is set aside as
    /// `.moai.rust-<UTC>` and the Go backup becomes `.moai` again. All or nothing: anything that
    /// came back at an original path, or a changed backup, refuses the revert.
    GoStateRevert {
        project: PathBuf,
        #[arg(long)]
        backup: String,
        #[arg(long)]
        plan_digest: String,
        #[arg(long)]
        permission_reference: String,
    },
    /// Read-only UPDATE preview over every host projection (04-B #1999; the command name
    /// is P06 HOLD). HOME is the user home root; STAGE is the package directory to apply.
    UpdatePlan {
        /// User home root whose `.moai-cli` and host projections are inspected.
        home: PathBuf,
        /// Package directory carrying SKILL.md (staged new state).
        stage: PathBuf,
    },
    /// Apply a previewed UPDATE host by host (04-B #1999; pending→committed receipts).
    /// The previewed plan travels on stdin and must match a fresh recompute.
    UpdateApply {
        home: PathBuf,
        stage: PathBuf,
        #[arg(long)]
        plan_digest: String,
        #[arg(long)]
        permission_reference: String,
        /// Confirm one host's backup-then-replace of an unmanaged `moai` skill folder
        /// with the `confirm_token` its plan previews (#2483); repeatable.
        #[arg(long, value_name = "TOKEN")]
        confirm_backup_replace: Vec<String>,
    },
    /// Read-only recovery plan for UPDATE hosts left `pending` (04-C #2000; names P06 HOLD).
    RecoveryPlan {
        /// User home root whose `.moai-cli` host receipts and projections are read.
        home: PathBuf,
    },
    /// Revert only this request's changes on pending hosts (04-C #2000). The previewed
    /// plan travels on stdin and must match a fresh recompute.
    RecoveryApply {
        home: PathBuf,
        #[arg(long)]
        plan_digest: String,
        #[arg(long)]
        permission_reference: String,
    },
    /// Claude Code session launcher: `cc -l|-a`.
    Cc(HostLaunch),
    /// Claude Code on z.ai GLM: `glm -l|-a` (needs ANTHROPIC_AUTH_TOKEN in the environment or a key file).
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
/// The config writer takes an explicit user home. Refuse a project target that is also
/// the host's relocated user config directory before either preview or apply can run.
fn codex_config_home(project: &std::path::Path) -> Result<PathBuf, String> {
    let home = std::env::var_os("HOME")
        .or_else(|| std::env::var_os("USERPROFILE"))
        .map(PathBuf::from)
        .ok_or("home directory missing; config target refused")?;
    let home = std::fs::canonicalize(home)
        .map_err(|_| "home directory unavailable; config target refused")?;
    if let Some(path) = std::env::var_os("CODEX_HOME") {
        let user = std::fs::canonicalize(path)
            .map_err(|_| "CODEX_HOME unavailable; user config target refused")?;
        if std::fs::canonicalize(project.join(".codex")).is_ok_and(|target| target == user) {
            return Err("project .codex is CODEX_HOME; user config target refused".into());
        }
    }
    Ok(home)
}
/// Upper bound on a typed plan fed back to its `*-apply` (#2349). A plan grows with the
/// files it covers (about 330 bytes per file per host), so the 64 KiB request bound cut
/// `update-apply` off near 99 files. 16 MiB matches the package walk's byte bound
/// (`update::WALK_BYTES`); the plan is never trusted for its size, since every apply checks
/// it against `--plan-digest`.
const PLAN_INPUT_LIMIT: usize = 16 * 1024 * 1024;
fn plan_input<T: serde::de::DeserializeOwned>() -> Result<T, String> {
    let mut bytes = Vec::new();
    io::stdin()
        .take(PLAN_INPUT_LIMIT as u64 + 1)
        .read_to_end(&mut bytes)
        .map_err(|_| "stdin read failed")?;
    if bytes.len() > PLAN_INPUT_LIMIT {
        return Err("plan input exceeds 16 MiB".into());
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
    use moai_cli::session::{launch_settings, peek_agent_number, take_agent_number, AdvisorSource};
    // User decision 2026-11-10 (#2359, comment 6760976883): `-l|-a` stay. `-l` is `leader`,
    // `-a` is `agent-N` from the project's counter; access is resolved before numbering.
    let role = leader.map(|l| if l { "leader" } else { "agent" });
    // `-a` without `--name` is numbered; `--name` takes no number.
    let numbered = leader == Some(false) && o.name.is_none();
    let name = o
        .name
        .or_else(|| (leader == Some(true)).then(|| "leader".into()));
    // Validate user configuration even for roleless sessions or a CLI override. Resolution
    // precedes host composition and any agent-number allocation. No advisor by default (user
    // decision 2026-11-12, #2476): only `--advisor` or the user setting `launcher.advisor`.
    let settings = launch_settings(
        host,
        role.is_some(),
        o.permissions,
        o.advisor.as_deref(),
        o.no_advisor,
    )?;
    let advisor_note = (host == "claude"
        && settings.advisor.as_ref().map(|(_, s)| *s) == Some(AdvisorSource::UserSetting))
    .then_some("--advisor comes from the user setting launcher.advisor (~/.moai-cli/config/launcher.yaml); --no-advisor turns it off");
    let mut request = SessionRequest {
        role: role.map(str::to_string),
        host: host.into(),
        name,
        model: o.model,
        effort: o.effort,
        advisor: settings.advisor.map(|(a, _)| a),
        skip_permissions: settings.skip_permissions,
        prompt,
        claude_bin: "claude".into(),
        codex_bin: "codex".into(),
    };
    // #2481 (user decisions 6810797000, 6810818827): the first launch in a folder that is not a
    // project initializes it (git init, settings, AGENTS.md). Options are composed first, so a
    // launch refused anyway writes nothing; numbering follows the initialization.
    let init = match std::env::current_dir() {
        Ok(here) => {
            let mut probe = request.clone();
            if numbered {
                probe.name = Some("agent-0".into());
            }
            moai_cli::session::compose(&probe)?;
            moai_cli::launcher_init::prepare(&here, host, o.dry_run)?
        }
        Err(_) => moai_cli::launcher_init::Prepared::default(),
    };
    if let (false, Some(note)) = (o.dry_run, &init.note) {
        eprintln!("{note}");
    }
    let mut number_note = None;
    if numbered {
        let here = std::env::current_dir()
            .map_err(|_| "agent numbering: the current directory is unreadable; pass --name")?;
        let n = if o.dry_run && init.git_pending {
            number_note = Some(
                "agent-N preview: --dry-run takes no number (the store is created by git init)"
                    .to_string(),
            );
            1
        } else if o.dry_run {
            let dir = moai_cli::session::agent_seq_dir(&here)?;
            number_note = Some(format!(
                "agent-N preview: --dry-run takes no number (store {})",
                dir.display()
            ));
            peek_agent_number(&here)?
        } else {
            // Compose first, so a launch that is refused anyway (a blocked glm, an invalid
            // option) takes no number.
            request.name = Some("agent-0".into());
            moai_cli::session::compose(&request)?;
            let dir = moai_cli::session::agent_seq_dir(&here)?;
            let n = take_agent_number(&here)?;
            number_note = Some(format!("agent number {n} taken (store {})", dir.display()));
            n
        };
        request.name = Some(format!("agent-{n}"));
    }
    let mut v = session_launch(&request, o.dry_run)?;
    if let Some(notes) = v["notes"].as_array_mut() {
        notes.extend(advisor_note.map(Into::into));
        notes.extend(init.note.map(Into::into));
        notes.extend(number_note.map(Into::into));
    }
    Ok(v)
}

/// Operation name of an internal command (01-C #1990); `None` for the launchers, which are
/// developer tools outside the envelope contract. Exhaustive on purpose (#2254, REQ-D8-01):
/// a `Command` variant without an arm here does not compile, where the old word table only
/// panicked at run time. Put a new arm at its alphabetical place, not at the end.
fn operation(command: &Command) -> Option<&'static str> {
    use Command::*;
    Some(match command {
        // Not envelope operations: the launchers, and the raw-text statusline (#2350).
        Cc(_)
        | Glm(_)
        | Codex(_)
        | Session { .. }
        | Statusline { .. }
        | PrSelfAuditCheck { .. } => return None,
        Op(op) => op.operation(),
        AgentsApply { .. } => "agents.apply",
        AgentsPlan { .. } => "agents.plan",
        ApplyRecover { .. } => "apply.recover",
        ArchiveApply { .. } => "archive.apply",
        ArchivePlan { .. } => "archive.plan",
        ArchiveRestore { .. } => "archive.restore",
        ArchiveSearch { .. } => "archive.search",
        AtlasTagScan { .. } => "atlas_tag.scan",
        AuditClaim => "audit.claim",
        AuditClonePlan { .. } => "audit.clone.plan",
        AuditDelta { .. } => "audit.delta",
        AuditReduce { .. } => "audit.reduce",
        BaselineInventory { .. } => "baseline.inventory",
        ClaimAcquire { .. } => "claim.acquire",
        ClaimHandoff { .. } => "claim.handoff",
        ClaimRelease { .. } => "claim.release",
        ClaimShow { .. } => "claim.show",
        ClaimVerify { .. } => "claim.verify",
        ClaudeDoctor { .. } => "claude.doctor",
        ClaudeSettingsApply { .. } => "claude.settings_apply",
        ClaudeSettingsPlan { .. } => "claude.settings_plan",
        CodexConfigApply { .. } => "codex.config_apply",
        CodexConfigPlan { .. } => "codex.config_plan",
        CodexDoctor { .. } => "codex.doctor",
        CodexProjectionApply { .. } => "codex.projection_apply",
        CodexProjectionPlan { .. } => "codex.projection_plan",
        ContextBundle { .. } => "context.bundle",
        ContextVerify { .. } => "context.verify",
        CostReport { .. } => "cost.report",
        DecisionAppend { .. } => "decision.append",
        DecisionRecord { .. } => "decision.record",
        EffectConfirm { .. } => "effect.confirm",
        EffectIntend { .. } => "effect.intend",
        EffectsReconcile { .. } => "effects.reconcile",
        FlowWipCheck { .. } => "flow.wip_check",
        ForgeAuditStatus { .. } => "forge.audit_status",
        ForgeLinksProbe { .. } => "forge.links_probe",
        ForgeRead { .. } => "forge.read",
        ForgeReport { .. } => "forge.report",
        GenerationApply { .. } => "generation.apply",
        GenerationPlan { .. } => "generation.plan",
        GithubApply { .. } => "github.apply",
        GithubPlan { .. } => "github.plan",
        HookDispatch { .. } => "hook.dispatch",
        HookReceipt { .. } => "hook.receipt",
        HooksApply { .. } => "hooks.apply",
        HooksPlan { .. } => "hooks.plan",
        Identity { .. } => "identity",
        IndexBuild { .. } => "index.build",
        IndexQuery { .. } => "index.query",
        InstallInspect { .. } => "install.inspect",
        KeyImportApply { .. } => "key.import_apply",
        LabelsApply { .. } => "labels.apply",
        LabelsPlan { .. } => "labels.plan",
        KeyImportPlan { .. } => "key.import_plan",
        KeyRemove { .. } => "key.remove",
        KeySet { .. } => "key.set",
        KeyStatus { .. } => "key.status",
        LaneHandoffCheck { .. } => "lane_handoff.check",
        LaneStatus { .. } => "lane.status",
        LanguageSetup { .. } => "language.setup",
        LayoutApply { .. } => "layout.apply",
        LayoutPlan { .. } => "layout.plan",
        LeaderTick { .. } => "leader.tick",
        DispatchRecord { .. } => "dispatch.record",
        DispatchAck { .. } => "dispatch.ack",
        DispatchCheck => "dispatch.check",
        MessageAck { .. } => "message.ack",
        MessageReconcile { .. } => "message.reconcile",
        MessageRecord { .. } => "message.record",
        MigrateApply { .. } => "migrate.apply",
        MigratePlan { .. } => "migrate.plan",
        MigrateRecover { .. } => "migrate.recover",
        MigrateRecoveryPlan { .. } => "migrate.recovery_plan",
        PackageResolve { .. } => "package.resolve",
        ParityCheck { .. } => "parity.check",
        PlanAdmit { .. } => "plan.admit",
        PrBodyRender { .. } => "pr.body.render",
        PrConfirm { .. } => "pr.confirm",
        ProjectApply { .. } => "project.apply",
        ProjectDetect { .. } => "project.detect",
        ProjectPlan { .. } => "project.plan",
        ProjectSkillApply { .. } => "project_skill.apply",
        ProjectSkillPlan { .. } => "project_skill.plan",
        ProjectSkillRevert { .. } => "project_skill.revert",
        HeavySlotAcquire { .. } => "heavy_slot.acquire",
        HeavySlotStatus { .. } => "heavy_slot.status",
        QueueNext { .. } => "queue.next",
        QueueFree { .. } => "queue.free",
        QueueQuery { .. } => "queue.query",
        RecoveryApply { .. } => "recovery.apply",
        RecoveryPlan { .. } => "recovery.plan",
        RunVerify { .. } => "run.verify",
        ScopeCheck { .. } => "scope.check",
        Search { .. } => "search.query",
        AtlasMap { .. } => "atlas.map",
        AtlasGraph { .. } => "atlas.graph",
        SnapshotDescribe { .. } => "snapshot.describe",
        Update { .. } => "update.run",
        UpdateApply { .. } => "update.apply",
        ReviewCheck { .. } => "review.check",
        ReviewPackage { .. } => "review.package",
        GoStatePlan { .. } => "go_state.plan",
        GoStateApply { .. } => "go_state.apply",
        GoStateRevertPlan { .. } => "go_state.revert_plan",
        GoStateRevert { .. } => "go_state.revert",
        UpdatePlan { .. } => "update.plan",
        WorktreeApply { .. } => "worktree.apply",
        WorktreePlan { .. } => "worktree.plan",
        WorktreeReport { .. } => "worktree.report",
        WorktreeGcPlan { .. } => "worktree.gc.plan",
        WorktreeGcApply { .. } => "worktree.gc.apply",
    })
}

/// Operation of a subcommand word when the command line did not parse (`usage_error`): the
/// word's subcommand is parsed once with each required argument filled in, so the name
/// still comes from `operation` alone. A word that names no internal operation is `None`.
fn operation_of(word: &str) -> Option<&'static str> {
    let cli = Cli::command();
    let sub = cli.get_subcommands().find(|s| s.get_name() == word)?;
    let mut argv = vec!["moai-cli".to_string(), word.to_string()];
    for arg in sub.get_arguments().filter(|a| a.is_required_set()) {
        let value = arg
            .get_possible_values()
            .first()
            .map_or_else(|| "1".to_string(), |v| v.get_name().to_string());
        if let Some(long) = arg.get_long() {
            argv.push(format!("--{long}"));
        }
        argv.push(value);
    }
    operation(Cli::try_parse_from(argv).ok()?.command.as_ref()?)
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
            (false, None) => {
                // Bare no-argument run: the banner rides stdout for humans only
                // (Issue #2363); the usage error contract below is unchanged.
                moai_cli::banner::maybe_print_banner();
                Err(
                    "usage: moai-cli -l | -a [cc|glm|codex] | moai-cli cc|glm|codex -l|-a | moai-cli <command>; see --help".into(),
                )
            }
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
/// Journal-writing operations and message operations return typed errors directly;
/// the remaining operations keep the classified `String` fallback.
/// Commands that may run longer than two seconds and print static progress lines on a stderr
/// terminal (#2806, design #2799 D2 §5). The `wait`, `gate-loop`, `gate-check` and
/// `labels-check` operations install the same sink in their own files.
fn shows_progress(command: &Command) -> bool {
    matches!(
        command,
        Command::IndexBuild { .. }
            | Command::LabelsPlan { .. }
            | Command::QueueFree { .. }
            | Command::LeaderTick { .. }
            | Command::LaneStatus { .. }
            | Command::ForgeRead { .. }
            | Command::FlowWipCheck { .. }
            | Command::DispatchCheck
    )
}

fn run_internal(
    command: Command,
) -> std::result::Result<serde_json::Value, moai_cli::contract::OpError> {
    if shows_progress(&command) {
        moai_cli::progress::install_stderr();
    }
    match command {
        Command::ClaimAcquire { project } => moai_cli::claim::acquire_cli(
            &project,
            input::<moai_cli::claim::ClaimRequest>()?,
            &moai_cli::forge::transport::GhTransport::from_env()?,
            &moai_cli::forge::policy::SystemClock,
        ),
        Command::QueueNext { project } => moai_cli::queue::next_cli(
            &project,
            input::<moai_cli::queue::QueueRequest>()?,
            &moai_cli::forge::transport::GhTransport::from_env()?,
            &moai_cli::forge::policy::SystemClock,
        ),
        Command::QueueFree { project: _ } => moai_cli::queue::free_cli(
            input::<moai_cli::queue::FreeRequest>()?,
            &moai_cli::forge::transport::GhTransport::from_env()?,
            &moai_cli::forge::policy::SystemClock,
        ),
        Command::DispatchRecord {
            project,
            permission_reference,
        } => moai_cli::dispatch::record_cli(
            &project,
            input::<moai_cli::dispatch::RecordRequest>()?,
            &permission_reference,
            &moai_cli::forge::transport::GhTransport::from_env()?,
            &moai_cli::forge::policy::SystemClock,
        ),
        Command::DispatchAck {
            project,
            permission_reference,
        } => moai_cli::dispatch::ack_cli(
            &project,
            input::<moai_cli::dispatch::AckRequest>()?,
            &permission_reference,
            &moai_cli::forge::transport::GhTransport::from_env()?,
            &moai_cli::forge::policy::SystemClock,
        ),
        Command::DispatchCheck => moai_cli::dispatch::check_cli(
            input::<moai_cli::dispatch::CheckRequest>()?,
            &moai_cli::forge::transport::GhTransport::from_env()?,
            &moai_cli::forge::policy::SystemClock,
        ),
        Command::MessageRecord {
            project,
            permission_reference,
        } => moai_cli::message::record_cli(
            &project,
            input::<moai_cli::message::RecordRequest>()?,
            &permission_reference,
            &moai_cli::forge::transport::GhTransport::from_env()?,
            &moai_cli::forge::policy::SystemClock,
        ),
        Command::MessageAck {
            project,
            permission_reference,
        } => moai_cli::message::ack_cli(
            &project,
            input::<moai_cli::message::AckRequest>()?,
            &permission_reference,
            &moai_cli::forge::transport::GhTransport::from_env()?,
            &moai_cli::forge::policy::SystemClock,
        ),
        Command::MessageReconcile { project } => moai_cli::message::reconcile_cli(
            &project,
            input::<moai_cli::message::ReconcileRequest>()?,
            &moai_cli::forge::transport::GhTransport::from_env()?,
            &moai_cli::forge::policy::SystemClock,
        ),
        Command::QueueQuery { project } => Ok(moai_cli::queue_query::query_cli(
            &project,
            input::<moai_cli::queue_query::QueueQueryRequest>()?,
            &moai_cli::forge::transport::GhTransport::from_env()?,
            &moai_cli::forge::policy::SystemClock,
        )?),
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
        Command::ClaimHandoff {
            project,
            task,
            from_epoch,
            reason,
            force_explicit,
            host_id,
        } => moai_cli::handoff::claim_handoff_cli(
            &project,
            &task,
            from_epoch,
            &reason,
            force_explicit,
            host_id.as_deref(),
            &moai_cli::forge::policy::SystemClock,
        ),
        Command::EffectIntend {
            project,
            task,
            epoch,
            operation,
            project_id,
            local_id,
            attempt,
            kind,
        } => moai_cli::handoff::effect_intend_cli(
            &project,
            &task,
            epoch,
            &operation,
            &project_id,
            &local_id,
            attempt,
            &kind,
            input::<serde_json::Value>()?.to_string().as_bytes(),
            &moai_cli::forge::policy::SystemClock,
        ),
        Command::EffectConfirm {
            project,
            task,
            operation_id,
            response_id,
            observed_state,
        } => moai_cli::handoff::effect_confirm_cli(
            &project,
            &task,
            &operation_id,
            &response_id,
            &observed_state,
            &moai_cli::forge::policy::SystemClock,
        ),
        Command::EffectsReconcile {
            project,
            task,
            remote_probe,
        } => moai_cli::handoff::effects_reconcile_cli(
            &project,
            &task,
            remote_probe.as_deref(),
            &moai_cli::forge::transport::GhTransport::from_env()?,
            &moai_cli::forge::policy::SystemClock,
        ),
        Command::LayoutApply {
            project,
            plan_digest,
            permission_reference,
        } => moai_cli::layout::apply_cli(&project, &plan_digest, &permission_reference),
        Command::WorktreeApply {
            project,
            source,
            branch_prefix,
            path,
            plan_digest,
            disk_floor_bytes,
            warn_total_bytes,
        } => moai_cli::worktree::apply_and_record(
            &project,
            &source.load(&project)?,
            source.task.is_some(),
            &branch_prefix,
            path.as_deref(),
            moai_cli::home_workdir::home().ok().as_deref(),
            &plan_digest,
            moai_cli::worktree::Capacity {
                floor_bytes: disk_floor_bytes,
                warn_bytes: warn_total_bytes,
            },
            &moai_cli::forge::policy::SystemClock,
        ),
        Command::AtlasMap { project } => serde_json::to_value(moai_cli::atlas_query::atlas_map(
            &project,
            &input::<moai_cli::atlas_query::MapRequest>()?,
        )?)
        .map_err(|_| "encode failed".into()),
        Command::AtlasGraph { project } => {
            serde_json::to_value(moai_cli::atlas_query::atlas_graph(
                &project,
                &input::<moai_cli::atlas_query::GraphRequest>()?,
            )?)
            .map_err(|_| "encode failed".into())
        }
        Command::Search { project } => serde_json::to_value(moai_cli::search::search(
            &project,
            &input::<moai_cli::search::SearchRequest>()?,
        )?)
        .map_err(|_| "encode failed".into()),
        Command::ScopeCheck {
            project,
            source,
            integration_order,
        } => {
            let mine = source.load(&project)?;
            let others = moai_cli::worktree::active_records(&project)?;
            // #2404: the PR state of overlapping handed-off claims, read before the journal
            // lock; with no overlap with a `pr_ready` record the forge is not set up at all.
            let clock = moai_cli::forge::policy::SystemClock;
            let handed_off = if moai_cli::scope::needs_pr_status(&mine, &others) {
                let transport = moai_cli::forge::transport::GhTransport::from_env()?;
                let client = moai_cli::forge::Client::new(
                    &transport,
                    &clock,
                    Default::default(),
                    Default::default(),
                );
                moai_cli::scope::handed_off_status(&mine, &others, &client)
            } else {
                Default::default()
            };
            moai_cli::scope::check_and_record(
                &project,
                &mine,
                source.task.is_some(),
                &others,
                integration_order.as_deref(),
                &handed_off,
                &clock,
            )
        }
        Command::PrConfirm {
            project,
            task,
            operation_id,
            response_id,
            observed_state,
        } => moai_cli::pr_link::confirm_and_record(
            &project,
            &task,
            &operation_id,
            &response_id,
            &observed_state,
            &moai_cli::forge::policy::SystemClock,
        ),
        Command::PrBodyRender {
            project,
            source,
            receipts,
            base_sha,
            base_branch,
            default_branch,
            allow_close,
            not_run,
            review_link,
            gap,
            pr_seq,
        } => {
            let rec = source.load(&project)?;
            let bytes = std::fs::read(&receipts).map_err(|_| "receipts file not found")?;
            let receipts: Vec<moai_cli::run::Receipt> =
                serde_json::from_slice(&bytes).map_err(|_| "invalid receipts file")?;
            let head = moai_cli::worktree::head(&project)?;
            let changes = if moai_cli::worktree::is_sha(&base_sha) {
                moai_cli::worktree::changes(&project, &base_sha)?
            } else {
                Vec::new()
            };
            moai_cli::pr_link::render_and_record(
                &project,
                &rec,
                source.task.is_some(),
                &receipts,
                &moai_cli::pr_link::RenderInput {
                    base_branch: &base_branch,
                    default_branch: &default_branch,
                    base_sha: &base_sha,
                    candidate_sha: &head,
                    allow_close,
                    not_run: &not_run,
                    review_link: review_link.as_deref(),
                    changes: &changes,
                    gaps: &gap,
                    pr_seq,
                },
                &moai_cli::forge::policy::SystemClock,
            )
        }
        other => run_internal_classified(other).map_err(moai_cli::contract::OpError::from),
    }
}

fn run_internal_classified(command: Command) -> Result<serde_json::Value, String> {
    match command {
        Command::Op(op) => op.run(),
        Command::ProjectDetect { project } => moai_cli::project_detect::detect_cli(&project),
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
            plan_input::<Plan>()?,
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
            plan_input::<AgentsPlan>()?,
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
        Command::ForgeAuditStatus {
            project,
            permission_reference,
        } => moai_cli::forge::audit_status::audit_status_cli(
            &project,
            input::<moai_cli::forge::audit_status::AuditStatusRequest>()?,
            &permission_reference,
            moai_cli::forge::transport::GhTransport::from_env()?,
            &moai_cli::forge::policy::SystemClock,
        ),
        Command::ForgeRead { project } => moai_cli::forge::forge_read_cli(
            &project,
            input::<moai_cli::forge::ReadRequest>()?,
            &moai_cli::forge::transport::GhTransport::from_env()?,
            &moai_cli::forge::policy::SystemClock,
        ),
        Command::ForgeLinksProbe { project } => moai_cli::forge::links::links_probe_cli(
            &project,
            input::<moai_cli::forge::links::LinksRequest>()?,
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
        Command::IndexBuild { project } => serde_json::to_value(moai_cli::index::index_build(
            &project,
            &input::<moai_cli::index::IndexBuildRequest>()?,
        )?)
        .map_err(|_| "encode failed".into()),
        Command::IndexQuery { project } => serde_json::to_value(moai_cli::index::index_query(
            &project,
            &input::<moai_cli::index::IndexQueryRequest>()?,
        )?)
        .map_err(|_| "encode failed".into()),
        Command::KeyImportApply {
            provider,
            plan_digest,
            permission_reference,
            shared,
        } => {
            let scope = moai_cli::key_vault::scope_here(shared)?;
            moai_cli::key_vault::import_apply(
                &moai_cli::key_vault::home()?,
                &scope,
                &provider,
                &plan_digest,
                &permission_reference,
            )
        }
        Command::KeyImportPlan { provider, shared } => {
            let scope = moai_cli::key_vault::scope_here(shared)?;
            moai_cli::key_vault::import_plan(&moai_cli::key_vault::home()?, &scope, &provider)
        }
        Command::KeyRemove {
            provider,
            yes,
            shared,
        } => {
            let scope = moai_cli::key_vault::scope_here(shared)?;
            let home = moai_cli::key_vault::home()?;
            if yes {
                return moai_cli::key_vault::remove(&home, &scope, &provider, true);
            }
            // Without `--yes`: the preview removes nothing. On a terminal (stdin and stderr)
            // with an entry to delete, ask once (#2806, design #2799 D2 §7); elsewhere the
            // preview with its `next:` hint is the answer, byte for byte as before.
            let preview = moai_cli::key_vault::remove(&home, &scope, &provider, false)?;
            use std::io::IsTerminal as _;
            let terminal = io::stdin().is_terminal() && io::stderr().is_terminal();
            if !(terminal && preview["exists"] == serde_json::json!(true)) {
                return Ok(preview);
            }
            let what = format!(
                "key-remove deletes 1 file\n  {provider} key, {} scope: {}\n",
                preview["scope"].as_str().unwrap_or("?"),
                preview["path"].as_str().unwrap_or("?")
            );
            if moai_cli::progress::ask_terminal(&what, "Remove it?")? {
                moai_cli::key_vault::remove(&home, &scope, &provider, true)
            } else {
                eprintln!("next: run with --yes to remove it without asking");
                Ok(preview)
            }
        }
        Command::KeySet { provider, shared } => {
            use std::io::IsTerminal;
            if io::stdin().is_terminal() {
                return Err(
                    "key-set reads the value from stdin and refuses a terminal (the \
                            value would be echoed); pipe it in instead"
                        .into(),
                );
            }
            let scope = moai_cli::key_vault::scope_here(shared)?;
            let value = moai_cli::key_vault::read_value(io::stdin())?;
            moai_cli::key_vault::set(&moai_cli::key_vault::home()?, &scope, &provider, &value)
        }
        Command::KeyStatus { provider, shared } => {
            let scope = moai_cli::key_vault::scope_here(shared)?;
            moai_cli::key_vault::status(&moai_cli::key_vault::home()?, &scope, &provider)
        }
        Command::LaneHandoffCheck { project } => {
            let mut request = input::<moai_cli::lane_handoff::LaneHandoffRequest>()?;
            // The CLI positional project wins over any stdin value (consistency
            // with every other project-taking op).
            request.project = project.clone();

            let report = moai_cli::lane_handoff::handoff_check(&request)?;
            if report["status"] == serde_json::json!("failed") {
                eprintln!(
                    "{}",
                    serde_json::to_string_pretty(&report).unwrap_or_default()
                );
                return Err(moai_cli::lane_handoff::blocked_message(&report)
                    .unwrap_or_else(|| "lane handoff blocked".into()));
            }
            Ok(report)
        }
        Command::LaneStatus { project } => moai_cli::lane_status::status_cli(
            &project,
            input::<moai_cli::lane_status::LaneStatusRequest>()?,
            &moai_cli::forge::transport::GhTransport::from_env()?,
            &moai_cli::forge::policy::SystemClock,
        ),
        Command::LanguageSetup {
            project,
            questions,
            answers,
            conversation,
        } => {
            use moai_cli::language::{setup, SetupMode};
            let mode = if questions {
                SetupMode::Questions
            } else if answers {
                let value = input::<serde_json::Value>()?;
                SetupMode::Answers(
                    value
                        .as_object()
                        .cloned()
                        .ok_or("answers must be a JSON object")?,
                )
            } else {
                SetupMode::Interactive
            };
            setup(
                &project,
                mode,
                conversation.as_deref(),
                &mut io::stdin().lock(),
                &mut io::stderr(),
            )
        }
        Command::LeaderTick { project } => moai_cli::leader::tick_cli(
            &project,
            input::<moai_cli::leader::TickRequest>()?,
            &moai_cli::forge::transport::GhTransport::from_env()?,
            &moai_cli::forge::policy::SystemClock,
        ),
        Command::ClaimVerify { project, hook } => {
            let clock = moai_cli::forge::policy::SystemClock;
            if hook {
                // Host hook stdin can carry large tool input; read a bounded prefix.
                let mut bytes = Vec::new();
                let _ = io::stdin().take(4 * 1024 * 1024).read_to_end(&mut bytes);
                // #2348: a command the guard does not check never reaches the transport
                // setup, so a bad `MOAI_FORGE_GH_BIN` cannot turn every Bash call into a
                // refusal. `hook_cli` repeats this check and stays the single decision.
                let input: serde_json::Value =
                    serde_json::from_slice(&bytes).unwrap_or(serde_json::Value::Null);
                let command = input["tool_input"]["command"].as_str().unwrap_or("");
                if input["tool_name"] != "Bash"
                    || !moai_cli::claim_verify::is_commit_or_push(command)
                {
                    return Ok(serde_json::json!({
                        "allowed": true, "reason": "not a git commit or push"
                    }));
                }
                let transport = moai_cli::forge::transport::GhTransport::from_env()?;
                moai_cli::claim_verify::hook_cli(&project, &bytes, &transport, &clock)
            } else {
                let transport = moai_cli::forge::transport::GhTransport::from_env()?;
                let request = input::<moai_cli::claim_verify::VerifyRequest>()?;
                moai_cli::claim_verify::verify_cli(&request, &transport, &clock)
            }
        }
        Command::ClaimShow { project, task } => moai_cli::claim::show_cli(&project, &task),
        Command::MigratePlan { home } => {
            let request = input::<moai_cli::migrate::PlanRequest>()?;
            moai_cli::migrate::plan_cli(&home, &request)
        }
        Command::MigrateApply {
            home,
            plan_digest,
            permission_reference,
        } => {
            let plan = plan_input::<moai_cli::migrate::Plan>()?;
            moai_cli::migrate::apply_cli(&home, &plan, &plan_digest, &permission_reference)
        }
        Command::MigrateRecoveryPlan { home, backup } => {
            moai_cli::migrate::recovery_plan_cli(&home, &backup)
        }
        Command::MigrateRecover {
            home,
            backup,
            plan_digest,
            permission_reference,
        } => {
            let plan = plan_input::<moai_cli::migrate::RecoveryPlan>()?;
            moai_cli::migrate::recover_cli(
                &home,
                &backup,
                &plan,
                &plan_digest,
                &permission_reference,
            )
        }
        Command::LayoutPlan { project } => moai_cli::layout::plan_cli(&project),
        Command::FlowWipCheck {
            owner,
            repo,
            lane,
            lane_cap,
            total_cap,
        } => {
            let transport = moai_cli::forge::transport::GhTransport::from_env()?;
            let clock = moai_cli::forge::policy::SystemClock;
            let client = moai_cli::forge::Client::new(
                &transport,
                &clock,
                Default::default(),
                Default::default(),
            );
            moai_cli::flow::wip_check_json(&client, &owner, &repo, &lane, lane_cap, total_cap)
        }
        Command::ClaudeDoctor { project } => moai_cli::claude::doctor(&project),
        Command::HooksPlan { project, host } => {
            let home = std::env::var_os("HOME").map(PathBuf::from);
            let mut out = moai_cli::hooks::plan(&project, &host, home.as_deref())?;
            merge_binary_state(&mut out);
            // Outside the plan digest: the #2391 hook state from local evidence, read only.
            let path_env = std::env::var("PATH").unwrap_or_default();
            out["hook_status"] =
                moai_cli::hooks::status(&project, &host, home.as_deref(), &path_env)?;
            Ok(out)
        }
        Command::HooksApply {
            project,
            host,
            plan_digest,
            permission_reference,
        } => {
            let home = std::env::var_os("HOME").map(PathBuf::from);
            let mut out = moai_cli::hooks::apply(
                &project,
                &host,
                home.as_deref(),
                &plan_digest,
                &permission_reference,
            )?;
            merge_binary_state(&mut out);
            Ok(out)
        }
        // Printed by `hook_dispatch_main`, which `main` calls before any internal dispatch.
        Command::HookDispatch { .. } => {
            Err("hook-dispatch is answered before internal dispatch".into())
        }
        Command::HookReceipt { host } => {
            // Hook stdin can be large (tool output); read a bounded prefix. A cut-off object
            // fails to parse and is skipped, never an error: the hook must not fail.
            let mut bytes = Vec::new();
            let _ = io::stdin().take(4 * 1024 * 1024).read_to_end(&mut bytes);
            let now = std::time::SystemTime::now()
                .duration_since(std::time::UNIX_EPOCH)
                .map_or(0, |d| d.as_millis());
            Ok(match std::env::current_dir() {
                Ok(cwd) => moai_cli::hooks::receipt(&host, &bytes, &cwd, now),
                Err(_) => {
                    serde_json::json!({"recorded": false, "reason": "working directory unavailable"})
                }
            })
        }
        // Answered in `main` before internal dispatch (no envelope, the command's exit code).
        Command::HeavySlotAcquire { .. } => {
            Err("heavy-slot-acquire is answered before internal dispatch".into())
        }
        Command::HeavySlotStatus { dir } => {
            let dir = match dir {
                Some(d) => d,
                None => moai_cli::heavy_slot::default_dir()?,
            };
            moai_cli::heavy_slot::status(&dir)
        }
        Command::ClaudeSettingsPlan { project, scope } => {
            let scope = moai_cli::claude::SettingsScope::parse(&scope)?;
            let desired = moai_cli::claude::user_settings_desired_in(&project, &input()?, scope)?;
            moai_cli::claude::settings_plan_in(&project, &desired, scope)
        }
        Command::ClaudeSettingsApply {
            project,
            plan_digest,
            permission_reference,
            scope,
            confirm_replace,
        } => {
            let scope = moai_cli::claude::SettingsScope::parse(&scope)?;
            let desired = moai_cli::claude::user_settings_desired_in(&project, &input()?, scope)?;
            moai_cli::claude::settings_apply_in(
                &project,
                &desired,
                &plan_digest,
                &permission_reference,
                scope,
                confirm_replace.as_deref(),
            )
        }
        Command::CodexConfigPlan { project } => {
            let home = codex_config_home(&project)?;
            let desired = input()?;
            moai_cli::hosts::codex_config::config_plan(&project, &home, &desired)
        }
        Command::CodexConfigApply {
            project,
            plan_digest,
            permission_reference,
        } => {
            let home = codex_config_home(&project)?;
            let desired = input()?;
            moai_cli::hosts::codex_config::config_apply(
                &project,
                &home,
                &desired,
                &plan_digest,
                &permission_reference,
            )
        }
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
        Command::CodexProjectionApply {
            package,
            target,
            display_name,
            short_description,
            default_prompt,
            allow_implicit_invocation,
            plan_digest,
            permission_reference,
        } => codex::projection_apply(
            &package,
            &target,
            &codex::OpenAiYaml {
                display_name,
                short_description,
                default_prompt,
                allow_implicit_invocation,
            },
            &plan_digest,
            &permission_reference,
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
        Command::WorktreePlan {
            project,
            source,
            branch_prefix,
            path,
        } => moai_cli::worktree::plan(
            &project,
            &source.load(&project)?,
            &branch_prefix,
            path.as_deref(),
            moai_cli::home_workdir::home().ok().as_deref(),
        ),
        Command::WorktreeGcPlan {
            project,
            integration_ref,
            recent_minutes,
            budget_ms,
            now_unix,
        } => moai_cli::worktree_gc::plan(
            &project,
            &integration_ref,
            std::time::Duration::from_secs(recent_minutes.saturating_mul(60)),
            std::time::Duration::from_millis(budget_ms),
            now_unix,
        ),
        Command::WorktreeGcApply {
            project,
            integration_ref,
            recent_minutes,
            budget_ms,
            now_unix,
            plan_digest,
            permission_reference,
        } => moai_cli::worktree_gc::apply(
            &project,
            &integration_ref,
            std::time::Duration::from_secs(recent_minutes.saturating_mul(60)),
            std::time::Duration::from_millis(budget_ms),
            now_unix,
            &plan_digest,
            &permission_reference,
        ),
        Command::AuditClonePlan {
            project,
            pr,
            head,
            head_branch,
            base_branch,
            repo_url,
            tmp_root,
        } => moai_cli::audit_clone::plan(
            &project,
            &moai_cli::audit_clone::Request {
                pr,
                head: &head,
                head_branch: &head_branch,
                base_branch: &base_branch,
                repo_url: repo_url.as_deref(),
                tmp_root: &tmp_root.unwrap_or_else(|| PathBuf::from("/tmp")),
            },
        ),
        Command::AuditDelta {
            project,
            pass_head,
            tip,
            head,
        } => serde_json::to_value(moai_cli::audit_delta::inspect(
            &project,
            moai_cli::audit_delta::Request {
                pass_head,
                tip,
                head,
            },
        )?)
        .map_err(|_| "encode failed".into()),
        Command::WorktreeReport {
            project,
            tmp_root,
            budget_ms,
        } => moai_cli::worktree_report::report(
            &project,
            &tmp_root.unwrap_or_else(|| PathBuf::from("/tmp")),
            std::time::Duration::from_millis(budget_ms),
        ),
        Command::ParityCheck {
            template,
            generated,
            doc,
        } => moai_cli::parity::parity_check(&template, &generated, doc.as_deref()),
        Command::AtlasTagScan { path } => moai_cli::atlas_tag::scan_path(&path),
        Command::InstallInspect { json: _ } => {
            let home = std::env::var_os("HOME")
                .filter(|h| !h.is_empty())
                .map(PathBuf::from)
                .ok_or("home directory missing; install-inspect requires HOME")?;
            let path_env = std::env::var("PATH").unwrap_or_default();
            let env_dir = |k: &str| {
                std::env::var_os(k)
                    .filter(|v| !v.is_empty())
                    .map(PathBuf::from)
            };
            let (project_dir, gobin) = (env_dir("CLAUDE_PROJECT_DIR"), env_dir("GOBIN"));
            let places = moai_cli::install::CandidatePlaces {
                project_dir: project_dir.as_deref(),
                gobin: gobin.as_deref(),
            };
            moai_cli::install::inspect(&home, &path_env, &places)
        }
        Command::Update { confirm } => {
            let home = std::env::var_os("HOME")
                .filter(|h| !h.is_empty())
                .map(PathBuf::from)
                .ok_or("home directory missing; update requires HOME")?;
            // The shared confirm prompt (#2806): same terminal rule and question as before.
            moai_cli::update::update_command(&home, confirm.as_deref(), |summary| {
                moai_cli::progress::ask_terminal(
                    &format!("moai-cli update preview:\n{summary}"),
                    "Apply this update?",
                )
            })
        }
        Command::ReviewCheck { project, head } => {
            let findings: serde_json::Value = input()?;
            serde_json::to_value(moai_cli::review::check(&project, &head, &findings)?)
                .map_err(|_| "encode failed".into())
        }
        Command::ReviewPackage {
            project,
            base,
            head,
        } => serde_json::to_value(moai_cli::review::package(&project, &base, &head)?)
            .map_err(|_| "encode failed".into()),
        Command::GoStatePlan { project } => {
            serde_json::to_value(moai_cli::go_state::plan(&project)?)
                .map_err(|_| "encode failed".into())
        }
        Command::GoStateApply {
            project,
            plan_digest,
            permission_reference,
        } => serde_json::to_value(moai_cli::go_state::apply(
            &project,
            &plan_digest,
            &permission_reference,
        )?)
        .map_err(|_| "encode failed".into()),
        Command::GoStateRevertPlan { project, backup } => {
            serde_json::to_value(moai_cli::go_state::revert_plan(&project, &backup)?)
                .map_err(|_| "encode failed".into())
        }
        Command::GoStateRevert {
            project,
            backup,
            plan_digest,
            permission_reference,
        } => serde_json::to_value(moai_cli::go_state::revert(
            &project,
            &backup,
            &plan_digest,
            &permission_reference,
        )?)
        .map_err(|_| "encode failed".into()),
        Command::UpdatePlan { home, stage } => {
            let staged = moai_cli::update::stage_package(&stage)?;
            serde_json::to_value(moai_cli::update::plan_update(&home, &staged)?)
                .map_err(|_| "encode failed".into())
        }
        Command::UpdateApply {
            home,
            stage,
            plan_digest,
            permission_reference,
            confirm_backup_replace,
        } => {
            let staged = moai_cli::update::stage_package(&stage)?;
            let plan = plan_input::<moai_cli::update::UpdatePlan>()?;
            moai_cli::update::apply_update(
                &home,
                &staged,
                &plan,
                &plan_digest,
                &permission_reference,
                &confirm_backup_replace,
            )
        }
        Command::RecoveryPlan { home } => {
            serde_json::to_value(moai_cli::recovery::recovery_plan(&home)?)
                .map_err(|_| "encode failed".into())
        }
        Command::RecoveryApply {
            home,
            plan_digest,
            permission_reference,
        } => {
            let plan = plan_input::<moai_cli::recovery::RecoveryPlan>()?;
            moai_cli::recovery::recovery_apply(&home, &plan, &plan_digest, &permission_reference)
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
            serde_json::to_value(moai_cli::run::run_verify(&project, &request, &current)?)
                .map_err(|_| "encode failed".into())
        }
        Command::SnapshotDescribe { project } => moai_cli::context::snapshot(&project),
        Command::ArchivePlan {
            project,
            grace_days,
        } => serde_json::to_value(moai_cli::archive::plan(
            &project,
            std::time::SystemTime::now(),
            grace_days,
        )?)
        .map_err(|_| "encode failed".into()),
        Command::ArchiveApply {
            project,
            plan_digest,
            permission_reference,
            grace_days,
        } => moai_cli::archive::apply(
            &project,
            std::time::SystemTime::now(),
            grace_days,
            &plan_digest,
            &permission_reference,
        ),
        Command::ArchiveRestore {
            project,
            spec,
            permission_reference,
        } => moai_cli::archive::restore(&project, &spec, &permission_reference),
        Command::ArchiveSearch { project, query } => moai_cli::archive::search(&project, &query),
        Command::CostReport {} => {
            // cost_report checks protocol_version itself (one check, not two).
            moai_cli::run::cost_report(input::<moai_cli::run::CostRequest>()?)
        }
        Command::ProjectSkillPlan { project } => {
            let mut request = input::<moai_cli::project_skills::ProjectSkillRequest>()?;
            request.project_root = project;
            serde_json::to_value(moai_cli::project_skills::project_skill_plan(&request)?)
                .map_err(|_| "encode failed".into())
        }
        Command::ProjectSkillApply { project } => {
            let mut request = input::<moai_cli::project_skills::ProjectSkillRequest>()?;
            request.project_root = project;
            serde_json::to_value(moai_cli::project_skills::project_skill_apply(&request)?)
                .map_err(|_| "encode failed".into())
        }
        Command::ProjectSkillRevert { project } => {
            let mut request = input::<moai_cli::project_skills::ProjectSkillRevertRequest>()?;
            request.project_root = project;
            serde_json::to_value(moai_cli::project_skills::project_skill_revert(&request)?)
                .map_err(|_| "encode failed".into())
        }
        Command::GithubPlan { project } => {
            serde_json::to_value(moai_cli::github_templates::github_plan(
                &project,
                input::<moai_cli::github_templates::GithubRequest>()?,
            )?)
            .map_err(|_| "encode failed".into())
        }
        Command::LabelsPlan { project } => moai_cli::forge::labels::labels_plan_cli(
            &project,
            input::<moai_cli::forge::labels::LabelsRequest>()?,
            &moai_cli::forge::transport::GhTransport::from_env()?,
            &moai_cli::forge::policy::SystemClock,
        ),
        Command::LabelsApply {
            project,
            plan_digest,
            permission_reference,
        } => {
            if !moai_cli::contract::permission_reference_ok(&permission_reference) {
                return Err("permission reference required".into());
            }
            moai_cli::forge::labels::labels_apply_cli(
                &project,
                plan_input::<serde_json::Value>()?,
                &plan_digest,
                &permission_reference,
                &moai_cli::forge::transport::GhTransport::from_env()?,
                &moai_cli::forge::policy::SystemClock,
            )
        }
        Command::GithubApply {
            project,
            plan_digest,
            permission_reference,
        } => moai_cli::github_templates::github_apply(
            &project,
            plan_input::<moai_cli::github_templates::GithubPlan>()?,
            &plan_digest,
            &permission_reference,
        ),
        Command::GenerationPlan { project } => {
            serde_json::to_value(moai_cli::generation::generation_plan(
                &project,
                input::<moai_cli::generation::GenerationRequest>()?,
            )?)
            .map_err(|_| "encode failed".into())
        }
        Command::GenerationApply {
            project,
            plan_digest,
            permission_reference,
        } => moai_cli::generation::generation_apply(
            &project,
            plan_input::<moai_cli::generation::GenerationPlan>()?,
            &plan_digest,
            &permission_reference,
        ),
        Command::ContextBundle { project } => moai_cli::context::bundle(&project, &input()?),
        Command::ContextVerify { project } => moai_cli::context::verify(&project, &input()?),
        Command::DecisionRecord {
            project,
            run,
            permission_reference,
        } => moai_cli::decision::append(&project, &run, &input()?, &permission_reference),
        Command::DecisionAppend {
            project,
            log,
            apply,
            permission_reference,
        } => moai_cli::decision_log::append(
            &project,
            &log,
            &input()?,
            apply,
            permission_reference.as_deref(),
        ),
        Command::AuditClaim => moai_cli::audit_claim::reserve_cli(
            input::<moai_cli::audit_claim::AuditClaimRequest>()?,
            &moai_cli::forge::transport::GhTransport::from_env()?,
            &moai_cli::forge::policy::SystemClock,
        ),
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

/// The envelope `operation` of a command line whose first word names no command (#2799 D3
/// decision 1 (a)); no operation ran, so it names the command line itself.
const UNKNOWN_COMMAND_OPERATION: &str = "cli";
const UNKNOWN_COMMAND: &str = "unknown command; see --help";

/// True when clap refused `first` as a subcommand and `first` is a bare word that names no
/// command (nor `help`): `moai no-such-command`. An option, `help <word>` or a known command
/// followed by a wrong word is not this case and keeps its own handling.
fn names_no_command(e: &clap::Error, first: &str) -> bool {
    matches!(e.kind(), ErrorKind::InvalidSubcommand)
        && !first.starts_with('-')
        && first != "help"
        && Cli::command().find_subcommand(first).is_none()
}

/// A command-line error on an internal operation becomes a `usage` envelope (exit 2) with a
/// message that echoes no argument; stderr keeps clap's first line for people. Help, version,
/// the launchers and words that name no operation keep clap's own output.
fn usage_error(e: clap::Error) -> ExitCode {
    let first = std::env::args().nth(1);
    if first.as_deref().is_some_and(|w| names_no_command(&e, w)) {
        // #2799 D3 decision 1 (a): an unknown first word is a usage error like any other, so
        // stdout carries the `usage` envelope too. Its message echoes no argument (the word
        // may be anything pasted); stderr and the exit code stay clap's own, byte for byte.
        let _ = e.print();
        println!(
            "{}",
            contract::failure(
                UNKNOWN_COMMAND_OPERATION,
                contract::Kind::Usage,
                UNKNOWN_COMMAND
            )
        );
        return ExitCode::from(2);
    }
    let op = first.and_then(|w| operation_of(&w));
    let Some(op) =
        op.filter(|_| !matches!(e.kind(), ErrorKind::DisplayHelp | ErrorKind::DisplayVersion))
    else {
        // Root help carries the logo above clap's help body (Issue #2363);
        // subcommand help and version stay logo-free.
        if matches!(e.kind(), ErrorKind::DisplayHelp) {
            let rest: Vec<String> = std::env::args().skip(1).collect();
            moai_cli::banner::maybe_print_logo_for(&rest);
        }
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
    // A key command's stray argument may be a key typed in the wrong place (#2393 S1 B2), so
    // its stderr carries only the fixed text, never clap's line that quotes the argument.
    if op.starts_with("key.") {
        eprintln!("{message}");
    } else {
        let rendered = e.render().to_string();
        eprintln!("{}", rendered.lines().next().unwrap_or(message));
    }
    println!("{}", contract::failure(op, contract::Kind::Usage, message));
    ExitCode::from(2)
}

/// True when the command line is a generated host hook (#2348): `hook-receipt` or any word
/// run with `--hook`. Read from the raw arguments so a command line that fails to parse is
/// still handled as a hook. Both hosts read a hook's stdout as hook output and its exit 2 as
/// "block", so in hook mode stdout stays empty and only the guard's refusal exits 2.
fn hook_mode_argv(args: &[String]) -> bool {
    args.get(1)
        .is_some_and(|w| w == "hook-receipt" || w == "hook-dispatch")
        || args.iter().skip(1).any(|a| a == "--hook")
}

/// `hook-dispatch` (#2391): prints exactly the host's native answer — no contract envelope —
/// and exits with the engine's code (0 allow/deny/context, 1 engine error).
fn hook_dispatch_main(host: &str, event: &str) -> ExitCode {
    let mut bytes = Vec::new();
    let _ = io::stdin()
        .take(moai_cli::hook_engine::STDIN_LIMIT)
        .read_to_end(&mut bytes);
    let answer = match std::env::current_dir() {
        Ok(cwd) => {
            let now = std::time::SystemTime::now()
                .duration_since(std::time::UNIX_EPOCH)
                .map_or(0, |d| d.as_millis());
            moai_cli::hook_engine::dispatch(host, event, &bytes, &cwd, now)
        }
        Err(_) => moai_cli::hook_engine::Answer {
            stdout: String::new(),
            stderr: "moai-cli hook-dispatch: engine error: working directory unavailable".into(),
            exit: moai_cli::hook_engine::ENGINE_ERROR_EXIT,
        },
    };
    if !answer.stdout.is_empty() {
        println!("{}", answer.stdout);
    }
    if !answer.stderr.is_empty() {
        eprintln!("{}", answer.stderr);
    }
    ExitCode::from(answer.exit)
}

/// Adds whether the generated hook command resolves on this process's PATH (#2348) to a
/// hooks-plan / hooks-apply result.
fn merge_binary_state(out: &mut serde_json::Value) {
    let path_env = std::env::var("PATH").unwrap_or_default();
    if let (Some(out), serde_json::Value::Object(state)) = (
        out.as_object_mut(),
        moai_cli::hooks::binary_state(&path_env),
    ) {
        out.extend(state);
    }
}

/// Exit code of a hook-mode error that is not the guard's refusal: 1, which both hosts treat
/// as a non-blocking hook error; the reason goes to stderr.
const HOOK_NON_BLOCKING_ERROR: u8 = 1;

/// `pr-self-audit-check` (#2708): the Python checker's text and exit codes. With no `--file`,
/// the body comes from stdin; a terminal on stdin is a usage error, as before.
fn pr_self_audit_check(reaudit: bool, file: Option<&std::path::Path>) -> ExitCode {
    use std::io::IsTerminal as _;
    let body = match file {
        Some(path) => match std::fs::read(path) {
            Ok(bytes) => String::from_utf8(bytes).map_err(|_| "body is not UTF-8".to_string()),
            Err(e) => Err(format!("{}: {e}", path.display())),
        },
        None if io::stdin().is_terminal() => {
            eprintln!("usage: moai-cli pr-self-audit-check [--reaudit] [--file BODY.md] < BODY.md");
            return ExitCode::from(2);
        }
        None => {
            let mut bytes = Vec::new();
            io::stdin()
                .read_to_end(&mut bytes)
                .map_err(|e| e.to_string())
                .and_then(|_| String::from_utf8(bytes).map_err(|_| "body is not UTF-8".into()))
        }
    };
    let body = match body {
        Ok(body) => body,
        Err(e) => {
            eprintln!("pr-self-audit-check: {e}");
            return ExitCode::from(1);
        }
    };
    let result = moai_cli::self_audit::check(&body, reaudit);
    println!("{}", moai_cli::self_audit::render(&result));
    if result.is_ok() {
        ExitCode::SUCCESS
    } else {
        ExitCode::from(1)
    }
}

/// The first word of a Go-generation hook call (#2790 R0): the Go `moai` ran project hooks as
/// `moai hook <event> [args]` (34 events in the Go templates' `.claude/hooks/moai/*.sh`, which
/// look for `~/go/bin/moai` before PATH).
const GO_HOOK_WORD: &str = "hook";

/// `moai hook <event> …` from a Go-generation project hook (#2790 R0, impact table C1). This
/// build has no `hook` command; answering with clap's usage error would exit 2, which hosts
/// read as "block this tool call". It is a hook call, so it gets hook mode's quiet answer:
/// the host's input on stdin is read and dropped (bounded, never from a terminal), stdout
/// stays empty, one line on stderr says what happened, and the exit is 0 — nothing blocked.
fn go_hook_main(argv: &[String]) -> ExitCode {
    use std::io::IsTerminal as _;
    if !io::stdin().is_terminal() {
        let _ = io::copy(
            &mut io::stdin().take(moai_cli::hook_engine::STDIN_LIMIT),
            &mut io::sink(),
        );
    }
    let event = argv
        .get(2)
        .map(String::as_str)
        .filter(|e| {
            !e.is_empty() && e.len() <= 64 && e.bytes().all(|b| b.is_ascii_lowercase() || b == b'-')
        })
        .unwrap_or("?");
    eprintln!(
        "moai: ignored a Go-generation hook call (hook {event}); this Rust build does not run Go hooks, nothing was blocked"
    );
    ExitCode::SUCCESS
}

/// Entry point of both binaries: `moai-cli` (this file) and `moai` (`src/bin/moai.rs` loads
/// this file as a module and calls this function; #2790 R1).
pub(crate) fn main() -> ExitCode {
    let argv: Vec<String> = std::env::args().collect();
    if argv.get(1).is_some_and(|w| w == GO_HOOK_WORD) {
        return go_hook_main(&argv);
    }
    let hook_mode = hook_mode_argv(&argv);
    let cli = match Cli::try_parse() {
        Ok(cli) => cli,
        Err(e) if hook_mode => {
            if matches!(e.kind(), ErrorKind::DisplayHelp | ErrorKind::DisplayVersion) {
                e.exit()
            }
            let rendered = e.render().to_string();
            eprintln!(
                "moai-cli hook: {}",
                rendered.lines().next().unwrap_or("invalid command line")
            );
            return ExitCode::from(HOOK_NON_BLOCKING_ERROR);
        }
        Err(e) => return usage_error(e),
    };
    if let Some(Command::Statusline {
        refresh_counts: true,
        board_root: Some(root),
        ..
    }) = &cli.command
    {
        moai_cli::statusline::counts::refresh(root);
        return ExitCode::SUCCESS;
    }
    if let Some(Command::Statusline {
        render_budget_ms,
        now_unix,
        ..
    }) = cli.command
    {
        let budget = render_budget_ms
            .map(std::time::Duration::from_millis)
            .unwrap_or(moai_cli::statusline::RENDER_BUDGET);
        print!("{}", moai_cli::statusline::run(budget, now_unix));
        return ExitCode::SUCCESS;
    }
    if let Some(Command::PrSelfAuditCheck { reaudit, file }) = &cli.command {
        return pr_self_audit_check(*reaudit, file.as_deref());
    }
    let op = cli.command.as_ref().and_then(operation);
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
    if let Some(Command::HookDispatch { host, event }) = &cli.command {
        return hook_dispatch_main(host, event);
    }
    if let Some(Command::HeavySlotAcquire {
        max,
        name,
        dir,
        min_free_gb,
        max_swap_pct,
        command,
    }) = &cli.command
    {
        // No contract envelope: the wrapped command's own exit code and output (#2677).
        let dir = match dir
            .clone()
            .map(Ok)
            .unwrap_or_else(moai_cli::heavy_slot::default_dir)
        {
            Ok(d) => d,
            Err(e) => {
                eprintln!("heavy-slot: {e}");
                return ExitCode::from(moai_cli::heavy_slot::USAGE);
            }
        };
        let limits = moai_cli::heavy_slot::Limits {
            min_free_gb: *min_free_gb,
            max_swap_pct: *max_swap_pct,
        };
        return ExitCode::from(moai_cli::heavy_slot::acquire(
            &dir, *max, name, command, limits,
        ));
    }
    // A host hook blocks a tool call only on exit 2 (Claude Code and Codex hooks docs); the
    // guard's refusal is the one error that may block. A receipt hook never blocks.
    let guard = matches!(&cli.command, Some(Command::ClaimVerify { hook: true, .. }));
    match run_internal(cli.command.expect("internal operation")) {
        Ok(_) if hook_mode => ExitCode::SUCCESS,
        Ok(result) => {
            // A computed delta is evidence, not audit admission; only a pure replay exits 0.
            let exit = if op == "audit.delta" && result["status"] != "pure_merge" {
                ExitCode::from(1)
            } else {
                ExitCode::SUCCESS
            };
            println!("{}", contract::success(op, result));
            exit
        }
        Err(e) if hook_mode => {
            eprintln!("{e}");
            ExitCode::from(if guard { 2 } else { HOOK_NON_BLOCKING_ERROR })
        }
        Err(e) => {
            // #2345: the journal lock-busy error carries its kind from the raise site;
            // every other error was classified at the dispatch boundary.
            println!("{}", contract::failure(op, e.kind, &e.message));
            eprintln!("{e}");
            ExitCode::from(1)
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    const LAUNCHERS: [&str; 4] = ["cc", "glm", "codex", "session"];
    /// Raw-text subcommands outside the envelope (#2350); documented in interfaces.md.
    const RAW_OUTPUT: [&str; 2] = ["statusline", "pr-self-audit-check"];

    /// #2722 AC-5: `--version` shows the build version, which is the package version plus
    /// `+local.<n>` exactly when this build set `MOAI_LOCAL_BUILD`.
    #[test]
    fn version_flag_reports_the_local_build_number() {
        let expected = moai_cli::local_build::version(
            env!("CARGO_PKG_VERSION"),
            option_env!("MOAI_LOCAL_BUILD"),
        )
        .unwrap();
        assert_eq!(moai_cli::BUILD_VERSION, expected);
        // `--version` prints the build version and the generation mark (#2790 R1).
        let signature = format!("{} (rust)", moai_cli::BUILD_VERSION);
        assert_eq!(Cli::command().get_version(), Some(signature.as_str()));
        assert_eq!(moai_cli::VERSION_SIGNATURE, signature);
        if option_env!("MOAI_LOCAL_BUILD").is_none_or(str::is_empty) {
            assert_eq!(moai_cli::BUILD_VERSION, env!("CARGO_PKG_VERSION"));
        }
    }

    /// Every clap subcommand word names exactly one operation through `operation` (#2254): a
    /// word whose required arguments cannot be filled, or two words with one name, fail here.
    /// `MOAI_DUMP_OPERATIONS=1` prints the `word operation` pairs (the before/after diff).
    #[test]
    fn every_subcommand_word_maps_to_one_operation() {
        let mut seen = std::collections::BTreeMap::new();
        let cli = Cli::command();
        for sub in cli.get_subcommands() {
            let word = sub.get_name();
            if word == "help" {
                continue;
            }
            let op = operation_of(word);
            if LAUNCHERS.contains(&word) {
                assert_eq!(op, None, "{word} is a launcher");
                continue;
            }
            if RAW_OUTPUT.contains(&word) {
                assert_eq!(op, None, "{word} prints raw text, not the envelope");
                continue;
            }
            let op = op.unwrap_or_else(|| panic!("{word} maps to no operation"));
            if let Some(other) = seen.insert(op, word) {
                panic!("{word} and {other} share {op}");
            }
        }
        // Every non-launcher word was counted once: no hand-kept total to bump.
        let words = cli
            .get_subcommands()
            .filter(|s| s.get_name() != "help")
            .count();
        assert_eq!(seen.len() + LAUNCHERS.len() + RAW_OUTPUT.len(), words);
        if std::env::var_os("MOAI_DUMP_OPERATIONS").is_some() {
            let mut pairs: Vec<_> = seen.iter().map(|(op, w)| format!("{w} {op}")).collect();
            pairs.sort();
            println!("{}", pairs.join("\n"));
        }
    }

    #[test]
    fn a_word_that_is_no_subcommand_has_no_operation() {
        for word in ["", "nope", "package", "package.inspect", "--help"] {
            assert_eq!(operation_of(word), None, "{word:?}");
        }
    }
}

#[cfg(test)]
mod operations_doc {
    //! rc.2 D7 #2246: `docs/design-review/appendix/interfaces.md` § "RC1 OPERATIONS 전수" must list
    //! exactly the internal subcommand words mapped by `operation`, so a PR that adds or removes an operation
    //! fails here unless it updates the document in the same change.
    //!
    //! The list is one `- `word`` line per operation in alphabetical order, with no count
    //! (#2241 repair, leader rule ③): PRs that each add an operation then touch different lines
    //! instead of all rewriting one shared count and one shared line.
    use super::{operation_of, Cli};
    use clap::CommandFactory;
    use std::collections::BTreeSet;

    const DOC: &str = include_str!("../docs/design-review/appendix/interfaces.md");
    const HEADING: &str = "## RC1 OPERATIONS 전수";

    /// The words of the item lines (`- `word``) that directly follow the heading.
    fn doc_list(doc: &str) -> Result<Vec<String>, String> {
        let mut lines = doc.lines();
        lines
            .by_ref()
            .find(|l| l.trim_end() == HEADING)
            .ok_or_else(|| format!("heading `{HEADING}` not found"))?;
        let mut words = Vec::new();
        for line in lines.skip_while(|l| l.trim().is_empty()) {
            if !line.starts_with("- ") {
                break;
            }
            let word = line
                .strip_prefix("- `")
                .and_then(|rest| rest.strip_suffix('`'))
                .filter(|w| !w.is_empty() && !w.contains('`'))
                .ok_or_else(|| format!("malformed list line (expected - `word`): {line:?}"))?;
            words.push(word.to_string());
        }
        if words.is_empty() {
            return Err("operation list missing under the heading".into());
        }
        Ok(words)
    }

    fn duplicates<'a>(items: impl IntoIterator<Item = &'a str>) -> BTreeSet<&'a str> {
        let mut seen = BTreeSet::new();
        items.into_iter().filter(|w| !seen.insert(*w)).collect()
    }

    fn compare(code: &[&str], doc: &str) -> Result<(), String> {
        let words = doc_list(doc)?;
        let mut problems = Vec::new();
        let doc_dups = duplicates(words.iter().map(String::as_str));
        if !doc_dups.is_empty() {
            problems.push(format!("duplicate in document: {doc_dups:?}"));
        }
        let code_dups = duplicates(code.iter().copied());
        if !code_dups.is_empty() {
            problems.push(format!("duplicate in OPERATIONS: {code_dups:?}"));
        }
        if let Some(pair) = words.windows(2).find(|p| p[0] >= p[1]) {
            if pair[0] != pair[1] {
                problems.push(format!(
                    "document list not alphabetical: {:?} before {:?}",
                    pair[0], pair[1]
                ));
            }
        }
        let doc_set: BTreeSet<&str> = words.iter().map(String::as_str).collect();
        let code_set: BTreeSet<&str> = code.iter().copied().collect();
        let missing: Vec<_> = code_set.difference(&doc_set).collect();
        let extra: Vec<_> = doc_set.difference(&code_set).collect();
        if !missing.is_empty() {
            problems.push(format!("missing from document: {missing:?}"));
        }
        if !extra.is_empty() {
            problems.push(format!("not in OPERATIONS: {extra:?}"));
        }
        if problems.is_empty() {
            Ok(())
        } else {
            Err(problems.join("; "))
        }
    }

    fn code_words(cli: &clap::Command) -> Vec<&str> {
        cli.get_subcommands()
            .filter_map(|sub| operation_of(sub.get_name()).map(|_| sub.get_name()))
            .collect()
    }

    /// #2617: the one-file operations are documented by `operations/<word>.md` fragments, not
    /// by list lines, so two branches that each add one touch no shared line.
    const FRAGMENT_DIR: &str = concat!(
        env!("CARGO_MANIFEST_DIR"),
        "/docs/design-review/appendix/operations"
    );

    /// Subcommand words of the operations collected from `src/ops/`.
    fn file_op_words() -> Vec<String> {
        use clap::Subcommand;
        super::ops::Ops::augment_subcommands(clap::Command::new("ops"))
            .get_subcommands()
            .map(|s| s.get_name().to_string())
            .collect()
    }

    /// `(subcommand word, first line)` of every `*.md` fragment, the word being the file stem
    /// with `_` as `-`.
    fn fragments(dir: &str) -> Result<Vec<(String, String)>, String> {
        let mut out = Vec::new();
        for entry in std::fs::read_dir(dir).map_err(|e| format!("{dir}: {e}"))? {
            let path = entry.map_err(|e| e.to_string())?.path();
            let name = path.file_name().unwrap_or_default().to_string_lossy();
            let Some(stem) = name.strip_suffix(".md") else {
                return Err(format!("not a fragment file: {name}"));
            };
            let text = std::fs::read_to_string(&path).map_err(|e| format!("{name}: {e}"))?;
            let first = text.lines().next().unwrap_or_default().to_string();
            out.push((stem.replace('_', "-"), first));
        }
        out.sort();
        Ok(out)
    }

    fn compare_fragments(
        file_ops: &[String],
        fragments: &[(String, String)],
        list: &[String],
    ) -> Result<(), String> {
        let mut problems = Vec::new();
        let ops: BTreeSet<&str> = file_ops.iter().map(String::as_str).collect();
        let docs: BTreeSet<&str> = fragments.iter().map(|(w, _)| w.as_str()).collect();
        let missing: Vec<_> = ops.difference(&docs).collect();
        let extra: Vec<_> = docs.difference(&ops).collect();
        if !missing.is_empty() {
            problems.push(format!("operation file without a fragment: {missing:?}"));
        }
        if !extra.is_empty() {
            problems.push(format!("fragment without an operation file: {extra:?}"));
        }
        for (word, first) in fragments {
            if *first != format!("# `{word}`") {
                problems.push(format!("fragment {word}: first line must be # `{word}`"));
            }
        }
        let listed: Vec<_> = list.iter().filter(|w| ops.contains(w.as_str())).collect();
        if !listed.is_empty() {
            problems.push(format!("one-file operation also on the list: {listed:?}"));
        }
        if problems.is_empty() {
            Ok(())
        } else {
            Err(problems.join("; "))
        }
    }

    #[test]
    fn operations_doc_matches_code() {
        let cli = Cli::command();
        let file_ops = file_op_words();
        let listed: Vec<&str> = code_words(&cli)
            .into_iter()
            .filter(|w| !file_ops.iter().any(|f| f == w))
            .collect();
        if let Err(problems) = compare(&listed, DOC) {
            panic!("interfaces.md § RC1 OPERATIONS 전수 is out of sync with src/main.rs OPERATIONS: {problems}");
        }
        let frags = fragments(FRAGMENT_DIR).unwrap();
        if let Err(problems) = compare_fragments(&file_ops, &frags, &doc_list(DOC).unwrap()) {
            panic!("interfaces.md operations/ fragments are out of sync with src/ops: {problems}");
        }
    }

    /// #2617: every one-file operation is a real operation, and the fragment check names what
    /// is wrong.
    #[test]
    fn fragment_mutations_are_detected() {
        let file_ops = file_op_words();
        assert!(
            file_ops.iter().any(|w| w == "package-inspect"),
            "{file_ops:?}"
        );
        for w in &file_ops {
            assert!(operation_of(w).is_some(), "{w} maps to no operation");
        }
        let frags = fragments(FRAGMENT_DIR).unwrap();
        let list = doc_list(DOC).unwrap();
        assert_eq!(compare_fragments(&file_ops, &frags, &list), Ok(()));
        let mut more = file_ops.clone();
        more.push("zz-new-op".into());
        assert!(compare_fragments(&more, &frags, &list)
            .unwrap_err()
            .contains("operation file without a fragment: [\"zz-new-op\"]"));
        let mut ghost = frags.clone();
        ghost.push(("zz-ghost".into(), "# `zz-ghost`".into()));
        assert!(compare_fragments(&file_ops, &ghost, &list)
            .unwrap_err()
            .contains("fragment without an operation file: [\"zz-ghost\"]"));
        let mut bad = frags.clone();
        bad[0].1 = "# wrong".into();
        assert!(compare_fragments(&file_ops, &bad, &list)
            .unwrap_err()
            .contains("first line must be"));
        let mut both = list.clone();
        both.push(file_ops[0].clone());
        assert!(compare_fragments(&file_ops, &frags, &both)
            .unwrap_err()
            .contains("also on the list"));
    }

    /// #2350: `statusline` prints raw text, so it is a subcommand outside OPERATIONS and the
    /// document names it as the one envelope exception.
    #[test]
    fn statusline_is_the_documented_envelope_exception() {
        let cli = Cli::command();
        assert!(cli.get_subcommands().any(|s| s.get_name() == "statusline"));
        assert_eq!(operation_of("statusline"), None);
        assert!(!doc_list(DOC).unwrap().iter().any(|w| w == "statusline"));
        assert!(DOC.contains("envelope 예외 하위 명령: `statusline`(#2350)"));
        // The render-budget test seam is documented and hidden from help.
        assert!(DOC.contains("`--render-budget-ms`"));
        let sub = cli.find_subcommand("statusline").unwrap();
        for name in ["render-budget-ms", "now-unix"] {
            assert!(DOC.contains(&format!("`--{name}`")), "{name}");
            let seam = sub
                .get_arguments()
                .find(|a| a.get_long() == Some(name))
                .unwrap();
            assert!(seam.is_hide_set(), "{name}");
        }
    }

    #[test]
    fn pr_self_audit_check_is_the_documented_envelope_exception() {
        let cli = Cli::command();
        assert!(cli
            .get_subcommands()
            .any(|s| s.get_name() == "pr-self-audit-check"));
        assert_eq!(operation_of("pr-self-audit-check"), None);
        assert!(!doc_list(DOC)
            .unwrap()
            .iter()
            .any(|w| w == "pr-self-audit-check"));
        assert!(DOC.contains("envelope 예외 하위 명령: `pr-self-audit-check`(#2708)"));
    }

    fn render(words: &[&str]) -> String {
        let list: String = words.iter().map(|w| format!("- `{w}`\n")).collect();
        format!("{HEADING}\n\n{list}\n(note)\n")
    }

    fn err(code: &[&str], doc: &str) -> String {
        compare(code, doc).expect_err("mutation must be detected")
    }

    #[test]
    fn mutations_are_detected() {
        let cli = Cli::command();
        let code = code_words(&cli);
        let mut sorted = code.clone();
        sorted.sort_unstable();
        assert_eq!(compare(&code, &render(&sorted)), Ok(()));

        // A document line removed: reported by name.
        let first = sorted[0];
        let shorter: Vec<&str> = sorted[1..].to_vec();
        assert!(
            err(&code, &render(&shorter)).contains(&format!("missing from document: [{first:?}]"))
        );
        // An operation added to the code but not to the document.
        let mut more = code.clone();
        more.push("zz-new-op");
        assert!(err(&more, &render(&sorted)).contains("missing from document: [\"zz-new-op\"]"));
        // A document line with no operation behind it.
        let mut extra = sorted.clone();
        extra.push("zz-ghost-op");
        assert!(err(&code, &render(&extra)).contains("not in OPERATIONS: [\"zz-ghost-op\"]"));
        // A renamed line: one missing and one extra, both named.
        let mut renamed = sorted.clone();
        renamed[0] = "aaa-renamed-op";
        let e = err(&code, &render(&renamed));
        assert!(e.contains(first) && e.contains("aaa-renamed-op"), "{e}");
        // A duplicated line, on either side.
        let mut doubled = sorted.clone();
        doubled.insert(1, first);
        assert!(err(&code, &render(&doubled))
            .contains(&format!("duplicate in document: {{{first:?}}}")));
        let mut twice = code.clone();
        twice.push(code[0]);
        assert!(err(&twice, &render(&sorted)).contains("duplicate in OPERATIONS"));
        // Out of alphabetical order.
        let mut unsorted = sorted.clone();
        unsorted.swap(0, 1);
        assert!(err(&code, &render(&unsorted)).contains("not alphabetical"));
        // A malformed line, an empty list, a missing heading.
        let malformed = render(&sorted).replacen(&format!("- `{first}`"), &format!("- {first}"), 1);
        assert!(err(&code, &malformed).contains("malformed list line"));
        assert!(err(&code, &format!("{HEADING}\n\n(note)\n")).contains("list missing"));
        assert!(
            err(&code, &render(&sorted).replacen(HEADING, "## Other", 1)).contains("not found")
        );
    }
}
