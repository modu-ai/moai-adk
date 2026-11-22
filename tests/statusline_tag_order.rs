//! #2928 (user instruction 6878415244, 2026-11-14, replacing S-1's tag-first line 3): line 1 of
//! the Claude statusline starts with `🏷️ <Leader|Agent-N|agent> | ` (#2871), and line 3 is
//! `🔋 CW … | 🪫 7D …` without the tag, the same bytes under every language. The worktree label
//! (`🌳 <name>`, or `🌳 WT-c<card>-<name>`) closes line 2 when there is one, and with a ledger a
//! flow line stands between line 2 and the windows (SPEC-20261122-statusline-flow); these cases
//! have no ledger, so the check below still finds CW on line 3.
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
    for d in [&home, &moai_home, &project] {
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

#[test]
fn line1_tag_first_under_ko_en_and_80_40_columns() {
    let manifest: Value =
        serde_json::from_slice(&std::fs::read(Path::new(FIXTURES).join("cases.json")).unwrap())
            .unwrap();
    let mut runs = 0;
    for name in CASES {
        let case = case(&manifest, name);
        let want_tag = match name {
            "d01-leader-mode" => "🏷️ Leader | ",
            "d04-agent-name" => "🏷️ Agent-27 | ",
            _ => "🏷️ agent | ",
        };
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
                    assert_eq!(
                        got, golden,
                        "{name} {variant} LANG={lang} COLUMNS={columns}"
                    );
                    let text = strip_sgr(&got);
                    let line1 = text.lines().next().unwrap_or_default();
                    assert!(
                        line1.starts_with(want_tag),
                        "{name} {variant} LANG={lang} COLUMNS={columns}: line 1 {line1:?}"
                    );
                    let line3 = text.lines().nth(2).unwrap_or_default();
                    // The worktree label closed line 2 since the picked layout
                    // (SPEC-20261121-statusline-detached REQ-4); line 3 holds the counts and
                    // the windows as numbers (REQ-5: no graph, no colon).
                    let line3 = match line3.strip_prefix("🌳 ") {
                        Some(rest) => rest.split_once(" | ").map_or(line3, |(_, after)| after),
                        None => line3,
                    };
                    assert!(
                        line3.starts_with("🔋 CW ") || line3.starts_with("🪫 CW "),
                        "{name} {variant} LANG={lang} COLUMNS={columns}: line 3 {line3:?}"
                    );
                    assert!(
                        !text.lines().skip(1).any(|l| l.contains("🏷️")),
                        "{name} {variant} LANG={lang} COLUMNS={columns}: tag after line 1"
                    );
                    runs += 1;
                }
            }
        }
    }
    assert_eq!(runs, CASES.len() * 2 * 2 * 2);
}
