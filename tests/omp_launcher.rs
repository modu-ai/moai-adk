//! SPEC-20261118-omp-launcher (card C-62 slice B; user instruction 2026-11-17): `moai omp
//! [-l|-a]` is the third launcher beside `cc` and `glm` (the fourth until Codex was withdrawn,
//! SPEC-20261121-codex-launcher-and-models-removed). It replaces the 2026-11-10
//! decision of Issue #2378 that the launcher knew three hosts (`tests/launcher_three_hosts.rs`
//! pinned that decision and is gone). Every case runs in a temporary git project with a temporary
//! HOME and a stub `omp` on PATH that leaves traces; no real host starts. `--print-command` is
//! another spelling of `--dry-run`.
#![cfg(unix)]
mod common;
use serde_json::Value;
use std::io::Write;
use std::os::unix::fs::PermissionsExt;
use std::{
    fs,
    path::Path,
    process::{Command, Output, Stdio},
};

const ENV_KEY: &str = "moai-test-env-key-not-real";
const VAULT_KEY: &str = "moai-test-vault-key-not-real";
const GLM_MODEL: &str = "zai/glm-5.3";

/// The stub host: records that it started, the value of `ZAI_API_KEY` (or `<unset>`) and its
/// arguments, one per line, in the folder `MOAI_TEST_MARK` names.
const STUB: &str = "#!/bin/sh
printf started > \"$MOAI_TEST_MARK/started\"
printf '%s' \"${ZAI_API_KEY-<unset>}\" > \"$MOAI_TEST_MARK/zai\"
printf '%s\\n' \"$@\" > \"$MOAI_TEST_MARK/args\"
env | cut -d= -f1 | sort > \"$MOAI_TEST_MARK/envnames\"
exit 0
";

struct Fixture {
    project: tempfile::TempDir,
    bin: tempfile::TempDir,
    home: tempfile::TempDir,
}

impl Fixture {
    fn new() -> Self {
        let project = tempfile::tempdir().unwrap();
        let ok = common::git::command()
            .args(["init", "-q"])
            .current_dir(project.path())
            .stdout(Stdio::null())
            .stderr(Stdio::null())
            .status()
            .unwrap()
            .success();
        assert!(ok);
        let bin = tempfile::tempdir().unwrap();
        let stub = bin.path().join("omp");
        fs::write(&stub, STUB).unwrap();
        fs::set_permissions(&stub, fs::Permissions::from_mode(0o755)).unwrap();
        Fixture {
            project,
            bin,
            home: tempfile::tempdir().unwrap(),
        }
    }

    fn command(&self, args: &[&str]) -> Command {
        let mut c = Command::new(crate::common::MOAI_BIN);
        c.args(args)
            .current_dir(self.project.path())
            .env(
                "PATH",
                format!("{}:/usr/bin:/bin", self.bin.path().display()),
            )
            .env("HOME", self.home.path())
            .env("MOAI_TEST_MARK", self.bin.path())
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_SYSTEM", "/dev/null")
            .env_remove("ANTHROPIC_AUTH_TOKEN")
            .env_remove("ZAI_API_KEY")
            .env_remove("CLAUDE_CODE_DISABLE_ADVISOR_TOOL")
            .env_remove(moai_cli::session::AGENT_SEQ_TEST_HOLD_VAR)
            .stdin(Stdio::null());
        c
    }

    fn run(&self, args: &[&str]) -> Output {
        self.command(args).output().unwrap()
    }

    /// The launcher with `--dry-run`: its JSON, which must have exited 0.
    fn dry(&self, args: &[&str]) -> Value {
        let mut all = args.to_vec();
        all.push("--dry-run");
        let out = self.run(&all);
        assert!(out.status.success(), "{args:?}: {out:?}");
        serde_json::from_slice(&out.stdout).unwrap()
    }

    fn started(&self) -> bool {
        self.bin.path().join("started").exists()
    }

    fn zai_seen_by_host(&self) -> String {
        fs::read_to_string(self.bin.path().join("zai")).unwrap()
    }

    /// Stores `value` as the shared glm key in the temporary HOME with the binary's own `key-set`.
    fn store_shared_key(&self, value: &str) {
        let mut child = self
            .command(&["key-set", "glm", "--shared"])
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .spawn()
            .unwrap();
        child
            .stdin
            .take()
            .unwrap()
            .write_all(value.as_bytes())
            .unwrap();
        let out = child.wait_with_output().unwrap();
        let v: Value = serde_json::from_slice(&out.stdout).unwrap();
        assert_eq!(v["ok"], true, "{v}");
    }

    /// A legacy key file: the glm launcher never reads it and stops with `legacy_key_only`.
    fn plant_legacy_key_file(&self) {
        fs::create_dir_all(self.home.path().join(".moai")).unwrap();
        fs::write(
            self.home.path().join(".moai/.env.glm"),
            "GLM_API_KEY=moai-test-legacy-key-not-real\n",
        )
        .unwrap();
    }

    fn number_live(&self, n: u32) -> bool {
        self.project
            .path()
            .join(format!(".git/.moai-cli/sessions/agent-{n}.live"))
            .exists()
    }
}

fn argv(v: &Value) -> Vec<String> {
    serde_json::from_value(v["argv"].clone()).unwrap()
}

fn notes(v: &Value) -> Vec<String> {
    serde_json::from_value(v["notes"].clone()).unwrap()
}

fn env_keys(v: &Value) -> Vec<String> {
    serde_json::from_value(v["env_keys"].clone()).unwrap()
}

fn text(out: &Output) -> String {
    format!(
        "{}{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    )
}

/// AC-1: `moai omp`, `-l`, `-a` and the legacy `-a omp` compose a command that starts with `omp`.
#[test]
fn omp_roles_compose_an_omp_command() {
    let f = Fixture::new();
    let plain = f.dry(&["omp"]);
    assert_eq!(plain["host"], "omp");
    assert_eq!(plain["role"], Value::Null);
    assert_eq!(argv(&plain), ["omp"]);
    assert_eq!(plain["executed"], false);

    let leader = f.dry(&["omp", "-l"]);
    assert_eq!(leader["role"], "leader");
    assert_eq!(leader["name"], "leader");
    assert_eq!(argv(&leader)[0], "omp");

    let agent = f.dry(&["omp", "-a"]);
    assert_eq!(agent["role"], "agent");
    assert_eq!(agent["name"], "agent-1");
    assert_eq!(argv(&agent)[0], "omp");

    let named = f.dry(&["omp", "-a", "--name", "lane-x"]);
    assert_eq!(named["name"], "lane-x");

    let legacy = f.dry(&["-a", "omp"]);
    assert_eq!(legacy["host"], "omp");
    assert_eq!(argv(&legacy)[0], "omp");
}

/// AC-2: `--print-command` is `--dry-run` on every host word, and starts nothing.
#[test]
fn print_command_is_dry_run_on_every_host() {
    let f = Fixture::new();
    for host in ["cc", "glm", "omp"] {
        let run = |flag: &str| {
            let mut c = f.command(&[host, "-l", flag]);
            c.env("ANTHROPIC_AUTH_TOKEN", ENV_KEY);
            let out = c.output().unwrap();
            assert!(out.status.success(), "{host} {flag}: {out:?}");
            serde_json::from_slice::<Value>(&out.stdout).unwrap()
        };
        let a = run("--dry-run");
        let b = run("--print-command");
        assert_eq!(a, b, "{host}");
        assert_eq!(b["executed"], false, "{host}");
    }
    assert!(!f.started());
}

#[test]
fn omp_dry_run_starts_nothing() {
    let f = Fixture::new();
    f.dry(&["omp", "-a"]);
    f.dry(&["omp", "-l", "--model", GLM_MODEL]);
    assert!(!f.started(), "a dry run started the host");
    assert!(!f.number_live(1), "a dry run locked an agent number");
}

/// AC-3: `--model` and `--effort` go to `--model` and `--thinking`.
#[test]
fn omp_model_and_thinking() {
    let f = Fixture::new();
    let v = f.dry(&["omp", "--model", "opus", "--effort", "high"]);
    assert_eq!(argv(&v), ["omp", "--model", "opus", "--thinking", "high"]);
    for effort in ["low", "medium", "high", "xhigh", "max"] {
        let v = f.dry(&["omp", "--effort", effort]);
        assert_eq!(argv(&v), ["omp", "--thinking", effort], "{effort}");
    }
    let v = f.dry(&["omp", "--model", "opus"]);
    assert_eq!(argv(&v), ["omp", "--model", "opus"]);
}

#[test]
fn omp_unknown_effort_is_a_note() {
    let f = Fixture::new();
    let v = f.dry(&["omp", "--effort", "ultracode"]);
    assert!(!argv(&v).contains(&"--thinking".to_string()), "{v}");
    assert!(
        notes(&v)
            .iter()
            .any(|n| n.contains("--effort was not passed to omp")),
        "{v}"
    );
}

/// AC-4: the access of the role rules (user decision 2026-11-10, #2359) in OMP's words.
#[test]
fn omp_approval_modes() {
    let f = Fixture::new();
    let modes = |args: &[&str]| -> Vec<String> {
        let a = argv(&f.dry(args));
        a.iter()
            .position(|w| w == "--approval-mode")
            .map(|i| a[i + 1..].iter().take(1).cloned().collect())
            .unwrap_or_default()
    };
    assert_eq!(modes(&["omp"]), Vec::<String>::new(), "no role: no flag");
    assert_eq!(modes(&["omp", "-l"]), ["yolo"]);
    assert_eq!(modes(&["omp", "-a"]), ["yolo"]);
    assert_eq!(
        modes(&["omp", "-l", "--permissions"]),
        ["always-ask-cached"]
    );
    assert_eq!(
        modes(&["omp", "-a", "--permissions"]),
        ["always-ask-cached"]
    );
    assert_eq!(modes(&["omp", "--permissions"]), Vec::<String>::new());
}

/// AC-5: no role is entered by a first prompt; the notes name the lines typed inside.
#[test]
fn omp_has_no_role_prompt_and_names_the_entry_lines() {
    let f = Fixture::new();
    let leader = f.dry(&["omp", "-l"]);
    assert_eq!(argv(&leader), ["omp", "--approval-mode", "yolo"]);
    assert!(
        notes(&leader)
            .iter()
            .any(|n| n.contains("`/skill:moai --mode-leader`")),
        "{leader}"
    );
    let agent = f.dry(&["omp", "-a"]);
    assert_eq!(argv(&agent), ["omp", "--approval-mode", "yolo"]);
    let agent_notes = notes(&agent);
    assert!(
        agent_notes
            .iter()
            .any(|n| n.contains("`/skill:moai --mode-agent`")),
        "{agent}"
    );
    assert!(
        agent_notes.iter().all(|n| !n.contains("`/moai --mode")),
        "{agent}"
    );
}

#[test]
fn omp_advisor_is_a_note() {
    let f = Fixture::new();
    let v = f.dry(&["omp", "--advisor", "opus"]);
    assert!(!argv(&v).contains(&"--advisor".to_string()), "{v}");
    assert!(
        notes(&v).iter().any(|n| n.contains("not passed to omp")),
        "{v}"
    );
}

/// AC-6: the glm key reaches the OMP process for a z.ai model only, as `ZAI_API_KEY`.
#[test]
fn omp_key_reaches_the_process_for_a_glm_model_only() {
    let f = Fixture::new();
    f.store_shared_key(VAULT_KEY);
    let out = f.run(&["omp", "-l", "--model", GLM_MODEL]);
    assert!(out.status.success(), "{out:?}");
    assert!(f.started());
    assert_eq!(
        f.zai_seen_by_host(),
        VAULT_KEY,
        "z.ai model, key from vault"
    );

    let f2 = Fixture::new();
    f2.store_shared_key(VAULT_KEY);
    let out = f2.run(&["omp", "-l", "--model", "opus"]);
    assert!(out.status.success(), "{out:?}");
    assert_eq!(f2.zai_seen_by_host(), "<unset>", "not a z.ai model");
    let out = f2.run(&["omp", "-l"]);
    assert!(out.status.success(), "{out:?}");
    assert_eq!(f2.zai_seen_by_host(), "<unset>", "no model named");

    // The variable wins over the vault, as for the glm launcher; a `glm` name is a z.ai model.
    let f3 = Fixture::new();
    f3.store_shared_key(VAULT_KEY);
    let mut c = f3.command(&["omp", "-l", "--model", "glm-5.3-flash"]);
    c.env("ANTHROPIC_AUTH_TOKEN", ENV_KEY);
    let out = c.output().unwrap();
    assert!(out.status.success(), "{out:?}");
    assert_eq!(f3.zai_seen_by_host(), ENV_KEY);
}

/// AC-6: the key goes to the OMP process only. A claude process started with a z.ai model name
/// never sees `ZAI_API_KEY` (audit #3211 N2).
#[test]
fn the_key_is_given_to_omp_only() {
    for host in ["claude"] {
        let f = Fixture::new();
        f.store_shared_key(VAULT_KEY);
        let stub = f.bin.path().join(host);
        fs::write(&stub, STUB).unwrap();
        fs::set_permissions(&stub, fs::Permissions::from_mode(0o755)).unwrap();
        let mut c = f.command(&["cc", "-l", "--model", GLM_MODEL]);
        c.env("ANTHROPIC_AUTH_TOKEN", ENV_KEY);
        let out = c.output().unwrap();
        assert!(out.status.success(), "{host}: {out:?}");
        assert!(f.started(), "{host}");
        assert_eq!(
            f.zai_seen_by_host(),
            "<unset>",
            "{host} must not get the key"
        );
    }
}

/// AC-6: for a z.ai model the OMP process gets exactly one more variable than for any other
/// model, `ZAI_API_KEY` (audit #3211 N3).
#[test]
fn omp_gets_one_more_variable_for_a_zai_model() {
    let f = Fixture::new();
    f.store_shared_key(VAULT_KEY);
    let names = |f: &Fixture| -> Vec<String> {
        fs::read_to_string(f.bin.path().join("envnames"))
            .unwrap()
            .lines()
            .map(str::to_string)
            .collect()
    };
    let out = f.run(&["omp", "-l", "--model", "opus"]);
    assert!(out.status.success(), "{out:?}");
    let plain = names(&f);
    assert!(!plain.contains(&"ZAI_API_KEY".to_string()), "{plain:?}");
    let out = f.run(&["omp", "-l", "--model", GLM_MODEL]);
    assert!(out.status.success(), "{out:?}");
    let with_key = names(&f);
    let extra: Vec<&String> = with_key.iter().filter(|n| !plain.contains(n)).collect();
    assert_eq!(extra, ["ZAI_API_KEY"], "{with_key:?}");
    assert_eq!(with_key.len(), plain.len() + 1, "{with_key:?}");
}

/// AC-6: the value is in no command, JSON, note or stream; the JSON lists the name only.
#[test]
fn omp_key_never_appears_in_output() {
    let f = Fixture::new();
    f.store_shared_key(VAULT_KEY);
    let dry = f.run(&["omp", "-a", "--model", GLM_MODEL, "--dry-run"]);
    assert!(dry.status.success(), "{dry:?}");
    assert!(!text(&dry).contains(VAULT_KEY), "{}", text(&dry));
    let v: Value = serde_json::from_slice(&dry.stdout).unwrap();
    assert!(env_keys(&v).contains(&"ZAI_API_KEY".to_string()), "{v}");
    assert!(!argv(&v).iter().any(|w| w.contains(VAULT_KEY)));
    assert_eq!(v["key_source"], "vault:shared:glm", "{v}");

    let run = f.run(&["omp", "-a", "--model", GLM_MODEL]);
    assert!(run.status.success(), "{run:?}");
    assert!(!text(&run).contains(VAULT_KEY), "{}", text(&run));
    let args = fs::read_to_string(f.bin.path().join("args")).unwrap();
    assert!(!args.contains(VAULT_KEY), "{args}");

    let mut c = f.command(&["omp", "-a", "--model", GLM_MODEL]);
    c.env("ANTHROPIC_AUTH_TOKEN", ENV_KEY);
    let out = c.output().unwrap();
    assert!(!text(&out).contains(ENV_KEY), "{}", text(&out));
}

/// AC-7: a vault that cannot give the key stops an OMP launch the way it stops a glm launch.
#[test]
fn omp_blocked_vault_stops_like_glm() {
    let f = Fixture::new();
    f.plant_legacy_key_file();
    let glm = f.run(&["glm", "-a", "--dry-run"]);
    assert!(!glm.status.success(), "{glm:?}");
    let blocked = |out: &Output| -> String {
        String::from_utf8_lossy(&out.stderr)
            .lines()
            .find(|l| l.contains("blocked:"))
            .unwrap_or_default()
            .to_string()
    };
    let want = blocked(&glm);
    assert!(want.contains("legacy_key_only"), "{want}");

    let omp = f.run(&["omp", "-a", "--model", GLM_MODEL]);
    assert!(!omp.status.success(), "{omp:?}");
    assert_eq!(blocked(&omp), want);
    assert!(!f.started(), "a blocked launch started the host");
    assert!(!f.number_live(1), "a blocked launch took a number");

    // An OMP launch that does not need the key is not blocked by it.
    let plain = f.run(&["omp", "-a", "--model", "opus", "--dry-run"]);
    assert!(plain.status.success(), "{plain:?}");
}

#[test]
fn omp_without_a_key_goes_on_with_a_note() {
    let f = Fixture::new();
    let v = f.dry(&["omp", "-l", "--model", GLM_MODEL]);
    assert!(!env_keys(&v).contains(&"ZAI_API_KEY".to_string()), "{v}");
    assert!(
        notes(&v)
            .iter()
            .any(|n| n.contains("OMP uses its own login")),
        "{v}"
    );
    let out = f.run(&["omp", "-l", "--model", GLM_MODEL]);
    assert!(out.status.success(), "{out:?}");
    assert_eq!(f.zai_seen_by_host(), "<unset>");
}

/// AC-8: `-a` takes the next `agent-N` while the host runs, and a refusal takes none.
#[test]
fn omp_agent_takes_a_number_and_a_refusal_takes_none() {
    let f = Fixture::new();
    let out = f.run(&["omp", "-a"]);
    assert!(out.status.success(), "{out:?}");
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["name"], "agent-1", "{v}");
    assert_eq!(v["executed"], true);

    // A refused launch (a blocked vault) takes no number: a fresh project has no `.live` file
    // afterwards, and the next accepted launch is `agent-1`.
    let g = Fixture::new();
    g.plant_legacy_key_file();
    let refused = g.run(&["omp", "-a", "--model", GLM_MODEL]);
    assert!(!refused.status.success(), "{refused:?}");
    assert!(!g.number_live(1), "a refused launch took a number");
    assert!(!g.started());
    let accepted = g.run(&["omp", "-a"]);
    let v: Value = serde_json::from_slice(&accepted.stdout).unwrap();
    assert_eq!(v["name"], "agent-1", "{v}");
}

/// AC-9: the help, the command list, the usage line and the unknown-host error name `omp`.
#[test]
fn omp_is_named_in_help_usage_and_errors() {
    let f = Fixture::new();
    let words = |s: &str| -> Vec<String> {
        s.split(|c: char| !c.is_ascii_alphanumeric())
            .filter(|w| !w.is_empty())
            .map(str::to_ascii_lowercase)
            .collect()
    };
    let root = f.run(&["--help"]);
    assert!(root.status.success());
    let root = String::from_utf8_lossy(&root.stdout).to_string();
    for host in ["cc", "glm", "omp"] {
        assert!(
            words(&root).contains(&host.to_string()),
            "{host} in:\n{root}"
        );
    }
    let all = f.run(&["help", "--all"]);
    assert!(all.status.success());
    let all = String::from_utf8_lossy(&all.stdout).to_string();
    let commands: Vec<&str> = all
        .lines()
        .filter_map(|l| l.strip_prefix("  "))
        .filter_map(|l| l.split_whitespace().next())
        .collect();
    for host in ["cc", "glm", "omp"] {
        assert!(commands.contains(&host), "{host} missing from:\n{all}");
    }
    assert!(
        !commands.contains(&"codex"),
        "codex is not a launcher word:\n{all}"
    );

    let bare = f.run(&[]);
    assert!(text(&bare).contains("cc|glm|omp"), "{}", text(&bare));
    let unknown = f.run(&["-a", "nope", "--dry-run"]);
    assert!(
        text(&unknown).contains("(cc|glm|omp)"),
        "{}",
        text(&unknown)
    );

    let help = f.run(&["omp", "--help"]);
    assert!(help.status.success(), "{help:?}");
    let help = String::from_utf8_lossy(&help.stdout)
        .split_whitespace()
        .collect::<Vec<_>>()
        .join(" ");
    assert!(help.contains("/skill:moai --mode-agent"), "{help}");
    assert!(help.contains("--print-command"), "{help}");
}

/// AC-11: the support document no longer records OMP as unsupported (the 2026-11-10 line is
/// replaced by the instruction of 2026-11-17).
#[test]
fn the_support_matrix_no_longer_says_omp_is_unsupported() {
    let root = Path::new(env!("CARGO_MANIFEST_DIR"));
    let matrix = fs::read_to_string(root.join("docs/host-support-matrix.md")).unwrap();
    assert!(
        !matrix.contains("지원하지 않음(2026-11-10 결정"),
        "old line is still there"
    );
    assert!(matrix.contains("2026-11-17"));
}
