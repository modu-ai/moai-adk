//! #2237: expose the existing project config writer without using real host state.
mod common;
use common::CleanEnv;

use serde_json::{json, Value};
use std::{
    fs,
    path::Path,
    process::{Command, Stdio},
};

fn run(
    project: &Path,
    home: &Path,
    codex_home: Option<&Path>,
    apply: Option<(&str, &str)>,
    input: &[u8],
) -> (i32, Value) {
    run_vars(project, &[("HOME", home)], codex_home, apply, input)
}

/// Like `run`, but the caller chooses which home variables (`HOME`, `USERPROFILE`) exist.
fn run_vars(
    project: &Path,
    homes: &[(&str, &Path)],
    codex_home: Option<&Path>,
    apply: Option<(&str, &str)>,
    input: &[u8],
) -> (i32, Value) {
    let mut cmd = Command::new(crate::common::MOAI_BIN);
    cmd.clean_env();
    for (name, path) in homes {
        cmd.env(name, path);
    }
    if let Some(path) = std::env::var_os("PATH") {
        cmd.env("PATH", path);
    }
    if let Some(path) = codex_home {
        cmd.env("CODEX_HOME", path);
    }
    cmd.arg(if apply.is_some() {
        "codex-config-apply"
    } else {
        "codex-config-plan"
    })
    .arg(project);
    if let Some((digest, permission)) = apply {
        cmd.args([
            "--plan-digest",
            digest,
            "--permission-reference",
            permission,
        ]);
    }
    let mut child = cmd
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    common::feed_stdin(&mut child, input);
    let out = child.wait_with_output().unwrap();
    (
        out.status.code().unwrap(),
        serde_json::from_slice(&out.stdout)
            .unwrap_or_else(|_| panic!("{}", String::from_utf8_lossy(&out.stderr))),
    )
}

fn plan(project: &Path, home: &Path, want: &Value) -> Value {
    let (code, envelope) = run(project, home, None, None, want.to_string().as_bytes());
    assert_eq!(code, 0, "{envelope}");
    assert_eq!(envelope["operation"], "codex.config_plan");
    assert_eq!(envelope["ok"], true);
    envelope["result"].clone()
}

fn refused(code: i32, envelope: &Value) {
    assert_eq!(code, 1, "{envelope}");
    assert_eq!(envelope["ok"], false, "{envelope}");
    assert_eq!(envelope["error"]["kind"], "invalid_input", "{envelope}");
}

#[test]
fn project_config_round_trip_keeps_comments_and_unmanaged_keys() {
    let (home, project) = (tempfile::tempdir().unwrap(), tempfile::tempdir().unwrap());
    fs::create_dir(project.path().join(".codex")).unwrap();
    let original = "# User settings\nmodel = \"user-choice\" # keep this\n";
    fs::write(project.path().join(".codex/config.toml"), original).unwrap();
    let want =
        json!({"project_doc_max_bytes": 32768, "project_doc_fallback_filenames": ["TEAM.md"]});
    let p = plan(project.path(), home.path(), &want);
    assert_eq!(p["writes"], false);
    assert_eq!(p["action"], "update");
    assert_eq!(
        fs::read_to_string(project.path().join(".codex/config.toml")).unwrap(),
        original
    );
    assert!(!project.path().join(".moai-cli").exists());
    let (code, e) = run(
        project.path(),
        home.path(),
        None,
        Some((p["plan_digest"].as_str().unwrap(), "fixture-2237")),
        want.to_string().as_bytes(),
    );
    assert_eq!(code, 0, "{e}");
    assert_eq!(e["operation"], "codex.config_apply");
    assert_eq!(e["result"]["state"], "applied");
    assert_eq!(e["result"]["trust"]["writes"], false);
    assert_eq!(e["result"]["host_loaded"], "not_run");
    let bytes = fs::read(project.path().join(".codex/config.toml")).unwrap();
    assert!(String::from_utf8_lossy(&bytes).contains(original));
    assert!(project
        .path()
        .join(".moai-cli/manifests/codex-config.json")
        .is_file());
    let again = plan(project.path(), home.path(), &want);
    assert_eq!(again["action"], "unchanged");
    let (code, e) = run(
        project.path(),
        home.path(),
        None,
        Some((again["plan_digest"].as_str().unwrap(), "fixture-2237")),
        want.to_string().as_bytes(),
    );
    assert_eq!(code, 0, "{e}");
    assert_eq!(e["result"]["state"], "unchanged");
    assert_eq!(
        fs::read(project.path().join(".codex/config.toml")).unwrap(),
        bytes
    );
    assert!(!home.path().join(".codex").exists());
}

#[test]
fn stale_digest_permission_and_unmanaged_values_never_write() {
    let (home, project) = (tempfile::tempdir().unwrap(), tempfile::tempdir().unwrap());
    for want in [
        json!({"model":"x"}),
        json!({"projects.x.trust_level":"trusted"}),
        json!({"approval_policy":"never"}),
        json!({"mcp_servers.x.command":"x"}),
        json!({"project_doc_max_bytes":0}),
        json!({"project_doc_max_bytes":1048577}),
    ] {
        let (code, e) = run(
            project.path(),
            home.path(),
            None,
            None,
            want.to_string().as_bytes(),
        );
        refused(code, &e);
    }
    let want = json!({"project_doc_max_bytes":1048576});
    let p = plan(project.path(), home.path(), &want);
    for permission in ["", "null", "placeholder"] {
        let (code, e) = run(
            project.path(),
            home.path(),
            None,
            Some((p["plan_digest"].as_str().unwrap(), permission)),
            want.to_string().as_bytes(),
        );
        assert_eq!(code, 1, "{e}");
        assert_eq!(e["error"]["kind"], "permission_reference_missing", "{e}");
    }
    let (code, e) = run(
        project.path(),
        home.path(),
        None,
        Some(("sha256:wrong", "fixture-2237")),
        want.to_string().as_bytes(),
    );
    assert_eq!(code, 1, "{e}");
    assert_eq!(e["error"]["kind"], "stale_plan", "{e}");
    assert!(!project.path().join(".codex").exists());
    assert!(!project.path().join(".moai-cli").exists());
}

#[test]
fn user_edit_conflicts_without_overwrite() {
    let (home, project) = (tempfile::tempdir().unwrap(), tempfile::tempdir().unwrap());
    let first = json!({"project_doc_max_bytes":32768});
    let p = plan(project.path(), home.path(), &first);
    let (code, e) = run(
        project.path(),
        home.path(),
        None,
        Some((p["plan_digest"].as_str().unwrap(), "fixture-2237")),
        first.to_string().as_bytes(),
    );
    assert_eq!(code, 0, "{e}");
    let user = b"project_doc_max_bytes = 65536 # user edit\n";
    fs::write(project.path().join(".codex/config.toml"), user).unwrap();
    let want = json!({"project_doc_max_bytes":131072});
    let p = plan(project.path(), home.path(), &want);
    assert_eq!(p["action"], "conflict");
    let (code, e) = run(
        project.path(),
        home.path(),
        None,
        Some((p["plan_digest"].as_str().unwrap(), "fixture-2237")),
        want.to_string().as_bytes(),
    );
    assert_eq!(code, 1, "{e}");
    assert_eq!(e["error"]["kind"], "conflict");
    assert_eq!(
        fs::read(project.path().join(".codex/config.toml")).unwrap(),
        user
    );
}

#[test]
fn home_and_codex_home_targets_are_refused() {
    let home = tempfile::tempdir().unwrap();
    let input = br#"{"project_doc_max_bytes":32768}"#;
    let (code, e) = run(home.path(), home.path(), None, None, input);
    refused(code, &e);
    let project = tempfile::tempdir().unwrap();
    let codex_home = project.path().join(".codex");
    fs::create_dir(&codex_home).unwrap();
    let (code, e) = run(project.path(), home.path(), Some(&codex_home), None, input);
    refused(code, &e);
    let (code, e) = run(
        project.path(),
        home.path(),
        Some(&codex_home),
        Some(("wrong", "fixture-2237")),
        input,
    );
    refused(code, &e);
    assert!(!codex_home.join("config.toml").exists());
    assert!(!project.path().join(".moai-cli").exists());
}

#[test]
fn request_input_has_a_literal_64_kib_boundary() {
    let (home, project) = (tempfile::tempdir().unwrap(), tempfile::tempdir().unwrap());
    let mut input = b"{}".to_vec();
    input.resize(65536, b' ');
    let (code, e) = run(project.path(), home.path(), None, None, &input);
    assert_eq!(code, 0, "{e}");
    input.push(b' ');
    let (code, e) = run(project.path(), home.path(), None, None, &input);
    assert_eq!(code, 1, "{e}");
    assert_eq!(e["error"]["kind"], "input_too_large", "{e}");
    assert!(e["error"]["message"].as_str().unwrap().contains("64 KiB"));
    for raw in [b"[1]".as_slice(), b"{", b"null"] {
        let (code, e) = run(project.path(), home.path(), None, None, raw);
        refused(code, &e);
    }
    assert!(!project.path().join(".moai-cli").exists());
}

#[cfg(unix)]
#[test]
fn symlink_config_directory_never_writes_the_target() {
    let (home, project, outside) = (
        tempfile::tempdir().unwrap(),
        tempfile::tempdir().unwrap(),
        tempfile::tempdir().unwrap(),
    );
    std::os::unix::fs::symlink(outside.path(), project.path().join(".codex")).unwrap();
    let (code, e) = run(
        project.path(),
        home.path(),
        None,
        None,
        br#"{"project_doc_max_bytes":32768}"#,
    );
    refused(code, &e);
    assert_eq!(fs::read_dir(outside.path()).unwrap().count(), 0);
}

#[test]
fn unrelated_codex_home_is_preserved_and_unavailable_home_is_refused() {
    let (home, project, user) = (
        tempfile::tempdir().unwrap(),
        tempfile::tempdir().unwrap(),
        tempfile::tempdir().unwrap(),
    );
    let before = b"# Synthetic user config\nmodel = \"user-only\"\n";
    fs::write(user.path().join("config.toml"), before).unwrap();
    let input = br#"{"project_doc_max_bytes":32768}"#;
    let (code, e) = run(project.path(), home.path(), Some(user.path()), None, input);
    assert_eq!(code, 0, "{e}");
    let (code, e) = run(
        project.path(),
        home.path(),
        Some(user.path()),
        Some((e["result"]["plan_digest"].as_str().unwrap(), "fixture-2237")),
        input,
    );
    assert_eq!(code, 0, "{e}");
    assert_eq!(fs::read(user.path().join("config.toml")).unwrap(), before);
    let missing = home.path().join("unavailable");
    let (code, e) = run(project.path(), home.path(), Some(&missing), None, input);
    assert_eq!(code, 1, "{e}");
    assert_eq!(e["error"]["kind"], "io", "{e}");
    assert!(!missing.exists());
}

/// Pre-audit FAIL 6812471540 B1: the library's "project is home" check needs a resolved
/// home, so an unresolvable or missing home must refuse plan and apply before any write.
#[test]
fn unresolvable_or_missing_home_is_refused_before_any_write() {
    let (scratch, project) = (tempfile::tempdir().unwrap(), tempfile::tempdir().unwrap());
    let missing = scratch.path().join("no-such-home");
    let input = br#"{"project_doc_max_bytes":32768}"#;
    let unresolvable: [&[(&str, &Path)]; 2] = [
        &[("HOME", missing.as_path())],
        &[("USERPROFILE", missing.as_path())],
    ];
    for homes in unresolvable {
        for apply in [None, Some(("any-digest", "fixture-2237"))] {
            let (code, e) = run_vars(project.path(), homes, None, apply, input);
            assert_eq!(code, 1, "{homes:?} {e}");
            assert_eq!(e["ok"], false, "{e}");
            assert_eq!(e["error"]["kind"], "io", "{e}");
            assert_eq!(
                e["error"]["message"], "home directory unavailable; config target refused",
                "{e}"
            );
        }
    }
    for apply in [None, Some(("any-digest", "fixture-2237"))] {
        let (code, e) = run_vars(project.path(), &[], None, apply, input);
        refused(code, &e);
        assert_eq!(
            e["error"]["message"], "home directory missing; config target refused",
            "{e}"
        );
    }
    assert!(!missing.exists());
    assert_eq!(fs::read_dir(project.path()).unwrap().count(), 0);
}

#[test]
fn apply_has_the_same_literal_64_kib_input_boundary() {
    let (home, project) = (tempfile::tempdir().unwrap(), tempfile::tempdir().unwrap());
    let p = plan(project.path(), home.path(), &json!({}));
    let mut input = b"{}".to_vec();
    input.resize(65536, b' ');
    let (code, e) = run(
        project.path(),
        home.path(),
        None,
        Some((p["plan_digest"].as_str().unwrap(), "fixture-2237")),
        &input,
    );
    assert_eq!(code, 0, "{e}");
    let receipt_path = project.path().join(".moai-cli/manifests/codex-config.json");
    let before = fs::read(&receipt_path).unwrap();
    input.push(b' ');
    let (code, e) = run(
        project.path(),
        home.path(),
        None,
        Some((p["plan_digest"].as_str().unwrap(), "fixture-2237")),
        &input,
    );
    assert_eq!(code, 1, "{e}");
    assert_eq!(e["error"]["kind"], "input_too_large", "{e}");
    assert_eq!(fs::read(receipt_path).unwrap(), before);
    assert!(!project.path().join(".codex/config.toml").exists());
}
