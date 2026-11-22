//! #2350 / #2337: the statusline's git read is scrubbed like every product git call. An
//! inherited `GIT_DIR` / `GIT_WORK_TREE` / `GIT_INDEX_FILE` pointing at another repository
//! (set by a host hook or an outer git command) must not change the rendered branch or
//! counts: the render describes the repository of the session directory only.
#![cfg(unix)]

mod common;
use common::CleanEnv;

use std::{
    fs,
    io::Write as _,
    path::Path,
    process::{Command, Stdio},
};

fn git(dir: &Path, args: &[&str]) {
    let status = common::git::command()
        .args(args)
        .current_dir(dir)
        .env_remove("GIT_DIR")
        .env_remove("GIT_WORK_TREE")
        .env_remove("GIT_INDEX_FILE")
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .env("GIT_AUTHOR_NAME", "t")
        .env("GIT_AUTHOR_EMAIL", "t@example.invalid")
        .env("GIT_COMMITTER_NAME", "t")
        .env("GIT_COMMITTER_EMAIL", "t@example.invalid")
        .status()
        .unwrap();
    assert!(status.success(), "git {args:?}");
}

/// One plain render with `stdin` naming `project`; `budget_ms` is the hidden render-budget
/// seam (a large value keeps this independent of machine load).
fn render(home: &Path, project: &Path, extra: &[(&str, &Path)], budget_ms: u64) -> String {
    let stdin = format!(
        r#"{{"model":{{"display_name":"Opus"}},"workspace":{{"current_dir":"{0}","project_dir":"{0}"}}}}"#,
        project.display()
    );
    let mut cmd = Command::new(crate::common::MOAI_BIN);
    cmd.args(["statusline", "--render-budget-ms", &budget_ms.to_string()])
        .current_dir(project)
        .clean_env()
        .env("PATH", std::env::var_os("PATH").unwrap_or_default())
        .env("HOME", home)
        .env("MOAI_HOME", home.join(".moai"))
        .env("NO_COLOR", "1")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped());
    for (k, v) in extra {
        cmd.env(k, v);
    }
    let mut child = cmd.spawn().unwrap();
    child
        .stdin
        .take()
        .unwrap()
        .write_all(stdin.as_bytes())
        .unwrap();
    let out = child.wait_with_output().unwrap();
    assert!(out.status.success());
    String::from_utf8(out.stdout).unwrap()
}

#[test]
fn an_inherited_git_dir_cannot_change_the_rendered_branch_or_counts() {
    let tmp = tempfile::tempdir().unwrap();
    let root = fs::canonicalize(tmp.path()).unwrap();
    let home = root.join("home");
    fs::create_dir_all(home.join(".moai")).unwrap();

    // The session's repository: branch `main`, clean.
    let project = root.join("project");
    fs::create_dir(&project).unwrap();
    git(&project, &["init", "-q", "-b", "main"]);
    fs::write(project.join("a.txt"), "a\n").unwrap();
    git(&project, &["add", "a.txt"]);
    git(&project, &["commit", "-q", "-m", "a"]);

    // Another repository: branch `intruder`, one modified and two untracked files.
    let other = root.join("other");
    fs::create_dir(&other).unwrap();
    git(&other, &["init", "-q", "-b", "intruder"]);
    fs::write(other.join("b.txt"), "b\n").unwrap();
    git(&other, &["add", "b.txt"]);
    git(&other, &["commit", "-q", "-m", "b"]);
    fs::write(other.join("b.txt"), "changed\n").unwrap();
    fs::write(other.join("u1.txt"), "u\n").unwrap();
    fs::write(other.join("u2.txt"), "u\n").unwrap();

    let clean = render(&home, &project, &[], WITHIN);
    assert!(clean.contains("🅱️ main"), "{clean}");
    // #2756: the repository line is the second of the three lines, not the last. A clean
    // branch has no `+N` (the other repository's three changes would show as `+3`); there is
    // no `💾` segment (leader decision 6862179411). The fixture has no remote and no forge
    // cache, so the repository and the counts say `n/a`
    // (SPEC-20261121-statusline-detached REQ-2, REQ-3) — the line stays, never a zero.
    let repository_line = clean.lines().find(|l| l.contains("🅱️")).unwrap();
    assert_eq!(
        repository_line, "📡 n/a, ⬇️ n/a 🔀 n/a | 🅱️ main",
        "{clean}"
    );
    assert!(!clean.contains("📁"), "{clean}");

    let git_dir = other.join(".git");
    let index = git_dir.join("index");
    let contaminated = render(
        &home,
        &project,
        &[
            ("GIT_DIR", git_dir.as_path()),
            ("GIT_WORK_TREE", other.as_path()),
            ("GIT_INDEX_FILE", index.as_path()),
        ],
        WITHIN,
    );
    assert_eq!(contaminated, clean);
    assert!(!contaminated.contains("intruder"), "{contaminated}");
}

/// A budget no git call can miss under any load.
const WITHIN: u64 = 3_600_000;

/// The over-budget path, deterministically: with an already expired budget no git is
/// started, so the repository row is omitted entirely. Without the seam (budget ignored)
/// git would run and this fails.
#[test]
fn an_expired_render_budget_drops_the_git_segment() {
    let tmp = tempfile::tempdir().unwrap();
    let root = fs::canonicalize(tmp.path()).unwrap();
    let home = root.join("home");
    fs::create_dir_all(home.join(".moai")).unwrap();
    let project = root.join("project");
    fs::create_dir(&project).unwrap();
    git(&project, &["init", "-q", "-b", "main"]);

    let within = render(&home, &project, &[], WITHIN);
    // #2756: the repository line sits between the environment and usage lines; the
    // fixture's missing remote and forge cache render as `n/a` slots
    // (SPEC-20261121-statusline-detached REQ-2, REQ-3).
    assert_eq!(
        within.lines().nth(1),
        Some("📡 n/a, ⬇️ n/a 🔀 n/a | 🅱️ main"),
        "{within}"
    );
    for _ in 0..5 {
        let over = render(&home, &project, &[], 0);
        assert!(
            !over.contains("🅱️") && !over.contains("📁") && !over.contains("📡"),
            "{over}"
        );
        let within_without_repo: Vec<_> = within
            .lines()
            .filter(|l| !l.starts_with("📡") && !l.starts_with("🅱️"))
            .collect();
        let over_lines: Vec<_> = over.lines().collect();
        assert_eq!(over_lines, within_without_repo, "{over}");
    }
}

// ---- SPEC-20261122-statusline-where: the `HEAD` file, one git process, the session cache ----
//
// Every case below runs the real binary against a repository whose `git` on PATH is a script
// that counts its starts and answers `status` with a canned text, so the number of processes and
// the origin of each shown fact (the `HEAD` file or git's output) are both visible. The canned
// text names a branch that is never the one in `HEAD`: a branch name that shows up from it was
// read from git, not from the file. TMPDIR and HOME are temporary folders of the test.

use std::os::unix::fs::{symlink, PermissionsExt};
use std::path::PathBuf;
use std::time::{Duration, Instant, SystemTime};

/// The render's clock of the cases (the hidden `--now-unix` seam).
const T: i64 = 1_800_000_000;
/// A made-up commit.
const SHA_A: &str = "0123456789abcdef0123456789abcdef01234567";
const SHA_B: &str = "fedcba9876543210fedcba9876543210fedcba98";
/// What the scripted `git status` says: a branch nobody checked out, one changed file.
const CANNED: &str = "## elsewhere\n M a.txt\n";
const BUDGET_WITHIN: u64 = 3_600_000;

struct Scripted {
    _guard: tempfile::TempDir,
    home: PathBuf,
    project: PathBuf,
    /// `TMPDIR` of the render.
    tmp: PathBuf,
    /// The folder holding the scripted `git`.
    bin: PathBuf,
    /// One line per start of the scripted `git`: its arguments.
    log: PathBuf,
}

impl Scripted {
    /// A repository (a real `git init`) whose `HEAD` file is then replaced by `head`.
    fn new(head: &str) -> Scripted {
        let guard = tempfile::tempdir().unwrap();
        let root = fs::canonicalize(guard.path()).unwrap();
        let home = root.join("home");
        let project = root.join("project");
        let tmp = root.join("tmp");
        let bin = root.join("bin");
        for d in [&home, &project, &tmp, &bin] {
            fs::create_dir_all(d).unwrap();
        }
        fs::create_dir_all(home.join(".moai")).unwrap();
        common::git::git_ok(&project, &["init", "-q", "-b", "main"]);
        common::git::git_in(
            &project,
            &["commit", "-q", "--allow-empty", "-m", "fixture"],
        );
        let log = root.join("git-starts.log");
        fs::write(&log, "").unwrap();
        let script = format!(
            "#!/bin/sh\n\
             echo \"$*\" >> '{log}'\n\
             for a in \"$@\"; do\n\
             \x20 if [ \"$a\" = status ]; then cat '{canned}'; exit 0; fi\n\
             done\n\
             exit 1\n",
            log = log.display(),
            canned = root.join("canned.txt").display(),
        );
        fs::write(root.join("canned.txt"), CANNED).unwrap();
        let git = bin.join("git");
        fs::write(&git, script).unwrap();
        fs::set_permissions(&git, fs::Permissions::from_mode(0o755)).unwrap();
        let s = Scripted {
            _guard: guard,
            home,
            project,
            tmp,
            bin,
            log,
        };
        s.set_head(head);
        s
    }

    fn git_dir(&self) -> PathBuf {
        self.project.join(".git")
    }

    fn set_head(&self, content: &str) {
        let head = self.git_dir().join("HEAD");
        let _ = fs::remove_file(&head);
        fs::write(head, content).unwrap();
    }

    /// The project profile names `name` as the integration branch, and its ref file holds `sha`.
    fn integration(&self, name: &str, sha: &str) {
        let state = self.project.join(".moai-cli");
        fs::create_dir_all(&state).unwrap();
        fs::write(
            state.join("project.yaml"),
            format!(
                "{{\"schema_version\":2,\"project_id\":\"demo\",\"work_flow\":{{\"flow\":\"m-flow\",\"integration_branch\":\"{name}\"}}}}\n"
            ),
        )
        .unwrap();
        self.set_integration_commit(name, sha);
    }

    fn set_integration_commit(&self, name: &str, sha: &str) {
        fs::write(
            self.git_dir().join("refs/heads").join(name),
            format!("{sha}\n"),
        )
        .unwrap();
    }

    /// The arguments of each start of `git`, in order.
    fn starts(&self) -> Vec<String> {
        fs::read_to_string(&self.log)
            .unwrap()
            .lines()
            .map(str::to_string)
            .collect()
    }

    fn cache_dir(&self) -> PathBuf {
        self.tmp.join("moai-statusline")
    }

    fn run(&self, now: i64, session: &str) -> String {
        self.run_with(now, Some(session), BUDGET_WITHIN)
    }

    /// One plain render; the exit code must be 0 and the render must end within 30 seconds.
    fn run_with(&self, now: i64, session: Option<&str>, budget_ms: u64) -> String {
        let mut stdin = serde_json::json!({
            "model": {"display_name": "Opus"},
            "workspace": {"current_dir": self.project, "project_dir": self.project},
        });
        if let Some(id) = session {
            stdin["session_id"] = id.into();
        }
        let mut cmd = Command::new(common::MOAI_BIN);
        cmd.args([
            "statusline",
            "--render-budget-ms",
            &budget_ms.to_string(),
            "--now-unix",
            &now.to_string(),
        ])
        .current_dir(&self.project)
        .clean_env()
        .env("PATH", format!("{}:/usr/bin:/bin", self.bin.display()))
        .env("HOME", &self.home)
        .env("MOAI_HOME", self.home.join(".moai"))
        .env("TMPDIR", &self.tmp)
        .env("NO_COLOR", "1")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped());
        let mut child = cmd.spawn().unwrap();
        child
            .stdin
            .take()
            .unwrap()
            .write_all(stdin.to_string().as_bytes())
            .unwrap();
        let started = Instant::now();
        loop {
            match child.try_wait().unwrap() {
                Some(_) => break,
                None if started.elapsed() > Duration::from_secs(30) => {
                    let _ = child.kill();
                    panic!("the render did not end within 30 seconds");
                }
                None => std::thread::sleep(Duration::from_millis(5)),
            }
        }
        let out = child.wait_with_output().unwrap();
        assert_eq!(
            out.status.code(),
            Some(0),
            "{}",
            String::from_utf8_lossy(&out.stderr)
        );
        String::from_utf8(out.stdout).unwrap()
    }
}

/// The line that carries the branch item (the old and the new first line alike).
fn branch_line(out: &str) -> &str {
    out.lines().find(|l| l.contains("🅱️")).unwrap_or("")
}

/// Every regular file under `dir`, recursively.
fn files_under(dir: &Path) -> Vec<PathBuf> {
    let mut out = Vec::new();
    let mut stack = vec![dir.to_path_buf()];
    while let Some(d) = stack.pop() {
        for e in fs::read_dir(&d).unwrap() {
            let p = e.unwrap().path();
            let meta = fs::symlink_metadata(&p).unwrap();
            if meta.is_dir() {
                stack.push(p);
            } else {
                out.push(p);
            }
        }
    }
    out.sort();
    out
}

fn mode_of(path: &Path) -> u32 {
    fs::symlink_metadata(path).unwrap().permissions().mode() & 0o777
}

/// AC-2, REQ-2: the branch is read from the `HEAD` file, never from git's output, and the one
/// `git status` process of the render supplies the numbers.
#[test]
fn head_file_and_process_count_branch() {
    let s = Scripted::new("ref: refs/heads/main\n");
    let out = s.run(T, "s-branch");
    let line = branch_line(&out);
    assert!(
        !out.contains("elsewhere"),
        "the branch name came from git's output, not from the HEAD file: {out}"
    );
    assert!(line.contains("main"), "the HEAD file's branch: {out}");
    assert!(
        line.contains("+1"),
        "the dirty count of the one process: {out}"
    );
    assert_eq!(s.starts().len(), 1, "one process: {:?}", s.starts());
    assert!(s.starts()[0].contains("status"), "{:?}", s.starts());
}

/// AC-2, REQ-2, REQ-7, D-2: a detached HEAD shows the first seven characters of the commit in
/// the `HEAD` file and starts no process for it.
#[test]
fn head_file_and_process_count_detached_head() {
    let s = Scripted::new(&format!("{SHA_A}\n"));
    fs::write(
        s.project.parent().unwrap().join("canned.txt"),
        "## HEAD (no branch)\n M a.txt\n",
    )
    .unwrap();
    let out = s.run(T, "s-detached");
    let starts = s.starts();
    assert!(starts.len() <= 1, "at most one process: {starts:?}");
    for start in &starts {
        assert!(
            start.contains("status")
                && !start.contains("rev-parse")
                && !start.contains("symbolic-ref"),
            "no process for the branch or the short commit: {starts:?}"
        );
    }
    let line = branch_line(&out);
    assert!(line.contains("0123456"), "the short commit: {out}");
    assert!(
        !line.contains("01234567"),
        "seven characters, no more: {out}"
    );
}

/// AC-2, REQ-2: a missing `HEAD` leaves the branch items out and starts at most one process.
#[test]
fn head_file_and_process_count_missing_head() {
    let s = Scripted::new("ref: refs/heads/main\n");
    fs::remove_file(s.git_dir().join("HEAD")).unwrap();
    let out = s.run(T, "s-missing");
    assert!(s.starts().len() <= 1, "{:?}", s.starts());
    assert!(!out.contains("🅱️"), "no branch item: {out}");
    assert!(!out.contains("🧭"), "no distance item: {out}");
    assert!(!out.contains("elsewhere"), "{out}");
}

/// A `HEAD` that is a link, a named pipe, a folder or text that is no branch and no commit is
/// refused, never followed or waited on: no branch item, and the render ends.
#[test]
fn head_file_and_process_count_refuses_links_pipes_and_folders() {
    let kinds = [
        "link",
        "pipe",
        "folder",
        "garbage",
        "control-bytes",
        "empty",
    ];
    for kind in kinds {
        let s = Scripted::new("ref: refs/heads/main\n");
        let head = s.git_dir().join("HEAD");
        fs::remove_file(&head).unwrap();
        match kind {
            "link" => {
                let real = s.project.parent().unwrap().join("real-head");
                fs::write(&real, "ref: refs/heads/main\n").unwrap();
                symlink(&real, &head).unwrap();
            }
            "pipe" => {
                let ok = Command::new("mkfifo")
                    .arg(&head)
                    .status()
                    .unwrap()
                    .success();
                assert!(ok, "mkfifo");
            }
            "folder" => fs::create_dir(&head).unwrap(),
            "garbage" => fs::write(&head, "not a head\n").unwrap(),
            "control-bytes" => fs::write(&head, "ref: refs/heads/a\x1b[31mb\n").unwrap(),
            _ => fs::write(&head, "").unwrap(),
        }
        let out = s.run(T, "s-refuse");
        assert!(!out.contains("🅱️"), "{kind}: no branch item: {out}");
        assert!(!out.contains("elsewhere"), "{kind}: {out}");
        assert!(
            !out.contains('\x1b'),
            "{kind}: no escape from a HEAD file: {out:?}"
        );
        assert!(s.starts().len() <= 1, "{kind}: {:?}", s.starts());
    }
}

/// Lane decision 1: the lane branch is measured against the integration branch by the one
/// process, with the two `-c` overrides in front of `status`.
#[test]
fn head_file_and_process_count_lane_branch_is_measured_against_the_integration_branch() {
    let s = Scripted::new("ref: refs/heads/lane\n");
    s.integration("next", SHA_A);
    s.run(T, "s-lane");
    assert_eq!(
        s.starts(),
        vec![
            "--no-optional-locks -c branch.lane.remote=. -c branch.lane.merge=refs/heads/next status --porcelain --branch"
                .to_string()
        ]
    );
    // On the integration branch itself git's own numbers: no override.
    let s = Scripted::new("ref: refs/heads/next\n");
    s.integration("next", SHA_A);
    s.run(T, "s-next");
    assert_eq!(
        s.starts(),
        vec!["--no-optional-locks status --porcelain --branch".to_string()]
    );
}

/// A branch name git could take as an option or a revision range never reaches git, whatever
/// the `HEAD` file says.
#[test]
fn head_file_and_process_count_unsafe_branch_names_never_reach_git() {
    let long = "a".repeat(101);
    for name in [
        "--upload-pack=x",
        "-x",
        "a..b",
        "HEAD",
        "a b",
        "a=b",
        long.as_str(),
    ] {
        let s = Scripted::new(&format!("ref: refs/heads/{name}\n"));
        s.integration("next", SHA_A);
        s.run(T, "s-unsafe");
        for start in s.starts() {
            assert!(!start.contains(name), "{name:?} reached git: {start}");
            assert!(!start.contains("-c "), "an override for {name:?}: {start}");
        }
        assert!(s.starts().len() <= 1, "{name:?}: {:?}", s.starts());
    }
}

/// The render deadline: a spent budget starts no git process (the render still ends with 0).
#[test]
fn head_file_and_process_count_a_spent_budget_starts_no_git() {
    let s = Scripted::new("ref: refs/heads/main\n");
    s.run_with(T, Some("s-spent"), 0);
    assert_eq!(s.starts(), Vec::<String>::new());
}

/// AC-4, REQ-6: a second render within ten seconds with the same `HEAD` starts no git, and
/// writes the same line.
#[test]
fn session_cache_second_render_within_ten_seconds_starts_no_git() {
    let s = Scripted::new("ref: refs/heads/main\n");
    let first = s.run(T, "s-hit");
    assert_eq!(s.starts().len(), 1);
    let second = s.run(T + 1, "s-hit");
    assert_eq!(
        s.starts().len(),
        1,
        "a cache hit starts no git: {:?}",
        s.starts()
    );
    assert_eq!(second, first);
    let third = s.run(T + 5, "s-hit");
    assert_eq!(s.starts().len(), 1, "{:?}", s.starts());
    assert_eq!(third, first);
}

/// REQ-6: a changed `HEAD` starts one git, and the new key is cached.
#[test]
fn session_cache_a_changed_head_starts_one_git() {
    let s = Scripted::new("ref: refs/heads/main\n");
    s.run(T, "s-head");
    s.run(T + 1, "s-head");
    assert_eq!(s.starts().len(), 1, "{:?}", s.starts());
    s.set_head("ref: refs/heads/other\n");
    let out = s.run(T + 2, "s-head");
    assert_eq!(s.starts().len(), 2, "{:?}", s.starts());
    assert!(branch_line(&out).contains("other"), "{out}");
    let again = s.run(T + 3, "s-head");
    assert_eq!(
        s.starts().len(),
        2,
        "the new key is cached: {:?}",
        s.starts()
    );
    assert_eq!(again, out);
}

/// REQ-6: a changed commit of the integration branch starts one git, from a loose ref and from
/// `packed-refs` alike.
#[test]
fn session_cache_a_changed_integration_commit_starts_one_git() {
    let s = Scripted::new("ref: refs/heads/lane\n");
    s.integration("next", SHA_A);
    s.run(T, "s-int");
    s.run(T + 1, "s-int");
    assert_eq!(s.starts().len(), 1, "{:?}", s.starts());
    s.set_integration_commit("next", SHA_B);
    s.run(T + 2, "s-int");
    assert_eq!(s.starts().len(), 2, "{:?}", s.starts());
    s.run(T + 3, "s-int");
    assert_eq!(s.starts().len(), 2, "{:?}", s.starts());

    let s = Scripted::new("ref: refs/heads/lane\n");
    s.integration("next", SHA_A);
    fs::remove_file(s.git_dir().join("refs/heads/next")).unwrap();
    let packed = |sha: &str| {
        format!("# pack-refs with: peeled fully-peeled sorted \n{sha} refs/heads/next\n")
    };
    fs::write(s.git_dir().join("packed-refs"), packed(SHA_A)).unwrap();
    s.run(T, "s-packed");
    s.run(T + 1, "s-packed");
    assert_eq!(s.starts().len(), 1, "{:?}", s.starts());
    fs::write(s.git_dir().join("packed-refs"), packed(SHA_B)).unwrap();
    s.run(T + 2, "s-packed");
    assert_eq!(s.starts().len(), 2, "{:?}", s.starts());
}

/// REQ-6 at the edge of the age: 9 seconds is a hit, 10 and 11 start one git, and the git that
/// an old cache asks for rewrites it.
#[test]
fn session_cache_age_edges() {
    for (age, starts) in [(0, 1), (9, 1), (10, 2), (11, 2), (3_600, 2)] {
        let s = Scripted::new("ref: refs/heads/main\n");
        s.run(T, "s-age");
        s.run(T + age, "s-age");
        assert_eq!(s.starts().len(), starts, "age {age}: {:?}", s.starts());
    }
    let s = Scripted::new("ref: refs/heads/main\n");
    s.run(T, "s-age");
    s.run(T + 11, "s-age");
    assert_eq!(s.starts().len(), 2);
    s.run(T + 15, "s-age");
    assert_eq!(s.starts().len(), 2, "rewritten at T + 11: {:?}", s.starts());
    s.run(T + 21, "s-age");
    assert_eq!(
        s.starts().len(),
        3,
        "10 seconds after the rewrite: {:?}",
        s.starts()
    );
}

/// REQ-5: a cache folder that cannot be written (a file in its place, no write permission, a
/// link) still gives the same line, and git is asked each time.
#[test]
fn session_cache_an_unwritable_folder_still_renders() {
    let reference = {
        let s = Scripted::new("ref: refs/heads/main\n");
        s.run(T, "s-bad")
    };
    for kind in [
        "tmp-is-a-file",
        "read-only-folder",
        "folder-is-a-file",
        "folder-is-a-link",
    ] {
        let s = Scripted::new("ref: refs/heads/main\n");
        let decoy = s.project.parent().unwrap().join("decoy");
        fs::create_dir(&decoy).unwrap();
        match kind {
            "tmp-is-a-file" => {
                fs::remove_dir(&s.tmp).unwrap();
                fs::write(&s.tmp, "x").unwrap();
            }
            "read-only-folder" => {
                fs::create_dir(s.cache_dir()).unwrap();
                fs::set_permissions(s.cache_dir(), fs::Permissions::from_mode(0o500)).unwrap();
            }
            "folder-is-a-file" => fs::write(s.cache_dir(), "x").unwrap(),
            _ => symlink(&decoy, s.cache_dir()).unwrap(),
        }
        let first = s.run(T, "s-bad");
        let second = s.run(T + 1, "s-bad");
        assert_eq!(first, reference, "{kind}");
        assert_eq!(second, reference, "{kind}");
        assert_eq!(
            s.starts().len(),
            2,
            "{kind}: git each time: {:?}",
            s.starts()
        );
        assert!(
            fs::read_dir(&decoy).unwrap().next().is_none(),
            "{kind}: nothing written through the link"
        );
        if kind == "read-only-folder" {
            fs::set_permissions(s.cache_dir(), fs::Permissions::from_mode(0o700)).unwrap();
        }
    }
}

/// REQ-5: the folder is made 0700, the file 0600, and the file is named by the session id and
/// holds the `HEAD` content; no temporary file stays behind.
#[test]
fn session_cache_file_is_private_and_named_by_the_session_id() {
    let s = Scripted::new("ref: refs/heads/main\n");
    s.run(T, "sess_A-1");
    let dir = s.cache_dir();
    assert!(dir.is_dir(), "the cache folder under TMPDIR");
    assert_eq!(mode_of(&dir), 0o700, "the folder");
    let file = dir.join("sess_A-1.json");
    assert_eq!(mode_of(&file), 0o600, "the file");
    assert_eq!(
        common::home::TempHome::names_in(&dir),
        vec!["sess_A-1.json"]
    );
    let text = fs::read_to_string(&file).unwrap();
    let value: serde_json::Value = serde_json::from_str(&text).expect("the cache is JSON");
    assert!(value.is_object(), "{text}");
    assert!(
        text.contains("ref: refs/heads/main"),
        "the HEAD content is the key: {text}"
    );
    // A second session has a file of its own, and the first is not touched.
    let before = fs::read(&file).unwrap();
    s.run(T + 1, "other-session");
    assert_eq!(
        common::home::TempHome::names_in(&dir),
        vec!["other-session.json", "sess_A-1.json"]
    );
    assert_eq!(fs::read(&file).unwrap(), before);
}

/// REQ-5: only a session id of 1 to 64 characters of `A-Z a-z 0-9 _ -` names a file; any other
/// id or none means no cache and git each time. A path in the id writes nothing outside.
#[test]
fn session_cache_a_bad_session_id_writes_no_file() {
    let long = "a".repeat(65);
    let ids: Vec<Option<&str>> = vec![
        None,
        Some(""),
        Some("../escape"),
        Some("a/b"),
        Some("a.b"),
        Some("a b"),
        Some("é"),
        Some(".."),
        Some(long.as_str()),
    ];
    for id in ids {
        let s = Scripted::new("ref: refs/heads/main\n");
        let first = s.run_with(T, id, BUDGET_WITHIN);
        let second = s.run_with(T + 1, id, BUDGET_WITHIN);
        assert_eq!(first, second, "{id:?}");
        assert_eq!(
            s.starts().len(),
            2,
            "{id:?}: git each time: {:?}",
            s.starts()
        );
        let root = s.project.parent().unwrap();
        let strays: Vec<_> = files_under(root)
            .into_iter()
            .filter(|p| p.extension().is_some_and(|e| e == "json") && !p.starts_with(&s.project))
            .collect();
        assert!(strays.is_empty(), "{id:?}: {strays:?}");
        assert!(
            !s.cache_dir().exists() || files_under(&s.cache_dir()).is_empty(),
            "{id:?}: a file in the cache folder"
        );
    }
}

/// REQ-5 at the edge of the length: 64 characters name a file, and it serves a second render.
#[test]
fn session_cache_a_session_id_of_sixty_four_characters_names_a_file() {
    let id = "a-_Z9".repeat(13)[..64].to_string();
    let s = Scripted::new("ref: refs/heads/main\n");
    s.run(T, &id);
    assert!(s.cache_dir().join(format!("{id}.json")).is_file());
    s.run(T + 1, &id);
    assert_eq!(s.starts().len(), 1, "{:?}", s.starts());
}

/// REQ-6: a cache that is malformed, empty, of the wrong shape or unreadable asks git once and
/// is rewritten, so the next render within ten seconds starts none.
#[test]
fn session_cache_a_malformed_file_asks_git_once_and_is_rewritten() {
    for content in [
        "not json",
        "",
        "{}",
        "[]",
        "null",
        "{\"head\":1,\"written\":\"x\"}",
        "\u{0}\u{1}",
    ] {
        let s = Scripted::new("ref: refs/heads/main\n");
        fs::create_dir(s.cache_dir()).unwrap();
        fs::set_permissions(s.cache_dir(), fs::Permissions::from_mode(0o700)).unwrap();
        let file = s.cache_dir().join("s-bad.json");
        fs::write(&file, content).unwrap();
        fs::set_permissions(&file, fs::Permissions::from_mode(0o600)).unwrap();
        let first = s.run(T, "s-bad");
        assert_eq!(
            s.starts().len(),
            1,
            "{content:?}: one git: {:?}",
            s.starts()
        );
        let second = s.run(T + 1, "s-bad");
        assert_eq!(
            s.starts().len(),
            1,
            "{content:?}: rewritten: {:?}",
            s.starts()
        );
        assert_eq!(second, first, "{content:?}");
        let text = fs::read_to_string(&file).unwrap();
        serde_json::from_str::<serde_json::Value>(&text).unwrap_or_else(|e| {
            panic!("{content:?}: the rewritten cache is not JSON ({e}): {text}")
        });
    }
    // A cache file this user cannot read is replaced, never followed to another file.
    let s = Scripted::new("ref: refs/heads/main\n");
    fs::create_dir(s.cache_dir()).unwrap();
    fs::set_permissions(s.cache_dir(), fs::Permissions::from_mode(0o700)).unwrap();
    let file = s.cache_dir().join("s-bad.json");
    fs::write(&file, "{}").unwrap();
    fs::set_permissions(&file, fs::Permissions::from_mode(0o000)).unwrap();
    s.run(T, "s-bad");
    s.run(T + 1, "s-bad");
    assert_eq!(s.starts().len(), 1, "{:?}", s.starts());
}

/// REQ-5: a cache file that is a link is never written through; the link is replaced.
#[test]
fn session_cache_a_linked_file_is_never_written_through() {
    let s = Scripted::new("ref: refs/heads/main\n");
    fs::create_dir(s.cache_dir()).unwrap();
    fs::set_permissions(s.cache_dir(), fs::Permissions::from_mode(0o700)).unwrap();
    let decoy = s.project.parent().unwrap().join("decoy.json");
    fs::write(&decoy, "decoy").unwrap();
    symlink(&decoy, s.cache_dir().join("s-link.json")).unwrap();
    s.run(T, "s-link");
    assert_eq!(fs::read_to_string(&decoy).unwrap(), "decoy");
    assert!(
        !fs::symlink_metadata(s.cache_dir().join("s-link.json"))
            .unwrap()
            .file_type()
            .is_symlink(),
        "the link was replaced by a file"
    );
}

/// Risk 4: the render that writes a cache removes cache files older than a day, at most ten per
/// render, by the file's age against the render's clock; a render that hits the cache removes
/// none, and a file that is no cache file stays.
#[test]
fn session_cache_removes_old_files_at_most_ten_per_render() {
    let age_file = |dir: &Path, name: &str, age: i64| {
        let path = dir.join(name);
        fs::write(&path, "{}").unwrap();
        fs::set_permissions(&path, fs::Permissions::from_mode(0o600)).unwrap();
        let when = SystemTime::UNIX_EPOCH + Duration::from_secs((T - age) as u64);
        fs::OpenOptions::new()
            .write(true)
            .open(&path)
            .unwrap()
            .set_modified(when)
            .unwrap();
    };
    let setup = |old: usize| {
        let s = Scripted::new("ref: refs/heads/main\n");
        fs::create_dir(s.cache_dir()).unwrap();
        fs::set_permissions(s.cache_dir(), fs::Permissions::from_mode(0o700)).unwrap();
        for i in 0..old {
            age_file(&s.cache_dir(), &format!("old-{i:02}.json"), 2 * 86_400);
        }
        age_file(&s.cache_dir(), "recent.json", 3_600);
        age_file(&s.cache_dir(), "edge-kept.json", 86_400);
        age_file(&s.cache_dir(), "edge-removed.json", 86_401);
        age_file(&s.cache_dir(), "notes.txt", 5 * 86_400);
        s
    };
    let names = |s: &Scripted| common::home::TempHome::names_in(&s.cache_dir());
    // `edge-removed.json` is one more old file: 9 old files make 10 in all, and 10 make 11.
    let s = setup(9);
    s.run(T, "new");
    assert_eq!(
        names(&s),
        ["edge-kept.json", "new.json", "notes.txt", "recent.json"],
        "10 old files: all removed"
    );
    let s = setup(10);
    s.run(T, "new");
    let left = names(&s);
    assert_eq!(
        left.len(),
        4 + 1,
        "11 old files, 10 removed, 1 stays: {left:?}"
    );
    assert!(left.contains(&"new.json".to_string()));
    assert!(left.contains(&"recent.json".to_string()));
    assert!(left.contains(&"edge-kept.json".to_string()));
    assert!(left.contains(&"notes.txt".to_string()));
    // 12 old files and the edge file: ten removed, three stay.
    let s = setup(12);
    s.run(T, "new");
    let left = names(&s);
    let old_left = left
        .iter()
        .filter(|n| n.starts_with("old-") || *n == "edge-removed.json")
        .count();
    assert_eq!(old_left, 3, "{left:?}");
    // A render that hits the cache removes nothing.
    let s = setup(3);
    s.run(T, "new");
    let after_write = names(&s);
    for i in 0..3 {
        age_file(&s.cache_dir(), &format!("late-{i}.json"), 3 * 86_400);
    }
    s.run(T + 1, "new");
    assert_eq!(s.starts().len(), 1, "a hit: {:?}", s.starts());
    let mut want = after_write;
    want.extend((0..3).map(|i| format!("late-{i}.json")));
    want.sort();
    assert_eq!(names(&s), want, "a hit removes no file");
}
