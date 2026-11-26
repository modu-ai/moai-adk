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
/// condition for one card. Only `set` writes, and only the profile's `lane_run` key. Nothing
/// here starts a loop or a goal.
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
    ///
    /// Example: moai lane-run goal --card <card>
    Goal {
        #[arg(long)]
        card: String,
        #[arg(long, default_value_t = DEFAULT_TURNS)]
        turns: u32,
    },
}

pub const OPERATION: &str = "lane.run";

const DEFAULT_TURNS: u32 = 40;
const MAX_TURNS: u32 = 200;
/// The longest `/goal` condition Claude Code takes.
const GOAL_MAX_CHARS: usize = 4000;
/// The host versions the loop and the watch need, each with the condition it holds under. moai
/// reads no host version: this is text for the user.
const FLOORS: &str = "Claude Code 2.1.248 or later for the self-paced loop on Amazon Bedrock, \
                      Claude Platform on AWS, Google Cloud's Agent Platform or Microsoft \
                      Foundry, or with feature-flag fetching off; 2.1.310 or later for Monitor \
                      on those providers, or with telemetry or non-essential traffic off; Git \
                      Bash on Windows; before 2.1.330 a loop ends after 7 days instead of 30";

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
                             command with no interval); it ends 30 days after it starts (7 days \
                             before Claude Code 2.1.330) and is not restored by a resume",
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
    }
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

/// The plain text a non-stop lane gives to the self-paced loop. It names no card and no path.
/// At every wake the lane reads the leader entry: with a registered leader it waits for a
/// dispatch, and only a session that has never seen one takes what `moai card next` names. Its
/// one watch is `moai card wait` under a Monitor.
fn prompt(lane: &str) -> String {
    format!(
        "Lane {lane}: keep working without stopping, as a self-paced loop with no interval.\n\
         At every wake, first run `moai session-register --show --name leader`. held means a \
         leader is registered. free means there is none, unless this session has already seen a \
         registered leader or received a dispatch: then the entry has lapsed. On a lapsed \
         entry, any other answer or a failure, take no new card, continue only the cards you \
         hold, and write one note that names the answer on the card you worked last.\n\
         1. Keep the current card: if a card is in progress, continue it first.\n\
         2. With a leader registered: after your READY note, hand the card back, send the one \
         completion message and wait for the leader's dispatch; never take the next card on \
         your own. With no leader: when the ledger shows the previous card done or released, \
         run `moai card next`, take the card it names and read its body and notes whole.\n\
         3. Watch your cards: those `moai card next` lists under mine, and each card of `moai \
         card list --state review` whose `moai card show` roles.authors holds the \
         source_session_id that `moai session-current` prints. Keep a Monitor, with its timeout \
         at 30 minutes, that runs `moai card wait --card <card> --timeout 25m --not-author \
         <your session id>`, one --card for each watched card. With no card to watch, start no \
         watch. On a first start, start it and then read each watched card with `moai card \
         show`. When it ends, start it again with \
         --since set to the highest note_id it returned, or to the mark it answered on a \
         timeout; if it ended with no answer, do a first start again.\n\
         4. A verdict or a repair request comes first.\n\
         5. A decision you need goes into a LANE-BLOCKED kind=decision note on its card; do not \
         ask the user from the loop.\n\
         6. never end a turn with neither a card in progress, a watch nor a scheduled wake: \
         call ScheduleWakeup (1200 to 1800 seconds while a Monitor runs, 3600 with nothing to \
         watch).\n\
         7. When there is no card you can work (a card that waits for a decision does not \
         count), print one idle line (lane {lane} idle: no card to take) once. If this session \
         has seen no leader, stop the watch and end the loop with ScheduleWakeup and stop: \
         true. Otherwise keep one wake pending at 3600 seconds, and end the loop only on the \
         leader's session-end notice or the user's word.\n\
         Read and write the ledger only through moai commands. The loop ends 30 days after it \
         starts (7 days before Claude Code 2.1.330) and is not restored by a resume."
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
