//! #2638: the MoAI routing section in the global instruction files, against a temporary HOME
//! only — never the real one. Covers create, idempotence, the user's bytes outside the block,
//! conflicts, symlinks, a stale digest, revert, and the hosts reported as not written.

mod common;
use moai_cli::home_routing::{apply, plan, render, revert, HOSTS, SECTION, SOURCE};
use std::fs;
use std::path::{Path, PathBuf};

const PERM: &str = "test-permission";

fn package() -> PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai")
}

fn home() -> (tempfile::TempDir, PathBuf) {
    let dir = tempfile::tempdir().unwrap();
    let path = fs::canonicalize(dir.path()).unwrap();
    (dir, path)
}

fn file(home: &Path, host: &str) -> PathBuf {
    home.join(HOSTS.iter().find(|(h, _, _)| *h == host).unwrap().1)
}

fn action(p: &moai_cli::home_routing::Plan, host: &str) -> String {
    let h = p.hosts.iter().find(|h| h.host == host).unwrap();
    match &h.section {
        Some(s) => s.action.clone(),
        None => h.preflight.clone(),
    }
}

fn block(host: &str) -> String {
    let src = fs::read_to_string(package().join(SOURCE)).unwrap();
    render(&src, host).unwrap()
}

#[test]
fn fresh_home_gets_one_block_per_host_and_a_second_run_is_unchanged() {
    let (_t, home) = home();
    let p = plan(&home, &package()).unwrap();
    assert_eq!(action(&p, "claude"), "create");
    assert_eq!(action(&p, "codex"), "create");
    let hosts: Vec<&str> = p.not_supported.iter().map(|u| u.host.as_str()).collect();
    assert_eq!(hosts, ["omp", "kiro", "cursor"]);
    assert!(p.not_supported.iter().all(|u| !u.reason.is_empty()));
    let r = apply(&home, &package(), &p.plan_digest, PERM).unwrap();
    assert_eq!(r["state"], "applied");
    let claude = fs::read_to_string(file(&home, "claude")).unwrap();
    let codex = fs::read_to_string(file(&home, "codex")).unwrap();
    assert_eq!(claude, block("claude"));
    assert_eq!(codex, block("codex"));
    assert!(claude.contains("(/moai)") && codex.contains("($moai)"));
    assert!(claude.starts_with(&format!("<!-- moai:managed:begin id={SECTION} -->\n")));
    // `~/.agents/AGENTS.md` is never created (user decision 6845654776).
    assert!(!home.join(".agents/AGENTS.md").exists());
    let again = plan(&home, &package()).unwrap();
    assert_eq!(action(&again, "claude"), "unchanged");
    assert_eq!(action(&again, "codex"), "unchanged");
    let r = apply(&home, &package(), &again.plan_digest, PERM).unwrap();
    assert_eq!(r["state"], "unchanged");
    assert_eq!(fs::read_to_string(file(&home, "claude")).unwrap(), claude);
}

#[test]
fn existing_files_keep_every_byte_outside_the_block() {
    let (_t, home) = home();
    fs::create_dir_all(home.join(".claude")).unwrap();
    fs::create_dir_all(home.join(".codex")).unwrap();
    // No trailing newline, and a file that already ends with a blank line.
    let claude_user = "# Mine\n\nKeep this line.";
    let codex_user = "Codex notes\n\n";
    fs::write(file(&home, "claude"), claude_user).unwrap();
    fs::write(file(&home, "codex"), codex_user).unwrap();
    let p = plan(&home, &package()).unwrap();
    assert_eq!(action(&p, "claude"), "append");
    apply(&home, &package(), &p.plan_digest, PERM).unwrap();
    let claude = fs::read_to_string(file(&home, "claude")).unwrap();
    assert_eq!(claude, format!("{claude_user}\n\n{}", block("claude")));
    let codex = fs::read_to_string(file(&home, "codex")).unwrap();
    assert_eq!(codex, format!("{codex_user}{}", block("codex")));
    // The user adds text after the block; a later run keeps it and changes nothing.
    fs::write(file(&home, "claude"), format!("{claude}\nAfter.\n")).unwrap();
    let p = plan(&home, &package()).unwrap();
    assert_eq!(action(&p, "claude"), "unchanged");
}

/// A temporary package holding only what a plan reads: `SKILL.md` and the section source.
fn temp_package(source: &str) -> tempfile::TempDir {
    let pkg = tempfile::tempdir().unwrap();
    fs::write(pkg.path().join("SKILL.md"), "package\n").unwrap();
    let src = pkg.path().join(SOURCE);
    fs::create_dir_all(src.parent().unwrap()).unwrap();
    fs::write(&src, source).unwrap();
    pkg
}

#[test]
fn an_edited_block_is_kept_and_conflicts_once_the_source_changes() {
    let (_t, home) = home();
    let pkg = temp_package("## MoAI\n\nUse the skill ({invoke}).\n");
    let p = plan(&home, pkg.path()).unwrap();
    apply(&home, pkg.path(), &p.plan_digest, PERM).unwrap();
    let edited = fs::read_to_string(file(&home, "claude"))
        .unwrap()
        .replace("Use the skill", "Use my own skill");
    fs::write(file(&home, "claude"), &edited).unwrap();
    let codex = fs::read_to_string(file(&home, "codex")).unwrap();
    let broken = format!("{codex}{codex}");
    fs::write(file(&home, "codex"), &broken).unwrap();
    // Same source: the user's edit is kept, nothing is written; doubled markers conflict.
    let p = plan(&home, pkg.path()).unwrap();
    assert_eq!(action(&p, "claude"), "unchanged");
    assert_eq!(action(&p, "codex"), "conflict");
    // A new source and an edited block: a conflict, never an overwrite.
    fs::write(
        pkg.path().join(SOURCE),
        "## MoAI\n\nUse the moai skill ({invoke}).\n",
    )
    .unwrap();
    let p = plan(&home, pkg.path()).unwrap();
    assert_eq!(action(&p, "claude"), "conflict");
    assert_eq!(action(&p, "codex"), "conflict");
    let r = apply(&home, pkg.path(), &p.plan_digest, PERM).unwrap();
    assert_eq!(r["state"], "unchanged");
    assert_eq!(fs::read_to_string(file(&home, "claude")).unwrap(), edited);
    assert_eq!(fs::read_to_string(file(&home, "codex")).unwrap(), broken);
}

#[cfg(unix)]
#[test]
fn a_symlinked_file_or_folder_blocks_its_host() {
    let (_t, home) = home();
    let elsewhere = home.join("elsewhere");
    fs::create_dir_all(&elsewhere).unwrap();
    fs::write(elsewhere.join("CLAUDE.md"), "theirs").unwrap();
    fs::create_dir_all(home.join(".claude")).unwrap();
    std::os::unix::fs::symlink(elsewhere.join("CLAUDE.md"), file(&home, "claude")).unwrap();
    std::os::unix::fs::symlink(&elsewhere, home.join(".codex")).unwrap();
    let p = plan(&home, &package()).unwrap();
    assert_eq!(action(&p, "claude"), "blocked");
    assert_eq!(action(&p, "codex"), "blocked");
    let r = apply(&home, &package(), &p.plan_digest, PERM).unwrap();
    assert_eq!(r["state"], "unchanged");
    assert_eq!(
        fs::read_to_string(elsewhere.join("CLAUDE.md")).unwrap(),
        "theirs"
    );
    assert!(!elsewhere.join("AGENTS.md").exists());
}

/// Audit FAIL 6846213349 B1 (#2650): after the apply, both host folders are moved outside HOME
/// and linked back. Revert must skip them — never remove the created Codex file or restore the
/// Claude file through the link.
#[cfg(unix)]
#[test]
fn revert_never_writes_through_a_folder_linked_after_the_apply() {
    let (_t, home) = home();
    let outside = tempfile::tempdir().unwrap();
    let outside = fs::canonicalize(outside.path()).unwrap();
    fs::create_dir_all(home.join(".claude")).unwrap();
    fs::write(file(&home, "claude"), "mine\n").unwrap();
    let p = plan(&home, &package()).unwrap();
    let r = apply(&home, &package(), &p.plan_digest, PERM).unwrap();
    let id = r["receipt"].as_str().unwrap().to_string();
    let applied_claude = fs::read_to_string(file(&home, "claude")).unwrap();
    let applied_codex = fs::read_to_string(file(&home, "codex")).unwrap();
    for dir in [".claude", ".codex"] {
        fs::rename(home.join(dir), outside.join(dir)).unwrap();
        std::os::unix::fs::symlink(outside.join(dir), home.join(dir)).unwrap();
    }
    let v = revert(&home, &id, PERM).unwrap();
    assert_eq!(v["reverted"], serde_json::json!([]));
    let skipped = v["skipped"].as_array().unwrap();
    assert_eq!(skipped.len(), 2);
    assert!(skipped.iter().all(|s| s["reason"]
        .as_str()
        .unwrap()
        .starts_with("host_folder_not_plain")));
    assert_eq!(
        fs::read_to_string(outside.join(".claude/CLAUDE.md")).unwrap(),
        applied_claude
    );
    assert_eq!(
        fs::read_to_string(outside.join(".codex/AGENTS.md")).unwrap(),
        applied_codex
    );
}

#[test]
fn a_stale_digest_or_a_missing_permission_is_refused() {
    let (_t, home) = home();
    let p = plan(&home, &package()).unwrap();
    assert!(apply(&home, &package(), &p.plan_digest, "")
        .unwrap_err()
        .contains("permission"));
    fs::create_dir_all(home.join(".claude")).unwrap();
    fs::write(file(&home, "claude"), "changed after the preview\n").unwrap();
    assert!(apply(&home, &package(), &p.plan_digest, PERM)
        .unwrap_err()
        .contains("stale"));
    assert_eq!(
        fs::read_to_string(file(&home, "claude")).unwrap(),
        "changed after the preview\n"
    );
}

#[test]
fn revert_restores_only_files_still_holding_what_the_apply_wrote() {
    let (_t, home) = home();
    fs::create_dir_all(home.join(".claude")).unwrap();
    fs::write(file(&home, "claude"), "mine\n").unwrap();
    let p = plan(&home, &package()).unwrap();
    let r = apply(&home, &package(), &p.plan_digest, PERM).unwrap();
    let id = r["receipt"].as_str().unwrap().to_string();
    // The user edits the Codex file after the apply: it is skipped, Claude is restored.
    fs::write(file(&home, "codex"), "user wrote this\n").unwrap();
    let v = revert(&home, &id, PERM).unwrap();
    assert_eq!(v["reverted"], serde_json::json!(["claude"]));
    assert_eq!(v["skipped"][0]["reason"], "user_changed_since_apply");
    assert_eq!(fs::read_to_string(file(&home, "claude")).unwrap(), "mine\n");
    assert_eq!(
        fs::read_to_string(file(&home, "codex")).unwrap(),
        "user wrote this\n"
    );
    assert!(revert(&home, &id, PERM).unwrap_err().contains("already"));
    assert!(revert(&home, "agents-1-0", PERM).is_err());
    // After the revert the Claude file has no block and no ownership: append again.
    assert_eq!(
        action(&plan(&home, &package()).unwrap(), "claude"),
        "append"
    );
}

#[test]
fn a_changed_source_merges_and_its_revert_restores_the_earlier_block() {
    let (_t, home) = home();
    let pkg = temp_package("## MoAI\n\nUse the skill ({invoke}).\n");
    let src = pkg.path().join(SOURCE);
    let p = plan(&home, pkg.path()).unwrap();
    apply(&home, pkg.path(), &p.plan_digest, PERM).unwrap();
    let first = fs::read_to_string(file(&home, "claude")).unwrap();
    fs::write(&src, "## MoAI\n\nUse the moai skill ({invoke}) first.\n").unwrap();
    let p = plan(&home, pkg.path()).unwrap();
    assert_eq!(action(&p, "claude"), "merge");
    let r = apply(&home, pkg.path(), &p.plan_digest, PERM).unwrap();
    assert!(fs::read_to_string(file(&home, "claude"))
        .unwrap()
        .contains("first."));
    revert(&home, r["receipt"].as_str().unwrap(), PERM).unwrap();
    assert_eq!(fs::read_to_string(file(&home, "claude")).unwrap(), first);
    // The ownership went back to the earlier block too, so the newer source merges again —
    // without it the restored block would read as a user edit and conflict.
    assert_eq!(action(&plan(&home, pkg.path()).unwrap(), "claude"), "merge");
}

#[test]
fn a_bad_section_source_is_refused() {
    let (_t, home) = home();
    let empty = tempfile::tempdir().unwrap();
    assert!(plan(&home, empty.path()).unwrap_err().contains("missing"));
    // A package that is otherwise complete, so only the source check can refuse it.
    let pkg = temp_package("ok ({invoke})\n");
    assert!(plan(&home, pkg.path()).is_ok());
    for bad in [
        "no placeholder\n",
        "{invoke} {invoke}\n",
        "{invoke} <!-- x -->\n",
        "{invoke} moai:managed:end id=moai-routing\n",
        "{invoke} no newline",
    ] {
        fs::write(pkg.path().join(SOURCE), bad).unwrap();
        let e = plan(&home, pkg.path()).unwrap_err();
        assert!(e.contains("must hold one"), "{bad:?}: {e}");
    }
}

/// The shipped section is a short pointer: it names the package files it points to, and
/// each of them exists; it does not copy the chain table.
#[test]
fn the_shipped_section_points_at_real_package_files() {
    let src = fs::read_to_string(package().join(SOURCE)).unwrap();
    for rel in ["references/routes.json", "references/chains/index.md"] {
        assert!(src.contains(&format!("`{rel}`")), "{rel}");
        assert!(package().join(rel).is_file(), "{rel}");
    }
    assert!(
        !src.contains("| Request |"),
        "the chain table is not copied"
    );
    assert!(src.contains("The agent that implemented a change never audits it."));
    for (host, _, invoke) in HOSTS {
        assert!(render(&src, host).unwrap().contains(&format!("({invoke})")));
    }
}

/// The three operations, through the binary, against a temporary HOME passed as an argument.
#[test]
fn the_operations_run_through_the_binary() {
    let (_t, home) = home();
    let bin = crate::common::MOAI_BIN;
    let run = |args: &[&str]| {
        let out = std::process::Command::new(bin)
            .args(args)
            .env("HOME", &home)
            .output()
            .unwrap();
        assert!(
            out.status.success(),
            "{}",
            String::from_utf8_lossy(&out.stderr)
        );
        serde_json::from_slice::<serde_json::Value>(&out.stdout).unwrap()
    };
    let h = home.to_str().unwrap();
    let pkg = package();
    let pkg = pkg.to_str().unwrap();
    let p = run(&["home-routing-plan", h, pkg]);
    assert_eq!(p["operation"], "home.routing_plan");
    let digest = p["result"]["plan_digest"].as_str().unwrap().to_string();
    let a = run(&[
        "home-routing-apply",
        h,
        pkg,
        "--plan-digest",
        &digest,
        "--permission-reference",
        PERM,
    ]);
    let id = a["result"]["receipt"].as_str().unwrap().to_string();
    let r = run(&[
        "home-routing-revert",
        h,
        "--receipt",
        &id,
        "--permission-reference",
        PERM,
    ]);
    assert_eq!(r["result"]["state"], "reverted");
    assert!(!file(&home, "claude").exists());
}
