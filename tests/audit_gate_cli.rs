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

/// The interpreter the step would reach when the test's own stand-in is not first on PATH: the
/// first `python3` along the caller's PATH, asked for its own executable so that a version
/// manager's shim (which reads HOME) is not needed once the step's HOME is the test's.
fn real_python3() -> std::path::PathBuf {
    let first = std::env::split_paths(&std::env::var_os("PATH").unwrap_or_default())
        .map(|d| d.join("python3"))
        .find(|p| p.is_file())
        .expect("a python3 on PATH: the step parses the gate's envelope with it");
    let out = std::process::Command::new(&first)
        .args(["-c", "import sys; print(sys.executable)"])
        // This question runs in the test's own HOME too: no bytecode cache there either.
        .env("PYTHONDONTWRITEBYTECODE", "1")
        .output()
        .expect("ask python3 for its executable");
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    std::path::PathBuf::from(String::from_utf8_lossy(&out.stdout).trim())
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
        // SPEC-20261124-cleanup-leftovers REQ-16: the step's shell runs in a HOME of its own that
        // must stay empty, whichever `python3` is first on PATH. The Xcode `python3` of macOS
        // keeps a bytecode cache under `$HOME/Library/Caches` unless it is told not to; the
        // stand-in does the same on every machine, then runs the real `python3` that parses the
        // envelope.
        let step_home = tmp.path().join("home");
        fs::create_dir_all(&step_home).unwrap();
        fs::write(
            bin.join("python3"),
            format!(
                "#!/bin/sh\nif [ -z \"$PYTHONDONTWRITEBYTECODE\" ]; then mkdir -p \"$HOME/Library/Caches/com.apple.python\" && : > \"$HOME/Library/Caches/com.apple.python/step.pyc\"; fi\nexec '{}' \"$@\"\n",
                real_python3().display()
            ),
        )
        .unwrap();
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
        for exe in [bin.join("gh"), bin.join("python3"), target.join("moai")] {
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
            .env("HOME", &step_home)
            // Whichever `python3` the step reaches keeps its bytecode cache out of the step's HOME
            // (the Xcode one of macOS would write under `$HOME/Library/Caches`).
            .env("PYTHONDONTWRITEBYTECODE", "1")
            .current_dir(tmp.path())
            .output()
            .expect("run the step block");
        let at_post = fs::read_to_string(bin.join("runner-temp-at-post.txt")).unwrap_or_default();
        assert!(
            at_post.lines().any(|l| l.starts_with("audit-gate.")),
            "{name}: the envelope was not made in RUNNER_TEMP: {at_post:?}"
        );
        // Listed with `ls -A`, as the stub gh lists RUNNER_TEMP: this test walks only its own
        // temporary folders, and the repository-scan table is for tests that walk the tree.
        for dir in [&runner_temp, &tmpdir, &step_home] {
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

/// SPEC-20261125-release-audit-summary AC-5 (REQ-11 to REQ-14): a release summary comment of a
/// member decides the gate through the command, with the merge base the input carries. The
/// summary names landings and sessions only by id, and the gate reads it without a commit list.
#[test]
fn a_release_summary_comment_decides_the_gate() {
    use moai_cli::release_audit_summary::{self as summary, Kind, Line, Summary, Verdict};

    const MB: &str = "4444444444444444444444444444444444444444";
    const S_AUTH: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
    const S_REV: &str = "cccccccc-0000-4000-8000-00000000000c";

    let pass = |excluded: Vec<usize>| {
        Kind::Pass(vec![Verdict {
            card: 1,
            head: "abcdef012345".into(),
            tier: 0,
            reviewer: 1,
            excluded,
            event: 7,
        }])
    };
    // Two landings from `base` to `H`: the first one passes, the second is `second`.
    let comment = |id: u64, base: &str, second: Kind, association: &str| -> Value {
        let first = "5".repeat(40);
        let s = Summary {
            head: H.into(),
            base: base.into(),
            sessions: vec![S_AUTH.into(), S_REV.into()],
            lines: vec![
                Line {
                    sha: first.clone(),
                    parent: base[..12].into(),
                    kind: pass(vec![0]),
                },
                Line {
                    sha: H.into(),
                    parent: first[..12].into(),
                    kind: second,
                },
            ],
        };
        json!({"id": id, "author_association": association,
               "body": summary::comment(&s).unwrap()})
    };
    // A release PR: no file list and no commit list, so tier 2 and unknown authors.
    let release = |comments: Vec<Value>, merge_base: Option<&str>| -> Value {
        let mut v = json!({"schema_version": 1, "pr": 9, "head": H,
                           "body": "audit-tier: 2 release\n\nbody", "comments": comments});
        if let Some(b) = merge_base {
            v["release_base"] = json!(b);
        }
        v
    };

    // A complete summary of a non-author auditor passes, and `--require-pass` exits 0.
    let good = comment(41, MB, pass(vec![0]), "MEMBER");
    let r = decide(&release(vec![good.clone()], Some(MB)));
    assert_eq!(r["state"], "pass", "{r}");
    assert_eq!(r["rule"], "release-summary", "{r}");
    assert_eq!(r["tier"], json!(2), "{r}");
    assert_eq!(r["comment"], json!(41), "{r}");
    let ok = run(&release(vec![good.clone()], Some(MB)), &["--require-pass"]);
    assert_eq!(
        ok.status.code(),
        Some(0),
        "{}",
        String::from_utf8_lossy(&ok.stderr)
    );

    // Without the merge base, or with another one, the summary does not match the PR.
    for merge_base in [None, Some(H)] {
        let r = decide(&release(vec![good.clone()], merge_base));
        assert_eq!(
            (r["state"].clone(), r["rule"].clone()),
            (json!("missing"), json!("summary-base")),
            "{merge_base:?}: {r}"
        );
    }

    // A landing with no verdict is missing, by name, and the job's exit code is 1.
    let none = Kind::None {
        cards: vec![3],
        title: "a direct commit".into(),
    };
    let gap = release(vec![comment(41, MB, none, "MEMBER")], Some(MB));
    let r = decide(&gap);
    assert_eq!(
        (r["state"].clone(), r["rule"].clone()),
        (json!("missing"), json!("summary-landings")),
        "{r}"
    );
    assert!(r["reason"].as_str().unwrap().contains(&H[..12]), "{r}");
    let out = run(&gap, &["--require-pass"]);
    assert_eq!(out.status.code(), Some(1));
    assert!(String::from_utf8_lossy(&out.stdout).contains("audit gate missing"));

    // An auditor that is among the excluded sessions is no pass.
    let own = comment(41, MB, pass(vec![0, 1]), "MEMBER");
    let r = decide(&release(vec![own], Some(MB)));
    assert_eq!(r["rule"], "summary-landings", "{r}");

    // A FAIL line fails the gate; a non-member's summary is not read at all.
    let fail = Kind::Fail {
        why: "late_fail".into(),
        cards: vec![2],
        title: "a late FAIL".into(),
    };
    let r = decide(&release(
        vec![comment(41, MB, fail.clone(), "MEMBER")],
        Some(MB),
    ));
    assert_eq!(
        (r["state"].clone(), r["rule"].clone()),
        (json!("fail"), json!("summary-fail")),
        "{r}"
    );
    let r = decide(&release(vec![comment(41, MB, fail, "NONE")], Some(MB)));
    assert_eq!(r["rule"], "pass-on-head", "{r}");
    let r = decide(&release(
        vec![comment(41, MB, pass(vec![0]), "NONE")],
        Some(MB),
    ));
    assert_eq!(r["rule"], "pass-on-head", "{r}");

    // A merge base that is no string is a malformed input, refused, not ignored.
    let mut bad = release(vec![good], Some(MB));
    bad["release_base"] = json!(5);
    assert_eq!(run(&bad, &[]).status.code(), Some(1));
}
