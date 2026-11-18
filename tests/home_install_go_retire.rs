//! SPEC-20261118-home-install-go-retire (card C-50): the install preview after the Go retire,
//! the refusal that stays without a retire record, one answer for the ledger registration, and
//! the reads after the move. Temporary HOME folders only: the real HOME is never read or
//! written, every case passes `--home <tempdir>`, secrets are synthetic placeholders.
#![cfg(unix)]

mod common;
use serde_json::{json, Value};
use std::fs;
use std::io::Write as _;
use std::os::unix::fs::symlink;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};

const BIN: &str = crate::common::MOAI_BIN;
const RETIRE_NOW: &str = "2026-11-14T12:00:00Z";
const RETIRE_STAMP: &str = "20261114T120000Z";
const NOW: &str = "2026-11-14T13:00:00Z";
/// One setup step of a refusal case, and one row of a tree listing.
type Setup = Box<dyn Fn(&Fx)>;
type Row = (String, String, Vec<u8>);
const PERM: &str = "user approved in the fixture";

fn write(root: &Path, rel: &str, body: &str) {
    let p = root.join(rel);
    fs::create_dir_all(p.parent().unwrap()).unwrap();
    fs::write(p, body).unwrap();
}

/// A temporary HOME and a temporary git project that carries a project profile (the install
/// reads its session registry, the ledger registration reads its project id).
struct Fx {
    _home: tempfile::TempDir,
    home: PathBuf,
    _project: tempfile::TempDir,
    project: PathBuf,
}

/// This generation's items in `~/.moai-cli`: the vault, a version with its binary, the `path`
/// stub linking into it, and an item that is not in the name table.
fn write_rust_items(home: &Path) {
    for (rel, body) in [
        (".moai-cli/config/launcher.yaml", "a: 1\n"),
        (".moai-cli/state/last-model.txt", "m\n"),
        (
            ".moai-cli/vault/.master.key",
            "PLACEHOLDER-KEY-0123456789-ABCDE",
        ),
        (".moai-cli/vault/anthropic", "PLACEHOLDER-SECRET"),
        (".moai-cli/versions/1.0.0/bin/moai-cli", "bin-1"),
        (".moai-cli/path/moai", "#!/bin/sh\n"),
        (".moai-cli/mystery/m.txt", "x"),
    ] {
        write(home, rel, body);
    }
    // The key store refuses a vault other users can read.
    use std::os::unix::fs::PermissionsExt as _;
    fs::set_permissions(
        home.join(".moai-cli/vault"),
        fs::Permissions::from_mode(0o700),
    )
    .unwrap();
    for n in [".master.key", "anthropic"] {
        let k = home.join(".moai-cli/vault").join(n);
        fs::set_permissions(k, fs::Permissions::from_mode(0o600)).unwrap();
    }
    symlink(
        home.join(".moai-cli/versions/1.0.0/bin/moai-cli"),
        home.join(".moai-cli/path/moai-cli"),
    )
    .unwrap();
}

fn fx(go_state: bool) -> Fx {
    let dir = tempfile::tempdir().unwrap();
    let home = fs::canonicalize(dir.path()).unwrap();
    write_rust_items(&home);
    if go_state {
        // Items the retire takes, and items it keeps (user data, unknown owner, unknown name).
        write(&home, ".moai/state/active-sessions.json", "[]");
        write(&home, ".moai/user-assets.json", "{}");
        write(&home, ".moai/cache/a.bin", "x");
        write(&home, ".moai/specs/SPEC-1/spec.md", "# user spec\n");
        write(&home, ".moai/reports/r.md", "# report\n");
        write(&home, ".moai/my notes/n.txt", "kept\n");
    } else {
        // An `~/.moai` the table does not know: unknown, with no Go marker.
        write(&home, ".moai/my notes/n.txt", "mine\n");
        write(&home, ".moai/reports/r.md", "# report\n");
    }
    let pdir = tempfile::tempdir().unwrap();
    let project = fs::canonicalize(pdir.path()).unwrap();
    common::git::git_ok(&project, &["init", "-q"]);
    write(
        &project,
        ".moai-cli/project.yaml",
        "{\"schema_version\":2,\"project_id\":\"demo\"}\n",
    );
    Fx {
        _home: dir,
        home,
        _project: pdir,
        project,
    }
}

fn call(args: &[&str]) -> Value {
    let out = Command::new(BIN)
        .args(args)
        .env_remove("MOAI_HOME")
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .stdin(Stdio::null())
        .output()
        .unwrap();
    serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("{}", String::from_utf8_lossy(&out.stderr)))
}

fn result(v: Value) -> Value {
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

/// The Go retire of the fixture: preview, apply, and the answer.
fn retire(f: &Fx) {
    let h = f.home.to_str().unwrap();
    let p = result(call(&[
        "home-retire-plan",
        "--home",
        h,
        "--now",
        RETIRE_NOW,
    ]));
    assert_eq!(p["action"], json!("retire"), "{p}");
    let a = call(&[
        "home-retire-apply",
        "--home",
        h,
        "--now",
        RETIRE_NOW,
        "--plan-digest",
        p["plan_digest"].as_str().unwrap(),
        "--permission-reference",
        PERM,
    ]);
    let r = result(a);
    assert_eq!(r["journal_state"], json!("committed"), "{r}");
}

fn install_plan(f: &Fx) -> Value {
    result(call(&[
        "home-install-plan",
        "--home",
        f.home.to_str().unwrap(),
        "--now",
        NOW,
        "--project",
        f.project.to_str().unwrap(),
    ]))
}

fn install_apply(f: &Fx, digest: &str) -> Value {
    call(&[
        "home-install-apply",
        "--home",
        f.home.to_str().unwrap(),
        "--project",
        f.project.to_str().unwrap(),
        "--now",
        NOW,
        "--plan-digest",
        digest,
        "--permission-reference",
        PERM,
    ])
}

fn ledger_plan(f: &Fx) -> Value {
    result(call(&[
        "ledger-register-plan",
        f.project.to_str().unwrap(),
        "--home",
        f.home.to_str().unwrap(),
    ]))
}

fn blocks(p: &Value) -> Vec<String> {
    p["blocks"]
        .as_array()
        .unwrap()
        .iter()
        .map(|b| b.as_str().unwrap().to_string())
        .collect()
}

/// Every entry under `root`: relative path, kind and, for a file, its bytes; links by target.
fn tree(root: &Path) -> Vec<Row> {
    fn walk(dir: &Path, rel: &str, out: &mut Vec<Row>) {
        let mut names: Vec<_> = fs::read_dir(dir)
            .unwrap()
            .map(|e| e.unwrap().file_name())
            .collect();
        names.sort();
        for n in names {
            let p = dir.join(&n);
            let r = format!("{rel}/{}", n.to_string_lossy());
            let m = fs::symlink_metadata(&p).unwrap();
            if m.file_type().is_symlink() {
                let t = fs::read_link(&p).unwrap();
                out.push((r, "link".into(), t.to_string_lossy().as_bytes().to_vec()));
            } else if m.is_dir() {
                out.push((r.clone(), "dir".into(), Vec::new()));
                walk(&p, &r, out);
            } else {
                out.push((r, "file".into(), fs::read(&p).unwrap()));
            }
        }
    }
    let mut out = Vec::new();
    walk(root, "", &mut out);
    out
}

fn generation(f: &Fx) -> &'static str {
    moai_cli::state_root::classify_home(&f.home).generation
}

#[test]
fn a_retired_go_home_gets_a_movable_install_preview_and_the_move_reverts() {
    let f = fx(true);
    assert_eq!(generation(&f), "go");
    retire(&f);
    // The retire took every Go marker away; what stays has no generation marker.
    assert_eq!(generation(&f), "unknown");
    let before_old = tree(&f.home.join(".moai-cli"));

    let p = install_plan(&f);
    assert_eq!(p["action"], json!("install"), "{p}");
    assert_eq!(blocks(&p), Vec::<String>::new(), "{p}");
    assert!(p["confirm_token"].as_str().unwrap().starts_with("install:"));
    assert_eq!(p["writes"], json!(false));
    assert_eq!(
        tree(&f.home.join(".moai-cli")),
        before_old,
        "the preview wrote"
    );

    let a = install_apply(&f, p["plan_digest"].as_str().unwrap());
    let r = result(a);
    assert_eq!(r["state"], json!("applied"), "{r}");
    // The Rust items moved, the marker is there, the kept entries are untouched.
    for n in ["vault", "versions", "path", "config", "state"] {
        assert!(
            fs::symlink_metadata(f.home.join(".moai").join(n)).is_ok(),
            "{n}"
        );
        assert!(
            fs::symlink_metadata(f.home.join(".moai-cli").join(n)).is_err(),
            "{n} stayed behind"
        );
    }
    assert_eq!(generation(&f), "rust-first");
    assert!(f.home.join(".moai/my notes/n.txt").is_file());
    assert!(f.home.join(".moai/specs/SPEC-1/spec.md").is_file());

    let rv = result(call(&[
        "home-install-revert",
        "--home",
        f.home.to_str().unwrap(),
        "--permission-reference",
        PERM,
    ]));
    assert!(rv.is_object(), "{rv}");
    for n in ["vault", "versions", "path", "config", "state"] {
        assert!(
            fs::symlink_metadata(f.home.join(".moai-cli").join(n)).is_ok(),
            "{n} not restored"
        );
    }
    assert_eq!(generation(&f), "unknown");
}

#[test]
fn an_unknown_home_without_a_retire_record_is_still_refused() {
    let manifest = |kind: &str, home: &str| {
        json!({"schema_version": 1, "kind": kind, "stamp": RETIRE_STAMP, "home": home, "items": [], "kept": []})
            .to_string()
    };
    let cases: Vec<(&str, Setup)> = vec![
        ("no record", Box::new(|_f: &Fx| {})),
        (
            "another home",
            Box::new(move |f: &Fx| {
                write(
                    &f.home,
                    &format!(".moai-go-retired/retire-manifest-{RETIRE_STAMP}.json"),
                    &manifest("go-retire", "/somewhere/else"),
                );
            }),
        ),
        (
            "another kind",
            Box::new(move |f: &Fx| {
                write(
                    &f.home,
                    &format!(".moai-go-retired/retire-manifest-{RETIRE_STAMP}.json"),
                    &manifest("home-install", f.home.to_str().unwrap()),
                );
            }),
        ),
        (
            "a link",
            Box::new(move |f: &Fx| {
                let real = f.home.join("elsewhere.json");
                fs::write(&real, manifest("go-retire", f.home.to_str().unwrap())).unwrap();
                fs::create_dir_all(f.home.join(".moai-go-retired")).unwrap();
                symlink(
                    &real,
                    f.home.join(format!(
                        ".moai-go-retired/retire-manifest-{RETIRE_STAMP}.json"
                    )),
                )
                .unwrap();
            }),
        ),
        (
            "not a manifest",
            Box::new(|f: &Fx| {
                write(
                    &f.home,
                    &format!(".moai-go-retired/retire-manifest-{RETIRE_STAMP}.json"),
                    "this is not json",
                );
            }),
        ),
    ];
    for (label, setup) in cases {
        let f = fx(false);
        setup(&f);
        assert_eq!(generation(&f), "unknown", "{label}");
        let before = tree(&f.home);
        let p = install_plan(&f);
        assert_eq!(p["action"], json!("blocked"), "{label}: {p}");
        assert!(
            blocks(&p).contains(&"generation_unknown".to_string()),
            "{label}: {p}"
        );
        assert!(p["confirm_token"].is_null(), "{label}: {p}");
        let help = p["help"]
            .as_array()
            .unwrap_or_else(|| panic!("{label}: {p}"));
        assert!(
            help.iter()
                .any(|h| h.as_str().is_some_and(|t| t.contains("retire"))),
            "{label}: the help line names no way out: {p}"
        );
        assert_eq!(tree(&f.home), before, "{label}: the preview wrote");
    }
}

#[test]
fn after_the_install_readers_resolve_under_moai_and_moai_cli_stays_untouched() {
    let f = fx(true);
    retire(&f);
    let home = f.home.clone();
    let run = |args: &[&str], stdin: Option<&str>| {
        let mut child = Command::new(BIN)
            .args(args)
            .arg("--shared")
            .current_dir(&home)
            .env("HOME", &home)
            .env("GIT_CEILING_DIRECTORIES", &home)
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
            .unwrap();
        {
            let mut pipe = child.stdin.take().unwrap();
            if let Some(t) = stdin {
                pipe.write_all(t.as_bytes()).unwrap();
            }
        }
        let out = child.wait_with_output().unwrap();
        serde_json::from_slice::<Value>(&out.stdout)
            .unwrap_or_else(|_| panic!("{}", String::from_utf8_lossy(&out.stderr)))
    };
    let key = format!("moai-test-key-{}", std::process::id());
    let set = run(&["key-set", "glm"], Some(&key));
    assert_eq!(set["result"]["written"], json!(true), "{set}");

    let p = install_plan(&f);
    let r = result(install_apply(&f, p["plan_digest"].as_str().unwrap()));
    assert_eq!(r["state"], json!("applied"), "{r}");

    assert_eq!(moai_cli::home_root::dir(&f.home), ".moai");
    let status = run(&["key-status", "glm"], None);
    assert_eq!(status["result"]["entry"], json!("present"), "{status}");
    // The `path` stub follows the move: its link points into the new versions folder.
    let target = fs::read_link(f.home.join(".moai/path/moai-cli")).unwrap();
    assert!(
        target.starts_with(f.home.join(".moai/versions")),
        "{}",
        target.display()
    );
    // Reading again, and previewing again, writes nothing into the earlier folder.
    let old = f.home.join(".moai-cli");
    let before = tree(&old);
    let _ = run(&["key-status", "glm"], None);
    let again = install_plan(&f);
    assert_eq!(again["action"], json!("none"), "{again}");
    assert_eq!(tree(&old), before, "something wrote into ~/.moai-cli");
}

#[test]
fn ledger_register_and_install_agree_on_home_readiness() {
    // With a finished retire on record: both offer.
    let f = fx(true);
    retire(&f);
    let i = install_plan(&f);
    let l = ledger_plan(&f);
    assert_eq!(i["action"], json!("install"), "{i}");
    assert_eq!(l["action"], json!("register"), "{l}");
    // Once installed: the install has nothing left, the registration is offered.
    result(install_apply(&f, i["plan_digest"].as_str().unwrap()));
    assert_eq!(install_plan(&f)["action"], json!("none"));
    assert_eq!(ledger_plan(&f)["action"], json!("register"));

    // An unknown `~/.moai` with no record: both refuse, the ledger says why.
    let g = fx(false);
    let i = install_plan(&g);
    assert!(
        blocks(&i).contains(&"generation_unknown".to_string()),
        "{i}"
    );
    let l = ledger_plan(&g);
    assert_eq!(l["action"], json!("blocked"), "{l}");
    let before = tree(&g.home);
    let _ = ledger_plan(&g);
    assert_eq!(tree(&g.home), before, "the preview wrote");
}

/// The journal of the retire the fixture ran.
fn retire_journal(f: &Fx) -> PathBuf {
    let root = f.home.join(".moai-go-retired");
    fs::read_dir(&root)
        .unwrap()
        .map(|e| e.unwrap().path())
        .find(|p| {
            p.file_name()
                .is_some_and(|n| n.to_string_lossy().starts_with("move-journal-"))
        })
        .unwrap_or_else(|| panic!("no retire journal in {}", root.display()))
}

#[test]
fn a_retire_that_did_not_finish_does_not_unlock_the_install() {
    // Codex review P2 on #3199: the manifest is written before the first rename, so it is not
    // the proof; the journal beside it must be committed with every item moved.
    for (label, edit) in [
        (
            "pending",
            ("\"state\": \"committed\"", "\"state\": \"pending\""),
        ),
        (
            "reverted",
            ("\"state\": \"committed\"", "\"state\": \"reverted\""),
        ),
        (
            "an item not moved",
            ("\"status\": \"moved\"", "\"status\": \"skipped\""),
        ),
    ] {
        let f = fx(true);
        retire(&f);
        let journal = retire_journal(&f);
        let text = fs::read_to_string(&journal).unwrap();
        assert!(text.contains(edit.0), "{label}: {text}");
        fs::write(&journal, text.replacen(edit.0, edit.1, 1)).unwrap();
        assert_eq!(generation(&f), "unknown", "{label}");
        let p = install_plan(&f);
        assert_eq!(p["action"], json!("blocked"), "{label}: {p}");
        assert!(
            blocks(&p).contains(&"generation_unknown".to_string()),
            "{label}: {p}"
        );
    }
    // And with the journal gone.
    let f = fx(true);
    retire(&f);
    fs::remove_file(retire_journal(&f)).unwrap();
    let p = install_plan(&f);
    assert_eq!(p["action"], json!("blocked"), "{p}");
}

fn ledger_apply(f: &Fx, digest: &str) -> Value {
    call(&[
        "ledger-register-apply",
        f.project.to_str().unwrap(),
        "--home",
        f.home.to_str().unwrap(),
        "--plan-digest",
        digest,
        "--permission-reference",
        PERM,
    ])
}

#[test]
fn a_blocked_registration_digest_cannot_authorize_after_the_home_turns_ready() {
    // Codex review P2 on #3199: the blocked preview advertises no writes, so its digest must
    // not stay valid for the registration once the unknown entries are gone.
    let f = fx(false);
    let l = ledger_plan(&f);
    assert_eq!(l["action"], json!("blocked"), "{l}");
    let digest = l["plan_digest"].as_str().unwrap().to_string();
    fs::remove_dir_all(f.home.join(".moai/my notes")).unwrap();
    fs::remove_dir_all(f.home.join(".moai/reports")).unwrap();
    let a = ledger_apply(&f, &digest);
    assert_eq!(a["ok"], json!(false), "{a}");
    assert!(
        a["error"]["message"].as_str().unwrap().contains("stale"),
        "{a}"
    );
    assert!(!f.home.join(".moai/projects").exists(), "a ledger was made");
}

#[test]
fn the_registration_apply_refuses_a_blocked_home_and_writes_nothing() {
    // Audit N2 (M1) on #3199: the apply itself refuses, not only the preview.
    let f = fx(false);
    let l = ledger_plan(&f);
    assert_eq!(l["action"], json!("blocked"), "{l}");
    let before = tree(&f.home);
    let a = ledger_apply(&f, l["plan_digest"].as_str().unwrap());
    assert_eq!(a["ok"], json!(false), "{a}");
    assert!(
        a["error"]["message"]
            .as_str()
            .unwrap()
            .contains("nothing was written"),
        "{a}"
    );
    assert_eq!(tree(&f.home), before);
}

#[test]
fn an_existing_ledger_keeps_its_idempotent_registration() {
    // Codex review P2 on #3199: the new readiness guard applies only where the registration
    // would create the ledger.
    let f = fx(false);
    fs::remove_dir_all(f.home.join(".moai/my notes")).unwrap();
    fs::remove_dir_all(f.home.join(".moai/reports")).unwrap();
    let first = ledger_plan(&f);
    assert_eq!(first["action"], json!("register"), "{first}");
    let a = result(ledger_apply(&f, first["plan_digest"].as_str().unwrap()));
    assert!(a.is_object(), "{a}");
    // A HOME that holds only the ledger's own folder is the ledger's doing: still ready.
    assert_eq!(generation(&f), "unknown");
    assert_eq!(ledger_plan(&f)["action"], json!("none"));
    // Another entry appears without a generation marker or a retire record.
    write(&f.home, ".moai/logs/x.log", "x\n");
    assert_eq!(generation(&f), "unknown");
    let again = ledger_plan(&f);
    assert_eq!(again["action"], json!("none"), "{again}");
    let b = result(ledger_apply(&f, again["plan_digest"].as_str().unwrap()));
    assert_eq!(b["state"], json!("unchanged"), "{b}");
}

#[test]
fn a_manifest_that_names_its_home_as_a_relative_path_is_not_evidence() {
    // Audit N4 on #3199: `"home": "."` would match whatever folder the command runs in.
    let f = fx(true);
    retire(&f);
    let root = f.home.join(".moai-go-retired");
    let manifest = fs::read_dir(&root)
        .unwrap()
        .map(|e| e.unwrap().path())
        .find(|p| {
            p.file_name()
                .is_some_and(|n| n.to_string_lossy().starts_with("retire-manifest-"))
        })
        .unwrap();
    let mut v: Value = serde_json::from_slice(&fs::read(&manifest).unwrap()).unwrap();
    v["home"] = json!(".");
    fs::write(&manifest, serde_json::to_vec(&v).unwrap()).unwrap();
    let out = Command::new(BIN)
        .args([
            "home-install-plan",
            "--home",
            f.home.to_str().unwrap(),
            "--now",
            NOW,
            "--project",
            f.project.to_str().unwrap(),
        ])
        .current_dir(&f.home)
        .env_remove("MOAI_HOME")
        .stdin(Stdio::null())
        .output()
        .unwrap();
    let p = result(serde_json::from_slice(&out.stdout).unwrap());
    assert_eq!(p["action"], json!("blocked"), "{p}");
    assert!(
        blocks(&p).contains(&"generation_unknown".to_string()),
        "{p}"
    );
}
