//! #2871 U3b (R5): an install id is a hash of the binary's canonical path, so moving the version
//! folders from `~/.moai-cli` to `~/.moai` changes it while the receipts keep the id they were
//! written under. The committed install journal records old id to new id; the readers
//! (`resolve_host_receipt` for the update baseline, `install-inspect` for self) use it. Without
//! a committed journal nothing changes. Every case runs copies of the built binary inside a
//! temporary HOME; the real HOME is never touched and the move is done by the journal module
//! itself, not by a command of this build (the apply step is a later piece).

#![cfg(unix)]

mod clock_wait;
mod common;
use common::CleanEnv;

use moai_cli::install::{plan_receipt, record_receipt, InstallerOwner};
use moai_cli::move_journal::{never_in_use, MoveEnv, PlannedMove};
use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::{
    collections::BTreeMap,
    fs,
    io::Write as _,
    path::{Path, PathBuf},
    process::{Command, Output, Stdio},
};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;
const HOSTS: [&str; 2] = ["claude", "codex"];
const NOW: &str = "2026-11-14T10:00:00Z";

struct Fixture {
    _tmp: TempDir,
    home: PathBuf,
    v1: PathBuf,
    v2: PathBuf,
}

fn skill_md(body: &str) -> String {
    format!("---\nname: moai\ndescription: fixture package\n---\n# MoAI\n{body}\n")
}

fn write_pkg(dir: &Path, files: &[(&str, String)]) {
    for (rel, body) in files {
        let p = dir.join(rel);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        fs::write(p, body).unwrap();
    }
}

fn fixture() -> Fixture {
    let tmp = TempDir::new().unwrap();
    let root = fs::canonicalize(tmp.path()).unwrap();
    let home = root.join("home");
    fs::create_dir_all(&home).unwrap();
    let v1 = root.join("stage-v1");
    let v2 = root.join("stage-v2");
    write_pkg(
        &v1,
        &[
            ("SKILL.md", skill_md("body v1")),
            ("workflows/update.md", "# UPDATE\nv1\n".into()),
        ],
    );
    write_pkg(
        &v2,
        &[
            ("SKILL.md", skill_md("body v2")),
            ("workflows/update.md", "# UPDATE\nv2\n".into()),
        ],
    );
    Fixture {
        _tmp: tmp,
        home,
        v1,
        v2,
    }
}

/// A copy of the built binary at `<state>/versions/<version>/bin/moai-cli`.
fn version_exe(state: &Path, version: &str) -> PathBuf {
    let bin = state.join("versions").join(version).join("bin");
    fs::create_dir_all(&bin).unwrap();
    let exe = bin.join("moai-cli");
    common::copy_executable(Path::new(EXE), &exe);
    exe
}

fn run(exe: &Path, home: &Path, args: &[&str], stdin: Option<&Value>) -> Output {
    let mut child = Command::new(exe)
        .args(args)
        .clean_env()
        .env("PATH", "/usr/bin:/bin")
        .env("HOME", home)
        .env("LC_ALL", "C")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    let mut pipe = child.stdin.take().unwrap();
    if let Some(v) = stdin {
        let _ = pipe.write_all(v.to_string().as_bytes());
    }
    drop(pipe);
    child.wait_with_output().unwrap()
}

fn ok(out: &Output) -> Value {
    let v: Value = serde_json::from_slice(&out.stdout).expect("envelope JSON on stdout");
    assert_eq!(
        v["ok"],
        json!(true),
        "{}{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    v["result"].clone()
}

fn plan(exe: &Path, home: &Path, stage: &Path) -> Value {
    ok(&run(
        exe,
        home,
        &[
            "update-plan",
            home.to_str().unwrap(),
            stage.to_str().unwrap(),
        ],
        None,
    ))
}

fn install(exe: &Path, home: &Path, stage: &Path) {
    let p = plan(exe, home, stage);
    let digest = p["plan_digest"].as_str().unwrap().to_string();
    let r = ok(&run(
        exe,
        home,
        &[
            "update-apply",
            home.to_str().unwrap(),
            stage.to_str().unwrap(),
            "--plan-digest",
            &digest,
            "--permission-reference",
            "fixture-approved",
        ],
        Some(&p),
    ));
    for h in HOSTS {
        let o = r["hosts"]
            .as_array()
            .unwrap()
            .iter()
            .find(|x| x["host"] == h)
            .unwrap();
        assert_eq!(o["status"], json!("applied"), "{r}");
    }
    // Receipts carry second-resolution timestamps; keep the order of installs observable.
    clock_wait::wait_for_next_second();
}

fn host<'a>(plan: &'a Value, name: &str) -> &'a Value {
    plan["hosts"]
        .as_array()
        .unwrap()
        .iter()
        .find(|h| h["host"] == name)
        .unwrap()
}

fn all_unchanged(plan: &Value) {
    for h in HOSTS {
        let hp = host(plan, h);
        assert_eq!(hp["preflight"], json!("ok"), "{h}: {plan}");
        assert!(
            hp["files"]
                .as_array()
                .unwrap()
                .iter()
                .all(|f| f["action"] == "unchanged"),
            "{h}: {plan}"
        );
    }
}

fn blocked_unmanaged(plan: &Value) {
    for h in HOSTS {
        let hp = host(plan, h);
        assert_eq!(hp["preflight"], json!("blocked"), "{h}: {plan}");
        let reason = hp["blocked_reason"].as_str().unwrap();
        assert!(
            reason.contains("unmanaged existing package"),
            "{h}: {reason}"
        );
    }
}

fn id_of(exe: &Path) -> String {
    format!(
        "{:x}",
        Sha256::digest(fs::canonicalize(exe).unwrap().to_str().unwrap().as_bytes())
    )
}

/// Moves `versions` and `installations` from `~/.moai-cli` to `~/.moai` the way the install
/// step will (one journal, ids recorded in it), writes the schema-2 marker, and returns the
/// journal path. `names` are the version folders; their ids before and after become the aliases.
fn move_home(home: &Path, names: &[&str]) -> PathBuf {
    let old = home.join(".moai-cli");
    let new = home.join(".moai");
    fs::create_dir_all(&new).unwrap();
    fs::write(
        new.join("moai-cli.json"),
        r#"{"schema_version":2,"generation":"moai-cli"}"#,
    )
    .unwrap();
    let aliases: BTreeMap<String, String> = names
        .iter()
        .map(|n| {
            (
                id_of(&old.join("versions").join(n).join("bin/moai-cli")),
                // Canonical form of the destination path: the folder does not exist there yet,
                // so build it from the canonical home.
                format!(
                    "{:x}",
                    Sha256::digest(
                        new.join("versions")
                            .join(n)
                            .join("bin/moai-cli")
                            .to_str()
                            .unwrap()
                            .as_bytes()
                    )
                ),
            )
        })
        .collect();
    let items = [
        PlannedMove {
            from: old.join("versions"),
            to: new.join("versions"),
        },
        PlannedMove {
            from: old.join("installations"),
            to: new.join("installations"),
        },
    ];
    let dir = old.join("home-install");
    fs::create_dir_all(&dir).unwrap();
    let env = MoveEnv::new(&never_in_use);
    let report = moai_cli::move_journal::apply(
        &dir,
        moai_cli::home_install::KIND,
        &items,
        &aliases,
        NOW,
        &env,
    )
    .unwrap();
    assert_eq!(
        serde_json::to_value(report.state).unwrap(),
        json!("committed")
    );
    PathBuf::from(report.journal)
}

fn set_state(journal: &Path, state: &str) {
    let mut v: Value = serde_json::from_slice(&fs::read(journal).unwrap()).unwrap();
    v["state"] = json!(state);
    fs::write(journal, serde_json::to_vec_pretty(&v).unwrap()).unwrap();
}

// The folder's own receipt: installed under the old id, planned by the same build after it
// moved. With the committed journal the baseline is found and nothing changes; without it the
// old id means nothing and the unmanaged-package protection blocks, as before.
#[test]
fn a_moved_folder_finds_its_own_receipt_through_the_committed_journal() {
    let f = fixture();
    let old = f.home.join(".moai-cli");
    let a = version_exe(&old, "a");
    install(&a, &f.home, &f.v1);
    let a_old = id_of(&a);
    let journal = move_home(&f.home, &["a"]);
    let moved = f.home.join(".moai/versions/a/bin/moai-cli");
    assert!(moved.is_file());

    let p = plan(&moved, &f.home, &f.v1);
    all_unchanged(&p);

    // A journal that is not committed lends nothing.
    for state in ["pending", "reverted", "revert_partial"] {
        set_state(&journal, state);
        blocked_unmanaged(&plan(&moved, &f.home, &f.v1));
    }
    set_state(&journal, "committed");
    all_unchanged(&plan(&moved, &f.home, &f.v1));

    // A journal of another kind lends nothing either.
    let mut v: Value = serde_json::from_slice(&fs::read(&journal).unwrap()).unwrap();
    v["kind"] = json!("something-else");
    fs::write(&journal, serde_json::to_vec_pretty(&v).unwrap()).unwrap();
    blocked_unmanaged(&plan(&moved, &f.home, &f.v1));
    v["kind"] = json!(moai_cli::home_install::KIND);
    fs::write(&journal, serde_json::to_vec_pretty(&v).unwrap()).unwrap();
    all_unchanged(&plan(&moved, &f.home, &f.v1));

    // An old-id receipt that names another installation is not this folder's, journal or not.
    for h in HOSTS {
        let path = f
            .home
            .join(".moai/installations")
            .join(&a_old)
            .join(format!("hosts/moai/{h}.json"));
        let mut r: Value = serde_json::from_slice(&fs::read(&path).unwrap()).unwrap();
        r["install_id"] = json!("0".repeat(64));
        fs::write(&path, serde_json::to_vec_pretty(&r).unwrap()).unwrap();
    }
    blocked_unmanaged(&plan(&moved, &f.home, &f.v1));
}

// Going back and forth across the move: an earlier folder's receipt, written under its old id,
// is adopted by the build in another folder, and the newest of them wins.
#[test]
fn an_earlier_moved_folders_receipts_are_adopted_and_the_newest_wins() {
    let f = fixture();
    let old = f.home.join(".moai-cli");
    let a = version_exe(&old, "a");
    let b = version_exe(&old, "b");
    install(&a, &f.home, &f.v1);
    install(&b, &f.home, &f.v2);
    let b_old = id_of(&b);
    move_home(&f.home, &["a", "b"]);
    let a2 = f.home.join(".moai/versions/a/bin/moai-cli");
    let b2 = f.home.join(".moai/versions/b/bin/moai-cli");

    // b's receipt is the newest: a (its own receipt is the older one) reads b's state, from the
    // receipt b wrote under its old id.
    let p = plan(&a2, &f.home, &f.v2);
    all_unchanged(&p);
    for h in HOSTS {
        assert_eq!(host(&p, h)["receipt_from"], json!(b_old), "{h}: {p}");
    }
    // b itself finds its own receipt: nothing was adopted from elsewhere.
    let own = plan(&b2, &f.home, &f.v2);
    all_unchanged(&own);
    for h in HOSTS {
        assert!(host(&own, h).get("receipt_from").is_none(), "{h}: {own}");
    }
}

// One folder, two receipts: the old-id one written before the move and a newer one the moved
// build wrote under its new id. The newest stands for the folder, so another build reads the
// newer state and the folder never ties with itself.
#[test]
fn of_one_folders_old_and_new_receipts_the_newest_stands_for_it() {
    let f = fixture();
    let v3 = f.home.parent().unwrap().join("stage-v3");
    write_pkg(
        &v3,
        &[
            ("SKILL.md", skill_md("body v3")),
            ("workflows/update.md", "# UPDATE\nv3\n".into()),
        ],
    );
    let old = f.home.join(".moai-cli");
    let a = version_exe(&old, "a");
    let b = version_exe(&old, "b");
    install(&a, &f.home, &f.v1);
    install(&b, &f.home, &f.v2);
    let b_old = id_of(&b);
    move_home(&f.home, &["a", "b"]);
    let a2 = f.home.join(".moai/versions/a/bin/moai-cli");
    let b2 = f.home.join(".moai/versions/b/bin/moai-cli");

    // b, now at its new place, installs v3 under its new id; its old-id receipt (v2) stays.
    install(&b2, &f.home, &v3);
    let p = plan(&a2, &f.home, &v3);
    for h in HOSTS {
        let hp = host(&p, h);
        assert_eq!(hp["preflight"], json!("ok"), "{h}: {p}");
        assert_eq!(
            hp["receipt_from"],
            json!(id_of(&b2)),
            "{h}: new b {} old b {} from {}",
            id_of(&b2),
            b_old,
            hp["receipt_from"]
        );
        assert!(
            hp["files"]
                .as_array()
                .unwrap()
                .iter()
                .all(|x| x["action"] == "unchanged"),
            "{h}: {p}"
        );
    }
}

// install-inspect: a receipt recorded for the build before the move is `self` after it, through
// the committed journal; without one it is no longer this binary's (its path is gone).
#[test]
fn install_inspect_knows_a_receipt_recorded_before_the_move() {
    let f = fixture();
    let old = f.home.join(".moai-cli");
    let a = version_exe(&old, "a");
    let receipt = plan_receipt(&a, &f.home, InstallerOwner::Unknown).unwrap();
    record_receipt(&f.home, &receipt).unwrap();
    let a_old = id_of(&a);
    let journal = move_home(&f.home, &["a"]);
    let moved = f.home.join(".moai/versions/a/bin/moai-cli");

    let with = ok(&run(&moved, &f.home, &["install-inspect", "--json"], None));
    assert_eq!(with["self"]["state"], json!("self"), "{with}");
    let listed = with["installations"].as_array().unwrap();
    assert_eq!(listed.len(), 1, "{with}");
    assert_eq!(listed[0]["match_state"], json!("self"), "{with}");

    // The id is the journal's word for the path; the digest still has to match the binary.
    let file = f
        .home
        .join(".moai/installations")
        .join(format!("{a_old}.json"));
    let original = fs::read(&file).unwrap();
    let mut r: Value = serde_json::from_slice(&original).unwrap();
    r["binary_sha256"] = json!(format!("sha256:{}", "0".repeat(64)));
    fs::write(&file, serde_json::to_vec_pretty(&r).unwrap()).unwrap();
    let stale = ok(&run(&moved, &f.home, &["install-inspect", "--json"], None));
    assert_eq!(stale["self"]["state"], json!("stale_receipt"), "{stale}");
    fs::write(&file, original).unwrap();

    set_state(&journal, "reverted");
    let without = ok(&run(&moved, &f.home, &["install-inspect", "--json"], None));
    assert_eq!(without["self"]["state"], json!("unrecorded"), "{without}");
}

/// Replaces the journal's alias table (the file is the user's own state, so a test may write
/// what a hand edit or a bad build could).
fn set_aliases(journal: &Path, pairs: &[(&str, &str)]) {
    let mut v: Value = serde_json::from_slice(&fs::read(journal).unwrap()).unwrap();
    v["id_aliases"] = Value::Object(
        pairs
            .iter()
            .map(|(a, b)| (a.to_string(), json!(b)))
            .collect(),
    );
    fs::write(journal, serde_json::to_vec_pretty(&v).unwrap()).unwrap();
}

/// A host receipt file of installation `id`, as JSON.
fn host_receipt_file(state: &Path, id: &str, host: &str) -> PathBuf {
    state
        .join("installations")
        .join(id)
        .join(format!("hosts/moai/{host}.json"))
}

// Audit 6884416142 B1: an alias is a path part under `installations/`. A journal edited (or
// written by a bad build) to name `../../elsewhere` must not make `update` read a receipt
// outside the store and take it as the baseline.
#[test]
fn a_journal_alias_shaped_like_a_path_lends_nothing_for_the_own_receipt() {
    let f = fixture();
    let old = f.home.join(".moai-cli");
    let a = version_exe(&old, "a");
    install(&a, &f.home, &f.v1);
    let a_old = id_of(&a);
    let journal = move_home(&f.home, &["a"]);
    let moved = f.home.join(".moai/versions/a/bin/moai-cli");
    let a_new = id_of(&moved);
    all_unchanged(&plan(&moved, &f.home, &f.v1));

    let state = f.home.join(".moai");
    for h in HOSTS {
        let src = host_receipt_file(&state, &a_old, h);
        let mut r: Value = serde_json::from_slice(&fs::read(&src).unwrap()).unwrap();
        r["install_id"] = json!("../../elsewhere");
        let dst = f.home.join(format!("elsewhere/hosts/moai/{h}.json"));
        fs::create_dir_all(dst.parent().unwrap()).unwrap();
        fs::write(&dst, serde_json::to_vec_pretty(&r).unwrap()).unwrap();
    }
    fs::remove_dir_all(state.join("installations").join(&a_old)).unwrap();
    set_aliases(&journal, &[("../../elsewhere", &a_new)]);
    blocked_unmanaged(&plan(&moved, &f.home, &f.v1));
}

// The same through the other reader: an earlier folder's receipts, looked up by the ids the
// journal lends it.
#[test]
fn a_journal_alias_shaped_like_a_path_lends_nothing_for_an_earlier_folder() {
    let f = fixture();
    let old = f.home.join(".moai-cli");
    let a = version_exe(&old, "a");
    let b = version_exe(&old, "b");
    install(&a, &f.home, &f.v1);
    install(&b, &f.home, &f.v2);
    let (a_old, b_old) = (id_of(&a), id_of(&b));
    let journal = move_home(&f.home, &["a", "b"]);
    let a2 = f.home.join(".moai/versions/a/bin/moai-cli");
    let b_new = id_of(&f.home.join(".moai/versions/b/bin/moai-cli"));
    let a_new = id_of(&a2);

    let state = f.home.join(".moai");
    for h in HOSTS {
        let src = host_receipt_file(&state, &b_old, h);
        let mut r: Value = serde_json::from_slice(&fs::read(&src).unwrap()).unwrap();
        r["install_id"] = json!("../../elsewhere");
        let dst = f.home.join(format!("elsewhere/hosts/moai/{h}.json"));
        fs::create_dir_all(dst.parent().unwrap()).unwrap();
        fs::write(&dst, serde_json::to_vec_pretty(&r).unwrap()).unwrap();
    }
    fs::remove_dir_all(state.join("installations").join(&b_old)).unwrap();
    set_aliases(&journal, &[(&a_old, &a_new), ("../../elsewhere", &b_new)]);
    // a keeps its own (older) receipt as the baseline; the outside file, newer and written for
    // the other package, is not read. The disk holds b's package (installed last), so a's v1
    // baseline shows as `preserved`; an adopted v2 baseline would equal the disk instead.
    let p = plan(&a2, &f.home, &f.v1);
    for h in HOSTS {
        let hp = host(&p, h);
        assert_eq!(hp["preflight"], json!("ok"), "{h}: {p}");
        assert!(hp.get("receipt_from").is_none(), "{h}: {p}");
        for file in hp["files"].as_array().unwrap() {
            assert_eq!(file["baseline_sha256"], file["new_sha256"], "{h}: {p}");
            assert_ne!(file["baseline_sha256"], file["current_sha256"], "{h}: {p}");
        }
    }
}

// Audit B2: an alias counts for `install-inspect` only when it leads to THIS build's id. A
// receipt that another moved build wrote is that build's: its old path is gone, so from here it
// is foreign (the digest alone, equal for two copies of one binary, must not make it `self`).
#[test]
fn install_inspect_does_not_take_another_builds_moved_receipt_for_itself() {
    let f = fixture();
    let old = f.home.join(".moai-cli");
    let a = version_exe(&old, "a");
    let b = version_exe(&old, "b");
    for exe in [&a, &b] {
        let receipt = plan_receipt(exe, &f.home, InstallerOwner::Unknown).unwrap();
        record_receipt(&f.home, &receipt).unwrap();
    }
    let (a_old, b_old) = (id_of(&a), id_of(&b));
    move_home(&f.home, &["a", "b"]);
    let b2 = f.home.join(".moai/versions/b/bin/moai-cli");

    let out = ok(&run(&b2, &f.home, &["install-inspect", "--json"], None));
    let state_of = |id: &str| {
        out["installations"]
            .as_array()
            .unwrap()
            .iter()
            .find(|e| e["install_id"] == json!(id))
            .map(|e| e["match_state"].clone())
            .unwrap_or_else(|| panic!("{id} not listed: {out}"))
    };
    assert_eq!(state_of(&b_old), json!("self"), "{out}");
    assert_eq!(state_of(&a_old), json!("foreign"), "{out}");
    assert_eq!(out["self"]["state"], json!("self"), "{out}");
}

// Audit B3: when the journal lends a folder two earlier ids, the receipt that finished last is
// the baseline. The older id sorts first and its receipt is an interrupted (pending) write; if
// the comparison picked the older one the plan would stop on that pending receipt.
#[test]
fn of_two_earlier_ids_of_this_folder_the_newer_receipt_is_the_baseline() {
    let f = fixture();
    let old = f.home.join(".moai-cli");
    let a = version_exe(&old, "a");
    install(&a, &f.home, &f.v1);
    let a_old = id_of(&a);
    let journal = move_home(&f.home, &["a"]);
    let moved = f.home.join(".moai/versions/a/bin/moai-cli");
    let a_new = id_of(&moved);

    let x2 = "f".repeat(64);
    assert!(a_old < x2, "the older id must be read first");
    let state = f.home.join(".moai");
    for h in HOSTS {
        let src = host_receipt_file(&state, &a_old, h);
        let finished: Value = serde_json::from_slice(&fs::read(&src).unwrap()).unwrap();
        // The newer receipt, under the second id: a copy of the finished one.
        let mut newer = finished.clone();
        newer["install_id"] = json!(x2);
        let dst = host_receipt_file(&state, &x2, h);
        fs::create_dir_all(dst.parent().unwrap()).unwrap();
        fs::write(&dst, serde_json::to_vec_pretty(&newer).unwrap()).unwrap();
        // The older one, under the first id: an apply that never finished.
        let mut older = finished;
        older["status"] = json!("pending");
        older["finished_at"] = Value::Null;
        older["started_at"] = json!("2000-01-01T00:00:00Z");
        fs::write(&src, serde_json::to_vec_pretty(&older).unwrap()).unwrap();
    }
    set_aliases(&journal, &[(&a_old, &a_new), (&x2, &a_new)]);
    all_unchanged(&plan(&moved, &f.home, &f.v1));
}
