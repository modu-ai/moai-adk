//! SPEC-20261121-skill-text-branch-pr-base: the skill that ships to user projects names no branch
//! of the form `card/<n>`, no pull request line `card: C-<n>` and no `--base v2026.11.0` (the
//! integration branch of moai-adk's own development). The claim step of `mode-agent.md` points at
//! the branch `moai spec take <SPEC-ID>` names and lets the project's own rule come first, its
//! pull-request step puts the SPEC id in the `Moai-Spec` trailer, and the pull-request step of
//! `sync.md` reads the base from the profile. The scan reads the files under `skills/` of this
//! checkout; nothing is run and no HOME is touched.

use std::{
    fs,
    path::{Path, PathBuf},
};

fn root() -> &'static Path {
    Path::new(env!("CARGO_MANIFEST_DIR"))
}

fn is_word(c: char) -> bool {
    c.is_ascii_alphanumeric() || c == '_'
}

/// The patterns one line carries: `card-branch` (`card/` followed by `<`, a digit, `{` or `C-`),
/// `card-line` (`card: C-` followed by `<` or a digit) and `fixed-base` (`--base` followed by
/// `v2026.11.0`, with a space or an equals sign). A word before `card` (`discard/`) is not it.
fn patterns(line: &str) -> Vec<&'static str> {
    let mut out = Vec::new();
    for (at, _) in line.match_indices("card/") {
        let before = line[..at].chars().next_back();
        if before.is_some_and(|c| is_word(c) || c == '-' || c == '/') {
            continue;
        }
        let rest = &line[at + "card/".len()..];
        let first = rest.chars().next();
        let card_dash = rest.starts_with("C-");
        if matches!(first, Some('<') | Some('{'))
            || first.is_some_and(|c| c.is_ascii_digit())
            || card_dash
        {
            out.push("card-branch");
        }
    }
    for (at, _) in line.match_indices("card: C-") {
        let before = line[..at].chars().next_back();
        if before.is_some_and(is_word) {
            continue;
        }
        let next = line[at + "card: C-".len()..].chars().next();
        if next.is_some_and(|c| c == '<' || c.is_ascii_digit()) {
            out.push("card-line");
        }
    }
    for sep in [' ', '='] {
        if line.contains(&format!("--base{sep}v2026.11.0")) {
            out.push("fixed-base");
        }
    }
    out
}

/// `path:line pattern` for every hit of `text`.
fn found(path: &str, text: &str) -> Vec<String> {
    let mut out = Vec::new();
    for (n, line) in text.lines().enumerate() {
        for p in patterns(line) {
            out.push(format!("{path}:{} {p}", n + 1));
        }
    }
    out
}

fn files_under(dir: &Path, out: &mut Vec<PathBuf>) {
    let Ok(entries) = fs::read_dir(dir) else {
        return;
    };
    for e in entries.flatten() {
        let p = e.path();
        if p.is_dir() {
            files_under(&p, out);
        } else {
            out.push(p);
        }
    }
}

/// The shipped files under `skills/` as (relative path, text); a file that is not UTF-8 text is
/// not read.
fn shipped() -> Vec<(String, String)> {
    let mut files = Vec::new();
    files_under(&root().join("skills"), &mut files);
    let mut out: Vec<(String, String)> = files
        .iter()
        .filter_map(|p| {
            let text = fs::read_to_string(p).ok()?;
            let rel = p
                .strip_prefix(root())
                .ok()?
                .components()
                .map(|c| c.as_os_str().to_string_lossy().into_owned())
                .collect::<Vec<_>>()
                .join("/");
            Some((rel, text))
        })
        .collect();
    out.sort();
    out
}

fn workflow(name: &str) -> String {
    fs::read_to_string(root().join("skills/moai/workflows").join(name)).unwrap()
}

/// The one line of a workflow that starts with `start` (a numbered step).
fn step(text: &str, start: &str) -> String {
    let lines: Vec<&str> = text.lines().filter(|l| l.starts_with(start)).collect();
    assert_eq!(lines.len(), 1, "exactly one line starts with `{start}`");
    lines[0].to_string()
}

/// AC-1: the scan sees each of the three patterns and leaves the allowed forms alone.
#[test]
fn scan_flags_each_pattern_and_leaves_allowed_forms_alone() {
    for (line, kind) in [
        (
            "on branch `card/<n>-<slug>` (`git worktree add -b card/<n>-<slug> p`)",
            "card-branch",
        ),
        ("`git branch -D card/12-fix-login`", "card-branch"),
        ("the branch card/{n}-x", "card-branch"),
        ("a name like card/C-12 is refused", "card-branch"),
        ("Second line `card: C-<n>`.", "card-line"),
        ("card: C-12", "card-line"),
        (
            "open it with `gh pr create --base v2026.11.0 --body-file <file>`",
            "fixed-base",
        ),
        ("`gh pr create --base=v2026.11.0`", "fixed-base"),
    ] {
        assert!(
            patterns(line).contains(&kind),
            "`{line}` should carry a {kind}: {:?}",
            patterns(line)
        );
    }
    for line in [
        "The branch is `spec/<slug>-<yyyymmdd>` and carries no card number.",
        "open it with `gh pr create --base <integration branch> --body-file <file>`",
        "discard/12 and a card/lane word are prose, and so is a card: of cards",
        "a card is taken with `moai card take`; the `Moai-Spec` trailer names the SPEC",
        "--base v2026.11.1 is another branch name",
        "a card: C-word and card/ alone are not numbers",
    ] {
        assert!(
            patterns(line).is_empty(),
            "`{line}` should be clean: {:?}",
            patterns(line)
        );
    }
}

/// AC-1: a hit is named by its file, its line and its pattern.
#[test]
fn scan_names_the_file_the_line_and_the_pattern() {
    let text = "clean line\nbranch `card/<n>-x` and `--base v2026.11.0`\ncard: C-5\n";
    assert_eq!(
        found("skills/x.md", text),
        [
            "skills/x.md:2 card-branch",
            "skills/x.md:2 fixed-base",
            "skills/x.md:3 card-line"
        ]
    );
}

/// AC-1: no shipped file under `skills/` carries one of the three patterns.
#[test]
fn scan_shipped_skills_carry_none() {
    let files = shipped();
    assert!(
        files
            .iter()
            .any(|(p, _)| p == "skills/moai/workflows/mode-agent.md"),
        "the scan reaches mode-agent.md"
    );
    assert!(
        files
            .iter()
            .any(|(p, _)| p == "skills/moai/workflows/sync.md"),
        "the scan reaches sync.md"
    );
    let mut bad = Vec::new();
    for (path, text) in &files {
        bad.extend(found(path, text));
    }
    assert!(
        bad.is_empty(),
        "shipped text names a card branch, a card line or a fixed base:\n{}",
        bad.join("\n")
    );
}

/// AC-2: step 5 gives the worktree place and the branch of `moai spec take`, passes that branch
/// to `card take`, and says the project's own rule comes first.
#[test]
fn claim_step_names_the_spec_branch_and_the_project_rule() {
    let claim = step(&workflow("mode-agent.md"), "5. **claim**");
    for phrase in [
        "~/.moai/worktrees/<project>/<name>/",
        "`moai spec take <SPEC-ID>`",
        "`spec/<slug>-<yyyymmdd>`",
        "--branch <branch>",
        "The project's own rule comes first",
        "`AGENTS.local.md`",
    ] {
        assert!(claim.contains(phrase), "step 5 must say `{phrase}`");
    }
}

/// AC-2: step 8 names the `Moai-Spec` trailer and carries no card line.
#[test]
fn claim_pr_step_puts_the_spec_id_in_the_trailer() {
    let gate = step(&workflow("mode-agent.md"), "8. **pr-gate**");
    assert!(
        gate.contains("`Moai-Spec` trailer"),
        "step 8 must name the `Moai-Spec` trailer"
    );
    assert!(
        !gate.contains("Second line"),
        "step 8 asks for no second body line"
    );
}

/// AC-3: the pull-request step of `sync.md` reads the base from the profile.
#[test]
fn sync_pull_request_step_reads_the_base_from_the_profile() {
    let pr = step(&workflow("sync.md"), "5. **Pull request.**");
    for phrase in [
        "--base <integration branch>",
        "`work_flow.integration_branch`",
    ] {
        assert!(
            pr.contains(phrase),
            "the pull-request step must say `{phrase}`"
        );
    }
}
