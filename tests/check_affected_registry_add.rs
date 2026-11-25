//! SPEC-20261125-registry-add-check-set: a change of `tools/repo_scans.json` that only adds lines
//! keeps `check affected --card` on the card set (with the added targets and the tests that read
//! the table); every other change of the registry, an unreadable text and any other wide path give
//! the full set as before. Temporary git repositories and fake `cargo` and `rustfmt` scripts only;
//! no file of the checked-in tree is read, so the registry needs no line for this file.
mod common;
use common::git::{git, git_ok};
use serde_json::{json, Value};
use std::fs;
use std::path::Path;
use std::process::{Command, Stdio};

const BIN: &str = crate::common::MOAI_BIN;

fn put(root: &Path, rel: &str, text: &str) {
    let path = root.join(rel);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, text).unwrap();
}

/// The registry of the base commit: an `always` scan with two rule files and a scan with two
/// checks, each naming one cargo target.
fn base_registry() -> Value {
    json!({
        "about": "fixture",
        "scans": [
            {
                "id": "alpha",
                "always": true,
                "checks": ["cargo test --offline --locked --test scan_alpha"],
                "scope": [{"prefix": "src/", "suffixes": [".rs"]}],
                "rules": ["tests/scan_alpha.rs", "tests/scan_alpha_two.rs"]
            },
            {
                "id": "beta",
                "checks": [
                    "cargo test --offline --locked --test scan_beta",
                    "cargo test --offline --locked --test scan_beta_two"
                ],
                "scope": [{"prefix": "docs/"}],
                "rules": ["docs/note.md"]
            }
        ]
    })
}

fn text(v: &Value) -> String {
    serde_json::to_string_pretty(v).unwrap() + "\n"
}

/// The base registry with a registration: a rule file, a second `--test` on a cargo line and a
/// scan added.
fn registered() -> Value {
    let mut v = base_registry();
    let scans = v["scans"].as_array_mut().unwrap();
    scans[0]["rules"]
        .as_array_mut()
        .unwrap()
        .push(json!("tests/scan_alpha_three.rs"));
    scans[1]["checks"][0] =
        json!("cargo test --offline --locked --test scan_beta --test scan_beta_three");
    scans.push(json!({
        "id": "gamma",
        "checks": ["cargo test --offline --locked --test scan_gamma"],
        "scope": [{"prefix": "docs/"}],
        "rules": ["docs/gamma.md"]
    }));
    v
}

/// A repository whose `tests/` holds twelve `card_other_*` targets (a common word), the scan
/// targets and a table test, a base commit with `base` as the registry (none: no registry file)
/// and one commit on top of it that writes `head` and the `other` files.
struct Fixture {
    dir: tempfile::TempDir,
    base: String,
    head: String,
}

impl Fixture {
    fn new(base: Option<&str>, head: &str, other: &[(&str, &str)]) -> Fixture {
        let dir = tempfile::tempdir().unwrap();
        let p = dir.path();
        git_ok(p, &["init", "-q", "-b", "main"]);
        git_ok(p, &["config", "user.email", "t@example.invalid"]);
        git_ok(p, &["config", "user.name", "t"]);
        put(p, "README.md", "x\n");
        put(p, "Cargo.toml", "[package]\nname = \"x\"\n");
        put(p, "docs/note.md", "note\n");
        for n in 0..12 {
            put(p, &format!("tests/card_other_{n}.rs"), "fn main() {}\n");
        }
        for t in [
            "scan_alpha",
            "scan_alpha_two",
            "scan_alpha_three",
            "scan_beta",
            "scan_beta_two",
            "scan_beta_three",
            "scan_gamma",
        ] {
            put(p, &format!("tests/{t}.rs"), "fn main() {}\n");
        }
        put(
            p,
            "tests/repo_scans_table.rs",
            "// reads tools/repo_scans.json\nfn main() {}\n",
        );
        if let Some(text) = base {
            put(p, "tools/repo_scans.json", text);
        }
        git_ok(p, &["add", "-A"]);
        git_ok(p, &["commit", "-q", "-m", "base"]);
        let base = git(p, &["rev-parse", "HEAD"]).trim().to_string();
        put(p, "tools/repo_scans.json", head);
        for (rel, body) in other {
            put(p, rel, body);
        }
        git_ok(p, &["add", "-A"]);
        git_ok(p, &["commit", "-q", "-m", "change"]);
        let head = git(p, &["rev-parse", "HEAD"]).trim().to_string();
        Fixture { dir, base, head }
    }

    /// A change of the registry from the base registry to `head`, nothing else.
    fn registry(head: &Value) -> Fixture {
        Fixture::new(Some(&text(&base_registry())), &text(head), &[])
    }

    fn root(&self) -> &Path {
        self.dir.path()
    }

    fn range(&self) -> String {
        format!("{}..{}", self.base, self.head)
    }
}

fn affected(fx: &Fixture, extra: &[&str], path: Option<&str>) -> (bool, Value) {
    let mut cmd = Command::new(BIN);
    cmd.args(["check", "affected", &fx.range(), "--dir"])
        .arg(fx.root())
        .args(extra)
        .env_remove("MOAI_HOME")
        .env_remove("CARGO_TARGET_DIR")
        .stdin(Stdio::null());
    if let Some(p) = path {
        cmd.env("PATH", p);
    }
    let out = cmd.output().unwrap();
    let text = String::from_utf8_lossy(&out.stdout).to_string();
    let v: Value = serde_json::from_str(&text)
        .unwrap_or_else(|_| panic!("{text} / {}", String::from_utf8_lossy(&out.stderr)));
    let ok = v["ok"] == Value::Bool(true);
    (
        ok,
        if ok {
            v["result"].clone()
        } else {
            v["error"].clone()
        },
    )
}

fn names(v: &Value) -> Vec<String> {
    v.as_array()
        .unwrap()
        .iter()
        .map(|s| s.as_str().unwrap().to_string())
        .collect()
}

const REGISTRY: &str = "tools/repo_scans.json";

/// The answer is the full set, with `reason` as the first wide reason and `path` among its paths.
fn assert_full(r: &Value, reason: &str, path: &str, what: &str) {
    assert_eq!(r["selection"], "full", "{what}: {r}");
    assert_eq!(r["card_wide"]["reason"], reason, "{what}: {r}");
    assert!(
        names(&r["card_wide"]["paths"]).contains(&path.to_string()),
        "{what}: {r}"
    );
}

#[test]
fn an_added_line_keeps_the_card_set_and_adds_its_targets() {
    let fx = Fixture::registry(&registered());
    let (ok, r) = affected(&fx, &["--list", "--card"], None);
    assert!(ok, "{r}");
    assert_eq!(r["selection"], "card", "{r}");
    assert!(r["card_wide"].is_null(), "{r}");
    let rust = names(&r["rust_targets"]);
    // The target the new `--test` names, the target of the added scan and the test that reads the
    // table (REQ-2); a target the base text already listed is not added by this rule.
    for want in ["scan_beta_three", "scan_gamma", "repo_scans_table"] {
        assert!(rust.contains(&want.to_string()), "{want} in {rust:?}");
    }
    for old in ["scan_beta", "scan_beta_two"] {
        assert!(!rust.contains(&old.to_string()), "{old} in {rust:?}");
    }
    // The area rule of the full set would take every `card_*` target; the card set does not.
    assert!(rust.len() <= 6, "{} targets: {rust:?}", rust.len());
    for n in 0..12 {
        let t = format!("card_other_{n}");
        assert!(!rust.contains(&t), "{t} in {rust:?}");
    }
}

#[test]
fn an_edit_or_a_removal_still_gives_the_full_set() {
    let cases: Vec<(&str, Box<dyn Fn(&mut Value)>)> = vec![
        (
            "a scan removed",
            Box::new(|v| {
                v["scans"].as_array_mut().unwrap().remove(1);
            }),
        ),
        (
            "a check line removed",
            Box::new(|v| {
                v["scans"][1]["checks"].as_array_mut().unwrap().pop();
            }),
        ),
        (
            "a check line replaced",
            Box::new(|v| {
                v["scans"][1]["checks"][0] =
                    json!("cargo test --offline --locked --test scan_other");
            }),
        ),
        (
            "`always` cleared",
            Box::new(|v| {
                v["scans"][0].as_object_mut().unwrap().remove("always");
            }),
        ),
        (
            "a rule file removed",
            Box::new(|v| {
                v["scans"][0]["rules"].as_array_mut().unwrap().pop();
            }),
        ),
        (
            "a scope narrowed",
            Box::new(|v| {
                v["scans"][0]["scope"][0]["suffixes"] = json!([".toml"]);
            }),
        ),
        (
            "a text field edited",
            Box::new(|v| {
                v["about"] = json!("edited");
            }),
        ),
    ];
    for (what, edit) in cases {
        let mut head = base_registry();
        edit(&mut head);
        let fx = Fixture::registry(&head);
        let (ok, r) = affected(&fx, &["--list", "--card"], None);
        assert!(ok, "{what}: {r}");
        assert_full(&r, "audit floor", REGISTRY, what);
    }
}

#[test]
fn an_unreadable_registry_gives_the_full_set() {
    // The head text is not JSON.
    let fx = Fixture::new(Some(&text(&base_registry())), "{ not json\n", &[]);
    let (ok, r) = affected(&fx, &["--list", "--card"], None);
    assert!(ok, "{r}");
    assert_full(&r, "audit floor", REGISTRY, "head text is not JSON");
    // The base text is not JSON.
    let fx = Fixture::new(Some("{ not json\n"), &text(&registered()), &[]);
    let (ok, r) = affected(&fx, &["--list", "--card"], None);
    assert!(ok, "{r}");
    assert_full(&r, "audit floor", REGISTRY, "base text is not JSON");
    // The base has no registry: a whole file is not an addition to one.
    let fx = Fixture::new(None, &text(&registered()), &[]);
    let (ok, r) = affected(&fx, &["--list", "--card"], None);
    assert!(ok, "{r}");
    assert_full(&r, "audit floor", REGISTRY, "base has no registry");
}

#[test]
fn another_wide_path_still_gives_the_full_set() {
    let fx = Fixture::new(
        Some(&text(&base_registry())),
        &text(&registered()),
        &[("Cargo.toml", "[package]\nname = \"y\"\n")],
    );
    let (ok, r) = affected(&fx, &["--list", "--card"], None);
    assert!(ok, "{r}");
    assert_eq!(r["selection"], "full", "{r}");
    // The reason is the other path's, and the registry is not named as a wide path.
    assert_eq!(r["card_wide"]["reason"], "dependencies", "{r}");
    assert_eq!(names(&r["card_wide"]["paths"]), vec!["Cargo.toml"], "{r}");
}

/// A folder of fake `cargo` and `rustfmt` that exit 0, in front of the system's `git`.
fn fake_path() -> (tempfile::TempDir, String) {
    let dir = tempfile::tempdir().unwrap();
    for n in ["cargo", "rustfmt"] {
        let path = dir.path().join(n);
        fs::write(&path, "#!/bin/sh\nexit 0\n").unwrap();
        #[cfg(unix)]
        {
            use std::os::unix::fs::PermissionsExt;
            fs::set_permissions(&path, fs::Permissions::from_mode(0o755)).unwrap();
        }
    }
    let git_dir = Command::new("sh")
        .args(["-c", "dirname \"$(command -v git)\""])
        .output()
        .unwrap();
    let path = format!(
        "{}:{}:/usr/bin:/bin",
        dir.path().display(),
        String::from_utf8_lossy(&git_dir.stdout).trim()
    );
    (dir, path)
}

fn slot_dir() -> tempfile::TempDir {
    let d = tempfile::tempdir().unwrap();
    fs::write(
        d.path().join("limits.json"),
        r#"{"max":4,"min_free_gb":0,"max_swap_pct":100,"min_mem_free_pct":0}"#,
    )
    .unwrap();
    d
}

#[test]
fn the_list_and_a_run_pick_the_same_set() {
    let (_tools, path) = fake_path();
    let slots = slot_dir();
    let slot = slots.path().to_string_lossy().to_string();
    // A registration: the list says `card`, and the run's receipt says `selection=card`.
    let fx = Fixture::registry(&registered());
    let (ok, listed) = affected(&fx, &["--list", "--card"], None);
    assert!(ok, "{listed}");
    let (ok, run) = affected(&fx, &["--card", "--slot-dir", slot.as_str()], Some(&path));
    assert!(ok, "{run}");
    assert_eq!(listed["selection"], run["selection"], "{listed} / {run}");
    let line = run["summary"].as_str().unwrap();
    assert!(line.ends_with(" result=PASS"), "{line}");
    assert!(line.contains(" selection=card"), "{line}");
    // An edit of the registry: both fall back to the full set, and the receipt says nothing of a
    // selection.
    let mut edited = base_registry();
    edited["about"] = json!("edited");
    let fx = Fixture::registry(&edited);
    let (ok, listed) = affected(&fx, &["--list", "--card"], None);
    assert!(ok, "{listed}");
    let (ok, run) = affected(&fx, &["--card", "--slot-dir", slot.as_str()], Some(&path));
    assert!(ok, "{run}");
    assert_eq!(listed["selection"], "full", "{listed}");
    assert_eq!(run["selection"], "full", "{run}");
    assert!(
        !run["summary"].as_str().unwrap().contains("selection="),
        "{run}"
    );
}
