//! The target `lane-next` offers a lane, kept for a short lease (#3015 C2, N4).
//!
//! Two lanes calling `lane-next` a moment apart used to be offered the same audit or Issue,
//! because the first lane's `moai-audit-claim` or `moai-claim` comment is not on GitHub yet.
//! Here the offer to a registered lane is written to `<session registry>/proposals/<lane>.json`
//! (`{schema_version, lane, kind, target, proposed_at, lease_until}`), and another lane's
//! decision skips a target that is offered to a *live* lane whose lease still runs.
//!
//! - Live means the offered lane has a session registry entry with `lease_until > now`; an offer
//!   to a lane with no entry or an expired one blocks nobody, and a lane that was never
//!   registered is told `caller_not_registered` (its offer is kept but nobody honours it).
//! - The lease is [`PROPOSAL_LEASE_SECS`]. The same lane asking again for the same target keeps
//!   the first `lease_until`, so a stuck lane cannot hold a target forever.
//! - Reading, deciding and writing happen under the registry's exclusive lock, so two lanes
//!   that ask at once get different targets; the GitHub reads stay outside the lock.
//! - Any trouble with the folder (missing, a link, oversized or malformed file, no lock) leaves
//!   the decision as it was without offers, and the answer says `proposal.status: unavailable`
//!   with a fixed reason (no path, no file text). A lane is never blocked by this.
//! - GitHub stays the real reservation; this only closes the gap before it shows, on this
//!   machine and clone.

use crate::lane_next::{decide, ActiveProposal, Snapshot};
use crate::lane_registry::{is_live, lock_existing, refuse_symlink, valid_name};
use serde_json::{json, Value};
use std::fs;
use std::io::Write as _;
use std::path::{Path, PathBuf};

/// How long a target stays offered to one lane: longer than the default `--wait` poll (432 s),
/// so a lane that waits is covered, and short enough that a lane that never acted frees it
/// within two polls.
pub const PROPOSAL_LEASE_SECS: i64 = 600;

const MAX_FILE_BYTES: u64 = 4_096;

/// The task kinds that are offered to any lane and so are kept.
const KINDS: [&str; 2] = ["reserve_audit", "pick_issue"];

/// The proposal folder inside the registry folder.
pub fn folder(registry: &Path) -> PathBuf {
    registry.join("proposals")
}

struct Record {
    lane: String,
    kind: String,
    number: u64,
    head: Option<String>,
    proposed_at: i64,
    lease_until: i64,
}

fn parse(text: &str, lane: &str) -> Option<Record> {
    let v: Value = serde_json::from_str(text).ok()?;
    let kind = v["kind"].as_str()?.to_string();
    if v["schema_version"].as_u64() != Some(1)
        || v["lane"].as_str() != Some(lane)
        || !KINDS.contains(&kind.as_str())
    {
        return None;
    }
    Some(Record {
        lane: lane.to_string(),
        kind,
        number: v["target"]["number"].as_u64()?,
        head: v["target"]["head"].as_str().map(str::to_string),
        proposed_at: v["proposed_at"].as_i64()?,
        lease_until: v["lease_until"].as_i64()?,
    })
}

/// Every record in the folder; a non-regular, oversized or malformed `<lane>.json` is an
/// error (never "absent"). Other files (a temporary file of a write in progress) are ignored.
fn read_all(registry: &Path) -> Result<Vec<Record>, &'static str> {
    let dir = folder(registry);
    refuse_symlink(&dir).map_err(|_| "proposal folder is a link")?;
    let entries = match fs::read_dir(&dir) {
        Ok(e) => e,
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => return Ok(Vec::new()),
        Err(_) => return Err("proposal folder unreadable"),
    };
    let mut out = Vec::new();
    for entry in entries {
        let entry = entry.map_err(|_| "proposal folder unreadable")?;
        let file_name = entry.file_name();
        let Some(lane) = file_name
            .to_str()
            .and_then(|n| n.strip_suffix(".json"))
            .filter(|l| valid_name(l))
        else {
            continue;
        };
        let path = entry.path();
        match fs::symlink_metadata(&path) {
            Ok(m) if m.file_type().is_file() && m.len() <= MAX_FILE_BYTES => {}
            _ => return Err("proposal file unusable"),
        }
        let text = fs::read_to_string(&path).map_err(|_| "proposal file unusable")?;
        out.push(parse(&text, lane).ok_or("proposal file malformed")?);
    }
    Ok(out)
}

fn write_record(registry: &Path, r: &Record) -> Result<(), &'static str> {
    let dir = folder(registry);
    refuse_symlink(&dir).map_err(|_| "proposal folder is a link")?;
    fs::create_dir_all(&dir).map_err(|_| "proposal not written")?;
    let file = dir.join(format!("{}.json", r.lane));
    refuse_symlink(&file).map_err(|_| "proposal file is a link")?;
    let mut target = json!({"number": r.number});
    if let Some(h) = &r.head {
        target["head"] = json!(h);
    }
    let body = json!({
        "schema_version": 1,
        "lane": r.lane,
        "kind": r.kind,
        "target": target,
        "proposed_at": r.proposed_at,
        "lease_until": r.lease_until,
    });
    let tmp = dir.join(format!("{}.json.tmp-{}", r.lane, std::process::id()));
    let written = fs::File::create(&tmp)
        .and_then(|mut f| {
            f.write_all(body.to_string().as_bytes())
                .and_then(|()| f.sync_all())
        })
        .and_then(|()| fs::rename(&tmp, &file));
    if written.is_err() {
        let _ = fs::remove_file(&tmp);
        return Err("proposal not written");
    }
    Ok(())
}

/// Decides for `lane` with the offers other live lanes hold, then keeps the offer made now.
/// `registry` is the session registry folder (it must exist); `now` is Unix seconds. The answer
/// is [`decide`]'s plus `proposal: {status, ...}`: `recorded`, `kept` (same target, first lease
/// kept), `none` (nothing offered, an older offer removed), `caller_not_registered` or
/// `unavailable` with a reason.
pub fn decide_and_record(registry: &Path, lane: &str, snap: &mut Snapshot, now: i64) -> Value {
    let unavailable = |snap: &Snapshot, why: &str| {
        let mut v = decide(lane, snap);
        v["proposal"] = json!({"status": "unavailable", "reason": why});
        v
    };
    if !valid_name(lane) {
        return unavailable(snap, "lane name is not a registry name");
    }
    let _lock = match lock_existing(registry) {
        Ok(l) => l,
        Err(_) => return unavailable(snap, "session registry unavailable"),
    };
    let records = match read_all(registry) {
        Ok(r) => r,
        Err(why) => return unavailable(snap, why),
    };
    let caller_live = match is_live(registry, lane, now) {
        Ok(b) => b,
        Err(_) => return unavailable(snap, "session registry entry unreadable"),
    };
    let mut offers = Vec::new();
    for r in records
        .iter()
        .filter(|r| r.lane != lane && r.lease_until > now)
    {
        match is_live(registry, &r.lane, now) {
            Ok(true) => offers.push(ActiveProposal {
                lane: r.lane.clone(),
                kind: r.kind.clone(),
                number: r.number,
                head: r.head.clone(),
                lease_until: r.lease_until,
            }),
            Ok(false) => {}
            Err(_) => return unavailable(snap, "session registry entry unreadable"),
        }
    }
    snap.proposals = offers;
    let mut v = decide(lane, snap);
    let own = records.iter().find(|r| r.lane == lane);
    let next = &v["next"];
    let kind = next["kind"].as_str().filter(|k| KINDS.contains(k));
    let proposal = match (kind, next["target"]["number"].as_u64()) {
        (Some(kind), Some(number)) => {
            let head = next["target"]["head"].as_str().map(str::to_string);
            let same = own.filter(|o| {
                o.kind == kind && o.number == number && o.head == head && o.lease_until > now
            });
            match same {
                Some(o) => json!({"status": "kept", "lease_until": o.lease_until}),
                None => {
                    let r = Record {
                        lane: lane.to_string(),
                        kind: kind.to_string(),
                        number,
                        head,
                        proposed_at: now,
                        lease_until: now + PROPOSAL_LEASE_SECS,
                    };
                    match write_record(registry, &r) {
                        Err(why) => json!({"status": "unavailable", "reason": why}),
                        Ok(()) if caller_live => {
                            json!({"status": "recorded", "lease_until": r.lease_until})
                        }
                        Ok(()) => {
                            json!({"status": "caller_not_registered", "lease_until": r.lease_until})
                        }
                    }
                }
            }
        }
        _ => {
            // Another kind of task (or none): an older offer of this lane is moot.
            if own.is_some() {
                let _ = fs::remove_file(folder(registry).join(format!("{lane}.json")));
            }
            json!({"status": "none"})
        }
    };
    v["proposal"] = proposal;
    v
}

#[cfg(test)]
mod tests {
    use super::*;
    use serde_json::json;

    const NOW: i64 = 1_800_000_000;

    fn registry() -> tempfile::TempDir {
        tempfile::TempDir::new().unwrap()
    }

    /// A live (or expired) registry entry for `name`.
    fn register(dir: &Path, name: &str, lease_until: i64) {
        let body = json!({"schema_version": 1, "name": name,
            "instance": crate::lane_registry::new_instance(), "lease_until": lease_until,
            "registered_at": NOW - 10});
        fs::write(dir.join(format!("{name}.json")), body.to_string()).unwrap();
    }

    fn head(c: char) -> String {
        c.to_string().repeat(40)
    }

    /// Two other lanes' pull requests with no auditor, oldest first.
    fn two_prs() -> Snapshot {
        serde_json::from_value(json!({"prs": [
            {"number": 40, "head": head('a'), "author_lane": "agent-28", "comments": []},
            {"number": 41, "head": head('b'), "author_lane": "agent-28", "comments": []},
        ]}))
        .unwrap()
    }

    fn target(v: &Value) -> (String, u64) {
        (
            v["next"]["kind"].as_str().unwrap().to_string(),
            v["next"]["target"]["number"].as_u64().unwrap(),
        )
    }

    #[test]
    fn a_second_live_lane_is_offered_the_next_target() {
        let dir = registry();
        register(dir.path(), "agent-1", NOW + 3000);
        register(dir.path(), "agent-2", NOW + 3000);
        let a = decide_and_record(dir.path(), "agent-1", &mut two_prs(), NOW);
        let b = decide_and_record(dir.path(), "agent-2", &mut two_prs(), NOW + 1);
        assert_eq!(target(&a), ("reserve_audit".into(), 40), "{a}");
        assert_eq!(target(&b), ("reserve_audit".into(), 41), "{b}");
        assert_eq!(a["proposal"]["status"], "recorded");
        assert_eq!(
            b["pr_skipped"][0]["reason"], "proposed_to_other_lane",
            "{b}"
        );
        assert_eq!(b["pr_skipped"][0]["by"], "agent-1");
    }

    #[test]
    fn the_offer_ends_at_its_lease_exactly() {
        let dir = registry();
        register(dir.path(), "agent-1", NOW + 3000);
        register(dir.path(), "agent-2", NOW + 3000);
        let a = decide_and_record(dir.path(), "agent-1", &mut two_prs(), NOW);
        let lease = a["proposal"]["lease_until"].as_i64().unwrap();
        assert_eq!(lease, NOW + PROPOSAL_LEASE_SECS);
        // One second before the lease ends the target is still taken; at the lease it is free.
        let before = decide_and_record(dir.path(), "agent-2", &mut two_prs(), lease - 1);
        assert_eq!(target(&before).1, 41, "{before}");
        // The record of agent-2 is now the live one for 41; remove it to look at 40 alone.
        fs::remove_file(folder(dir.path()).join("agent-2.json")).unwrap();
        let at = decide_and_record(dir.path(), "agent-2", &mut two_prs(), lease);
        assert_eq!(target(&at).1, 40, "{at}");
        let after = decide_and_record(dir.path(), "agent-2", &mut two_prs(), lease + 1);
        assert_eq!(target(&after).1, 40, "{after}");
    }

    #[test]
    fn an_offer_to_a_lane_that_is_not_live_blocks_nobody() {
        let dir = registry();
        register(dir.path(), "agent-2", NOW + 3000);
        // agent-1 was never registered: its offer is kept but not honoured.
        let a = decide_and_record(dir.path(), "agent-1", &mut two_prs(), NOW);
        assert_eq!(a["proposal"]["status"], "caller_not_registered", "{a}");
        let b = decide_and_record(dir.path(), "agent-2", &mut two_prs(), NOW + 1);
        assert_eq!(target(&b).1, 40, "{b}");
        // An entry that ends exactly at `now` is over; one second later than that is live.
        fs::remove_file(folder(dir.path()).join("agent-2.json")).unwrap();
        register(dir.path(), "agent-3", NOW + 5);
        let c = decide_and_record(dir.path(), "agent-3", &mut two_prs(), NOW);
        assert_eq!(c["proposal"]["status"], "recorded", "{c}");
        // At NOW + 5 agent-3's entry ends exactly now: its offer of 40 blocks nobody.
        let d = decide_and_record(dir.path(), "agent-2", &mut two_prs(), NOW + 5);
        assert_eq!(target(&d).1, 40, "{d}");
        // Had the entry run one second longer, agent-3 would still be live and hold 40.
        fs::remove_file(folder(dir.path()).join("agent-2.json")).unwrap();
        register(dir.path(), "agent-3", NOW + 6);
        let e = decide_and_record(dir.path(), "agent-2", &mut two_prs(), NOW + 5);
        assert_eq!(
            target(&e).1,
            41,
            "agent-3 is live (6 > 5) and holds 40: {e}"
        );
    }

    #[test]
    fn the_same_lane_asking_again_keeps_its_first_lease() {
        let dir = registry();
        register(dir.path(), "agent-1", NOW + 3000);
        let a = decide_and_record(dir.path(), "agent-1", &mut two_prs(), NOW);
        let b = decide_and_record(dir.path(), "agent-1", &mut two_prs(), NOW + 100);
        assert_eq!(target(&b).1, 40, "{b}");
        assert_eq!(b["proposal"]["status"], "kept");
        assert_eq!(b["proposal"]["lease_until"], a["proposal"]["lease_until"]);
        // After the lease the same lane gets it again with a fresh lease.
        let c = decide_and_record(dir.path(), "agent-1", &mut two_prs(), NOW + 700);
        assert_eq!(c["proposal"]["status"], "recorded", "{c}");
        assert_eq!(
            c["proposal"]["lease_until"],
            NOW + 700 + PROPOSAL_LEASE_SECS
        );
    }

    #[test]
    fn a_new_head_of_the_same_pull_request_is_not_covered_by_the_old_offer() {
        let dir = registry();
        register(dir.path(), "agent-1", NOW + 3000);
        register(dir.path(), "agent-2", NOW + 3000);
        decide_and_record(dir.path(), "agent-1", &mut two_prs(), NOW);
        let mut moved: Snapshot = serde_json::from_value(json!({"prs": [
            {"number": 40, "head": head('c'), "author_lane": "agent-28", "comments": []}]}))
        .unwrap();
        let b = decide_and_record(dir.path(), "agent-2", &mut moved, NOW + 1);
        assert_eq!(target(&b).1, 40, "{b}");
    }

    #[test]
    fn an_offer_of_a_lane_that_moved_on_is_removed() {
        let dir = registry();
        register(dir.path(), "agent-1", NOW + 3000);
        decide_and_record(dir.path(), "agent-1", &mut two_prs(), NOW);
        assert!(folder(dir.path()).join("agent-1.json").exists());
        let mut empty = Snapshot::default();
        let v = decide_and_record(dir.path(), "agent-1", &mut empty, NOW + 1);
        assert_eq!(v["proposal"]["status"], "none", "{v}");
        assert!(!folder(dir.path()).join("agent-1.json").exists());
    }

    #[test]
    fn issues_are_offered_once_too() {
        let dir = registry();
        register(dir.path(), "agent-1", NOW + 3000);
        register(dir.path(), "agent-2", NOW + 3000);
        let snap = || -> Snapshot {
            serde_json::from_value(json!({"open_issues": [
                {"number": 7, "title": "a", "labels": [], "comments": []},
                {"number": 8, "title": "b", "labels": [], "comments": []}]}))
            .unwrap()
        };
        let a = decide_and_record(dir.path(), "agent-1", &mut snap(), NOW);
        let b = decide_and_record(dir.path(), "agent-2", &mut snap(), NOW + 1);
        assert_eq!(target(&a), ("pick_issue".into(), 7), "{a}");
        assert_eq!(target(&b), ("pick_issue".into(), 8), "{b}");
        assert_eq!(b["skipped"][0]["reason"], "proposed_to_other_lane", "{b}");
    }

    #[test]
    fn an_issue_offer_does_not_cover_a_pull_request_with_the_same_number() {
        let dir = registry();
        register(dir.path(), "agent-1", NOW + 3000);
        register(dir.path(), "agent-2", NOW + 3000);
        let mut issue: Snapshot = serde_json::from_value(json!({"open_issues": [
            {"number": 40, "title": "a", "labels": [], "comments": []}]}))
        .unwrap();
        let a = decide_and_record(dir.path(), "agent-1", &mut issue, NOW);
        assert_eq!(target(&a), ("pick_issue".into(), 40), "{a}");
        let mut pr: Snapshot = serde_json::from_value(json!({"prs": [
            {"number": 40, "head": head('a'), "author_lane": "agent-28", "comments": []}]}))
        .unwrap();
        let b = decide_and_record(dir.path(), "agent-2", &mut pr, NOW + 1);
        assert_eq!(target(&b), ("reserve_audit".into(), 40), "{b}");
    }

    #[test]
    fn a_lane_name_that_is_not_a_registry_name_writes_nothing() {
        let dir = registry();
        let inner = dir.path().join("reg");
        fs::create_dir_all(&inner).unwrap();
        for lane in ["../escape", "a/b", "", "Agent One"] {
            let v = decide_and_record(&inner, lane, &mut two_prs(), NOW);
            assert_eq!(v["proposal"]["status"], "unavailable", "{lane:?}: {v}");
            assert_eq!(target(&v).1, 40, "the decision is made anyway: {v}");
        }
        assert!(!dir.path().join("escape.json").exists());
        assert!(
            !folder(&inner).exists(),
            "nothing was written under the registry"
        );
    }

    #[test]
    fn trouble_with_the_folder_leaves_the_decision_and_says_why() {
        // No registry folder at all.
        let dir = registry();
        let missing = dir.path().join("nope");
        let v = decide_and_record(&missing, "agent-1", &mut two_prs(), NOW);
        assert_eq!(v["proposal"]["status"], "unavailable", "{v}");
        assert_eq!(target(&v).1, 40);
        // A malformed, a symlinked and an oversized proposal file; the reason names none of
        // the file's text or path.
        let reg = registry();
        register(reg.path(), "agent-1", NOW + 3000);
        let proposals = folder(reg.path());
        fs::create_dir_all(&proposals).unwrap();
        fs::write(proposals.join("agent-2.json"), "{not json").unwrap();
        let v = decide_and_record(reg.path(), "agent-1", &mut two_prs(), NOW);
        assert_eq!(v["proposal"]["status"], "unavailable", "{v}");
        assert!(!v["proposal"]["reason"]
            .as_str()
            .unwrap()
            .contains("not json"));
        assert_eq!(target(&v).1, 40, "the decision is the old one: {v}");
        fs::remove_file(proposals.join("agent-2.json")).unwrap();
        #[cfg(unix)]
        {
            let elsewhere = reg.path().join("elsewhere.json");
            fs::write(&elsewhere, "{}").unwrap();
            std::os::unix::fs::symlink(&elsewhere, proposals.join("agent-2.json")).unwrap();
            let v = decide_and_record(reg.path(), "agent-1", &mut two_prs(), NOW);
            assert_eq!(v["proposal"]["status"], "unavailable", "link: {v}");
            fs::remove_file(proposals.join("agent-2.json")).unwrap();
        }
        // 4096 bytes is read (and found malformed), 4097 is refused before reading.
        for (len, reason) in [
            (4096, "proposal file malformed"),
            (4097, "proposal file unusable"),
        ] {
            fs::write(proposals.join("agent-2.json"), "x".repeat(len)).unwrap();
            let v = decide_and_record(reg.path(), "agent-1", &mut two_prs(), NOW);
            assert_eq!(v["proposal"]["reason"], reason, "{len}: {v}");
        }
    }

    #[test]
    fn two_lanes_asking_at_once_get_different_targets() {
        let dir = registry();
        register(dir.path(), "agent-1", NOW + 3000);
        register(dir.path(), "agent-2", NOW + 3000);
        let barrier = std::sync::Arc::new(std::sync::Barrier::new(2));
        let run = |lane: &'static str| {
            let reg = dir.path().to_path_buf();
            let barrier = barrier.clone();
            std::thread::spawn(move || {
                barrier.wait();
                let v = decide_and_record(&reg, lane, &mut two_prs(), NOW);
                target(&v).1
            })
        };
        let (a, b) = (run("agent-1"), run("agent-2"));
        let mut got = [a.join().unwrap(), b.join().unwrap()];
        got.sort_unstable();
        assert_eq!(got, [40, 41]);
    }
}
