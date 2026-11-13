//! #2561: Go-generation project state — the read-only preview. The Go manifest
//! (`.moai/manifest.json`) is untrusted input: only a regular, unlinked file it records as
//! `template_managed` whose bytes still hash to its `deployed_hash` is proven tool-made and
//! offered for moving; everything else stays, with a reason. Every case builds a temporary
//! project; the real checkout is never read.

#![cfg(unix)]

mod common;
use common::CleanEnv;
use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::{
    fs,
    os::unix::fs::symlink,
    path::{Path, PathBuf},
    process::{Command, Output},
};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;

fn sha(bytes: &[u8]) -> String {
    format!("sha256:{:x}", Sha256::digest(bytes))
}

struct Fx {
    _tmp: TempDir,
    root: PathBuf,
    files: serde_json::Map<String, Value>,
}

impl Fx {
    fn new() -> Fx {
        let tmp = TempDir::new().unwrap();
        let root = fs::canonicalize(tmp.path()).unwrap().join("project");
        fs::create_dir_all(&root).unwrap();
        Fx {
            _tmp: tmp,
            root,
            files: serde_json::Map::new(),
        }
    }

    fn write(&self, rel: &str, body: &str) {
        let p = self.root.join(rel);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        fs::write(p, body).unwrap();
    }

    /// A file the Go deployer wrote and recorded.
    fn deployed(&mut self, rel: &str, body: &str) {
        self.write(rel, body);
        self.record(rel, "template_managed", &sha(body.as_bytes()));
    }

    fn record(&mut self, rel: &str, provenance: &str, deployed: &str) {
        self.files.insert(
            rel.into(),
            json!({"provenance": provenance, "template_hash": deployed,
                   "deployed_hash": deployed, "current_hash": deployed}),
        );
    }

    fn manifest(&self) {
        self.write(
            ".moai/manifest.json",
            &json!({"version": "v3.2.1-rc.6", "deployed_at": "2026-11-01T00:00:00Z",
                    "files": self.files})
            .to_string(),
        );
    }

    fn tree(&self) -> Vec<(PathBuf, Vec<u8>)> {
        let mut out = Vec::new();
        let mut stack = vec![self.root.clone()];
        while let Some(d) = stack.pop() {
            for e in fs::read_dir(&d).unwrap() {
                let p = e.unwrap().path();
                let m = fs::symlink_metadata(&p).unwrap();
                if m.is_dir() {
                    stack.push(p);
                } else if m.file_type().is_symlink() {
                    out.push((
                        p.clone(),
                        fs::read_link(&p)
                            .unwrap()
                            .into_os_string()
                            .into_encoded_bytes(),
                    ));
                } else {
                    out.push((p.clone(), fs::read(&p).unwrap()));
                }
            }
        }
        out.sort();
        out
    }
}

fn run(args: &[&str]) -> Output {
    Command::new(EXE)
        .args(args)
        .clean_env()
        .env("PATH", "/usr/bin:/bin")
        .env("LC_ALL", "C")
        .output()
        .unwrap()
}

fn plan(fx: &Fx) -> Value {
    let before = fx.tree();
    let out = run(&["go-state-plan", fx.root.to_str().unwrap()]);
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

fn action<'a>(plan: &'a Value, rel: &str) -> (&'a str, &'a str) {
    let item = plan["items"]
        .as_array()
        .unwrap()
        .iter()
        .find(|i| i["path"] == rel)
        .unwrap_or_else(|| panic!("{rel} not in plan: {plan}"));
    (
        item["action"].as_str().unwrap(),
        item["reason"].as_str().unwrap(),
    )
}

#[test]
fn only_unchanged_template_files_are_offered_for_moving() {
    let mut fx = Fx::new();
    fx.deployed(".moai/config/sections/a.yaml", "a: 1\n");
    fx.deployed(".claude/workflows/w.md", "# w\n");
    fx.deployed(".claude/rules/b.md", "b: 1\n");
    fx.write(".claude/rules/b.md", "b: 2 # edited\n");
    fx.write("AGENTS.md", "mine\n");
    fx.record("AGENTS.md", "user_created", &sha(b"mine\n"));
    fx.write(".claude/rules/u.md", "u: 1\n");
    fx.record(".claude/rules/u.md", "user_modified", &sha(b"u: 1\n"));
    fx.write(".moai/specs/SPEC-1/spec.md", "# spec\n");
    fx.manifest();
    let p = plan(&fx);
    assert_eq!(p["manifest_version"], json!("v3.2.1-rc.6"));
    // Decision 1: `.moai/` moves as one folder; nothing under it moves on its own.
    assert_eq!(
        action(&p, ".moai"),
        (
            "move",
            "the Go .moai/ folder as a whole (decision 1, #2561 comment 6831345102)"
        )
    );
    assert_eq!(
        action(&p, ".moai/config/sections/a.yaml"),
        ("with_folder", "moves with .moai/ as a whole")
    );
    assert_eq!(action(&p, ".claude/workflows/w.md").0, "move");
    assert_eq!(
        action(&p, ".claude/rules/b.md"),
        ("leave", "changed since the Go deploy")
    );
    assert_eq!(
        action(&p, ".claude/rules/u.md"),
        ("leave", "not template_managed")
    );
    assert_eq!(
        action(&p, "AGENTS.md"),
        ("leave", "outside .moai/ and .claude/")
    );
    // Not recorded by the Go manifest: never moved, counted under its folder.
    assert_eq!(p["unrecorded"][".moai/specs"], json!(1));
    assert!(p["plan_digest"].as_str().unwrap().starts_with("sha256:"));
}

#[test]
fn links_and_escaping_keys_are_never_moved() {
    let mut fx = Fx::new();
    fs::create_dir_all(fx.root.join("elsewhere")).unwrap();
    fs::write(fx.root.join("elsewhere/h.sh"), "#!/bin/sh\n").unwrap();
    fs::create_dir_all(fx.root.join(".claude")).unwrap();
    symlink(fx.root.join("elsewhere"), fx.root.join(".claude/hooks")).unwrap();
    fx.record(
        ".claude/hooks/h.sh",
        "template_managed",
        &sha(b"#!/bin/sh\n"),
    );
    fx.write(".claude/rules/real.md", "x\n");
    symlink(
        fx.root.join(".claude/rules/real.md"),
        fx.root.join(".claude/rules/link.md"),
    )
    .unwrap();
    fx.record(".claude/rules/link.md", "template_managed", &sha(b"x\n"));
    fx.record(".moai/../escape.txt", "template_managed", &sha(b""));
    fx.record("/etc/passwd", "template_managed", &sha(b""));
    fx.manifest();
    let p = plan(&fx);
    assert_eq!(
        action(&p, ".claude/hooks/h.sh"),
        ("leave", "a path component is a symlink")
    );
    assert_eq!(
        action(&p, ".claude/rules/link.md"),
        ("leave", "a path component is a symlink")
    );
    assert_eq!(
        action(&p, ".moai/../escape.txt"),
        ("leave", "not a plain relative path")
    );
    assert_eq!(
        action(&p, "/etc/passwd"),
        ("leave", "not a plain relative path")
    );
}

#[test]
fn files_settings_still_reference_stay() {
    let mut fx = Fx::new();
    fx.deployed(".claude/hooks/moai/handle-stop.sh", "#!/bin/sh\n");
    fx.deployed(
        ".claude/output-styles/moai/MoAI-Easy.md",
        "---\nname: MoAI-Easy\n---\n",
    );
    fx.deployed(".claude/workflows/free.md", "# free\n");
    fx.write(
        ".claude/settings.json",
        &json!({"hooks": {"Stop": [{"hooks": [{"type": "command",
            "command": "\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-stop.sh\""}]}]},
            "outputStyle": "MoAI-Easy"})
        .to_string(),
    );
    fx.manifest();
    let p = plan(&fx);
    assert_eq!(
        action(&p, ".claude/hooks/moai/handle-stop.sh"),
        ("leave", "referenced by .claude/settings.json")
    );
    assert_eq!(
        action(&p, ".claude/output-styles/moai/MoAI-Easy.md"),
        ("leave", "referenced by .claude/settings.json")
    );
    assert_eq!(action(&p, ".claude/workflows/free.md").0, "move");
}

/// #2670: the Go generation's `.worktreeinclude` lists `.env` and `.env.*`, so Claude Code
/// copies them into every new worktree. The preview names each such line and changes nothing;
/// a safe list or no file gives no warning.
#[test]
fn a_worktreeinclude_that_copies_secrets_is_reported_and_left_alone() {
    let has = |p: &Value, needle: &str| {
        p["warnings"]
            .as_array()
            .unwrap()
            .iter()
            .any(|w| w.as_str().unwrap().contains(needle))
    };
    let fx = Fx::new();
    fx.write(".env", "PLACEHOLDER=1\n");
    fx.write(
        ".worktreeinclude",
        "# Environment files\n.claude/settings.local.json\n.env\n.env.*\n",
    );
    let p = plan(&fx);
    assert!(
        has(&p, ".worktreeinclude line 3 `.env` copies secret files"),
        "{p}"
    );
    assert!(
        has(&p, ".worktreeinclude line 4 `.env.*` copies secret files"),
        "{p}"
    );
    assert!(!has(&p, "line 2"), "{p}");
    assert_eq!(
        fs::read_to_string(fx.root.join(".worktreeinclude")).unwrap(),
        "# Environment files\n.claude/settings.local.json\n.env\n.env.*\n",
        "never edited"
    );
    let safe = Fx::new();
    safe.write(".worktreeinclude", ".claude/settings.local.json\n.envrc\n");
    assert!(!has(&plan(&safe), ".worktreeinclude"));
    assert!(!has(&plan(&Fx::new()), ".worktreeinclude"));
}

#[test]
fn without_a_valid_manifest_nothing_moves() {
    for body in [
        None,
        Some("not json"),
        Some(r#"{"version":"v3","files":[]}"#),
    ] {
        let fx = Fx::new();
        fx.write(".moai/config/a.yaml", "a\n");
        if let Some(b) = body {
            fx.write(".moai/manifest.json", b);
        }
        let p = plan(&fx);
        assert_eq!(p["items"], json!([]), "{body:?}: {p}");
        assert!(
            p["warnings"]
                .as_array()
                .unwrap()
                .iter()
                .any(|w| w.as_str().unwrap().contains("no usable Go manifest")),
            "{p}"
        );
    }
}

#[test]
fn a_linked_moai_folder_or_manifest_is_refused() {
    let mut fx = Fx::new();
    fx.deployed(".moai/config/a.yaml", "a\n");
    fx.manifest();
    let real = fx.root.join("real-moai");
    fs::rename(fx.root.join(".moai"), &real).unwrap();
    symlink(&real, fx.root.join(".moai")).unwrap();
    let out = run(&["go-state-plan", fx.root.to_str().unwrap()]);
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], json!(false), "{v}");
    assert!(v.to_string().contains("symlink"), "{v}");

    // A linked .moai with no manifest in it is refused too, not read as "no manifest".
    let g = Fx::new();
    let real = g.root.join("real-moai");
    fs::create_dir_all(real.join("config")).unwrap();
    symlink(&real, g.root.join(".moai")).unwrap();
    let out = run(&["go-state-plan", g.root.to_str().unwrap()]);
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], json!(false), "{v}");
    let _ = Path::new("");
}

// ---- stage 2: apply and revert (#2561 AC-2, AC-3, AC-4) ----

const REF: &str = "test-permission-ref";

fn call(args: &[&str]) -> Value {
    let out = run(args);
    serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
        panic!(
            "{args:?}: {}{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    })
}

/// The project tree without the backup areas: AC-3 compares the project as the user sees
/// it; the backup folder, the set-aside `.moai.rust-<UTC>` (#2696) and the receipt stay
/// after a revert as the record of both acts.
fn project_tree(fx: &Fx) -> Vec<(PathBuf, Vec<u8>)> {
    let backups = fx.root.join(".moai-cli/backups");
    fx.tree()
        .into_iter()
        .filter(|(p, _)| !p.starts_with(&backups))
        .filter(|(p, _)| {
            !p.strip_prefix(&fx.root)
                .unwrap()
                .to_string_lossy()
                .starts_with(".moai.rust-")
        })
        .collect()
}

/// The folder a revert set this generation's `.moai/` aside as (#2696).
fn rust_aside(fx: &Fx) -> PathBuf {
    let mut found: Vec<PathBuf> = fs::read_dir(&fx.root)
        .unwrap()
        .map(|e| e.unwrap().path())
        .filter(|p| {
            p.file_name()
                .unwrap()
                .to_string_lossy()
                .starts_with(".moai.rust-")
        })
        .collect();
    assert_eq!(found.len(), 1, "{found:?}");
    found.pop().unwrap()
}

/// The `.moai/backups/<UTC>/` area of an apply that moved the Go folder (#2696).
fn area(fx: &Fx, backup: &str) -> PathBuf {
    fx.root.join(".moai/backups").join(backup)
}

fn movable_fixture() -> Fx {
    let mut fx = Fx::new();
    fx.deployed(".moai/config/sections/a.yaml", "a: 1\n");
    fx.deployed(".claude/workflows/w.md", "# w\n");
    fx.deployed(".moai/config/sections/b.yaml", "b: 1\n");
    fx.write(".moai/config/sections/b.yaml", "b: 2 # edited\n");
    fx.write(".moai/specs/SPEC-1/spec.md", "# spec\n");
    fx.manifest();
    fx
}

fn apply(fx: &Fx, digest: &str) -> Value {
    call(&[
        "go-state-apply",
        fx.root.to_str().unwrap(),
        "--plan-digest",
        digest,
        "--permission-reference",
        REF,
    ])
}

fn moves(plan: &Value) -> usize {
    plan["items"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|i| i["action"] == "move")
        .count()
}

/// #2725 (leader condition 6854854683): a `.moai/` that also holds the Rust marker is
/// `mixed` — the preview moves nothing, restores nothing and says why; the apply then has
/// nothing to move either.
#[test]
fn a_mixed_moai_folder_moves_nothing_and_says_why() {
    let fx = movable_fixture();
    assert!(moves(&plan(&fx)) > 0, "the fixture alone is movable");
    fx.write(
        ".moai/moai-cli.json",
        r#"{"schema_version":1,"generation":"moai-cli"}"#,
    );
    let p = plan(&fx);
    assert_eq!(moves(&p), 0, "{p}");
    assert_eq!(p["restores"], json!([]), "{p}");
    let warnings = p["warnings"].to_string();
    assert!(warnings.contains("state_root is `mixed`"), "{warnings}");
    assert!(
        p["items"]
            .as_array()
            .unwrap()
            .iter()
            .any(|i| i["reason"] == "state_root `mixed`: nothing moves (#2725)"),
        "{p}"
    );
    let before = project_tree(&fx);
    apply(&fx, p["plan_digest"].as_str().unwrap());
    assert_eq!(project_tree(&fx), before, "nothing moved");
}

/// AUDIT #2776 B1: a stopped migration (`interrupted`) holds the preview the same way — the
/// Go files would otherwise move (2 moves and a SPEC restore on this fixture). #2696: the
/// verdict now comes from a Go backup beside a `.moai` that holds a go-state receipt and no
/// valid marker.
#[test]
fn an_interrupted_migration_moves_nothing_and_says_why() {
    let fx = movable_fixture();
    fx.write(".moai-backup/manifest.json", "{}");
    fx.write(".moai/backups/20261113T000000Z/go-state-receipt.json", "{}");
    let p = plan(&fx);
    assert_eq!(moves(&p), 0, "{p}");
    assert_eq!(p["restores"], json!([]), "{p}");
    assert!(
        p["warnings"]
            .to_string()
            .contains("state_root is `interrupted`"),
        "{p}"
    );
    assert!(
        p["items"]
            .as_array()
            .unwrap()
            .iter()
            .any(|i| i["reason"] == "state_root `interrupted`: nothing moves (#2725)"),
        "{p}"
    );
    let before = project_tree(&fx);
    apply(&fx, p["plan_digest"].as_str().unwrap());
    assert_eq!(project_tree(&fx), before, "nothing moved");
}

#[test]
fn apply_moves_only_the_planned_files_and_needs_the_current_digest() {
    let fx = movable_fixture();
    let before = fx.tree();
    let p = plan(&fx);
    assert_eq!(moves(&p), 2, "{p}");

    // AC-2: without the confirmed digest nothing changes, not even the backup area.
    for wrong in ["", "sha256:0000", "not-a-digest"] {
        let v = apply(&fx, wrong);
        assert_eq!(v["ok"], json!(false), "{wrong}: {v}");
        assert_eq!(
            fx.tree(),
            before,
            "{wrong}: an unconfirmed apply writes nothing"
        );
    }

    assert_eq!(p["go_backup"], json!(".moai-backup"), "{p}");
    assert_eq!(
        p["left_in_backup"],
        json!(["config", "manifest.json", "specs"]),
        "{p}"
    );
    let v = apply(&fx, p["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], json!(true), "{v}");
    let r = &v["result"];
    let backup = r["backup"].as_str().unwrap().to_string();
    assert_eq!(r["state"], json!("committed"), "{r}");
    assert_eq!(r["permission_reference"], json!(REF));
    assert_eq!(r["go_backup"], json!(".moai-backup"), "{r}");
    let dir = area(&fx, &backup);
    // The proven `.claude/` file moved into the new `.moai/`'s backup area.
    assert!(!fx.root.join(".claude/workflows/w.md").exists());
    assert_eq!(
        fs::read_to_string(dir.join("go-state/.claude/workflows/w.md")).unwrap(),
        "# w\n"
    );
    // #2696: the whole Go folder is `.moai-backup` now, changed, unrecorded and manifest
    // files included, byte for byte.
    let go = fx.root.join(".moai-backup");
    for (rel, body) in [
        ("config/sections/a.yaml", "a: 1\n"),
        ("config/sections/b.yaml", "b: 2 # edited\n"),
        ("specs/SPEC-1/spec.md", "# spec\n"),
    ] {
        assert_eq!(fs::read_to_string(go.join(rel)).unwrap(), body, "{rel}");
    }
    assert!(go.join("manifest.json").is_file());
    // The new `.moai/` is this generation's: the marker is there and nothing of Go's.
    let marker: Value =
        serde_json::from_slice(&fs::read(fx.root.join(".moai/moai-cli.json")).unwrap()).unwrap();
    assert_eq!(marker["schema_version"], json!(1), "{marker}");
    assert_eq!(marker["generation"], json!("moai-cli"), "{marker}");
    assert_eq!(marker["go_backup"], json!(".moai-backup"), "{marker}");
    assert!(!fx.root.join(".moai/manifest.json").exists());
    assert!(!fx.root.join(".moai/config").exists());
    let receipt: Value =
        serde_json::from_slice(&fs::read(dir.join("go-state-receipt.json")).unwrap()).unwrap();
    assert_eq!(receipt["state"], json!("committed"), "{receipt}");
    assert_eq!(receipt["go_backup"], json!(".moai-backup"), "{receipt}");
    assert_eq!(
        receipt["marker"],
        json!(sha(&fs::read(fx.root.join(".moai/moai-cli.json")).unwrap()))
    );
    let folder = receipt["moves"][0].clone();
    assert_eq!(
        folder["path"],
        json!(".moai"),
        "the folder moves first: {receipt}"
    );
    assert_eq!(folder["kind"], json!("folder"), "{receipt}");
    assert_eq!(folder["moved"], json!(true));
    assert!(folder["sha256"].as_str().unwrap().starts_with("sha256:"));
    assert_eq!(receipt["plan_digest"], p["plan_digest"]);

    // AC-2: the preview after the apply offers nothing more (the folder is `rust` now), and
    // a second apply is a no-op that creates no new backup folder.
    let again = plan(&fx);
    assert_eq!(moves(&again), 0, "{again}");
    assert!(
        again["warnings"]
            .to_string()
            .contains("state_root is `rust`"),
        "{again}"
    );
    let after = fx.tree();
    let v = apply(&fx, again["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], json!(true), "{v}");
    assert_eq!(v["result"]["state"], json!("unchanged"), "{v}");
    assert_eq!(fx.tree(), after);
}

#[test]
fn a_file_changed_after_the_preview_stops_the_apply() {
    let fx = movable_fixture();
    let p = plan(&fx);
    fx.write(".claude/workflows/w.md", "# w, edited after the preview\n");
    let before = fx.tree();
    let v = apply(&fx, p["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], json!(false), "{v}");
    assert_eq!(fx.tree(), before);
}

/// A linked earlier root keeps the Go folder in place (its state cannot move into the new
/// `.moai/`), and the stage-2 backup area under it is refused as a link.
#[test]
fn a_linked_backup_area_is_refused() {
    let fx = movable_fixture();
    let elsewhere = fx.root.parent().unwrap().join("elsewhere");
    fs::create_dir_all(&elsewhere).unwrap();
    symlink(&elsewhere, fx.root.join(".moai-cli")).unwrap();
    let p = plan(&fx);
    let before = fx.tree();
    let v = apply(&fx, p["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], json!(false), "{v}");
    assert!(v.to_string().contains("symlink"), "{v}");
    assert_eq!(fx.tree(), before);
    assert_eq!(fs::read_dir(&elsewhere).unwrap().count(), 0);
}

/// Step 2 fails right after the rename (debug-build fault seam): the rename is undone at
/// once, no `.moai-backup` and no new `.moai/` are left, and the project is exactly as before.
#[test]
fn a_failure_right_after_the_rename_puts_the_go_folder_back() {
    let fx = movable_fixture();
    let before = fx.tree();
    let p = plan(&fx);
    let out = Command::new(EXE)
        .args([
            "go-state-apply",
            fx.root.to_str().unwrap(),
            "--plan-digest",
            p["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            REF,
        ])
        .clean_env()
        .env("PATH", "/usr/bin:/bin")
        .env("LC_ALL", "C")
        .env("MOAI_CLI_TEST_FAULT", "go_state_after_rename")
        .output()
        .unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], json!(false), "{v}");
    assert!(
        v.to_string().contains("the Go folder was renamed back"),
        "{v}"
    );
    assert_eq!(fx.tree(), before);
    assert!(!fx.root.join(".moai-backup").exists());
    assert!(fx.root.join(".moai/manifest.json").is_file());
}

fn revert_plan(fx: &Fx, backup: &str) -> Value {
    call(&[
        "go-state-revert-plan",
        fx.root.to_str().unwrap(),
        "--backup",
        backup,
    ])
}

fn revert(fx: &Fx, backup: &str, digest: &str) -> Value {
    call(&[
        "go-state-revert",
        fx.root.to_str().unwrap(),
        "--backup",
        backup,
        "--plan-digest",
        digest,
        "--permission-reference",
        REF,
    ])
}

#[test]
fn revert_restores_the_project_tree() {
    let fx = movable_fixture();
    let before = project_tree(&fx);
    let p = plan(&fx);
    let v = apply(&fx, p["plan_digest"].as_str().unwrap());
    let backup = v["result"]["backup"].as_str().unwrap().to_string();
    assert_ne!(project_tree(&fx), before);

    let rp = revert_plan(&fx, &backup);
    assert_eq!(rp["ok"], json!(true), "{rp}");
    let rp = &rp["result"];
    let restore = rp["items"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|i| i["action"] == "restore")
        .count();
    assert_eq!(restore, 2, "{rp}");

    let wrong = revert(&fx, &backup, "sha256:0000");
    assert_eq!(wrong["ok"], json!(false), "{wrong}");

    let v = revert(&fx, &backup, rp["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], json!(true), "{v}");
    // AC-3: the project tree outside the backup areas is what it was before the apply: the
    // Go folder is `.moai` again and no `.moai-backup` is left.
    assert_eq!(project_tree(&fx), before);
    assert!(!fx.root.join(".moai-backup").exists());
    // This generation's `.moai/` was set aside whole, marker removed, receipt `reverted`.
    let aside = rust_aside(&fx);
    assert!(!aside.join("moai-cli.json").exists());
    let (path, receipt) = receipt_of(&fx, &backup);
    assert!(path.starts_with(&aside), "{path:?}");
    assert_eq!(receipt["state"], json!("reverted"), "{receipt}");
    assert_eq!(receipt["marker"], Value::Null, "{receipt}");
    assert!(receipt["rust_aside"]
        .as_str()
        .is_some_and(|a| aside.ends_with(a)));
}

#[test]
fn revert_never_overwrites_a_file_that_came_back() {
    let fx = movable_fixture();
    let p = plan(&fx);
    let v = apply(&fx, p["plan_digest"].as_str().unwrap());
    let backup = v["result"]["backup"].as_str().unwrap().to_string();
    fx.write(".claude/workflows/w.md", "# new, written after the apply\n");
    let rp = revert_plan(&fx, &backup);
    let rp = &rp["result"];
    let item = rp["items"]
        .as_array()
        .unwrap()
        .iter()
        .find(|i| i["path"] == ".claude/workflows/w.md")
        .unwrap()
        .clone();
    assert_eq!(item["action"], json!("conflict"), "{rp}");
    let before = fx.tree();
    let v = revert(&fx, &backup, rp["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], json!(false), "{v}");
    assert_eq!(fx.tree(), before, "a refused revert writes nothing");
}

#[test]
fn revert_refuses_a_backup_name_that_is_not_a_stamp() {
    let fx = movable_fixture();
    for bad in ["../x", "20261112T000000Z/..", "", "x"] {
        let v = revert_plan(&fx, bad);
        assert_eq!(v["ok"], json!(false), "{bad}: {v}");
    }
    // A name that climbs out of the backup area must be refused for its shape, even when a
    // well-formed receipt is waiting at the place it points to.
    let escape = "../../evil";
    let receipt = json!({"schema_version": 1, "project": fx.root.to_str().unwrap(),
        "backup": escape, "plan_digest": "sha256:x", "permission_reference": "r",
        "state": "committed", "moves": []});
    fx.write("evil/go-state-receipt.json", &receipt.to_string());
    let v = revert_plan(&fx, escape);
    assert_eq!(v["ok"], json!(false), "{v}");
    assert!(v.to_string().contains("invalid backup name"), "{v}");
}

#[test]
fn a_changed_backup_copy_refuses_the_revert() {
    let fx = movable_fixture();
    let p = plan(&fx);
    let v = apply(&fx, p["plan_digest"].as_str().unwrap());
    let backup = v["result"]["backup"].as_str().unwrap().to_string();
    fx.write(
        &format!(".moai/backups/{backup}/go-state/.claude/workflows/w.md"),
        "# tampered\n",
    );
    let rp = revert_plan(&fx, &backup);
    let rp = &rp["result"];
    let item = rp["items"]
        .as_array()
        .unwrap()
        .iter()
        .find(|i| i["path"] == ".claude/workflows/w.md")
        .unwrap()
        .clone();
    assert_eq!(item["action"], json!("missing"), "{rp}");
    let before = fx.tree();
    let v = revert(&fx, &backup, rp["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], json!(false), "{v}");
    assert_eq!(fx.tree(), before);
}

/// A failure midway leaves a `pending` receipt that names what already moved, `state_root`
/// says `interrupted` (so nothing else moves), and the revert brings exactly that back.
#[test]
fn a_failure_midway_leaves_a_pending_receipt_the_revert_can_use() {
    use std::os::unix::fs::PermissionsExt;
    let fx = movable_fixture();
    fx.write(".moai-cli/index/a.json", "{}");
    fx.write(".moai-cli/project.yaml", "schema_version: 2\n");
    let before = project_tree(&fx);
    let p = plan(&fx);
    // Step 4 moves `.moai-cli/index` first; a locked `.moai-cli` cannot lose an entry.
    let locked = fx.root.join(".moai-cli");
    fs::set_permissions(&locked, fs::Permissions::from_mode(0o555)).unwrap();
    let v = apply(&fx, p["plan_digest"].as_str().unwrap());
    fs::set_permissions(&locked, fs::Permissions::from_mode(0o755)).unwrap();
    assert_eq!(v["ok"], json!(false), "{v}");
    assert!(v.to_string().contains(".moai-cli/index"), "{v}");
    let backups = fx.root.join(".moai/backups");
    let mut stamps: Vec<String> = fs::read_dir(&backups)
        .unwrap()
        .map(|e| e.unwrap().file_name().into_string().unwrap())
        .collect();
    assert_eq!(stamps.len(), 1);
    let backup = stamps.pop().unwrap();
    let (_, receipt) = receipt_of(&fx, &backup);
    assert_eq!(receipt["state"], json!("pending"), "{receipt}");
    let moved: Vec<(&str, bool)> = receipt["moves"]
        .as_array()
        .unwrap()
        .iter()
        .map(|m| (m["path"].as_str().unwrap(), m["moved"].as_bool().unwrap()))
        .collect();
    assert_eq!(
        moved,
        vec![
            (".moai", true),
            (".claude/workflows/w.md", true),
            (".moai-cli/index", false),
            (".moai-cli/project.yaml", false),
        ]
    );
    assert!(!fx.root.join(".moai/moai-cli.json").exists(), "no marker");
    // Stopped midway: every further preview holds.
    let again = plan(&fx);
    assert_eq!(moves(&again), 0, "{again}");
    assert!(
        again["warnings"]
            .to_string()
            .contains("state_root is `interrupted`"),
        "{again}"
    );

    let rp = revert_plan(&fx, &backup);
    let v = revert(&fx, &backup, rp["result"]["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], json!(true), "{v}");
    assert_eq!(project_tree(&fx), before);
}

#[test]
fn an_empty_permission_reference_writes_nothing() {
    let fx = movable_fixture();
    let p = plan(&fx);
    let before = fx.tree();
    let v = call(&[
        "go-state-apply",
        fx.root.to_str().unwrap(),
        "--plan-digest",
        p["plan_digest"].as_str().unwrap(),
        "--permission-reference",
        " ",
    ]);
    assert_eq!(v["ok"], json!(false), "{v}");
    assert_eq!(fx.tree(), before);

    let v = apply(&fx, p["plan_digest"].as_str().unwrap());
    let backup = v["result"]["backup"].as_str().unwrap().to_string();
    let rp = revert_plan(&fx, &backup);
    let before = fx.tree();
    let v = call(&[
        "go-state-revert",
        fx.root.to_str().unwrap(),
        "--backup",
        &backup,
        "--plan-digest",
        rp["result"]["plan_digest"].as_str().unwrap(),
        "--permission-reference",
        "",
    ]);
    assert_eq!(v["ok"], json!(false), "{v}");
    assert_eq!(fx.tree(), before);
}

#[test]
fn a_reverted_apply_is_not_reverted_twice() {
    let fx = movable_fixture();
    let p = plan(&fx);
    let v = apply(&fx, p["plan_digest"].as_str().unwrap());
    let backup = v["result"]["backup"].as_str().unwrap().to_string();
    let rp = revert_plan(&fx, &backup);
    let v = revert(&fx, &backup, rp["result"]["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], json!(true), "{v}");
    let rp = revert_plan(&fx, &backup);
    assert_eq!(rp["result"]["receipt_state"], json!("reverted"), "{rp}");
    let before = fx.tree();
    let v = revert(&fx, &backup, rp["result"]["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], json!(false), "{v}");
    assert!(v.to_string().contains("already reverted"), "{v}");
    assert_eq!(fx.tree(), before);
}

/// The receipt is not its own authority: a receipt edited after the apply (or planted)
/// must not steer a revert outside what an apply can have moved.
#[test]
fn a_receipt_that_points_elsewhere_is_refused() {
    let fx = movable_fixture();
    let p = plan(&fx);
    let v = apply(&fx, p["plan_digest"].as_str().unwrap());
    let backup = v["result"]["backup"].as_str().unwrap().to_string();
    let dir = area(&fx, &backup);
    let receipt_path = dir.join("go-state-receipt.json");
    let original: Value = serde_json::from_slice(&fs::read(&receipt_path).unwrap()).unwrap();
    assert_eq!(
        original["moves"][1]["path"],
        json!(".claude/workflows/w.md")
    );
    // A copy staged where a forged path would restore from, with the recorded hash.
    let staged = dir.join("go-state/src/evil.rs");
    fs::create_dir_all(staged.parent().unwrap()).unwrap();
    fs::write(&staged, "evil\n").unwrap();
    type Forge = fn(&mut Value);
    let cases: Vec<(&str, Forge)> = vec![
        ("another project", |r: &mut Value| {
            r["project"] = json!("/elsewhere")
        }),
        ("a path that climbs", |r: &mut Value| {
            r["moves"][1]["path"] = json!(".claude/../src/evil.rs")
        }),
        ("a path outside .moai/ and .claude/", |r: &mut Value| {
            r["moves"][1]["path"] = json!("src/evil.rs");
            r["moves"][1]["sha256"] = json!(sha(b"evil\n"));
        }),
        ("a state entry this build never moves", |r: &mut Value| {
            r["moves"][1]["path"] = json!(".moai-cli/worktrees");
            r["moves"][1]["kind"] = json!("state");
        }),
        ("a state entry outside the earlier root", |r: &mut Value| {
            r["moves"][1]["path"] = json!("src/evil.rs");
            r["moves"][1]["kind"] = json!("state");
        }),
        (
            "no Go backup name (the stage-2 layout found under .moai)",
            |r: &mut Value| {
                r.as_object_mut().unwrap().remove("go_backup");
            },
        ),
        ("a Go backup name that is not one", |r: &mut Value| {
            r["go_backup"] = json!(".moai-backups");
        }),
        ("an aside name that is not one", |r: &mut Value| {
            r["rust_aside"] = json!("src");
        }),
        ("a root-file copy elsewhere", |r: &mut Value| {
            r["copied"] = json!([{"path": "src/evil.rs", "sha256": sha(b"evil\n")}]);
        }),
    ];
    for (name, edit) in cases {
        let mut forged = original.clone();
        edit(&mut forged);
        fs::write(&receipt_path, forged.to_string()).unwrap();
        let rp = revert_plan(&fx, &backup);
        assert_eq!(rp["ok"], json!(false), "{name}: {rp}");
        assert!(!fx.root.join("src/evil.rs").exists(), "{name}");
    }
}

/// AUDIT #2584 B1: an unchanged Go `.claude/settings.json` is a recorded, unchanged
/// template file, yet it is the live host settings (permissions.deny, hooks, statusLine).
/// Moving it would drop those rules, so both settings files always stay, and a file the
/// local settings name stays like one the shared settings name.
#[test]
fn the_live_settings_files_always_stay() {
    let mut fx = Fx::new();
    let shared = json!({"permissions": {"deny": ["Bash(rm -rf:*)", "Read(./.env)"]},
        "hooks": {"PreToolUse": [{"hooks": [{"type": "command",
            "command": "\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-pre-tool.sh\""}]}]}})
    .to_string();
    fx.deployed(".claude/settings.json", &shared);
    let local = json!({"statusLine": {"type": "command",
        "command": "\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/status.sh\""}})
    .to_string();
    fx.deployed(".claude/settings.local.json", &local);
    fx.deployed(".claude/hooks/moai/handle-pre-tool.sh", "#!/bin/sh\n");
    fx.deployed(".claude/hooks/moai/status.sh", "#!/bin/sh\necho s\n");
    fx.deployed(".claude/workflows/free.md", "# free\n");
    fx.manifest();
    let p = plan(&fx);
    for rel in [".claude/settings.json", ".claude/settings.local.json"] {
        assert_eq!(action(&p, rel).0, "leave", "{rel}: {p}");
    }
    assert_eq!(
        action(&p, ".claude/hooks/moai/status.sh"),
        ("leave", "referenced by .claude/settings.local.json")
    );
    assert_eq!(action(&p, ".claude/workflows/free.md").0, "move");
    let v = apply(&fx, p["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], json!(true), "{v}");
    assert_eq!(
        fs::read_to_string(fx.root.join(".claude/settings.json")).unwrap(),
        shared,
        "the deny rules and hooks are still in place after the apply"
    );
    assert!(fx.root.join(".claude/settings.local.json").is_file());
    assert!(fx.root.join(".claude/hooks/moai/status.sh").is_file());
}

// ---- stage 3a: `.moai/` as one folder (decision 1, #2561 comment 6831345102) ----

/// The receipt of `backup`, wherever its layout keeps it: `.moai/backups/` (#2696), the
/// earlier root's `backups/` (the Go folder stayed), or a set-aside `.moai.rust-<UTC>`.
fn receipt_of(fx: &Fx, backup: &str) -> (PathBuf, Value) {
    let mut bases = vec![fx.root.join(".moai"), fx.root.join(".moai-cli")];
    bases.extend(
        fs::read_dir(&fx.root)
            .unwrap()
            .map(|e| e.unwrap().path())
            .filter(|p| p.to_string_lossy().contains("/.moai.rust-")),
    );
    let path = bases
        .iter()
        .map(|b| b.join("backups").join(backup).join("go-state-receipt.json"))
        .find(|p| p.is_file())
        .unwrap_or_else(|| panic!("no receipt for {backup}"));
    let v = serde_json::from_slice(&fs::read(&path).unwrap()).unwrap();
    (path, v)
}

/// Decision 2 (fail-closed): while a settings file names a path under `.moai/`, the folder
/// and everything in it stay; `.claude/` files still move one by one, and the receipt
/// has no folder entry (so it reads like a stage-2 receipt, without `kind`).
#[test]
fn settings_that_reference_moai_keep_the_whole_folder() {
    for file in [".claude/settings.json", ".claude/settings.local.json"] {
        let mut fx = movable_fixture();
        let text = json!({"statusLine": {"type": "command",
            "command": "\"$CLAUDE_PROJECT_DIR/.moai/status_line.sh\""}})
        .to_string();
        fx.write(file, &text);
        fx.record(file, "user_created", &sha(text.as_bytes()));
        fx.manifest();
        let p = plan(&fx);
        assert_eq!(
            action(&p, ".moai"),
            (
                "leave",
                "settings reference .moai/; the folder stays until they are cleaned (#2581)"
            ),
            "{file}"
        );
        assert_eq!(
            action(&p, ".moai/config/sections/a.yaml"),
            ("leave", "the .moai/ folder stays (see the .moai item)")
        );
        assert_eq!(moves(&p), 1, "{file}: only the .claude file: {p}");
        let v = apply(&fx, p["plan_digest"].as_str().unwrap());
        assert_eq!(v["ok"], json!(true), "{v}");
        assert!(fx.root.join(".moai/config/sections/a.yaml").is_file());
        let (_, r) = receipt_of(&fx, v["result"]["backup"].as_str().unwrap());
        assert!(!r.to_string().contains("\"kind\""), "{r}");
    }
}

/// A receipt of the #2561 stage-3 layout (an earlier build moved `.moai/` itself to
/// `.moai-cli/backups/<UTC>/go-state/.moai`), made by hand: revert still reads it.
const LEGACY: &str = "20261101T000000Z";

fn legacy_stage3(fx: &Fx, moved: bool, sha256: &str) {
    let dest = fx
        .root
        .join(".moai-cli/backups")
        .join(LEGACY)
        .join("go-state");
    fs::create_dir_all(&dest).unwrap();
    fs::rename(fx.root.join(".moai"), dest.join(".moai")).unwrap();
    fx.write(
        &format!(".moai-cli/backups/{LEGACY}/go-state-receipt.json"),
        &json!({"schema_version": 1, "project": fx.root.to_str().unwrap(),
            "backup": LEGACY, "plan_digest": "sha256:legacy", "permission_reference": "r",
            "state": "committed",
            "moves": [{"path": ".moai", "sha256": sha256, "moved": moved, "kind": "folder"}]})
        .to_string(),
    );
}

fn item_action(rp: &Value, path: &str) -> Value {
    rp["result"]["items"]
        .as_array()
        .unwrap_or_else(|| panic!("{rp}"))
        .iter()
        .find(|i| i["path"] == path)
        .unwrap_or_else(|| panic!("{path}: {rp}"))["action"]
        .clone()
}

/// Legacy layout: a live hook recreated `.moai/` after the apply; the revert never merges
/// into it.
#[test]
fn a_moai_folder_that_came_back_refuses_the_revert() {
    let fx = movable_fixture();
    legacy_stage3(&fx, true, "");
    fx.write(".moai/logs/hook.log", "written by a live hook\n");
    let rp = revert_plan(&fx, LEGACY);
    assert_eq!(item_action(&rp, ".moai"), json!("conflict"), "{rp}");
    let before = fx.tree();
    let v = revert(&fx, LEGACY, rp["result"]["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], json!(false), "{v}");
    assert!(v.to_string().contains(".moai is conflict"), "{v}");
    assert_eq!(fx.tree(), before);
}

/// Legacy layout: the folder's backup is checked against the tree digest the apply recorded,
/// and a folder move the receipt missed (`moved: false`, no digest) is still brought back.
#[test]
fn a_legacy_folder_backup_is_checked_and_restored() {
    let fx = movable_fixture();
    legacy_stage3(&fx, true, &sha(b"not this tree"));
    let rp = revert_plan(&fx, LEGACY);
    assert_eq!(item_action(&rp, ".moai"), json!("missing"), "{rp}");
    let before = fx.tree();
    let v = revert(&fx, LEGACY, rp["result"]["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], json!(false), "{v}");
    assert_eq!(fx.tree(), before);

    let fx = movable_fixture();
    let before = project_tree(&fx);
    legacy_stage3(&fx, false, "");
    let rp = revert_plan(&fx, LEGACY);
    assert_eq!(item_action(&rp, ".moai"), json!("restore"), "{rp}");
    let v = revert(&fx, LEGACY, rp["result"]["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], json!(true), "{v}");
    assert_eq!(project_tree(&fx), before);
}

/// #2696: the Go backup is checked against the tree digest the apply recorded.
#[test]
fn a_changed_folder_backup_refuses_the_revert() {
    let fx = movable_fixture();
    let p = plan(&fx);
    let v = apply(&fx, p["plan_digest"].as_str().unwrap());
    let backup = v["result"]["backup"].as_str().unwrap().to_string();
    fx.write(".moai-backup/specs/SPEC-1/spec.md", "# tampered\n");
    let rp = revert_plan(&fx, &backup);
    assert_eq!(item_action(&rp, ".moai"), json!("missing"), "{rp}");
    let before = fx.tree();
    let v = revert(&fx, &backup, rp["result"]["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], json!(false), "{v}");
    assert_eq!(fx.tree(), before);
}

/// #2696: the folder was renamed but the receipt write after it did not happen (no digest):
/// the disk decides, and the revert brings it back.
#[test]
fn a_folder_move_the_receipt_missed_is_still_restored() {
    let fx = movable_fixture();
    let before = project_tree(&fx);
    let p = plan(&fx);
    let v = apply(&fx, p["plan_digest"].as_str().unwrap());
    let backup = v["result"]["backup"].as_str().unwrap().to_string();
    let (path, mut r) = receipt_of(&fx, &backup);
    r["moves"][0]["sha256"] = json!("");
    r["state"] = json!("pending");
    fs::write(&path, r.to_string()).unwrap();
    let rp = revert_plan(&fx, &backup);
    assert_eq!(item_action(&rp, ".moai"), json!("restore"), "{rp}");
    let v = revert(&fx, &backup, rp["result"]["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], json!(true), "{v}");
    assert_eq!(project_tree(&fx), before);
}

/// #2696: a revert that stopped right after setting `.moai/` aside finishes on the next run
/// (the receipt is found in the aside folder), and never merges into a `.moai` that came
/// back meanwhile.
#[test]
fn a_revert_stopped_after_the_set_aside_finishes_on_the_next_run() {
    let fx = movable_fixture();
    let before = project_tree(&fx);
    let p = plan(&fx);
    let v = apply(&fx, p["plan_digest"].as_str().unwrap());
    let backup = v["result"]["backup"].as_str().unwrap().to_string();
    // What the revert does up to the set-aside: marker removed, files back, aside recorded.
    let rp = revert_plan(&fx, &backup);
    assert_eq!(item_action(&rp, ".moai/moai-cli.json"), json!("remove"));
    fs::remove_file(fx.root.join(".moai/moai-cli.json")).unwrap();
    fs::rename(
        area(&fx, &backup).join("go-state/.claude/workflows/w.md"),
        fx.root.join(".claude/workflows/w.md"),
    )
    .unwrap();
    let (path, mut r) = receipt_of(&fx, &backup);
    r["moves"][1]["moved"] = json!(false);
    r.as_object_mut().unwrap().remove("marker");
    r["rust_aside"] = json!(".moai.rust-20261113T190000Z");
    fs::write(&path, r.to_string()).unwrap();
    fs::rename(
        fx.root.join(".moai"),
        fx.root.join(".moai.rust-20261113T190000Z"),
    )
    .unwrap();
    // A receipt found in an aside folder must name that folder.
    let (path, original) = receipt_of(&fx, &backup);
    let mut forged = original.clone();
    forged["rust_aside"] = json!(".moai.rust-20261113T190001Z");
    fs::write(&path, forged.to_string()).unwrap();
    assert_eq!(revert_plan(&fx, &backup)["ok"], json!(false));
    fs::write(&path, original.to_string()).unwrap();
    // Interrupted now: the Go backup is beside a missing `.moai`.
    assert!(plan(&fx)["warnings"]
        .to_string()
        .contains("state_root is `interrupted`"));
    // Something at `.moai` again: never merged.
    fx.write(".moai/logs/hook.log", "written by a live hook\n");
    let rp = revert_plan(&fx, &backup);
    assert_eq!(item_action(&rp, ".moai"), json!("conflict"), "{rp}");
    assert_eq!(
        item_action(&rp, ".moai/moai-cli.json"),
        json!("skip"),
        "{rp}"
    );
    fs::remove_dir_all(fx.root.join(".moai")).unwrap();
    let rp = revert_plan(&fx, &backup);
    assert_eq!(item_action(&rp, ".moai"), json!("restore"), "{rp}");
    let v = revert(&fx, &backup, rp["result"]["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], json!(true), "{v}");
    assert_eq!(project_tree(&fx), before);
    let (path, r) = receipt_of(&fx, &backup);
    assert!(path.starts_with(fx.root.join(".moai.rust-20261113T190000Z")));
    assert_eq!(r["state"], json!("reverted"), "{r}");
}

/// A receipt may only name the folder as `.moai` with kind `folder`, and a file entry is
/// never the folder itself.
#[test]
fn a_receipt_with_a_wrong_kind_is_refused() {
    let fx = movable_fixture();
    let p = plan(&fx);
    let v = apply(&fx, p["plan_digest"].as_str().unwrap());
    let backup = v["result"]["backup"].as_str().unwrap().to_string();
    let (path, original) = receipt_of(&fx, &backup);
    type Forge = fn(&mut Value);
    let cases: Vec<(&str, Forge)> = vec![
        ("folder kind on a file", |r: &mut Value| {
            r["moves"][1]["kind"] = json!("folder")
        }),
        ("file kind on the folder", |r: &mut Value| {
            r["moves"][0]["kind"] = json!("");
        }),
        ("unknown kind", |r: &mut Value| {
            r["moves"][1]["kind"] = json!("tree")
        }),
        ("a Go backup without a folder move", |r: &mut Value| {
            r["moves"].as_array_mut().unwrap().remove(0);
        }),
    ];
    for (name, edit) in cases {
        let mut forged = original.clone();
        edit(&mut forged);
        fs::write(&path, forged.to_string()).unwrap();
        let rp = revert_plan(&fx, &backup);
        assert_eq!(rp["ok"], json!(false), "{name}: {rp}");
    }
}

// ---- stage 3b: restore the Go SPECs into `.moai/specs` (AC-5) ----

fn restore_action<'a>(plan: &'a Value, from: &str) -> &'a str {
    plan["restores"]
        .as_array()
        .unwrap()
        .iter()
        .find(|r| r["from"] == from)
        .unwrap_or_else(|| panic!("{from} not in restores: {plan}"))["action"]
        .as_str()
        .unwrap()
}

/// AC-5: the SPEC is copied to `.moai/specs` from the Go backup and the original stays in
/// the backup; a SPEC that holds a secret-looking value is not copied; the receipt names
/// each copy with its hash.
#[test]
fn specs_are_restored_from_the_backup_and_secrets_are_not() {
    let fx = movable_fixture();
    fx.write(
        ".moai/specs/SPEC-2/spec.md",
        "api_key: sk-live-0123456789abcdef\n",
    );
    let p = plan(&fx);
    assert_eq!(
        restore_action(&p, ".moai-backup/specs/SPEC-1/spec.md"),
        "copy"
    );
    assert_eq!(
        restore_action(&p, ".moai-backup/specs/SPEC-2/spec.md"),
        "excluded_secret"
    );
    let v = apply(&fx, p["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], json!(true), "{v}");
    let backup = v["result"]["backup"].as_str().unwrap().to_string();
    assert_eq!(
        fs::read_to_string(fx.root.join(".moai/specs/SPEC-1/spec.md")).unwrap(),
        "# spec\n"
    );
    assert!(!fx.root.join(".moai/specs/SPEC-2").exists());
    assert!(fx.root.join(".moai-backup/specs/SPEC-1/spec.md").is_file());
    assert!(fx.root.join(".moai-backup/specs/SPEC-2/spec.md").is_file());
    let (_, r) = receipt_of(&fx, &backup);
    assert_eq!(
        r["restored"],
        json!([{"path": ".moai/specs/SPEC-1/spec.md", "sha256": sha(b"# spec\n")}])
    );
}

/// An existing target is never overwritten: a SPEC already in the earlier root's `specs/`
/// (which moves to `.moai/specs` first) makes the preview say `conflict` and the apply
/// write nothing at all.
#[test]
fn an_existing_restore_target_stops_the_apply() {
    let fx = movable_fixture();
    fx.write(".moai-cli/specs/SPEC-1/spec.md", "# mine, already here\n");
    let p = plan(&fx);
    assert_eq!(
        restore_action(&p, ".moai-backup/specs/SPEC-1/spec.md"),
        "conflict"
    );
    let before = fx.tree();
    let v = apply(&fx, p["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], json!(false), "{v}");
    assert!(v.to_string().contains("never overwritten"), "{v}");
    assert_eq!(fx.tree(), before);
}

/// The revert removes an unchanged restored copy and keeps (and reports) a changed one; both
/// are in this generation's `.moai/`, which is set aside whole, so the Go SPECs are untouched.
#[test]
fn revert_removes_only_unchanged_restored_copies() {
    let fx = movable_fixture();
    fx.write(".moai/specs/SPEC-3/spec.md", "# three\n");
    let p = plan(&fx);
    let v = apply(&fx, p["plan_digest"].as_str().unwrap());
    let backup = v["result"]["backup"].as_str().unwrap().to_string();
    fx.write(
        ".moai/specs/SPEC-3/spec.md",
        "# three, edited after the apply\n",
    );
    let rp = revert_plan(&fx, &backup);
    assert_eq!(
        item_action(&rp, ".moai/specs/SPEC-1/spec.md"),
        json!("remove_restored")
    );
    assert_eq!(
        item_action(&rp, ".moai/specs/SPEC-3/spec.md"),
        json!("keep_changed")
    );
    let v = revert(&fx, &backup, rp["result"]["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], json!(true), "{v}");
    let aside = rust_aside(&fx);
    assert!(!aside.join("specs/SPEC-1/spec.md").exists());
    assert_eq!(
        fs::read_to_string(aside.join("specs/SPEC-3/spec.md")).unwrap(),
        "# three, edited after the apply\n"
    );
    assert_eq!(
        fs::read_to_string(fx.root.join(".moai/specs/SPEC-3/spec.md")).unwrap(),
        "# three\n"
    );
    assert!(fx.root.join(".moai/specs/SPEC-1/spec.md").is_file());
}

/// A receipt may only name restored copies under `.moai/specs/`.
#[test]
fn a_receipt_restored_path_elsewhere_is_refused() {
    let fx = movable_fixture();
    let p = plan(&fx);
    let v = apply(&fx, p["plan_digest"].as_str().unwrap());
    let backup = v["result"]["backup"].as_str().unwrap().to_string();
    let (path, mut r) = receipt_of(&fx, &backup);
    for elsewhere in ["src/main.rs", ".moai-cli/specs/SPEC-1/spec.md"] {
        r["restored"][0]["path"] = json!(elsewhere);
        fs::write(&path, r.to_string()).unwrap();
        let rp = revert_plan(&fx, &backup);
        assert_eq!(rp["ok"], json!(false), "{elsewhere}: {rp}");
    }
}

/// #2561 F2 (comment 6856196324): moving `.moai/` also moves every git worktree registered
/// under it, and git then loses them. A worktree there keeps the whole folder in place, with
/// the count in the reason; a worktree elsewhere does not.
#[test]
fn a_worktree_registered_under_moai_keeps_the_whole_folder() {
    let id = [
        "-c",
        "user.name=fixture",
        "-c",
        "user.email=fixture@example.invalid",
    ];
    let git = |args: &[&str], fx: &Fx| {
        let mut all: Vec<&str> = id.to_vec();
        all.extend_from_slice(args);
        common::git::git(&fx.root, &all)
    };
    // A checkout whose .moai/ holds nothing registered: the folder still moves.
    let fx = movable_fixture();
    git(&["init", "-q", "-b", "main"], &fx);
    fx.write("README.md", "r\n");
    git(&["add", "README.md"], &fx);
    git(&["commit", "-q", "-m", "init"], &fx);
    let outside = fx.root.join("wt-outside");
    git(
        &[
            "worktree",
            "add",
            "-q",
            "-b",
            "o",
            outside.to_str().unwrap(),
        ],
        &fx,
    );
    assert_eq!(action(&plan(&fx), ".moai").0, "move");

    // Two worktrees under .moai/: the folder stays, the reason names the count.
    for (branch, rel) in [
        ("a", ".moai/worktrees/a"),
        ("b", ".moai/worktrees/projects/x/b"),
    ] {
        let wt = fx.root.join(rel);
        git(
            &["worktree", "add", "-q", "-b", branch, wt.to_str().unwrap()],
            &fx,
        );
    }
    let p = plan(&fx);
    let (act, why) = action(&p, ".moai");
    assert_eq!(act, "leave", "{p}");
    assert!(
        why.starts_with("2 git worktree(s) are registered under .moai/"),
        "{why}"
    );
    assert_eq!(
        action(&p, ".moai/config/sections/a.yaml"),
        ("leave", "the .moai/ folder stays (see the .moai item)")
    );
    let v = apply(&fx, p["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], json!(true), "{v}");
    assert!(fx.root.join(".moai/worktrees/a/README.md").is_file());
    let list = git(&["worktree", "list", "--porcelain"], &fx);
    assert!(!list.contains("prunable"), "{list}");
}

/// F2: when the project is a git checkout whose worktree list cannot be read, nobody knows
/// whether a worktree lives under `.moai/`, so the folder stays (fail-closed).
#[test]
fn an_unreadable_worktree_list_keeps_the_whole_folder() {
    let fx = movable_fixture();
    fx.write(".git", "gitdir: /nonexistent/moai-fixture-gitdir\n");
    let p = plan(&fx);
    let (act, why) = action(&p, ".moai");
    assert_eq!(act, "leave", "{p}");
    assert!(why.contains("could not be read"), "{why}");
}

/// #2561 F1 (comment 6856196324): a project's worktrees under `.claude/worktrees/` can hold
/// hundreds of thousands of files (a preview once ran 14 minutes). A folder holding a `.git`
/// entry is skipped unread and counted in a warning; a plain folder is still counted.
#[test]
fn nested_checkouts_are_skipped_and_counted() {
    let fx = movable_fixture();
    for wt in [".claude/worktrees/w1", ".moai/worktrees/w2"] {
        fx.write(&format!("{wt}/.git"), "gitdir: /elsewhere\n");
        fx.write(&format!("{wt}/node_modules/x/index.js"), "x\n");
        fx.write(&format!("{wt}/src/a.rs"), "a\n");
    }
    fx.write(".claude/worktrees/plain/notes.md", "n\n");
    let p = plan(&fx);
    assert_eq!(p["unrecorded"][".claude/worktrees"], json!(1), "{p}");
    assert_eq!(p["unrecorded"][".moai/worktrees"], Value::Null, "{p}");
    let warned: Vec<&str> = p["warnings"]
        .as_array()
        .unwrap()
        .iter()
        .filter_map(|w| w.as_str())
        .filter(|w| w.contains("nested checkout"))
        .collect();
    assert_eq!(warned.len(), 1, "{p}");
    assert!(
        warned[0].starts_with("2 nested checkout(s)"),
        "{}",
        warned[0]
    );
    // Without nested checkouts there is no such warning.
    let clean = movable_fixture();
    assert!(!plan(&clean)["warnings"]
        .to_string()
        .contains("nested checkout"));
}

/// F1: the `unrecorded` count stops after a fixed number of entries and says so; at exactly
/// the limit it does not stop. `.moai/` here holds three entries: `config/`, `manifest.json`
/// and `config/a.yaml`.
#[test]
fn the_unrecorded_count_stops_at_its_entry_limit() {
    let mut fx = Fx::new();
    fx.deployed(".moai/config/a.yaml", "a\n");
    fx.manifest();
    let with_limit = |n: &str| -> Value {
        let out = Command::new(EXE)
            .args(["go-state-plan", fx.root.to_str().unwrap()])
            .clean_env()
            .env("PATH", "/usr/bin:/bin")
            .env("LC_ALL", "C")
            .env("MOAI_CLI_TEST_UNRECORDED_LIMIT", n)
            .output()
            .unwrap();
        let v: Value = serde_json::from_slice(&out.stdout).unwrap();
        assert_eq!(v["ok"], json!(true), "{v}");
        v["result"].clone()
    };
    let stopped = |p: &Value| p["warnings"].to_string().contains("stopped after");
    assert!(
        !stopped(&with_limit("3")),
        "three entries fit a limit of three"
    );
    let p = with_limit("2");
    assert!(stopped(&p), "{p}");
    assert!(
        p["warnings"]
            .to_string()
            .contains("stopped after 2 entries"),
        "{p}"
    );
    assert!(!stopped(&plan(&fx)), "the default limit is far above three");
}

/// #2753 audit F1 (6858965696): a worktree that another repository registered under `.moai/`
/// is not in this project's worktree list, yet moving the folder would cut that repository
/// off from it. Any nested checkout under `.moai/` keeps the whole folder in place.
#[test]
fn a_foreign_worktree_under_moai_keeps_the_whole_folder() {
    let id = [
        "-c",
        "user.name=fixture",
        "-c",
        "user.email=fixture@example.invalid",
    ];
    let git = |dir: &Path, args: &[&str]| {
        let mut all: Vec<&str> = id.to_vec();
        all.extend_from_slice(args);
        common::git::git(dir, &all)
    };
    let fx = movable_fixture();
    let other = fx.root.parent().unwrap().join("other-repo");
    fs::create_dir_all(&other).unwrap();
    git(&other, &["init", "-q", "-b", "main"]);
    fs::write(other.join("o.txt"), "o\n").unwrap();
    git(&other, &["add", "o.txt"]);
    git(&other, &["commit", "-q", "-m", "init"]);
    let foreign = fx.root.join(".moai/worktrees/foreign");
    git(
        &other,
        &[
            "worktree",
            "add",
            "-q",
            "-b",
            "f",
            foreign.to_str().unwrap(),
        ],
    );
    let p = plan(&fx);
    let (act, why) = action(&p, ".moai");
    assert_eq!(act, "leave", "{p}");
    assert!(why.starts_with("1 nested checkout(s)"), "{why}");
    let v = apply(&fx, p["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], json!(true), "{v}");
    assert!(foreign.join("o.txt").is_file());
    let list = git(&other, &["worktree", "list", "--porcelain"]);
    assert!(!list.contains("prunable"), "{list}");
}

/// F1 repair: a `.moai/` too large to check within the entry limit stays (fail-closed).
#[test]
fn a_moai_folder_past_the_entry_limit_stays() {
    let fx = movable_fixture();
    let out = Command::new(EXE)
        .args(["go-state-plan", fx.root.to_str().unwrap()])
        .clean_env()
        .env("PATH", "/usr/bin:/bin")
        .env("LC_ALL", "C")
        .env("MOAI_CLI_TEST_UNRECORDED_LIMIT", "2")
        .output()
        .unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    let (act, why) = action(&v["result"], ".moai");
    assert_eq!(act, "leave", "{v}");
    assert!(why.contains("was not fully checked"), "{why}");
    assert_eq!(
        action(&plan(&fx), ".moai").0,
        "move",
        "under the default limit it moves"
    );
}

/// #2753 re-audit F2 (6859648118): `.moai/` itself may be another repository's linked
/// worktree (`.moai/.git` is a gitdir file). Moving it would cut that repository off, so the
/// whole folder stays.
#[test]
fn a_moai_folder_that_is_itself_a_worktree_stays() {
    let id = [
        "-c",
        "user.name=fixture",
        "-c",
        "user.email=fixture@example.invalid",
    ];
    let git = |dir: &Path, args: &[&str]| {
        let mut all: Vec<&str> = id.to_vec();
        all.extend_from_slice(args);
        common::git::git(dir, &all)
    };
    let mut fx = Fx::new();
    let other = fx.root.parent().unwrap().join("owner-repo");
    fs::create_dir_all(&other).unwrap();
    git(&other, &["init", "-q", "-b", "main"]);
    fs::write(other.join("o.txt"), "o\n").unwrap();
    git(&other, &["add", "o.txt"]);
    git(&other, &["commit", "-q", "-m", "init"]);
    let moai = fx.root.join(".moai");
    git(
        &other,
        &["worktree", "add", "-q", "-b", "m", moai.to_str().unwrap()],
    );
    fx.deployed(".moai/config/sections/a.yaml", "a: 1\n");
    fx.deployed(".claude/workflows/w.md", "# w\n");
    fx.manifest();
    let p = plan(&fx);
    let (act, why) = action(&p, ".moai");
    assert_eq!(act, "leave", "{p}");
    assert!(why.starts_with("1 nested checkout(s)"), "{why}");
    let v = apply(&fx, p["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], json!(true), "{v}");
    assert!(moai.join("o.txt").is_file());
    let list = git(&other, &["worktree", "list", "--porcelain"]);
    assert!(!list.contains("prunable"), "{list}");
}

// ---- #2696 (#2725 S2c): the Go folder becomes `.moai-backup`, a new `.moai/` is made ----

fn state_move<'a>(plan: &'a Value, from: &str) -> &'a Value {
    plan["state_moves"]
        .as_array()
        .unwrap()
        .iter()
        .find(|m| m["from"] == from)
        .unwrap_or_else(|| panic!("{from} not in state_moves: {plan}"))
}

/// The earlier root's product state moves into the new `.moai/`; worktrees, reports, earlier
/// backups and anything unknown stay; the revert brings every moved entry back.
#[test]
fn the_earlier_roots_product_state_moves_into_the_new_moai() {
    let fx = movable_fixture();
    fx.write(".moai-cli/project.yaml", "schema_version: 2\n");
    fx.write(".moai-cli/index/a.json", "{\"a\":1}");
    fx.write(".moai-cli/receipts/r1.json", "{}");
    fx.write(".moai-cli/worktrees/lane/x", "x");
    fx.write(".moai-cli/reports/r.md", "# r\n");
    fx.write(".moai-cli/backups/20261101T000000Z/x", "x");
    fx.write(".moai-cli/notes.txt", "mine");
    let before = project_tree(&fx);
    let p = plan(&fx);
    for (from, action) in [
        (".moai-cli/project.yaml", "move"),
        (".moai-cli/index", "move"),
        (".moai-cli/receipts", "move"),
        (".moai-cli/worktrees", "leave"),
        (".moai-cli/reports", "leave"),
        (".moai-cli/backups", "leave"),
        (".moai-cli/notes.txt", "leave"),
    ] {
        assert_eq!(state_move(&p, from)["action"], json!(action), "{from}: {p}");
    }
    assert_eq!(
        state_move(&p, ".moai-cli/index")["to"],
        json!(".moai/index")
    );
    let v = apply(&fx, p["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], json!(true), "{v}");
    let backup = v["result"]["backup"].as_str().unwrap().to_string();
    for (rel, body) in [
        (".moai/project.yaml", "schema_version: 2\n"),
        (".moai/index/a.json", "{\"a\":1}"),
        (".moai/receipts/r1.json", "{}"),
        (".moai-cli/worktrees/lane/x", "x"),
        (".moai-cli/reports/r.md", "# r\n"),
        (".moai-cli/backups/20261101T000000Z/x", "x"),
        (".moai-cli/notes.txt", "mine"),
    ] {
        assert_eq!(
            fs::read_to_string(fx.root.join(rel)).unwrap(),
            body,
            "{rel}"
        );
    }
    assert!(!fx.root.join(".moai-cli/project.yaml").exists());
    assert!(!fx.root.join(".moai-cli/index").exists());
    let (_, r) = receipt_of(&fx, &backup);
    let states: Vec<&str> = r["moves"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|m| m["kind"] == "state")
        .map(|m| m["path"].as_str().unwrap())
        .collect();
    assert_eq!(
        states,
        vec![
            ".moai-cli/index",
            ".moai-cli/project.yaml",
            ".moai-cli/receipts"
        ]
    );
    let rp = revert_plan(&fx, &backup);
    assert_eq!(item_action(&rp, ".moai-cli/index"), json!("restore"));
    let v = revert(&fx, &backup, rp["result"]["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], json!(true), "{v}");
    assert_eq!(project_tree(&fx), before);
}

/// An entry that came back at the earlier root after the apply is never overwritten: the
/// revert refuses before its first write.
#[test]
fn a_state_entry_that_came_back_refuses_the_revert() {
    let fx = movable_fixture();
    fx.write(".moai-cli/project.yaml", "schema_version: 2\n");
    let p = plan(&fx);
    let v = apply(&fx, p["plan_digest"].as_str().unwrap());
    let backup = v["result"]["backup"].as_str().unwrap().to_string();
    fx.write(".moai-cli/project.yaml", "written after the apply\n");
    let rp = revert_plan(&fx, &backup);
    assert_eq!(
        item_action(&rp, ".moai-cli/project.yaml"),
        json!("conflict")
    );
    let before = fx.tree();
    let v = revert(&fx, &backup, rp["result"]["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], json!(false), "{v}");
    assert_eq!(fx.tree(), before);
}

/// D2: an existing `.moai-backup` is never overwritten; the Go folder gets
/// `.moai-backup-<UTC>`, and after a revert the user's folder is still there and the project
/// is plain `go` again (not `interrupted`).
#[test]
fn a_taken_backup_name_gets_the_apply_stamp() {
    let fx = movable_fixture();
    fx.write(".moai-backup/mine.txt", "the user's own folder\n");
    let before = project_tree(&fx);
    let p = plan(&fx);
    assert_eq!(p["go_backup"], json!(".moai-backup-<UTC>"), "{p}");
    assert_eq!(
        restore_action(&p, ".moai-backup-<UTC>/specs/SPEC-1/spec.md"),
        "copy"
    );
    let v = apply(&fx, p["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], json!(true), "{v}");
    let backup = v["result"]["backup"].as_str().unwrap().to_string();
    let named = format!(".moai-backup-{backup}");
    assert_eq!(v["result"]["go_backup"], json!(named), "{v}");
    assert!(fx.root.join(&named).join("manifest.json").is_file());
    assert_eq!(
        fs::read_to_string(fx.root.join(".moai-backup/mine.txt")).unwrap(),
        "the user's own folder\n"
    );
    assert_eq!(
        fs::read_to_string(fx.root.join(".moai/specs/SPEC-1/spec.md")).unwrap(),
        "# spec\n"
    );
    let rp = revert_plan(&fx, &backup);
    let v = revert(&fx, &backup, rp["result"]["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], json!(true), "{v}");
    assert_eq!(project_tree(&fx), before);
    let again = plan(&fx);
    assert_eq!(again["go_backup"], json!(".moai-backup-<UTC>"), "{again}");
    assert!(moves(&again) > 0, "plain go again: {again}");
}

/// A product-state entry (or the earlier root itself) that is a link is never moved through:
/// the Go folder stays, with the reason; an earlier `backups/` link is left alone untouched.
#[test]
fn a_linked_state_entry_keeps_the_go_folder() {
    for linked in [".moai-cli/specs", ".moai-cli/project.yaml", ".moai-cli"] {
        let fx = movable_fixture();
        let elsewhere = fx.root.parent().unwrap().join("elsewhere");
        fs::create_dir_all(&elsewhere).unwrap();
        fs::create_dir_all(fx.root.join(".moai-cli")).unwrap();
        if linked == ".moai-cli" {
            fs::remove_dir(fx.root.join(".moai-cli")).unwrap();
        }
        symlink(&elsewhere, fx.root.join(linked)).unwrap();
        let p = plan(&fx);
        let (action, reason) = action(&p, ".moai");
        assert_eq!(action, "leave", "{linked}: {p}");
        assert!(reason.contains("Go folder stays"), "{linked}: {reason}");
        assert_eq!(p["go_backup"], Value::Null, "{linked}");
        assert_eq!(p["state_moves"], json!([]), "{linked}");
    }
    let fx = movable_fixture();
    let elsewhere = fx.root.parent().unwrap().join("elsewhere");
    fs::create_dir_all(&elsewhere).unwrap();
    fs::create_dir_all(fx.root.join(".moai-cli")).unwrap();
    symlink(&elsewhere, fx.root.join(".moai-cli/backups")).unwrap();
    let p = plan(&fx);
    assert_eq!(action(&p, ".moai").0, "move", "{p}");
    assert_eq!(
        state_move(&p, ".moai-cli/backups")["action"],
        json!("leave")
    );
    let v = apply(&fx, p["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], json!(true), "{v}");
    assert_eq!(fs::read_dir(&elsewhere).unwrap().count(), 0);
}

/// `CLAUDE.md` and `AGENTS.md` are copied aside byte for byte and stay as they are (the
/// user decides later, default `keep`); a linked one is not copied; a file changed after the
/// preview stops the apply before anything is written.
#[test]
fn root_instruction_files_are_copied_aside_and_kept() {
    let fx = movable_fixture();
    fx.write("CLAUDE.md", "# project rules\n");
    fs::write(fx.root.parent().unwrap().join("agents.md"), "# shared\n").unwrap();
    symlink(
        fx.root.parent().unwrap().join("agents.md"),
        fx.root.join("AGENTS.md"),
    )
    .unwrap();
    let p = plan(&fx);
    assert_eq!(
        p["root_files"],
        json!([
            {"path": "CLAUDE.md", "sha256": sha(b"# project rules\n"), "action": "copy",
             "backup_copy": ".moai/backups/<UTC>/root-files/CLAUDE.md", "user_choice": "keep",
             "reason": "copied aside; the file stays as it is until the user decides"},
            {"path": "AGENTS.md", "sha256": "", "action": "leave",
             "backup_copy": ".moai/backups/<UTC>/root-files/AGENTS.md", "user_choice": "keep",
             "reason": "a link; not copied, never followed"}
        ]),
        "{p}"
    );
    // Changed after the preview: the digest no longer matches, nothing is written.
    fx.write("CLAUDE.md", "# project rules, edited\n");
    let before = fx.tree();
    let v = apply(&fx, p["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], json!(false), "{v}");
    assert_eq!(fx.tree(), before);

    let p = plan(&fx);
    let v = apply(&fx, p["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], json!(true), "{v}");
    let backup = v["result"]["backup"].as_str().unwrap().to_string();
    let copy = format!(".moai/backups/{backup}/root-files/CLAUDE.md");
    assert_eq!(
        fs::read_to_string(fx.root.join(&copy)).unwrap(),
        "# project rules, edited\n"
    );
    assert_eq!(
        fs::read_to_string(fx.root.join("CLAUDE.md")).unwrap(),
        "# project rules, edited\n"
    );
    assert!(!area(&fx, &backup).join("root-files/AGENTS.md").exists());
    assert_eq!(
        v["result"]["copied"],
        json!([{"path": copy, "sha256": sha(b"# project rules, edited\n")}])
    );
}

/// Without a Go folder that moves, the preview carries none of the #2696 fields.
#[test]
fn a_staying_go_folder_plans_no_new_moai() {
    let mut fx = movable_fixture();
    let text = json!({"statusLine": {"command": ".moai/status_line.sh"}}).to_string();
    fx.write(".claude/settings.json", &text);
    fx.record(
        ".claude/settings.json",
        "user_created",
        &sha(text.as_bytes()),
    );
    fx.manifest();
    fx.write("CLAUDE.md", "# rules\n");
    let p = plan(&fx);
    assert_eq!(action(&p, ".moai").0, "leave");
    for key in ["state_moves", "root_files", "left_in_backup", "restores"] {
        assert_eq!(p[key], json!([]), "{key}: {p}");
    }
    assert_eq!(p["go_backup"], Value::Null);
}

// ---- #2696 C4a: the earlier `.moai-cli/` alone moves into a new `.moai/` ----

fn legacy_fixture() -> Fx {
    let fx = Fx::new();
    fx.write(".moai-cli/project.yaml", "schema_version: 2\n");
    fx.write(".moai-cli/index/a.json", "{\"a\":1}");
    fx.write(".moai-cli/worktrees/lane/x", "x");
    fx.write(".moai-cli/reports/r.md", "# r\n");
    fx.write(".claude/agents/mine.md", "# mine\n");
    fx
}

/// Preview → apply → re-run → revert: the state entries move into a new `.moai/` with the
/// marker last; worktrees and reports stay; nothing else in the project changes; the revert
/// brings the project back and sets the new `.moai/` aside (never deleted).
#[test]
fn an_earlier_root_project_moves_into_a_new_moai_and_back() {
    let fx = legacy_fixture();
    let before = project_tree(&fx);
    let p = plan(&fx);
    assert_eq!(
        action(&p, ".moai-cli"),
        (
            "move",
            "the earlier root's product state moves into a new .moai/ (#2696 C4a)"
        )
    );
    assert_eq!(moves(&p), 1, "{p}");
    assert_eq!(p["go_backup"], Value::Null, "{p}");
    assert_eq!(
        state_move(&p, ".moai-cli/project.yaml")["to"],
        json!(".moai/project.yaml")
    );
    assert_eq!(
        state_move(&p, ".moai-cli/worktrees")["action"],
        json!("leave")
    );
    let v = apply(&fx, p["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], json!(true), "{v}");
    let r = &v["result"];
    assert_eq!(r["state"], json!("committed"), "{r}");
    assert_eq!(r["legacy_move"], json!(true), "{r}");
    assert_eq!(r["go_backup"], Value::Null, "{r}");
    let backup = r["backup"].as_str().unwrap().to_string();
    let marker: Value =
        serde_json::from_slice(&fs::read(fx.root.join(".moai/moai-cli.json")).unwrap()).unwrap();
    assert_eq!(marker["generation"], json!("moai-cli"), "{marker}");
    for (rel, body) in [
        (".moai/project.yaml", "schema_version: 2\n"),
        (".moai/index/a.json", "{\"a\":1}"),
        (".moai-cli/worktrees/lane/x", "x"),
        (".moai-cli/reports/r.md", "# r\n"),
        (".claude/agents/mine.md", "# mine\n"),
    ] {
        assert_eq!(
            fs::read_to_string(fx.root.join(rel)).unwrap(),
            body,
            "{rel}"
        );
    }
    assert!(!fx.root.join(".moai-cli/project.yaml").exists());
    assert!(!fx.root.join(".moai-backup").exists());
    // Re-run: the project is `rust` now, nothing more moves.
    let again = plan(&fx);
    assert_eq!(moves(&again), 0, "{again}");
    // Revert.
    let rp = revert_plan(&fx, &backup);
    assert_eq!(
        item_action(&rp, ".moai/moai-cli.json"),
        json!("remove"),
        "{rp}"
    );
    assert_eq!(
        item_action(&rp, ".moai-cli/index"),
        json!("restore"),
        "{rp}"
    );
    assert_eq!(item_action(&rp, ".moai"), json!("set_aside"), "{rp}");
    let v = revert(&fx, &backup, rp["result"]["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], json!(true), "{v}");
    assert_eq!(project_tree(&fx), before);
    assert!(!fx.root.join(".moai").exists());
    let aside = rust_aside(&fx);
    let (path, receipt) = receipt_of(&fx, &backup);
    assert!(path.starts_with(&aside), "{path:?}");
    assert_eq!(receipt["state"], json!("reverted"), "{receipt}");
    // Back on the earlier root: the move is offered again.
    assert_eq!(moves(&plan(&fx)), 1);
}

/// A product-state entry that is a link keeps the earlier root where it is, with the reason;
/// a project without any product-state entry is offered nothing.
#[test]
fn an_earlier_root_with_a_linked_entry_or_no_state_stays() {
    let fx = legacy_fixture();
    let elsewhere = fx.root.parent().unwrap().join("elsewhere");
    fs::create_dir_all(&elsewhere).unwrap();
    symlink(&elsewhere, fx.root.join(".moai-cli/specs")).unwrap();
    let p = plan(&fx);
    let (act, reason) = action(&p, ".moai-cli");
    assert_eq!(act, "leave", "{p}");
    assert!(reason.contains("so it stays where it is"), "{reason}");
    assert_eq!(p["state_moves"], json!([]), "{p}");
    let before = fx.tree();
    apply(&fx, p["plan_digest"].as_str().unwrap());
    assert_eq!(fx.tree(), before);

    let fx = Fx::new();
    fx.write(".moai-cli/worktrees/lane/x", "x");
    let p = plan(&fx);
    assert_eq!(moves(&p), 0, "{p}");
    assert!(
        !p["items"].to_string().contains("\".moai-cli\""),
        "no item without product state: {p}"
    );
}

/// A failure midway leaves the new `.moai/` with a pending receipt and no marker:
/// `state_root` says `interrupted` (no Go backup is involved), further previews hold, and the
/// revert brings the project back.
#[test]
fn an_earlier_root_move_stopped_midway_is_interrupted_and_reverts() {
    use std::os::unix::fs::PermissionsExt;
    let fx = legacy_fixture();
    let before = project_tree(&fx);
    let p = plan(&fx);
    let locked = fx.root.join(".moai-cli");
    fs::set_permissions(&locked, fs::Permissions::from_mode(0o555)).unwrap();
    let v = apply(&fx, p["plan_digest"].as_str().unwrap());
    fs::set_permissions(&locked, fs::Permissions::from_mode(0o755)).unwrap();
    assert_eq!(v["ok"], json!(false), "{v}");
    assert!(!fx.root.join(".moai/moai-cli.json").exists());
    let again = plan(&fx);
    assert_eq!(moves(&again), 0, "{again}");
    assert!(
        again["warnings"]
            .to_string()
            .contains("state_root is `interrupted`"),
        "{again}"
    );
    let backup: String = fs::read_dir(fx.root.join(".moai/backups"))
        .unwrap()
        .map(|e| e.unwrap().file_name().into_string().unwrap())
        .next()
        .unwrap();
    let rp = revert_plan(&fx, &backup);
    assert_eq!(
        item_action(&rp, ".moai/moai-cli.json"),
        json!("skip"),
        "{rp}"
    );
    let v = revert(&fx, &backup, rp["result"]["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], json!(true), "{v}");
    assert_eq!(project_tree(&fx), before);
}

/// A C4a receipt cannot also claim a Go folder, a folder entry or copies.
#[test]
fn an_earlier_root_receipt_with_go_parts_is_refused() {
    let fx = legacy_fixture();
    let p = plan(&fx);
    let v = apply(&fx, p["plan_digest"].as_str().unwrap());
    let backup = v["result"]["backup"].as_str().unwrap().to_string();
    let (path, original) = receipt_of(&fx, &backup);
    type Forge = fn(&mut Value);
    let cases: Vec<(&str, Forge)> = vec![
        ("a Go backup too", |r: &mut Value| {
            r["go_backup"] = json!(".moai-backup")
        }),
        ("a folder entry", |r: &mut Value| {
            r["moves"]
                .as_array_mut()
                .unwrap()
                .push(json!({"path": ".moai", "sha256": "", "moved": true, "kind": "folder"}))
        }),
        (
            "copies",
            |r: &mut Value| r["copied"] = json!([{"path": format!(".moai/backups/{}/root-files/CLAUDE.md", r["backup"].as_str().unwrap()), "sha256": "sha256:00"}]),
        ),
        ("not a C4a receipt (the flag removed)", |r: &mut Value| {
            r.as_object_mut().unwrap().remove("legacy_move");
        }),
    ];
    for (name, edit) in cases {
        let mut forged = original.clone();
        edit(&mut forged);
        fs::write(&path, forged.to_string()).unwrap();
        let rp = revert_plan(&fx, &backup);
        assert_eq!(rp["ok"], json!(false), "{name}: {rp}");
    }
}
