//! SPEC-20261120-index-state-folders: when `go-state-plan` stops checking `.moai/` at its entry
//! limit, the reason of the `.moai` item names the largest folders under it and, when `reports`
//! is one of them, the command that moves the reports out of the project.
//!
//! The case behind it: a project whose `.moai/reports` held 74 GB left the preview with "the
//! folder stays" and nothing else. The entry limit is lowered with the debug-build variable
//! `MOAI_CLI_TEST_UNRECORDED_LIMIT`, so a few files reach it. Every case builds a temporary
//! project; the real checkout is never read.

#![cfg(unix)]

mod common;
use common::home::TempHome;
use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::{fs, path::PathBuf};
use tempfile::TempDir;

struct Fx {
    _tmp: TempDir,
    root: PathBuf,
}

impl Fx {
    /// A Go-generation project: a deploy manifest that records one unchanged template file,
    /// which is what makes `.moai/` movable under the default limit.
    fn new() -> Fx {
        let tmp = TempDir::new().unwrap();
        let root = fs::canonicalize(tmp.path()).unwrap().join("project");
        fs::create_dir_all(&root).unwrap();
        let fx = Fx { _tmp: tmp, root };
        let body = "a: 1\n";
        fx.write(".moai/config/sections/a.yaml", body);
        let hash = format!("sha256:{:x}", Sha256::digest(body.as_bytes()));
        fx.write(
            ".moai/manifest.json",
            &json!({"version": "v3.2.1-rc.6", "deployed_at": "2026-11-01T00:00:00Z",
                    "files": {".moai/config/sections/a.yaml": {
                        "provenance": "template_managed", "template_hash": hash,
                        "deployed_hash": hash, "current_hash": hash}}})
            .to_string(),
        );
        fx
    }

    fn write(&self, rel: &str, body: &str) {
        let p = self.root.join(rel);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        fs::write(p, body).unwrap();
    }

    /// `count` files directly in `.moai/<folder>/`.
    fn files(&self, folder: &str, count: usize) {
        for n in 0..count {
            self.write(&format!(".moai/{folder}/f-{n:03}"), "x");
        }
    }

    fn tree(&self) -> Vec<(PathBuf, Vec<u8>)> {
        let mut out = Vec::new();
        let mut stack = vec![self.root.clone()];
        while let Some(d) = stack.pop() {
            for e in fs::read_dir(&d).unwrap() {
                let p = e.unwrap().path();
                if fs::symlink_metadata(&p).unwrap().is_dir() {
                    stack.push(p);
                } else {
                    out.push((p.clone(), fs::read(&p).unwrap()));
                }
            }
        }
        out.sort();
        out
    }
}

/// The preview with the entry limit lowered to `limit` (or the default when `None`).
fn plan(fx: &Fx, limit: Option<usize>) -> Value {
    let before = fx.tree();
    // The binary runs through a temporary HOME with a clean environment, never the real one.
    let home = TempHome::new();
    let mut cmd = home.command(&["go-state-plan", fx.root.to_str().unwrap()]);
    cmd.env("PATH", "/usr/bin:/bin").env("LC_ALL", "C");
    if let Some(n) = limit {
        cmd.env("MOAI_CLI_TEST_UNRECORDED_LIMIT", n.to_string());
    }
    let out = cmd.output().unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
        panic!(
            "{}{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    });
    assert_eq!(v["ok"], json!(true), "{v}");
    assert_eq!(fx.tree(), before, "a preview writes nothing");
    v["result"].clone()
}

/// The `(action, reason)` of the `.moai` item.
fn moai_item(plan: &Value) -> (String, String) {
    let item = plan["items"]
        .as_array()
        .unwrap()
        .iter()
        .find(|i| i["path"] == ".moai")
        .unwrap_or_else(|| panic!("no .moai item: {plan}"));
    (
        item["action"].as_str().unwrap().to_string(),
        item["reason"].as_str().unwrap().to_string(),
    )
}

fn position(text: &str, needle: &str) -> usize {
    text.find(needle)
        .unwrap_or_else(|| panic!("`{needle}` is not in: {text}"))
}

/// AC-5 (REQ-5, REQ-7, REQ-8): the reason names the largest folders, largest first, and a
/// folder over the limit is shown as `more than <limit> entries`.
#[test]
fn the_blocked_reason_names_the_largest_folders() {
    let fx = Fx::new();
    fx.files("reports", 8);
    fx.files("logs", 3);
    fx.files("cache", 2);
    fx.files("memory", 1);
    let (action, reason) = moai_item(&plan(&fx, Some(5)));
    assert_eq!(action, "leave", "{reason}");
    assert!(reason.contains("was not fully checked"), "{reason}");
    assert!(reason.contains("Largest folders under .moai/:"), "{reason}");
    let reports = position(&reason, "reports (more than 5 entries)");
    let logs = position(&reason, "logs (3 entries)");
    // `config` holds `sections/` and `a.yaml` (2 entries), the same as `cache`: equal counts
    // are ordered by name.
    let cache = position(&reason, "cache (2 entries)");
    let config = position(&reason, "config (2 entries)");
    let memory = position(&reason, "memory (1 entry)");
    assert!(
        reports < logs && logs < cache && cache < config && config < memory,
        "largest first, equal counts by name: {reason}"
    );
}

/// AC-5 (REQ-5): at most five folders are named.
#[test]
fn at_most_five_folders_are_named() {
    let fx = Fx::new();
    for (n, folder) in ["f1", "f2", "f3", "f4", "f5", "f6"].iter().enumerate() {
        fx.files(folder, 10 - n);
    }
    let (_, reason) = moai_item(&plan(&fx, Some(3)));
    for named in ["f1", "f2", "f3", "f4", "f5"] {
        assert!(
            reason.contains(&format!("{named} (more than 3")),
            "{reason}"
        );
    }
    assert!(!reason.contains("f6 ("), "{reason}");
    assert!(!reason.contains("config ("), "{reason}");
}

/// AC-5 (REQ-8), the edge: a folder of exactly the limit shows its count, one more shows
/// `more than`.
#[test]
fn a_folder_at_the_limit_is_counted_and_one_over_is_more_than() {
    let at = Fx::new();
    at.files("reports", 5);
    let (_, reason) = moai_item(&plan(&at, Some(5)));
    assert!(reason.contains("reports (5 entries)"), "{reason}");

    let over = Fx::new();
    over.files("reports", 6);
    let (_, reason) = moai_item(&plan(&over, Some(5)));
    assert!(reason.contains("reports (more than 5 entries)"), "{reason}");
}

/// AC-6 (REQ-6): the move command and the folder it moves to are given when `reports` is among
/// the named folders, and are not when it is not.
#[test]
fn the_reason_gives_the_report_move_only_when_reports_is_named() {
    let with = Fx::new();
    with.files("reports", 8);
    let (_, reason) = moai_item(&plan(&with, Some(5)));
    assert!(reason.contains("`moai reports-move-plan .`"), "{reason}");
    assert!(reason.contains("~/.moai/reports/<project>/"), "{reason}");

    let without = Fx::new();
    without.files("logs", 8);
    let (_, reason) = moai_item(&plan(&without, Some(5)));
    assert!(reason.contains("logs (more than 5 entries)"), "{reason}");
    assert!(!reason.contains("reports-move-plan"), "{reason}");
    assert!(!reason.contains("~/.moai/reports"), "{reason}");
}

/// AC-5 (REQ-7): the blocked folder stays, the preview writes nothing (`plan` compares the
/// tree), and under the default limit the same project moves.
#[test]
fn the_blocked_folder_stays_and_the_preview_writes_nothing() {
    let fx = Fx::new();
    fx.files("reports", 8);
    let (action, _) = moai_item(&plan(&fx, Some(5)));
    assert_eq!(action, "leave");
    let (action, reason) = moai_item(&plan(&fx, None));
    assert_eq!(action, "move", "{reason}");
    assert!(!reason.contains("Largest folders"), "{reason}");
}
