//! SPEC-20261121-xtask-local-check REQ-4 to REQ-8: `xtask local-check` and `tools/local_check.sh`
//! run in the same temporary repository, with the same environment and the same stub tools on
//! `PATH`, and must print the same lines, call the same tools in the same folders and exit with
//! the same code. The only fields that may differ are the ones that depend on the clock, the
//! process, the script file and the temporary folder; `mask` hides them and each scenario also
//! asserts what the shell script did, so two identical wrong answers cannot pass.
//!
//! The stubs stand in for `cargo`, `python3`, `rustfmt`, `rustc` and (when asked) `moai`: no
//! real build runs here.

use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};
use std::sync::atomic::{AtomicUsize, Ordering};
use std::time::{Duration, Instant};

const XTASK: &str = env!("CARGO_BIN_EXE_xtask");
const BASE_BRANCH: &str = "v2026.11.0";

fn repo_root() -> PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR"))
        .parent()
        .unwrap()
        .to_path_buf()
}

/// A temporary folder removed when it goes out of scope.
struct Tmp(PathBuf);

impl Tmp {
    fn new() -> Tmp {
        static COUNT: AtomicUsize = AtomicUsize::new(0);
        let n = COUNT.fetch_add(1, Ordering::SeqCst);
        let path = std::env::temp_dir().join(format!("xt-parity-{}-{n}", std::process::id()));
        fs::create_dir_all(&path).unwrap();
        Tmp(fs::canonicalize(path).unwrap())
    }
}

impl Drop for Tmp {
    fn drop(&mut self) {
        let _ = fs::remove_dir_all(&self.0);
    }
}

#[derive(Clone, Copy, PartialEq, Debug)]
enum Tool {
    Shell,
    Xtask,
}

#[derive(Debug, PartialEq)]
struct Outcome {
    code: Option<i32>,
    stdout: String,
    stderr: String,
    /// What the stub tools were called with, in order.
    calls: Vec<String>,
    /// Where they ran: folder, tree, first and second parent.
    wheres: Vec<String>,
    /// The arguments `moai heavy-slot-acquire` received.
    slots: Vec<String>,
    /// The step summary file, when one was named.
    summary: String,
}

type Env<'a> = &'a [(&'a str, Option<&'a str>)];

struct Fx {
    tmp: Tmp,
    repo: PathBuf,
    bin: PathBuf,
    tmpdir: PathBuf,
    /// Candidate mode: a bare origin and a clone; otherwise one plain repository.
    candidate: bool,
}

fn git_env(home: &Path) -> Vec<(String, String)> {
    [
        ("PATH", "/usr/bin:/bin"),
        ("HOME", home.to_str().unwrap()),
        ("LC_ALL", "C"),
        ("GIT_CONFIG_NOSYSTEM", "1"),
        ("GIT_AUTHOR_NAME", "t"),
        ("GIT_AUTHOR_EMAIL", "t@example.invalid"),
        ("GIT_COMMITTER_NAME", "t"),
        ("GIT_COMMITTER_EMAIL", "t@example.invalid"),
    ]
    .iter()
    .map(|(k, v)| (k.to_string(), v.to_string()))
    .collect()
}

fn git_in(home: &Path, cwd: &Path, args: &[&str]) -> String {
    let out = Command::new("git")
        .args(args)
        .current_dir(cwd)
        .env_clear()
        .envs(git_env(home))
        .output()
        .unwrap();
    assert!(
        out.status.success(),
        "git {args:?} in {cwd:?}: {}",
        String::from_utf8_lossy(&out.stderr)
    );
    String::from_utf8_lossy(&out.stdout).trim().to_string()
}

impl Fx {
    /// One repository with two commits, the script copied in, as the Python harness has it.
    fn in_place() -> Fx {
        Fx::build(false, false)
    }

    /// A bare origin, a clone on `feature` and a private `TMPDIR`.
    fn candidate() -> Fx {
        Fx::build(true, false)
    }

    fn with_slot() -> Fx {
        Fx::build(false, true)
    }

    fn build(candidate: bool, slot: bool) -> Fx {
        let tmp = Tmp::new();
        let base = tmp.0.clone();
        let repo = base.join("repo");
        let bin = base.join("bin");
        let tmpdir = base.join("tmpdir");
        for dir in [&repo, &bin, &tmpdir] {
            fs::create_dir_all(dir).unwrap();
        }
        fs::create_dir_all(repo.join("tools")).unwrap();
        fs::create_dir_all(repo.join("xtask/src")).unwrap();
        fs::copy(
            repo_root().join("tools/local_check.sh"),
            repo.join("tools/local_check.sh"),
        )
        .unwrap();
        // The file `xtask` hashes for the `script=` field must be there for a clean tree.
        fs::copy(
            repo_root().join("xtask/src/local_check.rs"),
            repo.join("xtask/src/local_check.rs"),
        )
        .unwrap_or_else(|_| {
            fs::write(repo.join("xtask/src/local_check.rs"), "// stand-in\n").unwrap();
            0
        });
        let fx = Fx {
            tmp,
            repo,
            bin,
            tmpdir,
            candidate,
        };
        fx.write_stubs(slot);
        if candidate {
            let origin = base.join("origin.git");
            fx.git(
                &base,
                &[
                    "init",
                    "-q",
                    "--bare",
                    "-b",
                    BASE_BRANCH,
                    origin.to_str().unwrap(),
                ],
            );
            fx.git(&fx.repo, &["init", "-q", "-b", BASE_BRANCH]);
            fx.git(
                &fx.repo,
                &["remote", "add", "origin", origin.to_str().unwrap()],
            );
            fx.commit(
                &fx.repo,
                &[("README.md", "one\ntwo\n"), (".gitignore", "*.log\n")],
                "base",
            );
            fx.git(&fx.repo, &["push", "-q", "origin", BASE_BRANCH]);
            fx.git(&fx.repo, &["checkout", "-q", "-b", "feature"]);
        } else {
            fx.git(&fx.repo, &["init", "-q", "-b", "main"]);
            fx.commit(
                &fx.repo,
                &[("README.md", "one\n"), (".gitignore", "*.log\n")],
                "one",
            );
            fx.commit(&fx.repo, &[("README.md", "one\ntwo\n")], "two");
        }
        fx
    }

    fn git(&self, cwd: &Path, args: &[&str]) -> String {
        git_in(&self.tmp.0, cwd, args)
    }

    fn commit(&self, cwd: &Path, files: &[(&str, &str)], message: &str) -> String {
        for (name, text) in files {
            fs::write(cwd.join(name), text).unwrap();
        }
        self.git(cwd, &["add", "-A"]);
        self.git(cwd, &["commit", "-q", "-m", message]);
        self.git(cwd, &["rev-parse", "HEAD"])
    }

    fn origin(&self) -> PathBuf {
        self.tmp.0.join("origin.git")
    }

    /// A commit on origin's base branch made by someone else, not by the clone.
    fn advance_origin(&self, files: &[(&str, &str)]) {
        let other = self.tmp.0.join("other");
        if !other.exists() {
            self.git(
                &self.tmp.0,
                &[
                    "clone",
                    "-q",
                    self.origin().to_str().unwrap(),
                    other.to_str().unwrap(),
                ],
            );
        }
        self.git(&other, &["fetch", "-q", "origin"]);
        self.git(
            &other,
            &[
                "checkout",
                "-q",
                "-B",
                "work",
                &format!("origin/{BASE_BRANCH}"),
            ],
        );
        self.commit(&other, files, "origin moved");
        self.git(
            &other,
            &[
                "push",
                "-q",
                "origin",
                &format!("HEAD:refs/heads/{BASE_BRANCH}"),
            ],
        );
    }

    /// origin's base branch becomes a root commit with nothing in common with the head.
    fn replace_origin_history(&self, files: &[(&str, &str)]) {
        self.advance_origin(&[("README.md", "x\n")]);
        let other = self.tmp.0.join("other");
        self.git(&other, &["checkout", "-q", "--orphan", "rewritten"]);
        self.git(&other, &["rm", "-rfq", "."]);
        self.commit(&other, files, "unrelated root");
        self.git(
            &other,
            &[
                "push",
                "-q",
                "--force",
                "origin",
                &format!("HEAD:refs/heads/{BASE_BRANCH}"),
            ],
        );
    }

    fn write_stubs(&self, slot: bool) {
        let body = |name: &str| {
            format!(
                "#!/bin/sh\n\
                 echo \"{name} $*\" >> \"$STUB_CALLS\"\n\
                 printf '%s\\t%s\\t%s\\t%s\\n' \"$(pwd -P)\" \"$(git rev-parse 'HEAD^{{tree}}' 2>/dev/null)\" \
                 \"$(git rev-parse 'HEAD^1' 2>/dev/null)\" \"$(git rev-parse 'HEAD^2' 2>/dev/null)\" >> \"$STUB_WHERE\"\n\
                 if [ -n \"${{STUB_SLEEP:-}}\" ]; then sleep \"$STUB_SLEEP\"; fi\n\
                 if [ -f rule_forbids_forbidden.txt ] && [ -f forbidden.txt ]; then\n\
                 \x20 echo 'rule: forbidden.txt must not exist' >&2; exit 1\n\
                 fi\n\
                 case \"{name} $*\" in *\"${{STUB_FAIL:-@none@}}\"*) exit 1;; esac\n\
                 exit 0\n"
            )
        };
        for name in ["cargo", "python3", "rustfmt"] {
            self.script(&name.to_string(), &body(name));
        }
        self.script(
            "rustc",
            "#!/bin/sh\n\
             if [ -n \"${STUB_RUSTC_OUT+x}\" ]; then\n\
             \x20 printf '%s\\n' \"$STUB_RUSTC_OUT\"; exit \"${STUB_RUSTC_EXIT:-0}\"\n\
             fi\n\
             echo 'rustc 1.99.0 (stub 2026-01-01)'\n",
        );
        if slot {
            // `moai heavy-slot-acquire`: records its arguments and runs the command after `--`,
            // or exits 75 (the slot's refusal) when STUB_SLOT_REFUSE is set.
            self.script(
                "moai",
                "#!/bin/sh\n\
                 [ \"$1 $2\" = 'heavy-slot-acquire --help' ] && exit 0\n\
                 echo \"$*\" >> \"$STUB_SLOT_CALLS\"\n\
                 [ -n \"${STUB_SLOT_REFUSE:-}\" ] && exit 75\n\
                 while [ \"$#\" -gt 0 ] && [ \"$1\" != -- ]; do shift; done\n\
                 shift\n\
                 exec \"$@\"\n",
            );
        }
    }

    fn script(&self, name: &str, text: &str) {
        use std::os::unix::fs::PermissionsExt;
        let path = self.bin.join(name);
        fs::write(&path, text).unwrap();
        fs::set_permissions(&path, fs::Permissions::from_mode(0o755)).unwrap();
    }

    fn files(&self, tag: &str) -> [PathBuf; 4] {
        let dir = self.tmp.0.join(format!("out-{tag}"));
        fs::create_dir_all(&dir).unwrap();
        [
            dir.join("calls"),
            dir.join("where"),
            dir.join("slots"),
            dir.join("summary"),
        ]
    }

    /// The command that runs `tool` here, with the environment the Python harness gave the
    /// script. `extra` adds a variable (`Some`) or removes one (`None`).
    fn command(&self, tool: Tool, extra: Env, tag: &str) -> (Command, [PathBuf; 4]) {
        let [calls, wheres, slots, summary] = self.files(tag);
        let mut env: Vec<(String, String)> = git_env(&self.tmp.0);
        let mut set = |k: &str, v: Option<&str>| {
            env.retain(|(name, _)| name != k);
            if let Some(v) = v {
                env.push((k.to_string(), v.to_string()));
            }
        };
        set(
            "PATH",
            Some(&format!("{}:/usr/bin:/bin", self.bin.display())),
        );
        set("TMPDIR", Some(self.tmpdir.to_str().unwrap()));
        set("STUB_CALLS", Some(calls.to_str().unwrap()));
        set("STUB_WHERE", Some(wheres.to_str().unwrap()));
        set("STUB_SLOT_CALLS", Some(slots.to_str().unwrap()));
        if !self.candidate {
            // The in-place scenarios pin CI's behaviour unless one names its own mode.
            set("LOCAL_CHECK_MODE", Some("in-place"));
        }
        for (k, v) in extra {
            // `@file` names this run's own summary file, so the two runs do not share one.
            if *k == "GITHUB_STEP_SUMMARY" && *v == Some("@file") {
                set(k, Some(summary.to_str().unwrap()));
            } else {
                set(k, *v);
            }
        }
        let mut cmd = match tool {
            Tool::Shell => {
                let mut c = Command::new("bash");
                c.arg(self.repo.join("tools/local_check.sh"));
                c.current_dir(&self.tmp.0);
                c
            }
            Tool::Xtask => {
                let mut c = Command::new(XTASK);
                c.arg("local-check");
                c.current_dir(&self.repo);
                c
            }
        };
        cmd.env_clear().envs(env);
        (cmd, [calls, wheres, slots, summary])
    }

    fn run(&self, tool: Tool, extra: Env) -> Outcome {
        let tag = format!("{tool:?}");
        let (mut cmd, files) = self.command(tool, extra, &tag);
        let out = cmd.output().unwrap();
        outcome(&out, &files)
    }
}

fn read_lines(path: &Path) -> Vec<String> {
    fs::read_to_string(path)
        .map(|t| t.lines().map(String::from).collect())
        .unwrap_or_default()
}

fn outcome(out: &std::process::Output, files: &[PathBuf; 4]) -> Outcome {
    Outcome {
        code: out.status.code(),
        stdout: String::from_utf8_lossy(&out.stdout).into_owned(),
        stderr: String::from_utf8_lossy(&out.stderr).into_owned(),
        calls: read_lines(&files[0]),
        wheres: read_lines(&files[1]),
        slots: read_lines(&files[2]),
        summary: fs::read_to_string(&files[3]).unwrap_or_default(),
    }
}

/// Replaces the `len` characters after every `marker` with `#`, so a field that depends on the
/// clock, the process or a temporary name compares equal.
fn mask_field(text: &str, marker: &str, len: usize) -> String {
    let mut out = String::new();
    let mut rest = text;
    while let Some(at) = rest.find(marker) {
        let end = at + marker.len();
        out.push_str(&rest[..end]);
        out.push('#');
        let skip: usize = rest[end..].chars().take(len).map(char::len_utf8).sum();
        rest = &rest[end + skip..];
    }
    out.push_str(rest);
    out
}

fn mask(text: &str) -> String {
    let text = mask_field(text, " at=", 20);
    let text = mask_field(&text, " run=", 12);
    let text = mask_field(&text, " script=", 40);
    let text = mask_field(&text, "local-check-cand.", 6);
    mask_field(&text, ", commit ", 40)
}

fn masked(o: &Outcome) -> Outcome {
    Outcome {
        code: o.code,
        stdout: mask(&o.stdout),
        stderr: mask(&o.stderr),
        calls: o.calls.iter().map(|l| mask(l)).collect(),
        wheres: o.wheres.iter().map(|l| mask(l)).collect(),
        slots: o.slots.clone(),
        summary: o.summary.clone(),
    }
}

/// Runs the script and `xtask` in `fx` with `extra`, asserts they agree, and returns the
/// script's outcome for the scenario's own assertions.
fn parity(name: &str, fx: &Fx, extra: Env) -> Outcome {
    let shell = fx.run(Tool::Shell, extra);
    let xtask = fx.run(Tool::Xtask, extra);
    assert_eq!(
        masked(&xtask),
        masked(&shell),
        "scenario `{name}`: xtask (left) differs from the script (right)"
    );
    shell
}

fn last_line(o: &Outcome) -> &str {
    o.stdout
        .lines()
        .rfind(|l| !l.trim().is_empty())
        .unwrap_or("")
}

fn leftovers(fx: &Fx) -> Vec<String> {
    let mut names: Vec<String> = fs::read_dir(&fx.tmpdir)
        .unwrap()
        .map(|e| e.unwrap().file_name().to_string_lossy().into_owned())
        .collect();
    names.sort();
    names
}

// ---------------------------------------------------------------------------------------------
// in-place mode
// ---------------------------------------------------------------------------------------------

#[test]
fn in_place_runs_match_the_script() {
    const DRY: Env = &[
        ("LOCAL_CHECK_DRY_RUN", Some("1")),
        ("LOCAL_CHECK_BASE", Some("BASE")),
        ("LOCAL_CHECK_DIFF_HEAD", Some("HEAD")),
    ];
    let o = parity("dry run", &Fx::in_place(), DRY);
    assert_eq!(o.code, Some(1));
    assert_eq!(
        o.stdout.lines().filter(|l| l.starts_with("step ")).count(),
        8,
        "{}",
        o.stdout
    );
    assert!(o.calls.is_empty(), "a dry run ran {:?}", o.calls);
    assert!(last_line(&o).contains("result=fail"), "{}", o.stdout);

    let o = parity("clean pass", &Fx::in_place(), &[]);
    assert_eq!(o.code, Some(0), "{}{}", o.stdout, o.stderr);
    assert!(last_line(&o).contains(" result=pass cmd=tools/local_check.sh at="));
    assert_eq!(o.calls.len(), 7, "{:?}", o.calls);

    let o = parity(
        "first failure stops",
        &Fx::in_place(),
        &[("STUB_FAIL", Some("cargo clippy"))],
    );
    assert_eq!(o.code, Some(1));
    assert_eq!(o.calls.last().unwrap().split(' ').nth(1), Some("clippy"));
    assert!(o.stderr.contains("step clippy: exit 1"), "{}", o.stderr);

    let o = parity(
        "keep going runs every step",
        &Fx::in_place(),
        &[
            ("STUB_FAIL", Some("cargo fmt --check")),
            ("LOCAL_CHECK_KEEP_GOING", Some("1")),
        ],
    );
    assert_eq!(o.code, Some(1));
    assert_eq!(o.calls.len(), 7, "{:?}", o.calls);

    let fx = Fx::in_place();
    fs::write(fx.repo.join("README.md"), "changed\n").unwrap();
    let o = parity("modified tracked file is dirty", &fx, &[]);
    assert_eq!(o.code, Some(1));
    assert!(last_line(&o).contains(" dirty=1 cmd="), "{}", o.stdout);

    let fx = Fx::in_place();
    fs::write(fx.repo.join("new.txt"), "new\n").unwrap();
    let o = parity("untracked file is dirty", &fx, &[]);
    assert_eq!(o.code, Some(1));

    let fx = Fx::in_place();
    fs::write(fx.repo.join("debug.log"), "ignored\n").unwrap();
    let o = parity("an ignored file is not dirty", &fx, &[]);
    assert_eq!(o.code, Some(0), "{}{}", o.stdout, o.stderr);

    let fx = Fx::in_place();
    fx.commit(&fx.repo, &[("bad.txt", "trailing  \n")], "whitespace");
    let o = parity("diff-check failure", &fx, &[]);
    assert_eq!(o.code, Some(1));
    assert!(o.stderr.contains("step diff-check: exit"), "{}", o.stderr);

    let o = parity(
        "step summary rows",
        &Fx::in_place(),
        &[
            ("GITHUB_STEP_SUMMARY", Some("@file")),
            ("STUB_FAIL", Some("cargo clippy")),
            ("LOCAL_CHECK_KEEP_GOING", Some("1")),
        ],
    );
    assert_eq!(o.code, Some(1));
    assert!(
        o.summary.contains("| clippy |"),
        "no summary row: {}",
        o.summary
    );
    assert_eq!(o.summary.lines().count(), 8, "{}", o.summary);

    let o = parity(
        "base and diff head named",
        &Fx::in_place(),
        &[
            ("LOCAL_CHECK_BASE", Some("HEAD~1")),
            ("LOCAL_CHECK_DIFF_HEAD", Some("HEAD")),
            ("GITHUB_ACTIONS", Some("true")),
        ],
    );
    assert_eq!(o.code, Some(0), "{}{}", o.stdout, o.stderr);
    assert!(!o.stderr.contains("it is for CI"), "{}", o.stderr);
}

#[test]
fn modes_and_legs_match_the_script() {
    let o = parity(
        "unknown mode",
        &Fx::in_place(),
        &[("LOCAL_CHECK_MODE", Some("sideways"))],
    );
    assert_eq!(o.code, Some(2));
    assert!(o.stdout.is_empty(), "{}", o.stdout);
    assert!(o
        .stderr
        .contains("LOCAL_CHECK_MODE must be candidate or in-place"));

    for (leg, expect) in [("", 7usize), ("nextest", 3), ("rest", 5)] {
        let o = parity(
            &format!("leg `{leg}`"),
            &Fx::in_place(),
            &[
                ("LOCAL_CHECK_LEG", Some(leg)),
                ("GITHUB_ACTIONS", Some("true")),
            ],
        );
        assert_eq!(o.code, Some(0), "{leg}: {}{}", o.stdout, o.stderr);
        assert_eq!(o.calls.len(), expect, "{leg}: {:?}", o.calls);
    }

    let o = parity(
        "unknown leg",
        &Fx::in_place(),
        &[
            ("LOCAL_CHECK_LEG", Some("everything")),
            ("GITHUB_ACTIONS", Some("true")),
        ],
    );
    assert_eq!(o.code, Some(2));
    assert!(o.stdout.is_empty() && o.calls.is_empty());

    let o = parity(
        "a leg outside CI",
        &Fx::in_place(),
        &[("LOCAL_CHECK_LEG", Some("nextest"))],
    );
    assert_eq!(o.code, Some(2));
    assert!(o.stderr.contains("is for CI"), "{}", o.stderr);

    let o = parity(
        "a leg in candidate mode",
        &Fx::in_place(),
        &[
            ("LOCAL_CHECK_MODE", Some("candidate")),
            ("LOCAL_CHECK_LEG", Some("rest")),
        ],
    );
    assert_eq!(o.code, Some(2));
    assert!(o
        .stderr
        .contains("candidate mode always checks the full set"));

    let o = parity(
        "a leg ignores the other leg's failure",
        &Fx::in_place(),
        &[
            ("LOCAL_CHECK_LEG", Some("nextest")),
            ("GITHUB_ACTIONS", Some("true")),
            ("STUB_FAIL", Some("cargo clippy")),
        ],
    );
    assert_eq!(o.code, Some(0), "{}{}", o.stdout, o.stderr);

    let o = parity(
        "a leg stops at its own first failure",
        &Fx::in_place(),
        &[
            ("LOCAL_CHECK_LEG", Some("rest")),
            ("GITHUB_ACTIONS", Some("true")),
            ("STUB_FAIL", Some("cargo clippy")),
        ],
    );
    assert_eq!(o.code, Some(1));
    assert_eq!(o.calls.len(), 2, "{:?}", o.calls);

    let o = parity(
        "keep going finishes the leg",
        &Fx::in_place(),
        &[
            ("LOCAL_CHECK_LEG", Some("rest")),
            ("GITHUB_ACTIONS", Some("true")),
            ("STUB_FAIL", Some("cargo clippy")),
            ("LOCAL_CHECK_KEEP_GOING", Some("1")),
        ],
    );
    assert_eq!(o.code, Some(1));
    assert_eq!(o.calls.len(), 5, "{:?}", o.calls);
}

// ---------------------------------------------------------------------------------------------
// candidate mode
// ---------------------------------------------------------------------------------------------

#[test]
fn candidate_runs_match_the_script() {
    let fx = Fx::candidate();
    fx.commit(&fx.repo, &[("feature.txt", "feature\n")], "feature work");
    let o = parity("a head that merges cleanly", &fx, &[]);
    assert_eq!(o.code, Some(0), "{}{}", o.stdout, o.stderr);
    let line = last_line(&o);
    assert!(line.starts_with("local-check: head="), "{line}");
    assert!(line.contains(" result=pass base="), "{line}");
    assert!(line.contains(" rustc=1.99.0 "), "{line}");
    assert_eq!(
        leftovers(&fx),
        Vec::<String>::new(),
        "a candidate folder is left"
    );
    assert!(
        o.wheres.iter().all(|w| w.contains("local-check-cand.")),
        "steps did not run in the candidate: {:?}",
        o.wheres
    );

    let fx = Fx::candidate();
    fx.commit(&fx.repo, &[("README.md", "feature side\n")], "feature");
    fx.advance_origin(&[("README.md", "origin side\n")]);
    let o = parity("a conflict with the base tip", &fx, &[]);
    assert_eq!(o.code, Some(1));
    assert!(last_line(&o).contains(" result=fail base="), "{}", o.stdout);
    assert!(last_line(&o).contains(" tree=none "), "{}", o.stdout);
    assert!(o.calls.is_empty());

    let fx = Fx::candidate();
    fx.commit(&fx.repo, &[("a.txt", "a\n")], "feature");
    fx.replace_origin_history(&[("other.txt", "unrelated\n")]);
    let o = parity("unrelated history", &fx, &[]);
    assert_eq!(o.code, Some(1));
    assert!(last_line(&o).contains(" tree=none "), "{}", o.stdout);

    let fx = Fx::candidate();
    fx.commit(&fx.repo, &[("a.txt", "a\n")], "feature");
    fs::rename(fx.origin(), fx.tmp.0.join("origin-gone.git")).unwrap();
    let o = parity("a failed fetch", &fx, &[]);
    assert_eq!(o.code, Some(1));
    assert!(
        last_line(&o).contains(" base=none tree=none "),
        "{}",
        o.stdout
    );
    assert!(o.stderr.contains("fetching origin/"), "{}", o.stderr);

    let fx = Fx::candidate();
    fx.commit(&fx.repo, &[("a.txt", "a\n")], "feature");
    fs::write(fx.repo.join("a.txt"), "dirty\n").unwrap();
    let o = parity("a dirty working folder", &fx, &[]);
    assert_eq!(o.code, Some(1));
    assert!(o.calls.is_empty());
    assert!(
        o.stderr.contains("commit or stash them first"),
        "{}",
        o.stderr
    );

    let fx = Fx::candidate();
    fx.commit(&fx.repo, &[("a.txt", "a\n")], "feature");
    let o = parity("a dry run", &fx, &[("LOCAL_CHECK_DRY_RUN", Some("1"))]);
    assert_eq!(o.code, Some(1));
    assert!(o.calls.is_empty());
    assert!(
        last_line(&o).contains(" result=fail base=none tree=none "),
        "{}",
        o.stdout
    );

    for name in ["-x", "a..b", "has space", "semi;colon", ""] {
        let fx = Fx::candidate();
        fx.commit(&fx.repo, &[("a.txt", "a\n")], "feature");
        let o = parity(
            &format!("base branch name `{name}`"),
            &fx,
            &[("LOCAL_CHECK_BASE_BRANCH", Some(name))],
        );
        // An empty name falls back to the default, so that one passes.
        if name.is_empty() {
            assert_eq!(o.code, Some(0), "{}{}", o.stdout, o.stderr);
        } else {
            assert_eq!(o.code, Some(1), "{name}: {}{}", o.stdout, o.stderr);
            assert!(
                o.stderr.contains("is not a branch name"),
                "{name}: {}",
                o.stderr
            );
        }
    }

    // The word of `rustc --version`: 32 characters are the longest accepted.
    let long32 = "x".repeat(32);
    let long33 = "x".repeat(33);
    for (out, exit, passes) in [
        (format!("rustc {long32} (x)"), "0", true),
        (format!("rustc {long33} (x)"), "0", false),
        ("rustc bad/word (x)".to_string(), "0", false),
        ("rustc".to_string(), "0", false),
        (String::new(), "0", false),
        ("rustc 1.2.3 (x)".to_string(), "1", false),
    ] {
        let fx = Fx::candidate();
        fx.commit(&fx.repo, &[("a.txt", "a\n")], "feature");
        let o = parity(
            &format!("rustc prints `{out}` and exits {exit}"),
            &fx,
            &[
                ("STUB_RUSTC_OUT", Some(&out)),
                ("STUB_RUSTC_EXIT", Some(exit)),
            ],
        );
        assert_eq!(
            o.code,
            Some(i32::from(!passes)),
            "`{out}`: {}{}",
            o.stdout,
            o.stderr
        );
    }

    let fx = Fx::candidate();
    fx.commit(&fx.repo, &[("a.txt", "a\n")], "feature");
    fs::create_dir(fx.tmpdir.join("local-check-cand.OLD123")).unwrap();
    let o = parity("an earlier candidate folder", &fx, &[]);
    assert_eq!(o.code, Some(0), "{}{}", o.stdout, o.stderr);
    assert!(
        o.stderr.contains("an earlier candidate folder exists"),
        "{}",
        o.stderr
    );
    assert!(fx.tmpdir.join("local-check-cand.OLD123").exists());

    let fx = Fx::candidate();
    fx.commit(&fx.repo, &[("a.txt", "a\n")], "feature");
    let o = parity("a step fails", &fx, &[("STUB_FAIL", Some("cargo clippy"))]);
    assert_eq!(o.code, Some(1));
    assert!(last_line(&o).contains(" result=fail base="), "{}", o.stdout);
    assert!(!last_line(&o).contains("tree=none"), "{}", o.stdout);
    assert_eq!(leftovers(&fx), Vec::<String>::new());

    let fx = Fx::candidate();
    fx.commit(&fx.repo, &[("bad.txt", "trailing  \n")], "whitespace");
    let o = parity("the diff check reads base...head", &fx, &[]);
    assert_eq!(o.code, Some(1));
    assert!(o.stderr.contains("step diff-check: exit"), "{}", o.stderr);

    let fx = Fx::candidate();
    fx.commit(&fx.repo, &[("a.txt", "a\n")], "feature");
    let o = parity(
        "the old overrides are ignored",
        &fx,
        &[
            ("LOCAL_CHECK_BASE", Some("HEAD~1")),
            ("LOCAL_CHECK_DIFF_HEAD", Some("HEAD")),
        ],
    );
    assert_eq!(o.code, Some(0), "{}{}", o.stdout, o.stderr);
    assert!(
        o.stderr.contains("candidate mode ignores LOCAL_CHECK_BASE"),
        "{}",
        o.stderr
    );

    // A rule the base adds that a head cannot know: the head passes alone, the merge fails.
    let fx = Fx::candidate();
    fx.commit(&fx.repo, &[("forbidden.txt", "x\n")], "feature");
    fx.advance_origin(&[("rule_forbids_forbidden.txt", "no\n")]);
    let o = parity("a rule the base added", &fx, &[]);
    assert_eq!(o.code, Some(1));
    assert!(last_line(&o).contains(" result=fail "), "{}", o.stdout);
}

// ---------------------------------------------------------------------------------------------
// slots
// ---------------------------------------------------------------------------------------------

#[test]
fn slot_runs_match_the_script() {
    let o = parity("default slots", &Fx::with_slot(), &[]);
    assert_eq!(o.code, Some(0), "{}{}", o.stdout, o.stderr);
    // test, doctest and clippy are heavy; unittest is the one python step.
    assert_eq!(o.slots.len(), 4, "{:?}", o.slots);
    assert!(o.slots[0].contains("--max 3 --min-free-gb 50 --name local-check -- bash -c"));
    assert!(o.slots[3].contains("--dir "), "{:?}", o.slots);
    assert!(
        o.slots[3].contains("--name local-check-py"),
        "{:?}",
        o.slots
    );

    let fx = Fx::with_slot();
    let cache = fx.tmp.0.join("cache");
    parity(
        "XDG_CACHE_HOME names the python slot folder",
        &fx,
        &[("XDG_CACHE_HOME", Some(cache.to_str().unwrap()))],
    );

    let o = parity(
        "folder, limits and name are settable",
        &Fx::with_slot(),
        &[
            ("MOAI_PYTHON_SLOT_DIR", Some("/slots/py")),
            ("MOAI_PYTHON_SLOT_MAX", Some("5")),
            ("MOAI_HEAVY_SLOT_MAX", Some("2")),
            ("MOAI_HEAVY_SLOT_MIN_FREE_GB", Some("7")),
            ("MOAI_HEAVY_SLOT_NAME", Some("mine")),
        ],
    );
    assert!(o.slots[3].contains("--dir /slots/py --max 5 --min-free-gb 7 --name mine-py"));

    parity(
        "empty settings mean the defaults",
        &Fx::with_slot(),
        &[
            ("MOAI_PYTHON_SLOT_DIR", Some("")),
            ("MOAI_PYTHON_SLOT_MAX", Some("")),
            ("XDG_CACHE_HOME", Some("")),
        ],
    );

    let o = parity(
        "no HOME and no XDG_CACHE_HOME: the python step runs without a slot",
        &Fx::with_slot(),
        &[("HOME", None), ("XDG_CACHE_HOME", None)],
    );
    assert_eq!(o.code, Some(0), "{}{}", o.stdout, o.stderr);
    assert_eq!(o.slots.len(), 3, "{:?}", o.slots);
    assert!(o.stderr.contains("no python slot folder"), "{}", o.stderr);

    let o = parity(
        "an empty HOME falls back the same way",
        &Fx::with_slot(),
        &[("HOME", Some("")), ("XDG_CACHE_HOME", Some(""))],
    );
    assert_eq!(o.slots.len(), 3, "{:?}", o.slots);

    let o = parity(
        "the slot refuses",
        &Fx::with_slot(),
        &[("STUB_SLOT_REFUSE", Some("1"))],
    );
    assert_eq!(o.code, Some(1));
    assert!(o.stderr.contains("step test: exit 75"), "{}", o.stderr);

    // Without a `moai` on PATH every step runs directly, which the plain scenarios cover.
}

// ---------------------------------------------------------------------------------------------
// the tool list and the signal
// ---------------------------------------------------------------------------------------------

#[test]
fn an_unknown_tool_lists_the_tools_and_exits_2() {
    for args in [vec![], vec!["no-such-tool"]] {
        let out = Command::new(XTASK).args(&args).output().unwrap();
        assert_eq!(out.status.code(), Some(2), "{args:?}");
        assert!(out.stdout.is_empty());
        let stderr = String::from_utf8_lossy(&out.stderr);
        assert!(stderr.contains("local-check"), "{args:?}: {stderr}");
    }
}

fn wait_for(path: &Path, what: &str) {
    let start = Instant::now();
    while !path.exists() {
        assert!(
            start.elapsed() < Duration::from_secs(30),
            "{what} never started"
        );
        std::thread::sleep(Duration::from_millis(50));
    }
}

/// A stub `cargo` that sleeps while the tool is signalled: the script defers its `trap` until
/// the foreground step ends, and `xtask` must too.
fn signalled(fx: &Fx, tool: Tool, extra: Env) -> Outcome {
    let tag = format!("sig-{tool:?}");
    let mut vars: Vec<(&str, Option<&str>)> = vec![("STUB_SLEEP", Some("2"))];
    vars.extend(extra.iter().copied());
    let (mut cmd, files) = fx.command(tool, &vars, &tag);
    let out = fx.tmp.0.join(format!("sig-out-{tag}"));
    let err = fx.tmp.0.join(format!("sig-err-{tag}"));
    cmd.stdout(fs::File::create(&out).unwrap());
    cmd.stderr(fs::File::create(&err).unwrap());
    cmd.stdin(Stdio::null());
    let mut child = cmd.spawn().unwrap();
    wait_for(&files[0], "the first step");
    let sent = Command::new("kill")
        .args(["-TERM", &child.id().to_string()])
        .status()
        .unwrap();
    assert!(sent.success());
    let status = child.wait().unwrap();
    Outcome {
        code: status.code(),
        stdout: fs::read_to_string(&out).unwrap(),
        stderr: fs::read_to_string(&err).unwrap(),
        calls: read_lines(&files[0]),
        wheres: read_lines(&files[1]),
        slots: read_lines(&files[2]),
        summary: String::new(),
    }
}

#[test]
fn a_terminate_signal_ends_with_a_fail_line_and_cleans_up() {
    let fx = Fx::in_place();
    let shell = signalled(&fx, Tool::Shell, &[]);
    let xtask = signalled(&fx, Tool::Xtask, &[]);
    assert_eq!(
        masked(&xtask),
        masked(&shell),
        "in-place: xtask (left), script (right)"
    );
    assert_eq!(shell.code, Some(1), "{}{}", shell.stdout, shell.stderr);
    assert!(
        last_line(&shell).contains(" result=fail "),
        "{}",
        shell.stdout
    );
    assert_eq!(
        shell.calls.len(),
        1,
        "the steps after the signal ran: {:?}",
        shell.calls
    );

    let fx = Fx::candidate();
    fx.commit(&fx.repo, &[("a.txt", "a\n")], "feature");
    let shell = signalled(&fx, Tool::Shell, &[]);
    assert_eq!(
        leftovers(&fx),
        Vec::<String>::new(),
        "the script left a candidate folder"
    );
    let xtask = signalled(&fx, Tool::Xtask, &[]);
    assert_eq!(
        leftovers(&fx),
        Vec::<String>::new(),
        "xtask left a candidate folder"
    );
    assert_eq!(
        masked(&xtask),
        masked(&shell),
        "candidate: xtask (left), script (right)"
    );
    assert_eq!(shell.code, Some(1), "{}{}", shell.stdout, shell.stderr);
    assert!(
        last_line(&shell).contains(" result=fail base="),
        "{}",
        shell.stdout
    );
    let worktrees = fx.git(&fx.repo, &["worktree", "list", "--porcelain"]);
    assert_eq!(worktrees.matches("worktree ").count(), 1, "{worktrees}");
}
