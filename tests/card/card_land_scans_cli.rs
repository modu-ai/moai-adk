//! `moai card land` and the repo-wide scans (SPEC-20261123-land-repo-scans): before the landing
//! moves the integration branch it works out, from `tools/repo_scans.json`, which scan tests the
//! merge tree needs, and it asks for their receipt: a step `scan:<id>` at exit 0 in the receipt
//! of a PASS that names the landing head. Every test runs on a temporary home and a temporary
//! repository whose integration branch is `integ`; nothing here touches the real ledger.
//!
//! The repository: the first commit holds the table, `head` is the card's change, `integ` points at
//! `tip` (a child of `head`, or one that adds a file), a lane commit sits on `head` and the
//! candidate is the merge of the lane on `tip` that the leader lands, made with `commit-tree` so
//! the working tree never changes.

use crate::common;

use common::home::TempHome;
use moai_cli::ledger;
use rusqlite::Connection;
use serde_json::{json, Value};
use std::cell::RefCell;
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Output, Stdio};

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
/// The lane's author and card holder, the independent reviewer, the lander (the leader).
const S_A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const S_AUD: &str = "cccccccc-0000-4000-8000-00000000000c";
const S_LEAD: &str = "dddddddd-0000-4000-8000-00000000000d";
const WORKTREE_KEY: &str = "0123456789ab";

const OBSERVE_PROFILE: &[u8] = b"{\"schema_version\": 1, \"project_id\": \"demo\", \"purpose\": \"fixture\", \"tools\": [\"rust\"], \"development_methods\": [\"sdd\"], \"hosts\": [\"claude\"], \"work_flow\": {\"flow\": \"m-flow\", \"integration_branch\": \"integ\", \"land_verdict\": \"observe\"}}";
const ENFORCE_PROFILE: &[u8] = b"{\"schema_version\": 1, \"project_id\": \"demo\", \"purpose\": \"fixture\", \"tools\": [\"rust\"], \"development_methods\": [\"sdd\"], \"hosts\": [\"claude\"], \"work_flow\": {\"flow\": \"m-flow\", \"integration_branch\": \"integ\", \"land_verdict\": \"enforce\"}}";
/// No `land_verdict` key: the default, which is observe.
const DEFAULT_PROFILE: &[u8] = b"{\"schema_version\": 1, \"project_id\": \"demo\", \"purpose\": \"fixture\", \"tools\": [\"rust\"], \"development_methods\": [\"sdd\"], \"hosts\": [\"claude\"], \"work_flow\": {\"flow\": \"m-flow\", \"integration_branch\": \"integ\"}}";

const AREA_CHECK: &str = "cargo test --offline --locked --test test_area_layout";
const ENV_CHECK: &str = "cargo test --offline --locked --test cargo_env_developer_dir tests_clear_child_environments_only_through_clean_env";
const CROSS_CHECK: &str = "cargo test --offline --locked --lib cross_rule";

/// The fixture's table: the two whole-tree scans with `always`, and one scan that hits only when
/// its rule file and a path of its scope change on different sides.
fn table() -> String {
    json!({"scans": [
        {"id": "test-area-layout", "always": true,
         "tests": ["tests/test_area_layout.rs::area_table_maps_every_real_test_file"],
         "checks": [AREA_CHECK],
         "rules": ["tests/test_area_layout.rs", "tests/areas.toml"],
         "scope": [{"prefix": "tests/", "suffixes": [".rs", ".toml"]}],
         "rejects": "a test file whose first word is not in tests/areas.toml"},
        {"id": "clean-env-only", "always": true,
         "tests": ["tests/cargo_env_developer_dir.rs::tests_clear_child_environments_only_through_clean_env"],
         "checks": [ENV_CHECK],
         "rules": ["tests/cargo_env_developer_dir.rs"],
         "scope": [{"prefix": "tests/", "suffixes": [".rs"]}],
         "rejects": "a test that clears a child environment itself"},
        {"id": "cross-scan",
         "tests": ["src/cross.rs::tests::cross_rule"],
         "checks": [CROSS_CHECK],
         "rules": ["src/rule.rs"],
         "scope": [{"prefix": "src/", "suffixes": [".rs"]}],
         "rejects": "a file in src/ that the rule file refuses"}
    ]})
    .to_string()
}

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

/// A commit message with the trailer block of `session`.
fn signed(subject: &str, session: &str) -> String {
    format!(
        "{subject}\n\nMoai-Spec: SPEC-20261123-fixture\nMoai-Session: {session}:{WORKTREE_KEY}\nMoai-Verified: none"
    )
}

/// A commit on top of `parent` that adds `path` with `content`; no ref moves.
fn put_on(project: &Path, parent: &str, path: &str, content: &str, message: &str) -> String {
    let blob_file = project.join("put-on-blob.tmp");
    fs::write(&blob_file, content).unwrap();
    let blob = git(project, &["hash-object", "-w", "put-on-blob.tmp"]);
    fs::remove_file(&blob_file).unwrap();
    let index = project.join(".git/put-on-index");
    let with_index = |args: &[&str]| -> String {
        let out = common::git::command()
            .arg("-C")
            .arg(project)
            .args(args)
            .env("GIT_INDEX_FILE", &index)
            .output()
            .unwrap();
        assert!(out.status.success(), "git {args:?}");
        String::from_utf8_lossy(&out.stdout).trim().to_string()
    };
    with_index(&["read-tree", parent]);
    with_index(&[
        "update-index",
        "--add",
        "--cacheinfo",
        &format!("100644,{blob},{path}"),
    ]);
    let tree = with_index(&["write-tree"]);
    fs::remove_file(&index).unwrap();
    git(
        project,
        &["commit-tree", &tree, "-p", parent, "-m", message],
    )
}

fn git_executable() -> PathBuf {
    let path = std::env::var_os("PATH").expect("PATH");
    std::env::split_paths(&path)
        .map(|dir| dir.join("git"))
        .find(|p| p.is_file())
        .expect("a git on PATH")
}

fn envelope(out: &Output) -> Value {
    serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
        panic!(
            "stdout is not JSON ({e}): {} {}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    })
}

/// Asserts exit 0 and a success envelope, and unwraps the result.
fn result(out: &Output) -> Value {
    assert_eq!(
        out.status.code(),
        Some(0),
        "{} {}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    let v = envelope(out);
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn dump_tables(conn: &Connection, tables: &[&str]) -> Vec<String> {
    let mut out = Vec::new();
    for table in tables {
        let mut st = conn
            .prepare(&format!("SELECT * FROM {table} ORDER BY 1, 2"))
            .unwrap();
        let n = st.column_count();
        let lines: Vec<String> = st
            .query_map([], |r| {
                let mut line = format!("{table}:");
                for i in 0..n {
                    line.push_str(&format!("{:?}|", r.get_ref(i)?));
                }
                Ok(line)
            })
            .unwrap()
            .map(Result::unwrap)
            .collect();
        out.extend(lines);
    }
    out
}

fn number(name: &str) -> i64 {
    name.trim_start_matches("C-").parse().unwrap()
}

struct World {
    home: TempHome,
    project: PathBuf,
    p: String,
    profile: RefCell<Vec<u8>>,
    git_bin: PathBuf,
    head: String,
    tip: String,
}

impl World {
    /// A repository whose first commit holds `table` (none: no table at all); the tip adds
    /// `tip_adds` on top of the card's change when it is given.
    fn new(table: Option<&str>, tip_adds: Option<(&str, &str)>, profile: &[u8]) -> Self {
        let home = TempHome::new();
        let project = home.project("demo");
        ledger::register(home.path(), "demo").unwrap();
        git(&project, &["init", "-q", "-b", "main"]);
        fs::write(project.join("a.txt"), "a\n").unwrap();
        git(&project, &["add", "a.txt"]);
        if let Some(text) = table {
            fs::create_dir_all(project.join("tools")).unwrap();
            fs::write(project.join("tools/repo_scans.json"), text).unwrap();
            git(&project, &["add", "tools/repo_scans.json"]);
        }
        git(&project, &["commit", "-q", "-m", "init"]);
        fs::write(project.join("b.txt"), "b\n").unwrap();
        git(&project, &["add", "b.txt"]);
        // Signed by the lane session: a PASS of a head the branch already holds names the range
        // the head added, and the verdict check needs the authors of it.
        git(
            &project,
            &["commit", "-q", "-m", &signed("the card's change", S_A)],
        );
        let head = git(&project, &["rev-parse", "HEAD"]);
        let tip = match tip_adds {
            Some((path, content)) => put_on(&project, &head, path, content, "integration moved on"),
            None => {
                let tree = git(&project, &["rev-parse", &format!("{head}^{{tree}}")]);
                git(
                    &project,
                    &[
                        "commit-tree",
                        &tree,
                        "-p",
                        &head,
                        "-m",
                        "integration moved on",
                    ],
                )
            }
        };
        git(&project, &["update-ref", "refs/heads/integ", &tip]);
        World {
            p: project.to_str().unwrap().to_string(),
            home,
            project,
            profile: RefCell::new(profile.to_vec()),
            git_bin: git_executable(),
            head,
            tip,
        }
    }

    /// The common case: the fixture table, the tip a plain child of the card's change.
    fn with_table(profile: &[u8]) -> Self {
        World::new(Some(&table()), None, profile)
    }

    fn set_profile(&self, bytes: &[u8]) {
        *self.profile.borrow_mut() = bytes.to_vec();
    }

    fn card(&self, session: Option<&str>, args: &[&str]) -> Output {
        let cfg = moai_cli::state_root::path(&self.project, "project.yaml");
        fs::create_dir_all(cfg.parent().unwrap()).unwrap();
        fs::write(&cfg, &*self.profile.borrow()).unwrap();
        let mut all = vec!["card"];
        all.extend_from_slice(args);
        all.extend_from_slice(&["--project", &self.p]);
        let mut cmd: Command = self.home.command(&all);
        cmd.env("MOAI_GIT_BIN", &self.git_bin);
        if let Some(id) = session {
            cmd.env(SESSION_VAR, id);
        }
        cmd.stdin(Stdio::null());
        cmd.output().expect("run moai")
    }

    /// A task card with one open criterion, taken by `S_A` and sent to review; its name.
    fn review_card(&self) -> String {
        let added = result(&self.card(
            None,
            &[
                "add",
                "--title",
                "fixture card",
                "--kind",
                "task",
                "--scope",
                "apps/a/",
                "--ac",
                "one",
            ],
        ));
        let name = added["name"].as_str().unwrap().to_string();
        result(&self.card(None, &["ready", &name]));
        result(&self.card(Some(S_A), &["take", &name]));
        result(&self.card(Some(S_A), &["review", &name, "--generation", "1"]));
        name
    }

    /// A lane commit on the card's change that adds `path`, written by the card's holder.
    fn lane_adding(&self, path: &str) -> String {
        put_on(
            &self.project,
            &self.head,
            path,
            "lane\n",
            &signed("the lane change", S_A),
        )
    }

    /// A lane commit that adds `path` straight on the tip: landing it moves the branch by a
    /// fast-forward, with no merge commit on the first-parent line.
    fn fast_forward_adding(&self, path: &str) -> String {
        put_on(
            &self.project,
            &self.tip,
            path,
            "lane\n",
            &signed("the lane change on the tip", S_A),
        )
    }

    /// The commit the leader lands: the lane on the tip, the tree git makes of the two, signed by
    /// the lander.
    fn join(&self, lane: &str) -> String {
        let tree = git(
            &self.project,
            &["merge-tree", "--write-tree", &self.tip, lane],
        );
        git(
            &self.project,
            &[
                "commit-tree",
                &tree,
                "-p",
                &self.tip,
                "-p",
                lane,
                "-m",
                &signed("merge the lane on the tip", S_LEAD),
            ],
        )
    }

    /// A lane that adds `path`, joined on the tip: the candidate the leader lands.
    fn candidate_adding(&self, path: &str) -> String {
        let lane = self.lane_adding(path);
        self.join(&lane)
    }

    /// One PASS verdict of `card` on `head` at tier 1 by the independent reviewer, written as the
    /// `card verdict` command writes it; the receipt names `head` and the steps `(name, exit)`.
    fn pass(&self, card: &str, head: &str, steps: &[(&str, i64)]) {
        self.pass_naming(card, head, head, 1, steps);
    }

    /// [`World::pass`] at tier 0, where the verdict check asks for no receipt, so what the scan
    /// check reads from the receipt is judged by the scan check alone.
    fn pass_tier0(&self, card: &str, head: &str, receipt_head: &str, steps: &[(&str, i64)]) {
        self.pass_naming(card, head, receipt_head, 0, steps);
    }

    /// [`World::pass`] with the head the receipt names and the tier given.
    fn pass_naming(
        &self,
        card: &str,
        head: &str,
        receipt_head: &str,
        tier: u8,
        steps: &[(&str, i64)],
    ) {
        let conn =
            Connection::open(ledger::ledger_path(self.home.path(), "demo").unwrap()).unwrap();
        let rows = self.count("SELECT count(*) FROM event");
        let steps: Vec<Value> = steps
            .iter()
            .map(|(name, exit)| json!({"name": name, "exit": exit}))
            .collect();
        // The range the head added: from the tip, or from the first parent of a head that the
        // branch already holds (the card's own change).
        let base = if head == self.head {
            git(&self.project, &["rev-parse", &format!("{head}^")])
        } else {
            self.tip.clone()
        };
        let payload = json!({
            "head": head, "base": base, "verdict": "pass", "tier": tier,
            "evidence": "the fixture's evidence",
            "receipt": {"head": receipt_head, "steps": steps},
            "reviewer": S_AUD,
        });
        conn.execute(
            "INSERT INTO event (event_key, at, actor, card_id, kind, payload) \
             VALUES (?1, unixepoch(), ?2, ?3, 'card.verdict', ?4)",
            rusqlite::params![
                format!("test-verdict-{rows}"),
                S_AUD,
                number(card),
                payload.to_string()
            ],
        )
        .unwrap();
    }

    /// `card land` as the leader session for `head`.
    fn land(&self, card: &str, head: &str) -> Output {
        self.card(
            Some(S_LEAD),
            &[
                "land",
                card,
                "--head",
                head,
                "--met",
                "1=the test",
                "--by",
                "leader",
            ],
        )
    }

    fn integ(&self) -> String {
        git(&self.project, &["rev-parse", "refs/heads/integ"])
    }

    fn dump(&self) -> Vec<String> {
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        dump_tables(
            l.conn(),
            &[
                "card",
                "card_ac",
                "card_note",
                "card_link",
                "reservation",
                "reservation_scope",
                "event",
            ],
        )
    }

    fn count(&self, sql: &str) -> i64 {
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        l.conn().query_row(sql, [], |r| r.get(0)).unwrap()
    }

    fn events(&self, card: &str, kind: &str) -> i64 {
        self.count(&format!(
            "SELECT count(*) FROM event WHERE card_id = {} AND kind = '{kind}'",
            number(card)
        ))
    }

    /// Asserts a refusal that wrote nothing and moved nothing: exit 1, an `invalid_input` error
    /// whose message holds `code`, every table as it was and `integ` at the tip. Returns the
    /// message.
    fn refuse(&self, out: &Output, code: &str, before: &[String]) -> String {
        assert_eq!(
            out.status.code(),
            Some(1),
            "{} {}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        );
        let v = envelope(out);
        assert_eq!(v["ok"], json!(false), "{v}");
        assert_eq!(v["error"]["kind"], json!("invalid_input"), "{v}");
        let message = v["error"]["message"].as_str().unwrap().to_string();
        assert!(message.contains(code), "the reason lacks {code}: {message}");
        assert_eq!(&self.dump(), before, "the refused call wrote to the ledger");
        assert_eq!(self.integ(), self.tip, "the refused call moved integ");
        message
    }
}

/// The fmt and test steps every receipt of the fixture carries.
const BASE_STEPS: [(&str, i64); 2] = [("fmt", 0), ("test", 0)];

fn with_steps<'a>(extra: &[(&'a str, i64)]) -> Vec<(&'a str, i64)> {
    let mut steps = BASE_STEPS.to_vec();
    steps.extend_from_slice(extra);
    steps
}

/// AC-1 (REQ-1, REQ-3, REQ-5, REQ-10): a head that adds a test file whose first word is not in
/// `tests/areas.toml` refuses with `scans_missing` that names the area scan and its command, and
/// nothing is written, not even a start row.
#[test]
fn ac1_a_new_test_file_needs_the_area_layout_scan_receipt() {
    let w = World::with_table(ENFORCE_PROFILE);
    let card = w.review_card();
    let cand = w.candidate_adding("tests/branch_gc_cli.rs");
    w.pass(&card, &cand, &BASE_STEPS);
    let before = w.dump();
    let out = w.land(&card, &cand);
    let message = w.refuse(&out, "scans_missing", &before);
    assert!(message.contains("test-area-layout"), "{message}");
    assert!(message.contains(AREA_CHECK), "{message}");
    assert!(message.contains("scan:test-area-layout"), "{message}");
    assert_eq!(w.events(&card, "card.land_started"), 0);
}

/// AC-2 (REQ-1, REQ-10): a head that adds a test file that clears the environment itself needs the
/// clean-env scan as well, each scan with its own command on the line.
#[test]
fn ac2_a_new_test_file_needs_the_clean_env_scan_too() {
    let w = World::with_table(ENFORCE_PROFILE);
    let card = w.review_card();
    let cand = w.candidate_adding("tests/fake_exe_marker.rs");
    w.pass(&card, &cand, &BASE_STEPS);
    let before = w.dump();
    let out = w.land(&card, &cand);
    let message = w.refuse(&out, "scans_missing", &before);
    for needle in [
        "test-area-layout",
        AREA_CHECK,
        "clean-env-only",
        ENV_CHECK,
        "scan:clean-env-only",
    ] {
        assert!(message.contains(needle), "{needle}: {message}");
    }
    assert!(!message.contains('\n'), "one line: {message}");
}

/// AC-3 (REQ-4, REQ-5): the receipt of a PASS of the landing head with each needed scan at exit 0
/// lets the landing move and close; a step at exit 1, a receipt of another head and a receipt that
/// lacks one scan each refuse.
#[test]
fn ac3_the_scan_steps_of_the_landing_head_cover_the_scans() {
    let both = [("scan:test-area-layout", 0), ("scan:clean-env-only", 0)];
    let w = World::with_table(ENFORCE_PROFILE);
    let card = w.review_card();
    let cand = w.candidate_adding("tests/branch_gc_cli.rs");
    w.pass(&card, &cand, &with_steps(&both));
    let got = result(&w.land(&card, &cand));
    assert_eq!(got["moved"], json!(true), "{got}");
    assert_eq!(got["closed"], json!(true), "{got}");
    assert_eq!(w.integ(), cand);
    assert_eq!(got["scans"]["missing"], json!([]), "{got}");
    let required: Vec<&str> = got["scans"]["required"]
        .as_array()
        .unwrap()
        .iter()
        .map(|s| s["id"].as_str().unwrap())
        .collect();
    assert_eq!(required, ["test-area-layout", "clean-env-only"], "{got}");
    assert!(got["scans"]["required"]
        .as_array()
        .unwrap()
        .iter()
        .all(|s| s["covered"] == json!(true)));

    // A scan step at exit 1 does not cover its scan (tier 0: the verdict check reads no receipt).
    let w = World::with_table(ENFORCE_PROFILE);
    let card = w.review_card();
    let cand = w.candidate_adding("tests/branch_gc_cli.rs");
    w.pass_tier0(
        &card,
        &cand,
        &cand,
        &with_steps(&[("scan:test-area-layout", 1), ("scan:clean-env-only", 0)]),
    );
    let before = w.dump();
    let message = w.refuse(&w.land(&card, &cand), "scans_missing", &before);
    assert!(message.contains("test-area-layout"), "{message}");
    assert!(!message.contains("clean-env-only"), "{message}");

    // A receipt of another head than the verdict's covers nothing.
    let w = World::with_table(ENFORCE_PROFILE);
    let card = w.review_card();
    let cand = w.candidate_adding("tests/branch_gc_cli.rs");
    w.pass_tier0(&card, &cand, &w.tip, &with_steps(&both));
    let before = w.dump();
    let message = w.refuse(&w.land(&card, &cand), "scans_missing", &before);
    assert!(message.contains("test-area-layout"), "{message}");
    assert!(message.contains("clean-env-only"), "{message}");

    // At tier 1 the verdict check itself asks every receipt step for exit 0, so a red scan step
    // is refused there before the scan check.
    let w = World::with_table(ENFORCE_PROFILE);
    let card = w.review_card();
    let cand = w.candidate_adding("tests/branch_gc_cli.rs");
    w.pass(
        &card,
        &cand,
        &with_steps(&[("scan:test-area-layout", 1), ("scan:clean-env-only", 0)]),
    );
    let before = w.dump();
    w.refuse(&w.land(&card, &cand), "receipt_missing", &before);

    // A receipt that lacks one scan.
    let w = World::with_table(ENFORCE_PROFILE);
    let card = w.review_card();
    let cand = w.candidate_adding("tests/branch_gc_cli.rs");
    w.pass(&card, &cand, &with_steps(&[("scan:test-area-layout", 0)]));
    let before = w.dump();
    let message = w.refuse(&w.land(&card, &cand), "scans_missing", &before);
    assert!(message.contains("clean-env-only"), "{message}");
    assert!(!message.contains("scan:test-area-layout"), "{message}");
}

/// AC-4 (REQ-3): a lane that changes a rule file while the tip changed a path of that scan's scope
/// is a cross hit and needs the scan; a docs-only head needs none and lands with no scan step.
#[test]
fn ac4_a_cross_hit_needs_its_scan_and_a_docs_only_head_needs_none() {
    let w = World::new(
        Some(&table()),
        Some(("src/new_scope.rs", "scope\n")),
        ENFORCE_PROFILE,
    );
    let card = w.review_card();
    let cand = w.candidate_adding("src/rule.rs");
    w.pass(&card, &cand, &BASE_STEPS);
    let before = w.dump();
    let message = w.refuse(&w.land(&card, &cand), "scans_missing", &before);
    assert!(message.contains("cross-scan"), "{message}");
    assert!(message.contains(CROSS_CHECK), "{message}");
    assert!(!message.contains("test-area-layout"), "{message}");
    w.pass(&card, &cand, &with_steps(&[("scan:cross-scan", 0)]));
    let got = result(&w.land(&card, &cand));
    assert_eq!(got["moved"], json!(true), "{got}");

    let w = World::with_table(ENFORCE_PROFILE);
    let card = w.review_card();
    let cand = w.candidate_adding("notes/x.md");
    // Tier 0: no receipt at all.
    let payload = json!({
        "head": cand, "base": w.tip, "verdict": "pass", "tier": 0,
        "evidence": "docs only", "receipt": null, "reviewer": S_AUD,
    });
    let conn = Connection::open(ledger::ledger_path(w.home.path(), "demo").unwrap()).unwrap();
    conn.execute(
        "INSERT INTO event (event_key, at, actor, card_id, kind, payload) \
         VALUES ('test-verdict-docs', unixepoch(), ?1, ?2, 'card.verdict', ?3)",
        rusqlite::params![S_AUD, number(&card), payload.to_string()],
    )
    .unwrap();
    drop(conn);
    let got = result(&w.land(&card, &cand));
    assert_eq!(got["moved"], json!(true), "{got}");
    assert_eq!(got["scans"]["required"], json!([]), "{got}");
}

/// AC-5 (REQ-6, REQ-7): with no `land_verdict` key and with `observe` the landing moves, and the
/// answer says what it would have refused.
#[test]
fn ac5_observe_mode_reports_the_missing_scans_and_lands() {
    for profile in [DEFAULT_PROFILE, OBSERVE_PROFILE] {
        let w = World::with_table(profile);
        let card = w.review_card();
        let cand = w.candidate_adding("tests/branch_gc_cli.rs");
        w.pass(&card, &cand, &BASE_STEPS);
        let got = result(&w.land(&card, &cand));
        assert_eq!(got["moved"], json!(true), "{got}");
        assert_eq!(got["closed"], json!(true), "{got}");
        assert_eq!(w.integ(), cand);
        assert_eq!(w.events(&card, "card.landed"), 1);
        assert_eq!(
            got["scans"]["would_refuse"]["code"],
            json!("scans_missing"),
            "{got}"
        );
        assert_eq!(
            got["scans"]["missing"],
            json!(["test-area-layout", "clean-env-only"]),
            "{got}"
        );
        assert!(got["scans"]["would_refuse"]["why"]
            .as_str()
            .unwrap()
            .contains(AREA_CHECK));
    }
}

/// AC-6 (REQ-8, REQ-13): a table that is not JSON refuses in enforce mode and is a would-refuse in
/// observe mode; a repository with no table needs no scan.
#[test]
fn ac6_an_unreadable_table_is_refused_and_no_table_needs_no_scan() {
    let w = World::new(Some("this is not json"), None, ENFORCE_PROFILE);
    let card = w.review_card();
    let cand = w.candidate_adding("notes/x.md");
    w.pass(&card, &cand, &BASE_STEPS);
    let before = w.dump();
    w.refuse(&w.land(&card, &cand), "scans_not_read", &before);

    w.set_profile(OBSERVE_PROFILE);
    let got = result(&w.land(&card, &cand));
    assert_eq!(got["moved"], json!(true), "{got}");
    assert_eq!(
        got["scans"]["would_refuse"]["code"],
        json!("scans_not_read"),
        "{got}"
    );

    let w = World::new(None, None, ENFORCE_PROFILE);
    let card = w.review_card();
    let cand = w.candidate_adding("tests/branch_gc_cli.rs");
    w.pass(&card, &cand, &BASE_STEPS);
    let got = result(&w.land(&card, &cand));
    assert_eq!(got["moved"], json!(true), "{got}");
    assert_eq!(got["scans"]["required"], json!([]), "{got}");
}

/// AC-7 (REQ-9): a head that the integration branch already holds closes without a scan
/// computation, even with a table that cannot be read.
#[test]
fn ac7_a_head_inside_the_tip_skips_the_scans() {
    let w = World::new(Some("this is not json"), None, DEFAULT_PROFILE);
    let card = w.review_card();
    w.pass(&card, &w.head, &[("fmt", 0)]);
    let got = result(&w.land(&card, &w.head));
    assert_eq!(got["moved"], json!(false), "{got}");
    assert_eq!(got["closed"], json!(true), "{got}");
    assert!(
        got["scans"].is_null() || got["scans"]["would_refuse"].is_null(),
        "{got}"
    );
}

/// AC-11 (REQ-3): a landing by fast-forward, whose first-parent line holds no merge commit, needs
/// the scans of the range from the tip to the head as one change: the missing receipt refuses and
/// a covered one lets the branch move.
#[test]
fn ac11_a_fast_forward_landing_needs_the_scans_of_its_range() {
    let both = [("scan:test-area-layout", 0), ("scan:clean-env-only", 0)];
    let w = World::with_table(ENFORCE_PROFILE);
    let card = w.review_card();
    let cand = w.fast_forward_adding("tests/branch_gc_cli.rs");
    assert_eq!(
        git(
            &w.project,
            &["rev-list", "--merges", &format!("{}..{cand}", w.tip)]
        ),
        "",
        "the fixture head has a merge commit on its line"
    );
    w.pass(&card, &cand, &BASE_STEPS);
    let before = w.dump();
    let message = w.refuse(&w.land(&card, &cand), "scans_missing", &before);
    assert!(message.contains("test-area-layout"), "{message}");
    assert!(message.contains("clean-env-only"), "{message}");

    w.pass(&card, &cand, &with_steps(&both));
    let got = result(&w.land(&card, &cand));
    assert_eq!(got["moved"], json!(true), "{got}");
    assert_eq!(w.integ(), cand);
    assert_eq!(got["scans"]["missing"], json!([]), "{got}");
}

/// The fixture table with two more whole-tree scans that a head touching `tests/` hits: one whose
/// only check is a Python module and one that lists a cargo check and a Python module.
fn table_with_python_checks() -> String {
    let mut value: Value = serde_json::from_str(&table()).unwrap();
    value["scans"].as_array_mut().unwrap().extend([
        json!({"id": "python-only", "always": true,
               "tests": ["tests/py_only_test.py"],
               "checks": [PY_CHECK],
               "rules": ["tests/py_only_rule.rs"],
               "scope": [{"prefix": "tests/", "suffixes": [".rs"]}],
               "rejects": "a fixture that only a Python module reads"}),
        json!({"id": "cargo-and-python", "always": true,
               "tests": ["tests/mixed_rule.rs::mixed"],
               "checks": [MIXED_CHECK, PY_CHECK],
               "rules": ["tests/mixed_rule.rs"],
               "scope": [{"prefix": "tests/", "suffixes": [".rs"]}],
               "rejects": "a fixture that a cargo test and a Python module read"}),
    ]);
    value.to_string()
}

const PY_CHECK: &str = "python3 -m unittest tests.fixture_test";
const MIXED_CHECK: &str = "cargo test --offline --locked --test mixed_rule";

/// AC-12 (REQ-15): the needed list and the refusal line name cargo checks only. A scan whose checks
/// are all Python is not required (the gate does not ask for a Python run) and a scan with both
/// keeps its cargo check alone; the table itself keeps its Python lines.
#[test]
fn ac12_the_required_scans_name_cargo_checks_only() {
    let w = World::new(Some(&table_with_python_checks()), None, ENFORCE_PROFILE);
    let card = w.review_card();
    let cand = w.candidate_adding("tests/branch_gc_cli.rs");
    w.pass(&card, &cand, &BASE_STEPS);
    let before = w.dump();
    let message = w.refuse(&w.land(&card, &cand), "scans_missing", &before);
    assert!(message.contains(MIXED_CHECK), "{message}");
    assert!(message.contains("scan:cargo-and-python"), "{message}");
    assert!(message.contains(AREA_CHECK), "{message}");
    assert!(!message.contains("python3"), "{message}");
    assert!(!message.contains("scan:python-only"), "{message}");

    // Receipts for the cargo scans alone let the landing move: no step for the Python scan.
    let cargo_steps = [
        ("scan:test-area-layout", 0),
        ("scan:clean-env-only", 0),
        ("scan:cargo-and-python", 0),
    ];
    w.pass(&card, &cand, &with_steps(&cargo_steps));
    let got = result(&w.land(&card, &cand));
    assert_eq!(got["moved"], json!(true), "{got}");
    let required = got["scans"]["required"].as_array().unwrap();
    let ids: Vec<&str> = required.iter().map(|s| s["id"].as_str().unwrap()).collect();
    assert_eq!(
        ids,
        ["test-area-layout", "clean-env-only", "cargo-and-python"],
        "{got}"
    );
    for scan in required {
        let checks = scan["checks"].as_array().unwrap();
        assert!(!checks.is_empty(), "{scan}");
        assert!(
            checks
                .iter()
                .all(|c| c.as_str().unwrap().starts_with("cargo ")),
            "{scan}"
        );
    }
}
