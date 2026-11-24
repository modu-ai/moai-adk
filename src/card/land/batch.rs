//! `moai card land-batch` (SPEC-20261124-card-select-batch-land): 2 to 4 audited heads are joined
//! one after another onto the integration tip, the wide check runs once on the last join in a
//! private worktree, and each head lands through [`super::land`] in the given order.
//!
//! The loop, so that one wide run serves as many heads as it can and no result outlives its tip:
//!
//! 1. the tip is read and the pending heads are joined onto it in order (a head the tip already
//!    holds or stands on needs no join; a head that cannot be joined is dropped);
//! 2. the check runs on the last join. On PASS the heads land in order, each by its own
//!    compare-and-swap; the tip has moved, so what is still pending is joined and checked again.
//!    A tip that moved between the run and the landing sends the same set through step 1 again;
//! 3. on FAIL a set of one head drops that head; a larger set is cut to its first half and that
//!    half runs on its own. A passing half lands; the rest is checked again on the new tip.
//!
//! Each head is judged by the landing itself (verdict, scans, stale tip): a refusal stops the
//! batch at that head and the heads after it are reported as not run. A join candidate is landed
//! with its audited head named, so the verdict of the audited head covers it.

use super::LandRequest;
use crate::card::input::{self, CardRef};
use crate::card::landed::{self, Containment, Standing};
use crate::ledger::workflow::LEADER_BY;
use serde_json::{json, Value};
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};

type Result<T> = std::result::Result<T, String>;

/// The most heads one batch lands (the card says 2 to 4).
pub const MAX_HEADS: usize = 4;

/// What `card land-batch` was given.
pub struct BatchRequest<'a> {
    pub session: Option<&'a str>,
    pub by: &'a str,
    /// Each `C-n=<head>`: a card and its audited head (40 lowercase hex).
    pub pairs: &'a [String],
    /// The check program: it gets `<tip>..<join>` as its one argument, runs in a private worktree
    /// at the join and passes with exit 0. `None` runs the wide `check affected` in this process.
    pub check_cmd: Option<&'a Path>,
}

#[derive(Clone, Debug)]
struct Item {
    name: String,
    head: String,
}

/// One head of a set after the join: the commit that lands for it.
struct Candidate {
    item: Item,
    commit: String,
    joined: bool,
}

fn refused(why: impl std::fmt::Display) -> String {
    format!("card land-batch refused: {why}; nothing was changed")
}

fn parse_pairs(pairs: &[String]) -> Result<Vec<Item>> {
    if !(2..=MAX_HEADS).contains(&pairs.len()) {
        return Err(refused(format!(
            "give 2 to {MAX_HEADS} `C-n=<head>` pairs, got {}",
            pairs.len()
        )));
    }
    let mut items: Vec<Item> = Vec::new();
    for pair in pairs {
        let Some((card, head)) = pair.split_once('=') else {
            return Err(refused(format!("`{pair}` is not `C-n=<40 hex head>`")));
        };
        if !crate::gate::is_full_sha(head) {
            return Err(refused(format!(
                "the head of `{pair}` must be 40 lowercase hex"
            )));
        }
        let name = match input::parse_ref(card).map_err(refused)? {
            CardRef::Id(n) => input::card_name(n),
            CardRef::SourceNumber(_) => {
                return Err(refused("name each card by its C-number"));
            }
        };
        if items.iter().any(|i| i.name == name || i.head == head) {
            return Err(refused(format!("`{pair}` repeats a card or a head")));
        }
        items.push(Item {
            name,
            head: head.to_string(),
        });
    }
    Ok(items)
}

fn short(sha: &str) -> &str {
    sha.get(..12).unwrap_or(sha)
}

/// The trailers of a join commit: the lander's session and no verification kind.
fn trailers(project: &Path, session: Option<&str>) -> String {
    let mut text = String::new();
    if let (Some(id), Some(key)) = (session, crate::trailers::worktree_key(project)) {
        text.push_str(&format!("Moai-Session: {id}:{key}\n"));
    }
    text.push_str("Moai-Verified: none\n");
    text
}

/// Join `items` in order onto `tip`. A head the cursor already holds or stands on lands as it is;
/// any other is merged with `git merge-tree --write-tree` and a merge commit that has the cursor
/// as its first parent. The heads that cannot be joined come back with the reason.
fn join(
    project: &Path,
    session: Option<&str>,
    branch: &str,
    tip: &str,
    items: &[Item],
) -> Result<(Vec<Candidate>, Vec<(Item, String)>)> {
    let mut cursor = tip.to_string();
    let mut out = Vec::new();
    let mut bad = Vec::new();
    for item in items {
        match landed::contained(project, &item.head, &cursor)? {
            Containment::NotACommit => {
                bad.push((
                    item.clone(),
                    format!(
                        "not_a_commit: {} is not a commit of this repository",
                        item.head
                    ),
                ));
                continue;
            }
            Containment::Inside => {
                out.push(Candidate {
                    item: item.clone(),
                    commit: item.head.clone(),
                    joined: false,
                });
                continue;
            }
            Containment::Outside => {}
        }
        if landed::standing(project, &item.head, &cursor)? == Standing::OnTip {
            cursor = item.head.clone();
            out.push(Candidate {
                item: item.clone(),
                commit: item.head.clone(),
                joined: false,
            });
            continue;
        }
        let tree = match super::check::git_text(
            project,
            &[
                "merge-tree",
                "--write-tree",
                "--no-messages",
                &cursor,
                &item.head,
            ],
        ) {
            Ok(text) => text.lines().next().unwrap_or("").trim().to_string(),
            Err(why) => {
                bad.push((item.clone(), format!("join_failed: {why}")));
                continue;
            }
        };
        let message = format!(
            "merge(next-flow): {} into {branch}, head {}\n\n{}",
            item.name,
            short(&item.head),
            trailers(project, session)
        );
        let commit = super::check::git_text(
            project,
            &[
                "commit-tree",
                &tree,
                "-p",
                &cursor,
                "-p",
                &item.head,
                "-m",
                &message,
            ],
        )
        .map_err(|why| {
            format!(
                "join_failed: the merge commit of {} cannot be made: {why}",
                item.name
            )
        })?
        .trim()
        .to_string();
        cursor = commit.clone();
        out.push(Candidate {
            item: item.clone(),
            commit,
            joined: true,
        });
    }
    Ok((out, bad))
}

/// What one check run gave.
struct Run {
    exit: i32,
    log: PathBuf,
}

/// Run the check on `last` in a private worktree under `<project>/target/land-batch/`, which the
/// run removes again. `Err` means the check could not start; an exit code other than 0 is a FAIL.
fn run_check(
    project: &Path,
    check: Option<&Path>,
    tip: &str,
    last: &str,
    number: usize,
) -> Result<Run> {
    let base = project.join("target").join("land-batch");
    std::fs::create_dir_all(&base)
        .map_err(|e| format!("check_not_started: cannot create {}: {e}", base.display()))?;
    let work = base.join(format!("{}-{number}", short(last)));
    let log = base.join(format!("{}-{number}.log", short(last)));
    let work_text = work.to_string_lossy().to_string();
    super::check::git_text(project, &["worktree", "add", "--detach", &work_text, last])
        .map_err(|why| format!("check_not_started: the private worktree cannot be made: {why}"))?;
    let range = format!("{tip}..{last}");
    let result = match check {
        Some(program) => {
            let file = std::fs::File::create(&log)
                .map_err(|e| format!("check_not_started: cannot write {}: {e}", log.display()));
            match file {
                Err(e) => Err(e),
                Ok(file) => {
                    let again = file
                        .try_clone()
                        .map_err(|e| format!("check_not_started: {e}"));
                    match again {
                        Err(e) => Err(e),
                        Ok(again) => Command::new(program)
                            .arg(&range)
                            .current_dir(&work)
                            .stdin(Stdio::null())
                            .stdout(file)
                            .stderr(again)
                            .status()
                            .map(|s| s.code().unwrap_or(1))
                            .map_err(|e| {
                                format!("check_not_started: {} cannot run: {e}", program.display())
                            }),
                    }
                }
            }
        }
        None => {
            let opts = crate::check_affected::RunOptions {
                slot_dir: None,
                max: 4,
                min_free_gb: 50,
                out: None,
                pick: crate::check_affected::Pick::Full,
                integration_ref: None,
            };
            match crate::check_affected::run(&work, &range, &opts) {
                Ok(answer) => {
                    let _ = std::fs::write(&log, format!("{}\n", answer["summary"]));
                    Ok(i32::from(answer["result"] != "PASS"))
                }
                Err(why) => Err(format!("check_not_started: {why}")),
            }
        }
    };
    // The worktree is the run's own: git removes it with its checkout.
    let _ = super::check::git_text(project, &["worktree", "remove", "--force", &work_text]);
    let _ = super::check::git_text(project, &["worktree", "prune"]);
    result.map(|exit| Run { exit, log })
}

fn entry(item: &Item, why: Option<&str>) -> Value {
    match why {
        Some(why) => json!({"card": item.name, "head": item.head, "reason": why}),
        None => json!({"card": item.name, "head": item.head}),
    }
}

/// `card land-batch`: see the module comment.
pub fn land_batch(project: &Path, home: &Path, req: &BatchRequest<'_>) -> Result<Value> {
    if req.by != LEADER_BY {
        return Err(refused(format!(
            "only the leader lands a card until the typed verdict exists: give --by {LEADER_BY}"
        )));
    }
    let mut pending = parse_pairs(req.pairs)?;
    let branch = landed::integration_branch(project).map_err(|why| {
        refused(format!(
            "a landing is made on the integration branch, and the profile gives none ({why})"
        ))
    })?;

    let mut landed_cards: Vec<Value> = Vec::new();
    let mut dropped: Vec<Value> = Vec::new();
    let mut runs: Vec<Value> = Vec::new();
    let mut stopped: Value = Value::Null;
    let mut not_run: Vec<Value> = Vec::new();

    'outer: while !pending.is_empty() {
        let mut take = pending.len();
        loop {
            let tip = landed::tip_of(project, &branch).map_err(refused)?;
            let set: Vec<Item> = pending[..take].to_vec();
            let (cands, bad) = join(project, req.session, &branch, &tip, &set)?;
            if !bad.is_empty() {
                for (item, why) in &bad {
                    dropped.push(entry(item, Some(why)));
                    pending.retain(|p| p.name != item.name);
                }
                continue 'outer;
            }
            let last = cands.last().map(|c| c.commit.clone()).unwrap_or_default();
            // Heads the tip already holds need no check; a set of them lands as it is.
            let moves = cands.iter().any(|c| landed_moves(project, &c.commit, &tip));
            let exit = if moves {
                match run_check(project, req.check_cmd, &tip, &last, runs.len() + 1) {
                    Ok(run) => {
                        runs.push(json!({
                            "cards": set.iter().map(|i| i.name.clone()).collect::<Vec<_>>(),
                            "range": format!("{tip}..{last}"),
                            "exit": run.exit,
                            "log": run.log.to_string_lossy(),
                        }));
                        run.exit
                    }
                    Err(why) => {
                        stopped = json!({"code": "check_not_started", "why": why});
                        break 'outer;
                    }
                }
            } else {
                0
            };
            if exit != 0 {
                if take == 1 {
                    let item = pending.remove(0);
                    dropped.push(entry(
                        &item,
                        Some(&format!(
                            "check_failed: the check exited {exit} on the join of this head alone"
                        )),
                    ));
                    continue 'outer;
                }
                take = take.div_ceil(2);
                continue;
            }
            // A tip that moved since the run has no say in this result: join and check again.
            if landed::tip_of(project, &branch).map_err(refused)? != tip {
                continue;
            }
            for cand in &cands {
                let answer = super::land(
                    project,
                    home,
                    &cand.item.name,
                    &LandRequest {
                        session: req.session,
                        head: &cand.commit,
                        met: &[],
                        generation: None,
                        by: req.by,
                        keep_open: false,
                        audited: cand.joined.then_some(cand.item.head.as_str()),
                    },
                );
                match answer {
                    Ok(done) => {
                        let mut e = entry(&cand.item, None);
                        e["candidate"] = json!(cand.commit);
                        e["moved"] = done["moved"].clone();
                        e["closed"] = done["closed"].clone();
                        landed_cards.push(e);
                        pending.retain(|p| p.name != cand.item.name);
                    }
                    Err(why) => {
                        stopped = json!({
                            "code": "land_refused", "card": cand.item.name, "why": why,
                        });
                        break 'outer;
                    }
                }
            }
            continue 'outer;
        }
    }
    for item in &pending {
        not_run.push(entry(
            item,
            Some("not run: the batch stopped before this head"),
        ));
    }
    Ok(json!({
        "action": "land-batch", "branch": branch,
        "landed": landed_cards, "dropped": dropped, "not_run": not_run,
        "runs": runs, "checks_run": runs.len(), "stopped": stopped,
        "writes": !landed_cards.is_empty(),
    }))
}

/// Whether landing `commit` on `tip` changes the branch (it is not inside the tip).
fn landed_moves(project: &Path, commit: &str, tip: &str) -> bool {
    !matches!(
        landed::contained(project, commit, tip),
        Ok(Containment::Inside)
    )
}
