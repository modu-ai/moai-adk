//! SPEC-20261126-claude-md-fresh-merge `moai claude-md-plan`. The `base_` cases (AC-2: REQ-3,
//! REQ-4, D-2, D-4): the command lists a project's old instruction files in the fixed order with
//! their facts, says which lines of each are the project's own against the Go manifest and git
//! history, refuses a source it must not read, and writes nothing. The `hunk_` cases (AC-3: REQ-5,
//! with the preview half of REQ-6, REQ-7 and REQ-9, D-6, D-8): the project lines as hunks, the
//! `--keep` picks, the texts they compose, the warnings, the plan digest and the apply command.
//! The `apply_` cases (AC-4: REQ-6, REQ-7, REQ-8, D-1, D-3, D-5): the printed command moves every
//! source into the backup folder, writes the new files and saves a receipt, and a step that fails
//! leaves that receipt pending. The `refuse_` cases (AC-5: REQ-9, D-8): each refusal says why and
//! leaves every byte, `.git` included, as it was.
//! Every case runs the real binary on a temporary project with a temporary HOME; a
//! credential-shaped value is built at run time and appears in no output.

#![cfg(unix)]

mod common;
use common::{CleanEnv, MOAI_BIN};
use moai_cli::agents::{is_generated_local, AGENTS_SECTION};
use moai_cli::project_update::claude_md::{HISTORY_LIMIT, RESULT_LINE_LIMIT, SOURCE_LIMIT};
use moai_cli::project_update::PERMISSION_PLACEHOLDER;
use moai_cli::state_root::{marker_bytes, GO_BACKUP, MOAI_DIR, RUST_MARKER};
use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::{
    collections::BTreeMap,
    fs,
    os::unix::fs::PermissionsExt,
    path::{Path, PathBuf},
    process::{Command, Stdio},
};
use tempfile::TempDir;

fn sha(body: &[u8]) -> String {
    format!("sha256:{:x}", Sha256::digest(body))
}

/// A path under the project's local rules folder.
fn rules(tail: &str) -> String {
    format!("{MOAI_DIR}/docs/local-rules/{tail}")
}

fn rules_dir() -> String {
    format!("{MOAI_DIR}/docs/local-rules")
}

struct Fx {
    _tmp: TempDir,
    base: PathBuf,
    home: PathBuf,
    root: PathBuf,
    recorded: serde_json::Map<String, Value>,
    /// Values no output of any run may hold.
    secrets: Vec<String>,
}

struct Out {
    code: Option<i32>,
    stdout: String,
    stderr: String,
}

impl Fx {
    fn new() -> Fx {
        let tmp = TempDir::new().unwrap();
        let base = fs::canonicalize(tmp.path()).unwrap();
        let (home, root) = (base.join("home"), base.join("project"));
        fs::create_dir_all(&home).unwrap();
        fs::create_dir_all(&root).unwrap();
        Fx {
            _tmp: tmp,
            base,
            home,
            root,
            recorded: serde_json::Map::new(),
            secrets: Vec::new(),
        }
    }

    /// A value built at run time that no output may hold: `fill` repeated.
    fn secret(&mut self, fill: &str) -> String {
        let value = fill.repeat(24);
        self.secrets.push(value.clone());
        value
    }

    /// A usable project profile with the writing languages `language`, in a state folder marked
    /// as this generation's, as a project that has moved its Go state has it: without the marker
    /// a Go backup beside the folder reads as a migration stopped midway.
    fn profile(&self, language: Value) {
        self.write(
            &format!("{MOAI_DIR}/{RUST_MARKER}"),
            marker_bytes("fixture", &[]),
        );
        let mut profile = json!({
            "schema_version": 2, "project_id": "zq-demo", "purpose": "zq purpose",
            "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude"],
        });
        if !language.is_null() {
            profile["language"] = language;
        }
        self.write(&profile_rel(), profile.to_string());
    }

    /// A root `CLAUDE.md` of nothing but Go curator markers: a Go-generation file, without
    /// which no apply command is printed, that has no project line and so no hunk.
    fn go_marker(&self) {
        self.write(
            "CLAUDE.md",
            "<!-- moai:learned-start -->\n<!-- moai:learned-end -->\n",
        );
    }

    fn write(&self, rel: &str, body: impl AsRef<[u8]>) {
        let p = self.root.join(rel);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        fs::write(p, body).unwrap();
    }

    fn link(&self, rel: &str, target: &Path) {
        let p = self.root.join(rel);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        std::os::unix::fs::symlink(target, p).unwrap();
    }

    /// Records `body` as what the Go tool deployed at `rel` (the file itself is not written).
    fn record(&mut self, rel: &str, body: &str) {
        let hash = sha(body.as_bytes());
        self.recorded.insert(
            rel.into(),
            json!({"provenance": "template_managed", "template_hash": hash,
                   "deployed_hash": hash, "current_hash": hash}),
        );
    }

    /// The Go manifest as an update leaves it, under the Go backup folder.
    fn manifest(&self) {
        self.write(
            &format!("{GO_BACKUP}/manifest.json"),
            json!({"version": "v3.2.1-rc.6", "deployed_at": "2026-11-01T00:00:00Z",
                   "files": self.recorded})
            .to_string(),
        );
    }

    fn git(&self, args: &[&str]) -> String {
        common::git::git(&self.root, args)
    }

    fn git_init(&self) {
        self.git(&["init", "-q"]);
    }

    fn commit(&self, paths: &[&str]) {
        let mut add = vec!["add", "-f", "--"];
        add.extend_from_slice(paths);
        self.git(&add);
        self.git(&[
            "-c",
            "user.name=fixture",
            "-c",
            "user.email=fixture@example.invalid",
            "-c",
            "commit.gpgsign=false",
            "commit",
            "-q",
            "-m",
            "fixture",
        ]);
    }

    /// Builds history from a `git fast-import` stream.
    fn fast_import(&self, stream: &str) {
        let mut child = common::git::command()
            .args(["fast-import", "--quiet"])
            .current_dir(&self.root)
            .stdin(Stdio::piped())
            .stdout(Stdio::null())
            .stderr(Stdio::inherit())
            .spawn()
            .expect("spawn git fast-import");
        common::feed_stdin(&mut child, stream.as_bytes());
        assert!(child.wait().unwrap().success(), "git fast-import failed");
    }

    /// Everything under the project (`.git` included) and HOME: the kind, the mode and the
    /// bytes of each entry; links are never followed and a pipe is never opened.
    fn snapshot(&self) -> BTreeMap<PathBuf, (String, Vec<u8>)> {
        let mut out = BTreeMap::new();
        let mut stack = vec![self.root.clone(), self.home.clone()];
        while let Some(d) = stack.pop() {
            let Ok(entries) = fs::read_dir(&d) else {
                continue;
            };
            for e in entries {
                let p = e.unwrap().path();
                let m = fs::symlink_metadata(&p).unwrap();
                let mode = m.permissions().mode();
                let (kind, bytes) = if m.file_type().is_symlink() {
                    let target = fs::read_link(&p).unwrap();
                    ("link", target.display().to_string().into_bytes())
                } else if m.is_dir() {
                    stack.push(p.clone());
                    ("dir", Vec::new())
                } else if m.is_file() {
                    match fs::read(&p) {
                        Ok(bytes) => ("file", bytes),
                        Err(_) => ("unreadable", Vec::new()),
                    }
                } else {
                    ("other", Vec::new())
                };
                out.insert(p, (format!("{kind} {mode:o}"), bytes));
            }
        }
        out
    }

    fn run(&self, args: &[&str]) -> Out {
        self.run_env(args, &[])
    }

    fn run_env(&self, args: &[&str], extra: &[(&str, &str)]) -> Out {
        let out = Command::new(MOAI_BIN)
            .args(args)
            .clean_env()
            .env("HOME", &self.home)
            .env("PATH", "/usr/bin:/bin")
            .env("LC_ALL", "C")
            .envs(extra.iter().copied())
            .stdin(Stdio::null())
            .output()
            .expect("spawn the binary");
        let out = Out {
            code: out.status.code(),
            stdout: String::from_utf8_lossy(&out.stdout).into_owned(),
            stderr: String::from_utf8_lossy(&out.stderr).into_owned(),
        };
        for secret in &self.secrets {
            assert!(
                !out.stdout.contains(secret) && !out.stderr.contains(secret),
                "a credential-shaped value was printed by {args:?}: {}{}",
                out.stdout,
                out.stderr
            );
        }
        out
    }

    /// `claude-md-plan` with these `--keep` values, leaving every byte as it was.
    fn run_plan(&self, keep: &[&str]) -> Out {
        let mut args = vec!["claude-md-plan", self.root.to_str().unwrap()];
        for value in keep {
            args.extend(["--keep", value]);
        }
        let before = self.snapshot();
        let out = self.run(&args);
        assert_eq!(self.snapshot(), before, "claude-md-plan wrote something");
        out
    }

    /// `claude-md-plan`: it succeeds and leaves every byte as it was.
    fn plan(&self) -> Value {
        self.plan_with(&[])
    }

    fn plan_with(&self, keep: &[&str]) -> Value {
        let out = self.run_plan(keep);
        assert_eq!(
            out.code,
            Some(0),
            "claude-md-plan failed: stdout {} stderr {}",
            out.stdout,
            out.stderr
        );
        let v: Value = serde_json::from_str(&out.stdout).expect("a JSON answer");
        assert_eq!(v["ok"], json!(true), "{v}");
        v["result"].clone()
    }
}

fn paths(result: &Value) -> Vec<String> {
    result["sources"]
        .as_array()
        .expect("sources")
        .iter()
        .map(|s| s["path"].as_str().unwrap().to_string())
        .collect()
}

fn source<'v>(result: &'v Value, path: &str) -> &'v Value {
    result["sources"]
        .as_array()
        .expect("sources")
        .iter()
        .find(|s| s["path"] == json!(path))
        .unwrap_or_else(|| panic!("no source {path}: {result}"))
}

/// The source's comparison and project lines, checked against its count.
fn assert_lines(result: &Value, path: &str, comparison: Value, lines: &[usize]) {
    let s = source(result, path);
    assert_eq!(s["comparison"], comparison, "{path}: {s}");
    assert_eq!(s["project_lines"], json!(lines), "{path}: {s}");
    assert_eq!(s["project_line_count"], json!(lines.len()), "{path}: {s}");
}

fn all_lines(reason: &str) -> Value {
    json!({"used": "all_lines", "reason": reason})
}

fn history(blob: &str) -> Value {
    json!({"used": "git_history", "blob": blob})
}

fn refused(result: &Value) -> Vec<(String, String)> {
    result["refused"]
        .as_array()
        .expect("refused")
        .iter()
        .map(|r| {
            (
                r["path"].as_str().unwrap().to_string(),
                r["reason"].as_str().unwrap().to_string(),
            )
        })
        .collect()
}

fn profile_rel() -> String {
    format!("{MOAI_DIR}/project.yaml")
}

fn list<'v>(result: &'v Value, key: &str) -> &'v Vec<Value> {
    result[key]
        .as_array()
        .unwrap_or_else(|| panic!("no list {key}: {result}"))
}

fn words(values: &[Value]) -> Vec<String> {
    values
        .iter()
        .map(|v| v.as_str().unwrap().to_string())
        .collect()
}

fn ids(result: &Value) -> Vec<String> {
    list(result, "hunks")
        .iter()
        .map(|h| h["id"].as_str().unwrap().to_string())
        .collect()
}

fn hunk<'v>(result: &'v Value, id: &str) -> &'v Value {
    list(result, "hunks")
        .iter()
        .find(|h| h["id"] == json!(id))
        .unwrap_or_else(|| panic!("no hunk {id}: {result}"))
}

/// The hunk's heading, text and hidden line count.
fn assert_hunk(result: &Value, id: &str, heading: Option<&str>, text: &str, hidden: usize) {
    let h = hunk(result, id);
    assert_eq!(h["heading"], json!(heading), "{id}: {h}");
    assert_eq!(h["text"], json!(text), "{id}: {h}");
    assert_eq!(h["same_as"], json!(null), "{id}: {h}");
    assert_eq!(h["hidden_lines"], json!(hidden), "{id}: {h}");
    let shown = if text.is_empty() {
        0
    } else {
        text.split('\n').count()
    };
    assert_eq!(h["line_count"], json!(shown + hidden), "{id}: {h}");
}

fn composed<'v>(result: &'v Value, path: &str) -> Option<&'v Value> {
    list(result, "results")
        .iter()
        .find(|r| r["path"] == json!(path))
}

/// The text of a resulting file, checked against its line count and (with no hidden line) its
/// digest.
fn composed_text(result: &Value, path: &str) -> String {
    let file = composed(result, path).unwrap_or_else(|| panic!("no result {path}: {result}"));
    let text = file["text"].as_str().expect("text").to_string();
    if file["hidden_lines"] == json!(0) {
        assert_eq!(file["sha256"], json!(sha(text.as_bytes())), "{file}");
        assert_eq!(file["line_count"], json!(text.lines().count()), "{file}");
    }
    text
}

/// The `<path>:<line>` and reason of every line a resulting file leaves out.
fn left_out(result: &Value, path: &str) -> Vec<(String, String)> {
    let file = composed(result, path).unwrap_or_else(|| panic!("no result {path}: {result}"));
    list(file, "left_out")
        .iter()
        .map(|l| {
            (
                l["line"].as_str().unwrap().to_string(),
                l["reason"].as_str().unwrap().to_string(),
            )
        })
        .collect()
}

/// The answer carries no digest and no command, and one of its reasons holds every needle.
fn assert_no_apply(result: &Value, needles: &[&str]) {
    assert_eq!(result["apply"], json!(null), "{result}");
    assert_eq!(result["plan_digest"], json!(null), "{result}");
    let reasons = words(list(result, "apply_blocked"));
    assert!(
        reasons
            .iter()
            .any(|r| needles.iter().all(|needle| r.contains(needle))),
        "no reason holds {needles:?}: {reasons:?}"
    );
}

/// The digest of an answer that prints a command, with that command's words.
fn assert_apply(result: &Value) -> (String, Vec<String>) {
    assert_eq!(result["apply_blocked"], json!([]), "{result}");
    let digest = result["plan_digest"].as_str().expect("a digest");
    assert!(
        digest.starts_with("sha256:") && digest.len() == 71,
        "{digest}"
    );
    assert_eq!(result["apply"]["stdin"], json!(null), "{result}");
    let argv = words(result["apply"]["argv"].as_array().expect("argv"));
    (digest.to_string(), argv)
}

fn block(id: &str) -> (String, String) {
    (
        format!("<!-- moai:managed:begin id={id} -->"),
        format!("<!-- moai:managed:end id={id} -->"),
    )
}

const GO_CLAUDE: &str = "# zq-go template\n@AGENTS.md\n@AGENTS.local.md\n";
const GO_AGENTS: &str = "# zq-go agents\n\nzq-template one\nzq-template two\n";
const EDITED_AGENTS: &str = "# zq-go agents\n\nzq-template one\nzq-project rule\nzq-template two\n";

/// A project shaped like the one the SPEC read: a Go manifest in the backup folder, an
/// unchanged Go `CLAUDE.md`, an `AGENTS.md` whose Go original is an earlier commit, two ignored
/// personal files and tracked local rules.
fn go_project() -> Fx {
    let mut fx = Fx::new();
    fx.record("CLAUDE.md", GO_CLAUDE);
    fx.record("AGENTS.md", GO_AGENTS);
    fx.manifest();
    fx.git_init();
    fx.write(".gitignore", "AGENTS.local.md\nCLAUDE.local.md\n");
    fx.write("CLAUDE.md", GO_CLAUDE);
    fx.write("AGENTS.md", GO_AGENTS);
    fx.write(&rules("kanban.md"), "# zq-kanban\nzq-old rule\n");
    fx.write(&rules("gitflow.md"), "# zq-gitflow\n");
    fx.write(&rules("sub/deep.md"), "zq-deep without a newline");
    fx.commit(&[
        ".gitignore",
        "CLAUDE.md",
        "AGENTS.md",
        &rules("kanban.md"),
        &rules("gitflow.md"),
        &rules("sub/deep.md"),
    ]);
    fx.write("AGENTS.md", EDITED_AGENTS);
    fx.commit(&["AGENTS.md"]);
    fx.write("AGENTS.local.md", "# zq-local\nzq-token=zq-secret-value\n");
    fx.write("CLAUDE.local.md", "# zq-local\nzq-other\n");
    fx.write(&rules("zz-new.md"), "zq-not yet committed\n");
    fx
}

#[test]
fn base_the_sources_are_listed_in_order_with_their_facts() {
    let fx = go_project();
    let result = fx.plan();
    let want = [
        ("CLAUDE.md".to_string(), 3, "tracked"),
        ("AGENTS.md".to_string(), 5, "tracked"),
        ("AGENTS.local.md".to_string(), 2, "untracked"),
        ("CLAUDE.local.md".to_string(), 2, "untracked"),
        (rules("gitflow.md"), 1, "tracked"),
        (rules("kanban.md"), 2, "tracked"),
        (rules("sub/deep.md"), 1, "tracked"),
        (rules("zz-new.md"), 1, "untracked"),
    ];
    let order: Vec<String> = want.iter().map(|w| w.0.clone()).collect();
    assert_eq!(paths(&result), order, "{result}");
    for (path, lines, git) in &want {
        let s = source(&result, path);
        assert_eq!(
            s["sha256"],
            json!(sha(&fs::read(fx.root.join(path)).unwrap())),
            "{path}"
        );
        assert_eq!(s["line_count"], json!(lines), "{path}");
        assert_eq!(s["git"], json!(git), "{path}");
        assert!(s["comparison"]["used"].is_string(), "{path}: {s}");
    }
    assert_eq!(refused(&result), vec![], "{result}");
    assert_eq!(result["read_only"], json!(true));
    assert_eq!(result["go_claude_md"], json!("manifest"));
    assert_eq!(
        result["go_manifest"],
        json!(format!("{GO_BACKUP}/manifest.json"))
    );
    assert_lines(&result, "AGENTS.local.md", all_lines("no_record"), &[1, 2]);
    assert_lines(
        &result,
        &rules("kanban.md"),
        all_lines("no_record"),
        &[1, 2],
    );
}

#[test]
fn base_a_claude_md_equal_to_the_record_has_no_project_line() {
    let fx = go_project();
    let result = fx.plan();
    assert_lines(&result, "CLAUDE.md", json!({"used": "manifest_equal"}), &[]);
}

#[test]
fn base_an_edited_agents_md_lists_only_the_lines_the_go_original_lacks() {
    let fx = go_project();
    let original = fx.git(&["rev-parse", "HEAD~1:AGENTS.md"]);
    assert_lines(&fx.plan(), "AGENTS.md", history(&original), &[4]);
    // An edit that is not committed yet counts against the same original.
    fx.write(
        "AGENTS.md",
        format!("zq-first\n{EDITED_AGENTS}zq-last\nzq-template one\n"),
    );
    assert_lines(&fx.plan(), "AGENTS.md", history(&original), &[1, 5, 7]);
}

#[test]
fn base_without_a_matching_version_every_line_counts() {
    let mut fx = go_project();
    fx.record("AGENTS.md", "zq-bytes that were never committed\n");
    fx.manifest();
    assert_lines(
        &fx.plan(),
        "AGENTS.md",
        all_lines("no_matching_version"),
        &[1, 2, 3, 4, 5],
    );
    // No manifest at all: nothing is recorded for any path.
    fs::remove_file(fx.root.join(GO_BACKUP).join("manifest.json")).unwrap();
    let result = fx.plan();
    assert_eq!(result["go_manifest"], json!(null));
    assert_lines(&result, "CLAUDE.md", all_lines("no_record"), &[1, 2, 3]);
}

#[test]
fn base_outside_a_git_repository_every_line_counts() {
    let mut fx = Fx::new();
    fx.record("CLAUDE.md", GO_CLAUDE);
    fx.record("AGENTS.md", GO_AGENTS);
    fx.manifest();
    fx.write("CLAUDE.md", GO_CLAUDE);
    fx.write("AGENTS.md", EDITED_AGENTS);
    let result = fx.plan();
    assert_lines(&result, "CLAUDE.md", json!({"used": "manifest_equal"}), &[]);
    assert_lines(
        &result,
        "AGENTS.md",
        all_lines("not_a_repository"),
        &[1, 2, 3, 4, 5],
    );
    assert_eq!(source(&result, "AGENTS.md")["git"], json!("unknown"));
}

#[test]
fn base_a_failed_git_read_counts_every_line() {
    let fx = go_project();
    fx.write(".git/config", "[zq-broken\n");
    let result = fx.plan();
    assert_lines(
        &result,
        "AGENTS.md",
        all_lines("git_failed"),
        &[1, 2, 3, 4, 5],
    );
    assert_eq!(source(&result, "AGENTS.md")["git"], json!("unknown"));
}

/// One `fast-import` commit on `branch` writing `body` to `AGENTS.md`; `parents` holds the
/// `from` and `merge` lines.
fn imported(branch: &str, mark: usize, parents: &str, body: &str) -> String {
    format!(
        "commit refs/heads/{branch}\nmark :{mark}\n\
         committer fixture <fixture@example.invalid> {} +0000\ndata 2\nc\n{parents}\
         M 100644 inline AGENTS.md\ndata {}\n{body}\n",
        1_700_000_000 + mark,
        body.len()
    )
}

#[test]
fn base_the_history_search_stops_after_the_latest_two_hundred_versions() {
    for (versions, found) in [(HISTORY_LIMIT, true), (HISTORY_LIMIT + 1, false)] {
        let mut fx = Fx::new();
        let original = "zq-original\n";
        fx.record("AGENTS.md", original);
        fx.manifest();
        fx.git_init();
        let mut stream = imported("keep", 1, "", original);
        for i in 2..=versions {
            stream.push_str(&imported(
                "keep",
                i,
                "",
                &format!("zq-original\nzq-version {i}\n"),
            ));
        }
        fx.fast_import(&stream);
        fx.write("AGENTS.md", "zq-original\nzq-mine\n");
        let result = fx.plan();
        if found {
            let blob = fx.git(&["rev-parse", &format!("keep~{}:AGENTS.md", versions - 1)]);
            assert_lines(&result, "AGENTS.md", history(&blob), &[2]);
        } else {
            assert_lines(
                &result,
                "AGENTS.md",
                all_lines("no_matching_version"),
                &[1, 2],
            );
        }
    }
}

#[test]
fn base_the_history_search_reads_other_branches_and_both_sides_of_a_merge() {
    let original = "zq-original\n";
    // The original only on a branch that is not checked out.
    let other = [
        imported("keep", 1, "", "zq-base\n"),
        imported("other", 2, "from :1\n", original),
    ]
    .concat();
    // The original on a side branch whose change the merge did not keep.
    let discarded = [
        imported("keep", 1, "", "zq-base\n"),
        imported("side", 2, "from :1\n", original),
        imported("keep", 3, "", "zq-main\n"),
        imported("keep", 4, "merge :2\n", "zq-main\n"),
    ]
    .concat();
    // The original first written by the merge itself.
    let resolved = [
        imported("keep", 1, "", "zq-base\n"),
        imported("side", 2, "from :1\n", "zq-side\n"),
        imported("keep", 3, "", "zq-main\n"),
        imported("keep", 4, "merge :2\n", original),
    ]
    .concat();
    for (name, stream) in [
        ("another branch", other),
        ("a discarded side", discarded),
        ("a merge result", resolved),
    ] {
        let mut fx = Fx::new();
        fx.record("AGENTS.md", original);
        fx.manifest();
        fx.git_init();
        fx.fast_import(&stream);
        if name == "a discarded side" {
            fx.git(&["update-ref", "-d", "refs/heads/side"]);
        }
        fx.write("AGENTS.md", "zq-original\nzq-mine\n");
        let result = fx.plan();
        assert_eq!(
            source(&result, "AGENTS.md")["comparison"]["used"],
            json!("git_history"),
            "{name}: {result}"
        );
        assert_eq!(
            source(&result, "AGENTS.md")["project_lines"],
            json!([2]),
            "{name}"
        );
    }
}

#[test]
fn base_the_managed_block_and_curator_marker_lines_are_never_project_lines() {
    let (begin, end) = block(AGENTS_SECTION);
    let (other_begin, other_end) = block("other");
    let body = [
        "zq-a",                                       // 1
        "<!-- moai:learned-start id=\"zq\" -->",      // 2
        "zq-learned",                                 // 3
        "<!-- moai:learned-end -->",                  // 4
        "<!-- moai:harness-start -->",                // 5
        "<!-- moai:harness-end -->",                  // 6
        "  <!-- moai:learned-local-start -->",        // 7
        "<!-- moai:learned-local-end -->",            // 8
        "zq-see <!-- moai:learned-start --> in text", // 9
        "<!-- moai:learned-start --> zq-trailing",    // 10
        "<!-- moai:learned-start --> zq <!-- zq -->", // 11
        &begin,                                       // 12: no end before the next begin
        "zq-between",                                 // 13
        &begin,                                       // 14
        "zq-block body",                              // 15
        &end,                                         // 16
        "zq-after",                                   // 17
        &other_begin,                                 // 18
        &other_end,                                   // 19
        &begin,                                       // 20: never ended
        "zq-tail",                                    // 21
    ]
    .join("\n");
    let mut fx = Fx::new();
    // `AGENTS.md` counts every line; `CLAUDE.md` is compared with a committed original.
    fx.write("AGENTS.md", &body);
    fx.record("CLAUDE.md", "zq-a\n");
    fx.manifest();
    fx.git_init();
    fx.write("CLAUDE.md", "zq-a\n");
    fx.commit(&["CLAUDE.md"]);
    fx.write("CLAUDE.md", &body);
    let result = fx.plan();
    let counted = [1, 3, 9, 10, 11, 12, 13, 17, 18, 19, 20, 21];
    assert_lines(&result, "AGENTS.md", all_lines("no_record"), &counted);
    let blob = fx.git(&["rev-parse", "HEAD:CLAUDE.md"]);
    assert_lines(&result, "CLAUDE.md", history(&blob), &counted[1..]);
}

#[test]
fn base_a_personal_file_that_is_a_moai_first_text_has_no_project_line() {
    let en = "# Personal supplement\n\nResponse language: English\n\n<!-- Personal file Claude Code loads next to CLAUDE.md. Keep secrets out; never share it in the repo. -->\n";
    let ko = "# 개인 작업 보충\n\n응답 언어: 한국어\n\n<!-- Claude Code가 CLAUDE.md 옆에서 함께 읽는 개인 파일. 비밀값을 넣지 않으며 repo에 공유하지 않는다. -->\n";
    assert!(is_generated_local(en.as_bytes()) && is_generated_local(ko.as_bytes()));
    let fx = Fx::new();
    fx.write("CLAUDE.local.md", en);
    fx.write("AGENTS.local.md", ko);
    // The same bytes anywhere else are not a personal file.
    fx.write("AGENTS.md", en);
    fx.write(&rules("copy.md"), en);
    let result = fx.plan();
    for path in ["CLAUDE.local.md", "AGENTS.local.md"] {
        assert_lines(&result, path, json!({"used": "moai_first_text"}), &[]);
        // Such a file has no hunk, so its text is shown nowhere under its name.
        assert!(
            ids(&result).iter().all(|id| !id.starts_with(path)),
            "{path}: {result}"
        );
    }
    for path in ["AGENTS.md".to_string(), rules("copy.md")] {
        assert_lines(&result, &path, all_lines("no_record"), &[1, 2, 3, 4, 5]);
    }
    // One added line makes it the user's file: every line counts.
    fx.write("CLAUDE.local.md", format!("{en}zq-mine\n"));
    assert_lines(
        &fx.plan(),
        "CLAUDE.local.md",
        all_lines("no_record"),
        &[1, 2, 3, 4, 5, 6],
    );
}

#[test]
fn base_the_go_claude_md_verdict_follows_markers_the_block_and_the_manifest() {
    let (begin, end) = block(AGENTS_SECTION);
    let managed = format!("{GO_CLAUDE}{begin}\nzq-facts\n{end}\n");
    let marked = |marker: &str| format!("{managed}<!-- {marker} -->\n");
    let cases: Vec<(&str, Option<String>, bool, Value)> = vec![
        ("recorded", Some(GO_CLAUDE.into()), true, json!("manifest")),
        (
            "recorded and edited",
            Some("zq-mine\n".into()),
            true,
            json!("manifest"),
        ),
        (
            "recorded with a block",
            Some(managed.clone()),
            true,
            json!(null),
        ),
        (
            "learned marker with a block",
            Some(marked("moai:learned-start")),
            true,
            json!("marker"),
        ),
        (
            "harness marker without a manifest",
            Some(marked("moai:harness-start")),
            false,
            json!("marker"),
        ),
        (
            "local marker without a manifest",
            Some("<!-- moai:learned-local-start -->\n".into()),
            false,
            json!("marker"),
        ),
        ("no proof", Some(GO_CLAUDE.into()), false, json!(null)),
        ("no file", None, true, json!(null)),
    ];
    for (name, body, recorded, want) in cases {
        let mut fx = Fx::new();
        fx.record(if recorded { "CLAUDE.md" } else { "zq.md" }, GO_CLAUDE);
        fx.manifest();
        if let Some(body) = &body {
            fx.write("CLAUDE.md", body);
        }
        assert_eq!(fx.plan()["go_claude_md"], want, "{name}");
    }
    // A link is never the Go file, whatever it points at.
    let mut fx = Fx::new();
    fx.record("CLAUDE.md", GO_CLAUDE);
    fx.manifest();
    let outside = fx.base.join("outside.md");
    fs::write(&outside, marked("moai:learned-start")).unwrap();
    fx.link("CLAUDE.md", &outside);
    assert_eq!(fx.plan()["go_claude_md"], json!(null), "a link");
}

#[test]
fn base_an_unsafe_source_is_refused_and_named() {
    let fx = Fx::new();
    let outside = fx.base.join("outside");
    fs::create_dir_all(&outside).unwrap();
    fs::write(outside.join("inner.md"), "zq-outside\n").unwrap();
    fx.write("CLAUDE.md", "zq-fine\n");
    fx.link("AGENTS.md", &outside.join("inner.md"));
    fx.write("AGENTS.local.md", [b'z', 0xff, 0xfe, b'\n']);
    fs::create_dir_all(fx.root.join("CLAUDE.local.md")).unwrap();
    fx.write(&rules("empty.md"), "");
    let limit = usize::try_from(SOURCE_LIMIT).unwrap();
    fx.write(&rules("limit.md"), vec![b'a'; limit]);
    fx.write(&rules("over.md"), vec![b'a'; limit + 1]);
    fx.link(&rules("dirlink"), &outside);
    fx.link(&rules("filelink.md"), &outside.join("inner.md"));
    let pipe = fx.root.join(rules("pipe.md"));
    let made = Command::new("/usr/bin/mkfifo").arg(&pipe).status().unwrap();
    assert!(made.success(), "mkfifo");
    fx.write(&rules("locked.md"), "zq-locked\n");
    fx.write(&rules("lockeddir/in.md"), "zq-in\n");
    let locked = fx.root.join(rules("locked.md"));
    let lockeddir = fx.root.join(rules("lockeddir"));
    fs::set_permissions(&locked, fs::Permissions::from_mode(0o000)).unwrap();
    fs::set_permissions(&lockeddir, fs::Permissions::from_mode(0o000)).unwrap();

    let result = fx.plan();
    fs::set_permissions(&lockeddir, fs::Permissions::from_mode(0o755)).unwrap();

    let want: Vec<(String, String)> = [
        ("AGENTS.md".to_string(), "is a link"),
        ("AGENTS.local.md".to_string(), "is not UTF-8 text"),
        ("CLAUDE.local.md".to_string(), "is not a regular file"),
        (rules("dirlink"), "is a link"),
        (rules("filelink.md"), "is a link"),
        (rules("locked.md"), "cannot be read"),
        (rules("lockeddir"), "cannot be listed"),
        (rules("over.md"), "is larger than 1 MiB"),
        (rules("pipe.md"), "is not a regular file"),
    ]
    .into_iter()
    .map(|(p, r)| (p, r.to_string()))
    .collect();
    assert_eq!(refused(&result), want, "{result}");
    assert_eq!(
        paths(&result),
        vec![
            "CLAUDE.md".to_string(),
            rules("empty.md"),
            rules("limit.md")
        ],
        "{result}"
    );
    assert_eq!(source(&result, &rules("empty.md"))["line_count"], json!(0));
    assert_eq!(source(&result, &rules("limit.md"))["line_count"], json!(1));
    assert!(
        !result.to_string().contains(&sha(b"zq-outside\n"))
            && !result.to_string().contains("zq-outside"),
        "the link's target was read: {result}"
    );
}

#[test]
fn base_a_linked_folder_refuses_what_lies_below_it() {
    // The local rules folder itself is a link, to a folder or to nothing.
    for dangling in [false, true] {
        let fx = Fx::new();
        let outside = fx.base.join("outside");
        if !dangling {
            fs::create_dir_all(&outside).unwrap();
            fs::write(outside.join("x.md"), "zq-outside\n").unwrap();
        }
        fx.link(&rules_dir(), &outside);
        let result = fx.plan();
        assert!(!result.to_string().contains("zq-outside"), "{result}");
        assert_eq!(
            refused(&result),
            vec![(rules_dir(), "is a link".to_string())],
            "dangling {dangling}"
        );
        assert_eq!(paths(&result), Vec::<String>::new());
    }
    // A folder above it is a link.
    for (linked, below) in [
        (format!("{MOAI_DIR}/docs"), "local-rules/y.md"),
        (MOAI_DIR.to_string(), "docs/local-rules/y.md"),
    ] {
        let fx = Fx::new();
        let outside = fx.base.join("outside");
        fs::create_dir_all(outside.join(below).parent().unwrap()).unwrap();
        fs::write(outside.join(below), "zq-outside\n").unwrap();
        fx.link(&linked, &outside);
        let result = fx.plan();
        assert!(!result.to_string().contains("zq-outside"), "{result}");
        assert_eq!(
            refused(&result),
            vec![(rules_dir(), "lies below a link folder".to_string())],
            "{linked}"
        );
        assert_eq!(paths(&result), Vec::<String>::new(), "{linked}");
    }
}

#[test]
fn hunk_one_hunk_per_heading_section_holds_its_lines_verbatim() {
    let fx = Fx::new();
    let body = [
        "",                 // 1
        "",                 // 2: a preamble of blank lines is no hunk
        "# zq-one",         // 3
        "zq-a  ",           // 4
        "####### zq-seven", // 5
        "#zq-nospace",      // 6
        "  ```text",        // 7
        "# zq-in-fence",    // 8
        "```",              // 9
        "~~~~",             // 10
        "```",              // 11: another character does not close
        "## zq-in-tilde",   // 12
        "~~~",              // 13: a shorter run does not close
        "~~~~~",            // 14
        "   ### zq-three",  // 15
        "\tzq-b",           // 16
        "    # zq-code",    // 17: four spaces make no heading
        "```zq```",         // 18: a backtick in the info text opens no fence
        "###### zq-six",    // 19
        "#",                // 20
        "zq-c",             // 21
    ];
    fx.write("AGENTS.md", body.join("\n"));
    fx.write("CLAUDE.md", "zq-before any heading\n# zq-h\n");
    fx.write("AGENTS.local.md", "");
    let result = fx.plan();
    assert_eq!(
        ids(&result),
        [
            "CLAUDE.md:1",
            "CLAUDE.md:2",
            "AGENTS.md:3",
            "AGENTS.md:15",
            "AGENTS.md:19",
            "AGENTS.md:20"
        ],
        "{result}"
    );
    assert_hunk(&result, "CLAUDE.md:1", None, "zq-before any heading", 0);
    assert_hunk(&result, "CLAUDE.md:2", Some("# zq-h"), "# zq-h", 0);
    for (id, heading, from, to) in [
        ("AGENTS.md:3", "# zq-one", 3, 14),
        ("AGENTS.md:15", "   ### zq-three", 15, 18),
        ("AGENTS.md:19", "###### zq-six", 19, 19),
        ("AGENTS.md:20", "#", 20, 21),
    ] {
        assert_hunk(
            &result,
            id,
            Some(heading),
            &body[from - 1..to].join("\n"),
            0,
        );
        assert_eq!(hunk(&result, id)["path"], json!("AGENTS.md"));
    }
}

#[test]
fn hunk_holds_only_the_project_lines_of_its_section() {
    let mut fx = Fx::new();
    let base = "# zq-base\nzq-one\n\n## zq-two\nzq-two body\n## zq-three\nzq-three body\n";
    fx.record("AGENTS.md", base);
    fx.manifest();
    fx.git_init();
    fx.write("AGENTS.md", base);
    fx.commit(&["AGENTS.md"]);
    // The second section gains only a blank line the original lacks.
    fx.write(
        "AGENTS.md",
        "# zq-base\nzq-one\n\nzq-added\n## zq-two\n   \nzq-two body\n## zq-three\nzq-new\n\
         zq-three body\nzq-newer\n",
    );
    let result = fx.plan();
    assert_eq!(ids(&result), ["AGENTS.md:4", "AGENTS.md:9"], "{result}");
    assert_hunk(&result, "AGENTS.md:4", Some("# zq-base"), "zq-added", 0);
    assert_hunk(
        &result,
        "AGENTS.md:9",
        Some("## zq-three"),
        "zq-new\nzq-newer",
        0,
    );
}

#[test]
fn hunk_an_equal_section_shows_its_text_once() {
    let fx = Fx::new();
    fx.profile(json!(null));
    let same = "## zq-same\nzq-twice-only\n";
    fx.write("AGENTS.local.md", same);
    fx.write("CLAUDE.local.md", format!("{same}## zq-own\nzq-own line\n"));
    let out = fx.run_plan(&[]);
    assert_eq!(
        out.stdout.matches("zq-twice-only").count(),
        1,
        "{}",
        out.stdout
    );
    let result = fx.plan();
    assert_eq!(
        ids(&result),
        [
            "AGENTS.local.md:1",
            "CLAUDE.local.md:1",
            "CLAUDE.local.md:3"
        ],
        "{result}"
    );
    assert_hunk(
        &result,
        "AGENTS.local.md:1",
        Some("## zq-same"),
        "## zq-same\nzq-twice-only",
        0,
    );
    let second = hunk(&result, "CLAUDE.local.md:1");
    assert_eq!(second["same_as"], json!("AGENTS.local.md:1"), "{second}");
    assert_eq!(second["text"], json!(null), "{second}");
    assert_eq!(second["heading"], json!(null), "{second}");
    assert_eq!(second["line_count"], json!(2), "{second}");
    assert_eq!(
        hunk(&result, "CLAUDE.local.md:3")["same_as"],
        json!(null),
        "{result}"
    );
    // A hunk that shows no text can still be picked.
    let result = fx.plan_with(&["CLAUDE.local.md:1=shared"]);
    assert!(
        composed_text(&result, "CLAUDE.md").ends_with("\n## zq-same\nzq-twice-only\n"),
        "{result}"
    );
}

/// A private key marker line, built at run time.
fn key_line(edge: &str, kind: &str) -> String {
    format!("-----{edge} {kind}PRIVATE KEY-----")
}

#[test]
fn hunk_credential_shaped_lines_are_counted_and_never_printed() {
    let mut fx = Fx::new();
    fx.profile(json!(null));
    fx.go_marker();
    let token = format!("ghp_{}", fx.secret("Q"));
    let split = format!("gh\u{200B}p_{}", fx.secret("W"));
    let opaque = fx.secret("Z");
    let password = format!("X1{}", fx.secret("X"));
    let key = [fx.secret("V"), fx.secret("K"), fx.secret("J")];
    let heading = format!("ghp_{}", fx.secret("G"));
    let body = [
        "# zq-secrets".to_string(),                                      // 1
        "zq-before".to_string(),                                         // 2
        format!("zq-key {token}"),                                       // 3
        format!("zq-split {split}"),                                     // 4
        format!("zq-url https://{opaque}@example.invalid/x"),            // 5
        format!("zq-pass password = {password}"),                        // 6
        "zq-ask @zq-someone, see https://example.invalid/p".to_string(), // 7
        "zq-user https://example.invalid/@zq-someone".to_string(),       // 8
        "zq-form https://${ZQ_TOKEN}@example.invalid/".to_string(),      // 9
        key_line("BEGIN", ""),                                           // 10
        key[0].clone(),                                                  // 11
        key_line("END", ""),                                             // 12
        "zq-after".to_string(),                                          // 13
        format!("## {heading}"),                                         // 14
        "zq-under".to_string(),                                          // 15
        "## zq-open".to_string(),                                        // 16
        key_line("BEGIN", "RSA "),                                       // 17
        key[1].clone(),                                                  // 18
        "## zq-inside-the-key".to_string(),                              // 19
        key[2].clone(),                                                  // 20: never ended
    ];
    fx.write("AGENTS.md", body.join("\n"));
    fx.write("CLAUDE.local.md", "## zq-clean\nzq-clean line\n");
    let result = fx.plan();
    let visible = [&body[0..2], &body[6..9], &body[12..13]].concat();
    assert_hunk(
        &result,
        "AGENTS.md:1",
        Some("# zq-secrets"),
        &visible.join("\n"),
        7,
    );
    assert_hunk(&result, "AGENTS.md:14", None, "zq-under", 1);
    assert_hunk(&result, "AGENTS.md:16", Some("## zq-open"), "## zq-open", 2);
    assert_hunk(&result, "AGENTS.md:19", None, "", 2);

    // Kept as personal, the values are composed and still not printed.
    let all = [
        "AGENTS.md:1=personal",
        "AGENTS.md:14=personal",
        "AGENTS.md:16=personal",
        "AGENTS.md:19=personal",
    ];
    let result = fx.plan_with(&all);
    assert_apply(&result);
    let local = composed(&result, "CLAUDE.local.md").expect("the personal file");
    assert_eq!(local["hidden_lines"], json!(12), "{local}");
    let shown = composed_text(&result, "CLAUDE.local.md");
    assert!(shown.contains("zq-after") && shown.contains("zq-under"));
    assert_eq!(
        local["line_count"],
        json!(shown.lines().count() + 12),
        "{local}"
    );

    // Kept as shared, every such hunk is named and no command is printed.
    for id in [
        "AGENTS.md:1",
        "AGENTS.md:14",
        "AGENTS.md:16",
        "AGENTS.md:19",
    ] {
        let result = fx.plan_with(&[&format!("{id}=shared")]);
        assert_no_apply(&result, &[id, "credential"]);
    }
    // A hunk without such a line may be shared while others hold one.
    assert_apply(&fx.plan_with(&["CLAUDE.local.md:1=shared"]));

    // A refused pick value does not echo a credential.
    for keep in [
        format!("{token}=shared"),
        format!("AGENTS.md:1={token}"),
        token.clone(),
    ] {
        let out = fx.run_plan(&[&keep]);
        assert_ne!(out.code, Some(0), "{}{}", out.stdout, out.stderr);
    }
}

#[test]
fn hunk_a_private_key_that_begins_in_the_original_hides_the_added_lines() {
    let mut fx = Fx::new();
    let begin = key_line("BEGIN", "");
    let end = key_line("END", "");
    let base = format!("# zq-k\n{begin}\n{end}\n");
    fx.record("AGENTS.md", &base);
    fx.manifest();
    fx.git_init();
    fx.write("AGENTS.md", &base);
    fx.commit(&["AGENTS.md"]);
    let added = fx.secret("V");
    fx.write("AGENTS.md", format!("# zq-k\n{begin}\n{added}\n{end}\n"));
    let result = fx.plan();
    assert_eq!(ids(&result), ["AGENTS.md:3"], "{result}");
    assert_hunk(&result, "AGENTS.md:3", Some("# zq-k"), "", 1);
}

#[test]
fn hunk_a_hunk_names_the_moved_paths_its_text_names() {
    let fx = Fx::new();
    let dir = rules_dir();
    let kanban = rules("kanban.md");
    fx.write("CLAUDE.md", "zq-claude\n");
    fx.write(&kanban, "zq-kanban\n");
    fx.write(
        "AGENTS.md",
        format!(
            "## zq-plain\nzq read AGENTS.md first.\n\
             ## zq-forms\nzq `./CLAUDE.md` and @./AGENTS.md를 zq\n\
             ## zq-folder\nzq see {dir}/ for more\n\
             ## zq-file\nzq ({kanban})\n\
             ## zq-miss\nzq .claude/CLAUDE.md xAGENTS.md AGENTS.md.bak sub/AGENTS.md\n\
             zq AGENTS.md-old CLAUDE.local.md {dir}-old\n"
        ),
    );
    let result = fx.plan();
    for (id, want) in [
        ("AGENTS.md:1", vec!["AGENTS.md".to_string()]),
        (
            "AGENTS.md:3",
            vec!["CLAUDE.md".to_string(), "AGENTS.md".to_string()],
        ),
        ("AGENTS.md:5", vec![dir.clone()]),
        ("AGENTS.md:7", vec![kanban.clone(), dir.clone()]),
        // Other files, and a source that does not exist.
        ("AGENTS.md:9", vec![]),
    ] {
        assert_eq!(
            hunk(&result, id)["moved_paths"],
            json!(want),
            "{id}: {result}"
        );
    }
}

#[test]
fn hunk_picks_refuse_an_unknown_id_a_duplicate_and_a_bad_destination() {
    let fx = Fx::new();
    fx.profile(json!(null));
    fx.write("AGENTS.md", "## zq-a\nzq-a line\n");
    let cases: [(&[&str], &str, &str); 7] = [
        (&["AGENTS.md:2=shared"], "AGENTS.md:2", "no hunk"),
        (&["zq.md:1=personal"], "zq.md:1", "no hunk"),
        (
            &["AGENTS.md:1=shared", "AGENTS.md:1=personal"],
            "AGENTS.md:1",
            "twice",
        ),
        (
            &["AGENTS.md:1=shared", "AGENTS.md:1=shared"],
            "AGENTS.md:1",
            "twice",
        ),
        (&["AGENTS.md:1=team"], "AGENTS.md:1=team", "destination"),
        (&["AGENTS.md:1="], "AGENTS.md:1=", "destination"),
        (&["AGENTS.md:1"], "AGENTS.md:1", "destination"),
    ];
    for (keep, named, why) in cases {
        let out = fx.run_plan(keep);
        assert_ne!(out.code, Some(0), "{keep:?}: {}{}", out.stdout, out.stderr);
        assert!(
            out.stdout.contains(named) && out.stdout.contains(why),
            "{keep:?}: {}{}",
            out.stdout,
            out.stderr
        );
    }
}

#[test]
fn hunk_composition_keeps_the_picked_hunks_in_source_order() {
    let fx = Fx::new();
    fx.profile(json!(null));
    let kanban = rules("kanban.md");
    let kept = [
        "## zq-a",
        "zq-a rule",
        "zq `@AGENTS.md` quoted",
        "zq ``x @AGENTS.md y`` span",
        "@docs/zq-other.md",
        "```sh",
        "@AGENTS.md",
        "# zq-not-a-heading",
        "```",
    ];
    let mut agents: Vec<&str> = vec![
        "## zq-a",                       // 1
        "zq-a rule",                     // 2
        "@AGENTS.md",                    // 3
        "zq see @./CLAUDE.local.md now", // 4
    ];
    agents.extend(&kept[2..]); // 5..=11
    agents.extend([
        "<!-- moai:learned-start id=\"zq\" -->", // 12: never a project line
        "zq mid <!-- moai:harness-start --> text", // 13
        "",                                      // 14
        "## zq-b",                               // 15
        "zq-b rule",                             // 16
        "## zq-unpicked",                        // 17
        "zq-unpicked rule",                      // 18
    ]);
    fx.write("AGENTS.md", agents.join("\n"));
    fx.write(
        "CLAUDE.local.md",
        "## zq-local\nzq-mine\n`x` @AGENTS.md\nzq <!-- moai:learned-end --> tail\n",
    );
    fx.write(
        &kanban,
        format!(
            "## zq-rules\n@sub/deep.md\n@./gitflow.md\n@{kanban}\n@sub/zq-none.md\n\
             zq-rules end\n"
        ),
    );
    fx.write(&rules("gitflow.md"), "zq-gitflow\n");
    fx.write(&rules("sub/deep.md"), "zq-deep\n");

    // With no pick, `CLAUDE.md` is the managed block alone and there is no personal file.
    let result = fx.plan();
    let block_text = composed_text(&result, "CLAUDE.md");
    let (begin, end) = block(AGENTS_SECTION);
    assert!(
        block_text.starts_with(&format!("{begin}\n"))
            && block_text.ends_with(&format!("{end}\n"))
            && block_text.contains("zq-demo"),
        "{block_text}"
    );
    assert_eq!(list(&result, "results").len(), 1, "{result}");
    assert_eq!(result["picks"], json!([]));

    // Shared picks alone make no personal file; the order of the picks does not matter.
    let kanban_pick = format!("{kanban}:1=shared");
    let shared = [
        kanban_pick.as_str(),
        "AGENTS.md:15=shared",
        "AGENTS.md:1=shared",
    ];
    let result = fx.plan_with(&shared);
    let want = format!(
        "{block_text}\n{}\n\n## zq-b\nzq-b rule\n\n## zq-rules\n@sub/zq-none.md\nzq-rules end\n",
        kept.join("\n")
    );
    assert_eq!(composed_text(&result, "CLAUDE.md"), want, "{result}");
    assert_eq!(list(&result, "results").len(), 1, "{result}");
    assert_eq!(
        composed(&result, "CLAUDE.md").unwrap()["hunks"],
        json!(["AGENTS.md:1", "AGENTS.md:15", format!("{kanban}:1")])
    );
    let import = "imports_a_moved_file".to_string();
    let mut out: Vec<(String, String)> = vec![
        ("AGENTS.md:3".into(), import.clone()),
        ("AGENTS.md:4".into(), import.clone()),
        ("AGENTS.md:13".into(), "go_curator_marker".into()),
    ];
    out.extend((2..=4).map(|n| (format!("{kanban}:{n}"), import.clone())));
    assert_eq!(left_out(&result, "CLAUDE.md"), out, "{result}");

    // A personal pick makes the personal file: moai's first text, then the hunk.
    let result = fx.plan_with(&["CLAUDE.local.md:1=personal"]);
    assert_eq!(composed_text(&result, "CLAUDE.md"), block_text);
    let local = composed_text(&result, "CLAUDE.local.md");
    let first = local
        .strip_suffix("\n## zq-local\nzq-mine\n")
        .unwrap_or_else(|| panic!("the hunk does not close the file: {local}"));
    assert!(is_generated_local(first.as_bytes()), "{first}");
    assert_eq!(
        left_out(&result, "CLAUDE.local.md"),
        vec![
            ("CLAUDE.local.md:3".to_string(), import),
            (
                "CLAUDE.local.md:4".to_string(),
                "go_curator_marker".to_string()
            ),
        ]
    );
    assert_eq!(
        result["picks"],
        json!([{"id": "CLAUDE.local.md:1", "destination": "personal"}])
    );
}

#[test]
fn hunk_a_byte_order_mark_is_no_text_of_a_line() {
    let fx = Fx::new();
    // A file of the mark alone, as some editors save an empty file.
    fx.write(&rules("bom.md"), "\u{FEFF}");
    fx.write("AGENTS.md", "\u{FEFF}# zq-marked\nzq line\n");
    let result = fx.plan();
    assert_eq!(ids(&result), ["AGENTS.md:1"], "{result}");
    assert_hunk(
        &result,
        "AGENTS.md:1",
        Some("# zq-marked"),
        "# zq-marked\nzq line",
        0,
    );
}

#[test]
fn hunk_a_url_user_name_is_not_a_credential() {
    let fx = Fx::new();
    fx.profile(json!(null));
    fx.go_marker();
    let lines = "## zq-url\nzq ssh://git@example.invalid/org/repo.git\n\
                 zq https://deploy@example.invalid/x";
    fx.write("AGENTS.md", lines);
    let result = fx.plan_with(&["AGENTS.md:1=shared"]);
    assert_hunk(&result, "AGENTS.md:1", Some("## zq-url"), lines, 0);
    assert_apply(&result);
}

#[test]
fn hunk_an_import_that_loses_its_fence_is_left_out() {
    let mut fx = Fx::new();
    fx.profile(json!(null));
    let base = "# zq\n```\n```\n";
    fx.record("AGENTS.md", base);
    fx.manifest();
    fx.git_init();
    fx.write("AGENTS.md", base);
    fx.commit(&["AGENTS.md"]);
    // Both added lines are code here; the fence lines are the original's, so the new file would
    // hold them outside a fence.
    fx.write("AGENTS.md", "# zq\n```\n@AGENTS.md\nzq-example\n```\n");
    let block_text = composed_text(&fx.plan(), "CLAUDE.md");
    let result = fx.plan_with(&["AGENTS.md:3=shared"]);
    assert_eq!(
        composed_text(&result, "CLAUDE.md"),
        format!("{block_text}\nzq-example\n"),
        "{result}"
    );
}

#[test]
fn hunk_the_composed_language_follows_the_profile() {
    for (language, korean) in [
        (json!(null), false),
        (json!({"instruction_files": "ko"}), true),
        (json!({"user_chat": "ko"}), true),
        (json!({"instruction_files": "en", "user_chat": "ko"}), false),
    ] {
        let fx = Fx::new();
        fx.profile(language.clone());
        fx.write("AGENTS.md", "## zq-a\nzq-a line\n");
        let result = fx.plan_with(&["AGENTS.md:1=personal"]);
        let shared = composed_text(&result, "CLAUDE.md");
        let personal = composed_text(&result, "CLAUDE.local.md");
        assert_eq!(
            (
                shared.contains("## MoAI 프로젝트 사실"),
                shared.contains("## MoAI project facts"),
                personal.starts_with("# 개인 작업 보충\n"),
                personal.starts_with("# Personal supplement\n"),
            ),
            (korean, !korean, korean, !korean),
            "{language}: {shared}{personal}"
        );
    }
}

#[test]
fn hunk_a_result_over_the_line_limit_is_warned() {
    let fx = Fx::new();
    fx.profile(json!(null));
    fx.go_marker();
    let section = |lines: usize| format!("## zq-big\n{}", "zq-line\n".repeat(lines - 1));
    for (path, keep) in [
        ("CLAUDE.md", "AGENTS.md:1=shared"),
        ("CLAUDE.local.md", "AGENTS.md:1=personal"),
    ] {
        fx.write("AGENTS.md", section(10));
        let small = fx.plan_with(&[keep]);
        let around = composed(&small, path).unwrap()["line_count"]
            .as_u64()
            .unwrap() as usize
            - 10;
        for (total, warned) in [(RESULT_LINE_LIMIT, false), (RESULT_LINE_LIMIT + 1, true)] {
            fx.write("AGENTS.md", section(total - around));
            let result = fx.plan_with(&[keep]);
            assert_eq!(
                composed(&result, path).unwrap()["line_count"],
                json!(total),
                "{result}"
            );
            let warnings = words(list(&result, "warnings"));
            if warned {
                assert!(
                    warnings.len() == 1
                        && warnings[0].starts_with(&format!("{path} "))
                        && warnings[0].contains(&total.to_string()),
                    "{warnings:?}"
                );
            } else {
                assert!(warnings.is_empty(), "{path}: {warnings:?}");
            }
            // A warning never withholds the command.
            assert_apply(&result);
        }
    }
}

#[test]
fn hunk_a_claude_md_in_the_claude_folder_is_warned_and_never_read() {
    let rel = ".claude/CLAUDE.md";
    for pipe in [false, true] {
        let fx = Fx::new();
        fx.profile(json!(null));
        fx.go_marker();
        assert_eq!(fx.plan()["warnings"], json!([]));
        if pipe {
            fs::create_dir_all(fx.root.join(".claude")).unwrap();
            let made = Command::new("/usr/bin/mkfifo")
                .arg(fx.root.join(rel))
                .status()
                .unwrap();
            assert!(made.success(), "mkfifo");
        } else {
            fx.write(rel, "## zq-dot\nzq-dot-claude\n");
        }
        let result = fx.plan();
        let warnings = words(list(&result, "warnings"));
        assert!(
            warnings.len() == 1 && warnings[0].contains(rel),
            "{warnings:?}"
        );
        assert!(!result.to_string().contains("zq-dot"), "{result}");
        // The root file alone is a source.
        assert_eq!(paths(&result), ["CLAUDE.md"], "{result}");
        assert_apply(&result);
    }
}

#[test]
fn hunk_the_digest_binds_the_sources_the_profile_and_the_picks() {
    let fx = Fx::new();
    fx.profile(json!(null));
    fx.go_marker();
    let agents = "## zq-a\nzq-a line\n## zq-b\nzq-b line\n";
    fx.write("AGENTS.md", agents);
    let (zero, argv) = assert_apply(&fx.plan());
    let root = fx.root.to_str().unwrap();
    let command = |keep: &[&str], digest: &str| {
        let mut want = vec!["moai", "claude-md-apply", root];
        for word in keep {
            want.extend(["--keep", word]);
        }
        want.extend([
            "--plan-digest",
            digest,
            "--permission-reference",
            PERMISSION_PLACEHOLDER,
        ]);
        want.into_iter().map(String::from).collect::<Vec<_>>()
    };
    assert_eq!(argv, command(&[], &zero));
    assert_eq!(assert_apply(&fx.plan()).0, zero, "the digest is stable");

    // The picks: which, and where to; not their order.
    let (a, b) = ("AGENTS.md:1=shared", "AGENTS.md:3=personal");
    let (both, argv) = assert_apply(&fx.plan_with(&[b, a]));
    assert_eq!(argv, command(&[a, b], &both));
    assert_eq!(assert_apply(&fx.plan_with(&[a, b])).0, both);
    let mut seen = vec![zero.clone(), both];
    for keep in [a, "AGENTS.md:1=personal", "AGENTS.md:3=shared"] {
        seen.push(assert_apply(&fx.plan_with(&[keep])).0);
    }

    // A source byte that changes no resulting text.
    fx.write("AGENTS.md", agents.replace("zq-b line", "zq-B line"));
    seen.push(assert_apply(&fx.plan()).0);
    fx.write("AGENTS.md", agents);
    assert_eq!(assert_apply(&fx.plan()).0, zero);
    // A new source.
    fx.write(&rules("new.md"), "");
    seen.push(assert_apply(&fx.plan()).0);
    fs::remove_file(fx.root.join(rules("new.md"))).unwrap();
    // A profile byte that changes no resulting text.
    let profile = fs::read_to_string(fx.root.join(profile_rel())).unwrap();
    fx.write(&profile_rel(), format!("{profile}\n"));
    seen.push(assert_apply(&fx.plan()).0);

    let distinct: std::collections::BTreeSet<&String> = seen.iter().collect();
    assert_eq!(distinct.len(), seen.len(), "{seen:?}");
}

#[test]
fn hunk_no_command_without_a_usable_profile_or_with_a_refused_source() {
    let fx = Fx::new();
    fx.go_marker();
    fx.write("AGENTS.md", "## zq-a\nzq-a line\n");
    let keep = ["AGENTS.md:1=shared"];

    // No profile: the hunks are listed, nothing is composed.
    let result = fx.plan_with(&keep);
    assert_no_apply(&result, &["project profile missing", "/moai project"]);
    assert_eq!(ids(&result), ["AGENTS.md:1"]);
    assert_eq!(result["results"], json!([]));

    fx.write(&profile_rel(), "zq: not json\n");
    assert_no_apply(&fx.plan_with(&keep), &["/moai project"]);

    // The placeholder a first launch writes.
    fx.write(
        &profile_rel(),
        json!({"schema_version": 2, "project_id": "zq-demo", "purpose": "zq",
               "tools": ["unspecified"], "development_methods": ["unspecified"],
               "hosts": ["claude"]})
        .to_string(),
    );
    let result = fx.plan_with(&keep);
    assert_no_apply(&result, &["placeholder", "/moai project"]);
    assert_eq!(result["results"], json!([]));

    // A value that would open or close a managed block.
    fx.profile(json!(null));
    let profile = fs::read_to_string(fx.root.join(profile_rel())).unwrap();
    fx.write(
        &profile_rel(),
        profile.replace("zq purpose", "zq <!-- purpose"),
    );
    let result = fx.plan_with(&keep);
    assert_no_apply(&result, &["managed-marker"]);
    assert_eq!(result["results"], json!([]));

    // A usable profile prints the command; one refused source withholds it.
    fx.profile(json!(null));
    assert_apply(&fx.plan_with(&keep));
    let outside = fx.base.join("outside.md");
    fs::write(&outside, "zq-outside\n").unwrap();
    fx.link("CLAUDE.local.md", &outside);
    let result = fx.plan_with(&keep);
    assert_no_apply(&result, &["CLAUDE.local.md", "is a link"]);
    assert!(!result.to_string().contains("zq-outside"), "{result}");
}

#[test]
fn base_an_empty_project_has_no_source_and_a_missing_one_is_an_error() {
    let fx = Fx::new();
    let result = fx.plan();
    assert_eq!(paths(&result), Vec::<String>::new());
    assert_eq!(refused(&result), vec![]);
    assert_eq!(result["go_claude_md"], json!(null));
    assert_eq!(result["go_manifest"], json!(null));

    let before = fx.snapshot();
    let missing = fx.base.join("no-such-project");
    let out = fx.run(&["claude-md-plan", missing.to_str().unwrap()]);
    assert_ne!(out.code, Some(0), "{}{}", out.stdout, out.stderr);
    assert!(
        out.stdout.contains("the project is not a folder"),
        "{}{}",
        out.stdout,
        out.stderr
    );
    assert_eq!(fx.snapshot(), before);
}

/// The ignore rule that keeps the backup folder out of git.
fn backups_rule() -> String {
    format!("{MOAI_DIR}/backups/")
}

/// [`go_project`] with a usable profile, whose `.gitignore` keeps the backup folder and both
/// personal files out of git.
fn apply_project() -> Fx {
    let fx = go_project();
    fx.profile(json!(null));
    fx.write(
        ".gitignore",
        format!("AGENTS.local.md\nCLAUDE.local.md\n{}\n", backups_rule()),
    );
    fx
}

/// The sources of [`go_project`] in the order the step lists them.
fn go_sources() -> Vec<String> {
    let mut sources: Vec<String> = [
        "CLAUDE.md",
        "AGENTS.md",
        "AGENTS.local.md",
        "CLAUDE.local.md",
    ]
    .map(String::from)
    .to_vec();
    sources.extend(["gitflow.md", "kanban.md", "sub/deep.md", "zz-new.md"].map(rules));
    sources
}

const APPROVAL: &str = "lane-approval-1";

fn mode(path: &Path) -> u32 {
    fs::symlink_metadata(path).unwrap().permissions().mode() & 0o777
}

fn absent(path: &Path) -> bool {
    fs::symlink_metadata(path).is_err()
}

impl Fx {
    fn chmod(&self, rel: &str, mode: u32) {
        fs::set_permissions(self.root.join(rel), fs::Permissions::from_mode(mode)).unwrap();
    }

    fn read(&self, rel: &str) -> Vec<u8> {
        fs::read(self.root.join(rel)).unwrap_or_else(|e| panic!("{rel}: {e}"))
    }

    /// The command `claude-md-plan` prints for these picks, with the permission filled in, as
    /// the arguments after `moai`.
    fn command(&self, keep: &[&str]) -> Vec<String> {
        let (_, argv) = assert_apply(&self.plan_with(keep));
        assert_eq!(argv[0], "moai");
        argv[1..]
            .iter()
            .map(|word| {
                if word == PERMISSION_PLACEHOLDER {
                    APPROVAL.to_string()
                } else {
                    word.clone()
                }
            })
            .collect()
    }

    fn run_words(&self, words: &[String], extra: &[(&str, &str)]) -> Out {
        let args: Vec<&str> = words.iter().map(String::as_str).collect();
        self.run_env(&args, extra)
    }

    /// Runs the printed command, which must succeed.
    fn apply(&self, words: &[String]) -> Value {
        let out = self.run_words(words, &[]);
        assert_eq!(
            out.code,
            Some(0),
            "claude-md-apply failed: stdout {} stderr {}",
            out.stdout,
            out.stderr
        );
        let v: Value = serde_json::from_str(&out.stdout).expect("a JSON answer");
        assert_eq!(v["operation"], json!("claude_md.apply"), "{v}");
        assert_eq!(v["ok"], json!(true), "{v}");
        v["result"].clone()
    }

    /// A failed apply's error kind and message.
    fn failed(&self, words: &[String], extra: &[(&str, &str)]) -> (String, String) {
        let out = self.run_words(words, extra);
        assert_eq!(
            out.code,
            Some(1),
            "{words:?}: stdout {} stderr {}",
            out.stdout,
            out.stderr
        );
        let v: Value = serde_json::from_str(&out.stdout).expect("a JSON answer");
        assert_eq!(v["operation"], json!("claude_md.apply"), "{v}");
        assert_eq!(v["ok"], json!(false), "{v}");
        (
            v["error"]["kind"].as_str().unwrap().to_string(),
            v["error"]["message"].as_str().unwrap().to_string(),
        )
    }

    /// The apply refuses before its first write: it names every needle and leaves every byte,
    /// `.git` included, as it was. Returns the error kind.
    fn refuses(&self, words: &[String], needles: &[&str]) -> String {
        let before = self.snapshot();
        let (kind, message) = self.failed(words, &[]);
        for needle in needles.iter().chain(&["refused", "nothing was written"]) {
            assert!(message.contains(needle), "no {needle:?} in: {message}");
        }
        assert_eq!(self.snapshot(), before, "a refused apply wrote: {message}");
        kind
    }

    /// The preview withholds the digest and the command for this reason, and the apply of an
    /// earlier command refuses for it too.
    fn blocks(&self, keep: &[&str], words: &[String], needles: &[&str]) {
        assert_no_apply(&self.plan_with(keep), needles);
        self.refuses(words, needles);
    }

    fn receipt_rel(stamp: &str) -> String {
        format!("{MOAI_DIR}/backups/{stamp}/claude-md-receipt.json")
    }

    fn area(stamp: &str) -> String {
        format!("{MOAI_DIR}/backups/{stamp}/claude-md")
    }

    fn receipt(&self, stamp: &str) -> Value {
        serde_json::from_slice(&self.read(&Fx::receipt_rel(stamp))).expect("a receipt")
    }

    /// The stamp of the one apply that left a backup folder.
    fn only_stamp(&self) -> String {
        let names: Vec<String> = fs::read_dir(self.root.join(MOAI_DIR).join("backups"))
            .expect("a backup folder")
            .map(|e| e.unwrap().file_name().into_string().unwrap())
            .collect();
        assert_eq!(names.len(), 1, "{names:?}");
        names[0].clone()
    }

    /// Whether git ignores the path.
    fn ignored(&self, rel: &str) -> bool {
        common::git::command()
            .args(["check-ignore", "--no-index", "-q", "--", rel])
            .current_dir(&self.root)
            .status()
            .unwrap()
            .success()
    }
}

/// The `moved` flags of a receipt, in the order of its moves.
fn moved_flags(receipt: &Value) -> Vec<bool> {
    list(receipt, "moves")
        .iter()
        .map(|m| m["moved"].as_bool().unwrap())
        .collect()
}

/// Each written path of a receipt with its flag.
fn written_flags(receipt: &Value) -> Vec<(String, bool)> {
    list(receipt, "writes")
        .iter()
        .map(|w| {
            (
                w["path"].as_str().unwrap().to_string(),
                w["written"].as_bool().unwrap(),
            )
        })
        .collect()
}

#[test]
fn apply_the_printed_command_moves_every_source_and_writes_the_picked_hunks() {
    let fx = apply_project();
    // A project line that imports a moved file and one that holds a curator marker.
    fx.write(
        "AGENTS.md",
        "# zq-go agents\n\nzq-template one\nzq-project rule\n@AGENTS.local.md\n\
         zq <!-- moai:learned-start --> x\nzq-template two\n",
    );
    let sources = go_sources();
    let originals: Vec<Vec<u8>> = sources.iter().map(|p| fx.read(p)).collect();
    let kanban = format!("{}:1=shared", rules("kanban.md"));
    let keep = [
        "AGENTS.md:4=shared",
        kanban.as_str(),
        "CLAUDE.local.md:1=personal",
    ];
    let preview = fx.plan_with(&keep);
    let shared = composed_text(&preview, "CLAUDE.md");
    let personal = composed_text(&preview, "CLAUDE.local.md");
    let ignore_before = fx.read(".gitignore");

    let words = fx.command(&keep);
    let result = fx.apply(&words);

    assert_eq!(result["state"], json!("committed"), "{result}");
    assert_eq!(result["moved"], json!(sources), "{result}");
    assert_eq!(
        result["written"],
        json!(["CLAUDE.md", "CLAUDE.local.md"]),
        "{result}"
    );
    assert_eq!(result["ignore_rule"], json!(null), "{result}");
    assert_eq!(result["backup_ignore"], json!("git"), "{result}");
    assert_eq!(result["plan_digest"], preview["plan_digest"], "{result}");
    let notice = result["notice"].as_str().expect("a notice");
    assert!(
        notice.contains("settings apply at once")
            && notice.contains("running sessions and lanes")
            && notice.contains("restart"),
        "{notice}"
    );
    let stamp = result["receipt"].as_str().expect("a receipt id");
    assert_eq!(stamp, fx.only_stamp());
    assert_eq!(result["backup"], json!(Fx::area(stamp)), "{result}");

    // The new files: the managed block and the picked hunks, nothing else.
    assert_eq!(fx.read("CLAUDE.md"), shared.as_bytes());
    assert_eq!(fx.read("CLAUDE.local.md"), personal.as_bytes());
    let (begin, end) = block(AGENTS_SECTION);
    assert!(shared.starts_with(&format!("{begin}\n")), "{shared}");
    assert_eq!(
        shared.split_once(&format!("{end}\n")).expect("a block").1,
        "\nzq-project rule\n\n# zq-kanban\nzq-old rule\n",
        "{shared}"
    );
    for gone in [
        "@AGENTS.md",
        "@AGENTS.local.md",
        "moai:learned",
        "zq-template",
    ] {
        assert!(!shared.contains(gone), "{gone}: {shared}");
    }
    assert!(personal.ends_with("\n# zq-local\nzq-other\n"), "{personal}");
    assert!(fx.ignored("CLAUDE.local.md"));

    // Every source lies in the backup byte for byte, out of git; nothing was deleted.
    for (path, original) in sources.iter().zip(&originals) {
        let backup = format!("{}/{path}", Fx::area(stamp));
        assert_eq!(&fx.read(&backup), original, "{path}");
        assert!(fx.ignored(&backup), "{backup}");
        if !["CLAUDE.md", "CLAUDE.local.md"].contains(&path.as_str()) {
            assert!(absent(&fx.root.join(path)), "{path} stayed");
        }
    }
    assert!(
        fx.root.join(rules("sub")).is_dir(),
        "an emptied folder stays"
    );
    assert!(fx.ignored(&Fx::receipt_rel(stamp)));
    assert!(!fx
        .git(&["status", "--porcelain", "--untracked-files=all"])
        .contains("backups"));
    assert_eq!(fx.read(".gitignore"), ignore_before);

    // The new file is no Go-generation file: nothing more is offered, and the command that was
    // run is refused.
    let again = fx.plan();
    assert_eq!(again["go_claude_md"], json!(null), "{again}");
    assert_no_apply(&again, &["not a Go-generation instruction file"]);
    // The run command picks a hunk of a file that was moved, which is what refuses it first;
    // without the picks the reason is the new file.
    fx.refuses(&words, &["no hunk has this id"]);
    let unpicked: Vec<String> = words
        .iter()
        .enumerate()
        .filter(|(i, word)| *word != "--keep" && (*i == 0 || words[i - 1] != "--keep"))
        .map(|(_, word)| word.clone())
        .collect();
    fx.refuses(&unpicked, &["not a Go-generation instruction file"]);
}

#[test]
fn apply_without_a_personal_pick_writes_no_personal_file() {
    for keep in [&[][..], &["AGENTS.md:4=shared"][..]] {
        let fx = apply_project();
        let shared = composed_text(&fx.plan_with(keep), "CLAUDE.md");
        let result = fx.apply(&fx.command(keep));
        assert_eq!(result["written"], json!(["CLAUDE.md"]), "{result}");
        assert_eq!(fx.read("CLAUDE.md"), shared.as_bytes());
        assert!(absent(&fx.root.join("CLAUDE.local.md")));
        let stamp = fx.only_stamp();
        assert_eq!(
            fx.read(&format!("{}/CLAUDE.local.md", Fx::area(&stamp))),
            b"# zq-local\nzq-other\n"
        );
        assert_eq!(
            written_flags(&fx.receipt(&stamp)),
            [("CLAUDE.md".to_string(), true)]
        );
    }
}

#[test]
fn apply_adds_the_ignore_rule_when_no_rule_ignores_the_personal_file() {
    let rule = backups_rule();
    // (the `.gitignore` before, the one after); `None`: no such file.
    let cases: [(Option<String>, String); 3] = [
        (
            Some(format!("AGENTS.local.md\n{rule}")),
            format!("AGENTS.local.md\n{rule}\nCLAUDE.local.md\n"),
        ),
        (
            Some(format!("{rule}\r\n")),
            format!("{rule}\r\nCLAUDE.local.md\r\n"),
        ),
        (None, "CLAUDE.local.md\n".to_string()),
    ];
    for (before, after) in cases {
        let fx = apply_project();
        match &before {
            Some(text) => fx.write(".gitignore", text),
            None => {
                fs::remove_file(fx.root.join(".gitignore")).unwrap();
                fx.write(".git/info/exclude", format!("{rule}\n"));
            }
        }
        assert!(!fx.ignored("CLAUDE.local.md"));
        let result = fx.apply(&fx.command(&["CLAUDE.local.md:1=personal"]));
        assert_eq!(fx.read(".gitignore"), after.as_bytes(), "{before:?}");
        assert!(fx.ignored("CLAUDE.local.md"), "{before:?}");
        assert_eq!(
            result["written"],
            json!([".gitignore", "CLAUDE.md", "CLAUDE.local.md"]),
            "{result}"
        );
        assert_eq!(
            result["ignore_rule"],
            json!({"path": ".gitignore", "line": "CLAUDE.local.md"}),
            "{result}"
        );
        let receipt = fx.receipt(&fx.only_stamp());
        assert_eq!(
            receipt["ignore"],
            json!({
                "path": ".gitignore",
                "line": "CLAUDE.local.md",
                "before_sha256": before.as_ref().map(|text| sha(text.as_bytes())),
                "before_len": before.as_ref().map_or(0, String::len),
                "after_sha256": sha(after.as_bytes()),
                "written": true,
            }),
            "{before:?}: {receipt}"
        );
    }
    // A rule outside `.gitignore` already ignores the file: nothing is added.
    let fx = apply_project();
    fx.write(".gitignore", format!("{rule}\n"));
    fx.write(".git/info/exclude", "CLAUDE.local.md\n");
    let result = fx.apply(&fx.command(&["CLAUDE.local.md:1=personal"]));
    assert_eq!(result["ignore_rule"], json!(null), "{result}");
    assert_eq!(fx.read(".gitignore"), format!("{rule}\n").as_bytes());
    assert_eq!(fx.receipt(&fx.only_stamp())["ignore"], json!(null));
}

#[test]
fn apply_saves_a_receipt_of_paths_digests_and_modes() {
    let mut fx = apply_project();
    let token = format!("ghp_{}", fx.secret("Q"));
    fx.write("CLAUDE.local.md", format!("# zq-local\nzq-key {token}\n"));
    let modes = [
        ("CLAUDE.md".to_string(), 0o640),
        ("AGENTS.md".to_string(), 0o644),
        ("AGENTS.local.md".to_string(), 0o600),
        ("CLAUDE.local.md".to_string(), 0o644),
        (rules("kanban.md"), 0o400),
    ];
    for (path, bits) in &modes {
        fx.chmod(path, *bits);
    }
    let sources = go_sources();
    let facts: Vec<(Vec<u8>, u32)> = sources
        .iter()
        .map(|p| (fx.read(p), mode(&fx.root.join(p))))
        .collect();
    let keep = ["CLAUDE.local.md:1=personal"];
    let digest = fx.plan_with(&keep)["plan_digest"].clone();
    let words = fx.command(&keep);
    let result = fx.apply(&words);
    let stamp = result["receipt"].as_str().unwrap().to_string();

    // A personal file is its owner's alone, `CLAUDE.md` keeps the mode it had, and each backup
    // keeps its source's mode in a folder that is its owner's alone.
    assert_eq!(mode(&fx.root.join("CLAUDE.local.md")), 0o600);
    assert_eq!(mode(&fx.root.join("CLAUDE.md")), 0o640);
    for (path, (_, bits)) in sources.iter().zip(&facts) {
        let backup = fx.root.join(Fx::area(&stamp)).join(path);
        assert_eq!(mode(&backup), *bits, "{path}");
    }
    assert_eq!(
        mode(&fx.root.join(MOAI_DIR).join("backups").join(&stamp)),
        0o700
    );
    assert_eq!(mode(&fx.root.join(Fx::receipt_rel(&stamp))), 0o600);

    let tracked = |path: &str| {
        let personal = ["AGENTS.local.md", "CLAUDE.local.md"].contains(&path);
        if personal || path == rules("zz-new.md") {
            "untracked"
        } else {
            "tracked"
        }
    };
    let moves: Vec<Value> = sources
        .iter()
        .zip(&facts)
        .map(|(path, (bytes, bits))| {
            json!({
                "path": path,
                "backup": format!("{}/{path}", Fx::area(&stamp)),
                "sha256": sha(bytes),
                "mode": bits,
                "git": tracked(path),
                "moved": true,
            })
        })
        .collect();
    let writes: Vec<Value> = [("CLAUDE.md", 0o640), ("CLAUDE.local.md", 0o600)]
        .iter()
        .map(|(path, bits)| {
            json!({"path": path, "sha256": sha(&fx.read(path)), "mode": bits, "written": true})
        })
        .collect();
    assert_eq!(
        fx.receipt(&stamp),
        json!({
            "schema_version": 1,
            "operation": "claude_md.apply",
            "project": fx.root.to_str().unwrap(),
            "stamp": stamp,
            "plan_digest": digest,
            "permission_reference": APPROVAL,
            "state": "committed",
            "moves": moves,
            "writes": writes,
            "ignore": null,
        })
    );
    // The value went into the personal file and its backup, and into no receipt.
    let personal = String::from_utf8(fx.read("CLAUDE.local.md")).unwrap();
    assert!(personal.contains(&token));
    let receipt = String::from_utf8(fx.read(&Fx::receipt_rel(&stamp))).unwrap();
    assert!(!receipt.contains(&fx.secrets[0]) && !receipt.contains("zq-key"));
}

#[test]
fn apply_outside_a_git_repository_adds_no_ignore_rule() {
    let fx = Fx::new();
    fx.profile(json!(null));
    fx.go_marker();
    fx.write("CLAUDE.local.md", "## zq-local\nzq-mine\n");
    let result = fx.apply(&fx.command(&["CLAUDE.local.md:1=personal"]));
    assert_eq!(
        result["backup_ignore"],
        json!("not_a_repository"),
        "{result}"
    );
    assert_eq!(result["ignore_rule"], json!(null), "{result}");
    assert!(absent(&fx.root.join(".gitignore")));
    assert!(fx.read("CLAUDE.local.md").ends_with(b"zq-mine\n"));
    let receipt = fx.receipt(&fx.only_stamp());
    assert_eq!(receipt["ignore"], json!(null));
    assert_eq!(receipt["moves"][0]["git"], json!("unknown"), "{receipt}");
}

/// A failed apply that began: its message says `partial apply` and holds every needle, and the
/// receipt it left is pending.
fn stopped_apply(fx: &Fx, fault: &str, keep: &[&str], needles: &[&str]) -> (String, Value) {
    let words = fx.command(keep);
    let (kind, message) = fx.failed(&words, &[("MOAI_TEST_FAULT", fault)]);
    assert_eq!(kind, "partial", "{message}");
    for needle in needles.iter().chain(&["partial apply", "pending"]) {
        assert!(message.contains(needle), "no {needle:?} in: {message}");
    }
    let stamp = fx.only_stamp();
    assert!(message.contains(&Fx::receipt_rel(&stamp)), "{message}");
    let receipt = fx.receipt(&stamp);
    assert_eq!(receipt["state"], json!("pending"), "{receipt}");
    (stamp, receipt)
}

#[test]
fn apply_a_source_edited_before_its_move_stops_with_a_pending_receipt() {
    let fx = apply_project();
    let sources = go_sources();
    let original = fx.read("CLAUDE.md");
    let (stamp, receipt) = stopped_apply(
        &fx,
        "claude_md_source_edit",
        &[],
        &["AGENTS.md", "changed after the preview"],
    );
    let mut want = vec![false; sources.len()];
    want[0] = true;
    assert_eq!(moved_flags(&receipt), want, "{receipt}");
    assert_eq!(written_flags(&receipt), [("CLAUDE.md".to_string(), false)]);
    assert_eq!(
        fx.read(&format!("{}/CLAUDE.md", Fx::area(&stamp))),
        original
    );
    assert!(absent(&fx.root.join("CLAUDE.md")), "a new file was written");
    // The edited file and every later source stay where they are.
    assert!(fx.read("AGENTS.md").starts_with(EDITED_AGENTS.as_bytes()));
    for path in &sources[1..] {
        assert!(fx.root.join(path).is_file(), "{path}");
        assert!(absent(&fx.root.join(Fx::area(&stamp)).join(path)), "{path}");
    }
}

#[test]
fn apply_never_overwrites_a_file_at_a_backup_path() {
    let fx = apply_project();
    let (stamp, receipt) = stopped_apply(
        &fx,
        "claude_md_backup_taken",
        &[],
        &["AGENTS.md", "destination already exists"],
    );
    assert_eq!(moved_flags(&receipt)[..2], [true, false], "{receipt}");
    assert_eq!(fx.read("AGENTS.md"), EDITED_AGENTS.as_bytes());
    assert_eq!(
        fx.read(&format!("{}/AGENTS.md", Fx::area(&stamp))),
        b"fault\n"
    );
}

#[test]
fn apply_never_overwrites_a_file_that_appears_at_a_destination() {
    let fx = apply_project();
    let keep = ["CLAUDE.local.md:1=personal"];
    let (stamp, receipt) = stopped_apply(
        &fx,
        "claude_md_destination_taken",
        &keep,
        &["CLAUDE.local.md"],
    );
    assert!(
        moved_flags(&receipt).iter().all(|moved| *moved),
        "{receipt}"
    );
    assert_eq!(
        written_flags(&receipt),
        [
            ("CLAUDE.md".to_string(), true),
            ("CLAUDE.local.md".to_string(), false)
        ]
    );
    assert_eq!(fx.read("CLAUDE.local.md"), b"fault\n");
    assert_eq!(
        fx.read(&format!("{}/CLAUDE.local.md", Fx::area(&stamp))),
        b"# zq-local\nzq-other\n"
    );
}

#[test]
fn refuse_without_a_permission_reference_of_the_accepted_form() {
    let mut fx = apply_project();
    let token = format!("ghp_{}", fx.secret("Q"));
    let words = fx.command(&[]);
    let long = "a".repeat(129);
    for reference in [
        "",
        "  ",
        "null",
        "placeholder",
        PERMISSION_PLACEHOLDER,
        long.as_str(),
        "zq\nzq",
        token.as_str(),
    ] {
        let mut words = words.clone();
        *words.last_mut().unwrap() = reference.to_string();
        let kind = fx.refuses(&words, &["permission reference"]);
        assert_eq!(kind, "permission_reference_missing", "{reference:?}");
    }
    // The longest accepted one passes.
    let mut words = words;
    *words.last_mut().unwrap() = "a".repeat(128);
    fx.apply(&words);
}

#[test]
fn refuse_a_claude_md_that_is_no_go_generation_file() {
    // No manifest records it; another manifest still leaves its lines to the project.
    let fx = apply_project();
    let words = fx.command(&[]);
    fs::remove_file(fx.root.join(GO_BACKUP).join("manifest.json")).unwrap();
    fx.blocks(&[], &words, &["not a Go-generation instruction file"]);
    // No file at all.
    fs::remove_file(fx.root.join("CLAUDE.md")).unwrap();
    fx.blocks(&[], &words, &["not a Go-generation instruction file"]);
}

#[test]
fn refuse_a_missing_or_placeholder_profile() {
    let fx = apply_project();
    let words = fx.command(&[]);
    fs::remove_file(fx.root.join(profile_rel())).unwrap();
    fx.blocks(&[], &words, &["project profile missing", "/moai project"]);
    fx.write(
        &profile_rel(),
        json!({"schema_version": 2, "project_id": "zq-demo", "purpose": "zq",
               "tools": ["unspecified"], "development_methods": ["unspecified"],
               "hosts": ["claude"]})
        .to_string(),
    );
    fx.blocks(&[], &words, &["placeholder", "/moai project"]);
}

#[test]
fn refuse_a_source_that_must_not_be_read() {
    type Spoil = fn(&Fx);
    let cases: [(Spoil, String, &str); 4] = [
        (
            // Swapped for a link to the very same bytes after the preview.
            |fx| {
                let outside = fx.base.join("outside.md");
                fs::rename(fx.root.join("AGENTS.md"), &outside).unwrap();
                fx.link("AGENTS.md", &outside);
            },
            "AGENTS.md".to_string(),
            "is a link",
        ),
        (
            |fx| {
                let outside = fx.base.join("outside-rules");
                fs::rename(fx.root.join(rules_dir()), &outside).unwrap();
                fx.link(&rules_dir(), &outside);
            },
            rules_dir(),
            "is a link",
        ),
        (
            |fx| fx.write("AGENTS.local.md", [b'z', 0xff, 0xfe, b'\n']),
            "AGENTS.local.md".to_string(),
            "is not UTF-8 text",
        ),
        (
            |fx| {
                let over = usize::try_from(SOURCE_LIMIT).unwrap() + 1;
                fx.write(&rules("kanban.md"), vec![b'a'; over]);
            },
            rules("kanban.md"),
            "is larger than 1 MiB",
        ),
    ];
    for (spoil, path, why) in cases {
        let fx = apply_project();
        let words = fx.command(&[]);
        spoil(&fx);
        fx.blocks(&[], &words, &[&path, why]);
    }
}

#[test]
fn refuse_a_source_or_a_profile_changed_since_the_preview() {
    type Change = fn(&Fx);
    let changes: [Change; 5] = [
        |fx| fx.write("CLAUDE.local.md", "# zq-local\nzq-Other\n"),
        |fx| fx.write(&rules("later.md"), "zq-later\n"),
        |fx| fs::remove_file(fx.root.join("AGENTS.local.md")).unwrap(),
        |fx| {
            let mut bytes = fx.read(&profile_rel());
            bytes.push(b'\n');
            fx.write(&profile_rel(), bytes);
        },
        |fx| fx.write("AGENTS.md", GO_AGENTS),
    ];
    for change in changes {
        let fx = apply_project();
        let words = fx.command(&[]);
        change(&fx);
        let kind = fx.refuses(&words, &["stale or altered plan"]);
        assert_eq!(kind, "stale_plan");
    }
    // A digest that is not the preview's, and one the preview made for other picks.
    let fx = apply_project();
    let words = fx.command(&[]);
    let at = words.iter().position(|w| w == "--plan-digest").unwrap() + 1;
    let mut altered = words.clone();
    altered[at] = sha(b"zq");
    fx.refuses(&altered, &["stale or altered plan"]);
    let mut other = words;
    other.extend(["--keep".to_string(), "AGENTS.md:4=shared".to_string()]);
    fx.refuses(&other, &["stale or altered plan"]);
}

#[test]
fn refuse_an_unknown_id_a_duplicate_and_a_bad_destination() {
    let fx = apply_project();
    let words = fx.command(&["AGENTS.md:4=shared"]);
    for (value, why) in [
        ("AGENTS.md:5=shared", "no hunk"),
        ("AGENTS.md:4=personal", "twice"),
        ("CLAUDE.local.md:1=team", "destination"),
    ] {
        let mut words = words.clone();
        words.extend(["--keep".to_string(), value.to_string()]);
        fx.refuses(&words, &[why]);
    }
}

#[test]
fn refuse_a_shared_hunk_with_a_credential_shaped_line() {
    let mut fx = apply_project();
    let token = format!("ghp_{}", fx.secret("Q"));
    fx.write("CLAUDE.local.md", format!("# zq-local\nzq-key {token}\n"));
    let mut words = fx.command(&["CLAUDE.local.md:1=personal"]);
    let at = words.iter().position(|w| w == "--keep").unwrap() + 1;
    words[at] = "CLAUDE.local.md:1=shared".to_string();
    fx.blocks(
        &["CLAUDE.local.md:1=shared"],
        &words,
        &["CLAUDE.local.md:1", "credential"],
    );
}

#[test]
fn refuse_a_backup_folder_git_would_not_ignore() {
    let rule = backups_rule();
    let none = "AGENTS.local.md\nCLAUDE.local.md\n".to_string();
    // Ignored, and taken back by a later rule.
    let negated = format!("{none}{MOAI_DIR}/*\n!{rule}\n");
    // The receipt would be ignored and a moved file would not.
    let partly = format!("{none}{MOAI_DIR}/backups/*/*.json\n");
    for rules_text in [none, negated, partly] {
        let fx = apply_project();
        let words = fx.command(&[]);
        fx.write(".gitignore", &rules_text);
        fx.blocks(&[], &words, &["would not ignore", &rule]);
    }
    // Git cannot say.
    let fx = apply_project();
    let words = fx.command(&[]);
    fx.write(".git/config", "[zq-broken\n");
    fx.blocks(&[], &words, &["git could not say"]);
    // A repository marker git rejects is not "no repository".
    let fx = Fx::new();
    fx.profile(json!(null));
    fx.go_marker();
    let words = fx.command(&[]);
    fx.write(".git", "gitdir: zq-gone\n");
    fx.blocks(&[], &words, &["git could not say"]);
}

#[test]
fn refuse_a_link_or_a_file_where_the_backup_folder_goes() {
    for link in [true, false] {
        let fx = apply_project();
        let words = fx.command(&[]);
        let backups = format!("{MOAI_DIR}/backups");
        let outside = fx.base.join("outside");
        fs::create_dir_all(&outside).unwrap();
        if link {
            fx.link(&backups, &outside);
        } else {
            fx.write(&backups, "zq-file\n");
        }
        fx.blocks(&[], &words, &["backup folder", &backups]);
        assert_eq!(fs::read_dir(&outside).unwrap().count(), 0);
    }
}

#[test]
fn refuse_a_link_at_a_destination() {
    // The personal file's place is taken by a link after the preview.
    let fx = apply_project();
    fs::remove_file(fx.root.join("CLAUDE.local.md")).unwrap();
    let words = fx.command(&["AGENTS.local.md:1=personal"]);
    let outside = fx.base.join("outside.md");
    fx.link("CLAUDE.local.md", &outside);
    fx.blocks(
        &["AGENTS.local.md:1=personal"],
        &words,
        &["CLAUDE.local.md", "is a link"],
    );
    assert!(absent(&outside));
    // The new `CLAUDE.md`'s place.
    let fx = apply_project();
    let words = fx.command(&[]);
    let outside = fx.base.join("outside.md");
    fs::rename(fx.root.join("CLAUDE.md"), &outside).unwrap();
    fx.link("CLAUDE.md", &outside);
    fx.blocks(&[], &words, &["CLAUDE.md", "is a link"]);
    // The `.gitignore` that would gain the rule.
    let fx = apply_project();
    let words = fx.command(&["CLAUDE.local.md:1=personal"]);
    let outside = fx.base.join("outside-ignore");
    fs::write(&outside, "zq-outside\n").unwrap();
    fs::remove_file(fx.root.join(".gitignore")).unwrap();
    fx.link(".gitignore", &outside);
    fx.write(".git/info/exclude", format!("{}\n", backups_rule()));
    fx.blocks(
        &["CLAUDE.local.md:1=personal"],
        &words,
        &[".gitignore", "link"],
    );
    assert_eq!(fs::read(&outside).unwrap(), b"zq-outside\n");
}

#[test]
fn refuse_a_personal_pick_git_would_not_keep_out() {
    let keep = ["CLAUDE.local.md:1=personal"];
    // Git tracks the personal file.
    let fx = apply_project();
    let words = fx.command(&keep);
    fx.git(&["add", "-f", "--", "CLAUDE.local.md"]);
    fx.blocks(&keep, &words, &["CLAUDE.local.md", "tracks"]);
    // Without a personal pick nothing personal is written, so the plan goes on.
    assert_apply(&fx.plan());
    // A rule the user wrote takes the file back.
    let fx = apply_project();
    let words = fx.command(&keep);
    fx.write(
        ".gitignore",
        format!("{}\n!CLAUDE.local.md\n", backups_rule()),
    );
    fx.blocks(&keep, &words, &["CLAUDE.local.md", "un-ignores"]);
    assert_apply(&fx.plan());
}

#[test]
fn refuse_an_ignore_file_among_the_sources() {
    let fx = apply_project();
    let words = fx.command(&[]);
    fx.write(&rules("sub/.gitignore"), "!*\n");
    fx.blocks(&[], &words, &[&rules("sub/.gitignore")]);
}
