//! #2376: the Claude three-line contract runs inside `cargo test`, comparing real native
//! output with the Rust goldens in `tests/fixtures/statusline-claude/`.
//!
//! The harness lives in this file (it was `tools/statusline_golden.py` until
//! SPEC-20261125-rust-only-finish): a fresh HOME and project per case, the binary copied under
//! another name, an environment built from scratch (`CleanEnv`), and on macOS the
//! network/HOME/exec sandbox. Unix only: the fixtures use POSIX paths and `/usr/bin` git;
//! Windows CI compiles this file without running it.
//!
//! Wall-clock independence (#2350 leader decision): the product's 800 ms render budget drops
//! the git segment when git is slow, which a loaded machine makes happen at random (18 of 20
//! runs under a concurrent cargo build, re-audit 6765987319). The goldens describe the render
//! within budget, so the harness passes the hidden seam `--render-budget-ms 3600000`; the
//! over-budget path is pinned on its own in `tests/statusline_git_env.rs`.
//!
//! Clock independence (#2756 AC-8): the prompt cache timer counts from `--now-unix`, a hidden
//! seam; the goldens are recorded and checked with the same value (`ARGV`).
//!
//! Variants: `plain` pipes stdout with `NO_COLOR=1`; `ansi` and `ansi256` attach stdout to a
//! real pseudo terminal with and without `COLORTERM=truecolor`. Claude Code reads the
//! statusline through a pipe, so a pipe without `NO_COLOR` must carry the same bytes as the
//! terminal (escape-free output there was the #2559 regression).
//!
//! Recording the goldens again: run the ignored test `record_statusline_goldens`
//! (`cargo test --test statusline_golden -- --ignored record_statusline_goldens`, or
//! `cargo nextest run --test statusline_golden --run-ignored only`). `MOAI_GOLDEN_CASES` and
//! `MOAI_GOLDEN_VARIANTS` (comma lists) restrict it; read the diff of every changed golden.
#![cfg(unix)]

mod common;
use common::pty::Terminal;
use common::CleanEnv;
use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::Command;
use std::sync::OnceLock;
use std::time::Duration;

const FIXTURES: &str = concat!(
    env!("CARGO_MANIFEST_DIR"),
    "/tests/fixtures/statusline-claude"
);
const VARIANTS: [&str; 3] = ["plain", "ansi", "ansi256"];
const VERSION_PLACEHOLDER: &str = "<binary-version>";
const RECORDING: &str =
    "Actual native binary outputs; provenance is recorded per case and variant.";
/// Recording and checking share these hidden seams: `--render-budget-ms` (a slow git cannot
/// drop the git segment, #2350) and `--now-unix` (the prompt cache timer counts from a fixed
/// clock, #2756 AC-8).
const ARGV: [&str; 5] = [
    "statusline",
    "--render-budget-ms",
    "3600000",
    "--now-unix",
    "1738429000",
];
const GIT_ENV: [(&str, &str); 8] = [
    ("GIT_AUTHOR_NAME", "fixture"),
    ("GIT_AUTHOR_EMAIL", "fixture@example.invalid"),
    ("GIT_COMMITTER_NAME", "fixture"),
    ("GIT_COMMITTER_EMAIL", "fixture@example.invalid"),
    ("GIT_AUTHOR_DATE", "1735689600 +0000"),
    ("GIT_COMMITTER_DATE", "1735689600 +0000"),
    ("GIT_CONFIG_GLOBAL", "/dev/null"),
    ("GIT_CONFIG_NOSYSTEM", "1"),
];
const RUN_LIMIT: Duration = Duration::from_secs(30);

fn fixtures() -> &'static Path {
    Path::new(FIXTURES)
}

fn variant_env(variant: &str) -> &'static [(&'static str, &'static str)] {
    match variant {
        "plain" => &[("NO_COLOR", "1")],
        "ansi" => &[("COLORTERM", "truecolor")],
        "ansi256" => &[],
        other => panic!("unknown variant {other:?}"),
    }
}

fn manifest_path() -> PathBuf {
    fixtures().join("cases.json")
}

fn load_manifest() -> Value {
    let text = fs::read_to_string(manifest_path()).expect("cases.json is readable");
    serde_json::from_str(&text).expect("cases.json is JSON")
}

/// The manifest as the recorder writes it: two-space indent, non-ASCII as `\uXXXX` escapes
/// (UTF-16 units, lowercase hex) and one final newline, so a rewrite changes only what changed.
fn manifest_text(manifest: &Value) -> String {
    let text = serde_json::to_string_pretty(manifest).expect("the manifest serializes");
    let mut out = String::with_capacity(text.len() + 1);
    for c in text.chars() {
        if c.is_ascii() {
            out.push(c);
        } else {
            let mut units = [0u16; 2];
            for unit in c.encode_utf16(&mut units) {
                out.push_str(&format!("\\u{unit:04x}"));
            }
        }
    }
    out.push('\n');
    out
}

fn path_var() -> String {
    std::env::var("PATH").unwrap_or_else(|_| "/usr/bin:/bin".to_string())
}

/// One `git` call in the temporary project, with the fixed author, dates and no config files.
fn run_git(cwd: &Path, args: &[&str]) -> Result<(), String> {
    let mut cmd = Command::new("git");
    cmd.clean_env()
        .arg("-C")
        .arg(cwd)
        .args(args)
        .envs(GIT_ENV)
        .env("PATH", path_var());
    let out = common::pty::output_within(&mut cmd, None);
    if out.status.success() {
        Ok(())
    } else {
        Err(format!(
            "git {args:?} exit {:?}: {}",
            out.status.code(),
            String::from_utf8_lossy(&out.stderr).trim()
        ))
    }
}

/// One setup step from cases.json. Paths are relative to the project unless the step says
/// otherwise (`moai_home`: the Go MOAI_HOME the goldens were recorded with; `cli_home`: the
/// home state folder moai-cli reads, #2455); every write lands inside the temporary root. A
/// step that is refused creates nothing.
fn apply_setup(step: &Value, root: &Path, project: &Path, moai_home: &Path) -> Result<(), String> {
    let text = |key: &str| step[key].as_str();
    let kind = text("kind").ok_or("a setup step without a kind")?;
    match kind {
        "write" => {
            let base = match text("under") {
                Some("moai_home") => moai_home.to_path_buf(),
                Some("cli_home") => moai_home
                    .parent()
                    .ok_or("moai_home has no parent")?
                    .join(moai_cli::HOME_STATE_DIR),
                None => project.to_path_buf(),
                Some(other) => return Err(format!("unknown write base {other:?}")),
            };
            let target = base.join(text("path").ok_or("a write step without a path")?);
            let parent = target.parent().ok_or("a write target without a parent")?;
            fs::create_dir_all(parent).map_err(|e| e.to_string())?;
            fs::write(
                &target,
                text("content").ok_or("a write step without content")?,
            )
            .map_err(|e| e.to_string())
        }
        "copy" => {
            // Card C-65: a checked-in binary fixture (a project ledger), copied byte for byte
            // from the fixture folder into the temporary HOME's `.moai`.
            if text("under") != Some("moai_home") {
                return Err(format!("unknown copy base {:?}", text("under")));
            }
            let from = text("from").ok_or("a copy step without a source")?;
            let outside = || format!("copy source outside the fixtures: {from:?}");
            let source = fixtures()
                .join(from)
                .canonicalize()
                .map_err(|_| outside())?;
            let folder = fixtures().canonicalize().map_err(|e| e.to_string())?;
            if !source.starts_with(&folder) {
                return Err(outside());
            }
            if !source.is_file() {
                return Err(format!("copy source is not a file: {from:?}"));
            }
            let target = moai_home.join(text("path").ok_or("a copy step without a path")?);
            let parent = target.parent().ok_or("a copy target without a parent")?;
            fs::create_dir_all(parent).map_err(|e| e.to_string())?;
            fs::copy(&source, &target)
                .map(|_| ())
                .map_err(|e| e.to_string())
        }
        "git_init" => run_git(
            project,
            &["init", "-q", "-b", text("branch").unwrap_or("main")],
        ),
        "git_commit_all" => {
            run_git(project, &["add", "-A"])?;
            run_git(
                project,
                &[
                    "commit",
                    "-q",
                    "--allow-empty",
                    "-m",
                    text("message").unwrap_or("fixture"),
                ],
            )
        }
        "git_upstream" => {
            let bare = root.join("origin.git");
            let bare_text = bare.to_string_lossy().into_owned();
            let mut cmd = Command::new("git");
            cmd.clean_env()
                .args(["init", "-q", "--bare", &bare_text])
                .envs(GIT_ENV)
                .env("PATH", path_var());
            let out = common::pty::output_within(&mut cmd, None);
            if !out.status.success() {
                return Err(format!("git init --bare failed: {:?}", out.status.code()));
            }
            run_git(project, &["remote", "add", "origin", &bare_text])?;
            run_git(
                project,
                &[
                    "push",
                    "-q",
                    "-u",
                    "origin",
                    text("branch").unwrap_or("main"),
                ],
            )
        }
        other => Err(format!("unknown setup kind {other:?}")),
    }
}

/// Sandbox clauses for what a `git` call may exec on macOS: the /usr/bin shim, xcrun, and the
/// developer-tools bin directory the shim resolves git into (git runs more than the one `git`
/// file from there; measured on #2350).
#[cfg(target_os = "macos")]
fn git_exec_clauses() -> Vec<String> {
    let mut clauses = vec![
        r#"(literal "/usr/bin/git")"#.to_string(),
        r#"(literal "/usr/bin/xcrun")"#.to_string(),
    ];
    let found = Command::new("/usr/bin/xcrun")
        .args(["--find", "git"])
        .output();
    if let Ok(found) = found {
        let path = String::from_utf8_lossy(&found.stdout).trim().to_string();
        if found.status.success() && !path.is_empty() {
            if let Some(dev_bin) = fs::canonicalize(&path)
                .ok()
                .as_deref()
                .and_then(Path::parent)
            {
                clauses.push(format!(r#"(subpath "{}")"#, dev_bin.display()));
            }
        }
    }
    clauses
}

/// Deny all network, every read or write under the real HOME, and every exec except the
/// binary under test and git.
#[cfg(target_os = "macos")]
fn sandbox_profile(exe: &Path) -> String {
    let home = std::env::var_os("HOME")
        .map(PathBuf::from)
        .unwrap_or_default();
    let real_home = fs::canonicalize(&home).unwrap_or(home);
    let mut allowed = vec![format!(r#"(literal "{}")"#, exe.display())];
    allowed.extend(git_exec_clauses());
    format!(
        r#"(version 1)(allow default)(deny network*)(deny file-read* (subpath "{home}"))(deny file-write* (subpath "{home}"))(deny process-exec*)(allow process-exec* {allowed})"#,
        home = real_home.display(),
        allowed = allowed.join(" ")
    )
}

/// The program and arguments that run `exe args`: through the sandbox on macOS, as they are
/// elsewhere.
fn wrapped(exe: &Path, args: &[&str]) -> (PathBuf, Vec<String>) {
    let args: Vec<String> = args.iter().map(|a| a.to_string()).collect();
    #[cfg(target_os = "macos")]
    {
        let mut all = vec![
            "-p".to_string(),
            sandbox_profile(exe),
            exe.to_string_lossy().into_owned(),
        ];
        all.extend(args);
        (PathBuf::from("/usr/bin/sandbox-exec"), all)
    }
    #[cfg(not(target_os = "macos"))]
    {
        (exe.to_path_buf(), args)
    }
}

fn normalize_version(text: &str, version: &str) -> String {
    let mut text = text.to_string();
    for mark in ["🗿 v", "🗿 ", "-> "] {
        text = text.replace(
            &format!("{mark}{version}"),
            &format!("{mark}{VERSION_PLACEHOLDER}"),
        );
    }
    text
}

/// A version token, with an optional leading `v` taken off: three dot-separated digit runs,
/// then only letters, digits and `.+-` (`2026.11.0-rc.1+local.3`).
fn version_token(token: &str) -> Option<String> {
    let token = token.strip_prefix('v').unwrap_or(token);
    let bytes = token.as_bytes();
    let mut at = 0;
    for part in 0..3 {
        let start = at;
        while bytes.get(at).is_some_and(u8::is_ascii_digit) {
            at += 1;
        }
        if at == start {
            return None;
        }
        if part < 2 {
            if bytes.get(at) != Some(&b'.') {
                return None;
            }
            at += 1;
        }
    }
    let tail_ok = bytes[at..]
        .iter()
        .all(|c| c.is_ascii_alphanumeric() || matches!(c, b'.' | b'+' | b'-'));
    tail_ok.then(|| token.to_string())
}

/// The version in a `--version` run's first line (`moai-cli 2026.11.0-rc.1 (rust)`,
/// `moai version v3.2.1`): its last token, or the token before a trailing generation mark such
/// as `(rust)` (#2790), without a leading `v`; `None` for a failed run or no version.
fn parse_version_output(code: Option<i32>, stdout: &str) -> Option<String> {
    let mut tokens: Vec<&str> = stdout.lines().next()?.split_whitespace().collect();
    if code != Some(0) || tokens.is_empty() {
        return None;
    }
    if tokens.len() > 1 {
        let last = tokens[tokens.len() - 1];
        if last.starts_with('(') && last.ends_with(')') {
            tokens.pop();
        }
    }
    version_token(tokens[tokens.len() - 1])
}

/// `<binary> --version`, run the way the cases run it: a copy under another name, a throwaway
/// HOME, a minimal PATH and, on macOS, the same sandbox.
fn detect_binary_version(binary: &Path) -> Option<String> {
    let tmp = tempfile::tempdir().ok()?;
    let root = tmp.path().canonicalize().ok()?;
    let exe = root.join("bin").join("sl-under-test");
    fs::create_dir_all(exe.parent()?).ok()?;
    fs::create_dir_all(root.join("home")).ok()?;
    fs::copy(binary, &exe).ok()?;
    let (program, args) = wrapped(&exe, &["--version"]);
    let mut cmd = Command::new(program);
    cmd.clean_env()
        .args(args)
        .current_dir(&root)
        .env("PATH", "/usr/bin:/bin")
        .env("HOME", root.join("home"));
    let out = cmd.output().ok()?;
    parse_version_output(out.status.code(), &String::from_utf8_lossy(&out.stdout))
}

fn replace_bytes(haystack: &[u8], needle: &[u8], with: &[u8]) -> Vec<u8> {
    let mut out = Vec::with_capacity(haystack.len());
    let mut at = 0;
    while at < haystack.len() {
        if haystack[at..].starts_with(needle) {
            out.extend_from_slice(with);
            at += needle.len();
        } else {
            out.push(haystack[at]);
            at += 1;
        }
    }
    out
}

struct Ran {
    code: Option<i32>,
    stdout: String,
    stderr: String,
}

/// One run of the binary on a case's stdin in its own temporary HOME, project and `TMPDIR`.
/// `stdout_tty` overrides the variant's choice (`plain` is a pipe, the colored ones a terminal).
fn run_case(
    binary: &Path,
    case: &Value,
    variant: &str,
    stdin_bytes: Option<&[u8]>,
    stdout_tty: Option<bool>,
) -> Result<Ran, String> {
    let tmp = tempfile::Builder::new()
        .prefix("sl-golden-")
        .tempdir()
        .map_err(|e| e.to_string())?;
    let root = tmp.path().canonicalize().map_err(|e| e.to_string())?;
    let home = root.join("home");
    let moai_home = home.join(".moai");
    let project = root.join("work").join("demo-project");
    // The render keeps a per-session cache under `$TMPDIR` (SPEC-20261122-statusline-where): a
    // folder of its own per run, so no case reads another case's cache and nothing reaches the
    // real temporary folder.
    let tmpdir = root.join("tmp");
    for dir in [&home, &moai_home, &project, &tmpdir] {
        fs::create_dir_all(dir).map_err(|e| e.to_string())?;
    }
    for step in case["setup"].as_array().into_iter().flatten() {
        apply_setup(step, &root, &project, &moai_home)?;
    }
    let exe = root.join("bin").join("sl-under-test");
    fs::create_dir_all(exe.parent().unwrap()).map_err(|e| e.to_string())?;
    fs::copy(binary, &exe).map_err(|e| e.to_string())?;
    // Bytes, not text: a fixture may hold invalid UTF-8 on purpose (#2350 audit F1).
    let raw = match stdin_bytes {
        Some(bytes) => bytes.to_vec(),
        None => {
            let name = case["stdin"].as_str().ok_or("a case without stdin")?;
            fs::read(fixtures().join(name)).map_err(|e| e.to_string())?
        }
    };
    let stdin = replace_bytes(&raw, b"{PROJECT}", project.to_string_lossy().as_bytes());
    let (program, args) = wrapped(&exe, &ARGV);
    let tty = stdout_tty.unwrap_or(variant != "plain");
    let mut terminal = Terminal::new(program)
        .args(args)
        .cwd(&project)
        .stdin_text(stdin)
        .stderr_not_a_terminal()
        .limit(RUN_LIMIT)
        .env("PATH", "/usr/bin:/bin")
        .env("HOME", &home)
        .env("MOAI_HOME", &moai_home)
        .env("TZ", "UTC")
        .env("LANG", "C.UTF-8")
        .env("TMPDIR", &tmpdir);
    for (key, value) in variant_env(variant) {
        terminal = terminal.env(key, value);
    }
    if !tty {
        terminal = terminal.stdout_not_a_terminal();
    }
    let screen = terminal.run();
    Ok(Ran {
        code: screen.code,
        stdout: if tty { screen.text } else { screen.stdout },
        stderr: screen.stderr,
    })
}

/// The binary under test and its version, read once.
fn binary() -> &'static Path {
    Path::new(common::MOAI_BIN)
}

fn binary_version() -> &'static str {
    static VERSION: OnceLock<String> = OnceLock::new();
    VERSION.get_or_init(|| {
        detect_binary_version(binary())
            .expect("`--version` printed no version for the binary under test")
    })
}

/// Whether a recorded golden and a fresh output are the same text once each side's own build
/// version is replaced by the placeholder.
fn golden_matches(golden: &str, recorded_version: &str, out: &str, current_version: &str) -> bool {
    normalize_version(golden, recorded_version) == normalize_version(out, current_version)
}

/// The output of one case and variant, after the checks that hold for both recording and
/// checking: a clean exit with text, no escapes under `NO_COLOR`, and a pipe that carries the
/// bytes the terminal showed (#2559).
fn render(binary: &Path, case: &Value, variant: &str) -> Result<String, String> {
    let ran = run_case(binary, case, variant, None, None)?;
    if ran.code != Some(0) || ran.stdout.trim().is_empty() {
        let err: String = ran.stderr.trim().chars().take(200).collect();
        return Err(format!(
            "exit {:?}, stdout {:?}, stderr {err}",
            ran.code, ran.stdout
        ));
    }
    if variant == "plain" && ran.stdout.contains('\x1b') {
        return Err("NO_COLOR output contains terminal escapes".to_string());
    }
    if variant != "plain" {
        // Claude Code reads stdout through a pipe; the piped bytes must equal the terminal's.
        let pipe = run_case(binary, case, variant, None, Some(false))?;
        if pipe.code != Some(0) || pipe.stdout != ran.stdout {
            return Err(format!(
                "pipe output differs from the terminal: {:?} {:?} {:?}",
                pipe.code, pipe.stdout, pipe.stderr
            ));
        }
    }
    Ok(ran.stdout)
}

fn golden_path(case: &Value, variant: &str) -> PathBuf {
    let name = case["name"].as_str().expect("a case has a name");
    fixtures()
        .join("golden")
        .join(format!("{name}.{variant}.txt"))
}

fn check_one(case: &Value, variant: &str) -> Result<(), String> {
    let out = render(binary(), case, variant)?;
    let recorded = case["recorded_with"]["variants"][variant]["version"]
        .as_str()
        .ok_or_else(|| format!("no recorded version for {variant}"))?;
    let golden = fs::read_to_string(golden_path(case, variant)).map_err(|e| e.to_string())?;
    if golden_matches(&golden, recorded, &out, binary_version()) {
        Ok(())
    } else {
        Err("golden differs".to_string())
    }
}

#[test]
fn statusline_matches_every_claude_golden() {
    let manifest = load_manifest();
    let mut failures = Vec::new();
    let mut checked = 0;
    for case in manifest["cases"].as_array().expect("cases is a list") {
        for variant in VARIANTS {
            match check_one(case, variant) {
                Ok(()) => checked += 1,
                Err(why) => failures.push(format!("FAIL {} {variant}: {why}", case["name"])),
            }
        }
    }
    assert!(checked > 0, "no golden outputs were checked");
    assert!(
        failures.is_empty(),
        "{} of {} golden outputs failed:\n{}",
        failures.len(),
        checked + failures.len(),
        failures.join("\n")
    );
}

/// Removes every SGR color sequence (`ESC [ digits and ; m`).
fn strip_sgr(text: &str) -> String {
    let bytes = text.as_bytes();
    let mut out = Vec::with_capacity(bytes.len());
    let mut at = 0;
    while at < bytes.len() {
        if bytes[at] == 0x1b && bytes.get(at + 1) == Some(&b'[') {
            let mut end = at + 2;
            while bytes
                .get(end)
                .is_some_and(|c| c.is_ascii_digit() || *c == b';')
            {
                end += 1;
            }
            if bytes.get(end) == Some(&b'm') {
                at = end + 1;
                continue;
            }
        }
        out.push(bytes[at]);
        at += 1;
    }
    String::from_utf8_lossy(&out).into_owned()
}

/// Removes every OSC 8 hyperlink opener or closer: `ESC ] 8 ; params ; URI ESC \`.
fn strip_osc8(text: &str) -> String {
    let bytes = text.as_bytes();
    let mut out = Vec::with_capacity(bytes.len());
    let mut at = 0;
    while at < bytes.len() {
        if bytes[at..].starts_with(b"\x1b]8;") {
            let mut end = at + 4;
            while bytes.get(end).is_some_and(|c| *c != b';' && *c != 0x1b) {
                end += 1;
            }
            if bytes.get(end) == Some(&b';') {
                end += 1;
                while bytes.get(end).is_some_and(|c| *c != 0x1b) {
                    end += 1;
                }
                if bytes.get(end) == Some(&0x1b) && bytes.get(end + 1) == Some(&b'\\') {
                    at = end + 2;
                    continue;
                }
            }
        }
        out.push(bytes[at]);
        at += 1;
    }
    String::from_utf8_lossy(&out).into_owned()
}

/// Color on a pipe (how Claude Code runs it), `NO_COLOR` on a terminal, and both palettes.
/// #2559: a pipe without `NO_COLOR` must be colored; an escape-free pipe fails here.
#[test]
fn the_color_contract_holds_on_a_pipe_and_on_a_terminal() {
    let case = json!({"stdin": "../statusline/stdin/a02-doc-full.json"});
    let plain = run_case(binary(), &case, "plain", None, None).unwrap();
    assert!(
        plain.code == Some(0) && !plain.stdout.trim().is_empty() && !plain.stdout.contains('\x1b'),
        "plain pipe failed: {:?} {:?} {:?}",
        plain.code,
        plain.stdout,
        plain.stderr
    );
    for (label, variant, tty) in [
        ("pipe without NO_COLOR", "ansi256", Some(false)),
        ("truecolor pipe", "ansi", Some(false)),
        ("TTY with NO_COLOR", "plain", Some(true)),
        ("truecolor TTY", "ansi", None),
        ("ANSI256 TTY", "ansi256", None),
    ] {
        let ran = run_case(binary(), &case, variant, None, tty).unwrap();
        // SGR colors and (#2756 AC-7) OSC 8 hyperlinks are the only escapes color adds.
        assert!(
            ran.code == Some(0) && strip_osc8(&strip_sgr(&ran.stdout)) == plain.stdout,
            "{label} content differs: {:?} {:?} {:?}",
            ran.code,
            ran.stdout,
            ran.stderr
        );
        if variant == "plain" {
            assert!(
                !ran.stdout.contains('\x1b'),
                "{label} contains terminal escapes"
            );
        }
        if variant == "ansi" {
            assert!(
                ran.stdout.contains("38;2;"),
                "truecolor TTY has no truecolor escapes"
            );
        }
        if variant == "ansi256" {
            assert!(
                ran.stdout.contains("38;5;") && !ran.stdout.contains("38;2;"),
                "ANSI256 TTY does not use the ANSI256 palette"
            );
        }
    }
}

/// Exercises the native stdin cap, including a valid JSON prefix beyond the cap.
#[test]
fn the_stdin_cap_and_malformed_input_fall_back_to_the_empty_input_line() {
    let case = json!({"stdin": "../statusline/stdin/a01-minimal.json"});
    let run = |bytes: &[u8]| run_case(binary(), &case, "plain", Some(bytes), None).unwrap();
    let fallback = run(b"");
    assert!(
        fallback.code == Some(0)
            && !fallback.stdout.trim().is_empty()
            && (fallback.stdout.contains("MoAI") || fallback.stdout.contains('🗿')),
        "empty-input fallback failed: {:?} {:?} {:?}",
        fallback.code,
        fallback.stdout,
        fallback.stderr
    );
    let cap = 4 * 1024 * 1024;
    let valid = br#"{"model":{"display_name":"BoundaryModel"}}"#;
    let mut exact = valid.to_vec();
    exact.resize(cap, b' ');
    let ran = run(&exact);
    assert!(
        ran.code == Some(0) && ran.stdout.contains("BoundaryModel") && !ran.stdout.contains('\x1b'),
        "exact-cap valid input failed: {:?} {:?} {:?}",
        ran.code,
        ran.stdout,
        ran.stderr
    );
    let mut oversized = br#"{"session_name":""#.to_vec();
    oversized.extend(std::iter::repeat_n(b'x', cap));
    let mut over_cap = exact.clone();
    over_cap.push(b' ');
    for (label, raw) in [
        ("malformed", b"{".to_vec()),
        (
            "invalid typed field",
            br#"{"model":{"display_name":42}}"#.to_vec(),
        ),
        ("oversized JSON string", oversized),
        ("valid prefix over cap", over_cap),
    ] {
        let ran = run(&raw);
        assert!(
            ran.code == Some(0) && ran.stdout == fallback.stdout && !ran.stdout.contains('\x1b'),
            "{label} did not fall back: {:?} {:?} {:?}",
            ran.code,
            ran.stdout,
            ran.stderr
        );
    }
}

/// Re-records the goldens and their provenance from the binary this build produced. Ignored: it
/// rewrites checked-in files. `MOAI_GOLDEN_CASES` and `MOAI_GOLDEN_VARIANTS` restrict it.
#[test]
#[ignore = "rewrites the checked-in goldens; run it by name"]
fn record_statusline_goldens() {
    let list = |key: &str| -> Option<Vec<String>> {
        std::env::var(key)
            .ok()
            .map(|v| v.split(',').map(|s| s.trim().to_string()).collect())
    };
    let only_cases = list("MOAI_GOLDEN_CASES");
    let only_variants = list("MOAI_GOLDEN_VARIANTS");
    let mut manifest = load_manifest();
    let digest = format!(
        "{:x}",
        Sha256::digest(fs::read(binary()).expect("the binary reads"))
    );
    let version = binary_version();
    let known: Vec<String> = manifest["cases"]
        .as_array()
        .unwrap()
        .iter()
        .map(|c| c["name"].as_str().unwrap().to_string())
        .collect();
    for name in only_cases.iter().flatten() {
        assert!(known.contains(name), "unknown case: {name}");
    }
    for name in only_variants.iter().flatten() {
        assert!(VARIANTS.contains(&name.as_str()), "unknown variant: {name}");
    }
    let mut written = 0;
    for case in manifest["cases"].as_array_mut().unwrap() {
        let name = case["name"].as_str().unwrap().to_string();
        if only_cases.as_ref().is_some_and(|l| !l.contains(&name)) {
            continue;
        }
        for variant in VARIANTS {
            if only_variants
                .as_ref()
                .is_some_and(|l| !l.iter().any(|v| v == variant))
            {
                continue;
            }
            let out = render(binary(), case, variant)
                .unwrap_or_else(|why| panic!("{name} {variant}: {why}"));
            fs::create_dir_all(fixtures().join("golden")).unwrap();
            fs::write(golden_path(case, variant), out).unwrap();
            case["recorded_with"]["implementation"] = json!("Rust moai-cli");
            case["recorded_with"]["variants"][variant] = json!({
                "version": version,
                "sha256": digest,
                "stdout": if variant == "plain" { "pipe" } else { "PTY" },
            });
            written += 1;
        }
    }
    assert!(written > 0, "no selected cases or variants");
    manifest["recording"] = json!(RECORDING);
    fs::write(manifest_path(), manifest_text(&manifest)).expect("cases.json is writable");
    eprintln!("recorded {written} golden outputs");
}

// ---- the harness itself (what tests/statusline_fixtures_test.py pinned) ----

/// #2350: `--version` is read the way the cases run the binary, and only the version token
/// after the marks is normalized.
#[test]
fn the_last_version_token_is_taken_without_a_leading_v() {
    let ok = |line: &str| parse_version_output(Some(0), line);
    assert_eq!(
        ok("moai-cli 2026.11.0-rc.1\n").as_deref(),
        Some("2026.11.0-rc.1")
    );
    assert_eq!(ok("moai version v3.2.1\n").as_deref(), Some("3.2.1"));
    // #2790: the Rust build's line ends with its generation mark.
    assert_eq!(
        ok("moai-cli 2026.11.0-rc.1 (rust)\n").as_deref(),
        Some("2026.11.0-rc.1")
    );
    assert_eq!(
        ok("moai-cli 2026.11.0-rc.1+local.3 (rust)\n").as_deref(),
        Some("2026.11.0-rc.1+local.3")
    );
    assert_eq!(ok("(rust)\n"), None);
}

#[test]
fn no_version_or_a_failed_run_is_none() {
    assert_eq!(parse_version_output(Some(0), "moai-cli unknown\n"), None);
    assert_eq!(parse_version_output(Some(0), ""), None);
    assert_eq!(parse_version_output(Some(1), "moai-cli 1.2.3\n"), None);
    assert_eq!(parse_version_output(None, "moai-cli 1.2.3\n"), None);
}

#[test]
fn only_the_named_version_token_is_replaced() {
    let norm = normalize_version;
    assert_eq!(
        norm("🧠 Opus | 🗿 v3.2.1\n", "3.2.1"),
        "🧠 Opus | 🗿 v<binary-version>\n"
    );
    // b04: the template version stays visible; only the binary side is normalized.
    assert_eq!(
        norm("🗿 v2.0.0 -> 🗿 v3.2.1", "3.2.1"),
        "🗿 v2.0.0 -> 🗿 v<binary-version>"
    );
    assert_eq!(norm("🗿 v3.2.10", "3.2.1"), "🗿 v<binary-version>0");
    assert_eq!(norm("v3.2.1", "3.2.1"), "v3.2.1");
}

/// A binary that prints no version is not a golden run: the version is unknown, not guessed.
#[test]
fn a_file_that_is_no_program_has_no_readable_version() {
    let tmp = tempfile::tempdir().unwrap();
    let not_a_binary = tmp.path().join("not-a-binary");
    fs::write(&not_a_binary, "plain text\n").unwrap();
    assert_eq!(detect_binary_version(&not_a_binary), None);
}

/// The guard of the comparison: a changed line is reported, a changed build version alone is
/// not, and a template version that is not the build's stays visible.
#[test]
fn a_changed_golden_is_told_apart_from_a_changed_build_version() {
    let golden = "🧠 Opus | 🗿 v1.0.0\nline two\n";
    assert!(golden_matches(
        golden,
        "1.0.0",
        "🧠 Opus | 🗿 v2.0.0\nline two\n",
        "2.0.0"
    ));
    assert!(!golden_matches(
        golden,
        "1.0.0",
        "🧠 Opus | 🗿 v2.0.0\nline 2\n",
        "2.0.0"
    ));
    assert!(!golden_matches(
        golden,
        "1.0.0",
        "🧠 Sonnet | 🗿 v2.0.0\nline two\n",
        "2.0.0"
    ));
    let two_versions = "🗿 v9.9.9 -> 🗿 v1.0.0";
    assert!(!golden_matches(
        two_versions,
        "1.0.0",
        "🗿 v8.8.8 -> 🗿 v2.0.0",
        "2.0.0"
    ));
}

#[test]
fn color_and_link_escapes_are_stripped_and_nothing_else() {
    assert_eq!(strip_sgr("\x1b[38;2;1;2;3mhi\x1b[0m"), "hi");
    assert_eq!(strip_sgr("\x1b[Kkeep \x1b[3"), "\x1b[Kkeep \x1b[3");
    assert_eq!(
        strip_osc8("\x1b]8;;https://x.invalid\x1b\\link\x1b]8;;\x1b\\"),
        "link"
    );
    assert_eq!(
        strip_osc8("\x1b]8;id=1;no-terminator"),
        "\x1b]8;id=1;no-terminator"
    );
}

/// The recorder writes `cases.json` byte for byte as it is checked in, so a rewrite of one case
/// changes that case only (non-ASCII text stays `\u` escaped, as the earlier tool wrote it).
#[test]
fn the_manifest_writer_reproduces_the_checked_in_file() {
    let text = fs::read_to_string(manifest_path()).unwrap();
    assert_eq!(manifest_text(&load_manifest()), text);
    assert!(
        text.contains("\\ud83d\\udd00"),
        "the fixture keeps a surrogate-pair escape"
    );
}

/// Every golden file belongs to a case and a variant, every case has all three with a recorded
/// provenance, and the manifest names the same files the folder holds.
#[test]
fn every_case_has_its_three_goldens_with_provenance_and_no_golden_is_orphaned() {
    let manifest = load_manifest();
    let mut expected = Vec::new();
    for case in manifest["cases"].as_array().unwrap() {
        for variant in VARIANTS {
            let path = golden_path(case, variant);
            assert!(path.is_file(), "missing golden {}", path.display());
            let record = &case["recorded_with"]["variants"][variant];
            for key in ["version", "sha256", "stdout"] {
                assert!(
                    record[key].is_string(),
                    "{} {variant}: no {key}",
                    case["name"]
                );
            }
            expected.push(path.file_name().unwrap().to_string_lossy().into_owned());
        }
    }
    let mut found: Vec<String> = fs::read_dir(fixtures().join("golden"))
        .unwrap()
        .map(|e| e.unwrap().file_name().to_string_lossy().into_owned())
        .collect();
    found.sort();
    expected.sort();
    assert_eq!(found, expected);
}

/// Card C-65: the `copy` step puts a checked-in fixture file (the ledger of `b13-card-counts`)
/// into the temporary HOME's `.moai`, byte for byte.
#[test]
fn a_fixture_file_is_copied_under_the_moai_home() {
    let tmp = tempfile::tempdir().unwrap();
    let root = tmp.path();
    let moai_home = root.join("home").join(".moai");
    let step = json!({"kind": "copy", "from": "ledger/cards-5-50-12.sqlite",
                      "under": "moai_home", "path": "projects/demo/ledger.db"});
    apply_setup(&step, root, &root.join("project"), &moai_home).unwrap();
    let copied = fs::read(moai_home.join("projects/demo/ledger.db")).unwrap();
    let source = fs::read(fixtures().join("ledger/cards-5-50-12.sqlite")).unwrap();
    assert_eq!(copied, source);
    assert_eq!(&copied[..16], b"SQLite format 3\0");
}

/// A copy source outside the fixture folder, or a base other than the Go home, is refused and
/// creates nothing.
#[test]
fn a_copy_source_outside_the_fixtures_or_another_base_is_refused() {
    for step in [
        json!({"kind": "copy", "from": "../statusline/cases.json", "under": "moai_home", "path": "x"}),
        json!({"kind": "copy", "from": "/etc/hosts", "under": "moai_home", "path": "x"}),
        json!({"kind": "copy", "from": "ledger", "under": "moai_home", "path": "x"}),
        json!({"kind": "copy", "from": "ledger/cards-5-50-12.sqlite", "path": "x"}),
        json!({"kind": "copy", "from": "ledger/cards-5-50-12.sqlite", "under": "cli_home", "path": "x"}),
        json!({"kind": "write", "path": "x", "content": "y", "under": "elsewhere"}),
        json!({"kind": "no-such-step"}),
    ] {
        let tmp = tempfile::tempdir().unwrap();
        let root = tmp.path();
        let moai_home = root.join("home").join(".moai");
        let result = apply_setup(&step, root, &root.join("project"), &moai_home);
        assert!(result.is_err(), "{step} was not refused");
        assert!(!root.join("home").exists(), "{step} created the home");
    }
}

/// A `write` step lands inside the temporary root, under the base the step names.
#[test]
fn a_write_step_lands_under_the_base_it_names() {
    let tmp = tempfile::tempdir().unwrap();
    let root = tmp.path();
    let project = root.join("work").join("demo-project");
    let moai_home = root.join("home").join(".moai");
    for (under, base) in [
        (None, project.clone()),
        (Some("moai_home"), moai_home.clone()),
        (
            Some("cli_home"),
            root.join("home").join(moai_cli::HOME_STATE_DIR),
        ),
    ] {
        let mut step = json!({"kind": "write", "path": "a/b.txt", "content": "text"});
        if let Some(under) = under {
            step["under"] = json!(under);
        }
        apply_setup(&step, root, &project, &moai_home).unwrap();
        assert_eq!(fs::read_to_string(base.join("a/b.txt")).unwrap(), "text");
    }
}

/// #2350 AC-2/AC-5: the harness sandbox denies real-HOME reads, exec of anything but the binary
/// under test and git, and network. Each probe has a control that must succeed, so a profile
/// that fails everything does not pass. macOS only (`sandbox-exec`).
#[cfg(target_os = "macos")]
mod sandbox {
    use super::*;

    fn run_in(exe: &str, args: &[&str]) -> std::process::Output {
        Command::new("/usr/bin/sandbox-exec")
            .args(["-p", &sandbox_profile(Path::new(exe)), exe])
            .args(args)
            .current_dir("/")
            .output()
            .expect("sandbox-exec runs")
    }

    fn denied(out: &std::process::Output) -> bool {
        !out.status.success()
            && String::from_utf8_lossy(&out.stderr).contains("Operation not permitted")
    }

    #[test]
    fn real_home_is_not_readable() {
        let home = std::env::var("HOME").unwrap();
        let home = fs::canonicalize(&home).unwrap_or_else(|_| PathBuf::from(&home));
        assert!(run_in("/bin/ls", &["/"]).status.success());
        let out = run_in("/bin/ls", &[&home.to_string_lossy()]);
        assert!(denied(&out), "{:?}", String::from_utf8_lossy(&out.stderr));
    }

    #[test]
    fn only_the_binary_and_git_are_executed() {
        assert!(run_in("/bin/bash", &["-c", "echo builtin"])
            .status
            .success());
        let out = run_in("/bin/bash", &["-c", "/bin/echo external"]);
        assert!(denied(&out), "{:?}", String::from_utf8_lossy(&out.stderr));
    }

    #[test]
    fn network_is_denied() {
        // Port 9 on loopback: without the deny this is `Connection refused`.
        let out = run_in("/bin/bash", &["-c", "exec 3<>/dev/tcp/127.0.0.1/9"]);
        assert!(denied(&out), "{:?}", String::from_utf8_lossy(&out.stderr));
    }
}
