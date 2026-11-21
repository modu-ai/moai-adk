//! card landed + card done under m-flow (SPEC-20261120-card-done-mflow): a review card whose
//! landed head is contained in the profile's integration branch closes without a pull
//! request; a project without an m-flow work_flow keeps the merged-pull-request rule. Every
//! case runs in a temporary HOME and a temporary git repository; the real `~/.moai` is not
//! touched.

mod common;

use common::home::TempHome;
use moai_cli::ledger::{self, WriteLedger};
use serde_json::{json, Value};
use std::path::Path;
use std::process::Output;

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
const S_A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";

fn git(dir: &Path, args: &[&str]) -> String {
    let mut all = vec![
        "-c",
        "user.name=fixture",
        "-c",
        "user.email=fixture@example.invalid",
    ];
    all.extend_from_slice(args);
    common::git::git(dir, &all)
}

fn run(home: &TempHome, args: &[&str]) -> Output {
    home.command(args)
        .env(SESSION_VAR, S_A)
        .output()
        .expect("run moai")
}

/// Asserts exit 0 and a success envelope, and unwraps the result.
fn result(out: &Output) -> Value {
    assert_eq!(
        out.status.code(),
        Some(0),
        "{} {}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

/// Asserts exit 1 and returns the whole refusal envelope.
fn refusal(out: &Output) -> Value {
    assert_eq!(out.status.code(), Some(1));
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], json!(false), "{v}");
    v
}

/// The world: a temporary git repository as the project, the ledger registered for it, one
/// review card held by this session with a met criterion, and — for the m-flow cases — the
/// profile naming the m-flow work_flow with `integ` as the integration branch. `integ`
/// starts at the card's head commit; `main` moves ahead of it. Returns the home, the project
/// path, the card id and the head sha.
fn world(mflow: bool) -> (TempHome, String, i64, String) {
    let home = TempHome::new();
    let project = home.project("demo");
    let p = project.to_str().unwrap().to_string();
    ledger::register(home.path(), "demo").unwrap();
    git(&project, &["init", "-q", "-b", "main"]);
    std::fs::write(project.join("a.txt"), "a\n").unwrap();
    git(&project, &["add", "a.txt"]);
    git(&project, &["commit", "-q", "-m", "init"]);
    let head = git(&project, &["rev-parse", "HEAD"]).trim().to_string();
    git(&project, &["branch", "integ"]);
    // The card: added and made ready through the CLI, then reserved for this session, one
    // met criterion, review — the ledger holds the state the CLI commands read.
    let added = result(&run(
        &home,
        &[
            "card",
            "add",
            "--title",
            "land me",
            "--scope",
            "apps/demo/",
            "--project",
            &p,
        ],
    ));
    let name = added["name"].as_str().unwrap().to_string();
    result(&run(&home, &["card", "ready", &name, "--project", &p]));
    let mut w: WriteLedger = ledger::open_write(home.path(), "demo").unwrap();
    let r = moai_cli::ledger::reservation::reserve(
        &mut w,
        1,
        S_A,
        &["apps/demo/".to_string()],
        3600,
        None,
        None,
    )
    .unwrap();
    let generation = r.generation;
    w.with_write_tx(|tx, now| {
        Ok(tx.execute(
            "INSERT INTO card_ac (card_id, seq, text, status, evidence_ref, updated_at) \
             VALUES (1, 1, 'works', 'met', 'fixture', ?1)",
            rusqlite::params![now],
        )?)
    })
    .unwrap();
    moai_cli::ledger::workflow::review(&mut w, 1, S_A, generation).unwrap();
    drop(w);
    // The profile goes last: some card commands rewrite it (the id stub), and done must read
    // the m-flow work_flow from the final bytes.
    if mflow {
        let cfg = project.join(".moai-cli/project.yaml");
        std::fs::create_dir_all(cfg.parent().unwrap()).unwrap();
        std::fs::write(
            &cfg,
            &b"{\"schema_version\": 1, \"project_id\": \"demo\", \"purpose\": \"fixture\", \"tools\": [\"rust\"], \"development_methods\": [\"sdd\"], \"hosts\": [\"claude\"], \"work_flow\": {\"flow\": \"m-flow\", \"integration_branch\": \"integ\"}}"[..],
        )
        .unwrap();
    }
    (home, p, 1, head)
}

/// Reads the newest event payload of one kind through the read-only ledger.
fn last_event(home: &TempHome, kind: &str) -> Value {
    let ro = ledger::open_read_only(home.path(), "demo").unwrap();
    let conn = ro.conn();
    let (payload, actor): (String, String) = conn
        .query_row(
            "SELECT payload, actor FROM event WHERE kind = ?1 ORDER BY event_id DESC LIMIT 1",
            [kind],
            |r| Ok((r.get(0)?, r.get(1)?)),
        )
        .unwrap();
    let mut v: Value = serde_json::from_str(&payload).unwrap();
    v["actor"] = json!(actor);
    v
}

/// AC-1: `card landed` records the event with the head, the by and the writing session.
#[test]
fn landed_records_the_head_and_refuses_a_bad_head() {
    let (home, p, _card, head) = world(true);
    let v = result(&run(
        &home,
        &[
            "card",
            "landed",
            "C-1",
            "--head",
            &head,
            "--by",
            "leader",
            "--project",
            &p,
        ],
    ));
    assert_eq!(v["action"], json!("landed"), "{v}");
    let event = last_event(&home, "card.landed");
    assert_eq!(event["head"], json!(head), "{event}");
    assert_eq!(event["by"], json!("leader"), "{event}");
    assert_eq!(event["actor"], json!(S_A), "{event}");
    let bad = run(
        &home,
        &[
            "card",
            "landed",
            "C-1",
            "--head",
            &head[..39],
            "--by",
            "leader",
            "--project",
            &p,
        ],
    );
    let bad: Value = serde_json::from_slice(&bad.stdout).unwrap();
    assert_eq!(bad["error"]["kind"], json!("invalid_input"), "{bad}");
}

/// AC-2: the landed head sits on the integration branch, so done closes without a PR and the
/// card.done event names the head and the tip.
#[test]
fn m_flow_done_closes_on_a_landed_head_contained_in_the_integration_branch() {
    let (home, p, _card, head) = world(true);
    result(&run(
        &home,
        &[
            "card",
            "landed",
            "C-1",
            "--head",
            &head,
            "--by",
            "leader",
            "--project",
            &p,
        ],
    ));
    let done = result(&run(
        &home,
        &["card", "done", "C-1", "--generation", "1", "--project", &p],
    ));
    assert_eq!(done["name"], json!("C-1"), "{done}");
    let event = last_event(&home, "card.done");
    assert_eq!(event["mflow"]["head"], json!(head), "{event}");
    let tip = git(home.project("demo").as_path(), &["rev-parse", "integ"])
        .trim()
        .to_string();
    assert_eq!(event["mflow"]["tip"], json!(tip), "{event}");
}

/// AC-2: a landed head that is not an ancestor of the integration branch is refused at done
/// with the rule naming why.
#[test]
fn m_flow_done_refuses_a_head_outside_the_integration_branch() {
    let (home, p, _card, _head) = world(true);
    let repo = Path::new(&p).to_path_buf();
    git(Path::new(&p), &["checkout", "-q", "-b", "scratch"]);
    std::fs::write(repo.join("b.txt"), "b\n").unwrap();
    git(Path::new(&p), &["add", "b.txt"]);
    git(Path::new(&p), &["commit", "-q", "-m", "orphan"]);
    let orphan = git(Path::new(&p), &["rev-parse", "scratch"])
        .trim()
        .to_string();
    result(&run(
        &home,
        &[
            "card",
            "landed",
            "C-1",
            "--head",
            &orphan,
            "--by",
            "leader",
            "--project",
            &p,
        ],
    ));
    let dout = run(
        &home,
        &["card", "done", "C-1", "--generation", "1", "--project", &p],
    );
    let v = refusal(&dout);
    let message = v["error"]["message"].as_str().unwrap();
    assert!(
        message.contains("not contained in the integration branch"),
        "{v}"
    );
}

/// AC-3: a project without an m-flow work_flow keeps the merged-PR rule even with a landing.
#[test]
fn a_project_without_m_flow_keeps_the_pr_rule() {
    let (home, p, _card, head) = world(false);
    result(&run(
        &home,
        &[
            "card",
            "landed",
            "C-1",
            "--head",
            &head,
            "--by",
            "leader",
            "--project",
            &p,
        ],
    ));
    let v = refusal(&run(
        &home,
        &["card", "done", "C-1", "--generation", "1", "--project", &p],
    ));
    let message = v["error"]["message"].as_str().unwrap();
    assert!(message.contains("no merged pull request is linked"), "{v}");
}

// ---- SPEC-20261121-mixed-profile-read: which profile `card done` reads, and what it says when it
// cannot use one. -------------------------------------------------------------------------------

const PROFILE_PLAIN: &str = "{\"schema_version\": 1, \"project_id\": \"demo\", \"purpose\": \"fixture\", \"tools\": [\"rust\"], \"development_methods\": [\"sdd\"], \"hosts\": [\"claude\"]}";
const PROFILE_MFLOW: &str = "{\"schema_version\": 1, \"project_id\": \"demo\", \"purpose\": \"fixture\", \"tools\": [\"rust\"], \"development_methods\": [\"sdd\"], \"hosts\": [\"claude\"], \"work_flow\": {\"flow\": \"m-flow\", \"integration_branch\": \"integ\"}}";
const PROFILE_GITHUB: &str = "{\"schema_version\": 1, \"project_id\": \"demo\", \"purpose\": \"fixture\", \"tools\": [\"rust\"], \"development_methods\": [\"sdd\"], \"hosts\": [\"claude\"], \"work_flow\": {\"flow\": \"github-flow\", \"integration_branch\": \"integ\"}}";
/// A profile whose project id is fine (so the card commands find the ledger) but whose
/// `work_flow` names a flow that does not exist: `read_profile` refuses it.
const PROFILE_BROKEN: &str = "{\"schema_version\": 1, \"project_id\": \"demo\", \"purpose\": \"fixture\", \"tools\": [\"rust\"], \"development_methods\": [\"sdd\"], \"hosts\": [\"claude\"], \"work_flow\": {\"flow\": \"bogus-flow\"}}";

/// A world without a profile of its own (the card is in review) plus the profile files named:
/// `(project-relative path, bytes)`. Every earlier profile file is removed first, so the
/// layout is exactly what the case asks for.
fn world_with_profiles(files: &[(&str, &str)]) -> (TempHome, String, String) {
    let (home, p, _card, head) = world(false);
    let project = Path::new(&p).to_path_buf();
    let _ = std::fs::remove_file(project.join(".moai-cli/project.yaml"));
    for (rel, bytes) in files {
        let path = project.join(rel);
        std::fs::create_dir_all(path.parent().unwrap()).unwrap();
        std::fs::write(path, bytes).unwrap();
    }
    (home, p, head)
}

fn land(home: &TempHome, p: &str, head: &str) {
    result(&run(
        home,
        &[
            "card",
            "landed",
            "C-1",
            "--head",
            head,
            "--by",
            "leader",
            "--project",
            p,
        ],
    ));
}

fn done(home: &TempHome, p: &str) -> Output {
    run(
        home,
        &["card", "done", "C-1", "--generation", "1", "--project", p],
    )
}

/// `run` with stdin: the plan/apply dance feeds the desired profile and the plan.
fn run_stdin(home: &TempHome, args: &[&str], stdin: &[u8]) -> Output {
    let mut child = home
        .command(args)
        .stdin(std::process::Stdio::piped())
        .stdout(std::process::Stdio::piped())
        .stderr(std::process::Stdio::piped())
        .spawn()
        .expect("spawn moai");
    use std::io::Write as _;
    let mut pipe = child.stdin.take().unwrap();
    let _ = pipe.write_all(stdin);
    drop(pipe);
    child.wait_with_output().unwrap()
}

fn message_of(v: &Value) -> String {
    v["error"]["message"]
        .as_str()
        .unwrap_or_default()
        .to_string()
}

fn done_events(home: &TempHome) -> i64 {
    let ro = ledger::open_read_only(home.path(), "demo").unwrap();
    ro.conn()
        .query_row(
            "SELECT COUNT(*) FROM event WHERE kind = 'card.done'",
            [],
            |r| r.get(0),
        )
        .unwrap()
}

/// REQ-1 + REQ-7: the earlier profile has no work_flow, the newer one names m-flow; done reads the
/// newer one and closes on the landing.
#[test]
fn a_mixed_layout_reads_the_newer_profile_and_closes_on_the_landing() {
    let (home, p, head) = world_with_profiles(&[
        (".moai-cli/project.yaml", PROFILE_PLAIN),
        (".moai/project.yaml", PROFILE_MFLOW),
    ]);
    land(&home, &p, &head);
    let v = result(&done(&home, &p));
    assert_eq!(v["name"], json!("C-1"), "{v}");
    let event = last_event(&home, "card.done");
    assert_eq!(event["mflow"]["head"], json!(head), "{event}");
}

/// REQ-3: a newer profile that cannot be read is named; the earlier m-flow profile is not a
/// fallback (falling back would close the card).
#[test]
fn an_unreadable_newer_profile_is_named_and_the_earlier_one_is_no_fallback() {
    let (home, p, head) = world_with_profiles(&[
        (".moai-cli/project.yaml", PROFILE_MFLOW),
        (".moai/project.yaml", PROFILE_BROKEN),
    ]);
    land(&home, &p, &head);
    let v = refusal(&done(&home, &p));
    let message = message_of(&v);
    assert!(
        message.contains("no merged pull request is linked"),
        "{message}"
    );
    assert!(message.contains("a local landing is recorded"), "{message}");
    assert!(message.contains(".moai/project.yaml"), "{message}");
    assert!(message.contains("could not be read"), "{message}");
    assert_eq!(done_events(&home), 0);
}

/// REQ-4: the only profile cannot be read; the merged-PR refusal keeps its words and names the
/// profile and the reason.
#[test]
fn an_unreadable_profile_is_named_in_the_refusal() {
    let (home, p, head) = world_with_profiles(&[(".moai-cli/project.yaml", PROFILE_BROKEN)]);
    land(&home, &p, &head);
    let v = refusal(&done(&home, &p));
    let message = message_of(&v);
    assert!(
        message.contains("no merged pull request is linked"),
        "{message}"
    );
    assert!(message.contains("a local landing is recorded"), "{message}");
    assert!(message.contains(".moai-cli/project.yaml"), "{message}");
    assert!(message.contains("could not be read"), "{message}");
    assert_eq!(done_events(&home), 0);
}

/// REQ-6: a readable profile that is not m-flow, with a landing on record: the refusal keeps the
/// merged-PR wording and adds which profile was read and why it was not used.
#[test]
fn a_profile_that_is_not_m_flow_is_named_when_a_landing_is_on_record() {
    for (bytes, why) in [
        (PROFILE_PLAIN, "no work_flow"),
        (PROFILE_GITHUB, "github-flow"),
    ] {
        let (home, p, head) = world_with_profiles(&[(".moai-cli/project.yaml", bytes)]);
        land(&home, &p, &head);
        let v = refusal(&done(&home, &p));
        let message = message_of(&v);
        assert!(
            message.contains("no merged pull request is linked"),
            "{message}"
        );
        assert!(message.contains("a local landing is recorded"), "{message}");
        assert!(message.contains(".moai-cli/project.yaml"), "{message}");
        assert!(message.contains(why), "{why}: {message}");
        assert_eq!(done_events(&home), 0);
    }
}

/// REQ-5 + REQ-6 (no landing recorded): the plain merged-PR refusal stays as it was, with no hint.
#[test]
fn without_a_landing_the_refusal_carries_no_profile_hint() {
    // `TempHome::project` would rewrite the profile, so the path comes from the world itself.
    let (home, p, _head) = world_with_profiles(&[(".moai-cli/project.yaml", PROFILE_PLAIN)]);
    let v = refusal(&done(&home, &p));
    let message = message_of(&v);
    assert!(
        message.contains("no merged pull request is linked"),
        "{message}"
    );
    assert!(
        !message.contains("a local landing is recorded"),
        "{message}"
    );
}

// ---- The delta audit's B2, B3, B4 and N3 repairs. ---------------------------------------------

/// An m-flow profile that never names an integration branch (the B3 hint's case).
const PROFILE_MFLOW_NO_BRANCH: &str = "{\"schema_version\": 1, \"project_id\": \"demo\", \"purpose\": \"fixture\", \"tools\": [\"rust\"], \"development_methods\": [\"sdd\"], \"hosts\": [\"claude\"], \"work_flow\": {\"flow\": \"m-flow\"}}";
/// A plain profile of a DIFFERENT project (the N3 mixed-layout disagreement).
const PROFILE_OTHER_ID: &str = "{\"schema_version\": 1, \"project_id\": \"other\", \"purpose\": \"fixture\", \"tools\": [\"rust\"], \"development_methods\": [\"sdd\"], \"hosts\": [\"claude\"]}";
/// Not JSON at all (the B2 path-naming case).
const PROFILE_NOT_JSON: &str = "{not json";

/// B3 regression: an m-flow work_flow with no integration_branch gets the landing hint on the
/// landing-related refusal, naming the profile it read.
#[test]
fn an_m_flow_without_an_integration_branch_names_the_profile_in_the_hint() {
    let (home, p, head) = world_with_profiles(&[
        (".moai-cli/project.yaml", PROFILE_PLAIN),
        (".moai/project.yaml", PROFILE_MFLOW_NO_BRANCH),
    ]);
    land(&home, &p, &head);
    let v = refusal(&done(&home, &p));
    let message = message_of(&v);
    assert!(
        message.contains("no merged pull request is linked"),
        "{message}"
    );
    assert!(
        message.contains(".moai/project.yaml"),
        "the hint names the profile: {message}"
    );
    assert!(
        message.contains("does not name the m-flow work_flow with an integration branch"),
        "{message}"
    );
    assert_eq!(done_events(&home), 0);
}

/// B2 regression: a broken newer profile in a mixed layout is refused naming the file and the
/// reason, not a bare preservation-review sentence.
#[test]
fn a_broken_newer_profile_is_refused_with_its_path() {
    let (home, p, _head) = world_with_profiles(&[
        (".moai-cli/project.yaml", PROFILE_PLAIN),
        (".moai/project.yaml", PROFILE_NOT_JSON),
    ]);
    let v = refusal(&done(&home, &p));
    let message = message_of(&v);
    assert!(message.contains(".moai/project.yaml"), "{message}");
    assert!(
        message.contains("existing YAML requires preservation review"),
        "{message}"
    );
}

/// N3 regression: a mixed layout whose two profiles name different projects is refused, naming
/// both files and both ids, by a card command.
#[test]
fn two_profiles_with_different_project_ids_are_refused() {
    let (home, p, _head) = world_with_profiles(&[
        (".moai-cli/project.yaml", PROFILE_OTHER_ID),
        (".moai/project.yaml", PROFILE_MFLOW),
    ]);
    let v = refusal(&done(&home, &p));
    let message = message_of(&v);
    assert!(message.contains(".moai/project.yaml"), "{message}");
    assert!(message.contains(".moai-cli/project.yaml"), "{message}");
    assert!(message.contains("`demo`"), "{message}");
    assert!(message.contains("`other`"), "{message}");
    assert_eq!(done_events(&home), 0);
}

/// B4 regression: in a mixed layout a profile write follows the read — `project-apply` updates
/// the newer `.moai/project.yaml` and leaves the older one untouched, and `language-setup`
/// reads the written value back. (`work_flow` is a user-edited key: REQ-03A-03 keeps it out of
/// a plan's delta, so its mixed-layout read is pinned by the done tests above.)
#[test]
fn a_mixed_layout_write_updates_the_newer_profile_and_is_read_back() {
    let (home, p, _head) = world_with_profiles(&[
        (".moai-cli/project.yaml", PROFILE_PLAIN),
        (".moai/project.yaml", PROFILE_MFLOW),
    ]);
    let project = Path::new(&p).to_path_buf();
    // The plan/apply dance writes the whole profile: the desired one (schema v2, what
    // project-plan accepts) adds a managed `language` block and keeps the rest.
    let desired = "{\"schema_version\": 2, \"project_id\": \"demo\", \"purpose\": \"fixture\", \"tools\": [\"rust\"], \"development_methods\": [\"sdd\"], \"hosts\": [\"claude\"], \"work_flow\": {\"flow\": \"m-flow\", \"integration_branch\": \"integ\"}, \"language\": {\"user_chat\": \"en\", \"commit_messages\": \"en\", \"documents\": \"ja\", \"instruction_files\": \"en\"}}";
    let plan = result(&run_stdin(&home, &["project-plan", &p], desired.as_bytes()));
    let _ = result(&run_stdin(
        &home,
        &[
            "project-apply",
            &p,
            "--plan-digest",
            plan["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "fixture-approved",
        ],
        plan.to_string().as_bytes(),
    ));
    let newer: Value =
        serde_json::from_str(&std::fs::read_to_string(project.join(".moai/project.yaml")).unwrap())
            .unwrap();
    assert_eq!(
        newer["language"]["documents"],
        json!("ja"),
        "the write landed in the newer profile: {newer}"
    );
    let older: Value = serde_json::from_str(
        &std::fs::read_to_string(project.join(".moai-cli/project.yaml")).unwrap(),
    )
    .unwrap();
    assert!(
        older.get("language").is_none(),
        "the older profile is untouched: {older}"
    );
    // The write is read back: language-setup reads the chosen (newer) profile.
    let q = result(&run(&home, &["language-setup", &p, "--questions"]));
    let documents = q["questions"]["items"]
        .as_array()
        .unwrap()
        .iter()
        .find(|i| i["id"] == "documents")
        .unwrap()
        .clone();
    assert_eq!(documents["current"], json!("ja"), "{q}");
}
