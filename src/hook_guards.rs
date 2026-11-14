//! tool-preflight guards, #2391 stage 5: the minimal set that keeps everyday development safe
//! without the Go hooks (Issue #2391 classification 6771810987, corrected in 6772429210).
//!
//! Every rule below is a deny with no override, so each one is limited to damage that cannot be
//! undone and is pinned by allowed look-alikes in its tests. Stage 6 adds a separate
//! confirmation class ([`bash_confirm`], decision B2 = A): `git reset --hard`, a forced
//! `git clean` and a force push to another branch are asked about on Claude Code and refused
//! on Codex (which cannot ask); a deny anywhere on the line always wins over a confirmation. The reasons are fixed strings; the command
//! text, paths and file contents are never echoed.
//!
//! - A1 `rm` aimed at `/`, the home directory, a top-level directory, a bare `*`/`.`/`..`, or a
//!   `.git` directory (Go issue #1658). `node_modules` is not included: it can be regenerated.
//! - A2 irreversible commands: a force push to `main`/`master`, a force delete of the
//!   `main`/`master` branch, `mkfs*`, `dd` onto a device, a fork bomb, and a database client
//!   running `DROP DATABASE`, `DROP SCHEMA` or `TRUNCATE TABLE`. The integration branch of this
//!   repository is not covered (a recorded gap; widening it is configuration, not a constant).
//! - A3 `git commit` that skips the repository's hooks: the long bypass option or a unique
//!   abbreviation of it (nine characters or more), `-n` in a short-option cluster before any
//!   option that takes a value (`-uno`, `-m n` and `--message -n` are values, not options), or
//!   `core.hooksPath` set on git's own arguments (`-c`, `-c<key>=`, `--config-env`), any key
//!   case. Measured against git 2.54.0 with a failing pre-commit hook in a fixture repository.
//! - A4 file writes, wherever they land: inside a `.git` directory, into credential directories
//!   (`.ssh`, `.aws`, `.azure`, `.kube`, `.gnupg`, gcloud's configuration), onto a private-key
//!   file named `id_rsa`/`id_ed25519`/`id_ecdsa`/`id_dsa`, or content that carries a private-key
//!   block or a GitHub token. A file extension alone (`.pem`, `.key`) decides nothing: public
//!   certificates and test fixtures use them too. Writes outside the project are not refused
//!   here (decision pending: the hook cannot see a session's extra directories).
//!
//! Known gaps (allowed, not refused): shell redirections (`>`, `tee`) into protected paths; a
//! force push with no refspec while `main` is checked out (the branch is not read here); a
//! `--mirror` or `--all` push; deleting `main`/`master` on a remote (`--delete`, `:main`);
//! `core.hooksPath` set through `GIT_CONFIG_*` environment variables or a config file; Codex
//! `apply_patch` paths are read from `*** Add/Update/Delete File:` and `*** Move to:` lines, a
//! format the official pages do not define — a patch without such lines is allowed without a
//! path check (its added lines are still checked for secrets); input past the 4 MiB stdin
//! bound is an engine error and not checked at all (decision D1).

use serde_json::Value;
use std::path::{Component, Path, PathBuf};

const RM_TARGET: &str =
    "MoAI tool-preflight: rm aimed at the root, the home directory, a top-level directory, a bare wildcard or dot, or a .git directory is refused.";
const FORCE_PUSH: &str = "MoAI tool-preflight: a force push to main or master is refused.";
const BRANCH_DELETE: &str =
    "MoAI tool-preflight: force-deleting the main or master branch is refused.";
const DEVICE: &str =
    "MoAI tool-preflight: formatting a filesystem or writing raw data onto a device is refused.";
const FORK_BOMB: &str = "MoAI tool-preflight: a fork bomb is refused.";
const DATABASE: &str =
    "MoAI tool-preflight: dropping a database or schema, or truncating a table, is refused.";
const HOOK_BYPASS: &str =
    "MoAI tool-preflight: a commit that skips the repository's commit hooks is refused.";
const PROTECTED_PATH: &str =
    "MoAI tool-preflight: writing inside a .git directory, a credential directory or a private-key file is refused.";
const SECRET_CONTENT: &str =
    "MoAI tool-preflight: writing a private key or a GitHub token into a file is refused.";

const PROTECTED_BRANCHES: &[&str] = &["main", "master"];
const DB_CLIENTS: &[&str] = &["psql", "mysql", "mariadb", "sqlite3", "cockroach"];
const DB_STATEMENTS: &[&str] = &["drop database", "drop schema", "truncate table"];
const CREDENTIAL_DIRS: &[&str] = &[".ssh", ".aws", ".azure", ".kube", ".gnupg"];
const KEY_FILES: &[&str] = &["id_rsa", "id_ed25519", "id_ecdsa", "id_dsa"];
/// `git commit` short options that take a value (`-m msg`, `-F file`, `-C rev`, `-c rev`,
/// `-t file`); after one of them the rest of a cluster is that value, not more options.
const COMMIT_VALUE_OPTS: &[char] = &['m', 'F', 'C', 'c', 't'];
/// `git commit` short options whose value can only be attached (`-uno`, `-S<keyid>`): the rest
/// of the cluster is that value, and the next word is not consumed. Measured with git 2.54.0:
/// `-uno` and `-unormal` run the pre-commit hook.
const COMMIT_ATTACHED_OPTS: &[char] = &['u', 'S'];
/// `git commit` long options that take the next word as their value unless written `--x=v`.
const COMMIT_LONG_VALUE_OPTS: &[&str] = &[
    "--message",
    "--file",
    "--author",
    "--date",
    "--reuse-message",
    "--reedit-message",
    "--fixup",
    "--squash",
    "--template",
    "--cleanup",
    "--trailer",
    "--pathspec-from-file",
];
/// The hook-bypass long option, assembled so this file never holds it as one literal. Git
/// accepts a unique abbreviation; with git 2.54.0 the shortest one that skips the hook is
/// nine characters (shorter ones are ambiguous with `--no-verbose`).
fn bypass_flag() -> String {
    format!("--no-{}", "verify")
}
const BYPASS_MIN_PREFIX: usize = 9;

/// The reason a Bash command line is refused, or `None`.
pub fn bash_guard(command: &str) -> Option<&'static str> {
    // A fork bomb defines a function made of punctuation, which the word splitter does not
    // keep, so it is read from text: outside quotes, in the line and in every script the line
    // hands to a shell or `eval`. Quoted data stays data (pre-audit H5-B4); a quoted script
    // that runs is read again (re-audit regression, `sh -c '…'` / `eval '…'`).
    let bomb = |text: &str| {
        let bare = unquoted(text);
        bare.contains(":(){") || bare.contains(": () {")
    };
    if crate::claim_verify::executed_scripts(command)
        .iter()
        .any(|s| bomb(s))
    {
        return Some(FORK_BOMB);
    }
    for (name, args) in crate::claim_verify::shell_commands(command) {
        let reason = match name.as_str() {
            "rm" => rm_guard(&args),
            "git" => git_guard(&args),
            "dd" => args
                .iter()
                .any(|a| is_device_target(a.strip_prefix("of=").unwrap_or("")))
                .then_some(DEVICE),
            n if n.starts_with("mkfs") => Some(DEVICE),
            // A statement inside an SQL string literal is data, not a statement (H5-B5).
            n if DB_CLIENTS.contains(&n) => args
                .iter()
                .any(|a| {
                    let lower = outside_sql_literals(a).to_ascii_lowercase();
                    DB_STATEMENTS.iter().any(|s| lower.contains(s))
                })
                .then_some(DATABASE),
            _ => None,
        };
        if reason.is_some() {
            return reason;
        }
    }
    None
}

/// The shell text with single- and double-quoted spans removed (backslash escapes kept as is).
fn unquoted(text: &str) -> String {
    let mut out = String::new();
    let mut quote: Option<char> = None;
    let mut chars = text.chars();
    while let Some(c) = chars.next() {
        match (quote, c) {
            (None, '\\') => {
                out.push(c);
                out.extend(chars.next());
            }
            (None, '\'' | '"') => quote = Some(c),
            (None, _) => out.push(c),
            (Some('"'), '\\') => {
                chars.next();
            }
            (Some(q), _) if c == q => quote = None,
            (Some(_), _) => {}
        }
    }
    out
}

/// An SQL argument with its single-quoted string literals removed (`''` inside a literal
/// closes and reopens it, which removes the same text).
fn outside_sql_literals(sql: &str) -> String {
    let mut out = String::new();
    let mut in_literal = false;
    for c in sql.chars() {
        if c == '\'' {
            in_literal = !in_literal;
        } else if !in_literal {
            out.push(c);
        }
    }
    out
}

fn is_device_target(path: &str) -> bool {
    path.starts_with("/dev/")
        && !matches!(
            path,
            "/dev/null" | "/dev/zero" | "/dev/stdout" | "/dev/stderr" | "/dev/tty"
        )
}

/// Every argument is checked as a possible target: no dangerous target starts with `-`, so
/// option words (and `--`) can never match and need no separate parsing.
fn rm_guard(args: &[String]) -> Option<&'static str> {
    args.iter()
        .any(|a| is_dangerous_rm_target(a))
        .then_some(RM_TARGET)
}

fn is_dangerous_rm_target(target: &str) -> bool {
    let t = target.trim_end_matches('/');
    let t = if t.is_empty() && target.starts_with('/') {
        "/"
    } else {
        t
    };
    if matches!(
        t,
        "/" | "/*" | "~" | "~/*" | "$HOME" | "${HOME}" | "$HOME/*" | "${HOME}/*" | "*" | "." | ".."
    ) {
        return true;
    }
    // A top-level directory such as `/usr` or `/Users`.
    if let Some(rest) = t.strip_prefix('/') {
        if !rest.is_empty() && !rest.contains('/') {
            return true;
        }
    }
    // A `.git` directory anywhere.
    t == ".git" || t.ends_with("/.git")
}

/// Each subcommand the line may run — an alias defined on the same line resolved to its verb
/// and the words it carries (pre-audit H5-B7, the same reading as the claim check).
fn git_guard(args: &[String]) -> Option<&'static str> {
    for sub in crate::claim_verify::git_subcommands(args) {
        let rest = &sub.rest;
        let reason = match sub.verb.as_str() {
            "push" => force_push_to_protected(rest).then_some(FORCE_PUSH),
            "branch" => force_delete_protected(rest).then_some(BRANCH_DELETE),
            "commit" => {
                (skips_hooks(rest) || overrides_hooks_path(&sub.global)).then_some(HOOK_BYPASS)
            }
            _ => None,
        };
        if reason.is_some() {
            return reason;
        }
    }
    None
}

const CONFIRM_RESET: &str =
    "MoAI tool-preflight: `git reset --hard` discards uncommitted changes; confirm it.";
const CONFIRM_CLEAN: &str =
    "MoAI tool-preflight: a forced `git clean` deletes untracked files; confirm it.";
const CONFIRM_FORCE_PUSH: &str =
    "MoAI tool-preflight: a force push rewrites the remote branch; confirm it.";

/// The confirmation class (decision B2 = A, Issue #2389 c6774899518 row 25): commands that
/// discard work but are ordinary in development, so they are asked about (Claude) rather than
/// refused. Read with the same parser, aliases, short-option clusters and long-option prefix
/// rule ([`long_opt`]) as the deny guards.
/// Only consulted when no deny applies to the line.
pub fn bash_confirm(command: &str) -> Option<&'static str> {
    for (name, args) in crate::claim_verify::shell_commands(command) {
        if name != "git" {
            continue;
        }
        for sub in crate::claim_verify::git_subcommands(&args) {
            let rest = &sub.rest;
            let reason = match sub.verb.as_str() {
                // `--hard` or a prefix git takes for it (`--h` too: no other reset option
                // starts with `h`; measured with git 2.54.0).
                "reset" => rest
                    .iter()
                    .take_while(|a| *a != "--")
                    .any(|a| long_opt(a, "hard", RESET_LONG))
                    .then_some(CONFIRM_RESET),
                "clean" => {
                    let letters = short_letters(rest, &['e']);
                    let has = |w: &str| rest.iter().any(|a| long_opt(a, w, CLEAN_LONG));
                    let forced = letters.contains(&'f') || has("force");
                    let dry = letters.contains(&'n') || has("dry-run");
                    (forced && !dry).then_some(CONFIRM_CLEAN)
                }
                // A force push that the deny guard did not refuse: another branch, or no
                // refspec at all (the current branch is not read here).
                "push" => {
                    let forced_flag = short_letters(rest, &['o']).contains(&'f')
                        || rest.iter().any(|a| push_force_option(a));
                    let plus = push_positionals(rest)
                        .iter()
                        .skip(1)
                        .any(|s| s.starts_with('+'));
                    (forced_flag || plus).then_some(CONFIRM_FORCE_PUSH)
                }
                _ => None,
            };
            if reason.is_some() {
                return reason;
            }
        }
    }
    None
}

/// The option letters of every short-option cluster (`-vf` → v, f) up to `--`. A letter in
/// `value_opts` takes the rest of its cluster, or the next word when it ends the cluster, as
/// its value (pre-audit H5-B3: git aggregates short options).
fn short_letters(rest: &[String], value_opts: &[char]) -> Vec<char> {
    let mut out = Vec::new();
    let mut skip_next = false;
    for a in rest {
        if skip_next {
            skip_next = false;
            continue;
        }
        if a == "--" {
            break;
        }
        if a.starts_with("--") || !a.starts_with('-') || a.len() < 2 {
            continue;
        }
        for (k, c) in a[1..].char_indices() {
            out.push(c);
            if value_opts.contains(&c) {
                skip_next = k + c.len_utf8() == a.len() - 1;
                break;
            }
        }
    }
    out
}

/// git's long options per subcommand, from `git <cmd> --git-completion-helper-all` (git
/// 2.54.0), without `--` and `=`. git accepts any prefix of a long option that no other one
/// shares, so a guard must read `--force-w` as `--force-with-lease` (re-audit observation
/// 6777666763) and `--dele --forc` as `--delete --force`.
pub(crate) const PUSH_LONG: &[&str] = &[
    "all",
    "atomic",
    "branches",
    "delete",
    "dry-run",
    "exec",
    "follow-tags",
    "force",
    "force-if-includes",
    "force-with-lease",
    "ipv4",
    "ipv6",
    "mirror",
    "no-all",
    "no-atomic",
    "no-branches",
    "no-delete",
    "no-dry-run",
    "no-exec",
    "no-follow-tags",
    "no-force",
    "no-force-if-includes",
    "no-force-with-lease",
    "no-mirror",
    "no-porcelain",
    "no-progress",
    "no-prune",
    "no-push-option",
    "no-quiet",
    "no-receive-pack",
    "no-recurse-submodules",
    "no-repo",
    "no-set-upstream",
    "no-signed",
    "no-tags",
    "no-thin",
    "no-verbose",
    "no-verify",
    "porcelain",
    "progress",
    "prune",
    "push-option",
    "quiet",
    "receive-pack",
    "recurse-submodules",
    "repo",
    "set-upstream",
    "signed",
    "tags",
    "thin",
    "verbose",
    "verify",
];
pub(crate) const BRANCH_LONG: &[&str] = &[
    "abbrev",
    "all",
    "color",
    "column",
    "contains",
    "copy",
    "create-reflog",
    "delete",
    "edit-description",
    "force",
    "format",
    "ignore-case",
    "list",
    "merged",
    "move",
    "no-abbrev",
    "no-color",
    "no-column",
    "no-contains",
    "no-copy",
    "no-create-reflog",
    "no-delete",
    "no-edit-description",
    "no-force",
    "no-format",
    "no-ignore-case",
    "no-list",
    "no-merged",
    "no-move",
    "no-omit-empty",
    "no-points-at",
    "no-quiet",
    "no-recurse-submodules",
    "no-set-upstream",
    "no-set-upstream-to",
    "no-show-current",
    "no-sort",
    "no-track",
    "no-unset-upstream",
    "no-verbose",
    "omit-empty",
    "points-at",
    "quiet",
    "recurse-submodules",
    "remotes",
    "set-upstream",
    "set-upstream-to",
    "show-current",
    "sort",
    "track",
    "unset-upstream",
    "verbose",
    "with",
    "without",
];

pub(crate) const CLEAN_LONG: &[&str] = &[
    "dry-run",
    "exclude",
    "force",
    "interactive",
    "no-dry-run",
    "no-force",
    "no-interactive",
    "no-quiet",
    "quiet",
];
pub(crate) const RESET_LONG: &[&str] = &[
    "auto-advance",
    "hard",
    "intent-to-add",
    "inter-hunk-context",
    "keep",
    "merge",
    "mixed",
    "no-auto-advance",
    "no-intent-to-add",
    "no-patch",
    "no-pathspec-file-nul",
    "no-pathspec-from-file",
    "no-quiet",
    "no-recurse-submodules",
    "no-refresh",
    "patch",
    "pathspec-file-nul",
    "pathspec-from-file",
    "quiet",
    "recurse-submodules",
    "refresh",
    "soft",
    "unified",
];

/// Whether `arg` selects the long option `wanted` among a subcommand's long options `all`,
/// as git's parse-options reads it: the exact name, or a prefix that no other option in
/// `all` (negations included) shares. `--name=value` is read by its name.
pub(crate) fn long_opt(arg: &str, wanted: &str, all: &[&str]) -> bool {
    let Some(body) = arg.strip_prefix("--") else {
        return false;
    };
    let name = body.split('=').next().unwrap_or("");
    if name.is_empty() {
        return false;
    }
    if name == wanted {
        return true;
    }
    let mut hits = all.iter().filter(|o| o.starts_with(name));
    hits.next() == Some(&wanted) && hits.next().is_none()
}

/// Whether `arg` is a force option of `git push`: one git 2.54 reads as `--force`,
/// `--force-with-lease` or `--force-if-includes`, or any prefix of those from `--forc` on.
/// git 2.54 refuses `--forc` and `--force-` as ambiguous, but a git without
/// `--force-if-includes` (before 2.30) takes `--force-` for `--force-with-lease`, so the
/// guards treat the whole family as force (#2417 audit 6778052342, non-blocking note).
pub(crate) fn push_force_option(arg: &str) -> bool {
    const FAMILY: [&str; 3] = ["force", "force-with-lease", "force-if-includes"];
    if FAMILY.iter().any(|w| long_opt(arg, w, PUSH_LONG)) {
        return true;
    }
    let name = arg
        .strip_prefix("--")
        .and_then(|b| b.split('=').next())
        .unwrap_or("");
    name.len() >= 4 && name.starts_with("forc") && FAMILY.iter().any(|w| w.starts_with(name))
}

/// The words that are not options or option values: for `git push`, the remote then the
/// refspecs. `-o` and the long options below take the next word as their value.
fn push_positionals(rest: &[String]) -> Vec<&String> {
    const LONG_VALUE: &[&str] = &["push-option", "repo", "receive-pack", "exec"];
    let mut out = Vec::new();
    let mut skip_next = false;
    let mut options_done = false;
    for a in rest {
        if skip_next {
            skip_next = false;
            continue;
        }
        if !options_done && a == "--" {
            options_done = true;
            continue;
        }
        if !options_done && a.starts_with('-') && a.len() > 1 {
            skip_next = (!a.contains('=') && LONG_VALUE.iter().any(|w| long_opt(a, w, PUSH_LONG)))
                || (!a.starts_with("--")
                    && a.ends_with('o')
                    && a[1..].find('o') == Some(a.len() - 2));
            continue;
        }
        out.push(a);
    }
    out
}

fn branch_of_refspec(spec: &str) -> &str {
    let spec = spec.trim_start_matches('+');
    let dst = spec.rsplit(':').next().unwrap_or(spec);
    dst.strip_prefix("refs/heads/").unwrap_or(dst)
}

fn force_push_to_protected(rest: &[String]) -> bool {
    let force_flag =
        short_letters(rest, &['o']).contains(&'f') || rest.iter().any(|a| push_force_option(a));
    // The first positional word is the remote; the others are refspecs.
    push_positionals(rest).iter().skip(1).any(|spec| {
        let forced = force_flag || spec.starts_with('+');
        forced && PROTECTED_BRANCHES.contains(&branch_of_refspec(spec))
    })
}

fn force_delete_protected(rest: &[String]) -> bool {
    let letters = short_letters(rest, &['u']);
    let has = |w: &str| rest.iter().any(|a| long_opt(a, w, BRANCH_LONG));
    let force_delete = letters.contains(&'D')
        || ((letters.contains(&'d') || has("delete")) && (letters.contains(&'f') || has("force")));
    force_delete
        && rest
            .iter()
            .any(|a| PROTECTED_BRANCHES.contains(&a.as_str()))
}

fn skips_hooks(rest: &[String]) -> bool {
    let flag = bypass_flag();
    let mut take_value = false;
    for a in rest {
        if take_value {
            take_value = false;
            continue;
        }
        if a == "--" {
            return false;
        }
        if a.len() >= BYPASS_MIN_PREFIX && flag.starts_with(a.as_str()) {
            return true;
        }
        // A long value option, written out or as a prefix git accepts (`--mess` for
        // `--message`, pre-audit H5-B6). An ambiguous prefix makes git refuse the command, so
        // reading it as a value option can only allow a call that does not run.
        if a.len() >= 3
            && a.starts_with("--")
            && !a.contains('=')
            && COMMIT_LONG_VALUE_OPTS
                .iter()
                .any(|o| o.starts_with(a.as_str()))
        {
            take_value = true;
            continue;
        }
        if a.starts_with("--") || !a.starts_with('-') || a.len() < 2 {
            continue;
        }
        // A short-option cluster: `-n` counts unless a value-taking option came first.
        for (k, c) in a[1..].char_indices() {
            if c == 'n' {
                return true;
            }
            if COMMIT_ATTACHED_OPTS.contains(&c) {
                break;
            }
            if COMMIT_VALUE_OPTS.contains(&c) {
                // `-m` alone takes the next word; `-mtext` holds its value inline.
                take_value = k + c.len_utf8() == a.len() - 1;
                break;
            }
        }
    }
    false
}

/// True when git's own arguments (before the subcommand) set `core.hooksPath` — git reads the
/// key case-insensitively, and pointing it elsewhere skips the repository's hooks (measured
/// with git 2.54.0). Covers `-c key=value`, the attached `-ckey=value` and `--config-env`.
fn overrides_hooks_path(global: &[String]) -> bool {
    let sets = |s: &str| {
        s.split('=')
            .next()
            .is_some_and(|k| k.eq_ignore_ascii_case("core.hooksPath"))
    };
    global.iter().enumerate().any(|(i, a)| {
        (a == "-c" && global.get(i + 1).is_some_and(|v| sets(v)))
            || a.strip_prefix("-c")
                .is_some_and(|v| !v.is_empty() && sets(v))
            || a.strip_prefix("--config-env=").is_some_and(sets)
    })
}

/// The reason a file write is refused, or `None`. `cwd` resolves relative paths.
/// The paths a file-writing tool names and the text it would write; `None` for other tools.
fn write_targets(tool: &str, tool_input: &Value) -> Option<(Vec<String>, Vec<String>)> {
    let mut paths: Vec<String> = Vec::new();
    let mut contents: Vec<String> = Vec::new();
    match tool {
        "Write" | "Edit" | "MultiEdit" | "NotebookEdit" => {
            for key in ["file_path", "notebook_path"] {
                if let Some(p) = tool_input[key].as_str() {
                    paths.push(p.to_string());
                }
            }
            for key in ["content", "new_string", "new_source"] {
                if let Some(c) = tool_input[key].as_str() {
                    contents.push(c.to_string());
                }
            }
            for edit in tool_input["edits"].as_array().into_iter().flatten() {
                if let Some(c) = edit["new_string"].as_str() {
                    contents.push(c.to_string());
                }
            }
        }
        "apply_patch" => {
            let patch = tool_input["command"].as_str().unwrap_or("");
            for line in patch.lines() {
                for head in [
                    "*** Add File: ",
                    "*** Update File: ",
                    "*** Delete File: ",
                    "*** Move to: ",
                ] {
                    if let Some(p) = line.strip_prefix(head) {
                        paths.push(p.trim().to_string());
                    }
                }
                if let Some(added) = line.strip_prefix('+') {
                    contents.push(added.to_string());
                }
            }
            if !contents.is_empty() {
                contents = vec![contents.join("\n")];
            }
        }
        _ => return None,
    }
    Some((paths, contents))
}

pub fn write_guard(tool: &str, tool_input: &Value, cwd: &Path) -> Option<&'static str> {
    let (paths, contents) = write_targets(tool, tool_input)?;
    if paths.iter().any(|p| is_protected_path(Path::new(p), cwd)) {
        return Some(PROTECTED_PATH);
    }
    if contents.iter().any(|c| carries_secret(c)) {
        return Some(SECRET_CONTENT);
    }
    None
}

const OUTSIDE_PROJECT: &str = "MoAI tool-preflight: a write outside the project is refused; add the folder to `allow` under [hooks.outside_project_writes] in ~/.moai-cli/config to permit it.";
const PRIMARY_CHECKOUT: &str = "MoAI tool-preflight: this primary checkout is shared with linked worktrees; switching, resetting, stashing, rebasing, merging or changing its branches is refused (use a worktree; add the exact command to `allow_commands` under [hooks.primary_checkout] in ~/.moai-cli/config to permit it).";
const PROTECTED_BRANCH_COMMIT: &str = "MoAI tool-preflight: this primary checkout is on a branch listed in `deny_commits_on` under [hooks.primary_checkout] in ~/.moai-cli/config; a commit, revert or cherry-pick there is refused (commit in a worktree on its own branch).";

/// The folder holding the nearest `.git` (a folder or a file) at or above `dir`, resolved.
/// Found from the file system only: no git process.
fn work_tree_top(dir: &Path) -> Option<PathBuf> {
    let mut d = crate::hook_config::canonical(dir);
    loop {
        if std::fs::symlink_metadata(d.join(".git")).is_ok() {
            return Some(d);
        }
        if !d.pop() {
            return None;
        }
    }
}

/// B9 (opt-in, decision 6779829826): a file-writing tool's path must be inside the project
/// (the work tree holding the session's directory) or under an `allow` folder. Paths are
/// judged after symlinks and `..` are resolved, so a link or a `..` out of the project counts
/// as outside. Outside any git work tree only `allow` remains.
pub fn outside_project_guard(
    tool: &str,
    tool_input: &Value,
    cwd: &Path,
    rule: &crate::hook_config::OutsideRule,
) -> Option<&'static str> {
    let (paths, _) = write_targets(tool, tool_input)?;
    let project = work_tree_top(cwd);
    let outside = |p: &String| {
        let path = Path::new(p);
        let absolute = if path.is_absolute() {
            path.to_path_buf()
        } else {
            cwd.join(path)
        };
        let resolved = crate::hook_config::canonical(&absolute);
        let inside = project
            .as_ref()
            .is_some_and(|top| resolved.starts_with(top))
            || rule
                .allow
                .iter()
                .any(|a| resolved.starts_with(crate::hook_config::canonical(a)));
        !inside
    };
    paths.iter().any(outside).then_some(OUTSIDE_PROJECT)
}

/// Whether a git subcommand changes which commit or branch the work tree holds, or its
/// uncommitted state (B3): `switch`; `checkout` naming a branch or commit (not a path restore
/// after `--`, not a bare `checkout`); every `reset`; `stash` other than `list`/`show`;
/// every `rebase`; every `merge` (#2878 E2); a `git branch` that changes branches.
fn moves_the_checkout(sub: &crate::claim_verify::GitSub) -> bool {
    let rest = &sub.rest;
    match sub.verb.as_str() {
        "switch" | "rebase" | "reset" | "merge" => true,
        "checkout" => !rest.iter().any(|a| a == "--") && !rest.is_empty(),
        "stash" => !matches!(rest.first().map(String::as_str), Some("list" | "show")),
        "branch" => changes_branches(rest),
        _ => false,
    }
}

/// `git branch` arguments that create, delete, rename, copy, force or re-track a branch
/// (#2878 E2, the Go generation's branch guard). Listing forms — no arguments, `--list`,
/// `-a`, `-r`, `-v`, `--contains <c>`, `--merged <c>`, `--show-current` … — change nothing.
fn changes_branches(args: &[String]) -> bool {
    const CHANGING_LONG: &[&str] = &[
        "--delete",
        "--move",
        "--copy",
        "--force",
        "--edit-description",
        "--set-upstream-to",
        "--unset-upstream",
        "--track",
        "--no-track",
        "--create-reflog",
        "--recurse-submodules",
    ];
    // Listing options that take the next word as their value.
    const LIST_WITH_VALUE: &[&str] = &[
        "--contains",
        "--no-contains",
        "--merged",
        "--no-merged",
        "--points-at",
        "--format",
        "--sort",
        "--column",
    ];
    const LIST_ONLY: &[&str] = &[
        "-l",
        "--list",
        "-a",
        "--all",
        "-r",
        "--remotes",
        "-v",
        "-vv",
        "--verbose",
        "--show-current",
        "-i",
        "--ignore-case",
        "--no-column",
        "--color",
        "--no-color",
        "--abbrev",
        "--no-abbrev",
        "-q",
        "--quiet",
    ];
    let mut listing = false;
    let mut positional = false;
    let mut it = args.iter();
    while let Some(a) = it.next() {
        let flag = a.split('=').next().unwrap_or(a);
        if CHANGING_LONG.contains(&flag) || flag.starts_with("--set-upstream") {
            return true;
        }
        if LIST_WITH_VALUE.contains(&flag) {
            listing = true;
            if !a.contains('=') {
                it.next();
            }
            continue;
        }
        if LIST_ONLY.contains(&flag)
            || flag.starts_with("--color=")
            || flag.starts_with("--abbrev=")
        {
            if !matches!(flag, "-q" | "--quiet") {
                listing = true;
            }
            continue;
        }
        if let Some(cluster) = a
            .strip_prefix('-')
            .filter(|c| !c.starts_with('-') && !c.is_empty())
        {
            // A short-option cluster such as `-D`, `-df`, `-Mf`: any changing letter changes.
            if cluster.chars().any(|c| "dDmMcCfut".contains(c)) {
                return true;
            }
            if cluster.chars().any(|c| "lavr".contains(c)) {
                listing = true;
            }
            continue;
        }
        if a.starts_with("--") {
            continue;
        }
        positional = true;
    }
    // `git branch <name> [<start>]` creates a branch; with a listing flag the words are
    // patterns.
    positional && !listing
}

/// The branch a primary checkout's HEAD names (`ref: refs/heads/<b>` in `.git/HEAD`), read
/// from the file system; `None` when detached, missing or unreadable.
fn head_branch(top: &Path) -> Option<String> {
    let head = std::fs::read_to_string(top.join(".git").join("HEAD")).ok()?;
    head.trim()
        .strip_prefix("ref: refs/heads/")
        .map(str::to_string)
}

/// Whether a git subcommand makes a new commit on the current branch (#2878 E2): `commit`,
/// `revert` and `cherry-pick` (their `--abort`/`--quit`/`--skip` end an operation without one).
fn makes_a_commit(sub: &crate::claim_verify::GitSub) -> bool {
    matches!(sub.verb.as_str(), "commit" | "revert" | "cherry-pick")
        && !sub
            .rest
            .iter()
            .any(|a| matches!(a.as_str(), "--abort" | "--quit" | "--skip"))
}

/// The directory a git subcommand runs in: `cwd`, changed by each `-C <dir>` before the verb.
fn git_dir_of(global: &[String], cwd: &Path) -> PathBuf {
    let mut dir = cwd.to_path_buf();
    let mut it = global.iter();
    while let Some(a) = it.next() {
        if a == "-C" {
            if let Some(d) = it.next() {
                dir = dir.join(d);
            }
        }
    }
    dir
}

/// A primary worktree (its `.git` is a folder) that has at least one linked worktree.
fn shared_primary(top: &Path) -> bool {
    let git = top.join(".git");
    std::fs::symlink_metadata(&git).is_ok_and(|m| m.is_dir())
        && std::fs::read_dir(git.join("worktrees"))
            .is_ok_and(|mut entries| entries.next().is_some())
}

/// B3 (opt-in, decision 6779829826): in a primary checkout listed in `repos` that has linked
/// worktrees, a git command that switches, resets, stashes, rebases, merges or changes its
/// branches is refused unless the whole command line is in `allow_commands`. In a listed
/// primary checkout whose HEAD is a `deny_commits_on` branch, a commit, revert or cherry-pick
/// is refused too, with or without linked worktrees (#2878 E2). File writes are out of scope.
pub fn primary_checkout_guard(
    command: &str,
    cwd: &Path,
    rule: &crate::hook_config::PrimaryRule,
) -> Option<&'static str> {
    if rule.allow_commands.iter().any(|c| c == command.trim()) {
        return None;
    }
    for (name, args) in crate::claim_verify::shell_commands(command) {
        if name != "git" {
            continue;
        }
        for sub in crate::claim_verify::git_subcommands(&args) {
            let moves = moves_the_checkout(&sub);
            let commits = makes_a_commit(&sub);
            if !moves && !commits {
                continue;
            }
            let Some(top) = work_tree_top(&git_dir_of(&sub.global, cwd)) else {
                continue;
            };
            if !rule.repos.contains(&top) {
                continue;
            }
            if moves && shared_primary(&top) {
                return Some(PRIMARY_CHECKOUT);
            }
            let primary = std::fs::symlink_metadata(top.join(".git")).is_ok_and(|m| m.is_dir());
            if commits
                && primary
                && head_branch(&top).is_some_and(|b| rule.deny_commits_on.contains(&b))
            {
                return Some(PROTECTED_BRANCH_COMMIT);
            }
        }
    }
    None
}

/// Author and committer emails this repository's tests use (#2878 E3). A commit under one of
/// them in a real checkout means a test's identity leaked into git config or the environment.
pub const TEST_IDENTITY_EMAILS: &[&str] = &[
    "t@example.invalid",
    "fixture@example.invalid",
    "test@example.invalid",
    "f@example.invalid",
    "bench@example.invalid",
    "t@example.com",
    "t@t",
];

/// The git verbs that write a commit with an author and a committer (#2878 E3).
const COMMITTING_VERBS: &[&str] = &[
    "commit",
    "merge",
    "cherry-pick",
    "revert",
    "rebase",
    "am",
    "commit-tree",
    "pull",
];

/// The email inside `Name <email>` (or a bare `email`), trimmed and lowercased.
fn email_of(ident: &str) -> Option<String> {
    let inner = match (ident.find('<'), ident.rfind('>')) {
        (Some(a), Some(b)) if a < b => &ident[a + 1..b],
        _ => ident,
    };
    let e = inner.trim().to_ascii_lowercase();
    e.contains('@').then_some(e)
}

/// `git var <name>` in `dir` with the inherited git environment scrubbed, so the probe reads
/// the identity the repository's configuration gives; `None` on any failure.
fn git_ident(dir: &Path, name: &str) -> Option<String> {
    let mut cmd = std::process::Command::new("git");
    crate::identity::scrub_git_env(&mut cmd);
    let out = cmd
        .arg("-C")
        .arg(dir)
        .args(["var", name])
        .stdin(std::process::Stdio::null())
        .stderr(std::process::Stdio::null())
        .output()
        .ok()?;
    out.status
        .success()
        .then(|| String::from_utf8_lossy(&out.stdout).trim().to_string())
}

/// #2878 E3 (opt-in, the Go generation's commit-identity guard): a git command that writes a
/// commit is refused when its author or committer email is a test address
/// ([`TEST_IDENTITY_EMAILS`]) or one in `deny_emails`. The email comes from the command itself
/// when it sets one (`--author`, `-c user.email=…`, `GIT_AUTHOR_EMAIL=`/`GIT_COMMITTER_EMAIL=`
/// before `git`), otherwise from `git var GIT_AUTHOR_IDENT`/`GIT_COMMITTER_IDENT` in the
/// folder the command runs in. Anything it cannot read lets the command through (fail-open,
/// like the Go guard); commands inside `sh -c` are not opened.
pub fn commit_identity_guard(
    command: &str,
    cwd: &Path,
    rule: &crate::hook_config::IdentityRule,
) -> Option<String> {
    let denied = |email: &str| {
        TEST_IDENTITY_EMAILS.contains(&email) || rule.deny_emails.iter().any(|d| d == email)
    };
    for (env, name, args) in crate::claim_verify::shell_commands_with_env(command) {
        if name != "git" {
            continue;
        }
        for sub in crate::claim_verify::git_subcommands(&args) {
            if !COMMITTING_VERBS.contains(&sub.verb.as_str()) {
                continue;
            }
            let from_env = |key: &str| {
                env.iter()
                    .find_map(|w| w.strip_prefix(&format!("{key}=")))
                    .and_then(email_of)
            };
            let mut user_email = None;
            let mut g = sub.global.iter();
            while let Some(a) = g.next() {
                let kv = if a == "-c" {
                    g.next().map(String::as_str)
                } else {
                    a.strip_prefix("-c")
                };
                // Config keys ignore case (`USER.EMAIL` is `user.email`); the value is kept.
                if let Some((key, v)) = kv.and_then(|kv| kv.split_once('=')) {
                    if key.eq_ignore_ascii_case("user.email") {
                        user_email = email_of(v);
                    }
                }
            }
            let mut author_flag = None;
            let mut r = sub.rest.iter();
            while let Some(a) = r.next() {
                if a == "--author" {
                    author_flag = r.next().and_then(|v| email_of(v));
                } else if let Some(v) = a.strip_prefix("--author=") {
                    author_flag = email_of(v);
                }
            }
            let dir = git_dir_of(&sub.global, cwd);
            let author = author_flag
                .or_else(|| from_env("GIT_AUTHOR_EMAIL"))
                .or_else(|| user_email.clone())
                .or_else(|| git_ident(&dir, "GIT_AUTHOR_IDENT").and_then(|i| email_of(&i)));
            let committer = from_env("GIT_COMMITTER_EMAIL")
                .or_else(|| user_email.clone())
                .or_else(|| git_ident(&dir, "GIT_COMMITTER_IDENT").and_then(|i| email_of(&i)));
            for (role, email) in [("author", author), ("committer", committer)] {
                if let Some(e) = email.filter(|e| denied(e)) {
                    return Some(format!(
                        "MoAI tool-preflight: the {role} identity <{e}> is a test address; the commit is refused. Fix the identity git would use (git config user.email, or GIT_AUTHOR_EMAIL/GIT_COMMITTER_EMAIL) — a test's identity has leaked into this checkout ([hooks.commit_identity] in ~/.moai-cli/config)."
                    ));
                }
            }
        }
    }
    None
}

/// The path as written and as resolved through existing symlinks; either one in a protected
/// place is enough.
fn is_protected_path(path: &Path, cwd: &Path) -> bool {
    let absolute = if path.is_absolute() {
        path.to_path_buf()
    } else {
        cwd.join(path)
    };
    is_protected_lexical(&absolute) || is_protected_lexical(&resolve_existing(&absolute))
}

/// Resolves the longest existing prefix through symlinks and appends the rest unchanged.
fn resolve_existing(path: &Path) -> PathBuf {
    let mut existing = path.to_path_buf();
    let mut tail = Vec::new();
    while !existing.exists() {
        match (
            existing.file_name().map(|n| n.to_os_string()),
            existing.parent(),
        ) {
            (Some(name), Some(parent)) => {
                tail.push(name);
                existing = parent.to_path_buf();
            }
            _ => return path.to_path_buf(),
        }
    }
    let mut out = std::fs::canonicalize(&existing).unwrap_or(existing);
    for name in tail.into_iter().rev() {
        out.push(name);
    }
    out
}

fn is_protected_lexical(path: &Path) -> bool {
    let parts: Vec<String> = path
        .components()
        .filter_map(|c| match c {
            Component::Normal(s) => Some(s.to_string_lossy().into_owned()),
            _ => None,
        })
        .collect();
    let in_dir = |name: &str| parts.iter().any(|p| p == name);
    if in_dir(".git") || CREDENTIAL_DIRS.iter().any(|d| in_dir(d)) {
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

fn carries_secret(content: &str) -> bool {
    // Assembled so this file itself never holds the whole marker.
    let begin = concat!("-----", "BEGIN");
    let private = concat!("PRIVATE", " KEY-----");
    if content.contains(begin) && content.contains(private) {
        return true;
    }
    has_token(content, "ghp_", 36, false) || has_token(content, "github_pat_", 22, true)
}

/// True when `prefix` is followed by at least `min` token characters.
fn has_token(content: &str, prefix: &str, min: usize, underscore: bool) -> bool {
    content.match_indices(prefix).any(|(at, _)| {
        content[at + prefix.len()..]
            .chars()
            .take_while(|c| c.is_ascii_alphanumeric() || (underscore && *c == '_'))
            .count()
            >= min
    })
}

#[cfg(test)]
mod tests {
    //! Literal cases only; nothing here runs a command. Secret-shaped and hook-bypass
    //! fixtures are assembled at run time so no source line holds them whole.
    use super::*;
    use serde_json::json;

    fn bypass() -> String {
        format!("--no-{}", "verify")
    }

    #[test]
    fn rm_refuses_only_irreversible_targets() {
        for cmd in [
            "rm -rf /",
            "rm -rf /*",
            "rm -rf ~",
            "rm -rf ~/",
            "rm -fr $HOME",
            "rm -rf ${HOME}/*",
            "rm -rf *",
            "rm -rf .",
            "rm -rf ..",
            "rm -rf /usr",
            "rm -rf /Users/",
            "rm -rf .git",
            "rm -rf repo/.git/",
            "rm -rf -- /",
            "sudo rm -rf /",
            "cd x && rm -rf ~",
            "bash -c 'rm -rf /'",
        ] {
            assert_eq!(bash_guard(cmd), Some(RM_TARGET), "{cmd}");
        }
        for cmd in [
            "rm -rf build",
            "rm -rf ./build",
            "rm -rf node_modules",
            "rm -rf /tmp/x",
            "rm -f .gitignore",
            "rm -rf src/.github",
            "echo rm -rf /",
            "grep -n 'rm -rf /' notes.md",
            "rm -- -weird-name",
        ] {
            assert_eq!(bash_guard(cmd), None, "{cmd}");
        }
    }

    #[test]
    fn force_push_and_delete_of_main_or_master_are_refused() {
        for cmd in [
            "git push --force origin main",
            "git push -f origin master",
            "git push origin +main",
            "git push --force-with-lease origin HEAD:main",
            "git push -f origin refs/heads/master",
            "git -C repo push --force origin main",
        ] {
            assert_eq!(bash_guard(cmd), Some(FORCE_PUSH), "{cmd}");
        }
        for cmd in ["git branch -D main", "git branch --delete --force master"] {
            assert_eq!(bash_guard(cmd), Some(BRANCH_DELETE), "{cmd}");
        }
        for cmd in [
            "git push origin main",
            "git push --force origin feature",
            "git push -f origin issue/1-x",
            "git push --force origin v2026.11.0",
            // The first word is the remote, even when it is spelled like a protected branch.
            "git push --force master feature",
            "git branch -d main",
            "git branch -D feature",
            "git log --force main",
        ] {
            assert_eq!(bash_guard(cmd), None, "{cmd}");
        }
    }

    /// Re-audit observation 6777666763: git takes a unique prefix of a long option (measured
    /// with git 2.54.0: `push --force-w` overwrites, `branch --dele --forc` deletes, `--forc`
    /// and `--force-` are ambiguous for push, `branch --f` is ambiguous with `--format`).
    #[test]
    fn long_option_prefixes_are_read_like_git_reads_them() {
        for cmd in [
            "git push --force-w origin main",
            "git push --force-wi origin main",
            "git push --force-w=main:abc origin main",
            "git push --force-w origin feature:main",
            "git push --force-i origin main",
            "git push --forc origin main",
            "git push --force- origin main",
            "git push --force-wit=main:abc origin main",
        ] {
            assert_eq!(bash_guard(cmd), Some(FORCE_PUSH), "{cmd}");
        }
        for cmd in [
            "git branch --dele --forc main",
            "git branch --d --force master",
            "git branch --delete --forc main",
        ] {
            assert_eq!(bash_guard(cmd), Some(BRANCH_DELETE), "{cmd}");
        }
        for cmd in [
            // Ambiguous for git 2.54 (it refuses them), not for an older git (see
            // `push_force_option`).
            "git branch --d --f main",
            "git branch --delete --fo main",
            // Another option's prefix or a negation.
            "git push --fo origin main",
            "git push --no-force origin main",
            "git branch --d main",
        ] {
            assert_eq!(bash_guard(cmd), None, "{cmd}");
        }
        // A value option's prefix takes the next word: in `--push-o x main`, `x` is the
        // option's value and `main` is the remote, so no refspec names a protected branch.
        assert_eq!(bash_guard("git push --force --push-o x main"), None);
        assert!(long_opt("--force-w", "force-with-lease", PUSH_LONG));
        assert!(!long_opt("--force", "force-with-lease", PUSH_LONG));
        assert!(!long_opt("-f", "force", PUSH_LONG));
        assert!(!long_opt("--", "force", PUSH_LONG));
    }

    /// The option lists are git's own (`--git-completion-helper-all`); when the installed git
    /// answers, its list must equal ours, so a new git option that changes a prefix shows here.
    #[test]
    fn option_lists_match_the_installed_git() {
        // The lists were measured with git 2.54; another version may differ in options that
        // do not matter here, so only that version is compared.
        // Every git spawn runs with a scrubbed environment (#2337, tests/git_env_scrub.rs).
        let git = |args: &[&str]| {
            let mut cmd = std::process::Command::new("git");
            cmd.args(args);
            crate::identity::scrub_git_env(&mut cmd);
            cmd.output()
        };
        let Ok(version) = git(&["--version"]) else {
            eprintln!("skipped: git not found");
            return;
        };
        if !String::from_utf8_lossy(&version.stdout).starts_with("git version 2.54.") {
            eprintln!("skipped: the lists were measured with git 2.54");
            return;
        }
        for (cmd, ours) in [
            ("push", PUSH_LONG),
            ("branch", BRANCH_LONG),
            ("clean", CLEAN_LONG),
            ("reset", RESET_LONG),
        ] {
            let Ok(out) = git(&[cmd, "--git-completion-helper-all"]) else {
                eprintln!("skipped: git not found");
                return;
            };
            let text = String::from_utf8_lossy(&out.stdout);
            let mut theirs: Vec<&str> = text
                .split_whitespace()
                .filter_map(|w| w.strip_prefix("--"))
                .map(|w| w.trim_end_matches('='))
                .filter(|w| !w.is_empty())
                .collect();
            theirs.sort_unstable();
            theirs.dedup();
            let mut ours = ours.to_vec();
            ours.sort_unstable();
            assert_eq!(theirs, ours, "git {cmd}");
        }
    }

    #[test]
    fn devices_fork_bombs_and_database_drops_are_refused() {
        for cmd in [
            "mkfs.ext4 /dev/sdb1",
            "sudo mkfs -t ext4 /dev/sdc",
            "dd if=img of=/dev/disk2 bs=1m",
        ] {
            assert_eq!(bash_guard(cmd), Some(DEVICE), "{cmd}");
        }
        assert_eq!(bash_guard(":(){ :|:& };:"), Some(FORK_BOMB));
        for cmd in [
            "psql -c 'DROP DATABASE app'",
            "mysql -e \"drop schema s\"",
            "sqlite3 db.sqlite 'TRUNCATE TABLE t'",
        ] {
            assert_eq!(bash_guard(cmd), Some(DATABASE), "{cmd}");
        }
        for cmd in [
            "dd if=/dev/zero of=img bs=1m count=1",
            "dd if=x of=/dev/null",
            "psql -c 'select 1'",
            "echo 'DROP DATABASE app'",
            "cat mkfs-notes.txt",
        ] {
            assert_eq!(bash_guard(cmd), None, "{cmd}");
        }
    }

    #[test]
    fn a_commit_that_skips_hooks_is_refused_only_as_a_commit_option() {
        let flag = bypass();
        for cmd in [
            format!("git commit {flag} -m x"),
            "git commit -n -m x".to_string(),
            "git commit -anm x".to_string(),
            "git commit -nm x".to_string(),
            format!("git -C repo commit {flag}"),
            // Unique abbreviations git accepts (9 and 10 characters, measured with 2.54.0).
            format!("git commit {} -m x", &flag[..9]),
            format!("git commit {} -m x", &flag[..10]),
            // A hooks directory override on git's own arguments, any key case.
            "git -c core.hooksPath=/dev/null commit -m x".to_string(),
            "git -c core.hookspath=/tmp/none commit -m x".to_string(),
            "git -ccore.hooksPath=x commit -m x".to_string(),
            "git --config-env=core.hooksPath=HOOKS commit -m x".to_string(),
        ] {
            assert_eq!(bash_guard(&cmd), Some(HOOK_BYPASS), "{cmd}");
        }
        for cmd in [
            "git commit -m x".to_string(),
            // `n` here is the message text, not an option.
            "git commit -mn".to_string(),
            "git commit -m -n".to_string(),
            "git commit -am x".to_string(),
            format!("git log {flag}"),
            "git push -n origin main".to_string(),
            format!("git commit -m 'skip {flag}'"),
            // Eight characters is ambiguous for git and runs the hook (measured).
            format!("git commit {} -m x", &flag[..8]),
            // `-u` and `-S` hold an attached value: `n` there is not an option (measured:
            // `-uno` and `-unormal` run the hook).
            "git commit -uno -m x".to_string(),
            "git commit -unormal -m x".to_string(),
            "git commit -Snkey -m x".to_string(),
            // A long option's separate value is not an option either.
            "git commit --message -n".to_string(),
            "git commit --author -n -m x".to_string(),
            // Other configuration, or the override on a command that is not a commit.
            "git -c user.name=n commit -m x".to_string(),
            "git -c core.hooksPath=/dev/null log".to_string(),
            "git -c core.hooksPathX=1 commit -m x".to_string(),
        ] {
            assert_eq!(bash_guard(&cmd), None, "{cmd}");
        }
    }

    /// Pre-audit FAIL (#2391 stage 5, pr-4) B2–B7: the literal inputs of the report, each with
    /// its control. Nothing here runs a command.
    #[test]
    fn pre_audit_literals_are_judged_as_git_and_the_shell_read_them() {
        let flag = bypass();
        for (cmd, want) in [
            // B2: `eval` runs its joined arguments as a script.
            ("eval 'rm -rf /'".to_string(), Some(RM_TARGET)),
            ("eval rm -rf /".to_string(), Some(RM_TARGET)),
            ("sh -c 'rm -rf /'".to_string(), Some(RM_TARGET)),
            ("echo 'rm -rf /'".to_string(), None),
            // B3: short options aggregate.
            ("git push -vf origin main".to_string(), Some(FORCE_PUSH)),
            ("git push -fv origin main".to_string(), Some(FORCE_PUSH)),
            ("git branch -df main".to_string(), Some(BRANCH_DELETE)),
            ("git branch -fd main".to_string(), Some(BRANCH_DELETE)),
            ("git push -v origin main".to_string(), None),
            ("git branch -dv main".to_string(), None),
            // `-o` takes a value: `f` inside it is not the force option, and its value is not
            // the remote.
            ("git push -o f origin feature".to_string(), None),
            ("git push -of origin main".to_string(), None),
            (
                "git push -o ci.skip -f origin main".to_string(),
                Some(FORCE_PUSH),
            ),
            // An option value is not the remote: here `main` is the remote, with no refspec.
            ("git push -f --push-option x main".to_string(), None),
            ("git push -f -o x main".to_string(), None),
            // `-u` takes a value for `git branch`.
            ("git branch -uf origin/x -d main".to_string(), None),
            // B4: quoted punctuation is data.
            ("printf '%s' ':(){'".to_string(), None),
            ("echo \":(){ :|:& };:\"".to_string(), None),
            (":(){ :|:& };:".to_string(), Some(FORK_BOMB)),
            // Re-audit: a quoted script that a shell or `eval` runs is code, not data.
            ("sh -c ':(){ :|:& };:'".to_string(), Some(FORK_BOMB)),
            ("bash -c \":(){ :|:& };:\"".to_string(), Some(FORK_BOMB)),
            ("eval ':(){ :|:& };:'".to_string(), Some(FORK_BOMB)),
            ("eval \":(){ :|:& };:\"".to_string(), Some(FORK_BOMB)),
            (
                "cd x && sh -c ': () { :|:& };:'".to_string(),
                Some(FORK_BOMB),
            ),
            // ...while data stays data, also inside a script that runs.
            ("git commit -m ':(){ :|:& };:'".to_string(), None),
            ("sh -c 'echo \":(){\"'".to_string(), None),
            ("sh -c \"printf '%s' ':(){'\"".to_string(), None),
            // B5: a statement inside an SQL string literal is data.
            ("psql -c \"SELECT 'DROP DATABASE app';\"".to_string(), None),
            (
                "psql -c \"SELECT 'it''s'; DROP DATABASE app;\"".to_string(),
                Some(DATABASE),
            ),
            ("psql -c 'DROP DATABASE app'".to_string(), Some(DATABASE)),
            // B6: a unique prefix of a long value option takes the next word.
            ("git commit --mess -n".to_string(), None),
            ("git commit --auth -n -m x".to_string(), None),
            ("git commit --mess x -n".to_string(), Some(HOOK_BYPASS)),
            // B7: an alias defined on the same line, including the words it carries.
            (
                "git -c alias.ci=commit ci -n -m x".to_string(),
                Some(HOOK_BYPASS),
            ),
            (
                "git -calias.ci=commit ci -n -m x".to_string(),
                Some(HOOK_BYPASS),
            ),
            (
                format!("git -c 'alias.ci=commit {flag}' ci -m x"),
                Some(HOOK_BYPASS),
            ),
            (
                "git -c 'alias.pf=push -f' pf origin main".to_string(),
                Some(FORCE_PUSH),
            ),
            (
                "git -c 'alias.nuke=!git branch -D' nuke main".to_string(),
                Some(BRANCH_DELETE),
            ),
            ("git -c alias.ci=commit ci -m x".to_string(), None),
            ("git -c alias.lg=log lg -n 3".to_string(), None),
        ] {
            assert_eq!(bash_guard(&cmd), want, "{cmd}");
        }
    }

    /// Decision B2 = A: the confirmation class, with its look-alikes. Strings only.
    #[test]
    fn the_confirmation_class_is_read_like_git_reads_it() {
        for (cmd, want) in [
            ("git reset --hard", Some(CONFIRM_RESET)),
            ("git reset --hard HEAD~1", Some(CONFIRM_RESET)),
            ("git reset --ha", Some(CONFIRM_RESET)),
            ("git -c 'alias.rh=reset --hard' rh", Some(CONFIRM_RESET)),
            ("git reset --soft HEAD~1", None),
            ("git reset HEAD a.txt", None),
            // git takes `--h` for `--hard` (measured, git 2.54.0).
            ("git reset --h", Some(CONFIRM_RESET)),
            ("git reset --so HEAD~1", None),
            ("git reset -- --hard", None),
            ("git clean -fd", Some(CONFIRM_CLEAN)),
            ("git clean -fdx", Some(CONFIRM_CLEAN)),
            ("git clean --force", Some(CONFIRM_CLEAN)),
            ("git clean -f -e f", Some(CONFIRM_CLEAN)),
            ("git clean --f", Some(CONFIRM_CLEAN)),
            ("git clean --fo -d", Some(CONFIRM_CLEAN)),
            ("git clean --forc", Some(CONFIRM_CLEAN)),
            ("git clean --f --dry", None),
            ("git clean --h", None),
            ("git clean -n", None),
            ("git clean -fn", None),
            ("git clean -fd --dry-run", None),
            ("git clean -ef", None),
            ("git push --force origin feature", Some(CONFIRM_FORCE_PUSH)),
            ("git push -f", Some(CONFIRM_FORCE_PUSH)),
            ("git push origin +feature", Some(CONFIRM_FORCE_PUSH)),
            ("git push --force-with-lease", Some(CONFIRM_FORCE_PUSH)),
            (
                "git push --force-w origin feature",
                Some(CONFIRM_FORCE_PUSH),
            ),
            (
                "git push --force-wi=feature:abc origin feature",
                Some(CONFIRM_FORCE_PUSH),
            ),
            (
                "git push --force-i origin feature",
                Some(CONFIRM_FORCE_PUSH),
            ),
            // Ambiguous for git 2.54, the force family for an older git (`push_force_option`).
            ("git push --forc origin feature", Some(CONFIRM_FORCE_PUSH)),
            ("git push --force- origin feature", Some(CONFIRM_FORCE_PUSH)),
            ("git push origin feature", None),
            ("git push -o f origin feature", None),
            ("git log --hard", None),
            ("echo git reset --hard", None),
        ] {
            assert_eq!(bash_confirm(cmd), want, "{cmd}");
        }
        // Protected branches stay a deny, not a question.
        assert_eq!(bash_guard("git push --force origin main"), Some(FORCE_PUSH));
    }

    /// A temporary folder holding `repo/` (with a `.git` folder), `allowed/` and `outside/`.
    fn layout() -> (tempfile::TempDir, PathBuf) {
        let dir = tempfile::TempDir::new().unwrap();
        let root = std::fs::canonicalize(dir.path()).unwrap();
        for d in ["repo/.git", "repo/src", "allowed", "outside", "nogit"] {
            std::fs::create_dir_all(root.join(d)).unwrap();
        }
        (dir, root)
    }

    fn write(path: &str) -> Value {
        json!({"file_path": path, "content": "x"})
    }

    /// B9 (stage 7): inside the project or under `allow` passes; anything that resolves
    /// outside (absolute, `..`, a `..` through a folder that does not exist, a symlink) is
    /// refused. Outside any work tree only `allow` remains.
    #[test]
    fn writes_outside_the_project_are_refused_when_turned_on() {
        let (_d, root) = layout();
        let repo = root.join("repo");
        let rule = crate::hook_config::OutsideRule {
            allow: vec![root.join("allowed")],
        };
        let out = root.join("outside/x").to_string_lossy().into_owned();
        let allowed = root.join("allowed/x").to_string_lossy().into_owned();
        #[cfg(unix)]
        std::os::unix::fs::symlink(root.join("outside"), repo.join("link")).unwrap();
        let mut cases = vec![
            ("src/a.rs".to_string(), None),
            (repo.join("src/b.rs").to_string_lossy().into_owned(), None),
            ("new/dir/file".to_string(), None),
            (allowed.clone(), None),
            (out.clone(), Some(OUTSIDE_PROJECT)),
            ("../outside/x".to_string(), Some(OUTSIDE_PROJECT)),
            ("nothere/../../outside/x".to_string(), Some(OUTSIDE_PROJECT)),
            ("../repo-twin/x".to_string(), Some(OUTSIDE_PROJECT)),
        ];
        #[cfg(unix)]
        cases.push(("link/x".to_string(), Some(OUTSIDE_PROJECT)));
        for (path, want) in cases {
            assert_eq!(
                outside_project_guard("Write", &write(&path), &repo, &rule),
                want,
                "{path}"
            );
        }
        // Every file-writing tool, the patch format included; other tools are not judged.
        let edit = json!({"file_path": out, "new_string": "x"});
        assert_eq!(
            outside_project_guard("Edit", &edit, &repo, &rule),
            Some(OUTSIDE_PROJECT)
        );
        let patch =
            json!({"command": format!("*** Begin Patch\n*** Add File: {out}\n+x\n*** End Patch")});
        assert_eq!(
            outside_project_guard("apply_patch", &patch, &repo, &rule),
            Some(OUTSIDE_PROJECT)
        );
        assert_eq!(
            outside_project_guard("Read", &write(&out), &repo, &rule),
            None
        );
        // Outside any work tree: only `allow`.
        let nogit = root.join("nogit");
        assert_eq!(
            outside_project_guard("Write", &write("x"), &nogit, &rule),
            Some(OUTSIDE_PROJECT)
        );
        assert_eq!(
            outside_project_guard("Write", &write(&allowed), &nogit, &rule),
            None
        );
    }

    /// B3 (stage 7): only in a listed primary checkout that has linked worktrees, and only
    /// for commands that move it. Look-alikes and exceptions pass.
    #[test]
    fn a_shared_primary_checkout_is_not_moved_when_turned_on() {
        let (_d, root) = layout();
        let repo = root.join("repo");
        std::fs::create_dir_all(repo.join(".git/worktrees/lane-1")).unwrap();
        let rule = crate::hook_config::PrimaryRule {
            repos: vec![repo.clone()],
            allow_commands: vec!["git switch v2026.11.0".into()],
            deny_commits_on: vec![],
        };
        for (cmd, want) in [
            ("git switch main", Some(PRIMARY_CHECKOUT)),
            ("git checkout main", Some(PRIMARY_CHECKOUT)),
            ("git checkout -b new", Some(PRIMARY_CHECKOUT)),
            ("git reset --soft HEAD~1", Some(PRIMARY_CHECKOUT)),
            ("git reset", Some(PRIMARY_CHECKOUT)),
            ("git stash", Some(PRIMARY_CHECKOUT)),
            ("git stash pop", Some(PRIMARY_CHECKOUT)),
            ("git rebase main", Some(PRIMARY_CHECKOUT)),
            ("cd . && git switch main", Some(PRIMARY_CHECKOUT)),
            ("git checkout -- a.txt", None),
            ("git checkout", None),
            ("git stash list", None),
            ("git stash show", None),
            ("git status", None),
            ("git log --oneline", None),
            ("git switch v2026.11.0", None),
            ("echo git switch main", None),
        ] {
            assert_eq!(primary_checkout_guard(cmd, &repo, &rule), want, "{cmd}");
        }
        // `-C` names the checkout the command runs in.
        assert_eq!(
            primary_checkout_guard("git -C repo switch main", &root, &rule),
            Some(PRIMARY_CHECKOUT)
        );
        assert_eq!(
            primary_checkout_guard("git switch main", &repo.join("src"), &rule),
            Some(PRIMARY_CHECKOUT)
        );
        // Not listed, or no linked worktree, or a linked worktree itself: not this rule.
        let unlisted = crate::hook_config::PrimaryRule {
            repos: vec![],
            allow_commands: vec![],
            deny_commits_on: vec![],
        };
        assert_eq!(
            primary_checkout_guard("git switch main", &repo, &unlisted),
            None
        );
        std::fs::remove_dir_all(repo.join(".git/worktrees")).unwrap();
        assert_eq!(
            primary_checkout_guard("git switch main", &repo, &rule),
            None
        );
        let linked = root.join("nogit");
        std::fs::write(linked.join(".git"), "gitdir: /elsewhere\n").unwrap();
        let rule2 = crate::hook_config::PrimaryRule {
            repos: vec![linked.clone()],
            allow_commands: vec![],
            deny_commits_on: vec![],
        };
        assert_eq!(
            primary_checkout_guard("git switch main", &linked, &rule2),
            None
        );
    }

    /// #2878 E2 (the Go generation's branch guard): `git merge` and every `git branch` that
    /// changes branches are refused in a shared primary checkout; listing forms pass.
    #[test]
    fn merges_and_branch_changes_are_refused_in_a_shared_primary_checkout() {
        let (_d, root) = layout();
        let repo = root.join("repo");
        std::fs::create_dir_all(repo.join(".git/worktrees/lane-1")).unwrap();
        let rule = crate::hook_config::PrimaryRule {
            repos: vec![repo.clone()],
            allow_commands: vec![],
            deny_commits_on: vec![],
        };
        for (cmd, want) in [
            ("git merge feature/x", Some(PRIMARY_CHECKOUT)),
            ("git merge --no-ff lane", Some(PRIMARY_CHECKOUT)),
            ("git merge --abort", Some(PRIMARY_CHECKOUT)),
            ("git branch newbranch", Some(PRIMARY_CHECKOUT)),
            ("git branch newbranch origin/main", Some(PRIMARY_CHECKOUT)),
            ("git branch -d old", Some(PRIMARY_CHECKOUT)),
            ("git branch -D old", Some(PRIMARY_CHECKOUT)),
            ("git branch -df old", Some(PRIMARY_CHECKOUT)),
            ("git branch -m a b", Some(PRIMARY_CHECKOUT)),
            ("git branch -M a b", Some(PRIMARY_CHECKOUT)),
            ("git branch -c a b", Some(PRIMARY_CHECKOUT)),
            ("git branch -f main HEAD~1", Some(PRIMARY_CHECKOUT)),
            ("git branch -u origin/x", Some(PRIMARY_CHECKOUT)),
            (
                "git branch --set-upstream-to=origin/x",
                Some(PRIMARY_CHECKOUT),
            ),
            ("git branch --unset-upstream", Some(PRIMARY_CHECKOUT)),
            ("git branch --delete old", Some(PRIMARY_CHECKOUT)),
            ("git branch --move a b", Some(PRIMARY_CHECKOUT)),
            ("git branch -q quiet-new", Some(PRIMARY_CHECKOUT)),
            ("git branch", None),
            ("git branch -a", None),
            ("git branch -vv", None),
            ("git branch -r", None),
            ("git branch --list 'issue/*'", None),
            ("git branch -l 'issue/*'", None),
            ("git branch --contains HEAD", None),
            ("git branch --merged main", None),
            ("git branch --no-merged=main", None),
            ("git branch --show-current", None),
            ("git branch --format '%(refname)'", None),
            ("git merge-base HEAD main", None),
            ("git merge-tree --write-tree a b", None),
        ] {
            assert_eq!(primary_checkout_guard(cmd, &repo, &rule), want, "{cmd}");
        }
        // A linked worktree is not this rule, whatever the command.
        let linked = root.join("nogit");
        std::fs::write(linked.join(".git"), "gitdir: /elsewhere\n").unwrap();
        let rule2 = crate::hook_config::PrimaryRule {
            repos: vec![linked.clone()],
            allow_commands: vec![],
            deny_commits_on: vec![],
        };
        assert_eq!(primary_checkout_guard("git merge x", &linked, &rule2), None);
    }

    /// #2878 E2 (the Go generation's `deny_commits_on`): in a listed primary checkout whose
    /// HEAD is a protected branch, a new commit, revert or cherry-pick is refused — with or
    /// without linked worktrees; other branches, a detached HEAD and `--abort` pass.
    #[test]
    fn commits_on_a_protected_branch_are_refused_in_the_primary_checkout() {
        let (_d, root) = layout();
        let repo = root.join("repo");
        std::fs::create_dir_all(repo.join(".git")).unwrap();
        let head = |s: &str| std::fs::write(repo.join(".git/HEAD"), s).unwrap();
        head("ref: refs/heads/main\n");
        let rule = crate::hook_config::PrimaryRule {
            repos: vec![repo.clone()],
            allow_commands: vec!["git commit -m release".into()],
            deny_commits_on: vec!["main".into()],
        };
        for (cmd, want) in [
            ("git commit -m x", Some(PROTECTED_BRANCH_COMMIT)),
            (
                "git commit --amend --no-edit",
                Some(PROTECTED_BRANCH_COMMIT),
            ),
            ("git revert abc123", Some(PROTECTED_BRANCH_COMMIT)),
            ("git cherry-pick abc123", Some(PROTECTED_BRANCH_COMMIT)),
            ("git -C . commit -m x", Some(PROTECTED_BRANCH_COMMIT)),
            ("git cherry-pick --abort", None),
            ("git revert --quit", None),
            ("git commit -m release", None),
            ("git status", None),
        ] {
            assert_eq!(primary_checkout_guard(cmd, &repo, &rule), want, "{cmd}");
        }
        head("ref: refs/heads/issue/1-x\n");
        assert_eq!(
            primary_checkout_guard("git commit -m x", &repo, &rule),
            None
        );
        head("0123456789abcdef0123456789abcdef01234567\n");
        assert_eq!(
            primary_checkout_guard("git commit -m x", &repo, &rule),
            None
        );
        // Without the list, commits are not this rule's business.
        head("ref: refs/heads/main\n");
        let no_list = crate::hook_config::PrimaryRule {
            repos: vec![repo.clone()],
            allow_commands: vec![],
            deny_commits_on: vec![],
        };
        assert_eq!(
            primary_checkout_guard("git commit -m x", &repo, &no_list),
            None
        );
        // Not listed: not this rule.
        let unlisted = crate::hook_config::PrimaryRule {
            repos: vec![],
            allow_commands: vec![],
            deny_commits_on: vec!["main".into()],
        };
        assert_eq!(
            primary_checkout_guard("git commit -m x", &repo, &unlisted),
            None
        );
    }

    #[test]
    fn writes_into_git_and_credential_paths_are_refused() {
        let dir = tempfile::TempDir::new().unwrap();
        let cwd = dir.path();
        let write =
            |path: &str| write_guard("Write", &json!({"file_path": path, "content": "x"}), cwd);
        for path in [
            ".git/config",
            ".git/hooks/pre-commit",
            "/home/u/.ssh/authorized_keys",
            "~/.ssh/config",
            "/home/u/.aws/credentials",
            "/home/u/.config/gcloud/credentials.db",
            "/home/u/.kube/config",
            "/home/u/.gnupg/pubring.kbx",
            "keys/id_ed25519",
            "/home/u/id_rsa",
        ] {
            assert_eq!(write(path), Some(PROTECTED_PATH), "{path}");
        }
        for path in [
            ".gitignore",
            ".github/workflows/ci.yml",
            "src/git/mod.rs",
            "certs/server.pem",
            "keys/id_ed25519.pub",
            "docs/ssh.md",
            "/tmp/scratch/out.txt",
        ] {
            assert_eq!(write(path), None, "{path}");
        }
        for tool in ["Edit", "MultiEdit"] {
            assert_eq!(
                write_guard(tool, &json!({"file_path": ".git/HEAD"}), cwd),
                Some(PROTECTED_PATH)
            );
        }
        assert_eq!(
            write_guard(
                "NotebookEdit",
                &json!({"notebook_path": ".git/x.ipynb"}),
                cwd
            ),
            Some(PROTECTED_PATH)
        );
        assert_eq!(
            write_guard("Read", &json!({"file_path": ".git/config"}), cwd),
            None
        );
    }

    // Creating the symlink fixture uses the Unix API; the resolution logic itself is portable.
    #[cfg(unix)]
    #[test]
    fn a_symlink_into_a_protected_place_is_followed() {
        let dir = tempfile::TempDir::new().unwrap();
        let root = std::fs::canonicalize(dir.path()).unwrap();
        std::fs::create_dir_all(root.join("real/.ssh")).unwrap();
        std::os::unix::fs::symlink(root.join("real/.ssh"), root.join("innocent")).unwrap();
        assert_eq!(
            write_guard("Write", &json!({"file_path": "innocent/config"}), &root),
            Some(PROTECTED_PATH)
        );
    }

    #[test]
    fn private_keys_and_github_tokens_in_content_are_refused() {
        let dir = tempfile::TempDir::new().unwrap();
        let cwd = dir.path();
        let key = format!(
            "{}{} RSA {}\nMIIB\n",
            "-----",
            "BEGIN",
            concat!("PRIVATE", " KEY-----")
        );
        let ghp = format!("{}{}", "ghp_", "a".repeat(36));
        let pat = format!("{}{}", "github_pat_", "b_".repeat(11));
        for content in [&key, &ghp, &pat] {
            assert_eq!(
                write_guard(
                    "Write",
                    &json!({"file_path": "a.txt", "content": content}),
                    cwd
                ),
                Some(SECRET_CONTENT)
            );
        }
        assert_eq!(
            write_guard(
                "Edit",
                &json!({"file_path": "a.txt", "new_string": key}),
                cwd
            ),
            Some(SECRET_CONTENT)
        );
        assert_eq!(
            write_guard(
                "MultiEdit",
                &json!({"file_path": "a.txt", "edits": [{"new_string": "ok"}, {"new_string": ghp}]}),
                cwd
            ),
            Some(SECRET_CONTENT)
        );
        let short = format!("{}{}", "ghp_", "a".repeat(35));
        let public = format!("{}{} PUBLIC KEY-----", "-----", "BEGIN");
        for content in [
            short.as_str(),
            public.as_str(),
            "token = env_var",
            "use ghp_ prefix",
        ] {
            assert_eq!(
                write_guard(
                    "Write",
                    &json!({"file_path": "a.txt", "content": content}),
                    cwd
                ),
                None,
                "{content}"
            );
        }
    }

    #[test]
    fn codex_patches_are_read_from_their_file_lines() {
        let dir = tempfile::TempDir::new().unwrap();
        let cwd = dir.path();
        let patch =
            |body: &str| json!({"command": format!("*** Begin Patch\n{body}\n*** End Patch")});
        for body in [
            "*** Update File: .git/config\n@@\n-a\n+b",
            "*** Add File: /home/u/.ssh/config\n+Host x",
            "*** Delete File: .git/HEAD",
            "*** Update File: a.txt\n*** Move to: .git/x",
        ] {
            assert_eq!(
                write_guard("apply_patch", &patch(body), cwd),
                Some(PROTECTED_PATH),
                "{body}"
            );
        }
        let key = format!(
            "+{}{} EC {}",
            "-----",
            "BEGIN",
            concat!("PRIVATE", " KEY-----")
        );
        assert_eq!(
            write_guard(
                "apply_patch",
                &patch(&format!("*** Add File: k.txt\n{key}")),
                cwd
            ),
            Some(SECRET_CONTENT)
        );
        assert_eq!(
            write_guard(
                "apply_patch",
                &patch("*** Update File: src/a.rs\n@@\n-a\n+b"),
                cwd
            ),
            None
        );
        // An unreadable patch is not path-checked (documented gap), but never an error.
        assert_eq!(
            write_guard("apply_patch", &json!({"command": "garbage"}), cwd),
            None
        );
    }
}
