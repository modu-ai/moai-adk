//! #2592 S1, #2721: the base agents ship as tool-neutral sources in `agents/`, outside the
//! `moai` skill tree, and render into each host's agent file. Reads the sources and the
//! package; writes nothing.

use moai_cli::agent_defs::{package_agents, render, HOSTS, SOURCE_DIR};
use std::path::Path;

/// The renderings that still exist: the home installer's host and the OMP file the project layer
/// writes until it moves out.
const RENDERED_HOSTS: [&str; 2] = ["claude", "omp"];

fn package() -> std::path::PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai")
}

fn agents() -> std::path::PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR")).join(SOURCE_DIR)
}

/// #2721 AC-1: the skill tree carries no agent definition — no `agents/` folder at any depth
/// and no file with agent frontmatter — and every source sits in the folder beside `skills/`.
#[test]
fn the_skill_tree_carries_no_agent_definition() {
    fn walk(dir: &Path, out: &mut Vec<std::path::PathBuf>) {
        for e in std::fs::read_dir(dir).unwrap() {
            let p = e.unwrap().path();
            if p.is_dir() {
                assert_ne!(
                    p.file_name().unwrap(),
                    "agents",
                    "agent folder inside the skill: {}",
                    p.display()
                );
                walk(&p, out);
            } else {
                out.push(p);
            }
        }
    }
    let mut files = Vec::new();
    walk(&package(), &mut files);
    assert!(files.iter().any(|p| p.ends_with("SKILL.md")));
    for p in &files {
        let text = std::fs::read_to_string(p).unwrap_or_default();
        assert!(
            !(text.starts_with("---\nname: moai-") && text.contains("\nshape: ")),
            "agent source inside the skill: {}",
            p.display()
        );
    }
    let manifest =
        std::fs::read_to_string(package().join("references/source-manifest.json")).unwrap();
    assert!(!manifest.contains("references/agents/"));
    assert_eq!(package_agents(&agents(), &package()).unwrap().len(), 5);
}

/// #2721 (scope comment 6854556046): every base agent defines its purpose, when to use it and
/// when not, its input, its boundaries (write scope and tools), its steps, its verification,
/// its output and its handoff, in English, and the description says both when to use it and
/// when not to.
#[test]
fn every_agent_defines_purpose_boundaries_verification_and_handoff() {
    const SECTIONS: [&str; 8] = [
        "## Purpose\n",
        "## When to use\n",
        "## Input\n",
        "## Boundaries\n",
        "## Steps\n",
        "## Verification\n",
        "## Output\n",
        "## Handoff\n",
    ];
    for d in package_agents(&agents(), &package()).unwrap() {
        let mut at = 0;
        for s in SECTIONS {
            let found = d.body[at..]
                .find(s)
                .unwrap_or_else(|| panic!("{}: {s:?} missing or out of order", d.name));
            at += found + s.len();
        }
        assert!(d.body.contains("- Write scope: "), "{}", d.name);
        assert!(d.body.contains("- Tools: "), "{}", d.name);
        assert!(d.description.contains(" Use "), "{}", d.name);
        assert!(d.description.contains("Do not use "), "{}", d.name);
        assert!(
            d.body.is_ascii() && d.description.is_ascii(),
            "{}: agent text is English ASCII",
            d.name
        );
    }
}

/// #3054 AC1: the agents that build (`moai-impl`, `moai-audit`) and the AUDIT and RUN workflows
/// say how to clean up: inside the agent's own worktree, by relative path, only its own build
/// output, never by an absolute path, and logs stay until the verdict. No instruction text
/// shows a recursive delete of an absolute path, which the project's deny rules refuse.
#[test]
fn builders_clean_up_by_relative_path_inside_their_own_worktree() {
    for d in package_agents(&agents(), &package()).unwrap() {
        if d.name != "moai-impl" && d.name != "moai-audit" {
            continue;
        }
        let rule = d
            .body
            .lines()
            .find(|l| l.starts_with("- Cleanup: "))
            .unwrap_or_else(|| panic!("{}: no `- Cleanup: ` rule in Boundaries", d.name));
        // The implementation agent has a worktree of its own; the audit agent has none and
        // cleans the audit worktree named in `scope` (SPEC-20261123-audit-scope-worktree-text).
        let (made_in, inside) = if d.name == "moai-audit" {
            (
                "build output you made in the audit worktree named in `scope`",
                "from inside the audit worktree with a relative path",
            )
        } else {
            (
                "build output you made in your own worktree",
                "from inside your own worktree with a relative path",
            )
        };
        for needle in [
            made_in,
            inside,
            "rm -rf target/debug",
            "Never pass an absolute path",
            "Keep logs and evidence folders until",
            "report the path and its size under `Cleanup`",
        ] {
            assert!(
                rule.contains(needle),
                "{}: {needle:?} missing: {rule}",
                d.name
            );
        }
        assert!(
            d.body.contains("- `Cleanup`: "),
            "{}: Output lacks `Cleanup`",
            d.name
        );
        for l in d.body.lines() {
            assert!(
                !l.contains("rm -rf /") && !l.contains("rm -rf ~") && !l.contains("rm -rf \"/"),
                "{}: absolute recursive delete shown: {l}",
                d.name
            );
        }
    }
    for flow in ["workflows/audit.md", "workflows/run.md"] {
        let text = std::fs::read_to_string(package().join(flow)).unwrap();
        let rule = text
            .lines()
            .find(|l| l.starts_with("- Cleanup: ") || l.starts_with("- Cleanup ("))
            .unwrap_or_else(|| panic!("{flow}: no cleanup rule"));
        assert!(
            rule.contains("relative path") && rule.contains("`rm -rf target/debug`"),
            "{flow}: {rule}"
        );
        assert!(
            rule.contains("never") && rule.contains("absolute path"),
            "{flow}: {rule}"
        );
        assert!(
            !text.contains("rm -rf /"),
            "{flow}: absolute recursive delete shown"
        );
    }
    // The cleanup rule of the cargo output moved into `internal/build-slots`, which RUN step 1
    // names (SPEC-20261125-run-sync-procedure-receipt): it deletes by relative path too.
    let slots =
        std::fs::read_to_string(package().join("references/internal/build-slots.md")).unwrap();
    assert!(
        slots.contains(
            "by relative path (`rm -rf target/debug`), never through an absolute or home path"
        ),
        "build-slots.md: the cleanup rule does not say relative path and never absolute"
    );
    assert!(
        !slots.contains("rm -rf /") && !slots.contains("rm -rf ~"),
        "build-slots.md: absolute recursive delete shown"
    );
}

/// #2721: a symlinked or non-folder source path is refused, never followed.
#[cfg(unix)]
#[test]
fn a_symlinked_or_file_source_folder_is_refused() {
    let tmp = tempfile::tempdir().unwrap();
    let link = tmp.path().join("agents");
    std::os::unix::fs::symlink(agents(), &link).unwrap();
    assert!(package_agents(&link, &package())
        .unwrap_err()
        .contains("symlink"));
    let file = tmp.path().join("file");
    std::fs::write(&file, "x").unwrap();
    assert!(package_agents(&file, &package())
        .unwrap_err()
        .contains("not a folder"));
}

#[test]
fn the_package_holds_the_five_roles_of_the_contract() {
    let defs = package_agents(&agents(), &package()).unwrap();
    let names: Vec<&str> = defs.iter().map(|d| d.name.as_str()).collect();
    assert_eq!(
        names,
        [
            "moai-audit",
            "moai-docs",
            "moai-impl",
            "moai-plan",
            "moai-research"
        ]
    );
    let get = |n: &str| defs.iter().find(|d| d.name == n).unwrap();
    // Contract 6840103566: audit, research and plan write nothing; research fans out;
    // implementation and audit set no isolation of their own: audit reads the worktree its brief
    // gives (SPEC-20261123-audit-rule-text D-4) and implementation works in the worktree its
    // parent is already in (C-493, user decision 2026-11-25), because a second worktree rebuilds
    // from nothing.
    for n in ["moai-audit", "moai-research", "moai-plan"] {
        assert!(get(n).readonly, "{n}");
    }
    for n in ["moai-impl", "moai-docs"] {
        assert!(!get(n).readonly, "{n}");
    }
    assert_eq!(get("moai-research").shape, "fanout");
    assert_eq!(get("moai-impl").isolation, "none", "moai-impl");
    assert_eq!(get("moai-audit").isolation, "none", "moai-audit");
    for d in &defs {
        let role = d.name.trim_start_matches("moai-");
        // The fixed first line any parent reads, and the no-nesting rule.
        assert!(
            d.body
                .contains(&format!("`RESULT {role} <ok|blocked|not_run> ")),
            "{}",
            d.name
        );
        assert!(d.body.contains("Do not start other agents"), "{}", d.name);
    }
}

#[test]
fn every_role_renders_for_every_host_deterministically() {
    // The home installer writes Claude Code only; the OMP rendering stays for the project layer
    // (`omp_agents`) and the Codex rendering is gone.
    assert_eq!(HOSTS, ["claude"]);
    let defs = package_agents(&agents(), &package()).unwrap();
    for d in &defs {
        for host in RENDERED_HOSTS {
            let a = render(d, host).unwrap();
            assert_eq!(a, render(d, host).unwrap());
            assert!(a.path.starts_with("agents/moai-"), "{}", a.path);
            assert!(a.text.contains(&d.name));
        }
        assert!(render(d, "codex").is_err(), "{}", d.name);
        // #2633: OMP enforces read-only by its tool list and names per-call isolation.
        let omp = render(d, "omp").unwrap();
        assert_eq!(
            omp.text.contains(", edit, write\n"),
            !d.readonly,
            "{}",
            d.name
        );
        assert_eq!(
            omp.not_supported.iter().any(|n| n.starts_with("isolation")),
            d.isolation == "worktree",
            "{}",
            d.name
        );
    }
}

/// #3096 S7: a rendered agent file names no model and no effort. In the one run per product
/// measured on 2026-11-16 (`docs/agent-model-effort-inheritance.md`) a child of such a file
/// showed its parent's model and effort. This test pins the files; it does not prove the host.
/// The Claude Code and OMP frontmatter is read line by line (#3110 audit 6905677737 B2).
#[test]
fn rendered_agent_files_name_no_model_and_no_effort() {
    const KEYS_CLAUDE_OMP: [&str; 5] = [
        "name",
        "description",
        "disallowedTools",
        "isolation",
        "tools",
    ];
    let defs = package_agents(&agents(), &package()).unwrap();
    assert!(!defs.is_empty());
    for d in &defs {
        for host in RENDERED_HOSTS {
            let text = render(d, host).unwrap().text;
            let keys: Vec<String> = text
                .lines()
                .skip(1)
                .take_while(|l| *l != "---")
                .map(|l| {
                    l.split_once(':')
                        .unwrap_or_else(|| panic!("{} {host}: frontmatter line {l:?}", d.name))
                        .0
                        .to_string()
                })
                .collect();
            assert!(!keys.is_empty(), "{} {host}", d.name);
            for k in &keys {
                assert!(
                    KEYS_CLAUDE_OMP.contains(&k.as_str()),
                    "{} {host}: key {k:?}",
                    d.name
                );
            }
        }
    }
}

#[test]
fn a_package_without_sources_has_no_agents_and_a_stray_file_is_refused() {
    let tmp = tempfile::tempdir().unwrap();
    let dir = tmp.path().join("agents");
    assert!(package_agents(&dir, tmp.path()).unwrap().is_empty());
    std::fs::create_dir_all(&dir).unwrap();
    std::fs::write(dir.join("notes.txt"), "x").unwrap();
    assert!(package_agents(&dir, tmp.path())
        .unwrap_err()
        .contains("only regular .md files"));
    std::fs::remove_file(dir.join("notes.txt")).unwrap();
    std::fs::write(
        dir.join("impl.md"),
        "---\nname: moai-other\ndescription: d\nreadonly: false\nisolation: none\nshape: sequential\n---\n\nx\n",
    )
    .unwrap();
    assert!(package_agents(&dir, tmp.path())
        .unwrap_err()
        .contains("name must be moai-impl"));
}

/// #2637: each base agent's `reads:` list is the final design's §3 table (#2634) —
/// `moai-research` reads the two adapters #2636 added, and the five Go-generation references
/// #2582 adopted (leader confirmation 6856677422) go to `moai-impl`, `moai-audit` and
/// `moai-docs`; every entry is a file of the package and every rendered agent opens with the
/// generated line that names exactly those files.
#[test]
fn reads_lists_match_the_design_and_render_on_every_host() {
    let defs = package_agents(&agents(), &package()).unwrap();
    let reads = |n: &str| {
        defs.iter()
            .find(|d| d.name == n)
            .unwrap()
            .reads
            .iter()
            .map(String::as_str)
            .collect::<Vec<_>>()
    };
    // SPEC-20261117-atlas-in-read-plan: the two read-only agents of READ and PLAN check the
    // code atlas before they search.
    assert_eq!(
        reads("moai-plan"),
        [
            "internal/core-rules",
            "internal/decide",
            "internal/interview",
            "internal/code-atlas"
        ]
    );
    assert_eq!(
        reads("moai-research"),
        [
            "internal/core-rules",
            "adapters/aside",
            "adapters/orca",
            "internal/code-atlas"
        ]
    );
    assert_eq!(
        reads("moai-impl"),
        [
            "internal/core-rules",
            "internal/test-design",
            "internal/git-safety",
            "internal/tdd-results",
            // C-77: the pre-PR self-check list, run before the change goes to the audit.
            // SPEC-20261122-next-flow-skills-agents slice C: `internal/build-folder` left the
            // package; where builds go is RUN step 1 (`internal/build-slots`), which the body
            // points to.
            "internal/self-check"
        ]
    );
    assert_eq!(
        reads("moai-audit"),
        [
            // SPEC-20261122-next-flow-skills-agents slice B: the review note became the audit
            // contract `references/audit.md`, which no `<folder>/<name>` entry can name; the
            // source body points to it instead. Slice C: `internal/build-folder` left the
            // package as well; the body points to RUN step 1.
            "internal/core-rules",
            "internal/test-design",
            "internal/check-completeness",
            "internal/boundary-checks",
            "internal/security-checklist"
        ]
    );
    // SPEC-20261117-sdd-guidance: moai-docs is SYNC's writer and keeps the SPEC Result.
    assert_eq!(
        reads("moai-docs"),
        [
            "internal/core-rules",
            "internal/decide",
            "internal/doc-claims",
            "internal/sdd"
        ]
    );
    for d in &defs {
        for entry in &d.reads {
            assert!(
                package().join(format!("references/{entry}.md")).is_file(),
                "{entry}"
            );
        }
        for host in RENDERED_HOSTS {
            let text = render(d, host).unwrap().text;
            assert_eq!(
                text.contains("Read first,"),
                !d.reads.is_empty(),
                "{}",
                d.name
            );
            for entry in &d.reads {
                assert!(
                    text.contains(&format!("`references/{entry}.md`")),
                    "{entry}"
                );
            }
        }
        // The source body never repeats the list by hand.
        assert!(!d.body.contains("Read first,"), "{}", d.name);
    }
}

#[test]
fn a_reads_entry_without_its_file_is_refused() {
    let tmp = tempfile::tempdir().unwrap();
    let dir = tmp.path().join("agents");
    std::fs::create_dir_all(&dir).unwrap();
    std::fs::write(
        dir.join("impl.md"),
        "---\nname: moai-impl\ndescription: d\nreadonly: false\nisolation: none\nshape: sequential\nreads: internal/test-design\n---\n\nx\n",
    )
    .unwrap();
    let pkg = tmp.path().join("skills/moai");
    assert!(package_agents(&dir, &pkg)
        .unwrap_err()
        .contains("reads entry internal/test-design has no file"));
    // A file beside the source folder does not count: the entry resolves in the skill package.
    std::fs::create_dir_all(tmp.path().join("references/internal")).unwrap();
    std::fs::write(tmp.path().join("references/internal/test-design.md"), "t").unwrap();
    assert!(package_agents(&dir, &pkg)
        .unwrap_err()
        .contains("reads entry internal/test-design has no file"));
    std::fs::create_dir_all(pkg.join("references/internal")).unwrap();
    std::fs::write(pkg.join("references/internal/test-design.md"), "t").unwrap();
    assert_eq!(package_agents(&dir, &pkg).unwrap().len(), 1);
}

/// The internal references (#2637): one short note each, opening with the line that says
/// when to read it, and every one of them read by some agent or by a chain.
#[test]
fn internal_references_say_when_to_read_them() {
    let names = [
        "decide",
        "interview",
        // C-51/C-52 (user decision 2026-11-16): how a question reaches the user.
        "questions",
        "test-design",
        "git-safety",
        // #2582: adopted from the Go generation (leader confirmation 6856677422).
        "tdd-results",
        "check-completeness",
        "boundary-checks",
        "doc-claims",
        "security-checklist",
        // #2876: the rules every route, host and base agent follows.
        "core-rules",
        // #3096 S1: four ways of looking, read only when a question or plan needs one.
        "lenses",
        // C-69: the three versioning schemes the PROJECT interview offers.
        "versioning",
        // C-77: the checks an implementer runs before asking for an audit.
        "self-check",
        // SPEC-20261117-sdd-guidance: the agent-facing SDD rules.
        "sdd",
        // SPEC-20261117-atlas-in-read-plan: READ and PLAN read the code atlas when it is fresh.
        "code-atlas",
        // `build-folder` left the package in SPEC-20261122-next-flow-skills-agents slice C:
        // its rule, rewritten (D-14), became part of RUN's "Check-run rules" and is now
        // `internal/build-slots` (SPEC-20261125-run-sync-procedure-receipt).
        // SDD-6 (user decision U-B): the identity trailers every agent commit carries.
        "commit-trailers",
        // SPEC-20261117-quality-guides: the project's own check commands, AC-to-test tracing
        // and the dependency existence check, read by RUN through the chain rules.
        "quality",
        // SPEC-20261123-skill-structure: when a timer may repeat the leader round.
        "loop",
        // SPEC-20261125-run-sync-procedure-receipt: the long rules that left RUN's text, one
        // reference each, read from the step that names it (RUN steps 1, 2 and 5, and the
        // failure table's refused-command row).
        "build-slots",
        "check-rules",
        "shell-guards",
        "web-pages",
    ];
    let mut found: Vec<String> = std::fs::read_dir(package().join("references/internal"))
        .unwrap()
        .map(|e| e.unwrap().file_name().to_string_lossy().into_owned())
        .collect();
    found.sort();
    let mut want: Vec<String> = names.iter().map(|n| format!("{n}.md")).collect();
    want.sort();
    assert_eq!(found, want);
    let defs = package_agents(&agents(), &package()).unwrap();
    let chains = std::fs::read_to_string(package().join("references/chains/index.md")).unwrap();
    for n in names {
        let text =
            std::fs::read_to_string(package().join(format!("references/internal/{n}.md"))).unwrap();
        assert!(
            text.starts_with(&format!("# Internal: {n}\n\nWhen to read: ")),
            "{n}"
        );
        let entry = format!("internal/{n}");
        assert!(
            defs.iter().any(|d| d.reads.contains(&entry)) || chains.contains(&entry),
            "{entry} is read by no agent and no chain"
        );
    }
}

/// The chain table (#2637): every route is a route of `routes.json`, every agent a base agent,
/// every entry the parent reads a package file; `routes.json` names the table; the rules that
/// keep implementation and audit apart are stated.
#[test]
fn chains_name_real_routes_agents_and_references() {
    use serde_json::Value;
    let routes: Value =
        serde_json::from_slice(&std::fs::read(package().join("references/routes.json")).unwrap())
            .unwrap();
    assert_eq!(routes["chains"], "chains/index.md");
    let text = std::fs::read_to_string(package().join("references/chains/index.md")).unwrap();
    let agents: Vec<String> = package_agents(&agents(), &package())
        .unwrap()
        .into_iter()
        .map(|d| d.name)
        .collect();
    let rows: Vec<Vec<&str>> = text
        .lines()
        .filter(|l| l.starts_with("| ") && !l.starts_with("| Request") && !l.starts_with("| ---"))
        .map(|l| l.trim_matches('|').split('|').map(str::trim).collect())
        .collect();
    // Nine rows since SPEC-20261124-moai-leader-route REQ-11 added the LEADER row; eleven since
    // SPEC-20261120-plan-output-write-set REQ-7 added the flaky-test and the repair rows.
    assert_eq!(rows.len(), 11);
    let mut covered: Vec<&str> = Vec::new();
    for row in &rows {
        assert_eq!(row.len(), 4, "{row:?}");
        for route in row[1].split(", ") {
            assert!(routes["routes"][route].is_object(), "route {route}");
            covered.push(route);
        }
        if row[2] != "none" {
            for agent in row[2].split(", ") {
                assert!(agents.iter().any(|a| a == agent), "agent {agent}");
            }
        }
        if row[3] != "none" {
            for entry in row[3].split(", ") {
                assert!(
                    package().join(format!("references/{entry}.md")).is_file(),
                    "{entry}"
                );
            }
        }
    }
    for route in [
        "PLAN", "RUN", "AUDIT", "READ", "PROJECT", "SYNC", "WORK", "LEADER",
    ] {
        assert!(covered.contains(&route), "{route} has no chain");
    }
    for rule in [
        "The parent session starts each agent in turn",
        "An agent never starts another agent.",
        "The agent that implemented a change never audits it.",
        // Slice B audit N2: who records the verdict is the audit contract's rule alone.
        "the agent never writes the verdict",
        "`references/audit.md`, \"Who records the verdict\"",
    ] {
        assert!(text.contains(rule), "{rule}");
    }
    // The build chain runs implementation before the audit, never in the same agent; SYNC
    // (moai-docs) writes before AUDIT checks (SPEC-20261117-sdd-guidance).
    let build = rows.iter().find(|r| r[0] == "Build something new").unwrap();
    assert_eq!(build[1], "PLAN, RUN, SYNC, AUDIT");
    assert_eq!(build[2], "moai-plan, moai-impl, moai-docs, moai-audit");
    let fix = rows.iter().find(|r| r[0] == "Fix a bug").unwrap();
    assert_eq!(fix[1], "PLAN, RUN, SYNC, AUDIT");
    assert_eq!(
        fix[2],
        "moai-research, moai-plan, moai-impl, moai-docs, moai-audit"
    );
}

/// The question bank (#2639, SPEC-20261124-moai-project-interview): three tables, what is asked,
/// what is confirmed and the operating rules that are never asked; every row has five cells and
/// every answer lands in `decisions` or a real profile field.
#[test]
fn interview_question_bank_covers_the_project_axes() {
    let text = std::fs::read_to_string(package().join("references/internal/interview.md")).unwrap();
    let rows: Vec<Vec<String>> = text
        .lines()
        .filter(|l| l.starts_with("| `"))
        .map(|l| {
            l.trim_matches('|')
                .split(" | ")
                .map(|c| c.trim().to_string())
                .collect()
        })
        .collect();
    let ids: Vec<&str> = rows.iter().map(|r| r[0].trim_matches('`')).collect();
    assert_eq!(
        ids,
        [
            // Asked.
            "purpose",
            // Confirmed.
            "stack",
            "test_command",
            "build_command",
            "method",
            "hosts",
            // C-69: the versioning scheme, then the one detail the pick may need.
            "versioning",
            "versioning_format",
            "versioning_head",
            // C-70: the git work flow, GitHub flow recommended.
            "work_flow",
            // Defaults, never asked.
            "protected",
            "review_blocks",
            "autonomy",
            "done"
        ]
    );
    for r in &rows {
        assert_eq!(r.len(), 5, "{r:?}");
        let id = r[0].trim_matches('`');
        let never = ["protected", "review_blocks", "autonomy", "done"].contains(&id);
        assert_eq!(r[1] == "never", never, "{id}");
        if never {
            assert_eq!(r[2], "not asked", "{id}");
            assert_eq!(r[4], "decisions", "{id}");
            continue;
        }
        assert!(r[2].contains('?'), "{id}: the question cell asks");
        if id != "purpose" {
            assert!(
                r[3].contains("(recommended") && r[3].contains(';'),
                "{id}: options with a recommended one: {}",
                r[3]
            );
        }
        let to = r[4].as_str();
        assert!(
            to == "decisions"
                || [
                    "profile `purpose`",
                    "profile `tools`",
                    "profile `development_methods`",
                    "profile `hosts`",
                    "profile `versioning`",
                    "profile `work_flow`"
                ]
                .contains(&to),
            "{id}: {to}"
        );
    }
    assert!(text.contains("skip every row that reading already answered"));
    assert!(text.contains("\"source\": \"repo\""));
    assert!(text.contains("\"source\": \"user\""));
    assert!(text.contains("\"source\": \"default\""));
}

/// Card C-70 AC-1/AC-4: the `work_flow` row comes after the versioning rows, offers GitHub flow
/// first and recommended (user decision 2026-11-20) with three alternatives, and the bank states
/// each flow in a few lines;
/// the PROJECT workflow previews the worktree base write and applies it only after the pick.
#[test]
fn work_flow_row_offers_m_flow_first_and_the_bank_states_each_flow() {
    let read = |rel: &str| std::fs::read_to_string(package().join(rel)).unwrap();
    let interview = read("references/internal/interview.md");
    let rows: Vec<&str> = interview.lines().filter(|l| l.starts_with("| `")).collect();
    let at = |id: &str| {
        rows.iter()
            .position(|l| l.starts_with(&format!("| `{id}` |")))
            .unwrap_or_else(|| panic!("no `{id}` row"))
    };
    assert!(at("work_flow") > at("versioning_head"));
    let cells: Vec<String> = rows[at("work_flow")]
        .trim_matches('|')
        .split(" | ")
        .map(|c| c.trim().to_string())
        .collect();
    assert_eq!(cells[1], "always (the profile's value, else the default)");
    assert!(
        cells[3].starts_with("GitHub flow (recommended)"),
        "{cells:?}"
    );
    for alt in ["next-flow", "trunk-based", "git-flow"] {
        assert!(cells[3].contains(alt), "{alt}");
    }
    assert_eq!(cells[4], "profile `work_flow`");
    let section = &interview[interview.find("## Git work flow").expect("flow section")..];
    for text in [
        "- **next-flow — NeXT-Flow**",
        "- **GitHub flow**",
        "- **trunk-based**",
        "- **git-flow**",
        "`worktree.baseRef: head`",
        "the lane submits head and base and never re-merges a moved tip",
        "`{\"flow\": \"next-flow\" | \"github-flow\" | \"trunk-based\" | \"git-flow\"",
        // Audit B1 (decision 2026-11-17): one SPEC = one branch = one landing.
        "branch `spec/<slug>-<yyyymmdd>`",
        "one SPEC, one branch, one landing",
    ] {
        assert!(section.contains(text), "the flow section lacks {text:?}");
    }
    // A card number never goes into a shared place, so no `card/` branch is described.
    assert!(!section.contains("card/"), "{section}");
    let project = read("workflows/project.md");
    for text in [
        "the `work_flow` row",
        "`moai claude-settings-plan <project> --keep-owned`",
        "`{\"worktree\": {\"baseRef\": \"head\"}}`",
        "only after the user picks it",
        // Audit N1: a user's own `worktree` object is completed by the user, never replaced.
        "ask the user to add `\"baseRef\": \"head\"` to their own `worktree` object",
    ] {
        assert!(project.contains(text), "project.md lacks {text:?}");
    }
}

/// Card C-69: the `versioning` rows offer keeping the detected scheme first, else SemVer first,
/// and ask a format or a head only when the pick needs one; the reference they name states the
/// three schemes, and the PROJECT workflow and its chain row name that reference.
#[test]
fn versioning_rows_and_reference_cover_the_three_schemes() {
    let read = |rel: &str| std::fs::read_to_string(package().join(rel)).unwrap();
    let interview = read("references/internal/interview.md");
    let row = |id: &str| -> Vec<String> {
        interview
            .lines()
            .find(|l| l.starts_with(&format!("| `{id}` |")))
            .unwrap_or_else(|| panic!("no `{id}` row"))
            .trim_matches('|')
            .split(" | ")
            .map(|c| c.trim().to_string())
            .collect()
    };
    let versioning = row("versioning");
    assert_eq!(
        versioning[1],
        "always (the profile's value, else detection, else the default)"
    );
    let options = &versioning[3];
    // Keep-existing comes first and is the recommended one when a scheme was detected.
    let keep = options.find("keep the existing one").expect("keep option");
    let fresh = options
        .find("SemVer (recommended)")
        .expect("SemVer recommended without a detected scheme");
    assert!(keep < fresh, "{options}");
    for text in [
        "`versioning.scheme`",
        "`versioning.latest`",
        "(recommended); then the other two",
        "SemVer (recommended); HeadVer; CalVer",
        "`mixed`",
    ] {
        assert!(options.contains(text), "versioning options lack {text:?}");
    }
    let format = row("versioning_format");
    assert!(format[1].contains("CalVer"), "{format:?}");
    assert!(
        format[3].starts_with("`YYYY.0M.MICRO` (recommended"),
        "{format:?}"
    );
    // Audit N9: the two-part format is offered too.
    for text in ["`YY.0M.MICRO`", "`YYYY.0M.0D`", "`YY.0M`"] {
        assert!(format[3].contains(text), "{text}");
    }
    let head = row("versioning_head");
    assert!(head[1].contains("HeadVer"), "{head:?}");
    for text in ["plus one (recommended", "`0`"] {
        assert!(head[3].contains(text), "head options lack {text:?}");
    }
    assert!(interview.contains("`internal/versioning`"));

    let reference = read("references/internal/versioning.md");
    for text in [
        "MAJOR.MINOR.PATCH",
        "{head}.{yearweek}.{build}",
        "`YYYY.0M.MICRO`",
        "2134",
        "## Keep the existing scheme",
        "`semver`, `headver`, `calver` or `custom`",
        "https://semver.org",
        "https://github.com/line/headver",
        "https://calver.org",
    ] {
        assert!(reference.contains(text), "versioning.md lacks {text:?}");
    }
    // Each scheme has its own section with the same six points.
    for scheme in ["## SemVer", "## HeadVer", "## CalVer"] {
        let section = reference
            .split(scheme)
            .nth(1)
            .unwrap_or_else(|| panic!("no {scheme} section"))
            .split("\n## ")
            .next()
            .unwrap();
        for point in [
            "- Format:",
            "- Who bumps what:",
            "- Pick it when:",
            "- Next version:",
            "- Pre-release:",
            "- Examples:",
        ] {
            assert!(section.contains(point), "{scheme} lacks {point:?}");
        }
    }
    // Audit N9: what "Other" does on each row, and that the keep option can be corrected.
    for text in [
        "An answer through the host's \"Other\" to the `versioning` row is recorded as `{\"scheme\": \"custom\"",
        "An \"Other\" answer to `versioning_format` stays CalVer when it is a format the profile takes",
        "so the user can correct a wrong reading",
    ] {
        assert!(interview.contains(text), "interview.md lacks {text:?}");
    }
    // Audit N2 and N3: the overlap is named, and the limits are the ones the profile enforces.
    for text in [
        "Every HeadVer version is also valid SemVer",
        "at most 128 characters",
        "at most 9999",
        "version order, not date",
        "no program is started",
        // Card C-90: the credential-shape rule the profile applies to custom text.
        "a key prefix at the start of a word or a token or password assignment is refused, while a prefix inside a word, as in `task-based` or `risk-free`, is ordinary text);",
    ] {
        assert!(reference.contains(text), "versioning.md lacks {text:?}");
    }
    assert!(interview.contains(
        "the text rules it must pass, credential shapes included, are in `internal/versioning`"
    ));
    let project = read("workflows/project.md");
    // Audit B1: detection starts no program, the versioning fact included.
    for text in [
        "`internal/versioning`",
        "It never reads HOME and never starts a program, because `project.yaml` is tracked and holds project facts only. The one read outside the project is the `lsp_plugins` fact: it reads the names of the files in the PATH folders (never their content), starts nothing, writes nothing and is reported beside `detected`, never recorded. That holds for `result.versioning` too",
        "never from a git object or the git configuration",
        "is followed only for a linked worktree",
        "`tags: unavailable` with the reason",
    ] {
        assert!(project.contains(text), "project.md lacks {text:?}");
    }
    assert!(read("references/chains/index.md").contains(
        "The `versioning` rows of the PROJECT interview are asked with `internal/versioning`"
    ));
}

/// Card C-72: the HeadVer part states what each number means and who sets it, from the author's
/// article; `## At every release` is the checklist an agent runs at every release, and the SYNC
/// route, where a release is one of the effects, names it.
#[test]
fn versioning_reference_states_headver_numbers_and_the_release_checklist() {
    let read = |rel: &str| std::fs::read_to_string(package().join(rel)).unwrap();
    let reference = read("references/internal/versioning.md");
    let section = |heading: &str| -> String {
        reference
            .split(&format!("\n{heading}\n"))
            .nth(1)
            .unwrap_or_else(|| panic!("no {heading} section"))
            .split("\n## ")
            .next()
            .unwrap()
            .to_string()
    };
    let headver = section("## HeadVer");
    for text in [
        // The three numbers: who sets each and what it means.
        "`head`: set by a person",
        "raise it on every delivery that reaches users",
        "`yearweek`: when it was built",
        "computed by a program at build time, never typed",
        "`build`: which build it is",
        "the build server's counter",
        "the same code built twice is two builds",
        // Metadata, tags and how people talk about versions.
        "is not part of the version",
        "never make a commit that only edits a version number",
        "plan by head (`v3`), debug by build (`b79`)",
        "`date -u +%g%V`",
        "Observed in the author's team screenshot (the article's image 07), not a rule: development builds tagged with a `-dev` suffix",
        "A release tag on the same build number as its `-dev` tag names the same build, so it is not a reuse.",
        "MoAI's rule: a HeadVer release candidate carries no suffix; it is a build tag marked as a pre-release on the release page",
        "MoAI's rule: never use such an alias.",
        "UTC is MoAI's choice, since the article names no time zone",
    ] {
        assert!(headver.contains(text), "HeadVer section lacks {text:?}");
    }
    assert!(reference.contains(
        "https://techblog.lycorp.co.jp/ko/headver-new-versioning-system-for-product-teams"
    ));
    // Audit of C-72: whole sentences, so a negation or a dropped clause fails the test.
    let release = section("## At every release");
    for text in [
        "Run this list yourself, without being asked, whenever you prepare, tag or publish a release or a release candidate.",
        // Was: "follows the existing release rule: `SKILL.md` cases (2) and (3) and the autonomy
        // levels", where `semi` sent releases to case (4).
        "The list itself runs autonomously; pushing a tag or publishing a release is `internal/decide` case (2) at every autonomy level, `full` included: the user picks it and it waits, with no deadline default.",
        "Without the profile key, run the `versioning` interview row (detect first so the keep option is offered) and record the pick; never compute from the detection alone.",
        "never raise it silently, unless the project's rule says every release raises it.",
        "GitHub Actions keeps `run_number` on a re-run and raises only `run_attempt`, so there use `run_number` only when `run_attempt` is 1, or a dedicated counter.",
        "A person decides. Under `0.y.z` anything may change (SemVer item 4), so a breaking change may raise only MINOR.",
        "The new tag must not exist yet, and a HeadVer build number names one build only",
        "Never reuse, move or delete a tag: a wrong published version is replaced by the next version.",
        "Never require or make a version-only commit to put it in a manifest.",
        "the lock file, the changelog or release notes, install scripts and the built binary's `--version` agree with the tag.",
        "tag the merge commit on the integration branch after the merge, never on a feature, lane or release branch.",
        "when any step fails, do not tag or publish;",
    ] {
        assert!(release.contains(text), "At every release lacks {text:?}");
    }
    assert!(read("workflows/sync.md").contains(
        "When a release or release candidate is prepared, run `internal/versioning` `## At every release`"
    ));
    assert!(read("references/chains/index.md").contains(
        "- Any route that prepares, tags or publishes a release or release candidate runs `internal/versioning` `## At every release` first."
    ));
    // The MODE-LEADER pin of the same rule left with its route (SPEC-20261122-next-flow-skills-agents
    // REQ-2); SYNC and the chain rules above carry it.
}

/// #2673 AC-2: git-safety names, per host, how to enter a worktree — the built-in tool first,
/// `git -C <worktree>` where there is none — and the Claude Code guard's refused shapes.
#[test]
fn git_safety_tells_each_host_how_to_work_in_a_worktree() {
    let text =
        std::fs::read_to_string(package().join("references/internal/git-safety.md")).unwrap();
    for line in [
        "Enter and leave a worktree with the host's built-in worktree tool where it has one",
        "name the worktree on every git call as `git -C <worktree> <verb>`, one git command per call",
        "a line that commits or pushes in two folders is refused",
        "| Claude Code | `EnterWorktree` (`path` for an existing one) and `ExitWorktree`",
        "create the worktree with `git worktree add` first and enter it by `path`",
        "(`$VAR`, `$?`, `$(…)`)",
        // The `| Codex CLI |` row is retired with Codex (SPEC-20261122-next-flow-skills-agents
        // REQ-8); the Cursor row now states its git use itself.
        "| Cursor CLI | `-w` / `--worktree` at start only. | Inside a session started there, plain `git <verb>`; otherwise `git -C <worktree> <verb>`. |",
        "| Kiro CLI | None documented. | `git -C <worktree> <verb>`. |",
    ] {
        assert!(text.contains(line), "git-safety.md lacks: {line}");
    }
    assert!(
        !text.contains("| Codex CLI |"),
        "git-safety.md keeps a Codex row"
    );
}

/// Card C-77: the pre-PR self-check list ships in the package, holds the rule that turned pull
/// requests red most often on their own (the package file bound of `profile::LIMIT`), and is
/// named by `moai-impl` (its `reads:` and its Verification) and by the rules every chain follows.
///
/// SPEC-20261125-run-sync-procedure-receipt (D-5) retired the needles of rules that decisions of
/// 2026-11-23 to 2026-11-25 overturned: the whole `--bin moai` and `--bin moai-cli` runs, the
/// `--test git_env_scrub`, `--test home_isolation_scan` and `--test ac_command_lint` items and
/// `rustfmt --check src/ops` (the receipt of RUN step 7 is the one set before READY), the
/// in-repository runner `tools/self_check.sh` with `--full-python` and its `exit=75` line, the
/// Python `tools/source_manifest.py` pair (the manifest hash is `shasum -a 256`), the
/// `profile::LIMIT` wording of the size item (the bound itself is still read from the code
/// below), and "at most 5 for audit tier 1 and 12 for tier 2" (mutants are asked for at audit
/// tier 2 only).
#[test]
fn self_check_lists_the_pre_pr_checks_and_is_named_by_impl_and_the_chains() {
    let read = |rel: &str| std::fs::read_to_string(package().join(rel)).unwrap();
    let text = read("references/internal/self-check.md");
    let manifest = read("references/source-manifest.json");
    assert!(manifest.contains("\"references/internal/self-check.md\""));
    for needle in [
        "`moai package-inspect skills/moai` exits 0 and no package file is over 65,536 bytes",
        "`cargo clippy --offline --locked --all-targets -- -D warnings`",
        "`cargo fmt --check`",
        "`moai pr-self-audit-check --file BODY.md`",
        "`audit-tier: ",
        "skip it with the reason",
        // Audit of 421b11a2c9, N1-N6.
        "`git diff --check <base>...HEAD`",
        "no commit message, branch name or pull request title carries a card number",
        // The receipt of RUN step 7, the mutants of audit tier 2, and the manifest hash.
        "Receipt: run the `verify:` commands of the SPEC's criteria, then the receipt of `workflows/run.md`, step 7.",
        "at audit tier 2 only, one mutation per new guard, the riskiest first, at most 12; below tier 2 none is asked for",
        "`shasum -a 256`",
    ] {
        assert!(text.contains(needle), "self-check.md lacks {needle:?}");
    }
    // The bound in the text is the bound in the code.
    let profile =
        std::fs::read_to_string(Path::new(env!("CARGO_MANIFEST_DIR")).join("src/profile.rs"))
            .unwrap();
    assert!(profile.contains("pub const LIMIT: u64 = 65536;"));

    let defs = package_agents(&agents(), &package()).unwrap();
    let imp = defs.iter().find(|d| d.name == "moai-impl").unwrap();
    assert!(imp.reads.contains(&"internal/self-check".to_string()));
    assert!(
        imp.body.lines().any(|l| l.starts_with("- ")
            && l.contains("`internal/self-check`")
            && l.contains("`Not run`")),
        "moai-impl Verification does not name internal/self-check"
    );
    for d in defs.iter().filter(|d| d.name != "moai-impl") {
        assert!(
            !d.reads.contains(&"internal/self-check".to_string()),
            "{}",
            d.name
        );
    }
    // Audit N4: the list sends moai-impl only to references its `reads:` gives it, since the
    // rendered agent says "Read no other internal reference".
    let mut named = 0;
    for (at, _) in text.match_indices("`internal/") {
        let entry = &text[at + 1..];
        let entry = &entry[..entry.find('`').unwrap()];
        named += 1;
        assert!(
            imp.reads.iter().any(|r| r == entry),
            "self-check.md names {entry}, which moai-impl does not read"
        );
    }
    assert!(named >= 1, "the internal/ mentions moved");
    let chains = read("references/chains/index.md");
    let rules = chains.split("Rules for every chain:").nth(1).unwrap();
    assert!(
        rules.lines().any(|l| l.starts_with("- ")
            && l.contains("`internal/self-check`")
            && l.contains("before")
            && l.contains("audit")),
        "chains/index.md rules do not name internal/self-check"
    );
}

/// The words that send a reader to where the build rule lives now: by path from an agent or an
/// internal reference, by name from another workflow (a workflow names no package path).
const BUILD_RULE_POINTER: &str = "(`workflows/run.md`, step 1)";
const BUILD_RULE_WORKFLOW_POINTER: &str = "(the RUN workflow, step 1)";

/// A user decision of 2026-11-17 wrote the build rule as the reference
/// `internal/build-folder`; SPEC-20261122-next-flow-skills-agents slice C (D-14, the user's build
/// rule of 2026-11-22) removed that reference and rewrote the rule in RUN's "Check-run rules";
/// SPEC-20261125-run-sync-procedure-receipt moved it to `internal/build-slots`, which RUN step 1
/// names, so the agents and references point to "step 1" now.
/// The pins whose rule still holds moved with it: one build folder per worktree, never shared;
/// the end-of-task cleanup of the cargo output the worker made; evidence kept apart from that
/// cleanup; `CARGO_INCREMENTAL=0` in the builder agents; leftover agent build folders removed
/// only as the rows the user picks (UPDATE). Retired, because each worktree now builds in its
/// own folder and nothing is built in the real HOME: the folder per worktree in the home state
/// folder, its owner record `.moai-build.json`, the refusal when HOME is not usable, and the
/// HOME exception of the builder agents with its 0700 parents.
#[test]
fn build_rule_lives_in_run_and_the_builders_point_to_it() {
    let read = |rel: &str| std::fs::read_to_string(package().join(rel)).unwrap();
    assert!(!package()
        .join("references/internal/build-folder.md")
        .exists());
    assert!(!read("references/source-manifest.json").contains("internal/build-folder"));
    let run = read("workflows/run.md");
    // RUN's step 1 sends the reader to the reference that holds the rule.
    let step1 = run
        .lines()
        .find(|l| l.starts_with("1. "))
        .expect("run.md has a step 1");
    assert!(
        step1.contains("`internal/build-slots`"),
        "run.md step 1 does not name internal/build-slots: {step1}"
    );
    let slots = read("references/internal/build-slots.md");
    let section = |name: &str| -> String {
        slots
            .split(&format!("\n## {name}\n"))
            .nth(1)
            .and_then(|s| s.split("\n## ").next())
            .unwrap_or_else(|| panic!("build-slots.md has no `## {name}` section"))
            .to_string()
    };
    for (name, needles) in [
        (
            "Worktrees and build folders",
            [
                "Use one build folder per worktree",
                "never share it with another worktree",
            ],
        ),
        (
            "Cleanup",
            [
                "a lane removes only the cargo output it made itself",
                "the notes, evidence and scripts under `target/<name>/`",
            ],
        ),
    ] {
        let rules = section(name);
        for needle in needles {
            assert!(
                rules.contains(needle),
                "build-slots.md \"{name}\" lacks {needle:?}"
            );
        }
    }
    assert!(
        read("workflows/update.md").contains("go only as the rows the user picks"),
        "update.md lacks the user-picked removal"
    );
    let home_build = format!("~/{}/build", moai_cli::state_root::MOAI_DIR);
    let defs = package_agents(&agents(), &package()).unwrap();
    for d in &defs {
        assert!(
            !d.reads.iter().any(|r| r == "internal/build-folder"),
            "{}",
            d.name
        );
    }
    for name in ["moai-impl", "moai-audit"] {
        let d = defs.iter().find(|d| d.name == name).unwrap();
        assert!(d.body.contains("`CARGO_INCREMENTAL=0`"), "{name}");
        assert!(d.body.contains(BUILD_RULE_POINTER), "{name}");
        assert!(!d.body.contains(&home_build), "{name} names {home_build}");
        for host in RENDERED_HOSTS {
            let text = render(d, host).unwrap().text;
            assert!(text.contains(BUILD_RULE_POINTER), "{name} on {host}");
        }
    }
    for (rel, pointer) in [
        ("references/internal/git-safety.md", BUILD_RULE_POINTER),
        ("references/internal/self-check.md", BUILD_RULE_POINTER),
        // UPDATE carries the retired OPTIMIZE route (SPEC-20261122-next-flow-skills-agents).
        ("workflows/update.md", BUILD_RULE_WORKFLOW_POINTER),
    ] {
        let text = read(rel);
        assert!(text.contains(pointer), "{rel}");
        assert!(!text.contains("internal/build-folder"), "{rel}");
    }
    // RUN holds the rule itself; its other sections name it without the removed reference.
    assert!(!run.contains("internal/build-folder"), "workflows/run.md");
}

/// SPEC-20261124-check-affected-touch AC-5 (REQ-4, REQ-5): the guard-mutant line of the
/// self-check page tells a lane how to restore a mutated file: `git checkout` or a plain `cp`,
/// never `cp -p`, with the reason (the old file time makes cargo keep the mutant build).
#[test]
fn the_guard_mutant_line_names_the_restore_rule() {
    let page =
        std::fs::read_to_string(package().join("references/internal/self-check.md")).unwrap();
    let line = page
        .lines()
        .find(|l| l.starts_with("2. Guards:") && l.contains("guard-mutant:"))
        .expect("the Guards line of the Tests list");
    for needle in [
        "`git checkout -- <path>`",
        "plain `cp`",
        "never `cp -p`",
        "file time",
    ] {
        assert!(
            line.contains(needle),
            "the Guards line lacks {needle}: {line}"
        );
    }
}
