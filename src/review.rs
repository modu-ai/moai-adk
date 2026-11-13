//! `review-package` (#2614): the deterministic review package an AUDIT hands to a reviewer.
//!
//! It holds the diff from the merge base to the head, the changed files, the review
//! instructions (`REVIEW.md`, its `@` includes, and the subdirectory `REVIEW.md` files on the
//! changed paths) and the project context (`AGENTS.md`, `CLAUDE.md` at the root and above each
//! changed file). Instructions and context are read from the BASE commit only, never from the
//! change under review, following Claude Code's Code Review documentation
//! (code.claude.com/docs/en/code-review-customize): one level of `@name` includes beside the
//! root file, at most eight different files, root plus includes under 150,000 bytes combined
//! (a file that would pass the limit is skipped whole), more than 64 Markdown links skip the
//! root file, subdirectory files share their own 150,000 bytes, and directories with test,
//! fixture, vendored or generated names are skipped. Branch protection cannot be checked
//! offline, so includes always apply and the package says `protection: not_checked`
//! (design 6839486269, decision D1). Skill includes are not supported yet (D2).
//!
//! Git runs with the caller's configuration and environment removed
//! (`audit_delta::command`) and with fixed diff options, so the same base and head give the
//! same bytes on every machine and in every tool (AC-2). The repository's own diff settings are
//! overridden on the command line, attributes come from an empty tree instead of the worktree,
//! and a repository with `.git/info/attributes` is refused (audit FAIL 6839909483 B1). Paths and
//! diff text that are not UTF-8 are carried with U+FFFD replacements (deterministic, and
//! `diff_sha256` hashes that replaced text). Reads only; writes nothing.

use serde::Serialize;
use std::collections::BTreeSet;
use std::path::Path;

type Result<T> = std::result::Result<T, String>;

pub const REVIEW_FILE: &str = "REVIEW.md";
pub const BYTE_LIMIT: usize = 150_000;
pub const INCLUDE_LIMIT: usize = 8;
pub const LINK_LIMIT: usize = 64;
const CONTEXT_FILES: [&str; 2] = ["AGENTS.md", "CLAUDE.md"];
/// Directory names whose subdirectory `REVIEW.md` is not read, matched whole and in any case
/// at any depth ([customize]).
const SKIPPED_DIRS: [&str; 31] = [
    "test",
    "tests",
    "__tests__",
    "testdata",
    "fixtures",
    "__fixtures__",
    "golden",
    "corpus",
    "snapshots",
    "__snapshots__",
    "example",
    "examples",
    "sample",
    "samples",
    "demo",
    "demos",
    "generated",
    "vendor",
    "third_party",
    "third-party",
    "thirdparty",
    "node_modules",
    "bower_components",
    "external",
    "extern",
    "deps",
    "godeps",
    "pods",
    "carthage",
    "site-packages",
    "contrib",
];

#[derive(Debug, Clone, Serialize, PartialEq, Eq)]
pub struct Changed {
    pub path: String,
    pub status: String,
}

#[derive(Debug, Clone, Serialize, PartialEq, Eq)]
pub struct File {
    pub path: String,
    /// `root`, `include` or `subdirectory` for review files; the file name for context.
    pub role: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub included_by: Option<String>,
    pub bytes: usize,
    pub sha256: String,
    pub text: String,
}

#[derive(Debug, Clone, Serialize, PartialEq, Eq)]
pub struct Skipped {
    pub path: String,
    pub reason: String,
}

#[derive(Debug, Clone, Serialize, PartialEq, Eq)]
pub struct Package {
    pub schema_version: u32,
    pub base: String,
    pub head: String,
    pub merge_base: String,
    pub protection: String,
    pub changed: Vec<Changed>,
    pub diff: String,
    pub diff_sha256: String,
    pub review: Vec<File>,
    pub context: Vec<File>,
    pub skipped: Vec<Skipped>,
    pub package_digest: String,
}

fn git(repo: &Path, args: &[&str]) -> Result<Vec<u8>> {
    let out = crate::audit_delta::command(repo)
        .args(args)
        .output()
        .map_err(|e| format!("review package git failed: cannot execute Git: {e}"))?;
    if !out.status.success() {
        return Err(format!(
            "review package git failed: {}",
            crate::identity::failure_summary(&out.stderr)
        ));
    }
    Ok(out.stdout)
}

fn null_device() -> &'static str {
    if cfg!(windows) {
        "NUL"
    } else {
        "/dev/null"
    }
}

/// `git --attr-source=<empty tree> -c … diff`: the configuration keys a repository could set
/// and the command line has no option for.
fn diff_prefix(attr_source: &str) -> Vec<&str> {
    vec![
        attr_source,
        "-c",
        "core.quotePath=true",
        "-c",
        "core.bigFileThreshold=512m",
        "-c",
        "diff.suppressBlankEmpty=false",
        "-c",
        "diff.noprefix=false",
        "-c",
        "diff.mnemonicPrefix=false",
        "diff",
    ]
}

/// The repository's top level. Every path in the package is relative to it, and git runs
/// there, so PROJECT given as a subfolder reads the same root `REVIEW.md` and the same diff
/// (follow-up 6840613078).
fn toplevel(repo: &Path) -> Result<std::path::PathBuf> {
    let out = git(repo, &["rev-parse", "--show-toplevel"])?;
    let top = String::from_utf8_lossy(&out).trim().to_string();
    if top.is_empty() {
        return Err("not inside a git work tree".into());
    }
    Ok(std::path::PathBuf::from(top))
}

fn commit_of(repo: &Path, rev: &str) -> Result<String> {
    if rev.is_empty() || rev.starts_with('-') {
        return Err(format!("invalid revision {rev:?}"));
    }
    let out = git(
        repo,
        &[
            "rev-parse",
            "--verify",
            "--end-of-options",
            &format!("{rev}^{{commit}}"),
        ],
    )
    .map_err(|_| format!("unknown revision {rev:?}"))?;
    Ok(String::from_utf8_lossy(&out).trim().to_string())
}

/// What a path is in `commit`: `Some(Ok(text))` for a regular UTF-8 file, `Some(Err(reason))`
/// for anything else that exists there, `None` when it is absent.
fn blob_text(
    repo: &Path,
    commit: &str,
    path: &str,
) -> Result<Option<std::result::Result<String, &'static str>>> {
    let listing = git(repo, &["ls-tree", "-z", commit, "--", path])?;
    let Some(entry) = listing.split(|b| *b == 0).find(|e| !e.is_empty()) else {
        return Ok(None);
    };
    let entry = String::from_utf8_lossy(entry);
    let (meta, name) = entry.split_once('\t').ok_or("unreadable tree entry")?;
    // `ls-tree -- <path>` also answers for `.` and other pathspecs; only the exact entry counts.
    if name != path {
        return Ok(None);
    }
    let fields: Vec<&str> = meta.split(' ').collect();
    if fields.len() != 3 || fields[1] != "blob" || !(fields[0] == "100644" || fields[0] == "100755")
    {
        return Ok(Some(Err("not-a-regular-file")));
    }
    let bytes = git(repo, &["cat-file", "blob", fields[2]])?;
    Ok(Some(String::from_utf8(bytes).map_err(|_| "not-utf8")))
}

fn file(path: &str, role: &str, included_by: Option<&str>, text: String) -> File {
    File {
        path: path.to_string(),
        role: role.to_string(),
        included_by: included_by.map(str::to_string),
        bytes: text.len(),
        sha256: crate::digest(text.as_bytes()),
        text,
    }
}

/// Markdown links counted toward [`LINK_LIMIT`]: inline links `[text](destination)` and
/// reference definitions `[label]: destination` (up to three spaces of indentation). Spaces,
/// tabs and one line break may come before a destination and a destination may be written in
/// `<…>`, as CommonMark allows (AC-4 record 6841710561). `http:`, `https:`, `mailto:` and `#`
/// destinations are not counted.
fn counted_links(text: &str) -> usize {
    let mut count = 0;
    let mut rest = text;
    while let Some(at) = rest.find("](") {
        let after = &rest[at + 2..];
        if counted_destination(destination(after)) {
            count += 1;
        }
        rest = after;
    }
    for line in text.lines() {
        let indent = line.len() - line.trim_start_matches(' ').len();
        let Some(label) = line[indent..].strip_prefix('[').filter(|_| indent <= 3) else {
            continue;
        };
        let Some(close) = label.find("]:") else {
            continue;
        };
        if !label[..close].contains(']') && counted_destination(destination(&label[close + 2..])) {
            count += 1;
        }
    }
    count
}

/// The destination that starts `after`: leading spaces and tabs and at most one line break
/// skipped, then a `<…>` destination unwrapped or a bare one up to `)` or whitespace.
fn destination(after: &str) -> &str {
    let mut rest = after;
    let mut line_breaks = 0;
    loop {
        if let Some(r) = rest.strip_prefix([' ', '\t', '\r']) {
            rest = r;
        } else if let Some(r) = rest.strip_prefix('\n').filter(|_| line_breaks == 0) {
            line_breaks += 1;
            rest = r;
        } else {
            break;
        }
    }
    if let Some(inner) = rest.strip_prefix('<') {
        return inner.split(['>', '\n']).next().unwrap_or("").trim();
    }
    rest.split([')', ' ', '\t', '\n']).next().unwrap_or("")
}

fn counted_destination(destination: &str) -> bool {
    let lower = destination.to_ascii_lowercase();
    !destination.is_empty()
        && !lower.starts_with("http:")
        && !lower.starts_with("https:")
        && !lower.starts_with("mailto:")
        && !destination.starts_with('#')
}

/// `@name` lines of the root file, in order, without repeats.
fn include_lines(text: &str) -> Vec<String> {
    let mut seen = BTreeSet::new();
    let mut out = Vec::new();
    for line in text.lines() {
        let line = line.strip_suffix('\r').unwrap_or(line);
        if let Some(name) = line.strip_prefix('@') {
            if !name.is_empty()
                && !name.contains(char::is_whitespace)
                && seen.insert(name.to_string())
            {
                out.push(name.to_string());
            }
        }
    }
    out
}

fn skipped_dir(dir: &str) -> bool {
    dir.split('/').any(|part| {
        let part = part.to_ascii_lowercase();
        SKIPPED_DIRS.contains(&part.as_str())
    })
}

/// Every proper ancestor directory of the changed paths, shallow first, sorted.
fn ancestor_dirs(changed: &[Changed]) -> Vec<String> {
    let mut dirs = BTreeSet::new();
    for c in changed {
        let mut path = c.path.as_str();
        while let Some((parent, _)) = path.rsplit_once('/') {
            dirs.insert(parent.to_string());
            path = parent;
        }
    }
    let mut dirs: Vec<String> = dirs.into_iter().collect();
    dirs.sort_by(|a, b| {
        a.matches('/')
            .count()
            .cmp(&b.matches('/').count())
            .then(a.cmp(b))
    });
    dirs
}

pub fn package(repo: &Path, base: &str, head: &str) -> Result<Package> {
    let top = toplevel(repo)?;
    let repo = top.as_path();
    let base = commit_of(repo, base)?;
    let head = commit_of(repo, head)?;
    let merge_base = String::from_utf8_lossy(&git(repo, &["merge-base", &base, &head])?)
        .trim()
        .to_string();
    if merge_base.is_empty() {
        return Err("base and head share no history".into());
    }
    // Attributes come from an empty tree, never from the worktree, whose `.gitattributes` may
    // be uncommitted or belong to another branch (B1). `.git/info/attributes` is read whatever
    // the attribute source, so a repository that has one is refused.
    let info =
        String::from_utf8_lossy(&git(repo, &["rev-parse", "--git-path", "info/attributes"])?)
            .trim()
            .to_string();
    let info = if Path::new(&info).is_absolute() {
        Path::new(&info).to_path_buf()
    } else {
        repo.join(&info)
    };
    if std::fs::read(&info).is_ok_and(|b| b.iter().any(|c| !c.is_ascii_whitespace())) {
        return Err(format!(
            "{} sets attributes that would change the package; remove it or build the package in another clone",
            info.display()
        ));
    }
    let empty_tree =
        String::from_utf8_lossy(&git(repo, &["hash-object", "-t", "tree", "--stdin"])?)
            .trim()
            .to_string();
    let attr_source = format!("--attr-source={empty_tree}");
    // Every diff setting a repository's own `.git/config` could change is fixed here, on the
    // command line, which wins over that file (audit FAIL 6839909483 B1).
    let fixed = [
        "--no-ext-diff",
        "--full-index",
        "--inter-hunk-context=0",
        "--indent-heuristic",
        "--no-relative",
        "--ignore-submodules=none",
        "--submodule=short",
        "--no-textconv",
        "--no-color",
        "--no-renames",
        "--src-prefix=a/",
        "--dst-prefix=b/",
        "--diff-algorithm=myers",
        "-U3",
    ];
    let order_none = format!("-O{}", null_device());
    let mut args = diff_prefix(&attr_source);
    args.push(&order_none);
    args.extend(["--name-status", "-z"]);
    args.extend(fixed);
    args.extend([merge_base.as_str(), head.as_str()]);
    let listing = git(repo, &args)?;
    let fields: Vec<String> = listing
        .split(|b| *b == 0)
        .filter(|f| !f.is_empty())
        .map(|f| String::from_utf8_lossy(f).to_string())
        .collect();
    let changed: Vec<Changed> = fields
        .chunks(2)
        .filter(|pair| pair.len() == 2)
        .map(|pair| Changed {
            status: pair[0].clone(),
            path: pair[1].clone(),
        })
        .collect();
    let mut args = diff_prefix(&attr_source);
    args.push(&order_none);
    args.extend(fixed);
    args.extend([merge_base.as_str(), head.as_str()]);
    let diff = String::from_utf8_lossy(&git(repo, &args)?).to_string();

    let mut review = Vec::new();
    let mut skipped = Vec::new();
    // Root file and its includes.
    match blob_text(repo, &base, REVIEW_FILE)? {
        None => {}
        Some(Err(reason)) => skipped.push(Skipped {
            path: REVIEW_FILE.into(),
            reason: reason.into(),
        }),
        Some(Ok(root)) => {
            let mut files = vec![file(REVIEW_FILE, "root", None, root.clone())];
            let mut used = root.len();
            // Different include files taken so far (the root file is not one of them).
            let mut includes = 0;
            if used > BYTE_LIMIT {
                skipped.push(Skipped {
                    path: REVIEW_FILE.into(),
                    reason: "byte-limit".into(),
                });
                files.clear();
            }
            let root_read = !files.is_empty();
            for name in include_lines(&root).into_iter().filter(|_| root_read) {
                let reason = if name.starts_with(".claude/skills/") {
                    Some("skill-include-not-supported")
                } else if name.contains('/') || name.contains('\\') || name == "." || name == ".." {
                    Some("include-not-beside-root")
                } else if includes >= INCLUDE_LIMIT {
                    Some("include-count")
                } else {
                    None
                };
                if let Some(reason) = reason {
                    skipped.push(Skipped {
                        path: name,
                        reason: reason.into(),
                    });
                    continue;
                }
                // A git failure on one include line skips that include, not the package (B2).
                let found = blob_text(repo, &base, &name).unwrap_or(Some(Err("unreadable")));
                match found {
                    None => skipped.push(Skipped {
                        path: name,
                        reason: "missing".into(),
                    }),
                    Some(Err(reason)) => skipped.push(Skipped {
                        path: name,
                        reason: reason.into(),
                    }),
                    Some(Ok(text)) => {
                        includes += 1;
                        if used + text.len() > BYTE_LIMIT {
                            skipped.push(Skipped {
                                path: name,
                                reason: "byte-limit".into(),
                            });
                        } else {
                            used += text.len();
                            files.push(file(&name, "include", Some(REVIEW_FILE), text));
                        }
                    }
                }
            }
            let links: usize = files.iter().map(|f| counted_links(&f.text)).sum();
            if links > LINK_LIMIT {
                skipped.push(Skipped {
                    path: REVIEW_FILE.into(),
                    reason: "too-many-links".into(),
                });
            } else {
                review.extend(files);
            }
        }
    }
    // Subdirectory files on the changed paths, sharing their own byte limit.
    let dirs = ancestor_dirs(&changed);
    let mut used = 0;
    for dir in &dirs {
        let path = format!("{dir}/{REVIEW_FILE}");
        match blob_text(repo, &base, &path)? {
            None => {}
            Some(_) if skipped_dir(dir) => skipped.push(Skipped {
                path,
                reason: "skipped-directory-name".into(),
            }),
            Some(Err(reason)) => skipped.push(Skipped {
                path,
                reason: reason.into(),
            }),
            Some(Ok(text)) => {
                if used + text.len() > BYTE_LIMIT {
                    skipped.push(Skipped {
                        path,
                        reason: "byte-limit".into(),
                    });
                } else {
                    used += text.len();
                    review.push(file(&path, "subdirectory", None, text));
                }
            }
        }
    }
    // Context: the root and every ancestor directory, each file within the byte limit.
    let mut context = Vec::new();
    for dir in std::iter::once(String::new()).chain(dirs) {
        for name in CONTEXT_FILES {
            let path = if dir.is_empty() {
                name.to_string()
            } else {
                format!("{dir}/{name}")
            };
            match blob_text(repo, &base, &path)? {
                None => {}
                Some(Err(reason)) => skipped.push(Skipped {
                    path,
                    reason: reason.into(),
                }),
                Some(Ok(text)) if text.len() > BYTE_LIMIT => skipped.push(Skipped {
                    path,
                    reason: "byte-limit".into(),
                }),
                Some(Ok(text)) => context.push(file(&path, name, None, text)),
            }
        }
    }
    let mut package = Package {
        schema_version: 1,
        diff_sha256: crate::digest(diff.as_bytes()),
        base,
        head,
        merge_base,
        protection: "not_checked".into(),
        changed,
        diff,
        review,
        context,
        skipped,
        package_digest: String::new(),
    };
    let body = serde_json::to_vec(&package).map_err(|_| "encode failed".to_string())?;
    package.package_digest = crate::digest(&body);
    Ok(package)
}

/// One finding accepted by [`check`].
#[derive(Debug, Clone, Serialize, PartialEq, Eq)]
pub struct Finding {
    pub severity: String,
    pub file: String,
    pub line: u64,
    pub title: String,
    pub evidence: String,
    pub method: String,
}

#[derive(Debug, Clone, Serialize, PartialEq, Eq)]
pub struct Counts {
    pub important: usize,
    pub nit: usize,
    pub pre_existing: usize,
}

#[derive(Debug, Clone, Serialize, PartialEq, Eq)]
pub struct CheckReport {
    pub schema_version: u32,
    pub head: String,
    /// `FAIL` when any finding is `important` (design D3), else `PASS`.
    pub verdict: String,
    pub counts: Counts,
    pub findings: Vec<Finding>,
    /// The post body: a `Severity | File:Line | Issue` table and a machine line.
    pub body: String,
    pub digest: String,
}

/// How far (in lines) quoted evidence may sit from the cited line.
pub const EVIDENCE_WINDOW: u64 = 3;
/// The marker of the post body's machine-readable line.
pub const MACHINE_MARKER: &str = "moai-review-severity";
const SEVERITIES: [&str; 3] = ["important", "nit", "pre_existing"];

fn squash(text: &str) -> String {
    text.split_whitespace().collect::<Vec<_>>().join(" ")
}

/// Text for one table cell: one line, `|` escaped, and `&`, `<`, `>` as entities so a title or
/// path can neither split the table nor open an HTML comment that forges the machine line
/// (pre-audit 6840720403 B1, N1).
fn table_cell(text: &str) -> String {
    squash(text)
        .replace('&', "&amp;")
        .replace('<', "&lt;")
        .replace('>', "&gt;")
        .replace('|', "\\|")
}

fn text_field<'a>(v: &'a serde_json::Value, key: &str) -> Option<&'a str> {
    v.get(key)?.as_str().filter(|s| !s.trim().is_empty())
}

/// `review-check` (#2614 AC-3): every finding needs a known severity, a `file` that is a
/// regular file in HEAD, a `line` inside it, a title, evidence quoted from the code within
/// [`EVIDENCE_WINDOW`] lines of that line (whitespace-insensitive), and a verification with a
/// method and the result `confirmed`. Any malformed finding refuses the whole set, naming each
/// one by index and reason, so nothing unverified is posted. Reads git only.
pub fn check(repo: &Path, head: &str, input: &serde_json::Value) -> Result<CheckReport> {
    let top = toplevel(repo)?;
    let repo = top.as_path();
    if input.get("schema_version").and_then(|v| v.as_u64()) != Some(1) {
        return Err("unsupported schema_version; supported: 1".into());
    }
    let list = input
        .get("findings")
        .and_then(|v| v.as_array())
        .ok_or("findings must be an array")?;
    let head = commit_of(repo, head)?;
    let mut problems = Vec::new();
    let mut accepted: Vec<Finding> = Vec::new();
    for (i, f) in list.iter().enumerate() {
        let severity = f.get("severity").and_then(|v| v.as_str()).unwrap_or("");
        if !SEVERITIES.contains(&severity) {
            problems.push(format!("finding {i}: severity"));
            continue;
        }
        // The machine line's marker is reserved; a title naming it could pass for a second
        // machine line in the posted body (pre-audit 6840720403 B1).
        let Some(title) = text_field(f, "title").filter(|t| !t.contains(MACHINE_MARKER)) else {
            problems.push(format!("finding {i}: title"));
            continue;
        };
        let Some(evidence) = text_field(f, "evidence") else {
            problems.push(format!("finding {i}: evidence"));
            continue;
        };
        let verification = f.get("verification");
        let method = verification.and_then(|v| text_field(v, "method"));
        let result = verification
            .and_then(|v| v.get("result"))
            .and_then(|v| v.as_str());
        let Some(method) = method.filter(|_| result == Some("confirmed")) else {
            problems.push(format!("finding {i}: not-verified"));
            continue;
        };
        let file = f.get("file").and_then(|v| v.as_str()).unwrap_or("");
        let plain = !file.is_empty()
            && !file.starts_with('/')
            && !file.contains('\\')
            && file
                .split('/')
                .all(|p| !p.is_empty() && p != "." && p != "..");
        let text = match plain.then(|| blob_text(repo, &head, file)) {
            Some(Ok(Some(Ok(text)))) => text,
            _ => {
                problems.push(format!("finding {i}: file-not-in-head"));
                continue;
            }
        };
        let lines: Vec<&str> = text.lines().collect();
        let line = f.get("line").and_then(|v| v.as_u64()).unwrap_or(0);
        if line == 0 || line > lines.len() as u64 {
            problems.push(format!("finding {i}: line-out-of-range"));
            continue;
        }
        let from = line.saturating_sub(EVIDENCE_WINDOW).max(1) as usize - 1;
        let to = (line + EVIDENCE_WINDOW).min(lines.len() as u64) as usize;
        if !squash(&lines[from..to].join("\n")).contains(&squash(evidence)) {
            problems.push(format!("finding {i}: evidence-not-at-line"));
            continue;
        }
        let finding = Finding {
            severity: severity.into(),
            file: file.into(),
            line,
            title: title.into(),
            evidence: evidence.into(),
            method: method.into(),
        };
        if accepted
            .iter()
            .any(|a| a.file == finding.file && a.line == finding.line && a.title == finding.title)
        {
            problems.push(format!("finding {i}: duplicate"));
            continue;
        }
        accepted.push(finding);
    }
    if !problems.is_empty() {
        return Err(format!("review findings refused: {}", problems.join("; ")));
    }
    let rank = |s: &str| {
        SEVERITIES
            .iter()
            .position(|x| *x == s)
            .unwrap_or(SEVERITIES.len())
    };
    accepted.sort_by(|a, b| {
        rank(&a.severity)
            .cmp(&rank(&b.severity))
            .then(a.file.cmp(&b.file))
            .then(a.line.cmp(&b.line))
            .then(a.title.cmp(&b.title))
    });
    let count = |s: &str| accepted.iter().filter(|f| f.severity == s).count();
    let counts = Counts {
        important: count("important"),
        nit: count("nit"),
        pre_existing: count("pre_existing"),
    };
    let mut body = String::from("| Severity | File:Line | Issue |\n| --- | --- | --- |\n");
    for f in &accepted {
        let label = match f.severity.as_str() {
            "important" => "Important",
            "nit" => "Nit",
            _ => "Pre-existing",
        };
        body.push_str(&format!(
            "| {label} | {}:{} | {} |\n",
            table_cell(&f.file),
            f.line,
            table_cell(&f.title)
        ));
    }
    body.push_str(&format!(
        "\n<!-- {MACHINE_MARKER}: {{\"important\":{},\"nit\":{},\"pre_existing\":{}}} -->\n",
        counts.important, counts.nit, counts.pre_existing
    ));
    let mut report = CheckReport {
        schema_version: 1,
        head,
        verdict: if counts.important > 0 { "FAIL" } else { "PASS" }.into(),
        counts,
        findings: accepted,
        body,
        digest: String::new(),
    };
    let bytes = serde_json::to_vec(&report).map_err(|_| "encode failed".to_string())?;
    report.digest = crate::digest(&bytes);
    Ok(report)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn links_count_only_relative_targets() {
        assert_eq!(
            counted_links("[a](x.md) [b](https://e) [c](HTTP://e) [d](mailto:m) [e](#f)"),
            1
        );
        assert_eq!(counted_links("no links ] ( here"), 0);
    }

    /// AC-4 record 6841710561 (found by Codex CLI): CommonMark allows spaces, tabs and one
    /// line break before an inline link's destination, a destination in `<…>`, and reference
    /// definitions `[label]: destination`. All of them count; the exempt schemes stay exempt.
    #[test]
    fn links_with_space_angle_or_reference_destinations_count() {
        assert_eq!(counted_links("[a]( x.md)"), 1);
        assert_eq!(counted_links("[a](\tx.md)"), 1);
        assert_eq!(counted_links("[a](\n  x.md)"), 1);
        assert_eq!(counted_links("[a](<x y.md>)"), 1);
        assert_eq!(counted_links("[a]( https://e)"), 0);
        assert_eq!(counted_links("[a](< mailto:m >)"), 0);
        assert_eq!(counted_links("[a]( #f)"), 0);
        assert_eq!(
            counted_links("[ref]: x.md\n   [two]: <y.md>\n[web]: https://e\n"),
            2
        );
        // Four spaces of indentation is a code block, not a definition.
        assert_eq!(counted_links("    [ref]: x.md\n"), 0);
        // Two line breaks end the link: nothing is counted.
        assert_eq!(counted_links("[a](\n\nx.md)"), 0);
    }

    #[test]
    fn include_lines_are_whole_lines_without_spaces_or_repeats() {
        assert_eq!(
            include_lines("@a.md\n @b.md\n@c d.md\n@a.md\r\n@e.md\r\ntext @f.md\n@\n"),
            ["a.md", "e.md"]
        );
    }

    #[test]
    fn skipped_directory_names_match_whole_parts_in_any_case() {
        assert!(skipped_dir("src/Tests"));
        assert!(skipped_dir("contrib/x"));
        assert!(!skipped_dir("src/testing"));
        assert!(!skipped_dir("latest"));
    }
}
