//! #2592 S1: the base agents ship in the `moai` skill package as tool-neutral sources under
//! `references/agents/` (decision A) and render into each host's agent file. Reads the package;
//! writes nothing.

use moai_cli::agent_defs::{package_agents, render, HOSTS, RENDERED_HOSTS};
use std::path::Path;

fn package() -> std::path::PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai")
}

#[test]
fn the_package_holds_the_five_roles_of_the_contract() {
    let defs = package_agents(&package()).unwrap();
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
    // #2633: OMP is rendered (2026-11-13 confirmation of five supported hosts; Kiro and Cursor
    // CLI follow) but joins the installed hosts only with its folder in the installer.
    assert_eq!(RENDERED_HOSTS, ["claude", "codex", "omp"]);
    assert_eq!(HOSTS, ["claude", "codex"]);
    let defs = package_agents(&package()).unwrap();
    for d in &defs {
        for host in RENDERED_HOSTS {
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
    assert!(package_agents(tmp.path()).unwrap().is_empty());
    let dir = tmp.path().join("references/agents");
    std::fs::create_dir_all(&dir).unwrap();
    std::fs::write(dir.join("notes.txt"), "x").unwrap();
    assert!(package_agents(tmp.path())
        .unwrap_err()
        .contains("only regular .md files"));
    std::fs::remove_file(dir.join("notes.txt")).unwrap();
    std::fs::write(
        dir.join("impl.md"),
        "---\nname: moai-other\ndescription: d\nreadonly: false\nisolation: none\nshape: sequential\n---\n\nx\n",
    )
    .unwrap();
    assert!(package_agents(tmp.path())
        .unwrap_err()
        .contains("name must be moai-impl"));
}

/// #2637: each base agent's `reads:` list is the final design's §3 table (#2634), with the
/// adapters of `moai-research` left to #2636; every entry is a file of the package and every
/// rendered agent opens with the generated line that names exactly those files.
#[test]
fn reads_lists_match_the_design_and_render_on_every_host() {
    let defs = package_agents(&package()).unwrap();
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
    assert_eq!(reads("moai-research"), Vec::<&str>::new());
    assert_eq!(
        reads("moai-impl"),
        ["internal/test-design", "internal/git-safety"]
    );
    assert_eq!(
        reads("moai-audit"),
        ["internal/review", "internal/test-design"]
    );
    assert_eq!(reads("moai-docs"), ["internal/decide"]);
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
    let dir = tmp.path().join("references/agents");
    std::fs::create_dir_all(&dir).unwrap();
    std::fs::write(
        dir.join("impl.md"),
        "---\nname: moai-impl\ndescription: d\nreadonly: false\nisolation: none\nshape: sequential\nreads: internal/test-design\n---\n\nx\n",
    )
    .unwrap();
    assert!(package_agents(tmp.path())
        .unwrap_err()
        .contains("reads entry internal/test-design has no file"));
    std::fs::create_dir_all(tmp.path().join("references/internal")).unwrap();
    std::fs::write(tmp.path().join("references/internal/test-design.md"), "t").unwrap();
    assert_eq!(package_agents(tmp.path()).unwrap().len(), 1);
}

/// The internal references (#2637): one short note each, opening with the line that says
/// when to read it, and every one of them read by some agent or by a chain.
#[test]
fn internal_references_say_when_to_read_them() {
    let names = ["review", "decide", "interview", "test-design", "git-safety"];
    let mut found: Vec<String> = std::fs::read_dir(package().join("references/internal"))
        .unwrap()
        .map(|e| e.unwrap().file_name().to_string_lossy().into_owned())
        .collect();
    found.sort();
    let mut want: Vec<String> = names.iter().map(|n| format!("{n}.md")).collect();
    want.sort();
    assert_eq!(found, want);
    let defs = package_agents(&package()).unwrap();
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
    let agents: Vec<String> = package_agents(&package())
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
