//! SPEC-20261125-headver-weekly-release AC-3 and AC-4: the release workflow names a release
//! `release/v{head}.{yearweek}.{n}`. These tests take the plan script and the verify script out
//! of `.github/workflows/release.yml` and run them in temporary git repositories and folders
//! with a fixed clock, so the real HOME, the network and the real repository are never touched.
//! The week arithmetic of the plan script (plain integer math, because the date commands of
//! Linux and macOS differ) is checked against `headver::yearweek` over a table of days.

#![cfg(unix)]

use moai_cli::headver;
use std::{
    collections::HashMap,
    fs,
    path::{Path, PathBuf},
    process::Command,
    sync::LazyLock,
};
use tempfile::TempDir;

/// A file of this repository, read when the test runs (not `include_str!`, which the scan tables
/// would ask to list).
fn repo_text(relative: &str) -> String {
    fs::read_to_string(Path::new(env!("CARGO_MANIFEST_DIR")).join(relative))
        .unwrap_or_else(|e| panic!("{relative}: {e}"))
}

static WORKFLOW: LazyLock<String> = LazyLock::new(|| repo_text(".github/workflows/release.yml"));
static CI: LazyLock<String> = LazyLock::new(|| repo_text(".github/workflows/ci.yml"));
const NUMBER_STEP: &str = "Number the release";
const VERIFY_STEP: &str = "Binaries, tag and release note agree";

/// The `run: |` script of the step named `step`, without its indentation.
fn step_script(step: &str) -> String {
    let lines: Vec<&str> = WORKFLOW.lines().collect();
    let header = format!("- name: {step}");
    let start = lines
        .iter()
        .position(|l| l.trim() == header)
        .unwrap_or_else(|| panic!("no step {step:?}"));
    let run = (start..lines.len())
        .find(|&i| lines[i].trim() == "run: |")
        .unwrap_or_else(|| panic!("step {step:?} has no run block"));
    script_at(&lines, run)
}

/// The block that follows the `run: |` line at index `run`.
fn script_at(lines: &[&str], run: usize) -> String {
    let first = lines[run + 1];
    let indent = first.len() - first.trim_start().len();
    let mut out = String::new();
    for line in &lines[run + 1..] {
        if line.trim().is_empty() {
            out.push('\n');
            continue;
        }
        if line.len() - line.trim_start().len() < indent {
            break;
        }
        out.push_str(&line[indent..]);
        out.push('\n');
    }
    out
}

/// Every `run: |` script of the workflow.
fn all_scripts() -> Vec<String> {
    let lines: Vec<&str> = WORKFLOW.lines().collect();
    (0..lines.len())
        .filter(|&i| lines[i].trim() == "run: |")
        .map(|i| script_at(&lines, i))
        .collect()
}

fn git(dir: &Path, args: &[&str]) -> String {
    let out = Command::new("git")
        .args(["-c", "user.name=t", "-c", "user.email=t@example.invalid"])
        .args(["-c", "commit.gpgsign=false"])
        .args(args)
        .current_dir(dir)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .output()
        .unwrap();
    assert!(
        out.status.success(),
        "git {args:?}: {}",
        String::from_utf8_lossy(&out.stderr)
    );
    String::from_utf8(out.stdout).unwrap().trim().to_string()
}

struct Repo {
    _tmp: TempDir,
    dir: PathBuf,
    /// The one commit; the remote NeXT branch points at it.
    sha: String,
}

/// A repository with one commit holding `HEADVER` (when given) and a remote NeXT branch at it.
fn repo(head_text: Option<&str>) -> Repo {
    let tmp = TempDir::new().unwrap();
    let dir = fs::canonicalize(tmp.path()).unwrap().join("clone");
    fs::create_dir_all(&dir).unwrap();
    git(&dir, &["init", "-q", "-b", "work"]);
    fs::write(dir.join("file.txt"), "x\n").unwrap();
    if let Some(text) = head_text {
        fs::write(dir.join("HEADVER"), text).unwrap();
    }
    git(&dir, &["add", "-A"]);
    git(&dir, &["commit", "-q", "-m", "one"]);
    let sha = git(&dir, &["rev-parse", "HEAD"]);
    git(&dir, &["update-ref", "refs/remotes/origin/NeXT", &sha]);
    Repo {
        _tmp: tmp,
        dir,
        sha,
    }
}

/// Days since 1970-01-01 of a Gregorian date (written here, apart from the library's copy).
fn days(y: i64, m: i64, d: i64) -> i64 {
    let y = if m <= 2 { y - 1 } else { y };
    let era = y.div_euclid(400);
    let yoe = y - era * 400;
    let doy = (153 * (if m > 2 { m - 3 } else { m + 9 }) + 2) / 5 + d - 1;
    era * 146_097 + yoe * 365 + yoe / 4 - yoe / 100 + doy - 719_468
}

/// Seconds since the epoch of a UTC date at noon.
fn noon(y: i64, m: i64, d: i64) -> i64 {
    days(y, m, d) * 86_400 + 12 * 3600
}

/// The yearweek of the next Monday after `epoch`, by the library's ISO week code.
fn library_yearweek(epoch: i64) -> String {
    let day = epoch.div_euclid(86_400);
    let weekday = (day + 3).rem_euclid(7) + 1;
    let (y, m, d) = headver::civil_from_days(day + 8 - weekday);
    format!("{:04}", headver::yearweek(y, m, d))
}

struct Run {
    ok: bool,
    stderr: String,
    outputs: HashMap<String, String>,
}

/// Runs the plan script as GitHub does (`bash -eo pipefail`), in `repo`, with a fixed clock.
fn plan(repo: &Repo, branch: &str, pushed: &str, epoch: i64) -> Run {
    plan_with(repo, branch, pushed, Some(epoch), &[])
}

/// `epoch` of `None` leaves `SOURCE_DATE_EPOCH` unset, so the script reads the clock itself.
fn plan_with(
    repo: &Repo,
    branch: &str,
    pushed: &str,
    epoch: Option<i64>,
    extra: &[(&str, &str)],
) -> Run {
    let work = repo.dir.parent().unwrap();
    let script = work.join("plan.sh");
    fs::write(&script, step_script(NUMBER_STEP)).unwrap();
    let out_file = work.join("github_output");
    let _ = fs::remove_file(&out_file);
    fs::write(&out_file, "").unwrap();
    let mut cmd = Command::new("bash");
    cmd.args(["-eo", "pipefail"])
        .arg(&script)
        .current_dir(&repo.dir)
        .env_clear()
        .env("PATH", std::env::var("PATH").unwrap())
        .env("HOME", work)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .env("REF_NAME", branch)
        .env("PUSHED", pushed)
        .env("GITHUB_OUTPUT", &out_file);
    if let Some(epoch) = epoch {
        cmd.env("SOURCE_DATE_EPOCH", epoch.to_string());
    }
    for (k, v) in extra {
        cmd.env(k, v);
    }
    let out = cmd.output().unwrap();
    let outputs = fs::read_to_string(&out_file)
        .unwrap()
        .lines()
        .filter_map(|l| {
            l.split_once('=')
                .map(|(k, v)| (k.to_string(), v.to_string()))
        })
        .collect();
    Run {
        ok: out.status.success(),
        stderr: String::from_utf8_lossy(&out.stderr).into_owned(),
        outputs,
    }
}

/// REQ-5: the first release of a head and week is serial 0, and the week is the next Monday's.
#[test]
fn a_push_names_the_next_mondays_week_with_serial_zero() {
    let r = repo(Some("4\n"));
    // Wednesday 2026-11-25 is ISO week 2648; the next Monday, 2026-11-30, opens week 2649.
    let run = plan(&r, "release/v4.2649.0", &r.sha, noon(2026, 11, 25));
    assert!(run.ok, "{}", run.stderr);
    for (key, want) in [
        ("version", "4.2649.0"),
        ("tag", "moai-adk-v4.2649.0"),
        ("head", "4"),
        ("yearweek", "2649"),
        ("serial", "0"),
        ("prerelease", "false"),
    ] {
        assert_eq!(run.outputs[key], want, "{key}: {:?}", run.outputs);
    }
    assert_eq!(run.outputs["sha"], r.sha);
    assert_eq!(run.outputs["epoch"], noon(2026, 11, 25).to_string());
    // A release candidate is the same version on a `release/rc/` branch, marked a pre-release.
    let rc = plan(&r, "release/rc/v4.2649.0", &r.sha, noon(2026, 11, 25));
    assert!(rc.ok, "{}", rc.stderr);
    assert_eq!(rc.outputs["version"], "4.2649.0");
    assert_eq!(rc.outputs["tag"], "moai-adk-v4.2649.0");
    assert_eq!(rc.outputs["prerelease"], "true");
    // The head comes from HEADVER: a repository at head 12 plans 12.2649.0.
    let twelve = repo(Some("12\n"));
    let run = plan(
        &twelve,
        "release/v12.2649.0",
        &twelve.sha,
        noon(2026, 11, 25),
    );
    assert!(run.ok, "{}", run.stderr);
    assert_eq!(run.outputs["version"], "12.2649.0");
}

/// REQ-5 and the year boundary of the SPEC: the week comes from the Monday's ISO week-year.
/// Literals: Monday 2026-12-28 opens 2653, Monday 2027-01-04 opens 2701.
#[test]
fn the_release_week_follows_the_iso_week_year_at_a_year_boundary() {
    let r = repo(Some("4\n"));
    for (y, m, d, week) in [
        (2026, 12, 24, "2653"), // Thursday: the next Monday is 2026-12-28
        (2026, 12, 27, "2653"), // Sunday: the next Monday is 2026-12-28
        (2026, 12, 28, "2701"), // Monday: the next Monday is 2027-01-04, never the same day
        (2026, 12, 30, "2701"),
        (2027, 1, 3, "2701"), // Sunday of week 2653
        (2027, 1, 4, "2702"),
        (2025, 12, 22, "2601"), // the next Monday, 2025-12-29, is ISO week 1 of 2026
        (2025, 12, 29, "2602"),
        (2026, 11, 29, "2649"), // Sunday
        (2026, 11, 30, "2650"), // Monday
    ] {
        let name = format!("release/v4.{week}.0");
        let run = plan(&r, &name, &r.sha, noon(y, m, d));
        assert!(run.ok, "{y}-{m}-{d} {name}: {}", run.stderr);
        assert_eq!(run.outputs["yearweek"], week, "{y}-{m}-{d}");
        // The week before and after are refused: the name must spell the computed week.
        let wrong: u32 = week.parse::<u32>().unwrap() + 1;
        let run = plan(&r, &format!("release/v4.{wrong}.0"), &r.sha, noon(y, m, d));
        assert!(!run.ok, "{y}-{m}-{d} accepted week {wrong}");
        assert!(
            run.stderr.contains(&format!("4.{week}.0")),
            "{}",
            run.stderr
        );
    }
}

/// AC-3: the script's integer arithmetic and the library's ISO week code agree on the days around
/// every year boundary of a long span (53-week years included) and on a sample of ordinary days.
#[test]
fn the_script_agrees_with_headver_yearweek_over_a_table_of_days() {
    assert_eq!(days(2026, 11, 17) * 86_400, 1_794_873_600);
    assert_eq!(days(2027, 1, 4) * 86_400, 1_799_020_800);
    let r = repo(Some("4\n"));
    let mut table: Vec<(i64, i64, i64)> = Vec::new();
    for (y, m, d, span) in [
        (2020, 12, 24, 20),
        (2025, 12, 20, 20),
        (2026, 12, 20, 25),
        (2027, 12, 25, 15),
        (2032, 12, 25, 15),
        (2099, 12, 24, 20),
    ] {
        for step in 0..span {
            let day = days(y, m, d) + step;
            let (yy, mm, dd) = headver::civil_from_days(day);
            table.push((yy, i64::from(mm), i64::from(dd)));
        }
    }
    for month in 1..=12 {
        for day in [1, 15, 28] {
            table.push((2026, month, day));
            table.push((2028, month, day));
        }
    }
    for (y, m, d) in table {
        let epoch = noon(y, m, d);
        let want = library_yearweek(epoch);
        let run = plan(&r, &format!("release/v4.{want}.0"), &r.sha, epoch);
        assert!(run.ok, "{y}-{m}-{d} want {want}: {}", run.stderr);
        assert_eq!(run.outputs["yearweek"], want, "{y}-{m}-{d}");
    }
}

/// REQ-5: the serial is the lowest one no tag of the same head and week uses; tags of another
/// head or week do not count, and a name that spells a used serial is refused.
#[test]
fn the_serial_is_the_lowest_one_no_tag_uses() {
    let r = repo(Some("4\n"));
    let clock = noon(2026, 11, 25);
    for tag in [
        "moai-adk-v4.2648.0",
        "moai-adk-v5.2649.0",
        "moai-adk-v4.2650.3",
    ] {
        git(&r.dir, &["tag", tag]);
    }
    // None of those is of head 4 and week 2649, so the serial is still 0.
    assert!(plan(&r, "release/v4.2649.0", &r.sha, clock).ok);
    git(&r.dir, &["tag", "moai-adk-v4.2649.0"]);
    let used = plan(&r, "release/v4.2649.0", &r.sha, clock);
    assert!(!used.ok, "a serial a tag uses is never reused");
    assert!(used.stderr.contains("4.2649.1"), "{}", used.stderr);
    let next = plan(&r, "release/v4.2649.1", &r.sha, clock);
    assert!(next.ok, "{}", next.stderr);
    assert_eq!(next.outputs["serial"], "1");
    assert_eq!(next.outputs["tag"], "moai-adk-v4.2649.1");
    // A gap does not matter: the next serial is one above the highest.
    git(&r.dir, &["tag", "moai-adk-v4.2649.5"]);
    let six = plan(&r, "release/v4.2649.6", &r.sha, clock);
    assert!(six.ok, "{}", six.stderr);
    assert!(!plan(&r, "release/v4.2649.1", &r.sha, clock).ok);
    assert!(!plan(&r, "release/v4.2649.2", &r.sha, clock).ok);
}

/// REQ-5: tags of another family, or of another shape of name that merely contains the prefix,
/// never block the plan.
#[test]
fn tags_of_other_families_do_not_block_the_plan() {
    let r = repo(Some("4\n"));
    for tag in [
        "moai-adk-v4.2640.6",
        "v3.1.3",
        "go-v1.2.3",
        "moai-cli-v2026.11.0-rc.1",
        "xmoai-adk-v4.2649.50",
    ] {
        git(&r.dir, &["tag", tag]);
    }
    let run = plan(&r, "release/v4.2649.0", &r.sha, noon(2026, 11, 25));
    assert!(run.ok, "{}", run.stderr);
    assert_eq!(run.outputs["tag"], "moai-adk-v4.2649.0");
}

/// Without `SOURCE_DATE_EPOCH` the script reads the clock itself, once, and passes that reading
/// on as `epoch`. The week is computed from the reading taken before the run; a run that crosses
/// a Monday midnight is skipped so the test never depends on the wall clock's luck.
#[test]
fn without_a_given_epoch_the_clock_is_read_once() {
    let r = repo(Some("4\n"));
    let now = || {
        std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap()
            .as_secs() as i64
    };
    let before = now();
    let week = library_yearweek(before);
    let run = plan_with(&r, &format!("release/v4.{week}.0"), &r.sha, None, &[]);
    let after = now();
    if library_yearweek(after) != week {
        return;
    }
    assert!(run.ok, "{}", run.stderr);
    let epoch: i64 = run.outputs["epoch"].parse().unwrap();
    assert!(
        before <= epoch && epoch <= after,
        "{epoch} not in {before}..{after}"
    );
    assert_eq!(run.outputs["yearweek"], week);
}

/// REQ-5: a tag of the same head and week with a shape the plan does not read stops it.
#[test]
fn a_malformed_tag_of_the_same_week_stops_the_plan() {
    for bad in [
        "moai-adk-v4.2649.01",
        "moai-adk-v4.2649.x",
        "moai-adk-v4.2649.4294967296",
    ] {
        let r = repo(Some("4\n"));
        git(&r.dir, &["tag", bad]);
        let run = plan(&r, "release/v4.2649.0", &r.sha, noon(2026, 11, 25));
        assert!(!run.ok, "{bad} was skipped");
        assert!(run.stderr.contains(bad), "{bad}: {}", run.stderr);
    }
}

/// REQ-6: a head, week or serial that is not the computed one is refused, and the refusal names
/// the version the release should have.
#[test]
fn a_name_that_spells_another_version_is_refused() {
    let r = repo(Some("4\n"));
    let clock = noon(2026, 11, 25);
    for name in [
        "release/v5.2649.0", // head is not the HEADVER head
        "release/v3.2649.0",
        "release/v4.2648.0",    // this week, not the next Monday's
        "release/v4.2650.0",    // two weeks on
        "release/v4.2649.1",    // a serial no tag precedes
        "release/rc/v4.2650.0", // the rc shape is held to the same rule
    ] {
        let run = plan(&r, name, &r.sha, clock);
        assert!(!run.ok, "{name} was accepted");
        assert!(
            run.stderr
                .contains("the version of this release is 4.2649.0"),
            "{name}: {}",
            run.stderr
        );
        assert!(
            run.outputs.is_empty(),
            "{name} wrote outputs: {:?}",
            run.outputs
        );
    }
}

/// REQ-6: the earlier `release/<40 hex>` shape and every other shape is refused.
#[test]
fn the_earlier_and_other_shapes_are_refused() {
    let r = repo(Some("4\n"));
    let clock = noon(2026, 11, 25);
    let hex = r.sha.clone();
    let mut names = vec![
        format!("release/{hex}"),
        format!("release/rc/{hex}"),
        format!("release/v{hex}"),
    ];
    for name in [
        "release/v4.2649",
        "release/v4.2649.0.1",
        "release/v4.649.0",
        "release/v4.26490.0",
        "release/v04.2649.0",
        "release/v10000.2649.0",
        "release/v4.2649.00",
        "release/v4.2649.01",
        "release/v4.2649.-1",
        "release/v4.2649.0-dev",
        "release/v4.2649.0+meta",
        "release/V4.2649.0",
        "release/4.2649.0",
        "release/rc/4.2649.0",
        "release/rc/rc/v4.2649.0",
        "release/x/v4.2649.0",
        "release/v4.2649.0/x",
        "release//v4.2649.0",
        "release/v4.2649.4294967296",
        "release/v4.2649.99999999999",
        "release/",
        "release",
        "main",
        "",
    ] {
        names.push(name.to_string());
    }
    for name in names {
        let run = plan(&r, &name, &r.sha, clock);
        assert!(!run.ok, "{name:?} was accepted");
        assert!(run.outputs.is_empty(), "{name:?} wrote outputs");
    }
}

/// REQ-7: the pushed commit must be on the remote NeXT branch, and the branch must be there.
#[test]
fn a_commit_outside_the_remote_next_branch_is_refused() {
    let r = repo(Some("4\n"));
    let clock = noon(2026, 11, 25);
    let first = r.sha.clone();
    fs::write(r.dir.join("more.txt"), "y\n").unwrap();
    git(&r.dir, &["add", "-A"]);
    git(&r.dir, &["commit", "-q", "-m", "two"]);
    let second = git(&r.dir, &["rev-parse", "HEAD"]);
    // NeXT is at the first commit: the second is not on it.
    let outside = plan(&r, "release/v4.2649.0", &second, clock);
    assert!(!outside.ok);
    assert!(
        outside.stderr.contains("is not on the remote NeXT branch"),
        "{}",
        outside.stderr
    );
    // The first commit is on NeXT.
    assert!(plan(&r, "release/v4.2649.0", &first, clock).ok);
    // Once NeXT holds the second commit, both are on it.
    git(&r.dir, &["update-ref", "refs/remotes/origin/NeXT", &second]);
    assert!(plan(&r, "release/v4.2649.0", &second, clock).ok);
    assert!(plan(&r, "release/v4.2649.0", &first, clock).ok);
    // No NeXT branch in the clone at all.
    git(&r.dir, &["update-ref", "-d", "refs/remotes/origin/NeXT"]);
    let missing = plan(&r, "release/v4.2649.0", &second, clock);
    assert!(!missing.ok);
    assert!(
        missing.stderr.contains("remote NeXT branch"),
        "{}",
        missing.stderr
    );
    // The branch name of the clone's local `work` branch does not stand in for it.
    let local = plan_with(
        &r,
        "release/v4.2649.0",
        &second,
        Some(clock),
        &[("NEXT_REF", "work")],
    );
    assert!(local.ok, "NEXT_REF is the test's seam: {}", local.stderr);
}

/// A pushed value that is not a full lowercase commit id is refused.
#[test]
fn a_pushed_commit_that_is_not_a_full_commit_id_is_refused() {
    let r = repo(Some("4\n"));
    let clock = noon(2026, 11, 25);
    let upper = r.sha.to_uppercase();
    for pushed in [&r.sha[..12], upper.as_str(), "", "HEAD"] {
        let run = plan(&r, "release/v4.2649.0", pushed, clock);
        assert!(!run.ok, "{pushed:?} was accepted");
    }
}

/// REQ-5: the head is the one line of HEADVER, 0 to 9999, no sign, no leading zero.
#[test]
fn the_head_file_must_hold_one_head() {
    let clock = noon(2026, 11, 25);
    for (text, name, ok) in [
        (Some("0\n"), "release/v0.2649.0", true),
        (Some("9999\r\n"), "release/v9999.2649.0", true),
        (Some("4"), "release/v4.2649.0", true),
        (Some("04\n"), "release/v4.2649.0", false),
        (Some("10000\n"), "release/v10000.2649.0", false),
        (Some("+4\n"), "release/v4.2649.0", false),
        (Some("4\n5\n"), "release/v4.2649.0", false),
        (Some(""), "release/v4.2649.0", false),
        (None, "release/v4.2649.0", false),
    ] {
        let r = repo(text);
        let run = plan(&r, name, &r.sha, clock);
        assert_eq!(run.ok, ok, "{text:?} {name}: {}", run.stderr);
    }
}

/// Fake `moai-cli` and `moai` that print the given two `--version` lines.
fn fake_binaries(dir: &Path, first: &str, second: &str) {
    let bin = dir.join("target/release");
    fs::create_dir_all(&bin).unwrap();
    for name in ["moai-cli", "moai"] {
        let path = bin.join(name);
        fs::write(
            &path,
            format!("#!/bin/sh\nprintf '%s\\n' '{first}'\nprintf '%s\\n' '{second}'\n"),
        )
        .unwrap();
        use std::os::unix::fs::PermissionsExt;
        fs::set_permissions(&path, fs::Permissions::from_mode(0o755)).unwrap();
    }
}

struct Verify {
    ok: bool,
    stderr: String,
    summary: String,
}

fn verify(
    first: &str,
    second: &str,
    version: &str,
    serial: &str,
    tag: &str,
    branch: &str,
) -> Verify {
    verify_with(first, second, version, serial, tag, branch, "false")
}

fn verify_with(
    first: &str,
    second: &str,
    version: &str,
    serial: &str,
    tag: &str,
    branch: &str,
    prerelease: &str,
) -> Verify {
    let tmp = TempDir::new().unwrap();
    let dir = fs::canonicalize(tmp.path()).unwrap();
    fake_binaries(&dir, first, second);
    let script = dir.join("verify.sh");
    fs::write(&script, step_script(VERIFY_STEP)).unwrap();
    let summary = dir.join("summary.md");
    fs::write(&summary, "").unwrap();
    let out = Command::new("bash")
        .args(["-eo", "pipefail"])
        .arg(&script)
        .current_dir(&dir)
        .env_clear()
        .env("PATH", std::env::var("PATH").unwrap())
        .env("VERSION", version)
        .env("SERIAL", serial)
        .env("TAG", tag)
        .env("SHA", "a".repeat(40))
        .env("PRERELEASE", prerelease)
        .env("REF_NAME", branch)
        .env("GITHUB_STEP_SUMMARY", &summary)
        .output()
        .unwrap();
    Verify {
        ok: out.status.success(),
        stderr: String::from_utf8_lossy(&out.stderr).into_owned(),
        summary: fs::read_to_string(&summary).unwrap(),
    }
}

const RELEASE_LINE: &str = "release build 0 (MOAI_RELEASE_VERSION), built 2026-11-25 UTC (ISO week 2026-W48), package 2026.11.0-rc.1";

/// REQ-8 and REQ-9: the build job writes the release note only when both binaries and the tag
/// spell the version.
#[test]
fn the_build_job_checks_both_lines_and_the_tag() {
    let good = verify(
        "moai-cli 4.2649.0 (rust)",
        RELEASE_LINE,
        "4.2649.0",
        "0",
        "moai-adk-v4.2649.0",
        "release/v4.2649.0",
    );
    assert!(good.ok, "{}", good.stderr);
    for text in [
        "## Release note: moai-adk-v4.2649.0",
        "- version: 4.2649.0",
        "- serial: 0",
        "- pre-release: no",
    ] {
        assert!(good.summary.contains(text), "{text}: {}", good.summary);
    }
    // The `moai` binary may say `moai` and the release candidate branch is accepted.
    let rc = verify_with(
        "moai 4.2649.0 (rust)",
        RELEASE_LINE,
        "4.2649.0",
        "0",
        "moai-adk-v4.2649.0",
        "release/rc/v4.2649.0",
        "true",
    );
    assert!(rc.ok, "{}", rc.stderr);
    // A candidate's note says it is a pre-release.
    assert!(rc.summary.contains("- pre-release: yes"), "{}", rc.summary);

    let bad: Vec<(&str, &str, &str, &str, &str, &str)> = vec![
        // The first line carries another version, a suffix, or the wrong generation mark.
        (
            "moai-cli 4.2649.1 (rust)",
            RELEASE_LINE,
            "4.2649.0",
            "0",
            "moai-adk-v4.2649.0",
            "release/v4.2649.0",
        ),
        (
            "moai-cli 4.2649.0-dev (rust)",
            RELEASE_LINE,
            "4.2649.0",
            "0",
            "moai-adk-v4.2649.0",
            "release/v4.2649.0",
        ),
        (
            "moai-cli 4.2649.0 (go)",
            RELEASE_LINE,
            "4.2649.0",
            "0",
            "moai-adk-v4.2649.0",
            "release/v4.2649.0",
        ),
        // The second line says another kind of build or another serial.
        (
            "moai-cli 4.2649.0 (rust)",
            "local build 5 (MOAI_LOCAL_BUILD), built 2026-11-25 UTC",
            "4.2649.0",
            "0",
            "moai-adk-v4.2649.0",
            "release/v4.2649.0",
        ),
        (
            "moai-cli 4.2649.0 (rust)",
            "CI build 0 (MOAI_BUILD_NUMBER), built 2026-11-25 UTC",
            "4.2649.0",
            "0",
            "moai-adk-v4.2649.0",
            "release/v4.2649.0",
        ),
        (
            "moai-cli 4.2649.0 (rust)",
            "unnumbered build (no MOAI_LOCAL_BUILD or MOAI_BUILD_NUMBER), built 2026-11-25 UTC",
            "4.2649.0",
            "0",
            "moai-adk-v4.2649.0",
            "release/v4.2649.0",
        ),
        (
            "moai-cli 4.2649.0 (rust)",
            "release build 1 (MOAI_RELEASE_VERSION), built 2026-11-25 UTC",
            "4.2649.0",
            "0",
            "moai-adk-v4.2649.0",
            "release/v4.2649.0",
        ),
        (
            "moai-cli 4.2649.0 (rust)",
            "",
            "4.2649.0",
            "0",
            "moai-adk-v4.2649.0",
            "release/v4.2649.0",
        ),
        // The tag, the serial or the branch does not spell the version.
        (
            "moai-cli 4.2649.0 (rust)",
            RELEASE_LINE,
            "4.2649.0",
            "0",
            "moai-adk-v4.2649.1",
            "release/v4.2649.0",
        ),
        (
            "moai-cli 4.2649.0 (rust)",
            RELEASE_LINE,
            "4.2649.0",
            "0",
            "moai-cli-v4.2649.0",
            "release/v4.2649.0",
        ),
        (
            "moai-cli 4.2649.0 (rust)",
            RELEASE_LINE,
            "4.2649.0",
            "1",
            "moai-adk-v4.2649.0",
            "release/v4.2649.0",
        ),
        (
            "moai-cli 4.2649.0 (rust)",
            RELEASE_LINE,
            "4.2649.0",
            "0",
            "moai-adk-v4.2649.0",
            "release/v4.2649.1",
        ),
        (
            "moai-cli 4.2649.0 (rust)",
            RELEASE_LINE,
            "4.2649.0",
            "0",
            "moai-adk-v4.2649.0",
            "release/4.2649.0",
        ),
    ];
    for (first, second, version, serial, tag, branch) in bad {
        let run = verify(first, second, version, serial, tag, branch);
        assert!(
            !run.ok,
            "{first:?} / {second:?} / {tag} / {serial} / {branch} passed"
        );
        assert!(
            !run.summary.contains("Release note"),
            "a failing check wrote the note: {}",
            run.summary
        );
    }
    // A missing binary fails.
    let tmp = TempDir::new().unwrap();
    let dir = fs::canonicalize(tmp.path()).unwrap();
    let script = dir.join("verify.sh");
    fs::write(&script, step_script(VERIFY_STEP)).unwrap();
    let out = Command::new("bash")
        .args(["-eo", "pipefail"])
        .arg(&script)
        .current_dir(&dir)
        .env_clear()
        .env("PATH", std::env::var("PATH").unwrap())
        .env("VERSION", "4.2649.0")
        .env("SERIAL", "0")
        .env("TAG", "moai-adk-v4.2649.0")
        .env("SHA", "a".repeat(40))
        .env("PRERELEASE", "false")
        .env("REF_NAME", "release/v4.2649.0")
        .env("GITHUB_STEP_SUMMARY", dir.join("summary.md"))
        .output()
        .unwrap();
    assert!(!out.status.success());
}

/// REQ-10: the workflow starts from a push to `release/**` only, holds a read-only token and no
/// secret, builds with the release version and no other build number, and writes nothing out.
#[test]
fn the_workflow_publishes_nothing() {
    let before_jobs = WORKFLOW.split("\njobs:").next().unwrap();
    let triggers = before_jobs
        .split("\non:\n")
        .nth(1)
        .unwrap()
        .split("\npermissions:")
        .next()
        .unwrap();
    assert_eq!(
        triggers.trim_end(),
        "  push:\n    branches: ['release/**']",
        "only a push to a release branch starts it"
    );
    // The workflow without its comment lines: the rules below read code, not prose.
    let code: String = WORKFLOW
        .lines()
        .filter(|l| !l.trim_start().starts_with('#'))
        .collect::<Vec<_>>()
        .join("\n");
    for text in [
        "workflow_dispatch",
        "pull_request",
        "schedule:",
        "concurrency",
    ] {
        assert!(!code.contains(text), "{text}");
    }
    // Every permissions block is `contents: read` and nothing else.
    let mut blocks = 0;
    let lines: Vec<&str> = WORKFLOW.lines().collect();
    for (i, line) in lines.iter().enumerate() {
        if line.trim() == "permissions:" {
            blocks += 1;
            assert_eq!(lines[i + 1].trim(), "contents: read", "line {}", i + 1);
        }
    }
    assert_eq!(blocks, 3, "the workflow and its two jobs");
    assert!(!code.contains("write"), "no write permission or step");
    for script in all_scripts() {
        for word in [
            "gh release",
            "gh api",
            "gh pr",
            "git tag",
            "git push",
            "git remote",
            "curl ",
            "upload-artifact",
            "attest",
            "secrets.",
            "id-token",
            "GITHUB_TOKEN",
            "${{",
        ] {
            assert!(!script.contains(word), "a script holds {word:?}:\n{script}");
        }
    }
    // The build is a release build: the plan's version, and no other number source.
    let build_env = WORKFLOW
        .split("  build:\n")
        .nth(1)
        .unwrap()
        .split("    steps:")
        .next()
        .unwrap();
    assert!(
        build_env.contains("MOAI_RELEASE_VERSION: ${{ needs.plan.outputs.version }}"),
        "{build_env}"
    );
    for var in ["MOAI_LOCAL_BUILD", "MOAI_BUILD_NUMBER"] {
        assert!(
            !code
                .lines()
                .any(|l| l.trim_start().starts_with(&format!("{var}:"))),
            "{var} is set"
        );
    }
    // Every action is pinned to a full commit id, the one ci.yml uses.
    let mut uses = 0;
    for line in WORKFLOW
        .lines()
        .filter(|l| l.trim_start().starts_with("- uses:"))
    {
        let spec = line.trim_start().trim_start_matches("- uses:").trim();
        let (action, sha) = spec.split(' ').next().unwrap().split_once('@').unwrap();
        assert!(
            sha.len() == 40 && sha.bytes().all(|b| b.is_ascii_hexdigit()),
            "{spec}"
        );
        assert!(
            CI.contains(&format!("{action}@{sha}")),
            "ci.yml does not pin {spec}"
        );
        uses += 1;
    }
    assert!(uses >= 2, "{uses}");
}

/// The settings of the two jobs the plan and the build rely on: no stored credential, the tags
/// fetched for the plan, the build on the commit the plan checked, the plan's clock reading, bash
/// as the shell, and no secret, token scope or publishing action anywhere in the file.
#[test]
fn the_jobs_keep_their_settings() {
    let code: String = WORKFLOW
        .lines()
        .filter(|l| !l.trim_start().starts_with('#'))
        .collect::<Vec<_>>()
        .join("\n");
    for word in [
        "secrets.",
        "secrets:",
        "GITHUB_TOKEN",
        "id-token",
        "upload-artifact",
        "attest",
        "gh release",
        "gh api",
        "git push",
    ] {
        assert!(!code.contains(word), "the workflow holds {word:?}");
    }
    let (plan_job, build_job) = code.split_once("\n  build:\n").expect("a build job");
    assert!(plan_job.contains("\n  plan:\n"), "a plan job");
    // Every checkout keeps no credential; the plan reads every tag of the clone it checks out.
    assert_eq!(code.matches("uses: actions/checkout@").count(), 2);
    assert_eq!(code.matches("persist-credentials: false").count(), 2);
    assert!(
        plan_job.contains("fetch-tags: true"),
        "the plan reads the tags"
    );
    assert!(
        plan_job.contains("fetch-depth: 0"),
        "the plan reads the history"
    );
    // The build waits for the plan, builds the commit the plan checked and takes its clock reading.
    assert!(build_job.contains("needs: plan"));
    assert!(build_job.contains("ref: ${{ needs.plan.outputs.sha }}"));
    assert!(build_job.contains("SOURCE_DATE_EPOCH: ${{ needs.plan.outputs.epoch }}"));
    // Both jobs run their scripts under bash.
    assert_eq!(code.matches("shell: bash").count(), 2);
    assert!(plan_job.contains("shell: bash") && build_job.contains("shell: bash"));
}
