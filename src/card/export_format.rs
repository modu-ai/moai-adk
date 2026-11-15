//! The file format of the card ledger export (#3070 K0, acceptance criterion AC6).
//!
//! `render` turns the rows of the five exported tables (`card`, `card_ac`, `card_note`,
//! `card_link`, `event`) into the bytes of an export folder; `parse` turns those bytes back
//! into the same rows. The two are exact inverses: rendering what `parse` returned gives the
//! same bytes, and the same ledger always gives the same bytes, whatever order its rows were
//! read in. Nothing here opens a ledger or touches a file; the commands do that.
//!
//! The folder holds `index.json`, one `C-<card_id>.md` per card and `events.jsonl`.
//! - Every text column is kept as one JSON string, exactly as stored. That includes the JSON
//!   columns `write_scope` and `event.payload`, so no number or key order is ever re-written.
//! - A card file is a marker line, one line of compact JSON (the card with its criteria, notes
//!   and links — the only part `parse` reads), a second marker line, then a human view made
//!   from the same values. Text the user typed can hold marker lines or a Markdown fence
//!   without effect: JSON escapes every line break, and the human view puts such text on one
//!   line.
//! - `index.json` lists the cards and the SHA-256 and size of every other file, so a
//!   truncated or edited file is named before anything is restored.
//! - `parse` renders what it read and compares the bytes with its input. A file that is not
//!   exactly what `render` writes (an edited human view, extra spaces, a field the format
//!   does not know) is refused as `Modified`, never half-accepted.
//!
//! Ids are kept as they are: restore must give back the same `card_id`, `seq`, `note_id`,
//! `link_id` and `event_id`.

use serde::{Deserialize, Serialize};
use sha2::{Digest, Sha256};
use std::collections::{BTreeMap, BTreeSet};
use std::fmt;

/// The version this module writes and reads.
pub const FORMAT: u32 = 1;

const CARD_MARK: &str = "<!-- moai-card-export 1 -->";
const VIEW_MARK: &str = "<!-- human view: ignored on restore -->";
const EVENTS_FILE: &str = "events.jsonl";

/// A row of `card`.
#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct Card {
    pub card_id: i64,
    pub title: String,
    pub body: String,
    pub kind: String,
    pub state: String,
    pub priority: String,
    /// The stored JSON array text, not a parsed value.
    pub write_scope: String,
    pub generation: i64,
    pub rev: i64,
    pub created_by: String,
    pub created_at: i64,
    pub updated_at: i64,
    pub closed_at: Option<i64>,
    pub source_host: Option<String>,
    pub source_project: Option<String>,
    pub source_number: Option<i64>,
    pub source_url: Option<String>,
}

/// A row of `card_ac`.
#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct Ac {
    pub ac_id: i64,
    pub card_id: i64,
    pub seq: i64,
    pub text: String,
    pub verify: String,
    pub status: String,
    pub evidence_ref: Option<String>,
    pub withdrawn_reason: Option<String>,
    pub updated_at: i64,
}

/// A row of `card_note`.
#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct Note {
    pub note_id: i64,
    pub card_id: i64,
    pub kind: String,
    pub author: String,
    pub origin: String,
    pub body: String,
    #[serde(rename = "ref")]
    pub reference: Option<String>,
    pub created_at: i64,
}

/// A row of `card_link`.
#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct Link {
    pub link_id: i64,
    pub from_card: i64,
    pub kind: String,
    pub to_card: Option<i64>,
    pub target_host: Option<String>,
    pub target_project: Option<String>,
    pub target_number: Option<i64>,
    pub pr_state: Option<String>,
    pub merge_ref: Option<String>,
    pub state: String,
    pub created_at: i64,
    pub updated_at: i64,
}

/// A row of `event`.
#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct Event {
    pub event_id: i64,
    pub event_key: String,
    pub at: i64,
    pub actor: String,
    pub card_id: Option<i64>,
    pub kind: String,
    /// The stored JSON text, not a parsed value.
    pub payload: String,
}

/// The five exported tables of one ledger, with the `meta` facts a restore checks.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Snapshot {
    pub project_id: String,
    pub schema_version: i64,
    pub cards: Vec<Card>,
    pub acs: Vec<Ac>,
    pub notes: Vec<Note>,
    pub links: Vec<Link>,
    pub events: Vec<Event>,
}

/// The bytes of an export folder, file by file.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Export {
    /// `index.json`.
    pub index: String,
    /// `(file name, contents)` of every card file, by card number.
    pub cards: Vec<(String, String)>,
    /// `events.jsonl`.
    pub events: String,
}

/// Why rows cannot be rendered or files cannot be read back.
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum ExportError {
    /// Two rows of one table share an id.
    Duplicate { table: &'static str, id: i64 },
    /// A row names a card that is not in the export.
    Orphan {
        table: &'static str,
        id: i64,
        card: i64,
    },
    /// The files are not what this format writes (unreadable JSON, wrong name, wrong order).
    Format(String),
    /// The index is a version this build does not read.
    Unsupported(u32),
    /// A file is missing from the folder.
    MissingFile(String),
    /// A file is in the folder but not in the index.
    ExtraFile(String),
    /// A file's SHA-256 or size differs from the index.
    Digest(String),
    /// A file reads fine but is not byte for byte what `render` writes.
    Modified(String),
}

impl fmt::Display for ExportError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::Duplicate { table, id } => write!(f, "{table}: id {id} appears twice"),
            Self::Orphan { table, id, card } => {
                write!(
                    f,
                    "{table}: row {id} names card {card}, which is not exported"
                )
            }
            Self::Format(m) => write!(f, "export format: {m}"),
            Self::Unsupported(v) => write!(
                f,
                "export format {v} is not supported (this build reads {FORMAT})"
            ),
            Self::MissingFile(n) => write!(f, "export file {n} is missing"),
            Self::ExtraFile(n) => write!(f, "export file {n} is not in the index"),
            Self::Digest(n) => write!(
                f,
                "export file {n} does not match the index (size or SHA-256)"
            ),
            Self::Modified(n) => write!(
                f,
                "export file {n} was edited: it is not what the exporter writes"
            ),
        }
    }
}

impl std::error::Error for ExportError {}

#[derive(Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
struct Index {
    format: u32,
    project_id: String,
    schema_version: i64,
    cards: Vec<IndexCard>,
    events: IndexEvents,
}

#[derive(Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
struct IndexCard {
    card_id: i64,
    file: String,
    title: String,
    state: String,
    source: Option<String>,
    bytes: usize,
    sha256: String,
}

#[derive(Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
struct IndexEvents {
    file: String,
    count: usize,
    bytes: usize,
    sha256: String,
}

#[derive(Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
struct CardFile {
    card: Card,
    acs: Vec<Ac>,
    notes: Vec<Note>,
    links: Vec<Link>,
}

fn sha256_hex(text: &str) -> String {
    format!("{:x}", Sha256::digest(text.as_bytes()))
}

fn card_file_name(card_id: i64) -> String {
    format!("C-{card_id}.md")
}

fn json<T: Serialize>(value: &T) -> Result<String, ExportError> {
    serde_json::to_string(value).map_err(|e| ExportError::Format(e.to_string()))
}

/// `rows` sorted by id, or the first id that appears twice.
fn sorted<T: Clone>(
    table: &'static str,
    rows: &[T],
    id: impl Fn(&T) -> i64,
) -> Result<Vec<T>, ExportError> {
    let mut out = rows.to_vec();
    out.sort_by_key(&id);
    if let Some(pair) = out.windows(2).find(|p| id(&p[0]) == id(&p[1])) {
        return Err(ExportError::Duplicate {
            table,
            id: id(&pair[0]),
        });
    }
    Ok(out)
}

/// The same rows in the order every export uses (by id), after checking that no id repeats
/// and that every row belongs to an exported card.
pub fn canonical(snapshot: &Snapshot) -> Result<Snapshot, ExportError> {
    let cards = sorted("card", &snapshot.cards, |c| c.card_id)?;
    let acs = sorted("card_ac", &snapshot.acs, |a| a.ac_id)?;
    let notes = sorted("card_note", &snapshot.notes, |n| n.note_id)?;
    let links = sorted("card_link", &snapshot.links, |l| l.link_id)?;
    let events = sorted("event", &snapshot.events, |e| e.event_id)?;
    let ids: BTreeSet<i64> = cards.iter().map(|c| c.card_id).collect();
    let check = |table: &'static str, id: i64, card: i64| {
        if ids.contains(&card) {
            Ok(())
        } else {
            Err(ExportError::Orphan { table, id, card })
        }
    };
    for a in &acs {
        check("card_ac", a.ac_id, a.card_id)?;
    }
    for n in &notes {
        check("card_note", n.note_id, n.card_id)?;
    }
    for l in &links {
        check("card_link", l.link_id, l.from_card)?;
        if let Some(to) = l.to_card {
            check("card_link", l.link_id, to)?;
        }
    }
    for e in &events {
        if let Some(card) = e.card_id {
            check("event", e.event_id, card)?;
        }
    }
    Ok(Snapshot {
        project_id: snapshot.project_id.clone(),
        schema_version: snapshot.schema_version,
        cards,
        acs,
        notes,
        links,
        events,
    })
}

/// A value on one line: every control character (a line break, a tab, an escape) becomes a
/// space. Used only by the human view, which `parse` never reads.
fn oneline(text: &str) -> String {
    text.chars()
        .map(|c| if c.is_control() { ' ' } else { c })
        .collect()
}

fn source_of(card: &Card) -> Option<String> {
    match (&card.source_host, &card.source_project, card.source_number) {
        (Some(host), Some(project), Some(number)) => Some(format!("{host}:{project}#{number}")),
        _ => None,
    }
}

fn human_view(file: &CardFile) -> String {
    let c = &file.card;
    let mut out = format!(
        "# C-{} · {}\n\nstate: {} · kind: {} · priority: {} · rev: {}\nsource: {}\n",
        c.card_id,
        oneline(&c.title),
        oneline(&c.state),
        oneline(&c.kind),
        oneline(&c.priority),
        c.rev,
        source_of(c).map_or_else(|| "none".to_string(), |s| oneline(&s)),
    );
    out.push_str("\n## Body\n\n");
    out.push_str(&oneline(&c.body));
    out.push_str("\n\n## Acceptance criteria\n\n");
    for a in &file.acs {
        out.push_str(&format!(
            "- AC{} [{}] {}\n",
            a.seq,
            oneline(&a.status),
            oneline(&a.text)
        ));
    }
    out.push_str("\n## Notes\n\n");
    for n in &file.notes {
        out.push_str(&format!(
            "- #{} {} by {} ({}): {}\n",
            n.note_id,
            oneline(&n.kind),
            oneline(&n.author),
            oneline(&n.origin),
            oneline(&n.body)
        ));
    }
    out.push_str("\n## Links\n\n");
    for l in &file.links {
        let target = match l.to_card {
            Some(to) => format!("C-{to}"),
            None => format!(
                "{}:{}#{}",
                oneline(l.target_host.as_deref().unwrap_or("")),
                oneline(l.target_project.as_deref().unwrap_or("")),
                l.target_number.unwrap_or(0)
            ),
        };
        out.push_str(&format!(
            "- {} {} ({})\n",
            oneline(&l.kind),
            target,
            oneline(&l.state)
        ));
    }
    // One line break at the end, whichever sections are empty.
    format!("{}\n", out.trim_end())
}

fn card_text(file: &CardFile) -> Result<String, ExportError> {
    Ok(format!(
        "{CARD_MARK}\n{}\n{VIEW_MARK}\n{}",
        json(file)?,
        human_view(file)
    ))
}

fn events_text(events: &[Event]) -> Result<String, ExportError> {
    let mut out = String::new();
    for e in events {
        out.push_str(&json(e)?);
        out.push('\n');
    }
    Ok(out)
}

fn group<T: Clone>(rows: &[T], card: impl Fn(&T) -> i64) -> BTreeMap<i64, Vec<T>> {
    let mut map: BTreeMap<i64, Vec<T>> = BTreeMap::new();
    for row in rows {
        map.entry(card(row)).or_default().push(row.clone());
    }
    map
}

/// The export folder of `snapshot`: the same bytes for the same rows, in whatever order they
/// were read.
pub fn render(snapshot: &Snapshot) -> Result<Export, ExportError> {
    let s = canonical(snapshot)?;
    let mut acs = group(&s.acs, |a| a.card_id);
    let mut notes = group(&s.notes, |n| n.card_id);
    let mut links = group(&s.links, |l| l.from_card);
    let mut files = Vec::new();
    let mut index_cards = Vec::new();
    for card in &s.cards {
        let id = card.card_id;
        let file = CardFile {
            card: card.clone(),
            acs: acs.remove(&id).unwrap_or_default(),
            notes: notes.remove(&id).unwrap_or_default(),
            links: links.remove(&id).unwrap_or_default(),
        };
        let text = card_text(&file)?;
        let name = card_file_name(id);
        index_cards.push(IndexCard {
            card_id: id,
            file: name.clone(),
            title: card.title.clone(),
            state: card.state.clone(),
            source: source_of(card),
            bytes: text.len(),
            sha256: sha256_hex(&text),
        });
        files.push((name, text));
    }
    let events = events_text(&s.events)?;
    let index = Index {
        format: FORMAT,
        project_id: s.project_id.clone(),
        schema_version: s.schema_version,
        cards: index_cards,
        events: IndexEvents {
            file: EVENTS_FILE.to_string(),
            count: s.events.len(),
            bytes: events.len(),
            sha256: sha256_hex(&events),
        },
    };
    let mut index_text =
        serde_json::to_string_pretty(&index).map_err(|e| ExportError::Format(e.to_string()))?;
    index_text.push('\n');
    Ok(Export {
        index: index_text,
        cards: files,
        events,
    })
}

fn check_digest(name: &str, text: &str, bytes: usize, sha256: &str) -> Result<(), ExportError> {
    if text.len() == bytes && sha256_hex(text) == sha256 {
        Ok(())
    } else {
        Err(ExportError::Digest(name.to_string()))
    }
}

/// The first line after the marker: the JSON of a card file.
fn card_json_line<'a>(name: &str, text: &'a str) -> Result<&'a str, ExportError> {
    let mut lines = text.split('\n');
    if lines.next() != Some(CARD_MARK) {
        return Err(ExportError::Format(format!(
            "{name}: first line is not the export marker"
        )));
    }
    lines
        .next()
        .ok_or_else(|| ExportError::Format(format!("{name}: no data line")))
}

/// The rows of an export folder. It refuses a folder that is incomplete, does not match its
/// index, or is not exactly what `render` writes.
pub fn parse(export: &Export) -> Result<Snapshot, ExportError> {
    let index: Index = serde_json::from_str(&export.index)
        .map_err(|e| ExportError::Format(format!("index.json: {e}")))?;
    if index.format != FORMAT {
        return Err(ExportError::Unsupported(index.format));
    }
    let listed: BTreeSet<&str> = index.cards.iter().map(|c| c.file.as_str()).collect();
    if listed.len() != index.cards.len() {
        return Err(ExportError::Format(
            "index.json lists a card file twice".into(),
        ));
    }
    if let Some((name, _)) = export
        .cards
        .iter()
        .find(|(n, _)| !listed.contains(n.as_str()))
    {
        return Err(ExportError::ExtraFile(name.clone()));
    }
    let mut cards = Vec::new();
    let mut acs = Vec::new();
    let mut notes = Vec::new();
    let mut links = Vec::new();
    let mut previous: Option<i64> = None;
    for entry in &index.cards {
        if entry.file != card_file_name(entry.card_id) {
            return Err(ExportError::Format(format!(
                "index.json: {} is not the file name of card {}",
                entry.file, entry.card_id
            )));
        }
        if previous.is_some_and(|p| p >= entry.card_id) {
            return Err(ExportError::Format(
                "index.json: cards are not in ascending order".into(),
            ));
        }
        previous = Some(entry.card_id);
        let (_, text) = export
            .cards
            .iter()
            .find(|(n, _)| *n == entry.file)
            .ok_or_else(|| ExportError::MissingFile(entry.file.clone()))?;
        check_digest(&entry.file, text, entry.bytes, &entry.sha256)?;
        let file: CardFile = serde_json::from_str(card_json_line(&entry.file, text)?)
            .map_err(|e| ExportError::Format(format!("{}: {e}", entry.file)))?;
        if file.card.card_id != entry.card_id {
            return Err(ExportError::Format(format!(
                "{}: holds another card",
                entry.file
            )));
        }
        cards.push(file.card);
        acs.extend(file.acs);
        notes.extend(file.notes);
        links.extend(file.links);
    }
    if index.events.file != EVENTS_FILE {
        return Err(ExportError::Format(
            "index.json: events file has another name".into(),
        ));
    }
    check_digest(
        EVENTS_FILE,
        &export.events,
        index.events.bytes,
        &index.events.sha256,
    )?;
    if !export.events.is_empty() && !export.events.ends_with('\n') {
        return Err(ExportError::Format(format!(
            "{EVENTS_FILE}: last line has no line break"
        )));
    }
    let events = export
        .events
        .split_terminator('\n')
        .map(|line| {
            serde_json::from_str::<Event>(line)
                .map_err(|e| ExportError::Format(format!("{EVENTS_FILE}: {e}")))
        })
        .collect::<Result<Vec<_>, _>>()?;
    if events.len() != index.events.count {
        return Err(ExportError::Format(format!(
            "{EVENTS_FILE}: line count differs from the index"
        )));
    }
    let snapshot = canonical(&Snapshot {
        project_id: index.project_id,
        schema_version: index.schema_version,
        cards,
        acs,
        notes,
        links,
        events,
    })?;
    // What was read must be exactly what the exporter would write for it.
    let again = render(&snapshot)?;
    for (name, text) in &again.cards {
        let same = export.cards.iter().any(|(n, t)| n == name && t == text);
        if !same {
            return Err(ExportError::Modified(name.clone()));
        }
    }
    if again.events != export.events {
        return Err(ExportError::Modified(EVENTS_FILE.into()));
    }
    // Last: the index holds the digests of the files, so a file edit shows up there too.
    if again.index != export.index {
        return Err(ExportError::Modified("index.json".into()));
    }
    Ok(snapshot)
}

#[cfg(test)]
mod tests {
    use super::*;

    fn card(id: i64) -> Card {
        Card {
            card_id: id,
            title: format!("Card {id}"),
            body: String::new(),
            kind: "task".into(),
            state: "draft".into(),
            priority: "p2".into(),
            write_scope: "[]".into(),
            generation: 0,
            rev: 1,
            created_by: "agent-7".into(),
            created_at: 1_700_000_000,
            updated_at: 1_700_000_000,
            closed_at: None,
            source_host: None,
            source_project: None,
            source_number: None,
            source_url: None,
        }
    }

    fn ac(id: i64, card: i64, seq: i64) -> Ac {
        Ac {
            ac_id: id,
            card_id: card,
            seq,
            text: format!("criterion {seq}"),
            verify: "run the test".into(),
            status: "open".into(),
            evidence_ref: None,
            withdrawn_reason: None,
            updated_at: 1_700_000_001,
        }
    }

    fn note(id: i64, card: i64) -> Note {
        Note {
            note_id: id,
            card_id: card,
            kind: "comment".into(),
            author: "agent-7".into(),
            origin: "local".into(),
            body: "a note".into(),
            reference: None,
            created_at: 1_700_000_002,
        }
    }

    fn link(id: i64, from: i64, to: Option<i64>) -> Link {
        let pr = to.is_none();
        Link {
            link_id: id,
            from_card: from,
            kind: if pr { "pr" } else { "blocks" }.into(),
            to_card: to,
            target_host: pr.then(|| "github".to_string()),
            target_project: pr.then(|| "modu-ai/moai-adk".to_string()),
            target_number: pr.then_some(3074),
            pr_state: pr.then(|| "merged".to_string()),
            merge_ref: pr.then(|| "abc123".to_string()),
            state: "active".into(),
            created_at: 1_700_000_003,
            updated_at: 1_700_000_004,
        }
    }

    fn event(id: i64, card: Option<i64>) -> Event {
        Event {
            event_id: id,
            event_key: format!("card_add:{}:0123abcd", card.unwrap_or(0)),
            at: 1_700_000_005,
            actor: "agent-7".into(),
            card_id: card,
            kind: "card_add".into(),
            payload: "{\"b\":1.0,\"a\":[1,2]}".into(),
        }
    }

    fn sample() -> Snapshot {
        let mut c2 = card(2);
        c2.source_host = Some("github".into());
        c2.source_project = Some("modu-ai/moai-adk".into());
        c2.source_number = Some(3066);
        c2.source_url = Some("https://github.com/modu-ai/moai-adk/issues/3066".into());
        c2.state = "done".into();
        c2.closed_at = Some(1_700_000_099);
        Snapshot {
            project_id: "p-1".into(),
            schema_version: 1,
            cards: vec![card(1), c2, card(5)],
            acs: vec![ac(1, 1, 1), ac(2, 1, 2), ac(3, 2, 1)],
            notes: vec![note(1, 1), note(2, 2)],
            links: vec![link(1, 1, Some(2)), link(2, 2, None)],
            events: vec![event(1, Some(1)), event(2, None), event(3, Some(2))],
        }
    }

    /// Text a user can type that looks like the file's own structure.
    const NASTY: &str = "line1\r\n---\n<!-- moai-card-export 1 -->\n<!-- human view: ignored on restore -->\n```\n## Notes\n- #9 x\n\u{1}\ttab \"quote\" \\ 한글 🗿 \u{2028}";

    fn nasty() -> Snapshot {
        let mut s = sample();
        s.cards[0].title = NASTY.into();
        s.cards[0].body = NASTY.into();
        s.acs[0].text = NASTY.into();
        s.acs[0].verify = NASTY.into();
        s.notes[0].body = NASTY.into();
        s.notes[0].author = NASTY.into();
        s.events[0].payload = format!("{{\"k\":{}}}", serde_json::to_string(NASTY).unwrap());
        s.events[0].actor = NASTY.into();
        s
    }

    /// Rebuild the index digests after a test edited a file, so only the edit is wrong.
    fn redigest(e: &mut Export) {
        let mut index: Index = serde_json::from_str(&e.index).unwrap();
        for entry in &mut index.cards {
            let text = &e.cards.iter().find(|(n, _)| *n == entry.file).unwrap().1;
            entry.bytes = text.len();
            entry.sha256 = sha256_hex(text);
        }
        index.events.bytes = e.events.len();
        index.events.sha256 = sha256_hex(&e.events);
        e.index = serde_json::to_string_pretty(&index).unwrap() + "\n";
    }

    #[test]
    fn rows_come_back_exactly_and_the_bytes_do_not_change() {
        for s in [sample(), nasty()] {
            let out = render(&s).unwrap();
            let back = parse(&out).unwrap();
            assert_eq!(back, canonical(&s).unwrap());
            assert_eq!(render(&back).unwrap(), out);
        }
    }

    #[test]
    fn the_same_rows_in_any_read_order_give_the_same_bytes() {
        let a = sample();
        let mut b = sample();
        b.cards.reverse();
        b.acs.reverse();
        b.notes.reverse();
        b.links.reverse();
        b.events.reverse();
        assert_eq!(render(&a).unwrap(), render(&b).unwrap());
        assert_eq!(render(&a).unwrap(), render(&a).unwrap());
    }

    #[test]
    fn an_empty_ledger_round_trips() {
        let s = Snapshot {
            project_id: "p-1".into(),
            schema_version: 1,
            cards: vec![],
            acs: vec![],
            notes: vec![],
            links: vec![],
            events: vec![],
        };
        let out = render(&s).unwrap();
        assert!(out.cards.is_empty());
        assert_eq!(out.events, "");
        assert_eq!(parse(&out).unwrap(), s);
    }

    #[test]
    fn nothing_and_the_empty_string_stay_different() {
        let mut s = sample();
        s.acs[0].evidence_ref = Some(String::new());
        s.acs[1].evidence_ref = None;
        s.notes[0].reference = Some(String::new());
        s.notes[1].reference = None;
        let back = parse(&render(&s).unwrap()).unwrap();
        assert_eq!(back.acs[0].evidence_ref, Some(String::new()));
        assert_eq!(back.acs[1].evidence_ref, None);
        assert_eq!(back.notes[0].reference, Some(String::new()));
        assert_eq!(back.notes[1].reference, None);
    }

    #[test]
    fn the_stored_json_text_is_kept_character_for_character() {
        let mut s = sample();
        s.cards[0].write_scope = "[ \"src/a.rs\" ,\"src/b.rs\"]".into();
        s.events[0].payload = "{\"z\": 1.50, \"a\" :[ ]}".into();
        let back = parse(&render(&s).unwrap()).unwrap();
        assert_eq!(back.cards[0].write_scope, "[ \"src/a.rs\" ,\"src/b.rs\"]");
        assert_eq!(back.events[0].payload, "{\"z\": 1.50, \"a\" :[ ]}");
    }

    #[test]
    fn ids_and_times_at_the_limits_come_back() {
        let mut s = sample();
        s.cards[0].card_id = i64::MAX;
        s.acs[0].card_id = i64::MAX;
        s.acs[1].card_id = i64::MAX;
        s.notes[0].card_id = i64::MAX;
        s.links[0].from_card = i64::MAX;
        s.events[0].card_id = Some(i64::MAX);
        s.cards[0].created_at = 0;
        s.cards[0].rev = i64::MAX;
        s.notes[0].note_id = i64::MAX;
        s.events[0].at = i64::MAX;
        let back = parse(&render(&s).unwrap()).unwrap();
        assert_eq!(back, canonical(&s).unwrap());
        assert!(back.cards.iter().any(|c| c.card_id == i64::MAX));
    }

    #[test]
    fn typed_text_cannot_break_the_file_or_its_human_view() {
        let out = render(&nasty()).unwrap();
        for (_, text) in &out.cards {
            let lines: Vec<&str> = text.split('\n').collect();
            assert_eq!(lines[0], CARD_MARK);
            assert!(
                lines[1].starts_with('{') && lines[1].ends_with('}'),
                "{}",
                lines[1]
            );
            assert_eq!(lines[2], VIEW_MARK);
            assert!(text.ends_with('\n') && !text.ends_with("\n\n"), "{text:?}");
            assert!(!text.contains('\r'));
        }
        assert!(!out.events.contains('\r'));
        assert_eq!(out.events.matches('\n').count(), 3);
        assert!(!out.index.contains('\r') && out.index.ends_with("}\n"));
    }

    #[test]
    fn the_index_lists_the_cards_in_order_with_their_source() {
        let out = render(&sample()).unwrap();
        let index: Index = serde_json::from_str(&out.index).unwrap();
        let ids: Vec<i64> = index.cards.iter().map(|c| c.card_id).collect();
        assert_eq!(ids, vec![1, 2, 5]);
        assert_eq!(index.cards[0].source, None);
        assert_eq!(
            index.cards[1].source.as_deref(),
            Some("github:modu-ai/moai-adk#3066")
        );
        assert_eq!(index.cards[1].file, "C-2.md");
        assert_eq!(index.events.count, 3);
        assert_eq!(index.format, FORMAT);
        let names: Vec<&str> = out.cards.iter().map(|(n, _)| n.as_str()).collect();
        assert_eq!(names, vec!["C-1.md", "C-2.md", "C-5.md"]);
    }

    #[test]
    fn rows_that_name_no_exported_card_are_refused() {
        type Edit = fn(&mut Snapshot);
        let cases: [(Edit, &str); 5] = [
            (|s| s.acs[0].card_id = 99, "card_ac"),
            (|s| s.notes[0].card_id = 99, "card_note"),
            (|s| s.links[0].from_card = 99, "card_link"),
            (|s| s.links[0].to_card = Some(99), "card_link"),
            (|s| s.events[0].card_id = Some(99), "event"),
        ];
        for (edit, table) in cases {
            let mut s = sample();
            edit(&mut s);
            match render(&s) {
                Err(ExportError::Orphan {
                    table: t, card: 99, ..
                }) => assert_eq!(t, table),
                other => panic!("{table}: {other:?}"),
            }
        }
    }

    #[test]
    fn a_repeated_id_is_refused_in_every_table() {
        type Edit = fn(&mut Snapshot);
        let cases: [(Edit, &str); 5] = [
            (|s| s.cards[1].card_id = 1, "card"),
            (|s| s.acs[1].ac_id = 1, "card_ac"),
            (|s| s.notes[1].note_id = 1, "card_note"),
            (|s| s.links[1].link_id = 1, "card_link"),
            (|s| s.events[1].event_id = 1, "event"),
        ];
        for (edit, table) in cases {
            let mut s = sample();
            edit(&mut s);
            match render(&s) {
                Err(ExportError::Duplicate { table: t, id: 1 }) => assert_eq!(t, table),
                other => panic!("{table}: {other:?}"),
            }
        }
    }

    #[test]
    fn an_edited_human_view_is_refused_even_with_a_matching_index() {
        let mut out = render(&sample()).unwrap();
        out.cards[0].1 = out.cards[0].1.replace("# C-1 · Card 1", "# C-1 · Card one");
        redigest(&mut out);
        assert_eq!(parse(&out), Err(ExportError::Modified("C-1.md".into())));
    }

    #[test]
    fn a_data_line_that_is_valid_but_not_what_the_exporter_writes_is_refused() {
        let mut out = render(&sample()).unwrap();
        out.cards[0].1 = out.cards[0]
            .1
            .replacen("\"card_id\":1,", "\"card_id\": 1,", 1);
        redigest(&mut out);
        assert_eq!(parse(&out), Err(ExportError::Modified("C-1.md".into())));
    }

    #[test]
    fn an_edit_the_index_does_not_know_about_is_named_as_a_digest_error() {
        let mut out = render(&sample()).unwrap();
        out.cards[1].1 = out.cards[1].1.replace("Card 2", "Card X");
        assert_eq!(parse(&out), Err(ExportError::Digest("C-2.md".into())));
        let mut out = render(&sample()).unwrap();
        out.events.push('\n');
        assert_eq!(parse(&out), Err(ExportError::Digest(EVENTS_FILE.into())));
    }

    #[test]
    fn a_truncated_or_changed_size_is_a_digest_error() {
        let mut out = render(&sample()).unwrap();
        let text = out.cards[0].1.clone();
        out.cards[0].1 = text[..text.len() - 1].to_string();
        assert_eq!(parse(&out), Err(ExportError::Digest("C-1.md".into())));
    }

    #[test]
    fn a_missing_file_and_an_extra_file_are_named() {
        let mut out = render(&sample()).unwrap();
        let (name, _) = out.cards.remove(1);
        assert_eq!(parse(&out), Err(ExportError::MissingFile(name)));
        let mut out = render(&sample()).unwrap();
        out.cards.push(("C-9.md".into(), "x".into()));
        assert_eq!(parse(&out), Err(ExportError::ExtraFile("C-9.md".into())));
    }

    #[test]
    fn another_format_version_is_not_read() {
        let mut out = render(&sample()).unwrap();
        out.index = out.index.replacen("\"format\": 1", "\"format\": 2", 1);
        assert_eq!(parse(&out), Err(ExportError::Unsupported(2)));
        let mut out = render(&sample()).unwrap();
        out.index = out.index.replacen("\"format\": 1", "\"format\": 0", 1);
        assert_eq!(parse(&out), Err(ExportError::Unsupported(0)));
    }

    #[test]
    fn a_field_the_format_does_not_know_is_refused() {
        let mut out = render(&sample()).unwrap();
        out.cards[0].1 = out.cards[0]
            .1
            .replacen("\"rev\":1,", "\"rev\":1,\"extra\":true,", 1);
        redigest(&mut out);
        assert!(
            matches!(parse(&out), Err(ExportError::Format(_))),
            "{:?}",
            parse(&out)
        );
    }

    #[test]
    fn a_wrong_file_name_or_order_in_the_index_is_refused() {
        let mut out = render(&sample()).unwrap();
        out.index = out.index.replacen("\"C-1.md\"", "\"C-7.md\"", 1);
        out.cards[0].0 = "C-7.md".into();
        assert!(
            matches!(parse(&out), Err(ExportError::Format(_))),
            "{:?}",
            parse(&out)
        );
        let mut out = render(&sample()).unwrap();
        let mut index: Index = serde_json::from_str(&out.index).unwrap();
        index.cards.swap(0, 1);
        out.index = serde_json::to_string_pretty(&index).unwrap() + "\n";
        assert!(
            matches!(parse(&out), Err(ExportError::Format(_))),
            "{:?}",
            parse(&out)
        );
    }

    #[test]
    fn a_card_file_that_holds_another_card_is_refused() {
        let mut out = render(&sample()).unwrap();
        let other = out.cards[1].1.clone();
        out.cards[0].1 = other;
        redigest(&mut out);
        assert!(
            matches!(parse(&out), Err(ExportError::Format(_))),
            "{:?}",
            parse(&out)
        );
    }

    #[test]
    fn a_missing_marker_or_event_line_break_is_refused() {
        let mut out = render(&sample()).unwrap();
        out.cards[0].1 = out.cards[0].1.replacen(CARD_MARK, "<!-- other -->", 1);
        redigest(&mut out);
        assert!(
            matches!(parse(&out), Err(ExportError::Format(_))),
            "{:?}",
            parse(&out)
        );
        let mut out = render(&sample()).unwrap();
        out.events.pop();
        redigest(&mut out);
        assert!(
            matches!(parse(&out), Err(ExportError::Format(_))),
            "{:?}",
            parse(&out)
        );
    }

    #[test]
    fn the_event_count_in_the_index_must_match_the_lines() {
        let mut out = render(&sample()).unwrap();
        out.index = out.index.replacen("\"count\": 3", "\"count\": 2", 1);
        assert!(
            matches!(parse(&out), Err(ExportError::Format(_))),
            "{:?}",
            parse(&out)
        );
    }

    #[test]
    fn the_human_view_puts_typed_text_on_one_line() {
        let out = render(&nasty()).unwrap();
        let view = out.cards[0]
            .1
            .split(&format!("{VIEW_MARK}\n"))
            .nth(1)
            .unwrap();
        assert!(!view.contains("\n---\n"));
        assert!(!view.contains("\n```"));
        // The typed text holds "## Notes", but never on a line of its own.
        for heading in ["## Body", "## Acceptance criteria", "## Notes", "## Links"] {
            assert_eq!(
                view.lines().filter(|l| *l == heading).count(),
                1,
                "{heading}"
            );
        }
    }

    #[test]
    fn an_index_that_is_not_what_the_exporter_writes_is_refused() {
        let mut out = render(&sample()).unwrap();
        out.index = out.index.replacen("{\n", "{ \n", 1);
        assert_eq!(parse(&out), Err(ExportError::Modified("index.json".into())));
    }

    #[test]
    fn an_events_file_that_is_valid_but_not_what_the_exporter_writes_is_refused() {
        let mut out = render(&sample()).unwrap();
        out.events = out
            .events
            .replacen("{\"event_id\":1,", "{\"event_id\": 1,", 1);
        assert_ne!(out.events, render(&sample()).unwrap().events);
        redigest(&mut out);
        assert_eq!(parse(&out), Err(ExportError::Modified(EVENTS_FILE.into())));
    }

    #[test]
    fn the_events_file_must_keep_its_name_in_the_index() {
        let mut out = render(&sample()).unwrap();
        out.index = out.index.replacen("events.jsonl", "other.jsonl", 1);
        assert!(
            matches!(parse(&out), Err(ExportError::Format(_))),
            "{:?}",
            parse(&out)
        );
    }
}
