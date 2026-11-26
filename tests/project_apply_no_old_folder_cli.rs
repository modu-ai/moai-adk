//! SPEC-20261126-no-state-folder-writes: a project with no earlier state folder whose state
//! verdict is `go`, `mixed`, `interrupted` or `refuse` gets nothing written by `project-apply`
//! or by a first `cc` launch; both say the next step in one line, and the launch still starts
//! the host. Every run uses a temporary project, a temporary HOME and a stub `claude` on PATH;
//! no real host starts. The state folder names come from the resolver's own constants.
#![cfg(unix)]
mod common;
use moai_cli::state_root;
use serde_json::Value;
use std::{
    collections::BTreeMap,
    fs,
    io::Write,
    os::unix::fs::{symlink, PermissionsExt},
    path::Path,
    process::{Command, Output, Stdio},
};

const STOP: &str = "moai: no MoAI state is created here";
const MOAI: &str = state_root::MOAI_DIR;
const EARLIER: &str = state_root::EARLIER_PROJECT_DIR;

/// The four verdicts that stop a writer, each with the words its next step must hold.
const STOPPED: [(&str, &str); 4] = [
    ("interrupted", "moai go-state-revert-plan"),
    ("go", "moai go-state-plan"),
    ("mixed", "you decide"),
    ("refuse", "you decide"),
];

/// `rel` inside this generation's state folder, as a project-relative path.
fn state(rel: &str) -> String {
    format!("{MOAI}/{rel}")
}

/// The profile's project-relative path inside the earlier state folder.
fn earlier_profile() -> String {
    format!("{EARLIER}/project.yaml")
}

struct Fixture {
    bin: tempfile::TempDir,
    home: tempfile::TempDir,
}

fn fixture() -> Fixture {
    let bin = tempfile::tempdir().unwrap();
    let claude = bin.path().join("claude");
    fs::write(
        &claude,
        "#!/bin/sh\nprintf started > \"$MOAI_TEST_HOST_MARKER\"\nexit 0\n",
    )
    .unwrap();
    fs::set_permissions(&claude, fs::Permissions::from_mode(0o755)).unwrap();
    Fixture {
        bin,
        home: tempfile::tempdir().unwrap(),
    }
}

impl Fixture {
    fn command(&self, dir: &Path, args: &[&str]) -> Command {
        let mut c = Command::new(crate::common::MOAI_BIN);
        c.args(args)
            .current_dir(dir)
            .env(
                "PATH",
                format!("{}:/usr/bin:/bin", self.bin.path().display()),
            )
            .env("HOME", self.home.path())
            .env("USERPROFILE", self.home.path())
            .env("MOAI_TEST_HOST_MARKER", self.home.path().join("started"))
            .env("ANTHROPIC_AUTH_TOKEN", "moai-test-dummy-not-real")
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_SYSTEM", "/dev/null")
            .env_remove("CLAUDE_CODE_DISABLE_ADVISOR_TOOL")
            .stdin(Stdio::null());
        c
    }

    fn run(&self, dir: &Path, args: &[&str]) -> Output {
        self.command(dir, args).output().unwrap()
    }

    fn run_with(&self, dir: &Path, args: &[&str], stdin: &[u8]) -> Output {
        let mut child = self
            .command(dir, args)
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
            .unwrap();
        // A command that refuses before reading its input closes the pipe; that is no failure.
        let _ = child.stdin.take().unwrap().write_all(stdin);
        child.wait_with_output().unwrap()
    }

    fn host_started(&self) -> bool {
        self.home.path().join("started").exists()
    }

    /// A profile as a first launch writes it, taken from a fresh temporary project.
    fn profile(&self) -> Vec<u8> {
        let fresh = tempfile::tempdir().unwrap();
        git_init(fresh.path());
        let out = self.run(fresh.path(), &["cc"]);
        assert!(out.status.success(), "{}", text(&out));
        fs::remove_file(self.home.path().join("started")).unwrap();
        fs::read(fresh.path().join(state("project.yaml"))).unwrap()
    }

    /// `project-plan` then `project-apply` of the plan it printed.
    fn plan_and_apply(&self, dir: &Path, profile: &[u8]) -> Output {
        let shown = dir.to_str().unwrap();
        let planned = self.run_with(dir, &["project-plan", shown], profile);
        assert!(planned.status.success(), "{}", text(&planned));
        let plan: Value = serde_json::from_slice(&planned.stdout).unwrap();
        let result = &plan["result"];
        let digest = result["plan_digest"].as_str().unwrap();
        self.run_with(
            dir,
            &[
                "project-apply",
                "--plan-digest",
                digest,
                "--permission-reference",
                "test",
                shown,
            ],
            result.to_string().as_bytes(),
        )
    }
}

fn text(out: &Output) -> String {
    format!(
        "{}{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    )
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

fn put(root: &Path, rel: &str, body: &str) {
    let path = root.join(rel);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, body).unwrap();
}

/// A temporary project whose state verdict is `verdict`, with no earlier state folder. The
/// second folder is the target of the `refuse` link, kept alive and checked for writes.
fn stopped_project(verdict: &str) -> (tempfile::TempDir, tempfile::TempDir) {
    let project = tempfile::tempdir().unwrap();
    let elsewhere = tempfile::tempdir().unwrap();
    let root = project.path();
    put(root, "README.md", "practice\n");
    match verdict {
        "interrupted" => put(
            root,
            &format!("{}/manifest.json", state_root::GO_BACKUP),
            "{}",
        ),
        "go" => put(root, &state("manifest.json"), "{}"),
        // The earlier marker name beside a Go marker owns nothing.
        "mixed" => {
            put(root, &state("manifest.json"), "{}");
            put(
                root,
                &state(state_root::LEGACY_MARKER),
                "{\"schema_version\": 1, \"generation\": \"moai-cli\"}",
            );
        }
        "refuse" => symlink(elsewhere.path(), root.join(MOAI)).unwrap(),
        other => panic!("no fixture for {other}"),
    }
    assert_eq!(state_root::classify(root).generation, verdict);
    (project, elsewhere)
}

/// Every entry under `root`: a folder, a link with its target (never followed) or a file with
/// its bytes.
fn tree(root: &Path) -> BTreeMap<String, Vec<u8>> {
    fn walk(root: &Path, dir: &Path, into: &mut BTreeMap<String, Vec<u8>>) {
        for entry in fs::read_dir(dir).unwrap() {
            let path = entry.unwrap().path();
            let name = path.strip_prefix(root).unwrap().display().to_string();
            let meta = fs::symlink_metadata(&path).unwrap();
            if meta.file_type().is_symlink() {
                let target = fs::read_link(&path).unwrap();
                into.insert(name, format!("link {}", target.display()).into_bytes());
            } else if meta.is_dir() {
                into.insert(format!("{name}/"), Vec::new());
                walk(root, &path, into);
            } else {
                into.insert(name, fs::read(&path).unwrap());
            }
        }
    }
    let mut into = BTreeMap::new();
    walk(root, root, &mut into);
    into
}

#[test]
fn apply_writes_nothing_and_names_the_next_step_for_every_stopped_verdict() {
    let f = fixture();
    let profile = f.profile();
    for (verdict, next) in STOPPED {
        let (project, elsewhere) = stopped_project(verdict);
        let before = tree(project.path());
        let out = f.plan_and_apply(project.path(), &profile);
        assert!(!out.status.success(), "{verdict}: {}", text(&out));
        let answer: Value = serde_json::from_slice(&out.stdout).unwrap();
        assert_eq!(
            answer["error"]["kind"], "invalid_input",
            "{verdict}: {answer}"
        );
        let message = answer["error"]["message"].as_str().unwrap();
        for word in [STOP, &format!("`{verdict}`"), next] {
            assert!(message.contains(word), "{verdict}: {word}: {message}");
        }
        assert!(!message.contains('\n'), "{verdict}: {message}");
        assert_eq!(tree(project.path()), before, "{verdict}");
        assert!(tree(elsewhere.path()).is_empty(), "{verdict}");
    }
}

#[test]
fn apply_refuses_before_it_reads_the_plan() {
    let f = fixture();
    let (project, _elsewhere) = stopped_project("go");
    let before = tree(project.path());
    let out = f.run_with(
        project.path(),
        &[
            "project-apply",
            "--plan-digest",
            "none",
            "--permission-reference",
            "test",
            ".",
        ],
        b"not a plan",
    );
    assert!(!out.status.success(), "{}", text(&out));
    assert!(text(&out).contains(STOP), "{}", text(&out));
    assert_eq!(tree(project.path()), before);
}

#[test]
fn launch_creates_nothing_says_one_line_and_starts_the_host() {
    for (verdict, next) in STOPPED {
        for repository in [true, false] {
            let f = fixture();
            let (project, elsewhere) = stopped_project(verdict);
            if repository {
                git_init(project.path());
            }
            let before = tree(project.path());
            let case = format!("{verdict}, repository={repository}");
            let out = f.run(project.path(), &["cc"]);
            assert!(out.status.success(), "{case}: {}", text(&out));
            assert!(f.host_started(), "{case}: {}", text(&out));
            let stderr = String::from_utf8_lossy(&out.stderr);
            let lines: Vec<&str> = stderr.lines().filter(|l| l.starts_with(STOP)).collect();
            assert_eq!(lines.len(), 1, "{case}: {stderr}");
            assert!(lines[0].contains(next), "{case}: {stderr}");
            assert!(!stderr.contains("moai: initialized"), "{case}: {stderr}");
            // The whole tree, `.git` included: no repository, marker, profile or instruction file.
            assert_eq!(tree(project.path()), before, "{case}");
            assert!(tree(elsewhere.path()).is_empty(), "{case}");
        }
    }
}

#[test]
fn launch_dry_run_reports_the_same_line_and_writes_nothing() {
    for (verdict, next) in STOPPED {
        for repository in [true, false] {
            let f = fixture();
            let (project, _elsewhere) = stopped_project(verdict);
            if repository {
                git_init(project.path());
            }
            let before = tree(project.path());
            let case = format!("{verdict}, repository={repository}");
            let out = f.run(project.path(), &["cc", "--dry-run"]);
            assert!(out.status.success(), "{case}: {}", text(&out));
            let answer: Value = serde_json::from_slice(&out.stdout).unwrap();
            let notes = answer["notes"].to_string();
            assert!(notes.contains(STOP), "{case}: {notes}");
            assert!(notes.contains(next), "{case}: {notes}");
            assert!(!notes.contains("would initialize"), "{case}: {notes}");
            assert!(!f.host_started(), "{case}");
            assert_eq!(tree(project.path()), before, "{case}");
        }
    }
}

#[test]
fn control_apply_writes_where_it_wrote_before() {
    let f = fixture();
    let profile = f.profile();
    // A new project, an empty state folder and one holding unmarked files are this generation's.
    for kind in ["new", "empty", "unknown"] {
        let project = tempfile::tempdir().unwrap();
        match kind {
            "empty" => fs::create_dir(project.path().join(MOAI)).unwrap(),
            "unknown" => put(project.path(), &state("notes.txt"), "kept\n"),
            _ => {}
        }
        let out = f.plan_and_apply(project.path(), &profile);
        assert!(out.status.success(), "{kind}: {}", text(&out));
        assert!(
            project.path().join(state("project.yaml")).is_file(),
            "{kind}"
        );
        assert!(!project.path().join(EARLIER).exists(), "{kind}");
    }
    // An earlier folder that is already there keeps being written, and Go's folder is not.
    let project = tempfile::tempdir().unwrap();
    put(project.path(), &state("manifest.json"), "{}");
    fs::create_dir(project.path().join(EARLIER)).unwrap();
    let go_before = tree(&project.path().join(MOAI));
    let out = f.plan_and_apply(project.path(), &profile);
    assert!(out.status.success(), "{}", text(&out));
    assert!(project.path().join(earlier_profile()).is_file());
    assert_eq!(tree(&project.path().join(MOAI)), go_before);
}

#[test]
fn control_launch_initializes_a_new_project_and_one_with_the_earlier_folder() {
    for (earlier, profile) in [(false, state("project.yaml")), (true, earlier_profile())] {
        let f = fixture();
        let project = tempfile::tempdir().unwrap();
        git_init(project.path());
        if earlier {
            put(project.path(), &state("manifest.json"), "{}");
            fs::create_dir(project.path().join(EARLIER)).unwrap();
        }
        let out = f.run(project.path(), &["cc"]);
        assert!(out.status.success(), "{earlier}: {}", text(&out));
        assert!(f.host_started(), "{earlier}");
        assert!(
            project.path().join(&profile).is_file(),
            "{earlier}: {}",
            text(&out)
        );
        assert!(!text(&out).contains(STOP), "{earlier}: {}", text(&out));
    }
}

/// A valid current marker beside a Go marker (git brought the Go marker back). Which folder the
/// resolver answers for this tree is the resolver's business; both writers follow its answer.
fn returned_marker_project() -> tempfile::TempDir {
    let project = tempfile::tempdir().unwrap();
    put(project.path(), &state("manifest.json"), "{}");
    put(
        project.path(),
        &state(state_root::RUST_MARKER),
        "{\"schema_version\": 1, \"generation\": \"moai\"}",
    );
    assert_eq!(state_root::classify(project.path()).generation, "mixed");
    project
}

#[test]
fn follow_apply_writes_exactly_when_the_resolver_answers_moai() {
    let f = fixture();
    let profile = f.profile();
    let project = returned_marker_project();
    let writes = state_root::dir(project.path()) == MOAI;
    let before = tree(project.path());
    let out = f.plan_and_apply(project.path(), &profile);
    if writes {
        assert!(out.status.success(), "{}", text(&out));
        assert!(project.path().join(state("project.yaml")).is_file());
    } else {
        assert!(!out.status.success(), "{}", text(&out));
        assert!(text(&out).contains(STOP), "{}", text(&out));
        assert_eq!(tree(project.path()), before);
    }
    assert!(!project.path().join(EARLIER).exists());
}

#[test]
fn follow_launch_initializes_exactly_when_the_resolver_answers_moai() {
    let f = fixture();
    let project = returned_marker_project();
    git_init(project.path());
    let writes = state_root::dir(project.path()) == MOAI;
    let before = tree(project.path());
    let out = f.run(project.path(), &["cc"]);
    assert!(out.status.success(), "{}", text(&out));
    assert!(f.host_started(), "{}", text(&out));
    if writes {
        assert!(
            project.path().join(state("project.yaml")).is_file(),
            "{}",
            text(&out)
        );
        assert!(!text(&out).contains(STOP), "{}", text(&out));
    } else {
        assert!(text(&out).contains(STOP), "{}", text(&out));
        assert_eq!(tree(project.path()), before);
    }
    assert!(!project.path().join(EARLIER).exists());
}
