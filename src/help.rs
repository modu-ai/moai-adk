//! The people-facing help layout (#2802 S1; design #2799 D2 section 9, leader D3 6865936070).
//!
//! `moai --help` lists the commands a person types, in five groups; `moai help --all` lists
//! every command in eight groups. The JSON envelope on stdout is not touched by anything here:
//! help is text for people. clap has no way to put subcommands under several headings, so the
//! root help body is drawn from [`GROUPS`] and handed to clap as the root's help template.

use clap::Command;
use unicode_width::UnicodeWidthStr;

/// One help group: its heading and its commands, each with whether the root help shows it.
pub struct Group {
    pub heading: &'static str,
    pub commands: &'static [(&'static str, bool)],
}

/// Every command word in exactly one group (pinned by `help_table_names_every_command_once`).
/// `true` marks the commands the root help shows (#2799 B 6865555998, root draft 6865796537);
/// the rest appear only in `help --all`.
pub const GROUPS: &[Group] = &[
    Group {
        heading: "Start",
        commands: &[("cc", true), ("codex", true), ("glm", true)],
    },
    Group {
        heading: "Project",
        commands: &[
            ("project-detect", true),
            ("project-plan", true),
            ("project-apply", true),
            ("language-setup", true),
            ("index-build", true),
            ("agents-plan", false),
            ("agents-apply", false),
            ("generation-plan", false),
            ("generation-apply", false),
            ("github-plan", false),
            ("github-apply", false),
            ("labels-plan", false),
            ("labels-apply", false),
            ("claude-settings-plan", false),
            ("claude-settings-apply", false),
            ("codex-config-plan", false),
            ("codex-config-apply", false),
            ("hooks-plan", false),
            ("hooks-apply", false),
        ],
    },
    Group {
        heading: "Update and move",
        commands: &[
            ("update", true),
            ("go-state-plan", true),
            ("go-state-apply", true),
            ("go-state-revert-plan", true),
            ("go-state-revert", true),
            ("home-state-plan", true),
            ("home-state-apply", true),
            ("home-state-revert", true),
            ("worktree-move-plan", true),
            ("worktree-move-apply", true),
            ("update-plan", false),
            ("update-apply", false),
            ("recovery-plan", false),
            ("recovery-apply", false),
            ("apply-recover", false),
            ("layout-plan", false),
            ("layout-apply", false),
            ("migrate-plan", false),
            ("migrate-apply", false),
            ("migrate-recovery-plan", false),
            ("migrate-recover", false),
            ("skills-legacy", false),
            ("bin-names-plan", false),
            ("bin-names-apply", false),
            ("agents-home-plan", false),
            ("agents-home-apply", false),
            ("agents-home-revert", false),
            ("home-routing-plan", false),
            ("home-routing-apply", false),
            ("home-routing-revert", false),
        ],
    },
    Group {
        heading: "Keys",
        commands: &[
            ("key-set", true),
            ("key-status", true),
            ("key-remove", true),
            ("key-import-plan", true),
            ("key-import-apply", true),
        ],
    },
    Group {
        heading: "Diagnose",
        commands: &[
            ("claude-doctor", true),
            ("codex-doctor", true),
            ("install-inspect", true),
            ("identity", true),
            ("worktree-report", true),
            ("optimize-plan", true),
            ("baseline-inventory", true),
        ],
    },
    Group {
        heading: "Lane and leader",
        commands: &[
            ("claim-acquire", false),
            ("claim-verify", false),
            ("claim-show", false),
            ("claim-release", false),
            ("claim-handoff", false),
            ("queue-next", false),
            ("queue-free", false),
            ("queue-query", false),
            ("leader-tick", false),
            ("lane-status", false),
            ("lane-handoff-check", false),
            ("lane-safe-to-reset", false),
            ("session-register", false),
            ("usage-limit", false),
            ("heavy-slot-acquire", false),
            ("heavy-slot-status", false),
            ("goal-show", false),
            ("goal-plan", false),
            ("goal-apply", false),
            ("goal-block", false),
            ("gate-check", false),
            ("gate-merge", false),
            ("gate-loop", false),
            ("wait", false),
            ("scope-check", false),
            ("worktree-plan", false),
            ("worktree-apply", false),
            ("pr-body-render", false),
            ("pr-confirm", false),
            ("audit-claim", false),
            ("audit-clone-plan", false),
            ("labels-check", false),
            ("decision-due", false),
            ("decision-file-check", false),
            ("decision-append", false),
            ("flow-wip-check", false),
            ("dispatch-record", false),
            ("dispatch-ack", false),
            ("dispatch-check", false),
            ("message-record", false),
            ("message-ack", false),
            ("message-reconcile", false),
            ("effect-intend", false),
            ("effect-confirm", false),
            ("effects-reconcile", false),
        ],
    },
    Group {
        heading: "Checks",
        commands: &[
            ("review-package", false),
            ("review-check", false),
            ("repo-scans", false),
            ("pr-self-audit-check", false),
            ("flake-classify", false),
            ("flake-record", false),
            ("audit-delta", false),
            ("audit-reduce", false),
            ("parity-check", false),
            ("plan-admit", false),
            ("run-verify", false),
        ],
    },
    Group {
        heading: "Internal",
        commands: &[
            ("statusline", false),
            ("hook-receipt", false),
            ("hook-dispatch", false),
            ("forge-read", false),
            ("forge-links-probe", false),
            ("forge-report", false),
            ("forge-audit-status", false),
            ("index-query", false),
            ("atlas-map", false),
            ("atlas-graph", false),
            ("atlas-tag-scan", false),
            ("search", false),
            ("context-bundle", false),
            ("context-verify", false),
            ("decision-record", false),
            ("cost-report", false),
            ("snapshot-describe", false),
            ("package-inspect", false),
            ("package-resolve", false),
            ("codex-projection-plan", false),
            ("codex-projection-apply", false),
            ("project-skill-plan", false),
            ("project-skill-apply", false),
            ("project-skill-revert", false),
            ("archive-plan", false),
            ("archive-apply", false),
            ("archive-restore", false),
            ("archive-search", false),
            ("worktree-gc-plan", false),
            ("worktree-gc-apply", false),
            ("session", false),
        ],
    },
];

/// Examples per command (#2802 S1c): shown under `Examples:` in `<command> --help`, each
/// after the binary name. Every one parses with this binary's own command line (pinned by
/// `every_example_parses`); ` < file` marks standard input and is not part of the arguments.
/// An example carries ` < file` exactly when its command reads standard input (pinned by
/// `stdin_examples_match_the_commands_that_read_stdin`).
pub const EXAMPLES: &[(&str, &[&str])] = &[
    ("cc", &["cc -l", "cc -a"]),
    ("codex", &["codex -a"]),
    ("glm", &["glm -a"]),
    ("project-detect", &["project-detect ."]),
    ("project-plan", &["project-plan . < request.json", "project-apply --plan-digest <digest> --permission-reference <ref> . < request.json"]),
    ("project-apply", &["project-apply --plan-digest <digest> --permission-reference <ref> . < request.json"]),
    ("language-setup", &["language-setup . < request.json"]),
    ("index-build", &["index-build . < request.json"]),
    ("agents-plan", &["agents-plan . < request.json", "agents-apply --plan-digest <digest> --permission-reference <ref> . < request.json"]),
    ("agents-apply", &["agents-apply --plan-digest <digest> --permission-reference <ref> . < request.json"]),
    ("generation-plan", &["generation-plan . < request.json", "generation-apply --plan-digest <digest> --permission-reference <ref> . < request.json"]),
    ("generation-apply", &["generation-apply --plan-digest <digest> --permission-reference <ref> . < request.json"]),
    ("github-plan", &["github-plan . < request.json", "github-apply --plan-digest <digest> --permission-reference <ref> . < request.json"]),
    ("github-apply", &["github-apply --plan-digest <digest> --permission-reference <ref> . < request.json"]),
    ("labels-plan", &["labels-plan . < request.json", "labels-apply --plan-digest <digest> --permission-reference <ref> . < request.json"]),
    ("labels-apply", &["labels-apply --plan-digest <digest> --permission-reference <ref> . < request.json"]),
    ("claude-settings-plan", &["claude-settings-plan . < request.json", "claude-settings-apply --plan-digest <digest> --permission-reference <ref> . < request.json"]),
    ("claude-settings-apply", &["claude-settings-apply --plan-digest <digest> --permission-reference <ref> . < request.json"]),
    ("codex-config-plan", &["codex-config-plan . < request.json", "codex-config-apply --plan-digest <digest> --permission-reference <ref> . < request.json"]),
    ("codex-config-apply", &["codex-config-apply --plan-digest <digest> --permission-reference <ref> . < request.json"]),
    ("hooks-plan", &["hooks-plan --host claude .", "hooks-apply --host claude --plan-digest <digest> --permission-reference <ref> ."]),
    ("hooks-apply", &["hooks-apply --host claude --plan-digest <digest> --permission-reference <ref> ."]),
    ("update", &["update"]),
    ("go-state-plan", &["go-state-plan .", "go-state-apply --plan-digest <digest> --permission-reference <ref> ."]),
    ("go-state-apply", &["go-state-apply --plan-digest <digest> --permission-reference <ref> ."]),
    ("go-state-revert-plan", &["go-state-revert-plan --backup <UTC> ."]),
    ("go-state-revert", &["go-state-revert --backup <UTC> --plan-digest <digest> --permission-reference <ref> ."]),
    ("home-state-plan", &["home-state-plan", "home-state-apply --now 2026-11-14T00:00:00Z --plan-digest <digest> --permission-reference <ref>"]),
    ("home-state-apply", &["home-state-apply --now 2026-11-14T00:00:00Z --plan-digest <digest> --permission-reference <ref>"]),
    ("home-state-revert", &["home-state-revert --permission-reference <ref>"]),
    ("worktree-move-plan", &["worktree-move-plan .", "worktree-move-apply --plan-digest <digest> --permission-reference <ref> ."]),
    ("worktree-move-apply", &["worktree-move-apply --plan-digest <digest> --permission-reference <ref> ."]),
    ("update-plan", &["update-plan ~ <stage>", "update-apply --plan-digest <digest> --permission-reference <ref> ~ <stage> < request.json"]),
    ("update-apply", &["update-apply --plan-digest <digest> --permission-reference <ref> ~ <stage> < request.json"]),
    ("recovery-plan", &["recovery-plan ~", "recovery-apply --plan-digest <digest> --permission-reference <ref> ~ < request.json"]),
    ("recovery-apply", &["recovery-apply --plan-digest <digest> --permission-reference <ref> ~ < request.json"]),
    ("apply-recover", &["apply-recover --receipt <receipt-id> --permission-reference <ref> ."]),
    ("layout-plan", &["layout-plan .", "layout-apply --plan-digest <digest> --permission-reference <ref> ."]),
    ("layout-apply", &["layout-apply --plan-digest <digest> --permission-reference <ref> ."]),
    ("migrate-plan", &["migrate-plan ~ < request.json", "migrate-apply --plan-digest <digest> --permission-reference <ref> ~ < request.json"]),
    ("migrate-apply", &["migrate-apply --plan-digest <digest> --permission-reference <ref> ~ < request.json"]),
    ("migrate-recovery-plan", &["migrate-recovery-plan --backup <UTC> ~"]),
    ("migrate-recover", &["migrate-recover --backup <UTC> --plan-digest <digest> --permission-reference <ref> ~ < request.json"]),
    ("skills-legacy", &["skills-legacy ~"]),
    ("bin-names-plan", &["bin-names-plan", "bin-names-apply --plan-digest <digest> --permission-reference <ref>"]),
    ("bin-names-apply", &["bin-names-apply --plan-digest <digest> --permission-reference <ref>"]),
    ("agents-home-plan", &["agents-home-plan ~ skills/moai agents", "agents-home-apply --plan-digest <digest> --permission-reference <ref> ~ skills/moai agents"]),
    ("agents-home-apply", &["agents-home-apply --plan-digest <digest> --permission-reference <ref> ~ skills/moai agents"]),
    ("agents-home-revert", &["agents-home-revert --receipt <receipt-id> --permission-reference <ref> ~"]),
    ("home-routing-plan", &["home-routing-plan ~ skills/moai", "home-routing-apply --plan-digest <digest> --permission-reference <ref> ~ skills/moai"]),
    ("home-routing-apply", &["home-routing-apply --plan-digest <digest> --permission-reference <ref> ~ skills/moai"]),
    ("home-routing-revert", &["home-routing-revert --receipt <receipt-id> --permission-reference <ref> ~"]),
    ("key-set", &["key-set glm < key.txt"]),
    ("key-status", &["key-status glm"]),
    ("key-remove", &["key-remove glm"]),
    ("key-import-plan", &["key-import-plan glm", "key-import-apply --plan-digest <digest> --permission-reference <ref> glm"]),
    ("key-import-apply", &["key-import-apply --plan-digest <digest> --permission-reference <ref> glm"]),
    ("claude-doctor", &["claude-doctor ."]),
    ("codex-doctor", &["codex-doctor ."]),
    ("install-inspect", &["install-inspect"]),
    ("identity", &["identity ."]),
    ("worktree-report", &["worktree-report ."]),
    ("optimize-plan", &["optimize-plan --integration-ref origin/main ."]),
    ("baseline-inventory", &["baseline-inventory ."]),
    ("claim-acquire", &["claim-acquire . < request.json"]),
    ("claim-verify", &["claim-verify . < request.json"]),
    ("claim-show", &["claim-show --task <task> ."]),
    ("claim-release", &["claim-release --task <task> --epoch 1 ."]),
    ("claim-handoff", &["claim-handoff --task <task> --from-epoch 1 --reason <reason> ."]),
    ("queue-next", &["queue-next . < request.json"]),
    ("queue-free", &["queue-free . < request.json"]),
    ("queue-query", &["queue-query . < request.json"]),
    ("leader-tick", &["leader-tick . < request.json"]),
    ("lane-status", &["lane-status . < request.json"]),
    ("lane-handoff-check", &["lane-handoff-check . < request.json"]),
    ("lane-safe-to-reset", &["lane-safe-to-reset --lane agent-1 ."]),
    ("session-register", &["session-register --name agent-1 ."]),
    ("usage-limit", &["usage-limit ~"]),
    ("heavy-slot-acquire", &["heavy-slot-acquire --max 2 --name agent-1 -- cargo test"]),
    ("heavy-slot-status", &["heavy-slot-status"]),
    ("goal-show", &["goal-show ."]),
    ("goal-plan", &["goal-plan --request <request> .", "goal-apply --request <request> --plan-digest <digest> ."]),
    ("goal-apply", &["goal-apply --request <request> --plan-digest <digest> ."]),
    ("goal-block", &["goal-block < request.json"]),
    ("gate-check", &["gate-check < request.json"]),
    ("gate-merge", &["gate-merge --pr 2824 --permission-reference <ref>"]),
    ("gate-loop", &["gate-loop --permission-reference <ref>"]),
    ("wait", &["wait --repo modu-ai/moai-adk ci"]),
    ("scope-check", &["scope-check ."]),
    ("worktree-plan", &["worktree-plan --branch-prefix issue/ .", "worktree-apply --branch-prefix issue/ --plan-digest <digest> ."]),
    ("worktree-apply", &["worktree-apply --branch-prefix issue/ --plan-digest <digest> ."]),
    ("pr-body-render", &["pr-body-render --receipts <receipts> --base-sha <sha> --base-branch <base-branch> --default-branch <default-branch> ."]),
    ("pr-confirm", &["pr-confirm --task <task> --operation-id <operation-id> --observed-state <observed-state> ."]),
    ("audit-claim", &["audit-claim < request.json"]),
    ("audit-clone-plan", &["audit-clone-plan --pr 2824 --head <sha> --head-branch <head-branch> --base-branch <base-branch> ."]),
    ("labels-check", &["labels-check --repo modu-ai/moai-adk ."]),
    ("decision-due", &["decision-due docs/decisions/bundle.md"]),
    ("decision-file-check", &["decision-file-check docs/decisions/bundle.md"]),
    ("decision-append", &["decision-append --log docs/decisions-log.md . < request.json"]),
    ("flow-wip-check", &["flow-wip-check --owner modu-ai --repo modu-ai/moai-adk --lane agent-1"]),
    ("dispatch-record", &["dispatch-record --permission-reference <ref> . < request.json"]),
    ("dispatch-ack", &["dispatch-ack --permission-reference <ref> . < request.json"]),
    ("dispatch-check", &["dispatch-check < request.json"]),
    ("message-record", &["message-record --permission-reference <ref> . < request.json"]),
    ("message-ack", &["message-ack --permission-reference <ref> . < request.json"]),
    ("message-reconcile", &["message-reconcile . < request.json"]),
    ("effect-intend", &["effect-intend --task <task> --epoch 1 --operation <operation> --project-id <project-id> --local-id <local-id> --attempt 1 --kind ci . < request.json"]),
    ("effect-confirm", &["effect-confirm --task <task> --operation-id <operation-id> --observed-state <observed-state> ."]),
    ("effects-reconcile", &["effects-reconcile --task <task> ."]),
    ("review-package", &["review-package --base <sha> --head <sha> ."]),
    ("review-check", &["review-check --head <sha> . < findings.json"]),
    ("repo-scans", &["repo-scans < request.json"]),
    ("pr-self-audit-check", &["pr-self-audit-check < BODY.md"]),
    ("flake-classify", &["flake-classify --failed <failed> --later <later>"]),
    ("flake-record", &["flake-record --classified classified.json --run 43529973148 --permission-reference <ref>"]),
    ("audit-delta", &["audit-delta --pass-head <sha> --tip <sha> --head <sha> ."]),
    ("audit-reduce", &["audit-reduce . < request.json"]),
    ("parity-check", &["parity-check template generated"]),
    ("plan-admit", &["plan-admit . < request.json"]),
    ("run-verify", &["run-verify . < request.json"]),
    ("statusline", &["statusline < session.json"]),
    ("hook-receipt", &["hook-receipt --host claude < hook.json"]),
    ("hook-dispatch", &["hook-dispatch --host claude --event <event> < hook.json"]),
    ("forge-read", &["forge-read . < request.json"]),
    ("forge-links-probe", &["forge-links-probe . < request.json"]),
    ("forge-report", &["forge-report --permission-reference <ref> . < request.json"]),
    ("forge-audit-status", &["forge-audit-status --permission-reference <ref> . < request.json"]),
    ("index-query", &["index-query . < request.json"]),
    ("atlas-map", &["atlas-map . < request.json"]),
    ("atlas-graph", &["atlas-graph . < request.json"]),
    ("atlas-tag-scan", &["atlas-tag-scan ."]),
    ("search", &["search . < request.json"]),
    ("context-bundle", &["context-bundle . < request.json"]),
    ("context-verify", &["context-verify . < request.json"]),
    ("decision-record", &["decision-record --run <run> --permission-reference <ref> . < request.json"]),
    ("cost-report", &["cost-report < request.json"]),
    ("snapshot-describe", &["snapshot-describe ."]),
    ("package-inspect", &["package-inspect skills/moai"]),
    ("package-resolve", &["package-resolve ."]),
    ("codex-projection-plan", &["codex-projection-plan --target <folder> skills/moai", "codex-projection-apply --target <folder> --plan-digest <digest> --permission-reference <ref> skills/moai"]),
    ("codex-projection-apply", &["codex-projection-apply --target <folder> --plan-digest <digest> --permission-reference <ref> skills/moai"]),
    ("project-skill-plan", &["project-skill-plan . < request.json", "project-skill-apply . < request.json"]),
    ("project-skill-apply", &["project-skill-apply . < request.json"]),
    ("project-skill-revert", &["project-skill-revert . < request.json"]),
    ("archive-plan", &["archive-plan .", "archive-apply --plan-digest <digest> --permission-reference <ref> ."]),
    ("archive-apply", &["archive-apply --plan-digest <digest> --permission-reference <ref> ."]),
    ("archive-restore", &["archive-restore --spec SPEC-001 --permission-reference <ref> ."]),
    ("archive-search", &["archive-search --query cache ."]),
    ("worktree-gc-plan", &["worktree-gc-plan --integration-ref origin/main .", "worktree-gc-apply --integration-ref origin/main --plan-digest <digest> --permission-reference <ref> ."]),
    ("worktree-gc-apply", &["worktree-gc-apply --integration-ref origin/main --plan-digest <digest> --permission-reference <ref> ."]),
    ("session", &["session --name agent-1"]),
];

/// Descriptions of arguments that share one meaning wherever they appear (#2802 S1d), by clap
/// argument id. `configure` gives one to an argument only when the source carries none, so a
/// doc comment on the field always wins.
pub const ARG_HELP: &[(&str, &str)] = &[
    ("project", "Project root"),
    (
        "permission_reference",
        "Caller-supplied permission reference (recorded; not an authorization verifier)",
    ),
    ("plan_digest", "The preview's `plan_digest`"),
    (
        "task",
        "The 16-hex task key of a record in the local claim journal",
    ),
    ("operation_id", "Operation id of the intended effect"),
    (
        "provider",
        "Provider name, e.g. `glm` (lower-case letters, digits and `-`)",
    ),
];

/// Descriptions of arguments whose meaning depends on the command (#2802 S1d):
/// (command, argument id, text). Checked before `ARG_HELP`.
pub const ARG_HELP_FOR: &[(&str, &str, &str)] = &[
    (
        "decision-due",
        "file",
        "Decision bundle file under `docs/decisions/`",
    ),
    (
        "decision-file-check",
        "file",
        "Decision bundle file under `docs/decisions/`",
    ),
    ("package-inspect", "package", "Skill package folder"),
    (
        "package-resolve",
        "repo_package",
        "Repository package folder; required when the profile pins a package",
    ),
    (
        "package-resolve",
        "user_package",
        "User package folder; required when the profile pins none",
    ),
    ("identity", "path", "Any path inside a clone or worktree"),
    (
        "key-remove",
        "yes",
        "Remove the key; without it only report what would go",
    ),
    (
        "migrate-apply",
        "home",
        "Synthetic home: a temporary directory carrying `.moai-cli-synthetic-home`",
    ),
    (
        "migrate-recovery-plan",
        "home",
        "Synthetic home: a temporary directory carrying `.moai-cli-synthetic-home`",
    ),
    (
        "migrate-recover",
        "home",
        "Synthetic home: a temporary directory carrying `.moai-cli-synthetic-home`",
    ),
    (
        "migrate-recover",
        "backup",
        "The apply's backup id (its plan digest)",
    ),
    (
        "heavy-slot-status",
        "dir",
        "Slot folder (default `${XDG_CACHE_HOME:-~/.cache}/moai-cli/heavy-slots`)",
    ),
    (
        "language-setup",
        "questions",
        "Print the question set for a host question tool",
    ),
    (
        "language-setup",
        "answers",
        "Read the question tool's `{item: code}` JSON on stdin",
    ),
    (
        "claim-verify",
        "hook",
        "Read a host PreToolUse hook input instead of a typed request (refusal exits 2)",
    ),
    (
        "claim-release",
        "epoch",
        "The epoch of the current claim record",
    ),
    (
        "claim-handoff",
        "reason",
        "Why the claim is handed off; recorded",
    ),
    (
        "effect-intend",
        "project_id",
        "Project part of the operation id (07-B format)",
    ),
    (
        "effect-intend",
        "local_id",
        "Local part of the operation id",
    ),
    (
        "effect-intend",
        "attempt",
        "Attempt number of the operation id",
    ),
    ("hooks-apply", "host", "`claude` or `codex`"),
    ("hook-receipt", "host", "`claude` or `codex`"),
    (
        "codex-projection-plan",
        "package",
        "Common skill package folder to project",
    ),
    (
        "codex-projection-apply",
        "package",
        "Common skill package folder to project",
    ),
    (
        "codex-projection-plan",
        "display_name",
        "`display_name` written to `agents/openai.yaml`",
    ),
    (
        "codex-projection-apply",
        "display_name",
        "`display_name` written to `agents/openai.yaml`",
    ),
    (
        "codex-projection-plan",
        "short_description",
        "`short_description` written to `agents/openai.yaml`",
    ),
    (
        "codex-projection-apply",
        "short_description",
        "`short_description` written to `agents/openai.yaml`",
    ),
    (
        "codex-projection-plan",
        "default_prompt",
        "`default_prompt` written to `agents/openai.yaml`",
    ),
    (
        "codex-projection-apply",
        "default_prompt",
        "`default_prompt` written to `agents/openai.yaml`",
    ),
    (
        "codex-projection-apply",
        "allow_implicit_invocation",
        "Write `policy.allow_implicit_invocation`; omitted keeps Codex's default",
    ),
    (
        "worktree-apply",
        "branch_prefix",
        "Prefix of the new branch name (required; no default)",
    ),
    (
        "worktree-apply",
        "path",
        "Worktree path; default `<primary>/.moai-cli/worktrees/<issue>-<logical_key>`",
    ),
    (
        "worktree-gc-apply",
        "integration_ref",
        "Local ref of the integration branch, e.g. `origin/v2026.11.0`; nothing is fetched",
    ),
    (
        "worktree-gc-apply",
        "recent_minutes",
        "A change newer than this many minutes keeps a worktree",
    ),
    (
        "worktree-gc-apply",
        "budget_ms",
        "Time budget in milliseconds; past it the remaining worktrees are `skip`",
    ),
    ("parity-check", "template", "Template tree"),
    (
        "parity-check",
        "generated",
        "Generated tree compared with the template",
    ),
    (
        "parity-check",
        "doc",
        "Optional third set: the documented tree",
    ),
    ("atlas-tag-scan", "path", "File or directory to scan"),
    (
        "pr-body-render",
        "base_sha",
        "Exact commit of the base branch",
    ),
    ("pr-body-render", "base_branch", "Branch the PR merges into"),
    (
        "pr-body-render",
        "default_branch",
        "The repository's default branch (closing keywords work only there)",
    ),
    (
        "pr-body-render",
        "not_run",
        "A check that did not run, listed under not_run (repeatable)",
    ),
    (
        "pr-body-render",
        "review_link",
        "Link to the independent review; a waiting line when absent",
    ),
    (
        "archive-plan",
        "grace_days",
        "Inactivity window in days instead of the profile's",
    ),
    (
        "archive-apply",
        "grace_days",
        "Inactivity window in days instead of the profile's",
    ),
    (
        "archive-restore",
        "spec",
        "Id of the archived SPEC to restore",
    ),
    ("archive-search", "query", "Text to search for"),
    (
        "decision-record",
        "run",
        "Run id; the record goes to `.moai-cli/decisions/<RUN>.jsonl`",
    ),
    ("audit-clone-plan", "pr", "Pull request number"),
    ("audit-clone-plan", "head_branch", "Branch of the PR head"),
    (
        "audit-clone-plan",
        "base_branch",
        "Branch the PR merges into",
    ),
    (
        "decision-append",
        "apply",
        "Write the rows; without it only preview",
    ),
    (
        "go-state-revert",
        "backup",
        "The apply's backup stamp (`YYYYMMDDTHHMMSSZ`)",
    ),
    (
        "update-apply",
        "home",
        "The HOME to change (default `$HOME`). Tests pass a temporary folder",
    ),
    (
        "update-apply",
        "stage",
        "Package directory carrying SKILL.md (staged new state)",
    ),
    (
        "recovery-apply",
        "home",
        "The HOME to change (default `$HOME`). Tests pass a temporary folder",
    ),
    (
        "session",
        "host",
        "Only `claude`; other hosts start with `cc|glm|codex -l|-a`",
    ),
    ("session", "model", "Claude model"),
    ("session", "effort", "Claude effort"),
    ("session", "claude_bin", "Claude Code executable to run"),
];

/// The description `configure` gives an argument of `command` that has none.
pub fn arg_help(command: &str, id: &str) -> Option<&'static str> {
    ARG_HELP_FOR
        .iter()
        .find(|(c, i, _)| *c == command && *i == id)
        .map(|(_, _, t)| *t)
        .or_else(|| ARG_HELP.iter().find(|(i, _)| *i == id).map(|(_, t)| *t))
}

/// The launchers the root help shows on one line.
const LAUNCHERS: [&str; 3] = ["cc", "codex", "glm"];
const LAUNCHER_LINE: (&str, &str) = (
    "cc, codex, glm",
    "Start a Claude Code, Codex or GLM session",
);

/// The group heading of a command word, or `None` for a word in no group.
pub fn group_of(word: &str) -> Option<&'static str> {
    GROUPS
        .iter()
        .find(|g| g.commands.iter().any(|(w, _)| *w == word))
        .map(|g| g.heading)
}

/// The name a person typed: `moai` or `moai-cli` from `argv[0]` (#2802, D2 section 9); any
/// other name (a renamed copy) reads as `moai-cli`, the canonical one.
pub fn bin_name(argv0: &str) -> &'static str {
    let stem = std::path::Path::new(argv0)
        .file_stem()
        .and_then(|s| s.to_str())
        .unwrap_or("");
    if stem == "moai" {
        "moai"
    } else {
        "moai-cli"
    }
}

/// The width help lines wrap at: `COLUMNS` when set and sane, else 100 (clap's own cap).
fn width() -> usize {
    std::env::var("COLUMNS")
        .ok()
        .and_then(|c| c.trim().parse::<usize>().ok())
        .filter(|w| (20..=1000).contains(w))
        .unwrap_or(100)
        .min(100)
}

/// `  name  about` rows under each heading, the about column wrapped to `width` with a hanging
/// indent; `all` lists every command, otherwise only the root ones (launchers on one line).
pub fn render_groups(cmd: &Command, all: bool, width: usize) -> String {
    let about = |word: &str| -> String {
        cmd.find_subcommand(word)
            .and_then(|s| s.get_about())
            .map(|a| a.to_string())
            .unwrap_or_default()
    };
    let mut rows: Vec<(&str, Vec<(String, String)>)> = Vec::new();
    for g in GROUPS {
        let mut list = Vec::new();
        for (word, root) in g.commands {
            if !all && !root {
                continue;
            }
            if !all && LAUNCHERS.contains(word) {
                if *word == LAUNCHERS[0] {
                    list.push((LAUNCHER_LINE.0.to_string(), LAUNCHER_LINE.1.to_string()));
                }
                continue;
            }
            list.push((word.to_string(), about(word)));
        }
        if !list.is_empty() {
            rows.push((g.heading, list));
        }
    }
    let name_w = rows
        .iter()
        .flat_map(|(_, l)| l.iter().map(|(n, _)| n.width()))
        .max()
        .unwrap_or(0);
    let indent = 2 + name_w + 2;
    let mut out = String::new();
    for (heading, list) in rows {
        out.push_str(heading);
        out.push('\n');
        for (name, text) in list {
            let pad = name_w - name.width();
            let row = format!(
                "  {name}{}  {}",
                " ".repeat(pad),
                wrap(&text, indent, width)
            );
            // A command without an about leaves no padding at the line end.
            out.push_str(row.trim_end());
            out.push('\n');
        }
    }
    out
}

/// `text` broken at spaces so no line passes `width` display cells (a wide character counts
/// two), later lines indented by `indent`.
fn wrap(text: &str, indent: usize, width: usize) -> String {
    let room = width.saturating_sub(indent).max(20);
    let mut lines: Vec<String> = Vec::new();
    let mut line = String::new();
    for word in text.split_whitespace() {
        if !line.is_empty() && line.width() + 1 + word.width() > room {
            lines.push(std::mem::take(&mut line));
        }
        if !line.is_empty() {
            line.push(' ');
        }
        line.push_str(word);
    }
    if !line.is_empty() || lines.is_empty() {
        lines.push(line);
    }
    lines.join(&format!("\n{}", " ".repeat(indent)))
}

/// The root command with the people-facing layout: usage and help name the typed `bin`, and the
/// root help body is the grouped list (`all` for `help --all`).
pub fn configure(cmd: Command, bin: &'static str, version: &str, all: bool) -> Command {
    let groups = render_groups(&cmd, all, width());
    let footer = if all {
        format!("Run '{bin} <command> --help' for details and examples.")
    } else {
        format!(
            "Run '{bin} help --all' for every command.\n\
             Run '{bin} <command> --help' for details and examples."
        )
    };
    // The template is literal text plus clap's `{usage-heading} {usage}` and, for `--all`,
    // `{options}`; braces in the group text would read as template tags, and command abouts
    // carry none (pinned by `help_text_has_no_template_braces`).
    let options = if all { "\nOptions:\n{options}\n" } else { "" };
    let template = format!(
        "MoAI-ADK {version} — agentic development kit\n\n{{usage-heading}} {{usage}}\n\n{groups}{options}\n{footer}\n"
    );
    // `mut_subcommands` keeps the command order and never panics on a name the table lacks
    // (`mut_subcommand` moves the command to the end and panics on an unknown one).
    let cmd = cmd.mut_subcommands(
        |c| match EXAMPLES.iter().find(|(w, _)| *w == c.get_name()) {
            Some((_, examples)) => {
                let text = examples
                    .iter()
                    .map(|e| format!("  {bin} {e}"))
                    .collect::<Vec<_>>()
                    .join("\n");
                c.after_long_help(format!("Examples:\n{text}"))
            }
            None => c,
        },
    );
    let cmd = cmd.mut_subcommands(|c| {
        let name = c.get_name().to_string();
        c.mut_args(|a| {
            if a.get_help().is_some() || a.get_long_help().is_some() {
                return a;
            }
            match arg_help(&name, a.get_id().as_str()) {
                Some(text) => a.help(text),
                None => a,
            }
        })
    });
    cmd.bin_name(bin)
        .override_usage(format!(
            "{bin} <command> [options]\n       {bin} cc|codex|glm -l|-a    lead or agent session"
        ))
        .help_template(template)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn bin_name_reads_the_typed_name() {
        assert_eq!(bin_name("/usr/local/bin/moai"), "moai");
        assert_eq!(bin_name("moai"), "moai");
        assert_eq!(bin_name("/x/moai-cli"), "moai-cli");
        assert_eq!(bin_name("/x/moai-copy"), "moai-cli");
        assert_eq!(bin_name(""), "moai-cli");
    }

    #[test]
    fn wrap_keeps_lines_inside_the_width() {
        let text = "Preview moving the Go-era .moai folder and everything it holds";
        let out = wrap(text, 24, 60);
        for line in out.lines() {
            assert!(line.chars().count() <= 60, "{line:?}");
        }
        assert_eq!(out.split_whitespace().collect::<Vec<_>>().join(" "), text);
        assert_eq!(wrap("short", 24, 60), "short");
    }

    /// Hangul counts two cells, so a Korean about wraps at half the characters (S6 will need it).
    #[test]
    fn wrap_counts_wide_characters_as_two_cells() {
        let text = "한글 설명 문장은 두 칸씩 차지해요 그래서 더 일찍 줄을 바꿔요";
        let out = wrap(text, 4, 30);
        for line in out.lines() {
            assert!(UnicodeWidthStr::width(line) <= 30, "{line:?}");
        }
        assert!(out.lines().count() >= 3, "{out}");
    }

    #[test]
    fn group_headings_are_unique() {
        let mut seen = std::collections::BTreeSet::new();
        for g in GROUPS {
            assert!(seen.insert(g.heading), "{}", g.heading);
        }
    }
}
