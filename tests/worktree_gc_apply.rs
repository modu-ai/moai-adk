//! #2498 (WT-9, #2380): `moai-cli worktree-gc-apply` and the WT-4 audit notes N1–N3, on a
//! synthetic repository in a temporary folder only. This machine's real worktrees are never
//! planned or cleaned here (the real run is a user-confirmed step, HOLD).
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

struct Fixture {
    _dir: tempfile::TempDir,
    root: PathBuf,
    home: PathBuf,
}

impl Fixture {
    fn wt(&self, name: &str) -> PathBuf {
        self.root.join(".moai-cli/worktrees").join(name)
    }
}

fn cargo_output(wt: &Path, profile: &str) {
    fs::create_dir_all(wt.join("target").join(profile).join("deps")).unwrap();
    fs::write(
        wt.join("target/CACHEDIR.TAG"),
        "Signature: 8a477f597d28d172789f06886806bc55\n",
    )
    .unwrap();
    fs::write(wt.join("target").join(profile).join(".cargo-lock"), "").unwrap();
    fs::write(
        wt.join("target").join(profile).join("deps/lib.rlib"),
        vec![7u8; 4096],
    )
    .unwrap();
}

/// - `a` merged, clean, cargo output only -> remove
/// - `b` merged, cargo output plus a lane note `target/notes/pr.md` -> trim, note preserved
/// - `c` one commit not in `integ` -> skip not_merged
/// - `g` merged, clean, nothing in it at all (a fresh, unstarted worktree) -> skip (N3)
fn fixture() -> Fixture {
    let dir = tempfile::TempDir::new().unwrap();
    let base = fs::canonicalize(dir.path()).unwrap();
    let root = base.join("repo");
    let home = base.join("home");
    fs::create_dir_all(&root).unwrap();
    fs::create_dir_all(&home).unwrap();
    git(&root, &["init", "-q", "-b", "main"]);
    fs::write(root.join(".gitignore"), ".moai-cli/\ntarget/\n").unwrap();
    fs::write(root.join("a.txt"), "a\n").unwrap();
    git(&root, &["add", ".gitignore", "a.txt"]);
    git(&root, &["commit", "-q", "-m", "init"]);
    git(&root, &["branch", "integ"]);
    for name in ["a", "b", "c", "g"] {
        let wt = root.join(".moai-cli/worktrees").join(name);
        let branch = format!("issue/1-{name}");
        git(
            &root,
            &["worktree", "add", "-q", "-b", &branch, wt.to_str().unwrap()],
        );
    }
    let fx = Fixture {
        _dir: dir,
        root,
        home,
    };
    cargo_output(&fx.wt("a"), "debug");
    cargo_output(&fx.wt("b"), "debug");
    fs::create_dir_all(fx.wt("b").join("target/notes")).unwrap();
    fs::write(fx.wt("b").join("target/notes/pr.md"), "lane note\n").unwrap();
    fs::write(fx.wt("c").join("c.txt"), "c\n").unwrap();
    git(&fx.wt("c"), &["add", "c.txt"]);
    git(&fx.wt("c"), &["commit", "-q", "-m", "unmerged"]);
    fx
}

/// A day from now: every fixture change is older than the default window.
fn later() -> u64 {
    std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .unwrap()
        .as_secs()
        + 86_400
}

fn run(fx: &Fixture, args: &[&str]) -> Output {
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
    serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
        panic!(
            "envelope on stdout; stderr {}",
            String::from_utf8_lossy(&out.stderr)
        )
    })
}

fn plan_args(fx: &Fixture, now: u64, extra: &[String]) -> Vec<String> {
    let mut a = vec![
        "worktree-gc-plan".to_string(),
        fx.root.to_str().unwrap().to_string(),
        "--integration-ref".into(),
        "integ".into(),
        "--now-unix".into(),
        now.to_string(),
    ];
    a.extend_from_slice(extra);
    a
}

fn plan(fx: &Fixture, now: u64) -> Value {
    let args = plan_args(fx, now, &[]);
    let refs: Vec<&str> = args.iter().map(String::as_str).collect();
    let v = envelope(&run(fx, &refs));
    assert_eq!(v["ok"], true, "{v}");
    v["result"].clone()
}

fn apply(fx: &Fixture, now: u64, digest: &str) -> Value {
    let args = [
        "worktree-gc-apply",
        fx.root.to_str().unwrap(),
        "--integration-ref",
        "integ",
        "--now-unix",
        &now.to_string(),
        "--plan-digest",
        digest,
        "--permission-reference",
        "fixture-approved",
    ];
    envelope(&run(fx, &args))
}

fn entry<'a>(r: &'a Value, path: &str) -> &'a Value {
    r["candidates"]
        .as_array()
        .unwrap()
        .iter()
        .find(|c| c["path"] == path)
        .unwrap_or_else(|| panic!("no {path}: {r}"))
}

/// Every file and folder under the repository (relative), for "nothing else changed" checks.
fn tree(root: &Path) -> Vec<String> {
    fn walk(base: &Path, dir: &Path, out: &mut Vec<String>) {
        let Ok(rd) = fs::read_dir(dir) else { return };
        for e in rd.flatten() {
            let p = e.path();
            out.push(p.strip_prefix(base).unwrap().to_string_lossy().into_owned());
            if fs::symlink_metadata(&p)
                .map(|m| m.is_dir())
                .unwrap_or(false)
            {
                walk(base, &p, out);
            }
        }
    }
    let mut out = Vec::new();
    walk(root, root, &mut out);
    out.sort();
    out
}

// N3: a fresh worktree with nothing in it is never a removal candidate.
#[test]
fn an_empty_new_worktree_is_kept() {
    let fx = fixture();
    // Three hours on: past the two-hour recent window, still inside the empty worktree's day.
    let soon = later() - 86_400 + 3 * 3_600;
    let r = plan(&fx, soon);
    let g = entry(&r, ".moai-cli/worktrees/g");
    assert_eq!(g["action"], "skip", "{g}");
    assert_eq!(g["reasons"], serde_json::json!(["empty_worktree"]), "{g}");
    // A built and finished worktree is still removable.
    assert_eq!(entry(&r, ".moai-cli/worktrees/a")["action"], "remove");
    // #2879: two days on the empty merged worktree goes too (it holds only its checkout).
    let r = plan(&fx, later() + 86_400);
    assert_eq!(
        entry(&r, ".moai-cli/worktrees/g")["action"],
        "remove",
        "{r}"
    );
}

// N2: `--recent-minutes` has a floor of 60 — 59 is refused, 60 is accepted (literal values).
#[test]
fn recent_minutes_below_sixty_is_refused() {
    let fx = fixture();
    for (minutes, ok) in [("59", false), ("60", true), ("0", false)] {
        let args = plan_args(&fx, later(), &["--recent-minutes".into(), minutes.into()]);
        let refs: Vec<&str> = args.iter().map(String::as_str).collect();
        let v = envelope(&run(&fx, &refs));
        assert_eq!(v["ok"], ok, "--recent-minutes {minutes}: {v}");
        if !ok {
            assert!(
                v["error"]["message"]
                    .as_str()
                    .unwrap_or_default()
                    .contains("60"),
                "{v}"
            );
        }
    }
}

// N1: the digest binds the trim paths — a new cargo folder after the plan changes it.
#[test]
fn the_digest_binds_the_trim_paths() {
    let fx = fixture();
    let now = later();
    let before = plan(&fx, now)["plan_digest"].as_str().unwrap().to_string();
    cargo_output(&fx.wt("b"), "release");
    let after = plan(&fx, now)["plan_digest"].as_str().unwrap().to_string();
    assert_ne!(before, after, "a new trim path must change the digest");
}

// AC-1: a stale digest is refused before any write.
#[test]
fn a_stale_digest_is_refused_and_nothing_changes() {
    let fx = fixture();
    let now = later();
    let digest = plan(&fx, now)["plan_digest"].as_str().unwrap().to_string();
    cargo_output(&fx.wt("b"), "release"); // a trim path the plan never listed
    let before = tree(&fx.root);
    let v = apply(&fx, now, &digest);
    assert_eq!(v["ok"], false, "{v}");
    assert!(
        v["error"]["message"]
            .as_str()
            .unwrap()
            .contains("re-preview"),
        "{v}"
    );
    assert_eq!(tree(&fx.root), before, "nothing was written");
}

// AC-1 + AC-3: with the plan's digest, apply removes and trims exactly what the plan
// listed and nothing else.
#[test]
fn apply_acts_only_on_the_planned_paths() {
    let fx = fixture();
    // Three hours on, so the empty `g` is still inside its day and stays a `skip` (#2879).
    let now = later() - 86_400 + 3 * 3_600;
    let r = plan(&fx, now);
    let digest = r["plan_digest"].as_str().unwrap().to_string();
    let v = apply(&fx, now, &digest);
    assert_eq!(v["ok"], true, "{v}");
    let res = &v["result"];
    assert_eq!(res["plan_digest"], digest);
    // `a` (remove): the worktree folder is gone, its branch is kept.
    assert!(!fx.wt("a").exists(), "{res}");
    assert!(git(&fx.root, &["branch", "--list", "issue/1-a"]).contains("issue/1-a"));
    // `b` (trim): cargo output gone, the lane note and the checkout kept.
    assert!(!fx.wt("b").join("target/debug").exists());
    assert!(!fx.wt("b").join("target/CACHEDIR.TAG").exists());
    assert_eq!(
        fs::read_to_string(fx.wt("b").join("target/notes/pr.md")).unwrap(),
        "lane note\n"
    );
    assert!(fx.wt("b").join("a.txt").exists());
    // skip entries are untouched.
    assert!(fx.wt("c").join("c.txt").exists());
    assert!(fx.wt("g").join("a.txt").exists());
    let done: Vec<(String, String)> = res["results"]
        .as_array()
        .unwrap()
        .iter()
        .map(|x| {
            (
                x["path"].as_str().unwrap().to_string(),
                x["status"].as_str().unwrap().to_string(),
            )
        })
        .collect();
    assert_eq!(
        done,
        [
            (".moai-cli/worktrees/a".to_string(), "removed".to_string()),
            (".moai-cli/worktrees/b".to_string(), "trimmed".to_string()),
        ]
    );
}

/// `--worktree` reaches the apply too (#2879; kept when the command moved to `src/ops/`,
/// #2834 R2c): a plan and an apply limited to `a` remove `a` and leave `b`'s cargo output.
#[test]
fn apply_with_worktree_acts_only_on_that_worktree() {
    let fx = fixture();
    let now = later() - 86_400 + 3 * 3_600;
    let a = fx.wt("a");
    let only = ["--worktree".to_string(), a.to_str().unwrap().to_string()];
    let args = plan_args(&fx, now, &only);
    let refs: Vec<&str> = args.iter().map(String::as_str).collect();
    let r = envelope(&run(&fx, &refs));
    assert_eq!(r["ok"], true, "{r}");
    let digest = r["result"]["plan_digest"].as_str().unwrap().to_string();
    let now_s = now.to_string();
    let v = envelope(&run(
        &fx,
        &[
            "worktree-gc-apply",
            fx.root.to_str().unwrap(),
            "--integration-ref",
            "integ",
            "--now-unix",
            &now_s,
            "--worktree",
            a.to_str().unwrap(),
            "--plan-digest",
            &digest,
            "--permission-reference",
            "fixture-approved",
        ],
    ));
    assert_eq!(v["ok"], true, "{v}");
    assert!(!a.exists(), "{v}");
    assert!(fx.wt("b").join("target/debug").exists(), "{v}");
}

// AC-3: a worktree that comes into use after the plan (a process working in it) is not
// touched — the recomputed plan differs, so the apply refuses before writing.
#[test]
fn a_worktree_in_use_after_the_plan_is_not_touched() {
    let fx = fixture();
    let now = later();
    let digest = plan(&fx, now)["plan_digest"].as_str().unwrap().to_string();
    let mut sleeper = Command::new("sleep")
        .arg("30")
        .current_dir(fx.wt("a"))
        .spawn()
        .unwrap();
    let before = tree(&fx.root);
    let v = apply(&fx, now, &digest);
    let _ = sleeper.kill();
    let _ = sleeper.wait();
    assert_eq!(v["ok"], false, "{v}");
    assert_eq!(tree(&fx.root), before);
}

// AC-3: a file held open inside a worktree by a process working elsewhere is invisible to
// the plan (it reads working directories only), so the digest still matches — the apply's
// own re-check right before the deletion refuses that entry and leaves it whole.
#[test]
fn a_file_held_open_inside_a_worktree_refuses_that_entry() {
    let fx = fixture();
    let now = later();
    let digest = plan(&fx, now)["plan_digest"].as_str().unwrap().to_string();
    let held = fx.wt("a").join("target/debug/deps/lib.rlib");
    let mut tail = Command::new("tail")
        .args(["-f", held.to_str().unwrap()])
        .current_dir(&fx.home)
        .stdout(std::process::Stdio::null())
        .spawn()
        .unwrap();
    std::thread::sleep(std::time::Duration::from_millis(300)); // let tail open the file
    let v = apply(&fx, now, &digest);
    let _ = tail.kill();
    let _ = tail.wait();
    assert_eq!(v["ok"], true, "{v}");
    let a = v["result"]["results"]
        .as_array()
        .unwrap()
        .iter()
        .find(|x| x["path"] == ".moai-cli/worktrees/a")
        .unwrap()
        .clone();
    assert_eq!(a["status"], "refused", "{a}");
    assert!(
        a["detail"].as_str().unwrap().contains("file_open_in_use"),
        "{a}"
    );
    assert!(held.exists(), "the held file and its worktree stay");
    assert_eq!(v["result"]["status"], "partial");
}

// AC-3: an uncommitted change after the plan blocks the apply the same way.
#[test]
fn an_uncommitted_change_after_the_plan_is_not_touched() {
    let fx = fixture();
    let now = later();
    let digest = plan(&fx, now)["plan_digest"].as_str().unwrap().to_string();
    fs::write(fx.wt("a").join("a.txt"), "edited after the plan\n").unwrap();
    let before = tree(&fx.root);
    let v = apply(&fx, now, &digest);
    assert_eq!(v["ok"], false, "{v}");
    assert_eq!(tree(&fx.root), before);
    assert_eq!(
        fs::read_to_string(fx.wt("a").join("a.txt")).unwrap(),
        "edited after the plan\n"
    );
}

// AC-3: a recent change (inside the window) after the plan blocks the apply.
#[test]
fn a_recent_change_after_the_plan_is_not_touched() {
    let fx = fixture();
    // Plan "a day later", apply at the real clock: every fixture change is now recent.
    let digest = plan(&fx, later())["plan_digest"]
        .as_str()
        .unwrap()
        .to_string();
    let now = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .unwrap()
        .as_secs();
    let before = tree(&fx.root);
    let v = apply(&fx, now, &digest);
    assert_eq!(v["ok"], false, "{v}");
    assert_eq!(tree(&fx.root), before);
}

// The permission reference is required, as for every other apply.
#[test]
fn apply_needs_a_permission_reference() {
    let fx = fixture();
    let now = later();
    let digest = plan(&fx, now)["plan_digest"].as_str().unwrap().to_string();
    let before = tree(&fx.root);
    let out = run(
        &fx,
        &[
            "worktree-gc-apply",
            fx.root.to_str().unwrap(),
            "--integration-ref",
            "integ",
            "--now-unix",
            &now.to_string(),
            "--plan-digest",
            &digest,
            "--permission-reference",
            "placeholder",
        ],
    );
    assert_eq!(envelope(&out)["ok"], false);
    assert_eq!(tree(&fx.root), before);
}

/// Audit B1 on PR #2557 (6827556008, agent-25): a worktree whose `target` is a symlink to a
/// folder outside it (for example a shared build folder) is never trimmed — the plan skips it
/// and apply leaves the outside folder exactly as it was.
#[cfg(unix)]
#[test]
fn a_target_symlinked_outside_the_worktree_is_never_trimmed() {
    let fx = fixture();
    let outside = fx.root.parent().unwrap().join("shared-target");
    cargo_output(&outside.join("x"), "debug");
    let outside = outside.join("x/target");
    fs::write(outside.join("debug/deps/other-lane.rlib"), "other lane\n").unwrap();
    fs::remove_dir_all(fx.wt("b").join("target")).unwrap();
    std::os::unix::fs::symlink(&outside, fx.wt("b").join("target")).unwrap();
    let before_outside = tree(&outside);
    let now = later();
    let r = plan(&fx, now);
    let b = entry(&r, ".moai-cli/worktrees/b");
    assert_ne!(b["action"], "trim", "{b}");
    assert_ne!(b["action"], "remove", "{b}");
    let digest = r["plan_digest"].as_str().unwrap().to_string();
    let v = apply(&fx, now, &digest);
    assert_eq!(v["ok"], true, "{v}");
    assert_eq!(
        tree(&outside),
        before_outside,
        "the outside folder is untouched"
    );
    assert!(fs::symlink_metadata(fx.wt("b").join("target"))
        .unwrap()
        .file_type()
        .is_symlink());
}

// ---- #2811 W2: old, new and Go worktree folders together (AC-3, D3, D8a) ----

/// Adds a merged worktree with cargo output at `wt` (branch `issue/1-<name>`).
fn add_merged(fx: &Fixture, wt: &Path, name: &str) {
    let branch = format!("issue/1-{name}");
    git(
        &fx.root,
        &["worktree", "add", "-q", "-b", &branch, wt.to_str().unwrap()],
    );
    cargo_output(wt, "debug");
}

/// A new-place worktree (`<HOME>/.moai/worktrees/repo-1-h`) is cleaned like an old-place one,
/// while a worktree in the Go generation's folder — reached through a `~/.moai` that is a
/// symbolic link to another folder — is planned `skip preserved_location` and the apply
/// leaves it, its cargo output and the link's target folder exactly as they were.
#[test]
fn new_place_worktrees_are_cleaned_and_go_place_ones_never_touched() {
    let fx = fixture();
    let real = fx.home.parent().unwrap().join("real-moai");
    fs::create_dir_all(real.join("worktrees/projects/repo")).unwrap();
    fs::write(real.join("settings.yaml"), "go: true\n").unwrap();
    std::os::unix::fs::symlink(&real, fx.home.join(".moai")).unwrap();
    let new_wt = fx.home.join(".moai/worktrees/repo-1-h");
    let go_wt = fx.home.join(".moai/worktrees/projects/repo/feat");
    add_merged(&fx, &new_wt, "h");
    add_merged(&fx, &go_wt, "go");
    let before_real = tree(&real.join("worktrees/projects"));

    let now = later();
    let r = plan(&fx, now);
    // git lists the resolved path; the plan still names both places.
    let new_e = entry(&r, real.join("worktrees/repo-1-h").to_str().unwrap());
    assert_eq!(new_e["location"], "home", "{new_e}");
    assert_eq!(new_e["action"], "remove", "{new_e}");
    let go_e = entry(
        &r,
        real.join("worktrees/projects/repo/feat").to_str().unwrap(),
    );
    assert_eq!(go_e["location"], "go", "{go_e}");
    assert_eq!(go_e["action"], "skip", "{go_e}");
    assert_eq!(go_e["reasons"], serde_json::json!(["preserved_location"]));
    assert_eq!(entry(&r, ".moai-cli/worktrees/a")["location"], "project");
    assert_eq!(entry(&r, ".")["location"], "primary");

    let v = apply(&fx, now, r["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], true, "{v}");
    assert!(!real.join("worktrees/repo-1-h").exists(), "{v}");
    assert!(!fx.wt("a").exists(), "old place cleaned the same way: {v}");
    // The Go folder: same files, same cargo output, link and its other files intact.
    assert_eq!(tree(&real.join("worktrees/projects")), before_real);
    assert!(go_wt.join("target/debug").is_dir());
    assert_eq!(
        fs::read_to_string(real.join("settings.yaml")).unwrap(),
        "go: true\n"
    );
    assert!(fs::symlink_metadata(fx.home.join(".moai"))
        .unwrap()
        .file_type()
        .is_symlink());
}

/// Without any home in the environment the Go folder's shape alone keeps its worktree.
#[test]
fn without_a_home_the_go_folder_shape_still_preserves() {
    let fx = fixture();
    let go_wt = fx.home.join(".moai/worktrees/projects/repo/feat");
    add_merged(&fx, &go_wt, "go");
    let out = Command::new(BIN)
        .args(plan_args(&fx, later(), &[]))
        .current_dir(&fx.root)
        .env_remove("HOME")
        .env_remove("USERPROFILE")
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .output()
        .unwrap();
    let r = envelope(&out)["result"].clone();
    let e = entry(&r, go_wt.to_str().unwrap());
    assert_eq!(e["action"], "skip", "{e}");
    assert_eq!(e["reasons"], serde_json::json!(["preserved_location"]));
}

/// The recent-change window at its edge for a new-place worktree: a change exactly
/// `window - 1` seconds old keeps it, one `window` seconds old does not (N/N+1).
#[test]
fn the_recent_window_edge_holds_for_a_new_place_worktree() {
    let fx = fixture();
    let new_wt = fx.home.join(".moai/worktrees/repo-1-h");
    add_merged(&fx, &new_wt, "h");
    let probe = later();
    let shown = new_wt.to_str().unwrap().to_string();
    let age = plan(&fx, probe)["candidates"]
        .as_array()
        .unwrap()
        .iter()
        .find(|c| c["path"] == shown.as_str())
        .unwrap()["change_age_secs"]
        .as_u64()
        .unwrap();
    let changed_at = probe - age;
    let window = 120 * 60;
    let kept = plan(&fx, changed_at + window - 1);
    let e = entry(&kept, &shown);
    assert_eq!(e["action"], "skip", "{e}");
    assert!(e["reasons"].to_string().contains("recent_change"), "{e}");
    let gone = plan(&fx, changed_at + window);
    assert_eq!(entry(&gone, &shown)["action"], "remove");
}
