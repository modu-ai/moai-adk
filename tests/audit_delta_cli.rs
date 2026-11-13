//! Local three-SHA audit behavior through the real CLI, with disposable Git repositories.
mod common;

use serde_json::{json, Value};
use std::collections::BTreeMap;
use std::fs;
use std::io::Write;
use std::path::{Path, PathBuf};
use std::process::{Command, Output, Stdio};

const BIN: &str = crate::common::MOAI_BIN;
const A: &str = "pass-file.txt";
const B: &str = "tip-file.txt";
const PASS: &str = "pass branch change\n";
const TIP: &str = "tip branch change\n";
const MANUAL: &str = "pass branch change\nAUDIT_DELTA_MANUAL_LITERAL_2403\n";

fn command(dir: &Path) -> Command {
    let mut cmd = common::git::command();
    for key in [
        "GIT_DIR",
        "GIT_WORK_TREE",
        "GIT_COMMON_DIR",
        "GIT_INDEX_FILE",
        "GIT_OBJECT_DIRECTORY",
        "GIT_ALTERNATE_OBJECT_DIRECTORIES",
        "GIT_CONFIG",
        "GIT_CONFIG_COUNT",
        "GIT_NAMESPACE",
    ] {
        cmd.env_remove(key);
    }
    cmd.current_dir(dir)
        .args([
            "-c",
            "user.name=fixture",
            "-c",
            "user.email=fixture@example.invalid",
        ])
        .env("GIT_OPTIONAL_LOCKS", "0")
        .env("GIT_TERMINAL_PROMPT", "0")
        .env("GIT_AUTHOR_DATE", "2026-01-01T00:00:00 +0000")
        .env("GIT_COMMITTER_DATE", "2026-01-01T00:00:00 +0000")
        .stdin(Stdio::null());
    cmd
}

fn git(dir: &Path, args: &[&str]) -> String {
    let out = command(dir).args(args).output().expect("spawn fixture Git");
    assert!(
        out.status.success(),
        "git {args:?}: {}",
        String::from_utf8_lossy(&out.stderr)
    );
    String::from_utf8(out.stdout).unwrap().trim().to_owned()
}

fn git_input(dir: &Path, args: &[&str], input: &[u8]) -> String {
    let mut child = command(dir)
        .args(args)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    child.stdin.take().unwrap().write_all(input).unwrap();
    let out = child.wait_with_output().unwrap();
    assert!(
        out.status.success(),
        "git {args:?}: {}",
        String::from_utf8_lossy(&out.stderr)
    );
    String::from_utf8(out.stdout).unwrap().trim().to_owned()
}

struct Fixture {
    _dir: tempfile::TempDir,
    project: PathBuf,
    home: PathBuf,
    tmp: PathBuf,
    root: String,
    pass: String,
    tip: String,
}

impl Fixture {
    fn new(conflict: bool, attributes: bool) -> Self {
        let dir = tempfile::TempDir::new().unwrap();
        let base = fs::canonicalize(dir.path()).unwrap();
        let project = base.join("project");
        let home = base.join("home");
        let tmp = base.join("tmp");
        for path in [&project, &home, &tmp] {
            fs::create_dir(path).unwrap();
        }
        git(
            &project,
            &["init", "-q", "--object-format=sha1", "-b", "main"],
        );
        fs::write(project.join(A), "base a\n").unwrap();
        fs::write(project.join(B), "base b\n").unwrap();
        if attributes {
            fs::write(
                project.join(".gitattributes"),
                "*.txt merge=audit-tripwire\n",
            )
            .unwrap();
        }
        git(&project, &["add", "."]);
        git(&project, &["commit", "-q", "-m", "root"]);
        let root = git(&project, &["rev-parse", "HEAD"]);
        git(&project, &["switch", "-q", "-c", "pass"]);
        fs::write(project.join(A), PASS).unwrap();
        git(&project, &["add", "."]);
        git(&project, &["commit", "-q", "-m", "pass"]);
        let pass = git(&project, &["rev-parse", "HEAD"]);
        git(&project, &["switch", "-q", "-c", "tip", &root]);
        fs::write(project.join(if conflict { A } else { B }), TIP).unwrap();
        git(&project, &["add", "."]);
        git(&project, &["commit", "-q", "-m", "tip"]);
        let tip = git(&project, &["rev-parse", "HEAD"]);
        Self {
            _dir: dir,
            project,
            home,
            tmp,
            root,
            pass,
            tip,
        }
    }

    fn clean_merge(&self) -> String {
        git(&self.project, &["switch", "-q", "pass"]);
        git(
            &self.project,
            &["merge", "-q", "--no-ff", "-m", "actual merge", &self.tip],
        );
        git(&self.project, &["rev-parse", "HEAD"])
    }

    fn commit_tree(&self, tree: &str, parents: &[&str]) -> String {
        let mut args = vec!["commit-tree", tree, "-m", "actual committed merge"];
        for parent in parents {
            args.extend(["-p", *parent]);
        }
        git(&self.project, &args)
    }

    fn manual_merge(&self) -> String {
        // Never merge or merge-tree in this source: the automatic root tree must not exist.
        git(&self.project, &["switch", "-q", "pass"]);
        fs::write(self.project.join(A), MANUAL).unwrap();
        fs::write(self.project.join(B), TIP).unwrap();
        git(&self.project, &["add", "."]);
        let actual = git(&self.project, &["write-tree"]);
        self.commit_tree(&actual, &[&self.pass, &self.tip])
    }

    fn automatic_tree_without_writing(&self) -> String {
        // Independently encode the two-entry Git tree from the branch blobs. hash-object
        // deliberately has no -w, so even this expected root object stays absent.
        let mut raw = Vec::new();
        for (name, commit) in [(A, &self.pass), (B, &self.tip)] {
            let oid = git(&self.project, &["rev-parse", &format!("{commit}:{name}")]);
            raw.extend_from_slice(b"100644 ");
            raw.extend_from_slice(name.as_bytes());
            raw.push(0);
            for i in (0..40).step_by(2) {
                raw.push(u8::from_str_radix(&oid[i..i + 2], 16).unwrap());
            }
        }
        git_input(
            &self.project,
            &["hash-object", "-t", "tree", "--stdin"],
            &raw,
        )
    }

    fn invoke(&self, project: &Path, head: &str) -> Output {
        self.invoke_with_tmp(project, head, &self.tmp)
    }

    fn invoke_with_tmp(&self, project: &Path, head: &str, tmp: &Path) -> Output {
        let mut cmd = Command::new(BIN);
        for key in [
            "GIT_DIR",
            "GIT_WORK_TREE",
            "GIT_COMMON_DIR",
            "GIT_INDEX_FILE",
            "GIT_OBJECT_DIRECTORY",
            "GIT_ALTERNATE_OBJECT_DIRECTORIES",
            "GIT_CONFIG",
            "GIT_CONFIG_COUNT",
            "GIT_NAMESPACE",
        ] {
            cmd.env_remove(key);
        }
        cmd.arg("audit-delta")
            .arg(project)
            .args([
                "--pass-head",
                &self.pass,
                "--tip",
                &self.tip,
                "--head",
                head,
            ])
            .env("HOME", &self.home)
            .env("XDG_CONFIG_HOME", &self.home)
            .env("TMPDIR", tmp)
            .env("TMP", tmp)
            .env("TEMP", tmp)
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_SYSTEM", "/dev/null")
            .env("GIT_TERMINAL_PROMPT", "0")
            .env("LC_ALL", "C")
            .stdin(Stdio::null())
            .output()
            .expect("spawn actual moai-cli")
    }

    fn report(&self, project: &Path, head: &str, exit: i32, status: &str) -> Value {
        let actual_tree = git(&self.project, &["rev-parse", &format!("{head}^{{tree}}")]);
        let out = self.invoke(project, head);
        // Deliberately before JSON parsing: the missing base command must be a behavioral
        // RED assertion (usage exit 2), not a JSON/setup panic.
        assert_eq!(
            out.status.code(),
            Some(exit),
            "audit-delta must compute {status}; stdout={} stderr={}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        );
        let envelope: Value = serde_json::from_slice(&out.stdout).expect("audit-delta JSON report");
        assert_eq!(
            envelope["ok"], true,
            "computed ineligible outcomes are DATA: {envelope}"
        );
        assert_eq!(envelope["operation"], "audit.delta");
        let r = envelope["result"].clone();
        assert_eq!(r["schema_version"], 1);
        assert_eq!(r["status"], status);
        assert_eq!(r["pass_head"], self.pass);
        assert_eq!(r["tip"], self.tip);
        assert_eq!(r["head"], head);
        assert_eq!(r["actual_tree"], actual_tree);
        r
    }
}

fn changes(paths: &[&str]) -> Value {
    json!(paths
        .iter()
        .map(|path| json!({"path": path, "status": "M"}))
        .collect::<Vec<_>>())
}

fn assert_clean(r: &Value, fx: &Fixture) {
    assert_eq!(r["actual_parents"], json!([fx.pass, fx.tip]));
    assert_eq!(r["replayed_tree"], r["actual_tree"]);
    assert_eq!(r["replay_exit"], 0);
    assert_eq!(r["manual_delta"], false);
    assert_eq!(r["manual_text_hunks"], false);
    assert_eq!(r["tip_changed_files"], changes(&[A]));
    assert_eq!(r["pass_changed_files"], changes(&[B]));
    assert_eq!(r["manual_changed_files"], json!([]));
    assert_eq!(r["conflicted_files"], json!([]));
    assert_eq!(r["refusal"], Value::Null);
}

fn assert_refused(r: &Value, parents: &[&str], code: &str) {
    assert_eq!(r["actual_parents"], json!(parents));
    assert_eq!(r["refusal"]["code"], code);
    for key in [
        "replayed_tree",
        "replay_exit",
        "manual_delta",
        "manual_text_hunks",
    ] {
        assert_eq!(
            r[key],
            Value::Null,
            "unrun replay must not pretend false: {key}"
        );
    }
    assert_eq!(r["manual_changed_files"], json!([]));
    assert_eq!(r["conflicted_files"], json!([]));
}

// Raw bytes, including loose objects, refs, logs, config, index, HEAD, tracked and
// untracked content. Comparing the entire source also detects new lock/object files.
fn snapshot(root: &Path) -> BTreeMap<PathBuf, Vec<u8>> {
    fn visit(root: &Path, dir: &Path, files: &mut BTreeMap<PathBuf, Vec<u8>>) {
        for entry in fs::read_dir(dir).unwrap() {
            let path = entry.unwrap().path();
            if path.is_dir() {
                visit(root, &path, files);
            } else {
                files.insert(
                    path.strip_prefix(root).unwrap().to_owned(),
                    fs::read(path).unwrap(),
                );
            }
        }
    }
    let mut files = BTreeMap::new();
    visit(root, root, &mut files);
    files
}

#[test]
fn clean_actual_two_parent_merge_is_pure_and_exits_zero() {
    let fx = Fixture::new(false, false);
    let expected = fx.automatic_tree_without_writing();
    let head = fx.clean_merge();
    let r = fx.report(&fx.project, &head, 0, "pure_merge");
    assert_clean(&r, &fx);
    assert_eq!(r["actual_tree"], expected);
}

#[test]
fn manual_commit_tree_hunk_is_delta_and_preserves_dirty_source_and_object_database() {
    let fx = Fixture::new(false, false);
    let expected = fx.automatic_tree_without_writing();
    let head = fx.manual_merge();
    let absent = command(&fx.project)
        .args(["cat-file", "-e", &expected])
        .output()
        .unwrap();
    assert!(
        !absent.status.success(),
        "automatic root must be absent before the SUT"
    );
    fs::write(fx.project.join(A), "staged dirty bytes\n").unwrap();
    git(&fx.project, &["add", A]);
    fs::write(fx.project.join(A), "unstaged dirty bytes\0\n").unwrap();
    fs::write(fx.project.join("untracked.txt"), "untracked bytes\0\n").unwrap();
    let before = snapshot(&fx.project);
    let r = fx.report(&fx.project, &head, 1, "manual_delta");
    assert_eq!(
        snapshot(&fx.project),
        before,
        "no source ODB/index/HEAD/refs/config/content writes"
    );
    assert_eq!(r["actual_parents"], json!([fx.pass, fx.tip]));
    assert_eq!(r["replayed_tree"], expected);
    assert_ne!(r["replayed_tree"], r["actual_tree"]);
    assert_eq!(r["replay_exit"], 0);
    assert_eq!(r["manual_delta"], true);
    assert_eq!(r["manual_text_hunks"], true);
    assert_eq!(r["manual_changed_files"], changes(&[A]));
    assert_eq!(r["tip_changed_files"], changes(&[A]));
    assert_eq!(r["pass_changed_files"], changes(&[A, B]));
    assert_eq!(r["refusal"], Value::Null);
    assert_eq!(r["conflicted_files"], json!([]));
}

#[test]
fn reversed_actual_parents_are_preserved_and_refused_without_replay() {
    let fx = Fixture::new(false, false);
    let clean = fx.clean_merge();
    let tree = git(&fx.project, &["rev-parse", &format!("{clean}^{{tree}}")]);
    let head = fx.commit_tree(&tree, &[&fx.tip, &fx.pass]);
    let r = fx.report(&fx.project, &head, 1, "refused");
    assert_refused(&r, &[&fx.tip, &fx.pass], "parent_mismatch");
}

#[test]
fn three_actual_parents_are_refused_as_octopus() {
    let fx = Fixture::new(false, false);
    let tree = git(
        &fx.project,
        &["rev-parse", &format!("{}^{{tree}}", fx.pass)],
    );
    let head = fx.commit_tree(&tree, &[&fx.pass, &fx.tip, &fx.root]);
    let r = fx.report(&fx.project, &head, 1, "refused");
    assert_refused(&r, &[&fx.pass, &fx.tip, &fx.root], "octopus");
}

#[test]
fn one_parent_and_root_commits_are_refused_as_not_merge() {
    let fx = Fixture::new(false, false);
    let r = fx.report(&fx.project, &fx.pass, 1, "refused");
    assert_refused(&r, &[&fx.root], "not_merge");
    let r = fx.report(&fx.project, &fx.root, 1, "refused");
    assert_refused(&r, &[], "not_merge");
}

#[test]
fn conflicting_replay_is_exit_one_report_data_not_transport_error() {
    let fx = Fixture::new(true, false);
    let tree = git(
        &fx.project,
        &["rev-parse", &format!("{}^{{tree}}", fx.pass)],
    );
    let head = fx.commit_tree(&tree, &[&fx.pass, &fx.tip]);
    let before = snapshot(&fx.project);
    let r = fx.report(&fx.project, &head, 1, "replay_conflict");
    assert_eq!(snapshot(&fx.project), before);
    assert_eq!(r["actual_parents"], json!([fx.pass, fx.tip]));
    assert_eq!(r["replay_exit"], 1);
    assert_eq!(r["conflicted_files"], json!([A]));
    assert_eq!(r["refusal"], Value::Null);
    let replay = r["replayed_tree"]
        .as_str()
        .expect("conflict tree is still DATA");
    assert_eq!(replay.len(), 40);
    assert!(replay
        .bytes()
        .all(|b| b.is_ascii_digit() || (b'a'..=b'f').contains(&b)));
    assert_ne!(r["replayed_tree"], r["actual_tree"]);
    assert_eq!(r["tip_changed_files"], changes(&[A]));
    assert_eq!(r["pass_changed_files"], json!([]));
}

#[test]
fn linked_worktree_reads_common_objects_without_changing_either_worktree() {
    let fx = Fixture::new(false, false);
    let head = fx.clean_merge();
    let linked = fx.project.parent().unwrap().join("linked worktree");
    git(
        &fx.project,
        &[
            "worktree",
            "add",
            "-q",
            "--detach",
            linked.to_str().unwrap(),
            &fx.pass,
        ],
    );
    assert!(linked.join(".git").is_file());
    fs::write(linked.join(B), "linked dirty bytes\n").unwrap();
    let source_before = snapshot(&fx.project);
    let linked_before = snapshot(&linked);
    let r = fx.report(&linked, &head, 0, "pure_merge");
    assert_clean(&r, &fx);
    assert_eq!(
        snapshot(&fx.project),
        source_before,
        "common object store and primary checkout unchanged"
    );
    assert_eq!(
        snapshot(&linked),
        linked_before,
        "linked HEAD/index/content unchanged"
    );
}

#[cfg(unix)]
#[test]
fn source_configured_external_merge_driver_is_never_invoked() {
    use std::os::unix::fs::PermissionsExt;
    let fx = Fixture::new(true, true);
    let marker = fx.project.parent().unwrap().join("driver-invoked");
    let driver = fx.project.parent().unwrap().join("merge-driver.sh");
    // A bounded collaborator: if invoked it creates evidence immediately and returns
    // success, incorrectly turning the builtin conflict into a clean replay.
    fs::write(&driver, "#!/bin/sh\nprintf 'invoked\\n' > \"$1\"\nexit 0\n").unwrap();
    fs::set_permissions(&driver, fs::Permissions::from_mode(0o700)).unwrap();
    let quote = |p: &Path| format!("'{}'", p.to_str().unwrap().replace('\'', "'\\''"));
    let driver_command = format!("{} {} %O %A %B", quote(&driver), quote(&marker));
    git(
        &fx.project,
        &["config", "merge.audit-tripwire.driver", &driver_command],
    );
    let tree = git(
        &fx.project,
        &["rev-parse", &format!("{}^{{tree}}", fx.pass)],
    );
    let head = fx.commit_tree(&tree, &[&fx.pass, &fx.tip]);
    let before = snapshot(&fx.project);
    let r = fx.report(&fx.project, &head, 1, "replay_conflict");
    assert!(!marker.exists(), "source external merge driver executed");
    assert_eq!(snapshot(&fx.project), before);
    assert_eq!(r["actual_parents"], json!([fx.pass, fx.tip]));
    assert_eq!(r["replay_exit"], 1);
    assert_eq!(r["conflicted_files"], json!([A]));
    assert_eq!(r["refusal"], Value::Null);
}

#[test]
fn temporary_root_inside_repository_is_blocked_even_from_subdirectory() {
    let fx = Fixture::new(false, false);
    let head = fx.clean_merge();
    let subdir = fx.project.join("subdir");
    let inside_tmp = fx.project.join("inside-tmp");
    fs::create_dir(&subdir).unwrap();
    fs::create_dir(&inside_tmp).unwrap();
    let before = snapshot(&fx.project);
    let control = fx.report(&subdir, &head, 0, "pure_merge");
    assert_clean(&control, &fx);
    assert_eq!(snapshot(&fx.project), before);

    let out = fx.invoke_with_tmp(&subdir, &head, &inside_tmp);
    assert_eq!(
        out.status.code(),
        Some(1),
        "source-contained temporary root must be blocked; stdout={} stderr={}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    let envelope: Value = serde_json::from_slice(&out.stdout).expect("blocked audit-delta JSON");
    assert_eq!(envelope["operation"], "audit.delta");
    assert_eq!(envelope["ok"], false);
    assert_eq!(envelope["error"]["kind"], "blocked");
    assert_eq!(
        snapshot(&fx.project),
        before,
        "blocked scratch must not change source contents"
    );
    assert_eq!(fs::read_dir(&inside_tmp).unwrap().count(), 0);
}
#[test]
fn temporary_root_inside_a_bare_source_is_blocked_by_the_common_boundary() {
    // #2506: with the redundant `project` clause removed, the bare case must still
    // refuse a scratch inside the source — here the common directory IS the bare
    // project, so only that clause can catch it.
    let fx = Fixture::new(false, false);
    let head = fx.clean_merge();
    let bare = fx.project.parent().unwrap().join("bare-source.git");
    git(
        &fx.project,
        &[
            "clone",
            "-q",
            "--bare",
            &fx.project.display().to_string(),
            &bare.display().to_string(),
        ],
    );
    let inside = bare.join("inside-tmp");
    fs::create_dir(&inside).unwrap();
    let before = snapshot(&bare);
    let out = fx.invoke_with_tmp(&bare, &head, &inside);
    assert_eq!(
        out.status.code(),
        Some(1),
        "bare source-contained temporary root must be blocked; stdout={} stderr={}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    let envelope: Value = serde_json::from_slice(&out.stdout).expect("blocked audit-delta JSON");
    assert_eq!(envelope["operation"], "audit.delta");
    assert_eq!(envelope["ok"], false);
    assert_eq!(envelope["error"]["kind"], "blocked");
    assert_eq!(
        snapshot(&bare),
        before,
        "blocked scratch must not touch the bare source"
    );
    assert_eq!(fs::read_dir(&inside).unwrap().count(), 0);
}
#[test]
fn tmp_inside_project_is_blocked_when_core_worktree_points_elsewhere() {
    // #2529 B1: with core.worktree at another folder, --show-toplevel is that
    // folder and --git-common-dir is PROJECT/.git, so neither the worktree nor
    // the common clause sees a temporary root under PROJECT. Only the PROJECT
    // clause keeps the scratch out of the audited tree.
    let fx = Fixture::new(false, false);
    let base = fx.project.parent().unwrap();
    let holder = base.join("worktree-holder");
    fs::create_dir(&holder).unwrap();
    git(
        &holder,
        &["init", "-q", "--object-format=sha1", "-b", "main"],
    );
    let elsewhere = base.join("elsewhere");
    fs::create_dir(&elsewhere).unwrap();
    git(
        &holder,
        &["config", "core.worktree", &elsewhere.display().to_string()],
    );
    let sub = holder.join("sub");
    fs::create_dir(&sub).unwrap();
    let inside = sub.join("tmp");
    fs::create_dir(&inside).unwrap();
    let before = snapshot(&holder);
    let out = fx.invoke_with_tmp(&sub, &fx.pass, &inside);
    assert_eq!(
        out.status.code(),
        Some(1),
        "core.worktree-shaped source-contained temporary root must be blocked; stdout={} stderr={}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    let envelope: Value = serde_json::from_slice(&out.stdout).expect("blocked audit-delta JSON");
    assert_eq!(envelope["operation"], "audit.delta");
    assert_eq!(envelope["ok"], false);
    assert_eq!(envelope["error"]["kind"], "blocked");
    assert_eq!(
        snapshot(&holder),
        before,
        "blocked scratch must not touch the core.worktree holder"
    );
    assert_eq!(fs::read_dir(&inside).unwrap().count(), 0);
}

#[test]
fn filesystem_failures_keep_typed_io_error_kind() {
    let fx = Fixture::new(false, false);
    let head = fx.clean_merge();
    let missing = fx.project.parent().unwrap().join("nonexistent-project");
    assert!(!missing.exists());
    let before = snapshot(&fx.project);
    let out = fx.invoke(&missing, &head);
    assert_eq!(
        out.status.code(),
        Some(1),
        "missing project must yield a typed I/O failure; stdout={} stderr={}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    let envelope: Value = serde_json::from_slice(&out.stdout).expect("I/O audit-delta JSON");
    assert_eq!(envelope["operation"], "audit.delta");
    assert_eq!(envelope["ok"], false);
    assert_eq!(envelope["error"]["kind"], "io");
    assert_eq!(snapshot(&fx.project), before);
    assert!(!missing.exists());
}

#[test]
fn missing_full_commit_oid_is_invalid_input_without_source_writes() {
    let fx = Fixture::new(false, false);
    let missing_head = "0000000000000000000000000000000000000000";
    let before = snapshot(&fx.project);
    let out = fx.invoke(&fx.project, missing_head);
    assert_eq!(
        out.status.code(),
        Some(1),
        "missing full commit OID must yield invalid input; stdout={} stderr={}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    let envelope: Value =
        serde_json::from_slice(&out.stdout).expect("missing-object audit-delta JSON");
    assert_eq!(envelope["operation"], "audit.delta");
    assert_eq!(envelope["ok"], false);
    assert_eq!(envelope["error"]["kind"], "invalid_input");
    assert_eq!(snapshot(&fx.project), before);
}
