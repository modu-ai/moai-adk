//! #2592 S1: the base agents ship in the `moai` skill package as tool-neutral sources under
//! `references/agents/` (decision A) and render into each host's agent file. Reads the package;
//! writes nothing.

use moai_cli::agent_defs::{package_agents, render, HOSTS};
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
    let defs = package_agents(&package()).unwrap();
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
