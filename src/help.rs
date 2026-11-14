//! The people-facing help layout (#2802 S1; design #2799 D2 section 9, leader D3 6865936070).
//!
//! `moai --help` lists the commands a person types, in five groups; `moai help --all` lists
//! every command in eight groups. The JSON envelope on stdout is not touched by anything here:
//! help is text for people. clap has no way to put subcommands under several headings, so the
//! root help body is drawn from [`GROUPS`] and handed to clap as the root's help template.

use crate::ui_text::Lang;
use clap::Command;
use unicode_width::UnicodeWidthStr;

/// One help group: its heading and its commands, each with whether the root help shows it.
pub struct Group {
    pub heading: &'static str,
    /// The heading on a Korean terminal (S6d, #2825).
    pub heading_ko: &'static str,
    pub commands: &'static [(&'static str, bool)],
}

/// Every command word in exactly one group (pinned by `help_table_names_every_command_once`).
/// `true` marks the commands the root help shows (#2799 B 6865555998, root draft 6865796537);
/// the rest appear only in `help --all`.
pub const GROUPS: &[Group] = &[
    Group {
        heading: "Start",
        heading_ko: "시작",
        commands: &[("cc", true), ("codex", true), ("glm", true)],
    },
    Group {
        heading: "Project",
        heading_ko: "프로젝트",
        commands: &[
            ("init", false),
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
            ("hooks-revert", false),
            ("go-hooks-plan", false),
            ("go-hooks-apply", false),
            ("go-hooks-revert", false),
        ],
    },
    Group {
        heading: "Update and move",
        heading_ko: "업데이트와 이동",
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
            ("cleanup-plan", false),
            ("cleanup-apply", false),
            ("cleanup-revert-plan", false),
            ("cleanup-revert", false),
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
            ("home-style-plan", false),
            ("home-style-apply", false),
            ("home-style-revert", false),
        ],
    },
    Group {
        heading: "Keys",
        heading_ko: "키",
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
        heading_ko: "진단",
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
        heading_ko: "레인과 리더",
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
            ("lane-next", false),
            ("lane-loop-line", false),
            ("lane-stop-hook", false),
            ("lane-resume-note", false),
            ("lane-resume-check", false),
            ("lane-duty-line", false),
            ("lane-status", false),
            ("lane-handoff-check", false),
            ("lane-safe-to-reset", false),
            ("lane-env", false),
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
            ("audit-verdict-line", false),
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
        heading_ko: "검사",
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
        heading_ko: "내부",
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
    ("init", &["init", "init --new demo", "init --new demo --dry-run"]),
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
    ("hooks-plan", &["hooks-plan --host claude .", "hooks-apply --host claude --plan-digest <digest> --permission-reference <ref> .", "hooks-plan --host codex --lane-loop agent-1 --repo modu-ai/moai-adk ."]),
    ("hooks-apply", &["hooks-apply --host claude --plan-digest <digest> --permission-reference <ref> ."]),
    ("hooks-revert", &["hooks-revert --receipt <id> --permission-reference <ref> ."]),
    ("go-hooks-plan", &["go-hooks-plan .", "go-hooks-apply --plan-digest <digest> --permission-reference <ref> ."]),
    ("go-hooks-apply", &["go-hooks-apply --plan-digest <digest> --permission-reference <ref> ."]),
    ("go-hooks-revert", &["go-hooks-revert --receipt <id> --permission-reference <ref> ."]),
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
    ("cleanup-plan", &["cleanup-plan ~ <stage>", "cleanup-apply --confirm-cleanup <token> --permission-reference <ref> ~ <stage>"]),
    ("cleanup-apply", &["cleanup-apply --confirm-cleanup <token> --permission-reference <ref> ~ <stage>"]),
    ("cleanup-revert-plan", &["cleanup-revert-plan --stamp <stamp> ~", "cleanup-revert --stamp <stamp> --permission-reference <ref> ~"]),
    ("cleanup-revert", &["cleanup-revert --stamp <stamp> --permission-reference <ref> ~"]),
    ("recovery-plan", &["recovery-plan ~","recovery-apply --plan-digest <digest> --permission-reference <ref> ~ < request.json"]),
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
    ("home-style-plan", &["home-style-plan ~ skills/moai", "home-style-apply --plan-digest <digest> --permission-reference <ref> ~ skills/moai"]),
    ("home-style-apply", &["home-style-apply --plan-digest <digest> --permission-reference <ref> ~ skills/moai"]),
    ("home-style-revert", &["home-style-revert --receipt <receipt-id> --permission-reference <ref> ~"]),
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
    (
        "lane-next",
        &[
            "lane-next --lane agent-1 --repo modu-ai/moai-adk",
            "lane-next --lane agent-1 --repo modu-ai/moai-adk --wait 1800",
        ],
    ),
    (
        "lane-loop-line",
        &[
            "lane-loop-line --host claude --lane agent-1",
            "lane-loop-line --host codex --lane agent-1 --repo modu-ai/moai-adk",
        ],
    ),
    (
        "lane-resume-note",
        &["lane-resume-note --lane agent-1 --issue 2904 --claim 6874514212 --lease-until 2026-11-14T14:00:00Z --branch issue/2904-x --head <40-hex> --uncommitted 0 --next push"],
    ),
    ("lane-resume-check", &["lane-resume-check --file comment.md"]),
    (
        "lane-stop-hook",
        &["lane-stop-hook --lane agent-1 --repo modu-ai/moai-adk < stop.json"],
    ),
    (
        "lane-duty-line",
        &[
            "lane-duty-line --lane 1",
            "lane-duty-line --lane 1 --host codex --first <task>",
        ],
    ),
    ("lane-status", &["lane-status . < request.json"]),
    ("lane-handoff-check", &["lane-handoff-check . < request.json"]),
    ("lane-safe-to-reset", &["lane-safe-to-reset --lane agent-1 ."]),
    (
        "session-register",
        &[
            "session-register --name agent-1 .",
            "session-register --show --name agent-1 .",
            "session-register --release --by leader --name agent-1 --instance <instance> .",
        ],
    ),
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
    ("audit-verdict-line", &["audit-verdict-line --pr 2850 --head <head> --verdict PASS --title <scope>", "audit-verdict-line --pr 2850 --head <head> --verdict FAIL --title <scope> --base <earlier-head>"]),
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
    ("lane-env", &["lane-env --lane agent-1"]),
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
        "Remove without asking; without it a preview, or one question on a terminal",
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
        "Worktree path; default `<HOME>/.moai/worktrees/<project>-<issue>-<logical_key>`",
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
        "Run id; the record goes to `decisions/<RUN>.jsonl` in the project state folder",
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

/// The Korean one-line about of each command the root help shows (S6d, #2825; design 8). The
/// other commands keep their English about in `help --all` until S6e moves the command help.
/// `the_korean_abouts_cover_exactly_the_root_commands` keeps this table in step with [`GROUPS`].
pub const ABOUT_KO: &[(&str, &str)] = &[
    ("project-detect", "프로젝트의 사실 알아내기"),
    ("project-plan", "프로젝트 프로필 미리 보기"),
    ("project-apply", "미리 본 프로젝트 프로필 쓰기"),
    ("language-setup", "쓰는 언어를 항목마다 고르기"),
    ("index-build", "코드 지도 색인 만들기"),
    ("update", "설치된 패키지로 MoAI 업데이트"),
    ("go-state-plan", "Go 세대 .moai 폴더 옮기기 미리 보기"),
    ("go-state-apply", "미리 본 Go 세대 .moai 폴더 옮기기"),
    (
        "go-state-revert-plan",
        "Go 세대 폴더 옮기기 되돌리기 미리 보기",
    ),
    ("go-state-revert", "Go 세대 폴더 옮기기 되돌리기"),
    ("home-state-plan", "Go 세대 ~/.moai 폴더 옮기기 미리 보기"),
    ("home-state-apply", "미리 본 Go 세대 ~/.moai 폴더 옮기기"),
    ("home-state-revert", "~/.moai 폴더 옮기기 되돌리기"),
    ("worktree-move-plan", "이 클론의 worktree 옮기기 미리 보기"),
    ("worktree-move-apply", "미리 본 worktree 옮기기"),
    ("key-set", "공급자 키를 금고에 저장"),
    ("key-status", "공급자 키를 보이지 않고 확인"),
    ("key-remove", "금고에서 공급자 키 지우기"),
    ("key-import-plan", "옛 키 파일 가져오기 미리 보기"),
    ("key-import-apply", "미리 본 옛 키 파일 가져오기"),
    ("claude-doctor", "Claude Code 가 AGENTS.md 를 읽는지 확인"),
    ("codex-doctor", "Codex 지침과 스킬 확인"),
    ("install-inspect", "설치된 MoAI 실행 파일 보기"),
    ("identity", "경로가 클론 안 어디에 있는지 보기"),
    ("worktree-report", "worktree 가 쓰는 공간 보고"),
    ("optimize-plan", "비울 수 있는 공간 미리 보기"),
    ("baseline-inventory", "체크아웃의 기준 사실 목록"),
];

/// The launcher line's about on a Korean terminal.
const LAUNCHER_LINE_KO: &str = "Claude Code·Codex·GLM 세션 시작";
/// clap's `Usage:` heading on a Korean terminal.
const USAGE_KO: &str = "사용법:";
/// clap's `Options:` heading of `help --all` on a Korean terminal.
const OPTIONS_KO: &str = "옵션:";
/// A subcommand help's `Arguments:` and `Examples:` headings on a Korean terminal (S6e-1).
const ARGUMENTS_KO: &str = "인자:";
const EXAMPLES_KO: &str = "예시:";
/// The help flag's line in a subcommand's short (`-h`) and long (`--help`) help (S6e-1).
const HELP_FLAG_SHORT_KO: &str = "도움말 보기 (자세히는 '--help')";
const HELP_FLAG_LONG_KO: &str = "도움말 보기 (요약은 '-h')";
/// The usage's second line ends with what `-l`/`-a` start.
const LEAD_OR_AGENT: &str = "lead or agent session";
const LEAD_OR_AGENT_KO: &str = "리더·에이전트 세션";

/// A group's heading in `lang`.
pub fn heading_in(g: &Group, lang: Lang) -> &'static str {
    match lang {
        Lang::En => g.heading,
        Lang::Ko => g.heading_ko,
    }
}

/// The about of a root command on a Korean terminal, when the table has one.
fn about_ko(word: &str) -> Option<&'static str> {
    ABOUT_KO.iter().find(|(w, _)| *w == word).map(|(_, t)| *t)
}

/// The lines under the grouped list: where every command is, and where the details are.
fn footer(lang: Lang, bin: &str, all: bool) -> String {
    let (every, details) = match lang {
        Lang::En => (
            format!("Run '{bin} help --all' for every command."),
            format!("Run '{bin} <command> --help' for details and examples."),
        ),
        Lang::Ko => (
            format!("모든 명령 보기: '{bin} help --all'"),
            format!("자세한 설명과 예시: '{bin} <command> --help'"),
        ),
    };
    if all {
        details
    } else {
        format!("{every}\n{details}")
    }
}

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

/// The narrowest `COLUMNS` the help follows; a smaller (or unreadable) value means [`MAX_WIDTH`].
pub const MIN_COLUMNS: usize = 20;
/// The widest any help line gets, the grouped list and clap's own sections alike.
pub const MAX_WIDTH: usize = 100;

/// The width help lines wrap at: `COLUMNS` when set, readable and at least [`MIN_COLUMNS`],
/// never more than [`MAX_WIDTH`]; otherwise [`MAX_WIDTH`].
pub fn width_from(columns: Option<&str>) -> usize {
    columns
        .and_then(|c| c.trim().parse::<usize>().ok())
        .filter(|w| *w >= MIN_COLUMNS)
        .unwrap_or(MAX_WIDTH)
        .min(MAX_WIDTH)
}

fn width() -> usize {
    width_from(std::env::var("COLUMNS").ok().as_deref())
}

/// `  name  about` rows under each heading, the about column wrapped to `width` with a hanging
/// indent; `all` lists every command, otherwise only the root ones (launchers on one line).
pub fn render_groups(cmd: &Command, all: bool, width: usize) -> String {
    render_groups_in(cmd, all, width, Lang::En)
}

/// [`render_groups`] in `lang`: headings, the launcher line and the root commands' abouts.
pub fn render_groups_in(cmd: &Command, all: bool, width: usize, lang: Lang) -> String {
    let about = |word: &str| -> String {
        if lang == Lang::Ko {
            if let Some(text) = about_ko(word) {
                return text.to_string();
            }
        }
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
                    let text = match lang {
                        Lang::En => LAUNCHER_LINE.1,
                        Lang::Ko => LAUNCHER_LINE_KO,
                    };
                    list.push((LAUNCHER_LINE.0.to_string(), text.to_string()));
                }
                continue;
            }
            list.push((word.to_string(), about(word)));
        }
        if !list.is_empty() {
            rows.push((heading_in(g, lang), list));
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
    let room = width.saturating_sub(indent).max(MIN_COLUMNS);
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
    // Help goes to stdout, so its language is stdout's: English whenever stdout is not a
    // terminal (S6-2), whatever stderr is.
    let lang = crate::ui::ui_lang(crate::ui::Stream::Stdout);
    configure_in(cmd, bin, version, all, lang)
}

/// [`configure`] in `lang` (S6d, #2825). The first line names the product and stays English
/// (design 1, like the banner's identity line).
pub fn configure_in(
    cmd: Command,
    bin: &'static str,
    version: &str,
    all: bool,
    lang: Lang,
) -> Command {
    let groups = render_groups_in(&cmd, all, width(), lang);
    let footer = footer(lang, bin, all);
    let options_heading = match lang {
        Lang::En => "Options:",
        Lang::Ko => OPTIONS_KO,
    };
    // clap prints the root usage in a usage error on stderr too (`moai --bogus`), so the usage
    // it holds stays English whatever stdout's language (audit #2909 B1, S6-2): a Korean help
    // spells its usage lines out in the template instead. The second line starts under the
    // first one's command (`Usage:` is six cells, a Korean heading two per letter).
    let usage = format!(
        "{bin} <command> [options]\n{}{bin} cc|codex|glm -l|-a    {LEAD_OR_AGENT}",
        " ".repeat("Usage:".width() + 1)
    );
    let usage_lines = match lang {
        Lang::En => "{usage-heading} {usage}".to_string(),
        Lang::Ko => format!(
            "{USAGE_KO} {bin} <command> [options]\n{}{bin} cc|codex|glm -l|-a    {LEAD_OR_AGENT_KO}",
            " ".repeat(USAGE_KO.width() + 1)
        ),
    };
    // The template is literal text plus clap's `{usage-heading} {usage}` and, for `--all`,
    // `{options}`; braces in the group text would read as template tags, and command abouts
    // carry none (pinned by `help_text_has_no_template_braces`).
    let options = if all {
        format!("\n{options_heading}\n{{options}}\n")
    } else {
        String::new()
    };
    let template = format!(
        "MoAI-ADK {version} — agentic development kit\n\n{usage_lines}\n\n{groups}{options}\n{footer}\n"
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
                let heading = match lang {
                    Lang::En => "Examples:",
                    Lang::Ko => EXAMPLES_KO,
                };
                c.after_long_help(format!("{heading}\n{text}"))
            }
            None => c,
        },
    );
    let cmd = match lang {
        Lang::En => cmd,
        Lang::Ko => cmd.mut_subcommands(korean_frame),
    };
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
    // clap wraps its own sections (options, arguments) at the terminal width it reads, `COLUMNS`
    // included, with no cap of its own; the cap holds every help line to `MAX_WIDTH`.
    cmd.bin_name(bin)
        .max_term_width(MAX_WIDTH)
        .override_usage(usage)
        .help_template(template)
}

/// The groups whose root commands' own help texts read Korean on a Korean terminal (S6e-2…,
/// one group per slice; `the_moved_groups_read_korean_in_their_help` in `main.rs` checks
/// every about, long about and argument of them).
pub const KOREAN_GROUPS: &[&str] = &["Project"];

/// The Korean long about of a moved command that has an English one (S6e-2). Wrapped by
/// [`wrap`] at the help width before clap sees it: clap counts a Hangul letter as one cell.
const LONG_ABOUT_KO: &[(&str, &str)] = &[
    (
        "project-detect",
        "`/moai project` 경로(#2496)가 쓰는 프로젝트의 사실을 알아내요: 새 프로젝트인지 기존 \
         프로젝트인지, manifest, 언어, 언어 서버, 설정된 MCP 서버 이름, 데이터베이스. 읽기만 하고 \
         아무것도 쓰지 않아요. `project-plan` 에 넘길, `detected` 가 채워진 프로필을 돌려줘요.",
    ),
    (
        "language-setup",
        "쓰는 언어를 항목마다 정해요(#2293): 항목마다 질문을 하나씩 하고 `project-plan` 에 넘길 \
         완전한 프로필을 돌려줘요. 아무것도 쓰지 않아요. `--questions` 는 호스트 질문 도구용 질문 \
         묶음을 찍고, `--answers` 는 그 도구의 JSON 답을 stdin 에서 읽어요. 두 플래그가 없으면 \
         터미널에서 물어요(질문은 stderr, 답은 stdin).",
    ),
    (
        "index-build",
        "증분 색인의 새 세대를 만들어요(C6 #2182): 추적 파일을 훑고, 바뀌지 않는 manifest 폴더를 \
         준비한 뒤 원자적 포인터 교체로 게시해요. 빌드가 실패하면 앞 세대를 그대로 두고 partial 로 \
         표시해요(REQ-C6-01).",
    ),
];

/// The Korean help of a moved command's argument (S6e-2): `(command, id, text)`, an empty
/// command for an argument id every command shares.
const ARG_HELP_KO: &[(&str, &str, &str)] = &[
    ("", "project", "프로젝트 루트"),
    ("", "plan_digest", "미리 보기의 `plan_digest`"),
    (
        "",
        "permission_reference",
        "호출자가 준 권한 참조(기록만, 권한 검증기 아님)",
    ),
    (
        "language-setup",
        "questions",
        "호스트 질문 도구용 질문 묶음 찍기",
    ),
    (
        "language-setup",
        "answers",
        "질문 도구의 JSON 답을 stdin 에서 읽기",
    ),
    (
        "language-setup",
        "conversation",
        "대화 기본 항목의 대화 언어 코드. 없으면 프로필의 `language.user_chat`, 그것도 없으면 `en`",
    ),
];

/// The Korean help of argument `id` of `command`, when the table has one.
fn arg_help_ko(command: &str, id: &str) -> Option<&'static str> {
    ARG_HELP_KO
        .iter()
        .find(|(c, i, _)| *c == command && *i == id)
        .or_else(|| {
            ARG_HELP_KO
                .iter()
                .find(|(c, i, _)| c.is_empty() && *i == id)
        })
        .map(|(_, _, t)| *t)
}

/// The column clap starts an argument's help at in a long help (`          <help>`).
const LONG_ARG_INDENT: usize = 10;

/// clap's own help template with the usage heading spelled out (clap 4.6 `DEFAULT_TEMPLATE`).
const SUBCOMMAND_TEMPLATE_KO: &str =
    "{before-help}{about-with-newline}\n사용법: {usage}\n\n{all-args}{after-help}";

/// A subcommand's help frame on a Korean terminal (S6e-1, #2825; design 5): the usage,
/// arguments and options headings and the help flag's line. The command's about, argument
/// texts and examples are left as they are; later S6e slices move them group by group.
fn korean_frame(c: Command) -> Command {
    let name = c.get_name().to_string();
    let moved = GROUPS
        .iter()
        .filter(|g| KOREAN_GROUPS.contains(&g.heading))
        .any(|g| g.commands.iter().any(|(w, root)| *root && *w == name));
    let c = if moved {
        korean_texts(c, &name, width())
    } else {
        c
    };
    let c = c.mut_args(|a| {
        let heading = if a.is_positional() {
            ARGUMENTS_KO
        } else {
            OPTIONS_KO
        };
        a.help_heading(heading.trim_end_matches(':'))
    });
    // clap's generated `-h, --help` cannot be edited before the command is built, so it is
    // replaced by the same flag with Korean lines (`ArgAction::Help`: `-h` short, `--help` long).
    c.disable_help_flag(true)
        .arg(
            clap::Arg::new("help")
                .short('h')
                .long("help")
                .action(clap::ArgAction::Help)
                .help(HELP_FLAG_SHORT_KO)
                .long_help(HELP_FLAG_LONG_KO)
                .help_heading(OPTIONS_KO.trim_end_matches(':')),
        )
        .help_template(SUBCOMMAND_TEMPLATE_KO)
}

/// A moved command's own texts in Korean (S6e-2): its about (the root list's), its long about
/// and its arguments' help, each wrapped at `width` display cells — clap would count a Hangul
/// letter as one cell and run the line past the terminal.
fn korean_texts(c: Command, name: &str, width: usize) -> Command {
    let c = match about_ko(name) {
        Some(about) => c.about(about),
        None => c,
    };
    let c = match LONG_ABOUT_KO.iter().find(|(w, _)| *w == name) {
        Some((_, long)) => c.long_about(wrap(long, 0, width)),
        None => c,
    };
    let name = name.to_string();
    c.mut_args(move |a| match arg_help_ko(&name, a.get_id().as_str()) {
        Some(text) => {
            let wrapped = wrap(text, 0, width.saturating_sub(LONG_ARG_INDENT));
            a.help(wrapped.clone()).long_help(wrapped)
        }
        None => a,
    })
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

    /// #2824 re-audit B3: `COLUMNS` below 20 or unreadable means 100, 20 is followed, and nothing
    /// goes past 100.
    #[test]
    fn width_follows_columns_between_twenty_and_one_hundred() {
        for (columns, want) in [
            (None, 100),
            (Some("19"), 100),
            (Some("20"), 20),
            (Some(" 60 "), 60),
            (Some("100"), 100),
            (Some("101"), 100),
            (Some("1001"), 100),
            (Some("0"), 100),
            (Some("wide"), 100),
            (Some("-5"), 100),
        ] {
            assert_eq!(width_from(columns), want, "{columns:?}");
        }
    }

    /// The description column keeps at least 20 columns however deep the indent: a column
    /// squeezed below that would put every word on its own line.
    #[test]
    fn wrap_keeps_a_twenty_column_floor() {
        let out = wrap("aaaa bbbb cccc dddd eeee", 55, 60);
        assert_eq!(out.lines().next(), Some("aaaa bbbb cccc dddd"), "{out:?}");
        assert_eq!(out.lines().count(), 2, "{out:?}");
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

    /// A stand-in root command: every grouped word with an English about of its own.
    fn root_cmd() -> Command {
        let mut cmd = Command::new("moai");
        for g in GROUPS {
            for (w, _) in g.commands {
                cmd = cmd.subcommand(Command::new(*w).about(format!("About {w}")));
            }
        }
        cmd
    }

    /// S6d AC-1 (#2825): every heading has its own Korean twin; English stays the table's word.
    #[test]
    fn every_heading_has_a_distinct_korean_twin() {
        let mut seen = std::collections::BTreeSet::new();
        for g in GROUPS {
            assert_eq!(heading_in(g, Lang::En), g.heading);
            let ko = heading_in(g, Lang::Ko);
            assert!(!ko.is_empty() && ko != g.heading, "{}", g.heading);
            assert!(ko.chars().any(|c| ('가'..='힣').contains(&c)), "{ko}");
            assert!(seen.insert(ko), "{ko}");
        }
    }

    /// S6d AC-2: the Korean table covers exactly the commands the root help shows (launchers on
    /// their own line), so a new root command without a Korean about fails here.
    #[test]
    fn the_korean_abouts_cover_exactly_the_root_commands() {
        let root: std::collections::BTreeSet<&str> = GROUPS
            .iter()
            .flat_map(|g| g.commands.iter())
            .filter(|(w, r)| *r && !LAUNCHERS.contains(w))
            .map(|(w, _)| *w)
            .collect();
        let table: std::collections::BTreeSet<&str> = ABOUT_KO.iter().map(|(w, _)| *w).collect();
        assert_eq!(table, root);
        assert_eq!(ABOUT_KO.len(), table.len(), "a word twice");
    }

    /// S6d AC-2 and design 3.2: Korean constants carry no control character and no template
    /// brace (the group text goes into clap's help template).
    #[test]
    fn korean_help_text_is_plain() {
        let mut all: Vec<&str> = ABOUT_KO.iter().map(|(_, t)| *t).collect();
        all.extend(GROUPS.iter().map(|g| g.heading_ko));
        all.extend([LAUNCHER_LINE_KO, USAGE_KO, OPTIONS_KO, LEAD_OR_AGENT_KO]);
        all.extend(
            [
                footer(Lang::Ko, "moai", true),
                footer(Lang::Ko, "moai", false),
            ]
            .map(|s| Box::leak(s.into_boxed_str()) as &str),
        );
        for t in all {
            assert!(!t.is_empty());
            assert!(!t.chars().any(|c| c.is_control() && c != '\n'), "{t:?}");
            assert!(!t.contains('{') && !t.contains('}'), "{t:?}");
        }
    }

    /// S6d AC-1/2: a Korean root list has Korean headings and abouts, the launcher line in
    /// Korean, and the same names; English output is today's.
    #[test]
    fn a_korean_root_list_is_korean_and_english_is_unchanged() {
        let cmd = root_cmd();
        let en = render_groups(&cmd, false, 100);
        assert_eq!(render_groups_in(&cmd, false, 100, Lang::En), en);
        assert!(en.starts_with("Start\n  cc, codex, glm"), "{en}");
        let ko = render_groups_in(&cmd, false, 100, Lang::Ko);
        assert!(ko.starts_with("시작\n  cc, codex, glm"), "{ko}");
        assert!(ko.contains(LAUNCHER_LINE_KO), "{ko}");
        assert!(!ko.contains("About "), "{ko}");
        assert!(ko.contains("  project-detect  "), "{ko}");
        assert_eq!(ko.lines().count(), en.lines().count());
    }

    /// S6d AC-2: `help --all` shows the root commands in Korean and keeps the other abouts in
    /// English until S6e.
    #[test]
    fn a_korean_full_list_keeps_the_other_abouts_english() {
        let cmd = root_cmd();
        let ko = render_groups_in(&cmd, true, 100, Lang::Ko);
        assert!(ko.contains("About init"), "{ko}");
        assert!(!ko.contains("About project-detect"), "{ko}");
        for g in GROUPS {
            assert!(
                ko.contains(&format!("{}\n", g.heading_ko)),
                "{}",
                g.heading_ko
            );
        }
    }

    /// A stand-in subcommand `name`'s help, configured as `main` does, in `lang`.
    fn sub_help_of(name: &'static str, lang: Lang, long: bool) -> String {
        let cmd = Command::new("moai").subcommand(
            Command::new(name)
                .about("Detect the project's facts")
                .long_about("Detect the project's facts, reading only.")
                .arg(
                    clap::Arg::new("project")
                        .value_name("PROJECT")
                        .required(true),
                )
                .arg(
                    clap::Arg::new("json")
                        .long("json")
                        .action(clap::ArgAction::SetTrue),
                ),
        );
        let mut cmd = configure_in(cmd, "moai", "9.9.9", false, lang);
        cmd.build();
        let sub = cmd.find_subcommand_mut(name).unwrap();
        if long {
            sub.render_long_help().to_string()
        } else {
            sub.render_help().to_string()
        }
    }

    /// The frame tests use `identity`: a command of a group whose own texts have not moved.
    fn sub_help(lang: Lang, long: bool) -> String {
        sub_help_of("identity", lang, long)
    }

    /// S6e-2: a command of a moved group takes its Korean about, long about and argument help;
    /// an argument the table lacks keeps its own; English is untouched.
    #[test]
    fn a_moved_command_reads_its_korean_texts() {
        let ko = sub_help_of("project-detect", Lang::Ko, true);
        assert!(ko.contains("`/moai project` 경로(#2496)"), "{ko}");
        assert!(!ko.contains("Detect the project's facts"), "{ko}");
        assert!(ko.contains("프로젝트 루트"), "{ko}");
        let short = sub_help_of("project-detect", Lang::Ko, false);
        assert!(short.starts_with("프로젝트의 사실 알아내기\n"), "{short}");
        let en = sub_help_of("project-detect", Lang::En, true);
        assert!(
            en.starts_with("Detect the project's facts, reading only."),
            "{en}"
        );
        assert!(!en.contains("프로젝트"), "{en}");
        // Wrapped at the width in display cells, whatever clap would count.
        let cmd = korean_texts(Command::new("language-setup"), "language-setup", 40);
        let long = cmd.get_long_about().unwrap().to_string();
        assert!(long.lines().count() > 3, "{long}");
        for line in long.lines() {
            assert!(line.width() <= 40, "{line:?}");
        }
    }

    /// S6e-1 (#2825): a subcommand's help frame — the usage, arguments, options and examples
    /// headings and the help flag — reads Korean on a Korean terminal; the command's own texts
    /// stay as they are (later S6e slices), and English is today's frame.
    #[test]
    fn a_korean_subcommand_help_has_a_korean_frame() {
        let en = sub_help(Lang::En, true);
        for head in [
            "\nUsage: ",
            "\nArguments:\n",
            "\nOptions:\n",
            "\nExamples:\n",
        ] {
            assert!(en.contains(head), "{head:?} in {en}");
        }
        assert!(en.contains("Print help (see a summary with '-h')"), "{en}");
        let ko = sub_help(Lang::Ko, true);
        for head in [
            format!("\n{USAGE_KO} "),
            format!("\n{ARGUMENTS_KO}\n"),
            format!("\n{OPTIONS_KO}\n"),
            format!("\n{EXAMPLES_KO}\n"),
        ] {
            assert!(ko.contains(&head), "{head:?} in {ko}");
        }
        for head in [
            "Usage:",
            "Arguments:",
            "Options:",
            "Examples:",
            "Print help",
        ] {
            assert!(!ko.contains(head), "{head:?} in {ko}");
        }
        assert!(ko.contains(HELP_FLAG_LONG_KO), "{ko}");
        assert!(
            ko.contains("Detect the project's facts, reading only."),
            "{ko}"
        );
        assert!(ko.contains("moai identity ."), "{ko}");
        // `-h` keeps the short form with its own Korean line.
        let short = sub_help(Lang::Ko, false);
        assert!(short.contains(HELP_FLAG_SHORT_KO), "{short}");
        assert!(
            !short.contains("Examples") && !short.contains(EXAMPLES_KO),
            "{short}"
        );
        let en_short = sub_help(Lang::En, false);
        assert!(
            en_short.contains("Print help (see more with '--help')"),
            "{en_short}"
        );
    }

    /// S6d AC-3: the Korean help names its usage and options headings and the two footer lines
    /// in Korean; the usage's second line stays under the first command; English is today's.
    #[test]
    fn the_korean_frame_has_korean_heads_and_footer() {
        // Audit #2909 B1: clap prints the root usage in a usage error on stderr, so a Korean
        // help leaves the usage clap holds English — the error bytes do not follow stdout.
        for args in [&["moai", "--bogus"][..], &["moai", "bogus"][..]] {
            let error = |lang: Lang| {
                configure_in(root_cmd(), "moai", "9.9.9", false, lang)
                    .try_get_matches_from(args)
                    .unwrap_err()
                    .render()
                    .to_string()
            };
            let en = error(Lang::En);
            assert!(en.contains("lead or agent session"), "{en}");
            assert_eq!(error(Lang::Ko), en, "{args:?}");
        }
        let render = |all: bool, lang: Lang| -> String {
            configure_in(root_cmd(), "moai", "9.9.9", all, lang)
                .render_help()
                .to_string()
        };
        let en = render(false, Lang::En);
        assert!(
            en.contains("\nUsage: moai <command> [options]\n       moai cc|"),
            "{en}"
        );
        assert!(
            en.contains("Run 'moai help --all' for every command."),
            "{en}"
        );
        let ko = render(false, Lang::Ko);
        assert!(
            ko.starts_with("MoAI-ADK 9.9.9 — agentic development kit\n"),
            "{ko}"
        );
        let usage = format!("\n{USAGE_KO} moai <command> [options]\n");
        assert!(ko.contains(&usage), "{ko}");
        let second = ko.lines().find(|l| l.contains("cc|codex|glm")).unwrap();
        let first = ko.lines().find(|l| l.starts_with(USAGE_KO)).unwrap();
        assert_eq!(
            second.find("moai").map(|i| second[..i].width()),
            first.find("moai").map(|i| first[..i].width()),
            "{ko}"
        );
        assert!(second.ends_with(LEAD_OR_AGENT_KO), "{second}");
        assert!(
            ko.ends_with(&format!("{}\n", footer(Lang::Ko, "moai", false))),
            "{ko}"
        );
        assert!(!ko.contains("Run '"), "{ko}");
        let all = render(true, Lang::Ko);
        assert!(all.contains(&format!("\n{OPTIONS_KO}\n")), "{all}");
        assert!(!all.contains("\nOptions:\n"), "{all}");
    }
}
