//! #2592 S1, #2721: the base agents ship as tool-neutral sources in `agents/`, outside the
//! `moai` skill tree, and render into each host's agent file. Reads the sources and the
//! package; writes nothing.

use moai_cli::agent_defs::{package_agents, render, HOSTS, SOURCE_DIR};
use std::path::Path;

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
    // implementation and audit run isolated where the host can.
    for n in ["moai-audit", "moai-research", "moai-plan"] {
        assert!(get(n).readonly, "{n}");
    }
    for n in ["moai-impl", "moai-docs"] {
        assert!(!get(n).readonly, "{n}");
    }
    assert_eq!(get("moai-research").shape, "fanout");
    for n in ["moai-impl", "moai-audit"] {
        assert_eq!(get(n).isolation, "worktree", "{n}");
    }
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
    // #2633: OMP is rendered and installed (2026-11-13 confirmation of five supported hosts;
    // Kiro and Cursor CLI follow).
    assert_eq!(HOSTS, ["claude", "codex", "omp"]);
    let defs = package_agents(&agents(), &package()).unwrap();
    for d in &defs {
        for host in HOSTS {
            let a = render(d, host).unwrap();
            assert_eq!(a, render(d, host).unwrap());
            assert!(a.path.starts_with("agents/moai-"), "{}", a.path);
            assert!(a.text.contains(&d.name));
        }
        let codex = render(d, "codex").unwrap();
        // Codex cannot enforce read-only or isolation: named, never silently dropped.
        assert_eq!(
            codex
                .not_supported
                .iter()
                .any(|n| n.starts_with("readonly")),
            d.readonly,
            "{}",
            d.name
        );
        assert_eq!(
            codex
                .not_supported
                .iter()
                .any(|n| n.starts_with("isolation")),
            d.isolation == "worktree",
            "{}",
            d.name
        );
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
    assert_eq!(
        reads("moai-plan"),
        ["internal/decide", "internal/interview"]
    );
    assert_eq!(reads("moai-research"), ["adapters/aside", "adapters/orca"]);
    assert_eq!(
        reads("moai-impl"),
        [
            "internal/test-design",
            "internal/git-safety",
            "internal/tdd-results"
        ]
    );
    assert_eq!(
        reads("moai-audit"),
        [
            "internal/review",
            "internal/test-design",
            "internal/check-completeness",
            "internal/boundary-checks",
            "internal/security-checklist"
        ]
    );
    assert_eq!(
        reads("moai-docs"),
        ["internal/decide", "internal/doc-claims"]
    );
    for d in &defs {
        for entry in &d.reads {
            assert!(
                package().join(format!("references/{entry}.md")).is_file(),
                "{entry}"
            );
        }
        for host in HOSTS {
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
        "review",
        "decide",
        "interview",
        "test-design",
        "git-safety",
        // #2582: adopted from the Go generation (leader confirmation 6856677422).
        "tdd-results",
        "check-completeness",
        "boundary-checks",
        "doc-claims",
        "security-checklist",
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
    assert_eq!(rows.len(), 8);
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
        "PLAN",
        "RUN",
        "AUDIT",
        "READ",
        "VERIFY",
        "PROJECT",
        "SYNC",
        "MODE-AGENT",
    ] {
        assert!(covered.contains(&route), "{route} has no chain");
    }
    for rule in [
        "The parent session starts each agent in turn",
        "An agent never starts another agent.",
        "The agent that implemented a change never audits it.",
        "the parent session, not the agent, writes the verdict",
    ] {
        assert!(text.contains(rule), "{rule}");
    }
    // The build chain runs implementation before the audit, never in the same agent.
    let build = rows.iter().find(|r| r[0] == "Build something new").unwrap();
    assert_eq!(build[2], "moai-plan, moai-impl, moai-audit, moai-docs");
}

/// The question bank (#2639): one row per question, every always-asked axis of the PROJECT
/// workflow is in it, and every answer lands in `decisions` or a real profile field.
#[test]
fn interview_question_bank_covers_the_project_axes() {
    let text = std::fs::read_to_string(package().join("references/internal/interview.md")).unwrap();
    let bank = text
        .split("## Question bank")
        .nth(1)
        .expect("question bank");
    let rows: Vec<Vec<String>> = bank
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
            "purpose",
            "protected",
            "done",
            "review_blocks",
            "autonomy",
            "test_command",
            "build_command",
            "stack",
            "method",
            "hosts"
        ]
    );
    for r in &rows {
        assert_eq!(r.len(), 4, "{r:?}");
        let always = r[1] == "always";
        let id = r[0].trim_matches('`');
        assert_eq!(
            always,
            ["protected", "done", "review_blocks", "autonomy"].contains(&id),
            "{id}"
        );
        let to = r[3].as_str();
        assert!(
            to == "decisions"
                || [
                    "profile `purpose`",
                    "profile `tools`",
                    "profile `development_methods`",
                    "profile `hosts`"
                ]
                .contains(&to),
            "{id}: {to}"
        );
    }
    assert!(bank.contains("skip every row that reading already answered"));
    assert!(bank.contains("\"source\": \"repo\""));
    assert!(bank.contains("\"source\": \"user\""));
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
        "| Codex CLI |",
        "| Cursor CLI |",
        "| OMP | None for the main session",
        "| Kiro CLI | None documented. | `git -C <worktree> <verb>`. |",
    ] {
        assert!(text.contains(line), "git-safety.md lacks: {line}");
    }
}
