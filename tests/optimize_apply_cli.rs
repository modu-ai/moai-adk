//! SPEC-20261120-worktree-cap-and-cleanup, the apply half: `optimize-apply` removes only the
//! worktrees the caller picked with the digest the preview gave, one by one, the deepest first,
//! checking each again just before it goes, and writes a record that restores them.
//!
//! Temporary repository, temporary HOME; the real worktrees are never touched. `git worktree
//! remove` runs without `--force`, so git itself refuses a worktree that changed.
#![cfg(unix)]
mod common;

use serde_json::Value;
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Output};

const BIN: &str = crate::common::MOAI_BIN;

fn git(dir: &Path, args: &[&str]) -> String {
    let mut all = vec![
        "-c",
        "user.name=fixture",
        "-c",
        "user.email=fixture@example.invalid",
    ];
    all.extend_from_slice(args);
    common::git::git(dir, &all)
}

struct Fx {
    _dir: tempfile::TempDir,
    root: PathBuf,
    home: PathBuf,
}

fn fx() -> Fx {
    let dir = tempfile::TempDir::new().unwrap();
    let base = fs::canonicalize(dir.path()).unwrap();
    let (root, home) = (base.join("repo"), base.join("home"));
    fs::create_dir_all(&root).unwrap();
    fs::create_dir_all(&home).unwrap();
    git(&root, &["init", "-q", "-b", "main"]);
    fs::write(root.join(".gitignore"), ".moai-cli/\n.claude/\ntarget/\n").unwrap();
    fs::write(root.join("a.txt"), "a\n").unwrap();
    git(&root, &["add", ".gitignore", "a.txt"]);
    git(&root, &["commit", "-q", "-m", "init"]);
    git(&root, &["branch", "integ"]);
    fs::create_dir_all(root.join(".moai-cli")).unwrap();
    fs::write(
        root.join(".moai-cli/project.yaml"),
        r#"{"project_id":"demo"}"#,
    )
    .unwrap();
    moai_cli::ledger::register(&home, "demo").expect("register ledger");
    Fx {
        _dir: dir,
        root,
        home,
    }
}

/// A merged (nothing of its own), clean worktree with cargo output.
fn add(fx: &Fx, path: &Path, branch: &str) -> PathBuf {
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    git(
        &fx.root,
        &[
            "worktree",
            "add",
            "-q",
            "-b",
            branch,
            path.to_str().unwrap(),
            "integ",
        ],
    );
    let t = path.join("target/debug/deps");
    fs::create_dir_all(&t).unwrap();
    fs::write(
        path.join("target/CACHEDIR.TAG"),
        "Signature: 8a477f597d28d172789f06886806bc55\n",
    )
    .unwrap();
    fs::write(path.join("target/debug/.cargo-lock"), "").unwrap();
    fs::write(t.join("lib.rlib"), vec![7u8; 65_536]).unwrap();
    fs::canonicalize(path).unwrap()
}

fn wt(fx: &Fx, name: &str) -> PathBuf {
    fx.root.join(".moai-cli/worktrees").join(name)
}

fn now() -> String {
    (std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .unwrap()
        .as_secs()
        + 30 * 86_400)
        .to_string()
}

fn run(fx: &Fx, args: &[&str]) -> Output {
    Command::new(BIN)
        .args(args)
        .current_dir(&fx.root)
        .env("HOME", &fx.home)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .output()
        .unwrap()
}

fn envelope(out: &Output) -> Value {
    serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("{}", String::from_utf8_lossy(&out.stderr)))
}

fn preview(fx: &Fx) -> Value {
    let now = now();
    let v = envelope(&run(
        fx,
        &[
            "optimize-plan",
            fx.root.to_str().unwrap(),
            "--integration-ref",
            "integ",
            "--now-unix",
            &now,
        ],
    ));
    assert_eq!(v["ok"], true, "{v}");
    v["result"]["cleanup"].clone()
}

/// `path@digest` of the candidate whose path ends with `tail`.
fn pick(cleanup: &Value, tail: &str) -> String {
    let c = cleanup["candidates"]
        .as_array()
        .unwrap()
        .iter()
        .find(|c| c["path"].as_str().is_some_and(|p| p.ends_with(tail)))
        .unwrap_or_else(|| panic!("{tail} is not a candidate: {cleanup}"));
    format!(
        "{}@{}",
        c["path"].as_str().unwrap(),
        c["item_digest"].as_str().unwrap()
    )
}

fn apply(fx: &Fx, permission: &str, picks: &[String]) -> Value {
    let now = now();
    let mut args = vec![
        "optimize-apply".to_string(),
        fx.root.to_str().unwrap().into(),
        "--integration-ref".into(),
        "integ".into(),
        "--now-unix".into(),
        now,
        "--permission-reference".into(),
        permission.into(),
    ];
    for p in picks {
        args.push("--pick".into());
        args.push(p.clone());
    }
    let refs: Vec<&str> = args.iter().map(String::as_str).collect();
    envelope(&run(fx, &refs))
}

fn removed_paths(v: &Value) -> Vec<String> {
    v["result"]["removed"]
        .as_array()
        .map(|a| {
            a.iter()
                .filter_map(|r| r["path"].as_str().map(String::from))
                .collect()
        })
        .unwrap_or_default()
}

#[test]
fn apply_removes_only_the_picked_clean_worktrees_inside_out() {
    let fx = fx();
    let parent = add(&fx, &wt(&fx, "parent"), "b-parent");
    let inner = add(&fx, &parent.join("target/inner"), "b-inner");
    let keep = add(&fx, &wt(&fx, "keepme"), "b-keep");
    let dirty = add(&fx, &wt(&fx, "dirty"), "b-dirty");
    let late = add(&fx, &wt(&fx, "late"), "b-late");
    let cleanup = preview(&fx);
    // parent is named before inner on purpose: the order of removal is the tool's
    let picks = vec![
        pick(&cleanup, "worktrees/parent"),
        pick(&cleanup, "target/inner"),
        pick(&cleanup, "worktrees/dirty"),
        pick(&cleanup, "worktrees/late"),
    ];
    // a worktree that changes after the preview is refused, the others go on
    fs::write(dirty.join("a.txt"), "changed after the preview\n").unwrap();

    // without a usable permission reference nothing is removed
    let none = apply(&fx, "", &picks);
    assert_eq!(none["ok"], false, "{none}");
    assert!(parent.is_dir() && inner.is_dir() && late.is_dir());

    // a digest that is not the preview's refuses that one item only
    let mut bad = picks.clone();
    bad[3] = format!("{}@0000000000000000", bad[3].split('@').next().unwrap());
    let v = apply(&fx, "test:cap", &bad);
    assert_eq!(v["ok"], true, "{v}");
    let gone = removed_paths(&v);
    assert!(
        !gone.iter().any(|p| p.ends_with("worktrees/late")),
        "a wrong digest must not remove: {v}"
    );
    assert!(late.is_dir());
    let refused = v["result"]["refused"].as_array().unwrap();
    assert!(
        refused.iter().any(|r| r["path"]
            .as_str()
            .is_some_and(|p| p.ends_with("worktrees/late"))
            && r["reason"] == "digest_mismatch"),
        "{v}"
    );
    // inside out: the nested worktree is removed before the one it sits in
    let at = |tail: &str| gone.iter().position(|p| p.ends_with(tail));
    assert!(
        at("target/inner").unwrap() < at("worktrees/parent").unwrap(),
        "{gone:?}"
    );
    // the dirty one is refused by our own fresh check (git would refuse it too, but later)
    // and the reason says what changed
    let dirty_row = refused
        .iter()
        .find(|r| {
            r["path"]
                .as_str()
                .is_some_and(|p| p.ends_with("worktrees/dirty"))
        })
        .unwrap_or_else(|| panic!("dirty not refused: {v}"));
    assert_eq!(dirty_row["reason"], "not_a_candidate", "{dirty_row}");
    assert!(
        dirty_row["detail"]
            .as_array()
            .is_some_and(|d| d.iter().any(|x| x == "tracked_changes")),
        "{dirty_row}"
    );
    assert!(dirty.is_dir());
    assert!(!parent.exists() && !inner.exists());
    // not picked: untouched; branches are always kept
    assert!(keep.is_dir());
    for b in ["b-parent", "b-inner", "b-dirty", "b-keep", "b-late"] {
        assert!(
            !git(&fx.root, &["branch", "--list", b]).trim().is_empty(),
            "{b} must stay"
        );
    }

    // the right digest for the one left goes through, and the dirty one is still refused
    let again = apply(&fx, "test:cap", &[picks[3].clone(), picks[2].clone()]);
    assert_eq!(removed_paths(&again).len(), 1, "{again}");
    assert!(!late.exists() && dirty.is_dir());
}

#[test]
fn apply_writes_a_record_that_restores_the_worktree() {
    let fx = fx();
    let one = add(&fx, &wt(&fx, "one"), "b-one");
    let head = git(&one, &["rev-parse", "HEAD"]).trim().to_string();
    let cleanup = preview(&fx);
    let v = apply(&fx, "test:cap", &[pick(&cleanup, "worktrees/one")]);
    assert_eq!(v["ok"], true, "{v}");
    assert!(!one.exists());
    let r = &v["result"];
    // the reclaimed space is the change of the free bytes, never a sum of sizes
    for k in ["free_bytes_before", "free_bytes_after", "free_bytes_delta"] {
        assert!(r[k].is_number(), "{k}: {r}");
    }
    assert!(
        r["space_basis"].as_str().is_some_and(|s| s.contains("df")),
        "{r}"
    );
    // the record is under the repository's git folder and names what a restore needs
    let record = PathBuf::from(r["record"].as_str().expect("record path"));
    assert!(
        record.starts_with(fx.root.join(".git/.moai-cli/worktree-cleanup")),
        "{record:?}"
    );
    let rec: Value = serde_json::from_str(&fs::read_to_string(&record).unwrap()).unwrap();
    let item = &rec["items"][0];
    assert_eq!(item["branch"], "b-one", "{rec}");
    assert_eq!(item["head"], head, "{rec}");
    let restore = item["restore"]
        .as_str()
        .expect("restore command")
        .to_string();
    assert!(restore.starts_with("git worktree add"), "{restore}");
    // running it brings the worktree back on its branch
    let out = Command::new("sh")
        .arg("-c")
        .arg(&restore)
        .current_dir(&fx.root)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .output()
        .unwrap();
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    assert!(one.join("a.txt").is_file());
    assert_eq!(git(&one, &["rev-parse", "HEAD"]).trim(), head);

    // a record that cannot be written is reported together with what was removed
    let two = add(&fx, &wt(&fx, "two"), "b-two");
    let cleanup = preview(&fx);
    let _ = fs::remove_dir_all(fx.root.join(".git/.moai-cli"));
    fs::write(fx.root.join(".git/.moai-cli"), "in the way").unwrap();
    let failed = apply(&fx, "test:cap", &[pick(&cleanup, "worktrees/two")]);
    assert_eq!(failed["ok"], false, "{failed}");
    let msg = failed["error"]["message"].as_str().unwrap_or_default();
    assert!(msg.contains("worktrees/two"), "{msg}");
    assert!(!two.exists(), "the removal itself had happened");
}
