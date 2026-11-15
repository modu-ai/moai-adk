//! `check-tree` (#3009 S3a, design comment 6887461482 §2 "(가) 검사"): what the tracked MoAI
//! folders of a project hold that they should not. The team-shared folder is for small, hand-
//! edited text; machine-written data lives in the HOME. This reads the files git tracks under
//! `.moai/` and `.moai-cli/` and warns about five things:
//!
//! * `large_file`: a file over [`SIZE_LIMIT`] (1 MiB; exactly 1 MiB is fine);
//! * `absolute_path`: text that holds `/Users/<name>` or `/home/<name>`;
//! * `session_id`: a session UUID in a path, or in a text line that mentions a session;
//! * `secret_name`: a file named like a credential (`.env`, `*.pem`, `*.key`, ...);
//! * `project_id_not_random`: the profile's `project_id` is a name a person chose, which can be
//!   the same as another project's (decision A, comment 6889383420); it is never changed here.
//!
//! This is the warning stage: the command always answers, never fails on a finding, and
//! writes nothing (one `git ls-files`, then reads). A finding names the path and a count or a
//! size; the matched text (a user name, a UUID, an id) is never repeated. A file is read only
//! when every folder above it is a real folder (a linked one could lead out of the project),
//! and only the paths in scope are decoded (a file name that is not UTF-8 elsewhere changes
//! nothing); what is not read is counted in `not_read`.
use crate::{contract::PROJECT_PATH_UNAVAILABLE, project_id, worktree_include::secret_name};
use serde_json::{json, Value};
use std::collections::BTreeMap;
use std::fs::{self, File};
use std::io::Read;
use std::path::Path;
use std::process::Command;

type Result<T> = std::result::Result<T, String>;

/// Largest file that is not a warning: exactly this many bytes passes, one more is flagged.
pub const SIZE_LIMIT: u64 = 1024 * 1024;

/// Findings listed in the answer; the counts always cover all of them.
pub const FINDINGS_MAX: usize = 200;

pub const LARGE_FILE: &str = "large_file";
pub const ABSOLUTE_PATH: &str = "absolute_path";
pub const SESSION_ID: &str = "session_id";
pub const SECRET_NAME: &str = "secret_name";
pub const ID_NOT_RANDOM: &str = "project_id_not_random";
pub const ID_UNUSABLE: &str = "project_id_unusable";

/// One warning: where, what kind, and how much (`bytes` for a large file, `hits` otherwise).
#[derive(Clone, Debug, PartialEq, Eq, PartialOrd, Ord)]
pub struct Finding {
    pub path: String,
    pub kind: &'static str,
    pub amount: Option<(&'static str, u64)>,
}

impl Finding {
    fn to_json(&self) -> Value {
        let mut v = json!({"path": self.path, "kind": self.kind});
        if let Some((unit, n)) = self.amount {
            v[unit] = json!(n);
        }
        v
    }
}

/// Whether a tracked path is under one of the two MoAI folders at the top of the project.
pub fn in_scope(path: &str) -> bool {
    in_scope_bytes(path.as_bytes())
}

/// [`in_scope`] on the raw bytes git printed, so a path that is not UTF-8 can be sorted into
/// "in scope" or "not" before anything tries to decode it.
fn in_scope_bytes(path: &[u8]) -> bool {
    let Some(slash) = path.iter().position(|&b| b == b'/') else {
        return false;
    };
    let first = &path[..slash];
    [crate::state_root::MOAI_DIR, crate::PROJECT_STATE_DIR]
        .iter()
        .any(|d| first.eq_ignore_ascii_case(d.as_bytes()))
}

fn name_byte(b: u8) -> bool {
    b.is_ascii_alphanumeric() || matches!(b, b'.' | b'_' | b'-')
}

/// Whether the text at `bytes[at..]` starts an account name: an ASCII name byte, or a letter
/// or digit of another script (`홍길동`, `élise`) read as UTF-8. A byte that is not the start of
/// a whole UTF-8 character, a space (also U+00A0) and a symbol start no name.
fn name_start(bytes: &[u8], at: usize) -> bool {
    let Some(&first) = bytes.get(at) else {
        return false;
    };
    if first < 0x80 {
        return name_byte(first);
    }
    let rest = &bytes[at..bytes.len().min(at + 4)];
    let valid = match std::str::from_utf8(rest) {
        Ok(s) => s,
        Err(e) => std::str::from_utf8(&rest[..e.valid_up_to()]).unwrap_or(""),
    };
    valid.chars().next().is_some_and(char::is_alphanumeric)
}

/// Bytes that may stand right before an absolute path in text (a space, a quote, a bracket,
/// `=`, `:`, `,`), so `lib/Users/x` and `a/home/b` are not mistaken for one.
fn path_start_boundary(b: u8) -> bool {
    b.is_ascii_whitespace()
        || matches!(
            b,
            b'"' | b'\'' | b'`' | b'(' | b'[' | b'{' | b'<' | b'=' | b':' | b','
        )
}

/// The shared credential-name rule (`.env` family, keys, certificates, `credentials`), plus
/// the design's `*.env` (`prod.env`), which the shared rule leaves out.
pub fn secret_name_here(name: &str) -> bool {
    let n = name.to_ascii_lowercase();
    secret_name(name) || (n.len() > ".env".len() && n.ends_with(".env"))
}

/// How many times `bytes` holds `/Users/<name>` or `/home/<name>` as a whole path start.
pub fn absolute_path_hits(bytes: &[u8]) -> usize {
    let mut hits = 0;
    for prefix in [&b"/Users/"[..], &b"/home/"[..]] {
        let mut from = 0;
        while let Some(at) = bytes[from..]
            .windows(prefix.len())
            .position(|w| w == prefix)
            .map(|i| i + from)
        {
            let before_ok = at == 0 || path_start_boundary(bytes[at - 1]);
            let after = at + prefix.len();
            let name_ok = name_start(bytes, after);
            if before_ok && name_ok {
                hits += 1;
            }
            from = at + prefix.len();
        }
    }
    hits
}

fn uuid_at(b: &[u8], i: usize) -> bool {
    if i + 36 > b.len() {
        return false;
    }
    let s = &b[i..i + 36];
    let shape = s.iter().enumerate().all(|(k, &c)| {
        if matches!(k, 8 | 13 | 18 | 23) {
            c == b'-'
        } else {
            c.is_ascii_hexdigit()
        }
    });
    let alnum_before = i > 0 && b[i - 1].is_ascii_alphanumeric();
    let alnum_after = b.get(i + 36).is_some_and(u8::is_ascii_alphanumeric);
    shape && !alnum_before && !alnum_after
}

/// How many UUIDs (8-4-4-4-12 hex digits, not part of a longer word) `bytes` holds.
pub fn uuid_hits(bytes: &[u8]) -> usize {
    let (mut n, mut i) = (0, 0);
    while i < bytes.len() {
        if uuid_at(bytes, i) {
            n += 1;
            i += 36;
        } else {
            i += 1;
        }
    }
    n
}

/// How many text lines mention a session (any case) and hold a UUID.
pub fn session_id_lines(bytes: &[u8]) -> usize {
    bytes
        .split(|&b| b == b'\n')
        .filter(|line| {
            line.windows(7).any(|w| w.eq_ignore_ascii_case(b"session")) && uuid_hits(line) > 0
        })
        .count()
}

/// Whether the start of a file looks binary (a NUL byte in its first 8 KiB).
fn looks_binary(bytes: &[u8]) -> bool {
    bytes.iter().take(8192).any(|&b| b == 0)
}

/// The files git tracks in a project, sorted by what concerns this check.
struct Tracked {
    /// Every distinct tracked path, whatever its encoding.
    total: usize,
    /// The paths under the two MoAI folders that are UTF-8, sorted.
    scoped: Vec<String>,
    /// Paths under the two MoAI folders whose names are not UTF-8: they cannot be named in the
    /// answer, so they are counted as not read.
    undecodable: usize,
}

/// What the files git tracks in `project` are: `git ls-files -z`, split on the raw bytes. Only
/// the paths in scope are decoded, so a file name outside the MoAI folders, whatever its
/// encoding, never stops the check.
fn tracked(project: &Path) -> Result<Tracked> {
    let mut cmd = Command::new("git");
    cmd.arg("-C")
        .arg(project)
        .args(["--no-optional-locks", "ls-files", "-z"]);
    crate::identity::scrub_git_env(&mut cmd);
    let out = cmd.output().map_err(|_| "git unavailable".to_string())?;
    if !out.status.success() {
        return Err(format!("not a git work tree: {}", project.display()));
    }
    let mut raw: Vec<&[u8]> = out
        .stdout
        .split(|&b| b == 0)
        .filter(|p| !p.is_empty())
        .collect();
    raw.sort();
    raw.dedup();
    let mut scoped = Vec::new();
    let mut undecodable = 0;
    for path in raw.iter().filter(|p| in_scope_bytes(p)) {
        match std::str::from_utf8(path) {
            Ok(s) => scoped.push(s.to_string()),
            Err(_) => undecodable += 1,
        }
    }
    Ok(Tracked {
        total: raw.len(),
        scoped,
        undecodable,
    })
}

/// How the folders above a tracked file stand: every step a real folder, one of them a link
/// (the file is then not inside the project, whatever its name says), or one of them gone.
#[derive(PartialEq, Eq)]
enum Parents {
    Real,
    Linked,
    Gone,
}

/// Looks at each folder between `root` and the file `rel`, one step at a time without
/// following any link. `symlink_metadata` of the file alone only judges its last name, and a
/// linked folder above it would lead the read out of the project.
fn parents_of(root: &Path, rel: &str) -> Parents {
    let mut dir = root.to_path_buf();
    let mut parts = rel.split('/').peekable();
    while let Some(part) = parts.next() {
        if parts.peek().is_none() {
            break;
        }
        dir.push(part);
        match fs::symlink_metadata(&dir) {
            Err(_) => return Parents::Gone,
            Ok(m) if m.file_type().is_symlink() => return Parents::Linked,
            Ok(m) if !m.is_dir() => return Parents::Gone,
            Ok(_) => {}
        }
    }
    Parents::Real
}

/// The text of a regular file of at most [`SIZE_LIMIT`] bytes; `None` for anything else.
fn read_small(path: &Path, len: u64) -> Option<Vec<u8>> {
    if len > SIZE_LIMIT {
        return None;
    }
    let mut buf = Vec::with_capacity(len as usize);
    File::open(path)
        .ok()?
        .take(SIZE_LIMIT + 1)
        .read_to_end(&mut buf)
        .ok()?;
    (buf.len() as u64 <= SIZE_LIMIT).then_some(buf)
}

/// The warnings of the tracked MoAI folders of `project`, as the command's answer.
pub fn check(project: &Path) -> Result<Value> {
    let root = fs::canonicalize(project).map_err(|_| PROJECT_PATH_UNAVAILABLE.to_string())?;
    let all = tracked(&root)?;
    let mut findings: Vec<Finding> = Vec::new();
    let scoped = all.scoped.len() + all.undecodable;
    let (mut missing, mut skipped) = (0usize, all.undecodable);
    for rel in &all.scoped {
        let mut add = |kind: &'static str, amount: Option<(&'static str, u64)>| {
            findings.push(Finding {
                path: rel.clone(),
                kind,
                amount,
            });
        };
        let name = rel.rsplit('/').next().unwrap_or(rel);
        if secret_name_here(name) {
            add(SECRET_NAME, None);
        }
        if uuid_hits(rel.as_bytes()) > 0 {
            add(SESSION_ID, Some(("hits", 1)));
        }
        match parents_of(&root, rel) {
            Parents::Real => {}
            Parents::Linked => {
                skipped += 1;
                continue;
            }
            Parents::Gone => {
                missing += 1;
                continue;
            }
        }
        let full = root.join(rel);
        let Ok(meta) = fs::symlink_metadata(&full) else {
            missing += 1;
            continue;
        };
        if !meta.is_file() {
            skipped += 1;
            continue;
        }
        if meta.len() > SIZE_LIMIT {
            add(LARGE_FILE, Some(("bytes", meta.len())));
            continue;
        }
        let Some(bytes) = read_small(&full, meta.len()) else {
            skipped += 1;
            continue;
        };
        if looks_binary(&bytes) {
            continue;
        }
        let abs = absolute_path_hits(&bytes);
        if abs > 0 {
            add(ABSOLUTE_PATH, Some(("hits", abs as u64)));
        }
        let sessions = session_id_lines(&bytes);
        if sessions > 0 && uuid_hits(rel.as_bytes()) == 0 {
            add(SESSION_ID, Some(("hits", sessions as u64)));
        }
    }
    let profile_rel = crate::state_root::rel(&root, "project.yaml");
    match project_id::read(&root) {
        Ok(Some(id)) if !project_id::is_random_form(&id) => {
            findings.push(Finding {
                path: profile_rel,
                kind: ID_NOT_RANDOM,
                amount: None,
            });
        }
        Ok(_) => {}
        Err(_) => findings.push(Finding {
            path: profile_rel,
            kind: ID_UNUSABLE,
            amount: None,
        }),
    }
    findings.sort();
    let mut counts: BTreeMap<&str, u64> = BTreeMap::new();
    for f in &findings {
        *counts.entry(f.kind).or_default() += 1;
    }
    let total = findings.len();
    let listed: Vec<Value> = findings
        .iter()
        .take(FINDINGS_MAX)
        .map(Finding::to_json)
        .collect();
    Ok(json!({
        "stage": "warning",
        "status": if total == 0 { "clean" } else { "warnings" },
        "tracked": all.total,
        "in_scope": scoped,
        "missing": missing,
        "not_read": skipped,
        "size_limit_bytes": SIZE_LIMIT,
        "counts": counts,
        "findings": listed,
        "omitted": total.saturating_sub(FINDINGS_MAX),
        "writes": false,
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    const UUID: &str = "123e4567-e89b-12d3-a456-426614174000";

    #[test]
    fn the_scope_is_the_two_folders_at_the_top() {
        let earlier = crate::PROJECT_STATE_DIR;
        for p in [
            ".moai/project.yaml".to_string(),
            format!("{earlier}/state/x"),
            ".MOAI/a".to_string(),
            ".moai/a/b/c.md".to_string(),
        ] {
            assert!(in_scope(&p), "{p}");
        }
        for p in [
            "src/.moai/x".to_string(),
            ".moai".to_string(),
            earlier.to_string(),
            ".moaix/a".to_string(),
            "moai/a".to_string(),
            "app/.moai/a".to_string(),
            String::new(),
        ] {
            assert!(!in_scope(&p), "{p:?}");
        }
    }

    #[test]
    fn a_credential_name_includes_star_dot_env_but_not_the_example_files() {
        for n in [
            ".env",
            ".env.local",
            "prod.env",
            "PROD.ENV",
            "a.pem",
            "id_rsa",
            "x.key",
        ] {
            assert!(secret_name_here(n), "{n}");
        }
        for n in [
            ".env.example",
            ".env.sample",
            "env",
            ".env_",
            "x.envx",
            "readme.md",
            "key",
        ] {
            assert!(!secret_name_here(n), "{n}");
        }
    }

    #[test]
    fn an_absolute_path_needs_a_name_after_it_and_a_boundary_before_it() {
        assert_eq!(absolute_path_hits(b"/Users/alice/x"), 1);
        assert_eq!(absolute_path_hits(b"path: /home/bob"), 1);
        assert_eq!(absolute_path_hits(b"\"/Users/a\" and (/home/b)"), 2);
        assert_eq!(absolute_path_hits(b"/Users/"), 0, "no name");
        assert_eq!(absolute_path_hits(b"/Users/ x"), 0, "a space is no name");
        assert_eq!(absolute_path_hits(b"lib/Users/x"), 0, "inside a path");
        assert_eq!(absolute_path_hits(b"a/home/b"), 0, "inside a path");
        assert_eq!(absolute_path_hits(b"/users/a /HOME/b"), 0, "case matters");
        assert_eq!(absolute_path_hits(b"nothing here"), 0);
        assert_eq!(absolute_path_hits(b""), 0);
    }

    #[test]
    fn a_uuid_is_8_4_4_4_12_hex_not_inside_a_longer_word() {
        assert_eq!(uuid_hits(UUID.as_bytes()), 1);
        assert_eq!(uuid_hits(UUID.to_uppercase().as_bytes()), 1);
        assert_eq!(uuid_hits(format!("a {UUID},{UUID}.").as_bytes()), 2);
        assert_eq!(
            uuid_hits(format!("x{UUID}").as_bytes()),
            0,
            "a letter before"
        );
        assert_eq!(uuid_hits(format!("{UUID}0").as_bytes()), 0, "a digit after");
        assert_eq!(uuid_hits(&UUID.as_bytes()[..35]), 0, "one short");
        let bad = UUID.replacen('-', "_", 1);
        assert_eq!(uuid_hits(bad.as_bytes()), 0, "a wrong separator");
        let nonhex = UUID.replacen('1', "g", 1);
        assert_eq!(uuid_hits(nonhex.as_bytes()), 0, "a non-hex digit");
        assert_eq!(uuid_hits(b""), 0);
    }

    #[test]
    fn a_session_line_needs_both_the_word_and_a_uuid() {
        let text = format!("Session: {UUID}\nsession only\n{UUID} alone\nMy SESSION {UUID}\n");
        assert_eq!(session_id_lines(text.as_bytes()), 2);
        assert_eq!(session_id_lines(b""), 0);
    }

    #[test]
    fn binary_is_a_nul_in_the_first_8_kib() {
        let mut b = vec![b'a'; 9000];
        assert!(!looks_binary(&b));
        b[8191] = 0;
        assert!(looks_binary(&b));
        b[8191] = b'a';
        b[8192] = 0;
        assert!(!looks_binary(&b), "byte 8192 is past the look");
    }

    #[test]
    fn a_finding_prints_its_unit_and_never_text() {
        let f = Finding {
            path: "a".into(),
            kind: LARGE_FILE,
            amount: Some(("bytes", 7)),
        };
        assert_eq!(
            f.to_json(),
            json!({"path": "a", "kind": "large_file", "bytes": 7})
        );
        let g = Finding {
            path: "a".into(),
            kind: SECRET_NAME,
            amount: None,
        };
        assert_eq!(g.to_json(), json!({"path": "a", "kind": "secret_name"}));
    }

    #[test]
    fn a_missing_folder_says_so() {
        let d = tempfile::tempdir().unwrap();
        let err = check(&d.path().join("nowhere")).unwrap_err();
        assert_eq!(err, PROJECT_PATH_UNAVAILABLE);
    }

    #[test]
    fn an_account_name_may_start_with_a_letter_of_any_script() {
        // B2 of the #3044 audit: Korean and accented account names were passed as clean.
        for text in [
            "/home/홍길동/project",
            "/Users/élise/project",
            "/Users/Ünal",
            "/home/日本語",
            "/home/1",
            "x=/home/_u",
        ] {
            assert_eq!(absolute_path_hits(text.as_bytes()), 1, "{text}");
        }
        // The edges: nothing, a separator, a space (also U+00A0), a symbol, a byte that starts
        // no UTF-8 character, a character cut in two at the end.
        let cases: [&[u8]; 9] = [
            b"/home/",
            b"/home//x",
            b"/home/ x",
            "/home/\u{a0}x".as_bytes(),
            "/home/\u{1f600}x".as_bytes(),
            b"/home/\xff",
            b"/home/\x80x",
            &"/home/홍".as_bytes()[..8],
            b"/Users/\n",
        ];
        for bytes in cases {
            assert_eq!(absolute_path_hits(bytes), 0, "{bytes:?}");
        }
        // A whole character before a cut one still counts once.
        assert_eq!(absolute_path_hits(&"/home/홍길".as_bytes()[..10]), 1);
    }

    #[test]
    fn a_path_is_sorted_into_scope_before_it_is_decoded() {
        // B1/N1 of the #3044 audit: a name that is not UTF-8 can be told in or out of scope.
        assert!(in_scope_bytes(b".moai/\xff"));
        assert!(in_scope_bytes(b".MOAI/a/\xfe\xff"));
        assert!(!in_scope_bytes(b"src/\xff"));
        assert!(!in_scope_bytes(b"\xff/a"));
        assert!(!in_scope_bytes(b".moai"));
        assert!(!in_scope_bytes(b""));
        assert!(!in_scope_bytes(b".moai\xff/a"));
    }

    #[cfg(unix)]
    #[test]
    fn a_linked_or_missing_folder_above_a_file_is_told_from_a_real_one() {
        let d = tempfile::tempdir().unwrap();
        let root = d.path().join("project");
        let outside = d.path().join("outside");
        fs::create_dir_all(root.join(".moai/real/deep")).unwrap();
        fs::create_dir_all(&outside).unwrap();
        fs::write(root.join(".moai/file"), "x").unwrap();
        std::os::unix::fs::symlink(&outside, root.join(".moai/linked")).unwrap();
        std::os::unix::fs::symlink(&outside, root.join("top")).unwrap();
        assert!(parents_of(&root, ".moai/a") == Parents::Real, "one step");
        assert!(parents_of(&root, ".moai/real/deep/a") == Parents::Real);
        assert!(
            parents_of(&root, ".moai/linked/a") == Parents::Linked,
            "mid"
        );
        assert!(
            parents_of(&root, ".moai/linked/x/y") == Parents::Linked,
            "first link wins"
        );
        assert!(parents_of(&root, "top/a") == Parents::Linked, "top");
        assert!(parents_of(&root, ".moai/nowhere/a") == Parents::Gone);
        assert!(
            parents_of(&root, ".moai/file/a") == Parents::Gone,
            "a file is no folder"
        );
        // The last name is not looked at here: a link as the file itself is `check`'s business.
        assert!(parents_of(&root, ".moai/linked") == Parents::Real);
    }

    #[test]
    fn a_folder_that_is_no_git_work_tree_is_refused() {
        let d = tempfile::tempdir().unwrap();
        let err = check(d.path()).unwrap_err();
        assert!(err.starts_with("not a git work tree"), "{err}");
        assert_eq!(
            crate::contract::classify(&err),
            crate::contract::Kind::InvalidInput
        );
    }
}
