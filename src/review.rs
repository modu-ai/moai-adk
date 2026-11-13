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
//! same bytes on every machine and in every tool (AC-2). Reads only; writes nothing.

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
    let (meta, _) = entry.split_once('\t').ok_or("unreadable tree entry")?;
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

/// Inline Markdown links `[text](target)` whose target is not `http:`, `https:`, `mailto:` or
/// a `#` anchor.
fn counted_links(text: &str) -> usize {
    let mut count = 0;
    let mut rest = text;
    while let Some(at) = rest.find("](") {
        let after = &rest[at + 2..];
        let target = after.split([')', ' ', '\n']).next().unwrap_or("");
        let lower = target.to_ascii_lowercase();
        if !target.is_empty()
            && !lower.starts_with("http:")
            && !lower.starts_with("https:")
            && !lower.starts_with("mailto:")
            && !target.starts_with('#')
        {
            count += 1;
        }
        rest = after;
    }
    count
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
    let base = commit_of(repo, base)?;
    let head = commit_of(repo, head)?;
    let merge_base = String::from_utf8_lossy(&git(repo, &["merge-base", &base, &head])?)
        .trim()
        .to_string();
    if merge_base.is_empty() {
        return Err("base and head share no history".into());
    }
    let fixed = [
        "--no-ext-diff",
        "--no-textconv",
        "--no-color",
        "--no-renames",
        "--src-prefix=a/",
        "--dst-prefix=b/",
        "--diff-algorithm=myers",
        "-U3",
    ];
    let mut args = vec!["-c", "core.quotePath=true", "diff"];
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
    let mut args = vec!["-c", "core.quotePath=true", "diff"];
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
                } else if name.contains('/') || name.contains('\\') {
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
                match blob_text(repo, &base, &name)? {
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
