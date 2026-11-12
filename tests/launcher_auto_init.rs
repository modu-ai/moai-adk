//! #2481 (user decisions 6810797000 and 6810818827, 2026-11-12): the first `cc|glm|codex`
//! launch in a folder that is not a git repository initializes it without asking — `git init`,
//! then the MoAI project settings (`.moai-cli/project.yaml`) and `AGENTS.md` through the
//! existing `project-apply` and `generation-apply` code — and says so on one stderr line. Files
//! that already exist are never touched. HOME itself and the file-system root are never
//! initialized (guidance, exit 1). Without git on PATH the launcher prints the install guidance
//! for this OS (exit 1). `--dry-run` creates nothing and reports what it would initialize.
//! Every launch runs in a temporary folder with a temporary HOME and stub `claude` / `codex`
//! executables on PATH; no real host starts.
#![cfg(unix)]
mod common;
use serde_json::Value;
use std::os::unix::fs::PermissionsExt;
use std::{
    fs,
    path::{Path, PathBuf},
    process::{Command, Output, Stdio},
};

const INIT_LINE: &str = "moai-cli: initialized";

struct Fixture {
    bin: tempfile::TempDir,
    home: tempfile::TempDir,
}

fn fixture() -> Fixture {
    let bin = tempfile::tempdir().unwrap();
    for host in ["claude", "codex"] {
        let path = bin.path().join(host);
        fs::write(
            &path,
            "#!/bin/sh\nprintf started > \"$MOAI_TEST_HOST_MARKER\"\nexit 0\n",
        )
        .unwrap();
        fs::set_permissions(&path, fs::Permissions::from_mode(0o755)).unwrap();
    }
    Fixture {
        bin,
        home: tempfile::tempdir().unwrap(),
    }
}

fn marker(f: &Fixture) -> PathBuf {
    f.home.path().join("host-started")
}

fn command(f: &Fixture, dir: &Path, path_var: &str, args: &[&str]) -> Command {
    let mut c = Command::new(env!("CARGO_BIN_EXE_moai-cli"));
    c.args(args)
        .current_dir(dir)
        .env("PATH", path_var)
        .env("HOME", f.home.path())
        .env("USERPROFILE", f.home.path())
        .env("MOAI_TEST_HOST_MARKER", marker(f))
        .env("ANTHROPIC_AUTH_TOKEN", "moai-test-dummy-not-real")
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .env_remove("CLAUDE_CODE_DISABLE_ADVISOR_TOOL")
        .env_remove(moai_cli::session::AGENT_SEQ_TEST_HOLD_VAR)
        .stdin(Stdio::null());
    c
}

fn with_git(f: &Fixture) -> String {
    format!("{}:/usr/bin:/bin", f.bin.path().display())
}

fn run_in(f: &Fixture, dir: &Path, args: &[&str]) -> Output {
    command(f, dir, &with_git(f), args).output().unwrap()
}

fn text(out: &Output) -> String {
    format!(
        "{}{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    )
}

fn json(out: &Output) -> Value {
    assert!(out.status.success(), "{}", text(out));
    serde_json::from_slice(&out.stdout).unwrap()
}

fn init_lines(out: &Output) -> usize {
    String::from_utf8_lossy(&out.stderr)
        .lines()
        .filter(|l| l.starts_with(INIT_LINE))
        .count()
}

fn git_init(dir: &Path) {
    let ok = common::git::command()
        .args(["init", "-q"])
        .current_dir(dir)
        .status()
        .unwrap()
        .success();
    assert!(ok);
}

/// A refused launch: exit 1, nothing on stdout, no host started, nothing created in `dir`.
fn refused_without_writes(f: &Fixture, out: &Output, dir: &Path, before: &[String]) -> String {
    let stderr = String::from_utf8_lossy(&out.stderr).to_string();
    assert_eq!(out.status.code(), Some(1), "{}", text(out));
    assert!(out.stdout.is_empty(), "{}", text(out));
    assert!(!marker(f).exists(), "host must not start: {stderr}");
    assert_eq!(entries(dir), before, "nothing may be created: {stderr}");
    stderr
}

fn entries(dir: &Path) -> Vec<String> {
    let mut v: Vec<String> = fs::read_dir(dir)
        .unwrap()
        .map(|e| e.unwrap().file_name().to_string_lossy().into_owned())
        .collect();
    v.sort();
    v
}

fn profile(dir: &Path) -> Value {
    serde_json::from_slice(&fs::read(dir.join(".moai-cli/project.yaml")).unwrap()).unwrap()
}

#[test]
fn the_first_launch_outside_git_initializes_the_folder_on_every_host() {
    for (host, profile_host) in [("cc", "claude"), ("glm", "claude"), ("codex", "codex")] {
        let f = fixture();
        let plain = tempfile::tempdir().unwrap();
        let dir = plain.path().join("My Project");
        fs::create_dir(&dir).unwrap();
        let out = run_in(&f, &dir, &[host, "-a"]);
        let v = json(&out);
        assert_eq!(v["name"], "agent-1", "{host}");
        assert_eq!(init_lines(&out), 1, "{host}: {}", text(&out));
        assert!(dir.join(".git").is_dir(), "{host}");
        assert!(
            dir.join(".git/.moai-cli/sessions/agent-seq").is_file(),
            "{host}"
        );
        let p = profile(&dir);
        assert_eq!(p["schema_version"], 2, "{host}");
        assert_eq!(p["project_id"], "my-project", "{host}");
        assert_eq!(p["hosts"], serde_json::json!([profile_host]), "{host}");
        let agents = fs::read_to_string(dir.join("AGENTS.md")).unwrap();
        assert!(
            agents.contains("<!-- moai:managed:begin id=moai-workflow -->"),
            "{host}: {agents}"
        );
        assert_eq!(fs::read(&marker(&f)).unwrap(), b"started", "{host}");

        // The second launch does not initialize again and keeps every file byte for byte.
        let yaml = fs::read(dir.join(".moai-cli/project.yaml")).unwrap();
        let agents = fs::read(dir.join("AGENTS.md")).unwrap();
        let out = run_in(&f, &dir, &[host, "-a"]);
        assert_eq!(json(&out)["name"], "agent-2", "{host}");
        assert_eq!(init_lines(&out), 0, "{host}: {}", text(&out));
        assert_eq!(fs::read(dir.join(".moai-cli/project.yaml")).unwrap(), yaml);
        assert_eq!(fs::read(dir.join("AGENTS.md")).unwrap(), agents);
    }
}

#[test]
fn leader_and_roleless_launches_initialize_too() {
    for args in [&["cc", "-l"][..], &["codex"][..]] {
        let f = fixture();
        let plain = tempfile::tempdir().unwrap();
        let out = run_in(&f, plain.path(), args);
        json(&out);
        assert_eq!(init_lines(&out), 1, "{args:?}: {}", text(&out));
        assert!(plain.path().join(".git").is_dir(), "{args:?}");
        assert!(
            plain.path().join(".moai-cli/project.yaml").is_file(),
            "{args:?}"
        );
    }
}

#[test]
fn a_repository_without_settings_gets_the_settings_only_and_keeps_its_agents_md() {
    let f = fixture();
    let p = tempfile::tempdir().unwrap();
    git_init(p.path());
    fs::write(p.path().join("AGENTS.md"), "mine\n").unwrap();
    let out = run_in(&f, p.path(), &["cc", "-a"]);
    assert_eq!(json(&out)["name"], "agent-1");
    assert_eq!(init_lines(&out), 1, "{}", text(&out));
    assert!(p.path().join(".moai-cli/project.yaml").is_file());
    assert_eq!(fs::read(p.path().join("AGENTS.md")).unwrap(), b"mine\n");
}

#[test]
fn a_project_that_has_settings_is_left_exactly_as_it_was() {
    let f = fixture();
    let p = tempfile::tempdir().unwrap();
    git_init(p.path());
    fs::create_dir(p.path().join(".moai-cli")).unwrap();
    fs::write(p.path().join(".moai-cli/project.yaml"), "kept\n").unwrap();
    let out = run_in(&f, p.path(), &["cc", "-a"]);
    assert_eq!(json(&out)["name"], "agent-1");
    assert_eq!(init_lines(&out), 0, "{}", text(&out));
    assert!(!String::from_utf8_lossy(&out.stderr).contains("moai-cli:"));
    assert_eq!(
        fs::read(p.path().join(".moai-cli/project.yaml")).unwrap(),
        b"kept\n"
    );
    assert!(!p.path().join("AGENTS.md").exists());
}

#[test]
fn home_itself_is_never_initialized_also_through_a_symlinked_home() {
    for args in [&["cc", "-a"][..], &["cc"][..], &["codex", "-l"][..]] {
        let f = fixture();
        let home = f.home.path().to_path_buf();
        let before = entries(&home);
        let out = run_in(&f, &home, args);
        let err = refused_without_writes(&f, &out, &home, &before);
        assert!(err.contains("HOME"), "{args:?}: {err}");
    }
    // HOME given as a symlink to the folder the launch runs in.
    let f = fixture();
    let links = tempfile::tempdir().unwrap();
    let link = links.path().join("home-link");
    std::os::unix::fs::symlink(f.home.path(), &link).unwrap();
    let before = entries(f.home.path());
    let out = command(&f, f.home.path(), &with_git(&f), &["cc", "-a"])
        .env("HOME", &link)
        .output()
        .unwrap();
    let err = refused_without_writes(&f, &out, f.home.path(), &before);
    assert!(err.contains("HOME"), "{err}");
}

#[test]
fn without_git_on_path_the_launcher_prints_install_guidance_for_this_os() {
    for args in [&["cc", "-a"][..], &["cc"][..], &["codex", "-l"][..]] {
        let f = fixture();
        let plain = tempfile::tempdir().unwrap();
        let only_stubs = f.bin.path().display().to_string();
        let out = command(&f, plain.path(), &only_stubs, args)
            .output()
            .unwrap();
        let err = refused_without_writes(&f, &out, plain.path(), &[]);
        let page = if cfg!(target_os = "macos") {
            "https://git-scm.com/install/mac"
        } else {
            "https://git-scm.com/install/linux"
        };
        assert!(err.contains("git is not installed"), "{args:?}: {err}");
        assert!(err.contains(page), "{args:?}: {err}");
    }
}

#[test]
fn dry_run_creates_nothing_and_reports_what_it_would_initialize() {
    let f = fixture();
    let plain = tempfile::tempdir().unwrap();
    let out = run_in(&f, plain.path(), &["cc", "-a", "--dry-run"]);
    let v = json(&out);
    assert_eq!(v["name"], "agent-1");
    let notes = v["notes"].to_string();
    for item in ["git init", ".moai-cli/project.yaml", "AGENTS.md"] {
        assert!(notes.contains(item), "{item}: {notes}");
    }
    assert_eq!(init_lines(&out), 0, "{}", text(&out));
    assert!(
        entries(plain.path()).is_empty(),
        "{:?}",
        entries(plain.path())
    );
    assert!(!marker(&f).exists());
    // A repository whose settings exist reports nothing to initialize.
    let p = tempfile::tempdir().unwrap();
    git_init(p.path());
    fs::create_dir(p.path().join(".moai-cli")).unwrap();
    fs::write(p.path().join(".moai-cli/project.yaml"), "kept\n").unwrap();
    let v = json(&run_in(&f, p.path(), &["cc", "-a", "--dry-run"]));
    assert!(!v["notes"].to_string().contains("would initialize"), "{v}");
}

#[test]
fn concurrent_first_launches_initialize_once_and_never_share_a_number() {
    let f = fixture();
    let plain = tempfile::tempdir().unwrap();
    let children: Vec<_> = (0..4)
        .map(|_| {
            command(&f, plain.path(), &with_git(&f), &["cc", "-a"])
                .stdout(Stdio::piped())
                .stderr(Stdio::piped())
                .spawn()
                .unwrap()
        })
        .collect();
    let mut names: Vec<String> = children
        .into_iter()
        .map(|c| {
            let out = c.wait_with_output().unwrap();
            json(&out)["name"].as_str().unwrap().to_string()
        })
        .collect();
    names.sort();
    assert_eq!(names, ["agent-1", "agent-2", "agent-3", "agent-4"]);
    assert!(plain.path().join(".moai-cli/project.yaml").is_file());
    assert!(plain.path().join("AGENTS.md").is_file());
}
