//! SPEC-20261126-check-affected-trailers-flaky (AC-1, AC-2): `check affected` applies the trailer
//! rules of `moai trailer-check` to its range as a cheap `trailers` step. A failing finding puts
//! `trailers=1` and `stopped_at=trailers` into the one summary line and skips the heavy steps; a
//! range without a failing finding keeps the summary it always had. Temporary git repositories on a
//! `spec/` branch, a temporary HOME and a fake `cargo` that exits 0 and records its calls: the real
//! tools and the machine's heavy slots are never used.
#![cfg(unix)]
mod common;

use common::git::{git, git_ok};
use serde_json::Value;
use std::fs;
use std::os::unix::fs::PermissionsExt;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};

const BIN: &str = crate::common::MOAI_BIN;
const ROOMY: &str = "free_gb=1000,swap_pct=0";

/// A `spec/` branch: a missing `Moai-Spec` fails the trailer rules only there.
const BRANCH: &str = "spec/demo-20261126";
const SPEC: &str = "SPEC-20261126-demo";
const SESSION: &str = "abc12345:123456789012";
const ATTRIBUTION: &str = "Co-Authored-By: Claude Sonnet 5.5 <noreply@anthropic.com>";

/// A fake `cargo` that records every call and succeeds, so a step that runs shows in the call log.
const FAKE_CARGO: &str = "#!/bin/sh\necho \"cargo $*\" >> \"$FAKE_CALLS\"\nexit 0\n";

fn put(root: &Path, rel: &str, text: &str) {
    let path = root.join(rel);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, text).unwrap();
}

/// A repository on the `spec/` branch with a base commit; the range commits come from `commit`.
struct Fixture {
    dir: tempfile::TempDir,
    base: String,
}

impl Fixture {
    fn new() -> Fixture {
        let dir = tempfile::tempdir().unwrap();
        let p = dir.path();
        git_ok(p, &["init", "-q", "-b", BRANCH]);
        git_ok(p, &["config", "user.email", "t@example.invalid"]);
        git_ok(p, &["config", "user.name", "t"]);
        put(p, "README.md", "x\n");
        put(p, "Cargo.toml", "[package]\nname = \"x\"\n");
        put(p, "src/update_models.rs", "// a\n");
        put(p, "src/bundle_pack.rs", "// a\n");
        put(p, "tests/update_command_cli.rs", "fn main() {}\n");
        put(p, "tests/bundle_run_cli.rs", "fn main() {}\n");
        git_ok(p, &["add", "-A"]);
        git_ok(p, &["commit", "-q", "-m", "base"]);
        let base = git(p, &["rev-parse", "HEAD"]);
        Fixture { dir, base }
    }

    fn root(&self) -> &Path {
        self.dir.path()
    }

    /// One more commit with exactly this message; it changes `src/update_models.rs`.
    fn commit(&self, message: &str) {
        let file = self.root().join("src/update_models.rs");
        let mut text = fs::read_to_string(&file).unwrap();
        text.push_str("// next\n");
        fs::write(&file, text).unwrap();
        git_ok(self.root(), &["add", "-A"]);
        git_ok(self.root(), &["commit", "-q", "-m", message]);
    }

    fn head(&self) -> String {
        git(self.root(), &["rev-parse", "HEAD"])
    }

    fn range(&self) -> String {
        format!("{}..{}", self.base, self.head())
    }
}

/// The fake tool folder, the slot folder, the parent of the log folder and a temporary HOME.
struct Run {
    tools: tempfile::TempDir,
    slots: tempfile::TempDir,
    outs: tempfile::TempDir,
    home: tempfile::TempDir,
}

impl Run {
    fn new() -> Run {
        let tools = tempfile::tempdir().unwrap();
        for (n, body) in [("cargo", FAKE_CARGO), ("rustfmt", "#!/bin/sh\nexit 0\n")] {
            let path = tools.path().join(n);
            fs::write(&path, body).unwrap();
            fs::set_permissions(&path, fs::Permissions::from_mode(0o755)).unwrap();
        }
        let slots = tempfile::tempdir().unwrap();
        fs::write(
            slots.path().join("limits.json"),
            r#"{"max":4,"min_free_gb":0,"max_swap_pct":100,"min_mem_free_pct":0}"#,
        )
        .unwrap();
        Run {
            tools,
            slots,
            outs: tempfile::tempdir().unwrap(),
            home: tempfile::tempdir().unwrap(),
        }
    }

    fn calls(&self) -> PathBuf {
        self.tools.path().join("calls.log")
    }

    /// The cargo calls so far, oldest first.
    fn cargo_calls(&self) -> Vec<String> {
        fs::read_to_string(self.calls())
            .unwrap_or_default()
            .lines()
            .map(String::from)
            .collect()
    }

    fn out(&self) -> PathBuf {
        self.outs.path().join("check")
    }

    /// `moai check affected <range> --dir … --out …` with the fakes on PATH; the answer's result.
    fn go(&self, fx: &Fixture) -> Value {
        let git_dir = Command::new("sh")
            .args(["-c", "dirname \"$(command -v git)\""])
            .output()
            .unwrap();
        let path = format!(
            "{}:{}:/usr/bin:/bin",
            self.tools.path().display(),
            String::from_utf8_lossy(&git_dir.stdout).trim()
        );
        let out = Command::new(BIN)
            .args(["check", "affected", &fx.range(), "--dir"])
            .arg(fx.root())
            .arg("--slot-dir")
            .arg(self.slots.path())
            .arg("--out")
            .arg(self.out())
            .env("HOME", self.home.path())
            .env_remove("MOAI_HOME")
            .env_remove("XDG_CACHE_HOME")
            .env_remove("CARGO_TARGET_DIR")
            .env_remove("MOAI_HEAVY_SLOT_FD")
            .env_remove("RUST_TEST_THREADS")
            .env_remove("NEXTEST_TEST_THREADS")
            .env("MOAI_TEST_MACHINE", ROOMY)
            .env("PATH", path)
            .env("FAKE_CALLS", self.calls())
            .stdin(Stdio::null())
            .output()
            .unwrap();
        let text = String::from_utf8_lossy(&out.stdout);
        let at = text.find('{').unwrap_or(0);
        let v: Value = serde_json::from_str(&text[at..])
            .unwrap_or_else(|_| panic!("{text} / {}", String::from_utf8_lossy(&out.stderr)));
        assert_eq!(v["ok"], true, "{v}");
        v["result"].clone()
    }
}

fn summary(v: &Value) -> String {
    v["summary"].as_str().unwrap().to_string()
}

/// The receipt words a trailer failure gives: `trailers=1`, the stop at that step, `FAIL`, and no
/// heavy step started (no word for it, listed as skipped, no cargo call for it).
fn assert_stopped_at_trailers(r: &Run, v: &Value) {
    let line = summary(v);
    assert!(line.contains(" trailers=1"), "{line}");
    assert!(line.contains(" stopped_at=trailers"), "{line}");
    assert!(line.ends_with(" result=FAIL"), "{line}");
    assert_eq!(v["stopped_at"], "trailers", "{v}");
    for heavy in [" clippy=", " test="] {
        assert!(!line.contains(heavy), "{heavy}: {line}");
    }
    let skipped: Vec<&str> = v["skipped"]
        .as_array()
        .unwrap()
        .iter()
        .map(|s| s.as_str().unwrap())
        .collect();
    assert!(
        skipped.contains(&"clippy") && skipped.contains(&"test"),
        "{v}"
    );
    let calls = r.cargo_calls();
    assert!(
        !calls
            .iter()
            .any(|c| c.starts_with("cargo clippy") || c.starts_with("cargo nextest run")),
        "{calls:?}"
    );
    // The one-line receipt on disk says the same.
    let text = fs::read_to_string(r.out().join("summary.txt")).unwrap();
    assert_eq!(text.trim_end(), line, "{text}");
}

/// AC-1 (REQ-1, REQ-2): the `Moai-Spec` line alone in a paragraph before the last one is never
/// read, so the range fails on a `spec/` branch and the receipt says where it stopped.
#[test]
fn a_split_trailer_paragraph_fails_the_receipt_with_trailers_1() {
    let fx = Fixture::new();
    fx.commit(&format!(
        "fix: demo\n\nMoai-Spec: {SPEC}\n\nMoai-Session: {SESSION}\nMoai-Verified: none\n{ATTRIBUTION}"
    ));
    let r = Run::new();
    let v = r.go(&fx);
    assert_stopped_at_trailers(&r, &v);
}

/// AC-1 (REQ-1, REQ-2, item I): a host attribution line between the `Moai-*` lines makes git read
/// no trailer from the last paragraph, so the range fails the same way.
#[test]
fn an_attribution_line_in_the_trailer_paragraph_fails_the_receipt_with_trailers_1() {
    let fx = Fixture::new();
    fx.commit(&format!(
        "fix: demo\n\nMoai-Spec: {SPEC}\n\u{1f5ff} MoAI\nMoai-Session: {SESSION}\nMoai-Verified: none\n{ATTRIBUTION}"
    ));
    let r = Run::new();
    let v = r.go(&fx);
    assert_stopped_at_trailers(&r, &v);
}

/// AC-2 (REQ-3): a range whose trailers are clean, and one whose only finding does not fail (a
/// missing `Moai-Session`), keep the summary words they had before: no `trailers` word and no stop
/// at that step.
#[test]
fn a_clean_range_keeps_the_summary_without_a_trailers_word() {
    let fx = Fixture::new();
    fx.commit(&format!(
        "fix: demo\n\nMoai-Spec: {SPEC}\nMoai-Session: {SESSION}\nMoai-Verified: tests\n{ATTRIBUTION}"
    ));
    let r = Run::new();
    let v = r.go(&fx);
    let line = summary(&v);
    assert!(!line.contains("trailers"), "{line}");
    assert!(!line.contains("stopped_at=trailers"), "{line}");
    assert_eq!(v["stopped_at"], Value::Null, "{v}");
    assert!(line.ends_with(" clippy=0 test=0 result=PASS"), "{line}");

    // A finding that does not fail (no session line) adds no word either.
    fx.commit(&format!(
        "fix: demo again\n\nMoai-Spec: {SPEC}\nMoai-Verified: none\n{ATTRIBUTION}"
    ));
    let v = r.go(&fx);
    let line = summary(&v);
    assert!(!line.contains("trailers"), "{line}");
    assert!(line.ends_with(" clippy=0 test=0 result=PASS"), "{line}");
}
