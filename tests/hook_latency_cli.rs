//! SPEC-20261123-hook-entry-skeleton: `moai hook latency`, the Rust tool that measures how long
//! `moai hook <name>` takes. It reports p50, p90 and the maximum per case and judges nothing:
//! no number here is asserted, only the shape of the answer, the order of the percentiles and
//! the nearest-rank arithmetic on literal values.
#![cfg(unix)]
mod common;

use serde_json::Value;
use std::fs;
use std::path::Path;
use std::process::{Command, Stdio};

struct Run {
    code: i32,
    stdout: String,
    stderr: String,
}

/// Runs `moai hook latency` with `TMPDIR` set to `tmp`, so the measuring folders land where the
/// test says and never in the shared temporary folder.
fn latency_in(tmp: &Path, args: &[&str]) -> Run {
    let home = tempfile::TempDir::new().unwrap();
    let cwd = tempfile::TempDir::new().unwrap();
    let out = Command::new(common::MOAI_BIN)
        .arg("hook")
        .arg("latency")
        .args(args)
        .current_dir(cwd.path())
        .env("HOME", home.path())
        .env("TMPDIR", tmp)
        .env_remove("CLAUDE_PROJECT_DIR")
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .stdin(Stdio::null())
        .output()
        .unwrap();
    Run {
        code: out.status.code().unwrap(),
        stdout: String::from_utf8_lossy(&out.stdout).into_owned(),
        stderr: String::from_utf8_lossy(&out.stderr).into_owned(),
    }
}

/// Like [`latency_in`] with a temporary folder of its own, which must be empty again afterwards:
/// a run, refused or finished, leaves no measuring folder behind.
fn latency(args: &[&str]) -> Run {
    let tmp = tempfile::TempDir::new().unwrap();
    let run = latency_in(tmp.path(), args);
    let left: Vec<_> = fs::read_dir(tmp.path())
        .unwrap()
        .flatten()
        .map(|e| e.file_name())
        .collect();
    assert!(left.is_empty(), "{args:?} left {left:?} in TMPDIR");
    run
}

fn number(v: &Value, key: &str) -> f64 {
    v[key].as_f64().unwrap_or_else(|| panic!("{key} in {v}"))
}

/// AC-8: both cases are answered with the sample count and ordered percentiles; bad arguments
/// are refused before any sample is taken.
#[test]
fn ac8_both_cases_are_measured_with_ordered_percentiles() {
    let r = latency(&["--bin", common::MOAI_BIN, "--samples", "3"]);
    assert_eq!(r.code, 0, "{}", r.stderr);
    let v: Value = serde_json::from_str(r.stdout.trim()).expect(&r.stdout);
    assert_eq!(v["name"], "session-context");
    assert_eq!(v["samples"], 3);
    assert!(v["profile"].is_string(), "{v}");
    let cases = v["cases"].as_array().expect("cases");
    let names: Vec<&str> = cases.iter().map(|c| c["case"].as_str().unwrap()).collect();
    assert_eq!(names, ["not_moai_project", "moai_project"]);
    for c in cases {
        assert_eq!(c["samples"], 3, "{c}");
        let (p50, p90, max) = (
            number(c, "p50_ms"),
            number(c, "p90_ms"),
            number(c, "max_ms"),
        );
        assert!(0.0 <= p50 && p50 <= p90 && p90 <= max, "{c}");
    }
}

#[test]
fn ac8_bad_arguments_are_refused_with_exit_2_before_any_sample() {
    let bad: [&[&str]; 6] = [
        &["--bin", "/no/such/moai", "--samples", "3"],
        &["--bin", common::MOAI_BIN, "--samples", "0"],
        &["--bin", common::MOAI_BIN, "--samples", "1001"],
        &["--bin", common::MOAI_BIN, "--samples", "x"],
        &["--bin", common::MOAI_BIN, "--name", "pre-tool"],
        &["--samples", "3"],
    ];
    for args in bad {
        let r = latency(args);
        assert_eq!(r.code, 2, "{args:?}: {}", r.stderr);
        assert_eq!(r.stdout, "", "{args:?}");
        assert!(!r.stderr.trim().is_empty(), "{args:?}");
    }
}

/// AC-8: both ends of `--samples` (1 and 1000) are accepted, 0 and 1001 are refused with the
/// range message. The 1000 end runs against a stand-in binary that counts its starts and fails:
/// the run is accepted, then stops at the first sample, so exactly one process starts.
#[test]
fn ac8_the_samples_range_takes_both_ends_and_refuses_one_beyond() {
    let one = latency(&["--bin", common::MOAI_BIN, "--samples", "1"]);
    assert_eq!(one.code, 0, "{}", one.stderr);
    let v: Value = serde_json::from_str(one.stdout.trim()).expect(&one.stdout);
    assert_eq!(v["samples"], 1);
    for c in v["cases"].as_array().expect("cases") {
        assert_eq!(c["samples"], 1, "{c}");
        assert_eq!(c["p50_ms"], c["max_ms"], "{c}");
    }

    let dir = tempfile::TempDir::new().unwrap();
    let starts = dir.path().join("starts");
    let stand_in = dir.path().join("stand-in");
    common::write_executable(
        &stand_in,
        &format!("#!/bin/sh\necho start >> '{}'\nexit 1\n", starts.display()),
    );
    let stand_in = stand_in.to_str().unwrap();
    for samples in ["0", "1001"] {
        let r = latency(&["--bin", stand_in, "--samples", samples]);
        assert_eq!(r.code, 2, "{samples}: {}", r.stderr);
        assert!(
            r.stderr.contains("--samples must be 1 to 1000"),
            "{samples}: {}",
            r.stderr
        );
        assert!(!starts.exists(), "{samples}: no sample may start");
    }
    let top = latency(&["--bin", stand_in, "--samples", "1000"]);
    assert_eq!(top.code, 2, "{}", top.stderr);
    assert!(
        !top.stderr.contains("--samples must be"),
        "1000 is inside the range: {}",
        top.stderr
    );
    assert!(
        top.stderr.contains("not 0"),
        "the stand-in's exit is the failure: {}",
        top.stderr
    );
    assert_eq!(
        fs::read_to_string(&starts).unwrap().lines().count(),
        1,
        "the first sample failed, so one process started and the other 1999 never did"
    );
}

/// AC-8: the measuring folders are made under `TMPDIR`, only the folder this run made is
/// removed (a folder with the same name prefix that was already there stays), and an unusable
/// `TMPDIR` is one usage-style failure that creates and removes nothing.
#[test]
fn ac8_only_the_folder_the_run_made_is_removed() {
    let tmp = tempfile::TempDir::new().unwrap();
    let theirs = tmp.path().join("moai-hook-latency-1-1");
    fs::create_dir_all(theirs.join("inner")).unwrap();
    fs::write(theirs.join("inner/keep"), "mine").unwrap();
    let r = latency_in(tmp.path(), &["--bin", common::MOAI_BIN, "--samples", "1"]);
    assert_eq!(r.code, 0, "{}", r.stderr);
    let names: Vec<_> = fs::read_dir(tmp.path())
        .unwrap()
        .flatten()
        .map(|e| e.file_name().to_string_lossy().into_owned())
        .collect();
    assert_eq!(
        names,
        ["moai-hook-latency-1-1"],
        "only the pre-existing folder is left"
    );
    assert_eq!(
        fs::read_to_string(theirs.join("inner/keep")).unwrap(),
        "mine"
    );

    let missing = tmp.path().join("no-such-folder");
    let r = latency_in(&missing, &["--bin", common::MOAI_BIN, "--samples", "1"]);
    assert_eq!(r.code, 2, "{}", r.stderr);
    assert_eq!(r.stdout, "");
    assert!(
        r.stderr.contains("cannot prepare the measuring folders"),
        "{}",
        r.stderr
    );
    assert!(!missing.exists(), "the run must not make its parent");
    assert_eq!(
        fs::read_to_string(theirs.join("inner/keep")).unwrap(),
        "mine"
    );
}

/// AC-9: the nearest rank on literal values; nothing is started and nothing is judged.
#[test]
fn ac9_values_give_the_nearest_rank_and_judge_nothing() {
    let summary = |list: &str| -> Value {
        let r = latency(&["--values", list]);
        assert_eq!(r.code, 0, "{list}: {}", r.stderr);
        serde_json::from_str(r.stdout.trim()).expect(&r.stdout)
    };
    let ten = summary("10,1,9,2,8,3,7,4,6,5");
    assert_eq!(
        (
            number(&ten, "p50_ms"),
            number(&ten, "p90_ms"),
            number(&ten, "max_ms")
        ),
        (5.0, 9.0, 10.0)
    );
    assert_eq!(ten["samples"], 10);
    let twenty = summary(
        &(1..=20)
            .map(|n| n.to_string())
            .collect::<Vec<_>>()
            .join(","),
    );
    assert_eq!(number(&twenty, "p90_ms"), 18.0);
    assert_eq!(number(&twenty, "p50_ms"), 10.0);
    let one = summary("7.5");
    assert_eq!(
        (
            number(&one, "p50_ms"),
            number(&one, "p90_ms"),
            number(&one, "max_ms")
        ),
        (7.5, 7.5, 7.5)
    );
    for v in [&ten, &twenty, &one] {
        assert!(v.get("cases").is_none(), "no process is started: {v}");
        for verdict in [
            "ok", "pass", "passed", "fail", "verdict", "target", "within",
        ] {
            assert!(v.get(verdict).is_none(), "{verdict} in {v}");
        }
    }
    for bad in ["", "1,,2", "abc", "1,-2", "NaN", "1e999"] {
        let r = latency(&["--values", bad]);
        assert_eq!(r.code, 2, "{bad:?}: {}", r.stderr);
        assert_eq!(r.stdout, "", "{bad:?}");
    }
}

/// AC-11 (SPEC-20261123-hook-session-start): the `moai_project` case runs the card path of
/// `session-context`: the tool registers a ledger in its temporary HOME and reserves one card for
/// the measured project folder. A stand-in keeps what the built binary prints, with the folder it
/// ran in; every `moai_project` sample shows the card line and no `not_moai_project` sample does.
#[test]
fn ac11_the_project_case_runs_the_card_path() {
    let dir = tempfile::TempDir::new().unwrap();
    let log = dir.path().join("log");
    let stand_in = dir.path().join("stand-in");
    common::write_executable(
        &stand_in,
        &format!(
            "#!/bin/sh\necho \"cwd=$PWD\" >> '{log}'\nexec '{real}' \"$@\" >> '{log}'\n",
            log = log.display(),
            real = common::MOAI_BIN
        ),
    );
    let r = latency(&["--bin", stand_in.to_str().unwrap(), "--samples", "3"]);
    assert_eq!(r.code, 0, "{}", r.stderr);
    let text = fs::read_to_string(&log).unwrap();
    let mut folder = "";
    let (mut project, mut plain, mut cards) = (0, 0, 0);
    for line in text.lines() {
        if let Some(cwd) = line.strip_prefix("cwd=") {
            folder = if cwd.ends_with("/project") {
                project += 1;
                "project"
            } else {
                plain += 1;
                "plain"
            };
        } else {
            assert_eq!(folder, "project", "output outside a project: {line}");
            assert!(
                line.contains("\"hookEventName\":\"SessionStart\""),
                "{line}"
            );
            assert!(line.contains(" · card C-1 \\\""), "{line}");
            cards += 1;
        }
    }
    assert_eq!((project, plain, cards), (3, 3, 3), "{text}");
}
