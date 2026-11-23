//! SPEC-20261122-remove-omp-launcher (REQ-1 to REQ-4, REQ-6, REQ-8, REQ-9): `moai omp` prints a
//! removal notice, `moai help omp` and `moai help codex` print the notices, the host word `omp`
//! is refused like any other unknown host, the `cc` and `glm` launches name no OMP program, key
//! or spelling and do not change with a saved OMP value, and the help, the recorded usage error
//! and the install texts name no `moai omp` and no `moai codex`. Every case that runs the binary
//! uses a temporary git project and a temporary HOME; stub `omp` and `claude` programs on PATH
//! prove that nothing starts.
#![cfg(unix)]
mod common;
use common::CleanEnv;
use serde_json::{json, Value};
use std::os::unix::fs::PermissionsExt;
use std::process::{Command, Output};
use std::{fs, path::Path};

struct Fixture {
    project: tempfile::TempDir,
    bin: tempfile::TempDir,
    home: tempfile::TempDir,
}

fn fixture() -> Fixture {
    let project = tempfile::tempdir().unwrap();
    let init = common::git::command()
        .args(["init", "-q"])
        .current_dir(project.path())
        .status()
        .unwrap();
    assert!(init.success());
    // An initialized project, so no launch runs the first-launch initialization. The state
    // folder comes from the resolver, so the case follows the folder the product reads.
    let state = project
        .path()
        .join(moai_cli::state_root::dir(project.path()));
    fs::create_dir_all(&state).unwrap();
    fs::write(state.join("project.yaml"), "{}\n").unwrap();
    let bin = tempfile::tempdir().unwrap();
    // Stub `omp` and `claude` programs that leave a trace when they start.
    for host in ["omp", "claude"] {
        let path = bin.path().join(host);
        fs::write(
            &path,
            "#!/bin/sh\nprintf started > \"$MOAI_TEST_HOST_MARKER\"\nexit 0\n",
        )
        .unwrap();
        fs::set_permissions(&path, fs::Permissions::from_mode(0o755)).unwrap();
    }
    let home = tempfile::tempdir().unwrap();
    // The Claude skill copy, so a launch prints no missing-skill warning that names the HOME.
    let skill = home.path().join(".claude/skills/moai");
    fs::create_dir_all(&skill).unwrap();
    fs::write(skill.join("SKILL.md"), "skill\n").unwrap();
    // A file a person keeps in the OMP home folder.
    let omp = home.path().join(".omp/agent");
    fs::create_dir_all(&omp).unwrap();
    fs::write(omp.join("settings.json"), "{\"mine\": true}\n").unwrap();
    Fixture { project, bin, home }
}

fn run(f: &Fixture, args: &[&str]) -> Output {
    Command::new(common::MOAI_BIN)
        .clean_env()
        .args(args)
        .current_dir(f.project.path())
        .env("PATH", format!("{}:/usr/bin:/bin", f.bin.path().display()))
        .env("HOME", f.home.path())
        .env("USERPROFILE", f.home.path())
        .env("LANG", "C")
        .env("MOAI_TEST_HOST_MARKER", f.home.path().join("host-started"))
        .env("ANTHROPIC_AUTH_TOKEN", "moai-test-dummy-not-real")
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .output()
        .unwrap()
}

fn started(f: &Fixture) -> bool {
    f.home.path().join("host-started").exists()
}

fn out_text(out: &Output) -> String {
    format!(
        "{}{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    )
}

/// The bytes of the file a person keeps in the OMP home folder.
fn omp_file(f: &Fixture) -> Vec<u8> {
    fs::read(f.home.path().join(".omp/agent/settings.json")).unwrap()
}

/// The one line `moai omp` prints now (English; the terminal language is `C` here).
const OMP_REMOVED: &str =
    "moai omp was removed: MoAI supports Claude Code and GLM only (moai cc | moai glm)";

fn usage_envelope(message: &str) -> Value {
    json!({
        "protocol_version": 1,
        "operation": "cli",
        "ok": false,
        "error": {"kind": "usage", "message": message, "retryable": false}
    })
}

/// AC-1: `moai omp` is gone: the notice once on stderr, the usage envelope with the same sentence,
/// exit 2, nothing started, and the OMP home folder unchanged.
#[test]
fn moai_omp_says_it_was_removed() {
    let f = fixture();
    let before = omp_file(&f);
    for args in [
        &["omp"][..],
        &["omp", "-a"][..],
        &["omp", "-l", "--model", "x"][..],
        &["omp", "-a", "--dry-run"][..],
    ] {
        let out = run(&f, args);
        assert_eq!(out.status.code(), Some(2), "{args:?}: {}", out_text(&out));
        let v: Value = serde_json::from_slice(&out.stdout)
            .unwrap_or_else(|e| panic!("{args:?}: stdout is not the envelope ({e}): {out:?}"));
        assert_eq!(v, usage_envelope(OMP_REMOVED), "{args:?}");
        let stderr = String::from_utf8_lossy(&out.stderr);
        assert_eq!(stderr.matches(OMP_REMOVED).count(), 1, "{args:?}: {stderr}");
        assert!(!started(&f), "{args:?}: a host was started");
    }
    assert_eq!(omp_file(&f), before, "the OMP home file changed");
}

/// AC-2: `moai help omp` and `moai help codex` print what `moai omp` and `moai codex` print.
#[test]
fn help_omp_and_help_codex_print_the_removal_notice() {
    let f = fixture();
    for word in ["omp", "codex"] {
        let direct = run(&f, &[word]);
        let help = run(&f, &["help", word]);
        assert_eq!(
            direct.status.code(),
            Some(2),
            "{word}: {}",
            out_text(&direct)
        );
        assert_eq!(
            help.status.code(),
            direct.status.code(),
            "help {word}: {}",
            out_text(&help)
        );
        assert_eq!(
            String::from_utf8_lossy(&help.stdout),
            String::from_utf8_lossy(&direct.stdout),
            "help {word}: stdout"
        );
        assert_eq!(
            String::from_utf8_lossy(&help.stderr),
            String::from_utf8_lossy(&direct.stderr),
            "help {word}: stderr"
        );
    }
    assert!(!started(&f));
}

/// AC-3: the host word `omp` after `-a` and after `session --host` is refused exactly like an
/// unknown word, and the message lists `cc|glm` only.
#[test]
fn the_omp_host_word_is_refused_like_any_unknown_host() {
    let f = fixture();
    let nope = run(&f, &["-a", "nope", "--dry-run"]);
    let omp = run(&f, &["-a", "omp", "--dry-run"]);
    assert_ne!(nope.status.code(), Some(0), "-a nope: {nope:?}");
    assert_eq!(
        omp.status.code(),
        nope.status.code(),
        "-a omp: {}",
        out_text(&omp)
    );
    let text = out_text(&omp);
    assert!(text.contains("(cc|glm)"), "{text}");
    assert_eq!(
        text.replace("omp", "WORD"),
        out_text(&nope).replace("nope", "WORD"),
        "the refusal differs from the unknown-host refusal"
    );
    let session_nope = run(
        &f,
        &["session", "--host", "nope", "--name", "x", "--dry-run"],
    );
    let session_omp = run(
        &f,
        &["session", "--host", "omp", "--name", "x", "--dry-run"],
    );
    assert_ne!(session_nope.status.code(), Some(0), "{session_nope:?}");
    assert_eq!(
        session_omp.status.code(),
        session_nope.status.code(),
        "{}",
        out_text(&session_omp)
    );
    let text = out_text(&session_omp);
    assert!(!text.contains("|omp") && !text.contains("omp|"), "{text}");
    assert!(!started(&f), "a host was started");
}

const LAUNCHES: [&[&str]; 4] = [
    &["cc", "-l", "--dry-run"],
    &["cc", "-a", "--name", "agent-9", "--dry-run"],
    &["glm", "-l", "--dry-run"],
    &["glm", "-a", "--name", "agent-9", "--dry-run"],
];

/// The dry-run JSON of a launch, with the temporary HOME and project paths replaced.
fn dry_run(f: &Fixture, args: &[&str]) -> String {
    let out = run(f, args);
    assert!(out.status.success(), "{args:?}: {}", out_text(&out));
    let v: Value = serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|e| panic!("{args:?}: not JSON ({e}): {}", out_text(&out)));
    v.to_string()
        .replace(&f.home.path().display().to_string(), "<home>")
        .replace(&f.project.path().display().to_string(), "<project>")
        .replace(&f.bin.path().display().to_string(), "<bin>")
}

/// AC-4: the `cc` and `glm` launches name no OMP program, no OMP key variable and no
/// `/skill:moai` spelling in their command, environment or notes.
#[test]
fn cc_and_glm_launches_are_unchanged_and_name_no_omp() {
    let f = fixture();
    for args in LAUNCHES {
        let text = dry_run(&f, args);
        // Whole words only: "prompt" holds the letters too.
        let lower = text.to_ascii_lowercase();
        let named = lower
            .split(|c: char| !c.is_ascii_alphanumeric())
            .any(|word| word == "omp");
        assert!(!named, "{args:?}: {text}");
        assert!(!text.contains("/skill:moai"), "{args:?}: {text}");
        assert!(!text.contains("ZAI_API_KEY"), "{args:?}: {text}");
        assert!(!text.contains("Oh My Pi"), "{args:?}: {text}");
    }
    assert!(!started(&f));
}

fn save_launcher(f: &Fixture, body: &str) {
    let dir = moai_cli::home_root::path(f.home.path()).join("config");
    fs::create_dir_all(&dir).unwrap();
    fs::write(dir.join("launcher.yaml"), body).unwrap();
}

/// AC-6: a saved `launcher.models.omp` value of any shape leaves every `cc` and `glm` dry run as
/// it is without that value.
#[test]
fn a_saved_omp_value_of_any_shape_never_changes_a_cc_or_glm_launch() {
    let f = fixture();
    let plain: Vec<String> = LAUNCHES.iter().map(|a| dry_run(&f, a)).collect();
    for omp in [
        r#"{"leader":{"model":"zai/glm-5.3-flash","effort":"max"},"agent":{"model":"zai/glm-5.3","effort":"low"}}"#,
        r#""SECRET-OMP-MODEL""#,
        "7",
    ] {
        save_launcher(
            &f,
            &format!(r#"{{"launcher":{{"models":{{"omp":{omp}}}}}}}"#),
        );
        for (args, want) in LAUNCHES.iter().zip(&plain) {
            let got = dry_run(&f, args);
            assert_eq!(&got, want, "{omp}: {args:?}");
            assert!(!got.contains("SECRET-OMP-MODEL"), "{omp}: {args:?}: {got}");
        }
    }
    assert!(!started(&f));
}

fn repo(rel: &str) -> String {
    let path = Path::new(env!("CARGO_MANIFEST_DIR")).join(rel);
    fs::read_to_string(&path).unwrap_or_else(|e| panic!("{rel}: {e}"))
}

/// AC-8: the help goldens, the recorded usage error and the install texts name no `omp` launcher
/// and no `moai codex`. Named files only; no folder is walked.
#[test]
fn no_help_usage_or_install_text_names_omp_or_moai_codex() {
    let files = [
        "tests/fixtures/help/root.txt",
        "tests/fixtures/help/root-60.txt",
        "tests/fixtures/help/root-80.txt",
        "tests/fixtures/help/root-dumb.txt",
        "tests/fixtures/help/all.txt",
        "tests/fixtures/help/all-60.txt",
        "tests/fixtures/help/all-80.txt",
        "tests/fixtures/help/all-dumb.txt",
        "tests/fixtures/protocol/v1/current/01-usage-unknown-subcommand.json",
        "install/install.sh",
        "install/install.ps1",
        "install/README.md",
    ];
    let mut hits = Vec::new();
    for rel in files {
        for (n, line) in repo(rel).lines().enumerate() {
            let first = line.split_whitespace().next().unwrap_or("");
            let named = line.contains("moai omp")
                || line.contains("moai codex")
                || line.contains("|omp")
                || line.contains(", omp")
                || first == "omp";
            if named {
                hits.push(format!("{rel}:{}: {line}", n + 1));
            }
        }
    }
    assert!(hits.is_empty(), "{}", hits.join("\n"));
}
