//! Audit verdict line property tests (SPEC-20261120-property-tests-wave2): the line the merge
//! gate reads is the line the formatter prints, the two readers never both read one line, a
//! changed line is never read, and only a comment's first line is a verdict.
//!
//! Every property is a checker that returns a result, so the canary test can feed the same
//! checker a deliberately broken variant and require it to fail: a checker that cannot fail
//! proves nothing. A failing case is replayed with `PROP_SEED=<seed> PROP_ONLY=<i>`.

#[path = "common/prop.rs"]
mod prop;

use moai_cli::gate::{format_verdict_line, parse_verdict, parse_verdict_short_base, verdict_line};
use prop::{run_property, Rng, MIN_CASES};

type Reading = Option<(String, String, String)>;
type Format = dyn Fn(u64, &str, &str, &str, &str) -> Result<String, String>;
type Reader = dyn Fn(&str) -> Reading;
type FirstLine = dyn for<'a> Fn(&'a str) -> Option<&'a str>;

const HEX: &[u8] = b"0123456789abcdef";
const FULL_VERDICT_SEPARATOR: &str = " — verdict: ";

fn hex(rng: &mut Rng, len: usize) -> String {
    (0..len)
        .map(|_| HEX[rng.below(HEX.len() as u64) as usize] as char)
        .collect()
}

fn verdict_word(rng: &mut Rng) -> &'static str {
    let word: &&'static str = rng.pick(&["PASS", "FAIL"]);
    word
}

/// A one-line title the formatter must accept: words, parentheses, the words the parser keys
/// on (`head`, `base`, `verdict:`), non-ASCII text, a tab, and sometimes a whole verdict line
/// quoted inside it (the real verdict is the last ` head ` of the line).
fn title(rng: &mut Rng) -> String {
    const PIECES: &[&str] = &[
        "fix",
        "audit",
        "C",
        "(",
        ")",
        "#12",
        "head",
        "base",
        "verdict:",
        "—",
        "감사",
        "수리",
        "PASS",
        "x-y_z",
        "a.b",
        "\"q\"",
        "tab\there",
    ];
    let count = 1 + rng.below(6) as usize;
    let mut words: Vec<String> = (0..count)
        .map(|_| (*rng.pick(PIECES)).to_string())
        .collect();
    if rng.below(4) == 0 {
        let (h, b) = (hex(rng, 40), hex(rng, 40));
        words.push(format!("head {h} base {b}{FULL_VERDICT_SEPARATOR}PASS"));
    }
    let mut text = words.join(" ");
    if rng.below(4) == 0 {
        text.insert(0, ' ');
    }
    if rng.below(4) == 0 {
        text.push(' ');
    }
    text
}

/// A head or base the formatter must refuse: wrong length, uppercase, a non-hex letter, empty.
fn bad_hex(rng: &mut Rng) -> String {
    match rng.below(5) {
        0 => hex(rng, 39),
        1 => hex(rng, 41),
        2 => format!("A{}", hex(rng, 39)),
        3 => format!("g{}", hex(rng, 39)),
        _ => String::new(),
    }
}

/// REQ-1: a valid input is formatted, and the line reads back as the same head, base and
/// verdict, and is its own verdict line.
fn check_round_trip(fmt: &Format, rng: &mut Rng) -> Result<(), String> {
    let pr = rng.below(100_000);
    let title = title(rng);
    let head = hex(rng, 40);
    let mut base = hex(rng, 40);
    if base == head {
        base.replace_range(0..1, if head.starts_with('0') { "1" } else { "0" });
    }
    let verdict = verdict_word(rng);
    let line = fmt(pr, &title, &head, &base, verdict)
        .map_err(|e| format!("a valid input was refused ({title:?}): {e}"))?;
    let want = Some((head.clone(), base.clone(), verdict.to_string()));
    let got = parse_verdict(&line);
    if got != want {
        return Err(format!(
            "line {line:?} read back as {got:?}, wanted {want:?}"
        ));
    }
    if !line.starts_with(&format!("AUDIT #{pr} (")) || verdict_line(&line) != Some(line.as_str()) {
        return Err(format!("line {line:?} is not its own verdict line"));
    }
    Ok(())
}

/// REQ-3: every input the formatter must refuse is refused.
fn check_refusals(fmt: &Format, rng: &mut Rng) -> Result<(), String> {
    let pr = rng.below(100_000);
    let (mut head, mut base) = (hex(rng, 40), hex(rng, 40));
    let mut verdict = "PASS".to_string();
    let mut title = "fix (x)".to_string();
    let what = match rng.below(4) {
        0 => {
            head = bad_hex(rng);
            "head"
        }
        1 => {
            base = bad_hex(rng);
            "base"
        }
        2 => {
            verdict = (*rng.pick(&["pass", "fail", "MAYBE", "", "PASS ", "PASSFAIL"])).to_string();
            "verdict"
        }
        _ => {
            title = (*rng.pick(&["", "   ", "a\nb", "a\rb", "\n"])).to_string();
            "title"
        }
    };
    match fmt(pr, &title, &head, &base, &verdict) {
        Err(_) => Ok(()),
        Ok(line) => Err(format!("a bad {what} was formatted: {line:?}")),
    }
}

/// REQ-5: a 40-hex base is read by the full reader only, a 7 to 39 hex base by the short
/// reader only; no line is read by both.
fn check_readers(full: &Reader, short: &Reader, rng: &mut Rng) -> Result<(), String> {
    let head = hex(rng, 40);
    let verdict = verdict_word(rng);
    let is_full = rng.bool();
    let base_len = if is_full {
        40
    } else {
        7 + rng.below(33) as usize
    };
    let base = hex(rng, base_len);
    let pr = rng.below(100_000);
    let title = title(rng);
    let line =
        format!("AUDIT #{pr} ({title}) head {head} base {base}{FULL_VERDICT_SEPARATOR}{verdict}");
    let want = Some((head, base, verdict.to_string()));
    let (by_full, by_short) = (full(&line), short(&line));
    let (want_full, want_short) = if is_full { (want, None) } else { (None, want) };
    if by_full != want_full || by_short != want_short {
        return Err(format!(
            "line {line:?} (base of {base_len} hex): full reader {by_full:?}, short reader {by_short:?}"
        ));
    }
    Ok(())
}

/// REQ-7: a valid line, changed in one of seven ways, is never read.
fn check_mutations(parse: &Reader, rng: &mut Rng) -> Result<(), String> {
    let (head, base) = (hex(rng, 40), hex(rng, 40));
    let verdict = verdict_word(rng);
    let title = title(rng);
    let mut line =
        format!("AUDIT #7 ({title}) head {head} base {base}{FULL_VERDICT_SEPARATOR}{verdict}");
    let head_at = line.rfind(" head ").expect("the line holds its own head") + 6;
    let base_at = head_at + 40 + " base ".len();
    let which = rng.below(7);
    let at_head = head_at + rng.below(40) as usize;
    let at_base = base_at + rng.below(40) as usize;
    let replacement = (*rng.pick(&["g", "G", "A", "z", "-", " ", "é"])).to_string();
    let what = match which {
        0 => {
            line.replace_range(at_head..at_head + 1, &replacement);
            "a head character replaced"
        }
        1 => {
            line.replace_range(at_base..at_base + 1, &replacement);
            "a base character replaced"
        }
        2 => {
            line.remove(at_head);
            "a head character dropped"
        }
        3 => {
            line.remove(at_base);
            "a base character dropped"
        }
        4 => {
            line.push_str(" extra");
            "a word appended"
        }
        5 => {
            line = line.replace(FULL_VERDICT_SEPARATOR, " - verdict: ");
            "the separator changed"
        }
        _ => {
            line.truncate(line.len() - verdict.len());
            "the verdict cut off"
        }
    };
    match parse(&line) {
        None => Ok(()),
        Some(read) => Err(format!("{what}: {line:?} was read as {read:?}")),
    }
}

/// REQ-9: the verdict is a comment's first line; a verdict line further down is a quote.
fn check_first_line(first_line: &FirstLine, rng: &mut Rng) -> Result<(), String> {
    let (head, base) = (hex(rng, 40), hex(rng, 40));
    let title = title(rng);
    let line = format_verdict_line(rng.below(100_000), &title, &head, &base, verdict_word(rng))
        .map_err(|e| format!("a valid input was refused: {e}"))?;
    let quoted = format!("AUDIT #1 (x) head {head} base {base}{FULL_VERDICT_SEPARATOR}PASS");
    let extra = rng.below(4) as usize;
    let tail: Vec<&str> = (0..extra)
        .map(|_| *rng.pick(&[quoted.as_str(), "> quoted", "", "note"]))
        .collect();
    let first_body = std::iter::once(line.as_str())
        .chain(tail)
        .collect::<Vec<_>>()
        .join("\n");
    if first_line(&first_body) != Some(line.as_str()) {
        return Err(format!(
            "the first line was not the verdict in {first_body:?}"
        ));
    }
    let lead = *rng.pick(&["note", "", "audit #1", " AUDIT #1 (x)", "> AUDIT #1"]);
    let later = format!("{lead}\n{line}");
    match first_line(&later) {
        None => Ok(()),
        Some(read) => Err(format!("a later line {read:?} was read as the verdict")),
    }
}

fn pass(result: Result<(), String>) {
    if let Err(reason) = result {
        panic!("{reason}");
    }
}

#[test]
fn a_formatted_verdict_line_reads_back() {
    run_property("verdict_round_trip", MIN_CASES, |rng| {
        pass(check_round_trip(&format_verdict_line, rng));
    });
}

#[test]
fn the_formatter_refuses_what_the_reader_would_refuse() {
    run_property("verdict_refusals", MIN_CASES, |rng| {
        pass(check_refusals(&format_verdict_line, rng));
    });
}

#[test]
fn no_line_is_read_by_both_readers() {
    run_property("verdict_readers", MIN_CASES, |rng| {
        pass(check_readers(
            &parse_verdict,
            &parse_verdict_short_base,
            rng,
        ));
    });
}

#[test]
fn a_changed_verdict_line_is_never_read() {
    run_property("verdict_mutations", MIN_CASES, |rng| {
        pass(check_mutations(&parse_verdict, rng));
    });
}

#[test]
fn only_the_first_line_is_the_verdict() {
    run_property("verdict_first_line", MIN_CASES, |rng| {
        pass(check_first_line(&verdict_line, rng));
    });
}

/// How many of `cases` the checker fails, on its own fixed seed.
fn failures(cases: u64, mut checker: impl FnMut(&mut Rng) -> Result<(), String>) -> u64 {
    (0..cases)
        .filter(|i| {
            let mut rng = Rng::new(0xCA4A_2026 ^ i.wrapping_mul(0x9e37_79b9_7f4a_7c15 | 1));
            checker(&mut rng).is_err()
        })
        .count() as u64
}

/// The canary of the suite: every checker fails on a deliberately broken variant, so none of
/// them is a checker that cannot fail.
#[test]
fn canary_every_checker_fails_on_its_broken_variant() {
    let swapped = |pr: u64, t: &str, h: &str, b: &str, v: &str| format_verdict_line(pr, t, b, h, v);
    assert!(
        failures(MIN_CASES, |rng| check_round_trip(&swapped, rng)) > 0,
        "the round-trip checker accepted a formatter that swaps head and base"
    );
    let unchecked = |pr: u64, t: &str, h: &str, b: &str, v: &str| {
        Ok(format!(
            "AUDIT #{pr} ({}) head {h} base {b}{FULL_VERDICT_SEPARATOR}{v}",
            t.trim()
        ))
    };
    assert!(
        failures(MIN_CASES, |rng| check_refusals(&unchecked, rng)) > 0,
        "the refusal checker accepted a formatter that checks nothing"
    );
    let either = |line: &str| parse_verdict(line).or_else(|| parse_verdict_short_base(line));
    assert!(
        failures(MIN_CASES, |rng| check_readers(
            &either,
            &parse_verdict_short_base,
            rng
        )) > 0,
        "the reader checker accepted a full reader that also reads short bases"
    );
    let lowercased = |line: &str| parse_verdict(&line.to_lowercase());
    assert!(
        failures(MIN_CASES, |rng| check_mutations(&lowercased, rng)) > 0,
        "the mutation checker accepted a parser that lowercases its input"
    );
    fn any_line(body: &str) -> Option<&str> {
        body.lines().find(|l| l.starts_with("AUDIT #"))
    }
    assert!(
        failures(MIN_CASES, |rng| check_first_line(&any_line, rng)) > 0,
        "the first-line checker accepted a reader that searches every line"
    );
}
