//! `moai bundle ready | check | body | build | verify` (card C-119, SPEC-20261119-bundle-ready-check,
//! SPEC-20261119-bundle-build and SPEC-20261119-bundle-verify): read the `LANE-READY-INTEGRATION`
//! heads of the ledger, judge each head, write the M-Flow bundle pull request body from facts,
//! build the integration branch and run the check list on it (`verify`, which runs commands).
//! Nothing here fetches, pushes, opens a pull request, writes a card note or a ledger row
//! (REQ-12). `ready`, `check` and `body` write no ref and no branch either: their git commands
//! are the list of the SPEC's Plan, run through [`crate::audit_delta::command`] (scrubbed
//! environment). [`build`] writes a new branch, its worktree and merge commits; [`verify`] writes one new
//! folder and runs commands.

pub mod build;
pub mod verify;

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

/// The text merge GitHub runs: a merge driver named in the configuration must not decide.
const TEXT_MERGE: &str = "git merge-file -L ours -L base -L theirs %A %O %B";

/// `-c merge.<name>.driver=<text merge>` for the manifest driver and for every driver the
/// configuration names (`configured`, from [`driver_names`]): a driver registered in the clone,
/// the manifest's or another one chosen by a `.gitattributes` line, must not decide (D-12 of
/// SPEC-20261119-bundle-build, D-8 of SPEC-20261119-bundle-ready-check).
fn text_merge_args(configured: &[String]) -> Vec<String> {
    let mut names: Vec<&str> = vec!["moai-manifest"];
    for n in configured {
        if !names.contains(&n.as_str()) {
            names.push(n);
        }
    }
    names
        .iter()
        .flat_map(|n| ["-c".to_string(), format!("merge.{n}.driver={TEXT_MERGE}")])
        .collect()
}

/// The driver names in the output of `git config --get-regexp '^merge\..*\.driver$'` (one
/// `<key> <value>` per line); a name git's `-c` cannot carry (a space or `=`) is left out.
fn driver_names(config_output: &str) -> Vec<String> {
    config_output
        .lines()
        .filter_map(|l| l.split_whitespace().next())
        .filter_map(|k| k.strip_prefix("merge.")?.strip_suffix(".driver"))
        .filter(|n| !n.is_empty() && !n.contains(['=', ' ']))
        .map(String::from)
        .collect()
}

const AUDIT_NOTE: &str = "a `PASS` note naming the head counts; a stated `tier <N>` before ` head ` is read (a head whose floor is above it has `audit_tier_low`); the tool cannot tell who wrote it";

/// Byte ranges of the card numbers in `text`: `C-` or `c-` and one or more digits, the letter not
/// inside a longer word (`AC-1` and `SPEC-2026…` are not card numbers). The card reference reader
/// takes both cases (`card::input::parse_ref`), so this reads both too.
/// SPEC-20261120-board-bundle-link.
fn card_number_spans(text: &str) -> Vec<(usize, usize)> {
    let b = text.as_bytes();
    let mut spans = Vec::new();
    let mut i = 0;
    while i + 2 < b.len() {
        let word_start = i == 0 || !b[i - 1].is_ascii_alphanumeric();
        if word_start
            && matches!(b[i], b'C' | b'c')
            && b[i + 1] == b'-'
            && b[i + 2].is_ascii_digit()
        {
            let end = b[i + 2..]
                .iter()
                .position(|c| !c.is_ascii_digit())
                .map_or(b.len(), |p| i + 2 + p);
            spans.push((i, end));
            i = end;
        } else {
            i += 1;
        }
    }
    spans
}

/// `text` with every card number replaced by the placeholder `<lane>`.
fn without_card_numbers(text: &str) -> String {
    let mut out = String::new();
    let mut at = 0;
    for (start, end) in card_number_spans(text) {
        out.push_str(&text[at..start]);
        out.push_str("<lane>");
        at = end;
    }
    out.push_str(&text[at..]);
    out
}

/// What a text that leaves the repository must not carry and still does: a card number or a
/// session id (SPEC-20261120-board-bundle-link REQ-5). The session id shape is the one the
/// repository scans already use (`check_tree::uuid_hits`), not a second copy of it.
pub fn outward_leak(text: &str) -> Option<&'static str> {
    if !card_number_spans(text).is_empty() {
        Some("a card number (`C-` or `c-` and digits)")
    } else if crate::check_tree::uuid_hits(text.as_bytes()) > 0 {
        Some("a session id (hex groups 8-4-4-4-12)")
    } else {
        None
    }
}

/// Names the field of the body a leaking line belongs to, without printing the line's value: the
/// reader of a refusal needs to know which value to reword (the value is the leak itself).
fn leak_field(line: &str) -> String {
    if let Some(rest) = line.strip_prefix("- ") {
        if rest.starts_with('`') {
            return "a command of the run record".to_string();
        }
        if let Some((key, _)) = rest.split_once(": ") {
            return match key {
                "Not done locally" => "the not-run reasons (`Not done locally`)".to_string(),
                "not run" => "a not-run reason of the run record".to_string(),
                k if !k.contains('`') && outward_leak(k).is_none() => {
                    format!("the Self-Audit field `{k}`")
                }
                _ => "a Self-Audit field".to_string(),
            };
        }
    }
    let kind = if line.starts_with('|') {
        "the table of heads (SPEC id, branch name or head)"
    } else if line.starts_with("Moai-Spec:") {
        "a `Moai-Spec` line"
    } else if line.starts_with("audit-tier:") {
        "the `audit-tier` line (a changed path)"
    } else if line.starts_with("Integration:") {
        "the integration line (the branch name)"
    } else if line.starts_with("SPEC-") {
        "a SPEC result line"
    } else {
        "the text"
    };
    kind.to_string()
}

/// The tier a verdict line states before its last ` head ` (the verdict's own, as
/// `gate::parse_verdict` takes it): the word `tier`, a space and one digit from 0 to 2, the word
/// starting the text or following a space, `(` or `,` and the digit followed by the end of the
/// text, a space, `,`, `)` or `;`. One distinct value is the tier; none, two different values or
/// a digit above 2 give `None` (SPEC-20261119-bundle-audit-tier, D-4).
pub fn stated_tier(line: &str) -> Option<u8> {
    let prefix = &line[..line.rfind(" head ")?];
    let mut found: Option<u8> = None;
    for (i, _) in prefix.match_indices("tier ") {
        if i > 0 && !matches!(prefix.as_bytes()[i - 1], b' ' | b'(' | b',') {
            continue;
        }
        let mut rest = prefix[i + 5..].chars();
        let Some(d) = rest.next().filter(char::is_ascii_digit) else {
            continue;
        };
        if !matches!(rest.next(), None | Some(' ' | ',' | ')' | ';')) {
            continue;
        }
        let v = d as u8 - b'0';
        if v > 2 {
            return None;
        }
        match found {
            None => found = Some(v),
            Some(f) if f == v => {}
            Some(_) => return None,
        }
    }
    found
}

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

    /// The merge drivers the clone's configuration names (this wrapper runs with the global
    /// configuration off, so the clone's own file is what applies).
    fn configured_drivers(&self) -> Vec<String> {
        self.ok(&["config", "--get-regexp", r"^merge\..*\.driver$"])
            .map(|s| driver_names(&s))
            .unwrap_or_default()
    }

    /// `Ok(None)` for a clean merge, `Ok(Some(files))` for a conflict, run as GitHub would (the
    /// text merge, whatever driver the configuration names). The exit status is the process's.
    fn conflict(&self, a: &str, b: &str) -> Result<Option<Vec<String>>> {
        let drivers = text_merge_args(&self.configured_drivers());
        let mut all: Vec<&str> = drivers.iter().map(String::as_str).collect();
        all.extend([
            "merge-tree",
            "--write-tree",
            "--name-only",
            "--no-messages",
            a,
            b,
        ]);
        let o = self.raw(&all)?;
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
    /// The highest tier a `PASS` naming the head states (`stated_tier`), if any states one.
    pass_tier: Option<u8>,
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
        pass_tier: None,
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
    let named: Vec<&(i64, String, String)> = passes.iter().filter(|p| p.1 == head).collect();
    if let Some((id, _, line)) = named.iter().max_by_key(|p| p.0) {
        j.pass_note = Some(*id);
        j.pass_line = Some(line.clone());
    }
    // D-7: the tier a `PASS` states. One `PASS` without a stated tier ends the question (it
    // counts as before); a floor above the highest stated tier is a finding.
    let tiers: Vec<(i64, Option<u8>)> = named.iter().map(|p| (p.0, stated_tier(&p.2))).collect();
    j.pass_tier = tiers.iter().filter_map(|t| t.1).max();
    if let Some(best) = j.pass_tier {
        if j.floor > best && tiers.iter().all(|t| t.1.is_some()) {
            let note_id = tiers
                .iter()
                .filter(|t| t.1 == Some(best))
                .map(|t| t.0)
                .max();
            j.findings.push(finding(
                "audit_tier_low",
                &arg.card,
                &head,
                json!({"floor": j.floor, "pass_tier": best, "note_id": note_id}),
            ));
        }
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
                   "pass_note": j.pass_note, "pass_tier": j.pass_tier, "specs": j.specs,
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

/// The one body line that counts the commits of the branch per `Moai-Verified` kind and the
/// reverted ones, read from the answer of `trailers::list` (SPEC-20261120-verified-trailer-in-use
/// REQ-5…REQ-7). A commit without the line counts under `none`, as the listing reads it. The line
/// holds counts and fixed words only, and it does not begin with a trailer key. An answer that is
/// an error or lacks a count refuses the body (REQ-6): a body states no figure it did not read.
fn verified_summary(listing: &Result<Value>) -> Result<String> {
    let list = listing.as_ref().map_err(|e| {
        format!("bundle body refused: the Moai-Verified listing of the branch cannot be read: {e}")
    })?;
    let lacks =
        || "bundle body refused: the Moai-Verified listing of the branch lacks a count".to_string();
    let commits = list["commits"].as_array().ok_or_else(lacks)?;
    let mut by_kind = Vec::new();
    for kind in trailers::VERIFIED_KINDS {
        let n = list["counts"][kind].as_u64().ok_or_else(lacks)?;
        by_kind.push(format!("{kind} {n}"));
    }
    let reverted = commits
        .iter()
        .filter(|c| c["reverted"].as_bool() == Some(true))
        .count();
    let n = commits.len();
    Ok(format!(
        "Verification kinds (`Moai-Verified`, {n} commit{}): {} (a commit without the line counts \
         as none); reverted {reverted}.",
        if n == 1 { "" } else { "s" },
        by_kind.join(", ")
    ))
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
    // The commits the branch adds beyond the tip, by `Moai-Verified` kind; an unreadable listing
    // refuses the body before anything is written.
    let summary = verified_summary(&trailers::list(dir, &format!("{base}..{pinned}")))?;
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
         | SPEC | Branch head | Audit before integration |\n| --- | --- | --- |\n",
    );
    for j in &judged {
        let head = j.head.as_deref().unwrap_or("");
        let branch_name = j
            .origin_branch
            .as_deref()
            .map(|n| format!(" (branch `{n}`)"))
            .unwrap_or_default();
        // The text leaves the repository: it names the verdict and the exact head, never the card
        // or the note that holds the verdict (SPEC-20261120-board-bundle-link REQ-4, D-6).
        let audit = match (&j.pass_note, &j.pass_line) {
            (Some(_), Some(line)) => {
                let tier = stated_tier(line)
                    .map(|t| format!(" (tier {t})"))
                    .unwrap_or_default();
                format!("`PASS` on `{head}`{tier}")
            }
            _ => format!("audit floor {}: no verdict needed", j.floor),
        };
        b.push_str(&format!(
            "| {} | `{head}`{branch_name} | {audit} |\n",
            j.specs.join(", ")
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
         {files} file{} changed, {ins} insertion{}(+), {del} deletion{}(-).\n\n{summary}\n\n## Self-Audit\n\n",
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
        // A card number inside a command (`bundle check C-1 C-2`) is a machine-made argument:
        // it becomes a placeholder instead of leaving the repository.
        let cmd = &without_card_numbers(cmd);
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
                "{id}: {acs} AC, {met} met, {withdrawn} withdrawn{tail} — see the SPEC's Result.\n\n"
            ));
        }
    }
    b.push_str("🤖 Generated with [Claude Code](https://claude.com/claude-code)\n");
    // The whole text is read once more, line by line: a card number or a session id that is still
    // in it (a free-text value the leader wrote, a branch name) refuses the body and nothing is
    // written. The refusal names the field, never the value.
    for line in b.lines() {
        if let Some(what) = outward_leak(line) {
            return Err(format!(
                "bundle body refused: {} carries {what}; reword that value (the text leaves the repository with SPEC ids and links only)",
                leak_field(line)
            ));
        }
    }
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
    fn card_numbers_are_found_at_word_starts_only() {
        fn spans(t: &str) -> Vec<&str> {
            card_number_spans(t)
                .into_iter()
                .map(|(a, b)| &t[a..b])
                .collect()
        }
        assert_eq!(spans("C-1"), ["C-1"]);
        assert_eq!(spans("see C-123, C-4)"), ["C-123", "C-4"]);
        assert_eq!(spans("(C-9)"), ["C-9"]);
        assert_eq!(spans("check C-1=0123 C-22"), ["C-1", "C-22"]);
        // The letter may be lower case, as the card reference reader takes it.
        assert_eq!(spans("c-1 and (c-22)"), ["c-1", "c-22"]);
        // Not card numbers: inside a longer word or without digits.
        assert!(spans("AC-1 SPEC-20261119-x xC-1 xc-1 C- C-x c-x AC-12").is_empty());
        assert!(spans("").is_empty() && spans("C").is_empty() && spans("C-").is_empty());
        assert_eq!(
            without_card_numbers("moai-cli bundle check C-1=abc C-22"),
            "moai-cli bundle check <lane>=abc <lane>"
        );
        assert_eq!(without_card_numbers("no card here"), "no card here");
    }

    #[test]
    fn a_lower_case_card_number_is_a_card_number_too() {
        // The card reference reader takes `c-12` as well as `C-12`.
        assert!(outward_leak("lane c-12 only")
            .unwrap()
            .contains("card number"));
        assert!(outward_leak("lane2/c-12").unwrap().contains("card number"));
        assert_eq!(
            without_card_numbers("bundle check c-12 C-13 c-4"),
            "bundle check <lane> <lane> <lane>"
        );
        // Still not card numbers: inside a longer word, no digit, a longer name.
        for text in [
            "abc-12",
            "spec-12",
            "etc-1 x",
            "c-x",
            "c- 1",
            "ac-9",
            "fine text",
        ] {
            assert_eq!(outward_leak(text), None, "{text}");
        }
    }

    #[test]
    fn a_session_id_is_found_by_the_scan_shape_the_repository_uses() {
        // `outward_leak` asks `check_tree::uuid_hits`: the shape itself is tested there.
        for text in [
            "x 00000000-0000-4000-8000-000000000000 y",
            "ABCDEF01-2345-6789-abcd-ef0123456789",
            "session-00000000-0000-4000-8000-000000000000",
        ] {
            assert!(outward_leak(text).unwrap().contains("session id"), "{text}");
        }
        for text in [
            "0000000-0000-4000-8000-000000000000",
            "a 40-hex sha 0123456789abcdef0123456789abcdef01234567",
            "fine text",
        ] {
            assert_eq!(outward_leak(text), None, "{text}");
        }
    }

    #[test]
    fn a_leak_is_named_by_its_field_and_never_by_its_value() {
        let field = |line: &str| leak_field(line);
        assert_eq!(field("- host: lane c-9"), "the Self-Audit field `host`");
        assert_eq!(
            field("- Not done locally: unit (c-9)."),
            "the not-run reasons (`Not done locally`)"
        );
        assert_eq!(
            field("- not run: unit — c-9."),
            "a not-run reason of the run record"
        );
        assert_eq!(
            field("- `moai: c-9` → exit 0."),
            "a command of the run record"
        );
        assert!(field("| SPEC-1 | `abc` | x |").starts_with("the table of heads"));
        assert!(field("Integration: `lane2/c-9` built").contains("integration line"));
        assert!(field("SPEC-1: 2 AC").contains("SPEC result line"));
        // A key that itself carries a card number is not repeated.
        assert_eq!(field("- c-9: x"), "a Self-Audit field");
        assert_eq!(field("other"), "the text");
    }

    #[test]
    fn the_stated_tier_is_read_from_before_the_last_head() {
        let line = |prefix: &str| format!("AUDIT C-1 {prefix}head {H} base {H} — verdict: PASS");
        assert_eq!(stated_tier(&line("(SPEC-20261119-x, tier 1) ")), Some(1));
        assert_eq!(stated_tier(&line("tier 2 ")), Some(2));
        assert_eq!(stated_tier(&line("(tier 0, own agent) ")), Some(0));
        assert_eq!(stated_tier(&line("(a, tier 1; b) ")), Some(1));
        // The same tier twice is one tier.
        assert_eq!(stated_tier(&line("(tier 1, tier 1) ")), Some(1));
        // None, two different tiers, a digit above 2, no digit, a longer number: no tier.
        assert_eq!(stated_tier(&line("(SPEC-20261119-x) ")), None);
        assert_eq!(stated_tier(&line("(tier 1, tier 2) ")), None);
        assert_eq!(stated_tier(&line("(tier 3) ")), None);
        assert_eq!(stated_tier(&line("(tier x) ")), None);
        assert_eq!(stated_tier(&line("(tier 12) ")), None);
        // The word must start a word: `multitier 1` is not it.
        assert_eq!(stated_tier(&line("(multitier 1) ")), None);
        // A `tier 1` after the verdict's own head is not read; neither is a line with no head.
        assert_eq!(
            stated_tier(&format!(
                "AUDIT C-1 head {H} base {H} — verdict: PASS tier 1"
            )),
            None
        );
        assert_eq!(stated_tier("AUDIT C-1 tier 1 without a head"), None);
        // A title that holds the word ` head ` is cut at the last one, as `parse_verdict` does.
        let titled =
            format!("AUDIT C-1 (the head of tier 2, tier 2) head {H} base {H} — verdict: PASS");
        assert_eq!(stated_tier(&titled), Some(2));
    }

    #[test]
    fn merge_driver_names_come_from_the_configuration_lines() {
        let out = "merge.moai-manifest.driver true\nmerge.ours.driver true\nmerge.a b.driver x\nmerge.default x\nmerge..driver y\n";
        assert_eq!(driver_names(out), vec!["moai-manifest", "ours"]);
        let args = text_merge_args(&driver_names(out));
        let want: Vec<String> = vec![
            "-c".into(),
            format!("merge.moai-manifest.driver={TEXT_MERGE}"),
            "-c".into(),
            format!("merge.ours.driver={TEXT_MERGE}"),
        ];
        assert_eq!(args, want);
        // The manifest driver is always overridden, configured or not.
        assert_eq!(text_merge_args(&[]).len(), 2);
    }

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

    /// SPEC-20261120-verified-trailer-in-use REQ-5, REQ-7: the line from a listing answer.
    #[test]
    fn verified_summary_names_the_counts_in_the_pinned_order() {
        let listing = json!({
            "commits": [
                {"sha": "a", "kinds": ["tests"], "reverted": true},
                {"sha": "b", "kinds": ["none"], "reverted": false},
            ],
            "counts": {"tests": 1, "mutants": 0, "e2e": 0, "smoke": 0, "none": 1},
        });
        assert_eq!(
            verified_summary(&Ok(listing)).unwrap(),
            "Verification kinds (`Moai-Verified`, 2 commits): tests 1, mutants 0, e2e 0, smoke 0, \
             none 1 (a commit without the line counts as none); reverted 1."
        );
        let one = json!({
            "commits": [{"sha": "a", "kinds": ["smoke"], "reverted": false}],
            "counts": {"tests": 0, "mutants": 0, "e2e": 0, "smoke": 1, "none": 0},
        });
        let line = verified_summary(&Ok(one)).unwrap();
        assert!(line.contains("1 commit)"), "{line}");
        assert!(!line.starts_with("Moai-"), "{line}");
    }

    /// REQ-6: an error from the listing, or a listing without a count, refuses the body.
    #[test]
    fn verified_summary_refuses_an_unreadable_or_incomplete_listing() {
        let err = verified_summary(&Err("git log failed: bad range".into())).unwrap_err();
        assert!(
            err.starts_with("bundle body refused:") && err.contains("cannot be read"),
            "{err}"
        );
        assert!(err.contains("git log failed: bad range"), "{err}");
        let full = |counts: Value| json!({"commits": [], "counts": counts});
        for broken in [
            json!({"counts": {"tests": 0, "mutants": 0, "e2e": 0, "smoke": 0, "none": 0}}),
            full(json!({"tests": 0, "mutants": 0, "e2e": 0, "smoke": 0})),
            full(json!({"tests": 0, "mutants": 0, "e2e": 0, "smoke": 0, "none": -1})),
            full(json!({"tests": "0", "mutants": 0, "e2e": 0, "smoke": 0, "none": 0})),
            json!({}),
        ] {
            let err = verified_summary(&Ok(broken.clone())).unwrap_err();
            assert!(err.contains("lacks a count"), "{broken}: {err}");
        }
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
