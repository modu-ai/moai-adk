//! SPEC-20261121-security-scan-turn: `moai hook-dispatch --event security-turn` (Claude Code, Stop)
//! reads what a turn changed — the added lines of the tracked changes against HEAD and the
//! untracked files the project does not ignore — scans each file with the scanner of
//! security-scan, and only ever notifies: the rule `security_turn` in `[hooks.modes]` of the HOME
//! hook-config.toml is `observe` by default (one `would_notify` line, no output), `enforce`
//! prints a `systemMessage` and writes a `notified` line, `off` does nothing at all. The same
//! findings are reported once per session. Every case runs the real binary in a temporary git
//! repository and HOME; secret-shaped words are assembled at run time.
#![cfg(unix)]
mod common;

use common::CleanEnv;
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Output, Stdio};
use std::time::Instant;

const MIB: usize = 1024 * 1024;

struct Fx {
    _dir: tempfile::TempDir,
    root: PathBuf,
    proj: PathBuf,
    home: PathBuf,
}

fn fx() -> Fx {
    let dir = tempfile::TempDir::new().unwrap();
    let root = fs::canonicalize(dir.path()).unwrap();
    let (proj, home) = (root.join("proj"), root.join("home"));
    fs::create_dir_all(proj.join(".moai-cli")).unwrap();
    fs::create_dir_all(home.join(".moai-cli")).unwrap();
    Fx {
        _dir: dir,
        root,
        proj,
        home,
    }
}

/// What one run printed: (exit, stdout, stderr).
type Ran = (i32, String, String);

fn real_git() -> PathBuf {
    std::env::split_paths(&std::env::var_os("PATH").unwrap_or_default())
        .map(|dir| dir.join("git"))
        .find(|path| path.is_file())
        .expect("git on PATH")
}

/// The classes of the table: (class, severity, one line that holds only that class), in the
/// order unsafe-deserialization, dom-injection-xss, hardcoded-secret, code-injection-eval,
/// sql-injection, command-injection.
fn samples() -> Vec<(&'static str, &'static str, String)> {
    let name = format!("{}{}", "pass", "word");
    vec![
        (
            "unsafe-deserialization",
            "high",
            "data = pickle.loads(blob)".to_string(),
        ),
        (
            "dom-injection-xss",
            "high",
            "el.innerHTML = html;".to_string(),
        ),
        (
            "hardcoded-secret",
            "critical",
            format!("{name} = \"hunter2\""),
        ),
        (
            "code-injection-eval",
            "high",
            "result = eval(user_text)".to_string(),
        ),
        (
            "sql-injection",
            "high",
            "q = \"SELECT * FROM users WHERE id = \" + uid".to_string(),
        ),
        ("command-injection", "high", "os.system(cmd)".to_string()),
    ]
}

fn line_of(class: &str) -> String {
    samples()
        .into_iter()
        .find(|(name, _, _)| *name == class)
        .unwrap()
        .2
}

impl Fx {
    fn config(&self, text: &str) {
        fs::write(self.home.join(".moai-cli/hook-config.toml"), text).unwrap();
    }

    /// The rule `security_turn` in the mode `mode` (a bare word of the three modes, or any word).
    fn mode(&self, mode: &str) {
        self.config(&format!("[hooks.modes]\nsecurity_turn = \"{mode}\"\n"));
    }

    /// A git run of the test itself, with the author set and no global configuration.
    fn git(&self, args: &[&str]) {
        let out = Command::new(real_git())
            .clean_env()
            .env("PATH", std::env::var_os("PATH").unwrap_or_default())
            .env("HOME", &self.home)
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_SYSTEM", "/dev/null")
            .args(["-c", "user.name=t", "-c", "user.email=t@e.test"])
            .args([
                "-c",
                "commit.gpgsign=false",
                "-c",
                "init.defaultBranch=main",
            ])
            .args(args)
            .current_dir(&self.proj)
            .output()
            .unwrap();
        assert!(
            out.status.success(),
            "git {args:?}: {}",
            String::from_utf8_lossy(&out.stderr)
        );
    }

    /// The project becomes a git repository whose one commit holds `files`.
    fn repo(&self, files: &[(&str, &str)]) {
        self.git(&["init", "-q"]);
        for (rel, content) in files {
            self.put(rel, content.as_bytes());
        }
        self.git(&["add", "-A"]);
        self.git(&["commit", "-q", "-m", "base"]);
    }

    fn put(&self, rel: &str, content: &[u8]) {
        let path = self.proj.join(rel);
        fs::create_dir_all(path.parent().unwrap()).unwrap();
        fs::write(path, content).unwrap();
    }

    /// A fake `git` first on PATH: it records its arguments in `calls.txt` and then either runs
    /// the real git or, for the arguments `guard` names (a `case` body), fails or hangs.
    fn fake_git(&self, guard: &str) -> PathBuf {
        let bin = self.root.join("fakebin");
        fs::create_dir_all(&bin).unwrap();
        let script = format!(
            "#!/bin/sh\necho \"$*\" >> \"{calls}\"\nfor a in \"$@\"; do\n  case \"$a\" in\n{guard}\n  esac\ndone\nexec \"{real}\" \"$@\"\n",
            calls = self.root.join("calls.txt").display(),
            real = real_git().display(),
        );
        common::write_executable(&bin.join("git"), &script);
        bin
    }

    /// The git calls the binary made through the fake git.
    fn calls(&self) -> String {
        fs::read_to_string(self.root.join("calls.txt")).unwrap_or_default()
    }

    fn dispatch(&self, host: &str, event: &str, stdin: &[u8], fake: Option<&Path>) -> Ran {
        let mut path = std::env::var_os("PATH").unwrap_or_default();
        if let Some(bin) = fake {
            let mut dirs = vec![bin.to_path_buf()];
            dirs.extend(std::env::split_paths(&path));
            path = std::env::join_paths(dirs).unwrap();
        }
        let mut child = Command::new(common::MOAI_BIN)
            .clean_env()
            .env("PATH", path)
            .env("HOME", &self.home)
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_SYSTEM", "/dev/null")
            .current_dir(&self.proj)
            .args(["hook-dispatch", "--host", host, "--event", event])
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
            .unwrap();
        common::feed_stdin(&mut child, stdin);
        let out: Output = child.wait_with_output().unwrap();
        (
            out.status.code().unwrap_or(-1),
            String::from_utf8_lossy(&out.stdout).trim().to_string(),
            String::from_utf8_lossy(&out.stderr).trim().to_string(),
        )
    }

    /// A Claude Code Stop input of the session `session`.
    fn stop_input(&self, session: &str, active: bool) -> Vec<u8> {
        json!({"session_id": session, "hook_event_name": "Stop", "stop_hook_active": active,
               "cwd": self.proj})
        .to_string()
        .into_bytes()
    }

    /// One Stop of the session `session` through the real git.
    fn turn_end(&self, session: &str) -> Ran {
        self.dispatch(
            "claude",
            "security-turn",
            &self.stop_input(session, false),
            None,
        )
    }

    /// The JSON lines of the log of this event.
    fn lines(&self) -> Vec<Value> {
        self.lines_of("security-scan.jsonl")
    }

    fn lines_of(&self, file: &str) -> Vec<Value> {
        fs::read_to_string(self.proj.join(".moai-cli/receipts").join(file))
            .unwrap_or_default()
            .lines()
            .filter_map(|l| serde_json::from_str(l).ok())
            .collect()
    }

    fn of_kind(&self, kind: &str) -> Vec<Value> {
        self.lines()
            .into_iter()
            .filter(|l| l["kind"] == json!(kind))
            .collect()
    }

    /// The state of the SPEC's first case: a commit with an old finding, a tracked file edited to
    /// hold three classes, an untracked file with three more, and an ignored file with one.
    fn scenario(&self) {
        self.repo(&[
            (".gitignore", "ignored.py\n"),
            ("src/a.py", "x = 1\n"),
            (
                "src/old.py",
                &format!("{}\n", line_of("unsafe-deserialization")),
            ),
        ]);
        self.put(
            "src/a.py",
            format!(
                "x = 1\n{}\n{}\n{}\n",
                line_of("code-injection-eval"),
                line_of("dom-injection-xss"),
                line_of("hardcoded-secret")
            )
            .as_bytes(),
        );
        self.put(
            "src/old.py",
            format!("{}\ny = 2\n", line_of("unsafe-deserialization")).as_bytes(),
        );
        self.put(
            "src/new.py",
            format!(
                "{}\n{}\n{}\n",
                line_of("unsafe-deserialization"),
                line_of("command-injection"),
                line_of("sql-injection")
            )
            .as_bytes(),
        );
        self.put(
            "ignored.py",
            format!("{}\n", line_of("hardcoded-secret")).as_bytes(),
        );
    }
}

/// The `systemMessage` of an answer, after checking that it is the only member.
fn system_message(stdout: &str) -> String {
    let v: Value = serde_json::from_str(stdout).unwrap_or_else(|e| panic!("{e}: {stdout}"));
    let members: Vec<&String> = v.as_object().unwrap().keys().collect();
    assert_eq!(members, ["systemMessage"], "only systemMessage: {v}");
    v["systemMessage"]
        .as_str()
        .unwrap_or_else(|| panic!("no systemMessage: {v}"))
        .to_string()
}

/// The finding entries of a notification: lines `- <class> (<severity>) <path>:<line>`.
fn entries(text: &str) -> Vec<String> {
    text.lines()
        .filter(|l| l.starts_with("- ") && !l.starts_with("- and "))
        .map(str::to_string)
        .collect()
}

fn assert_silent(ran: &Ran, why: &str) {
    assert_eq!(ran, &(0, String::new(), String::new()), "{why}");
}

/// AC-1, REQ-1 / REQ-2 / REQ-4 / REQ-9 / REQ-10: in enforce mode the notification lists the first
/// five findings, tracked files by path and then untracked files by path, each by line, counts
/// the sixth, names no matched text and no file that held nothing new, and the observe line names
/// all six classes. The file `.gitignore` names is not read, the line that stood in the file
/// before the turn is not reported.
#[test]
fn notify_lists_new_findings_of_tracked_and_untracked_files_by_relative_path() {
    let f = fx();
    f.scenario();
    f.mode("enforce");
    let (code, out, err) = f.turn_end("s1");
    assert_eq!((code, err.as_str()), (0, ""), "{out}");
    let text = system_message(&out);
    assert_eq!(
        entries(&text),
        [
            "- code-injection-eval (high) src/a.py:2",
            "- dom-injection-xss (high) src/a.py:3",
            "- hardcoded-secret (critical) src/a.py:4",
            "- unsafe-deserialization (high) src/new.py:1",
            "- command-injection (high) src/new.py:2",
        ],
        "{text}"
    );
    assert!(text.contains("- and 1 more findings not listed."), "{text}");
    for absent in [
        "ignored.py",
        "src/old.py",
        "hunter2",
        "pickle.loads",
        "eval(",
    ] {
        assert!(!text.contains(absent), "{absent} in {text}");
    }

    let o = fx();
    o.scenario();
    o.mode("observe");
    assert_silent(&o.turn_end("s1"), "observe prints nothing");
    let seen = o.of_kind("would_notify");
    assert_eq!(seen.len(), 1, "{:?}", o.lines());
    assert_eq!(seen[0]["findings"], json!(6));
    // The line keeps the class names in table order, as the write-time duty's line does.
    let classes: Vec<&str> = samples().iter().map(|(name, _, _)| *name).collect();
    assert_eq!(seen[0]["classes"], json!(classes));
}

/// REQ-1: a new file that is staged is a tracked change of HEAD and is read like one.
#[test]
fn notify_reads_a_staged_new_file() {
    let f = fx();
    f.repo(&[("src/a.py", "x = 1\n")]);
    f.put(
        "src/s.py",
        format!("{}\n", line_of("code-injection-eval")).as_bytes(),
    );
    f.git(&["add", "src/s.py"]);
    f.mode("enforce");
    let (code, out, err) = f.turn_end("s1");
    assert_eq!((code, err.as_str()), (0, ""), "{out}");
    assert_eq!(
        entries(&system_message(&out)),
        ["- code-injection-eval (high) src/s.py:1"]
    );
}

/// AC-2, REQ-3 / REQ-4 / REQ-9: without an entry, with observe, with an unknown word and with a
/// differently cased Enforce, nothing is printed and one `would_notify` line of the rule
/// `security_turn` is written to the log of this event — class names, count, digest, host,
/// session, time, no path and no text; enforce prints the object and writes a `notified` line.
#[test]
fn modes_other_than_enforce_only_log_and_enforce_notifies() {
    for config in [None, Some("observe"), Some("bogus"), Some("Enforce")] {
        let f = fx();
        f.scenario();
        if let Some(word) = config {
            f.mode(word);
        }
        assert_silent(&f.turn_end("s1"), &format!("{config:?}"));
        let seen = f.of_kind("would_notify");
        assert_eq!(seen.len(), 1, "{config:?}: {:?}", f.lines());
        assert!(f.of_kind("notified").is_empty(), "{config:?}");
        assert!(f.lines_of("hook-engine.jsonl").is_empty(), "{config:?}");
        let line = &seen[0];
        assert_eq!(line["rule"], json!("security_turn"));
        assert_eq!(line["host"], json!("claude"));
        assert_eq!(line["findings"], json!(6));
        assert!(
            line["session"].is_string() && line["call"].is_null(),
            "{line}"
        );
        assert!(line["unix_ms"].as_u64().unwrap() > 0);
        let digest = line["digest"].as_str().unwrap_or_else(|| panic!("{line}"));
        assert!(
            digest.len() == 16 && digest.bytes().all(|b| b.is_ascii_hexdigit()),
            "{line}"
        );
        let raw = line.to_string();
        for absent in ["src/", "pickle", "hunter2", "eval(", "ignored"] {
            assert!(!raw.contains(absent), "{absent} in {raw}");
        }
    }

    let f = fx();
    f.scenario();
    f.mode("enforce");
    let (code, out, err) = f.turn_end("s1");
    assert_eq!((code, err.as_str()), (0, ""));
    system_message(&out);
    assert!(f.of_kind("would_notify").is_empty(), "{:?}", f.lines());
    let sent = f.of_kind("notified");
    assert_eq!(sent.len(), 1, "{:?}", f.lines());
    assert_eq!(sent[0]["rule"], json!("security_turn"));
    assert_eq!(sent[0]["findings"], json!(6));
}

/// REQ-5: `off` runs no git command, reads no file and writes and prints nothing — the fake git
/// on PATH records every call, and the files hold findings that would be read otherwise.
#[test]
fn modes_off_runs_no_git_and_writes_nothing() {
    let f = fx();
    f.scenario();
    f.mode("off");
    let bin = f.fake_git("");
    let ran = f.dispatch(
        "claude",
        "security-turn",
        &f.stop_input("s1", false),
        Some(&bin),
    );
    assert_silent(&ran, "off");
    assert_eq!(f.calls(), "", "no git command under off");
    assert!(f.lines().is_empty() && f.lines_of("hook-engine.jsonl").is_empty());
}

/// AC-3, REQ-7: the same findings are reported once per session, in enforce and in observe; a
/// second session is told again; a changed set of findings is told again.
#[test]
fn once_per_session_and_again_when_the_findings_change() {
    let f = fx();
    f.repo(&[("src/a.py", "x = 1\n")]);
    f.put(
        "src/new.py",
        format!("{}\n", line_of("code-injection-eval")).as_bytes(),
    );
    f.mode("enforce");
    let (_, first, _) = f.turn_end("s1");
    system_message(&first);
    assert_silent(&f.turn_end("s1"), "the same findings, the same session");
    assert_eq!(f.of_kind("notified").len(), 1, "{:?}", f.lines());
    let (_, other, _) = f.turn_end("s2");
    system_message(&other);
    assert_eq!(f.of_kind("notified").len(), 2, "{:?}", f.lines());
    // A new finding after the first was told: the set changed, so it is told again — once.
    f.put(
        "src/new.py",
        format!(
            "{}\n{}\n",
            line_of("code-injection-eval"),
            line_of("command-injection")
        )
        .as_bytes(),
    );
    let (_, changed, _) = f.turn_end("s1");
    assert_eq!(entries(&system_message(&changed)).len(), 2, "{changed}");
    assert_silent(&f.turn_end("s1"), "the changed set, told once");
    assert_eq!(f.of_kind("notified").len(), 3, "{:?}", f.lines());

    let o = fx();
    o.repo(&[("src/a.py", "x = 1\n")]);
    o.put(
        "src/new.py",
        format!("{}\n", line_of("code-injection-eval")).as_bytes(),
    );
    assert_silent(&o.turn_end("s1"), "observe");
    assert_silent(&o.turn_end("s1"), "observe, again");
    assert_eq!(o.of_kind("would_notify").len(), 1, "{:?}", o.lines());
}

/// AC-3, REQ-6: a turn that already continues because of a stop hook runs no git command, writes
/// no line and prints nothing.
#[test]
fn once_a_continuing_turn_is_left_alone() {
    let f = fx();
    f.scenario();
    f.mode("enforce");
    let bin = f.fake_git("");
    let ran = f.dispatch(
        "claude",
        "security-turn",
        &f.stop_input("s9", true),
        Some(&bin),
    );
    assert_silent(&ran, "stop_hook_active");
    assert_eq!(f.calls(), "", "no git command while the turn continues");
    assert!(f.lines().is_empty());
}

/// REQ-8: outside a repository and in a repository without a commit nothing is printed or
/// written, and the exit is 0.
#[test]
fn skips_a_folder_that_is_no_repository_and_a_repository_without_a_commit() {
    let f = fx();
    f.mode("enforce");
    f.put(
        "src/new.py",
        format!("{}\n", line_of("code-injection-eval")).as_bytes(),
    );
    assert_silent(&f.turn_end("s1"), "no repository");
    assert!(f.lines().is_empty());
    f.git(&["init", "-q"]);
    assert_silent(&f.turn_end("s1"), "no commit");
    assert!(f.lines().is_empty());
}

/// REQ-8: a git call that fails, in either of the two calls, or runs past three seconds, gives no
/// output, no line and exit 0 — even though the other part holds findings. The hang is cut within
/// five seconds.
#[test]
fn skips_when_git_fails_or_hangs() {
    for (guard, why) in [
        ("    diff) exit 1;;", "the diff fails"),
        ("    ls-files) exit 1;;", "the untracked list fails"),
    ] {
        let f = fx();
        f.scenario();
        f.mode("enforce");
        let bin = f.fake_git(guard);
        let ran = f.dispatch(
            "claude",
            "security-turn",
            &f.stop_input("s1", false),
            Some(&bin),
        );
        assert_silent(&ran, why);
        assert!(f.lines().is_empty(), "{why}");
    }

    let f = fx();
    f.scenario();
    f.mode("enforce");
    let bin = f.fake_git("    diff) exec sleep 30;;");
    let started = Instant::now();
    let ran = f.dispatch(
        "claude",
        "security-turn",
        &f.stop_input("s1", false),
        Some(&bin),
    );
    assert!(
        started.elapsed().as_secs_f64() < 5.0,
        "cut within five seconds: {:?}",
        started.elapsed()
    );
    assert_silent(&ran, "the diff hangs");
    assert!(f.lines().is_empty());
}

/// REQ-11: a file that is binary, not UTF-8, a link (to a file outside the project and to one
/// inside it) is skipped; the one readable file is the only finding.
#[test]
fn skips_files_that_are_binary_not_utf8_or_links() {
    let f = fx();
    f.repo(&[
        (
            "src/real.py",
            &format!("{}\n", line_of("code-injection-eval")),
        ),
        ("src/a.py", "x = 1\n"),
    ]);
    let eval = line_of("code-injection-eval");
    f.put("bin.py", format!("\0\n{eval}\n").as_bytes());
    let mut latin = vec![0xff, b'\n'];
    latin.extend_from_slice(eval.as_bytes());
    f.put("latin.py", &latin);
    fs::write(f.root.join("outside.py"), format!("{eval}\n")).unwrap();
    std::os::unix::fs::symlink("../outside.py", f.proj.join("out_link.py")).unwrap();
    std::os::unix::fs::symlink("src/real.py", f.proj.join("in_link.py")).unwrap();
    f.put(
        "ok.py",
        format!("{}\n", line_of("command-injection")).as_bytes(),
    );
    assert_silent(&f.turn_end("s1"), "observe");
    let seen = f.of_kind("would_notify");
    assert_eq!(seen.len(), 1, "{:?}", f.lines());
    assert_eq!(seen[0]["findings"], json!(1));
    assert_eq!(seen[0]["classes"], json!(["command-injection"]));
}

/// A text of exactly `size` bytes whose first line is a finding of class eval.
fn text_of(size: usize) -> String {
    let head = format!("{}\n", line_of("code-injection-eval"));
    format!("{head}{}", "#".repeat(size - head.len()))
}

/// REQ-11: one mebibyte of text is read, one byte more is not — for an untracked file and for a
/// tracked one.
#[test]
fn skips_a_file_of_more_than_one_mebibyte() {
    for (tracked, size, findings) in [
        (false, MIB, 1),
        (false, MIB + 1, 0),
        (true, MIB, 1),
        (true, MIB + 1, 0),
    ] {
        let f = fx();
        f.repo(&[("src/a.py", "x = 1\n")]);
        if tracked {
            f.put("src/a.py", text_of(size).as_bytes());
        } else {
            f.put("src/big.py", text_of(size).as_bytes());
        }
        assert_silent(&f.turn_end("s1"), "observe");
        let seen = f.of_kind("would_notify");
        let label = format!("tracked={tracked} size={size}");
        if findings == 0 {
            assert!(seen.is_empty(), "{label}: {:?}", f.lines());
        } else {
            assert_eq!(seen.len(), 1, "{label}: {:?}", f.lines());
            assert_eq!(seen[0]["findings"], json!(findings), "{label}");
        }
    }
}

/// REQ-11: two hundred untracked files are read, the two hundred and first is not.
#[test]
fn skips_untracked_files_past_two_hundred() {
    for (count, findings) in [(200, 200), (201, 200)] {
        let f = fx();
        f.repo(&[("src/a.py", "x = 1\n")]);
        for n in 0..count {
            f.put(
                &format!("many/f{n:03}.py"),
                format!("{}\n", line_of("code-injection-eval")).as_bytes(),
            );
        }
        assert_silent(&f.turn_end("s1"), "observe");
        let seen = f.of_kind("would_notify");
        assert_eq!(seen.len(), 1, "{count}: {:?}", f.lines());
        assert_eq!(seen[0]["findings"], json!(findings), "{count} files");
    }
}

/// The length of what the binary's diff call prints in the project.
fn diff_len(f: &Fx) -> usize {
    let out = Command::new(real_git())
        .clean_env()
        .env("PATH", std::env::var_os("PATH").unwrap_or_default())
        .env("HOME", &f.home)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .args([
            "diff",
            "HEAD",
            "-U0",
            "--no-ext-diff",
            "--no-color",
            "--no-renames",
        ])
        .current_dir(&f.proj)
        .output()
        .unwrap();
    assert!(out.status.success());
    out.stdout.len()
}

/// REQ-11: a diff of exactly 4 MiB is read; one of 4 MiB plus one byte is not read at all — the
/// untracked file is still read. Five tracked files each hold one finding on their first line.
#[test]
fn skips_a_diff_of_more_than_four_mebibytes() {
    for (extra, tracked_findings) in [(0usize, 5usize), (1, 0)] {
        let f = fx();
        let base: Vec<(String, String)> = (0..5)
            .map(|n| (format!("src/t{n}.py"), "x = 1\n".into()))
            .collect();
        let refs: Vec<(&str, &str)> = base.iter().map(|(p, c)| (p.as_str(), c.as_str())).collect();
        f.repo(&refs);
        let head = format!("{}\n", line_of("code-injection-eval"));
        for n in 0..4 {
            f.put(
                &format!("src/t{n}.py"),
                format!("{head}{}\n", "#".repeat(900_000)).as_bytes(),
            );
        }
        f.put(
            "src/t4.py",
            format!("{head}{}\n", "#".repeat(1000)).as_bytes(),
        );
        let measured = diff_len(&f);
        let pad = 1000
            + (4 * MIB + extra)
                .checked_sub(measured)
                .expect("the diff is under 4 MiB");
        f.put(
            "src/t4.py",
            format!("{head}{}\n", "#".repeat(pad)).as_bytes(),
        );
        assert_eq!(
            diff_len(&f),
            4 * MIB + extra,
            "the diff is tuned to the edge"
        );
        f.put(
            "src/u.py",
            format!("{}\n", line_of("command-injection")).as_bytes(),
        );
        assert_silent(&f.turn_end("s1"), "observe");
        let seen = f.of_kind("would_notify");
        assert_eq!(seen.len(), 1, "extra={extra}: {:?}", f.lines());
        assert_eq!(
            seen[0]["findings"],
            json!(tracked_findings + 1),
            "extra={extra}"
        );
    }
}

/// REQ-12: the hosts codex and omp get the usage error of the engine, exit 1, and nothing is
/// written.
#[test]
fn skips_hosts_other_than_claude_with_a_usage_error() {
    for host in ["codex", "omp"] {
        let f = fx();
        f.scenario();
        f.mode("enforce");
        let (code, out, err) = f.dispatch(host, "security-turn", &f.stop_input("s1", false), None);
        assert_eq!(code, 1, "{host}: {err}");
        assert_eq!(out, "", "{host}");
        assert!(
            err.contains("security-turn is Claude Code only"),
            "{host}: {err}"
        );
        assert!(f.lines().is_empty(), "{host}");
    }
}

/// AC-5: the usage error for an unknown event lists the new word.
#[test]
fn usage_error_lists_the_event_word() {
    let f = fx();
    let (code, out, err) = f.dispatch("claude", "no-such-event", b"{}", None);
    assert_eq!(code, 1, "{err}");
    assert_eq!(out, "");
    assert!(err.contains("security-turn"), "{err}");
}
