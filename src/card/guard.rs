//! The guard of the integration branch (SPEC-20261121-ref-guard): a git `reference-transaction`
//! hook refuses every create, move and delete of the integration branch unless the intent record
//! the landing command wrote names exactly that move, keeps the case variants and the symbolic
//! aliases of its name out of `refs/heads`, and lets `pack-refs` and `gc` through.
//!
//! Two files live in the common git folder, one value for the whole repository:
//!
//! - [`RECORD_FILE`]: the full reference of the protected branch (`refs/heads/<name>`). The
//!   profile is a tracked file, so each worktree reads the one of its own branch; the record is
//!   what the hook reads. It is written by the landing command (and the later setup that installs
//!   the hook) from the profile read in the primary worktree, only when it is absent.
//! - [`INTENT_FILE`]: one JSON object naming `ref`, `old` and `new`, written by the lander just
//!   before its `update-ref` and removed after it; single use.
//!
//! The hook text ([`hook_text`]) filters in shell, so an ordinary commit starts no program: it
//! starts `moai card guard <state>` only for a transaction that names the recorded branch, a
//! variant of it or a symbolic value. [`run`] is that command; [`judge`] is its decision, a pure
//! function over the transaction lines and what git says about the refs.
//!
//! Git is run two ways: for the lander the way every card command runs it (through the git
//! environment scrub, in the project folder), and for the hook with the environment git itself
//! set, because the hook must work in the repository git is working in.

use super::landed;
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Output, Stdio};

/// The record of the protected branch, in the common git folder.
pub const RECORD_FILE: &str = "moai/integration-branch";
/// The intent record of the lander, in the common git folder.
pub const INTENT_FILE: &str = "moai/land-intent";

const ZERO: &str = "0000000000000000000000000000000000000000";
const HEADS: &str = "refs/heads/";

/// The protected branch as the one reader finds it.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Protected {
    /// The full reference, `refs/heads/<name>`.
    pub full_ref: String,
    /// The profile's branch name when it disagrees with the record.
    pub profile_disagrees: Option<String>,
    /// Whether this call wrote the record.
    pub written: bool,
}

/// The hook text (a POSIX shell script) for the managed-hook installer. It holds no branch name:
/// it reads the record.
pub fn hook_text() -> &'static str {
    r#"#!/bin/sh
# moai reference-transaction hook: the integration branch moves only through the landing command
# (SPEC-20261121-ref-guard). Written by moai; changes are overwritten.
case "$1" in prepared | committed) ;; *) exit 0 ;; esac
common=${GIT_COMMON_DIR:-}
if [ -z "$common" ]; then
  case "$0" in /*/hooks/*) common=${0%/hooks/*} ;; esac
  [ -f "$common/moai/integration-branch" ] || common=$(git rev-parse --git-common-dir 2>/dev/null) || exit 0
fi
case "$common" in /*) ;; *) common=$(pwd)/$common ;; esac
record="$common/moai/integration-branch"
[ -f "$record" ] || exit 0
line=
IFS= read -r line < "$record" || :
line=${line#"${line%%[![:space:]]*}"}
line=${line%"${line##*[![:space:]]}"}
case "$line" in refs/heads/?*) name=${line#refs/heads/} ;; *) exit 0 ;; esac
input=$(cat)
lowered=$(printf '%s\n%s\n' "$name" "$input" | tr 'A-Z' 'a-z')
nl='
'
name=${lowered%%"$nl"*}
lines=${lowered#*"$nl"}
hit=0
deletes=0
source=
while read -r old new ref; do
  for value in "$old" "$new"; do
    case "$value" in "ref:refs/heads/$name" | "ref:refs/heads/$name"/*) hit=1 ;; esac
  done
  case "$ref" in "refs/heads/$name" | "refs/heads/$name"/*) hit=1 ;; esac
  case "$ref" in refs/heads/*) case "$new" in *[!0]*) ;; *) deletes=1; source=$ref ;; esac ;; esac
  [ "$hit" = 1 ] && break
done <<EOF
$lines
EOF
if [ "$hit" = 0 ] && [ "$deletes" = 1 ] && [ "$1" = prepared ]; then
  words=$(ps -o command= -p "$PPID" 2>/dev/null | tr 'A-Z ' 'a-z\n')
  case "$nl$words$nl" in
    *"$nl$name$nl"* | *"${nl}refs/heads/$name$nl"*)
      # git moved the source's reflog aside before this delete and does not move it back
      moved="$common/logs/refs/.tmp-renamed-log"
      if [ -n "$source" ] && [ -f "$moved" ] && [ ! -e "$common/logs/$source" ]; then
        mkdir -p "${common}/logs/${source%/*}" 2>/dev/null
        mv "$moved" "$common/logs/$source" 2>/dev/null
      fi
      echo "moai ref guard refused: this command renames a branch onto the integration branch, and git deletes the source branch before it touches the integration branch; nothing was changed. The integration branch moves only through the landing command (\`moai card land\`)" >&2
      exit 1 ;;
  esac
fi
[ "$hit" = 1 ] || exit 0
moai=$(command -v moai 2>/dev/null)
if [ -z "$moai" ] && [ -x "$HOME/.moai/bin/moai" ]; then moai="$HOME/.moai/bin/moai"; fi
if [ -z "$moai" ]; then
  echo "moai ref guard refused: the moai program was not found on PATH or in \$HOME/.moai/bin, so a change of the integration branch cannot be checked" >&2
  [ "$1" = prepared ] && exit 1
  exit 0
fi
printf '%s\n' "$input" | "$moai" card guard "$1" >/dev/null
rc=$?
[ "$1" = prepared ] && exit $rc
exit 0
"#
}

// ---- git -------------------------------------------------------------------------------------

/// Where git runs and with which environment.
#[derive(Clone, Copy)]
enum Env<'a> {
    /// The lander: through the git environment scrub, in the project folder.
    Project(&'a Path),
    /// The hook: with the environment git set for it, in the current folder.
    Hook,
}

fn git(env: Env<'_>, args: &[&str]) -> Result<Output, String> {
    // `MOAI_GIT_BIN` names the git program for a runner whose child carries no `PATH`.
    let program = std::env::var("MOAI_GIT_BIN").unwrap_or_else(|_| "git".to_string());
    let mut cmd = Command::new(program);
    match env {
        Env::Project(project) => {
            crate::identity::scrub_git_env(&mut cmd);
            cmd.arg("--no-optional-locks").arg("-C").arg(project);
        }
        Env::Hook => {
            cmd.env("LC_ALL", "C").arg("--no-optional-locks");
        }
    }
    cmd.args(args)
        .stdin(Stdio::null())
        .output()
        .map_err(|e| format!("git could not run: {e}"))
}

fn text_of(out: &Output) -> String {
    String::from_utf8_lossy(&out.stdout).trim().to_string()
}

/// The common git folder of the repository `env` names (the shared one for every worktree).
fn common_dir(env: Env<'_>) -> Result<PathBuf, String> {
    let out = git(env, &["rev-parse", "--git-common-dir"])?;
    if !out.status.success() {
        return Err(format!(
            "git cannot find the repository: {}",
            String::from_utf8_lossy(&out.stderr).trim()
        ));
    }
    let found = PathBuf::from(text_of(&out));
    if found.is_absolute() {
        return Ok(found);
    }
    let base = match env {
        Env::Project(project) => project.to_path_buf(),
        Env::Hook => std::env::current_dir().map_err(|e| format!("no current folder: {e}"))?,
    };
    Ok(base.join(found))
}

/// The common git folder of the repository of `project`.
pub fn common_dir_of(project: &Path) -> Result<PathBuf, String> {
    common_dir(Env::Project(project))
}

/// The commit `reference` names now (a symbolic ref names what it points at), or `None`.
fn resolve(env: Env<'_>, reference: &str) -> Option<String> {
    let out = git(env, &["rev-parse", "-q", "--verify", reference]).ok()?;
    out.status.success().then(|| text_of(&out))
}

fn is_ancestor(env: Env<'_>, old: &str, new: &str) -> bool {
    git(env, &["merge-base", "--is-ancestor", old, new]).is_ok_and(|out| out.status.success())
}

/// The value `reference` has in the packed store, when it has one.
fn packed_value(common: &Path, reference: &str) -> Option<String> {
    let text = fs::read_to_string(common.join("packed-refs")).ok()?;
    text.lines()
        .filter(|l| !l.starts_with('#') && !l.starts_with('^'))
        .find_map(|l| {
            let (sha, name) = l.split_once(' ')?;
            (name == reference).then(|| sha.to_string())
        })
}

// ---- the record and the intent ----------------------------------------------------------------

/// Opens a new file for writing; `private` makes it readable and writable by its owner only.
#[cfg(unix)]
fn create_file(path: &Path, private: bool) -> std::io::Result<fs::File> {
    use std::os::unix::fs::OpenOptionsExt;
    let mut options = fs::OpenOptions::new();
    options.write(true).create_new(true);
    if private {
        options.mode(0o600);
    }
    options.open(path)
}

#[cfg(not(unix))]
fn create_file(path: &Path, _private: bool) -> std::io::Result<fs::File> {
    fs::OpenOptions::new()
        .write(true)
        .create_new(true)
        .open(path)
}

/// Writes `bytes` to `path` through a temporary file and a rename, so a reader never sees half.
fn write_atomically(path: &Path, bytes: &[u8], private: bool) -> Result<(), String> {
    use std::io::Write;
    let folder = path.parent().ok_or("no folder to write in")?;
    fs::create_dir_all(folder).map_err(|e| format!("cannot make {}: {e}", folder.display()))?;
    let tmp = folder.join(format!(
        ".{}.tmp{}",
        path.file_name()
            .and_then(|n| n.to_str())
            .unwrap_or("record"),
        std::process::id()
    ));
    let _ = fs::remove_file(&tmp);
    create_file(&tmp, private)
        .and_then(|mut f| f.write_all(bytes))
        .map_err(|e| format!("cannot write {}: {e}", tmp.display()))?;
    fs::rename(&tmp, path).map_err(|e| format!("cannot write {}: {e}", path.display()))
}

/// The full reference the record names, or `None` when there is no usable record.
fn read_record(common: &Path) -> Option<String> {
    let text = fs::read_to_string(common.join(RECORD_FILE)).ok()?;
    let line = text.lines().next()?.trim();
    (line.starts_with(HEADS) && line.len() > HEADS.len()).then(|| line.to_string())
}

fn write_record(common: &Path, full_ref: &str) -> Result<(), String> {
    write_atomically(
        &common.join(RECORD_FILE),
        format!("{full_ref}\n").as_bytes(),
        false,
    )
}

/// The branch name (without `refs/heads/`) the record of `project`'s repository names, or `None`.
pub fn recorded_branch(project: &Path) -> Option<String> {
    let common = common_dir_of(project).ok()?;
    read_record(&common).map(|full| full[HEADS.len()..].to_string())
}

/// What the lander says it is about to do: `old` to `new` of `reference`.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Intent {
    pub reference: String,
    pub old: String,
    pub new: String,
}

fn read_intent(common: &Path) -> Option<Intent> {
    let text = fs::read_to_string(common.join(INTENT_FILE)).ok()?;
    let value: Value = serde_json::from_str(&text).ok()?;
    let field = |key: &str| value.get(key)?.as_str().map(str::to_string);
    Some(Intent {
        reference: field("ref")?,
        old: field("old")?,
        new: field("new")?,
    })
}

/// Writes the intent record of a move of `reference` from `old` to `new`, replacing any older one.
pub fn write_intent(common: &Path, reference: &str, old: &str, new: &str) -> Result<(), String> {
    let at = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map_or(0, |d| d.as_secs());
    let body = json!({
        "ref": reference, "old": old, "new": new, "by": "card land", "at": at,
    });
    write_atomically(&common.join(INTENT_FILE), body.to_string().as_bytes(), true)
}

/// Removes the intent record; a record that is not there is fine.
pub fn remove_intent(common: &Path) {
    let _ = fs::remove_file(common.join(INTENT_FILE));
}

// ---- the one reader of the protected name -------------------------------------------------------

/// Whether `project` is the primary worktree of its repository.
fn is_primary_worktree(project: &Path) -> bool {
    let Ok(out) = git(Env::Project(project), &["worktree", "list", "--porcelain"]) else {
        return false;
    };
    let text = String::from_utf8_lossy(&out.stdout).into_owned();
    let first = text
        .lines()
        .find_map(|l| l.strip_prefix("worktree "))
        .map(PathBuf::from);
    let Some(first) = first else {
        return false;
    };
    match (fs::canonicalize(first), fs::canonicalize(project)) {
        (Ok(a), Ok(b)) => a == b,
        _ => false,
    }
}

/// The protected branch of the repository of `project`: the record first; for a repository with
/// no record the profile's value, written as the record when `write_if_absent` and `project` is
/// the primary worktree. A profile that disagrees with the record is reported, not obeyed. `None`
/// when nothing is protected (no record, and no profile value to seed one from or no leave to).
pub fn protected(project: &Path, write_if_absent: bool) -> Result<Option<Protected>, String> {
    let common = common_dir_of(project)?;
    let profile = landed::profile_branch(project)
        .ok()
        .and_then(|name| landed::branch_ref(&name).ok());
    if let Some(full_ref) = read_record(&common) {
        let profile_disagrees = profile
            .filter(|p| *p != full_ref)
            .map(|p| p.strip_prefix(HEADS).unwrap_or(&p).to_string());
        return Ok(Some(Protected {
            full_ref,
            profile_disagrees,
            written: false,
        }));
    }
    let Some(full_ref) = profile else {
        return Ok(None);
    };
    if !write_if_absent {
        return Ok(None);
    }
    let written = is_primary_worktree(project);
    if written {
        write_record(&common, &full_ref)?;
    }
    Ok(Some(Protected {
        full_ref,
        profile_disagrees: None,
        written,
    }))
}

/// Whether `name` is a case variant of the protected `protected`: equal to it without regard to
/// letter case but spelled differently, or beginning with such a spelling and a slash.
pub fn is_variant(name: &str, protected: &str) -> bool {
    if name == protected || name.starts_with(&format!("{protected}/")) {
        return false;
    }
    let (n, p) = (name.to_lowercase(), protected.to_lowercase());
    n == p || n.starts_with(&format!("{p}/"))
}

/// The branch references that are case variants of `full_ref`, sorted.
pub fn variants(project: &Path, full_ref: &str) -> Result<Vec<String>, String> {
    let out = git(
        Env::Project(project),
        &["for-each-ref", "--format=%(refname)", "refs/heads"],
    )?;
    if !out.status.success() {
        return Err(format!(
            "git could not list the branches: {}",
            String::from_utf8_lossy(&out.stderr).trim()
        ));
    }
    let mut found: Vec<String> = String::from_utf8_lossy(&out.stdout)
        .lines()
        .filter(|name| is_variant(name, full_ref))
        .map(str::to_string)
        .collect();
    found.sort();
    Ok(found)
}

// ---- the decision ---------------------------------------------------------------------------------

/// One line of a reference transaction as git hands it to the hook.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Line {
    pub old: String,
    pub new: String,
    pub name: String,
}

/// The lines of the hook's standard input; a line that is not three words is skipped.
pub fn parse_lines(input: &str) -> Vec<Line> {
    input
        .lines()
        .filter_map(|l| {
            let mut words = l.split_whitespace();
            let (old, new, name) = (words.next()?, words.next()?, words.next()?);
            Some(Line {
                old: old.to_string(),
                new: new.to_string(),
                name: name.to_string(),
            })
        })
        .collect()
}

/// What the decision needs to know about the repository.
pub struct Facts<'a> {
    /// The protected reference.
    pub protected: &'a str,
    /// The commit a reference names now.
    pub current: &'a dyn Fn(&str) -> Option<String>,
    /// Whether the reference's loose file exists.
    pub loose: &'a dyn Fn(&str) -> bool,
    /// The reference's value in the packed store.
    pub packed: &'a dyn Fn(&str) -> Option<String>,
    /// The lander's intent record, when there is a readable one.
    pub intent: Option<Intent>,
    /// Whether the first commit is an ancestor of the second.
    pub is_ancestor: &'a dyn Fn(&str, &str) -> bool,
}

/// Decides one transaction. `Err` is the reason it is refused, without the guard's frame.
pub fn judge(facts: &Facts<'_>, lines: &[Line]) -> Result<(), String> {
    let protected = facts.protected;
    let spelled = protected.strip_prefix(HEADS).unwrap_or(protected);
    for line in lines {
        let name = line.name.as_str();
        if !name.starts_with(HEADS) {
            continue;
        }
        let branch = &name[HEADS.len()..];
        let symbolic_target = line
            .new
            .strip_prefix("ref:")
            .or_else(|| line.old.strip_prefix("ref:"));
        if is_variant(name, protected) {
            return Err(format!(
                "a branch named `{branch}` differs from the integration branch `{spelled}` only in \
                 letter case, and on a case-insensitive file system it would overwrite the branch's \
                 own file; the right spelling is `{spelled}`, which is never created, moved or \
                 deleted by hand"
            ));
        }
        if name == protected {
            if symbolic_target.is_some() {
                return Err(format!(
                    "the integration branch `{spelled}` cannot be made a symbolic ref"
                ));
            }
            judge_move(facts, line)?;
        } else if let Some(target) = symbolic_target {
            if target == protected || is_variant(target, protected) {
                return Err(format!(
                    "`{branch}` would be an alias of the integration branch `{spelled}`; an alias \
                     is never made, because the name it hides is the branch's own"
                ));
            }
        }
    }
    Ok(())
}

/// A create, move or delete of the protected reference itself.
fn judge_move(facts: &Facts<'_>, line: &Line) -> Result<(), String> {
    let protected = facts.protected;
    let current = (facts.current)(protected);
    if line.new == ZERO {
        // The prune of a loose copy after packing deletes nothing: the loose file is there, the
        // packed store holds the same value and the transaction names that value as the old one.
        let prune = current.as_deref().is_some_and(|cur| {
            (facts.loose)(protected)
                && (facts.packed)(protected).as_deref() == Some(cur)
                && line.old == cur
        });
        return if prune {
            Ok(())
        } else {
            Err(
                "the integration branch cannot be deleted, and no intent record allows a deletion"
                    .to_string(),
            )
        };
    }
    if current.as_deref() == Some(line.new.as_str()) {
        return Ok(()); // no change of value (packing, a repeated update)
    }
    let Some(intent) = &facts.intent else {
        return Err("no intent record names this move".to_string());
    };
    let old = current.as_deref().unwrap_or(ZERO);
    if intent.reference != protected || intent.old != old || intent.new != line.new {
        return Err(format!(
            "the intent record names {} {} -> {}, not this move ({} -> {})",
            intent.reference, intent.old, intent.new, old, line.new
        ));
    }
    if let Some(cur) = &current {
        if !(facts.is_ancestor)(cur, &line.new) {
            return Err(
                "the intent record names a move that is not a fast-forward (the current value is \
                 not an ancestor of the new one)"
                    .to_string(),
            );
        }
    }
    Ok(())
}

// ---- the entry ------------------------------------------------------------------------------------

/// What `card guard <state>` does with the transaction lines git hands the hook (at `prepared`:
/// decide; at `committed`: consume the intent record of a move that went through). Any other state
/// does nothing. `Err` is the refusal, framed for the user of the git command.
pub fn run(state: &str, input: &str) -> Result<(), String> {
    if state != "prepared" && state != "committed" {
        return Ok(());
    }
    let env = Env::Hook;
    let common = common_dir(env).map_err(|why| format!("moai ref guard refused: {why}"))?;
    let Some(protected) = read_record(&common) else {
        return Ok(()); // no record, nothing to protect
    };
    let lines = parse_lines(input);
    let intent = read_intent(&common);
    if state == "committed" {
        let used = lines.iter().any(|l| {
            intent
                .as_ref()
                .is_some_and(|i| l.name == protected && i.reference == protected && l.new == i.new)
        });
        if used {
            remove_intent(&common);
        }
        return Ok(());
    }
    let current = |r: &str| resolve(env, r);
    let loose = |r: &str| common.join(r).is_file();
    let packed = |r: &str| packed_value(&common, r);
    let ancestor = |a: &str, b: &str| is_ancestor(env, a, b);
    let facts = Facts {
        protected: &protected,
        current: &current,
        loose: &loose,
        packed: &packed,
        intent,
        is_ancestor: &ancestor,
    };
    judge(&facts, &lines).map_err(|why| {
        let spelled = protected.strip_prefix(HEADS).unwrap_or(&protected);
        let mut message = format!(
            "moai ref guard refused: {why}. The integration branch `{spelled}` moves only through \
             the landing command (`moai card land`), which writes the intent record this guard \
             checks."
        );
        if let Some(note) = profile_note(&protected) {
            message.push(' ');
            message.push_str(&note);
        }
        message
    })
}

/// A sentence about the checked-out profile when it names another integration branch than the
/// record (the record wins), for the guard's refusal.
fn profile_note(protected: &str) -> Option<String> {
    let top = git(Env::Hook, &["rev-parse", "--show-toplevel"]).ok()?;
    if !top.status.success() {
        return None;
    }
    let name = landed::profile_branch(Path::new(&text_of(&top))).ok()?;
    let full = landed::branch_ref(&name).ok()?;
    (full != protected).then(|| {
        let spelled = protected.strip_prefix(HEADS).unwrap_or(protected);
        format!(
            "The profile of this worktree names `{name}`; the record names `{spelled}` and wins."
        )
    })
}
