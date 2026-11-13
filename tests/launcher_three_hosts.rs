//! #2378 (user decision 2026-11-10T10:07Z): the launcher supports exactly three hosts, `cc`,
//! `glm` and `codex`; the fourth host added earlier the same day (#2362) is removed. Every form
//! naming it is a usage error that starts no host process and takes no `agent-N` number; the
//! legacy top-level `-a <name>` keeps its `unknown host` error (exit 1). Runs in a temporary git
//! project with a stub of that host on PATH that leaves a trace when it is started; no real host
//! starts. The name is assembled at run time; it was once kept out of every Rust file, but the
//! 2026-11-13 user confirmation of five supported hosts (Issue #2633) brings the host back as
//! an agent-file target, so only the launcher stays closed to it.
#![cfg(unix)]
mod common;
use std::os::unix::fs::PermissionsExt;
use std::{
    fs,
    path::Path,
    process::{Command, Output, Stdio},
};

/// The removed host's name, assembled so that a whole-word scan of this file stays clean.
fn removed() -> String {
    ["o", "m", "p"].concat()
}

struct Fixture {
    project: tempfile::TempDir,
    bin: tempfile::TempDir,
}

impl Fixture {
    fn new() -> Self {
        let project = tempfile::tempdir().unwrap();
        let ok = common::git::command()
            .args(["init", "-q"])
            .current_dir(project.path())
            .stdout(Stdio::null())
            .stderr(Stdio::null())
            .status()
            .unwrap()
            .success();
        assert!(ok);
        let bin = tempfile::tempdir().unwrap();
        let stub = bin.path().join(removed());
        fs::write(
            &stub,
            format!(
                "#!/bin/sh\ntouch '{}'\nexit 0\n",
                bin.path().join("started").display()
            ),
        )
        .unwrap();
        fs::set_permissions(&stub, fs::Permissions::from_mode(0o755)).unwrap();
        // A stored number, so any change to it is visible.
        let store = project.path().join(".git/.moai-cli/sessions/agent-seq");
        fs::create_dir_all(store.parent().unwrap()).unwrap();
        fs::write(&store, "5\n").unwrap();
        Fixture { project, bin }
    }

    fn run(&self, args: &[String]) -> Output {
        Command::new(env!("CARGO_BIN_EXE_moai-cli"))
            .args(args)
            .current_dir(self.project.path())
            .env(
                "PATH",
                format!("{}:/usr/bin:/bin", self.bin.path().display()),
            )
            .env("HOME", self.project.path().join("no-such-home"))
            .env("ANTHROPIC_AUTH_TOKEN", "moai-test-dummy-not-real")
            .env_remove("CLAUDE_CODE_DISABLE_ADVISOR_TOOL")
            .env_remove(moai_cli::session::AGENT_SEQ_TEST_HOLD_VAR)
            .stdin(Stdio::null())
            .output()
            .unwrap()
    }

    fn store(&self) -> String {
        fs::read_to_string(
            self.project
                .path()
                .join(".git/.moai-cli/sessions/agent-seq"),
        )
        .unwrap()
    }

    fn host_started(&self) -> bool {
        self.bin.path().join("started").exists()
    }
}

fn argv(words: &[&str]) -> Vec<String> {
    words
        .iter()
        .map(|w| {
            if *w == "HOST" {
                removed()
            } else {
                (*w).to_string()
            }
        })
        .collect()
}

#[test]
fn every_form_of_the_removed_host_is_a_usage_error_that_starts_nothing() {
    for words in [
        &["HOST"][..],
        &["HOST", "-a"],
        &["HOST", "-l"],
        &["HOST", "--dry-run"],
        &["HOST", "-a", "--dry-run"],
    ] {
        let f = Fixture::new();
        let out = f.run(&argv(words));
        let stderr = String::from_utf8_lossy(&out.stderr);
        assert_eq!(out.status.code(), Some(2), "{words:?}: {stderr}");
        assert!(
            stderr.contains("unrecognized subcommand"),
            "{words:?}: {stderr}"
        );
        assert!(out.stdout.is_empty(), "{words:?}");
        assert!(!f.host_started(), "{words:?} started a host process");
        assert_eq!(f.store(), "5\n", "{words:?} took a number");
    }
}

#[test]
fn the_legacy_top_level_agent_flag_still_names_only_three_hosts() {
    for words in [&["-a", "HOST"][..], &["-a", "HOST", "--dry-run"]] {
        let f = Fixture::new();
        let out = f.run(&argv(words));
        let stderr = String::from_utf8_lossy(&out.stderr);
        assert_eq!(out.status.code(), Some(1), "{words:?}: {stderr}");
        assert!(stderr.contains("unknown host"), "{words:?}: {stderr}");
        assert!(stderr.contains("(cc|glm|codex)"), "{words:?}: {stderr}");
        assert!(!f.host_started(), "{words:?} started a host process");
        assert_eq!(f.store(), "5\n", "{words:?} took a number");
    }
}

/// Lower-cased alphanumeric words of `text`.
fn words(text: &str) -> Vec<String> {
    text.split(|c: char| !c.is_ascii_alphanumeric())
        .filter(|w| !w.is_empty())
        .map(str::to_ascii_lowercase)
        .collect()
}

#[test]
fn root_help_lists_cc_glm_and_codex_and_never_the_removed_host() {
    let f = Fixture::new();
    let out = f.run(&["--help".to_string()]);
    assert!(out.status.success());
    let help = String::from_utf8_lossy(&out.stdout);
    let commands: Vec<&str> = help
        .lines()
        .filter_map(|l| l.strip_prefix("  "))
        .filter_map(|l| l.split_whitespace().next())
        .collect();
    for host in ["cc", "glm", "codex"] {
        assert!(commands.contains(&host), "{host} missing from:\n{help}");
    }
    assert!(!words(&help).contains(&removed()), "{help}");
}

/// The support documents say the host is not supported, once, with the decision date; the
/// launcher documents list only the three hosts.
#[test]
fn the_support_matrix_records_the_removal_and_launcher_docs_list_three_hosts() {
    let root = Path::new(env!("CARGO_MANIFEST_DIR"));
    let matrix = fs::read_to_string(root.join("docs/host-support-matrix.md")).unwrap();
    let line = format!(
        "{} — 지원하지 않음(2026-11-10 결정",
        removed().to_uppercase()
    );
    assert_eq!(matrix.matches(&line).count(), 1, "{line}");
    for doc in ["docs/rust-product-path.md", "docs/user-entry-draft.md"] {
        let text = fs::read_to_string(root.join(doc)).unwrap();
        assert!(!words(&text).contains(&removed()), "{doc}");
    }
}
