//! SPEC-20261117-installer-skill-agents REQ-10: `moai cc` and `moai glm` warn on stderr when
//! the host's copy of the MoAI skill has no `SKILL.md`, name the folder and offer the
//! `moai update` preview; the host still starts. Every launch runs in a temporary project with a
//! temporary HOME and a stub `claude` executable on PATH; no real host starts.
#![cfg(unix)]
mod common;
use std::os::unix::fs::PermissionsExt;
use std::{
    fs,
    path::Path,
    process::{Command, Output, Stdio},
};

const WARNING: &str = "the MoAI skill is not installed for";

struct Fixture {
    bin: tempfile::TempDir,
    home: tempfile::TempDir,
    project: tempfile::TempDir,
}

fn fixture() -> Fixture {
    let bin = tempfile::tempdir().unwrap();
    let path = bin.path().join("claude");
    fs::write(
        &path,
        "#!/bin/sh\nprintf started > \"$MOAI_TEST_HOST_MARKER\"\n",
    )
    .unwrap();
    fs::set_permissions(&path, fs::Permissions::from_mode(0o755)).unwrap();
    // A project that already has its settings: the launch initializes nothing, so the only
    // launcher line on stderr is the one under test.
    let project = tempfile::tempdir().unwrap();
    let ok = common::git::command()
        .args(["init", "-q"])
        .current_dir(project.path())
        .status()
        .unwrap()
        .success();
    assert!(ok);
    fs::create_dir(project.path().join(".moai-cli")).unwrap();
    fs::write(project.path().join(".moai-cli/project.yaml"), "kept\n").unwrap();
    Fixture {
        bin,
        home: tempfile::tempdir().unwrap(),
        project,
    }
}

fn launch(f: &Fixture, host: &str) -> Output {
    let marker = f.home.path().join("host-started");
    let _ = fs::remove_file(&marker);
    let out = Command::new(common::MOAI_BIN)
        .arg(host)
        .current_dir(f.project.path())
        .env("PATH", format!("{}:/usr/bin:/bin", f.bin.path().display()))
        .env("HOME", f.home.path())
        .env("USERPROFILE", f.home.path())
        .env("MOAI_TEST_HOST_MARKER", &marker)
        .env("ANTHROPIC_AUTH_TOKEN", "moai-test-dummy-not-real")
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .env_remove("CLAUDE_CODE_DISABLE_ADVISOR_TOOL")
        .stdin(Stdio::null())
        .output()
        .unwrap();
    assert!(out.status.success(), "{}", stderr(&out));
    // The warning never stops the host.
    assert!(marker.exists(), "host must start: {}", stderr(&out));
    out
}

fn stderr(out: &Output) -> String {
    String::from_utf8_lossy(&out.stderr).into_owned()
}

fn put_skill(home: &Path, rel: &str) {
    let dir = home.join(rel);
    fs::create_dir_all(&dir).unwrap();
    fs::write(dir.join("SKILL.md"), "skill\n").unwrap();
}

fn warnings(out: &Output) -> Vec<String> {
    stderr(out)
        .lines()
        .filter(|l| l.contains(WARNING))
        .map(str::to_string)
        .collect()
}

#[test]
fn a_missing_claude_skill_is_named_with_its_folder_and_the_update_preview() {
    let f = fixture();
    let out = launch(&f, "cc");
    let lines = warnings(&out);
    assert_eq!(lines.len(), 1, "{}", stderr(&out));
    let line = &lines[0];
    assert!(line.contains("Claude Code"), "{line}");
    let folder = f.home.path().join(".claude/skills/moai");
    assert!(line.contains(&folder.display().to_string()), "{line}");
    assert!(line.contains("moai update"), "{line}");
    // Nothing is written into HOME's host folders by the check.
    assert!(!f.home.path().join(".claude").exists());
}

/// SPEC-20261121-codex-launcher-and-models-removed REQ-9: no `codex` host is known, so the check
/// has no Codex folder to look in. `moai codex` is an unknown command: no skill line, no host
/// start, and a skill copy under `.agents/skills` is neither read nor counted for any host.
#[test]
fn codex_is_an_unknown_host_for_the_skill_check() {
    let f = fixture();
    let marker = f.home.path().join("host-started");
    let out = Command::new(common::MOAI_BIN)
        .arg("codex")
        .current_dir(f.project.path())
        .env("PATH", format!("{}:/usr/bin:/bin", f.bin.path().display()))
        .env("HOME", f.home.path())
        .env("USERPROFILE", f.home.path())
        .env("MOAI_TEST_HOST_MARKER", &marker)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .stdin(Stdio::null())
        .output()
        .unwrap();
    assert_eq!(out.status.code(), Some(2), "{}", stderr(&out));
    assert!(warnings(&out).is_empty(), "{}", stderr(&out));
    assert!(!marker.exists(), "no host may start");
    assert!(!String::from_utf8_lossy(&out.stdout).contains("Codex"));
    // A skill copy only under `.agents/skills` does not satisfy the Claude Code check.
    put_skill(f.home.path(), ".agents/skills/moai");
    let out = launch(&f, "cc");
    let lines = warnings(&out);
    assert_eq!(lines.len(), 1, "{}", stderr(&out));
    assert!(lines[0].contains("Claude Code"), "{}", lines[0]);
}

#[test]
fn an_installed_skill_gives_no_warning() {
    let f = fixture();
    put_skill(f.home.path(), ".claude/skills/moai");
    for host in ["cc", "glm"] {
        let out = launch(&f, host);
        assert!(warnings(&out).is_empty(), "{host}: {}", stderr(&out));
    }
}

#[test]
fn a_skill_folder_without_skill_md_still_warns() {
    let f = fixture();
    fs::create_dir_all(f.home.path().join(".claude/skills/moai")).unwrap();
    let out = launch(&f, "cc");
    assert_eq!(warnings(&out).len(), 1, "{}", stderr(&out));
}
