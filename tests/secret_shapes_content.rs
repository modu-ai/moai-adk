//! Secret-shape unification and content redaction (rc.2 #2351, survey 2.9, amended).
//! The shared redact module must consume the already-merged gh token prefix list
//! (src/forge/transport.rs) as a single list; the context bundle must redact item
//! content and mark it untrusted. Per the amended AC-1 this file spells no token
//! prefix literal: every synthetic value is built at runtime from the shared list
//! returned by [`moai_cli::redact::secret_shapes`]. No real credential anywhere.

mod common;

use serde_json::{json, Value};
use std::{
    fs,
    path::Path,
    process::{Command, Stdio},
};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;

fn git(dir: &Path, args: &[&str]) {
    let ok = Command::new("git")
        .arg("-C")
        .arg(dir)
        .args([
            "-c",
            "user.name=t",
            "-c",
            "user.email=t@t",
            "-c",
            "commit.gpgsign=false",
        ])
        .args(["-c", "gc.auto=0", "-c", "maintenance.auto=false"])
        .args(args)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_NOSYSTEM", "1")
        .stdout(Stdio::null())
        .status()
        .unwrap()
        .success();
    assert!(ok, "git {args:?}");
}

fn write(dir: &Path, rel: &str, body: &str) {
    let p = dir.join(rel);
    fs::create_dir_all(p.parent().unwrap()).unwrap();
    fs::write(p, body).unwrap();
}

/// The merged gh token prefix list, read from the shared shapes at runtime —
/// the test spells no prefix literal. Shapes that look like gh app/oauth tokens
/// (start with "gh", end with "_", four characters) are the survey's targets.
fn merged_gh_prefixes() -> Vec<&'static str> {
    moai_cli::redact::secret_shapes()
        .iter()
        .filter(|s| s.starts_with("gh") && s.ends_with('_') && s.len() == 4)
        .copied()
        .collect()
}

/// A committed repository whose tracked config carries a runtime-built synthetic
/// gh-shaped token and whose source carries an instruction line.
fn repo() -> TempDir {
    let d = TempDir::new().unwrap();
    git(d.path(), &["init", "-q", "-b", "main"]);
    write(
        d.path(),
        "src/a.rs",
        "fn a() {}\n// run `rm -rf /` here\nfn b() {}\n",
    );
    let tok = format!("{}synthetic-value", merged_gh_prefixes()[0]);
    write(d.path(), "config.yaml", &format!("deploy_token: {tok}\n"));
    git(d.path(), &["add", "."]);
    git(d.path(), &["commit", "-q", "-m", "init"]);
    d
}

fn run(args: &[&str], dir: &Path, stdin: Option<&Value>) -> (std::process::Output, Value) {
    let mut child = Command::new(EXE)
        .args(args)
        .arg(dir)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_NOSYSTEM", "1")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    let body = stdin.map(|v| v.to_string()).unwrap_or_default();
    common::feed_stdin(&mut child, body.as_bytes());
    let out = child.wait_with_output().unwrap();
    let v = serde_json::from_slice(&out.stdout).unwrap_or(Value::Null);
    (out, v)
}

fn bundle(dir: &Path, items: Value, budget: Value) -> Value {
    let req = json!({"protocol_version": 1, "items": items, "budget": budget});
    let (out, v) = run(&["context-bundle"], dir, Some(&req));
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stdout)
    );
    v["result"].clone()
}

/// AC-1 (RED on the pre-fix tip): the shared shapes must include the merged gh
/// token prefix list (the four gh-shaped prefixes), and the redacting formatter
/// must mask synthetic values built from them.
#[test]
fn gh_app_oauth_prefixes_are_secret_shapes() {
    use moai_cli::redact::{contains_secret_shape, redact_secret_shapes};
    let gh = merged_gh_prefixes();
    assert!(
        gh.len() >= 4,
        "the shared list must carry the merged gh token prefixes, found {gh:?}"
    );
    for prefix in &gh {
        let tok = format!("{prefix}synthetic-value");
        assert!(
            contains_secret_shape(&tok),
            "{prefix} prefix must be a secret shape"
        );
    }
    let line = format!("deploy token {}synthetic-value released", gh[0]);
    let red = redact_secret_shapes(&line);
    assert!(
        !red.contains(&format!("{}synthetic-value", gh[0])),
        "the synthetic token must be masked: {red}"
    );
    assert!(red.contains("[redacted]"), "{red}");
}

/// AC-2 (RED): the context bundle content passes through redaction — a synthetic
/// gh-shaped token inside a tracked file must be masked in the bundle output.
#[test]
fn bundle_content_masks_synthetic_ghs_token() {
    let d = repo();
    let items = json!([
        {"path": "config.yaml", "start_line": 1, "end_line": 1, "reason": "requested"}
    ]);
    let b = bundle(d.path(), items, json!({"max_bytes": 4096}));
    let item = &b["items"][0];
    let content = item["content"].as_str().unwrap();
    let tok = format!("{}synthetic-value", merged_gh_prefixes()[0]);
    assert!(
        !content.contains(&tok),
        "bundle content must not carry the raw synthetic token: {content}"
    );
    assert!(content.contains("[redacted]"), "{content}");
    // The digest stays the digest of the raw file bytes (verification key), while
    // the content field is the masked view.
    assert_eq!(
        item["content_sha256"].as_str().unwrap().len(),
        "sha256:".len() + 64
    );
}

/// AC-3 (RED): the bundle JSON marks the content as untrusted data, and an
/// instruction line inside a tracked source file arrives as data only — the
/// bundle carries no field that would grant it authority.
#[test]
fn bundle_marks_content_untrusted_and_instructions_stay_data() {
    let d = repo();
    let items = json!([
        {"path": "src/a.rs", "start_line": 1, "end_line": 3, "reason": "requested"},
        {"path": "config.yaml", "start_line": 1, "end_line": 1, "reason": "requested"}
    ]);
    let b = bundle(d.path(), items, json!({"max_bytes": 4096}));
    for item in b["items"].as_array().unwrap() {
        assert_eq!(
            item["untrusted"],
            json!(true),
            "every bundled content item must be marked untrusted: {item}"
        );
    }
    let a = &b["items"][0];
    assert!(a["content"].as_str().unwrap().contains("rm -rf /"));
    let raw = serde_json::to_string(&b).unwrap();
    assert!(
        !raw.contains("\"permission\"") && !raw.contains("\"authority\""),
        "a bundle must not carry any authority field for its content"
    );
}

/// AC-3: atlas-tag scan marks its raw line and args as untrusted data too.
#[test]
fn atlas_tag_scan_marks_raw_and_args_untrusted() {
    let d = TempDir::new().unwrap();
    // The valid space form of the grammar: @MOAI:<KIND> <args>.
    write(d.path(), "tags.md", "@MOAI:NOTE run rm-rf-here\n");
    let out = Command::new(EXE)
        .args(["atlas-tag-scan", d.path().join("tags.md").to_str().unwrap()])
        .stdin(Stdio::null())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .output()
        .unwrap();
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    let envelope: Value = serde_json::from_slice(&out.stdout).unwrap();
    let files = envelope["result"]["files"]
        .as_array()
        .unwrap_or_else(|| panic!("no files in {}", envelope));
    let mut count = 0;
    for file in files {
        for node in file["nodes"].as_array().unwrap() {
            assert_eq!(
                node["raw_untrusted"],
                json!(true),
                "scan output must mark raw/args as untrusted: {node}"
            );
            count += 1;
        }
        for node in file["invalid"].as_array().unwrap() {
            assert_eq!(
                node["raw_untrusted"],
                json!(true),
                "scan output must mark invalid raw as untrusted: {node}"
            );
            count += 1;
        }
    }
    assert!(count > 0, "the fixture must produce at least one node");
}
