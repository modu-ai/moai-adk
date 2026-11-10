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
    /// Session name (default `leader` for -l, the project's next `agent-N` for -a; none without a role).
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
    /// Claim a task (06-A #2005): typed JSON request on stdin; posts exactly one claim
    /// comment (⑤) behind the same-host journal serialization (①) and records the record.
    ClaimAcquire {
        project: PathBuf,
    },
    /// Self-claim the next card (rc.2 A2 #2161): typed JSON request on stdin; re-reads the
    /// open Issues, skips epics, pull requests and open-blocked cards, and claims the first
    /// candidate through `claim-acquire` under one journal lock.
    QueueNext {
        project: PathBuf,
    },
    /// Leader supervision tick (rc.2 A3 #2162): typed JSON request on stdin; GET-only read
    /// of open PRs, reports only what changed since the last tick, drafts assignee hints.
    LeaderTick {
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
    /// Flow control (A13 #2279): the integration gate verdict, read-only — checks
    /// tip containment (behind_by 0), changed-since-audit, claim lines, required check
    /// runs' conclusions, head stability. Status posting and merging stay with the
    /// leader flow.
    FlowGateVerdict {
        /// GitHub owner.
        #[arg(long)]
        owner: String,
        /// GitHub repository name.
        #[arg(long)]
        repo: String,
        /// The pull request number to judge.
        #[arg(long)]
        pr: i64,
        /// The head the last PASS audit was measured on.
        #[arg(long = "audit-head")]
        audit_head: String,
        /// The base ref the tip is read from (e.g. v2026.11.0).
        #[arg(long = "base-ref", default_value = "v2026.11.0")]
        base_ref: String,
        /// Comma-separated required check-run names (#2335). The verdict reads these
        /// runs' conclusions; unrelated checks and the combined commit status are
        /// excluded from the judgment.
        #[arg(long = "required-checks", default_value = "rc-ci")]
        required_checks: String,
    },
    /// Read-only Claude Code check: whether a session started in PROJECT reads AGENTS.md.
    ClaudeDoctor {
        project: PathBuf,
    },
    /// Preview a three-way merge of the managed keys (`outputStyle`, `statusLine`, `permissions`) on stdin
    /// into PROJECT/.claude/settings.json (#2239); other keys are refused. Writes nothing.
    ClaudeSettingsPlan {
        project: PathBuf,
    },
    /// Apply a previewed settings merge: same stdin, the preview's digest and a permission
    /// reference. Writes the project settings file and its receipt only; home is refused.
    ClaudeSettingsApply {
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
        /// Worktree path; default `<primary>/.moai-cli/worktrees/<issue>-<logical_key>`.
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
    /// Plan external skill source candidates read-only (03-B #1996): typed request on
    /// stdin; judgment envelope on stdout; writes nothing.
    SkillSourcePlan {
        project: PathBuf,
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
    /// Install approved skill sources into the host official skill folders (03-B
    /// #1996): the exact plan on stdin, `--plan-digest` and `--permission-reference`.
    SkillSourceApply {
        project: PathBuf,
        #[arg(long)]
        plan_digest: String,
        #[arg(long)]
        permission_reference: String,
    },
    /// Record that every file of an installed skill was actually read (03-B #1996).
    SkillReadReceipt {
        project: PathBuf,
        /// Installed skill directory to walk and hash.
        #[arg(long)]
        dir: PathBuf,
        /// Installed skill name (Agent Skills spec shape).
        #[arg(long)]
        name: String,
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
    /// Temporary (05-D #2004; the name is fixed after P06): independent audit of PROJECT. Re-runs
    /// the approved verification itself and decides admission; receipts on stdin never count.
    AuditReduce {
        project: PathBuf,
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
    use moai_cli::session::{full_access, peek_agent_number, take_agent_number};
    // User decision 2026-11-10 (#2359, comment 6760976883): `-l|-a` stay. `-l` is `leader`,
    // `-a` is `agent-N` from the project's counter; both start with full access.
    let role = leader.map(|l| if l { "leader" } else { "agent" });
    // `-a` without `--name` is numbered; `--name` takes no number.
    let numbered = leader == Some(false) && o.name.is_none();
    let name = o
        .name
        .or_else(|| (leader == Some(true)).then(|| "leader".into()));
    let mut request = SessionRequest {
        role: role.map(str::to_string),
        host: host.into(),
        name,
        model: o.model,
        effort: o.effort,
        // Claude defaults (user decision 2026-11-08, HOLD-L1 = A, #2053): roleless `cc`
        // takes the launcher defaults (advisor fable, model, effort) too.
        advisor: (host != "codex" && !o.no_advisor && moai_cli::session::launcher_defaults(role))
            .then(|| "fable".into()),
        // A role session starts with full access unless `--permissions` is given; a roleless
        // one gets no permission flag (user decision 2026-11-10, #2359; replaces #2244).
        skip_permissions: full_access(role.is_some(), o.permissions),
        prompt,
        claude_bin: "claude".into(),
        codex_bin: "codex".into(),
    };
    let mut number_note = None;
    if numbered {
        let here = std::env::current_dir()
            .map_err(|_| "agent numbering: the current directory is unreadable; pass --name")?;
        let dir = moai_cli::session::agent_seq_dir(&here)?;
        let n = if o.dry_run {
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
            let n = take_agent_number(&here)?;
            number_note = Some(format!("agent number {n} taken (store {})", dir.display()));
            n
        };
        request.name = Some(format!("agent-{n}"));
    }
    let mut v = session_launch(&request, o.dry_run)?;
    if let (Some(note), Some(notes)) = (number_note, v["notes"].as_array_mut()) {
        notes.push(note.into());
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
        Cc(_) | Glm(_) | Codex(_) | Session { .. } => return None,
        AgentsApply { .. } => "agents.apply",
        AgentsPlan { .. } => "agents.plan",
        ApplyRecover { .. } => "apply.recover",
        ArchiveApply { .. } => "archive.apply",
        ArchivePlan { .. } => "archive.plan",
        ArchiveRestore { .. } => "archive.restore",
        ArchiveSearch { .. } => "archive.search",
        AtlasTagScan { .. } => "atlas_tag.scan",
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
        CodexDoctor { .. } => "codex.doctor",
        CodexProjectionApply { .. } => "codex.projection_apply",
        CodexProjectionPlan { .. } => "codex.projection_plan",
        ContextBundle { .. } => "context.bundle",
        ContextVerify { .. } => "context.verify",
        CostReport { .. } => "cost.report",
        DecisionRecord { .. } => "decision.record",
        EffectConfirm { .. } => "effect.confirm",
        EffectIntend { .. } => "effect.intend",
        EffectsReconcile { .. } => "effects.reconcile",
        FlowGateVerdict { .. } => "flow.gate_verdict",
        FlowWipCheck { .. } => "flow.wip_check",
        ForgeAuditStatus { .. } => "forge.audit_status",
        ForgeLinksProbe { .. } => "forge.links_probe",
        ForgeRead { .. } => "forge.read",
        ForgeReport { .. } => "forge.report",
        GenerationApply { .. } => "generation.apply",
        GenerationPlan { .. } => "generation.plan",
        GithubApply { .. } => "github.apply",
        GithubPlan { .. } => "github.plan",
        HookReceipt { .. } => "hook.receipt",
        HooksApply { .. } => "hooks.apply",
        HooksPlan { .. } => "hooks.plan",
        Identity { .. } => "identity",
        InstallInspect { .. } => "install.inspect",
        LaneHandoffCheck { .. } => "lane_handoff.check",
        LayoutApply { .. } => "layout.apply",
        LayoutPlan { .. } => "layout.plan",
        LeaderTick { .. } => "leader.tick",
        PackageInspect { .. } => "package.inspect",
        PackageResolve { .. } => "package.resolve",
        ParityCheck { .. } => "parity.check",
        PlanAdmit { .. } => "plan.admit",
        PrBodyRender { .. } => "pr.body.render",
        PrConfirm { .. } => "pr.confirm",
        ProjectApply { .. } => "project.apply",
        ProjectPlan { .. } => "project.plan",
        ProjectSkillApply { .. } => "project_skill.apply",
        ProjectSkillPlan { .. } => "project_skill.plan",
        ProjectSkillRevert { .. } => "project_skill.revert",
        QueueNext { .. } => "queue.next",
        RecoveryApply { .. } => "recovery.apply",
        RecoveryPlan { .. } => "recovery.plan",
        RunVerify { .. } => "run.verify",
        ScopeCheck { .. } => "scope.check",
        SkillReadReceipt { .. } => "skill_source.read_receipt",
        SkillSourceApply { .. } => "skill_source.apply",
        SkillSourcePlan { .. } => "skill_source.plan",
        SnapshotDescribe { .. } => "snapshot.describe",
        TodoPromoteApply { .. } => "todo.promote.apply",
        TodoPromotePreview { .. } => "todo.promote.plan",
        UpdateApply { .. } => "update.apply",
        UpdatePlan { .. } => "update.plan",
        WorktreeApply { .. } => "worktree.apply",
        WorktreePlan { .. } => "worktree.plan",
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
/// #2345: the eleven journal-writing operations raise the journal lock-busy kind at its
/// source and return it typed, so the envelope never re-derives it from message text;
/// every other operation keeps the classified `String` fallback.
fn run_internal(
    command: Command,
) -> std::result::Result<serde_json::Value, moai_cli::contract::OpError> {
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
        Command::WorktreeApply {
            project,
            source,
            branch_prefix,
            path,
            plan_digest,
        } => moai_cli::worktree::apply_and_record(
            &project,
            &source.load(&project)?,
            source.task.is_some(),
            &branch_prefix,
            path.as_deref(),
            &plan_digest,
            &moai_cli::forge::policy::SystemClock,
        ),
        Command::ScopeCheck {
            project,
            source,
            integration_order,
        } => {
            let mine = source.load(&project)?;
            let others = moai_cli::worktree::active_records(&project)?;
            moai_cli::scope::check_and_record(
                &project,
                &mine,
                source.task.is_some(),
                &others,
                integration_order.as_deref(),
                &moai_cli::forge::policy::SystemClock,
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
        Command::LayoutPlan { project } => moai_cli::layout::plan_cli(&project),
        Command::LayoutApply {
            project,
            plan_digest,
            permission_reference,
        } => moai_cli::layout::apply_cli(&project, &plan_digest, &permission_reference),
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
        Command::FlowGateVerdict {
            owner,
            repo,
            pr,
            audit_head,
            base_ref,
            required_checks,
        } => {
            let transport = moai_cli::forge::transport::GhTransport::from_env()?;
            let clock = moai_cli::forge::policy::SystemClock;
            let client = moai_cli::forge::Client::new(
                &transport,
                &clock,
                Default::default(),
                Default::default(),
            );
            let checks: Vec<String> = required_checks
                .split(',')
                .map(str::trim)
                .filter(|s| !s.is_empty())
                .map(str::to_string)
                .collect();
            moai_cli::flow::gate_verdict(
                &client,
                &owner,
                &repo,
                pr,
                &audit_head,
                &base_ref,
                &checks,
            )
        }
        Command::ClaudeDoctor { project } => moai_cli::claude::doctor(&project),
        Command::HooksPlan { project, host } => {
            let home = std::env::var_os("HOME").map(PathBuf::from);
            let mut out = moai_cli::hooks::plan(&project, &host, home.as_deref())?;
            merge_binary_state(&mut out);
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
        Command::ClaudeSettingsPlan { project } => {
            let desired = moai_cli::claude::user_settings_desired(&project, &input()?)?;
            moai_cli::claude::settings_plan(&project, &desired)
        }
        Command::ClaudeSettingsApply {
            project,
            plan_digest,
            permission_reference,
        } => {
            let desired = moai_cli::claude::user_settings_desired(&project, &input()?)?;
            moai_cli::claude::settings_apply(
                &project,
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
            moai_cli::install::inspect(&home, &path_env)
        }
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
        } => {
            let staged = moai_cli::update::stage_package(&stage)?;
            let plan = plan_input::<moai_cli::update::UpdatePlan>()?;
            moai_cli::update::apply_update(
                &home,
                &staged,
                &plan,
                &plan_digest,
                &permission_reference,
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
            serde_json::to_value(moai_cli::run::run(&project, &request, &current)?)
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
        Command::SkillSourcePlan { project } => serde_json::to_value(moai_cli::sources::plan(
            &project,
            input::<moai_cli::sources::SourcesRequest>()?,
        )?)
        .map_err(|_| "encode failed".into()),
        Command::SkillSourceApply {
            project,
            plan_digest,
            permission_reference,
        } => {
            let home = std::env::var_os("HOME")
                .map(PathBuf::from)
                .ok_or("home directory missing; set HOME")?;
            moai_cli::sources::sources_apply(
                &project,
                &home,
                plan_input::<moai_cli::sources::SourcePlan>()?,
                &plan_digest,
                &permission_reference,
            )
        }
        Command::SkillReadReceipt { project, dir, name } => {
            moai_cli::sources::read_receipt(&project, &dir, &name)
        }
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

/// True when the command line is a generated host hook (#2348): `hook-receipt` or any word
/// run with `--hook`. Read from the raw arguments so a command line that fails to parse is
/// still handled as a hook. Both hosts read a hook's stdout as hook output and its exit 2 as
/// "block", so in hook mode stdout stays empty and only the guard's refusal exits 2.
fn hook_mode_argv(args: &[String]) -> bool {
    args.get(1).is_some_and(|w| w == "hook-receipt") || args.iter().skip(1).any(|a| a == "--hook")
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

fn main() -> ExitCode {
    let argv: Vec<String> = std::env::args().collect();
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
    // A host hook blocks a tool call only on exit 2 (Claude Code and Codex hooks docs); the
    // guard's refusal is the one error that may block. A receipt hook never blocks.
    let guard = matches!(&cli.command, Some(Command::ClaimVerify { hook: true, .. }));
    match run_internal(cli.command.expect("internal operation")) {
        Ok(_) if hook_mode => ExitCode::SUCCESS,
        Ok(result) => {
            println!("{}", contract::success(op, result));
            ExitCode::SUCCESS
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
        assert_eq!(seen.len() + LAUNCHERS.len(), words);
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

    #[test]
    fn operations_doc_matches_code() {
        let cli = Cli::command();
        if let Err(problems) = compare(&code_words(&cli), DOC) {
            panic!("interfaces.md § RC1 OPERATIONS 전수 is out of sync with src/main.rs OPERATIONS: {problems}");
        }
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
