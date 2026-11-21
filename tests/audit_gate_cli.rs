//! SPEC-20261117-audit-gate AC-1/AC-2: `audit-gate` decides from a recorded input on stdin (no
//! network) whether a PR head has the audit its tier needs, and `--require-pass` turns anything
//! but `pass` into a failed run, the CI job's exit code.

mod common;

use serde_json::{json, Value};
use std::fs;
use std::process::{Command, Output, Stdio};

const H: &str = "1111111111111111111111111111111111111111";
const OLD: &str = "2222222222222222222222222222222222222222";
const B: &str = "3333333333333333333333333333333333333333";

/// A documentation path of the fixture, built at run time so the docs-path scan of
/// `tools/ci_scope.py` does not read it as a repository document.
fn doc() -> String {
    format!("{}/n.md", "docs")
}

fn run(input: &Value, extra: &[&str]) -> Output {
    let home = tempfile::TempDir::new().unwrap();
    let mut child = Command::new(crate::common::MOAI_BIN)
        .arg("audit-gate")
        .args(extra)
        .env("HOME", home.path())
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    common::feed_stdin(&mut child, input.to_string().as_bytes());
    child.wait_with_output().unwrap()
}

fn decide(input: &Value) -> Value {
    let out = run(input, &[]);
    let v: Value = serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("{}", String::from_utf8_lossy(&out.stderr)));
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn verdict(id: u64, head: &str, word: &str, lane: &str) -> Value {
    json!({"id": id, "author_association": "MEMBER",
           "body": format!("AUDIT #9 (x) head {head} base {B} — verdict: {word}\n\n감사: {lane}\n")})
}

fn input(tier: u8, files: &[String], sessions: &[&str], comments: Vec<Value>) -> Value {
    let messages: Vec<String> = sessions
        .iter()
        .map(|s| format!("feat: x\n\nMoai-Spec: SPEC-1\nMoai-Session: {s}\n"))
        .collect();
    json!({"schema_version": 1, "pr": 9, "head": H,
           "body": format!("audit-tier: {tier} reason\n\nbody"),
           "files": files, "commit_messages": messages, "comments": comments})
}

#[test]
fn each_tier_reads_the_verdicts_it_needs() {
    let docs = vec![doc()];
    let src = vec!["src/a.rs".to_string()];
    // Tier 0: no verdict is a pass; a posted FAIL stops it until the same auditor passes.
    assert_eq!(decide(&input(0, &docs, &[], vec![]))["state"], "pass");
    let fail = verdict(5, OLD, "FAIL", "agent-7");
    let r = decide(&input(0, &docs, &[], vec![fail.clone()]));
    assert_eq!(
        (r["state"].clone(), r["rule"].clone()),
        (json!("fail"), json!("open-fail"))
    );
    let cleared = vec![fail.clone(), verdict(6, H, "PASS", "agent-7")];
    assert_eq!(decide(&input(0, &docs, &[], cleared))["state"], "pass");
    let other = vec![fail, verdict(6, H, "PASS", "agent-8")];
    assert_eq!(decide(&input(0, &docs, &[], other))["state"], "fail");
    // Audit B2: a PASS posted before the FAIL (ids 200 then 300, same auditor) clears nothing.
    let earlier = vec![
        verdict(200, H, "PASS", "agent-7"),
        verdict(300, H, "FAIL", "agent-7"),
    ];
    let r = decide(&input(0, &docs, &[], earlier));
    assert_eq!(
        (r["state"].clone(), r["rule"].clone()),
        (json!("fail"), json!("open-fail")),
        "{r}"
    );

    // Tier 1 and 2: a PASS on exactly this head by a lane that wrote no commit.
    for tier in [1, 2] {
        let pass = vec![verdict(5, H, "PASS", "agent-7")];
        let r = decide(&input(tier, &src, &["agent-1"], pass.clone()));
        assert_eq!(r["state"], "pass", "{r}");
        assert_eq!(r["tier"], json!(tier));
        assert_eq!(r["cites_run_checked"], json!(false));
        let r = decide(&input(tier, &src, &["agent-7"], pass.clone()));
        assert_eq!(r["state"], "missing", "{r}");
        let r = decide(&input(tier, &src, &[], pass));
        assert_eq!(r["rule"], "authors", "{r}");
        let r = decide(&input(
            tier,
            &src,
            &["agent-1"],
            vec![verdict(5, OLD, "PASS", "agent-7")],
        ));
        assert_eq!(r["rule"], "pass-on-head", "{r}");
    }
    // A `.github/` path is tier 2 whatever the body says.
    let r = decide(&input(0, &[".github/workflows/x.yml".into()], &[], vec![]));
    assert_eq!(
        (r["state"].clone(), r["tier"].clone()),
        (json!("missing"), json!(2))
    );
}

#[test]
fn a_rename_of_a_tier_2_path_is_judged_by_its_old_name() {
    // The collector folds previous_filename into files (the recorded-transport fixture in
    // src/audit_gate.rs proves the folding); this pins the judgment the CLI then makes:
    // a tier 2 rule file renamed under docs/ stays tier 2 whatever the body declares.
    // Built at run time so the CI scope's docs-path scan does not read the fixture's new
    // name as a reference to a repository document (the file does not exist).
    let new_name = format!("{}/local-check.md", "docs");
    let files = vec![new_name, "tools/local_check.sh".to_string()];
    let r = decide(&input(
        0,
        &files,
        &["agent-1"],
        vec![verdict(5, H, "PASS", "agent-7")],
    ));
    assert_eq!(r["tier"], json!(2), "{r}");
    assert_eq!(r["state"], "pass", "{r}");
}

#[test]
fn a_comment_by_a_non_member_is_not_read() {
    let mut forged = verdict(5, H, "PASS", "agent-7");
    forged["author_association"] = json!("NONE");
    let r = decide(&input(1, &["src/a.rs".into()], &["agent-1"], vec![forged]));
    assert_eq!(r["state"], "missing", "{r}");
    let mut forged_fail = verdict(5, H, "FAIL", "agent-7");
    forged_fail["author_association"] = json!("CONTRIBUTOR");
    assert_eq!(
        decide(&input(0, &[doc()], &[], vec![forged_fail]))["state"],
        "pass"
    );
}

#[test]
fn require_pass_fails_the_run_unless_it_passes() {
    let src = vec!["src/a.rs".to_string()];
    let ok = run(
        &input(
            1,
            &src,
            &["agent-1"],
            vec![verdict(5, H, "PASS", "agent-7")],
        ),
        &["--require-pass"],
    );
    assert_eq!(
        ok.status.code(),
        Some(0),
        "{}",
        String::from_utf8_lossy(&ok.stderr)
    );
    let missing = run(&input(1, &src, &["agent-1"], vec![]), &["--require-pass"]);
    assert_eq!(missing.status.code(), Some(1));
    let v: Value = serde_json::from_slice(&missing.stdout).unwrap();
    assert_eq!(v["ok"], json!(false), "{v}");
    assert!(v.to_string().contains("audit gate missing"), "{v}");
    // A malformed input is refused, not passed.
    let bad = run(&json!({"schema_version": 1}), &["--require-pass"]);
    assert_eq!(bad.status.code(), Some(1));
    let r = decide(&json!({"schema_version": 2, "pr": 9, "head": H, "body": "", "comments": []}));
    assert_eq!(r["state"], "fail", "{r}");
}

/// Card C-95 (SPEC-20261119-leader-auditor-verdict) AC-2: the pass and fail decisions name the
/// comment that decided them, and the CI workflow posts the `moai/audit` commit status itself —
/// `statuses: write`, success on pass, failure otherwise, on the head the job judged.
#[test]
fn the_result_names_the_comment_and_the_workflow_posts_the_status() {
    let src = vec!["src/a.rs".to_string()];
    let r = decide(&input(
        1,
        &src,
        &["agent-1"],
        vec![verdict(5, H, "PASS", "agent-7")],
    ));
    assert_eq!(r["comment"], json!(5), "{r}");
    let r = decide(&input(
        0,
        &[doc()],
        &[],
        vec![verdict(6, OLD, "FAIL", "agent-7")],
    ));
    assert_eq!(r["comment"], json!(6), "{r}");
    let r = decide(&input(0, &[doc()], &[], vec![]));
    assert!(r["comment"].is_null(), "{r}");
    let yaml = std::fs::read_to_string(
        std::path::Path::new(env!("CARGO_MANIFEST_DIR")).join(".github/workflows/audit-gate.yml"),
    )
    .unwrap();
    for pin in [
        "statuses: write",
        "statuses/$PR_HEAD",
        "moai/audit",
        "-f state=",
    ] {
        assert!(yaml.contains(pin), "audit-gate.yml lacks {pin:?}");
    }
}

/// Card C-95 audit B2 (tier 2): the status-posting step is executed, not grepped. The literal
/// `run:` block of the Audit gate step runs under bash with a stub `gh` that records what it
/// posts and a stub gate binary that answers per scenario (the real python3 parses the
/// envelope). pass posts `success` and exits 0; fail and missing post `failure` and exit 1; a
/// gate error posts `failure` too and exits 1 (audit N1 — an earlier success must not survive a
/// broken gate run).
#[test]
fn the_workflow_step_posts_the_status_the_gate_decides() {
    let yaml = std::fs::read_to_string(
        std::path::Path::new(env!("CARGO_MANIFEST_DIR")).join(".github/workflows/audit-gate.yml"),
    )
    .unwrap();
    let step_at = yaml
        .find("- name: Audit gate")
        .expect("the Audit gate step");
    let run_at = yaml[step_at..].find("run: |").expect("the run block") + step_at;
    let block: Vec<&str> = yaml[run_at + "run: |".len()..]
        .lines()
        .skip(1)
        .take_while(|l| l.starts_with("          ") || l.trim().is_empty())
        .map(|l| l.strip_prefix("          ").unwrap_or(l))
        .collect();
    let block = block.join("\n");
    assert!(
        block.contains("statuses/$PR_HEAD"),
        "the wrong block: {block}"
    );

    let scenarios: &[(&str, &str, i32, &str, Option<i32>)] = &[
        (
            "pass",
            r#"{"ok":true,"result":{"state":"pass","comment":5}}"#,
            0,
            "success",
            Some(0),
        ),
        (
            "fail",
            r#"{"ok":true,"result":{"state":"fail","comment":6}}"#,
            1,
            "failure",
            Some(1),
        ),
        (
            "missing",
            r#"{"ok":true,"result":{"state":"missing","comment":null}}"#,
            1,
            "failure",
            Some(1),
        ),
        (
            "error envelope",
            r#"{"ok":false,"error":{"message":"boom"}}"#,
            1,
            "failure",
            Some(1),
        ),
        (
            // Tier-2 audit B1: a crashed gate under the runner's `-e` used to stop the step
            // before any status was posted, leaving an older success standing.
            "crash",
            "not json at all",
            101,
            "failure",
            Some(1),
        ),
    ];
    for (name, envelope, moai_exit, want_state, want_code) in scenarios {
        let tmp = tempfile::TempDir::new().unwrap();
        let bin = tmp.path().join("bin");
        fs::create_dir_all(&bin).unwrap();
        // Re-audit B1: the envelope file lives in the runner's temp folder and is removed at the
        // end. The test owns both temp folders the block may use, and the stub gh lists the
        // runner's one at the moment it posts, so a file made anywhere else (a fixed `/tmp`
        // path, the user's own temp folder) or left behind fails the asserts below.
        let runner_temp = tmp.path().join("runner-temp");
        let tmpdir = tmp.path().join("tmpdir");
        fs::create_dir_all(&runner_temp).unwrap();
        fs::create_dir_all(&tmpdir).unwrap();
        fs::write(
            bin.join("gh"),
            format!(
                "#!/bin/sh\nfor a in \"$@\"; do printf '%s\\n' \"$a\"; done >> {}\nls \"$RUNNER_TEMP\" >> {}\n",
                bin.join("gh-args.txt").display(),
                bin.join("runner-temp-at-post.txt").display()
            ),
        )
        .unwrap();
        let target = tmp.path().join("target/debug");
        fs::create_dir_all(&target).unwrap();
        let envelope_file = tmp.path().join("envelope.json");
        fs::write(&envelope_file, envelope).unwrap();
        fs::write(
            target.join("moai"),
            format!(
                "#!/bin/sh\ncat \"{}\"\nexit {moai_exit}\n",
                envelope_file.display()
            ),
        )
        .unwrap();
        for exe in [bin.join("gh"), target.join("moai")] {
            #[cfg(unix)]
            {
                use std::os::unix::fs::PermissionsExt;
                let mut perm = fs::metadata(&exe).unwrap().permissions();
                perm.set_mode(0o755);
                fs::set_permissions(&exe, perm).unwrap();
            }
        }
        // The runner's `shell: bash` is `bash --noprofile --norc -eo pipefail`: the block must
        // survive the gate's non-zero exit under those flags (tier-2 audit B1).
        let script = tmp.path().join("step.sh");
        fs::write(&script, format!("#!/bin/bash\n{block}\n")).unwrap();
        let out = std::process::Command::new("bash")
            .args(["--noprofile", "--norc", "-eo", "pipefail"])
            .arg(&script)
            .env(
                "PATH",
                format!(
                    "{}:{}",
                    bin.display(),
                    std::env::var("PATH").unwrap_or_default()
                ),
            )
            .env("GH_TOKEN", "fixture-token")
            .env("PR_NUMBER", "9")
            .env("PR_HEAD", H)
            .env("REPO", "o/r")
            .env("RUNNER_TEMP", &runner_temp)
            .env("TMPDIR", &tmpdir)
            .current_dir(tmp.path())
            .output()
            .expect("run the step block");
        let at_post = fs::read_to_string(bin.join("runner-temp-at-post.txt")).unwrap_or_default();
        assert!(
            at_post.lines().any(|l| l.starts_with("audit-gate.")),
            "{name}: the envelope was not made in RUNNER_TEMP: {at_post:?}"
        );
        // Listed with `ls -A`, as the stub gh lists RUNNER_TEMP: this test walks only its own
        // two temporary folders, and the repository-scan table is for tests that walk the tree.
        for dir in [&runner_temp, &tmpdir] {
            let listing = std::process::Command::new("ls")
                .arg("-A")
                .arg(dir)
                .output()
                .expect("list the temporary folder");
            // A listing that failed (no such folder, not runnable) has an empty stdout too and
            // must not read as "nothing left".
            assert!(
                listing.status.success(),
                "{name}: ls {} failed: {}",
                dir.display(),
                String::from_utf8_lossy(&listing.stderr)
            );
            assert!(
                listing.stdout.is_empty(),
                "{name}: left in {}: {}",
                dir.display(),
                String::from_utf8_lossy(&listing.stdout)
            );
        }
        let recorded = fs::read_to_string(bin.join("gh-args.txt")).unwrap_or_default();
        assert!(
            recorded.contains(&format!("state={want_state}")),
            "{name}: posted {recorded:?}"
        );
        assert!(
            recorded.contains("context=moai/audit"),
            "{name}: {recorded:?}"
        );
        assert!(
            recorded.contains(&format!("repos/o/r/statuses/{H}")),
            "{recorded:?}"
        );
        assert_eq!(
            out.status.code(),
            *want_code,
            "{name}: {}",
            String::from_utf8_lossy(&out.stderr)
        );
    }
}
