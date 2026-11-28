//! SPEC-20261124-audit-prep: `audit-prep <card> --head H` takes an audit slot for the card,
//! switches it detached to the head, inherits the lane's build and runs one incremental
//! `cargo test --no-run` of the named test targets; `audit-done <card>` finishes the card's slot
//! and never discards or forces. The binary runs on temporary repositories and a temporary
//! crate in a clean environment; no test builds the real repository.
mod common;

use common::git::{git, git_raw};
use common::home::TempHome;
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::Command;

const LIB: &str = "pub fn two() -> u32 { auditdep::one() + 1 }\n";
const WHERE_TEST: &str = "use auditapp::two;\n#[test]\nfn where_am_i() {\n    println!(\"MANIFEST={}\", env!(\"CARGO_MANIFEST_DIR\"));\n    assert_eq!(two(), 2);\n}\n";

struct Fixture {
    _dir: tempfile::TempDir,
    home: TempHome,
    root: PathBuf,
    lane: PathBuf,
    /// The first commit, on `main`.
    first: String,
    /// The lane's commit on top of `first`, checked out in `lane`.
    second: String,
}

fn commit(dir: &Path, message: &str) {
    git(
        dir,
        &[
            "-c",
            "user.name=t",
            "-c",
            "user.email=t@example.invalid",
            "commit",
            "-q",
            "-m",
            message,
        ],
    );
}

/// The `cargo` of this test run, as Cargo or nextest name it.
fn cargo_path() -> String {
    common::cargo::cargo_path()
        .into_string()
        .unwrap_or_else(|_| "cargo".into())
}

/// A `cargo` run for the fixture's own set-up and for what the tests measure directly.
fn cargo(dir: &Path, args: &[&str]) -> std::process::Output {
    common::cargo::cargo_at(cargo_path())
        .args(args)
        .current_dir(dir)
        .env_remove("CARGO_TARGET_DIR")
        .env_remove("CARGO_BUILD_TARGET_DIR")
        .env_remove("CARGO_INCREMENTAL")
        .output()
        .unwrap()
}

fn cargo_ok(dir: &Path, args: &[&str]) -> String {
    let out = cargo(dir, args);
    assert!(
        out.status.success(),
        "cargo {args:?}: {}",
        String::from_utf8_lossy(&out.stderr)
    );
    String::from_utf8_lossy(&out.stdout).to_string()
}

/// A crate `auditapp` with a path dependency `auditdep` outside the repository (as a registry
/// crate sits outside it), a test target `where` that prints `CARGO_MANIFEST_DIR`, a test target
/// `second`, and a test target `broken` that does not compile (a mismatched type, E0308).
/// `Cargo.lock` is committed, so `--locked` works.
fn fixture() -> Fixture {
    let dir = tempfile::tempdir().unwrap();
    let base = fs::canonicalize(dir.path()).unwrap();
    let dep = base.join("dep");
    fs::create_dir_all(&dep).unwrap();
    fs::write(
        dep.join("Cargo.toml"),
        "[package]\nname = \"auditdep\"\nversion = \"0.1.0\"\nedition = \"2021\"\n[lib]\npath = \"lib.rs\"\n",
    )
    .unwrap();
    fs::write(dep.join("lib.rs"), "pub fn one() -> u32 { 1 }\n").unwrap();
    let root = base.join("repo");
    fs::create_dir_all(root.join("src")).unwrap();
    fs::create_dir_all(root.join("tests")).unwrap();
    fs::write(
        root.join("Cargo.toml"),
        format!(
            "[package]\nname = \"auditapp\"\nversion = \"0.1.0\"\nedition = \"2021\"\n[workspace]\n[dependencies]\nauditdep = {{ path = \"{}\" }}\n",
            dep.display()
        ),
    )
    .unwrap();
    fs::write(root.join("src/lib.rs"), LIB).unwrap();
    fs::write(root.join("tests/where.rs"), WHERE_TEST).unwrap();
    fs::write(
        root.join("tests/second.rs"),
        "use auditapp::two;\n#[test]\nfn second() {\n    assert_eq!(two(), 2);\n}\n",
    )
    .unwrap();
    fs::write(
        root.join("tests/broken.rs"),
        "fn number() -> u32 {\n    \"text\"\n}\n",
    )
    .unwrap();
    fs::write(root.join(".gitignore"), "/target/\n/.claude/\n").unwrap();
    git(&root, &["init", "-q", "-b", "main"]);
    cargo_ok(&root, &["generate-lockfile", "--offline"]);
    git(&root, &["add", "-A"]);
    commit(&root, "one");
    let first = git(&root, &["rev-parse", "HEAD"]);
    let lane = base.join("lane");
    git(
        &root,
        &[
            "worktree",
            "add",
            "-q",
            "-b",
            "work",
            lane.to_str().unwrap(),
        ],
    );
    fs::write(lane.join("b.txt"), "b\n").unwrap();
    git(&lane, &["add", "b.txt"]);
    commit(&lane, "two");
    let second = git(&lane, &["rev-parse", "HEAD"]);
    Fixture {
        _dir: dir,
        home: TempHome::new(),
        root,
        lane,
        first,
        second,
    }
}

/// The `compiler-artifact` messages of a `--message-format json` run.
fn artifacts(stdout: &str) -> Vec<Value> {
    stdout
        .lines()
        .filter_map(|l| serde_json::from_str::<Value>(l).ok())
        .filter(|v| v["reason"] == json!("compiler-artifact"))
        .collect()
}

/// `(name, kind)` of the artifacts under `folder` that Cargo compiled (not `Fresh`).
fn compiled_in(list: &[Value], folder: &Path) -> Vec<(String, String)> {
    let mut out: Vec<(String, String)> = list
        .iter()
        .filter(|a| a["fresh"] == json!(false))
        .filter(|a| {
            a["manifest_path"]
                .as_str()
                .is_some_and(|m| Path::new(m).starts_with(folder))
        })
        .map(|a| {
            (
                a["target"]["name"].as_str().unwrap().to_string(),
                a["target"]["kind"][0].as_str().unwrap().to_string(),
            )
        })
        .collect();
    out.sort();
    out
}

fn contains_bytes(hay: &[u8], needle: &[u8]) -> bool {
    hay.windows(needle.len()).any(|w| w == needle)
}

impl Fixture {
    fn slot(&self, name: &str) -> PathBuf {
        self.root.join(".claude/worktrees").join(name)
    }

    fn lane_str(&self) -> &str {
        self.lane.to_str().unwrap()
    }

    /// The environment the child needs to find the test run's own `cargo` and `rustc`: the
    /// child's environment is empty, and a rustup proxy cannot work in a temporary HOME.
    fn toolchain(&self) -> Vec<(String, String)> {
        let cargo = cargo_path();
        let bin = Path::new(&cargo)
            .parent()
            .filter(|p| !p.as_os_str().is_empty())
            .map(|p| format!("{}:", p.display()))
            .unwrap_or_default();
        vec![
            ("CARGO".into(), cargo),
            ("PATH".into(), format!("{bin}/usr/bin:/bin:/usr/sbin:/sbin")),
        ]
    }

    fn command(&self, op: &str, args: &[&str], env: &[(&str, &str)]) -> Command {
        let mut cmd = self.home.command(&[op]);
        cmd.current_dir(&self.root)
            .args(args)
            .args(["--project", self.root.to_str().unwrap()]);
        for (k, v) in self.toolchain() {
            cmd.env(k, v);
        }
        for (k, v) in env {
            cmd.env(k, v);
        }
        cmd
    }

    fn json(&self, op: &str, args: &[&str], env: &[(&str, &str)]) -> (i32, Value) {
        let out = self.command(op, args, env).output().unwrap();
        let v: Value = serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
            panic!(
                "{op} {args:?}: not JSON: {} / {}",
                String::from_utf8_lossy(&out.stdout),
                String::from_utf8_lossy(&out.stderr)
            )
        });
        (out.status.code().unwrap_or(-1), v)
    }

    /// `audit-prep CARD --head HEAD` with the heavy slot folder and floor of this HOME.
    fn prep_with(
        &self,
        card: &str,
        head: &str,
        extra: &[&str],
        env: &[(&str, &str)],
    ) -> (i32, Value) {
        let slots = self.home.path().join("heavy-slots");
        fs::create_dir_all(&slots).unwrap();
        let mut args = vec![
            card,
            "--head",
            head,
            "--slot-dir",
            slots.to_str().unwrap(),
            "--max",
            "2",
            "--min-free-gb",
            "0",
        ];
        args.extend_from_slice(extra);
        self.json("audit-prep", &args, env)
    }

    fn prep(&self, card: &str, head: &str, extra: &[&str]) -> (i32, Value) {
        self.prep_with(card, head, extra, &[])
    }

    fn prep_ok(&self, card: &str, head: &str, extra: &[&str]) -> Value {
        let (code, v) = self.prep(card, head, extra);
        assert_eq!((code, &v["ok"]), (0, &json!(true)), "{card} {extra:?}: {v}");
        v["result"].clone()
    }

    fn done(&self, card: &str) -> (i32, Value) {
        self.json("audit-done", &[card], &[])
    }

    fn done_ok(&self, card: &str) -> Value {
        let (code, v) = self.done(card);
        assert_eq!((code, &v["ok"]), (0, &json!(true)), "{card}: {v}");
        v["result"].clone()
    }

    /// A refusal: non-zero exit, `invalid_input`, and a message holding every word of `words`.
    fn refused(&self, op: &str, args: &[&str], words: &[&str]) -> String {
        let (code, v) = self.json(op, args, &[]);
        assert_ne!(code, 0, "{op} {args:?}: {v}");
        assert_eq!(
            v["error"]["kind"],
            json!("invalid_input"),
            "{op} {args:?}: {v}"
        );
        let m = v["error"]["message"].as_str().unwrap().to_string();
        for w in words {
            assert!(m.contains(w), "{op} {args:?}: {w:?} not in {m}");
        }
        m
    }

    /// `audit-slot` as another auditor, to occupy slots.
    fn hold(&self, slot: &str, holder: &str) {
        let (code, v) = self.json(
            "audit-slot",
            &["prep", slot, "--head", &self.first, "--holder", holder],
            &[],
        );
        assert_eq!((code, &v["ok"]), (0, &json!(true)), "{slot} {holder}: {v}");
    }

    /// The lane builds its test targets, as a lane does before it says READY.
    fn build_lane(&self) -> String {
        cargo_ok(
            &self.lane,
            &[
                "test",
                "--no-run",
                "--offline",
                "--locked",
                "--message-format",
                "json",
                "--test",
                "where",
            ],
        )
    }

    fn holder_record(&self, slot: &str) -> PathBuf {
        self.root
            .join(".git/moai-audit-slots")
            .join(format!("{slot}.json"))
    }

    /// Log files `audit-prep` left in a slot's `target/`.
    fn prep_logs(&self, slot: &str) -> Vec<String> {
        fs::read_dir(self.slot(slot).join("target"))
            .map(|d| {
                d.flatten()
                    .map(|e| e.file_name().to_string_lossy().to_string())
                    .filter(|n| n.starts_with("audit-prep-"))
                    .collect()
            })
            .unwrap_or_default()
    }
}

fn detached_at(slot: &Path, head: &str) {
    assert_eq!(git(slot, &["rev-parse", "HEAD"]), head);
    assert!(!git_raw(slot, &["symbolic-ref", "-q", "HEAD"]).success);
    assert_eq!(
        git(slot, &["status", "--porcelain", "--untracked-files=all"]),
        ""
    );
}

// ---- the card takes a slot (AC-1: REQ-1, REQ-2, REQ-3, REQ-4)

#[cfg(unix)]
#[test]
fn prep_takes_a_slot_for_the_card_again_the_same_one_and_refuses_when_all_are_held() {
    let r = fixture();
    r.build_lane();
    let first = r.prep_ok("C-1", &r.second, &["--lane", r.lane_str()]);
    assert_eq!(first["slot"], json!("audit-1"), "{first}");
    assert_eq!(first["holder"]["holder"], json!("C-1"), "{first}");
    assert_eq!(first["build"]["source"], json!("lane"), "{first}");
    assert_eq!(first["cargo"]["exit"], json!(0), "{first}");
    detached_at(&r.slot("audit-1"), &r.second);
    let again = r.prep_ok("C-1", &r.second, &["--lane", r.lane_str()]);
    assert_eq!(again["slot"], json!("audit-1"), "{again}");
    // Two other auditors take the other two slots; C-9 finds none free and nothing runs.
    r.hold("audit-2", "h2");
    r.hold("audit-3", "h3");
    let m = r.refused(
        "audit-prep",
        &[
            "C-9",
            "--head",
            &r.second,
            "--slot-dir",
            r.home.path().join("heavy-slots").to_str().unwrap(),
            "--max",
            "2",
            "--min-free-gb",
            "0",
        ],
        &["all audit slots are held", "C-1", "h2", "h3"],
    );
    assert!(m.contains("nothing was changed"), "{m}");
    assert!(r.prep_logs("audit-1").iter().all(|n| !n.contains("C-9")));
}

#[cfg(unix)]
#[test]
fn prep_prepares_the_slot_the_card_already_holds_not_the_first_free_one() {
    let r = fixture();
    r.hold("audit-1", "h1");
    let a = r.prep_ok("C-1", &r.second, &[]);
    assert_eq!(a["slot"], json!("audit-2"), "{a}");
    // audit-1 is free again; the card still gets the slot it holds.
    let (code, v) = r.json("audit-slot", &["done", "audit-1", "--holder", "h1"], &[]);
    assert_eq!((code, &v["ok"]), (0, &json!(true)), "{v}");
    let b = r.prep_ok("C-1", &r.second, &[]);
    assert_eq!(b["slot"], json!("audit-2"), "{b}");
    assert!(!r.holder_record("audit-1").exists());
}

#[cfg(unix)]
#[test]
fn prep_refuses_what_audit_slot_refuses_and_runs_no_build() {
    let r = fixture();
    let slots = r.home.path().join("heavy-slots");
    let common = [
        "--slot-dir",
        slots.to_str().unwrap(),
        "--max",
        "2",
        "--min-free-gb",
        "0",
    ];
    let with = |first: &[&str]| -> Vec<String> {
        first
            .iter()
            .chain(common.iter())
            .map(|s| s.to_string())
            .collect()
    };
    let go = |first: &[&str], words: &[&str]| {
        let args = with(first);
        let args: Vec<&str> = args.iter().map(String::as_str).collect();
        r.refused("audit-prep", &args, words);
    };
    // Not a card name, a short head, a lane that is not a worktree.
    go(&["nonsense", "--head", &r.second], &["card"]);
    go(&["C-1", "--head", &r.second[..12]], &["40"]);
    go(
        &[
            "C-1",
            "--head",
            &r.second,
            "--lane",
            r.root.join("nope").to_str().unwrap(),
        ],
        &["lane"],
    );
    assert!(!r.slot("audit-1").exists(), "nothing was added");
    // A slot that holds an earlier audit's file.
    r.hold("audit-1", "h1");
    r.json("audit-slot", &["done", "audit-1", "--holder", "h1"], &[]);
    fs::write(r.slot("audit-1").join("stray.txt"), "left\n").unwrap();
    go(&["C-1", "--head", &r.second], &["stray.txt"]);
    assert!(r.prep_logs("audit-1").is_empty(), "no build ran");
}

// ---- one incremental build (AC-2: REQ-5, REQ-6, REQ-7)

/// The dependency outside the repository is not compiled again: the answer lists the two
/// workspace units compiled once each and the dependency as fresh; a third `--test` is refused.
#[cfg(target_os = "macos")]
#[test]
fn one_build_compiles_the_workspace_units_once_and_every_dependency_fresh() {
    let r = fixture();
    r.build_lane();
    let a = r.prep_ok(
        "C-1",
        &r.second,
        &["--lane", r.lane_str(), "--test", "where"],
    );
    let c = &a["cargo"];
    assert_eq!(c["exit"], json!(0), "{a}");
    assert_eq!(c["compiled_dependencies"], json!(0), "{a}");
    assert_eq!(c["compiled_workspace"], json!(2), "{a}");
    let units = c["units"].as_array().unwrap();
    let mut built: Vec<(&str, &str)> = units
        .iter()
        .filter(|u| u["workspace"] == json!(true) && u["fresh"] == json!(false))
        .map(|u| (u["name"].as_str().unwrap(), u["kind"].as_str().unwrap()))
        .collect();
    built.sort();
    assert_eq!(built, vec![("auditapp", "lib"), ("where", "test")], "{a}");
    assert!(
        units
            .iter()
            .any(|u| u["name"] == json!("auditdep") && u["fresh"] == json!(true)),
        "{a}"
    );
    assert!(
        r.slot("audit-1")
            .join("target")
            .join(c["log"].as_str().unwrap())
            .exists()
            || Path::new(c["log"].as_str().unwrap()).exists()
    );
    // Three named targets are over the cap: refused before anything changes.
    let r2 = fixture();
    let m = r2.refused(
        "audit-prep",
        &[
            "C-2", "--head", &r2.second, "--test", "a", "--test", "b", "--test", "c",
        ],
        &["at most two"],
    );
    assert!(m.contains("nothing was changed"), "{m}");
    assert!(!r2.slot("audit-1").exists());
}

// ---- a build that fails (AC-3: REQ-8)

#[cfg(unix)]
#[test]
fn a_build_that_does_not_compile_keeps_the_slot_held_and_names_the_errors() {
    let r = fixture();
    r.build_lane();
    let (code, v) = r.prep(
        "C-1",
        &r.second,
        &["--lane", r.lane_str(), "--test", "broken"],
    );
    assert_ne!(code, 0, "{v}");
    assert_eq!(v["ok"], json!(false), "{v}");
    let m = v["error"]["message"].as_str().unwrap();
    // The compiler's own sentence, not the word `broken` (Cargo's summary line has it too): the
    // rendered error comes only from the JSON messages the build reads.
    assert!(m.contains("error[E0308]"), "the compiler's error code: {m}");
    assert!(
        m.contains("mismatched types"),
        "the compiler's message: {m}"
    );
    assert!(m.contains("audit-1"), "{m}");
    let rec: Value =
        serde_json::from_slice(&fs::read(r.holder_record("audit-1")).unwrap()).unwrap();
    assert_eq!(rec["holder"], json!("C-1"), "the card still holds the slot");
    // The auditor still finishes it.
    let done = r.done_ok("C-1");
    assert_eq!(done["released"], json!(true), "{done}");
}

// ---- the rebuilt executable reads the slot (AC-4: REQ-9)

#[cfg(unix)]
#[test]
fn a_rebuilt_test_executable_reads_the_slot_and_holds_no_byte_of_the_lane_folder() {
    let r = fixture();
    let lane_build = r.build_lane();
    let lane_exe = artifacts(&lane_build)
        .into_iter()
        .find(|a| a["target"]["name"] == json!("where"))
        .and_then(|a| a["executable"].as_str().map(PathBuf::from))
        .expect("the lane's test executable");
    let lane_bytes = fs::read(&lane_exe).unwrap();
    assert!(
        contains_bytes(&lane_bytes, r.lane_str().as_bytes()),
        "the check is meaningful: the lane's own executable holds the lane's folder"
    );
    let a = r.prep_ok(
        "C-1",
        &r.second,
        &["--lane", r.lane_str(), "--test", "where"],
    );
    let exe = a["cargo"]["units"]
        .as_array()
        .unwrap()
        .iter()
        .find(|u| u["name"] == json!("where") && u["kind"] == json!("test"))
        .and_then(|u| u["executable"].as_str())
        .map(PathBuf::from)
        .expect("the slot's test executable");
    assert!(exe.starts_with(r.slot("audit-1")), "{exe:?}");
    let out = Command::new(&exe).arg("--nocapture").output().unwrap();
    let text = String::from_utf8_lossy(&out.stdout).to_string();
    let slot = r.slot("audit-1");
    assert!(
        text.contains(&format!("MANIFEST={}", slot.display())),
        "{text}"
    );
    assert!(!text.contains(r.lane_str()), "{text}");
    let bytes = fs::read(&exe).unwrap();
    assert!(
        !contains_bytes(&bytes, r.lane_str().as_bytes()),
        "the rebuilt executable names the lane's folder"
    );
}

// ---- audit-done (AC-5: REQ-10, REQ-11, REQ-12)

#[cfg(unix)]
#[test]
fn audit_done_refuses_a_left_mutant_then_releases_and_takes_no_discard_or_force() {
    let r = fixture();
    r.prep_ok("C-1", &r.second, &[]);
    let slot = r.slot("audit-1");
    fs::write(slot.join("src/lib.rs"), "pub fn two() -> u32 { 3 }\n").unwrap();
    let m = r.refused("audit-done", &["C-1"], &["src/lib.rs"]);
    assert!(m.contains("nothing was changed"), "{m}");
    assert!(r.holder_record("audit-1").exists(), "the hold stays");
    // Neither option exists; the mutant stays where it is.
    for flag in ["--discard", "--force"] {
        let out = r
            .command("audit-done", &["C-1", flag], &[])
            .output()
            .unwrap();
        assert!(!out.status.success(), "{flag} must not be accepted");
        assert_eq!(
            fs::read_to_string(slot.join("src/lib.rs")).unwrap(),
            "pub fn two() -> u32 { 3 }\n"
        );
    }
    git(&slot, &["checkout", "--", "src/lib.rs"]);
    let done = r.done_ok("C-1");
    assert_eq!(done["slot"], json!("audit-1"), "{done}");
    assert_eq!(done["released"], json!(true), "{done}");
    assert_eq!(done["target_kept"], json!(true), "{done}");
    assert_eq!(
        git(&slot, &["status", "--porcelain", "--untracked-files=all"]),
        ""
    );
    // A card that holds no slot.
    r.refused("audit-done", &["C-1"], &["holds no audit slot"]);
}

// ---- a mutant recompiles only the workspace (AC-6: REQ-5)

#[cfg(target_os = "macos")]
#[test]
fn a_one_line_mutant_on_a_prepared_slot_recompiles_only_the_workspace_units() {
    let r = fixture();
    r.build_lane();
    let a = r.prep_ok(
        "C-1",
        &r.second,
        &["--lane", r.lane_str(), "--test", "where"],
    );
    // The prepared build itself used the lane's flags and compiled no dependency.
    assert_eq!(a["cargo"]["compiled_dependencies"], json!(0), "{a}");
    let slot = r.slot("audit-1");
    fs::write(
        slot.join("src/lib.rs"),
        "pub fn two() -> u32 { auditdep::one() + 2 }\n",
    )
    .unwrap();
    let out = cargo_ok(
        &slot,
        &[
            "test",
            "--no-run",
            "--offline",
            "--locked",
            "--message-format",
            "json",
            "--test",
            "where",
        ],
    );
    let list = artifacts(&out);
    let compiled = compiled_in(&list, &slot);
    assert!(
        compiled.contains(&("auditapp".to_string(), "lib".to_string())),
        "{compiled:?}"
    );
    let outside = list
        .iter()
        .filter(|a| a["fresh"] == json!(false))
        .filter(|a| !Path::new(a["manifest_path"].as_str().unwrap()).starts_with(&slot))
        .count();
    assert_eq!(outside, 0, "no dependency was compiled again");
}

// ---- the card is its own holder, in any session (D-6)

#[cfg(unix)]
#[test]
fn the_card_is_the_holder_whatever_session_finishes_it() {
    let r = fixture();
    let a = r.prep_with(
        "C-1",
        &r.second,
        &[],
        &[(
            "CLAUDE_CODE_SESSION_ID",
            "11111111-1111-1111-1111-111111111111",
        )],
    );
    assert_eq!((a.0, &a.1["ok"]), (0, &json!(true)), "{}", a.1);
    // The record names the card and no session, so no session can be the other holder.
    let rec: Value =
        serde_json::from_slice(&fs::read(r.holder_record("audit-1")).unwrap()).unwrap();
    assert_eq!(rec["holder"], json!("C-1"), "{rec}");
    assert!(rec["session"].is_null(), "{rec}");
    let (code, v) = r.json(
        "audit-done",
        &["C-1"],
        &[(
            "CLAUDE_CODE_SESSION_ID",
            "22222222-2222-2222-2222-222222222222",
        )],
    );
    assert_eq!((code, &v["ok"]), (0, &json!(true)), "{v}");
    assert_eq!(v["result"]["released"], json!(true), "{v}");
}

// ---- the repair of the first audit (B1-B4, N1)

/// The cap is two: two named targets are built, in one run (a third is refused above).
#[cfg(unix)]
#[test]
fn prep_builds_exactly_the_most_test_targets_the_cap_allows() {
    let r = fixture();
    let a = r.prep_ok("C-1", &r.second, &["--test", "where", "--test", "second"]);
    let c = &a["cargo"];
    assert_eq!(c["exit"], json!(0), "{a}");
    let args: Vec<&str> = c["args"]
        .as_array()
        .unwrap()
        .iter()
        .filter_map(Value::as_str)
        .collect();
    assert_eq!(
        args.iter().filter(|a| **a == "--test").count(),
        2,
        "both targets are named to cargo: {args:?}"
    );
    let tests: Vec<&str> = c["units"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|u| u["kind"] == json!("test") && u["workspace"] == json!(true))
        .filter_map(|u| u["name"].as_str())
        .collect();
    assert!(tests.contains(&"where") && tests.contains(&"second"), "{a}");
}

/// The build gets none of the caller's wrapper or target-folder variables: with them set to
/// places that cannot work or must stay empty, it still succeeds and writes only the slot.
#[cfg(unix)]
#[test]
fn the_build_drops_the_callers_wrapper_and_target_folder_variables() {
    let r = fixture();
    let elsewhere = r.home.path().join("elsewhere-target");
    let elsewhere_build = r.home.path().join("elsewhere-build-target");
    let wrapper = r.home.path().join("no-such-wrapper");
    let (code, v) = r.prep_with(
        "C-1",
        &r.second,
        &[],
        &[
            ("RUSTC_WRAPPER", wrapper.to_str().unwrap()),
            ("CARGO_BUILD_RUSTC_WRAPPER", wrapper.to_str().unwrap()),
            ("CARGO_TARGET_DIR", elsewhere.to_str().unwrap()),
            ("CARGO_BUILD_TARGET_DIR", elsewhere_build.to_str().unwrap()),
        ],
    );
    assert_eq!((code, &v["ok"]), (0, &json!(true)), "{v}");
    assert_eq!(v["result"]["cargo"]["exit"], json!(0), "{v}");
    assert!(
        r.slot("audit-1").join("target/debug").is_dir(),
        "the build wrote the slot's own target/"
    );
    assert!(!elsewhere.exists(), "CARGO_TARGET_DIR was not followed");
    assert!(
        !elsewhere_build.exists(),
        "CARGO_BUILD_TARGET_DIR was not followed"
    );
}

/// A slot whose `target/` is a link is refused before anything is written: the build and its
/// logs would land in the folder the link names (a lane's build).
#[cfg(unix)]
#[test]
fn prep_refuses_a_slot_whose_target_is_a_link_and_writes_nothing_through_it() {
    let r = fixture();
    r.hold("audit-1", "h1");
    let (code, v) = r.json("audit-slot", &["done", "audit-1", "--holder", "h1"], &[]);
    assert_eq!((code, &v["ok"]), (0, &json!(true)), "{v}");
    let victim = r.home.path().join("victim-target");
    fs::create_dir_all(&victim).unwrap();
    let link = r.slot("audit-1").join("target");
    assert!(!link.exists(), "a slot made with no build has no target/");
    std::os::unix::fs::symlink(&victim, &link).unwrap();
    let slots = r.home.path().join("heavy-slots");
    let m = r.refused(
        "audit-prep",
        &[
            "C-1",
            "--head",
            &r.second,
            "--slot-dir",
            slots.to_str().unwrap(),
            "--max",
            "2",
            "--min-free-gb",
            "0",
        ],
        &[
            "audit-prep refused",
            "target/",
            "link",
            "nothing was changed",
        ],
    );
    assert!(m.contains("audit-1"), "{m}");
    assert_eq!(
        fs::read_dir(&victim).unwrap().count(),
        0,
        "nothing was written through the link"
    );
    assert!(!r.holder_record("audit-1").exists(), "no hold was taken");
}

/// A line of cargo's output that is not UTF-8 does not end the read: the units after it are
/// still counted and the child is read to its end (an unread pipe would hold the slot).
#[cfg(unix)]
#[test]
fn a_build_line_that_is_not_utf8_does_not_end_the_read_of_the_build_messages() {
    use std::os::unix::fs::PermissionsExt;
    let r = fixture();
    let fake = r.home.path().join("fake-cargo");
    fs::write(
        &fake,
        concat!(
            "#!/bin/sh\n",
            "printf '{\"reason\":\"compiler-artifact\",\"manifest_path\":\"/elsewhere/dep/Cargo.toml\",\"target\":{\"name\":\"auditdep\",\"kind\":[\"lib\"]},\"fresh\":true,\"executable\":null}\\n'\n",
            "printf 'not utf8 \\377\\376 here\\n'\n",
            "printf '{\"reason\":\"compiler-artifact\",\"manifest_path\":\"%s/Cargo.toml\",\"target\":{\"name\":\"auditapp\",\"kind\":[\"lib\"]},\"fresh\":false,\"executable\":null}\\n' \"$PWD\"\n",
            "exit 0\n",
        ),
    )
    .unwrap();
    fs::set_permissions(&fake, fs::Permissions::from_mode(0o755)).unwrap();
    let (code, v) = r.prep_with("C-1", &r.second, &[], &[("CARGO", fake.to_str().unwrap())]);
    assert_eq!((code, &v["ok"]), (0, &json!(true)), "{v}");
    let c = &v["result"]["cargo"];
    assert_eq!(c["units"].as_array().unwrap().len(), 2, "{v}");
    assert_eq!(c["compiled"], json!(1), "{v}");
    assert_eq!(c["fresh"], json!(1), "{v}");
    assert_eq!(c["compiled_workspace"], json!(1), "{v}");
}
