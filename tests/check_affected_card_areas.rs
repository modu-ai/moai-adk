//! SPEC-20261125-check-affected-card-areas: `check affected --card` no longer falls to the full set
//! for a `mod.rs`, a name word that more than ten targets carry or a source path that reaches no
//! test. It takes the targets of the path's table area (`tests/areas.toml`), narrows the library
//! unit tests to the changed module, and keeps every wide path, the bare run and the lane set as
//! they were. Temporary git repositories and fake `cargo` and `rustfmt` scripts only.
mod common;
use common::git::{git, git_ok};
use serde_json::Value;
use std::collections::BTreeSet;
use std::fs;
use std::path::Path;
use std::process::{Command, Stdio};

const BIN: &str = crate::common::MOAI_BIN;

/// Three areas: `card` lists the words `card` and `ledger`, `spec` lists `check`, `update` lists
/// `update`. A word that no area lists (`zz`, `zzz`, `lonely`) belongs to none.
const AREAS: &str = "[areas]\ncard = [\"card\", \"ledger\"]\nhome = [\"home\"]\nspec = [\"spec\", \"check\"]\nupdate = [\"update\"]\n\n[solo]\nfiles = []\n";

fn put(root: &Path, rel: &str, text: &str) {
    let path = root.join(rel);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, text).unwrap();
}

/// A repository with an area table, twelve `card_*` targets (the word `card` is common), one
/// target per other area, eleven `zz_check_pipeline_*` targets that no area lists, and the source
/// files the tests change; one commit on top of the base that changes `changed`.
struct Fixture {
    dir: tempfile::TempDir,
    base: String,
    head: String,
}

const CARD_TARGETS: [&str; 12] = [
    "card_a_cli",
    "card_b_cli",
    "card_c_cli",
    "card_d_cli",
    "card_e_cli",
    "card_f_cli",
    "card_g_cli",
    "card_h_cli",
    "card_i_cli",
    "card_j_cli",
    "card_k_cli",
    "card_l_cli",
];

impl Fixture {
    fn new(table: Option<&str>, changed: &[(&str, &str)]) -> Fixture {
        let dir = tempfile::tempdir().unwrap();
        let p = dir.path();
        git_ok(p, &["init", "-q", "-b", "main"]);
        git_ok(p, &["config", "user.email", "t@example.invalid"]);
        git_ok(p, &["config", "user.name", "t"]);
        put(p, "README.md", "x\n");
        if let Some(text) = table {
            put(p, "tests/areas.toml", text);
        }
        for t in CARD_TARGETS {
            put(p, &format!("tests/{t}.rs"), "fn main() {}\n");
        }
        put(
            p,
            "tests/ledger_sweep.rs",
            "use moai_cli::ledger::sweep::run;\nfn main() { run(); }\n",
        );
        put(p, "tests/spec_check_cli.rs", "fn main() {}\n");
        put(p, "tests/update_command_cli.rs", "fn main() {}\n");
        put(p, "tests/home_cli.rs", "fn main() {}\n");
        put(
            p,
            "tests/pins_scans.rs",
            "// reads tools/repo_scans.json\nfn main() {}\n",
        );
        put(p, "tests/check_one.rs", "fn main() {}\n");
        for n in 0..11 {
            put(
                p,
                &format!("tests/zz_check_pipeline_{n}.rs"),
                "fn main() {}\n",
            );
        }
        for src in [
            "src/lib.rs",
            "src/main.rs",
            "src/lonely.rs",
            "src/check_pipeline.rs",
            "src/card/mod.rs",
            "src/card/work.rs",
            "src/card/zeta.rs",
            "src/ops/card.rs",
            "src/ledger/sweep.rs",
            "src/zzz/foo.rs",
            "src/card/land/check.rs",
            "src/update_models.rs",
            "src/home_style.rs",
            "src/self_update.rs",
            "src/gate.rs",
            "src/hook_engine.rs",
        ] {
            put(p, src, "// a\n");
        }
        put(p, "Cargo.lock", "# lock\n");
        put(p, "docs/note.md", "note\n");
        git_ok(p, &["add", "-A"]);
        git_ok(p, &["commit", "-q", "-m", "base"]);
        let base = git(p, &["rev-parse", "HEAD"]).trim().to_string();
        for (rel, text) in changed {
            put(p, rel, text);
        }
        git_ok(p, &["add", "-A"]);
        git_ok(p, &["commit", "-q", "-m", "change"]);
        let head = git(p, &["rev-parse", "HEAD"]).trim().to_string();
        Fixture { dir, base, head }
    }

    /// The usual fixture: the area table and one changed path.
    fn with(paths: &[&str]) -> Fixture {
        let changed: Vec<(&str, &str)> = paths.iter().map(|p| (*p, "// b\n")).collect();
        Fixture::new(Some(AREAS), &changed)
    }

    fn root(&self) -> &Path {
        self.dir.path()
    }

    fn range(&self) -> String {
        format!("{}..{}", self.base, self.head)
    }
}

/// A folder of fake `cargo` and `rustfmt` that exit 0 (so `cargo nextest --version` succeeds and
/// the planned runner is nextest), in front of the system's `git`.
fn fake_path() -> (tempfile::TempDir, String) {
    let dir = tempfile::tempdir().unwrap();
    for n in ["cargo", "rustfmt"] {
        let path = dir.path().join(n);
        fs::write(&path, "#!/bin/sh\nexit 0\n").unwrap();
        use std::os::unix::fs::PermissionsExt;
        fs::set_permissions(&path, fs::Permissions::from_mode(0o755)).unwrap();
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

fn affected(fx: &Fixture, extra: &[&str]) -> Value {
    let (_tools, path) = fake_path();
    let out = Command::new(BIN)
        .args(["check", "affected", &fx.range(), "--dir"])
        .arg(fx.root())
        .args(extra)
        .env("PATH", &path)
        .env_remove("MOAI_HOME")
        .env_remove("CARGO_TARGET_DIR")
        .stdin(Stdio::null())
        .output()
        .unwrap();
    let text = String::from_utf8_lossy(&out.stdout).to_string();
    let v: Value = serde_json::from_str(&text)
        .unwrap_or_else(|_| panic!("{text} / {}", String::from_utf8_lossy(&out.stderr)));
    assert_eq!(v["ok"], Value::Bool(true), "{text}");
    v["result"].clone()
}

fn names(v: &Value) -> BTreeSet<String> {
    v.as_array()
        .unwrap()
        .iter()
        .map(|s| s.as_str().unwrap().to_string())
        .collect()
}

fn set(of: &[&str]) -> BTreeSet<String> {
    of.iter().map(|s| s.to_string()).collect()
}

/// The words of the `-E` filterset of the planned `test` step, `None` without one.
fn filterset(r: &Value) -> Option<String> {
    let step = r["steps"]
        .as_array()
        .unwrap()
        .iter()
        .find(|s| s["label"] == "test")?;
    let args = step["command"].as_array().unwrap();
    let at = args.iter().position(|a| a == "-E")?;
    args.get(at + 1).and_then(Value::as_str).map(String::from)
}

fn by_of(r: &Value, path: &str) -> Vec<String> {
    r["card_picks"]
        .as_array()
        .unwrap()
        .iter()
        .find(|p| p["path"] == path)
        .map(|p| {
            p["by"]
                .as_array()
                .unwrap()
                .iter()
                .map(|s| s.as_str().unwrap().to_string())
                .collect()
        })
        .unwrap_or_default()
}

/// AC-1 (REQ-1, REQ-2): the three ways that fell to the full set take the path's area instead.
#[test]
fn card_areas_replace_the_full_set() {
    let card_area: BTreeSet<String> = CARD_TARGETS
        .iter()
        .map(|s| s.to_string())
        .chain(["ledger_sweep".to_string()])
        .collect();
    // A name word that twelve targets carry: the carrying targets inside the area.
    let fx = Fixture::with(&["src/ops/card.rs"]);
    let r = affected(&fx, &["--list", "--card"]);
    assert_eq!(r["selection"], "card", "{r}");
    assert!(r["card_wide"].is_null(), "{r}");
    assert_eq!(names(&r["rust_targets"]), set(&CARD_TARGETS), "{r}");
    let by = by_of(&r, "src/ops/card.rs");
    assert!(
        by.contains(&"area:card".to_string()) && by.contains(&"module:card".to_string()),
        "{by:?}"
    );
    // A source path that reaches no test, and a `mod.rs`: the whole area.
    for path in ["src/card/zeta.rs", "src/card/mod.rs"] {
        let fx = Fixture::with(&[path]);
        let r = affected(&fx, &["--list", "--card"]);
        assert_eq!(r["selection"], "card", "{path}: {r}");
        assert!(r["card_wide"].is_null(), "{path}: {r}");
        assert_eq!(names(&r["rust_targets"]), card_area, "{path}: {r}");
        assert!(
            by_of(&r, path).contains(&"area:card".to_string()),
            "{path}: {r}"
        );
        assert!(r["card_by"].is_string(), "{r}");
    }
    // The area `spec` for a file whose first word `check` it lists and that no test names.
    let fx = Fixture::with(&["src/check_other.rs"]);
    let r = affected(&fx, &["--list", "--card"]);
    assert_eq!(r["selection"], "card", "{r}");
    assert_eq!(
        names(&r["rust_targets"]),
        set(&["check_one", "spec_check_cli"]),
        "{r}"
    );
}

/// AC-2 (REQ-3): a module filter that picks nothing inside the area runs the whole area, and no
/// card selection is empty.
#[test]
fn an_empty_module_filter_runs_the_whole_area() {
    // Eleven targets carry the word `check_pipeline`, but none is in the area `spec` that the first word `check` lists.
    let fx = Fixture::with(&["src/check_pipeline.rs"]);
    let r = affected(&fx, &["--list", "--card"]);
    assert_eq!(r["selection"], "card", "{r}");
    assert_eq!(
        names(&r["rust_targets"]),
        set(&["check_one", "spec_check_cli"]),
        "{r}"
    );
    assert_eq!(
        names(&r["card_ignored_words"]),
        set(&["check_pipeline"]),
        "{r}"
    );
    assert_eq!(
        by_of(&r, "src/check_pipeline.rs"),
        ["units:check_pipeline".to_string(), "area:spec".to_string()],
        "{r}"
    );
    // The binary tests stay in every selection, so the planned run is never empty.
    let filter = filterset(&r).expect("a filterset");
    assert!(filter.contains("kind(bin)"), "{filter}");
}

/// AC-3 (REQ-4): a path with no area adds no integration test and does not fall to the full set.
#[test]
fn a_path_without_an_area_adds_no_integration_test() {
    for paths in [
        vec!["src/lib.rs"],
        vec!["src/main.rs"],
        vec!["src/lonely.rs"],
        vec!["src/zzz/foo.rs"],
        vec![
            "src/lib.rs",
            "src/main.rs",
            "src/lonely.rs",
            "src/zzz/foo.rs",
        ],
    ] {
        let fx = Fixture::with(&paths);
        let r = affected(&fx, &["--list", "--card"]);
        assert_eq!(r["selection"], "card", "{paths:?}: {r}");
        assert!(r["card_wide"].is_null(), "{paths:?}: {r}");
        assert!(names(&r["rust_targets"]).is_empty(), "{paths:?}: {r}");
    }
}

/// AC-4 (REQ-5, REQ-6, REQ-7): the library unit tests follow the changed module; with no changed
/// library module they stay whole.
#[test]
fn the_library_unit_tests_follow_the_changed_module() {
    let fx = Fixture::with(&["src/card/work.rs"]);
    let filter = filterset(&affected(&fx, &["--list", "--card"])).expect("a filterset");
    assert!(
        filter.starts_with("(kind(lib) & test(/^card::work::/)) | kind(bin)"),
        "{filter}"
    );
    // Two modules: one prefix each, in order; `mod.rs` names its folder.
    let fx = Fixture::with(&["src/card/mod.rs", "src/check_pipeline.rs"]);
    let filter = filterset(&affected(&fx, &["--list", "--card"])).expect("a filterset");
    assert!(
        filter.starts_with(
            "(kind(lib) & test(/^card::/)) | (kind(lib) & test(/^check_pipeline::/)) | kind(bin)"
        ),
        "{filter}"
    );
    // A crate root among the changes keeps the whole library.
    let fx = Fixture::with(&["src/card/work.rs", "src/lib.rs"]);
    let filter = filterset(&affected(&fx, &["--list", "--card"]));
    assert!(
        filter
            .as_deref()
            .is_none_or(|f| !f.contains("(kind(lib) &")),
        "{filter:?}"
    );
    // No changed library module (a document only): no narrowing filterset at all.
    let fx = Fixture::with(&["docs/note.md"]);
    let r = affected(&fx, &["--list", "--card"]);
    assert_eq!(r["selection"], "card", "{r}");
    assert!(filterset(&r).is_none(), "{r}");
    // The binary's `src/ops/` is no library module.
    let fx = Fixture::with(&["src/ops/card.rs"]);
    let filter = filterset(&affected(&fx, &["--list", "--card"]));
    assert!(
        filter
            .as_deref()
            .is_none_or(|f| !f.contains("(kind(lib) &")),
        "{filter:?}"
    );
}

/// AC-5 (REQ-8): the wide paths and an area table that cannot be read keep the full set.
#[test]
fn wide_paths_keep_the_full_set() {
    for (path, reason) in [
        ("Cargo.lock", "dependencies"),
        ("src/gate.rs", "audit floor"),
        ("src/self_update.rs", "audit floor"),
        ("src/hook_engine.rs", "permission"),
    ] {
        let fx = Fixture::with(&[path]);
        let r = affected(&fx, &["--list", "--card"]);
        assert_eq!(r["selection"], "full", "{path}: {r}");
        assert_eq!(r["card_wide"]["reason"], reason, "{path}: {r}");
    }
    // An absent table and a table with no area need an area they cannot give.
    for table in [None, Some("# no areas here\n")] {
        let fx = Fixture::new(table, &[("src/card/zeta.rs", "// b\n")]);
        let r = affected(&fx, &["--list", "--card"]);
        assert_eq!(r["selection"], "full", "{table:?}: {r}");
        assert_eq!(r["card_wide"]["reason"], "area table", "{table:?}: {r}");
    }
    // A path that a name word picks needs no table.
    let fx = Fixture::new(None, &[("src/ledger/sweep.rs", "// b\n")]);
    let r = affected(&fx, &["--list", "--card"]);
    assert_eq!(r["selection"], "card", "{r}");
    assert_eq!(names(&r["rust_targets"]), set(&["ledger_sweep"]), "{r}");
}

/// AC-6 (REQ-9, REQ-10): a card run says why its set is narrow, right after `selection=card`.
#[test]
fn the_summary_line_names_why_the_set_is_narrow() {
    let (_tools, path) = fake_path();
    let slots = slot_dir();
    let slot = slots.path().to_string_lossy().to_string();
    let fx = Fixture::with(&["src/card/zeta.rs"]);
    let out = Command::new(BIN)
        .args(["check", "affected", &fx.range(), "--dir"])
        .arg(fx.root())
        .args(["--card", "--slot-dir", slot.as_str()])
        .env("PATH", &path)
        .env_remove("MOAI_HOME")
        .env_remove("CARGO_TARGET_DIR")
        .stdin(Stdio::null())
        .output()
        .unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    let r = &v["result"];
    let line = r["summary"].as_str().unwrap();
    assert!(line.ends_with(" result=PASS"), "{line}");
    let words: Vec<&str> = line.split(' ').collect();
    let at = words.iter().position(|w| w.starts_with("scans=")).unwrap();
    assert_eq!(words[at + 1], "selection=card", "{line}");
    assert_eq!(
        words[at + 2],
        "card_by=area:card,units:card::zeta",
        "{line}"
    );
    assert_eq!(r["card_by"], "area:card,units:card::zeta", "{r}");
    // A card set with nothing to narrow says so.
    let fx = Fixture::with(&["docs/note.md"]);
    let out = Command::new(BIN)
        .args(["check", "affected", &fx.range(), "--dir"])
        .arg(fx.root())
        .args(["--card", "--slot-dir", slot.as_str()])
        .env("PATH", &path)
        .env_remove("MOAI_HOME")
        .env_remove("CARGO_TARGET_DIR")
        .stdin(Stdio::null())
        .output()
        .unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    let line = v["result"]["summary"].as_str().unwrap();
    assert!(line.contains(" selection=card card_by=none "), "{line}");
}

/// AC-7 (REQ-11): the bare command and `--lane` pick what they picked before this change for the
/// same range, and write no `card_by` word.
#[test]
fn the_bare_and_lane_selections_are_unchanged() {
    let fx = Fixture::with(&["src/ops/card.rs", "src/card/work.rs"]);
    // The area rule of the full and lane sets: every target whose first word is `card`.
    let bare = affected(&fx, &["--list"]);
    assert_eq!(bare["selection"], "full", "{bare}");
    assert_eq!(names(&bare["rust_targets"]), set(&CARD_TARGETS), "{bare}");
    assert!(
        bare["card_by"].is_null() && bare["card_picks"].is_null(),
        "{bare}"
    );
    let lane = affected(&fx, &["--list", "--lane"]);
    assert_eq!(lane["selection"], "lane", "{lane}");
    assert_eq!(names(&lane["rust_targets"]), set(&CARD_TARGETS), "{lane}");
    // Neither narrows the library unit tests: no filterset names a library prefix.
    for r in [&bare, &lane] {
        let filter = filterset(r);
        assert!(
            filter
                .as_deref()
                .is_none_or(|f| !f.contains("(kind(lib) &")),
            "{filter:?}"
        );
    }
}

/// AC-9 (REQ-13, REQ-14): the floor's ordinary source paths and the landing code take their area,
/// the scan table pins the tests that name it, and the gate stays wide.
#[test]
fn narrowed_floor_paths_take_their_area_and_the_scan_table_pins_its_tests() {
    let card_area: BTreeSet<String> = CARD_TARGETS
        .iter()
        .map(|s| s.to_string())
        .chain(["ledger_sweep".to_string()])
        .collect();
    for (path, want) in [
        ("src/update_models.rs", set(&["update_command_cli"])),
        ("src/home_style.rs", set(&["home_cli"])),
        ("src/card/land/check.rs", card_area),
    ] {
        let fx = Fixture::with(&[path]);
        let r = affected(&fx, &["--list", "--card"]);
        assert_eq!(r["selection"], "card", "{path}: {r}");
        assert!(r["card_wide"].is_null(), "{path}: {r}");
        assert_eq!(names(&r["rust_targets"]), want, "{path}: {r}");
    }
    // The scan table: the tests that name it, as a pinned path, never the full set.
    let fx = Fixture::new(
        Some(AREAS),
        &[("tools/repo_scans.json", "{\"about\":\"x\",\"scans\":[]}\n")],
    );
    let r = affected(&fx, &["--list", "--card"]);
    assert_eq!(r["selection"], "card", "{r}");
    assert_eq!(names(&r["rust_targets"]), set(&["pins_scans"]), "{r}");
    assert_eq!(
        by_of(&r, "tools/repo_scans.json"),
        ["pins:tools/repo_scans.json".to_string()],
        "{r}"
    );
    // The gate next to it stays wide.
    let fx = Fixture::with(&["src/gate.rs", "src/update_models.rs"]);
    let r = affected(&fx, &["--list", "--card"]);
    assert_eq!(r["selection"], "full", "{r}");
    assert_eq!(
        r["card_wide"]["paths"],
        serde_json::json!(["src/gate.rs"]),
        "{r}"
    );
}
