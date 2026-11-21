//! delete_guard — refuses recursive deletes that can wipe the filesystem root, the
//! home folder, the project root (or its parent), or a path built from a shell
//! variable that may be empty or unset.
//!
//! Pure classification: no git, no filesystem access. The caller (the hook guard,
//! once the wiring piece lands) supplies the project root as read from
//! `git rev-parse --show-toplevel`. Tests include this file by `#[path]`, so
//! `lib.rs` does not declare it yet — the declaration and the `hook_guards.rs`
//! delegation are a separate wiring slice (leader decision, waits on C-234).

use std::path::{Component, Path};

const PREFIX: &str = "MoAI delete-guard:";
const TAIL: &str = " — use a named subfolder, or guard the variable with ${VAR:?}.";

/// The shapes `rm` spells recursion with: `-r`, `-R`, their clusters (`-rf`, `-fr`,
/// `-rRf`, ...) and `--recursive`.
fn recursive(args: &[String]) -> bool {
    args.iter().any(|a| {
        a == "--recursive"
            || (a.starts_with('-') && !a.starts_with("--") && {
                let mut seen = false;
                for c in a.chars().skip(1) {
                    if c == 'r' || c == 'R' {
                        seen = true;
                    }
                }
                seen
            })
    })
}

/// One refusal with the text that names what was hit.
struct Refusal {
    what: String,
}

impl Refusal {
    fn message(&self) -> String {
        format!(
            "{PREFIX} recursive delete of {} is refused{TAIL}",
            self.what
        )
    }
}

/// Classify one command: `rm`, `find` with `-delete`/`-exec rm`, `xargs rm`/`xargs find`,
/// and the `sudo`-wrapped forms of those. Returns the reason line, or `None` when the
/// command deletes nothing dangerous.
pub fn check(name: &str, args: &[String], project_root: Option<&Path>) -> Option<String> {
    let refusal = match name {
        "rm" => rm_check(args, project_root),
        "find" => find_check(args, project_root),
        "xargs" => wrapped_check(args, project_root),
        "sudo" => wrapped_check(args, project_root),
        _ => None,
    }?;
    Some(refusal.message())
}

/// `sudo rm -rf /` / `xargs rm -rf /`: unwrap one layer and re-check the inner command.
fn wrapped_check(args: &[String], root: Option<&Path>) -> Option<Refusal> {
    match args.split_first() {
        Some((first, rest)) if matches!(first.as_str(), "rm" | "find" | "xargs") => {
            match first.as_str() {
                "rm" => rm_check(rest, root),
                "find" => find_check(rest, root),
                _ => wrapped_check(rest, root),
            }
        }
        _ => None,
    }
}

/// Every argument is checked as a possible target: no dangerous target starts with
/// `-`, so option words (and `--`) can never match and need no separate parsing.
fn rm_check(args: &[String], root: Option<&Path>) -> Option<Refusal> {
    if !recursive(args) {
        return None;
    }
    let mut targets = Vec::new();
    let mut past_dd = false;
    for a in args {
        if past_dd {
            targets.push(a);
            continue;
        }
        if a == "--" {
            past_dd = true;
            continue;
        }
        if a.starts_with('-') && a.len() > 1 {
            continue; // an option or an option cluster
        }
        targets.push(a);
    }
    targets.iter().find_map(|t| classify(t, root))
}

/// `find <path> ... -delete` and `find <path> ... -exec rm ...`: the leading
/// non-option words are the paths find walks.
fn find_check(args: &[String], root: Option<&Path>) -> Option<Refusal> {
    let deletes = args.iter().any(|a| a == "-delete")
        || args
            .windows(2)
            .any(|w| (w[0] == "-exec" || w[0] == "-execdir") && w[1] == "rm");
    if !deletes {
        return None;
    }
    let mut paths = Vec::new();
    for a in args {
        if a.starts_with('-') {
            break;
        }
        paths.push(a);
    }
    paths.iter().find_map(|p| classify(p, root))
}

/// A variable-shaped first segment (`$M`, `${M}`, `"$M"` after the shell parser): the
/// expansion may be empty, and an empty expansion makes `rm -rf $M/` delete the
/// working folder. `${VAR:?}` and `${VAR:-lit}` cannot expand empty, so they stay
/// allowed; `${VAR:+word}` and any other spelling is refused.
fn variable_refusal(target: &str) -> Option<Refusal> {
    let first = target.split('/').next().unwrap_or(target);
    if !first.starts_with('$') {
        return None;
    }
    if first == "$HOME" || first == "${HOME}" {
        return Some(home_refusal());
    }
    let inner = first.strip_prefix("${").and_then(|s| s.strip_suffix('}'));
    if let Some(inner) = inner {
        // `${VAR:?}` and `${VAR:-lit}` cannot expand empty — allowed.
        if inner.contains(":?") || inner.contains(":-") {
            return None;
        }
        return Some(empty_variable_refusal());
    }
    let name = &first[1..];
    if name
        .chars()
        .next()
        .is_some_and(|c| c.is_ascii_alphabetic() || c == '_')
    {
        return Some(empty_variable_refusal());
    }
    None
}

fn home_refusal() -> Refusal {
    Refusal {
        what: "the home folder".to_string(),
    }
}

fn empty_variable_refusal() -> Refusal {
    Refusal {
        what: "a path built from a variable that may be empty or unset".to_string(),
    }
}

/// The normalized shape of one target: absolute flag plus the name segments with
/// trailing slashes, `/.`, and duplicate slashes gone.
fn segments(target: &str) -> (bool, Vec<String>) {
    let mut absolute = false;
    let mut out = Vec::new();
    for c in Path::new(target).components() {
        match c {
            Component::RootDir => absolute = true,
            Component::CurDir => {}
            Component::ParentDir => out.push("..".to_string()),
            Component::Normal(s) => out.push(s.to_string_lossy().into_owned()),
            Component::Prefix(_) => {}
        }
    }
    (absolute, out)
}

fn classify(target: &str, root: Option<&Path>) -> Option<Refusal> {
    if let Some(r) = variable_refusal(target) {
        return Some(r);
    }
    // Bare wildcards and dots, and the home tilde, read before path parsing.
    let t = target.trim_end_matches('/');
    if matches!(t, "*" | "." | ".." | "~") {
        let what = match t {
            "~" => "the home folder".to_string(),
            _ => "a bare wildcard or dot path".to_string(),
        };
        return Some(Refusal { what });
    }
    if t == "~/*" {
        return Some(home_refusal());
    }
    let (absolute, segs) = segments(target);
    if segs.last().is_some_and(|s| s == ".git") {
        return Some(Refusal {
            what: "a .git directory".to_string(),
        });
    }
    if absolute {
        if segs.is_empty() {
            return Some(Refusal {
                what: "the filesystem root".to_string(),
            });
        }
        if segs.len() == 1 {
            return Some(Refusal {
                what: format!("the top-level directory /{}", segs[0]),
            });
        }
        if let Some(root) = root {
            let (_, root_segs) = segments(root.to_string_lossy().as_ref());
            if root_segs == segs {
                return Some(Refusal {
                    what: format!("the project root {}", root.display()),
                });
            }
            let parent: &[String] = root_segs.split_last().map(|(_, p)| p).unwrap_or(&[]);
            if parent == segs {
                return Some(Refusal {
                    what: format!(
                        "a parent of the project root ({})",
                        root.parent()
                            .map_or_else(|| "/".to_string(), |p| p.display().to_string(),)
                    ),
                });
            }
        }
    }
    // Relative non-dot targets keep working (named subfolders, scratch paths).
    None
}
