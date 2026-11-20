//! SPEC-20261119-launcher-reads-models (card C-47 slice 8, criteria 5 and 7): the launchers read the
//! model and effort that `moai update` saved under `launcher.models.<tool>.<role>` in the user
//! launcher settings file. The `glm` host follows in SPEC-20261119-glm-launcher-settings-receipt
//! (criterion 5, glm part): its saved id carries the `zai/` provider prefix of the shared table, and
//! the host takes the bare z.ai model name. Every case uses `--dry-run` with a synthetic HOME and a
//! temporary git project; no host is started and no real settings file is read.

#![cfg(unix)]

mod common;

use serde_json::Value;
use std::fs;
use std::os::unix::fs::PermissionsExt;
use std::path::Path;
use std::process::{Command, Output};

/// Every saved choice the support table accepts, one per tool and role.
const SAVED: &str = r#"{"launcher":{"future_key":"SENTINEL-SAVED-TEXT","models":{
    "claude":{"leader":{"model":"claude-opus-5-5","effort":"high"},
              "agent":{"model":"claude-haiku-5-5","effort":"low"}},
    "codex":{"leader":{"model":"gpt-6.2-sol","effort":"max"},
             "agent":{"model":"gpt-6-luna","effort":"low"}},
    "glm":{"leader":{"model":"zai/glm-5.3-flash","effort":"max"},
           "agent":{"model":"zai/glm-5.3","effort":"low"}}}}}"#;

/// The same tools and roles with other supported values, for the rewrite test.
const SAVED_OTHER: &str = r#"{"launcher":{"models":{
    "claude":{"leader":{"model":"claude-sonnet-5-5","effort":"medium"}}}}}"#;

struct Fixture {
    project: tempfile::TempDir,
    bin: tempfile::TempDir,
    home: tempfile::TempDir,
}

impl Fixture {
    fn new() -> Self {
        let project = tempfile::tempdir().unwrap();
        assert!(common::git::command()
            .args(["init", "-q"])
            .current_dir(project.path())
            .status()
            .unwrap()
            .success());
        let bin = tempfile::tempdir().unwrap();
        let stub = bin.path().join("omp");
        fs::write(&stub, "#!/bin/sh\nexit 0\n").unwrap();
        fs::set_permissions(&stub, fs::Permissions::from_mode(0o755)).unwrap();
        let home = tempfile::tempdir().unwrap();
        // The skill copies of both hosts, so a launch prints no missing-skill warning (that line
        // names the HOME folder, and the outputs compared below come from different HOMEs).
        for dir in [".claude/skills/moai", ".agents/skills/moai"] {
            fs::create_dir_all(home.path().join(dir)).unwrap();
            fs::write(home.path().join(dir).join("SKILL.md"), "skill\n").unwrap();
        }
        Fixture { project, bin, home }
    }

    fn config(&self, text: &str) {
        let dir = self.home.path().join(".moai-cli/config");
        fs::create_dir_all(&dir).unwrap();
        fs::write(dir.join("launcher.yaml"), text).unwrap();
    }

    fn run(&self, args: &[&str]) -> Output {
        run_in(self.project.path(), self.bin.path(), self.home.path(), args)
    }

    /// The launcher with `--dry-run`: its JSON, which must have exited 0.
    fn dry(&self, args: &[&str]) -> Value {
        let out = self.run(args);
        assert!(
            out.status.success(),
            "{args:?}: {}",
            String::from_utf8_lossy(&out.stderr)
        );
        serde_json::from_slice(&out.stdout).unwrap()
    }
}

fn run_in(project: &Path, bin: &Path, home: &Path, args: &[&str]) -> Output {
    Command::new(crate::common::MOAI_BIN)
        .args(args)
        .arg("--dry-run")
        .current_dir(project)
        .env("PATH", format!("{}:/usr/bin:/bin", bin.display()))
        .env("HOME", home)
        .env("USERPROFILE", home)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .env_remove("ZAI_API_KEY")
        .env_remove("CLAUDE_CODE_DISABLE_ADVISOR_TOOL")
        .env("ANTHROPIC_AUTH_TOKEN", "moai-test-dummy-not-real")
        .output()
        .unwrap()
}

fn argv(v: &Value) -> Vec<String> {
    serde_json::from_value(v["argv"].clone()).unwrap()
}

fn notes(v: &Value) -> String {
    v["notes"].to_string()
}

/// True when `flag` is followed directly by `value` in the argv.
fn has_pair(argv: &[String], flag: &str, value: &str) -> bool {
    argv.windows(2).any(|w| w[0] == flag && w[1] == value)
}

fn effort_arg(e: &str) -> String {
    format!("model_reasoning_effort=\"{e}\"")
}

/// `(launcher word, role flag, model flag, model, effort pair)` for the three hosts and two roles
/// of `SAVED`: the flag each host takes the model in and the effort in.
fn cases() -> Vec<(
    &'static str,
    &'static str,
    &'static str,
    &'static str,
    [String; 2],
)> {
    vec![
        (
            "cc",
            "-l",
            "--model",
            "claude-opus-5-5",
            ["--effort".into(), "high".into()],
        ),
        (
            "cc",
            "-a",
            "--model",
            "claude-haiku-5-5",
            ["--effort".into(), "low".into()],
        ),
        (
            "codex",
            "-l",
            "-m",
            "gpt-6.2-sol",
            ["-c".into(), effort_arg("max")],
        ),
        (
            "codex",
            "-a",
            "-m",
            "gpt-6-luna",
            ["-c".into(), effort_arg("low")],
        ),
        (
            "glm",
            "-l",
            "--model",
            "glm-5.3-flash",
            ["--effort".into(), "max".into()],
        ),
        (
            "glm",
            "-a",
            "--model",
            "glm-5.3",
            ["--effort".into(), "low".into()],
        ),
    ]
}

/// AC-1: a saved choice reaches the argv in each host's own flags, for every host and role.
#[test]
fn saved_values_reach_the_argv_of_each_host_and_role() {
    let f = Fixture::new();
    f.config(SAVED);
    for (word, role, model_flag, model, effort) in cases() {
        let v = f.dry(&[word, role]);
        let a = argv(&v);
        assert!(has_pair(&a, model_flag, model), "{word} {role}: {a:?}");
        assert!(has_pair(&a, &effort[0], &effort[1]), "{word} {role}: {a:?}");
    }
}

/// AC-2: an explicit `--model` replaces only the model, an explicit `--effort` only the effort.
#[test]
fn explicit_flags_win_per_field() {
    let f = Fixture::new();
    f.config(SAVED);
    // claude leader: saved claude-opus-5-5 / high.
    let a = argv(&f.dry(&["cc", "-l", "--model", "claude-sonnet-5-5"]));
    assert!(has_pair(&a, "--model", "claude-sonnet-5-5"), "{a:?}");
    assert!(has_pair(&a, "--effort", "high"), "{a:?}");
    let a = argv(&f.dry(&["cc", "-l", "--effort", "medium"]));
    assert!(has_pair(&a, "--model", "claude-opus-5-5"), "{a:?}");
    assert!(has_pair(&a, "--effort", "medium"), "{a:?}");
    // codex agent: saved gpt-6-luna / low.
    let a = argv(&f.dry(&["codex", "-a", "--effort", "high"]));
    assert!(has_pair(&a, "-m", "gpt-6-luna"), "{a:?}");
    assert!(has_pair(&a, "-c", &effort_arg("high")), "{a:?}");
    assert!(!has_pair(&a, "-c", &effort_arg("low")), "{a:?}");
    let a = argv(&f.dry(&["codex", "-a", "--model", "gpt-6.1-sol"]));
    assert!(has_pair(&a, "-m", "gpt-6.1-sol"), "{a:?}");
    assert!(has_pair(&a, "-c", &effort_arg("low")), "{a:?}");
    // glm leader: saved zai/glm-5.3-flash / max; an explicit alias stays as written (the zai/
    // prefix strip applies to the saved value only, and the alias still reaches the z.ai model
    // through the environment mapping).
    let a = argv(&f.dry(&["glm", "-l", "--model", "opus"]));
    assert!(has_pair(&a, "--model", "opus"), "{a:?}");
    assert!(has_pair(&a, "--effort", "max"), "{a:?}");
    // A prefixed explicit model also stays as written: the strip is a translation of the saved
    // table's OMP form, not input handling (audit N2 — a pin with a prefix on the flag).
    let a = argv(&f.dry(&["glm", "-l", "--model", "zai/glm-5.3"]));
    assert!(has_pair(&a, "--model", "zai/glm-5.3"), "{a:?}");
    assert!(has_pair(&a, "--effort", "max"), "{a:?}");
    // both given: nothing of the file is used and no saved-value note is added.
    let v = f.dry(&[
        "cc",
        "-l",
        "--model",
        "claude-sonnet-5-5",
        "--effort",
        "medium",
    ]);
    let a = argv(&v);
    assert!(has_pair(&a, "--model", "claude-sonnet-5-5"), "{a:?}");
    assert!(has_pair(&a, "--effort", "medium"), "{a:?}");
    assert!(!notes(&v).contains("launcher.models"), "{}", notes(&v));
}

/// AC-3 (REQ-4): with no saved choice for that host and role the launch is what it was.
#[test]
fn without_a_saved_choice_nothing_changes() {
    let f = Fixture::new();
    // No file at all.
    let a = argv(&f.dry(&["cc", "-l"]));
    assert!(has_pair(&a, "--model", "opus"), "{a:?}");
    assert!(has_pair(&a, "--effort", "medium"), "{a:?}");
    let a = argv(&f.dry(&["codex", "-l"]));
    assert!(!a.iter().any(|x| x == "-m" || x == "-c"), "{a:?}");
    // A file that saves only codex: the claude and glm launches stay as they are (a glm role
    // session without a saved glm entry keeps the alias default of `claude_argv`, D-4).
    let bare = (argv(&f.dry(&["cc", "-a"])), argv(&f.dry(&["glm", "-l"])));
    f.config(
        r#"{"launcher":{"models":{"codex":{"leader":{"model":"gpt-6.1-sol","effort":"high"}}}}}"#,
    );
    assert_eq!(argv(&f.dry(&["cc", "-a"])), bare.0);
    assert_eq!(argv(&f.dry(&["glm", "-l"])), bare.1);
    // The codex agent has no saved entry either: nothing is passed for it.
    let a = argv(&f.dry(&["codex", "-a"]));
    assert!(!a.iter().any(|x| x == "-m" || x == "-c"), "{a:?}");
}

/// AC-3 (REQ-4 of the glm SPEC): a roleless session — `glm` included — ignores the saved choices;
/// the glm role session reads them now (SPEC-20261119-glm-launcher-settings-receipt).
#[test]
fn a_roleless_session_ignores_saved_choices() {
    let with = Fixture::new();
    with.config(SAVED);
    let without = Fixture::new();
    let launches: [&[&str]; 3] = [&["cc"], &["codex"], &["glm"]];
    for args in launches {
        // Both homes launch in the same project: a note names the project folder.
        let launch = |home: &Path| {
            let out = run_in(with.project.path(), with.bin.path(), home, args);
            assert!(
                out.status.success(),
                "{args:?}: {}",
                String::from_utf8_lossy(&out.stderr)
            );
            serde_json::from_slice::<Value>(&out.stdout).unwrap()
        };
        let a = launch(with.home.path());
        let b = launch(without.home.path());
        assert_eq!(argv(&a), argv(&b), "{args:?}");
        assert_eq!(a["env_keys"], b["env_keys"], "{args:?}");
        assert_eq!(a["notes"], b["notes"], "{args:?}");
    }
}

/// AC-4: the file is read at every launch, so a rewritten file is used by the next one.
#[test]
fn a_rewritten_file_is_used_by_the_next_launch() {
    let f = Fixture::new();
    f.config(SAVED);
    let first = argv(&f.dry(&["cc", "-l"]));
    assert!(has_pair(&first, "--model", "claude-opus-5-5"), "{first:?}");
    assert!(has_pair(&first, "--effort", "high"), "{first:?}");
    f.config(SAVED_OTHER);
    let second = argv(&f.dry(&["cc", "-l"]));
    assert!(
        has_pair(&second, "--model", "claude-sonnet-5-5"),
        "{second:?}"
    );
    assert!(has_pair(&second, "--effort", "medium"), "{second:?}");
    // Removing the file puts the launch back to its defaults without another step.
    fs::remove_file(f.home.path().join(".moai-cli/config/launcher.yaml")).unwrap();
    let third = argv(&f.dry(&["cc", "-l"]));
    assert!(has_pair(&third, "--model", "opus"), "{third:?}");
    assert!(has_pair(&third, "--effort", "medium"), "{third:?}");
}

/// AC-5: a saved entry outside the support table refuses a role launch on claude, codex, omp and
/// glm (the entry of another tool counts: the whole file is checked), names key paths only, and
/// does not refuse a roleless launch.
#[test]
fn an_unsupported_saved_entry_refuses_a_role_launch() {
    // The file and the key path the refusal has to name (audit N5 of the first head).
    let bad_files = [
        // An unknown model for the tool.
        (
            r#"{"launcher":{"models":{"claude":{"leader":{"model":"SECRET-unknown-model","effort":"high"}}}}}"#,
            "launcher.models.claude.leader",
        ),
        // An effort the model does not offer (the glm models take low, high and max only).
        (
            // A tool this version does not know any more (the dropped OMP).
            r#"{"launcher":{"models":{"omp":{"agent":{"model":"zai/glm-5.3","effort":"medium"}}}}}"#,
            "launcher.models.omp.agent",
        ),
        // The bad entry belongs to another tool than the one launched below.
        (
            r#"{"launcher":{"models":{"codex":{"agent":{"model":"SECRET-other-tool","effort":"low"}}}}}"#,
            "launcher.models.codex.agent",
        ),
    ];
    for (bad, key_path) in bad_files {
        let f = Fixture::new();
        f.config(bad);
        for args in [["cc", "-l"], ["codex", "-a"], ["glm", "-l"]] {
            let out = f.run(&args);
            assert_eq!(out.status.code(), Some(1), "{args:?} {bad}");
            assert!(out.stdout.is_empty(), "{args:?} {bad}");
            let err = String::from_utf8_lossy(&out.stderr).to_string();
            assert!(!err.contains("SECRET"), "{args:?} {bad}: {err}");
            assert!(err.contains(key_path), "{args:?} {bad}: {err}");
        }
        for args in [vec!["cc"], vec!["glm"]] {
            let out = f.run(&args);
            assert!(
                out.status.success(),
                "{args:?} {bad}: {}",
                String::from_utf8_lossy(&out.stderr)
            );
        }
    }
}

/// AC-6: a saved value adds a note with its key path and the output holds no other saved text.
#[test]
fn a_saved_value_adds_a_note_with_its_key_path() {
    let f = Fixture::new();
    f.config(SAVED);
    for (word, role, _, _, _) in cases() {
        let tool = if word == "cc" { "claude" } else { word };
        let role_name = if role == "-l" { "leader" } else { "agent" };
        let key = format!("launcher.models.{tool}.{role_name}");
        let out = f.run(&[word, role]);
        assert!(out.status.success(), "{word} {role}");
        let text = String::from_utf8_lossy(&out.stdout).to_string();
        let v: Value = serde_json::from_str(&text).unwrap();
        assert!(notes(&v).contains(&key), "{word} {role}: {}", notes(&v));
        assert!(
            !text.contains("SENTINEL-SAVED-TEXT"),
            "{word} {role}: {text}"
        );
    }
}

/// AC-6 (REQ-7, the start path): a saved glm choice reaches the started host with the saved model
/// and effort in its argv and the z.ai environment unchanged by the choice (endpoint, alias
/// mapping, the glm token). A stub `claude` records the arguments and the environment it was
/// started with.
#[test]
fn a_saved_glm_model_reaches_the_started_host() {
    let f = Fixture::new();
    f.config(SAVED);
    let stub = f.bin.path().join("claude");
    fs::write(
        &stub,
        "#!/bin/sh\nprintf '%s' \"${ANTHROPIC_AUTH_TOKEN-<unset>}\" > \"$MOAI_TEST_MARK/token\"\nprintf '%s' \"${ANTHROPIC_DEFAULT_OPUS_MODEL-<unset>}\" > \"$MOAI_TEST_MARK/opus\"\nprintf '%s' \"${ANTHROPIC_BASE_URL-<unset>}\" > \"$MOAI_TEST_MARK/base\"\nprintf '%s\\n' \"$@\" > \"$MOAI_TEST_MARK/args\"\nexit 0\n",
    )
    .unwrap();
    fs::set_permissions(&stub, fs::Permissions::from_mode(0o755)).unwrap();
    let out = Command::new(crate::common::MOAI_BIN)
        .args(["glm", "-l"])
        .current_dir(f.project.path())
        .env("PATH", format!("{}:/usr/bin:/bin", f.bin.path().display()))
        .env("HOME", f.home.path())
        .env("USERPROFILE", f.home.path())
        .env("MOAI_TEST_MARK", f.bin.path())
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .env_remove("ZAI_API_KEY")
        .env("ANTHROPIC_AUTH_TOKEN", "moai-test-token-not-real")
        .stdin(std::process::Stdio::null())
        .output()
        .unwrap();
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    // The environment is the z.ai mapping of an unchanged launch: the alias still resolves to
    // glm-5.3 even though the saved choice names the flash model, and the token is carried under
    // its two names.
    let seen_token = fs::read_to_string(f.bin.path().join("token")).unwrap();
    assert_eq!(seen_token, "moai-test-token-not-real");
    let seen_opus = fs::read_to_string(f.bin.path().join("opus")).unwrap();
    assert_eq!(seen_opus, "glm-5.3");
    let seen_base = fs::read_to_string(f.bin.path().join("base")).unwrap();
    assert_eq!(seen_base, "https://api.z.ai/api/anthropic");
    // The saved leader choice (zai/glm-5.3-flash / max) reaches the argv as the bare z.ai id.
    let args = fs::read_to_string(f.bin.path().join("args")).unwrap();
    let args: Vec<&str> = args.lines().collect();
    assert!(
        args.windows(2).any(|w| w == ["--model", "glm-5.3-flash"]),
        "{args:?}"
    );
    assert!(
        args.windows(2).any(|w| w == ["--effort", "max"]),
        "{args:?}"
    );
}
