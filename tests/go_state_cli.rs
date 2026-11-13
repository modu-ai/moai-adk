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

const EXE: &str = env!("CARGO_BIN_EXE_moai-cli");

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

/// The project tree without the backup area: AC-3 compares the project as the user sees
/// it; the backup folder and its receipt stay after a revert as the record of both acts.
fn project_tree(fx: &Fx) -> Vec<(PathBuf, Vec<u8>)> {
    let backups = fx.root.join(".moai-cli/backups");
    fx.tree()
        .into_iter()
        .filter(|(p, _)| !p.starts_with(&backups))
        .collect()
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

    let v = apply(&fx, p["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], json!(true), "{v}");
    let r = &v["result"];
    let backup = r["backup"].as_str().unwrap().to_string();
    assert_eq!(r["state"], json!("committed"), "{r}");
    assert_eq!(r["permission_reference"], json!(REF));
    let dir = fx.root.join(".moai-cli/backups").join(&backup);
    for (rel, body) in [
        (".moai/config/sections/a.yaml", "a: 1\n"),
        (".claude/workflows/w.md", "# w\n"),
    ] {
        assert!(!fx.root.join(rel).exists(), "{rel} moved away");
        assert_eq!(
            fs::read_to_string(dir.join("go-state").join(rel)).unwrap(),
            body,
            "{rel} kept byte for byte in the backup"
        );
    }
    // Decision 1: the whole `.moai/` folder went to the backup in one rename, changed,
    // unrecorded and manifest files included, byte for byte.
    assert!(!fx.root.join(".moai").exists(), ".moai moved away");
    for (rel, body) in [
        (".moai/config/sections/b.yaml", "b: 2 # edited\n"),
        (".moai/specs/SPEC-1/spec.md", "# spec\n"),
    ] {
        assert_eq!(
            fs::read_to_string(dir.join("go-state").join(rel)).unwrap(),
            body,
            "{rel}"
        );
    }
    assert!(dir.join("go-state/.moai/manifest.json").is_file());
    let receipt: Value =
        serde_json::from_slice(&fs::read(dir.join("go-state-receipt.json")).unwrap()).unwrap();
    assert_eq!(receipt["state"], json!("committed"), "{receipt}");
    let folder = receipt["moves"]
        .as_array()
        .unwrap()
        .iter()
        .find(|m| m["path"] == ".moai")
        .unwrap()
        .clone();
    assert_eq!(folder["kind"], json!("folder"), "{receipt}");
    assert_eq!(folder["moved"], json!(true));
    assert!(folder["sha256"].as_str().unwrap().starts_with("sha256:"));
    assert_eq!(
        receipt["moves"].as_array().unwrap().last().unwrap()["path"],
        json!(".moai"),
        "the folder moves last"
    );
    assert_eq!(receipt["plan_digest"], p["plan_digest"]);

    // AC-2: the preview after the apply offers nothing more, and a second apply is a no-op
    // that creates no new backup folder.
    let again = plan(&fx);
    assert_eq!(moves(&again), 0, "{again}");
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

#[test]
fn a_linked_backup_area_is_refused() {
    for linked in [".moai-cli", ".moai-cli/backups"] {
        let fx = movable_fixture();
        let elsewhere = fx.root.parent().unwrap().join("elsewhere");
        fs::create_dir_all(&elsewhere).unwrap();
        let link = fx.root.join(linked);
        fs::create_dir_all(link.parent().unwrap()).unwrap();
        symlink(&elsewhere, &link).unwrap();
        let p = plan(&fx);
        let before = fx.tree();
        let v = apply(&fx, p["plan_digest"].as_str().unwrap());
        assert_eq!(v["ok"], json!(false), "{linked}: {v}");
        assert!(v.to_string().contains("symlink"), "{linked}: {v}");
        assert_eq!(fx.tree(), before, "{linked}");
        assert_eq!(fs::read_dir(&elsewhere).unwrap().count(), 0, "{linked}");
    }
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
    // AC-3: the project tree outside the backup area is what it was before the apply.
    assert_eq!(project_tree(&fx), before);
    let receipt: Value = serde_json::from_slice(
        &fs::read(
            fx.root
                .join(".moai-cli/backups")
                .join(&backup)
                .join("go-state-receipt.json"),
        )
        .unwrap(),
    )
    .unwrap();
    assert_eq!(receipt["state"], json!("reverted"), "{receipt}");
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
        &format!(".moai-cli/backups/{backup}/go-state/.claude/workflows/w.md"),
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

/// A failure midway leaves a `pending` receipt that names what already moved, and the
/// revert brings exactly those files back.
#[test]
fn a_failure_midway_leaves_a_pending_receipt_the_revert_can_use() {
    use std::os::unix::fs::PermissionsExt;
    let fx = movable_fixture();
    let before = project_tree(&fx);
    let p = plan(&fx);
    // Files move first, then `.moai/` in one rename. A folder that cannot be written
    // cannot change its parent (`..`), so locking `.moai` stops only the folder move.
    let locked = fx.root.join(".moai");
    fs::set_permissions(&locked, fs::Permissions::from_mode(0o555)).unwrap();
    let v = apply(&fx, p["plan_digest"].as_str().unwrap());
    fs::set_permissions(&locked, fs::Permissions::from_mode(0o755)).unwrap();
    assert_eq!(v["ok"], json!(false), "{v}");
    let backups = fx.root.join(".moai-cli/backups");
    let mut stamps: Vec<String> = fs::read_dir(&backups)
        .unwrap()
        .map(|e| e.unwrap().file_name().into_string().unwrap())
        .collect();
    assert_eq!(stamps.len(), 1);
    let backup = stamps.pop().unwrap();
    let receipt: Value = serde_json::from_slice(
        &fs::read(backups.join(&backup).join("go-state-receipt.json")).unwrap(),
    )
    .unwrap();
    assert_eq!(receipt["state"], json!("pending"), "{receipt}");
    let moved: Vec<(&str, bool)> = receipt["moves"]
        .as_array()
        .unwrap()
        .iter()
        .map(|m| (m["path"].as_str().unwrap(), m["moved"].as_bool().unwrap()))
        .collect();
    assert_eq!(
        moved,
        vec![(".claude/workflows/w.md", true), (".moai", false)]
    );
    assert!(fx.root.join(".moai/config/sections/a.yaml").is_file());

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
    let dir = fx.root.join(".moai-cli/backups").join(&backup);
    let receipt_path = dir.join("go-state-receipt.json");
    let original: Value = serde_json::from_slice(&fs::read(&receipt_path).unwrap()).unwrap();
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
            r["moves"][0]["path"] = json!(".claude/../src/evil.rs")
        }),
        ("a path outside .moai/ and .claude/", |r: &mut Value| {
            r["moves"][0]["path"] = json!("src/evil.rs");
            r["moves"][0]["sha256"] = json!(sha(b"evil\n"));
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

fn receipt_of(fx: &Fx, backup: &str) -> (PathBuf, Value) {
    let path = fx
        .root
        .join(".moai-cli/backups")
        .join(backup)
        .join("go-state-receipt.json");
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

/// A live hook recreates `.moai/` after the apply: the revert never merges into it.
#[test]
fn a_moai_folder_that_came_back_refuses_the_revert() {
    let fx = movable_fixture();
    let p = plan(&fx);
    let v = apply(&fx, p["plan_digest"].as_str().unwrap());
    let backup = v["result"]["backup"].as_str().unwrap().to_string();
    fx.write(".moai/logs/hook.log", "written by a live hook\n");
    let rp = revert_plan(&fx, &backup);
    let item = rp["result"]["items"]
        .as_array()
        .unwrap()
        .iter()
        .find(|i| i["path"] == ".moai")
        .unwrap()
        .clone();
    assert_eq!(item["action"], json!("conflict"), "{rp}");
    let before = fx.tree();
    let v = revert(&fx, &backup, rp["result"]["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], json!(false), "{v}");
    assert!(v.to_string().contains(".moai is conflict"), "{v}");
    assert_eq!(fx.tree(), before);
}

/// The folder's backup is checked against the tree digest the apply recorded.
#[test]
fn a_changed_folder_backup_refuses_the_revert() {
    let fx = movable_fixture();
    let p = plan(&fx);
    let v = apply(&fx, p["plan_digest"].as_str().unwrap());
    let backup = v["result"]["backup"].as_str().unwrap().to_string();
    fx.write(
        &format!(".moai-cli/backups/{backup}/go-state/.moai/specs/SPEC-1/spec.md"),
        "# tampered\n",
    );
    let rp = revert_plan(&fx, &backup);
    let item = rp["result"]["items"]
        .as_array()
        .unwrap()
        .iter()
        .find(|i| i["path"] == ".moai")
        .unwrap()
        .clone();
    assert_eq!(item["action"], json!("missing"), "{rp}");
    let before = fx.tree();
    let v = revert(&fx, &backup, rp["result"]["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], json!(false), "{v}");
    assert_eq!(fx.tree(), before);
}

/// The folder was renamed but the receipt write after it did not happen (the receipt
/// still says `moved: false`, no digest): the disk decides, and the revert brings it back.
#[test]
fn a_folder_move_the_receipt_missed_is_still_restored() {
    let fx = movable_fixture();
    let before = project_tree(&fx);
    let p = plan(&fx);
    let v = apply(&fx, p["plan_digest"].as_str().unwrap());
    let backup = v["result"]["backup"].as_str().unwrap().to_string();
    let (path, mut r) = receipt_of(&fx, &backup);
    for m in r["moves"].as_array_mut().unwrap() {
        if m["path"] == ".moai" {
            m["moved"] = json!(false);
            m["sha256"] = json!("");
        }
    }
    r["state"] = json!("pending");
    fs::write(&path, r.to_string()).unwrap();
    let rp = revert_plan(&fx, &backup);
    let v = revert(&fx, &backup, rp["result"]["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], json!(true), "{v}");
    assert_eq!(project_tree(&fx), before);
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
            r["moves"][0]["kind"] = json!("folder")
        }),
        ("file kind on the folder", |r: &mut Value| {
            let last = r["moves"].as_array().unwrap().len() - 1;
            r["moves"][last]["kind"] = json!("");
        }),
        ("unknown kind", |r: &mut Value| {
            r["moves"][0]["kind"] = json!("tree")
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

// ---- stage 3b: restore `.moai/specs` into `.moai-cli/specs` (AC-5) ----

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

/// AC-5: the SPEC is copied to `.moai-cli/specs` from the backup and the original stays in
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
    assert_eq!(restore_action(&p, ".moai/specs/SPEC-1/spec.md"), "copy");
    assert_eq!(
        restore_action(&p, ".moai/specs/SPEC-2/spec.md"),
        "excluded_secret"
    );
    let v = apply(&fx, p["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], json!(true), "{v}");
    let backup = v["result"]["backup"].as_str().unwrap().to_string();
    assert_eq!(
        fs::read_to_string(fx.root.join(".moai-cli/specs/SPEC-1/spec.md")).unwrap(),
        "# spec\n"
    );
    assert!(!fx.root.join(".moai-cli/specs/SPEC-2").exists());
    let dir = fx.root.join(".moai-cli/backups").join(&backup);
    assert!(dir.join("go-state/.moai/specs/SPEC-1/spec.md").is_file());
    assert!(dir.join("go-state/.moai/specs/SPEC-2/spec.md").is_file());
    let (_, r) = receipt_of(&fx, &backup);
    assert_eq!(
        r["restored"],
        json!([{"path": ".moai-cli/specs/SPEC-1/spec.md", "sha256": sha(b"# spec\n")}])
    );
}

/// An existing target is never overwritten: the preview says `conflict` and the apply
/// writes nothing at all.
#[test]
fn an_existing_restore_target_stops_the_apply() {
    let fx = movable_fixture();
    fx.write(".moai-cli/specs/SPEC-1/spec.md", "# mine, already here\n");
    let p = plan(&fx);
    assert_eq!(restore_action(&p, ".moai/specs/SPEC-1/spec.md"), "conflict");
    let before = fx.tree();
    let v = apply(&fx, p["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], json!(false), "{v}");
    assert!(v.to_string().contains("never overwritten"), "{v}");
    assert_eq!(fx.tree(), before);
}

/// The revert removes an unchanged restored copy and keeps (and reports) a changed one.
#[test]
fn revert_removes_only_unchanged_restored_copies() {
    let fx = movable_fixture();
    fx.write(".moai/specs/SPEC-3/spec.md", "# three\n");
    let p = plan(&fx);
    let v = apply(&fx, p["plan_digest"].as_str().unwrap());
    let backup = v["result"]["backup"].as_str().unwrap().to_string();
    fx.write(
        ".moai-cli/specs/SPEC-3/spec.md",
        "# three, edited after the apply\n",
    );
    let rp = revert_plan(&fx, &backup);
    let items = rp["result"]["items"].as_array().unwrap().clone();
    let act = |path: &str| {
        items
            .iter()
            .find(|i| i["path"] == path)
            .unwrap_or_else(|| panic!("{path}: {rp}"))["action"]
            .clone()
    };
    assert_eq!(
        act(".moai-cli/specs/SPEC-1/spec.md"),
        json!("remove_restored")
    );
    assert_eq!(act(".moai-cli/specs/SPEC-3/spec.md"), json!("keep_changed"));
    let v = revert(&fx, &backup, rp["result"]["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], json!(true), "{v}");
    assert!(!fx.root.join(".moai-cli/specs/SPEC-1/spec.md").exists());
    assert_eq!(
        fs::read_to_string(fx.root.join(".moai-cli/specs/SPEC-3/spec.md")).unwrap(),
        "# three, edited after the apply\n"
    );
    assert!(fx.root.join(".moai/specs/SPEC-3/spec.md").is_file());
}

/// A receipt may only name restored copies under `.moai-cli/specs/`.
#[test]
fn a_receipt_restored_path_elsewhere_is_refused() {
    let fx = movable_fixture();
    let p = plan(&fx);
    let v = apply(&fx, p["plan_digest"].as_str().unwrap());
    let backup = v["result"]["backup"].as_str().unwrap().to_string();
    let (path, mut r) = receipt_of(&fx, &backup);
    r["restored"][0]["path"] = json!("src/main.rs");
    fs::write(&path, r.to_string()).unwrap();
    let rp = revert_plan(&fx, &backup);
    assert_eq!(rp["ok"], json!(false), "{rp}");
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
