//! SPEC-20261124-launcher-retire-role-flags (card C-449; user decision D-1 2026-11-09, restated
//! 2026-11-24): the launcher retires `-l|-a`. A role flag is refused with one message that names
//! `--name`, `/moai work` and `/moai leader`; the launcher then starts no host, takes no number
//! or lock and writes no file. No help screen offers a role flag, and no live text tells anyone
//! to use one or names a `mode-` spelling. Every case runs in a temporary git project with a
//! temporary or missing HOME; no real host is started (a stub `claude` stands in).
mod common;
use common::CleanEnv;
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Output};

/// Each role spelling the launcher used to take, with the host word it came after.
const ROLE_FORMS: [&[&str]; 6] = [
    &["cc", "-l"],
    &["cc", "-a"],
    &["glm", "--leader"],
    &["glm", "--agent"],
    &["-l"],
    &["-a", "cc"],
];

fn git_project() -> tempfile::TempDir {
    let project = tempfile::tempdir().unwrap();
    let init = common::git::command()
        .args(["init", "-q"])
        .current_dir(project.path())
        .status()
        .unwrap();
    assert!(init.success());
    project
}

fn text(out: &Output) -> String {
    format!(
        "{}{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    )
}

/// The launcher in `dir` with a clean environment; `home` is the HOME it sees.
fn launcher(dir: &Path, home: &Path, args: &[&str]) -> Command {
    let mut cmd = Command::new(common::MOAI_BIN);
    cmd.args(args)
        .current_dir(dir)
        .clean_env()
        .env("HOME", home)
        .env("ANTHROPIC_AUTH_TOKEN", "dummy-token-for-tests");
    cmd
}

/// AC-1 (REQ-1): every role spelling exits non-zero and says the replacement, with and without
/// `--dry-run`; the message names `--name`, `/moai work` and `/moai leader` and no `mode-`.
#[test]
fn every_role_flag_is_refused_with_the_replacement() {
    let home = tempfile::tempdir().unwrap();
    for form in ROLE_FORMS {
        for dry in [false, true] {
            let project = git_project();
            let mut args: Vec<&str> = form.to_vec();
            if dry {
                args.push("--dry-run");
            }
            // A stub host that fails the case when it is run: a refused launch starts nothing.
            let bin = tempfile::tempdir().unwrap();
            let mark = bin.path().join("host-was-run");
            write_stub(bin.path(), &mark);
            let out = launcher(project.path(), home.path(), &args)
                .env("PATH", stub_path(bin.path()))
                .output()
                .unwrap();
            let said = text(&out);
            assert!(!out.status.success(), "{args:?} was accepted: {said}");
            for want in ["--name", "/moai work", "/moai leader"] {
                assert!(said.contains(want), "{args:?} lacks {want}: {said}");
            }
            assert!(
                !said.contains("mode-"),
                "{args:?} names a mode spelling: {said}"
            );
            assert!(!mark.exists(), "{args:?} started the host");
        }
    }
}

/// A `claude` stub that leaves a mark when it runs.
fn write_stub(dir: &Path, mark: &Path) {
    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt;
        let path = dir.join("claude");
        fs::write(&path, format!("#!/bin/sh\ntouch '{}'\n", mark.display())).unwrap();
        fs::set_permissions(&path, fs::Permissions::from_mode(0o755)).unwrap();
    }
    #[cfg(not(unix))]
    let _ = (dir, mark);
}

fn stub_path(stub: &Path) -> String {
    format!("{}:/usr/bin:/bin", stub.display())
}

/// Every file and folder under `dir`, relative, sorted; `.git` is listed as one entry only
/// when it holds something this test did not create, so the git folder's own files are skipped.
fn tree(dir: &Path) -> Vec<PathBuf> {
    fn walk(base: &Path, dir: &Path, out: &mut Vec<PathBuf>) {
        for entry in fs::read_dir(dir).unwrap() {
            let entry = entry.unwrap();
            let path = entry.path();
            out.push(path.strip_prefix(base).unwrap().to_path_buf());
            if path.is_dir() && entry.file_name() != ".git" {
                walk(base, &path, out);
            }
        }
    }
    let mut out = Vec::new();
    walk(dir, dir, &mut out);
    out.sort();
    out
}

/// AC-2 (REQ-1): a refused launch leaves the project and its clone folder as they were: no
/// project file, no `agent-N` number, no `.live` lock and no session folder.
#[cfg(unix)]
#[test]
fn a_refused_role_flag_starts_and_writes_nothing() {
    let home = tempfile::tempdir().unwrap();
    for form in ROLE_FORMS {
        let project = git_project();
        let before = tree(project.path());
        let sessions = project.path().join(".git/.moai-cli");
        let bin = tempfile::tempdir().unwrap();
        let mark = bin.path().join("host-was-run");
        write_stub(bin.path(), &mark);
        let out = launcher(project.path(), home.path(), form)
            .env("PATH", stub_path(bin.path()))
            .output()
            .unwrap();
        assert!(
            !out.status.success(),
            "{form:?} was accepted: {}",
            text(&out)
        );
        assert!(!mark.exists(), "{form:?} started the host");
        assert_eq!(
            tree(project.path()),
            before,
            "{form:?} wrote a project file"
        );
        assert!(!sessions.exists(), "{form:?} created the session folder");
    }
}

/// The help texts a person reads to learn how to start a session.
fn help_screens() -> Vec<(String, String)> {
    let home = tempfile::tempdir().unwrap();
    let mut out = Vec::new();
    for args in [
        &["cc", "--help"][..],
        &["glm", "--help"][..],
        &["--help"][..],
        &["help", "--all"][..],
    ] {
        let result = Command::new(common::MOAI_BIN)
            .clean_env()
            .env("HOME", home.path())
            .env("LANG", "C")
            .env("COLUMNS", "200")
            .args(args)
            .output()
            .unwrap();
        assert!(result.status.success(), "{args:?}: {result:?}");
        let flat = String::from_utf8(result.stdout)
            .unwrap()
            .split_whitespace()
            .collect::<Vec<_>>()
            .join(" ");
        out.push((args.join(" "), flat));
    }
    out
}

/// AC-4 (REQ-2): no launcher help or root screen offers a role flag.
#[test]
fn no_help_screen_offers_a_role_flag() {
    const OFFERS: [&str; 9] = [
        "-l, --leader",
        "-a, --agent",
        "--leader",
        "--agent <HOST>",
        "cc|glm -l|-a",
        "cc -l",
        "glm -a",
        "moai -l",
        "moai -a",
    ];
    for (name, text) in help_screens() {
        for offer in OFFERS {
            assert!(
                !text.contains(offer),
                "`moai {name}` still offers `{offer}`"
            );
        }
    }
}

/// AC-6 (REQ-3): `--name` names the session; without it no name and no `agent-N` is added.
#[test]
fn a_session_is_named_by_name_only() {
    let home = tempfile::tempdir().unwrap();
    let run = |args: &[&str]| -> serde_json::Value {
        let project = git_project();
        let out = launcher(project.path(), home.path(), args)
            .arg("--dry-run")
            .output()
            .unwrap();
        assert!(out.status.success(), "{args:?}: {}", text(&out));
        serde_json::from_slice(&out.stdout).unwrap()
    };
    let named = run(&["cc", "--name", "lane-x"]);
    assert_eq!(named["name"], "lane-x");
    let argv = named["argv"].to_string();
    assert!(argv.contains("\"--name\",\"lane-x\""), "{argv}");
    for _ in 0..2 {
        let plain = run(&["cc"]);
        assert_eq!(plain["name"], serde_json::Value::Null);
        assert!(!plain.to_string().contains("agent-"), "{plain}");
    }
}

fn root_file(rel: &str) -> String {
    fs::read_to_string(Path::new(env!("CARGO_MANIFEST_DIR")).join(rel))
        .unwrap_or_else(|e| panic!("{rel}: {e}"))
}

/// The files that tell a person how to start a session (REQ-8, REQ-11): the list is named here,
/// not found by a repository-wide scan. History files (the decisions log, closed SPECs, the
/// design reviews, the release notes) are not in it.
const LIVE_FILES: [&str; 5] = [
    "CLAUDE.md",
    "install/README.md",
    "docs/user-entry-draft.md",
    "docs/host-support-matrix.md",
    "docs/rust-product-path.md",
];

/// A line that names a retired form is a record when it also says so.
const RETIRED_MARKERS: [&str; 10] = [
    "retired",
    "refuse",
    "earlier",
    "record",
    "withdrawn",
    "폐기",
    "거절",
    "옛",
    "이전",
    "없어졌",
];

/// AC-11 (REQ-8): no live line tells a person to use `-l|-a`; a line that mentions the old
/// forms says in the same line that they are retired or a record.
#[test]
fn no_live_text_tells_anyone_to_use_the_role_flags() {
    const FORMS: [&str; 14] = [
        "moai cc -l",
        "moai cc -a",
        "moai glm -l",
        "moai glm -a",
        "cc -l",
        "cc -a",
        "glm -l",
        "glm -a",
        "moai -l",
        "moai -a",
        "-l|-a",
        "-l\\|-a",
        "`-l`",
        "`-a`",
    ];
    let mut bad = Vec::new();
    for file in LIVE_FILES {
        for (n, line) in root_file(file).lines().enumerate() {
            let lower = line.to_lowercase();
            // A line about the withdrawn Codex CLI names that CLI's own `-s` and `-a`.
            let other_cli = lower.contains("codex");
            if FORMS.iter().any(|f| line.contains(f))
                && !other_cli
                && !RETIRED_MARKERS.iter().any(|m| lower.contains(m))
            {
                bad.push(format!("{file}:{}", n + 1));
            }
        }
    }
    assert!(
        bad.is_empty(),
        "live text still names the role flags: {bad:?}"
    );
}

/// AC-13 (REQ-11): no live line, and no launcher help screen, names a `mode-` spelling. Two
/// records are exempt: the withdrawn Codex first prompt (`$moai --mode-agent`) and the old
/// package listing of the workflow files.
#[test]
fn no_live_text_names_a_mode_spelling() {
    const SPELLINGS: [&str; 4] = ["mode-leader", "mode-agent", "--leader-mode", "--agent-mode"];
    const EXEMPT: [&str; 2] = ["$moai --mode-agent", "mode-leader,mode-agent}.md"];
    let mut bad = Vec::new();
    for file in LIVE_FILES {
        for (n, line) in root_file(file).lines().enumerate() {
            if SPELLINGS.iter().any(|s| line.contains(s))
                && !EXEMPT.iter().any(|e| line.contains(e))
            {
                bad.push(format!("{file}:{}", n + 1));
            }
        }
    }
    for (name, text) in help_screens() {
        for s in SPELLINGS {
            if text.contains(s) {
                bad.push(format!("`moai {name}` names {s}"));
            }
        }
    }
    assert!(
        bad.is_empty(),
        "live text still names a mode spelling: {bad:?}"
    );
}

/// A temporary project and HOME that print no missing-skill warning, with an optional user
/// launcher settings file; the host is never started (every case passes `--dry-run`).
struct Fixture {
    project: tempfile::TempDir,
    home: tempfile::TempDir,
}

impl Fixture {
    fn new(settings: Option<&str>) -> Self {
        let home = tempfile::tempdir().unwrap();
        for dir in [".claude/skills/moai", ".agents/skills/moai"] {
            fs::create_dir_all(home.path().join(dir)).unwrap();
            fs::write(home.path().join(dir).join("SKILL.md"), "skill\n").unwrap();
        }
        if let Some(text) = settings {
            let dir = home.path().join(".moai-cli/config");
            fs::create_dir_all(&dir).unwrap();
            fs::write(dir.join("launcher.yaml"), text).unwrap();
        }
        Fixture {
            project: git_project(),
            home,
        }
    }

    /// The dry-run JSON of `args`, which must have exited 0.
    fn dry(&self, args: &[&str]) -> serde_json::Value {
        let out = launcher(self.project.path(), self.home.path(), args)
            .arg("--dry-run")
            .output()
            .unwrap();
        assert!(out.status.success(), "{args:?}: {}", text(&out));
        serde_json::from_slice(&out.stdout).unwrap()
    }
}

fn argv_of(v: &serde_json::Value) -> Vec<String> {
    serde_json::from_value(v["argv"].clone()).unwrap()
}

fn has_pair(argv: &[String], flag: &str, value: &str) -> bool {
    argv.windows(2).any(|w| w[0] == flag && w[1] == value)
}

fn has_flag(argv: &[String], flag: &str) -> bool {
    argv.iter().any(|a| a == flag)
}

const BYPASS: &str = "--dangerously-skip-permissions";

/// AC-16 (REQ-13, REQ-15): with no setting, `cc` starts in auto mode at effort medium and `glm`
/// in the bypass at effort high, both with the model alias `opus`; the glm environment maps that
/// alias to `glm-5.4[1m]` and lists the variable names only.
#[test]
fn each_host_starts_with_its_own_permission_and_effort_default() {
    let fx = Fixture::new(None);
    let cc = fx.dry(&["cc"]);
    let argv = argv_of(&cc);
    assert!(has_pair(&argv, "--permission-mode", "auto"), "{argv:?}");
    assert!(!has_flag(&argv, BYPASS), "{argv:?}");
    assert!(has_pair(&argv, "--effort", "medium"), "{argv:?}");
    assert!(has_pair(&argv, "--model", "opus"), "{argv:?}");

    let glm = fx.dry(&["glm"]);
    let argv = argv_of(&glm);
    assert!(has_flag(&argv, BYPASS), "{argv:?}");
    assert!(!has_flag(&argv, "--permission-mode"), "{argv:?}");
    assert!(has_pair(&argv, "--effort", "high"), "{argv:?}");
    assert!(has_pair(&argv, "--model", "opus"), "{argv:?}");
    let keys = glm["env_keys"].to_string();
    assert!(keys.contains("ANTHROPIC_DEFAULT_OPUS_MODEL"), "{keys}");
    assert!(
        !glm.to_string().contains("dummy-token-for-tests"),
        "a token value is printed"
    );
    let notes = glm["notes"].to_string();
    assert!(
        notes.contains("opus, sonnet and fable to glm-5.4[1m]"),
        "{notes}"
    );
}

/// AC-16 (REQ-13, REQ-15): a given `--effort` wins on both hosts, and `--permissions` adds no
/// permission flag on either host.
#[test]
fn a_given_effort_and_permissions_win_over_the_host_defaults() {
    let fx = Fixture::new(None);
    for (host, effort) in [("cc", "low"), ("glm", "max")] {
        let argv = argv_of(&fx.dry(&[host, "--effort", effort]));
        assert!(has_pair(&argv, "--effort", effort), "{host}: {argv:?}");
        let argv = argv_of(&fx.dry(&[host, "--permissions"]));
        assert!(!has_flag(&argv, BYPASS), "{host}: {argv:?}");
        assert!(!has_flag(&argv, "--permission-mode"), "{host}: {argv:?}");
    }
}

/// AC-16 (REQ-13): `launcher.role_access` decides both hosts over the host default.
#[test]
fn launcher_role_access_decides_both_hosts() {
    for (setting, bypass, mode) in [
        ("auto", false, Some("auto")),
        ("full", true, None),
        ("prompt", false, None),
    ] {
        let fx = Fixture::new(Some(&format!(
            r#"{{"launcher":{{"role_access":"{setting}"}}}}"#
        )));
        for host in ["cc", "glm"] {
            let argv = argv_of(&fx.dry(&[host]));
            assert_eq!(
                has_flag(&argv, BYPASS),
                bypass,
                "{host} {setting}: {argv:?}"
            );
            match mode {
                Some(m) => assert!(has_pair(&argv, "--permission-mode", m), "{host}: {argv:?}"),
                None => assert!(!has_flag(&argv, "--permission-mode"), "{host}: {argv:?}"),
            }
        }
    }
}

/// AC-7 and AC-16 (REQ-4, REQ-15): a saved flat `launcher.models.<host>` entry gives the model
/// and effort over the host default; the `zai/` prefix is dropped for glm.
#[test]
fn the_flat_saved_entry_fills_model_and_effort() {
    let fx = Fixture::new(Some(
        r#"{"launcher":{"models":{
            "claude":{"model":"claude-opus-5-5","effort":"xhigh"},
            "glm":{"model":"zai/glm-5.4[1m]","effort":"max"}}}}"#,
    ));
    let cc = fx.dry(&["cc"]);
    let argv = argv_of(&cc);
    assert!(has_pair(&argv, "--model", "claude-opus-5-5"), "{argv:?}");
    assert!(has_pair(&argv, "--effort", "xhigh"), "{argv:?}");
    assert!(
        cc["notes"].to_string().contains("launcher.models.claude"),
        "{cc}"
    );
    let argv = argv_of(&fx.dry(&["glm"]));
    assert!(has_pair(&argv, "--model", "glm-5.4[1m]"), "{argv:?}");
    assert!(has_pair(&argv, "--effort", "max"), "{argv:?}");
    // A given flag still wins per field.
    let argv = argv_of(&fx.dry(&["glm", "--effort", "low"]));
    assert!(has_pair(&argv, "--effort", "low"), "{argv:?}");
    assert!(has_pair(&argv, "--model", "glm-5.4[1m]"), "{argv:?}");
}

/// AC-7 (REQ-4): with no entry for its host a launch keeps the host defaults, and an entry of the
/// other host is not read.
#[test]
fn no_saved_entry_keeps_the_roleless_defaults() {
    let fx = Fixture::new(Some(
        r#"{"launcher":{"models":{"glm":{"model":"zai/glm-5.4[1m]","effort":"max"}}}}"#,
    ));
    let argv = argv_of(&fx.dry(&["cc"]));
    assert!(has_pair(&argv, "--model", "opus"), "{argv:?}");
    assert!(has_pair(&argv, "--effort", "medium"), "{argv:?}");
}

/// AC-8 (REQ-5): saved `leader` and `agent` entries are not read and add one note that names both
/// key paths and `moai update`.
#[test]
fn old_role_entries_are_ignored_with_one_note() {
    let fx = Fixture::new(Some(
        r#"{"launcher":{"models":{"claude":{
            "leader":{"model":"claude-opus-5-5","effort":"high"},
            "agent":{"model":"claude-sonnet-5-5","effort":"high"}}}}}"#,
    ));
    let cc = fx.dry(&["cc"]);
    let argv = argv_of(&cc);
    assert!(has_pair(&argv, "--model", "opus"), "{argv:?}");
    assert!(has_pair(&argv, "--effort", "medium"), "{argv:?}");
    let notes: Vec<String> = serde_json::from_value(cc["notes"].clone()).unwrap();
    let ours: Vec<&String> = notes
        .iter()
        .filter(|n| n.contains("launcher.models.claude.leader"))
        .collect();
    assert_eq!(ours.len(), 1, "{notes:?}");
    for want in ["launcher.models.claude.agent", "moai update"] {
        assert!(ours[0].contains(want), "{}", ours[0]);
    }
}

/// AC-10 (REQ-7): the permission notes and the long help say that `launcher.role_access` sets the
/// mode of every launch and name no role as the one that chooses access.
#[test]
fn the_access_notes_name_no_role() {
    let fx = Fixture::new(None);
    const ROLE_ACCESS: [&str; 3] = ["sets its access", "role chooses", "role sets"];
    for host in ["cc", "glm"] {
        let v = fx.dry(&[host]);
        let notes = v["notes"].to_string();
        for bad in ROLE_ACCESS {
            assert!(!notes.contains(bad), "`{host}` notes say `{bad}`: {notes}");
        }
        let note: Vec<String> = serde_json::from_value(v["notes"].clone()).unwrap();
        let access = note
            .iter()
            .find(|n| n.contains("launcher.role_access"))
            .unwrap_or_else(|| panic!("`{host}` has no access note: {note:?}"));
        assert!(access.contains("every launch"), "`{host}`: {access}");
    }
    for (name, text) in help_screens() {
        for bad in ROLE_ACCESS {
            assert!(!text.contains(bad), "`moai {name}` says `{bad}`");
        }
        if name == "cc --help" || name == "glm --help" {
            assert!(
                text.contains("every launch"),
                "`moai {name}` lacks `every launch`"
            );
        }
    }
}
