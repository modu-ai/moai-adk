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
//! [`SECRET_PROBES`] — names a project keeps out of git because they hold credentials — or a
//! path the file itself names whose last name is a secret name ([`secret_name`]: the `.env`
//! family, keys and certificates, credential files), so `.env.development` written out is
//! caught too (pre-audit FAIL 6851377127). Names compare without case, as git does on the
//! default macOS file system (`core.ignorecase`).

use std::{fs, path::Path};

/// The file Claude Code reads at the project root.
pub const FILE: &str = ".worktreeinclude";
/// Larger files are not read (a real include list is a few lines).
const LIMIT: u64 = 1024 * 1024;

/// Probe paths standing for secret files: environment files at the root and nested, private
/// keys and certificates, and credential or secret documents.
pub const SECRET_PROBES: [&str; 18] = [
    ".env",
    ".env.local",
    ".env.production",
    ".env.development",
    "apps/web/.env",
    "apps/web/.env.local",
    "id_rsa",
    "certs/server.key",
    "certs/server.pem",
    "certs/client.p12",
    "config/master.key",
    "credentials.json",
    "service-account.json",
    "config/secrets.yml",
    "secrets/token",
    ".aws/credentials",
    ".npmrc",
    ".netrc",
];

/// Example files a project keeps in git on purpose: never a secret name.
const EXAMPLE_NAMES: [&str; 4] = [".env.example", ".env.sample", ".env.template", ".env.dist"];

/// Whether a file name (no folder) is one a project keeps out of git because it holds
/// credentials: the `.env` family (not its example files), private keys and certificates, and
/// the usual credential files. Compared without case.
pub fn secret_name(name: &str) -> bool {
    let n = name.to_ascii_lowercase();
    if EXAMPLE_NAMES.contains(&n.as_str()) {
        return false;
    }
    n == ".env"
        || n.starts_with(".env.")
        || [".key", ".pem", ".p12", ".pfx"]
            .iter()
            .any(|ext| n.len() > ext.len() && n.ends_with(ext))
        || matches!(
            n.as_str(),
            ".npmrc"
                | ".netrc"
                | ".pgpass"
                | "credentials"
                | "credentials.json"
                | "service-account.json"
                | "id_rsa"
                | "id_ecdsa"
                | "id_ed25519"
        )
}

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

/// Whether `rule`'s pattern matches `path` itself (a `/`-separated path from the project root;
/// `is_file` false for a folder). Folders above `path` are judged by [`deciding_rule`].
fn matches_at(rule: &Rule, path: &str, is_file: bool) -> bool {
    if rule.dir_only && is_file {
        return false;
    }
    // Without case: git on the default macOS file system (`core.ignorecase`) matches `.ENV`
    // to `.env`, so reading case strictly would miss a line that copies it.
    let glob = rule.glob.to_ascii_lowercase();
    let path = path.to_ascii_lowercase();
    if rule.anchored {
        glob_match(glob.as_bytes(), path.as_bytes())
    } else {
        let name = path.rsplit('/').next().unwrap_or(&path);
        glob_match(glob.as_bytes(), name.as_bytes())
    }
}

/// The line that decides whether the file at `path` is selected, as git does (audit FAIL
/// 6851346525 B1): the folders above it are judged first, from the top; once a folder is
/// selected (its last matching line is not a `!`), that line decides — "it is not possible to
/// re-include a file if a parent directory of that file is excluded" (gitignore docs), so a
/// later `!` for the file does not count. Otherwise the last line matching the file decides.
fn deciding_rule<'a>(rules: &'a [Rule], path: &str) -> Option<&'a Rule> {
    let parts: Vec<&str> = path.split('/').collect();
    for n in 1..parts.len() {
        let folder = parts[..n].join("/");
        if let Some(r) = rules.iter().rev().find(|r| matches_at(r, &folder, false)) {
            if !r.negate {
                return Some(r);
            }
        }
    }
    rules.iter().rev().find(|r| matches_at(r, path, true))
}

/// The lines of an include list that select a secret probe, in line order.
pub fn secret_lines(text: &str) -> Vec<Finding> {
    let rules = parse(text);
    // The fixed probes, plus every path a selecting line names outright whose last name is a
    // secret name (`.env.development`, `config/master.key`): a list that writes the family
    // out name by name copies the same secrets as `.env.*`.
    let mut probes: Vec<String> = SECRET_PROBES.iter().map(|p| p.to_string()).collect();
    for rule in rules.iter().filter(|r| !r.negate && !r.dir_only) {
        let path = rule.glob.trim_start_matches("**/");
        let last = path.rsplit('/').next().unwrap_or(path);
        let literal = !path.contains(['*', '?', '[']);
        if literal && secret_name(last) && !probes.iter().any(|p| p.eq_ignore_ascii_case(path)) {
            probes.push(path.to_string());
        }
    }
    let mut findings: Vec<Finding> = Vec::new();
    for probe in probes.iter().map(String::as_str) {
        // A selected folder above the file decides; otherwise the last matching line does,
        // and a negation un-selects (`.gitignore` order).
        let Some(last) = deciding_rule(&rules, probe) else {
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

    /// Pre-audit FAIL 6851377127 (agent-25) B1, N1, N2: the `.env` family written out name by
    /// name, keys and the usual credential files are caught; example files are not; case does
    /// not matter.
    #[test]
    fn named_secret_files_keys_and_any_case_are_flagged_but_examples_are_not() {
        for line in [
            ".env.development",
            ".env.staging",
            ".env.test",
            "config/master.key",
            "deploy/prod.pem",
            "*.p12",
            "secrets/",
            ".npmrc",
            ".netrc",
            "service-account.json",
            "**/.env.staging",
            ".ENV",
            ".Env.Local",
        ] {
            assert_eq!(
                lines(&format!("{line}\n")),
                vec![(1, line.to_string())],
                "{line:?}"
            );
        }
        for line in [
            ".env.example",
            ".env.sample",
            ".env.template",
            ".env.dist",
            "README.md",
        ] {
            assert_eq!(lines(&format!("{line}\n")), Vec::new(), "{line:?}");
        }
        // A named secret that a later line negates is not flagged.
        assert_eq!(lines(".env.staging\n!.env.staging\n"), Vec::new());
        assert!(secret_name(".env.ci") && secret_name("server.KEY"));
        assert!(!secret_name(".key") && !secret_name(".envrc") && !secret_name("env"));
    }

    /// Audit FAIL 6851346525 (agent-29) B1: a `!` cannot re-include a file whose folder is
    /// selected, as in git. The rows are git's own answers (`git ls-files -o -i
    /// --exclude-from=list`, measured in the audit): `certs/` then `!` lines still copies both
    /// files, while `certs/*` then `!certs/server.key` copies only the `.pem`.
    #[test]
    fn a_negation_under_a_selected_folder_does_not_re_include_like_git() {
        let probes = |text: &str| -> Vec<(usize, Vec<String>)> {
            secret_lines(text)
                .into_iter()
                .map(|f| (f.line, f.probes))
                .collect()
        };
        let both = vec![
            "certs/server.key".to_string(),
            "certs/server.pem".to_string(),
            "certs/client.p12".to_string(),
        ];
        assert_eq!(
            probes("certs/\n!certs/server.key\n"),
            vec![(1, both.clone())]
        );
        assert_eq!(
            probes("certs/\n!certs/server.key\n!certs/server.pem\n"),
            vec![(1, both)]
        );
        assert_eq!(
            probes("certs/*\n!certs/server.key\n"),
            vec![(
                1,
                vec![
                    "certs/server.pem".to_string(),
                    "certs/client.p12".to_string()
                ]
            )]
        );
        // A negated folder is not selected: the file's own lines decide below it.
        assert_eq!(probes("certs/\n!certs/\n"), Vec::new());
        // Audit 6851794046 N3: below a negated folder a later file line still selects, and
        // that line (3) is the one reported, not the negated folder (git: `certs/server.key`).
        assert_eq!(
            probes("certs/\n!certs/\ncerts/server.key\n"),
            vec![(3, vec!["certs/server.key".to_string()])]
        );
        // Non-blocking N2: `?` stays inside one name (anchored, so the whole path is matched).
        assert_eq!(lines("/certs?server.key\n"), Vec::new());
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
        // Audit 6851346525 N1: the size limit at its edge — exactly LIMIT bytes is read,
        // one byte more is named and not read.
        let mut at_limit = b".env\n".to_vec();
        at_limit.resize(LIMIT as usize - 1, b'#');
        at_limit.push(b'\n');
        assert_eq!(at_limit.len() as u64, LIMIT);
        fs::write(p.join(FILE), &at_limit).unwrap();
        let w = warnings(p);
        assert!(w[0].starts_with(".worktreeinclude line 1 `.env`"), "{w:?}");
        at_limit.push(b'\n');
        fs::write(p.join(FILE), &at_limit).unwrap();
        assert!(warnings(p)[0].contains("not a small regular file"));
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
