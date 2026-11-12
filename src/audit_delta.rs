//! Local three-commit merge replay, with all object writes confined to owned scratch.
//!
//! Replay uses Git's built-in default merge semantics, not the source machine's merge
//! drivers, global attributes, or configuration. Repository attributes still select Git's
//! built-in drivers; an unknown custom driver falls back to Git's built-in text driver.

use crate::identity::scrub_git_env;
use serde::Serialize;
use std::fs;
use std::io::Write;
use std::path::{Path, PathBuf};
use std::process::{Command, Output, Stdio};

type Result<T, E = String> = std::result::Result<T, E>;

pub struct Request {
    pub pass_head: String,
    pub tip: String,
    pub head: String,
}

#[derive(Debug, Serialize, PartialEq, Eq)]
#[serde(rename_all = "snake_case")]
pub enum Status {
    PureMerge,
    ManualDelta,
    ReplayConflict,
    Refused,
}

#[derive(Debug, Serialize)]
pub enum FileStatus {
    A,
    M,
    D,
    T,
}

#[derive(Debug, Serialize)]
pub struct FileChange {
    pub path: String,
    pub status: FileStatus,
}

#[derive(Debug, Serialize)]
#[serde(rename_all = "snake_case")]
pub enum RefusalCode {
    NotMerge,
    Octopus,
    ParentMismatch,
}

#[derive(Debug, Serialize)]
pub struct Refusal {
    pub code: RefusalCode,
    pub message: String,
}

#[derive(Debug, Serialize)]
pub struct Writes {
    pub source: bool,
    pub temporary_scratch: bool,
}

#[derive(Debug, Serialize)]
pub struct Report {
    pub schema_version: u32,
    pub status: Status,
    pub pass_head: String,
    pub tip: String,
    pub head: String,
    pub actual_parents: Vec<String>,
    pub actual_tree: String,
    pub replayed_tree: Option<String>,
    pub replay_exit: Option<i32>,
    pub conflicted_files: Vec<String>,
    pub manual_delta: Option<bool>,
    pub manual_text_hunks: Option<bool>,
    pub tip_changed_files: Vec<FileChange>,
    pub pass_changed_files: Vec<FileChange>,
    pub manual_changed_files: Vec<FileChange>,
    pub refusal: Option<Refusal>,
    pub writes: Writes,
    pub network: &'static str,
}

fn is_sha(value: &str) -> bool {
    value.len() == 40
        && value
            .bytes()
            .all(|b| b.is_ascii_digit() || (b'a'..=b'f').contains(&b))
}

fn blocked(message: &str) -> String {
    format!("audit delta blocked: {message}")
}

fn malformed(message: &str) -> String {
    format!("audit delta git failed: unexpected {message}")
}

fn null_device() -> &'static str {
    if cfg!(windows) {
        "NUL"
    } else {
        "/dev/null"
    }
}

/// Strip all caller Git variables, including object redirects, config injection, exec
/// paths, alternate stores and tracing. Never mutate the process environment or cwd.
fn command(dir: &Path) -> Command {
    let mut cmd = Command::new("git");
    scrub_git_env(&mut cmd);
    for (key, _) in std::env::vars_os() {
        if key.to_string_lossy().starts_with("GIT_") {
            cmd.env_remove(key);
        }
    }
    cmd.env("LC_ALL", "C")
        .env("GIT_CONFIG_NOSYSTEM", "1")
        .env("GIT_CONFIG_SYSTEM", null_device())
        .env("GIT_CONFIG_GLOBAL", null_device())
        .env("GIT_ATTR_NOSYSTEM", "1")
        .env("GIT_NO_REPLACE_OBJECTS", "1")
        .env("GIT_NO_LAZY_FETCH", "1")
        .env("GIT_ALLOW_PROTOCOL", "")
        .env("GIT_PROTOCOL_FROM_USER", "0")
        .env("GIT_TERMINAL_PROMPT", "0")
        .env("GIT_OPTIONAL_LOCKS", "0")
        .arg("--no-optional-locks")
        .arg("-c")
        .arg("gc.auto=0")
        .arg("-c")
        .arg("maintenance.auto=false")
        .arg("-c")
        .arg(format!("core.attributesFile={}", null_device()))
        .arg("-C")
        .arg(dir)
        .stdin(Stdio::null());
    cmd
}

fn output(dir: &Path, args: &[&str]) -> Result<Output> {
    command(dir)
        .args(args)
        .output()
        .map_err(|e| format!("audit delta git failed: cannot execute Git: {e}"))
}

fn git(dir: &Path, args: &[&str]) -> Result<Vec<u8>> {
    let out = output(dir, args)?;
    if !out.status.success() {
        return Err(format!(
            "audit delta git failed: {}",
            crate::identity::failure_summary(&out.stderr)
        ));
    }
    Ok(out.stdout)
}

fn text(bytes: &[u8]) -> Result<&str> {
    std::str::from_utf8(bytes).map_err(|_| blocked("Git output is not representable as UTF-8"))
}

/// PROJECT may name a subdirectory. Protect the entire actual worktree, not only
/// that argument. Bare sources have no worktree; their common directory equals the
/// passed directory. With `core.worktree` elsewhere the toplevel is not an ancestor
/// of PROJECT at all, so the passed directory is guarded in its own right (#2529 B1).
fn source_boundaries(project: &Path) -> Result<(PathBuf, Option<PathBuf>)> {
    let raw = git(
        project,
        &[
            "rev-parse",
            "--path-format=absolute",
            "--is-bare-repository",
            "--git-common-dir",
        ],
    )?;
    let raw = raw
        .strip_suffix(b"\n")
        .ok_or_else(|| malformed("source boundary output"))?;
    let mut fields = raw.splitn(2, |&b| b == b'\n');
    let bare = match fields.next() {
        Some(b"true") => true,
        Some(b"false") => false,
        _ => return Err(malformed("bare repository indicator")),
    };
    let path = fields
        .next()
        .ok_or_else(|| malformed("common directory output"))?;
    let path = Path::new(text(path)?);
    if !path.is_absolute() {
        return Err(malformed("relative common directory"));
    }
    let common = fs::canonicalize(path)
        .map_err(|e| format!("audit delta I/O failed: cannot resolve common directory: {e}"))?;
    let worktree =
        if bare {
            None
        } else {
            let raw = git(
                project,
                &["rev-parse", "--path-format=absolute", "--show-toplevel"],
            )?;
            let raw = raw
                .strip_suffix(b"\n")
                .ok_or_else(|| malformed("worktree root output"))?;
            let path = Path::new(text(raw)?);
            if !path.is_absolute() {
                return Err(malformed("relative worktree root"));
            }
            Some(fs::canonicalize(path).map_err(|e| {
                format!("audit delta I/O failed: cannot resolve worktree root: {e}")
            })?)
        };
    Ok((common, worktree))
}

/// An alternates file is one literal absolute pathname per line. Refuse paths that could
/// instead be parsed as quoting or another line; no lossy conversion or shell quoting.
fn alternates_path(objects: &Path) -> Result<String> {
    let path = objects
        .to_str()
        .ok_or_else(|| blocked("object-store path is not UTF-8"))?;
    if path.contains(['\n', '\r', '\0']) || path.starts_with('"') {
        return Err(blocked(
            "object-store path cannot be represented in an alternates file",
        ));
    }
    Ok(format!("{path}\n"))
}

fn setup(scratch: &Path, objects: &str) -> Result<PathBuf> {
    let repo = scratch.join("replay.git");
    let template = scratch.join("empty-template");
    fs::create_dir(&template)
        .map_err(|e| format!("audit delta I/O failed: cannot create empty template: {e}"))?;
    let out = command(scratch)
        .args(["init", "--bare", "--quiet", "--object-format=sha1"])
        .arg("--template")
        .arg(&template)
        .arg(&repo)
        .output()
        .map_err(|e| format!("audit delta git failed: cannot initialize scratch: {e}"))?;
    if !out.status.success() {
        return Err(format!(
            "audit delta git failed: {}",
            crate::identity::failure_summary(&out.stderr)
        ));
    }
    fs::write(repo.join("objects/info/alternates"), objects)
        .map_err(|e| format!("audit delta I/O failed: cannot borrow source objects: {e}"))?;
    Ok(repo)
}

/// Batch-check's machine protocol distinguishes absent local objects from genuine Git
/// failures without classifying localized stderr. Close stdin before waiting, and reap
/// the child even if writing its bounded three-line request fails.
fn commit_check(repo: &Path, request: &Request) -> Result<()> {
    let mut child = command(repo)
        .arg("cat-file")
        .arg("--batch-check=%(objectname) %(objecttype)")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .map_err(|e| format!("audit delta git failed: cannot execute commit validation: {e}"))?;
    let write_result = match child.stdin.take() {
        Some(mut stdin) => (|| {
            for oid in [&request.pass_head, &request.tip, &request.head] {
                writeln!(stdin, "{oid}")?;
            }
            std::io::Result::Ok(())
        })(),
        None => Err(std::io::Error::new(
            std::io::ErrorKind::BrokenPipe,
            "commit validation stdin unavailable",
        )),
    };
    let out = child
        .wait_with_output()
        .map_err(|e| format!("audit delta git failed: cannot collect commit validation: {e}"))?;
    if !out.status.success() {
        return Err(format!(
            "audit delta git failed: {}",
            crate::identity::failure_summary(&out.stderr)
        ));
    }
    write_result
        .map_err(|e| format!("audit delta git failed: cannot write commit validation: {e}"))?;
    let raw = out
        .stdout
        .strip_suffix(b"\n")
        .ok_or_else(|| malformed("commit validation termination"))?;
    let mut lines = raw.split(|&b| b == b'\n');
    let mut invalid_input = false;
    for oid in [&request.pass_head, &request.tip, &request.head] {
        let line = lines
            .next()
            .ok_or_else(|| malformed("commit validation line count"))?;
        let kind = line
            .strip_prefix(oid.as_bytes())
            .and_then(|suffix| suffix.strip_prefix(b" "))
            .ok_or_else(|| malformed("commit validation object identity"))?;
        match kind {
            b"commit" => {}
            b"missing" | b"tag" | b"tree" | b"blob" => invalid_input = true,
            _ => return Err(malformed("commit validation object type")),
        }
    }
    if lines.next().is_some() {
        return Err(malformed("commit validation line count"));
    }
    if invalid_input {
        return Err("audit delta refused: each input must name an available local commit object, not a missing object, tag or other object".into());
    }
    Ok(())
}

fn actual_commit(repo: &Path, head: &str) -> Result<(String, Vec<String>)> {
    let raw = git(
        repo,
        &["show", "--no-patch", "--format=%T%x00%P", head, "--"],
    )?;
    let raw = raw
        .strip_suffix(b"\n")
        .ok_or_else(|| malformed("commit metadata"))?;
    let mut fields = raw.split(|&b| b == 0);
    let tree = text(fields.next().ok_or_else(|| malformed("commit tree"))?)?;
    let parents = text(fields.next().ok_or_else(|| malformed("commit parents"))?)?;
    if !is_sha(tree) || fields.next().is_some() {
        return Err(malformed("commit metadata"));
    }
    let parents: Vec<String> = if parents.is_empty() {
        Vec::new()
    } else {
        parents.split(' ').map(str::to_owned).collect()
    };
    if parents.iter().any(|oid| !is_sha(oid)) {
        return Err(malformed("parent object IDs"));
    }
    Ok((tree.to_owned(), parents))
}

fn nul_fields(raw: &[u8]) -> Result<impl Iterator<Item = &[u8]>> {
    let body = raw
        .strip_suffix(b"\0")
        .ok_or_else(|| malformed("non-NUL-terminated file output"))?;
    Ok(body.split(|&b| b == 0))
}

fn changes(repo: &Path, before: &str, after: &str) -> Result<Vec<FileChange>> {
    let raw = git(
        repo,
        &[
            "diff-tree",
            "--no-commit-id",
            "-r",
            "--name-status",
            "-z",
            "--no-renames",
            "--no-ext-diff",
            "--no-textconv",
            before,
            after,
            "--",
        ],
    )?;
    if raw.is_empty() {
        return Ok(Vec::new());
    }
    let mut fields = nul_fields(&raw)?;
    let mut result = Vec::new();
    while let Some(status) = fields.next() {
        let status = match status {
            b"A" => FileStatus::A,
            b"M" => FileStatus::M,
            b"D" => FileStatus::D,
            b"T" => FileStatus::T,
            _ => return Err(malformed("file status")),
        };
        let path = fields
            .next()
            .ok_or_else(|| malformed("file status/path pair"))?;
        if path.is_empty() {
            return Err(malformed("empty file path"));
        }
        result.push(FileChange {
            path: text(path)?.to_owned(),
            status,
        });
    }
    result.sort_unstable_by(|a, b| a.path.cmp(&b.path));
    Ok(result)
}

/// Numeric numstat distinguishes textual edits from binary or mode-only tree changes
/// without allocating patches. Paths may themselves contain tabs and newlines.
fn text_hunks(repo: &Path, before: &str, after: &str) -> Result<bool> {
    let raw = git(
        repo,
        &[
            "diff",
            "--numstat",
            "-z",
            "--no-ext-diff",
            "--no-textconv",
            "--no-renames",
            before,
            after,
            "--",
        ],
    )?;
    if raw.is_empty() {
        return Ok(false);
    }
    let mut changed = false;
    for row in nul_fields(&raw)? {
        let mut fields = row.splitn(3, |&b| b == b'\t');
        let added = fields
            .next()
            .ok_or_else(|| malformed("numstat additions"))?;
        let removed = fields.next().ok_or_else(|| malformed("numstat removals"))?;
        let path = fields.next().ok_or_else(|| malformed("numstat path"))?;
        if path.is_empty() {
            return Err(malformed("empty numstat path"));
        }
        text(path)?;
        if added == b"-" && removed == b"-" {
            continue;
        }
        for count in [added, removed] {
            if count.is_empty() || !count.iter().all(u8::is_ascii_digit) {
                return Err(malformed("numstat count"));
            }
            changed |= count.iter().any(|&b| b != b'0');
        }
    }
    Ok(changed)
}

fn replay(repo: &Path, report: &mut Report) -> Result<()> {
    let out = output(
        repo,
        &[
            "merge-tree",
            "--write-tree",
            "--name-only",
            "-z",
            "--no-messages",
            &report.actual_parents[0],
            &report.actual_parents[1],
        ],
    )?;
    let exit = match out.status.code() {
        Some(code @ (0 | 1)) => code,
        _ => {
            return Err(format!(
                "audit delta git failed: {}",
                crate::identity::failure_summary(&out.stderr)
            ))
        }
    };
    let mut fields = nul_fields(&out.stdout)?;
    let tree = text(fields.next().ok_or_else(|| malformed("replay tree"))?)?;
    if !is_sha(tree) {
        return Err(malformed("replay tree object ID"));
    }
    for path in fields {
        if path.is_empty() {
            return Err(malformed("empty conflict path"));
        }
        report.conflicted_files.push(text(path)?.to_owned());
    }
    report.conflicted_files.sort_unstable();
    report.conflicted_files.dedup();
    if exit == 0 && !report.conflicted_files.is_empty() {
        return Err(malformed("conflict paths in clean replay"));
    }
    report.replayed_tree = Some(tree.to_owned());
    report.replay_exit = Some(exit);
    if exit == 1 {
        report.status = Status::ReplayConflict;
        return Ok(());
    }
    let delta = tree != report.actual_tree;
    report.manual_text_hunks = Some(false);
    if delta {
        report.manual_changed_files = changes(repo, tree, &report.actual_tree)?;
        report.manual_text_hunks = Some(text_hunks(repo, tree, &report.actual_tree)?);
    }
    report.manual_delta = Some(delta);
    report.status = if delta {
        Status::ManualDelta
    } else {
        Status::PureMerge
    };
    Ok(())
}

fn inspect_scratch(repo: &Path, request: Request) -> Result<Report> {
    commit_check(repo, &request)?;
    let (actual_tree, actual_parents) = actual_commit(repo, &request.head)?;
    let tip_changed_files = changes(repo, &request.tip, &request.head)?;
    let pass_changed_files = changes(repo, &request.pass_head, &request.head)?;
    let refusal = match actual_parents.len() {
        0 | 1 => Some(Refusal {
            code: RefusalCode::NotMerge,
            message: "head is not a merge commit".into(),
        }),
        2 if actual_parents[0] != request.pass_head || actual_parents[1] != request.tip => {
            Some(Refusal {
                code: RefusalCode::ParentMismatch,
                message: "head parents are not exactly [pass_head, tip] in that order".into(),
            })
        }
        2 => None,
        _ => Some(Refusal {
            code: RefusalCode::Octopus,
            message: "head has more than two parents".into(),
        }),
    };
    let mut report = Report {
        schema_version: 1,
        status: Status::Refused,
        pass_head: request.pass_head,
        tip: request.tip,
        head: request.head,
        actual_parents,
        actual_tree,
        replayed_tree: None,
        replay_exit: None,
        conflicted_files: Vec::new(),
        manual_delta: None,
        manual_text_hunks: None,
        tip_changed_files,
        pass_changed_files,
        manual_changed_files: Vec::new(),
        refusal,
        writes: Writes {
            source: false,
            temporary_scratch: true,
        },
        network: "none",
    };
    if report.refusal.is_none() {
        replay(repo, &mut report)?;
    }
    Ok(report)
}

/// Inspect exact local commits without checking out, updating or writing source objects.
/// Cleanup is explicit even on errors: a failed removal is an I/O error, never success.
pub fn inspect(project: &Path, request: Request) -> Result<Report> {
    for (name, oid) in [
        ("pass-head", &request.pass_head),
        ("tip", &request.tip),
        ("head", &request.head),
    ] {
        if !is_sha(oid) {
            return Err(format!(
                "audit delta refused: --{name} must be a 40-hex lowercase commit"
            ));
        }
    }
    let project = fs::canonicalize(project)
        .map_err(|e| format!("audit delta I/O failed: cannot resolve project: {e}"))?;
    let (common, worktree) = source_boundaries(&project)?;
    let objects_dir = fs::canonicalize(common.join("objects"))
        .map_err(|e| format!("audit delta I/O failed: cannot resolve object directory: {e}"))?;
    let objects = alternates_path(&objects_dir)?;
    let temp_root = fs::canonicalize(std::env::temp_dir())
        .map_err(|e| format!("audit delta I/O failed: cannot resolve temporary root: {e}"))?;
    // Guard the passed directory itself as well: with `core.worktree` pointing at
    // another folder, --show-toplevel is that folder (not an ancestor of PROJECT)
    // and --git-common-dir is PROJECT/.git, so a temporary root under PROJECT
    // escapes the worktree and common clauses (#2529 B1). Pinned by the
    // subdirectory, bare-source and core.worktree CLI tests.
    if temp_root.starts_with(&project)
        || temp_root.starts_with(&common)
        || temp_root.starts_with(&objects_dir)
        || worktree
            .as_ref()
            .is_some_and(|root| temp_root.starts_with(root))
    {
        return Err(blocked(
            "temporary root must be outside the source worktree, project and common Git directory",
        ));
    }
    let scratch = tempfile::Builder::new()
        .prefix("moai-audit-delta-")
        .tempdir_in(&temp_root)
        .map_err(|e| format!("audit delta I/O failed: cannot create scratch: {e}"))?;
    let result = setup(scratch.path(), &objects).and_then(|repo| inspect_scratch(&repo, request));
    scratch
        .close()
        .map_err(|e| format!("audit delta I/O failed: cannot clean owned scratch: {e}"))?;
    result
}
