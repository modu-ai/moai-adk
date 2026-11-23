//! SPEC-20261123-hook-entry-skeleton: `moai hook latency`, the Rust tool that measures how long
//! `moai hook <name>` takes. It reports p50, p90 and the maximum per case and judges nothing:
//! no number here is asserted, only the shape of the answer, the order of the percentiles and
//! the nearest-rank arithmetic on literal values.
#![cfg(unix)]
mod common;

use serde_json::Value;
use std::process::{Command, Stdio};

struct Run {
    code: i32,
    stdout: String,
    stderr: String,
}

fn latency(args: &[&str]) -> Run {
    let home = tempfile::TempDir::new().unwrap();
    let cwd = tempfile::TempDir::new().unwrap();
    let out = Command::new(common::MOAI_BIN)
        .arg("hook")
        .arg("latency")
        .args(args)
        .current_dir(cwd.path())
        .env("HOME", home.path())
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
