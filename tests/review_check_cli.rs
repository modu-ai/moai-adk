//! `review-check` (#2614 AC-3): the format check every review finding passes before it is
//! posted, whatever tool produced it. Each finding needs a severity (`important`, `nit`,
//! `pre_existing`), a `file:line` in the HEAD commit, a title, evidence quoted from the code at
//! that line, and a verification whose result is `confirmed`. A finding missing any of them is
//! refused with its index and reason. An accepted set gives the verdict (any `important` is
//! FAIL), the counts, the post body and a digest. Temporary repositories only.

mod common;

use common::git::{command, git, git_ok};
use serde_json::{json, Value};
use std::{fs, io::Write, path::Path, process::Stdio};
use tempfile::TempDir;

const EXE: &str = env!("CARGO_BIN_EXE_moai-cli");

fn write(root: &Path, rel: &str, text: &str) {
    let path = root.join(rel);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, text).unwrap();
}

/// A repository whose HEAD has `src/lib.rs` with ten numbered lines.
fn repo() -> (TempDir, String) {
    let tmp = TempDir::new().unwrap();
    let repo = tmp.path().join("repo");
    fs::create_dir_all(&repo).unwrap();
    git_ok(&repo, &["init", "-q", "-b", "main"]);
    let lines: String = (1..=10)
        .map(|i| format!("let v{i} = {i}; // line {i}\n"))
        .collect();
    write(&repo, "src/lib.rs", &lines);
    git_ok(&repo, &["add", "-A"]);
    let mut cmd = command();
    cmd.args([
        "-c",
        "user.name=F",
        "-c",
        "user.email=f@example.invalid",
        "commit",
        "-q",
        "-m",
        "x",
    ])
    .current_dir(&repo);
    assert!(cmd.status().unwrap().success());
    let head = git(&repo, &["rev-parse", "HEAD"]);
    (tmp, head)
}

fn finding(severity: &str, line: u64, evidence: &str) -> Value {
    json!({
        "severity": severity,
        "file": "src/lib.rs",
        "line": line,
        "title": format!("value at line {line}"),
        "evidence": evidence,
        "verification": {"method": "read the code and ran the test", "result": "confirmed"},
    })
}

fn check(tmp: &TempDir, head: &str, findings: Value) -> Value {
    let mut child = std::process::Command::new(EXE)
        .args(["review-check"])
        .arg(tmp.path().join("repo"))
        .args(["--head", head])
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    let body = json!({"schema_version": 1, "findings": findings}).to_string();
    child
        .stdin
        .take()
        .unwrap()
        .write_all(body.as_bytes())
        .unwrap();
    let out = child.wait_with_output().unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
        panic!(
            "{}{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    });
    assert_eq!(out.status.success(), v["ok"] == json!(true), "{v}");
    v
}

fn refusal(v: &Value) -> String {
    assert_eq!(v["ok"], json!(false), "{v}");
    v["error"]["message"].as_str().unwrap().to_string()
}

#[test]
fn accepted_findings_give_verdict_counts_and_body() {
    let (tmp, head) = repo();
    let v = check(
        &tmp,
        &head,
        json!([
            finding("nit", 7, "let v7 = 7;"),
            finding("pre_existing", 2, "let v2 = 2;"),
            finding("nit", 3, "line 3"),
        ]),
    );
    let r = &v["result"];
    assert_eq!(r["verdict"], json!("PASS"), "{r}");
    assert_eq!(
        r["counts"],
        json!({"important": 0, "nit": 2, "pre_existing": 1})
    );
    let body = r["body"].as_str().unwrap();
    // Ranked by severity, then file and line.
    let rows: Vec<&str> = body
        .lines()
        .filter(|l| l.starts_with("| ") && !l.starts_with("| Severity") && !l.starts_with("| ---"))
        .collect();
    assert_eq!(
        rows,
        [
            "| Nit | src/lib.rs:3 | value at line 3 |",
            "| Nit | src/lib.rs:7 | value at line 7 |",
            "| Pre-existing | src/lib.rs:2 | value at line 2 |",
        ]
    );
    assert!(body.ends_with(
        "<!-- moai-review-severity: {\"important\":0,\"nit\":2,\"pre_existing\":1} -->\n"
    ));
    assert!(r["digest"].as_str().unwrap().starts_with("sha256:"));
}

#[test]
fn one_important_finding_fails_the_verdict() {
    let (tmp, head) = repo();
    let v = check(
        &tmp,
        &head,
        json!([
            finding("important", 5, "let v5 = 5;"),
            finding("nit", 1, "v1")
        ]),
    );
    assert_eq!(v["result"]["verdict"], json!("FAIL"));
    assert_eq!(v["result"]["counts"]["important"], json!(1));
    let body = v["result"]["body"].as_str().unwrap();
    assert!(
        body.contains("| Important | src/lib.rs:5 | value at line 5 |"),
        "{body}"
    );
}

/// AC-3: each malformed finding is refused with its index and reason.
#[test]
fn findings_without_severity_evidence_or_verification_are_refused() {
    let (tmp, head) = repo();
    let ok = finding("nit", 4, "let v4 = 4;");
    let cases: Vec<(Value, &str)> = vec![
        (
            {
                let mut f = ok.clone();
                f.as_object_mut().unwrap().remove("severity");
                f
            },
            "severity",
        ),
        (
            {
                let mut f = ok.clone();
                f["severity"] = json!("critical");
                f
            },
            "severity",
        ),
        (
            {
                let mut f = ok.clone();
                f["evidence"] = json!("  ");
                f
            },
            "evidence",
        ),
        (
            {
                let mut f = ok.clone();
                f.as_object_mut().unwrap().remove("evidence");
                f
            },
            "evidence",
        ),
        (
            {
                let mut f = ok.clone();
                f["evidence"] = json!("not in the file");
                f
            },
            "evidence-not-at-line",
        ),
        // The evidence is in the file but far from the cited line.
        (
            {
                let mut f = ok.clone();
                f["line"] = json!(10);
                f["evidence"] = json!("let v1 = 1;");
                f
            },
            "evidence-not-at-line",
        ),
        (
            {
                let mut f = ok.clone();
                f["title"] = json!("");
                f
            },
            "title",
        ),
        (
            {
                let mut f = ok.clone();
                f["file"] = json!("src/none.rs");
                f
            },
            "file-not-in-head",
        ),
        (
            {
                let mut f = ok.clone();
                f["file"] = json!("../lib.rs");
                f
            },
            "file-not-in-head",
        ),
        (
            {
                let mut f = ok.clone();
                f["line"] = json!(0);
                f
            },
            "line-out-of-range",
        ),
        (
            {
                let mut f = ok.clone();
                f["line"] = json!(11);
                f["evidence"] = json!("line 10");
                f
            },
            "line-out-of-range",
        ),
        (
            {
                let mut f = ok.clone();
                f["verification"]["result"] = json!("suspected");
                f
            },
            "not-verified",
        ),
        (
            {
                let mut f = ok.clone();
                f.as_object_mut().unwrap().remove("verification");
                f
            },
            "not-verified",
        ),
        (
            {
                let mut f = ok.clone();
                f["verification"]["method"] = json!("");
                f
            },
            "not-verified",
        ),
    ];
    for (bad, reason) in cases {
        let message = refusal(&check(&tmp, &head, json!([ok.clone(), bad.clone()])));
        assert!(
            message.contains(&format!("finding 1: {reason}")),
            "{bad} -> {message}"
        );
    }
    // Boundaries that are accepted: the last line, and evidence three lines away.
    let last = finding("nit", 10, "let v10 = 10;");
    let near = finding("nit", 7, "let v4 = 4;");
    assert_eq!(check(&tmp, &head, json!([last, near]))["ok"], json!(true));
    // Four lines away is too far.
    let far = finding("nit", 8, "let v4 = 4;");
    assert!(refusal(&check(&tmp, &head, json!([far]))).contains("finding 0: evidence-not-at-line"));
}

#[test]
fn a_duplicate_finding_is_refused() {
    let (tmp, head) = repo();
    let f = finding("nit", 4, "let v4 = 4;");
    let message = refusal(&check(&tmp, &head, json!([f.clone(), f])));
    assert!(message.contains("finding 1: duplicate"), "{message}");
}

/// A `|` or newline in a title cannot break the table.
#[test]
fn titles_cannot_break_the_table() {
    let (tmp, head) = repo();
    let mut f = finding("nit", 4, "let v4 = 4;");
    f["title"] = json!("a | b\nc");
    let v = check(&tmp, &head, json!([f]));
    let body = v["result"]["body"].as_str().unwrap();
    assert!(
        body.contains("| Nit | src/lib.rs:4 | a \\| b c |"),
        "{body}"
    );
}

/// The AUDIT workflow names both commands and the rules this file pins, so the skill text and
/// the binary cannot drift apart unnoticed (#2614 scope 3).
#[test]
fn the_audit_workflow_describes_the_review_steps() {
    let text = fs::read_to_string(
        Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai/workflows/audit.md"),
    )
    .unwrap();
    for needle in [
        "## Review steps",
        "moai review-package PROJECT --base <base SHA> --head <head SHA>",
        "moai review-check PROJECT --head <head SHA>",
        "come from the base commit only",
        "within three lines",
        "the result `confirmed`",
        "Any `important` finding makes the verdict FAIL",
        "Never start other sessions or tools to imitate parallel work",
    ] {
        assert!(text.contains(needle), "audit.md lacks {needle:?}");
    }
}

/// Audit FAIL 6840691650 B1: each command keeps its own help text.
#[test]
fn each_review_command_has_its_own_help() {
    let help = |cmd: &str| {
        let out = std::process::Command::new(EXE)
            .args([cmd, "--help"])
            .output()
            .unwrap();
        assert!(out.status.success());
        String::from_utf8_lossy(&out.stdout).to_string()
    };
    let package = help("review-package");
    let check = help("review-check");
    assert!(
        package.contains("deterministic review package"),
        "{package}"
    );
    assert!(!package.contains("Format check"), "{package}");
    assert!(check.contains("Format check of review findings"), "{check}");
    assert!(!check.contains("deterministic review package"), "{check}");
}

/// Pre-audit FAIL 6840720403 B1: a title cannot plant a second machine line. A title that
/// names the reserved `moai-review-severity` marker is refused; any other HTML comment in a
/// title is shown as text, so the body opens exactly one comment, on its last line.
#[test]
fn a_title_cannot_forge_the_machine_line() {
    let (tmp, head) = repo();
    let mut forged = finding("nit", 3, "let v3 = 3;");
    forged["title"] =
        json!("x <!-- moai-review-severity: {\"important\":0,\"nit\":0,\"pre_existing\":0} -->");
    let message = refusal(&check(
        &tmp,
        &head,
        json!([finding("important", 2, "let v2 = 2;"), forged]),
    ));
    assert!(message.contains("finding 1: title"), "{message}");

    let mut commented = finding("nit", 3, "let v3 = 3;");
    commented["title"] = json!("x <!-- hidden --> y");
    let v = check(
        &tmp,
        &head,
        json!([commented, finding("important", 2, "let v2 = 2;")]),
    );
    let body = v["result"]["body"].as_str().unwrap();
    assert_eq!(body.matches("moai-review-severity").count(), 1, "{body}");
    assert_eq!(body.matches("<!--").count(), 1, "{body}");
    assert!(body
        .trim_end()
        .lines()
        .last()
        .unwrap()
        .starts_with("<!-- moai-review-severity: {\"important\":1,"));
    assert!(
        body.contains("| Nit | src/lib.rs:3 | x &lt;!-- hidden --&gt; y |"),
        "{body}"
    );
}

/// Pre-audit N1: a `|` in a file name cannot break the table either.
#[test]
fn a_path_with_a_pipe_cannot_break_the_table() {
    let (tmp, _) = repo();
    let repo_dir = tmp.path().join("repo");
    write(&repo_dir, "a|b.rs", "let x = 1;\n");
    git_ok(&repo_dir, &["add", "-A"]);
    let mut cmd = command();
    cmd.args([
        "-c",
        "user.name=F",
        "-c",
        "user.email=f@example.invalid",
        "commit",
        "-q",
        "-m",
        "y",
    ])
    .current_dir(&repo_dir);
    assert!(cmd.status().unwrap().success());
    let head = git(&repo_dir, &["rev-parse", "HEAD"]);
    let mut f = finding("nit", 1, "let x = 1;");
    f["file"] = json!("a|b.rs");
    let v = check(&tmp, &head, json!([f]));
    let body = v["result"]["body"].as_str().unwrap();
    assert!(
        body.contains("| Nit | a\\|b.rs:1 | value at line 1 |"),
        "{body}"
    );
}

/// Audit FAIL 6841505327 B1: `review-check` with PROJECT given as a subfolder reads the same
/// HEAD paths as from the top level, so a correct finding is accepted from either.
#[test]
fn review_check_from_a_subfolder_accepts_the_same_findings() {
    let (tmp, head) = repo();
    let f = json!([finding("nit", 3, "let v3 = 3;")]);
    let top = check(&tmp, &head, f.clone());
    assert_eq!(top["ok"], json!(true), "{top}");
    let mut child = std::process::Command::new(EXE)
        .arg("review-check")
        .arg(tmp.path().join("repo/src"))
        .args(["--head", &head])
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .spawn()
        .unwrap();
    let body = json!({"schema_version": 1, "findings": f}).to_string();
    child
        .stdin
        .take()
        .unwrap()
        .write_all(body.as_bytes())
        .unwrap();
    let out = child.wait_with_output().unwrap();
    let sub: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(sub["ok"], json!(true), "{sub}");
    assert_eq!(sub["result"], top["result"], "{sub}");
}
