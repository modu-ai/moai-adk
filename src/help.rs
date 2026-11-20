//! The people-facing help layout (#2802 S1; design #2799 D2 section 9, leader D3 6865936070).
//!
//! `moai --help` lists the commands a person types, in five groups; `moai help --all` lists
//! every command in eight groups. The JSON envelope on stdout is not touched by anything here:
//! help is text for people. clap has no way to put subcommands under several headings, so the
//! root help body is drawn from [`GROUPS`] and handed to clap as the root's help template.

use crate::ui::{Painter, Token};
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
        commands: &[("cc", true), ("codex", true), ("glm", true), ("omp", true)],
    },
    Group {
        heading: "Project",
        heading_ko: "프로젝트",
        commands: &[
            ("project-detect", true),
            ("project-plan", true),
            ("project-apply", true),
            ("project-update-plan", false),
            ("work-flow-plan", true),
            ("language-setup", true),
            ("language-set", false),
            ("index-build", true),
            ("check-tree", false),
            ("check", false),
            ("ledger-status", false),
            ("card", false),
            ("spec", false),
            ("tdd", false),
            ("ledger-register-plan", false),
            ("ledger-register-apply", false),
            ("card-import-plan", false),
            ("card-import-apply", false),
            ("card-summary-plan", false),
            ("card-summary-apply", false),
            ("ledger-recover-plan", false),
            ("ledger-recover-apply", false),
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
            ("go-settings-plan", false),
            ("go-settings-apply", false),
            ("go-settings-revert", false),
            ("agents-merge-plan", false),
            ("agents-merge-apply", false),
            ("agents-merge-revert", false),
            ("agents-project-plan", false),
            ("agents-project-apply", false),
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
            ("home-retire-plan", false),
            ("home-retire-apply", false),
            ("home-retire-revert", false),
            ("home-install-plan", false),
            ("home-install-apply", false),
            ("home-install-revert", false),
            ("worktree-move-plan", true),
            ("worktree-move-apply", true),
            ("update-plan", false),
            ("update-apply", false),
            ("reports-move-plan", false),
            ("reports-move-apply", false),
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
            ("keys-revert", false),
            ("models-revert", false),
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
            ("claim-timeout-check", false),
            ("claim-expired", false),
            ("audit-claim-stale", false),
            ("claim-show", false),
            ("claim-release", false),
            ("claim-handoff", false),
            ("claim-link", false),
            ("queue-free", false),
            ("queue-query", false),
            ("leader-tick", false),
            ("leader-pulse", false),
            ("leader-round", false),
            ("lane-next", false),
            ("lane-loop-line", false),
            ("lane-stop-hook", false),
            ("lane-resume-note", false),
            ("lane-resume-check", false),
            ("lane-duty-line", false),
            ("lane-status", false),
            ("lane-screen", false),
            ("lane-send", false),
            ("lane-handoff-check", false),
            ("lane-safe-to-reset", false),
            ("close-check", false),
            ("card-times", false),
            ("lane-env", false),
            ("session-register", false),
            ("session-current", false),
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
            ("gate-schedule-plan", false),
            ("gate-schedule-apply", false),
            ("wait", false),
            ("scope-check", false),
            ("worktree-plan", false),
            ("worktree-apply", false),
            ("pr-body-render", false),
            ("pr-confirm", false),
            ("audit-claim", false),
            ("audit-verdict-line", false),
            ("audit-gate", false),
            ("trailers", false),
            ("trailer-check", false),
            ("bundle", false),
            ("audit-clone-plan", false),
            ("labels-check", false),
            ("decision-due", false),
            ("decision-file-check", false),
            ("decision-append", false),
            ("flow-wip-check", false),
            ("dispatch-record", false),
            ("dispatch-ack", false),
            ("dispatch-start", false),
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
            ("worktree-install-plan", false),
            ("hygiene-plan", false),
            ("hygiene-apply", false),
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
    ("omp", &["omp -a", "omp -l --print-command"]),
    ("project-detect", &["project-detect ."]),
    ("project-plan", &["project-plan . < request.json", "project-apply --plan-digest <digest> --permission-reference <ref> . < request.json"]),
    ("project-apply", &["project-apply --plan-digest <digest> --permission-reference <ref> . < request.json"]),
    ("project-update-plan", &["project-update-plan ."]),
    ("work-flow-plan", &["work-flow-plan --project ."]),
    ("language-setup", &["language-setup . < request.json"]),
    ("language-set", &["language-set --home ko"]),
    ("index-build", &["index-build . < request.json"]),
    ("check-tree", &["check-tree ."]),
    (
        "check",
        &[
            "check affected <base>..<head> --list",
            "check affected <base>..<head>",
        ],
    ),
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
    ("go-settings-plan", &["go-settings-plan .", "go-settings-apply --plan-digest <digest> --permission-reference <ref> ."]),
    ("go-settings-apply", &["go-settings-apply --plan-digest <digest> --permission-reference <ref> ."]),
    ("go-settings-revert", &["go-settings-revert --receipt <id> --permission-reference <ref> ."]),
    ("agents-merge-plan", &["agents-merge-plan .", "agents-merge-apply --plan-digest <digest> --permission-reference <ref> ."]),
    ("agents-merge-apply", &["agents-merge-apply --plan-digest <digest> --permission-reference <ref> ."]),
    ("agents-merge-revert", &["agents-merge-revert --receipt <id> --permission-reference <ref> ."]),
    ("agents-project-plan", &["agents-project-plan . skills/moai agents", "agents-project-apply --plan-digest <digest> --permission-reference <ref> . skills/moai agents"]),
    ("agents-project-apply", &["agents-project-apply --plan-digest <digest> --permission-reference <ref> . skills/moai agents"]),
    ("ledger-status", &["ledger-status ."]),
    (
        "card",
        &[
            "card add --title fix-login --ac login-works --scope src/login/ --project .",
            "card ready C-1 --project .",
            "card note C-1 --kind report --project . < note.md",
        ],
    ),
    (
        "spec",
        &[
            "spec new login-fix --title fix-login --minimal --project .",
            "spec check --all --project .",
            "spec show SPEC-20261117-login-fix --project .",
        ],
    ),
    (
        "tdd",
        &[
            "tdd red --check ./test-login.sh --project .",
            "tdd green --project .",
            "tdd reopen --reason the-test-named-the-wrong-error --project .",
        ],
    ),
    ("ledger-register-plan", &["ledger-register-plan .", "ledger-register-apply --plan-digest <digest> --permission-reference <ref> ."]),
    ("ledger-register-apply", &["ledger-register-apply --plan-digest <digest> --permission-reference <ref> ."]),
    ("card-import-plan", &["card-import-plan --from github --repo owner/repo .", "card-import-apply --from github --repo owner/repo --plan-digest <digest> --permission-reference <ref> ."]),
    ("card-import-apply", &["card-import-apply --from github --repo owner/repo --plan-digest <digest> --permission-reference <ref> ."]),
    ("card-summary-plan", &["card-summary-plan --card C-1 --to github .", "card-summary-apply --card C-1 --to github --plan-digest <digest> --permission-reference <ref> ."]),
    ("card-summary-apply", &["card-summary-apply --card C-1 --to github --plan-digest <digest> --permission-reference <ref> ."]),
    ("ledger-recover-plan", &["ledger-recover-plan --source <copy> .", "ledger-recover-apply --source <copy> --plan-digest <digest> --permission-reference <ref> ."]),
    ("ledger-recover-apply", &["ledger-recover-apply --source <copy> --plan-digest <digest> --permission-reference <ref> ."]),
    ("update", &["update"]),
    ("go-state-plan", &["go-state-plan .", "go-state-apply --plan-digest <digest> --permission-reference <ref> ."]),
    ("go-state-apply", &["go-state-apply --plan-digest <digest> --permission-reference <ref> ."]),
    ("go-state-revert-plan", &["go-state-revert-plan --backup <UTC> ."]),
    ("go-state-revert", &["go-state-revert --backup <UTC> --plan-digest <digest> --permission-reference <ref> ."]),
    ("home-state-plan", &["home-state-plan", "home-state-apply --now 2026-11-14T00:00:00Z --plan-digest <digest> --permission-reference <ref>"]),
    ("home-state-apply", &["home-state-apply --now 2026-11-14T00:00:00Z --plan-digest <digest> --permission-reference <ref>"]),
    ("home-state-revert", &["home-state-revert --permission-reference <ref>"]),
    ("home-retire-plan", &["home-retire-plan", "home-retire-apply --now 2026-11-14T00:00:00Z --plan-digest <digest> --permission-reference <ref>"]),
    ("home-retire-apply", &["home-retire-apply --now 2026-11-14T00:00:00Z --plan-digest <digest> --permission-reference <ref>"]),
    ("home-retire-revert", &["home-retire-revert --permission-reference <ref>"]),
    ("home-install-plan", &["home-install-plan", "home-install-apply --home <HOME> --now 2026-11-14T00:00:00Z --plan-digest <digest> --permission-reference <ref>"]),
    ("home-install-apply", &["home-install-apply --home <HOME> --now 2026-11-14T00:00:00Z --plan-digest <digest> --permission-reference <ref>"]),
    ("home-install-revert", &["home-install-revert --home <HOME> --permission-reference <ref>"]),
    ("worktree-move-plan",&["worktree-move-plan .", "worktree-move-apply --plan-digest <digest> --permission-reference <ref> ."]),
    ("worktree-move-apply", &["worktree-move-apply --plan-digest <digest> --permission-reference <ref> ."]),
    ("update-plan", &["update-plan ~ <stage>", "update-apply --plan-digest <digest> --permission-reference <ref> ~ <stage> < request.json"]),
    ("update-apply", &["update-apply --plan-digest <digest> --permission-reference <ref> ~ <stage> < request.json"]),
    ("reports-move-plan", &["reports-move-plan .", "reports-move-apply --plan-digest <digest> --permission-reference <ref> . < plan.json"]),
    ("reports-move-apply", &["reports-move-apply --plan-digest <digest> --permission-reference <ref> . < plan.json"]),
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
    ("key-set", &["key-set glm < key.txt", "key-set glm --prompt"]),
    ("key-status", &["key-status glm"]),
    ("key-remove", &["key-remove glm"]),
    ("key-import-plan", &["key-import-plan glm", "key-import-apply --plan-digest <digest> --permission-reference <ref> glm"]),
    ("key-import-apply", &["key-import-apply --plan-digest <digest> --permission-reference <ref> glm"]),
    ("keys-revert", &["keys-revert --receipt <keys-id> --permission-reference <ref> ~"]),
    ("models-revert", &["models-revert --receipt <models-id> --permission-reference <ref> ~"]),
    ("claude-doctor", &["claude-doctor ."]),
    ("codex-doctor", &["codex-doctor ."]),
    ("install-inspect", &["install-inspect"]),
    ("identity", &["identity ."]),
    ("worktree-report", &["worktree-report ."]),
    ("optimize-plan", &["optimize-plan --integration-ref origin/main ."]),
    ("baseline-inventory", &["baseline-inventory ."]),
    ("claim-acquire", &["claim-acquire . < request.json"]),
    ("claim-verify", &["claim-verify . < request.json"]),
    (
        "claim-timeout-check",
        &["claim-timeout-check --issue 7 --claim 101 --request 202 --lane agent-1"],
    ),
    ("claim-expired", &["claim-expired ."]),
    ("audit-claim-stale", &["audit-claim-stale --lane agent-1 ."]),
    ("claim-show", &["claim-show --task <task> ."]),
    ("claim-release", &["claim-release --task <task> --epoch 1 ."]),
    ("claim-handoff", &["claim-handoff --task <task> --from-epoch 1 --reason <reason> ."]),
    ("claim-link", &["claim-link --issue 7 --comment 101 --lane agent-1 ."]),
    ("queue-free", &["queue-free . < request.json"]),
    ("queue-query", &["queue-query . < request.json"]),
    ("leader-tick", &["leader-tick . < request.json"]),
    (
        "leader-pulse",
        &["leader-pulse --terminal term_abc --text continue-round --interval-secs 300"],
    ),
    (
        "leader-round",
        &[
            "leader-round --observations screens.json --repo modu-ai/moai-adk .",
            "leader-round --observations screens.json --snapshot fleet.json .",
        ],
    ),
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
    (
        "lane-screen",
        &[
            "lane-screen < screen.txt",
            "lane-screen --home <dir> < screen.json",
        ],
    ),
    (
        "lane-send",
        &[
            "lane-send --terminal term_123 --text <one-line-text>",
            "lane-send --terminal term_123 --text <one-line-text> --enter",
        ],
    ),
    ("lane-handoff-check", &["lane-handoff-check . < request.json"]),
    ("lane-safe-to-reset", &["lane-safe-to-reset --lane agent-1 ."]),
    ("close-check", &["close-check 2903 --repo modu-ai/moai-adk"]),
    ("card-times", &["card-times < request.json"]),
    (
        "session-register",
        &[
            "session-register --name agent-1 .",
            "session-register --show --name agent-1 .",
            "session-register --release --by leader --name agent-1 --instance <instance> .",
        ],
    ),
    ("session-current", &["session-current"]),
    ("usage-limit", &["usage-limit ~"]),
    ("heavy-slot-acquire", &["heavy-slot-acquire --name agent-1 -- cargo test"]),
    ("heavy-slot-status", &["heavy-slot-status"]),
    ("goal-show", &["goal-show ."]),
    ("goal-plan", &["goal-plan --request <request> .", "goal-apply --request <request> --plan-digest <digest> ."]),
    ("goal-apply", &["goal-apply --request <request> --plan-digest <digest> ."]),
    ("goal-block", &["goal-block < request.json"]),
    ("gate-check", &["gate-check < request.json"]),
    ("gate-merge", &["gate-merge --pr 2824 --permission-reference <ref>"]),
    ("gate-loop", &["gate-loop --permission-reference <ref>"]),
    ("gate-schedule-plan", &["gate-schedule-plan ~ --os macos", "gate-schedule-apply ~ --os macos --plan-digest <digest> --permission-reference <ref>"]),
    ("gate-schedule-apply", &["gate-schedule-apply ~ --os macos --plan-digest <digest> --permission-reference <ref>"]),
    ("wait", &["wait --repo modu-ai/moai-adk ci"]),
    ("scope-check", &["scope-check ."]),
    ("worktree-plan", &["worktree-plan --branch-prefix issue/ .", "worktree-apply --branch-prefix issue/ --plan-digest <digest> ."]),
    ("worktree-apply", &["worktree-apply --branch-prefix issue/ --plan-digest <digest> ."]),
    ("pr-body-render", &["pr-body-render --receipts <receipts> --base-sha <sha> --base-branch <base-branch> --default-branch <default-branch> ."]),
    ("pr-confirm", &["pr-confirm --task <task> --operation-id <operation-id> --observed-state <observed-state> ."]),
    ("audit-claim", &["audit-claim < request.json"]),
    ("audit-verdict-line", &["audit-verdict-line --pr 2850 --head <head> --verdict PASS --title <scope>", "audit-verdict-line --pr 2850 --head <head> --verdict FAIL --title <scope> --base <earlier-head>"]),
    ("trailers", &["trailers --spec SPEC-20261117-session-trailer"]),
    ("trailer-check", &["trailer-check origin/v2026.11.0..HEAD"]),
    ("bundle", &["bundle check C-1 C-2", "bundle build C-1 C-2"]),
    (
        "audit-gate",
        &["audit-gate < input.json", "audit-gate --require-pass < input.json"],
    ),
    ("audit-clone-plan", &["audit-clone-plan --pr 2824 --head <sha> --head-branch <head-branch> --base-branch <base-branch> ."]),
    ("labels-check", &["labels-check --repo modu-ai/moai-adk ."]),
    ("decision-due", &["decision-due docs/decisions/bundle.md"]),
    ("decision-file-check", &["decision-file-check docs/decisions/bundle.md"]),
    ("decision-append", &["decision-append --log docs/decisions-log.md . < request.json"]),
    ("flow-wip-check", &["flow-wip-check --owner modu-ai --repo modu-ai/moai-adk --lane agent-1"]),
    ("dispatch-record", &["dispatch-record --permission-reference <ref> . < request.json"]),
    ("dispatch-ack", &["dispatch-ack --permission-reference <ref> . < request.json"]),
    ("dispatch-start", &["dispatch-start --permission-reference <ref> . < request.json"]),
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
    ("worktree-install-plan", &["worktree-install-plan --scope web ."]),
    ("hygiene-plan", &["hygiene-plan --integration-ref origin/main ."]),
    ("hygiene-apply", &["hygiene-apply --integration-ref origin/main --plan-digest <digest> --permission-reference <ref> ."]),
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
        "Only `claude`; other hosts start with `cc|glm|codex|omp -l|-a`",
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
const LAUNCHERS: [&str; 4] = ["cc", "codex", "glm", "omp"];
const LAUNCHER_LINE: (&str, &str) = (
    "cc, codex, glm, omp",
    "Start a Claude Code, Codex, GLM or OMP session",
);

/// The Korean one-line about of each command the root help shows (S6d, #2825; design 8). The
/// other commands keep their English about in `help --all` until S6e moves the command help.
/// `the_korean_abouts_cover_exactly_the_root_commands` keeps this table in step with [`GROUPS`].
pub const ABOUT_KO: &[(&str, &str)] = &[
    ("project-detect", "프로젝트의 사실 알아내기"),
    ("project-plan", "프로젝트 프로필 미리 보기"),
    ("project-apply", "미리 본 프로젝트 프로필 쓰기"),
    ("work-flow-plan", "통합 가지 고르기의 근거 보기"),
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
const LAUNCHER_LINE_KO: &str = "Claude Code·Codex·GLM·OMP 세션 시작";
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

/// The name help, usage lines, footers and usage-error text print: always `moai`, whichever
/// name was typed (`moai`, `moai-cli`, a renamed copy; #3004 P1). `moai-cli` stays an alias of
/// the executable and the `--version` signature, but a person reads and types `moai`. The
/// `argv[0]` argument stays so every caller goes through this one place when the rule changes.
pub fn bin_name(_argv0: &str) -> &'static str {
    "moai"
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
    let rows = group_rows(cmd, all, lang);
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

/// The first steps the root help names, in the order a new person takes them (#3004 P1): the
/// command after the binary name and what it does. A command starting with `/` is typed inside
/// the session, not to the shell. `the_getting_started_and_example_commands_exist` (in
/// `main.rs`) parses every one against the real command tree, so a removed command breaks it.
/// `update` comes before the session (SPEC-20261117-installer-skill-agents REQ-11): it installs
/// the skill that `/moai project` needs.
pub const GETTING_STARTED: &[(&str, &str)] = &[
    ("install-inspect", "Check the installed MoAI"),
    ("update", "Install or update the MoAI skill and agents"),
    ("cc", "Start a session in this folder (or: codex, glm, omp)"),
    (
        "/moai project",
        "In the session: set up this project ($moai in Codex)",
    ),
];

/// A few commands to copy, each with what it does (#3004 P1); parsed against the real command
/// tree by the same test.
pub const ROOT_EXAMPLES: &[(&str, &str)] = &[
    // SPEC-20261118-launcher-agent-entry: the role is entered inside the session; at 60
    // columns this ends on column 59, so it never wraps.
    ("cc -l", "leader, then /moai --mode-leader"),
    ("update", "preview and apply an update"),
    ("project-detect .", "show the facts of this project"),
];

/// Where the root help sends a reader for more: the word after the binary name and why.
const ROOT_MORE: &[(&str, &str)] = &[
    ("help --all", "every command"),
    ("<command> --help", "options and examples of one command"),
];

/// One row of the root help: `lead` spaces, the left cell (`plain` is what it reads, `styled`
/// the same with style codes), then the about in the column every block shares.
struct Row {
    lead: usize,
    plain: String,
    styled: String,
    about: String,
}

/// The root help body for English (#3004 P1): the first steps, the grouped commands, a few
/// examples and where the rest is, one block each with a blank line between. The about column
/// is shared by every block, so the eye finds it once. `color` adds style codes to the
/// headings and command names; the words and the line breaks are the same without it.
pub fn render_root(cmd: &Command, bin: &str, width: usize, color: bool) -> String {
    let painter = Painter {
        color,
        unicode: false,
        lang: Lang::En,
    };
    let command_row = |text: &str, about: &str| {
        let full = format!("{bin} {text}");
        Row {
            lead: 2,
            styled: painter.paint(Token::Code, &full),
            plain: full,
            about: about.to_string(),
        }
    };
    let mut steps = Vec::new();
    for (i, (text, about)) in GETTING_STARTED.iter().enumerate() {
        let mut row = if text.starts_with('/') {
            Row {
                lead: 2,
                styled: painter.paint(Token::Code, text),
                plain: text.to_string(),
                about: about.to_string(),
            }
        } else {
            command_row(text, about)
        };
        let number = format!("{}. ", i + 1);
        row.plain.insert_str(0, &number);
        row.styled.insert_str(0, &number);
        steps.push(row);
    }
    let mut commands = Vec::new();
    for (heading, list) in group_rows(cmd, false, Lang::En) {
        // A group heading is a row without an about: its own line, one level in.
        commands.push(Row {
            lead: 2,
            styled: painter.paint(Token::Accent, heading),
            plain: heading.to_string(),
            about: String::new(),
        });
        for (name, about) in list {
            commands.push(Row {
                lead: 4,
                styled: painter.paint(Token::Code, &name),
                plain: name,
                about,
            });
        }
    }
    let examples: Vec<Row> = ROOT_EXAMPLES
        .iter()
        .map(|(text, about)| command_row(text, about))
        .collect();
    let more: Vec<Row> = ROOT_MORE
        .iter()
        .map(|(text, about)| command_row(text, about))
        .collect();
    let blocks = [
        ("Getting started", steps),
        ("Commands", commands),
        ("Examples", examples),
        ("More", more),
    ];
    // The about column starts two cells after the widest left cell of any row that has one.
    let column = blocks
        .iter()
        .flat_map(|(_, rows)| rows.iter())
        .filter(|r| !r.about.is_empty())
        .map(|r| r.lead + r.plain.width())
        .max()
        .unwrap_or(0)
        + 2;
    let mut out = String::new();
    for (i, (heading, rows)) in blocks.iter().enumerate() {
        if i > 0 {
            out.push('\n');
        }
        out.push_str(&painter.paint(Token::Strong, heading));
        out.push('\n');
        for r in rows {
            let mut line = format!("{}{}", " ".repeat(r.lead), r.styled);
            if !r.about.is_empty() {
                line.push_str(&" ".repeat(column - r.lead - r.plain.width()));
                line.push_str(&wrap(&r.about, column, width));
            }
            out.push_str(line.trim_end());
            out.push('\n');
        }
    }
    out
}

/// The `(name, about)` rows under each group heading in `lang`; `all` lists every command,
/// otherwise only the root ones (the launchers on one line). A group with no row is left out.
fn group_rows(cmd: &Command, all: bool, lang: Lang) -> Vec<(&'static str, Vec<(String, String)>)> {
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
    rows
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
    // Colour and bold follow stdout like the rest of the human layer (#2803): off on a pipe,
    // `NO_COLOR` and `TERM=dumb`, on a terminal; `--color` wins when it was given.
    let color = Painter::for_stream(crate::ui::Stream::Stdout, crate::ui::ColorFlag::Auto).color;
    configure_styled(cmd, bin, version, all, lang, color)
}

/// [`configure`] in `lang` (S6d, #2825), without style codes. The first line names the product
/// and stays English (design 1, like the banner's identity line).
pub fn configure_in(
    cmd: Command,
    bin: &'static str,
    version: &str,
    all: bool,
    lang: Lang,
) -> Command {
    configure_styled(cmd, bin, version, all, lang, false)
}

/// [`configure_in`] with the root help's style codes when `color` is set (#3004 P1). Only the
/// English root help carries style: `help --all` is printed as plain text, and the Korean help
/// (unused while the CLI's human output is English only, user decision 6878265192) keeps its
/// earlier layout.
pub fn configure_styled(
    cmd: Command,
    bin: &'static str,
    version: &str,
    all: bool,
    lang: Lang,
    color: bool,
) -> Command {
    // The English root help (#3004 P1) is its own layout: first steps, the grouped commands,
    // examples and pointers. `help --all` and the Korean help keep the grouped list and its
    // footer.
    let root_en = lang == Lang::En && !all;
    let groups = if root_en {
        render_root(&cmd, bin, width(), color)
    } else {
        render_groups_in(&cmd, all, width(), lang)
    };
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
        "{bin} <command> [options]\n{}{bin} cc|codex|glm|omp -l|-a    {LEAD_OR_AGENT}",
        " ".repeat("Usage:".width() + 1)
    );
    let usage_lines = match lang {
        Lang::En => "{usage-heading} {usage}".to_string(),
        Lang::Ko => format!(
            "{USAGE_KO} {bin} <command> [options]\n{}{bin} cc|codex|glm|omp -l|-a    {LEAD_OR_AGENT_KO}",
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
    // The English root help ends with its own "More" block instead of the footer lines.
    let template = if root_en {
        format!("MoAI-ADK {version} — agentic development kit\n\n{usage_lines}\n\n{groups}")
    } else {
        format!(
            "MoAI-ADK {version} — agentic development kit\n\n{usage_lines}\n\n{groups}{options}\n{footer}\n"
        )
    };
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
pub const KOREAN_GROUPS: &[&str] = &["Start", "Project", "Update and move", "Keys", "Diagnose"];

/// The Korean long about of a moved command that has an English one (S6e-2). Wrapped by
/// [`wrap`] at the help width before clap sees it: clap counts a Hangul letter as one cell.
const LONG_ABOUT_KO: &[(&str, &str)] = &[
    (
        "work-flow-plan",
        "`work_flow` 인터뷰 단계의 통합 가지 질문 근거를 읽기 전용으로 내보내요. 후보 가지마다 \
         `canonical` 플래그를 붙이고, 원격의 풀 요청 후보 이름공간(예: `refs/remotes/origin/pr/<n>`)은 \
         `canonical: false` 와 candidate 라벨이 붙어 보이기만 하고 통합 대상으로 추천되지 않아요. 추천은 \
         값 하나: 프로필에 `work_flow.integration_branch` 가 기록돼 있으면 다시 실행에서도 그 값을 그대로 \
         추천하고(이유 문구가 recorded 라고 말해요), 없으면 `origin/HEAD` 가 가리키는 원격 기본 가지, \
         그것도 없으면 현재 가지를 추천해요. 기록된 대상이 저장소에 없으면 `missing-target`, 다른 \
         워크트리가 이미 체크아웃 중이면 `occupied-worktree` 상태와 안전한 선택지를 내보내요. 가지·\
         워크트리·파일·ref 를 하나도 만들거나 옮기거나 지우지 않고, 네트워크 명령도 돌리지 않아요.",
    ),
    (
        "project-detect",
        "`/moai project` 경로가 쓰는 프로젝트의 사실을 알아내요: 새 프로젝트인지 기존 \
         프로젝트인지, manifest, 언어, 언어 서버, 설정된 MCP 서버 이름, 데이터베이스. 읽기만 하고 \
         아무것도 쓰지 않아요. `project-plan` 에 넘길, `detected` 가 채워진 프로필을 돌려줘요.",
    ),
    (
        "language-setup",
        "쓰는 언어를 항목마다 정해요: 항목마다 질문을 하나씩 하고 `project-plan` 에 넘길 \
         완전한 프로필을 돌려줘요. 아무것도 쓰지 않아요. `--questions` 는 호스트 질문 도구용 질문 \
         묶음을 찍고, `--answers` 는 그 도구의 JSON 답을 stdin 에서 읽어요. 두 플래그가 없으면 \
         터미널에서 물어요(질문은 stderr, 답은 stdin).",
    ),
    (
        "index-build",
        "증분 색인의 새 세대를 만들어요: 추적 파일을 훑고, 바뀌지 않는 manifest 폴더를 \
         준비한 뒤 원자적 포인터 교체로 게시해요. 빌드가 실패하면 앞 세대를 그대로 두고 partial 로 \
         표시해요(REQ-C6-01).",
    ),
    (
        "update",
        "이 실행 파일과 함께 설치된 패키지로 스킬을 업데이트해요: update-plan 처럼 미리 \
         보여 주고, 확인하기 전에는 아무것도 쓰지 않아요 — 터미널에서는 한 번 묻고, 그 밖에서는 \
         미리 보기의 digest 를 --confirm 으로 넘겨요. HOME 은 환경에서 읽어요.",
    ),
    (
        "go-state-plan",
        "PROJECT 의 Go 세대 상태를 읽기만 하고 미리 봐요: Go manifest 가 도구가 만들고 \
         바뀌지 않았다고 증명하는 파일은 무엇인지, 나머지는 왜 그대로 남는지. 아무것도 쓰지 \
         않아요.",
    ),
    (
        "go-state-apply",
        "확인한 go-state 미리 보기를 실행해요: Go `.moai/` 는 `.moai-backup`(또는 \
         `.moai-backup-<UTC>`)이 되고, 새 `.moai/` 가 앞서의 `.moai-cli` 제품 상태를 받고 표식은 \
         마지막에 써요. 남는 Go 폴더는 파일 이동 단계를 그대로 해요. 새로 계산한 미리 보기의 \
         digest 가 같을 때만 실행하고, 절대 덮어쓰지 않아요.",
    ),
    (
        "go-state-revert-plan",
        "go-state 실행 하나를 되돌리면 무엇을 하는지 읽기만 해요.",
    ),
    (
        "go-state-revert",
        "go-state 실행 하나를 되돌려요: 새 `.moai/` 는 `.moai.rust-<UTC>` 로 비켜 \
         두고, Go 백업이 다시 `.moai` 가 돼요. 전부 아니면 전무: 원래 경로에 무언가 다시 생겼거나 \
         백업이 바뀌었으면 되돌리기를 거절해요.",
    ),
    (
        "home-state-plan",
        "HOME 상태 폴더를 미리 봐요: 프로젝트 판정 표와 HOME 의 Go 표식으로 `~/.moai` 를 \
         나누고, 백업 이름(`~/.moai-backup`, 이미 있으면 `~/.moai-backup-<UTC>` — 절대 덮어쓰지 \
         않음)을 정하고, 이동을 막는 것(살아 있는 Go 세션, 안에서 도는 프로세스, 안쪽 체크아웃, \
         확인하기에 너무 큼)과 사용자가 다시 복사해 올 수 있는 항목(기본 없음, 비밀은 절대 제안 \
         안 함)을 보여 줘요. 아무것도 쓰지 않아요.",
    ),
    (
        "home-state-apply",
        "Go `~/.moai` 를 미리 보기가 정한 백업으로 비켜 두고(절대 덮어쓰지 않음) Rust `~/.moai` 를 \
         만들고, `--copy` 로 이름 댄 항목만 복사하고(비밀·Go 표식은 절대 안 함), 영수증을 쓴 뒤 Rust \
         표식을 써요. 같은 입력으로 다시 계산한 미리 보기의 `plan_digest` 가 같고 동작이 \
         `migrate` 일 때만 실행해요. 실제 HOME 에서는 사용자가 확인한 뒤에만 리더가 실행해요.",
    ),
    (
        "home-state-revert",
        "Go `~/.moai` 를 되돌려 놓아요: 실행이 만든 Rust `~/.moai`(올바른 표식과 영수증)는 \
         `~/.moai.rust-<UTC>` 로 비켜 두고, 영수증이 이름 댄 백업이 `~/.moai` 로 돌아와요. 아무것도 \
         지우지 않아요.",
    ),
    (
        "home-retire-plan",
        "`~/.moai` 안 Go 세대 항목의 은퇴를 미리 봐요: HOME 이름 표로 맨 위 항목마다 \
         `retire`(`~/.moai-go-retired/<UTC>/` 로 이동)·`keep`(사용자 자료, 주인 모름, 표에 없는 \
         이름, 이 세대 이름, 병합 단계를 기다리는 키 파일, 링크)·`skip-in-use`(돌고 있는 \
         프로그램이 씀)로 나눠요. 비밀은 이름만 보고 절대 열지 않아요. 이동을 막는 것과 적용에 \
         필요한 `plan_digest` 를 보여 줘요. 아무것도 쓰지 않아요.",
    ),
    (
        "home-retire-apply",
        "미리 보기가 `retire` 로 정한 항목을 `~/.moai-go-retired/<UTC>/` 로 옮겨요: 먼저 \
         매니페스트와 이동 저널을 쓰고, 항목마다 rename 한 번(덮어쓰지 않음, 같은 파일 \
         시스템)이에요. 같은 입력으로 다시 계산한 미리 보기의 `plan_digest` 가 같고 동작이 \
         `retire` 일 때만 실행해요. 복사도 삭제도 없고, 돌고 있는 프로그램이 쥔 항목은 건너뛰어요. \
         실제 HOME 에서는 사용자가 확인한 뒤에만 리더가 실행해요.",
    ),
    (
        "home-retire-revert",
        "은퇴 저널을 거꾸로 되짚어 항목을 되돌려 놓아요: 항목은 옮긴 그 항목이 \
         그대로이고 옛 자리가 비어 있을 때만 돌아가요. 그 사이 바뀐 항목은 그 자리에 두고 \
         상태가 `revert_partial` 이에요. `--stamp` 가 없으면 가장 새 은퇴 저널을 써요. 아무것도 \
         지우지 않아요.",
    ),
    (
        "home-install-plan",
        "`~/.moai-cli` 의 이 세대 항목을 `~/.moai` 맨 위로 옮기는 일을 미리 봐요: \
         HOME 이름 표로 맨 위 항목마다 `move`(`~/.moai/<이름>` 으로 rename)·`keep`(사용자 자료, \
         옛 worktrees, 표에 없는 이름, 링크, 표식)·`blocked`(자리가 이미 있음)로 나눠요. Go \
         항목이 아직 은퇴 전이면 막혀요. 버전 폴더마다 이동 전후 설치 id, `path` 항목을 어디로 \
         다시 가리킬지, vault 의 모드를 함께 보여 줘요(이름·모드만 보고 열지 않아요). `--project` \
         클론에 등록된 Rust 세션이 몇 개인지 보여 주고(적용은 있으면 거부해요), 다른 클론은 HOME \
         에서 보이지 않는다고 알려 줘요. 아무것도 쓰지 않고 확인 토큰을 줘요.",
    ),
    (
        "home-install-apply",
        "미리 보기가 `move` 로 정한 `~/.moai-cli` 항목을 `~/.moai` 로 옮겨요: 먼저 \
         매니페스트와 이동 저널을 쓰고, 항목마다 rename 한 번(덮어쓰지 않음, 같은 파일 시스템, \
         `versions` 와 `path` 는 맨 마지막)이에요. 다음에 `path` 의 링크를 옮겨진 `versions` 로 \
         다시 가리키고, vault 는 이름·모드만 확인하고(열지도 모드를 바꾸지도 않아요), schema 2 \
         표식을 맨 마지막에 써서 모든 읽기가 `~/.moai` 로 가는 순간이 전부 끝난 뒤가 되게 해요. \
         같은 입력으로 다시 계산한 미리 보기의 `plan_digest` 가 같고 동작이 `install` 이며 \
         `--project` 클론에 등록된 Rust 세션이 없을 때만 실행해요. 돌고 있는 프로그램이 쥔 항목이 \
         있으면 옮긴 것을 되돌려 `rolled_back` 으로 끝나요. 중간에 멈추면 같은 명령이 이어서 \
         끝내요. 복사도 삭제도 없어요. 실제 HOME 에서는 사용자가 확인한 뒤에만 리더가 실행해요.",
    ),
    (
        "home-install-revert",
        "이동을 되돌려 놓아요: 적용이 쓴 표식(고쳐졌으면 거부)을 `~/.moai-cli/home-install/` \
         로 치우고, `path` 의 링크는 아직 새 대상을 가리킬 때만 옛 대상으로 돌리고, 이동 저널을 \
         거꾸로 되짚어요. 항목은 옮긴 그 항목이 그대로이고 옛 자리가 비어 있을 때만 돌아가요. \
         `~/.moai` 는 적용이 만들었고 비어 있을 때만 없애요. `--stamp` 가 없으면 가장 새 설치를 \
         되돌려요. 그 밖에는 아무것도 지우지 않아요.",
    ),
    (
        "ledger-status",
        "프로젝트의 로컬 SQLite 원장 상태를 읽기만 하고 보여 줘요: 있는지, 스키마 버전, WAL \
         모드가 아닌지, 카드가 상태별로 몇 장인지. 원장이 없는 것은 정상 답(`exists: false`)이고 \
         폴더도 파일도 만들지 않아요. 쓸 수 없는 파일은 `problem` 코드로 알리고 고치지 않아요.",
    ),
    (
        "card",
        "프로젝트의 로컬 카드를 읽고 써요: `list`(목록), `show C-n`(카드 하나와 기준·메모·링크), \
         `add`(draft 카드 하나와 기준을 한 트랜잭션으로), `ready`(draft 를 ready 로), `take`·`release`·`handoff`(예약을 잡고, 돌려주고, 넘겨요). `list`·`show` 는 읽기만 하고, 원장 파일과 \
         폴더는 어떤 동작도 만들지 않아요(`ledger-register-apply` 만 만들어요). 원장이 없으면 \
         `list` 는 빈 목록, 나머지는 `ledger not found` 오류와 안내를 줘요.",
    ),
    (
        "ledger-register-plan",
        "프로젝트의 로컬 원장을 `<HOME>/.moai/projects/<project_id>/` 에 만들면 무엇이 생기는지 \
         미리 봐요. 프로젝트 id 는 프로파일의 `project_id` 예요. 이미 있는 파일은 읽기만 \
         하고 바꾸지 않아요. 아무것도 쓰지 않고, 답에 `plan_digest` 가 있어요.",
    ),
    (
        "ledger-register-apply",
        "미리 본 프로젝트 원장을 만들어요 — 같은 입력으로 다시 계산한 계획의 `plan_digest` 가 \
         같을 때만, 원장 폴더와 파일을 만드는 유일한 명령이에요. Go 세대의 `~/.moai` 에는 만들지 \
         않아요. 이미 있는 유효한 원장은 그대로 두고, 이 프로젝트 것이 아닌 파일은 대체하지 않아요.",
    ),
    (
        "card-import-plan",
        "코드 호스트의 열린 이슈를 로컬 원장의 카드로 가져오면 무엇이 되는지 미리 봐요 — \
         모든 페이지를 읽고(끝까지 못 읽으면 오류), 이슈마다 만들어질 카드와 읽어 낸 합격 기준, \
         표를 읽지 못한 이슈, 남길 구성원 댓글과 걸러 낼 제어 줄을 보여 줘요. 이슈 글은 글로만 \
         다루고 명령으로 실행하지 않아요. 읽기(GET)만 하고 원장도 폴더도 만들지 않으며, 답에 \
         `plan_digest` 가 있어요.",
    ),
    (
        "card-import-apply",
        "미리 본 열린 이슈를 로컬 원장에 카드(draft)로 써요 — 호스트가 지금 주는 이슈 집합이 \
         미리 보기의 `plan_digest` 와 같을 때만, 한 트랜잭션으로 전부 쓰거나 아무것도 쓰지 않아요. \
         이미 카드가 있는 이슈는 그대로 두고, 원장이 없으면 호스트도 읽지 않고 멈춰요. 이슈를 \
         닫거나 바꾸지 않아요.",
    ),
    (
        "card-summary-plan",
        "카드 한 장의 요약을 그 카드의 이슈(또는 `--number` 로 정한 PR)에 올리면 무엇이 올라가는지 \
         미리 봐요 — 올라갈 글(올라가는 바이트 그대로, 제어 줄·멘션·비밀 모양을 씻은 뒤), \
         출처, 대상, 두 번 올라가지 않게 하는 표지, `plan_digest` 를 보여 줘요. 카드는 `--to` 로 적은 호스트에서 온 것이어야 해요. 로컬 원장만 읽고 통신도 쓰기도 \
         하지 않아요.",
    ),
    (
        "card-summary-apply",
        "미리 본 카드 요약을 이슈에 한 번 올려요 — 미리 보기의 `plan_digest` 가 그대로이고 \
         권한 참조가 있을 때만, 호스트에 쓰는 유일한 명령이에요. 올리기 전에 호스트를 읽어 구성원 또는 \
         토큰 계정의 댓글에 이 카드의 표지가 이미 있으면 아무것도 올리지 않고 `already_posted` 로 답해요.",
    ),
    (
        "ledger-recover-plan",
        "망가진 프로젝트 원장을 읽을 수 있는 사본에서 다시 만들면 무엇이 생기는지 미리 봐요. \
         사본은 `--source` 로 넘겨요(원장 파일 자체, `.bak`, 또는 원장 모양의 파일). 표마다 읽다가 \
         오류를 만나면 마지막 좋은 행에서 멈추고 잃은 행 수를 알려 줘요. 예약은 클레임 저널의 \
         lease 가 살아 있고, worktree 가 그 가지에 있고, 원격 관찰 파일이 그 가지를 보여 줄 때만 \
         `held` 로 되살리고, 나머지 활성 예약은 `held_for_review` 로 남겨 카드를 새로 잡지 못하게 \
         해요. 답에 `plan_digest` 가 있어요. 아무것도 쓰지 않아요(사본에도). 망가진 파일은 지우지도 \
         덮어쓰지도 않아요.",
    ),
    (
        "ledger-recover-apply",
        "미리 본 원장 복구를 해요 — 같은 입력으로 다시 계산한 계획의 `plan_digest` 가 같고 \
         권한 참조가 있을 때만. 새 원장을 `<원장>.recovering` 에 만들어 `integrity_check`·\
         `foreign_key_check` 를 통과시킨 뒤, 망가진 파일을 `<원장>.damaged-<UTC>` 로 남기고(하드 \
         링크라 원래 이름이 빈 적이 없어요) 새 파일을 원래 이름으로 바꿔 놓아요. 아무것도 지우지 \
         않아요. 같은 계획을 다시 하면 `unchanged` 이고 아무것도 쓰지 않아요. 이 계획의 결과가 \
         아닌 쓸 수 있는 원장은 대체하지 않아요.",
    ),
    (
        "worktree-move-plan",
        "이 클론의 linked worktree 를 `<root>/<project>-<name>` 으로 옮기는 일을 미리 봐요 — \
         `worktree-plan` 이 새 worktree 를 만드는 자리예요. worktree 마다 `location` 과 \
         `move`·`keep`(이미 root 안, 앞서의 `~/.moai-cli/worktrees/<key>/` 또는 Go 세대의 \
         `~/.moai/worktrees/projects/` 안)·`blocked` 와 그 까닭을 보여 줘요: main worktree, 없음, \
         잠김, submodule, 커밋 안 한·추적 안 하는 파일, push 안 한 커밋, 그 가지의 stash, 안에서 \
         도는 프로세스, 붙잡힌 가지, 붙잡힘 검사를 안 함(`holds` 관측 없음). 아무것도 쓰지 않고, \
         답에 `plan_digest` 가 있어요.",
    ),
    (
        "worktree-move-apply",
        "`worktree-move-plan` 이 `move` 로 적은 worktree 를 옮겨요 — 같은 입력으로 다시 \
         계산한 계획의 `plan_digest` 가 같을 때만. 옮길 때마다 프로세스와 열린 파일을 다시 보고, \
         `git worktree move` 를 돌리고(`--force` 는 절대 안 함), git 이 새 경로에 같은 HEAD·가지로 \
         적었는지 확인해요(아니면 `git worktree repair` 한 번). 옮긴 worktree 마다 되돌리는 명령이 \
         함께 와요.",
    ),
    (
        "cc",
        "Claude Code 세션 실행기: `cc -l|-a`. `-l|-a` 는 세션 이름과 권한만 정하고, 역할은 세션 \
         안의 `/moai --mode-leader`·`/moai --mode-agent` 로 시작.",
    ),
    (
        "codex",
        "Codex CLI 세션 실행기: `codex -l|-a`. `-l|-a` 는 세션 이름과 권한만 정하고, 역할은 세션 \
         안의 `$moai --mode-leader`·`$moai --mode-agent` 로 시작. `codex -a` 의 첫 프롬프트는 \
         `$moai --mode-agent` 줄부터.",
    ),
    (
        "glm",
        "z.ai GLM 위의 Claude Code: `glm -l|-a`(환경의 ANTHROPIC_AUTH_TOKEN 이나 키 파일이 \
         필요해요). `-l|-a` 는 세션 이름과 권한만 정하고, 역할은 세션 안의 \
         `/moai --mode-leader`·`/moai --mode-agent` 로 시작.",
    ),
    (
        "omp",
        "Oh My Pi(OMP) 세션 실행기: `omp -l|-a`. `-l|-a` 는 세션 이름과 권한(`--approval-mode`)만 \
         정하고, 역할은 세션 안의 `/skill:moai --mode-leader`·`/skill:moai --mode-agent` 로 \
         시작. z.ai 모델이면 glm 키가 `ZAI_API_KEY` 로 OMP 에 가요(명령에는 없어요).",
    ),
    (
        "claude-doctor",
        "Claude Code 를 읽기만 하고 확인해요: PROJECT 에서 시작한 세션이 AGENTS.md 를 읽는지.",
    ),
    (
        "codex-doctor",
        "Codex 를 읽기만 하고 진단해요: AGENTS 사슬과 크기 한도, 겹친 스킬, `.codex/` 층.",
    ),
    (
        "install-inspect",
        "설치를 읽기만 하고 살펴봐요: PATH 의 세대, `~/.moai-cli/installations/` 의 \
         영수증, Go 세대 경로가 있는지. 아무것도 쓰지 않아요.",
    ),
    (
        "identity",
        "경로의 클론·작업 공간 정체와 상태 루트를 읽기만 하고 알려 줘요(stdout 에 typed JSON).",
    ),
    (
        "worktree-report",
        "저장소의 worktree, 그 빌드 폴더, 임시 `moai-*` 폴더가 쓰는 공간을 읽기만 하고 보고해요. \
         사실만 — 지울 후보는 판단하지 않아요. 아무것도 쓰지 않아요.",
    ),
    (
        "worktree-install-plan",
        "프로필 `scopes` 의 scope 하나가 고른 폴더만 설치하는 명령을 보여 주기만 해요. \
         명령을 실행하지 않고 아무것도 쓰지 않아요 — 패키지 설치는 네트워크로 받는 일이에요. \
         잠금 파일로 매니저를 고르고, 문서에 없는 것은 `unverified` 에 적어요.",
    ),
    (
        "optimize-plan",
        "비울 수 있는 공간을 읽기만 하고 미리 봐요: 병합 뒤의 linked worktree(squash 병합 \
         포함), 이름 댄 루트 아래의 cargo 빌드 폴더, 보고서 — 크기, 마지막 변경, 보관 기간, 남기는 \
         까닭, 되살리는 법. worktree 는 정확한 HEAD 에 대한 독립 감사 PASS 가 있을 때만 `remove` \
         후보이고, 살아 있는 claim·HOLD·관측 안 된 부분이 있으면 남겨요. 아무것도 지우지 않아요 — \
         적용은 사용자 확인 뒤 `worktree-gc-apply`.",
    ),
    (
        "baseline-inventory",
        "체크아웃의 기준 사실 목록을 읽기만 하고 만들어요(stdout 에 typed JSON).",
    ),
    (
        "key-set",
        "공급자 키를 암호화해 금고에 저장해요: 지금 프로젝트용, 또는 `--shared` 면 모든 \
         프로젝트용. 값은 stdin 에서 읽고(argv 로는 받지 않음) 대화형 터미널이면 거절해요. 다만 \
         `--prompt` 를 붙이면 터미널에서 화면에 보이지 않게 입력받아요. 답에는 값이 절대 들어가지 \
         않아요.",
    ),
    (
        "key-status",
        "공급자 항목(지금 프로젝트, 또는 `--shared`)이 있는지, 안전한지, 복호화되는지 확인해요. \
         값·길이·지문은 절대 보이지 않아요.",
    ),
    (
        "key-remove",
        "금고에서 공급자 키를 지워요: 지금 프로젝트의 항목, 또는 `--shared` 면 공유 항목. \
         `--yes` 가 없으면 지울 것만 알려 줘요. 기계 키와 옛 키 파일은 절대 건드리지 않아요.",
    ),
    (
        "key-import-plan",
        "옛 키 파일을 암호화 금고로 옮기는 일을 미리 봐요: 어느 파일, 어느 폴더로 — 값은 \
         절대 보이지 않아요. 아무것도 쓰지 않아요.",
    ),
    (
        "key-import-apply",
        "옛 키 파일의 키를 암호화 금고로 복사해요: 미리 보기의 digest 와 권한 참조가 \
         있을 때만, 다시 읽은 값이 같아야 하고, 옛 파일은 그대로 둬요. 목적지는 지금 프로젝트 폴더, \
         또는 `--shared` 면 `_shared`.",
    ),
];

/// The launchers' own abouts (S6e-6): the root list shows the three on one line instead.
const LAUNCHER_ABOUT_KO: &[(&str, &str)] = &[
    ("cc", "Claude Code 세션 시작"),
    ("codex", "Codex CLI 세션 시작"),
    ("glm", "z.ai GLM 으로 Claude Code 세션 시작"),
    ("omp", "Oh My Pi(OMP) 세션 시작"),
];

/// The Korean help of a moved command's argument (S6e-2): `(command, id, text)`, an empty
/// command for an argument id every command shares.
const ARG_HELP_KO: &[(&str, &str, &str)] = &[
    ("", "project", "프로젝트 루트"),
    (
        "cc|codex|glm|omp",
        "prompt",
        "첫 프롬프트에 덧붙일 글(codex 만)",
    ),
    (
        "cc|codex|glm|omp",
        "leader",
        "리더 세션(이름 `leader` 와 권한만 정함, 역할은 세션 안의 `/moai --mode-leader`, Codex 는 \
         `$moai --mode-leader`, OMP 는 `/skill:moai --mode-leader`)",
    ),
    (
        "cc|codex|glm|omp",
        "agent",
        "에이전트 세션(이름과 권한만 정함, 역할은 세션 안의 `/moai --mode-agent`, Codex 는 첫 \
         프롬프트의 `$moai --mode-agent`, OMP 는 `/skill:moai --mode-agent`)",
    ),
    (
        "cc|codex|glm|omp",
        "name",
        "세션 이름(기본: -l 은 `leader`, -a 는 프로젝트의 다음 `agent-N`, 역할 없으면 없음)",
    ),
    (
        "cc|codex|glm|omp",
        "model",
        "주 모델(Claude 호스트 기본 `opus`, codex·omp 는 정하지 않으면 자기 기본값)",
    ),
    (
        "cc|codex|glm|omp",
        "effort",
        "effort(Claude 기본 `medium`, codex·omp 는 지정할 때만 low|medium|high|xhigh|max, omp 는 `--thinking` 으로)",
    ),
    (
        "cc|codex|glm|omp",
        "advisor",
        "Claude 호스트의 advisor 모델(fable|opus|sonnet 또는 claude-* 전체 id). 기본 없음 — \
         사용자 설정 `launcher.advisor` 로 켤 수 있음. 사용량 크레딧이 청구될 수 있음",
    ),
    (
        "cc|codex|glm|omp",
        "no_advisor",
        "사용자 설정 `launcher.advisor` 가 켜도 `--advisor` 를 넘기지 않기",
    ),
    (
        "cc|codex|glm|omp",
        "permissions",
        "권한 질문 유지: `--dangerously-skip-permissions` 없음(Claude 호스트), codex 는 \
         `workspace-write` / `on-request` sandbox, omp 는 `--approval-mode always-ask-cached`. 없으면 \
         -l|-a 세션은 전체 접근(omp: `--approval-mode yolo`), 역할 없는 세션은 권한 플래그를 받지 않음",
    ),
    (
        "cc|codex|glm|omp",
        "dry_run",
        "호스트를 시작하는 대신 만든 명령을 JSON 으로 찍기(`--print-command` 도 같은 뜻)",
    ),
    (
        "codex-doctor",
        "project",
        "프로젝트 루트(Codex 는 여기서 작업 폴더까지 내려가며 읽음)",
    ),
    (
        "codex-doctor",
        "cwd",
        "프로젝트 안의 작업 폴더(기본: 프로젝트 루트)",
    ),
    (
        "codex-doctor",
        "codex_home",
        "전역 AGENTS 파일이 있는 Codex home(기본: $CODEX_HOME, 없으면 ~/.codex)",
    ),
    (
        "codex-doctor",
        "home",
        "`.agents/skills` 가 있는 home(기본: $HOME)",
    ),
    (
        "codex-doctor",
        "fallback",
        "`project_doc_fallback_filenames` 항목(되풀이 가능, config.toml 은 안 읽음)",
    ),
    (
        "codex-doctor",
        "max_bytes",
        "`project_doc_max_bytes`(기본 32768, config.toml 은 안 읽음)",
    ),
    (
        "install-inspect",
        "json",
        "REQ-04A-02 호출 형식용으로 받음 — 출력은 늘 JSON 봉투",
    ),
    ("identity", "path", "클론이나 worktree 안의 아무 경로"),
    (
        "worktree-report",
        "tmp_root",
        "`moai-*` 자식을 나눌 폴더(기본 `/tmp`)",
    ),
    (
        "worktree-report",
        "budget_ms",
        "크기 훑기의 시간 예산(밀리초). 넘으면 그 갈래는 `partial`",
    ),
    ("worktree-install-plan", "project", "저장소의 아무 체크아웃"),
    (
        "worktree-install-plan",
        "scope",
        "프로필 `scopes` 에 선언된 scope 이름",
    ),
    ("optimize-plan", "project", "저장소의 아무 체크아웃"),
    (
        "optimize-plan",
        "integration_ref",
        "통합 가지의 로컬 ref, 예: `origin/v2026.11.0`. 아무것도 fetch 하지 않음",
    ),
    (
        "optimize-plan",
        "recent_minutes",
        "이 분보다 최근에 바뀌면 그 줄을 남김(최소 60)",
    ),
    (
        "optimize-plan",
        "budget_ms",
        "시간 예산(밀리초). 넘으면 남은 worktree 는 `skip`, 크기는 partial",
    ),
    (
        "optimize-plan",
        "retention_days",
        "worktree 밖 빌드 폴더의 자식은 이 날수가 지나야 셈",
    ),
    (
        "optimize-plan",
        "build_roots",
        "자식 폴더가 cargo 빌드 폴더인 절대 폴더(되풀이 가능, 예: `/tmp/moai-rc-target-*` 용 \
         `/tmp`). 기본 없음",
    ),
    (
        "optimize-plan",
        "observations",
        "JSON 파일: `schema_version` 1 과 선택 `audits`·`claims`·`holds`(각각 source·items). \
         빠진 부분이 필요한 줄은 남김",
    ),
    (
        "baseline-inventory",
        "go_baseline_commit",
        "`.claude` 트리를 비교할 Go 세대 기준 커밋(ls-tree 로만 읽음)",
    ),
    (
        "baseline-inventory",
        "go_repo",
        "프로젝트에 기준 커밋이 없을 때 찾아볼 추가 저장소",
    ),
    ("update", "json", "터미널에서도 화면 대신 JSON 봉투 찍기"),
    (
        "update",
        "confirm",
        "확인하는 미리 보기의 `plan_digest` — 아직 같을 때만 실행",
    ),
    (
        "update",
        "confirm_cleanup",
        "미리 보기의 `cleanup.confirm_token`: 거기 적힌 Go 잔여물을 보관 폴더로. \
         --confirm 이 없으면 나머지는 미리 보기로 남음",
    ),
    (
        "update",
        "confirm_home",
        "미리 보기의 `home.confirm_token`: 거기 적힌 Go 세대 `~/.moai` 항목을 \
         `~/.moai-go-retired/<UTC>/` 로 옮김. --confirm 이 없으면 나머지는 미리 보기로 남고, \
         --confirm-cleanup 과 같이 못 씀",
    ),
    (
        "update",
        "confirm_switch",
        "미리 보기의 `switch.confirm_token`: 이름 전환 스크립트 `<상태 폴더>/path/moai` 를 \
         새 것으로 바꿈. 옛 파일은 옆에 `moai.before-<UTC>` 로 보관됨. --confirm 이 없으면 \
         나머지는 미리 보기로 남음",
    ),
    (
        "update",
        "revert_switch",
        "앞선 --confirm-switch 가 `moai.before-<STAMP>` 로 보관한 옛 전환 스크립트를 되돌림. \
         혼자 실행되며 --confirm 류 플래그와 같이 못 씀",
    ),
    (
        "update",
        "allow_older",
        "설치된 더 새 버전 위로 이 옛 빌드로 돌아감(없으면 막힘). 미리 보기가 호스트마다 \
         경고하고, 확인할 때도 이 플래그가 다시 필요함",
    ),
    (
        "go-state-plan",
        "project",
        "Go `.moai/manifest.json` 이 있는 프로젝트 루트",
    ),
    ("", "backup", "실행의 백업 표지(`YYYYMMDDTHHMMSSZ`)"),
    (
        "home-state-plan",
        "home",
        "읽을 HOME(기본 `$HOME`). 시험은 임시 폴더를 넘김",
    ),
    (
        "ledger-status",
        "home",
        "읽을 HOME(기본 `$HOME`). 시험은 임시 폴더를 넘김",
    ),
    (
        "ledger-register-plan",
        "home",
        "읽을 HOME(기본 `$HOME`). 시험은 임시 폴더를 넘김",
    ),
    (
        "ledger-register-apply",
        "home",
        "미리 보기가 읽은 것과 같은 HOME(기본 `$HOME`)",
    ),
    (
        "card-import-plan",
        "home",
        "원장을 찾을 HOME(기본 `$HOME`). 시험은 임시 폴더를 넘김",
    ),
    (
        "card-import-apply",
        "home",
        "미리 보기가 읽은 것과 같은 HOME(기본 `$HOME`)",
    ),
    (
        "card-summary-plan",
        "home",
        "원장을 찾을 HOME(기본 `$HOME`). 시험은 임시 폴더를 넘김",
    ),
    (
        "card-summary-apply",
        "home",
        "미리 보기가 읽은 것과 같은 HOME(기본 `$HOME`)",
    ),
    (
        "ledger-recover-plan",
        "home",
        "읽을 HOME(기본 `$HOME`). 시험은 임시 폴더를 넘김",
    ),
    (
        "ledger-recover-apply",
        "home",
        "미리 보기가 읽은 것과 같은 HOME(기본 `$HOME`)",
    ),
    (
        "ledger-recover-plan",
        "source",
        "읽을 수 있는 가장 좋은 원장 사본: 망가진 파일 자체, `.bak`, 또는 원장 모양의 파일. 링크·폴더·WAL \
         파일·다른 프로젝트의 파일은 거절",
    ),
    (
        "ledger-recover-apply",
        "source",
        "미리 보기가 읽은 것과 같은 사본",
    ),
    (
        "ledger-recover-plan",
        "journal",
        "클레임 저널 폴더(기본: 프로젝트의 것). 없거나 못 읽으면 저널 확인이 모자라 모든 활성 예약이 \
         `held_for_review`",
    ),
    (
        "ledger-recover-apply",
        "journal",
        "미리 보기가 읽은 것과 같은 저널 폴더",
    ),
    (
        "ledger-recover-plan",
        "observations",
        "JSON 파일: 선택 `journal`(holder·generation·live)·`worktrees`(path·exists·branch·clean)·\
         `remotes`(branch·exists, 선택 card_id·holder·generation). 원격 관찰이 없으면 그 예약은 \
         `held_for_review`. `worktrees` 는 이 컴퓨터에 없는 경로에만 쓰임",
    ),
    (
        "ledger-recover-apply",
        "observations",
        "미리 보기가 읽은 것과 같은 관측 파일",
    ),
    (
        "ledger-recover-plan",
        "now",
        "`YYYY-MM-DDTHH:MM:SSZ`. 없으면 시스템 시계(저널 lease 를 이때와 견줌)",
    ),
    (
        "ledger-recover-apply",
        "now",
        "미리 보기가 읽은 것과 같은 `--now`",
    ),
    (
        "",
        "home",
        "바꿀 HOME(기본 `$HOME`). 시험은 임시 폴더를 넘김",
    ),
    (
        "home-state-plan",
        "now",
        "`YYYY-MM-DDTHH:MM:SSZ`. 없으면 시스템 시계",
    ),
    (
        "home-state-apply",
        "now",
        "미리 보기가 읽은 것과 같은 `--now`",
    ),
    (
        "home-retire-plan",
        "now",
        "`YYYY-MM-DDTHH:MM:SSZ`. 없으면 시스템 시계(날짜 폴더 이름)",
    ),
    (
        "home-retire-apply",
        "now",
        "미리 보기가 읽은 것과 같은 `--now`",
    ),
    (
        "home-retire-revert",
        "stamp",
        "되돌릴 은퇴의 표지(`YYYYMMDDTHHMMSSZ`). 없으면 가장 새 은퇴",
    ),
    (
        "home-install-plan",
        "now",
        "`YYYY-MM-DDTHH:MM:SSZ`. 없으면 시스템 시계(저널 이름)",
    ),
    (
        "home-install-plan",
        "project",
        "Rust 세션을 셀 클론 안의 폴더(기본: 현재 폴더). 다른 클론은 HOME 에서 보이지 않음",
    ),
    (
        "home-install-apply",
        "home",
        "바꿀 HOME(필수). 실제 HOME 이 기본이 되는 일은 없음. 시험은 임시 폴더를 넘김",
    ),
    (
        "home-install-apply",
        "project",
        "Rust 세션을 셀 클론 안의 폴더(기본: 현재 폴더). 등록된 세션이 있거나 레지스트리를 못 \
         읽으면 거부",
    ),
    (
        "home-install-apply",
        "now",
        "미리 보기가 읽은 것과 같은 `--now`",
    ),
    (
        "home-install-revert",
        "home",
        "바꿀 HOME(필수). 실제 HOME 이 기본이 되는 일은 없음. 시험은 임시 폴더를 넘김",
    ),
    (
        "home-install-revert",
        "stamp",
        "되돌릴 설치의 표지(`YYYYMMDDTHHMMSSZ`). 없으면 가장 새 설치",
    ),
    (
        "home-state-revert",
        "now",
        "`YYYY-MM-DDTHH:MM:SSZ`. 없으면 시스템 시계(비켜 둘 폴더 이름)",
    ),
    (
        "home-state-plan",
        "selective",
        "폴더 전체를 옮기는 대신 맨 위 항목마다 나누기: 제자리·백업으로 옮김·비밀 \
         수동(절대 안 엶)·모름(아무것도 안 옮김)",
    ),
    (
        "home-state-apply",
        "permission_reference",
        "사용자 승인에 대한 호출자의 참조",
    ),
    (
        "home-state-revert",
        "permission_reference",
        "사용자 승인에 대한 호출자의 참조",
    ),
    (
        "worktree-move-apply",
        "permission_reference",
        "이 이동들에 대한 사용자 승인의 호출자 참조",
    ),
    (
        "home-state-apply",
        "copy",
        "Rust 폴더로 복사할 `~/.moai` 의 맨 위 항목. 여럿이면 되풀이",
    ),
    (
        "worktree-move-plan",
        "project",
        "저장소의 아무 체크아웃(primary 또는 worktree)",
    ),
    (
        "worktree-move-apply",
        "project",
        "저장소의 아무 체크아웃(primary 또는 worktree)",
    ),
    (
        "worktree-move-plan",
        "root",
        "worktree 를 옮겨 넣을 절대 폴더(기본 `<HOME>/.moai/worktrees`). 모든 worktree \
         밖이어야 함",
    ),
    (
        "worktree-move-apply",
        "root",
        "미리 보기가 읽은 것과 같은 `--root`(기본도 미리 보기와 같음)",
    ),
    (
        "worktree-move-plan",
        "observations",
        "JSON 파일: `schema_version` 1 과 선택 `holds`(source, branches: branch, reason). \
         `holds` 가 없으면 모든 worktree 가 `hold_check_not_run` 으로 막힘",
    ),
    (
        "worktree-move-apply",
        "observations",
        "미리 보기가 읽은 것과 같은 관측 파일",
    ),
    ("", "provider", "공급자 이름, 예: `glm`(소문자·숫자·`-`)"),
    (
        "",
        "shared",
        "모든 프로젝트가 쓰는 키(`~/.moai-cli/vault/_shared/`)",
    ),
    (
        "key-import-plan",
        "shared",
        "프로젝트 대신 `~/.moai-cli/vault/_shared/`(모든 프로젝트)로 가져오기",
    ),
    (
        "key-import-apply",
        "shared",
        "프로젝트 대신 `~/.moai-cli/vault/_shared/`(모든 프로젝트)로 가져오기",
    ),
    (
        "key-set",
        "prompt",
        "값을 숨긴 입력란에 직접 입력(화면에 보이지 않음). stdin 과 stderr 가 터미널일 때만. \
         답은 Enter 뒤 1초 지나서 나오고, 그 1초 안에 입력이 더 들어오면 값을 거절",
    ),
    (
        "key-remove",
        "yes",
        "묻지 않고 지우기. 없으면 미리 보기, 터미널에서는 질문 하나",
    ),
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
    // A command field may name several commands, `cc|codex|glm`.
    ARG_HELP_KO
        .iter()
        .find(|(c, i, _)| c.split('|').any(|w| w == command) && *i == id)
        .or_else(|| {
            ARG_HELP_KO
                .iter()
                .find(|(c, i, _)| c.is_empty() && *i == id)
        })
        .map(|(_, _, t)| *t)
}

/// The column clap starts an argument's help at in a long help (`          <help>`).
const LONG_ARG_INDENT: usize = 10;

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
    // The arguments and options sections are spelled out in the template, arguments first as in
    // English: a custom `help_heading` would be drawn in the order the arguments are declared,
    // which put `옵션:` before `인자:` for the launchers (their options come first).
    let positionals = if c.get_positionals().any(|a| !a.is_hide_set()) {
        format!("{ARGUMENTS_KO}\n{{positionals}}\n\n")
    } else {
        String::new()
    };
    let template = format!(
        "{{before-help}}{{about-with-newline}}\n{USAGE_KO} {{usage}}\n\n{positionals}{OPTIONS_KO}\n{{options}}{{after-help}}"
    );
    // clap's generated `-h, --help` cannot be edited before the command is built, so it is
    // replaced by the same flag with Korean lines (`ArgAction::Help`: `-h` short, `--help` long).
    // Every help starts on the line after its name, `LONG_ARG_INDENT` in, in `-h` as in
    // `--help`: clap counts a Hangul letter as one cell, so beside a wide name it would judge a
    // Korean help to fit when it does not (`codex-projection-plan -h` ran to 72 cells at 60).
    let room = width().saturating_sub(LONG_ARG_INDENT);
    c.disable_help_flag(true)
        .arg(
            clap::Arg::new("help")
                .short('h')
                .long("help")
                .action(clap::ArgAction::Help)
                .help(wrap(HELP_FLAG_SHORT_KO, 0, room))
                .long_help(wrap(HELP_FLAG_LONG_KO, 0, room)),
        )
        .next_line_help(true)
        .help_template(template)
}

/// A moved command's own texts in Korean (S6e-2): its about (the root list's), its long about
/// and its arguments' help, each wrapped at `width` display cells — clap would count a Hangul
/// letter as one cell and run the line past the terminal.
fn korean_texts(c: Command, name: &str, width: usize) -> Command {
    let about = about_ko(name).or_else(|| {
        LAUNCHER_ABOUT_KO
            .iter()
            .find(|(w, _)| *w == name)
            .map(|(_, t)| *t)
    });
    let c = match about {
        Some(about) => c.about(about),
        None => c,
    };
    let c = match LONG_ABOUT_KO.iter().find(|(w, _)| *w == name) {
        Some((_, long)) => c.long_about(wrap(long, 0, width)),
        None => c,
    };
    // The Korean frame puts every help on the line after its name (`korean_frame`), so one
    // wrap at the indent fits `-h` and `--help` alike.
    let room = width.saturating_sub(LONG_ARG_INDENT);
    let name = name.to_string();
    c.mut_args(move |a| match arg_help_ko(&name, a.get_id().as_str()) {
        Some(text) => {
            // A short help ends its last line with clap's ` [default: …]` and
            // ` [possible values: …]`; a long help puts them on lines of their own.
            let short = wrap(text, 0, room.saturating_sub(short_suffix_width(&a)));
            a.help(short).long_help(wrap(text, 0, room))
        }
        None => a,
    })
}

/// The width of what clap appends to an argument's short help line: ` [default: …]` and
/// ` [possible values: …]` (clap's words, left in English).
fn short_suffix_width(a: &clap::Arg) -> usize {
    // A flag (`--json`) takes no value: clap shows neither its `false` default nor `true|false`.
    if !a.get_action().takes_values() {
        return 0;
    }
    let join = |v: Vec<String>| v.join(", ");
    let defaults: Vec<String> = a
        .get_default_values()
        .iter()
        .map(|v| v.to_string_lossy().into_owned())
        .collect();
    let mut width = 0;
    if !defaults.is_empty() && !a.is_hide_default_value_set() {
        width += format!(" [default: {}]", join(defaults)).width();
    }
    let possible: Vec<String> = a
        .get_possible_values()
        .iter()
        .filter(|p| !p.is_hide_set())
        .map(|p| p.get_name().to_string())
        .collect();
    if !possible.is_empty() && !a.is_hide_possible_values_set() {
        width += format!(" [possible values: {}]", join(possible)).width();
    }
    width
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn bin_name_is_moai_whatever_name_was_typed() {
        for argv0 in [
            "/usr/local/bin/moai",
            "moai",
            "/x/moai-cli",
            "moai-cli",
            "C:/Users/u/.moai/path/moai-cli.exe",
            "/x/moai-copy",
            "",
        ] {
            assert_eq!(bin_name(argv0), "moai", "{argv0:?}");
        }
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

    /// `text` without its SGR style codes (`ESC [ … m`).
    fn without_styles(text: &str) -> String {
        let mut out = String::new();
        let mut chars = text.chars();
        while let Some(c) = chars.next() {
            if c == '\x1b' {
                for d in chars.by_ref() {
                    if d == 'm' {
                        break;
                    }
                }
            } else {
                out.push(c);
            }
        }
        out
    }

    /// #3004 P1: the English root help has its four blocks in order, a blank line between them,
    /// the groups one level under "Commands", one about column for every block, and no line
    /// past the width (60, 80 and 100), a trailing space or a style code.
    #[test]
    fn the_root_help_has_its_blocks_and_fits_its_width() {
        let cmd = root_cmd();
        for width in [60usize, 80, 100] {
            let text = render_root(&cmd, "moai", width, false);
            let heads: Vec<&str> = text
                .lines()
                .filter(|l| !l.is_empty() && !l.starts_with(' '))
                .collect();
            assert_eq!(
                heads,
                ["Getting started", "Commands", "Examples", "More"],
                "{text}"
            );
            assert!(
                text.contains("\n\nCommands\n  Start\n    cc, codex, glm, omp  "),
                "{text}"
            );
            assert!(text.contains("\n\nExamples\n  moai cc -l  "), "{text}");
            assert!(text.ends_with("command\n"), "{text}");
            assert!(!text.contains("\n\n\n"), "{text}");
            for line in text.lines() {
                assert!(line.width() <= width, "{width}: {line:?}");
                assert_eq!(line, line.trim_end(), "{width}: trailing space {line:?}");
                assert!(!line.contains('\x1b'), "{line:?}");
            }
            // One about column: the first about of each block starts at the same cell.
            let starts: std::collections::BTreeSet<usize> = text
                .lines()
                .filter_map(|l| {
                    [
                        "  1. moai install-inspect",
                        "  moai cc -l",
                        "  moai help --all",
                        "    cc, codex, glm, omp",
                    ]
                    .iter()
                    .find(|p| l.starts_with(**p))
                    .map(|p| p.len() + l[p.len()..].chars().take_while(|c| *c == ' ').count())
                })
                .collect();
            assert_eq!(starts.len(), 1, "{width}: {starts:?}\n{text}");
        }
    }

    /// #3004 P1 AC-5: style codes only when asked, and the same words and breaks either way.
    #[test]
    fn the_root_help_styles_only_when_asked_and_keeps_its_words() {
        let cmd = root_cmd();
        let plain = render_root(&cmd, "moai", 80, false);
        let styled = render_root(&cmd, "moai", 80, true);
        assert!(!plain.contains('\x1b'), "{plain}");
        assert!(
            styled.contains("\x1b[1mGetting started\x1b[0m"),
            "{styled:?}"
        );
        assert!(
            styled.contains("\x1b[1mmoai install-inspect\x1b[0m"),
            "{styled:?}"
        );
        assert_eq!(without_styles(&styled), plain);
    }

    /// #3004 P1: every name the root help types is a root command of the table (the real command
    /// tree is checked in `main.rs`), and the binary name is the one given.
    #[test]
    fn the_first_steps_name_root_commands_and_the_given_binary() {
        let root_words: std::collections::BTreeSet<&str> = GROUPS
            .iter()
            .flat_map(|g| g.commands.iter())
            .filter(|(_, root)| *root)
            .map(|(w, _)| *w)
            .collect();
        for (text, _) in GETTING_STARTED.iter().chain(ROOT_EXAMPLES.iter()) {
            if text.starts_with('/') {
                continue;
            }
            let word = text.split_whitespace().next().unwrap();
            assert!(root_words.contains(word), "{text}");
        }
        let text = render_root(&root_cmd(), "bin-x", 80, false);
        assert!(text.contains("  1. bin-x install-inspect  "), "{text}");
        assert!(!text.contains(" moai "), "{text}");
    }

    /// SPEC-20261117-installer-skill-agents REQ-11: `update` installs the skill, so it comes
    /// before the session and before `/moai project` in the first steps.
    #[test]
    fn the_first_steps_install_the_skill_before_the_session() {
        let at = |word: &str| {
            GETTING_STARTED
                .iter()
                .position(|(text, _)| *text == word)
                .unwrap_or_else(|| panic!("{word} is not a first step"))
        };
        assert!(at("update") < at("cc"));
        assert!(at("cc") < at("/moai project"));
        let text = render_root(&root_cmd(), "moai", 80, false);
        assert!(text.contains("  2. moai update  "), "{text}");
    }

    /// S6d AC-2: `help --all` shows the root commands in Korean and keeps the other abouts in
    /// English until S6e.
    #[test]
    fn a_korean_full_list_keeps_the_other_abouts_english() {
        let cmd = root_cmd();
        let ko = render_groups_in(&cmd, true, 100, Lang::Ko);
        assert!(ko.contains("About agents-plan"), "{ko}");
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

    /// The frame tests use `index-query`: a command outside the root list, whose own texts move
    /// in the last slice.
    fn sub_help(lang: Lang, long: bool) -> String {
        sub_help_of("index-query", lang, long)
    }

    /// S6e-2: a command of a moved group takes its Korean about, long about and argument help;
    /// an argument the table lacks keeps its own; English is untouched.
    #[test]
    fn a_moved_command_reads_its_korean_texts() {
        let ko = sub_help_of("project-detect", Lang::Ko, true);
        assert!(ko.contains("`/moai project` 경로가 쓰는"), "{ko}");
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
        assert!(ko.contains("moai index-query . < request.json"), "{ko}");
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
        // The English root help ends with its "More" block (#3004 P1), not the footer lines;
        // `help --all` keeps the footer.
        assert!(en.contains("\nMore\n  moai help --all "), "{en}");
        assert!(!en.contains("Run '"), "{en}");
        assert!(
            render(true, Lang::En)
                .ends_with("Run 'moai <command> --help' for details and examples.\n"),
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
