//! `moai release-audit-summary` (SPEC-20261125-release-audit-summary): the command that gathers
//! the verdicts of a release range from the ledger into one pull request comment. Every test runs
//! on a temporary HOME, a temporary ledger and a temporary git repository; the post path runs
//! against a fake `gh` executable. Nothing here reads the real HOME, the real ledger or GitHub.
#![cfg(unix)]

use crate::common;

use common::fake_exe::write_marking_executable;
use common::home::TempHome;
use moai_cli::audit_gate;
use moai_cli::ledger;
use moai_cli::release_audit_summary::{self as summary, Kind, Summary, Verdict};
use rusqlite::Connection;
use serde_json::{json, Value};
use std::cell::Cell;
use std::fs;
use std::io::Write;
use std::path::{Path, PathBuf};
use std::process::{Output, Stdio};

const S_AUTH: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const S_LEAD: &str = "bbbbbbbb-0000-4000-8000-00000000000b";
const S_REV: &str = "cccccccc-0000-4000-8000-00000000000c";
const S_REV2: &str = "dddddddd-0000-4000-8000-00000000000d";
const S_HOLD: &str = "eeeeeeee-0000-4000-8000-00000000000e";

/// The fake `gh`: logs every call, answers `GET /user` as `bot`, serves the comments of the pull
/// request (none before a POST, the comment the POST left after it) and takes a POST once.
const GH_SCRIPT: &str = "#!/bin/sh\n\
printf '%s\\n' \"$*\" >> '@LOG@'\n\
url=''\n\
for a in \"$@\"; do case \"$a\" in https://*) url=\"$a\";; esac; done\n\
if [ \"$3\" = POST ]; then\n\
  cat > '@DIR@/post-body.txt'\n\
  : > '@DIR@/posted'\n\
  printf 'HTTP/2.0 201 Created\\r\\n\\r\\n{\"id\": 9001}'\n\
  exit 0\n\
fi\n\
case \"$url\" in\n\
  */user) printf 'HTTP/2.0 200 OK\\r\\n\\r\\n{\"login\": \"bot\"}';;\n\
  */issues/*/comments*)\n\
    printf 'HTTP/2.0 200 OK\\r\\n\\r\\n'\n\
    if [ -f '@DIR@/posted' ]; then cat '@DIR@/posted-comments.json'; else printf '[]'; fi;;\n\
  *) printf 'HTTP/2.0 404 Not Found\\r\\n\\r\\n{}'; exit 1;;\n\
esac\n";

fn git(dir: &Path, args: &[&str]) -> String {
    let mut all = vec![
        "-c",
        "user.name=fixture",
        "-c",
        "user.email=fixture@example.invalid",
        "-c",
        "commit.gpgsign=false",
    ];
    all.extend_from_slice(args);
    common::git::git(dir, &all)
}

fn envelope(out: &Output) -> Value {
    serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
        panic!(
            "stdout is not JSON ({e}): {} {}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    })
}

fn result(out: &Output) -> Value {
    assert_eq!(
        out.status.code(),
        Some(0),
        "{} {}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    let v = envelope(out);
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

/// The message of a refused call (exit 1, `ok` false).
fn refusal(out: &Output) -> String {
    assert_eq!(
        out.status.code(),
        Some(1),
        "{} {}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    let v = envelope(out);
    assert_eq!(v["ok"], json!(false), "{v}");
    v["error"]["message"].as_str().unwrap_or("").to_string()
}

struct World {
    home: TempHome,
    project: PathBuf,
    p: String,
    /// The fake `gh`, its call log and its answers live outside the HOME.
    host: tempfile::TempDir,
    gh: PathBuf,
    log: PathBuf,
    n: Cell<u32>,
}

impl World {
    fn new() -> Self {
        let home = TempHome::new();
        let project = home.project("demo");
        ledger::register(home.path(), "demo").unwrap();
        git(&project, &["init", "-q", "-b", "main"]);
        let host = tempfile::Builder::new()
            .prefix("moai-test-host-")
            .tempdir()
            .unwrap();
        let dir = host.path().to_path_buf();
        let log = dir.join("gh-calls.log");
        let script = GH_SCRIPT
            .replace("@LOG@", &log.display().to_string())
            .replace("@DIR@", &dir.display().to_string());
        let gh = dir.join("gh-stub");
        common::write_executable(&gh, &script);
        let w = World {
            p: project.to_str().unwrap().to_string(),
            home,
            project,
            host,
            gh,
            log,
            n: Cell::new(0),
        };
        w.commit("init", None);
        w
    }

    fn bump(&self) -> u32 {
        self.n.set(self.n.get() + 1);
        self.n.get()
    }

    // ---- git ---------------------------------------------------------------------------------

    /// One commit on `main` that adds its own file; `session` becomes the `Moai-Session`
    /// trailer in the last paragraph.
    fn commit(&self, subject: &str, session: Option<&str>) -> String {
        let n = self.bump();
        let file = format!("f{n}.txt");
        fs::write(self.project.join(&file), format!("{n}\n")).unwrap();
        git(&self.project, &["add", &file]);
        let trailer = session.map(|s| format!("Moai-Session: {s}"));
        let mut args = vec!["commit", "-q", "-m", subject];
        if let Some(t) = &trailer {
            args.extend(["-m", t]);
        }
        git(&self.project, &args);
        self.head()
    }

    fn head(&self) -> String {
        git(&self.project, &["rev-parse", "HEAD"])
    }

    /// A lane of commits on a side branch, joined onto `main` by a real merge. Returns the merge
    /// commit and the last commit of the lane (its second parent).
    fn merged_lane(&self, subject: &str, lane: usize, session: &str) -> (String, String) {
        let n = self.bump();
        git(&self.project, &["switch", "-q", "-c", &format!("lane{n}")]);
        for i in 0..lane {
            self.commit(&format!("lane commit {i}"), Some(session));
        }
        let second = self.head();
        git(&self.project, &["switch", "-q", "main"]);
        git(
            &self.project,
            &["merge", "-q", "--no-ff", "-m", subject, &format!("lane{n}")],
        );
        (self.head(), second)
    }

    /// The first-parent commits of `base..head`, oldest first, as git lists them.
    fn first_parent(&self, base: &str, head: &str) -> Vec<String> {
        git(
            &self.project,
            &[
                "rev-list",
                "--first-parent",
                "--reverse",
                &format!("{base}..{head}"),
            ],
        )
        .lines()
        .map(str::to_string)
        .collect()
    }

    /// `count` direct commits in one `git fast-import` run, each with a long subject.
    fn fast_history(&self, count: usize) {
        let from = self.head();
        let mut stream = String::new();
        for i in 0..count {
            let message = format!("a direct commit with a long subject, number {i:06}");
            stream.push_str(&format!(
                "commit refs/heads/main\ncommitter fixture <fixture@example.invalid> {} +0000\ndata {}\n{message}\n",
                1_700_000_000 + i,
                message.len()
            ));
            if i == 0 {
                stream.push_str(&format!("from {from}\n"));
            }
            stream.push('\n');
        }
        let mut child = common::git::command()
            .args(["fast-import", "--quiet"])
            .current_dir(&self.project)
            .stdin(Stdio::piped())
            .spawn()
            .expect("spawn git fast-import");
        child
            .stdin
            .take()
            .unwrap()
            .write_all(stream.as_bytes())
            .unwrap();
        assert!(child.wait().unwrap().success());
        git(&self.project, &["reset", "-q", "--hard", "main"]);
    }

    // ---- the ledger --------------------------------------------------------------------------

    fn conn(&self) -> Connection {
        Connection::open(ledger::ledger_path(self.home.path(), "demo").unwrap()).unwrap()
    }

    fn card(&self) -> i64 {
        let conn = self.conn();
        conn.execute(
            "INSERT INTO card (title, created_by, created_at, updated_at) \
             VALUES ('a fixture card', 'test', 0, 0)",
            [],
        )
        .unwrap();
        conn.last_insert_rowid()
    }

    fn event(&self, actor: &str, card: i64, kind: &str, payload: Value) -> i64 {
        let conn = self.conn();
        let key = format!("test-{kind}-{}", self.bump());
        conn.execute(
            "INSERT INTO event (event_key, at, actor, card_id, kind, payload) \
             VALUES (?1, unixepoch(), ?2, ?3, ?4, ?5)",
            rusqlite::params![key, actor, card, kind, payload.to_string()],
        )
        .unwrap();
        conn.last_insert_rowid()
    }

    /// One `card.verdict` event as `card verdict` writes it, with a good receipt of the head.
    fn verdict(&self, card: i64, head: &str, pass: bool, reviewer: &str) -> i64 {
        self.event(
            reviewer,
            card,
            "card.verdict",
            json!({
                "head": head, "base": head, "verdict": if pass { "pass" } else { "fail" },
                "tier": 0, "evidence": "the fixture's evidence",
                "receipt": {"head": head, "steps": [{"name": "test", "exit": 0}]},
                "reviewer": reviewer,
            }),
        )
    }

    /// The `card.landed` event of `card` on `head`, naming the verdict event it relied on.
    fn landed(&self, card: i64, head: &str, relied: Option<(i64, &str)>) -> i64 {
        let mut payload = json!({
            "head": head, "tip": head, "generation": 1, "by": "leader", "report": null,
            "note": null, "via": "finish", "no_code": false,
        });
        if let Some((event, reviewer)) = relied {
            payload["verdict"] =
                json!({"event": event, "reviewer": reviewer, "tier": 0, "lander": S_LEAD});
        }
        self.event(S_LEAD, card, "card.landed", payload)
    }

    /// The `card.land_started` event: the tip the branch was moved from.
    fn started(&self, card: i64, head: &str, tip: &str) {
        self.event(
            S_LEAD,
            card,
            "card.land_started",
            json!({"branch": "main", "tip": tip, "head": head, "by": "leader"}),
        );
    }

    /// `holder` held a reservation of the card.
    fn holder(&self, card: i64, holder: &str) {
        let conn = self.conn();
        conn.execute(
            "INSERT INTO reservation (card_id, generation, holder, status, acquired_at, \
             lease_until, updated_at) VALUES (?1, \
             (SELECT COALESCE(MAX(generation), 0) + 1 FROM reservation WHERE card_id = ?1), \
             ?2, 'released', 0, 0, 0)",
            rusqlite::params![card, holder],
        )
        .unwrap();
    }

    /// A card that landed `head` on a verdict of `reviewer`, the way `card land` records it.
    fn passed(&self, head: &str, reviewer: &str) -> i64 {
        let card = self.card();
        let event = self.verdict(card, head, true, reviewer);
        self.landed(card, head, Some((event, reviewer)));
        card
    }

    // ---- the command -------------------------------------------------------------------------

    fn run(&self, args: &[&str]) -> Output {
        let mut cmd = self.home.command(args);
        cmd.env("MOAI_FORGE_GH_BIN", &self.gh);
        cmd.output().expect("run moai")
    }

    fn preview(&self, base: &str, head: &str, extra: &[&str]) -> Output {
        let mut args = vec![
            "release-audit-summary",
            &self.p,
            "--base",
            base,
            "--head",
            head,
        ];
        args.extend_from_slice(extra);
        self.run(&args)
    }

    fn plan(&self, base: &str, head: &str) -> Value {
        result(&self.preview(base, head, &[]))
    }

    fn post(&self, base: &str, head: &str, digest: &str, extra: &[&str]) -> Output {
        let mut args = vec![
            "--post",
            "--pr",
            "7",
            "--repo",
            "acme/widgets",
            "--plan-digest",
            digest,
            "--permission-reference",
            "user-approved-test",
        ];
        args.extend_from_slice(extra);
        self.preview(base, head, &args)
    }

    /// What the fake `gh` answers once a POST was made: a comment of its own account with the
    /// marker of `head`.
    fn comment_left_by_the_post(&self, head: &str) {
        let comments = json!([{
            "id": 9001, "body": format!("summary\n\n{}", summary::marker(head)),
            "user": {"login": "bot"}, "author_association": "CONTRIBUTOR",
            "updated_at": "2026-11-03T00:00:00Z"
        }]);
        fs::write(
            self.host.path().join("posted-comments.json"),
            comments.to_string(),
        )
        .unwrap();
    }

    fn calls(&self) -> Vec<String> {
        fs::read_to_string(&self.log)
            .map(|t| t.lines().map(str::to_string).collect())
            .unwrap_or_default()
    }

    fn posts(&self) -> usize {
        self.calls()
            .iter()
            .filter(|c| c.contains("-X POST"))
            .count()
    }

    fn posted_body(&self) -> String {
        let raw =
            fs::read_to_string(self.host.path().join("post-body.txt")).expect("a POST was sent");
        let value: Value = serde_json::from_str(&raw).unwrap();
        value["body"].as_str().unwrap().to_string()
    }
}

/// The parsed summary the plan carries.
fn read(plan: &Value) -> Summary {
    let text = plan["text"].as_str().expect("the plan has a text");
    summary::parse(text).unwrap_or_else(|e| panic!("the text does not parse ({e}): {text}"))
}

fn pass_verdicts(line: &Kind) -> &[Verdict] {
    match line {
        Kind::Pass(v) => v,
        other => panic!("not a PASS line: {other:?}"),
    }
}

fn session(s: &Summary, index: usize) -> &str {
    &s.sessions[index]
}

/// AC-1 (REQ-1, REQ-2): one line per first-parent commit, oldest first; a direct commit with no
/// landing row and a card with no verdict are lines, none is dropped.
#[test]
fn one_line_per_first_parent_commit_and_none_is_dropped() {
    let w = World::new();
    let base = w.head();
    let c1 = w.commit("card one", Some(S_AUTH));
    let first = w.passed(&c1, S_REV);
    let c2 = w.commit("a direct commit", None);
    let c3 = w.commit("card two", Some(S_AUTH));
    let second = w.card();
    w.landed(second, &c3, None);

    let plan = w.plan(&base, &c3);
    let s = read(&plan);
    // The same commits as git lists them, oldest first.
    let listed: Vec<&str> = s.lines.iter().map(|l| l.sha.as_str()).collect();
    assert_eq!(listed, w.first_parent(&base, &c3), "{}", plan["text"]);
    assert_eq!(listed, [c1.as_str(), c2.as_str(), c3.as_str()]);
    assert_eq!(
        (s.base.as_str(), s.head.as_str()),
        (base.as_str(), c3.as_str())
    );
    // Each line names the first parent of its commit.
    assert_eq!(s.lines[0].parent, base[..12]);
    assert_eq!(s.lines[1].parent, c1[..12]);
    assert_eq!(s.lines[2].parent, c2[..12]);
    // The PASS line, the direct commit (no card) and the landing with no verdict.
    assert_eq!(pass_verdicts(&s.lines[0].kind)[0].card, first);
    assert!(
        matches!(&s.lines[1].kind, Kind::None { cards, title } if cards.is_empty() && title == "a direct commit"),
        "{:?}",
        s.lines[1].kind
    );
    assert!(
        matches!(&s.lines[2].kind, Kind::None { cards, .. } if cards == &vec![second]),
        "{:?}",
        s.lines[2].kind
    );
    assert_eq!(plan["landings"], json!(3));
    assert_eq!(
        plan["counts"],
        json!({"pass": 1, "abs": 0, "none": 2, "fail": 0})
    );
    assert_eq!(plan["complete"], json!(false));
    // Every first-parent commit has exactly one `L` line in the text itself.
    let text = plan["text"].as_str().unwrap();
    assert_eq!(text.lines().filter(|l| l.starts_with("L ")).count(), 3);
}

/// AC-1 (REQ-2): a landing with no verdict is listed by name: the commit and the card, in the
/// text and in `missing`; `--require-complete` turns it into an error.
#[test]
fn a_landing_without_a_verdict_is_listed_by_name() {
    let w = World::new();
    let base = w.head();
    let c1 = w.commit("a card with no verdict", Some(S_AUTH));
    let card = w.card();
    w.landed(card, &c1, None);

    let plan = w.plan(&base, &c1);
    let missing = plan["missing"].as_array().expect("missing is a list");
    assert_eq!(missing.len(), 1, "{plan}");
    assert_eq!(missing[0]["commit"], json!(c1));
    assert_eq!(missing[0]["state"], json!("NONE"));
    assert_eq!(missing[0]["cards"], json!([format!("C-{card}")]));
    assert_eq!(missing[0]["why"], json!("no_verdict"), "{plan}");
    let text = plan["text"].as_str().unwrap();
    assert!(
        text.contains(&format!(
            "L {c1} p={} NONE c={card} | a card with no verdict",
            &base[..12]
        )),
        "{text}"
    );

    // A direct commit is named too: no card, `no_landing` as the reason.
    let c2 = w.commit("a direct commit", None);
    let plan = w.plan(&base, &c2);
    let missing = plan["missing"].as_array().unwrap();
    assert_eq!(missing.len(), 2, "{plan}");
    assert_eq!(missing[1]["commit"], json!(c2));
    assert_eq!(missing[1]["cards"], json!([]));
    assert_eq!(missing[1]["why"], json!("no_landing_row"));

    // `--require-complete` makes the incomplete summary an error that names a landing.
    let why = refusal(&w.preview(&base, &c2, &["--require-complete"]));
    assert!(why.contains(&c1[..12]), "{why}");
    // A complete summary passes it.
    let w = World::new();
    let base = w.head();
    let c1 = w.commit("card", Some(S_AUTH));
    w.passed(&c1, S_REV);
    let plan = result(&w.preview(&base, &c1, &["--require-complete"]));
    assert_eq!(plan["complete"], json!(true));
    assert_eq!(plan["missing"], json!([]));
}

/// AC-1 (REQ-3, REQ-4): a commit that closed several cards is one line that lists every one; it
/// is a PASS line only while every card has a passing verdict.
#[test]
fn a_batch_commit_lists_every_card_it_closed_on_one_line() {
    let w = World::new();
    let base = w.head();
    // One commit, two cards, both passing.
    let c1 = w.commit("a batch of two", Some(S_AUTH));
    let a = w.passed(&c1, S_REV);
    let b = w.passed(&c1, S_REV2);
    // One commit, two cards, the second without a verdict.
    let c2 = w.commit("a batch with a hole", Some(S_AUTH));
    let c = w.passed(&c2, S_REV);
    let d = w.card();
    w.landed(d, &c2, None);
    // A slice landed twice on the same head names its card once.
    let c3 = w.commit("a card landed twice", Some(S_AUTH));
    let e = w.passed(&c3, S_REV);
    w.landed(e, &c3, None);

    let plan = w.plan(&base, &c3);
    let s = read(&plan);
    assert_eq!(s.lines.len(), 3, "{}", plan["text"]);
    let first = pass_verdicts(&s.lines[0].kind);
    assert_eq!(first.iter().map(|v| v.card).collect::<Vec<_>>(), [a, b]);
    // Not a PASS while one card has none: NONE, and both cards are named.
    assert!(
        matches!(&s.lines[1].kind, Kind::None { cards, .. } if cards == &vec![c, d]),
        "{:?}",
        s.lines[1].kind
    );
    let third = pass_verdicts(&s.lines[2].kind);
    assert_eq!(third.iter().map(|v| v.card).collect::<Vec<_>>(), [e]);
    assert_eq!(
        plan["counts"],
        json!({"pass": 2, "abs": 0, "none": 1, "fail": 0})
    );
}

/// AC-1 (REQ-1): a lane that was fast-forwarded in is several first-parent commits and one
/// landing: the commits between the tip the landing started from and the landed head are `ABS`
/// lines of it, and nothing else is absorbed.
#[test]
fn a_fast_forwarded_lane_absorbs_only_its_own_commits() {
    let w = World::new();
    let base = w.head();
    // A lane right after the base: its start row names the base.
    let x1 = w.commit("lane x one", Some(S_AUTH));
    let x2 = w.commit("lane x two", Some(S_AUTH));
    let cx = w.passed(&x2, S_REV);
    w.started(cx, &x2, &base);
    // A direct commit, then a lane of three that started from it.
    let d = w.commit("a direct commit", None);
    let l1 = w.commit("lane l one", Some(S_AUTH));
    let l2 = w.commit("lane l two", Some(S_AUTH));
    let l3 = w.commit("lane l three", Some(S_AUTH));
    let cl = w.passed(&l3, S_REV);
    w.started(cl, &l3, &d);
    let e = w.commit("a direct commit again", None);
    // A lane with no start row: its first commit stays a line of its own (D-7).
    let m1 = w.commit("lane m one", Some(S_AUTH));
    let m2 = w.commit("lane m two", Some(S_AUTH));
    w.passed(&m2, S_REV);
    // A lane whose start row names a tip that is not before it: nothing is absorbed.
    let n1 = w.commit("lane n one", Some(S_AUTH));
    let n2 = w.commit("lane n two", Some(S_AUTH));
    let cn = w.passed(&n2, S_REV);
    let z = w.commit("the last direct commit", None);
    w.started(cn, &n2, &z);
    // A start row whose tip is an unrelated commit absorbs nothing either.
    let o1 = w.commit("lane o one", Some(S_AUTH));
    let o2 = w.commit("lane o two", Some(S_AUTH));
    let co = w.passed(&o2, S_REV);
    w.started(co, &o2, &"f".repeat(40));

    let plan = w.plan(&base, &o2);
    let s = read(&plan);
    let at = |sha: &str| s.lines.iter().position(|l| l.sha == sha).unwrap();
    let abs = |sha: &str, target: &str| {
        assert!(
            matches!(&s.lines[at(sha)].kind, Kind::Abs { target: t } if t == &target[..12]),
            "{sha}: {:?}",
            s.lines[at(sha)].kind
        );
    };
    abs(&x1, &x2);
    abs(&l1, &l3);
    abs(&l2, &l3);
    assert!(matches!(s.lines[at(&x2)].kind, Kind::Pass(_)));
    assert!(matches!(s.lines[at(&l3)].kind, Kind::Pass(_)));
    // The commits around the lanes are not absorbed.
    for direct in [&d, &e, &z] {
        assert!(
            matches!(&s.lines[at(direct)].kind, Kind::None { cards, .. } if cards.is_empty()),
            "{direct}: {:?}",
            s.lines[at(direct)].kind
        );
    }
    // No start row, a tip after the lane, an unrelated tip: each first commit is its own line.
    for own in [&m1, &n1, &o1] {
        assert!(
            matches!(&s.lines[at(own)].kind, Kind::None { cards, .. } if cards.is_empty()),
            "{own}: {:?}",
            s.lines[at(own)].kind
        );
    }
    assert_eq!(
        plan["counts"],
        json!({"pass": 5, "abs": 3, "none": 6, "fail": 0}),
        "{}",
        plan["text"]
    );
    assert_eq!(s.lines.len(), 14);
    let listed: Vec<&str> = s.lines.iter().map(|l| l.sha.as_str()).collect();
    assert_eq!(listed, w.first_parent(&base, &o2));
}

/// AC-2 (REQ-5, REQ-6): a verdict whose reviewer wrote a commit of the head or held the card is
/// no pass; a PASS line lists the sessions that wrote or held its cards.
#[test]
fn a_reviewer_who_wrote_or_held_the_card_is_not_a_pass_line() {
    let w = World::new();
    let base = w.head();
    // The author reviews their own work.
    let c1 = w.commit("written and reviewed by one session", Some(S_AUTH));
    let own = w.passed(&c1, S_AUTH);
    // The reviewer held the card.
    let c2 = w.commit("reviewed by its holder", Some(S_AUTH));
    let held = w.passed(&c2, S_HOLD);
    w.holder(held, S_HOLD);
    // An independent reviewer, with an author and a holder on the card.
    let c3 = w.commit("reviewed by another session", Some(S_AUTH));
    let fine = w.passed(&c3, S_REV);
    w.holder(fine, S_HOLD);

    let plan = w.plan(&base, &c3);
    let s = read(&plan);
    assert!(
        matches!(&s.lines[0].kind, Kind::None { cards, .. } if cards == &vec![own]),
        "{:?}",
        s.lines[0].kind
    );
    assert!(
        matches!(&s.lines[1].kind, Kind::None { cards, .. } if cards == &vec![held]),
        "{:?}",
        s.lines[1].kind
    );
    let verdicts = pass_verdicts(&s.lines[2].kind);
    assert_eq!(verdicts.len(), 1);
    let v = &verdicts[0];
    assert_eq!(v.card, fine);
    assert_eq!(session(&s, v.reviewer), S_REV);
    let excluded: Vec<&str> = v.excluded.iter().map(|&i| session(&s, i)).collect();
    assert!(excluded.contains(&S_AUTH), "{excluded:?}");
    assert!(excluded.contains(&S_HOLD), "{excluded:?}");
    assert!(!excluded.contains(&S_REV), "{excluded:?}");
    assert!(v.counts(&s));
    // The reasons are in `missing`, not in the comment.
    let why: Vec<&str> = plan["missing"]
        .as_array()
        .unwrap()
        .iter()
        .map(|m| m["why"].as_str().unwrap())
        .collect();
    assert_eq!(why, ["self_verdict", "self_verdict"], "{plan}");
}

/// AC-2 (REQ-5): a landing row that names no verdict is judged by the rules `card land` applies
/// to the verdict events of its card, on the commit or on the audited second parent of a clean
/// join; a row that names the verdict it relied on reads that event.
#[test]
fn a_pass_event_without_a_relied_row_is_judged_by_the_land_rules() {
    let w = World::new();
    let base = w.head();
    // A verdict event on the commit, a row that names none.
    let c1 = w.commit("verdict on the commit", Some(S_AUTH));
    let one = w.card();
    let e1 = w.verdict(one, &c1, true, S_REV);
    w.landed(one, &c1, None);
    // A verdict event on another head only.
    let c2 = w.commit("verdict on another head", Some(S_AUTH));
    let two = w.card();
    w.verdict(two, &base, true, S_REV);
    w.landed(two, &c2, None);
    // No verdict event at all.
    let c3 = w.commit("no verdict event", Some(S_AUTH));
    let three = w.card();
    w.landed(three, &c3, None);
    // A clean join of an audited lane: the verdict is on the second parent, and the row names
    // nothing.
    let (m1, a1) = w.merged_lane("merge(next-flow): lane four", 2, S_AUTH);
    let four = w.card();
    let e4 = w.verdict(four, &a1, true, S_REV);
    w.landed(four, &m1, None);
    // The same join, the row names the verdict event it relied on.
    let (m2, a2) = w.merged_lane("merge(next-flow): lane five", 1, S_AUTH);
    let five = w.card();
    let e5 = w.verdict(five, &a2, true, S_REV2);
    w.landed(five, &m2, Some((e5, S_REV2)));
    // A reviewer's FAIL stands even when another reviewer passed the head.
    let c6 = w.commit("one fail, one pass", Some(S_AUTH));
    let six = w.card();
    w.verdict(six, &c6, false, S_REV2);
    w.verdict(six, &c6, true, S_REV);
    w.landed(six, &c6, None);

    let plan = w.plan(&base, &c6);
    let s = read(&plan);
    let listed: Vec<&str> = s.lines.iter().map(|l| l.sha.as_str()).collect();
    assert_eq!(listed, w.first_parent(&base, &c6), "{}", plan["text"]);
    let find = |sha: &str| &s.lines[s.lines.iter().position(|l| l.sha == sha).unwrap()].kind;
    let v1 = &pass_verdicts(find(&c1))[0];
    assert_eq!((v1.card, v1.event), (one, e1));
    assert_eq!(v1.head, c1[..12]);
    assert!(matches!(find(&c2), Kind::None { cards, .. } if cards == &vec![two]));
    assert!(matches!(find(&c3), Kind::None { cards, .. } if cards == &vec![three]));
    let v4 = &pass_verdicts(find(&m1))[0];
    assert_eq!((v4.card, v4.event), (four, e4));
    assert_eq!(v4.head, a1[..12], "the audited head is the second parent");
    let v5 = &pass_verdicts(find(&m2))[0];
    assert_eq!((v5.card, v5.event), (five, e5));
    assert_eq!(v5.head, a2[..12]);
    assert!(
        matches!(find(&c6), Kind::Fail { why, cards, .. } if why == "verdict_fail" && cards == &vec![six]),
        "{:?}",
        find(&c6)
    );
    // The lane commits are not first-parent commits of the range.
    assert!(!listed.contains(&a1.as_str()) && !listed.contains(&a2.as_str()));
}

/// A merge that carries more than the join of its parents is not a clean join: an audit of its
/// second parent says nothing about the added change, so the line has no passing verdict.
#[test]
fn a_merge_that_adds_a_change_is_not_a_clean_join() {
    let w = World::new();
    let base = w.head();
    let n = w.bump();
    git(&w.project, &["switch", "-q", "-c", &format!("lane{n}")]);
    let lane = w.commit("lane commit", Some(S_AUTH));
    git(&w.project, &["switch", "-q", "main"]);
    git(
        &w.project,
        &["merge", "-q", "--no-ff", "--no-commit", &format!("lane{n}")],
    );
    fs::write(w.project.join("added-by-the-merge.txt"), "unaudited\n").unwrap();
    git(&w.project, &["add", "added-by-the-merge.txt"]);
    git(
        &w.project,
        &["commit", "-q", "-m", "merge(next-flow): lane with an extra"],
    );
    let merge = w.head();
    let card = w.card();
    w.verdict(card, &lane, true, S_REV);
    w.landed(card, &merge, None);

    let plan = w.plan(&base, &merge);
    let s = read(&plan);
    let line = s.lines.last().unwrap();
    assert_eq!(line.sha, merge);
    assert!(
        matches!(&line.kind, Kind::None { cards, .. } if cards == &vec![card]),
        "{:?}\n{}",
        line.kind,
        plan["text"]
    );
    assert_eq!(plan["missing"][0]["why"], "verdict_other_head", "{plan}");
}

/// AC-2 (REQ-2): a FAIL recorded after the landing marks the line `FAIL`; the newest verdict of
/// a reviewer is the one that counts.
#[test]
fn a_fail_recorded_after_the_landing_marks_the_line_fail() {
    let w = World::new();
    let base = w.head();
    // Passed by one reviewer, failed by another after the landing.
    let c1 = w.commit("failed later by another reviewer", Some(S_AUTH));
    let one = w.passed(&c1, S_REV);
    w.verdict(one, &c1, false, S_REV2);
    // Failed first, passed by the same reviewer afterwards: the PASS is the newest.
    let c2 = w.commit("failed then passed", Some(S_AUTH));
    let two = w.card();
    w.verdict(two, &c2, false, S_REV);
    let pass = w.verdict(two, &c2, true, S_REV);
    w.landed(two, &c2, Some((pass, S_REV)));
    // Passed, then failed by the same reviewer after the landing.
    let c3 = w.commit("failed later by the same reviewer", Some(S_AUTH));
    let three = w.passed(&c3, S_REV);
    w.verdict(three, &c3, false, S_REV);

    let plan = w.plan(&base, &c3);
    let s = read(&plan);
    assert!(
        matches!(&s.lines[0].kind, Kind::Fail { why, cards, .. } if why == "verdict_fail" && cards == &vec![one]),
        "{:?}",
        s.lines[0].kind
    );
    assert_eq!(pass_verdicts(&s.lines[1].kind)[0].card, two);
    assert!(
        matches!(&s.lines[2].kind, Kind::Fail { cards, .. } if cards == &vec![three]),
        "{:?}",
        s.lines[2].kind
    );
    assert_eq!(
        plan["counts"],
        json!({"pass": 1, "abs": 0, "none": 0, "fail": 2})
    );
    let states: Vec<&str> = plan["missing"]
        .as_array()
        .unwrap()
        .iter()
        .map(|m| m["state"].as_str().unwrap())
        .collect();
    assert_eq!(states, ["FAIL", "FAIL"]);
    // A FAIL line is not complete, whatever its reason.
    assert_eq!(plan["complete"], json!(false));
}

/// AC-3 (REQ-7): the default run prints the comment, its digest and the missing landings, and
/// sends nothing: the fake `gh` is never started and the ledger does not change.
#[test]
fn the_dry_run_prints_the_comment_and_digest_and_sends_nothing() {
    let w = World::new();
    let base = w.head();
    let c1 = w.commit("passed", Some(S_AUTH));
    w.passed(&c1, S_REV);
    let c2 = w.commit("a direct commit", None);
    // The program the command would send with is a fake that leaves a marker when it starts.
    let marker = w.host.path().join("gh-started");
    let gh = w.host.path().join("gh-marking");
    write_marking_executable(&gh, &marker, "exit 0\n");
    let ledger_before = fs::read(ledger::ledger_path(w.home.path(), "demo").unwrap()).unwrap();
    let project = w.p.clone();
    let mut cmd = w.home.command(&[
        "release-audit-summary",
        &project,
        "--base",
        &base,
        "--head",
        &c2,
    ]);
    cmd.env("MOAI_FORGE_GH_BIN", &gh);
    let plan = result(&cmd.output().unwrap());

    assert!(!marker.exists(), "the dry run started gh");
    assert!(!w.log.exists(), "the dry run called the stub gh");
    assert_eq!(
        fs::read(ledger::ledger_path(w.home.path(), "demo").unwrap()).unwrap(),
        ledger_before,
        "the dry run changed the ledger"
    );
    assert_eq!(plan["writes"], json!(false));
    assert_eq!(plan["network"], json!("none"));
    assert_eq!(plan["landings"], json!(2));
    assert_eq!(plan["complete"], json!(false));
    assert_eq!(plan["missing"].as_array().unwrap().len(), 1);
    // The text is the comment as it is posted, marker included, and the digest covers it.
    let text = plan["text"].as_str().unwrap();
    assert!(text.starts_with("RELEASE-AUDIT-SUMMARY head="), "{text}");
    assert!(text.ends_with(&summary::marker(&c2)), "{text}");
    assert_eq!(plan["marker"], json!(summary::marker(&c2)));
    assert_eq!(plan["plan_digest"], json!(summary::plan_digest(text)));
    assert_eq!(plan["bytes"], json!(text.len()));
    // Without `--post` the other post flags change nothing.
    let again = result(&w.preview(
        &base,
        &c2,
        &[
            "--pr",
            "7",
            "--repo",
            "acme/widgets",
            "--plan-digest",
            "sha256:0",
        ],
    ));
    assert_eq!(again["plan_digest"], plan["plan_digest"]);
    assert_eq!(w.posts(), 0);
}

/// AC-3 (REQ-8, REQ-9): `--post` needs the flag, a PR, a repository, the plan's digest, a
/// permission reference and a complete summary; a refusal never reaches `gh`.
#[test]
fn posting_needs_flag_pr_digest_permission_and_a_complete_summary() {
    let w = World::new();
    let base = w.head();
    let c1 = w.commit("passed", Some(S_AUTH));
    w.passed(&c1, S_REV);
    let digest = w.plan(&base, &c1)["plan_digest"]
        .as_str()
        .unwrap()
        .to_string();
    let full = [
        "--post",
        "--pr",
        "7",
        "--repo",
        "acme/widgets",
        "--plan-digest",
        digest.as_str(),
        "--permission-reference",
        "user-approved-test",
    ];
    // Each value left out is refused by name.
    for (left_out, name) in [
        ("--pr", "--pr"),
        ("--repo", "--repo"),
        ("--plan-digest", "--plan-digest"),
        ("--permission-reference", "permission reference"),
    ] {
        let mut args: Vec<&str> = Vec::new();
        let mut i = 0;
        while i < full.len() {
            if full[i] == left_out {
                i += 2;
            } else {
                args.push(full[i]);
                i += 1;
            }
        }
        let why = refusal(&w.preview(&base, &c1, &args));
        assert!(why.contains(name), "{left_out}: {why}");
    }
    // A permission reference that says nothing, a digest that is not the plan's, a malformed
    // repository.
    for reference in ["", "null", "placeholder"] {
        let mut args = full.to_vec();
        *args.last_mut().unwrap() = reference;
        let why = refusal(&w.preview(&base, &c1, &args));
        assert!(why.contains("permission reference"), "{reference:?}: {why}");
    }
    let mut args = full.to_vec();
    args[6] = "sha256:0000000000000000000000000000000000000000000000000000000000000000";
    let why = refusal(&w.preview(&base, &c1, &args));
    assert!(why.contains("stale or altered plan"), "{why}");
    let mut args = full.to_vec();
    args[4] = "not-a-repository";
    assert!(!refusal(&w.preview(&base, &c1, &args)).is_empty());
    assert!(
        w.calls().is_empty(),
        "a refusal reached gh: {:?}",
        w.calls()
    );

    // An incomplete summary is refused without `--allow-incomplete`, and posted with it.
    let c2 = w.commit("a direct commit", None);
    let digest2 = w.plan(&base, &c2)["plan_digest"]
        .as_str()
        .unwrap()
        .to_string();
    let why = refusal(&w.post(&base, &c2, &digest2, &[]));
    assert!(why.contains("--allow-incomplete"), "{why}");
    assert!(w.calls().is_empty(), "an incomplete summary reached gh");
    w.comment_left_by_the_post(&c2);
    let posted = result(&w.post(&base, &c2, &digest2, &["--allow-incomplete"]));
    assert_eq!(posted["state"], json!("posted"), "{posted}");
    assert_eq!(w.posts(), 1);
}

/// AC-3 (REQ-8): a complete summary is posted once with exactly the planned bytes; the second
/// run finds its marker on the pull request and posts nothing.
#[test]
fn one_post_with_the_planned_bytes_then_nothing() {
    let w = World::new();
    let base = w.head();
    let c1 = w.commit("one", Some(S_AUTH));
    w.passed(&c1, S_REV);
    let c2 = w.commit("two", Some(S_AUTH));
    w.passed(&c2, S_REV2);
    let plan = w.plan(&base, &c2);
    assert_eq!(plan["complete"], json!(true), "{plan}");
    let digest = plan["plan_digest"].as_str().unwrap();
    w.comment_left_by_the_post(&c2);

    let first = result(&w.post(&base, &c2, digest, &[]));
    assert_eq!(first["state"], json!("posted"), "{first}");
    assert_eq!(first["comment_id"], json!(9001));
    assert_eq!(first["plan_digest"], json!(digest));
    assert_eq!(w.posts(), 1);
    // The body that left is the planned text, byte for byte, and the gate reads it back.
    let body = w.posted_body();
    assert_eq!(body, plan["text"].as_str().unwrap());
    assert_eq!(summary::plan_digest(&body), digest);
    assert!(body.ends_with(&summary::marker(&c2)));
    let read_back = summary::parse(&body).expect("the posted comment parses");
    assert_eq!(read_back.lines.len(), 2);
    // The post went to the comments of pull request 7 of the named repository.
    let post = w
        .calls()
        .into_iter()
        .find(|c| c.contains("-X POST"))
        .unwrap();
    assert!(
        post.contains("/repos/acme/widgets/issues/7/comments"),
        "{post}"
    );

    // The second run posts nothing.
    let second = result(&w.post(&base, &c2, digest, &[]));
    assert_eq!(second["state"], json!("already_posted"), "{second}");
    assert_eq!(w.posts(), 1, "{:?}", w.calls());
}

/// AC-3 (REQ-9): a ledger or a range that changed after the plan posts nothing.
#[test]
fn a_changed_ledger_or_range_after_the_plan_posts_nothing() {
    let w = World::new();
    let base = w.head();
    let c1 = w.commit("one", Some(S_AUTH));
    let card = w.passed(&c1, S_REV);
    let c2 = w.commit("two", Some(S_AUTH));
    w.passed(&c2, S_REV);
    let digest = w.plan(&base, &c2)["plan_digest"]
        .as_str()
        .unwrap()
        .to_string();
    w.comment_left_by_the_post(&c2);

    // A FAIL recorded after the plan changes a line, so the plan is no longer the plan.
    w.verdict(card, &c1, false, S_REV2);
    let why = refusal(&w.post(&base, &c2, &digest, &[]));
    assert!(why.contains("stale or altered plan"), "{why}");
    assert_eq!(w.posts(), 0);
    assert!(w.calls().is_empty(), "{:?}", w.calls());

    // A range that moved: the old digest posts nothing for the new head.
    let fresh = w.plan(&base, &c2)["plan_digest"]
        .as_str()
        .unwrap()
        .to_string();
    let c3 = w.commit("three", Some(S_AUTH));
    w.passed(&c3, S_REV);
    let why = refusal(&w.post(&base, &c3, &fresh, &[]));
    assert!(why.contains("stale or altered plan"), "{why}");
    // A narrower range (another base) is another plan too.
    let why = refusal(&w.post(&c1, &c2, &fresh, &[]));
    assert!(why.contains("stale or altered plan"), "{why}");
    assert!(w.calls().is_empty(), "{:?}", w.calls());
}

/// AC-3 (REQ-10): a summary over 60000 bytes is refused whole, in the preview and on the post;
/// nothing is cut and nothing is sent.
#[test]
fn a_summary_over_the_cap_is_refused_not_cut() {
    let w = World::new();
    let base = w.head();
    w.fast_history(700);
    let head = w.head();
    assert_eq!(w.first_parent(&base, &head).len(), 700);

    let why = refusal(&w.preview(&base, &head, &[]));
    assert!(why.contains("60000"), "{why}");
    assert!(
        why.contains("700"),
        "the refusal names the size of the range: {why}"
    );
    let why = refusal(&w.post(&base, &head, "sha256:0", &["--allow-incomplete"]));
    assert!(why.contains("60000"), "{why}");
    assert!(w.calls().is_empty(), "{:?}", w.calls());

    // A range that fits is whole: one line per commit, none cut.
    let near = w.first_parent(&base, &head)[199].clone();
    let plan = w.plan(&base, &near);
    assert_eq!(plan["landings"], json!(200));
    assert_eq!(read(&plan).lines.len(), 200);
    assert!(plan["bytes"].as_u64().unwrap() <= 60_000);
}

/// The arguments are read before anything else: a range that is not a first-parent line from
/// the base, a commit that is not there and a malformed value are refused by name.
#[test]
fn a_range_that_is_not_a_first_parent_line_from_the_base_is_refused() {
    let w = World::new();
    let base = w.head();
    let c1 = w.commit("one", Some(S_AUTH));
    git(&w.project, &["switch", "-q", "-c", "side", &base]);
    let side = w.commit("a commit beside the line", Some(S_AUTH));
    git(&w.project, &["switch", "-q", "main"]);

    // The head is not built on the base's first-parent line.
    let why = refusal(&w.preview(&c1, &side, &[]));
    assert!(!why.is_empty());
    // The base is the head: no landing.
    let why = refusal(&w.preview(&c1, &c1, &[]));
    assert!(why.contains("no first-parent commit"), "{why}");
    // Malformed and missing values, each named.
    for (b, h) in [
        ("1", c1.as_str()),
        (base.as_str(), "1"),
        (base.as_str(), &"0".repeat(40)),
    ] {
        let why = refusal(&w.preview(b, h, &[]));
        assert!(!why.is_empty(), "{b} {h}");
    }
    assert!(w.calls().is_empty());
    // And the control: the good range reads.
    assert_eq!(w.plan(&base, &c1)["landings"], json!(1));
}

/// AC-5 (REQ-11, REQ-12, REQ-13): the comment the command prints is the comment the audit gate
/// reads. The dry-run text of a complete range, fed as a member's comment of a release PR, passes
/// the gate; a landing the ledger has no verdict for prints a `NONE` line and the gate is
/// `missing`, naming it; a line cut out of the printed text is a comment that does not read, and
/// the gate fails it.
#[test]
fn the_comment_the_command_prints_is_what_audit_gate_reads() {
    let w = World::new();
    let base = w.head();
    let c1 = w.commit("card one", Some(S_AUTH));
    w.passed(&c1, S_REV);
    let c2 = w.commit("card two", Some(S_AUTH));
    w.passed(&c2, S_REV);

    // A release PR: more than 250 commits and no file list, so the gate knows neither the authors
    // nor the files; the printed lines have to carry the decision.
    let gate = |head: &str, merge_base: &str, text: &str| -> Value {
        let input: audit_gate::Input = serde_json::from_value(json!({
            "schema_version": 1, "pr": 9, "head": head, "body": "audit-tier: 2\n\nbody",
            "comments": [{"id": 31, "body": text, "author_association": "MEMBER"}],
            "release_base": merge_base,
        }))
        .unwrap();
        audit_gate::decide(&input)
    };

    let plan = w.plan(&base, &c2);
    assert_eq!(plan["complete"], json!(true), "{plan}");
    let text = plan["text"].as_str().unwrap().to_string();
    let decided = gate(&c2, &base, &text);
    assert_eq!(decided["state"], "pass", "{decided}");
    assert_eq!(decided["rule"], "release-summary", "{decided}");
    assert_eq!(decided["comment"], 31, "{decided}");

    // Cutting any one session or landing line out of the printed text is refused, never a pass.
    let lines: Vec<&str> = text.lines().collect();
    let cuttable: Vec<usize> = (0..lines.len())
        .filter(|&i| lines[i].starts_with("S ") || lines[i].starts_with("L "))
        .collect();
    assert!(cuttable.len() >= 4, "{text}");
    for cut in cuttable {
        let without: Vec<&str> = lines
            .iter()
            .enumerate()
            .filter(|(i, _)| *i != cut)
            .map(|(_, l)| *l)
            .collect();
        let decided = gate(&c2, &base, &without.join("\n"));
        assert_eq!(decided["state"], "fail", "cut `{}`: {decided}", lines[cut]);
        assert_eq!(decided["rule"], "summary-malformed", "{decided}");
    }

    // The merge base the gate read must be the base the command printed; a text of an older head
    // is no summary of this PR.
    let decided = gate(&c2, &c1, &text);
    assert_eq!(decided["rule"], "summary-base", "{decided}");
    assert_eq!(decided["state"], "missing", "{decided}");

    // A landing with no verdict on the ledger prints a NONE line and the gate names the commit.
    let c3 = w.commit("a direct commit", None);
    let plan = w.plan(&base, &c3);
    assert_eq!(plan["complete"], json!(false), "{plan}");
    let text3 = plan["text"].as_str().unwrap().to_string();
    let decided = gate(&c3, &base, &text3);
    assert_eq!(decided["state"], "missing", "{decided}");
    assert_eq!(decided["rule"], "summary-landings", "{decided}");
    assert!(
        decided["reason"].as_str().unwrap().contains(&c3[..12]),
        "{decided}"
    );
    let stale = gate(&c3, &base, &text);
    assert_eq!(
        stale["rule"], "pass-on-head",
        "an older head's text: {stale}"
    );
}
