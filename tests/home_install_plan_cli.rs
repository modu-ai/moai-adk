//! #2871 U3a: `home-install-plan`, the read-only preview of moving this generation's items
//! from `~/.moai-cli` to the `~/.moai` top level. Temporary HOME folders only: the real HOME is
//! never read or written, every case passes `--home <tempdir>`, secrets are synthetic names
//! with placeholder or no content.
#![cfg(unix)]

mod common;
use serde_json::{json, Value};
use std::fs;
use std::os::unix::fs::{symlink, MetadataExt, PermissionsExt};
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};

const BIN: &str = crate::common::MOAI_BIN;
const NOW: &str = "2026-11-14T12:00:00Z";
const LATER: &str = "2026-11-14T13:00:00Z";
const MARKER_V2: &str = r#"{"schema_version":2,"generation":"moai-cli"}"#;

fn write(root: &Path, rel: &str, body: &str) {
    let p = root.join(rel);
    fs::create_dir_all(p.parent().unwrap()).unwrap();
    fs::write(p, body).unwrap();
}

fn chmod(p: &Path, mode: u32) {
    fs::set_permissions(p, fs::Permissions::from_mode(mode)).unwrap();
}

/// A Rust-only HOME: every item this generation writes sits in `~/.moai-cli`, no `~/.moai`.
fn rust_home() -> (tempfile::TempDir, PathBuf) {
    let dir = tempfile::tempdir().unwrap();
    let home = fs::canonicalize(dir.path()).unwrap();
    for (rel, body) in [
        (".moai-cli/agents/a.json", "{}"),
        (".moai-cli/backups/b.txt", "x"),
        (".moai-cli/cleanup/c.json", "{}"),
        (".moai-cli/config/launcher.yaml", "a: 1\n"),
        (".moai-cli/home-routing/r.json", "{}"),
        (".moai-cli/home-style/s.json", "{}"),
        (".moai-cli/installations/i.json", "{}"),
        (".moai-cli/path/moai", "#!/bin/sh\n"),
        (".moai-cli/receipts/r.json", "{}"),
        (".moai-cli/schedule/s.json", "{}"),
        (".moai-cli/state/last-model.txt", "m\n"),
        (".moai-cli/vault/.master.key", "PLACEHOLDER-SECRET"),
        (".moai-cli/vault/anthropic", "PLACEHOLDER-SECRET"),
        (".moai-cli/versions/1.0.0/bin/moai-cli", "bin-1"),
        (".moai-cli/worktrees/projects/p/wt/README", "x\n"),
        (".moai-cli/projects/p/notes.md", "x\n"),
        (".moai-cli/bin/old-tool", "x"),
        (".moai-cli/mystery/m.txt", "x"),
        (".moai-cli/moai-cli.json", "{}"),
        (".moai-cli/home-install/keep.txt", "x"),
    ] {
        write(&home, rel, body);
    }
    symlink("agents", home.join(".moai-cli/link-x")).unwrap();
    (dir, home)
}

/// A Go `~/.moai` beside the Rust `~/.moai-cli`.
fn add_go_state(home: &Path) {
    write(
        home,
        ".moai/user-assets.json",
        r#"{"schema_version": 1, "files": {}}"#,
    );
    write(home, ".moai/state/active-sessions.json", "[]");
}

fn call(args: &[&str], envs: &[(&str, &str)]) -> Value {
    let out = Command::new(BIN)
        .args(args)
        .env_remove("MOAI_HOME")
        .envs(envs.iter().copied())
        .stdin(Stdio::null())
        .output()
        .unwrap();
    serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("{}", String::from_utf8_lossy(&out.stderr)))
}

fn plan_with(home: &Path, now: &str, envs: &[(&str, &str)]) -> Value {
    let v = call(
        &[
            "home-install-plan",
            "--home",
            home.to_str().unwrap(),
            "--now",
            now,
        ],
        envs,
    );
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn plan(home: &Path) -> Value {
    plan_with(home, NOW, &[])
}

fn entry<'a>(plan: &'a Value, name: &str) -> &'a Value {
    plan["entries"]
        .as_array()
        .unwrap()
        .iter()
        .find(|e| e["name"] == name)
        .unwrap_or_else(|| panic!("{name} missing: {plan}"))
}

fn act(plan: &Value, name: &str) -> (String, String) {
    let e = entry(plan, name);
    (
        e["action"].as_str().unwrap().to_string(),
        e["reason"].as_str().unwrap().to_string(),
    )
}

fn blocks(plan: &Value) -> Vec<String> {
    plan["blocks"]
        .as_array()
        .unwrap()
        .iter()
        .map(|b| b.as_str().unwrap().to_string())
        .collect()
}

/// Every path under `root`: kind, mode and, for a file, its bytes. A link is not followed.
fn tree(root: &Path) -> Vec<String> {
    fn walk(p: &Path, out: &mut Vec<String>) {
        let m = fs::symlink_metadata(p).unwrap();
        let mode = m.permissions().mode() & 0o7777;
        if m.file_type().is_symlink() {
            out.push(format!(
                "{} link {:o} -> {}",
                p.display(),
                mode,
                fs::read_link(p).unwrap().display()
            ));
        } else if m.is_dir() {
            out.push(format!("{} dir {mode:o}", p.display()));
            let mut names: Vec<_> = fs::read_dir(p)
                .unwrap()
                .map(|e| e.unwrap().path())
                .collect();
            names.sort();
            for n in names {
                walk(&n, out);
            }
        } else {
            let bytes = fs::read(p).unwrap_or_default();
            out.push(format!(
                "{} file {mode:o} {}",
                p.display(),
                String::from_utf8_lossy(&bytes)
            ));
        }
    }
    let mut out = Vec::new();
    walk(root, &mut out);
    out
}

const MOVED: [&str; 13] = [
    "agents",
    "backups",
    "cleanup",
    "config",
    "home-routing",
    "home-style",
    "installations",
    "path",
    "receipts",
    "schedule",
    "state",
    "vault",
    "versions",
];

#[test]
fn install_preview_writes_nothing() {
    let (_d, home) = rust_home();
    let before = tree(&home);
    let p = plan(&home);
    assert_eq!(p["writes"], json!(false), "{p}");
    assert_eq!(p["mode"], json!("install"), "{p}");
    assert_eq!(tree(&home), before);
    assert!(!home.join(".moai").exists());
}

#[test]
fn install_preview_follows_the_table() {
    let (_d, home) = rust_home();
    let p = plan(&home);
    assert_eq!(p["action"], json!("install"), "{p}");
    assert!(blocks(&p).is_empty(), "{p}");
    for n in MOVED {
        assert_eq!(act(&p, n), ("move".into(), "rust_item".into()), "{n}: {p}");
        assert_eq!(
            entry(&p, n)["destination"],
            json!(home.join(".moai").join(n).display().to_string()),
            "{n}"
        );
    }
    for (n, why) in [
        ("worktrees", "worktrees_stay"),
        ("projects", "user_data"),
        ("bin", "not_written_by_this_generation"),
        ("mystery", "not_in_table"),
        ("moai-cli.json", "marker_written_last"),
        ("home-install", "install_journal"),
        ("link-x", "link"),
    ] {
        assert_eq!(act(&p, n), ("keep".into(), why.into()), "{n}: {p}");
        assert_eq!(entry(&p, n)["destination"], Value::Null, "{n}");
    }
    assert_eq!(p["counts"]["move"], json!(13), "{p}");
    assert_eq!(p["counts"]["keep"], json!(7), "{p}");
    assert_eq!(p["counts"]["blocked"], json!(0), "{p}");
}

/// Codex review of #3019: the user's hook rules (`hook-config.toml`) move with the HOME, since
/// after the marker the loader reads only `~/.moai`. The folder `config/` moves as before.
#[test]
fn the_hook_config_file_moves_with_the_home() {
    let (_d, home) = rust_home();
    write(&home, ".moai-cli/hook-config.toml", "[hooks]\n");
    let p = plan(&home);
    assert_eq!(p["action"], json!("install"), "{p}");
    assert_eq!(
        act(&p, "hook-config.toml"),
        ("move".into(), "rust_item".into()),
        "{p}"
    );
    assert_eq!(
        entry(&p, "hook-config.toml")["destination"],
        json!(home.join(".moai/hook-config.toml").display().to_string()),
        "{p}"
    );
    assert_eq!(p["counts"]["move"], json!(14), "{p}");
}

#[test]
fn schedule_is_a_rust_table_row() {
    let row = moai_cli::home_layout::lookup("schedule").expect("a table row for `schedule`");
    assert_eq!(row.rust, moai_cli::home_layout::RustUse::Writes);
    assert!(!row.go);
    assert_eq!(row.class, moai_cli::home_layout::Class::Rust);
    assert_eq!(row.rule, moai_cli::home_layout::Rule::RustName);
}

#[test]
fn install_blocks_a_case_only_clash() {
    let (_d, home) = rust_home();
    add_go_state(&home);
    write(&home, ".moai/Agents/x", "x");
    let p = plan(&home);
    assert_eq!(p["action"], json!("blocked"), "{p}");
    assert!(
        blocks(&p).contains(&"destination_taken:agents".to_string()),
        "{p}"
    );
    assert_eq!(
        act(&p, "agents"),
        ("blocked".into(), "destination_taken".into())
    );
}

#[test]
fn install_blocks_on_a_go_state_still_in_place() {
    let (_d, home) = rust_home();
    add_go_state(&home);
    let p = plan(&home);
    assert_eq!(p["action"], json!("blocked"), "{p}");
    assert!(
        blocks(&p).contains(&"destination_taken:state".to_string()),
        "{p}"
    );
    assert_eq!(
        act(&p, "state"),
        ("blocked".into(), "destination_taken".into())
    );
    // The Go items have not retired yet: the user's order is retire first, then install.
    assert!(
        blocks(&p)
            .iter()
            .any(|b| b.starts_with("go_items_not_retired:")),
        "{p}"
    );
    // The names that are free still show as a move.
    assert_eq!(act(&p, "vault").0, "move");
}

#[test]
fn install_entry_limit_is_500() {
    for (count, blocked) in [(500usize, false), (501, true)] {
        let dir = tempfile::tempdir().unwrap();
        let home = fs::canonicalize(dir.path()).unwrap();
        for i in 0..count {
            fs::create_dir_all(home.join(format!(".moai-cli/n{i:04}"))).unwrap();
        }
        let p = plan(&home);
        assert_eq!(
            blocks(&p).contains(&"not_fully_listed".to_string()),
            blocked,
            "{count}: {p}"
        );
        assert_eq!(p["total"], json!(count), "{count}");
    }
}

#[test]
fn install_refuses_links_and_never_follows() {
    // `~/.moai` is a link.
    let (_d, home) = rust_home();
    let other = tempfile::tempdir().unwrap();
    symlink(other.path(), home.join(".moai")).unwrap();
    let p = plan(&home);
    assert_eq!(p["action"], json!("blocked"), "{p}");
    assert!(blocks(&p).contains(&"generation_refuse".to_string()), "{p}");

    // `~/.moai-cli` is a link: nothing behind it is listed.
    let dir = tempfile::tempdir().unwrap();
    let home = fs::canonicalize(dir.path()).unwrap();
    let behind = tempfile::tempdir().unwrap();
    write(behind.path(), "agents/a.json", "{}");
    symlink(behind.path(), home.join(".moai-cli")).unwrap();
    let p = plan(&home);
    assert_eq!(p["action"], json!("blocked"), "{p}");
    assert!(
        blocks(&p).contains(&"state_dir_not_a_folder".to_string()),
        "{p}"
    );
    assert_eq!(p["entries"], json!([]), "{p}");
}

#[test]
fn install_never_opens_a_secret() {
    let (_d, home) = rust_home();
    for f in ["vault/.master.key", "vault/anthropic"] {
        chmod(&home.join(".moai-cli").join(f), 0o000);
    }
    let out = Command::new(BIN)
        .args([
            "home-install-plan",
            "--home",
            home.to_str().unwrap(),
            "--now",
            NOW,
        ])
        .env_remove("MOAI_HOME")
        .stdin(Stdio::null())
        .output()
        .unwrap();
    let text = String::from_utf8_lossy(&out.stdout).to_string();
    assert!(!text.contains("PLACEHOLDER-SECRET"), "{text}");
    let v: Value = serde_json::from_str(&text).unwrap();
    let p = &v["result"];
    assert_eq!(p["action"], json!("install"), "{p}");
    assert_eq!(p["vault"]["kind"], json!("folder"), "{p}");
    assert_eq!(p["vault"]["master_key_mode"], json!("000"), "{p}");
    // Put the modes back so the tempdir can be removed.
    for f in ["vault/.master.key", "vault/anthropic"] {
        chmod(&home.join(".moai-cli").join(f), 0o600);
    }
}

#[test]
fn install_digest_binds_identity() {
    let (d, home) = rust_home();
    let first = plan(&home);
    // Swap `agents` for a new folder of the same name while the old one is still alive, so the
    // new one cannot reuse its inode.
    let parked = d.path().join("parked-agents");
    fs::rename(home.join(".moai-cli/agents"), &parked).unwrap();
    write(&home, ".moai-cli/agents/a.json", "{}");
    let second = plan(&home);
    assert_ne!(first["content_digest"], second["content_digest"]);
    assert_ne!(first["plan_digest"], second["plan_digest"]);
    assert_ne!(first["confirm_token"], second["confirm_token"]);
}

#[test]
fn install_blocks_an_interrupted_home() {
    let (_d, home) = rust_home();
    write(
        &home,
        ".moai/backups/20261114T100000Z/go-state-receipt.json",
        "{}",
    );
    let p = plan(&home);
    assert_eq!(p["action"], json!("blocked"), "{p}");
    assert!(
        blocks(&p).contains(&"generation_interrupted".to_string()),
        "{p}"
    );
    assert_eq!(p["confirm_token"], Value::Null, "{p}");
}

#[test]
fn install_preview_is_deterministic() {
    let (_d, home) = rust_home();
    let a = plan_with(&home, NOW, &[]);
    let b = plan_with(&home, LATER, &[]);
    assert_eq!(a["content_digest"], b["content_digest"]);
    assert_eq!(a["confirm_token"], b["confirm_token"]);
    assert_ne!(a["plan_digest"], b["plan_digest"]);
    assert_eq!(plan_with(&home, NOW, &[])["plan_digest"], a["plan_digest"]);
}

#[test]
fn install_none_when_nothing_to_move() {
    // No `~/.moai-cli` at all, then an empty one.
    let dir = tempfile::tempdir().unwrap();
    let home = fs::canonicalize(dir.path()).unwrap();
    let p = plan(&home);
    assert_eq!(p["action"], json!("none"), "{p}");
    assert!(blocks(&p).is_empty(), "{p}");
    assert_eq!(p["confirm_token"], Value::Null, "{p}");
    fs::create_dir(home.join(".moai-cli")).unwrap();
    let p = plan(&home);
    assert_eq!(p["action"], json!("none"), "{p}");
    assert_eq!(p["counts"]["move"], json!(0), "{p}");
}

#[test]
fn install_reports_late_arrivals_only() {
    // The marker is already schema 2: the move happened. An older build wrote back into
    // `~/.moai-cli`; the preview names those entries and moves nothing.
    let dir = tempfile::tempdir().unwrap();
    let home = fs::canonicalize(dir.path()).unwrap();
    write(&home, ".moai/moai-cli.json", MARKER_V2);
    write(&home, ".moai/agents/a.json", "{}");
    write(&home, ".moai-cli/state/x.txt", "x");
    write(&home, ".moai-cli/mystery/m.txt", "x");
    // The install journal folder is not an entry that came back.
    write(&home, ".moai-cli/home-install/j.txt", "x");
    let p = plan(&home);
    assert_eq!(p["action"], json!("none"), "{p}");
    assert_eq!(p["counts"]["move"], json!(0), "{p}");
    let late: Vec<String> = p["late_arrivals"]
        .as_array()
        .unwrap()
        .iter()
        .map(|e| e["name"].as_str().unwrap().to_string())
        .collect();
    assert_eq!(
        late,
        vec!["mystery".to_string(), "state".to_string()],
        "{p}"
    );
    assert_eq!(p["confirm_token"], Value::Null, "{p}");
}

#[test]
fn install_reports_vault_modes_without_changing_them() {
    let (_d, home) = rust_home();
    let vault = home.join(".moai-cli/vault");
    chmod(&vault, 0o755);
    chmod(&vault.join(".master.key"), 0o644);
    let before = tree(&home);
    let p = plan(&home);
    assert_eq!(p["vault"]["mode"], json!("755"), "{p}");
    assert_eq!(p["vault"]["master_key_mode"], json!("644"), "{p}");
    assert_eq!(p["vault"]["loose"], json!(true), "{p}");
    assert_eq!(p["vault"]["entries"], json!(2), "{p}");
    assert_eq!(tree(&home), before);
    // Tight modes are not loose.
    chmod(&vault, 0o700);
    chmod(&vault.join(".master.key"), 0o600);
    assert_eq!(plan(&home)["vault"]["loose"], json!(false));
}

#[test]
fn alias_table_maps_every_version_folder() {
    let (_d, home) = rust_home();
    write(&home, ".moai-cli/versions/2.0.0/bin/moai-cli", "bin-2");
    write(&home, ".moai-cli/versions/3.0.0/readme", "no binary here");
    let p = plan(&home);
    let id = |path: PathBuf| {
        let d = moai_cli::digest(path.to_str().unwrap().as_bytes());
        d.strip_prefix("sha256:").unwrap().to_string()
    };
    let table = p["alias_table"].as_object().unwrap();
    assert_eq!(table.len(), 2, "{p}");
    for v in ["1.0.0", "2.0.0"] {
        let old = id(home.join(format!(".moai-cli/versions/{v}/bin/moai-cli")));
        let new = id(home.join(format!(".moai/versions/{v}/bin/moai-cli")));
        assert_eq!(table[&old], json!(new), "{v}: {p}");
    }
    assert_eq!(p["alias_skipped"], json!(["3.0.0"]), "{p}");
}

#[test]
fn install_token_has_its_own_prefix() {
    let (_d, home) = rust_home();
    let p = plan(&home);
    let t = p["confirm_token"].as_str().unwrap();
    let hex = t.strip_prefix("install:").unwrap_or_else(|| panic!("{t}"));
    assert_eq!(hex.len(), 16, "{t}");
    assert!(hex.chars().all(|c| c.is_ascii_hexdigit()), "{t}");
    assert!(!t.starts_with("home:"));
}

#[test]
fn install_plans_the_path_stub() {
    let (_d, home) = rust_home();
    let version_bin = home.join(".moai-cli/versions/1.0.0/bin/moai-cli");
    symlink(&version_bin, home.join(".moai-cli/path/moai-cli")).unwrap();
    let p = plan(&home);
    let stub = p["stub"].as_array().unwrap();
    let find = |n: &str| stub.iter().find(|e| e["name"] == n).unwrap().clone();
    assert_eq!(find("moai-cli")["kind"], json!("link"), "{p}");
    assert_eq!(find("moai-cli")["repoint"], json!(true), "{p}");
    assert_eq!(
        find("moai-cli")["target"],
        json!(version_bin.display().to_string()),
        "{p}"
    );
    assert_eq!(find("moai")["kind"], json!("file"), "{p}");
    assert_eq!(find("moai")["repoint"], json!(false), "{p}");
}

// #2871 U3c-pre (plan 5.2): a link into the old `versions` folder is re-pointed only when
// `versions` and `path` both move. A `versions` that stays (here a link) leaves the old target
// valid, and the new one would not exist.
#[test]
fn install_does_not_repoint_a_stub_when_versions_stays() {
    let (d, home) = rust_home();
    fs::create_dir_all(home.join(".moai-cli/versions/1.0.0")).unwrap();
    let version_bin = home.join(".moai-cli/versions/1.0.0/bin/moai-cli");
    symlink(&version_bin, home.join(".moai-cli/path/moai-cli")).unwrap();
    let repoint = |p: &Value| {
        p["stub"]
            .as_array()
            .unwrap()
            .iter()
            .find(|e| e["name"] == "moai-cli")
            .unwrap()["repoint"]
            .clone()
    };
    // Control: both move, so the link is re-pointed.
    let moving = plan(&home);
    assert_eq!(act(&moving, "versions").0, "move", "{moving}");
    assert_eq!(repoint(&moving), json!(true), "{moving}");
    // `versions` becomes a link: it stays where it is.
    let parked = d.path().join("parked-versions");
    fs::rename(home.join(".moai-cli/versions"), &parked).unwrap();
    symlink(&parked, home.join(".moai-cli/versions")).unwrap();
    let staying = plan(&home);
    assert_eq!(act(&staying, "versions").0, "keep", "{staying}");
    assert_eq!(repoint(&staying), json!(false), "{staying}");
}

// #3001 audit note N1 (mutant J3): when `versions` moves but `path` does not (here its destination
// is taken, which blocks the plan), the stub must not be re-pointed either: the apply would leave
// a `path` entry aimed at a `versions` that is no longer where the link says.
#[test]
fn install_does_not_repoint_a_stub_when_path_stays() {
    let (_d, home) = rust_home();
    add_go_state(&home);
    write(&home, ".moai/path/x", "x");
    let version_bin = home.join(".moai-cli/versions/1.0.0/bin/moai-cli");
    symlink(&version_bin, home.join(".moai-cli/path/moai-cli")).unwrap();
    let p = plan(&home);
    assert_eq!(act(&p, "versions").0, "move", "{p}");
    assert_eq!(
        act(&p, "path"),
        ("blocked".into(), "destination_taken".into()),
        "{p}"
    );
    let stub = p["stub"].as_array().unwrap();
    let link = stub.iter().find(|e| e["name"] == "moai-cli").unwrap();
    assert_eq!(link["repoint"], json!(false), "{p}");
}

// Codex P2 4576239282 on #3001: the table lookup ignores ASCII case, so a `Versions` folder is a
// Rust name that moves, and the stub plan must agree (on a case-insensitive volume such as default
// APFS the link below reaches it by the lowercase spelling).
#[test]
fn install_repoints_a_stub_when_versions_moves_under_another_case() {
    let (_d, home) = rust_home();
    let version_bin = home.join(".moai-cli/versions/1.0.0/bin/moai-cli");
    symlink(&version_bin, home.join(".moai-cli/path/moai-cli")).unwrap();
    fs::rename(
        home.join(".moai-cli/versions"),
        home.join(".moai-cli/Versions"),
    )
    .unwrap();
    let p = plan(&home);
    assert_eq!(act(&p, "Versions").0, "move", "{p}");
    let stub = p["stub"].as_array().unwrap();
    let link = stub.iter().find(|e| e["name"] == "moai-cli").unwrap();
    assert_eq!(link["repoint"], json!(true), "{p}");
}

// #2871 U3c-pre (plan 5.1): the apply creates the install journal folder (for its lock and its
// manifest) before it re-plans; that arrival must not change the digest it compares. The folder
// is still listed, as kept.
#[test]
fn install_digest_ignores_the_journal_folder_the_apply_creates() {
    let (_d, home) = rust_home();
    // The fixture already holds a journal folder; start from a HOME without one.
    fs::remove_dir_all(home.join(".moai-cli/home-install")).unwrap();
    let before = plan(&home);
    assert!(
        before["entries"]
            .as_array()
            .unwrap()
            .iter()
            .all(|e| e["name"] != "home-install"),
        "{before}"
    );
    fs::create_dir_all(home.join(".moai-cli/home-install")).unwrap();
    let after = plan(&home);
    assert_eq!(
        act(&after, "home-install"),
        ("keep".into(), "install_journal".into()),
        "{after}"
    );
    assert_eq!(before["content_digest"], after["content_digest"], "{after}");
    assert_eq!(before["confirm_token"], after["confirm_token"], "{after}");
    // The digest still binds every other entry: a new top-level name changes it.
    write(&home, ".moai-cli/zz-new/x.json", "{}");
    let changed = plan(&home);
    assert_ne!(
        after["content_digest"], changed["content_digest"],
        "{changed}"
    );
}

#[test]
fn install_blocks_when_moai_home_points_elsewhere() {
    let (_d, home) = rust_home();
    let p = plan_with(&home, NOW, &[("MOAI_HOME", "/somewhere/else")]);
    assert_eq!(p["action"], json!("blocked"), "{p}");
    assert!(
        blocks(&p).contains(&"moai_home_elsewhere".to_string()),
        "{p}"
    );
}

#[test]
fn install_blocks_a_taken_journal_stamp() {
    let (_d, home) = rust_home();
    write(
        &home,
        ".moai-cli/home-install/move-journal-20261114T120000Z.json",
        "{}",
    );
    let p = plan(&home);
    assert!(blocks(&p).contains(&"stamp_taken".to_string()), "{p}");
    assert_eq!(plan_with(&home, LATER, &[])["action"], json!("install"));
}

#[test]
fn install_counts_a_multi_linked_vault_entry_without_opening_it() {
    // A hard link among the vault entries is reported as a count, never followed or opened.
    let (_d, home) = rust_home();
    fs::hard_link(
        home.join(".moai-cli/vault/anthropic"),
        home.join(".moai-cli/vault/anthropic-copy"),
    )
    .unwrap();
    assert!(
        fs::metadata(home.join(".moai-cli/vault/anthropic"))
            .unwrap()
            .nlink()
            >= 2
    );
    let p = plan(&home);
    assert_eq!(p["vault"]["entries"], json!(3), "{p}");
    assert_eq!(p["vault"]["multi_link"], json!(2), "{p}");
}

#[test]
fn install_blocks_a_case_only_clash_from_the_other_side() {
    // The old folder holds `Vault`, the new one already has `vault`: the same name on the
    // default file system, so the move would land on it.
    let dir = tempfile::tempdir().unwrap();
    let home = fs::canonicalize(dir.path()).unwrap();
    write(&home, ".moai-cli/Vault/.master.key", "x");
    write(
        &home,
        ".moai/user-assets.json",
        r#"{"schema_version": 1, "files": {}}"#,
    );
    write(&home, ".moai/vault/other", "x");
    let p = plan(&home);
    assert!(
        blocks(&p).contains(&"destination_taken:Vault".to_string()),
        "{p}"
    );
    assert_eq!(
        act(&p, "Vault"),
        ("blocked".into(), "destination_taken".into())
    );
}

#[test]
fn install_repoints_only_links_into_the_old_versions_folder() {
    let (_d, home) = rust_home();
    symlink("/usr/bin/true", home.join(".moai-cli/path/elsewhere")).unwrap();
    let p = plan(&home);
    let stub = p["stub"].as_array().unwrap();
    let e = stub.iter().find(|e| e["name"] == "elsewhere").unwrap();
    assert_eq!(e["kind"], json!("link"), "{p}");
    assert_eq!(e["target"], json!("/usr/bin/true"), "{p}");
    assert_eq!(e["repoint"], json!(false), "{p}");
}

#[test]
fn install_calls_a_group_writable_vault_loose() {
    let (_d, home) = rust_home();
    let vault = home.join(".moai-cli/vault");
    chmod(&vault, 0o750);
    chmod(&vault.join(".master.key"), 0o600);
    assert_eq!(plan(&home)["vault"]["loose"], json!(true));
    chmod(&vault, 0o700);
    chmod(&vault.join(".master.key"), 0o640);
    assert_eq!(plan(&home)["vault"]["loose"], json!(true));
}

// Codex review P2 on #2987: a version folder or its `bin` that is a link would make the canonical
// path (old identity) differ from the moved path (new identity); such a folder gets no alias.
#[test]
fn alias_table_skips_a_version_folder_or_bin_that_is_a_link() {
    let (_d, home) = rust_home();
    let outside = tempfile::tempdir().unwrap();
    let outside = fs::canonicalize(outside.path()).unwrap();
    write(&outside, "folder/bin/moai-cli", "external-1");
    write(&outside, "bin/moai-cli", "external-2");
    symlink(
        outside.join("folder"),
        home.join(".moai-cli/versions/9.0.0"),
    )
    .unwrap();
    fs::create_dir_all(home.join(".moai-cli/versions/8.0.0")).unwrap();
    symlink(
        outside.join("bin"),
        home.join(".moai-cli/versions/8.0.0/bin"),
    )
    .unwrap();
    let p = plan(&home);
    let table = p["alias_table"].as_object().unwrap();
    assert_eq!(table.len(), 1, "only 1.0.0 is a real folder: {p}");
    assert_eq!(p["alias_skipped"], json!(["8.0.0", "9.0.0"]), "{p}");
}

// Codex review P2 on #2987: a folder that cannot be listed is not an empty one.
#[test]
fn install_blocks_a_state_folder_it_cannot_list() {
    let (_d, home) = rust_home();
    let old = home.join(".moai-cli");
    chmod(&old, 0o000);
    let listable = fs::read_dir(&old).is_ok();
    let p = plan(&home);
    chmod(&old, 0o755);
    if listable {
        return; // running as a user the mode does not bind (root): nothing to show
    }
    assert_eq!(p["action"], json!("blocked"), "{p}");
    assert!(
        blocks(&p).contains(&"state_dir_unreadable".to_string()),
        "{p}"
    );
    assert!(
        p["confirm_token"].is_null(),
        "no token for a blocked plan: {p}"
    );
}

// Codex review P2 on #2987: the retirement preview lists at most 500 names. When it is blocked
// for that, "zero pending" proves nothing, so the install preview is blocked too.
#[test]
fn install_blocks_when_the_go_retirement_preview_is_incomplete() {
    let (_d, home) = rust_home();
    add_go_state(&home);
    for i in 0..501 {
        fs::create_dir_all(home.join(format!(".moai/a{i:04}"))).unwrap();
    }
    let p = plan(&home);
    assert_eq!(p["action"], json!("blocked"), "{p}");
    assert!(
        blocks(&p).contains(&"go_retire_preview_blocked".to_string()),
        "{p}"
    );
    assert!(p["confirm_token"].is_null(), "{p}");
}

// Audit 6883839420 note 1 (needed before U3c): the preview never opens a vault entry. A FIFO in
// place of the master key makes any open for reading wait forever, so a read shows up as a hang
// that this test turns into a failure (a mode-000 file would only fail the read silently).
#[test]
fn install_does_not_open_a_master_key_that_is_a_fifo() {
    use std::io::Read as _;
    let (_d, home) = rust_home();
    let key = home.join(".moai-cli/vault/.master.key");
    fs::remove_file(&key).unwrap();
    assert!(Command::new("mkfifo").arg(&key).status().unwrap().success());
    let mut child = Command::new(BIN)
        .args([
            "home-install-plan",
            "--home",
            home.to_str().unwrap(),
            "--now",
            NOW,
        ])
        .env_remove("MOAI_HOME")
        .stdin(Stdio::null())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    let start = std::time::Instant::now();
    loop {
        if child.try_wait().unwrap().is_some() {
            break;
        }
        if start.elapsed() > std::time::Duration::from_secs(20) {
            let _ = child.kill();
            let _ = child.wait();
            panic!("the preview waited on the FIFO: it opened the master key");
        }
        std::thread::sleep(std::time::Duration::from_millis(50));
    }
    let mut text = String::new();
    child
        .stdout
        .take()
        .unwrap()
        .read_to_string(&mut text)
        .unwrap();
    let v: Value = serde_json::from_str(&text).unwrap();
    let p = &v["result"];
    assert_eq!(p["vault"]["kind"], json!("folder"), "{p}");
    assert!(p["vault"]["master_key_mode"].is_string(), "{p}");
}

// Audit note 1: modes and link counts are the entry's own (lstat), never the target's.
#[test]
fn install_reports_a_linked_vault_entry_by_its_own_mode_and_link_count() {
    let (_d, home) = rust_home();
    let outside = tempfile::tempdir().unwrap();
    let outside = fs::canonicalize(outside.path()).unwrap();
    write(&outside, "key", "OUTSIDE-SECRET");
    chmod(&outside.join("key"), 0o600);
    fs::hard_link(outside.join("key"), outside.join("key-twin")).unwrap();
    let vault = home.join(".moai-cli/vault");
    fs::remove_file(vault.join(".master.key")).unwrap();
    symlink(outside.join("key"), vault.join(".master.key")).unwrap();
    symlink(outside.join("key"), vault.join("ext")).unwrap();
    let p = plan(&home);
    // The target is mode 0600 with two names; neither fact may be taken from it.
    assert_ne!(p["vault"]["master_key_mode"], json!("600"), "{p}");
    assert_eq!(p["vault"]["links"], json!(2), "{p}");
    // Only `anthropic` (a plain file with one name) is left besides the two links: no multi-name.
    assert_eq!(p["vault"]["multi_link"], json!(0), "{p}");
}

// Audit note 5: a folder's link count grows with its sub-folders; that is not a second name.
#[test]
fn install_does_not_call_a_vault_subfolder_a_multi_named_entry() {
    let (_d, home) = rust_home();
    fs::create_dir_all(home.join(".moai-cli/vault/sub/deeper")).unwrap();
    let p = plan(&home);
    assert_eq!(p["vault"]["multi_link"], json!(0), "{p}");
}

// Audit note 2: the destination is asked of the file system too, so a volume that folds more
// than ASCII case still reports the clash. Needs a volume that folds U+017F to `s`; without one
// there is nothing to show (the listing compare alone is the ASCII-case rule).
#[test]
fn install_asks_the_file_system_about_a_folded_destination_name() {
    let (_d, home) = rust_home();
    add_go_state(&home);
    write(&home, ".moai/\u{17f}chedule/x", "x");
    if fs::symlink_metadata(home.join(".moai/schedule")).is_err() {
        return;
    }
    let p = plan(&home);
    assert_eq!(
        act(&p, "schedule"),
        ("blocked".into(), "destination_taken".into()),
        "{p}"
    );
}

// Audit note 4: the `path` stub plan (50) and the version alias table (500) are bounded lists the
// later steps rely on; one name more than the bound blocks instead of cutting the list silently.
#[test]
fn install_blocks_when_the_stub_or_alias_list_would_be_cut() {
    for (extra, blocked) in [(49usize, false), (50, true)] {
        let (_d, home) = rust_home();
        for i in 0..extra {
            write(&home, &format!(".moai-cli/path/p{i:03}"), "x");
        }
        let p = plan(&home);
        assert_eq!(
            blocks(&p).contains(&"path_not_fully_listed".to_string()),
            blocked,
            "{extra}: {p}"
        );
    }
    for (extra, blocked) in [(499usize, false), (500, true)] {
        let (_d, home) = rust_home();
        for i in 0..extra {
            fs::create_dir_all(home.join(format!(".moai-cli/versions/v{i:04}"))).unwrap();
        }
        let p = plan(&home);
        assert_eq!(
            blocks(&p).contains(&"versions_not_fully_listed".to_string()),
            blocked,
            "{extra}: blocks {:?}",
            blocks(&p)
        );
    }
}

// #2989 audit item 1: a `path` or `versions` that is a link (or a plain file) is not a list the
// later steps walk, so the names behind the link must not trip the bound. The same number of names
// in a real folder does block, so the count below is a real one.
#[test]
fn install_does_not_count_a_linked_list_folder_against_its_bound() {
    for (sub, many, block) in [
        ("path", 60usize, "path_not_fully_listed"),
        ("versions", 600, "versions_not_fully_listed"),
    ] {
        let (_d, home) = rust_home();
        for i in 0..many {
            fs::create_dir_all(home.join(format!("elsewhere/{sub}/e{i:04}"))).unwrap();
        }
        let list = home.join(format!(".moai-cli/{sub}"));

        // Control: the names in a real folder block.
        let _ = fs::remove_dir_all(&list);
        fs::create_dir_all(list.parent().unwrap()).unwrap();
        for i in 0..many {
            fs::create_dir_all(list.join(format!("e{i:04}"))).unwrap();
        }
        let p = plan(&home);
        assert!(blocks(&p).contains(&block.to_string()), "{sub} real: {p}");

        // The same names behind a link do not.
        fs::remove_dir_all(&list).unwrap();
        symlink(home.join(format!("elsewhere/{sub}")), &list).unwrap();
        let p = plan(&home);
        assert!(!blocks(&p).contains(&block.to_string()), "{sub} link: {p}");

        // Nor behind a plain file.
        fs::remove_file(&list).unwrap();
        fs::write(&list, "x").unwrap();
        let p = plan(&home);
        assert!(!blocks(&p).contains(&block.to_string()), "{sub} file: {p}");
    }
}
