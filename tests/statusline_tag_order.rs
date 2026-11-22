//! SPEC-20261122-statusline-where (it replaces the tag-first rule of #2928): line 1 of the Claude
//! statusline has no tag. It opens with the model, and the usage windows `🔋 CW … | 🪫 7D …` are
//! the last line, the same bytes under every language. A session name changes nothing (see the
//! last cases of this file).
//!
//! The renderer has no language text, and these cases have no ledger, so no line has more than
//! one form: the width ladder only changes the flow line, and the other lines are the same bytes
//! at `COLUMNS` 80 and 40. "Korean and English, 80 and 40 columns" is therefore proven as
//! "byte-equal to the recorded golden": each case of `tests/fixtures/statusline-claude/cases.json` named below runs under
//! `LANG` ko_KR.UTF-8 / en_US.UTF-8 × `COLUMNS` 80 / 40 × plain (`NO_COLOR=1`) / truecolor
//! (`COLORTERM=truecolor`, through a pipe as Claude Code reads it), and must print exactly the
//! `plain` / `ansi` golden. The run mirrors `tools/statusline_golden.py`: a fresh HOME and
//! project per run, the case's setup steps, the same hidden seams, and the binary copied under
//! another name. Unix only, like the golden harness.
#![cfg(unix)]

mod common;
use common::CleanEnv;
use moai_cli::statusline::width::display_width;
use serde_json::Value;
use std::io::Write;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};

const FIXTURES: &str = concat!(
    env!("CARGO_MANIFEST_DIR"),
    "/tests/fixtures/statusline-claude"
);
/// `DEFAULT_ARGV` of the golden harness.
const ARGV: [&str; 5] = [
    "statusline",
    "--render-budget-ms",
    "3600000",
    "--now-unix",
    "1738429000",
];
/// `a02`: no registry (agent, every window); `d01`: a live leader entry; `d03`: a title with
/// escape bytes as `session_name` (agent); `d04`: the exact name `agent-27` (`Agent-27`, #2871).
const CASES: [&str; 4] = [
    "a02-doc-full",
    "d01-leader-mode",
    "d03-title-name",
    "d04-agent-name",
];

fn case(manifest: &Value, name: &str) -> Value {
    manifest["cases"]
        .as_array()
        .unwrap()
        .iter()
        .find(|c| c["name"] == name)
        .unwrap_or_else(|| panic!("case {name} in cases.json"))
        .clone()
}

/// The harness's `apply_setup` for the step kinds these cases use.
fn apply_setup(step: &Value, project: &Path) {
    match step["kind"].as_str().unwrap() {
        "git_init" => common::git::git_ok(
            project,
            &[
                "init",
                "-q",
                "-b",
                step["branch"].as_str().unwrap_or("main"),
            ],
        ),
        "write" => {
            assert!(step.get("under").is_none(), "{step}");
            let target = project.join(step["path"].as_str().unwrap());
            std::fs::create_dir_all(target.parent().unwrap()).unwrap();
            std::fs::write(target, step["content"].as_str().unwrap()).unwrap();
        }
        other => panic!("setup kind {other} is not used by these cases"),
    }
}

fn run_case(case: &Value, lang: &str, columns: &str, color: bool) -> String {
    let tmp = tempfile::tempdir().unwrap();
    let root = std::fs::canonicalize(tmp.path()).unwrap();
    let home = root.join("home");
    let moai_home = home.join(".moai");
    let project: PathBuf = root.join("work/demo-project");
    // The render keeps a per-session cache under `$TMPDIR`: a folder of this run, like the
    // harness gives it.
    let tmpdir = root.join("tmp");
    for d in [&home, &moai_home, &project, &tmpdir] {
        std::fs::create_dir_all(d).unwrap();
    }
    for step in case["setup"].as_array().into_iter().flatten() {
        apply_setup(step, &project);
    }
    let exe = root.join("bin/sl-under-test");
    std::fs::create_dir_all(exe.parent().unwrap()).unwrap();
    common::copy_executable(Path::new(common::MOAI_BIN), &exe);
    let stdin = std::fs::read(Path::new(FIXTURES).join(case["stdin"].as_str().unwrap())).unwrap();
    let stdin = String::from_utf8_lossy(&stdin)
        .replace("{PROJECT}", project.to_str().unwrap())
        .into_bytes();
    let mut cmd = Command::new(&exe);
    cmd.args(ARGV)
        .current_dir(&project)
        .clean_env()
        .env("PATH", "/usr/bin:/bin")
        .env("HOME", &home)
        .env("MOAI_HOME", &moai_home)
        .env("TZ", "UTC")
        .env("TMPDIR", &tmpdir)
        .env("LANG", lang)
        .env("COLUMNS", columns)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped());
    if color {
        cmd.env("COLORTERM", "truecolor");
    } else {
        cmd.env("NO_COLOR", "1");
    }
    let mut child = cmd.spawn().unwrap();
    child.stdin.take().unwrap().write_all(&stdin).unwrap();
    let out = child.wait_with_output().unwrap();
    assert!(
        out.status.success(),
        "exit {:?}: {}",
        out.status.code(),
        String::from_utf8_lossy(&out.stderr)
    );
    String::from_utf8(out.stdout).unwrap()
}

/// Removes SGR color sequences, so the line can be read as text.
fn strip_sgr(text: &str) -> String {
    let mut out = String::new();
    let mut rest = text;
    while let Some(at) = rest.find("\x1b[") {
        out.push_str(&rest[..at]);
        let tail = &rest[at..];
        rest = &tail[tail.find('m').unwrap() + 1..];
    }
    out.push_str(rest);
    out
}

/// Removes the OSC 8 hyperlink openers and closers (`ESC ] 8 ; ; <uri> ESC \`).
fn strip_links(text: &str) -> String {
    let mut out = String::new();
    let mut rest = text;
    while let Some(at) = rest.find("\x1b]8;") {
        out.push_str(&rest[..at]);
        let tail = &rest[at..];
        rest = &tail[tail.find("\x1b\\").map_or(tail.len(), |e| e + 2)..];
    }
    out.push_str(rest);
    out
}

#[test]
fn line1_has_no_tag_under_ko_en_and_80_40_columns() {
    let manifest: Value =
        serde_json::from_slice(&std::fs::read(Path::new(FIXTURES).join("cases.json")).unwrap())
            .unwrap();
    let mut runs = 0;
    for name in CASES {
        let case = case(&manifest, name);
        for (variant, color) in [("plain", false), ("ansi", true)] {
            let recorded = case["recorded_with"]["variants"][variant]["version"]
                .as_str()
                .unwrap_or_else(|| panic!("{name} {variant} recorded version"));
            let golden = std::fs::read_to_string(
                Path::new(FIXTURES).join(format!("golden/{name}.{variant}.txt")),
            )
            .unwrap()
            .replace(
                &format!("🗿 v{recorded}"),
                &format!("🗿 v{}", moai_cli::BUILD_VERSION),
            );
            for lang in ["ko_KR.UTF-8", "en_US.UTF-8"] {
                for columns in ["80", "40"] {
                    let got = run_case(&case, lang, columns, color);
                    // Line 1 has its own forms by width (SPEC-20261122-statusline-where REQ-11), so
                    // it is read below; every line after it is the same bytes as the golden.
                    assert_eq!(
                        got.lines().skip(1).collect::<Vec<_>>(),
                        golden.lines().skip(1).collect::<Vec<_>>(),
                        "{name} {variant} LANG={lang} COLUMNS={columns}"
                    );
                    let text = strip_sgr(&got);
                    let line1 = strip_links(text.lines().next().unwrap_or_default());
                    let line1 = line1.as_str();
                    let budget: usize = columns.parse::<usize>().unwrap() - 2;
                    assert!(
                        line1.starts_with("🧠 Opus"),
                        "{name} {variant} LANG={lang} COLUMNS={columns}: line 1 {line1:?}"
                    );
                    assert!(
                        display_width(line1) <= budget || !line1.contains(" | "),
                        "{name} {variant} LANG={lang} COLUMNS={columns}: line 1 {line1:?} is too wide"
                    );
                    assert!(
                        !text.contains("🏷️"),
                        "{name} {variant} LANG={lang} COLUMNS={columns}: a tag in {text:?}"
                    );
                    // The windows are the last line: counts as numbers, no graph, no colon.
                    let last = text.lines().last().unwrap_or_default();
                    assert!(
                        last.starts_with("🔋 CW ") || last.starts_with("🪫 CW "),
                        "{name} {variant} LANG={lang} COLUMNS={columns}: last line {last:?}"
                    );
                    runs += 1;
                }
            }
        }
    }
    assert_eq!(runs, CASES.len() * 2 * 2 * 2);
}

// ---- SPEC-20261122-statusline-where AC-5: no tag, and a session name changes nothing ----

const STATE_SOURCE: &str = include_str!("../src/statusline/state.rs");
const RENDER_SOURCE: &str = include_str!("../src/statusline/render.rs");

/// The registry entry the golden case `d01-leader-mode` uses: a live `leader`.
const LIVE_LEADER: &str = "{\"schema_version\":1,\"name\":\"leader\",\"instance\":\"11111111111111111111111111111111\",\"lease_until\":1738432600,\"registered_at\":1738429000,\"mode\":\"leader\"}\n";

/// The first line of one render in a fresh repository, with the stdin session name `name` (no
/// such field when `None`) and, when `registry` is set, a live `leader` entry in the lane
/// registry of the repository.
fn first_line_with(name: Option<&str>, registry: bool, color: bool) -> String {
    let tmp = tempfile::tempdir().unwrap();
    let root = std::fs::canonicalize(tmp.path()).unwrap();
    let home = root.join("home");
    let project = root.join("work/demo-project");
    let moai_home = home.join(moai_cli::home_root::MOAI_DIR);
    for d in [&home, &moai_home, &project, &root.join("tmp")] {
        std::fs::create_dir_all(d).unwrap();
    }
    common::git::git_ok(&project, &["init", "-q", "-b", "main"]);
    if registry {
        // The lane registry where the clone resolver puts it.
        let file = moai_cli::clone_root::path(&project.join(".git"))
            .join("sessions")
            .join("leader.json");
        std::fs::create_dir_all(file.parent().unwrap()).unwrap();
        std::fs::write(file, LIVE_LEADER).unwrap();
    }
    let mut stdin = serde_json::json!({
        "session_id": "fixture-where",
        "cwd": project,
        "model": {"display_name": "Opus"},
        "workspace": {"current_dir": project, "project_dir": project},
    });
    if let Some(name) = name {
        stdin["session_name"] = name.into();
    }
    let mut cmd = Command::new(common::MOAI_BIN);
    cmd.args(ARGV)
        .current_dir(&project)
        .clean_env()
        .env("PATH", "/usr/bin:/bin")
        .env("HOME", &home)
        .env("MOAI_HOME", &moai_home)
        .env("TMPDIR", root.join("tmp"))
        .env("TZ", "UTC")
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped());
    if color {
        cmd.env("COLORTERM", "truecolor");
    } else {
        cmd.env("NO_COLOR", "1");
    }
    let mut child = cmd.spawn().unwrap();
    child
        .stdin
        .take()
        .unwrap()
        .write_all(stdin.to_string().as_bytes())
        .unwrap();
    let out = child.wait_with_output().unwrap();
    assert!(
        out.status.success(),
        "exit {:?}: {}",
        out.status.code(),
        String::from_utf8_lossy(&out.stderr)
    );
    String::from_utf8(out.stdout)
        .unwrap()
        .lines()
        .next()
        .unwrap_or("")
        .to_string()
}

/// REQ-8: a session name of `leader`, `agent-3` or anything else changes no byte of line 1,
/// with or without a live `leader` entry in the lane registry, in plain text and in color; and
/// the first line carries no tag.
#[test]
fn the_session_name_changes_nothing() {
    let names: [Option<&str>; 8] = [
        None,
        Some(""),
        Some("leader"),
        Some("agent-3"),
        Some("agent-027"),
        Some("release"),
        Some("\u{1b}]0;a title\u{7}"),
        Some("LEADER"),
    ];
    for color in [false, true] {
        let baseline = first_line_with(None, false, color);
        assert!(!baseline.is_empty(), "a first line to compare");
        for name in names {
            for registry in [false, true] {
                assert_eq!(
                    first_line_with(name, registry, color),
                    baseline,
                    "name {name:?} registry {registry} color {color}"
                );
            }
        }
        for tag in ["🏷️", "Leader", "Agent-", "agent |"] {
            assert!(!baseline.contains(tag), "{tag} in {baseline:?}");
        }
    }
}

/// REQ-8: the status line reads no lane registry and no session name: its sources name neither,
/// and it keeps no tag text.
#[test]
fn the_session_name_changes_nothing_in_the_sources() {
    for (file, text) in [("state.rs", STATE_SOURCE), ("render.rs", RENDER_SOURCE)] {
        for gone in [
            "lane_registry",
            "session_name",
            "is_leader",
            "tag_text",
            "🏷️",
        ] {
            assert!(
                !text.contains(gone),
                "{file} still holds `{gone}`: the first line has no tag"
            );
        }
    }
}
