//! Commit trailer property tests (SPEC-20261120-property-tests-wave2): the log record reads back
//! as written, the lines `trailers::format_verified` prints are lines the checker accepts, and a
//! raw separator inside a commit message never gives a silent wrong reading.
//!
//! Every property is a checker that returns a result, so the canary test can feed the same
//! checker a deliberately broken variant and require it to fail. A failing case is replayed with
//! `PROP_SEED=<seed> PROP_ONLY=<i>`.

#[path = "common/prop.rs"]
mod prop;

use moai_cli::session_current::SOURCES;
use moai_cli::trailers::{
    format_verified, judge, parse_log, verified_kinds, CommitTrailers, SESSION_KEY, SPEC_KEY,
    VERIFIED_KEY, VERIFIED_KINDS,
};
use prop::{run_property, Rng, MIN_CASES};
use serde_json::Value;

type Parse = dyn Fn(&str) -> Vec<CommitTrailers>;
type Getter<'a> = &'a dyn Fn(&str) -> Option<String>;
type FormatVerified = dyn Fn(&str, &[String], Getter<'_>, Option<&str>) -> Result<Value, String>;

const HEX: &[u8] = b"0123456789abcdef";
const ID_ALPHABET: &[u8] = b"0123456789abcdefABCDEF-";
const WORD_ALPHABET: &[u8] = b"abcdefghijklmnopqrstuvwxyz0123456789";
const PIECES: &[&str] = &[
    "fix",
    "trailers",
    "(parse)",
    "감사",
    "SPEC-20261120-x",
    "\"q\"",
    "tab\tz",
    "#",
    ":",
    "a: b",
    "{}",
    "—",
];

fn hex(rng: &mut Rng, len: usize) -> String {
    (0..len)
        .map(|_| HEX[rng.below(HEX.len() as u64) as usize] as char)
        .collect()
}

fn word(rng: &mut Rng) -> String {
    let len = 1 + rng.below(6) as usize;
    (0..len)
        .map(|_| WORD_ALPHABET[rng.below(WORD_ALPHABET.len() as u64) as usize] as char)
        .collect()
}

/// `SPEC-<yyyymmdd>-<slug>`: a real calendar date and a slug of up to three words.
fn spec_id(rng: &mut Rng) -> String {
    let (year, month, day) = (2000 + rng.below(100), 1 + rng.below(12), 1 + rng.below(28));
    let words = 1 + rng.below(3);
    let slug = (0..words).map(|_| word(rng)).collect::<Vec<_>>().join("-");
    format!("SPEC-{year:04}{month:02}{day:02}-{slug}")
}

/// A host session id as the checker reads it: hex digits (either case) and `-`, 8 to 64 long.
fn session_id(rng: &mut Rng) -> String {
    let len = 8 + rng.below(57) as usize;
    (0..len)
        .map(|_| ID_ALPHABET[rng.below(ID_ALPHABET.len() as u64) as usize] as char)
        .collect()
}

fn words(rng: &mut Rng, max: u64) -> String {
    let count = 1 + rng.below(max) as usize;
    (0..count)
        .map(|_| *rng.pick(PIECES))
        .collect::<Vec<_>>()
        .join(" ")
}

fn subject(rng: &mut Rng) -> String {
    if rng.below(5) == 0 {
        String::new()
    } else {
        words(rng, 6)
    }
}

fn body(rng: &mut Rng) -> String {
    if rng.below(4) == 0 {
        return String::new();
    }
    let lines = 1 + rng.below(4) as usize;
    let mut text = words(rng, 5);
    for _ in 1..lines {
        text.push_str(if rng.bool() { "\n" } else { "\n\n" });
        text.push_str(&words(rng, 5));
    }
    text
}

/// One generated commit, with the trailer values already in the shape `parse_log` returns.
#[derive(Clone)]
struct Gen {
    sha: String,
    specs: Vec<String>,
    sessions: Vec<String>,
    verified: Vec<String>,
    subject: String,
    body: String,
}

impl Gen {
    fn new(rng: &mut Rng) -> Self {
        let many = |rng: &mut Rng, make: &dyn Fn(&mut Rng) -> String| -> Vec<String> {
            (0..rng.below(3)).map(|_| make(rng)).collect()
        };
        let sha_len = 7 + rng.below(34) as usize;
        Gen {
            sha: hex(rng, sha_len),
            specs: many(rng, &spec_id),
            sessions: many(rng, &|r| format!("{}:{}", session_id(r), hex(r, 12))),
            verified: many(rng, &|r| {
                let a = *r.pick(&VERIFIED_KINDS);
                let b = *r.pick(&VERIFIED_KINDS);
                format!("{a}, {b}")
            }),
            subject: subject(rng),
            body: body(rng),
        }
    }

    fn expected(&self) -> CommitTrailers {
        CommitTrailers {
            sha: self.sha.clone(),
            specs: self.specs.clone(),
            sessions: self.sessions.clone(),
            verified: self.verified.clone(),
            subject: self.subject.clone(),
            body: self.body.clone(),
            garbled: false,
        }
    }

    /// The six fields of the `git log` record: sha, specs, sessions, verified, subject, body.
    fn fields(&self) -> [String; 6] {
        [
            self.sha.clone(),
            self.specs.join("\x03"),
            self.sessions.join("\x03"),
            self.verified.join("\x03"),
            self.subject.clone(),
            self.body.clone(),
        ]
    }
}

fn serialize(records: &[[String; 6]]) -> String {
    records
        .iter()
        .map(|f| format!("\x1e{}", f.join("\x1f")))
        .collect()
}

fn commits(rng: &mut Rng) -> Vec<Gen> {
    (0..1 + rng.below(5)).map(|_| Gen::new(rng)).collect()
}

/// REQ-11: what the log format writes, `parse_log` reads back unchanged and never garbled.
fn check_log_round_trip(parse: &Parse, rng: &mut Rng) -> Result<(), String> {
    let generated = commits(rng);
    let text = serialize(&generated.iter().map(Gen::fields).collect::<Vec<_>>());
    let want: Vec<CommitTrailers> = generated.iter().map(Gen::expected).collect();
    let got = parse(&text);
    if got != want {
        return Err(format!(
            "log {text:?} read back as {got:?}, wanted {want:?}"
        ));
    }
    Ok(())
}

/// REQ-13: the lines `format_verified` prints, read through `parse_log` and `judge`.
fn check_format_round_trip(format: &FormatVerified, rng: &mut Rng) -> Result<(), String> {
    let spec = spec_id(rng);
    let id = session_id(rng);
    let key = hex(rng, 12);
    let source = rng.below(4);
    let with_key = rng.below(4) != 0;
    let kinds: Vec<String> = if rng.bool() {
        Vec::new()
    } else {
        (0..1 + rng.below(3))
            .map(|_| (*rng.pick(&VERIFIED_KINDS)).to_string())
            .collect()
    };
    let get = |name: &str| -> Option<String> {
        match (source, name) {
            (0 | 1, n) if n == SOURCES[0] => Some(id.clone()),
            (2, n) if n == SOURCES[1] => Some(id.clone()),
            _ => None,
        }
    };
    let printed = format(&spec, &kinds, &get, with_key.then_some(key.as_str()))
        .map_err(|e| format!("a valid input was refused ({spec}): {e}"))?;
    let text = printed["text"].as_str().ok_or("no text in the answer")?;
    let (mut specs, mut sessions, mut verified) = (Vec::new(), Vec::new(), Vec::new());
    for line in text.lines() {
        let (name, value) = line
            .split_once(": ")
            .ok_or_else(|| format!("line {line:?} is not a trailer"))?;
        match name {
            n if n == SPEC_KEY => specs.push(value.to_string()),
            n if n == SESSION_KEY => sessions.push(value.to_string()),
            n if n == VERIFIED_KEY => verified.push(value.to_string()),
            other => return Err(format!("an unknown trailer {other:?} in {text:?}")),
        }
    }
    let record = Gen {
        sha: hex(rng, 40),
        specs,
        sessions,
        verified,
        subject: "subject".into(),
        body: String::new(),
    };
    let read = parse_log(&serialize(&[record.fields()]));
    let [c] = read.as_slice() else {
        return Err(format!("one record was read as {} commits", read.len()));
    };
    let known = source < 3 && with_key;
    let want_sessions = if known {
        vec![format!("{id}:{key}")]
    } else {
        Vec::new()
    };
    if c.garbled
        || c.specs != [spec.clone()]
        || c.sessions != want_sessions
        || verified_kinds(&c.verified) != kinds
    {
        return Err(format!("the lines {text:?} were read as {c:?}"));
    }
    let verdict = judge(&read, "spec/example");
    if verdict["failing"].as_u64() != Some(0) {
        return Err(format!(
            "the checker rejects the printed lines {text:?}: {verdict}"
        ));
    }
    Ok(())
}

/// What an injected separator carries after it.
const WORD_TAILS: &[&str] = &["zz top", "zz\x1f", "zz\x1fSPEC-20261120-x\x1f\x1f\x1f\x1f"];
const HEX_TAILS: &[&str] = &[
    "",
    "abc",
    "ab12cd",
    "abc\x1f",
    "abc\x1fSPEC-20261120-x\x1f\x1f\x1f\x1f",
];

/// REQ-15: a raw separator put in the middle of one commit's message fields gives either the
/// original commits or at least one garbled commit, never a silent wrong reading. The sha is
/// written by git, never by the author, so the injection never lands in it.
fn check_injection(parse: &Parse, tails: &[&str], rng: &mut Rng) -> Result<(), String> {
    let generated = commits(rng);
    let mut records: Vec<[String; 6]> = generated.iter().map(Gen::fields).collect();
    let victim = rng.below(records.len() as u64) as usize;
    let field = 1 + rng.below(5) as usize;
    let text = records[victim][field].clone();
    let boundaries: Vec<usize> = text
        .char_indices()
        .map(|(i, _)| i)
        .chain(std::iter::once(text.len()))
        .collect();
    let at = boundaries[rng.below(boundaries.len() as u64) as usize];
    let separator = if rng.bool() { '\x1e' } else { '\x1f' };
    let payload = format!("{separator}{}", rng.pick(tails));
    records[victim][field].insert_str(at, &payload);
    let log = serialize(&records);
    let original: Vec<CommitTrailers> = generated.iter().map(Gen::expected).collect();
    let got = parse(&log);
    if got == original || got.iter().any(|c| c.garbled) {
        Ok(())
    } else {
        Err(format!(
            "payload {payload:?} injected at byte {at} of field {field} of commit {victim} was read \
             silently as {got:?} (original {original:?})"
        ))
    }
}

fn pass(result: Result<(), String>) {
    if let Err(reason) = result {
        panic!("{reason}");
    }
}

#[test]
fn a_log_record_reads_back_as_written() {
    run_property("trailers_log_round_trip", MIN_CASES, |rng| {
        pass(check_log_round_trip(&parse_log, rng));
    });
}

#[test]
fn printed_trailer_lines_are_lines_the_checker_accepts() {
    run_property("trailers_format_round_trip", MIN_CASES, |rng| {
        pass(check_format_round_trip(
            &|spec, kinds, get, key| format_verified(spec, kinds, get, key),
            rng,
        ));
    });
}

#[test]
fn a_separator_followed_by_text_is_never_read_silently() {
    run_property("trailers_injection_words", MIN_CASES, |rng| {
        pass(check_injection(&parse_log, WORD_TAILS, rng));
    });
}

/// The hole (observed, not fixed here): a separator whose tail is hex only turns the rest of a
/// record into a second, ungarbled "commit" that takes over the fields after it, and the commit
/// it was injected into loses them. An empty tail is a different loss: the record has no sha
/// and is dropped, so the fields before the separator are kept but the ones after it vanish
/// without a trace. `src/trailers.rs` is outside this suite's scope, so the replay is recorded
/// on the card and the test stays ignored. Run it with `cargo test --test prop_trailers --
/// --ignored`.
///
/// The last `HEX_TAILS` value is a complete fake record (a hex id and every field after it).
/// No change to the garbled rule alone can tell it from a real commit; it needs a change of
/// the log format (a count or length frame). So the test cannot turn green by a reader fix of
/// the garbled rule only: drop that value, or change the format, before removing the `#[ignore]`.
#[test]
#[ignore = "finding: a separator in a message followed by hex-looking text is read silently; \
            fails at case 6 with the default seed (PROP_ONLY=6)"]
fn a_separator_followed_by_a_commit_id_look_is_never_read_silently() {
    run_property("trailers_injection_hex", MIN_CASES, |rng| {
        pass(check_injection(&parse_log, HEX_TAILS, rng));
    });
}

/// How many of `cases` the checker fails, on its own fixed seed.
fn failures(cases: u64, mut checker: impl FnMut(&mut Rng) -> Result<(), String>) -> u64 {
    (0..cases)
        .filter(|i| {
            let mut rng = Rng::new(0xCA4A_2027 ^ i.wrapping_mul(0x9e37_79b9_7f4a_7c15 | 1));
            checker(&mut rng).is_err()
        })
        .count() as u64
}

/// The canary of the suite: every checker fails on a deliberately broken variant.
#[test]
fn canary_every_checker_fails_on_its_broken_variant() {
    let reversed = |text: &str| {
        let mut read = parse_log(text);
        read.reverse();
        read
    };
    assert!(
        failures(MIN_CASES, |rng| check_log_round_trip(&reversed, rng)) > 0,
        "the log checker accepted a reader that reverses the commits"
    );
    let renamed = |spec: &str, kinds: &[String], get: Getter<'_>, key: Option<&str>| {
        let mut printed = format_verified(spec, kinds, get, key)?;
        let text = printed["text"]
            .as_str()
            .unwrap_or("")
            .replace(SESSION_KEY, "Moai-Sessions");
        printed["text"] = Value::String(text);
        Ok(printed)
    };
    assert!(
        failures(MIN_CASES, |rng| check_format_round_trip(&renamed, rng)) > 0,
        "the format checker accepted a formatter that renames the session trailer"
    );
    let never_garbled = |text: &str| {
        let mut read = parse_log(text);
        for c in &mut read {
            c.garbled = false;
        }
        read
    };
    assert!(
        failures(MIN_CASES, |rng| check_injection(
            &never_garbled,
            WORD_TAILS,
            rng
        )) > 0,
        "the injection checker accepted a reader that never marks a commit garbled"
    );
}
