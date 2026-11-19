//! `moai bundle ready | check | body | build` (card C-119, SPEC-20261119-bundle-ready-check and
//! SPEC-20261119-bundle-build): read the `LANE-READY-INTEGRATION` heads of the ledger, judge each
//! head, write the M-Flow bundle pull request body from facts, and build the integration branch.
//! Nothing here fetches, pushes, opens a pull request, writes a card note or a ledger row
//! (REQ-12). `ready`, `check` and `body` write no ref and no branch either: their git commands
//! are the list of the SPEC's Plan, run through [`crate::audit_delta::command`] (scrubbed
//! environment). Only [`build`] writes, and only a new branch, its worktree and merge commits.

pub mod build;

use crate::card::{command as card_command, input as card_input};
use crate::{gate, self_audit, spec_doc, trailers};
use serde_json::{json, Value};
use std::collections::BTreeSet;
use std::path::{Path, PathBuf};

type Result<T> = std::result::Result<T, String>;

/// The first-line marker of a card note that names a head ready for a bundle.
pub const MARKER: &str = "LANE-READY-INTEGRATION";

/// The checks of the bundle's verification list, pinned by a test (D-4 of the parent SPEC): the
/// keys a run record names. A check neither run nor named not run is a missing fact for `body`.
pub const CHECKS: [&str; 10] = [
    "build",
    "lib_tests",
    "home_isolation_scan",
    "product_cli",
    "fmt",
    "clippy",
    "trailer_check",
    "spec_check",
    "result_coverage",
    "unittest",
];

/// The self-audit lines a run record carries (what `self_audit::check` asks for).
pub const SELF_AUDIT_KEYS: [&str; 6] = ["plan", "red", "mutation", "clone", "host", "receipt"];

/// The text merge GitHub runs: a merge driver registered in the clone must not decide.
const TEXT_DRIVER: &str =
    "merge.moai-manifest.driver=git merge-file -L ours -L base -L theirs %A %O %B";

const AUDIT_NOTE: &str =
    "a `PASS` note naming the head counts; the tool cannot tell who wrote it or at which tier";

// ---------------------------------------------------------------------------------------------
// git (read-only)

struct Out {
    code: Option<i32>,
    out: String,
    err: String,
}

struct Git<'a> {
    dir: &'a Path,
}

impl Git<'_> {
    fn raw(&self, args: &[&str]) -> Result<Out> {
        let o = crate::audit_delta::command(self.dir)
            .args(args)
            .output()
            .map_err(|e| format!("cannot execute git: {e}"))?;
        Ok(Out {
            code: o.status.code(),
            out: String::from_utf8_lossy(&o.stdout).into_owned(),
            err: String::from_utf8_lossy(&o.stderr).trim().to_string(),
        })
    }

    /// stdout of a command that succeeded; `None` when it exited non-zero.
    fn ok(&self, args: &[&str]) -> Option<String> {
        self.raw(args)
            .ok()
            .filter(|o| o.code == Some(0))
            .map(|o| o.out)
    }

    fn lines(&self, args: &[&str]) -> Result<Vec<String>> {
        let o = self.raw(args)?;
        if o.code != Some(0) {
            return Err(format!(
                "git {} failed: {}",
                args.first().unwrap_or(&""),
                o.err
            ));
        }
        Ok(o.out
            .lines()
            .map(str::trim)
            .filter(|l| !l.is_empty())
            .map(String::from)
            .collect())
    }

    /// The commit a revision names, or `None`.
    fn commit(&self, rev: &str) -> Option<String> {
        if rev.is_empty() || rev.starts_with('-') {
            return None;
        }
        self.ok(&[
            "rev-parse",
            "--verify",
            "--quiet",
            &format!("{rev}^{{commit}}"),
        ])
        .map(|s| s.trim().to_string())
        .filter(|s| gate::is_full_sha(s))
    }

    fn is_ancestor(&self, a: &str, b: &str) -> bool {
        self.raw(&["merge-base", "--is-ancestor", a, b])
            .is_ok_and(|o| o.code == Some(0))
    }

    /// The `origin` branches that contain `sha` (without `origin/HEAD`).
    fn origin_branches(&self, sha: &str, points_at: bool) -> Vec<String> {
        let flag = if points_at {
            "--points-at"
        } else {
            "--contains"
        };
        self.lines(&[
            "for-each-ref",
            flag,
            sha,
            "--format=%(refname)",
            "refs/remotes/origin",
        ])
        .unwrap_or_default()
        .into_iter()
        .filter_map(|r| r.strip_prefix("refs/remotes/origin/").map(String::from))
        .filter(|b| b != "HEAD")
        .collect()
    }

    /// `Moai-Spec` values of the non-merge commits of `tip..head`, sorted, without repeats.
    fn spec_ids(&self, tip: &str, head: &str) -> Vec<String> {
        let range = format!("{tip}..{head}");
        let lines = self
            .lines(&[
                "log",
                "--no-merges",
                "--format=%(trailers:key=Moai-Spec,valueonly)",
                &range,
            ])
            .unwrap_or_default();
        lines
            .into_iter()
            .collect::<BTreeSet<_>>()
            .into_iter()
            .collect()
    }

    /// Paths changed since the merge base of `a` and `b` (three dots: a head behind the tip does
    /// not count the tip's own changes).
    fn changed_paths(&self, a: &str, b: &str) -> Result<Vec<String>> {
        self.lines(&["diff", "--name-only", &format!("{a}...{b}")])
    }

    /// `Ok(None)` for a clean merge, `Ok(Some(files))` for a conflict, run as GitHub would (the
    /// text driver, not a driver registered in the clone). The exit status is the process's.
    fn conflict(&self, a: &str, b: &str) -> Result<Option<Vec<String>>> {
        let o = self.raw(&[
            "-c",
            TEXT_DRIVER,
            "merge-tree",
            "--write-tree",
            "--name-only",
            "--no-messages",
            a,
            b,
        ])?;
        match o.code {
            Some(0) => Ok(None),
            Some(1) => {
                let mut files: Vec<String> = Vec::new();
                for l in o.out.lines().skip(1) {
                    if l.is_empty() {
                        break;
                    }
                    if !files.iter().any(|f| f == l) {
                        files.push(l.to_string());
                    }
                }
                Ok(Some(files))
            }
            _ => Err(format!("git merge-tree failed: {}", o.err)),
        }
    }
}

// ---------------------------------------------------------------------------------------------
// ledger (read-only)

struct Note {
    id: i64,
    body: String,
}

impl Note {
    fn first_line(&self) -> &str {
        self.body.lines().next().unwrap_or("")
    }
}

fn card_notes(project: &Path, home: &Path, name: &str) -> Result<Vec<Note>> {
    let v = card_command::show(project, home, name)?;
    Ok(v["notes"]
        .as_array()
        .map(|a| {
            a.iter()
                .filter_map(|n| {
                    Some(Note {
                        id: n["note_id"].as_i64()?,
                        body: n["body"].as_str()?.to_string(),
                    })
                })
                .collect()
        })
        .unwrap_or_default())
}

/// One card's newest marker note and what it says.
struct Ready {
    card: String,
    note_id: i64,
    /// `None`: the note does not read under the grammar (`unreadable`).
    head: Option<String>,
    spec: Option<String>,
    lines: usize,
}

/// The newest note of the card whose first line starts with the marker; the newest wins even
/// when it is the one that does not read (an older head would be stale).
fn newest_marker(card: &str, notes: &[Note]) -> Option<Ready> {
    let marked: Vec<&Note> = notes
        .iter()
        .filter(|n| n.first_line().starts_with(MARKER))
        .collect();
    let newest = marked.iter().max_by_key(|n| n.id)?;
    let parsed = parse_ready(newest.first_line(), card);
    Some(Ready {
        card: card.to_string(),
        note_id: newest.id,
        head: parsed.as_ref().map(|p| p.0.clone()),
        spec: parsed.and_then(|p| p.1),
        lines: marked.len(),
    })
}

fn spec_id_shaped(s: &str) -> bool {
    // SPEC-<8 digits>-<slug of a-z, 0-9 and ->
    let Some(rest) = s.strip_prefix("SPEC-") else {
        return false;
    };
    let Some((date, slug)) = rest.split_once('-') else {
        return false;
    };
    date.len() == 8
        && date.bytes().all(|b| b.is_ascii_digit())
        && !slug.is_empty()
        && slug
            .bytes()
            .all(|b| b.is_ascii_lowercase() || b.is_ascii_digit() || b == b'-')
}

/// D-6 of SPEC-20261119-bundle-tool: `LANE-READY-INTEGRATION C-<n> head=<40 hex>` then the end
/// of the line or a space, with an optional ` spec=<SPEC id>` right after the head; the card
/// must be the card the note is on. Returns the head and the optional `spec=` value.
pub fn parse_ready(line: &str, card: &str) -> Option<(String, Option<String>)> {
    let rest = line.strip_prefix(MARKER)?.strip_prefix(' ')?;
    let (named, rest) = rest.split_once(' ')?;
    if named != card {
        return None;
    }
    let rest = rest.strip_prefix("head=")?;
    let head = rest.get(..40)?;
    if !gate::is_full_sha(head) {
        return None;
    }
    let after = &rest[40..];
    if !(after.is_empty() || after.starts_with(' ')) {
        return None;
    }
    let spec = after
        .strip_prefix(" spec=")
        .map(|s| s.split(' ').next().unwrap_or(""))
        .filter(|s| spec_id_shaped(s))
        .map(String::from);
    Some((head.to_string(), spec))
}

fn all_cards(project: &Path, home: &Path) -> Result<Vec<String>> {
    let list = card_command::list(project, home, None)?;
    if list["ledger"] == "absent" {
        return Err("bundle refused: the project has no ledger".into());
    }
    Ok(list["cards"]
        .as_array()
        .map(|a| {
            a.iter()
                .filter_map(|c| c["name"].as_str().map(String::from))
                .collect()
        })
        .unwrap_or_default())
}

// ---------------------------------------------------------------------------------------------
// the tip

struct Tip {
    name: String,
    sha: String,
    committed: String,
}

fn tip_of(git: &Git, tip: &str) -> Result<Tip> {
    let sha = git.commit(tip).ok_or_else(|| {
        format!("bundle refused: the tip `{tip}` is not a commit here (fetch first)")
    })?;
    let committed = git
        .ok(&["log", "-1", "--format=%cI", &sha])
        .unwrap_or_default()
        .trim()
        .to_string();
    Ok(Tip {
        name: tip.to_string(),
        sha,
        committed,
    })
}

fn tip_json(t: &Tip) -> Value {
    json!({"name": t.name, "sha": t.sha, "committed": t.committed,
           "note": "the tool never fetches; git keeps no fetch time shared by every worktree"})
}

// ---------------------------------------------------------------------------------------------
// ready

/// `bundle ready`: one entry per card whose newest marker note reads, the unreadable ones named.
pub fn ready(project: &Path, home: &Path, dir: &Path, tip: &str) -> Result<Value> {
    let git = Git { dir };
    let tip = tip_of(&git, tip)?;
    let mut heads = Vec::new();
    let mut unreadable = Vec::new();
    for name in all_cards(project, home)? {
        let notes = card_notes(project, home, &name)?;
        let Some(r) = newest_marker(&name, &notes) else {
            continue;
        };
        match &r.head {
            None => unreadable.push(json!({"card": r.card, "note_id": r.note_id})),
            Some(h) => {
                let present = git.commit(h).is_some();
                let specs = if present {
                    git.spec_ids(&tip.sha, h)
                } else {
                    vec![]
                };
                let on_origin = present && !git.origin_branches(h, false).is_empty();
                heads.push(json!({
                    "card": r.card, "note_id": r.note_id, "head": h, "spec": r.spec,
                    "specs": specs, "present": present, "on_origin": on_origin,
                    "already_in_tip": present && git.is_ancestor(h, &tip.sha),
                    "ready_lines": r.lines,
                }));
            }
        }
    }
    Ok(json!({
        "action": "ready", "tip": tip_json(&tip), "count": heads.len(), "heads": heads,
        "unreadable": unreadable, "writes": false,
    }))
}

// ---------------------------------------------------------------------------------------------
// check

struct HeadArg {
    card: String,
    sha: Option<String>,
}

fn parse_head_args(args: &[String]) -> Result<Vec<HeadArg>> {
    if args.is_empty() {
        return Err("bundle refused: name at least one head as C-<n> or C-<n>=<40 hex>".into());
    }
    let mut out: Vec<HeadArg> = Vec::new();
    for a in args {
        let (card, sha) = match a.split_once('=') {
            Some((c, s)) => (c, Some(s)),
            None => (a.as_str(), None),
        };
        let id = match card_input::parse_ref(card) {
            Ok(card_input::CardRef::Id(n)) => n,
            _ => {
                return Err(format!(
                    "bundle refused: `{a}` is not C-<n> or C-<n>=<40 hex>"
                ))
            }
        };
        let card = card_input::card_name(id);
        let sha = match sha {
            None => None,
            Some(s) if gate::is_full_sha(s) => Some(s.to_string()),
            Some(_) => {
                return Err(format!(
                    "bundle refused: `{a}`: the head after `=` is 40 lowercase hex digits"
                ))
            }
        };
        if out.iter().any(|h| h.card == card) {
            return Err(format!("bundle refused: {card} is named twice"));
        }
        out.push(HeadArg { card, sha });
    }
    Ok(out)
}

fn finding(code: &str, card: &str, head: &str, detail: Value) -> Value {
    json!({"code": code, "card": card, "head": head, "detail": detail})
}

/// What `check` and `body` learn about one head.
struct Judged {
    card: String,
    head: Option<String>,
    findings: Vec<Value>,
    floor: u8,
    pass_note: Option<i64>,
    pass_line: Option<String>,
    specs: Vec<String>,
    /// The `## Result` outcomes of each named SPEC at the head: `(id, ac count, outcomes)`.
    spec_results: Vec<(String, usize, Vec<String>)>,
    unreadable_verdicts: Vec<Value>,
    origin_branch: Option<String>,
}

fn judge(
    git: &Git,
    project: &Path,
    home: &Path,
    tip: &Tip,
    arg: &HeadArg,
    earlier: &[(String, String)],
) -> Result<Judged> {
    let mut j = Judged {
        card: arg.card.clone(),
        head: None,
        findings: vec![],
        floor: 0,
        pass_note: None,
        pass_line: None,
        specs: vec![],
        spec_results: vec![],
        unreadable_verdicts: vec![],
        origin_branch: None,
    };
    let notes = card_notes(project, home, &arg.card)?;
    let head = match &arg.sha {
        Some(s) => s.clone(),
        None => match newest_marker(&arg.card, &notes) {
            Some(Ready { head: Some(h), .. }) => h,
            Some(r) => {
                j.findings.push(finding(
                    "no_ready_line",
                    &arg.card,
                    "",
                    json!(format!(
                        "the newest marker note {} does not read",
                        r.note_id
                    )),
                ));
                return Ok(j);
            }
            None => {
                j.findings.push(finding(
                    "no_ready_line",
                    &arg.card,
                    "",
                    json!("no marker note"),
                ));
                return Ok(j);
            }
        },
    };
    j.head = Some(head.clone());
    if git.commit(&head).is_none() {
        j.findings.push(finding(
            "head_missing",
            &arg.card,
            &head,
            json!("fetch first"),
        ));
        return Ok(j);
    }
    // origin
    let containing = git.origin_branches(&head, false);
    if containing.is_empty() {
        j.findings.push(finding(
            "not_on_origin",
            &arg.card,
            &head,
            json!("no branch of origin contains the head (push it, then fetch)"),
        ));
    }
    j.origin_branch = git
        .origin_branches(&head, true)
        .into_iter()
        .next()
        .or_else(|| containing.first().cloned());
    if git.is_ancestor(&head, &tip.sha) {
        j.findings.push(finding(
            "already_merged",
            &arg.card,
            &head,
            json!(format!("the head is already in {}", tip.name)),
        ));
    }
    // audit evidence
    // A head with no merge base with the tip (an unrelated history) is one finding, not an abort:
    // the other heads are still judged (REQ-3).
    let paths = match git.changed_paths(&tip.sha, &head) {
        Ok(p) => p,
        Err(e) => {
            j.findings
                .push(finding("no_merge_base", &arg.card, &head, json!(e)));
            return Ok(j);
        }
    };
    j.floor = gate::tier_of(Some(0), &paths);
    let mut passes: Vec<(i64, String, String)> = Vec::new(); // (note, head, line)
    for n in &notes {
        let first = n.first_line();
        if !first.starts_with("AUDIT ") {
            continue;
        }
        match gate::parse_verdict(first) {
            Some((h, _base, verdict)) => {
                if verdict == "PASS" {
                    passes.push((n.id, h, first.to_string()));
                }
            }
            None => j
                .unreadable_verdicts
                .push(json!({"card": arg.card, "note_id": n.id})),
        }
    }
    if let Some((id, _, line)) = passes.iter().filter(|p| p.1 == head).max_by_key(|p| p.0) {
        j.pass_note = Some(*id);
        j.pass_line = Some(line.clone());
    }
    if j.floor >= 1 && j.pass_note.is_none() {
        let nearest = passes
            .iter()
            .filter(|p| p.1 != head && git.is_ancestor(&p.1, &head))
            .filter_map(|p| {
                let behind = git
                    .ok(&["rev-list", "--count", &format!("{}..{}", p.1, head)])?
                    .trim()
                    .parse::<u64>()
                    .ok()?;
                Some((behind, p))
            })
            .min_by_key(|(b, _)| *b);
        let ancestor_pass =
            nearest.map(|(b, p)| json!({"head": p.1, "note_id": p.0, "commits_behind": b}));
        j.findings.push(finding(
            "audit_missing",
            &arg.card,
            &head,
            json!({"floor": j.floor, "ancestor_pass": ancestor_pass}),
        ));
    }
    // trailers
    let branch = j.origin_branch.clone().unwrap_or_else(|| "detached".into());
    match trailers::check(
        git.dir,
        &format!("{}..{}", tip.sha, head),
        Some(&branch),
        None,
    ) {
        Ok(v) if v["ok"] == json!(false) => {
            let fails: Vec<Value> = v["findings"]
                .as_array()
                .map(|a| {
                    a.iter()
                        .filter(|f| f["fails"] == json!(true))
                        .take(5)
                        .cloned()
                        .collect()
                })
                .unwrap_or_default();
            j.findings.push(finding(
                "trailers",
                &arg.card,
                &head,
                json!({"failing": v["failing"], "findings": fails}),
            ));
        }
        Ok(_) => {}
        Err(e) => j
            .findings
            .push(finding("trailers", &arg.card, &head, json!(e))),
    }
    // SPECs
    j.specs = git.spec_ids(&tip.sha, &head);
    if j.specs.is_empty() && !git.is_ancestor(&head, &tip.sha) {
        j.findings.push(finding(
            "spec",
            &arg.card,
            &head,
            json!("no Moai-Spec trailer in the commits of the head"),
        ));
    }
    for id in j.specs.clone() {
        if spec_doc::id_form(&id).is_none() {
            j.findings.push(finding(
                "spec",
                &arg.card,
                &head,
                json!(format!("`{id}` is not a SPEC id")),
            ));
            continue;
        }
        let Some(text) = git.ok(&["show", &format!("{head}:.moai/specs/{id}/spec.md")]) else {
            j.findings.push(finding(
                "spec",
                &arg.card,
                &head,
                json!(format!("{id}: no spec.md at the head")),
            ));
            continue;
        };
        let doc = spec_doc::parse(
            &text,
            &spec_doc::Options {
                folder: Some(id.clone()),
                closure: false,
            },
        );
        if doc.errors() > 0 {
            let codes: Vec<String> = doc
                .findings
                .iter()
                .filter(|f| f.severity == spec_doc::Severity::Error)
                .take(3)
                .map(|f| f.code.to_string())
                .collect();
            j.findings.push(finding(
                "spec",
                &arg.card,
                &head,
                json!({"spec": id, "errors": doc.errors(), "codes": codes}),
            ));
        }
        let have: BTreeSet<u32> = doc.result.iter().map(|o| o.number).collect();
        let missing: Vec<String> = doc
            .acceptance
            .iter()
            .filter(|c| !have.contains(&c.number))
            .map(|c| c.id.clone())
            .collect();
        if !missing.is_empty() {
            j.findings.push(finding(
                "result_missing",
                &arg.card,
                &head,
                json!({"spec": id, "acs": missing}),
            ));
        }
        j.spec_results.push((
            id,
            doc.acceptance.len(),
            doc.result
                .iter()
                .map(|o| format!("{} {}", o.ac, o.outcome))
                .collect(),
        ));
    }
    // conflicts: with the tip, then with every earlier head, two at a time
    let mut others: Vec<(String, String)> = vec![(tip.name.clone(), tip.sha.clone())];
    others.extend(earlier.iter().cloned());
    for (label, other) in &others {
        match git.conflict(other, &head) {
            Ok(None) => {}
            Ok(Some(files)) => j.findings.push(finding(
                "conflict",
                &arg.card,
                &head,
                json!({"with": label, "files": files}),
            )),
            Err(e) => j.findings.push(finding(
                "conflict",
                &arg.card,
                &head,
                json!({"with": label, "error": e}),
            )),
        }
    }
    Ok(j)
}

fn judge_all(
    git: &Git,
    project: &Path,
    home: &Path,
    tip: &Tip,
    args: &[HeadArg],
) -> Result<Vec<Judged>> {
    let mut out: Vec<Judged> = Vec::new();
    let mut earlier: Vec<(String, String)> = Vec::new();
    for a in args {
        let j = judge(git, project, home, tip, a, &earlier)?;
        let unrelated = j.findings.iter().any(|f| f["code"] == "no_merge_base");
        if let Some(h) = &j.head {
            if git.commit(h).is_some() && !unrelated {
                earlier.push((a.card.clone(), h.clone()));
            }
        }
        out.push(j);
    }
    Ok(out)
}

/// `bundle check`: every check for every head, all findings in one result.
pub fn check(
    project: &Path,
    home: &Path,
    dir: &Path,
    tip: &str,
    heads: &[String],
) -> Result<Value> {
    let args = parse_head_args(heads)?;
    let git = Git { dir };
    let tip = tip_of(&git, tip)?;
    let judged = judge_all(&git, project, home, &tip, &args)?;
    let findings: Vec<Value> = judged.iter().flat_map(|j| j.findings.clone()).collect();
    let unreadable: Vec<Value> = judged
        .iter()
        .flat_map(|j| j.unreadable_verdicts.clone())
        .collect();
    let heads_json: Vec<Value> = judged
        .iter()
        .map(|j| {
            json!({"card": j.card, "head": j.head, "audit_floor": j.floor,
                   "pass_note": j.pass_note, "specs": j.specs,
                   "findings": j.findings.len()})
        })
        .collect();
    Ok(json!({
        "action": "check", "ok": findings.is_empty(), "tip": tip_json(&tip),
        "heads": heads_json, "findings": findings,
        "unreadable_verdict": unreadable,
        "audit_note": AUDIT_NOTE,
        "conflict_note": "each head is merged with the tip and with every earlier head, two at a time; `build` is the exact three-way check",
        "writes": false,
    }))
}

// ---------------------------------------------------------------------------------------------
// body

fn nonempty_str<'a>(v: &'a Value, key: &str) -> Option<&'a str> {
    v[key].as_str().map(str::trim).filter(|s| !s.is_empty())
}

/// What the run record says, checked for the facts `body` needs; every gap is one string.
struct Record {
    /// `(command, exit, counts)` per command, in file order.
    commands: Vec<(String, i64, Option<String>, Option<String>)>,
    not_run: Vec<(String, String)>,
    self_audit: Vec<(String, String)>,
}

fn read_record(path: &Path, gaps: &mut Vec<String>) -> Result<Record> {
    let text = std::fs::read_to_string(path).map_err(|e| {
        format!(
            "bundle refused: cannot read the run record {}: {e}",
            path.display()
        )
    })?;
    let v: Value = serde_json::from_str(&text)
        .map_err(|e| format!("bundle refused: the run record is not JSON: {e}"))?;
    if !v.is_object() {
        return Err("bundle refused: the run record is not a JSON object".into());
    }
    let mut rec = Record {
        commands: vec![],
        not_run: vec![],
        self_audit: vec![],
    };
    for c in v["commands"].as_array().map(Vec::as_slice).unwrap_or(&[]) {
        match (nonempty_str(c, "command"), c["exit"].as_i64()) {
            (Some(cmd), Some(exit)) => rec.commands.push((
                cmd.to_string(),
                exit,
                nonempty_str(c, "counts").map(String::from),
                nonempty_str(c, "check").map(String::from),
            )),
            _ => gaps.push("a `commands` entry lacks its `command` text or integer `exit`".into()),
        }
    }
    for n in v["not_run"].as_array().map(Vec::as_slice).unwrap_or(&[]) {
        match (nonempty_str(n, "check"), nonempty_str(n, "reason")) {
            (Some(c), Some(r)) => rec.not_run.push((c.to_string(), r.to_string())),
            _ => gaps.push("a `not_run` entry lacks its `check` or its `reason`".into()),
        }
    }
    for k in SELF_AUDIT_KEYS {
        match nonempty_str(&v["self_audit"], k) {
            Some(s) => rec.self_audit.push((k.to_string(), s.to_string())),
            None => gaps.push(format!("the Self-Audit line `{k}` has no text")),
        }
    }
    let line_break = |s: &str| s.contains(['\n', '\r']);
    for (cmd, _, counts, _) in &rec.commands {
        if line_break(cmd) || counts.as_deref().is_some_and(line_break) {
            gaps.push(format!(
                "a command text or its counts has a line break: `{}`",
                cmd.lines().next().unwrap_or("")
            ));
        }
    }
    for (c, r) in &rec.not_run {
        if line_break(r) {
            gaps.push(format!("the not_run reason of `{c}` has a line break"));
        }
    }
    for (k, v) in &rec.self_audit {
        if line_break(v) {
            gaps.push(format!("the Self-Audit line `{k}` has a line break"));
        }
    }
    for key in CHECKS {
        let ran = rec.commands.iter().any(|c| c.3.as_deref() == Some(key));
        let not = rec.not_run.iter().any(|n| n.0 == key);
        match (ran, not) {
            (false, false) => gaps.push(format!(
                "the check `{key}` neither ran nor is named not run"
            )),
            (true, true) => gaps.push(format!("the check `{key}` is both run and named not run")),
            _ => {}
        }
    }
    for (cmd, exit, _, _) in &rec.commands {
        if *exit != 0 {
            gaps.push(format!("the command `{cmd}` ended with status {exit}"));
        }
    }
    Ok(rec)
}

/// `bundle body`: the pull request body, from the heads, the branch and the run record.
#[allow(clippy::too_many_arguments)]
pub fn body(
    project: &Path,
    home: &Path,
    dir: &Path,
    tip: &str,
    heads: &[String],
    branch: &str,
    record: &Path,
    out: Option<&Path>,
) -> Result<Value> {
    let args = parse_head_args(heads)?;
    let git = Git { dir };
    let tip = tip_of(&git, tip)?;
    let pinned = git
        .commit(branch)
        .ok_or_else(|| format!("bundle refused: the branch `{branch}` is not a commit here"))?;
    let mut gaps: Vec<String> = Vec::new();
    let rec = read_record(record, &mut gaps)?;
    let judged = judge_all(&git, project, home, &tip, &args)?;
    for j in &judged {
        for f in &j.findings {
            // A body is written only for heads that pass every check of `check`.
            let code = f["code"].as_str().unwrap_or("");
            gaps.push(format!("{} {}: {}", j.card, code, f["detail"]));
        }
        if let Some(h) = &j.head {
            if git.commit(h).is_some() && !git.is_ancestor(h, &pinned) {
                gaps.push(format!(
                    "{}: the head {} is not in the branch `{branch}`",
                    j.card,
                    &h[..10]
                ));
            }
        }
    }
    if !gaps.is_empty() {
        return Err(format!(
            "bundle body refused: {} missing fact(s): {}",
            gaps.len(),
            gaps.join("; ")
        ));
    }
    if let Some(p) = out {
        if std::fs::symlink_metadata(p).is_ok() {
            return Err(format!(
                "bundle body refused: {} already exists",
                p.display()
            ));
        }
    }
    // What the branch adds is counted from its merge base with the tip: when the tip has moved on
    // since the branch was built, the tip's own changes are not the bundle's.
    let base = git
        .ok(&["merge-base", &tip.sha, &pinned])
        .map(|s| s.trim().to_string())
        .filter(|s| gate::is_full_sha(s))
        .ok_or_else(|| {
            format!("bundle body refused: the branch `{branch}` has no merge base with the tip")
        })?;
    let paths = git.changed_paths(&tip.sha, &pinned)?;
    let tier = gate::tier_of(Some(0), &paths);
    let merges = git
        .ok(&[
            "rev-list",
            "--merges",
            "--count",
            &format!("{base}..{pinned}"),
        ])
        .unwrap_or_default()
        .trim()
        .to_string();
    let (mut files, mut ins, mut del) = (0u64, 0u64, 0u64);
    for l in git
        .lines(&["diff", "--numstat", &base, &pinned])
        .unwrap_or_default()
    {
        let mut it = l.split('\t');
        files += 1;
        ins += it.next().and_then(|s| s.parse::<u64>().ok()).unwrap_or(0);
        del += it.next().and_then(|s| s.parse::<u64>().ok()).unwrap_or(0);
    }
    let all_specs: BTreeSet<&String> = judged.iter().flat_map(|j| j.specs.iter()).collect();

    let mut b = String::new();
    if paths.is_empty() {
        b.push_str(&format!("audit-tier: {tier}\n\n"));
    } else {
        b.push_str(&format!("audit-tier: {tier} {}\n\n", paths.join(" ")));
    }
    for s in &all_specs {
        b.push_str(&format!("Moai-Spec: {s}\n"));
    }
    b.push_str(
        "\nM-Flow bundle (lanes push only, the leader integrates audited heads locally, verifies the \
         pinned SHA locally, then opens one pull request; CI, independent audit and the ruleset are unchanged).\n\n\
         | SPEC | Card | Branch head | Audit before integration |\n| --- | --- | --- | --- |\n",
    );
    for j in &judged {
        let head = j.head.as_deref().unwrap_or("");
        let branch_name = j
            .origin_branch
            .as_deref()
            .map(|n| format!(" (branch `{n}`)"))
            .unwrap_or_default();
        let audit = match (&j.pass_note, &j.pass_line) {
            (Some(id), Some(line)) => format!("`PASS` card note {id}: {}", line.replace('|', "/")),
            _ => format!("audit floor {}: no verdict needed", j.floor),
        };
        b.push_str(&format!(
            "| {} | {} | `{head}`{branch_name} | {audit} |\n",
            j.specs.join(", "),
            j.card
        ));
    }
    let built = if base == tip.sha {
        format!("built from the tip `{base}`")
    } else {
        format!(
            "built from `{base}` (the tip `{}` has moved on since)",
            tip.sha
        )
    };
    let s = |n: u64| if n == 1 { "" } else { "s" };
    b.push_str(&format!(
        "\nIntegration: `{branch}` {built} with {merges} merge commits → pinned SHA `{pinned}`. \
         {files} file{} changed, {ins} insertion{}(+), {del} deletion{}(-).\n\n## Self-Audit\n\n",
        s(files),
        s(ins),
        s(del)
    ));
    for (k, v) in &rec.self_audit {
        b.push_str(&format!("- {k}: {v}\n"));
    }
    if !rec.not_run.is_empty() {
        let list: Vec<String> = rec
            .not_run
            .iter()
            .map(|(c, r)| format!("{c} ({r})"))
            .collect();
        b.push_str(&format!("- Not done locally: {}.\n", list.join("; ")));
    }
    b.push_str(&format!(
        "\n## Result\n\nLocal verification on `{pinned}`:\n\n"
    ));
    for (cmd, exit, counts, _) in &rec.commands {
        match counts {
            Some(c) => b.push_str(&format!("- `{cmd}` → exit {exit}, {c}.\n")),
            None => b.push_str(&format!("- `{cmd}` → exit {exit}.\n")),
        }
    }
    for (c, r) in &rec.not_run {
        b.push_str(&format!("- not run: {c} — {r}.\n"));
    }
    b.push('\n');
    for j in &judged {
        for (id, acs, outcomes) in &j.spec_results {
            let met = outcomes.iter().filter(|o| o.ends_with(" met")).count();
            let withdrawn = outcomes
                .iter()
                .filter(|o| o.ends_with(" withdrawn"))
                .count();
            let other: Vec<&String> = outcomes
                .iter()
                .filter(|o| !o.ends_with(" met") && !o.ends_with(" withdrawn"))
                .collect();
            let tail = if other.is_empty() {
                String::new()
            } else {
                format!(
                    ", other: {}",
                    other
                        .iter()
                        .map(|s| s.as_str())
                        .collect::<Vec<_>>()
                        .join("; ")
                )
            };
            b.push_str(&format!(
                "{id} ({}): {acs} AC, {met} met, {withdrawn} withdrawn{tail} — see the SPEC's Result.\n\n",
                j.card
            ));
        }
    }
    b.push_str("🤖 Generated with [Claude Code](https://claude.com/claude-code)\n");
    if let Err(missing) = self_audit::check(&b, false) {
        return Err(format!(
            "bundle body refused: the text fails the Self-Audit checker: {}",
            missing.join(", ")
        ));
    }
    if let Some(p) = out {
        use std::io::Write;
        let mut f = std::fs::OpenOptions::new()
            .write(true)
            .create_new(true)
            .open(p)
            .map_err(|e| format!("bundle body refused: cannot create {}: {e}", p.display()))?;
        f.write_all(b.as_bytes())
            .map_err(|e| format!("cannot write {}: {e}", p.display()))?;
    }
    Ok(json!({
        "action": "body", "tier": tier, "pinned": pinned, "tip": tip_json(&tip),
        "specs": all_specs, "out": out.map(PathBuf::from), "body": b,
        "writes": out.is_some(),
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    const H: &str = "0123456789abcdef0123456789abcdef01234567";

    #[test]
    fn the_ready_grammar_reads_head_with_optional_spec_and_text() {
        let line = format!("{MARKER} C-12 head={H} spec=SPEC-20261119-x-y");
        assert_eq!(
            parse_ready(&line, "C-12"),
            Some((H.into(), Some("SPEC-20261119-x-y".into())))
        );
        let plain = format!("{MARKER} C-12 head={H} — repaired");
        assert_eq!(parse_ready(&plain, "C-12"), Some((H.into(), None)));
        let bare = format!("{MARKER} C-12 head={H}");
        assert_eq!(parse_ready(&bare, "C-12"), Some((H.into(), None)));
    }

    #[test]
    fn the_ready_grammar_refuses_the_lines_it_cannot_read() {
        // another card, a short head, a head glued to text, upper-case hex, the old free form
        assert!(parse_ready(&format!("{MARKER} C-13 head={H}"), "C-12").is_none());
        assert!(parse_ready(&format!("{MARKER} C-12 head={}", &H[..39]), "C-12").is_none());
        assert!(parse_ready(&format!("{MARKER} C-12 head={H}x"), "C-12").is_none());
        assert!(parse_ready(&format!("{MARKER} C-12 head={}", H.to_uppercase()), "C-12").is_none());
        assert!(parse_ready(&format!("{MARKER} agent-1 C-12 head {H}"), "C-12").is_none());
        assert!(parse_ready("LANE-READY-INTEGRATIONC-12 head=x", "C-12").is_none());
    }

    #[test]
    fn a_malformed_spec_value_leaves_the_line_readable_without_a_spec() {
        let line = format!("{MARKER} C-12 head={H} spec=nope");
        assert_eq!(parse_ready(&line, "C-12"), Some((H.into(), None)));
    }

    #[test]
    fn the_check_keys_are_pinned() {
        assert_eq!(
            CHECKS,
            [
                "build",
                "lib_tests",
                "home_isolation_scan",
                "product_cli",
                "fmt",
                "clippy",
                "trailer_check",
                "spec_check",
                "result_coverage",
                "unittest"
            ]
        );
    }

    #[test]
    fn head_arguments_are_cards_with_an_optional_full_head() {
        let ok = parse_head_args(&["C-5".into(), format!("C-6={H}")]).unwrap();
        assert_eq!(ok[0].card, "C-5");
        assert_eq!(ok[1].sha.as_deref(), Some(H));
        for bad in [
            "5",
            "C-x",
            "C-5=abc",
            "#5",
            "C-5=0123456789ABCDEF0123456789ABCDEF01234567",
        ] {
            assert!(parse_head_args(&[bad.into()]).is_err(), "{bad}");
        }
        assert!(parse_head_args(&["C-5".into(), "C-5".into()]).is_err());
        assert!(parse_head_args(&[]).is_err());
    }
}
