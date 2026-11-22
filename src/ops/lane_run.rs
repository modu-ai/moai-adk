//! `lane-run` (SPEC-20261122-nonstop-lane): the pick of running a lane non-stop or one card at a
//! time, kept in the project profile's top-level `lane_run` key, and the texts that follow it.

use moai_cli::profile::{read_profile, read_rel, LaneRunSettings, LANE_RUNS};
use serde_json::{json, Value};
use std::io::Write as _;
use std::path::{Path, PathBuf};

/// Run a lane non-stop or one card at a time
///
/// `show` reads the pick in force for the project or a lane (the lane's own value before the
/// project's) and the host facts that turn it into `single`; `set` saves a pick without asking;
/// `prompt` prints the self-paced loop text for a non-stop lane; `goal` prints a `/goal`
/// condition for one card; `watch` is the command a Monitor runs: one line per new note of the
/// named cards, until its own deadline. Only `set` writes, and only the profile's `lane_run`
/// key. Nothing here starts a loop or a goal.
#[derive(clap::Args)]
pub struct Args {
    #[command(subcommand)]
    action: Action,
    /// The project folder.
    #[arg(long, global = true, default_value = ".")]
    project: PathBuf,
}

#[derive(clap::Subcommand)]
enum Action {
    /// The saved pick, the pick in force and the facts behind it.
    Show {
        /// A lane name, as `session-register` gave it.
        #[arg(long)]
        lane: Option<String>,
    },
    /// Save `nonstop` or `single` for the project, or for one lane with `--lane`.
    Set {
        value: String,
        #[arg(long)]
        lane: Option<String>,
    },
    /// The loop text for a lane whose pick in force is `nonstop`; none for `single`.
    Prompt {
        #[arg(long)]
        lane: String,
    },
    /// A `/goal` condition for one card, with a turn limit.
    Goal {
        #[arg(long)]
        card: String,
        #[arg(long, default_value_t = DEFAULT_TURNS)]
        turns: u32,
    },
    /// One line per new note of the named cards until the deadline; reads only. For Monitor.
    Watch {
        /// A card to watch, by its C-number; repeat for each card the lane holds.
        #[arg(long = "card", required = true)]
        cards: Vec<String>,
        /// Seconds until the watch ends with status 0 (60 to 1700), before Monitor's 30 minutes.
        #[arg(long, default_value_t = WATCH_DEADLINE_SECS)]
        deadline_secs: u64,
        /// Seconds between reads (1 to 60).
        #[arg(long, default_value_t = WATCH_EVERY_SECS)]
        every_secs: u64,
    },
}

pub const OPERATION: &str = "lane.run";

const DEFAULT_TURNS: u32 = 40;
const WATCH_DEADLINE_SECS: u64 = 1500;
const WATCH_DEADLINES: std::ops::RangeInclusive<u64> = 60..=1700;
const WATCH_EVERY_SECS: u64 = 30;
const WATCH_EVERIES: std::ops::RangeInclusive<u64> = 1..=60;
/// Events read per timeline call; more pages follow while `more` is true.
const WATCH_PAGE: i64 = 500;
const MAX_TURNS: u32 = 200;
/// The longest `/goal` condition Claude Code takes.
const GOAL_MAX_CHARS: usize = 4000;
const FLOORS: &str = "Claude Code 2.1.248 or later for the self-paced loop; 2.1.310 or later \
                      for Monitor on Bedrock, Claude Platform on AWS, Google Cloud's Agent \
                      Platform, Microsoft Foundry, or with telemetry or non-essential traffic \
                      off; Git Bash on Windows";

pub fn run(args: Args) -> Result<Value, String> {
    let root = args.project;
    match args.action {
        Action::Show { lane } => {
            let lane = checked_lane(lane.as_deref())?;
            let state = State::read(&root, lane)?;
            Ok(state.json(lane))
        }
        Action::Set { value, lane } => set(&root, &value, checked_lane(lane.as_deref())?),
        Action::Prompt { lane } => {
            let lane = checked_lane(Some(&lane))?.unwrap_or_default();
            let state = State::read(&root, Some(lane))?;
            if state.effective == "nonstop" {
                Ok(json!({
                    "lane": lane, "pick": "nonstop", "started": false,
                    "prompt": prompt(lane),
                    "note": "give this text to the self-paced loop of Claude Code (the loop \
                             command with no interval); it ends after seven days and is not \
                             restored by a resume",
                }))
            } else {
                Ok(json!({
                    "lane": lane, "pick": "single", "started": false, "prompt": null,
                    "facts": state.facts,
                    "note": "the pick in force is single: no loop is started",
                }))
            }
        }
        Action::Goal { card, turns } => goal(&card, turns),
        Action::Watch {
            cards,
            deadline_secs,
            every_secs,
        } => watch(&root, &cards, deadline_secs, every_secs),
    }
}

/// One watched card: where its reading stands.
enum Watched {
    /// Read up to this timeline cursor; only later notes are printed.
    At(String),
    /// Not readable now; when it reads again, it starts after the notes it has then.
    Unreadable,
}

/// The cursor after every event the card has now, or why it cannot be read.
fn watch_start(root: &Path, home: &Path, card: &str) -> Result<String, String> {
    let page = moai_cli::card::timeline::timeline(root, home, None, card, None, Some(1), false)?;
    // `end` names the newest event; a card always has its `card.add` event.
    page["end"]
        .as_str()
        .or_else(|| page["next"].as_str())
        .map(str::to_string)
        .ok_or_else(|| "the timeline gave no cursor".to_string())
}

/// The notes of `card` after `cursor`, as printed lines, and the cursor after them.
fn watch_read(
    root: &Path,
    home: &Path,
    card: &str,
    cursor: &str,
) -> Result<(Vec<String>, String), String> {
    let mut lines = Vec::new();
    let mut at = cursor.to_string();
    loop {
        let page = moai_cli::card::timeline::timeline(
            root,
            home,
            None,
            card,
            Some(&at),
            Some(WATCH_PAGE),
            false,
        )?;
        for e in page["events"].as_array().into_iter().flatten() {
            if e["kind"] != json!("card.note") {
                continue;
            }
            let line = e["note"]["line"].as_str().unwrap_or("");
            lines.push(format!(
                "{card} note {} {} {}",
                e["detail"]["note_id"],
                e["detail"]["kind"].as_str().unwrap_or("note"),
                line.lines().next().unwrap_or("")
            ));
        }
        at = page["next"].as_str().unwrap_or(&at).to_string();
        if page["more"] != json!(true) {
            return Ok((lines, at));
        }
    }
}

/// `watch`: the command a lane's Monitor runs (SPEC REQ-7, REQ-8, D-5). It reads the named
/// cards read-only through the card timeline cursor every `every` seconds and prints one line
/// per note that is new since it started, nothing for no change, and one `unreadable` line when
/// a card stops reading. It ends with status 0 at its own deadline, before Monitor's.
fn watch(root: &Path, cards: &[String], deadline: u64, every: u64) -> Result<Value, String> {
    if !WATCH_DEADLINES.contains(&deadline) {
        return Err(format!(
            "watch refused: a deadline of {deadline} seconds is outside 60 to 1700"
        ));
    }
    if !WATCH_EVERIES.contains(&every) {
        return Err(format!(
            "watch refused: a read every {every} seconds is outside 1 to 60"
        ));
    }
    let home = moai_cli::ledger::command::resolve_home(None)?;
    let end = std::time::Instant::now() + std::time::Duration::from_secs(deadline);
    let mut out = std::io::stdout().lock();
    let mut printed = 0usize;
    let mut say = |text: &str| {
        // A closed pipe means the Monitor is gone; the deadline still ends the run.
        let _ = writeln!(out, "{text}");
        let _ = out.flush();
    };
    let mut state: Vec<Watched> = Vec::new();
    for card in cards {
        state.push(match watch_start(root, &home, card) {
            Ok(cursor) => Watched::At(cursor),
            Err(e) => {
                say(&format!("{card} unreadable: {e}"));
                Watched::Unreadable
            }
        });
    }
    loop {
        let now = std::time::Instant::now();
        if now >= end {
            break;
        }
        std::thread::sleep((end - now).min(std::time::Duration::from_secs(every)));
        for (card, watched) in cards.iter().zip(state.iter_mut()) {
            let next = match watched {
                Watched::At(cursor) => match watch_read(root, &home, card, cursor) {
                    Ok((lines, at)) => {
                        for l in &lines {
                            say(l);
                        }
                        printed += lines.len();
                        Watched::At(at)
                    }
                    Err(e) => {
                        say(&format!("{card} unreadable: {e}"));
                        Watched::Unreadable
                    }
                },
                Watched::Unreadable => match watch_start(root, &home, card) {
                    Ok(cursor) => Watched::At(cursor),
                    Err(_) => Watched::Unreadable,
                },
            };
            *watched = next;
        }
    }
    let unreadable: Vec<&String> = cards
        .iter()
        .zip(&state)
        .filter(|(_, w)| matches!(w, Watched::Unreadable))
        .map(|(c, _)| c)
        .collect();
    Ok(json!({
        "action": "watch", "cards": cards, "deadline_secs": deadline, "every_secs": every,
        "printed": printed, "unreadable": unreadable, "ended": "deadline", "writes": false,
    }))
}

fn checked_lane(lane: Option<&str>) -> Result<Option<&str>, String> {
    match lane {
        Some(l) if !moai_cli::lane_registry::valid_name(l) => Err(
            "invalid lane name: 1 to 64 letters, digits, `.`, `_` or `-`, starting with a \
             letter or digit"
                .into(),
        ),
        other => Ok(other),
    }
}

/// The saved pick, where it came from, the facts and the pick in force.
struct State {
    saved: &'static str,
    source: &'static str,
    effective: &'static str,
    facts: Value,
}

fn pick_name(v: &str) -> &'static str {
    LANE_RUNS
        .iter()
        .copied()
        .find(|p| *p == v)
        .unwrap_or("unset")
}

/// A true value of an environment switch, the way Claude Code's switches are written.
fn truthy(v: &str) -> bool {
    matches!(
        v.trim().to_ascii_lowercase().as_str(),
        "1" | "true" | "yes" | "on"
    )
}

/// The one rule for the key's shape, on `show` and on `set`: `lane_run` is an object or absent.
/// The profile reader takes `null` as unset and `[]` as an empty setting; `lane-run` refuses
/// both, as it refuses any other non-object value, before it reads or writes the key.
fn lane_run_shape(bytes: &[u8]) -> Result<(), String> {
    let map: serde_json::Map<String, Value> =
        serde_json::from_slice(bytes).map_err(|_| "profile is not a JSON object".to_string())?;
    match map.get("lane_run") {
        Some(v) if !v.is_object() => Err("invalid lane_run setting; no values were logged".into()),
        _ => Ok(()),
    }
}

impl State {
    fn read(root: &Path, lane: Option<&str>) -> Result<State, String> {
        let (profile, bytes) = read_profile(root)?;
        lane_run_shape(&bytes)?;
        let settings = profile.lane_run.unwrap_or_default();
        let (saved, source) = match lane.and_then(|l| settings.lanes.get(l)) {
            Some(v) => (pick_name(v), "lane"),
            None => match settings.default.as_deref() {
                Some(v) => (pick_name(v), "project"),
                None => ("unset", "none"),
            },
        };
        let cron_off = std::env::var("CLAUDE_CODE_DISABLE_CRON").is_ok_and(|v| truthy(&v));
        let facts = json!([
            {"name": "CLAUDE_CODE_DISABLE_CRON", "read": true, "blocks": cron_off,
             "says": "set to a true value, it removes the loop and the scheduled wakes"},
            {"name": "loop_available", "read": false, "blocks": false,
             "says": "whether Claude Code can be asked for a loop here; moai cannot read it"},
            {"name": "claude_code_version", "read": false, "blocks": false,
             "says": "moai reads no host version; see floors"},
        ]);
        let effective = if saved == "nonstop" && !cron_off {
            "nonstop"
        } else {
            "single"
        };
        Ok(State {
            saved,
            source,
            effective,
            facts,
        })
    }

    fn json(&self, lane: Option<&str>) -> Value {
        json!({
            "lane": lane, "saved": self.saved, "source": self.source,
            "effective": self.effective, "facts": self.facts, "floors": FLOORS,
        })
    }
}

fn set(root: &Path, value: &str, lane: Option<&str>) -> Result<Value, String> {
    if !LANE_RUNS.contains(&value) {
        return Err("invalid pick: nonstop or single".into());
    }
    // The profile must read before it is changed; its raw bytes keep every other key in order.
    let (_, bytes) = read_profile(root)?;
    lane_run_shape(&bytes)?;
    let path = root.join(read_rel(root));
    let meta = std::fs::symlink_metadata(&path).map_err(|e| format!("profile: {e}"))?;
    if !meta.file_type().is_file() {
        return Err("profile is not a regular file; nothing was written".into());
    }
    let mut map: serde_json::Map<String, Value> =
        serde_json::from_slice(&bytes).map_err(|_| "profile is not a JSON object".to_string())?;
    let mut settings: LaneRunSettings = match map.get("lane_run") {
        Some(v) => serde_json::from_value(v.clone()).map_err(|_| "invalid lane_run setting")?,
        None => LaneRunSettings::default(),
    };
    match lane {
        Some(l) => {
            settings.lanes.insert(l.to_string(), value.to_string());
        }
        None => settings.default = Some(value.to_string()),
    }
    map.insert(
        "lane_run".into(),
        serde_json::to_value(&settings).map_err(|e| e.to_string())?,
    );
    let mut out = serde_json::to_vec_pretty(&map).map_err(|e| e.to_string())?;
    out.push(b'\n');
    let dir = path.parent().ok_or("profile has no folder")?;
    let mut tmp = tempfile::NamedTempFile::new_in(dir).map_err(|e| format!("profile: {e}"))?;
    // The temporary file is made 0600; the profile keeps the mode it had, as `apply` keeps it.
    #[cfg(unix)]
    tmp.as_file()
        .set_permissions(meta.permissions())
        .map_err(|e| format!("profile: {e}"))?;
    tmp.write_all(&out)
        .and_then(|_| tmp.as_file().sync_all())
        .map_err(|e| format!("profile: {e}"))?;
    tmp.persist(&path)
        .map_err(|e| format!("profile: {}", e.error))?;
    // Read back: the written profile must still read, with the value in it.
    let check = State::read(root, lane)?;
    Ok(json!({
        "written": true, "key": "lane_run", "lane": lane, "value": value,
        "saved": check.saved, "source": check.source, "profile": read_rel(root),
    }))
}

/// The plain text a non-stop lane gives to the self-paced loop: no card, no leader and no
/// path; the next card is whatever `moai card next` names.
fn prompt(lane: &str) -> String {
    format!(
        "Lane {lane}: keep running cards without stopping, as a self-paced loop with no \
         interval.\n\
         1. Keep the current card: if a card is in progress, continue it first.\n\
         2. After you post a READY note, run `moai card next` and take the card it names; read \
         its body and notes whole, then write the SPEC and the tests first.\n\
         3. While you wait for an audit verdict or an answer, keep a Monitor that runs `moai \
         lane-run watch --card <card>` with one --card for each card you hold; it prints one \
         line per new note and ends by itself before the Monitor does. Work the next card \
         meanwhile, and start the watch again when it ends.\n\
         4. A verdict or a repair request comes first: handle it before anything else.\n\
         5. A decision you need goes into a LANE-BLOCKED kind=decision note on its card; do not \
         ask the user from the loop.\n\
         6. never end a turn with neither a card in progress nor a watch: before a turn ends, \
         schedule the next wake with ScheduleWakeup (1200 to 1800 seconds while a Monitor \
         runs).\n\
         7. When `moai card next` names no card and no pull request of this lane is open, print \
         one idle line (lane {lane} idle: no card to take), stop the watch and end the loop \
         with ScheduleWakeup and stop: true.\n\
         Read and write the ledger only through moai commands. The loop ends after seven days \
         and is not restored by a resume."
    )
}

/// A card name the goal text takes: `C-` and 1 to 6 digits, no leading zero.
fn card_name(s: &str) -> bool {
    s.strip_prefix("C-").is_some_and(|n| {
        (1..=6).contains(&n.len()) && !n.starts_with('0') && n.bytes().all(|b| b.is_ascii_digit())
    })
}

fn goal(card: &str, turns: u32) -> Result<Value, String> {
    if !card_name(card) {
        return Err("invalid card name: C- and 1 to 6 digits".into());
    }
    if !(1..=MAX_TURNS).contains(&turns) {
        return Err(format!("invalid turn limit: 1 to {MAX_TURNS}"));
    }
    let condition = format!(
        "Card {card} is in review: its READY note exists and names the head and base read from \
         git (the head commit of the card's branch and its merge base with the integration \
         branch), the card's own tests pass on that head, and `moai card show {card}` reads its \
         state as review. Or stop after {turns} turns and say what is left."
    );
    debug_assert!(condition.chars().count() <= GOAL_MAX_CHARS);
    Ok(json!({
        "card": card, "turns": turns, "condition": condition, "started": false,
        "max_chars": GOAL_MAX_CHARS,
        "note": "the user types /goal with this condition; a goal adds a second model's check \
                 after every turn and its tokens",
    }))
}
