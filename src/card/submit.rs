//! `moai card submit <card> --generation <g>`: the whole hand-in of a lane in one command (card
//! C-503, NeXT-Flow 2 step 1).
//!
//! In this order, and the ledger is written only after every step before it passed:
//!
//! 1. the preconditions: the caller holds the card at the generation, the reserved worktree is
//!    clean, its HEAD is the tip of the reserved branch, no criterion is still `open`;
//! 2. the fast gates, all of them, so one refusal lists every failure: `diff_check`, `fmt`,
//!    `ops_rustfmt`, `tables` and `trailers`. A failing gate starts no long step;
//! 3. the receipt: a PASS summary of the same head and base is quoted, otherwise
//!    `moai check affected --card <base>..<head>` runs once;
//! 4. the READY note, built from the receipt and written through `card note`, so the receipt
//!    gate and the `leftovers:` rule judge it as they judge a hand-written one;
//! 5. `card review` (state `review`, the reserved paths given back).
//!
//! A path outside the reservation's scope is never refused: it is the `outside-scope:` line of
//! the note, which the auditor reads. The tier is the floor of the changed paths, the rule of
//! `moai bundle check`; a tier 2 head needs the lane's `guard-mutant:` lines (`--mutants FILE`).
//! The command writes no verdict and lands nothing.

use super::input;
use super::land::check::floor_tier;
use super::landed;
use super::work;
use crate::check_affected;
use crate::ledger::reservation;
use crate::ledger::workflow::covers;
use crate::ledger::{self, LedgerError};
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};

type Result<T> = std::result::Result<T, String>;

/// What the caller typed (and the tests' check program, never a flag).
pub struct SubmitRequest<'a> {
    pub card: &'a str,
    /// This session's id: the holder the reservation is fenced to.
    pub holder: &'a str,
    pub generation: i64,
    /// `--mutants FILE`: the lane's `guard-mutant:` lines, needed for a tier 2 head.
    pub mutants: Option<&'a Path>,
    /// A program that stands in for `moai check affected` (a debug build's test seam): run as
    /// `<program> <worktree> <base> <head> <out folder>`.
    pub check_cmd: Option<PathBuf>,
}

/// The most bytes of a `--mutants` file.
const MAX_MUTANTS_BYTES: u64 = 64 * 1024;

/// One criterion as the ledger holds it.
struct Criterion {
    seq: i64,
    text: String,
    status: String,
    evidence: Option<String>,
    reason: Option<String>,
}

/// One fast gate's result.
struct Gate {
    label: &'static str,
    /// `None` when the gate had nothing to check.
    code: Option<i32>,
    log: PathBuf,
}

fn short(rev: &str) -> &str {
    rev.get(..12).unwrap_or(rev)
}

/// One git command in `dir`: its trimmed standard output, or its failure in words.
fn git_out(dir: &Path, args: &[&str]) -> Result<String> {
    let out = landed::git(dir, args)?;
    if out.status.success() {
        Ok(String::from_utf8_lossy(&out.stdout).trim_end().to_string())
    } else {
        Err(format!(
            "git {} failed: {}",
            args.first().copied().unwrap_or(""),
            String::from_utf8_lossy(&out.stderr).trim()
        ))
    }
}

/// `card submit`: see the module text.
pub fn submit(project: &Path, home: &Path, req: &SubmitRequest<'_>) -> Result<Value> {
    let refuse = |why: String| format!("card submit refused: {why}");
    let card_id = match input::parse_ref(req.card).map_err(|e| refuse(e.to_string()))? {
        input::CardRef::Id(n) => n,
        input::CardRef::SourceNumber(_) => {
            return Err(refuse(
                "name the card by its C-number (`card show #N` finds it)".to_string(),
            ))
        }
    };
    let name = input::card_name(card_id);

    // ---- 1. what the ledger holds, read once ------------------------------------------------
    let (held, criteria) = read_card(project, home, card_id)?;
    if held.holder != req.holder || held.generation != req.generation || held.status != "held" {
        return Err(refuse(format!(
            "{name} is not held by this session at generation {} (it is held by {} at generation {}, \
             status {}); `card show {name}` reads the reservation",
            req.generation, held.holder, held.generation, held.status
        )));
    }
    let worktree = PathBuf::from(
        held.worktree
            .clone()
            .ok_or_else(|| refuse(format!("{name}'s reservation names no worktree")))?,
    );
    let branch = held
        .branch
        .clone()
        .ok_or_else(|| refuse(format!("{name}'s reservation names no branch")))?;
    if !worktree.is_dir() {
        return Err(refuse(format!(
            "the reserved worktree {} is not a folder",
            worktree.display()
        )));
    }

    // ---- 2. preconditions, listed together --------------------------------------------------
    let mut problems: Vec<String> = Vec::new();
    let status = git_out(
        &worktree,
        &["status", "--porcelain=v1", "--untracked-files=all"],
    )
    .map_err(refuse)?;
    if !status.is_empty() {
        let listed: Vec<&str> = status.lines().take(20).collect();
        problems.push(format!(
            "the worktree is not clean ({} path(s); commit or remove them): {}",
            status.lines().count(),
            listed.join(" | ")
        ));
    }
    let head = git_out(&worktree, &["rev-parse", "HEAD"]).map_err(refuse)?;
    let on = git_out(&worktree, &["symbolic-ref", "--short", "-q", "HEAD"]).unwrap_or_default();
    let tip =
        git_out(&worktree, &["rev-parse", &format!("refs/heads/{branch}")]).unwrap_or_default();
    if on != branch || tip != head {
        problems.push(format!(
            "HEAD ({}) is not the tip of the reserved branch {branch} (on `{on}`, tip {})",
            short(&head),
            short(&tip)
        ));
    }
    for c in criteria.iter().filter(|c| c.status == "open") {
        problems.push(format!(
            "criterion {} is still open (\"{}\"); mark it with `card criterion {name} {} --generation {} --met --evidence REF` or withdraw it",
            c.seq, c.text, c.seq, req.generation
        ));
    }
    if criteria.iter().all(|c| c.status != "met") && problems.is_empty() {
        problems.push("no criterion of the card is met".to_string());
    }
    if !problems.is_empty() {
        return Err(refuse(format!(
            "the card is not ready to hand in (ledger unchanged):\n  - {}",
            problems.join("\n  - ")
        )));
    }

    // ---- 3. the range, the tier and the paths outside the scope ------------------------------
    let integration = landed::integration_branch(project).map_err(refuse)?;
    let base = git_out(&worktree, &["merge-base", &integration, &head]).map_err(|e| {
        refuse(format!(
            "no merge base of {integration} and {} ({e})",
            short(&head)
        ))
    })?;
    let range = format!("{base}..{head}");
    let changed = changed_paths(&worktree, &base, &head).map_err(refuse)?;
    let outside: Vec<String> = changed
        .iter()
        .filter(|p| !held.write_scope.iter().any(|s| covers(s, p)))
        .cloned()
        .collect();
    let tier = floor_tier(project, &base, &head, 0).map_err(refuse)?;
    let mutants = read_mutants(req.mutants, tier).map_err(refuse)?;

    // ---- 4. the fast gates, every one, before any long step ---------------------------------
    let logs = worktree
        .join("target")
        .join(format!("submit-{}", short(&head)));
    fs::create_dir_all(&logs)
        .map_err(|e| refuse(format!("cannot create {}: {e}", logs.display())))?;
    let gates = fast_gates(&worktree, &range, &changed, &branch, &logs);
    let failed: Vec<&Gate> = gates
        .iter()
        .filter(|g| g.code.is_some_and(|c| c != 0))
        .collect();
    if !failed.is_empty() {
        let lines: Vec<String> = failed
            .iter()
            .map(|g| {
                format!(
                    "{} exit {} (log {})",
                    g.label,
                    g.code.unwrap_or(1),
                    g.log.display()
                )
            })
            .collect();
        return Err(refuse(format!(
            "{} fast gate(s) failed; no long step was started and the ledger is unchanged:\n  - {}",
            failed.len(),
            lines.join("\n  - ")
        )));
    }

    // ---- 5. the receipt: quote a PASS of this head and base, or run the check once ---------
    let receipt_dir = worktree
        .join("target")
        .join(format!("check-affected-{}", short(&head)));
    let summary_path = receipt_dir.join(super::ready_receipt::SUMMARY_FILE);
    let (summary, reused) = match pass_summary(&summary_path, &head, &base) {
        Some(line) => (line, true),
        None => {
            let code = run_check(&worktree, &base, &head, &receipt_dir, &logs, req)?;
            match pass_summary(&summary_path, &head, &base) {
                Some(line) => (line, false),
                None => {
                    let said = fs::read_to_string(&summary_path)
                        .map(|s| s.trim().to_string())
                        .unwrap_or_else(|_| "no summary was written".to_string());
                    return Err(refuse(format!(
                        "the check did not end in result=PASS; the ledger is unchanged\n  - summary: {said}\n  - check program exit {code} (the summary decides)\n  - log folder: {}",
                        receipt_dir.display()
                    )));
                }
            }
        }
    };

    // ---- 6. the READY note, then the hand-in -------------------------------------------------
    let text = ready_text(&NoteParts {
        card: &name,
        head: &head,
        base: &base,
        branch: &branch,
        tier,
        summary: &summary,
        reused,
        gates: &gates,
        receipt_dir: &receipt_dir,
        criteria: &criteria,
        outside: &outside,
        mutants: &mutants,
    });
    let noted = work::note(
        project,
        home,
        &name,
        &work::NoteRequest {
            kind: "report",
            text: text.as_bytes(),
            reference: None,
            by: None,
            session: Some(req.holder),
        },
    )?;
    let moved = work::review_keeping(project, home, &name, req.holder, req.generation, &[])?;
    let next = work::next(project, home, Some(req.holder))
        .ok()
        .and_then(|v| v.get("next").cloned())
        .unwrap_or(Value::Null);
    Ok(json!({
        "action": "submit", "name": name, "card_id": card_id,
        "head": head, "base": base, "tier": tier, "state": "review",
        "note_id": noted["note_id"],
        "receipt": {"reused": reused, "summary": summary, "folder": receipt_dir},
        "outside_scope": outside,
        "given_back": moved["given_back"],
        "next": next,
        "writes": true,
    }))
}

/// The reservation the card holds and its criteria, from a read-only ledger.
fn read_card(
    project: &Path,
    home: &Path,
    card_id: i64,
) -> Result<(reservation::Reservation, Vec<Criterion>)> {
    let id = ledger::command::project_id_of(project)?;
    let ledger = match ledger::open_read_only(home, &id) {
        Ok(l) => l,
        Err(LedgerError::Absent) => {
            return Err("card submit refused: the project has no ledger".to_string())
        }
        Err(e) => return Err(e.to_string()),
    };
    let held = reservation::current(ledger.conn(), card_id)
        .map_err(|e| e.to_string())?
        .ok_or_else(|| {
            format!(
                "card submit refused: {} holds no reservation",
                input::card_name(card_id)
            )
        })?;
    let mut st = ledger
        .conn()
        .prepare(
            "SELECT seq, text, status, evidence_ref, withdrawn_reason FROM card_ac \
             WHERE card_id = ?1 ORDER BY seq",
        )
        .map_err(|e| e.to_string())?;
    let rows = st
        .query_map([card_id], |r| {
            Ok(Criterion {
                seq: r.get(0)?,
                text: r.get(1)?,
                status: r.get(2)?,
                evidence: r.get(3)?,
                reason: r.get(4)?,
            })
        })
        .map_err(|e| e.to_string())?
        .collect::<std::result::Result<Vec<_>, _>>()
        .map_err(|e| e.to_string())?;
    Ok((held, rows))
}

/// The paths `base..head` changes (new names of renames included), sorted.
fn changed_paths(dir: &Path, base: &str, head: &str) -> Result<Vec<String>> {
    let out = git_out(
        dir,
        &[
            "diff",
            "--name-only",
            "--no-renames",
            &format!("{base}..{head}"),
        ],
    )?;
    let mut paths: Vec<String> = out.lines().map(str::to_string).collect();
    paths.sort();
    paths.dedup();
    Ok(paths)
}

/// The `guard-mutant:` lines of `--mutants`: required for tier 2, read for any tier when given,
/// and every line of the file must be one such line.
fn read_mutants(file: Option<&Path>, tier: u8) -> Result<Vec<String>> {
    let Some(file) = file else {
        if tier >= 2 {
            return Err(format!(
                "the head is tier {tier} (a path under the tier 2 floor): pass `--mutants FILE` with one \
                 `guard-mutant: <test> | <code removed> | exit <n>` line per new guard test, or \
                 `guard-mutant: none | <reason>`"
            ));
        }
        return Ok(Vec::new());
    };
    let meta =
        fs::metadata(file).map_err(|e| format!("cannot read --mutants {}: {e}", file.display()))?;
    if !meta.is_file() || meta.len() > MAX_MUTANTS_BYTES {
        return Err(format!(
            "--mutants {} is not a regular file of at most {MAX_MUTANTS_BYTES} bytes",
            file.display()
        ));
    }
    let text = fs::read_to_string(file)
        .map_err(|e| format!("cannot read --mutants {}: {e}", file.display()))?;
    let mut lines = Vec::new();
    for line in text.lines().map(str::trim).filter(|l| !l.is_empty()) {
        if !line.starts_with("guard-mutant: ") {
            return Err(format!(
                "--mutants holds a line that is not `guard-mutant: <test> | <code removed> | exit <n>`: {line}"
            ));
        }
        lines.push(line.to_string());
    }
    if lines.is_empty() && tier >= 2 {
        return Err("--mutants holds no `guard-mutant:` line".to_string());
    }
    Ok(lines)
}

/// Runs one program with its output in `log`; its exit code (127 when it cannot start).
fn run_logged(dir: &Path, program: &str, args: &[String], log: &Path, unset: &[&str]) -> i32 {
    let Ok(file) = fs::File::create(log) else {
        return 127;
    };
    let Ok(err) = file.try_clone() else {
        return 127;
    };
    let mut cmd = Command::new(program);
    if program == "git" {
        crate::identity::scrub_git_env(&mut cmd);
    }
    cmd.args(args)
        .current_dir(dir)
        .stdin(Stdio::null())
        .stdout(file)
        .stderr(err);
    for var in unset {
        cmd.env_remove(var);
    }
    match cmd.status() {
        Ok(status) => status.code().unwrap_or(1),
        Err(_) => 127,
    }
}

fn strings(args: &[&str]) -> Vec<String> {
    args.iter().map(|s| s.to_string()).collect()
}

/// This program, to start `moai trailer-check` and `moai check affected` as the user would.
fn this_program() -> Result<String> {
    std::env::current_exe()
        .map_err(|e| format!("cannot find this program to run a check: {e}"))
        .map(|p| p.to_string_lossy().into_owned())
}

/// Every fast gate of the range, run to the end (each result is kept, none stops the others).
fn fast_gates(
    worktree: &Path,
    range: &str,
    changed: &[String],
    branch: &str,
    logs: &Path,
) -> Vec<Gate> {
    let log = |n: usize, label: &str| logs.join(format!("{n:02}-{label}.log"));
    let mut gates = Vec::new();

    // diff_check: whitespace errors the range adds.
    let path = log(1, "diff_check");
    let code = match landed::git(worktree, &["diff", "--check", range]) {
        Ok(out) => {
            let mut text = out.stdout.clone();
            text.extend_from_slice(&out.stderr);
            let _ = fs::write(&path, text);
            out.status.code().unwrap_or(1)
        }
        Err(e) => {
            let _ = fs::write(&path, e);
            127
        }
    };
    gates.push(Gate {
        label: "diff_check",
        code: Some(code),
        log: path,
    });

    // fmt: only a Cargo project has one.
    let path = log(2, "fmt");
    let code = worktree.join("Cargo.toml").is_file().then(|| {
        run_logged(
            worktree,
            "cargo",
            &strings(&["fmt", "--all", "--", "--check"]),
            &path,
            &[],
        )
    });
    gates.push(Gate {
        label: "fmt",
        code,
        log: path,
    });

    // ops_rustfmt: `cargo fmt` does not reach `src/ops` as written in this project.
    let path = log(3, "ops_rustfmt");
    let ops: Vec<String> = changed
        .iter()
        .filter(|p| p.starts_with("src/ops/") && p.ends_with(".rs") && worktree.join(p).is_file())
        .cloned()
        .collect();
    let code = (!ops.is_empty()).then(|| {
        let mut args = strings(&["--edition", "2021", "--check"]);
        args.extend(ops);
        run_logged(worktree, "rustfmt", &args, &path, &[])
    });
    gates.push(Gate {
        label: "ops_rustfmt",
        code,
        log: path,
    });

    // tables: an added top-level test file must be in the area table and the allow list.
    let path = log(4, "tables");
    let code = table_gate(worktree, range, &path);
    gates.push(Gate {
        label: "tables",
        code,
        log: path,
    });

    // trailers: every commit of the range carries the identity trailers in one final block.
    let path = log(5, "trailers");
    let code = match this_program() {
        Ok(me) => run_logged(
            worktree,
            &me,
            &strings(&[
                "trailer-check",
                range,
                "--dir",
                &worktree.to_string_lossy(),
                "--branch",
                branch,
            ]),
            &path,
            &[],
        ),
        Err(e) => {
            let _ = fs::write(&path, e);
            127
        }
    };
    gates.push(Gate {
        label: "trailers",
        code: Some(code),
        log: path,
    });
    gates
}

/// The table check of `check affected`'s `test_tables` step, for a range that adds a top-level
/// `tests/*.rs` file: `None` when it adds none; otherwise 0 or 1 with the verdict lines in `log`.
fn table_gate(worktree: &Path, range: &str, log: &Path) -> Option<i32> {
    let added = git_out(
        worktree,
        &[
            "diff",
            "--name-only",
            "--diff-filter=A",
            "--no-renames",
            range,
        ],
    )
    .ok()?;
    let adds_test = added.lines().any(|p| {
        p.strip_prefix("tests/")
            .and_then(|r| r.strip_suffix(".rs"))
            .is_some_and(|s| !s.contains('/'))
    });
    if !adds_test {
        return None;
    }
    let lines = match check_affected::read_facts(worktree, range) {
        Ok(facts) => {
            let mut lines = check_affected::table_violations(
                &facts.added,
                &facts.tests,
                facts.areas_text.as_deref(),
                facts.allow_text.as_deref(),
                &facts.moved,
            );
            lines.extend(check_affected::scan_violations(
                &facts.added,
                &facts.tests,
                facts.scan_list.as_deref(),
                facts.tracked.as_deref(),
            ));
            for (table, error) in &facts.unreadable {
                lines.push(format!(
                    "{table}: the table exists but cannot be read ({error})"
                ));
            }
            lines
        }
        Err(e) => vec![format!("the range cannot be read: {e}")],
    };
    let _ = fs::write(log, lines.join("\n"));
    Some(i32::from(!lines.is_empty()))
}

/// The summary line of a PASS receipt of this head and base, or `None`: the file is one line,
/// names this head and this base, is no `--lane` run and ends in `result=PASS`. The receipt gate
/// of `card note` judges the same file again when the note is written, so a stricter reading
/// there only makes the check run.
fn pass_summary(path: &Path, head: &str, base: &str) -> Option<String> {
    let meta = fs::symlink_metadata(path).ok()?;
    if !meta.is_file() || meta.len() > 4096 {
        return None;
    }
    let text = fs::read_to_string(path).ok()?;
    let line = text.strip_suffix('\n').unwrap_or(&text);
    if line.contains('\n') {
        return None;
    }
    let words: Vec<&str> = line.split(' ').collect();
    let named = |key: &str| words.iter().find_map(|w| w.strip_prefix(key));
    let same_base = named("base=").is_some_and(|b| b.len() >= 12 && base.starts_with(b));
    (words.first() == Some(&"check-affected")
        && named("head=") == Some(short(head))
        && same_base
        && !words.contains(&"selection=lane")
        && words.last() == Some(&"result=PASS"))
    .then(|| line.to_string())
}

/// Runs the long check once: the test seam when the caller gave one, else `moai check affected
/// --card <range>` with the session id unset (as the rules say). Its exit code.
fn run_check(
    worktree: &Path,
    base: &str,
    head: &str,
    receipt_dir: &Path,
    logs: &Path,
    req: &SubmitRequest<'_>,
) -> Result<i32> {
    let log = logs.join("06-check_affected.log");
    Ok(match &req.check_cmd {
        Some(program) => run_logged(
            worktree,
            &program.to_string_lossy(),
            &[
                worktree.to_string_lossy().into_owned(),
                base.to_string(),
                head.to_string(),
                receipt_dir.to_string_lossy().into_owned(),
            ],
            &log,
            &[],
        ),
        None => {
            let me = this_program()?;
            run_logged(
                worktree,
                &me,
                &[
                    "check".to_string(),
                    "affected".to_string(),
                    "--card".to_string(),
                    format!("{base}..{head}"),
                    "--dir".to_string(),
                    worktree.to_string_lossy().into_owned(),
                ],
                &log,
                &["CLAUDE_CODE_SESSION_ID"],
            )
        }
    })
}

/// What the READY note is made of.
struct NoteParts<'a> {
    card: &'a str,
    head: &'a str,
    base: &'a str,
    branch: &'a str,
    tier: u8,
    summary: &'a str,
    reused: bool,
    gates: &'a [Gate],
    receipt_dir: &'a Path,
    criteria: &'a [Criterion],
    outside: &'a [String],
    mutants: &'a [String],
}

/// The READY note. The first line is the submission marker the receipt gate reads; the last is
/// the one `leftovers:` line (the tree was clean, and `target/` is expected state).
fn ready_text(p: &NoteParts<'_>) -> String {
    let mut out = String::new();
    out.push_str(&format!(
        "LANE-READY-INTEGRATION {} head={}\n",
        p.card, p.head
    ));
    out.push_str(&format!(
        "base: {}\nbranch: {}\ntier: {}\n",
        p.base, p.branch, p.tier
    ));
    out.push_str(&format!(
        "receipt ({}): {}\nlog folder: {}\n",
        if p.reused {
            "quoted, not run again"
        } else {
            "run by submit"
        },
        p.summary,
        p.receipt_dir.display()
    ));
    let gates: Vec<String> = p
        .gates
        .iter()
        .map(|g| match g.code {
            Some(c) => format!("{}={c}", g.label),
            None => format!("{}=skipped", g.label),
        })
        .collect();
    out.push_str(&format!("fast gates: {}\n", gates.join(" ")));
    out.push_str("criteria:\n");
    for c in p.criteria {
        match c.status.as_str() {
            "met" => out.push_str(&format!(
                "  {} met: {}\n",
                c.seq,
                c.evidence.as_deref().unwrap_or("")
            )),
            _ => out.push_str(&format!(
                "  {} {}: {}\n",
                c.seq,
                c.status,
                c.reason.as_deref().unwrap_or(&c.text)
            )),
        }
    }
    if p.outside.is_empty() {
        out.push_str("outside-scope: none\n");
    } else {
        out.push_str(&format!("outside-scope: {}\n", p.outside.join(", ")));
    }
    for m in p.mutants {
        out.push_str(m);
        out.push('\n');
    }
    out.push_str("leftovers: none\n");
    out
}

#[cfg(test)]
mod tests {
    use super::*;

    /// The test seam for the long check is read by a debug build only: the line that reads
    /// `MOAI_TEST_SUBMIT_CHECK` stands under `cfg(debug_assertions)` and a release build takes
    /// `None`, so no program named by the environment is ever started by a release binary.
    #[test]
    fn the_check_program_seam_is_read_by_debug_builds_only() {
        let source = include_str!("../ops/card.rs");
        let reads: Vec<usize> = source
            .lines()
            .enumerate()
            .filter(|(_, l)| l.contains("env(\"MOAI_TEST_SUBMIT_CHECK\")"))
            .map(|(i, _)| i)
            .collect();
        assert_eq!(reads.len(), 1, "the seam is read at one place: {reads:?}");
        let lines: Vec<&str> = source.lines().collect();
        let before = lines[..reads[0]]
            .iter()
            .rev()
            .map(|l| l.trim())
            .find(|l| !l.is_empty() && !l.starts_with("//"));
        assert_eq!(before, Some("#[cfg(debug_assertions)]"));
        let after = lines[reads[0] + 1..]
            .iter()
            .map(|l| l.trim())
            .find(|l| !l.is_empty() && !l.starts_with("//"));
        assert_eq!(after, Some("#[cfg(not(debug_assertions))]"));
        assert!(
            lines[reads[0] + 2].contains("let submit_check: Option<PathBuf> = None;"),
            "a release build takes no check program: {}",
            lines[reads[0] + 2]
        );
    }

    #[test]
    fn a_receipt_is_quoted_only_for_this_head_and_base_and_a_pass() {
        let dir = tempfile::tempdir().unwrap();
        let path = dir.path().join("summary.txt");
        let head = "0123456789abcdef0123456789abcdef01234567";
        let base = "fedcba9876543210fedcba9876543210fedcba98";
        let line = |extra: &str, result: &str| {
            format!(
                "check-affected base={} head={} targets=3 scans=listed:0{extra} test-threads=5 diff_check=0 fmt=0 clippy=0 test=0 result={result}\n",
                &base[..12],
                &head[..12]
            )
        };
        fs::write(&path, line(" selection=card", "PASS")).unwrap();
        assert!(pass_summary(&path, head, base).is_some());
        fs::write(&path, line("", "FAIL")).unwrap();
        assert!(
            pass_summary(&path, head, base).is_none(),
            "a FAIL is no receipt"
        );
        fs::write(&path, line(" selection=lane", "PASS")).unwrap();
        assert!(
            pass_summary(&path, head, base).is_none(),
            "a lane run is no receipt"
        );
        fs::write(&path, line("", "PASS")).unwrap();
        assert!(pass_summary(&path, head, "1111111111112222222222223333333333334444").is_none());
        assert!(
            pass_summary(&path, "aaaaaaaaaaaabcdef0123456789abcdef01234567", base).is_none(),
            "another head"
        );
    }
}
