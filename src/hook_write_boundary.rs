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
/// A path whose links cannot be followed: denied in the class `protected-path` with its own rule.
const UNFOLLOWED: &str = "unfollowed-link";
/// An input too large to read: denied in the class `protected-path` with its own rule.
const OVERSIZE: &str = "over-limit-input";

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

/// The denial of a call whose input is over the entry's limit: neither its path nor its content
/// was read, so no check can judge it (REQ-13).
pub fn over_limit() -> Value {
    deny(OVERSIZE)
}

/// The denial of a call whose input the entry could not read, by the entry's receipt word:
/// `not_utf8`, `not_json`, else `not_object`. Neither path nor content is known, so the class is
/// `protected-path`, with a fixed text per cause (SPEC-20261126-hook-unreadable-input REQ-1,
/// REQ-3).
pub fn unreadable(word: &str) -> Value {
    let cause = match word {
        "not_utf8" => "is not valid UTF-8",
        "not_json" => "is not valid JSON",
        _ => "is not a JSON object",
    };
    deny_object(
        PROTECTED,
        &format!(
            "the hook input {cause}, so its path and content cannot be checked; send the call \
             again with plain text, and if every call is refused, tell the user, who can remove \
             or disable the moai hook entry"
        ),
    )
}

fn classify(target: &str, content: Option<&str>, cwd: &Path, env: &Env) -> Option<&'static str> {
    let written = if Path::new(target).is_absolute() {
        PathBuf::from(target)
    } else {
        cwd.join(target)
    };
    // A path whose links cannot be followed (more than 32, or one that cannot be read) cannot be
    // judged, so it is refused rather than skipped.
    let Some(resolved) = resolve(&written).map(|p| on_disk(&p)) else {
        return Some(UNFOLLOWED);
    };
    if is_protected(&written) || is_protected(&resolved) {
        return Some(PROTECTED);
    }
    if session_checkout(cwd, env).is_some_and(|own| in_other_checkout(&own, &resolved)) {
        return Some(OTHER);
    }
    if content.is_some_and(carries_secret) {
        return Some(SECRET);
    }
    None
}

fn deny(found: &str) -> Value {
    let (class, rule) = match found {
        PROTECTED => (
            PROTECTED,
            "a write into a .git folder, a credential folder or a private-key file is refused; \
             change files of your own checkout instead",
        ),
        UNFOLLOWED => (
            PROTECTED,
            "the symbolic links of this path cannot be followed, so it cannot be checked; \
             write to the real file by its own path",
        ),
        OVERSIZE => (
            PROTECTED,
            "this call is over the 4 MiB the check reads, so its path and content cannot be \
             checked; write the file in calls under 4 MiB",
        ),
        OTHER => (
            OTHER,
            "this file lies in another checkout of the repository; write in your own worktree \
             and let the leader land the change",
        ),
        _ => (
            SECRET,
            "the content carries a private key or an access token; keep secrets out of files \
             and name the variable instead",
        ),
    };
    deny_object(class, rule)
}

/// The deny object of one class and its rule text.
fn deny_object(class: &str, rule: &str) -> Value {
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
    let here = resolve(cwd)
        .map(|p| on_disk(&p))
        .as_deref()
        .and_then(checkout_of);
    let project = env
        .project_dir
        .and_then(resolve)
        .map(|p| on_disk(&p))
        .as_deref()
        .and_then(checkout_of);
    match (here, project) {
        (Some(h), Some(p)) if h.common != p.common => Some(p),
        (h, _) => h,
    }
}

/// A name as a file system that ignores case sees it: ASCII letters in lower case, and the
/// non-ASCII letters that Unicode case folding turns into ASCII letters folded the same way (the
/// long s, the Kelvin sign, the sharp s and the Latin ligatures U+FB00 to U+FB06), so `.GIT`,
/// `.\u{17f}sh` and `.con\u{fb01}g` all spell a protected name. No other letter folds to ASCII
/// letters alone.
fn fold_name(name: &str) -> String {
    let mut out = String::with_capacity(name.len());
    for c in name.chars() {
        match c {
            '\u{17f}' => out.push('s'),
            '\u{212a}' => out.push('k'),
            '\u{df}' | '\u{1e9e}' => out.push_str("ss"),
            '\u{fb00}' => out.push_str("ff"),
            '\u{fb01}' => out.push_str("fi"),
            '\u{fb02}' => out.push_str("fl"),
            '\u{fb03}' => out.push_str("ffi"),
            '\u{fb04}' => out.push_str("ffl"),
            '\u{fb05}' | '\u{fb06}' => out.push_str("st"),
            _ => out.push(c.to_ascii_lowercase()),
        }
    }
    out
}

/// True when a component of the path, written or resolved, is a protected place, compared as
/// `fold_name` spells it (the file system of a Mac ignores case: `.GIT` is the `.git` folder).
fn is_protected(path: &Path) -> bool {
    let parts: Vec<String> = path
        .components()
        .filter_map(|c| match c {
            Component::Normal(s) => Some(fold_name(&s.to_string_lossy())),
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

/// True when `path` lies in a checkout of the session's repository other than its own. The deepest
/// of two answers decides: the checkout found by walking up from the path, and the registered
/// worktrees of the common folder. So a registered worktree folder that lost its `.git` file is
/// still another checkout, and another repository nested inside one is not.
fn in_other_checkout(own: &Checkout, path: &Path) -> bool {
    let depth = |p: &Path| p.components().count();
    let listed = registered_roots(&own.common)
        .into_iter()
        .filter(|root| path.starts_with(root))
        .max_by_key(|root| depth(root));
    match checkout_of(path) {
        Some(f) if listed.as_ref().is_none_or(|r| depth(&f.root) >= depth(r)) => {
            f.common == own.common && f.root != own.root
        }
        _ => listed.is_some_and(|root| root != own.root),
    }
}

/// The checkout roots a common folder knows: the folder holding it when it is named `.git`, and
/// the folder of the `.git` file that each `worktrees/*/gitdir` names.
fn registered_roots(common: &Path) -> Vec<PathBuf> {
    let mut roots = Vec::new();
    if common.ends_with(".git") {
        roots.extend(common.parent().map(Path::to_path_buf));
    }
    if let Ok(entries) = fs::read_dir(common.join("worktrees")) {
        for entry in entries.flatten() {
            let Ok(text) = fs::read_to_string(entry.path().join("gitdir")) else {
                continue;
            };
            if text.trim().is_empty() {
                continue;
            }
            // Absolute, or relative to the folder of this registry entry.
            let named = entry.path().join(text.trim());
            if let Some(root) = named.parent().and_then(resolve) {
                roots.push(on_disk(&root));
            }
        }
    }
    roots
}

/// The spelling the disk holds for `path`: the deepest ancestor that exists is canonicalized (on a
/// file system that ignores case this returns the stored spelling, so `Repo/.GIT` and `repo/.git`
/// compare equal) and the components after it, which do not exist, are kept as written.
fn on_disk(path: &Path) -> PathBuf {
    let mut rest: Vec<&std::ffi::OsStr> = Vec::new();
    let mut cur = path;
    loop {
        if let Ok(real) = fs::canonicalize(cur) {
            return rest.iter().rev().fold(real, |p, name| p.join(name));
        }
        match (cur.file_name(), cur.parent()) {
            (Some(name), Some(parent)) => {
                rest.push(name);
                cur = parent;
            }
            _ => return path.to_path_buf(),
        }
    }
}

/// The checkout holding `path` (the nearest ancestor with a `.git` entry), read without a
/// process: a `.git` folder is a primary checkout, a `.git` file names the folder of a linked
/// worktree whose `commondir` leads to the primary `.git`. A `.git` file that cannot be read or
/// parsed names no checkout, and the search goes on with the parent folder, so a broken file
/// never hides the checkout around it.
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
                if let Some(common) = linked_common(dir, &dotgit) {
                    return Some(Checkout {
                        root: dir.to_path_buf(),
                        common,
                    });
                }
            }
        }
        cur = dir.parent();
    }
    None
}

/// The common `.git` folder a linked worktree's `.git` file leads to, or `None` when the file
/// cannot be read or parsed or a link on the way cannot be followed.
fn linked_common(dir: &Path, dotgit: &Path) -> Option<PathBuf> {
    let text = fs::read_to_string(dotgit).ok()?;
    let rel = text.lines().next()?.strip_prefix("gitdir:")?.trim();
    let gitdir = on_disk(&resolve(&dir.join(rel))?);
    Some(match fs::read_to_string(gitdir.join("commondir")) {
        Ok(c) => on_disk(&resolve(&gitdir.join(c.trim()))?),
        Err(_) => gitdir,
    })
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
    fn folded_spellings_of_a_protected_name_are_protected() {
        for yes in [
            "/a/.GIT/config",
            "/a/.\u{17f}sh/x",
            "/a/.\u{17f}\u{17f}h/x",
            "/a/.\u{df}h/x",
            "/a/.\u{212a}ube/x",
            "/a/ID_R\u{17f}A",
            "/a/.config/GCLOUD/x",
        ] {
            assert!(is_protected(Path::new(yes)), "{yes}");
        }
        assert!(!is_protected(Path::new("/a/.config/G\u{212a}loud/x")));
        assert_eq!(fold_name("\u{fb06}x"), "stx");
        assert_eq!(
            fold_name("\u{fb00}|\u{fb01}|\u{fb02}|\u{fb03}|\u{fb04}|\u{fb05}"),
            "ff|fi|fl|ffi|ffl|st"
        );
        assert!(is_protected(Path::new("/a/.con\u{fb01}g/gcloud/x")));
        assert_eq!(fold_name("\u{e9}"), "\u{e9}");
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
