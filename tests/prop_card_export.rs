//! Card ledger export property tests (SPEC-20261120-property-tests-wave2): what `render` writes,
//! `parse` reads back as the canonical rows; the bytes do not depend on the order the rows were
//! read in; and an export with one character changed is refused, never half-accepted.
//!
//! Every property is a checker that returns a result, so the canary test can feed the same
//! checker a deliberately broken variant and require it to fail. A failing case is replayed with
//! `PROP_SEED=<seed> PROP_ONLY=<i>`.

#[path = "common/prop.rs"]
mod prop;

use moai_cli::card::export_format::{
    canonical, parse, render, Ac, Card, Event, Export, ExportError, Link, Note, Snapshot,
};
use prop::{run_property, Rng, MIN_CASES};

type Render = dyn Fn(&Snapshot) -> Result<Export, ExportError>;
type Parse = dyn Fn(&Export) -> Result<Snapshot, ExportError>;

/// Text the user may type: line breaks, quotes, a backslash, both marker lines of the format, a
/// Markdown fence, an escape, non-ASCII and an emoji.
const PIECES: &[&str] = &[
    "",
    "title",
    "multi\nline",
    "with \"quotes\" and \\ backslash",
    "<!-- moai-card-export 1 -->",
    "<!-- human view: ignored on restore -->",
    "```",
    "한글 제목",
    "tab\tchar",
    "\u{1b}[31m",
    "{\"a\":1}",
    "[\"zone/**\"]",
    "line one\r\nline two",
    "emoji 🙂",
];

fn text(rng: &mut Rng) -> String {
    let count = 1 + rng.below(3) as usize;
    let mut out = String::new();
    for i in 0..count {
        if i > 0 {
            let separator: &&str = rng.pick(&["", " ", "\n"]);
            out.push_str(separator);
        }
        let piece: &&str = rng.pick(PIECES);
        out.push_str(piece);
    }
    out
}

fn opt_text(rng: &mut Rng) -> Option<String> {
    rng.bool().then(|| text(rng))
}

fn opt_number(rng: &mut Rng) -> Option<i64> {
    rng.bool().then(|| rng.below(1_000_000) as i64)
}

fn number(rng: &mut Rng) -> i64 {
    rng.below(2_000_000_000) as i64
}

/// Ids that never repeat but do not run in a row, so a sort by id has work to do.
struct Ids(i64);

impl Ids {
    fn new(rng: &mut Rng) -> Self {
        Ids(1 + rng.below(1000) as i64)
    }

    fn next(&mut self, rng: &mut Rng) -> i64 {
        let id = self.0;
        self.0 += 1 + rng.below(9) as i64;
        id
    }
}

/// The stored JSON array text of a write scope, or any text (the format keeps it as it is).
fn write_scope(rng: &mut Rng) -> String {
    if rng.below(3) == 0 {
        return text(rng);
    }
    let paths: Vec<String> = (0..rng.below(4))
        .map(|_| format!("\"{}/**\"", rng.name(6)))
        .collect();
    format!("[{}]", paths.join(","))
}

fn card(rng: &mut Rng, card_id: i64) -> Card {
    Card {
        card_id,
        title: text(rng),
        body: text(rng),
        kind: (*rng.pick(&["fix", "feature"])).to_string(),
        state: (*rng.pick(&["ready", "active", "review", "done"])).to_string(),
        priority: (*rng.pick(&["p0", "p1", "p2", "p3"])).to_string(),
        write_scope: write_scope(rng),
        generation: rng.below(20) as i64,
        rev: rng.below(20) as i64,
        created_by: text(rng),
        created_at: number(rng),
        updated_at: number(rng),
        closed_at: opt_number(rng),
        source_host: opt_text(rng),
        source_project: opt_text(rng),
        source_number: opt_number(rng),
        source_url: opt_text(rng),
    }
}

/// A snapshot of one to four cards with their criteria, notes and links, and a few events;
/// every id is unique and every row belongs to an exported card.
fn snapshot(rng: &mut Rng) -> Snapshot {
    let mut card_ids = Ids::new(rng);
    let mut cards = Vec::new();
    for _ in 0..1 + rng.below(4) {
        let id = card_ids.next(rng);
        cards.push(card(rng, id));
    }
    let ids: Vec<i64> = cards.iter().map(|c| c.card_id).collect();
    let (mut ac_ids, mut note_ids, mut link_ids, mut event_ids) =
        (Ids::new(rng), Ids::new(rng), Ids::new(rng), Ids::new(rng));
    let (mut acs, mut notes, mut links, mut events) =
        (Vec::new(), Vec::new(), Vec::new(), Vec::new());
    for &card_id in &ids {
        for seq in 1..=rng.below(4) as i64 {
            let ac_id = ac_ids.next(rng);
            acs.push(Ac {
                ac_id,
                card_id,
                seq,
                text: text(rng),
                verify: text(rng),
                status: (*rng.pick(&["open", "met", "withdrawn"])).to_string(),
                evidence_ref: opt_text(rng),
                withdrawn_reason: opt_text(rng),
                updated_at: number(rng),
            });
        }
        for _ in 0..rng.below(4) {
            let note_id = note_ids.next(rng);
            notes.push(Note {
                note_id,
                card_id,
                kind: (*rng.pick(&["report", "decision", "comment", "observation"])).to_string(),
                author: text(rng),
                origin: (*rng.pick(&["local", "github"])).to_string(),
                body: text(rng),
                reference: opt_text(rng),
                created_at: number(rng),
            });
        }
        for _ in 0..rng.below(3) {
            let link_id = link_ids.next(rng);
            let to_card = rng.bool().then(|| *rng.pick(&ids));
            links.push(Link {
                link_id,
                from_card: card_id,
                kind: (*rng.pick(&["pr", "blocks", "relates"])).to_string(),
                to_card,
                target_host: opt_text(rng),
                target_project: opt_text(rng),
                target_number: opt_number(rng),
                pr_state: opt_text(rng),
                merge_ref: opt_text(rng),
                state: (*rng.pick(&["open", "merged"])).to_string(),
                created_at: number(rng),
                updated_at: number(rng),
            });
        }
    }
    for _ in 0..rng.below(7) {
        let event_id = event_ids.next(rng);
        events.push(Event {
            event_id,
            event_key: format!("key-{event_id}"),
            at: number(rng),
            actor: text(rng),
            card_id: rng.bool().then(|| *rng.pick(&ids)),
            kind: (*rng.pick(&["card.take", "card.note", "card.review"])).to_string(),
            payload: text(rng),
        });
    }
    let project_id = rng.name(8);
    let ordered = Snapshot {
        project_id,
        schema_version: 1,
        cards,
        acs,
        notes,
        links,
        events,
    };
    permuted(rng, &ordered)
}

/// The same rows with every table in another order.
fn permuted(rng: &mut Rng, s: &Snapshot) -> Snapshot {
    Snapshot {
        project_id: s.project_id.clone(),
        schema_version: s.schema_version,
        cards: rng.shuffled(&s.cards),
        acs: rng.shuffled(&s.acs),
        notes: rng.shuffled(&s.notes),
        links: rng.shuffled(&s.links),
        events: rng.shuffled(&s.events),
    }
}

/// REQ-17: the rows read back are the canonical rows, canonical order is stable, and rendering
/// what was read gives the bytes that were read.
fn check_round_trip(render: &Render, parse: &Parse, rng: &mut Rng) -> Result<(), String> {
    let s = snapshot(rng);
    let canon = canonical(&s).map_err(|e| format!("a valid snapshot was refused: {e}"))?;
    let written = render(&s).map_err(|e| format!("a valid snapshot was not rendered: {e}"))?;
    let read = parse(&written).map_err(|e| format!("a rendered export was refused: {e}"))?;
    if read != canon {
        return Err(format!(
            "the rows read back differ from the canonical rows of {s:?}"
        ));
    }
    if canonical(&canon).as_ref() != Ok(&canon) {
        return Err("canonical order is not stable".into());
    }
    // The order itself, stated apart from `canonical`: every kind of row ascends by its own id.
    let ascending = |name: &str, ids: Vec<i64>| -> Result<(), String> {
        if ids.windows(2).all(|pair| pair[0] < pair[1]) {
            Ok(())
        } else {
            Err(format!(
                "the {name} rows are not in ascending id order: {ids:?}"
            ))
        }
    };
    ascending("card", read.cards.iter().map(|r| r.card_id).collect())?;
    ascending("criterion", read.acs.iter().map(|r| r.ac_id).collect())?;
    ascending("note", read.notes.iter().map(|r| r.note_id).collect())?;
    ascending("link", read.links.iter().map(|r| r.link_id).collect())?;
    ascending("event", read.events.iter().map(|r| r.event_id).collect())?;
    let again = render(&read).map_err(|e| format!("the rows read were not rendered: {e}"))?;
    if again != written {
        return Err("rendering the rows read gave other bytes than were read".into());
    }
    Ok(())
}

/// REQ-19: the bytes do not depend on the order the rows came in.
fn check_order(render: &Render, rng: &mut Rng) -> Result<(), String> {
    let s = snapshot(rng);
    let other = permuted(rng, &s);
    let a = render(&s).map_err(|e| format!("a valid snapshot was not rendered: {e}"))?;
    let b = render(&other).map_err(|e| format!("a reordered snapshot was not rendered: {e}"))?;
    if a != b {
        return Err(format!("two orders of {s:?} rendered differently"));
    }
    Ok(())
}

/// REQ-21: one changed character in any file and the export is refused, or it reads back with
/// every row kept and only the index's own facts (project id, schema version) changed.
fn check_edit(parse: &Parse, rng: &mut Rng) -> Result<(), String> {
    let s = snapshot(rng);
    let written = render(&s).map_err(|e| format!("a valid snapshot was not rendered: {e}"))?;
    let mut files: Vec<(String, String)> = vec![("index.json".into(), written.index.clone())];
    files.extend(written.cards.iter().cloned());
    if !written.events.is_empty() {
        files.push(("events.jsonl".into(), written.events.clone()));
    }
    let (name, original) = files[rng.below(files.len() as u64) as usize].clone();
    let chars: Vec<char> = original.chars().collect();
    let at = rng.below(chars.len() as u64) as usize;
    let replacement = if chars[at] == 'x' { 'y' } else { 'x' };
    let edited: String = chars
        .iter()
        .enumerate()
        .map(|(i, c)| if i == at { replacement } else { *c })
        .collect();
    let mut export = written.clone();
    match name.as_str() {
        "index.json" => export.index = edited,
        "events.jsonl" => export.events = edited,
        _ => {
            for (file, contents) in &mut export.cards {
                if *file == name {
                    *contents = edited.clone();
                }
            }
        }
    }
    match parse(&export) {
        Err(_) => Ok(()),
        // The project id and the schema version live in the index only, with no digest, so an
        // edit of either is a different but valid export. Whatever else is accepted must keep
        // every row and render back to exactly the edited bytes.
        Ok(read) => {
            let canon = canonical(&s).map_err(|e| format!("a valid snapshot was refused: {e}"))?;
            let rows_kept = read.cards == canon.cards
                && read.acs == canon.acs
                && read.notes == canon.notes
                && read.links == canon.links
                && read.events == canon.events;
            let bytes_kept = render(&read).as_ref() == Ok(&export);
            if rows_kept && bytes_kept {
                Ok(())
            } else {
                Err(format!(
                    "{name} with character {at} changed to {replacement:?} was accepted and \
                     changed a row (rows kept: {rows_kept}, bytes kept: {bytes_kept})"
                ))
            }
        }
    }
}

fn pass(result: Result<(), String>) {
    if let Err(reason) = result {
        panic!("{reason}");
    }
}

#[test]
fn a_rendered_export_reads_back_as_the_canonical_rows() {
    run_property("export_round_trip", MIN_CASES, |rng| {
        pass(check_round_trip(&render, &parse, rng));
    });
}

#[test]
fn the_export_bytes_do_not_depend_on_row_order() {
    run_property("export_order", MIN_CASES, |rng| {
        pass(check_order(&render, rng));
    });
}

#[test]
fn an_export_with_one_character_changed_is_refused() {
    run_property("export_edit", MIN_CASES, |rng| {
        pass(check_edit(&parse, rng));
    });
}

/// How many of `cases` the checker fails, on its own fixed seed.
fn failures(cases: u64, mut checker: impl FnMut(&mut Rng) -> Result<(), String>) -> u64 {
    (0..cases)
        .filter(|i| {
            let mut rng = Rng::new(0xCA4A_2028 ^ i.wrapping_mul(0x9e37_79b9_7f4a_7c15 | 1));
            checker(&mut rng).is_err()
        })
        .count() as u64
}

/// The canary of the suite: every checker fails on a deliberately broken variant.
#[test]
fn canary_every_checker_fails_on_its_broken_variant() {
    let lossy = |s: &Snapshot| {
        let mut t = s.clone();
        if let Some(note) = t.notes.last_mut() {
            note.body.push('x');
        }
        render(&t)
    };
    assert!(
        failures(MIN_CASES, |rng| check_round_trip(&lossy, &parse, rng)) > 0,
        "the round-trip checker accepted a renderer that changes a note"
    );
    let first_only = |s: &Snapshot| {
        let mut t = s.clone();
        t.events.truncate(1);
        render(&t)
    };
    assert!(
        failures(MIN_CASES, |rng| check_order(&first_only, rng)) > 0,
        "the order checker accepted a renderer that keeps the first event it was given"
    );
    let lenient = |e: &Export| {
        parse(e).or_else(|_| {
            Ok(Snapshot {
                project_id: String::new(),
                schema_version: 0,
                cards: Vec::new(),
                acs: Vec::new(),
                notes: Vec::new(),
                links: Vec::new(),
                events: Vec::new(),
            })
        })
    };
    assert!(
        failures(MIN_CASES, |rng| check_edit(&lenient, rng)) > 0,
        "the edit checker accepted a parser that never refuses"
    );
}
