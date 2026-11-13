//! #2708: the PR body checker is `moai-cli pr-self-audit-check`, so the shipped skill never
//! asks a user project to run `tools/check_pr_self_audit.py`, which only this repository has.
//! Its rules are checked case by case against the Python tool by the three
//! `tests/check_pr_*_test.py` files (`tests/self_audit_parity.py`).
mod common;
use std::{fs, path::Path, process::Command};

fn skill_files(dir: &Path, out: &mut Vec<std::path::PathBuf>) {
    for e in fs::read_dir(dir).unwrap() {
        let p = e.unwrap().path();
        if p.is_dir() {
            skill_files(&p, out);
        } else {
            out.push(p);
        }
    }
}

#[test]
fn the_shipped_skill_never_runs_the_python_checker() {
    let mut files = Vec::new();
    skill_files(
        &Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai"),
        &mut files,
    );
    let hits: Vec<String> = files
        .iter()
        .filter(|p| p.extension().is_some_and(|e| e == "md"))
        .filter(|p| {
            fs::read_to_string(p)
                .unwrap()
                .contains("tools/check_pr_self_audit.py")
        })
        .map(|p| p.display().to_string())
        .collect();
    assert_eq!(hits, Vec::<String>::new());
    let audit = fs::read_to_string(
        Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai/workflows/audit.md"),
    )
    .unwrap();
    assert!(audit.contains("`moai pr-self-audit-check --reaudit --file BODY.md`"));
}

/// #2708 3/4 (default A after the 15:00Z deadline of 6854550310): `tools/ci_scope.py` stays
/// this repository's CI script, so every skill mention of it, or of its test, names the
/// moai-adk repository within the 300 bytes before it.
#[test]
fn every_ci_scope_mention_in_the_skill_names_this_repository() {
    let mut files = Vec::new();
    skill_files(
        &Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai"),
        &mut files,
    );
    let mut seen = 0;
    let mut bare = Vec::new();
    for p in files
        .iter()
        .filter(|p| p.extension().is_some_and(|e| e == "md"))
    {
        let text = fs::read_to_string(p).unwrap();
        for (at, _) in text.match_indices("ci_scope") {
            seen += 1;
            let mut from = at.saturating_sub(300);
            while !text.is_char_boundary(from) {
                from += 1;
            }
            if !text[from..at].contains("moai-adk repository") {
                bare.push(format!("{}: byte {at}", p.display()));
            }
        }
    }
    assert!(seen >= 3, "the three mentions moved: {seen}");
    assert_eq!(bare, Vec::<String>::new());
}

/// #2708 AC-1: a user project has no `tools/`, so the shipped skill never tells a session to
/// run a repository Python tool. No `python3 … tools/…` command, and every `tools/<name>.py`
/// mention names the moai-adk repository within the 300 bytes before it. The leader's scan
/// judge is the `moai-cli repo-scans` command (#2708 4/4).
#[test]
fn the_shipped_skill_runs_no_repository_python_tool() {
    let root = Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai");
    let mut files = Vec::new();
    skill_files(&root, &mut files);
    let mut mentions = 0;
    let mut bad = Vec::new();
    for p in files
        .iter()
        .filter(|p| p.extension().is_some_and(|e| e == "md"))
    {
        let text = fs::read_to_string(p).unwrap();
        for (at, _) in text.match_indices("python3") {
            let rest = &text[at..];
            let end = rest.find(['`', '\n']).unwrap_or(rest.len());
            if rest[..end].contains("tools/") {
                bad.push(format!("{}: byte {at}: runs {}", p.display(), &rest[..end]));
            }
        }
        for (at, _) in text.match_indices("tools/") {
            let name = &text[at + "tools/".len()..];
            let stem = name
                .find(|c: char| !(c.is_ascii_lowercase() || c.is_ascii_digit() || c == '_'))
                .unwrap_or(name.len());
            if stem == 0 || !name[stem..].starts_with(".py") {
                continue;
            }
            mentions += 1;
            let mut from = at.saturating_sub(300);
            while !text.is_char_boundary(from) {
                from += 1;
            }
            if !text[from..at].contains("moai-adk repository") {
                bad.push(format!(
                    "{}: byte {at}: bare tools/{}.py",
                    p.display(),
                    &name[..stem]
                ));
            }
        }
    }
    assert!(
        mentions >= 3,
        "the moai-adk repository mentions moved: {mentions}"
    );
    assert_eq!(bad, Vec::<String>::new());
    let leader = fs::read_to_string(root.join("workflows/mode-leader.md")).unwrap();
    assert!(leader.contains("run `moai repo-scans --refs <tip> <head>` from any checkout"));
    assert!(leader.contains("when `result.remerge` is `true`"));
}

fn run(args: &[&str], body: &str) -> (i32, String) {
    use std::io::Write;
    let mut child = Command::new(crate::common::MOAI_BIN)
        .arg("pr-self-audit-check")
        .args(args)
        .stdin(std::process::Stdio::piped())
        .stdout(std::process::Stdio::piped())
        .stderr(std::process::Stdio::piped())
        .spawn()
        .unwrap();
    child
        .stdin
        .take()
        .unwrap()
        .write_all(body.as_bytes())
        .unwrap();
    let out = child.wait_with_output().unwrap();
    (
        out.status.code().unwrap(),
        String::from_utf8_lossy(&out.stdout).into_owned(),
    )
}

const FULL: &str = "## Self-Audit\n\n- mutation: guard removed -> check FAILED\n- clone: fresh clone\n- host: fake gh\n- receipt: digest re-read\n";

/// The command's own boundaries, without Python: pass, a missing line, an unmeasured
/// mutation, a re-audit without its section, and a usage error.
#[test]
fn the_command_prints_one_line_and_the_python_exit_codes() {
    assert_eq!(
        run(&[], FULL),
        (
            0,
            "self-audit: pass (mutation/clone/host/receipt measured)\n".into()
        )
    );
    assert_eq!(
        run(&[], &FULL.replace("- receipt: digest re-read\n", "")),
        (1, "self-audit: blocked missing=receipt\n".into())
    );
    assert_eq!(
        run(&[], &FULL.replace("FAILED", "passed")),
        (1, "self-audit: blocked missing=mutation:FAILED\n".into())
    );
    assert_eq!(
        run(&["--reaudit"], FULL),
        (1, "self-audit: blocked missing=## 독립 review\n".into())
    );
    assert_eq!(run(&["bogus"], FULL).0, 2);
    assert_eq!(run(&["--reaudit", "--reaudit"], FULL).0, 2);
}
