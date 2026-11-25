//! `moai hook write-boundary` (H3, SPEC-20261126-hook-write-boundary): the PreToolUse judgment of
//! a `Write`, `Edit` or `NotebookEdit` call. The handler reads the path and the new content the
//! tool is about to write and denies a call that
//!
//! - reaches a protected place: a `.git` folder, a credential folder or a private-key file
//!   (class `protected-path`),
//! - reaches a file of another checkout of the session's repository, the primary checkout or a
//!   linked worktree that is not the session's own (class `other-checkout`), or
//! - carries a private-key block or an access token (class `secret-content`).
//!
//! A call that breaks none of them gets no decision at all: the handler prints nothing, never
//! `ask`, never `allow`, never added context. The session checkout is the nearest ancestor of
//! the hook input `cwd` that has a `.git` entry, read from the file system without a process
//! (the definition of the Bash boundary hook, so the two never disagree). A path in no checkout
//! of that repository (a temporary folder, the scratchpad, the home state folder, the user's
//! Claude folder) is not a check of this handler. The judgment never reads `permission_mode`, the
//! environment, the network or the card ledger, and writes no file.

use serde_json::{json, Map, Value};
use std::fs;
use std::path::{Component, Path, PathBuf};

/// What the entry hands over: the folder `CLAUDE_PROJECT_DIR` names, if set, and the folder the
/// hook process started in.
pub struct Env<'a> {
    pub project_dir: Option<&'a Path>,
    pub process_cwd: &'a Path,
}

/// The reason classes, in the order the checks run.
const PROTECTED: &str = "protected-path";
const OTHER: &str = "other-checkout";
const SECRET: &str = "secret-content";

const CREDENTIAL_DIRS: &[&str] = &[".ssh", ".aws", ".azure", ".kube", ".gnupg"];
const KEY_FILES: &[&str] = &["id_rsa", "id_ed25519", "id_ecdsa", "id_dsa"];

/// Judges one hook input. `None` is no decision; `Some` is the deny object to print.
pub fn decide(input: &Map<String, Value>, env: &Env) -> Option<Value> {
    let (path_key, content_key) = match input.get("tool_name")?.as_str()? {
        "Write" => ("file_path", "content"),
        "Edit" => ("file_path", "new_string"),
        "NotebookEdit" => ("notebook_path", "new_source"),
        _ => return None,
    };
    let args = input.get("tool_input")?.as_object()?;
    let target = args.get(path_key)?.as_str().filter(|p| !p.is_empty())?;
    let content = args.get(content_key).and_then(Value::as_str);
    let cwd = input
        .get("cwd")
        .and_then(Value::as_str)
        .map(PathBuf::from)
        .filter(|p| p.is_absolute())
        .unwrap_or_else(|| env.process_cwd.to_path_buf());
    classify(target, content, &cwd, env).map(deny)
}

fn classify(target: &str, content: Option<&str>, cwd: &Path, env: &Env) -> Option<&'static str> {
    let written = if Path::new(target).is_absolute() {
        PathBuf::from(target)
    } else {
        cwd.join(target)
    };
    let resolved = resolve(&written);
    if is_protected(&written) || resolved.as_deref().is_some_and(is_protected) {
        return Some(PROTECTED);
    }
    if let (Some(own), Some(there)) = (
        session_checkout(cwd, env),
        resolved.as_deref().and_then(checkout_of),
    ) {
        if own.common == there.common && own.root != there.root {
            return Some(OTHER);
        }
    }
    if content.is_some_and(carries_secret) {
        return Some(SECRET);
    }
    None
}

fn deny(class: &str) -> Value {
    let rule = match class {
        PROTECTED => {
            "a write into a .git folder, a credential folder or a private-key file is refused; \
             change files of your own checkout instead"
        }
        OTHER => {
            "this file lies in another checkout of the repository; write in your own worktree \
             and let the leader land the change"
        }
        _ => {
            "the content carries a private key or an access token; keep secrets out of files \
             and name the variable instead"
        }
    };
    json!({
        "hookSpecificOutput": {
            "hookEventName": "PreToolUse",
            "permissionDecision": "deny",
            "permissionDecisionReason": format!("moai-boundary/{class}: {rule}"),
        }
    })
}

// ------------------------------------------------------------------ paths

struct Checkout {
    root: PathBuf,
    common: PathBuf,
}

/// The checkout of the session: the one `cwd` lies in, unless `CLAUDE_PROJECT_DIR` names a folder
/// of another repository, which then is the session's own.
fn session_checkout(cwd: &Path, env: &Env) -> Option<Checkout> {
    let here = resolve(cwd).as_deref().and_then(checkout_of);
    let project = env
        .project_dir
        .and_then(resolve)
        .as_deref()
        .and_then(checkout_of);
    match (here, project) {
        (Some(h), Some(p)) if h.common != p.common => Some(p),
        (h, _) => h,
    }
}

/// True when a component of the path, written or resolved, is a protected place.
fn is_protected(path: &Path) -> bool {
    let parts: Vec<String> = path
        .components()
        .filter_map(|c| match c {
            Component::Normal(s) => Some(s.to_string_lossy().into_owned()),
            _ => None,
        })
        .collect();
    let has = |name: &str| parts.iter().any(|p| p == name);
    if has(".git") || CREDENTIAL_DIRS.iter().any(|d| has(d)) {
        return true;
    }
    if parts
        .windows(2)
        .any(|w| w[0] == ".config" && w[1] == "gcloud")
    {
        return true;
    }
    parts
        .last()
        .is_some_and(|name| KEY_FILES.contains(&name.as_str()))
}

/// Resolves links component by component, as the kernel does for an absolute path; a component
/// that does not exist is kept as written, so the path of a new file in a missing folder is
/// judged by the folders that exist. `None` after more than 32 links.
fn resolve(path: &Path) -> Option<PathBuf> {
    let mut links = 0u32;
    resolve_inner(path, &mut links)
}

fn resolve_inner(path: &Path, links: &mut u32) -> Option<PathBuf> {
    let mut cur = PathBuf::from("/");
    for comp in path.components() {
        match comp {
            Component::RootDir | Component::Prefix(_) => cur = PathBuf::from("/"),
            Component::CurDir => {}
            Component::ParentDir => {
                cur.pop();
            }
            Component::Normal(name) => {
                let next = cur.join(name);
                match fs::symlink_metadata(&next) {
                    Ok(m) if m.file_type().is_symlink() => {
                        *links += 1;
                        if *links > 32 {
                            return None;
                        }
                        let target = fs::read_link(&next).ok()?;
                        let base = if target.is_absolute() {
                            target
                        } else {
                            cur.join(target)
                        };
                        cur = resolve_inner(&base, links)?;
                    }
                    _ => cur = next,
                }
            }
        }
    }
    Some(cur)
}

/// The checkout holding `path` (the nearest ancestor with a `.git` entry), read without a
/// process: a `.git` folder is a primary checkout, a `.git` file names the folder of a linked
/// worktree whose `commondir` leads to the primary `.git`.
fn checkout_of(path: &Path) -> Option<Checkout> {
    let mut cur = Some(path);
    while let Some(dir) = cur {
        let dotgit = dir.join(".git");
        if let Ok(meta) = fs::symlink_metadata(&dotgit) {
            if meta.is_dir() {
                return Some(Checkout {
                    root: dir.to_path_buf(),
                    common: dotgit,
                });
            }
            if meta.is_file() {
                let text = fs::read_to_string(&dotgit).ok()?;
                let rel = text.lines().next()?.strip_prefix("gitdir:")?.trim();
                let gitdir = resolve(&dir.join(rel))?;
                let common = match fs::read_to_string(gitdir.join("commondir")) {
                    Ok(c) => resolve(&gitdir.join(c.trim()))?,
                    Err(_) => gitdir,
                };
                return Some(Checkout {
                    root: dir.to_path_buf(),
                    common,
                });
            }
        }
        cur = dir.parent();
    }
    None
}

// ---------------------------------------------------------------- secrets

/// One access-token shape: a prefix, how many accepted characters must follow it, which
/// characters count, and whether the prefix counts only at a word start (first in the content or
/// after a character that is not an ASCII letter or digit). A candidate shorter than `min` is not
/// a token, so prose that only names a prefix is allowed. The values are the data of the live
/// write guard (`src/hook_guards.rs`, `TOKEN_SHAPES`), kept equal until that guard is retired.
struct Shape {
    prefix: &'static str,
    min: usize,
    accepts: fn(char) -> bool,
    word_start: bool,
}

fn alnum(c: char) -> bool {
    c.is_ascii_alphanumeric()
}

fn alnum_underscore(c: char) -> bool {
    c.is_ascii_alphanumeric() || c == '_'
}

fn alnum_hyphen(c: char) -> bool {
    c.is_ascii_alphanumeric() || c == '-'
}

fn alnum_hyphen_underscore(c: char) -> bool {
    c.is_ascii_alphanumeric() || c == '-' || c == '_'
}

fn upper_digit(c: char) -> bool {
    c.is_ascii_uppercase() || c.is_ascii_digit()
}

const fn shape(
    prefix: &'static str,
    min: usize,
    accepts: fn(char) -> bool,
    word_start: bool,
) -> Shape {
    Shape {
        prefix,
        min,
        accepts,
        word_start,
    }
}

const SHAPES: &[Shape] = &[
    // GitHub personal, fine-grained and OAuth tokens, GitLab personal access token.
    shape("ghp_", 36, alnum, false),
    shape("github_pat_", 22, alnum_underscore, false),
    shape("gho_", 36, alnum, false),
    shape("glpat-", 20, alnum_hyphen_underscore, false),
    // OpenAI-style key, then the project, service, admin and Anthropic keys, whose name and
    // hyphens follow a short prefix.
    shape("sk-", 32, alnum, true),
    shape("sk-proj-", 40, alnum_hyphen_underscore, true),
    shape("sk-svcacct-", 40, alnum_hyphen_underscore, true),
    shape("sk-admin-", 40, alnum_hyphen_underscore, true),
    shape("sk-ant-", 40, alnum_hyphen_underscore, true),
    // Slack: the letter after the three-letter prefix is one of b, a, p, r, s.
    shape("xoxb-", 10, alnum_hyphen, true),
    shape("xoxa-", 10, alnum_hyphen, true),
    shape("xoxp-", 10, alnum_hyphen, true),
    shape("xoxr-", 10, alnum_hyphen, true),
    shape("xoxs-", 10, alnum_hyphen, true),
    // AWS access key id, long-term and temporary.
    shape("AKIA", 16, upper_digit, true),
    shape("ASIA", 16, upper_digit, true),
    // Google OAuth access token.
    shape("ya29.", 20, alnum_hyphen_underscore, true),
];

fn carries_secret(content: &str) -> bool {
    // Assembled so this file itself never holds the whole marker.
    let begin = concat!("-----", "BEGIN");
    let private = concat!("PRIVATE", " KEY-----");
    if content.contains(begin) && content.contains(private) {
        return true;
    }
    SHAPES.iter().any(|s| has_token(content, s))
}

/// True when the prefix of `s`, at a word start where the shape needs one, is followed by at
/// least `s.min` accepted characters. A run is counted only up to `min`, so the scan grows with
/// the content and never with the length of one run.
fn has_token(content: &str, s: &Shape) -> bool {
    content.match_indices(s.prefix).any(|(at, _)| {
        let at_word_start = !s.word_start
            || content[..at]
                .chars()
                .next_back()
                .is_none_or(|c| !c.is_ascii_alphanumeric());
        at_word_start
            && content[at + s.prefix.len()..]
                .chars()
                .take(s.min)
                .take_while(|c| (s.accepts)(*c))
                .count()
                == s.min
    })
}

#[cfg(test)]
mod tests {
    use super::*;

    fn token(prefix: &str, n: usize) -> String {
        format!("{prefix}{}", "a".repeat(n))
    }

    #[test]
    fn a_shape_needs_its_minimum_length_after_the_prefix() {
        for s in SHAPES {
            let c = if (s.accepts)('a') { 'a' } else { 'A' };
            let full = format!("{} {}{}", "x", s.prefix, c.to_string().repeat(s.min));
            let short = format!("{} {}{}", "x", s.prefix, c.to_string().repeat(s.min - 1));
            assert!(carries_secret(&full), "{} at {}", s.prefix, s.min);
            assert!(!carries_secret(&short), "{} one short", s.prefix);
        }
    }

    #[test]
    fn the_openai_prefix_counts_only_at_a_word_start() {
        assert!(!carries_secret(&token("task-", 40)));
        assert!(carries_secret(&token("sk-", 32)));
        assert!(carries_secret(&format!("key={}", token("sk-", 32))));
    }

    #[test]
    fn a_private_key_block_needs_both_halves() {
        let begin = format!("{}{}", "-----", "BEGIN RSA ");
        let end = format!("{} {}", "PRIVATE", "KEY-----");
        assert!(carries_secret(&format!("{begin}{end}")));
        assert!(!carries_secret(&begin));
        assert!(!carries_secret(&end));
    }

    #[test]
    fn protected_places_match_whole_components_only() {
        for yes in [
            "/a/.git/config",
            "/a/.ssh/x",
            "/a/.config/gcloud/x",
            "/a/id_rsa",
        ] {
            assert!(is_protected(Path::new(yes)), "{yes}");
        }
        for no in [
            "/a/.gitignore",
            "/a/.github/x",
            "/a/my.ssh.md",
            "/a/id_rsa.pub",
        ] {
            assert!(!is_protected(Path::new(no)), "{no}");
        }
    }
}
