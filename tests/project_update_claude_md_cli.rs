//! SPEC-20261126-claude-md-fresh-merge AC-2 (REQ-3, REQ-4, D-2, D-4): `moai claude-md-plan`
//! lists a project's old instruction files in the fixed order with their facts, says which lines
//! of each are the project's own against the Go manifest and git history, refuses a source it
//! must not read, and writes nothing. Every case runs the real binary on a temporary project
//! with a temporary HOME; line text is never printed.

#![cfg(unix)]

mod common;
use common::{CleanEnv, MOAI_BIN};
use moai_cli::agents::{is_generated_local, AGENTS_SECTION};
use moai_cli::project_update::claude_md::{HISTORY_LIMIT, SOURCE_LIMIT};
use moai_cli::state_root::{GO_BACKUP, MOAI_DIR};
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

/// Every line of every fixture body starts with this, so one check proves no text is printed.
const TEXT: &str = "zq-";

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
        }
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
        let out = Command::new(MOAI_BIN)
            .args(args)
            .clean_env()
            .env("HOME", &self.home)
            .env("PATH", "/usr/bin:/bin")
            .env("LC_ALL", "C")
            .stdin(Stdio::null())
            .output()
            .expect("spawn the binary");
        Out {
            code: out.status.code(),
            stdout: String::from_utf8_lossy(&out.stdout).into_owned(),
            stderr: String::from_utf8_lossy(&out.stderr).into_owned(),
        }
    }

    /// `claude-md-plan`: it succeeds, prints no line text and leaves every byte as it was.
    fn plan(&self) -> Value {
        let before = self.snapshot();
        let out = self.run(&["claude-md-plan", self.root.to_str().unwrap()]);
        assert_eq!(self.snapshot(), before, "claude-md-plan wrote something");
        assert_eq!(
            out.code,
            Some(0),
            "claude-md-plan failed: stdout {} stderr {}",
            out.stdout,
            out.stderr
        );
        assert!(
            !out.stdout.contains(TEXT) && !out.stderr.contains(TEXT),
            "line text was printed: {}{}",
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
    // The fixture text itself has no `zq-` prefix, so the answer is read for its title instead.
    let result = fx.plan();
    assert!(!result.to_string().contains("Personal supplement"));
    for path in ["CLAUDE.local.md", "AGENTS.local.md"] {
        assert_lines(&result, path, json!({"used": "moai_first_text"}), &[]);
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
        !result.to_string().contains(&sha(b"zq-outside\n")),
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
        assert_eq!(
            refused(&result),
            vec![(rules_dir(), "lies below a link folder".to_string())],
            "{linked}"
        );
        assert_eq!(paths(&result), Vec::<String>::new(), "{linked}");
    }
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
