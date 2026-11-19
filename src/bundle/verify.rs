//! `moai bundle verify` (card C-119, SPEC-20261119-bundle-verify): the check list run in the
//! worktree of a bundle branch under one heavy slot, the run record `bundle body` reads, and the
//! heads named for each failure. It runs commands and writes one new folder (a script, a log per
//! command, `run.json`); it never fetches, pushes, opens a pull request, writes a card note or a
//! ledger row, changes a setting or moves a ref (REQ-9 of the SPEC, REQ-12 of the parent).

use super::{parse_head_args, tip_json, tip_of, Git, HeadArg, Result, CHECKS};
use serde_json::{json, Value};
use std::collections::BTreeSet;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};

/// What `verify` is given besides the heads (every option of the command).
pub struct Options<'a> {
    pub branch: &'a str,
    pub out: Option<&'a Path>,
    pub self_audit: Option<&'a Path>,
    pub moai_cli: Option<&'a Path>,
    pub python: &'a str,
    pub slot_dir: Option<&'a Path>,
    pub max: Option<u32>,
    pub min_free_gb: Option<u32>,
}

/// One command of the list: its check key (none for a changed test or a SPEC target), the name
/// a failure goes by, the words it runs, and whether it needs the built `moai-cli`.
struct Cmd {
    key: Option<&'static str>,
    label: String,
    words: Vec<String>,
    needs_binary: bool,
}

impl Cmd {
    fn text(&self) -> String {
        self.words
            .iter()
            .map(|w| quote(w))
            .collect::<Vec<_>>()
            .join(" ")
    }
}

/// A word as a shell reads it back: quoted only when it holds a character a shell treats as
/// more than a word.
fn quote(s: &str) -> String {
    let plain = !s.is_empty()
        && s.chars()
            .all(|c| c.is_ascii_alphanumeric() || "_@%+=:,./-".contains(c));
    if plain {
        s.to_string()
    } else {
        format!("'{}'", s.replace('\'', r"'\''"))
    }
}

fn cargo(args: &[&str]) -> Vec<String> {
    let mut w = vec!["cargo".to_string()];
    w.extend(args.iter().map(|a| a.to_string()));
    w
}

/// The `--test <name>` targets a `verify:` line of a SPEC names.
fn verify_targets(spec_text: &str) -> Vec<String> {
    let mut out = Vec::new();
    for line in spec_text.lines().filter(|l| l.contains("verify:")) {
        let mut words = line.split_whitespace();
        while let Some(w) = words.next() {
            if w == "--test" {
                if let Some(name) = words.next() {
                    let name = name.trim_matches(|c: char| !c.is_ascii_alphanumeric() && c != '_');
                    if !name.is_empty() {
                        out.push(name.to_string());
                    }
                }
            }
        }
    }
    out
}

/// `tests/<name>.rs` (a top-level file) as the cargo target `<name>`.
fn test_target(path: &str) -> Option<&str> {
    let name = path.strip_prefix("tests/")?.strip_suffix(".rs")?;
    (!name.is_empty() && !name.contains('/')).then_some(name)
}

/// The sums `counts` reads from a log: `N passed` over the `test result:` lines, else
/// `Ran N tests` of a unittest run.
fn counts(log: &str) -> Option<String> {
    let mut passed: Option<u64> = None;
    for l in log.lines().filter(|l| l.starts_with("test result:")) {
        let mut prev = "";
        for w in l.split_whitespace() {
            if w.trim_end_matches(';') == "passed" {
                if let Ok(n) = prev.parse::<u64>() {
                    passed = Some(passed.unwrap_or(0) + n);
                }
            }
            prev = w;
        }
    }
    if let Some(n) = passed {
        return Some(format!("{n} passed"));
    }
    for l in log.lines() {
        if let Some(rest) = l.strip_prefix("Ran ") {
            if let Some(n) = rest
                .split_whitespace()
                .next()
                .and_then(|n| n.parse::<u64>().ok())
            {
                return Some(format!("Ran {n} tests"));
            }
        }
    }
    None
}

/// Paths a failing output names that the branch changed: each whitespace-separated word, cut at
/// the first `:` (a `file:line:col`) and without a leading `<worktree>/` (`cargo fmt --check`
/// prints `Diff in <absolute path>:<line>:`), that is one of `changed`. `roots` are the spellings
/// of the worktree folder (as given and canonical).
fn named_paths(log: &str, changed: &BTreeSet<String>, roots: &[String]) -> BTreeSet<String> {
    let mut out = BTreeSet::new();
    for w in log.split_whitespace() {
        let w =
            w.trim_matches(|c: char| matches!(c, '"' | '\'' | '`' | '(' | ')' | ',' | '[' | ']'));
        let w = w.split(':').next().unwrap_or("");
        let w = roots
            .iter()
            .find_map(|r| w.strip_prefix(r.trim_end_matches('/'))?.strip_prefix('/'))
            .unwrap_or(w);
        let w = w.strip_prefix("./").unwrap_or(w);
        if changed.contains(w) {
            out.insert(w.to_string());
        }
    }
    out
}

/// The cards whose head's `tip...head` diff touches one of `paths`, in the order given; empty
/// when none does (REQ-7).
fn attributed(paths: &BTreeSet<String>, per_head: &[(String, BTreeSet<String>)]) -> Vec<String> {
    per_head
        .iter()
        .filter(|(_, p)| p.iter().any(|f| paths.contains(f)))
        .map(|(card, _)| card.clone())
        .collect()
}

/// `(path, HEAD sha, branch ref)` of every entry of `git worktree list --porcelain`.
fn worktree_entries(porcelain: &str) -> Vec<(PathBuf, String, Option<String>)> {
    let mut out = Vec::new();
    let mut cur: Option<(PathBuf, String, Option<String>)> = None;
    for l in porcelain.lines() {
        if let Some(p) = l.strip_prefix("worktree ") {
            if let Some(e) = cur.take() {
                out.push(e);
            }
            cur = Some((PathBuf::from(p), String::new(), None));
        } else if let Some(h) = l.strip_prefix("HEAD ") {
            if let Some(e) = cur.as_mut() {
                e.1 = h.to_string();
            }
        } else if let Some(b) = l.strip_prefix("branch ") {
            if let Some(e) = cur.as_mut() {
                e.2 = Some(b.to_string());
            }
        }
    }
    if let Some(e) = cur {
        out.push(e);
    }
    out
}

fn short(sha: &str) -> &str {
    sha.get(..10).unwrap_or(sha)
}

/// A file's text; a byte that is not UTF-8 becomes U+FFFD so a log with one never loses its counts
/// or its attribution.
fn read_text(path: &Path) -> Option<String> {
    Some(String::from_utf8_lossy(&std::fs::read(path).ok()?).into_owned())
}

/// A path as an absolute one: a relative path is the caller's, so it is read against the current
/// folder here, before anything runs in the worktree.
fn absolute(p: &Path) -> PathBuf {
    if p.is_absolute() {
        p.to_path_buf()
    } else {
        std::env::current_dir().map_or_else(|_| p.to_path_buf(), |d| d.join(p))
    }
}

/// The file stem of a command's log and exit status: its place in the list, then its label with
/// every character a file name or a shell could mind turned into `-`.
fn stem(i: usize, label: &str) -> String {
    let clean: String = label
        .chars()
        .map(|c| {
            if c.is_ascii_alphanumeric() || matches!(c, '_' | '.' | '-') {
                c
            } else {
                '-'
            }
        })
        .collect();
    format!("{:02}-{clean}", i + 1)
}

fn read_int(path: &Path) -> Option<i64> {
    read_text(path)?.trim().parse().ok()
}

/// The list as data: the ten keys of [`CHECKS`] in one order, with the commands of the table of
/// SPEC-20261119-bundle-verify, the targets of changed tests and of SPEC `verify:` lines in
/// place after `product_cli`, and one `result_coverage` per SPEC id. `not_run` gets the targets
/// that have no file.
fn the_list(
    cli: &str,
    python: &str,
    tip: &str,
    spec_ids: &[String],
    targets: &[String],
    exists: impl Fn(&str) -> bool,
) -> (Vec<Cmd>, Vec<(String, String)>) {
    let mut list: Vec<Cmd> = Vec::new();
    let mut not_run: Vec<(String, String)> = Vec::new();
    let plain = |key: &'static str, words: Vec<String>, needs_binary: bool| Cmd {
        key: Some(key),
        label: key.to_string(),
        words,
        needs_binary,
    };
    list.push(plain(
        "build",
        cargo(&["build", "--locked", "--offline", "--bin", "moai-cli"]),
        false,
    ));
    list.push(plain(
        "lib_tests",
        cargo(&["test", "--locked", "--offline", "--lib"]),
        false,
    ));
    for (key, target) in [
        ("home_isolation_scan", "home_isolation_scan"),
        ("product_cli", "product_cli"),
    ] {
        list.push(plain(
            key,
            cargo(&["test", "--locked", "--offline", "--test", target]),
            false,
        ));
    }
    for name in targets {
        if exists(name) {
            list.push(Cmd {
                key: None,
                label: format!("test:{name}"),
                words: cargo(&["test", "--locked", "--offline", "--test", name]),
                needs_binary: false,
            });
        } else {
            not_run.push((
                format!("test:{name}"),
                format!("a SPEC names the target `{name}` and there is no tests/{name}.rs in the branch"),
            ));
        }
    }
    list.push(plain("fmt", cargo(&["fmt", "--check"]), false));
    list.push(plain(
        "clippy",
        cargo(&[
            "clippy",
            "--locked",
            "--offline",
            "--all-targets",
            "--",
            "-D",
            "warnings",
        ]),
        false,
    ));
    let bin = |args: &[&str]| -> Vec<String> {
        let mut w = vec![cli.to_string()];
        w.extend(args.iter().map(|a| a.to_string()));
        w
    };
    list.push(plain(
        "trailer_check",
        bin(&["trailer-check", &format!("{tip}..HEAD")]),
        true,
    ));
    list.push(plain("spec_check", bin(&["spec", "check", "--all"]), true));
    if spec_ids.is_empty() {
        not_run.push((
            "result_coverage".to_string(),
            "the heads name no SPEC (no `Moai-Spec` trailer in the branch)".to_string(),
        ));
    }
    for id in spec_ids {
        list.push(Cmd {
            key: Some("result_coverage"),
            label: format!("result_coverage-{id}"),
            words: bin(&["spec", "check", id, "--closure"]),
            needs_binary: true,
        });
    }
    list.push(plain(
        "unittest",
        vec![
            python.to_string(),
            "-m".into(),
            "unittest".into(),
            "discover".into(),
            "-s".into(),
            "tests".into(),
            "-p".into(),
            "*_test.py".into(),
        ],
        false,
    ));
    (list, not_run)
}

/// The script: one block per command (its log, exit status, start and end), `.began` first so a
/// refused slot is told from a list that ran; a command that needs the built binary is skipped
/// with a reason when the build failed or the binary is not there (REQ-4).
fn script(worktree: &Path, out: &Path, bin: &str, list: &[Cmd]) -> String {
    let mut s = String::from("#!/bin/sh\n");
    s.push_str(&format!("OUT={}\n", quote(&out.to_string_lossy())));
    s.push_str(&format!("BIN={}\n", quote(bin)));
    s.push_str(&format!(
        "cd {} || exit 70\n: > \"$OUT/.began\"\n",
        quote(&worktree.to_string_lossy())
    ));
    for (i, c) in list.iter().enumerate() {
        let n = stem(i, &c.label);
        let run = format!(
            "date +%s > \"$OUT/{n}.start\"\n{cmd} > \"$OUT/{n}.log\" 2>&1\nprintf '%s' \"$?\" > \"$OUT/{n}.exit\"\ndate +%s > \"$OUT/{n}.end\"\n",
            cmd = c.text()
        );
        if c.needs_binary {
            s.push_str(&format!(
                "if [ \"$(cat \"$OUT/01-build.exit\" 2>/dev/null)\" != \"0\" ]; then\n  printf '%s' \"the build failed (exit $(cat \"$OUT/01-build.exit\" 2>/dev/null)), so there is no built moai-cli\" > \"$OUT/{n}.skip\"\n\
                 elif [ ! -x \"$BIN\" ]; then\n  printf '%s' \"no executable moai-cli at $BIN\" > \"$OUT/{n}.skip\"\nelse\n{run}fi\n"
            ));
        } else {
            s.push_str(&run);
        }
    }
    s
}

/// `bundle verify`.
pub fn verify(
    project: &Path,
    home: &Path,
    dir: &Path,
    tip: &str,
    heads: &[String],
    opts: &Options,
) -> Result<Value> {
    let args = parse_head_args(heads)?;
    let git = Git { dir };
    let tip = tip_of(&git, tip)?;
    let mut problems: Vec<String> = Vec::new();

    // REQ-2: the branch, its worktree, the heads, the folder and the file, every fact at once.
    let branch_ref = format!("refs/heads/{}", opts.branch);
    let pinned = git.commit(&branch_ref);
    if pinned.is_none() {
        problems.push(format!("the branch `{}` is not here", opts.branch));
    }
    let mut worktree: Option<PathBuf> = None;
    if let Some(pinned) = &pinned {
        let entries = git
            .ok(&["worktree", "list", "--porcelain"])
            .map(|o| worktree_entries(&o))
            .unwrap_or_default();
        match entries
            .iter()
            .find(|e| e.2.as_deref() == Some(branch_ref.as_str()))
        {
            None => problems.push(format!(
                "no worktree has the branch `{}` checked out (a worktree on another HEAD does not count)",
                opts.branch
            )),
            Some((path, head, _)) => {
                if head != pinned {
                    problems.push(format!(
                        "the HEAD of the worktree is {}, not the branch's SHA {}",
                        short(head),
                        short(pinned)
                    ));
                }
                match (Git { dir: path }).ok(&["status", "--porcelain", "--untracked-files=no"]) {
                    Some(s) if s.trim().is_empty() => {}
                    Some(_) => problems.push(format!(
                        "the worktree {} has uncommitted changes to tracked files",
                        path.display()
                    )),
                    None => problems.push(format!(
                        "git status cannot be read in the worktree {}",
                        path.display()
                    )),
                }
                worktree = Some(path.clone());
            }
        }
    }
    let mut wanted: Vec<(&HeadArg, String)> = Vec::new();
    for a in &args {
        let head = match &a.sha {
            Some(s) => Some(s.clone()),
            None => match super::card_notes(project, home, &a.card) {
                Ok(notes) => super::newest_marker(&a.card, &notes).and_then(|r| r.head),
                Err(e) => {
                    problems.push(format!("{}: {e}", a.card));
                    continue;
                }
            },
        };
        let Some(head) = head else {
            problems.push(format!(
                "{}: no readable LANE-READY-INTEGRATION line",
                a.card
            ));
            continue;
        };
        let Some(commit) = git.commit(&head) else {
            problems.push(format!(
                "{}: the head {} is not a commit here (fetch first)",
                a.card,
                short(&head)
            ));
            continue;
        };
        let in_tip = git.is_ancestor(&commit, &tip.sha);
        let in_branch = pinned
            .as_deref()
            .is_some_and(|p| git.is_ancestor(&commit, p));
        if !in_tip && !in_branch && pinned.is_some() {
            problems.push(format!(
                "{}: the head {} is in neither the tip nor the branch `{}`",
                a.card,
                short(&commit),
                opts.branch
            ));
        }
        if !in_tip {
            wanted.push((a, commit));
        }
    }
    let self_audit: Value = match opts.self_audit {
        None => json!({}),
        Some(p) => match read_text(p).and_then(|t| serde_json::from_str::<Value>(&t).ok()) {
            Some(v) if v.is_object() => v,
            _ => {
                problems.push(format!(
                    "--self-audit {} is not a readable JSON object",
                    p.display()
                ));
                json!({})
            }
        },
    };
    let (Some(pinned), Some(worktree)) = (pinned, worktree) else {
        return Err(refused(&problems));
    };
    let out_dir = match opts.out {
        Some(p) => absolute(p),
        None => worktree
            .join("target")
            .join("bundle-verify")
            .join(short(&pinned)),
    };
    if out_dir.exists() {
        problems.push(format!("the output folder {} exists", out_dir.display()));
    }
    if !problems.is_empty() {
        return Err(refused(&problems));
    }

    // The list: the changed tests and the SPEC targets, the SPEC ids, the binary.
    let changed: Vec<String> = git.changed_paths(&tip.sha, &pinned)?;
    let changed_set: BTreeSet<String> = changed.iter().cloned().collect();
    let mut targets: BTreeSet<String> = changed
        .iter()
        .filter_map(|p| test_target(p).map(String::from))
        .collect();
    for p in &changed {
        if p.starts_with(".moai/specs/") && p.ends_with("/spec.md") {
            if let Some(text) = read_text(&worktree.join(p)) {
                targets.extend(verify_targets(&text));
            }
        }
    }
    let targets: Vec<String> = targets.into_iter().collect();
    let spec_ids = git.spec_ids(&tip.sha, &pinned);
    let target_dir = std::env::var_os("CARGO_TARGET_DIR")
        .map(|d| worktree.join(d))
        .unwrap_or_else(|| worktree.join("target"));
    let bin: String = match opts.moai_cli {
        Some(p) => absolute(p).to_string_lossy().into_owned(),
        None => target_dir
            .join("debug")
            .join("moai-cli")
            .to_string_lossy()
            .into_owned(),
    };
    let (list, mut not_run) = the_list(&bin, opts.python, &tip.name, &spec_ids, &targets, |n| {
        worktree.join("tests").join(format!("{n}.rs")).is_file()
    });
    debug_assert!(CHECKS
        .iter()
        .all(|k| list.iter().any(|c| c.key == Some(*k)) || not_run.iter().any(|n| n.0 == *k)));
    std::fs::create_dir_all(&out_dir).map_err(|e| {
        format!(
            "bundle verify refused: cannot create the output folder {}: {e}",
            out_dir.display()
        )
    })?;
    let script_path = out_dir.join("run.sh");
    std::fs::write(&script_path, script(&worktree, &out_dir, &bin, &list)).map_err(|e| {
        format!(
            "bundle verify refused: cannot write {}: {e}",
            script_path.display()
        )
    })?;

    // One slot hold for the whole list: this executable's own `heavy-slot-acquire`.
    let exe = std::env::current_exe().map_err(|e| format!("cannot find this executable: {e}"))?;
    let mut slot = Command::new(exe);
    slot.arg("heavy-slot-acquire");
    if let Some(n) = opts.max {
        slot.args(["--max", &n.to_string()]);
    }
    if let Some(n) = opts.min_free_gb {
        slot.args(["--min-free-gb", &n.to_string()]);
    }
    if let Some(d) = opts.slot_dir {
        slot.arg("--dir").arg(absolute(d));
    }
    slot.args(["--name", "bundle-verify", "--", "sh"])
        .arg(&script_path)
        .current_dir(&worktree)
        .stdin(Stdio::null())
        .stdout(Stdio::null())
        .stderr(Stdio::piped());
    let o = slot
        .output()
        .map_err(|e| format!("bundle verify refused: cannot start the heavy slot: {e}"))?;
    if !out_dir.join(".began").exists() {
        let why = String::from_utf8_lossy(&o.stderr)
            .lines()
            .map(str::trim)
            .rfind(|l| !l.is_empty())
            .unwrap_or("no message")
            .to_string();
        let _ = std::fs::remove_dir_all(&out_dir);
        return Err(format!(
            "bundle verify refused: the heavy slot was not obtained or the script did not start (exit {}): {why}; nothing was recorded",
            o.status.code().map_or("signal".to_string(), |c| c.to_string())
        ));
    }

    // The record: what each command did, the failures with their heads.
    let per_head: Vec<(String, BTreeSet<String>)> = wanted
        .iter()
        .map(|(a, head)| {
            let paths = git
                .changed_paths(&tip.sha, head)
                .unwrap_or_default()
                .into_iter()
                .collect();
            (a.card.clone(), paths)
        })
        .collect();
    let mut roots = vec![worktree.to_string_lossy().into_owned()];
    if let Ok(c) = std::fs::canonicalize(&worktree) {
        roots.push(c.to_string_lossy().into_owned());
    }
    let mut commands: Vec<Value> = Vec::new();
    let mut failed: Vec<Value> = Vec::new();
    for (i, c) in list.iter().enumerate() {
        let n = stem(i, &c.label);
        if let Some(reason) = read_text(&out_dir.join(format!("{n}.skip"))) {
            not_run.push((c.key.unwrap_or(&c.label).to_string(), reason));
            continue;
        }
        let log_name = format!("{n}.log");
        let log = read_text(&out_dir.join(&log_name)).unwrap_or_default();
        let Some(exit) = read_int(&out_dir.join(format!("{n}.exit"))) else {
            not_run.push((
                c.key.unwrap_or(&c.label).to_string(),
                "the command did not run (no exit status was recorded)".to_string(),
            ));
            continue;
        };
        let mut entry = json!({
            "command": c.text(), "exit": exit,
            "started": read_int(&out_dir.join(format!("{n}.start"))).unwrap_or(0),
            "ended": read_int(&out_dir.join(format!("{n}.end"))).unwrap_or(0),
            "log": log_name,
        });
        if let Some(k) = c.key {
            entry["check"] = json!(k);
        }
        if let Some(cn) = counts(&log) {
            entry["counts"] = json!(cn);
        }
        commands.push(entry);
        if exit != 0 {
            let mut paths = named_paths(&log, &changed_set, &roots);
            if let Some(name) = c.label.strip_prefix("test:") {
                let file = format!("tests/{name}.rs");
                if changed_set.contains(&file) {
                    paths.insert(file);
                }
            }
            let mut heads = attributed(&paths, &per_head);
            if heads.is_empty() {
                heads.push("unattributed".to_string());
            }
            failed.push(json!({
                "check": c.label, "command": c.text(), "exit": exit,
                "log": format!("{n}.log"), "heads": heads,
            }));
        }
    }
    let not_run_json: Vec<Value> = not_run
        .iter()
        .map(|(k, r)| json!({"check": k, "reason": r}))
        .collect();
    let record = json!({
        "commands": commands, "not_run": not_run_json, "self_audit": self_audit,
        "pinned": pinned, "branch": opts.branch, "tip": tip.sha,
    });
    let record_path = out_dir.join("run.json");
    let mut text = serde_json::to_string_pretty(&record).map_err(|e| e.to_string())?;
    text.push('\n');
    std::fs::write(&record_path, text).map_err(|e| {
        format!(
            "bundle verify refused: cannot write {}: {e}",
            record_path.display()
        )
    })?;
    Ok(json!({
        "action": "verify", "branch": opts.branch, "pinned": pinned,
        "worktree": worktree.to_string_lossy(), "tip": tip_json(&tip),
        "out": out_dir.to_string_lossy(), "record": record_path.to_string_lossy(),
        "commands": commands, "not_run": not_run_json, "failed": failed,
        "ok": failed.is_empty(), "writes": true,
        "note": "a command that cannot run is in not_run, never a failure; `body` refuses a record with a failed command; the Self-Audit lines are the leader's (--self-audit)",
    }))
}

fn refused(problems: &[String]) -> String {
    format!(
        "bundle verify refused: {} problem(s), nothing was run or written: {}",
        problems.len(),
        problems.join("; ")
    )
}

#[cfg(test)]
mod tests {
    use super::*;

    fn set(items: &[&str]) -> BTreeSet<String> {
        items.iter().map(|s| s.to_string()).collect()
    }

    #[test]
    fn counts_sum_passed_lines_and_read_unittest() {
        let log = "test result: ok. 3 passed; 0 failed\ntest result: ok. 4 passed; 1 failed";
        assert_eq!(counts(log).as_deref(), Some("7 passed"));
        assert_eq!(
            counts("Ran 12 tests in 0.5s\n\nOK").as_deref(),
            Some("Ran 12 tests")
        );
        assert_eq!(counts("error: boom"), None);
    }

    #[test]
    fn paths_are_the_words_of_an_output_that_the_branch_changed() {
        let changed = set(&["src/a.rs", "tests/alpha.rs"]);
        let log = "error: x\n --> src/a.rs:12:5\nDiff in ./tests/alpha.rs:3:\n(src/other.rs)";
        let roots = vec!["/wt".to_string(), "/private/wt/".to_string()];
        assert_eq!(
            named_paths(log, &changed, &roots),
            set(&["src/a.rs", "tests/alpha.rs"])
        );
        assert_eq!(named_paths("nothing here", &changed, &roots), set(&[]));
        // `cargo fmt --check` names the absolute path, in either spelling of the worktree.
        let fmt = "Diff in /wt/src/a.rs:1:\nDiff in /private/wt/tests/alpha.rs:3:";
        assert_eq!(
            named_paths(fmt, &changed, &roots),
            set(&["src/a.rs", "tests/alpha.rs"])
        );
        assert_eq!(
            named_paths("Diff in /other/src/a.rs:1:", &changed, &roots),
            set(&[])
        );
    }

    #[test]
    fn a_stem_keeps_only_name_characters_and_the_place() {
        assert_eq!(stem(0, "build"), "01-build");
        assert_eq!(stem(9, "rc-SPEC-1\"x $(y)"), "10-rc-SPEC-1-x---y-");
        assert_eq!(stem(2, "test:a/b"), "03-test-a-b");
        assert!(absolute(Path::new("x")).is_absolute());
    }

    #[test]
    fn heads_are_named_by_their_own_diff_in_the_order_given() {
        let per = vec![
            ("C-1".to_string(), set(&["a.rs", "shared"])),
            ("C-2".to_string(), set(&["b.rs", "shared"])),
        ];
        assert_eq!(attributed(&set(&["shared"]), &per), vec!["C-1", "C-2"]);
        assert_eq!(attributed(&set(&["b.rs"]), &per), vec!["C-2"]);
        assert!(attributed(&set(&["c.rs"]), &per).is_empty());
    }

    #[test]
    fn spec_targets_come_from_verify_lines_only() {
        let text = "- AC-1 (REQ-1): x.\n  verify: `cargo test --locked --offline --test alpha`; `cargo test --test beta -- filter`\n  manual: cargo test --test gamma\n";
        assert_eq!(verify_targets(text), vec!["alpha", "beta"]);
        assert_eq!(test_target("tests/alpha.rs"), Some("alpha"));
        assert_eq!(test_target("tests/common/mod.rs"), None);
        assert_eq!(test_target("src/alpha.rs"), None);
    }

    #[test]
    fn the_list_has_the_ten_checks_in_one_order() {
        let (list, not_run) = the_list(
            "moai-cli",
            "python3",
            "origin/x",
            &["SPEC-1".to_string()],
            &[],
            |_| true,
        );
        let mut keys: Vec<&str> = Vec::new();
        for c in &list {
            if let Some(k) = c.key {
                if !keys.contains(&k) {
                    keys.push(k);
                }
            }
        }
        assert_eq!(keys, CHECKS.to_vec());
        assert!(not_run.is_empty());
        assert_eq!(
            list.last().unwrap().text(),
            "python3 -m unittest discover -s tests -p '*_test.py'"
        );
    }

    #[test]
    fn worktree_entries_read_branch_and_head() {
        let p =
            "worktree /a\nHEAD aaaa\nbranch refs/heads/main\n\nworktree /b\nHEAD bbbb\ndetached\n";
        let e = worktree_entries(p);
        assert_eq!(e.len(), 2);
        assert_eq!(e[0].2.as_deref(), Some("refs/heads/main"));
        assert_eq!(e[1].2, None);
        assert_eq!(e[1].1, "bbbb");
    }
}
