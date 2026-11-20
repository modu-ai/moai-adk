//! SPEC-20261120-test-weakening-detection: the weakened-test detector reads a unified diff and
//! finds five kinds; each kind has a diff that fires and a pair that stays quiet. The diffs are
//! built here from lines, and the marker words are joined at run time so no repository scan
//! reads this file's strings as an ignored or sleeping test of its own.

use moai_cli::gate::test_weakening::{detect, is_golden_path, is_test_path, Finding};
use moai_cli::gate::{
    effective_tier, tier_of, tier_with_weakening, Snapshot, TEST_WEAKENING_ENFORCED,
};
use serde_json::json;

fn ignore_attr() -> String {
    ["#[", "ignore]"].concat()
}

fn sleep_call(unit: &str) -> String {
    match unit {
        "rust" => format!(
            "    std::thread::{}(std::time::Duration::from_secs(1));",
            "sleep"
        ),
        _ => format!("    time.{}(1)", "sleep"),
    }
}

/// A hunk of one file: `old`/`new` are the removed and added lines; `at` is the first line.
fn change(path: &str, at: u64, old: &[String], new: &[String]) -> String {
    let (os, ns) = (
        if old.is_empty() { at - 1 } else { at },
        if new.is_empty() { at - 1 } else { at },
    );
    let mut s = format!(
        "diff --git a/{path} b/{path}\nindex 1111111..2222222 100644\n--- a/{path}\n+++ b/{path}\n@@ -{os},{} +{ns},{} @@\n",
        old.len(),
        new.len()
    );
    for l in old {
        s.push_str(&format!("-{l}\n"));
    }
    for l in new {
        s.push_str(&format!("+{l}\n"));
    }
    s
}

/// A file the diff creates.
fn created(path: &str, lines: &[String]) -> String {
    let mut s = format!(
        "diff --git a/{path} b/{path}\nnew file mode 100644\nindex 0000000..2222222\n--- /dev/null\n+++ b/{path}\n@@ -0,0 +1,{} @@\n",
        lines.len()
    );
    for l in lines {
        s.push_str(&format!("+{l}\n"));
    }
    s
}

fn v(lines: &[&str]) -> Vec<String> {
    lines.iter().map(|s| s.to_string()).collect()
}

fn kinds(findings: &[Finding]) -> Vec<&'static str> {
    findings.iter().map(|f| f.kind).collect()
}

fn of_kind<'a>(findings: &'a [Finding], kind: &str) -> Vec<&'a Finding> {
    findings.iter().filter(|f| f.kind == kind).collect()
}

/// AC-1 .. AC-5: for each kind a diff that fires, with its file and line, and a pair that does not.
#[test]
fn each_kind_fires_on_its_fixture_and_stays_quiet_on_its_pair() {
    // ---- test_deleted (AC-1)
    let rust_test = v(&["#[test]", "fn a_thing_works() {", "    run();", "}"]);
    let d = change("tests/thing.rs", 10, &rust_test, &[]);
    let f = detect(&d);
    let del = of_kind(&f, "test_deleted");
    assert_eq!(del.len(), 1, "{f:?}");
    assert_eq!((del[0].path.as_str(), del[0].line), ("tests/thing.rs", 10));
    assert!(del[0].detail.contains("a_thing_works"), "{:?}", del[0]);

    let py = v(&["def test_adds():", "    assert add(1, 2) == 3", ""]);
    let f = detect(&change("tests/test_math.py", 4, &py, &[]));
    assert_eq!(of_kind(&f, "test_deleted").len(), 1, "{f:?}");
    assert!(of_kind(&f, "test_deleted")[0].detail.contains("test_adds"));

    // Moved: the same test name starts again in another file of the diff.
    let moved = format!(
        "{}{}",
        change("tests/thing.rs", 10, &rust_test, &[]),
        created("tests/more.rs", &rust_test)
    );
    assert!(of_kind(&detect(&moved), "test_deleted").is_empty());
    // A test attribute swapped for another one is no deletion.
    let swapped = change(
        "tests/thing.rs",
        9,
        &v(&["#[test]"]),
        &v(&["#[tokio::test]"]),
    );
    assert!(of_kind(&detect(&swapped), "test_deleted").is_empty());
    // A bare attribute removed (the function stays) disables the test.
    let bare = change("tests/thing.rs", 9, &v(&["#[test]"]), &[]);
    assert_eq!(of_kind(&detect(&bare), "test_deleted").len(), 1);

    // ---- skip_added (AC-2): the five notations
    let skips = [
        ("tests/a.rs", ignore_attr()),
        ("tests/test_a.py", "@unittest.skip(\"later\")".to_string()),
        (
            "tests/test_b.py",
            "@pytest.mark.skip(reason=\"x\")".to_string(),
        ),
        ("tests/a_test.go", "\tt.Skip(\"later\")".to_string()),
        (
            "web/a.spec.ts",
            "  it.skip(\"works\", () => {});".to_string(),
        ),
    ];
    for (path, line) in &skips {
        let f = detect(&change(path, 7, &[], std::slice::from_ref(line)));
        let hit = of_kind(&f, "skip_added");
        assert_eq!(hit.len(), 1, "{path}: {f:?}");
        assert_eq!((hit[0].path.as_str(), hit[0].line), (*path, 7));
    }
    // Quiet pairs: the skip is deleted, only written about, or an iterator method.
    assert!(of_kind(
        &detect(&change("tests/a.rs", 7, &[ignore_attr()], &[])),
        "skip_added"
    )
    .is_empty());
    let prose = change(
        "docs/testing.md",
        3,
        &[],
        &[format!("Mark a slow test with `{}`.", ignore_attr())],
    );
    assert!(detect(&prose).is_empty(), "{:?}", detect(&prose));
    let iter = change(
        "src/lib.rs",
        3,
        &[],
        &v(&["    let rest = items.iter().skip(1);"]),
    );
    assert!(of_kind(&detect(&iter), "skip_added").is_empty());
    // A skip that only moves inside its file is not new.
    let moved_skip = change("tests/a.rs", 7, &[ignore_attr()], &[ignore_attr()]);
    assert!(of_kind(&detect(&moved_skip), "skip_added").is_empty());

    // ---- assertions_reduced (AC-3)
    let two_gone = change(
        "tests/x.rs",
        5,
        &v(&["    assert_eq!(a, 1);", "    assert!(b);"]),
        &v(&["    assert!(b);"]),
    );
    let f = detect(&two_gone);
    let hit = of_kind(&f, "assertions_reduced");
    assert_eq!(hit.len(), 1, "{f:?}");
    assert_eq!((hit[0].path.as_str(), hit[0].line), ("tests/x.rs", 5));
    assert!(
        hit[0].detail.contains("2 assertions removed, 1 added"),
        "{:?}",
        hit[0]
    );
    let moved = change(
        "tests/x.rs",
        5,
        &v(&["    assert_eq!(a, 1);"]),
        &v(&["    assert_eq!(a, 2);"]),
    );
    assert!(of_kind(&detect(&moved), "assertions_reduced").is_empty());
    let more = change(
        "tests/x.rs",
        5,
        &v(&["    assert!(b);"]),
        &v(&["    assert!(b);", "    assert!(c);"]),
    );
    assert!(of_kind(&detect(&more), "assertions_reduced").is_empty());
    let not_a_test = change("src/x.rs", 5, &v(&["    assert!(b);"]), &[]);
    assert!(of_kind(&detect(&not_a_test), "assertions_reduced").is_empty());
    let py = change(
        "tests/test_y.py",
        3,
        &v(&["    self.assertEqual(a, 1)", "    assert b"]),
        &[],
    );
    assert_eq!(of_kind(&detect(&py), "assertions_reduced").len(), 1);

    // ---- sleep_added (AC-4)
    let f = detect(&change("tests/x.rs", 12, &[], &[sleep_call("rust")]));
    assert_eq!(kinds(&f), ["sleep_added"], "{f:?}");
    assert_eq!(f[0].line, 12);
    let f = detect(&change("tests/test_z.py", 8, &[], &[sleep_call("py")]));
    assert_eq!(kinds(&f), ["sleep_added"], "{f:?}");
    assert!(detect(&change("src/x.rs", 12, &[], &[sleep_call("rust")])).is_empty());
    let commented = change(
        "tests/x.rs",
        12,
        &[],
        &[format!("    // std::thread::{}(...) was removed", "sleep")],
    );
    assert!(detect(&commented).is_empty());

    // ---- golden_only (AC-5)
    let golden = change("tests/golden/out.txt", 1, &v(&["old"]), &v(&["new"]));
    let f = detect(&golden);
    assert_eq!(kinds(&f), ["golden_only"], "{f:?}");
    assert_eq!((f[0].path.as_str(), f[0].line), ("tests/golden/out.txt", 1));
    let with_docs = format!("{golden}{}", change("docs/n.md", 2, &v(&["a"]), &v(&["b"])));
    assert_eq!(kinds(&detect(&with_docs)), ["golden_only"]);
    // Source code changes with the golden file: the expectation follows the code.
    let with_code = format!(
        "{golden}{}",
        change("src/main.rs", 4, &v(&["a"]), &v(&["b"]))
    );
    assert!(detect(&with_code).is_empty());
    // A golden file the diff creates is a new expectation, not a rewritten one.
    assert!(detect(&created("tests/golden/new.txt", &v(&["x"]))).is_empty());
    for p in [
        "a/golden/x.txt",
        "x.snap",
        "x.expected",
        "x.golden",
        "tests/fixtures/y.json",
    ] {
        assert!(is_golden_path(p), "{p}");
    }
    assert!(!is_golden_path("src/lib.rs"));
    for p in [
        "tests/a.rs",
        "crates/c/tests/b.rs",
        "pkg/test_x.py",
        "pkg/x_test.py",
        "pkg/x_test.go",
        "src/x_test.rs",
    ] {
        assert!(is_test_path(p), "{p}");
    }
    assert!(!is_test_path("src/lib.rs") && !is_test_path("src/attests.rs"));

    // Rendered form: `kind path:line detail`.
    let r = detect(&golden)[0].render();
    assert!(r.starts_with("golden_only tests/golden/out.txt:1 "), "{r}");
}

/// AC-6: a test-only change that weakens nothing is an empty list and keeps tier 0.
#[test]
fn a_test_only_change_that_weakens_nothing_stays_tier_zero() {
    let diff = format!(
        "{}{}",
        created(
            "tests/new_case.rs",
            &v(&[
                "#[test]",
                "fn it_holds() {",
                "    assert_eq!(1 + 1, 2);",
                "}"
            ]),
        ),
        change(
            "tests/old_case.rs",
            8,
            &v(&["    assert_eq!(f(1), 1);"]),
            &v(&["    assert_eq!(f(1), 1);", "    assert_eq!(f(2), 2);"]),
        )
    );
    assert!(detect(&diff).is_empty(), "{:?}", detect(&diff));
    let files = vec![
        "tests/new_case.rs".to_string(),
        "tests/old_case.rs".to_string(),
    ];
    assert_eq!(tier_of(Some(0), &files), 0);
    // An empty diff is an empty list too.
    assert!(detect("").is_empty());
}

fn snapshot(findings: Option<Vec<&str>>) -> Snapshot {
    let mut v = json!({
        "schema_version": 1,
        "pr": 9,
        "head": "1111111111111111111111111111111111111111",
        "comments": [],
        "checks": [],
        "pr_files": ["tests/a.rs"],
        "tip_files_since_verdict_base": [],
        "merge_clean": true,
        "body_check_ok": true,
        "body_tier": 0
    });
    if let Some(list) = findings {
        v["test_weakening"] = json!(list);
    }
    serde_json::from_value(v).expect("a snapshot")
}

/// AC-9: in observe the findings change no tier; enforced, a list that is not empty, and a diff
/// that was not read, make it at least 1.
#[test]
fn observe_reports_and_enforce_raises_the_tier() {
    // Pinned: turning the gate on is a decision, so it takes a change to this test as well.
    const { assert!(!TEST_WEAKENING_ENFORCED) };
    let hit = snapshot(Some(vec!["skip_added tests/a.rs:3 skips a test"]));
    let clean = snapshot(Some(vec![]));
    let unread = snapshot(None);
    for s in [&hit, &clean, &unread] {
        assert_eq!(effective_tier(s), 0, "observe ignores the findings");
    }
    let list = ["skip_added tests/a.rs:3 skips a test".to_string()];
    assert_eq!(tier_with_weakening(0, Some(&list), false), 0);
    assert_eq!(tier_with_weakening(0, Some(&list), true), 1);
    assert_eq!(tier_with_weakening(0, Some(&[]), true), 0);
    assert_eq!(tier_with_weakening(0, None, true), 1);
    assert_eq!(tier_with_weakening(2, Some(&list), true), 2);
    assert_eq!(tier_with_weakening(1, None, false), 1);
    // The report names the tier the findings would give.
    let r = moai_cli::gate::weakening_report(&hit);
    assert_eq!(r["enforced"], json!(false));
    assert_eq!(r["tier_with_findings"], json!(1));
    assert_eq!(
        r["findings"],
        json!(["skip_added tests/a.rs:3 skips a test"])
    );
    assert_eq!(
        moai_cli::gate::weakening_report(&clean)["tier_with_findings"],
        json!(0)
    );
    assert_eq!(
        moai_cli::gate::weakening_report(&unread)["findings"],
        json!("unknown")
    );
    assert_eq!(
        moai_cli::gate::weakening_report(&unread)["tier_with_findings"],
        json!(1)
    );
}
