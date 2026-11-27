//! SPEC-20261117-sdd-guidance: the agent-facing SDD rules live in one reference
//! (`skills/moai/references/internal/sdd.md`), name only commands this build has, and the
//! workflows and the chain table follow one order, written once in the reference:
//! PLAN → RUN → VERIFY → SYNC → AUDIT → merge.
mod common;

use common::CleanEnv;
use std::{fs, path::PathBuf, process::Command};

const ORDER: &str = "PLAN → RUN → VERIFY → SYNC → AUDIT → merge";

fn root() -> PathBuf {
    PathBuf::from(env!("CARGO_MANIFEST_DIR"))
}

fn read(rel: &str) -> String {
    fs::read_to_string(root().join(rel)).unwrap_or_else(|e| panic!("{rel}: {e}"))
}

fn sdd() -> String {
    read("skills/moai/references/internal/sdd.md")
}

#[test]
fn the_reference_states_the_rules_the_user_decided() {
    let text = sdd();
    assert!(text.starts_with("# Internal: sdd\n\nWhen to read: "));
    for phrase in [
        // SPEC-20261125-plan-card-branch-without-spec REQ-7: the old "every change has a SPEC"
        // sentence is replaced; a SPEC is for tier 2 changes and the card is the design below it.
        "A SPEC is for tier 2 changes",
        // SPEC-20261122-next-flow-skills-agents (slice C audit B2): a card lands once through the
        // landing command; the pull request is a release's.
        "One SPEC = one branch `spec/<slug>-<yyyymmdd>` = one landing",
        "--minimal",
        "exactly one `shall`",
        "IF <condition>, THEN",
        "(REQ-<n>)",
        "· failure: none",
        "[NEEDS CLARIFICATION",
        "RUN does not start while one is open",
        "Only the user's in-session selection approves a SPEC",
        "digest",
        "`Moai-Spec: <SPEC-ID>`",
        "`@cc:later`",
        "preserve-then-change",
        ORDER,
    ] {
        assert!(text.contains(phrase), "sdd.md lacks {phrase:?}");
    }
    // The session trailer exists since SDD-6: the line names the commands that print and check
    // it and the reference that holds its rules, and no longer calls it planned.
    let session = text
        .lines()
        .find(|l| l.contains("Moai-Session"))
        .expect("sdd.md names the session trailer");
    for phrase in [
        "`moai trailers --spec <SPEC-ID>`",
        "`moai trailer-check <base>..HEAD`",
        "`internal/commit-trailers`",
        // SPEC-20261118-skill-audit-followups REQ-1 (SDD-5 audit N-a): the session line is
        // conditional on the host giving an id.
        "when the host gives a session id",
        "otherwise `Moai-Spec` only",
    ] {
        assert!(session.contains(phrase), "{session} lacks {phrase}");
    }
    assert!(!session.contains("planned"), "{session}");
}

/// SPEC-20261118-skill-audit-followups REQ-3 (SDD-6 audit N1): the trailer reference tells CI
/// and a detached checkout to name the branch, in the `trailer-check` item itself.
#[test]
fn the_trailer_reference_names_the_branch_flag_for_ci() {
    let text = read("skills/moai/references/internal/commit-trailers.md");
    let item = text
        .lines()
        .find(|l| l.contains("`moai trailer-check <base>..HEAD`"))
        .expect("commit-trailers.md names trailer-check");
    for phrase in ["`--branch <name>`", "detached", "CI"] {
        assert!(item.contains(phrase), "{item} lacks {phrase}");
    }
}

/// Every `moai spec <action>` the reference names is a subcommand of this build. The card
/// link (`moai card spec`, `--require-approved-spec`) comes with a separate change: while this
/// build lacks it, every line that names it says it is planned; once the build has it, no such
/// line may still say so (audit N1: the reference must then be updated).
#[test]
fn the_reference_names_only_commands_this_build_has() {
    let text = sdd();
    let mut actions: Vec<&str> = text
        .match_indices("`moai spec ")
        .map(|(i, m)| {
            let rest = &text[i + m.len()..];
            rest.split(|c: char| !c.is_ascii_lowercase())
                .next()
                .unwrap()
        })
        .collect();
    actions.sort_unstable();
    actions.dedup();
    assert_eq!(actions, ["check", "close", "list", "new", "show", "take"]);
    let home = tempfile::tempdir().unwrap();
    for action in &actions {
        let out = Command::new(common::MOAI_BIN)
            .clean_env()
            .env("HOME", home.path())
            .args(["spec", action, "--help"])
            .output()
            .unwrap();
        assert!(out.status.success(), "moai spec {action} --help");
    }
    // The trailer commands the session line names (SDD-6) are subcommands of this build.
    for command in ["trailers", "trailer-check"] {
        assert!(text.contains(&format!("`moai {command} ")), "{command}");
        let out = Command::new(common::MOAI_BIN)
            .clean_env()
            .env("HOME", home.path())
            .args([command, "--help"])
            .output()
            .unwrap();
        assert!(out.status.success(), "moai {command} --help");
    }
    let card = Command::new(common::MOAI_BIN)
        .clean_env()
        .env("HOME", home.path())
        .args(["card", "spec", "--help"])
        .output()
        .unwrap();
    let lines: Vec<&str> = text
        .lines()
        .filter(|l| l.contains("moai card spec") || l.contains("--require-approved-spec"))
        .collect();
    assert!(!lines.is_empty(), "sdd.md does not mention the card link");
    let available = card.status.success();
    for line in lines {
        let planned = line.contains("planned") || line.contains("once that change lands");
        if available {
            assert!(
                !planned,
                "`moai card spec` exists: say it is available: {line}"
            );
        } else {
            assert!(
                line.contains("planned") && line.contains("once that change lands"),
                "the card link is named as available: {line}"
            );
        }
    }
}

/// Audit B2: the two user rules no test held — existing code changes start from committed
/// characterization tests, and agents and tools add and link `@cc:` comments automatically
/// (the second rule replaced "agents never add them" on 2026-11-20).
#[test]
fn preserve_then_change_and_the_cc_rule_are_stated() {
    for (rel, phrase) in [
        (
            "skills/moai/workflows/run.md",
            "For a change to existing code (preserve-then-change), first commit characterization \
             tests that pass on the base and pin the behaviour the change must keep, then make the \
             change and show they still pass.",
        ),
        (
            "skills/moai/references/internal/sdd.md",
            "preserve-then-change for existing code",
        ),
        (
            "skills/moai/references/internal/sdd.md",
            "Agents and tools add them and link them automatically; SYNC keeps them true to the code.",
        ),
        (
            "skills/moai/workflows/sync.md",
            "keeps the `@cc:` comments true to the code and adds those the change calls for",
        ),
        (
            "agents/docs.md",
            "add or change only `@cc:` comment lines, where the change calls for them.",
        ),
        (
            "skills/moai/references/chains/index.md",
            "`@cc:` comments (agents and tools add and link them automatically)",
        ),
        (
            "skills/moai/references/internal/quality.md",
            "which agents and tools add and link automatically",
        ),
    ] {
        assert!(read(rel).contains(phrase), "{rel} lacks {phrase:?}");
    }
    // SPEC-20261120-acc-auto-link REQ-7: the name, the marker, the kinds and the read-only
    // rule for the old tags are unchanged.
    let text = sdd();
    for phrase in [
        "Agent Context Comments in code: `@cc:later`, `@cc:careful`, `@cc:proves`.",
        "Old `@MX:` and `@MOAI:` tags are read, never written.",
    ] {
        assert!(text.contains(phrase), "sdd.md lacks {phrase:?}");
    }
}

/// SPEC-20261120-acc-auto-link AC-3: the documentation agent may add `@cc:` comments and the
/// implementation agent is told to.
#[test]
fn the_agents_may_add_cc_comments() {
    let impl_agent = read("agents/impl.md");
    for kind in ["`@cc:later`", "`@cc:careful`", "`@cc:proves`"] {
        assert!(impl_agent.contains(kind), "agents/impl.md lacks {kind}");
    }
    assert!(
        impl_agent.contains("tools link it automatically"),
        "agents/impl.md does not say the link is automatic"
    );
}

/// SPEC-20261120-acc-auto-link AC-4: the three sentences that forbade adding `@cc:` comments
/// are in no rule text, agent definition or workflow.
#[test]
fn the_old_cc_sentences_are_gone() {
    let old = [
        "Agents never add them on their own",
        "`@cc:` comment lines, never add one",
        "it never adds new ones",
    ];
    let mut files: Vec<String> = ["sdd", "quality", "interview", "questions", "test-design"]
        .iter()
        .map(|n| format!("skills/moai/references/internal/{n}.md"))
        .collect();
    files.push("skills/moai/references/chains/index.md".into());
    files.push("skills/moai/SKILL.md".into());
    // RUN carries the retired VERIFY route (SPEC-20261122-next-flow-skills-agents).
    for wf in ["run", "sync", "audit", "plan"] {
        files.push(format!("skills/moai/workflows/{wf}.md"));
    }
    for agent in ["plan", "research", "impl", "audit", "docs"] {
        files.push(format!("agents/{agent}.md"));
    }
    for rel in files {
        let text = read(&rel);
        for phrase in old {
            assert!(!text.contains(phrase), "{rel} still says {phrase:?}");
        }
    }
}

/// SPEC-20261120-acc-auto-link AC-1: the decisions log has the 2026-11-20 row that names the
/// 2026-11-17 ACC row and supersedes its clause; that row is left as it was.
#[test]
fn the_decisions_log_supersedes_the_acc_clause() {
    let log = read("docs/decisions-2026-11-08.md");
    let row = log
        .lines()
        .find(|l| l.starts_with("| ") && l.contains("(11-20)") && l.contains("ACC @cc"))
        .expect("a 2026-11-20 ACC row");
    for phrase in [
        "2026-11-17",
        "ACC 코드 표식",
        "대체",
        "달고 자동으로 연결한다",
        "SPEC-20261120-acc-auto-link",
    ] {
        assert!(row.contains(phrase), "the new row lacks {phrase:?}: {row}");
    }
    assert!(
        log.lines()
            .any(|l| l.contains("(11-17)") && l.contains("에이전트는 자동 생성하지 않는다")),
        "the 2026-11-17 row was changed"
    );
}

#[test]
fn no_card_number_or_ddd_wording_in_the_package_and_agents() {
    let card = |s: &str| {
        s.match_indices("C-")
            .any(|(i, _)| s[i + 2..].starts_with(|c: char| c.is_ascii_digit()))
    };
    assert!(!card(&sdd()), "sdd.md names a card number");
    // Named files only (no folder walk): every workflow the route table names, the internal
    // references the method question and the order live in, the skill entry and the agents.
    let routes: serde_json::Value =
        serde_json::from_str(&read("skills/moai/references/routes.json")).unwrap();
    let mut files: Vec<String> = routes["routes"]
        .as_object()
        .unwrap()
        .values()
        .map(|r| {
            let rel = r["reference"].as_str().unwrap();
            format!(
                "skills/moai/{}",
                rel.strip_prefix("../").expect("route reference")
            )
        })
        .collect();
    for name in ["sdd", "interview", "questions", "test-design", "core-rules"] {
        files.push(format!("skills/moai/references/internal/{name}.md"));
    }
    files.push("skills/moai/references/chains/index.md".into());
    files.push("skills/moai/SKILL.md".into());
    for agent in ["plan", "research", "impl", "audit", "docs"] {
        files.push(format!("agents/{agent}.md"));
    }
    // Eight route texts (SPEC-20261122-next-flow-skills-agents REQ-1; twelve before), five
    // internal references, the chain table, SKILL.md and five agents.
    assert!(files.len() >= 20, "{files:?}");
    for rel in files {
        assert!(!read(&rel).contains("DDD"), "{rel} still says DDD");
    }
}

/// The four files that order work by the stages (the chain table and three workflows). Each
/// says that the order is the one `internal/sdd` gives and does not write it out
/// (SPEC-20261121-stage-order-once, which replaces REQ-3 of this file's first SPEC). RUN
/// carries the retired VERIFY route (SPEC-20261122-next-flow-skills-agents).
const POINTER_FILES: [&str; 4] = [
    "skills/moai/references/chains/index.md",
    "skills/moai/workflows/run.md",
    "skills/moai/workflows/sync.md",
    "skills/moai/workflows/audit.md",
];

/// The sentence every pointer file carries instead of the order (compared in lower case).
const POINTER: &str = "order of the stages is the one `internal/sdd` gives";

/// SPEC-20261121-stage-order-once AC-1: the full order is in `sdd.md` only; the four pointer
/// files name `internal/sdd` as its owner. Named files only (no folder walk): the entry, every
/// workflow the route table names, the chain table and the internal references.
#[test]
fn stage_order_is_written_once() {
    let routes: serde_json::Value =
        serde_json::from_str(&read("skills/moai/references/routes.json")).unwrap();
    let mut files: Vec<String> = routes["routes"]
        .as_object()
        .unwrap()
        .values()
        .map(|r| {
            let rel = r["reference"].as_str().unwrap();
            format!(
                "skills/moai/{}",
                rel.strip_prefix("../").expect("route reference")
            )
        })
        .collect();
    for name in ["sdd", "interview", "questions", "test-design", "core-rules"] {
        files.push(format!("skills/moai/references/internal/{name}.md"));
    }
    files.push("skills/moai/references/chains/index.md".into());
    files.push("skills/moai/SKILL.md".into());
    let with_order: Vec<&String> = files
        .iter()
        .filter(|rel| read(rel).contains(ORDER))
        .collect();
    assert_eq!(
        with_order,
        ["skills/moai/references/internal/sdd.md"],
        "the full order belongs to sdd.md only"
    );
    for rel in POINTER_FILES {
        assert!(
            read(rel).to_ascii_lowercase().contains(POINTER),
            "{rel} does not say that internal/sdd owns the order"
        );
    }
}

/// SPEC-20261121-stage-order-once AC-2: taking the order out leaves what each stage reads,
/// writes or refuses where it was.
#[test]
fn each_stage_workflow_keeps_its_own_duty() {
    for (rel, phrase) in [
        (
            "skills/moai/workflows/audit.md",
            "reads only, and any drift between the SPEC (its Scope, requirements and `## Result`), \
             the code, the tests, the documents and the `@cc:` comments is a non-blocking finding, \
             never a fix, unless it is one of the four FAIL kinds",
        ),
        (
            "skills/moai/workflows/sync.md",
            "SYNC writes, AUDIT checks. On the candidate branch, before the audit, SYNC makes the \
             SPEC, the documents and the `@cc:` comments match the code",
        ),
        (
            "skills/moai/workflows/run.md",
            "SYNC writes them into the SPEC's `## Result` before the audit",
        ),
        (
            "skills/moai/workflows/run.md",
            "RUN follows the SPEC as `internal/sdd` says: only its Scope, and never while a \
             `[NEEDS CLARIFICATION …]` question is open.",
        ),
        (
            "skills/moai/references/chains/index.md",
            "SYNC writes the SPEC Result, documents and `@cc:` comments (agents and tools add and \
             link them automatically) on the same branch, and AUDIT only checks them.",
        ),
    ] {
        assert!(read(rel).contains(phrase), "{rel} lacks {phrase:?}");
    }
}

/// The old build order (audit before docs) is gone.
#[test]
fn every_place_states_one_order() {
    for rel in POINTER_FILES {
        assert!(
            read(rel).to_ascii_lowercase().contains(POINTER),
            "{rel} does not point to the one order"
        );
    }
    for rel in [
        "skills/moai/references/chains/index.md",
        "skills/moai/SKILL.md",
        "skills/moai/workflows/run.md",
        "skills/moai/workflows/sync.md",
        "skills/moai/workflows/audit.md",
    ] {
        assert!(
            !read(rel).contains("moai-impl, moai-audit, moai-docs"),
            "{rel} keeps the old order"
        );
    }
    for rel in [
        "skills/moai/workflows/plan.md",
        "skills/moai/workflows/run.md",
        "skills/moai/workflows/sync.md",
        "skills/moai/workflows/audit.md",
    ] {
        assert!(
            read(rel).contains("`internal/sdd`"),
            "{rel} does not point to internal/sdd"
        );
    }
}
/// Verify wave 2: a new check that blocks starts in observe — the rule sentence in the SDD
/// guidance, the one-line reminder in the full SPEC template, and the manifest entry reason
/// naming this rule (the digest itself stays with the source-manifest check).
#[test]
fn new_blocking_gates_start_in_observe() {
    let sdd = sdd();
    for phrase in [
        "starts in observe",
        "would have blocked",
        "false positives",
        "the switch to enforce is a recorded decision",
    ] {
        assert!(sdd.contains(phrase), "sdd.md lacks {phrase:?}");
    }
    let template = read("templates/spec/full.md");
    assert!(
        template.contains("starts in observe"),
        "templates/spec/full.md lacks the observe reminder"
    );
    let manifest = read("skills/moai/references/source-manifest.json");
    let at = manifest
        .find("\"references/internal/sdd.md\"")
        .expect("the manifest lists the guidance file");
    // The entry runs to the next top-level key, not to a fixed window: an entry before this
    // one that grows long must not push the reason out of the searched span.
    let entry_end = manifest[at + 1..]
        .find("\n    \"")
        .map(|at| at + 1)
        .unwrap_or(manifest.len());
    let entry = &manifest[at..at + entry_end];
    assert!(
        entry.contains("gates-start-in-observe"),
        "the manifest entry reason does not name the rule change"
    );
}

/// SPEC-20261123-work-count-guidance AC-1: `test-design.md` carries one numbered rule that says how
/// to bound a time limit by a count of work, with the parts the method needs.
#[test]
fn test_design_has_one_work_count_rule() {
    let text = read("skills/moai/references/internal/test-design.md");
    let rules: Vec<&str> = text
        .lines()
        .filter(|line| {
            let digits = line.chars().take_while(|c| c.is_ascii_digit()).count();
            digits > 0 && line[digits..].starts_with(". ") && line.contains("count of work")
        })
        .collect();
    assert_eq!(
        rules.len(),
        1,
        "test-design.md needs exactly one numbered rule about a count of work"
    );
    for part in [
        "wall clock",
        "on the thread that runs the code",
        "input length",
        "stops the code at the limit",
        "counted less than the input length",
        "quadratic",
        "only the work that calls the counter",
    ] {
        assert!(
            rules[0].contains(part),
            "the work-count rule lacks {part:?}"
        );
    }
}

/// The text of `rel` from `heading` to the next `## ` heading or the end of the file.
fn card_branch_part(rel: &str, heading: &str) -> String {
    let text = read(rel);
    let start = text
        .find(heading)
        .unwrap_or_else(|| panic!("{rel} has no {heading:?}"));
    let rest = &text[start + heading.len()..];
    let end = rest.find("\n## ").unwrap_or(rest.len());
    rest[..end].to_string()
}

/// SPEC-20261125-plan-card-branch-without-spec AC-1 (REQ-1 to REQ-4): the PLAN workflow sorts a
/// card by its Write set and has one step for a card with no SPEC and one for tier 2.
#[test]
fn card_branch_ac1_plan_has_a_branch_for_each_tier() {
    let part = card_branch_part("skills/moai/workflows/plan.md", "## Card branches");
    for phrase in [
        "1. **Sort.**",
        "tier 2 when",
        "tier 0 or 1 otherwise",
        "treat the card as tier 2",
        "2. **Tier 0 or 1: no SPEC.**",
        "body and criteria are the design",
        "Write no SPEC and ask for no approval",
        "does not start with `spec/`",
        "`card/<slug>-<yyyymmdd>`",
        "never a card number",
        "`moai card land`",
        "does not take a card that names no SPEC",
        "3. **Tier 2: SPEC and approval.**",
        "`## Interview`",
        "hand only the approved revision to RUN",
        "never recorded by the lane",
    ] {
        assert!(
            part.contains(phrase),
            "plan.md Card branches lacks {phrase:?}"
        );
    }
    // The two branches come after the sort, in this order.
    let at = |p: &str| part.find(p).unwrap_or_else(|| panic!("lacks {p:?}"));
    assert!(at("1. **Sort.**") < at("2. **Tier 0 or 1: no SPEC.**"));
    assert!(at("2. **Tier 0 or 1: no SPEC.**") < at("3. **Tier 2: SPEC and approval.**"));
}

/// AC-2 (REQ-5, REQ-6): where the planning agent's Write set, Shared spots and Open questions
/// go, with only commands this build has.
#[test]
fn card_branch_ac2_plan_places_the_plan_agent_output() {
    let part = card_branch_part("skills/moai/workflows/plan.md", "## Card branches");
    for phrase in [
        "4. **Planning agent output.**",
        "`Write set`",
        "`Tests that read changed text`",
        "the card's scope",
        "`moai card scope`",
        "`Shared spots`",
        "`moai card line-lease`",
        "`Open questions`",
        "open-question marker (`internal/sdd` item 5)",
        "`LANE-BLOCKED kind=decision`",
    ] {
        assert!(
            part.contains(phrase),
            "plan.md Card branches lacks {phrase:?}"
        );
    }
    // Every `moai card <verb>` the section names is a command of this build.
    let home = tempfile::tempdir().unwrap();
    let mut seen = 0;
    for chunk in part.split("`moai card ").skip(1) {
        let verb: String = chunk
            .chars()
            .take_while(|c| c.is_ascii_lowercase() || *c == '-')
            .collect();
        let out = Command::new(common::MOAI_BIN)
            .clean_env()
            .env("HOME", home.path())
            .args(["card", &verb, "--help"])
            .output()
            .unwrap();
        assert!(out.status.success(), "`moai card {verb}` is not a command");
        seen += 1;
    }
    assert!(seen >= 4, "the section names too few card commands: {seen}");
}

/// AC-3 (REQ-7, REQ-8, REQ-10): the SDD reference says a SPEC is for tier 2 and a FAIL is for
/// the four kinds, and the PLAN workflow names no Issue.
#[test]
fn card_branch_ac3_sdd_and_plan_follow_the_recorded_decisions() {
    let sdd = sdd();
    for phrase in [
        "A SPEC is for tier 2 changes",
        "the card's body and criteria are the design",
        "does not start with `spec/`",
        "it FAILs for the four kinds that `references/audit.md` names",
        "is a non-blocking finding",
    ] {
        assert!(sdd.contains(phrase), "sdd.md lacks {phrase:?}");
    }
    for gone in [
        "Every change that modifies the repository has a SPEC",
        "FAILs on any drift",
        "cards stay personal to-do items",
    ] {
        assert!(!sdd.contains(gone), "sdd.md still says {gone:?}");
    }
    let plan = read("skills/moai/workflows/plan.md");
    assert!(!plan.contains("Issue"), "plan.md still names an Issue");
}

/// AC-4 (REQ-9): READ answers a one-fact question directly and keeps the fan-out for the rest.
#[test]
fn card_branch_ac4_read_does_not_fan_out_one_fact() {
    let read_md = read("skills/moai/workflows/read.md");
    for phrase in [
        "A question that is one fact one source answers",
        "answered directly",
        "start no `moai-research` for it",
        "Otherwise READ fans out: split the question into independent parts",
    ] {
        assert!(read_md.contains(phrase), "read.md lacks {phrase:?}");
    }
}

/// The lines of `text` that hold every one of `needles`.
fn lines_with<'a>(text: &'a str, needles: &[&str]) -> Vec<&'a str> {
    text.lines()
        .filter(|l| needles.iter().all(|n| l.contains(n)))
        .collect()
}

/// C-607 criterion 1: the SPEC that `moai spec new` writes names, in its Requirements section,
/// the one-`shall` rule and the failure-path rule on one line each, so an author learns them
/// before `moai spec check` refuses the file. The file is written by the real command, minimal and
/// full.
#[test]
fn a_new_spec_names_the_one_shall_and_failure_path_rules_in_its_requirements() {
    let home = tempfile::tempdir().unwrap();
    let project = home.path().join("proj");
    // The state folder of the fresh project comes from the resolver, as the command finds it.
    let state = moai_cli::state_root::path(&project, "");
    fs::create_dir_all(&state).unwrap();
    fs::write(
        state.join("moai.json"),
        moai_cli::state_root::marker_bytes("test", &[]),
    )
    .unwrap();
    for (slug, minimal) in [("rules-full", false), ("rules-minimal", true)] {
        let mut args = vec![
            "spec",
            "new",
            slug,
            "--title",
            "Rules",
            "--date",
            "2026-11-27",
        ];
        if minimal {
            args.push("--minimal");
        }
        let out = Command::new(common::MOAI_BIN)
            .clean_env()
            .env("HOME", home.path())
            .args(&args)
            .arg("--project")
            .arg(&project)
            .output()
            .unwrap();
        assert!(
            out.status.success(),
            "{slug}: {}",
            String::from_utf8_lossy(&out.stdout)
        );
        let written = fs::read_to_string(moai_cli::state_root::path(
            &project,
            &format!("specs/SPEC-20261127-{slug}/spec.md"),
        ))
        .unwrap();
        let from = written.find("## Requirements").expect("Requirements");
        let section = &written[from..];
        let section = &section[..section.find("## Acceptance").expect("Acceptance")];
        let one_shall = lines_with(section, &["exactly one `shall`", "req_gears"]);
        assert_eq!(
            one_shall.len(),
            1,
            "{slug}: the one-shall rule is not on one line of Requirements:\n{section}"
        );
        let failure = lines_with(
            section,
            &["IF", "THEN", "failure: none", "req_failure_path"],
        );
        assert_eq!(
            failure.len(),
            1,
            "{slug}: the failure-path rule is not on one line of Requirements:\n{section}"
        );
        // Both sit inside one multi-line HTML comment, so the author's requirement lines stay
        // the only `REQ-` lines. The comment is not one line `<!-- … -->`: `moai spec check`
        // reads such a line as template text still to fill and refuses it once the SPEC is
        // approved, and the opening line `<!--` ends without `>`, so it is no such text.
        let open = section.find("<!--\n").expect("the rules open a comment");
        let close = open + section[open..].find("\n-->").expect("the comment closes");
        for line in one_shall.iter().chain(failure.iter()) {
            let at = section.find(line).unwrap();
            assert!(
                open < at && at < close,
                "{slug}: not inside the comment: {line}"
            );
            assert!(
                !line.contains(['<', '>']),
                "{slug}: reads as template text: {line}"
            );
        }
    }
}

/// C-607 criterion 2: `sdd.md` states each rule on a line of its own, with the `moai spec check`
/// code that refuses it.
#[test]
fn the_sdd_reference_states_each_check_rule_on_one_line() {
    let text = sdd();
    let one_shall = lines_with(&text, &["exactly one `shall`", "req_gears"]);
    assert_eq!(one_shall.len(), 1, "one-shall rule lines: {one_shall:?}");
    let failure = lines_with(&text, &["· failure: none", "req_failure_path"]);
    assert_eq!(failure.len(), 1, "failure-path rule lines: {failure:?}");
    assert_ne!(one_shall[0], failure[0], "both rules share a line");
}
