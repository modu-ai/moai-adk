//! SPEC-20261120-release-pr-three-os: a release pull request (head branch `release/<40 hex>` or
//! `release/rc/<40 hex>`) runs the check set on Windows x86_64 before merge, and the `ci`
//! aggregate fails unless that job succeeded; every other run is judged as before. The tests run
//! the workflow's real step scripts in a temporary folder, with stand-in `gh` and `python3`
//! programs that record what they were asked. Unix only: the stand-ins are sh scripts.

#![cfg(unix)]

use serde_json::{json, Value};
use std::{
    fs,
    os::unix::fs::PermissionsExt,
    path::{Path, PathBuf},
    process::{Command, Stdio},
};
use tempfile::TempDir;

/// A commit id of 40 lowercase hex digits, the shape a release branch name carries.
const HEX: &str = "0123456789abcdef0123456789abcdef01234567";

/// The stand-in for the files API: it answers the three calls of the decide step and fails on
/// request.
const GH: &str = r#"if [ "${STUB_API_FAIL:-0}" = 1 ]; then exit 1; fi
case "$*" in
  *"--jq length"*) printf '2\n' ;;
  *changedFiles*) printf '2\n' ;;
  *) printf 'README.md\nsrc/lib.rs\n' ;;
esac
"#;

/// The stand-in for Python. A call whose first argument is not `-c` is the platform-sensitive
/// classifier: it answers with the value of `STUB_MACOS_NEEDED` and fails on request. A `-c`
/// call is one of the two lines that read that answer. The stand-in tells the calls apart by
/// their first argument, so this file spells no path under `tools/` and the file-reader scan of
/// the repository does not take it for a reader of the classifier it replaces.
const PYTHON3: &str = r#"case "$1" in
  -c) printf '%s\n' "${STUB_MACOS_NEEDED:-False}" ;;
  *)
    if [ "${STUB_SCOPE_FAIL:-0}" = 1 ]; then exit 1; fi
    printf '{"scope": "full", "macos_needed": "%s", "macos_reasons": []}\n' "${STUB_MACOS_NEEDED:-False}"
    ;;
esac
"#;

fn repo_file(rel: &str) -> String {
    fs::read_to_string(Path::new(env!("CARGO_MANIFEST_DIR")).join(rel)).unwrap()
}

/// The body of the `run: |` block of the workflow step with the given `id:`, dedented.
fn step_script(workflow: &str, id: &str) -> String {
    let lines: Vec<&str> = workflow.lines().collect();
    let id_line = format!("id: {id}");
    let at = lines
        .iter()
        .position(|l| l.trim() == id_line)
        .unwrap_or_else(|| panic!("no step `{id_line}`"));
    let run = at
        + lines[at..]
            .iter()
            .position(|l| l.trim() == "run: |")
            .expect("the step has a `run: |` block");
    let indent = lines[run].len() - lines[run].trim_start().len() + 2;
    let mut script = String::new();
    for l in &lines[run + 1..] {
        if l.trim().is_empty() {
            script.push('\n');
        } else if l.len() - l.trim_start().len() >= indent {
            script.push_str(&l[indent..]);
            script.push('\n');
        } else {
            break;
        }
    }
    script
}

/// The lines of the top-level job `job` of a workflow file: from `  <job>:` up to the next
/// top-level key or the end of the file.
fn job_block<'a>(workflow: &'a str, job: &str) -> Vec<&'a str> {
    let lines: Vec<&str> = workflow.lines().collect();
    let head = format!("  {job}:");
    let at = lines
        .iter()
        .position(|l| *l == head)
        .unwrap_or_else(|| panic!("no job `{job}`"));
    let end = lines[at + 1..]
        .iter()
        .position(|l| is_key(l))
        .map(|n| at + 1 + n)
        .unwrap_or(lines.len());
    lines[at..end].to_vec()
}

/// A line that opens a top-level key of the `jobs:` mapping: two spaces, a name and a colon.
fn is_key(line: &str) -> bool {
    match line.strip_prefix("  ") {
        Some(rest) if !rest.starts_with(' ') && !rest.starts_with('#') => rest.ends_with(':'),
        _ => false,
    }
}

/// The `key: value` lines of the `with:` block of a job block, in order.
fn inputs(job: &[&str]) -> Vec<(String, String)> {
    let at = job
        .iter()
        .position(|l| l.trim() == "with:")
        .expect("the job has a `with:` block");
    job[at + 1..]
        .iter()
        .take_while(|l| l.trim().is_empty() || l.len() - l.trim_start().len() >= 6)
        .filter_map(|l| {
            l.trim()
                .split_once(": ")
                .map(|(k, v)| (k.to_string(), v.trim().to_string()))
        })
        .collect()
}

struct Ran {
    ok: bool,
    stderr: String,
}

/// Runs a step script the way the runner runs it (`bash -eo pipefail`), with only the given
/// environment beside the one the test process has.
fn run(script: &str, dir: &Path, env: &[(&str, &str)]) -> Ran {
    let out = Command::new("bash")
        .args(["-eo", "pipefail", "-c", script])
        .current_dir(dir)
        .envs(env.iter().copied())
        .stdin(Stdio::null())
        .output()
        .unwrap();
    Ran {
        ok: out.status.success(),
        stderr: String::from_utf8_lossy(&out.stderr).into_owned(),
    }
}

/// A stand-in program: a sh script that records its name and its arguments in `STUB_LOG` as
/// `<name> <arguments>`, then runs `body`.
fn stand_in(bin: &Path, name: &str, body: &str) {
    let path = bin.join(name);
    fs::write(
        &path,
        format!("#!/bin/sh\nprintf '%s %s\\n' '{name}' \"$*\" >> \"$STUB_LOG\"\n{body}"),
    )
    .unwrap();
    fs::set_permissions(&path, fs::Permissions::from_mode(0o755)).unwrap();
}

/// A temporary folder with the stand-in programs on its `PATH`, what they were asked, and the
/// file a step writes its outputs to.
struct Folder {
    dir: TempDir,
    bin: PathBuf,
    log: PathBuf,
    out: PathBuf,
}

impl Folder {
    fn with_stand_ins() -> Self {
        let dir = TempDir::new().unwrap();
        let bin = dir.path().join("bin");
        fs::create_dir(&bin).unwrap();
        stand_in(&bin, "gh", GH);
        stand_in(&bin, "python3", PYTHON3);
        let log = dir.path().join("calls.log");
        let out = dir.path().join("step-output.txt");
        fs::write(&out, "").unwrap();
        Self { dir, bin, log, out }
    }

    /// The `PATH` the step sees: the stand-ins first, then the system tools it uses.
    fn path(&self) -> String {
        format!("{}:/usr/bin:/bin", self.bin.display())
    }

    fn calls(&self) -> Vec<String> {
        fs::read_to_string(&self.log)
            .unwrap_or_default()
            .lines()
            .map(str::to_string)
            .collect()
    }

    /// The value a step wrote to its output file as `<key>=<value>`, or the empty string.
    fn output(&self, key: &str) -> String {
        let prefix = format!("{key}=");
        fs::read_to_string(&self.out)
            .unwrap_or_default()
            .lines()
            .find_map(|l| l.strip_prefix(prefix.as_str()))
            .unwrap_or_default()
            .to_string()
    }
}

struct Decision {
    ok: bool,
    release: String,
    /// The `matrix` output as far as it parses as JSON, else `Value::Null`.
    matrix: Value,
    calls: Vec<String>,
    stderr: String,
}

/// Runs the real decide step of the `ci` workflow for an event and a head branch name, with the
/// given settings for the stand-ins.
fn decide(event: &str, head_ref: &str, stubs: &[(&str, &str)]) -> Decision {
    let script = step_script(&repo_file(".github/workflows/ci.yml"), "decide")
        .replace("${{ github.event_name }}", event);
    let f = Folder::with_stand_ins();
    let temp = f.dir.path().join("runner-temp");
    fs::create_dir(&temp).unwrap();
    let (path, log, out, temp) = (
        f.path(),
        f.log.to_string_lossy().into_owned(),
        f.out.to_string_lossy().into_owned(),
        temp.to_string_lossy().into_owned(),
    );
    let mut env: Vec<(&str, &str)> = vec![
        ("PATH", path.as_str()),
        ("GITHUB_OUTPUT", out.as_str()),
        ("GH_TOKEN", "x"),
        ("PR_NUMBER", "7"),
        ("REPO", "example/example"),
        ("RUNNER_TEMP", temp.as_str()),
        ("STUB_LOG", log.as_str()),
        ("HEAD_REF", head_ref),
    ];
    env.extend_from_slice(stubs);
    let ran = run(&script, f.dir.path(), &env);
    let matrix_text = f.output("matrix");
    Decision {
        ok: ran.ok,
        release: f.output("release"),
        matrix: serde_json::from_str(&matrix_text).unwrap_or(Value::Null),
        calls: f.calls(),
        stderr: ran.stderr,
    }
}

/// The matrix every run gets that is not narrowed: both targets of the support decision, each
/// as a nextest leg and a rest leg.
fn full_matrix() -> Value {
    json!([
        {"target": "linux-x86_64", "runner": "ubuntu-24.04", "timeout": 30, "leg": "nextest"},
        {"target": "linux-x86_64", "runner": "ubuntu-24.04", "timeout": 30, "leg": "rest"},
        {"target": "macos-arm64", "runner": "macos-15", "timeout": 45, "leg": "nextest"},
        {"target": "macos-arm64", "runner": "macos-15", "timeout": 45, "leg": "rest"},
    ])
}

/// The matrix a pull request with no platform-sensitive path gets: the two linux legs.
fn linux_matrix() -> Value {
    json!([
        {"target": "linux-x86_64", "runner": "ubuntu-24.04", "timeout": 30, "leg": "nextest"},
        {"target": "linux-x86_64", "runner": "ubuntu-24.04", "timeout": 30, "leg": "rest"},
    ])
}

/// Whether the release workflow's plan step accepts `name` as a release branch: a name it accepts
/// gets past its three name checks and stops at the missing `HEADVER` file of an empty folder.
fn release_workflow_plan_accepts(name: &str) -> bool {
    let script = step_script(&repo_file(".github/workflows/release.yml"), "plan");
    let dir = TempDir::new().unwrap();
    let ran = run(
        &script,
        dir.path(),
        &[
            ("REF_NAME", name),
            ("PUSHED", HEX),
            ("RUN_NUMBER", "7"),
            ("RUN_ATTEMPT", "1"),
        ],
    );
    ran.stderr.contains("HEADVER refused: the file is missing")
}

/// A pull request whose head branch is a release name runs the full matrix and reports a release
/// pull request; the classifier and the files API are not asked for it.
#[test]
fn a_release_pull_request_gets_the_full_matrix_and_the_windows_flag() {
    for head in [format!("release/{HEX}"), format!("release/rc/{HEX}")] {
        let d = decide("pull_request", &head, &[]);
        assert!(d.ok, "{head}: {}", d.stderr);
        assert_eq!(d.matrix, full_matrix(), "{head}");
        assert_eq!(d.release, "true", "{head}");
        assert!(
            d.calls.is_empty(),
            "{head}: the classifier and the files API are not asked: {:?}",
            d.calls
        );
    }
}

/// The names the decide step takes as a release pull request are exactly the names the release
/// workflow's plan step numbers; both real scripts answer for each name of one list.
#[test]
fn the_decide_step_and_the_release_workflow_accept_the_same_branch_names() {
    let names = [
        format!("release/{HEX}"),
        format!("release/rc/{HEX}"),
        format!("release/{}", &HEX[..39]),
        format!("release/{HEX}0"),
        format!("release/{}", HEX.to_uppercase()),
        "release/v4.2648.23".to_string(),
        format!("candidate/{HEX}"),
        format!("Release/{HEX}"),
        format!("release/{HEX}/x"),
        "release/".to_string(),
    ];
    let mut accepted = 0;
    for name in &names {
        let d = decide("pull_request", name, &[]);
        let plan = release_workflow_plan_accepts(name);
        assert_eq!(
            d.release == "true",
            plan,
            "{name}: the decide step says release={:?}, the release workflow {}",
            d.release,
            if plan {
                "accepts the name"
            } else {
                "refuses the name"
            }
        );
        accepted += usize::from(plan);
    }
    assert_eq!(
        accepted, 2,
        "the list holds two release names and eight others"
    );
}

/// An ordinary pull request keeps today's rule: the classifier narrows the matrix to linux when
/// it answers not needed, the full matrix stays when it answers needed or fails, and a branch
/// that only looks like a release is an ordinary pull request.
#[test]
fn an_ordinary_pull_request_keeps_the_classifier_rule() {
    let d = decide(
        "pull_request",
        "feature/login-fix",
        &[("STUB_MACOS_NEEDED", "False")],
    );
    assert!(d.ok, "{}", d.stderr);
    assert_eq!(d.matrix, linux_matrix());
    assert_eq!(d.release, "false");
    assert!(
        d.calls.iter().any(|c| c.contains("/pulls/7/files")),
        "the files API is asked: {:?}",
        d.calls
    );
    // A Python call that is not the JSON reader `-c` is the classifier.
    assert!(
        d.calls
            .iter()
            .any(|c| c.starts_with("python3 ") && !c.contains(" -c ")),
        "the classifier is asked: {:?}",
        d.calls
    );

    let d = decide(
        "pull_request",
        "feature/login-fix",
        &[("STUB_MACOS_NEEDED", "True")],
    );
    assert!(d.ok, "{}", d.stderr);
    assert_eq!(d.matrix, full_matrix());
    assert_eq!(d.release, "false");

    for stubs in [[("STUB_SCOPE_FAIL", "1")], [("STUB_API_FAIL", "1")]] {
        let d = decide("pull_request", "feature/login-fix", &stubs);
        assert!(d.ok, "{stubs:?}: {}", d.stderr);
        assert_eq!(d.matrix, full_matrix(), "{stubs:?}");
        assert_eq!(d.release, "false", "{stubs:?}");
    }

    for head in [
        format!("release/{}", &HEX[..39]),
        "release/v4.2648.23".to_string(),
        format!("Release/{HEX}"),
    ] {
        let d = decide("pull_request", &head, &[("STUB_MACOS_NEEDED", "False")]);
        assert!(d.ok, "{head}: {}", d.stderr);
        assert_eq!(d.release, "false", "{head}");
        assert_eq!(d.matrix, linux_matrix(), "{head}");
    }
}

/// A push and a manual run keep the full matrix, report no release pull request and ask nothing;
/// the event decides, not the branch name.
#[test]
fn push_and_dispatch_keep_the_full_matrix() {
    for event in ["push", "workflow_dispatch"] {
        for head in ["", &format!("release/{HEX}")] {
            let d = decide(event, head, &[]);
            assert!(d.ok, "{event} {head:?}: {}", d.stderr);
            assert_eq!(d.matrix, full_matrix(), "{event} {head:?}");
            assert_eq!(d.release, "false", "{event} {head:?}");
            assert!(d.calls.is_empty(), "{event} {head:?}: {:?}", d.calls);
        }
    }
}

/// The Windows job of the `ci` workflow calls the target workflow the candidate verification
/// calls, with its five Windows inputs unchanged, on the pull request's head commit, and only
/// for a release pull request. Its name stays outside the shapes other readers take for jobs of
/// the matrix: the merge gate (`src/gate.rs` `watched_check`) and the superseded verdict
/// (`tools/ci_superseded.py`) read every check run named `ci / …` and count a skipped one as
/// failed, and this job is skipped on every ordinary run.
#[test]
fn the_windows_job_calls_the_candidate_target_with_the_candidates_windows_inputs() {
    let ci = repo_file(".github/workflows/ci.yml");
    let job = job_block(&ci, "windows-x86_64");
    let line = |text: &str| job.iter().any(|l| l.trim() == text);
    assert!(line("name: windows-check"), "{job:#?}");
    assert!(
        !job.iter().any(|l| l.trim().starts_with("name: ci")),
        "{job:#?}"
    );
    assert!(
        line("uses: ./.github/workflows/candidate-target.yml"),
        "{job:#?}"
    );
    assert!(
        line("if: ${{ needs.targets.outputs.release == 'true' }}"),
        "{job:#?}"
    );
    let candidate = inputs(&job_block(
        &repo_file(".github/workflows/candidate.yml"),
        "windows-x86_64",
    ));
    let got = inputs(&job);
    let without_sha = |v: &[(String, String)]| {
        v.iter()
            .filter(|(k, _)| k != "sha")
            .cloned()
            .collect::<Vec<_>>()
    };
    assert_eq!(without_sha(&got), without_sha(&candidate));
    let five = [
        ("target", "windows-x86_64"),
        ("runner", "windows-2025"),
        ("timeout", "60"),
        ("nextest_asset", "x86_64-pc-windows-msvc"),
        (
            "nextest_sha256",
            "e51bdb4b5883df48f335381b287378edb244318fac65ae0adaf6b9e69298246b",
        ),
    ]
    .map(|(k, v)| (k.to_string(), v.to_string()));
    assert_eq!(without_sha(&got), five);
    assert_eq!(
        got.iter()
            .find(|(k, _)| k == "sha")
            .map(|(_, v)| v.as_str()),
        Some("${{ github.event.pull_request.head.sha }}"),
        "{got:?}"
    );
}

/// The real step of the `ci` aggregate, run over every result its jobs can have: an ordinary
/// run passes with the Windows job skipped, a release pull request passes only when the Windows
/// check set succeeded, and a failed matrix never passes.
#[test]
fn the_aggregate_fails_when_a_release_pull_requests_windows_check_did_not_succeed() {
    let script = step_script(&repo_file(".github/workflows/ci.yml"), "aggregate");
    for (check, release, windows, passes) in [
        ("success", "false", "skipped", true),
        ("success", "false", "success", true),
        ("success", "true", "success", true),
        ("success", "true", "skipped", false),
        ("success", "true", "failure", false),
        ("success", "true", "cancelled", false),
        ("success", "true", "", false),
        ("success", "", "skipped", false),
        ("failure", "false", "skipped", false),
        ("failure", "true", "success", false),
        ("cancelled", "true", "success", false),
        ("skipped", "", "skipped", false),
    ] {
        let dir = TempDir::new().unwrap();
        let summary = dir.path().join("summary.md").to_string_lossy().into_owned();
        let env = [
            ("PATH", "/usr/bin:/bin"),
            ("RELEASE", release),
            ("WINDOWS", windows),
            ("GITHUB_STEP_SUMMARY", summary.as_str()),
        ];
        let ran = run(
            &script.replace("${{ needs.check.result }}", check),
            dir.path(),
            &env,
        );
        assert_eq!(
            ran.ok, passes,
            "check={check} release={release:?} windows={windows:?}: {}",
            ran.stderr
        );
    }
}

/// The build job runs after a merge as before and on a pull request whose head branch begins
/// with `release/`: its one condition is today's condition with that case added.
#[test]
fn the_build_job_runs_on_a_release_pull_request_and_as_before_after_a_merge() {
    let ci = repo_file(".github/workflows/ci.yml");
    let job = job_block(&ci, "windows-build");
    // The job's own condition sits at the job's indentation; its steps have conditions too.
    let conditions: Vec<&str> = job
        .iter()
        .copied()
        .filter(|l| l.starts_with("    if: "))
        .map(str::trim)
        .collect();
    assert_eq!(
        conditions,
        ["if: ${{ github.event_name != 'pull_request' || startsWith(github.head_ref, 'release/') }}"],
        "{job:#?}"
    );
}

/// The head branch name is the pull request author's text: it reaches the scripts of the `ci`
/// workflow through `env` only, never through an expression inside a script or a condition.
#[test]
fn the_head_branch_name_reaches_scripts_through_env_only() {
    let ci = repo_file(".github/workflows/ci.yml");
    let elsewhere: Vec<&str> = ci
        .lines()
        .filter(|l| l.contains("github.head_ref"))
        .map(|l| l.trim())
        .filter(|l| !(l.starts_with("if: ") || l.starts_with("HEAD_REF: ")))
        .collect();
    assert!(elsewhere.is_empty(), "{elsewhere:#?}");
    let decide = job_block(&ci, "targets");
    let run_at = decide
        .iter()
        .position(|l| l.trim() == "run: |")
        .expect("the targets job has a script");
    assert!(
        decide[..run_at]
            .iter()
            .any(|l| l.trim() == "HEAD_REF: ${{ github.head_ref }}"),
        "{decide:#?}"
    );
}
