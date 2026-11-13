//! #2670: a project's `.worktreeinclude` that copies secret files into every new worktree.
//!
//! Claude Code copies the gitignored files a `.worktreeinclude` selects into every worktree it
//! creates with git — `--worktree`, subagent worktrees (`isolation: worktree`, as the base
//! agents `moai-impl` and `moai-audit` use) and desktop parallel sessions
//! (https://code.claude.com/docs/en/worktrees.md, read 2026-11-13). The Go generation's
//! template lists `.env` and `.env.*` there, so each worktree gets a copy of the project's
//! secrets. This module only reads and reports: the same docs show `.env` as an example, so a
//! line may be deliberate and is never edited here.
//!
//! The file uses `.gitignore` syntax. A line "selects a secret" when, after every later line
//! (including `!` negations) is applied, it is the line that selects one of
//! [`SECRET_PROBES`] — names a project keeps out of git because they hold credentials.

use std::{fs, path::Path};

/// The file Claude Code reads at the project root.
pub const FILE: &str = ".worktreeinclude";
/// Larger files are not read (a real include list is a few lines).
const LIMIT: u64 = 1024 * 1024;

/// Probe paths standing for secret files: environment files at the root and nested, private
/// keys and certificates, and credential or secret documents.
pub const SECRET_PROBES: [&str; 11] = [
    ".env",
    ".env.local",
    ".env.production",
    "apps/web/.env",
    "apps/web/.env.local",
    "id_rsa",
    "certs/server.key",
    "certs/server.pem",
    "credentials.json",
    "config/secrets.yml",
    ".aws/credentials",
];

/// One line that selects secret files.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Finding {
    /// 1-based line number.
    pub line: usize,
    /// The pattern as written (trimmed).
    pub pattern: String,
    /// The probe paths this line selects.
    pub probes: Vec<String>,
}

struct Rule {
    line: usize,
    text: String,
    negate: bool,
    dir_only: bool,
    anchored: bool,
    glob: String,
}

fn parse(text: &str) -> Vec<Rule> {
    let mut rules = Vec::new();
    for (i, raw) in text.lines().enumerate() {
        let line = raw.trim_end_matches('\r');
        // `.gitignore`: trailing spaces are ignored unless escaped; leading spaces count.
        let trimmed = line.trim_end();
        if trimmed.is_empty() || trimmed.starts_with('#') {
            continue;
        }
        let (negate, body) = match trimmed.strip_prefix('!') {
            Some(rest) => (true, rest),
            None => (false, trimmed),
        };
        // `\#` and `\!` stand for a literal first character.
        let body = body
            .strip_prefix('\\')
            .filter(|b| b.starts_with('#') || b.starts_with('!'))
            .unwrap_or(body);
        let dir_only = body.ends_with('/');
        let body = body.trim_end_matches('/');
        if body.is_empty() {
            continue;
        }
        let anchored = body.contains('/');
        let glob = body.trim_start_matches('/').to_string();
        rules.push(Rule {
            line: i + 1,
            text: trimmed.to_string(),
            negate,
            dir_only,
            anchored,
            glob,
        });
    }
    rules
}

/// `.gitignore` glob over a `/`-separated path: `**` crosses folders, `*` and `?` stay inside
/// one name, `[...]` is a character class.
fn glob_match(pattern: &[u8], text: &[u8]) -> bool {
    match pattern.first() {
        None => text.is_empty(),
        Some(b'*') if pattern.get(1) == Some(&b'*') => {
            // `**/` matches zero or more folders; a bare `**` matches everything left.
            let rest = &pattern[2..];
            let rest = rest.strip_prefix(b"/").unwrap_or(rest);
            (0..=text.len()).any(|i| {
                (i == 0 || text[i - 1] == b'/') && glob_match(rest, &text[i..]) || rest.is_empty()
            })
        }
        Some(b'*') => (0..=text.len())
            .take_while(|&i| i == 0 || text[i - 1] != b'/')
            .any(|i| glob_match(&pattern[1..], &text[i..])),
        Some(b'?') => !text.is_empty() && text[0] != b'/' && glob_match(&pattern[1..], &text[1..]),
        Some(b'[') => {
            let Some(end) = pattern.iter().skip(1).position(|&c| c == b']') else {
                return text.first() == Some(&b'[') && glob_match(&pattern[1..], &text[1..]);
            };
            let class = &pattern[1..1 + end];
            let (negated, class) = match class.first() {
                Some(b'!') | Some(b'^') => (true, &class[1..]),
                _ => (false, class),
            };
            let Some(&c) = text.first() else { return false };
            let mut hit = false;
            let mut j = 0;
            while j < class.len() {
                if j + 2 < class.len() && class[j + 1] == b'-' {
                    hit |= class[j] <= c && c <= class[j + 2];
                    j += 3;
                } else {
                    hit |= class[j] == c;
                    j += 1;
                }
            }
            hit != negated && c != b'/' && glob_match(&pattern[2 + end..], &text[1..])
        }
        Some(&p) => text.first() == Some(&p) && glob_match(&pattern[1..], &text[1..]),
    }
}

/// Whether `rule` selects the file at `path` (a `/`-separated path from the project root).
fn selects(rule: &Rule, path: &str) -> bool {
    let parts: Vec<&str> = path.split('/').collect();
    // A file is selected by a pattern naming it, or naming any folder above it.
    let candidates = (1..=parts.len()).map(|n| (parts[..n].join("/"), n == parts.len()));
    candidates.into_iter().any(|(prefix, is_file)| {
        if rule.dir_only && is_file {
            return false;
        }
        if rule.anchored {
            glob_match(rule.glob.as_bytes(), prefix.as_bytes())
        } else {
            let name = prefix.rsplit('/').next().unwrap_or(&prefix);
            glob_match(rule.glob.as_bytes(), name.as_bytes())
        }
    })
}

/// The lines of an include list that select a secret probe, in line order.
pub fn secret_lines(text: &str) -> Vec<Finding> {
    let rules = parse(text);
    let mut findings: Vec<Finding> = Vec::new();
    for probe in SECRET_PROBES {
        // The last matching line decides (`.gitignore` order); a negation un-selects.
        let Some(last) = rules.iter().rev().find(|r| selects(r, probe)) else {
            continue;
        };
        if last.negate {
            continue;
        }
        match findings.iter_mut().find(|f| f.line == last.line) {
            Some(f) => f.probes.push(probe.to_string()),
            None => findings.push(Finding {
                line: last.line,
                pattern: last.text.clone(),
                probes: vec![probe.to_string()],
            }),
        }
    }
    findings.sort_by_key(|f| f.line);
    findings
}

/// Read-only check of `<project>/.worktreeinclude`, as warning lines for a preview. A missing
/// file gives none; a link is never followed and a file that cannot be read is named, never
/// guessed. Nothing is written.
pub fn warnings(project: &Path) -> Vec<String> {
    let path = project.join(FILE);
    let meta = match fs::symlink_metadata(&path) {
        Ok(m) => m,
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => return Vec::new(),
        Err(_) => return vec![format!("{FILE} could not be inspected; check it by hand")],
    };
    if meta.file_type().is_symlink() {
        return vec![format!(
            "{FILE} is a symlink; not followed, check by hand whether it copies secret files into worktrees"
        )];
    }
    if !meta.is_file() || meta.len() > LIMIT {
        return vec![format!(
            "{FILE} is not a small regular file; not read, check it by hand"
        )];
    }
    let text = match fs::read(&path).map(String::from_utf8) {
        Ok(Ok(t)) => t,
        _ => {
            return vec![format!(
                "{FILE} is not readable UTF-8; not checked, check it by hand"
            )]
        }
    };
    secret_lines(&text)
        .into_iter()
        .map(|f| {
            format!(
                "{FILE} line {} `{}` copies secret files ({}) into every new Claude Code worktree; \
                 remove the line unless you mean it (not changed here, #2670)",
                f.line,
                f.pattern,
                f.probes.join(", ")
            )
        })
        .collect()
}

#[cfg(test)]
mod tests {
    use super::*;

    /// The Go generation's template lines (`main` `internal/template/templates/.worktreeinclude`
    /// at 95c7ad1bcf, lines 17–25), the negative control: it must be flagged.
    const GO_TEMPLATE: &str = "# Per-machine Claude Code settings.\n.claude/settings.local.json\n\n# Environment files\n.env\n.env.*\n";

    fn lines(text: &str) -> Vec<(usize, String)> {
        secret_lines(text)
            .into_iter()
            .map(|f| (f.line, f.pattern))
            .collect()
    }

    #[test]
    fn the_go_template_is_flagged_on_its_env_lines() {
        assert_eq!(
            lines(GO_TEMPLATE),
            vec![(5, ".env".to_string()), (6, ".env.*".to_string())]
        );
        let f = secret_lines(GO_TEMPLATE);
        assert!(f[0].probes.contains(&"apps/web/.env".to_string()), "{f:?}");
        assert!(f[1].probes.contains(&".env.local".to_string()), "{f:?}");
    }

    #[test]
    fn safe_lists_comments_and_look_alikes_are_not_flagged() {
        for text in [
            "",
            "# .env\n",
            ".claude/settings.local.json\n",
            ".envrc\n",
            "node_modules/\n",
            "env\n",
            "\\#.env\n",
        ] {
            assert_eq!(lines(text), Vec::new(), "{text:?}");
        }
    }

    #[test]
    fn negation_order_folders_globs_and_crlf_follow_gitignore() {
        // A later negation un-selects; an earlier one is overridden by a later pattern.
        assert_eq!(lines(".env*\n!.env*\n"), Vec::new());
        assert_eq!(lines("!.env\n.env\n"), vec![(2, ".env".to_string())]);
        // Globs and folders.
        assert_eq!(lines("*.env\n"), vec![(1, "*.env".to_string())]);
        assert_eq!(lines("**/.env\n"), vec![(1, "**/.env".to_string())]);
        assert_eq!(lines("certs/\n"), vec![(1, "certs/".to_string())]);
        assert_eq!(lines("*.pem\n"), vec![(1, "*.pem".to_string())]);
        assert_eq!(lines(".[e]nv\n"), vec![(1, ".[e]nv".to_string())]);
        // Anchored to the root: only the root file.
        let f = secret_lines("/.env\n");
        assert_eq!(f.len(), 1);
        assert_eq!(f[0].probes, vec![".env".to_string()]);
        // `*` stays inside one name: an anchored `/*.env` is only the root file.
        let f = secret_lines("/*.env\n");
        assert_eq!(f.len(), 1);
        assert_eq!(f[0].probes, vec![".env".to_string()]);
        // A folder pattern does not select a file of that name.
        assert_eq!(lines("id_rsa/\n"), Vec::new());
        // CRLF and trailing spaces.
        assert_eq!(lines(".env  \r\n"), vec![(1, ".env".to_string())]);
    }

    #[test]
    fn the_file_check_reads_only_and_never_follows_a_link() {
        let tmp = tempfile::tempdir().unwrap();
        let p = tmp.path();
        assert!(warnings(p).is_empty(), "no file, no warning");
        fs::write(p.join(FILE), GO_TEMPLATE).unwrap();
        let before = fs::read(p.join(FILE)).unwrap();
        let w = warnings(p);
        assert_eq!(w.len(), 2, "{w:?}");
        assert!(
            w[0].starts_with(".worktreeinclude line 5 `.env` copies secret files"),
            "{w:?}"
        );
        assert_eq!(fs::read(p.join(FILE)).unwrap(), before, "nothing written");
        fs::write(p.join(FILE), [0xff, 0xfe, b'\n']).unwrap();
        assert!(warnings(p)[0].contains("not readable UTF-8"));
        #[cfg(unix)]
        {
            fs::remove_file(p.join(FILE)).unwrap();
            let outside = tempfile::tempdir().unwrap();
            fs::write(outside.path().join("list"), GO_TEMPLATE).unwrap();
            std::os::unix::fs::symlink(outside.path().join("list"), p.join(FILE)).unwrap();
            let w = warnings(p);
            assert_eq!(w.len(), 1);
            assert!(w[0].contains("symlink; not followed"), "{w:?}");
        }
    }
}
