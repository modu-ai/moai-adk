//! Read-only baseline inventory (Issue #1985, 00-A). Records the checkout state, a sha256
//! inventory of tracked and changed files, and a byte-level `.claude` comparison between the
//! Go baseline commit, its install template, and the current checkout's installed copy.
//!
//! The collector never writes: git is invoked with `--no-optional-locks` (no index refresh),
//! the Go baseline is read with `cat-file`/`ls-tree` only, blob ids are computed with
//! `hash-object` without `-w`, and the user home is checked for existence only (REQ-00A-08).
//! Output paths are repository-relative; absolute paths never reach the result (REQ-00A-09).
//! Classification is mechanical; meaning-level judgments stay with the design pages.
use serde_json::{json, Value};
use std::{
    collections::{BTreeMap, BTreeSet},
    fs,
    io::Write,
    path::{Path, PathBuf},
    process::{Command, Stdio},
    time::{SystemTime, UNIX_EPOCH},
};

type Result<T> = std::result::Result<T, String>;

/// Lane worktrees nest inside the primary checkout; they are not part of its state.
const EXCLUDED_PREFIXES: [&str; 3] = [".claude/worktrees/", ".moai-cli/worktrees/", ".worktrees/"];
const HOME_ENTRIES: [&str; 3] = [".claude", ".moai", ".agents"];
const TEMPLATE_PREFIX: &str = "internal/template/templates/";

fn git(dir: &Path, args: &[&str], stdin: Option<&[u8]>) -> Result<(i32, Vec<u8>)> {
    run_git(dir, args, stdin).map(|(code, out, _)| (code, out))
}
/// Runs git with a fixed argv. The caller's `GIT_DIR`, `GIT_WORK_TREE`, injected config and
/// discovery variables are removed and the locale is fixed (`identity::scrub_git_env`), so
/// PROJECT is the repository read (Issue #2054). Returns exit code, stdout and stderr; stderr
/// is only used to classify a failure.
fn run_git(dir: &Path, args: &[&str], stdin: Option<&[u8]>) -> Result<(i32, Vec<u8>, Vec<u8>)> {
    let mut cmd = Command::new("git");
    cmd.arg("--no-optional-locks").arg("-C").arg(dir).args(args);
    crate::identity::scrub_git_env(&mut cmd);
    let mut child = cmd
        .env("GIT_OPTIONAL_LOCKS", "0")
        .env("GIT_TERMINAL_PROMPT", "0")
        .stdin(if stdin.is_some() {
            Stdio::piped()
        } else {
            Stdio::null()
        })
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .map_err(|_| "git unavailable")?;
    if let Some(bytes) = stdin {
        child
            .stdin
            .take()
            .ok_or("git stdin unavailable")?
            .write_all(bytes)
            .map_err(|_| "git stdin write failed")?;
    }
    let out = child.wait_with_output().map_err(|_| "git failed")?;
    Ok((out.status.code().unwrap_or(-1), out.stdout, out.stderr))
}
/// A failure other than "not a git repository" is `git failed: <path-free summary>`
/// (`git_failed`, as in #2033); "not a git repository" keeps `git <cmd> exited <code>`.
fn git_ok(dir: &Path, args: &[&str]) -> Result<Vec<u8>> {
    match run_git(dir, args, None)? {
        (0, out, _) => Ok(out),
        (code, _, err) => {
            let summary = crate::identity::failure_summary(&err);
            if err.is_empty() || summary.starts_with(crate::identity::NOT_A_REPOSITORY) {
                Err(format!("git {} exited {code}", args[0]))
            } else {
                Err(format!("git failed: {summary}"))
            }
        }
    }
}
fn text(bytes: Vec<u8>) -> Result<String> {
    String::from_utf8(bytes).map_err(|_| "non UTF-8 git output".into())
}
fn fields(bytes: &[u8]) -> Result<Vec<String>> {
    bytes
        .split(|b| *b == 0)
        .filter(|f| !f.is_empty())
        .map(|f| String::from_utf8(f.to_vec()).map_err(|_| "non UTF-8 path".into()))
        .collect()
}
fn excluded(path: &str) -> bool {
    excluded_by(path, &EXCLUDED_PREFIXES)
}
/// `rel` is excluded when it equals an entry (trailing `/` ignored) or lies under it.
fn excluded_by(rel: &str, excludes: &[&str]) -> bool {
    excludes.iter().any(|e| {
        let e = e.trim_end_matches('/');
        !e.is_empty() && (rel == e || rel.starts_with(&format!("{e}/")))
    })
}
fn hex40(s: &str) -> bool {
    s.len() == 40
        && s.bytes()
            .all(|b| b.is_ascii_digit() || (b'a'..=b'f').contains(&b))
}

/// Working-tree bytes of a repository-relative path, without following symlinks.
/// Returns (kind, sha256 of file bytes or link target).
fn content(root: &Path, rel: &str) -> Result<(&'static str, Option<String>)> {
    let path = root.join(rel);
    let Ok(m) = fs::symlink_metadata(&path) else {
        return Ok(("missing", None));
    };
    if m.file_type().is_symlink() {
        let target = fs::read_link(&path).map_err(|_| "symlink unreadable")?;
        let raw = target.to_str().ok_or("non UTF-8 symlink target")?;
        return Ok(("symlink", Some(crate::digest(raw.as_bytes()))));
    }
    if m.is_file() {
        let bytes = fs::read(&path).map_err(|_| format!("unreadable file {rel}"))?;
        return Ok(("file", Some(crate::digest(&bytes))));
    }
    Ok((if m.is_dir() { "directory" } else { "special" }, None))
}

/// RFC 3339 UTC seconds from the system clock (civil-from-days, no external crate).
pub(crate) fn utc_now() -> String {
    utc_from_epoch(
        SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .map(|d| d.as_secs() as i64)
            .unwrap_or(0),
    )
}

/// RFC 3339 UTC seconds for a Unix timestamp (civil-from-days, no external crate).
pub(crate) fn utc_from_epoch(secs: i64) -> String {
    let (days, rem) = (secs.div_euclid(86400), secs.rem_euclid(86400));
    let z = days + 719_468;
    let era = z.div_euclid(146_097);
    let doe = z - era * 146_097;
    let yoe = (doe - doe / 1460 + doe / 36524 - doe / 146_096) / 365;
    let doy = doe - (365 * yoe + yoe / 4 - yoe / 100);
    let mp = (5 * doy + 2) / 153;
    let d = doy - (153 * mp + 2) / 5 + 1;
    let m = if mp < 10 { mp + 3 } else { mp - 9 };
    let y = yoe + era * 400 + i64::from(m <= 2);
    format!(
        "{y:04}-{m:02}-{d:02}T{:02}:{:02}:{:02}Z",
        rem / 3600,
        rem % 3600 / 60,
        rem % 60
    )
}

fn relativize(root: &Path, raw: &str) -> String {
    let p = fs::canonicalize(raw).unwrap_or_else(|_| PathBuf::from(raw));
    match p.strip_prefix(root) {
        Ok(r) if r.as_os_str().is_empty() => ".".into(),
        Ok(r) => r.to_string_lossy().into_owned(),
        Err(_) => "<outside-project>".into(),
    }
}

fn worktrees(root: &Path) -> Result<Vec<Value>> {
    let out = git_ok(root, &["worktree", "list", "--porcelain", "-z"])?;
    let mut rows = Vec::new();
    let mut cur: BTreeMap<String, Value> = BTreeMap::new();
    for field in out.split(|b| *b == 0) {
        if field.is_empty() {
            if !cur.is_empty() {
                rows.push(json!(std::mem::take(&mut cur)));
            }
            continue;
        }
        let f = String::from_utf8_lossy(field);
        let (key, val) = f.split_once(' ').unwrap_or((&f, ""));
        match key {
            "worktree" => cur.insert("path".into(), json!(relativize(root, val))),
            "HEAD" => cur.insert("head".into(), json!(val)),
            "branch" => cur.insert("branch".into(), json!(val)),
            // Lock and prune reasons are free text that may carry paths; keep the flag only.
            "detached" | "bare" | "locked" | "prunable" => cur.insert(key.into(), json!(true)),
            _ => None,
        };
    }
    if !cur.is_empty() {
        rows.push(json!(cur));
    }
    rows.sort_by_key(|r| r["path"].as_str().unwrap_or("").to_string());
    Ok(rows)
}

fn remote_heads(root: &Path) -> Value {
    match git(root, &["ls-remote", "--heads", "origin"], None) {
        Ok((0, out)) => {
            let mut heads: Vec<Value> = String::from_utf8_lossy(&out)
                .lines()
                .filter_map(|l| l.split_once('\t'))
                .map(|(sha, r)| json!({"ref": r, "sha": sha}))
                .collect();
            heads.sort_by_key(|h| h["ref"].as_str().unwrap_or("").to_string());
            json!({"status": "collected", "heads": heads})
        }
        // stderr is discarded: it can echo a remote URL carrying credentials.
        Ok((code, _)) => {
            json!({"status": "unavailable", "reason": format!("ls-remote exited {code}")})
        }
        Err(e) => json!({"status": "unavailable", "reason": e}),
    }
}

/// Parses `git status --porcelain=v2 -z --untracked-files=all`. Header lines (`#`) and
/// unknown record types are skipped: only v1 carries a documented stability guarantee.
fn changes(root: &Path, excluded_count: &mut usize) -> Result<Vec<Value>> {
    let out = git_ok(
        root,
        &["status", "--porcelain=v2", "-z", "--untracked-files=all"],
    )?;
    let recs = fields(&out)?;
    let mut rows = Vec::new();
    let mut i = 0;
    while i < recs.len() {
        let r = &recs[i];
        i += 1;
        let parts: Vec<&str> = r.splitn(11, ' ').collect();
        let (kind, xy, head, index, path, orig) = match parts[0] {
            "1" if parts.len() >= 9 => {
                let rest: Vec<&str> = r.splitn(9, ' ').collect();
                (
                    "ordinary",
                    rest[1],
                    rest[6],
                    rest[7],
                    rest[8].to_string(),
                    None,
                )
            }
            "2" if parts.len() >= 10 => {
                let rest: Vec<&str> = r.splitn(10, ' ').collect();
                let orig = recs.get(i).cloned();
                i += 1;
                (
                    "renamed",
                    rest[1],
                    rest[6],
                    rest[7],
                    rest[9].to_string(),
                    orig,
                )
            }
            "u" if parts.len() >= 11 => ("unmerged", parts[1], "", "", parts[10].to_string(), None),
            "?" => ("untracked", "??", "", "", r[2..].to_string(), None),
            _ => continue,
        };
        if excluded(&path) {
            *excluded_count += 1;
            continue;
        }
        let (file_kind, sha) = content(root, &path)?;
        let opt = |s: &str| {
            if hex40(s) && s != "0".repeat(40) {
                json!(s)
            } else {
                Value::Null
            }
        };
        rows.push(json!({
            "path": path, "status": kind, "xy": xy, "kind": file_kind, "sha256": sha,
            "object_head": opt(head), "object_index": opt(index), "orig_path": orig,
        }));
    }
    rows.sort_by_key(|r| r["path"].as_str().unwrap_or("").to_string());
    Ok(rows)
}

fn tracked(root: &Path) -> Result<Vec<Value>> {
    let out = git_ok(root, &["ls-files", "-s", "-z"])?;
    let mut seen = BTreeMap::new();
    for rec in fields(&out)? {
        let (meta, path) = rec.split_once('\t').ok_or("malformed ls-files record")?;
        let m: Vec<&str> = meta.split(' ').collect();
        if m.len() != 3 || seen.contains_key(path) {
            continue;
        }
        let (kind, sha) = if m[0] == "160000" {
            ("gitlink", None)
        } else {
            content(root, path)?
        };
        seen.insert(
            path.to_string(),
            json!({"path": path, "mode": m[0], "object": m[1], "kind": kind, "sha256": sha}),
        );
    }
    Ok(seen.into_values().collect())
}

/// Blob ids of active files, computed without writing objects.
fn blob_ids(root: &Path, files: &[(String, &'static str)]) -> Result<BTreeMap<String, String>> {
    let mut ids = BTreeMap::new();
    let regular: Vec<&String> = files
        .iter()
        .filter(|f| f.1 == "file")
        .map(|f| &f.0)
        .collect();
    if !regular.is_empty() {
        let list: String = regular.iter().map(|p| format!("{p}\n")).collect();
        let (code, out) = git(
            root,
            &["hash-object", "--no-filters", "--stdin-paths"],
            Some(list.as_bytes()),
        )?;
        let out = text(out)?;
        let lines: Vec<&str> = out.lines().collect();
        if code != 0 || lines.len() != regular.len() {
            return Err("hash-object failed".into());
        }
        for (p, id) in regular.iter().zip(lines) {
            ids.insert((*p).clone(), id.to_string());
        }
    }
    for (p, _) in files.iter().filter(|f| f.1 == "symlink") {
        let target = fs::read_link(root.join(p)).map_err(|_| "symlink unreadable")?;
        let raw = target.to_str().ok_or("non UTF-8 symlink target")?;
        let (code, out) = git(
            root,
            &["hash-object", "--no-filters", "--stdin"],
            Some(raw.as_bytes()),
        )?;
        if code != 0 {
            return Err("hash-object failed".into());
        }
        ids.insert(p.clone(), text(out)?.trim().to_string());
    }
    Ok(ids)
}

/// Lists regular files and symlinks under `dir` (paths relative to `root`, symlinks never
/// followed). Entries matching `excludes` are skipped and their files counted in `skipped`;
/// special files are refused.
fn walk(
    root: &Path,
    dir: &Path,
    excludes: &[&str],
    out: &mut Vec<(String, &'static str)>,
    skipped: &mut usize,
) -> Result<()> {
    let Ok(entries) = fs::read_dir(dir) else {
        return Ok(());
    };
    for entry in entries {
        let path = entry.map_err(|_| "directory entry unavailable")?.path();
        let rel = path
            .strip_prefix(root)
            .map_err(|_| "path outside root")?
            .to_str()
            .ok_or("non UTF-8 path")?
            .to_string();
        let m = fs::symlink_metadata(&path).map_err(|_| "metadata unavailable")?;
        if excluded_by(&rel, excludes) {
            *skipped += if m.is_dir() { count_files(&path) } else { 1 };
        } else if m.is_dir() {
            walk(root, &path, excludes, out, skipped)?;
        } else if m.file_type().is_symlink() {
            out.push((rel, "symlink"));
        } else if m.is_file() {
            out.push((rel, "file"));
        } else {
            return Err(format!("special file refused: {rel}"));
        }
    }
    Ok(())
}

// @MX:ANCHOR tree digest owned by 00-A (#1985); 05-C #2003 and 07-D #2011 call it, never re-implement.
/// Content digest of the file tree under `root` (format `moai-tree-v1`). Paths are relative,
/// `/`-separated, UTF-8 and sorted bytewise; each entry is framed as kind byte (`f` file,
/// `l` symlink), u64 big-endian path length, path, u64 big-endian content length, content
/// (file bytes, or the link target for a symlink, which is never followed). Empty directories
/// do not contribute. `excludes` are root-relative paths; an entry equal to or under one is
/// left out. Special files are refused. Returns `sha256:<64 lowercase hex>`.
pub fn tree_digest(root: &Path, excludes: &[&str]) -> Result<String> {
    use sha2::{Digest, Sha256};
    let mut files = Vec::new();
    walk(root, root, excludes, &mut files, &mut 0)?;
    files.sort();
    let mut h = Sha256::new();
    h.update(b"moai-tree-v1\0");
    for (rel, kind) in files {
        let path = root.join(&rel);
        let (tag, bytes) = if kind == "symlink" {
            let t = fs::read_link(&path).map_err(|_| "symlink unreadable")?;
            (
                b'l',
                t.to_str()
                    .ok_or("non UTF-8 symlink target")?
                    .as_bytes()
                    .to_vec(),
            )
        } else {
            (
                b'f',
                fs::read(&path).map_err(|_| format!("unreadable file {rel}"))?,
            )
        };
        let name = rel.replace(std::path::MAIN_SEPARATOR, "/");
        h.update([tag]);
        h.update((name.len() as u64).to_be_bytes());
        h.update(name.as_bytes());
        h.update((bytes.len() as u64).to_be_bytes());
        h.update(&bytes);
    }
    Ok(format!("sha256:{:x}", h.finalize()))
}
fn count_files(dir: &Path) -> usize {
    fs::read_dir(dir)
        .map(|es| {
            es.filter_map(|e| e.ok())
                .map(|e| match fs::symlink_metadata(e.path()) {
                    Ok(m) if m.is_dir() => count_files(&e.path()),
                    Ok(_) => 1,
                    Err(_) => 0,
                })
                .sum()
        })
        .unwrap_or(0)
}

fn scope(path: &str) -> &'static str {
    let name = path.rsplit('/').next().unwrap_or(path);
    if name.contains(".local.") || name.ends_with(".local") {
        "private"
    } else {
        "shared"
    }
}

fn claude_diff(
    root: &Path,
    commit: Option<&str>,
    go_repo: Option<&Path>,
    tracked_paths: &BTreeSet<String>,
) -> Result<Value> {
    let rules = json!({
        "key": "path relative to the .claude directory of each set",
        "class": "baseline-only|template-only|active-only when present in one set; identical when present in two or more sets with one blob id; differing otherwise",
        "tracking": "current checkout: tracked (ls-files), ignored (check-ignore), untracked, or absent",
        "origin": "generated when the template set holds the path, else source",
        "scope": "private when the file name contains .local. or ends with .local, else shared",
        "excluded": EXCLUDED_PREFIXES,
    });
    let Some(commit) = commit else {
        return Ok(
            json!({"status": "not_collected", "reason": "no --go-baseline-commit", "rules": rules}),
        );
    };
    if !hex40(commit) {
        return Err("go baseline commit must be 40 lowercase hex".into());
    }
    let spec = format!("{commit}^{{commit}}");
    let has = |dir: &Path| matches!(git(dir, &["cat-file", "-e", &spec], None), Ok((0, _)));
    let (repo, source): (PathBuf, &str) = if has(root) {
        (root.to_path_buf(), "project")
    } else if let Some(g) = go_repo.filter(|g| has(g)) {
        (g.to_path_buf(), "go-repo")
    } else {
        return Ok(
            json!({"status": "commit_not_found", "go_baseline_commit": commit, "rules": rules}),
        );
    };
    let tree = git_ok(
        &repo,
        &[
            "ls-tree",
            "-r",
            "-z",
            commit,
            "--",
            ".claude",
            "internal/template/templates",
        ],
    )?;
    let mut baseline = BTreeMap::new();
    let mut template = BTreeMap::new();
    let mut roots = BTreeSet::new();
    for rec in fields(&tree)? {
        let (meta, path) = rec.split_once('\t').ok_or("malformed ls-tree record")?;
        let object = meta
            .split(' ')
            .nth(2)
            .ok_or("malformed ls-tree record")?
            .to_string();
        if let Some(rel) = path.strip_prefix(".claude/") {
            baseline.insert(rel.to_string(), object);
        } else if let Some(rest) = path.strip_prefix(TEMPLATE_PREFIX) {
            let (troot, rel) = if let Some(rel) = rest.strip_prefix(".claude/") {
                ("", rel)
            } else if let Some(at) = rest.find("/.claude/") {
                (&rest[..at], &rest[at + 9..])
            } else {
                continue;
            };
            roots.insert(troot.to_string());
            template.insert(rel.to_string(), object);
        }
    }
    if roots.len() > 1 {
        return Ok(json!({"status": "multiple_template_roots", "roots": roots, "rules": rules}));
    }
    let mut active_files = Vec::new();
    let mut skipped = 0;
    walk(
        root,
        &root.join(".claude"),
        &EXCLUDED_PREFIXES,
        &mut active_files,
        &mut skipped,
    )?;
    let active_tree_digest = tree_digest(&root.join(".claude"), &["worktrees/"])?;
    let ids = blob_ids(root, &active_files)?;
    let mut active = BTreeMap::new();
    for (p, kind) in &active_files {
        let rel = p.strip_prefix(".claude/").unwrap_or(p).to_string();
        let (_, sha) = content(root, p)?;
        active.insert(rel, (ids.get(p).cloned(), sha, *kind));
    }
    let active_list: String = active.keys().map(|k| format!(".claude/{k}\0")).collect();
    let ignored: BTreeSet<String> = match git(
        root,
        &["check-ignore", "-z", "--stdin"],
        Some(active_list.as_bytes()),
    )? {
        (0 | 1, out) => fields(&out)?.into_iter().collect(),
        (code, _) => return Err(format!("git check-ignore exited {code}")),
    };
    let keys: BTreeSet<&String> = baseline
        .keys()
        .chain(template.keys())
        .chain(active.keys())
        .collect();
    let mut counts: BTreeMap<&str, usize> = BTreeMap::new();
    let mut rows = Vec::new();
    for k in keys {
        let b = baseline.get(k);
        let t = template.get(k);
        let a = active.get(k);
        let ids: Vec<Option<&String>> = vec![b, t, a.map(|a| a.0.as_ref()).unwrap_or(None)];
        let present = [b.is_some(), t.is_some(), a.is_some()];
        let n = present.iter().filter(|p| **p).count();
        let class = match (n, present) {
            (1, [true, _, _]) => "baseline-only",
            (1, [_, true, _]) => "template-only",
            (1, _) => "active-only",
            _ => {
                let distinct: BTreeSet<Option<&String>> = ids
                    .iter()
                    .zip(present)
                    .filter(|(_, p)| *p)
                    .map(|(i, _)| *i)
                    .collect();
                if distinct.len() == 1 && !distinct.contains(&None) {
                    "identical"
                } else {
                    "differing"
                }
            }
        };
        *counts.entry(class).or_default() += 1;
        let full = format!(".claude/{k}");
        let tracking = if a.is_none() {
            "absent"
        } else if tracked_paths.contains(&full) {
            "tracked"
        } else if ignored.contains(&full) {
            "ignored"
        } else {
            "untracked"
        };
        rows.push(json!({
            "path": k, "class": class,
            "baseline": b, "template": t,
            "active": a.map(|(id, sha, kind)| json!({"object": id, "sha256": sha, "kind": kind})),
            "tracking": tracking,
            "origin": if t.is_some() { "generated" } else { "source" },
            "scope": scope(k),
        }));
    }
    let total = rows.len();
    Ok(json!({
        "status": "collected", "go_baseline_commit": commit, "source": source,
        "template_root": roots.iter().next().map(|r| format!("{TEMPLATE_PREFIX}{r}")),
        "rules": rules, "counts": counts, "total": total,
        "set_sizes": {"baseline": baseline.len(), "template": template.len(), "active": active.len()},
        "excluded_active_files": skipped,
        "active_tree_digest": active_tree_digest,
        "active_tree_digest_excludes": ["worktrees/"],
        "rows": rows,
    }))
}

/// Collects the read-only baseline for `project`. Returns
/// `{schema_version, collected_at, baseline, inventory, claude_diff, checksums}`.
pub fn inventory(project: &Path, go_commit: Option<&str>, go_repo: Option<&Path>) -> Result<Value> {
    let collected_at = utc_now();
    let top = text(git_ok(project, &["rev-parse", "--show-toplevel"])?)?;
    let root = fs::canonicalize(top.trim()).map_err(|_| "project root unavailable")?;
    let head = text(git_ok(&root, &["rev-parse", "HEAD"])?)?
        .trim()
        .to_string();
    if !hex40(&head) {
        return Err("HEAD is not a 40 hex commit".into());
    }
    let branch = text(git_ok(&root, &["branch", "--show-current"])?)?
        .trim()
        .to_string();
    let home = std::env::var_os("HOME").map(PathBuf::from);
    let home_rows: Vec<Value> = HOME_ENTRIES
        .iter()
        .map(|e| {
            let exists = home
                .as_ref()
                .map(|h| fs::symlink_metadata(h.join(e)).is_ok());
            json!({"path": format!("~/{e}"), "exists": exists})
        })
        .collect();
    let files = tracked(&root)?;
    let tracked_paths: BTreeSet<String> = files
        .iter()
        .filter_map(|f| f["path"].as_str().map(String::from))
        .collect();
    let mut excluded_changes = 0;
    let changed = changes(&root, &mut excluded_changes)?;
    let mut sums = BTreeMap::new();
    // shasum-format lines cannot carry these paths unescaped; they stay in `files`/`changes`
    // and are listed here so the omission is visible rather than silent.
    let mut unrepresentable = BTreeSet::new();
    for f in files.iter().chain(changed.iter()) {
        if let (Some(p), Some(s), Some("file")) =
            (f["path"].as_str(), f["sha256"].as_str(), f["kind"].as_str())
        {
            if p.contains('\n') || p.contains('\\') {
                unrepresentable.insert(p.to_string());
            } else {
                sums.insert(p.to_string(), s[7..].to_string());
            }
        }
    }
    let checksums: String = sums.iter().map(|(p, h)| format!("{h}  {p}\n")).collect();
    let mut diff = claude_diff(&root, go_commit, go_repo, &tracked_paths)?;
    diff["schema_version"] = json!(1);
    diff["collected_at"] = json!(collected_at);
    diff["head_sha"] = json!(head);
    let count = |s: &str| changed.iter().filter(|c| c["status"] == s).count();
    let measurements = json!([
        {"name": "tracked_files", "command": "git ls-files -s -z", "value": files.len()},
        {"name": "changed_tracked", "command": "git status --porcelain=v2 -z --untracked-files=all", "value": changed.len() - count("untracked")},
        {"name": "untracked_files", "command": "git status --porcelain=v2 -z --untracked-files=all", "value": count("untracked")},
        {"name": "checksum_rows", "command": "sha256 of regular files among tracked and changed rows", "value": sums.len()},
    ]);
    Ok(json!({
        "schema_version": 1,
        "collected_at": collected_at,
        "baseline": {
            "schema_version": 1, "collected_at": collected_at,
            "head_sha": head, "branch": if branch.is_empty() { Value::Null } else { json!(branch) },
            "worktrees": worktrees(&root)?,
            "remote_heads": remote_heads(&root),
            "home": home_rows,
            "path_rule": "paths are relative to the project root; worktrees outside it read <outside-project>; home entries are ~-relative and recorded as existence only",
        },
        "inventory": {
            "schema_version": 1, "collected_at": collected_at, "head_sha": head,
            "files": files, "changes": changed,
            "excluded_prefixes": EXCLUDED_PREFIXES, "excluded_changes": excluded_changes,
            "checksum_unrepresentable": unrepresentable,
            "measurements": measurements,
            "measured_at": "collected_at",
        },
        "claude_diff": diff,
        "checksums": checksums,
    }))
}
