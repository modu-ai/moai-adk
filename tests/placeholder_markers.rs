//! SPEC-20261121-placeholder-markers: a first launch marks every value it invents. The profile a
//! first `cc|glm|codex|omp` launch writes holds one `source: "default"` decision for each of
//! purpose, tools, development methods and hosts, so the project interview can tell a launcher
//! placeholder from an answer. Every launch runs in a temporary folder with a temporary HOME and
//! a stub host on PATH; no real host starts.
#![cfg(unix)]
mod common;
use serde_json::{json, Value};
use std::os::unix::fs::PermissionsExt;
use std::{fs, path::Path, process::Command};

const EXE: &str = crate::common::MOAI_BIN;
/// The four values a first launch writes that are not an answer.
const MARKED: [&str; 4] = ["purpose", "tools", "development_methods", "hosts"];
/// The decision id of each marked topic; the three earlier ids keep their names.
const MARKER_IDS: [(&str, &str); 4] = [
    ("purpose", "auto-init-purpose"),
    ("tools", "auto-init-tools"),
    ("development_methods", "auto-init-methods"),
    ("hosts", "auto-init-hosts"),
];

/// The profile a first launch of `host` writes in a fresh folder, and the host it names.
fn first_launch_profile(host: &str) -> Value {
    let bin = tempfile::tempdir().unwrap();
    for stub in ["claude", "codex", "omp"] {
        let path = bin.path().join(stub);
        fs::write(&path, "#!/bin/sh\nexit 0\n").unwrap();
        fs::set_permissions(&path, fs::Permissions::from_mode(0o755)).unwrap();
    }
    let home = tempfile::tempdir().unwrap();
    let plain = tempfile::tempdir().unwrap();
    let dir = plain.path().join("My Project");
    fs::create_dir(&dir).unwrap();
    let out = Command::new(EXE)
        .args([host, "-a"])
        .current_dir(&dir)
        .env("PATH", format!("{}:/usr/bin:/bin", bin.path().display()))
        .env("HOME", home.path())
        .env("USERPROFILE", home.path())
        .env("ANTHROPIC_AUTH_TOKEN", "moai-test-dummy-not-real")
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .stdin(std::process::Stdio::null())
        .output()
        .unwrap();
    assert!(
        out.status.success(),
        "{host}: {}{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    read_profile(&dir)
}

fn read_profile(dir: &Path) -> Value {
    serde_json::from_slice(&fs::read(dir.join(".moai/project.yaml")).unwrap()).unwrap()
}

/// AC-1 (REQ-1, REQ-2): exactly one `source: default` decision per marked value, with the id
/// `auto-init-<topic>`; the hosts decision names the launching host.
#[test]
fn a_first_launch_marks_purpose_tools_methods_and_hosts_as_placeholders() {
    for (host, profile_host) in [
        ("cc", "claude"),
        ("glm", "claude"),
        ("codex", "codex"),
        ("omp", "omp"),
    ] {
        let p = first_launch_profile(host);
        let decisions = p["decisions"].as_array().unwrap();
        let mut topics: Vec<&str> = decisions
            .iter()
            .filter(|d| d["source"] == json!("default"))
            .map(|d| d["topic"].as_str().unwrap())
            .collect();
        topics.sort_unstable();
        let mut want = MARKED.to_vec();
        want.sort_unstable();
        assert_eq!(topics, want, "{host}: {decisions:?}");
        assert_eq!(
            decisions.len(),
            MARKED.len(),
            "{host}: no other decision is written"
        );
        for d in decisions {
            let topic = d["topic"].as_str().unwrap();
            let id = MARKER_IDS.iter().find(|(t, _)| *t == topic).unwrap().1;
            assert_eq!(d["id"], json!(id), "{host}: {topic}");
            assert_eq!(d["scope"], json!("project"), "{host}");
        }
        let hosts = decisions
            .iter()
            .find(|d| d["topic"] == json!("hosts"))
            .unwrap();
        assert_eq!(hosts["value"], json!(profile_host), "{host}");
        assert_eq!(p["hosts"], json!([profile_host]), "{host}");
    }
}

/// AC-2 (REQ-1): the marked profile is still the placeholder profile, so the rules that wait
/// for the interview (no managed AGENTS.md section) still hold.
#[test]
fn the_marked_profile_is_still_the_placeholder_profile() {
    let p = first_launch_profile("cc");
    let profile: moai_cli::profile::Profile = serde_json::from_value(p).unwrap();
    assert!(profile.is_placeholder());
}
